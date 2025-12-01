// Package handlers implements MCP tool handlers for knowledge graph operations.
package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/knowledge"
)

// KnowledgeHandlers provides MCP handlers for knowledge graph tools
type KnowledgeHandlers struct {
	kg *knowledge.KnowledgeGraph
}

// NewKnowledgeHandlers creates knowledge graph handlers
func NewKnowledgeHandlers(kg *knowledge.KnowledgeGraph) *KnowledgeHandlers {
	return &KnowledgeHandlers{kg: kg}
}

// StoreEntityRequest represents the store-entity tool request
type StoreEntityRequest struct {
	EntityID    string                 `json:"entity_id"`
	Label       string                 `json:"label"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Content     string                 `json:"content"` // For semantic search
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StoreEntityResponse represents the response
type StoreEntityResponse struct {
	EntityID  string `json:"entity_id"`
	Status    string `json:"status"`
	Cached    bool   `json:"cached"`
	CreatedAt int64  `json:"created_at"`
}

// HandleStoreEntity implements the store-entity MCP tool
func (kh *KnowledgeHandlers) HandleStoreEntity(request *StoreEntityRequest) (*mcp.CallToolResult, *StoreEntityResponse, error) {
	if !kh.kg.IsEnabled() {
		return nil, nil, fmt.Errorf("knowledge graph not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	entity := &knowledge.Entity{
		ID:          request.EntityID,
		Label:       request.Label,
		Type:        knowledge.EntityType(request.Type),
		Description: request.Description,
		Metadata:    request.Metadata,
	}

	if err := kh.kg.StoreEntity(ctx, entity, request.Content); err != nil {
		return nil, nil, fmt.Errorf("failed to store entity: %w", err)
	}

	response := &StoreEntityResponse{
		EntityID:  entity.ID,
		Status:    "stored",
		Cached:    true, // Embedding was cached
		CreatedAt: entity.CreatedAt,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// SearchKnowledgeGraphRequest represents the search-knowledge-graph tool request
type SearchKnowledgeGraphRequest struct {
	Query         string `json:"query"`
	SearchType    string `json:"search_type"` // "semantic", "graph", "hybrid"
	Limit         int    `json:"limit,omitempty"`
	MaxHops       int    `json:"max_hops,omitempty"`
	MinSimilarity float32 `json:"min_similarity,omitempty"`
}

// SearchKnowledgeGraphResponse represents the response
type SearchKnowledgeGraphResponse struct {
	Results       []*knowledge.Entity `json:"results"`
	ResultCount   int                 `json:"result_count"`
	SearchType    string              `json:"search_type"`
	LatencyMs     int64               `json:"latency_ms"`
}

// HandleSearchKnowledgeGraph implements the search-knowledge-graph MCP tool
func (kh *KnowledgeHandlers) HandleSearchKnowledgeGraph(request *SearchKnowledgeGraphRequest) (*mcp.CallToolResult, *SearchKnowledgeGraphResponse, error) {
	if !kh.kg.IsEnabled() {
		return nil, nil, fmt.Errorf("knowledge graph not enabled")
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var results []*knowledge.Entity
	var err error

	switch request.SearchType {
	case "semantic":
		// Semantic search only
		searchResults, searchErr := kh.kg.SearchSemantic(ctx, request.Query, request.Limit, request.MinSimilarity)
		if searchErr != nil {
			return nil, nil, fmt.Errorf("semantic search failed: %w", searchErr)
		}

		for _, sr := range searchResults {
			entity, getErr := kh.kg.GetEntity(ctx, sr.ID)
			if getErr == nil {
				results = append(results, entity)
			}
		}

	case "hybrid":
		// Combined semantic + graph search
		results, err = kh.kg.HybridSearch(ctx, request.Query, request.Limit, request.MaxHops)
		if err != nil {
			return nil, nil, fmt.Errorf("hybrid search failed: %w", err)
		}

	default:
		return nil, nil, fmt.Errorf("invalid search_type: %s (use 'semantic' or 'hybrid')", request.SearchType)
	}

	response := &SearchKnowledgeGraphResponse{
		Results:     results,
		ResultCount: len(results),
		SearchType:  request.SearchType,
		LatencyMs:   time.Since(start).Milliseconds(),
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// CreateRelationshipRequest represents the create-relationship tool request
type CreateRelationshipRequest struct {
	RelationshipID string                 `json:"relationship_id"`
	FromID         string                 `json:"from_id"`
	ToID           string                 `json:"to_id"`
	Type           string                 `json:"type"`
	Strength       float64                `json:"strength"`
	Confidence     float64                `json:"confidence"`
	Source         string                 `json:"source,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRelationshipResponse represents the response
type CreateRelationshipResponse struct {
	RelationshipID string `json:"relationship_id"`
	Status         string `json:"status"`
	CreatedAt      int64  `json:"created_at"`
}

// HandleCreateRelationship implements the create-relationship MCP tool
func (kh *KnowledgeHandlers) HandleCreateRelationship(request *CreateRelationshipRequest) (*mcp.CallToolResult, *CreateRelationshipResponse, error) {
	if !kh.kg.IsEnabled() {
		return nil, nil, fmt.Errorf("knowledge graph not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rel := &knowledge.Relationship{
		ID:         request.RelationshipID,
		FromID:     request.FromID,
		ToID:       request.ToID,
		Type:       knowledge.RelationshipType(request.Type),
		Strength:   request.Strength,
		Confidence: request.Confidence,
		Source:     request.Source,
		Metadata:   request.Metadata,
	}

	if err := kh.kg.CreateRelationship(ctx, rel); err != nil {
		return nil, nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	response := &CreateRelationshipResponse{
		RelationshipID: rel.ID,
		Status:         "created",
		CreatedAt:      rel.CreatedAt,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

