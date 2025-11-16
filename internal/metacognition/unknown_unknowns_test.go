package metacognition

import (
	"context"
	"testing"
	"time"

	"unified-thinking/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestNewUnknownUnknownsDetector(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	assert.NotNil(t, detector)
	assert.NotNil(t, detector.knownPatterns)
	assert.NotNil(t, detector.domainChecklists)
	assert.NotEmpty(t, detector.knownPatterns)
	assert.NotEmpty(t, detector.domainChecklists)

	// Check initialized patterns
	assert.Contains(t, detector.knownPatterns, "confirmation_bias")
	assert.Contains(t, detector.knownPatterns, "incomplete")
	assert.Contains(t, detector.knownPatterns, "single_perspective")

	// Check domain checklists
	assert.Contains(t, detector.domainChecklists, "software")
	assert.Contains(t, detector.domainChecklists, "business")
}

func TestDetectBlindSpots_ConfirmationBias(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content:    "This confirms my hypothesis and validates my original point, as expected.",
		Confidence: 0.8,
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.BlindSpots)

	// Should detect confirmation bias
	hasConfirmationBias := false
	for _, spot := range result.BlindSpots {
		if spot.Type == BlindSpotConfirmationBias {
			hasConfirmationBias = true
			assert.NotEmpty(t, spot.Suggestions)
		}
	}
	assert.True(t, hasConfirmationBias, "Should detect confirmation bias")
}

func TestDetectBlindSpots_NarrowFraming(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content:    "This is the only option available, there's no alternative approach.",
		Confidence: 0.7,
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.BlindSpots)

	// Should detect narrow framing
	hasNarrowFraming := false
	for _, spot := range result.BlindSpots {
		if spot.Type == BlindSpotNarrowFraming {
			hasNarrowFraming = true
			assert.Greater(t, spot.Severity, 0.6)
			assert.Contains(t, spot.Suggestions, "What alternative approaches exist?")
		}
	}
	assert.True(t, hasNarrowFraming, "Should detect narrow framing")
}

func TestDetectBlindSpots_Overconfidence(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content:    "This is the solution.", // Very short
		Confidence: 0.95,                    // Very high confidence
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.BlindSpots)

	// Should detect overconfidence
	hasOverconfidence := false
	for _, spot := range result.BlindSpots {
		if spot.Type == BlindSpotOverconfidence {
			hasOverconfidence = true
			assert.GreaterOrEqual(t, spot.Severity, 0.8)
			assert.Contains(t, spot.Suggestions, "What could go wrong?")
		}
	}
	assert.True(t, hasOverconfidence, "Should detect overconfidence")
}

func TestDetectBlindSpots_NoOverconfidence_LongContent(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	// Long content with high confidence should NOT trigger overconfidence
	longContent := "This is a detailed analysis that covers multiple perspectives and considerations. " +
		"We've examined the problem from various angles including technical, business, and user perspectives. " +
		"The solution accounts for edge cases and potential failure modes."

	req := &GapAnalysisRequest{
		Content:    longContent,
		Confidence: 0.95,
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Should NOT detect overconfidence
	for _, spot := range result.BlindSpots {
		assert.NotEqual(t, BlindSpotOverconfidence, spot.Type)
	}
}

func TestDetectBlindSpots_DomainChecklist_Software(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content: "We'll build a new API endpoint to handle user requests.",
		Domain:  "software",
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.MissingConsiderations)

	// Should identify missing software considerations
	missingCount := 0
	for _, consideration := range result.MissingConsiderations {
		if consideration != "" {
			missingCount++
		}
	}
	assert.Greater(t, missingCount, 0, "Should identify missing software considerations")
}

func TestDetectBlindSpots_DomainChecklist_Business(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content: "We should launch this new product feature immediately.",
		Domain:  "business",
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.MissingConsiderations)

	// Should identify missing business considerations
	foundMissing := false
	for _, consideration := range result.MissingConsiderations {
		if consideration != "" {
			foundMissing = true
			break
		}
	}
	assert.True(t, foundMissing, "Should identify missing business considerations")
}

func TestDetectBlindSpots_ImplicitAssumptions(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content: "Users will always prefer the faster option, and this is definitely the best approach.",
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.UnchallengedAssumptions)

	// Should detect absolute statements
	foundAbsolute := false
	for _, assumption := range result.UnchallengedAssumptions {
		if assumption != "" {
			foundAbsolute = true
			break
		}
	}
	assert.True(t, foundAbsolute, "Should detect implicit assumptions from absolute statements")
}

