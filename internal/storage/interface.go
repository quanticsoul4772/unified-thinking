package storage

import "unified-thinking/internal/types"

// ThoughtRepository manages thought persistence and retrieval
type ThoughtRepository interface {
	StoreThought(thought *types.Thought) error
	GetThought(id string) (*types.Thought, error)
	SearchThoughts(query string, mode types.ThinkingMode, limit, offset int) []*types.Thought
}

// BranchRepository manages branch operations
type BranchRepository interface {
	StoreBranch(branch *types.Branch) error
	GetBranch(id string) (*types.Branch, error)
	ListBranches() []*types.Branch
	GetActiveBranch() (*types.Branch, error)
	SetActiveBranch(branchID string) error
	UpdateBranchAccess(branchID string) error
	AppendThoughtToBranch(branchID string, thought *types.Thought) error
	AppendInsightToBranch(branchID string, insight *types.Insight) error
	AppendCrossRefToBranch(branchID string, crossRef *types.CrossRef) error
	UpdateBranchPriority(branchID string, priority float64) error
	UpdateBranchConfidence(branchID string, confidence float64) error
	GetRecentBranches() ([]*types.Branch, error)
}

// InsightRepository manages insights
type InsightRepository interface {
	StoreInsight(insight *types.Insight) error
}

// ValidationRepository manages validation results
type ValidationRepository interface {
	StoreValidation(validation *types.Validation) error
}

// RelationshipRepository manages thought relationships
type RelationshipRepository interface {
	StoreRelationship(relationship *types.Relationship) error
}

// MetricsProvider provides system metrics
type MetricsProvider interface {
	GetMetrics() *Metrics
}

// Storage combines all repository interfaces for unified access
// This is the main interface that modes and handlers should depend on
// Note: Trajectory persistence is handled separately via StorageBackend interface in memory package
type Storage interface {
	ThoughtRepository
	BranchRepository
	InsightRepository
	ValidationRepository
	RelationshipRepository
	MetricsProvider
}

// Verify MemoryStorage implements Storage interface
var _ Storage = (*MemoryStorage)(nil)
