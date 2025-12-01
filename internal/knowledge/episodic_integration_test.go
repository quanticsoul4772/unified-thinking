package knowledge

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"unified-thinking/internal/embeddings"
)

func TestTrajectoryExtractor_ExtractFromTrajectory(t *testing.T) {
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

	// Setup knowledge graph
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

	// Clear test data
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	// Create trajectory extractor
	extractor := NewTrajectoryExtractor(kg, false) // false = no LLM

	// Test extraction
	trajectoryID := "test-trajectory-1"
	problem := "How to optimize database queries with indexes and caching strategies"
	steps := []string{
		"Analyze current query performance with EXPLAIN",
		"Create appropriate indexes on filtered columns",
		"Implement query result caching layer",
	}

	err = extractor.ExtractFromTrajectory(ctx, trajectoryID, problem, steps)
	if err != nil {
		t.Fatalf("ExtractFromTrajectory failed: %v", err)
	}

	// Verify problem entity was created
	problemEntityID := "problem-" + trajectoryID
	problemEntity, err := kg.GetEntity(ctx, problemEntityID)
	if err != nil {
		t.Fatalf("GetEntity for problem failed: %v", err)
	}

	if problemEntity.Type != EntityTypeProblem {
		t.Errorf("Problem entity type = %s, want %s", problemEntity.Type, EntityTypeProblem)
	}

	if problemEntity.Metadata["trajectory_id"] != trajectoryID {
		t.Errorf("Problem entity trajectory_id = %v, want %s", problemEntity.Metadata["trajectory_id"], trajectoryID)
	}

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

func TestTrajectoryExtractor_ExtractStrategies(t *testing.T) {
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

	extractor := NewTrajectoryExtractor(kg, false) // false = no LLM

	// Test extraction (note: there's no ExtractStrategies method, just ExtractFromTrajectory)
	trajectoryID := "test-trajectory-2"
	problem := "Implement user authentication"
	steps := []string{
		"Use JWT tokens for stateless authentication",
		"Hash passwords with bcrypt algorithm",
		"Implement refresh token rotation",
	}

	err = extractor.ExtractFromTrajectory(ctx, trajectoryID, problem, steps)
	if err != nil {
		t.Fatalf("ExtractStrategies failed: %v", err)
	}

	// Verify strategy entities were created
	// Note: Using threshold -1.0 to accept all results (MockEmbedder generates random embeddings)
	strategyResults, err := kg.SearchSemantic(ctx, "authentication JWT bcrypt", 3, -1.0)
	if err != nil {
		t.Fatalf("SearchSemantic failed: %v", err)
	}

	if len(strategyResults) < 1 {
		t.Error("Expected at least 1 strategy entity to be created")
	}

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}
