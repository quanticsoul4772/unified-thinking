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

// RegisterKnowledgeGraphTools registers all knowledge graph MCP tools
func RegisterKnowledgeGraphTools(mcpServer *mcp.Server, kh *KnowledgeHandlers) {
	// Tool 64: store-entity
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "store-entity",
		Description: `Store an entity in the knowledge graph with semantic indexing.

Stores entities in both Neo4j graph database and chromem-go vector search for hybrid retrieval.

**Parameters:**
- entity_id (required): Unique entity identifier
- label (required): Human-readable entity label
- type (required): Entity type (Concept, Person, Tool, File, Decision, Strategy, Problem)
- content (required): Content for semantic search embedding
- description (optional): Detailed description
- metadata (optional): Additional metadata as JSON object

**Returns:** entity_id, status, cached, created_at

**Example:** {"entity_id": "optimization-1", "label": "Database Optimization", "type": "Concept", "content": "Techniques for optimizing database query performance"}`,
	}, kh.HandleStoreEntity)

	// Tool 65: search-knowledge-graph
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "search-knowledge-graph",
		Description: `Search for entities using semantic similarity or graph traversal.

**Parameters:**
- query (required): Search query
- search_type (required): "semantic" or "hybrid"
- limit (optional): Max results (default: 10)
- max_hops (optional): For hybrid search, max graph hops (default: 1)
- min_similarity (optional): Minimum similarity threshold 0.0-1.0 (default: 0.7)

**Returns:** results array, result_count, search_type, latency_ms

**Search Types:**
- semantic: Vector similarity search only (fast, finds semantically similar entities)
- hybrid: Semantic search + graph traversal (finds similar + connected entities)

**Example:** {"query": "database performance", "search_type": "semantic", "limit": 5, "min_similarity": 0.7}`,
	}, kh.HandleSearchKnowledgeGraph)

	// Tool 66: create-relationship
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "create-relationship",
		Description: `Create a typed relationship between entities in the knowledge graph.

**Parameters:**
- relationship_id (required): Unique relationship identifier
- from_id (required): Source entity ID
- to_id (required): Target entity ID
- type (required): Relationship type (CAUSES, ENABLES, CONTRADICTS, BUILDS_UPON, RELATES_TO)
- strength (required): Relationship strength 0.0-1.0
- confidence (required): Confidence in relationship 0.0-1.0
- source (optional): Source of relationship (e.g., "trajectory_extraction", "manual")
- metadata (optional): Additional metadata as JSON object

**Returns:** relationship_id, status, created_at

**Example:** {"relationship_id": "rel-1", "from_id": "optimization-1", "to_id": "performance-1", "type": "ENABLES", "strength": 0.9, "confidence": 0.95}`,
	}, kh.HandleCreateRelationship)
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
func (kh *KnowledgeHandlers) HandleStoreEntity(ctx context.Context, req *mcp.CallToolRequest, request StoreEntityRequest) (*mcp.CallToolResult, *StoreEntityResponse, error) {
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	entity := &knowledge.Entity{
		ID:          request.EntityID,
		Label:       request.Label,
		Type:        knowledge.EntityType(request.Type),
		Description: request.Description,
		Metadata:    request.Metadata,
	}

	if err := kh.kg.StoreEntity(opCtx, entity, request.Content); err != nil {
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
	Query         string  `json:"query"`
	SearchType    string  `json:"search_type"` // "semantic", "graph", "hybrid"
	Limit         int     `json:"limit,omitempty"`
	MaxHops       int     `json:"max_hops,omitempty"`
	MinSimilarity float32 `json:"min_similarity,omitempty"`
}

// SearchKnowledgeGraphResponse represents the response
type SearchKnowledgeGraphResponse struct {
	Results     []*knowledge.Entity `json:"results"`
	ResultCount int                 `json:"result_count"`
	SearchType  string              `json:"search_type"`
	LatencyMs   int64               `json:"latency_ms"`
}

// HandleSearchKnowledgeGraph implements the search-knowledge-graph MCP tool
func (kh *KnowledgeHandlers) HandleSearchKnowledgeGraph(ctx context.Context, req *mcp.CallToolRequest, request SearchKnowledgeGraphRequest) (*mcp.CallToolResult, *SearchKnowledgeGraphResponse, error) {
	start := time.Now()
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Initialize as empty slice, not nil - MCP schema requires array, not null
	results := make([]*knowledge.Entity, 0)

	switch request.SearchType {
	case "semantic":
		// Semantic search only
		searchResults, searchErr := kh.kg.SearchSemantic(opCtx, request.Query, request.Limit, request.MinSimilarity)
		if searchErr != nil {
			return nil, nil, fmt.Errorf("semantic search failed: %w", searchErr)
		}

		for _, sr := range searchResults {
			entity, getErr := kh.kg.GetEntity(opCtx, sr.ID)
			if getErr == nil {
				results = append(results, entity)
			}
		}

	case "hybrid":
		// Combined semantic + graph search
		hybridResults, hybridErr := kh.kg.HybridSearch(opCtx, request.Query, request.Limit, request.MaxHops)
		if hybridErr != nil {
			return nil, nil, fmt.Errorf("hybrid search failed: %w", hybridErr)
		}
		// Ensure we never return nil - MCP schema requires array
		if hybridResults != nil {
			results = hybridResults
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
func (kh *KnowledgeHandlers) HandleCreateRelationship(ctx context.Context, req *mcp.CallToolRequest, request CreateRelationshipRequest) (*mcp.CallToolResult, *CreateRelationshipResponse, error) {
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
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

	if err := kh.kg.CreateRelationship(opCtx, rel); err != nil {
		return nil, nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	response := &CreateRelationshipResponse{
		RelationshipID: rel.ID,
		Status:         "created",
		CreatedAt:      rel.CreatedAt,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}
