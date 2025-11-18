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

func TestAnalogicalReasoner_ListAnalogies(t *testing.T) {
	ar := NewAnalogicalReasoner()

	// Initially should be empty
	analogies := ar.ListAnalogies()
	if len(analogies) != 0 {
		t.Errorf("Expected empty list, got %d analogies", len(analogies))
	}

	// Create an analogy and verify it can be listed
	created, err := ar.FindAnalogy("Water flows through pipes with pressure", "Electricity moves through wires with voltage", nil)
	if err != nil {
		t.Fatalf("Failed to create analogy: %v", err)
	}

	// List should now have 1 analogy
	analogies = ar.ListAnalogies()
	if len(analogies) != 1 {
		t.Errorf("Expected 1 analogy, got %d", len(analogies))
	}

	// Verify the created analogy is in the list
	found := false
	for _, a := range analogies {
		if a.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created analogy not found in list")
	}

	// Verify each analogy has required fields
	for _, a := range analogies {
		if a.ID == "" {
			t.Error("Analogy ID should not be empty")
		}
		if a.SourceDomain == "" {
			t.Error("SourceDomain should not be empty")
		}
		if a.TargetDomain == "" {
			t.Error("TargetDomain should not be empty")
		}
	}
}

func TestAnalogicalReasoner_SemanticSimilarityEdgeCases(t *testing.T) {
	ar := NewAnalogicalReasoner()

	tests := []struct {
		name          string
		concept1      string
		concept2      string
		minSimilarity float64
		maxSimilarity float64
	}{
		{
			name:          "exact match",
			concept1:      "system",
			concept2:      "system",
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "case insensitive exact match",
			concept1:      "System",
			concept2:      "system",
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "substring containment",
			concept1:      "network",
			concept2:      "networking",
			minSimilarity: 0.7,
			maxSimilarity: 0.8,
		},
		{
			name:          "semantic pair - evolution",
			concept1:      "evolution",
			concept2:      "evolve",
			minSimilarity: 0.8,
			maxSimilarity: 0.9,
		},
		{
			name:          "semantic pair - selection",
			concept1:      "selection",
			concept2:      "select",
			minSimilarity: 0.7,
			maxSimilarity: 1.0,
		},
		{
			name:          "semantic pair - adaptation",
			concept1:      "adaptation",
			concept2:      "adapting",
			minSimilarity: 0.8,
			maxSimilarity: 0.9,
		},
		{
			name:          "empty strings",
			concept1:      "",
			concept2:      "",
			minSimilarity: 0.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "one empty string",
			concept1:      "test",
			concept2:      "",
			minSimilarity: 0.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "no match",
			concept1:      "apple",
			concept2:      "banana",
			minSimilarity: 0.0,
			maxSimilarity: 0.1,
		},
		{
			name:          "word prefix match",
			concept1:      "testing framework",
			concept2:      "test",
			minSimilarity: 0.3,
			maxSimilarity: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := ar.semanticSimilarity(tt.concept1, tt.concept2)

			if sim < tt.minSimilarity || sim > tt.maxSimilarity {
				t.Errorf("Similarity %.3f not in expected range [%.3f, %.3f]",
					sim, tt.minSimilarity, tt.maxSimilarity)
			}
		})
	}
}

func TestAnalogicalReasoner_CalculateAnalogyStrength(t *testing.T) {
	ar := NewAnalogicalReasoner()

	tests := []struct {
		name        string
		mapping     map[string]string
		source      []string
		target      []string
		minStrength float64
		maxStrength float64
	}{
		{
			name: "full coverage",
			mapping: map[string]string{
				"a": "x",
				"b": "y",
				"c": "z",
			},
			source:      []string{"a", "b", "c"},
			target:      []string{"x", "y", "z"},
			minStrength: 0.9,
			maxStrength: 1.0,
		},
		{
			name: "partial coverage",
			mapping: map[string]string{
				"a": "x",
			},
			source:      []string{"a", "b", "c"},
			target:      []string{"x", "y", "z"},
			minStrength: 0.2,
			maxStrength: 0.5,
		},
		{
			name:        "empty mapping",
			mapping:     map[string]string{},
			source:      []string{"a", "b"},
			target:      []string{"x", "y"},
			minStrength: 0.0,
			maxStrength: 0.1,
		},
		{
			name:        "empty source",
			mapping:     map[string]string{},
			source:      []string{},
			target:      []string{"x"},
			minStrength: 0.0,
			maxStrength: 0.0,
		},
		{
			name: "high source coverage low target",
			mapping: map[string]string{
				"a": "x",
				"b": "y",
			},
			source:      []string{"a", "b"},
			target:      []string{"x", "y", "z", "w"},
			minStrength: 0.5,
			maxStrength: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := ar.calculateAnalogyStrength(tt.mapping, tt.source, tt.target)

			if strength < tt.minStrength || strength > tt.maxStrength {
				t.Errorf("Strength %.3f not in expected range [%.3f, %.3f]",
					strength, tt.minStrength, tt.maxStrength)
			}
		})
	}
}

