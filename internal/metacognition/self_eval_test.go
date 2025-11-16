package metacognition

import (
	"testing"
	"time"

	"unified-thinking/internal/types"
)

func TestNewSelfEvaluator(t *testing.T) {
	se := NewSelfEvaluator()
	if se == nil {
		t.Fatal("NewSelfEvaluator returned nil")
	}
}

func TestEvaluateThought(t *testing.T) {
	se := NewSelfEvaluator()

	tests := []struct {
		name            string
		thought         *types.Thought
		minQuality      float64
		minCompleteness float64
		minCoherence    float64
	}{
		{
			name: "high quality thought",
			thought: &types.Thought{
				ID:         "thought-1",
				Content:    "Based on the evidence from multiple studies and data analysis, we can therefore conclude that this approach is effective. However, we should also consider alternative explanations.",
				Mode:       types.ModeLinear,
				Confidence: 0.7,
				KeyPoints:  []string{"Evidence-based", "Multiple studies", "Alternative explanations"},
				Timestamp:  time.Now(),
			},
			minQuality:      0.7,
			minCompleteness: 0.6,
			minCoherence:    0.7,
		},
		{
			name: "low quality thought",
			thought: &types.Thought{
				ID:         "thought-2",
				Content:    "Maybe it works",
				Mode:       types.ModeLinear,
				Confidence: 0.9,
				KeyPoints:  []string{},
				Timestamp:  time.Now(),
			},
			minQuality:      0.0,
			minCompleteness: 0.0,
			minCoherence:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eval, err := se.EvaluateThought(tt.thought)
			if err != nil {
				t.Errorf("EvaluateThought() error = %v", err)
				return
			}
			if eval == nil {
				t.Fatal("EvaluateThought() returned nil")
			}

			if eval.QualityScore < tt.minQuality {
				t.Errorf("QualityScore = %v, want >= %v", eval.QualityScore, tt.minQuality)
			}
			if eval.CompletenessScore < tt.minCompleteness {
				t.Errorf("CompletenessScore = %v, want >= %v", eval.CompletenessScore, tt.minCompleteness)
			}
			if eval.CoherenceScore < tt.minCoherence {
				t.Errorf("CoherenceScore = %v, want >= %v", eval.CoherenceScore, tt.minCoherence)
			}

			// Check score ranges
			if eval.QualityScore < 0 || eval.QualityScore > 1 {
				t.Errorf("QualityScore %v out of range [0,1]", eval.QualityScore)
			}
			if eval.CompletenessScore < 0 || eval.CompletenessScore > 1 {
				t.Errorf("CompletenessScore %v out of range [0,1]", eval.CompletenessScore)
			}
			if eval.CoherenceScore < 0 || eval.CoherenceScore > 1 {
				t.Errorf("CoherenceScore %v out of range [0,1]", eval.CoherenceScore)
			}
		})
	}
}

func TestEvaluateBranch(t *testing.T) {
	se := NewSelfEvaluator()

	branch := &types.Branch{
		ID:         "branch-1",
		State:      types.StateActive,
		Priority:   0.8,
		Confidence: 0.7,
		Thoughts: []*types.Thought{
			{
				ID:        "thought-1",
				Content:   "First analysis with evidence and data",
				KeyPoints: []string{"Evidence", "Data"},
			},
			{
				ID:        "thought-2",
				Content:   "Second analysis building on the first",
				KeyPoints: []string{"Building"},
			},
		},
		Insights: []*types.Insight{
			{
				ID:      "insight-1",
				Content: "Key insight discovered",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	eval, err := se.EvaluateBranch(branch)
	if err != nil {
		t.Fatalf("EvaluateBranch() error = %v", err)
	}
	if eval == nil {
		t.Fatal("EvaluateBranch() returned nil")
	}

	// Check that scores are in valid range
	if eval.QualityScore < 0 || eval.QualityScore > 1 {
		t.Errorf("QualityScore %v out of range [0,1]", eval.QualityScore)
	}
	if eval.CompletenessScore < 0 || eval.CompletenessScore > 1 {
		t.Errorf("CompletenessScore %v out of range [0,1]", eval.CompletenessScore)
	}
	if eval.CoherenceScore < 0 || eval.CoherenceScore > 1 {
		t.Errorf("CoherenceScore %v out of range [0,1]", eval.CoherenceScore)
	}

	// Should have at least one strength
	if len(eval.Strengths) == 0 && eval.QualityScore > 0.7 {
		t.Error("High quality branch should have identified strengths")
	}
}
