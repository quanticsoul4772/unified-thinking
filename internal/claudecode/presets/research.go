package presets

// ResearchSynthesisPreset returns the research-synthesis workflow preset
func ResearchSynthesisPreset() *WorkflowPreset {
	return &WorkflowPreset{
		ID:          "research-synthesis",
		Name:        "Research Synthesis",
		Description: "Deep research with Graph-of-Thoughts for exploring multiple paths and synthesizing findings",
		Category:    "research",
		InputSchema: map[string]ParamSpec{
			"topic": {
				Type:        "string",
				Required:    true,
				Description: "The research topic or question",
				Examples:    []any{"Best practices for microservices", "Comparison of state management solutions"},
			},
			"depth": {
				Type:        "string",
				Required:    false,
				Default:     "medium",
				Description: "Research depth (shallow, medium, deep)",
			},
			"focus_areas": {
				Type:        "array",
				Required:    false,
				Description: "Specific areas to focus on",
			},
		},
		OutputFormat:  "synthesis_report",
		EstimatedTime: "4-6 minutes",
		Tags:          []string{"research", "synthesis", "got"},
		Steps: []PresetStep{
			{
				StepID:      "init_graph",
				Tool:        "got-initialize",
				Description: "Start exploration graph",
				InputMap: map[string]string{
					"initial_thought": "topic",
				},
				StaticInputs: map[string]any{
					"graph_id": "research-${timestamp}",
					"config": map[string]any{
						"max_vertices": 50,
						"max_depth":    5,
					},
				},
				StoreAs: "graph",
			},
			{
				StepID:      "generate",
				Tool:        "got-generate",
				Description: "Generate multiple exploration paths",
				InputMap: map[string]string{
					"graph_id": "graph.graph_id",
					"problem":  "topic",
				},
				StaticInputs: map[string]any{
					"k": 3,
				},
				DependsOn: []string{"init_graph"},
				StoreAs:   "paths",
			},
			{
				StepID:      "aggregate",
				Tool:        "got-aggregate",
				Description: "Synthesize findings from all paths",
				InputMap: map[string]string{
					"graph_id": "graph.graph_id",
					"problem":  "topic",
				},
				DependsOn: []string{"generate"},
				StoreAs:   "synthesis",
			},
			{
				StepID:      "refine",
				Tool:        "got-refine",
				Description: "Polish the synthesis",
				InputMap: map[string]string{
					"graph_id":  "graph.graph_id",
					"vertex_id": "synthesis.vertex_id",
					"problem":   "topic",
				},
				DependsOn: []string{"aggregate"},
				StoreAs:   "refined",
			},
			{
				StepID:      "finalize",
				Tool:        "got-finalize",
				Description: "Extract conclusions",
				InputMap: map[string]string{
					"graph_id": "graph.graph_id",
				},
				DependsOn: []string{"refine"},
				StoreAs:   "conclusions",
			},
		},
	}
}
