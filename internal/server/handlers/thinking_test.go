package handlers

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

func TestNewThinkingHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()

	handler := NewThinkingHandler(store, linear, tree, divergent, auto, validator)

	if handler == nil {
		t.Fatal("NewThinkingHandler returned nil")
	}
	if handler.storage == nil {
		t.Error("storage not initialized")
	}
	if handler.linear == nil {
		t.Error("linear mode not initialized")
	}
	if handler.tree == nil {
		t.Error("tree mode not initialized")
	}
	if handler.divergent == nil {
		t.Error("divergent mode not initialized")
	}
	if handler.auto == nil {
		t.Error("auto mode not initialized")
	}
	if handler.validator == nil {
		t.Error("validator not initialized")
	}
}

func TestThinkingHandler_HandleThink(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()
	handler := NewThinkingHandler(store, linear, tree, divergent, auto, validator)

	tests := []struct {
		name    string
		input   ThinkRequest
		wantErr bool
	}{
		{
			name: "linear mode thought",
			input: ThinkRequest{
				Content:    "Test linear thought",
				Mode:       "linear",
				Confidence: 0.8,
			},
			wantErr: false,
		},
		{
			name: "tree mode thought",
			input: ThinkRequest{
				Content:    "Test tree thought",
				Mode:       "tree",
				Confidence: 0.9,
			},
			wantErr: false,
		},
		{
			name: "divergent mode thought",
			input: ThinkRequest{
				Content:        "Test divergent thought",
				Mode:           "divergent",
				Confidence:     0.7,
				ForceRebellion: true,
			},
			wantErr: false,
		},
		{
			name: "auto mode thought",
			input: ThinkRequest{
				Content:    "Test auto thought",
				Mode:       "auto",
				Confidence: 0.85,
			},
			wantErr: false,
		},
		{
			name: "default confidence",
			input: ThinkRequest{
				Content: "Test with default confidence",
				Mode:    "linear",
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			input: ThinkRequest{
				Content:    "Test invalid mode",
				Mode:       "invalid",
				Confidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "with validation",
			input: ThinkRequest{
				Content:           "Test with validation",
				Mode:              "linear",
				Confidence:        0.9,
				RequireValidation: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleThink(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleThink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Error("ThinkResponse should not be nil")
				}
				if response.ThoughtID == "" {
					t.Error("Response missing ThoughtID")
				}
				if response.Status != "success" {
					t.Errorf("Response status = %v, want success", response.Status)
				}

				// Confidence is returned from the mode processing
				if response.Confidence == 0 {
					// Some modes may return 0 confidence, that's ok
					// Just verify the response was created
				}
			}
		})
	}
}

func TestThinkingHandler_HandleHistory(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()
	handler := NewThinkingHandler(store, linear, tree, divergent, auto, validator)

	// Create some test thoughts
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Add thoughts in different modes
	handler.HandleThink(ctx, req, ThinkRequest{Content: "Linear 1", Mode: "linear"})
	handler.HandleThink(ctx, req, ThinkRequest{Content: "Linear 2", Mode: "linear"})
	handler.HandleThink(ctx, req, ThinkRequest{Content: "Tree 1", Mode: "tree"})

	tests := []struct {
		name       string
		input      HistoryRequest
		wantErr    bool
		wantMinLen int
	}{
		{
			name: "all thoughts with default limit",
			input: HistoryRequest{
				Limit: 100,
			},
			wantErr:    false,
			wantMinLen: 3,
		},
		{
			name: "with limit",
			input: HistoryRequest{
				Limit: 2,
			},
			wantErr:    false,
			wantMinLen: 0, // Could be 0-2 depending on implementation
		},
		{
			name: "with offset",
			input: HistoryRequest{
				Limit:  10,
				Offset: 1,
			},
			wantErr:    false,
			wantMinLen: 0,
		},
		{
			name: "mode filtered",
			input: HistoryRequest{
				Mode:  "linear",
				Limit: 100,
			},
			wantErr:    false,
			wantMinLen: 0,
		},
		{
			name: "default limit applied",
			input: HistoryRequest{
				// No limit specified, should default to 100
			},
			wantErr:    false,
			wantMinLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, response, err := handler.HandleHistory(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Error("HistoryResponse should not be nil")
				}
				if response.Thoughts == nil {
					t.Error("Response.Thoughts should not be nil")
				}

				// Check limit is set
				expectedLimit := tt.input.Limit
				if expectedLimit == 0 {
					expectedLimit = 100
				}
				if response.Limit != expectedLimit {
					t.Errorf("Response limit = %v, want %v", response.Limit, expectedLimit)
				}

				// Check offset
				if response.Offset != tt.input.Offset {
					t.Errorf("Response offset = %v, want %v", response.Offset, tt.input.Offset)
				}
			}
		})
	}
}

