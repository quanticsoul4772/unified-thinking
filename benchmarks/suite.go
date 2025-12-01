// Package benchmarks provides the benchmark suite runner.
package benchmarks

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LoadSuite loads a benchmark suite from a JSON file
func LoadSuite(path string) (*BenchmarkSuite, error) {
	// Sanitize path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Validate path contains only safe characters and no traversal
	if filepath.IsAbs(cleanPath) {
		return nil, fmt.Errorf("absolute paths not allowed")
	}
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("path traversal not allowed")
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read suite file: %w", err)
	}

	var suite BenchmarkSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse suite JSON: %w", err)
	}

	// Populate problem.Category from suite.Category
	for _, problem := range suite.Problems {
		if problem.Category == "" {
			problem.Category = suite.Category
		}
	}

	return &suite, nil
}

// RunSuite executes all problems in a suite
func RunSuite(suite *BenchmarkSuite, evaluator Evaluator, executor ProblemExecutor) (*BenchmarkRun, error) {
	run := &BenchmarkRun{
		RunID:            fmt.Sprintf("run_%d", time.Now().Unix()),
		SuiteName:        suite.Name,
		Timestamp:        time.Now(),
		Results:          make([]*Result, 0, len(suite.Problems)),
		ModeDistribution: make(map[string]int),
	}

	// Check if executor is RL-aware
	rlExecutor, isRLExecutor := executor.(*RLExecutor)
	if isRLExecutor {
		run.RLEnabled = true
		run.StrategyDistribution = make(map[string]int)
		run.StrategySuccessRate = make(map[string]float64)
		run.LearningCurve = make([]float64, 0)
		run.RLMetadata = make(map[string]interface{})
	}

	for _, problem := range suite.Problems {
		result, err := executor.Execute(problem, evaluator)
		if err != nil {
			result = &Result{
				ProblemID: problem.ID,
				Correct:   false,
				Score:     0.0,
				Error:     err.Error(),
			}
		}

		run.Results = append(run.Results, result)
		run.TotalProblems++
		if result.Correct {
			run.CorrectProblems++
		}

		if result.Mode != "" {
			run.ModeDistribution[result.Mode]++
		}

		// Track RL-specific metrics
		if isRLExecutor && result.Metadata != nil {
			if strategyID, ok := result.Metadata["strategy_id"].(string); ok {
				run.StrategyDistribution[strategyID]++
			}
		}
	}

	// Compute overall metrics
	run.OverallAccuracy = float64(run.CorrectProblems) / float64(run.TotalProblems)
	run.OverallECE = computeECE(run.Results)
	run.AvgLatency = computeAvgLatency(run.Results)

	// Compute RL metrics if applicable
	if isRLExecutor {
		computeRLMetrics(run, rlExecutor)
	}

	return run, nil
}

// ProblemExecutor executes a single problem
type ProblemExecutor interface {
	Execute(problem *Problem, evaluator Evaluator) (*Result, error)
}

// computeECE calculates Expected Calibration Error
func computeECE(results []*Result) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Bucket results by confidence (10 buckets: 0-0.1, 0.1-0.2, ..., 0.9-1.0)
	buckets := make([]struct {
		totalCount    int
		correctCount  int
		avgConfidence float64
		sumConfidence float64
	}, 10)

	for _, result := range results {
		if result.Confidence < 0 || result.Confidence > 1 {
			continue // Skip invalid confidence values
		}

		bucketIdx := int(result.Confidence * 10)
		if bucketIdx >= 10 {
			bucketIdx = 9
		}

		buckets[bucketIdx].totalCount++
		if result.Correct {
			buckets[bucketIdx].correctCount++
		}
		buckets[bucketIdx].sumConfidence += result.Confidence
	}

	// Calculate ECE
	totalCount := len(results)
	ece := 0.0

	for _, bucket := range buckets {
		if bucket.totalCount == 0 {
			continue
		}

		bucketAccuracy := float64(bucket.correctCount) / float64(bucket.totalCount)
		avgConfidence := bucket.sumConfidence / float64(bucket.totalCount)
		weight := float64(bucket.totalCount) / float64(totalCount)

		ece += weight * abs(bucketAccuracy-avgConfidence)
	}

	return ece
}

