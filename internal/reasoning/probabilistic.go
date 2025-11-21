// Package reasoning provides advanced cognitive reasoning capabilities including
// probabilistic reasoning, Bayesian inference, and logical extensions.
package reasoning

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"unified-thinking/internal/metrics"
	"unified-thinking/internal/types"
)

// ProbabilisticReasoner performs Bayesian inference and probabilistic reasoning
type ProbabilisticReasoner struct {
	mu        sync.RWMutex
	beliefs   map[string]*types.ProbabilisticBelief
	counter   int
	metrics   *metrics.ProbabilisticMetrics
	estimator LikelihoodEstimator
}

// NewProbabilisticReasoner creates a new probabilistic reasoner with default settings
func NewProbabilisticReasoner() *ProbabilisticReasoner {
	return &ProbabilisticReasoner{
		beliefs:   make(map[string]*types.ProbabilisticBelief),
		metrics:   metrics.NewProbabilisticMetrics(),
		estimator: NewStandardEstimator(nil), // Use default profile
	}
}

// NewProbabilisticReasonerWithEstimator creates a reasoner with custom likelihood estimator
func NewProbabilisticReasonerWithEstimator(estimator LikelihoodEstimator) *ProbabilisticReasoner {
	if estimator == nil {
		estimator = NewStandardEstimator(nil)
	}
	return &ProbabilisticReasoner{
		beliefs:   make(map[string]*types.ProbabilisticBelief),
		metrics:   metrics.NewProbabilisticMetrics(),
		estimator: estimator,
	}
}

// CreateBelief creates a new probabilistic belief with prior probability
func (pr *ProbabilisticReasoner) CreateBelief(statement string, priorProb float64) (*types.ProbabilisticBelief, error) {
	if priorProb < 0 || priorProb > 1 {
		return nil, fmt.Errorf("probability must be between 0 and 1, got: %f", priorProb)
	}

	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.counter++
	belief := &types.ProbabilisticBelief{
		ID:          fmt.Sprintf("belief-%d", pr.counter),
		Statement:   statement,
		Probability: priorProb,
		PriorProb:   priorProb,
		Evidence:    []string{},
		UpdatedAt:   time.Now(),
		Metadata:    map[string]interface{}{},
	}

	pr.beliefs[belief.ID] = belief

	if pr.metrics != nil {
		pr.metrics.RecordBeliefCreated()
	}

	return belief, nil
}

// UpdateBelief applies Bayesian update based on new evidence
//
// DEPRECATED: This method is mathematically questionable and should not be used in new code.
//
// PROBLEMS WITH THIS API:
// 1. Assumes P(E|¬H) = 0.5 (neutral), which is rarely correct in real scenarios
// 2. The "evidenceProb" parameter is ignored because it conflates P(E) with likelihoods
// 3. Cannot properly apply Bayes' theorem without both P(E|H) and P(E|¬H)
// 4. Produces incorrect posteriors when evidence has asymmetric likelihoods
//
// MIGRATION PATH - Replace this:
//   pr.UpdateBelief(beliefID, evidenceID, 0.8, 0.5)
//
// With this:
//   pr.UpdateBeliefFull(beliefID, evidenceID, 0.8, 0.2)
//   // Explicitly specify: 0.8 = P(E|H), 0.2 = P(E|¬H)
//
// Or use UpdateBeliefWithEvidence if you have an Evidence struct that includes quality scores.
//
// DO NOT USE THIS METHOD IN NEW CODE. Use UpdateBeliefFull for correctness.
//
// Parameters:
//   - beliefID: ID of the belief to update
//   - evidenceID: ID of the evidence being applied
//   - likelihood: P(E|H) - probability of evidence given hypothesis is true
//   - evidenceProb: P(E) - DEPRECATED, this parameter is ignored. Method now always
//     calculates P(E) properly using Bayes' theorem with P(E|¬H) = 0.5
//
// Returns the updated belief or an error
func (pr *ProbabilisticReasoner) UpdateBelief(beliefID string, evidenceID string, likelihood, evidenceProb float64) (*types.ProbabilisticBelief, error) {
	// For backward compatibility, we use a default P(E|¬H) = 0.5
	// This is not ideal but maintains the existing API behavior
	// Users should migrate to UpdateBeliefFull for proper Bayesian inference

	// Calculate what P(E|¬H) would need to be if evidenceProb was correct
	// However, we'll just use the default for consistency
	likelihoodIfFalse := 0.5

	// Note: evidenceProb parameter is now ignored to avoid mathematical errors
	// Always use the full Bayesian update with the default P(E|¬H)
	return pr.UpdateBeliefFull(beliefID, evidenceID, likelihood, likelihoodIfFalse)
}

