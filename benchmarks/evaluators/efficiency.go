// Package evaluators provides efficiency metric implementations.
package evaluators

import (
	"time"
)

// EfficiencyMetrics tracks performance characteristics
type EfficiencyMetrics struct {
	TotalLatency   time.Duration // Total time spent
	AvgLatency     time.Duration // Average latency per problem
	P50Latency     time.Duration // Median latency
	P95Latency     time.Duration // 95th percentile latency
	P99Latency     time.Duration // 99th percentile latency
	TotalTokens    int           // Total tokens used (if tracked)
	AvgTokens      float64       // Average tokens per problem
	ProblemsPerSec float64       // Throughput
}

// LatencyResult represents timing information for a single problem
type LatencyResult struct {
	ProblemID string
	Latency   time.Duration
	Tokens    int
}

// ComputeEfficiency calculates efficiency metrics from latency results
func ComputeEfficiency(results []LatencyResult) *EfficiencyMetrics {
	if len(results) == 0 {
		return &EfficiencyMetrics{}
	}

	var totalLatency time.Duration
	totalTokens := 0
	latencies := make([]time.Duration, len(results))

	for i, result := range results {
		totalLatency += result.Latency
		totalTokens += result.Tokens
		latencies[i] = result.Latency
	}

	// Sort latencies for percentiles
	sortDurations(latencies)

	avgLatency := totalLatency / time.Duration(len(results))
	avgTokens := float64(totalTokens) / float64(len(results))

	// Calculate throughput
	problemsPerSec := 0.0
	if totalLatency > 0 {
		problemsPerSec = float64(len(results)) / totalLatency.Seconds()
	}

	return &EfficiencyMetrics{
		TotalLatency:   totalLatency,
		AvgLatency:     avgLatency,
		P50Latency:     percentile(latencies, 0.50),
		P95Latency:     percentile(latencies, 0.95),
		P99Latency:     percentile(latencies, 0.99),
		TotalTokens:    totalTokens,
		AvgTokens:      avgTokens,
		ProblemsPerSec: problemsPerSec,
	}
}

// sortDurations sorts a slice of durations in place
func sortDurations(durations []time.Duration) {
	// Simple bubble sort for small slices
	n := len(durations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}

// percentile computes the pth percentile from sorted durations
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}

	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	// Linear interpolation
	index := p * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[lower]
	}

	// Interpolate
	fraction := index - float64(lower)
	lowerVal := float64(sorted[lower])
	upperVal := float64(sorted[upper])

	return time.Duration(lowerVal + fraction*(upperVal-lowerVal))
}

// EfficiencyCategory categorizes latency performance
func EfficiencyCategory(avgLatency time.Duration) string {
	if avgLatency < 100*time.Millisecond {
		return "excellent"
	} else if avgLatency < 500*time.Millisecond {
		return "good"
	} else if avgLatency < time.Second {
		return "acceptable"
	} else if avgLatency < 5*time.Second {
		return "slow"
	}
	return "very slow"
}
