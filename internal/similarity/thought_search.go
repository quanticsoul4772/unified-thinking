// Package similarity provides thought similarity search using semantic embeddings.
package similarity

import (
	"context"
	"fmt"
	"sort"

	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// ThoughtSearcher provides semantic search over stored thoughts
type ThoughtSearcher struct {
	storage  storage.Storage
	embedder embeddings.Embedder
	reranker embeddings.Reranker
}

// NewThoughtSearcher creates a new thought searcher
func NewThoughtSearcher(store storage.Storage, embedder embeddings.Embedder) *ThoughtSearcher {
	return &ThoughtSearcher{
		storage:  store,
		embedder: embedder,
	}
}

// SetReranker sets the optional reranker for result optimization
func (ts *ThoughtSearcher) SetReranker(reranker embeddings.Reranker) {
	ts.reranker = reranker
}

// SimilarThought represents a thought with similarity score
type SimilarThought struct {
	Thought    *types.Thought
	Similarity float32
}

// SearchSimilar finds thoughts similar to query content
func (ts *ThoughtSearcher) SearchSimilar(ctx context.Context, query string, limit int, minSimilarity float32) ([]*SimilarThought, error) {
	if ts.embedder == nil {
		return nil, fmt.Errorf("embedder not configured")
	}

	// Generate query embedding
	queryEmbedding, err := ts.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Get all thoughts from storage (empty query, any mode)
	allThoughts := ts.storage.SearchThoughts("", "", 10000, 0)

	// Calculate similarity for each thought
	results := make([]*SimilarThought, 0, len(allThoughts))

	for _, thought := range allThoughts {
		// Generate embedding for thought content
		thoughtEmbedding, err := ts.embedder.Embed(ctx, thought.Content)
		if err != nil {
			continue // Skip thoughts we can't embed
		}

		// Calculate cosine similarity
		similarity := cosineSimilarity(queryEmbedding, thoughtEmbedding)

		if similarity >= minSimilarity {
			results = append(results, &SimilarThought{
				Thought:    thought,
				Similarity: similarity,
			})
		}
	}

	// Sort by similarity descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// If reranker is configured, use it to optimize results
	if ts.reranker != nil && len(results) > 0 {
		results, err = ts.rerankResults(ctx, query, results, limit)
		if err != nil {
			return nil, fmt.Errorf("reranking failed: %w - reranker errors must not be silently ignored", err)
		}
	}

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// rerankResults uses the reranker to optimize search results
func (ts *ThoughtSearcher) rerankResults(ctx context.Context, query string, results []*SimilarThought, limit int) ([]*SimilarThought, error) {
	if len(results) == 0 {
		return results, nil
	}

	// Get more candidates for reranking (2x limit or all if less)
	candidateCount := limit * 2
	if candidateCount > len(results) {
		candidateCount = len(results)
	}
	candidates := results[:candidateCount]

	// Extract document content for reranking
	documents := make([]string, len(candidates))
	for i, st := range candidates {
		documents[i] = st.Thought.Content
	}

	// Rerank using the reranker
	rerankResults, err := ts.reranker.Rerank(ctx, query, documents, limit)
	if err != nil {
		return results, err
	}

	// Build reranked results
	reranked := make([]*SimilarThought, 0, len(rerankResults))
	for _, rr := range rerankResults {
		if rr.Index < len(candidates) {
			reranked = append(reranked, &SimilarThought{
				Thought:    candidates[rr.Index].Thought,
				Similarity: float32(rr.RelevanceScore), // Use rerank score
			})
		}
	}

	return reranked, nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	// Calculate square roots using math package
	normASqrt := float32(1.0)
	for i := 0; i < 10; i++ {
		normASqrt = (normASqrt + normA/normASqrt) / 2.0
	}

	normBSqrt := float32(1.0)
	for i := 0; i < 10; i++ {
		normBSqrt = (normBSqrt + normB/normBSqrt) / 2.0
	}

	return dotProduct / (normASqrt * normBSqrt)
}
