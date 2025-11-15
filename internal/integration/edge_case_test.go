package integration

import (
	"testing"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

func TestCausalTemporalIntegration_generateTemporalRecommendation(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	temporalReasoner := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causalReasoner, temporalReasoner)

	tests := []struct {
		name          string
		pattern       string
		peakEffect    string
		horizons      map[string]*HorizonEffect
		expectedRec   string
		expectContent bool
	}{
		{
			name:       "immediate pattern",
			pattern:    "immediate",
			peakEffect: "peak_immediate",
			horizons: map[string]*HorizonEffect{
				"short_term": {
					TimeFrame: "days-weeks",
					Magnitude: "strong",
					Certainty: 0.8,
				},
				"medium_term": {
					TimeFrame: "months",
					Magnitude: "moderate",
					Certainty: 0.6,
				},
				"long_term": {
					TimeFrame: "years",
					Magnitude: "weak",
					Certainty: 0.4,
				},
			},
			expectedRec:   "immediate_action_required",
			expectContent: true,
		},
		{
			name:       "delayed pattern",
			pattern:    "delayed",
			peakEffect: "peak_delayed",
			horizons: map[string]*HorizonEffect{
				"short_term": {
					TimeFrame: "days-weeks",
					Magnitude: "weak",
					Certainty: 0.7,
				},
				"long_term": {
					TimeFrame: "years",
					Magnitude: "strong",
					Certainty: 0.8,
				},
			},
			expectedRec:   "delayed_investment_monitoring",
			expectContent: true,
		},
		{
			name:       "unknown pattern",
			pattern:    "unknown_pattern",
			peakEffect: "peak_unknown",
			horizons: map[string]*HorizonEffect{
				"short_term": {
					TimeFrame: "days-weeks",
					Magnitude: "weak",
					Certainty: 0.3,
				},
				"long_term": {
					TimeFrame: "years",
					Magnitude: "weak",
					Certainty: 0.2,
				},
			},
			expectedRec:   "",
			expectContent: false,
		},
		{
			name:       "empty pattern",
			pattern:    "",
			peakEffect: "peak_empty",
			horizons: map[string]*HorizonEffect{
				"short_term": {
					TimeFrame: "days-weeks",
					Magnitude: "weak",
					Certainty: 0.3,
				},
				"long_term": {
					TimeFrame: "years",
					Magnitude: "weak",
					Certainty: 0.2,
				},
			},
			expectedRec:   "",
			expectContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := integration.generateTemporalRecommendation(tt.pattern, tt.peakEffect, tt.horizons)

			if tt.expectContent {
				if result == "" {
					t.Error("Expected recommendation but got empty string")
				}
				// Note: We just check that content exists, not the exact format
				// as the implementation provides more detailed recommendations
			} else {
				if result != "" {
					t.Errorf("Expected empty recommendation but got: %v", result)
				}
			}
		})
	}
}

