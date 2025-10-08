package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalibrationTracker_RecordPrediction(t *testing.T) {
	tracker := NewCalibrationTracker()

	tests := []struct {
		name        string
		prediction  *Prediction
		expectError bool
	}{
		{
			name: "valid prediction",
			prediction: &Prediction{
				ThoughtID:  "test-1",
				Confidence: 0.8,
				Mode:       "linear",
			},
			expectError: false,
		},
		{
			name: "missing thought_id",
			prediction: &Prediction{
				Confidence: 0.8,
				Mode:       "linear",
			},
			expectError: true,
		},
		{
			name: "confidence too high",
			prediction: &Prediction{
				ThoughtID:  "test-2",
				Confidence: 1.5,
				Mode:       "linear",
			},
			expectError: true,
		},
		{
			name: "confidence too low",
			prediction: &Prediction{
				ThoughtID:  "test-3",
				Confidence: -0.1,
				Mode:       "linear",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tracker.RecordPrediction(tt.prediction)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.prediction.Timestamp)
			}
		})
	}
}

func TestCalibrationTracker_RecordOutcome(t *testing.T) {
	tracker := NewCalibrationTracker()

	// Record prediction first
	pred := &Prediction{
		ThoughtID:  "test-1",
		Confidence: 0.8,
		Mode:       "linear",
	}
	err := tracker.RecordPrediction(pred)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		outcome     *Outcome
		expectError bool
	}{
		{
			name: "valid outcome",
			outcome: &Outcome{
				ThoughtID:        "test-1",
				WasCorrect:       true,
				ActualConfidence: 0.9,
				Source:           OutcomeSourceValidation,
			},
			expectError: false,
		},
		{
			name: "missing thought_id",
			outcome: &Outcome{
				WasCorrect:       true,
				ActualConfidence: 0.9,
			},
			expectError: true,
		},
		{
			name: "no prediction exists",
			outcome: &Outcome{
				ThoughtID:        "nonexistent",
				WasCorrect:       true,
				ActualConfidence: 0.9,
			},
			expectError: true,
		},
		{
			name: "invalid actual_confidence",
			outcome: &Outcome{
				ThoughtID:        "test-1",
				WasCorrect:       true,
				ActualConfidence: 1.5,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tracker.RecordOutcome(tt.outcome)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.outcome.Timestamp)
			}
		})
	}
}

func TestCalibrationTracker_GetCalibrationReport_NoData(t *testing.T) {
	tracker := NewCalibrationTracker()

	report := tracker.GetCalibrationReport()

	assert.NotNil(t, report)
	assert.Equal(t, 0, report.TotalPredictions)
	assert.Equal(t, 0, report.TotalOutcomes)
	assert.NotEmpty(t, report.Recommendations)
	assert.Contains(t, report.Recommendations[0], "No outcomes recorded")
}

func TestCalibrationTracker_GetCalibrationReport_WellCalibrated(t *testing.T) {
	tracker := NewCalibrationTracker()

	// Create well-calibrated data across multiple confidence levels
	// 50% confidence = 50% accuracy, 70% = 70%, 90% = 90%
	confidenceLevels := []struct {
		conf    float64
		correct int
		total   int
	}{
		{0.5, 5, 10},
		{0.7, 7, 10},
		{0.9, 9, 10},
	}

	idx := 0
	for _, level := range confidenceLevels {
		for i := 0; i < level.total; i++ {
			pred := &Prediction{
				ThoughtID:  string(rune('a' + idx)),
				Confidence: level.conf,
				Mode:       "linear",
			}
			tracker.RecordPrediction(pred)

			outcome := &Outcome{
				ThoughtID:        pred.ThoughtID,
				WasCorrect:       i < level.correct,
				ActualConfidence: level.conf,
				Source:           OutcomeSourceValidation,
			}
			tracker.RecordOutcome(outcome)
			idx++
		}
	}

	report := tracker.GetCalibrationReport()

	assert.Equal(t, 30, report.TotalPredictions)
	assert.Equal(t, 30, report.TotalOutcomes)
	assert.InDelta(t, 0.7, report.OverallAccuracy, 0.05) // ~70% overall

	// With well-calibrated data, bias should be small
	assert.LessOrEqual(t, report.Bias.Magnitude, 0.1)

	// Should not recommend major adjustments for well-calibrated data
	if len(report.Recommendations) > 0 {
		assert.NotContains(t, report.Recommendations[0], "Significant")
	}
}

