package reinforcement

import (
	"fmt"
	"math"
	"testing"
)

func TestThompsonSelector_AddStrategy(t *testing.T) {
	ts := NewThompsonSelector(42)

	strategy := &Strategy{
		ID:       "test_strategy",
		Name:     "Test Strategy",
		Mode:     "linear",
		IsActive: true,
	}

	ts.AddStrategy(strategy)

	// Verify uniform prior was set
	retrieved, err := ts.GetStrategy("test_strategy")
	if err != nil {
		t.Fatalf("Failed to get strategy: %v", err)
	}

	if retrieved.Alpha != 1.0 {
		t.Errorf("Alpha = %v, want 1.0 (uniform prior)", retrieved.Alpha)
	}

	if retrieved.Beta != 1.0 {
		t.Errorf("Beta = %v, want 1.0 (uniform prior)", retrieved.Beta)
	}
}

func TestThompsonSelector_SelectStrategy(t *testing.T) {
	ts := NewThompsonSelector(42)

	// Add two strategies
	ts.AddStrategy(&Strategy{
		ID:       "strategy_1",
		Name:     "Strategy 1",
		Mode:     "linear",
		IsActive: true,
		Alpha:    1.0,
		Beta:     1.0,
	})

	ts.AddStrategy(&Strategy{
		ID:       "strategy_2",
		Name:     "Strategy 2",
		Mode:     "tree",
		IsActive: true,
		Alpha:    1.0,
		Beta:     1.0,
	})

	// Select strategy
	ctx := ProblemContext{Description: "test problem"}
	selected, err := ts.SelectStrategy(ctx)
	if err != nil {
		t.Fatalf("SelectStrategy failed: %v", err)
	}

	if selected == nil {
		t.Fatal("Expected strategy, got nil")
	}

	// Should select one of the two strategies
	if selected.ID != "strategy_1" && selected.ID != "strategy_2" {
		t.Errorf("Selected unexpected strategy: %s", selected.ID)
	}
}

func TestThompsonSelector_RecordOutcome(t *testing.T) {
	ts := NewThompsonSelector(42)

	ts.AddStrategy(&Strategy{
		ID:       "test",
		Name:     "Test",
		Mode:     "linear",
		IsActive: true,
	})

	// Record success
	err := ts.RecordOutcome("test", true)
	if err != nil {
		t.Fatalf("RecordOutcome failed: %v", err)
	}

	strategy, _ := ts.GetStrategy("test")
	if strategy.Alpha != 2.0 {
		t.Errorf("After success, Alpha = %v, want 2.0", strategy.Alpha)
	}
	if strategy.TotalSuccesses != 1 {
		t.Errorf("TotalSuccesses = %d, want 1", strategy.TotalSuccesses)
	}
	if strategy.TotalTrials != 1 {
		t.Errorf("TotalTrials = %d, want 1", strategy.TotalTrials)
	}

	// Record failure
	err = ts.RecordOutcome("test", false)
	if err != nil {
		t.Fatalf("RecordOutcome failed: %v", err)
	}

	strategy, _ = ts.GetStrategy("test")
	if strategy.Beta != 2.0 {
		t.Errorf("After failure, Beta = %v, want 2.0", strategy.Beta)
	}
	if strategy.TotalTrials != 2 {
		t.Errorf("TotalTrials = %d, want 2", strategy.TotalTrials)
	}
}

func TestThompsonSelector_ExplorationExploitation(t *testing.T) {
	ts := NewThompsonSelector(42)

	// Add strategy with many successes (should be exploited)
	ts.AddStrategy(&Strategy{
		ID:       "good",
		Name:     "Good Strategy",
		Mode:     "linear",
		IsActive: true,
		Alpha:    20.0, // 19 successes
		Beta:     2.0,  // 1 failure
	})

	// Add strategy with no data (should be explored)
	ts.AddStrategy(&Strategy{
		ID:       "unknown",
		Name:     "Unknown Strategy",
		Mode:     "tree",
		IsActive: true,
		Alpha:    1.0,
		Beta:     1.0,
	})

	// Run many selections
	selections := make(map[string]int)
	ctx := ProblemContext{Description: "test"}

	for i := 0; i < 1000; i++ {
		selected, _ := ts.SelectStrategy(ctx)
		selections[selected.ID]++
	}

	// Good strategy should be selected more often (but not always)
	goodCount := selections["good"]
	unknownCount := selections["unknown"]

	if goodCount < 700 {
		t.Errorf("Good strategy selected only %d/1000 times, expected >700 (exploitation)", goodCount)
	}

	if unknownCount < 50 {
		t.Errorf("Unknown strategy selected only %d/1000 times, expected >50 (exploration)", unknownCount)
	}

	t.Logf("Selection distribution: good=%d (%.1f%%), unknown=%d (%.1f%%)",
		goodCount, float64(goodCount)/10.0, unknownCount, float64(unknownCount)/10.0)
}

