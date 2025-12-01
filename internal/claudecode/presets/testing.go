package presets

// TestStrategyPreset returns the test-strategy workflow preset
func TestStrategyPreset() *WorkflowPreset {
	return &WorkflowPreset{
		ID:          "test-strategy",
		Name:        "Test Strategy",
		Description: "Comprehensive test planning with component analysis, failure prediction, and coverage optimization",
		Category:    "testing",
		InputSchema: map[string]ParamSpec{
			"target": {
				Type:        "string",
				Required:    true,
				Description: "System or component to test",
				Examples:    []any{"User registration flow", "Payment processing module"},
			},
			"test_types": {
				Type:        "array",
				Required:    false,
				Default:     []string{"unit", "integration", "e2e"},
				Description: "Types of tests to plan",
			},
			"priority": {
				Type:        "string",
				Required:    false,
				Default:     "coverage",
				Description: "Testing priority (coverage, speed, risk)",
			},
		},
		OutputFormat:  "test_plan",
		EstimatedTime: "3-4 minutes",
		Tags:          []string{"testing", "strategy", "coverage"},
		Steps: []PresetStep{
			{
				StepID:      "decompose",
				Tool:        "decompose-problem",
				Description: "Identify testable components",
				InputMap: map[string]string{
					"problem": "target",
				},
				StaticInputs: map[string]any{
					"strategy": "components",
				},
				StoreAs: "components",
			},
			{
				StepID:      "perspectives",
				Tool:        "analyze-perspectives",
				Description: "Consider tester, developer, and user views",
				InputMap: map[string]string{
					"topic": "target",
				},
				StaticInputs: map[string]any{
					"stakeholders": []string{"developer", "tester", "end_user", "ops"},
				},
				DependsOn: []string{"decompose"},
				StoreAs:   "perspectives",
			},
			{
				StepID:      "blind_spots",
				Tool:        "detect-blind-spots",
				Description: "Find untested scenarios",
				InputMap: map[string]string{
					"context": "target",
				},
				DependsOn: []string{"perspectives"},
				StoreAs:   "gaps",
			},
			{
				StepID:      "predict_failures",
				Tool:        "generate-hypotheses",
				Description: "Predict potential failure modes",
				StaticInputs: map[string]any{
					"type":           "failure_modes",
					"max_hypotheses": 10,
				},
				DependsOn: []string{"blind_spots"},
				StoreAs:   "failures",
			},
			{
				StepID:      "generate_tests",
				Tool:        "think",
				Description: "Generate test cases",
				InputMap: map[string]string{
					"content": "target",
				},
				StaticInputs: map[string]any{
					"mode":       "tree",
					"confidence": 0.8,
					"key_points": []string{"edge cases", "error handling", "happy paths"},
				},
				DependsOn: []string{"predict_failures"},
				StoreAs:   "test_cases",
			},
		},
	}
}
