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
