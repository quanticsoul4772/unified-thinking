package modes

import (
	"context"
	"fmt"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// TreeMode implements multi-branch exploration
type TreeMode struct {
	storage storage.Storage
}

// NewTreeMode creates a new tree mode handler
func NewTreeMode(storage storage.Storage) *TreeMode {
	return &TreeMode{storage: storage}
}

// ProcessThought processes a thought in tree mode with branching
func (m *TreeMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
	// Determine branch (use provided or active)
	branchID := input.BranchID
	if branchID == "" {
		if activeBranch, err := m.storage.GetActiveBranch(); err == nil && activeBranch != nil {
			branchID = activeBranch.ID
		} else {
			// Create new branch
			branch := &types.Branch{
				State:      types.StateActive,
				Priority:   1.0,
				Confidence: input.Confidence,
				Thoughts:   make([]*types.Thought, 0),
				Insights:   make([]*types.Insight, 0),
				CrossRefs:  make([]*types.CrossRef, 0),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err := m.storage.StoreBranch(branch); err != nil {
				return nil, err
			}
			branchID = branch.ID
		}
	} else {
		// BranchID was provided - check if it exists, create if it doesn't
		_, err := m.storage.GetBranch(branchID)
		if err != nil {
			// Branch doesn't exist, create it
			branch := &types.Branch{
				ID:         branchID,
				State:      types.StateActive,
				Priority:   1.0,
				Confidence: input.Confidence,
				Thoughts:   make([]*types.Thought, 0),
				Insights:   make([]*types.Insight, 0),
				CrossRefs:  make([]*types.CrossRef, 0),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err := m.storage.StoreBranch(branch); err != nil {
				return nil, err
			}
		}
	}

	// Update branch access tracking (branch is guaranteed to exist now)
	if err := m.storage.UpdateBranchAccess(branchID); err != nil {
		return nil, err
	}

	// Create thought
	thought := &types.Thought{
		Content:    input.Content,
		Mode:       types.ModeTree,
		BranchID:   branchID,
		ParentID:   input.ParentID,
		Type:       input.Type,
		Confidence: input.Confidence,
		KeyPoints:  input.KeyPoints,
		Timestamp:  time.Now(),
	}

	if err := m.storage.StoreThought(thought); err != nil {
		return nil, err
	}

	// Append thought directly to branch (optimized - avoids deep copy)
	if err := m.storage.AppendThoughtToBranch(branchID, thought); err != nil {
		return nil, err
	}

	// Generate insights from key points
	if len(input.KeyPoints) > 0 {
		insight := &types.Insight{
			Type:               types.InsightObservation,
			Content:            fmt.Sprintf("Key points identified: %v", input.KeyPoints),
			Context:            []string{input.Type},
			ApplicabilityScore: input.Confidence,
			SupportingEvidence: map[string]interface{}{},
			CreatedAt:          time.Now(),
		}
		if err := m.storage.StoreInsight(insight); err != nil {
			return nil, err
		}
		// Append insight directly to branch (optimized - avoids deep copy)
		if err := m.storage.AppendInsightToBranch(branchID, insight); err != nil {
			return nil, err
		}
	}

	// Handle cross-references (still requires Get-Modify-Store for metrics calculation)
	if len(input.CrossRefs) > 0 {
		for _, xref := range input.CrossRefs {
			crossRef := &types.CrossRef{
				FromBranch: branchID,
				ToBranch:   xref.ToBranch,
				Type:       types.CrossRefType(xref.Type),
				Reason:     xref.Reason,
				Strength:   xref.Strength,
				CreatedAt:  time.Now(),
			}
			// Append cross-ref directly to branch (optimized - avoids deep copy)
			if err := m.storage.AppendCrossRefToBranch(branchID, crossRef); err != nil {
				return nil, err
			}
		}
	}

	// Get branch for result metrics
	branch, err := m.storage.GetBranch(branchID)
	if err != nil {
		return nil, err
	}

	// Update branch metrics if needed
	if len(input.CrossRefs) > 0 {
		m.updateBranchMetrics(branch)
		// Store updated priority and confidence
		if err := m.storage.UpdateBranchPriority(branchID, branch.Priority); err != nil {
			return nil, err
		}
		if err := m.storage.UpdateBranchConfidence(branchID, branch.Confidence); err != nil {
			return nil, err
		}
	}

	result := &ThoughtResult{
		ThoughtID:     thought.ID,
		BranchID:      branchID,
		Mode:          string(types.ModeTree),
		BranchState:   string(branch.State),
		Priority:      branch.Priority,
		InsightCount:  len(branch.Insights),
		CrossRefCount: len(branch.CrossRefs),
	}

	return result, nil
}

func (m *TreeMode) updateBranchMetrics(branch *types.Branch) {
	// Calculate average confidence
	if len(branch.Thoughts) > 0 {
		totalConf := 0.0
		for _, t := range branch.Thoughts {
			totalConf += t.Confidence
		}
		branch.Confidence = totalConf / float64(len(branch.Thoughts))
	}

	// Calculate priority
	// TIER 2 OPTIMIZATION: Priority calculation is simple and fast (no caching needed)
	// The cost of cache management would exceed calculation cost
	insightScore := float64(len(branch.Insights)) * 0.1
	crossRefScore := 0.0
	for _, xref := range branch.CrossRefs {
		crossRefScore += xref.Strength * 0.1
	}
	branch.Priority = branch.Confidence + insightScore + crossRefScore
}

// ListBranches returns all branches
func (m *TreeMode) ListBranches(ctx context.Context) ([]*types.Branch, error) {
	branches := m.storage.ListBranches()
	return branches, nil
}

// GetBranchHistory returns detailed history of a branch
func (m *TreeMode) GetBranchHistory(ctx context.Context, branchID string) (*BranchHistory, error) {
	branch, err := m.storage.GetBranch(branchID)
	if err != nil {
		return nil, err
	}

	history := &BranchHistory{
		BranchID:   branch.ID,
		State:      string(branch.State),
		Priority:   branch.Priority,
		Confidence: branch.Confidence,
		Thoughts:   branch.Thoughts,
		Insights:   branch.Insights,
		CrossRefs:  branch.CrossRefs,
	}

	return history, nil
}

// SetActiveBranch changes the active branch
func (m *TreeMode) SetActiveBranch(ctx context.Context, branchID string) error {
	return m.storage.SetActiveBranch(branchID)
}
