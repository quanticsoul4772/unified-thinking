package validation

import (
	"testing"
)

// TestLogicalReasoningAccuracy tests the fundamental correctness of logical reasoning
// These tests expose the critical flaw: string matching instead of formal logic
func TestLogicalReasoningAccuracy(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name       string
		premises   []string
		conclusion string
		shouldProve bool
		description string
	}{
		// MODUS PONENS - Should be 100% accurate
		{
			name:       "Modus Ponens - Basic",
			premises:   []string{"If it rains, the ground is wet", "It rains"},
			conclusion: "The ground is wet",
			shouldProve: true,
			description: "Classic modus ponens inference",
		},
		{
			name:       "Modus Ponens - Formal",
			premises:   []string{"P implies Q", "P"},
			conclusion: "Q",
			shouldProve: true,
			description: "Formal logic notation",
		},

		// MODUS TOLLENS - Should be 100% accurate
		{
			name:       "Modus Tollens - Basic",
			premises:   []string{"If it rains, the ground is wet", "The ground is not wet"},
			conclusion: "It does not rain",
			shouldProve: true,
			description: "Classic modus tollens inference",
		},

		// INVALID INFERENCES - Should detect as invalid
		{
			name:       "Affirming Consequent - Invalid",
			premises:   []string{"If it rains, the ground is wet", "The ground is wet"},
			conclusion: "It rains",
			shouldProve: false,
			description: "Common fallacy - should NOT prove",
		},
		{
			name:       "Denying Antecedent - Invalid",
			premises:   []string{"If it rains, the ground is wet", "It does not rain"},
			conclusion: "The ground is not wet",
			shouldProve: false,
			description: "Common fallacy - should NOT prove",
		},

		// SYLLOGISMS
		{
			name:       "Universal Instantiation",
			premises:   []string{"All humans are mortal", "Socrates is human"},
			conclusion: "Socrates is mortal",
			shouldProve: true,
			description: "Classic syllogism",
		},
		{
			name:       "Invalid Syllogism - Undistributed Middle",
			premises:   []string{"All cats are mammals", "All dogs are mammals"},
			conclusion: "All cats are dogs",
			shouldProve: false,
			description: "Invalid syllogism - should detect",
		},

		// COMPLEX CASES
		{
			name:       "Hypothetical Syllogism",
			premises:   []string{"If P then Q", "If Q then R"},
			conclusion: "If P then R",
			shouldProve: true,
			description: "Chain of implications",
		},
		{
			name:       "Disjunctive Syllogism",
			premises:   []string{"P or Q", "Not P"},
			conclusion: "Q",
			shouldProve: true,
			description: "Elimination reasoning",
		},
	}

	passCount := 0
	failCount := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Prove(tt.premises, tt.conclusion)

			if result.IsProvable != tt.shouldProve {
				failCount++
				t.Errorf("%s FAILED\n  Expected: %v\n  Got: %v\n  Description: %s\n  Steps: %v",
					tt.name, tt.shouldProve, result.IsProvable, tt.description, result.Steps)
			} else {
				passCount++
				t.Logf("%s PASSED ✓", tt.name)
			}
		})
	}

	// Calculate accuracy
	total := passCount + failCount
	accuracy := float64(passCount) / float64(total) * 100.0

	t.Logf("\n=== LOGICAL REASONING ACCURACY ===")
	t.Logf("Pass: %d/%d (%.1f%%)", passCount, total, accuracy)
	t.Logf("Fail: %d/%d", failCount, total)
	t.Logf("TARGET: 95%% accuracy")

	if accuracy < 95.0 {
		t.Errorf("CRITICAL: Logical reasoning accuracy (%.1f%%) is below 95%% target", accuracy)
	}
}

