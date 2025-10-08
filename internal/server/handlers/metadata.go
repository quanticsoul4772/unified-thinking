package handlers

import (
	"fmt"
	"strings"

	"unified-thinking/internal/types"
)

// MetadataGenerator generates response metadata for Claude orchestration
type MetadataGenerator struct{}

// NewMetadataGenerator creates a new metadata generator
func NewMetadataGenerator() *MetadataGenerator {
	return &MetadataGenerator{}
}

// GenerateThinkMetadata generates metadata for think tool responses
func (mg *MetadataGenerator) GenerateThinkMetadata(
	thought *types.Thought,
	mode types.ThinkingMode,
	confidence float64,
	hasInsights bool,
	hasCrossRefs bool,
) *types.ResponseMetadata {
	metadata := &types.ResponseMetadata{
		SuggestedNextTools:      []types.ToolSuggestion{},
		ValidationOpportunities: []string{},
		ActionRecommendations:   []types.ActionRecommendation{},
		ExportFormats:           make(map[string]interface{}),
	}

	// Low confidence triggers validation suggestions
	if confidence < 0.7 {
		metadata.ValidationOpportunities = append(metadata.ValidationOpportunities,
			"Consider using web_search to gather additional evidence",
			"Query conversation:conversation_search for historical context",
		)
		metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
			ServerTool: "brave-search:brave_web_search",
			Reason:     "Gather external evidence to improve confidence",
			InputHint:  "Extract key terms from thought content for search query",
			Priority:   "recommended",
		})
	}

	// Mode-specific suggestions
	switch mode {
	case types.ModeLinear:
		mg.addLinearModeSuggestions(metadata, thought, confidence)
	case types.ModeTree:
		mg.addTreeModeSuggestions(metadata, thought, hasInsights, hasCrossRefs)
	case types.ModeDivergent:
		mg.addDivergentModeSuggestions(metadata, thought)
	}

	// Always suggest persistence for high-quality thoughts
	if confidence >= 0.8 {
		metadata.ActionRecommendations = append(metadata.ActionRecommendations, types.ActionRecommendation{
			Type:        "persist",
			Description: "Store this high-quality thought for future reference",
			ToolChain:   []string{"memory:create_entities", "obsidian:create-note"},
		})
	}

	return metadata
}

// addLinearModeSuggestions adds linear-mode specific suggestions
func (mg *MetadataGenerator) addLinearModeSuggestions(metadata *types.ResponseMetadata, thought *types.Thought, confidence float64) {
	// Suggest continuing the reasoning chain
	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "unified-thinking:think",
		Reason:     "Continue linear reasoning chain with next step",
		InputHint:  fmt.Sprintf("Set previous_thought_id to '%s' and mode to 'linear'", thought.ID),
		Priority:   "recommended",
	})

	// Suggest validation for completed reasoning
	if confidence >= 0.7 {
		metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
			ServerTool: "unified-thinking:validate",
			Reason:     "Validate logical consistency of reasoning chain",
			InputHint:  fmt.Sprintf("Validate thought_id '%s'", thought.ID),
			Priority:   "optional",
		})
	}
}

// addTreeModeSuggestions adds tree-mode specific suggestions
func (mg *MetadataGenerator) addTreeModeSuggestions(metadata *types.ResponseMetadata, thought *types.Thought, hasInsights bool, hasCrossRefs bool) {
	if hasInsights {
		metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
			ServerTool: "memory:create_entities",
			Reason:     "Persist branch insights as knowledge graph entities",
			InputHint:  "Use branch insights as entity observations",
			Priority:   "recommended",
		})
	}

	if hasCrossRefs {
		metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
			ServerTool: "memory:create_relations",
			Reason:     "Represent cross-branch relationships in knowledge graph",
			InputHint:  "Map cross-refs to entity relations",
			Priority:   "optional",
		})
	}

	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "unified-thinking:synthesize-insights",
		Reason:     "Combine insights from multiple branches",
		InputHint:  "Collect thoughts from related branches and synthesize",
		Priority:   "recommended",
	})
}

