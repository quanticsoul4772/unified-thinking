package modes

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// DivergentMode implements creative/rebellious ideation
type DivergentMode struct {
	storage *storage.MemoryStorage
}

// NewDivergentMode creates a new divergent mode handler
func NewDivergentMode(storage *storage.MemoryStorage) *DivergentMode {
	return &DivergentMode{storage: storage}
}

// ProcessThought generates creative/unconventional thoughts
func (m *DivergentMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
	// Generate creative thought
	creativeContent := m.generateCreativeThought(input.Content, input.ForceRebellion)

	thought := &types.Thought{
		Content:              creativeContent,
		Mode:                 types.ModeDivergent,
		Type:                 input.Type,
		Confidence:           input.Confidence,
		Timestamp:            time.Now(),
		IsRebellion:          input.ForceRebellion || rand.Float64() > 0.5,
		ChallengesAssumption: rand.Float64() > 0.3,
	}

	if input.PreviousThoughtID != "" {
		thought.ParentID = input.PreviousThoughtID
	}

	if err := m.storage.StoreThought(thought); err != nil {
		return nil, err
	}

	result := &ThoughtResult{
		ThoughtID:            thought.ID,
		Mode:                 string(types.ModeDivergent),
		Content:              creativeContent,
		IsRebellion:          thought.IsRebellion,
		ChallengesAssumption: thought.ChallengesAssumption,
	}

	return result, nil
}

func (m *DivergentMode) generateCreativeThought(problem string, forceRebellion bool) string {
	approaches := []string{
		fmt.Sprintf("What if we completely eliminated the concept of %s?", problem),
		fmt.Sprintf("Imagine if %s operated in reverse - what opportunities would that create?", problem),
		fmt.Sprintf("If we had infinite resources and no physical limitations, how would we solve %s?", problem),
		fmt.Sprintf("What if we combined %s with its exact opposite?", problem),
		fmt.Sprintf("How would an alien civilization with completely different logic solve %s?", problem),
		fmt.Sprintf("Challenge: Everyone assumes %s - what if it doesn't?", problem),
		fmt.Sprintf("The conventional wisdom about %s is probably wrong. Here's why:", problem),
		fmt.Sprintf("What if the 'problem' of %s is actually a feature, not a bug?", problem),
		fmt.Sprintf("Let's deliberately break every rule about %s and see what happens.", problem),
	}

	if forceRebellion {
		rebellious := []string{
			fmt.Sprintf("Everyone assumes %s - what if it doesn't?", problem),
			fmt.Sprintf("The conventional wisdom about %s is probably wrong. Here's why:", problem),
			fmt.Sprintf("What if the 'problem' of %s is actually a feature, not a bug?", problem),
			fmt.Sprintf("Let's deliberately break every rule about %s and see what happens.", problem),
		}
		return rebellious[rand.Intn(len(rebellious))]
	}

	return approaches[rand.Intn(len(approaches))]
}

// BranchThought creates a new creative branch from an existing thought
func (m *DivergentMode) BranchThought(ctx context.Context, thoughtID string, direction string) (*ThoughtResult, error) {
	sourceThought, err := m.storage.GetThought(thoughtID)
	if err != nil {
		return nil, err
	}

	branchedContent := m.generateBranchedThought(sourceThought, direction)

	thought := &types.Thought{
		Content:              branchedContent,
		Mode:                 types.ModeDivergent,
		ParentID:             thoughtID,
		Type:                 "branched_" + direction,
		Confidence:           0.7,
		Timestamp:            time.Now(),
		IsRebellion:          direction == "opposite",
		ChallengesAssumption: true,
	}

	if err := m.storage.StoreThought(thought); err != nil {
		return nil, err
	}

	result := &ThoughtResult{
		ThoughtID: thought.ID,
		Mode:      string(types.ModeDivergent),
		Content:   branchedContent,
		Direction: direction,
	}

	return result, nil
}

func (m *DivergentMode) generateBranchedThought(source *types.Thought, direction string) string {
	switch direction {
	case "more_extreme":
		return fmt.Sprintf("Taking it to the extreme: %s ... but multiplied by 1000x", source.Content)
	case "opposite":
		return fmt.Sprintf("Complete reversal: What if the exact opposite of '%s' is the real answer?", source.Content)
	case "tangential":
		return fmt.Sprintf("Unexpected connection: %s ... applied to a completely different domain", source.Content)
	case "simplified":
		return fmt.Sprintf("Radical simplification: What's the absolute simplest version of '%s'?", source.Content)
	case "combined":
		return fmt.Sprintf("Synthesis: Combine '%s' with something completely unrelated", source.Content)
	default:
		return fmt.Sprintf("New direction from: %s", source.Content)
	}
}

// ListThoughts returns all divergent thoughts
func (m *DivergentMode) ListThoughts(ctx context.Context) ([]*types.Thought, error) {
	thoughts := m.storage.SearchThoughts("", types.ModeDivergent)
	return thoughts, nil
}
