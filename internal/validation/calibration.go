// Package validation provides confidence calibration tracking.
package validation

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// CalibrationTracker tracks confidence predictions and their outcomes
type CalibrationTracker struct {
	predictions map[string]*Prediction
	outcomes    map[string]*Outcome
	mu          sync.RWMutex
}

// Prediction represents a confidence prediction for a thought
type Prediction struct {
	ThoughtID  string                 `json:"thought_id"`
	Confidence float64                `json:"confidence"` // 0-1
	Mode       string                 `json:"mode"`       // linear, tree, divergent
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Outcome represents the actual result of a prediction
type Outcome struct {
	ThoughtID        string                 `json:"thought_id"`
	WasCorrect       bool                   `json:"was_correct"`
	ActualConfidence float64                `json:"actual_confidence"` // 0-1, determined by validation/verification
	Source           OutcomeSource          `json:"source"`            // validation, verification, user_feedback
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// OutcomeSource indicates how the outcome was determined
type OutcomeSource string

const (
	OutcomeSourceValidation   OutcomeSource = "validation"
	OutcomeSourceVerification OutcomeSource = "verification"
	OutcomeSourceUserFeedback OutcomeSource = "user_feedback"
)

// CalibrationBucket represents a confidence range and its accuracy
type CalibrationBucket struct {
	MinConfidence float64 `json:"min_confidence"`
	MaxConfidence float64 `json:"max_confidence"`
	Count         int     `json:"count"`
	CorrectCount  int     `json:"correct_count"`
	Accuracy      float64 `json:"accuracy"`    // actual proportion correct
	Calibration   float64 `json:"calibration"` // difference from expected
}

// CalibrationReport provides overall calibration metrics
type CalibrationReport struct {
	TotalPredictions int                         `json:"total_predictions"`
	TotalOutcomes    int                         `json:"total_outcomes"`
	Buckets          []CalibrationBucket         `json:"buckets"`
	OverallAccuracy  float64                     `json:"overall_accuracy"`
	Calibration      float64                     `json:"calibration"` // Expected Calibration Error (ECE)
	Bias             CalibrationBias             `json:"bias"`
	ByMode           map[string]*ModeCalibration `json:"by_mode"`
	Recommendations  []string                    `json:"recommendations"`
	GeneratedAt      time.Time                   `json:"generated_at"`
}

// CalibrationBias indicates systematic over/under confidence
type CalibrationBias struct {
	Type        BiasType `json:"type"`
	Magnitude   float64  `json:"magnitude"` // 0-1, how severe
	Description string   `json:"description"`
}

// BiasType categorizes calibration bias
type BiasType string

const (
	BiasNone           BiasType = "none"
	BiasOverconfident  BiasType = "overconfident"
	BiasUnderconfident BiasType = "underconfident"
)

// ModeCalibration tracks calibration by thinking mode
type ModeCalibration struct {
	Mode            string  `json:"mode"`
	PredictionCount int     `json:"prediction_count"`
	OutcomeCount    int     `json:"outcome_count"`
	Accuracy        float64 `json:"accuracy"`
	Calibration     float64 `json:"calibration"`
}

// NewCalibrationTracker creates a new calibration tracker
func NewCalibrationTracker() *CalibrationTracker {
	return &CalibrationTracker{
		predictions: make(map[string]*Prediction),
		outcomes:    make(map[string]*Outcome),
	}
}

// RecordPrediction stores a confidence prediction
func (ct *CalibrationTracker) RecordPrediction(prediction *Prediction) error {
	if prediction.ThoughtID == "" {
		return fmt.Errorf("thought_id is required")
	}
	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	prediction.Timestamp = time.Now()
	ct.predictions[prediction.ThoughtID] = prediction
	return nil
}

// RecordOutcome stores an outcome for a prediction
func (ct *CalibrationTracker) RecordOutcome(outcome *Outcome) error {
	if outcome.ThoughtID == "" {
		return fmt.Errorf("thought_id is required")
	}
	if outcome.ActualConfidence < 0 || outcome.ActualConfidence > 1 {
		return fmt.Errorf("actual_confidence must be between 0 and 1")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Check if prediction exists
	if _, exists := ct.predictions[outcome.ThoughtID]; !exists {
		return fmt.Errorf("no prediction found for thought_id: %s", outcome.ThoughtID)
	}

	outcome.Timestamp = time.Now()
	ct.outcomes[outcome.ThoughtID] = outcome
	return nil
}

// GetCalibrationReport generates a comprehensive calibration report
func (ct *CalibrationTracker) GetCalibrationReport() *CalibrationReport {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	report := &CalibrationReport{
		TotalPredictions: len(ct.predictions),
		TotalOutcomes:    len(ct.outcomes),
		Buckets:          []CalibrationBucket{},
		ByMode:           make(map[string]*ModeCalibration),
		Recommendations:  []string{},
		GeneratedAt:      time.Now(),
	}

	// Collect matched prediction-outcome pairs
	var pairs []struct {
		prediction *Prediction
		outcome    *Outcome
	}

	for thoughtID, prediction := range ct.predictions {
		if outcome, exists := ct.outcomes[thoughtID]; exists {
			pairs = append(pairs, struct {
				prediction *Prediction
				outcome    *Outcome
			}{prediction, outcome})
		}
	}

	if len(pairs) == 0 {
		report.Recommendations = []string{"No outcomes recorded yet. Record outcomes to calculate calibration."}
		return report
	}

	// Calculate buckets (10 buckets: 0-0.1, 0.1-0.2, ..., 0.9-1.0)
	buckets := make([]*CalibrationBucket, 10)
	for i := 0; i < 10; i++ {
		buckets[i] = &CalibrationBucket{
			MinConfidence: float64(i) / 10.0,
			MaxConfidence: float64(i+1) / 10.0,
		}
	}

	// Populate buckets and mode stats
	correctCount := 0
	modeStats := make(map[string]*struct {
		total   int
		correct int
		errors  []float64
	})

	for _, pair := range pairs {
		pred := pair.prediction
		out := pair.outcome

		// Find bucket
		bucketIdx := int(pred.Confidence * 10)
		if bucketIdx >= 10 {
			bucketIdx = 9
		}
		bucket := buckets[bucketIdx]
		bucket.Count++
		if out.WasCorrect {
			bucket.CorrectCount++
			correctCount++
		}

		// Update mode stats
		if _, exists := modeStats[pred.Mode]; !exists {
			modeStats[pred.Mode] = &struct {
				total   int
				correct int
				errors  []float64
			}{}
		}
		stats := modeStats[pred.Mode]
		stats.total++
		if out.WasCorrect {
			stats.correct++
		}
		// Calibration error for this prediction
		expectedCorrect := pred.Confidence
		actualCorrect := 0.0
		if out.WasCorrect {
			actualCorrect = 1.0
		}
		stats.errors = append(stats.errors, math.Abs(expectedCorrect-actualCorrect))
	}

	// Calculate bucket metrics and ECE
	var ece float64 // Expected Calibration Error
	validBuckets := 0
	reportBuckets := []CalibrationBucket{}

	for _, bucket := range buckets {
		if bucket.Count > 0 {
			bucket.Accuracy = float64(bucket.CorrectCount) / float64(bucket.Count)
			expectedAccuracy := (bucket.MinConfidence + bucket.MaxConfidence) / 2
			bucket.Calibration = bucket.Accuracy - expectedAccuracy

			// Contribute to ECE
			weight := float64(bucket.Count) / float64(len(pairs))
			ece += weight * math.Abs(bucket.Calibration)

			validBuckets++
			reportBuckets = append(reportBuckets, *bucket)
		}
	}

	report.Buckets = reportBuckets
	report.OverallAccuracy = float64(correctCount) / float64(len(pairs))
	report.Calibration = ece

	// Determine bias
	report.Bias = ct.calculateBias(reportBuckets)

	// Calculate per-mode calibration
	for mode, stats := range modeStats {
		modeCalib := &ModeCalibration{
			Mode:            mode,
			PredictionCount: stats.total,
			OutcomeCount:    stats.total,
			Accuracy:        float64(stats.correct) / float64(stats.total),
		}

		// Average calibration error for this mode
		if len(stats.errors) > 0 {
			sum := 0.0
			for _, err := range stats.errors {
				sum += err
			}
			modeCalib.Calibration = sum / float64(len(stats.errors))
		}

		report.ByMode[mode] = modeCalib
	}

	// Generate recommendations
	report.Recommendations = ct.generateRecommendations(report)

	return report
}

// calculateBias determines if there's systematic over/under confidence
func (ct *CalibrationTracker) calculateBias(buckets []CalibrationBucket) CalibrationBias {
	if len(buckets) == 0 {
		return CalibrationBias{Type: BiasNone}
	}

	// Calculate weighted average bias
	totalWeight := 0
	weightedBias := 0.0

	for _, bucket := range buckets {
		weight := bucket.Count
		totalWeight += weight
		weightedBias += float64(weight) * bucket.Calibration
	}

	if totalWeight == 0 {
		return CalibrationBias{Type: BiasNone}
	}

	avgBias := weightedBias / float64(totalWeight)
	magnitude := math.Abs(avgBias)

	bias := CalibrationBias{
		Magnitude: magnitude,
	}

	// Classify bias
	if magnitude < 0.05 {
		bias.Type = BiasNone
		bias.Description = "Well calibrated - confidence matches accuracy"
	} else if avgBias > 0 {
		bias.Type = BiasUnderconfident
		if magnitude > 0.15 {
			bias.Description = fmt.Sprintf("Significantly underconfident - actual accuracy %.1f%% higher than stated confidence", magnitude*100)
		} else {
			bias.Description = fmt.Sprintf("Slightly underconfident - actual accuracy %.1f%% higher than stated confidence", magnitude*100)
		}
	} else {
		bias.Type = BiasOverconfident
		if magnitude > 0.15 {
			bias.Description = fmt.Sprintf("Significantly overconfident - actual accuracy %.1f%% lower than stated confidence", magnitude*100)
		} else {
			bias.Description = fmt.Sprintf("Slightly overconfident - actual accuracy %.1f%% lower than stated confidence", magnitude*100)
		}
	}

	return bias
}

// generateRecommendations creates actionable recommendations
func (ct *CalibrationTracker) generateRecommendations(report *CalibrationReport) []string {
	var recommendations []string

	// Bias recommendations
	switch report.Bias.Type {
	case BiasOverconfident:
		if report.Bias.Magnitude > 0.15 {
			recommendations = append(recommendations,
				"Significant overconfidence detected. Consider:",
				"- Reduce confidence scores by 10-15%",
				"- Increase use of validation tools before high-confidence claims",
				"- Challenge assumptions more frequently")
		} else {
			recommendations = append(recommendations,
				"Slight overconfidence detected. Consider reducing confidence by 5-10%.")
		}

	case BiasUnderconfident:
		if report.Bias.Magnitude > 0.15 {
			recommendations = append(recommendations,
				"Significant underconfidence detected. Consider:",
				"- Increase confidence scores by 10-15%",
				"- Trust validated reasoning more",
				"- Reduce unnecessary hedging")
		} else {
			recommendations = append(recommendations,
				"Slight underconfidence detected. Consider increasing confidence by 5-10%.")
		}

	case BiasNone:
		recommendations = append(recommendations,
			"Excellent calibration! Confidence scores match actual accuracy.")
	}

	// Mode-specific recommendations
	for mode, modeCalib := range report.ByMode {
		if modeCalib.OutcomeCount >= 10 {
			if modeCalib.Calibration > 0.1 {
				recommendations = append(recommendations,
					fmt.Sprintf("Mode '%s': Poorly calibrated (ECE=%.2f). Review confidence estimation for this mode.", mode, modeCalib.Calibration))
			}
		}
	}

	// Bucket-specific recommendations
	poorBuckets := []string{}
	for _, bucket := range report.Buckets {
		if bucket.Count >= 5 && math.Abs(bucket.Calibration) > 0.2 {
			poorBuckets = append(poorBuckets,
				fmt.Sprintf("%.0f%%-%.0f%%", bucket.MinConfidence*100, bucket.MaxConfidence*100))
		}
	}
	if len(poorBuckets) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Poor calibration in confidence ranges: %v", poorBuckets))
	}

	// Minimum data recommendation (add at end as context)
	if report.TotalOutcomes < 20 {
		recommendations = append(recommendations,
			fmt.Sprintf("Note: Only %d outcomes recorded. Collect 50+ for more reliable calibration.", report.TotalOutcomes))
	}

	return recommendations
}

// GetPrediction retrieves a prediction by thought ID
func (ct *CalibrationTracker) GetPrediction(thoughtID string) (*Prediction, error) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	pred, exists := ct.predictions[thoughtID]
	if !exists {
		return nil, fmt.Errorf("prediction not found for thought_id: %s", thoughtID)
	}

	return pred, nil
}

// GetOutcome retrieves an outcome by thought ID
func (ct *CalibrationTracker) GetOutcome(thoughtID string) (*Outcome, error) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	outcome, exists := ct.outcomes[thoughtID]
	if !exists {
		return nil, fmt.Errorf("outcome not found for thought_id: %s", thoughtID)
	}

	return outcome, nil
}

// ListPredictions returns all predictions, optionally filtered by mode
func (ct *CalibrationTracker) ListPredictions(mode string, limit int) []*Prediction {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	var predictions []*Prediction
	for _, pred := range ct.predictions {
		if mode == "" || pred.Mode == mode {
			predictions = append(predictions, pred)
		}
	}

	// Sort by timestamp descending
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Timestamp.After(predictions[j].Timestamp)
	})

	if limit > 0 && len(predictions) > limit {
		predictions = predictions[:limit]
	}

	return predictions
}

// Clear removes all predictions and outcomes
func (ct *CalibrationTracker) Clear() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.predictions = make(map[string]*Prediction)
	ct.outcomes = make(map[string]*Outcome)
}
