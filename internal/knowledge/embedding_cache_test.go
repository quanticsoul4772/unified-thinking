package knowledge

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// setupEmbeddingCacheTestDB creates an in-memory SQLite database with the required schema
func setupEmbeddingCacheTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create the entity_embeddings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS entity_embeddings (
			entity_id TEXT PRIMARY KEY,
			entity_label TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			embedding TEXT NOT NULL,
			model TEXT NOT NULL,
			provider TEXT NOT NULL,
			dimension INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestNewEmbeddingCache(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache.db != db {
		t.Error("expected cache to have the same db instance")
	}
}

func TestEmbeddingCache_Store(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	embedding := &EntityEmbedding{
		EntityID:    "test-entity-1",
		EntityLabel: "Test Entity",
		EntityType:  "Concept",
		Embedding:   []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		Model:       "voyage-3-lite",
		Provider:    "voyage",
		Dimension:   5,
	}

	err := cache.Store(embedding)
	if err != nil {
		t.Fatalf("failed to store embedding: %v", err)
	}

	// Verify timestamps were set
	if embedding.CreatedAt == 0 {
		t.Error("expected CreatedAt to be set")
	}
	if embedding.UpdatedAt == 0 {
		t.Error("expected UpdatedAt to be set")
	}

	// Verify by retrieving
	retrieved, err := cache.Get("test-entity-1")
	if err != nil {
		t.Fatalf("failed to get embedding: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected non-nil embedding")
	}
	if retrieved.EntityLabel != "Test Entity" {
		t.Errorf("expected label 'Test Entity', got '%s'", retrieved.EntityLabel)
	}
}

func TestEmbeddingCache_Store_Update(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Store initial embedding
	embedding1 := &EntityEmbedding{
		EntityID:    "test-entity-1",
		EntityLabel: "Initial Label",
		EntityType:  "Concept",
		Embedding:   []float32{0.1, 0.2, 0.3},
		Model:       "voyage-3-lite",
		Provider:    "voyage",
		Dimension:   3,
	}
	err := cache.Store(embedding1)
	if err != nil {
		t.Fatalf("failed to store initial embedding: %v", err)
	}

	initialCreatedAt := embedding1.CreatedAt

	// Update with new label
	embedding2 := &EntityEmbedding{
		EntityID:    "test-entity-1",
		EntityLabel: "Updated Label",
		EntityType:  "Concept",
		Embedding:   []float32{0.4, 0.5, 0.6},
		Model:       "voyage-3-lite",
		Provider:    "voyage",
		Dimension:   3,
		CreatedAt:   initialCreatedAt, // Preserve original created_at
	}
	err = cache.Store(embedding2)
	if err != nil {
		t.Fatalf("failed to update embedding: %v", err)
	}

	// Verify update
	retrieved, err := cache.Get("test-entity-1")
	if err != nil {
		t.Fatalf("failed to get embedding: %v", err)
	}
	if retrieved.EntityLabel != "Updated Label" {
		t.Errorf("expected label 'Updated Label', got '%s'", retrieved.EntityLabel)
	}
	if len(retrieved.Embedding) != 3 || retrieved.Embedding[0] != 0.4 {
		t.Error("embedding was not updated correctly")
	}
}

func TestEmbeddingCache_Get_NotFound(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	retrieved, err := cache.Get("non-existent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved != nil {
		t.Error("expected nil for non-existent entity")
	}
}

func TestEmbeddingCache_GetByType(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Store multiple embeddings of different types
	embeddings := []*EntityEmbedding{
		{
			EntityID:    "concept-1",
			EntityLabel: "Concept 1",
			EntityType:  "Concept",
			Embedding:   []float32{0.1, 0.2},
			Model:       "test",
			Provider:    "test",
			Dimension:   2,
		},
		{
			EntityID:    "concept-2",
			EntityLabel: "Concept 2",
			EntityType:  "Concept",
			Embedding:   []float32{0.3, 0.4},
			Model:       "test",
			Provider:    "test",
			Dimension:   2,
		},
		{
			EntityID:    "person-1",
			EntityLabel: "Person 1",
			EntityType:  "Person",
			Embedding:   []float32{0.5, 0.6},
			Model:       "test",
			Provider:    "test",
			Dimension:   2,
		},
	}

	for _, e := range embeddings {
		if err := cache.Store(e); err != nil {
			t.Fatalf("failed to store embedding: %v", err)
		}
	}

	// Get by type "Concept"
	concepts, err := cache.GetByType("Concept", 10)
	if err != nil {
		t.Fatalf("failed to get by type: %v", err)
	}
	if len(concepts) != 2 {
		t.Errorf("expected 2 concepts, got %d", len(concepts))
	}

	// Get by type "Person"
	persons, err := cache.GetByType("Person", 10)
	if err != nil {
		t.Fatalf("failed to get by type: %v", err)
	}
	if len(persons) != 1 {
		t.Errorf("expected 1 person, got %d", len(persons))
	}

	// Get non-existent type
	tools, err := cache.GetByType("Tool", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}
}

func TestEmbeddingCache_GetByType_Limit(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Store 5 embeddings
	for i := 0; i < 5; i++ {
		e := &EntityEmbedding{
			EntityID:    "concept-" + string(rune('a'+i)),
			EntityLabel: "Concept",
			EntityType:  "Concept",
			Embedding:   []float32{float32(i)},
			Model:       "test",
			Provider:    "test",
			Dimension:   1,
		}
		if err := cache.Store(e); err != nil {
			t.Fatalf("failed to store embedding: %v", err)
		}
	}

	// Get with limit 3
	results, err := cache.GetByType("Concept", 3)
	if err != nil {
		t.Fatalf("failed to get by type: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Get with limit 0 (should default to 100)
	results, err = cache.GetByType("Concept", 0)
	if err != nil {
		t.Fatalf("failed to get by type: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("expected 5 results with default limit, got %d", len(results))
	}
}

func TestEmbeddingCache_Delete(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Store an embedding
	embedding := &EntityEmbedding{
		EntityID:    "to-delete",
		EntityLabel: "Delete Me",
		EntityType:  "Concept",
		Embedding:   []float32{0.1},
		Model:       "test",
		Provider:    "test",
		Dimension:   1,
	}
	if err := cache.Store(embedding); err != nil {
		t.Fatalf("failed to store embedding: %v", err)
	}

	// Verify it exists
	retrieved, _ := cache.Get("to-delete")
	if retrieved == nil {
		t.Fatal("expected embedding to exist before delete")
	}

	// Delete
	if err := cache.Delete("to-delete"); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Verify it's gone
	retrieved, _ = cache.Get("to-delete")
	if retrieved != nil {
		t.Error("expected embedding to be deleted")
	}
}

func TestEmbeddingCache_Delete_NonExistent(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Delete non-existent should not error
	err := cache.Delete("non-existent")
	if err != nil {
		t.Errorf("unexpected error deleting non-existent: %v", err)
	}
}

func TestEmbeddingCache_Count(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Initially empty
	count, err := cache.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// Add some embeddings
	for i := 0; i < 3; i++ {
		e := &EntityEmbedding{
			EntityID:    "entity-" + string(rune('a'+i)),
			EntityLabel: "Entity",
			EntityType:  "Concept",
			Embedding:   []float32{float32(i)},
			Model:       "test",
			Provider:    "test",
			Dimension:   1,
		}
		if err := cache.Store(e); err != nil {
			t.Fatalf("failed to store: %v", err)
		}
	}

	count, err = cache.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestEmbeddingCache_GetCacheStats(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Store embeddings of different types
	embeddings := []*EntityEmbedding{
		{EntityID: "c1", EntityLabel: "C1", EntityType: "Concept", Embedding: []float32{0.1}, Model: "test", Provider: "test", Dimension: 1},
		{EntityID: "c2", EntityLabel: "C2", EntityType: "Concept", Embedding: []float32{0.2}, Model: "test", Provider: "test", Dimension: 1},
		{EntityID: "p1", EntityLabel: "P1", EntityType: "Person", Embedding: []float32{0.3}, Model: "test", Provider: "test", Dimension: 1},
		{EntityID: "t1", EntityLabel: "T1", EntityType: "Tool", Embedding: []float32{0.4}, Model: "test", Provider: "test", Dimension: 1},
	}

	for _, e := range embeddings {
		if err := cache.Store(e); err != nil {
			t.Fatalf("failed to store: %v", err)
		}
	}

	stats, err := cache.GetCacheStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	// Check total
	total, ok := stats["total_cached"].(int)
	if !ok || total != 4 {
		t.Errorf("expected total_cached=4, got %v", stats["total_cached"])
	}

	// Check by type
	byType, ok := stats["by_type"].(map[string]int)
	if !ok {
		t.Fatal("expected by_type to be map[string]int")
	}
	if byType["Concept"] != 2 {
		t.Errorf("expected 2 Concepts, got %d", byType["Concept"])
	}
	if byType["Person"] != 1 {
		t.Errorf("expected 1 Person, got %d", byType["Person"])
	}
	if byType["Tool"] != 1 {
		t.Errorf("expected 1 Tool, got %d", byType["Tool"])
	}
}

func TestEmbeddingCache_LargeEmbedding(t *testing.T) {
	db := setupEmbeddingCacheTestDB(t)
	defer db.Close()

	cache := NewEmbeddingCache(db)

	// Create a large embedding (512 dimensions like voyage-3-lite)
	embedding := make([]float32, 512)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001
	}

	entity := &EntityEmbedding{
		EntityID:    "large-entity",
		EntityLabel: "Large Entity",
		EntityType:  "Concept",
		Embedding:   embedding,
		Model:       "voyage-3-lite",
		Provider:    "voyage",
		Dimension:   512,
	}

	if err := cache.Store(entity); err != nil {
		t.Fatalf("failed to store large embedding: %v", err)
	}

	retrieved, err := cache.Get("large-entity")
	if err != nil {
		t.Fatalf("failed to get large embedding: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected non-nil embedding")
	}
	if len(retrieved.Embedding) != 512 {
		t.Errorf("expected 512 dimensions, got %d", len(retrieved.Embedding))
	}
	// Verify first and last values
	if retrieved.Embedding[0] != 0.0 {
		t.Errorf("expected first value 0.0, got %f", retrieved.Embedding[0])
	}
	// Use approximate comparison for floating point
	expectedLast := float32(0.511)
	tolerance := float32(0.0001)
	if retrieved.Embedding[511] < expectedLast-tolerance || retrieved.Embedding[511] > expectedLast+tolerance {
		t.Errorf("expected last value ~0.511, got %f", retrieved.Embedding[511])
	}
}
