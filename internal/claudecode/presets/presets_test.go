package presets

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry should not return nil")
	}
}

func TestRegistryGetBuiltInPresets(t *testing.T) {
	registry := NewRegistry()

	expectedPresets := []string{
		"code-review",
		"debug-analysis",
		"architecture-decision",
		"research-synthesis",
		"refactoring-plan",
		"test-strategy",
		"documentation-gen",
		"incident-investigation",
	}

	for _, id := range expectedPresets {
		preset, err := registry.Get(id)
		if err != nil {
			t.Errorf("Expected built-in preset '%s' to exist: %v", id, err)
			continue
		}
		if preset == nil {
			t.Errorf("Preset '%s' should not be nil", id)
			continue
		}
		if preset.ID != id {
			t.Errorf("Preset ID mismatch: expected %s, got %s", id, preset.ID)
		}
	}
}

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()
	presets := registry.List("")

	if len(presets) < 8 {
		t.Errorf("Expected at least 8 built-in presets, got %d", len(presets))
	}

	// Verify all presets have required fields
	for _, preset := range presets {
		if preset.ID == "" {
			t.Error("Preset ID should not be empty")
		}
		if preset.Name == "" {
			t.Error("Preset Name should not be empty")
		}
		if preset.Description == "" {
			t.Error("Preset Description should not be empty")
		}
		if preset.Category == "" {
			t.Error("Preset Category should not be empty")
		}
	}
}

