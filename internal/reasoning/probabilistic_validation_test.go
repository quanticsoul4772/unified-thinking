package reasoning

import (
	"math"
	"testing"
)

// TestBayesianAccuracy validates the mathematical correctness of Bayesian updates
// CRITICAL: Current implementation has incorrect formula
func TestBayesianAccuracy(t *testing.T) {
	reasoner := NewProbabilisticReasoner()

	tests := []struct {
		name          string
		priorProb     float64
		likelihood    float64   // P(E|H)
		evidenceProb  float64   // P(E)
		likelihoodNot float64   // P(E|¬H) - needed for correct calculation
		expectedPost  float64   // Correct posterior P(H|E)
		description   string
	}{
		{
			name:         "Medical Test - Classic Bayes",
			priorProb:    0.01,  // 1% have disease
			likelihood:   0.99,  // 99% true positive rate
			likelihoodNot: 0.05, // 5% false positive rate
			evidenceProb: 0.05,  // Will be calculated: 0.99*0.01 + 0.05*0.99 = 0.0594
			expectedPost: 0.166, // Correct: (0.99*0.01) / 0.0594 ≈ 0.166
			description:  "Rare disease with accurate test - posterior should be ~16.6%",
		},
		{
			name:         "Fair Coin - No Update",
			priorProb:    0.5,
			likelihood:   0.5,   // Fair coin
			likelihoodNot: 0.5,
			evidenceProb: 0.5,
			expectedPost: 0.5,   // Should stay 0.5
			description:  "Uninformative evidence should not change belief",
		},
		{
			name:         "Strong Evidence",
			priorProb:    0.3,
			likelihood:   0.9,   // Strong positive evidence
			likelihoodNot: 0.1,  // Weak evidence if not H
			evidenceProb: 0.34,  // 0.9*0.3 + 0.1*0.7 = 0.34
			expectedPost: 0.794, // (0.9*0.3) / 0.34 ≈ 0.794
			description:  "Strong evidence should significantly increase belief",
		},
		{
			name:         "Monty Hall - Switch Door",
			priorProb:    1.0/3.0,  // Initially 1/3 chance car behind chosen door
			likelihood:   0.0,      // P(host shows goat | car behind chosen) = 0
			likelihoodNot: 0.5,     // P(host shows goat | car elsewhere) = 1/2
			evidenceProb: 1.0/3.0,  // P(host shows goat) = 1/3
			expectedPost: 0.0,      // Should realize switching is better
			description:  "Monty Hall problem - evidence changes probability",
		},
	}

	totalMAE := 0.0
	passCount := 0
	failCount := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create belief with prior
			belief, err := reasoner.CreateBelief(tt.description, tt.priorProb)
			if err != nil {
				t.Fatalf("Failed to create belief: %v", err)
			}

			// Calculate correct posterior using Bayes' theorem
			// P(H|E) = P(E|H)P(H) / [P(E|H)P(H) + P(E|¬H)P(¬H)]
			priorNot := 1.0 - tt.priorProb
			numerator := tt.likelihood * tt.priorProb
			denominator := tt.likelihood*tt.priorProb + tt.likelihoodNot*priorNot
			correctPosterior := numerator / denominator

			// Update belief using current implementation
			evidenceID := "test_evidence"
			updated, err := reasoner.UpdateBelief(belief.ID, evidenceID, tt.likelihood, tt.evidenceProb)
			if err != nil {
				t.Fatalf("Failed to update belief: %v", err)
			}

			// Calculate error
			mae := math.Abs(updated.Probability - correctPosterior)
			totalMAE += mae

			t.Logf("Prior: %.3f", tt.priorProb)
			t.Logf("P(E|H): %.3f, P(E|¬H): %.3f", tt.likelihood, tt.likelihoodNot)
			t.Logf("Correct Posterior: %.3f", correctPosterior)
			t.Logf("Computed Posterior: %.3f", updated.Probability)
			t.Logf("Absolute Error: %.3f", mae)

			// Tolerance: MAE should be < 0.01
			if mae < 0.01 {
				passCount++
				t.Logf("✓ PASS - Error within tolerance")
			} else {
				failCount++
				t.Errorf("✗ FAIL - Error %.3f exceeds tolerance 0.01", mae)
				t.Errorf("  Description: %s", tt.description)
				t.Errorf("  Expected: %.3f, Got: %.3f", correctPosterior, updated.Probability)
			}
		})
	}

	avgMAE := totalMAE / float64(len(tests))

	t.Logf("\n=== BAYESIAN INFERENCE ACCURACY ===")
	t.Logf("Tests Passed: %d/%d", passCount, len(tests))
	t.Logf("Tests Failed: %d/%d", failCount, len(tests))
	t.Logf("Average MAE: %.4f (target: <0.01)", avgMAE)

	if avgMAE >= 0.01 {
		t.Errorf("CRITICAL: Bayesian update MAE (%.4f) exceeds 0.01 threshold", avgMAE)
		t.Errorf("Current formula is mathematically incorrect!")
		t.Errorf("Missing P(E|¬H)P(¬H) term in denominator")
	}
}

