package reasoning_test

import (
	"fmt"
	"math"
	"testing"
	"testing/quick"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

// Property: Posterior probability always in [0, 1] regardless of valid inputs
func TestProperty_PosteriorInRange(t *testing.T) {
	pr := reasoning.NewProbabilisticReasoner()

	f := func(priorProb, likelihoodTrue, likelihoodFalse float64) bool {
		// Normalize inputs to valid ranges [0, 1]
		priorProb = clamp(priorProb, 0, 1)
		likelihoodTrue = clamp(likelihoodTrue, 0, 1)
		likelihoodFalse = clamp(likelihoodFalse, 0, 1)

		// Create belief
		belief, err := pr.CreateBelief("Property test", priorProb)
		if err != nil {
			return true // Skip invalid inputs (shouldn't happen after clamping)
		}

		// Update with evidence
		updated, err := pr.UpdateBeliefFull(belief.ID, "ev1", likelihoodTrue, likelihoodFalse)
		if err != nil {
			return true // Skip if update fails
		}

		// Property: posterior must be in [0, 1]
		if updated.Probability < 0 || updated.Probability > 1 {
			t.Logf("VIOLATION: posterior = %f (prior=%f, P(E|H)=%f, P(E|¬H)=%f)",
				updated.Probability, priorProb, likelihoodTrue, likelihoodFalse)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 1000, // Run 1000 random tests
	}

	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property: Sequential updates never produce probabilities outside [0, 1]
func TestProperty_SequentialUpdatesStayInRange(t *testing.T) {
	pr := reasoning.NewProbabilisticReasoner()

	f := func(priorProb float64, updates []struct{ LTrue, LFalse float64 }) bool {
		priorProb = clamp(priorProb, 0, 1)

		// Limit number of updates to avoid excessive test time
		if len(updates) > 10 {
			updates = updates[:10]
		}

		belief, err := pr.CreateBelief("Sequential property test", priorProb)
		if err != nil {
			return true
		}

		// Apply sequential updates
		for i, update := range updates {
			lTrue := clamp(update.LTrue, 0, 1)
			lFalse := clamp(update.LFalse, 0, 1)

			belief, err = pr.UpdateBeliefFull(belief.ID, fmt.Sprintf("ev%d", i), lTrue, lFalse)
			if err != nil {
				return true
			}

			// Check probability stays in range
			if belief.Probability < 0 || belief.Probability > 1 {
				t.Logf("VIOLATION at update %d: posterior = %f", i, belief.Probability)
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 500, // 500 random test cases
	}

	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property: Uninformative evidence (P(E|H) = P(E|¬H)) doesn't change posterior
func TestProperty_UninformativeEvidenceNoChange(t *testing.T) {
	pr := reasoning.NewProbabilisticReasoner()

	f := func(priorProb, likelihood float64) bool {
		priorProb = clamp(priorProb, 0, 1)
		likelihood = clamp(likelihood, 0, 1)

		belief, err := pr.CreateBelief("Uninformative test", priorProb)
		if err != nil {
			return true
		}

		// Update with uninformative evidence (same likelihood for both)
		updated, err := pr.UpdateBeliefFull(belief.ID, "ev-uninformative", likelihood, likelihood)
		if err != nil {
			return true
		}

		// Property: posterior should equal prior (within floating point tolerance)
		if math.Abs(updated.Probability-priorProb) > 1e-9 {
			t.Logf("VIOLATION: prior=%f, posterior=%f (uninformative evidence changed belief)",
				priorProb, updated.Probability)
			return false
		}

		// Should be marked as uninformative
		if uninformative, ok := updated.Metadata["last_update_uninformative"].(bool); !ok || !uninformative {
			t.Logf("VIOLATION: uninformative evidence not marked in metadata")
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 1000,
	}

	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property: Combining beliefs with AND always produces probability ≤ minimum input
func TestProperty_CombineAnd_NotGreaterThanMin(t *testing.T) {
	pr := reasoning.NewProbabilisticReasoner()

	f := func(prob1, prob2 float64) bool {
		prob1 = clamp(prob1, 0, 1)
		prob2 = clamp(prob2, 0, 1)

		belief1, _ := pr.CreateBelief("Belief 1", prob1)
		belief2, _ := pr.CreateBelief("Belief 2", prob2)

		combined, err := pr.CombineBeliefs([]string{belief1.ID, belief2.ID}, "and")
		if err != nil {
			return true
		}

		minProb := math.Min(prob1, prob2)

		// Property: P(A and B) ≤ min(P(A), P(B))
		if combined > minProb+1e-9 {
			t.Logf("VIOLATION: P(A and B) = %f > min(%f, %f) = %f",
				combined, prob1, prob2, minProb)
			return false
		}

		// Also check it equals P(A) * P(B) for independent events
		expected := prob1 * prob2
		if math.Abs(combined-expected) > 1e-9 {
			t.Logf("VIOLATION: P(A and B) = %f != P(A) * P(B) = %f",
				combined, expected)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 1000,
	}

	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property: Combining beliefs with OR always produces probability ≥ maximum input
func TestProperty_CombineOr_NotLessThanMax(t *testing.T) {
	pr := reasoning.NewProbabilisticReasoner()

	f := func(prob1, prob2 float64) bool {
		prob1 = clamp(prob1, 0, 1)
		prob2 = clamp(prob2, 0, 1)

		belief1, _ := pr.CreateBelief("Belief 1", prob1)
		belief2, _ := pr.CreateBelief("Belief 2", prob2)

		combined, err := pr.CombineBeliefs([]string{belief1.ID, belief2.ID}, "or")
		if err != nil {
			return true
		}

		maxProb := math.Max(prob1, prob2)

		// Property: P(A or B) ≥ max(P(A), P(B))
		if combined < maxProb-1e-9 {
			t.Logf("VIOLATION: P(A or B) = %f < max(%f, %f) = %f",
				combined, prob1, prob2, maxProb)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 1000,
	}

	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property: Likelihood estimator always produces valid likelihoods [0, 1]
func TestProperty_EstimatorProducesValidLikelihoods(t *testing.T) {
	estimator := reasoning.NewStandardEstimator(nil)

	f := func(score float64, supportsClaim bool) bool {
		score = clamp(score, 0, 1)

		evidence := &types.Evidence{
			ID:            "prop-test",
			SupportsClaim: supportsClaim,
			OverallScore:  score,
		}

		ifTrue, ifFalse, err := estimator.EstimateLikelihoods(evidence)
		if err != nil {
			t.Logf("Estimator error: %v (score=%f, supports=%v)", err, score, supportsClaim)
			return false
		}

		// Property: both likelihoods must be in [0, 1]
		if ifTrue < 0 || ifTrue > 1 {
			t.Logf("VIOLATION: P(E|H) = %f out of range", ifTrue)
			return false
		}

		if ifFalse < 0 || ifFalse > 1 {
			t.Logf("VIOLATION: P(E|¬H) = %f out of range", ifFalse)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 1000,
	}

	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property: Strong supporting evidence increases belief
func TestProperty_SupportingEvidenceIncreasesBelief(t *testing.T) {
	pr := reasoning.NewProbabilisticReasoner()

	f := func(priorProb, likelihoodTrue float64) bool {
		priorProb = clamp(priorProb, 0.01, 0.99) // Avoid extremes
		likelihoodTrue = clamp(likelihoodTrue, 0.6, 1.0)
		likelihoodFalse := clamp(1.0-likelihoodTrue, 0, 0.4)

		belief, _ := pr.CreateBelief("Property test", priorProb)
		updated, err := pr.UpdateBeliefFull(belief.ID, "ev-support", likelihoodTrue, likelihoodFalse)
		if err != nil {
			return true
		}

		// Property: if P(E|H) > P(E|¬H), posterior should increase
		if likelihoodTrue > likelihoodFalse && updated.Probability <= priorProb {
			t.Logf("VIOLATION: supporting evidence didn't increase belief (prior=%f, post=%f, P(E|H)=%f, P(E|¬H)=%f)",
				priorProb, updated.Probability, likelihoodTrue, likelihoodFalse)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 500,
	}

	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Helper function to clamp values to [min, max]
func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