func TestDetectBlindSpots_CausalClaims(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content: "Users left because the UI was slow, therefore we need to optimize performance.",
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.UnchallengedAssumptions)

	// Should question causal claims
	foundCausal := false
	for _, assumption := range result.UnchallengedAssumptions {
		if assumption != "" {
			foundCausal = true
			break
		}
	}
	assert.True(t, foundCausal, "Should detect causal claims")
}

func TestDetectBlindSpots_ProbingQuestions(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content: "This confirms my point, it's the only option.",
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.SuggestedQuestions)

	// Should generate probing questions
	assert.LessOrEqual(t, len(result.SuggestedQuestions), 10)
	assert.Contains(t, result.SuggestedQuestions, "What assumptions are you making that might not hold?")
}

func TestDetectBlindSpots_RiskCalculation_HighRisk(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content:    "Obviously this is the only way, and it confirms everything.",
		Confidence: 0.95, // High confidence
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Multiple blind spots + high confidence = higher risk
	if len(result.BlindSpots) > 0 {
		assert.Greater(t, result.OverallRisk, 0.0)
	}
}

func TestDetectBlindSpots_RiskCalculation_LowRisk(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content:    "Based on analysis of multiple perspectives, considering various alternatives and potential failure modes, this approach seems reasonable.",
		Confidence: 0.6,
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.LessOrEqual(t, result.OverallRisk, 1.0)
}

func TestDetectBlindSpots_AnalysisSummary(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content: "This confirms my point.",
	}

	result, err := detector.DetectBlindSpots(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Analysis)
	assert.Contains(t, result.Analysis, "Detected")
}

