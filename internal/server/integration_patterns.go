package server

// IntegrationPattern represents a common multi-server workflow pattern
type IntegrationPattern struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	UseCase     string   `json:"use_case"`
	Servers     []string `json:"servers"` // MCP servers involved
}

// GetIntegrationPatterns returns all documented integration patterns
func GetIntegrationPatterns() []IntegrationPattern {
	return []IntegrationPattern{
		{
			Name:        "Research-Enhanced Thinking",
			Description: "Gather external evidence before reasoning, then validate conclusions",
			Steps: []string{
				"1. brave-search:brave_web_search - Search for relevant information",
				"2. unified-thinking:think - Reason with gathered context",
				"3. unified-thinking:assess-evidence - Validate evidence quality",
				"4. unified-thinking:synthesize-insights - Combine findings",
			},
			UseCase: "When reasoning about topics requiring external validation or current information",
			Servers: []string{"brave-search", "unified-thinking"},
		},
		{
			Name:        "Knowledge-Backed Decision Making",
			Description: "Query existing knowledge before making decisions, then document results",
			Steps: []string{
				"1. memory:traverse_graph - Find related concepts from knowledge base",
				"2. conversation:conversation_search - Check past discussions",
				"3. unified-thinking:make-decision - Decide with full context",
				"4. memory:create_entities - Store decision rationale",
				"5. obsidian:create-note - Document decision for future reference",
			},
			UseCase: "Important decisions that benefit from organizational memory and history",
			Servers: []string{"memory", "conversation", "unified-thinking", "obsidian"},
		},
		{
			Name:        "Causal Model to Knowledge Graph",
			Description: "Build causal models and persist them as structured knowledge",
			Steps: []string{
				"1. brave-search:brave_web_search - Research causal relationships",
				"2. unified-thinking:build-causal-graph - Create causal model",
				"3. memory:create_entities - Persist variables as entities (use export_formats.memory_entities)",
				"4. memory:create_relations - Persist causal links (use export_formats.memory_relations)",
				"5. unified-thinking:simulate-intervention - Test interventions",
			},
			UseCase: "Understanding complex systems and testing hypothetical interventions",
			Servers: []string{"brave-search", "unified-thinking", "memory"},
		},
		{
			Name:        "Problem Decomposition Workflow",
			Description: "Break down complex problems, research solutions, and track progress",
			Steps: []string{
				"1. unified-thinking:decompose-problem - Break into subproblems",
				"2. obsidian:create-note - Create tracking checklist (use export_formats.obsidian_note)",
				"3. brave-search:brave_web_search - Research each subproblem",
				"4. unified-thinking:think - Solve subproblems systematically",
				"5. unified-thinking:synthesize-insights - Combine solutions",
			},
			UseCase: "Tackling large, complex problems systematically with progress tracking",
			Servers: []string{"unified-thinking", "obsidian", "brave-search"},
		},
		{
			Name:        "Temporal Decision Analysis",
			Description: "Analyze time tradeoffs before making decisions",
			Steps: []string{
				"1. conversation:conversation_search - Find similar past decisions",
				"2. unified-thinking:analyze-temporal - Analyze short/long-term implications",
				"3. unified-thinking:make-decision - Decide with temporal insights",
				"4. obsidian:create-note - Document with temporal context (use export_formats)",
			},
			UseCase: "Decisions with significant time-dependent tradeoffs",
			Servers: []string{"conversation", "unified-thinking", "obsidian"},
		},
		{
			Name:        "Stakeholder-Aware Planning",
			Description: "Analyze multiple perspectives before planning actions",
			Steps: []string{
				"1. obsidian:search-notes - Find existing stakeholder documentation",
				"2. unified-thinking:analyze-perspectives - Identify concerns and priorities",
				"3. memory:create_entities - Store stakeholder profiles (use export_formats.memory_entities)",
				"4. unified-thinking:synthesize-insights - Find common ground",
				"5. unified-thinking:make-decision - Decide considering all perspectives",
			},
			UseCase: "Complex initiatives requiring alignment across multiple stakeholders",
			Servers: []string{"obsidian", "unified-thinking", "memory"},
		},
		{
			Name:        "Validated File Operations",
			Description: "Think through file operations before executing them",
			Steps: []string{
				"1. unified-thinking:analyze-temporal - Consider short/long-term implications",
				"2. unified-thinking:think - Reason through the operation",
				"3. unified-thinking:validate - Check logical consistency",
				"4. filesystem:write_file OR windows-cli:execute_command - Execute",
				"5. obsidian:create-note - Document what was done and why",
			},
			UseCase: "Potentially destructive operations that benefit from validation",
			Servers: []string{"unified-thinking", "filesystem", "windows-cli", "obsidian"},
		},
		{
			Name:        "Evidence-Based Causal Reasoning",
			Description: "Research, build causal models, and validate with evidence",
			Steps: []string{
				"1. brave-search:brave_web_search - Gather causal evidence",
				"2. unified-thinking:build-causal-graph - Create model",
				"3. unified-thinking:assess-evidence - Validate each causal link",
				"4. unified-thinking:probabilistic-reasoning - Quantify uncertainty",
				"5. memory:create_entities - Persist validated model",
			},
			UseCase: "Scientific or analytical reasoning requiring strong evidence",
			Servers: []string{"brave-search", "unified-thinking", "memory"},
		},
		{
			Name:        "Iterative Problem Refinement",
			Description: "Solve problems iteratively with continuous validation",
			Steps: []string{
				"1. unified-thinking:think - Initial reasoning",
				"2. unified-thinking:self-evaluate - Assess quality",
				"3. brave-search:brave_web_search - If quality low, gather more info",
				"4. unified-thinking:think - Refine with new information (challenge_assumptions: true)",
				"5. unified-thinking:validate - Check logical consistency",
			},
			UseCase: "Complex reasoning that benefits from iterative refinement",
			Servers: []string{"unified-thinking", "brave-search"},
		},
		{
			Name:        "Knowledge Discovery Pipeline",
			Description: "Comprehensive research with structured knowledge capture",
			Steps: []string{
				"1. brave-search:brave_web_search - Research topic",
				"2. obsidian:search-notes - Check existing knowledge",
				"3. unified-thinking:decompose-problem - Structure the investigation",
				"4. unified-thinking:think (tree mode) - Explore multiple angles",
				"5. unified-thinking:synthesize-insights - Combine findings",
				"6. memory:create_entities + create_relations - Build knowledge graph",
				"7. obsidian:create-note - Create comprehensive documentation",
			},
			UseCase: "Deep research projects requiring organization and knowledge retention",
			Servers: []string{"brave-search", "obsidian", "unified-thinking", "memory"},
		},
	}
}
