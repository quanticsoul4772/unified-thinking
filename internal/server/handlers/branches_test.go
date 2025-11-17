package handlers

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestNewBranchHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	tree := modes.NewTreeMode(store)

	handler := NewBranchHandler(store, tree)

	if handler == nil {
		t.Fatal("NewBranchHandler returned nil")
	}
	if handler.storage == nil {
		t.Error("storage not initialized")
	}
	if handler.tree == nil {
		t.Error("tree mode not initialized")
	}
}

func TestBranchHandler_HandleListBranches(t *testing.T) {
	store := storage.NewMemoryStorage()
	tree := modes.NewTreeMode(store)
	handler := NewBranchHandler(store, tree)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create some test branches
	branch1 := types.NewBranch().Build()
	branch2 := types.NewBranch().Build()
	_ = store.StoreBranch(branch1)
	_ = store.StoreBranch(branch2)

	result, response, err := handler.HandleListBranches(ctx, req, EmptyRequest{})

	if err != nil {
		t.Fatalf("HandleListBranches() error = %v", err)
	}

	if result == nil {
		t.Error("CallToolResult should not be nil")
	}

	if response == nil {
		t.Fatal("ListBranchesResponse should not be nil")
	}

	if response.Count != len(response.Branches) {
		t.Errorf("Count = %v, but len(Branches) = %v", response.Count, len(response.Branches))
	}

	if len(response.Branches) < 2 {
		t.Errorf("Expected at least 2 branches, got %v", len(response.Branches))
	}
}

func TestBranchHandler_HandleListBranches_Empty(t *testing.T) {
	store := storage.NewMemoryStorage()
	tree := modes.NewTreeMode(store)
	handler := NewBranchHandler(store, tree)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, response, err := handler.HandleListBranches(ctx, req, EmptyRequest{})

	if err != nil {
		t.Fatalf("HandleListBranches() error = %v", err)
	}

	if response.Count != 0 {
		t.Errorf("Count = %v, want 0", response.Count)
	}

	if len(response.Branches) != 0 {
		t.Errorf("Expected 0 branches, got %v", len(response.Branches))
	}

	if result == nil {
		t.Error("CallToolResult should not be nil")
	}
}

func TestBranchHandler_HandleFocusBranch(t *testing.T) {
	store := storage.NewMemoryStorage()
	tree := modes.NewTreeMode(store)
	handler := NewBranchHandler(store, tree)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create a test branch
	branch := types.NewBranch().Build()
	_ = store.StoreBranch(branch)

	tests := []struct {
		name       string
		input      FocusBranchRequest
		wantErr    bool
		wantStatus string
	}{
		{
			name: "focus valid branch",
			input: FocusBranchRequest{
				BranchID: branch.ID,
			},
			wantErr:    false,
			wantStatus: "success",
		},
		{
			name: "empty branch ID",
			input: FocusBranchRequest{
				BranchID: "",
			},
			wantErr: true,
		},
		{
			name: "non-existent branch",
			input: FocusBranchRequest{
				BranchID: "non-existent-id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, response, err := handler.HandleFocusBranch(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleFocusBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("FocusBranchResponse should not be nil")
				}
				if response.Status != tt.wantStatus {
					t.Errorf("Status = %v, want %v", response.Status, tt.wantStatus)
				}
				if response.ActiveBranchID != tt.input.BranchID {
					t.Errorf("ActiveBranchID = %v, want %v", response.ActiveBranchID, tt.input.BranchID)
				}

				// Verify branch is actually set as active
				activeBranch, _ := store.GetActiveBranch()
				if activeBranch == nil {
					t.Error("Active branch should be set")
				} else if activeBranch.ID != tt.input.BranchID {
					t.Errorf("Active branch ID = %v, want %v", activeBranch.ID, tt.input.BranchID)
				}
			}
		})
	}
}

func TestBranchHandler_HandleBranchHistory(t *testing.T) {
	store := storage.NewMemoryStorage()
	tree := modes.NewTreeMode(store)
	handler := NewBranchHandler(store, tree)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create a branch with thoughts
	branch := types.NewBranch().
		WithThought(types.NewThought().Content("Thought 1").Build()).
		WithThought(types.NewThought().Content("Thought 2").Build()).
		Build()
	_ = store.StoreBranch(branch)

	tests := []struct {
		name    string
		input   BranchHistoryRequest
		wantErr bool
	}{
		{
			name: "valid branch",
			input: BranchHistoryRequest{
				BranchID: branch.ID,
			},
			wantErr: false,
		},
		{
			name: "empty branch ID",
			input: BranchHistoryRequest{
				BranchID: "",
			},
			wantErr: true,
		},
		{
			name: "non-existent branch",
			input: BranchHistoryRequest{
				BranchID: "non-existent-id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, history, err := handler.HandleBranchHistory(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleBranchHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if history == nil {
					t.Fatal("BranchHistory should not be nil")
				}
				if history.BranchID != tt.input.BranchID {
					t.Errorf("BranchID = %v, want %v", history.BranchID, tt.input.BranchID)
				}
			}
		})
	}
}

