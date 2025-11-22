package benchmarks

import (
	"fmt"
	"testing"

	"unified-thinking/benchmarks/evaluators"
	"unified-thinking/internal/storage"
)

func TestCausalReasoning(t *testing.T) {
	// Load causal inference suite
	suite, err := LoadSuite("datasets/causal/causal_inference.json")
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
	t.Logf("\n=== Causal Reasoning Benchmark ===")
	t.Logf("Suite: %s", run.SuiteName)
	t.Logf("Total Problems: %d", run.TotalProblems)
	t.Logf("Correct: %d", run.CorrectProblems)
	t.Logf("Accuracy: %.2f%%", run.OverallAccuracy*100)
	t.Logf("ECE (Calibration): %.4f", run.OverallECE)
	t.Logf("Calibration Quality: %s", evaluators.CalibrationQuality(run.OverallECE))
	t.Logf("Avg Latency: %v", run.AvgLatency)

	// Compute calibration details
	calibResults := make([]evaluators.CalibrationResult, len(run.Results))
	for i, r := range run.Results {
		calibResults[i] = evaluators.CalibrationResult{
			Correct:    r.Correct,
			Confidence: r.Confidence,
		}
	}
	calibMetrics := evaluators.ComputeCalibration(calibResults)

	t.Logf("\n=== Calibration Details ===")
	t.Logf("ECE: %.4f", calibMetrics.ECE)
	t.Logf("MCE (Max Error): %.4f", calibMetrics.MCE)
	t.Logf("Brier Score: %.4f", calibMetrics.Brier)
	t.Logf("\nPer-Bucket Calibration:")
	for bucket := 0; bucket < 10; bucket++ {
		if count, exists := calibMetrics.BucketCounts[bucket]; exists && count > 0 {
			confRange := fmt.Sprintf("%.1f-%.1f%%", float64(bucket)*10, float64(bucket+1)*10)
			accuracy := calibMetrics.BucketAccuracies[bucket]
			avgConf := calibMetrics.BucketConfidence[bucket]
			t.Logf("  %s: %d samples, accuracy: %.2f%%, avg confidence: %.2f%%",
				confRange, count, accuracy*100, avgConf*100)
		}
	}

	// Show problem breakdown by type
	typeCorrect := make(map[string]int)
	typeTotal := make(map[string]int)
	for _, problem := range suite.Problems {
		problemType, ok := problem.Metadata["type"].(string)
		if !ok {
			problemType = "unknown"
		}
		typeTotal[problemType]++

		// Find result for this problem
		for _, result := range run.Results {
			if result.ProblemID == problem.ID && result.Correct {
				typeCorrect[problemType]++
				break
			}
		}
	}

	t.Logf("\n=== Performance by Problem Type ===")
	for problemType, total := range typeTotal {
		correct := typeCorrect[problemType]
		accuracy := float64(correct) / float64(total) * 100
		t.Logf("  %s: %d/%d (%.1f%%)", problemType, correct, total, accuracy)
	}

	// Baseline for regression detection
	t.Logf("\n=== Baseline Established ===")
	t.Logf("Causal Reasoning Baseline: %.2f%% accuracy", run.OverallAccuracy*100)
	t.Logf("Run ID: %s", run.RunID)
}