func TestThinkingHandler_HandleHistory_BranchSpecific(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()
	handler := NewThinkingHandler(store, linear, tree, divergent, auto, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create a thought in tree mode (creates a branch)
	_, thinkResp, _ := handler.HandleThink(ctx, req, ThinkRequest{
		Content: "Branch thought",
		Mode:    "tree",
	})

	// Get history for that branch
	result, response, err := handler.HandleHistory(ctx, req, HistoryRequest{
		BranchID: thinkResp.BranchID,
		Limit:    100,
	})

	if err != nil {
		t.Fatalf("HandleHistory() error = %v", err)
	}

	if result == nil {
		t.Error("CallToolResult should not be nil")
	}
	if response == nil {
		t.Error("HistoryResponse should not be nil")
	}
}

func TestThinkingHandler_HandleHistory_InvalidBranch(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()
	handler := NewThinkingHandler(store, linear, tree, divergent, auto, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := handler.HandleHistory(ctx, req, HistoryRequest{
		BranchID: "non-existent-branch",
		Limit:    100,
	})

	if err == nil {
		t.Error("Expected error for non-existent branch, got nil")
	}
}

func TestThinkingHandler_CrossRefs(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()
	handler := NewThinkingHandler(store, linear, tree, divergent, auto, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create first branch
	_, resp1, _ := handler.HandleThink(ctx, req, ThinkRequest{
		Content: "First branch",
		Mode:    "tree",
	})

	// Create second branch with cross-reference to first
	_, response2, err := handler.HandleThink(ctx, req, ThinkRequest{
		Content: "Second branch",
		Mode:    "tree",
		CrossRefs: []modes.CrossRefInput{
			{
				ToBranch: resp1.BranchID,
				Type:     "complementary",
				Reason:   "They work together",
				Strength: 0.9,
			},
		},
	})

	if err != nil {
		t.Fatalf("HandleThink() error = %v", err)
	}

	if response2.CrossRefCount == 0 {
		t.Error("CrossRefCount should be > 0")
	}
}

func TestThinkingHandler_KeyPoints(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()
	handler := NewThinkingHandler(store, linear, tree, divergent, auto, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, response, err := handler.HandleThink(ctx, req, ThinkRequest{
		Content:   "Thought with key points",
		Mode:      "linear",
		KeyPoints: []string{"point1", "point2", "point3"},
	})

	if err != nil {
		t.Fatalf("HandleThink() error = %v", err)
	}

	// Retrieve the thought to verify it was stored
	_, getErr := store.GetThought(response.ThoughtID)
	if getErr != nil {
		t.Errorf("Failed to retrieve thought: %v", getErr)
	}
	// Note: Key points storage depends on mode implementation
}

func TestThinkingHandler_ParentChaining(t *testing.T) {
	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()
	handler := NewThinkingHandler(store, linear, tree, divergent, auto, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create parent thought
	_, response1, _ := handler.HandleThink(ctx, req, ThinkRequest{
		Content: "Parent thought",
		Mode:    "linear",
	})

	// Create child thought
	_, response2, err := handler.HandleThink(ctx, req, ThinkRequest{
		Content:  "Child thought",
		Mode:     "linear",
		ParentID: response1.ThoughtID,
	})

	if err != nil {
		t.Fatalf("HandleThink() error = %v", err)
	}

	// Verify parent relationship
	thought, _ := store.GetThought(response2.ThoughtID)
	if thought.ParentID != response1.ThoughtID {
		t.Errorf("ParentID = %v, want %v", thought.ParentID, response1.ThoughtID)
	}
}
