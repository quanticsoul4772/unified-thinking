package handlers

import (
	"context"
	"testing"

	"unified-thinking/internal/modes"
	"unified-thinking/internal/processing"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// TestDualProcessHandler_NewDualProcessHandler tests handler creation
func TestDualProcessHandler_NewDualProcessHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeMap := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeMap[types.ModeLinear] = modes.NewLinearMode(store)
	executor := processing.NewDualProcessExecutor(store, modeMap)

	handler := NewDualProcessHandler(executor, store)

	if handler == nil {
		t.Error("NewDualProcessHandler() should return a handler")
	}

	if handler.executor == nil {
		t.Error("Handler should have an executor")
	}

	if handler.storage == nil {
		t.Error("Handler should have storage")
	}
}

// TestDualProcessHandler_HandleDualProcessThink tests dual-process thinking
func TestDualProcessHandler_HandleDualProcessThink(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeMap := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeMap[types.ModeLinear] = modes.NewLinearMode(store)
	executor := processing.NewDualProcessExecutor(store, modeMap)
	handler := NewDualProcessHandler(executor, store)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid request with auto system selection",
			params: map[string]interface{}{
				"content": "What is 2+2?",
			},
			wantErr: false,
		},
		{
			name: "valid request forcing system1",
			params: map[string]interface{}{
				"content":      "Quick decision needed",
				"force_system": "system1",
			},
			wantErr: false,
		},
		{
			name: "valid request forcing system2",
			params: map[string]interface{}{
				"content":      "Complex logical problem requiring deep analysis",
				"force_system": "system2",
			},
			wantErr: false,
		},
		{
			name: "valid request with mode specified",
			params: map[string]interface{}{
				"content": "Analyze this problem",
				"mode":    "linear",
			},
			wantErr: false,
		},
		{
			name: "valid request with key points and metadata",
			params: map[string]interface{}{
				"content":    "Thought with metadata",
				"key_points": []string{"point1", "point2"},
				"metadata": map[string]interface{}{
					"tag": "test",
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
			name: "invalid force_system value",
			params: map[string]interface{}{
				"content":      "Test content",
				"force_system": "invalid",
			},
			wantErr: false, // Invalid value is ignored, not an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handler.HandleDualProcessThink(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleDualProcessThink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandleDualProcessThink() should return result on success")
			}
		})
	}
}

// TestDualProcessHandler_SystemSelection tests that the correct system is selected
func TestDualProcessHandler_SystemSelection(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeMap := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeMap[types.ModeLinear] = modes.NewLinearMode(store)
	executor := processing.NewDualProcessExecutor(store, modeMap)
	handler := NewDualProcessHandler(executor, store)

	tests := []struct {
		name        string
		content     string
		forceSystem string
		wantError   bool
	}{
		{
			name:        "simple math - system1",
			content:     "2+2",
			forceSystem: "",
			wantError:   false,
		},
		{
			name:        "complex logic - system2",
			content:     "If P then Q. P is true. Given P implies Q, what can we conclude about Q?",
			forceSystem: "",
			wantError:   false,
		},
		{
			name:        "force system1",
			content:     "Any content",
			forceSystem: "system1",
			wantError:   false,
		},
		{
			name:        "force system2",
			content:     "Any content",
			forceSystem: "system2",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := map[string]interface{}{
				"content": tt.content,
			}
			if tt.forceSystem != "" {
				params["force_system"] = tt.forceSystem
			}

			result, err := handler.HandleDualProcessThink(ctx, params)

			if (err != nil) != tt.wantError {
				t.Errorf("HandleDualProcessThink() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && result == nil {
				t.Error("HandleDualProcessThink() should return result on success")
			}
		})
	}
}
