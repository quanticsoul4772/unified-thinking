// Package handlers provides MCP tool handlers for enhanced reasoning capabilities.
//
// This module adds handlers for analogical reasoning, argument decomposition,
// fallacy detection, workflow orchestration, and evidence pipeline integration.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

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
		Description: "Find analogies between source and target domains for cross-domain reasoning. Required: source_domain (string), target_problem (string). Optional: constraints (array). Example: {\"source_domain\": \"biology: immune system\", \"target_problem\": \"How to protect computer network?\"}",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input FindAnalogyRequest) (*mcp.CallToolResult, *FindAnalogyResponse, error) {
		if err := ValidateFindAnalogyRequest(&input); err != nil {
			return nil, nil, err
		}
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
		Description: "Apply an existing analogy to a new context. Required: analogy_id (from find-analogy), target_context (string). Example: {\"analogy_id\": \"analogy_123\", \"target_context\": \"New security scenario\"}",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ApplyAnalogyRequest) (*mcp.CallToolResult, *ApplyAnalogyResponse, error) {
		if err := ValidateApplyAnalogyRequest(&input); err != nil {
			return nil, nil, err
		}
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
		Description: "Break down an argument into premises, claims, assumptions, and inference chains. Required: argument (string). Example: {\"argument\": \"We should adopt policy X because studies show it reduces Y by 30%\"}",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DecomposeArgumentRequest) (*mcp.CallToolResult, *DecomposeArgumentResponse, error) {
		if err := ValidateDecomposeArgumentRequest(&input); err != nil {
			return nil, nil, err
		}
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
		Description: "Generate counter-arguments for a given argument using multiple strategies. Required: argument_id (from decompose-argument). Example: {\"argument_id\": \"arg_123\"}",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GenerateCounterArgumentsRequest) (*mcp.CallToolResult, *GenerateCounterArgumentsResponse, error) {
		if err := ValidateGenerateCounterArgumentsRequest(&input); err != nil {
			return nil, nil, err
		}
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
		Description: "Detect formal and informal logical fallacies in reasoning. Required: content (string). Optional: check_formal (bool), check_informal (bool). Detects ad hominem, strawman, false dichotomy, etc. NOTE: For cognitive biases (confirmation bias, anchoring, etc.), use detect-biases instead. Example: {\"content\": \"Everyone says X is true, so it must be\", \"check_formal\": true, \"check_informal\": true}",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DetectFallaciesRequest) (*mcp.CallToolResult, *DetectFallaciesResponse, error) {
		if err := ValidateDetectFallaciesRequest(&input); err != nil {
			return nil, nil, err
		}
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
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
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
		Description: "Process evidence and auto-update beliefs, causal graphs, and decisions. Required: content (string), source (string), supports_claim (bool). Optional: claim_id. Example: {\"content\": \"Study shows X increases Y by 30%\", \"source\": \"Journal of Science 2024\", \"supports_claim\": true}",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ProcessEvidencePipelineRequest) (*mcp.CallToolResult, *ProcessEvidencePipelineResponse, error) {
		if err := ValidateProcessEvidencePipelineRequest(&input); err != nil {
			return nil, nil, err
		}
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
		Description: "Analyze how causal effects evolve across different time horizons. Required: graph_id (from build-causal-graph), variable_id, intervention_type (\"increase\"/\"decrease\"/\"remove\"/\"introduce\"). Example: {\"graph_id\": \"graph_123\", \"variable_id\": \"marketing_spend\", \"intervention_type\": \"increase\"}",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeTemporalCausalEffectsRequest) (*mcp.CallToolResult, *AnalyzeTemporalCausalEffectsResponse, error) {
		if err := ValidateAnalyzeTemporalCausalEffectsRequest(&input); err != nil {
			return nil, nil, err
		}
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
		Description: "Determine optimal timing for decisions based on causal and temporal analysis. Required: situation (string). Optional: causal_graph_id. Example: {\"situation\": \"When to launch product?\", \"causal_graph_id\": \"graph_123\"}",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeDecisionTimingRequest) (*mcp.CallToolResult, *AnalyzeDecisionTimingResponse, error) {
		if err := ValidateAnalyzeDecisionTimingRequest(&input); err != nil {
			return nil, nil, err
		}
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

// Validation constants
const (
	MaxContentLength    = 100000
	MaxConstraints      = 50
	MaxAnalogyIDLength  = 100
	MaxArgumentIDLength = 100
	MaxGraphIDLength    = 100
	MaxVariableIDLength = 100
	MaxSourceLength     = 1000
	MaxContextLength    = 10000
	MaxBranchIDLength   = 100
	MaxQueryLength      = 1000
)

// ValidationError represents a validation error with helpful context
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// Validation functions for enhanced tools

func ValidateFindAnalogyRequest(req *FindAnalogyRequest) error {
	if req.SourceDomain == "" {
		return &ValidationError{"source_domain", "source_domain is required. Example: {\"source_domain\": \"biology: immune system\", \"target_problem\": \"How to protect network?\"}"}
	}
	if len(req.SourceDomain) > MaxContentLength {
		return &ValidationError{"source_domain", fmt.Sprintf("source_domain exceeds max length of %d", MaxContentLength)}
	}
	if req.TargetProblem == "" {
		return &ValidationError{"target_problem", "target_problem is required"}
	}
	if len(req.TargetProblem) > MaxContentLength {
		return &ValidationError{"target_problem", fmt.Sprintf("target_problem exceeds max length of %d", MaxContentLength)}
	}
	if len(req.Constraints) > MaxConstraints {
		return &ValidationError{"constraints", fmt.Sprintf("too many constraints (max %d)", MaxConstraints)}
	}
	return nil
}

func ValidateApplyAnalogyRequest(req *ApplyAnalogyRequest) error {
	if req.AnalogyID == "" {
		return &ValidationError{"analogy_id", "analogy_id is required. First use find-analogy to create an analogy. Example: {\"analogy_id\": \"analogy_123\", \"target_context\": \"...\"}"}
	}
	if len(req.AnalogyID) > MaxAnalogyIDLength {
		return &ValidationError{"analogy_id", "analogy_id too long"}
	}
	if req.TargetContext == "" {
		return &ValidationError{"target_context", "target_context is required"}
	}
	if len(req.TargetContext) > MaxContextLength {
		return &ValidationError{"target_context", fmt.Sprintf("target_context exceeds max length of %d", MaxContextLength)}
	}
	return nil
}

func ValidateDecomposeArgumentRequest(req *DecomposeArgumentRequest) error {
	if req.Argument == "" {
		return &ValidationError{"argument", "argument is required. Example: {\"argument\": \"We should adopt policy X because...\"}"}
	}
	if len(req.Argument) > MaxContentLength {
		return &ValidationError{"argument", fmt.Sprintf("argument exceeds max length of %d", MaxContentLength)}
	}
	return nil
}

func ValidateGenerateCounterArgumentsRequest(req *GenerateCounterArgumentsRequest) error {
	if req.ArgumentID == "" {
		return &ValidationError{"argument_id", "argument_id is required. First use decompose-argument. Example: {\"argument_id\": \"arg_123\"}"}
	}
	if len(req.ArgumentID) > MaxArgumentIDLength {
		return &ValidationError{"argument_id", "argument_id too long"}
	}
	return nil
}

func ValidateDetectFallaciesRequest(req *DetectFallaciesRequest) error {
	if req.Content == "" {
		return &ValidationError{"content", "content is required. Example: {\"content\": \"Everyone says X, so it must be true\", \"check_formal\": true, \"check_informal\": true}"}
	}
	if len(req.Content) > MaxContentLength {
		return &ValidationError{"content", fmt.Sprintf("content exceeds max length of %d", MaxContentLength)}
	}
	return nil
}

func ValidateProcessEvidencePipelineRequest(req *ProcessEvidencePipelineRequest) error {
	if req.Content == "" {
		return &ValidationError{"content", "content is required"}
	}
	if len(req.Content) > MaxContentLength {
		return &ValidationError{"content", fmt.Sprintf("content exceeds max length of %d", MaxContentLength)}
	}
	if req.Source == "" {
		return &ValidationError{"source", "source is required. Example: \"Research paper\", \"Expert testimony\""}
	}
	if len(req.Source) > MaxSourceLength {
		return &ValidationError{"source", fmt.Sprintf("source exceeds max length of %d", MaxSourceLength)}
	}
	return nil
}

func ValidateAnalyzeTemporalCausalEffectsRequest(req *AnalyzeTemporalCausalEffectsRequest) error {
	if req.GraphID == "" {
		return &ValidationError{"graph_id", "graph_id is required. First use build-causal-graph. Example: {\"graph_id\": \"graph_123\", \"variable_id\": \"var_x\", \"intervention_type\": \"increase\"}"}
	}
	if len(req.GraphID) > MaxGraphIDLength {
		return &ValidationError{"graph_id", "graph_id too long"}
	}
	if req.VariableID == "" {
		return &ValidationError{"variable_id", "variable_id is required"}
	}
	if len(req.VariableID) > MaxVariableIDLength {
		return &ValidationError{"variable_id", "variable_id too long"}
	}
	if req.InterventionType == "" {
		return &ValidationError{"intervention_type", "intervention_type is required. Common types: \"increase\", \"decrease\", \"remove\", \"introduce\""}
	}
	validInterventions := map[string]bool{"increase": true, "decrease": true, "remove": true, "introduce": true, "set": true}
	if !validInterventions[req.InterventionType] {
		return &ValidationError{"intervention_type", fmt.Sprintf("invalid intervention_type '%s'. Valid: increase, decrease, remove, introduce, set", req.InterventionType)}
	}
	return nil
}

func ValidateAnalyzeDecisionTimingRequest(req *AnalyzeDecisionTimingRequest) error {
	if req.Situation == "" {
		return &ValidationError{"situation", "situation is required"}
	}
	if len(req.Situation) > MaxContentLength {
		return &ValidationError{"situation", fmt.Sprintf("situation exceeds max length of %d", MaxContentLength)}
	}
	if req.CausalGraphID != "" && len(req.CausalGraphID) > MaxGraphIDLength {
		return &ValidationError{"causal_graph_id", "causal_graph_id too long"}
	}
	return nil
}
