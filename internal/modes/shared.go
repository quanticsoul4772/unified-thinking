// Package modes implements different cognitive thinking patterns for the unified system.
//
// This package provides four thinking modes:
//   - LinearMode: Sequential step-by-step reasoning
//   - TreeMode: Multi-branch parallel exploration with cross-references
//   - DivergentMode: Creative and unconventional ideation
//   - AutoMode: Automatic mode selection based on input analysis
//
// Each mode implements thought processing appropriate to its cognitive pattern
// and uses the shared storage layer for persistence.
package modes

import "unified-thinking/internal/types"

// ThoughtInput represents input parameters for thought processing across all modes.
//
// Field usage:
//   - Content: The thought content to process (required)
//   - Type: Type of thought (optional, mode-specific)
//   - BranchID: For tree mode, specifies which branch to use
//   - ParentID: For tree mode, parent thought in the branch
//   - PreviousThoughtID: For linear mode, previous step in sequence
//   - Confidence: Confidence level (0.0-1.0, defaults to 0.8)
//   - KeyPoints: Important points extracted from the thought
//   - ForceRebellion: For divergent mode, forces unconventional thinking
//   - CrossRefs: For tree mode, cross-references to other branches
type ThoughtInput struct {
	Content           string
	Type              string
	BranchID          string
	ParentID          string
	PreviousThoughtID string
	Confidence        float64
	KeyPoints         []string
	ForceRebellion    bool
	CrossRefs         []CrossRefInput
}

// CrossRefInput represents a cross-reference input
type CrossRefInput struct {
	ToBranch string
	Type     string
	Reason   string
	Strength float64
}

// ThoughtResult represents the result of processing a thought
type ThoughtResult struct {
	ThoughtID            string
	Mode                 string
	BranchID             string
	BranchState          string
	Status               string
	Priority             float64
	Confidence           float64
	InsightCount         int
	CrossRefCount        int
	Content              string
	Direction            string
	IsRebellion          bool
	ChallengesAssumption bool
}

// BranchHistory represents detailed branch history
type BranchHistory struct {
	BranchID   string
	State      string
	Priority   float64
	Confidence float64
	Thoughts   []*types.Thought
	Insights   []*types.Insight
	CrossRefs  []*types.CrossRef
}