// addDivergentModeSuggestions adds divergent-mode specific suggestions
func (mg *MetadataGenerator) addDivergentModeSuggestions(metadata *types.ResponseMetadata, thought *types.Thought) {
	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "brave-search:brave_web_search",
		Reason:     "Find real-world examples of unconventional ideas",
		InputHint:  "Search for innovative examples related to this divergent thinking",
		Priority:   "optional",
	})

	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "unified-thinking:assess-evidence",
		Reason:     "Evaluate feasibility of unconventional ideas",
		InputHint:  "Assess evidence for and against this divergent approach",
		Priority:   "recommended",
	})
}

// GenerateDecisionMetadata generates metadata for decision tool responses
func (mg *MetadataGenerator) GenerateDecisionMetadata(decision *types.Decision) *types.ResponseMetadata {
	metadata := &types.ResponseMetadata{
		SuggestedNextTools:      []types.ToolSuggestion{},
		ValidationOpportunities: []string{},
		ActionRecommendations:   []types.ActionRecommendation{},
		ExportFormats:           make(map[string]interface{}),
	}

	// Add Obsidian export format
	if decision != nil {
		obsidianContent := mg.formatDecisionAsMarkdown(decision)
		metadata.ExportFormats["obsidian_note"] = types.ObsidianNoteExport{
			Title:   fmt.Sprintf("Decision: %s", decision.Question),
			Content: obsidianContent,
			Tags:    []string{"decision", "reasoning"},
			Properties: map[string]string{
				"decision_id": decision.ID,
				"confidence":  fmt.Sprintf("%.2f", decision.Confidence),
			},
		}
	}

	// Suggest documenting the decision
	metadata.ActionRecommendations = append(metadata.ActionRecommendations, types.ActionRecommendation{
		Type:        "persist",
		Description: "Document decision for future reference",
		ToolChain:   []string{"obsidian:create-note"},
	})

	// Suggest validation
	if decision != nil && decision.Confidence < 0.7 {
		metadata.ValidationOpportunities = append(metadata.ValidationOpportunities,
			"Search for additional criteria or data to improve confidence",
			"Consult conversation history for similar past decisions",
		)
	}

	return metadata
}

// GenerateCausalGraphMetadata generates metadata for causal graph responses
func (mg *MetadataGenerator) GenerateCausalGraphMetadata(graph *types.CausalGraph) *types.ResponseMetadata {
	metadata := &types.ResponseMetadata{
		SuggestedNextTools:      []types.ToolSuggestion{},
		ValidationOpportunities: []string{},
		ActionRecommendations:   []types.ActionRecommendation{},
		ExportFormats:           make(map[string]interface{}),
	}

	if graph == nil {
		return metadata
	}

	// Add Memory KG export format
	entities := []types.MemoryEntityExport{}
	relations := []types.MemoryRelationExport{}

	for _, variable := range graph.Variables {
		entities = append(entities, types.MemoryEntityExport{
			Name:         variable.Name,
			EntityType:   "causal_variable",
			Observations: []string{variable.Description, fmt.Sprintf("Type: %s", variable.Type)},
		})
	}

	for _, link := range graph.Links {
		// Find variable names
		fromName := mg.findVariableName(graph, link.From)
		toName := mg.findVariableName(graph, link.To)

		relations = append(relations, types.MemoryRelationExport{
			From:         fromName,
			To:           toName,
			RelationType: fmt.Sprintf("causes_%s", link.Type),
		})
	}

	metadata.ExportFormats["memory_entities"] = entities
	metadata.ExportFormats["memory_relations"] = relations

	// Suggest next steps
	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "memory:create_entities",
		Reason:     "Persist causal model in knowledge graph",
		InputHint:  "Use export_formats.memory_entities from this response",
		Priority:   "recommended",
	})

	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "unified-thinking:simulate-intervention",
		Reason:     "Test interventions on this causal model",
		InputHint:  fmt.Sprintf("Use graph_id '%s' to simulate interventions", graph.ID),
		Priority:   "recommended",
	})

	metadata.ValidationOpportunities = append(metadata.ValidationOpportunities,
		"Search web to validate causal relationships",
		"Query existing knowledge graph for related causal patterns",
	)

	return metadata
}

