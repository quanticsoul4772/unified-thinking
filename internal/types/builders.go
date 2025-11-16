package types

import (
	"fmt"
	"time"
)

// ThoughtBuilder provides fluent API for thought construction
type ThoughtBuilder struct {
	thought *Thought
}

// NewThought creates a new ThoughtBuilder with sensible defaults
func NewThought() *ThoughtBuilder {
	return &ThoughtBuilder{
		thought: &Thought{
			Timestamp:  time.Now(),
			Confidence: 0.8, // Sensible default
			Metadata:   map[string]interface{}{},
		},
	}
}

// Content sets the thought content
func (b *ThoughtBuilder) Content(content string) *ThoughtBuilder {
	b.thought.Content = content
	return b
}

// Mode sets the thinking mode
func (b *ThoughtBuilder) Mode(mode ThinkingMode) *ThoughtBuilder {
	b.thought.Mode = mode
	return b
}

// Confidence sets the confidence level (overrides default)
func (b *ThoughtBuilder) Confidence(confidence float64) *ThoughtBuilder {
	if confidence > 0 {
		b.thought.Confidence = confidence
	}
	return b
}

// InBranch sets the branch ID
func (b *ThoughtBuilder) InBranch(branchID string) *ThoughtBuilder {
	b.thought.BranchID = branchID
	return b
}

// WithParent sets the parent thought ID
func (b *ThoughtBuilder) WithParent(parentID string) *ThoughtBuilder {
	b.thought.ParentID = parentID
	return b
}

// Type sets the thought type
func (b *ThoughtBuilder) Type(thoughtType string) *ThoughtBuilder {
	b.thought.Type = thoughtType
	return b
}

// KeyPoints sets the key points
func (b *ThoughtBuilder) KeyPoints(points []string) *ThoughtBuilder {
	b.thought.KeyPoints = points
	return b
}

// AsRebellion marks the thought as a rebellion
func (b *ThoughtBuilder) AsRebellion() *ThoughtBuilder {
	b.thought.IsRebellion = true
	return b
}

// ChallengesAssumptions marks that this thought challenges assumptions
func (b *ThoughtBuilder) ChallengesAssumptions() *ThoughtBuilder {
	b.thought.ChallengesAssumption = true
	return b
}

// WithMetadata sets a metadata key-value pair
func (b *ThoughtBuilder) WithMetadata(key string, value interface{}) *ThoughtBuilder {
	if b.thought.Metadata == nil {
		b.thought.Metadata = make(map[string]interface{})
	}
	b.thought.Metadata[key] = value
	return b
}

// Build returns the constructed thought
func (b *ThoughtBuilder) Build() *Thought {
	return b.thought
}

// Validate ensures thought meets minimum requirements
func (b *ThoughtBuilder) Validate() error {
	if b.thought.Content == "" {
		return fmt.Errorf("thought content cannot be empty")
	}
	if b.thought.Confidence < 0 || b.thought.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	return nil
}

// BranchBuilder provides fluent API for branch construction
type BranchBuilder struct {
	branch *Branch
}

