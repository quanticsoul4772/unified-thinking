package modes

import (
	"context"
	"testing"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestNewAutoMode(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)

	auto := NewAutoMode(linear, tree, divergent)

	if auto == nil {
		t.Fatal("NewAutoMode returned nil")
	}

	if auto.linear == nil {
		t.Error("AutoMode linear mode not initialized")
	}

	if auto.tree == nil {
		t.Error("AutoMode tree mode not initialized")
	}

	if auto.divergent == nil {
		t.Error("AutoMode divergent mode not initialized")
	}
}

func TestAutoMode_DetectMode(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	tests := []struct {
		name       string
		input      ThoughtInput
		wantMode   types.ThinkingMode
	}{
		{
			name: "detect divergent by keyword creative",
			input: ThoughtInput{
				Content: "Let's think creatively about this problem",
			},
			wantMode: types.ModeDivergent,
		},
		{
			name: "detect divergent by keyword what if",
			input: ThoughtInput{
				Content: "What if we approached this differently?",
			},
			wantMode: types.ModeDivergent,
		},
		{
			name: "detect divergent by keyword imagine",
			input: ThoughtInput{
				Content: "Imagine a world where this works",
			},
			wantMode: types.ModeDivergent,
		},
		{
			name: "detect divergent by keyword challenge",
			input: ThoughtInput{
				Content: "Let's challenge the assumptions here",
			},
			wantMode: types.ModeDivergent,
		},
		{
			name: "detect divergent by keyword innovative",
			input: ThoughtInput{
				Content: "We need an innovative solution",
			},
			wantMode: types.ModeDivergent,
		},
		{
			name: "detect tree by branch ID",
			input: ThoughtInput{
				Content:  "Normal content",
				BranchID: "branch-123",
			},
			wantMode: types.ModeTree,
		},
		{
			name: "detect tree by cross refs",
			input: ThoughtInput{
				Content: "Normal content",
				CrossRefs: []CrossRefInput{
					{ToBranch: "b1", Type: "complementary"},
				},
			},
			wantMode: types.ModeTree,
		},
		{
			name: "detect tree by key points",
			input: ThoughtInput{
				Content:   "Normal content",
				KeyPoints: []string{"point1", "point2"},
			},
			wantMode: types.ModeTree,
		},
		{
			name: "detect tree by keyword branch",
			input: ThoughtInput{
				Content: "Let's branch this thought into alternatives",
			},
			wantMode: types.ModeTree,
		},
		{
			name: "detect tree by keyword explore",
			input: ThoughtInput{
				Content: "We should explore multiple options",
			},
			wantMode: types.ModeTree,
		},
		{
			name: "detect tree by keyword parallel",
			input: ThoughtInput{
				Content: "Consider these parallel approaches",
			},
			wantMode: types.ModeTree,
		},
		{
			name: "detect tree by keyword alternative",
			input: ThoughtInput{
				Content: "What are the alternative solutions?",
			},
			wantMode: types.ModeTree,
		},
		{
			name: "detect linear by default",
			input: ThoughtInput{
				Content: "This is a straightforward observation",
			},
			wantMode: types.ModeLinear,
		},
		{
			name: "detect linear for simple content",
			input: ThoughtInput{
				Content: "Step one is to analyze the data",
			},
			wantMode: types.ModeLinear,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := auto.detectMode(tt.input)

			if mode != tt.wantMode {
				t.Errorf("detectMode() = %v, want %v", mode, tt.wantMode)
			}
		})
	}
}

func TestAutoMode_ProcessThought(t *testing.T) {
	tests := []struct {
		name         string
		input        ThoughtInput
		expectedMode types.ThinkingMode
		wantErr      bool
	}{
		{
			name: "process divergent thought",
			input: ThoughtInput{
				Content:    "Let's think creatively about this",
				Type:       "creative",
				Confidence: 0.8,
			},
			expectedMode: types.ModeDivergent,
			wantErr:      false,
		},
		{
			name: "process tree thought with branch",
			input: ThoughtInput{
				Content:    "Analyzing multiple branches",
				Type:       "analysis",
				BranchID:   "", // Will create new branch
				Confidence: 0.85,
			},
			expectedMode: types.ModeTree,
			wantErr:      false,
		},
		{
			name: "process tree thought with key points",
			input: ThoughtInput{
				Content:    "Exploration with key insights",
				Type:       "exploration",
				KeyPoints:  []string{"point1", "point2"},
				Confidence: 0.9,
			},
			expectedMode: types.ModeTree,
			wantErr:      false,
		},
		{
			name: "process linear thought",
			input: ThoughtInput{
				Content:    "Simple sequential thought",
				Type:       "observation",
				Confidence: 0.8,
			},
			expectedMode: types.ModeLinear,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			linear := NewLinearMode(store)
			tree := NewTreeMode(store)
			divergent := NewDivergentMode(store)
			auto := NewAutoMode(linear, tree, divergent)
			ctx := context.Background()

			result, err := auto.ProcessThought(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessThought() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatal("ProcessThought() returned nil result")
				}

				if result.ThoughtID == "" {
					t.Error("Result missing ThoughtID")
				}

				// Verify the correct mode was used
				if result.Mode != string(tt.expectedMode) {
					t.Errorf("Result mode = %v, want %v", result.Mode, string(tt.expectedMode))
				}

				// Verify thought was stored
				thought, err := store.GetThought(result.ThoughtID)
				if err != nil {
					t.Errorf("Failed to retrieve stored thought: %v", err)
				}

				if thought.Mode != tt.expectedMode {
					t.Errorf("Stored thought mode = %v, want %v", thought.Mode, tt.expectedMode)
				}
			}
		})
	}
}

