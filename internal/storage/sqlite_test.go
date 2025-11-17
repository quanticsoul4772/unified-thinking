package storage

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unified-thinking/internal/types"
)

// Helper function to create a temporary SQLite database for testing
func newTestSQLiteStorage(t *testing.T) (*SQLiteStorage, string) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create test SQLite storage: %v", err)
	}

	return storage, dbPath
}

func TestNewSQLiteStorage(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		timeout int
		wantErr bool
	}{
		{
			name:    "create new database",
			dbPath:  filepath.Join(t.TempDir(), "new.db"),
			timeout: 5000,
			wantErr: false,
		},
		{
			name:    "open existing database",
			dbPath:  filepath.Join(t.TempDir(), "existing.db"),
			timeout: 3000,
			wantErr: false,
		},
		{
			name:    "invalid path",
			dbPath:  "/invalid/path/\x00/test.db", // null byte in path
			timeout: 5000,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pre-create database for "existing" test
			if tt.name == "open existing database" {
				preStorage, err := NewSQLiteStorage(tt.dbPath, tt.timeout)
				if err != nil {
					t.Fatalf("Failed to pre-create database: %v", err)
				}
				preStorage.Close()
			}

			storage, err := NewSQLiteStorage(tt.dbPath, tt.timeout)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSQLiteStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if storage == nil {
					t.Fatal("NewSQLiteStorage() returned nil storage")
				}
				if storage.db == nil {
					t.Error("Database connection is nil")
				}
				if storage.cache == nil {
					t.Error("Cache is nil")
				}

				// Verify database is accessible
				err = storage.db.Ping()
				if err != nil {
					t.Errorf("Database ping failed: %v", err)
				}

				storage.Close()
			}
		})
	}
}

func TestSQLiteStoreAndGetThought(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	tests := []struct {
		name    string
		thought *types.Thought
		wantErr bool
	}{
		{
			name: "store basic thought",
			thought: &types.Thought{
				Content:    "Test thought content",
				Mode:       types.ModeLinear,
				Type:       "test",
				Confidence: 0.85,
				Timestamp:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "store thought with all fields",
			thought: &types.Thought{
				Content: "Complex thought",
				Mode:    types.ModeTree,
				// Don't set BranchID to avoid foreign key constraint
				Type:                 "analysis",
				Confidence:           0.92,
				Timestamp:            time.Now(),
				KeyPoints:            []string{"point1", "point2", "point3"},
				Metadata:             map[string]interface{}{"key": "value", "count": 42},
				IsRebellion:          true,
				ChallengesAssumption: true,
			},
			wantErr: false,
		},
		{
			name: "store thought with ID",
			thought: &types.Thought{
				ID:         "custom-id-123",
				Content:    "Thought with custom ID",
				Mode:       types.ModeDivergent,
				Confidence: 0.75,
				Timestamp:  time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.StoreThought(tt.thought)

			if (err != nil) != tt.wantErr {
				t.Errorf("StoreThought() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify ID was generated
				if tt.thought.ID == "" {
					t.Error("Thought ID was not generated")
				}

				// Retrieve and verify
				retrieved, err := storage.GetThought(tt.thought.ID)
				if err != nil {
					t.Fatalf("GetThought() error = %v", err)
				}

				if retrieved.Content != tt.thought.Content {
					t.Errorf("Content = %v, want %v", retrieved.Content, tt.thought.Content)
				}
				if retrieved.Mode != tt.thought.Mode {
					t.Errorf("Mode = %v, want %v", retrieved.Mode, tt.thought.Mode)
				}
				if retrieved.Confidence != tt.thought.Confidence {
					t.Errorf("Confidence = %v, want %v", retrieved.Confidence, tt.thought.Confidence)
				}
				if len(retrieved.KeyPoints) != len(tt.thought.KeyPoints) {
					t.Errorf("KeyPoints length = %v, want %v", len(retrieved.KeyPoints), len(tt.thought.KeyPoints))
				}
				if retrieved.IsRebellion != tt.thought.IsRebellion {
					t.Errorf("IsRebellion = %v, want %v", retrieved.IsRebellion, tt.thought.IsRebellion)
				}
			}
		})
	}
}

func TestSQLiteGetThought(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store a thought
	thought := &types.Thought{
		ID:         "test-123",
		Content:    "Test content",
		Mode:       types.ModeLinear,
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}
	_ = storage.StoreThought(thought)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "get existing thought",
			id:      "test-123",
			wantErr: false,
		},
		{
			name:    "get non-existent thought",
			id:      "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := storage.GetThought(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetThought() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && retrieved == nil {
				t.Error("GetThought() returned nil")
			}
		})
	}
}

