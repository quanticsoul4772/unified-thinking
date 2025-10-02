package modes

import (
	"context"
	"fmt"
	"sync"

	"unified-thinking/internal/types"
)

// ThinkingMode represents a cognitive thinking pattern
type ThinkingMode interface {
	// ProcessThought processes input and returns result
	ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error)

	// Name returns the mode identifier
	Name() types.ThinkingMode

	// CanHandle determines if this mode can process the input
	CanHandle(input ThoughtInput) bool
}

// BranchingMode extends ThinkingMode for modes that support branching
type BranchingMode interface {
	ThinkingMode
	ListBranches(ctx context.Context) ([]*types.Branch, error)
	SetActiveBranch(ctx context.Context, branchID string) error
}

// DivergentProcessor extends ThinkingMode for divergent thinking operations
type DivergentProcessor interface {
	ThinkingMode
	BranchThought(ctx context.Context, thoughtID string, branchType string) (*ThoughtResult, error)
}

// Registry manages available thinking modes
type Registry struct {
	mu    sync.RWMutex
	modes map[types.ThinkingMode]ThinkingMode
}

// NewRegistry creates a new mode registry
func NewRegistry() *Registry {
	return &Registry{
		modes: make(map[types.ThinkingMode]ThinkingMode),
	}
}

// Register adds a mode to the registry
func (r *Registry) Register(mode ThinkingMode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := mode.Name()
	if _, exists := r.modes[name]; exists {
		return fmt.Errorf("mode already registered: %s", name)
	}

	r.modes[name] = mode
	return nil
}

// Get retrieves a mode by name
func (r *Registry) Get(name types.ThinkingMode) (ThinkingMode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mode, exists := r.modes[name]
	if !exists {
		return nil, fmt.Errorf("unknown mode: %s", name)
	}
	return mode, nil
}

// SelectBest chooses the best mode for given input
// This is used by auto mode or for intelligent mode selection
func (r *Registry) SelectBest(input ThoughtInput) (ThinkingMode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Priority order: divergent > tree > linear
	// This ensures more specialized modes are checked first

	// Check divergent mode
	if mode, exists := r.modes[types.ModeDivergent]; exists && mode.CanHandle(input) {
		return mode, nil
	}

	// Check tree mode
	if mode, exists := r.modes[types.ModeTree]; exists && mode.CanHandle(input) {
		return mode, nil
	}

	// Check linear mode
	if mode, exists := r.modes[types.ModeLinear]; exists && mode.CanHandle(input) {
		return mode, nil
	}

	// Fallback to linear if available
	if mode, exists := r.modes[types.ModeLinear]; exists {
		return mode, nil
	}

	return nil, fmt.Errorf("no suitable mode found")
}

// Available returns all registered mode names
func (r *Registry) Available() []types.ThinkingMode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]types.ThinkingMode, 0, len(r.modes))
	for name := range r.modes {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered modes
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.modes)
}

// Update existing modes to implement the ThinkingMode interface methods

// Name returns the mode identifier for LinearMode
func (m *LinearMode) Name() types.ThinkingMode {
	return types.ModeLinear
}

// CanHandle determines if linear mode can process the input
func (m *LinearMode) CanHandle(input ThoughtInput) bool {
	// Linear handles simple, sequential reasoning
	// It's the default fallback mode
	return true
}

// Name returns the mode identifier for TreeMode
func (m *TreeMode) Name() types.ThinkingMode {
	return types.ModeTree
}

// CanHandle determines if tree mode can process the input
func (m *TreeMode) CanHandle(input ThoughtInput) bool {
	// Tree mode is preferred when:
	// 1. Branch ID is explicitly provided
	// 2. Key points are provided (for parallel exploration)
	// 3. Cross-references are present
	// 4. Content suggests branching/exploration

	if input.BranchID != "" {
		return true
	}

	if len(input.KeyPoints) > 0 {
		return true
	}

	if len(input.CrossRefs) > 0 {
		return true
	}

	// Check for tree-mode keywords
	treeKeywords := []string{"branch", "explore", "alternative", "parallel", "option"}
	for _, keyword := range treeKeywords {
		if containsIgnoreCase(input.Content, keyword) {
			return true
		}
	}

	return false
}

// Name returns the mode identifier for DivergentMode
func (m *DivergentMode) Name() types.ThinkingMode {
	return types.ModeDivergent
}

// CanHandle determines if divergent mode can process the input
func (m *DivergentMode) CanHandle(input ThoughtInput) bool {
	// Divergent mode is preferred when:
	// 1. Force rebellion is requested
	// 2. Challenge assumptions is requested
	// 3. Content suggests creative/unconventional thinking

	if input.ForceRebellion {
		return true
	}

	// Check for divergent-mode keywords
	divergentKeywords := []string{
		"creative", "unconventional", "what if", "imagine",
		"challenge", "rebel", "different", "innovative",
		"brainstorm", "unusual", "unique",
	}

	for _, keyword := range divergentKeywords {
		if containsIgnoreCase(input.Content, keyword) {
			return true
		}
	}

	return false
}

// Name returns the mode identifier for AutoMode
func (m *AutoMode) Name() types.ThinkingMode {
	return types.ModeAuto
}

// CanHandle determines if auto mode can process the input
func (m *AutoMode) CanHandle(input ThoughtInput) bool {
	// Auto mode can handle anything - it delegates to other modes
	return true
}

// Helper function for case-insensitive string contains
func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return contains(sLower, substrLower)
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
