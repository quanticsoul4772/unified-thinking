package errors

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewStructuredError(t *testing.T) {
	err := NewStructuredError(ErrThoughtNotFound, "Thought with ID 'xyz' not found")

	if err.Code != ErrThoughtNotFound {
		t.Errorf("Expected code %s, got %s", ErrThoughtNotFound, err.Code)
	}
	if err.Message != "Thought with ID 'xyz' not found" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
	if err.RecoverySuggestions == nil {
		t.Error("RecoverySuggestions should not be nil")
	}
}

func TestStructuredErrorWithDetails(t *testing.T) {
	err := NewStructuredError(ErrInvalidParameter, "Invalid parameter").
		WithDetails("Parameter 'mode' must be one of: linear, tree, divergent")

	if err.Details != "Parameter 'mode' must be one of: linear, tree, divergent" {
		t.Errorf("Unexpected details: %s", err.Details)
	}
}

func TestStructuredErrorWithRecovery(t *testing.T) {
	err := NewStructuredError(ErrBranchNotFound, "Branch not found").
		WithRecovery("Use list-branches to find available branches").
		WithRecovery("Create a new branch with think in tree mode")

	if len(err.RecoverySuggestions) != 2 {
		t.Errorf("Expected 2 recovery suggestions, got %d", len(err.RecoverySuggestions))
	}
}

func TestStructuredErrorWithRelatedTools(t *testing.T) {
	err := NewStructuredError(ErrSessionActive, "Session already active").
		WithRelatedTools("complete-reasoning-session", "search-trajectories")

	if len(err.RelatedTools) != 2 {
		t.Errorf("Expected 2 related tools, got %d", len(err.RelatedTools))
	}
}

func TestStructuredErrorWithExample(t *testing.T) {
	err := NewStructuredError(ErrInvalidMode, "Invalid thinking mode").
		WithExample("think", map[string]any{
			"content": "example content",
			"mode":    "linear",
		})

	if err.ExampleFix == nil {
		t.Error("ExampleFix should not be nil")
	}

	example, ok := err.ExampleFix.(map[string]any)
	if !ok {
		t.Fatal("ExampleFix should be a map")
	}

	if example["tool"] != "think" {
		t.Errorf("Expected tool 'think', got %v", example["tool"])
	}
}

func TestStructuredErrorError(t *testing.T) {
	err := NewStructuredError(ErrRateLimited, "Rate limited")
	errorString := err.Error()

	if errorString != "[ERR_5001_RATE_LIMITED] Rate limited" {
		t.Errorf("Unexpected error string: %s", errorString)
	}
}

func TestStructuredErrorJSONSerialization(t *testing.T) {
	err := NewStructuredError(ErrInvalidParameter, "Invalid parameter").
		WithDetails("Must provide content").
		WithRecovery("Add content field to request").
		WithRelatedTools("think").
		WithExample("think", map[string]any{"content": "example"})

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal error: %v", jsonErr)
	}

	var decoded StructuredError
	if jsonErr := json.Unmarshal(data, &decoded); jsonErr != nil {
		t.Fatalf("Failed to unmarshal error: %v", jsonErr)
	}

	if decoded.Code != err.Code {
		t.Errorf("Code mismatch after round-trip: %s != %s", decoded.Code, err.Code)
	}
	if decoded.Message != err.Message {
		t.Errorf("Message mismatch after round-trip: %s != %s", decoded.Message, err.Message)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := WrapError(ErrStorageFailed, originalErr)

	if wrapped.Code != ErrStorageFailed {
		t.Errorf("Expected code %s, got %s", ErrStorageFailed, wrapped.Code)
	}
	if wrapped.Message != "original error" {
		t.Errorf("Unexpected message: %s", wrapped.Message)
	}
}

func TestWrapErrorNil(t *testing.T) {
	wrapped := WrapError(ErrStorageFailed, nil)
	if wrapped != nil {
		t.Error("WrapError should return nil for nil input")
	}
}

