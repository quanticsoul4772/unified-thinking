// Package server implements the MCP (Model Context Protocol) server for unified thinking.
//
// This package provides the MCP server implementation that exposes 63 tools for
// thought processing, validation, search, and advanced cognitive reasoning. All
// responses are JSON formatted for consumption by Claude AI via stdio transport.
//
// Core Tools (11):
//   - think, history, list-branches, focus-branch, branch-history, recent-branches
//   - validate, prove, check-syntax, search, get-metrics
//
// Probabilistic & Evidence Tools (4):
//   - probabilistic-reasoning, assess-evidence, detect-contradictions, sensitivity-analysis
//
// Decision & Problem-Solving Tools (3):
//   - make-decision, decompose-problem, verify-thought
//
// Metacognition Tools (3):
//   - self-evaluate, detect-biases, detect-blind-spots
//
// Hallucination & Calibration Tools (4):
//   - get-hallucination-report, record-prediction, record-outcome, get-calibration-report
//
// Temporal & Perspective Tools (4):
//   - analyze-perspectives, analyze-temporal, compare-time-horizons, identify-optimal-timing
//
// Causal Reasoning Tools (5):
//   - build-causal-graph, simulate-intervention, generate-counterfactual
//   - analyze-correlation-vs-causation, get-causal-graph
//
// Integration & Synthesis Tools (6):
//   - synthesize-insights, detect-emergent-patterns
//   - execute-workflow, list-workflows, register-workflow, list-integration-patterns
//
// Advanced Reasoning Tools (10):
//   - dual-process-think, create-checkpoint, restore-checkpoint, list-checkpoints
//   - generate-hypotheses, evaluate-hypotheses, retrieve-similar-cases, perform-cbr-cycle
//   - prove-theorem, check-constraints
//
// Enhanced Tools (8):
//   - find-analogy, apply-analogy
//   - decompose-argument, generate-counter-arguments, detect-fallacies
//   - process-evidence-pipeline, analyze-temporal-causal-effects, analyze-decision-timing
//
// Episodic Memory & Learning Tools (5):
//   - start-reasoning-session, complete-reasoning-session
//   - get-recommendations, search-trajectories, analyze-trajectory
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/integration"
	"unified-thinking/internal/memory"
	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/processing"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/server/handlers"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// UnifiedServer coordinates all thinking modes and provides MCP tool handlers.
type UnifiedServer struct {
	storage               storage.Storage
	linear                *modes.LinearMode
	tree                  *modes.TreeMode
	divergent             *modes.DivergentMode
	auto                  *modes.AutoMode
	validator             *validation.LogicValidator
	probabilisticReasoner *reasoning.ProbabilisticReasoner
	evidenceAnalyzer      *analysis.EvidenceAnalyzer
	contradictionDetector *analysis.ContradictionDetector
	decisionMaker         *reasoning.DecisionMaker
	problemDecomposer     *reasoning.ProblemDecomposer
	sensitivityAnalyzer   *analysis.SensitivityAnalyzer
	selfEvaluator         *metacognition.SelfEvaluator
	biasDetector          *metacognition.BiasDetector
	fallacyDetector       *validation.FallacyDetector
	// Phase 1: Handler delegates
	probabilisticHandler *handlers.ProbabilisticHandler
	decisionHandler      *handlers.DecisionHandler
	metacognitionHandler *handlers.MetacognitionHandler
	// Phase 2: Handler delegates
	temporalHandler *handlers.TemporalHandler
	causalHandler   *handlers.CausalHandler
	// Phase 2-3: Advanced reasoning modules
	perspectiveAnalyzer *analysis.PerspectiveAnalyzer
	temporalReasoner    *reasoning.TemporalReasoner
	causalReasoner      *reasoning.CausalReasoner
	synthesizer         *integration.Synthesizer
	// Workflow orchestration
	orchestrator *orchestration.Orchestrator
	// Hallucination detection (Phase 1 implementation)
	hallucinationHandler *handlers.HallucinationHandler
	// Confidence calibration tracking (Phase 1 implementation)
	calibrationHandler *handlers.CalibrationHandler
	// Phase 2-3: New reasoning handlers
	dualProcessHandler     *handlers.DualProcessHandler
	backtrackingHandler    *handlers.BacktrackingHandler
	abductiveHandler       *handlers.AbductiveHandler
	caseBasedHandler       *handlers.CaseBasedHandler
	unknownUnknownsHandler *handlers.UnknownUnknownsHandler
	symbolicHandler        *handlers.SymbolicHandler
	// Enhanced tools components
	analogicalReasoner        *reasoning.AnalogicalReasoner
	argumentAnalyzer          *analysis.ArgumentAnalyzer
	evidencePipeline          *integration.EvidencePipeline
	causalTemporalIntegration *integration.CausalTemporalIntegration
	// Episodic memory system (Phase 2)
	episodicMemoryHandler *handlers.EpisodicMemoryHandler
}

func NewUnifiedServer(
	store storage.Storage,
	linear *modes.LinearMode,
	tree *modes.TreeMode,
	divergent *modes.DivergentMode,
	auto *modes.AutoMode,
	validator *validation.LogicValidator,
) *UnifiedServer {
	// Initialize core reasoning engines
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()
	decisionMaker := reasoning.NewDecisionMaker()
	problemDecomposer := reasoning.NewProblemDecomposer()
	sensitivityAnalyzer := analysis.NewSensitivityAnalyzer()
	perspectiveAnalyzer := analysis.NewPerspectiveAnalyzer()
	temporalReasoner := reasoning.NewTemporalReasoner()
	causalReasoner := reasoning.NewCausalReasoner()

	s := &UnifiedServer{
		storage:               store,
		linear:                linear,
		tree:                  tree,
		divergent:             divergent,
		auto:                  auto,
		validator:             validator,
		probabilisticReasoner: probabilisticReasoner,
		evidenceAnalyzer:      evidenceAnalyzer,
		contradictionDetector: contradictionDetector,
		decisionMaker:         decisionMaker,
		problemDecomposer:     problemDecomposer,
		sensitivityAnalyzer:   sensitivityAnalyzer,
		selfEvaluator:         metacognition.NewSelfEvaluator(),
		biasDetector:          metacognition.NewBiasDetector(),
		fallacyDetector:       validation.NewFallacyDetector(),
		// Phase 1: Initialize handler delegates
		probabilisticHandler: handlers.NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector),
		decisionHandler:      handlers.NewDecisionHandler(store, decisionMaker, problemDecomposer, sensitivityAnalyzer),
		metacognitionHandler: handlers.NewMetacognitionHandler(store, metacognition.NewSelfEvaluator(), metacognition.NewBiasDetector(), validation.NewFallacyDetector()),
		// Phase 2: Initialize temporal handler delegate
		temporalHandler: handlers.NewTemporalHandler(perspectiveAnalyzer, temporalReasoner),
		// Phase 2-3: Initialize advanced reasoning modules
		perspectiveAnalyzer: perspectiveAnalyzer,
		temporalReasoner:    temporalReasoner,
		causalReasoner:      causalReasoner,
		// Phase 2: Initialize causal handler delegate (FIXED: reuse causalReasoner instance)
		causalHandler: handlers.NewCausalHandler(causalReasoner),
		synthesizer:   integration.NewSynthesizer(),
		// Initialize hallucination handler
		hallucinationHandler: handlers.NewHallucinationHandler(store),
		// Initialize calibration tracker
		calibrationHandler: handlers.NewCalibrationHandler(),
	}

	// Initialize Phase 2-3 handlers
	s.initializeAdvancedHandlers()

	return s
}

