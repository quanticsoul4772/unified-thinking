package presets

// DocumentationGenPreset returns the documentation-gen workflow preset
func DocumentationGenPreset() *WorkflowPreset {
	return &WorkflowPreset{
		ID:          "documentation-gen",
		Name:        "Documentation Generation",
		Description: "Generate documentation from code analysis with multi-audience consideration and structured output",
		Category:    "documentation",
		InputSchema: map[string]ParamSpec{
			"code": {
				Type:        "string",
				Required:    true,
				Description: "Code or module to document",
				Examples:    []any{"API endpoints", "Utility functions", "Data models"},
			},
			"doc_type": {
				Type:        "string",
				Required:    false,
				Default:     "api",
				Description: "Documentation type (api, readme, guide, inline)",
			},
			"audiences": {
				Type:        "array",
				Required:    false,
				Default:     []string{"developers", "users"},
				Description: "Target audiences",
			},
		},
		OutputFormat:  "documentation",
		EstimatedTime: "2-3 minutes",
		Tags:          []string{"documentation", "api", "readme"},
		Steps: []PresetStep{
			{
				StepID:      "analyze",
				Tool:        "think",
				Description: "Analyze code structure and purpose",
				InputMap: map[string]string{
					"content": "code",
				},
				StaticInputs: map[string]any{
					"mode":       "linear",
					"confidence": 0.8,
				},
				StoreAs: "analysis",
			},
			{
				StepID:      "perspectives",
				Tool:        "analyze-perspectives",
				Description: "Consider different audience needs",
				InputMap: map[string]string{
					"topic":        "code",
					"stakeholders": "audiences",
				},
				DependsOn: []string{"analyze"},
				StoreAs:   "perspectives",
			},
			{
				StepID:      "structure",
				Tool:        "decompose-argument",
				Description: "Structure documentation outline",
				InputMap: map[string]string{
					"content": "code",
				},
				StaticInputs: map[string]any{
					"type": "documentation_structure",
				},
				DependsOn: []string{"perspectives"},
				StoreAs:   "outline",
			},
			{
				StepID:      "synthesize",
				Tool:        "synthesize-insights",
				Description: "Combine perspectives into coherent docs",
				StaticInputs: map[string]any{
					"mode": "documentation",
				},
				DependsOn: []string{"structure"},
				StoreAs:   "synthesis",
			},
			{
				StepID:      "generate",
				Tool:        "think",
				Description: "Generate final documentation",
				InputMap: map[string]string{
					"content": "code",
				},
				StaticInputs: map[string]any{
					"mode":       "linear",
					"confidence": 0.85,
					"key_points": []string{"usage examples", "parameters", "return values"},
				},
				DependsOn: []string{"synthesize"},
				StoreAs:   "documentation",
			},
		},
	}
}
