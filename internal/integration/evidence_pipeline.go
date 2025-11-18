// Package integration provides cross-system integration capabilities.
//
// This module enables automatic propagation of evidence updates to probabilistic
// beliefs, causal graphs, and decision frameworks, ensuring consistency across
// all reasoning systems.
package integration

import (
	"fmt"
	"sync"

	"unified-thinking/internal/analysis"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

// EvidencePipeline manages evidence-driven updates across reasoning systems
type EvidencePipeline struct {
	probabilisticReasoner *reasoning.ProbabilisticReasoner
	causalReasoner        *reasoning.CausalReasoner
	decisionMaker         *reasoning.DecisionMaker
	evidenceAnalyzer      *analysis.EvidenceAnalyzer

	// Track relationships
	evidenceToBeliefs   map[string][]string // evidence_id -> belief_ids
	evidenceToCausal    map[string][]string // evidence_id -> graph_ids
	evidenceToDecisions map[string][]string // evidence_id -> decision_ids

	mu sync.RWMutex
}

// NewEvidencePipeline creates a new evidence integration pipeline
func NewEvidencePipeline(
	probReasoner *reasoning.ProbabilisticReasoner,
	causalReasoner *reasoning.CausalReasoner,
	decisionMaker *reasoning.DecisionMaker,
	evidenceAnalyzer *analysis.EvidenceAnalyzer,
) *EvidencePipeline {
	return &EvidencePipeline{
		probabilisticReasoner: probReasoner,
		causalReasoner:        causalReasoner,
		decisionMaker:         decisionMaker,
		evidenceAnalyzer:      evidenceAnalyzer,
		evidenceToBeliefs:     make(map[string][]string),
		evidenceToCausal:      make(map[string][]string),
		evidenceToDecisions:   make(map[string][]string),
	}
}

// PipelineResult contains the results of evidence propagation
type PipelineResult struct {
	EvidenceID       string                       `json:"evidence_id"`
	UpdatedBeliefs   []*types.ProbabilisticBelief `json:"updated_beliefs"`
	UpdatedGraphs    []*types.CausalGraph         `json:"updated_graphs,omitempty"`
	UpdatedDecisions []*types.Decision            `json:"updated_decisions,omitempty"`
	Changes          []string                     `json:"changes"`
	Status           string                       `json:"status"`
}

// ProcessEvidence assesses evidence and propagates updates across all systems
func (ep *EvidencePipeline) ProcessEvidence(content, source, claimID string, supportsClaim bool) (*PipelineResult, error) {
	// Step 1: Assess evidence quality
	evidence, err := ep.evidenceAnalyzer.AssessEvidence(content, source, claimID, supportsClaim)
	if err != nil {
		return nil, fmt.Errorf("evidence assessment failed: %w", err)
	}

	result := &PipelineResult{
		EvidenceID:       evidence.ID,
		UpdatedBeliefs:   []*types.ProbabilisticBelief{},
		UpdatedGraphs:    []*types.CausalGraph{},
		UpdatedDecisions: []*types.Decision{},
		Changes:          []string{},
		Status:           "success",
	}

	// Step 2: Update probabilistic beliefs
	updatedBeliefs, err := ep.updateBeliefs(evidence)
	if err != nil {
		result.Changes = append(result.Changes, fmt.Sprintf("Belief update warning: %v", err))
	} else {
		result.UpdatedBeliefs = updatedBeliefs
		for _, belief := range updatedBeliefs {
			result.Changes = append(result.Changes,
				fmt.Sprintf("Updated belief '%s': %.3f -> %.3f",
					belief.Statement, belief.PriorProb, belief.Probability))
		}
	}

	// Step 3: Update causal graph edge strengths
	updatedGraphs, err := ep.updateCausalGraphs(evidence)
	if err != nil {
		result.Changes = append(result.Changes, fmt.Sprintf("Causal graph update warning: %v", err))
	} else {
		result.UpdatedGraphs = updatedGraphs
		for _, graph := range updatedGraphs {
			result.Changes = append(result.Changes,
				fmt.Sprintf("Updated causal graph '%s' with evidence", graph.ID))
		}
	}

	// Step 4: Re-evaluate affected decisions
	updatedDecisions, err := ep.updateDecisions(evidence)
	if err != nil {
		result.Changes = append(result.Changes, fmt.Sprintf("Decision update warning: %v", err))
	} else {
		result.UpdatedDecisions = updatedDecisions
		for _, decision := range updatedDecisions {
			result.Changes = append(result.Changes,
				fmt.Sprintf("Re-evaluated decision '%s'", decision.Question))
		}
	}

	return result, nil
}

// LinkEvidenceToBelief creates a relationship between evidence and belief
func (ep *EvidencePipeline) LinkEvidenceToBelief(evidenceID, beliefID string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if beliefs, exists := ep.evidenceToBeliefs[evidenceID]; exists {
		// Check if already linked
		for _, id := range beliefs {
			if id == beliefID {
				return
			}
		}
		ep.evidenceToBeliefs[evidenceID] = append(beliefs, beliefID)
	} else {
		ep.evidenceToBeliefs[evidenceID] = []string{beliefID}
	}
}

// LinkEvidenceToCausalGraph links evidence to a causal graph
func (ep *EvidencePipeline) LinkEvidenceToCausalGraph(evidenceID, graphID string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if graphs, exists := ep.evidenceToCausal[evidenceID]; exists {
		for _, id := range graphs {
			if id == graphID {
				return
			}
		}
		ep.evidenceToCausal[evidenceID] = append(graphs, graphID)
	} else {
		ep.evidenceToCausal[evidenceID] = []string{graphID}
	}
}

// LinkEvidenceToDecision links evidence to a decision
func (ep *EvidencePipeline) LinkEvidenceToDecision(evidenceID, decisionID string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if decisions, exists := ep.evidenceToDecisions[evidenceID]; exists {
		for _, id := range decisions {
			if id == decisionID {
				return
			}
		}
		ep.evidenceToDecisions[evidenceID] = append(decisions, decisionID)
	} else {
		ep.evidenceToDecisions[evidenceID] = []string{decisionID}
	}
}

// updateBeliefs updates probabilistic beliefs based on evidence
func (ep *EvidencePipeline) updateBeliefs(evidence *types.Evidence) ([]*types.ProbabilisticBelief, error) {
	ep.mu.RLock()
	beliefIDs, exists := ep.evidenceToBeliefs[evidence.ID]
	ep.mu.RUnlock()

	if !exists || len(beliefIDs) == 0 {
		// Try to find beliefs related to the claim
		if evidence.ClaimID != "" {
			// Create belief from claim if doesn't exist
			belief, err := ep.probabilisticReasoner.CreateBelief(evidence.ClaimID, 0.5)
			if err != nil {
				return nil, err
			}
			ep.LinkEvidenceToBelief(evidence.ID, belief.ID)
			beliefIDs = []string{belief.ID}
		} else {
			return []*types.ProbabilisticBelief{}, nil
		}
	}

	updatedBeliefs := []*types.ProbabilisticBelief{}

	for _, beliefID := range beliefIDs {
		// Calculate likelihood based on evidence quality
		likelihood := ep.calculateLikelihood(evidence)

		// Evidence probability (how likely we'd see this evidence)
		evidenceProb := evidence.OverallScore

		// Update belief using Bayesian inference
		updatedBelief, err := ep.probabilisticReasoner.UpdateBelief(
			beliefID,
			evidence.ID,
			likelihood,
			evidenceProb,
		)
		if err != nil {
			continue
		}

		updatedBeliefs = append(updatedBeliefs, updatedBelief)
	}

	return updatedBeliefs, nil
}

// updateCausalGraphs updates causal graph edge strengths based on evidence
func (ep *EvidencePipeline) updateCausalGraphs(evidence *types.Evidence) ([]*types.CausalGraph, error) {
	ep.mu.RLock()
	graphIDs, exists := ep.evidenceToCausal[evidence.ID]
	ep.mu.RUnlock()

	if !exists || len(graphIDs) == 0 {
		return []*types.CausalGraph{}, nil
	}

	updatedGraphs := []*types.CausalGraph{}

	for _, graphID := range graphIDs {
		graph, err := ep.causalReasoner.GetGraph(graphID)
		if err != nil {
			continue
		}

		// Update causal link strengths based on evidence
		strengthAdjustment := ep.calculateStrengthAdjustment(evidence)

		for _, link := range graph.Links {
			// Add evidence to link
			link.Evidence = append(link.Evidence, evidence.ID)

			// Adjust confidence based on evidence quality
			if evidence.SupportsClaim {
				link.Confidence = min(link.Confidence+strengthAdjustment, 1.0)
			} else {
				link.Confidence = max(link.Confidence-strengthAdjustment, 0.0)
			}
		}

		updatedGraphs = append(updatedGraphs, graph)
	}

	return updatedGraphs, nil
}

// updateDecisions re-evaluates decisions based on new evidence
func (ep *EvidencePipeline) updateDecisions(evidence *types.Evidence) ([]*types.Decision, error) {
	ep.mu.RLock()
	decisionIDs, exists := ep.evidenceToDecisions[evidence.ID]
	ep.mu.RUnlock()

	if !exists || len(decisionIDs) == 0 {
		return []*types.Decision{}, nil
	}

	if ep.decisionMaker == nil {
		return []*types.Decision{}, nil
	}

	updatedDecisions := []*types.Decision{}

	for _, decisionID := range decisionIDs {
		decision, err := ep.decisionMaker.GetDecision(decisionID)
		if err != nil {
			continue
		}

		// Calculate score adjustments based on evidence
		scoreAdjustments := ep.calculateDecisionScoreAdjustments(evidence, decision)
		if len(scoreAdjustments) == 0 {
			continue
		}

		// Recalculate the decision with adjusted scores
		updatedDecision, err := ep.decisionMaker.RecalculateDecision(decisionID, scoreAdjustments)
		if err != nil {
			continue
		}

		updatedDecisions = append(updatedDecisions, updatedDecision)
	}

	return updatedDecisions, nil
}

// calculateDecisionScoreAdjustments determines how evidence affects decision option scores
func (ep *EvidencePipeline) calculateDecisionScoreAdjustments(evidence *types.Evidence, decision *types.Decision) map[string]map[string]float64 {
	adjustments := make(map[string]map[string]float64)

	// Calculate base adjustment from evidence quality
	baseAdjustment := ep.calculateScoreAdjustment(evidence)

	for _, option := range decision.Options {
		// Check if evidence relates to this option
		if ep.evidenceRelatesToOption(evidence, option) {
			optionAdjustments := make(map[string]float64)

			// Apply adjustment to all criteria scores for this option
			for criterionID := range option.Scores {
				if evidence.SupportsClaim {
					optionAdjustments[criterionID] = baseAdjustment
				} else {
					optionAdjustments[criterionID] = -baseAdjustment
				}
			}

			if len(optionAdjustments) > 0 {
				adjustments[option.ID] = optionAdjustments
			}
		}
	}

	return adjustments
}

// calculateScoreAdjustment determines how much to adjust decision scores
func (ep *EvidencePipeline) calculateScoreAdjustment(evidence *types.Evidence) float64 {
	// Adjustment proportional to evidence quality
	return evidence.OverallScore * 0.15 // Max adjustment of 0.15
}

// evidenceRelatesToOption checks if evidence relates to a decision option
func (ep *EvidencePipeline) evidenceRelatesToOption(evidence *types.Evidence, option *types.DecisionOption) bool {
	// Simple heuristic: check if evidence content mentions option name
	// In production, this would use more sophisticated matching
	contentLower := toLower(evidence.Content)
	optionLower := toLower(option.Name)
	descLower := toLower(option.Description)

	return contains(contentLower, optionLower) || contains(contentLower, descLower)
}

// calculateLikelihood determines likelihood P(E|H) based on evidence quality
func (ep *EvidencePipeline) calculateLikelihood(evidence *types.Evidence) float64 {
	// Strong evidence has high likelihood if it supports, low if refutes
	if evidence.SupportsClaim {
		// P(seeing strong evidence | hypothesis true) = high
		return 0.5 + (evidence.OverallScore * 0.45) // Range: 0.5-0.95
	} else {
		// P(seeing refuting evidence | hypothesis true) = low
		return 0.5 - (evidence.OverallScore * 0.45) // Range: 0.05-0.5
	}
}

// calculateStrengthAdjustment determines how much to adjust causal link strength
func (ep *EvidencePipeline) calculateStrengthAdjustment(evidence *types.Evidence) float64 {
	// Adjustment proportional to evidence quality
	return evidence.OverallScore * 0.1 // Max adjustment of 0.1
}

// GetEvidenceImpact returns all systems affected by an evidence piece
func (ep *EvidencePipeline) GetEvidenceImpact(evidenceID string) map[string]interface{} {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	impact := make(map[string]interface{})

	if beliefs, exists := ep.evidenceToBeliefs[evidenceID]; exists {
		impact["beliefs"] = beliefs
	}

	if graphs, exists := ep.evidenceToCausal[evidenceID]; exists {
		impact["causal_graphs"] = graphs
	}

	if decisions, exists := ep.evidenceToDecisions[evidenceID]; exists {
		impact["decisions"] = decisions
	}

	return impact
}

// Helper functions

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
