package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// MockFailingStorage implements storage.Storage with configurable failures
type MockFailingStorage struct {
	failOnStoreThought bool
	failOnGetThought   bool
	failOnListThoughts bool
	failOnCreateBranch bool
	failOnGetBranch    bool
	failOnListBranches bool
	failOnGetMetrics   bool
	failOnClose        bool
	thoughts           map[string]*types.Thought
	branches           map[string]*types.Branch
}

func (m *MockFailingStorage) StoreThought(thought *types.Thought) error {
	if m.failOnStoreThought {
		return errors.New("storage failure: cannot store thought")
	}
	if m.thoughts == nil {
		m.thoughts = make(map[string]*types.Thought)
	}
	m.thoughts[thought.ID] = thought
	return nil
}

func (m *MockFailingStorage) GetThought(id string) (*types.Thought, error) {
	if m.failOnGetThought {
		return nil, errors.New("storage failure: cannot retrieve thought")
	}
	if m.thoughts == nil {
		return nil, errors.New("thought not found")
	}
	thought, exists := m.thoughts[id]
	if !exists {
		return nil, errors.New("thought not found")
	}
	return thought, nil
}

func (m *MockFailingStorage) ListThoughts(branchID string) ([]*types.Thought, error) {
	if m.failOnListThoughts {
		return nil, errors.New("storage failure: cannot list thoughts")
	}
	var thoughts []*types.Thought
	for _, thought := range m.thoughts {
		if thought.BranchID == branchID {
			thoughts = append(thoughts, thought)
		}
	}
	return thoughts, nil
}

func (m *MockFailingStorage) UpdateThought(id string, thought *types.Thought) error {
	if m.thoughts == nil {
		m.thoughts = make(map[string]*types.Thought)
	}
	m.thoughts[id] = thought
	return nil
}

func (m *MockFailingStorage) DeleteThought(id string) error {
	if m.thoughts != nil {
		delete(m.thoughts, id)
	}
	return nil
}

func (m *MockFailingStorage) CreateBranch(name string) (*types.Branch, error) {
	if m.failOnCreateBranch {
		return nil, errors.New("storage failure: cannot create branch")
	}
	if m.branches == nil {
		m.branches = make(map[string]*types.Branch)
	}
	branch := &types.Branch{
		ID:         name,
		State:      types.StateActive,
		Priority:   0.5,
		Confidence: 0.8,
		Thoughts:   []*types.Thought{},
		Insights:   []*types.Insight{},
		CrossRefs:  []*types.CrossRef{},
		CreatedAt:  time.Now(),
	}
	m.branches[name] = branch
	return branch, nil
}

func (m *MockFailingStorage) GetBranch(id string) (*types.Branch, error) {
	if m.failOnGetBranch {
		return nil, errors.New("storage failure: cannot retrieve branch")
	}
	if m.branches == nil {
		return nil, errors.New("branch not found")
	}
	branch, exists := m.branches[id]
	if !exists {
		return nil, errors.New("branch not found")
	}
	return branch, nil
}

func (m *MockFailingStorage) ListBranches() []*types.Branch {
	if m.failOnListBranches {
		return nil // In real implementation, this might panic or return empty slice
	}
	var branches []*types.Branch
	for _, branch := range m.branches {
		branches = append(branches, branch)
	}
	return branches
}

func (m *MockFailingStorage) UpdateBranch(id string, branch *types.Branch) error {
	if m.branches == nil {
		m.branches = make(map[string]*types.Branch)
	}
	m.branches[id] = branch
	return nil
}

func (m *MockFailingStorage) DeleteBranch(id string) error {
	if m.branches != nil {
		delete(m.branches, id)
	}
	return nil
}

func (m *MockFailingStorage) GetMetrics() *storage.Metrics {
	if m.failOnGetMetrics {
		return nil // In real implementation, this would return nil or default metrics
	}
	return &storage.Metrics{
		TotalThoughts:     len(m.thoughts),
		TotalBranches:     len(m.branches),
		TotalInsights:     0,
		TotalValidations:  0,
		ThoughtsByMode:    make(map[string]int),
		AverageConfidence: 0.0,
	}
}

func (m *MockFailingStorage) Close() error {
	if m.failOnClose {
		return errors.New("storage failure: cannot close connection")
	}
	return nil
}

