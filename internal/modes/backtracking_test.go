package modes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestNewBacktrackingManager(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	assert.NotNil(t, bm)
	assert.NotNil(t, bm.storage)
	assert.NotNil(t, bm.snapshots)
	assert.NotNil(t, bm.deltas)
	assert.NotNil(t, bm.checkpoints)
}

func TestBacktrackingManager_CreateCheckpoint(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create a branch
	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		Thoughts:   []*types.Thought{},
		Insights:   []*types.Insight{},
		CrossRefs:  []*types.CrossRef{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := store.StoreBranch(branch)
	assert.NoError(t, err)

	ctx := context.Background()
	checkpoint, err := bm.CreateCheckpoint(ctx, "branch-1", "checkpoint1", "First checkpoint")

	assert.NoError(t, err)
	assert.NotNil(t, checkpoint)
	assert.Equal(t, "checkpoint1", checkpoint.Name)
	assert.Equal(t, "First checkpoint", checkpoint.Description)
	assert.Equal(t, "branch-1", checkpoint.BranchID)
	assert.NotEmpty(t, checkpoint.ID)

	// Verify snapshot was created
	assert.NotNil(t, bm.snapshots["branch-1"])
}

func TestBacktrackingManager_RecordChange(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	thought := &types.Thought{
		ID:      "thought-1",
		Content: "Test thought",
	}

	err := bm.RecordChange("branch-1", DeltaAdd, EntityThought, "thought-1", thought)

	assert.NoError(t, err)
	assert.Len(t, bm.deltas["branch-1"], 1)
	assert.Equal(t, DeltaAdd, bm.deltas["branch-1"][0].Operation)
	assert.Equal(t, EntityThought, bm.deltas["branch-1"][0].EntityType)
	assert.Equal(t, "thought-1", bm.deltas["branch-1"][0].EntityID)
}

func TestBacktrackingManager_RestoreCheckpoint(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create initial branch
	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		Thoughts: []*types.Thought{
			{ID: "thought-1", Content: "Initial thought"},
		},
		Insights:  []*types.Insight{},
		CrossRefs: []*types.CrossRef{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := store.StoreBranch(branch)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create checkpoint
	checkpoint, err := bm.CreateCheckpoint(ctx, "branch-1", "cp1", "Test checkpoint")
	assert.NoError(t, err)

	// Add more thoughts after checkpoint
	thought2 := &types.Thought{ID: "thought-2", Content: "Second thought"}
	err = bm.RecordChange("branch-1", DeltaAdd, EntityThought, "thought-2", thought2)
	assert.NoError(t, err)

	// Modify branch in storage
	branch.Thoughts = append(branch.Thoughts, thought2)
	err = store.StoreBranch(branch)
	assert.NoError(t, err)

	// Restore to checkpoint
	restored, err := bm.RestoreCheckpoint(ctx, checkpoint.ID)

	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Len(t, restored.Thoughts, 1) // Should only have initial thought
	assert.Equal(t, "thought-1", restored.Thoughts[0].ID)
}

func TestBacktrackingManager_ForkFromCheckpoint(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create initial branch
	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		Thoughts: []*types.Thought{
			{ID: "thought-1", Content: "Initial thought"},
		},
		Insights:  []*types.Insight{},
		CrossRefs: []*types.CrossRef{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := store.StoreBranch(branch)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create checkpoint
	checkpoint, err := bm.CreateCheckpoint(ctx, "branch-1", "cp1", "Fork point")
	assert.NoError(t, err)

	// Fork from checkpoint
	forked, err := bm.ForkFromCheckpoint(ctx, checkpoint.ID, "forked-branch")

	assert.NoError(t, err)
	assert.NotNil(t, forked)
	assert.NotEqual(t, "branch-1", forked.ID) // Should have different ID
	assert.Equal(t, "branch-1", forked.ParentBranchID)
	assert.Len(t, forked.Thoughts, 1)
}

func TestBacktrackingManager_ListCheckpoints(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create branch
	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		Thoughts:   []*types.Thought{},
		Insights:   []*types.Insight{},
		CrossRefs:  []*types.CrossRef{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := store.StoreBranch(branch)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create multiple checkpoints
	cp1, _ := bm.CreateCheckpoint(ctx, "branch-1", "cp1", "First")
	cp2, _ := bm.CreateCheckpoint(ctx, "branch-1", "cp2", "Second")
	cp3, _ := bm.CreateCheckpoint(ctx, "branch-1", "cp3", "Third")

	// List checkpoints
	checkpoints := bm.ListCheckpoints("branch-1")

	assert.Len(t, checkpoints, 3)
	ids := make(map[string]bool)
	for _, cp := range checkpoints {
		ids[cp.ID] = true
	}
	assert.True(t, ids[cp1.ID])
	assert.True(t, ids[cp2.ID])
	assert.True(t, ids[cp3.ID])
}

func TestBacktrackingManager_GetCheckpointDiff(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create branch
	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		Thoughts:   []*types.Thought{},
		Insights:   []*types.Insight{},
		CrossRefs:  []*types.CrossRef{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := store.StoreBranch(branch)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create first checkpoint
	cp1, _ := bm.CreateCheckpoint(ctx, "branch-1", "cp1", "Before changes")

	// Add thoughts to branch
	thought1 := &types.Thought{ID: "thought-1", Content: "First"}
	thought2 := &types.Thought{ID: "thought-2", Content: "Second"}
	branch.Thoughts = append(branch.Thoughts, thought1, thought2)
	_ = store.StoreBranch(branch)

	// Record the changes
	_ = 	bm.RecordChange("branch-1", DeltaAdd, EntityThought, "thought-1", thought1)
	_ = 	bm.RecordChange("branch-1", DeltaAdd, EntityThought, "thought-2", thought2)

	// Create second checkpoint
	cp2, _ := bm.CreateCheckpoint(ctx, "branch-1", "cp2", "After changes")

	// Get diff
	diff, err := bm.GetCheckpointDiff(cp1.ID, cp2.ID)

	assert.NoError(t, err)
	assert.NotNil(t, diff)
	assert.Len(t, diff.ThoughtsAdded, 2)
	assert.Contains(t, diff.ThoughtsAdded, "thought-1")
	assert.Contains(t, diff.ThoughtsAdded, "thought-2")
}

func TestBacktrackingManager_PruneBranch(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create branch
	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		Thoughts: []*types.Thought{
			{ID: "thought-1", Content: "Test"},
		},
		Insights:  []*types.Insight{},
		CrossRefs: []*types.CrossRef{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := store.StoreBranch(branch)
	assert.NoError(t, err)

	// Prune branch
	err = bm.PruneBranch("branch-1", "Failed exploration")

	assert.NoError(t, err)

	// Verify branch was marked as dead end
	pruned, err := store.GetBranch("branch-1")
	assert.NoError(t, err)
	assert.Equal(t, types.StateDeadEnd, pruned.State)
	assert.True(t, pruned.Thoughts[0].Metadata["pruned"].(bool))
	assert.Equal(t, "Failed exploration", pruned.Thoughts[0].Metadata["prune_reason"])
}

func TestBacktrackingManager_SnapshotCreation(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create branch
	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   1.0,
		Confidence: 0.8,
		Thoughts:   []*types.Thought{},
		Insights:   []*types.Insight{},
		CrossRefs:  []*types.CrossRef{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := store.StoreBranch(branch)
	assert.NoError(t, err)

	ctx := context.Background()

	// Add 11 changes to trigger snapshot (threshold is 10)
	for i := 0; i < 11; i++ {
		thought := &types.Thought{
			ID:      fmt.Sprintf("thought-%d", i),
			Content: fmt.Sprintf("Thought %d", i),
		}
		_ = bm.RecordChange("branch-1", DeltaAdd, EntityThought, thought.ID, thought)
	}

	// Create checkpoint (should trigger snapshot creation)
	checkpoint, err := bm.CreateCheckpoint(ctx, "branch-1", "cp1", "After many changes")

	assert.NoError(t, err)
	assert.NotNil(t, checkpoint)

	// Verify snapshot was created
	assert.NotNil(t, bm.snapshots["branch-1"])

	// Verify deltas were cleared after snapshot
	assert.Len(t, bm.deltas["branch-1"], 0)
}

func TestBacktrackingManager_DeepCopy(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create thought with metadata
	original := &types.Thought{
		ID:         "thought-1",
		Content:    "Test thought",
		KeyPoints:  []string{"point1", "point2"},
		Confidence: 0.8,
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	// Deep copy
	copied := bm.deepCopyThought(original)

	// Verify copy is independent
	assert.Equal(t, original.ID, copied.ID)
	assert.Equal(t, original.Content, copied.Content)
	assert.Equal(t, len(original.KeyPoints), len(copied.KeyPoints))

	// Modify original
	original.Content = "Modified"
	original.KeyPoints[0] = "modified"
	original.Metadata["key1"] = "modified"

	// Verify copy wasn't affected
	assert.Equal(t, "Test thought", copied.Content)
	assert.Equal(t, "point1", copied.KeyPoints[0])
	assert.Equal(t, "value1", copied.Metadata["key1"])
}

func TestBacktrackingManager_ApplyDelta(t *testing.T) {
	store := storage.NewMemoryStorage()
	bm := NewBacktrackingManager(store)

	// Create branch
	branch := &types.Branch{
		ID:        "branch-1",
		Thoughts:  []*types.Thought{},
		Insights:  []*types.Insight{},
		CrossRefs: []*types.CrossRef{},
	}

	// Create delta to add thought
	thought := &types.Thought{
		ID:      "thought-1",
		Content: "Test",
	}
	delta := &BranchDelta{
		Operation:  DeltaAdd,
		EntityType: EntityThought,
		EntityID:   "thought-1",
		Entity:     thought,
	}

	// Apply delta
	err := bm.applyDelta(branch, delta)

	assert.NoError(t, err)
	assert.Len(t, branch.Thoughts, 1)
	assert.Equal(t, "thought-1", branch.Thoughts[0].ID)

	// Create delta to remove thought
	removeDelta := &BranchDelta{
		Operation:  DeltaRemove,
		EntityType: EntityThought,
		EntityID:   "thought-1",
	}

	// Apply remove delta
	err = bm.applyDelta(branch, removeDelta)

	assert.NoError(t, err)
	assert.Len(t, branch.Thoughts, 0)
}
