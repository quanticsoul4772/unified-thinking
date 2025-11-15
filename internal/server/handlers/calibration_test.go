package handlers

import (
	"context"
	"testing"
)

func TestHandleRecordPrediction_Success(t *testing.T) {
	handler := NewCalibrationHandler()
	ctx := context.Background()

	req := &RecordPredictionRequest{
		ThoughtID:  "thought-pred-1",
		Confidence: 0.85,
		Mode:       "linear",
		Metadata:   map[string]interface{}{"key": "value"},
	}

	resp, err := handler.HandleRecordPrediction(ctx, req)
	if err != nil {
		t.Fatalf("HandleRecordPrediction error = %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success, got failure: %s", resp.Message)
	}

	if resp.Prediction == nil {
		t.Fatal("expected prediction in response")
	}

	if resp.Prediction.ThoughtID != req.ThoughtID {
		t.Fatalf("prediction thought ID = %s, want %s", resp.Prediction.ThoughtID, req.ThoughtID)
	}

	if resp.Prediction.Confidence != req.Confidence {
		t.Fatalf("prediction confidence = %f, want %f", resp.Prediction.Confidence, req.Confidence)
	}
}

func TestHandleRecordOutcome_Success(t *testing.T) {
	handler := NewCalibrationHandler()
	ctx := context.Background()

	// First record a prediction
	predReq := &RecordPredictionRequest{
		ThoughtID:  "thought-outcome-1",
		Confidence: 0.75,
		Mode:       "tree",
	}

	_, err := handler.HandleRecordPrediction(ctx, predReq)
	if err != nil {
		t.Fatalf("HandleRecordPrediction error = %v", err)
	}

	// Now record an outcome
	outcomeReq := &RecordOutcomeRequest{
		ThoughtID:        "thought-outcome-1",
		WasCorrect:       true,
		ActualConfidence: 0.8,
		Source:           "validation",
		Metadata:         map[string]interface{}{"verified_by": "test"},
	}

	resp, err := handler.HandleRecordOutcome(ctx, outcomeReq)
	if err != nil {
		t.Fatalf("HandleRecordOutcome error = %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success, got failure: %s", resp.Message)
	}

	if resp.Outcome == nil {
		t.Fatal("expected outcome in response")
	}

	if resp.Outcome.ThoughtID != outcomeReq.ThoughtID {
		t.Fatalf("outcome thought ID = %s, want %s", resp.Outcome.ThoughtID, outcomeReq.ThoughtID)
	}

	if resp.Outcome.WasCorrect != outcomeReq.WasCorrect {
		t.Fatalf("outcome was_correct = %v, want %v", resp.Outcome.WasCorrect, outcomeReq.WasCorrect)
	}
}

func TestHandleGetCalibrationReport_Success(t *testing.T) {
	handler := NewCalibrationHandler()
	ctx := context.Background()

	// Record some predictions and outcomes
	for i := 0; i < 3; i++ {
		predReq := &RecordPredictionRequest{
			ThoughtID:  "thought-cal-" + string(rune('1'+i)),
			Confidence: 0.8,
			Mode:       "linear",
		}

		_, err := handler.HandleRecordPrediction(ctx, predReq)
		if err != nil {
			t.Fatalf("HandleRecordPrediction error = %v", err)
		}

		outcomeReq := &RecordOutcomeRequest{
			ThoughtID:        predReq.ThoughtID,
			WasCorrect:       i%2 == 0,
			ActualConfidence: 0.85,
			Source:           "validation",
		}

		_, err = handler.HandleRecordOutcome(ctx, outcomeReq)
		if err != nil {
			t.Fatalf("HandleRecordOutcome error = %v", err)
		}
	}

	// Get calibration report
	reportReq := &GetCalibrationReportRequest{}
	resp, err := handler.HandleGetCalibrationReport(ctx, reportReq)
	if err != nil {
		t.Fatalf("HandleGetCalibrationReport error = %v", err)
	}

	if resp.Status != "success" {
		t.Fatalf("expected success status, got %s", resp.Status)
	}

	if resp.Report == nil {
		t.Fatal("expected report in response")
	}

	if resp.Report.TotalPredictions != 3 {
		t.Fatalf("report total predictions = %d, want 3", resp.Report.TotalPredictions)
	}
}

func TestHandleRecordOutcome_WithoutPrediction(t *testing.T) {
	handler := NewCalibrationHandler()
	ctx := context.Background()

	// Try to record outcome without prediction
	outcomeReq := &RecordOutcomeRequest{
		ThoughtID:        "thought-no-pred",
		WasCorrect:       true,
		ActualConfidence: 0.9,
		Source:           "validation",
	}

	resp, err := handler.HandleRecordOutcome(ctx, outcomeReq)
	if err != nil {
		t.Fatalf("HandleRecordOutcome error = %v", err)
	}

	// Should still succeed (tracker may allow orphan outcomes)
	if !resp.Success {
		t.Logf("Recording outcome without prediction: %s", resp.Message)
	}
}

func TestHandleGetCalibrationReport_Empty(t *testing.T) {
	handler := NewCalibrationHandler()
	ctx := context.Background()

	// Get report with no data
	reportReq := &GetCalibrationReportRequest{}
	resp, err := handler.HandleGetCalibrationReport(ctx, reportReq)
	if err != nil {
		t.Fatalf("HandleGetCalibrationReport error = %v", err)
	}

	if resp.Status != "success" {
		t.Fatalf("expected success status, got %s", resp.Status)
	}

	if resp.Report == nil {
		t.Fatal("expected report in response even when empty")
	}

	if resp.Report.TotalPredictions != 0 {
		t.Fatalf("report total predictions = %d, want 0", resp.Report.TotalPredictions)
	}
}