// initializeAdvancedHandlers initializes Phase 2-3 reasoning handlers
func (s *UnifiedServer) initializeAdvancedHandlers() {
	// Dual-process executor
	dualProcessExecutor := processing.NewDualProcessExecutor(s.storage, map[types.ThinkingMode]modes.ThinkingMode{
		types.ModeLinear:    s.linear,
		types.ModeTree:      s.tree,
		types.ModeDivergent: s.divergent,
	})
	s.dualProcessHandler = handlers.NewDualProcessHandler(dualProcessExecutor, s.storage)

	// Backtracking manager
	backtrackingManager := modes.NewBacktrackingManager(s.storage)
	s.backtrackingHandler = handlers.NewBacktrackingHandler(backtrackingManager, s.storage)

	// Abductive reasoner
	abductiveReasoner := reasoning.NewAbductiveReasoner(s.storage)
	s.abductiveHandler = handlers.NewAbductiveHandler(abductiveReasoner, s.storage)

	// Case-based reasoner
	caseBasedReasoner := reasoning.NewCaseBasedReasoner(s.storage)
	s.caseBasedHandler = handlers.NewCaseBasedHandler(caseBasedReasoner, s.storage)

	// Unknown unknowns detector
	unknownUnknownsDetector := metacognition.NewUnknownUnknownsDetector()
	s.unknownUnknownsHandler = handlers.NewUnknownUnknownsHandler(unknownUnknownsDetector, s.storage)

	// Symbolic reasoner
	symbolicReasoner := validation.NewSymbolicReasoner()
	s.symbolicHandler = handlers.NewSymbolicHandler(symbolicReasoner, s.storage)

	// Enhanced tools components
	s.analogicalReasoner = reasoning.NewAnalogicalReasoner()
	s.argumentAnalyzer = analysis.NewArgumentAnalyzer()
	s.evidencePipeline = integration.NewEvidencePipeline(
		s.probabilisticReasoner,
		s.causalReasoner,
		s.decisionMaker,
		s.evidenceAnalyzer,
	)
	s.causalTemporalIntegration = integration.NewCausalTemporalIntegration(
		s.causalReasoner,
		s.temporalReasoner,
	)

	// Initialize episodic memory system (Phase 2)
	s.initializeEpisodicMemory()
}

// initializeEpisodicMemory initializes the episodic reasoning memory system
func (s *UnifiedServer) initializeEpisodicMemory() {
	// Create episodic memory store
	store := memory.NewEpisodicMemoryStore()

	// Create session tracker
	tracker := memory.NewSessionTracker(store)

	// Create learning engine
	learner := memory.NewLearningEngine(store)

	// Create episodic memory handler
	s.episodicMemoryHandler = handlers.NewEpisodicMemoryHandler(store, tracker, learner)
}

// SetOrchestrator sets the workflow orchestrator for the server
// This is a separate method to handle circular dependency between server and orchestrator
func (s *UnifiedServer) SetOrchestrator(orchestrator *orchestration.Orchestrator) {
	s.orchestrator = orchestrator
}

func (s *UnifiedServer) RegisterTools(mcpServer *mcp.Server) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "think",
		Description: `Main thinking tool supporting multiple cognitive modes (linear, tree, divergent, auto).

**Parameters:**
- content (required): The thought to process
- mode: "linear" (step-by-step), "tree" (multi-branch), "divergent" (creative), "auto" (automatic selection)
- confidence: 0.0-1.0 (default: 0.8)
- key_points: Array of key observations
- branch_id: For tree mode continuation

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

**Example:** {"content": "Analyze database performance", "mode": "linear", "confidence": 0.7}`,
	}, s.handleThink)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "history",
		Description: "View thinking history",
	}, s.handleHistory)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list-branches",
		Description: "List all thinking branches",
	}, s.handleListBranches)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "focus-branch",
		Description: "Switch the active thinking branch",
	}, s.handleFocusBranch)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "branch-history",
		Description: "Get detailed history of a specific branch",
	}, s.handleBranchHistory)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "validate",
		Description: "Validate a thought for logical consistency",
	}, s.handleValidate)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "prove",
		Description: "Attempt to prove a logical conclusion from premises",
	}, s.handleProve)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "check-syntax",
		Description: "Validate syntax of logical statements",
	}, s.handleCheckSyntax)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "search",
		Description: "Search through all thoughts",
	}, s.handleSearch)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "get-metrics",
		Description: "Get system performance and usage metrics",
	}, s.handleGetMetrics)

	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleListIntegrationPatterns)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "recent-branches",
		Description: "Get recently accessed branches for quick context switching",
	}, s.handleRecentBranches)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "probabilistic-reasoning",
		Description: "Perform Bayesian inference and update probabilistic beliefs based on evidence. Required: operation (\"create\", \"update\", \"get\", or \"combine\"). For create: statement, prior_prob (0-1). For update: belief_id, evidence_id, likelihood (0-1), evidence_prob (0-1). For get: belief_id. For combine: belief_ids (array), combine_op (\"and\" or \"or\"). Example: {\"operation\": \"create\", \"statement\": \"X is true\", \"prior_prob\": 0.5}",
	}, s.handleProbabilisticReasoning)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "assess-evidence",
		Description: "Assess the quality, reliability, and relevance of evidence for claims",
	}, s.handleAssessEvidence)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "detect-contradictions",
		Description: "Detect contradictions among a set of thoughts or statements",
	}, s.handleDetectContradictions)

	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleMakeDecision)

	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleDecomposeProblem)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "sensitivity-analysis",
		Description: "Test robustness of conclusions to changes in underlying assumptions",
	}, s.handleSensitivityAnalysis)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "self-evaluate",
		Description: "Perform metacognitive self-assessment of reasoning quality and completeness",
	}, s.handleSelfEvaluate)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "detect-biases",
		Description: "Identify cognitive biases AND logical fallacies in reasoning (comprehensive analysis). Required: EITHER thought_id OR branch_id (not both). Detects cognitive biases (confirmation bias, anchoring, availability heuristic, etc.) AND logical fallacies (ad hominem, straw man, affirming the consequent, etc.). Returns separate lists plus a unified 'combined' list. Example: {\"thought_id\": \"thought_123\"}",
	}, s.handleDetectBiases)

	// Phase 1: Hallucination Detection Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleVerifyThought)

	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleGetHallucinationReport)

	// Phase 1: Confidence Calibration Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleRecordPrediction)

	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleRecordOutcome)

	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleGetCalibrationReport)

	// Phase 2: Multi-Perspective Analysis Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleAnalyzePerspectives)

	// Phase 2: Temporal Reasoning Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleAnalyzeTemporal)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "compare-time-horizons",
		Description: "Compare how a decision looks across different time horizons (days-weeks, months, years)",
	}, s.handleCompareTimeHorizons)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "identify-optimal-timing",
		Description: "Determine optimal timing for a decision based on situation and constraints",
	}, s.handleIdentifyOptimalTiming)

	// Phase 3: Causal Reasoning Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
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
	}, s.handleBuildCausalGraph)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "simulate-intervention",
		Description: "Simulate the effects of intervening on a variable in a causal graph",
	}, s.handleSimulateIntervention)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "generate-counterfactual",
		Description: "Generate a counterfactual scenario ('what if') by changing variables in a causal model",
	}, s.handleGenerateCounterfactual)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "analyze-correlation-vs-causation",
		Description: "Analyze whether an observed relationship is likely correlation or causation",
	}, s.handleAnalyzeCorrelationVsCausation)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "get-causal-graph",
		Description: "Retrieve a previously built causal graph by ID",
	}, s.handleGetCausalGraph)

	// Phase 3: Cross-Mode Synthesis Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "synthesize-insights",
		Description: "Synthesize insights from multiple reasoning modes, identifying synergies and conflicts",
	}, s.handleSynthesizeInsights)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "detect-emergent-patterns",
		Description: "Detect emergent patterns that become visible when combining multiple reasoning modes",
	}, s.handleDetectEmergentPatterns)

	// Workflow orchestration tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "execute-workflow",
		Description: "Execute a predefined workflow that coordinates multiple reasoning tools automatically. Required: workflow_id (string), input (object with workflow parameters). Common workflows: \"comprehensive-analysis\", \"validation-pipeline\". Input must include \"problem\" field. Use list-workflows to see available workflows. Example: {\"workflow_id\": \"comprehensive-analysis\", \"input\": {\"problem\": \"Optimize system\"}}",
	}, s.handleExecuteWorkflow)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list-workflows",
		Description: "List all available automated workflows for multi-tool reasoning pipelines",
	}, s.handleListWorkflows)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "register-workflow",
		Description: "Register a new custom workflow for automated tool coordination",
	}, s.handleRegisterWorkflow)

	// Phase 2-3: Advanced reasoning tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "dual-process-think",
		Description: "Execute dual-process reasoning (System 1: fast/intuitive, System 2: slow/analytical). Auto-detects complexity and escalates as needed. Parameters: content (required), mode, branch_id, force_system ('system1'/'system2'), key_points. Returns: thought_id, system_used, complexity, escalated, timings, confidence",
	}, s.handleDualProcessThink)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "create-checkpoint",
		Description: "Create a backtracking checkpoint in tree mode. Save current branch state for later restoration. Parameters: branch_id (required), name (required), description. Returns: checkpoint_id, thought_count, insight_count, created_at",
	}, s.handleCreateCheckpoint)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "restore-checkpoint",
		Description: "Restore branch from a checkpoint. Enables backtracking in tree exploration. Parameters: checkpoint_id (required). Returns: branch_id, thought_count, insight_count, message",
	}, s.handleRestoreCheckpoint)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list-checkpoints",
		Description: "List available checkpoints for backtracking. Parameters: branch_id (optional - filter by branch). Returns: array of checkpoints with id, name, description, branch_id, thought_count, created_at",
	}, s.handleListCheckpoints)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "generate-hypotheses",
		Description: "Generate hypotheses from observations using abductive reasoning (inference to best explanation). Parameters: observations (array of {description, confidence}), max_hypotheses, min_parsimony. Returns: array of hypotheses with id, description, parsimony, prior_probability",
	}, s.handleGenerateHypotheses)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "evaluate-hypotheses",
		Description: "Evaluate and rank hypotheses using Bayesian inference, parsimony, and explanatory power. Parameters: observations (required), hypotheses (required), method ('bayesian'/'parsimony'/'combined'). Returns: ranked_hypotheses with posterior_probability, explanatory_power, parsimony scores",
	}, s.handleEvaluateHypotheses)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "retrieve-similar-cases",
		Description: "Retrieve similar cases from case library using CBR (case-based reasoning). Parameters: problem {description, context, goals, constraints}, domain, max_cases, min_similarity. Returns: array of similar cases with similarity scores, solutions, success_rates",
	}, s.handleRetrieveCases)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "perform-cbr-cycle",
		Description: "Perform full CBR cycle: Retrieve similar cases, Reuse/adapt solution, provide recommendations. Parameters: problem {description, context, goals, constraints}, domain. Returns: retrieved_count, best_case, adapted_solution, strategy, confidence",
	}, s.handlePerformCBRCycle)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "detect-blind-spots",
		Description: "Detect unknown unknowns, blind spots, and knowledge gaps using metacognitive analysis. Parameters: content (required), domain, context, assumptions, confidence. Returns: blind_spots (array), missing_considerations, unchallenged_assumptions, suggested_questions, overall_risk, risk_level, analysis",
	}, s.handleDetectBlindSpots)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "prove-theorem",
		Description: "Attempt to prove a theorem using symbolic reasoning and logical inference rules (modus ponens, simplification, conjunction). Parameters: name, premises (array), conclusion. Returns: status, is_valid, confidence, proof with steps (step_number, statement, justification, rule, dependencies)",
	}, s.handleProveTheorem)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "check-constraints",
		Description: "Check consistency of symbolic constraints. Detect conflicts and contradictions. Parameters: symbols (array of {name, type, domain}), constraints (array of {type, expression, symbols}). Returns: is_consistent, conflicts (array), explanation",
	}, s.handleCheckConstraints)

	// Register enhanced tools (analogical reasoning, argument analysis, fallacy detection, evidence pipeline, temporal-causal integration)
	handlers.RegisterEnhancedTools(
		mcpServer,
		s.analogicalReasoner,
		s.argumentAnalyzer,
		s.fallacyDetector,
		s.orchestrator,
		s.evidencePipeline,
		s.causalTemporalIntegration,
	)

	// Register episodic memory tools (Phase 2)
	handlers.RegisterEpisodicMemoryTools(mcpServer, s.episodicMemoryHandler)
}

