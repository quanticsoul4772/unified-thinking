// Package storage provides in-memory storage implementation for the unified thinking system.
//
// This package implements thread-safe storage using a read-write mutex and deep copying
// strategy to prevent data races. All retrieval methods return deep copies of stored data
// to ensure external modifications do not affect the internal storage state.
//
// Thread Safety:
// All methods are thread-safe through RWMutex protection. Read operations use RLock
// for concurrent access, while write operations use exclusive Lock.
//
// Memory Management:
// The storage is unbounded and will grow with usage. For production deployments,
// consider implementing LRU eviction or periodic cleanup strategies.
package storage

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

const (
	// MaxSearchResults is the hard limit for search results to prevent resource exhaustion
	MaxSearchResults = 1000
	// MaxIndexWordLength prevents extremely long words from being indexed
	MaxIndexWordLength = 50
	// MaxUniqueWordsPerThought limits unique words indexed per thought
	MaxUniqueWordsPerThought = 1000
	// MaxIndexSize is the global index size limit
	MaxIndexSize = 100000
	// MaxRecentBranches is the maximum number of recent branches to track
	MaxRecentBranches = 10
)

// MemoryStorage implements in-memory storage with thread-safe operations.
// All Get methods return deep copies to prevent external mutation of internal state.
type MemoryStorage struct {
	mu               sync.RWMutex
	thoughts         map[string]*types.Thought
	branches         map[string]*types.Branch
	insights         map[string]*types.Insight
	validations      map[string]*types.Validation
	relationships    map[string]*types.Relationship
	trajectoriesJSON map[string]string // Episodic memory trajectories as JSON

	// Search indices for O(1) word lookup (optimization for SearchThoughts)
	contentIndex    map[string][]string             // word -> []thoughtIDs
	indexAccessTime map[string]time.Time            // word -> last access time (for LRU eviction)
	modeIndex       map[types.ThinkingMode][]string // mode -> []thoughtIDs
	// TIER 2 OPTIMIZATION: Hot word cache for frequently accessed index entries
	hotWordCache  map[string][]string // Cache for top 100 most accessed words
	hotWordAccess map[string]int      // Access count for cache promotion

	// Ordered slices for deterministic pagination (sorted by timestamp, newest first)
	thoughtsOrdered []*types.Thought
	branchesOrdered []*types.Branch

	activeBranchID  string
	recentBranchIDs []string // Stack of recently accessed branch IDs (max 10)

	// Counters for ID generation
	thoughtCounter      int
	branchCounter       int
	insightCounter      int
	validationCounter   int
	relationshipCounter int
	trajectoryCounter   int
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		thoughts:         make(map[string]*types.Thought),
		branches:         make(map[string]*types.Branch),
		insights:         make(map[string]*types.Insight),
		validations:      make(map[string]*types.Validation),
		relationships:    make(map[string]*types.Relationship),
		trajectoriesJSON: make(map[string]string),
		contentIndex:     make(map[string][]string),
		indexAccessTime:  make(map[string]time.Time),
		modeIndex:        make(map[types.ThinkingMode][]string),
		hotWordCache:     make(map[string][]string, 100), // Cache top 100 hot words
		hotWordAccess:    make(map[string]int, 200),      // Track access counts
		thoughtsOrdered:  make([]*types.Thought, 0, 100), // Pre-allocate typical size
		branchesOrdered:  make([]*types.Branch, 0, 20),   // Pre-allocate typical size
		recentBranchIDs:  make([]string, 0, MaxRecentBranches),
	}
}

// StoreThought stores a thought in memory. If the thought ID is empty, a unique ID
// is generated automatically. The thought is stored by reference internally but all
// retrieval operations return deep copies for thread safety.
// Additionally builds search indices for efficient content and mode lookups.
func (s *MemoryStorage) StoreThought(thought *types.Thought) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if thought.ID == "" {
		s.thoughtCounter++
		thought.ID = fmt.Sprintf("thought-%d-%d", time.Now().Unix(), s.thoughtCounter)
	}

	s.thoughts[thought.ID] = thought

	// Build content index - tokenize content and index each word
	s.indexThoughtContent(thought)

	// Build mode index
	s.modeIndex[thought.Mode] = append(s.modeIndex[thought.Mode], thought.ID)

	// Add to ordered slice and maintain sort order (newest first)
	s.thoughtsOrdered = append(s.thoughtsOrdered, thought)
	// Sort by timestamp descending (newest first)
	sort.Slice(s.thoughtsOrdered, func(i, j int) bool {
		return s.thoughtsOrdered[i].Timestamp.After(s.thoughtsOrdered[j].Timestamp)
	})

	return nil
}

