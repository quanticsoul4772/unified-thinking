package analysis

import (
	"strings"
	"testing"
)

func TestNewPerspectiveAnalyzer(t *testing.T) {
	pa := NewPerspectiveAnalyzer()
	if pa == nil {
		t.Fatal("NewPerspectiveAnalyzer() returned nil")
	}
}

func TestAnalyzePerspectives(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	tests := []struct {
		name             string
		situation        string
		stakeholderHints []string
		expectError      bool
		minPerspectives  int
	}{
		{
			name:            "empty situation",
			situation:       "",
			expectError:     true,
			minPerspectives: 0,
		},
		{
			name:             "with stakeholder hints",
			situation:        "We need to decide whether to implement a new feature that will improve user experience but increase server costs",
			stakeholderHints: []string{"users", "management", "engineers"},
			expectError:      false,
			minPerspectives:  3,
		},
		{
			name:            "auto-detect stakeholders",
			situation:       "The company plans to implement a new policy affecting employees and customers",
			expectError:     false,
			minPerspectives: 2, // Should detect employees and customers
		},
		{
			name:            "no obvious stakeholders",
			situation:       "This is a generic decision without specific stakeholders mentioned",
			expectError:     false,
			minPerspectives: 1, // Should use default stakeholders
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perspectives, err := pa.AnalyzePerspectives(tt.situation, tt.stakeholderHints)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("AnalyzePerspectives() failed: %v", err)
			}

			if len(perspectives) < tt.minPerspectives {
				t.Errorf("Expected at least %d perspectives, got %d", tt.minPerspectives, len(perspectives))
			}

			// Verify perspective structure
			for _, p := range perspectives {
				if p.ID == "" {
					t.Error("Perspective ID is empty")
				}
				if p.Stakeholder == "" {
					t.Error("Stakeholder is empty")
				}
				if p.Viewpoint == "" {
					t.Error("Viewpoint is empty")
				}
				if len(p.Concerns) == 0 {
					t.Error("Concerns are empty")
				}
				if len(p.Priorities) == 0 {
					t.Error("Priorities are empty")
				}
				if len(p.Constraints) == 0 {
					t.Error("Constraints are empty")
				}
				if p.Confidence < 0 || p.Confidence > 1 {
					t.Errorf("Confidence %f is out of range [0,1]", p.Confidence)
				}
				if p.CreatedAt.IsZero() {
					t.Error("CreatedAt is zero")
				}
			}
		})
	}
}

func TestDetectStakeholders(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	tests := []struct {
		name                string
		situation           string
		expectedStakeholder string // At least one expected
	}{
		{
			name:                "user mentioned",
			situation:           "The user interface needs improvement",
			expectedStakeholder: "users",
		},
		{
			name:                "employee mentioned",
			situation:           "This policy will affect all employees",
			expectedStakeholder: "employees",
		},
		{
			name:                "management mentioned",
			situation:           "The CEO and executive team must decide",
			expectedStakeholder: "management",
		},
		{
			name:                "investor mentioned",
			situation:           "Shareholders are concerned about returns",
			expectedStakeholder: "investors",
		},
		{
			name:                "community mentioned",
			situation:           "The public has concerns about this proposal",
			expectedStakeholder: "community",
		},
		{
			name:                "no specific stakeholders",
			situation:           "This is a generic situation",
			expectedStakeholder: "decision-maker", // Should use defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stakeholders := pa.detectStakeholders(tt.situation)

			if len(stakeholders) == 0 {
				t.Error("No stakeholders detected")
			}

			found := false
			for _, s := range stakeholders {
				if s == tt.expectedStakeholder {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected stakeholder '%s' not found in %v", tt.expectedStakeholder, stakeholders)
			}
		})
	}
}