func TestCalibrationTracker_GetCalibrationReport_Overconfident(t *testing.T) {
	tracker := NewCalibrationTracker()

	// Create overconfident data: 90% confidence but only 60% accuracy
	for i := 0; i < 10; i++ {
		pred := &Prediction{
			ThoughtID:  string(rune('a' + i)),
			Confidence: 0.9,
			Mode:       "linear",
		}
		tracker.RecordPrediction(pred)

		outcome := &Outcome{
			ThoughtID:        pred.ThoughtID,
			WasCorrect:       i < 6, // 60% correct
			ActualConfidence: 0.6,
			Source:           OutcomeSourceValidation,
		}
		tracker.RecordOutcome(outcome)
	}

	report := tracker.GetCalibrationReport()

	assert.Equal(t, 10, report.TotalPredictions)
	assert.Equal(t, 10, report.TotalOutcomes)
	assert.InDelta(t, 0.6, report.OverallAccuracy, 0.01)
	assert.Equal(t, BiasOverconfident, report.Bias.Type)
	assert.Greater(t, report.Bias.Magnitude, 0.15)
	assert.Contains(t, report.Recommendations[0], "overconfidence")
}

func TestCalibrationTracker_GetCalibrationReport_Underconfident(t *testing.T) {
	tracker := NewCalibrationTracker()

	// Create underconfident data: 60% confidence but 85% accuracy
	for i := 0; i < 20; i++ {
		pred := &Prediction{
			ThoughtID:  string(rune('a' + i)),
			Confidence: 0.6,
			Mode:       "linear",
		}
		tracker.RecordPrediction(pred)

		outcome := &Outcome{
			ThoughtID:        pred.ThoughtID,
			WasCorrect:       i < 17, // 85% correct
			ActualConfidence: 0.85,
			Source:           OutcomeSourceValidation,
		}
		tracker.RecordOutcome(outcome)
	}

	report := tracker.GetCalibrationReport()

	assert.Equal(t, 20, report.TotalPredictions)
	assert.Equal(t, 20, report.TotalOutcomes)
	assert.InDelta(t, 0.85, report.OverallAccuracy, 0.01)
	assert.Equal(t, BiasUnderconfident, report.Bias.Type)
	assert.Greater(t, report.Bias.Magnitude, 0.15)
	assert.Contains(t, report.Recommendations[0], "underconfidence")
}

func TestCalibrationTracker_GetCalibrationReport_ByMode(t *testing.T) {
	tracker := NewCalibrationTracker()

	// Linear mode: well calibrated
	for i := 0; i < 10; i++ {
		pred := &Prediction{
			ThoughtID:  "linear-" + string(rune('a'+i)),
			Confidence: 0.8,
			Mode:       "linear",
		}
		tracker.RecordPrediction(pred)

		outcome := &Outcome{
			ThoughtID:        pred.ThoughtID,
			WasCorrect:       i < 8,
			ActualConfidence: 0.8,
			Source:           OutcomeSourceValidation,
		}
		tracker.RecordOutcome(outcome)
	}

	// Tree mode: overconfident
	for i := 0; i < 10; i++ {
		pred := &Prediction{
			ThoughtID:  "tree-" + string(rune('a'+i)),
			Confidence: 0.9,
			Mode:       "tree",
		}
		tracker.RecordPrediction(pred)

		outcome := &Outcome{
			ThoughtID:        pred.ThoughtID,
			WasCorrect:       i < 5,
			ActualConfidence: 0.5,
			Source:           OutcomeSourceValidation,
		}
		tracker.RecordOutcome(outcome)
	}

	report := tracker.GetCalibrationReport()

	assert.Equal(t, 20, report.TotalPredictions)
	assert.Equal(t, 20, report.TotalOutcomes)
	assert.Len(t, report.ByMode, 2)

	linearMode := report.ByMode["linear"]
	assert.NotNil(t, linearMode)
	assert.Equal(t, 10, linearMode.PredictionCount)
	assert.InDelta(t, 0.8, linearMode.Accuracy, 0.01)

	treeMode := report.ByMode["tree"]
	assert.NotNil(t, treeMode)
	assert.Equal(t, 10, treeMode.PredictionCount)
	assert.InDelta(t, 0.5, treeMode.Accuracy, 0.01)
}