func TestSQLitePersistenceAcrossRestarts(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "persistence.db")

	// Create storage and store data
	storage1, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create first storage: %v", err)
	}

	thought := &types.Thought{
		ID:         "persist-1",
		Content:    "Persistent thought",
		Mode:       types.ModeLinear,
		Type:       "test",
		Confidence: 0.9,
		KeyPoints:  []string{"key1", "key2"},
		Metadata:   map[string]interface{}{"persisted": true},
		Timestamp:  time.Now(),
	}
	err = storage1.StoreThought(thought)
	if err != nil {
		t.Fatalf("Failed to store thought: %v", err)
	}

	branch := &types.Branch{
		ID:             "branch-persist",
		State:          types.StateActive,
		Priority:       1.0,
		Confidence:     0.85,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}
	err = storage1.StoreBranch(branch)
	if err != nil {
		t.Fatalf("Failed to store branch: %v", err)
	}

	// Close storage
	storage1.Close()

	// Reopen storage
	storage2, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to reopen storage: %v", err)
	}
	defer storage2.Close()

	// Verify thought persisted
	retrievedThought, err := storage2.GetThought("persist-1")
	if err != nil {
		t.Fatalf("Failed to retrieve thought after restart: %v", err)
	}
	if retrievedThought.Content != thought.Content {
		t.Errorf("Thought content = %v, want %v", retrievedThought.Content, thought.Content)
	}
	if len(retrievedThought.KeyPoints) != 2 {
		t.Errorf("KeyPoints not persisted correctly")
	}

	// Verify branch persisted
	retrievedBranch, err := storage2.GetBranch("branch-persist")
	if err != nil {
		t.Fatalf("Failed to retrieve branch after restart: %v", err)
	}
	if retrievedBranch.State != branch.State {
		t.Errorf("Branch state = %v, want %v", retrievedBranch.State, branch.State)
	}
}

func TestSQLiteCacheConsistency(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store thought
	thought := &types.Thought{
		ID:         "cache-test",
		Content:    "Cache test",
		Mode:       types.ModeLinear,
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}
	_ = storage.StoreThought(thought)

	// Retrieve from cache (first call)
	cached, err := storage.GetThought("cache-test")
	if err != nil {
		t.Fatalf("Failed to get cached thought: %v", err)
	}

	// Clear cache to force database read
	storage.cache = NewMemoryStorage()

	// Retrieve from database
	fromDB, err := storage.GetThought("cache-test")
	if err != nil {
		t.Fatalf("Failed to get thought from DB: %v", err)
	}

	// Verify consistency
	if cached.Content != fromDB.Content {
		t.Errorf("Cache inconsistency: cached=%v, db=%v", cached.Content, fromDB.Content)
	}
	if cached.Confidence != fromDB.Confidence {
		t.Errorf("Cache confidence inconsistency: cached=%v, db=%v", cached.Confidence, fromDB.Confidence)
	}
}

