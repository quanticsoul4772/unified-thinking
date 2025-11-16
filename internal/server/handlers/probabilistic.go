package handlers

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// ProbabilisticHandler handles probabilistic reasoning and evidence operations
type ProbabilisticHandler struct {
	storage               storage.Storage
	probabilisticReasoner *reasoning.ProbabilisticReasoner
	evidenceAnalyzer      *analysis.EvidenceAnalyzer
	contradictionDetector *analysis.ContradictionDetector
}

// NewProbabilisticHandler creates a new probabilistic handler
func NewProbabilisticHandler(
	store storage.Storage,
	probabilisticReasoner *reasoning.ProbabilisticReasoner,
	evidenceAnalyzer *analysis.EvidenceAnalyzer,
	contradictionDetector *analysis.ContradictionDetector,
) *ProbabilisticHandler {
	return &ProbabilisticHandler{
		storage:               store,
		probabilisticReasoner: probabilisticReasoner,
		evidenceAnalyzer:      evidenceAnalyzer,
		contradictionDetector: contradictionDetector,
	}
}

// ============================================================================
// Request/Response Types
// ============================================================================

// ProbabilisticReasoningRequest represents a probabilistic reasoning request
type ProbabilisticReasoningRequest struct {
	Operation    string   `json:"operation"`               // "create", "update", "get", or "combine"
	Statement    string   `json:"statement,omitempty"`     // For create operation
	PriorProb    float64  `json:"prior_prob,omitempty"`    // For create operation
	BeliefID     string   `json:"belief_id,omitempty"`     // For update/get operations
	EvidenceID   string   `json:"evidence_id,omitempty"`   // For update operation
	Likelihood   float64  `json:"likelihood,omitempty"`    // For update operation
	EvidenceProb float64  `json:"evidence_prob,omitempty"` // For update operation
	BeliefIDs    []string `json:"belief_ids,omitempty"`    // For combine operation
	CombineOp    string   `json:"combine_op,omitempty"`    // "and" or "or" for combine
}

// ProbabilisticReasoningResponse represents a probabilistic reasoning response
type ProbabilisticReasoningResponse struct {
	Belief       *types.ProbabilisticBelief `json:"belief,omitempty"`
	CombinedProb float64                    `json:"combined_prob,omitempty"`
	Operation    string                     `json:"operation"`
	Status       string                     `json:"status"`
}

// AssessEvidenceRequest represents an evidence assessment request
type AssessEvidenceRequest struct {
	Content       string `json:"content"`
	Source        string `json:"source"`
	ClaimID       string `json:"claim_id,omitempty"`
	SupportsClaim bool   `json:"supports_claim"`
}

// AssessEvidenceResponse represents an evidence assessment response
type AssessEvidenceResponse struct {
	Evidence *types.Evidence `json:"evidence"`
	Status   string          `json:"status"`
}

// DetectContradictionsRequest represents a contradiction detection request
type DetectContradictionsRequest struct {
	ThoughtIDs []string `json:"thought_ids,omitempty"` // Specific thought IDs to check
	BranchID   string   `json:"branch_id,omitempty"`   // Or check all thoughts in a branch
	Mode       string   `json:"mode,omitempty"`        // Or check all thoughts in a mode
}

// DetectContradictionsResponse represents a contradiction detection response
type DetectContradictionsResponse struct {
	Contradictions []*types.Contradiction `json:"contradictions"`
	Count          int                    `json:"count"`
	Status         string                 `json:"status"`
}

// ============================================================================
// Handler Methods
// ============================================================================

