package integration

import (
	"testing"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

func TestNewCausalTemporalIntegration(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()

	integration := NewCausalTemporalIntegration(causal, temporal)
	if integration == nil {
		t.Fatal("expected integration instance")
	}

	if integration.causalReasoner != causal {
		t.Error("causal reasoner not set correctly")
	}

	if integration.temporalReasoner != temporal {
		t.Error("temporal reasoner not set correctly")
	}
}

func TestAnalyzeTemporalCausalEffects_ErrorWithInvalidGraph(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	// Try to analyze effects for a non-existent graph
	result, err := integration.AnalyzeTemporalCausalEffects("nonexistent-graph", "variable", "increase")
	if err == nil {
		t.Fatal("expected error for invalid graph")
	}

	if result != nil {
		t.Error("expected nil result when error occurs")
	}
}

func TestAnalyzeDecisionTiming_ErrorWithInvalidGraph(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	// Try to analyze timing for a non-existent graph
	result, err := integration.AnalyzeDecisionTiming("situation", "nonexistent-graph")
	if err == nil {
		t.Fatal("expected error for invalid graph")
	}

	if result != nil {
		t.Error("expected nil result when error occurs")
	}
}

func TestAnalyzeShortTermEffects(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	// Create a mock intervention with effects
	intervention := &types.CausalIntervention{
		ID: "test-intervention",
		PredictedEffects: []*types.PredictedEffect{
			{
				Variable:    "var1",
				Magnitude:   0.8,
				Probability: 0.9,
				PathLength:  1, // Direct effect
			},
			{
				Variable:    "var2",
				Magnitude:   0.6,
				Probability: 0.7,
				PathLength:  2, // Indirect effect
			},
		},
	}

	horizon := integration.analyzeShortTermEffects(intervention)
	if horizon == nil {
		t.Fatal("expected horizon effect")
	}

	if horizon.TimeFrame != "days-weeks" {
		t.Errorf("expected timeframe 'days-weeks', got %s", horizon.TimeFrame)
	}

	// Should have direct effect (path length 1)
	if len(horizon.DirectEffects) != 1 {
		t.Errorf("expected 1 direct effect, got %d", len(horizon.DirectEffects))
	}

	// Should have indirect effect (path length 2)
	if len(horizon.IndirectEffects) != 1 {
		t.Errorf("expected 1 indirect effect, got %d", len(horizon.IndirectEffects))
	}

	if horizon.Magnitude == "" {
		t.Error("expected magnitude to be calculated")
	}

	if horizon.Certainty <= 0 {
		t.Errorf("expected certainty > 0, got %f", horizon.Certainty)
	}
}

func TestAnalyzeMediumTermEffects(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	// Create a mock intervention with effects
	intervention := &types.CausalIntervention{
		ID: "test-intervention",
		PredictedEffects: []*types.PredictedEffect{
			{
				Variable:    "var1",
				Magnitude:   0.7,
				Probability: 0.8,
				PathLength:  2, // Medium term direct effect
			},
			{
				Variable:    "var2",
				Magnitude:   0.5,
				Probability: 0.6,
				PathLength:  3, // Medium term indirect effect
			},
		},
	}

	horizon := integration.analyzeMediumTermEffects(intervention)
	if horizon == nil {
		t.Fatal("expected horizon effect")
	}

	if horizon.TimeFrame != "months" {
		t.Errorf("expected timeframe 'months', got %s", horizon.TimeFrame)
	}

	if len(horizon.DirectEffects) != 1 {
		t.Errorf("expected 1 direct effect, got %d", len(horizon.DirectEffects))
	}

	if len(horizon.IndirectEffects) != 1 {
		t.Errorf("expected 1 indirect effect, got %d", len(horizon.IndirectEffects))
	}

	// Certainty should be reduced for medium term
	if horizon.Certainty >= 1.0 {
		t.Errorf("expected reduced certainty for medium term, got %f", horizon.Certainty)
	}
}

func TestAnalyzeLongTermEffects(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	// Create a mock intervention with effects
	intervention := &types.CausalIntervention{
		ID: "test-intervention",
		PredictedEffects: []*types.PredictedEffect{
			{
				Variable:    "var1",
				Magnitude:   0.6,
				Probability: 0.7,
				PathLength:  3, // Long term direct effect
			},
			{
				Variable:    "var2",
				Magnitude:   0.4,
				Probability: 0.5,
				PathLength:  4, // Long term indirect effect
			},
		},
	}

	horizon := integration.analyzeLongTermEffects(intervention)
	if horizon == nil {
		t.Fatal("expected horizon effect")
	}

	if horizon.TimeFrame != "years" {
		t.Errorf("expected timeframe 'years', got %s", horizon.TimeFrame)
	}

	if len(horizon.DirectEffects) != 2 {
		t.Errorf("expected 2 direct effects (path length 3+), got %d", len(horizon.DirectEffects))
	}

	// Certainty should be further reduced for long term
	if horizon.Certainty >= 1.0 {
		t.Errorf("expected further reduced certainty for long term, got %f", horizon.Certainty)
	}
}

func TestDetermineTemporalPattern(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	// Test increasing pattern
	horizons := map[string]*HorizonEffect{
		"short_term": {
			Magnitude: "weak",
		},
		"medium_term": {
			Magnitude: "moderate",
		},
		"long_term": {
			Magnitude: "strong",
		},
	}

	pattern := integration.determineTemporalPattern(horizons)
	if pattern != "increasing" {
		t.Errorf("expected 'increasing' pattern, got %s", pattern)
	}

	// Test decreasing pattern
	horizons["short_term"].Magnitude = "strong"
	horizons["medium_term"].Magnitude = "moderate"
	horizons["long_term"].Magnitude = "weak"

	pattern = integration.determineTemporalPattern(horizons)
	if pattern != "decreasing" {
		t.Errorf("expected 'decreasing' pattern, got %s", pattern)
	}

	// Test oscillating pattern
	horizons["short_term"].Magnitude = "strong"
	horizons["medium_term"].Magnitude = "weak"
	horizons["long_term"].Magnitude = "moderate"

	pattern = integration.determineTemporalPattern(horizons)
	if pattern != "oscillating" {
		t.Errorf("expected 'oscillating' pattern, got %s", pattern)
	}

	// Test stable pattern
	horizons["short_term"].Magnitude = "moderate"
	horizons["medium_term"].Magnitude = "moderate"
	horizons["long_term"].Magnitude = "moderate"

	pattern = integration.determineTemporalPattern(horizons)
	if pattern != "stable" {
		t.Errorf("expected 'stable' pattern, got %s", pattern)
	}
}

func TestIdentifyPeakEffectTime(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	horizons := map[string]*HorizonEffect{
		"short_term": {
			Magnitude: "weak",
			Certainty: 0.3,
		},
		"medium_term": {
			Magnitude: "strong",
			Certainty: 0.8,
		},
		"long_term": {
			Magnitude: "moderate",
			Certainty: 0.5,
		},
	}

	peakTime := integration.identifyPeakEffectTime(horizons)
	if peakTime == "" {
		t.Error("expected peak time description")
	}

	// The medium term should be peak since it has highest magnitude * certainty
	if !containsString(peakTime, "months") {
		t.Errorf("expected peak time to reference months, got %s", peakTime)
	}
}

func TestCalculateMagnitude(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	// Test with no effects
	magnitude := integration.calculateMagnitude([]*types.PredictedEffect{})
	if magnitude != "weak" {
		t.Errorf("expected 'weak' for no effects, got %s", magnitude)
	}

	// Test with weak effects
	weakEffects := []*types.PredictedEffect{
		{Magnitude: 0.3, Probability: 0.4},
		{Magnitude: 0.2, Probability: 0.3},
	}

	magnitude = integration.calculateMagnitude(weakEffects)
	if magnitude != "weak" {
		t.Errorf("expected 'weak' for weak effects, got %s", magnitude)
	}

	// Test with moderate effects - need more than 30% to be strong
	modEffects := []*types.PredictedEffect{
		{Magnitude: 0.8, Probability: 0.9}, // Strong (both > thresholds)
		{Magnitude: 0.2, Probability: 0.3}, // Weak
		{Magnitude: 0.2, Probability: 0.3}, // Weak
	}

	magnitude = integration.calculateMagnitude(modEffects)
	// 1 out of 3 = 33.3% > 30% but <= 60% -> moderate
	if magnitude != "moderate" {
		t.Errorf("expected 'moderate' for moderate effects, got %s", magnitude)
	}

	// Test with strong effects - need more than 60% to be strong
	strongEffects := []*types.PredictedEffect{
		{Magnitude: 0.8, Probability: 0.9}, // Strong
		{Magnitude: 0.8, Probability: 0.9}, // Strong
		{Magnitude: 0.8, Probability: 0.9}, // Strong
		{Magnitude: 0.2, Probability: 0.3}, // Weak
	}

	magnitude = integration.calculateMagnitude(strongEffects)
	// 3 out of 4 = 75% > 60% -> strong
	if magnitude != "strong" {
		t.Errorf("expected 'strong' for strong effects, got %s", magnitude)
	}
}

func TestCalculateCertainty(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	// Test with no effects
	certainty := integration.calculateCertainty([]*types.PredictedEffect{})
	if certainty != 0.0 {
		t.Errorf("expected 0.0 for no effects, got %f", certainty)
	}

	// Test with effects
	effects := []*types.PredictedEffect{
		{Probability: 0.6},
		{Probability: 0.8},
		{Probability: 0.4},
	}

	certainty = integration.calculateCertainty(effects)
	expected := (0.6 + 0.8 + 0.4) / 3.0
	if certainty != expected {
		t.Errorf("expected %f, got %f", expected, certainty)
	}
}

func TestMagnitudeValue(t *testing.T) {
	causal := reasoning.NewCausalReasoner()
	temporal := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causal, temporal)

	tests := []struct {
		magnitude string
		expected  float64
	}{
		{"strong", 1.0},
		{"moderate", 0.6},
		{"weak", 0.3},
		{"invalid", 0.0},
		{"", 0.0},
	}

	for _, test := range tests {
		got := integration.magnitudeValue(test.magnitude)
		if got != test.expected {
			t.Errorf("magnitudeValue(%q) = %f, want %f", test.magnitude, got, test.expected)
		}
	}
}

func TestJoinStrings(t *testing.T) {

	// Test empty slice
	result := joinStrings([]string{}, ", ")
	if result != "" {
		t.Errorf("expected empty string for empty slice, got %q", result)
	}

	// Test single string
	result = joinStrings([]string{"single"}, ", ")
	if result != "single" {
		t.Errorf("expected 'single', got %q", result)
	}

	// Test multiple strings
	result = joinStrings([]string{"first", "second", "third"}, ", ")
	if result != "first, second, third" {
		t.Errorf("expected 'first, second, third', got %q", result)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
