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

// ValidateProbabilisticReasoningRequest validates a ProbabilisticReasoningRequest
func ValidateProbabilisticReasoningRequest(req *ProbabilisticReasoningRequest) error {
	// Validate operation
	validOps := map[string]bool{"create": true, "update": true, "get": true, "combine": true}
	if !validOps[req.Operation] {
		return &ValidationError{"operation", "operation must be 'create', 'update', 'get', or 'combine'"}
	}

	// Validate based on operation
	switch req.Operation {
	case "create":
		if len(req.Statement) == 0 {
			return &ValidationError{"statement", "statement is required for create operation"}
		}
		if len(req.Statement) > MaxContentLength {
			return &ValidationError{"statement", fmt.Sprintf("statement exceeds max length of %d", MaxContentLength)}
		}
		if !utf8.ValidString(req.Statement) {
			return &ValidationError{"statement", "statement must be valid UTF-8"}
		}
		if req.PriorProb < 0 || req.PriorProb > 1 {
			return &ValidationError{"prior_prob", "prior_prob must be between 0 and 1"}
		}

	case "update":
		if len(req.BeliefID) == 0 {
			return &ValidationError{"belief_id", "belief_id is required for update operation"}
		}
		if len(req.EvidenceID) == 0 {
			return &ValidationError{"evidence_id", "evidence_id is required for update operation"}
		}
		if req.Likelihood < 0 || req.Likelihood > 1 {
			return &ValidationError{"likelihood", "likelihood must be between 0 and 1"}
		}
		if req.EvidenceProb <= 0 || req.EvidenceProb > 1 {
			return &ValidationError{"evidence_prob", "evidence_prob must be between 0 and 1 (exclusive 0)"}
		}

	case "get":
		if len(req.BeliefID) == 0 {
			return &ValidationError{"belief_id", "belief_id is required for get operation"}
		}

	case "combine":
		if len(req.BeliefIDs) == 0 {
			return &ValidationError{"belief_ids", "at least one belief_id is required for combine operation"}
		}
		if len(req.BeliefIDs) > 50 {
			return &ValidationError{"belief_ids", "too many belief_ids (max 50)"}
		}
		validCombineOps := map[string]bool{"and": true, "or": true}
		if !validCombineOps[req.CombineOp] {
			return &ValidationError{"combine_op", "combine_op must be 'and' or 'or'"}
		}
	}

	return nil
}

// ValidateAssessEvidenceRequest validates an AssessEvidenceRequest
func ValidateAssessEvidenceRequest(req *AssessEvidenceRequest) error {
	if len(req.Content) == 0 {
		return &ValidationError{"content", "content is required"}
	}
	if len(req.Content) > MaxContentLength {
		return &ValidationError{"content", fmt.Sprintf("content exceeds max length of %d", MaxContentLength)}
	}
	if !utf8.ValidString(req.Content) {
		return &ValidationError{"content", "content must be valid UTF-8"}
	}

	if len(req.Source) == 0 {
		return &ValidationError{"source", "source is required"}
	}
	if len(req.Source) > MaxQueryLength {
		return &ValidationError{"source", fmt.Sprintf("source exceeds max length of %d", MaxQueryLength)}
	}
	if !utf8.ValidString(req.Source) {
		return &ValidationError{"source", "source must be valid UTF-8"}
	}

	if len(req.ClaimID) > MaxBranchIDLength {
		return &ValidationError{"claim_id", "claim_id too long"}
	}

	return nil
}

// ValidateDetectContradictionsRequest validates a DetectContradictionsRequest
func ValidateDetectContradictionsRequest(req *DetectContradictionsRequest) error {
	if len(req.ThoughtIDs) > 100 {
		return &ValidationError{"thought_ids", "too many thought_ids (max 100)"}
	}

	if len(req.BranchID) > MaxBranchIDLength {
		return &ValidationError{"branch_id", "branch_id too long"}
	}

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
		return &ValidationError{"question", "question is required"}
	}
	if len(req.Question) > MaxContentLength {
		return &ValidationError{"question", fmt.Sprintf("question exceeds max length of %d", MaxContentLength)}
	}
	if !utf8.ValidString(req.Question) {
		return &ValidationError{"question", "question must be valid UTF-8"}
	}

	if len(req.Options) == 0 {
		return &ValidationError{"options", "at least one option is required"}
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
	}

	if len(req.Criteria) == 0 {
		return &ValidationError{"criteria", "at least one criterion is required"}
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
			return &ValidationError{"criteria", fmt.Sprintf("criterion[%d].weight must be non-negative", i)}
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
