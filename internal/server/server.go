package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

type UnifiedServer struct {
	storage   *storage.MemoryStorage
	linear    *modes.LinearMode
	tree      *modes.TreeMode
	divergent *modes.DivergentMode
	auto      *modes.AutoMode
	validator *validation.LogicValidator
}

func NewUnifiedServer(
	store *storage.MemoryStorage,
	linear *modes.LinearMode,
	tree *modes.TreeMode,
	divergent *modes.DivergentMode,
	auto *modes.AutoMode,
	validator *validation.LogicValidator,
) *UnifiedServer {
	return &UnifiedServer{
		storage:   store,
		linear:    linear,
		tree:      tree,
		divergent: divergent,
		auto:      auto,
		validator: validator,
	}
}

func (s *UnifiedServer) RegisterTools(mcpServer *mcp.Server) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "think",
		Description: "Main thinking tool supporting multiple cognitive modes",
	}, s.handleThink)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "history",
		Description: "View thinking history",
	}, s.handleHistory)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list-branches",
		Description: "List all thinking branches",
	}, s.handleListBranches)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "focus-branch",
		Description: "Switch the active thinking branch",
	}, s.handleFocusBranch)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "branch-history",
		Description: "Get detailed history of a specific branch",
	}, s.handleBranchHistory)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "validate",
		Description: "Validate a thought for logical consistency",
	}, s.handleValidate)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "prove",
		Description: "Attempt to prove a logical conclusion from premises",
	}, s.handleProve)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "check-syntax",
		Description: "Validate syntax of logical statements",
	}, s.handleCheckSyntax)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "search",
		Description: "Search through all thoughts",
	}, s.handleSearch)
}

type ThinkRequest struct {
	Content              string          `json:"content"`
	Mode                 string          `json:"mode"`
	Type                 string          `json:"type,omitempty"`
	BranchID             string          `json:"branch_id,omitempty"`
	ParentID             string          `json:"parent_id,omitempty"`
	Confidence           float64         `json:"confidence,omitempty"`
	KeyPoints            []string        `json:"key_points,omitempty"`
	RequireValidation    bool            `json:"require_validation,omitempty"`
	ChallengeAssumptions bool            `json:"challenge_assumptions,omitempty"`
	ForceRebellion       bool            `json:"force_rebellion,omitempty"`
	CrossRefs            []CrossRefInput `json:"cross_refs,omitempty"`
}

type CrossRefInput struct {
	ToBranch string  `json:"to_branch"`
	Type     string  `json:"type"`
	Reason   string  `json:"reason"`
	Strength float64 `json:"strength"`
}

type ThinkResponse struct {
	ThoughtID    string  `json:"thought_id"`
	Mode         string  `json:"mode"`
	BranchID     string  `json:"branch_id,omitempty"`
	Status       string  `json:"status"`
	Priority     float64 `json:"priority,omitempty"`
	Confidence   float64 `json:"confidence"`
	InsightCount int     `json:"insight_count,omitempty"`
	IsValid      bool    `json:"is_valid,omitempty"`
}

func (s *UnifiedServer) handleThink(ctx context.Context, req *mcp.CallToolRequest, input ThinkRequest) (*mcp.CallToolResult, *ThinkResponse, error) {
	thoughtInput := modes.ThoughtInput{
		Content:        input.Content,
		Type:           input.Type,
		BranchID:       input.BranchID,
		ParentID:       input.ParentID,
		Confidence:     input.Confidence,
		KeyPoints:      input.KeyPoints,
		ForceRebellion: input.ForceRebellion,
		CrossRefs:      convertCrossRefs(input.CrossRefs),
	}

	if thoughtInput.Confidence == 0 {
		thoughtInput.Confidence = 0.8
	}

	var result *modes.ThoughtResult
	var err error

	mode := types.ThinkingMode(input.Mode)
	if mode == "" || mode == types.ModeAuto {
		result, err = s.auto.ProcessThought(ctx, thoughtInput)
	} else {
		switch mode {
		case types.ModeLinear:
			result, err = s.linear.ProcessThought(ctx, thoughtInput)
		case types.ModeTree:
			result, err = s.tree.ProcessThought(ctx, thoughtInput)
		case types.ModeDivergent:
			result, err = s.divergent.ProcessThought(ctx, thoughtInput)
		default:
			return nil, nil, fmt.Errorf("unknown mode: %s", mode)
		}
	}

	if err != nil {
		return nil, nil, err
	}

	isValid := true
	if input.RequireValidation {
		thought, _ := s.storage.GetThought(result.ThoughtID)
		if thought != nil {
			validationResult, _ := s.validator.ValidateThought(thought)
			if validationResult != nil {
				isValid = validationResult.IsValid
			}
		}
	}

	response := &ThinkResponse{
		ThoughtID:    result.ThoughtID,
		Mode:         result.Mode,
		BranchID:     result.BranchID,
		Status:       "success",
		Priority:     result.Priority,
		Confidence:   result.Confidence,
		InsightCount: result.InsightCount,
		IsValid:      isValid,
	}

	return nil, response, nil
}

type HistoryRequest struct {
	Mode     string `json:"mode,omitempty"`
	BranchID string `json:"branch_id,omitempty"`
}

