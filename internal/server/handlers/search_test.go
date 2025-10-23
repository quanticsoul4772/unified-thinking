package handlers

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestNewSearchHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewSearchHandler(store)

	if handler == nil {
		t.Fatal("NewSearchHandler returned nil")
	}
	if handler.storage == nil {
		t.Error("storage not initialized")
	}
}

func TestSearchHandler_HandleSearch(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewSearchHandler(store)

	// Create some test thoughts
	thought1 := types.NewThought().
		Content("Test search functionality").
		Mode(types.ModeLinear).
		Build()
	thought2 := types.NewThought().
		Content("Another test thought").
		Mode(types.ModeTree).
		Build()
	thought3 := types.NewThought().
		Content("Different content here").
		Mode(types.ModeLinear).
		Build()

	_ = store.StoreThought(thought1)
	_ = store.StoreThought(thought2)
	_ = store.StoreThought(thought3)

	tests := []struct {
		name       string
		input      SearchRequest
		wantErr    bool
		wantMinLen int
	}{
		{
			name: "search all thoughts",
			input: SearchRequest{
				Query: "",
				Limit: 100,
			},
			wantErr:    false,
			wantMinLen: 3,
		},
		{
			name: "search with query",
			input: SearchRequest{
				Query: "test",
				Limit: 100,
			},
			wantErr:    false,
			wantMinLen: 0,
		},
		{
			name: "search with mode filter",
			input: SearchRequest{
				Query: "",
				Mode:  "linear",
				Limit: 100,
			},
			wantErr:    false,
			wantMinLen: 0,
		},
		{
			name: "search with limit",
			input: SearchRequest{
				Query: "",
				Limit: 2,
			},
			wantErr:    false,
			wantMinLen: 0,
		},
		{
			name: "search with offset",
			input: SearchRequest{
				Query:  "",
				Limit:  10,
				Offset: 1,
			},
			wantErr:    false,
			wantMinLen: 0,
		},
		{
			name: "search with default limit",
			input: SearchRequest{
				Query: "",
			},
			wantErr:    false,
			wantMinLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleSearch(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleSearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("SearchResponse should not be nil")
				}
				if response.Thoughts == nil {
					t.Error("Response.Thoughts should not be nil")
				}
				if response.Query != tt.input.Query {
					t.Errorf("Response.Query = %v, want %v", response.Query, tt.input.Query)
				}
				if response.Count != len(response.Thoughts) {
					t.Errorf("Response.Count = %v, but len(Thoughts) = %v", response.Count, len(response.Thoughts))
				}
				if len(response.Thoughts) < tt.wantMinLen {
					t.Errorf("Expected at least %v thoughts, got %v", tt.wantMinLen, len(response.Thoughts))
				}
			}
		})
	}
}

func TestSearchHandler_HandleSearch_EmptyStore(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewSearchHandler(store)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, response, err := handler.HandleSearch(ctx, req, SearchRequest{
		Query: "test",
		Limit: 100,
	})

	if err != nil {
		t.Fatalf("HandleSearch() error = %v", err)
	}

	if response.Count != 0 {
		t.Errorf("Count = %v, want 0", response.Count)
	}

	if len(response.Thoughts) != 0 {
		t.Errorf("Expected 0 thoughts, got %v", len(response.Thoughts))
	}

	if result == nil {
		t.Error("CallToolResult should not be nil")
	}
}

