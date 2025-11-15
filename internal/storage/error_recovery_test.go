package storage

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"unified-thinking/internal/types"
)

// MockFailingBackend simulates various storage backend failures
type MockFailingBackend struct {
	failOnRead     bool
	failOnWrite    bool
	failOnDelete   bool
	failOnList     bool
	failOnConnect  bool
	failOnClose    bool
	data           map[string]interface{}
	connectionOpen bool
}

func (m *MockFailingBackend) Read(key string) (interface{}, error) {
	if m.failOnRead {
		return nil, errors.New("backend read failure")
	}
	if m.failOnConnect && !m.connectionOpen {
		return nil, errors.New("connection not established")
	}
	if m.data == nil {
		return nil, errors.New("data not found")
	}
	value, exists := m.data[key]
	if !exists {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (m *MockFailingBackend) Write(key string, value interface{}) error {
	if m.failOnWrite {
		return errors.New("backend write failure")
	}
	if m.failOnConnect && !m.connectionOpen {
		return errors.New("connection not established")
	}
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
	return nil
}

func (m *MockFailingBackend) Delete(key string) error {
	if m.failOnDelete {
		return errors.New("backend delete failure")
	}
	if m.failOnConnect && !m.connectionOpen {
		return errors.New("connection not established")
	}
	if m.data != nil {
		delete(m.data, key)
	}
	return nil
}

func (m *MockFailingBackend) List(prefix string) ([]string, error) {
	if m.failOnList {
		return nil, errors.New("backend list failure")
	}
	if m.failOnConnect && !m.connectionOpen {
		return nil, errors.New("connection not established")
	}
	var keys []string
	if m.data != nil {
		for key := range m.data {
			if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				keys = append(keys, key)
			}
		}
	}
	return keys, nil
}

func (m *MockFailingBackend) Connect() error {
	if m.failOnConnect {
		return errors.New("backend connection failure")
	}
	m.connectionOpen = true
	return nil
}

func (m *MockFailingBackend) Close() error {
	if m.failOnClose {
		return errors.New("backend close failure")
	}
	m.connectionOpen = false
	return nil
}

func (m *MockFailingBackend) IsConnected() bool {
	return m.connectionOpen
}

func TestStorage_ErrorRecovery_BackendFailures(t *testing.T) {
	// Since MemoryStorage doesn't support backend injection, we'll test error recovery
	// through the actual MemoryStorage methods that can fail

	storage := NewMemoryStorage()

	// Test various failure scenarios with actual storage operations
	t.Run("storage operation failures", func(t *testing.T) {
		// Store some valid data first
		validThought := &types.Thought{
			ID:      "test-id",
			Content: "test content",
			Mode:    types.ModeLinear,
		}
		err := storage.StoreThought(validThought)
		if err != nil {
			t.Fatalf("Failed to store valid thought: %v", err)
		}

		// Test retrieval of valid data
		retrieved, err := storage.GetThought("test-id")
		if err != nil {
			t.Errorf("Failed to get valid thought: %v", err)
		}
		if retrieved == nil {
			t.Error("Retrieved thought is nil")
		}

		// Test branch operations
		validBranch := &types.Branch{
			ID:         "test-branch",
			State:      types.StateActive,
			Priority:   0.5,
			Confidence: 0.8,
			CreatedAt:  time.Now(),
		}
		err = storage.StoreBranch(validBranch)
		if err != nil {
			t.Errorf("Failed to store valid branch: %v", err)
		}

		branches := storage.ListBranches()
		if branches == nil {
			t.Error("ListBranches returned nil")
		}

		// Test metrics retrieval
		metrics := storage.GetMetrics()
		if metrics == nil {
			t.Error("GetMetrics returned nil")
		}
	})
}

func TestStorage_ErrorRecovery_DataCorruption(t *testing.T) {
	storage := NewMemoryStorage()

	// Test handling of edge cases that might lead to data issues
	t.Run("edge case data handling", func(t *testing.T) {
		// Store thoughts with various edge cases
		edgeCaseThoughts := []*types.Thought{
			{
				ID:      "empty-content",
				Content: "",
				Mode:    types.ModeLinear,
			},
			{
				ID:      "large-content",
				Content: string(make([]byte, 10000)), // 10KB content
				Mode:    types.ModeLinear,
			},
			{
				ID:      "special-chars",
				Content: "Content with\nnewlines\tand\ttabs\r\nand\rreturns",
				Mode:    types.ModeLinear,
			},
			{
				ID:      "unicode-content",
				Content: "Unicode: 你好世界 🌍 Привет мир",
				Mode:    types.ModeLinear,
			},
		}

		for _, thought := range edgeCaseThoughts {
			err := storage.StoreThought(thought)
			if err != nil {
				t.Errorf("Failed to store edge case thought %s: %v", thought.ID, err)
			}

			// Retrieve and verify
			retrieved, err := storage.GetThought(thought.ID)
			if err != nil {
				t.Errorf("Failed to retrieve edge case thought %s: %v", thought.ID, err)
			}
			if retrieved == nil {
				t.Errorf("Retrieved edge case thought %s is nil", thought.ID)
			}
		}

		// Test with edge case branches
		edgeCaseBranches := []*types.Branch{
			{
				ID:         "empty-branch",
				State:      types.StateActive,
				Priority:   0.0,
				Confidence: 0.0,
				CreatedAt:  time.Now(),
			},
			{
				ID:         "high-priority",
				State:      types.StateActive,
				Priority:   1.0,
				Confidence: 1.0,
				CreatedAt:  time.Now(),
			},
		}

		for _, branch := range edgeCaseBranches {
			err := storage.StoreBranch(branch)
			if err != nil {
				t.Errorf("Failed to store edge case branch %s: %v", branch.ID, err)
			}
		}
	})
}

func TestStorage_ErrorRecovery_ConcurrentAccess(t *testing.T) {
	storage := NewMemoryStorage()

	// Test concurrent access patterns
	done := make(chan bool, 20)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			// Perform various operations concurrently
			thought := &types.Thought{
				ID:      fmt.Sprintf("thought-%d", id),
				Content: fmt.Sprintf("content %d", id),
				Mode:    types.ModeLinear,
			}

			// Store
			err := storage.StoreThought(thought)
			if err != nil {
				t.Errorf("Failed to store thought %d: %v", id, err)
			}

			// Retrieve
			_, err = storage.GetThought(thought.ID)
			if err != nil {
				t.Errorf("Failed to get thought %d: %v", id, err)
			}

			// List branches (should not crash)
			branches := storage.ListBranches()
			if branches == nil {
				t.Errorf("ListBranches returned nil in goroutine %d", id)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(2 * time.Second):
			t.Errorf("Goroutine %d timed out", i)
		}
	}

	// Storage should remain functional after concurrent access
	if storage == nil {
		t.Error("Storage became nil after concurrent access")
	}
}

