// Package modes - Mock LLM client for testing
package modes

import (
	"context"
	"fmt"
	"strings"
)

// MockLLMClient provides deterministic responses for testing
type MockLLMClient struct {
	GenerateResponses []string
	GenerateIndex     int
}

// NewMockLLMClient creates a new mock client
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		GenerateResponses: []string{
			"Generated continuation 1",
			"Generated continuation 2",
			"Generated continuation 3",
		},
		GenerateIndex: 0,
	}
}

// Generate returns mock continuations
func (m *MockLLMClient) Generate(ctx context.Context, prompt string, k int) ([]string, error) {
	results := make([]string, k)
	for i := 0; i < k; i++ {
		idx := (m.GenerateIndex + i) % len(m.GenerateResponses)
		results[i] = fmt.Sprintf("%s (from: %s)", m.GenerateResponses[idx], truncate(prompt, 30))
		m.GenerateIndex++
	}
	return results, nil
}

// Aggregate returns mock synthesis
func (m *MockLLMClient) Aggregate(ctx context.Context, thoughts []string, problem string) (string, error) {
	return fmt.Sprintf("Aggregated from %d thoughts: %s", len(thoughts), strings.Join(thoughts, " + ")), nil
}

// Refine returns mock refinement
func (m *MockLLMClient) Refine(ctx context.Context, thought string, problem string, refinementCount int) (string, error) {
	return fmt.Sprintf("Refined v%d: %s", refinementCount+1, thought), nil
}

// Score returns mock quality scores
func (m *MockLLMClient) Score(ctx context.Context, thought string, problem string, criteria map[string]float64) (float64, map[string]float64, error) {
	breakdown := map[string]float64{
		"confidence":   0.8,
		"validity":     0.9,
		"relevance":    0.7,
		"novelty":      0.6,
		"depth_factor": 0.8,
	}
	overall := 0.0
	for _, weight := range criteria {
		overall += weight
	}
	return overall / float64(len(criteria)), breakdown, nil
}

// ExtractKeyPoints returns mock key points
func (m *MockLLMClient) ExtractKeyPoints(ctx context.Context, thought string) ([]string, error) {
	return []string{
		"Key point 1",
		"Key point 2",
		"Key point 3",
	}, nil
}

// CalculateNovelty returns mock novelty score
func (m *MockLLMClient) CalculateNovelty(ctx context.Context, thought string, siblings []string) (float64, error) {
	if len(siblings) == 0 {
		return 1.0, nil
	}
	return 0.7, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