func TestRegistryListByCategory(t *testing.T) {
	registry := NewRegistry()

	categories := map[string]int{
		"code":          3, // code-review, debug-analysis, refactoring-plan
		"architecture":  1, // architecture-decision
		"research":      1, // research-synthesis
		"testing":       1, // test-strategy
		"documentation": 1, // documentation-gen
		"operations":    1, // incident-investigation
	}

	for category, expectedCount := range categories {
		presets := registry.List(category)
		if len(presets) != expectedCount {
			t.Errorf("Category '%s': expected %d presets, got %d", category, expectedCount, len(presets))
		}
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()

	customPreset := &WorkflowPreset{
		ID:          "custom-preset",
		Name:        "Custom Preset",
		Description: "A custom workflow",
		Category:    "custom",
		Steps: []PresetStep{
			{StepID: "step1", Tool: "think", Description: "Think about it"},
		},
	}

	err := registry.Register(customPreset)
	if err != nil {
		t.Errorf("Failed to register custom preset: %v", err)
	}

	preset, err := registry.Get("custom-preset")
	if err != nil {
		t.Errorf("Custom preset should be registered: %v", err)
	}
	if preset.Name != "Custom Preset" {
		t.Errorf("Expected name 'Custom Preset', got '%s'", preset.Name)
	}

	// Should appear in category list
	customPresets := registry.List("custom")
	if len(customPresets) != 1 {
		t.Errorf("Expected 1 custom preset, got %d", len(customPresets))
	}
}

func TestRegistryRegisterValidation(t *testing.T) {
	registry := NewRegistry()

	// Missing ID
	err := registry.Register(&WorkflowPreset{
		Name:  "Test",
		Steps: []PresetStep{{StepID: "s1", Tool: "think"}},
	})
	if err == nil {
		t.Error("Should reject preset without ID")
	}

	// Missing Name
	err = registry.Register(&WorkflowPreset{
		ID:    "test",
		Steps: []PresetStep{{StepID: "s1", Tool: "think"}},
	})
	if err == nil {
		t.Error("Should reject preset without Name")
	}

	// Missing Steps
	err = registry.Register(&WorkflowPreset{
		ID:   "test",
		Name: "Test",
	})
	if err == nil {
		t.Error("Should reject preset without Steps")
	}
}

func TestRegistryGetNonExistent(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("non-existent-preset")
	if err == nil {
		t.Error("Non-existent preset should return error")
	}
}

func TestRegistryCount(t *testing.T) {
	registry := NewRegistry()
	count := registry.Count()
	if count != 8 {
		t.Errorf("Expected 8 built-in presets, got %d", count)
	}
}

func TestRegistryCategories(t *testing.T) {
	registry := NewRegistry()
	categories := registry.Categories()

	if len(categories) < 5 {
		t.Errorf("Expected at least 5 categories, got %d", len(categories))
	}
}

func TestCodeReviewPreset(t *testing.T) {
	preset := CodeReviewPreset()

	if preset.ID != "code-review" {
		t.Errorf("Expected ID 'code-review', got '%s'", preset.ID)
	}
	if preset.Category != "code" {
		t.Errorf("Expected category 'code', got '%s'", preset.Category)
	}

	// Verify required input params
	codeParam, exists := preset.InputSchema["code"]
	if !exists {
		t.Error("Should have 'code' input parameter")
	}
	if !codeParam.Required {
		t.Error("'code' parameter should be required")
	}

	// Verify steps
	if len(preset.Steps) != 5 {
		t.Errorf("Expected 5 steps, got %d", len(preset.Steps))
	}
}

func TestDebugAnalysisPreset(t *testing.T) {
	preset := DebugAnalysisPreset()

	if preset.ID != "debug-analysis" {
		t.Errorf("Expected ID 'debug-analysis', got '%s'", preset.ID)
	}
	if preset.Category != "code" {
		t.Errorf("Expected category 'code', got '%s'", preset.Category)
	}

	// Verify required input params - the field is "problem" not "error"
	problemParam, exists := preset.InputSchema["problem"]
	if !exists {
		t.Error("Should have 'problem' input parameter")
	}
	if !problemParam.Required {
		t.Error("'problem' parameter should be required")
	}

	// Verify steps
	if len(preset.Steps) != 5 {
		t.Errorf("Expected 5 steps, got %d", len(preset.Steps))
	}
}

func TestArchitectureDecisionPreset(t *testing.T) {
	preset := ArchitectureDecisionPreset()

	if preset.ID != "architecture-decision" {
		t.Errorf("Expected ID 'architecture-decision', got '%s'", preset.ID)
	}

	// Verify required input params
	decisionParam, exists := preset.InputSchema["decision"]
	if !exists {
		t.Error("Should have 'decision' input parameter")
	}
	if !decisionParam.Required {
		t.Error("'decision' parameter should be required")
	}

	optionsParam, exists := preset.InputSchema["options"]
	if !exists {
		t.Error("Should have 'options' input parameter")
	}
	if !optionsParam.Required {
		t.Error("'options' parameter should be required")
	}
}

func TestResearchSynthesisPreset(t *testing.T) {
	preset := ResearchSynthesisPreset()

	if preset.ID != "research-synthesis" {
		t.Errorf("Expected ID 'research-synthesis', got '%s'", preset.ID)
	}

	// Should use GoT tools
	gotToolsUsed := false
	for _, step := range preset.Steps {
		if step.Tool == "got-initialize" || step.Tool == "got-generate" || step.Tool == "got-aggregate" {
			gotToolsUsed = true
			break
		}
	}
	if !gotToolsUsed {
		t.Error("Research synthesis should use Graph-of-Thoughts tools")
	}
}

func TestRefactoringPlanPreset(t *testing.T) {
	preset := RefactoringPlanPreset()

	if preset.ID != "refactoring-plan" {
		t.Errorf("Expected ID 'refactoring-plan', got '%s'", preset.ID)
	}

	// Verify goals is required
	goalsParam, exists := preset.InputSchema["goals"]
	if !exists {
		t.Error("Should have 'goals' input parameter")
	}
	if !goalsParam.Required {
		t.Error("'goals' parameter should be required")
	}
}

func TestTestStrategyPreset(t *testing.T) {
	preset := TestStrategyPreset()

	if preset.ID != "test-strategy" {
		t.Errorf("Expected ID 'test-strategy', got '%s'", preset.ID)
	}

	// Verify target is required
	targetParam, exists := preset.InputSchema["target"]
	if !exists {
		t.Error("Should have 'target' input parameter")
	}
	if !targetParam.Required {
		t.Error("'target' parameter should be required")
	}
}

func TestDocumentationGenPreset(t *testing.T) {
	preset := DocumentationGenPreset()

	if preset.ID != "documentation-gen" {
		t.Errorf("Expected ID 'documentation-gen', got '%s'", preset.ID)
	}

	// Verify code is required
	codeParam, exists := preset.InputSchema["code"]
	if !exists {
		t.Error("Should have 'code' input parameter")
	}
	if !codeParam.Required {
		t.Error("'code' parameter should be required")
	}
}

func TestIncidentInvestigationPreset(t *testing.T) {
	preset := IncidentInvestigationPreset()

	if preset.ID != "incident-investigation" {
		t.Errorf("Expected ID 'incident-investigation', got '%s'", preset.ID)
	}
	if preset.Category != "operations" {
		t.Errorf("Expected category 'operations', got '%s'", preset.Category)
	}

	// Verify incident is required
	incidentParam, exists := preset.InputSchema["incident"]
	if !exists {
		t.Error("Should have 'incident' input parameter")
	}
	if !incidentParam.Required {
		t.Error("'incident' parameter should be required")
	}

	// Should have reasoning session steps
	hasSessionStart := false
	hasSessionComplete := false
	for _, step := range preset.Steps {
		if step.Tool == "start-reasoning-session" {
			hasSessionStart = true
		}
		if step.Tool == "complete-reasoning-session" {
			hasSessionComplete = true
		}
	}
	if !hasSessionStart || !hasSessionComplete {
		t.Error("Incident investigation should track reasoning session")
	}
}

func TestPresetStepDependencies(t *testing.T) {
	presets := []*WorkflowPreset{
		CodeReviewPreset(),
		DebugAnalysisPreset(),
		ArchitectureDecisionPreset(),
		ResearchSynthesisPreset(),
		RefactoringPlanPreset(),
		TestStrategyPreset(),
		DocumentationGenPreset(),
		IncidentInvestigationPreset(),
	}

	for _, preset := range presets {
		stepIDs := make(map[string]bool)
		for _, step := range preset.Steps {
			stepIDs[step.StepID] = true
		}

		for _, step := range preset.Steps {
			for _, dep := range step.DependsOn {
				if !stepIDs[dep] {
					t.Errorf("Preset %s: step %s depends on non-existent step %s",
						preset.ID, step.StepID, dep)
				}
			}
		}
	}
}

func TestPresetStepToolNames(t *testing.T) {
	// Valid tools in the unified-thinking server
	validTools := map[string]bool{
		"think":                           true,
		"history":                         true,
		"search":                          true,
		"validate":                        true,
		"prove":                           true,
		"decompose-problem":               true,
		"make-decision":                   true,
		"probabilistic-reasoning":         true,
		"assess-evidence":                 true,
		"detect-contradictions":           true,
		"sensitivity-analysis":            true,
		"self-evaluate":                   true,
		"detect-biases":                   true,
		"detect-blind-spots":              true,
		"analyze-perspectives":            true,
		"analyze-temporal":                true,
		"build-causal-graph":              true,
		"simulate-intervention":           true,
		"generate-counterfactual":         true,
		"synthesize-insights":             true,
		"detect-emergent-patterns":        true,
		"generate-hypotheses":             true,
		"evaluate-hypotheses":             true,
		"start-reasoning-session":         true,
		"complete-reasoning-session":      true,
		"get-recommendations":             true,
		"got-initialize":                  true,
		"got-generate":                    true,
		"got-aggregate":                   true,
		"got-refine":                      true,
		"got-finalize":                    true,
		"decompose-argument":              true,
		"detect-fallacies":                true,
		"process-evidence-pipeline":       true,
		"analyze-temporal-causal-effects": true,
	}

	presets := []*WorkflowPreset{
		CodeReviewPreset(),
		DebugAnalysisPreset(),
		ArchitectureDecisionPreset(),
		ResearchSynthesisPreset(),
		RefactoringPlanPreset(),
		TestStrategyPreset(),
		DocumentationGenPreset(),
		IncidentInvestigationPreset(),
	}

	for _, preset := range presets {
		for _, step := range preset.Steps {
			if !validTools[step.Tool] {
				t.Errorf("Preset %s: step %s uses unknown tool '%s'",
					preset.ID, step.StepID, step.Tool)
			}
		}
	}
}

func TestPresetHasTags(t *testing.T) {
	presets := []*WorkflowPreset{
		CodeReviewPreset(),
		DebugAnalysisPreset(),
		ArchitectureDecisionPreset(),
		ResearchSynthesisPreset(),
		RefactoringPlanPreset(),
		TestStrategyPreset(),
		DocumentationGenPreset(),
		IncidentInvestigationPreset(),
	}

	for _, preset := range presets {
		if len(preset.Tags) == 0 {
			t.Errorf("Preset %s should have tags", preset.ID)
		}
	}
}

func TestPresetHasEstimatedTime(t *testing.T) {
	presets := []*WorkflowPreset{
		CodeReviewPreset(),
		DebugAnalysisPreset(),
		ArchitectureDecisionPreset(),
		ResearchSynthesisPreset(),
		RefactoringPlanPreset(),
		TestStrategyPreset(),
		DocumentationGenPreset(),
		IncidentInvestigationPreset(),
	}

	for _, preset := range presets {
		if preset.EstimatedTime == "" {
			t.Errorf("Preset %s should have estimated time", preset.ID)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Test the global default registry
	preset, err := GetPreset("code-review")
	if err != nil {
		t.Errorf("GetPreset should find code-review: %v", err)
	}
	if preset == nil {
		t.Error("GetPreset should return non-nil preset")
	}

	presets := ListPresets("")
	if len(presets) < 8 {
		t.Errorf("ListPresets should return at least 8 presets, got %d", len(presets))
	}
}

func TestPresetToSummary(t *testing.T) {
	preset := CodeReviewPreset()
	summary := preset.ToSummary()

	if summary.ID != preset.ID {
		t.Errorf("Summary ID mismatch: got %s, want %s", summary.ID, preset.ID)
	}
	if summary.Name != preset.Name {
		t.Errorf("Summary Name mismatch: got %s, want %s", summary.Name, preset.Name)
	}
	if summary.Category != preset.Category {
		t.Errorf("Summary Category mismatch: got %s, want %s", summary.Category, preset.Category)
	}
	if summary.StepCount != len(preset.Steps) {
		t.Errorf("Summary StepCount mismatch: got %d, want %d", summary.StepCount, len(preset.Steps))
	}
}