func TestIsStructuredError(t *testing.T) {
	structErr := NewStructuredError(ErrThoughtNotFound, "Not found")
	regularErr := errors.New("regular error")

	if !IsStructuredError(structErr) {
		t.Error("IsStructuredError should return true for StructuredError")
	}
	if IsStructuredError(regularErr) {
		t.Error("IsStructuredError should return false for regular error")
	}
}

func TestAsStructuredError(t *testing.T) {
	structErr := NewStructuredError(ErrThoughtNotFound, "Not found")
	regularErr := errors.New("regular error")

	se, ok := AsStructuredError(structErr)
	if !ok || se == nil {
		t.Error("AsStructuredError should return the error for StructuredError")
	}

	se, ok = AsStructuredError(regularErr)
	if ok || se != nil {
		t.Error("AsStructuredError should return nil for regular error")
	}
}

func TestToStructuredError(t *testing.T) {
	// Test with StructuredError
	structErr := NewStructuredError(ErrThoughtNotFound, "Not found")
	result := ToStructuredError(structErr)
	if result.Code != ErrThoughtNotFound {
		t.Error("ToStructuredError should return unchanged StructuredError")
	}

	// Test with regular error
	regularErr := errors.New("regular error")
	result = ToStructuredError(regularErr)
	if result == nil {
		t.Error("ToStructuredError should wrap regular errors")
	}
	if result.Code != ErrInvalidOperation {
		t.Errorf("Expected generic code, got %s", result.Code)
	}

	// Test with nil
	result = ToStructuredError(nil)
	if result != nil {
		t.Error("ToStructuredError should return nil for nil input")
	}
}

func TestRecoveryGenerator(t *testing.T) {
	gen := NewRecoveryGenerator()

	// Test default recovery for known error
	suggestions := gen.GetSuggestions(ErrThoughtNotFound)
	if len(suggestions) == 0 {
		t.Error("Should have default recovery for ErrThoughtNotFound")
	}

	// Test unknown error code
	suggestions = gen.GetSuggestions("UNKNOWN_CODE")
	if len(suggestions) == 0 {
		t.Error("Should have generic recovery for unknown code")
	}
}

func TestRecoveryGeneratorRelatedTools(t *testing.T) {
	gen := NewRecoveryGenerator()

	tools := gen.GetRelatedTools(ErrThoughtNotFound)
	if len(tools) == 0 {
		t.Error("Should have related tools for ErrThoughtNotFound")
	}
}

func TestRecoveryGeneratorExample(t *testing.T) {
	gen := NewRecoveryGenerator()

	example := gen.GetExample(ErrThoughtNotFound)
	if example == nil {
		t.Error("Should have example for ErrThoughtNotFound")
	}
}

func TestRecoveryGeneratorEnhance(t *testing.T) {
	gen := NewRecoveryGenerator()
	err := NewStructuredError(ErrBranchNotFound, "Branch not found")

	enhanced := gen.Enhance(err)

	if len(enhanced.RecoverySuggestions) == 0 {
		t.Error("Enhanced error should have recovery suggestions")
	}
	if len(enhanced.RelatedTools) == 0 {
		t.Error("Enhanced error should have related tools")
	}
}

func TestEnhanceError(t *testing.T) {
	err := NewStructuredError(ErrSessionNotFound, "Session not found")

	enhanced := EnhanceError(err)

	if len(enhanced.RecoverySuggestions) == 0 {
		t.Error("EnhanceError should add recovery suggestions")
	}
}

func TestErrorCategory(t *testing.T) {
	tests := []struct {
		code     string
		category string
	}{
		{ErrThoughtNotFound, "resource"},
		{ErrInvalidParameter, "validation"},
		{ErrSessionActive, "state"},
		{ErrEmbeddingFailed, "external"},
		{ErrRateLimited, "limit"},
	}

	for _, tt := range tests {
		category := ErrorCategory(tt.code)
		if category != tt.category {
			t.Errorf("ErrorCategory(%s): got %s, want %s", tt.code, category, tt.category)
		}
	}
}

