package validation

import (
	"context"
	"testing"
	"time"

	"unified-thinking/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestHallucinationDetector_FastVerification(t *testing.T) {
	detector := NewHallucinationDetector()

	tests := []struct {
		name               string
		thought            *types.Thought
		expectedRiskMin    float64
		expectedRiskMax    float64
		expectRecommendations bool
	}{
		{
			name: "high confidence with uncertainty markers",
			thought: &types.Thought{
				ID:         "test-1",
				Content:    "I'm not sure, but maybe this might possibly work.",
				Confidence: 0.9,
			},
			expectedRiskMin:    0.3,
			expectedRiskMax:    0.8,
			expectRecommendations: true,
		},
		{
			name: "low confidence with definitive language",
			thought: &types.Thought{
				ID:         "test-2",
				Content:    "This definitely always works and must be true.",
				Confidence: 0.3,
			},
			expectedRiskMin:    0.2,
			expectedRiskMax:    0.7,
			expectRecommendations: true,
		},
		{
			name: "consistent high confidence",
			thought: &types.Thought{
				ID:         "test-3",
				Content:    "PostgreSQL uses MVCC for transaction isolation.",
				Confidence: 0.85,
			},
			expectedRiskMin:    0.0,
			expectedRiskMax:    0.4,
			expectRecommendations: false,
		},
		{
			name: "causal claim extraction",
			thought: &types.Thought{
				ID:         "test-4",
				Content:    "Smoking causes lung cancer. This leads to higher mortality rates.",
				Confidence: 0.8,
			},
			expectedRiskMin:    0.0,
			expectedRiskMax:    0.5,
			expectRecommendations: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := detector.fastVerification(tt.thought)

			assert.NotNil(t, report)
			assert.Equal(t, tt.thought.ID, report.ThoughtID)
			assert.Equal(t, VerificationFast, report.VerificationLevel)

			assert.GreaterOrEqual(t, report.OverallRisk, tt.expectedRiskMin,
				"Risk should be at least %f, got %f", tt.expectedRiskMin, report.OverallRisk)
			assert.LessOrEqual(t, report.OverallRisk, tt.expectedRiskMax,
				"Risk should be at most %f, got %f", tt.expectedRiskMax, report.OverallRisk)

			if tt.expectRecommendations {
				assert.NotEmpty(t, report.Recommendations, "Expected recommendations")
			}

			// Verify semantic uncertainty is calculated
			assert.NotNil(t, report.SemanticUncertainty)
			assert.GreaterOrEqual(t, report.SemanticUncertainty.Overall, 0.0)
			assert.LessOrEqual(t, report.SemanticUncertainty.Overall, 1.0)
		})
	}
}

func TestHallucinationDetector_ClaimExtraction(t *testing.T) {
	detector := NewHallucinationDetector()

	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectedTypes []ClaimType
	}{
		{
			name:          "causal claims",
			content:       "Smoking causes cancer. This results in health problems.",
			expectedCount: 2,
			expectedTypes: []ClaimType{ClaimCausal, ClaimCausal},
		},
		{
			name:          "statistical claims",
			content:       "95% of users prefer this. The average response time is 100ms.",
			expectedCount: 2,
			expectedTypes: []ClaimType{ClaimStatistical, ClaimStatistical},
		},
		{
			name:          "mixed claims",
			content:       "PostgreSQL uses MVCC. This causes better concurrency. Studies show 80% improvement.",
			expectedCount: 3,
			expectedTypes: []ClaimType{ClaimFactual, ClaimCausal, ClaimStatistical},
		},
		{
			name:          "opinions filtered out",
			content:       "I think this is good. I believe it works. This is a fact.",
			expectedCount: 1,
			expectedTypes: []ClaimType{ClaimFactual},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := detector.extractClaims(tt.content)

			assert.Len(t, claims, tt.expectedCount, "Expected %d claims, got %d", tt.expectedCount, len(claims))

			for i, claim := range claims {
				assert.Equal(t, tt.expectedTypes[i], claim.ClaimType,
					"Claim %d should be type %s, got %s", i, tt.expectedTypes[i], claim.ClaimType)
				assert.Equal(t, StatusPending, claim.VerificationStatus)
			}
		})
	}
}

func TestHallucinationDetector_UncertaintyCategorization(t *testing.T) {
	detector := NewHallucinationDetector()

	tests := []struct {
		uncertainty     float64
		expectedType    UncertaintyType
	}{
		{0.1, UncertaintyLow},
		{0.25, UncertaintyLow},
		{0.4, UncertaintyModerate},
		{0.55, UncertaintyModerate},
		{0.7, UncertaintyHigh},
		{0.75, UncertaintyHigh},
		{0.85, UncertaintyCritical},
		{0.95, UncertaintyCritical},
	}

	for _, tt := range tests {
		result := detector.categorizeUncertainty(tt.uncertainty)
		assert.Equal(t, tt.expectedType, result,
			"Uncertainty %f should be %s, got %s", tt.uncertainty, tt.expectedType, result)
	}
}