type HistoryResponse struct {
	Thoughts []*types.Thought `json:"thoughts"`
}

func (s *UnifiedServer) handleHistory(ctx context.Context, req *mcp.CallToolRequest, input HistoryRequest) (*mcp.CallToolResult, *HistoryResponse, error) {
	var thoughts []*types.Thought

	if input.BranchID != "" {
		branch, err := s.storage.GetBranch(input.BranchID)
		if err != nil {
			return nil, nil, err
		}
		thoughts = branch.Thoughts
	} else {
		mode := types.ThinkingMode(input.Mode)
		thoughts = s.storage.SearchThoughts("", mode)
	}

	return nil, &HistoryResponse{Thoughts: thoughts}, nil
}

type EmptyRequest struct{}

type ListBranchesResponse struct {
	Branches       []*types.Branch `json:"branches"`
	ActiveBranchID string          `json:"active_branch_id"`
}

func (s *UnifiedServer) handleListBranches(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *ListBranchesResponse, error) {
	branches := s.storage.ListBranches()

	activeBranch, _ := s.storage.GetActiveBranch()
	activeID := ""
	if activeBranch != nil {
		activeID = activeBranch.ID
	}

	return nil, &ListBranchesResponse{
		Branches:       branches,
		ActiveBranchID: activeID,
	}, nil
}

type FocusBranchRequest struct {
	BranchID string `json:"branch_id"`
}

type FocusBranchResponse struct {
	Status         string `json:"status"`
	ActiveBranchID string `json:"active_branch_id"`
}

func (s *UnifiedServer) handleFocusBranch(ctx context.Context, req *mcp.CallToolRequest, input FocusBranchRequest) (*mcp.CallToolResult, *FocusBranchResponse, error) {
	if err := s.storage.SetActiveBranch(input.BranchID); err != nil {
		return nil, nil, err
	}

	return nil, &FocusBranchResponse{
		Status:         "success",
		ActiveBranchID: input.BranchID,
	}, nil
}

type BranchHistoryRequest struct {
	BranchID string `json:"branch_id"`
}

func (s *UnifiedServer) handleBranchHistory(ctx context.Context, req *mcp.CallToolRequest, input BranchHistoryRequest) (*mcp.CallToolResult, *modes.BranchHistory, error) {
	history, err := s.tree.GetBranchHistory(ctx, input.BranchID)
	return nil, history, err
}

type ValidateRequest struct {
	ThoughtID string `json:"thought_id"`
}

type ValidateResponse struct {
	IsValid bool   `json:"is_valid"`
	Reason  string `json:"reason"`
}

func (s *UnifiedServer) handleValidate(ctx context.Context, req *mcp.CallToolRequest, input ValidateRequest) (*mcp.CallToolResult, *ValidateResponse, error) {
	thought, err := s.storage.GetThought(input.ThoughtID)
	if err != nil {
		return nil, nil, err
	}

	validationResult, err := s.validator.ValidateThought(thought)
	if err != nil {
		return nil, nil, err
	}

	return nil, &ValidateResponse{
		IsValid: validationResult.IsValid,
		Reason:  validationResult.Reason,
	}, nil
}

type ProveRequest struct {
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
}

type ProveResponse struct {
	IsProvable bool     `json:"is_provable"`
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
	Steps      []string `json:"steps"`
}

func (s *UnifiedServer) handleProve(ctx context.Context, req *mcp.CallToolRequest, input ProveRequest) (*mcp.CallToolResult, *ProveResponse, error) {
	result := s.validator.Prove(input.Premises, input.Conclusion)

	return nil, &ProveResponse{
		IsProvable: result.IsProvable,
		Premises:   result.Premises,
		Conclusion: result.Conclusion,
		Steps:      result.Steps,
	}, nil
}

type CheckSyntaxRequest struct {
	Statements []string `json:"statements"`
}

type CheckSyntaxResponse struct {
	Checks []validation.StatementCheck `json:"checks"`
}

func (s *UnifiedServer) handleCheckSyntax(ctx context.Context, req *mcp.CallToolRequest, input CheckSyntaxRequest) (*mcp.CallToolResult, *CheckSyntaxResponse, error) {
	checks := s.validator.CheckWellFormed(input.Statements)

	return nil, &CheckSyntaxResponse{
		Checks: checks,
	}, nil
}

type SearchRequest struct {
	Query string `json:"query"`
	Mode  string `json:"mode,omitempty"`
}

type SearchResponse struct {
	Thoughts []*types.Thought `json:"thoughts"`
}

func (s *UnifiedServer) handleSearch(ctx context.Context, req *mcp.CallToolRequest, input SearchRequest) (*mcp.CallToolResult, *SearchResponse, error) {
	mode := types.ThinkingMode(input.Mode)
	thoughts := s.storage.SearchThoughts(input.Query, mode)

	return nil, &SearchResponse{Thoughts: thoughts}, nil
}

func convertCrossRefs(input []CrossRefInput) []modes.CrossRefInput {
	result := make([]modes.CrossRefInput, len(input))
	for i, xref := range input {
		result[i] = modes.CrossRefInput{
			ToBranch: xref.ToBranch,
			Type:     xref.Type,
			Reason:   xref.Reason,
			Strength: xref.Strength,
		}
	}
	return result
}