func TestIsRetryable(t *testing.T) {
	// External errors should be retryable
	if !IsRetryable(ErrEmbeddingFailed) {
		t.Error("External errors should be retryable")
	}
	if !IsRetryable(ErrLLMFailed) {
		t.Error("LLM errors should be retryable")
	}
	if !IsRetryable(ErrRateLimited) {
		t.Error("Rate limited should be retryable")
	}

	// Resource errors should not be retryable
	if IsRetryable(ErrThoughtNotFound) {
		t.Error("Resource errors should not be retryable")
	}
}

func TestErrorCategories(t *testing.T) {
	// Verify error codes follow the expected format
	tests := []struct {
		code     string
		category string
	}{
		{ErrThoughtNotFound, "1"},  // 1xxx = Resource errors
		{ErrInvalidParameter, "2"}, // 2xxx = Validation errors
		{ErrSessionActive, "3"},    // 3xxx = State errors
		{ErrEmbeddingFailed, "4"},  // 4xxx = External errors
		{ErrRateLimited, "5"},      // 5xxx = Limit errors
	}

	for _, tt := range tests {
		// Extract category from code (format: ERR_XXXX_...)
		if len(tt.code) < 5 {
			t.Errorf("Invalid code format: %s", tt.code)
			continue
		}
		// Code format: ERR_1001_NAME
		categoryDigit := string(tt.code[4])
		if categoryDigit != tt.category {
			t.Errorf("Code %s: expected category %s, got %s", tt.code, tt.category, categoryDigit)
		}
	}
}

func TestStructuredErrorChaining(t *testing.T) {
	err := NewStructuredError(ErrInvalidParameter, "Invalid parameter").
		WithDetails("Field 'content' is required").
		WithRecovery("Provide a non-empty content field").
		WithRecovery("Check the API documentation for required fields").
		WithRelatedTools("think", "history").
		WithExample("think", map[string]any{
			"content": "Your thought content here",
			"mode":    "linear",
		})

	// Verify all fields are set
	if err.Details == "" {
		t.Error("Details should be set")
	}
	if len(err.RecoverySuggestions) != 2 {
		t.Errorf("Expected 2 recovery suggestions, got %d", len(err.RecoverySuggestions))
	}
	if len(err.RelatedTools) != 2 {
		t.Errorf("Expected 2 related tools, got %d", len(err.RelatedTools))
	}
	if err.ExampleFix == nil {
		t.Error("ExampleFix should be set")
	}
}

func TestAllErrorCodesHaveRecovery(t *testing.T) {
	gen := NewRecoveryGenerator()

	codes := []string{
		ErrThoughtNotFound,
		ErrBranchNotFound,
		ErrSessionNotFound,
		ErrGraphNotFound,
		ErrCheckpointNotFound,
		ErrDecisionNotFound,
		ErrWorkflowNotFound,
		ErrPresetNotFound,
		ErrInvalidParameter,
		ErrMissingRequired,
		ErrInvalidMode,
		ErrInvalidConfidence,
		ErrInvalidFormat,
		ErrSessionActive,
		ErrSessionNotActive,
		ErrGraphFinalized,
		ErrEmbeddingFailed,
		ErrNeo4jConnection,
		ErrLLMFailed,
		ErrStorageFailed,
		ErrRateLimited,
		ErrContextTooLarge,
		ErrTooManyBranches,
		ErrMaxDepthReached,
	}

	for _, code := range codes {
		suggestions := gen.GetSuggestions(code)
		if len(suggestions) == 0 {
			t.Errorf("No recovery suggestions for code %s", code)
		}
	}
}

func TestToMap(t *testing.T) {
	err := NewStructuredError(ErrInvalidParameter, "Invalid parameter").
		WithDetails("Must provide content").
		WithRecovery("Add content field").
		WithRelatedTools("think").
		WithExample("think", map[string]any{"content": "example"})

	m := err.ToMap()

	if m["error_code"] != ErrInvalidParameter {
		t.Errorf("Expected error_code %s, got %v", ErrInvalidParameter, m["error_code"])
	}
	if m["message"] != "Invalid parameter" {
		t.Errorf("Expected message 'Invalid parameter', got %v", m["message"])
	}
	if m["details"] != "Must provide content" {
		t.Errorf("Expected details 'Must provide content', got %v", m["details"])
	}
}
