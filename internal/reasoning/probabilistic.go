// Package reasoning provides advanced cognitive reasoning capabilities including
// probabilistic reasoning, Bayesian inference, and logical extensions.
package reasoning

import (
	"fmt"
	"math"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// ProbabilisticReasoner performs Bayesian inference and probabilistic reasoning
type ProbabilisticReasoner struct {
	mu      sync.RWMutex
	beliefs map[string]*types.ProbabilisticBelief
	counter int
}

// NewProbabilisticReasoner creates a new probabilistic reasoner
func NewProbabilisticReasoner() *ProbabilisticReasoner {
	return &ProbabilisticReasoner{
		beliefs: make(map[string]*types.ProbabilisticBelief),
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
	return belief, nil
}

// UpdateBelief applies Bayesian update based on new evidence
//
// DEPRECATED: This method is mathematically incorrect when evidenceProb is provided directly.
// Use UpdateBeliefFull instead, which properly requires both P(E|H) and P(E|¬H).
//
// This method is retained for backward compatibility but now delegates to UpdateBeliefFull
// with a default P(E|¬H) = 0.5 when evidenceProb is not provided.
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

// UpdateBeliefFull applies Bayesian update with full parameters
// This is the mathematically correct implementation of Bayes' theorem:
// P(H|E) = P(E|H)P(H) / [P(E|H)P(H) + P(E|¬H)P(¬H)]
//
// Why both likelihoods are needed:
// - P(E|H): How likely we'd see this evidence if the hypothesis is true
// - P(E|¬H): How likely we'd see this evidence if the hypothesis is false
// Without both, we cannot properly normalize the posterior probability.
//
// Example: Medical test with 99% sensitivity (P(positive|disease) = 0.99)
// and 95% specificity (P(negative|healthy) = 0.95, so P(positive|healthy) = 0.05).
// For a disease with 1% prevalence, a positive test only gives ~17% probability
// of having the disease, NOT 99%! This is why we need both likelihoods.
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
		return nil, fmt.Errorf("P(E|H) must be between 0 and 1, got: %f", likelihoodIfTrue)
	}
	if likelihoodIfFalse < 0 || likelihoodIfFalse > 1 {
		return nil, fmt.Errorf("P(E|¬H) must be between 0 and 1, got: %f", likelihoodIfFalse)
	}

	// Check for degenerate cases where both likelihoods are identical
	// This means the evidence provides no information
	if math.Abs(likelihoodIfTrue-likelihoodIfFalse) < 1e-10 {
		// Evidence is equally likely regardless of hypothesis truth
		// No update needed - posterior equals prior
		belief.Evidence = append(belief.Evidence, evidenceID)
		belief.UpdatedAt = time.Now()
		if belief.Metadata == nil {
			belief.Metadata = make(map[string]interface{})
		}
		belief.Metadata["last_update_uninformative"] = true
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

	return belief, nil
}

// UpdateBeliefWithEvidence uses evidence strength to update belief
// This method now properly estimates both P(E|H) and P(E|¬H) from evidence quality
func (pr *ProbabilisticReasoner) UpdateBeliefWithEvidence(beliefID string, evidence *types.Evidence) (*types.ProbabilisticBelief, error) {
	_, exists := pr.beliefs[beliefID]
	if !exists {
		return nil, fmt.Errorf("belief not found: %s", beliefID)
	}

	// Estimate likelihoods from evidence quality and reliability
	// We need both P(E|H) and P(E|¬H) for proper Bayesian update
	var likelihoodIfTrue, likelihoodIfFalse float64

	if evidence.SupportsClaim {
		// Evidence supports the belief
		// Strong evidence means high P(E|H) and low P(E|¬H)
		likelihoodIfTrue = 0.5 + (evidence.OverallScore * 0.4)  // Range: 0.5-0.9
		likelihoodIfFalse = 0.5 - (evidence.OverallScore * 0.3) // Range: 0.2-0.5
	} else {
		// Evidence refutes the belief
		// Strong refuting evidence means low P(E|H) and high P(E|¬H)
		likelihoodIfTrue = 0.5 - (evidence.OverallScore * 0.4)  // Range: 0.1-0.5
		likelihoodIfFalse = 0.5 + (evidence.OverallScore * 0.3) // Range: 0.5-0.8
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

	return result, nil
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
