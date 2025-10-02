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
	LastAccessedAt time.Time      `json:"last_accessed_at"`
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

// EvidenceQuality represents the quality assessment of evidence
type EvidenceQuality string

const (
	EvidenceStrong    EvidenceQuality = "strong"
	EvidenceModerate  EvidenceQuality = "moderate"
	EvidenceWeak      EvidenceQuality = "weak"
	EvidenceAnecdotal EvidenceQuality = "anecdotal"
)

// Evidence represents a piece of evidence supporting or refuting a claim
type Evidence struct {
	ID             string                 `json:"id"`
	Content        string                 `json:"content"`
	Source         string                 `json:"source,omitempty"`
	Quality        EvidenceQuality        `json:"quality"`
	Reliability    float64                `json:"reliability"`     // 0.0-1.0
	Relevance      float64                `json:"relevance"`       // 0.0-1.0
	OverallScore   float64                `json:"overall_score"`   // Computed from quality, reliability, relevance
	SupportsClaim  bool                   `json:"supports_claim"`  // true = supports, false = refutes
	ClaimID        string                 `json:"claim_id"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// ProbabilisticBelief represents a belief with associated probability
type ProbabilisticBelief struct {
	ID           string                 `json:"id"`
	Statement    string                 `json:"statement"`
	Probability  float64                `json:"probability"`      // 0.0-1.0 (Bayesian probability)
	PriorProb    float64                `json:"prior_prob"`       // Prior probability before evidence
	Evidence     []string               `json:"evidence"`         // Evidence IDs supporting this belief
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Contradiction represents detected contradictions between thoughts
type Contradiction struct {
	ID              string    `json:"id"`
	ThoughtID1      string    `json:"thought_id_1"`
	ThoughtID2      string    `json:"thought_id_2"`
	ContradictoryAt string    `json:"contradictory_at"` // Description of the contradiction
	Severity        string    `json:"severity"`         // "high", "medium", "low"
	DetectedAt      time.Time `json:"detected_at"`
}

// Perspective represents a stakeholder viewpoint
type Perspective struct {
	ID             string                 `json:"id"`
	Stakeholder    string                 `json:"stakeholder"`     // Name or role of stakeholder
	Viewpoint      string                 `json:"viewpoint"`       // Their perspective on the issue
	Concerns       []string               `json:"concerns"`        // Key concerns
	Priorities     []string               `json:"priorities"`      // What they prioritize
	Constraints    []string               `json:"constraints"`     // Constraints they face
	Confidence     float64                `json:"confidence"`      // Confidence in this perspective modeling
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// TemporalAnalysis represents short-term vs long-term reasoning
type TemporalAnalysis struct {
	ID              string                 `json:"id"`
	ShortTermView   string                 `json:"short_term_view"`    // Immediate implications
	LongTermView    string                 `json:"long_term_view"`     // Long-term implications
	TimeHorizon     string                 `json:"time_horizon"`       // "days", "months", "years"
	Tradeoffs       []string               `json:"tradeoffs"`          // Short vs long term tradeoffs
	Recommendation  string                 `json:"recommendation"`     // Which to prioritize and why
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// Decision represents a structured decision with options and criteria
type Decision struct {
	ID             string                 `json:"id"`
	Question       string                 `json:"question"`
	Options        []*DecisionOption      `json:"options"`
	Criteria       []*DecisionCriterion   `json:"criteria"`
	Recommendation string                 `json:"recommendation"`
	Confidence     float64                `json:"confidence"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// DecisionOption represents an option in a decision
type DecisionOption struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Scores      map[string]float64   `json:"scores"`      // criterion_id -> score
	Pros        []string             `json:"pros"`
	Cons        []string             `json:"cons"`
	TotalScore  float64              `json:"total_score"` // Weighted sum
}

// DecisionCriterion represents a criterion for evaluating options
type DecisionCriterion struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`      // Importance weight (0.0-1.0)
	Maximize    bool    `json:"maximize"`    // true = higher is better, false = lower is better
}

// ProblemDecomposition represents breaking down a complex problem
type ProblemDecomposition struct {
	ID             string                 `json:"id"`
	Problem        string                 `json:"problem"`
	Subproblems    []*Subproblem          `json:"subproblems"`
	Dependencies   []*Dependency          `json:"dependencies"`
	SolutionPath   []string               `json:"solution_path"`   // Order to solve subproblems
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Subproblem represents a component of a larger problem
type Subproblem struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Complexity  string   `json:"complexity"`  // "low", "medium", "high"
	Priority    string   `json:"priority"`    // "low", "medium", "high"
	Status      string   `json:"status"`      // "pending", "in_progress", "solved"
	Solution    string   `json:"solution,omitempty"`
}

// Dependency represents a dependency between subproblems
type Dependency struct {
	FromSubproblem string `json:"from_subproblem"` // Must be solved first
	ToSubproblem   string `json:"to_subproblem"`   // Depends on from_subproblem
	Type           string `json:"type"`            // "required", "optional", "informative"
}

// Synthesis represents cross-mode insight integration
type Synthesis struct {
	ID              string                 `json:"id"`
	Sources         []string               `json:"sources"`         // Thought/Insight IDs from different modes
	IntegratedView  string                 `json:"integrated_view"` // Synthesized conclusion
	Synergies       []string               `json:"synergies"`       // How sources complement each other
	Conflicts       []string               `json:"conflicts"`       // Conflicting aspects
	Confidence      float64                `json:"confidence"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// SensitivityAnalysis represents robustness testing of conclusions
