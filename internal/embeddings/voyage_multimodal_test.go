// Package embeddings - Tests for VoyageMultimodalEmbedder
package embeddings

import (
	"context"
	"os"
	"testing"
)

func TestNewVoyageMultimodalEmbedder(t *testing.T) {
	tests := []struct {
		name            string
		apiKey          string
		textModel       string
		multimodalModel string
		wantMM          string
	}{
		{
			name:            "with all params",
			apiKey:          "test-key",
			textModel:       "voyage-3-lite",
			multimodalModel: "voyage-multimodal-3",
			wantMM:          "voyage-multimodal-3",
		},
		{
			name:            "with empty multimodal model uses default",
			apiKey:          "test-key",
			textModel:       "voyage-3",
			multimodalModel: "",
			wantMM:          "voyage-multimodal-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewVoyageMultimodalEmbedder(tt.apiKey, tt.textModel, tt.multimodalModel)

			if e == nil {
				t.Fatal("expected non-nil embedder")
			}

			if e.MultimodalModel() != tt.wantMM {
				t.Errorf("MultimodalModel() = %v, want %v", e.MultimodalModel(), tt.wantMM)
			}

			if !e.SupportsMultimodal() {
				t.Error("SupportsMultimodal() should return true")
			}

			if e.MultimodalDimension() != 1024 {
				t.Errorf("MultimodalDimension() = %v, want 1024", e.MultimodalDimension())
			}
		})
	}
}

func TestVoyageMultimodalEmbedder_EmbedImage_Empty(t *testing.T) {
	e := NewVoyageMultimodalEmbedder("test-key", "voyage-3-lite", "")

	_, err := e.EmbedImage(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty image data")
	}
}

func TestVoyageMultimodalEmbedder_EmbedImageURL_Empty(t *testing.T) {
	e := NewVoyageMultimodalEmbedder("test-key", "voyage-3-lite", "")

	_, err := e.EmbedImageURL(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty image URL")
	}
}

func TestVoyageMultimodalEmbedder_EmbedMultimodal_NoInputs(t *testing.T) {
	e := NewVoyageMultimodalEmbedder("test-key", "voyage-3-lite", "")

	_, err := e.EmbedMultimodal(context.Background(), []MultimodalInput{})
	if err == nil {
		t.Error("expected error for empty inputs")
	}
}

func TestMultimodalInput_ToAPIFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    MultimodalInput
		wantType string
		wantKey  string
	}{
		{
			name: "text input",
			input: MultimodalInput{
				Type: InputTypeText,
				Text: "hello world",
			},
			wantType: "text",
			wantKey:  "text",
		},
		{
			name: "image base64 input",
			input: MultimodalInput{
				Type:     InputTypeImageBase64,
				ImageB64: "iVBORw0KGgo...",
			},
			wantType: "image_base64",
			wantKey:  "image_base64",
		},
		{
			name: "image URL input",
			input: MultimodalInput{
				Type:     InputTypeImageURL,
				ImageURL: "https://example.com/image.png",
			},
			wantType: "image_url",
			wantKey:  "image_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToAPIFormat()

			if result["type"] != tt.wantType {
				t.Errorf("type = %v, want %v", result["type"], tt.wantType)
			}

			if _, ok := result[tt.wantKey]; !ok {
				t.Errorf("expected key %v in result", tt.wantKey)
			}
		})
	}
}

// Integration test - requires VOYAGE_API_KEY
func TestVoyageMultimodalEmbedder_Integration(t *testing.T) {
	apiKey := os.Getenv("VOYAGE_API_KEY")
	if apiKey == "" {
		t.Skip("VOYAGE_API_KEY not set, skipping integration test")
	}

	e := NewVoyageMultimodalEmbedder(apiKey, "voyage-3-lite", "voyage-multimodal-3")
	ctx := context.Background()

	// Test text embedding (using base embedder)
	t.Run("text embedding", func(t *testing.T) {
		embedding, err := e.Embed(ctx, "hello world")
		if err != nil {
			t.Fatalf("Embed failed: %v", err)
		}

		if len(embedding) != 512 { // voyage-3-lite produces 512-dim
			t.Errorf("expected 512 dimensions, got %d", len(embedding))
		}
	})

	// Note: Image embedding tests require actual images and consume API credits
	// They are disabled by default
}