type ThinkRequest struct {
	Content              string          `json:"content"`
	Mode                 string          `json:"mode"`
	Type                 string          `json:"type,omitempty"`
	BranchID             string          `json:"branch_id,omitempty"`
	ParentID             string          `json:"parent_id,omitempty"`
	Confidence           float64         `json:"confidence,omitempty"`
	KeyPoints            []string        `json:"key_points,omitempty"`
	RequireValidation    bool            `json:"require_validation,omitempty"`
	ChallengeAssumptions bool            `json:"challenge_assumptions,omitempty"`
	ForceRebellion       bool            `json:"force_rebellion,omitempty"`
	CrossRefs            []CrossRefInput `json:"cross_refs,omitempty"`
}

type CrossRefInput struct {
	ToBranch string  `json:"to_branch"`
	Type     string  `json:"type"`
	Reason   string  `json:"reason"`
	Strength float64 `json:"strength"`
}

type ThinkResponse struct {
	ThoughtID    string                  `json:"thought_id"`
	Mode         string                  `json:"mode"`
	BranchID     string                  `json:"branch_id,omitempty"`
	Status       string                  `json:"status"`
	Priority     float64                 `json:"priority,omitempty"`
	Confidence   float64                 `json:"confidence"`
	InsightCount int                     `json:"insight_count,omitempty"`
	IsValid      bool                    `json:"is_valid,omitempty"`
	Metadata     *types.ResponseMetadata `json:"metadata"`
}

