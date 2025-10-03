package types

import (
	"testing"
	"time"
)

func TestThinkingModeConstants(t *testing.T) {
	tests := []struct {
		name string
		mode ThinkingMode
		want string
	}{
		{"linear mode", ModeLinear, "linear"},
		{"tree mode", ModeTree, "tree"},
		{"divergent mode", ModeDivergent, "divergent"},
		{"auto mode", ModeAuto, "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.want {
				t.Errorf("ThinkingMode = %v, want %v", tt.mode, tt.want)
			}
		})
	}
}

func TestThoughtStateConstants(t *testing.T) {
	tests := []struct {
		name  string
		state ThoughtState
		want  string
	}{
		{"active state", StateActive, "active"},
		{"suspended state", StateSuspended, "suspended"},
		{"completed state", StateCompleted, "completed"},
		{"dead end state", StateDeadEnd, "dead_end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.state) != tt.want {
				t.Errorf("ThoughtState = %v, want %v", tt.state, tt.want)
			}
		})
	}
}

func TestInsightTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		typ  InsightType
		want string
	}{
		{"behavioral pattern", InsightBehavioralPattern, "behavioral_pattern"},
		{"feature integration", InsightFeatureIntegration, "feature_integration"},
		{"observation", InsightObservation, "observation"},
		{"connection", InsightConnection, "connection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.typ) != tt.want {
				t.Errorf("InsightType = %v, want %v", tt.typ, tt.want)
			}
		})
	}
}

func TestCrossRefTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		typ  CrossRefType
		want string
	}{
		{"complementary", CrossRefComplementary, "complementary"},
		{"contradictory", CrossRefContradictory, "contradictory"},
		{"builds upon", CrossRefBuildsUpon, "builds_upon"},
		{"alternative", CrossRefAlternative, "alternative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.typ) != tt.want {
				t.Errorf("CrossRefType = %v, want %v", tt.typ, tt.want)
			}
		})
	}
}

func TestEvidenceQualityConstants(t *testing.T) {
	tests := []struct {
		name    string
		quality EvidenceQuality
		want    string
	}{
		{"strong", EvidenceStrong, "strong"},
		{"moderate", EvidenceModerate, "moderate"},
		{"weak", EvidenceWeak, "weak"},
		{"anecdotal", EvidenceAnecdotal, "anecdotal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.quality) != tt.want {
				t.Errorf("EvidenceQuality = %v, want %v", tt.quality, tt.want)
			}
		})
	}
}

func TestThoughtStructure(t *testing.T) {
	now := time.Now()
	thought := &Thought{
		ID:                   "thought-1",
		Content:              "Test content",
		Mode:                 ModeLinear,
		BranchID:             "branch-1",
		ParentID:             "parent-1",
		Type:                 "observation",
		Confidence:           0.85,
		Timestamp:            now,
		KeyPoints:            []string{"point1", "point2"},
		Metadata:             map[string]interface{}{"key": "value"},
		IsRebellion:          true,
		ChallengesAssumption: true,
	}

	if thought.ID != "thought-1" {
		t.Errorf("Thought.ID = %v, want thought-1", thought.ID)
	}
	if thought.Content != "Test content" {
		t.Errorf("Thought.Content = %v, want Test content", thought.Content)
	}
	if thought.Mode != ModeLinear {
		t.Errorf("Thought.Mode = %v, want %v", thought.Mode, ModeLinear)
	}
	if thought.Confidence != 0.85 {
		t.Errorf("Thought.Confidence = %v, want 0.85", thought.Confidence)
	}
	if !thought.IsRebellion {
		t.Error("Thought.IsRebellion should be true")
	}
	if !thought.ChallengesAssumption {
		t.Error("Thought.ChallengesAssumption should be true")
	}
}

