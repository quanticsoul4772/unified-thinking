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
				ID:      "thought-1",
				Content: "This confirms our hypothesis and clearly shows we were right. Obviously this supports our view.",
				Mode:    types.ModeLinear,
				Confidence: 0.7,
				Timestamp: time.Now(),
			},
			expectBiasType: "confirmation",
		},
		{
			name: "overconfidence bias",
			thought: &types.Thought{
				ID:      "thought-2",
				Content: "This is definitely true without any doubt.",
				Mode:    types.ModeLinear,
				Confidence: 0.95,
				Timestamp: time.Now(),
			},
			expectBiasType: "overconfidence",
		},
		{
			name: "availability bias",
			thought: &types.Thought{
				ID:      "thought-3",
				Content: "I recently saw on the news and everyone knows this is common knowledge.",
				Mode:    types.ModeLinear,
				Confidence: 0.7,
				Timestamp: time.Now(),
			},
			expectBiasType: "availability",
		},
		{
			name: "no bias",
			thought: &types.Thought{
				ID:      "thought-4",
				Content: "Research data indicates this pattern, however alternative explanations should be considered.",
				Mode:    types.ModeLinear,
				Confidence: 0.6,
				Timestamp: time.Now(),
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
		ID:       "branch-1",
		State:    types.StateActive,
		Priority: 0.8,
		Confidence: 0.7,
		Thoughts: []*types.Thought{
			{
				ID:      "thought-1",
				Content: "This confirms our view.",
				Confidence: 0.8,
			},
			{
				ID:      "thought-2",
				Content: "This also confirms our view.",
				Confidence: 0.8,
			},
			{
				ID:      "thought-3",
				Content: "Yet another confirmation.",
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
