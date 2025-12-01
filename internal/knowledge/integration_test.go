package knowledge

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/knowledge/extraction"
)

// TestFullIntegration tests complete workflow: extraction -> storage -> search -> retrieval
func TestFullIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full integration test in short mode")
	}

	// Setup
	neo4jCfg := DefaultConfig()
	neo4jClient, err := NewNeo4jClient(neo4jCfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer neo4jClient.Close(context.Background())

	sqliteDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}
	defer sqliteDB.Close()

	// Create schema
	_, err = sqliteDB.Exec(`
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
		t.Fatalf("Failed to create table: %v", err)
	}

	mockEmbedder := embeddings.NewMockEmbedder(512)
	vectorCfg := VectorStoreConfig{
		Embedder: mockEmbedder,
	}

	kgCfg := KnowledgeGraphConfig{
		Neo4jConfig:  neo4jCfg,
		VectorConfig: vectorCfg,
		SQLiteDB:     sqliteDB,
		Enabled:      true,
	}

	kg, err := NewKnowledgeGraph(kgCfg)
	if err != nil {
		t.Fatalf("NewKnowledgeGraph failed: %v", err)
	}
	defer kg.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	// Test workflow
	testContent := `
The database optimization at https://db.example.com causes better performance.
This enables faster query execution which leads to improved user experience.
Contact admin@example.com for details about the config.json file.
The deployment on 2025-01-15 improved response time to 150ms.
`

	// Step 1: Extract entities and relationships
	extractor := extraction.NewHybridExtractor(extraction.HybridConfig{EnableLLM: false})
	extractResult, err := extractor.Extract(testContent)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	t.Logf("Extracted %d entities, %d relationships", extractResult.TotalEntities, extractResult.TotalRelationships)

	// Step 2: Store extracted entities in knowledge graph
	for _, extractedEntity := range extractResult.Entities {
		entity := &Entity{
			ID:    extractedEntity.Type + "-" + extractedEntity.Text,
			Label: extractedEntity.Text,
			Type:  mapExtractedTypeForTest(extractedEntity.Type),
			Metadata: map[string]interface{}{
				"confidence": extractedEntity.Confidence,
				"method":     extractedEntity.Method,
			},
		}

		if err := kg.StoreEntity(ctx, entity, extractedEntity.Text); err != nil {
			t.Logf("Failed to store entity %s: %v", entity.ID, err)
		}
	}

	// Step 3: Semantic search for similar entities
	results, err := kg.SearchSemantic(ctx, "database performance optimization", 5, 0.5)
	if err != nil {
		t.Fatalf("Semantic search failed: %v", err)
	}

	t.Logf("Semantic search found %d results", len(results))

	if len(results) > 0 {
		t.Logf("Top result: %s (similarity: %.4f)", results[0].ID, results[0].Similarity)
	}

	// Step 4: Verify embedding cache
	cache := NewEmbeddingCache(sqliteDB)
	stats, err := cache.GetCacheStats()
	if err != nil {
		t.Fatalf("GetCacheStats failed: %v", err)
	}

	t.Logf("Embedding cache stats: %+v", stats)

	if totalCached, ok := stats["total_cached"].(int); ok && totalCached < 1 {
		t.Error("Expected at least 1 cached embedding")
	}

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

func mapExtractedTypeForTest(extractedType string) EntityType {
	switch extractedType {
	case "url", "file_path":
		return EntityTypeTool
	case "email":
		return EntityTypePerson
	default:
		return EntityTypeConcept
	}
}