// HandleProbabilisticReasoning processes probabilistic reasoning requests
func (h *ProbabilisticHandler) HandleProbabilisticReasoning(ctx context.Context, req *mcp.CallToolRequest, input ProbabilisticReasoningRequest) (*mcp.CallToolResult, *ProbabilisticReasoningResponse, error) {
	if err := ValidateProbabilisticReasoningRequest(&input); err != nil {
		return nil, nil, err
	}

	response := &ProbabilisticReasoningResponse{
		Operation: input.Operation,
		Status:    "success",
	}

	switch input.Operation {
	case "create":
		belief, err := h.probabilisticReasoner.CreateBelief(input.Statement, input.PriorProb)
		if err != nil {
			return nil, nil, err
		}
		response.Belief = belief

	case "update":
		belief, err := h.probabilisticReasoner.UpdateBelief(input.BeliefID, input.EvidenceID, input.Likelihood, input.EvidenceProb)
		if err != nil {
			return nil, nil, err
		}
		response.Belief = belief

	case "get":
		belief, err := h.probabilisticReasoner.GetBelief(input.BeliefID)
		if err != nil {
			return nil, nil, err
		}
		response.Belief = belief

	case "combine":
		combinedProb, err := h.probabilisticReasoner.CombineBeliefs(input.BeliefIDs, input.CombineOp)
		if err != nil {
			return nil, nil, err
		}
		response.CombinedProb = combinedProb

	default:
		return nil, nil, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleAssessEvidence processes evidence assessment requests
func (h *ProbabilisticHandler) HandleAssessEvidence(ctx context.Context, req *mcp.CallToolRequest, input AssessEvidenceRequest) (*mcp.CallToolResult, *AssessEvidenceResponse, error) {
	if err := ValidateAssessEvidenceRequest(&input); err != nil {
		return nil, nil, err
	}

	evidence, err := h.evidenceAnalyzer.AssessEvidence(input.Content, input.Source, input.ClaimID, input.SupportsClaim)
	if err != nil {
		return nil, nil, err
	}

	response := &AssessEvidenceResponse{
		Evidence: evidence,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleDetectContradictions processes contradiction detection requests
func (h *ProbabilisticHandler) HandleDetectContradictions(ctx context.Context, req *mcp.CallToolRequest, input DetectContradictionsRequest) (*mcp.CallToolResult, *DetectContradictionsResponse, error) {
	if err := ValidateDetectContradictionsRequest(&input); err != nil {
		return nil, nil, err
	}

	var thoughts []*types.Thought

	// Gather thoughts based on input
	if len(input.ThoughtIDs) > 0 {
		for _, id := range input.ThoughtIDs {
			thought, err := h.storage.GetThought(id)
			if err != nil {
				return nil, nil, fmt.Errorf("thought not found: %s", id)
			}
			thoughts = append(thoughts, thought)
		}
	} else if input.BranchID != "" {
		branch, err := h.storage.GetBranch(input.BranchID)
		if err != nil {
			return nil, nil, err
		}
		thoughts = branch.Thoughts
	} else if input.Mode != "" {
		mode := types.ThinkingMode(input.Mode)
		thoughts = h.storage.SearchThoughts("", mode, 1000, 0)
	} else {
		// Check all thoughts
		thoughts = h.storage.SearchThoughts("", "", 1000, 0)
	}

	contradictions, err := h.contradictionDetector.DetectContradictions(thoughts)
	if err != nil {
		return nil, nil, err
	}

	response := &DetectContradictionsResponse{
		Contradictions: contradictions,
		Count:          len(contradictions),
		Status:         "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Validation Functions
// ============================================================================

// ValidateProbabilisticReasoningRequest validates a ProbabilisticReasoningRequest
func ValidateProbabilisticReasoningRequest(req *ProbabilisticReasoningRequest) error {
	// Validate operation
	validOps := map[string]bool{"create": true, "update": true, "get": true, "combine": true}
	if !validOps[req.Operation] {
		return &ValidationError{"operation", fmt.Sprintf("operation must be 'create', 'update', 'get', or 'combine'. You provided: '%s'", req.Operation)}
	}

	// Validate based on operation
	switch req.Operation {
	case "create":
		if len(req.Statement) == 0 {
			return &ValidationError{"statement", "statement is required for create operation. Example: {\"operation\": \"create\", \"statement\": \"It will rain tomorrow\", \"prior_prob\": 0.3}"}
		}
		if len(req.Statement) > MaxContentLength {
			return &ValidationError{"statement", fmt.Sprintf("statement exceeds max length of %d", MaxContentLength)}
		}
		if !utf8.ValidString(req.Statement) {
			return &ValidationError{"statement", "statement must be valid UTF-8"}
		}
		if req.PriorProb < 0 || req.PriorProb > 1 {
			return &ValidationError{"prior_prob", fmt.Sprintf("prior_prob must be between 0 and 1 (you provided: %.2f)", req.PriorProb)}
		}

	case "update":
		if len(req.BeliefID) == 0 {
			return &ValidationError{"belief_id", "belief_id is required for update operation. First create a belief, then update it with evidence. Example: {\"operation\": \"update\", \"belief_id\": \"belief_123\", \"evidence_id\": \"ev_456\", \"likelihood\": 0.8, \"evidence_prob\": 0.6}"}
		}
		if len(req.EvidenceID) == 0 {
			return &ValidationError{"evidence_id", "evidence_id is required for update operation"}
		}
		if req.Likelihood < 0 || req.Likelihood > 1 {
			return &ValidationError{"likelihood", fmt.Sprintf("likelihood must be between 0 and 1 (you provided: %.2f)", req.Likelihood)}
		}
		if req.EvidenceProb <= 0 || req.EvidenceProb > 1 {
			return &ValidationError{"evidence_prob", fmt.Sprintf("evidence_prob must be between 0 and 1 exclusive of 0 (you provided: %.2f)", req.EvidenceProb)}
		}

	case "get":
		if len(req.BeliefID) == 0 {
			return &ValidationError{"belief_id", "belief_id is required for get operation. Example: {\"operation\": \"get\", \"belief_id\": \"belief_123\"}"}
		}

	case "combine":
		if len(req.BeliefIDs) == 0 {
			return &ValidationError{"belief_ids", "at least one belief_id is required for combine operation. Example: {\"operation\": \"combine\", \"belief_ids\": [\"belief_1\", \"belief_2\"], \"combine_op\": \"and\"}"}
		}
		if len(req.BeliefIDs) > 50 {
			return &ValidationError{"belief_ids", "too many belief_ids (max 50)"}
		}
		validCombineOps := map[string]bool{"and": true, "or": true}
		if !validCombineOps[req.CombineOp] {
			return &ValidationError{"combine_op", fmt.Sprintf("combine_op must be 'and' or 'or' (you provided: '%s')", req.CombineOp)}
		}
	}

	return nil
}

// ValidateAssessEvidenceRequest validates an AssessEvidenceRequest
func ValidateAssessEvidenceRequest(req *AssessEvidenceRequest) error {
	if len(req.Content) == 0 {
		return &ValidationError{"content", "content is required"}
	}
	if len(req.Content) > MaxContentLength {
		return &ValidationError{"content", fmt.Sprintf("content exceeds max length of %d", MaxContentLength)}
	}
	if !utf8.ValidString(req.Content) {
		return &ValidationError{"content", "content must be valid UTF-8"}
	}

	if len(req.Source) == 0 {
		return &ValidationError{"source", "source is required"}
	}
	if len(req.Source) > MaxQueryLength {
		return &ValidationError{"source", fmt.Sprintf("source exceeds max length of %d", MaxQueryLength)}
	}
	if !utf8.ValidString(req.Source) {
		return &ValidationError{"source", "source must be valid UTF-8"}
	}

	if len(req.ClaimID) > MaxBranchIDLength {
		return &ValidationError{"claim_id", "claim_id too long"}
	}

	return nil
}

// ValidateDetectContradictionsRequest validates a DetectContradictionsRequest
func ValidateDetectContradictionsRequest(req *DetectContradictionsRequest) error {
	if len(req.ThoughtIDs) > 100 {
		return &ValidationError{"thought_ids", "too many thought_ids (max 100)"}
	}

	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}

	if req.Mode != "" {
		validModes := map[string]bool{"linear": true, "tree": true, "divergent": true}
		if !validModes[req.Mode] {
			return &ValidationError{"mode", fmt.Sprintf("mode must be 'linear', 'tree', or 'divergent' (you provided: '%s')", req.Mode)}
		}
	}

	return nil
}
