package server

import (
	"fmt"
	"unicode/utf8"
)

// Input validation limits to protect against resource exhaustion and
// malformed requests. These limits are based on typical usage patterns
// and memory constraints.
const (
	// MaxContentLength limits thought content to 100KB to prevent memory exhaustion
	// while allowing substantial thought processing.
	MaxContentLength = 100000

	// MaxKeyPoints limits extracted key points to 50 items to prevent
	// unbounded array growth while supporting detailed analysis.
	MaxKeyPoints = 50

	// MaxKeyPointLength limits individual key point size to 1KB for
	// reasonable summary lengths.
	MaxKeyPointLength = 1000

	// MaxStatements limits syntax check batch size to 100 statements
	// to prevent long processing times.
	MaxStatements = 100

	// MaxStatementLength limits individual statement size to 10KB
	// for complex logical expressions.
	MaxStatementLength = 10000

	// MaxPremises limits proof premise count to 50 for manageable
	// inference complexity.
	MaxPremises = 50

	// MaxPremiseLength limits premise size to 10KB for detailed
	// logical statements.
	MaxPremiseLength = 10000

	// MaxQueryLength limits search queries to 1KB for efficient
	// pattern matching.
	MaxQueryLength = 1000

	// MaxBranchIDLength limits branch identifiers to 100 bytes
	// for reasonable UUID or human-readable IDs.
	MaxBranchIDLength = 100

	// MaxTypeLength limits type field to 100 bytes for standard
	// categorization strings.
	MaxTypeLength = 100

	// MaxCrossRefs limits cross-references per thought to 20 connections
	// to prevent reference graph explosion.
	MaxCrossRefs = 20

	// MaxReasonLength limits reason text to 500 bytes for concise
	// explanations.
	MaxReasonLength = 500
)