func TestDetectBlindSpots_EmptyContent(t *testing.T) {
	detector := NewUnknownUnknownsDetector()
	ctx := context.Background()

	req := &GapAnalysisRequest{
		Content: "",
	}

	_, err := detector.DetectBlindSpots(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content required")
}

func TestIdentifyKnowledgeGaps_UncertaintyMarkers(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	thought := &types.Thought{
		Content:    "I'm unsure about this approach, perhaps it might work but I don't know.",
		Confidence: 0.7,
	}

	gaps, err := detector.IdentifyKnowledgeGaps(thought)

	assert.NoError(t, err)
	assert.NotEmpty(t, gaps)

	// Should detect uncertainty markers
	foundUncertainty := false
	for _, gap := range gaps {
		if gap != "" {
			foundUncertainty = true
			break
		}
	}
	assert.True(t, foundUncertainty, "Should detect uncertainty markers")
}

func TestIdentifyKnowledgeGaps_Questions(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	thought := &types.Thought{
		Content:    "How should we approach this problem? What are the alternatives?",
		Confidence: 0.6,
	}

	gaps, err := detector.IdentifyKnowledgeGaps(thought)

	assert.NoError(t, err)
	assert.NotEmpty(t, gaps)

	// Should detect questions as knowledge gaps
	foundQuestion := false
	for _, gap := range gaps {
		if gap == "Questions present - indicates knowledge gaps" {
			foundQuestion = true
			break
		}
	}
	assert.True(t, foundQuestion, "Should detect questions as knowledge gaps")
}

func TestIdentifyKnowledgeGaps_LowConfidence(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	thought := &types.Thought{
		Content:    "This is an approach.",
		Confidence: 0.3, // Low confidence
	}

	gaps, err := detector.IdentifyKnowledgeGaps(thought)

	assert.NoError(t, err)
	assert.NotEmpty(t, gaps)

	// Should detect low confidence as gap
	foundLowConfidence := false
	for _, gap := range gaps {
		if gap != "" && gap != "Questions present - indicates knowledge gaps" {
			foundLowConfidence = true
			break
		}
	}
	assert.True(t, foundLowConfidence, "Should detect low confidence as knowledge gap")
}

func TestIdentifyKnowledgeGaps_NoGaps(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	thought := &types.Thought{
		Content:    "Based on thorough analysis, this is the recommended approach.",
		Confidence: 0.85,
	}

	gaps, err := detector.IdentifyKnowledgeGaps(thought)

	assert.NoError(t, err)
	// May be empty or have minimal gaps
	assert.NotNil(t, gaps)
}

func TestExtractImplicitAssumptions_Generalizations(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	content := "All users prefer fast interfaces and every developer knows this."

	assumptions := detector.extractImplicitAssumptions(content)

	assert.NotEmpty(t, assumptions)

	// Should detect generalizations
	foundGeneralization := false
	for _, assumption := range assumptions {
		if assumption == "Generalization detected - are there exceptions?" {
			foundGeneralization = true
			break
		}
	}
	assert.True(t, foundGeneralization)
}

func TestCalculateOverallRisk_NoBlindSpots(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	risk := detector.calculateOverallRisk([]*BlindSpot{}, 0.7)

	assert.Equal(t, 0.1, risk) // Minimal risk
}

func TestCalculateOverallRisk_WithBlindSpots_LowConfidence(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	blindSpots := []*BlindSpot{
		{Severity: 0.6},
		{Severity: 0.7},
	}

	risk := detector.calculateOverallRisk(blindSpots, 0.5)

	assert.Greater(t, risk, 0.1)
	assert.LessOrEqual(t, risk, 1.0)
}

func TestCalculateOverallRisk_WithBlindSpots_HighConfidence(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	blindSpots := []*BlindSpot{
		{Severity: 0.6},
		{Severity: 0.7},
	}

	risk := detector.calculateOverallRisk(blindSpots, 0.95)

	// High confidence + blind spots = increased risk
	assert.Greater(t, risk, 0.5)
	assert.LessOrEqual(t, risk, 1.0)
}

func TestSummarizeAnalysis_HighRisk(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	result := &GapAnalysisResult{
		BlindSpots: []*BlindSpot{
			{
				Type:        BlindSpotOverconfidence,
				Description: "High confidence issue",
			},
		},
		OverallRisk: 0.85,
	}

	summary := detector.summarizeAnalysis(result)

	assert.NotEmpty(t, summary)
	assert.Contains(t, summary, "HIGH RISK")
	assert.Contains(t, summary, "Detected")
}

func TestSummarizeAnalysis_ModerateRisk(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	result := &GapAnalysisResult{
		BlindSpots: []*BlindSpot{
			{
				Type:        BlindSpotNarrowFraming,
				Description: "Limited perspective",
			},
		},
		OverallRisk: 0.55,
	}

	summary := detector.summarizeAnalysis(result)

	assert.NotEmpty(t, summary)
	assert.Contains(t, summary, "MODERATE RISK")
}

func TestSummarizeAnalysis_LowRisk(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	result := &GapAnalysisResult{
		BlindSpots:  []*BlindSpot{},
		OverallRisk: 0.2,
	}

	summary := detector.summarizeAnalysis(result)

	assert.NotEmpty(t, summary)
	assert.Contains(t, summary, "LOW RISK")
}

func TestBlindSpotCreation(t *testing.T) {
	now := time.Now()
	spot := &BlindSpot{
		ID:          "test-1",
		Type:        BlindSpotConfirmationBias,
		Description: "Test blind spot",
		Severity:    0.7,
		Indicators:  []string{"confirms", "validates"},
		Suggestions: []string{"What contradicts this?"},
		DetectedAt:  now,
		Metadata:    map[string]interface{}{"source": "test"},
	}

	assert.Equal(t, "test-1", spot.ID)
	assert.Equal(t, BlindSpotConfirmationBias, spot.Type)
	assert.Equal(t, 0.7, spot.Severity)
	assert.Len(t, spot.Indicators, 2)
	assert.Len(t, spot.Suggestions, 1)
}

func TestDomainChecklist_Software(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	checklist := detector.domainChecklists["software"]

	assert.NotNil(t, checklist)
	assert.Equal(t, "software", checklist.Domain)
	assert.NotEmpty(t, checklist.MustConsider)
	assert.Contains(t, checklist.MustConsider, "scalability")
	assert.Contains(t, checklist.MustConsider, "security")
	assert.NotEmpty(t, checklist.CommonGaps)
}

func TestDomainChecklist_Business(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	checklist := detector.domainChecklists["business"]

	assert.NotNil(t, checklist)
	assert.Equal(t, "business", checklist.Domain)
	assert.NotEmpty(t, checklist.MustConsider)
	assert.Contains(t, checklist.MustConsider, "cost")
	assert.Contains(t, checklist.MustConsider, "revenue")
	assert.NotEmpty(t, checklist.CommonGaps)
}

func TestGenerateProbingQuestions_Limit(t *testing.T) {
	detector := NewUnknownUnknownsDetector()

	// Create many blind spots with suggestions
	blindSpots := make([]*BlindSpot, 0)
	for i := 0; i < 10; i++ {
		blindSpots = append(blindSpots, &BlindSpot{
			Suggestions: []string{"Question 1", "Question 2", "Question 3"},
		})
	}

	req := &GapAnalysisRequest{Content: "test"}
	questions := detector.generateProbingQuestions(req, blindSpots)

	// Should limit to 10 questions
	assert.LessOrEqual(t, len(questions), 10)
}
