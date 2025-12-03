package metacognition

import (
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// BiasCalibration tracks bias detection outcomes for auto-calibration
type BiasCalibration struct {
	mu               sync.RWMutex
	detectionRecords map[string][]BiasDetectionRecord // biasType -> records
	falsePositives   map[string]int                   // biasType -> false positive count
	truePositives    map[string]int                   // biasType -> true positive count
	suppressionRates map[string]float64               // biasType -> suppression threshold
}

// BiasDetectionRecord records a single bias detection for calibration
type BiasDetectionRecord struct {
	ID           string
	BiasType     string
	Severity     string
	WasConfirmed bool      // Whether the bias was confirmed (true positive) or not (false positive)
	ConfirmedAt  time.Time // When the outcome was recorded
	DetectedAt   time.Time
	ThoughtID    string
}

// NewBiasCalibration creates a new bias calibration tracker
func NewBiasCalibration() *BiasCalibration {
	return &BiasCalibration{
		detectionRecords: make(map[string][]BiasDetectionRecord),
		falsePositives:   make(map[string]int),
		truePositives:    make(map[string]int),
		suppressionRates: make(map[string]float64),
	}
}

// RecordDetection records a bias detection for calibration
func (bc *BiasCalibration) RecordDetection(bias *types.CognitiveBias) string {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	record := BiasDetectionRecord{
		ID:         bias.ID,
		BiasType:   bias.BiasType,
		Severity:   bias.Severity,
		DetectedAt: time.Now(),
		ThoughtID:  bias.DetectedIn,
	}

	bc.detectionRecords[bias.BiasType] = append(bc.detectionRecords[bias.BiasType], record)
	return record.ID
}

// ConfirmDetection marks a detection as confirmed (true positive) or not (false positive)
func (bc *BiasCalibration) ConfirmDetection(biasType, recordID string, wasConfirmed bool) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	records, exists := bc.detectionRecords[biasType]
	if !exists {
		return nil // No records for this type
	}

	for i, record := range records {
		if record.ID == recordID {
			records[i].WasConfirmed = wasConfirmed
			records[i].ConfirmedAt = time.Now()

			if wasConfirmed {
				bc.truePositives[biasType]++
			} else {
				bc.falsePositives[biasType]++
			}

			// Update suppression rate
			bc.updateSuppressionRate(biasType)
			return nil
		}
	}

	return nil
}

// updateSuppressionRate calculates the false positive rate for a bias type
func (bc *BiasCalibration) updateSuppressionRate(biasType string) {
	fp := bc.falsePositives[biasType]
	tp := bc.truePositives[biasType]
	total := fp + tp

	if total < 5 {
		// Not enough data yet
		bc.suppressionRates[biasType] = 0
		return
	}

	// False positive rate
	fpr := float64(fp) / float64(total)
	bc.suppressionRates[biasType] = fpr
}

// ShouldSuppress determines if a bias detection should be suppressed based on calibration
func (bc *BiasCalibration) ShouldSuppress(biasType, severity string) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	fpr, exists := bc.suppressionRates[biasType]
	if !exists {
		return false // No calibration data, don't suppress
	}

	// Suppression rules based on severity and false positive rate:
	// - High severity: Never suppress
	// - Medium severity: Suppress if FPR > 50%
	// - Low severity: Suppress if FPR > 30%
	switch severity {
	case "high":
		return false // Never suppress high severity
	case "medium":
		return fpr > 0.50
	case "low":
		return fpr > 0.30
	default:
		return fpr > 0.50 // Default to medium severity threshold
	}
}

// GetCalibrationStats returns calibration statistics for a bias type
func (bc *BiasCalibration) GetCalibrationStats(biasType string) *BiasCalibrationStats {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	records := bc.detectionRecords[biasType]
	fp := bc.falsePositives[biasType]
	tp := bc.truePositives[biasType]
	fpr := bc.suppressionRates[biasType]

	return &BiasCalibrationStats{
		BiasType:          biasType,
		TotalDetections:   len(records),
		TruePositives:     tp,
		FalsePositives:    fp,
		FalsePositiveRate: fpr,
		IsCalibrated:      (tp + fp) >= 5,
	}
}

// GetAllStats returns calibration statistics for all bias types
func (bc *BiasCalibration) GetAllStats() map[string]*BiasCalibrationStats {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	stats := make(map[string]*BiasCalibrationStats)
	for biasType := range bc.detectionRecords {
		stats[biasType] = &BiasCalibrationStats{
			BiasType:          biasType,
			TotalDetections:   len(bc.detectionRecords[biasType]),
			TruePositives:     bc.truePositives[biasType],
			FalsePositives:    bc.falsePositives[biasType],
			FalsePositiveRate: bc.suppressionRates[biasType],
			IsCalibrated:      (bc.truePositives[biasType] + bc.falsePositives[biasType]) >= 5,
		}
	}
	return stats
}

// BiasCalibrationStats contains calibration statistics for a bias type
type BiasCalibrationStats struct {
	BiasType          string  `json:"bias_type"`
	TotalDetections   int     `json:"total_detections"`
	TruePositives     int     `json:"true_positives"`
	FalsePositives    int     `json:"false_positives"`
	FalsePositiveRate float64 `json:"false_positive_rate"`
	IsCalibrated      bool    `json:"is_calibrated"`
}
