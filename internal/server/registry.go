// Package server - Tool registration for the MCP server
package server

import (
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolRegistry maps tool names to their handler functions
type ToolRegistry struct {
	handlers map[string]interface{}
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		handlers: make(map[string]interface{}),
	}
}

// Register adds a handler for a tool
func (r *ToolRegistry) Register(name string, handler interface{}) {
	r.handlers[name] = handler
}

// Get retrieves a handler by tool name
func (r *ToolRegistry) Get(name string) (interface{}, bool) {
	handler, ok := r.handlers[name]
	return handler, ok
}

// RegisterAllTools registers all tool handlers with the MCP server
func (s *UnifiedServer) RegisterAllTools(mcpServer *mcp.Server) {
	// Create registry for handler lookup
	registry := NewToolRegistry()

	// Register core thinking handlers
	registry.Register("think", s.handleThink)
	registry.Register("history", s.handleHistory)
	registry.Register("list-branches", s.handleListBranches)
	registry.Register("focus-branch", s.handleFocusBranch)
	registry.Register("branch-history", s.handleBranchHistory)
	registry.Register("recent-branches", s.handleRecentBranches)

	// Register validation handlers
	registry.Register("validate", s.handleValidate)
	registry.Register("prove", s.handleProve)
	registry.Register("check-syntax", s.handleCheckSyntax)

	// Register search and metadata handlers
	registry.Register("search", s.handleSearch)
	registry.Register("get-metrics", s.handleGetMetrics)
	registry.Register("list-integration-patterns", s.handleListIntegrationPatterns)

	// Register probabilistic reasoning handlers
	registry.Register("probabilistic-reasoning", s.handleProbabilisticReasoning)
	registry.Register("assess-evidence", s.handleAssessEvidence)
	registry.Register("detect-contradictions", s.handleDetectContradictions)
	registry.Register("sensitivity-analysis", s.handleSensitivityAnalysis)

	// Register decision handlers
	registry.Register("make-decision", s.handleMakeDecision)
	registry.Register("decompose-problem", s.handleDecomposeProblem)
	registry.Register("verify-thought", s.handleVerifyThought)

	// Register metacognition handlers
	registry.Register("self-evaluate", s.handleSelfEvaluate)
	registry.Register("detect-biases", s.handleDetectBiases)
	registry.Register("detect-blind-spots", s.handleDetectBlindSpots)

	// Register hallucination & calibration handlers
	registry.Register("get-hallucination-report", s.handleGetHallucinationReport)
	registry.Register("record-prediction", s.handleRecordPrediction)
	registry.Register("record-outcome", s.handleRecordOutcome)
	registry.Register("get-calibration-report", s.handleGetCalibrationReport)

	// Register temporal & perspective handlers
	registry.Register("analyze-perspectives", s.handleAnalyzePerspectives)
	registry.Register("analyze-temporal", s.handleAnalyzeTemporal)
	registry.Register("compare-time-horizons", s.handleCompareTimeHorizons)
	registry.Register("identify-optimal-timing", s.handleIdentifyOptimalTiming)

	// Register causal reasoning handlers
	registry.Register("build-causal-graph", s.handleBuildCausalGraph)
	registry.Register("simulate-intervention", s.handleSimulateIntervention)
	registry.Register("generate-counterfactual", s.handleGenerateCounterfactual)
	registry.Register("analyze-correlation-vs-causation", s.handleAnalyzeCorrelationVsCausation)
	registry.Register("get-causal-graph", s.handleGetCausalGraph)

	// Register integration & synthesis handlers
	registry.Register("synthesize-insights", s.handleSynthesizeInsights)
	registry.Register("detect-emergent-patterns", s.handleDetectEmergentPatterns)

	// Register workflow orchestration handlers
	registry.Register("execute-workflow", s.handleExecuteWorkflow)
	registry.Register("list-workflows", s.handleListWorkflows)
	registry.Register("register-workflow", s.handleRegisterWorkflow)

	// Register dual-process reasoning handler
	registry.Register("dual-process-think", s.handleDualProcessThink)

	// Register backtracking handlers
	registry.Register("create-checkpoint", s.handleCreateCheckpoint)
	registry.Register("restore-checkpoint", s.handleRestoreCheckpoint)
	registry.Register("list-checkpoints", s.handleListCheckpoints)

	// Register abductive reasoning handlers
	registry.Register("generate-hypotheses", s.handleGenerateHypotheses)
	registry.Register("evaluate-hypotheses", s.handleEvaluateHypotheses)

	// Register case-based reasoning handlers
	registry.Register("retrieve-similar-cases", s.handleRetrieveCases)
	registry.Register("perform-cbr-cycle", s.handlePerformCBRCycle)

	// Register symbolic reasoning handlers
	registry.Register("prove-theorem", s.handleProveTheorem)
	registry.Register("check-constraints", s.handleCheckConstraints)

	// Register enhanced handlers (delegated to enhanced handler)
	if s.enhancedHandler != nil {
		registry.Register("find-analogy", s.enhancedHandler.HandleFindAnalogy)
		registry.Register("apply-analogy", s.enhancedHandler.HandleApplyAnalogy)
		registry.Register("decompose-argument", s.enhancedHandler.HandleDecomposeArgument)
		registry.Register("generate-counter-arguments", s.enhancedHandler.HandleGenerateCounterArguments)
		registry.Register("detect-fallacies", s.enhancedHandler.HandleDetectFallacies)
		registry.Register("process-evidence-pipeline", s.enhancedHandler.HandleProcessEvidencePipeline)
		registry.Register("analyze-temporal-causal-effects", s.enhancedHandler.HandleAnalyzeTemporalCausalEffects)
		registry.Register("analyze-decision-timing", s.enhancedHandler.HandleAnalyzeDecisionTiming)
	}

	// Register episodic memory handlers (delegated to episodic handler)
	if s.episodicHandler != nil {
		registry.Register("start-reasoning-session", s.episodicHandler.HandleStartSession)
		registry.Register("complete-reasoning-session", s.episodicHandler.HandleCompleteSession)
		registry.Register("get-recommendations", s.episodicHandler.HandleGetRecommendations)
		registry.Register("search-trajectories", s.episodicHandler.HandleSearchTrajectories)
		registry.Register("analyze-trajectory", s.episodicHandler.HandleAnalyzeTrajectory)
	}

	// Register all tools with MCP server using definitions from tools.go
	for _, tool := range ToolDefinitions {
		handler, exists := registry.Get(tool.Name)
		if exists {
			toolCopy := tool // Create a copy to avoid closure issues
			mcp.AddTool(mcpServer, &toolCopy, handler)
		} else {
			// Log warning about missing handler
			if debugMode() {
				println("WARNING: No handler found for tool:", tool.Name)
			}
		}
	}
}

// Helper function to check debug mode
func debugMode() bool {
	return os.Getenv("DEBUG") == "true"
}