// formatDecisionAsMarkdown formats a decision as Obsidian-friendly markdown
func (mg *MetadataGenerator) formatDecisionAsMarkdown(decision *types.Decision) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Decision: %s\n\n", decision.Question))
	sb.WriteString(fmt.Sprintf("**Confidence**: %.2f\n\n", decision.Confidence))
	sb.WriteString(fmt.Sprintf("**Recommendation**: %s\n\n", decision.Recommendation))

	sb.WriteString("## Options\n\n")
	for _, opt := range decision.Options {
		sb.WriteString(fmt.Sprintf("### %s (Score: %.2f)\n\n", opt.Name, opt.TotalScore))
		sb.WriteString(fmt.Sprintf("%s\n\n", opt.Description))

		if len(opt.Pros) > 0 {
			sb.WriteString("**Pros:**\n")
			for _, pro := range opt.Pros {
				sb.WriteString(fmt.Sprintf("- %s\n", pro))
			}
			sb.WriteString("\n")
		}

		if len(opt.Cons) > 0 {
			sb.WriteString("**Cons:**\n")
			for _, con := range opt.Cons {
				sb.WriteString(fmt.Sprintf("- %s\n", con))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("## Criteria\n\n")
	for _, crit := range decision.Criteria {
		sb.WriteString(fmt.Sprintf("- **%s** (weight: %.2f): %s\n", crit.Name, crit.Weight, crit.Description))
	}

	return sb.String()
}

// findVariableName finds a variable name by ID in a causal graph
func (mg *MetadataGenerator) findVariableName(graph *types.CausalGraph, varID string) string {
	for _, v := range graph.Variables {
		if v.ID == varID {
			return v.Name
		}
	}
	return varID // Fallback to ID if not found
}

// GenerateDecomposeProblemMetadata generates metadata for problem decomposition responses
func (mg *MetadataGenerator) GenerateDecomposeProblemMetadata(decomposition *types.ProblemDecomposition) *types.ResponseMetadata {
	metadata := &types.ResponseMetadata{
		SuggestedNextTools:      []types.ToolSuggestion{},
		ValidationOpportunities: []string{},
		ActionRecommendations:   []types.ActionRecommendation{},
		ExportFormats:           make(map[string]interface{}),
	}

	if decomposition == nil {
		return metadata
	}

	// Suggest research for gathering information
	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "brave-search:brave_web_search",
		Reason:     "Research subproblems to gather context and solutions",
		InputHint:  "Search for solutions to each subproblem individually",
		Priority:   "recommended",
	})

	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "obsidian:search-notes",
		Reason:     "Check existing notes for related problem-solving approaches",
		InputHint:  "Search notes for keywords related to subproblems",
		Priority:   "optional",
	})

	// Suggest think tool for solving subproblems
	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "unified-thinking:think",
		Reason:     "Reason through each subproblem systematically",
		InputHint:  "Use linear mode to solve subproblems in dependency order",
		Priority:   "recommended",
	})

	// Add Obsidian export format as checklist
	if len(decomposition.Subproblems) > 0 {
		checklistItems := []string{}
		for _, sub := range decomposition.Subproblems {
			checklistItems = append(checklistItems, fmt.Sprintf("- [ ] %s (%s priority, %s complexity)", sub.Description, sub.Priority, sub.Complexity))
		}

		metadata.ExportFormats["obsidian_note"] = types.ObsidianNoteExport{
			Title:   fmt.Sprintf("Problem Decomposition: %s", decomposition.Problem),
			Content: fmt.Sprintf("# Problem\n\n%s\n\n# Subproblems\n\n%s", decomposition.Problem, strings.Join(checklistItems, "\n")),
			Tags:    []string{"problem-solving", "decomposition"},
			Properties: map[string]string{
				"decomposition_id": decomposition.ID,
				"subproblem_count": fmt.Sprintf("%d", len(decomposition.Subproblems)),
			},
		}
	}

	return metadata
}

