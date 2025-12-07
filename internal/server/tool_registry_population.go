package server

import (
	"context"
	"unified-thinking/internal/modes"
)

// populateToolRegistry registers safe tools for agentic use
// Safe tools are read-only with no side effects
func (s *UnifiedServer) populateToolRegistry(registry *modes.ToolRegistry) {
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
}
