package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/streaming"
	"unified-thinking/internal/types"
)

// TemporalHandler handles temporal and perspective analysis operations
type TemporalHandler struct {
	perspectiveAnalyzer *analysis.PerspectiveAnalyzer
	temporalReasoner    *reasoning.TemporalReasoner
}

// NewTemporalHandler creates a new temporal handler
func NewTemporalHandler(
	perspectiveAnalyzer *analysis.PerspectiveAnalyzer,
	temporalReasoner *reasoning.TemporalReasoner,
) *TemporalHandler {
	return &TemporalHandler{
		perspectiveAnalyzer: perspectiveAnalyzer,
		temporalReasoner:    temporalReasoner,
	}
}

// ============================================================================
// Request/Response Types
// ============================================================================

// AnalyzePerspectivesRequest represents a perspective analysis request
type AnalyzePerspectivesRequest struct {
	Situation        string   `json:"situation"`
	StakeholderHints []string `json:"stakeholder_hints,omitempty"`
}

// AnalyzePerspectivesResponse represents a perspective analysis response
type AnalyzePerspectivesResponse struct {
	Perspectives []*types.Perspective    `json:"perspectives"`
	Count        int                     `json:"count"`
	Conflicts    []string                `json:"conflicts,omitempty"`
	Status       string                  `json:"status"`
	Metadata     *types.ResponseMetadata `json:"metadata,omitempty"`
}

// AnalyzeTemporalRequest represents a temporal analysis request
type AnalyzeTemporalRequest struct {
	Situation   string `json:"situation"`
	TimeHorizon string `json:"time_horizon,omitempty"`
}

// AnalyzeTemporalResponse represents a temporal analysis response
type AnalyzeTemporalResponse struct {
	Analysis *types.TemporalAnalysis `json:"analysis"`
	Status   string                  `json:"status"`
	Metadata *types.ResponseMetadata `json:"metadata,omitempty"`
}

// CompareTimeHorizonsRequest represents a time horizon comparison request
type CompareTimeHorizonsRequest struct {
	Situation string `json:"situation"`
}

// CompareTimeHorizonsResponse represents a time horizon comparison response
type CompareTimeHorizonsResponse struct {
	Analyses map[string]*types.TemporalAnalysis `json:"analyses"`
	Status   string                             `json:"status"`
}

// IdentifyOptimalTimingRequest represents an optimal timing identification request
type IdentifyOptimalTimingRequest struct {
	Situation   string   `json:"situation"`
	Constraints []string `json:"constraints,omitempty"`
}

// IdentifyOptimalTimingResponse represents an optimal timing identification response
type IdentifyOptimalTimingResponse struct {
	Recommendation string `json:"recommendation"`
	Status         string `json:"status"`
}

// ============================================================================
// Handler Methods
// ============================================================================

// HandleAnalyzePerspectives processes perspective analysis requests
func (h *TemporalHandler) HandleAnalyzePerspectives(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input AnalyzePerspectivesRequest,
) (*mcp.CallToolResult, *AnalyzePerspectivesResponse, error) {
	// Create progress reporter for streaming notifications
	reporter := streaming.CreateReporter(req, "analyze-perspectives")
	stakeholderCount := len(input.StakeholderHints)
	if stakeholderCount == 0 {
		stakeholderCount = 3 // Default stakeholders if none specified
	}
	totalSteps := stakeholderCount + 1 // stakeholders + conflicts analysis

	// Report start
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(1, totalSteps, "analyze", fmt.Sprintf("Analyzing %d stakeholder perspectives...", stakeholderCount))
	}

	perspectives, err := h.perspectiveAnalyzer.AnalyzePerspectives(input.Situation, input.StakeholderHints)
	if err != nil {
		return nil, nil, err
	}

	// Report completion
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(totalSteps, totalSteps, "complete", fmt.Sprintf("Identified %d perspectives", len(perspectives)))
	}

	// Generate metadata for Claude orchestration
	metadataGen := NewMetadataGenerator()
	metadata := metadataGen.GeneratePerspectiveAnalysisMetadata(perspectives)

	// Note: conflict detection is done internally, made available through ComparePerspectives if needed
	response := &AnalyzePerspectivesResponse{
		Perspectives: perspectives,
		Count:        len(perspectives),
		Status:       "success",
		Metadata:     metadata,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleAnalyzeTemporal processes temporal analysis requests
func (h *TemporalHandler) HandleAnalyzeTemporal(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input AnalyzeTemporalRequest,
) (*mcp.CallToolResult, *AnalyzeTemporalResponse, error) {
	analysis, err := h.temporalReasoner.AnalyzeTemporal(input.Situation, input.TimeHorizon)
	if err != nil {
		return nil, nil, err
	}

	// Generate metadata for Claude orchestration
	metadataGen := NewMetadataGenerator()
	metadata := metadataGen.GenerateTemporalAnalysisMetadata(analysis)

	response := &AnalyzeTemporalResponse{
		Analysis: analysis,
		Status:   "success",
		Metadata: metadata,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleCompareTimeHorizons processes time horizon comparison requests
func (h *TemporalHandler) HandleCompareTimeHorizons(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CompareTimeHorizonsRequest,
) (*mcp.CallToolResult, *CompareTimeHorizonsResponse, error) {
	analyses, err := h.temporalReasoner.CompareTimeHorizons(input.Situation)
	if err != nil {
		return nil, nil, err
	}

	response := &CompareTimeHorizonsResponse{
		Analyses: analyses,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleIdentifyOptimalTiming processes optimal timing identification requests
func (h *TemporalHandler) HandleIdentifyOptimalTiming(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input IdentifyOptimalTimingRequest,
) (*mcp.CallToolResult, *IdentifyOptimalTimingResponse, error) {
	recommendation, err := h.temporalReasoner.IdentifyOptimalTiming(input.Situation, input.Constraints)
	if err != nil {
		return nil, nil, err
	}

	response := &IdentifyOptimalTimingResponse{
		Recommendation: recommendation,
		Status:         "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}
