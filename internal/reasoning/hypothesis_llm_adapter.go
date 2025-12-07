package reasoning

import (
	"context"
)

// TextGenerator interface for raw text generation (used for hypothesis generation)
type TextGenerator interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

// AnthropicHypothesisGenerator adapts modes.AnthropicLLMClient for hypothesis generation
type AnthropicHypothesisGenerator struct {
	client TextGenerator
}

// NewAnthropicHypothesisGenerator creates hypothesis generator from Anthropic client
func NewAnthropicHypothesisGenerator(client TextGenerator) *AnthropicHypothesisGenerator {
	return &AnthropicHypothesisGenerator{client: client}
}

// GenerateHypotheses implements HypothesisGenerator interface
func (g *AnthropicHypothesisGenerator) GenerateHypotheses(ctx context.Context, prompt string) (string, error) {
	// Use GenerateText for free-form hypothesis generation (not structured output)
	return g.client.GenerateText(ctx, prompt)
}
