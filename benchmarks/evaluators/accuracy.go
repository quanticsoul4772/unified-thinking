// Package evaluators provides benchmark evaluation implementations.
package evaluators

import (
	"fmt"
	"strings"
)

// ExactMatchEvaluator checks for exact string matches
type ExactMatchEvaluator struct{}

// NewExactMatchEvaluator creates a new exact match evaluator
func NewExactMatchEvaluator() *ExactMatchEvaluator {
	return &ExactMatchEvaluator{}
}

// Evaluate compares response to expected for exact match
func (e *ExactMatchEvaluator) Evaluate(response interface{}, expected interface{}) (bool, float64, error) {
	responseStr, ok := response.(string)
	if !ok {
		return false, 0.0, fmt.Errorf("response must be string, got %T", response)
	}

	expectedStr, ok := expected.(string)
	if !ok {
		return false, 0.0, fmt.Errorf("expected must be string, got %T", expected)
	}

	// Normalize for comparison
	responseNorm := strings.TrimSpace(strings.ToLower(responseStr))
	expectedNorm := strings.TrimSpace(strings.ToLower(expectedStr))

	correct := responseNorm == expectedNorm
	score := 0.0
	if correct {
		score = 1.0
	}

	return correct, score, nil
}

// Name returns the evaluator name
func (e *ExactMatchEvaluator) Name() string {
	return "exact_match"
}

// ContainsEvaluator checks if response contains expected substring
type ContainsEvaluator struct{}

// NewContainsEvaluator creates a new contains evaluator
func NewContainsEvaluator() *ContainsEvaluator {
	return &ContainsEvaluator{}
}

// Evaluate checks if response contains expected substring
func (e *ContainsEvaluator) Evaluate(response interface{}, expected interface{}) (bool, float64, error) {
	responseStr, ok := response.(string)
	if !ok {
		return false, 0.0, fmt.Errorf("response must be string, got %T", response)
	}

	expectedStr, ok := expected.(string)
	if !ok {
		return false, 0.0, fmt.Errorf("expected must be string, got %T", expected)
	}

	// Normalize for comparison
	responseNorm := strings.ToLower(responseStr)
	expectedNorm := strings.ToLower(expectedStr)

	correct := strings.Contains(responseNorm, expectedNorm)
	score := 0.0
	if correct {
		score = 1.0
	}

	return correct, score, nil
}

// Name returns the evaluator name
func (e *ContainsEvaluator) Name() string {
	return "contains"
}

// NumericEvaluator checks if numeric response is within tolerance of expected
type NumericEvaluator struct {
	tolerance float64
}

// NewNumericEvaluator creates a new numeric evaluator with default tolerance 0.05
func NewNumericEvaluator() *NumericEvaluator {
	return &NumericEvaluator{tolerance: 0.05}
}

// Evaluate compares numeric response to expected within tolerance
func (e *NumericEvaluator) Evaluate(response interface{}, expected interface{}) (bool, float64, error) {
	// Convert response to float
	var responseVal float64
	switch r := response.(type) {
	case float64:
		responseVal = r
	case string:
		// Try parsing string as float
		var parsed float64
		if _, err := fmt.Sscanf(r, "%f", &parsed); err == nil {
			responseVal = parsed
		} else {
			return false, 0.0, fmt.Errorf("response string cannot be parsed as float: %s", r)
		}
	default:
		return false, 0.0, fmt.Errorf("response must be number or string, got %T", response)
	}

	// Convert expected to float
	var expectedVal float64
	switch exp := expected.(type) {
	case float64:
		expectedVal = exp
	case string:
		var parsed float64
		if _, err := fmt.Sscanf(exp, "%f", &parsed); err == nil {
			expectedVal = parsed
		} else {
			return false, 0.0, fmt.Errorf("expected string cannot be parsed as float: %s", exp)
		}
	default:
		return false, 0.0, fmt.Errorf("expected must be number or string, got %T", expected)
	}

	// Check if within tolerance
	diff := responseVal - expectedVal
	if diff < 0 {
		diff = -diff
	}

	correct := diff <= e.tolerance
	score := 0.0
	if correct {
		score = 1.0
	}

	return correct, score, nil
}

// Name returns the evaluator name
func (e *NumericEvaluator) Name() string {
	return "numeric"
}