// UpdateBeliefFull applies Bayesian update with full parameters.
//
// MATHEMATICAL FOUNDATION:
// This is the mathematically correct implementation of Bayes' theorem:
//
//   P(H|E) = P(E|H) × P(H) / [P(E|H) × P(H) + P(E|¬H) × P(¬H)]
//
// Where:
//   - P(H|E) = Posterior probability (updated belief after seeing evidence)
//   - P(E|H) = Likelihood if true (probability of evidence if hypothesis is true)
//   - P(H) = Prior probability (belief before seeing evidence)
//   - P(E|¬H) = Likelihood if false (probability of evidence if hypothesis is false)
//   - P(¬H) = 1 - P(H) (probability hypothesis is false)
//
// WHY BOTH LIKELIHOODS ARE REQUIRED:
// Without both P(E|H) and P(E|¬H), we cannot properly normalize the posterior probability.
// This is a common source of the "base rate fallacy" - ignoring how common the evidence
// is under alternative hypotheses leads to dramatically incorrect probability estimates.
//
// CRITICAL EXAMPLE - Medical Testing (Base Rate Fallacy):
// A medical test has:
//   - 99% sensitivity: P(positive|disease) = 0.99
//   - 95% specificity: P(negative|healthy) = 0.95, so P(positive|healthy) = 0.05
//   - Disease prevalence: 1% (prior probability)
//
// Question: If a patient tests positive, what is P(disease|positive)?
//
// INTUITIVE (WRONG) ANSWER: 99% (the test sensitivity)
//
// CORRECT CALCULATION:
//   P(disease|positive) = 0.99 × 0.01 / [0.99 × 0.01 + 0.05 × 0.99]
//                       = 0.0099 / [0.0099 + 0.0495]
//                       = 0.0099 / 0.0594
//                       ≈ 0.167 (16.7%)
//
// The positive test only gives a 16.7% probability of disease, NOT 99%!
// This counterintuitive result occurs because the disease is rare (1% base rate).
// Out of 10,000 people: 100 have disease (99 test positive), 9,900 are healthy (495 test positive).
// Total positive tests: 594. Of these, only 99 actually have the disease (99/594 ≈ 16.7%).
//
// WITHOUT P(E|¬H) = 0.05, WE CANNOT COMPUTE THIS CORRECTLY!
//
// WARNING TO FUTURE DEVELOPERS:
// Do NOT "optimize" this back to a single likelihood parameter. The mathematical correctness
// depends on having BOTH conditional probabilities. Any attempt to simplify this API will
// reintroduce the base rate fallacy and produce incorrect probability updates.
//
// If you think this can be simplified, read the medical test example above carefully and
// try to compute the correct posterior with only one likelihood - you will find it's impossible.
//
// Parameters:
//   - beliefID: ID of the belief to update
//   - evidenceID: ID of the evidence being applied
//   - likelihoodIfTrue: P(E|H) - probability of seeing evidence if hypothesis is true
//   - likelihoodIfFalse: P(E|¬H) - probability of seeing evidence if hypothesis is false
//
// Returns the updated belief with the posterior probability
func (pr *ProbabilisticReasoner) UpdateBeliefFull(beliefID string, evidenceID string, likelihoodIfTrue, likelihoodIfFalse float64) (*types.ProbabilisticBelief, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	belief, exists := pr.beliefs[beliefID]
	if !exists {
		return nil, fmt.Errorf("belief not found: %s", beliefID)
	}

	// Validate likelihood parameters
	if likelihoodIfTrue < 0 || likelihoodIfTrue > 1 {
		if pr.metrics != nil {
			pr.metrics.RecordError()
		}
		return nil, fmt.Errorf("P(E|H) must be between 0 and 1, got: %f", likelihoodIfTrue)
	}
	if likelihoodIfFalse < 0 || likelihoodIfFalse > 1 {
		if pr.metrics != nil {
			pr.metrics.RecordError()
		}
		return nil, fmt.Errorf("P(E|¬H) must be between 0 and 1, got: %f", likelihoodIfFalse)
	}

	// Check for degenerate cases where both likelihoods are identical
	// This means the evidence provides no information
	if math.Abs(likelihoodIfTrue-likelihoodIfFalse) < 1e-10 {
		// Evidence is equally likely regardless of hypothesis truth
		// No update needed - posterior equals prior

		// Log warning - this may indicate upstream issues with evidence quality
		if os.Getenv("DEBUG") != "" {
			log.Printf("[WARN] Uninformative evidence for belief %s: P(E|H)=%.4f, P(E|¬H)=%.4f (equal within tolerance). Evidence ID: %s",
				beliefID, likelihoodIfTrue, likelihoodIfFalse, evidenceID)
		}

		belief.Evidence = append(belief.Evidence, evidenceID)
		belief.UpdatedAt = time.Now()
		if belief.Metadata == nil {
			belief.Metadata = make(map[string]interface{})
		}
		belief.Metadata["last_update_uninformative"] = true
		belief.Metadata["uninformative_reason"] = "likelihoods_equal"

		if pr.metrics != nil {
			pr.metrics.RecordUninformative()
		}

		return belief, nil
	}

	// Full Bayes' theorem
	prior := belief.Probability
	priorNot := 1.0 - prior

	numerator := likelihoodIfTrue * prior
	denominator := (likelihoodIfTrue * prior) + (likelihoodIfFalse * priorNot)

	var posterior float64
	if denominator > 0 {
		posterior = numerator / denominator
	} else {
		posterior = prior // No update if denominator is zero
	}

	// Clamp to valid probability range (should already be in range, but safety)
	posterior = math.Max(0, math.Min(1, posterior))

	belief.Probability = posterior
	belief.Evidence = append(belief.Evidence, evidenceID)
	belief.UpdatedAt = time.Now()

	if pr.metrics != nil {
		pr.metrics.RecordUpdate()
	}

	return belief, nil
}