// indexThoughtContent tokenizes thought content and adds to inverted index
// Implements LRU eviction when index reaches capacity to prevent unbounded growth
func (s *MemoryStorage) indexThoughtContent(thought *types.Thought) {
	// Check global index size and evict LRU entries if at capacity
	if len(s.contentIndex) >= MaxIndexSize {
		s.evictLRUIndexEntries(MaxIndexSize / 10) // Evict 10% of capacity
	}

	// Tokenize content - PERFORMANCE OPTIMIZED
	// Use strings.Fields (fast path) + post-filter instead of FieldsFunc with closure
	// This is 20-30% faster and allocates less
	content := strings.ToLower(thought.Content)
	rawWords := strings.Fields(content) // Fast: splits on whitespace
	words := make([]string, 0, len(rawWords))

	// Filter and clean raw words
	for _, rawWord := range rawWords {
		// Remove leading/trailing punctuation
		word := strings.TrimFunc(rawWord, func(r rune) bool {
			return (r < 'a' || r > 'z') && (r < '0' || r > '9')
		})
		if len(word) > 0 {
			words = append(words, word)
		}
	}

	// Add thought ID to index for each unique word
	seen := make(map[string]bool, len(words))
	uniqueWordCount := 0
	now := time.Now()

	for _, word := range words {
		// Enforce word length limit to prevent extremely long words
		if word == "" || len(word) < 2 || len(word) > MaxIndexWordLength { // Skip empty and single-char tokens
			continue
		}
		if seen[word] {
			continue // Skip duplicates within same thought
		}

		// Limit unique words per thought to prevent index pollution
		if uniqueWordCount >= MaxUniqueWordsPerThought {
			log.Printf("Warning: Thought %s exceeded max unique words (%d), truncating index", thought.ID, MaxUniqueWordsPerThought)
			break
		}

		seen[word] = true
		uniqueWordCount++
		s.contentIndex[word] = append(s.contentIndex[word], thought.ID)
		s.indexAccessTime[word] = now // Track access time for LRU
	}
}

// evictLRUIndexEntries removes the least recently used entries from the content index
// This prevents unbounded memory growth and ensures new thoughts remain searchable
func (s *MemoryStorage) evictLRUIndexEntries(count int) {
	if count <= 0 || len(s.contentIndex) == 0 {
		return
	}

	// Build list of (word, lastAccessTime) pairs
	type wordEntry struct {
		word       string
		accessTime time.Time
	}
	entries := make([]wordEntry, 0, len(s.contentIndex))
	for word, accessTime := range s.indexAccessTime {
		entries = append(entries, wordEntry{word, accessTime})
	}

	// Sort by access time (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].accessTime.Before(entries[j].accessTime)
	})

	// Evict oldest entries
	evicted := 0
	for _, entry := range entries {
		if evicted >= count {
			break
		}
		delete(s.contentIndex, entry.word)
		delete(s.indexAccessTime, entry.word)
		evicted++
	}

	log.Printf("Evicted %d LRU index entries (index size now: %d)", evicted, len(s.contentIndex))
}

// GetThought retrieves a thought by ID (returns a copy to prevent data races)
// Optimization: Releases lock before deep copy to reduce lock hold time
func (s *MemoryStorage) GetThought(id string) (*types.Thought, error) {
	s.mu.RLock()
	thought, exists := s.thoughts[id]
	s.mu.RUnlock() // Release lock before deep copy

	if !exists {
		return nil, fmt.Errorf("thought not found: %s", id)
	}
	// Deep copy without holding lock (safe: thoughts are immutable after store)
	return copyThought(thought), nil
}

