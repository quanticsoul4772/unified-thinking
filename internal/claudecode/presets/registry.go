package presets

import (
	"fmt"
	"sync"
)

// Registry manages workflow presets
type Registry struct {
	presets map[string]*WorkflowPreset
	mu      sync.RWMutex
}

// NewRegistry creates a new preset registry with built-in presets
func NewRegistry() *Registry {
	r := &Registry{
		presets: make(map[string]*WorkflowPreset),
	}
	r.registerBuiltins()
	return r
}

// Register adds a preset to the registry
func (r *Registry) Register(preset *WorkflowPreset) error {
	if preset.ID == "" {
		return fmt.Errorf("preset ID is required")
	}
	if preset.Name == "" {
		return fmt.Errorf("preset name is required")
	}
	if len(preset.Steps) == 0 {
		return fmt.Errorf("preset must have at least one step")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.presets[preset.ID]; exists {
		return fmt.Errorf("preset %s already exists", preset.ID)
	}

	r.presets[preset.ID] = preset
	return nil
}

// Get retrieves a preset by ID
func (r *Registry) Get(id string) (*WorkflowPreset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	preset, ok := r.presets[id]
	if !ok {
		return nil, fmt.Errorf("preset not found: %s", id)
	}
	return preset, nil
}

// List returns all presets, optionally filtered by category
func (r *Registry) List(category string) []PresetSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summaries := make([]PresetSummary, 0, len(r.presets))
	for _, preset := range r.presets {
		if category == "" || preset.Category == category {
			summaries = append(summaries, preset.ToSummary())
		}
	}
	return summaries
}

// Categories returns all unique categories
func (r *Registry) Categories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make(map[string]bool)
	for _, preset := range r.presets {
		categories[preset.Category] = true
	}

	result := make([]string, 0, len(categories))
	for cat := range categories {
		result = append(result, cat)
	}
	return result
}

// Count returns the number of registered presets
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.presets)
}

// registerBuiltins registers all built-in presets
func (r *Registry) registerBuiltins() {
	// Code category
	_ = r.Register(CodeReviewPreset())
	_ = r.Register(DebugAnalysisPreset())
	_ = r.Register(RefactoringPlanPreset())

	// Architecture category
	_ = r.Register(ArchitectureDecisionPreset())

	// Research category
	_ = r.Register(ResearchSynthesisPreset())

	// Testing category
	_ = r.Register(TestStrategyPreset())

	// Documentation category
	_ = r.Register(DocumentationGenPreset())

	// Operations category
	_ = r.Register(IncidentInvestigationPreset())
}

// DefaultRegistry is the global preset registry
var DefaultRegistry = NewRegistry()

// GetPreset retrieves a preset from the default registry
func GetPreset(id string) (*WorkflowPreset, error) {
	return DefaultRegistry.Get(id)
}

// ListPresets lists presets from the default registry
func ListPresets(category string) []PresetSummary {
	return DefaultRegistry.List(category)
}
