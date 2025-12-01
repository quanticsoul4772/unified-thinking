// Package knowledge integrates knowledge graph with Thompson Sampling RL.
package knowledge

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RLContextRetriever retrieves contextual information from knowledge graph for RL strategy selection
type RLContextRetriever struct {
	kg                *KnowledgeGraph
	similarityThreshold float32
}

// NewRLContextRetriever creates a new RL context retriever with default similarity threshold (0.7)
func NewRLContextRetriever(kg *KnowledgeGraph) *RLContextRetriever {
	return &RLContextRetriever{
		kg:                kg,
		similarityThreshold: 0.7,
	}
}

// NewRLContextRetrieverWithThreshold creates retriever with custom similarity threshold (for testing)
func NewRLContextRetrieverWithThreshold(kg *KnowledgeGraph, threshold float32) *RLContextRetriever {
	return &RLContextRetriever{
		kg:                kg,
		similarityThreshold: threshold,
	}
}

// GetSimilarProblems retrieves similar past problems using semantic search
func (rcr *RLContextRetriever) GetSimilarProblems(ctx context.Context, problemDesc string, limit int) ([]*Entity, error) {
	if !rcr.kg.IsEnabled() {
		return nil, nil
	}

	// Semantic search for similar problems
	results, err := rcr.kg.SearchSemantic(ctx, problemDesc, limit, rcr.similarityThreshold)
	if err != nil {
		log.Printf("[WARN] Knowledge graph semantic search failed: %v", err)
		return nil, nil
	}

	// Convert search results to entities
	entities := make([]*Entity, 0, len(results))
	for _, result := range results {
		entity, err := rcr.kg.GetEntity(ctx, result.ID)
		if err != nil {
			continue
		}

		// Filter for problem entities only
		if entity.Type == EntityTypeProblem {
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

// GetStrategyPerformance retrieves strategy performance for similar problems
func (rcr *RLContextRetriever) GetStrategyPerformance(ctx context.Context, problemDesc string) (map[string]float64, error) {
	if !rcr.kg.IsEnabled() {
		return nil, nil
	}

	// Find similar problems
	similarProblems, err := rcr.GetSimilarProblems(ctx, problemDesc, 5)
	if err != nil || len(similarProblems) == 0 {
		return nil, nil
	}

	// Extract strategy performance from metadata
	strategyStats := make(map[string]struct {
		successes int
		total     int
	})

	for _, problem := range similarProblems {
		if metadata, ok := problem.Metadata["strategy"]; ok {
			if strategyName, ok := metadata.(string); ok {
				stats := strategyStats[strategyName]
				stats.total++

				if success, ok := problem.Metadata["success"].(bool); ok && success {
					stats.successes++
				}

				strategyStats[strategyName] = stats
			}
		}
	}

	// Calculate success rates
	performance := make(map[string]float64)
	for strategy, stats := range strategyStats {
		if stats.total > 0 {
			performance[strategy] = float64(stats.successes) / float64(stats.total)
		}
	}

	return performance, nil
}

// EnrichProblemContext adds knowledge graph context to problem description
func (rcr *RLContextRetriever) EnrichProblemContext(ctx context.Context, problemDesc string) (string, error) {
	if !rcr.kg.IsEnabled() {
		return problemDesc, nil
	}

	// Get similar problems
	similar, err := rcr.GetSimilarProblems(ctx, problemDesc, 3)
	if err != nil || len(similar) == 0 {
		return problemDesc, nil
	}

	// Build enriched context
	enriched := fmt.Sprintf("Problem: %s\n\nSimilar past problems:\n", problemDesc)
	for i, prob := range similar {
		enriched += fmt.Sprintf("%d. %s", i+1, prob.Description)

		if strategy, ok := prob.Metadata["strategy"].(string); ok {
			enriched += fmt.Sprintf(" (used: %s)", strategy)
		}

		if success, ok := prob.Metadata["success"].(bool); ok {
			if success {
				enriched += " [success]"
			} else {
				enriched += " [failed]"
			}
		}

		enriched += "\n"
	}

	return enriched, nil
}

// RecordStrategyOutcome records strategy selection outcome in knowledge graph
func (rcr *RLContextRetriever) RecordStrategyOutcome(ctx context.Context, problemDesc string, strategy string, success bool, confidence float64) error {
	if !rcr.kg.IsEnabled() {
		return nil
	}

	// Create problem entity
	problemEntity := &Entity{
		ID:          fmt.Sprintf("problem-%d", time.Now().Unix()),
		Label:       truncate(problemDesc, 100),
		Type:        EntityTypeProblem,
		Description: problemDesc,
		Metadata: map[string]interface{}{
			"strategy":   strategy,
			"success":    success,
			"confidence": confidence,
			"timestamp":  time.Now().Unix(),
		},
	}

	if err := rcr.kg.StoreEntity(ctx, problemEntity, problemDesc); err != nil {
		log.Printf("[WARN] Failed to record problem in KG: %v", err)
	}

	// Create strategy entity
	strategyEntity := &Entity{
		ID:    fmt.Sprintf("strategy-%s", strategy),
		Label: strategy,
		Type:  EntityTypeStrategy,
	}

	_ = rcr.kg.StoreEntity(ctx, strategyEntity, strategy)

	// Create relationship
	relType := RelationshipEnables
	if !success {
		relType = RelationshipContradicts
	}

	rel := &Relationship{
		ID:         fmt.Sprintf("rel-%s-%s", problemEntity.ID, strategyEntity.ID),
		FromID:     strategyEntity.ID,
		ToID:       problemEntity.ID,
		Type:       relType,
		Strength:   confidence,
		Confidence: confidence,
		Source:     "rl_outcome",
	}

	if err := rcr.kg.CreateRelationship(ctx, rel); err != nil {
		log.Printf("[WARN] Failed to create strategy relationship: %v", err)
	}

	return nil
}
