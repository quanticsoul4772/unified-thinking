package evaluators

import (
	"testing"
)

func TestExactMatchEvaluator_Evaluate(t *testing.T) {
	eval := NewExactMatchEvaluator()

	tests := []struct {
		name        string
		response    interface{}
		expected    interface{}
		wantCorrect bool
		wantScore   float64
		wantErr     bool
	}{
		{
			name:        "exact match - same case",
			response:    "valid",
			expected:    "valid",
			wantCorrect: true,
			wantScore:   1.0,
			wantErr:     false,
		},
		{
			name:        "exact match - different case",
			response:    "Valid",
			expected:    "valid",
			wantCorrect: true,
			wantScore:   1.0,
			wantErr:     false,
		},
		{
			name:        "exact match - with whitespace",
			response:    "  valid  ",
			expected:    "valid",
			wantCorrect: true,
			wantScore:   1.0,
			wantErr:     false,
		},
		{
			name:        "no match",
			response:    "invalid",
			expected:    "valid",
			wantCorrect: false,
			wantScore:   0.0,
			wantErr:     false,
		},
		{
			name:        "empty strings match",
			response:    "",
			expected:    "",
			wantCorrect: true,
			wantScore:   1.0,
			wantErr:     false,
		},
		{
			name:        "response not string",
			response:    123,
			expected:    "valid",
			wantCorrect: false,
			wantScore:   0.0,
			wantErr:     true,
		},
		{
			name:        "expected not string",
			response:    "valid",
			expected:    123,
			wantCorrect: false,
			wantScore:   0.0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correct, score, err := eval.Evaluate(tt.response, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if correct != tt.wantCorrect {
				t.Errorf("Evaluate() correct = %v, want %v", correct, tt.wantCorrect)
			}

			if score != tt.wantScore {
				t.Errorf("Evaluate() score = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestExactMatchEvaluator_Name(t *testing.T) {
	eval := NewExactMatchEvaluator()
	if got := eval.Name(); got != "exact_match" {
		t.Errorf("Name() = %v, want %v", got, "exact_match")
	}
}

func TestContainsEvaluator_Evaluate(t *testing.T) {
	eval := NewContainsEvaluator()

	tests := []struct {
		name        string
		response    interface{}
		expected    interface{}
		wantCorrect bool
		wantScore   float64
		wantErr     bool
	}{
		{
			name:        "contains - exact substring",
			response:    "The answer is valid",
			expected:    "valid",
			wantCorrect: true,
			wantScore:   1.0,
			wantErr:     false,
		},
		{
			name:        "contains - case insensitive",
			response:    "The answer is VALID",
			expected:    "valid",
			wantCorrect: true,
			wantScore:   1.0,
			wantErr:     false,
		},
		{
			name:        "contains - within longer text",
			response:    "This reasoning is invalid because...",
			expected:    "invalid",
			wantCorrect: true,
			wantScore:   1.0,
			wantErr:     false,
		},
		{
			name:        "does not contain",
			response:    "The answer is correct",
			expected:    "invalid",
			wantCorrect: false,
			wantScore:   0.0,
			wantErr:     false,
		},
		{
			name:        "empty expected in response",
			response:    "anything",
			expected:    "",
			wantCorrect: true,
			wantScore:   1.0,
			wantErr:     false,
		},
		{
			name:        "empty response, non-empty expected",
			response:    "",
			expected:    "valid",
			wantCorrect: false,
			wantScore:   0.0,
			wantErr:     false,
		},
		{
			name:        "response not string",
			response:    []string{"valid"},
			expected:    "valid",
			wantCorrect: false,
			wantScore:   0.0,
			wantErr:     true,
		},
		{
			name:        "expected not string",
			response:    "valid",
			expected:    true,
			wantCorrect: false,
			wantScore:   0.0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correct, score, err := eval.Evaluate(tt.response, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if correct != tt.wantCorrect {
				t.Errorf("Evaluate() correct = %v, want %v", correct, tt.wantCorrect)
			}

			if score != tt.wantScore {
				t.Errorf("Evaluate() score = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestContainsEvaluator_Name(t *testing.T) {
	eval := NewContainsEvaluator()
	if got := eval.Name(); got != "contains" {
		t.Errorf("Name() = %v, want %v", got, "contains")
	}
}
