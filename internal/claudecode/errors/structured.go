package errors

import (
	"encoding/json"
	"fmt"
)

// StructuredError provides actionable error information for Claude Code
type StructuredError struct {
	// Code is the error code (e.g., "ERR_1001_THOUGHT_NOT_FOUND")
	Code string `json:"error_code"`
	// Message is a human-readable error message
	Message string `json:"message"`
	// Details provides additional context about the error
	Details string `json:"details,omitempty"`
	// RecoverySuggestions are actionable steps to resolve the error
	RecoverySuggestions []string `json:"recovery_suggestions"`
	// RelatedTools are tools that might help resolve the error
	RelatedTools []string `json:"related_tools,omitempty"`
	// ExampleFix shows an example of how to fix the error
	ExampleFix any `json:"example_fix,omitempty"`
	// Cause is the underlying error if any
	Cause error `json:"-"`
}

// Error implements the error interface
func (e *StructuredError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *StructuredError) Unwrap() error {
	return e.Cause
}

// MarshalJSON implements custom JSON marshaling
func (e *StructuredError) MarshalJSON() ([]byte, error) {
	type alias StructuredError
	return json.Marshal((*alias)(e))
}

// NewStructuredError creates a new StructuredError with the given code and message
func NewStructuredError(code, message string) *StructuredError {
	return &StructuredError{
		Code:                code,
		Message:             message,
		RecoverySuggestions: make([]string, 0),
		RelatedTools:        make([]string, 0),
	}
}

// WrapError wraps an existing error with structured information
func WrapError(code string, err error) *StructuredError {
	if err == nil {
		return nil
	}
	return &StructuredError{
		Code:                code,
		Message:             err.Error(),
		Cause:               err,
		RecoverySuggestions: make([]string, 0),
		RelatedTools:        make([]string, 0),
	}
}

// WithDetails adds details to the error
func (e *StructuredError) WithDetails(details string) *StructuredError {
	e.Details = details
	return e
}

// WithRecovery adds a recovery suggestion
func (e *StructuredError) WithRecovery(suggestion string) *StructuredError {
	e.RecoverySuggestions = append(e.RecoverySuggestions, suggestion)
	return e
}

// WithRecoveries adds multiple recovery suggestions
func (e *StructuredError) WithRecoveries(suggestions ...string) *StructuredError {
	e.RecoverySuggestions = append(e.RecoverySuggestions, suggestions...)
	return e
}

// WithRelatedTools adds related tools
func (e *StructuredError) WithRelatedTools(tools ...string) *StructuredError {
	e.RelatedTools = append(e.RelatedTools, tools...)
	return e
}

// WithExample adds an example fix
func (e *StructuredError) WithExample(tool string, params map[string]any) *StructuredError {
	e.ExampleFix = map[string]any{
		"tool":   tool,
		"params": params,
	}
	return e
}

// WithCause sets the underlying cause
func (e *StructuredError) WithCause(err error) *StructuredError {
	e.Cause = err
	return e
}

// ToMap converts the error to a map for JSON response
func (e *StructuredError) ToMap() map[string]any {
	result := map[string]any{
		"error_code":           e.Code,
		"message":              e.Message,
		"recovery_suggestions": e.RecoverySuggestions,
	}
	if e.Details != "" {
		result["details"] = e.Details
	}
	if len(e.RelatedTools) > 0 {
		result["related_tools"] = e.RelatedTools
	}
	if e.ExampleFix != nil {
		result["example_fix"] = e.ExampleFix
	}
	return result
}

// IsStructuredError checks if an error is a StructuredError
func IsStructuredError(err error) bool {
	_, ok := err.(*StructuredError)
	return ok
}

// AsStructuredError converts an error to a StructuredError if possible
func AsStructuredError(err error) (*StructuredError, bool) {
	se, ok := err.(*StructuredError)
	return se, ok
}

// ToStructuredError converts any error to a StructuredError
// If the error is already a StructuredError, it returns it unchanged
// Otherwise, it wraps the error with a generic code
func ToStructuredError(err error) *StructuredError {
	if err == nil {
		return nil
	}
	if se, ok := err.(*StructuredError); ok {
		return se
	}
	return WrapError(ErrInvalidOperation, err)
}