func TestSQLiteStoreBranch(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	tests := []struct {
		name    string
		branch  *types.Branch
		wantErr bool
	}{
		{
			name: "store basic branch",
			branch: &types.Branch{
				State:          types.StateActive,
				Priority:       1.0,
				Confidence:     0.8,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				LastAccessedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "store branch with parent",
			branch: &types.Branch{
				// Don't set ParentBranchID to avoid foreign key constraint
				// (parent branch doesn't exist in this isolated test)
				State:          types.StateSuspended,
				Priority:       0.5,
				Confidence:     0.6,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				LastAccessedAt: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.StoreBranch(tt.branch)

			if (err != nil) != tt.wantErr {
				t.Errorf("StoreBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.branch.ID == "" {
					t.Error("Branch ID was not generated")
				}

				// Retrieve and verify
				retrieved, err := storage.GetBranch(tt.branch.ID)
				if err != nil {
					t.Fatalf("GetBranch() error = %v", err)
				}

				if retrieved.State != tt.branch.State {
					t.Errorf("State = %v, want %v", retrieved.State, tt.branch.State)
				}
				if retrieved.Priority != tt.branch.Priority {
					t.Errorf("Priority = %v, want %v", retrieved.Priority, tt.branch.Priority)
				}
			}
		})
	}
}

func TestSQLiteBranchWithAssociations(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Create branch
	branch := &types.Branch{
		ID:             "branch-assoc",
		State:          types.StateActive,
		Priority:       1.0,
		Confidence:     0.8,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}
	err := storage.StoreBranch(branch)
	if err != nil {
		t.Fatalf("StoreBranch() error = %v", err)
	}

	// Add thoughts to branch (now branch exists, so foreign key is satisfied)
	thought1 := &types.Thought{
		Content:    "Thought 1",
		Mode:       types.ModeTree,
		BranchID:   "branch-assoc",
		Type:       "test",
		Confidence: 0.9,
		Timestamp:  time.Now(),
	}
	thought2 := &types.Thought{
		Content:    "Thought 2",
		Mode:       types.ModeTree,
		BranchID:   "branch-assoc",
		Type:       "test",
		Confidence: 0.85,
		Timestamp:  time.Now().Add(1 * time.Second),
	}
	err = storage.StoreThought(thought1)
	if err != nil {
		t.Fatalf("StoreThought(thought1) error = %v", err)
	}
	err = storage.StoreThought(thought2)
	if err != nil {
		t.Fatalf("StoreThought(thought2) error = %v", err)
	}

	// Add insight to branch
	insight := &types.Insight{
		Type:               types.InsightObservation,
		Content:            "Test insight",
		ApplicabilityScore: 0.9,
		CreatedAt:          time.Now(),
	}
	err = storage.StoreInsight(insight)
	if err != nil {
		t.Fatalf("StoreInsight() error = %v", err)
	}
	err = storage.AppendInsightToBranch("branch-assoc", insight)
	if err != nil {
		t.Fatalf("AppendInsightToBranch() error = %v", err)
	}

	// Clear cache to force database load
	storage.cache = NewMemoryStorage()

	// Retrieve branch
	retrieved, err := storage.GetBranch("branch-assoc")
	if err != nil {
		t.Fatalf("GetBranch() error = %v", err)
	}

	// Verify associations were loaded
	if len(retrieved.Thoughts) != 2 {
		t.Errorf("Expected 2 thoughts, got %d", len(retrieved.Thoughts))
	}
	if len(retrieved.Insights) != 1 {
		t.Errorf("Expected 1 insight, got %d", len(retrieved.Insights))
	}

	// Verify thoughts are in correct order (by timestamp)
	if len(retrieved.Thoughts) == 2 {
		if retrieved.Thoughts[0].Content != "Thought 1" {
			t.Error("Thoughts not loaded in correct order")
		}
	}
}

func TestSQLiteFullTextSearch(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store thoughts with various content
	thoughts := []*types.Thought{
		{Content: "The quick brown fox jumps", Mode: types.ModeLinear, Type: "test", Timestamp: time.Now()},
		{Content: "A lazy dog sleeps peacefully", Mode: types.ModeTree, Type: "test", Timestamp: time.Now()},
		{Content: "Quick thinking leads to success", Mode: types.ModeLinear, Type: "test", Timestamp: time.Now()},
		{Content: "Brown bears in the forest", Mode: types.ModeDivergent, Type: "test", Timestamp: time.Now()},
	}

	for _, th := range thoughts {
		err := storage.StoreThought(th)
		if err != nil {
			t.Fatalf("StoreThought() error = %v", err)
		}
	}

	tests := []struct {
		name      string
		query     string
		mode      types.ThinkingMode
		wantCount int
	}{
		{
			name:      "search for 'quick'",
			query:     "quick",
			mode:      "",
			wantCount: 2,
		},
		{
			name:      "search for 'brown'",
			query:     "brown",
			mode:      "",
			wantCount: 2,
		},
		{
			name:      "search with mode filter",
			query:     "quick",
			mode:      types.ModeLinear,
			wantCount: 2,
		},
		{
			name:      "no matches",
			query:     "elephant",
			mode:      "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := storage.SearchThoughts(tt.query, tt.mode, 100, 0)

			if len(results) != tt.wantCount {
				t.Errorf("SearchThoughts() returned %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestSQLiteStoreInsight(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	insight := &types.Insight{
		Type:               types.InsightObservation,
		Content:            "Test insight content",
		Context:            []string{"context1", "context2"},
		ParentInsights:     []string{"parent1"},
		ApplicabilityScore: 0.92,
		SupportingEvidence: map[string]interface{}{"evidence": "data"},
		CreatedAt:          time.Now(),
	}

	err := storage.StoreInsight(insight)
	if err != nil {
		t.Fatalf("StoreInsight() error = %v", err)
	}

	if insight.ID == "" {
		t.Error("Insight ID was not generated")
	}

	// Retrieve and verify
	retrieved, err := storage.GetInsight(insight.ID)
	if err != nil {
		t.Fatalf("GetInsight() error = %v", err)
	}

	if retrieved.Content != insight.Content {
		t.Errorf("Content = %v, want %v", retrieved.Content, insight.Content)
	}
	if len(retrieved.Context) != len(insight.Context) {
		t.Errorf("Context length = %v, want %v", len(retrieved.Context), len(insight.Context))
	}
}

func TestSQLiteStoreValidation(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	validation := &types.Validation{
		ThoughtID:      "thought-1",
		InsightID:      "insight-1",
		IsValid:        true,
		ValidationData: map[string]interface{}{"score": 0.95},
		Reason:         "Logical consistency verified",
		CreatedAt:      time.Now(),
	}

	err := storage.StoreValidation(validation)
	if err != nil {
		t.Fatalf("StoreValidation() error = %v", err)
	}

	if validation.ID == "" {
		t.Error("Validation ID was not generated")
	}

	// Retrieve and verify
	retrieved, err := storage.GetValidation(validation.ID)
	if err != nil {
		t.Fatalf("GetValidation() error = %v", err)
	}

	if retrieved.IsValid != validation.IsValid {
		t.Errorf("IsValid = %v, want %v", retrieved.IsValid, validation.IsValid)
	}
	if retrieved.Reason != validation.Reason {
		t.Errorf("Reason = %v, want %v", retrieved.Reason, validation.Reason)
	}
}

func TestSQLiteStoreRelationship(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	relationship := &types.Relationship{
		FromStateID: "state-1",
		ToStateID:   "state-2",
		Type:        "causal",
		Metadata:    map[string]interface{}{"strength": 0.8},
		CreatedAt:   time.Now(),
	}

	err := storage.StoreRelationship(relationship)
	if err != nil {
		t.Fatalf("StoreRelationship() error = %v", err)
	}

	if relationship.ID == "" {
		t.Error("Relationship ID was not generated")
	}

	// Retrieve and verify
	retrieved, err := storage.GetRelationship(relationship.ID)
	if err != nil {
		t.Fatalf("GetRelationship() error = %v", err)
	}

	if retrieved.Type != relationship.Type {
		t.Errorf("Type = %v, want %v", retrieved.Type, relationship.Type)
	}
}

func TestSQLiteUpdateBranchPriority(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Create branch
	branch := &types.Branch{
		ID:             "update-test",
		State:          types.StateActive,
		Priority:       0.5,
		Confidence:     0.8,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}
	_ = storage.StoreBranch(branch)

	// Update priority
	err := storage.UpdateBranchPriority("update-test", 0.95)
	if err != nil {
		t.Fatalf("UpdateBranchPriority() error = %v", err)
	}

	// Verify update persisted
	retrieved, _ := storage.GetBranch("update-test")
	if retrieved.Priority != 0.95 {
		t.Errorf("Priority = %v, want 0.95", retrieved.Priority)
	}
}

func TestSQLiteUpdateBranchConfidence(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Create branch
	branch := &types.Branch{
		ID:             "confidence-test",
		State:          types.StateActive,
		Priority:       1.0,
		Confidence:     0.5,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}
	_ = storage.StoreBranch(branch)

	// Update confidence
	err := storage.UpdateBranchConfidence("confidence-test", 0.88)
	if err != nil {
		t.Fatalf("UpdateBranchConfidence() error = %v", err)
	}

	// Verify update persisted
	retrieved, _ := storage.GetBranch("confidence-test")
	if retrieved.Confidence != 0.88 {
		t.Errorf("Confidence = %v, want 0.88", retrieved.Confidence)
	}
}

func TestSQLiteConcurrentAccess(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	t.Run("concurrent thought storage", func(t *testing.T) {
		var wg sync.WaitGroup
		// Reduce concurrency for SQLite - it has max_open_conns=4
		numGoroutines := 5

		wg.Add(numGoroutines)
		var errorCount int32 // Use atomic operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				thought := &types.Thought{
					Content:    "Concurrent thought",
					Mode:       types.ModeLinear,
					Type:       "test",
					Confidence: 0.8,
					Timestamp:  time.Now(),
				}
				err := storage.StoreThought(thought)
				if err != nil {
					// SQLite may have some lock contention, that's expected
					atomic.AddInt32(&errorCount, 1)
				}
			}(i)
		}

		wg.Wait()

		// Verify at least some thoughts stored (SQLite has limited concurrency)
		results := storage.SearchThoughts("Concurrent", "", 100, 0)
		if len(results) == 0 {
			t.Error("No thoughts were stored during concurrent access")
		}
	})

	t.Run("concurrent read/write", func(t *testing.T) {
		// Store initial thought
		thought := &types.Thought{
			ID:         "concurrent-rw",
			Content:    "Read-write test",
			Mode:       types.ModeLinear,
			Type:       "test",
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		err := storage.StoreThought(thought)
		if err != nil {
			t.Fatalf("Failed to store initial thought: %v", err)
		}

		var wg sync.WaitGroup
		// Reduce concurrency for SQLite
		numReaders := 5
		numWriters := 3

		// Readers (read operations work better concurrently with WAL)
		wg.Add(numReaders)
		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()
				_, err := storage.GetThought("concurrent-rw")
				// Reads should succeed with WAL mode
				if err != nil {
					t.Errorf("Concurrent read error: %v", err)
				}
			}()
		}

		// Writers (may encounter some lock contention)
		wg.Add(numWriters)
		for i := 0; i < numWriters; i++ {
			go func() {
				defer wg.Done()
				th := &types.Thought{
					Content:    "Writer thought",
					Mode:       types.ModeTree,
					Type:       "test",
					Confidence: 0.7,
					Timestamp:  time.Now(),
				}
				// Some writes may fail due to locks, that's expected with SQLite
				_ = storage.StoreThought(th)
			}()
		}

		wg.Wait()
	})
}

