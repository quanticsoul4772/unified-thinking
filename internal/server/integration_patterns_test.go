package server

import (
	"testing"
)

func TestGetIntegrationPatterns(t *testing.T) {
	patterns := GetIntegrationPatterns()

	if len(patterns) == 0 {
		t.Fatal("expected at least one integration pattern")
	}

	// Verify all patterns have required fields
	for i, pattern := range patterns {
		if pattern.Name == "" {
			t.Errorf("pattern %d: Name should not be empty", i)
		}
		if pattern.Description == "" {
			t.Errorf("pattern %d (%s): Description should not be empty", i, pattern.Name)
		}
		if len(pattern.Steps) == 0 {
			t.Errorf("pattern %d (%s): Steps should not be empty", i, pattern.Name)
		}
		if pattern.UseCase == "" {
			t.Errorf("pattern %d (%s): UseCase should not be empty", i, pattern.Name)
		}
		if len(pattern.Servers) == 0 {
			t.Errorf("pattern %d (%s): Servers should not be empty", i, pattern.Name)
		}
	}
}

func TestGetIntegrationPatterns_KnownPatterns(t *testing.T) {
	patterns := GetIntegrationPatterns()

	// Check for expected pattern names
	expectedPatterns := []string{
		"Research-Enhanced Thinking",
		"Knowledge-Backed Decision Making",
		"Causal Model to Knowledge Graph",
		"Problem Decomposition Workflow",
		"Temporal Decision Analysis",
		"Stakeholder-Aware Planning",
		"Validated File Operations",
		"Evidence-Based Causal Reasoning",
		"Iterative Problem Refinement",
		"Knowledge Discovery Pipeline",
	}

	patternMap := make(map[string]bool)
	for _, p := range patterns {
		patternMap[p.Name] = true
	}

	for _, expected := range expectedPatterns {
		if !patternMap[expected] {
			t.Errorf("expected pattern %q not found", expected)
		}
	}
}

func TestGetIntegrationPatterns_UnifiedThinkingIncluded(t *testing.T) {
	patterns := GetIntegrationPatterns()

	// All patterns should include unified-thinking server
	for _, pattern := range patterns {
		hasUnifiedThinking := false
		for _, server := range pattern.Servers {
			if server == "unified-thinking" {
				hasUnifiedThinking = true
				break
			}
		}
		if !hasUnifiedThinking {
			t.Errorf("pattern %q should include unified-thinking server", pattern.Name)
		}
	}
}

func TestGetIntegrationPatterns_StepsHaveNumbers(t *testing.T) {
	patterns := GetIntegrationPatterns()

	for _, pattern := range patterns {
		for i, step := range pattern.Steps {
			// Each step should start with a number
			if len(step) == 0 || step[0] < '1' || step[0] > '9' {
				t.Errorf("pattern %q step %d should start with a number: %s", pattern.Name, i, step)
			}
		}
	}
}

func TestIntegrationPattern_FieldsNotNil(t *testing.T) {
	patterns := GetIntegrationPatterns()

	for _, pattern := range patterns {
		// Steps and Servers are slices, verify they're not nil
		if pattern.Steps == nil {
			t.Errorf("pattern %q: Steps should not be nil", pattern.Name)
		}
		if pattern.Servers == nil {
			t.Errorf("pattern %q: Servers should not be nil", pattern.Name)
		}
	}
}
