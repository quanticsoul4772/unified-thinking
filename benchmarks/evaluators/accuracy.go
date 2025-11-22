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