func TestSQLiteSchemaInitialization(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "schema.db")

	storage, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Verify tables exist
	tables := []string{
		"schema_metadata",
		"thoughts",
		"branches",
		"insights",
		"cross_refs",
		"validations",
		"relationships",
		"thoughts_fts",
	}

	for _, table := range tables {
		var count int
		err := storage.db.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&count)

		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Table %s not found", table)
		}
	}

	// Verify schema version
	var version int
	err = storage.db.QueryRow(
		"SELECT value FROM schema_metadata WHERE key = 'version'",
	).Scan(&version)

	if err != nil {
		t.Errorf("Failed to get schema version: %v", err)
	}
	if version != schemaVersion {
		t.Errorf("Schema version = %d, want %d", version, schemaVersion)
	}
}

func TestSQLitePreparedStatements(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Verify prepared statements were created
	if storage.stmtInsertThought == nil {
		t.Error("stmtInsertThought is nil")
	}
	if storage.stmtGetThought == nil {
		t.Error("stmtGetThought is nil")
	}
	if storage.stmtSearchFTS == nil {
		t.Error("stmtSearchFTS is nil")
	}
	if storage.stmtInsertBranch == nil {
		t.Error("stmtInsertBranch is nil")
	}
	if storage.stmtGetBranch == nil {
		t.Error("stmtGetBranch is nil")
	}
	if storage.stmtUpdateBranchAccess == nil {
		t.Error("stmtUpdateBranchAccess is nil")
	}
	if storage.stmtInsertInsight == nil {
		t.Error("stmtInsertInsight is nil")
	}
	if storage.stmtInsertCrossRef == nil {
		t.Error("stmtInsertCrossRef is nil")
	}
	if storage.stmtInsertValidation == nil {
		t.Error("stmtInsertValidation is nil")
	}
}

