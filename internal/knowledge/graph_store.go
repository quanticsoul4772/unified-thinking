// Package knowledge implements knowledge graph storage and retrieval operations.
package knowledge

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// GraphStore provides CRUD operations for the knowledge graph
type GraphStore struct {
	client   *Neo4jClient
	database string
}

// NewGraphStore creates a new graph store
func NewGraphStore(client *Neo4jClient, database string) *GraphStore {
	return &GraphStore{
		client:   client,
		database: database,
	}
}

// CreateEntity stores an entity in the knowledge graph
func (s *GraphStore) CreateEntity(ctx context.Context, entity *Entity) error {
	now := time.Now().Unix()
	if entity.CreatedAt == 0 {
		entity.CreatedAt = now
	}
	entity.UpdatedAt = now

	query := `
		CREATE (e:Entity {
			id: $id,
			label: $label,
			type: $type,
			description: $description,
			created_at: $created_at,
			updated_at: $updated_at,
			metadata: $metadata
		})
		RETURN e.id as id
	`

	params := map[string]interface{}{
		"id":          entity.ID,
		"label":       entity.Label,
		"type":        string(entity.Type),
		"description": entity.Description,
		"created_at":  entity.CreatedAt,
		"updated_at":  entity.UpdatedAt,
		"metadata":    entity.Metadata,
	}

	_, err := s.client.ExecuteWrite(ctx, s.database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		if result.Next(ctx) {
			return result.Record().AsMap(), nil
		}
		return nil, result.Err()
	})

	return err
}

// GetEntity retrieves an entity by ID
func (s *GraphStore) GetEntity(ctx context.Context, entityID string) (*Entity, error) {
	query := `
		MATCH (e:Entity {id: $id})
		RETURN e.id as id, e.label as label, e.type as type,
		       e.description as description, e.created_at as created_at,
		       e.updated_at as updated_at, e.metadata as metadata
	`

	params := map[string]interface{}{
		"id": entityID,
	}

	result, err := s.client.ExecuteRead(ctx, s.database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			record := res.Record()
			entity := &Entity{
				ID:          record.Values[0].(string),
				Label:       record.Values[1].(string),
				Type:        EntityType(record.Values[2].(string)),
				Description: getStringOrEmpty(record.Values[3]),
				CreatedAt:   record.Values[4].(int64),
				UpdatedAt:   record.Values[5].(int64),
			}

			if metadata, ok := record.Values[6].(map[string]interface{}); ok {
				entity.Metadata = metadata
			}

			return entity, nil
		}

		if err := res.Err(); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("entity not found: %s", entityID)
	})

	if err != nil {
		return nil, err
	}

	if entity, ok := result.(*Entity); ok {
		return entity, nil
	}

	return nil, fmt.Errorf("unexpected result type")
}