func TestBranchStructure(t *testing.T) {
	now := time.Now()
	branch := &Branch{
		ID:             "branch-1",
		ParentBranchID: "parent-branch",
		State:          StateActive,
		Priority:       0.9,
		Confidence:     0.85,
		Thoughts:       []*Thought{},
		Insights:       []*Insight{},
		CrossRefs:      []*CrossRef{},
		CreatedAt:      now,
		UpdatedAt:      now,
		LastAccessedAt: now,
	}

	if branch.ID != "branch-1" {
		t.Errorf("Branch.ID = %v, want branch-1", branch.ID)
	}
	if branch.State != StateActive {
		t.Errorf("Branch.State = %v, want %v", branch.State, StateActive)
	}
	if branch.Priority != 0.9 {
		t.Errorf("Branch.Priority = %v, want 0.9", branch.Priority)
	}
	if branch.Thoughts == nil {
		t.Error("Branch.Thoughts should not be nil")
	}
}

func TestInsightStructure(t *testing.T) {
	now := time.Now()
	insight := &Insight{
		ID:                 "insight-1",
		Type:               InsightObservation,
		Content:            "Test insight",
		Context:            []string{"ctx1", "ctx2"},
		ParentInsights:     []string{"parent1"},
		ApplicabilityScore: 0.8,
		SupportingEvidence: map[string]interface{}{"evidence": "data"},
		Validations:        []*Validation{},
		CreatedAt:          now,
	}

	if insight.ID != "insight-1" {
		t.Errorf("Insight.ID = %v, want insight-1", insight.ID)
	}
	if insight.Type != InsightObservation {
		t.Errorf("Insight.Type = %v, want %v", insight.Type, InsightObservation)
	}
	if insight.ApplicabilityScore != 0.8 {
		t.Errorf("Insight.ApplicabilityScore = %v, want 0.8", insight.ApplicabilityScore)
	}
}

func TestCrossRefStructure(t *testing.T) {
	now := time.Now()
	crossRef := &CrossRef{
		ID:         "xref-1",
		FromBranch: "branch-1",
		ToBranch:   "branch-2",
		Type:       CrossRefComplementary,
		Reason:     "They work well together",
		Strength:   0.9,
		TouchPoints: []TouchPoint{
			{FromThought: "t1", ToThought: "t2", Connection: "related"},
		},
		CreatedAt: now,
	}

	if crossRef.ID != "xref-1" {
		t.Errorf("CrossRef.ID = %v, want xref-1", crossRef.ID)
	}
	if crossRef.Type != CrossRefComplementary {
		t.Errorf("CrossRef.Type = %v, want %v", crossRef.Type, CrossRefComplementary)
	}
	if crossRef.Strength != 0.9 {
		t.Errorf("CrossRef.Strength = %v, want 0.9", crossRef.Strength)
	}
	if len(crossRef.TouchPoints) != 1 {
		t.Errorf("CrossRef.TouchPoints length = %v, want 1", len(crossRef.TouchPoints))
	}
}

func TestValidationStructure(t *testing.T) {
	now := time.Now()
	validation := &Validation{
		ID:             "val-1",
		InsightID:      "insight-1",
		ThoughtID:      "thought-1",
		IsValid:        true,
		ValidationData: map[string]interface{}{"score": 0.95},
		Reason:         "Logically sound",
		CreatedAt:      now,
	}

	if validation.ID != "val-1" {
		t.Errorf("Validation.ID = %v, want val-1", validation.ID)
	}
	if !validation.IsValid {
		t.Error("Validation.IsValid should be true")
	}
	if validation.Reason != "Logically sound" {
		t.Errorf("Validation.Reason = %v, want Logically sound", validation.Reason)
	}
}

func TestEvidenceStructure(t *testing.T) {
	now := time.Now()
	evidence := &Evidence{
		ID:            "ev-1",
		Content:       "Strong evidence",
		Source:        "Research paper",
		Quality:       EvidenceStrong,
		Reliability:   0.95,
		Relevance:     0.9,
		OverallScore:  0.925,
		SupportsClaim: true,
		ClaimID:       "claim-1",
		Metadata:      map[string]interface{}{"year": 2024},
		CreatedAt:     now,
	}

	if evidence.Quality != EvidenceStrong {
		t.Errorf("Evidence.Quality = %v, want %v", evidence.Quality, EvidenceStrong)
	}
	if evidence.Reliability != 0.95 {
		t.Errorf("Evidence.Reliability = %v, want 0.95", evidence.Reliability)
	}
	if !evidence.SupportsClaim {
		t.Error("Evidence.SupportsClaim should be true")
	}
}

