package evaluators

import (
	"strings"
	"testing"
)

func TestComputeLearning(t *testing.T) {
	tests := []struct {
		name                   string
		results                []IterationResult
		wantInitial            float64
		wantFinal              float64
		wantImprovement        float64
		wantSignificantLearning bool
	}{
		{
			name:                   "empty results",
			results:                []IterationResult{},
			wantInitial:            0.0,
			wantFinal:              0.0,
			wantImprovement:        0.0,
			wantSignificantLearning: false,
		},
		{
			name: "no learning - stable performance",
			results: []IterationResult{
				{Iteration: 1, Correct: 5, Total: 10, Accuracy: 0.5},
				{Iteration: 2, Correct: 5, Total: 10, Accuracy: 0.5},
				{Iteration: 3, Correct: 5, Total: 10, Accuracy: 0.5},
			},
			wantInitial:            0.5,
			wantFinal:              0.5,
			wantImprovement:        0.0,
			wantSignificantLearning: false,
		},
		{
			name: "significant learning - 0.3 to 0.8",
			results: []IterationResult{
				{Iteration: 1, Correct: 3, Total: 10, Accuracy: 0.3},
				{Iteration: 2, Correct: 5, Total: 10, Accuracy: 0.5},
				{Iteration: 3, Correct: 8, Total: 10, Accuracy: 0.8},
			},
			wantInitial:            0.3,
			wantFinal:              0.8,
			wantImprovement:        0.5,
			wantSignificantLearning: true,
		},
		{
			name: "weak learning - at threshold",
			results: []IterationResult{
				{Iteration: 1, Correct: 5, Total: 10, Accuracy: 0.5},
				{Iteration: 2, Correct: 6, Total: 10, Accuracy: 0.6},
			},
			wantInitial:            0.5,
			wantFinal:              0.6,
			wantImprovement:        0.1,
			wantSignificantLearning: false, // Will be slightly below due to float precision
		},
		{
			name: "regression - performance degrades",
			results: []IterationResult{
				{Iteration: 1, Correct: 8, Total: 10, Accuracy: 0.8},
				{Iteration: 2, Correct: 5, Total: 10, Accuracy: 0.5},
				{Iteration: 3, Correct: 3, Total: 10, Accuracy: 0.3},
			},
			wantInitial:            0.8,
			wantFinal:              0.3,
			wantImprovement:        -0.5,
			wantSignificantLearning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := ComputeLearning(tt.results)

			if metrics.InitialAccuracy != tt.wantInitial {
				t.Errorf("InitialAccuracy = %v, want %v", metrics.InitialAccuracy, tt.wantInitial)
			}

			if metrics.FinalAccuracy != tt.wantFinal {
				t.Errorf("FinalAccuracy = %v, want %v", metrics.FinalAccuracy, tt.wantFinal)
			}

			// Allow small floating point tolerance
			tolerance := 0.01
			if diff := metrics.ImprovementRate - tt.wantImprovement; diff > tolerance || diff < -tolerance {
				t.Errorf("ImprovementRate = %v, want %v (tolerance: %v)", metrics.ImprovementRate, tt.wantImprovement, tolerance)
			}

			if metrics.SignificantImprovement != tt.wantSignificantLearning {
				t.Errorf("SignificantImprovement = %v, want %v", metrics.SignificantImprovement, tt.wantSignificantLearning)
			}

			// Verify AccuracyByIter contains all iterations
			if len(tt.results) > 0 {
				if len(metrics.AccuracyByIter) != len(tt.results) {
					t.Errorf("AccuracyByIter has %d entries, want %d", len(metrics.AccuracyByIter), len(tt.results))
				}
			}
		})
	}
}

