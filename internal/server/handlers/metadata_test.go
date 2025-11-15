package handlers

import (
	"strings"
	"testing"

	"unified-thinking/internal/types"
)

func TestGenerateThinkMetadataLinearLowConfidence(t *testing.T) {
	mg := NewMetadataGenerator()
	thought := &types.Thought{ID: "thought-1"}

	metadata := mg.GenerateThinkMetadata(thought, types.ModeLinear, 0.65, false, false)

	if len(metadata.ValidationOpportunities) != 2 {
		t.Fatalf("expected 2 validation opportunities, got %d", len(metadata.ValidationOpportunities))
	}

	if !containsTool(metadata.SuggestedNextTools, "brave-search:brave_web_search") {
		t.Fatalf("expected brave search suggestion: %+v", metadata.SuggestedNextTools)
	}

	if !containsTool(metadata.SuggestedNextTools, "unified-thinking:think") {
		t.Fatalf("expected linear follow-up suggestion: %+v", metadata.SuggestedNextTools)
	}

	if containsAction(metadata.ActionRecommendations, "persist") {
		t.Fatalf("did not expect persist recommendation at low confidence")
	}
}

func TestGenerateThinkMetadataTreeHighConfidence(t *testing.T) {
	mg := NewMetadataGenerator()
	thought := &types.Thought{ID: "thought-tree"}

	metadata := mg.GenerateThinkMetadata(thought, types.ModeTree, 0.9, true, true)

	if !containsTool(metadata.SuggestedNextTools, "memory:create_entities") {
		t.Fatalf("expected memory entity suggestion: %+v", metadata.SuggestedNextTools)
	}

	if !containsTool(metadata.SuggestedNextTools, "memory:create_relations") {
		t.Fatalf("expected memory relation suggestion: %+v", metadata.SuggestedNextTools)
	}

	if !containsTool(metadata.SuggestedNextTools, "unified-thinking:synthesize-insights") {
		t.Fatalf("expected synthesis suggestion: %+v", metadata.SuggestedNextTools)
	}

	if !containsAction(metadata.ActionRecommendations, "persist") {
		t.Fatalf("expected persist recommendation for high confidence")
	}
}

func TestGenerateThinkMetadataDivergent(t *testing.T) {
	mg := NewMetadataGenerator()
	thought := &types.Thought{ID: "thought-div"}

	metadata := mg.GenerateThinkMetadata(thought, types.ModeDivergent, 0.75, false, false)

	if !containsTool(metadata.SuggestedNextTools, "brave-search:brave_web_search") {
		t.Fatalf("expected brave search suggestion for divergent mode")
	}

	if !containsTool(metadata.SuggestedNextTools, "unified-thinking:assess-evidence") {
		t.Fatalf("expected assess-evidence suggestion for divergent mode")
	}
}

func TestGenerateDecisionMetadata(t *testing.T) {
	mg := NewMetadataGenerator()
	decision := &types.Decision{
		ID:             "dec-1",
		Question:       "Choose platform",
		Confidence:     0.6,
		Recommendation: "Adopt platform A",
		Options: []*types.DecisionOption{
			{
				ID:          "opt-a",
				Name:        "Platform A",
				Description: "Reliable and scalable",
				Pros:        []string{"mature ecosystem"},
				Cons:        []string{"higher cost"},
				TotalScore:  0.82,
			},
		},
		Criteria: []*types.DecisionCriterion{
			{
				ID:          "crit-1",
				Name:        "Stability",
				Description: "Track record of uptime",
				Weight:      0.5,
			},
		},
	}

	metadata := mg.GenerateDecisionMetadata(decision)

	if len(metadata.ActionRecommendations) == 0 {
		t.Fatal("expected persist recommendation for decisions")
	}

	if len(metadata.ValidationOpportunities) != 2 {
		t.Fatalf("expected validation opportunities for low confidence decision, got %d", len(metadata.ValidationOpportunities))
	}

	note, ok := metadata.ExportFormats["obsidian_note"].(types.ObsidianNoteExport)
	if !ok {
		t.Fatalf("expected obsidian note export, got %T", metadata.ExportFormats["obsidian_note"])
	}

	if !strings.Contains(note.Content, "## Options") {
		t.Fatalf("expected options section in markdown, got %s", note.Content)
	}
}