func TestCalibrationTracker_CalibrationBuckets(t *testing.T) {
	tracker := NewCalibrationTracker()

	// Create data across multiple confidence levels
	confidenceLevels := []float64{0.1, 0.3, 0.5, 0.7, 0.9}
	for _, conf := range confidenceLevels {
		for i := 0; i < 10; i++ {
			pred := &Prediction{
				ThoughtID:  string(rune('a'+i)) + "-" + string(rune('0'+int(conf*10))),
				Confidence: conf,
				Mode:       "linear",
			}
			tracker.RecordPrediction(pred)

			// Make accuracy match confidence for well-calibrated data
			correct := float64(i) < conf*10
			outcome := &Outcome{
				ThoughtID:        pred.ThoughtID,
				WasCorrect:       correct,
				ActualConfidence: conf,
				Source:           OutcomeSourceValidation,
			}
			tracker.RecordOutcome(outcome)
		}
	}

	report := tracker.GetCalibrationReport()

	assert.Equal(t, 50, report.TotalPredictions)
	assert.Equal(t, 50, report.TotalOutcomes)
	assert.NotEmpty(t, report.Buckets)

	// Should have buckets for each confidence level
	assert.GreaterOrEqual(t, len(report.Buckets), 5)
}

func TestCalibrationTracker_GetPrediction(t *testing.T) {
	tracker := NewCalibrationTracker()

	pred := &Prediction{
		ThoughtID:  "test-1",
		Confidence: 0.8,
		Mode:       "linear",
	}
	tracker.RecordPrediction(pred)

	retrieved, err := tracker.GetPrediction("test-1")
	assert.NoError(t, err)
	assert.Equal(t, pred.ThoughtID, retrieved.ThoughtID)
	assert.Equal(t, pred.Confidence, retrieved.Confidence)

	_, err = tracker.GetPrediction("nonexistent")
	assert.Error(t, err)
}

func TestCalibrationTracker_GetOutcome(t *testing.T) {
	tracker := NewCalibrationTracker()

	pred := &Prediction{
		ThoughtID:  "test-1",
		Confidence: 0.8,
		Mode:       "linear",
	}
	tracker.RecordPrediction(pred)

	outcome := &Outcome{
		ThoughtID:        "test-1",
		WasCorrect:       true,
		ActualConfidence: 0.9,
		Source:           OutcomeSourceValidation,
	}
	tracker.RecordOutcome(outcome)

	retrieved, err := tracker.GetOutcome("test-1")
	assert.NoError(t, err)
	assert.Equal(t, outcome.ThoughtID, retrieved.ThoughtID)
	assert.Equal(t, outcome.WasCorrect, retrieved.WasCorrect)

	_, err = tracker.GetOutcome("nonexistent")
	assert.Error(t, err)
}

