// Package evaluators provides calibration metric implementations.
package evaluators

import (
	"math"
)

// CalibrationMetrics computes confidence calibration quality
type CalibrationMetrics struct {
	ECE              float64         // Expected Calibration Error
	MCE              float64         // Maximum Calibration Error
	Brier            float64         // Brier score
	BucketAccuracies map[int]float64 // Accuracy per confidence bucket
	BucketCounts     map[int]int     // Count per bucket
	BucketConfidence map[int]float64 // Avg confidence per bucket
}

// CalibrationResult represents a single prediction for calibration
type CalibrationResult struct {
	Correct    bool
	Confidence float64
}

// ComputeCalibration calculates calibration metrics from results
func ComputeCalibration(results []CalibrationResult) *CalibrationMetrics {
	if len(results) == 0 {
		return &CalibrationMetrics{
			BucketAccuracies: make(map[int]float64),
			BucketCounts:     make(map[int]int),
			BucketConfidence: make(map[int]float64),
		}
	}

	// Initialize 10 buckets (0-10%, 10-20%, ..., 90-100%)
	buckets := make([]struct {
		totalCount    int
		correctCount  int
		sumConfidence float64
	}, 10)

	brierSum := 0.0

	for _, result := range results {
		// Skip invalid confidence values
		if result.Confidence < 0 || result.Confidence > 1 {
			continue
		}

		// Determine bucket (0-9)
		bucketIdx := int(result.Confidence * 10)
		if bucketIdx >= 10 {
			bucketIdx = 9
		}

		buckets[bucketIdx].totalCount++
		if result.Correct {
			buckets[bucketIdx].correctCount++
		}
		buckets[bucketIdx].sumConfidence += result.Confidence

		// Compute Brier score component
		actual := 0.0
		if result.Correct {
			actual = 1.0
		}
		diff := result.Confidence - actual
		brierSum += diff * diff
	}

	// Calculate metrics
	totalCount := len(results)
	ece := 0.0
	mce := 0.0
	brier := brierSum / float64(totalCount)

	bucketAccuracies := make(map[int]float64)
	bucketCounts := make(map[int]int)
	bucketConfidence := make(map[int]float64)

	for i, bucket := range buckets {
		if bucket.totalCount == 0 {
			continue
		}

		bucketAccuracy := float64(bucket.correctCount) / float64(bucket.totalCount)
		avgConfidence := bucket.sumConfidence / float64(bucket.totalCount)
		weight := float64(bucket.totalCount) / float64(totalCount)

		calibrationError := math.Abs(bucketAccuracy - avgConfidence)
		ece += weight * calibrationError

		if calibrationError > mce {
			mce = calibrationError
		}

		bucketAccuracies[i] = bucketAccuracy
		bucketCounts[i] = bucket.totalCount
		bucketConfidence[i] = avgConfidence
	}

	return &CalibrationMetrics{
		ECE:              ece,
		MCE:              mce,
		Brier:            brier,
		BucketAccuracies: bucketAccuracies,
		BucketCounts:     bucketCounts,
		BucketConfidence: bucketConfidence,
	}
}

// CalibrationQuality categorizes calibration quality
func CalibrationQuality(ece float64) string {
	if ece < 0.05 {
		return "excellent"
	} else if ece < 0.10 {
		return "good"
	} else if ece < 0.15 {
		return "acceptable"
	} else if ece < 0.25 {
		return "poor"
	}
	return "very poor"
}
