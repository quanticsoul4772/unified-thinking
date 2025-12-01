// Package handlers - Thought similarity search handlers
package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/similarity"
)

// SimilarityHandler handles thought similarity search operations
type SimilarityHandler struct {
	searcher *similarity.ThoughtSearcher
}

// NewSimilarityHandler creates a new similarity handler
func NewSimilarityHandler(searcher *similarity.ThoughtSearcher) *SimilarityHandler {
	return &SimilarityHandler{searcher: searcher}
}

// SearchSimilarThoughtsRequest searches for similar thoughts
type SearchSimilarThoughtsRequest struct {
	Query         string  `json:"query"`
	Limit         int     `json:"limit,omitempty"`
	MinSimilarity float64 `json:"min_similarity,omitempty"`
}

// SearchSimilarThoughtsResponse returns similar thoughts
type SearchSimilarThoughtsResponse struct {
	Results []SimilarThoughtResult `json:"results"`
	Count   int                    `json:"count"`
}

// SimilarThoughtResult represents a similar thought with metadata
type SimilarThoughtResult struct {
	ThoughtID  string                 `json:"thought_id"`
	Content    string                 `json:"content"`
	Mode       string                 `json:"mode"`
	Confidence float64                `json:"confidence"`
	Similarity float64                `json:"similarity"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// HandleSearchSimilarThoughts searches for similar thoughts
func (h *SimilarityHandler) HandleSearchSimilarThoughts(ctx context.Context, req *mcp.CallToolRequest, request SearchSimilarThoughtsRequest) (*mcp.CallToolResult, *SearchSimilarThoughtsResponse, error) {
	if request.Query == "" {
		return nil, nil, fmt.Errorf("query is required")
	}

	limit := request.Limit
	if limit == 0 {
		limit = 5
	}

	minSimilarity := float32(request.MinSimilarity)
	if minSimilarity == 0 {
		minSimilarity = 0.5
	}

	// Search for similar thoughts
	results, err := h.searcher.SearchSimilar(ctx, request.Query, limit, minSimilarity)
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert to response format
	responseResults := make([]SimilarThoughtResult, len(results))
	for i, r := range results {
		responseResults[i] = SimilarThoughtResult{
			ThoughtID:  r.Thought.ID,
			Content:    r.Thought.Content,
			Mode:       string(r.Thought.Mode),
			Confidence: r.Thought.Confidence,
			Similarity: float64(r.Similarity),
			Metadata:   r.Thought.Metadata,
		}
	}

	response := &SearchSimilarThoughtsResponse{
		Results: responseResults,
		Count:   len(responseResults),
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// RegisterSimilarityTools registers thought similarity MCP tools
func RegisterSimilarityTools(mcpServer *mcp.Server, sh *SimilarityHandler) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "search-similar-thoughts",
		Description: `Search for thoughts similar to a query using semantic embeddings.

Finds past thoughts semantically similar to your query, allowing you to reuse proven reasoning chains.

**Parameters:**
- query (required): Text to find similar thoughts
- limit (optional): Maximum results (default: 5)
- min_similarity (optional): Threshold 0-1 (default: 0.5)

**Returns:** results (array of similar thoughts with similarity scores), count

**Example:** {"query": "database optimization", "limit": 3, "min_similarity": 0.6}`,
	}, sh.HandleSearchSimilarThoughts)
}
