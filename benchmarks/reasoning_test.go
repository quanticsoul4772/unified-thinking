package benchmarks

import (
	"testing"
	"time"

	"unified-thinking/benchmarks/evaluators"
	"unified-thinking/internal/storage"
)

func TestLogicReasoning(t *testing.T) {
	// Load logic puzzles suite
	suite, err := LoadSuite("datasets/reasoning/logic_puzzles.json")
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
	t.Logf("\n=== Benchmark Results ===")
	t.Logf("Suite: %s", run.SuiteName)
	t.Logf("Total Problems: %d", run.TotalProblems)
	t.Logf("Correct: %d", run.CorrectProblems)
	t.Logf("Accuracy: %.2f%%", run.OverallAccuracy*100)
	t.Logf("ECE (Calibration): %.4f", run.OverallECE)
	t.Logf("Avg Latency: %v", run.AvgLatency)
	t.Logf("\nMode Distribution:")
	for mode, count := range run.ModeDistribution {
		t.Logf("  %s: %d (%.1f%%)", mode, count, float64(count)/float64(run.TotalProblems)*100)
	}

	// Show individual results
	t.Logf("\nIndividual Results:")
	for _, result := range run.Results {
		status := "PASS"
		if !result.Correct {
			status = "FAIL"
		}
		t.Logf("  %s: %s (score: %.2f, confidence: %.2f, latency: %v)",
			result.ProblemID, status, result.Score, result.Confidence, result.Latency)
	}

	// Save results to JSON file
	resultsJSON, err := RunToJSON(run)
	if err != nil {
		t.Logf("Warning: failed to serialize results: %v", err)
	} else {
		t.Logf("\nResults JSON:\n%s", resultsJSON)
	}

	// Check for regressions (baseline TBD after first run)
	// For Phase 1, just report the baseline
	t.Logf("\n=== Baseline Established ===")
	t.Logf("Use this run as baseline for future regression detection")
	t.Logf("Run ID: %s", run.RunID)
}

// TestBenchmarkFramework validates the framework itself using mock executor
func TestBenchmarkFramework(t *testing.T) {
	// Create a simple test suite
	suite := &BenchmarkSuite{
		Name:     "Framework Test",
		Category: "test",
		Problems: []*Problem{
			{
				ID:          "test_001",
				Description: "Test problem 1",
				Input:       map[string]interface{}{"mode": "linear"},
				Expected:    "correct answer",
			},
			{
				ID:          "test_002",
				Description: "Test problem 2",
				Input:       map[string]interface{}{"mode": "linear"},
				Expected:    "another answer",
			},
		},
	}

	// Create mock executor with predefined results
	executor := NewMockExecutor()
	executor.AddResult("test_001", &Result{
		ProblemID:  "test_001",
		Correct:    true,
		Score:      1.0,
		Confidence: 0.9,
		Latency:    5 * time.Millisecond,
		Mode:       "linear",
		Response:   "correct answer",
	})
	executor.AddResult("test_002", &Result{
		ProblemID:  "test_002",
		Correct:    false,
		Score:      0.0,
		Confidence: 0.6,
		Latency:    8 * time.Millisecond,
		Mode:       "linear",
		Response:   "wrong answer",
	})

	evaluator := evaluators.NewExactMatchEvaluator()

	// Run suite
	run, err := RunSuite(suite, evaluator, executor)
	if err != nil {
		t.Fatalf("Failed to run suite: %v", err)
	}

	// Validate metrics computation
	expectedAccuracy := 0.5 // 1 out of 2 correct
	if run.OverallAccuracy != expectedAccuracy {
		t.Errorf("Expected accuracy %.2f, got %.2f", expectedAccuracy, run.OverallAccuracy)
	}

	if run.TotalProblems != 2 {
		t.Errorf("Expected 2 total problems, got %d", run.TotalProblems)
	}

	if run.CorrectProblems != 1 {
		t.Errorf("Expected 1 correct problem, got %d", run.CorrectProblems)
	}

	t.Logf("Framework validation passed: accuracy=%.2f%%, ECE=%.4f",
		run.OverallAccuracy*100, run.OverallECE)
}
