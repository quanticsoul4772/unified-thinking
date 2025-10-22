package modes

import (
	"context"
	"strings"
	"testing"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestNewDivergentMode(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)

	if mode == nil {
		t.Fatal("NewDivergentMode returned nil")
	}

	if mode.storage == nil {
		t.Error("DivergentMode storage not initialized")
	}
}

func TestDivergentMode_ProcessThought(t *testing.T) {
	tests := []struct {
		name           string
		input          ThoughtInput
		wantErr        bool
		checkRebellion bool
	}{
		{
			name: "basic divergent thought",
			input: ThoughtInput{
				Content:    "problem to solve",
				Type:       "creative",
				Confidence: 0.8,
			},
			wantErr: false,
		},
		{
			name: "forced rebellion",
			input: ThoughtInput{
				Content:        "conventional problem",
				Type:           "rebellious",
				Confidence:     0.7,
				ForceRebellion: true,
			},
			wantErr:        false,
			checkRebellion: true,
		},
		{
			name: "thought with previous ID",
			input: ThoughtInput{
				Content:           "divergent continuation",
				Type:              "creative",
				PreviousThoughtID: "prev-123",
				Confidence:        0.85,
			},
			wantErr: false,
		},
		{
			name: "high confidence divergent",
			input: ThoughtInput{
				Content:    "innovative idea",
				Type:       "innovation",
				Confidence: 1.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			mode := NewDivergentMode(store)
			ctx := context.Background()

			result, err := mode.ProcessThought(ctx, tt.input)

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

				if result.Mode != string(types.ModeDivergent) {
					t.Errorf("Result mode = %v, want %v", result.Mode, types.ModeDivergent)
				}

				if result.Content == "" {
					t.Error("Result missing content")
				}

				// Verify creative transformation occurred
				if result.Content == tt.input.Content {
					t.Error("Content should be transformed creatively")
				}

				if tt.checkRebellion && !result.IsRebellion {
					t.Error("ForceRebellion should set IsRebellion to true")
				}

				// Verify thought was stored
				thought, err := store.GetThought(result.ThoughtID)
				if err != nil {
					t.Errorf("Failed to retrieve stored thought: %v", err)
				}

				if thought.Mode != types.ModeDivergent {
					t.Errorf("Stored thought mode = %v, want %v", thought.Mode, types.ModeDivergent)
				}

				if thought.Type != tt.input.Type {
					t.Errorf("Stored thought type = %v, want %v", thought.Type, tt.input.Type)
				}

				if tt.input.PreviousThoughtID != "" {
					if thought.ParentID != tt.input.PreviousThoughtID {
						t.Errorf("Stored thought parent = %v, want %v", thought.ParentID, tt.input.PreviousThoughtID)
					}
				}
			}
		})
	}
}

func TestDivergentMode_GenerateCreativeThought(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)

	tests := []struct {
		name           string
		problem        string
		forceRebellion bool
	}{
		{
			name:           "creative without rebellion",
			problem:        "test problem",
			forceRebellion: false,
		},
		{
			name:           "creative with forced rebellion",
			problem:        "test problem",
			forceRebellion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creative := mode.generateCreativeThought(tt.problem, tt.forceRebellion)

			if creative == "" {
				t.Error("generateCreativeThought() returned empty string")
			}

			if !strings.Contains(creative, tt.problem) {
				t.Errorf("Creative thought should reference the problem: %s", creative)
			}

			// Generate multiple times to ensure variety is possible
			seen := make(map[string]bool)
			for i := 0; i < 10; i++ {
				creative := mode.generateCreativeThought(tt.problem, tt.forceRebellion)
				seen[creative] = true
			}

			// Should have some variety (at least 2 different outputs in 10 tries)
			if len(seen) < 2 {
				t.Error("generateCreativeThought() should produce variety")
			}
		})
	}
}