func TestGroupByIteration(t *testing.T) {
	results := []struct {
		Correct   bool
		Iteration int
	}{
		{Correct: true, Iteration: 1},
		{Correct: false, Iteration: 1},
		{Correct: true, Iteration: 1}, // Iter 1: 2/3 = 66.7%
		{Correct: true, Iteration: 2},
		{Correct: true, Iteration: 2}, // Iter 2: 2/2 = 100%
		{Correct: false, Iteration: 3},
		{Correct: false, Iteration: 3},
		{Correct: false, Iteration: 3}, // Iter 3: 0/3 = 0%
	}

	grouped := GroupByIteration(results)

	if len(grouped) != 3 {
		t.Fatalf("Expected 3 iterations, got %d", len(grouped))
	}

	// Check iteration 1
	iter1 := grouped[0]
	if iter1.Iteration != 1 {
		t.Errorf("First result should be iteration 1, got %d", iter1.Iteration)
	}
	if iter1.Total != 3 {
		t.Errorf("Iteration 1 total = %d, want 3", iter1.Total)
	}
	if iter1.Correct != 2 {
		t.Errorf("Iteration 1 correct = %d, want 2", iter1.Correct)
	}

	// Check iteration 2
	iter2 := grouped[1]
	if iter2.Accuracy != 1.0 {
		t.Errorf("Iteration 2 accuracy = %v, want 1.0", iter2.Accuracy)
	}

	// Check iteration 3
	iter3 := grouped[2]
	if iter3.Accuracy != 0.0 {
		t.Errorf("Iteration 3 accuracy = %v, want 0.0", iter3.Accuracy)
	}
}

func TestFormatLearningReport(t *testing.T) {
	metrics := &LearningMetrics{
		InitialAccuracy:     0.3,
		FinalAccuracy:       0.8,
		ImprovementRate:     0.5,
		RelativeImprovement: 166.67,
		Iterations:          3,
		LearningRate:        0.25,
		AccuracyByIter: map[int]float64{
			1: 0.3,
			2: 0.5,
			3: 0.8,
		},
		SignificantImprovement: true,
	}

	report := FormatLearningReport(metrics)

	// Verify report contains key information
	if !strings.Contains(report, "Initial Accuracy: 30.00%") {
		t.Error("Report should contain initial accuracy")
	}

	if !strings.Contains(report, "Final Accuracy: 80.00%") {
		t.Error("Report should contain final accuracy")
	}

	if !strings.Contains(report, "Improvement: 50.00%") {
		t.Error("Report should contain improvement rate")
	}

	if !strings.Contains(report, "SIGNIFICANT") {
		t.Error("Report should indicate significant improvement")
	}

	if !strings.Contains(report, "Iteration 1: 30.00%") {
		t.Error("Report should contain per-iteration breakdown")
	}
}

func TestDetectLearning(t *testing.T) {
	tests := []struct {
		name           string
		metrics        *LearningMetrics
		minImprovement float64
		want           bool
	}{
		{
			name: "above threshold",
			metrics: &LearningMetrics{
				InitialAccuracy: 0.3,
				FinalAccuracy:   0.5,
				ImprovementRate: 0.2,
			},
			minImprovement: 0.1,
			want:           true,
		},
		{
			name: "at threshold",
			metrics: &LearningMetrics{
				InitialAccuracy: 0.3,
				FinalAccuracy:   0.4,
				ImprovementRate: 0.1,
			},
			minImprovement: 0.1,
			want:           true,
		},
		{
			name: "below threshold",
			metrics: &LearningMetrics{
				InitialAccuracy: 0.3,
				FinalAccuracy:   0.35,
				ImprovementRate: 0.05,
			},
			minImprovement: 0.1,
			want:           false,
		},
		{
			name: "negative improvement",
			metrics: &LearningMetrics{
				InitialAccuracy: 0.8,
				FinalAccuracy:   0.5,
				ImprovementRate: -0.3,
			},
			minImprovement: 0.1,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLearning(tt.metrics, tt.minImprovement)
			if got != tt.want {
				t.Errorf("DetectLearning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLearningTrend(t *testing.T) {
	tests := []struct {
		improvementRate float64
		want            string
	}{
		{improvementRate: 0.25, want: "strong learning"},
		{improvementRate: 0.15, want: "moderate learning"},
		{improvementRate: 0.08, want: "weak learning"},
		{improvementRate: 0.02, want: "stable"},
		{improvementRate: -0.02, want: "stable"},
		{improvementRate: -0.08, want: "slight degradation"},
		{improvementRate: -0.15, want: "regression"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			metrics := &LearningMetrics{ImprovementRate: tt.improvementRate}
			got := LearningTrend(metrics)
			if got != tt.want {
				t.Errorf("LearningTrend(%v) = %v, want %v", tt.improvementRate, got, tt.want)
			}
		})
	}
}
