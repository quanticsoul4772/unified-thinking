package knowledge

import (
	"context"
	"testing"

	"unified-thinking/internal/embeddings"
)

func TestNewVectorStore_InMemory(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	cfg := VectorStoreConfig{
		PersistPath: "", // In-memory
		Embedder:    embedder,
	}

	vs, err := NewVectorStore(cfg)
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}
	if vs == nil {
		t.Fatal("expected non-nil vector store")
	}
	if vs.db == nil {
		t.Error("expected non-nil db")
	}
}

func TestNewVectorStore_Persistent(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	tempDir := t.TempDir()
	cfg := VectorStoreConfig{
		PersistPath: tempDir,
		Embedder:    embedder,
	}

	vs, err := NewVectorStore(cfg)
	if err != nil {
		t.Fatalf("failed to create persistent vector store: %v", err)
	}
	if vs == nil {
		t.Fatal("expected non-nil vector store")
	}
}

func TestVectorStore_CreateCollection(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()
	err = vs.CreateCollection(ctx, "test-collection", map[string]string{"type": "test"})
	if err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}

	// Verify collection exists
	collection := vs.GetCollection("test-collection")
	if collection == nil {
		t.Error("expected collection to exist after creation")
	}
}

func TestVectorStore_GetCollection_NonExistent(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	collection := vs.GetCollection("non-existent")
	if collection != nil {
		t.Error("expected nil for non-existent collection")
	}
}

func TestVectorStore_GetOrCreateCollection(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// First call creates
	coll1, err := vs.GetOrCreateCollection(ctx, "new-collection", nil)
	if err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}
	if coll1 == nil {
		t.Fatal("expected non-nil collection")
	}

	// Second call gets existing
	coll2, err := vs.GetOrCreateCollection(ctx, "new-collection", nil)
	if err != nil {
		t.Fatalf("failed to get collection: %v", err)
	}
	if coll2 == nil {
		t.Fatal("expected non-nil collection on second call")
	}
}

func TestVectorStore_AddDocument(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	err = vs.AddDocument(ctx, "entities", "doc-1", "This is a test document about AI", map[string]string{"type": "test"})
	if err != nil {
		t.Fatalf("failed to add document: %v", err)
	}

	// Verify document was added by checking collection count
	collection := vs.GetCollection("entities")
	if collection == nil {
		t.Fatal("expected collection to exist")
	}
	if collection.Count() != 1 {
		t.Errorf("expected 1 document, got %d", collection.Count())
	}
}

func TestVectorStore_AddDocument_MultipleDocuments(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	docs := []struct {
		id      string
		content string
	}{
		{"doc-1", "Machine learning is a subset of AI"},
		{"doc-2", "Deep learning uses neural networks"},
		{"doc-3", "Natural language processing handles text"},
	}

	for _, doc := range docs {
		err = vs.AddDocument(ctx, "ml-concepts", doc.id, doc.content, nil)
		if err != nil {
			t.Fatalf("failed to add document %s: %v", doc.id, err)
		}
	}

	collection := vs.GetCollection("ml-concepts")
	if collection.Count() != 3 {
		t.Errorf("expected 3 documents, got %d", collection.Count())
	}
}

func TestVectorStore_SearchSimilar(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// Add documents
	docs := []struct {
		id      string
		content string
	}{
		{"ml", "Machine learning is great for predictions"},
		{"dl", "Deep learning uses many layers"},
		{"nlp", "Natural language processing understands text"},
	}

	for _, doc := range docs {
		err = vs.AddDocument(ctx, "concepts", doc.id, doc.content, nil)
		if err != nil {
			t.Fatalf("failed to add document: %v", err)
		}
	}

	// Search
	results, err := vs.SearchSimilar(ctx, "concepts", "machine learning AI", 2)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Results should have content and similarity scores
	for _, result := range results {
		if result.Content == "" {
			t.Error("expected non-empty content")
		}
		if result.Similarity <= 0 {
			t.Error("expected positive similarity score")
		}
	}
}

