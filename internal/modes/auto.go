package modes

import (
	"context"
	"strings"

	"unified-thinking/internal/types"
)

// AutoMode implements automatic mode detection
type AutoMode struct {
	linear    *LinearMode
	tree      *TreeMode
	divergent *DivergentMode
}

// NewAutoMode creates a new auto mode detector
func NewAutoMode(linear *LinearMode, tree *TreeMode, divergent *DivergentMode) *AutoMode {
	return &AutoMode{
		linear:    linear,
		tree:      tree,
		divergent: divergent,
	}
}

// ProcessThought automatically selects the best mode and processes
func (m *AutoMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
	mode := m.detectMode(input)

	switch mode {
	case types.ModeLinear:
		return m.linear.ProcessThought(ctx, input)
	case types.ModeTree:
		return m.tree.ProcessThought(ctx, input)
	case types.ModeDivergent:
		return m.divergent.ProcessThought(ctx, input)
	default:
		return m.linear.ProcessThought(ctx, input)
	}
}

func (m *AutoMode) detectMode(input ThoughtInput) types.ThinkingMode {
	content := strings.ToLower(input.Content)

	// Divergent indicators
	divergentKeywords := []string{
		"creative", "unconventional", "what if", "imagine", "challenge", 
		"rebel", "outside the box", "innovative", "radical", "different",
	}
	for _, kw := range divergentKeywords {
		if strings.Contains(content, kw) {
			return types.ModeDivergent
		}
	}

	// Tree indicators
	if input.BranchID != "" || len(input.CrossRefs) > 0 || len(input.KeyPoints) > 0 {
		return types.ModeTree
	}

	treeKeywords := []string{
		"branch", "explore", "alternative", "parallel", "compare",
		"multiple", "options", "possibilities",
	}
	for _, kw := range treeKeywords {
		if strings.Contains(content, kw) {
			return types.ModeTree
		}
	}

	// Default to linear
	return types.ModeLinear
}