// ValidationError represents an input validation error with the specific
// field that failed validation and a descriptive error message.
type ValidationError struct {
	// Field is the name of the request field that failed validation
	Field string
	// Message describes why the validation failed
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// ToolSuggestion provides helpful tool recommendations when users may be using the wrong tool
type ToolSuggestion struct {
	CurrentTool   string
	SuggestedTool string
	Reason        string
	Example       string
}

// SuggestAlternativeTool analyzes the request and suggests a more appropriate tool if applicable
func SuggestAlternativeTool(toolName string, requestContent string) *ToolSuggestion {
	// Normalize content to lowercase for pattern matching
	contentLower := ""
	if requestContent != "" {
		contentLower = requestContent
		// Simple lowercase conversion
		contentLower = toLowerSimple(contentLower)
	}

	// detect-fallacies vs detect-biases confusion
	if toolName == "detect-fallacies" {
		if contains(contentLower, "confirmation bias") || contains(contentLower, "anchoring") ||
		   contains(contentLower, "availability heuristic") || contains(contentLower, "cognitive bias") {
			return &ToolSuggestion{
				CurrentTool:   "detect-fallacies",
				SuggestedTool: "detect-biases",
				Reason:        "You're looking for cognitive biases, not logical fallacies",
				Example:       `Use: {"thought_id": "...", "branch_id": "..."}. detect-biases now detects BOTH biases AND fallacies!`,
			}
		}
	}

	// Workflow confusion - user provides raw problem instead of workflow structure
	if toolName == "execute-workflow" {
		if !contains(contentLower, "workflow_id") && !contains(contentLower, "problem") {
			return &ToolSuggestion{
				CurrentTool:   "execute-workflow",
				SuggestedTool: "execute-workflow (correct format)",
				Reason:        "Workflow execution requires workflow_id and input.problem fields",
				Example:       `Use: {"workflow_id": "comprehensive-analysis", "input": {"problem": "Your problem here"}}. List available workflows with list-workflows tool.`,
			}
		}
	}

	// make-decision - missing required structure
	if toolName == "make-decision" {
		if contains(contentLower, "decision") && (!contains(contentLower, "options") || !contains(contentLower, "criteria")) {
			return &ToolSuggestion{
				CurrentTool:   "make-decision",
				SuggestedTool: "make-decision (correct format)",
				Reason:        "Decision-making requires structured options and criteria",
				Example:       `Required fields: question, options (array with id/name/scores), criteria (array with id/name/weight/maximize). Scores must map to criterion IDs.`,
			}
		}
	}

	return nil
}

// Helper functions for string operations
func toLowerSimple(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// ValidateThinkRequest validates a ThinkRequest
func ValidateThinkRequest(req *ThinkRequest) error {
	// Validate content
	if len(req.Content) == 0 {
		return &ValidationError{"content", "content cannot be empty"}
	}
	if len(req.Content) > MaxContentLength {
		return &ValidationError{"content", fmt.Sprintf("content exceeds maximum length of %d bytes", MaxContentLength)}
	}
	if !utf8.ValidString(req.Content) {
		return &ValidationError{"content", "content must be valid UTF-8"}
	}

	// Validate mode
	validModes := map[string]bool{"linear": true, "tree": true, "divergent": true, "auto": true, "": true}
	if !validModes[req.Mode] {
		return &ValidationError{"mode", fmt.Sprintf("invalid mode: %s (must be 'linear', 'tree', 'divergent', or 'auto')", req.Mode)}
	}

	// Validate type length
	if len(req.Type) > MaxTypeLength {
		return &ValidationError{"type", fmt.Sprintf("type exceeds maximum length of %d", MaxTypeLength)}
	}
	if req.Type != "" && !utf8.ValidString(req.Type) {
		return &ValidationError{"type", "type must be valid UTF-8"}
	}

	// Validate IDs
	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}
	if len(req.ParentID) > MaxBranchIDLength {
		return &ValidationError{"parent_id", "parent_id too long"}
	}

	// Validate confidence range
	if req.Confidence < 0.0 || req.Confidence > 1.0 {
		return &ValidationError{"confidence", "confidence must be between 0.0 and 1.0"}
	}

	// Validate key points
	if len(req.KeyPoints) > MaxKeyPoints {
		return &ValidationError{"key_points", fmt.Sprintf("too many key points (max %d)", MaxKeyPoints)}
	}
	for i, kp := range req.KeyPoints {
		if len(kp) > MaxKeyPointLength {
			return &ValidationError{"key_points", fmt.Sprintf("key_points[%d] exceeds max length of %d", i, MaxKeyPointLength)}
		}
		if !utf8.ValidString(kp) {
			return &ValidationError{"key_points", fmt.Sprintf("key_points[%d] must be valid UTF-8", i)}
		}
	}

	// Validate cross refs
	if len(req.CrossRefs) > MaxCrossRefs {
		return &ValidationError{"cross_refs", fmt.Sprintf("too many cross references (max %d)", MaxCrossRefs)}
	}
	for i, xref := range req.CrossRefs {
		if len(xref.ToBranch) == 0 {
			return &ValidationError{"cross_refs", fmt.Sprintf("cross_refs[%d].to_branch cannot be empty", i)}
		}
		if len(xref.ToBranch) > MaxBranchIDLength {
			return &ValidationError{"cross_refs", fmt.Sprintf("cross_refs[%d].to_branch too long", i)}
		}
		if xref.Strength < 0.0 || xref.Strength > 1.0 {
			return &ValidationError{"cross_refs", fmt.Sprintf("cross_refs[%d].strength must be 0.0-1.0", i)}
		}
		if len(xref.Reason) > MaxReasonLength {
			return &ValidationError{"cross_refs", fmt.Sprintf("cross_refs[%d].reason exceeds max length", i)}
		}
		if !utf8.ValidString(xref.Reason) {
			return &ValidationError{"cross_refs", fmt.Sprintf("cross_refs[%d].reason must be valid UTF-8", i)}
		}
		// Validate type
		validXRefTypes := map[string]bool{"complementary": true, "contradictory": true, "builds_upon": true, "alternative": true}
		if !validXRefTypes[xref.Type] {
			return &ValidationError{"cross_refs", fmt.Sprintf("cross_refs[%d].type invalid (must be 'complementary', 'contradictory', 'builds_upon', or 'alternative')", i)}
		}
	}

	return nil
}

// ValidateHistoryRequest validates a HistoryRequest
func ValidateHistoryRequest(req *HistoryRequest) error {
	// Validate mode if provided
	if req.Mode != "" {
		validModes := map[string]bool{"linear": true, "tree": true, "divergent": true}
		if !validModes[req.Mode] {
			return &ValidationError{"mode", fmt.Sprintf("invalid mode: %s", req.Mode)}
		}
	}

	// Validate branch ID length
	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}

	return nil
}

