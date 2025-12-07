package reasoning

import (
	"context"
	"testing"
	"time"

	"unified-thinking/internal/storage"

	"github.com/stretchr/testify/assert"
)

// mockHypothesisGenerator for testing
type mockHypothesisGenerator struct{}

func (m *mockHypothesisGenerator) GenerateHypotheses(ctx context.Context, prompt string) (string, error) {
	return `{"hypotheses": [{"description": "Mock hypothesis", "assumptions": ["test"], "predictions": ["test"], "parsimony": 0.8, "prior_probability": 0.5}]}`, nil
}

func TestNewAbductiveReasoner(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	assert.NotNil(t, ar)
	assert.NotNil(t, ar.storage)
}

func TestAbductiveReasoner_GenerateHypotheses_SingleCause(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	observations := []*Observation{
		{
			ID:          "obs-1",
			Description: "System performance degraded significantly",
			Confidence:  0.9,
			Timestamp:   time.Now(),
		},
		{
			ID:          "obs-2",
			Description: "Database queries are slow and performance issues detected",
			Confidence:  0.85,
			Timestamp:   time.Now().Add(5 * time.Minute),
		},
		{
			ID:          "obs-3",
			Description: "Users reporting performance problems",
			Confidence:  0.8,
			Timestamp:   time.Now().Add(10 * time.Minute),
		},
	}

	req := &GenerateHypothesesRequest{
		Observations:  observations,
		MaxHypotheses: 10,
		MinParsimony:  0.0,
	}

	hypotheses, err := ar.GenerateHypotheses(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, hypotheses)

	// LLM-based generation should produce at least one hypothesis
	assert.GreaterOrEqual(t, len(hypotheses), 1, "Should generate hypotheses")
	// All hypotheses should explain all observations
	for _, h := range hypotheses {
		assert.Equal(t, len(observations), len(h.Observations), "Hypothesis should explain all observations")
		assert.GreaterOrEqual(t, h.Parsimony, 0.0, "Parsimony should be valid")
		assert.LessOrEqual(t, h.Parsimony, 1.0, "Parsimony should be valid")
	}
}

func TestAbductiveReasoner_GenerateHypotheses_MultipleCauses(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	// Observations that don't seem related
	observations := []*Observation{
		{
			ID:          "obs-1",
			Description: "Database server crashed",
			Confidence:  0.9,
			Timestamp:   time.Now(),
		},
		{
			ID:          "obs-2",
			Description: "Network switch failed",
			Confidence:  0.85,
			Timestamp:   time.Now().Add(2 * time.Hour),
		},
	}

	req := &GenerateHypothesesRequest{
		Observations:  observations,
		MaxHypotheses: 10,
		MinParsimony:  0.0,
	}

	hypotheses, err := ar.GenerateHypotheses(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, hypotheses)
}

func TestAbductiveReasoner_GenerateHypotheses_MaxLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	observations := []*Observation{
		{ID: "obs-1", Description: "Event A", Confidence: 0.9, Timestamp: time.Now()},
		{ID: "obs-2", Description: "Event B", Confidence: 0.8, Timestamp: time.Now()},
	}

	req := &GenerateHypothesesRequest{
		Observations:  observations,
		MaxHypotheses: 2,
	}

	hypotheses, err := ar.GenerateHypotheses(ctx, req)

	assert.NoError(t, err)
	assert.LessOrEqual(t, len(hypotheses), 2)
}

func TestAbductiveReasoner_GenerateHypotheses_ParsimonyFilter(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	observations := []*Observation{
		{ID: "obs-1", Description: "Event A", Confidence: 0.9, Timestamp: time.Now()},
		{ID: "obs-2", Description: "Event B", Confidence: 0.8, Timestamp: time.Now()},
	}

	req := &GenerateHypothesesRequest{
		Observations:  observations,
		MaxHypotheses: 10,
		MinParsimony:  0.8, // High threshold
	}

	hypotheses, err := ar.GenerateHypotheses(ctx, req)

	assert.NoError(t, err)
	// All returned hypotheses should meet parsimony threshold
	for _, h := range hypotheses {
		assert.GreaterOrEqual(t, h.Parsimony, 0.8)
	}
}

