package knowledge

import (
	"testing"
)

func TestEntityType_Constants(t *testing.T) {
	// Verify entity type constants have expected values
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{EntityTypeConcept, "Concept"},
		{EntityTypePerson, "Person"},
		{EntityTypeTool, "Tool"},
		{EntityTypeFile, "File"},
		{EntityTypeDecision, "Decision"},
		{EntityTypeStrategy, "Strategy"},
		{EntityTypeProblem, "Problem"},
	}

	for _, tt := range tests {
		t.Run(string(tt.entityType), func(t *testing.T) {
			if string(tt.entityType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.entityType))
			}
		})
	}
}

func TestRelationshipType_Constants(t *testing.T) {
	// Verify relationship type constants have expected values
	tests := []struct {
		relType  RelationshipType
		expected string
	}{
		{RelationshipCauses, "CAUSES"},
		{RelationshipEnables, "ENABLES"},
		{RelationshipContradicts, "CONTRADICTS"},
		{RelationshipBuildsUpon, "BUILDS_UPON"},
		{RelationshipRelatesTo, "RELATES_TO"},
		{RelationshipHasObservation, "HAS_OBSERVATION"},
		{RelationshipUsedInContext, "USED_IN_CONTEXT"},
	}

	for _, tt := range tests {
		t.Run(string(tt.relType), func(t *testing.T) {
			if string(tt.relType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.relType))
			}
		})
	}
}

func TestEntity_Struct(t *testing.T) {
	entity := Entity{
		ID:          "test-id",
		Label:       "Test Entity",
		Type:        EntityTypeConcept,
		Description: "A test entity",
		CreatedAt:   1699900000,
		UpdatedAt:   1699900100,
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	if entity.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got '%s'", entity.ID)
	}
	if entity.Label != "Test Entity" {
		t.Errorf("expected Label 'Test Entity', got '%s'", entity.Label)
	}
	if entity.Type != EntityTypeConcept {
		t.Errorf("expected Type Concept, got '%s'", entity.Type)
	}
	if entity.Description != "A test entity" {
		t.Errorf("expected Description 'A test entity', got '%s'", entity.Description)
	}
	if entity.CreatedAt != 1699900000 {
		t.Errorf("expected CreatedAt 1699900000, got %d", entity.CreatedAt)
	}
	if entity.UpdatedAt != 1699900100 {
		t.Errorf("expected UpdatedAt 1699900100, got %d", entity.UpdatedAt)
	}
	if source, ok := entity.Metadata["source"].(string); !ok || source != "test" {
		t.Error("expected Metadata['source'] to be 'test'")
	}
}

func TestRelationship_Struct(t *testing.T) {
	rel := Relationship{
		ID:         "rel-1",
		FromID:     "entity-a",
		ToID:       "entity-b",
		Type:       RelationshipCauses,
		Strength:   0.8,
		Confidence: 0.9,
		Source:     "test-source",
		CreatedAt:  1699900000,
		Metadata: map[string]interface{}{
			"context": "test",
		},
	}

	if rel.ID != "rel-1" {
		t.Errorf("expected ID 'rel-1', got '%s'", rel.ID)
	}
	if rel.FromID != "entity-a" {
		t.Errorf("expected FromID 'entity-a', got '%s'", rel.FromID)
	}
	if rel.ToID != "entity-b" {
		t.Errorf("expected ToID 'entity-b', got '%s'", rel.ToID)
	}
	if rel.Type != RelationshipCauses {
		t.Errorf("expected Type CAUSES, got '%s'", rel.Type)
	}
	if rel.Strength != 0.8 {
		t.Errorf("expected Strength 0.8, got %f", rel.Strength)
	}
	if rel.Confidence != 0.9 {
		t.Errorf("expected Confidence 0.9, got %f", rel.Confidence)
	}
	if rel.Source != "test-source" {
		t.Errorf("expected Source 'test-source', got '%s'", rel.Source)
	}
}

func TestObservation_Struct(t *testing.T) {
	obs := Observation{
		ID:         "obs-1",
		EntityID:   "entity-a",
		Content:    "This is an observation",
		Confidence: 0.95,
		Source:     "test-source",
		Timestamp:  1699900000,
		Metadata: map[string]interface{}{
			"context": "test",
		},
	}

	if obs.ID != "obs-1" {
		t.Errorf("expected ID 'obs-1', got '%s'", obs.ID)
	}
	if obs.EntityID != "entity-a" {
		t.Errorf("expected EntityID 'entity-a', got '%s'", obs.EntityID)
	}
	if obs.Content != "This is an observation" {
		t.Errorf("expected Content 'This is an observation', got '%s'", obs.Content)
	}
	if obs.Confidence != 0.95 {
		t.Errorf("expected Confidence 0.95, got %f", obs.Confidence)
	}
	if obs.Source != "test-source" {
		t.Errorf("expected Source 'test-source', got '%s'", obs.Source)
	}
	if obs.Timestamp != 1699900000 {
		t.Errorf("expected Timestamp 1699900000, got %d", obs.Timestamp)
	}
}

func TestEntity_WithNilMetadata(t *testing.T) {
	entity := Entity{
		ID:       "no-metadata",
		Label:    "No Metadata Entity",
		Type:     EntityTypePerson,
		Metadata: nil,
	}

	if entity.Metadata != nil {
		t.Error("expected nil Metadata")
	}
}

func TestRelationship_EdgeCases(t *testing.T) {
	// Test with zero values
	rel := Relationship{
		Strength:   0.0,
		Confidence: 0.0,
	}

	if rel.Strength != 0.0 {
		t.Errorf("expected Strength 0.0, got %f", rel.Strength)
	}
	if rel.Confidence != 0.0 {
		t.Errorf("expected Confidence 0.0, got %f", rel.Confidence)
	}

	// Test with max values
	rel2 := Relationship{
		Strength:   1.0,
		Confidence: 1.0,
	}

	if rel2.Strength != 1.0 {
		t.Errorf("expected Strength 1.0, got %f", rel2.Strength)
	}
	if rel2.Confidence != 1.0 {
		t.Errorf("expected Confidence 1.0, got %f", rel2.Confidence)
	}
}

func TestEntityType_StringConversion(t *testing.T) {
	// Verify types can be converted to strings and back
	entityType := EntityTypeConcept
	str := string(entityType)
	back := EntityType(str)

	if back != EntityTypeConcept {
		t.Errorf("expected Concept after round-trip, got %s", back)
	}
}

func TestRelationshipType_StringConversion(t *testing.T) {
	// Verify relationship types can be converted to strings and back
	relType := RelationshipBuildsUpon
	str := string(relType)
	back := RelationshipType(str)

	if back != RelationshipBuildsUpon {
		t.Errorf("expected BUILDS_UPON after round-trip, got %s", back)
	}
}
