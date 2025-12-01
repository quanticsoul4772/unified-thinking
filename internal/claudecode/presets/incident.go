package presets

// IncidentInvestigationPreset returns the incident-investigation workflow preset
func IncidentInvestigationPreset() *WorkflowPreset {
	return &WorkflowPreset{
		ID:          "incident-investigation",
		Name:        "Incident Investigation",
		Description: "Post-incident analysis with timeline mapping, causal analysis, and prevention recommendations",
		Category:    "operations",
		InputSchema: map[string]ParamSpec{
			"incident": {
				Type:        "string",
				Required:    true,
				Description: "Description of the incident",
				Examples:    []any{"Production outage at 2pm", "Data corruption in orders table"},
			},
			"timeline": {
				Type:        "array",
				Required:    false,
				Description: "Timeline of events",
			},
			"severity": {
				Type:        "string",
				Required:    false,
				Default:     "medium",
				Description: "Incident severity (low, medium, high, critical)",
			},
		},
		OutputFormat:  "postmortem",
		EstimatedTime: "4-5 minutes",
		Tags:          []string{"incident", "postmortem", "operations"},
		Steps: []PresetStep{
			{
				StepID:      "start_session",
				Tool:        "start-reasoning-session",
				Description: "Track investigation reasoning",
				InputMap: map[string]string{
					"problem": "incident",
				},
				StaticInputs: map[string]any{
					"session_type": "investigation",
				},
				StoreAs: "session",
			},
			{
				StepID:      "timeline",
				Tool:        "build-causal-graph",
				Description: "Map incident timeline and causes",
				InputMap: map[string]string{
					"description": "incident",
				},
				StaticInputs: map[string]any{
					"type":            "incident_timeline",
					"include_timing":  true,
					"identify_causes": true,
				},
				DependsOn: []string{"start_session"},
				StoreAs:   "causal",
			},
			{
				StepID:      "counterfactual",
				Tool:        "generate-counterfactual",
				Description: "Explore 'what if' prevention scenarios",
				InputMap: map[string]string{
					"graph_id": "causal.graph_id",
				},
				StaticInputs: map[string]any{
					"focus": "prevention",
					"scenarios": []string{
						"earlier_detection",
						"different_architecture",
						"better_monitoring",
					},
				},
				DependsOn: []string{"timeline"},
				StoreAs:   "scenarios",
			},
			{
				StepID:      "bias_check",
				Tool:        "detect-biases",
				Description: "Check for hindsight and outcome bias",
				InputMap: map[string]string{
					"content": "incident",
				},
				StaticInputs: map[string]any{
					"focus_biases": []string{"hindsight_bias", "outcome_bias", "attribution_error"},
				},
				DependsOn: []string{"counterfactual"},
				StoreAs:   "biases",
			},
			{
				StepID:      "recommendations",
				Tool:        "make-decision",
				Description: "Recommend prevention measures",
				StaticInputs: map[string]any{
					"question": "Which prevention measures should be prioritized?",
					"criteria": []map[string]any{
						{"name": "effectiveness", "weight": 0.35, "maximize": true},
						{"name": "effort", "weight": 0.25, "maximize": false},
						{"name": "risk_reduction", "weight": 0.25, "maximize": true},
						{"name": "time_to_implement", "weight": 0.15, "maximize": false},
					},
				},
				DependsOn: []string{"bias_check"},
				StoreAs:   "prevention",
			},
			{
				StepID:      "complete",
				Tool:        "complete-reasoning-session",
				Description: "Generate postmortem document",
				InputMap: map[string]string{
					"session_id": "session.session_id",
				},
				StaticInputs: map[string]any{
					"success":       true,
					"generate_doc":  true,
					"include_graph": true,
				},
				DependsOn: []string{"recommendations"},
				StoreAs:   "postmortem",
			},
		},
	}
}