func TestAbductiveReasoner_EvaluateHypotheses_Combined(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	observations := []*Observation{
		{ID: "obs-1", Description: "Event A", Confidence: 0.9, Timestamp: time.Now()},
		{ID: "obs-2", Description: "Event B", Confidence: 0.8, Timestamp: time.Now()},
	}

	hypotheses := []*Hypothesis{
		{
			ID:               "hyp-1",
			Description:      "Simple explanation",
			Observations:     []string{"obs-1", "obs-2"},
			PriorProbability: 0.7,
			Assumptions:      []string{"One assumption"},
		},
		{
			ID:               "hyp-2",
			Description:      "Complex explanation with many words and assumptions",
			Observations:     []string{"obs-1"},
			PriorProbability: 0.3,
			Assumptions:      []string{"Assumption 1", "Assumption 2", "Assumption 3"},
		},
	}

	req := &EvaluateHypothesesRequest{
		Observations: observations,
		Hypotheses:   hypotheses,
		Method:       MethodCombined,
		Weights:      DefaultEvaluationWeights(),
	}

	ranked, err := ar.EvaluateHypotheses(ctx, req)

	assert.NoError(t, err)
	assert.Len(t, ranked, 2)

	// Hypotheses should be ranked (best first)
	assert.Greater(t, ranked[0].PosteriorProbability, ranked[1].PosteriorProbability)

	// First hypothesis should be better (explains more, simpler, higher prior)
	assert.Equal(t, "hyp-1", ranked[0].ID)
}

func TestAbductiveReasoner_EvaluateHypotheses_Bayesian(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	observations := []*Observation{
		{ID: "obs-1", Description: "Event", Confidence: 1.0, Timestamp: time.Now()},
	}

	hypotheses := []*Hypothesis{
		{
			ID:               "hyp-1",
			Description:      "Hypothesis",
			Observations:     []string{"obs-1"},
			PriorProbability: 0.8,
		},
	}

	req := &EvaluateHypothesesRequest{
		Observations: observations,
		Hypotheses:   hypotheses,
		Method:       MethodBayesian,
	}

	ranked, err := ar.EvaluateHypotheses(ctx, req)

	assert.NoError(t, err)
	assert.Len(t, ranked, 1)
	assert.Greater(t, ranked[0].PosteriorProbability, 0.0)
	assert.LessOrEqual(t, ranked[0].PosteriorProbability, 1.0)
}

func TestAbductiveReasoner_EvaluateHypotheses_Parsimony(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	observations := []*Observation{
		{ID: "obs-1", Description: "Event", Confidence: 1.0, Timestamp: time.Now()},
	}

	hypotheses := []*Hypothesis{
		{
			ID:           "hyp-simple",
			Description:  "Simple",
			Observations: []string{"obs-1"},
			Assumptions:  []string{"One"},
		},
		{
			ID:           "hyp-complex",
			Description:  "Very complex explanation with many words",
			Observations: []string{"obs-1"},
			Assumptions:  []string{"One", "Two", "Three", "Four"},
		},
	}

	req := &EvaluateHypothesesRequest{
		Observations: observations,
		Hypotheses:   hypotheses,
		Method:       MethodParsimony,
	}

	ranked, err := ar.EvaluateHypotheses(ctx, req)

	assert.NoError(t, err)
	// Simpler hypothesis should rank higher
	assert.Equal(t, "hyp-simple", ranked[0].ID)
	assert.Greater(t, ranked[0].PosteriorProbability, ranked[1].PosteriorProbability)
}

func TestAbductiveReasoner_PerformAbductiveInference(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	observations := []*Observation{
		{
			ID:          "obs-1",
			Description: "System load increased dramatically",
			Confidence:  0.95,
			Timestamp:   time.Now(),
		},
		{
			ID:          "obs-2",
			Description: "Memory usage spiked unexpectedly",
			Confidence:  0.9,
			Timestamp:   time.Now().Add(1 * time.Minute),
		},
		{
			ID:          "obs-3",
			Description: "Response times degraded across services",
			Confidence:  0.85,
			Timestamp:   time.Now().Add(2 * time.Minute),
		},
	}

	inference, err := ar.PerformAbductiveInference(ctx, observations, 5)

	assert.NoError(t, err)
	assert.NotNil(t, inference)
	assert.NotEmpty(t, inference.Hypotheses)
	assert.NotNil(t, inference.BestHypothesis)
	assert.NotEmpty(t, inference.RankedHypotheses)
	assert.Greater(t, inference.Confidence, 0.0)
	assert.Equal(t, len(observations), len(inference.Observations))
}

