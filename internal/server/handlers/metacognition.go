package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// MetacognitionHandler handles self-evaluation and bias detection operations
type MetacognitionHandler struct {
	storage         storage.Storage
	selfEvaluator   *metacognition.SelfEvaluator
	biasDetector    *metacognition.BiasDetector
	fallacyDetector *validation.FallacyDetector
}

// NewMetacognitionHandler creates a new metacognition handler
func NewMetacognitionHandler(
	store storage.Storage,
	selfEvaluator *metacognition.SelfEvaluator,
	biasDetector *metacognition.BiasDetector,
	fallacyDetector *validation.FallacyDetector,
) *MetacognitionHandler {
	return &MetacognitionHandler{
		storage:         store,
		selfEvaluator:   selfEvaluator,
		biasDetector:    biasDetector,
		fallacyDetector: fallacyDetector,
	}
}

// ============================================================================
// Request/Response Types
// ============================================================================

// SelfEvaluateRequest represents a self-evaluation request
type SelfEvaluateRequest struct {
	ThoughtID string `json:"thought_id,omitempty"`
	BranchID  string `json:"branch_id,omitempty"`
}

// SelfEvaluateResponse represents a self-evaluation response
type SelfEvaluateResponse struct {
	Evaluation *types.SelfEvaluation `json:"evaluation"`
	Status     string                `json:"status"`
}

// DetectBiasesRequest represents a bias/fallacy detection request
type DetectBiasesRequest struct {
	ThoughtID string `json:"thought_id,omitempty"`
	BranchID  string `json:"branch_id,omitempty"`
}

// DetectedIssue represents either a bias or fallacy with a unified structure
type DetectedIssue struct {
	Type        string  `json:"type"`        // "bias" or "fallacy"
	Name        string  `json:"name"`        // e.g., "confirmation_bias", "ad_hominem"
	Category    string  `json:"category"`    // e.g., "cognitive", "formal", "informal"
	Description string  `json:"description"` // What the issue is
	Location    string  `json:"location"`    // Where it was found
	Example     string  `json:"example"`     // The problematic content
	Mitigation  string  `json:"mitigation"`  // How to fix/avoid it
	Confidence  float64 `json:"confidence"`  // Detection confidence
}

// DetectBiasesResponse represents a bias/fallacy detection response
type DetectBiasesResponse struct {
	Biases    []*types.CognitiveBias       `json:"biases"`
	Fallacies []*validation.DetectedFallacy `json:"fallacies"`
	Combined  []*DetectedIssue             `json:"combined"` // Unified list of all issues
	Count     int                          `json:"count"`    // Total count
	Status    string                       `json:"status"`
}

// ============================================================================
// Handler Methods
// ============================================================================

