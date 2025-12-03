package metacognition

import (
	"testing"
	"time"

	"unified-thinking/internal/types"
)

func TestNewBiasDetector(t *testing.T) {
	bd := NewBiasDetector()
	if bd == nil {
		t.Fatal("NewBiasDetector returned nil")
	}
}

func TestDetectBiases(t *testing.T) {
	bd := NewBiasDetector()

	tests := []struct {
		name           string
		thought        *types.Thought
		expectBiasType string
	}{
		{
			name: "confirmation bias",
			thought: &types.Thought{
				ID:         "thought-1",
				Content:    "This confirms our hypothesis and clearly shows we were right. Obviously this supports our view.",
				Mode:       types.ModeLinear,
				Confidence: 0.7,
				Timestamp:  time.Now(),
			},
			expectBiasType: "confirmation",
		},
		{
			name: "overconfidence bias",
			thought: &types.Thought{
				ID:         "thought-2",
				Content:    "This is definitely true without any doubt.",
				Mode:       types.ModeLinear,
				Confidence: 0.95,
				Timestamp:  time.Now(),
			},
			expectBiasType: "overconfidence",
		},
		{
			name: "availability bias",
			thought: &types.Thought{
				ID:         "thought-3",
				Content:    "I recently saw on the news and everyone knows this is common knowledge.",
				Mode:       types.ModeLinear,
				Confidence: 0.7,
				Timestamp:  time.Now(),
			},
			expectBiasType: "availability",
		},
		{
			name: "no bias",
			thought: &types.Thought{
				ID:         "thought-4",
				Content:    "Research data indicates this pattern, however alternative explanations should be considered.",
				Mode:       types.ModeLinear,
				Confidence: 0.6,
				Timestamp:  time.Now(),
			},
			expectBiasType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			biases, err := bd.DetectBiases(tt.thought)
			if err != nil {
				t.Errorf("DetectBiases() error = %v", err)
				return
			}

			if tt.expectBiasType == "" {
				if len(biases) > 0 {
					t.Errorf("Expected no biases, but found %v", len(biases))
				}
			} else {
				found := false
				for _, bias := range biases {
					if bias.BiasType == tt.expectBiasType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected bias type %v not found in detected biases", tt.expectBiasType)
				}
			}
		})
	}
}

func TestDetectBiasesInBranch(t *testing.T) {
	bd := NewBiasDetector()

	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   0.8,
		Confidence: 0.7,
		Thoughts: []*types.Thought{
			{
				ID:         "thought-1",
				Content:    "This confirms our view.",
				Confidence: 0.8,
			},
			{
				ID:         "thought-2",
				Content:    "This also confirms our view.",
				Confidence: 0.8,
			},
			{
				ID:         "thought-3",
				Content:    "Yet another confirmation.",
				Confidence: 0.8,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	biases, err := bd.DetectBiasesInBranch(branch)
	if err != nil {
		t.Fatalf("DetectBiasesInBranch() error = %v", err)
	}

	// Should detect some biases given the repetitive confirming language
	if len(biases) == 0 {
		t.Error("Expected to detect biases in branch with repetitive confirming language")
	}
}

// Tests for bias calibration functionality (Phase 1.3 implementation)

func TestBiasCalibration_RecordAndConfirm(t *testing.T) {
	cal := NewBiasCalibration()

	bias := &types.CognitiveBias{
		ID:         "bias-1",
		BiasType:   "confirmation",
		Severity:   "medium",
		DetectedIn: "thought-1",
	}

	// Record detection
	id := cal.RecordDetection(bias)
	if id == "" {
		t.Error("RecordDetection should return non-empty ID")
	}

	// Confirm as true positive
	err := cal.ConfirmDetection("confirmation", id, true)
	if err != nil {
		t.Errorf("ConfirmDetection error: %v", err)
	}

	// Check stats
	stats := cal.GetCalibrationStats("confirmation")
	if stats.TruePositives != 1 {
		t.Errorf("Expected 1 true positive, got %d", stats.TruePositives)
	}
}

func TestBiasCalibration_FalsePositiveTracking(t *testing.T) {
	cal := NewBiasCalibration()

	// Record 10 detections, 7 false positives
	for i := 0; i < 10; i++ {
		bias := &types.CognitiveBias{
			ID:         "bias-" + string(rune('a'+i)),
			BiasType:   "anchoring",
			Severity:   "low",
			DetectedIn: "thought-" + string(rune('a'+i)),
		}
		id := cal.RecordDetection(bias)
		// 7 false positives, 3 true positives
		cal.ConfirmDetection("anchoring", id, i >= 7)
	}

	stats := cal.GetCalibrationStats("anchoring")
	if stats.FalsePositives != 7 {
		t.Errorf("Expected 7 false positives, got %d", stats.FalsePositives)
	}
	if stats.TruePositives != 3 {
		t.Errorf("Expected 3 true positives, got %d", stats.TruePositives)
	}

	// FPR should be 0.7 (70%)
	expectedFPR := 0.7
	if stats.FalsePositiveRate < 0.69 || stats.FalsePositiveRate > 0.71 {
		t.Errorf("Expected FPR ~0.7, got %f", stats.FalsePositiveRate)
	}
	_ = expectedFPR
}

func TestBiasCalibration_Suppression(t *testing.T) {
	cal := NewBiasCalibration()

	// Record 10 detections with high false positive rate (80%)
	for i := 0; i < 10; i++ {
		bias := &types.CognitiveBias{
			ID:         "bias-" + string(rune('a'+i)),
			BiasType:   "recency",
			Severity:   "low",
			DetectedIn: "thought-" + string(rune('a'+i)),
		}
		id := cal.RecordDetection(bias)
		// 8 false positives, 2 true positives = 80% FPR
		cal.ConfirmDetection("recency", id, i >= 8)
	}

	// Low severity with 80% FPR should be suppressed (threshold is 30%)
	if !cal.ShouldSuppress("recency", "low") {
		t.Error("Expected low severity bias with 80% FPR to be suppressed")
	}

	// High severity should never be suppressed
	if cal.ShouldSuppress("recency", "high") {
		t.Error("High severity bias should never be suppressed")
	}
}

func TestBiasDetector_WithCalibration(t *testing.T) {
	bd := NewBiasDetector()

	// Get calibration should return non-nil
	cal := bd.GetCalibration()
	if cal == nil {
		t.Fatal("GetCalibration should return non-nil")
	}

	// Detect biases should record to calibration
	thought := &types.Thought{
		ID:         "test-thought",
		Content:    "This confirms our view and supports our hypothesis as expected.",
		Mode:       types.ModeLinear,
		Confidence: 0.7,
		Timestamp:  time.Now(),
	}

	biases, err := bd.DetectBiases(thought)
	if err != nil {
		t.Fatalf("DetectBiases error: %v", err)
	}

	// Should detect confirmation bias
	if len(biases) == 0 {
		t.Error("Expected to detect confirmation bias")
	}

	// Check calibration recorded the detection
	stats := cal.GetCalibrationStats("confirmation")
	if stats.TotalDetections == 0 {
		t.Error("Expected calibration to record the detection")
	}
}
