package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// ThinkingHandler handles think and history operations
type ThinkingHandler struct {
	storage          storage.Storage
	linear           *modes.LinearMode
	tree             *modes.TreeMode
	divergent        *modes.DivergentMode
	auto             *modes.AutoMode
	validator        *validation.LogicValidator
	metadataGen      *MetadataGenerator
}

// NewThinkingHandler creates a new thinking handler
func NewThinkingHandler(
	store storage.Storage,
	linear *modes.LinearMode,
	tree *modes.TreeMode,
	divergent *modes.DivergentMode,
	auto *modes.AutoMode,
	validator *validation.LogicValidator,
) *ThinkingHandler {
	return &ThinkingHandler{
		storage:     store,
		linear:      linear,
		tree:        tree,
		divergent:   divergent,
		auto:        auto,
		validator:   validator,
		metadataGen: NewMetadataGenerator(),
	}
}

// ThinkRequest represents a thinking request
type ThinkRequest struct {
	Content           string                  `json:"content"`
	Mode              string                  `json:"mode,omitempty"`
	Type              string                  `json:"type,omitempty"`
	BranchID          string                  `json:"branch_id,omitempty"`
	ParentID          string                  `json:"parent_id,omitempty"`
	PreviousThoughtID string                  `json:"previous_thought_id,omitempty"`
	Confidence        float64                 `json:"confidence,omitempty"`
	KeyPoints         []string                `json:"key_points,omitempty"`
	ForceRebellion    bool                    `json:"force_rebellion,omitempty"`
	CrossRefs         []modes.CrossRefInput   `json:"cross_refs,omitempty"`
	RequireValidation bool                    `json:"require_validation,omitempty"`
}

// ThinkResponse represents a thinking response
type ThinkResponse struct {
	ThoughtID            string                `json:"thought_id"`
	Mode                 string                `json:"mode"`
	Status               string                `json:"status"`
	BranchID             string                `json:"branch_id,omitempty"`
	Confidence           float64               `json:"confidence"`
	IsRebellion          bool                  `json:"is_rebellion,omitempty"`
	ChallengesAssumption bool                  `json:"challenges_assumption,omitempty"`
	InsightCount         int                   `json:"insight_count,omitempty"`
	CrossRefCount        int                   `json:"cross_ref_count,omitempty"`
	IsValid              bool                  `json:"is_valid,omitempty"`
	Metadata             *types.ResponseMetadata `json:"metadata,omitempty"`
}

// HistoryRequest represents a history request
type HistoryRequest struct {
	BranchID string `json:"branch_id,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// HistoryResponse represents a history response
type HistoryResponse struct {
	Thoughts   interface{} `json:"thoughts"`
	TotalCount int         `json:"total_count"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

// HandleThink processes a thinking request
func (h *ThinkingHandler) HandleThink(ctx context.Context, req *mcp.CallToolRequest, input ThinkRequest) (*mcp.CallToolResult, *ThinkResponse, error) {
	// Set default confidence if not provided
	if input.Confidence == 0 {
		input.Confidence = 0.8
	}

	// Convert to mode input
	thoughtInput := modes.ThoughtInput{
		Content:           input.Content,
		Type:              input.Type,
		BranchID:          input.BranchID,
		ParentID:          input.ParentID,
		PreviousThoughtID: input.PreviousThoughtID,
		Confidence:        input.Confidence,
		KeyPoints:         input.KeyPoints,
		ForceRebellion:    input.ForceRebellion,
		CrossRefs:         input.CrossRefs,
	}

	// Select mode
	var result *modes.ThoughtResult
	var err error

	switch input.Mode {
	case "linear", "":
		result, err = h.linear.ProcessThought(ctx, thoughtInput)
	case "tree":
		result, err = h.tree.ProcessThought(ctx, thoughtInput)
	case "divergent":
		result, err = h.divergent.ProcessThought(ctx, thoughtInput)
	case "auto":
		result, err = h.auto.ProcessThought(ctx, thoughtInput)
	default:
		return nil, nil, fmt.Errorf("invalid mode: %s", input.Mode)
	}

	if err != nil {
		return nil, nil, err
	}

	// Validate if required
	isValid := false
	if input.RequireValidation {
		thought, _ := h.storage.GetThought(result.ThoughtID)
		if thought != nil {
			validationResult, _ := h.validator.ValidateThought(thought)
			if validationResult != nil {
				isValid = validationResult.IsValid
			}
		}
	}

	// Retrieve the thought for metadata generation
	thought, _ := h.storage.GetThought(result.ThoughtID)

	// Generate metadata to guide Claude's next actions
	var metadata *types.ResponseMetadata
	if thought != nil {
		metadata = h.metadataGen.GenerateThinkMetadata(
			thought,
			types.ThinkingMode(result.Mode),
			result.Confidence,
			result.InsightCount > 0,
			result.CrossRefCount > 0,
		)
	}

	response := &ThinkResponse{
		ThoughtID:            result.ThoughtID,
		Mode:                 string(result.Mode),
		Status:               "success",
		BranchID:             result.BranchID,
		Confidence:           result.Confidence,
		IsRebellion:          result.IsRebellion,
		ChallengesAssumption: result.ChallengesAssumption,
		InsightCount:         result.InsightCount,
		CrossRefCount:        result.CrossRefCount,
		IsValid:              isValid,
		Metadata:             metadata,
	}

	return &mcp.CallToolResult{}, response, nil
}

// HandleHistory retrieves thought history
func (h *ThinkingHandler) HandleHistory(ctx context.Context, req *mcp.CallToolRequest, input HistoryRequest) (*mcp.CallToolResult, *HistoryResponse, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	var thoughts interface{}

	if input.BranchID != "" {
		// Branch-specific history
		branch, err := h.storage.GetBranch(input.BranchID)
		if err != nil {
			return nil, nil, err
		}
		thoughts = branch.Thoughts
	} else if input.Mode != "" {
		// Mode-filtered history
		thoughts = h.storage.SearchThoughts("", types.ThinkingMode(input.Mode), limit, input.Offset)
	} else {
		// General history
		thoughts = h.storage.SearchThoughts("", "", limit, input.Offset)
	}

	response := &HistoryResponse{
		Thoughts:   thoughts,
		TotalCount: 0, // Would need to implement count
		Limit:      limit,
		Offset:     input.Offset,
	}

	return &mcp.CallToolResult{}, response, nil
}
