package knowledge

import (
	"context"
	"testing"
	"time"
)

// TestGraphStore_EntityOperations tests entity CRUD operations
func TestGraphStore_EntityOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := DefaultConfig()
	client, err := NewNeo4jClient(cfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer client.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize schema
	if err := InitializeSchema(ctx, client, cfg.Database); err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}

	// Clear test data
	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	store := NewGraphStore(client, cfg.Database)

	// Test CreateEntity
	entity := &Entity{
		ID:          "test-entity-1",
		Label:       "Test Entity",
		Type:        EntityTypeConcept,
		Description: "A test concept",
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	err = store.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("CreateEntity failed: %v", err)
	}

	// Verify timestamps were set
	if entity.CreatedAt == 0 {
		t.Error("CreatedAt should be set")
	}
	if entity.UpdatedAt == 0 {
		t.Error("UpdatedAt should be set")
	}

	// Test GetEntity
	retrieved, err := store.GetEntity(ctx, "test-entity-1")
	if err != nil {
		t.Fatalf("GetEntity failed: %v", err)
	}

	if retrieved.ID != entity.ID {
		t.Errorf("ID = %s, want %s", retrieved.ID, entity.ID)
	}
	if retrieved.Label != entity.Label {
		t.Errorf("Label = %s, want %s", retrieved.Label, entity.Label)
	}
	if retrieved.Type != entity.Type {
		t.Errorf("Type = %s, want %s", retrieved.Type, entity.Type)
	}

	// Test GetEntity not found
	_, err = store.GetEntity(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent entity")
	}

	// Cleanup
	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

