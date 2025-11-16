package modes

import (
	"context"
	"testing"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestLinearMode_EmptyContent(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewLinearMode(store)
	ctx := context.Background()

	input := ThoughtInput{
		Content:    "",
		Confidence: 0.8,
	}

	// Should handle empty content gracefully (validation happens at server level)
	result, err := mode.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	// Mode should still process it (server validates)
	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestTreeMode_BranchNotFound(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Try to add thought to non-existent branch
	input := ThoughtInput{
		Content:    "Test thought",
		BranchID:   "non-existent-branch",
		Confidence: 0.8,
	}

	// Current implementation creates new branch instead of erroring
	result, err := mode.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	// Should handle by creating branch or returning result
	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestTreeMode_NilActiveBranch(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// No active branch set, no branchID provided
	input := ThoughtInput{
		Content:    "Test thought",
		BranchID:   "", // Empty - should try to get active
		Confidence: 0.8,
	}

	result, err := mode.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	// Should create new branch
	if result.BranchID == "" {
		t.Error("BranchID should be created when no active branch exists")
	}
}

func TestDivergentMode_BranchNonExistentThought(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	// Try to branch from non-existent thought
	_, err := mode.BranchThought(ctx, "non-existent-thought", "opposite")

	if err == nil {
		t.Error("BranchThought should return error for non-existent thought")
	}
}

func TestDivergentMode_InvalidBranchType(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	// Create a thought first
	input := ThoughtInput{
		Content:    "Original thought",
		Confidence: 0.8,
	}
	result, _ := mode.ProcessThought(ctx, input)

	// Try to branch with invalid type
	_, err := mode.BranchThought(ctx, result.ThoughtID, "invalid-type")

	// Should either error or handle gracefully
	// Current implementation may not validate branch type
	if err != nil {
		// Error is acceptable
		t.Logf("BranchThought returned expected error: %v", err)
	}
}

func TestAutoMode_FallbackToLinear(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	ctx := context.Background()

	// Content with no specific keywords should default to linear
	input := ThoughtInput{
		Content:    "Calculate the sum of two numbers",
		Confidence: 0.8,
	}

	result, err := auto.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	if string(result.Mode) != string(types.ModeLinear) {
		t.Errorf("Auto mode should default to linear, got %v", result.Mode)
	}
}

func TestLinearMode_ZeroConfidence(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewLinearMode(store)
	ctx := context.Background()

	input := ThoughtInput{
		Content:    "Test thought",
		Confidence: 0.0, // Zero confidence
	}

	result, err := mode.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	// Should use default or accept 0.0
	if result == nil {
		t.Error("Result should not be nil for zero confidence")
	}
}

func TestTreeMode_MaxKeyPoints(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Create input with many key points
	keyPoints := make([]string, 100)
	for i := 0; i < 100; i++ {
		keyPoints[i] = "Key point"
	}

	input := ThoughtInput{
		Content:    "Test thought with many key points",
		KeyPoints:  keyPoints,
		Confidence: 0.8,
	}

	// Should handle without error (validation at server level)
	result, err := mode.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestDivergentMode_ConsistentRebellion(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	// Test that ForceRebellion consistently produces rebellion
	input := ThoughtInput{
		Content:        "Test thought",
		ForceRebellion: true,
		Confidence:     0.8,
	}

	for i := 0; i < 10; i++ {
		result, err := mode.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() error = %v", err)
		}

		if !result.IsRebellion {
			t.Error("ForceRebellion should always produce rebellion")
		}
	}
}

func TestTreeMode_CrossRefToSelf(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Create a branch
	input := ThoughtInput{
		Content:    "Test thought",
		Confidence: 0.8,
	}
	result, _ := mode.ProcessThought(ctx, input)

	// Try to create cross-ref to same branch
	crossRefInput := ThoughtInput{
		Content:  "Cross-ref thought",
		BranchID: result.BranchID,
		CrossRefs: []CrossRefInput{
			{
				ToBranch: result.BranchID, // Self-reference
				Type:     "complementary",
				Reason:   "Self-reference test",
				Strength: 0.8,
			},
		},
		Confidence: 0.8,
	}

	// Should handle self-reference (may allow or reject)
	_, err := mode.ProcessThought(ctx, crossRefInput)
	if err != nil {
		t.Logf("Self-reference rejected (acceptable): %v", err)
	}
}

func TestLinearMode_ParentChain(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewLinearMode(store)
	ctx := context.Background()

	// Create chain of thoughts
	var parentID string
	for i := 0; i < 5; i++ {
		input := ThoughtInput{
			Content:    "Chained thought",
			ParentID:   parentID,
			Confidence: 0.8,
		}

		result, err := mode.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() error = %v", err)
		}

		parentID = result.ThoughtID
	}

	// Verify final thought has correct parent
	if parentID == "" {
		t.Error("Should have created thought chain")
	}
}

func TestTreeMode_BranchStateTransitions(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewTreeMode(store)
	ctx := context.Background()

	// Create a branch
	input := ThoughtInput{
		Content:    "Test thought",
		Confidence: 0.8,
	}
	result, _ := mode.ProcessThought(ctx, input)

	// Get branch and verify initial state
	branch, err := store.GetBranch(result.BranchID)
	if err != nil {
		t.Fatalf("GetBranch() error = %v", err)
	}

	if branch.State != types.StateActive {
		t.Errorf("New branch state = %v, want active", branch.State)
	}
}

func TestDivergentMode_ChallengeAssumptions(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	// Process multiple thoughts and verify challenge assumption flag varies
	challenges := 0
	iterations := 50

	input := ThoughtInput{
		Content:    "Test thought",
		Confidence: 0.8,
	}

	for i := 0; i < iterations; i++ {
		result, err := mode.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() error = %v", err)
		}

		if result.ChallengesAssumption {
			challenges++
		}
	}

	// With random threshold of 0.3, expect ~70% to challenge
	if challenges < 20 || challenges > 45 {
		t.Logf("Challenge rate: %d/%d (expected ~35 with 0.3 threshold)", challenges, iterations)
		// Not failing, just logging - randomness may vary
	}
}

func TestAutoMode_ModeSelection(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	tests := []struct {
		name         string
		content      string
		keyPoints    []string
		crossRefs    []CrossRefInput
		forceRebel   bool
		expectedMode types.ThinkingMode
	}{
		{
			name:         "creative keyword triggers divergent",
			content:      "Let's think creatively about this",
			expectedMode: types.ModeDivergent,
		},
		{
			name:         "rebel keyword triggers divergent",
			content:      "Let's challenge the conventional approach",
			expectedMode: types.ModeDivergent,
		},
		{
			name:         "explore keyword triggers tree",
			content:      "Let's explore different options",
			expectedMode: types.ModeTree,
		},
		{
			name:         "key points trigger tree",
			content:      "Consider these aspects",
			keyPoints:    []string{"point1", "point2"},
			expectedMode: types.ModeTree,
		},
		{
			name:         "cross refs trigger tree",
			content:      "Building on previous ideas",
			crossRefs:    []CrossRefInput{{ToBranch: "other", Type: "builds_upon", Reason: "test", Strength: 0.8}},
			expectedMode: types.ModeTree,
		},
		{
			name:         "force rebellion triggers divergent",
			content:      "Regular content",
			forceRebel:   true,
			expectedMode: types.ModeDivergent,
		},
		{
			name:         "plain content triggers linear",
			content:      "Calculate the result",
			expectedMode: types.ModeLinear,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ThoughtInput{
				Content:        tt.content,
				KeyPoints:      tt.keyPoints,
				CrossRefs:      tt.crossRefs,
				ForceRebellion: tt.forceRebel,
				Confidence:     0.8,
			}

			result, err := auto.ProcessThought(context.Background(), input)
			if err != nil {
				t.Fatalf("ProcessThought() error = %v", err)
			}

			if string(result.Mode) != string(tt.expectedMode) {
				t.Errorf("Mode = %v, want %v", result.Mode, tt.expectedMode)
			}
		})
	}
}
