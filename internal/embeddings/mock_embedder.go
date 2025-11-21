package embeddings

import (
	"context"
	"fmt"
	"math"
	"math/rand"
)

// MockEmbedder provides a fake embedder for testing without external API dependencies.
// It generates deterministic embeddings based on text hash for reproducible tests.
type MockEmbedder struct {
	dimension int
	model     string
	provider  string
	failOnEmbed bool // Simulate API failures
}

// NewMockEmbedder creates a new mock embedder for testing
func NewMockEmbedder(dimension int) *MockEmbedder {
	return &MockEmbedder{
		dimension: dimension,
		model:     "mock-model",
		provider:  "mock",
	}
}

// NewFailingMockEmbedder creates a mock that always fails (for error path testing)
func NewFailingMockEmbedder() *MockEmbedder {
	return &MockEmbedder{
		dimension:   512,
		model:       "mock-model",
		provider:    "mock",
		failOnEmbed: true,
	}
}

// Embed generates a deterministic embedding based on text content
func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if m.failOnEmbed {
		return nil, fmt.Errorf("mock embedder configured to fail")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Generate deterministic embedding from text hash
	embedding := make([]float32, m.dimension)

	// Use text hash as seed for reproducibility
	seed := int64(0)
	for _, c := range text {
		seed = seed*31 + int64(c)
	}

	rng := rand.New(rand.NewSource(seed))

	// Generate random unit vector (deterministic due to seed)
	var sumSquares float64
	for i := 0; i < m.dimension; i++ {
		embedding[i] = float32(rng.NormFloat64())
		sumSquares += float64(embedding[i] * embedding[i])
	}

	// Normalize to unit vector: divide each component by sqrt(sumSquares)
	if sumSquares > 0 {
		magnitude := float32(math.Sqrt(sumSquares))
		for i := 0; i < m.dimension; i++ {
			embedding[i] /= magnitude
		}
	}

	return embedding, nil
}

// EmbedBatch generates embeddings for multiple texts
func (m *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if m.failOnEmbed {
		return nil, fmt.Errorf("mock embedder configured to fail")
	}

	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// Dimension returns the embedding dimension
func (m *MockEmbedder) Dimension() int {
	return m.dimension
}

// Model returns the model identifier
func (m *MockEmbedder) Model() string {
	return m.model
}

// Provider returns the provider name
func (m *MockEmbedder) Provider() string {
	return m.provider
}

// SetFailOnEmbed configures whether the mock should fail on embedding calls
func (m *MockEmbedder) SetFailOnEmbed(fail bool) {
	m.failOnEmbed = fail
}
