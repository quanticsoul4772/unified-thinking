// Package server implements the MCP (Model Context Protocol) server for unified thinking.
//
// This package provides the MCP server implementation that exposes 19 tools for
// thought processing, validation, search, and advanced cognitive reasoning. All
// responses are JSON formatted for consumption by Claude AI via stdio transport.
//
// Available tools:
//   - think: Main thinking tool with multiple cognitive modes
//   - history: View thinking history
//   - list-branches: List all thinking branches
//   - focus-branch: Switch active branch
//   - branch-history: Get detailed branch history
//   - validate: Validate thought logical consistency
//   - prove: Attempt logical proof
//   - check-syntax: Validate logical statement syntax
//   - search: Search through all thoughts
//   - get-metrics: Get system performance and usage metrics
//   - recent-branches: Get recently accessed branches for quick context switching
//   - probabilistic-reasoning: Bayesian inference and probabilistic belief updates
//   - assess-evidence: Evidence quality and strength assessment
//   - detect-contradictions: Find contradictions among thoughts
//   - make-decision: Structured multi-criteria decision making
//   - decompose-problem: Break complex problems into subproblems
//   - sensitivity-analysis: Test robustness of conclusions to assumption changes
//   - self-evaluate: Metacognitive self-assessment of reasoning quality
//   - detect-biases: Identify cognitive biases in reasoning
package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/integration"
	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// UnifiedServer coordinates all thinking modes and provides MCP tool handlers.
type UnifiedServer struct {
	storage                *storage.MemoryStorage
	linear                 *modes.LinearMode
	tree                   *modes.TreeMode
	divergent              *modes.DivergentMode
	auto                   *modes.AutoMode
	validator              *validation.LogicValidator
	probabilisticReasoner  *reasoning.ProbabilisticReasoner
	evidenceAnalyzer       *analysis.EvidenceAnalyzer
	contradictionDetector  *analysis.ContradictionDetector
	decisionMaker          *reasoning.DecisionMaker
	problemDecomposer      *reasoning.ProblemDecomposer
	sensitivityAnalyzer    *analysis.SensitivityAnalyzer
	selfEvaluator          *metacognition.SelfEvaluator
	biasDetector           *metacognition.BiasDetector
	// Phase 2-3: Advanced reasoning modules
	perspectiveAnalyzer    *analysis.PerspectiveAnalyzer
	temporalReasoner       *reasoning.TemporalReasoner
	causalReasoner         *reasoning.CausalReasoner
	synthesizer            *integration.Synthesizer
}

func NewUnifiedServer(
	store *storage.MemoryStorage,
	linear *modes.LinearMode,
	tree *modes.TreeMode,
	divergent *modes.DivergentMode,
	auto *modes.AutoMode,
	validator *validation.LogicValidator,
) *UnifiedServer {
	return &UnifiedServer{
		storage:                store,
		linear:                 linear,
		tree:                   tree,
		divergent:              divergent,
		auto:                   auto,
		validator:              validator,
		probabilisticReasoner:  reasoning.NewProbabilisticReasoner(),
		evidenceAnalyzer:       analysis.NewEvidenceAnalyzer(),
		contradictionDetector:  analysis.NewContradictionDetector(),
		decisionMaker:          reasoning.NewDecisionMaker(),
		problemDecomposer:      reasoning.NewProblemDecomposer(),
		sensitivityAnalyzer:    analysis.NewSensitivityAnalyzer(),
		selfEvaluator:          metacognition.NewSelfEvaluator(),
		biasDetector:           metacognition.NewBiasDetector(),
		// Phase 2-3: Initialize advanced reasoning modules
		perspectiveAnalyzer:    analysis.NewPerspectiveAnalyzer(),
		temporalReasoner:       reasoning.NewTemporalReasoner(),
		causalReasoner:         reasoning.NewCausalReasoner(),
		synthesizer:            integration.NewSynthesizer(),
	}
}

