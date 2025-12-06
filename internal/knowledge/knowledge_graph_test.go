package knowledge

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"unified-thinking/internal/embeddings"
)

// TestKnowledgeGraph_StoreAndRetrieve tests entity storage and retrieval
func TestKnowledgeGraph_StoreAndRetrieve(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup Neo4j
	neo4jCfg := DefaultConfig()
	neo4jClient, err := NewNeo4jClient(neo4jCfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer neo4jClient.Close(context.Background())

	// Setup SQLite for cache
	sqliteDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}
	defer sqliteDB.Close()

	// Create entity_embeddings table
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

	// Setup vector store with mock embedder
	mockEmbedder := embeddings.NewMockEmbedder(512)
	vectorCfg := VectorStoreConfig{
		Embedder: mockEmbedder,
	}

	// Create knowledge graph
	kgCfg := KnowledgeGraphConfig{
		Neo4jConfig:  neo4jCfg,
		VectorConfig: vectorCfg,
		SQLiteDB:     sqliteDB,
	}

	kg, err := NewKnowledgeGraph(kgCfg)
	if err != nil {
		t.Fatalf("NewKnowledgeGraph failed: %v", err)
	}
	defer kg.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clear test data
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	// Store entity
	entity := &Entity{
		ID:          "test-entity-1",
		Label:       "Test Concept",
		Type:        EntityTypeConcept,
		Description: "A test concept for knowledge graph",
	}

	content := "This is test content for semantic search about database optimization and query performance"

	err = kg.StoreEntity(ctx, entity, content)
	if err != nil {
		t.Fatalf("StoreEntity failed: %v", err)
	}

	// Retrieve entity from Neo4j
	retrieved, err := kg.GetEntity(ctx, "test-entity-1")
	if err != nil {
		t.Fatalf("GetEntity failed: %v", err)
	}

	if retrieved.Label != entity.Label {
		t.Errorf("Label = %s, want %s", retrieved.Label, entity.Label)
	}

	// Check embedding cache
	cache := NewEmbeddingCache(sqliteDB)
	cached, err := cache.Get("test-entity-1")
	if err != nil {
		t.Fatalf("Cache.Get failed: %v", err)
	}

	if cached == nil {
		t.Error("Expected embedding to be cached")
	} else {
		if cached.Dimension != 512 {
			t.Errorf("Dimension = %d, want 512", cached.Dimension)
		}
		if len(cached.Embedding) != 512 {
			t.Errorf("Embedding length = %d, want 512", len(cached.Embedding))
		}
	}

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

// TestKnowledgeGraph_SemanticSearch tests vector similarity search
func TestKnowledgeGraph_SemanticSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup (similar to previous test)
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

	// Create entity_embeddings table
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

	// Store test entities
	testData := []struct {
		id      string
		label   string
		content string
	}{
		{"e1", "Database Optimization", "Optimize database queries for better performance"},
		{"e2", "Query Performance", "Improve SQL query execution time"},
		{"e3", "Authentication System", "User authentication and authorization"},
	}

	for _, td := range testData {
		entity := &Entity{
			ID:    td.id,
			Label: td.label,
			Type:  EntityTypeConcept,
		}
		if err := kg.StoreEntity(ctx, entity, td.content); err != nil {
			t.Fatalf("StoreEntity failed: %v", err)
		}
	}

	// Semantic search for "database performance"
	// Note: Using threshold -1.0 to accept all results (MockEmbedder generates random embeddings)
	results, err := kg.SearchSemantic(ctx, "database performance optimization", 3, -1.0)
	if err != nil {
		t.Fatalf("SearchSemantic failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("Expected at least 1 semantic search result")
	}

	// First result should be most similar (database-related)
	if len(results) > 0 {
		t.Logf("Top result: %s (similarity: %.4f)", results[0].ID, results[0].Similarity)
	}

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

// TestKnowledgeGraph_HybridSearch tests combined semantic + graph search
func TestKnowledgeGraph_HybridSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	neo4jCfg := DefaultConfig()
	neo4jClient, err := NewNeo4jClient(neo4jCfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer neo4jClient.Close(context.Background())

	// Setup SQLite for cache
	sqliteDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}
	defer sqliteDB.Close()

	// Create entity_embeddings table
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

	// Create entities with relationships
	e1 := &Entity{ID: "optimization", Label: "Optimization", Type: EntityTypeConcept}
	e2 := &Entity{ID: "performance", Label: "Performance", Type: EntityTypeConcept}
	e3 := &Entity{ID: "indexing", Label: "Indexing", Type: EntityTypeConcept}

	if err := kg.StoreEntity(ctx, e1, "Database optimization techniques"); err != nil {
		t.Fatalf("StoreEntity e1 failed: %v", err)
	}
	if err := kg.StoreEntity(ctx, e2, "Performance monitoring and tuning"); err != nil {
		t.Fatalf("StoreEntity e2 failed: %v", err)
	}
	if err := kg.StoreEntity(ctx, e3, "Index creation for faster queries"); err != nil {
		t.Fatalf("StoreEntity e3 failed: %v", err)
	}

	// Create relationships
	rel1 := &Relationship{
		ID:         "r1",
		FromID:     "optimization",
		ToID:       "performance",
		Type:       RelationshipEnables,
		Strength:   0.9,
		Confidence: 0.95,
	}
	rel2 := &Relationship{
		ID:         "r2",
		FromID:     "indexing",
		ToID:       "optimization",
		Type:       RelationshipEnables,
		Strength:   0.85,
		Confidence: 0.9,
	}

	if err := kg.CreateRelationship(ctx, rel1); err != nil {
		t.Fatalf("CreateRelationship r1 failed: %v", err)
	}
	if err := kg.CreateRelationship(ctx, rel2); err != nil {
		t.Fatalf("CreateRelationship r2 failed: %v", err)
	}

	// Hybrid search: semantic + 1 hop graph traversal
	// Note: Using threshold -1.0 to accept all results (MockEmbedder generates random embeddings)
	results, err := kg.HybridSearchWithThreshold(ctx, "database performance", 3, 1, -1.0)
	if err != nil {
		t.Fatalf("HybridSearch failed: %v", err)
	}

	// Should find semantically similar entities + their connected entities
	if len(results) < 2 {
		t.Errorf("Expected at least 2 entities from hybrid search, got %d", len(results))
	}

	t.Logf("Hybrid search found %d entities", len(results))

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}
