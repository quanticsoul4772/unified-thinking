package handlers

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// DecisionHandler handles decision-making and problem decomposition operations
type DecisionHandler struct {
	storage              storage.Storage
	decisionMaker        *reasoning.DecisionMaker
	problemDecomposer    *reasoning.ProblemDecomposer
	llmProblemDecomposer *reasoning.LLMProblemDecomposer
	sensitivityAnalyzer  *analysis.SensitivityAnalyzer
	metadataGen          *MetadataGenerator
}

// NewDecisionHandler creates a new decision handler
func NewDecisionHandler(
	store storage.Storage,
	decisionMaker *reasoning.DecisionMaker,
	problemDecomposer *reasoning.ProblemDecomposer,
	sensitivityAnalyzer *analysis.SensitivityAnalyzer,
) *DecisionHandler {
	return &DecisionHandler{
		storage:             store,
		decisionMaker:       decisionMaker,
		problemDecomposer:   problemDecomposer,
		sensitivityAnalyzer: sensitivityAnalyzer,
		metadataGen:         NewMetadataGenerator(),
	}
}

// SetLLMProblemDecomposer sets the LLM-based problem decomposer
func (h *DecisionHandler) SetLLMProblemDecomposer(llmDecomposer *reasoning.LLMProblemDecomposer) {
	h.llmProblemDecomposer = llmDecomposer
}

// ============================================================================
// Request/Response Types
// ============================================================================

// MakeDecisionRequest represents a decision-making request
type MakeDecisionRequest struct {
	Question string                     `json:"question"`
	Options  []*types.DecisionOption    `json:"options"`
	Criteria []*types.DecisionCriterion `json:"criteria"`
}

// MakeDecisionResponse represents a decision-making response
type MakeDecisionResponse struct {
	Decision *types.Decision         `json:"decision"`
	Status   string                  `json:"status"`
	Metadata *types.ResponseMetadata `json:"metadata,omitempty"`
}

// DecomposeProblemRequest represents a problem decomposition request
type DecomposeProblemRequest struct {
	Problem string  `json:"problem"`
	Domain  *string `json:"domain,omitempty"` // Optional: "debugging", "proof", "architecture", "research", or auto-detect
}

// DecomposeProblemResponse represents a problem decomposition response
type DecomposeProblemResponse struct {
	Decomposition        *types.ProblemDecomposition `json:"decomposition,omitempty"`
	CanDecompose         bool                        `json:"can_decompose"`
	ProblemType          string                      `json:"problem_type,omitempty"`
	DetectedDomain       string                      `json:"detected_domain,omitempty"`     // Phase 2.3: Domain that was used
	DomainWasExplicit    bool                        `json:"domain_was_explicit,omitempty"` // Phase 2.3: Whether domain was specified or detected
	ClassificationReason string                      `json:"reason,omitempty"`
	Approach             string                      `json:"approach,omitempty"`
	SuggestedTools       []string                    `json:"suggested_tools,omitempty"`
	Status               string                      `json:"status"`
	Metadata             *types.ResponseMetadata     `json:"metadata,omitempty"`
}

// SensitivityAnalysisRequest represents a sensitivity analysis request
type SensitivityAnalysisRequest struct {
	TargetClaim    string   `json:"target_claim"`
	Assumptions    []string `json:"assumptions"`
	BaseConfidence float64  `json:"base_confidence"`
}

// SensitivityAnalysisResponse represents a sensitivity analysis response
type SensitivityAnalysisResponse struct {
	Analysis *types.SensitivityAnalysis `json:"analysis"`
	Status   string                     `json:"status"`
}

// ============================================================================
// Handler Methods
// ============================================================================

