// Package knowledge integrates knowledge graph with episodic memory.
package knowledge

import (
	"context"
	"fmt"
	"log"
	"time"

	"unified-thinking/internal/knowledge/extraction"
)

// TrajectoryExtractor extracts entities and relationships from reasoning trajectories
type TrajectoryExtractor struct {
	extractor *extraction.HybridExtractor
	kg        *KnowledgeGraph
}

// NewTrajectoryExtractor creates a new trajectory extractor
func NewTrajectoryExtractor(kg *KnowledgeGraph, enableLLM bool) *TrajectoryExtractor {
	return &TrajectoryExtractor{
		extractor: extraction.NewHybridExtractor(extraction.HybridConfig{
			EnableLLM: enableLLM,
		}),
		kg: kg,
	}
}

// ExtractFromTrajectory extracts entities from a trajectory and stores in knowledge graph
func (te *TrajectoryExtractor) ExtractFromTrajectory(ctx context.Context, trajectoryID string, problem string, steps []string) error {
	if !te.kg.IsEnabled() {
		return nil // Silently skip if KG disabled
	}

	// Extract entities from problem description
	problemResult, err := te.extractor.Extract(problem)
	if err != nil {
		return fmt.Errorf("failed to extract from problem: %w", err)
	}

	// Store problem as entity
	problemEntity := &Entity{
		ID:          fmt.Sprintf("problem-%s", trajectoryID),
		Label:       truncate(problem, 100),
		Type:        EntityTypeProblem,
		Description: problem,
		Metadata: map[string]interface{}{
			"trajectory_id": trajectoryID,
			"extracted_at":  time.Now().Unix(),
		},
	}

	if err := te.kg.StoreEntity(ctx, problemEntity, problem); err != nil {
		log.Printf("[WARN] Failed to store problem entity: %v", err)
	}

	// Extract and store entities from problem content
	for _, entity := range problemResult.Entities {
		e := &Entity{
			ID:    fmt.Sprintf("%s-%s", entity.Type, entity.Text),
			Label: entity.Text,
			Type:  mapExtractedType(entity.Type),
			Metadata: map[string]interface{}{
				"extracted_from": trajectoryID,
				"confidence":     entity.Confidence,
				"method":         entity.Method,
			},
		}

		if err := te.kg.StoreEntity(ctx, e, entity.Text); err != nil {
			log.Printf("[WARN] Failed to store entity %s: %v", e.ID, err)
			continue
		}

		// Create relationship from problem to entity
		rel := &Relationship{
			ID:         fmt.Sprintf("rel-%s-%s", problemEntity.ID, e.ID),
			FromID:     problemEntity.ID,
			ToID:       e.ID,
			Type:       RelationshipRelatesTo,
			Strength:   entity.Confidence,
			Confidence: entity.Confidence,
			Source:     "trajectory_extraction",
		}

		if err := te.kg.CreateRelationship(ctx, rel); err != nil {
			log.Printf("[WARN] Failed to create relationship: %v", err)
		}
	}

	// Extract and store causal relationships
	for _, rel := range problemResult.Relationships {
		fromID := fmt.Sprintf("concept-%s", rel.From)
		toID := fmt.Sprintf("concept-%s", rel.To)

		// Create entities if they don't exist
		fromEntity := &Entity{
			ID:    fromID,
			Label: rel.From,
			Type:  EntityTypeConcept,
		}
		toEntity := &Entity{
			ID:    toID,
			Label: rel.To,
			Type:  EntityTypeConcept,
		}

		te.kg.StoreEntity(ctx, fromEntity, rel.From)
		te.kg.StoreEntity(ctx, toEntity, rel.To)

		// Create relationship
		relationship := &Relationship{
			ID:         fmt.Sprintf("rel-%s-%s", fromID, toID),
			FromID:     fromID,
			ToID:       toID,
			Type:       mapRelationshipType(rel.Type),
			Strength:   rel.Strength,
			Confidence: rel.Confidence,
			Source:     "trajectory_extraction",
		}

		if err := te.kg.CreateRelationship(ctx, relationship); err != nil {
			log.Printf("[WARN] Failed to create causal relationship: %v", err)
		}
	}

	return nil
}

// mapExtractedType maps extraction entity types to knowledge graph entity types
func mapExtractedType(extractedType string) EntityType {
	switch extractedType {
	case "url", "file_path", "identifier":
		return EntityTypeTool
	case "email":
		return EntityTypePerson
	default:
		return EntityTypeConcept
	}
}

// mapRelationshipType maps extraction relationship types to knowledge graph types
func mapRelationshipType(extractedType string) RelationshipType {
	switch extractedType {
	case "CAUSES":
		return RelationshipCauses
	case "ENABLES":
		return RelationshipEnables
	case "CONTRADICTS":
		return RelationshipContradicts
	case "BUILDS_UPON":
		return RelationshipBuildsUpon
	default:
		return RelationshipRelatesTo
	}
}

// truncate shortens a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