// computeAvgLatency calculates average latency
func computeAvgLatency(results []*Result) time.Duration {
	if len(results) == 0 {
		return 0
	}

	var total time.Duration
	for _, result := range results {
		total += result.Latency
	}

	return total / time.Duration(len(results))
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// computeRLMetrics calculates Thompson Sampling RL metrics
func computeRLMetrics(run *BenchmarkRun, rlExecutor *RLExecutor) {
	if len(run.Results) == 0 {
		return
	}

	// Get strategy stats from executor
	selections, successes := rlExecutor.GetStrategyStats()

	// Compute strategy success rates
	for strategyID, count := range selections {
		successCount := successes[strategyID]
		run.StrategySuccessRate[strategyID] = float64(successCount) / float64(count)
	}

	// Compute initial vs final accuracy (first 20% vs last 20%)
	numProblems := len(run.Results)
	initialWindow := max(1, numProblems/5) // 20%
	finalWindow := max(1, numProblems/5)

	// Initial accuracy (first 20%)
	initialCorrect := 0
	for i := 0; i < initialWindow && i < numProblems; i++ {
		if run.Results[i].Correct {
			initialCorrect++
		}
	}
	run.InitialAccuracy = float64(initialCorrect) / float64(initialWindow)

	// Final accuracy (last 20%)
	finalCorrect := 0
	startIdx := max(0, numProblems-finalWindow)
	for i := startIdx; i < numProblems; i++ {
		if run.Results[i].Correct {
			finalCorrect++
		}
	}
	run.FinalAccuracy = float64(finalCorrect) / float64(finalWindow)

	// Accuracy improvement
	run.AccuracyImprovement = run.FinalAccuracy - run.InitialAccuracy

	// Compute learning curve (rolling window accuracy)
	windowSize := max(5, numProblems/10) // 10% window or min 5
	for i := 0; i <= numProblems-windowSize; i++ {
		windowCorrect := 0
		for j := i; j < i+windowSize; j++ {
			if run.Results[j].Correct {
				windowCorrect++
			}
		}
		windowAccuracy := float64(windowCorrect) / float64(windowSize)
		run.LearningCurve = append(run.LearningCurve, windowAccuracy)
	}

	// Compute exploration rate (how often non-greedy strategy was chosen)
	// Find strategy with highest success rate (greedy choice)
	var greedyStrategy string
	var maxSuccessRate float64
	for strategyID, rate := range run.StrategySuccessRate {
		if rate > maxSuccessRate {
			maxSuccessRate = rate
			greedyStrategy = strategyID
		}
	}

	if greedyStrategy != "" {
		greedyCount := selections[greedyStrategy]
		totalSelections := 0
		for _, count := range selections {
			totalSelections += count
		}
		if totalSelections > 0 {
			run.ExplorationRate = float64(totalSelections-greedyCount) / float64(totalSelections)
		}
	}

	// Compute strategy diversity (Shannon entropy)
	totalSelections := 0
	for _, count := range selections {
		totalSelections += count
	}

	if totalSelections > 0 {
		entropy := 0.0
		for _, count := range selections {
			if count > 0 {
				p := float64(count) / float64(totalSelections)
				entropy -= p * log2(p)
			}
		}
		run.StrategyDiversity = entropy
	}

	// Store additional metadata
	run.RLMetadata["total_strategy_selections"] = totalSelections
	run.RLMetadata["unique_strategies_used"] = len(selections)
	run.RLMetadata["greedy_strategy"] = greedyStrategy
	run.RLMetadata["max_success_rate"] = maxSuccessRate
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func log2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Log2(x)
}