type SensitivityAnalysis struct {
	ID             string                 `json:"id"`
	TargetClaim    string                 `json:"target_claim"`
	Variations     []*Variation           `json:"variations"`
	Robustness     float64                `json:"robustness"`      // 0.0-1.0, how stable is the conclusion
	KeyAssumptions []string               `json:"key_assumptions"` // Critical assumptions
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Variation represents a change in assumptions for sensitivity testing
type Variation struct {
	ID              string  `json:"id"`
	AssumptionChange string  `json:"assumption_change"`
	Impact          string  `json:"impact"`           // Description of how conclusion changes
	ImpactMagnitude float64 `json:"impact_magnitude"` // 0.0-1.0
}

// Analogy represents cross-domain reasoning
type Analogy struct {
	ID              string                 `json:"id"`
	SourceDomain    string                 `json:"source_domain"`
	TargetDomain    string                 `json:"target_domain"`
	Mapping         map[string]string      `json:"mapping"`         // source concept -> target concept
	Insight         string                 `json:"insight"`
	Strength        float64                `json:"strength"`        // 0.0-1.0
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// CausalGraph represents a causal model with variables and relationships
type CausalGraph struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Variables   []*CausalVariable      `json:"variables"`
	Links       []*CausalLink          `json:"links"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// CausalVariable represents a variable in a causal model
type CausalVariable struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"` // "binary", "continuous", "categorical"
	Observable  bool                   `json:"observable"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CausalLink represents a causal relationship between variables
type CausalLink struct {
	ID          string                 `json:"id"`
	From        string                 `json:"from"`        // Source variable ID
	To          string                 `json:"to"`          // Target variable ID
	Strength    float64                `json:"strength"`    // 0.0-1.0, strength of causal influence
	Type        string                 `json:"type"`        // "positive", "negative", "nonlinear"
	Confidence  float64                `json:"confidence"`  // 0.0-1.0, confidence in this link
	Evidence    []string               `json:"evidence,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CausalIntervention represents a hypothetical intervention and its effects
type CausalIntervention struct {
	ID                string                 `json:"id"`
	GraphID           string                 `json:"graph_id"`
	Variable          string                 `json:"variable"`        // Variable to intervene on
	InterventionType  string                 `json:"intervention_type"` // "set", "increase", "decrease"
	InterventionValue string                 `json:"intervention_value,omitempty"`
	PredictedEffects  []*PredictedEffect     `json:"predicted_effects"`
	Confidence        float64                `json:"confidence"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// PredictedEffect represents the predicted effect on a variable
type PredictedEffect struct {
	Variable     string  `json:"variable"`
	Effect       string  `json:"effect"` // "increase", "decrease", "no change"
	Magnitude    float64 `json:"magnitude,omitempty"` // Estimated magnitude (if quantifiable)
	Probability  float64 `json:"probability"` // 0.0-1.0
	Explanation  string  `json:"explanation"`
	PathLength   int     `json:"path_length"` // Number of causal steps
}

// Counterfactual represents a "what if" scenario
type Counterfactual struct {
	ID          string                 `json:"id"`
	GraphID     string                 `json:"graph_id"`
	Scenario    string                 `json:"scenario"` // Description of counterfactual
	Changes     map[string]string      `json:"changes"` // Variable -> counterfactual value
	Outcomes    map[string]string      `json:"outcomes"` // Variable -> predicted outcome
	Plausibility float64               `json:"plausibility"` // 0.0-1.0
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// SelfEvaluation represents metacognitive self-assessment
type SelfEvaluation struct {
	ID                  string                 `json:"id"`
	ThoughtID           string                 `json:"thought_id,omitempty"`
	BranchID            string                 `json:"branch_id,omitempty"`
	QualityScore        float64                `json:"quality_score"`        // 0.0-1.0
	CompletenessScore   float64                `json:"completeness_score"`   // 0.0-1.0
	CoherenceScore      float64                `json:"coherence_score"`      // 0.0-1.0
	Strengths           []string               `json:"strengths"`
	Weaknesses          []string               `json:"weaknesses"`
	ImprovementSuggestions []string            `json:"improvement_suggestions"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
}

// CognitiveBias represents detected cognitive biases
type CognitiveBias struct {
	ID          string                 `json:"id"`
	BiasType    string                 `json:"bias_type"`    // e.g., "confirmation", "anchoring", "availability"
	Description string                 `json:"description"`
	DetectedIn  string                 `json:"detected_in"`  // Thought ID or Branch ID
	Severity    string                 `json:"severity"`     // "low", "medium", "high"
	Mitigation  string                 `json:"mitigation"`   // How to address this bias
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}
