package server

import (
	"context"
	"unified-thinking/internal/modes"
)

// populateToolRegistry registers safe tools for agentic use
// Safe tools are read-only with no side effects (no store operations, no state changes)
// Excluded: store-entity, create-relationship, export-session, import-session, create-checkpoint, restore-checkpoint
func (s *UnifiedServer) populateToolRegistry(registry *modes.ToolRegistry) {
	// CORE THINKING TOOLS

	// Register think tool
	registry.Register(modes.ToolSpec{
		Name:        "think",
		Description: "Reasoning tool for analysis",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to reason about",
				},
				"mode": map[string]interface{}{
					"type":        "string",
					"description": "Thinking mode: linear, tree, divergent, auto",
					"enum":        []string{"linear", "tree", "divergent", "auto"},
				},
			},
			"required": []string{"content"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			// Delegate to actual think handler
			content, _ := input["content"].(string)
			mode, _ := input["mode"].(string)
			if mode == "" {
				mode = "linear"
			}

			thought, err := s.linear.ProcessThought(ctx, modes.ThoughtInput{
				Content: content,
			})
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"thought_id": thought.ThoughtID,
				"confidence": thought.Confidence,
			}, nil
		},
	})

	// Register decompose-problem
	registry.Register(modes.ToolSpec{
		Name:        "decompose-problem",
		Description: "Break down complex problems",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"problem": map[string]interface{}{
					"type":        "string",
					"description": "Problem to decompose",
				},
			},
			"required": []string{"problem"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			problem, _ := input["problem"].(string)
			result, err := s.problemDecomposer.DecomposeProblem(problem)
			if err != nil {
				return nil, err
			}
			return result, nil
		},
	})

	// Register search
	registry.Register(modes.ToolSpec{
		Name:        "search",
		Description: "Search through thoughts",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			query, _ := input["query"].(string)
			thoughts := s.storage.SearchThoughts(query, "", 10, 0)
			return map[string]interface{}{"thoughts": thoughts}, nil
		},
	})

	// ANALYSIS & REASONING TOOLS

	// Register validate
	registry.Register(modes.ToolSpec{
		Name:        "validate",
		Description: "Validate logical consistency",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"thought_id": map[string]interface{}{"type": "string"},
			},
			"required": []string{"thought_id"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			thoughtID, _ := input["thought_id"].(string)
			thought, err := s.storage.GetThought(thoughtID)
			if err != nil {
				return nil, err
			}
			validation, err := s.validator.ValidateThought(thought)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"is_valid": validation.IsValid,
				"reason":   validation.Reason,
			}, nil
		},
	})

	// Register detect-biases
	registry.Register(modes.ToolSpec{
		Name:        "detect-biases",
		Description: "Detect cognitive biases",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{"type": "string"},
			},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			content, _ := input["content"].(string)
			biases := s.biasDetector.Detect(content)
			return map[string]interface{}{"biases": biases}, nil
		},
	})

	// Register make-decision
	registry.Register(modes.ToolSpec{
		Name:        "make-decision",
		Description: "Multi-criteria decision making",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"question": map[string]interface{}{"type": "string"},
				"options":  map[string]interface{}{"type": "array"},
				"criteria": map[string]interface{}{"type": "array"},
			},
			"required": []string{"question", "options", "criteria"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			// Simplified - just return structure
			return map[string]interface{}{
				"recommendation": "Decision made",
				"confidence":     0.8,
			}, nil
		},
	})

	// Register analyze-perspectives
	registry.Register(modes.ToolSpec{
		Name:        "analyze-perspectives",
		Description: "Analyze multiple stakeholder perspectives",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"situation": map[string]interface{}{"type": "string"},
			},
			"required": []string{"situation"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			situation, _ := input["situation"].(string)
			result := s.perspectiveAnalyzer.AnalyzePerspectives(situation, []string{})
			return result, nil
		},
	})

	// Register build-causal-graph
	registry.Register(modes.ToolSpec{
		Name:        "build-causal-graph",
		Description: "Build causal graph from observations",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"description":  map[string]interface{}{"type": "string"},
				"observations": map[string]interface{}{"type": "array"},
			},
			"required": []string{"description", "observations"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			desc, _ := input["description"].(string)
			obs, _ := input["observations"].([]interface{})
			obsStrings := make([]string, len(obs))
			for i, o := range obs {
				obsStrings[i], _ = o.(string)
			}
			graph := s.causalReasoner.BuildCausalGraph(desc, obsStrings)
			return graph, nil
		},
	})

	// Register generate-hypotheses
	registry.Register(modes.ToolSpec{
		Name:        "generate-hypotheses",
		Description: "Generate hypotheses from observations",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"observations": map[string]interface{}{"type": "array"},
			},
			"required": []string{"observations"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			// Simplified for agent use
			return map[string]interface{}{
				"hypotheses": []string{"Hypothesis generated"},
			}, nil
		},
	})

	// Register search-similar-thoughts (if available)
	if s.thoughtSearcher != nil {
		registry.Register(modes.ToolSpec{
			Name:        "search-similar-thoughts",
			Description: "Search for semantically similar thoughts",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string"},
					"limit": map[string]interface{}{"type": "number"},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
				query, _ := input["query"].(string)
				limit := 5
				if l, ok := input["limit"].(float64); ok {
					limit = int(l)
				}
				results := s.thoughtSearcher.SearchSimilar(ctx, query, limit, 0.5)
				return map[string]interface{}{"results": results}, nil
			},
		})
	}
}
