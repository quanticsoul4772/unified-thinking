package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
)

// CaseBasedHandler handles case-based reasoning operations
type CaseBasedHandler struct {
	reasoner *reasoning.CaseBasedReasoner
	storage  storage.Storage
}

// NewCaseBasedHandler creates a new case-based handler
func NewCaseBasedHandler(reasoner *reasoning.CaseBasedReasoner, store storage.Storage) *CaseBasedHandler {
	return &CaseBasedHandler{
		reasoner: reasoner,
		storage:  store,
	}
}

// RetrieveCasesRequest represents a case retrieval request
type RetrieveCasesRequest struct {
	Problem       *ProblemDescriptionInput `json:"problem"`
	Domain        string                   `json:"domain,omitempty"`
	MaxCases      int                      `json:"max_cases,omitempty"`
	MinSimilarity float64                  `json:"min_similarity,omitempty"`
}

// ProblemDescriptionInput represents a problem
type ProblemDescriptionInput struct {
	Description string                 `json:"description"`
	Context     string                 `json:"context,omitempty"`
	Goals       []string               `json:"goals,omitempty"`
	Constraints []string               `json:"constraints,omitempty"`
	Features    map[string]interface{} `json:"features,omitempty"`
}

// RetrieveCasesResponse represents the response
type RetrieveCasesResponse struct {
	Cases     []*SimilarCaseOutput `json:"cases"`
	Retrieved int                  `json:"retrieved"`
}

// SimilarCaseOutput represents a similar case
type SimilarCaseOutput struct {
	CaseID      string                   `json:"case_id"`
	Problem     *ProblemDescriptionInfo  `json:"problem"`
	Solution    *SolutionDescriptionInfo `json:"solution"`
	Similarity  float64                  `json:"similarity"`
	SuccessRate float64                  `json:"success_rate"`
	Domain      string                   `json:"domain"`
}

// ProblemDescriptionInfo contains problem info
type ProblemDescriptionInfo struct {
	Description string   `json:"description"`
	Context     string   `json:"context,omitempty"`
	Goals       []string `json:"goals,omitempty"`
}

// SolutionDescriptionInfo contains solution info
type SolutionDescriptionInfo struct {
	Description string   `json:"description"`
	Approach    string   `json:"approach,omitempty"`
	Steps       []string `json:"steps,omitempty"`
}

// HandleRetrieveCases retrieves similar cases
func (h *CaseBasedHandler) HandleRetrieveCases(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	var req RetrieveCasesRequest
	if err := unmarshalParams(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request: " + err.Error())
	}

	if req.Problem == nil || req.Problem.Description == "" {
		return nil, fmt.Errorf("problem description is required")
	}

	// Convert problem
	problem := &reasoning.ProblemDescription{
		Description: req.Problem.Description,
		Context:     req.Problem.Context,
		Goals:       req.Problem.Goals,
		Constraints: req.Problem.Constraints,
		Features:    req.Problem.Features,
	}

	// Build request
	retrieveReq := &reasoning.RetrieveRequest{
		Problem:       problem,
		Domain:        req.Domain,
		MaxCases:      req.MaxCases,
		MinSimilarity: req.MinSimilarity,
	}

	if retrieveReq.MaxCases == 0 {
		retrieveReq.MaxCases = 5
	}
	if retrieveReq.MinSimilarity == 0 {
		retrieveReq.MinSimilarity = 0.3
	}

	// Retrieve cases
	result, err := h.reasoner.Retrieve(ctx, retrieveReq)
	if err != nil {
		return nil, fmt.Errorf("case retrieval failed: " + err.Error())
	}

	// Build response
	cases := make([]*SimilarCaseOutput, len(result.Cases))
	for i, sc := range result.Cases {
		cases[i] = &SimilarCaseOutput{
			CaseID: sc.Case.ID,
			Problem: &ProblemDescriptionInfo{
				Description: sc.Case.Problem.Description,
				Context:     sc.Case.Problem.Context,
				Goals:       sc.Case.Problem.Goals,
			},
			Solution: &SolutionDescriptionInfo{
				Description: sc.Case.Solution.Description,
				Approach:    sc.Case.Solution.Approach,
				Steps:       sc.Case.Solution.Steps,
			},
			Similarity:  sc.Similarity,
			SuccessRate: sc.Case.SuccessRate,
			Domain:      sc.Case.Domain,
		}
	}

	resp := &RetrieveCasesResponse{
		Cases:     cases,
		Retrieved: result.Retrieved,
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// PerformCBRCycleRequest represents a full CBR cycle request
type PerformCBRCycleRequest struct {
	Problem *ProblemDescriptionInput `json:"problem"`
	Domain  string                   `json:"domain,omitempty"`
}

// PerformCBRCycleResponse represents the response
type PerformCBRCycleResponse struct {
	Retrieved       int                      `json:"retrieved"`
	BestCase        *SimilarCaseOutput       `json:"best_case,omitempty"`
	AdaptedSolution *SolutionDescriptionInfo `json:"adapted_solution,omitempty"`
	Strategy        string                   `json:"strategy,omitempty"`
	Confidence      float64                  `json:"confidence"`
}

// HandlePerformCBRCycle performs a full CBR cycle
func (h *CaseBasedHandler) HandlePerformCBRCycle(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	var req PerformCBRCycleRequest
	if err := unmarshalParams(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request: " + err.Error())
	}

	if req.Problem == nil || req.Problem.Description == "" {
		return nil, fmt.Errorf("problem description is required")
	}

	// Convert problem
	problem := &reasoning.ProblemDescription{
		Description: req.Problem.Description,
		Context:     req.Problem.Context,
		Goals:       req.Problem.Goals,
		Constraints: req.Problem.Constraints,
		Features:    req.Problem.Features,
	}

	// Perform CBR cycle
	cycle, err := h.reasoner.PerformCBRCycle(ctx, problem, req.Domain)
	if err != nil {
		return nil, fmt.Errorf("CBR cycle failed: " + err.Error())
	}

	resp := &PerformCBRCycleResponse{
		Retrieved:  cycle.Retrieved.Retrieved,
		Confidence: 0.5,
	}

	// If we have a reused case
	if cycle.Reused != nil {
		resp.BestCase = &SimilarCaseOutput{
			CaseID: cycle.Reused.OriginalCase.ID,
			Problem: &ProblemDescriptionInfo{
				Description: cycle.Reused.OriginalCase.Problem.Description,
				Context:     cycle.Reused.OriginalCase.Problem.Context,
			},
			Solution: &SolutionDescriptionInfo{
				Description: cycle.Reused.OriginalCase.Solution.Description,
				Approach:    cycle.Reused.OriginalCase.Solution.Approach,
			},
			SuccessRate: cycle.Reused.OriginalCase.SuccessRate,
			Domain:      cycle.Reused.OriginalCase.Domain,
		}

		resp.AdaptedSolution = &SolutionDescriptionInfo{
			Description: cycle.Reused.AdaptedSolution.Description,
			Approach:    cycle.Reused.AdaptedSolution.Approach,
			Steps:       cycle.Reused.AdaptedSolution.Steps,
		}

		resp.Strategy = string(cycle.Reused.Strategy)
		resp.Confidence = cycle.Reused.Confidence
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}
