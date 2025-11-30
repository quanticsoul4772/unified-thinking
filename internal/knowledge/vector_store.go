// Package knowledge provides vector search capabilities using chromem-go.
package knowledge

import (
	"context"
	"fmt"
	"log"

	chromem "github.com/philippgille/chromem-go"
	"unified-thinking/internal/embeddings"
)

// VectorStore provides semantic similarity search using chromem-go
type VectorStore struct {
	db       *chromem.DB
	embedder embeddings.Embedder
}

// VectorStoreConfig holds vector store configuration
type VectorStoreConfig struct {
	PersistPath string // Path to persist vector database (empty = in-memory only)
	Embedder    embeddings.Embedder
}

// NewVectorStore creates a new vector store with chromem-go
func NewVectorStore(cfg VectorStoreConfig) (*VectorStore, error) {
	var db *chromem.DB
	var err error

	if cfg.PersistPath != "" {
		// Persistent store
		db, err = chromem.NewPersistentDB(cfg.PersistPath, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistent vector DB: %w", err)
		}
		log.Printf("[DEBUG] Vector store initialized with persistence at %s", cfg.PersistPath)
	} else {
		// In-memory only
		db = chromem.NewDB()
		log.Printf("[DEBUG] Vector store initialized (in-memory only)")
	}

	return &VectorStore{
		db:       db,
		embedder: cfg.Embedder,
	}, nil
}

// CreateCollection creates a new collection for storing entity embeddings
func (vs *VectorStore) CreateCollection(ctx context.Context, name string, metadata map[string]string) error {
	// chromem-go uses sync.Map internally, already thread-safe
	_, err := vs.db.CreateCollection(name, metadata, nil)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	return nil
}

// GetCollection retrieves an existing collection
func (vs *VectorStore) GetCollection(name string) *chromem.Collection {
	return vs.db.GetCollection(name, nil)
}

// GetOrCreateCollection gets existing collection or creates new one
func (vs *VectorStore) GetOrCreateCollection(ctx context.Context, name string, metadata map[string]string) (*chromem.Collection, error) {
	collection := vs.db.GetCollection(name, nil)
	if collection != nil {
		return collection, nil
	}
	// Collection doesn't exist, create it
	return vs.db.CreateCollection(name, metadata, nil)
}

// AddDocument adds a document to a collection with embedding
func (vs *VectorStore) AddDocument(ctx context.Context, collectionName string, id string, content string, metadata map[string]string) error {
	collection, err := vs.GetOrCreateCollection(ctx, collectionName, nil)
	if err != nil {
		return err
	}

	// Generate embedding using configured embedder
	embedding, err := vs.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Add document with embedding
	err = collection.AddDocument(ctx, chromem.Document{
		ID:        id,
		Content:   content,
		Metadata:  metadata,
		Embedding: embedding,
	})
	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	return nil
}

// SearchSimilar performs semantic similarity search
func (vs *VectorStore) SearchSimilar(ctx context.Context, collectionName string, query string, limit int) ([]chromem.Result, error) {
	if limit <= 0 {
		limit = 10
	}

	collection := vs.db.GetCollection(collectionName, nil)
	if collection == nil {
		return nil, fmt.Errorf("collection not found: %s", collectionName)
	}

	// Generate query embedding
	queryEmbedding, err := vs.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search with cosine similarity
	results, err := collection.QueryEmbedding(ctx, queryEmbedding, limit, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	return results, nil
}

// SearchSimilarWithThreshold performs semantic search with minimum similarity threshold
func (vs *VectorStore) SearchSimilarWithThreshold(ctx context.Context, collectionName string, query string, limit int, minSimilarity float32) ([]chromem.Result, error) {
	allResults, err := vs.SearchSimilar(ctx, collectionName, query, limit*2)
	if err != nil {
		return nil, err
	}

	// Filter by similarity threshold
	filtered := make([]chromem.Result, 0)
	for _, result := range allResults {
		if result.Similarity >= minSimilarity {
			filtered = append(filtered, result)
			if len(filtered) >= limit {
				break
			}
		}
	}

	return filtered, nil
}

// ListCollections returns all collection names
func (vs *VectorStore) ListCollections() ([]string, error) {
	collections := vs.db.ListCollections()
	names := make([]string, 0, len(collections))
	for name := range collections {
		names = append(names, name)
	}
	return names, nil
}

// DeleteCollection removes a collection
func (vs *VectorStore) DeleteCollection(name string) error {
	vs.db.DeleteCollection(name)
	return nil
}

// Close persists the vector database if configured
func (vs *VectorStore) Close() error {
	// chromem-go auto-persists on operations when configured with PersistPath
	// No explicit close needed
	return nil
}
