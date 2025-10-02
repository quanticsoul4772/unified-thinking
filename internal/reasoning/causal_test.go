package reasoning

import (
	"testing"
)

func TestNewCausalReasoner(t *testing.T) {
	cr := NewCausalReasoner()
	if cr == nil {
		t.Fatal("NewCausalReasoner() returned nil")
	}
	if cr.graphs == nil {
		t.Fatal("graphs map not initialized")
	}
}

func TestBuildCausalGraph(t *testing.T) {
	cr := NewCausalReasoner()

	tests := []struct {
		name         string
		description  string
		observations []string
		expectError  bool
		minVariables int
		minLinks     int
	}{
		{
			name:        "empty description",
			description: "",
			observations: []string{"A causes B"},
			expectError: true,
		},
		{
			name:        "no observations",
			description: "Test",
			observations: []string{},
			expectError: true,
		},
		{
			name:        "simple causal chain",
			description: "Temperature affects ice cream sales",
			observations: []string{
				"Temperature increases ice cream sales",
				"Ice cream sales increase revenue",
			},
			expectError:  false,
			minVariables: 2,
			minLinks:     1,
		},
		{
			name:        "multiple causal relationships",
			description: "Factors affecting plant growth",
			observations: []string{
				"Sunlight increases plant growth",
				"Water increases plant growth",
				"Fertilizer enhances plant growth",
			},
			expectError:  false,
			minVariables: 2,
			minLinks:     1,
		},
		{
			name:        "negative causation",
			description: "Pollution effects",
			observations: []string{
				"Pollution decreases air quality",
				"Air quality affects health",
			},
			expectError:  false,
			minVariables: 2,
			minLinks:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := cr.BuildCausalGraph(tt.description, tt.observations)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("BuildCausalGraph() failed: %v", err)
			}

			if graph.ID == "" {
				t.Error("Graph ID is empty")
			}
			if graph.Description != tt.description {
				t.Errorf("Expected description %q, got %q", tt.description, graph.Description)
			}
			if len(graph.Variables) < tt.minVariables {
				t.Errorf("Expected at least %d variables, got %d", tt.minVariables, len(graph.Variables))
			}
			if len(graph.Links) < tt.minLinks {
				t.Errorf("Expected at least %d links, got %d", tt.minLinks, len(graph.Links))
			}
			if graph.CreatedAt.IsZero() {
				t.Error("CreatedAt is zero")
			}

			// Verify variables have required fields
			for _, v := range graph.Variables {
				if v.ID == "" {
					t.Error("Variable ID is empty")
				}
				if v.Name == "" {
					t.Error("Variable name is empty")
				}
				if v.Type == "" {
					t.Error("Variable type is empty")
				}
			}

			// Verify links have required fields
			for _, link := range graph.Links {
				if link.ID == "" {
					t.Error("Link ID is empty")
				}
				if link.From == "" {
					t.Error("Link From is empty")
				}
				if link.To == "" {
					t.Error("Link To is empty")
				}
				if link.Strength < 0 || link.Strength > 1 {
					t.Errorf("Link strength %f out of range [0,1]", link.Strength)
				}
				if link.Confidence < 0 || link.Confidence > 1 {
					t.Errorf("Link confidence %f out of range [0,1]", link.Confidence)
				}
			}
		})
	}
}

func TestSimulateIntervention(t *testing.T) {
	cr := NewCausalReasoner()

	// Build a graph first
	graph, err := cr.BuildCausalGraph(
		"Temperature and sales",
		[]string{
			"Temperature increases ice cream sales",
			"Ice cream sales increase revenue",
		},
	)
	if err != nil {
		t.Fatalf("BuildCausalGraph() failed: %v", err)
	}

	tests := []struct {
		name             string
		graphID          string
		variableID       string
		interventionType string
		expectError      bool
		minEffects       int
	}{
		{
			name:             "invalid graph ID",
			graphID:          "nonexistent",
			variableID:       "temperature",
			interventionType: "increase",
			expectError:      true,
		},
		{
			name:             "invalid variable ID",
			graphID:          graph.ID,
			variableID:       "nonexistent",
			interventionType: "increase",
			expectError:      true,
		},
		{
			name:             "valid intervention",
			graphID:          graph.ID,
			variableID:       graph.Variables[0].ID,
			interventionType: "increase",
			expectError:      false,
			minEffects:       0, // May or may not have downstream effects
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intervention, err := cr.SimulateIntervention(tt.graphID, tt.variableID, tt.interventionType)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("SimulateIntervention() failed: %v", err)
			}

			if intervention.ID == "" {
				t.Error("Intervention ID is empty")
			}
			if intervention.GraphID != tt.graphID {
				t.Errorf("Expected GraphID %q, got %q", tt.graphID, intervention.GraphID)
			}
			if intervention.InterventionType != tt.interventionType {
				t.Errorf("Expected intervention type %q, got %q", tt.interventionType, intervention.InterventionType)
			}
			if intervention.Confidence < 0 || intervention.Confidence > 1 {
				t.Errorf("Confidence %f out of range [0,1]", intervention.Confidence)
			}
			if len(intervention.PredictedEffects) < tt.minEffects {
				t.Errorf("Expected at least %d effects, got %d", tt.minEffects, len(intervention.PredictedEffects))
			}

			// Verify predicted effects
			for _, effect := range intervention.PredictedEffects {
				if effect.Variable == "" {
					t.Error("Effect variable is empty")
				}
				if effect.Effect == "" {
					t.Error("Effect direction is empty")
				}
				if effect.Probability < 0 || effect.Probability > 1 {
					t.Errorf("Effect probability %f out of range [0,1]", effect.Probability)
				}
				if effect.PathLength < 1 {
					t.Errorf("Path length %d should be >= 1", effect.PathLength)
				}
			}
		})
	}
}

