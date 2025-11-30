package evaluators

import (
	"math"
	"testing"
	"time"
)

func TestComputeEfficiency(t *testing.T) {
	tests := []struct {
		name    string
		results []LatencyResult
		wantAvg time.Duration
		wantP50 time.Duration
	}{
		{
			name:    "empty results",
			results: []LatencyResult{},
			wantAvg: 0,
			wantP50: 0,
		},
		{
			name: "single result",
			results: []LatencyResult{
				{ProblemID: "p1", Latency: 100 * time.Millisecond, Tokens: 50},
			},
			wantAvg: 100 * time.Millisecond,
			wantP50: 100 * time.Millisecond,
		},
		{
			name: "multiple results - all same",
			results: []LatencyResult{
				{ProblemID: "p1", Latency: 50 * time.Millisecond, Tokens: 10},
				{ProblemID: "p2", Latency: 50 * time.Millisecond, Tokens: 10},
				{ProblemID: "p3", Latency: 50 * time.Millisecond, Tokens: 10},
			},
			wantAvg: 50 * time.Millisecond,
			wantP50: 50 * time.Millisecond,
		},
		{
			name: "multiple results - varied",
			results: []LatencyResult{
				{ProblemID: "p1", Latency: 10 * time.Millisecond, Tokens: 5},
				{ProblemID: "p2", Latency: 50 * time.Millisecond, Tokens: 20},
				{ProblemID: "p3", Latency: 100 * time.Millisecond, Tokens: 50},
			},
			wantAvg: 53333333 * time.Nanosecond, // (10+50+100)/3 ms
			wantP50: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := ComputeEfficiency(tt.results)

			if metrics.AvgLatency != tt.wantAvg {
				t.Errorf("ComputeEfficiency() AvgLatency = %v, want %v", metrics.AvgLatency, tt.wantAvg)
			}

			if metrics.P50Latency != tt.wantP50 {
				t.Errorf("ComputeEfficiency() P50Latency = %v, want %v", metrics.P50Latency, tt.wantP50)
			}

			// Verify all percentiles are in ascending order
			if metrics.P50Latency > metrics.P95Latency {
				t.Error("P50 should be <= P95")
			}
			if metrics.P95Latency > metrics.P99Latency {
				t.Error("P95 should be <= P99")
			}

			// Verify throughput calculation
			if len(tt.results) > 0 && metrics.TotalLatency > 0 {
				expectedThroughput := float64(len(tt.results)) / metrics.TotalLatency.Seconds()
				if metrics.ProblemsPerSec != expectedThroughput {
					t.Errorf("ProblemsPerSec = %v, want %v", metrics.ProblemsPerSec, expectedThroughput)
				}
			}
		})
	}
}

func TestComputeEfficiency_TokenTracking(t *testing.T) {
	results := []LatencyResult{
		{ProblemID: "p1", Latency: 10 * time.Millisecond, Tokens: 100},
		{ProblemID: "p2", Latency: 20 * time.Millisecond, Tokens: 200},
		{ProblemID: "p3", Latency: 30 * time.Millisecond, Tokens: 300},
	}

	metrics := ComputeEfficiency(results)

	expectedTotal := 600
	if metrics.TotalTokens != expectedTotal {
		t.Errorf("TotalTokens = %d, want %d", metrics.TotalTokens, expectedTotal)
	}

	expectedAvg := 200.0
	if metrics.AvgTokens != expectedAvg {
		t.Errorf("AvgTokens = %v, want %v", metrics.AvgTokens, expectedAvg)
	}
}

func TestComputeEfficiency_Percentiles(t *testing.T) {
	// Create 100 results with known distribution
	results := make([]LatencyResult, 100)
	for i := 0; i < 100; i++ {
		results[i] = LatencyResult{
			ProblemID: "p",
			Latency:   time.Duration(i+1) * time.Millisecond,
			Tokens:    10,
		}
	}

	metrics := ComputeEfficiency(results)

	// P50 should be around 50ms (median)
	expectedP50 := 50 * time.Millisecond
	tolerance := 5 * time.Millisecond
	if diff := time.Duration(math.Abs(float64(metrics.P50Latency - expectedP50))); diff > tolerance {
		t.Errorf("P50Latency = %v, want ~%v (tolerance %v)", metrics.P50Latency, expectedP50, tolerance)
	}

	// P95 should be around 95ms
	expectedP95 := 95 * time.Millisecond
	if diff := time.Duration(math.Abs(float64(metrics.P95Latency - expectedP95))); diff > tolerance {
		t.Errorf("P95Latency = %v, want ~%v", metrics.P95Latency, expectedP95)
	}

	// P99 should be around 99ms
	expectedP99 := 99 * time.Millisecond
	if diff := time.Duration(math.Abs(float64(metrics.P99Latency - expectedP99))); diff > tolerance {
		t.Errorf("P99Latency = %v, want ~%v", metrics.P99Latency, expectedP99)
	}
}

func TestEfficiencyCategory(t *testing.T) {
	tests := []struct {
		latency time.Duration
		want    string
	}{
		{latency: 50 * time.Millisecond, want: "excellent"},
		{latency: 200 * time.Millisecond, want: "good"},
		{latency: 700 * time.Millisecond, want: "acceptable"},
		{latency: 2 * time.Second, want: "slow"},
		{latency: 10 * time.Second, want: "very slow"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := EfficiencyCategory(tt.latency)
			if got != tt.want {
				t.Errorf("EfficiencyCategory(%v) = %v, want %v", tt.latency, got, tt.want)
			}
		})
	}
}

func TestSortDurations(t *testing.T) {
	durations := []time.Duration{
		100 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		5 * time.Millisecond,
		75 * time.Millisecond,
	}

	sortDurations(durations)

	// Verify sorted in ascending order
	for i := 1; i < len(durations); i++ {
		if durations[i] < durations[i-1] {
			t.Errorf("Durations not sorted: %v came before %v", durations[i-1], durations[i])
		}
	}

	// Check specific values
	expected := []time.Duration{
		5 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		75 * time.Millisecond,
		100 * time.Millisecond,
	}

	for i, want := range expected {
		if durations[i] != want {
			t.Errorf("Position %d: got %v, want %v", i, durations[i], want)
		}
	}
}

func TestPercentile(t *testing.T) {
	sorted := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}

	tests := []struct {
		p    float64
		want time.Duration
	}{
		{p: 0.0, want: 10 * time.Millisecond},
		{p: 0.5, want: 30 * time.Millisecond},
		{p: 1.0, want: 50 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := percentile(sorted, tt.p)
			if got != tt.want {
				t.Errorf("percentile(%v) = %v, want %v", tt.p, got, tt.want)
			}
		})
	}
}

func TestPercentile_EmptySlice(t *testing.T) {
	sorted := []time.Duration{}
	got := percentile(sorted, 0.5)
	if got != 0 {
		t.Errorf("percentile on empty slice = %v, want 0", got)
	}
}