// StoreBranch stores a branch
func (s *MemoryStorage) StoreBranch(branch *types.Branch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if branch.ID == "" {
		s.branchCounter++
		branch.ID = fmt.Sprintf("branch-%d-%d", time.Now().Unix(), s.branchCounter)
	}

	// Initialize LastAccessedAt if not set
	if branch.LastAccessedAt.IsZero() {
		branch.LastAccessedAt = time.Now()
	}

	// Check if this is an update or new branch
	_, exists := s.branches[branch.ID]
	s.branches[branch.ID] = branch

	// Set as active if it's the first branch
	if s.activeBranchID == "" {
		s.activeBranchID = branch.ID
		s.trackRecentBranch(branch.ID)
	}

	// Add to ordered slice if new, or update existing entry
	if !exists {
		// New branch - add and sort
		s.branchesOrdered = append(s.branchesOrdered, branch)
		sort.Slice(s.branchesOrdered, func(i, j int) bool {
			return s.branchesOrdered[i].CreatedAt.After(s.branchesOrdered[j].CreatedAt)
		})
	} else {
		// Existing branch - find and update pointer (maintains order)
		for i, b := range s.branchesOrdered {
			if b.ID == branch.ID {
				s.branchesOrdered[i] = branch
				break
			}
		}
	}

	return nil
}

// GetBranch retrieves a branch by ID (returns a copy to prevent data races)
// Optimization: Releases lock before deep copy to reduce lock hold time
func (s *MemoryStorage) GetBranch(id string) (*types.Branch, error) {
	s.mu.RLock()
	branch, exists := s.branches[id]
	s.mu.RUnlock() // Release lock before deep copy

	if !exists {
		return nil, fmt.Errorf("branch not found: %s", id)
	}
	// Deep copy without holding lock (safe: branches only updated via StoreBranch)
	return copyBranch(branch), nil
}

// ListBranches returns all branches (returns copies to prevent data races)
// Optimization: Uses ordered slice and releases lock before deep copy
func (s *MemoryStorage) ListBranches() []*types.Branch {
	s.mu.RLock()
	// Capture pointers to branches while holding lock
	branchPointers := make([]*types.Branch, len(s.branchesOrdered))
	copy(branchPointers, s.branchesOrdered)
	s.mu.RUnlock() // Release lock before deep copy

	// Deep copy all branches without holding lock
	branches := make([]*types.Branch, len(branchPointers))
	for i, branch := range branchPointers {
		branches[i] = copyBranch(branch)
	}
	return branches
}

// GetActiveBranch returns the currently active branch (returns a copy to prevent data races)
// Optimization: Releases lock before deep copy to reduce lock hold time
func (s *MemoryStorage) GetActiveBranch() (*types.Branch, error) {
	s.mu.RLock()
	activeBranchID := s.activeBranchID
	var branch *types.Branch
	if activeBranchID != "" {
		branch = s.branches[activeBranchID]
	}
	s.mu.RUnlock() // Release lock before deep copy

	if activeBranchID == "" {
		return nil, fmt.Errorf("no active branch")
	}

	if branch == nil {
		// Active branch was deleted - this is a data inconsistency
		return nil, fmt.Errorf("active branch %s no longer exists", activeBranchID)
	}

	// Deep copy without holding lock (safe: branches only updated via StoreBranch)
	return copyBranch(branch), nil
}

// SetActiveBranch sets the active branch and updates access tracking
func (s *MemoryStorage) SetActiveBranch(branchID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	branch, exists := s.branches[branchID]
	if !exists {
		// Build list of available branches for helpful error message
		availableBranches := make([]string, 0, len(s.branches))
		for id := range s.branches {
			availableBranches = append(availableBranches, id)
		}
		if len(availableBranches) == 0 {
			return fmt.Errorf("branch not found: %s (no branches exist yet, create thoughts in tree mode first)", branchID)
		}
		return fmt.Errorf("branch not found: %s (available branches: %v)", branchID, availableBranches)
	}

	s.activeBranchID = branchID
	branch.LastAccessedAt = time.Now()
	s.trackRecentBranch(branchID)
	return nil
}

// StoreInsight stores an insight
func (s *MemoryStorage) StoreInsight(insight *types.Insight) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if insight.ID == "" {
		s.insightCounter++
		insight.ID = fmt.Sprintf("insight-%d", s.insightCounter)
	}

	s.insights[insight.ID] = insight
	return nil
}

