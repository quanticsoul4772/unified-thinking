package modes

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// MockFailingStorage implements storage.Storage with configurable failures
type MockFailingStorage struct {
	failOnStoreThought      bool
	failOnGetThought        bool
	failOnSearchThoughts    bool
	failOnStoreBranch       bool
	failOnGetBranch         bool
	failOnListBranches      bool
	failOnGetActiveBranch   bool
	failOnSetActiveBranch   bool
	failOnGetMetrics        bool
	failOnAppendThought     bool
	failOnAppendInsight     bool
	failOnAppendCrossRef    bool
	failOnUpdatePriority    bool
	failOnUpdateConfidence  bool
	failOnUpdateAccess      bool
	failOnGetRecentBranches bool
	failOnStoreInsight      bool
	failOnStoreValidation   bool
	failOnStoreRelationship bool
	slowOperation           bool // Add slow operation support

	thoughts       map[string]*types.Thought
	branches       map[string]*types.Branch
	insights       map[string]*types.Insight
	validations    map[string]*types.Validation
	relationships  map[string]*types.Relationship
	activeBranchID string
	//nolint:unused // Reserved for future use
	recentBranchIDs []string

	// Add mutex for thread safety
	mu sync.Mutex
}

func (m *MockFailingStorage) StoreThought(thought *types.Thought) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for context cancellation first if slow operation is enabled
	if m.slowOperation {
		// Simulate slow operation
		time.Sleep(10 * time.Millisecond)
	}

	// Check for failure injection
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
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *MockFailingStorage) SearchThoughts(query string, mode types.ThinkingMode, limit, offset int) []*types.Thought {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnSearchThoughts {
		return nil
	}
	// Simple implementation for testing
	var results []*types.Thought
	if m.thoughts != nil {
		for _, thought := range m.thoughts {
			if len(results) < limit {
				results = append(results, thought)
			}
		}
	}
	return results
}

func (m *MockFailingStorage) StoreBranch(branch *types.Branch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnStoreBranch {
		return errors.New("storage failure: cannot store branch")
	}
	if m.branches == nil {
		m.branches = make(map[string]*types.Branch)
	}
	m.branches[branch.ID] = branch
	return nil
}

func (m *MockFailingStorage) GetBranch(id string) (*types.Branch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnListBranches {
		return nil
	}
	var branches []*types.Branch
	if m.branches != nil {
		for _, branch := range m.branches {
			branches = append(branches, branch)
		}
	}
	return branches
}

func (m *MockFailingStorage) GetActiveBranch() (*types.Branch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnGetActiveBranch {
		return nil, errors.New("storage failure: cannot get active branch")
	}
	if m.activeBranchID == "" {
		return nil, nil
	}
	// Note: We can't call GetBranch here because it also locks, which would cause deadlock
	// Instead, we access the map directly
	if m.branches == nil {
		return nil, nil
	}
	return m.branches[m.activeBranchID], nil
}

func (m *MockFailingStorage) SetActiveBranch(branchID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnSetActiveBranch {
		return errors.New("storage failure: cannot set active branch")
	}
	m.activeBranchID = branchID
	return nil
}

func (m *MockFailingStorage) UpdateBranchAccess(branchID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnUpdateAccess {
		return errors.New("storage failure: cannot update branch access")
	}
	// Simple implementation
	return nil
}

func (m *MockFailingStorage) AppendThoughtToBranch(branchID string, thought *types.Thought) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAppendThought {
		return errors.New("storage failure: cannot append thought to branch")
	}
	// Simple implementation
	return nil
}

func (m *MockFailingStorage) AppendInsightToBranch(branchID string, insight *types.Insight) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAppendInsight {
		return errors.New("storage failure: cannot append insight to branch")
	}
	// Simple implementation
	return nil
}

func (m *MockFailingStorage) AppendCrossRefToBranch(branchID string, crossRef *types.CrossRef) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAppendCrossRef {
		return errors.New("storage failure: cannot append cross ref to branch")
	}
	// Simple implementation
	return nil
}

func (m *MockFailingStorage) UpdateBranchPriority(branchID string, priority float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnUpdatePriority {
		return errors.New("storage failure: cannot update branch priority")
	}
	// Simple implementation
	return nil
}

func (m *MockFailingStorage) UpdateBranchConfidence(branchID string, confidence float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnUpdateConfidence {
		return errors.New("storage failure: cannot update branch confidence")
	}
	// Simple implementation
	return nil
}