func TestProbabilisticBeliefStructure(t *testing.T) {
	now := time.Now()
	belief := &ProbabilisticBelief{
		ID:          "belief-1",
		Statement:   "It will rain tomorrow",
		Probability: 0.7,
		PriorProb:   0.5,
		Evidence:    []string{"ev-1", "ev-2"},
		UpdatedAt:   now,
		Metadata:    map[string]interface{}{"source": "weather"},
	}

	if belief.Probability != 0.7 {
		t.Errorf("ProbabilisticBelief.Probability = %v, want 0.7", belief.Probability)
	}
	if belief.PriorProb != 0.5 {
		t.Errorf("ProbabilisticBelief.PriorProb = %v, want 0.5", belief.PriorProb)
	}
	if len(belief.Evidence) != 2 {
		t.Errorf("ProbabilisticBelief.Evidence length = %v, want 2", len(belief.Evidence))
	}
}

func TestDecisionStructure(t *testing.T) {
	now := time.Now()
	decision := &Decision{
		ID:       "dec-1",
		Question: "Which option to choose?",
		Options: []*DecisionOption{
			{ID: "opt-1", Name: "Option A", Scores: map[string]float64{"crit-1": 0.8}},
		},
		Criteria: []*DecisionCriterion{
			{ID: "crit-1", Name: "Cost", Weight: 0.5, Maximize: false},
		},
		Recommendation: "Choose Option A",
		Confidence:     0.85,
		Metadata:       map[string]interface{}{"method": "weighted"},
		CreatedAt:      now,
	}

	if len(decision.Options) != 1 {
		t.Errorf("Decision.Options length = %v, want 1", len(decision.Options))
	}
	if len(decision.Criteria) != 1 {
		t.Errorf("Decision.Criteria length = %v, want 1", len(decision.Criteria))
	}
	if decision.Confidence != 0.85 {
		t.Errorf("Decision.Confidence = %v, want 0.85", decision.Confidence)
	}
}

func TestPerspectiveStructure(t *testing.T) {
	now := time.Now()
	perspective := &Perspective{
		ID:          "persp-1",
		Stakeholder: "Customer",
		Viewpoint:   "Focus on usability",
		Concerns:    []string{"ease of use", "cost"},
		Priorities:  []string{"user experience", "value"},
		Constraints: []string{"budget", "time"},
		Confidence:  0.8,
		Metadata:    map[string]interface{}{"type": "end-user"},
		CreatedAt:   now,
	}

	if perspective.Stakeholder != "Customer" {
		t.Errorf("Perspective.Stakeholder = %v, want Customer", perspective.Stakeholder)
	}
	if len(perspective.Concerns) != 2 {
		t.Errorf("Perspective.Concerns length = %v, want 2", len(perspective.Concerns))
	}
}

func TestTemporalAnalysisStructure(t *testing.T) {
	now := time.Now()
	temporal := &TemporalAnalysis{
		ID:             "temp-1",
		ShortTermView:  "Quick wins",
		LongTermView:   "Sustainable growth",
		TimeHorizon:    "months",
		Tradeoffs:      []string{"speed vs quality"},
		Recommendation: "Balance both",
		Metadata:       map[string]interface{}{"priority": "high"},
		CreatedAt:      now,
	}

	if temporal.TimeHorizon != "months" {
		t.Errorf("TemporalAnalysis.TimeHorizon = %v, want months", temporal.TimeHorizon)
	}
	if len(temporal.Tradeoffs) != 1 {
		t.Errorf("TemporalAnalysis.Tradeoffs length = %v, want 1", len(temporal.Tradeoffs))
	}
}

func TestCausalGraphStructure(t *testing.T) {
	now := time.Now()
	graph := &CausalGraph{
		ID:          "graph-1",
		Description: "Test causal model",
		Variables: []*CausalVariable{
			{ID: "var-1", Name: "X", Type: "continuous", Observable: true},
		},
		Links: []*CausalLink{
			{ID: "link-1", From: "var-1", To: "var-2", Strength: 0.8, Type: "positive", Confidence: 0.9},
		},
		Metadata:  map[string]interface{}{"method": "structural"},
		CreatedAt: now,
	}

	if len(graph.Variables) != 1 {
		t.Errorf("CausalGraph.Variables length = %v, want 1", len(graph.Variables))
	}
	if len(graph.Links) != 1 {
		t.Errorf("CausalGraph.Links length = %v, want 1", len(graph.Links))
	}
}

