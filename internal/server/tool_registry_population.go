package server

import (
	"context"
	"fmt"
	"unified-thinking/internal/modes"
)

// populateToolRegistry registers safe tools for agentic use
func (s *UnifiedServer) populateToolRegistry(registry *modes.ToolRegistry) {
	// Register 30+ safe read-only tools
	safeTools := []string{
		"think", "validate", "search", "history", "decompose-problem",
		"make-decision", "analyze-perspectives", "analyze-temporal",
		"build-causal-graph", "generate-hypotheses", "evaluate-hypotheses",
		"detect-biases", "detect-fallacies", "detect-blind-spots",
		"self-evaluate", "assess-evidence", "detect-contradictions",
		"sensitivity-analysis", "prove", "check-syntax",
		"find-analogy", "apply-analogy", "decompose-argument",
		"search-similar-thoughts", "search-knowledge-graph",
		"probabilistic-reasoning", "got-initialize", "got-generate",
		"got-score", "got-aggregate", "synthesize-insights",
	}

	for _, name := range safeTools {
		_ = registry.Register(modes.ToolSpec{
			Name:        name,
			Description: fmt.Sprintf("Safe tool: %s", name),
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
				return map[string]interface{}{"status": "stub"}, nil
			},
		})
	}
}
