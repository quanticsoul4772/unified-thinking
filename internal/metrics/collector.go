// Package metrics provides quality measurement and tracking for the unified-thinking server
package metrics

import (
	"sync"
	"time"
	"unified-thinking/internal/types"
)

// MetricType represents different categories of metrics
type MetricType string

const (
	MetricAccuracy       MetricType = "accuracy"
	MetricCompleteness   MetricType = "completeness"
	MetricCoherence      MetricType = "coherence"
	MetricEfficiency     MetricType = "efficiency"
	MetricCalibration    MetricType = "calibration"
	MetricSoundness      MetricType = "soundness"
	MetricFallacyDetect  MetricType = "fallacy_detection"
	MetricCausalAccuracy MetricType = "causal_accuracy"
)

// MetricValue represents a single metric measurement
type MetricValue struct {
	Type      MetricType             `json:"type"`
	Tool      string                 `json:"tool"`
	Value     float64                `json:"value"`
	Target    float64                `json:"target"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// Collector manages metric collection and analysis
type Collector struct {
	mu              sync.RWMutex
	metrics         []MetricValue
	toolUsage       map[string]int
	alertThresholds map[string]float64
	windowSize      time.Duration
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		metrics:    make([]MetricValue, 0),
		toolUsage:  make(map[string]int),
		windowSize: 24 * time.Hour,
		alertThresholds: map[string]float64{
			"accuracy":     0.85,
			"completeness": 0.80,
			"calibration":  0.15,
			"success_rate": 0.95,
		},
	}
}

// RecordMetric records a new metric value
func (c *Collector) RecordMetric(metric MetricValue) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metric.Timestamp = time.Now()
	c.metrics = append(c.metrics, metric)

	if metric.Tool != "" {
		c.toolUsage[metric.Tool]++
	}
}

// RecordThoughtValidation records validation metrics
func (c *Collector) RecordThoughtValidation(thought *types.Thought, validation *types.Validation) {
	accuracy := 0.0
	if validation.IsValid {
		accuracy = 1.0
	}

	c.RecordMetric(MetricValue{
		Type:   MetricAccuracy,
		Tool:   "validate",
		Value:  accuracy,
		Target: 0.95,
		Context: map[string]interface{}{
			"thought_id": thought.ID,
			"mode":       thought.Mode,
		},
	})
}