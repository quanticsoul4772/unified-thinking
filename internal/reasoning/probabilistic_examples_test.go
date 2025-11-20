package reasoning_test

import (
	"fmt"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

// Example_medicalTest demonstrates the base rate fallacy and why both P(E|H) and P(E|¬H)
// are required for correct Bayesian inference.
//
// This example shows the counterintuitive result that a highly accurate medical test
// (99% sensitivity, 95% specificity) only gives a 16.7% probability of disease when
// testing positive, due to the low base rate (1% prevalence).
func Example_medicalTest() {
	pr := reasoning.NewProbabilisticReasoner()

	// Scenario: Medical test for rare disease
	// - Disease prevalence: 1% (prior probability)
	// - Test sensitivity: 99% (P(positive|disease) = 0.99)
	// - Test specificity: 95% (P(negative|healthy) = 0.95)
	//   Therefore: P(positive|healthy) = 1 - 0.95 = 0.05

	// Create initial belief about patient having the disease
	belief, _ := pr.CreateBelief("Patient has disease", 0.01)
	fmt.Printf("Prior probability: %.1f%%\n", belief.Probability*100)

	// Patient tests positive
	updated, _ := pr.UpdateBeliefFull(
		belief.ID,
		"test-result-positive",
		0.99, // P(positive|disease) - test sensitivity
		0.05, // P(positive|healthy) - false positive rate (1 - specificity)
	)

	// The calculation:
	// P(disease|positive) = 0.99 × 0.01 / (0.99 × 0.01 + 0.05 × 0.99)
	//                     = 0.0099 / 0.0594
	//                     = 0.167 (16.7%)
	//
	// Intuition: Out of 10,000 people:
	//   - 100 have disease: 99 test positive (true positives)
	//   - 9,900 are healthy: 495 test positive (false positives)
	//   - Total positive tests: 594
	//   - Actual disease: 99/594 = 16.7%

	fmt.Printf("Posterior probability after positive test: %.1f%%\n", updated.Probability*100)
	fmt.Printf("\nThis is NOT 99%% (the test accuracy)!\n")
	fmt.Printf("The low base rate (1%%) dominates the calculation.\n")

	// Output:
	// Prior probability: 1.0%
	// Posterior probability after positive test: 16.7%
	//
	// This is NOT 99% (the test accuracy)!
	// The low base rate (1%) dominates the calculation.
}

// Example_sequentialUpdates shows how beliefs evolve with multiple pieces of evidence.
// Each update uses the previous posterior as the new prior.
func Example_sequentialUpdates() {
	pr := reasoning.NewProbabilisticReasoner()

	// Initial belief: Feature will be popular
	belief, _ := pr.CreateBelief("Feature will be popular", 0.5)
	fmt.Printf("Initial belief: %.2f\n", belief.Probability)

	// First evidence: Positive user feedback survey
	// Users who will use the feature are 80% likely to give positive feedback
	// Users who won't use it are only 20% likely to give positive feedback
	belief, _ = pr.UpdateBeliefFull(
		belief.ID,
		"user-feedback",
		0.8, // P(positive feedback | will be popular)
		0.2, // P(positive feedback | won't be popular)
	)
	fmt.Printf("After user feedback: %.2f\n", belief.Probability)

	// Second evidence: Successful beta test
	// Features that will be popular have 90% success in beta
	// Features that won't be popular have 30% success in beta
	belief, _ = pr.UpdateBeliefFull(
		belief.ID,
		"beta-test",
		0.9, // P(beta success | will be popular)
		0.3, // P(beta success | won't be popular)
	)
	fmt.Printf("After beta test: %.2f\n", belief.Probability)

	// Third evidence: Competitor analysis shows similar feature failed
	// If our feature will be popular, competitor failure is unlikely (0.2)
	// If our feature won't be popular, competitor failure is likely (0.8)
	belief, _ = pr.UpdateBeliefFull(
		belief.ID,
		"competitor-failure",
		0.2, // P(competitor fails | our feature popular)
		0.8, // P(competitor fails | our feature not popular)
	)
	fmt.Printf("After competitor analysis: %.2f\n", belief.Probability)

	// Output:
	// Initial belief: 0.50
	// After user feedback: 0.80
	// After beta test: 0.92
	// After competitor analysis: 0.75
}

// Example_uninformativeEvidence demonstrates what happens when evidence provides no information.
// This occurs when P(E|H) = P(E|¬H), meaning the evidence is equally likely regardless of
// whether the hypothesis is true or false.
func Example_uninformativeEvidence() {
	pr := reasoning.NewProbabilisticReasoner()

	belief, _ := pr.CreateBelief("It will rain tomorrow", 0.6)
	fmt.Printf("Prior: %.2f\n", belief.Probability)

	// Uninformative evidence: "The sky exists"
	// P(sky exists | will rain) = 1.0
	// P(sky exists | won't rain) = 1.0
	// This evidence is equally likely under both hypotheses, so it provides no information
	updated, _ := pr.UpdateBeliefFull(
		belief.ID,
		"sky-exists",
		1.0, // P(sky exists | will rain)
		1.0, // P(sky exists | won't rain)
	)

	fmt.Printf("Posterior: %.2f\n", updated.Probability)
	fmt.Printf("Changed: %v\n", updated.Probability != belief.Probability)
	fmt.Printf("Uninformative: %v\n", updated.Metadata["last_update_uninformative"])

	// Output:
	// Prior: 0.60
	// Posterior: 0.60
	// Changed: false
	// Uninformative: true
}

// Example_extremeLikelihoods shows behavior with certainty (0.0 or 1.0 likelihoods).
func Example_extremeLikelihoods() {
	pr := reasoning.NewProbabilisticReasoner()

	belief, _ := pr.CreateBelief("Hypothesis is true", 0.5)
	fmt.Printf("Prior: %.2f\n", belief.Probability)

	// Absolute certainty: If hypothesis is true, we ALWAYS see this evidence (1.0)
	//                     If hypothesis is false, we NEVER see this evidence (0.0)
	updated, _ := pr.UpdateBeliefFull(
		belief.ID,
		"definitive-evidence",
		1.0, // P(E|H) = 1.0 - always see evidence if true
		0.0, // P(E|¬H) = 0.0 - never see evidence if false
	)

	// Calculation: 1.0 × 0.5 / (1.0 × 0.5 + 0.0 × 0.5) = 0.5 / 0.5 = 1.0
	fmt.Printf("Posterior with definitive evidence: %.2f\n", updated.Probability)

	// Output:
	// Prior: 0.50
	// Posterior with definitive evidence: 1.00
}

// Example_updateBeliefWithEvidence shows the helper method that estimates likelihoods
// from evidence quality scores.
func Example_updateBeliefWithEvidence() {
	pr := reasoning.NewProbabilisticReasoner()

	belief, _ := pr.CreateBelief("Database optimization will improve performance", 0.6)
	fmt.Printf("Prior: %.2f\n", belief.Probability)

	// High-quality supporting evidence (e.g., benchmark results)
	evidence := &types.Evidence{
		ID:            "benchmark-results",
		Content:       "Benchmark shows 40% improvement with indexing",
		Source:        "Performance Testing",
		SupportsClaim: true,
		OverallScore:  0.85, // High quality evidence
	}

	// This automatically estimates:
	// P(E|H) = 0.5 + (0.85 × 0.4) = 0.84
	// P(E|¬H) = 0.5 - (0.85 × 0.3) = 0.245
	updated, _ := pr.UpdateBeliefWithEvidence(belief.ID, evidence)

	fmt.Printf("Posterior with high-quality evidence: %.2f\n", updated.Probability)

	// Output:
	// Prior: 0.60
	// Posterior with high-quality evidence: 0.84
}