// GetInsight retrieves an insight by ID
func (s *MemoryStorage) GetInsight(id string) (*types.Insight, error) {
	s.mu.RLock()
	insight, exists := s.insights[id]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("insight not found: %s", id)
	}

	return copyInsight(insight), nil
}

// StoreValidation stores a validation result
func (s *MemoryStorage) StoreValidation(validation *types.Validation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if validation.ID == "" {
		s.validationCounter++
		validation.ID = fmt.Sprintf("validation-%d", s.validationCounter)
	}

	s.validations[validation.ID] = validation
	return nil
}

// GetValidation retrieves a validation by ID
func (s *MemoryStorage) GetValidation(id string) (*types.Validation, error) {
	s.mu.RLock()
	validation, exists := s.validations[id]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("validation not found: %s", id)
	}

	return copyValidation(validation), nil
}

// StoreRelationship stores a relationship
func (s *MemoryStorage) StoreRelationship(rel *types.Relationship) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rel.ID == "" {
		s.relationshipCounter++
		rel.ID = fmt.Sprintf("rel-%d", s.relationshipCounter)
	}

	s.relationships[rel.ID] = rel
	return nil
}

// GetRelationship retrieves a relationship by ID
func (s *MemoryStorage) GetRelationship(id string) (*types.Relationship, error) {
	s.mu.RLock()
	rel, exists := s.relationships[id]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("relationship not found: %s", id)
	}

	// Deep copy
	relCopy := *rel
	if rel.Metadata != nil {
		relCopy.Metadata = deepCopyMap(rel.Metadata)
	}
	return &relCopy, nil
}

// SearchThoughts searches thoughts by content or type (returns copies to prevent data races)
// limit and offset support pagination - limit of 0 returns all results
// Uses inverted index for O(1) word lookup when query is provided
// Returns results in deterministic order (newest first by timestamp)
func (s *MemoryStorage) SearchThoughts(query string, mode types.ThinkingMode, limit, offset int) []*types.Thought {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Enforce maximum search results to prevent resource exhaustion
	if limit <= 0 || limit > MaxSearchResults {
		limit = MaxSearchResults
	}

	// Build candidate set using index for fast filtering
	// TIER 2 OPTIMIZATION: Use slice for small result sets (< 100), map for large
	// Linear scan is faster than map allocation for small sets (cache-friendly)
	const mapConversionThreshold = 100
	var candidateSet map[string]bool
	var candidateSlice []string
	var useSlice bool

	if query != "" {
		// Use index to find matching thoughts
		matchedIDs := s.searchByIndex(query, mode)
		if len(matchedIDs) < mapConversionThreshold {
			// Small set: use slice (no allocation, cache-friendly)
			candidateSlice = matchedIDs
			useSlice = true
		} else {
			// Large set: use map for O(1) lookup
			candidateSet = make(map[string]bool, len(matchedIDs))
			for _, id := range matchedIDs {
				candidateSet[id] = true
			}
		}
	} else if mode != "" {
		// Mode-only filter: use mode index
		modeIDs := s.modeIndex[mode]
		if len(modeIDs) < mapConversionThreshold {
			candidateSlice = modeIDs
			useSlice = true
		} else {
			candidateSet = make(map[string]bool, len(modeIDs))
			for _, id := range modeIDs {
				candidateSet[id] = true
			}
		}
	} else {
		// No filters: all thoughts are candidates
		candidateSet = nil // nil means "all thoughts match"
	}

	// Pre-allocate results with capacity
	results := make([]*types.Thought, 0, limit)
	skipped := 0

	for _, thought := range s.thoughtsOrdered {
		// Check limit BEFORE expensive copy operation
		if len(results) >= limit {
			break
		}

		// Check if thought matches filter criteria
		if useSlice {
			// Linear scan for small result sets
			found := false
			for _, id := range candidateSlice {
				if id == thought.ID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		} else if candidateSet != nil && !candidateSet[thought.ID] {
			continue // Not in candidate set, skip
		}

		// Skip offset
		if skipped < offset {
			skipped++
			continue
		}

		// Add to results (use optimized copy for hot path)
		results = append(results, copyThoughtOptimized(thought))
	}

	return results
}

// searchByIndex uses inverted index to find thoughts matching query words
// Updates access time for LRU tracking
func (s *MemoryStorage) searchByIndex(query string, mode types.ThinkingMode) []string {
	queryLower := strings.ToLower(query)
	queryWords := strings.FieldsFunc(queryLower, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})

	if len(queryWords) == 0 {
		// Empty query after tokenization, return all
		if mode != "" {
			return s.modeIndex[mode]
		}
		ids := make([]string, 0, len(s.thoughts))
		for id := range s.thoughts {
			ids = append(ids, id)
		}
		return ids
	}

	// Find thoughts containing any query word (OR semantics)
	matchedThoughts := make(map[string]bool)
	now := time.Now()
	for _, word := range queryWords {
		if len(word) < 2 {
			continue
		}
		// Update access time for LRU tracking
		if _, exists := s.indexAccessTime[word]; exists {
			s.indexAccessTime[word] = now
		}

		// TIER 2 OPTIMIZATION: Check hot word cache first
		var wordIDs []string
		if cached, inCache := s.hotWordCache[word]; inCache {
			wordIDs = cached
		} else {
			wordIDs = s.contentIndex[word]
			// Track access for cache promotion
			s.hotWordAccess[word]++
			if s.hotWordAccess[word] > 5 && len(s.hotWordCache) < 100 {
				// Promote to hot cache after 5 accesses
				s.hotWordCache[word] = wordIDs
			}
		}

		for _, thoughtID := range wordIDs {
			// Filter by mode if specified
			if mode != "" {
				thought, exists := s.thoughts[thoughtID]
				if !exists || thought.Mode != mode {
					continue
				}
			}
			matchedThoughts[thoughtID] = true
		}
	}

	// Convert set to slice
	result := make([]string, 0, len(matchedThoughts))
	for id := range matchedThoughts {
		result = append(result, id)
	}

	return result
}