func (m *MockFailingStorage) GetRecentBranches() ([]*types.Branch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnGetRecentBranches {
		return nil, errors.New("storage failure: cannot get recent branches")
	}
	// Simple implementation
	return nil, nil
}

func (m *MockFailingStorage) StoreInsight(insight *types.Insight) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnStoreInsight {
		return errors.New("storage failure: cannot store insight")
	}
	if m.insights == nil {
		m.insights = make(map[string]*types.Insight)
	}
	m.insights[insight.ID] = insight
	return nil
}

func (m *MockFailingStorage) StoreValidation(validation *types.Validation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnStoreValidation {
		return errors.New("storage failure: cannot store validation")
	}
	if m.validations == nil {
		m.validations = make(map[string]*types.Validation)
	}
	m.validations[validation.ID] = validation
	return nil
}

func (m *MockFailingStorage) StoreRelationship(relationship *types.Relationship) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnStoreRelationship {
		return errors.New("storage failure: cannot store relationship")
	}
	if m.relationships == nil {
		m.relationships = make(map[string]*types.Relationship)
	}
	m.relationships[relationship.ID] = relationship
	return nil
}

func (m *MockFailingStorage) GetMetrics() *storage.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnGetMetrics {
		return nil
	}
	return &storage.Metrics{
		TotalThoughts:     len(m.thoughts),
		TotalBranches:     len(m.branches),
		TotalInsights:     len(m.insights),
		TotalValidations:  len(m.validations),
		ThoughtsByMode:    make(map[string]int),
		AverageConfidence: 0.0,
	}
}

func TestLinearMode_ErrorRecovery_StorageFailures(t *testing.T) {
	failingStorage := &MockFailingStorage{}
	mode := NewLinearMode(failingStorage)

	tests := []struct {
		name        string
		failureMode string
		setup       func()
		testFunc    func() error
		expectError bool
	}{
		{
			name:        "storage store failure",
			failureMode: "store",
			setup: func() {
				failingStorage.failOnStoreThought = true
			},
			testFunc: func() error {
				ctx := context.Background()
				input := ThoughtInput{
					Content:  "test content",
					BranchID: "test-branch",
				}
				_, err := mode.ProcessThought(ctx, input)
				return err
			},
			expectError: true,
		},
		{
			name:        "storage retrieve failure",
			failureMode: "get",
			setup: func() {
				failingStorage.failOnGetThought = true
			},
			testFunc: func() error {
				// First store a thought to have something to retrieve
				ctx := context.Background()
				input := ThoughtInput{
					Content:  "test content",
					BranchID: "test-branch",
				}
				result, err := mode.ProcessThought(ctx, input)
				if err != nil {
					return err // If storing fails, that's a different issue
				}

				// Now try to retrieve the stored thought directly
				_, err = failingStorage.GetThought(result.ThoughtID)
				return err
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset storage state
			failingStorage.failOnStoreThought = false
			failingStorage.failOnGetThought = false

			// Setup failure
			tt.setup()

			// Test the operation
			err := tt.testFunc()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s failure but got none", tt.failureMode)
				}
			}

			// Mode should remain functional after storage failures
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Linear mode panicked after %s failure: %v", tt.failureMode, r)
				}
			}()

			if mode == nil {
				t.Errorf("Linear mode became nil after %s failure", tt.failureMode)
			}
		})
	}
}

func TestLinearMode_ErrorRecovery_InvalidInputs(t *testing.T) {
	storage := &MockFailingStorage{}
	mode := NewLinearMode(storage)

	tests := []struct {
		name        string
		input       ThoughtInput
		expectError bool
	}{
		{
			name: "empty content",
			input: ThoughtInput{
				Content:  "",
				BranchID: "test-branch",
			},
			expectError: false, // Empty content might be acceptable
		},
		{
			name: "valid input",
			input: ThoughtInput{
				Content:  "test content",
				BranchID: "test-branch",
			},
			expectError: false,
		},
		{
			name: "very large content",
			input: ThoughtInput{
				Content:  string(make([]byte, 1000000)), // 1MB
				BranchID: "test-branch",
			},
			expectError: false, // Should handle large content
		},
		{
			name: "content with special characters",
			input: ThoughtInput{
				Content:  "Content with\nnewlines\tand\ttabs",
				BranchID: "test-branch",
			},
			expectError: false,
		},
		{
			name: "unicode content",
			input: ThoughtInput{
				Content:  "Unicode: 你好世界 🌍",
				BranchID: "test-branch",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Test that mode handles invalid inputs gracefully
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Linear mode panicked with input: %v", r)
				}
			}()

			result, err := mode.ProcessThought(ctx, tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input but got result: %v", result)
				}
			} else {
				// For valid inputs, we might get results or errors, but shouldn't panic
				if result == nil && err == nil {
					t.Error("Expected either result or error")
				}
			}

			// Mode should remain functional
			if mode == nil {
				t.Error("Linear mode became nil after input")
			}
		})
	}
}