func TestGenerateCounterfactual(t *testing.T) {
	cr := NewCausalReasoner()

	// Build a graph first
	graph, err := cr.BuildCausalGraph(
		"Weather and activities",
		[]string{
			"Rain affects outdoor activities",
			"Outdoor activities affect mood",
		},
	)
	if err != nil {
		t.Fatalf("BuildCausalGraph() failed: %v", err)
	}

	tests := []struct {
		name        string
		graphID     string
		scenario    string
		changes     map[string]string
		expectError bool
	}{
		{
			name:        "invalid graph ID",
			graphID:     "nonexistent",
			scenario:    "What if it didn't rain?",
			changes:     map[string]string{"rain": "no"},
			expectError: true,
		},
		{
			name:        "empty scenario",
			graphID:     graph.ID,
			scenario:    "",
			changes:     map[string]string{"rain": "no"},
			expectError: true,
		},
		{
			name:        "no changes",
			graphID:     graph.ID,
			scenario:    "What if nothing changed?",
			changes:     map[string]string{},
			expectError: true,
		},
		{
			name:        "valid counterfactual",
			graphID:     graph.ID,
			scenario:    "What if it didn't rain?",
			changes:     map[string]string{"rain": "decrease"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counterfactual, err := cr.GenerateCounterfactual(tt.graphID, tt.scenario, tt.changes)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("GenerateCounterfactual() failed: %v", err)
			}

			if counterfactual.ID == "" {
				t.Error("Counterfactual ID is empty")
			}
			if counterfactual.GraphID != tt.graphID {
				t.Errorf("Expected GraphID %q, got %q", tt.graphID, counterfactual.GraphID)
			}
			if counterfactual.Scenario != tt.scenario {
				t.Errorf("Expected scenario %q, got %q", tt.scenario, counterfactual.Scenario)
			}
			if counterfactual.Plausibility < 0 || counterfactual.Plausibility > 1 {
				t.Errorf("Plausibility %f out of range [0,1]", counterfactual.Plausibility)
			}
			if len(counterfactual.Changes) != len(tt.changes) {
				t.Errorf("Expected %d changes, got %d", len(tt.changes), len(counterfactual.Changes))
			}
		})
	}
}

func TestAnalyzeCorrelationVsCausation(t *testing.T) {
	cr := NewCausalReasoner()

	tests := []struct {
		name            string
		observation     string
		expectError     bool
		expectedKeyword string
	}{
		{
			name:        "empty observation",
			observation: "",
			expectError: true,
		},
		{
			name:            "causal language",
			observation:     "Smoking causes lung cancer",
			expectError:     false,
			expectedKeyword: "causal",
		},
		{
			name:            "correlation language",
			observation:     "Ice cream sales are correlated with drowning incidents",
			expectError:     false,
			expectedKeyword: "correlation",
		},
		{
			name:            "unclear relationship",
			observation:     "There is a relationship between education and income",
			expectError:     false,
			expectedKeyword: "unclear",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := cr.AnalyzeCorrelationVsCausation(tt.observation)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("AnalyzeCorrelationVsCausation() failed: %v", err)
			}

			if analysis == "" {
				t.Error("Analysis is empty")
			}

			if tt.expectedKeyword != "" {
				analysisLower := toLower(analysis)
				if !contains(analysisLower, tt.expectedKeyword) {
					t.Errorf("Expected analysis to contain %q, got: %q", tt.expectedKeyword, analysis)
				}
			}
		})
	}
}

func TestGetGraph(t *testing.T) {
	cr := NewCausalReasoner()

	// Build a graph
	graph, err := cr.BuildCausalGraph("Test", []string{"A causes B"})
	if err != nil {
		t.Fatalf("BuildCausalGraph() failed: %v", err)
	}

	// Test retrieval
	retrieved, err := cr.GetGraph(graph.ID)
	if err != nil {
		t.Fatalf("GetGraph() failed: %v", err)
	}

	if retrieved.ID != graph.ID {
		t.Errorf("Expected graph ID %q, got %q", graph.ID, retrieved.ID)
	}

	// Test nonexistent graph
	_, err = cr.GetGraph("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent graph, got nil")
	}
}

func TestConcurrentCausalReasoning(t *testing.T) {
	cr := NewCausalReasoner()

	observations := [][]string{
		{"A causes B", "B causes C"},
		{"X increases Y", "Y affects Z"},
		{"Temperature leads to expansion", "Expansion causes stress"},
	}

	done := make(chan bool)
	errors := make(chan error, len(observations))

	for i, obs := range observations {
		go func(idx int, obsSet []string) {
			_, err := cr.BuildCausalGraph("Concurrent test", obsSet)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i, obs)
	}

	// Wait for all to complete
	for i := 0; i < len(observations); i++ {
		<-done
	}

	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent graph building error: %v", err)
	}
}

// Helper functions
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
