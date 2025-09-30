package modes

import (
	"context"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// LinearMode implements sequential step-by-step reasoning
type LinearMode struct {
	storage *storage.MemoryStorage
}

// NewLinearMode creates a new linear mode handler
func NewLinearMode(storage *storage.MemoryStorage) *LinearMode {
	return &LinearMode{storage: storage}
}

// ProcessThought processes a thought in linear mode
func (m *LinearMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
	thought := &types.Thought{
		Content:    input.Content,
		Mode:       types.ModeLinear,
		Type:       input.Type,
		Confidence: input.Confidence,
		Timestamp:  time.Now(),
	}

	if input.ParentID != "" {
		thought.ParentID = input.ParentID
	}

	// Store the thought
	if err := m.storage.StoreThought(thought); err != nil {
		return nil, err
	}

	// Create simple state tracking
	result := &ThoughtResult{
		ThoughtID:  thought.ID,
		Mode:       string(types.ModeLinear),
		Status:     "processed",
		Confidence: thought.Confidence,
	}

	return result, nil
}

// GetHistory returns the linear history of thoughts
func (m *LinearMode) GetHistory(ctx context.Context) ([]*types.Thought, error) {
	thoughts := m.storage.SearchThoughts("", types.ModeLinear)
	return thoughts, nil
}
