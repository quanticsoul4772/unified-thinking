// Package server - Refactored tool registration using tool definitions
package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/server/handlers"
)

// RegisterToolsRefactored registers all MCP tools using the centralized definitions from tools.go
// This is a cleaner version that separates tool definitions from registration logic
func (s *UnifiedServer) RegisterToolsRefactored(mcpServer *mcp.Server) {
	// Map tool names to their handlers
	toolHandlers := map[string]interface{}{
		// Core thinking tools
		"think":           s.handleThink,
		"history":         s.handleHistory,
		"list-branches":   s.handleListBranches,
		"focus-branch":    s.handleFocusBranch,
		"branch-history":  s.handleBranchHistory,
		"recent-branches": s.handleRecentBranches,

		// Validation tools
		"validate":      s.handleValidate,
		"prove":         s.handleProve,
		"check-syntax":  s.handleCheckSyntax,
		"verify-thought": s.handleVerifyThought,

		// Search and metadata tools
		"search":                    s.handleSearch,
		"get-metrics":              s.handleGetMetrics,
		"list-integration-patterns": s.handleListIntegrationPatterns,

		// Probabilistic reasoning tools (delegated)
		"probabilistic-reasoning": s.handleProbabilisticReasoning,
		"assess-evidence":        s.handleAssessEvidence,
		"detect-contradictions":   s.handleDetectContradictions,
		"sensitivity-analysis":   s.handleSensitivityAnalysis,

		// Decision tools (delegated)
		"make-decision":     s.handleMakeDecision,
		"decompose-problem": s.handleDecomposeProblem,

		// Metacognition tools (delegated)
		"self-evaluate":      s.handleSelfEvaluate,
		"detect-biases":      s.handleDetectBiases,
		"detect-blind-spots": s.handleDetectBlindSpots,

		// Temporal & perspective tools (delegated)
		"analyze-perspectives":     s.handleAnalyzePerspectives,
		"analyze-temporal":        s.handleAnalyzeTemporal,
		"compare-time-horizons":   s.handleCompareTimeHorizons,
		"identify-optimal-timing": s.handleIdentifyOptimalTiming,

		// Causal reasoning tools (delegated)
		"build-causal-graph":               s.handleBuildCausalGraph,
		"simulate-intervention":            s.handleSimulateIntervention,
		"generate-counterfactual":          s.handleGenerateCounterfactual,
		"analyze-correlation-vs-causation": s.handleAnalyzeCorrelationVsCausation,
		"get-causal-graph":                 s.handleGetCausalGraph,

		// Integration & synthesis tools
		"synthesize-insights":      s.handleSynthesizeInsights,
		"detect-emergent-patterns": s.handleDetectEmergentPatterns,

		// Workflow orchestration tools
		"execute-workflow":   s.handleExecuteWorkflow,
		"list-workflows":     s.handleListWorkflows,
		"register-workflow":  s.handleRegisterWorkflow,

		// Hallucination & calibration tools (delegated)
		"get-hallucination-report": s.handleGetHallucinationReport,
		"record-prediction":        s.handleRecordPrediction,
		"record-outcome":           s.handleRecordOutcome,
		"get-calibration-report":   s.handleGetCalibrationReport,

		// Dual-process reasoning
		"dual-process-think": s.handleDualProcessThink,

		// Backtracking tools (delegated)
		"create-checkpoint":  s.handleCreateCheckpoint,
		"restore-checkpoint": s.handleRestoreCheckpoint,
		"list-checkpoints":   s.handleListCheckpoints,

		// Abductive reasoning tools (delegated)
		"generate-hypotheses": s.handleGenerateHypotheses,
		"evaluate-hypotheses": s.handleEvaluateHypotheses,

		// Case-based reasoning tools (delegated)
		"retrieve-similar-cases": s.handleRetrieveCases,
		"perform-cbr-cycle":      s.handlePerformCBRCycle,

		// Symbolic reasoning tools (delegated)
		"prove-theorem":      s.handleProveTheorem,
		"check-constraints":  s.handleCheckConstraints,
	}

	// Register all tools using centralized definitions
	for _, tool := range ToolDefinitions {
		handler, exists := toolHandlers[tool.Name]
		if exists {
			toolCopy := tool // Create a copy to avoid closure issues
			mcp.AddTool(mcpServer, &toolCopy, handler)
		} else if debugMode() {
			// Log warning for missing handlers (these might be registered elsewhere)
			println("INFO: Handler not found in main registry for tool:", tool.Name)
		}
	}

	// Register enhanced tools separately (they have their own registration function)
	handlers.RegisterEnhancedTools(
		mcpServer,
		s.analogicalReasoner,
		s.argumentAnalyzer,
		s.fallacyDetector,
		s.orchestrator,
		s.evidencePipeline,
		s.causalTemporalIntegration,
	)

	// Register episodic memory tools if handler is available
	if s.episodicMemoryHandler != nil {
		s.episodicMemoryHandler.RegisterTools(mcpServer)
	}
}

// RegisterTools is the original registration function - can be replaced with RegisterToolsRefactored
// once the refactoring is tested and validated
func (s *UnifiedServer) RegisterToolsOld(mcpServer *mcp.Server) {
	// ... original implementation ...
	// This would be deleted once the refactoring is complete
}