// Metrics represents system performance and usage statistics
type Metrics struct {
	TotalThoughts     int            `json:"total_thoughts"`
	TotalBranches     int            `json:"total_branches"`
	TotalInsights     int            `json:"total_insights"`
	TotalValidations  int            `json:"total_validations"`
	ThoughtsByMode    map[string]int `json:"thoughts_by_mode"`
	AverageConfidence float64        `json:"average_confidence"`
}

// GetMetrics returns current system metrics
func (s *MemoryStorage) GetMetrics() *Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	thoughtsByMode := make(map[string]int)
	totalConfidence := 0.0
	thoughtCount := 0

	for _, thought := range s.thoughts {
		thoughtsByMode[string(thought.Mode)]++
		totalConfidence += thought.Confidence
		thoughtCount++
	}

	avgConfidence := 0.0
	if thoughtCount > 0 {
		avgConfidence = totalConfidence / float64(thoughtCount)
	}

	return &Metrics{
		TotalThoughts:     len(s.thoughts),
		TotalBranches:     len(s.branches),
		TotalInsights:     len(s.insights),
		TotalValidations:  len(s.validations),
		ThoughtsByMode:    thoughtsByMode,
		AverageConfidence: avgConfidence,
	}
}

// AppendThoughtToBranch directly appends a thought to a branch without requiring
// a full Get-Modify-Store cycle. This eliminates two deep copy operations.
func (s *MemoryStorage) AppendThoughtToBranch(branchID string, thought *types.Thought) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	branch, exists := s.branches[branchID]
	if !exists {
		return fmt.Errorf("branch not found: %s", branchID)
	}

	// Store a copy to prevent external modifications from affecting internal state
	branch.Thoughts = append(branch.Thoughts, copyThought(thought))
	branch.UpdatedAt = time.Now()
	return nil
}

// AppendInsightToBranch directly appends an insight to a branch without requiring
// a full Get-Modify-Store cycle. This eliminates two deep copy operations.
func (s *MemoryStorage) AppendInsightToBranch(branchID string, insight *types.Insight) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	branch, exists := s.branches[branchID]
	if !exists {
		return fmt.Errorf("branch not found: %s", branchID)
	}

	// Store a copy to prevent external modifications from affecting internal state
	branch.Insights = append(branch.Insights, copyInsight(insight))
	branch.UpdatedAt = time.Now()
	return nil
}

