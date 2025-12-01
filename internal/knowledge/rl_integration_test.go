package knowledge

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"unified-thinking/internal/embeddings"
)

func TestRLContextRetriever_GetSimilarProblems(t *testing.T) {
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

	// Create RL context retriever with low threshold for MockEmbedder (random embeddings)
	retriever := NewRLContextRetrieverWithThreshold(kg, 0.0)

	// Store test problem entities
	testProblems := []struct {
		id      string
		label   string
		content string
	}{
		{"p1", "Database Performance", "Slow database queries need optimization"},
		{"p2", "Query Optimization", "SQL queries running too slow"},
		{"p3", "Authentication Issue", "User login failing intermittently"},
	}

	for _, tp := range testProblems {
		problem := &Entity{
			ID:    tp.id,
			Label: tp.label,
			Type:  EntityTypeProblem,
		}
		if err := kg.StoreEntity(ctx, problem, tp.content); err != nil {
			t.Fatalf("StoreEntity failed: %v", err)
		}
	}

	// Test GetSimilarProblems
	similarProblems, err := retriever.GetSimilarProblems(ctx, "database performance issues", 3)
	if err != nil {
		t.Fatalf("GetSimilarProblems failed: %v", err)
	}

	if len(similarProblems) < 1 {
		t.Error("Expected at least 1 similar problem")
	}

	// All returned entities should be problems
	for _, entity := range similarProblems {
		if entity.Type != EntityTypeProblem {
			t.Errorf("Entity %s has type %s, want %s", entity.ID, entity.Type, EntityTypeProblem)
		}
	}

	t.Logf("Found %d similar problems", len(similarProblems))

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

func TestRLContextRetriever_GetStrategyPerformance(t *testing.T) {
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

	// Create RL context retriever with low threshold for MockEmbedder (random embeddings)
	retriever := NewRLContextRetrieverWithThreshold(kg, 0.0)

	// Store problems with strategy metadata
	problems := []*Entity{
		{
			ID:    "p1",
			Label: "Database Performance",
			Type:  EntityTypeProblem,
			Metadata: map[string]interface{}{
				"strategy": "indexing",
				"success":  true,
			},
		},
		{
			ID:    "p2",
			Label: "Query Optimization",
			Type:  EntityTypeProblem,
			Metadata: map[string]interface{}{
				"strategy": "indexing",
				"success":  true,
			},
		},
		{
			ID:    "p3",
			Label: "Slow Queries",
			Type:  EntityTypeProblem,
			Metadata: map[string]interface{}{
				"strategy": "caching",
				"success":  false,
			},
		},
	}

	for _, problem := range problems {
		if err := kg.StoreEntity(ctx, problem, problem.Label+" problem"); err != nil {
			t.Fatalf("StoreEntity failed: %v", err)
		}
	}

	// Test GetStrategyPerformance
	performance, err := retriever.GetStrategyPerformance(ctx, "database performance issues")
	if err != nil {
		t.Fatalf("GetStrategyPerformance failed: %v", err)
	}

	if len(performance) == 0 {
		t.Error("Expected strategy performance data")
	}

	// Check indexing strategy performance (2 successes out of 2)
	if indexingPerf, ok := performance["indexing"]; ok {
		if indexingPerf != 1.0 {
			t.Errorf("Indexing performance = %.2f, want 1.0 (2/2 successes)", indexingPerf)
		}
	}

	// Check caching strategy performance (0 successes out of 1)
	if cachingPerf, ok := performance["caching"]; ok {
		if cachingPerf != 0.0 {
			t.Errorf("Caching performance = %.2f, want 0.0 (0/1 successes)", cachingPerf)
		}
	}

	t.Logf("Strategy performance: %v", performance)

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

func TestRLContextRetriever_RecordStrategyOutcome(t *testing.T) {
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

	// Create RL context retriever with low threshold for MockEmbedder (random embeddings)
	retriever := NewRLContextRetrieverWithThreshold(kg, 0.0)

	// Test RecordStrategyOutcome
	problem := "Optimize database queries"
	strategy := "indexing"
	success := true
	confidence := 0.9

	err = retriever.RecordStrategyOutcome(ctx, problem, strategy, success, confidence)
	if err != nil {
		t.Fatalf("RecordStrategyOutcome failed: %v", err)
	}

	// Verify problem entity was created with strategy metadata
	// The ID format is "problem-" + timestamp, so we search semantically instead
	similarProblems, err := retriever.GetSimilarProblems(ctx, problem, 1)
	if err != nil {
		t.Fatalf("GetSimilarProblems failed: %v", err)
	}

	if len(similarProblems) == 0 {
		t.Fatal("Expected to find recorded problem entity")
	}

	problemEntity := similarProblems[0]

	if problemEntity.Type != EntityTypeProblem {
		t.Errorf("Entity type = %s, want %s", problemEntity.Type, EntityTypeProblem)
	}

	if problemEntity.Metadata["strategy"] != strategy {
		t.Errorf("Strategy = %v, want %s", problemEntity.Metadata["strategy"], strategy)
	}

	if problemEntity.Metadata["success"] != success {
		t.Errorf("Success = %v, want %v", problemEntity.Metadata["success"], success)
	}

	// Cleanup
	if err := ClearAllData(ctx, neo4jClient, neo4jCfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}
