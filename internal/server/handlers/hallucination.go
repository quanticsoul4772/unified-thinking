// Package handlers provides MCP tool handlers for hallucination detection.
package handlers

import (
	"context"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

// HallucinationHandler handles hallucination detection requests
type HallucinationHandler struct {
	storage  storage.Storage
	detector *validation.HallucinationDetector
}

// NewHallucinationHandler creates a new hallucination handler
func NewHallucinationHandler(store storage.Storage) *HallucinationHandler {
	return &HallucinationHandler{
		storage:  store,
		detector: validation.NewHallucinationDetector(),
	}
}

// VerifyThoughtRequest is the request for thought verification
type VerifyThoughtRequest struct {
	ThoughtID         string `json:"thought_id"`
	VerificationLevel string `json:"verification_level,omitempty"` // "fast", "deep", "hybrid"
}

// VerifyThoughtResponse contains the verification results
type VerifyThoughtResponse struct {
	Report *validation.HallucinationReport `json:"report"`
	Status string                          `json:"status"`
}

// GetReportRequest requests a hallucination report
type GetReportRequest struct {
	ThoughtID string `json:"thought_id"`
}

// HandleVerifyThought verifies a thought for hallucinations
func (h *HallucinationHandler) HandleVerifyThought(ctx context.Context, request *VerifyThoughtRequest) (*VerifyThoughtResponse, error) {
	// Retrieve the thought
	thought, err := h.storage.GetThought(request.ThoughtID)
	if err != nil {
		return nil, err
	}

	// Verify the thought
	report, err := h.detector.VerifyThought(ctx, thought)
	if err != nil {
		return nil, err
	}

	response := &VerifyThoughtResponse{
		Report: report,
		Status: "success",
	}

	return response, nil
}

// HandleGetReport retrieves a cached hallucination report
func (h *HallucinationHandler) HandleGetReport(ctx context.Context, request *GetReportRequest) (*VerifyThoughtResponse, error) {
	report, err := h.detector.GetReport(request.ThoughtID)
	if err != nil {
		return nil, err
	}

	response := &VerifyThoughtResponse{
		Report: report,
		Status: "success",
	}

	return response, nil
}
