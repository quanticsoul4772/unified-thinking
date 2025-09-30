// Package types defines the core data structures for the unified thinking system.
//
// This package contains all type definitions for thoughts, branches, insights,
// cross-references, validations, and relationships. These types are used across
// all thinking modes (linear, tree, divergent, auto) and are designed to support
// concurrent access through deep copying in the storage layer.
//
// Key types:
//   - Thought: Represents a single unit of thinking with metadata
//   - Branch: Represents a parallel exploration path in tree mode
//   - Insight: Represents a discovered pattern or observation
//   - CrossRef: Represents connections between branches
//   - Validation: Represents logical consistency checks
package types

import "time"

// ThinkingMode represents the type of thinking
type ThinkingMode string

const (
	ModeLinear    ThinkingMode = "linear"
	ModeTree      ThinkingMode = "tree"
	ModeDivergent ThinkingMode = "divergent"
	ModeAuto      ThinkingMode = "auto"
)

// ThoughtState represents the state of a thought or branch.
type ThoughtState string

const (
	// StateActive indicates a thought or branch is currently being worked on
	StateActive ThoughtState = "active"
	// StateSuspended indicates a thought or branch is temporarily paused
	StateSuspended ThoughtState = "suspended"
	// StateCompleted indicates a thought or branch has been finished
	StateCompleted ThoughtState = "completed"
	// StateDeadEnd indicates a thought or branch has been abandoned
	StateDeadEnd ThoughtState = "dead_end"
)

// InsightType categorizes insights
type InsightType string

const (
	InsightBehavioralPattern  InsightType = "behavioral_pattern"
	InsightFeatureIntegration InsightType = "feature_integration"
	InsightObservation        InsightType = "observation"
	InsightConnection         InsightType = "connection"
)

// CrossRefType categorizes cross-references
type CrossRefType string

const (
	CrossRefComplementary CrossRefType = "complementary"
	CrossRefContradictory CrossRefType = "contradictory"
	CrossRefBuildsUpon    CrossRefType = "builds_upon"
	CrossRefAlternative   CrossRefType = "alternative"
)

// Thought represents a single thought in the system
type Thought struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Mode        ThinkingMode           `json:"mode"`
	BranchID    string                 `json:"branch_id,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Type        string                 `json:"type"`
	Confidence  float64                `json:"confidence"`
	Timestamp   time.Time              `json:"timestamp"`
	KeyPoints   []string               `json:"key_points,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`

	// Flags
	IsRebellion          bool `json:"is_rebellion"`
	ChallengesAssumption bool `json:"challenges_assumption"`
}

// Branch represents a branch in tree-mode thinking
type Branch struct {
	ID             string         `json:"id"`
	ParentBranchID string         `json:"parent_branch_id,omitempty"`
	State          ThoughtState   `json:"state"`
	Priority       float64        `json:"priority"`
	Confidence     float64        `json:"confidence"`
	Thoughts       []*Thought     `json:"thoughts"`
	Insights       []*Insight     `json:"insights"`
	CrossRefs      []*CrossRef    `json:"cross_refs"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// Insight represents a derived insight
type Insight struct {
	ID                 string                 `json:"id"`
	Type               InsightType            `json:"type"`
	Content            string                 `json:"content"`
	Context            []string               `json:"context"`
	ParentInsights     []string               `json:"parent_insights,omitempty"`
	ApplicabilityScore float64                `json:"applicability_score"`
	SupportingEvidence map[string]interface{} `json:"supporting_evidence"`
	Validations        []*Validation          `json:"validations,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

// CrossRef represents a cross-reference between branches
type CrossRef struct {
	ID          string       `json:"id"`
	FromBranch  string       `json:"from_branch"`
	ToBranch    string       `json:"to_branch"`
	Type        CrossRefType `json:"type"`
	Reason      string       `json:"reason"`
	Strength    float64      `json:"strength"`
	TouchPoints []TouchPoint `json:"touchpoints,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

// TouchPoint represents a connection point between thoughts
type TouchPoint struct {
	FromThought string `json:"from_thought"`
	ToThought   string `json:"to_thought"`
	Connection  string `json:"connection"`
}

// Validation represents logical validation results
type Validation struct {
	ID             string                 `json:"id"`
	InsightID      string                 `json:"insight_id,omitempty"`
	ThoughtID      string                 `json:"thought_id,omitempty"`
	IsValid        bool                   `json:"is_valid"`
	ValidationData map[string]interface{} `json:"validation_data,omitempty"`
	Reason         string                 `json:"reason,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Relationship represents connections between states
type Relationship struct {
	ID          string                 `json:"id"`
	FromStateID string                 `json:"from_state_id"`
	ToStateID   string                 `json:"to_state_id"`
	Type        string                 `json:"type"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}
