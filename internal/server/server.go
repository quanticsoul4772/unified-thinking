// Package server implements the MCP (Model Context Protocol) server for unified thinking.
//
// This package provides the MCP server implementation that exposes 11 tools for
// thought processing, validation, and search. All responses are JSON formatted
// for consumption by Claude AI via stdio transport.
//
// Available tools:
//   - think: Main thinking tool with multiple cognitive modes
//   - history: View thinking history
//   - list-branches: List all thinking branches
//   - focus-branch: Switch active branch
//   - branch-history: Get detailed branch history
//   - validate: Validate thought logical consistency
//   - prove: Attempt logical proof
//   - check-syntax: Validate logical statement syntax
//   - search: Search through all thoughts
//   - get-metrics: Get system performance and usage metrics
//   - recent-branches: Get recently accessed branches for quick context switching
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

// UnifiedServer coordinates all thinking modes and provides MCP tool handlers.
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

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "get-metrics",
		Description: "Get system performance and usage metrics",
	}, s.handleGetMetrics)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "recent-branches",
		Description: "Get recently accessed branches for quick context switching",
	}, s.handleRecentBranches)
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
	// Validate input
	if err := ValidateThinkRequest(&input); err != nil {
		return nil, nil, err
	}

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

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type HistoryRequest struct {
	Mode     string `json:"mode,omitempty"`
	BranchID string `json:"branch_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

type HistoryResponse struct {
	Thoughts []*types.Thought `json:"thoughts"`
}

func (s *UnifiedServer) handleHistory(ctx context.Context, req *mcp.CallToolRequest, input HistoryRequest) (*mcp.CallToolResult, *HistoryResponse, error) {
	// Validate input
	if err := ValidateHistoryRequest(&input); err != nil {
		return nil, nil, err
	}

	// Set default limit if not specified
	limit := input.Limit
	if limit == 0 {
		limit = 100 // Default to 100 results
	}

	var thoughts []*types.Thought

	if input.BranchID != "" {
		branch, err := s.storage.GetBranch(input.BranchID)
		if err != nil {
			return nil, nil, err
		}
		// Apply pagination to branch thoughts
		thoughts = paginateThoughts(branch.Thoughts, limit, input.Offset)
	} else {
		mode := types.ThinkingMode(input.Mode)
		thoughts = s.storage.SearchThoughts("", mode, limit, input.Offset)
	}

	response := &HistoryResponse{Thoughts: thoughts}
	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// paginateThoughts applies limit and offset to a slice of thoughts
func paginateThoughts(thoughts []*types.Thought, limit, offset int) []*types.Thought {
	// Handle offset beyond slice length
	if offset >= len(thoughts) {
		return []*types.Thought{}
	}

	start := offset
	end := offset + limit
	if end > len(thoughts) {
		end = len(thoughts)
	}

	return thoughts[start:end]
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

	response := &ListBranchesResponse{
		Branches:       branches,
		ActiveBranchID: activeID,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type FocusBranchRequest struct {
	BranchID string `json:"branch_id"`
}

type FocusBranchResponse struct {
	Status         string `json:"status"`
	ActiveBranchID string `json:"active_branch_id"`
}

func (s *UnifiedServer) handleFocusBranch(ctx context.Context, req *mcp.CallToolRequest, input FocusBranchRequest) (*mcp.CallToolResult, *FocusBranchResponse, error) {
	// Validate input
	if err := ValidateFocusBranchRequest(&input); err != nil {
		return nil, nil, err
	}

	// Check if branch is already active
	activeBranch, _ := s.storage.GetActiveBranch()
	if activeBranch != nil && activeBranch.ID == input.BranchID {
		response := &FocusBranchResponse{
			Status:         "already_active",
			ActiveBranchID: input.BranchID,
		}
		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	}

	if err := s.storage.SetActiveBranch(input.BranchID); err != nil {
		return nil, nil, err
	}

	response := &FocusBranchResponse{
		Status:         "success",
		ActiveBranchID: input.BranchID,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type BranchHistoryRequest struct {
	BranchID string `json:"branch_id"`
}

func (s *UnifiedServer) handleBranchHistory(ctx context.Context, req *mcp.CallToolRequest, input BranchHistoryRequest) (*mcp.CallToolResult, *modes.BranchHistory, error) {
	// Validate input
	if err := ValidateBranchHistoryRequest(&input); err != nil {
		return nil, nil, err
	}

	history, err := s.tree.GetBranchHistory(ctx, input.BranchID)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(history),
	}, history, nil
}

type ValidateRequest struct {
	ThoughtID string `json:"thought_id"`
}

type ValidateResponse struct {
	IsValid bool   `json:"is_valid"`
	Reason  string `json:"reason"`
}

func (s *UnifiedServer) handleValidate(ctx context.Context, req *mcp.CallToolRequest, input ValidateRequest) (*mcp.CallToolResult, *ValidateResponse, error) {
	// Validate input
	if err := ValidateValidateRequest(&input); err != nil {
		return nil, nil, err
	}

	thought, err := s.storage.GetThought(input.ThoughtID)
	if err != nil {
		return nil, nil, err
	}

	validationResult, err := s.validator.ValidateThought(thought)
	if err != nil {
		return nil, nil, err
	}

	response := &ValidateResponse{
		IsValid: validationResult.IsValid,
		Reason:  validationResult.Reason,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
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
	// Validate input
	if err := ValidateProveRequest(&input); err != nil {
		return nil, nil, err
	}

	result := s.validator.Prove(input.Premises, input.Conclusion)

	response := &ProveResponse{
		IsProvable: result.IsProvable,
		Premises:   result.Premises,
		Conclusion: result.Conclusion,
		Steps:      result.Steps,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type CheckSyntaxRequest struct {
	Statements []string `json:"statements"`
}

type CheckSyntaxResponse struct {
	Checks []validation.StatementCheck `json:"checks"`
}

func (s *UnifiedServer) handleCheckSyntax(ctx context.Context, req *mcp.CallToolRequest, input CheckSyntaxRequest) (*mcp.CallToolResult, *CheckSyntaxResponse, error) {
	// Validate input
	if err := ValidateCheckSyntaxRequest(&input); err != nil {
		return nil, nil, err
	}

	checks := s.validator.CheckWellFormed(input.Statements)

	response := &CheckSyntaxResponse{
		Checks: checks,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type SearchRequest struct {
	Query  string `json:"query"`
	Mode   string `json:"mode,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

type SearchResponse struct {
	Thoughts []*types.Thought `json:"thoughts"`
}

func (s *UnifiedServer) handleSearch(ctx context.Context, req *mcp.CallToolRequest, input SearchRequest) (*mcp.CallToolResult, *SearchResponse, error) {
	// Validate input
	if err := ValidateSearchRequest(&input); err != nil {
		return nil, nil, err
	}

	// Set default limit if not specified
	limit := input.Limit
	if limit == 0 {
		limit = 100 // Default to 100 results
	}

	mode := types.ThinkingMode(input.Mode)
	thoughts := s.storage.SearchThoughts(input.Query, mode, limit, input.Offset)

	response := &SearchResponse{Thoughts: thoughts}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
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

type MetricsResponse struct {
	TotalThoughts     int            `json:"total_thoughts"`
	TotalBranches     int            `json:"total_branches"`
	TotalInsights     int            `json:"total_insights"`
	TotalValidations  int            `json:"total_validations"`
	ThoughtsByMode    map[string]int `json:"thoughts_by_mode"`
	AverageConfidence float64        `json:"average_confidence"`
}

type RecentBranchesResponse struct {
	ActiveBranchID string          `json:"active_branch_id"`
	RecentBranches []*types.Branch `json:"recent_branches"`
	Count          int             `json:"count"`
}

func (s *UnifiedServer) handleGetMetrics(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *MetricsResponse, error) {
	metrics := s.storage.GetMetrics()

	response := &MetricsResponse{
		TotalThoughts:     metrics.TotalThoughts,
		TotalBranches:     metrics.TotalBranches,
		TotalInsights:     metrics.TotalInsights,
		TotalValidations:  metrics.TotalValidations,
		ThoughtsByMode:    metrics.ThoughtsByMode,
		AverageConfidence: metrics.AverageConfidence,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

func (s *UnifiedServer) handleRecentBranches(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *RecentBranchesResponse, error) {
	branches, err := s.storage.GetRecentBranches()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get recent branches: %w", err)
	}

	// Get active branch for context
	activeBranch, _ := s.storage.GetActiveBranch()
	activeBranchID := ""
	if activeBranch != nil {
		activeBranchID = activeBranch.ID
	}

	response := &RecentBranchesResponse{
		ActiveBranchID: activeBranchID,
		RecentBranches: branches,
		Count:          len(branches),
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}