// NewBranch creates a new BranchBuilder with sensible defaults
func NewBranch() *BranchBuilder {
	return &BranchBuilder{
		branch: &Branch{
			State:     StateActive,
			Priority:  1.0,
			Thoughts:  []*Thought{},
			Insights:  []*Insight{},
			CrossRefs: []*CrossRef{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// State sets the branch state
func (b *BranchBuilder) State(state ThoughtState) *BranchBuilder {
	b.branch.State = state
	return b
}

// Priority sets the branch priority
func (b *BranchBuilder) Priority(priority float64) *BranchBuilder {
	b.branch.Priority = priority
	return b
}

// Confidence sets the branch confidence (removed - not in Branch struct)
// func (b *BranchBuilder) Confidence(confidence float64) *BranchBuilder {
// 	b.branch.Confidence = confidence
// 	return b
// }

// ParentBranch sets the parent branch ID
func (b *BranchBuilder) ParentBranch(parentID string) *BranchBuilder {
	b.branch.ParentBranchID = parentID
	return b
}

// WithThought adds a thought to the branch
func (b *BranchBuilder) WithThought(thought *Thought) *BranchBuilder {
	b.branch.Thoughts = append(b.branch.Thoughts, thought)
	return b
}

// Build returns the constructed branch
func (b *BranchBuilder) Build() *Branch {
	return b.branch
}

// InsightBuilder provides fluent API for insight construction
type InsightBuilder struct {
	insight *Insight
}

// NewInsight creates a new InsightBuilder with sensible defaults
func NewInsight() *InsightBuilder {
	return &InsightBuilder{
		insight: &Insight{
			Type:               InsightObservation,
			ApplicabilityScore: 0.8,
			Context:            []string{},
			ParentInsights:     []string{},
			SupportingEvidence: map[string]interface{}{},
			Validations:        []*Validation{},
			CreatedAt:          time.Now(),
		},
	}
}

// Content sets the insight content
func (b *InsightBuilder) Content(content string) *InsightBuilder {
	b.insight.Content = content
	return b
}

// Type sets the insight type
func (b *InsightBuilder) Type(insightType InsightType) *InsightBuilder {
	b.insight.Type = insightType
	return b
}

// ApplicabilityScore sets the applicability score
func (b *InsightBuilder) ApplicabilityScore(score float64) *InsightBuilder {
	if score > 0 {
		b.insight.ApplicabilityScore = score
	}
	return b
}

// WithContext adds context information
func (b *InsightBuilder) WithContext(context ...string) *InsightBuilder {
	b.insight.Context = append(b.insight.Context, context...)
	return b
}

// WithEvidence adds supporting evidence
func (b *InsightBuilder) WithEvidence(key string, value interface{}) *InsightBuilder {
	if b.insight.SupportingEvidence == nil {
		b.insight.SupportingEvidence = make(map[string]interface{})
	}
	b.insight.SupportingEvidence[key] = value
	return b
}

// Build returns the constructed insight
func (b *InsightBuilder) Build() *Insight {
	return b.insight
}

// CrossRefBuilder provides fluent API for cross-reference construction
type CrossRefBuilder struct {
	crossRef *CrossRef
}

// NewCrossRef creates a new CrossRefBuilder
func NewCrossRef() *CrossRefBuilder {
	return &CrossRefBuilder{
		crossRef: &CrossRef{
			TouchPoints: []TouchPoint{},
		},
	}
}

// From sets the source branch
func (b *CrossRefBuilder) From(branchID string) *CrossRefBuilder {
	b.crossRef.FromBranch = branchID
	return b
}

// To sets the target branch
func (b *CrossRefBuilder) To(branchID string) *CrossRefBuilder {
	b.crossRef.ToBranch = branchID
	return b
}

// Type sets the cross-reference type
func (b *CrossRefBuilder) Type(refType CrossRefType) *CrossRefBuilder {
	b.crossRef.Type = refType
	return b
}

// Reason sets the reason for the cross-reference
func (b *CrossRefBuilder) Reason(reason string) *CrossRefBuilder {
	b.crossRef.Reason = reason
	return b
}

// Strength sets the relationship strength
func (b *CrossRefBuilder) Strength(strength float64) *CrossRefBuilder {
	b.crossRef.Strength = strength
	return b
}

// WithTouchPoint adds a touch point connecting specific thoughts
func (b *CrossRefBuilder) WithTouchPoint(fromThought, toThought, connection string) *CrossRefBuilder {
	touchPoint := TouchPoint{
		FromThought: fromThought,
		ToThought:   toThought,
		Connection:  connection,
	}
	b.crossRef.TouchPoints = append(b.crossRef.TouchPoints, touchPoint)
	return b
}

// Build returns the constructed cross-reference
func (b *CrossRefBuilder) Build() *CrossRef {
	return b.crossRef
}
