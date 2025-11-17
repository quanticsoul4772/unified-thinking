// Package server - Tool registration for the MCP server
package server

import (
	// Registry is for internal use, doesn't directly use MCP
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

// BuildToolRegistry creates a registry of tool handlers for reference
// NOTE: This cannot be used with mcp.AddTool directly due to Go's type system
// The actual registration must be done with type-specific handlers
func (s *UnifiedServer) BuildToolRegistry() *ToolRegistry {
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

	// Enhanced handlers are registered separately via RegisterEnhancedTools
	// They have different signatures and are handled differently

	// Register episodic memory handlers (delegated to episodic handler)
	if s.episodicMemoryHandler != nil {
		registry.Register("start-reasoning-session", s.episodicMemoryHandler.HandleStartSession)
		registry.Register("complete-reasoning-session", s.episodicMemoryHandler.HandleCompleteSession)
		registry.Register("get-recommendations", s.episodicMemoryHandler.HandleGetRecommendations)
		registry.Register("search-trajectories", s.episodicMemoryHandler.HandleSearchTrajectories)
		registry.Register("analyze-trajectory", s.episodicMemoryHandler.HandleAnalyzeTrajectory)
	}

	// Return the registry for reference
	// Actual tool registration must be done separately with type-specific handlers
	return registry
}