func TestHallucinationDetector_CoherenceCheck(t *testing.T) {
	detector := NewHallucinationDetector()

	tests := []struct {
		name     string
		content  string
		minScore float64
		maxScore float64
	}{
		{
			name:     "coherent text",
			content:  "PostgreSQL is a relational database. It supports ACID transactions.",
			minScore: 0.9,
			maxScore: 1.0,
		},
		{
			name:     "uncertain text",
			content:  "Maybe this works. Possibly it could help. Perhaps unclear.",
			minScore: 0.7,
			maxScore: 0.85,
		},
		{
			name:     "contradictory indicators",
			content:  "This works well. However, it might not work. But it definitely does.",
			minScore: 0.7,
			maxScore: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.checkInternalCoherence(tt.content)

			assert.GreaterOrEqual(t, score, tt.minScore,
				"Coherence score should be at least %f, got %f", tt.minScore, score)
			assert.LessOrEqual(t, score, tt.maxScore,
				"Coherence score should be at most %f, got %f", tt.maxScore, score)
		})
	}
}

func TestHallucinationDetector_ConfidenceMismatch(t *testing.T) {
	detector := NewHallucinationDetector()

	tests := []struct {
		name              string
		thought           *types.Thought
		expectedMismatch  float64
		tolerance         float64
	}{
		{
			name: "high confidence with uncertainty",
			thought: &types.Thought{
				Content:    "Maybe this might possibly work.",
				Confidence: 0.9,
			},
			expectedMismatch: 0.4,
			tolerance:        0.2,
		},
		{
			name: "low confidence with definitives",
			thought: &types.Thought{
				Content:    "This definitely always works.",
				Confidence: 0.3,
			},
			expectedMismatch: 0.3,
			tolerance:        0.2,
		},
		{
			name: "appropriate confidence",
			thought: &types.Thought{
				Content:    "This generally works in most cases.",
				Confidence: 0.7,
			},
			expectedMismatch: 0.0,
			tolerance:        0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mismatch := detector.estimateConfidenceMismatch(tt.thought)

			assert.InDelta(t, tt.expectedMismatch, mismatch, tt.tolerance,
				"Expected mismatch ~%f, got %f", tt.expectedMismatch, mismatch)
		})
	}
}

func TestHallucinationDetector_HybridVerification(t *testing.T) {
	detector := NewHallucinationDetector()

	thought := &types.Thought{
		ID:         "hybrid-test",
		Content:    "PostgreSQL uses MVCC. This provides good isolation. Maybe it's the best database.",
		Confidence: 0.75,
	}

	ctx := context.Background()
	report, err := detector.VerifyThought(ctx, thought)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, thought.ID, report.ThoughtID)

	// Fast verification should complete
	assert.NotEmpty(t, report.Claims)
	assert.NotNil(t, report.SemanticUncertainty)

	// Should have recommendations
	assert.NotEmpty(t, report.Recommendations)

	// Report should be cached
	cachedReport, ok := detector.getCachedReport(thought.ID)
	assert.True(t, ok)
	assert.Equal(t, report.ThoughtID, cachedReport.ThoughtID)
}

func TestHallucinationDetector_Caching(t *testing.T) {
	detector := NewHallucinationDetector()

	report := &HallucinationReport{
		ThoughtID:  "cache-test",
		OverallRisk: 0.5,
		AnalyzedAt:  time.Now(),
	}

	// Cache the report
	detector.cacheReport(report)

	// Retrieve from cache
	cached, ok := detector.getCachedReport("cache-test")
	assert.True(t, ok)
	assert.Equal(t, report.ThoughtID, cached.ThoughtID)
	assert.Equal(t, report.OverallRisk, cached.OverallRisk)

	// Try non-existent key
	_, ok = detector.getCachedReport("non-existent")
	assert.False(t, ok)
}

// MockKnowledgeSource for testing
type MockKnowledgeSource struct {
	verifyFunc func(ctx context.Context, claim string) (*VerificationResult, error)
}

func (m *MockKnowledgeSource) Verify(ctx context.Context, claim string) (*VerificationResult, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(ctx, claim)
	}
	return &VerificationResult{
		IsVerified: true,
		Confidence: 0.9,
		Source:     "mock",
	}, nil
}

func (m *MockKnowledgeSource) Type() string {
	return "mock"
}

func (m *MockKnowledgeSource) Confidence() float64 {
	return 0.9
}

func TestHallucinationDetector_WithKnowledgeSource(t *testing.T) {
	detector := NewHallucinationDetector()

	// Register mock knowledge source
	mockSource := &MockKnowledgeSource{
		verifyFunc: func(ctx context.Context, claim string) (*VerificationResult, error) {
			if claim == "True fact" {
				return &VerificationResult{
					IsVerified:         true,
					Confidence:         0.95,
					SupportingEvidence: []string{"verified source"},
					Source:             "mock",
				}, nil
			}
			return &VerificationResult{
				IsVerified:            false,
				Confidence:            0.8,
				ContradictingEvidence: []string{"contradicts known facts"},
				Source:                "mock",
			}, nil
		},
	}

	detector.RegisterKnowledgeSource(mockSource)

	thought := &types.Thought{
		ID:      "knowledge-test",
		Content: "True fact. False claim.",
	}

	report := detector.deepVerification(thought)

	assert.NotNil(t, report)
	assert.Equal(t, VerificationDeep, report.VerificationLevel)
	assert.Equal(t, 1, report.VerifiedCount, "Should have 1 verified claim")
	assert.Equal(t, 1, report.HallucinationCount, "Should have 1 hallucination")
}