// ValidateFocusBranchRequest validates a FocusBranchRequest
func ValidateFocusBranchRequest(req *FocusBranchRequest) error {
	if req.BranchID == "" {
		return &ValidationError{"branch_id", "branch_id is required"}
	}
	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}
	return nil
}

// ValidateBranchHistoryRequest validates a BranchHistoryRequest
func ValidateBranchHistoryRequest(req *BranchHistoryRequest) error {
	if req.BranchID == "" {
		return &ValidationError{"branch_id", "branch_id is required"}
	}
	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}
	return nil
}

// ValidateValidateRequest validates a ValidateRequest
func ValidateValidateRequest(req *ValidateRequest) error {
	if req.ThoughtID == "" {
		return &ValidationError{"thought_id", "thought_id is required"}
	}
	if len(req.ThoughtID) > MaxBranchIDLength {
		return &ValidationError{"thought_id", "thought_id too long"}
	}
	return nil
}

// ValidateProveRequest validates a ProveRequest
func ValidateProveRequest(req *ProveRequest) error {
	if len(req.Premises) == 0 {
		return &ValidationError{"premises", "at least one premise is required"}
	}
	if len(req.Premises) > MaxPremises {
		return &ValidationError{"premises", fmt.Sprintf("too many premises (max %d)", MaxPremises)}
	}
	for i, premise := range req.Premises {
		if len(premise) == 0 {
			return &ValidationError{"premises", fmt.Sprintf("premise[%d] cannot be empty", i)}
		}
		if len(premise) > MaxPremiseLength {
			return &ValidationError{"premises", fmt.Sprintf("premise[%d] too long", i)}
		}
		if !utf8.ValidString(premise) {
			return &ValidationError{"premises", fmt.Sprintf("premise[%d] not valid UTF-8", i)}
		}
	}

	if len(req.Conclusion) == 0 {
		return &ValidationError{"conclusion", "conclusion is required"}
	}
	if len(req.Conclusion) > MaxPremiseLength {
		return &ValidationError{"conclusion", "conclusion too long"}
	}
	if !utf8.ValidString(req.Conclusion) {
		return &ValidationError{"conclusion", "conclusion not valid UTF-8"}
	}

	return nil
}

// ValidateCheckSyntaxRequest validates a CheckSyntaxRequest
func ValidateCheckSyntaxRequest(req *CheckSyntaxRequest) error {
	if len(req.Statements) == 0 {
		return &ValidationError{"statements", "at least one statement is required"}
	}
	if len(req.Statements) > MaxStatements {
		return &ValidationError{"statements", fmt.Sprintf("too many statements (max %d)", MaxStatements)}
	}
	for i, stmt := range req.Statements {
		if len(stmt) > MaxStatementLength {
			return &ValidationError{"statements", fmt.Sprintf("statement[%d] too long", i)}
		}
		if !utf8.ValidString(stmt) {
			return &ValidationError{"statements", fmt.Sprintf("statement[%d] not valid UTF-8", i)}
		}
	}
	return nil
}

// ValidateSearchRequest validates a SearchRequest
func ValidateSearchRequest(req *SearchRequest) error {
	if len(req.Query) > MaxQueryLength {
		return &ValidationError{"query", fmt.Sprintf("query exceeds maximum length of %d", MaxQueryLength)}
	}
	if !utf8.ValidString(req.Query) {
		return &ValidationError{"query", "query must be valid UTF-8"}
	}

	// Validate mode if provided
	if req.Mode != "" {
		validModes := map[string]bool{"linear": true, "tree": true, "divergent": true}
		if !validModes[req.Mode] {
			return &ValidationError{"mode", fmt.Sprintf("invalid mode: %s", req.Mode)}
		}
	}

	return nil
}