func (s *UnifiedServer) RegisterTools(mcpServer *mcp.Server) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "think",
		Description: "Main thinking tool supporting multiple cognitive modes",
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
		Name:        "recent-branches",
		Description: "Get recently accessed branches for quick context switching",
	}, s.handleRecentBranches)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "probabilistic-reasoning",
		Description: "Perform Bayesian inference and update probabilistic beliefs based on evidence",
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
		Name:        "make-decision",
		Description: "Create structured multi-criteria decision framework and recommendations",
	}, s.handleMakeDecision)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "decompose-problem",
		Description: "Break down complex problems into manageable subproblems with dependencies",
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
		Description: "Identify cognitive biases in reasoning and suggest mitigation strategies",
	}, s.handleDetectBiases)

	// Phase 2: Multi-Perspective Analysis Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "analyze-perspectives",
		Description: "Analyze a situation from multiple stakeholder perspectives, identifying concerns, priorities, and conflicts",
	}, s.handleAnalyzePerspectives)

	// Phase 2: Temporal Reasoning Tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "analyze-temporal",
		Description: "Analyze short-term vs long-term implications of a decision, identifying tradeoffs and providing recommendations",
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
		Name:        "build-causal-graph",
		Description: "Construct a causal graph from observations, identifying variables and causal relationships",
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
	ThoughtID    string  `json:"thought_id"`
	Mode         string  `json:"mode"`
	BranchID     string  `json:"branch_id,omitempty"`
	Status       string  `json:"status"`
	Priority     float64 `json:"priority,omitempty"`
	Confidence   float64 `json:"confidence"`
	InsightCount int     `json:"insight_count,omitempty"`
	IsValid      bool    `json:"is_valid,omitempty"`
}

