// Package modes - Tests for domain-specific model configuration
package modes

import (
	"os"
	"testing"
)

func TestDefaultDomainModels(t *testing.T) {
	config := DefaultDomainModels()

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	// Test that all domains have valid models
	domains := []struct {
		name   string
		config ModelConfig
	}{
		{"Code", config.Code},
		{"Research", config.Research},
		{"Quick", config.Quick},
		{"Default", config.Default},
	}

	for _, d := range domains {
		t.Run(d.name, func(t *testing.T) {
			if d.config.Model == "" {
				t.Errorf("%s: model should not be empty", d.name)
			}
			if d.config.MaxTokens <= 0 {
				t.Errorf("%s: max tokens should be positive, got %d", d.name, d.config.MaxTokens)
			}
			if d.config.Temperature < 0 || d.config.Temperature > 1 {
				t.Errorf("%s: temperature should be in [0, 1], got %f", d.name, d.config.Temperature)
			}
		})
	}
}

func TestGetModelForDomain(t *testing.T) {
	config := DefaultDomainModels()

	tests := []struct {
		domain   TaskDomain
		expected string
	}{
		{DomainCode, config.Code.Model},
		{DomainResearch, config.Research.Model},
		{DomainQuick, config.Quick.Model},
		{DomainDefault, config.Default.Model},
		{"unknown", config.Default.Model}, // Unknown domain should return default
	}

	for _, tt := range tests {
		t.Run(string(tt.domain), func(t *testing.T) {
			result := config.GetModelForDomain(tt.domain)
			if result.Model != tt.expected {
				t.Errorf("expected model %q, got %q", tt.expected, result.Model)
			}
		})
	}
}

func TestDetectDomainFromProblem(t *testing.T) {
	tests := []struct {
		problem  string
		expected TaskDomain
	}{
		// Code domain
		{"Debug this function that throws an error", DomainCode},
		{"Implement a new API endpoint for users", DomainCode},
		{"Refactor this Python class to use inheritance", DomainCode},
		{"Fix the bug in this JavaScript code", DomainCode},

		// Research domain (need >= 2 keyword matches)
		{"Research and analyze the effects of climate change on agriculture", DomainResearch},
		{"Analyze the correlation between education and income", DomainResearch},
		{"Investigate the hypothesis that sleep affects memory", DomainResearch},
		{"Study the methodology used in recent experiments", DomainResearch},

		// Quick domain
		{"What is the capital of France?", DomainQuick},
		{"Quick question about dates", DomainQuick},
		{"Summarize this in one sentence", DomainQuick},
		{"Brief", DomainQuick},

		// Default domain (mixed or unclear)
		{"Help me with my project", DomainDefault},
		{"I need to think about this problem", DomainDefault},
		{"What should I do next?", DomainDefault},
	}

	for _, tt := range tests {
		t.Run(tt.problem[:min(30, len(tt.problem))], func(t *testing.T) {
			result := DetectDomainFromProblem(tt.problem)
			if result != tt.expected {
				t.Errorf("problem %q: expected domain %q, got %q", tt.problem, tt.expected, result)
			}
		})
	}
}

func TestEnvVarOverrides(t *testing.T) {
	// Save original values
	origCode := os.Getenv("GOT_MODEL_CODE")
	origResearch := os.Getenv("GOT_MODEL_RESEARCH")
	origQuick := os.Getenv("GOT_MODEL_QUICK")
	origDefault := os.Getenv("GOT_MODEL")

	// Restore after test
	defer func() {
		setOrUnset("GOT_MODEL_CODE", origCode)
		setOrUnset("GOT_MODEL_RESEARCH", origResearch)
		setOrUnset("GOT_MODEL_QUICK", origQuick)
		setOrUnset("GOT_MODEL", origDefault)
	}()

	// Set test values
	os.Setenv("GOT_MODEL_CODE", "test-code-model")
	os.Setenv("GOT_MODEL_RESEARCH", "test-research-model")
	os.Setenv("GOT_MODEL_QUICK", "test-quick-model")
	os.Setenv("GOT_MODEL", "test-default-model")

	config := DefaultDomainModels()

	if config.Code.Model != "test-code-model" {
		t.Errorf("expected code model 'test-code-model', got %q", config.Code.Model)
	}
	if config.Research.Model != "test-research-model" {
		t.Errorf("expected research model 'test-research-model', got %q", config.Research.Model)
	}
	if config.Quick.Model != "test-quick-model" {
		t.Errorf("expected quick model 'test-quick-model', got %q", config.Quick.Model)
	}
	if config.Default.Model != "test-default-model" {
		t.Errorf("expected default model 'test-default-model', got %q", config.Default.Model)
	}
}

func TestCountKeywordMatches(t *testing.T) {
	tests := []struct {
		text     string
		keywords []string
		expected int
	}{
		{"implement a function", []string{"function", "method"}, 1},
		{"debug the code error", []string{"debug", "error", "bug"}, 3}, // "bug" is substring of "debug"
		{"hello world", []string{"code", "debug"}, 0},
		{"research analyze study", []string{"research", "analyze", "study"}, 3},
	}

	for _, tt := range tests {
		result := countKeywordMatches(tt.text, tt.keywords)
		if result != tt.expected {
			t.Errorf("text %q: expected %d matches, got %d", tt.text, tt.expected, result)
		}
	}
}

func setOrUnset(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