func TestAutoMode_ProcessThoughtModeSelection(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)
	ctx := context.Background()

	// Test that divergent keywords take priority
	t.Run("divergent takes priority over tree keywords", func(t *testing.T) {
		input := ThoughtInput{
			Content:    "Let's explore creative and innovative solutions",
			Confidence: 0.8,
		}

		result, err := auto.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() error = %v", err)
		}

		if result.Mode != string(types.ModeDivergent) {
			t.Errorf("Result mode = %v, want %v (divergent should take priority)", result.Mode, types.ModeDivergent)
		}
	})

	// Test that branch ID forces tree mode (if branch exists)
	t.Run("branch ID forces tree mode", func(t *testing.T) {
		// Create a branch first
		branch := &types.Branch{
			ID:    "test-branch",
			State: types.StateActive,
		}
		_ = store.StoreBranch(branch)

		input := ThoughtInput{
			Content:    "Simple content",
			BranchID:   "test-branch",
			Confidence: 0.8,
		}

		result, err := auto.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() error = %v", err)
		}

		if result.Mode != string(types.ModeTree) {
			t.Errorf("Result mode = %v, want %v (branch ID should force tree)", result.Mode, types.ModeTree)
		}
	})

	// Test default to linear
	t.Run("defaults to linear for neutral content", func(t *testing.T) {
		input := ThoughtInput{
			Content:    "A neutral observation about data",
			Confidence: 0.8,
		}

		result, err := auto.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() error = %v", err)
		}

		if result.Mode != string(types.ModeLinear) {
			t.Errorf("Result mode = %v, want %v (should default to linear)", result.Mode, types.ModeLinear)
		}
	})
}

func TestAutoMode_CaseInsensitiveDetection(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	tests := []struct {
		content  string
		wantMode types.ThinkingMode
	}{
		{"CREATIVE solution needed", types.ModeDivergent},
		{"Creative Solution Needed", types.ModeDivergent},
		{"creative SOLUTION needed", types.ModeDivergent},
		{"EXPLORE multiple options", types.ModeTree},
		{"Explore Multiple Options", types.ModeTree},
		{"explore MULTIPLE options", types.ModeTree},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			input := ThoughtInput{Content: tt.content}
			mode := auto.detectMode(input)

			if mode != tt.wantMode {
				t.Errorf("detectMode(%s) = %v, want %v", tt.content, mode, tt.wantMode)
			}
		})
	}
}

func TestAutoMode_MultipleKeywords(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)
	ctx := context.Background()

	// Content with both divergent and tree keywords
	input := ThoughtInput{
		Content:    "We need to explore creative and innovative alternatives",
		Confidence: 0.8,
	}

	result, err := auto.ProcessThought(ctx, input)
	if err != nil {
		t.Fatalf("ProcessThought() error = %v", err)
	}

	// Divergent keywords should take priority
	if result.Mode != string(types.ModeDivergent) {
		t.Errorf("Result mode = %v, want %v (divergent should have priority)", result.Mode, types.ModeDivergent)
	}
}

func TestAutoMode_EmptyContent(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	input := ThoughtInput{
		Content: "",
	}

	mode := auto.detectMode(input)

	// Should default to linear for empty content
	if mode != types.ModeLinear {
		t.Errorf("detectMode() for empty content = %v, want %v", mode, types.ModeLinear)
	}
}

func TestAutoMode_AllDivergentKeywords(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	// Note: "different" removed to avoid false positives with tree-mode phrases
	// like "explore different options" or "compare different alternatives"
	keywords := []string{
		"creative", "unconventional", "what if", "imagine", "challenge",
		"rebel", "outside the box", "innovative", "radical",
	}

	for _, keyword := range keywords {
		t.Run("keyword_"+keyword, func(t *testing.T) {
			input := ThoughtInput{
				Content: "Let's use " + keyword + " thinking",
			}

			mode := auto.detectMode(input)

			if mode != types.ModeDivergent {
				t.Errorf("detectMode() with keyword '%s' = %v, want %v", keyword, mode, types.ModeDivergent)
			}
		})
	}
}

func TestAutoMode_AllTreeKeywords(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	keywords := []string{
		"branch", "explore", "alternative", "parallel", "compare",
		"multiple", "options", "possibilities",
	}

	for _, keyword := range keywords {
		t.Run("keyword_"+keyword, func(t *testing.T) {
			input := ThoughtInput{
				Content: "Consider " + keyword + " approaches",
			}

			mode := auto.detectMode(input)

			if mode != types.ModeTree {
				t.Errorf("detectMode() with keyword '%s' = %v, want %v", keyword, mode, types.ModeTree)
			}
		})
	}
}

func TestAutoMode_TreeModeIndicators(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	tests := []struct {
		name  string
		input ThoughtInput
	}{
		{
			name: "branch ID indicator",
			input: ThoughtInput{
				Content:  "Normal content",
				BranchID: "branch-123",
			},
		},
		{
			name: "cross refs indicator",
			input: ThoughtInput{
				Content: "Normal content",
				CrossRefs: []CrossRefInput{
					{ToBranch: "b1"},
				},
			},
		},
		{
			name: "key points indicator",
			input: ThoughtInput{
				Content:   "Normal content",
				KeyPoints: []string{"point1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := auto.detectMode(tt.input)

			if mode != types.ModeTree {
				t.Errorf("detectMode() = %v, want %v (should detect tree mode)", mode, types.ModeTree)
			}
		})
	}
}