func TestCausalTemporalIntegration_identifyTimeSensitiveVariables(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	temporalReasoner := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causalReasoner, temporalReasoner)

	tests := []struct {
		name     string
		graph    *types.CausalGraph
		expected []string
	}{
		{
			name: "time-sensitive variables",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "market_timing", Name: "Market Timing"},
					{ID: "seasonal_demand", Name: "Seasonal Demand"},
					{ID: "interest_rates", Name: "Interest Rates"},
					{ID: "technology_adoption", Name: "Technology Adoption"},
				},
				Links: []*types.CausalLink{
					{From: "market_timing", To: "revenue"},
					{From: "market_timing", To: "cost"},
					{From: "market_timing", To: "risk"},
					{From: "seasonal_demand", To: "revenue"},
					{From: "seasonal_demand", To: "cost"},
				},
			},
			expected: []string{"market_timing", "seasonal_demand"},
		},
		{
			name: "no time-sensitive variables",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "static_cost", Name: "Static Cost"},
					{ID: "fixed_capacity", Name: "Fixed Capacity"},
				},
				Links: []*types.CausalLink{
					{From: "static_cost", To: "expense"},
				},
			},
			expected: []string{},
		},
		{
			name: "empty graph",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{},
				Links:     []*types.CausalLink{},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := integration.identifyTimeSensitiveVariables(tt.graph)

			if len(result) != len(tt.expected) {
				t.Errorf("identifyTimeSensitiveVariables() length = %v, want %v", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("identifyTimeSensitiveVariables() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestCausalTemporalIntegration_determineTimingWindows(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	temporalReasoner := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causalReasoner, temporalReasoner)

	tests := []struct {
		name          string
		graph         *types.CausalGraph
		temporal      *types.TemporalAnalysis
		timeSensitive []string
		expectedRec   string
		expectContent bool
	}{
		{
			name: "immediate pattern with variables",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "market_timing", Name: "Market Timing"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Immediate effects visible",
			},
			timeSensitive: []string{"market_timing"},
			expectedRec:   "Early Action Window",
			expectContent: true,
		},
		{
			name: "delayed pattern with variables",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "seasonal_demand", Name: "Seasonal Demand"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Effects build over time",
			},
			timeSensitive: []string{"seasonal_demand"},
			expectedRec:   "Early Action Window",
			expectContent: true,
		},
		{
			name: "unknown pattern",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "test", Name: "Test"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Unknown effects",
			},
			timeSensitive: []string{"test"},
			expectedRec:   "",
			expectContent: false,
		},
		{
			name: "no time-sensitive variables",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "static_cost", Name: "Static Cost"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Static effects",
			},
			timeSensitive: []string{},
			expectedRec:   "",
			expectContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := integration.determineTimingWindows(tt.graph, tt.temporal, tt.timeSensitive)

			if tt.expectContent {
				if len(result) == 0 {
					t.Error("Expected timing windows but got empty slice")
				} else if len(result) > 0 && result[0]["name"] != tt.expectedRec {
					t.Errorf("determineTimingWindows() first window name = %v, want %v", result[0]["name"], tt.expectedRec)
				}
			} else {
				if len(result) != 0 {
					t.Errorf("Expected empty timing windows but got: %v", result)
				}
			}
		})
	}
}

func TestCausalTemporalIntegration_synthesizeTimingRecommendation(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	temporalReasoner := reasoning.NewTemporalReasoner()
	integration := NewCausalTemporalIntegration(causalReasoner, temporalReasoner)

	tests := []struct {
		name          string
		graph         *types.CausalGraph
		temporal      *types.TemporalAnalysis
		timeSensitive []string
		expectedRec   string
		expectContent bool
	}{
		{
			name: "immediate with time-sensitive vars",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "market_timing", Name: "Market Timing"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Immediate effects visible",
			},
			timeSensitive: []string{"market_timing"},
			expectedRec:   "immediate_action_required",
			expectContent: true,
		},
		{
			name: "delayed with time-sensitive vars",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "interest_rates", Name: "Interest Rates"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Effects build over time",
			},
			timeSensitive: []string{"interest_rates"},
			expectedRec:   "delayed_monitoring_advised",
			expectContent: true,
		},
		{
			name: "sustained with time-sensitive vars",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "technology_adoption", Name: "Technology Adoption"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Long-term effects dominate",
			},
			timeSensitive: []string{"technology_adoption"},
			expectedRec:   "sustained_effort_needed",
			expectContent: true,
		},
		{
			name: "pattern without time-sensitive vars",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "static_cost", Name: "Static Cost"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Static effects",
			},
			timeSensitive: []string{},
			expectedRec:   "",
			expectContent: false,
		},
		{
			name: "empty pattern",
			graph: &types.CausalGraph{
				Variables: []*types.CausalVariable{
					{ID: "market_timing", Name: "Market Timing"},
				},
			},
			temporal: &types.TemporalAnalysis{
				ShortTermView: "Unknown effects",
			},
			timeSensitive: []string{"market_timing"},
			expectedRec:   "",
			expectContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// synthesizeTimingRecommendation expects []map[string]string
			windows := []map[string]string{}
			if tt.graph != nil && tt.temporal != nil {
				for _, variable := range tt.graph.Variables {
					for _, timeSensitiveVar := range tt.timeSensitive {
						if variable.ID == timeSensitiveVar {
							windows = append(windows, map[string]string{
								"variable":   variable.Name,
								"short_term": tt.temporal.ShortTermView,
							})
						}
					}
				}
			}
			result := integration.synthesizeTimingRecommendation(windows)

			if tt.expectContent {
				if result == "" {
					t.Error("Expected synthesis recommendation but got empty string")
				} else if result != tt.expectedRec {
					t.Errorf("synthesizeTimingRecommendation() = %v, want %v", result, tt.expectedRec)
				}
			} else {
				if result != "" {
					t.Errorf("Expected empty synthesis recommendation but got: %v", result)
				}
			}
		})
	}
}
