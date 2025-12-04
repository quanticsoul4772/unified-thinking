// Package modes - ToolRegistry for programmatic tool calling
package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// ToolSpec defines a tool's interface for Claude
type ToolSpec struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
	Handler     ToolHandler            `json:"-"` // Internal handler reference
}

// ToolHandler executes a tool call
type ToolHandler func(ctx context.Context, input map[string]interface{}) (interface{}, error)

// ToolRegistry manages available tools for agentic use
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*ToolSpec
}

// NewToolRegistry creates an empty tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*ToolSpec),
	}
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(spec ToolSpec) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if spec.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if spec.Handler == nil {
		return fmt.Errorf("tool handler cannot be nil")
	}

	r.tools[spec.Name] = &spec
	return nil
}

// Unregister removes a tool from the registry
func (r *ToolRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (*ToolSpec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	spec, ok := r.tools[name]
	return spec, ok
}

// List returns all registered tool names
func (r *ToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GetToolsForClaude returns tools in Claude API format
func (r *ToolRegistry) GetToolsForClaude() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]map[string]interface{}, 0, len(r.tools))
	for _, spec := range r.tools {
		tools = append(tools, map[string]interface{}{
			"name":         spec.Name,
			"description":  spec.Description,
			"input_schema": spec.InputSchema,
		})
	}
	return tools
}

// Execute runs a tool by name with the given input
func (r *ToolRegistry) Execute(ctx context.Context, name string, input map[string]interface{}) (interface{}, error) {
	r.mu.RLock()
	spec, ok := r.tools[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	return spec.Handler(ctx, input)
}

// CreateFilteredRegistry creates a new registry with only the specified tools
func (r *ToolRegistry) CreateFilteredRegistry(allowedTools []string) *ToolRegistry {
	if len(allowedTools) == 0 {
		// Return a copy of all tools
		return r.Clone()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	allowed := make(map[string]bool)
	for _, name := range allowedTools {
		allowed[name] = true
	}

	filtered := NewToolRegistry()
	for name, spec := range r.tools {
		if allowed[name] {
			filtered.tools[name] = spec
		}
	}
	return filtered
}

// Clone creates a copy of the registry
func (r *ToolRegistry) Clone() *ToolRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := NewToolRegistry()
	for name, spec := range r.tools {
		clone.tools[name] = spec
	}
	return clone
}

// SafeToolSubset returns the list of tools safe for agentic use
// These tools do not have side effects or resource-intensive operations
var SafeToolSubset = []string{
	"think",                          // Process reasoning
	"search-similar-thoughts",        // Semantic search (read-only)
	"search-knowledge-graph",         // Knowledge retrieval (read-only)
	"build-causal-graph",             // Causal analysis
	"get-causal-graph",               // Retrieve causal graph (read-only)
	"generate-hypotheses",            // Abductive reasoning
	"evaluate-hypotheses",            // Hypothesis testing
	"analyze-perspectives",           // Multi-perspective analysis
	"detect-biases",                  // Bias detection
	"detect-fallacies",               // Fallacy detection
	"decompose-problem",              // Problem decomposition
	"make-decision",                  // Decision analysis
	"analyze-temporal",               // Temporal analysis
	"probabilistic-reasoning",        // Probabilistic reasoning
	"assess-evidence",                // Evidence assessment
	"detect-contradictions",          // Contradiction detection
	"self-evaluate",                  // Metacognitive analysis
	"detect-blind-spots",             // Unknown unknowns detection
	"dual-process-think",             // System 1/2 reasoning
	"verify-thought",                 // Hallucination detection
	"retrieve-similar-cases",         // Case-based reasoning (read-only)
	"decompose-argument",             // Argument analysis
	"generate-counter-arguments",     // Counter-argument generation
	"find-analogy",                   // Analogy finding
	"apply-analogy",                  // Analogy application
	"synthesize-insights",            // Cross-mode synthesis
	"detect-emergent-patterns",       // Pattern detection
	"got-generate",                   // Graph-of-Thoughts generation
	"got-aggregate",                  // Graph-of-Thoughts aggregation
	"got-refine",                     // Graph-of-Thoughts refinement
	"got-score",                      // Graph-of-Thoughts scoring
	"got-get-state",                  // Graph-of-Thoughts state (read-only)
}

// ExcludedTools lists tools that should NOT be available for agentic use
// These have side effects, are resource-intensive, or could cause recursion
var ExcludedTools = []string{
	// Side effects - write to storage
	"store-entity",
	"create-relationship",
	"start-reasoning-session",
	"complete-reasoning-session",
	// Session management
	"export-session",
	"import-session",
	// Could cause recursion
	"run-agent",
	"run-preset",
	"execute-workflow",
	// State modifying
	"got-initialize",
	"got-prune",
	"got-finalize",
	"create-checkpoint",
	"restore-checkpoint",
	// Resource intensive
	"embed-multimodal",
}

// BuildSchemaFromStruct creates a JSON schema from a struct definition
// This is a helper for building tool input schemas
func BuildSchemaFromStruct(description string, properties map[string]PropertyDef, required []string) map[string]interface{} {
	props := make(map[string]interface{})
	for name, def := range properties {
		prop := map[string]interface{}{
			"type":        def.Type,
			"description": def.Description,
		}
		if def.Default != nil {
			prop["default"] = def.Default
		}
		if def.Enum != nil {
			prop["enum"] = def.Enum
		}
		if def.Items != nil {
			prop["items"] = def.Items
		}
		props[name] = prop
	}

	return map[string]interface{}{
		"type":        "object",
		"description": description,
		"properties":  props,
		"required":    required,
	}
}

// PropertyDef defines a property in a JSON schema
type PropertyDef struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Items       interface{} `json:"items,omitempty"`
}

