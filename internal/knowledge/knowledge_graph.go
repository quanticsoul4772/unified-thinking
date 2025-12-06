// Package knowledge provides the unified knowledge graph API combining Neo4j and vector search.
package knowledge

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	chromem "github.com/philippgille/chromem-go"
	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/types"
)

// KnowledgeGraph combines Neo4j graph database with chromem-go vector search
type KnowledgeGraph struct {
	graphStore     *GraphStore
	VectorStore    *VectorStore // Exported for metrics access
	embeddingCache *EmbeddingCache
	neo4jClient    *Neo4jClient
	database       string
	reranker       embeddings.Reranker
}

// KnowledgeGraphConfig holds knowledge graph configuration
type KnowledgeGraphConfig struct {
	Neo4jConfig  Neo4jConfig
	VectorConfig VectorStoreConfig
	SQLiteDB     *sql.DB // For embedding cache
}

// NewKnowledgeGraph creates a new knowledge graph instance
// Knowledge graph is REQUIRED - will FAIL if Neo4j is not available
func NewKnowledgeGraph(cfg KnowledgeGraphConfig) (*KnowledgeGraph, error) {
	// Initialize Neo4j client
	neo4jClient, err := NewNeo4jClient(cfg.Neo4jConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j client: %w", err)
	}

	// Initialize schema
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Neo4jConfig.Timeout)
	defer cancel()

	if err := InitializeSchema(ctx, neo4jClient, cfg.Neo4jConfig.Database); err != nil {
		_ = neo4jClient.Close(ctx)
		return nil, fmt.Errorf("failed to initialize Neo4j schema: %w", err)
	}

	// Initialize graph store
	graphStore := NewGraphStore(neo4jClient, cfg.Neo4jConfig.Database)

	// Initialize vector store
	vectorStore, err := NewVectorStore(cfg.VectorConfig)
	if err != nil {
		_ = neo4jClient.Close(ctx)
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}

	// Initialize embedding cache
	var embeddingCache *EmbeddingCache
	if cfg.SQLiteDB != nil {
		embeddingCache = NewEmbeddingCache(cfg.SQLiteDB)
	}

	kg := &KnowledgeGraph{
		graphStore:     graphStore,
		VectorStore:    vectorStore,
		embeddingCache: embeddingCache,
		neo4jClient:    neo4jClient,
		database:       cfg.Neo4jConfig.Database,
	}

	log.Printf("[DEBUG] Knowledge graph initialized successfully")
	return kg, nil
}

// Close closes all knowledge graph connections
func (kg *KnowledgeGraph) Close(ctx context.Context) error {
	if kg.VectorStore != nil {
		if err := kg.VectorStore.Close(); err != nil {
			return err
		}
	}

	if kg.neo4jClient != nil {
		return kg.neo4jClient.Close(ctx)
	}

	return nil
}

// SetReranker sets the optional reranker for result optimization
func (kg *KnowledgeGraph) SetReranker(reranker embeddings.Reranker) {
	kg.reranker = reranker
}

// StoreEntity stores an entity in both Neo4j and vector search
func (kg *KnowledgeGraph) StoreEntity(ctx context.Context, entity *Entity, content string) error {
	// Store in Neo4j graph
	if err := kg.graphStore.CreateEntity(ctx, entity); err != nil {
		return fmt.Errorf("failed to store entity in graph: %w", err)
	}

	// Generate and cache embedding
	if kg.VectorStore != nil && kg.VectorStore.embedder != nil {
		embedding, err := kg.VectorStore.embedder.Embed(ctx, content)
		if err != nil {
			log.Printf("[WARN] Failed to generate embedding for entity %s: %v", entity.ID, err)
		} else {
			// Store in vector search
			metadata := map[string]string{
				"entity_id":   entity.ID,
				"entity_type": string(entity.Type),
				"label":       entity.Label,
			}

			if err := kg.VectorStore.AddDocument(ctx, "entities", entity.ID, content, metadata); err != nil {
				log.Printf("[WARN] Failed to add entity to vector store: %v", err)
			}

			// Cache embedding in SQLite
			if kg.embeddingCache != nil {
				embCache := &EntityEmbedding{
					EntityID:    entity.ID,
					EntityLabel: entity.Label,
					EntityType:  string(entity.Type),
					Embedding:   embedding,
					Model:       kg.VectorStore.embedder.Model(),
					Provider:    kg.VectorStore.embedder.Provider(),
					Dimension:   len(embedding),
				}

				if err := kg.embeddingCache.Store(embCache); err != nil {
					log.Printf("[WARN] Failed to cache embedding: %v", err)
				}
			}
		}
	}

	return nil
}

// GetEntity retrieves an entity from Neo4j
func (kg *KnowledgeGraph) GetEntity(ctx context.Context, entityID string) (*Entity, error) {
	return kg.graphStore.GetEntity(ctx, entityID)
}