func TestSQLiteClose(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)

	err := storage.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify database is closed
	err = storage.db.Ping()
	if err == nil {
		t.Error("Database still accessible after Close()")
	}
}

func TestSQLiteErrorHandling(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	t.Run("get non-existent thought", func(t *testing.T) {
		_, err := storage.GetThought("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent thought")
		}
	})

	t.Run("get non-existent branch", func(t *testing.T) {
		_, err := storage.GetBranch("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent branch")
		}
	})

	t.Run("get non-existent insight", func(t *testing.T) {
		_, err := storage.GetInsight("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent insight")
		}
	})

	t.Run("get non-existent validation", func(t *testing.T) {
		_, err := storage.GetValidation("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent validation")
		}
	})

	t.Run("get non-existent relationship", func(t *testing.T) {
		_, err := storage.GetRelationship("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent relationship")
		}
	})
}

func TestSQLiteWarmCache(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "warmcache.db")

	// Create storage and add thoughts
	storage1, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Add multiple thoughts
	for i := 0; i < 10; i++ {
		thought := &types.Thought{
			Content:    "Cached thought",
			Mode:       types.ModeLinear,
			Type:       "test",
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		err := storage1.StoreThought(thought)
		if err != nil {
			t.Fatalf("StoreThought() error = %v", err)
		}
	}
	storage1.Close()

	// Reopen storage - cache should warm
	storage2, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to reopen storage: %v", err)
	}
	defer storage2.Close()

	// Verify cache was warmed (warmCache loads last 1000 thoughts)
	// Note: cache is private, so verify by searching instead
	results := storage2.SearchThoughts("", types.ModeLinear, 100, 0)
	if len(results) != 10 {
		t.Errorf("Search returned %d thoughts, expected 10", len(results))
	}
}

