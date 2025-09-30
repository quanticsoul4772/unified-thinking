package storage

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
	mu            sync.RWMutex
	thoughts      map[string]*types.Thought
	branches      map[string]*types.Branch
	insights      map[string]*types.Insight
	validations   map[string]*types.Validation
	relationships map[string]*types.Relationship

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
	}
}

// StoreThought stores a thought
func (s *MemoryStorage) StoreThought(thought *types.Thought) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if thought.ID == "" {
		s.thoughtCounter++
		thought.ID = fmt.Sprintf("thought-%d-%d", time.Now().Unix(), s.thoughtCounter)
	}

	s.thoughts[thought.ID] = thought
	return nil
}

// GetThought retrieves a thought by ID
func (s *MemoryStorage) GetThought(id string) (*types.Thought, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	thought, exists := s.thoughts[id]
	if !exists {
		return nil, fmt.Errorf("thought not found: %s", id)
	}
	return thought, nil
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

// GetBranch retrieves a branch by ID
func (s *MemoryStorage) GetBranch(id string) (*types.Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	branch, exists := s.branches[id]
	if !exists {
		return nil, fmt.Errorf("branch not found: %s", id)
	}
	return branch, nil
}

// ListBranches returns all branches
func (s *MemoryStorage) ListBranches() []*types.Branch {
	s.mu.RLock()
	defer s.mu.RUnlock()

	branches := make([]*types.Branch, 0, len(s.branches))
	for _, branch := range s.branches {
		branches = append(branches, branch)
	}
	return branches
}

// GetActiveBranch returns the currently active branch
func (s *MemoryStorage) GetActiveBranch() (*types.Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.activeBranchID == "" {
		return nil, fmt.Errorf("no active branch")
	}

	return s.branches[s.activeBranchID], nil
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

// SearchThoughts searches thoughts by content or type
func (s *MemoryStorage) SearchThoughts(query string, mode types.ThinkingMode) []*types.Thought {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*types.Thought
	queryLower := strings.ToLower(query)
	
	for _, thought := range s.thoughts {
		// Search by content and mode
		matchesQuery := query == "" || strings.Contains(strings.ToLower(thought.Content), queryLower)
		matchesMode := mode == "" || thought.Mode == mode

		if matchesQuery && matchesMode {
			results = append(results, thought)
		}
	}
	return results
}
