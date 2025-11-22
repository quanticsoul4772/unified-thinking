package benchmarks

import (
	"testing"

	"unified-thinking/benchmarks/evaluators"
	"unified-thinking/internal/storage"
)

func TestLearningEffectiveness(t *testing.T) {
	// Load repeated problems suite
	suite, err := LoadSuite("datasets/learning/repeated_problems.json")
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

	// Group results by iteration
	type iterResult struct {
		Correct   bool
		Iteration int
	}

	iterResults := make([]iterResult, 0, len(run.Results))
	for i, result := range run.Results {
		problem := suite.Problems[i]
		iteration := 1
		if iter, ok := problem.Metadata["iteration"].(float64); ok {
			iteration = int(iter)
		}

		iterResults = append(iterResults, iterResult{
			Correct:   result.Correct,
			Iteration: iteration,
		})
	}

	// Convert to evaluators format
	evalIterResults := make([]struct {
		Correct   bool
		Iteration int
	}, len(iterResults))
	for i, r := range iterResults {
		evalIterResults[i] = struct {
			Correct   bool
			Iteration int
		}{
			Correct:   r.Correct,
			Iteration: r.Iteration,
		}
	}

	grouped := evaluators.GroupByIteration(evalIterResults)
	learningMetrics := evaluators.ComputeLearning(grouped)

	// Report results
	t.Logf("\n=== Learning Effectiveness Benchmark ===")
	t.Logf("Suite: %s", run.SuiteName)
	t.Logf("Total Problems: %d", run.TotalProblems)
	t.Logf("Correct: %d", run.CorrectProblems)
	t.Logf("Overall Accuracy: %.2f%%", run.OverallAccuracy*100)

	t.Logf("\n=== Learning Analysis ===")
	t.Log(evaluators.FormatLearningReport(learningMetrics))

	t.Logf("Learning Trend: %s", evaluators.LearningTrend(learningMetrics))

	// Check for learning
	hasLearning := evaluators.DetectLearning(learningMetrics, 0.10)
	if hasLearning {
		t.Logf("\n✓ LEARNING DETECTED: System improved by %.1f%% over %d iterations",
			learningMetrics.ImprovementRate*100, learningMetrics.Iterations)
	} else {
		t.Logf("\n✗ NO SIGNIFICANT LEARNING: Improvement %.1f%% (threshold: 10%%)",
			learningMetrics.ImprovementRate*100)
	}

	// Baseline
	t.Logf("\n=== Baseline Established ===")
	t.Logf("Learning Test Baseline: %.2f%% → %.2f%% over %d iterations",
		learningMetrics.InitialAccuracy*100,
		learningMetrics.FinalAccuracy*100,
		learningMetrics.Iterations)
	t.Logf("Run ID: %s", run.RunID)
}