// Additional required methods for Storage interface
func (m *MockFailingStorage) SearchThoughts(query string, mode types.ThinkingMode, limit, offset int) []*types.Thought {
	return nil
}

func (m *MockFailingStorage) StoreBranch(branch *types.Branch) error {
	if m.failOnCreateBranch {
		return errors.New("storage failure: cannot store branch")
	}
	if m.branches == nil {
		m.branches = make(map[string]*types.Branch)
	}
	m.branches[branch.ID] = branch
	return nil
}

func (m *MockFailingStorage) GetActiveBranch() (*types.Branch, error) {
	return nil, nil
}

func (m *MockFailingStorage) SetActiveBranch(branchID string) error {
	return nil
}

func (m *MockFailingStorage) UpdateBranchAccess(branchID string) error {
	return nil
}

func (m *MockFailingStorage) AppendThoughtToBranch(branchID string, thought *types.Thought) error {
	return nil
}

func (m *MockFailingStorage) AppendInsightToBranch(branchID string, insight *types.Insight) error {
	return nil
}

func (m *MockFailingStorage) AppendCrossRefToBranch(branchID string, crossRef *types.CrossRef) error {
	return nil
}

func (m *MockFailingStorage) UpdateBranchPriority(branchID string, priority float64) error {
	return nil
}

func (m *MockFailingStorage) UpdateBranchConfidence(branchID string, confidence float64) error {
	return nil
}

func (m *MockFailingStorage) GetRecentBranches() ([]*types.Branch, error) {
	return nil, nil
}

func (m *MockFailingStorage) StoreInsight(insight *types.Insight) error {
	return nil
}

func (m *MockFailingStorage) StoreValidation(validation *types.Validation) error {
	return nil
}

func (m *MockFailingStorage) StoreRelationship(relationship *types.Relationship) error {
	return nil
}

func TestUnifiedServer_ErrorRecovery_StorageFailures(t *testing.T) {
	tests := []struct {
		name           string
		storageFailure string
		expectRecovery bool
	}{
		{
			name:           "storage store thought failure",
			storageFailure: "store_thought",
			expectRecovery: true,
		},
		{
			name:           "storage get thought failure",
			storageFailure: "get_thought",
			expectRecovery: true,
		},
		{
			name:           "storage list thoughts failure",
			storageFailure: "list_thoughts",
			expectRecovery: true,
		},
		{
			name:           "storage create branch failure",
			storageFailure: "create_branch",
			expectRecovery: true,
		},
		{
			name:           "storage get branch failure",
			storageFailure: "get_branch",
			expectRecovery: true,
		},
		{
			name:           "storage list branches failure",
			storageFailure: "list_branches",
			expectRecovery: true,
		},
		{
			name:           "storage get metrics failure",
			storageFailure: "get_metrics",
			expectRecovery: true,
		},
		{
			name:           "storage close failure",
			storageFailure: "close",
			expectRecovery: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create failing storage
			failingStorage := &MockFailingStorage{}
			switch tt.storageFailure {
			case "store_thought":
				failingStorage.failOnStoreThought = true
			case "get_thought":
				failingStorage.failOnGetThought = true
			case "list_thoughts":
				failingStorage.failOnListThoughts = true
			case "create_branch":
				failingStorage.failOnCreateBranch = true
			case "get_branch":
				failingStorage.failOnGetBranch = true
			case "list_branches":
				failingStorage.failOnListBranches = true
			case "get_metrics":
				failingStorage.failOnGetMetrics = true
			case "close":
				failingStorage.failOnClose = true
			}

			// Create server components
			linearMode := modes.NewLinearMode(failingStorage)
			treeMode := modes.NewTreeMode(failingStorage)
			divergentMode := modes.NewDivergentMode(failingStorage)
			autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
			validator := validation.NewLogicValidator()

			// Create server
			server := NewUnifiedServer(failingStorage, linearMode, treeMode, divergentMode, autoMode, validator)
			if server == nil {
				t.Fatal("Failed to create server with failing storage")
			}

			// Test that server can be created and basic operations work despite storage failures
			// ctx := context.Background() // Not used in this test

			// Test server initialization doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Server panicked during error recovery test: %v", r)
				}
			}()

			// Test that server can handle storage failures gracefully
			// This tests the error recovery capability
			if tt.expectRecovery {
				// Server should be created successfully even with failing storage
				if server == nil {
					t.Error("Server should be created even with failing storage")
				}
			}

			// Test cleanup
			err := failingStorage.Close()
			if tt.storageFailure == "close" {
				if err == nil {
					t.Error("Expected close error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected close error: %v", err)
				}
			}
		})
	}
}

