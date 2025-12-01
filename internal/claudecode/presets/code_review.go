package presets

// CodeReviewPreset returns the code-review workflow preset
func CodeReviewPreset() *WorkflowPreset {
	return &WorkflowPreset{
		ID:          "code-review",
		Name:        "Code Review",
		Description: "Systematic code review with documented reasoning, bias detection, and decision tracking",
		Category:    "code",
		InputSchema: map[string]ParamSpec{
			"code": {
				Type:        "string",
				Required:    true,
				Description: "The code to review",
				Examples:    []any{"function example() {...}", "class MyClass {...}"},
			},
			"focus": {
				Type:        "string",
				Required:    false,
				Default:     "quality",
				Description: "Review focus area (quality, performance, security, readability)",
				Examples:    []any{"performance", "security", "readability"},
			},
			"context": {
				Type:        "string",
				Required:    false,
				Description: "Additional context about the code or requirements",
			},
		},
		OutputFormat:  "decision",
		EstimatedTime: "2-3 minutes",
		Tags:          []string{"code", "review", "quality"},
		Steps: []PresetStep{
			{
				StepID:      "decompose",
				Tool:        "decompose-problem",
				Description: "Break code into reviewable chunks",
				InputMap: map[string]string{
					"problem": "code",
				},
				StaticInputs: map[string]any{
					"strategy": "components",
				},
				StoreAs: "decomposed",
			},
			{
				StepID:      "analyze",
				Tool:        "think",
				Description: "Analyze each code section",
				InputMap: map[string]string{
					"content": "code",
				},
				StaticInputs: map[string]any{
					"mode":       "linear",
					"confidence": 0.8,
				},
				DependsOn: []string{"decompose"},
				StoreAs:   "analysis",
			},
			{
				StepID:      "fallacies",
				Tool:        "detect-fallacies",
				Description: "Check for logical errors in implementation",
				InputMap: map[string]string{
					"content": "code",
				},
				DependsOn: []string{"analyze"},
				StoreAs:   "fallacies",
				Optional:  true,
			},
			{
				StepID:      "biases",
				Tool:        "detect-biases",
				Description: "Identify review blind spots",
				InputMap: map[string]string{
					"thought_id": "analysis.thought_id",
				},
				DependsOn: []string{"analyze"},
				StoreAs:   "biases",
				Optional:  true,
			},
			{
				StepID:      "decide",
				Tool:        "make-decision",
				Description: "Decide whether to approve or request changes",
				StaticInputs: map[string]any{
					"question": "Should this code be approved?",
					"options": []map[string]any{
						{
							"name":        "approve",
							"description": "Code meets quality standards",
							"scores":      map[string]float64{"quality": 0.8, "risk": 0.2},
						},
						{
							"name":        "request_changes",
							"description": "Code needs improvements",
							"scores":      map[string]float64{"quality": 0.4, "risk": 0.6},
						},
					},
					"criteria": []map[string]any{
						{"name": "quality", "weight": 0.6, "maximize": true},
						{"name": "risk", "weight": 0.4, "maximize": false},
					},
				},
				DependsOn: []string{"analyze", "fallacies", "biases"},
				StoreAs:   "decision",
			},
		},
	}
}
