package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
)

// BacktrackingHandler handles backtracking operations
type BacktrackingHandler struct {
	manager *modes.BacktrackingManager
	storage storage.Storage
}

// NewBacktrackingHandler creates a new backtracking handler
func NewBacktrackingHandler(manager *modes.BacktrackingManager, store storage.Storage) *BacktrackingHandler {
	return &BacktrackingHandler{
		manager: manager,
		storage: store,
	}
}

// CreateCheckpointRequest represents a checkpoint creation request
type CreateCheckpointRequest struct {
	BranchID    string `json:"branch_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateCheckpointResponse represents the response
type CreateCheckpointResponse struct {
	CheckpointID string `json:"checkpoint_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	BranchID     string `json:"branch_id"`
	ThoughtCount int    `json:"thought_count"`
	InsightCount int    `json:"insight_count"`
	CreatedAt    string `json:"created_at"`
}

// HandleCreateCheckpoint creates a checkpoint
func (h *BacktrackingHandler) HandleCreateCheckpoint(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req, err := unmarshalRequest[CreateCheckpointRequest](params)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp, err := h.createCheckpoint(ctx, req)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// createCheckpoint is the typed internal implementation
func (h *BacktrackingHandler) createCheckpoint(ctx context.Context, req CreateCheckpointRequest) (*CreateCheckpointResponse, error) {
	if req.BranchID == "" {
		return nil, fmt.Errorf("branch_id is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	checkpoint, err := h.manager.CreateCheckpoint(ctx, req.BranchID, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("checkpoint creation failed: %w", err)
	}

	// Get branch to get counts
	branch, _ := h.storage.GetBranch(checkpoint.BranchID)
	thoughtCount := 0
	insightCount := 0
	if branch != nil {
		thoughtCount = len(branch.Thoughts)
		insightCount = len(branch.Insights)
	}

	return &CreateCheckpointResponse{
		CheckpointID: checkpoint.ID,
		Name:         checkpoint.Name,
		Description:  checkpoint.Description,
		BranchID:     checkpoint.BranchID,
		ThoughtCount: thoughtCount,
		InsightCount: insightCount,
		CreatedAt:    checkpoint.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// RestoreCheckpointRequest represents a restore request
type RestoreCheckpointRequest struct {
	CheckpointID string `json:"checkpoint_id"`
}

// RestoreCheckpointResponse represents the response
type RestoreCheckpointResponse struct {
	BranchID     string `json:"branch_id"`
	ThoughtCount int    `json:"thought_count"`
	InsightCount int    `json:"insight_count"`
	Message      string `json:"message"`
}

// HandleRestoreCheckpoint restores from a checkpoint
func (h *BacktrackingHandler) HandleRestoreCheckpoint(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req, err := unmarshalRequest[RestoreCheckpointRequest](params)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp, err := h.restoreCheckpoint(ctx, req)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// restoreCheckpoint is the typed internal implementation
func (h *BacktrackingHandler) restoreCheckpoint(ctx context.Context, req RestoreCheckpointRequest) (*RestoreCheckpointResponse, error) {
	if req.CheckpointID == "" {
		return nil, fmt.Errorf("checkpoint_id is required")
	}

	branch, err := h.manager.RestoreCheckpoint(ctx, req.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("checkpoint restoration failed: %w", err)
	}

	return &RestoreCheckpointResponse{
		BranchID:     branch.ID,
		ThoughtCount: len(branch.Thoughts),
		InsightCount: len(branch.Insights),
		Message:      "Checkpoint restored successfully",
	}, nil
}

// ListCheckpointsRequest represents a list request
type ListCheckpointsRequest struct {
	BranchID string `json:"branch_id,omitempty"`
}

// ListCheckpointsResponse represents the response
type ListCheckpointsResponse struct {
	Checkpoints []*CheckpointInfo `json:"checkpoints"`
	Count       int               `json:"count"`
}

// CheckpointInfo contains checkpoint information
type CheckpointInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	BranchID     string `json:"branch_id"`
	ThoughtCount int    `json:"thought_count"`
	CreatedAt    string `json:"created_at"`
}

// HandleListCheckpoints lists available checkpoints
func (h *BacktrackingHandler) HandleListCheckpoints(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req, err := unmarshalRequest[ListCheckpointsRequest](params)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp, err := h.listCheckpoints(ctx, req)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// listCheckpoints is the typed internal implementation
func (h *BacktrackingHandler) listCheckpoints(_ context.Context, req ListCheckpointsRequest) (*ListCheckpointsResponse, error) {
	checkpoints := h.manager.ListCheckpoints(req.BranchID)

	infos := make([]*CheckpointInfo, 0, len(checkpoints))
	for _, cp := range checkpoints {
		// Get branch for thought count
		branch, _ := h.storage.GetBranch(cp.BranchID)
		thoughtCount := 0
		if branch != nil {
			thoughtCount = len(branch.Thoughts)
		}

		info := &CheckpointInfo{
			ID:           cp.ID,
			Name:         cp.Name,
			Description:  cp.Description,
			BranchID:     cp.BranchID,
			ThoughtCount: thoughtCount,
			CreatedAt:    cp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		infos = append(infos, info)
	}

	return &ListCheckpointsResponse{
		Checkpoints: infos,
		Count:       len(infos),
	}, nil
}
