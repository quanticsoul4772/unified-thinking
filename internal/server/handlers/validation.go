package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

// ValidationHandler handles validation operations
type ValidationHandler struct {
	storage   storage.Storage
	validator *validation.LogicValidator
}

// NewValidationHandler creates a new validation handler
func NewValidationHandler(store storage.Storage, validator *validation.LogicValidator) *ValidationHandler {
	return &ValidationHandler{
		storage:   store,
		validator: validator,
	}
}

// ValidateRequest represents a validation request
type ValidateRequest struct {
	ThoughtID string `json:"thought_id"`
}

// ValidateResponse represents a validation response
type ValidateResponse struct {
	ValidationID string `json:"validation_id"`
	IsValid      bool   `json:"is_valid"`
	Reason       string `json:"reason,omitempty"`
}

// ProveRequest represents a prove request
type ProveRequest struct {
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
}

// ProveResponse represents a prove response
type ProveResponse struct {
	IsProvable bool     `json:"is_provable"`
	Steps      []string `json:"steps,omitempty"`
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
}

// CheckSyntaxRequest represents a check syntax request
type CheckSyntaxRequest struct {
	Statements []string `json:"statements"`
}

// CheckSyntaxResponse represents a check syntax response
type CheckSyntaxResponse struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// HandleValidate validates a thought
func (h *ValidationHandler) HandleValidate(ctx context.Context, req *mcp.CallToolRequest, input ValidateRequest) (*mcp.CallToolResult, *ValidateResponse, error) {
	if input.ThoughtID == "" {
		return nil, nil, fmt.Errorf("thought_id is required")
	}

	thought, err := h.storage.GetThought(input.ThoughtID)
	if err != nil {
		return nil, nil, err
	}

	validationResult, err := h.validator.ValidateThought(thought)
	if err != nil {
		return nil, nil, err
	}

	response := &ValidateResponse{
		ValidationID: validationResult.ID,
		IsValid:      validationResult.IsValid,
		Reason:       validationResult.Reason,
	}

	return &mcp.CallToolResult{}, response, nil
}

// HandleProve attempts to prove a conclusion from premises
func (h *ValidationHandler) HandleProve(ctx context.Context, req *mcp.CallToolRequest, input ProveRequest) (*mcp.CallToolResult, *ProveResponse, error) {
	result := h.validator.Prove(input.Premises, input.Conclusion)

	response := &ProveResponse{
		IsProvable: result.IsProvable,
		Steps:      result.Steps,
		Premises:   input.Premises,
		Conclusion: input.Conclusion,
	}

	return &mcp.CallToolResult{}, response, nil
}

// HandleCheckSyntax checks logical syntax
func (h *ValidationHandler) HandleCheckSyntax(ctx context.Context, req *mcp.CallToolRequest, input CheckSyntaxRequest) (*mcp.CallToolResult, *CheckSyntaxResponse, error) {
	isValid, errors := h.validator.CheckSyntax(input.Statements)

	response := &CheckSyntaxResponse{
		IsValid: isValid,
		Errors:  errors,
	}

	return &mcp.CallToolResult{}, response, nil
}