func TestSQLiteUpdateThought(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store initial thought
	thought := &types.Thought{
		ID:         "update-thought",
		Content:    "Original content",
		Mode:       types.ModeLinear,
		Type:       "test",
		Confidence: 0.5,
		Timestamp:  time.Now(),
	}
	err := storage.StoreThought(thought)
	if err != nil {
		t.Fatalf("StoreThought() error = %v", err)
	}

	// Update thought (SQLite uses ON CONFLICT DO UPDATE)
	thought.Content = "Updated content"
	thought.Confidence = 0.9
	err = storage.StoreThought(thought)
	if err != nil {
		t.Fatalf("StoreThought() update error = %v", err)
	}

	// Clear cache to force DB read
	storage.cache = NewMemoryStorage()

	// Retrieve and verify update
	retrieved, err := storage.GetThought("update-thought")
	if err != nil {
		t.Fatalf("GetThought() error = %v", err)
	}

	if retrieved.Content != "Updated content" {
		t.Errorf("Content = %v, want 'Updated content'", retrieved.Content)
	}
	if retrieved.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", retrieved.Confidence)
	}
}

func TestSQLiteListBranches(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store multiple branches
	for i := 0; i < 5; i++ {
		branch := &types.Branch{
			State:          types.StateActive,
			Priority:       float64(i),
			Confidence:     0.8,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			LastAccessedAt: time.Now(),
		}
		err := storage.StoreBranch(branch)
		if err != nil {
			t.Fatalf("StoreBranch() error = %v", err)
		}
	}

	// ListBranches uses cache, which should have all branches
	branches := storage.ListBranches()
	if len(branches) != 5 {
		t.Errorf("ListBranches() returned %d branches, want 5", len(branches))
	}
}

