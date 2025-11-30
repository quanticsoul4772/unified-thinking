// Package knowledge provides embedding caching for knowledge graph entities.
package knowledge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// EntityEmbedding represents a cached embedding for an entity
type EntityEmbedding struct {
	EntityID    string    `json:"entity_id"`
	EntityLabel string    `json:"entity_label"`
	EntityType  string    `json:"entity_type"`
	Embedding   []float32 `json:"embedding"`
	Model       string    `json:"model"`
	Provider    string    `json:"provider"`
	Dimension   int       `json:"dimension"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

// EmbeddingCache provides SQLite-based caching for entity embeddings
type EmbeddingCache struct {
	db *sql.DB
}

// NewEmbeddingCache creates a new embedding cache
func NewEmbeddingCache(db *sql.DB) *EmbeddingCache {
	return &EmbeddingCache{db: db}
}

// Store caches an entity embedding
func (ec *EmbeddingCache) Store(embedding *EntityEmbedding) error {
	now := time.Now().Unix()
	if embedding.CreatedAt == 0 {
		embedding.CreatedAt = now
	}
	embedding.UpdatedAt = now

	// Serialize embedding to JSON
	embeddingJSON, err := json.Marshal(embedding.Embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	_, err = ec.db.Exec(`
		INSERT OR REPLACE INTO entity_embeddings (
			entity_id, entity_label, entity_type, embedding,
			model, provider, dimension, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		embedding.EntityID,
		embedding.EntityLabel,
		embedding.EntityType,
		embeddingJSON,
		embedding.Model,
		embedding.Provider,
		embedding.Dimension,
		embedding.CreatedAt,
		embedding.UpdatedAt,
	)

	return err
}

// Get retrieves a cached embedding by entity ID
func (ec *EmbeddingCache) Get(entityID string) (*EntityEmbedding, error) {
	var entityLabel, entityType, embeddingJSON, model, provider string
	var dimension int
	var createdAt, updatedAt int64

	err := ec.db.QueryRow(`
		SELECT entity_label, entity_type, embedding, model, provider,
		       dimension, created_at, updated_at
		FROM entity_embeddings
		WHERE entity_id = ?
	`, entityID).Scan(
		&entityLabel, &entityType, &embeddingJSON, &model, &provider,
		&dimension, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query embedding cache: %w", err)
	}

	// Deserialize embedding
	var embedding []float32
	if err := json.Unmarshal([]byte(embeddingJSON), &embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}

	return &EntityEmbedding{
		EntityID:    entityID,
		EntityLabel: entityLabel,
		EntityType:  entityType,
		Embedding:   embedding,
		Model:       model,
		Provider:    provider,
		Dimension:   dimension,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// GetByType retrieves all cached embeddings of a specific entity type
func (ec *EmbeddingCache) GetByType(entityType string, limit int) ([]*EntityEmbedding, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := ec.db.Query(`
		SELECT entity_id, entity_label, entity_type, embedding, model, provider,
		       dimension, created_at, updated_at
		FROM entity_embeddings
		WHERE entity_type = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, entityType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query embeddings by type: %w", err)
	}
	defer rows.Close()

	embeddings := []*EntityEmbedding{}
	for rows.Next() {
		var entityID, entityLabel, entityType, embeddingJSON, model, provider string
		var dimension int
		var createdAt, updatedAt int64

		err := rows.Scan(
			&entityID, &entityLabel, &entityType, &embeddingJSON, &model, &provider,
			&dimension, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan embedding: %w", err)
		}

		var embedding []float32
		if err := json.Unmarshal([]byte(embeddingJSON), &embedding); err != nil {
			continue // Skip malformed embeddings
		}

		embeddings = append(embeddings, &EntityEmbedding{
			EntityID:    entityID,
			EntityLabel: entityLabel,
			EntityType:  entityType,
			Embedding:   embedding,
			Model:       model,
			Provider:    provider,
			Dimension:   dimension,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	return embeddings, rows.Err()
}

// Delete removes a cached embedding
func (ec *EmbeddingCache) Delete(entityID string) error {
	_, err := ec.db.Exec("DELETE FROM entity_embeddings WHERE entity_id = ?", entityID)
	return err
}

// Count returns the total number of cached embeddings
func (ec *EmbeddingCache) Count() (int, error) {
	var count int
	err := ec.db.QueryRow("SELECT COUNT(*) FROM entity_embeddings").Scan(&count)
	return count, err
}

// GetCacheStats returns cache statistics
func (ec *EmbeddingCache) GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	count, err := ec.Count()
	if err != nil {
		return nil, err
	}
	stats["total_cached"] = count

	// Count by type
	rows, err := ec.db.Query(`
		SELECT entity_type, COUNT(*) as count
		FROM entity_embeddings
		GROUP BY entity_type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byType := make(map[string]int)
	for rows.Next() {
		var entityType string
		var count int
		if err := rows.Scan(&entityType, &count); err == nil {
			byType[entityType] = count
		}
	}
	stats["by_type"] = byType

	return stats, nil
}