// TestGraphStore_QueryEntitiesByType tests querying entities by type
func TestGraphStore_QueryEntitiesByType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := DefaultConfig()
	client, err := NewNeo4jClient(cfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer client.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	store := NewGraphStore(client, cfg.Database)

	// Create test entities
	entities := []*Entity{
		{ID: "concept-1", Label: "Concept 1", Type: EntityTypeConcept},
		{ID: "concept-2", Label: "Concept 2", Type: EntityTypeConcept},
		{ID: "tool-1", Label: "Tool 1", Type: EntityTypeTool},
	}

	for _, entity := range entities {
		if err := store.CreateEntity(ctx, entity); err != nil {
			t.Fatalf("CreateEntity failed: %v", err)
		}
	}

	// Query concepts
	concepts, err := store.QueryEntitiesByType(ctx, EntityTypeConcept, 10)
	if err != nil {
		t.Fatalf("QueryEntitiesByType failed: %v", err)
	}

	if len(concepts) != 2 {
		t.Errorf("Found %d concepts, want 2", len(concepts))
	}

	// Cleanup
	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

// TestGraphStore_RelationshipOperations tests relationship CRUD
func TestGraphStore_RelationshipOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := DefaultConfig()
	client, err := NewNeo4jClient(cfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer client.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	store := NewGraphStore(client, cfg.Database)

	// Create entities
	entity1 := &Entity{ID: "e1", Label: "Entity 1", Type: EntityTypeConcept}
	entity2 := &Entity{ID: "e2", Label: "Entity 2", Type: EntityTypeConcept}

	if err := store.CreateEntity(ctx, entity1); err != nil {
		t.Fatalf("CreateEntity 1 failed: %v", err)
	}
	if err := store.CreateEntity(ctx, entity2); err != nil {
		t.Fatalf("CreateEntity 2 failed: %v", err)
	}

	// Create relationship
	rel := &Relationship{
		ID:         "rel-1",
		FromID:     "e1",
		ToID:       "e2",
		Type:       RelationshipCauses,
		Strength:   0.8,
		Confidence: 0.9,
		Source:     "test",
	}

	err = store.CreateRelationship(ctx, rel)
	if err != nil {
		t.Fatalf("CreateRelationship failed: %v", err)
	}

	// Verify CreatedAt was set
	if rel.CreatedAt == 0 {
		t.Error("CreatedAt should be set")
	}

	// Get outgoing relationships
	rels, err := store.GetRelationships(ctx, "e1", "outgoing")
	if err != nil {
		t.Fatalf("GetRelationships failed: %v", err)
	}

	if len(rels) != 1 {
		t.Errorf("Found %d relationships, want 1", len(rels))
	}

	if len(rels) > 0 {
		if rels[0].Type != RelationshipCauses {
			t.Errorf("Type = %s, want %s", rels[0].Type, RelationshipCauses)
		}
		if rels[0].Strength != 0.8 {
			t.Errorf("Strength = %.2f, want 0.8", rels[0].Strength)
		}
	}

	// Get incoming relationships
	inRels, err := store.GetRelationships(ctx, "e2", "incoming")
	if err != nil {
		t.Fatalf("GetRelationships incoming failed: %v", err)
	}

	if len(inRels) != 1 {
		t.Errorf("Found %d incoming relationships, want 1", len(inRels))
	}

	// Cleanup
	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

// TestGraphStore_QueryEntitiesWithinHops tests graph traversal
func TestGraphStore_QueryEntitiesWithinHops(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := DefaultConfig()
	client, err := NewNeo4jClient(cfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer client.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	store := NewGraphStore(client, cfg.Database)

	// Create entity chain: e1 -> e2 -> e3
	entities := []*Entity{
		{ID: "e1", Label: "Entity 1", Type: EntityTypeConcept},
		{ID: "e2", Label: "Entity 2", Type: EntityTypeConcept},
		{ID: "e3", Label: "Entity 3", Type: EntityTypeConcept},
	}

	for _, entity := range entities {
		if err := store.CreateEntity(ctx, entity); err != nil {
			t.Fatalf("CreateEntity failed: %v", err)
		}
	}

	relationships := []*Relationship{
		{ID: "r1", FromID: "e1", ToID: "e2", Type: RelationshipEnables, Strength: 0.9},
		{ID: "r2", FromID: "e2", ToID: "e3", Type: RelationshipBuildsUpon, Strength: 0.8},
	}

	for _, rel := range relationships {
		if err := store.CreateRelationship(ctx, rel); err != nil {
			t.Fatalf("CreateRelationship failed: %v", err)
		}
	}

	// Query 1 hop from e1 (should find e2)
	connected1, err := store.QueryEntitiesWithinHops(ctx, "e1", 1, nil)
	if err != nil {
		t.Fatalf("QueryEntitiesWithinHops failed: %v", err)
	}

	if len(connected1) != 1 {
		t.Errorf("Found %d entities within 1 hop, want 1", len(connected1))
	}

	// Query 2 hops from e1 (should find e2 and e3)
	connected2, err := store.QueryEntitiesWithinHops(ctx, "e1", 2, nil)
	if err != nil {
		t.Fatalf("QueryEntitiesWithinHops failed: %v", err)
	}

	if len(connected2) != 2 {
		t.Errorf("Found %d entities within 2 hops, want 2", len(connected2))
	}

	// Cleanup
	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

// TestGraphStore_SearchEntities tests fulltext search
func TestGraphStore_SearchEntities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := DefaultConfig()
	client, err := NewNeo4jClient(cfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer client.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	store := NewGraphStore(client, cfg.Database)

	// Create test entities
	entities := []*Entity{
		{ID: "search-1", Label: "Database Performance", Type: EntityTypeConcept, Description: "Optimizing database queries"},
		{ID: "search-2", Label: "Query Optimization", Type: EntityTypeConcept, Description: "SQL performance tuning"},
		{ID: "search-3", Label: "Authentication System", Type: EntityTypeTool, Description: "User authentication"},
	}

	for _, entity := range entities {
		if err := store.CreateEntity(ctx, entity); err != nil {
			t.Fatalf("CreateEntity failed: %v", err)
		}
	}

	// Wait for fulltext index
	time.Sleep(2 * time.Second)

	// Search for "database"
	results, err := store.SearchEntities(ctx, "database", 10)
	if err != nil {
		t.Fatalf("SearchEntities failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("Expected at least 1 result for 'database' search")
	}

	// Cleanup
	if err := ClearAllData(ctx, client, cfg.Database); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}