func TestExtractConcerns(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	tests := []struct {
		name        string
		situation   string
		stakeholder string
		minConcerns int
	}{
		{
			name:        "user stakeholder",
			situation:   "We need to improve usability and privacy",
			stakeholder: "users",
			minConcerns: 2,
		},
		{
			name:        "employee stakeholder",
			situation:   "This will affect workload and job security",
			stakeholder: "employees",
			minConcerns: 2,
		},
		{
			name:        "management stakeholder",
			situation:   "We need to consider profitability and risk",
			stakeholder: "management",
			minConcerns: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			concerns := pa.extractConcerns(tt.situation, tt.stakeholder)

			if len(concerns) < tt.minConcerns {
				t.Errorf("Expected at least %d concerns, got %d", tt.minConcerns, len(concerns))
			}
		})
	}
}

func TestExtractPriorities(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	stakeholders := []string{"users", "employees", "management", "investors", "community"}

	for _, stakeholder := range stakeholders {
		t.Run(stakeholder, func(t *testing.T) {
			priorities := pa.extractPriorities(stakeholder)

			if len(priorities) == 0 {
				t.Errorf("No priorities extracted for %s", stakeholder)
			}
		})
	}
}

func TestExtractConstraints(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	stakeholders := []string{"users", "employees", "management", "investors", "community"}

	for _, stakeholder := range stakeholders {
		t.Run(stakeholder, func(t *testing.T) {
			constraints := pa.extractConstraints(stakeholder)

			if len(constraints) == 0 {
				t.Errorf("No constraints extracted for %s", stakeholder)
			}
		})
	}
}

func TestEstimateConfidence(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	tests := []struct {
		name        string
		stakeholder string
		situation   string
		minConf     float64
		maxConf     float64
	}{
		{
			name:        "well-defined stakeholder",
			stakeholder: "users",
			situation:   "The users are concerned about privacy",
			minConf:     0.8,
			maxConf:     1.0,
		},
		{
			name:        "well-defined without mention",
			stakeholder: "employees",
			situation:   "This is a generic situation",
			minConf:     0.7,
			maxConf:     0.9,
		},
		{
			name:        "generic stakeholder",
			stakeholder: "random-group",
			situation:   "Generic situation",
			minConf:     0.5,
			maxConf:     0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := pa.estimateConfidence(tt.stakeholder, tt.situation)

			if confidence < tt.minConf || confidence > tt.maxConf {
				t.Errorf("Confidence %f not in expected range [%f, %f]", confidence, tt.minConf, tt.maxConf)
			}
		})
	}
}

func TestDetectPerspectiveConflicts(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	// Create conflicting perspectives
	situation := "Should we prioritize speed or thoroughness?"
	perspectives, err := pa.AnalyzePerspectives(situation, []string{"management", "quality-assurance"})
	if err != nil {
		t.Fatalf("AnalyzePerspectives() failed: %v", err)
	}

	conflicts := pa.detectPerspectiveConflicts(perspectives)

	// Note: Conflicts might not always be detected depending on the situation
	// Just verify the method runs without error
	if conflicts == nil {
		t.Error("detectPerspectiveConflicts() returned nil")
	}
}

func TestComparePerspectives(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	situation := "We need to balance user needs with business profitability"
	perspectives, err := pa.AnalyzePerspectives(situation, []string{"users", "management", "investors"})
	if err != nil {
		t.Fatalf("AnalyzePerspectives() failed: %v", err)
	}

	comparison, err := pa.ComparePerspectives(perspectives)
	if err != nil {
		t.Fatalf("ComparePerspectives() failed: %v", err)
	}

	// Verify comparison structure
	if _, ok := comparison["common_concerns"]; !ok {
		t.Error("Comparison missing 'common_concerns'")
	}
	if _, ok := comparison["common_priorities"]; !ok {
		t.Error("Comparison missing 'common_priorities'")
	}
	if _, ok := comparison["conflicts"]; !ok {
		t.Error("Comparison missing 'conflicts'")
	}
	if _, ok := comparison["synthesis"]; !ok {
		t.Error("Comparison missing 'synthesis'")
	}

	// Verify synthesis is a string
	if synthesis, ok := comparison["synthesis"].(string); !ok || synthesis == "" {
		t.Error("Synthesis is not a non-empty string")
	}
}

