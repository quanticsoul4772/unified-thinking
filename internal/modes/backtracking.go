// Package modes provides backtracking support for tree mode.
// Implements hybrid snapshot + delta architecture for efficient branch history.
package modes

import (
	"context"
	"fmt"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// BacktrackingManager manages branch checkpoints and history
type BacktrackingManager struct {
	storage           storage.Storage
	snapshots         map[string]*BranchSnapshot  // branchID -> latest snapshot
	deltas            map[string][]*BranchDelta   // branchID -> deltas since last snapshot
	checkpoints       map[string]*Checkpoint      // checkpointID -> checkpoint
	checkpointCounter int                         // Counter for unique checkpoint IDs
}

// NewBacktrackingManager creates a new backtracking manager
func NewBacktrackingManager(store storage.Storage) *BacktrackingManager {
	return &BacktrackingManager{
		storage:     store,
		snapshots:   make(map[string]*BranchSnapshot),
		deltas:      make(map[string][]*BranchDelta),
		checkpoints: make(map[string]*Checkpoint),
	}
}

// BranchSnapshot represents a full branch state at a point in time
type BranchSnapshot struct {
	ID             string
	BranchID       string
	Branch         *types.Branch // Deep copy of branch
	CreatedAt      time.Time
	ThoughtCount   int
	InsightCount   int
	CrossRefCount  int
}

// BranchDelta represents a change to a branch
type BranchDelta struct {
	ID         string
	BranchID   string
	Operation  DeltaOperation
	EntityType EntityType
	EntityID   string
	Entity     interface{} // Thought, Insight, or CrossRef
	CreatedAt  time.Time
}

// DeltaOperation represents type of change
type DeltaOperation string

const (
	DeltaAdd    DeltaOperation = "add"
	DeltaRemove DeltaOperation = "remove"
	DeltaModify DeltaOperation = "modify"
)

// EntityType represents what kind of entity changed
type EntityType string

const (
	EntityThought  EntityType = "thought"
	EntityInsight  EntityType = "insight"
	EntityCrossRef EntityType = "cross_ref"
)

// Checkpoint represents a named savepoint
type Checkpoint struct {
	ID          string
	Name        string
	Description string
	BranchID    string
	SnapshotID  string
	DeltaCount  int // Number of deltas to apply from snapshot
	CreatedAt   time.Time
	Metadata    map[string]interface{}
}

// CreateCheckpoint creates a checkpoint for the current branch state
func (bm *BacktrackingManager) CreateCheckpoint(ctx context.Context, branchID, name, description string) (*Checkpoint, error) {
	// Get current branch
	branch, err := bm.storage.GetBranch(branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}

	// Check if we need a new snapshot (every 10 deltas)
	deltaCount := len(bm.deltas[branchID])
	if deltaCount >= 10 || bm.snapshots[branchID] == nil {
		if err := bm.createSnapshot(branchID, branch); err != nil {
			return nil, fmt.Errorf("failed to create snapshot: %w", err)
		}
		// Deltas are cleared after snapshot, so checkpoint delta count is 0
		deltaCount = 0
	}

	// Create checkpoint - store full branch state as metadata for diffing
	bm.checkpointCounter++
	checkpoint := &Checkpoint{
		ID:          fmt.Sprintf("checkpoint-%d-%d-%s", bm.checkpointCounter, time.Now().UnixNano(), branchID),
		Name:        name,
		Description: description,
		BranchID:    branchID,
		SnapshotID:  bm.snapshots[branchID].ID,
		DeltaCount:  deltaCount,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Store thought/insight IDs for diffing
	thoughtIDs := make([]string, len(branch.Thoughts))
	for i, t := range branch.Thoughts {
		thoughtIDs[i] = t.ID
	}
	insightIDs := make([]string, len(branch.Insights))
	for i, ins := range branch.Insights {
		insightIDs[i] = ins.ID
	}
	checkpoint.Metadata["thought_ids"] = thoughtIDs
	checkpoint.Metadata["insight_ids"] = insightIDs

	bm.checkpoints[checkpoint.ID] = checkpoint

	return checkpoint, nil
}

// RestoreCheckpoint restores a branch to a checkpoint state
func (bm *BacktrackingManager) RestoreCheckpoint(ctx context.Context, checkpointID string) (*types.Branch, error) {
	checkpoint, exists := bm.checkpoints[checkpointID]
	if !exists {
		return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
	}

	// Get snapshot
	snapshot := bm.snapshots[checkpoint.BranchID]
	if snapshot == nil || snapshot.ID != checkpoint.SnapshotID {
		return nil, fmt.Errorf("snapshot not found for checkpoint: %s", checkpointID)
	}

	// Deep copy the snapshot branch
	restoredBranch := bm.deepCopyBranch(snapshot.Branch)

	// Apply deltas up to checkpoint
	deltas := bm.deltas[checkpoint.BranchID]
	if checkpoint.DeltaCount > len(deltas) {
		return nil, fmt.Errorf("delta count mismatch: checkpoint expects %d, have %d", checkpoint.DeltaCount, len(deltas))
	}

	for i := 0; i < checkpoint.DeltaCount; i++ {
		if err := bm.applyDelta(restoredBranch, deltas[i]); err != nil {
			return nil, fmt.Errorf("failed to apply delta %d: %w", i, err)
		}
	}

	// Store restored branch
	restoredBranch.UpdatedAt = time.Now()
	if err := bm.storage.StoreBranch(restoredBranch); err != nil {
		return nil, fmt.Errorf("failed to store restored branch: %w", err)
	}

	return restoredBranch, nil
}

// RecordChange records a change to track in deltas
func (bm *BacktrackingManager) RecordChange(branchID string, operation DeltaOperation, entityType EntityType, entityID string, entity interface{}) error {
	delta := &BranchDelta{
		ID:         fmt.Sprintf("delta-%d-%s", time.Now().UnixNano(), branchID),
		BranchID:   branchID,
		Operation:  operation,
		EntityType: entityType,
		EntityID:   entityID,
		Entity:     entity,
		CreatedAt:  time.Now(),
	}

	bm.deltas[branchID] = append(bm.deltas[branchID], delta)

	return nil
}

// ForkFromCheckpoint creates a new branch from a checkpoint
func (bm *BacktrackingManager) ForkFromCheckpoint(ctx context.Context, checkpointID, newBranchName string) (*types.Branch, error) {
	// Restore to checkpoint state
	restoredBranch, err := bm.RestoreCheckpoint(ctx, checkpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to restore checkpoint: %w", err)
	}

	// Create new branch from restored state
	newBranch := bm.deepCopyBranch(restoredBranch)
	newBranch.ID = fmt.Sprintf("branch-%d", time.Now().Unix())
	newBranch.ParentBranchID = restoredBranch.ID
	newBranch.CreatedAt = time.Now()
	newBranch.UpdatedAt = time.Now()

	// Add metadata about fork origin
	if len(newBranch.Thoughts) > 0 {
		lastThought := newBranch.Thoughts[len(newBranch.Thoughts)-1]
		if lastThought.Metadata == nil {
			lastThought.Metadata = make(map[string]interface{})
		}
		lastThought.Metadata["forked_from_checkpoint"] = checkpointID
		lastThought.Metadata["fork_time"] = time.Now()
	}

	// Store new branch
	if err := bm.storage.StoreBranch(newBranch); err != nil {
		return nil, fmt.Errorf("failed to store forked branch: %w", err)
	}

	return newBranch, nil
}

// ListCheckpoints lists all checkpoints for a branch
func (bm *BacktrackingManager) ListCheckpoints(branchID string) []*Checkpoint {
	checkpoints := make([]*Checkpoint, 0)
	for _, cp := range bm.checkpoints {
		if cp.BranchID == branchID {
			checkpoints = append(checkpoints, cp)
		}
	}
	return checkpoints
}

// GetCheckpointDiff compares two checkpoints and returns differences
func (bm *BacktrackingManager) GetCheckpointDiff(checkpoint1ID, checkpoint2ID string) (*CheckpointDiff, error) {
	cp1, exists1 := bm.checkpoints[checkpoint1ID]
	cp2, exists2 := bm.checkpoints[checkpoint2ID]

	if !exists1 || !exists2 {
		return nil, fmt.Errorf("one or both checkpoints not found")
	}

	if cp1.BranchID != cp2.BranchID {
		return nil, fmt.Errorf("checkpoints are from different branches")
	}

	diff := &CheckpointDiff{
		Checkpoint1:     cp1,
		Checkpoint2:     cp2,
		ThoughtsAdded:   make([]string, 0),
		ThoughtsRemoved: make([]string, 0),
		InsightsAdded:   make([]string, 0),
		InsightsRemoved: make([]string, 0),
	}

	// Get IDs from checkpoint metadata
	cp1Thoughts := cp1.Metadata["thought_ids"].([]string)
	cp2Thoughts := cp2.Metadata["thought_ids"].([]string)
	cp1Insights := cp1.Metadata["insight_ids"].([]string)
	cp2Insights := cp2.Metadata["insight_ids"].([]string)

	// Find thoughts added (in cp2 but not in cp1)
	cp1ThoughtMap := make(map[string]bool)
	for _, id := range cp1Thoughts {
		cp1ThoughtMap[id] = true
	}
	for _, id := range cp2Thoughts {
		if !cp1ThoughtMap[id] {
			diff.ThoughtsAdded = append(diff.ThoughtsAdded, id)
		}
	}

	// Find thoughts removed (in cp1 but not in cp2)
	cp2ThoughtMap := make(map[string]bool)
	for _, id := range cp2Thoughts {
		cp2ThoughtMap[id] = true
	}
	for _, id := range cp1Thoughts {
		if !cp2ThoughtMap[id] {
			diff.ThoughtsRemoved = append(diff.ThoughtsRemoved, id)
		}
	}

	// Find insights added
	cp1InsightMap := make(map[string]bool)
	for _, id := range cp1Insights {
		cp1InsightMap[id] = true
	}
	for _, id := range cp2Insights {
		if !cp1InsightMap[id] {
			diff.InsightsAdded = append(diff.InsightsAdded, id)
		}
	}

	// Find insights removed
	cp2InsightMap := make(map[string]bool)
	for _, id := range cp2Insights {
		cp2InsightMap[id] = true
	}
	for _, id := range cp1Insights {
		if !cp2InsightMap[id] {
			diff.InsightsRemoved = append(diff.InsightsRemoved, id)
		}
	}

	return diff, nil
}

// CheckpointDiff represents differences between two checkpoints
type CheckpointDiff struct {
	Checkpoint1     *Checkpoint
	Checkpoint2     *Checkpoint
	Reversed        bool
	ThoughtsAdded   []string
	ThoughtsRemoved []string
	InsightsAdded   []string
	InsightsRemoved []string
}

// PruneBranch marks a branch as abandoned (failed exploration)
func (bm *BacktrackingManager) PruneBranch(branchID string, reason string) error {
	branch, err := bm.storage.GetBranch(branchID)
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	branch.State = types.StateDeadEnd
	branch.UpdatedAt = time.Now()

	// Add prune metadata to last thought
	if len(branch.Thoughts) > 0 {
		lastThought := branch.Thoughts[len(branch.Thoughts)-1]
		if lastThought.Metadata == nil {
			lastThought.Metadata = make(map[string]interface{})
		}
		lastThought.Metadata["pruned"] = true
		lastThought.Metadata["prune_reason"] = reason
		lastThought.Metadata["pruned_at"] = time.Now()
	}

	return bm.storage.StoreBranch(branch)
}

// Helper methods

func (bm *BacktrackingManager) createSnapshot(branchID string, branch *types.Branch) error {
	snapshot := &BranchSnapshot{
		ID:            fmt.Sprintf("snapshot-%d-%s", time.Now().Unix(), branchID),
		BranchID:      branchID,
		Branch:        bm.deepCopyBranch(branch),
		CreatedAt:     time.Now(),
		ThoughtCount:  len(branch.Thoughts),
		InsightCount:  len(branch.Insights),
		CrossRefCount: len(branch.CrossRefs),
	}

	bm.snapshots[branchID] = snapshot

	// Clear old deltas since we have a new snapshot
	bm.deltas[branchID] = make([]*BranchDelta, 0)

	return nil
}

func (bm *BacktrackingManager) deepCopyBranch(branch *types.Branch) *types.Branch {
	// Deep copy branch
	copied := &types.Branch{
		ID:             branch.ID,
		ParentBranchID: branch.ParentBranchID,
		State:          branch.State,
		Priority:       branch.Priority,
		Confidence:     branch.Confidence,
		CreatedAt:      branch.CreatedAt,
		UpdatedAt:      branch.UpdatedAt,
		LastAccessedAt: branch.LastAccessedAt,
	}

	// Deep copy thoughts
	copied.Thoughts = make([]*types.Thought, len(branch.Thoughts))
	for i, thought := range branch.Thoughts {
		copied.Thoughts[i] = bm.deepCopyThought(thought)
	}

	// Deep copy insights
	copied.Insights = make([]*types.Insight, len(branch.Insights))
	for i, insight := range branch.Insights {
		copied.Insights[i] = bm.deepCopyInsight(insight)
	}

	// Deep copy cross-refs
	copied.CrossRefs = make([]*types.CrossRef, len(branch.CrossRefs))
	for i, ref := range branch.CrossRefs {
		copied.CrossRefs[i] = bm.deepCopyCrossRef(ref)
	}

	return copied
}

func (bm *BacktrackingManager) deepCopyThought(thought *types.Thought) *types.Thought {
	copied := &types.Thought{
		ID:         thought.ID,
		Content:    thought.Content,
		Mode:       thought.Mode,
		BranchID:   thought.BranchID,
		ParentID:   thought.ParentID,
		Type:       thought.Type,
		Confidence: thought.Confidence,
		Timestamp:  thought.Timestamp,
	}

	// Copy key points
	if thought.KeyPoints != nil {
		copied.KeyPoints = make([]string, len(thought.KeyPoints))
		copy(copied.KeyPoints, thought.KeyPoints)
	}

	// Copy metadata
	if thought.Metadata != nil {
		copied.Metadata = make(map[string]interface{})
		for k, v := range thought.Metadata {
			copied.Metadata[k] = v
		}
	}

	return copied
}

func (bm *BacktrackingManager) deepCopyInsight(insight *types.Insight) *types.Insight {
	copied := &types.Insight{
		ID:                 insight.ID,
		Type:               insight.Type,
		Content:            insight.Content,
		ApplicabilityScore: insight.ApplicabilityScore,
		CreatedAt:          insight.CreatedAt,
	}

	// Copy context
	if insight.Context != nil {
		copied.Context = make([]string, len(insight.Context))
		copy(copied.Context, insight.Context)
	}

	// Copy parent insights
	if insight.ParentInsights != nil {
		copied.ParentInsights = make([]string, len(insight.ParentInsights))
		copy(copied.ParentInsights, insight.ParentInsights)
	}

	// Copy supporting evidence
	if insight.SupportingEvidence != nil {
		copied.SupportingEvidence = make(map[string]interface{})
		for k, v := range insight.SupportingEvidence {
			copied.SupportingEvidence[k] = v
		}
	}

	// Copy validations (shallow copy of slice)
	if insight.Validations != nil {
		copied.Validations = make([]*types.Validation, len(insight.Validations))
		copy(copied.Validations, insight.Validations)
	}

	return copied
}

func (bm *BacktrackingManager) deepCopyCrossRef(ref *types.CrossRef) *types.CrossRef {
	copied := &types.CrossRef{
		ID:         ref.ID,
		FromBranch: ref.FromBranch,
		ToBranch:   ref.ToBranch,
		Type:       ref.Type,
		Reason:     ref.Reason,
		Strength:   ref.Strength,
		CreatedAt:  ref.CreatedAt,
	}

	// Copy touch points
	if ref.TouchPoints != nil {
		copied.TouchPoints = make([]types.TouchPoint, len(ref.TouchPoints))
		for i, tp := range ref.TouchPoints {
			copied.TouchPoints[i] = types.TouchPoint{
				FromThought: tp.FromThought,
				ToThought:   tp.ToThought,
				Connection:  tp.Connection,
			}
		}
	}

	return copied
}

func (bm *BacktrackingManager) applyDelta(branch *types.Branch, delta *BranchDelta) error {
	switch delta.EntityType {
	case EntityThought:
		return bm.applyThoughtDelta(branch, delta)
	case EntityInsight:
		return bm.applyInsightDelta(branch, delta)
	case EntityCrossRef:
		return bm.applyCrossRefDelta(branch, delta)
	default:
		return fmt.Errorf("unknown entity type: %s", delta.EntityType)
	}
}

func (bm *BacktrackingManager) applyThoughtDelta(branch *types.Branch, delta *BranchDelta) error {
	switch delta.Operation {
	case DeltaAdd:
		thought := delta.Entity.(*types.Thought)
		branch.Thoughts = append(branch.Thoughts, bm.deepCopyThought(thought))
	case DeltaRemove:
		for i, t := range branch.Thoughts {
			if t.ID == delta.EntityID {
				branch.Thoughts = append(branch.Thoughts[:i], branch.Thoughts[i+1:]...)
				break
			}
		}
	case DeltaModify:
		thought := delta.Entity.(*types.Thought)
		for i, t := range branch.Thoughts {
			if t.ID == delta.EntityID {
				branch.Thoughts[i] = bm.deepCopyThought(thought)
				break
			}
		}
	}
	return nil
}

func (bm *BacktrackingManager) applyInsightDelta(branch *types.Branch, delta *BranchDelta) error {
	switch delta.Operation {
	case DeltaAdd:
		insight := delta.Entity.(*types.Insight)
		branch.Insights = append(branch.Insights, bm.deepCopyInsight(insight))
	case DeltaRemove:
		for i, ins := range branch.Insights {
			if ins.ID == delta.EntityID {
				branch.Insights = append(branch.Insights[:i], branch.Insights[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (bm *BacktrackingManager) applyCrossRefDelta(branch *types.Branch, delta *BranchDelta) error {
	switch delta.Operation {
	case DeltaAdd:
		ref := delta.Entity.(*types.CrossRef)
		branch.CrossRefs = append(branch.CrossRefs, bm.deepCopyCrossRef(ref))
	case DeltaRemove:
		for i, r := range branch.CrossRefs {
			if r.ID == delta.EntityID {
				branch.CrossRefs = append(branch.CrossRefs[:i], branch.CrossRefs[i+1:]...)
				break
			}
		}
	}
	return nil
}
