// Package storage provides SQLite persistent storage implementation.
package storage

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"
	"unified-thinking/internal/types"
)

// SQLiteStorage implements persistent storage with SQLite + in-memory cache
type SQLiteStorage struct {
	db    *sql.DB
	cache *MemoryStorage // Write-through cache for fast reads

	mu             sync.RWMutex
	activeBranchID string
	thoughtCounter atomic.Int64
	branchCounter  atomic.Int64
	//nolint:unused // Reserved for future use
	_insightCounter atomic.Int64 // Reserved for future use
	//nolint:unused // Reserved for future use
	_validationCounter atomic.Int64 // Reserved for future use
	//nolint:unused // Reserved for future use
	_relationshipCounter atomic.Int64 // Reserved for future use

	// Prepared statements
	stmtInsertThought      *sql.Stmt
	stmtGetThought         *sql.Stmt
	stmtSearchFTS          *sql.Stmt
	stmtInsertBranch       *sql.Stmt
	stmtGetBranch          *sql.Stmt
	stmtUpdateBranchAccess *sql.Stmt
	stmtInsertInsight      *sql.Stmt
	stmtInsertCrossRef     *sql.Stmt
	stmtInsertValidation   *sql.Stmt
}

// NewSQLiteStorage creates a new SQLite storage backend
func NewSQLiteStorage(dbPath string, timeoutMs int) (*SQLiteStorage, error) {
	// Validate path
	if dbPath == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	// Build DSN with pragmas
	dsn := dbPath + fmt.Sprintf("?_busy_timeout=%d", timeoutMs)

	// Open database
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool (SQLite works best with limited connections)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(0)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		_ = db.Close() // Ignore close error during cleanup
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure SQLite pragmas
	if err := configureSQLite(db); err != nil {
		_ = db.Close() // Ignore close error during cleanup
		return nil, fmt.Errorf("failed to configure SQLite: %w", err)
	}

	// Initialize schema
	if err := initializeSchema(db); err != nil {
		_ = db.Close() // Ignore close error during cleanup
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	s := &SQLiteStorage{
		db:    db,
		cache: NewMemoryStorage(),
	}

	// Prepare statements
	if err := s.prepareStatements(); err != nil {
		_ = db.Close() // Ignore close error during cleanup
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	// Warm cache with recent data
	if err := s.warmCache(); err != nil {
		log.Printf("Warning: failed to warm cache: %v", err)
	}

	log.Printf("SQLite storage initialized successfully at %s", dbPath)
	return s, nil
}

// prepareStatements creates reusable prepared statements
func (s *SQLiteStorage) prepareStatements() error {
	var err error

	s.stmtInsertThought, err = s.db.Prepare(`
		INSERT INTO thoughts (
			id, content, mode, branch_id, parent_id, type, confidence,
			timestamp, key_points, metadata, is_rebellion, challenges_assumption
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content=excluded.content,
			confidence=excluded.confidence,
			timestamp=excluded.timestamp
	`)
	if err != nil {
		return fmt.Errorf("prepare insert thought: %w", err)
	}

	s.stmtGetThought, err = s.db.Prepare(`
		SELECT id, content, mode, branch_id, parent_id, type, confidence,
		       timestamp, key_points, metadata, is_rebellion, challenges_assumption
		FROM thoughts WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare get thought: %w", err)
	}

	s.stmtSearchFTS, err = s.db.Prepare(`
		SELECT t.id, t.content, t.mode, t.branch_id, t.parent_id, t.type,
		       t.confidence, t.timestamp, t.key_points, t.metadata,
		       t.is_rebellion, t.challenges_assumption
		FROM thoughts_fts fts
		JOIN thoughts t ON t.rowid = fts.rowid
		WHERE fts.content MATCH ? AND (? = '' OR t.mode = ?)
		ORDER BY t.timestamp DESC
		LIMIT ? OFFSET ?
	`)
	if err != nil {
		return fmt.Errorf("prepare FTS search: %w", err)
	}

	s.stmtInsertBranch, err = s.db.Prepare(`
		INSERT INTO branches (
			id, parent_branch_id, state, priority, confidence,
			created_at, updated_at, last_accessed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			state=excluded.state,
			priority=excluded.priority,
			confidence=excluded.confidence,
			updated_at=excluded.updated_at
	`)
	if err != nil {
		return fmt.Errorf("prepare insert branch: %w", err)
	}

	s.stmtGetBranch, err = s.db.Prepare(`
		SELECT id, parent_branch_id, state, priority, confidence,
		       created_at, updated_at, last_accessed_at
		FROM branches WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare get branch: %w", err)
	}

	s.stmtUpdateBranchAccess, err = s.db.Prepare(`
		UPDATE branches SET last_accessed_at = ? WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare update branch access: %w", err)
	}

	s.stmtInsertInsight, err = s.db.Prepare(`
		INSERT INTO insights (
			id, type, content, context, parent_insights,
			applicability_score, supporting_evidence, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert insight: %w", err)
	}

	s.stmtInsertCrossRef, err = s.db.Prepare(`
		INSERT INTO cross_refs (
			id, from_branch, to_branch, type, reason, strength, touchpoints, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert cross_ref: %w", err)
	}

	s.stmtInsertValidation, err = s.db.Prepare(`
		INSERT INTO validations (
			id, insight_id, thought_id, is_valid, validation_data, reason, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert validation: %w", err)
	}

	return nil
}

// warmCache loads recent thoughts and branches into memory on startup
func (s *SQLiteStorage) warmCache() error {
	// Load recent 1000 thoughts
	rows, err := s.db.Query(`
		SELECT id, content, mode, branch_id, parent_id, type, confidence,
		       timestamp, key_points, metadata, is_rebellion, challenges_assumption
		FROM thoughts
		ORDER BY timestamp DESC
		LIMIT 1000
	`)
	if err != nil {
		return fmt.Errorf("failed to query thoughts: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows.Err() will catch any real errors

	for rows.Next() {
		thought, err := s.scanThought(rows)
		if err != nil {
			log.Printf("Warning: failed to scan thought: %v", err)
			continue
		}
		if err := s.cache.StoreThought(thought); err != nil {
			log.Printf("Warning: failed to cache thought: %v", err)
		}
	}

	log.Printf("Warmed cache with %d thoughts", len(s.cache.thoughts))
	return nil
}

// StoreThought persists a thought to database and cache
func (s *SQLiteStorage) StoreThought(thought *types.Thought) error {
	// Generate ID if needed
	if thought.ID == "" {
		counter := s.thoughtCounter.Add(1)
		thought.ID = fmt.Sprintf("thought-%d-%d", time.Now().Unix(), counter)
	}

	// Marshal JSON fields
	keyPointsJSON, _ := json.Marshal(thought.KeyPoints)
	metadataJSON, _ := json.Marshal(thought.Metadata)

	// Handle optional foreign keys - convert empty strings to NULL
	var branchID, parentID interface{}
	if thought.BranchID != "" {
		branchID = thought.BranchID
	}
	if thought.ParentID != "" {
		parentID = thought.ParentID
	}

	// Write to database
	_, err := s.stmtInsertThought.Exec(
		thought.ID, thought.Content, thought.Mode, branchID, parentID,
		thought.Type, thought.Confidence, thought.Timestamp.Unix(),
		keyPointsJSON, metadataJSON,
		boolToInt(thought.IsRebellion), boolToInt(thought.ChallengesAssumption),
	)
	if err != nil {
		return fmt.Errorf("failed to insert thought: %w", err)
	}

	// Update cache
	return s.cache.StoreThought(thought)
}

// GetThought retrieves a thought by ID (cache-first)
func (s *SQLiteStorage) GetThought(id string) (*types.Thought, error) {
	// Try cache first
	if thought, err := s.cache.GetThought(id); err == nil {
		return thought, nil
	}

	// Cache miss: fetch from database
	thought, err := s.fetchThought(id)
	if err != nil {
		return nil, err
	}

	// Warm cache
	if err := s.cache.StoreThought(thought); err != nil {
		log.Printf("Warning: failed to warm cache with thought: %v", err)
	}
	return copyThought(thought), nil
}

// fetchThought retrieves a thought from database
func (s *SQLiteStorage) fetchThought(id string) (*types.Thought, error) {
	thought := &types.Thought{}
	var keyPointsJSON, metadataJSON []byte
	var branchID, parentID sql.NullString
	var isRebellion, challengesAssumption int
	var timestamp int64

	err := s.stmtGetThought.QueryRow(id).Scan(
		&thought.ID, &thought.Content, &thought.Mode, &branchID,
		&parentID, &thought.Type, &thought.Confidence, &timestamp,
		&keyPointsJSON, &metadataJSON, &isRebellion, &challengesAssumption,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("thought not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thought: %w", err)
	}

	if branchID.Valid {
		thought.BranchID = branchID.String
	}
	if parentID.Valid {
		thought.ParentID = parentID.String
	}
	thought.Timestamp = time.Unix(timestamp, 0)
	thought.IsRebellion = isRebellion == 1
	thought.ChallengesAssumption = challengesAssumption == 1

	if len(keyPointsJSON) > 0 {
		if err := json.Unmarshal(keyPointsJSON, &thought.KeyPoints); err != nil {
			log.Printf("Warning: failed to unmarshal thought key points: %v", err)
		}
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &thought.Metadata); err != nil {
			log.Printf("Warning: failed to unmarshal thought metadata: %v", err)
		}
	}
	if thought.Metadata == nil {
		thought.Metadata = make(map[string]interface{})
	}

	return thought, nil
}

// scanThought scans a thought from a SQL row
func (s *SQLiteStorage) scanThought(row interface{ Scan(...interface{}) error }) (*types.Thought, error) {
	thought := &types.Thought{}
	var keyPointsJSON, metadataJSON []byte
	var branchID, parentID sql.NullString
	var isRebellion, challengesAssumption int
	var timestamp int64

	err := row.Scan(
		&thought.ID, &thought.Content, &thought.Mode, &branchID,
		&parentID, &thought.Type, &thought.Confidence, &timestamp,
		&keyPointsJSON, &metadataJSON, &isRebellion, &challengesAssumption,
	)
	if err != nil {
		return nil, err
	}

	if branchID.Valid {
		thought.BranchID = branchID.String
	}
	if parentID.Valid {
		thought.ParentID = parentID.String
	}
	thought.Timestamp = time.Unix(timestamp, 0)
	thought.IsRebellion = isRebellion == 1
	thought.ChallengesAssumption = challengesAssumption == 1

	if len(keyPointsJSON) > 0 {
		if err := json.Unmarshal(keyPointsJSON, &thought.KeyPoints); err != nil {
			log.Printf("Warning: failed to unmarshal thought key points: %v", err)
		}
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &thought.Metadata); err != nil {
			log.Printf("Warning: failed to unmarshal thought metadata: %v", err)
		}
	}
	if thought.Metadata == nil {
		thought.Metadata = make(map[string]interface{})
	}

	return thought, nil
}

// scanInsight scans an insight from a database row
func (s *SQLiteStorage) scanInsight(row interface{ Scan(...interface{}) error }) (*types.Insight, error) {
	insight := &types.Insight{}
	var contextJSON, parentInsightsJSON, supportingEvidenceJSON []byte
	var createdAt int64

	err := row.Scan(
		&insight.ID, &insight.Type, &insight.Content, &contextJSON,
		&insight.ApplicabilityScore,
		&parentInsightsJSON, &supportingEvidenceJSON, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	insight.CreatedAt = time.Unix(createdAt, 0)

	if len(contextJSON) > 0 {
		if err := json.Unmarshal(contextJSON, &insight.Context); err != nil {
			log.Printf("Warning: failed to unmarshal insight context: %v", err)
		}
	}
	if len(parentInsightsJSON) > 0 {
		if err := json.Unmarshal(parentInsightsJSON, &insight.ParentInsights); err != nil {
			log.Printf("Warning: failed to unmarshal insight parent insights: %v", err)
		}
	}
	if len(supportingEvidenceJSON) > 0 {
		if err := json.Unmarshal(supportingEvidenceJSON, &insight.SupportingEvidence); err != nil {
			log.Printf("Warning: failed to unmarshal insight supporting evidence: %v", err)
		}
	}

	// Initialize empty slices if nil
	if insight.Context == nil {
		insight.Context = []string{}
	}
	if insight.ParentInsights == nil {
		insight.ParentInsights = []string{}
	}
	if insight.SupportingEvidence == nil {
		insight.SupportingEvidence = make(map[string]interface{})
	}

	return insight, nil
}

// SearchThoughts searches for thoughts using full-text search
func (s *SQLiteStorage) SearchThoughts(query string, mode types.ThinkingMode, limit, offset int) []*types.Thought {
	if limit <= 0 || limit > MaxSearchResults {
		limit = MaxSearchResults
	}

	// No query and no mode: use cache
	if query == "" && mode == "" {
		return s.cache.SearchThoughts(query, mode, limit, offset)
	}

	// Full-text search: use SQLite FTS5
	if query != "" {
		return s.searchThoughtsFTS(query, mode, limit, offset)
	}

	// Mode-only filter: use cache
	return s.cache.SearchThoughts(query, mode, limit, offset)
}

// searchThoughtsFTS performs full-text search using SQLite FTS5
func (s *SQLiteStorage) searchThoughtsFTS(query string, mode types.ThinkingMode, limit, offset int) []*types.Thought {
	modeStr := string(mode)
	rows, err := s.stmtSearchFTS.Query(query, modeStr, modeStr, limit, offset)
	if err != nil {
		log.Printf("FTS search error: %v", err)
		return nil
	}
	defer rows.Close() //nolint:errcheck // rows.Err() will catch any real errors

	results := make([]*types.Thought, 0, limit)
	for rows.Next() {
		thought, err := s.scanThought(rows)
		if err != nil {
			log.Printf("FTS scan error: %v", err)
			continue
		}
		results = append(results, thought)
	}

	return results
}

// StoreBranch persists a branch to database and cache
func (s *SQLiteStorage) StoreBranch(branch *types.Branch) error {
	// Generate ID if needed
	if branch.ID == "" {
		counter := s.branchCounter.Add(1)
		branch.ID = fmt.Sprintf("branch-%d-%d", time.Now().Unix(), counter)
	}

	// Handle optional foreign key - convert empty string to NULL
	var parentBranchID interface{}
	if branch.ParentBranchID != "" {
		parentBranchID = branch.ParentBranchID
	}

	// Write to database
	_, err := s.stmtInsertBranch.Exec(
		branch.ID, parentBranchID, branch.State, branch.Priority, branch.Confidence,
		branch.CreatedAt.Unix(), branch.UpdatedAt.Unix(), branch.LastAccessedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert branch: %w", err)
	}

	// Update cache
	return s.cache.StoreBranch(branch)
}

// GetBranch retrieves a branch by ID (cache-first)
func (s *SQLiteStorage) GetBranch(id string) (*types.Branch, error) {
	// Try cache first
	if branch, err := s.cache.GetBranch(id); err == nil {
		return branch, nil
	}

	// Cache miss: fetch from database
	branch := &types.Branch{}
	var parentBranchID sql.NullString
	var createdAt, updatedAt, lastAccessedAt int64

	err := s.stmtGetBranch.QueryRow(id).Scan(
		&branch.ID, &parentBranchID, &branch.State, &branch.Priority, &branch.Confidence,
		&createdAt, &updatedAt, &lastAccessedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("branch not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branch: %w", err)
	}

	if parentBranchID.Valid {
		branch.ParentBranchID = parentBranchID.String
	}
	branch.CreatedAt = time.Unix(createdAt, 0)
	branch.UpdatedAt = time.Unix(updatedAt, 0)
	branch.LastAccessedAt = time.Unix(lastAccessedAt, 0)

	// Load associated data
	thoughts, err := s.loadBranchThoughts(id)
	if err != nil {
		log.Printf("Warning: failed to load branch thoughts: %v", err)
		thoughts = []*types.Thought{}
	}
	branch.Thoughts = thoughts

	insights, err := s.loadBranchInsights(id)
	if err != nil {
		log.Printf("Warning: failed to load branch insights: %v", err)
		insights = []*types.Insight{}
	}
	branch.Insights = insights

	crossRefs, err := s.loadBranchCrossRefs(id)
	if err != nil {
		log.Printf("Warning: failed to load branch cross-refs: %v", err)
		crossRefs = []*types.CrossRef{}
	}
	branch.CrossRefs = crossRefs

	// Warm cache
	if err := s.cache.StoreBranch(branch); err != nil {
		log.Printf("Warning: failed to warm cache with branch: %v", err)
	}
	return copyBranch(branch), nil
}

// loadBranchThoughts loads all thoughts associated with a branch
func (s *SQLiteStorage) loadBranchThoughts(branchID string) ([]*types.Thought, error) {
	rows, err := s.db.Query(`
		SELECT id, content, mode, branch_id, parent_id, type, confidence,
		       timestamp, key_points, metadata, is_rebellion, challenges_assumption
		FROM thoughts
		WHERE branch_id = ?
		ORDER BY timestamp ASC
	`, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query branch thoughts: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows.Err() will catch any real errors

	thoughts := make([]*types.Thought, 0)
	for rows.Next() {
		thought, err := s.scanThought(rows)
		if err != nil {
			log.Printf("Warning: failed to scan thought: %v", err)
			continue
		}
		thoughts = append(thoughts, thought)
	}

	return thoughts, nil
}

// loadBranchInsights loads all insights associated with a branch
func (s *SQLiteStorage) loadBranchInsights(branchID string) ([]*types.Insight, error) {
	rows, err := s.db.Query(`
		SELECT id, type, content, context, applicability_score,
		       parent_insights, supporting_evidence, created_at
		FROM insights
		WHERE branch_id = ?
		ORDER BY created_at ASC
	`, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query branch insights: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows.Err() will catch any real errors

	insights := make([]*types.Insight, 0)
	for rows.Next() {
		insight, err := s.scanInsight(rows)
		if err != nil {
			log.Printf("Warning: failed to scan insight: %v", err)
			continue
		}
		insights = append(insights, insight)
	}

	return insights, nil
}

// loadBranchCrossRefs loads all cross-references associated with a branch
func (s *SQLiteStorage) loadBranchCrossRefs(branchID string) ([]*types.CrossRef, error) {
	rows, err := s.db.Query(`
		SELECT id, from_branch, to_branch, type, reason, strength, created_at
		FROM cross_refs
		WHERE from_branch = ?
		ORDER BY created_at ASC
	`, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query branch cross-refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows.Err() will catch any real errors

	crossRefs := make([]*types.CrossRef, 0)
	for rows.Next() {
		var id, fromBranch, toBranch, xrefType, reason string
		var strength float64
		var createdAt int64

		err := rows.Scan(&id, &fromBranch, &toBranch, &xrefType, &reason, &strength, &createdAt)
		if err != nil {
			log.Printf("Warning: failed to scan cross-ref: %v", err)
			continue
		}

		crossRef := &types.CrossRef{
			ID:         id,
			FromBranch: fromBranch,
			ToBranch:   toBranch,
			Type:       types.CrossRefType(xrefType),
			Reason:     reason,
			Strength:   strength,
			CreatedAt:  time.Unix(createdAt, 0),
		}
		crossRefs = append(crossRefs, crossRef)
	}

	return crossRefs, nil
}

// ListBranches returns all branches
func (s *SQLiteStorage) ListBranches() []*types.Branch {
	// Use cache for listing
	return s.cache.ListBranches()
}

// GetActiveBranch returns the currently active branch
func (s *SQLiteStorage) GetActiveBranch() (*types.Branch, error) {
	s.mu.RLock()
	branchID := s.activeBranchID
	s.mu.RUnlock()

	if branchID == "" {
		return nil, fmt.Errorf("no active branch")
	}

	return s.GetBranch(branchID)
}

// SetActiveBranch sets the active branch
func (s *SQLiteStorage) SetActiveBranch(branchID string) error {
	s.mu.Lock()
	s.activeBranchID = branchID
	s.mu.Unlock()

	return s.cache.SetActiveBranch(branchID)
}

// UpdateBranchAccess updates last accessed time
func (s *SQLiteStorage) UpdateBranchAccess(branchID string) error {
	now := time.Now()
	_, err := s.stmtUpdateBranchAccess.Exec(now.Unix(), branchID)
	if err != nil {
		return fmt.Errorf("failed to update branch access: %w", err)
	}

	// Update cache
	return s.cache.UpdateBranchAccess(branchID)
}

// StoreInsight persists an insight to database and cache
func (s *SQLiteStorage) StoreInsight(insight *types.Insight) error {
	// Generate ID if needed
	if insight.ID == "" {
		counter := s.thoughtCounter.Add(1) // Reuse counter for simplicity
		insight.ID = fmt.Sprintf("insight-%d-%d", time.Now().Unix(), counter)
	}

	// Marshal JSON fields
	contextJSON, _ := json.Marshal(insight.Context)
	parentInsightsJSON, _ := json.Marshal(insight.ParentInsights)
	supportingEvidenceJSON, _ := json.Marshal(insight.SupportingEvidence)

	// Write to database (branch_id is NULL for now, set via AppendInsightToBranch)
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO insights (id, branch_id, type, content, context, parent_insights,
		                                  applicability_score, supporting_evidence, created_at)
		VALUES (?, NULL, ?, ?, ?, ?, ?, ?, ?)
	`, insight.ID, insight.Type, insight.Content, contextJSON, parentInsightsJSON,
		insight.ApplicabilityScore, supportingEvidenceJSON, insight.CreatedAt.Unix())

	if err != nil {
		return fmt.Errorf("failed to insert insight: %w", err)
	}

	// Update cache
	return s.cache.StoreInsight(insight)
}

func (s *SQLiteStorage) GetInsight(id string) (*types.Insight, error) {
	// Try cache first
	if insight, err := s.cache.GetInsight(id); err == nil {
		return insight, nil
	}

	// Cache miss: fetch from database
	var contextJSON, parentInsightsJSON, supportingEvidenceJSON []byte
	var createdAt int64
	insight := &types.Insight{}

	err := s.db.QueryRow(`
		SELECT id, type, content, context, parent_insights, applicability_score,
		       supporting_evidence, created_at
		FROM insights
		WHERE id = ?
	`, id).Scan(&insight.ID, &insight.Type, &insight.Content, &contextJSON, &parentInsightsJSON,
		&insight.ApplicabilityScore, &supportingEvidenceJSON, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("insight not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch insight: %w", err)
	}

	insight.CreatedAt = time.Unix(createdAt, 0)

	// Unmarshal JSON fields
	if len(contextJSON) > 0 {
		if err := json.Unmarshal(contextJSON, &insight.Context); err != nil {
			log.Printf("Warning: failed to unmarshal insight context: %v", err)
		}
	}
	if len(parentInsightsJSON) > 0 {
		if err := json.Unmarshal(parentInsightsJSON, &insight.ParentInsights); err != nil {
			log.Printf("Warning: failed to unmarshal insight parent insights: %v", err)
		}
	}
	if len(supportingEvidenceJSON) > 0 {
		if err := json.Unmarshal(supportingEvidenceJSON, &insight.SupportingEvidence); err != nil {
			log.Printf("Warning: failed to unmarshal insight supporting evidence: %v", err)
		}
	}

	// Initialize empty slices
	if insight.Context == nil {
		insight.Context = []string{}
	}
	if insight.ParentInsights == nil {
		insight.ParentInsights = []string{}
	}
	if insight.SupportingEvidence == nil {
		insight.SupportingEvidence = make(map[string]interface{})
	}

	// Warm cache
	if err := s.cache.StoreInsight(insight); err != nil {
		log.Printf("Warning: failed to warm cache with insight: %v", err)
	}
	return copyInsight(insight), nil
}

func (s *SQLiteStorage) StoreValidation(validation *types.Validation) error {
	// Generate ID if needed
	if validation.ID == "" {
		counter := s.thoughtCounter.Add(1)
		validation.ID = fmt.Sprintf("validation-%d-%d", time.Now().Unix(), counter)
	}

	// Marshal validation_data JSON
	validationDataJSON, _ := json.Marshal(validation.ValidationData)

	// Write to database
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO validations (id, insight_id, thought_id, is_valid, validation_data, reason, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, validation.ID, validation.InsightID, validation.ThoughtID, boolToInt(validation.IsValid),
		validationDataJSON, validation.Reason, validation.CreatedAt.Unix())

	if err != nil {
		return fmt.Errorf("failed to insert validation: %w", err)
	}

	// Update cache
	return s.cache.StoreValidation(validation)
}

func (s *SQLiteStorage) GetValidation(id string) (*types.Validation, error) {
	// Try cache first
	if validation, err := s.cache.GetValidation(id); err == nil {
		return validation, nil
	}

	// Cache miss: fetch from database
	validation := &types.Validation{}
	var validationDataJSON []byte
	var isValid int
	var createdAt int64

	err := s.db.QueryRow(`
		SELECT id, insight_id, thought_id, is_valid, validation_data, reason, created_at
		FROM validations
		WHERE id = ?
	`, id).Scan(&validation.ID, &validation.InsightID, &validation.ThoughtID, &isValid,
		&validationDataJSON, &validation.Reason, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("validation not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch validation: %w", err)
	}

	validation.IsValid = isValid == 1
	validation.CreatedAt = time.Unix(createdAt, 0)

	// Unmarshal validation_data
	if len(validationDataJSON) > 0 {
		if err := json.Unmarshal(validationDataJSON, &validation.ValidationData); err != nil {
			log.Printf("Warning: failed to unmarshal validation data: %v", err)
		}
	}
	if validation.ValidationData == nil {
		validation.ValidationData = make(map[string]interface{})
	}

	// Warm cache
	if err := s.cache.StoreValidation(validation); err != nil {
		log.Printf("Warning: failed to warm cache with validation: %v", err)
	}
	return copyValidation(validation), nil
}

func (s *SQLiteStorage) StoreRelationship(rel *types.Relationship) error {
	// Generate ID if needed
	if rel.ID == "" {
		counter := s.thoughtCounter.Add(1)
		rel.ID = fmt.Sprintf("relationship-%d-%d", time.Now().Unix(), counter)
	}

	// Marshal metadata JSON
	metadataJSON, _ := json.Marshal(rel.Metadata)

	// Write to database
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO relationships (id, from_state_id, to_state_id, type, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, rel.ID, rel.FromStateID, rel.ToStateID, rel.Type, metadataJSON, rel.CreatedAt.Unix())

	if err != nil {
		return fmt.Errorf("failed to insert relationship: %w", err)
	}

	// Update cache
	return s.cache.StoreRelationship(rel)
}

func (s *SQLiteStorage) GetRelationship(id string) (*types.Relationship, error) {
	// Try cache first
	if relationship, err := s.cache.GetRelationship(id); err == nil {
		return relationship, nil
	}

	// Cache miss: fetch from database
	relationship := &types.Relationship{}
	var metadataJSON []byte
	var createdAt int64

	err := s.db.QueryRow(`
		SELECT id, from_state_id, to_state_id, type, metadata, created_at
		FROM relationships
		WHERE id = ?
	`, id).Scan(&relationship.ID, &relationship.FromStateID, &relationship.ToStateID,
		&relationship.Type, &metadataJSON, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("relationship not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationship: %w", err)
	}

	relationship.CreatedAt = time.Unix(createdAt, 0)

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &relationship.Metadata); err != nil {
			log.Printf("Warning: failed to unmarshal relationship metadata: %v", err)
		}
	}
	if relationship.Metadata == nil {
		relationship.Metadata = make(map[string]interface{})
	}

	// Warm cache
	if err := s.cache.StoreRelationship(relationship); err != nil {
		log.Printf("Warning: failed to warm cache with relationship: %v", err)
	}
	return copyRelationship(relationship), nil
}

func (s *SQLiteStorage) AppendThoughtToBranch(branchID string, thought *types.Thought) error {
	// Thought should already be persisted with branch_id via StoreThought
	// Just update cache
	return s.cache.AppendThoughtToBranch(branchID, thought)
}

func (s *SQLiteStorage) AppendInsightToBranch(branchID string, insight *types.Insight) error {
	// Update database to set branch_id for this insight
	_, err := s.db.Exec(`
		UPDATE insights SET branch_id = ? WHERE id = ?
	`, branchID, insight.ID)

	if err != nil {
		return fmt.Errorf("failed to update insight branch_id: %w", err)
	}

	// Update cache
	return s.cache.AppendInsightToBranch(branchID, insight)
}

func (s *SQLiteStorage) AppendCrossRefToBranch(branchID string, crossRef *types.CrossRef) error {
	// Cross-ref should already be persisted via storeBranch or directly
	// For now, just update cache (cross-refs are stored when branches are stored)
	return s.cache.AppendCrossRefToBranch(branchID, crossRef)
}

func (s *SQLiteStorage) UpdateBranchPriority(branchID string, priority float64) error {
	// Update database
	_, err := s.db.Exec(`
		UPDATE branches SET priority = ?, updated_at = ? WHERE id = ?
	`, priority, time.Now().Unix(), branchID)

	if err != nil {
		return fmt.Errorf("failed to update branch priority: %w", err)
	}

	// Update cache
	return s.cache.UpdateBranchPriority(branchID, priority)
}

func (s *SQLiteStorage) UpdateBranchConfidence(branchID string, confidence float64) error {
	// Update database
	_, err := s.db.Exec(`
		UPDATE branches SET confidence = ?, updated_at = ? WHERE id = ?
	`, confidence, time.Now().Unix(), branchID)

	if err != nil {
		return fmt.Errorf("failed to update branch confidence: %w", err)
	}

	// Update cache
	return s.cache.UpdateBranchConfidence(branchID, confidence)
}

func (s *SQLiteStorage) GetRecentBranches() ([]*types.Branch, error) {
	// Query database for recently accessed branches
	rows, err := s.db.Query(`
		SELECT id, parent_branch_id, state, priority, confidence,
		       created_at, updated_at, last_accessed_at
		FROM branches
		ORDER BY last_accessed_at DESC
		LIMIT ?
	`, MaxRecentBranches)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent branches: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows.Err() will catch any real errors

	branches := make([]*types.Branch, 0)
	for rows.Next() {
		branch := &types.Branch{}
		var parentBranchID sql.NullString
		var createdAt, updatedAt, lastAccessedAt int64

		err := rows.Scan(
			&branch.ID, &parentBranchID, &branch.State, &branch.Priority, &branch.Confidence,
			&createdAt, &updatedAt, &lastAccessedAt,
		)
		if err != nil {
			log.Printf("Warning: failed to scan branch: %v", err)
			continue
		}

		if parentBranchID.Valid {
			branch.ParentBranchID = parentBranchID.String
		}
		branch.CreatedAt = time.Unix(createdAt, 0)
		branch.UpdatedAt = time.Unix(updatedAt, 0)
		branch.LastAccessedAt = time.Unix(lastAccessedAt, 0)

		// Load associated data
		thoughts, _ := s.loadBranchThoughts(branch.ID)
		branch.Thoughts = thoughts

		insights, _ := s.loadBranchInsights(branch.ID)
		branch.Insights = insights

		crossRefs, _ := s.loadBranchCrossRefs(branch.ID)
		branch.CrossRefs = crossRefs

		branches = append(branches, branch)
	}

	return branches, nil
}

func (s *SQLiteStorage) GetMetrics() *Metrics {
	return s.cache.GetMetrics()
}

// StoreEmbedding stores a vector embedding in the database
func (s *SQLiteStorage) StoreEmbedding(problemID string, embedding []float32, model, provider string, dimension int, source string) error {
	// Generate ID
	id := fmt.Sprintf("emb-%s-%d", problemID, time.Now().Unix())

	// Serialize embedding to bytes
	embeddingBytes := make([]byte, len(embedding)*4)
	for i, val := range embedding {
		bits := math.Float32bits(val)
		binary.LittleEndian.PutUint32(embeddingBytes[i*4:], bits)
	}

	now := time.Now().Unix()

	// Insert or update embedding
	_, err := s.db.Exec(`
		INSERT INTO embeddings (id, problem_id, embedding, model, provider, dimension, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(problem_id) DO UPDATE SET
			embedding = excluded.embedding,
			model = excluded.model,
			provider = excluded.provider,
			dimension = excluded.dimension,
			source = excluded.source,
			updated_at = excluded.updated_at
	`, id, problemID, embeddingBytes, model, provider, dimension, source, now, now)

	if err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	return nil
}

// GetEmbedding retrieves an embedding for a problem ID
func (s *SQLiteStorage) GetEmbedding(problemID string) ([]float32, string, string, int, error) {
	var embeddingBytes []byte
	var model, provider, source string
	var dimension int

	err := s.db.QueryRow(`
		SELECT embedding, model, provider, dimension, source
		FROM embeddings
		WHERE problem_id = ?
	`, problemID).Scan(&embeddingBytes, &model, &provider, &dimension, &source)

	if err == sql.ErrNoRows {
		return nil, "", "", 0, fmt.Errorf("embedding not found for problem: %s", problemID)
	}
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("failed to fetch embedding: %w", err)
	}

	// Deserialize embedding
	embedding := make([]float32, len(embeddingBytes)/4)
	for i := 0; i < len(embedding); i++ {
		bits := binary.LittleEndian.Uint32(embeddingBytes[i*4:])
		embedding[i] = math.Float32frombits(bits)
	}

	return embedding, model, provider, dimension, nil
}

// GetAllEmbeddings retrieves all embeddings from the database
func (s *SQLiteStorage) GetAllEmbeddings() (map[string][]float32, error) {
	rows, err := s.db.Query(`
		SELECT problem_id, embedding
		FROM embeddings
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query embeddings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	embeddings := make(map[string][]float32)

	for rows.Next() {
		var problemID string
		var embeddingBytes []byte

		if err := rows.Scan(&problemID, &embeddingBytes); err != nil {
			log.Printf("Warning: failed to scan embedding row: %v", err)
			continue
		}

		// Deserialize embedding
		embedding := make([]float32, len(embeddingBytes)/4)
		for i := 0; i < len(embedding); i++ {
			bits := binary.LittleEndian.Uint32(embeddingBytes[i*4:])
			embedding[i] = math.Float32frombits(bits)
		}

		embeddings[problemID] = embedding
	}

	return embeddings, nil
}

// Close releases database resources
func (s *SQLiteStorage) Close() error {
	// Close prepared statements (ignore errors in cleanup)
	if s.stmtInsertThought != nil {
		_ = s.stmtInsertThought.Close()
	}
	if s.stmtGetThought != nil {
		_ = s.stmtGetThought.Close()
	}
	if s.stmtSearchFTS != nil {
		_ = s.stmtSearchFTS.Close()
	}
	if s.stmtInsertBranch != nil {
		_ = s.stmtInsertBranch.Close()
	}
	if s.stmtGetBranch != nil {
		_ = s.stmtGetBranch.Close()
	}
	if s.stmtUpdateBranchAccess != nil {
		_ = s.stmtUpdateBranchAccess.Close()
	}
	if s.stmtInsertInsight != nil {
		_ = s.stmtInsertInsight.Close()
	}
	if s.stmtInsertCrossRef != nil {
		_ = s.stmtInsertCrossRef.Close()
	}
	if s.stmtInsertValidation != nil {
		_ = s.stmtInsertValidation.Close()
	}

	// Close database
	return s.db.Close()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