// TestContradictionDetection tests the accuracy of contradiction detection
func TestContradictionDetection(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name        string
		content     string
		hasContradiction bool
		description string
	}{
		// TRUE CONTRADICTIONS
		{
			name:        "Direct Contradiction - Explicit",
			content:     "X is true and X is false",
			hasContradiction: true,
			description: "Explicit logical contradiction",
		},
		{
			name:        "Semantic Contradiction",
			content:     "The bachelor is married",
			hasContradiction: true,
			description: "Semantic contradiction (bachelor = unmarried man)",
		},
		{
			name:        "Numeric Contradiction",
			content:     "The temperature is above 100 degrees and below 50 degrees",
			hasContradiction: true,
			description: "Mathematical impossibility",
		},

		// FALSE POSITIVES TO AVOID
		{
			name:        "Color Contrast - Not Contradiction",
			content:     "The cat is black and the dog is not white",
			hasContradiction: false,
			description: "Different subjects - should NOT flag as contradiction",
		},
		{
			name:        "Temporal Difference - Not Contradiction",
			content:     "It was raining yesterday and it is not raining today",
			hasContradiction: false,
			description: "Different time frames - not contradictory",
		},
		{
			name:        "Probabilistic Statement - Not Contradiction",
			content:     "It might rain and it might not rain",
			hasContradiction: false,
			description: "Uncertainty, not contradiction",
		},

		// SUBTLE CONTRADICTIONS
		{
			name:        "Transitive Contradiction",
			content:     "A is greater than B. B is greater than C. C is greater than A.",
			hasContradiction: true,
			description: "Transitive relation violation",
		},
		{
			name:        "Modal Contradiction",
			content:     "It is necessarily true that X and it is possibly false that X",
			hasContradiction: true,
			description: "Modal logic contradiction",
		},
	}

	truePositives := 0
	falsePositives := 0
	trueNegatives := 0
	falseNegatives := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use validate to check for contradictions
			result := validator.detectContradiction(tt.content)
			detected := result != ""

			if tt.hasContradiction && detected {
				truePositives++
				t.Logf("%s: TRUE POSITIVE ✓", tt.name)
			} else if !tt.hasContradiction && !detected {
				trueNegatives++
				t.Logf("%s: TRUE NEGATIVE ✓", tt.name)
			} else if !tt.hasContradiction && detected {
				falsePositives++
				t.Errorf("%s: FALSE POSITIVE ✗\n  Detected: %s\n  Description: %s",
					tt.name, result, tt.description)
			} else if tt.hasContradiction && !detected {
				falseNegatives++
				t.Errorf("%s: FALSE NEGATIVE ✗\n  Missed contradiction\n  Description: %s",
					tt.name, tt.description)
			}
		})
	}

	// Calculate metrics
	precision := float64(truePositives) / float64(truePositives + falsePositives)
	recall := float64(truePositives) / float64(truePositives + falseNegatives)
	f1 := 2 * (precision * recall) / (precision + recall)

	t.Logf("\n=== CONTRADICTION DETECTION METRICS ===")
	t.Logf("True Positives: %d", truePositives)
	t.Logf("False Positives: %d", falsePositives)
	t.Logf("True Negatives: %d", trueNegatives)
	t.Logf("False Negatives: %d", falseNegatives)
	t.Logf("Precision: %.2f (target: >0.90)", precision)
	t.Logf("Recall: %.2f (target: >0.85)", recall)
	t.Logf("F1 Score: %.2f (target: >0.87)", f1)

	if precision < 0.90 {
		t.Errorf("Precision (%.2f) below target (0.90)", precision)
	}
	if recall < 0.85 {
		t.Errorf("Recall (%.2f) below target (0.85)", recall)
	}
}

// TestFormalLogicLimitations documents current limitations for improvement tracking
func TestFormalLogicLimitations(t *testing.T) {
	validator := NewLogicValidator()

	knownLimitations := []struct {
		name        string
		premises    []string
		conclusion  string
		shouldWork  bool
		currentlyWorks bool
		improvement string
	}{
		{
			name:        "First-Order Logic - Universal Quantifiers",
			premises:    []string{"For all x, if x is human then x is mortal", "Socrates is human"},
			conclusion:  "Socrates is mortal",
			shouldWork:  true,
			currentlyWorks: false,
			improvement: "Need FOL parser and quantifier handling",
		},
		{
			name:        "Existential Quantifiers",
			premises:    []string{"There exists an x such that P(x)"},
			conclusion:  "Something has property P",
			shouldWork:  true,
			currentlyWorks: false,
			improvement: "Need existential quantifier support",
		},
		{
			name:        "Nested Quantifiers",
			premises:    []string{"For all x, there exists y such that x loves y"},
			conclusion:  "Everyone loves someone",
			shouldWork:  true,
			currentlyWorks: false,
			improvement: "Need nested quantifier handling",
		},
		{
			name:        "Modal Logic",
			premises:    []string{"It is necessary that P", "P implies Q"},
			conclusion:  "It is necessary that Q",
			shouldWork:  true,
			currentlyWorks: false,
			improvement: "Need modal logic operators (necessary, possible)",
		},
	}

	t.Log("\n=== KNOWN LIMITATIONS (For Improvement Tracking) ===")
	for _, lim := range knownLimitations {
		result := validator.Prove(lim.premises, lim.conclusion)
		works := result.IsProvable == lim.shouldWork

		status := "❌ NOT IMPLEMENTED"
		if works {
			status = "✓ WORKS"
		}

		t.Logf("%s: %s", lim.name, status)
		t.Logf("  Should work: %v", lim.shouldWork)
		t.Logf("  Currently works: %v", works)
		t.Logf("  Improvement needed: %s\n", lim.improvement)
	}
}