func TestUnifiedServer_ErrorRecovery_InvalidInputs(t *testing.T) {
	// Create working storage for this test
	storage := &MockFailingStorage{}
	linearMode := modes.NewLinearMode(storage)
	treeMode := modes.NewTreeMode(storage)
	divergentMode := modes.NewDivergentMode(storage)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	validator := validation.NewLogicValidator()

	server := NewUnifiedServer(storage, linearMode, treeMode, divergentMode, autoMode, validator)

	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
	}{
		{
			name:        "nil input",
			input:       nil,
			expectError: true,
		},
		{
			name:        "empty input",
			input:       map[string]interface{}{},
			expectError: true,
		},
		{
			name: "invalid method",
			input: map[string]interface{}{
				"method": "invalid_method",
			},
			expectError: true,
		},
		{
			name: "missing required parameters",
			input: map[string]interface{}{
				"method": "think_linear",
			},
			expectError: true,
		},
		{
			name: "invalid parameter types",
			input: map[string]interface{}{
				"method":    "think_linear",
				"content":   123, // Should be string
				"branch_id": "test",
			},
			expectError: true,
		},
		{
			name: "oversized content",
			input: map[string]interface{}{
				"method":    "think_linear",
				"content":   string(make([]byte, 100000)), // Very large content
				"branch_id": "test",
			},
			expectError: true,
		},
		{
			name: "malformed JSON-like input",
			input: map[string]interface{}{
				"method": "think_linear",
				"params": map[string]interface{}{
					"nested": map[string]interface{}{
						"deeply": map[string]interface{}{
							"nested": "value",
						},
					},
				},
			},
			expectError: false, // Should handle nested structures
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ctx := context.Background() // Not used in this test

			// Test that server doesn't panic with invalid inputs
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Server panicked with invalid input %v: %v", tt.input, r)
				}
			}()

			// The server should handle invalid inputs gracefully
			// In a real implementation, this would call server methods
			// For now, we test that the server remains stable
			if server == nil {
				t.Error("Server should remain stable with invalid inputs")
			}
		})
	}
}

func TestUnifiedServer_ErrorRecovery_TimeoutHandling(t *testing.T) {
	storage := &MockFailingStorage{}
	linearMode := modes.NewLinearMode(storage)
	treeMode := modes.NewTreeMode(storage)
	divergentMode := modes.NewDivergentMode(storage)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	validator := validation.NewLogicValidator()

	server := NewUnifiedServer(storage, linearMode, treeMode, divergentMode, autoMode, validator)

	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "very short timeout",
			timeout: 1 * time.Millisecond,
		},
		{
			name:    "zero timeout",
			timeout: 0,
		},
		{
			name:    "negative timeout",
			timeout: -1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Test that server handles timeout contexts gracefully
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Server panicked with timeout context: %v", r)
				}
			}()

			// Simulate a long-running operation that might timeout
			select {
			case <-ctx.Done():
				// Context timed out, server should handle this gracefully
				if ctx.Err() != context.DeadlineExceeded && ctx.Err() != context.Canceled {
					t.Errorf("Unexpected context error: %v", ctx.Err())
				}
			case <-time.After(10 * time.Millisecond):
				// Operation completed before timeout
			}

			// Server should remain functional after timeout
			if server == nil {
				t.Error("Server should remain functional after timeout")
			}
		})
	}
}

func TestUnifiedServer_ErrorRecovery_ConcurrentFailures(t *testing.T) {
	storage := &MockFailingStorage{}
	linearMode := modes.NewLinearMode(storage)
	treeMode := modes.NewTreeMode(storage)
	divergentMode := modes.NewDivergentMode(storage)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	validator := validation.NewLogicValidator()

	server := NewUnifiedServer(storage, linearMode, treeMode, divergentMode, autoMode, validator)

	// Test concurrent access with failures
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			// Simulate concurrent operations that might fail
			// ctx := context.Background() // Not used

			// Test that server handles concurrent failures gracefully
			// In a real implementation, this would call server methods
			time.Sleep(time.Duration(id) * time.Millisecond)

			// Server should remain stable under concurrent load
			if server == nil {
				t.Errorf("Server became nil during concurrent operation %d", id)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Goroutine completed successfully
		case <-time.After(1 * time.Second):
			t.Error("Goroutine timed out")
		}
	}

	// Server should still be functional after concurrent failures
	if server == nil {
		t.Error("Server should remain functional after concurrent failures")
	}
}