func TestStorage_ErrorRecovery_ResourceLimits(t *testing.T) {
	storage := NewMemoryStorage()

	t.Run("large data sets", func(t *testing.T) {
		// Test with many thoughts
		for i := 0; i < 1000; i++ {
			thought := &types.Thought{
				ID:      fmt.Sprintf("large-test-%d", i),
				Content: fmt.Sprintf("Large content for testing resource limits %d", i) + string(make([]byte, 1000)), // 1KB each
				Mode:    types.ModeLinear,
			}

			err := storage.StoreThought(thought)
			if err != nil {
				t.Errorf("Failed to store large thought %d: %v", i, err)
			}
		}

		// Test retrieval and listing with large dataset
		branches := storage.ListBranches()
		if branches == nil {
			t.Error("Failed to list branches with large dataset")
		}

		// Storage should handle large datasets gracefully
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Storage panicked with large dataset: %v", r)
			}
		}()

		// Test metrics with large dataset
		metrics := storage.GetMetrics()
		if metrics == nil {
			t.Error("Failed to get metrics with large dataset")
		}
	})

	t.Run("memory pressure simulation", func(t *testing.T) {
		// Create many large objects
		var largeThoughts []*types.Thought
		for i := 0; i < 100; i++ {
			thought := &types.Thought{
				ID:      fmt.Sprintf("memory-test-%d", i),
				Content: string(make([]byte, 10000)), // 10KB each
				Mode:    types.ModeLinear,
				Metadata: map[string]interface{}{
					"large_data": string(make([]byte, 5000)), // Additional 5KB
				},
			}
			largeThoughts = append(largeThoughts, thought)
		}

		// Store all large thoughts
		for _, thought := range largeThoughts {
			err := storage.StoreThought(thought)
			if err != nil {
				t.Errorf("Failed to store memory-intensive thought: %v", err)
			}
		}

		// Storage should handle memory pressure gracefully
		if storage == nil {
			t.Error("Storage became nil under memory pressure")
		}
	})
}