// TestProbabilityCoherence validates that probability assignments follow axioms
func TestProbabilityCoherence(t *testing.T) {
	reasoner := NewProbabilisticReasoner()

	t.Run("Complementary Probabilities", func(t *testing.T) {
		// P(A) + P(¬A) = 1.0
		beliefA, _ := reasoner.CreateBelief("Event A", 0.7)

		// Test: If P(A) = 0.7, then P(¬A) should be 0.3
		// Current system: Does it track complementary probabilities?
		t.Logf("P(A) = %.2f", beliefA.Probability)
		t.Logf("P(¬A) should be %.2f", 1.0-beliefA.Probability)

		// Note: System doesn't explicitly track ¬A
		t.Log("⚠️ WARNING: System doesn't track complementary probabilities")
	})

	t.Run("Conjunction Bound", func(t *testing.T) {
		// P(A∧B) ≤ min(P(A), P(B))
		beliefA, _ := reasoner.CreateBelief("Event A", 0.7)
		beliefB, _ := reasoner.CreateBelief("Event B", 0.8)

		// Combine using AND
		combined, err := reasoner.CombineBeliefs([]string{beliefA.ID, beliefB.ID}, "and")
		if err != nil {
			t.Fatalf("Failed to combine beliefs: %v", err)
		}

		minProb := math.Min(0.7, 0.8)
		if combined > minProb {
			t.Errorf("Conjunction bound violated: P(A∧B)=%.2f > min(%.2f,%.2f)=%.2f",
				combined, 0.7, 0.8, minProb)
		} else {
			t.Logf("✓ Conjunction bound satisfied: %.2f ≤ %.2f", combined, minProb)
		}
	})

	t.Run("Disjunction Bound", func(t *testing.T) {
		// P(A∨B) ≥ max(P(A), P(B))
		beliefA, _ := reasoner.CreateBelief("Event A", 0.7)
		beliefB, _ := reasoner.CreateBelief("Event B", 0.8)

		// Combine using OR
		combined, err := reasoner.CombineBeliefs([]string{beliefA.ID, beliefB.ID}, "or")
		if err != nil {
			t.Fatalf("Failed to combine beliefs: %v", err)
		}

		maxProb := math.Max(0.7, 0.8)
		if combined < maxProb {
			t.Errorf("Disjunction bound violated: P(A∨B)=%.2f < max(%.2f,%.2f)=%.2f",
				combined, 0.7, 0.8, maxProb)
		} else {
			t.Logf("✓ Disjunction bound satisfied: %.2f ≥ %.2f", combined, maxProb)
		}
	})

	t.Run("Probability Range", func(t *testing.T) {
		// All probabilities must be in [0, 1]
		testCases := []float64{0.0, 0.5, 1.0}

		for _, p := range testCases {
			belief, err := reasoner.CreateBelief("Test", p)
			if err != nil {
				t.Errorf("Failed to create belief with valid probability %.2f: %v", p, err)
			}
			if belief.Probability < 0.0 || belief.Probability > 1.0 {
				t.Errorf("Probability %.2f out of range [0,1]", belief.Probability)
			}
		}

		// Should reject invalid probabilities
		invalidProbs := []float64{-0.1, 1.5, 2.0}
		for _, p := range invalidProbs {
			_, err := reasoner.CreateBelief("Test", p)
			if err == nil {
				t.Errorf("Should reject invalid probability %.2f", p)
			}
		}
	})
}

// TestConfidenceCalibration validates that confidence scores match actual accuracy
func TestConfidenceCalibration(t *testing.T) {
	// This test would require historical data
	// For now, document the requirement

	t.Log("=== CONFIDENCE CALIBRATION TEST ===")
	t.Log("Requirement: Track predictions and outcomes")
	t.Log("Metric: Expected Calibration Error (ECE) <0.05")
	t.Log("")
	t.Log("Implementation needed:")
	t.Log("1. Store prediction-outcome pairs")
	t.Log("2. Bin predictions by confidence level")
	t.Log("3. Calculate accuracy within each bin")
	t.Log("4. ECE = weighted average of |accuracy - confidence|")
	t.Log("")
	t.Log("⚠️ NOT CURRENTLY IMPLEMENTED")
}

// BenchmarkBayesianUpdate measures performance of Bayesian updates
func BenchmarkBayesianUpdate(b *testing.B) {
	reasoner := NewProbabilisticReasoner()
	belief, _ := reasoner.CreateBelief("Test", 0.5)
	evidenceID := "test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reasoner.UpdateBelief(belief.ID, evidenceID, 0.8, 0.6)
	}
}
