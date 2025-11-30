package benchmarks

import (
	"testing"

	"unified-thinking/benchmarks/evaluators"
	"unified-thinking/internal/reinforcement"
	"unified-thinking/internal/storage"
)

// TestRLExecutor_Basic tests basic RL executor functionality
func TestRLExecutor_Basic(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockRL := &mockRLStorage{
		strategies:      []*reinforcement.Strategy{},
		outcomes:        []*reinforcement.Outcome{},
		alphaIncrements: make(map[string]int),
		betaIncrements:  make(map[string]int),
	}

	executor := NewRLExecutor(store, mockRL, 0.7)

	if executor == nil {
		t.Fatal("Expected executor, got nil")
	}

	if executor.outcomeThreshold != 0.7 {
		t.Errorf("Expected threshold 0.7, got %.2f", executor.outcomeThreshold)
	}

	if !executor.trackOutcomes {
		t.Error("Expected tracking enabled by default")
	}
}

// TestRLExecutor_Execute tests problem execution with outcome recording
func TestRLExecutor_Execute(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockRL := &mockRLStorage{
		strategies:      []*reinforcement.Strategy{},
		outcomes:        []*reinforcement.Outcome{},
		alphaIncrements: make(map[string]int),
		betaIncrements:  make(map[string]int),
	}

	executor := NewRLExecutor(store, mockRL, 0.7)
	evaluator := evaluators.NewExactMatchEvaluator()

	problem := &Problem{
		ID:          "test_001",
		Description: "Test problem",
		Input: map[string]interface{}{
			"content": "Test content",
			"mode":    "linear",
		},
		Expected: "Test content",
		Category: "logical",
	}

	result, err := executor.Execute(problem, evaluator)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Check outcome was recorded
	if len(mockRL.outcomes) != 1 {
		t.Errorf("Expected 1 outcome, got %d", len(mockRL.outcomes))
	}

	if len(mockRL.outcomes) > 0 {
		outcome := mockRL.outcomes[0]

		if outcome.ProblemID != problem.ID {
			t.Errorf("Expected problem ID %s, got %s", problem.ID, outcome.ProblemID)
		}

		if outcome.ProblemType != problem.Category {
			t.Errorf("Expected category %s, got %s", problem.Category, outcome.ProblemType)
		}

		// Verify Thompson state was updated
		strategyID := outcome.StrategyID
		if outcome.Success {
			if mockRL.alphaIncrements[strategyID] != 1 {
				t.Errorf("Expected alpha increment, got %d", mockRL.alphaIncrements[strategyID])
			}
		} else {
			if mockRL.betaIncrements[strategyID] != 1 {
				t.Errorf("Expected beta increment, got %d", mockRL.betaIncrements[strategyID])
			}
		}
	}
}

// TestRLExecutor_Stats tests strategy statistics tracking
func TestRLExecutor_Stats(t *testing.T) {
	store := storage.NewMemoryStorage()
	mockRL := &mockRLStorage{
		strategies:      []*reinforcement.Strategy{},
		outcomes:        []*reinforcement.Outcome{},
		alphaIncrements: make(map[string]int),
		betaIncrements:  make(map[string]int),
	}

	executor := NewRLExecutor(store, mockRL, 0.7)
	evaluator := evaluators.NewExactMatchEvaluator()

	// Execute multiple problems
	problems := []*Problem{
		{
			ID:       "p1",
			Input:    map[string]interface{}{"content": "Problem 1"},
			Expected: "Problem 1",
			Category: "logical",
		},
		{
			ID:       "p2",
			Input:    map[string]interface{}{"content": "Problem 2"},
			Expected: "Different", // Will fail
			Category: "probabilistic",
		},
		{
			ID:       "p3",
			Input:    map[string]interface{}{"content": "Problem 3"},
			Expected: "Problem 3",
			Category: "causal",
		},
	}

	for _, problem := range problems {
		_, err := executor.Execute(problem, evaluator)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	}

	selections, successes := executor.GetStrategyStats()

	// Should have tracked all executions
	totalSelections := 0
	for _, count := range selections {
		totalSelections += count
	}

	if totalSelections != 3 {
		t.Errorf("Expected 3 selections, got %d", totalSelections)
	}

	// Verify outcomes were recorded
	if len(mockRL.outcomes) != 3 {
		t.Errorf("Expected 3 outcomes, got %d", len(mockRL.outcomes))
	}

	// Verify we have successes tracked
	_ = successes // Used below after reset

	// Test reset
	executor.ResetStats()
	selections, successes = executor.GetStrategyStats()

	if len(selections) != 0 {
		t.Errorf("Expected empty selections after reset, got %d", len(selections))
	}

	if len(successes) != 0 {
		t.Errorf("Expected empty successes after reset, got %d", len(successes))
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