func TestSearchHandler_HandleGetMetrics(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewSearchHandler(store)

	// Create test data
	thought1 := types.NewThought().
		Content("Test 1").
		Mode(types.ModeLinear).
		Confidence(0.8).
		Build()
	thought2 := types.NewThought().
		Content("Test 2").
		Mode(types.ModeTree).
		Confidence(0.9).
		Build()
	thought3 := types.NewThought().
		Content("Test 3").
		Mode(types.ModeLinear).
		Confidence(0.85).
		Build()

	_ = store.StoreThought(thought1)
	_ = store.StoreThought(thought2)
	_ = store.StoreThought(thought3)

	branch := types.NewBranch().Build()
	_ = store.StoreBranch(branch)

	insight := types.NewInsight().
		Content("Test insight").
		Build()
	store.StoreInsight(insight)

	validation := &types.Validation{
		IsValid: true,
		Reason:  "Test validation",
	}
	store.StoreValidation(validation)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, response, err := handler.HandleGetMetrics(ctx, req, EmptyRequest{})

	if err != nil {
		t.Fatalf("HandleGetMetrics() error = %v", err)
	}

	if result == nil {
		t.Error("CallToolResult should not be nil")
	}

	if response == nil {
		t.Fatal("MetricsResponse should not be nil")
	}

	if response.TotalThoughts != 3 {
		t.Errorf("TotalThoughts = %v, want 3", response.TotalThoughts)
	}

	if response.TotalBranches != 1 {
		t.Errorf("TotalBranches = %v, want 1", response.TotalBranches)
	}

	if response.TotalInsights != 1 {
		t.Errorf("TotalInsights = %v, want 1", response.TotalInsights)
	}

	if response.TotalValidations != 1 {
		t.Errorf("TotalValidations = %v, want 1", response.TotalValidations)
	}

	if response.ThoughtsByMode == nil {
		t.Error("ThoughtsByMode should not be nil")
	}

	if response.AverageConfidence == 0 {
		t.Error("AverageConfidence should not be 0")
	}

	// Verify average confidence calculation
	expectedAvg := (0.8 + 0.9 + 0.85) / 3
	if response.AverageConfidence < expectedAvg-0.01 || response.AverageConfidence > expectedAvg+0.01 {
		t.Errorf("AverageConfidence = %v, want ~%v", response.AverageConfidence, expectedAvg)
	}
}

func TestSearchHandler_HandleGetMetrics_EmptyStore(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewSearchHandler(store)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, response, err := handler.HandleGetMetrics(ctx, req, EmptyRequest{})

	if err != nil {
		t.Fatalf("HandleGetMetrics() error = %v", err)
	}

	if response.TotalThoughts != 0 {
		t.Errorf("TotalThoughts = %v, want 0", response.TotalThoughts)
	}

	if response.TotalBranches != 0 {
		t.Errorf("TotalBranches = %v, want 0", response.TotalBranches)
	}

	if response.TotalInsights != 0 {
		t.Errorf("TotalInsights = %v, want 0", response.TotalInsights)
	}

	if response.TotalValidations != 0 {
		t.Errorf("TotalValidations = %v, want 0", response.TotalValidations)
	}

	if result == nil {
		t.Error("CallToolResult should not be nil")
	}
}

func TestSearchHandler_SearchPagination(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewSearchHandler(store)

	// Create many thoughts
	for i := 0; i < 10; i++ {
		thought := types.NewThought().
			Content("Thought " + string(rune('0'+i))).
			Mode(types.ModeLinear).
			Build()
		_ = store.StoreThought(thought)
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test pagination with limit
	_, resp1, _ := handler.HandleSearch(ctx, req, SearchRequest{
		Query:  "",
		Limit:  5,
		Offset: 0,
	})

	if len(resp1.Thoughts) > 5 {
		t.Errorf("Expected at most 5 thoughts, got %v", len(resp1.Thoughts))
	}

	// Test pagination with offset
	_, resp2, _ := handler.HandleSearch(ctx, req, SearchRequest{
		Query:  "",
		Limit:  5,
		Offset: 5,
	})

	if len(resp2.Thoughts) > 5 {
		t.Errorf("Expected at most 5 thoughts, got %v", len(resp2.Thoughts))
	}
}

func TestSearchHandler_ModeFiltering(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewSearchHandler(store)

	// Create thoughts in different modes
	linearThought := types.NewThought().
		Content("Linear thought").
		Mode(types.ModeLinear).
		Build()
	treeThought := types.NewThought().
		Content("Tree thought").
		Mode(types.ModeTree).
		Build()
	divergentThought := types.NewThought().
		Content("Divergent thought").
		Mode(types.ModeDivergent).
		Build()

	_ = store.StoreThought(linearThought)
	_ = store.StoreThought(treeThought)
	_ = store.StoreThought(divergentThought)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name string
		mode string
	}{
		{"filter linear", "linear"},
		{"filter tree", "tree"},
		{"filter divergent", "divergent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, response, err := handler.HandleSearch(ctx, req, SearchRequest{
				Query: "",
				Mode:  tt.mode,
				Limit: 100,
			})

			if err != nil {
				t.Fatalf("HandleSearch() error = %v", err)
			}

			// Verify all returned thoughts match the mode filter
			for _, thought := range response.Thoughts {
				if string(thought.Mode) != tt.mode {
					t.Errorf("Expected mode %v, got %v", tt.mode, thought.Mode)
				}
			}
		})
	}
}
