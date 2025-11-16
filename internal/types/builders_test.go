package types

import (
	"testing"
	"time"
)

func TestNewThought(t *testing.T) {
	builder := NewThought()

	if builder == nil {
		t.Fatal("NewThought returned nil")
	}
	if builder.thought == nil {
		t.Fatal("ThoughtBuilder.thought is nil")
	}
	if builder.thought.Confidence != 0.8 {
		t.Errorf("Default confidence = %v, want 0.8", builder.thought.Confidence)
	}
	if builder.thought.Metadata == nil {
		t.Error("Metadata map not initialized")
	}
}

func TestThoughtBuilder_Content(t *testing.T) {
	thought := NewThought().
		Content("Test content").
		Build()

	if thought.Content != "Test content" {
		t.Errorf("Content = %v, want Test content", thought.Content)
	}
}

func TestThoughtBuilder_Mode(t *testing.T) {
	thought := NewThought().
		Mode(ModeTree).
		Build()

	if thought.Mode != ModeTree {
		t.Errorf("Mode = %v, want %v", thought.Mode, ModeTree)
	}
}

func TestThoughtBuilder_Confidence(t *testing.T) {
	tests := []struct {
		name       string
		confidence float64
		want       float64
	}{
		{"positive confidence", 0.95, 0.95},
		{"zero confidence", 0.0, 0.8},      // Should keep default
		{"negative confidence", -0.5, 0.8}, // Should keep default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thought := NewThought().
				Confidence(tt.confidence).
				Build()

			if thought.Confidence != tt.want {
				t.Errorf("Confidence = %v, want %v", thought.Confidence, tt.want)
			}
		})
	}
}

func TestThoughtBuilder_InBranch(t *testing.T) {
	thought := NewThought().
		InBranch("branch-123").
		Build()

	if thought.BranchID != "branch-123" {
		t.Errorf("BranchID = %v, want branch-123", thought.BranchID)
	}
}

func TestThoughtBuilder_WithParent(t *testing.T) {
	thought := NewThought().
		WithParent("parent-456").
		Build()

	if thought.ParentID != "parent-456" {
		t.Errorf("ParentID = %v, want parent-456", thought.ParentID)
	}
}

func TestThoughtBuilder_Type(t *testing.T) {
	thought := NewThought().
		Type("analysis").
		Build()

	if thought.Type != "analysis" {
		t.Errorf("Type = %v, want analysis", thought.Type)
	}
}

func TestThoughtBuilder_KeyPoints(t *testing.T) {
	points := []string{"point1", "point2", "point3"}
	thought := NewThought().
		KeyPoints(points).
		Build()

	if len(thought.KeyPoints) != 3 {
		t.Errorf("KeyPoints length = %v, want 3", len(thought.KeyPoints))
	}
	if thought.KeyPoints[0] != "point1" {
		t.Errorf("KeyPoints[0] = %v, want point1", thought.KeyPoints[0])
	}
}

func TestThoughtBuilder_AsRebellion(t *testing.T) {
	thought := NewThought().
		AsRebellion().
		Build()

	if !thought.IsRebellion {
		t.Error("IsRebellion should be true")
	}
}

func TestThoughtBuilder_ChallengesAssumptions(t *testing.T) {
	thought := NewThought().
		ChallengesAssumptions().
		Build()

	if !thought.ChallengesAssumption {
		t.Error("ChallengesAssumption should be true")
	}
}

func TestThoughtBuilder_WithMetadata(t *testing.T) {
	thought := NewThought().
		WithMetadata("key1", "value1").
		WithMetadata("key2", 123).
		Build()

	if thought.Metadata["key1"] != "value1" {
		t.Errorf("Metadata[key1] = %v, want value1", thought.Metadata["key1"])
	}
	if thought.Metadata["key2"] != 123 {
		t.Errorf("Metadata[key2] = %v, want 123", thought.Metadata["key2"])
	}
}

func TestThoughtBuilder_WithMetadata_NilMap(t *testing.T) {
	builder := &ThoughtBuilder{
		thought: &Thought{
			Metadata: nil,
		},
	}

	thought := builder.
		WithMetadata("key", "value").
		Build()

	if thought.Metadata == nil {
		t.Fatal("Metadata map not initialized")
	}
	if thought.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want value", thought.Metadata["key"])
	}
}

