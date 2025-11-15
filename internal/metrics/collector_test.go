package metrics

import (
	"testing"
	"time"

	"unified-thinking/internal/types"
)

func TestNewCollectorDefaults(t *testing.T) {
	collector := NewCollector()

	if collector == nil {
		t.Fatal("expected collector instance")
	}

	if collector.windowSize != 24*time.Hour {
		t.Fatalf("unexpected window size: %v", collector.windowSize)
	}

	if len(collector.metrics) != 0 {
		t.Fatalf("expected empty metrics slice, got %d", len(collector.metrics))
	}

	if collector.toolUsage == nil {
		t.Fatal("expected toolUsage map to be initialized")
	}

	if collector.alertThresholds["accuracy"] != 0.85 {
		t.Fatalf("unexpected accuracy threshold: %v", collector.alertThresholds["accuracy"])
	}

	if collector.alertThresholds["success_rate"] != 0.95 {
		t.Fatalf("unexpected success_rate threshold: %v", collector.alertThresholds["success_rate"])
	}
}

func TestRecordMetric(t *testing.T) {
	collector := NewCollector()

	start := time.Now()
	collector.RecordMetric(MetricValue{Type: MetricAccuracy, Tool: "think", Value: 0.9, Target: 1.0})

	if len(collector.metrics) != 1 {
		t.Fatalf("expected 1 metric recorded, got %d", len(collector.metrics))
	}

	recorded := collector.metrics[0]
	if recorded.Timestamp.IsZero() {
		t.Fatal("expected timestamp to be set")
	}

	if recorded.Timestamp.Before(start) {
		t.Fatal("expected timestamp to be set after start")
	}

	if collector.toolUsage["think"] != 1 {
		t.Fatalf("expected tool usage tracked, got %d", collector.toolUsage["think"])
	}
}

func TestRecordThoughtValidation(t *testing.T) {
	collector := NewCollector()

	thought := &types.Thought{ID: "thought-1", Mode: types.ModeLinear}
	validation := &types.Validation{IsValid: true}

	collector.RecordThoughtValidation(thought, validation)

	if len(collector.metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(collector.metrics))
	}

	recorded := collector.metrics[0]
	if recorded.Type != MetricAccuracy {
		t.Fatalf("expected accuracy metric, got %s", recorded.Type)
	}

	if recorded.Value != 1.0 {
		t.Fatalf("expected accuracy value 1.0, got %v", recorded.Value)
	}

	if recorded.Target != 0.95 {
		t.Fatalf("expected target 0.95, got %v", recorded.Target)
	}

	if recorded.Context["thought_id"] != thought.ID {
		t.Fatalf("expected thought_id context, got %v", recorded.Context["thought_id"])
	}

	if recorded.Context["mode"] != thought.Mode {
		t.Fatalf("expected mode context %s, got %v", thought.Mode, recorded.Context["mode"])
	}
}
