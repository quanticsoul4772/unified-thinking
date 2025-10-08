package handlers

import (
	"fmt"
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/processing"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// DualProcessHandler handles dual-process reasoning operations
type DualProcessHandler struct {
	executor *processing.DualProcessExecutor
	storage  storage.Storage
}

// NewDualProcessHandler creates a new dual-process handler
func NewDualProcessHandler(executor *processing.DualProcessExecutor, store storage.Storage) *DualProcessHandler {
	return &DualProcessHandler{
		executor: executor,
		storage:  store,
	}
}

// DualProcessThinkRequest represents a dual-process thinking request
type DualProcessThinkRequest struct {
	Content      string                 `json:"content"`
	Mode         string                 `json:"mode,omitempty"`
	BranchID     string                 `json:"branch_id,omitempty"`
	ForceSystem  string                 `json:"force_system,omitempty"` // "system1", "system2", or empty for auto
	KeyPoints    []string               `json:"key_points,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DualProcessThinkResponse represents the response
type DualProcessThinkResponse struct {
	ThoughtID        string                 `json:"thought_id"`
	SystemUsed       string                 `json:"system_used"`       // "system1" or "system2"
	Complexity       float64                `json:"complexity"`
	Escalated        bool                   `json:"escalated"`
	System1Time      string                 `json:"system1_time"`
	System2Time      string                 `json:"system2_time,omitempty"`
	Confidence       float64                `json:"confidence"`
	Content          string                 `json:"content"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// HandleDualProcessThink handles dual-process thinking requests
func (h *DualProcessHandler) HandleDualProcessThink(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	var req DualProcessThinkRequest
	if err := unmarshalParams(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request: " + err.Error())
	}

	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Default mode to linear if not specified
	mode := types.ModeLinear
	if req.Mode != "" {
		mode = types.ThinkingMode(req.Mode)
	}

	// Convert force_system from string to ProcessingSystem
	var forceSystem processing.ProcessingSystem
	if req.ForceSystem == "system1" {
		forceSystem = processing.System1
	} else if req.ForceSystem == "system2" {
		forceSystem = processing.System2
	}

	// Build processing request
	procReq := &processing.ProcessingRequest{
		Content:     req.Content,
		Mode:        mode,
		ForceSystem: forceSystem,
		KeyPoints:   req.KeyPoints,
		Confidence:  0.8, // Default
	}

	// Execute dual-process thinking
	result, err := h.executor.ProcessThought(ctx, procReq)
	if err != nil {
		return nil, fmt.Errorf("processing failed: " + err.Error())
	}

	// Convert system type to string
	systemUsed := "system1"
	if result.SystemUsed == processing.System2 {
		systemUsed = "system2"
	}

	// Build response
	resp := &DualProcessThinkResponse{
		ThoughtID:   result.Thought.ID,
		SystemUsed:  systemUsed,
		Complexity:  result.ComplexityScore,
		Escalated:   result.Escalated,
		System1Time: result.System1Time.String(),
		Confidence:  result.Thought.Confidence,
		Content:     result.Thought.Content,
		Metadata:    result.Thought.Metadata,
	}

	if result.System2Time > 0 {
		resp.System2Time = result.System2Time.String()
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}