func TestSelfEvaluationStructure(t *testing.T) {
	now := time.Now()
	eval := &SelfEvaluation{
		ID:                     "eval-1",
		ThoughtID:              "thought-1",
		QualityScore:           0.8,
		CompletenessScore:      0.9,
		CoherenceScore:         0.85,
		Strengths:              []string{"well-structured"},
		Weaknesses:             []string{"needs more evidence"},
		ImprovementSuggestions: []string{"add references"},
		Metadata:               map[string]interface{}{"version": "1.0"},
		CreatedAt:              now,
	}

	if eval.QualityScore != 0.8 {
		t.Errorf("SelfEvaluation.QualityScore = %v, want 0.8", eval.QualityScore)
	}
	if len(eval.Strengths) != 1 {
		t.Errorf("SelfEvaluation.Strengths length = %v, want 1", len(eval.Strengths))
	}
}

func TestCognitiveBiasStructure(t *testing.T) {
	now := time.Now()
	bias := &CognitiveBias{
		ID:          "bias-1",
		BiasType:    "confirmation",
		Description: "Seeking confirming evidence",
		DetectedIn:  "thought-1",
		Severity:    "medium",
		Mitigation:  "Seek disconfirming evidence",
		Metadata:    map[string]interface{}{"detected_by": "analyzer"},
		CreatedAt:   now,
	}

	if bias.BiasType != "confirmation" {
		t.Errorf("CognitiveBias.BiasType = %v, want confirmation", bias.BiasType)
	}
	if bias.Severity != "medium" {
		t.Errorf("CognitiveBias.Severity = %v, want medium", bias.Severity)
	}
}

func TestTouchPointStructure(t *testing.T) {
	tp := TouchPoint{
		FromThought: "thought-1",
		ToThought:   "thought-2",
		Connection:  "causally related",
	}

	if tp.FromThought != "thought-1" {
		t.Errorf("TouchPoint.FromThought = %v, want thought-1", tp.FromThought)
	}
	if tp.Connection != "causally related" {
		t.Errorf("TouchPoint.Connection = %v, want causally related", tp.Connection)
	}
}

func TestRelationshipStructure(t *testing.T) {
	now := time.Now()
	rel := &Relationship{
		ID:          "rel-1",
		FromStateID: "state-1",
		ToStateID:   "state-2",
		Type:        "transition",
		Metadata:    map[string]interface{}{"trigger": "event"},
		CreatedAt:   now,
	}

	if rel.Type != "transition" {
		t.Errorf("Relationship.Type = %v, want transition", rel.Type)
	}
}

func TestContradictionStructure(t *testing.T) {
	now := time.Now()
	contradiction := &Contradiction{
		ID:              "contra-1",
		ThoughtID1:      "thought-1",
		ThoughtID2:      "thought-2",
		ContradictoryAt: "claim about X",
		Severity:        "high",
		DetectedAt:      now,
	}

	if contradiction.Severity != "high" {
		t.Errorf("Contradiction.Severity = %v, want high", contradiction.Severity)
	}
}

func TestProblemDecompositionStructure(t *testing.T) {
	now := time.Now()
	decomp := &ProblemDecomposition{
		ID:      "decomp-1",
		Problem: "Complex problem",
		Subproblems: []*Subproblem{
			{ID: "sub-1", Description: "Part 1", Complexity: "medium", Priority: "high", Status: "pending"},
		},
		Dependencies: []*Dependency{
			{FromSubproblem: "sub-1", ToSubproblem: "sub-2", Type: "required"},
		},
		SolutionPath: []string{"sub-1", "sub-2"},
		Metadata:     map[string]interface{}{"method": "divide-conquer"},
		CreatedAt:    now,
	}

	if len(decomp.Subproblems) != 1 {
		t.Errorf("ProblemDecomposition.Subproblems length = %v, want 1", len(decomp.Subproblems))
	}
}

