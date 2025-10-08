package handlers

import (
	"fmt"
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/storage"
)

// UnknownUnknownsHandler handles unknown unknowns detection operations
type UnknownUnknownsHandler struct {
	detector *metacognition.UnknownUnknownsDetector
	storage  storage.Storage
}

// NewUnknownUnknownsHandler creates a new unknown unknowns handler
func NewUnknownUnknownsHandler(detector *metacognition.UnknownUnknownsDetector, store storage.Storage) *UnknownUnknownsHandler {
	return &UnknownUnknownsHandler{
		detector: detector,
		storage:  store,
	}
}

// DetectBlindSpotsRequest represents a blind spot detection request
type DetectBlindSpotsRequest struct {
	Content     string   `json:"content"`
	Domain      string   `json:"domain,omitempty"`
	Context     string   `json:"context,omitempty"`
	Assumptions []string `json:"assumptions,omitempty"`
	Confidence  float64  `json:"confidence,omitempty"`
}

// DetectBlindSpotsResponse represents the response
type DetectBlindSpotsResponse struct {
	BlindSpots                []*BlindSpotOutput `json:"blind_spots"`
	MissingConsiderations     []string           `json:"missing_considerations"`
	UnchallengedAssumptions   []string           `json:"unchallenged_assumptions"`
	SuggestedQuestions        []string           `json:"suggested_questions"`
	OverallRisk               float64            `json:"overall_risk"`
	RiskLevel                 string             `json:"risk_level"`
	Analysis                  string             `json:"analysis"`
}

// BlindSpotOutput represents a blind spot
type BlindSpotOutput struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Severity    float64  `json:"severity"`
	Indicators  []string `json:"indicators"`
	Suggestions []string `json:"suggestions"`
}

// HandleDetectBlindSpots detects blind spots and knowledge gaps
func (h *UnknownUnknownsHandler) HandleDetectBlindSpots(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	var req DetectBlindSpotsRequest
	if err := unmarshalParams(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request: " + err.Error())
	}

	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Default confidence
	if req.Confidence == 0 {
		req.Confidence = 0.7
	}

	// Build request
	gapReq := &metacognition.GapAnalysisRequest{
		Content:     req.Content,
		Domain:      req.Domain,
		Context:     req.Context,
		Assumptions: req.Assumptions,
		Confidence:  req.Confidence,
	}

	// Detect blind spots
	result, err := h.detector.DetectBlindSpots(ctx, gapReq)
	if err != nil {
		return nil, fmt.Errorf("blind spot detection failed: " + err.Error())
	}

	// Build response
	blindSpots := make([]*BlindSpotOutput, len(result.BlindSpots))
	for i, bs := range result.BlindSpots {
		blindSpots[i] = &BlindSpotOutput{
			Type:        string(bs.Type),
			Description: bs.Description,
			Severity:    bs.Severity,
			Indicators:  bs.Indicators,
			Suggestions: bs.Suggestions,
		}
	}

	// Determine risk level
	riskLevel := "LOW"
	if result.OverallRisk > 0.7 {
		riskLevel = "HIGH"
	} else if result.OverallRisk > 0.4 {
		riskLevel = "MODERATE"
	}

	resp := &DetectBlindSpotsResponse{
		BlindSpots:              blindSpots,
		MissingConsiderations:   result.MissingConsiderations,
		UnchallengedAssumptions: result.UnchallengedAssumptions,
		SuggestedQuestions:      result.SuggestedQuestions,
		OverallRisk:             result.OverallRisk,
		RiskLevel:               riskLevel,
		Analysis:                result.Analysis,
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}
