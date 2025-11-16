package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"unified-thinking/internal/types"
)

func TestAppendThoughtToBranch(t *testing.T) {
	store := NewMemoryStorage()

	// Create a branch
	branch := &types.Branch{
		ID:        "test-branch",
		State:     types.StateActive,
		Thoughts:  make([]*types.Thought, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreBranch(branch); err != nil {
		t.Fatalf("StoreBranch() error = %v", err)
	}

	// Create a thought
	thought := &types.Thought{
		ID:         "test-thought",
		Content:    "Test content",
		Mode:       types.ModeLinear,
		BranchID:   "test-branch",
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}
	if err := store.StoreThought(thought); err != nil {
		t.Fatalf("StoreThought() error = %v", err)
	}

	// Test append
	err := store.AppendThoughtToBranch("test-branch", thought)
	if err != nil {
		t.Fatalf("AppendThoughtToBranch() error = %v", err)
	}

	// Verify thought was appended
	retrieved, err := store.GetBranch("test-branch")
	if err != nil {
		t.Fatalf("GetBranch() error = %v", err)
	}

	if len(retrieved.Thoughts) != 1 {
		t.Errorf("Branch should have 1 thought, got %d", len(retrieved.Thoughts))
	}

	if retrieved.Thoughts[0].ID != "test-thought" {
		t.Errorf("Thought ID = %v, want test-thought", retrieved.Thoughts[0].ID)
	}
}

func TestAppendThoughtToBranch_Concurrency(t *testing.T) {
	store := NewMemoryStorage()

	branch := &types.Branch{
		ID:        "concurrent-branch",
		State:     types.StateActive,
		Thoughts:  make([]*types.Thought, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = store.StoreBranch(branch)

	var wg sync.WaitGroup
	numGoroutines := 20

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			thought := &types.Thought{
				ID:         fmt.Sprintf("thought-%d", id),
				Content:    "Concurrent thought",
				Mode:       types.ModeLinear,
				BranchID:   "concurrent-branch",
				Confidence: 0.8,
				Timestamp:  time.Now(),
			}
			_ = store.StoreThought(thought)
			store.AppendThoughtToBranch("concurrent-branch", thought)
		}(i)
	}

	wg.Wait()

	// Verify all thoughts were appended
	branch, _ = store.GetBranch("concurrent-branch")
	if len(branch.Thoughts) != numGoroutines {
		t.Errorf("Branch should have %d thoughts, got %d", numGoroutines, len(branch.Thoughts))
	}
}

func TestAppendInsightToBranch(t *testing.T) {
	store := NewMemoryStorage()

	branch := &types.Branch{
		ID:        "insight-branch",
		State:     types.StateActive,
		Insights:  make([]*types.Insight, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = store.StoreBranch(branch)

	insight := &types.Insight{
		ID:                 "test-insight",
		Content:            "Test insight",
		Type:               types.InsightObservation,
		ApplicabilityScore: 0.9,
		CreatedAt:          time.Now(),
		Context:            []string{},
		SupportingEvidence: make(map[string]interface{}),
	}
	store.StoreInsight(insight)

	err := store.AppendInsightToBranch("insight-branch", insight)
	if err != nil {
		t.Fatalf("AppendInsightToBranch() error = %v", err)
	}

	retrieved, _ := store.GetBranch("insight-branch")
	if len(retrieved.Insights) != 1 {
		t.Errorf("Branch should have 1 insight, got %d", len(retrieved.Insights))
	}
}

func TestAppendCrossRefToBranch(t *testing.T) {
	store := NewMemoryStorage()

	branch := &types.Branch{
		ID:        "crossref-branch",
		State:     types.StateActive,
		CrossRefs: make([]*types.CrossRef, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = store.StoreBranch(branch)

	crossRef := &types.CrossRef{
		FromBranch:  "crossref-branch",
		ToBranch:    "other-branch",
		Type:        types.CrossRefComplementary,
		Reason:      "Test cross-reference",
		Strength:    0.8,
		TouchPoints: []types.TouchPoint{},
		CreatedAt:   time.Now(),
	}

	err := store.AppendCrossRefToBranch("crossref-branch", crossRef)
	if err != nil {
		t.Fatalf("AppendCrossRefToBranch() error = %v", err)
	}

	retrieved, _ := store.GetBranch("crossref-branch")
	if len(retrieved.CrossRefs) != 1 {
		t.Errorf("Branch should have 1 cross-ref, got %d", len(retrieved.CrossRefs))
	}
}

func TestUpdateBranchPriority(t *testing.T) {
	store := NewMemoryStorage()

	branch := &types.Branch{
		ID:        "priority-branch",
		State:     types.StateActive,
		Priority:  1.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = store.StoreBranch(branch)

	err := store.UpdateBranchPriority("priority-branch", 2.5)
	if err != nil {
		t.Fatalf("UpdateBranchPriority() error = %v", err)
	}

	retrieved, _ := store.GetBranch("priority-branch")
	if retrieved.Priority != 2.5 {
		t.Errorf("Priority = %v, want 2.5", retrieved.Priority)
	}
}

func TestUpdateBranchConfidence(t *testing.T) {
	store := NewMemoryStorage()

	branch := &types.Branch{
		ID:         "confidence-branch",
		State:      types.StateActive,
		Confidence: 0.5,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_ = store.StoreBranch(branch)

	err := store.UpdateBranchConfidence("confidence-branch", 0.95)
	if err != nil {
		t.Fatalf("UpdateBranchConfidence() error = %v", err)
	}

	retrieved, _ := store.GetBranch("confidence-branch")
	if retrieved.Confidence != 0.95 {
		t.Errorf("Confidence = %v, want 0.95", retrieved.Confidence)
	}
}

func TestGetRecentBranches(t *testing.T) {
	store := NewMemoryStorage()

	// Create and access branches in order
	branchIDs := []string{"b1", "b2", "b3", "b4", "b5"}
	for _, id := range branchIDs {
		branch := &types.Branch{
			ID:        id,
			State:     types.StateActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = store.StoreBranch(branch)
		store.UpdateBranchAccess(id)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get recent branches
	recent, err := store.GetRecentBranches()
	if err != nil {
		t.Fatalf("GetRecentBranches() error = %v", err)
	}

	// Should return all 5 branches since we created 5 and the limit is 10
	if len(recent) != 5 {
		t.Fatalf("Expected 5 recent branches, got %d", len(recent))
	}

	// Verify most recent first (b5, b4, b3, b2, b1)
	expected := []string{"b5", "b4", "b3", "b2", "b1"}
	for i, branch := range recent {
		if branch.ID != expected[i] {
			t.Errorf("Recent[%d] = %v, want %v", i, branch.ID, expected[i])
		}
	}
}

func TestGetMetrics(t *testing.T) {
	store := NewMemoryStorage()

	// Add some data
	for i := 0; i < 5; i++ {
		thought := &types.Thought{
			Content:    "Test",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		_ = store.StoreThought(thought)
	}

	for i := 0; i < 3; i++ {
		branch := &types.Branch{
			State:     types.StateActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = store.StoreBranch(branch)
	}

	metrics := store.GetMetrics()

	if metrics.TotalThoughts != 5 {
		t.Errorf("TotalThoughts = %d, want 5", metrics.TotalThoughts)
	}

	if metrics.TotalBranches != 3 {
		t.Errorf("TotalBranches = %d, want 3", metrics.TotalBranches)
	}

	if metrics.TotalBranches != 3 {
		t.Errorf("TotalBranches = %d, want 3", metrics.TotalBranches)
	}
}

func TestAppendOperations_DataIsolation(t *testing.T) {
	// This test verifies that appended data is copied and not referenced
	store := NewMemoryStorage()

	branch := &types.Branch{
		ID:        "isolation-branch",
		State:     types.StateActive,
		Thoughts:  make([]*types.Thought, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = store.StoreBranch(branch)

	thought := &types.Thought{
		ID:         "mutable-thought",
		Content:    "Original content",
		Mode:       types.ModeLinear,
		BranchID:   "isolation-branch",
		Confidence: 0.8,
		Timestamp:  time.Now(),
		Metadata:   map[string]interface{}{"key": "original"},
	}

	// Append thought
	store.AppendThoughtToBranch("isolation-branch", thought)

	// Modify original thought's metadata
	thought.Metadata["key"] = "modified"
	thought.Content = "Modified content"

	// Retrieve and verify isolation
	retrieved, _ := store.GetBranch("isolation-branch")
	if retrieved.Thoughts[0].Content != "Original content" {
		t.Error("Thought content should not be affected by external modification")
	}

	if retrieved.Thoughts[0].Metadata["key"] != "original" {
		t.Error("Thought metadata should not be affected by external modification")
	}
}

func TestConcurrentUpdateOperations(t *testing.T) {
	store := NewMemoryStorage()

	branch := &types.Branch{
		ID:        "concurrent-update-branch",
		State:     types.StateActive,
		Priority:  1.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = store.StoreBranch(branch)

	var wg sync.WaitGroup
	updates := 100

	// Concurrent priority updates
	wg.Add(updates)
	for i := 0; i < updates; i++ {
		go func(val float64) {
			defer wg.Done()
			store.UpdateBranchPriority("concurrent-update-branch", val)
		}(float64(i))
	}

	wg.Wait()

	// Just verify no crashes occurred and final value is one of the updates
	retrieved, _ := store.GetBranch("concurrent-update-branch")
	if retrieved.Priority < 0 || retrieved.Priority >= float64(updates) {
		t.Errorf("Priority should be between 0 and %d, got %v", updates, retrieved.Priority)
	}
}

func TestUpdateBranchAccess_LRU(t *testing.T) {
	store := NewMemoryStorage()

	// Create more branches than the LRU limit
	for i := 0; i < 15; i++ {
		branch := &types.Branch{
			ID:        fmt.Sprintf("branch-%d", i),
			State:     types.StateActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = store.StoreBranch(branch)
		store.UpdateBranchAccess(fmt.Sprintf("branch-%d", i))
		time.Sleep(5 * time.Millisecond)
	}

	// Get recent branches (default limit is 10)
	recent, _ := store.GetRecentBranches()

	// Should respect MaxRecentBranches limit
	if len(recent) > MaxRecentBranches {
		t.Errorf("Recent branches should be capped at %d, got %d", MaxRecentBranches, len(recent))
	}

	// Most recent should be branch-14
	if len(recent) > 0 && recent[0].ID != "branch-14" {
		t.Errorf("Most recent branch should be branch-14, got %s", recent[0].ID)
	}
}

func TestAppendToBranchNotFound(t *testing.T) {
	store := NewMemoryStorage()

	thought := &types.Thought{
		ID:         "orphan-thought",
		Content:    "Test",
		Mode:       types.ModeLinear,
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}

	err := store.AppendThoughtToBranch("non-existent-branch", thought)
	if err == nil {
		t.Error("AppendThoughtToBranch should return error for non-existent branch")
	}

	insight := &types.Insight{
		ID:                 "orphan-insight",
		Content:            "Test",
		Type:               types.InsightObservation,
		ApplicabilityScore: 0.9,
		CreatedAt:          time.Now(),
		Context:            []string{},
		SupportingEvidence: make(map[string]interface{}),
	}

	err = store.AppendInsightToBranch("non-existent-branch", insight)
	if err == nil {
		t.Error("AppendInsightToBranch should return error for non-existent branch")
	}
}