// ValidateMakeDecisionRequest validates a MakeDecisionRequest
func ValidateMakeDecisionRequest(req *MakeDecisionRequest) error {
	if len(req.Question) == 0 {
		return &ValidationError{"question", "question is required. Example: 'Which cloud provider should we use?'"}
	}
	if len(req.Question) > MaxContentLength {
		return &ValidationError{"question", fmt.Sprintf("question exceeds max length of %d", MaxContentLength)}
	}
	if !utf8.ValidString(req.Question) {
		return &ValidationError{"question", "question must be valid UTF-8"}
	}

	if len(req.Options) == 0 {
		return &ValidationError{"options", "at least one option is required. Each option needs: id, name, description, scores (object mapping to criteria IDs), pros (array), cons (array). Example: {\"id\": \"opt1\", \"name\": \"AWS\", \"scores\": {\"cost\": 0.7, \"performance\": 0.9}}"}
	}
	if len(req.Options) > 50 {
		return &ValidationError{"options", "too many options (max 50)"}
	}

	for i, opt := range req.Options {
		if len(opt.Name) == 0 {
			return &ValidationError{"options", fmt.Sprintf("option[%d].name is required", i)}
		}
		if len(opt.Name) > MaxQueryLength {
			return &ValidationError{"options", fmt.Sprintf("option[%d].name too long", i)}
		}
		if !utf8.ValidString(opt.Name) {
			return &ValidationError{"options", fmt.Sprintf("option[%d].name must be valid UTF-8", i)}
		}
		// Validate that scores exist
		if len(opt.Scores) == 0 {
			return &ValidationError{"options", fmt.Sprintf("option[%d].scores is required and must map to criterion IDs. Example: {\"cost\": 0.8, \"performance\": 0.6}", i)}
		}
	}

	if len(req.Criteria) == 0 {
		return &ValidationError{"criteria", "at least one criterion is required. Each criterion needs: id, name, description, weight (number), maximize (bool). Example: {\"id\": \"cost\", \"name\": \"Cost\", \"weight\": 0.4, \"maximize\": false}"}
	}
	if len(req.Criteria) > 50 {
		return &ValidationError{"criteria", "too many criteria (max 50)"}
	}

	for i, crit := range req.Criteria {
		if len(crit.Name) == 0 {
			return &ValidationError{"criteria", fmt.Sprintf("criterion[%d].name is required", i)}
		}
		if len(crit.Name) > MaxQueryLength {
			return &ValidationError{"criteria", fmt.Sprintf("criterion[%d].name too long", i)}
		}
		if !utf8.ValidString(crit.Name) {
			return &ValidationError{"criteria", fmt.Sprintf("criterion[%d].name must be valid UTF-8", i)}
		}
		if crit.Weight < 0 {
			return &ValidationError{"criteria", fmt.Sprintf("criterion[%d].weight must be non-negative (you provided: %.2f)", i, crit.Weight)}
		}
		if crit.ID == "" {
			return &ValidationError{"criteria", fmt.Sprintf("criterion[%d].id is required - this ID must match keys in option.scores", i)}
		}
	}

	return nil
}

// ValidateDecomposeProblemRequest validates a DecomposeProblemRequest
func ValidateDecomposeProblemRequest(req *DecomposeProblemRequest) error {
	if len(req.Problem) == 0 {
		return &ValidationError{"problem", "problem is required"}
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
		return &ValidationError{"target_claim", "target_claim is required"}
	}
	if len(req.TargetClaim) > MaxContentLength {
		return &ValidationError{"target_claim", fmt.Sprintf("target_claim exceeds max length of %d", MaxContentLength)}
	}
	if !utf8.ValidString(req.TargetClaim) {
		return &ValidationError{"target_claim", "target_claim must be valid UTF-8"}
	}

	if len(req.Assumptions) == 0 {
		return &ValidationError{"assumptions", "at least one assumption is required"}
	}
	if len(req.Assumptions) > 50 {
		return &ValidationError{"assumptions", "too many assumptions (max 50)"}
	}

	for i, assumption := range req.Assumptions {
		if len(assumption) == 0 {
			return &ValidationError{"assumptions", fmt.Sprintf("assumption[%d] cannot be empty", i)}
		}
		if len(assumption) > MaxContentLength {
			return &ValidationError{"assumptions", fmt.Sprintf("assumption[%d] too long", i)}
		}
		if !utf8.ValidString(assumption) {
			return &ValidationError{"assumptions", fmt.Sprintf("assumption[%d] must be valid UTF-8", i)}
		}
	}

	if req.BaseConfidence < 0 || req.BaseConfidence > 1 {
		return &ValidationError{"base_confidence", "base_confidence must be between 0 and 1"}
	}

	return nil
}

