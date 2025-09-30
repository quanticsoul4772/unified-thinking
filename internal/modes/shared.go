package modes

import "unified-thinking/internal/types"

// ThoughtInput represents input for thinking
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