// HandleSelfEvaluate processes self-evaluation requests
func (h *MetacognitionHandler) HandleSelfEvaluate(ctx context.Context, req *mcp.CallToolRequest, input SelfEvaluateRequest) (*mcp.CallToolResult, *SelfEvaluateResponse, error) {
	if err := ValidateSelfEvaluateRequest(&input); err != nil {
		return nil, nil, err
	}

	var evaluation *types.SelfEvaluation
	var err error

	if input.ThoughtID != "" {
		thought, getErr := h.storage.GetThought(input.ThoughtID)
		if getErr != nil {
			return nil, nil, getErr
		}
		evaluation, err = h.selfEvaluator.EvaluateThought(thought)
	} else if input.BranchID != "" {
		branch, getErr := h.storage.GetBranch(input.BranchID)
		if getErr != nil {
			return nil, nil, getErr
		}
		evaluation, err = h.selfEvaluator.EvaluateBranch(branch)
	} else {
		return nil, nil, fmt.Errorf("either thought_id or branch_id must be provided")
	}

	if err != nil {
		return nil, nil, err
	}

	response := &SelfEvaluateResponse{
		Evaluation: evaluation,
		Status:     "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleDetectBiases processes bias and fallacy detection requests
func (h *MetacognitionHandler) HandleDetectBiases(ctx context.Context, req *mcp.CallToolRequest, input DetectBiasesRequest) (*mcp.CallToolResult, *DetectBiasesResponse, error) {
	if err := ValidateDetectBiasesRequest(&input); err != nil {
		return nil, nil, err
	}

	var biases []*types.CognitiveBias
	var fallacies []*validation.DetectedFallacy
	var content string
	var err error

	// Get the content to analyze
	if input.ThoughtID != "" {
		thought, getErr := h.storage.GetThought(input.ThoughtID)
		if getErr != nil {
			return nil, nil, getErr
		}
		content = thought.Content
		biases, err = h.biasDetector.DetectBiases(thought)
	} else if input.BranchID != "" {
		branch, getErr := h.storage.GetBranch(input.BranchID)
		if getErr != nil {
			return nil, nil, getErr
		}
		// Combine all thoughts in the branch for analysis
		for _, thought := range branch.Thoughts {
			if thought != nil {
				content += thought.Content + "\n"
			}
		}
		biases, err = h.biasDetector.DetectBiasesInBranch(branch)
	} else {
		return nil, nil, fmt.Errorf("either thought_id or branch_id must be provided")
	}

	if err != nil {
		return nil, nil, err
	}

	// Also detect fallacies in the content
	// Check both formal and informal fallacies by default
	fallacies = h.fallacyDetector.DetectFallacies(content, true, true)

	// Create combined list of all issues
	combined := make([]*DetectedIssue, 0, len(biases)+len(fallacies))

	// Add biases to combined list
	for _, bias := range biases {
		// Extract severity as confidence score
		confidenceScore := 0.5 // default
		switch bias.Severity {
		case "high":
			confidenceScore = 0.9
		case "medium":
			confidenceScore = 0.6
		case "low":
			confidenceScore = 0.3
		}

		combined = append(combined, &DetectedIssue{
			Type:        "bias",
			Name:        bias.BiasType,
			Category:    "cognitive",
			Description: bias.Description,
			Location:    bias.DetectedIn,
			Example:     "", // CognitiveBias doesn't have Example field
			Mitigation:  bias.Mitigation,
			Confidence:  confidenceScore,
		})
	}

	// Add fallacies to combined list
	for _, fallacy := range fallacies {
		combined = append(combined, &DetectedIssue{
			Type:        "fallacy",
			Name:        fallacy.Type,
			Category:    string(fallacy.Category),
			Description: fallacy.Explanation,
			Location:    fallacy.Location,
			Example:     fallacy.Example,
			Mitigation:  fallacy.Correction,
			Confidence:  fallacy.Confidence,
		})
	}

	response := &DetectBiasesResponse{
		Biases:    biases,
		Fallacies: fallacies,
		Combined:  combined,
		Count:     len(combined),
		Status:    "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Validation Functions
// ============================================================================

// ValidateSelfEvaluateRequest validates a SelfEvaluateRequest
func ValidateSelfEvaluateRequest(req *SelfEvaluateRequest) error {
	if req.ThoughtID == "" && req.BranchID == "" {
		return &ValidationError{"request", "either thought_id or branch_id must be provided"}
	}

	if req.ThoughtID != "" && req.BranchID != "" {
		return &ValidationError{"request", "only one of thought_id or branch_id should be provided"}
	}

	if len(req.ThoughtID) > MaxBranchIDLength {
		return &ValidationError{"thought_id", "thought_id too long"}
	}

	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}

	return nil
}

// ValidateDetectBiasesRequest validates a DetectBiasesRequest
func ValidateDetectBiasesRequest(req *DetectBiasesRequest) error {
	if req.ThoughtID == "" && req.BranchID == "" {
		return &ValidationError{"request", "either thought_id or branch_id must be provided. Example: {\"thought_id\": \"thought_123\"} or {\"branch_id\": \"branch_456\"}"}
	}

	if req.ThoughtID != "" && req.BranchID != "" {
		return &ValidationError{"request", "only one of thought_id or branch_id should be provided, not both"}
	}

	if len(req.ThoughtID) > MaxBranchIDLength {
		return &ValidationError{"thought_id", "thought_id too long"}
	}

	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}

	return nil
}
