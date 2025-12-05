// Package modes - Domain-specific model configuration for LLM operations
package modes

import (
	"os"
	"strings"
)

// TaskDomain represents the type of task being performed
type TaskDomain string

const (
	// DomainCode represents code-related tasks (generation, review, debugging)
	DomainCode TaskDomain = "code"
	// DomainResearch represents research and analysis tasks
	DomainResearch TaskDomain = "research"
	// DomainQuick represents quick, simple tasks requiring fast responses
	DomainQuick TaskDomain = "quick"
	// DomainDefault represents general-purpose tasks
	DomainDefault TaskDomain = "default"
)

// ModelConfig holds the configuration for a specific model
type ModelConfig struct {
	Model       string  // Model ID (e.g., "claude-sonnet-4-5-20250929")
	MaxTokens   int     // Maximum tokens for this model
	Temperature float64 // Temperature setting (0.0-1.0)
}

// DomainModelConfig holds model configurations for different task domains
type DomainModelConfig struct {
	Code     ModelConfig
	Research ModelConfig
	Quick    ModelConfig
	Default  ModelConfig
}

// DefaultDomainModels returns the default model configuration
// These can be overridden via environment variables
func DefaultDomainModels() *DomainModelConfig {
	return &DomainModelConfig{
		Code: ModelConfig{
			Model:       getEnvOrDefault("GOT_MODEL_CODE", "claude-sonnet-4-5-20250929"),
			MaxTokens:   4096,
			Temperature: 0.3, // Lower temperature for code
		},
		Research: ModelConfig{
			Model:       getEnvOrDefault("GOT_MODEL_RESEARCH", "claude-sonnet-4-5-20250929"),
			MaxTokens:   8192,
			Temperature: 0.7, // Higher temperature for creative research
		},
		Quick: ModelConfig{
			Model:       getEnvOrDefault("GOT_MODEL_QUICK", "claude-3-5-haiku-20241022"),
			MaxTokens:   1024,
			Temperature: 0.5,
		},
		Default: ModelConfig{
			Model:       getEnvOrDefault("GOT_MODEL", "claude-sonnet-4-5-20250929"),
			MaxTokens:   2048,
			Temperature: 0.5,
		},
	}
}

// GetModelForDomain returns the appropriate model configuration for a task domain
func (c *DomainModelConfig) GetModelForDomain(domain TaskDomain) ModelConfig {
	switch domain {
	case DomainCode:
		return c.Code
	case DomainResearch:
		return c.Research
	case DomainQuick:
		return c.Quick
	default:
		return c.Default
	}
}

// DetectDomainFromProblem analyzes a problem statement to determine the task domain
func DetectDomainFromProblem(problem string) TaskDomain {
	problemLower := strings.ToLower(problem)

	// Code-related keywords
	codeKeywords := []string{
		"code", "function", "debug", "error", "bug", "implement", "refactor",
		"class", "method", "api", "database", "sql", "javascript", "python",
		"golang", "typescript", "rust", "java", "programming", "algorithm",
		"compile", "syntax", "variable", "loop", "array", "object",
	}

	// Research-related keywords
	researchKeywords := []string{
		"research", "analyze", "study", "investigate", "compare", "evaluate",
		"literature", "theory", "hypothesis", "experiment", "methodology",
		"evidence", "conclusion", "findings", "paper", "journal", "academic",
		"scientific", "data", "statistics", "correlation", "causation",
	}

	// Quick/simple task keywords
	quickKeywords := []string{
		"simple", "quick", "brief", "short", "summarize", "define", "what is",
		"explain briefly", "one sentence", "quick question", "just tell me",
	}

	// Count keyword matches
	codeScore := countKeywordMatches(problemLower, codeKeywords)
	researchScore := countKeywordMatches(problemLower, researchKeywords)
	quickScore := countKeywordMatches(problemLower, quickKeywords)

	// Determine domain based on highest score
	if quickScore >= 2 || (len(problem) < 50 && quickScore > 0) {
		return DomainQuick
	}
	if codeScore > researchScore && codeScore >= 2 {
		return DomainCode
	}
	if researchScore > codeScore && researchScore >= 2 {
		return DomainResearch
	}

	return DomainDefault
}

// countKeywordMatches counts how many keywords appear in the text
func countKeywordMatches(text string, keywords []string) int {
	count := 0
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			count++
		}
	}
	return count
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
