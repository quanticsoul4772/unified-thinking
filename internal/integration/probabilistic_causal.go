// Package integration provides probabilistic-causal feedback integration.
package integration

import (
	"context"
	"fmt"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

// ProbabilisticCausalIntegration manages feedback between Bayesian beliefs and causal graphs
type ProbabilisticCausalIntegration struct {
	probReasoner   *reasoning.ProbabilisticReasoner
	causalReasoner *reasoning.CausalReasoner
}

// NewProbabilisticCausalIntegration creates a new integration
func NewProbabilisticCausalIntegration(
	probReasoner *reasoning.ProbabilisticReasoner,
	causalReasoner *reasoning.CausalReasoner,
) *ProbabilisticCausalIntegration {
	return &ProbabilisticCausalIntegration{
		probReasoner:   probReasoner,
		causalReasoner: causalReasoner,
	}
}

// UpdateBeliefFromCausalGraph updates a belief based on causal graph intervention
func (pci *ProbabilisticCausalIntegration) UpdateBeliefFromCausalGraph(
	ctx context.Context,
	beliefID string,
	graphID string,
	interventionVariable string,
) (*types.ProbabilisticBelief, error) {
	// Get the belief (for validation)
	_, err := pci.probReasoner.GetBelief(beliefID)
	if err != nil {
		return nil, fmt.Errorf("failed to get belief: %w", err)
	}

	// Get the causal graph (for validation)
	_, err = pci.causalReasoner.GetGraph(graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get causal graph: %w", err)
	}

	// Simulate intervention on the causal variable
	result, err := pci.causalReasoner.SimulateIntervention(graphID, interventionVariable, "increase")
	if err != nil {
		return nil, fmt.Errorf("failed to simulate intervention: %w", err)
	}

	// Extract posterior probability from intervention result
	posteriorProb := pci.extractPosteriorFromIntervention(result)

	// Update belief with new posterior
	updatedBelief, err := pci.probReasoner.UpdateBelief(beliefID, "causal-intervention", posteriorProb, posteriorProb)
	if err != nil {
		return nil, fmt.Errorf("failed to update belief: %w", err)
	}

	// Add metadata linking to causal graph
	if updatedBelief.Metadata == nil {
		updatedBelief.Metadata = make(map[string]interface{})
	}
	updatedBelief.Metadata["causal_graph_id"] = graphID
	updatedBelief.Metadata["intervention_variable"] = interventionVariable
	updatedBelief.Metadata["intervention_result"] = result

	return updatedBelief, nil
}

// UpdateCausalGraphFromBelief modifies causal graph based on belief update
func (pci *ProbabilisticCausalIntegration) UpdateCausalGraphFromBelief(
	ctx context.Context,
	graphID string,
	beliefID string,
	evidenceStrength float64,
) (*types.CausalGraph, error) {
	// Get the belief
	belief, err := pci.probReasoner.GetBelief(beliefID)
	if err != nil {
		return nil, fmt.Errorf("failed to get belief: %w", err)
	}

	// Get the causal graph
	graph, err := pci.causalReasoner.GetGraph(graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get causal graph: %w", err)
	}

	// Determine if belief strength warrants causal link modification
	if belief.Probability > 0.8 && evidenceStrength > 0.7 {
		// Strong belief + strong evidence = strengthen causal links
		pci.strengthenCausalLinks(graph, belief)
	} else if belief.Probability < 0.3 && evidenceStrength > 0.7 {
		// Weak belief + strong evidence = weaken causal links
		pci.weakenCausalLinks(graph, belief)
	}

	// Add metadata linking to belief
	if graph.Metadata == nil {
		graph.Metadata = make(map[string]interface{})
	}
	graph.Metadata["belief_id"] = beliefID
	graph.Metadata["posterior_probability"] = belief.Probability
	graph.Metadata["evidence_strength"] = evidenceStrength

	return graph, nil
}

// CreateFeedbackLoop establishes bidirectional feedback between belief and causal graph
func (pci *ProbabilisticCausalIntegration) CreateFeedbackLoop(
	ctx context.Context,
	beliefID string,
	graphID string,
	iterations int,
) (*FeedbackResult, error) {
	result := &FeedbackResult{
		Iterations: make([]FeedbackIteration, 0),
	}

	currentBelief, err := pci.probReasoner.GetBelief(beliefID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial belief: %w", err)
	}

	currentGraph, err := pci.causalReasoner.GetGraph(graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial graph: %w", err)
	}

	for i := 0; i < iterations; i++ {
		iteration := FeedbackIteration{
			IterationNum:    i + 1,
			BeliefPosterior: currentBelief.Probability,
			GraphComplexity: pci.calculateGraphComplexity(currentGraph),
		}

		// Step 1: Update belief based on causal graph
		if len(currentGraph.Variables) > 0 {
			updatedBelief, err := pci.UpdateBeliefFromCausalGraph(ctx, beliefID, graphID, currentGraph.Variables[0].ID)
			if err == nil {
				currentBelief = updatedBelief
				iteration.BeliefUpdated = true
			}
		}

		// Step 2: Update causal graph based on belief
		evidenceStrength := 0.8 // Could be calculated from evidence quality
		updatedGraph, err := pci.UpdateCausalGraphFromBelief(ctx, graphID, beliefID, evidenceStrength)
		if err == nil {
			currentGraph = updatedGraph
			iteration.GraphUpdated = true
		}

		iteration.ConvergenceScore = pci.calculateConvergence(currentBelief, currentGraph)
		result.Iterations = append(result.Iterations, iteration)

		// Check for convergence
		if iteration.ConvergenceScore > 0.95 {
			result.Converged = true
			result.ConvergenceIteration = i + 1
			break
		}
	}

	result.FinalBelief = currentBelief
	result.FinalGraph = currentGraph
	result.TotalIterations = len(result.Iterations)

	return result, nil
}

// FeedbackResult contains the results of feedback loop
type FeedbackResult struct {
	Iterations           []FeedbackIteration        `json:"iterations"`
	TotalIterations      int                        `json:"total_iterations"`
	Converged            bool                       `json:"converged"`
	ConvergenceIteration int                        `json:"convergence_iteration,omitempty"`
	FinalBelief          *types.ProbabilisticBelief `json:"final_belief"`
	FinalGraph           *types.CausalGraph         `json:"final_graph"`
}

// FeedbackIteration represents one iteration of the feedback loop
type FeedbackIteration struct {
	IterationNum     int     `json:"iteration_num"`
	BeliefPosterior  float64 `json:"belief_posterior"`
	GraphComplexity  float64 `json:"graph_complexity"`
	BeliefUpdated    bool    `json:"belief_updated"`
	GraphUpdated     bool    `json:"graph_updated"`
	ConvergenceScore float64 `json:"convergence_score"`
}

// Helper methods

func (pci *ProbabilisticCausalIntegration) extractPosteriorFromIntervention(result *types.CausalIntervention) float64 {
	// Extract posterior probability from intervention analysis
	// This is a simplified version - actual implementation would analyze predicted effects
	if len(result.PredictedEffects) > 0 {
		// Average the magnitudes of predicted effects
		totalMagnitude := 0.0
		for _, effect := range result.PredictedEffects {
			totalMagnitude += effect.Magnitude
		}
		avgMagnitude := totalMagnitude / float64(len(result.PredictedEffects))

		// Convert magnitude to probability (0.5 baseline + normalized magnitude)
		return 0.5 + (avgMagnitude * 0.4) // Scale to 0.1-0.9 range
	}
	return 0.5 // Default neutral probability
}

func (pci *ProbabilisticCausalIntegration) strengthenCausalLinks(graph *types.CausalGraph, belief *types.ProbabilisticBelief) {
	// Strengthen causal links related to the belief
	// This would modify edge weights in the graph
	// Simplified implementation - actual would be more sophisticated
	for i := range graph.Links {
		if graph.Links[i].Strength < 0.9 {
			graph.Links[i].Strength = graph.Links[i].Strength * 1.1
			if graph.Links[i].Strength > 1.0 {
				graph.Links[i].Strength = 1.0
			}
		}
	}
}

func (pci *ProbabilisticCausalIntegration) weakenCausalLinks(graph *types.CausalGraph, belief *types.ProbabilisticBelief) {
	// Weaken causal links related to the belief
	for i := range graph.Links {
		if graph.Links[i].Strength > 0.1 {
			graph.Links[i].Strength = graph.Links[i].Strength * 0.9
			if graph.Links[i].Strength < 0.0 {
				graph.Links[i].Strength = 0.0
			}
		}
	}
}

func (pci *ProbabilisticCausalIntegration) calculateGraphComplexity(graph *types.CausalGraph) float64 {
	// Simple complexity measure: ratio of edges to possible edges
	if len(graph.Variables) <= 1 {
		return 0.0
	}
	maxPossibleEdges := len(graph.Variables) * (len(graph.Variables) - 1)
	return float64(len(graph.Links)) / float64(maxPossibleEdges)
}

func (pci *ProbabilisticCausalIntegration) calculateConvergence(belief *types.ProbabilisticBelief, graph *types.CausalGraph) float64 {
	// Measure convergence based on belief stability and graph stability
	// Simplified: high posterior + moderate complexity = high convergence
	beliefStability := belief.Probability
	if beliefStability < 0.5 {
		beliefStability = 1.0 - beliefStability // Low posterior also counts as stable
	}

	graphComplexity := pci.calculateGraphComplexity(graph)
	// Prefer moderate complexity (too simple or too complex = less converged)
	complexityScore := 1.0 - (2.0 * abs(graphComplexity-0.5))

	return (beliefStability + complexityScore) / 2.0
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