func TestComparePerspectives_InsufficientPerspectives(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	situation := "Test situation"
	perspectives, _ := pa.AnalyzePerspectives(situation, []string{"users"})

	_, err := pa.ComparePerspectives(perspectives)
	if err == nil {
		t.Error("Expected error for insufficient perspectives, got nil")
	}
}

func TestPrioritiesConflict(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	tests := []struct {
		name           string
		priorities1    []string
		priorities2    []string
		expectConflict bool
	}{
		{
			name:           "speed vs thoroughness",
			priorities1:    []string{"speed", "quick delivery"},
			priorities2:    []string{"thoroughness", "careful analysis"},
			expectConflict: true,
		},
		{
			name:           "cost vs quality",
			priorities1:    []string{"minimize cost", "budget savings"},
			priorities2:    []string{"high quality", "premium standards"},
			expectConflict: true,
		},
		{
			name:           "no conflict",
			priorities1:    []string{"customer satisfaction", "quality service"},
			priorities2:    []string{"user happiness", "excellent experience"},
			expectConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflict := pa.prioritiesConflict(tt.priorities1, tt.priorities2)

			if conflict != tt.expectConflict {
				t.Errorf("prioritiesConflict() = %v, want %v", conflict, tt.expectConflict)
			}
		})
	}
}

func TestConcernsConflict(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	tests := []struct {
		name           string
		concerns1      []string
		concerns2      []string
		expectConflict bool
	}{
		{
			name:           "privacy vs transparency",
			concerns1:      []string{"data privacy", "confidentiality"},
			concerns2:      []string{"full transparency", "open disclosure"},
			expectConflict: true,
		},
		{
			name:           "security vs accessibility",
			concerns1:      []string{"maximum security", "strict controls"},
			concerns2:      []string{"easy accessibility", "minimal barriers"},
			expectConflict: true,
		},
		{
			name:           "no conflict",
			concerns1:      []string{"user safety", "system reliability"},
			concerns2:      []string{"data accuracy", "performance"},
			expectConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflict := pa.concernsConflict(tt.concerns1, tt.concerns2)

			if conflict != tt.expectConflict {
				t.Errorf("concernsConflict() = %v, want %v", conflict, tt.expectConflict)
			}
		})
	}
}

func TestConcurrentPerspectiveGeneration(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	situation := "A complex decision affecting multiple stakeholders"
	stakeholders := []string{"users", "employees", "management", "investors", "community"}

	// Generate perspectives concurrently
	done := make(chan bool)
	errors := make(chan error, len(stakeholders))

	for _, stakeholder := range stakeholders {
		go func(sh string) {
			_, err := pa.AnalyzePerspectives(situation, []string{sh})
			if err != nil {
				errors <- err
			}
			done <- true
		}(stakeholder)
	}

	// Wait for all to complete
	for i := 0; i < len(stakeholders); i++ {
		<-done
	}

	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent generation error: %v", err)
	}
}

