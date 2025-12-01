package presets

// ArchitectureDecisionPreset returns the architecture-decision workflow preset
func ArchitectureDecisionPreset() *WorkflowPreset {
	return &WorkflowPreset{
		ID:          "architecture-decision",
		Name:        "Architecture Decision Record",
		Description: "Generate comprehensive ADRs with stakeholder analysis, temporal considerations, and blind spot detection",
		Category:    "architecture",
		InputSchema: map[string]ParamSpec{
			"decision": {
				Type:        "string",
				Required:    true,
				Description: "The architectural decision to be made",
				Examples:    []any{"Choose database technology", "Select authentication approach"},
			},
			"options": {
				Type:        "array",
				Required:    true,
				Description: "Available options to evaluate",
				Examples:    []any{[]string{"PostgreSQL", "MongoDB", "DynamoDB"}},
			},
			"constraints": {
				Type:        "array",
				Required:    false,
				Description: "Known constraints or requirements",
			},
			"stakeholders": {
				Type:        "array",
				Required:    false,
				Default:     []string{"developers", "operations", "business"},
				Description: "Stakeholder groups to consider",
			},
		},
		OutputFormat:  "adr_document",
		EstimatedTime: "3-4 minutes",
		Tags:          []string{"architecture", "adr", "decision"},
		Steps: []PresetStep{
			{
				StepID:      "perspectives",
				Tool:        "analyze-perspectives",
				Description: "Gather all stakeholder viewpoints",
				InputMap: map[string]string{
					"topic":        "decision",
					"stakeholders": "stakeholders",
				},
				StoreAs: "perspectives",
			},
			{
				StepID:      "temporal",
				Tool:        "analyze-temporal",
				Description: "Evaluate short-term vs long-term implications",
				InputMap: map[string]string{
					"topic": "decision",
				},
				StaticInputs: map[string]any{
					"time_horizons": []string{"immediate", "1_year", "3_years"},
				},
				DependsOn: []string{"perspectives"},
				StoreAs:   "temporal",
			},
			{
				StepID:      "decide",
				Tool:        "make-decision",
				Description: "Multi-criteria evaluation of options",
				InputMap: map[string]string{
					"question": "decision",
				},
				StaticInputs: map[string]any{
					"criteria": []map[string]any{
						{"name": "scalability", "weight": 0.25, "maximize": true},
						{"name": "maintainability", "weight": 0.25, "maximize": true},
						{"name": "cost", "weight": 0.2, "maximize": false},
						{"name": "risk", "weight": 0.15, "maximize": false},
						{"name": "time_to_implement", "weight": 0.15, "maximize": false},
					},
				},
				DependsOn: []string{"temporal"},
				StoreAs:   "decision_result",
			},
			{
				StepID:      "blind_spots",
				Tool:        "detect-blind-spots",
				Description: "Find gaps in the analysis",
				InputMap: map[string]string{
					"context": "decision",
				},
				DependsOn: []string{"decide"},
				StoreAs:   "blind_spots",
				Optional:  true,
			},
			{
				StepID:      "document",
				Tool:        "think",
				Description: "Generate the ADR document",
				StaticInputs: map[string]any{
					"mode":       "linear",
					"confidence": 0.85,
					"content":    "Generate ADR based on analysis",
				},
				DependsOn: []string{"perspectives", "temporal", "decide", "blind_spots"},
				StoreAs:   "adr",
			},
		},
	}
}