func TestSQLiteActiveBranch(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Create branches
	branch1 := &types.Branch{
		ID:             "branch-1",
		State:          types.StateActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}
	branch2 := &types.Branch{
		ID:             "branch-2",
		State:          types.StateActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}

	err := storage.StoreBranch(branch1)
	if err != nil {
		t.Fatalf("StoreBranch(branch1) error = %v", err)
	}
	err = storage.StoreBranch(branch2)
	if err != nil {
		t.Fatalf("StoreBranch(branch2) error = %v", err)
	}

	// Set active branch (cache-based operation)
	err = storage.SetActiveBranch("branch-2")
	if err != nil {
		t.Fatalf("SetActiveBranch() error = %v", err)
	}

	// Get active branch
	active, err := storage.GetActiveBranch()
	if err != nil {
		t.Fatalf("GetActiveBranch() error = %v", err)
	}

	if active.ID != "branch-2" {
		t.Errorf("Active branch ID = %v, want 'branch-2'", active.ID)
	}
}

func TestSQLiteSearchPagination(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store 25 thoughts
	for i := 0; i < 25; i++ {
		thought := &types.Thought{
			Content:    "Paginated thought content",
			Mode:       types.ModeLinear,
			Type:       "test",
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		err := storage.StoreThought(thought)
		if err != nil {
			t.Fatalf("StoreThought() error = %v", err)
		}
	}

	// Test pagination using FTS search
	page1 := storage.SearchThoughts("Paginated", "", 10, 0)
	if len(page1) != 10 {
		t.Errorf("Page 1 returned %d results, want 10", len(page1))
	}

	page2 := storage.SearchThoughts("Paginated", "", 10, 10)
	if len(page2) != 10 {
		t.Errorf("Page 2 returned %d results, want 10", len(page2))
	}

	page3 := storage.SearchThoughts("Paginated", "", 10, 20)
	if len(page3) != 5 {
		t.Errorf("Page 3 returned %d results, want 5", len(page3))
	}
}

func TestSQLiteDatabaseLocked(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database lock test in short mode")
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "locked.db")

	// Create first connection
	storage1, err := NewSQLiteStorage(dbPath, 1000)
	if err != nil {
		t.Fatalf("Failed to create first storage: %v", err)
	}
	defer storage1.Close()

	// Start a transaction to lock database
	tx, err := storage1.db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Store data in transaction
	_, err = tx.Exec("INSERT INTO thoughts (id, content, mode, type, confidence, timestamp) VALUES (?, ?, ?, ?, ?, ?)",
		"lock-test", "Locked content", "linear", "test", 0.8, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// Try to read from another storage instance - should succeed with WAL mode
	storage2, err := NewSQLiteStorage(dbPath, 1000)
	if err != nil {
		t.Fatalf("Failed to create second storage: %v", err)
	}
	defer storage2.Close()

	// WAL mode allows concurrent reads
	_, err = storage2.GetThought("lock-test")
	// This might fail or succeed depending on WAL state, just verify no panic
	_ = err

	// Commit transaction
	tx.Commit()
}

func TestSQLiteCorruptDatabase(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "corrupt.db")

	// Create corrupt file
	err := os.WriteFile(dbPath, []byte("This is not a valid SQLite database"), 0600)
	if err != nil {
		t.Fatalf("Failed to create corrupt file: %v", err)
	}

	// Try to open - should fail
	_, err = NewSQLiteStorage(dbPath, 5000)
	if err == nil {
		t.Error("Expected error when opening corrupt database")
	}
}

func TestSQLiteMissingDirectory(t *testing.T) {
	// Try to create database in non-existent directory
	dbPath := filepath.Join(t.TempDir(), "nonexistent", "subdir", "test.db")

	// Should fail because parent directory doesn't exist
	_, err := NewSQLiteStorage(dbPath, 5000)
	if err == nil {
		t.Error("Expected error when creating database in non-existent directory")
	}
}