func TestAbductiveReasoner_CalculateExplanatoryPower(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	observations := []*Observation{
		{ID: "obs-1", Confidence: 1.0},
		{ID: "obs-2", Confidence: 0.8},
		{ID: "obs-3", Confidence: 0.6},
	}

	tests := []struct {
		name     string
		hyp      *Hypothesis
		minPower float64
	}{
		{
			name: "explains_all_observations",
			hyp: &Hypothesis{
				Observations: []string{"obs-1", "obs-2", "obs-3"},
			},
			minPower: 0.7, // Should be high
		},
		{
			name: "explains_some_observations",
			hyp: &Hypothesis{
				Observations: []string{"obs-1"},
			},
			minPower: 0.2, // Should be lower
		},
		{
			name: "explains_no_observations",
			hyp: &Hypothesis{
				Observations: []string{},
			},
			minPower: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			power := ar.calculateExplanatoryPower(tt.hyp, observations)
			assert.GreaterOrEqual(t, power, tt.minPower)
		})
	}
}

func TestAbductiveReasoner_CalculateParsimony(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	tests := []struct {
		name        string
		hyp         *Hypothesis
		expectRange [2]float64 // min, max
	}{
		{
			name: "simple_hypothesis",
			hyp: &Hypothesis{
				Description: "Simple",
				Assumptions: []string{"One"},
			},
			expectRange: [2]float64{0.4, 1.0},
		},
		{
			name: "complex_hypothesis",
			hyp: &Hypothesis{
				Description: "Very complex explanation with many words and detailed analysis",
				Assumptions: []string{"One", "Two", "Three", "Four", "Five"},
			},
			expectRange: [2]float64{0.0, 0.4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsimony := ar.calculateParsimony(tt.hyp)
			assert.GreaterOrEqual(t, parsimony, tt.expectRange[0])
			assert.LessOrEqual(t, parsimony, tt.expectRange[1])
		})
	}
}

func TestAbductiveReasoner_FindCommonThemes(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	observations := []*Observation{
		{Description: "The server performance degraded"},
		{Description: "Database performance issues detected"},
		{Description: "Performance monitoring shows problems"},
	}

	themes := ar.findCommonThemes(observations)

	assert.Contains(t, themes, "performance")
}

func TestAbductiveReasoner_HasTemporalPattern(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	now := time.Now()

	tests := []struct {
		name         string
		observations []*Observation
		expected     bool
	}{
		{
			name: "regular_intervals",
			observations: []*Observation{
				{Timestamp: now},
				{Timestamp: now.Add(1 * time.Hour)},
				{Timestamp: now.Add(2 * time.Hour)},
				{Timestamp: now.Add(3 * time.Hour)},
			},
			expected: true,
		},
		{
			name: "too_few_observations",
			observations: []*Observation{
				{Timestamp: now},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ar.hasTemporalPattern(tt.observations)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultEvaluationWeights(t *testing.T) {
	weights := DefaultEvaluationWeights()

	assert.NotNil(t, weights)
	assert.Equal(t, 0.4, weights.ExplanatoryPower)
	assert.Equal(t, 0.3, weights.Parsimony)
	assert.Equal(t, 0.3, weights.PriorProbability)

	// Weights should sum to 1.0
	sum := weights.ExplanatoryPower + weights.Parsimony + weights.PriorProbability
	assert.InDelta(t, 1.0, sum, 0.01)
}

func TestAbductiveReasoner_NoObservations(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	req := &GenerateHypothesesRequest{
		Observations: []*Observation{},
	}

	_, err := ar.GenerateHypotheses(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no observations")
}

func TestAbductiveReasoner_NoHypotheses(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockGen := &mockHypothesisGenerator{}
	ar := NewAbductiveReasoner(store, mockGen)

	ctx := context.Background()

	req := &EvaluateHypothesesRequest{
		Observations: []*Observation{{ID: "obs-1"}},
		Hypotheses:   []*Hypothesis{},
	}

	_, err := ar.EvaluateHypotheses(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no hypotheses")
}
