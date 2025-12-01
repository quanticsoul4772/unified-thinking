// Package handlers - Graph-of-Thoughts MCP tool handlers
package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/streaming"
)

// GoTHandler handles Graph-of-Thoughts operations
type GoTHandler struct {
	controller *modes.GraphController
	llm        modes.LLMClient
}

// NewGoTHandler creates a new GoT handler
func NewGoTHandler(controller *modes.GraphController, llm modes.LLMClient) *GoTHandler {
	return &GoTHandler{
		controller: controller,
		llm:        llm,
	}
}

// InitializeRequest for got-initialize
type InitializeRequest struct {
	GraphID        string              `json:"graph_id"`
	InitialThought string              `json:"initial_thought"`
	Config         *modes.GraphConfig `json:"config,omitempty"`
}

// InitializeResponse for got-initialize
type InitializeResponse struct {
	GraphID   string              `json:"graph_id"`
	RootID    string              `json:"root_id"`
	Status    string              `json:"status"`
	Config    *modes.GraphConfig `json:"config"`
}

// HandleInitialize creates a new GoT graph
func (h *GoTHandler) HandleInitialize(ctx context.Context, req *mcp.CallToolRequest, request InitializeRequest) (*mcp.CallToolResult, *InitializeResponse, error) {
	if request.GraphID == "" {
		return nil, nil, fmt.Errorf("graph_id is required")
	}
	if request.InitialThought == "" {
		return nil, nil, fmt.Errorf("initial_thought is required")
	}

	state, err := h.controller.Initialize(request.GraphID, request.InitialThought, request.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("initialization failed: %w", err)
	}

	response := &InitializeResponse{
		GraphID: state.ID,
		RootID:  state.RootIDs[0],
		Status:  "initialized",
		Config:  state.Config,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// GenerateRequest for got-generate
type GenerateRequest struct {
	GraphID   string   `json:"graph_id"`
	K         int      `json:"k"`
	SourceIDs []string `json:"source_ids,omitempty"`
	Problem   string   `json:"problem"`
}

// GenerateResponse for got-generate
type GenerateResponse struct {
	GraphID       string                 `json:"graph_id"`
	NewVertices   []VertexInfo          `json:"new_vertices"`
	Count         int                    `json:"count"`
	ActiveCount   int                    `json:"active_count"`
}

// VertexInfo represents vertex metadata
type VertexInfo struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Type       string  `json:"type"`
	Depth      int     `json:"depth"`
	Confidence float64 `json:"confidence"`
	Score      float64 `json:"score,omitempty"`
}

// HandleGenerate creates k diverse continuations
func (h *GoTHandler) HandleGenerate(ctx context.Context, req *mcp.CallToolRequest, request GenerateRequest) (*mcp.CallToolResult, *GenerateResponse, error) {
	if request.GraphID == "" {
		return nil, nil, fmt.Errorf("graph_id is required")
	}
	if request.K <= 0 {
		return nil, nil, fmt.Errorf("k must be positive")
	}

	// Create progress reporter for streaming notifications
	reporter := streaming.CreateReporter(req, "got-generate")

	// Report start
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(0, request.K, "initialize", fmt.Sprintf("Generating %d continuations for graph %s", request.K, request.GraphID))
	}

	genReq := modes.GenerateRequest{
		SourceVertexIDs: request.SourceIDs,
		K:               request.K,
		Problem:         request.Problem,
		MaxDepth:        0, // Use config default
	}

	// Inject reporter into context for the controller to use
	ctx = streaming.WithReporter(ctx, reporter)

	vertices, err := h.controller.Generate(ctx, request.GraphID, h.llm, genReq)
	if err != nil {
		return nil, nil, fmt.Errorf("generation failed: %w", err)
	}

	state, _ := h.controller.GetState(request.GraphID)

	newVertices := make([]VertexInfo, len(vertices))
	for i, v := range vertices {
		newVertices[i] = VertexInfo{
			ID:         v.ID,
			Content:    v.Content,
			Type:       string(v.Type),
			Depth:      v.Depth,
			Confidence: v.Confidence,
			Score:      v.Score,
		}

		// Report each vertex generation
		if reporter.IsEnabled() {
			_ = reporter.ReportStep(i+1, len(vertices), "generate", fmt.Sprintf("Generated vertex %s", v.ID))
			_ = reporter.ReportPartialResult("vertex", newVertices[i])
		}
	}

	// Report completion
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(len(vertices), len(vertices), "complete", fmt.Sprintf("Generated %d vertices", len(vertices)))
	}

	response := &GenerateResponse{
		GraphID:     request.GraphID,
		NewVertices: newVertices,
		Count:       len(newVertices),
		ActiveCount: len(state.ActiveIDs),
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// AggregateRequest for got-aggregate
type AggregateRequest struct {
	GraphID   string   `json:"graph_id"`
	VertexIDs []string `json:"vertex_ids"`
	Problem   string   `json:"problem"`
}

// AggregateResponse for got-aggregate
type AggregateResponse struct {
	GraphID          string     `json:"graph_id"`
	AggregatedVertex VertexInfo `json:"aggregated_vertex"`
	SourceCount      int        `json:"source_count"`
}

// HandleAggregate merges multiple thoughts
func (h *GoTHandler) HandleAggregate(ctx context.Context, req *mcp.CallToolRequest, request AggregateRequest) (*mcp.CallToolResult, *AggregateResponse, error) {
	if request.GraphID == "" {
		return nil, nil, fmt.Errorf("graph_id is required")
	}
	if len(request.VertexIDs) < 2 {
		return nil, nil, fmt.Errorf("need at least 2 vertices to aggregate")
	}

	// Create progress reporter for streaming notifications
	reporter := streaming.CreateReporter(req, "got-aggregate")
	sourceCount := len(request.VertexIDs)

	// Report start
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(0, sourceCount+1, "initialize", fmt.Sprintf("Aggregating %d vertices", sourceCount))
	}

	// Report collecting phase
	if reporter.IsEnabled() {
		for i, vid := range request.VertexIDs {
			_ = reporter.ReportStep(i+1, sourceCount+1, "collect", fmt.Sprintf("Collecting vertex %s", vid))
		}
	}

	aggReq := modes.AggregateRequest{
		VertexIDs: request.VertexIDs,
		Problem:   request.Problem,
	}

	// Report merge phase
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(sourceCount, sourceCount+1, "merge", "Merging vertices into unified insight")
	}

	vertex, err := h.controller.Aggregate(ctx, request.GraphID, h.llm, aggReq)
	if err != nil {
		return nil, nil, fmt.Errorf("aggregation failed: %w", err)
	}

	// Report completion
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(sourceCount+1, sourceCount+1, "complete", fmt.Sprintf("Created aggregated vertex %s", vertex.ID))
	}

	response := &AggregateResponse{
		GraphID: request.GraphID,
		AggregatedVertex: VertexInfo{
			ID:         vertex.ID,
			Content:    vertex.Content,
			Type:       string(vertex.Type),
			Depth:      vertex.Depth,
			Confidence: vertex.Confidence,
			Score:      vertex.Score,
		},
		SourceCount: len(request.VertexIDs),
	}

	// Report final result as partial data
	if reporter.IsEnabled() {
		_ = reporter.ReportPartialResult("aggregated", response.AggregatedVertex)
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// RefineRequest for got-refine
type RefineRequest struct {
	GraphID  string `json:"graph_id"`
	VertexID string `json:"vertex_id"`
	Problem  string `json:"problem"`
}

// RefineResponse for got-refine
type RefineResponse struct {
	GraphID        string     `json:"graph_id"`
	RefinedVertex  VertexInfo `json:"refined_vertex"`
	RefinementCount int       `json:"refinement_count"`
}

// HandleRefine iteratively improves a thought
func (h *GoTHandler) HandleRefine(ctx context.Context, req *mcp.CallToolRequest, request RefineRequest) (*mcp.CallToolResult, *RefineResponse, error) {
	if request.GraphID == "" {
		return nil, nil, fmt.Errorf("graph_id is required")
	}
	if request.VertexID == "" {
		return nil, nil, fmt.Errorf("vertex_id is required")
	}

	refReq := modes.RefineRequest{
		VertexID: request.VertexID,
		Problem:  request.Problem,
	}

	vertex, err := h.controller.Refine(ctx, request.GraphID, h.llm, refReq)
	if err != nil {
		return nil, nil, fmt.Errorf("refinement failed: %w", err)
	}

	response := &RefineResponse{
		GraphID: request.GraphID,
		RefinedVertex: VertexInfo{
			ID:         vertex.ID,
			Content:    vertex.Content,
			Type:       string(vertex.Type),
			Depth:      vertex.Depth,
			Confidence: vertex.Confidence,
			Score:      vertex.Score,
		},
		RefinementCount: vertex.RefinedCount,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// ScoreRequest for got-score
type ScoreRequest struct {
	GraphID  string `json:"graph_id"`
	VertexID string `json:"vertex_id"`
	Problem  string `json:"problem"`
}

// ScoreResponse for got-score
type ScoreResponse struct {
	GraphID   string  `json:"graph_id"`
	VertexID  string  `json:"vertex_id"`
	Breakdown modes.ScoreBreakdown `json:"breakdown"`
}

// HandleScore evaluates thought quality
func (h *GoTHandler) HandleScore(ctx context.Context, req *mcp.CallToolRequest, request ScoreRequest) (*mcp.CallToolResult, *ScoreResponse, error) {
	if request.GraphID == "" {
		return nil, nil, fmt.Errorf("graph_id is required")
	}
	if request.VertexID == "" {
		return nil, nil, fmt.Errorf("vertex_id is required")
	}

	scoreReq := modes.ScoreRequest{
		VertexID: request.VertexID,
		Problem:  request.Problem,
	}

	breakdown, err := h.controller.Score(ctx, request.GraphID, h.llm, scoreReq)
	if err != nil {
		return nil, nil, fmt.Errorf("scoring failed: %w", err)
	}

	response := &ScoreResponse{
		GraphID:   request.GraphID,
		VertexID:  request.VertexID,
		Breakdown: *breakdown,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// PruneRequest for got-prune
type PruneRequest struct {
	GraphID   string  `json:"graph_id"`
	Threshold float64 `json:"threshold,omitempty"`
}

// PruneResponse for got-prune
type PruneResponse struct {
	GraphID       string `json:"graph_id"`
	RemovedCount  int    `json:"removed_count"`
	RemainingCount int   `json:"remaining_count"`
	Threshold     float64 `json:"threshold"`
}

// HandlePrune removes low-quality vertices
func (h *GoTHandler) HandlePrune(ctx context.Context, req *mcp.CallToolRequest, request PruneRequest) (*mcp.CallToolResult, *PruneResponse, error) {
	if request.GraphID == "" {
		return nil, nil, fmt.Errorf("graph_id is required")
	}

	removed, err := h.controller.Prune(ctx, request.GraphID, request.Threshold)
	if err != nil {
		return nil, nil, fmt.Errorf("pruning failed: %w", err)
	}

	state, _ := h.controller.GetState(request.GraphID)

	response := &PruneResponse{
		GraphID:        request.GraphID,
		RemovedCount:   removed,
		RemainingCount: len(state.Vertices),
		Threshold:      request.Threshold,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// GetStateRequest for got-get-state
type GetStateRequest struct {
	GraphID string `json:"graph_id"`
}

// GetStateResponse for got-get-state
type GetStateResponse struct {
	GraphID       string       `json:"graph_id"`
	VertexCount   int          `json:"vertex_count"`
	EdgeCount     int          `json:"edge_count"`
	RootIDs       []string     `json:"root_ids"`
	ActiveIDs     []string     `json:"active_ids"`
	TerminalIDs   []string     `json:"terminal_ids"`
	Vertices      []VertexInfo `json:"vertices"`
	Config        *modes.GraphConfig `json:"config"`
}

// HandleGetState retrieves current graph state
func (h *GoTHandler) HandleGetState(ctx context.Context, req *mcp.CallToolRequest, request GetStateRequest) (*mcp.CallToolResult, *GetStateResponse, error) {
	if request.GraphID == "" {
		return nil, nil, fmt.Errorf("graph_id is required")
	}

	state, err := h.controller.GetState(request.GraphID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get state: %w", err)
	}

	vertices := make([]VertexInfo, 0, len(state.Vertices))
	for _, v := range state.Vertices {
		vertices = append(vertices, VertexInfo{
			ID:         v.ID,
			Content:    v.Content,
			Type:       string(v.Type),
			Depth:      v.Depth,
			Confidence: v.Confidence,
			Score:      v.Score,
		})
	}

	response := &GetStateResponse{
		GraphID:     state.ID,
		VertexCount: len(state.Vertices),
		EdgeCount:   len(state.Edges),
		RootIDs:     state.RootIDs,
		ActiveIDs:   state.ActiveIDs,
		TerminalIDs: state.TerminalIDs,
		Vertices:    vertices,
		Config:      state.Config,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// FinalizeRequest for got-finalize
type FinalizeRequest struct {
	GraphID     string   `json:"graph_id"`
	TerminalIDs []string `json:"terminal_ids"`
}

// FinalizeResponse for got-finalize
type FinalizeResponse struct {
	GraphID     string       `json:"graph_id"`
	TerminalIDs []string     `json:"terminal_ids"`
	Conclusions []VertexInfo `json:"conclusions"`
}

// HandleFinalize marks terminals and returns conclusions
func (h *GoTHandler) HandleFinalize(ctx context.Context, req *mcp.CallToolRequest, request FinalizeRequest) (*mcp.CallToolResult, *FinalizeResponse, error) {
	if request.GraphID == "" {
		return nil, nil, fmt.Errorf("graph_id is required")
	}
	if len(request.TerminalIDs) == 0 {
		return nil, nil, fmt.Errorf("terminal_ids cannot be empty")
	}

	if err := h.controller.SetTerminalVertices(request.GraphID, request.TerminalIDs); err != nil {
		return nil, nil, fmt.Errorf("failed to set terminals: %w", err)
	}

	state, _ := h.controller.GetState(request.GraphID)

	conclusions := make([]VertexInfo, 0, len(request.TerminalIDs))
	for _, termID := range request.TerminalIDs {
		if v, ok := state.Vertices[termID]; ok {
			conclusions = append(conclusions, VertexInfo{
				ID:         v.ID,
				Content:    v.Content,
				Type:       string(v.Type),
				Depth:      v.Depth,
				Confidence: v.Confidence,
				Score:      v.Score,
			})
		}
	}

	response := &FinalizeResponse{
		GraphID:     request.GraphID,
		TerminalIDs: request.TerminalIDs,
		Conclusions: conclusions,
	}

	return &mcp.CallToolResult{Content: toJSONContent(response)}, response, nil
}

// RegisterGoTTools registers all GoT MCP tools
func RegisterGoTTools(mcpServer *mcp.Server, handler *GoTHandler) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "got-initialize",
		Description: `Initialize a new Graph-of-Thoughts graph with an initial thought.

**Parameters:**
- graph_id (required): Unique identifier for this graph
- initial_thought (required): Starting thought content
- config (optional): GraphConfig with limits

**Returns:** graph_id, root_id, status, config

**Example:** {"graph_id": "sorting-problem", "initial_thought": "Sort [3,1,2] using comparisons"}`,
	}, handler.HandleInitialize)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "got-generate",
		Description: `Generate k diverse continuations from active or specified vertices.

**Parameters:**
- graph_id (required): Graph identifier
- k (required): Number of continuations per source (1-10)
- source_ids (optional): Specific vertices to expand from (default: active)
- problem (required): Original problem context

**Returns:** new_vertices array, count, active_count

**Example:** {"graph_id": "sorting-problem", "k": 3, "problem": "Sort the array"}`,
	}, handler.HandleGenerate)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "got-aggregate",
		Description: `Merge multiple parallel reasoning paths into a unified insight.

**Parameters:**
- graph_id (required): Graph identifier
- vertex_ids (required): Array of vertices to merge (min: 2)
- problem (required): Original problem context

**Returns:** aggregated_vertex, source_count

**Example:** {"graph_id": "sorting-problem", "vertex_ids": ["v1", "v2", "v3"], "problem": "Sort"}`,
	}, handler.HandleAggregate)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "got-refine",
		Description: `Iteratively improve a thought through self-critique.

**Parameters:**
- graph_id (required): Graph identifier
- vertex_id (required): Vertex to refine
- problem (required): Original problem context

**Returns:** refined_vertex, refinement_count

**Example:** {"graph_id": "sorting-problem", "vertex_id": "v1", "problem": "Sort"}`,
	}, handler.HandleRefine)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "got-score",
		Description: `Evaluate thought quality with multi-criteria breakdown.

**Parameters:**
- graph_id (required): Graph identifier
- vertex_id (required): Vertex to score
- problem (required): Original problem context

**Returns:** breakdown (confidence, validity, relevance, novelty, depth_factor, overall)

**Example:** {"graph_id": "sorting-problem", "vertex_id": "v1", "problem": "Sort"}`,
	}, handler.HandleScore)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "got-prune",
		Description: `Remove low-quality vertices below threshold (preserves roots and terminals).

**Parameters:**
- graph_id (required): Graph identifier
- threshold (optional): Minimum score to keep (default: config.PruneThreshold)

**Returns:** removed_count, remaining_count, threshold

**Example:** {"graph_id": "sorting-problem", "threshold": 0.3}`,
	}, handler.HandlePrune)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "got-get-state",
		Description: `Get current graph state with all vertices and metadata.

**Parameters:**
- graph_id (required): Graph identifier

**Returns:** vertex_count, edge_count, root_ids, active_ids, terminal_ids, vertices, config

**Example:** {"graph_id": "sorting-problem"}`,
	}, handler.HandleGetState)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "got-finalize",
		Description: `Mark terminal vertices and retrieve final conclusions.

**Parameters:**
- graph_id (required): Graph identifier
- terminal_ids (required): Array of final conclusion vertex IDs

**Returns:** terminal_ids, conclusions array

**Example:** {"graph_id": "sorting-problem", "terminal_ids": ["v10", "v15"]}`,
	}, handler.HandleFinalize)
}