func TestGenerateCausalGraphMetadata(t *testing.T) {
	mg := NewMetadataGenerator()
	graph := &types.CausalGraph{
		ID: "graph-1",
		Variables: []*types.CausalVariable{
			{ID: "a", Name: "Traffic", Description: "Website traffic", Type: "continuous"},
			{ID: "b", Name: "Signups", Description: "User signups", Type: "continuous"},
		},
		Links: []*types.CausalLink{
			{ID: "link", From: "a", To: "b", Type: "positive"},
		},
	}

	metadata := mg.GenerateCausalGraphMetadata(graph)

	entities, ok := metadata.ExportFormats["memory_entities"].([]types.MemoryEntityExport)
	if !ok || len(entities) != 2 {
		t.Fatalf("expected two memory entities, got %#v", metadata.ExportFormats["memory_entities"])
	}

	relations, ok := metadata.ExportFormats["memory_relations"].([]types.MemoryRelationExport)
	if !ok || len(relations) != 1 {
		t.Fatalf("expected one relation, got %#v", metadata.ExportFormats["memory_relations"])
	}

	if relations[0].From != "Traffic" || relations[0].To != "Signups" {
		t.Fatalf("expected relation to reference variable names, got %+v", relations[0])
	}

	if !containsTool(metadata.SuggestedNextTools, "unified-thinking:simulate-intervention") {
		t.Fatalf("expected simulate intervention suggestion")
	}

	if len(metadata.ValidationOpportunities) != 2 {
		t.Fatalf("expected two validation opportunities, got %d", len(metadata.ValidationOpportunities))
	}
}

func TestGeneratePerspectiveAnalysisMetadata(t *testing.T) {
	mg := NewMetadataGenerator()
	perspectives := []*types.Perspective{
		{
			Stakeholder: "Engineering",
			Viewpoint:   "Prioritize reliability",
			Concerns:    []string{"downtime"},
			Priorities:  []string{"resilience"},
		},
		{
			Stakeholder: "Marketing",
			Viewpoint:   "Focus on growth",
			Concerns:    []string{"lead volume"},
			Priorities:  []string{"campaign speed"},
		},
	}

	metadata := mg.GeneratePerspectiveAnalysisMetadata(perspectives)

	if !containsTool(metadata.SuggestedNextTools, "unified-thinking:synthesize-insights") {
		t.Fatalf("expected synthesis suggestion for perspectives")
	}

	entities, ok := metadata.ExportFormats["memory_entities"].([]types.MemoryEntityExport)
	if !ok || len(entities) != 2 {
		t.Fatalf("expected stakeholder entities export, got %#v", metadata.ExportFormats["memory_entities"])
	}
}

func TestGenerateDecomposeProblemMetadata(t *testing.T) {
	mg := NewMetadataGenerator()
	decomposition := &types.ProblemDecomposition{
		ID:      "dec-1",
		Problem: "Reduce support backlog",
		Subproblems: []*types.Subproblem{
			{ID: "s1", Description: "Automate triage", Priority: "high", Complexity: "medium"},
			{ID: "s2", Description: "Improve docs", Priority: "medium", Complexity: "low"},
		},
	}

	metadata := mg.GenerateDecomposeProblemMetadata(decomposition)

	if !containsTool(metadata.SuggestedNextTools, "unified-thinking:think") {
		t.Fatalf("expected think suggestion for subproblem reasoning")
	}

	note, ok := metadata.ExportFormats["obsidian_note"].(types.ObsidianNoteExport)
	if !ok {
		t.Fatalf("expected obsidian note export, got %T", metadata.ExportFormats["obsidian_note"])
	}

	if !strings.Contains(note.Content, "# Subproblems") {
		t.Fatalf("expected subproblems section in exported note, got %s", note.Content)
	}
}

func TestGenerateTemporalAnalysisMetadata(t *testing.T) {
	mg := NewMetadataGenerator()
	analysis := &types.TemporalAnalysis{
		TimeHorizon:    "6 months",
		ShortTermView:  "Immediate gains",
		LongTermView:   "Sustainable growth",
		Tradeoffs:      []string{"Resource allocation"},
		Recommendation: "Balance short and long term",
	}

	metadata := mg.GenerateTemporalAnalysisMetadata(analysis)

	if len(metadata.ActionRecommendations) == 0 {
		t.Fatalf("expected action recommendation to persist analysis")
	}

	note, ok := metadata.ExportFormats["obsidian_note"].(types.ObsidianNoteExport)
	if !ok {
		t.Fatalf("expected obsidian note export, got %T", metadata.ExportFormats["obsidian_note"])
	}

	if !strings.Contains(note.Content, "# Tradeoffs") {
		t.Fatalf("expected tradeoffs section in note, got %s", note.Content)
	}
}

func containsTool(suggestions []types.ToolSuggestion, tool string) bool {
	for _, suggestion := range suggestions {
		if suggestion.ServerTool == tool {
			return true
		}
	}
	return false
}

func containsAction(actions []types.ActionRecommendation, actionType string) bool {
	for _, action := range actions {
		if action.Type == actionType {
			return true
		}
	}
	return false
}