func TestThompsonSelector_Learning(t *testing.T) {
	ts := NewThompsonSelector(42)

	ts.AddStrategy(&Strategy{
		ID:       "strategy_a",
		Name:     "Strategy A",
		Mode:     "linear",
		IsActive: true,
	})

	ts.AddStrategy(&Strategy{
		ID:       "strategy_b",
		Name:     "Strategy B",
		Mode:     "tree",
		IsActive: true,
	})

	// Simulate: Strategy A is actually better (80% success rate)
	// Strategy B is worse (20% success rate)
	ctx := ProblemContext{Description: "test"}

	for i := 0; i < 100; i++ {
		selected, _ := ts.SelectStrategy(ctx)

		// Simulate outcome based on true success rates
		var success bool
		if selected.ID == "strategy_a" {
			success = ts.rng.Float64() < 0.8 // 80% success
		} else {
			success = ts.rng.Float64() < 0.2 // 20% success
		}

		ts.RecordOutcome(selected.ID, success)
	}

	// After learning, strategy_a should have higher success rate
	stratA, _ := ts.GetStrategy("strategy_a")
	stratB, _ := ts.GetStrategy("strategy_b")

	rateA := stratA.SuccessRate()
	rateB := stratB.SuccessRate()

	if rateA <= rateB {
		t.Errorf("After learning, strategy_a rate (%v) should be > strategy_b rate (%v)",
			rateA, rateB)
	}

	t.Logf("After 100 trials:")
	t.Logf("  Strategy A: %d trials, %.1f%% success", stratA.TotalTrials, rateA*100)
	t.Logf("  Strategy B: %d trials, %.1f%% success", stratB.TotalTrials, rateB*100)
}

func TestThompsonSelector_GetBestStrategy(t *testing.T) {
	ts := NewThompsonSelector(42)

	ts.AddStrategy(&Strategy{
		ID:             "weak",
		Name:           "Weak",
		IsActive:       true,
		TotalTrials:    10,
		TotalSuccesses: 3, // 30%
	})

	ts.AddStrategy(&Strategy{
		ID:             "strong",
		Name:           "Strong",
		IsActive:       true,
		TotalTrials:    10,
		TotalSuccesses: 8, // 80%
	})

	best := ts.GetBestStrategy()
	if best.ID != "strong" {
		t.Errorf("GetBestStrategy() = %s, want 'strong'", best.ID)
	}
}

func TestThompsonSelector_GetStrategyDistribution(t *testing.T) {
	ts := NewThompsonSelector(42)

	ts.AddStrategy(&Strategy{
		ID:       "strategy_1",
		Name:     "Strategy 1",
		IsActive: true,
		Alpha:    10.0,
		Beta:     2.0,
	})

	ts.AddStrategy(&Strategy{
		ID:       "strategy_2",
		Name:     "Strategy 2",
		IsActive: true,
		Alpha:    2.0,
		Beta:     10.0,
	})

	dist := ts.GetStrategyDistribution(10000)

	// Strategy 1 should be selected more often
	prob1 := dist["strategy_1"]
	prob2 := dist["strategy_2"]

	if prob1 <= prob2 {
		t.Errorf("Strategy 1 probability (%v) should be > Strategy 2 (%v)", prob1, prob2)
	}

	// Probabilities should sum to ~1.0
	total := prob1 + prob2
	if math.Abs(total-1.0) > 0.01 {
		t.Errorf("Total probability = %v, want ~1.0", total)
	}

	t.Logf("Distribution: strategy_1=%.1f%%, strategy_2=%.1f%%", prob1*100, prob2*100)
}

func TestThompsonSelector_ResetStrategy(t *testing.T) {
	ts := NewThompsonSelector(42)

	ts.AddStrategy(&Strategy{
		ID:             "test",
		Name:           "Test",
		IsActive:       true,
		Alpha:          10.0,
		Beta:           5.0,
		TotalTrials:    14,
		TotalSuccesses: 9,
	})

	// Reset
	err := ts.ResetStrategy("test")
	if err != nil {
		t.Fatalf("ResetStrategy failed: %v", err)
	}

	strategy, _ := ts.GetStrategy("test")
	if strategy.Alpha != 1.0 {
		t.Errorf("After reset, Alpha = %v, want 1.0", strategy.Alpha)
	}
	if strategy.Beta != 1.0 {
		t.Errorf("After reset, Beta = %v, want 1.0", strategy.Beta)
	}
	if strategy.TotalTrials != 0 {
		t.Errorf("After reset, TotalTrials = %d, want 0", strategy.TotalTrials)
	}
}

func TestThompsonSelector_NoStrategies(t *testing.T) {
	ts := NewThompsonSelector(42)

	ctx := ProblemContext{Description: "test"}
	_, err := ts.SelectStrategy(ctx)

	if err == nil {
		t.Error("SelectStrategy with no strategies should return error")
	}
}

func TestThompsonSelector_AllInactive(t *testing.T) {
	ts := NewThompsonSelector(42)

	ts.AddStrategy(&Strategy{
		ID:       "inactive",
		Name:     "Inactive",
		IsActive: false,
	})

	ctx := ProblemContext{Description: "test"}
	_, err := ts.SelectStrategy(ctx)

	if err == nil {
		t.Error("SelectStrategy with all inactive strategies should return error")
	}
}

// Benchmark Thompson selection
func BenchmarkThompsonSelect(b *testing.B) {
	ts := NewThompsonSelector(42)

	for i := 0; i < 5; i++ {
		ts.AddStrategy(&Strategy{
			ID:       fmt.Sprintf("strategy_%d", i),
			Name:     fmt.Sprintf("Strategy %d", i),
			IsActive: true,
			Alpha:    float64(i + 1),
			Beta:     float64(6 - i),
		})
	}

	ctx := ProblemContext{Description: "test"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ts.SelectStrategy(ctx)
	}
}
