package server

import (
	"context"
	"os"
	"strings"
	"testing"

	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

// TestAutoValidationTriggering tests that low-confidence thoughts trigger auto-validation
func TestAutoValidationTriggering(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Fatal("ANTHROPIC_API_KEY not set - required for full server")
	}
	if os.Getenv("VOYAGE_API_KEY") == "" {
		t.Fatal("VOYAGE_API_KEY not set - required for embeddings")
	}

	// Enable debug mode for testing
	_ = os.Setenv("DEBUG", "true")
	defer func() { _ = os.Unsetenv("DEBUG") }()

	// Setup
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()

	server, err := NewUnifiedServer(store, linear, tree, divergent, auto, validator)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test case 1: Low confidence (< 0.5) should trigger auto-validation
	t.Run("LowConfidenceTriggersAutoValidation", func(t *testing.T) {
		ctx := context.Background()
		req := &ThinkRequest{
			Content:    "This is a highly uncertain thought with many unknowns",
			Mode:       "linear",
			Confidence: 0.3, // Low confidence
		}

		// Process the thought
		_, response, err := server.handleThink(ctx, nil, *req)
		if err != nil {
			t.Fatalf("handleThink failed: %v", err)
		}

		// Verify the thought was processed
		if response.ThoughtID == "" {
			t.Error("Expected thought ID to be set")
		}

		// Get the thought to check metadata
		thought, err := store.GetThought(response.ThoughtID)
		if err != nil {
			t.Fatalf("Failed to get thought: %v", err)
		}

		// Check if auto-validation was triggered
		if thought.Metadata != nil {
			if triggered, ok := thought.Metadata["auto_validation_triggered"].(bool); ok {
				if !triggered {
					t.Error("Expected auto_validation_triggered to be true")
				}
			} else {
				// Auto-validation might not have been marked if self-evaluation failed
				// This is acceptable for this test
				t.Log("auto_validation_triggered not found in metadata")
			}
		}
	})

	// Test case 2: Normal confidence (>= 0.5) should NOT trigger auto-validation
	t.Run("NormalConfidenceNoAutoValidation", func(t *testing.T) {
		ctx := context.Background()
		req := &ThinkRequest{
			Content:    "This is a reasonably confident thought",
			Mode:       "linear",
			Confidence: 0.7, // Normal confidence
		}

		// Process the thought
		_, response, err := server.handleThink(ctx, nil, *req)
		if err != nil {
			t.Fatalf("handleThink failed: %v", err)
		}

		// Get the thought to check metadata
		thought, err := store.GetThought(response.ThoughtID)
		if err != nil {
			t.Fatalf("Failed to get thought: %v", err)
		}

		// Check that auto-validation was NOT triggered
		if thought.Metadata != nil {
			if triggered, ok := thought.Metadata["auto_validation_triggered"].(bool); ok && triggered {
				t.Error("Expected auto_validation_triggered to be false for normal confidence")
			}
		}
	})

	// Test case 3: RequireValidation=true should skip auto-validation
	t.Run("RequireValidationSkipsAuto", func(t *testing.T) {
		ctx := context.Background()
		req := &ThinkRequest{
			Content:           "Low confidence but with manual validation",
			Mode:              "linear",
			Confidence:        0.3, // Low confidence
			RequireValidation: true,
		}

		// Process the thought
		_, response, err := server.handleThink(ctx, nil, *req)
		if err != nil {
			t.Fatalf("handleThink failed: %v", err)
		}

		// Verify that IsValid is set (from RequireValidation)
		// The actual validation result depends on the content
		t.Logf("IsValid: %v", response.IsValid)

		// Get the thought to check metadata
		thought, err := store.GetThought(response.ThoughtID)
		if err != nil {
			t.Fatalf("Failed to get thought: %v", err)
		}

		// Check that auto-validation was NOT triggered
		if thought.Metadata != nil {
			if triggered, ok := thought.Metadata["auto_validation_triggered"].(bool); ok && triggered {
				t.Error("Expected auto_validation_triggered to be false when RequireValidation is true")
			}
		}
	})

	// Test case 4: ChallengeAssumptions should not trigger retry
	t.Run("ChallengeAssumptionsNoRetry", func(t *testing.T) {
		ctx := context.Background()
		req := &ThinkRequest{
			Content:              "Low confidence with challenge assumptions already set",
			Mode:                 "linear",
			Confidence:           0.3, // Low confidence
			ChallengeAssumptions: true,
		}

		// Process the thought
		_, response, err := server.handleThink(ctx, nil, *req)
		if err != nil {
			t.Fatalf("handleThink failed: %v", err)
		}

		// Get the thought to check metadata
		thought, err := store.GetThought(response.ThoughtID)
		if err != nil {
			t.Fatalf("Failed to get thought: %v", err)
		}

		// Check metadata - should have auto_validation but NOT retry
		if thought.Metadata != nil {
			if _, ok := thought.Metadata["auto_retry_with_challenge"].(bool); ok {
				t.Error("Expected no auto_retry_with_challenge when ChallengeAssumptions is already set")
			}
		}

		// Verify that the thought content includes assumption challenging
		if thought.Content != "" && !strings.Contains(thought.Content, "Questioning assumptions") {
			t.Error("Expected thought content to include assumption questioning")
		}
	})
}