// QueryEntitiesByType retrieves all entities of a specific type
func (s *GraphStore) QueryEntitiesByType(ctx context.Context, entityType EntityType, limit int) ([]*Entity, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		MATCH (e:Entity {type: $type})
		RETURN e.id as id, e.label as label, e.type as type,
		       e.description as description, e.created_at as created_at,
		       e.updated_at as updated_at, e.metadata as metadata
		ORDER BY e.created_at DESC
		LIMIT $limit
	`

	params := map[string]interface{}{
		"type":  string(entityType),
		"limit": limit,
	}

	result, err := s.client.ExecuteRead(ctx, s.database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		entities := []*Entity{}
		for res.Next(ctx) {
			record := res.Record()
			entity := &Entity{
				ID:          record.Values[0].(string),
				Label:       record.Values[1].(string),
				Type:        EntityType(record.Values[2].(string)),
				Description: getStringOrEmpty(record.Values[3]),
				CreatedAt:   record.Values[4].(int64),
				UpdatedAt:   record.Values[5].(int64),
			}

			if metadata, ok := record.Values[6].(map[string]interface{}); ok {
				entity.Metadata = metadata
			}

			entities = append(entities, entity)
		}

		return entities, res.Err()
	})

	if err != nil {
		return nil, err
	}

	if entities, ok := result.([]*Entity); ok {
		return entities, nil
	}

	return nil, fmt.Errorf("unexpected result type")
}

// CreateRelationship creates a relationship between two entities
func (s *GraphStore) CreateRelationship(ctx context.Context, rel *Relationship) error {
	now := time.Now().Unix()
	if rel.CreatedAt == 0 {
		rel.CreatedAt = now
	}

	query := `
		MATCH (from:Entity {id: $from_id})
		MATCH (to:Entity {id: $to_id})
		CREATE (from)-[r:` + string(rel.Type) + ` {
			id: $id,
			strength: $strength,
			confidence: $confidence,
			source: $source,
			created_at: $created_at,
			metadata: $metadata
		}]->(to)
		RETURN r.id as id
	`

	params := map[string]interface{}{
		"from_id":    rel.FromID,
		"to_id":      rel.ToID,
		"id":         rel.ID,
		"strength":   rel.Strength,
		"confidence": rel.Confidence,
		"source":     rel.Source,
		"created_at": rel.CreatedAt,
		"metadata":   rel.Metadata,
	}

	_, err := s.client.ExecuteWrite(ctx, s.database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		if result.Next(ctx) {
			return result.Record().AsMap(), nil
		}
		return nil, result.Err()
	})

	return err
}

// GetRelationships retrieves all relationships for an entity
func (s *GraphStore) GetRelationships(ctx context.Context, entityID string, direction string) ([]*Relationship, error) {
	var query string
	switch direction {
	case "outgoing":
		query = `
			MATCH (from:Entity {id: $id})-[r]->(to:Entity)
			RETURN type(r) as type, r.id as id, from.id as from_id, to.id as to_id,
			       r.strength as strength, r.confidence as confidence, r.source as source,
			       r.created_at as created_at, r.metadata as metadata
		`
	case "incoming":
		query = `
			MATCH (from:Entity)-[r]->(to:Entity {id: $id})
			RETURN type(r) as type, r.id as id, from.id as from_id, to.id as to_id,
			       r.strength as strength, r.confidence as confidence, r.source as source,
			       r.created_at as created_at, r.metadata as metadata
		`
	default: // both
		query = `
			MATCH (e:Entity {id: $id})
			MATCH (e)-[r]-(other:Entity)
			RETURN type(r) as type, r.id as id, startNode(r).id as from_id, endNode(r).id as to_id,
			       r.strength as strength, r.confidence as confidence, r.source as source,
			       r.created_at as created_at, r.metadata as metadata
		`
	}

	params := map[string]interface{}{
		"id": entityID,
	}

	result, err := s.client.ExecuteRead(ctx, s.database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		relationships := []*Relationship{}
		for res.Next(ctx) {
			record := res.Record()
			rel := &Relationship{
				Type:       RelationshipType(record.Values[0].(string)),
				ID:         record.Values[1].(string),
				FromID:     record.Values[2].(string),
				ToID:       record.Values[3].(string),
				Strength:   getFloat64OrZero(record.Values[4]),
				Confidence: getFloat64OrZero(record.Values[5]),
				Source:     getStringOrEmpty(record.Values[6]),
				CreatedAt:  record.Values[7].(int64),
			}

			if metadata, ok := record.Values[8].(map[string]interface{}); ok {
				rel.Metadata = metadata
			}

			relationships = append(relationships, rel)
		}

		return relationships, res.Err()
	})

	if err != nil {
		return nil, err
	}

	if relationships, ok := result.([]*Relationship); ok {
		return relationships, nil
	}

	return nil, fmt.Errorf("unexpected result type")
}

// CreateObservation stores a temporal observation about an entity
func (s *GraphStore) CreateObservation(ctx context.Context, obs *Observation) error {
	query := `
		MATCH (e:Entity {id: $entity_id})
		CREATE (o:Observation {
			id: $id,
			entity_id: $entity_id,
			content: $content,
			confidence: $confidence,
			source: $source,
			timestamp: $timestamp,
			metadata: $metadata
		})
		CREATE (e)-[:HAS_OBSERVATION]->(o)
		RETURN o.id as id
	`

	params := map[string]interface{}{
		"entity_id":  obs.EntityID,
		"id":         obs.ID,
		"content":    obs.Content,
		"confidence": obs.Confidence,
		"source":     obs.Source,
		"timestamp":  obs.Timestamp,
		"metadata":   obs.Metadata,
	}

	_, err := s.client.ExecuteWrite(ctx, s.database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		if result.Next(ctx) {
			return result.Record().AsMap(), nil
		}
		return nil, result.Err()
	})

	return err
}

// QueryEntitiesWithinHops finds entities within N hops of a starting entity
func (s *GraphStore) QueryEntitiesWithinHops(ctx context.Context, entityID string, maxHops int, relationshipTypes []RelationshipType) ([]*Entity, error) {
	if maxHops <= 0 {
		maxHops = 2
	}

	// Build relationship type filter
	relFilter := ""
	if len(relationshipTypes) > 0 {
		relFilter = ":"
		for i, relType := range relationshipTypes {
			if i > 0 {
				relFilter += "|"
			}
			relFilter += string(relType)
		}
	}

	query := fmt.Sprintf(`
		MATCH path = (start:Entity {id: $id})-[r%s*1..%d]-(connected:Entity)
		WHERE start.id <> connected.id
		RETURN DISTINCT connected.id as id, connected.label as label, connected.type as type,
		       connected.description as description, connected.created_at as created_at,
		       connected.updated_at as updated_at, connected.metadata as metadata,
		       length(path) as hops
		ORDER BY hops ASC, connected.created_at DESC
	`, relFilter, maxHops)

	params := map[string]interface{}{
		"id": entityID,
	}

	result, err := s.client.ExecuteRead(ctx, s.database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		entities := []*Entity{}
		for res.Next(ctx) {
			record := res.Record()
			entity := &Entity{
				ID:          record.Values[0].(string),
				Label:       record.Values[1].(string),
				Type:        EntityType(record.Values[2].(string)),
				Description: getStringOrEmpty(record.Values[3]),
				CreatedAt:   record.Values[4].(int64),
				UpdatedAt:   record.Values[5].(int64),
			}

			if metadata, ok := record.Values[6].(map[string]interface{}); ok {
				entity.Metadata = metadata
			}

			entities = append(entities, entity)
		}

		return entities, res.Err()
	})

	if err != nil {
		return nil, err
	}

	if entities, ok := result.([]*Entity); ok {
		return entities, nil
	}

	return nil, fmt.Errorf("unexpected result type")
}

// SearchEntities performs fulltext search across entities
func (s *GraphStore) SearchEntities(ctx context.Context, searchTerm string, limit int) ([]*Entity, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		CALL db.index.fulltext.queryNodes('entity_fulltext', $search_term)
		YIELD node, score
		RETURN node.id as id, node.label as label, node.type as type,
		       node.description as description, node.created_at as created_at,
		       node.updated_at as updated_at, node.metadata as metadata, score
		ORDER BY score DESC
		LIMIT $limit
	`

	params := map[string]interface{}{
		"search_term": searchTerm,
		"limit":       limit,
	}

	result, err := s.client.ExecuteRead(ctx, s.database, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		entities := []*Entity{}
		for res.Next(ctx) {
			record := res.Record()
			entity := &Entity{
				ID:          record.Values[0].(string),
				Label:       record.Values[1].(string),
				Type:        EntityType(record.Values[2].(string)),
				Description: getStringOrEmpty(record.Values[3]),
				CreatedAt:   record.Values[4].(int64),
				UpdatedAt:   record.Values[5].(int64),
			}

			if metadata, ok := record.Values[6].(map[string]interface{}); ok {
				entity.Metadata = metadata
			}

			entities = append(entities, entity)
		}

		return entities, res.Err()
	})

	if err != nil {
		return nil, err
	}

	if entities, ok := result.([]*Entity); ok {
		return entities, nil
	}

	return nil, fmt.Errorf("unexpected result type")
}

// Helper functions

func getStringOrEmpty(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

func getFloat64OrZero(value interface{}) float64 {
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}
