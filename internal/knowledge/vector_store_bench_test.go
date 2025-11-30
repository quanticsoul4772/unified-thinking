package knowledge

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
	"unified-thinking/internal/embeddings"
)

// BenchmarkVectorStore_AddDocument benchmarks document addition with embedding generation
func BenchmarkVectorStore_AddDocument(b *testing.B) {
	mockEmbedder := embeddings.NewMockEmbedder(512)
	cfg := VectorStoreConfig{
		Embedder: mockEmbedder,
	}

	vs, err := NewVectorStore(cfg)
	if err != nil {
		b.Fatalf("NewVectorStore failed: %v", err)
	}

	ctx := context.Background()
	if _, err := vs.GetOrCreateCollection(ctx, "bench", nil); err != nil {
		b.Fatalf("GetOrCreateCollection failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := vs.AddDocument(ctx, "bench", fmt.Sprintf("doc-%d", i), "Test content for benchmarking", nil)
		if err != nil {
			b.Fatalf("AddDocument failed: %v", err)
		}
	}
}

// BenchmarkVectorStore_SearchSimilar benchmarks semantic search
func BenchmarkVectorStore_SearchSimilar(b *testing.B) {
	mockEmbedder := embeddings.NewMockEmbedder(512)
	cfg := VectorStoreConfig{
		Embedder: mockEmbedder,
	}

	vs, err := NewVectorStore(cfg)
	if err != nil {
		b.Fatalf("NewVectorStore failed: %v", err)
	}

	ctx := context.Background()

	// Pre-populate with test documents
	for i := 0; i < 1000; i++ {
		content := fmt.Sprintf("Document %d about various topics", i)
		if err := vs.AddDocument(ctx, "bench", fmt.Sprintf("doc-%d", i), content, nil); err != nil {
			b.Fatalf("AddDocument failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vs.SearchSimilar(ctx, "bench", "test query about topics", 10)
		if err != nil {
			b.Fatalf("SearchSimilar failed: %v", err)
		}
	}
}

// BenchmarkVectorStore_SearchSimilar_Large benchmarks search on large dataset
func BenchmarkVectorStore_SearchSimilar_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	mockEmbedder := embeddings.NewMockEmbedder(512)
	cfg := VectorStoreConfig{
		Embedder: mockEmbedder,
	}

	vs, err := NewVectorStore(cfg)
	if err != nil {
		b.Fatalf("NewVectorStore failed: %v", err)
	}

	ctx := context.Background()

	// Pre-populate with 10,000 documents
	for i := 0; i < 10000; i++ {
		content := fmt.Sprintf("Document %d with semantic content about reasoning and knowledge", i)
		if err := vs.AddDocument(ctx, "bench_large", fmt.Sprintf("doc-%d", i), content, nil); err != nil {
			b.Fatalf("AddDocument failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vs.SearchSimilar(ctx, "bench_large", "reasoning and knowledge topics", 10)
		if err != nil {
			b.Fatalf("SearchSimilar failed: %v", err)
		}
	}
}

// BenchmarkEmbeddingCache_Store benchmarks SQLite embedding cache storage
func BenchmarkEmbeddingCache_Store(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Create test embedding
	embedding := make([]float32, 512)
	for i := range embedding {
		embedding[i] = float32(i) / 512.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityEmb := &EntityEmbedding{
			EntityID:    fmt.Sprintf("entity-%d", i),
			EntityLabel: "Test Entity",
			EntityType:  "Concept",
			Embedding:   embedding,
			Model:       "voyage-3-lite",
			Provider:    "voyage",
			Dimension:   512,
		}

		if err := cache.Store(entityEmb); err != nil {
			b.Fatalf("Store failed: %v", err)
		}
	}
}

// BenchmarkEmbeddingCache_Get benchmarks cache retrieval
func BenchmarkEmbeddingCache_Get(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Pre-populate cache
	embedding := make([]float32, 512)
	for i := 0; i < 100; i++ {
		entityEmb := &EntityEmbedding{
			EntityID:    fmt.Sprintf("entity-%d", i),
			EntityLabel: "Test Entity",
			EntityType:  "Concept",
			Embedding:   embedding,
			Model:       "voyage-3-lite",
			Provider:    "voyage",
			Dimension:   512,
		}
		if err := cache.Store(entityEmb); err != nil {
			b.Fatalf("Store failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := fmt.Sprintf("entity-%d", i%100)
		_, err := cache.Get(entityID)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func setupTestDB(b *testing.B) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open test database: %v", err)
	}

	// Create entity_embeddings table
	_, err = db.Exec(`
		CREATE TABLE entity_embeddings (
			entity_id TEXT PRIMARY KEY,
			entity_label TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			embedding BLOB NOT NULL,
			model TEXT NOT NULL,
			provider TEXT NOT NULL,
			dimension INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	return db
}
