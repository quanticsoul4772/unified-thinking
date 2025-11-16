package reasoning

import (
	"testing"
)

func TestAnalogicalReasoner_FindAnalogy(t *testing.T) {
	ar := NewAnalogicalReasoner()

	tests := []struct {
		name          string
		sourceDomain  string
		targetProblem string
		constraints   []string
		expectError   bool
		minStrength   float64
	}{
		{
			name:          "water flow to electrical current",
			sourceDomain:  "Water flow through pipes with pressure and resistance",
			targetProblem: "Electrical current through wires with voltage and resistance",
			constraints:   []string{"pressure->voltage", "flow->current"},
			expectError:   false,
			minStrength:   0.3,
		},
		{
			name:          "biological evolution to idea evolution",
			sourceDomain:  "Species evolve through natural selection and adaptation",
			targetProblem: "Ideas evolve through testing and refinement in the marketplace",
			constraints:   []string{},
			expectError:   false,
			minStrength:   0.2,
		},
		{
			name:          "empty source",
			sourceDomain:  "",
			targetProblem: "Some problem",
			constraints:   []string{},
			expectError:   true,
			minStrength:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analogy, err := ar.FindAnalogy(tt.sourceDomain, tt.targetProblem, tt.constraints)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if analogy == nil {
				t.Errorf("Expected analogy but got nil")
				return
			}

			if analogy.Strength < tt.minStrength {
				t.Errorf("Analogy strength %.3f is below minimum %.3f", analogy.Strength, tt.minStrength)
			}

			if analogy.SourceDomain != tt.sourceDomain {
				t.Errorf("Source domain mismatch: got %q, want %q", analogy.SourceDomain, tt.sourceDomain)
			}

			if len(analogy.Mapping) == 0 && len(tt.constraints) > 0 {
				t.Errorf("Expected mapping from constraints but got empty mapping")
			}
		})
	}
}

func TestAnalogicalReasoner_ApplyAnalogy(t *testing.T) {
	ar := NewAnalogicalReasoner()

	// Create an analogy first
	analogy, err := ar.FindAnalogy(
		"Water pressure drives flow through pipes",
		"Voltage drives current through wires",
		[]string{"pressure->voltage"},
	)
	if err != nil {
		t.Fatalf("Failed to create analogy: %v", err)
	}

	// Apply it to a new context
	result, err := ar.ApplyAnalogy(analogy.ID, "Battery voltage affects LED brightness")
	if err != nil {
		t.Errorf("Failed to apply analogy: %v", err)
	}

	if result == nil {
		t.Errorf("Expected result but got nil")
	}

	// Check result structure
	if _, ok := result["analogy_id"]; !ok {
		t.Errorf("Result missing analogy_id")
	}

	if _, ok := result["transferred_insights"]; !ok {
		t.Errorf("Result missing transferred_insights")
	}
}

func TestAnalogicalReasoner_GetAnalogy(t *testing.T) {
	ar := NewAnalogicalReasoner()

	// Create analogy
	created, err := ar.FindAnalogy("Source", "Target", nil)
	if err != nil {
		t.Fatalf("Failed to create analogy: %v", err)
	}

	// Retrieve it
	retrieved, err := ar.GetAnalogy(created.ID)
	if err != nil {
		t.Errorf("Failed to retrieve analogy: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Retrieved analogy ID mismatch: got %s, want %s", retrieved.ID, created.ID)
	}

	// Try non-existent ID
	_, err = ar.GetAnalogy("nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent analogy but got none")
	}
}

func TestAnalogicalReasoner_ExtractConcepts(t *testing.T) {
	ar := NewAnalogicalReasoner()

	tests := []struct {
		name           string
		text           string
		expectedMinLen int
	}{
		{
			name:           "technical domain",
			text:           "The system processes input through a complex network structure",
			expectedMinLen: 2,
		},
		{
			name:           "quoted concepts",
			text:           "The \"data flow\" moves through the \"processing pipeline\"",
			expectedMinLen: 2,
		},
		{
			name:           "empty text",
			text:           "",
			expectedMinLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			concepts := ar.extractConcepts(tt.text)

			if len(concepts) < tt.expectedMinLen {
				t.Errorf("Expected at least %d concepts, got %d", tt.expectedMinLen, len(concepts))
			}
		})
	}
}

func TestAnalogicalReasoner_SemanticSimilarity(t *testing.T) {
	ar := NewAnalogicalReasoner()

	tests := []struct {
		concept1      string
		concept2      string
		minSimilarity float64
	}{
		{"flow", "current", 0.0},             // No direct overlap
		{"water flow", "water current", 0.3}, // Partial overlap
		{"system process", "process", 0.4},   // Prefix match
	}

	for _, tt := range tests {
		sim := ar.semanticSimilarity(tt.concept1, tt.concept2)
		if sim < tt.minSimilarity {
			t.Errorf("Similarity between %q and %q is %.3f, expected at least %.3f",
				tt.concept1, tt.concept2, sim, tt.minSimilarity)
		}
	}
}
