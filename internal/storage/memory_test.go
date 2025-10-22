package storage

import (
	"sync"
	"testing"
	"time"
	"unified-thinking/internal/types"
)

func TestNewMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	if storage == nil {
		t.Fatal("NewMemoryStorage returned nil")
	}

	if storage.thoughts == nil {
		t.Error("thoughts map not initialized")
	}

	if storage.branches == nil {
		t.Error("branches map not initialized")
	}

	if storage.insights == nil {
		t.Error("insights map not initialized")
	}

	if storage.validations == nil {
		t.Error("validations map not initialized")
	}

	if storage.relationships == nil {
		t.Error("relationships map not initialized")
	}
}

func TestStoreThought(t *testing.T) {
	tests := []struct {
		name    string
		thought *types.Thought
		wantErr bool
	}{
		{
			name: "store thought with ID",
			thought: &types.Thought{
				ID:         "test-1",
				Content:    "Test content",
				Mode:       types.ModeLinear,
				Type:       "observation",
				Confidence: 0.8,
				Timestamp:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "store thought without ID (auto-generated)",
			thought: &types.Thought{
				Content:    "Test content 2",
				Mode:       types.ModeTree,
				Type:       "analysis",
				Confidence: 0.9,
				Timestamp:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "store thought with metadata",
			thought: &types.Thought{
				Content:    "Test with metadata",
				Mode:       types.ModeDivergent,
				Confidence: 0.7,
				KeyPoints:  []string{"point1", "point2"},
				Metadata:   map[string]interface{}{"key": "value"},
				Timestamp:  time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemoryStorage()
			err := storage.StoreThought(tt.thought)

			if (err != nil) != tt.wantErr {
				t.Errorf("StoreThought() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.thought.ID == "" {
				t.Error("Thought ID was not generated")
			}

			if !tt.wantErr {
				// Verify thought was stored
				retrieved, err := storage.GetThought(tt.thought.ID)
				if err != nil {
					t.Errorf("Failed to retrieve stored thought: %v", err)
				}
				if retrieved.Content != tt.thought.Content {
					t.Errorf("Retrieved thought content = %v, want %v", retrieved.Content, tt.thought.Content)
				}
			}
		})
	}
}

func TestGetThought(t *testing.T) {
	storage := NewMemoryStorage()

	// Store a thought
	thought := &types.Thought{
		ID:         "test-123",
		Content:    "Test thought",
		Mode:       types.ModeLinear,
		Type:       "test",
		Confidence: 0.8,
		KeyPoints:  []string{"key1", "key2"},
		Metadata:   map[string]interface{}{"test": "data"},
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

			if !tt.wantErr {
				if retrieved.ID != thought.ID {
					t.Errorf("Retrieved thought ID = %v, want %v", retrieved.ID, thought.ID)
				}
				if retrieved.Content != thought.Content {
					t.Errorf("Retrieved thought content = %v, want %v", retrieved.Content, thought.Content)
				}

				// Verify it's a copy - modify the retrieved thought's slice/map
				if len(retrieved.KeyPoints) > 0 {
					retrieved.KeyPoints[0] = "modified"
					retrieved2, _ := storage.GetThought(tt.id)
					if retrieved2.KeyPoints[0] == "modified" {
						t.Error("GetThought should return a deep copy of KeyPoints")
					}
				}
				if len(retrieved.Metadata) > 0 {
					retrieved.Metadata["test"] = "modified"
					retrieved2, _ := storage.GetThought(tt.id)
					if val, ok := retrieved2.Metadata["test"]; ok && val == "modified" {
						t.Error("GetThought should return a deep copy of Metadata")
					}
				}
			}
		})
	}
}

func TestStoreBranch(t *testing.T) {
	tests := []struct {
		name    string
		branch  *types.Branch
		wantErr bool
	}{
		{
			name: "store branch with ID",
			branch: &types.Branch{
				ID:         "branch-1",
				State:      types.StateActive,
				Priority:   1.0,
				Confidence: 0.8,
				Thoughts:   []*types.Thought{},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "store branch without ID (auto-generated)",
			branch: &types.Branch{
				State:      types.StateSuspended,
				Priority:   0.5,
				Confidence: 0.6,
				Thoughts:   []*types.Thought{},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemoryStorage()
			err := storage.StoreBranch(tt.branch)

			if (err != nil) != tt.wantErr {
				t.Errorf("StoreBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.branch.ID == "" {
				t.Error("Branch ID was not generated")
			}

			if !tt.wantErr {
				// First branch should be set as active
				activeBranch, err := storage.GetActiveBranch()
				if err != nil {
					t.Errorf("Failed to get active branch: %v", err)
				}
				if activeBranch.ID != tt.branch.ID {
					t.Errorf("Active branch ID = %v, want %v", activeBranch.ID, tt.branch.ID)
				}
			}
		})
	}
}

func TestGetBranch(t *testing.T) {
	storage := NewMemoryStorage()

	// Store a branch
	branch := &types.Branch{
		ID:         "branch-test",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_ = storage.StoreBranch(branch)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "get existing branch",
			id:      "branch-test",
			wantErr: false,
		},
		{
			name:    "get non-existent branch",
			id:      "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := storage.GetBranch(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if retrieved.ID != branch.ID {
					t.Errorf("Retrieved branch ID = %v, want %v", retrieved.ID, branch.ID)
				}
			}
		})
	}
}

