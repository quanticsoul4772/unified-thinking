package handlers

import (
	"context"
	"testing"

	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/storage"
)

// TestUnknownUnknownsHandler_NewUnknownUnknownsHandler tests handler creation
func TestUnknownUnknownsHandler_NewUnknownUnknownsHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	detector := metacognition.NewUnknownUnknownsDetector()

	handler := NewUnknownUnknownsHandler(detector, store)

	if handler == nil {
		t.Error("NewUnknownUnknownsHandler() should return a handler")
	}

	if handler.detector == nil {
		t.Error("Handler should have a detector")
	}

	if handler.storage == nil {
		t.Error("Handler should have storage")
	}
}

// TestUnknownUnknownsHandler_HandleDetectBlindSpots tests blind spot detection
func TestUnknownUnknownsHandler_HandleDetectBlindSpots(t *testing.T) {
	store := storage.NewMemoryStorage()
	detector := metacognition.NewUnknownUnknownsDetector()
	handler := NewUnknownUnknownsHandler(detector, store)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid request with full context",
			params: map[string]interface{}{
				"content":     "We should launch the new product immediately to capture market share.",
				"domain":      "business",
				"context":     "Startup planning product launch",
				"assumptions": []string{"Market is ready", "Product is stable"},
				"confidence":  0.8,
			},
			wantErr: false,
		},
		{
			name: "valid request with minimal info",
			params: map[string]interface{}{
				"content": "Simple statement to analyze",
			},
			wantErr: false,
		},
		{
			name: "valid request with default confidence",
			params: map[string]interface{}{
				"content": "Decision without explicit confidence",
				"domain":  "technology",
			},
			wantErr: false,
		},
		{
			name: "valid request with assumptions",
			params: map[string]interface{}{
				"content": "We can scale infinitely",
				"assumptions": []string{
					"Unlimited resources",
					"No technical limitations",
				},
			},
			wantErr: false,
		},
		{
			name:    "missing content",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "empty content",
			params: map[string]interface{}{
				"content": "",
			},
			wantErr: true,
		},
		{
			name: "invalid params structure",
			params: map[string]interface{}{
				"content":     123,
				"assumptions": "not an array",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handler.HandleDetectBlindSpots(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleDetectBlindSpots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandleDetectBlindSpots() should return result on success")
			}
		})
	}
}

// TestUnknownUnknownsHandler_RiskLevelDetermination tests risk level calculation
func TestUnknownUnknownsHandler_RiskLevelDetermination(t *testing.T) {
	store := storage.NewMemoryStorage()
	detector := metacognition.NewUnknownUnknownsDetector()
	handler := NewUnknownUnknownsHandler(detector, store)

	tests := []struct {
		name    string
		content string
		domain  string
	}{
		{
			name:    "high confidence statement",
			content: "This is definitely the best approach",
			domain:  "strategy",
		},
		{
			name:    "absolute statement",
			content: "We will always succeed with this method",
			domain:  "business",
		},
		{
			name:    "reasonable statement",
			content: "This approach seems promising based on initial data",
			domain:  "research",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := map[string]interface{}{
				"content": tt.content,
				"domain":  tt.domain,
			}

			result, err := handler.HandleDetectBlindSpots(ctx, params)

			if err != nil {
				t.Errorf("HandleDetectBlindSpots() unexpected error: %v", err)
			}

			if result == nil {
				t.Error("HandleDetectBlindSpots() should return result")
			}
		})
	}
}

// TestUnknownUnknownsHandler_WithAssumptions tests blind spot detection with assumptions
func TestUnknownUnknownsHandler_WithAssumptions(t *testing.T) {
	store := storage.NewMemoryStorage()
	detector := metacognition.NewUnknownUnknownsDetector()
	handler := NewUnknownUnknownsHandler(detector, store)

	ctx := context.Background()
	params := map[string]interface{}{
		"content": "Our AI model will solve all customer problems automatically",
		"assumptions": []string{
			"AI is perfect",
			"Customers will adopt immediately",
			"No ethical concerns",
		},
		"confidence": 0.95,
	}

	result, err := handler.HandleDetectBlindSpots(ctx, params)

	if err != nil {
		t.Errorf("HandleDetectBlindSpots() unexpected error: %v", err)
	}

	if result == nil {
		t.Error("HandleDetectBlindSpots() should return result")
	}
}

// TestUnknownUnknownsHandler_DomainSpecificAnalysis tests domain-specific detection
func TestUnknownUnknownsHandler_DomainSpecificAnalysis(t *testing.T) {
	store := storage.NewMemoryStorage()
	detector := metacognition.NewUnknownUnknownsDetector()
	handler := NewUnknownUnknownsHandler(detector, store)

	domains := []string{"technology", "business", "science", "healthcare", "finance"}

	for _, domain := range domains {
		t.Run("domain_"+domain, func(t *testing.T) {
			ctx := context.Background()
			params := map[string]interface{}{
				"content": "We have a solid plan for implementation",
				"domain":  domain,
				"context": "Strategic planning session",
			}

			result, err := handler.HandleDetectBlindSpots(ctx, params)

			if err != nil {
				t.Errorf("HandleDetectBlindSpots() unexpected error for domain %s: %v", domain, err)
			}

			if result == nil {
				t.Errorf("HandleDetectBlindSpots() should return result for domain %s", domain)
			}
		})
	}
}

// TestUnknownUnknownsHandler_ConfidenceLevels tests different confidence levels
func TestUnknownUnknownsHandler_ConfidenceLevels(t *testing.T) {
	store := storage.NewMemoryStorage()
	detector := metacognition.NewUnknownUnknownsDetector()
	handler := NewUnknownUnknownsHandler(detector, store)

	confidenceLevels := []float64{0.1, 0.3, 0.5, 0.7, 0.9, 0.99}

	for _, conf := range confidenceLevels {
		t.Run("confidence_"+string(rune(conf*100)), func(t *testing.T) {
			ctx := context.Background()
			params := map[string]interface{}{
				"content":    "Analysis with varying confidence",
				"confidence": conf,
			}

			result, err := handler.HandleDetectBlindSpots(ctx, params)

			if err != nil {
				t.Errorf("HandleDetectBlindSpots() unexpected error at confidence %.2f: %v", conf, err)
			}

			if result == nil {
				t.Errorf("HandleDetectBlindSpots() should return result at confidence %.2f", conf)
			}
		})
	}
}
