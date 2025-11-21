package embeddings_test

import (
	"context"
	"testing"

	"unified-thinking/internal/embeddings"
)

func TestMockEmbedder_Embed(t *testing.T) {
	mock := embeddings.NewMockEmbedder(512)

	embedding, err := mock.Embed(context.Background(), "test text")
	if err != nil {
		t.Fatalf("Embed() failed: %v", err)
	}

	if len(embedding) != 512 {
		t.Errorf("Expected embedding dimension 512, got %d", len(embedding))
	}

	// Verify it's a unit vector (approximately)
	var sumSquares float64
	for _, val := range embedding {
		sumSquares += float64(val * val)
	}

	// Should be close to 1.0 for unit vector
	if sumSquares < 0.99 || sumSquares > 1.01 {
		t.Errorf("Expected unit vector (sum of squares ≈ 1.0), got %.3f", sumSquares)
	}
}

func TestMockEmbedder_Deterministic(t *testing.T) {
	mock := embeddings.NewMockEmbedder(128)

	// Same text should produce same embedding
	text := "deterministic test"

	emb1, err := mock.Embed(context.Background(), text)
	if err != nil {
		t.Fatalf("First Embed() failed: %v", err)
	}

	emb2, err := mock.Embed(context.Background(), text)
	if err != nil {
		t.Fatalf("Second Embed() failed: %v", err)
	}

	if len(emb1) != len(emb2) {
		t.Fatal("Embeddings have different lengths")
	}

	for i := range emb1 {
		if emb1[i] != emb2[i] {
			t.Errorf("Embeddings differ at index %d: %.6f vs %.6f", i, emb1[i], emb2[i])
			break
		}
	}
}

func TestMockEmbedder_DifferentTexts(t *testing.T) {
	mock := embeddings.NewMockEmbedder(256)

	emb1, _ := mock.Embed(context.Background(), "first text")
	emb2, _ := mock.Embed(context.Background(), "second text")

	// Different texts should produce different embeddings
	identical := true
	for i := range emb1 {
		if emb1[i] != emb2[i] {
			identical = false
			break
		}
	}

	if identical {
		t.Error("Different texts produced identical embeddings")
	}
}

func TestMockEmbedder_EmbedBatch(t *testing.T) {
	mock := embeddings.NewMockEmbedder(128)

	texts := []string{"first", "second", "third"}

	embeddings, err := mock.EmbedBatch(context.Background(), texts)
	if err != nil {
		t.Fatalf("EmbedBatch() failed: %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}

	// Each embedding should match individual Embed() call
	for i, text := range texts {
		individual, _ := mock.Embed(context.Background(), text)

		for j := range individual {
			if embeddings[i][j] != individual[j] {
				t.Errorf("Batch embedding %d differs from individual at index %d", i, j)
				break
			}
		}
	}
}

func TestMockEmbedder_ContextCancellation(t *testing.T) {
	mock := embeddings.NewMockEmbedder(128)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := mock.Embed(ctx, "test")
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestMockEmbedder_Metadata(t *testing.T) {
	mock := embeddings.NewMockEmbedder(512)

	if mock.Dimension() != 512 {
		t.Errorf("Expected dimension 512, got %d", mock.Dimension())
	}

	if mock.Model() != "mock-model" {
		t.Errorf("Expected model 'mock-model', got %s", mock.Model())
	}

	if mock.Provider() != "mock" {
		t.Errorf("Expected provider 'mock', got %s", mock.Provider())
	}
}

func TestFailingMockEmbedder(t *testing.T) {
	mock := embeddings.NewFailingMockEmbedder()

	// Should fail on Embed
	_, err := mock.Embed(context.Background(), "test")
	if err == nil {
		t.Error("Expected error from failing mock embedder")
	}
	if err.Error() != "mock embedder configured to fail" {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Should fail on EmbedBatch
	_, err = mock.EmbedBatch(context.Background(), []string{"test"})
	if err == nil {
		t.Error("Expected error from failing mock embedder batch")
	}
}

func TestMockEmbedder_SetFailOnEmbed(t *testing.T) {
	mock := embeddings.NewMockEmbedder(128)

	// Initially should succeed
	_, err := mock.Embed(context.Background(), "test")
	if err != nil {
		t.Errorf("Initial embed should succeed, got error: %v", err)
	}

	// Configure to fail
	mock.SetFailOnEmbed(true)

	_, err = mock.Embed(context.Background(), "test")
	if err == nil {
		t.Error("Expected error after SetFailOnEmbed(true)")
	}

	// Configure to succeed again
	mock.SetFailOnEmbed(false)

	_, err = mock.Embed(context.Background(), "test")
	if err != nil {
		t.Errorf("Embed should succeed after SetFailOnEmbed(false), got error: %v", err)
	}
}

func TestMockEmbedder_EmptyText(t *testing.T) {
	mock := embeddings.NewMockEmbedder(128)

	// Empty string should still produce valid embedding
	embedding, err := mock.Embed(context.Background(), "")
	if err != nil {
		t.Fatalf("Embed() with empty string failed: %v", err)
	}

	if len(embedding) != 128 {
		t.Errorf("Expected embedding dimension 128, got %d", len(embedding))
	}

	// Should be deterministic for empty string too
	embedding2, _ := mock.Embed(context.Background(), "")

	for i := range embedding {
		if embedding[i] != embedding2[i] {
			t.Error("Empty string embeddings are not deterministic")
			break
		}
	}
}
