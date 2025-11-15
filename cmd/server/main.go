// Package main provides the entry point for the Unified Thinking MCP server.
//
// This server is designed to be spawned as a child process by Claude Desktop
// and communicates via stdio using the Model Context Protocol. It should not
// be run manually by users.
//
// The server consolidates multiple cognitive thinking patterns (linear, tree,
// divergent, and auto modes) into a single Go-based MCP server, providing
// 10 tools for thought processing, validation, search, and metrics.
//
// Environment variables:
//   - DEBUG: Set to "true" to enable debug logging
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/server"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

func main() {
	// Enable debug logging if DEBUG env var is set
	if os.Getenv("DEBUG") == "true" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Starting Unified Thinking Server in debug mode...")
	}

	// Initialize storage (configurable via environment variables)
	store, err := storage.NewStorageFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer func() {
		if err := storage.CloseStorage(store); err != nil {
			log.Printf("Warning: failed to close storage: %v", err)
		}
	}()

	// Initialize modes
	linearMode := modes.NewLinearMode(store)
	treeMode := modes.NewTreeMode(store)
	divergentMode := modes.NewDivergentMode(store)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	log.Println("Initialized thinking modes: linear, tree, divergent, auto")

	// Initialize validator
	validator := validation.NewLogicValidator()
	log.Println("Initialized logic validator")

	// Create unified server (without orchestrator initially)
	srv := server.NewUnifiedServer(store, linearMode, treeMode, divergentMode, autoMode, validator)
	log.Println("Created unified server")

	// Initialize orchestrator with server executor
	executor := server.NewServerToolExecutor(srv)
	orchestrator := orchestration.NewOrchestratorWithExecutor(executor)
	srv.SetOrchestrator(orchestrator)
	log.Println("Initialized workflow orchestrator")

	// Register predefined workflows (optional - can be added later via register-workflow tool)
	registerPredefinedWorkflows(orchestrator)

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "unified-thinking-server",
		Version: "1.0.0",
	}, nil)
	log.Println("Created MCP server")

	// Register tools
	srv.RegisterTools(mcpServer)
	log.Println("Registered tools: think, history, list-branches, focus-branch, branch-history, validate, prove, check-syntax, search, get-metrics, execute-workflow, list-workflows, register-workflow, and 21 additional reasoning tools")

	// Create stdio transport
	transport := &mcp.StdioTransport{}
	log.Println("Created stdio transport")

	// Run server
	ctx := context.Background()
	log.Println("Starting MCP server...")
	if err := mcpServer.Run(ctx, transport); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// registerPredefinedWorkflows registers commonly used workflow patterns
//
// Template Syntax:
// Workflow step inputs support both {{variable}} and $variable syntax for referencing
// workflow input parameters and previous step results. Examples:
//   - {{problem}} or $problem - reference workflow input parameter
//   - {{causal_graph}} or $causal_graph - reference result from step with StoreAs: "causal_graph"
//   - {{causal_graph.id}} or $causal_graph.id - reference nested field in previous result
//
// Resolution Order:
//   1. Reasoning context results (from previous steps)
//   2. Workflow input parameters
//   3. If unresolved, returns original template string for debugging
func registerPredefinedWorkflows(orchestrator *orchestration.Orchestrator) {
	// Guard against nil orchestrator
	if orchestrator == nil {
		log.Println("Warning: Cannot register workflows with nil orchestrator")
		return
	}

	// Register causal analysis workflow
	causalAnalysisWorkflow := &orchestration.Workflow{
		ID:          "causal-analysis",
		Name:        "Causal Analysis Pipeline",
		Description: "Complete causal analysis with fallacy detection",
		Type:        orchestration.WorkflowSequential,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:   "build-graph",
				Tool: "build-causal-graph",
				Input: map[string]interface{}{
					"description":  "{{problem}}",
					"observations": "{{observations}}",
				},
				StoreAs: "causal_graph",
			},
			{
				ID:   "detect-issues",
				Tool: "detect-biases",
				Input: map[string]interface{}{
					"content": "{{problem}}",
				},
				DependsOn: []string{"build-graph"},
				StoreAs:   "detected_issues",
			},
			{
				ID:   "think-about-results",
				Tool: "think",
				Input: map[string]interface{}{
					"content": "Analyze the causal graph and detected issues for: {{problem}}",
					"mode":    "linear",
				},
				DependsOn: []string{"detect-issues"},
				StoreAs:   "analysis",
			},
		},
	}

	// Register critical thinking workflow
	criticalThinkingWorkflow := &orchestration.Workflow{
		ID:          "critical-thinking",
		Name:        "Critical Thinking Analysis",
		Description: "Comprehensive critical thinking pipeline with bias detection and formal validation",
		Type:        orchestration.WorkflowSequential,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:   "detect-biases",
				Tool: "detect-biases",
				Input: map[string]interface{}{
					"content": "{{content}}",
				},
				StoreAs: "detected_biases",
			},
			{
				ID:   "check-syntax",
				Tool: "check-syntax",
				Input: map[string]interface{}{
					"statement": "{{content}}",
				},
				DependsOn: []string{"detect-biases"},
				StoreAs:   "syntax_check",
			},
			{
				ID:   "prove",
				Tool: "prove",
				Input: map[string]interface{}{
					"premises":   "{{premises}}",
					"conclusion": "{{conclusion}}",
				},
				DependsOn: []string{"check-syntax"},
				Condition: &orchestration.StepCondition{
					Type:  "result_match",
					Field: "syntax_check.is_valid",
					Value: true,
				},
				StoreAs: "proof",
			},
		},
	}

	// Register multi-perspective decision workflow
	decisionMakingWorkflow := &orchestration.Workflow{
		ID:          "multi-perspective-decision",
		Name:        "Multi-Perspective Decision Making",
		Description: "Analyze decision from multiple perspectives and make balanced choice",
		Type:        orchestration.WorkflowParallel,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:   "analyze-perspectives",
				Tool: "analyze-perspectives",
				Input: map[string]interface{}{
					"situation": "{{situation}}",
				},
				StoreAs: "perspectives",
			},
			{
				ID:   "sensitivity-analysis",
				Tool: "sensitivity-analysis",
				Input: map[string]interface{}{
					"target_claim": "{{decision}}",
					"assumptions":  "{{assumptions}}",
				},
				StoreAs: "sensitivity",
			},
			{
				ID:   "make-decision",
				Tool: "make-decision",
				Input: map[string]interface{}{
					"situation":    "{{situation}}",
					"criteria":     "{{criteria}}",
					"perspectives": "{{perspectives}}",
					"sensitivity":  "{{sensitivity}}",
				},
				DependsOn: []string{"analyze-perspectives", "sensitivity-analysis"},
			},
		},
	}

	// Register the predefined workflows
	if err := orchestrator.RegisterWorkflow(causalAnalysisWorkflow); err != nil {
		log.Printf("Failed to register causal-analysis workflow: %v", err)
	} else {
		log.Println("Registered workflow: causal-analysis")
	}

	if err := orchestrator.RegisterWorkflow(criticalThinkingWorkflow); err != nil {
		log.Printf("Failed to register critical-thinking workflow: %v", err)
	} else {
		log.Println("Registered workflow: critical-thinking")
	}

	if err := orchestrator.RegisterWorkflow(decisionMakingWorkflow); err != nil {
		log.Printf("Failed to register multi-perspective-decision workflow: %v", err)
	} else {
		log.Println("Registered workflow: multi-perspective-decision")
	}
}
