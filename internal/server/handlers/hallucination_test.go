package handlers

import (
	"context"
	"testing"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestHandleVerifyThought_Success(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewHallucinationHandler(store)

	thought := &types.Thought{
		ID:      "thought-verify-1",
		Content: "The capital of France is Paris. Paris is located on the Seine River.",
		Mode:    types.ModeLinear,
	}

	if err := store.StoreThought(thought); err != nil {
		t.Fatalf("StoreThought error = %v", err)
	}

	ctx := context.Background()
	req := &VerifyThoughtRequest{
		ThoughtID:         thought.ID,
		VerificationLevel: "fast",
	}

	resp, err := handler.HandleVerifyThought(ctx, req)
	if err != nil {
		t.Fatalf("HandleVerifyThought error = %v", err)
	}

	if resp.Status != "success" {
		t.Fatalf("expected success status, got %s", resp.Status)
	}

	if resp.Report == nil {
		t.Fatal("expected report in response")
	}

	if resp.Report.ThoughtID != thought.ID {
		t.Fatalf("report thought ID = %s, want %s", resp.Report.ThoughtID, thought.ID)
	}
}

func TestHandleVerifyThought_ThoughtNotFound(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewHallucinationHandler(store)

	ctx := context.Background()
	req := &VerifyThoughtRequest{
		ThoughtID:         "nonexistent-thought",
		VerificationLevel: "fast",
	}

	_, err := handler.HandleVerifyThought(ctx, req)
	if err == nil {
		t.Fatal("expected error for nonexistent thought")
	}
}

func TestHandleGetReport_Success(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewHallucinationHandler(store)

	thought := &types.Thought{
		ID:      "thought-report-1",
		Content: "Test content for report retrieval",
		Mode:    types.ModeLinear,
	}

	if err := store.StoreThought(thought); err != nil {
		t.Fatalf("StoreThought error = %v", err)
	}

	ctx := context.Background()

	// First verify to generate a report
	verifyReq := &VerifyThoughtRequest{
		ThoughtID:         thought.ID,
		VerificationLevel: "fast",
	}

	_, err := handler.HandleVerifyThought(ctx, verifyReq)
	if err != nil {
		t.Fatalf("HandleVerifyThought error = %v", err)
	}

	// Now retrieve the report
	getReq := &GetReportRequest{
		ThoughtID: thought.ID,
	}

	resp, err := handler.HandleGetReport(ctx, getReq)
	if err != nil {
		t.Fatalf("HandleGetReport error = %v", err)
	}

	if resp.Status != "success" {
		t.Fatalf("expected success status, got %s", resp.Status)
	}

	if resp.Report == nil {
		t.Fatal("expected report in response")
	}

	if resp.Report.ThoughtID != thought.ID {
		t.Fatalf("report thought ID = %s, want %s", resp.Report.ThoughtID, thought.ID)
	}
}

func TestHandleGetReport_NotFound(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewHallucinationHandler(store)

	ctx := context.Background()
	req := &GetReportRequest{
		ThoughtID: "no-report-exists",
	}

	_, err := handler.HandleGetReport(ctx, req)
	if err == nil {
		t.Fatal("expected error when report not found")
	}
}