func TestBranchHandler_HandleRecentBranches(t *testing.T) {
	store := storage.NewMemoryStorage()
	tree := modes.NewTreeMode(store)
	handler := NewBranchHandler(store, tree)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create and access some branches
	branch1 := types.NewBranch().Build()
	branch2 := types.NewBranch().Build()
	_ = store.StoreBranch(branch1)
	_ = store.StoreBranch(branch2)

	// Access them to populate recent list
	_ = store.SetActiveBranch(branch1.ID)
	_ = store.SetActiveBranch(branch2.ID)

	result, response, err := handler.HandleRecentBranches(ctx, req, EmptyRequest{})

	if err != nil {
		t.Fatalf("HandleRecentBranches() error = %v", err)
	}

	if result == nil {
		t.Error("CallToolResult should not be nil")
	}

	if response == nil {
		t.Fatal("RecentBranchesResponse should not be nil")
	}

	if response.Count != len(response.Branches) {
		t.Errorf("Count = %v, but len(Branches) = %v", response.Count, len(response.Branches))
	}

	if response.Branches == nil {
		t.Error("Branches should not be nil")
	}
}

func TestBranchHandler_ActiveBranchTracking(t *testing.T) {
	store := storage.NewMemoryStorage()
	tree := modes.NewTreeMode(store)
	handler := NewBranchHandler(store, tree)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create branches
	branch1 := types.NewBranch().Build()
	branch2 := types.NewBranch().Build()
	_ = store.StoreBranch(branch1)
	_ = store.StoreBranch(branch2)

	// Focus branch1
	_, _, _ = handler.HandleFocusBranch(ctx, req, FocusBranchRequest{BranchID: branch1.ID})

	// List branches should show branch1 as active
	_, _, _ = handler.HandleListBranches(ctx, req, EmptyRequest{})

	// Note: ListBranchesResponse doesn't have ActiveBranchID field in this handler
	// but we can verify through storage
	activeBranch, _ := store.GetActiveBranch()
	if activeBranch.ID != branch1.ID {
		t.Errorf("Active branch = %v, want %v", activeBranch.ID, branch1.ID)
	}

	// Focus branch2
	_, _, _ = handler.HandleFocusBranch(ctx, req, FocusBranchRequest{BranchID: branch2.ID})

	// Verify branch2 is now active
	activeBranch, _ = store.GetActiveBranch()
	if activeBranch.ID != branch2.ID {
		t.Errorf("Active branch = %v, want %v", activeBranch.ID, branch2.ID)
	}

	// List branches should have updated
	_, listResp, _ := handler.HandleListBranches(ctx, req, EmptyRequest{})
	if listResp.Count < 2 {
		t.Errorf("Expected at least 2 branches, got %v", listResp.Count)
	}
}

func TestBranchHandler_BranchLifecycle(t *testing.T) {
	store := storage.NewMemoryStorage()
	tree := modes.NewTreeMode(store)
	handler := NewBranchHandler(store, tree)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create branch
	branch := types.NewBranch().
		State(types.StateActive).
		Build()
	_ = store.StoreBranch(branch)

	// List should show it
	_, listResp, _ := handler.HandleListBranches(ctx, req, EmptyRequest{})  //nolint:errcheck // Test cleanup
	if listResp.Count == 0 {
		t.Error("Expected at least 1 branch")
	}

	// Focus it
	_, focusResp, _ := handler.HandleFocusBranch(ctx, req, FocusBranchRequest{BranchID: branch.ID})
	if focusResp.Status != "success" {
		t.Errorf("Focus status = %v, want success", focusResp.Status)
	}

	// Get history
	_, history, err := handler.HandleBranchHistory(ctx, req, BranchHistoryRequest{BranchID: branch.ID})
	if err != nil {
		t.Errorf("BranchHistory error = %v", err)
	}
	if history == nil {
		t.Error("History should not be nil")
	}

	// Check recent branches
	_, recentResp, _ := handler.HandleRecentBranches(ctx, req, EmptyRequest{})
	if recentResp.Count == 0 {
		t.Error("Expected at least 1 recent branch")
	}
}