// TestViewpointsAreDifferent verifies that different stakeholders produce different viewpoints
// This is a critical test to ensure we don't have identical template responses
func TestViewpointsAreDifferent(t *testing.T) {
	pa := NewPerspectiveAnalyzer()

	situation := "Should we fund dark matter research?"
	stakeholders := []string{"scientist", "policymaker", "taxpayer"}

	perspectives, err := pa.AnalyzePerspectives(situation, stakeholders)
	if err != nil {
		t.Fatalf("AnalyzePerspectives() failed: %v", err)
	}

	if len(perspectives) != len(stakeholders) {
		t.Fatalf("Expected %d perspectives, got %d", len(stakeholders), len(perspectives))
	}

	// Collect all viewpoints
	viewpoints := make(map[string]string)
	for _, p := range perspectives {
		viewpoints[p.Stakeholder] = p.Viewpoint
	}

	// Verify each pair of stakeholders has DIFFERENT viewpoints
	for i := 0; i < len(stakeholders); i++ {
		for j := i + 1; j < len(stakeholders); j++ {
			s1 := stakeholders[i]
			s2 := stakeholders[j]
			v1 := viewpoints[s1]
			v2 := viewpoints[s2]

			if v1 == v2 {
				t.Errorf("CRITICAL: %s and %s have IDENTICAL viewpoints!\n"+
					"This defeats the purpose of perspective analysis.\n"+
					"Viewpoint: %q", s1, s2, v1)
			}
		}
	}

	// Additional check: verify viewpoints contain stakeholder-specific language
	scientistViewpoint := viewpoints["scientist"]
	if !containsAny(scientistViewpoint, "scientist", "empirical", "evidence", "methodology", "data") {
		t.Errorf("Scientist viewpoint lacks scientific language: %q", scientistViewpoint)
	}

	policymakerViewpoint := viewpoints["policymaker"]
	if !containsAny(policymakerViewpoint, "policy", "stakeholder", "societal", "interests") {
		t.Errorf("Policymaker viewpoint lacks policy language: %q", policymakerViewpoint)
	}

	taxpayerViewpoint := viewpoints["taxpayer"]
	if !containsAny(taxpayerViewpoint, "taxpayer", "public", "fund", "money", "priorities") {
		t.Errorf("Taxpayer viewpoint lacks taxpayer language: %q", taxpayerViewpoint)
	}
}

// containsAny checks if text contains any of the given substrings (case-insensitive)
func containsAny(text string, substrings ...string) bool {
	textLower := strings.ToLower(text)
	for _, s := range substrings {
		if strings.Contains(textLower, strings.ToLower(s)) {
			return true
		}
	}
	return false
}

func TestSynthesizeViewpointVariety(t *testing.T) {
	pa := NewPerspectiveAnalyzer()
	concerns := []string{"resource allocation", "scientific value"}
	situation := "Research funding decision"

	// Test that different stakeholder types produce different viewpoints
	stakeholderTypes := []string{
		"scientist",
		"policymaker",
		"taxpayer",
		"investor",
		"engineer",
		"philosopher",
	}

	viewpoints := make(map[string]string)
	for _, stakeholder := range stakeholderTypes {
		viewpoint := pa.synthesizeViewpoint(situation, stakeholder, concerns)
		viewpoints[stakeholder] = viewpoint
	}

	// Verify all viewpoints are unique
	seen := make(map[string]string)
	for stakeholder, viewpoint := range viewpoints {
		if prevStakeholder, exists := seen[viewpoint]; exists {
			t.Errorf("Duplicate viewpoint! %s and %s have identical viewpoints:\n%s",
				prevStakeholder, stakeholder, viewpoint)
		}
		seen[viewpoint] = stakeholder
	}
}

func TestGenericViewpointGeneration(t *testing.T) {
	pa := NewPerspectiveAnalyzer()
	concerns := []string{"impact", "feasibility"}
	situation := "Test situation"

	// Test unknown stakeholder types get varied generic viewpoints
	unknownStakeholder := "alien-observer"
	viewpoint := pa.generateGenericViewpoint(situation, unknownStakeholder, concerns)

	if viewpoint == "" {
		t.Error("generateGenericViewpoint returned empty string")
	}

	// Should include the stakeholder name for variation
	if !strings.Contains(viewpoint, unknownStakeholder) {
		t.Errorf("Generic viewpoint should include stakeholder name %q, got: %q",
			unknownStakeholder, viewpoint)
	}

	// Should include concerns
	for _, concern := range concerns {
		if !strings.Contains(viewpoint, concern) {
			t.Errorf("Generic viewpoint should include concern %q, got: %q",
				concern, viewpoint)
		}
	}
}