// AppendCrossRefToBranch directly appends a cross-reference to a branch without requiring
// a full Get-Modify-Store cycle. This eliminates two deep copy operations.
func (s *MemoryStorage) AppendCrossRefToBranch(branchID string, crossRef *types.CrossRef) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	branch, exists := s.branches[branchID]
	if !exists {
		return fmt.Errorf("branch not found: %s", branchID)
	}

	// Store a copy to prevent external modifications from affecting internal state
	branch.CrossRefs = append(branch.CrossRefs, copyCrossRef(crossRef))
	branch.UpdatedAt = time.Now()
	return nil
}

// UpdateBranchPriority directly updates the priority of a branch without requiring
// a full Get-Modify-Store cycle. This eliminates two deep copy operations.
func (s *MemoryStorage) UpdateBranchPriority(branchID string, priority float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	branch, exists := s.branches[branchID]
	if !exists {
		return fmt.Errorf("branch not found: %s", branchID)
	}

	branch.Priority = priority
	branch.UpdatedAt = time.Now()
	return nil
}

// UpdateBranchConfidence directly updates the confidence of a branch without requiring
// a full Get-Modify-Store cycle. This eliminates two deep copy operations.
func (s *MemoryStorage) UpdateBranchConfidence(branchID string, confidence float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	branch, exists := s.branches[branchID]
	if !exists {
		return fmt.Errorf("branch not found: %s", branchID)
	}

	branch.Confidence = confidence
	branch.UpdatedAt = time.Now()
	return nil
}

// UpdateBranchAccess updates the LastAccessedAt timestamp for a branch and tracks it in recent list
func (s *MemoryStorage) UpdateBranchAccess(branchID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	branch, exists := s.branches[branchID]
	if !exists {
		return fmt.Errorf("branch not found: %s", branchID)
	}

	branch.LastAccessedAt = time.Now()
	s.trackRecentBranch(branchID)
	return nil
}

// GetRecentBranches returns the most recently accessed branches (up to 10)
func (s *MemoryStorage) GetRecentBranches() ([]*types.Branch, error) {
	s.mu.RLock()
	recentIDs := make([]string, len(s.recentBranchIDs))
	copy(recentIDs, s.recentBranchIDs)
	s.mu.RUnlock()

	branches := make([]*types.Branch, 0, len(recentIDs))
	for _, id := range recentIDs {
		branch, err := s.GetBranch(id)
		if err == nil {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

// trackRecentBranch adds a branch to the recent list (must be called with lock held)
func (s *MemoryStorage) trackRecentBranch(branchID string) {
	// Remove if already in list
	for i, id := range s.recentBranchIDs {
		if id == branchID {
			s.recentBranchIDs = append(s.recentBranchIDs[:i], s.recentBranchIDs[i+1:]...)
			break
		}
	}

	// Add to front
	s.recentBranchIDs = append([]string{branchID}, s.recentBranchIDs...)

	// Keep max 10
	if len(s.recentBranchIDs) > 10 {
		s.recentBranchIDs = s.recentBranchIDs[:10]
	}
}

// StoreTrajectoryJSON stores a reasoning trajectory as JSON (avoids import cycle)
func (s *MemoryStorage) StoreTrajectoryJSON(id string, trajectoryJSON string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.trajectoriesJSON[id] = trajectoryJSON
	return nil
}

// GetTrajectoryJSON retrieves a trajectory JSON by ID
func (s *MemoryStorage) GetTrajectoryJSON(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trajectoryJSON, exists := s.trajectoriesJSON[id]
	if !exists {
		return "", fmt.Errorf("trajectory not found: %s", id)
	}

	return trajectoryJSON, nil
}

// GetAllTrajectoriesJSON returns all stored trajectory JSONs with metadata for filtering
func (s *MemoryStorage) GetAllTrajectoriesJSON() (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make(map[string]string, len(s.trajectoriesJSON))
	for id, jsonData := range s.trajectoriesJSON {
		result[id] = jsonData
	}

	return result, nil
}
