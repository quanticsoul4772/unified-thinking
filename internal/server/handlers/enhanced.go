// Package handlers provides MCP tool handlers for enhanced reasoning capabilities.
//
// This module adds handlers for analogical reasoning, argument decomposition,
// fallacy detection, workflow orchestration, and evidence pipeline integration.
package handlers

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/integration"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/validation"
)

// RegisterEnhancedTools registers all enhanced reasoning tools
func RegisterEnhancedTools(
	mcpServer *mcp.Server,
	analogicalReasoner *reasoning.AnalogicalReasoner,
	argumentAnalyzer *analysis.ArgumentAnalyzer,
	fallacyDetector *validation.FallacyDetector,
	orchestrator *orchestration.Orchestrator,
	evidencePipeline *integration.EvidencePipeline,
	causalTemporalIntegration *integration.CausalTemporalIntegration,
) {
	// Analogical Reasoning Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "find-analogy",
		Description: "Find analogies between source and target domains for cross-domain reasoning",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input FindAnalogyRequest) (*mcp.CallToolResult, *FindAnalogyResponse, error) {
		analogy, err := analogicalReasoner.FindAnalogy(input.SourceDomain, input.TargetProblem, input.Constraints)
		if err != nil {
			return nil, nil, err
		}

		response := &FindAnalogyResponse{
			Analogy: analogy,
			Status:  "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "apply-analogy",
		Description: "Apply an existing analogy to a new context",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ApplyAnalogyRequest) (*mcp.CallToolResult, *ApplyAnalogyResponse, error) {
		result, err := analogicalReasoner.ApplyAnalogy(input.AnalogyID, input.TargetContext)
		if err != nil {
			return nil, nil, err
		}

		response := &ApplyAnalogyResponse{
			Result: result,
			Status: "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	// Argument Analysis Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "decompose-argument",
		Description: "Break down an argument into premises, claims, assumptions, and inference chains",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DecomposeArgumentRequest) (*mcp.CallToolResult, *DecomposeArgumentResponse, error) {
		decomposition, err := argumentAnalyzer.DecomposeArgument(input.Argument)
		if err != nil {
			return nil, nil, err
		}

		response := &DecomposeArgumentResponse{
			Decomposition: decomposition,
			Status:        "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "generate-counter-arguments",
		Description: "Generate counter-arguments for a given argument using multiple strategies",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GenerateCounterArgumentsRequest) (*mcp.CallToolResult, *GenerateCounterArgumentsResponse, error) {
		counterArgs, err := argumentAnalyzer.GenerateCounterArguments(input.ArgumentID)
		if err != nil {
			return nil, nil, err
		}

		response := &GenerateCounterArgumentsResponse{
			CounterArguments: counterArgs,
			Count:            len(counterArgs),
			Status:           "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	// Fallacy Detection Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "detect-fallacies",
		Description: "Detect formal and informal logical fallacies in reasoning",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DetectFallaciesRequest) (*mcp.CallToolResult, *DetectFallaciesResponse, error) {
		fallacies := fallacyDetector.DetectFallacies(input.Content, input.CheckFormal, input.CheckInformal)

		response := &DetectFallaciesResponse{
			Fallacies: fallacies,
			Count:     len(fallacies),
			Status:    "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	// Workflow Orchestration Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "execute-workflow",
		Description: "Execute a predefined reasoning workflow with automatic tool chaining",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ExecuteWorkflowRequest) (*mcp.CallToolResult, *ExecuteWorkflowResponse, error) {
		result, err := orchestrator.ExecuteWorkflow(ctx, input.WorkflowID, input.Input)
		if err != nil {
			return nil, nil, err
		}

		response := &ExecuteWorkflowResponse{
			Result: result,
			Status: "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list-workflows",
		Description: "List all available reasoning workflows",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EmptyRequest) (*mcp.CallToolResult, *ListWorkflowsResponse, error) {
		workflows := orchestrator.ListWorkflows()

		response := &ListWorkflowsResponse{
			Workflows: workflows,
			Count:     len(workflows),
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	// Evidence Pipeline Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "process-evidence-pipeline",
		Description: "Process evidence and auto-update beliefs, causal graphs, and decisions",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ProcessEvidencePipelineRequest) (*mcp.CallToolResult, *ProcessEvidencePipelineResponse, error) {
		result, err := evidencePipeline.ProcessEvidence(
			input.Content,
			input.Source,
			input.ClaimID,
			input.SupportsClaim,
		)
		if err != nil {
			return nil, nil, err
		}

		response := &ProcessEvidencePipelineResponse{
			Result: result,
			Status: "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	// Causal-Temporal Integration Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "analyze-temporal-causal-effects",
		Description: "Analyze how causal effects evolve across different time horizons",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeTemporalCausalEffectsRequest) (*mcp.CallToolResult, *AnalyzeTemporalCausalEffectsResponse, error) {
		result, err := causalTemporalIntegration.AnalyzeTemporalCausalEffects(
			input.GraphID,
			input.VariableID,
			input.InterventionType,
		)
		if err != nil {
			return nil, nil, err
		}

		response := &AnalyzeTemporalCausalEffectsResponse{
			Result: result,
			Status: "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "analyze-decision-timing",
		Description: "Determine optimal timing for decisions based on causal and temporal analysis",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeDecisionTimingRequest) (*mcp.CallToolResult, *AnalyzeDecisionTimingResponse, error) {
		result, err := causalTemporalIntegration.AnalyzeDecisionTiming(
			input.Situation,
			input.CausalGraphID,
		)
		if err != nil {
			return nil, nil, err
		}

		response := &AnalyzeDecisionTimingResponse{
			Result: result,
			Status: "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	})
}

// Request/Response types

type FindAnalogyRequest struct {
	SourceDomain  string   `json:"source_domain"`
	TargetProblem string   `json:"target_problem"`
	Constraints   []string `json:"constraints,omitempty"`
}

type FindAnalogyResponse struct {
	Analogy interface{} `json:"analogy"`
	Status  string      `json:"status"`
}

type ApplyAnalogyRequest struct {
	AnalogyID     string `json:"analogy_id"`
	TargetContext string `json:"target_context"`
}

type ApplyAnalogyResponse struct {
	Result interface{} `json:"result"`
	Status string      `json:"status"`
}

type DecomposeArgumentRequest struct {
	Argument string `json:"argument"`
}

type DecomposeArgumentResponse struct {
	Decomposition interface{} `json:"decomposition"`
	Status        string      `json:"status"`
}

type GenerateCounterArgumentsRequest struct {
	ArgumentID string `json:"argument_id"`
}

type GenerateCounterArgumentsResponse struct {
	CounterArguments interface{} `json:"counter_arguments"`
	Count            int         `json:"count"`
	Status           string      `json:"status"`
}

type DetectFallaciesRequest struct {
	Content       string `json:"content"`
	CheckFormal   bool   `json:"check_formal"`
	CheckInformal bool   `json:"check_informal"`
}

type DetectFallaciesResponse struct {
	Fallacies interface{} `json:"fallacies"`
	Count     int         `json:"count"`
	Status    string      `json:"status"`
}

type ExecuteWorkflowRequest struct {
	WorkflowID string                 `json:"workflow_id"`
	Input      map[string]interface{} `json:"input"`
}

type ExecuteWorkflowResponse struct {
	Result interface{} `json:"result"`
	Status string      `json:"status"`
}

// EmptyRequest type removed - using the one from branches.go

type ListWorkflowsResponse struct {
	Workflows interface{} `json:"workflows"`
	Count     int         `json:"count"`
}

type ProcessEvidencePipelineRequest struct {
	Content       string `json:"content"`
	Source        string `json:"source"`
	ClaimID       string `json:"claim_id,omitempty"`
	SupportsClaim bool   `json:"supports_claim"`
}

type ProcessEvidencePipelineResponse struct {
	Result interface{} `json:"result"`
	Status string      `json:"status"`
}

type AnalyzeTemporalCausalEffectsRequest struct {
	GraphID          string `json:"graph_id"`
	VariableID       string `json:"variable_id"`
	InterventionType string `json:"intervention_type"`
}

type AnalyzeTemporalCausalEffectsResponse struct {
	Result interface{} `json:"result"`
	Status string      `json:"status"`
}

type AnalyzeDecisionTimingRequest struct {
	Situation     string `json:"situation"`
	CausalGraphID string `json:"causal_graph_id"`
}

type AnalyzeDecisionTimingResponse struct {
	Result interface{} `json:"result"`
	Status string      `json:"status"`
}

// Helper function to convert responses to JSON content
func toJSONContent(data interface{}) []mcp.Content {
	jsonData, err := json.Marshal(data)
	if err != nil {
		errData := map[string]string{"error": err.Error()}
		jsonData, _ = json.Marshal(errData)
	}

	return []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}
}
