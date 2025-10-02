package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// BranchHandler handles branch operations
type BranchHandler struct {
	storage storage.Storage
	tree    *modes.TreeMode
}

// NewBranchHandler creates a new branch handler
func NewBranchHandler(store storage.Storage, tree *modes.TreeMode) *BranchHandler {
	return &BranchHandler{
		storage: store,
		tree:    tree,
	}
}

// EmptyRequest represents an empty request
type EmptyRequest struct{}

// ListBranchesResponse represents list branches response
type ListBranchesResponse struct {
	Branches []*types.Branch `json:"branches"`
	Count    int             `json:"count"`
}

// FocusBranchRequest represents focus branch request
type FocusBranchRequest struct {
	BranchID string `json:"branch_id"`
}

// FocusBranchResponse represents focus branch response
type FocusBranchResponse struct {
	ActiveBranchID string `json:"active_branch_id"`
	Status         string `json:"status"`
}

// BranchHistoryRequest represents branch history request
type BranchHistoryRequest struct {
	BranchID string `json:"branch_id"`
}

// RecentBranchesResponse represents recent branches response
type RecentBranchesResponse struct {
	Branches []*types.Branch `json:"branches"`
	Count    int             `json:"count"`
}

// HandleListBranches lists all branches
func (h *BranchHandler) HandleListBranches(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *ListBranchesResponse, error) {
	branches := h.storage.ListBranches()

	response := &ListBranchesResponse{
		Branches: branches,
		Count:    len(branches),
	}

	return &mcp.CallToolResult{}, response, nil
}

// HandleFocusBranch sets the active branch
func (h *BranchHandler) HandleFocusBranch(ctx context.Context, req *mcp.CallToolRequest, input FocusBranchRequest) (*mcp.CallToolResult, *FocusBranchResponse, error) {
	if input.BranchID == "" {
		return nil, nil, fmt.Errorf("branch_id is required")
	}

	// Verify branch exists
	_, err := h.storage.GetBranch(input.BranchID)
	if err != nil {
		return nil, nil, fmt.Errorf("branch not found: %s", input.BranchID)
	}

	// Set as active
	if err := h.storage.SetActiveBranch(input.BranchID); err != nil {
		return nil, nil, err
	}

	response := &FocusBranchResponse{
		ActiveBranchID: input.BranchID,
		Status:         "success",
	}

	return &mcp.CallToolResult{}, response, nil
}

// HandleBranchHistory retrieves detailed branch history
func (h *BranchHandler) HandleBranchHistory(ctx context.Context, req *mcp.CallToolRequest, input BranchHistoryRequest) (*mcp.CallToolResult, *modes.BranchHistory, error) {
	if input.BranchID == "" {
		return nil, nil, fmt.Errorf("branch_id is required")
	}

	history, err := h.tree.GetBranchHistory(ctx, input.BranchID)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{}, history, nil
}

// HandleRecentBranches retrieves recently accessed branches
func (h *BranchHandler) HandleRecentBranches(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *RecentBranchesResponse, error) {
	recent, err := h.storage.GetRecentBranches()
	if err != nil {
		return nil, nil, err
	}

	response := &RecentBranchesResponse{
		Branches: recent,
		Count:    len(recent),
	}

	return &mcp.CallToolResult{}, response, nil
}