func (s *UnifiedServer) handleThink(ctx context.Context, req *mcp.CallToolRequest, input ThinkRequest) (*mcp.CallToolResult, *ThinkResponse, error) {
	// Validate input
	if err := ValidateThinkRequest(&input); err != nil {
		return nil, nil, err
	}

	// Track if this is an auto-retry for metadata
	isAutoRetry := false
	autoValidationTriggered := false

	// Create a function to process the thought (can be called for retry)
	processThought := func(challengeAssumptions bool) (*modes.ThoughtResult, error) {
		thoughtInput := modes.ThoughtInput{
			Content:              input.Content,
			Type:                 input.Type,
			BranchID:             input.BranchID,
			ParentID:             input.ParentID,
			Confidence:           input.Confidence,
			KeyPoints:            input.KeyPoints,
			ForceRebellion:       input.ForceRebellion,
			ChallengeAssumptions: challengeAssumptions || input.ChallengeAssumptions,
			CrossRefs:            convertCrossRefs(input.CrossRefs),
		}

		if thoughtInput.Confidence == 0 {
			thoughtInput.Confidence = 0.8
		}

		var result *modes.ThoughtResult
		var err error

		mode := types.ThinkingMode(input.Mode)
		if mode == "" || mode == types.ModeAuto {
			result, err = s.auto.ProcessThought(ctx, thoughtInput)
		} else {
			switch mode {
			case types.ModeLinear:
				result, err = s.linear.ProcessThought(ctx, thoughtInput)
			case types.ModeTree:
				result, err = s.tree.ProcessThought(ctx, thoughtInput)
			case types.ModeDivergent:
				result, err = s.divergent.ProcessThought(ctx, thoughtInput)
			default:
				return nil, fmt.Errorf("unknown mode: %s", mode)
			}
		}

		return result, err
	}

	// Process the thought initially
	result, err := processThought(false)
	if err != nil {
		return nil, nil, err
	}

	// Auto-validation logic for low-confidence thoughts
	// Only trigger if confidence is below threshold and validation is not already required
	// Default threshold is 0.5, configurable via AUTO_VALIDATION_THRESHOLD env var
	autoValidationThreshold := 0.5
	if thresholdStr := os.Getenv("AUTO_VALIDATION_THRESHOLD"); thresholdStr != "" {
		if threshold, err := strconv.ParseFloat(thresholdStr, 64); err == nil && threshold >= 0 && threshold <= 1 {
			autoValidationThreshold = threshold
		}
	}

	if result.Confidence < autoValidationThreshold && !input.RequireValidation {
		autoValidationTriggered = true

		// Get the thought for self-evaluation
		thought, err := s.storage.GetThought(result.ThoughtID)
		if err == nil && thought != nil {
			// Run self-evaluation
			evaluation, evalErr := s.selfEvaluator.EvaluateThought(thought)
			if evalErr == nil && evaluation != nil {
				// Check if there are significant weaknesses
				hasSignificantIssues := evaluation.QualityScore < 0.5 ||
					evaluation.CompletenessScore < 0.5 ||
					evaluation.CoherenceScore < 0.5 ||
					len(evaluation.Weaknesses) > 2

				// If issues found and we haven't already challenged assumptions, retry
				if hasSignificantIssues && !input.ChallengeAssumptions && !isAutoRetry {
					isAutoRetry = true

					// Log the auto-retry for debugging
					if os.Getenv("DEBUG") == "true" {
						fmt.Printf("Auto-validation triggered: confidence=%.2f, quality=%.2f, completeness=%.2f, coherence=%.2f\n",
							result.Confidence, evaluation.QualityScore, evaluation.CompletenessScore, evaluation.CoherenceScore)
						fmt.Println("Retrying with ChallengeAssumptions=true")
					}

					// Retry with ChallengeAssumptions enabled
					retryResult, retryErr := processThought(true)
					if retryErr == nil && retryResult != nil {
						// Use the retry result
						result = retryResult

						// Update the thought's metadata to indicate auto-validation occurred
						if retryThought, err := s.storage.GetThought(result.ThoughtID); err == nil && retryThought != nil {
							if retryThought.Metadata == nil {
								retryThought.Metadata = make(map[string]interface{})
							}
							retryThought.Metadata["auto_validation_triggered"] = true
							retryThought.Metadata["auto_retry_with_challenge"] = true
							retryThought.Metadata["original_confidence"] = thought.Confidence
							retryThought.Metadata["auto_validation_scores"] = map[string]float64{
								"quality":      evaluation.QualityScore,
								"completeness": evaluation.CompletenessScore,
								"coherence":    evaluation.CoherenceScore,
							}
							// Re-store the thought with updated metadata
							_ = s.storage.StoreThought(retryThought)
						}
					}
				} else {
					// Just mark that auto-validation occurred without retry
					if thought.Metadata == nil {
						thought.Metadata = make(map[string]interface{})
					}
					thought.Metadata["auto_validation_triggered"] = true
					thought.Metadata["auto_validation_scores"] = map[string]float64{
						"quality":      evaluation.QualityScore,
						"completeness": evaluation.CompletenessScore,
						"coherence":    evaluation.CoherenceScore,
					}
					// Re-store the thought with updated metadata
					_ = s.storage.StoreThought(thought)
				}
			}
		}
	}

	// Standard validation if requested
	isValid := true
	if input.RequireValidation {
		thought, _ := s.storage.GetThought(result.ThoughtID)
		if thought != nil {
			validationResult, _ := s.validator.ValidateThought(thought)
			if validationResult != nil {
				isValid = validationResult.IsValid
			}
		}
	}

	// Generate metadata for Claude orchestration
	thought, _ := s.storage.GetThought(result.ThoughtID)
	var metadata *types.ResponseMetadata
	if thought != nil {
		metadataGen := handlers.NewMetadataGenerator()
		metadata = metadataGen.GenerateThinkMetadata(
			thought,
			types.ThinkingMode(result.Mode),
			result.Confidence,
			result.InsightCount > 0,
			false, // crossRefs - not directly available from result
		)
	}

	response := &ThinkResponse{
		ThoughtID:    result.ThoughtID,
		Mode:         result.Mode,
		BranchID:     result.BranchID,
		Status:       "success",
		Priority:     result.Priority,
		Confidence:   result.Confidence,
		InsightCount: result.InsightCount,
		IsValid:      isValid,
		Metadata:     metadata,
	}

	// Add auto-validation info to response metadata
	if autoValidationTriggered {
		if os.Getenv("DEBUG") == "true" {
			fmt.Printf("Auto-validation completed for thought %s (confidence: %.2f, retried: %v)\n",
				result.ThoughtID, result.Confidence, isAutoRetry)
		}
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type HistoryRequest struct {
	Mode     string `json:"mode,omitempty"`
	BranchID string `json:"branch_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

type HistoryResponse struct {
	Thoughts []*types.Thought `json:"thoughts"`
}

func (s *UnifiedServer) handleHistory(ctx context.Context, req *mcp.CallToolRequest, input HistoryRequest) (*mcp.CallToolResult, *HistoryResponse, error) {
	// Validate input
	if err := ValidateHistoryRequest(&input); err != nil {
		return nil, nil, err
	}

	// Set default limit if not specified
	limit := input.Limit
	if limit == 0 {
		limit = 100 // Default to 100 results
	}

	var thoughts []*types.Thought

	if input.BranchID != "" {
		branch, err := s.storage.GetBranch(input.BranchID)
		if err != nil {
			return nil, nil, err
		}
		// Apply pagination to branch thoughts
		thoughts = paginateThoughts(branch.Thoughts, limit, input.Offset)
	} else {
		mode := types.ThinkingMode(input.Mode)
		thoughts = s.storage.SearchThoughts("", mode, limit, input.Offset)
	}

	response := &HistoryResponse{Thoughts: thoughts}
	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// paginateThoughts applies limit and offset to a slice of thoughts
func paginateThoughts(thoughts []*types.Thought, limit, offset int) []*types.Thought {
	// Handle offset beyond slice length
	if offset >= len(thoughts) {
		return []*types.Thought{}
	}

	start := offset
	end := offset + limit
	if end > len(thoughts) {
		end = len(thoughts)
	}

	return thoughts[start:end]
}

type EmptyRequest struct{}

type ListBranchesResponse struct {
	Branches       []*types.Branch `json:"branches"`
	ActiveBranchID string          `json:"active_branch_id"`
}

func (s *UnifiedServer) handleListBranches(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *ListBranchesResponse, error) {
	branches := s.storage.ListBranches()

	activeBranch, _ := s.storage.GetActiveBranch()
	activeID := ""
	if activeBranch != nil {
		activeID = activeBranch.ID
	}

	response := &ListBranchesResponse{
		Branches:       branches,
		ActiveBranchID: activeID,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type FocusBranchRequest struct {
	BranchID string `json:"branch_id"`
}

type FocusBranchResponse struct {
	Status         string `json:"status"`
	ActiveBranchID string `json:"active_branch_id"`
}

func (s *UnifiedServer) handleFocusBranch(ctx context.Context, req *mcp.CallToolRequest, input FocusBranchRequest) (*mcp.CallToolResult, *FocusBranchResponse, error) {
	// Validate input
	if err := ValidateFocusBranchRequest(&input); err != nil {
		return nil, nil, err
	}

	// Check if branch is already active
	activeBranch, _ := s.storage.GetActiveBranch()
	if activeBranch != nil && activeBranch.ID == input.BranchID {
		response := &FocusBranchResponse{
			Status:         "already_active",
			ActiveBranchID: input.BranchID,
		}
		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	}

	if err := s.storage.SetActiveBranch(input.BranchID); err != nil {
		return nil, nil, err
	}

	response := &FocusBranchResponse{
		Status:         "success",
		ActiveBranchID: input.BranchID,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type BranchHistoryRequest struct {
	BranchID string `json:"branch_id"`
}

func (s *UnifiedServer) handleBranchHistory(ctx context.Context, req *mcp.CallToolRequest, input BranchHistoryRequest) (*mcp.CallToolResult, *modes.BranchHistory, error) {
	// Validate input
	if err := ValidateBranchHistoryRequest(&input); err != nil {
		return nil, nil, err
	}

	history, err := s.tree.GetBranchHistory(ctx, input.BranchID)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(history),
	}, history, nil
}

type ValidateRequest struct {
	ThoughtID string `json:"thought_id"`
}

type ValidateResponse struct {
	IsValid bool   `json:"is_valid"`
	Reason  string `json:"reason"`
}

func (s *UnifiedServer) handleValidate(ctx context.Context, req *mcp.CallToolRequest, input ValidateRequest) (*mcp.CallToolResult, *ValidateResponse, error) {
	// Validate input
	if err := ValidateValidateRequest(&input); err != nil {
		return nil, nil, err
	}

	thought, err := s.storage.GetThought(input.ThoughtID)
	if err != nil {
		return nil, nil, err
	}

	validationResult, err := s.validator.ValidateThought(thought)
	if err != nil {
		return nil, nil, err
	}

	response := &ValidateResponse{
		IsValid: validationResult.IsValid,
		Reason:  validationResult.Reason,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type ProveRequest struct {
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
}

type ProveResponse struct {
	IsProvable bool     `json:"is_provable"`
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
	Steps      []string `json:"steps"`
}

func (s *UnifiedServer) handleProve(ctx context.Context, req *mcp.CallToolRequest, input ProveRequest) (*mcp.CallToolResult, *ProveResponse, error) {
	// Validate input
	if err := ValidateProveRequest(&input); err != nil {
		return nil, nil, err
	}

	result := s.validator.Prove(input.Premises, input.Conclusion)

	response := &ProveResponse{
		IsProvable: result.IsProvable,
		Premises:   result.Premises,
		Conclusion: result.Conclusion,
		Steps:      result.Steps,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type CheckSyntaxRequest struct {
	Statements []string `json:"statements"`
}

type CheckSyntaxResponse struct {
	Checks []validation.StatementCheck `json:"checks"`
}

func (s *UnifiedServer) handleCheckSyntax(ctx context.Context, req *mcp.CallToolRequest, input CheckSyntaxRequest) (*mcp.CallToolResult, *CheckSyntaxResponse, error) {
	// Validate input
	if err := ValidateCheckSyntaxRequest(&input); err != nil {
		return nil, nil, err
	}

	checks := s.validator.CheckWellFormed(input.Statements)

	response := &CheckSyntaxResponse{
		Checks: checks,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type SearchRequest struct {
	Query  string `json:"query"`
	Mode   string `json:"mode,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

type SearchResponse struct {
	Thoughts []*types.Thought `json:"thoughts"`
}

func (s *UnifiedServer) handleSearch(ctx context.Context, req *mcp.CallToolRequest, input SearchRequest) (*mcp.CallToolResult, *SearchResponse, error) {
	// Validate input
	if err := ValidateSearchRequest(&input); err != nil {
		return nil, nil, err
	}

	// Set default limit if not specified
	limit := input.Limit
	if limit == 0 {
		limit = 100 // Default to 100 results
	}

	mode := types.ThinkingMode(input.Mode)
	thoughts := s.storage.SearchThoughts(input.Query, mode, limit, input.Offset)

	response := &SearchResponse{Thoughts: thoughts}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

func convertCrossRefs(input []CrossRefInput) []modes.CrossRefInput {
	result := make([]modes.CrossRefInput, len(input))
	for i, xref := range input {
		result[i] = modes.CrossRefInput{
			ToBranch: xref.ToBranch,
			Type:     xref.Type,
			Reason:   xref.Reason,
			Strength: xref.Strength,
		}
	}
	return result
}

type MetricsResponse struct {
	TotalThoughts     int            `json:"total_thoughts"`
	TotalBranches     int            `json:"total_branches"`
	TotalInsights     int            `json:"total_insights"`
	TotalValidations  int            `json:"total_validations"`
	ThoughtsByMode    map[string]int `json:"thoughts_by_mode"`
	AverageConfidence float64        `json:"average_confidence"`
}

type RecentBranchesResponse struct {
	ActiveBranchID string          `json:"active_branch_id"`
	RecentBranches []*types.Branch `json:"recent_branches"`
	Count          int             `json:"count"`
}

func (s *UnifiedServer) handleGetMetrics(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *MetricsResponse, error) {
	metrics := s.storage.GetMetrics()

	response := &MetricsResponse{
		TotalThoughts:     metrics.TotalThoughts,
		TotalBranches:     metrics.TotalBranches,
		TotalInsights:     metrics.TotalInsights,
		TotalValidations:  metrics.TotalValidations,
		ThoughtsByMode:    metrics.ThoughtsByMode,
		AverageConfidence: metrics.AverageConfidence,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Integration Patterns Tool
// ============================================================================

type ListIntegrationPatternsResponse struct {
	Patterns []IntegrationPattern `json:"patterns"`
	Count    int                  `json:"count"`
	Status   string               `json:"status"`
}

func (s *UnifiedServer) handleListIntegrationPatterns(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *ListIntegrationPatternsResponse, error) {
	patterns := GetIntegrationPatterns()

	response := &ListIntegrationPatternsResponse{
		Patterns: patterns,
		Count:    len(patterns),
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

func (s *UnifiedServer) handleRecentBranches(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *RecentBranchesResponse, error) {
	branches, err := s.storage.GetRecentBranches()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get recent branches: %w", err)
	}

	// Get active branch for context
	activeBranch, _ := s.storage.GetActiveBranch()
	activeBranchID := ""
	if activeBranch != nil {
		activeBranchID = activeBranch.ID
	}

	response := &RecentBranchesResponse{
		ActiveBranchID: activeBranchID,
		RecentBranches: branches,
		Count:          len(branches),
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Probabilistic Reasoning Tool
// ============================================================================

func (s *UnifiedServer) handleProbabilisticReasoning(ctx context.Context, req *mcp.CallToolRequest, input handlers.ProbabilisticReasoningRequest) (*mcp.CallToolResult, *handlers.ProbabilisticReasoningResponse, error) {
	return s.probabilisticHandler.HandleProbabilisticReasoning(ctx, req, input)
}

// ============================================================================
// Assess Evidence Tool
// ============================================================================

func (s *UnifiedServer) handleAssessEvidence(ctx context.Context, req *mcp.CallToolRequest, input handlers.AssessEvidenceRequest) (*mcp.CallToolResult, *handlers.AssessEvidenceResponse, error) {
	return s.probabilisticHandler.HandleAssessEvidence(ctx, req, input)
}

// ============================================================================
// Detect Contradictions Tool
// ============================================================================

func (s *UnifiedServer) handleDetectContradictions(ctx context.Context, req *mcp.CallToolRequest, input handlers.DetectContradictionsRequest) (*mcp.CallToolResult, *handlers.DetectContradictionsResponse, error) {
	return s.probabilisticHandler.HandleDetectContradictions(ctx, req, input)
}

// ============================================================================
// Make Decision Tool
// ============================================================================

func (s *UnifiedServer) handleMakeDecision(ctx context.Context, req *mcp.CallToolRequest, input handlers.MakeDecisionRequest) (*mcp.CallToolResult, *handlers.MakeDecisionResponse, error) {
	return s.decisionHandler.HandleMakeDecision(ctx, req, input)
}

// ============================================================================
// Decompose Problem Tool
// ============================================================================

func (s *UnifiedServer) handleDecomposeProblem(ctx context.Context, req *mcp.CallToolRequest, input handlers.DecomposeProblemRequest) (*mcp.CallToolResult, *handlers.DecomposeProblemResponse, error) {
	return s.decisionHandler.HandleDecomposeProblem(ctx, req, input)
}

// ============================================================================
// Sensitivity Analysis Tool
// ============================================================================

func (s *UnifiedServer) handleSensitivityAnalysis(ctx context.Context, req *mcp.CallToolRequest, input handlers.SensitivityAnalysisRequest) (*mcp.CallToolResult, *handlers.SensitivityAnalysisResponse, error) {
	return s.decisionHandler.HandleSensitivityAnalysis(ctx, req, input)
}

// ============================================================================
// Self-Evaluate Tool
// ============================================================================

func (s *UnifiedServer) handleSelfEvaluate(ctx context.Context, req *mcp.CallToolRequest, input handlers.SelfEvaluateRequest) (*mcp.CallToolResult, *handlers.SelfEvaluateResponse, error) {
	return s.metacognitionHandler.HandleSelfEvaluate(ctx, req, input)
}

// ============================================================================
// Detect Biases Tool
// ============================================================================

func (s *UnifiedServer) handleDetectBiases(ctx context.Context, req *mcp.CallToolRequest, input handlers.DetectBiasesRequest) (*mcp.CallToolResult, *handlers.DetectBiasesResponse, error) {
	return s.metacognitionHandler.HandleDetectBiases(ctx, req, input)
}

// ========================================
// Phase 2-3: Advanced Reasoning Tool Handlers
// ========================================

// Phase 2: Temporal Analysis - Delegated to handlers.TemporalHandler

func (s *UnifiedServer) handleAnalyzePerspectives(ctx context.Context, req *mcp.CallToolRequest, input handlers.AnalyzePerspectivesRequest) (*mcp.CallToolResult, *handlers.AnalyzePerspectivesResponse, error) {
	return s.temporalHandler.HandleAnalyzePerspectives(ctx, req, input)
}

func (s *UnifiedServer) handleAnalyzeTemporal(ctx context.Context, req *mcp.CallToolRequest, input handlers.AnalyzeTemporalRequest) (*mcp.CallToolResult, *handlers.AnalyzeTemporalResponse, error) {
	return s.temporalHandler.HandleAnalyzeTemporal(ctx, req, input)
}

func (s *UnifiedServer) handleCompareTimeHorizons(ctx context.Context, req *mcp.CallToolRequest, input handlers.CompareTimeHorizonsRequest) (*mcp.CallToolResult, *handlers.CompareTimeHorizonsResponse, error) {
	return s.temporalHandler.HandleCompareTimeHorizons(ctx, req, input)
}

func (s *UnifiedServer) handleIdentifyOptimalTiming(ctx context.Context, req *mcp.CallToolRequest, input handlers.IdentifyOptimalTimingRequest) (*mcp.CallToolResult, *handlers.IdentifyOptimalTimingResponse, error) {
	return s.temporalHandler.HandleIdentifyOptimalTiming(ctx, req, input)
}

// Phase 3: Causal Reasoning - Delegated to handlers.CausalHandler

func (s *UnifiedServer) handleBuildCausalGraph(ctx context.Context, req *mcp.CallToolRequest, input handlers.BuildCausalGraphRequest) (*mcp.CallToolResult, *handlers.BuildCausalGraphResponse, error) {
	return s.causalHandler.HandleBuildCausalGraph(ctx, req, input)
}

func (s *UnifiedServer) handleSimulateIntervention(ctx context.Context, req *mcp.CallToolRequest, input handlers.SimulateInterventionRequest) (*mcp.CallToolResult, *handlers.SimulateInterventionResponse, error) {
	return s.causalHandler.HandleSimulateIntervention(ctx, req, input)
}

func (s *UnifiedServer) handleGenerateCounterfactual(ctx context.Context, req *mcp.CallToolRequest, input handlers.GenerateCounterfactualRequest) (*mcp.CallToolResult, *handlers.GenerateCounterfactualResponse, error) {
	return s.causalHandler.HandleGenerateCounterfactual(ctx, req, input)
}

func (s *UnifiedServer) handleAnalyzeCorrelationVsCausation(ctx context.Context, req *mcp.CallToolRequest, input handlers.AnalyzeCorrelationVsCausationRequest) (*mcp.CallToolResult, *handlers.AnalyzeCorrelationVsCausationResponse, error) {
	return s.causalHandler.HandleAnalyzeCorrelationVsCausation(ctx, req, input)
}

func (s *UnifiedServer) handleGetCausalGraph(ctx context.Context, req *mcp.CallToolRequest, input handlers.GetCausalGraphRequest) (*mcp.CallToolResult, *handlers.GetCausalGraphResponse, error) {
	return s.causalHandler.HandleGetCausalGraph(ctx, req, input)
}

// Phase 3: Cross-Mode Synthesis

type SynthesizeInsightsRequest struct {
	Context string               `json:"context"`
	Inputs  []*integration.Input `json:"inputs"`
}

type SynthesizeInsightsResponse struct {
	Synthesis *types.Synthesis `json:"synthesis"`
	Status    string           `json:"status"`
}

func (s *UnifiedServer) handleSynthesizeInsights(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SynthesizeInsightsRequest,
) (*mcp.CallToolResult, *SynthesizeInsightsResponse, error) {
	synthesis, err := s.synthesizer.SynthesizeInsights(input.Inputs, input.Context)
	if err != nil {
		return nil, nil, err
	}

	response := &SynthesizeInsightsResponse{
		Synthesis: synthesis,
		Status:    "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type DetectEmergentPatternsRequest struct {
	Inputs []*integration.Input `json:"inputs"`
}

type DetectEmergentPatternsResponse struct {
	Patterns []string `json:"patterns"`
	Count    int      `json:"count"`
	Status   string   `json:"status"`
}

func (s *UnifiedServer) handleDetectEmergentPatterns(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input DetectEmergentPatternsRequest,
) (*mcp.CallToolResult, *DetectEmergentPatternsResponse, error) {
	patterns, err := s.synthesizer.DetectEmergentPatterns(input.Inputs)
	if err != nil {
		return nil, nil, err
	}

	response := &DetectEmergentPatternsResponse{
		Patterns: patterns,
		Count:    len(patterns),
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Internal Methods for Tool Executor
// ============================================================================

// ProcessThought processes a thought using the specified mode
func (s *UnifiedServer) ProcessThought(ctx context.Context, modeStr string, input modes.ThoughtInput) (*modes.ThoughtResult, error) {
	mode := types.ThinkingMode(modeStr)
	if mode == "" || mode == types.ModeAuto {
		return s.auto.ProcessThought(ctx, input)
	}

	switch mode {
	case types.ModeLinear:
		return s.linear.ProcessThought(ctx, input)
	case types.ModeTree:
		return s.tree.ProcessThought(ctx, input)
	case types.ModeDivergent:
		return s.divergent.ProcessThought(ctx, input)
	default:
		return nil, fmt.Errorf("unknown mode: %s", mode)
	}
}

// BuildCausalGraph builds a causal graph from problem and context
func (s *UnifiedServer) BuildCausalGraph(ctx context.Context, description string, observations []string) (*types.CausalGraph, error) {
	return s.causalReasoner.BuildCausalGraph(description, observations)
}

// ProbabilisticReasoning performs probabilistic reasoning operations
func (s *UnifiedServer) ProbabilisticReasoning(ctx context.Context, req handlers.ProbabilisticReasoningRequest) (interface{}, error) {
	switch req.Operation {
	case "create":
		return s.probabilisticReasoner.CreateBelief(req.Statement, req.PriorProb)
	case "update":
		return s.probabilisticReasoner.UpdateBelief(req.BeliefID, req.EvidenceID, req.Likelihood, req.EvidenceProb)
	case "get":
		return s.probabilisticReasoner.GetBelief(req.BeliefID)
	case "combine":
		prob, err := s.probabilisticReasoner.CombineBeliefs(req.BeliefIDs, req.CombineOp)
		if err != nil {
			return nil, err
		}
		return map[string]float64{"combined_prob": prob}, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}
}

// MakeDecision creates a structured decision
func (s *UnifiedServer) MakeDecision(ctx context.Context, req handlers.MakeDecisionRequest) (*types.Decision, error) {
	return s.decisionMaker.CreateDecision(req.Question, req.Options, req.Criteria)
}

// DetectContradictions finds contradictions in statements or thoughts
func (s *UnifiedServer) DetectContradictions(ctx context.Context, req handlers.DetectContradictionsRequest) ([]*types.Contradiction, error) {
	var thoughts []*types.Thought

	if len(req.ThoughtIDs) > 0 {
		for _, id := range req.ThoughtIDs {
			thought, err := s.storage.GetThought(id)
			if err != nil {
				return nil, fmt.Errorf("thought not found: %s", id)
			}
			thoughts = append(thoughts, thought)
		}
	} else if req.BranchID != "" {
		branch, err := s.storage.GetBranch(req.BranchID)
		if err != nil {
			return nil, err
		}
		thoughts = branch.Thoughts
	} else if req.Mode != "" {
		mode := types.ThinkingMode(req.Mode)
		thoughts = s.storage.SearchThoughts("", mode, 1000, 0)
	} else {
		// Check all thoughts
		thoughts = s.storage.SearchThoughts("", "", 1000, 0)
	}

	return s.contradictionDetector.DetectContradictions(thoughts)
}

// SynthesizeInsights synthesizes insights from multiple inputs
func (s *UnifiedServer) SynthesizeInsights(ctx context.Context, req SynthesizeInsightsRequest) (*types.Synthesis, error) {
	return s.synthesizer.SynthesizeInsights(req.Inputs, req.Context)
}

// DetectBiases identifies cognitive biases
func (s *UnifiedServer) DetectBiases(ctx context.Context, req handlers.DetectBiasesRequest) ([]*types.CognitiveBias, error) {
	if req.ThoughtID != "" {
		thought, err := s.storage.GetThought(req.ThoughtID)
		if err != nil {
			return nil, err
		}
		return s.biasDetector.DetectBiases(thought)
	} else if req.BranchID != "" {
		branch, err := s.storage.GetBranch(req.BranchID)
		if err != nil {
			return nil, err
		}
		return s.biasDetector.DetectBiasesInBranch(branch)
	}
	return nil, fmt.Errorf("either thought_id or branch_id must be provided")
}

// AssessEvidence evaluates the quality of evidence
func (s *UnifiedServer) AssessEvidence(ctx context.Context, req handlers.AssessEvidenceRequest) (*types.Evidence, error) {
	return s.evidenceAnalyzer.AssessEvidence(req.Content, req.Source, req.ClaimID, req.SupportsClaim)
}

// SelfEvaluate performs metacognitive self-assessment
func (s *UnifiedServer) SelfEvaluate(ctx context.Context, req handlers.SelfEvaluateRequest) (*types.SelfEvaluation, error) {
	if req.ThoughtID != "" {
		thought, err := s.storage.GetThought(req.ThoughtID)
		if err != nil {
			return nil, err
		}
		return s.selfEvaluator.EvaluateThought(thought)
	} else if req.BranchID != "" {
		branch, err := s.storage.GetBranch(req.BranchID)
		if err != nil {
			return nil, err
		}
		return s.selfEvaluator.EvaluateBranch(branch)
	}
	return nil, fmt.Errorf("either thought_id or branch_id must be provided")
}

// DecomposeProblem breaks down a complex problem
func (s *UnifiedServer) DecomposeProblem(ctx context.Context, req handlers.DecomposeProblemRequest) (*types.ProblemDecomposition, error) {
	return s.problemDecomposer.DecomposeProblem(req.Problem)
}

// SensitivityAnalysis tests robustness of conclusions
func (s *UnifiedServer) SensitivityAnalysis(ctx context.Context, req handlers.SensitivityAnalysisRequest) (*types.SensitivityAnalysis, error) {
	return s.sensitivityAnalyzer.AnalyzeSensitivity(req.TargetClaim, req.Assumptions, req.BaseConfidence)
}

// AnalyzePerspectives analyzes multiple stakeholder perspectives
func (s *UnifiedServer) AnalyzePerspectives(ctx context.Context, req handlers.AnalyzePerspectivesRequest) ([]*types.Perspective, error) {
	return s.perspectiveAnalyzer.AnalyzePerspectives(req.Situation, req.StakeholderHints)
}

// AnalyzeTemporal performs temporal reasoning
func (s *UnifiedServer) AnalyzeTemporal(ctx context.Context, req handlers.AnalyzeTemporalRequest) (*types.TemporalAnalysis, error) {
	return s.temporalReasoner.AnalyzeTemporal(req.Situation, req.TimeHorizon)
}

// SimulateIntervention simulates causal interventions
func (s *UnifiedServer) SimulateIntervention(ctx context.Context, req handlers.SimulateInterventionRequest) (*types.CausalIntervention, error) {
	return s.causalReasoner.SimulateIntervention(req.GraphID, req.VariableID, req.InterventionType)
}

// GenerateCounterfactual creates counterfactual scenarios
func (s *UnifiedServer) GenerateCounterfactual(ctx context.Context, req handlers.GenerateCounterfactualRequest) (*types.Counterfactual, error) {
	return s.causalReasoner.GenerateCounterfactual(req.GraphID, req.Scenario, req.Changes)
}

// DetectEmergentPatterns finds emergent patterns across modes
func (s *UnifiedServer) DetectEmergentPatterns(ctx context.Context, req DetectEmergentPatternsRequest) ([]string, error) {
	return s.synthesizer.DetectEmergentPatterns(req.Inputs)
}

// ============================================================================
// Workflow Orchestration Request/Response Types
// ============================================================================

type ExecuteWorkflowRequest struct {
	WorkflowID string                 `json:"workflow_id"`
	Input      map[string]interface{} `json:"input"`
}

type ExecuteWorkflowResponse struct {
	Result *orchestration.WorkflowResult `json:"result"`
	Status string                        `json:"status"`
	Error  string                        `json:"error,omitempty"`
}

type ListWorkflowsResponse struct {
	Workflows []*orchestration.Workflow `json:"workflows"`
	Count     int                       `json:"count"`
}

type RegisterWorkflowRequest struct {
	Workflow *orchestration.Workflow `json:"workflow"`
}

type RegisterWorkflowResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ============================================================================
// Workflow Orchestration Handlers
// ============================================================================

// handleExecuteWorkflow executes a predefined workflow
func (s *UnifiedServer) handleExecuteWorkflow(ctx context.Context, req *mcp.CallToolRequest, input ExecuteWorkflowRequest) (*mcp.CallToolResult, *ExecuteWorkflowResponse, error) {
	// Validate input
	if err := ValidateExecuteWorkflowRequest(&input); err != nil {
		return nil, nil, err
	}

	if s.orchestrator == nil {
		return nil, nil, fmt.Errorf("orchestrator not initialized")
	}

	result, err := s.orchestrator.ExecuteWorkflow(ctx, input.WorkflowID, input.Input)

	response := &ExecuteWorkflowResponse{
		Status: "completed",
	}

	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
	} else {
		response.Result = result
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// handleListWorkflows lists all available workflows
func (s *UnifiedServer) handleListWorkflows(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *ListWorkflowsResponse, error) {
	if s.orchestrator == nil {
		return nil, nil, fmt.Errorf("orchestrator not initialized")
	}

	workflows := s.orchestrator.ListWorkflows()

	response := &ListWorkflowsResponse{
		Workflows: workflows,
		Count:     len(workflows),
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// handleRegisterWorkflow registers a new workflow
func (s *UnifiedServer) handleRegisterWorkflow(ctx context.Context, req *mcp.CallToolRequest, input RegisterWorkflowRequest) (*mcp.CallToolResult, *RegisterWorkflowResponse, error) {
	// Validate input
	if err := ValidateRegisterWorkflowRequest(&input); err != nil {
		return nil, nil, err
	}

	if s.orchestrator == nil {
		return nil, nil, fmt.Errorf("orchestrator not initialized")
	}

	err := s.orchestrator.RegisterWorkflow(input.Workflow)

	response := &RegisterWorkflowResponse{
		Success: err == nil,
		Message: "Workflow registered successfully",
	}

	if err != nil {
		response.Success = false
		response.Message = "Failed to register workflow"
		response.Error = err.Error()
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// handleVerifyThought verifies a thought for hallucinations
func (s *UnifiedServer) handleVerifyThought(ctx context.Context, req *mcp.CallToolRequest, input handlers.VerifyThoughtRequest) (*mcp.CallToolResult, *handlers.VerifyThoughtResponse, error) {
	// Validate input
	if input.ThoughtID == "" {
		return nil, nil, fmt.Errorf("thought_id is required")
	}

	// Verify the thought
	response, err := s.hallucinationHandler.HandleVerifyThought(ctx, &handlers.VerifyThoughtRequest{
		ThoughtID:         input.ThoughtID,
		VerificationLevel: input.VerificationLevel,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("verification failed: %w", err)
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// handleGetHallucinationReport retrieves a cached hallucination report
func (s *UnifiedServer) handleGetHallucinationReport(ctx context.Context, req *mcp.CallToolRequest, input handlers.GetReportRequest) (*mcp.CallToolResult, *handlers.VerifyThoughtResponse, error) {
	// Validate input
	if input.ThoughtID == "" {
		return nil, nil, fmt.Errorf("thought_id is required")
	}

	// Get the report
	response, err := s.hallucinationHandler.HandleGetReport(ctx, &handlers.GetReportRequest{
		ThoughtID: input.ThoughtID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get report: %w", err)
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// handleRecordPrediction records a confidence prediction for calibration
func (s *UnifiedServer) handleRecordPrediction(ctx context.Context, req *mcp.CallToolRequest, input handlers.RecordPredictionRequest) (*mcp.CallToolResult, *handlers.RecordPredictionResponse, error) {
	// Validate input
	if input.ThoughtID == "" {
		return nil, nil, fmt.Errorf("thought_id is required")
	}
	if input.Confidence < 0 || input.Confidence > 1 {
		return nil, nil, fmt.Errorf("confidence must be between 0 and 1")
	}
	if input.Mode == "" {
		return nil, nil, fmt.Errorf("mode is required")
	}

	// Record prediction
	response, err := s.calibrationHandler.HandleRecordPrediction(ctx, &handlers.RecordPredictionRequest{
		ThoughtID:  input.ThoughtID,
		Confidence: input.Confidence,
		Mode:       input.Mode,
		Metadata:   input.Metadata,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to record prediction: %w", err)
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// handleRecordOutcome records an outcome for a prediction
func (s *UnifiedServer) handleRecordOutcome(ctx context.Context, req *mcp.CallToolRequest, input handlers.RecordOutcomeRequest) (*mcp.CallToolResult, *handlers.RecordOutcomeResponse, error) {
	// Validate input
	if input.ThoughtID == "" {
		return nil, nil, fmt.Errorf("thought_id is required")
	}
	if input.ActualConfidence < 0 || input.ActualConfidence > 1 {
		return nil, nil, fmt.Errorf("actual_confidence must be between 0 and 1")
	}
	if input.Source == "" {
		return nil, nil, fmt.Errorf("source is required")
	}

	// Record outcome
	response, err := s.calibrationHandler.HandleRecordOutcome(ctx, &handlers.RecordOutcomeRequest{
		ThoughtID:        input.ThoughtID,
		WasCorrect:       input.WasCorrect,
		ActualConfidence: input.ActualConfidence,
		Source:           input.Source,
		Metadata:         input.Metadata,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to record outcome: %w", err)
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// handleGetCalibrationReport generates a calibration report
func (s *UnifiedServer) handleGetCalibrationReport(ctx context.Context, req *mcp.CallToolRequest, input handlers.GetCalibrationReportRequest) (*mcp.CallToolResult, *handlers.GetCalibrationReportResponse, error) {
	// Get report
	response, err := s.calibrationHandler.HandleGetCalibrationReport(ctx, &handlers.GetCalibrationReportRequest{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get calibration report: %w", err)
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// Phase 2-3: Advanced reasoning handler methods

// handleDualProcessThink handles dual-process thinking
func (s *UnifiedServer) handleDualProcessThink(ctx context.Context, req *mcp.CallToolRequest, input handlers.DualProcessThinkRequest) (*mcp.CallToolResult, *handlers.DualProcessThinkResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.dualProcessHandler.HandleDualProcessThink(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.DualProcessThinkResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleCreateCheckpoint creates a backtracking checkpoint
func (s *UnifiedServer) handleCreateCheckpoint(ctx context.Context, req *mcp.CallToolRequest, input handlers.CreateCheckpointRequest) (*mcp.CallToolResult, *handlers.CreateCheckpointResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.backtrackingHandler.HandleCreateCheckpoint(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.CreateCheckpointResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleRestoreCheckpoint restores from a checkpoint
func (s *UnifiedServer) handleRestoreCheckpoint(ctx context.Context, req *mcp.CallToolRequest, input handlers.RestoreCheckpointRequest) (*mcp.CallToolResult, *handlers.RestoreCheckpointResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.backtrackingHandler.HandleRestoreCheckpoint(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.RestoreCheckpointResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleListCheckpoints lists available checkpoints
func (s *UnifiedServer) handleListCheckpoints(ctx context.Context, req *mcp.CallToolRequest, input handlers.ListCheckpointsRequest) (*mcp.CallToolResult, *handlers.ListCheckpointsResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.backtrackingHandler.HandleListCheckpoints(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.ListCheckpointsResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleGenerateHypotheses generates abductive hypotheses
func (s *UnifiedServer) handleGenerateHypotheses(ctx context.Context, req *mcp.CallToolRequest, input handlers.GenerateHypothesesRequest) (*mcp.CallToolResult, *handlers.GenerateHypothesesResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.abductiveHandler.HandleGenerateHypotheses(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.GenerateHypothesesResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleEvaluateHypotheses evaluates and ranks hypotheses
func (s *UnifiedServer) handleEvaluateHypotheses(ctx context.Context, req *mcp.CallToolRequest, input handlers.EvaluateHypothesesRequest) (*mcp.CallToolResult, *handlers.EvaluateHypothesesResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.abductiveHandler.HandleEvaluateHypotheses(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.EvaluateHypothesesResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleRetrieveCases retrieves similar cases using CBR
func (s *UnifiedServer) handleRetrieveCases(ctx context.Context, req *mcp.CallToolRequest, input handlers.RetrieveCasesRequest) (*mcp.CallToolResult, *handlers.RetrieveCasesResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.caseBasedHandler.HandleRetrieveCases(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.RetrieveCasesResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handlePerformCBRCycle performs full CBR cycle
func (s *UnifiedServer) handlePerformCBRCycle(ctx context.Context, req *mcp.CallToolRequest, input handlers.PerformCBRCycleRequest) (*mcp.CallToolResult, *handlers.PerformCBRCycleResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.caseBasedHandler.HandlePerformCBRCycle(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.PerformCBRCycleResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleDetectBlindSpots detects unknown unknowns and blind spots
func (s *UnifiedServer) handleDetectBlindSpots(ctx context.Context, req *mcp.CallToolRequest, input handlers.DetectBlindSpotsRequest) (*mcp.CallToolResult, *handlers.DetectBlindSpotsResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.unknownUnknownsHandler.HandleDetectBlindSpots(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.DetectBlindSpotsResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleProveTheorem attempts to prove a theorem symbolically
func (s *UnifiedServer) handleProveTheorem(ctx context.Context, req *mcp.CallToolRequest, input handlers.ProveTheoremRequest) (*mcp.CallToolResult, *handlers.ProveTheoremResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.symbolicHandler.HandleProveTheorem(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.ProveTheoremResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}

// handleCheckConstraints checks constraint consistency
func (s *UnifiedServer) handleCheckConstraints(ctx context.Context, req *mcp.CallToolRequest, input handlers.CheckConstraintsRequest) (*mcp.CallToolResult, *handlers.CheckConstraintsResponse, error) {
	params := make(map[string]interface{})
	data, _ := json.Marshal(input)
	json.Unmarshal(data, &params)

	result, err := s.symbolicHandler.HandleCheckConstraints(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	var response handlers.CheckConstraintsResponse
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &response)
		}
	}

	return result, &response, nil
}
