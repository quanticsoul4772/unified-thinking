package storage

import (
	"os"
	"testing"
	"time"

	"unified-thinking/internal/reinforcement"
)

func TestRLStorage_StoreAndGetStrategy(t *testing.T) {
	dbPath := "test_rl.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Store strategy
	strategy := &reinforcement.Strategy{
		ID:          "test_strategy",
		Name:        "Test Strategy",
		Description: "Test description",
		Mode:        "linear",
		Parameters:  map[string]interface{}{"param1": "value1"},
		IsActive:    true,
	}

	err = store.StoreRLStrategy(strategy)
	if err != nil {
		t.Fatalf("StoreRLStrategy failed: %v", err)
	}

	// Retrieve strategy
	retrieved, err := store.GetRLStrategy("test_strategy")
	if err != nil {
		t.Fatalf("GetRLStrategy failed: %v", err)
	}

	if retrieved.ID != strategy.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, strategy.ID)
	}

	if retrieved.Name != strategy.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, strategy.Name)
	}

	if retrieved.Mode != strategy.Mode {
		t.Errorf("Mode = %v, want %v", retrieved.Mode, strategy.Mode)
	}
}

func TestRLStorage_GetAllStrategies(t *testing.T) {
	dbPath := "test_rl_all.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Migration should have created 5 default strategies
	strategies, err := store.GetAllRLStrategies()
	if err != nil {
		t.Fatalf("GetAllRLStrategies failed: %v", err)
	}

	expectedCount := 5
	if len(strategies) != expectedCount {
		t.Errorf("Got %d strategies, want %d", len(strategies), expectedCount)
	}

	// Verify all have uniform priors
	for _, s := range strategies {
		if s.Alpha != 1.0 {
			t.Errorf("Strategy %s alpha = %v, want 1.0", s.ID, s.Alpha)
		}
		if s.Beta != 1.0 {
			t.Errorf("Strategy %s beta = %v, want 1.0", s.ID, s.Beta)
		}
	}

	t.Logf("Found %d default strategies", len(strategies))
}

func TestRLStorage_RecordOutcome(t *testing.T) {
	dbPath := "test_rl_outcome.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	outcome := &reinforcement.Outcome{
		StrategyID:         "strategy_linear",
		ProblemID:          "test_problem",
		ProblemType:        "logical",
		ProblemDescription: "Test problem description",
		Success:            true,
		ConfidenceBefore:   0.8,
		ConfidenceAfter:    0.9,
		ExecutionTimeNs:    1000000,
		TokenCount:         25,
		Timestamp:          time.Now().Unix(),
		ReasoningPath:      map[string]interface{}{"step1": "analysis"},
		Metadata:           map[string]interface{}{"test": true},
	}

	err = store.RecordRLOutcome(outcome)
	if err != nil {
		t.Fatalf("RecordRLOutcome failed: %v", err)
	}

	t.Log("Outcome recorded successfully")
}

func TestRLStorage_IncrementThompsonState(t *testing.T) {
	dbPath := "test_rl_increment.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	strategyID := "strategy_linear"

	// Get initial state
	initial, err := store.GetRLStrategy(strategyID)
	if err != nil {
		t.Fatalf("GetRLStrategy failed: %v", err)
	}

	initialAlpha := initial.Alpha
	initialBeta := initial.Beta

	// Increment alpha (success)
	err = store.IncrementThompsonAlpha(strategyID)
	if err != nil {
		t.Fatalf("IncrementThompsonAlpha failed: %v", err)
	}

	// Verify increment
	updated, err := store.GetRLStrategy(strategyID)
	if err != nil {
		t.Fatalf("GetRLStrategy failed: %v", err)
	}

	if updated.Alpha != initialAlpha+1 {
		t.Errorf("Alpha = %v, want %v", updated.Alpha, initialAlpha+1)
	}

	if updated.TotalSuccesses != 1 {
		t.Errorf("TotalSuccesses = %d, want 1", updated.TotalSuccesses)
	}

	if updated.TotalTrials != 1 {
		t.Errorf("TotalTrials = %d, want 1", updated.TotalTrials)
	}

	// Increment beta (failure)
	err = store.IncrementThompsonBeta(strategyID)
	if err != nil {
		t.Fatalf("IncrementThompsonBeta failed: %v", err)
	}

	// Verify increment
	updated, err = store.GetRLStrategy(strategyID)
	if err != nil {
		t.Fatalf("GetRLStrategy failed: %v", err)
	}

	if updated.Beta != initialBeta+1 {
		t.Errorf("Beta = %v, want %v", updated.Beta, initialBeta+1)
	}

	if updated.TotalTrials != 2 {
		t.Errorf("TotalTrials = %d, want 2", updated.TotalTrials)
	}

	successRate := updated.SuccessRate()
	if successRate != 0.5 {
		t.Errorf("SuccessRate = %v, want 0.5 (1/2)", successRate)
	}
}

func TestRLStorage_GetStrategyPerformance(t *testing.T) {
	dbPath := "test_rl_perf.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Record some outcomes
	store.IncrementThompsonAlpha("strategy_linear") // 1 success
	store.IncrementThompsonAlpha("strategy_linear") // 2 successes
	store.IncrementThompsonBeta("strategy_linear")  // 1 failure
	store.IncrementThompsonAlpha("strategy_tree")   // 1 success
	store.IncrementThompsonBeta("strategy_tree")    // 1 failure
	store.IncrementThompsonBeta("strategy_tree")    // 2 failures

	// Get performance
	perf, err := store.GetRLStrategyPerformance()
	if err != nil {
		t.Fatalf("GetRLStrategyPerformance failed: %v", err)
	}

	// Should be sorted by success rate (descending)
	// Linear: 2/3 = 66.7%, Tree: 1/3 = 33.3%
	if len(perf) < 2 {
		t.Fatalf("Expected at least 2 strategies, got %d", len(perf))
	}

	// First should be linear (highest success rate)
	if perf[0].ID != "strategy_linear" {
		t.Errorf("First strategy = %s, want strategy_linear", perf[0].ID)
	}

	linearRate := perf[0].SuccessRate()
	if linearRate < 0.66 || linearRate > 0.67 {
		t.Errorf("Linear success rate = %v, want ~0.667", linearRate)
	}

	t.Logf("Performance sorted correctly: %s (%.1f%%) > %s (%.1f%%)",
		perf[0].ID, perf[0].SuccessRate()*100,
		perf[1].ID, perf[1].SuccessRate()*100)
}
