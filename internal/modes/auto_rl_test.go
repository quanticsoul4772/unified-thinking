package modes

import (
	"context"
	"testing"

	"unified-thinking/internal/reinforcement"
	"unified-thinking/internal/storage"
)

// TestAutoMode_RLIntegration tests Thompson Sampling RL integration
func TestAutoMode_RLIntegration(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)

	auto := NewAutoMode(linear, tree, divergent)

	// Create mock RL storage with strategies
	mockStorage := &mockRLStorage{
		strategies: []*reinforcement.Strategy{
			{
				ID:             "strategy_linear",
				Name:           "Linear Sequential",
				Mode:           "linear",
				IsActive:       true,
				Alpha:          5.0, // Higher success rate
				Beta:           2.0,
				TotalTrials:    3,
				TotalSuccesses: 3,
			},
			{
				ID:             "strategy_tree",
				Name:           "Tree Exploration",
				Mode:           "tree",
				IsActive:       true,
				Alpha:          2.0,
				Beta:           3.0, // Lower success rate
				TotalTrials:    1,
				TotalSuccesses: 1,
			},
			{
				ID:             "strategy_divergent",
				Name:           "Divergent Creative",
				Mode:           "divergent",
				IsActive:       true,
				Alpha:          1.0, // No trials yet
				Beta:           1.0,
				TotalTrials:    0,
				TotalSuccesses: 0,
			},
		},
		outcomes:        []*reinforcement.Outcome{},
		alphaIncrements: make(map[string]int),
		betaIncrements:  make(map[string]int),
	}

	// Set RL storage (RL is always enabled when storage is available)
	t.Setenv("RL_OUTCOME_THRESHOLD", "0.8")

	err := auto.SetRLStorage(mockStorage)
	if err != nil {
		t.Fatalf("SetRLStorage failed: %v", err)
	}

	if !auto.rlEnabled {
		t.Fatal("RL should be enabled")
	}

	if auto.outcomeThreshold != 0.8 {
		t.Errorf("Expected threshold 0.8, got %.2f", auto.outcomeThreshold)
	}

	// Process a thought - should use Thompson selector
	input := ThoughtInput{
		Content:    "This is a general problem to solve",
		Confidence: 0.7,
	}

	result, err := auto.ProcessThought(context.Background(), input)
	if err != nil {
		t.Fatalf("ProcessThought failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Check that an outcome was recorded
	if len(mockStorage.outcomes) != 1 {
		t.Errorf("Expected 1 outcome recorded, got %d", len(mockStorage.outcomes))
	}

	if len(mockStorage.outcomes) > 0 {
		outcome := mockStorage.outcomes[0]

		// Verify outcome details
		if outcome.ProblemDescription != input.Content {
			t.Errorf("Expected problem description %q, got %q", input.Content, outcome.ProblemDescription)
		}

		if outcome.ConfidenceBefore != input.Confidence {
			t.Errorf("Expected confidence before %.2f, got %.2f", input.Confidence, outcome.ConfidenceBefore)
		}

		if outcome.ExecutionTimeNs < 0 {
			t.Error("Expected non-negative execution time")
		}

		// Check Thompson state update
		strategyID := outcome.StrategyID
		success := outcome.Success

		if success {
			if mockStorage.alphaIncrements[strategyID] != 1 {
				t.Errorf("Expected alpha increment for %s, got %d", strategyID, mockStorage.alphaIncrements[strategyID])
			}
		} else {
			if mockStorage.betaIncrements[strategyID] != 1 {
				t.Errorf("Expected beta increment for %s, got %d", strategyID, mockStorage.betaIncrements[strategyID])
			}
		}
	}
}

// TestAutoMode_RLFallbackNoStrategies tests that RL gracefully falls back when no strategies available
func TestAutoMode_RLFallbackNoStrategies(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)

	auto := NewAutoMode(linear, tree, divergent)

	// No strategies - RL should gracefully disable
	mockStorage := &mockRLStorage{
		strategies:      []*reinforcement.Strategy{},
		outcomes:        []*reinforcement.Outcome{},
		alphaIncrements: make(map[string]int),
		betaIncrements:  make(map[string]int),
	}

	err := auto.SetRLStorage(mockStorage)
	if err != nil {
		t.Fatalf("SetRLStorage failed: %v", err)
	}

	if auto.rlEnabled {
		t.Error("RL should be disabled when no strategies available")
	}

	// Process thought - should use keyword detection
	input := ThoughtInput{
		Content:    "This is a creative problem",
		Confidence: 0.7,
	}

	result, err := auto.ProcessThought(context.Background(), input)
	if err != nil {
		t.Fatalf("ProcessThought failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// No outcomes should be recorded
	if len(mockStorage.outcomes) != 0 {
		t.Errorf("Expected 0 outcomes when RL disabled, got %d", len(mockStorage.outcomes))
	}
}

// TestAutoMode_RLNoStrategies tests that RL is disabled when no strategies available
func TestAutoMode_RLNoStrategies(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)

	auto := NewAutoMode(linear, tree, divergent)

	mockStorage := &mockRLStorage{
		strategies:      []*reinforcement.Strategy{}, // No strategies
		outcomes:        []*reinforcement.Outcome{},
		alphaIncrements: make(map[string]int),
		betaIncrements:  make(map[string]int),
	}

	err := auto.SetRLStorage(mockStorage)
	if err != nil {
		t.Fatalf("SetRLStorage failed: %v", err)
	}

	// Should be disabled due to no strategies
	if auto.rlEnabled {
		t.Error("RL should be disabled when no strategies available")
	}
}

// TestDetectProblemType tests problem type detection
func TestDetectProblemType(t *testing.T) {
	tests := []struct {
		content      string
		expectedType string
	}{
		{"What is the cause of this effect?", "causal"},
		{"This intervention leads to better outcomes", "causal"},
		{"What is the probability of success?", "probabilistic"},
		{"How likely is this to happen?", "probabilistic"},
		{"If premise A, then conclusion B", "logical"},
		{"This implies the following result", "logical"},
		{"Analyze this general problem", "general"},
		{"", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			result := detectProblemType(tt.content)
			if result != tt.expectedType {
				t.Errorf("Expected type %q, got %q", tt.expectedType, result)
			}
		})
	}
}

// mockRLStorage implements RLStorage interface for testing
type mockRLStorage struct {
	strategies      []*reinforcement.Strategy
	outcomes        []*reinforcement.Outcome
	alphaIncrements map[string]int
	betaIncrements  map[string]int
}

func (m *mockRLStorage) GetAllRLStrategies() ([]*reinforcement.Strategy, error) {
	return m.strategies, nil
}

func (m *mockRLStorage) IncrementThompsonAlpha(strategyID string) error {
	m.alphaIncrements[strategyID]++
	return nil
}

func (m *mockRLStorage) IncrementThompsonBeta(strategyID string) error {
	m.betaIncrements[strategyID]++
	return nil
}

func (m *mockRLStorage) RecordRLOutcome(outcome *reinforcement.Outcome) error {
	m.outcomes = append(m.outcomes, outcome)
	return nil
}
