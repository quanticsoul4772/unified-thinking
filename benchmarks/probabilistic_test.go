package benchmarks

import (
	"testing"

	"unified-thinking/benchmarks/evaluators"
	"unified-thinking/internal/storage"
)

func TestProbabilisticReasoning(t *testing.T) {
	// Load Bayesian problems suite
	suite, err := LoadSuite("datasets/probabilistic/bayes_problems.json")
	if err != nil {
		t.Fatalf("Failed to load suite: %v", err)
	}

	t.Logf("Loaded suite: %s with %d problems", suite.Name, len(suite.Problems))

	// Create storage and executor
	store := storage.NewMemoryStorage()
	executor := NewDirectExecutor(store)
	evaluator := evaluators.NewContainsEvaluator()

	// Run benchmark suite
	run, err := RunSuite(suite, evaluator, executor)
	if err != nil {
		t.Fatalf("Failed to run suite: %v", err)
	}

	// Report results
	t.Logf("\n=== Probabilistic Reasoning Benchmark ===")
	t.Logf("Suite: %s", run.SuiteName)
	t.Logf("Total Problems: %d", run.TotalProblems)
	t.Logf("Correct: %d", run.CorrectProblems)
	t.Logf("Accuracy: %.2f%%", run.OverallAccuracy*100)
	t.Logf("ECE (Calibration): %.4f", run.OverallECE)
	t.Logf("Calibration Quality: %s", evaluators.CalibrationQuality(run.OverallECE))
	t.Logf("Avg Latency: %v", run.AvgLatency)

	// Compute efficiency metrics
	latencyResults := make([]evaluators.LatencyResult, len(run.Results))
	for i, r := range run.Results {
		latencyResults[i] = evaluators.LatencyResult{
			ProblemID: r.ProblemID,
			Latency:   r.Latency,
			Tokens:    r.Tokens,
		}
	}
	efficiency := evaluators.ComputeEfficiency(latencyResults)

	t.Logf("\n=== Efficiency Metrics ===")
	t.Logf("Total Latency: %v", efficiency.TotalLatency)
	t.Logf("Avg Latency: %v (%s)", efficiency.AvgLatency, evaluators.EfficiencyCategory(efficiency.AvgLatency))
	t.Logf("P50 Latency: %v", efficiency.P50Latency)
	t.Logf("P95 Latency: %v", efficiency.P95Latency)
	t.Logf("P99 Latency: %v", efficiency.P99Latency)
	t.Logf("Throughput: %.2f problems/sec", efficiency.ProblemsPerSec)
	t.Logf("Total Tokens: %d", efficiency.TotalTokens)
	t.Logf("Avg Tokens: %.1f", efficiency.AvgTokens)

	// Show individual results for failed problems
	t.Logf("\nFailed Problems:")
	for _, result := range run.Results {
		if !result.Correct {
			t.Logf("  %s: FAIL (confidence: %.2f)", result.ProblemID, result.Confidence)
		}
	}

	// Baseline for regression detection
	t.Logf("\n=== Baseline Established ===")
	t.Logf("Probabilistic Reasoning Baseline: %.2f%% accuracy", run.OverallAccuracy*100)
	t.Logf("Run ID: %s", run.RunID)
}
