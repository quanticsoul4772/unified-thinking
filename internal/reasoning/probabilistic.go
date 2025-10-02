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
		Metadata:    make(map[string]interface{}),
	}

	pr.beliefs[belief.ID] = belief
	return belief, nil
}

// UpdateBelief applies Bayesian update based on new evidence
// Uses simplified Bayes' theorem: P(H|E) = P(E|H) * P(H) / P(E)
// likelihood: P(E|H) - probability of evidence given hypothesis is true
// evidenceProb: P(E) - base rate probability of seeing this evidence
func (pr *ProbabilisticReasoner) UpdateBelief(beliefID string, evidenceID string, likelihood, evidenceProb float64) (*types.ProbabilisticBelief, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	belief, exists := pr.beliefs[beliefID]
	if !exists {
		return nil, fmt.Errorf("belief not found: %s", beliefID)
	}

	if likelihood < 0 || likelihood > 1 {
		return nil, fmt.Errorf("likelihood must be between 0 and 1")
	}
	if evidenceProb <= 0 || evidenceProb > 1 {
		return nil, fmt.Errorf("evidence probability must be between 0 and 1 (exclusive 0)")
	}

	// Bayes' theorem: P(H|E) = P(E|H) * P(H) / P(E)
	prior := belief.Probability
	posterior := (likelihood * prior) / evidenceProb

	// Clamp to valid probability range
	posterior = math.Max(0, math.Min(1, posterior))

	belief.Probability = posterior
	belief.Evidence = append(belief.Evidence, evidenceID)
	belief.UpdatedAt = time.Now()

	return belief, nil
}

// UpdateBeliefWithEvidence uses evidence strength to update belief
// This is a simplified approach where we estimate likelihood from evidence quality
func (pr *ProbabilisticReasoner) UpdateBeliefWithEvidence(beliefID string, evidence *types.Evidence) (*types.ProbabilisticBelief, error) {
	_, exists := pr.beliefs[beliefID]
	if !exists {
		return nil, fmt.Errorf("belief not found: %s", beliefID)
	}

	// Estimate likelihood from evidence quality and reliability
	// Strong supporting evidence increases likelihood
	var likelihood float64
	if evidence.SupportsClaim {
		// Evidence supports the belief
		likelihood = 0.5 + (evidence.OverallScore * 0.4) // Range: 0.5-0.9
	} else {
		// Evidence refutes the belief
		likelihood = 0.5 - (evidence.OverallScore * 0.4) // Range: 0.1-0.5
	}

	// Use evidence relevance as base rate estimate
	evidenceProb := 0.5 // Neutral base rate

	return pr.UpdateBelief(beliefID, evidence.ID, likelihood, evidenceProb)
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
