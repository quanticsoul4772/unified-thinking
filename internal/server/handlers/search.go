package handlers

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// SearchHandler handles search operations
type SearchHandler struct {
	storage storage.Storage
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(store storage.Storage) *SearchHandler {
	return &SearchHandler{
		storage: store,
	}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query  string `json:"query"`
	Mode   string `json:"mode,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Thoughts []*types.Thought `json:"thoughts"`
	Count    int              `json:"count"`
	Query    string           `json:"query"`
}

// MetricsResponse represents a metrics response
type MetricsResponse struct {
	TotalThoughts     int            `json:"total_thoughts"`
	TotalBranches     int            `json:"total_branches"`
	TotalInsights     int            `json:"total_insights"`
	TotalValidations  int            `json:"total_validations"`
	ThoughtsByMode    map[string]int `json:"thoughts_by_mode"`
	AverageConfidence float64        `json:"average_confidence"`
}

// HandleSearch searches for thoughts
func (h *SearchHandler) HandleSearch(ctx context.Context, req *mcp.CallToolRequest, input SearchRequest) (*mcp.CallToolResult, *SearchResponse, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 100
	}

	thoughts := h.storage.SearchThoughts(input.Query, types.ThinkingMode(input.Mode), limit, input.Offset)

	response := &SearchResponse{
		Thoughts: thoughts,
		Count:    len(thoughts),
		Query:    input.Query,
	}

	return &mcp.CallToolResult{}, response, nil
}

// HandleGetMetrics retrieves system metrics
func (h *SearchHandler) HandleGetMetrics(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *MetricsResponse, error) {
	metrics := h.storage.GetMetrics()

	response := &MetricsResponse{
		TotalThoughts:     metrics.TotalThoughts,
		TotalBranches:     metrics.TotalBranches,
		TotalInsights:     metrics.TotalInsights,
		TotalValidations:  metrics.TotalValidations,
		ThoughtsByMode:    metrics.ThoughtsByMode,
		AverageConfidence: metrics.AverageConfidence,
	}

	return &mcp.CallToolResult{}, response, nil
}