// GenerateTemporalAnalysisMetadata generates metadata for temporal analysis responses
func (mg *MetadataGenerator) GenerateTemporalAnalysisMetadata(analysis *types.TemporalAnalysis) *types.ResponseMetadata {
	metadata := &types.ResponseMetadata{
		SuggestedNextTools:      []types.ToolSuggestion{},
		ValidationOpportunities: []string{},
		ActionRecommendations:   []types.ActionRecommendation{},
		ExportFormats:           make(map[string]interface{}),
	}

	if analysis == nil {
		return metadata
	}

	// Suggest checking conversation history for past decisions
	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "conversation:conversation_search",
		Reason:     "Review historical decisions with similar time horizons",
		InputHint:  "Search for past temporal trade-off discussions",
		Priority:   "optional",
	})

	// Suggest documenting the analysis
	metadata.ActionRecommendations = append(metadata.ActionRecommendations, types.ActionRecommendation{
		Type:        "persist",
		Description: "Document temporal analysis for future reference",
		ToolChain:   []string{"obsidian:create-note"},
	})

	// Add Obsidian export format
	metadata.ExportFormats["obsidian_note"] = types.ObsidianNoteExport{
		Title: fmt.Sprintf("Temporal Analysis: %s horizon", analysis.TimeHorizon),
		Content: fmt.Sprintf("# Short-Term View\n\n%s\n\n# Long-Term View\n\n%s\n\n# Tradeoffs\n\n%s\n\n# Recommendation\n\n%s",
			analysis.ShortTermView,
			analysis.LongTermView,
			strings.Join(analysis.Tradeoffs, "\n- "),
			analysis.Recommendation),
		Tags: []string{"temporal-analysis", "decision-making"},
		Properties: map[string]string{
			"time_horizon": analysis.TimeHorizon,
		},
	}

	return metadata
}

// GeneratePerspectiveAnalysisMetadata generates metadata for perspective analysis responses
func (mg *MetadataGenerator) GeneratePerspectiveAnalysisMetadata(perspectives []*types.Perspective) *types.ResponseMetadata {
	metadata := &types.ResponseMetadata{
		SuggestedNextTools:      []types.ToolSuggestion{},
		ValidationOpportunities: []string{},
		ActionRecommendations:   []types.ActionRecommendation{},
		ExportFormats:           make(map[string]interface{}),
	}

	if len(perspectives) == 0 {
		return metadata
	}

	// Suggest researching stakeholders
	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "obsidian:search-notes",
		Reason:     "Find existing stakeholder documentation",
		InputHint:  "Search for stakeholder names and their documented concerns",
		Priority:   "recommended",
	})

	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "brave-search:brave_web_search",
		Reason:     "Research stakeholder backgrounds and public positions",
		InputHint:  "Search for each stakeholder's stated priorities and concerns",
		Priority:   "optional",
	})

	// Suggest synthesizing insights
	metadata.SuggestedNextTools = append(metadata.SuggestedNextTools, types.ToolSuggestion{
		ServerTool: "unified-thinking:synthesize-insights",
		Reason:     "Identify common ground and conflicts between perspectives",
		InputHint:  "Combine perspective analysis results",
		Priority:   "recommended",
	})

	// Add Memory KG export for stakeholders
	entities := []types.MemoryEntityExport{}
	for _, persp := range perspectives {
		entities = append(entities, types.MemoryEntityExport{
			Name:       persp.Stakeholder,
			EntityType: "stakeholder",
			Observations: append([]string{persp.Viewpoint},
				append(persp.Concerns, persp.Priorities...)...),
		})
	}
	metadata.ExportFormats["memory_entities"] = entities

	return metadata
}
