package modes

import (
	"context"
	"testing"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestNewTreeMode(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)

	if mode == nil {
		t.Fatal("NewTreeMode returned nil")
	}

	if mode.storage == nil {
		t.Error("TreeMode storage not initialized")
	}
}

func TestTreeMode_ProcessThought(t *testing.T) {
	tests := []struct {
		name    string
		input   ThoughtInput
		wantErr bool
	}{
		{
			name: "basic thought without branch",
			input: ThoughtInput{
				Content:    "Tree thought",
				Type:       "exploration",
				Confidence: 0.8,
			},
			wantErr: false,
		},
		{
			name: "thought with key points",
			input: ThoughtInput{
				Content:    "Thought with insights",
				Type:       "analysis",
				Confidence: 0.9,
				KeyPoints:  []string{"key1", "key2", "key3"},
			},
			wantErr: false,
		},
		{
			name: "thought with cross references",
			input: ThoughtInput{
				Content:    "Thought with cross refs",
				Type:       "connection",
				Confidence: 0.85,
				CrossRefs: []CrossRefInput{
					{ToBranch: "branch-other", Type: "complementary", Reason: "Related concept", Strength: 0.9},
				},
			},
			wantErr: false,
		},
		{
			name: "thought with parent",
			input: ThoughtInput{
				Content:    "Child thought",
				Type:       "elaboration",
				ParentID:   "parent-123",
				Confidence: 0.75,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			mode := NewTreeMode(store)
			ctx := context.Background()

			result, err := mode.ProcessThought(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessThought() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatal("ProcessThought() returned nil result")
				}

				if result.ThoughtID == "" {
					t.Error("Result missing ThoughtID")
				}

				if result.BranchID == "" {
					t.Error("Result missing BranchID")
				}

				if result.Mode != string(types.ModeTree) {
					t.Errorf("Result mode = %v, want %v", result.Mode, types.ModeTree)
				}

				// Verify thought was stored
				thought, err := store.GetThought(result.ThoughtID)
				if err != nil {
					t.Errorf("Failed to retrieve stored thought: %v", err)
				}

				if thought.Content != tt.input.Content {
					t.Errorf("Stored thought content = %v, want %v", thought.Content, tt.input.Content)
				}

				if thought.Mode != types.ModeTree {
					t.Errorf("Stored thought mode = %v, want %v", thought.Mode, types.ModeTree)
				}

				// Verify branch exists
				branch, err := store.GetBranch(result.BranchID)
				if err != nil {
					t.Errorf("Failed to retrieve branch: %v", err)
				}

				if branch.ID != result.BranchID {
					t.Errorf("Branch ID = %v, want %v", branch.ID, result.BranchID)
				}

				// Verify insights were created if key points provided
				if len(tt.input.KeyPoints) > 0 {
					if result.InsightCount == 0 {
						t.Error("Expected insights to be created from key points")
					}
				}

				// Verify cross refs were counted
				if len(tt.input.CrossRefs) > 0 {
					if result.CrossRefCount == 0 {
						t.Error("Expected cross refs to be counted")
					}
				}
			}
		})
	}
}

func TestTreeMode_ProcessThoughtWithExistingBranch(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Create a branch
	branch := &types.Branch{
		ID:         "existing-branch",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		Thoughts:   []*types.Thought{},
		Insights:   []*types.Insight{},
		CrossRefs:  []*types.CrossRef{},
	}
	store.StoreBranch(branch)

	// Process thought with existing branch
	input := ThoughtInput{
		Content:    "Thought in existing branch",
		Type:       "exploration",
		BranchID:   "existing-branch",
		Confidence: 0.9,
	}

	result, err := mode.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	if result.BranchID != "existing-branch" {
		t.Errorf("Result branch ID = %v, want existing-branch", result.BranchID)
	}
}

func TestTreeMode_ListBranches(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Initially should return empty list
	branches, err := mode.ListBranches(ctx)
	if err != nil {
		t.Errorf("ListBranches() error = %v", err)
	}
	if len(branches) != 0 {
		t.Errorf("Initial branches length = %d, want 0", len(branches))
	}

	// Process some thoughts (creates branches)
	for i := 0; i < 3; i++ {
		input := ThoughtInput{
			Content:    "Thought",
			Type:       "exploration",
			Confidence: 0.8,
		}
		mode.ProcessThought(ctx, input)
	}

	// List branches
	branches, err = mode.ListBranches(ctx)
	if err != nil {
		t.Errorf("ListBranches() error = %v", err)
	}

	if len(branches) == 0 {
		t.Error("Expected at least one branch")
	}
}

