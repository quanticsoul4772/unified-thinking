package presets

// DebugAnalysisPreset returns the debug-analysis workflow preset
func DebugAnalysisPreset() *WorkflowPreset {
	return &WorkflowPreset{
		ID:          "debug-analysis",
		Name:        "Debug Analysis",
		Description: "Structured debugging with causal analysis, hypothesis generation, and session tracking",
		Category:    "code",
		InputSchema: map[string]ParamSpec{
			"problem": {
				Type:        "string",
				Required:    true,
				Description: "Description of the bug or issue",
				Examples:    []any{"Users report 500 errors on checkout", "Memory leak in worker process"},
			},
			"observations": {
				Type:        "array",
				Required:    false,
				Description: "Observed symptoms or behaviors",
				Examples:    []any{[]string{"High CPU usage", "Slow response times"}},
			},
			"context": {
				Type:        "string",
				Required:    false,
				Description: "Relevant context (recent changes, environment)",
			},
		},
		OutputFormat:  "session_summary",
		EstimatedTime: "3-5 minutes",
		Tags:          []string{"debug", "analysis", "causal"},
		Steps: []PresetStep{
			{
				StepID:      "start_session",
				Tool:        "start-reasoning-session",
				Description: "Track debug session for learning",
				InputMap: map[string]string{
					"description": "problem",
				},
				StaticInputs: map[string]any{
					"session_id": "debug-${timestamp}",
					"domain":     "debugging",
				},
				StoreAs: "session",
			},
			{
				StepID:      "causal_graph",
				Tool:        "build-causal-graph",
				Description: "Map potential causes and relationships",
				InputMap: map[string]string{
					"description":  "problem",
					"observations": "observations",
				},
				DependsOn: []string{"start_session"},
				StoreAs:   "causal",
			},
			{
				StepID:      "hypotheses",
				Tool:        "generate-hypotheses",
				Description: "Generate possible explanations",
				InputMap: map[string]string{
					"observations": "observations",
				},
				StaticInputs: map[string]any{
					"max_hypotheses": 5,
				},
				DependsOn: []string{"causal_graph"},
				StoreAs:   "hypotheses",
			},
			{
				StepID:      "simulate",
				Tool:        "simulate-intervention",
				Description: "Test fix effectiveness",
				InputMap: map[string]string{
					"graph_id": "causal.graph_id",
				},
				StaticInputs: map[string]any{
					"intervention_type": "fix",
				},
				DependsOn: []string{"hypotheses"},
				StoreAs:   "simulation",
				Optional:  true,
			},
			{
				StepID:      "complete_session",
				Tool:        "complete-reasoning-session",
				Description: "Document findings and solution",
				InputMap: map[string]string{
					"session_id": "session.session_id",
				},
				StaticInputs: map[string]any{
					"status": "completed",
				},
				DependsOn: []string{"hypotheses", "simulate"},
				StoreAs:   "summary",
			},
		},
	}
}