func TestThoughtBuilder_Fluent(t *testing.T) {
	thought := NewThought().
		Content("Complex thought").
		Mode(ModeDivergent).
		Type("creative").
		Confidence(0.9).
		InBranch("branch-1").
		WithParent("parent-1").
		KeyPoints([]string{"point1", "point2"}).
		AsRebellion().
		ChallengesAssumptions().
		WithMetadata("source", "brainstorming").
		Build()

	if thought.Content != "Complex thought" {
		t.Errorf("Content = %v, want Complex thought", thought.Content)
	}
	if thought.Mode != ModeDivergent {
		t.Errorf("Mode = %v, want %v", thought.Mode, ModeDivergent)
	}
	if thought.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", thought.Confidence)
	}
	if !thought.IsRebellion {
		t.Error("IsRebellion should be true")
	}
	if !thought.ChallengesAssumption {
		t.Error("ChallengesAssumption should be true")
	}
}

func TestThoughtBuilder_Validate(t *testing.T) {
	tests := []struct {
		name    string
		builder *ThoughtBuilder
		wantErr bool
	}{
		{
			name: "valid thought",
			builder: NewThought().
				Content("Valid content").
				Confidence(0.8),
			wantErr: false,
		},
		{
			name: "empty content",
			builder: NewThought().
				Content(""),
			wantErr: true,
		},
		{
			name: "confidence at boundary 0",
			builder: NewThought().
				Content("Content"),
			wantErr: false,
		},
		{
			name: "confidence at boundary 1",
			builder: NewThought().
				Content("Content").
				Confidence(1.0),
			wantErr: false,
		},
		{
			name: "confidence too high",
			builder: NewThought().
				Content("Content").
				Confidence(1.1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.builder.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBranch(t *testing.T) {
	builder := NewBranch()

	if builder == nil {
		t.Fatal("NewBranch returned nil")
	}
	if builder.branch == nil {
		t.Fatal("BranchBuilder.branch is nil")
	}
	if builder.branch.State != StateActive {
		t.Errorf("Default state = %v, want %v", builder.branch.State, StateActive)
	}
	if builder.branch.Priority != 1.0 {
		t.Errorf("Default priority = %v, want 1.0", builder.branch.Priority)
	}
	if builder.branch.Thoughts == nil {
		t.Error("Thoughts slice not initialized")
	}
	if builder.branch.Insights == nil {
		t.Error("Insights slice not initialized")
	}
	if builder.branch.CrossRefs == nil {
		t.Error("CrossRefs slice not initialized")
	}
}

func TestBranchBuilder_State(t *testing.T) {
	branch := NewBranch().
		State(StateCompleted).
		Build()

	if branch.State != StateCompleted {
		t.Errorf("State = %v, want %v", branch.State, StateCompleted)
	}
}

func TestBranchBuilder_Priority(t *testing.T) {
	branch := NewBranch().
		Priority(0.75).
		Build()

	if branch.Priority != 0.75 {
		t.Errorf("Priority = %v, want 0.75", branch.Priority)
	}
}

func TestBranchBuilder_ParentBranch(t *testing.T) {
	branch := NewBranch().
		ParentBranch("parent-branch-1").
		Build()

	if branch.ParentBranchID != "parent-branch-1" {
		t.Errorf("ParentBranchID = %v, want parent-branch-1", branch.ParentBranchID)
	}
}

func TestBranchBuilder_WithThought(t *testing.T) {
	thought1 := &Thought{ID: "t1", Content: "Thought 1"}
	thought2 := &Thought{ID: "t2", Content: "Thought 2"}

	branch := NewBranch().
		WithThought(thought1).
		WithThought(thought2).
		Build()

	if len(branch.Thoughts) != 2 {
		t.Errorf("Thoughts length = %v, want 2", len(branch.Thoughts))
	}
	if branch.Thoughts[0].ID != "t1" {
		t.Errorf("Thoughts[0].ID = %v, want t1", branch.Thoughts[0].ID)
	}
	if branch.Thoughts[1].ID != "t2" {
		t.Errorf("Thoughts[1].ID = %v, want t2", branch.Thoughts[1].ID)
	}
}

func TestBranchBuilder_Fluent(t *testing.T) {
	branch := NewBranch().
		State(StateSuspended).
		Priority(0.5).
		ParentBranch("parent-1").
		WithThought(&Thought{ID: "t1"}).
		Build()

	if branch.State != StateSuspended {
		t.Errorf("State = %v, want %v", branch.State, StateSuspended)
	}
	if branch.Priority != 0.5 {
		t.Errorf("Priority = %v, want 0.5", branch.Priority)
	}
	if branch.ParentBranchID != "parent-1" {
		t.Errorf("ParentBranchID = %v, want parent-1", branch.ParentBranchID)
	}
	if len(branch.Thoughts) != 1 {
		t.Errorf("Thoughts length = %v, want 1", len(branch.Thoughts))
	}
}

func TestBranchBuilder_Timestamps(t *testing.T) {
	before := time.Now()
	branch := NewBranch().Build()
	after := time.Now()

	if branch.CreatedAt.Before(before) || branch.CreatedAt.After(after) {
		t.Error("CreatedAt timestamp not in expected range")
	}
	if branch.UpdatedAt.Before(before) || branch.UpdatedAt.After(after) {
		t.Error("UpdatedAt timestamp not in expected range")
	}
}

func TestNewInsight(t *testing.T) {
	builder := NewInsight()

	if builder == nil {
		t.Fatal("NewInsight returned nil")
	}
	if builder.insight == nil {
		t.Fatal("InsightBuilder.insight is nil")
	}
	if builder.insight.Type != InsightObservation {
		t.Errorf("Default type = %v, want %v", builder.insight.Type, InsightObservation)
	}
	if builder.insight.ApplicabilityScore != 0.8 {
		t.Errorf("Default applicability = %v, want 0.8", builder.insight.ApplicabilityScore)
	}
}

func TestInsightBuilder_Content(t *testing.T) {
	insight := NewInsight().
		Content("Interesting pattern").
		Build()

	if insight.Content != "Interesting pattern" {
		t.Errorf("Content = %v, want Interesting pattern", insight.Content)
	}
}

func TestInsightBuilder_Type(t *testing.T) {
	insight := NewInsight().
		Type(InsightConnection).
		Build()

	if insight.Type != InsightConnection {
		t.Errorf("Type = %v, want %v", insight.Type, InsightConnection)
	}
}

func TestInsightBuilder_ApplicabilityScore(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  float64
	}{
		{"positive score", 0.95, 0.95},
		{"zero score", 0.0, 0.8},      // Should keep default
		{"negative score", -0.5, 0.8}, // Should keep default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insight := NewInsight().
				ApplicabilityScore(tt.score).
				Build()

			if insight.ApplicabilityScore != tt.want {
				t.Errorf("ApplicabilityScore = %v, want %v", insight.ApplicabilityScore, tt.want)
			}
		})
	}
}

func TestInsightBuilder_WithContext(t *testing.T) {
	insight := NewInsight().
		WithContext("context1", "context2").
		WithContext("context3").
		Build()

	if len(insight.Context) != 3 {
		t.Errorf("Context length = %v, want 3", len(insight.Context))
	}
	if insight.Context[0] != "context1" {
		t.Errorf("Context[0] = %v, want context1", insight.Context[0])
	}
}

func TestInsightBuilder_WithEvidence(t *testing.T) {
	insight := NewInsight().
		WithEvidence("key1", "value1").
		WithEvidence("key2", 123).
		Build()

	if insight.SupportingEvidence["key1"] != "value1" {
		t.Errorf("SupportingEvidence[key1] = %v, want value1", insight.SupportingEvidence["key1"])
	}
	if insight.SupportingEvidence["key2"] != 123 {
		t.Errorf("SupportingEvidence[key2] = %v, want 123", insight.SupportingEvidence["key2"])
	}
}

func TestInsightBuilder_WithEvidence_NilMap(t *testing.T) {
	builder := &InsightBuilder{
		insight: &Insight{
			SupportingEvidence: nil,
		},
	}

	insight := builder.
		WithEvidence("key", "value").
		Build()

	if insight.SupportingEvidence == nil {
		t.Fatal("SupportingEvidence map not initialized")
	}
	if insight.SupportingEvidence["key"] != "value" {
		t.Errorf("SupportingEvidence[key] = %v, want value", insight.SupportingEvidence["key"])
	}
}

func TestInsightBuilder_Fluent(t *testing.T) {
	insight := NewInsight().
		Content("Complex pattern").
		Type(InsightBehavioralPattern).
		ApplicabilityScore(0.95).
		WithContext("ctx1", "ctx2").
		WithEvidence("source", "research").
		Build()

	if insight.Content != "Complex pattern" {
		t.Errorf("Content = %v, want Complex pattern", insight.Content)
	}
	if insight.Type != InsightBehavioralPattern {
		t.Errorf("Type = %v, want %v", insight.Type, InsightBehavioralPattern)
	}
	if len(insight.Context) != 2 {
		t.Errorf("Context length = %v, want 2", len(insight.Context))
	}
}

func TestNewCrossRef(t *testing.T) {
	builder := NewCrossRef()

	if builder == nil {
		t.Fatal("NewCrossRef returned nil")
	}
	if builder.crossRef == nil {
		t.Fatal("CrossRefBuilder.crossRef is nil")
	}
	if builder.crossRef.TouchPoints == nil {
		t.Error("TouchPoints slice not initialized")
	}
}

func TestCrossRefBuilder_From(t *testing.T) {
	xref := NewCrossRef().
		From("branch-1").
		Build()

	if xref.FromBranch != "branch-1" {
		t.Errorf("FromBranch = %v, want branch-1", xref.FromBranch)
	}
}

func TestCrossRefBuilder_To(t *testing.T) {
	xref := NewCrossRef().
		To("branch-2").
		Build()

	if xref.ToBranch != "branch-2" {
		t.Errorf("ToBranch = %v, want branch-2", xref.ToBranch)
	}
}

func TestCrossRefBuilder_Type(t *testing.T) {
	xref := NewCrossRef().
		Type(CrossRefBuildsUpon).
		Build()

	if xref.Type != CrossRefBuildsUpon {
		t.Errorf("Type = %v, want %v", xref.Type, CrossRefBuildsUpon)
	}
}

func TestCrossRefBuilder_Reason(t *testing.T) {
	xref := NewCrossRef().
		Reason("They complement each other").
		Build()

	if xref.Reason != "They complement each other" {
		t.Errorf("Reason = %v, want They complement each other", xref.Reason)
	}
}

func TestCrossRefBuilder_Strength(t *testing.T) {
	xref := NewCrossRef().
		Strength(0.85).
		Build()

	if xref.Strength != 0.85 {
		t.Errorf("Strength = %v, want 0.85", xref.Strength)
	}
}

func TestCrossRefBuilder_WithTouchPoint(t *testing.T) {
	xref := NewCrossRef().
		WithTouchPoint("thought-1", "thought-2", "related").
		WithTouchPoint("thought-3", "thought-4", "contradicts").
		Build()

	if len(xref.TouchPoints) != 2 {
		t.Errorf("TouchPoints length = %v, want 2", len(xref.TouchPoints))
	}
	if xref.TouchPoints[0].FromThought != "thought-1" {
		t.Errorf("TouchPoints[0].FromThought = %v, want thought-1", xref.TouchPoints[0].FromThought)
	}
	if xref.TouchPoints[0].Connection != "related" {
		t.Errorf("TouchPoints[0].Connection = %v, want related", xref.TouchPoints[0].Connection)
	}
}

func TestCrossRefBuilder_Fluent(t *testing.T) {
	xref := NewCrossRef().
		From("branch-1").
		To("branch-2").
		Type(CrossRefComplementary).
		Reason("Work together well").
		Strength(0.9).
		WithTouchPoint("t1", "t2", "supports").
		Build()

	if xref.FromBranch != "branch-1" {
		t.Errorf("FromBranch = %v, want branch-1", xref.FromBranch)
	}
	if xref.ToBranch != "branch-2" {
		t.Errorf("ToBranch = %v, want branch-2", xref.ToBranch)
	}
	if xref.Type != CrossRefComplementary {
		t.Errorf("Type = %v, want %v", xref.Type, CrossRefComplementary)
	}
	if xref.Strength != 0.9 {
		t.Errorf("Strength = %v, want 0.9", xref.Strength)
	}
	if len(xref.TouchPoints) != 1 {
		t.Errorf("TouchPoints length = %v, want 1", len(xref.TouchPoints))
	}
}