func TestCalibrationTracker_ListPredictions(t *testing.T) {
	tracker := NewCalibrationTracker()

	baseTime := time.Now()

	// Add predictions for different modes
	for i := 0; i < 5; i++ {
		pred := &Prediction{
			ThoughtID:  "linear-" + string(rune('a'+i)),
			Confidence: 0.8,
			Mode:       "linear",
		}
		tracker.RecordPrediction(pred)
		// Manually set timestamp after recording to avoid race conditions
		tracker.predictions[pred.ThoughtID].Timestamp = baseTime.Add(time.Duration(i) * time.Minute)
	}

	for i := 0; i < 3; i++ {
		pred := &Prediction{
			ThoughtID:  "tree-" + string(rune('a'+i)),
			Confidence: 0.7,
			Mode:       "tree",
		}
		tracker.RecordPrediction(pred)
		// Tree predictions are more recent
		tracker.predictions[pred.ThoughtID].Timestamp = baseTime.Add(time.Duration(i+10) * time.Minute)
	}

	// List all
	all := tracker.ListPredictions("", 0)
	assert.Len(t, all, 8)

	// List by mode
	linear := tracker.ListPredictions("linear", 0)
	assert.Len(t, linear, 5)

	tree := tracker.ListPredictions("tree", 0)
	assert.Len(t, tree, 3)

	// List with limit
	limited := tracker.ListPredictions("", 3)
	assert.Len(t, limited, 3)

	// Should be sorted by timestamp descending (most recent first)
	assert.True(t, limited[0].Timestamp.After(limited[1].Timestamp))
}

func TestCalibrationTracker_Clear(t *testing.T) {
	tracker := NewCalibrationTracker()

	// Add some data
	pred := &Prediction{
		ThoughtID:  "test-1",
		Confidence: 0.8,
		Mode:       "linear",
	}
	tracker.RecordPrediction(pred)

	outcome := &Outcome{
		ThoughtID:        "test-1",
		WasCorrect:       true,
		ActualConfidence: 0.9,
		Source:           OutcomeSourceValidation,
	}
	tracker.RecordOutcome(outcome)

	// Clear
	tracker.Clear()

	// Verify empty
	report := tracker.GetCalibrationReport()
	assert.Equal(t, 0, report.TotalPredictions)
	assert.Equal(t, 0, report.TotalOutcomes)

	_, err := tracker.GetPrediction("test-1")
	assert.Error(t, err)

	_, err = tracker.GetOutcome("test-1")
	assert.Error(t, err)
}

func TestCalibrationTracker_ExpectedCalibrationError(t *testing.T) {
	tracker := NewCalibrationTracker()

	// Create data with known ECE
	// Bucket 1: 0.5 confidence, 0.8 actual = 0.3 error, weight 0.5
	// Bucket 2: 0.9 confidence, 0.9 actual = 0.0 error, weight 0.5
	// Expected ECE = 0.5 * 0.3 + 0.5 * 0.0 = 0.15

	for i := 0; i < 10; i++ {
		pred := &Prediction{
			ThoughtID:  "bucket1-" + string(rune('a'+i)),
			Confidence: 0.5,
			Mode:       "linear",
		}
		tracker.RecordPrediction(pred)

		outcome := &Outcome{
			ThoughtID:        pred.ThoughtID,
			WasCorrect:       i < 8, // 80% correct
			ActualConfidence: 0.8,
			Source:           OutcomeSourceValidation,
		}
		tracker.RecordOutcome(outcome)
	}

	for i := 0; i < 10; i++ {
		pred := &Prediction{
			ThoughtID:  "bucket2-" + string(rune('a'+i)),
			Confidence: 0.9,
			Mode:       "linear",
		}
		tracker.RecordPrediction(pred)

		outcome := &Outcome{
			ThoughtID:        pred.ThoughtID,
			WasCorrect:       i < 9, // 90% correct
			ActualConfidence: 0.9,
			Source:           OutcomeSourceValidation,
		}
		tracker.RecordOutcome(outcome)
	}

	report := tracker.GetCalibrationReport()

	// ECE should be around 0.15 (allowing some tolerance for bucketing)
	assert.InDelta(t, 0.15, report.Calibration, 0.1)
}
