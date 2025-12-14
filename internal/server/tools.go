// Package server - Tool definitions for the MCP server
package server

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ToolDefinitions contains all MCP tool definitions for the unified thinking server.
// These are separated from the implementation for better maintainability.
var ToolDefinitions = []mcp.Tool{
	// Core Thinking Tools
	{
		Name: "think",
		Description: `Main thinking tool supporting multiple cognitive modes (linear, tree, divergent, auto).

**Parameters:**
- content (required): The thought to process
- mode: "linear" (step-by-step), "tree" (multi-branch), "divergent" (creative), "auto" (automatic selection)
- confidence: 0.0-1.0 (default: 0.8)
- key_points: Array of key observations
- branch_id: For tree mode continuation
- format_level: Response size control - "full" (default), "compact" (40-60% smaller), "minimal" (80% smaller)

**Returns:** thought_id, mode, confidence, metadata with:
- suggested_next_tools: Recommended next steps
- validation_opportunities: When to validate/research
- action_recommendations: Multi-tool workflows
- export_formats: Ready-to-use formats for other servers

**Works Well With:**
- Low confidence (<0.7) → brave-search:brave_web_search, conversation:conversation_search
- High confidence (≥0.8) → memory:create_entities, obsidian:create-note
- Tree mode → unified-thinking:synthesize-insights
- Linear mode → unified-thinking:validate

**Common Workflows:**
1. Research-Enhanced Thinking: brave_web_search → think → assess-evidence
2. Knowledge-Backed Reasoning: memory:traverse_graph → think (with context)
3. Documented Reasoning: think → obsidian:create-note
4. Validated Chain: think → validate → think (iterate)

**Example:** {"content": "Analyze database performance", "mode": "linear", "confidence": 0.7, "format_level": "compact"}`,
	},
	{
		Name:        "history",
		Description: "View thinking history",
	},
	{
		Name:        "list-branches",
		Description: "List all thinking branches",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
	{
		Name:        "focus-branch",
		Description: "Switch the active thinking branch",
	},
	{
		Name:        "branch-history",
		Description: "Get detailed history of a specific branch",
	},
	{
		Name:        "recent-branches",
		Description: "Get recently accessed branches for quick context switching",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},

	// Validation Tools
	{
		Name:        "validate",
		Description: "Validate a thought for logical consistency",
	},
	{
		Name:        "prove",
		Description: "Attempt to prove a logical conclusion from premises",
	},
	{
		Name:        "check-syntax",
		Description: "Validate syntax of logical statements",
	},

	// Search and Metadata Tools
	{
		Name:        "search",
		Description: "Search through all thoughts",
	},
	{
		Name:        "get-metrics",
		Description: "Get system performance and usage metrics",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
	{
		Name: "list-integration-patterns",
		Description: `List common multi-server workflow patterns for orchestrating tools across the MCP ecosystem.

**Returns:** Array of integration patterns, each with:
- name: Pattern identifier
- description: What the pattern does
- steps: Ordered list of tool calls with descriptions
- use_case: When to use this pattern
- servers: Which MCP servers are involved

**Patterns Included:**
- Research-Enhanced Thinking (brave-search + unified-thinking)
- Knowledge-Backed Decision Making (memory + conversation + unified-thinking + obsidian)
- Causal Model to Knowledge Graph (unified-thinking + memory)
- Problem Decomposition Workflow (unified-thinking + obsidian + brave-search)
- Temporal Decision Analysis (conversation + unified-thinking + obsidian)
- Stakeholder-Aware Planning (obsidian + unified-thinking + memory)
- Validated File Operations (unified-thinking + filesystem/windows-cli + obsidian)
- Evidence-Based Causal Reasoning (brave-search + unified-thinking + memory)
- Iterative Problem Refinement (unified-thinking + brave-search)
- Knowledge Discovery Pipeline (brave-search + obsidian + unified-thinking + memory)

**Use Cases:**
- Learning available integration patterns
- Choosing the right workflow for a task
- Understanding how to combine tools across servers
- Discovering new ways to use the MCP ecosystem

**Example:** {} (no parameters needed)`,
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},

	// Probabilistic Reasoning Tools
	{
		Name:        "probabilistic-reasoning",
		Description: "Perform Bayesian inference and update probabilistic beliefs based on evidence. Required: operation (\"create\", \"update\", \"get\", or \"combine\"). For create: statement, prior_prob (0-1). For update: belief_id, evidence_id, likelihood (0-1), evidence_prob (0-1). For get: belief_id. For combine: belief_ids (array), combine_op (\"and\" or \"or\"). Example: {\"operation\": \"create\", \"statement\": \"X is true\", \"prior_prob\": 0.5}",
	},
	{
		Name:        "assess-evidence",
		Description: "Assess the quality, reliability, and relevance of evidence for claims",
	},
	{
		Name:        "detect-contradictions",
		Description: "Detect contradictions among a set of thoughts or statements",
	},
	{
		Name:        "sensitivity-analysis",
		Description: "Test robustness of conclusions to changes in underlying assumptions",
	},

	// Decision & Problem-Solving Tools
	{
		Name: "make-decision",
		Description: `Create structured multi-criteria decision framework and recommendations.

**Parameters:**
- question (required): Decision question
- options (required): Array of options with id, name, description, scores, pros, cons
- criteria (required): Array of criteria with id, name, weight, maximize flag

**Returns:** decision with recommendation, confidence, and metadata with:
- export_formats.obsidian_note: Complete decision document in markdown
- action_recommendations: Persistence and execution suggestions
- validation_opportunities: Low confidence triggers validation suggestions

**Works Well With:**
- Before: decompose-problem (break down complex decisions)
- Before: analyze-temporal (understand time tradeoffs)
- After: obsidian:create-note (document decision)
- After: windows-cli:execute_command (implement decision)
- Validation: conversation:conversation_search (check similar past decisions)

**Common Workflows:**
1. Documented Decision: make-decision → obsidian:create-note (use export format)
2. Temporal Decision: analyze-temporal → make-decision → obsidian:create-note
3. Validated Decision: make-decision → conversation:conversation_search → make-decision (refine)
4. Action-Oriented: make-decision → windows-cli:execute_command

**Example:** {"question": "Which database?", "options": [{"id": "pg", "name": "PostgreSQL", "scores": {"cost": 0.8}}], "criteria": [{"id": "cost", "weight": 0.5}]}`,
	},
	{
		Name: "decompose-problem",
		Description: `Break down complex problems into manageable subproblems with dependencies.

**Parameters:**
- problem (required): Complex problem statement

**Returns:** decomposition with subproblems, dependencies, solution_path, and metadata with:
- suggested_next_tools: brave-search, obsidian:search-notes, think
- export_formats.obsidian_note: Problem breakdown as checklist

**Works Well With:**
- Research: brave-search:brave_web_search (research subproblems)
- Knowledge: obsidian:search-notes (find related solutions)
- Reasoning: unified-thinking:think (solve subproblems in order)
- Documentation: obsidian:create-note (track progress with checklist)

**Common Workflows:**
1. Research-Driven Solving: decompose-problem → brave_web_search (each subproblem) → think → synthesize-insights
2. Knowledge-Based Solving: decompose-problem → obsidian:search-notes → think (with context)
3. Tracked Progress: decompose-problem → obsidian:create-note (checklist) → update as solved
4. Team Collaboration: decompose-problem → memory:create_entities (subproblems as tasks)

**Example:** {"problem": "How to improve CI/CD pipeline performance?"}`,
	},

	// Metacognition Tools
	{
		Name:        "self-evaluate",
		Description: "Perform metacognitive self-assessment of reasoning quality and completeness",
	},
	{
		Name:        "detect-biases",
		Description: "Identify cognitive biases AND logical fallacies in reasoning (comprehensive analysis). Required: ONE OF content, thought_id, or branch_id. 'content' allows direct text analysis without storing a thought first. Detects cognitive biases (confirmation bias, anchoring, availability heuristic, sunk cost, overconfidence, recency bias) AND logical fallacies (ad hominem, straw man, affirming the consequent, etc.). Returns separate lists plus a unified 'combined' list. Example: {\"content\": \"Clearly this proves my theory. Obviously everyone knows this is true.\"} or {\"thought_id\": \"thought_123\"}",
	},
	{
		Name:        "detect-blind-spots",
		Description: "Detect unknown unknowns, blind spots, and knowledge gaps using metacognitive analysis. Parameters: content (required), domain, context, assumptions, confidence. Returns: blind_spots (array), missing_considerations, unchallenged_assumptions, suggested_questions, overall_risk, risk_level, analysis",
	},

	// Hallucination & Calibration Tools
	{
		Name: "verify-thought",
		Description: `Verify a thought for hallucinations using semantic uncertainty measurement.

**Parameters:**
- thought_id (required): ID of the thought to verify
- verification_level (optional): "fast", "deep", or "hybrid" (default: "hybrid")

**Returns:** HallucinationReport with:
- overall_risk: 0-1 score (higher = more likely hallucination)
- semantic_uncertainty: Breakdown of uncertainty types (aleatory, epistemic, confidence_mismatch)
- claims: Extracted factual claims with verification status
- verified_count: Number of verified claims
- hallucination_count: Number of likely hallucinations
- recommendations: Suggested actions for improvement

**Verification Levels:**
- fast: <100ms heuristic checks (confidence-content mismatch, uncertainty markers)
- deep: 1-5s with external knowledge sources (requires registered sources)
- hybrid: Fast check first, then async deep verification (default)

**Use Cases:**
- Verify thoughts before storing important decisions
- Check for confidence-content mismatches (e.g., high confidence + uncertain language)
- Identify potentially false factual claims
- Calibrate AI reasoning quality

**Example:** {"thought_id": "thought_123", "verification_level": "hybrid"}`,
	},
	{
		Name: "get-hallucination-report",
		Description: `Retrieve cached hallucination verification report for a thought.

**Parameters:**
- thought_id (required): ID of the thought

**Returns:** Previously generated HallucinationReport or error if not found

**Use Cases:**
- Check if a thought has been verified previously
- Retrieve verification results without re-running analysis
- Review historical verification reports

**Example:** {"thought_id": "thought_123"}`,
	},
	{
		Name: "record-prediction",
		Description: `Record a confidence prediction for calibration tracking.

**Parameters:**
- thought_id (required): ID of the thought
- confidence (required): Confidence score (0-1)
- mode (required): Thinking mode used
- metadata (optional): Additional metadata

**Returns:** Success status and recorded prediction

**Use Cases:**
- Track confidence predictions for later calibration analysis
- Build calibration history over time
- Enable confidence adjustment recommendations

**Example:** {"thought_id": "thought_123", "confidence": 0.8, "mode": "linear"}`,
	},
	{
		Name: "record-outcome",
		Description: `Record the actual outcome of a prediction for calibration.

**Parameters:**
- thought_id (required): ID of the thought (must have existing prediction)
- was_correct (required): Whether the thought was correct (boolean)
- actual_confidence (required): Actual confidence based on validation/verification (0-1)
- source (required): How outcome was determined ("validation", "verification", "user_feedback")
- metadata (optional): Additional context

**Returns:** Success status and recorded outcome

**Use Cases:**
- Record validation/verification results
- Track user feedback on thought quality
- Build calibration dataset for analysis

**Example:** {"thought_id": "thought_123", "was_correct": true, "actual_confidence": 0.9, "source": "validation"}`,
	},
	{
		Name: "get-calibration-report",
		Description: `Generate comprehensive confidence calibration report.

**Returns:** CalibrationReport with:
- total_predictions: Number of predictions tracked
- total_outcomes: Number of outcomes recorded
- buckets: Calibration by confidence range (0-10%, 10-20%, etc.)
- overall_accuracy: Actual success rate
- calibration: Expected Calibration Error (ECE)
- bias: Systematic over/underconfidence detection
- by_mode: Calibration breakdown by thinking mode
- recommendations: Actionable calibration improvements

**Use Cases:**
- Assess confidence calibration quality
- Detect systematic overconfidence or underconfidence
- Get specific recommendations for confidence adjustment
- Track calibration improvement over time

**Example:** {}`,
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},

	// Temporal & Perspective Tools
	{
		Name: "analyze-perspectives",
		Description: `Analyze a situation from multiple stakeholder perspectives, identifying concerns, priorities, and conflicts.

**Parameters:**
- situation (required): Situation to analyze
- stakeholder_hints: Array of stakeholder types/names (optional)

**Returns:** perspectives array with viewpoint, concerns, priorities, and metadata with:
- suggested_next_tools: obsidian:search-notes, brave-search, synthesize-insights
- export_formats.memory_entities: Stakeholders ready for Memory KG

**Works Well With:**
- Research: obsidian:search-notes (find stakeholder documentation)
- Research: brave-search:brave_web_search (research stakeholder positions)
- Synthesis: unified-thinking:synthesize-insights (find common ground)
- Persistence: memory:create_entities (store stakeholder profiles)

**Common Workflows:**
1. Stakeholder Research: obsidian:search-notes → analyze-perspectives → brave_web_search (validate)
2. Knowledge Base: analyze-perspectives → memory:create_entities (use export format)
3. Conflict Resolution: analyze-perspectives → synthesize-insights → make-decision
4. Documentation: analyze-perspectives → obsidian:create-note (stakeholder map)

**Example:** {"situation": "Implementing new auth system", "stakeholder_hints": ["developers", "security team", "users"]}`,
	},
	{
		Name: "analyze-temporal",
		Description: `Analyze short-term vs long-term implications of a decision, identifying tradeoffs and providing recommendations.

**Parameters:**
- situation (required): Decision or situation to analyze
- time_horizon: "days-weeks", "months", "years" (default: "months")

**Returns:** analysis with short/long-term views, tradeoffs, recommendation, and metadata with:
- suggested_next_tools: conversation:conversation_search (historical context)
- export_formats.obsidian_note: Structured temporal analysis document
- action_recommendations: Documentation suggestions

**Works Well With:**
- Before: conversation:conversation_search (check similar past decisions)
- After: make-decision (incorporate temporal insights)
- After: obsidian:create-note (document temporal analysis)
- Integration: memory:traverse_graph (connect to related temporal patterns)

**Common Workflows:**
1. Historical Context: conversation:conversation_search → analyze-temporal → make-decision
2. Documented Analysis: analyze-temporal → obsidian:create-note (use export format)
3. Pattern Recognition: analyze-temporal → memory:create_entities (temporal patterns)
4. Decision Support: analyze-temporal → make-decision → obsidian:create-note

**Example:** {"situation": "Refactor now or after release?", "time_horizon": "months"}`,
	},
	{
		Name:        "compare-time-horizons",
		Description: "Compare how a decision looks across different time horizons (days-weeks, months, years)",
	},
	{
		Name:        "identify-optimal-timing",
		Description: "Determine optimal timing for a decision based on situation and constraints",
	},

	// Causal Reasoning Tools
	{
		Name: "build-causal-graph",
		Description: `Construct a causal graph from observations, identifying variables and causal relationships.

**Parameters:**
- description (required): Context for the causal model
- observations (required): Array of causal statements

**Returns:** graph with variables, links, and metadata with:
- suggested_next_tools: memory:create_entities, simulate-intervention
- export_formats.memory_entities: Variables ready for Memory KG
- export_formats.memory_relations: Causal links ready for Memory KG

**Works Well With:**
- Before: brave-search:brave_web_search (gather causal evidence)
- After: memory:create_entities (persist causal model)
- After: unified-thinking:simulate-intervention (test interventions)
- Validation: brave-search (verify causal claims)

**Common Workflows:**
1. Research-Based Causal Model: brave_web_search → build-causal-graph → memory:create_entities
2. Intervention Planning: build-causal-graph → simulate-intervention → make-decision
3. Knowledge Integration: memory:traverse_graph → build-causal-graph (enrich with existing knowledge)

**Example:** {"description": "Sales process", "observations": ["Marketing increases awareness", "Awareness drives sales"]}`,
	},
	{
		Name:        "simulate-intervention",
		Description: "Simulate the effects of intervening on a variable in a causal graph",
	},
	{
		Name:        "generate-counterfactual",
		Description: "Generate a counterfactual scenario ('what if') by changing variables in a causal model",
	},
	{
		Name:        "analyze-correlation-vs-causation",
		Description: "Analyze whether an observed relationship is likely correlation or causation",
	},
	{
		Name:        "get-causal-graph",
		Description: "Retrieve a previously built causal graph by ID",
	},

	// Integration & Synthesis Tools
	{
		Name:        "synthesize-insights",
		Description: "Synthesize insights from multiple reasoning modes, identifying synergies and conflicts",
	},
	{
		Name:        "detect-emergent-patterns",
		Description: "Detect emergent patterns that become visible when combining multiple reasoning modes",
	},

	// Workflow Orchestration Tools
	{
		Name:        "execute-workflow",
		Description: "Execute a predefined reasoning workflow with automatic tool chaining",
	},
	{
		Name:        "list-workflows",
		Description: "List all available reasoning workflows",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
	{
		Name:        "register-workflow",
		Description: "Register a new custom workflow for automated tool coordination",
	},

	// Dual-Process Reasoning
	{
		Name:        "dual-process-think",
		Description: "Execute dual-process reasoning (System 1: fast/intuitive, System 2: slow/analytical). Auto-detects complexity and escalates as needed. Parameters: content (required), mode, branch_id, force_system ('system1'/'system2'), key_points. Returns: thought_id, system_used, complexity, escalated, timings, confidence",
	},

	// Backtracking Tools
	{
		Name:        "create-checkpoint",
		Description: "Create a backtracking checkpoint in tree mode. Save current branch state for later restoration. Parameters: branch_id (required), name (required), description. Returns: checkpoint_id, thought_count, insight_count, created_at",
	},
	{
		Name:        "restore-checkpoint",
		Description: "Restore branch from a checkpoint. Enables backtracking in tree exploration. Parameters: checkpoint_id (required). Returns: branch_id, thought_count, insight_count, message",
	},
	{
		Name:        "list-checkpoints",
		Description: "List available checkpoints for backtracking. Parameters: branch_id (optional - filter by branch). Returns: array of checkpoints with id, name, description, branch_id, thought_count, created_at",
	},

	// Abductive Reasoning Tools
	{
		Name:        "generate-hypotheses",
		Description: "Generate hypotheses from observations using abductive reasoning (inference to best explanation). Parameters: observations (array of {description, confidence}), max_hypotheses, min_parsimony. Returns: array of hypotheses with id, description, parsimony, prior_probability",
	},
	{
		Name:        "evaluate-hypotheses",
		Description: "Evaluate and rank hypotheses using Bayesian inference, parsimony, and explanatory power. Parameters: observations (required), hypotheses (required), method ('bayesian'/'parsimony'/'combined'). Returns: ranked_hypotheses with posterior_probability, explanatory_power, parsimony scores",
	},

	// Case-Based Reasoning Tools
	{
		Name:        "retrieve-similar-cases",
		Description: "Retrieve similar cases from case library using CBR (case-based reasoning). Parameters: problem {description, context, goals, constraints}, domain, max_cases, min_similarity. Returns: array of similar cases with similarity scores, solutions, success_rates",
	},
	{
		Name:        "perform-cbr-cycle",
		Description: "Perform full CBR cycle: Retrieve similar cases, Reuse/adapt solution, provide recommendations. Parameters: problem {description, context, goals, constraints}, domain. Returns: retrieved_count, best_case, adapted_solution, strategy, confidence",
	},

	// Symbolic Reasoning Tools
	{
		Name:        "prove-theorem",
		Description: "Attempt to prove a theorem using symbolic reasoning and logical inference rules (modus ponens, simplification, conjunction). Parameters: name, premises (array), conclusion. Returns: status, is_valid, confidence, proof with steps (step_number, statement, justification, rule, dependencies)",
	},
	{
		Name:        "check-constraints",
		Description: "Check consistency of symbolic constraints. Detect conflicts and contradictions. Parameters: symbols (array of {name, type, domain}), constraints (array of {type, expression, symbols}). Returns: is_consistent, conflicts (array), explanation",
	},

	// Enhanced Tools
	{
		Name:        "find-analogy",
		Description: "Find analogies between source and target domains for cross-domain reasoning. Required: source_domain (string), target_problem (string). Optional: constraints (array). Example: {\"source_domain\": \"biology: immune system\", \"target_problem\": \"How to protect computer network?\"}",
	},
	{
		Name:        "apply-analogy",
		Description: "Apply an existing analogy to a new context. Required: analogy_id (from find-analogy), target_context (string). Example: {\"analogy_id\": \"analogy_123\", \"target_context\": \"New security scenario\"}",
	},
	{
		Name:        "decompose-argument",
		Description: "Break down an argument into premises, claims, assumptions, and inference chains. Required: argument (string). Example: {\"argument\": \"We should adopt policy X because studies show it reduces Y by 30%\"}",
	},
	{
		Name:        "generate-counter-arguments",
		Description: "Generate counter-arguments for a given argument using multiple strategies. Required: argument_id (from decompose-argument). Example: {\"argument_id\": \"arg_123\"}",
	},
	{
		Name:        "detect-fallacies",
		Description: "Detect formal and informal logical fallacies in reasoning. Required: content (string). Optional: check_formal (bool), check_informal (bool). Detects ad hominem, strawman, false dichotomy, etc. NOTE: For cognitive biases (confirmation bias, anchoring, etc.), use detect-biases instead. Example: {\"content\": \"Everyone says X is true, so it must be\", \"check_formal\": true, \"check_informal\": true}",
	},
	{
		Name:        "process-evidence-pipeline",
		Description: "Process evidence and auto-update beliefs, causal graphs, and decisions. Required: content (string), source (string), supports_claim (bool). Optional: claim_id. Example: {\"content\": \"Study shows X increases Y by 30%\", \"source\": \"Journal of Science 2024\", \"supports_claim\": true}",
	},
	{
		Name:        "analyze-temporal-causal-effects",
		Description: "Analyze how causal effects evolve across different time horizons. Required: graph_id (from build-causal-graph), variable_id, intervention_type (\"increase\"/\"decrease\"/\"remove\"/\"introduce\"). Example: {\"graph_id\": \"graph_123\", \"variable_id\": \"marketing_spend\", \"intervention_type\": \"increase\"}",
	},
	{
		Name:        "analyze-decision-timing",
		Description: "Determine optimal timing for decisions based on causal and temporal analysis. Required: situation (string). Optional: causal_graph_id. Example: {\"situation\": \"When to launch product?\", \"causal_graph_id\": \"graph_123\"}",
	},

	// Episodic Memory & Learning Tools
	{
		Name: "start-reasoning-session",
		Description: `Start tracking a reasoning session to build episodic memory and learn from experience.

The episodic memory system enables the server to learn from past reasoning sessions,
recognize successful patterns, and provide adaptive recommendations.

**Parameters:**
- session_id (required): Unique session identifier
- description (required): Problem description
- goals (optional): Array of goals to achieve
- domain (optional): Problem domain (e.g., "software-engineering", "science", "business")
- context (optional): Additional context about the problem
- complexity (optional): Estimated complexity 0.0-1.0
- metadata (optional): Additional metadata

**Returns:**
- session_id: Session identifier
- problem_id: Problem fingerprint hash
- status: "active"
- suggestions: Array of recommendations based on similar past problems

**Use Cases:**
1. Before complex reasoning: Get suggestions from similar past successes
2. Learning from failures: System warns about approaches that historically fail
3. Continuous improvement: Performance improves with every reasoning session

**Example:**
{
  "session_id": "debug_2024_001",
  "description": "Optimize database query performance",
  "goals": ["Reduce query time", "Improve user experience"],
  "domain": "software-engineering",
  "complexity": 0.6
}`,
	},
	{
		Name: "complete-reasoning-session",
		Description: `Complete a reasoning session and store the trajectory for learning.

Marks a session as complete, calculates quality metrics, and triggers pattern learning.
The system learns which approaches work best for different problem types.

**Parameters:**
- session_id (required): Session to complete
- status (required): "success", "partial", or "failure"
- goals_achieved (optional): Array of achieved goals
- goals_failed (optional): Array of failed goals
- solution (optional): Description of solution
- confidence (optional): Confidence in solution (0.0-1.0)
- unexpected_outcomes (optional): Array of unexpected results

**Returns:**
- trajectory_id: Stored trajectory identifier
- session_id: Session identifier
- success_score: Calculated success score (0.0-1.0)
- quality_score: Overall quality score (0.0-1.0)
- patterns_found: Number of patterns updated
- status: "completed"

**Quality Metrics Calculated:**
- Efficiency: Steps taken vs optimal
- Coherence: Logical consistency
- Completeness: Goal coverage
- Innovation: Creative tool usage
- Reliability: Confidence in result

**Example:**
{
  "session_id": "debug_2024_001",
  "status": "success",
  "goals_achieved": ["Reduce query time"],
  "solution": "Added indexes and optimized queries",
  "confidence": 0.85
}`,
	},
	{
		Name: "get-recommendations",
		Description: `Get adaptive recommendations based on episodic memory of similar past problems.

Retrieves recommendations from the episodic memory system based on similarity to past
successful reasoning sessions. Includes learned patterns and historical success rates.

**Parameters:**
- description (required): Problem description
- goals (optional): Array of problem goals
- domain (optional): Problem domain
- context (optional): Additional context
- complexity (optional): Estimated complexity (0.0-1.0)
- limit (optional): Max recommendations (default: 5)

**Returns:**
- recommendations: Array of recommendations with:
  - type: "tool_sequence", "approach", "warning", or "optimization"
  - priority: Relevance score
  - suggestion: Specific advice
  - reasoning: Why this recommendation
  - success_rate: Historical success rate
- similar_cases: Count of similar past trajectories
- learned_patterns: Applicable learned patterns
- count: Number of recommendations

**Recommendation Types:**
1. **tool_sequence**: Proven tool sequences (success rate >70%)
2. **approach**: Successful reasoning strategies
3. **warning**: Approaches that historically fail (<40% success)
4. **optimization**: Performance improvements

**Example:**
{
  "description": "Need to implement user authentication",
  "domain": "security",
  "goals": ["Secure login", "Session management"],
  "limit": 3
}`,
	},
	{
		Name: "search-trajectories",
		Description: `Search for past reasoning trajectories to learn from experience.

Find past reasoning sessions by domain, tags, success rate, or problem type. Useful for
understanding what worked in the past and learning from both successes and failures.

**Parameters:**
- domain (optional): Filter by domain
- tags (optional): Array of tags to filter by
- min_success (optional): Minimum success score (0.0-1.0)
- problem_type (optional): Filter by problem type
- limit (optional): Max results (default: 10)

**Returns:**
- trajectories: Array of trajectory summaries with:
  - id: Trajectory identifier
  - session_id: Original session ID
  - problem: Problem description
  - domain: Problem domain
  - strategy: Strategy used
  - tools_used: Array of tools used
  - success_score: Success score (0.0-1.0)
  - duration: Session duration
  - tags: Array of tags
- count: Number of results

**Use Cases:**
1. Review successful approaches for a domain
2. Learn from failures (min_success: 0.0-0.4)
3. Find high-performing strategies (min_success: 0.8-1.0)
4. Analyze tool usage patterns

**Example:**
{
  "domain": "software-engineering",
  "min_success": 0.7,
  "limit": 5
}`,
	},
	{
		Name: "analyze-trajectory",
		Description: `Perform retrospective analysis of a completed reasoning session.

Provides comprehensive post-session analysis including strengths, weaknesses, actionable
improvements, lessons learned, and comparative analysis against similar past sessions.

**Parameters:**
- trajectory_id (required): ID of trajectory to analyze (returned from complete-reasoning-session)

**Returns:**
- summary: High-level assessment with success/quality scores, duration, strategy
- strengths: What went well (efficiency, coherence, completeness, innovation, reliability)
- weaknesses: Areas for improvement with specific metrics
- improvements: Prioritized actionable suggestions with expected impact
- lessons_learned: Key takeaways for future sessions
- comparative_analysis: How this session compares to similar past sessions (percentile rank)
- detailed_metrics: Deep dive into each quality metric with explanations and suggestions

**Quality Metrics Analyzed:**
1. **Efficiency**: Steps taken vs optimal (7-10 steps baseline)
2. **Coherence**: Logical consistency (contradictions, fallacies)
3. **Completeness**: Goal achievement rate
4. **Innovation**: Use of creative/advanced tools
5. **Reliability**: Confidence in results

**Improvement Categories:**
- efficiency: Reduce unnecessary steps
- quality: Improve logical consistency
- approach: Change reasoning strategy
- tools: Use different/better tools

**Use Cases:**
1. Learn from successful sessions - understand what worked
2. Improve future performance - get specific actionable advice
3. Track progress - see percentile rank vs similar problems
4. Identify patterns - discover your reasoning strengths/weaknesses

**Example:**
{
  "trajectory_id": "traj_session_001_problem_abc_1234567890"
}

**Returns comprehensive analysis including:**
- Overall assessment: "excellent", "good", "fair", or "poor"
- Top 3-5 strengths with metrics
- Top 3-5 weaknesses with root causes
- Prioritized improvement suggestions
- Percentile rank (e.g., "better than 75% of similar sessions")
- Detailed metric breakdowns with actionable next steps`,
	},
}

// GetToolByName returns a tool definition by name
func GetToolByName(name string) (*mcp.Tool, bool) {
	for _, tool := range ToolDefinitions {
		if tool.Name == name {
			return &tool, true
		}
	}
	return nil, false
}

// GetToolNames returns all tool names for easy reference
func GetToolNames() []string {
	names := make([]string, len(ToolDefinitions))
	for i, tool := range ToolDefinitions {
		names[i] = tool.Name
	}
	return names
}
