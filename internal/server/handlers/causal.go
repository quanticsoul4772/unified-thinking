package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/streaming"
	"unified-thinking/internal/types"
)

// CausalHandler handles causal reasoning operations
type CausalHandler struct {
	causalReasoner *reasoning.CausalReasoner
}

// NewCausalHandler creates a new causal handler
func NewCausalHandler(causalReasoner *reasoning.CausalReasoner) *CausalHandler {
	return &CausalHandler{
		causalReasoner: causalReasoner,
	}
}

// ============================================================================
// Request/Response Types
// ============================================================================

// BuildCausalGraphRequest represents a causal graph building request
type BuildCausalGraphRequest struct {
	Description  string   `json:"description"`
	Observations []string `json:"observations"`
}

// BuildCausalGraphResponse represents a causal graph building response
type BuildCausalGraphResponse struct {
	Graph    *types.CausalGraph      `json:"graph"`
	Status   string                  `json:"status"`
	Metadata *types.ResponseMetadata `json:"metadata,omitempty"`
}

// SimulateInterventionRequest represents an intervention simulation request
type SimulateInterventionRequest struct {
	GraphID          string `json:"graph_id"`
	VariableID       string `json:"variable_id"`
	InterventionType string `json:"intervention_type"`
}

// SimulateInterventionResponse represents an intervention simulation response
type SimulateInterventionResponse struct {
	Intervention *types.CausalIntervention `json:"intervention"`
	Status       string                    `json:"status"`
}

// GenerateCounterfactualRequest represents a counterfactual generation request
type GenerateCounterfactualRequest struct {
	GraphID  string            `json:"graph_id"`
	Scenario string            `json:"scenario"`
	Changes  map[string]string `json:"changes"`
}

// GenerateCounterfactualResponse represents a counterfactual generation response
type GenerateCounterfactualResponse struct {
	Counterfactual *types.Counterfactual `json:"counterfactual"`
	Status         string                `json:"status"`
}

// AnalyzeCorrelationVsCausationRequest represents a correlation vs causation analysis request
type AnalyzeCorrelationVsCausationRequest struct {
	Observation string `json:"observation"`
}

// AnalyzeCorrelationVsCausationResponse represents a correlation vs causation analysis response
type AnalyzeCorrelationVsCausationResponse struct {
	Analysis string `json:"analysis"`
	Status   string `json:"status"`
}

// GetCausalGraphRequest represents a causal graph retrieval request
type GetCausalGraphRequest struct {
	GraphID string `json:"graph_id"`
}

// GetCausalGraphResponse represents a causal graph retrieval response
type GetCausalGraphResponse struct {
	Graph  *types.CausalGraph `json:"graph"`
	Status string             `json:"status"`
}

// ============================================================================
// Handler Methods
// ============================================================================

// HandleBuildCausalGraph processes causal graph building requests
func (h *CausalHandler) HandleBuildCausalGraph(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input BuildCausalGraphRequest,
) (*mcp.CallToolResult, *BuildCausalGraphResponse, error) {
	// Create progress reporter for streaming notifications
	reporter := streaming.CreateReporter(req, "build-causal-graph")
	observationCount := len(input.Observations)
	totalSteps := observationCount + 2 // parse observations + build links + finalize

	// Report start
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(1, totalSteps, "parse", fmt.Sprintf("Parsing %d observations...", observationCount))
	}

	graph, err := h.causalReasoner.BuildCausalGraph(input.Description, input.Observations)
	if err != nil {
		return nil, nil, err
	}

	// Report completion
	if reporter.IsEnabled() {
		varCount := len(graph.Variables)
		linkCount := len(graph.Links)
		_ = reporter.ReportStep(totalSteps, totalSteps, "complete", fmt.Sprintf("Built graph: %d variables, %d links", varCount, linkCount))
	}

	// Generate metadata for Claude orchestration
	metadataGen := NewMetadataGenerator()
	metadata := metadataGen.GenerateCausalGraphMetadata(graph)

	response := &BuildCausalGraphResponse{
		Graph:    graph,
		Status:   "success",
		Metadata: metadata,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleSimulateIntervention processes intervention simulation requests
func (h *CausalHandler) HandleSimulateIntervention(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SimulateInterventionRequest,
) (*mcp.CallToolResult, *SimulateInterventionResponse, error) {
	intervention, err := h.causalReasoner.SimulateIntervention(
		input.GraphID,
		input.VariableID,
		input.InterventionType,
	)
	if err != nil {
		return nil, nil, err
	}

	response := &SimulateInterventionResponse{
		Intervention: intervention,
		Status:       "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleGenerateCounterfactual processes counterfactual generation requests
func (h *CausalHandler) HandleGenerateCounterfactual(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GenerateCounterfactualRequest,
) (*mcp.CallToolResult, *GenerateCounterfactualResponse, error) {
	counterfactual, err := h.causalReasoner.GenerateCounterfactual(
		input.GraphID,
		input.Scenario,
		input.Changes,
	)
	if err != nil {
		return nil, nil, err
	}

	response := &GenerateCounterfactualResponse{
		Counterfactual: counterfactual,
		Status:         "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleAnalyzeCorrelationVsCausation processes correlation vs causation analysis requests
func (h *CausalHandler) HandleAnalyzeCorrelationVsCausation(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input AnalyzeCorrelationVsCausationRequest,
) (*mcp.CallToolResult, *AnalyzeCorrelationVsCausationResponse, error) {
	analysis, err := h.causalReasoner.AnalyzeCorrelationVsCausation(input.Observation)
	if err != nil {
		return nil, nil, err
	}

	response := &AnalyzeCorrelationVsCausationResponse{
		Analysis: analysis,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleGetCausalGraph processes causal graph retrieval requests
func (h *CausalHandler) HandleGetCausalGraph(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetCausalGraphRequest,
) (*mcp.CallToolResult, *GetCausalGraphResponse, error) {
	graph, err := h.causalReasoner.GetGraph(input.GraphID)
	if err != nil {
		return nil, nil, err
	}

	response := &GetCausalGraphResponse{
		Graph:  graph,
		Status: "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}
