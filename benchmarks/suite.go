// Package benchmarks provides the benchmark suite runner.
package benchmarks

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LoadSuite loads a benchmark suite from a JSON file
func LoadSuite(path string) (*BenchmarkSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read suite file: %w", err)
	}

	var suite BenchmarkSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse suite JSON: %w", err)
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
	}

	// Compute overall metrics
	run.OverallAccuracy = float64(run.CorrectProblems) / float64(run.TotalProblems)
	run.OverallECE = computeECE(run.Results)
	run.AvgLatency = computeAvgLatency(run.Results)

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