// CommonSchemas provides reusable JSON schemas for tool inputs
var CommonSchemas = struct {
	Think                   map[string]interface{}
	SearchSimilarThoughts   map[string]interface{}
	BuildCausalGraph        map[string]interface{}
	GenerateHypotheses      map[string]interface{}
	AnalyzePerspectives     map[string]interface{}
	DecomposeProblem        map[string]interface{}
	MakeDecision            map[string]interface{}
	DetectBiases            map[string]interface{}
	DetectFallacies         map[string]interface{}
	AssessEvidence          map[string]interface{}
	ProbabilisticReasoning  map[string]interface{}
	SynthesizeInsights      map[string]interface{}
}{
	Think: BuildSchemaFromStruct(
		"Process a thought with structured reasoning",
		map[string]PropertyDef{
			"content": {Type: "string", Description: "The thought content to process"},
			"mode":    {Type: "string", Description: "Thinking mode: linear, tree, divergent, auto", Default: "auto"},
		},
		[]string{"content"},
	),
	SearchSimilarThoughts: BuildSchemaFromStruct(
		"Find semantically similar thoughts",
		map[string]PropertyDef{
			"query": {Type: "string", Description: "Search query text"},
			"limit": {Type: "integer", Description: "Maximum results", Default: 10},
		},
		[]string{"query"},
	),
	BuildCausalGraph: BuildSchemaFromStruct(
		"Build a causal graph from observations",
		map[string]PropertyDef{
			"observations": {Type: "array", Description: "List of observations", Items: map[string]string{"type": "string"}},
			"context":      {Type: "string", Description: "Context for causal analysis"},
		},
		[]string{"observations"},
	),
	GenerateHypotheses: BuildSchemaFromStruct(
		"Generate explanatory hypotheses for observations",
		map[string]PropertyDef{
			"observations": {Type: "array", Description: "Observations to explain", Items: map[string]string{"type": "string"}},
			"constraints":  {Type: "array", Description: "Constraints on hypotheses", Items: map[string]string{"type": "string"}},
		},
		[]string{"observations"},
	),
	AnalyzePerspectives: BuildSchemaFromStruct(
		"Analyze a situation from multiple perspectives",
		map[string]PropertyDef{
			"situation":    {Type: "string", Description: "The situation to analyze"},
			"stakeholders": {Type: "array", Description: "Stakeholder perspectives to consider", Items: map[string]string{"type": "string"}},
		},
		[]string{"situation"},
	),
	DecomposeProblem: BuildSchemaFromStruct(
		"Decompose a complex problem into subproblems",
		map[string]PropertyDef{
			"problem":   {Type: "string", Description: "The problem to decompose"},
			"max_depth": {Type: "integer", Description: "Maximum decomposition depth", Default: 3},
		},
		[]string{"problem"},
	),
	MakeDecision: BuildSchemaFromStruct(
		"Analyze a decision with options and criteria",
		map[string]PropertyDef{
			"decision":    {Type: "string", Description: "The decision to make"},
			"options":     {Type: "array", Description: "Available options", Items: map[string]string{"type": "string"}},
			"criteria":    {Type: "array", Description: "Evaluation criteria", Items: map[string]string{"type": "string"}},
			"constraints": {Type: "array", Description: "Decision constraints", Items: map[string]string{"type": "string"}},
		},
		[]string{"decision", "options"},
	),
	DetectBiases: BuildSchemaFromStruct(
		"Detect cognitive biases in reasoning",
		map[string]PropertyDef{
			"content": {Type: "string", Description: "Content to analyze for biases"},
			"context": {Type: "string", Description: "Context of the reasoning"},
		},
		[]string{"content"},
	),
	DetectFallacies: BuildSchemaFromStruct(
		"Detect logical fallacies in arguments",
		map[string]PropertyDef{
			"argument": {Type: "string", Description: "The argument to analyze"},
			"context":  {Type: "string", Description: "Context of the argument"},
		},
		[]string{"argument"},
	),
	AssessEvidence: BuildSchemaFromStruct(
		"Assess the quality and reliability of evidence",
		map[string]PropertyDef{
			"claim":    {Type: "string", Description: "The claim being supported"},
			"evidence": {Type: "array", Description: "Evidence items to assess", Items: map[string]string{"type": "string"}},
		},
		[]string{"claim", "evidence"},
	),
	ProbabilisticReasoning: BuildSchemaFromStruct(
		"Perform probabilistic reasoning with beliefs",
		map[string]PropertyDef{
			"hypothesis":   {Type: "string", Description: "The hypothesis to evaluate"},
			"evidence":     {Type: "array", Description: "Evidence items", Items: map[string]string{"type": "string"}},
			"prior":        {Type: "number", Description: "Prior probability", Default: 0.5},
		},
		[]string{"hypothesis", "evidence"},
	),
	SynthesizeInsights: BuildSchemaFromStruct(
		"Synthesize insights across multiple analyses",
		map[string]PropertyDef{
			"inputs":  {Type: "array", Description: "Analysis inputs to synthesize", Items: map[string]string{"type": "string"}},
			"context": {Type: "string", Description: "Context for synthesis"},
		},
		[]string{"inputs"},
	),
}

// ToJSON serializes any value to JSON string
func ToJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(data)
}