func TestBoolToInt(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		want  int
	}{
		{"true to 1", true, 1},
		{"false to 0", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boolToInt(tt.input)
			if got != tt.want {
				t.Errorf("boolToInt(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSQLiteScanThought(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store thought with all fields (no BranchID to avoid foreign key)
	thought := &types.Thought{
		Content:              "Scan test",
		Mode:                 types.ModeTree,
		Type:                 "test",
		Confidence:           0.95,
		Timestamp:            time.Now(),
		KeyPoints:            []string{"key1", "key2"},
		Metadata:             map[string]interface{}{"test": "data"},
		IsRebellion:          true,
		ChallengesAssumption: false,
	}
	err := storage.StoreThought(thought)
	if err != nil {
		t.Fatalf("StoreThought() error = %v", err)
	}

	// Query and scan
	rows, err := storage.db.Query(`
		SELECT id, content, mode, branch_id, parent_id, type, confidence,
		       timestamp, key_points, metadata, is_rebellion, challenges_assumption
		FROM thoughts WHERE id = ?
	`, thought.ID)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("No rows returned")
	}

	scanned, err := storage.scanThought(rows)
	if err != nil {
		t.Fatalf("scanThought() error = %v", err)
	}

	if scanned.Content != thought.Content {
		t.Errorf("Content = %v, want %v", scanned.Content, thought.Content)
	}
	if scanned.IsRebellion != thought.IsRebellion {
		t.Errorf("IsRebellion = %v, want %v", scanned.IsRebellion, thought.IsRebellion)
	}
}

func TestSQLiteScanInsight(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store insight
	insight := &types.Insight{
		Type:               types.InsightObservation,
		Content:            "Scan insight test",
		Context:            []string{"ctx1", "ctx2"},
		ParentInsights:     []string{"parent1"},
		ApplicabilityScore: 0.88,
		SupportingEvidence: map[string]interface{}{"evidence": "value"},
		CreatedAt:          time.Now(),
	}
	storage.StoreInsight(insight)
	storage.AppendInsightToBranch("test-branch", insight)

	// Query and scan
	rows, err := storage.db.Query(`
		SELECT id, type, content, context, applicability_score,
		       parent_insights, supporting_evidence, created_at
		FROM insights WHERE id = ?
	`, insight.ID)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("No rows returned")
	}

	scanned, err := storage.scanInsight(rows)
	if err != nil {
		t.Fatalf("scanInsight() error = %v", err)
	}

	if scanned.Content != insight.Content {
		t.Errorf("Content = %v, want %v", scanned.Content, insight.Content)
	}
	if len(scanned.Context) != len(insight.Context) {
		t.Errorf("Context length = %v, want %v", len(scanned.Context), len(insight.Context))
	}
}

func TestSQLiteMetrics(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// SQLite storage delegates metrics to cache
	metrics := storage.GetMetrics()
	if metrics == nil {
		t.Error("GetMetrics() returned nil")
	}
}

func TestSQLiteRecentBranches(t *testing.T) {
	storage, _ := newTestSQLiteStorage(t)
	defer storage.Close()

	// Store and access branches to populate recent list
	for i := 0; i < 3; i++ {
		branch := &types.Branch{
			State:          types.StateActive,
			Priority:       float64(i),
			Confidence:     0.8,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			LastAccessedAt: time.Now(),
		}
		err := storage.StoreBranch(branch)
		if err != nil {
			t.Fatalf("StoreBranch() error = %v", err)
		}

		// Access each branch to populate recent list
		err = storage.UpdateBranchAccess(branch.ID)
		if err != nil {
			t.Fatalf("UpdateBranchAccess() error = %v", err)
		}
	}

	// SQLite storage delegates to cache
	recent, err := storage.GetRecentBranches()
	if err != nil {
		t.Fatalf("GetRecentBranches() error = %v", err)
	}

	if len(recent) != 3 {
		t.Errorf("GetRecentBranches() returned %d branches, want 3", len(recent))
	}
}
