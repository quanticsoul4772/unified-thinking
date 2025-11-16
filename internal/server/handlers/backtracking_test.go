package handlers

import (
	"context"
	"testing"

	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// TestBacktrackingHandler_NewBacktrackingHandler tests handler creation
func TestBacktrackingHandler_NewBacktrackingHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	manager := modes.NewBacktrackingManager(store)

	handler := NewBacktrackingHandler(manager, store)

	if handler == nil {
		t.Fatal("NewBacktrackingHandler() should return a handler")
	}

	if handler.manager == nil {
		t.Error("Handler should have a manager")
	}

	if handler.storage == nil {
		t.Error("Handler should have storage")
	}
}

// TestBacktrackingHandler_HandleCreateCheckpoint tests checkpoint creation
func TestBacktrackingHandler_HandleCreateCheckpoint(t *testing.T) {
	// Create real storage and manager for testing
	store := storage.NewMemoryStorage()
	manager := modes.NewBacktrackingManager(store)
	handler := NewBacktrackingHandler(manager, store)

	// Create a test branch first
	branch := types.NewBranch().Build()
	_ = store.StoreBranch(branch)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid checkpoint",
			params: map[string]interface{}{
				"branch_id":   branch.ID,
				"name":        "Test Checkpoint",
				"description": "Test description",
			},
			wantErr: false,
		},
		{
			name: "valid checkpoint without description",
			params: map[string]interface{}{
				"branch_id": branch.ID,
				"name":      "Checkpoint No Desc",
			},
			wantErr: false,
		},
		{
			name: "missing branch_id",
			params: map[string]interface{}{
				"name": "Test Checkpoint",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			params: map[string]interface{}{
				"branch_id": branch.ID,
			},
			wantErr: true,
		},
		{
			name: "invalid branch_id",
			params: map[string]interface{}{
				"branch_id": "nonexistent",
				"name":      "Test Checkpoint",
			},
			wantErr: true,
		},
		{
			name:    "invalid params structure",
			params:  map[string]interface{}{"invalid": 123},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handler.HandleCreateCheckpoint(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCreateCheckpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandleCreateCheckpoint() should return result on success")
			}
		})
	}
}

// TestBacktrackingHandler_HandleRestoreCheckpoint tests checkpoint restoration
func TestBacktrackingHandler_HandleRestoreCheckpoint(t *testing.T) {
	store := storage.NewMemoryStorage()
	manager := modes.NewBacktrackingManager(store)
	handler := NewBacktrackingHandler(manager, store)

	// Create a test branch and checkpoint
	branch := types.NewBranch().Build()
	_ = store.StoreBranch(branch)

	ctx := context.Background()
	checkpoint, err := manager.CreateCheckpoint(ctx, branch.ID, "Test Checkpoint", "")
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid checkpoint restore",
			params: map[string]interface{}{
				"checkpoint_id": checkpoint.ID,
			},
			wantErr: false,
		},
		{
			name:    "missing checkpoint_id",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "invalid checkpoint_id",
			params: map[string]interface{}{
				"checkpoint_id": "nonexistent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.HandleRestoreCheckpoint(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleRestoreCheckpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandleRestoreCheckpoint() should return result on success")
			}
		})
	}
}

// TestBacktrackingHandler_HandleListCheckpoints tests checkpoint listing
func TestBacktrackingHandler_HandleListCheckpoints(t *testing.T) {
	store := storage.NewMemoryStorage()
	manager := modes.NewBacktrackingManager(store)
	handler := NewBacktrackingHandler(manager, store)

	// Create test branches and checkpoints
	branch1 := types.NewBranch().Build()
	branch2 := types.NewBranch().Build()
	_ = store.StoreBranch(branch1)
	_ = store.StoreBranch(branch2)

	ctx := context.Background()
	_, _ = manager.CreateCheckpoint(ctx, branch1.ID, "Checkpoint 1", "")
	_, _ = manager.CreateCheckpoint(ctx, branch1.ID, "Checkpoint 2", "")
	_, _ = manager.CreateCheckpoint(ctx, branch2.ID, "Checkpoint 3", "")

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "list all checkpoints",
			params:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "list checkpoints for specific branch",
			params: map[string]interface{}{
				"branch_id": branch1.ID,
			},
			wantErr: false,
		},
		{
			name: "list checkpoints for branch with no checkpoints",
			params: map[string]interface{}{
				"branch_id": "nonexistent",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.HandleListCheckpoints(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleListCheckpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandleListCheckpoints() should return result on success")
			}
		})
	}
}