func (s *UnifiedServer) handleThink(ctx context.Context, req *mcp.CallToolRequest, input ThinkRequest) (*mcp.CallToolResult, *ThinkResponse, error) {
	// Validate input
	if err := ValidateThinkRequest(&input); err != nil {
		return nil, nil, err
	}

	thoughtInput := modes.ThoughtInput{
		Content:        input.Content,
		Type:           input.Type,
		BranchID:       input.BranchID,
		ParentID:       input.ParentID,
		Confidence:     input.Confidence,
		KeyPoints:      input.KeyPoints,
		ForceRebellion: input.ForceRebellion,
		CrossRefs:      convertCrossRefs(input.CrossRefs),
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
			return nil, nil, fmt.Errorf("unknown mode: %s", mode)
		}
	}

	if err != nil {
		return nil, nil, err
	}

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

	response := &ThinkResponse{
		ThoughtID:    result.ThoughtID,
		Mode:         result.Mode,
		BranchID:     result.BranchID,
		Status:       "success",
		Priority:     result.Priority,
		Confidence:   result.Confidence,
		InsightCount: result.InsightCount,
		IsValid:      isValid,
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

type ProbabilisticReasoningRequest struct {
	Operation    string   `json:"operation"`              // "create", "update", or "combine"
	Statement    string   `json:"statement,omitempty"`    // For create operation
	PriorProb    float64  `json:"prior_prob,omitempty"`   // For create operation
	BeliefID     string   `json:"belief_id,omitempty"`    // For update/get operations
	EvidenceID   string   `json:"evidence_id,omitempty"`  // For update operation
	Likelihood   float64  `json:"likelihood,omitempty"`   // For update operation
	EvidenceProb float64  `json:"evidence_prob,omitempty"` // For update operation
	BeliefIDs    []string `json:"belief_ids,omitempty"`   // For combine operation
	CombineOp    string   `json:"combine_op,omitempty"`   // "and" or "or" for combine
}

type ProbabilisticReasoningResponse struct {
	Belief           *types.ProbabilisticBelief `json:"belief,omitempty"`
	CombinedProb     float64                    `json:"combined_prob,omitempty"`
	Operation        string                     `json:"operation"`
	Status           string                     `json:"status"`
}

func (s *UnifiedServer) handleProbabilisticReasoning(ctx context.Context, req *mcp.CallToolRequest, input ProbabilisticReasoningRequest) (*mcp.CallToolResult, *ProbabilisticReasoningResponse, error) {
	if err := ValidateProbabilisticReasoningRequest(&input); err != nil {
		return nil, nil, err
	}

	response := &ProbabilisticReasoningResponse{
		Operation: input.Operation,
		Status:    "success",
	}

	switch input.Operation {
	case "create":
		belief, err := s.probabilisticReasoner.CreateBelief(input.Statement, input.PriorProb)
		if err != nil {
			return nil, nil, err
		}
		response.Belief = belief

	case "update":
		belief, err := s.probabilisticReasoner.UpdateBelief(input.BeliefID, input.EvidenceID, input.Likelihood, input.EvidenceProb)
		if err != nil {
			return nil, nil, err
		}
		response.Belief = belief

	case "get":
		belief, err := s.probabilisticReasoner.GetBelief(input.BeliefID)
		if err != nil {
			return nil, nil, err
		}
		response.Belief = belief

	case "combine":
		combinedProb, err := s.probabilisticReasoner.CombineBeliefs(input.BeliefIDs, input.CombineOp)
		if err != nil {
			return nil, nil, err
		}
		response.CombinedProb = combinedProb

	default:
		return nil, nil, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Assess Evidence Tool
// ============================================================================

type AssessEvidenceRequest struct {
	Content       string `json:"content"`
	Source        string `json:"source"`
	ClaimID       string `json:"claim_id,omitempty"`
	SupportsClaim bool   `json:"supports_claim"`
}

type AssessEvidenceResponse struct {
	Evidence *types.Evidence `json:"evidence"`
	Status   string          `json:"status"`
}

func (s *UnifiedServer) handleAssessEvidence(ctx context.Context, req *mcp.CallToolRequest, input AssessEvidenceRequest) (*mcp.CallToolResult, *AssessEvidenceResponse, error) {
	if err := ValidateAssessEvidenceRequest(&input); err != nil {
		return nil, nil, err
	}

	evidence, err := s.evidenceAnalyzer.AssessEvidence(input.Content, input.Source, input.ClaimID, input.SupportsClaim)
	if err != nil {
		return nil, nil, err
	}

	response := &AssessEvidenceResponse{
		Evidence: evidence,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Detect Contradictions Tool
// ============================================================================

type DetectContradictionsRequest struct {
	ThoughtIDs []string `json:"thought_ids,omitempty"` // Specific thought IDs to check
	BranchID   string   `json:"branch_id,omitempty"`   // Or check all thoughts in a branch
	Mode       string   `json:"mode,omitempty"`        // Or check all thoughts in a mode
}

type DetectContradictionsResponse struct {
	Contradictions []*types.Contradiction `json:"contradictions"`
	Count          int                    `json:"count"`
	Status         string                 `json:"status"`
}

func (s *UnifiedServer) handleDetectContradictions(ctx context.Context, req *mcp.CallToolRequest, input DetectContradictionsRequest) (*mcp.CallToolResult, *DetectContradictionsResponse, error) {
	if err := ValidateDetectContradictionsRequest(&input); err != nil {
		return nil, nil, err
	}

	var thoughts []*types.Thought

	// Gather thoughts based on input
	if len(input.ThoughtIDs) > 0 {
		for _, id := range input.ThoughtIDs {
			thought, err := s.storage.GetThought(id)
			if err != nil {
				return nil, nil, fmt.Errorf("thought not found: %s", id)
			}
			thoughts = append(thoughts, thought)
		}
	} else if input.BranchID != "" {
		branch, err := s.storage.GetBranch(input.BranchID)
		if err != nil {
			return nil, nil, err
		}
		thoughts = branch.Thoughts
	} else if input.Mode != "" {
		mode := types.ThinkingMode(input.Mode)
		thoughts = s.storage.SearchThoughts("", mode, 1000, 0)
	} else {
		// Check all thoughts
		thoughts = s.storage.SearchThoughts("", "", 1000, 0)
	}

	contradictions, err := s.contradictionDetector.DetectContradictions(thoughts)
	if err != nil {
		return nil, nil, err
	}

	response := &DetectContradictionsResponse{
		Contradictions: contradictions,
		Count:          len(contradictions),
		Status:         "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Make Decision Tool
// ============================================================================

type MakeDecisionRequest struct {
	Question string                       `json:"question"`
	Options  []*types.DecisionOption      `json:"options"`
	Criteria []*types.DecisionCriterion   `json:"criteria"`
}

type MakeDecisionResponse struct {
	Decision *types.Decision `json:"decision"`
	Status   string          `json:"status"`
}

func (s *UnifiedServer) handleMakeDecision(ctx context.Context, req *mcp.CallToolRequest, input MakeDecisionRequest) (*mcp.CallToolResult, *MakeDecisionResponse, error) {
	if err := ValidateMakeDecisionRequest(&input); err != nil {
		return nil, nil, err
	}

	decision, err := s.decisionMaker.CreateDecision(input.Question, input.Options, input.Criteria)
	if err != nil {
		return nil, nil, err
	}

	response := &MakeDecisionResponse{
		Decision: decision,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Decompose Problem Tool
// ============================================================================

type DecomposeProblemRequest struct {
	Problem string `json:"problem"`
}

type DecomposeProblemResponse struct {
	Decomposition *types.ProblemDecomposition `json:"decomposition"`
	Status        string                      `json:"status"`
}

func (s *UnifiedServer) handleDecomposeProblem(ctx context.Context, req *mcp.CallToolRequest, input DecomposeProblemRequest) (*mcp.CallToolResult, *DecomposeProblemResponse, error) {
	if err := ValidateDecomposeProblemRequest(&input); err != nil {
		return nil, nil, err
	}

	decomposition, err := s.problemDecomposer.DecomposeProblem(input.Problem)
	if err != nil {
		return nil, nil, err
	}

	response := &DecomposeProblemResponse{
		Decomposition: decomposition,
		Status:        "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Sensitivity Analysis Tool
// ============================================================================

type SensitivityAnalysisRequest struct {
	TargetClaim    string   `json:"target_claim"`
	Assumptions    []string `json:"assumptions"`
	BaseConfidence float64  `json:"base_confidence"`
}

type SensitivityAnalysisResponse struct {
	Analysis *types.SensitivityAnalysis `json:"analysis"`
	Status   string                     `json:"status"`
}

func (s *UnifiedServer) handleSensitivityAnalysis(ctx context.Context, req *mcp.CallToolRequest, input SensitivityAnalysisRequest) (*mcp.CallToolResult, *SensitivityAnalysisResponse, error) {
	if err := ValidateSensitivityAnalysisRequest(&input); err != nil {
		return nil, nil, err
	}

	analysis, err := s.sensitivityAnalyzer.AnalyzeSensitivity(input.TargetClaim, input.Assumptions, input.BaseConfidence)
	if err != nil {
		return nil, nil, err
	}

	response := &SensitivityAnalysisResponse{
		Analysis: analysis,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Self-Evaluate Tool
// ============================================================================

type SelfEvaluateRequest struct {
	ThoughtID string `json:"thought_id,omitempty"`
	BranchID  string `json:"branch_id,omitempty"`
}

type SelfEvaluateResponse struct {
	Evaluation *types.SelfEvaluation `json:"evaluation"`
	Status     string                `json:"status"`
}

func (s *UnifiedServer) handleSelfEvaluate(ctx context.Context, req *mcp.CallToolRequest, input SelfEvaluateRequest) (*mcp.CallToolResult, *SelfEvaluateResponse, error) {
	if err := ValidateSelfEvaluateRequest(&input); err != nil {
		return nil, nil, err
	}

	var evaluation *types.SelfEvaluation
	var err error

	if input.ThoughtID != "" {
		thought, getErr := s.storage.GetThought(input.ThoughtID)
		if getErr != nil {
			return nil, nil, getErr
		}
		evaluation, err = s.selfEvaluator.EvaluateThought(thought)
	} else if input.BranchID != "" {
		branch, getErr := s.storage.GetBranch(input.BranchID)
		if getErr != nil {
			return nil, nil, getErr
		}
		evaluation, err = s.selfEvaluator.EvaluateBranch(branch)
	} else {
		return nil, nil, fmt.Errorf("either thought_id or branch_id must be provided")
	}

	if err != nil {
		return nil, nil, err
	}

	response := &SelfEvaluateResponse{
		Evaluation: evaluation,
		Status:     "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ============================================================================
// Detect Biases Tool
// ============================================================================

type DetectBiasesRequest struct {
	ThoughtID string `json:"thought_id,omitempty"`
	BranchID  string `json:"branch_id,omitempty"`
}

type DetectBiasesResponse struct {
	Biases []*types.CognitiveBias `json:"biases"`
	Count  int                    `json:"count"`
	Status string                 `json:"status"`
}

func (s *UnifiedServer) handleDetectBiases(ctx context.Context, req *mcp.CallToolRequest, input DetectBiasesRequest) (*mcp.CallToolResult, *DetectBiasesResponse, error) {
	if err := ValidateDetectBiasesRequest(&input); err != nil {
		return nil, nil, err
	}

	var biases []*types.CognitiveBias
	var err error

	if input.ThoughtID != "" {
		thought, getErr := s.storage.GetThought(input.ThoughtID)
		if getErr != nil {
			return nil, nil, getErr
		}
		biases, err = s.biasDetector.DetectBiases(thought)
	} else if input.BranchID != "" {
		branch, getErr := s.storage.GetBranch(input.BranchID)
		if getErr != nil {
			return nil, nil, getErr
		}
		biases, err = s.biasDetector.DetectBiasesInBranch(branch)
	} else {
		return nil, nil, fmt.Errorf("either thought_id or branch_id must be provided")
	}

	if err != nil {
		return nil, nil, err
	}

	response := &DetectBiasesResponse{
		Biases: biases,
		Count:  len(biases),
		Status: "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// ========================================
// Phase 2-3: Advanced Reasoning Tool Handlers
// ========================================

// Phase 2: Perspective Analysis

type AnalyzePerspectivesRequest struct {
	Situation        string   `json:"situation"`
	StakeholderHints []string `json:"stakeholder_hints,omitempty"`
}

type AnalyzePerspectivesResponse struct {
	Perspectives []*types.Perspective `json:"perspectives"`
	Count        int                  `json:"count"`
	Conflicts    []string             `json:"conflicts,omitempty"`
	Status       string               `json:"status"`
}

func (s *UnifiedServer) handleAnalyzePerspectives(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input AnalyzePerspectivesRequest,
) (*mcp.CallToolResult, *AnalyzePerspectivesResponse, error) {
	perspectives, err := s.perspectiveAnalyzer.AnalyzePerspectives(input.Situation, input.StakeholderHints)
	if err != nil {
		return nil, nil, err
	}

	// Note: conflict detection is done internally, made available through ComparePerspectives if needed
	response := &AnalyzePerspectivesResponse{
		Perspectives: perspectives,
		Count:        len(perspectives),
		Status:       "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// Phase 2: Temporal Reasoning

type AnalyzeTemporalRequest struct {
	Situation   string `json:"situation"`
	TimeHorizon string `json:"time_horizon,omitempty"`
}

type AnalyzeTemporalResponse struct {
	Analysis *types.TemporalAnalysis `json:"analysis"`
	Status   string                  `json:"status"`
}

func (s *UnifiedServer) handleAnalyzeTemporal(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input AnalyzeTemporalRequest,
) (*mcp.CallToolResult, *AnalyzeTemporalResponse, error) {
	analysis, err := s.temporalReasoner.AnalyzeTemporal(input.Situation, input.TimeHorizon)
	if err != nil {
		return nil, nil, err
	}

	response := &AnalyzeTemporalResponse{
		Analysis: analysis,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type CompareTimeHorizonsRequest struct {
	Situation string `json:"situation"`
}

type CompareTimeHorizonsResponse struct {
	Analyses map[string]*types.TemporalAnalysis `json:"analyses"`
	Status   string                             `json:"status"`
}

func (s *UnifiedServer) handleCompareTimeHorizons(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CompareTimeHorizonsRequest,
) (*mcp.CallToolResult, *CompareTimeHorizonsResponse, error) {
	analyses, err := s.temporalReasoner.CompareTimeHorizons(input.Situation)
	if err != nil {
		return nil, nil, err
	}

	response := &CompareTimeHorizonsResponse{
		Analyses: analyses,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type IdentifyOptimalTimingRequest struct {
	Situation   string   `json:"situation"`
	Constraints []string `json:"constraints,omitempty"`
}

type IdentifyOptimalTimingResponse struct {
	Recommendation string `json:"recommendation"`
	Status         string `json:"status"`
}

func (s *UnifiedServer) handleIdentifyOptimalTiming(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input IdentifyOptimalTimingRequest,
) (*mcp.CallToolResult, *IdentifyOptimalTimingResponse, error) {
	recommendation, err := s.temporalReasoner.IdentifyOptimalTiming(input.Situation, input.Constraints)
	if err != nil {
		return nil, nil, err
	}

	response := &IdentifyOptimalTimingResponse{
		Recommendation: recommendation,
		Status:         "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// Phase 3: Causal Reasoning

type BuildCausalGraphRequest struct {
	Description  string   `json:"description"`
	Observations []string `json:"observations"`
}

type BuildCausalGraphResponse struct {
	Graph  *types.CausalGraph `json:"graph"`
	Status string             `json:"status"`
}

func (s *UnifiedServer) handleBuildCausalGraph(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input BuildCausalGraphRequest,
) (*mcp.CallToolResult, *BuildCausalGraphResponse, error) {
	graph, err := s.causalReasoner.BuildCausalGraph(input.Description, input.Observations)
	if err != nil {
		return nil, nil, err
	}

	response := &BuildCausalGraphResponse{
		Graph:  graph,
		Status: "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type SimulateInterventionRequest struct {
	GraphID          string `json:"graph_id"`
	VariableID       string `json:"variable_id"`
	InterventionType string `json:"intervention_type"`
}

type SimulateInterventionResponse struct {
	Intervention *types.CausalIntervention `json:"intervention"`
	Status       string                    `json:"status"`
}

func (s *UnifiedServer) handleSimulateIntervention(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SimulateInterventionRequest,
) (*mcp.CallToolResult, *SimulateInterventionResponse, error) {
	intervention, err := s.causalReasoner.SimulateIntervention(
		input.GraphID,
		input.VariableID,
		input.InterventionType,
	)
	if err != nil {
		return nil, nil, err
	}

	response := &SimulateInterventionResponse{
		Intervention: intervention,
		Status:       "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type GenerateCounterfactualRequest struct {
	GraphID  string            `json:"graph_id"`
	Scenario string            `json:"scenario"`
	Changes  map[string]string `json:"changes"`
}

type GenerateCounterfactualResponse struct {
	Counterfactual *types.Counterfactual `json:"counterfactual"`
	Status         string                `json:"status"`
}

func (s *UnifiedServer) handleGenerateCounterfactual(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GenerateCounterfactualRequest,
) (*mcp.CallToolResult, *GenerateCounterfactualResponse, error) {
	counterfactual, err := s.causalReasoner.GenerateCounterfactual(
		input.GraphID,
		input.Scenario,
		input.Changes,
	)
	if err != nil {
		return nil, nil, err
	}

	response := &GenerateCounterfactualResponse{
		Counterfactual: counterfactual,
		Status:         "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type AnalyzeCorrelationVsCausationRequest struct {
	Observation string `json:"observation"`
}

type AnalyzeCorrelationVsCausationResponse struct {
	Analysis string `json:"analysis"`
	Status   string `json:"status"`
}

func (s *UnifiedServer) handleAnalyzeCorrelationVsCausation(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input AnalyzeCorrelationVsCausationRequest,
) (*mcp.CallToolResult, *AnalyzeCorrelationVsCausationResponse, error) {
	analysis, err := s.causalReasoner.AnalyzeCorrelationVsCausation(input.Observation)
	if err != nil {
		return nil, nil, err
	}

	response := &AnalyzeCorrelationVsCausationResponse{
		Analysis: analysis,
		Status:   "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

type GetCausalGraphRequest struct {
	GraphID string `json:"graph_id"`
}

type GetCausalGraphResponse struct {
	Graph  *types.CausalGraph `json:"graph"`
	Status string             `json:"status"`
}

func (s *UnifiedServer) handleGetCausalGraph(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetCausalGraphRequest,
) (*mcp.CallToolResult, *GetCausalGraphResponse, error) {
	graph, err := s.causalReasoner.GetGraph(input.GraphID)
	if err != nil {
		return nil, nil, err
	}

	response := &GetCausalGraphResponse{
		Graph:  graph,
		Status: "success",
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// Phase 3: Cross-Mode Synthesis

type SynthesizeInsightsRequest struct {
	Context string                   `json:"context"`
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