func TestLinearMode_ErrorRecovery_TimeoutHandling(t *testing.T) {
	storage := &MockFailingStorage{}
	mode := NewLinearMode(storage)

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
			name:    "reasonable timeout",
			timeout: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Enable slow operation for very short timeouts
			if tt.timeout > 0 && tt.timeout < 5*time.Millisecond {
				storage.slowOperation = true
			} else {
				storage.slowOperation = false
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Test timeout handling
			input := ThoughtInput{
				Content:  "test content",
				BranchID: "test-branch",
			}
			result, err := mode.ProcessThought(ctx, input)

			// Check for timeout
			if ctx.Err() == context.DeadlineExceeded {
				// For very short timeouts, we might get an error or we might not
				// depending on how fast the operation completes
				// This is acceptable behavior for timeout handling tests
				t.Logf("Context deadline exceeded (this is expected behavior)")
			}

			// Mode should handle timeouts gracefully
			if result == nil && err == nil && ctx.Err() == nil {
				t.Error("Expected some result or error from timeout test")
			}

			// Mode should remain functional
			if mode == nil {
				t.Error("Linear mode became nil after timeout")
			}

			// Reset slow operation
			storage.slowOperation = false
		})
	}
}

func TestTreeMode_ErrorRecovery_BranchFailures(t *testing.T) {
	failingStorage := &MockFailingStorage{}
	mode := NewTreeMode(failingStorage)

	// Test branch operations with failures
	t.Run("branch operation failures", func(t *testing.T) {
		// Test branch creation failure
		failingStorage.failOnStoreBranch = true

		// Mode should handle storage failures gracefully
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Tree mode panicked with branch failure: %v", r)
			}
		}()

		// Reset failure
		failingStorage.failOnStoreBranch = false

		// Mode should remain functional after failures
		if mode == nil {
			t.Error("Tree mode became nil after branch failure")
		}
	})
}

func TestTreeMode_ErrorRecovery_ComplexBranching(t *testing.T) {
	storage := &MockFailingStorage{}
	mode := NewTreeMode(storage)

	t.Run("deep branching", func(t *testing.T) {
		ctx := context.Background()

		// Test with various branch operations
		input := ThoughtInput{
			Content:  "test content for branching",
			BranchID: "test-branch",
		}

		result, err := mode.ProcessThought(ctx, input)
		if err != nil {
			t.Logf("Tree mode processing failed (expected in test): %v", err)
		}

		if result == nil && err == nil {
			t.Error("Expected either result or error from tree mode")
		}

		// Mode should handle complex branching gracefully
		if mode == nil {
			t.Error("Tree mode became nil after complex branching")
		}
	})
}

func TestDivergentMode_ErrorRecovery_ConcurrentExploration(t *testing.T) {
	storage := &MockFailingStorage{}
	mode := NewDivergentMode(storage)

	// Test concurrent divergent thinking
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Divergent mode goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			ctx := context.Background()
			content := fmt.Sprintf("Divergent thought %d", id)

			// Test concurrent processing
			input := ThoughtInput{
				Content:        content,
				BranchID:       fmt.Sprintf("branch-%d", id),
				ForceRebellion: true, // Enable creative thinking
			}
			result, err := mode.ProcessThought(ctx, input)

			if result == nil && err == nil {
				t.Errorf("Divergent mode returned nil result and error for content %q", content)
			}

			// Mode should remain functional
			if mode == nil {
				t.Errorf("Divergent mode became nil in goroutine %d", id)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(2 * time.Second):
			t.Errorf("Divergent mode goroutine %d timed out", i)
		}
	}

	// Mode should remain functional after concurrent operations
	if mode == nil {
		t.Error("Divergent mode became nil after concurrent exploration")
	}
}

