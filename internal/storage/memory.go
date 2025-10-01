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
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// MemoryStorage implements in-memory storage with thread-safe operations.
// All Get methods return deep copies to prevent external mutation of internal state.
type MemoryStorage struct {
	mu            sync.RWMutex
	thoughts      map[string]*types.Thought
	branches      map[string]*types.Branch
	insights      map[string]*types.Insight
	validations   map[string]*types.Validation
	relationships map[string]*types.Relationship

	// Search indices for O(1) word lookup (optimization for SearchThoughts)
	contentIndex map[string][]string              // word -> []thoughtIDs
	modeIndex    map[types.ThinkingMode][]string // mode -> []thoughtIDs

	activeBranchID string

	// Counters for ID generation
	thoughtCounter      int
	branchCounter       int
	insightCounter      int
	validationCounter   int
	relationshipCounter int
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		thoughts:      make(map[string]*types.Thought),
		branches:      make(map[string]*types.Branch),
		insights:      make(map[string]*types.Insight),
		validations:   make(map[string]*types.Validation),
		relationships: make(map[string]*types.Relationship),
		contentIndex:  make(map[string][]string),
		modeIndex:     make(map[types.ThinkingMode][]string),
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

	return nil
}

// indexThoughtContent tokenizes thought content and adds to inverted index
func (s *MemoryStorage) indexThoughtContent(thought *types.Thought) {
	// Tokenize content by splitting on whitespace and punctuation
	content := strings.ToLower(thought.Content)
	words := strings.FieldsFunc(content, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})

	// Add thought ID to index for each unique word
	seen := make(map[string]bool)
	for _, word := range words {
		if word == "" || len(word) < 2 { // Skip empty and single-char tokens
			continue
		}
		if seen[word] {
			continue // Skip duplicates within same thought
		}
		seen[word] = true
		s.contentIndex[word] = append(s.contentIndex[word], thought.ID)
	}
}

// GetThought retrieves a thought by ID (returns a copy to prevent data races)
func (s *MemoryStorage) GetThought(id string) (*types.Thought, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	thought, exists := s.thoughts[id]
	if !exists {
		return nil, fmt.Errorf("thought not found: %s", id)
	}
	// Return a deep copy to prevent external modification
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

	s.branches[branch.ID] = branch

	// Set as active if it's the first branch
	if s.activeBranchID == "" {
		s.activeBranchID = branch.ID
	}

	return nil
}

// GetBranch retrieves a branch by ID (returns a copy to prevent data races)
func (s *MemoryStorage) GetBranch(id string) (*types.Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	branch, exists := s.branches[id]
	if !exists {
		return nil, fmt.Errorf("branch not found: %s", id)
	}
	// Return a deep copy to prevent external modification
	return copyBranch(branch), nil
}

// ListBranches returns all branches (returns copies to prevent data races)
func (s *MemoryStorage) ListBranches() []*types.Branch {
	s.mu.RLock()
	defer s.mu.RUnlock()

	branches := make([]*types.Branch, 0, len(s.branches))
	for _, branch := range s.branches {
		// Return deep copies to prevent external modification
		branches = append(branches, copyBranch(branch))
	}
	return branches
}

// GetActiveBranch returns the currently active branch (returns a copy to prevent data races)
func (s *MemoryStorage) GetActiveBranch() (*types.Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.activeBranchID == "" {
		return nil, fmt.Errorf("no active branch")
	}

	branch, exists := s.branches[s.activeBranchID]
	if !exists {
		// Active branch was deleted - this is a data inconsistency
		return nil, fmt.Errorf("active branch %s no longer exists", s.activeBranchID)
	}

	// Return a deep copy to prevent external modification
	return copyBranch(branch), nil
}

// SetActiveBranch sets the active branch
func (s *MemoryStorage) SetActiveBranch(branchID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.branches[branchID]; !exists {
		return fmt.Errorf("branch not found: %s", branchID)
	}

	s.activeBranchID = branchID
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

// SearchThoughts searches thoughts by content or type (returns copies to prevent data races)
// limit and offset support pagination - limit of 0 returns all results
// Uses inverted index for O(1) word lookup when query is provided
func (s *MemoryStorage) SearchThoughts(query string, mode types.ThinkingMode, limit, offset int) []*types.Thought {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var candidateIDs []string

	// Fast path: Use index for query lookup if available
	if query != "" {
		candidateIDs = s.searchByIndex(query, mode)
	} else if mode != "" {
		// Mode-only filter: use mode index
		candidateIDs = s.modeIndex[mode]
	} else {
		// No filters: return all thoughts (fallback to map iteration)
		candidateIDs = make([]string, 0, len(s.thoughts))
		for id := range s.thoughts {
			candidateIDs = append(candidateIDs, id)
		}
	}

	// Apply pagination and return results
	results := make([]*types.Thought, 0, limit)
	for i := offset; i < len(candidateIDs); i++ {
		if limit > 0 && len(results) >= limit {
			break
		}

		thought, exists := s.thoughts[candidateIDs[i]]
		if !exists {
			continue // Thought may have been deleted (not implemented yet)
		}

		results = append(results, copyThought(thought))
	}

	return results
}

// searchByIndex uses inverted index to find thoughts matching query words
func (s *MemoryStorage) searchByIndex(query string, mode types.ThinkingMode) []string {
	queryLower := strings.ToLower(query)
	queryWords := strings.FieldsFunc(queryLower, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
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
	for _, word := range queryWords {
		if len(word) < 2 {
			continue
		}
		for _, thoughtID := range s.contentIndex[word] {
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

	branch.Thoughts = append(branch.Thoughts, thought)
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

	branch.Insights = append(branch.Insights, insight)
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

	branch.CrossRefs = append(branch.CrossRefs, crossRef)
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