func TestSynthesisStructure(t *testing.T) {
	now := time.Now()
	synthesis := &Synthesis{
		ID:             "synth-1",
		Sources:        []string{"src-1", "src-2"},
		IntegratedView: "Combined perspective",
		Synergies:      []string{"complementary insights"},
		Conflicts:      []string{"opposing views"},
		Confidence:     0.8,
		Metadata:       map[string]interface{}{"method": "integration"},
		CreatedAt:      now,
	}

	if len(synthesis.Sources) != 2 {
		t.Errorf("Synthesis.Sources length = %v, want 2", len(synthesis.Sources))
	}
	if synthesis.Confidence != 0.8 {
		t.Errorf("Synthesis.Confidence = %v, want 0.8", synthesis.Confidence)
	}
}

func TestSensitivityAnalysisStructure(t *testing.T) {
	now := time.Now()
	sensitivity := &SensitivityAnalysis{
		ID:          "sens-1",
		TargetClaim: "Claim to test",
		Variations: []*Variation{
			{ID: "var-1", AssumptionChange: "Change A", Impact: "Significant", ImpactMagnitude: 0.8},
		},
		Robustness:     0.7,
		KeyAssumptions: []string{"assumption 1"},
		Metadata:       map[string]interface{}{"method": "monte-carlo"},
		CreatedAt:      now,
	}

	if sensitivity.Robustness != 0.7 {
		t.Errorf("SensitivityAnalysis.Robustness = %v, want 0.7", sensitivity.Robustness)
	}
	if len(sensitivity.Variations) != 1 {
		t.Errorf("SensitivityAnalysis.Variations length = %v, want 1", len(sensitivity.Variations))
	}
}

func TestAnalogyStructure(t *testing.T) {
	now := time.Now()
	analogy := &Analogy{
		ID:           "ana-1",
		SourceDomain: "Biology",
		TargetDomain: "Business",
		Mapping:      map[string]string{"evolution": "innovation"},
		Insight:      "Adaptation is key",
		Strength:     0.75,
		Metadata:     map[string]interface{}{"type": "structural"},
		CreatedAt:    now,
	}

	if analogy.Strength != 0.75 {
		t.Errorf("Analogy.Strength = %v, want 0.75", analogy.Strength)
	}
	if len(analogy.Mapping) != 1 {
		t.Errorf("Analogy.Mapping length = %v, want 1", len(analogy.Mapping))
	}
}

func TestCausalInterventionStructure(t *testing.T) {
	now := time.Now()
	intervention := &CausalIntervention{
		ID:               "int-1",
		GraphID:          "graph-1",
		Variable:         "var-1",
		InterventionType: "increase",
		PredictedEffects: []*PredictedEffect{
			{Variable: "var-2", Effect: "increase", Magnitude: 0.5, Probability: 0.8, Explanation: "causal link", PathLength: 1},
		},
		Confidence: 0.85,
		Metadata:   map[string]interface{}{"method": "do-calculus"},
		CreatedAt:  now,
	}

	if intervention.InterventionType != "increase" {
		t.Errorf("CausalIntervention.InterventionType = %v, want increase", intervention.InterventionType)
	}
	if len(intervention.PredictedEffects) != 1 {
		t.Errorf("CausalIntervention.PredictedEffects length = %v, want 1", len(intervention.PredictedEffects))
	}
}

func TestCounterfactualStructure(t *testing.T) {
	now := time.Now()
	counterfactual := &Counterfactual{
		ID:           "cf-1",
		GraphID:      "graph-1",
		Scenario:     "What if X was different",
		Changes:      map[string]string{"var-1": "high"},
		Outcomes:     map[string]string{"var-2": "low"},
		Plausibility: 0.7,
		Metadata:     map[string]interface{}{"method": "counterfactual-reasoning"},
		CreatedAt:    now,
	}

	if counterfactual.Plausibility != 0.7 {
		t.Errorf("Counterfactual.Plausibility = %v, want 0.7", counterfactual.Plausibility)
	}
	if len(counterfactual.Changes) != 1 {
		t.Errorf("Counterfactual.Changes length = %v, want 1", len(counterfactual.Changes))
	}
}
