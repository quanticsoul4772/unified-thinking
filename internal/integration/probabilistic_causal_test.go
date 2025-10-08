package integration

import (
	"context"
	"testing"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestNewProbabilisticCausalIntegration(t *testing.T) {
	probReasoner := reasoning.NewProbabilisticReasoner()
	causalReasoner := reasoning.NewCausalReasoner()

	integration := NewProbabilisticCausalIntegration(probReasoner, causalReasoner)

	assert.NotNil(t, integration)
	assert.NotNil(t, integration.probReasoner)
	assert.NotNil(t, integration.causalReasoner)
}

func TestProbabilisticCausalIntegration_CalculateGraphComplexity(t *testing.T) {
	probReasoner := reasoning.NewProbabilisticReasoner()
	causalReasoner := reasoning.NewCausalReasoner()
	integration := NewProbabilisticCausalIntegration(probReasoner, causalReasoner)

	// Empty graph
	emptyGraph := &types.CausalGraph{
		Variables: []*types.CausalVariable{},
		Links:     []*types.CausalLink{},
	}
	assert.Equal(t, 0.0, integration.calculateGraphComplexity(emptyGraph))

	// Graph with nodes but no edges
	sparseGraph := &types.CausalGraph{
		Variables: []*types.CausalVariable{
			{ID: "A"},
			{ID: "B"},
			{ID: "C"},
		},
		Links: []*types.CausalLink{},
	}
	assert.Equal(t, 0.0, integration.calculateGraphComplexity(sparseGraph))

	// Graph with some edges
	partialGraph := &types.CausalGraph{
		Variables: []*types.CausalVariable{
			{ID: "A"},
			{ID: "B"},
			{ID: "C"},
		},
		Links: []*types.CausalLink{
			{From: "A", To: "B"},
			{From: "B", To: "C"},
		},
	}
	// 3 nodes = 3*2 = 6 possible edges, 2 actual edges = 2/6 = 0.333...
	assert.InDelta(t, 0.333, integration.calculateGraphComplexity(partialGraph), 0.01)
}

func TestProbabilisticCausalIntegration_ExtractPosteriorFromIntervention(t *testing.T) {
	probReasoner := reasoning.NewProbabilisticReasoner()
	causalReasoner := reasoning.NewCausalReasoner()
	integration := NewProbabilisticCausalIntegration(probReasoner, causalReasoner)

	// No effects
	emptyResult := &types.CausalIntervention{
		PredictedEffects: []*types.PredictedEffect{},
	}
	assert.Equal(t, 0.5, integration.extractPosteriorFromIntervention(emptyResult))

	// Positive effects
	positiveResult := &types.CausalIntervention{
		PredictedEffects: []*types.PredictedEffect{
			{Variable: "X", Magnitude: 0.5},
			{Variable: "Y", Magnitude: 0.7},
		},
	}
	posterior := integration.extractPosteriorFromIntervention(positiveResult)
	assert.Greater(t, posterior, 0.5)
	assert.LessOrEqual(t, posterior, 0.9)
}

func TestProbabilisticCausalIntegration_StrengthenCausalLinks(t *testing.T) {
	probReasoner := reasoning.NewProbabilisticReasoner()
	causalReasoner := reasoning.NewCausalReasoner()
	integration := NewProbabilisticCausalIntegration(probReasoner, causalReasoner)

	graph := &types.CausalGraph{
		Variables: []*types.CausalVariable{
			{ID: "A"},
			{ID: "B"},
		},
		Links: []*types.CausalLink{
			{From: "A", To: "B", Strength: 0.5},
		},
	}

	belief := &types.ProbabilisticBelief{
		Statement:   "Test",
		Probability: 0.9,
	}

	originalStrength := graph.Links[0].Strength
	integration.strengthenCausalLinks(graph, belief)

	assert.Greater(t, graph.Links[0].Strength, originalStrength)
	assert.LessOrEqual(t, graph.Links[0].Strength, 1.0)
}

func TestProbabilisticCausalIntegration_WeakenCausalLinks(t *testing.T) {
	probReasoner := reasoning.NewProbabilisticReasoner()
	causalReasoner := reasoning.NewCausalReasoner()
	integration := NewProbabilisticCausalIntegration(probReasoner, causalReasoner)

	graph := &types.CausalGraph{
		Variables: []*types.CausalVariable{
			{ID: "A"},
			{ID: "B"},
		},
		Links: []*types.CausalLink{
			{From: "A", To: "B", Strength: 0.5},
		},
	}

	belief := &types.ProbabilisticBelief{
		Statement:   "Test",
		Probability: 0.2,
	}

	originalStrength := graph.Links[0].Strength
	integration.weakenCausalLinks(graph, belief)

	assert.Less(t, graph.Links[0].Strength, originalStrength)
	assert.GreaterOrEqual(t, graph.Links[0].Strength, 0.0)
}

func TestProbabilisticCausalIntegration_CalculateConvergence(t *testing.T) {
	probReasoner := reasoning.NewProbabilisticReasoner()
	causalReasoner := reasoning.NewCausalReasoner()
	integration := NewProbabilisticCausalIntegration(probReasoner, causalReasoner)

	// High posterior belief + moderate complexity = high convergence
	highBelief := &types.ProbabilisticBelief{
		Statement:   "Test",
		Probability: 0.9,
	}

	moderateGraph := &types.CausalGraph{
		Variables: []*types.CausalVariable{
			{ID: "A"},
			{ID: "B"},
			{ID: "C"},
		},
		Links: []*types.CausalLink{
			{From: "A", To: "B"},
			{From: "B", To: "C"},
		},
	}

	convergence := integration.calculateConvergence(highBelief, moderateGraph)
	assert.Greater(t, convergence, 0.5)
	assert.LessOrEqual(t, convergence, 1.0)

	// Low posterior belief
	lowBelief := &types.ProbabilisticBelief{
		Statement:   "Test",
		Probability: 0.2,
	}

	lowConvergence := integration.calculateConvergence(lowBelief, moderateGraph)
	// Low posterior still counts as stable (certain it's false)
	assert.Greater(t, lowConvergence, 0.5)
}

func TestProbabilisticCausalIntegration_CreateFeedbackLoop(t *testing.T) {
	probReasoner := reasoning.NewProbabilisticReasoner()
	causalReasoner := reasoning.NewCausalReasoner()
	integration := NewProbabilisticCausalIntegration(probReasoner, causalReasoner)

	ctx := context.Background()

	// Create a belief
	belief, _ := probReasoner.CreateBelief("Test belief", 0.5)

	// Create a causal graph
	graph, _ := causalReasoner.BuildCausalGraph("Test graph", []string{
		"A causes B",
		"B causes C",
	})

	// Run feedback loop
	result, err := integration.CreateFeedbackLoop(ctx, belief.ID, graph.ID, 3)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.LessOrEqual(t, result.TotalIterations, 3)
	assert.NotEmpty(t, result.Iterations)
	assert.NotNil(t, result.FinalBelief)
	assert.NotNil(t, result.FinalGraph)

	// Check that iterations have proper structure
	for _, iteration := range result.Iterations {
		assert.Greater(t, iteration.IterationNum, 0)
		assert.GreaterOrEqual(t, iteration.BeliefPosterior, 0.0)
		assert.LessOrEqual(t, iteration.BeliefPosterior, 1.0)
		assert.GreaterOrEqual(t, iteration.ConvergenceScore, 0.0)
		assert.LessOrEqual(t, iteration.ConvergenceScore, 1.0)
	}
}
