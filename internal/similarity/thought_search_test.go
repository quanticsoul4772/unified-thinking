package similarity

import (
	"context"
	"testing"

	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestThoughtSearcher_SearchSimilar(t *testing.T) {
	store := storage.NewMemoryStorage()
	embedder := embeddings.NewMockEmbedder(512)
	searcher := NewThoughtSearcher(store, embedder)

	ctx := context.Background()

	// Store test thoughts
	thoughts := []*types.Thought{
		{ID: "t1", Content: "How to optimize database queries with indexes", Mode: types.ModeLinear, Confidence: 0.9},
		{ID: "t2", Content: "Best practices for SQL query performance tuning", Mode: types.ModeLinear, Confidence: 0.8},
		{ID: "t3", Content: "User authentication with JWT tokens", Mode: types.ModeLinear, Confidence: 0.85},
	}

	for _, th := range thoughts {
		_ = store.StoreThought(th)
	}

	// Search for database-related thoughts
	results, err := searcher.SearchSimilar(ctx, "database performance optimization", 5, 0.0)
	if err != nil {
		t.Fatalf("SearchSimilar failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("Expected at least 1 result")
	}

	// Results should be sorted by similarity
	if len(results) >= 2 {
		if results[0].Similarity < results[1].Similarity {
			t.Error("Results not sorted by similarity (descending)")
		}
	}

	t.Logf("Found %d similar thoughts", len(results))
	for i, r := range results {
		t.Logf("  %d. %s (similarity: %.4f)", i+1, r.Thought.Content, r.Similarity)
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float32
	}{
		{
			name:     "identical vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{1, 0, 0},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        []float32{1, 0},
			b:        []float32{0, 1},
			expected: 0.0,
		},
		{
			name:     "opposite vectors",
			a:        []float32{1, 0},
			b:        []float32{-1, 0},
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			// Allow small floating point error
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Errorf("cosineSimilarity() = %v, want %v", result, tt.expected)
			}
		})
	}
}