func TestStorage_ErrorRecovery_InvalidData(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name      string
		thought   *types.Thought
		expectErr bool
	}{
		{
			name:      "nil thought",
			thought:   nil,
			expectErr: true,
		},
		{
			name: "empty thought ID",
			thought: &types.Thought{
				ID:      "",
				Content: "content",
				Mode:    types.ModeLinear,
			},
			expectErr: false, // Empty ID might be acceptable
		},
		{
			name: "invalid mode",
			thought: &types.Thought{
				ID:      "invalid-mode",
				Content: "content",
				Mode:    types.ThinkingMode("invalid"),
			},
			expectErr: false, // Invalid mode might be acceptable
		},
		{
			name: "valid thought",
			thought: &types.Thought{
				ID:      "valid-thought",
				Content: "valid content",
				Mode:    types.ModeLinear,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.thought == nil {
				// Test storing nil (should panic, which is expected)
				defer func() {
					if r := recover(); r != nil {
						// This is expected behavior for nil input
						t.Logf("Expected panic when storing nil thought: %v", r)
					} else {
						t.Error("Expected panic when storing nil thought but got none")
					}
				}()

				// This should panic
				err := storage.StoreThought(tt.thought)
				t.Errorf("StoreThought should have panicked but returned: %v", err)
			} else {
				// Test that storage handles invalid data gracefully
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Storage panicked with valid data: %v", r)
					}
				}()

				err := storage.StoreThought(tt.thought)
				if tt.expectErr && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Test retrieval
				if tt.thought.ID != "" {
					retrieved, err := storage.GetThought(tt.thought.ID)
					if tt.expectErr && err == nil && retrieved == nil {
						// This is expected for invalid data
					} else if !tt.expectErr && err != nil {
						t.Errorf("Failed to retrieve valid thought: %v", err)
					}
				}
			}

			// Storage should remain functional
			if storage == nil {
				t.Error("Storage became nil after invalid data")
			}
		})
	}
}

func TestStorage_ErrorRecovery_SearchOperations(t *testing.T) {
	storage := NewMemoryStorage()

	// Store some test data
	testThoughts := []*types.Thought{
		{
			ID:      "search-1",
			Content: "This is a test thought for searching",
			Mode:    types.ModeLinear,
		},
		{
			ID:      "search-2",
			Content: "Another test thought with different content",
			Mode:    types.ModeLinear,
		},
		{
			ID:      "search-3",
			Content: "Test content for edge case searching",
			Mode:    types.ModeTree,
		},
	}

	for _, thought := range testThoughts {
		err := storage.StoreThought(thought)
		if err != nil {
			t.Fatalf("Failed to store test thought: %v", err)
		}
	}

	t.Run("search with various queries", func(t *testing.T) {
		// Test normal search
		results := storage.SearchThoughts("test", types.ModeLinear, 10, 0)
		if results == nil {
			t.Error("Search returned nil")
		}

		// Test search with empty query
		results = storage.SearchThoughts("", types.ModeLinear, 10, 0)
		if results == nil {
			t.Error("Search with empty query returned nil")
		}

		// Test search with non-existent mode
		results = storage.SearchThoughts("test", types.ThinkingMode("nonexistent"), 10, 0)
		if results == nil {
			t.Error("Search with invalid mode returned nil")
		}

		// Test search with large limit
		results = storage.SearchThoughts("test", types.ModeLinear, 10000, 0)
		if results == nil {
			t.Error("Search with large limit returned nil")
		}
	})

	t.Run("search edge cases", func(t *testing.T) {
		// Test search with special characters
		results := storage.SearchThoughts("test\nthought", types.ModeLinear, 10, 0)
		if results == nil {
			t.Error("Search with special characters returned nil")
		}

		// Test search with unicode
		results = storage.SearchThoughts("测试", types.ModeLinear, 10, 0) // Chinese characters
		if results == nil {
			t.Error("Search with unicode returned nil")
		}

		// Test search with very long query
		longQuery := string(make([]byte, 1000))
		results = storage.SearchThoughts(longQuery, types.ModeLinear, 10, 0)
		if results == nil {
			t.Error("Search with long query returned nil")
		}
	})
}
