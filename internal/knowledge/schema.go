// Package knowledge defines the knowledge graph schema for Neo4j.
package knowledge

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"unified-thinking/internal/types"
)

// EntityType represents the type of entity in the knowledge graph
type EntityType string

const (
	EntityTypeConcept  EntityType = "Concept"
	EntityTypePerson   EntityType = "Person"
	EntityTypeTool     EntityType = "Tool"
	EntityTypeFile     EntityType = "File"
	EntityTypeDecision EntityType = "Decision"
	EntityTypeStrategy EntityType = "Strategy"
	EntityTypeProblem  EntityType = "Problem"
)

// RelationshipType represents the type of relationship between entities
type RelationshipType string

const (
	RelationshipCauses         RelationshipType = "CAUSES"
	RelationshipEnables        RelationshipType = "ENABLES"
	RelationshipContradicts    RelationshipType = "CONTRADICTS"
	RelationshipBuildsUpon     RelationshipType = "BUILDS_UPON"
	RelationshipRelatesTo      RelationshipType = "RELATES_TO"
	RelationshipHasObservation RelationshipType = "HAS_OBSERVATION"
	RelationshipUsedInContext  RelationshipType = "USED_IN_CONTEXT"
)

// Entity represents a node in the knowledge graph
type Entity struct {
	ID          string         `json:"id"`
	Label       string         `json:"label"`
	Type        EntityType     `json:"type"`
	Description string         `json:"description,omitempty"`
	CreatedAt   int64          `json:"created_at"`
	UpdatedAt   int64          `json:"updated_at"`
	Metadata    types.Metadata `json:"metadata,omitempty"`
}

// Relationship represents an edge in the knowledge graph
type Relationship struct {
	ID         string           `json:"id"`
	FromID     string           `json:"from_id"`
	ToID       string           `json:"to_id"`
	Type       RelationshipType `json:"type"`
	Strength   float64          `json:"strength"`   // 0.0 to 1.0
	Confidence float64          `json:"confidence"` // 0.0 to 1.0
	Source     string           `json:"source,omitempty"`
	CreatedAt  int64            `json:"created_at"`
	Metadata   types.Metadata   `json:"metadata,omitempty"`
}

// Observation represents a temporal fact about an entity
type Observation struct {
	ID         string         `json:"id"`
	EntityID   string         `json:"entity_id"`
	Content    string         `json:"content"`
	Confidence float64        `json:"confidence"`
	Source     string         `json:"source"`
	Timestamp  int64          `json:"timestamp"`
	Metadata   types.Metadata `json:"metadata,omitempty"`
}

// InitializeSchema creates constraints and indexes for the knowledge graph
func InitializeSchema(ctx context.Context, client *Neo4jClient, database string) error {
	queries := []string{
		// Constraints for uniqueness
		"CREATE CONSTRAINT entity_id_unique IF NOT EXISTS FOR (e:Entity) REQUIRE e.id IS UNIQUE",
		"CREATE CONSTRAINT observation_id_unique IF NOT EXISTS FOR (o:Observation) REQUIRE o.id IS UNIQUE",

		// Indexes for performance (<50ms query target)
		"CREATE INDEX entity_type_idx IF NOT EXISTS FOR (e:Entity) ON (e.type)",
		"CREATE INDEX entity_label_idx IF NOT EXISTS FOR (e:Entity) ON (e.label)",
		"CREATE INDEX entity_created_idx IF NOT EXISTS FOR (e:Entity) ON (e.created_at)",
		"CREATE INDEX observation_timestamp_idx IF NOT EXISTS FOR (o:Observation) ON (o.timestamp)",
		"CREATE INDEX observation_entity_idx IF NOT EXISTS FOR (o:Observation) ON (o.entity_id)",

		// Fulltext index for semantic search
		"CREATE FULLTEXT INDEX entity_fulltext IF NOT EXISTS FOR (e:Entity) ON EACH [e.label, e.description]",
		"CREATE FULLTEXT INDEX observation_fulltext IF NOT EXISTS FOR (o:Observation) ON EACH [o.content]",
	}

	for _, query := range queries {
		_, err := client.ExecuteWrite(ctx, database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx, query, nil)
			if err != nil {
				return nil, err
			}
			return result.Consume(ctx)
		})
		if err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// DropSchema removes all constraints and indexes (for testing cleanup)
func DropSchema(ctx context.Context, client *Neo4jClient, database string) error {
	queries := []string{
		"DROP CONSTRAINT entity_id_unique IF EXISTS",
		"DROP CONSTRAINT observation_id_unique IF EXISTS",
		"DROP INDEX entity_type_idx IF EXISTS",
		"DROP INDEX entity_label_idx IF EXISTS",
		"DROP INDEX entity_created_idx IF EXISTS",
		"DROP INDEX observation_timestamp_idx IF EXISTS",
		"DROP INDEX observation_entity_idx IF EXISTS",
		"DROP FULLTEXT INDEX entity_fulltext IF EXISTS",
		"DROP FULLTEXT INDEX observation_fulltext IF EXISTS",
	}

	for _, query := range queries {
		_, err := client.ExecuteWrite(ctx, database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx, query, nil)
			if err != nil {
				return nil, err
			}
			return result.Consume(ctx)
		})
		if err != nil {
			// Ignore errors for DROP IF EXISTS
			continue
		}
	}

	return nil
}

// ClearAllData removes all nodes and relationships (for testing cleanup)
func ClearAllData(ctx context.Context, client *Neo4jClient, database string) error {
	query := "MATCH (n) DETACH DELETE n"
	_, err := client.ExecuteWrite(ctx, database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})
	return err
}