func TestVectorStore_SearchSimilar_EmptyCollection(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// Create empty collection
	err = vs.CreateCollection(ctx, "empty", nil)
	if err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}

	// Search should return empty results
	results, err := vs.SearchSimilar(ctx, "empty", "test query", 10)
	if err != nil {
		t.Fatalf("unexpected error searching empty collection: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestVectorStore_SearchSimilar_NonExistentCollection(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	_, err = vs.SearchSimilar(ctx, "does-not-exist", "test", 10)
	if err == nil {
		t.Error("expected error for non-existent collection")
	}
}

func TestVectorStore_SearchSimilar_LimitExceedsCount(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// Add 2 documents
	_ = vs.AddDocument(ctx, "small", "doc1", "Document one", nil)
	_ = vs.AddDocument(ctx, "small", "doc2", "Document two", nil)

	// Request 100 but only 2 exist - should return 2
	results, err := vs.SearchSimilar(ctx, "small", "document", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results (capped to collection size), got %d", len(results))
	}
}

func TestVectorStore_SearchSimilarWithThreshold(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// Add documents
	_ = vs.AddDocument(ctx, "threshold-test", "ai1", "Artificial intelligence machine learning", nil)
	_ = vs.AddDocument(ctx, "threshold-test", "ai2", "Deep neural networks AI", nil)
	_ = vs.AddDocument(ctx, "threshold-test", "other", "Cooking recipes and food", nil)

	// Search with high threshold - should filter low similarity results
	results, err := vs.SearchSimilarWithThreshold(ctx, "threshold-test", "AI machine learning", 10, 0.5)
	if err != nil {
		t.Fatalf("failed to search with threshold: %v", err)
	}

	// All results should have similarity >= threshold
	for _, result := range results {
		if result.Similarity < 0.5 {
			t.Errorf("result has similarity %f below threshold 0.5", result.Similarity)
		}
	}
}

func TestVectorStore_ListCollections(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// Initially empty
	names, err := vs.ListCollections()
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 collections initially, got %d", len(names))
	}

	// Create collections
	_ = vs.CreateCollection(ctx, "coll-a", nil)
	_ = vs.CreateCollection(ctx, "coll-b", nil)
	_ = vs.CreateCollection(ctx, "coll-c", nil)

	names, err = vs.ListCollections()
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("expected 3 collections, got %d", len(names))
	}
}

func TestVectorStore_DeleteCollection(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// Create and verify
	_ = vs.CreateCollection(ctx, "to-delete", nil)
	if vs.GetCollection("to-delete") == nil {
		t.Fatal("expected collection to exist")
	}

	// Delete
	err = vs.DeleteCollection("to-delete")
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Verify deleted
	if vs.GetCollection("to-delete") != nil {
		t.Error("expected collection to be deleted")
	}
}

func TestVectorStore_Close(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	// Close should not error
	err = vs.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestVectorStore_SearchSimilar_DefaultLimit(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// Add 15 documents
	for i := 0; i < 15; i++ {
		_ = vs.AddDocument(ctx, "limit-test", "doc-"+string(rune('a'+i)), "Document content", nil)
	}

	// Search with limit 0 should use default of 10
	results, err := vs.SearchSimilar(ctx, "limit-test", "document", 0)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}
	if len(results) != 10 {
		t.Errorf("expected 10 results (default limit), got %d", len(results))
	}

	// Search with negative limit should use default of 10
	results, err = vs.SearchSimilar(ctx, "limit-test", "document", -5)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}
	if len(results) != 10 {
		t.Errorf("expected 10 results (default limit), got %d", len(results))
	}
}

func TestVectorStore_AddDocument_FailingEmbedder(t *testing.T) {
	embedder := embeddings.NewFailingMockEmbedder()
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	err = vs.AddDocument(ctx, "test", "doc-1", "This should fail", nil)
	if err == nil {
		t.Error("expected error when embedder fails")
	}
}

func TestVectorStore_SearchSimilar_FailingEmbedder(t *testing.T) {
	embedder := embeddings.NewMockEmbedder(128)
	vs, err := NewVectorStore(VectorStoreConfig{Embedder: embedder})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	ctx := context.Background()

	// Add a document first with working embedder
	err = vs.AddDocument(ctx, "test", "doc-1", "Test document", nil)
	if err != nil {
		t.Fatalf("failed to add document: %v", err)
	}

	// Switch to failing embedder
	failingEmbedder := embeddings.NewFailingMockEmbedder()
	vs.embedder = failingEmbedder

	// Search should fail
	_, err = vs.SearchSimilar(ctx, "test", "query", 10)
	if err == nil {
		t.Error("expected error when embedder fails during search")
	}
}