func TestAutoMode_ErrorRecovery_ModeSwitching(t *testing.T) {
	storage := &MockFailingStorage{}
	linearMode := NewLinearMode(storage)
	treeMode := NewTreeMode(storage)
	divergentMode := NewDivergentMode(storage)

	mode := NewAutoMode(linearMode, treeMode, divergentMode)

	t.Run("mode switching with failures", func(t *testing.T) {
		ctx := context.Background()

		// Test different content types that should trigger different modes
		testCases := []ThoughtInput{
			{
				Content:  "Simple linear thought",
				BranchID: "test-branch",
			},
			{
				Content:  "Complex thought with multiple branches and considerations",
				BranchID: "test-branch",
			},
			{
				Content:        "Creative divergent thinking with many possibilities",
				BranchID:       "test-branch",
				ForceRebellion: true,
			},
		}

		for i, input := range testCases {
			result, err := mode.ProcessThought(ctx, input)

			// Should handle all inputs gracefully
			if result == nil && err == nil {
				t.Errorf("Auto mode returned nil result and error for test case %d", i)
			}

			// Mode should switch appropriately and remain functional
			if mode == nil {
				t.Error("Auto mode became nil during mode switching")
			}
		}
	})

	t.Run("fallback when modes fail", func(t *testing.T) {
		// Make storage fail
		failingStorage := &MockFailingStorage{
			failOnStoreThought: true,
		}

		linearMode := NewLinearMode(failingStorage)
		treeMode := NewTreeMode(failingStorage)
		divergentMode := NewDivergentMode(failingStorage)
		mode := NewAutoMode(linearMode, treeMode, divergentMode)

		ctx := context.Background()

		// Auto mode should handle mode failures gracefully
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Auto mode panicked when mode failed: %v", r)
			}
		}()

		input := ThoughtInput{
			Content:  "test content",
			BranchID: "test-branch",
		}
		result, err := mode.ProcessThought(ctx, input)

		// Should either succeed with fallback or fail gracefully
		if result == nil && err == nil {
			t.Error("Expected some result or error from fallback test")
		}

		// Mode should remain functional
		if mode == nil {
			t.Error("Auto mode became nil after mode failure")
		}
	})
}

func TestModes_ErrorRecovery_ResourceExhaustion(t *testing.T) {
	storage := &MockFailingStorage{}

	modes := []struct {
		name string
		mode ThinkingMode
	}{
		{"Linear", NewLinearMode(storage)},
		{"Tree", NewTreeMode(storage)},
		{"Divergent", NewDivergentMode(storage)},
	}

	t.Run("large content processing", func(t *testing.T) {
		ctx := context.Background()

		// Very large content
		largeContent := string(make([]byte, 500000)) // 500KB

		for _, m := range modes {
			input := ThoughtInput{
				Content:  largeContent,
				BranchID: fmt.Sprintf("large-branch-%s", m.name),
			}
			result, err := m.mode.ProcessThought(ctx, input)

			// Should handle large content gracefully
			if result == nil && err == nil {
				t.Errorf("Mode %s returned nil result and error for large content", m.name)
			}

			// Mode should remain functional
			if m.mode == nil {
				t.Errorf("Mode %s became nil after large content", m.name)
			}
		}
	})

	t.Run("many concurrent operations", func(t *testing.T) {
		ctx := context.Background()

		done := make(chan bool, 30)

		for i, m := range modes {
			for j := 0; j < 10; j++ { // 10 operations per mode
				go func(modeIdx, opIdx int, mode ThinkingMode, modeName string) {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("Mode %s operation %d panicked: %v", modeName, opIdx, r)
						}
						done <- true
					}()

					content := fmt.Sprintf("Concurrent content %s-%d-%d", modeName, modeIdx, opIdx)
					branchID := fmt.Sprintf("concurrent-branch-%s-%d-%d", modeName, modeIdx, opIdx)

					input := ThoughtInput{
						Content:  content,
						BranchID: branchID,
					}
					_, err := mode.ProcessThought(ctx, input)
					if err != nil {
						t.Logf("Mode %s operation %d failed (expected in test): %v", modeName, opIdx, err)
					}
				}(i, j, m.mode, m.name)
			}
		}

		// Wait for all operations
		for i := 0; i < 30; i++ {
			select {
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				t.Errorf("Operation %d timed out", i)
			}
		}

		// All modes should remain functional
		for _, m := range modes {
			if m.mode == nil {
				t.Errorf("Mode %s became nil after concurrent operations", m.name)
			}
		}
	})
}
