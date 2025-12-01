package presets

// RefactoringPlanPreset returns the refactoring-plan workflow preset
func RefactoringPlanPreset() *WorkflowPreset {
	return &WorkflowPreset{
		ID:          "refactoring-plan",
		Name:        "Refactoring Plan",
		Description: "Safe refactoring with risk assessment, dependency mapping, and phased execution plan",
		Category:    "code",
		InputSchema: map[string]ParamSpec{
			"target": {
				Type:        "string",
				Required:    true,
				Description: "Code or component to refactor",
				Examples:    []any{"Authentication module", "Data access layer"},
			},
			"goals": {
				Type:        "array",
				Required:    true,
				Description: "Refactoring goals",
				Examples:    []any{[]string{"Improve testability", "Reduce coupling"}},
			},
			"constraints": {
				Type:        "array",
				Required:    false,
				Description: "Constraints or limitations",
			},
		},
		OutputFormat:  "phased_plan",
		EstimatedTime: "3-4 minutes",
		Tags:          []string{"refactoring", "planning", "risk"},
		Steps: []PresetStep{
			{
				StepID:      "decompose",
				Tool:        "decompose-problem",
				Description: "Break refactoring into phases",
				InputMap: map[string]string{
					"problem": "target",
				},
				StaticInputs: map[string]any{
					"strategy": "phases",
				},
				StoreAs: "phases",
			},
			{
				StepID:      "dependencies",
				Tool:        "build-causal-graph",
				Description: "Map code dependencies",
				InputMap: map[string]string{
					"description": "target",
				},
				StaticInputs: map[string]any{
					"type": "dependency",
				},
				DependsOn: []string{"decompose"},
				StoreAs:   "deps",
			},
			{
				StepID:      "risk",
				Tool:        "sensitivity-analysis",
				Description: "Assess refactoring risks",
				InputMap: map[string]string{
					"target": "target",
				},
				StaticInputs: map[string]any{
					"factors": []string{"test_coverage", "dependencies", "complexity"},
				},
				DependsOn: []string{"dependencies"},
				StoreAs:   "risks",
			},
			{
				StepID:      "timing",
				Tool:        "analyze-temporal",
				Description: "Plan execution timing",
				InputMap: map[string]string{
					"topic": "target",
				},
				StaticInputs: map[string]any{
					"focus": "execution_order",
				},
				DependsOn: []string{"risk"},
				StoreAs:   "timing",
			},
			{
				StepID:      "prioritize",
				Tool:        "make-decision",
				Description: "Prioritize refactoring order",
				StaticInputs: map[string]any{
					"question": "Which phase should be executed first?",
					"criteria": []map[string]any{
						{"name": "risk_reduction", "weight": 0.3, "maximize": true},
						{"name": "effort", "weight": 0.25, "maximize": false},
						{"name": "impact", "weight": 0.25, "maximize": true},
						{"name": "dependencies", "weight": 0.2, "maximize": false},
					},
				},
				DependsOn: []string{"timing"},
				StoreAs:   "priority",
			},
		},
	}
}