// UpdateBeliefWithEvidence uses evidence strength to update belief.
// This method estimates both P(E|H) and P(E|¬H) from evidence quality using
// a configurable LikelihoodEstimator.
func (pr *ProbabilisticReasoner) UpdateBeliefWithEvidence(beliefID string, evidence *types.Evidence) (*types.ProbabilisticBelief, error) {
	_, exists := pr.beliefs[beliefID]
	if !exists {
		return nil, fmt.Errorf("belief not found: %s", beliefID)
	}

	// Use the likelihood estimator to convert evidence quality to conditional probabilities
	likelihoodIfTrue, likelihoodIfFalse, err := pr.estimator.EstimateLikelihoods(evidence)
	if err != nil {
		if pr.metrics != nil {
			pr.metrics.RecordError()
		}
		return nil, fmt.Errorf("failed to estimate likelihoods: %w", err)
	}

	// Use the mathematically correct UpdateBeliefFull method
	return pr.UpdateBeliefFull(beliefID, evidence.ID, likelihoodIfTrue, likelihoodIfFalse)
}

// GetBelief retrieves a belief by ID
func (pr *ProbabilisticReasoner) GetBelief(beliefID string) (*types.ProbabilisticBelief, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	belief, exists := pr.beliefs[beliefID]
	if !exists {
		return nil, fmt.Errorf("belief not found: %s", beliefID)
	}
	return belief, nil
}

// CombineBeliefs combines multiple independent beliefs using probability theory
// For independent events: P(A and B) = P(A) * P(B)
func (pr *ProbabilisticReasoner) CombineBeliefs(beliefIDs []string, operation string) (float64, error) {
	if len(beliefIDs) == 0 {
		return 0, fmt.Errorf("no beliefs provided")
	}

	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var result float64

	switch operation {
	case "and": // P(A and B) = P(A) * P(B) for independent events
		result = 1.0
		for _, id := range beliefIDs {
			belief, exists := pr.beliefs[id]
			if !exists {
				return 0, fmt.Errorf("belief not found: %s", id)
			}
			result *= belief.Probability
		}

	case "or": // P(A or B) = P(A) + P(B) - P(A)*P(B) for independent events
		result = 0.0
		for _, id := range beliefIDs {
			belief, exists := pr.beliefs[id]
			if !exists {
				return 0, fmt.Errorf("belief not found: %s", id)
			}
			// P(A or B) = 1 - P(not A and not B) = 1 - (1-P(A))*(1-P(B))
			result = 1 - (1-result)*(1-belief.Probability)
		}

	default:
		return 0, fmt.Errorf("unknown operation: %s (use 'and' or 'or')", operation)
	}

	if pr.metrics != nil {
		pr.metrics.RecordBeliefsCombined()
	}

	return result, nil
}

// GetMetrics returns current probabilistic reasoning metrics
func (pr *ProbabilisticReasoner) GetMetrics() map[string]interface{} {
	if pr.metrics == nil {
		return map[string]interface{}{}
	}

	stats := pr.metrics.GetStats()
	return map[string]interface{}{
		"updates_total":         stats["updates_total"],
		"updates_uninformative": stats["updates_uninformative"],
		"updates_error":         stats["updates_error"],
		"beliefs_created":       stats["beliefs_created"],
		"beliefs_combined":      stats["beliefs_combined"],
		"uninformative_rate":    pr.metrics.GetUninformativeRate(),
		"error_rate":            pr.metrics.GetErrorRate(),
	}
}

// EstimateConfidence estimates confidence in a conclusion based on supporting evidence
func (pr *ProbabilisticReasoner) EstimateConfidence(evidences []*types.Evidence) float64 {
	if len(evidences) == 0 {
		return 0.5 // Neutral confidence with no evidence
	}

	supportingScore := 0.0
	refutingScore := 0.0

	for _, ev := range evidences {
		if ev.SupportsClaim {
			supportingScore += ev.OverallScore
		} else {
			refutingScore += ev.OverallScore
		}
	}

	total := supportingScore + refutingScore
	if total == 0 {
		return 0.5
	}

	// Confidence is ratio of supporting evidence to total evidence
	confidence := supportingScore / total
	return math.Max(0, math.Min(1, confidence))
}

