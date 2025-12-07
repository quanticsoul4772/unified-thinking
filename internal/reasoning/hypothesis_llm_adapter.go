package reasoning

import (
	"context"
	"fmt"
)

// AnthropicHypothesisGenerator adapts modes.AnthropicLLMClient for hypothesis generation
type AnthropicHypothesisGenerator struct {
	client interface {
		Generate(ctx context.Context, prompt string, k int) ([]string, error)
	}
}

// NewAnthropicHypothesisGenerator creates hypothesis generator from Anthropic client
func NewAnthropicHypothesisGenerator(client interface {
	Generate(ctx context.Context, prompt string, k int) ([]string, error)
}) *AnthropicHypothesisGenerator {
	return &AnthropicHypothesisGenerator{client: client}
}

// GenerateHypotheses implements HypothesisGenerator interface
func (g *AnthropicHypothesisGenerator) GenerateHypotheses(ctx context.Context, prompt string) (string, error) {
	// Generate single response (k=1) for hypothesis generation
	responses, err := g.client.Generate(ctx, prompt, 1)
	if err != nil {
		return "", err
	}
	if len(responses) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}
	return responses[0], nil
}