func TestDivergentMode_BranchThought(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	// Create a source thought
	sourceThought := &types.Thought{
		ID:         "source-123",
		Content:    "Original idea",
		Mode:       types.ModeDivergent,
		Type:       "creative",
		Confidence: 0.8,
	}
	_ = store.StoreThought(sourceThought)

	directions := []string{"more_extreme", "opposite", "tangential", "simplified", "combined", "unknown"}

	for _, direction := range directions {
		t.Run("direction_"+direction, func(t *testing.T) {
			result, err := mode.BranchThought(ctx, "source-123", direction)

			if err != nil {
				t.Errorf("BranchThought() error = %v", err)
				return
			}

			if result == nil {
				t.Fatal("BranchThought() returned nil result")
			}

			if result.ThoughtID == "" {
				t.Error("Result missing ThoughtID")
			}

			if result.Content == "" {
				t.Error("Result missing content")
			}

			if result.Direction != direction {
				t.Errorf("Result direction = %v, want %v", result.Direction, direction)
			}

			// Verify thought was stored
			thought, err := store.GetThought(result.ThoughtID)
			if err != nil {
				t.Errorf("Failed to retrieve branched thought: %v", err)
			}

			if thought.ParentID != "source-123" {
				t.Errorf("Branched thought parent = %v, want source-123", thought.ParentID)
			}

			if thought.Type != "branched_"+direction {
				t.Errorf("Branched thought type = %v, want branched_%v", thought.Type, direction)
			}

			// Check rebellion flag for opposite direction
			if direction == "opposite" && !thought.IsRebellion {
				t.Error("Opposite direction should set IsRebellion to true")
			}

			// All branched thoughts should challenge assumptions
			if !thought.ChallengesAssumption {
				t.Error("Branched thought should challenge assumptions")
			}
		})
	}
}

func TestDivergentMode_BranchThoughtNonExistent(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	// Try to branch from non-existent thought
	_, err := mode.BranchThought(ctx, "non-existent", "opposite")

	if err == nil {
		t.Error("BranchThought() should return error for non-existent thought")
	}
}

func TestDivergentMode_ListThoughts(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	// Initially should return empty list
	thoughts, err := mode.ListThoughts(ctx)
	if err != nil {
		t.Errorf("ListThoughts() error = %v", err)
	}
	if len(thoughts) != 0 {
		t.Errorf("Initial thoughts length = %d, want 0", len(thoughts))
	}

	// Process some divergent thoughts
	for i := 0; i < 3; i++ {
		input := ThoughtInput{
			Content:    "problem",
			Type:       "creative",
			Confidence: 0.8,
		}
		_, _ = mode.ProcessThought(ctx, input)
	}

	// List thoughts
	thoughts, err = mode.ListThoughts(ctx)
	if err != nil {
		t.Errorf("ListThoughts() error = %v", err)
	}

	if len(thoughts) != 3 {
		t.Errorf("Thoughts length = %d, want 3", len(thoughts))
	}

	// Verify all thoughts are divergent mode
	for _, thought := range thoughts {
		if thought.Mode != types.ModeDivergent {
			t.Errorf("Thought mode = %v, want %v", thought.Mode, types.ModeDivergent)
		}
	}
}

func TestDivergentMode_GenerateBranchedThought(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)

	source := &types.Thought{
		Content: "Original thought",
	}

	directions := map[string]string{
		"more_extreme": "extreme",
		"opposite":     "reversal",
		"tangential":   "connection",
		"simplified":   "simplification",
		"combined":     "Synthesis",
		"unknown":      "direction",
	}

	for direction, expectedWord := range directions {
		t.Run(direction, func(t *testing.T) {
			branched := mode.generateBranchedThought(source, direction)

			if branched == "" {
				t.Error("generateBranchedThought() returned empty string")
			}

			if !strings.Contains(branched, expectedWord) {
				t.Errorf("Branched thought for %s should contain '%s': %s", direction, expectedWord, branched)
			}
		})
	}
}

func TestDivergentMode_MultipleProcessing(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	// Process multiple thoughts and verify they're all different
	contents := make(map[string]bool)

	for i := 0; i < 5; i++ {
		input := ThoughtInput{
			Content:    "same problem",
			Type:       "creative",
			Confidence: 0.8,
		}

		result, err := mode.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() iteration %d error = %v", i, err)
		}

		contents[result.Content] = true
	}

	// Should have variety in creative outputs
	if len(contents) < 2 {
		t.Error("Multiple ProcessThought calls should produce varied creative outputs")
	}
}

func TestDivergentMode_FlagsVariability(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewDivergentMode(store)
	ctx := context.Background()

	rebellionSeen := false
	challengesSeen := false

	// Process many thoughts to test randomization
	for i := 0; i < 20; i++ {
		input := ThoughtInput{
			Content:        "problem",
			Type:           "creative",
			Confidence:     0.8,
			ForceRebellion: false, // Not forced
		}

		result, err := mode.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() error = %v", err)
		}

		thought, _ := store.GetThought(result.ThoughtID)

		if thought.IsRebellion {
			rebellionSeen = true
		}

		if thought.ChallengesAssumption {
			challengesSeen = true
		}

		if rebellionSeen && challengesSeen {
			break
		}
	}

	// With randomization, we should see variety in flags
	if !challengesSeen {
		t.Error("ChallengesAssumption should be set for some thoughts")
	}
}
