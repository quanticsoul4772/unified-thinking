package evaluators

import (
	"math"
	"testing"
)

func TestComputeCalibration(t *testing.T) {
	tests := []struct {
		name    string
		results []CalibrationResult
		wantECE float64
		wantMCE float64
	}{
		{
			name:    "empty results",
			results: []CalibrationResult{},
			wantECE: 0.0,
			wantMCE: 0.0,
		},
		{
			name: "perfect calibration",
			results: []CalibrationResult{
				{Correct: true, Confidence: 0.9},
				{Correct: true, Confidence: 0.9},
				{Correct: false, Confidence: 0.1},
				{Correct: false, Confidence: 0.1},
			},
			wantECE: 0.1, // Bucket 0 (0-10%) has 0% accuracy but 10% confidence, Bucket 8 (80-90%) has 100% accuracy but 90% confidence
			wantMCE: 0.1,
		},
		{
			name: "overconfident - high confidence, low accuracy",
			results: []CalibrationResult{
				{Correct: false, Confidence: 0.9},
				{Correct: false, Confidence: 0.9},
				{Correct: false, Confidence: 0.9},
				{Correct: false, Confidence: 0.9},
			},
			wantECE: 0.9,
			wantMCE: 0.9,
		},
		{
			name: "underconfident - low confidence, high accuracy",
			results: []CalibrationResult{
				{Correct: true, Confidence: 0.1},
				{Correct: true, Confidence: 0.1},
				{Correct: true, Confidence: 0.1},
				{Correct: true, Confidence: 0.1},
			},
			wantECE: 0.9,
			wantMCE: 0.9,
		},
		{
			name: "mixed calibration",
			results: []CalibrationResult{
				{Correct: true, Confidence: 0.8},
				{Correct: false, Confidence: 0.8},
				{Correct: true, Confidence: 0.2},
				{Correct: false, Confidence: 0.2},
			},
			wantECE: 0.3, // Bucket 1 (10-20%): 50% accuracy but 20% confidence, Bucket 7 (70-80%): 50% accuracy but 80% confidence
			wantMCE: 0.3,
		},
		{
			name: "invalid confidence - negative",
			results: []CalibrationResult{
				{Correct: true, Confidence: -0.5},
				{Correct: true, Confidence: 0.9},
			},
			wantECE: 0.05, // Only valid result in bucket 8 (80-90%): 100% accuracy, 90% confidence
			wantMCE: 0.1,
		},
		{
			name: "invalid confidence - above 1",
			results: []CalibrationResult{
				{Correct: true, Confidence: 1.5},
				{Correct: true, Confidence: 0.9},
			},
			wantECE: 0.05,
			wantMCE: 0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := ComputeCalibration(tt.results)

			// Allow small floating point error
			tolerance := 0.01

			if math.Abs(metrics.ECE-tt.wantECE) > tolerance {
				t.Errorf("ComputeCalibration() ECE = %v, want %v", metrics.ECE, tt.wantECE)
			}

			if math.Abs(metrics.MCE-tt.wantMCE) > tolerance {
				t.Errorf("ComputeCalibration() MCE = %v, want %v", metrics.MCE, tt.wantMCE)
			}

			// Verify Brier score is in valid range [0, 1]
			if metrics.Brier < 0 || metrics.Brier > 1 {
				t.Errorf("ComputeCalibration() Brier = %v, must be in [0, 1]", metrics.Brier)
			}

			// Verify maps are initialized
			if metrics.BucketAccuracies == nil {
				t.Error("BucketAccuracies should not be nil")
			}
			if metrics.BucketCounts == nil {
				t.Error("BucketCounts should not be nil")
			}
			if metrics.BucketConfidence == nil {
				t.Error("BucketConfidence should not be nil")
			}
		})
	}
}

func TestComputeCalibration_BrierScore(t *testing.T) {
	// Test Brier score computation specifically
	results := []CalibrationResult{
		{Correct: true, Confidence: 1.0},  // (1-1)^2 = 0
		{Correct: false, Confidence: 0.0}, // (0-0)^2 = 0
		{Correct: true, Confidence: 0.5},  // (1-0.5)^2 = 0.25
		{Correct: false, Confidence: 0.5}, // (0-0.5)^2 = 0.25
	}

	metrics := ComputeCalibration(results)

	expectedBrier := (0.0 + 0.0 + 0.25 + 0.25) / 4.0 // 0.125
	if math.Abs(metrics.Brier-expectedBrier) > 0.01 {
		t.Errorf("Brier score = %v, want %v", metrics.Brier, expectedBrier)
	}
}

func TestCalibrationQuality(t *testing.T) {
	tests := []struct {
		ece  float64
		want string
	}{
		{ece: 0.02, want: "excellent"},
		{ece: 0.05, want: "good"},
		{ece: 0.08, want: "good"},
		{ece: 0.10, want: "acceptable"},
		{ece: 0.12, want: "acceptable"},
		{ece: 0.18, want: "poor"},
		{ece: 0.25, want: "very poor"},
		{ece: 0.50, want: "very poor"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := CalibrationQuality(tt.ece)
			if got != tt.want {
				t.Errorf("CalibrationQuality(%v) = %v, want %v", tt.ece, got, tt.want)
			}
		})
	}
}

func TestComputeCalibration_BucketDistribution(t *testing.T) {
	// Test that results are properly distributed into buckets
	results := []CalibrationResult{
		{Correct: true, Confidence: 0.15},  // Bucket 1 (10-20%)
		{Correct: true, Confidence: 0.18},  // Bucket 1
		{Correct: true, Confidence: 0.85},  // Bucket 8 (80-90%)
		{Correct: false, Confidence: 0.87}, // Bucket 8
	}

	metrics := ComputeCalibration(results)

	// Check bucket 1 (0.1-0.2)
	if count, exists := metrics.BucketCounts[1]; !exists || count != 2 {
		t.Errorf("Bucket 1 should have 2 results, got %d", count)
	}

	if acc, exists := metrics.BucketAccuracies[1]; !exists || acc != 1.0 {
		t.Errorf("Bucket 1 accuracy should be 1.0 (2/2), got %v", acc)
	}

	// Check bucket 8 (0.8-0.9)
	if count, exists := metrics.BucketCounts[8]; !exists || count != 2 {
		t.Errorf("Bucket 8 should have 2 results, got %d", count)
	}

	if acc, exists := metrics.BucketAccuracies[8]; !exists || acc != 0.5 {
		t.Errorf("Bucket 8 accuracy should be 0.5 (1/2), got %v", acc)
	}
}
