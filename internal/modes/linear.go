package modes

import (
	"context"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// LinearMode implements sequential step-by-step reasoning
type LinearMode struct {
	storage storage.Storage
}

// NewLinearMode creates a new linear mode handler
func NewLinearMode(storage storage.Storage) *LinearMode {
	return &LinearMode{storage: storage}
}

// ProcessThought processes a thought in linear mode
func (m *LinearMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
	// If ChallengeAssumptions is true, modify content to question assumptions
	content := input.Content
	challengesAssumption := false

	if input.ChallengeAssumptions {
		// Prepend questioning language to encourage assumption challenging
		content = "Questioning assumptions: " + content +
			" (What assumptions are being made here? Are there alternative viewpoints? What evidence supports or contradicts this?)"
		challengesAssumption = true

		// Slightly reduce confidence when challenging assumptions
		if input.Confidence > 0.3 {
			input.Confidence -= 0.1
		}
	}

	thought := &types.Thought{
		Content:    content,
		Mode:       types.ModeLinear,
		Type:       input.Type,
		Confidence: input.Confidence,
		Timestamp:  time.Now(),
	}

	if input.ParentID != "" {
		thought.ParentID = input.ParentID
	}

	// Add metadata about assumption challenging
	if challengesAssumption {
		if thought.Metadata == nil {
			thought.Metadata = make(map[string]interface{})
		}
		thought.Metadata["challenged_assumptions"] = true
	}

	// Store the thought
	if err := m.storage.StoreThought(thought); err != nil {
		return nil, err
	}

	// Create simple state tracking
	result := &ThoughtResult{
		ThoughtID:            thought.ID,
		Mode:                 string(types.ModeLinear),
		Status:               "processed",
		Confidence:           thought.Confidence,
		ChallengesAssumption: challengesAssumption,
	}

	return result, nil
}

// GetHistory returns the linear history of thoughts
func (m *LinearMode) GetHistory(ctx context.Context) ([]*types.Thought, error) {
	thoughts := m.storage.SearchThoughts("", types.ModeLinear, 0, 0)
	return thoughts, nil
}
