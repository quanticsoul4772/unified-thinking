// Package handlers - Research with Web Search MCP tool handler
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
)

// ResearchHandler handles web search augmented research operations
type ResearchHandler struct {
	llm modes.LLMClient
}

// NewResearchHandler creates a new research handler
func NewResearchHandler(llm modes.LLMClient) *ResearchHandler {
	return &ResearchHandler{
		llm: llm,
	}
}

// ResearchWithSearchRequest for research-with-search tool
type ResearchWithSearchRequest struct {
	Query   string `json:"query"`
	Problem string `json:"problem,omitempty"`
}

// ResearchWithSearchResponse for research-with-search tool
type ResearchWithSearchResponse struct {
	Findings    string         `json:"findings"`
	KeyInsights []string       `json:"key_insights"`
	Confidence  float64        `json:"confidence"`
	Citations   []CitationInfo `json:"citations"`
	Searches    int            `json:"searches_performed"`
	Status      string         `json:"status"`
}

// CitationInfo represents a citation from web search
type CitationInfo struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// HandleResearchWithSearch performs web-augmented research
func (h *ResearchHandler) HandleResearchWithSearch(ctx context.Context, req *mcp.CallToolRequest, request ResearchWithSearchRequest) (*mcp.CallToolResult, *ResearchWithSearchResponse, error) {
	if request.Query == "" {
		return nil, nil, fmt.Errorf("query is required")
	}

	// Check if LLM client is available
	if h.llm == nil {
		return nil, nil, fmt.Errorf("research-with-search requires ANTHROPIC_API_KEY with WEB_SEARCH_ENABLED=true")
	}

	// Use problem context if provided
	problem := request.Problem
	if problem == "" {
		problem = request.Query
	}

	// Perform research with web search
	result, err := h.llm.ResearchWithSearch(ctx, request.Query, problem)
	if err != nil {
		return nil, nil, fmt.Errorf("research failed: %w", err)
	}

	// Convert citations
	citations := make([]CitationInfo, len(result.Citations))
	for i, c := range result.Citations {
		citations[i] = CitationInfo{
			URL:   c.URL,
			Title: c.Title,
		}
	}

	response := &ResearchWithSearchResponse{
		Findings:    result.Findings,
		KeyInsights: result.KeyInsights,
		Confidence:  result.Confidence,
		Citations:   citations,
		Searches:    result.Searches,
		Status:      "completed",
	}

	return &mcp.CallToolResult{Content: researchToJSONContent(response)}, response, nil
}

// researchToJSONContent converts response to JSON content
func researchToJSONContent(data interface{}) []mcp.Content {
	jsonData, err := json.Marshal(data)
	if err != nil {
		errData := map[string]string{"error": err.Error()}
		jsonData, _ = json.Marshal(errData)
	}

	return []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}
}

// RegisterResearchTools registers all research MCP tools
func RegisterResearchTools(mcpServer *mcp.Server, handler *ResearchHandler) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "research-with-search",
		Description: `Perform web-augmented research using Anthropic's built-in web search.

Requires ANTHROPIC_API_KEY and WEB_SEARCH_ENABLED=true environment variables.

**Parameters:**
- query (required): The research query or question
- problem (optional): Context about the problem being researched

**Returns:**
- findings: Research findings and analysis
- key_insights: Array of key insights extracted
- confidence: Confidence score (0.0-1.0) based on source quality
- citations: Array of source URLs and titles
- searches_performed: Number of web searches executed

**Example:** {"query": "What are the latest advances in quantum computing?", "problem": "Research quantum computing for a technology report"}

**Note:** This tool uses Anthropic's server-side web search, which automatically
searches the web and integrates results into the analysis. The tool may perform
multiple searches to gather comprehensive information.`,
	}, handler.HandleResearchWithSearch)
}
