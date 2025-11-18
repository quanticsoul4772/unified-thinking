package contextbridge

import (
	"context"
	"log"

	"unified-thinking/internal/embeddings"
)

// EmbeddingSimilarity calculates similarity using semantic embeddings
type EmbeddingSimilarity struct {
	embedder   embeddings.Embedder
	fallback   SimilarityCalculator
	hybridMode bool // Combine embedding + concept similarity
	embedWeight float64 // Weight for embedding similarity in hybrid mode
}

// NewEmbeddingSimilarity creates a new embedding-based similarity calculator
func NewEmbeddingSimilarity(embedder embeddings.Embedder, fallback SimilarityCalculator, hybridMode bool) *EmbeddingSimilarity {
	return &EmbeddingSimilarity{
		embedder:    embedder,
		fallback:    fallback,
		hybridMode:  hybridMode,
		embedWeight: 0.7, // 70% embedding, 30% concept similarity
	}
}

// Calculate computes similarity between two signatures
// If embeddings are available, uses cosine similarity on embeddings
// Falls back to concept-based similarity if embeddings are missing
func (es *EmbeddingSimilarity) Calculate(sig1, sig2 *Signature) float64 {
	if sig1 == nil || sig2 == nil {
		return 0.0
	}

	// If both have embeddings, use embedding similarity
	if len(sig1.Embedding) > 0 && len(sig2.Embedding) > 0 {
		embeddingSim := embeddings.CosineSimilarity(sig1.Embedding, sig2.Embedding)

		// In hybrid mode, combine with concept similarity
		if es.hybridMode && es.fallback != nil {
			conceptSim := es.fallback.Calculate(sig1, sig2)
			return (embeddingSim * es.embedWeight) + (conceptSim * (1.0 - es.embedWeight))
		}

		return embeddingSim
	}

	// Fall back to concept-based similarity
	if es.fallback != nil {
		return es.fallback.Calculate(sig1, sig2)
	}

	return 0.0
}

// GenerateEmbedding generates an embedding for a signature's content
func (es *EmbeddingSimilarity) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if es.embedder == nil {
		return nil, nil
	}

	embedding, err := es.embedder.Embed(ctx, text)
	if err != nil {
		log.Printf("[WARN] Failed to generate embedding: %v", err)
		return nil, err
	}

	return embedding, nil
}

// HasEmbedder returns true if an embedder is configured
func (es *EmbeddingSimilarity) HasEmbedder() bool {
	return es.embedder != nil
}

// GetEmbedder returns the underlying embedder
func (es *EmbeddingSimilarity) GetEmbedder() embeddings.Embedder {
	return es.embedder
}
