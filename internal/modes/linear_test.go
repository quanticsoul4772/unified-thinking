package modes

import (
	"context"
	"testing"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestNewLinearMode(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewLinearMode(store)

	if mode == nil {
		t.Fatal("NewLinearMode returned nil")
	}

	if mode.storage == nil {
		t.Error("LinearMode storage not initialized")
	}
}

func TestLinearMode_ProcessThought(t *testing.T) {
	tests := []struct {
		name    string
		input   ThoughtInput
		wantErr bool
	}{
		{
			name: "basic thought",
			input: ThoughtInput{
				Content:    "This is a linear thought",
				Type:       "observation",
				Confidence: 0.8,
			},
			wantErr: false,
		},
		{
			name: "thought with parent",
			input: ThoughtInput{
				Content:    "This follows from the previous thought",
				Type:       "analysis",
				ParentID:   "parent-123",
				Confidence: 0.9,
			},
			wantErr: false,
		},
		{
			name: "thought with zero confidence",
			input: ThoughtInput{
				Content:    "Uncertain thought",
				Type:       "hypothesis",
				Confidence: 0.0,
			},
			wantErr: false,
		},
		{
			name: "thought with high confidence",
			input: ThoughtInput{
				Content:    "Very certain thought",
				Type:       "conclusion",
				Confidence: 1.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			mode := NewLinearMode(store)
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

				if result.Mode != string(types.ModeLinear) {
					t.Errorf("Result mode = %v, want %v", result.Mode, types.ModeLinear)
				}

				if result.Status != "processed" {
					t.Errorf("Result status = %v, want processed", result.Status)
				}

				if result.Confidence != tt.input.Confidence {
					t.Errorf("Result confidence = %v, want %v", result.Confidence, tt.input.Confidence)
				}

				// Verify thought was stored
				thought, err := store.GetThought(result.ThoughtID)
				if err != nil {
					t.Errorf("Failed to retrieve stored thought: %v", err)
				}

				if thought.Content != tt.input.Content {
					t.Errorf("Stored thought content = %v, want %v", thought.Content, tt.input.Content)
				}

				if thought.Mode != types.ModeLinear {
					t.Errorf("Stored thought mode = %v, want %v", thought.Mode, types.ModeLinear)
				}

				if thought.Type != tt.input.Type {
					t.Errorf("Stored thought type = %v, want %v", thought.Type, tt.input.Type)
				}

				if thought.ParentID != tt.input.ParentID {
					t.Errorf("Stored thought parent = %v, want %v", thought.ParentID, tt.input.ParentID)
				}
			}
		})
	}
}

func TestLinearMode_GetHistory(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewLinearMode(store)
	ctx := context.Background()

	// Initially should return empty history
	history, err := mode.GetHistory(ctx)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}
	if len(history) != 0 {
		t.Errorf("Initial history length = %d, want 0", len(history))
	}

	// Add some thoughts
	inputs := []ThoughtInput{
		{Content: "First thought", Type: "observation", Confidence: 0.8},
		{Content: "Second thought", Type: "analysis", Confidence: 0.9},
		{Content: "Third thought", Type: "conclusion", Confidence: 0.95},
	}

	for _, input := range inputs {
		_, err := mode.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() error = %v", err)
		}
	}

	// Get history
	history, err = mode.GetHistory(ctx)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}

	if len(history) != len(inputs) {
		t.Errorf("History length = %d, want %d", len(history), len(inputs))
	}

	// Verify all thoughts are linear mode
	for _, thought := range history {
		if thought.Mode != types.ModeLinear {
			t.Errorf("Thought mode = %v, want %v", thought.Mode, types.ModeLinear)
		}
	}
}

func TestLinearMode_ProcessMultipleThoughts(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewLinearMode(store)
	ctx := context.Background()

	// Process a chain of thoughts
	var previousID string
	for i := 0; i < 5; i++ {
		input := ThoughtInput{
			Content:    "Sequential thought",
			Type:       "step",
			ParentID:   previousID,
			Confidence: 0.8,
		}

		result, err := mode.ProcessThought(ctx, input)
		if err != nil {
			t.Fatalf("ProcessThought() iteration %d error = %v", i, err)
		}

		previousID = result.ThoughtID
	}

	// Verify all thoughts are stored
	history, err := mode.GetHistory(ctx)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}

	if len(history) != 5 {
		t.Errorf("History length = %d, want 5", len(history))
	}
}

func TestLinearMode_ContextCancellation(t *testing.T) {
	store := storage.NewMemoryStorage()
	mode := NewLinearMode(store)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	input := ThoughtInput{
		Content:    "Test thought",
		Type:       "test",
		Confidence: 0.8,
	}

	// ProcessThought should still work (doesn't check context cancellation in current implementation)
	result, err := mode.ProcessThought(ctx, input)
	if err != nil {
		t.Errorf("ProcessThought() with cancelled context error = %v", err)
	}

	if result == nil {
		t.Error("ProcessThought() returned nil result")
	}
}
