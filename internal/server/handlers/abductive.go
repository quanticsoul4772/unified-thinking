package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
)

// AbductiveHandler handles abductive reasoning operations
type AbductiveHandler struct {
	reasoner *reasoning.AbductiveReasoner
	storage  storage.Storage
}

// NewAbductiveHandler creates a new abductive handler
func NewAbductiveHandler(reasoner *reasoning.AbductiveReasoner, store storage.Storage) *AbductiveHandler {
	return &AbductiveHandler{
		reasoner: reasoner,
		storage:  store,
	}
}

// GenerateHypothesesRequest represents a hypothesis generation request
type GenerateHypothesesRequest struct {
	Observations  []*ObservationInput `json:"observations"`
	MaxHypotheses int                 `json:"max_hypotheses,omitempty"`
	MinParsimony  float64             `json:"min_parsimony,omitempty"`
}

// ObservationInput represents an observation
type ObservationInput struct {
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence,omitempty"`
}

// GenerateHypothesesResponse represents the response
type GenerateHypothesesResponse struct {
	Hypotheses []*HypothesisOutput `json:"hypotheses"`
	Count      int                 `json:"count"`
}

// HypothesisOutput represents a hypothesis
type HypothesisOutput struct {
	ID               string   `json:"id"`
	Description      string   `json:"description"`
	Observations     []string `json:"observations"`
	Parsimony        float64  `json:"parsimony"`
	PriorProbability float64  `json:"prior_probability"`
	Assumptions      []string `json:"assumptions,omitempty"`
}

// HandleGenerateHypotheses generates hypotheses from observations
func (h *AbductiveHandler) HandleGenerateHypotheses(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	var req GenerateHypothesesRequest
	if err := unmarshalParams(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if len(req.Observations) == 0 {
		return nil, fmt.Errorf("observations are required")
	}

	// Convert observations
	observations := make([]*reasoning.Observation, len(req.Observations))
	for i, obs := range req.Observations {
		confidence := obs.Confidence
		if confidence == 0 {
			confidence = 0.8 // Default
		}
		observations[i] = &reasoning.Observation{
			ID:          generateID("obs"),
			Description: obs.Description,
			Confidence:  confidence,
		}
	}

	// Build request
	genReq := &reasoning.GenerateHypothesesRequest{
		Observations:  observations,
		MaxHypotheses: req.MaxHypotheses,
		MinParsimony:  req.MinParsimony,
	}

	if genReq.MaxHypotheses == 0 {
		genReq.MaxHypotheses = 10
	}

	// Generate hypotheses
	hypotheses, err := h.reasoner.GenerateHypotheses(ctx, genReq)
	if err != nil {
		return nil, fmt.Errorf("hypothesis generation failed: %w", err)
	}

	// Build response
	outputs := make([]*HypothesisOutput, len(hypotheses))
	for i, hyp := range hypotheses {
		outputs[i] = &HypothesisOutput{
			ID:               hyp.ID,
			Description:      hyp.Description,
			Observations:     hyp.Observations,
			Parsimony:        hyp.Parsimony,
			PriorProbability: hyp.PriorProbability,
			Assumptions:      hyp.Assumptions,
		}
	}

	resp := &GenerateHypothesesResponse{
		Hypotheses: outputs,
		Count:      len(outputs),
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// EvaluateHypothesesRequest represents an evaluation request
type EvaluateHypothesesRequest struct {
	Observations []*ObservationInput `json:"observations"`
	Hypotheses   []*HypothesisInput  `json:"hypotheses"`
	Method       string              `json:"method,omitempty"` // "bayesian", "parsimony", "combined"
}

// HypothesisInput represents a hypothesis for evaluation
type HypothesisInput struct {
	Description      string   `json:"description"`
	Observations     []string `json:"observations"`
	PriorProbability float64  `json:"prior_probability,omitempty"`
	Assumptions      []string `json:"assumptions,omitempty"`
}

// EvaluateHypothesesResponse represents the response
type EvaluateHypothesesResponse struct {
	RankedHypotheses []*RankedHypothesis `json:"ranked_hypotheses"`
	BestHypothesis   *RankedHypothesis   `json:"best_hypothesis"`
	Method           string              `json:"method"`
}

// RankedHypothesis represents a ranked hypothesis
type RankedHypothesis struct {
	Description          string  `json:"description"`
	PosteriorProbability float64 `json:"posterior_probability"`
	ExplanatoryPower     float64 `json:"explanatory_power"`
	Parsimony            float64 `json:"parsimony"`
	Rank                 int     `json:"rank"`
}

// HandleEvaluateHypotheses evaluates and ranks hypotheses
func (h *AbductiveHandler) HandleEvaluateHypotheses(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	var req EvaluateHypothesesRequest
	if err := unmarshalParams(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if len(req.Observations) == 0 {
		return nil, fmt.Errorf("observations are required")
	}
	if len(req.Hypotheses) == 0 {
		return nil, fmt.Errorf("hypotheses are required")
	}

	// Convert observations
	observations := make([]*reasoning.Observation, len(req.Observations))
	for i, obs := range req.Observations {
		confidence := obs.Confidence
		if confidence == 0 {
			confidence = 0.8
		}
		observations[i] = &reasoning.Observation{
			ID:          generateID("obs"),
			Description: obs.Description,
			Confidence:  confidence,
		}
	}

	// Convert hypotheses
	hypotheses := make([]*reasoning.Hypothesis, len(req.Hypotheses))
	for i, hyp := range req.Hypotheses {
		prior := hyp.PriorProbability
		if prior == 0 {
			prior = 0.5
		}
		hypotheses[i] = &reasoning.Hypothesis{
			ID:               generateID("hyp"),
			Description:      hyp.Description,
			Observations:     hyp.Observations,
			PriorProbability: prior,
			Assumptions:      hyp.Assumptions,
		}
	}

	// Build request
	method := reasoning.MethodCombined
	if req.Method != "" {
		method = reasoning.EvaluationMethod(req.Method)
	}

	evalReq := &reasoning.EvaluateHypothesesRequest{
		Observations: observations,
		Hypotheses:   hypotheses,
		Method:       method,
		Weights:      reasoning.DefaultEvaluationWeights(),
	}

	// Evaluate
	ranked, err := h.reasoner.EvaluateHypotheses(ctx, evalReq)
	if err != nil {
		return nil, fmt.Errorf("hypothesis evaluation failed: %w", err)
	}

	// Build response
	outputs := make([]*RankedHypothesis, len(ranked))
	for i, hyp := range ranked {
		outputs[i] = &RankedHypothesis{
			Description:          hyp.Description,
			PosteriorProbability: hyp.PosteriorProbability,
			ExplanatoryPower:     hyp.ExplanatoryPower,
			Parsimony:            hyp.Parsimony,
			Rank:                 i + 1,
		}
	}

	resp := &EvaluateHypothesesResponse{
		RankedHypotheses: outputs,
		Method:           string(method),
	}

	if len(outputs) > 0 {
		resp.BestHypothesis = outputs[0]
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}