func TestAnalogicalReasoner_GenerateInsight(t *testing.T) {
	ar := NewAnalogicalReasoner()

	tests := []struct {
		name          string
		source        string
		target        string
		mapping       map[string]string
		expectContent bool
	}{
		{
			name:   "with mappings",
			source: "water system",
			target: "electrical circuit",
			mapping: map[string]string{
				"pressure": "voltage",
				"flow":     "current",
			},
			expectContent: true,
		},
		{
			name:          "empty mapping",
			source:        "source domain",
			target:        "target domain",
			mapping:       map[string]string{},
			expectContent: true,
		},
		{
			name:   "many mappings",
			source: "complex source",
			target: "complex target",
			mapping: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
				"d": "4",
				"e": "5",
				"f": "6",
			},
			expectContent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insight := ar.generateInsight(tt.source, tt.target, tt.mapping)

			if tt.expectContent && insight == "" {
				t.Error("Expected non-empty insight")
			}

			if len(tt.mapping) == 0 && insight == "" {
				t.Error("Expected default message for empty mapping")
			}

			// Check that insight mentions source and target for non-empty mappings
			if len(tt.mapping) > 0 {
				if !contains(insight, tt.source) {
					t.Errorf("Insight should mention source domain %q", tt.source)
				}
				if !contains(insight, tt.target) {
					t.Errorf("Insight should mention target domain %q", tt.target)
				}
			}
		})
	}
}

func TestAnalogicalReasoner_ApplyAnalogyNotFound(t *testing.T) {
	ar := NewAnalogicalReasoner()

	// Try to apply non-existent analogy
	_, err := ar.ApplyAnalogy("nonexistent-id", "some context")
	if err == nil {
		t.Error("Expected error for non-existent analogy but got none")
	}
}

func TestAnalogicalReasoner_BuildMapping(t *testing.T) {
	ar := NewAnalogicalReasoner()

	tests := []struct {
		name        string
		source      []string
		target      []string
		constraints []string
		minMappings int
	}{
		{
			name:        "direct matches",
			source:      []string{"system", "process", "network"},
			target:      []string{"system", "process", "network"},
			constraints: []string{},
			minMappings: 3,
		},
		{
			name:        "with constraints",
			source:      []string{"pressure", "flow"},
			target:      []string{"voltage", "current"},
			constraints: []string{"pressure->voltage", "flow->current"},
			minMappings: 2,
		},
		{
			name:        "no matches",
			source:      []string{"apple", "banana"},
			target:      []string{"car", "house"},
			constraints: []string{},
			minMappings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := ar.buildMapping(tt.source, tt.target, tt.constraints)

			if len(mapping) < tt.minMappings {
				t.Errorf("Expected at least %d mappings, got %d", tt.minMappings, len(mapping))
			}
		})
	}
}

func TestAnalogicalReasoner_ApplyConstraint(t *testing.T) {
	ar := NewAnalogicalReasoner()

	mapping := make(map[string]string)

	// Apply valid constraint
	ar.applyConstraint(mapping, "source->target")
	if mapping["source"] != "target" {
		t.Errorf("Expected mapping 'source'->'target', got %q", mapping["source"])
	}

	// Apply constraint with spaces
	ar.applyConstraint(mapping, "  key  ->  value  ")
	if mapping["key"] != "value" {
		t.Errorf("Expected mapping 'key'->'value', got %q", mapping["key"])
	}

	// Invalid constraint (no arrow)
	initialLen := len(mapping)
	ar.applyConstraint(mapping, "invalid constraint")
	if len(mapping) != initialLen {
		t.Error("Invalid constraint should not add to mapping")
	}
}

func TestAnalogicalReasoner_TransferInsights(t *testing.T) {
	ar := NewAnalogicalReasoner()

	// Create analogy with known mapping
	analogy, err := ar.FindAnalogy(
		"Water pressure drives flow through resistance",
		"Voltage drives current through resistance",
		[]string{"pressure->voltage", "flow->current"},
	)
	if err != nil {
		t.Fatalf("Failed to create analogy: %v", err)
	}

	// Transfer to context containing mapped concepts
	targetConcepts := []string{"voltage", "current", "power"}
	insights := ar.transferInsights(analogy, targetConcepts)

	if len(insights) == 0 {
		t.Error("Expected at least one transferred insight")
	}
}

func TestAnalogicalReasoner_IdentifyAdaptations(t *testing.T) {
	ar := NewAnalogicalReasoner()

	// Create analogy
	analogy, err := ar.FindAnalogy(
		"Simple source",
		"Simple target",
		[]string{"a->x"},
	)
	if err != nil {
		t.Fatalf("Failed to create analogy: %v", err)
	}

	// Identify adaptations for unmapped concepts
	targetConcepts := []string{"x", "y", "z"} // y and z are not mapped
	adaptations := ar.identifyAdaptations(analogy, targetConcepts)

	// Should identify y and z as needing adaptation
	if len(adaptations) < 1 {
		t.Error("Expected at least one adaptation recommendation")
	}
}

func TestAnalogicalReasoner_ContainsHelper(t *testing.T) {
	ar := NewAnalogicalReasoner()

	slice := []string{"apple", "Banana", "CHERRY"}

	tests := []struct {
		item   string
		expect bool
	}{
		{"apple", true},
		{"APPLE", true}, // Case insensitive
		{"banana", true},
		{"cherry", true},
		{"grape", false},
	}

	for _, tt := range tests {
		result := ar.contains(slice, tt.item)
		if result != tt.expect {
			t.Errorf("contains(%v, %q) = %v, expected %v", slice, tt.item, result, tt.expect)
		}
	}
}

func TestAnalogicalReasoner_IsValueInMap(t *testing.T) {
	ar := NewAnalogicalReasoner()

	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	// Value exists
	key, exists := ar.isValueInMap(m, "value1")
	if !exists {
		t.Error("Expected to find value1")
	}
	if key != "key1" {
		t.Errorf("Expected key 'key1', got %q", key)
	}

	// Value does not exist
	key, exists = ar.isValueInMap(m, "value3")
	if exists {
		t.Error("Did not expect to find value3")
	}
	if key != "" {
		t.Errorf("Expected empty key, got %q", key)
	}
}

// Helper function for tests
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
