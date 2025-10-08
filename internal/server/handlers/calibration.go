// Package handlers provides MCP tool handlers for confidence calibration.
package handlers

import (
	"context"

	"unified-thinking/internal/validation"
)

// CalibrationHandler handles confidence calibration requests
type CalibrationHandler struct {
	tracker *validation.CalibrationTracker
}

// NewCalibrationHandler creates a new calibration handler
func NewCalibrationHandler() *CalibrationHandler {
	return &CalibrationHandler{
		tracker: validation.NewCalibrationTracker(),
	}
}

// RecordPredictionRequest is the request for recording a prediction
type RecordPredictionRequest struct {
	ThoughtID  string                 `json:"thought_id"`
	Confidence float64                `json:"confidence"`
	Mode       string                 `json:"mode"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RecordPredictionResponse contains the result of recording a prediction
type RecordPredictionResponse struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Prediction *validation.Prediction `json:"prediction,omitempty"`
}

// HandleRecordPrediction records a confidence prediction
func (h *CalibrationHandler) HandleRecordPrediction(ctx context.Context, request *RecordPredictionRequest) (*RecordPredictionResponse, error) {
	prediction := &validation.Prediction{
		ThoughtID:  request.ThoughtID,
		Confidence: request.Confidence,
		Mode:       request.Mode,
		Metadata:   request.Metadata,
	}

	err := h.tracker.RecordPrediction(prediction)
	if err != nil {
		return &RecordPredictionResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &RecordPredictionResponse{
		Success:    true,
		Message:    "Prediction recorded successfully",
		Prediction: prediction,
	}, nil
}

// RecordOutcomeRequest is the request for recording an outcome
type RecordOutcomeRequest struct {
	ThoughtID        string                 `json:"thought_id"`
	WasCorrect       bool                   `json:"was_correct"`
	ActualConfidence float64                `json:"actual_confidence"`
	Source           string                 `json:"source"` // "validation", "verification", "user_feedback"
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RecordOutcomeResponse contains the result of recording an outcome
type RecordOutcomeResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Outcome *validation.Outcome   `json:"outcome,omitempty"`
}

// HandleRecordOutcome records an outcome for a prediction
func (h *CalibrationHandler) HandleRecordOutcome(ctx context.Context, request *RecordOutcomeRequest) (*RecordOutcomeResponse, error) {
	outcome := &validation.Outcome{
		ThoughtID:        request.ThoughtID,
		WasCorrect:       request.WasCorrect,
		ActualConfidence: request.ActualConfidence,
		Source:           validation.OutcomeSource(request.Source),
		Metadata:         request.Metadata,
	}

	err := h.tracker.RecordOutcome(outcome)
	if err != nil {
		return &RecordOutcomeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &RecordOutcomeResponse{
		Success: true,
		Message: "Outcome recorded successfully",
		Outcome: outcome,
	}, nil
}

// GetCalibrationReportRequest is the request for getting a calibration report
type GetCalibrationReportRequest struct {
	// No parameters needed - returns overall report
}

// GetCalibrationReportResponse contains the calibration report
type GetCalibrationReportResponse struct {
	Report *validation.CalibrationReport `json:"report"`
	Status string                         `json:"status"`
}

// HandleGetCalibrationReport generates and returns a calibration report
func (h *CalibrationHandler) HandleGetCalibrationReport(ctx context.Context, request *GetCalibrationReportRequest) (*GetCalibrationReportResponse, error) {
	report := h.tracker.GetCalibrationReport()

	return &GetCalibrationReportResponse{
		Report: report,
		Status: "success",
	}, nil
}
