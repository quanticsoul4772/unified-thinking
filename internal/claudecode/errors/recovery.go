package errors

// RecoveryGenerator provides recovery suggestions for common error scenarios
type RecoveryGenerator struct {
	suggestions map[string][]string
	relatedTools map[string][]string
	examples    map[string]map[string]any
}

// NewRecoveryGenerator creates a new RecoveryGenerator with default suggestions
func NewRecoveryGenerator() *RecoveryGenerator {
	g := &RecoveryGenerator{
		suggestions:  make(map[string][]string),
		relatedTools: make(map[string][]string),
		examples:     make(map[string]map[string]any),
	}
	g.registerDefaults()
	return g
}

// registerDefaults sets up default recovery suggestions for all error codes
func (g *RecoveryGenerator) registerDefaults() {
	// Resource errors (1xxx)
	g.register(ErrThoughtNotFound,
		[]string{
			"Use 'search' tool to find thoughts matching your criteria",
			"Use 'history' tool to list recent thoughts",
			"Verify the thought_id format matches 'thought-XXXXXXXXXX-N'",
		},
		[]string{"search", "history", "list-branches"},
		map[string]any{"tool": "search", "params": map[string]any{"query": "your search terms", "limit": 10}},
	)

	g.register(ErrBranchNotFound,
		[]string{
			"Use 'list-branches' to see all available branches",
			"Check if the branch was deleted or merged",
			"Create a new branch with 'think' tool in tree mode",
		},
		[]string{"list-branches", "think", "recent-branches"},
		map[string]any{"tool": "list-branches", "params": map[string]any{}},
	)

	g.register(ErrSessionNotFound,
		[]string{
			"Use 'search-trajectories' to find existing sessions",
			"Start a new session with 'start-reasoning-session'",
			"Check if session was completed and archived",
		},
		[]string{"search-trajectories", "start-reasoning-session"},
		map[string]any{"tool": "start-reasoning-session", "params": map[string]any{
			"session_id": "new-session", "description": "Your session description",
		}},
	)

	g.register(ErrGraphNotFound,
		[]string{
			"Use 'got-initialize' to create a new graph",
			"Check if the graph_id is spelled correctly",
			"Graphs are not persisted across restarts",
		},
		[]string{"got-initialize", "got-get-state"},
		map[string]any{"tool": "got-initialize", "params": map[string]any{
			"graph_id": "my-graph", "initial_thought": "Starting thought",
		}},
	)

	g.register(ErrCheckpointNotFound,
		[]string{
			"Use 'list-checkpoints' to see available checkpoints",
			"Create a checkpoint with 'create-checkpoint'",
			"Checkpoints may have been pruned due to age",
		},
		[]string{"list-checkpoints", "create-checkpoint"},
		map[string]any{"tool": "list-checkpoints", "params": map[string]any{}},
	)

	g.register(ErrDecisionNotFound,
		[]string{
			"Create a new decision with 'make-decision'",
			"Check if the decision_id is correct",
		},
		[]string{"make-decision"},
		map[string]any{"tool": "make-decision", "params": map[string]any{
			"question": "Your decision question",
			"options": []map[string]any{{"name": "Option A"}, {"name": "Option B"}},
		}},
	)

	g.register(ErrPresetNotFound,
		[]string{
			"Use 'list-presets' to see available presets",
			"Check the preset name spelling",
			"Built-in presets: code-review, debug-analysis, architecture-decision, research-synthesis, refactoring-plan, test-strategy, documentation-gen, incident-investigation",
		},
		[]string{"list-presets"},
		map[string]any{"tool": "list-presets", "params": map[string]any{}},
	)

	// Validation errors (2xxx)
	g.register(ErrInvalidParameter,
		[]string{
			"Check the parameter type and format",
			"Refer to API documentation for valid values",
		},
		[]string{},
		nil,
	)

	g.register(ErrMissingRequired,
		[]string{
			"Add the required parameter to your request",
			"Check API documentation for required parameters",
		},
		[]string{},
		nil,
	)

	g.register(ErrInvalidMode,
		[]string{
			"Valid modes: linear, tree, divergent, reflection, backtracking, auto",
			"Use 'auto' for automatic mode selection",
		},
		[]string{"think"},
		map[string]any{"tool": "think", "params": map[string]any{
			"content": "Your thought content", "mode": "linear",
		}},
	)

	g.register(ErrInvalidConfidence,
		[]string{
			"Confidence must be between 0.0 and 1.0",
			"Common values: 0.5 (moderate), 0.8 (high), 0.95 (very high)",
		},
		[]string{},
		nil,
	)

	g.register(ErrInvalidFormat,
		[]string{
			"Valid formats: full, compact, minimal",
			"Use 'compact' for reduced token usage",
			"Use 'minimal' for essential fields only",
		},
		[]string{},
		nil,
	)

	// State errors (3xxx)
	g.register(ErrSessionActive,
		[]string{
			"Complete the current session with 'complete-reasoning-session'",
			"Or export and close it to start a new one",
		},
		[]string{"complete-reasoning-session"},
		map[string]any{"tool": "complete-reasoning-session", "params": map[string]any{
			"status": "completed", "solution": "Session solution",
		}},
	)

	g.register(ErrSessionNotActive,
		[]string{
			"Start a new session with 'start-reasoning-session'",
			"Import a previous session with 'import-session'",
		},
		[]string{"start-reasoning-session", "import-session"},
		nil,
	)

	g.register(ErrGraphFinalized,
		[]string{
			"Finalized graphs cannot be modified",
			"Create a new graph for further exploration",
			"Use 'got-get-state' to retrieve final conclusions",
		},
		[]string{"got-initialize", "got-get-state"},
		nil,
	)

	// External errors (4xxx)
	g.register(ErrEmbeddingFailed,
		[]string{
			"Wait 60 seconds and retry the operation",
			"Check VOYAGE_API_KEY environment variable",
			"Operation will continue with hash-based fallback",
		},
		[]string{"get-metrics"},
		nil,
	)

	g.register(ErrNeo4jConnection,
		[]string{
			"Check NEO4J_URI and credentials in environment",
			"Verify Neo4j server is running",
			"Knowledge graph features will be disabled",
		},
		[]string{"get-metrics"},
		nil,
	)

	g.register(ErrLLMFailed,
		[]string{
			"Check ANTHROPIC_API_KEY environment variable",
			"Verify API quota and rate limits",
			"Retry after a brief delay",
		},
		[]string{"get-metrics"},
		nil,
	)

	g.register(ErrStorageFailed,
		[]string{
			"Check SQLITE_PATH permissions",
			"Verify disk space availability",
			"Restart the server if issue persists",
		},
		[]string{"get-metrics"},
		nil,
	)

	// Limit errors (5xxx)
	g.register(ErrRateLimited,
		[]string{
			"Wait 60 seconds before retrying",
			"Use 'format: compact' to reduce API calls",
			"Consider batching operations",
		},
		[]string{"get-metrics"},
		nil,
	)

	g.register(ErrContextTooLarge,
		[]string{
			"Break the content into smaller chunks",
			"Use 'decompose-problem' for large problems",
			"Remove unnecessary details from input",
		},
		[]string{"decompose-problem"},
		nil,
	)

	g.register(ErrTooManyBranches,
		[]string{
			"Focus on the most promising branches",
			"Complete or abandon inactive branches",
			"Use 'list-branches' to review current branches",
		},
		[]string{"list-branches", "focus-branch"},
		nil,
	)

	g.register(ErrMaxDepthReached,
		[]string{
			"Synthesize current findings before going deeper",
			"Use 'got-aggregate' to combine parallel paths",
			"Consider if further depth adds value",
		},
		[]string{"synthesize-insights", "got-aggregate"},
		nil,
	)
}