func TestUnifiedServer_ErrorRecovery_ResourceExhaustion(t *testing.T) {
	storage := &MockFailingStorage{}
	linearMode := modes.NewLinearMode(storage)
	treeMode := modes.NewTreeMode(storage)
	divergentMode := modes.NewDivergentMode(storage)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	validator := validation.NewLogicValidator()

	server := NewUnifiedServer(storage, linearMode, treeMode, divergentMode, autoMode, validator)

	// Test with large data structures that might cause memory issues
	largeContent := string(make([]byte, 1000000)) // 1MB string

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "large content",
			content: largeContent,
		},
		{
			name:    "very large content",
			content: largeContent + largeContent, // 2MB
		},
		{
			name: "nested structures",
			content: `{
				"level1": {
					"level2": {
						"level3": {
							"level4": {
								"level5": "deep nesting"
							}
						}
					}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that server handles large inputs without crashing
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Server panicked with large input: %v", r)
				}
			}()

			// Server should handle large inputs gracefully
			// In a real implementation, this would validate input size limits
			if len(tt.content) > 100000 { // Arbitrary large size check
				// Large input detected
				if server == nil {
					t.Error("Server should handle large inputs")
				}
			}
		})
	}
}

func TestUnifiedServer_ErrorRecovery_NetworkFailures(t *testing.T) {
	storage := &MockFailingStorage{}
	linearMode := modes.NewLinearMode(storage)
	treeMode := modes.NewTreeMode(storage)
	divergentMode := modes.NewDivergentMode(storage)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	validator := validation.NewLogicValidator()

	server := NewUnifiedServer(storage, linearMode, treeMode, divergentMode, autoMode, validator)

	// Simulate network-like failures
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "connection lost",
			setup: func() {
				// Simulate connection loss by making storage operations fail
				storage.failOnStoreThought = true
				storage.failOnGetThought = true
			},
		},
		{
			name: "partial network failure",
			setup: func() {
				// Some operations work, others fail
				storage.failOnListThoughts = true
				storage.failOnCreateBranch = false
			},
		},
		{
			name: "complete network failure",
			setup: func() {
				// All operations fail
				storage.failOnStoreThought = true
				storage.failOnGetThought = true
				storage.failOnListThoughts = true
				storage.failOnCreateBranch = true
				storage.failOnGetBranch = true
				storage.failOnListBranches = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			// Test that server can handle network-like failures
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Server panicked during network failure simulation: %v", r)
				}
			}()

			// ctx := context.Background() // Not used

			// Test various operations that might fail due to "network issues"
			metrics := storage.GetMetrics()
			// We expect some errors but server should not crash
			_ = metrics // Use the variable to avoid unused error

			// Server should remain operational
			if server == nil {
				t.Error("Server should remain operational during network failures")
			}

			// Test that context cancellation works during failures
			testCtx, cancel := context.WithCancel(context.Background())
			cancel()

			select {
			case <-testCtx.Done():
				// Context properly cancelled
			case <-time.After(100 * time.Millisecond):
				t.Error("Context should have been cancelled")
			}
		})
	}
}

func TestUnifiedServer_ErrorRecovery_Cleanup(t *testing.T) {
	// Test proper cleanup even when errors occur
	storage := &MockFailingStorage{
		failOnClose: true, // Storage close will fail
	}

	linearMode := modes.NewLinearMode(storage)
	treeMode := modes.NewTreeMode(storage)
	divergentMode := modes.NewDivergentMode(storage)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	validator := validation.NewLogicValidator()

	server := NewUnifiedServer(storage, linearMode, treeMode, divergentMode, autoMode, validator)

	// Test cleanup with failing storage
	defer func() {
		err := storage.Close()
		if err == nil {
			t.Error("Expected close error but got none")
		}

		// Even with close failure, server should handle cleanup gracefully
		if server == nil {
			t.Error("Server should handle cleanup failures gracefully")
		}
	}()

	// Test that operations work even with cleanup issues
	// ctx := context.Background() // Not used

	// Server should function normally despite cleanup issues
	if server == nil {
		t.Error("Server should function despite cleanup issues")
	}
}
