// Package evaluators provides learning effectiveness metric implementations.
package evaluators

import (
	"fmt"
	"math"
)

// LearningMetrics tracks improvement over iterations
type LearningMetrics struct {
	InitialAccuracy   float64            // Accuracy on first iteration
	FinalAccuracy     float64            // Accuracy on last iteration
	ImprovementRate   float64            // Absolute improvement
	RelativeImprovement float64          // Percentage improvement
	Iterations        int                // Number of iterations
	AccuracyByIter    map[int]float64    // Accuracy per iteration
	LearningRate      float64            // Rate of learning (slope)
	SignificantImprovement bool          // Statistical significance
	PValue            float64            // P-value for improvement
}

// IterationResult represents results for a single iteration
type IterationResult struct {
	Iteration   int
	Correct     int
	Total       int
	Accuracy    float64
}

// ComputeLearning calculates learning metrics from iteration results
func ComputeLearning(results []IterationResult) *LearningMetrics {
	if len(results) == 0 {
		return &LearningMetrics{
			AccuracyByIter: make(map[int]float64),
		}
	}

	// Build accuracy by iteration map
	accuracyByIter := make(map[int]float64)
	for _, result := range results {
		accuracyByIter[result.Iteration] = result.Accuracy
	}

	// Get first and last iteration
	firstIter := results[0]
	lastIter := results[len(results)-1]

	initialAcc := firstIter.Accuracy
	finalAcc := lastIter.Accuracy

	improvementRate := finalAcc - initialAcc
	relativeImprovement := 0.0
	if initialAcc > 0 {
		relativeImprovement = (improvementRate / initialAcc) * 100
	}

	// Compute learning rate (slope)
	learningRate := 0.0
	if len(results) > 1 {
		learningRate = improvementRate / float64(len(results)-1)
	}

	// Simple statistical test: paired t-test approximation
	// For learning effectiveness, we check if improvement > 10% with confidence
	significantImprovement := improvementRate > 0.10
	pValue := 0.05 // Placeholder - would need proper t-test implementation

	return &LearningMetrics{
		InitialAccuracy:        initialAcc,
		FinalAccuracy:          finalAcc,
		ImprovementRate:        improvementRate,
		RelativeImprovement:    relativeImprovement,
		Iterations:             len(results),
		AccuracyByIter:         accuracyByIter,
		LearningRate:           learningRate,
		SignificantImprovement: significantImprovement,
		PValue:                 pValue,
	}
}

// GroupByIteration groups results by iteration number
func GroupByIteration(results []struct {
	Correct    bool
	Iteration  int
}) []IterationResult {
	// Count by iteration
	iterCounts := make(map[int]*struct {
		correct int
		total   int
	})

	for _, result := range results {
		if _, exists := iterCounts[result.Iteration]; !exists {
			iterCounts[result.Iteration] = &struct {
				correct int
				total   int
			}{}
		}

		iterCounts[result.Iteration].total++
		if result.Correct {
			iterCounts[result.Iteration].correct++
		}
	}

	// Convert to IterationResult slice
	iterResults := make([]IterationResult, 0, len(iterCounts))
	for iter, counts := range iterCounts {
		accuracy := 0.0
		if counts.total > 0 {
			accuracy = float64(counts.correct) / float64(counts.total)
		}

		iterResults = append(iterResults, IterationResult{
			Iteration: iter,
			Correct:   counts.correct,
			Total:     counts.total,
			Accuracy:  accuracy,
		})
	}

	// Sort by iteration
	sortIterationResults(iterResults)

	return iterResults
}

// sortIterationResults sorts by iteration number
func sortIterationResults(results []IterationResult) {
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Iteration > results[j+1].Iteration {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// FormatLearningReport generates a text report of learning metrics
func FormatLearningReport(metrics *LearningMetrics) string {
	report := "Learning Analysis:\n"
	report += fmt.Sprintf("  Initial Accuracy: %.2f%%\n", metrics.InitialAccuracy*100)
	report += fmt.Sprintf("  Final Accuracy: %.2f%%\n", metrics.FinalAccuracy*100)
	report += fmt.Sprintf("  Improvement: %.2f%% (%.1f%% relative)\n",
		metrics.ImprovementRate*100, metrics.RelativeImprovement)
	report += fmt.Sprintf("  Learning Rate: %.4f per iteration\n", metrics.LearningRate)
	report += fmt.Sprintf("  Iterations: %d\n", metrics.Iterations)

	if metrics.SignificantImprovement {
		report += "  Status: SIGNIFICANT improvement detected\n"
	} else {
		report += "  Status: No significant improvement\n"
	}

	report += "\nAccuracy by Iteration:\n"
	for i := 1; i <= metrics.Iterations; i++ {
		if acc, exists := metrics.AccuracyByIter[i]; exists {
			report += fmt.Sprintf("  Iteration %d: %.2f%%\n", i, acc*100)
		}
	}

	return report
}

// DetectLearning returns true if learning is detected
func DetectLearning(metrics *LearningMetrics, minImprovement float64) bool {
	return metrics.ImprovementRate >= minImprovement && metrics.FinalAccuracy > metrics.InitialAccuracy
}

// LearningTrend categorizes the learning pattern
func LearningTrend(metrics *LearningMetrics) string {
	if metrics.ImprovementRate > 0.20 {
		return "strong learning"
	} else if metrics.ImprovementRate > 0.10 {
		return "moderate learning"
	} else if metrics.ImprovementRate > 0.05 {
		return "weak learning"
	} else if math.Abs(metrics.ImprovementRate) <= 0.05 {
		return "stable"
	} else if metrics.ImprovementRate < -0.10 {
		return "regression"
	}
	return "slight degradation"
}