// register adds recovery information for an error code
func (g *RecoveryGenerator) register(code string, suggestions []string, tools []string, example map[string]any) {
	g.suggestions[code] = suggestions
	g.relatedTools[code] = tools
	if example != nil {
		g.examples[code] = example
	}
}

// GetSuggestions returns recovery suggestions for an error code
func (g *RecoveryGenerator) GetSuggestions(code string) []string {
	if suggestions, ok := g.suggestions[code]; ok {
		return suggestions
	}
	return []string{"Check the error code and message for more details"}
}

// GetRelatedTools returns related tools for an error code
func (g *RecoveryGenerator) GetRelatedTools(code string) []string {
	if tools, ok := g.relatedTools[code]; ok {
		return tools
	}
	return nil
}

// GetExample returns an example fix for an error code
func (g *RecoveryGenerator) GetExample(code string) map[string]any {
	if example, ok := g.examples[code]; ok {
		return example
	}
	return nil
}

// Enhance adds recovery information to a StructuredError
func (g *RecoveryGenerator) Enhance(err *StructuredError) *StructuredError {
	if err == nil {
		return nil
	}

	// Only add suggestions if none exist
	if len(err.RecoverySuggestions) == 0 {
		err.RecoverySuggestions = g.GetSuggestions(err.Code)
	}

	// Only add related tools if none exist
	if len(err.RelatedTools) == 0 {
		err.RelatedTools = g.GetRelatedTools(err.Code)
	}

	// Only add example if none exists
	if err.ExampleFix == nil {
		err.ExampleFix = g.GetExample(err.Code)
	}

	return err
}

// DefaultGenerator is the default recovery generator instance
var DefaultGenerator = NewRecoveryGenerator()

// EnhanceError adds recovery information using the default generator
func EnhanceError(err *StructuredError) *StructuredError {
	return DefaultGenerator.Enhance(err)
}