func TestListBranches(t *testing.T) {
	storage := NewMemoryStorage()

	// Store multiple branches
	branch1 := &types.Branch{ID: "b1", State: types.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	branch2 := &types.Branch{ID: "b2", State: types.StateSuspended, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	branch3 := &types.Branch{ID: "b3", State: types.StateCompleted, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	_ = storage.StoreBranch(branch1)
	_ = storage.StoreBranch(branch2)
	_ = storage.StoreBranch(branch3)

	branches := storage.ListBranches()

	if len(branches) != 3 {
		t.Errorf("ListBranches() returned %d branches, want 3", len(branches))
	}

	// Verify all branches are present
	branchIDs := make(map[string]bool)
	for _, b := range branches {
		branchIDs[b.ID] = true
	}

	if !branchIDs["b1"] || !branchIDs["b2"] || !branchIDs["b3"] {
		t.Error("ListBranches() did not return all stored branches")
	}
}

func TestGetActiveBranch(t *testing.T) {
	storage := NewMemoryStorage()

	// Test with no active branch
	_, err := storage.GetActiveBranch()
	if err == nil {
		t.Error("GetActiveBranch() should return error when no active branch")
	}

	// Store a branch (becomes active)
	branch := &types.Branch{
		ID:        "active-test",
		State:     types.StateActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = storage.StoreBranch(branch)

	active, err := storage.GetActiveBranch()
	if err != nil {
		t.Errorf("GetActiveBranch() error = %v", err)
	}

	if active.ID != "active-test" {
		t.Errorf("Active branch ID = %v, want active-test", active.ID)
	}
}

func TestSetActiveBranch(t *testing.T) {
	storage := NewMemoryStorage()

	// Store two branches
	branch1 := &types.Branch{ID: "b1", State: types.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	branch2 := &types.Branch{ID: "b2", State: types.StateSuspended, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	_ = storage.StoreBranch(branch1)
	_ = storage.StoreBranch(branch2)

	tests := []struct {
		name     string
		branchID string
		wantErr  bool
	}{
		{
			name:     "set existing branch as active",
			branchID: "b2",
			wantErr:  false,
		},
		{
			name:     "set non-existent branch as active",
			branchID: "non-existent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.SetActiveBranch(tt.branchID)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetActiveBranch() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				active, _ := storage.GetActiveBranch()
				if active.ID != tt.branchID {
					t.Errorf("Active branch after set = %v, want %v", active.ID, tt.branchID)
				}
			}
		})
	}
}

func TestStoreInsight(t *testing.T) {
	storage := NewMemoryStorage()

	insight := &types.Insight{
		Type:               types.InsightObservation,
		Content:            "Test insight",
		ApplicabilityScore: 0.9,
		CreatedAt:          time.Now(),
	}

	err := storage.StoreInsight(insight)
	if err != nil {
		t.Errorf("StoreInsight() error = %v", err)
	}

	if insight.ID == "" {
		t.Error("Insight ID was not generated")
	}
}

func TestStoreValidation(t *testing.T) {
	storage := NewMemoryStorage()

	validation := &types.Validation{
		ThoughtID: "thought-123",
		IsValid:   true,
		Reason:    "Test reason",
		CreatedAt: time.Now(),
	}

	err := storage.StoreValidation(validation)
	if err != nil {
		t.Errorf("StoreValidation() error = %v", err)
	}

	if validation.ID == "" {
		t.Error("Validation ID was not generated")
	}
}

func TestStoreRelationship(t *testing.T) {
	storage := NewMemoryStorage()

	relationship := &types.Relationship{
		FromStateID: "state-1",
		ToStateID:   "state-2",
		Type:        "causal",
		CreatedAt:   time.Now(),
	}

	err := storage.StoreRelationship(relationship)
	if err != nil {
		t.Errorf("StoreRelationship() error = %v", err)
	}

	if relationship.ID == "" {
		t.Error("Relationship ID was not generated")
	}
}

func TestSearchThoughts(t *testing.T) {
	storage := NewMemoryStorage()

	// Store multiple thoughts
	thoughts := []*types.Thought{
		{Content: "Linear thinking about cats", Mode: types.ModeLinear, Timestamp: time.Now()},
		{Content: "Tree thinking about dogs", Mode: types.ModeTree, Timestamp: time.Now()},
		{Content: "Linear thinking about birds", Mode: types.ModeLinear, Timestamp: time.Now()},
		{Content: "Divergent thinking about cats", Mode: types.ModeDivergent, Timestamp: time.Now()},
	}

	for _, th := range thoughts {
		_ = storage.StoreThought(th)
	}

	tests := []struct {
		name      string
		query     string
		mode      types.ThinkingMode
		wantCount int
	}{
		{
			name:      "search all thoughts",
			query:     "",
			mode:      "",
			wantCount: 4,
		},
		{
			name:      "search by mode linear",
			query:     "",
			mode:      types.ModeLinear,
			wantCount: 2,
		},
		{
			name:      "search by mode tree",
			query:     "",
			mode:      types.ModeTree,
			wantCount: 1,
		},
		{
			name:      "search by content cats",
			query:     "cats",
			mode:      "",
			wantCount: 2,
		},
		{
			name:      "search by content and mode",
			query:     "cats",
			mode:      types.ModeLinear,
			wantCount: 1,
		},
		{
			name:      "search no matches",
			query:     "elephants",
			mode:      "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := storage.SearchThoughts(tt.query, tt.mode, 0, 0)

			if len(results) != tt.wantCount {
				t.Errorf("SearchThoughts() returned %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	storage := NewMemoryStorage()

	// Test concurrent writes
	t.Run("concurrent thought storage", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				thought := &types.Thought{
					Content:   "Concurrent thought",
					Mode:      types.ModeLinear,
					Timestamp: time.Now(),
				}
				_ = storage.StoreThought(thought)
			}(i)
		}

		wg.Wait()

		// Verify all thoughts were stored
		results := storage.SearchThoughts("", types.ModeLinear, 0, 0)
		if len(results) != numGoroutines {
			t.Errorf("Expected %d thoughts, got %d", numGoroutines, len(results))
		}
	})

	// Test concurrent reads
	t.Run("concurrent thought retrieval", func(t *testing.T) {
		thought := &types.Thought{
			ID:        "concurrent-test",
			Content:   "Test concurrent reads",
			Mode:      types.ModeLinear,
			Timestamp: time.Now(),
		}
		_ = storage.StoreThought(thought)

		var wg sync.WaitGroup
		numGoroutines := 20
		errors := make(chan error, numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				_, err := storage.GetThought("concurrent-test")
				if err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("Concurrent read error: %v", err)
		}
	})
}

func TestDataIsolation(t *testing.T) {
	storage := NewMemoryStorage()

	// Store a thought with key points and metadata
	original := &types.Thought{
		ID:        "isolation-test",
		Content:   "Original content",
		Mode:      types.ModeLinear,
		KeyPoints: []string{"point1", "point2"},
		Metadata:  map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
	}

	_ = storage.StoreThought(original)

	// Retrieve and modify
	retrieved, _ := storage.GetThought("isolation-test")
	retrieved.Content = "Modified content"
	retrieved.KeyPoints[0] = "modified"
	retrieved.Metadata["key"] = "modified"

	// Retrieve again and verify original is unchanged
	second, _ := storage.GetThought("isolation-test")

	if second.Content != "Original content" {
		t.Error("Content was modified through retrieved copy")
	}

	if second.KeyPoints[0] != "point1" {
		t.Error("KeyPoints was modified through retrieved copy")
	}

	if second.Metadata["key"] != "value" {
		t.Error("Metadata was modified through retrieved copy")
	}
}