// SearchSemantic performs semantic search using vector similarity
func (kg *KnowledgeGraph) SearchSemantic(ctx context.Context, query string, limit int, minSimilarity float32) ([]chromem.Result, error) {
	if kg.VectorStore == nil {
		return nil, fmt.Errorf("vector store not configured")
	}

	// Get more candidates for reranking if reranker is configured
	searchLimit := limit
	if kg.reranker != nil {
		searchLimit = limit * 2
	}

	results, err := kg.VectorStore.SearchSimilarWithThreshold(ctx, "entities", query, searchLimit, minSimilarity)
	if err != nil {
		return nil, err
	}

	// Apply reranking if configured
	if kg.reranker != nil && len(results) > 0 {
		results, err = kg.rerankSemanticResults(ctx, query, results, limit)
		if err != nil {
			// Log and continue with original results
			log.Printf("[WARN] Reranking failed, using embedding scores: %v", err)
		}
	}

	// Apply limit
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// rerankSemanticResults uses the reranker to optimize semantic search results
func (kg *KnowledgeGraph) rerankSemanticResults(ctx context.Context, query string, results []chromem.Result, limit int) ([]chromem.Result, error) {
	if len(results) == 0 {
		return results, nil
	}

	// Extract document content for reranking
	documents := make([]string, len(results))
	for i, result := range results {
		documents[i] = result.Content
	}

	// Rerank using the reranker
	rerankResults, err := kg.reranker.Rerank(ctx, query, documents, limit)
	if err != nil {
		return results, err
	}

	// Build reranked results
	reranked := make([]chromem.Result, 0, len(rerankResults))
	for _, rr := range rerankResults {
		if rr.Index < len(results) {
			result := results[rr.Index]
			result.Similarity = float32(rr.RelevanceScore) // Update score
			reranked = append(reranked, result)
		}
	}

	return reranked, nil
}

// SearchGraph performs graph traversal to find related entities
func (kg *KnowledgeGraph) SearchGraph(ctx context.Context, entityID string, maxHops int, relationshipTypes []RelationshipType) ([]*Entity, error) {
	return kg.graphStore.QueryEntitiesWithinHops(ctx, entityID, maxHops, relationshipTypes)
}

// HybridSearch combines semantic and graph search
func (kg *KnowledgeGraph) HybridSearch(ctx context.Context, query string, limit int, maxHops int) ([]*Entity, error) {
	return kg.HybridSearchWithThreshold(ctx, query, limit, maxHops, 0.7)
}

// HybridSearchWithThreshold combines semantic and graph search with configurable threshold
func (kg *KnowledgeGraph) HybridSearchWithThreshold(ctx context.Context, query string, limit int, maxHops int, minSimilarity float32) ([]*Entity, error) {
	// Step 1: Semantic search to find relevant starting entities
	// Use lower threshold (0.3) to get more starting points for graph traversal
	// Final filtering by user's minSimilarity happens after graph traversal
	semanticThreshold := float32(0.3)
	if minSimilarity < semanticThreshold {
		semanticThreshold = minSimilarity
	}

	semanticResults, err := kg.SearchSemantic(ctx, query, limit, semanticThreshold)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	if len(semanticResults) == 0 {
		return []*Entity{}, nil
	}

	// Step 2: Graph traversal from top semantic matches
	entityMap := make(map[string]*Entity)

	for _, result := range semanticResults {
		entityID := result.ID

		// Get the entity itself
		entity, err := kg.graphStore.GetEntity(ctx, entityID)
		if err != nil {
			log.Printf("[WARN] HybridSearch: failed to get entity %s: %v", entityID, err)
			continue
		}
		entityMap[entity.ID] = entity

		// Get connected entities
		if maxHops > 0 {
			connected, err := kg.graphStore.QueryEntitiesWithinHops(ctx, entityID, maxHops, nil)
			if err != nil {
				log.Printf("[WARN] HybridSearch: graph traversal failed for %s: %v", entityID, err)
				continue
			}

			for _, e := range connected {
				if _, exists := entityMap[e.ID]; !exists {
					entityMap[e.ID] = e
				}
			}
		}
	}

	// Convert map to slice
	entities := make([]*Entity, 0, len(entityMap))
	for _, entity := range entityMap {
		entities = append(entities, entity)
	}

	return entities, nil
}

// CreateRelationship creates a relationship between entities
func (kg *KnowledgeGraph) CreateRelationship(ctx context.Context, rel *Relationship) error {
	return kg.graphStore.CreateRelationship(ctx, rel)
}

// GetEmbeddingCacheStats returns embedding cache statistics
func (kg *KnowledgeGraph) GetEmbeddingCacheStats() (types.Metadata, error) {
	if kg.embeddingCache == nil {
		return nil, fmt.Errorf("embedding cache not available")
	}
	return kg.embeddingCache.GetCacheStats()
}