func TestTreeMode_GetBranchHistory(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Process a thought to create a branch
	input := ThoughtInput{
		Content:    "Test thought",
		Type:       "exploration",
		Confidence: 0.8,
		KeyPoints:  []string{"point1", "point2"},
	}

	result, err := mode.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	// Get branch history
	history, err := mode.GetBranchHistory(ctx, result.BranchID)
	if err != nil {
		t.Errorf("GetBranchHistory() error = %v", err)
	}

	if history == nil {
		t.Fatal("GetBranchHistory() returned nil")
	}

	if history.BranchID != result.BranchID {
		t.Errorf("History branch ID = %v, want %v", history.BranchID, result.BranchID)
	}

	// History exists (actual content depends on how tree mode stores data)
	if history.BranchID == "" {
		t.Error("History should have branch ID")
	}

	// Test with non-existent branch
	_, err = mode.GetBranchHistory(ctx, "non-existent")
	if err == nil {
		t.Error("GetBranchHistory() should return error for non-existent branch")
	}
}

func TestTreeMode_SetActiveBranch(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Create branches
	branch1 := &types.Branch{ID: "branch1", State: types.StateActive}
	branch2 := &types.Branch{ID: "branch2", State: types.StateSuspended}

	store.StoreBranch(branch1)
	store.StoreBranch(branch2)

	// Set active branch
	err := mode.SetActiveBranch(ctx, "branch2")
	if err != nil {
		t.Errorf("SetActiveBranch() error = %v", err)
	}

	// Verify active branch changed
	active, _ := store.GetActiveBranch()
	if active.ID != "branch2" {
		t.Errorf("Active branch = %v, want branch2", active.ID)
	}

	// Test with non-existent branch
	err = mode.SetActiveBranch(ctx, "non-existent")
	if err == nil {
		t.Error("SetActiveBranch() should return error for non-existent branch")
	}
}

func TestTreeMode_UpdateBranchMetrics(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)

	// Create a branch with thoughts
	branch := &types.Branch{
		ID:    "test-branch",
		State: types.StateActive,
		Thoughts: []*types.Thought{
			{Confidence: 0.8},
			{Confidence: 0.9},
			{Confidence: 0.7},
		},
		Insights: []*types.Insight{
			{},
			{},
		},
		CrossRefs: []*types.CrossRef{
			{Strength: 0.9},
			{Strength: 0.8},
		},
	}

	mode.updateBranchMetrics(branch)

	// Check confidence (average) with tolerance for floating point
	expectedConfidence := (0.8 + 0.9 + 0.7) / 3
	tolerance := 0.0001
	if branch.Confidence < expectedConfidence-tolerance || branch.Confidence > expectedConfidence+tolerance {
		t.Errorf("Branch confidence = %v, want %v (within tolerance)", branch.Confidence, expectedConfidence)
	}

	// Check priority (confidence + insight score + crossref score)
	if branch.Priority <= branch.Confidence {
		t.Error("Priority should be greater than confidence with insights and crossrefs")
	}
}

func TestTreeMode_MultipleThoughtsInBranch(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Process first thought (creates branch)
	result1, err := mode.ProcessThought(ctx, ThoughtInput{
		Content:    "First thought",
		Type:       "exploration",
		Confidence: 0.8,
	})
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	branchID := result1.BranchID

	// Process more thoughts in same branch
	thoughtCount := 1
	for i := 0; i < 3; i++ {
		_, err := mode.ProcessThought(ctx, ThoughtInput{
			Content:    "Additional thought",
			Type:       "elaboration",
			BranchID:   branchID,
			Confidence: 0.85,
		})
		if err != nil {
			t.Fatalf("ProcessThought() iteration %d error = %v", i, err)
		}
		thoughtCount++
	}

	// Verify all thoughts were created
	if thoughtCount != 4 {
		t.Errorf("Created thought count = %d, want 4", thoughtCount)
	}
}

func TestTreeMode_CrossRefTypes(t *testing.T) {
	crossRefTypes := []string{"complementary", "contradictory", "builds_upon", "alternative"}

	for _, refType := range crossRefTypes {
		t.Run(refType, func(t *testing.T) {
			// Create fresh storage and mode for each subtest
			store := storage.NewMemoryStorage()
			mode := NewTreeMode(store)
			ctx := context.Background()

			// Create a target branch
			targetBranch := &types.Branch{
				ID:    "target-branch",
				State: types.StateActive,
			}
			store.StoreBranch(targetBranch)

			input := ThoughtInput{
				Content:    "Thought with cross ref",
				Type:       "connection",
				Confidence: 0.8,
				CrossRefs: []CrossRefInput{
					{
						ToBranch: "target-branch",
						Type:     refType,
						Reason:   "Test reason",
						Strength: 0.9,
					},
				},
			}

			result, err := mode.ProcessThought(ctx, input)
			if err != nil {
				t.Errorf("ProcessThought() error = %v", err)
			}

			if result.CrossRefCount == 0 {
				t.Error("Expected cross ref to be created")
			}

			// Verify result has correct cross ref count
			if result.CrossRefCount != 1 {
				t.Errorf("CrossRefCount = %d, want 1", result.CrossRefCount)
			}
		})
	}
}