// ValidateSelfEvaluateRequest validates a SelfEvaluateRequest
func ValidateSelfEvaluateRequest(req *SelfEvaluateRequest) error {
	if req.ThoughtID == "" && req.BranchID == "" {
		return &ValidationError{"request", "either thought_id or branch_id must be provided"}
	}

	if req.ThoughtID != "" && req.BranchID != "" {
		return &ValidationError{"request", "only one of thought_id or branch_id should be provided"}
	}

	if len(req.ThoughtID) > MaxBranchIDLength {
		return &ValidationError{"thought_id", "thought_id too long"}
	}

	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}

	return nil
}

// ValidateDetectBiasesRequest validates a DetectBiasesRequest
func ValidateDetectBiasesRequest(req *DetectBiasesRequest) error {
	if req.ThoughtID == "" && req.BranchID == "" {
		return &ValidationError{"request", "either thought_id or branch_id must be provided. Example: {\"thought_id\": \"thought_123\"} or {\"branch_id\": \"branch_456\"}"}
	}

	if req.ThoughtID != "" && req.BranchID != "" {
		return &ValidationError{"request", "only one of thought_id or branch_id should be provided, not both"}
	}

	if len(req.ThoughtID) > MaxBranchIDLength {
		return &ValidationError{"thought_id", "thought_id too long"}
	}

	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}

	return nil
}

// ValidateExecuteWorkflowRequest validates an ExecuteWorkflowRequest
func ValidateExecuteWorkflowRequest(req *ExecuteWorkflowRequest) error {
	if req.WorkflowID == "" {
		return &ValidationError{"workflow_id", "workflow_id is required. Use list-workflows to see available workflows. Example: {\"workflow_id\": \"comprehensive-analysis\", \"input\": {\"problem\": \"...\"}}"}
	}

	if len(req.WorkflowID) > MaxBranchIDLength {
		return &ValidationError{"workflow_id", "workflow_id too long"}
	}

	if req.Input == nil {
		return &ValidationError{"input", "input object is required. Most workflows require a 'problem' field. Example: {\"problem\": \"How to optimize performance?\"}"}
	}

	// Check if common workflows have required fields
	switch req.WorkflowID {
	case "comprehensive-analysis", "validation-pipeline":
		if problem, ok := req.Input["problem"].(string); !ok || problem == "" {
			return &ValidationError{"input.problem", fmt.Sprintf("workflow '%s' requires 'problem' field in input. Example: {\"workflow_id\": \"%s\", \"input\": {\"problem\": \"Your problem statement\"}}", req.WorkflowID, req.WorkflowID)}
		}
	}

	return nil
}

// ValidateRegisterWorkflowRequest validates a RegisterWorkflowRequest
func ValidateRegisterWorkflowRequest(req *RegisterWorkflowRequest) error {
	if req.Workflow == nil {
		return &ValidationError{"workflow", "workflow is required"}
	}

	if req.Workflow.ID == "" {
		return &ValidationError{"workflow.id", "workflow ID is required"}
	}

	if req.Workflow.Name == "" {
		return &ValidationError{"workflow.name", "workflow name is required"}
	}

	if len(req.Workflow.Steps) == 0 {
		return &ValidationError{"workflow.steps", "workflow must have at least one step"}
	}

	// Validate workflow type
	validTypes := map[string]bool{"sequential": true, "parallel": true, "conditional": true}
	if !validTypes[string(req.Workflow.Type)] {
		return &ValidationError{"workflow.type", "workflow type must be 'sequential', 'parallel', or 'conditional'"}
	}

	// Validate each step
	for i, step := range req.Workflow.Steps {
		if step.ID == "" {
			return &ValidationError{"workflow.steps", fmt.Sprintf("step[%d].id is required", i)}
		}
		if step.Tool == "" {
			return &ValidationError{"workflow.steps", fmt.Sprintf("step[%d].tool is required", i)}
		}
		if step.Input == nil {
			return &ValidationError{"workflow.steps", fmt.Sprintf("step[%d].input is required (can be empty object)", i)}
		}
	}

	return nil
}