// HandleMakeDecision processes decision-making requests
func (h *DecisionHandler) HandleMakeDecision(ctx context.Context, req *mcp.CallToolRequest, input MakeDecisionRequest) (*mcp.CallToolResult, *MakeDecisionResponse, error) {
	if err := ValidateMakeDecisionRequest(&input); err != nil {
		return nil, nil, err
	}

	decision, err := h.decisionMaker.CreateDecision(input.Question, input.Options, input.Criteria)
	if err != nil {
		return nil, nil, err
	}

	// Generate metadata for Claude orchestration
	metadata := h.metadataGen.GenerateDecisionMetadata(decision)

	response := &MakeDecisionResponse{
		Decision: decision,
		Status:   "success",
		Metadata: metadata,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleDecomposeProblem processes problem decomposition requests
func (h *DecisionHandler) HandleDecomposeProblem(ctx context.Context, req *mcp.CallToolRequest, input DecomposeProblemRequest) (*mcp.CallToolResult, *DecomposeProblemResponse, error) {
	if err := ValidateDecomposeProblemRequest(&input); err != nil {
		return nil, nil, err
	}

	// SEMANTIC CLASSIFICATION: Check if problem is actually decomposable
	classifier := reasoning.NewProblemClassifier()
	classification := classifier.ClassifyProblem(input.Problem)

	// If problem is NOT decomposable, return classification result
	if classification.Type != reasoning.ProblemTypeDecomposable {
		// Extract suggested tools from approach text
		suggestedTools := extractToolSuggestions(classification.Approach)

		response := &DecomposeProblemResponse{
			CanDecompose:         false,
			ProblemType:          string(classification.Type),
			ClassificationReason: classification.Reasoning,
			Approach:             classification.Approach,
			SuggestedTools:       suggestedTools,
			Status:               "success",
		}

		return &mcp.CallToolResult{
			Content: toJSONContent(response),
		}, response, nil
	}

	// Problem IS decomposable - proceed with domain-aware decomposition

	// Convert explicit domain string to reasoning.Domain if provided
	var explicitDomain *reasoning.Domain
	if input.Domain != nil && *input.Domain != "" {
		domain := reasoning.Domain(*input.Domain)
		// Validate domain
		validDomains := reasoning.GetAllDomains()
		isValid := false
		for _, d := range validDomains {
			if d == domain {
				isValid = true
				break
			}
		}
		if isValid {
			explicitDomain = &domain
		}
	}

	// Use LLM decomposer if available, otherwise fall back to template-based
	var decomposition *types.ProblemDecomposition
	var err error
	if h.llmProblemDecomposer != nil && h.llmProblemDecomposer.HasGenerator() {
		decomposition, err = h.llmProblemDecomposer.DecomposeProblemWithDomain(ctx, input.Problem, explicitDomain)
	} else {
		decomposition, err = h.problemDecomposer.DecomposeProblemWithDomain(input.Problem, explicitDomain)
	}
	if err != nil {
		return nil, nil, err
	}

	// Extract domain from metadata
	detectedDomain := ""
	domainWasExplicit := false
	if decomposition.Metadata != nil {
		if d, ok := decomposition.Metadata["domain"].(string); ok {
			detectedDomain = d
		}
		if e, ok := decomposition.Metadata["domain_detected"].(bool); ok {
			domainWasExplicit = !e
		}
	}

	// Generate metadata for Claude orchestration
	metadata := h.metadataGen.GenerateDecomposeProblemMetadata(decomposition)

	response := &DecomposeProblemResponse{
		CanDecompose:      true,
		ProblemType:       string(reasoning.ProblemTypeDecomposable),
		Decomposition:     decomposition,
		DetectedDomain:    detectedDomain,
		DomainWasExplicit: domainWasExplicit,
		Status:            "success",
		Metadata:          metadata,
	}

	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}

// HandleSensitivityAnalysis processes sensitivity analysis requests
func (h *DecisionHandler) HandleSensitivityAnalysis(ctx context.Context, req *mcp.CallToolRequest, input SensitivityAnalysisRequest) (*mcp.CallToolResult, *SensitivityAnalysisResponse, error) {
	if err := ValidateSensitivityAnalysisRequest(&input); err != nil {
		return nil, nil, err
	}

	analysis, err := h.sensitivityAnalyzer.AnalyzeSensitivity(input.TargetClaim, input.Assumptions, input.BaseConfidence)
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
// Helper Functions
// ============================================================================

// extractToolSuggestions extracts tool names from approach text
func extractToolSuggestions(approach string) []string {
	tools := make([]string, 0)
	approachLower := strings.ToLower(approach)

	// Extract tool names mentioned in approach
	toolMentions := []string{
		"unified-thinking:think",
		"unified-thinking:analyze-perspectives",
		"unified-thinking:analyze-temporal",
		"unified-thinking:make-decision",
		"unified-thinking:find-analogy",
		"unified-thinking:build-causal-graph",
		"unified-thinking:simulate-intervention",
		"unified-thinking:retrieve-similar-cases",
		"unified-thinking:perform-cbr-cycle",
	}

	for _, tool := range toolMentions {
		if strings.Contains(approachLower, tool) ||
			strings.Contains(approachLower, strings.TrimPrefix(tool, "unified-thinking:")) {
			tools = append(tools, tool)
		}
	}

	// If no tools extracted, provide defaults based on context
	if len(tools) == 0 {
		if strings.Contains(approachLower, "philosophical") || strings.Contains(approachLower, "reflective") {
			tools = append(tools, "unified-thinking:think", "unified-thinking:analyze-perspectives")
		} else if strings.Contains(approachLower, "creative") || strings.Contains(approachLower, "divergent") {
			tools = append(tools, "unified-thinking:think (mode='divergent')", "unified-thinking:find-analogy")
		}
	}

	return tools
}

// ============================================================================
// Validation Functions
// ============================================================================

// ValidateMakeDecisionRequest validates a MakeDecisionRequest
func ValidateMakeDecisionRequest(req *MakeDecisionRequest) error {
	if len(req.Question) == 0 {
		return &ValidationError{"question", "question is required. Example: {\"question\": \"Which database?\", \"options\": [...], \"criteria\": [...]}"}
	}
	if len(req.Question) > MaxContentLength {
		return &ValidationError{"question", fmt.Sprintf("question exceeds max length of %d", MaxContentLength)}
	}
	if !utf8.ValidString(req.Question) {
		return &ValidationError{"question", "question must be valid UTF-8"}
	}

	if len(req.Options) == 0 {
		return &ValidationError{"options", "at least one option is required. Example: [{\"id\": \"pg\", \"name\": \"PostgreSQL\", \"description\": \"Relational DB\", \"scores\": {\"cost\": 0.8}, \"pros\": [...], \"cons\": [...]}]"}
	}
	if len(req.Options) > 20 {
		return &ValidationError{"options", "too many options (max 20)"}
	}

	// Validate each option
	for i, opt := range req.Options {
		if opt.ID == "" {
			return &ValidationError{fmt.Sprintf("options[%d].id", i), "option id is required"}
		}
		if opt.Name == "" {
			return &ValidationError{fmt.Sprintf("options[%d].name", i), "option name is required"}
		}
		if len(opt.Scores) == 0 {
			return &ValidationError{fmt.Sprintf("options[%d].scores", i), "option must have at least one score"}
		}
		for criterionID, score := range opt.Scores {
			if score < 0 || score > 1 {
				return &ValidationError{fmt.Sprintf("options[%d].scores[%s]", i, criterionID), fmt.Sprintf("score must be between 0 and 1 (got %.2f)", score)}
			}
		}
	}

	if len(req.Criteria) == 0 {
		return &ValidationError{"criteria", "at least one criterion is required. Example: [{\"id\": \"cost\", \"name\": \"Cost\", \"description\": \"Total cost\", \"weight\": 0.5, \"maximize\": false}]"}
	}
	if len(req.Criteria) > 10 {
		return &ValidationError{"criteria", "too many criteria (max 10)"}
	}

	// Validate each criterion
	totalWeight := 0.0
	for i, crit := range req.Criteria {
		if crit.ID == "" {
			return &ValidationError{fmt.Sprintf("criteria[%d].id", i), "criterion id is required"}
		}
		if crit.Name == "" {
			return &ValidationError{fmt.Sprintf("criteria[%d].name", i), "criterion name is required"}
		}
		if crit.Weight < 0 || crit.Weight > 1 {
			return &ValidationError{fmt.Sprintf("criteria[%d].weight", i), fmt.Sprintf("weight must be between 0 and 1 (got %.2f)", crit.Weight)}
		}
		totalWeight += crit.Weight
	}

	// Check that weights sum to approximately 1.0 (allow small floating point errors)
	if totalWeight < 0.99 || totalWeight > 1.01 {
		return &ValidationError{"criteria", fmt.Sprintf("criterion weights must sum to 1.0 (got %.2f)", totalWeight)}
	}

	return nil
}

// ValidateDecomposeProblemRequest validates a DecomposeProblemRequest
func ValidateDecomposeProblemRequest(req *DecomposeProblemRequest) error {
	if len(req.Problem) == 0 {
		return &ValidationError{"problem", "problem is required. Example: {\"problem\": \"How to improve system performance?\"}"}
	}
	if len(req.Problem) > MaxContentLength {
		return &ValidationError{"problem", fmt.Sprintf("problem exceeds max length of %d", MaxContentLength)}
	}
	if !utf8.ValidString(req.Problem) {
		return &ValidationError{"problem", "problem must be valid UTF-8"}
	}

	return nil
}

// ValidateSensitivityAnalysisRequest validates a SensitivityAnalysisRequest
func ValidateSensitivityAnalysisRequest(req *SensitivityAnalysisRequest) error {
	if len(req.TargetClaim) == 0 {
		return &ValidationError{"target_claim", "target_claim is required. Example: {\"target_claim\": \"X will succeed\", \"assumptions\": [...], \"base_confidence\": 0.8}"}
	}
	if len(req.TargetClaim) > MaxContentLength {
		return &ValidationError{"target_claim", fmt.Sprintf("target_claim exceeds max length of %d", MaxContentLength)}
	}
	if !utf8.ValidString(req.TargetClaim) {
		return &ValidationError{"target_claim", "target_claim must be valid UTF-8"}
	}

	if len(req.Assumptions) == 0 {
		return &ValidationError{"assumptions", "at least one assumption is required. Example: [\"Market remains stable\", \"No major competitors enter\"]"}
	}
	if len(req.Assumptions) > 20 {
		return &ValidationError{"assumptions", "too many assumptions (max 20)"}
	}

	for i, assumption := range req.Assumptions {
		if len(assumption) == 0 {
			return &ValidationError{fmt.Sprintf("assumptions[%d]", i), "assumption cannot be empty"}
		}
		if len(assumption) > MaxQueryLength {
			return &ValidationError{fmt.Sprintf("assumptions[%d]", i), "assumption too long"}
		}
		if !utf8.ValidString(assumption) {
			return &ValidationError{fmt.Sprintf("assumptions[%d]", i), "assumption must be valid UTF-8"}
		}
	}

	if req.BaseConfidence < 0 || req.BaseConfidence > 1 {
		return &ValidationError{"base_confidence", fmt.Sprintf("base_confidence must be between 0 and 1 (got %.2f)", req.BaseConfidence)}
	}

	return nil
}
