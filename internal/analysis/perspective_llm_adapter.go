package analysis

import (
	"context"
)

// TextGenerator interface for raw text generation (same as reasoning.TextGenerator)
type TextGenerator interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

// AnthropicPerspectiveGenerator adapts modes.AnthropicLLMClient for perspective generation
type AnthropicPerspectiveGenerator struct {
	client TextGenerator
}

// NewAnthropicPerspectiveGenerator creates perspective generator from Anthropic client
func NewAnthropicPerspectiveGenerator(client TextGenerator) *AnthropicPerspectiveGenerator {
	return &AnthropicPerspectiveGenerator{client: client}
}

// GeneratePerspectives implements PerspectiveGenerator interface
func (g *AnthropicPerspectiveGenerator) GeneratePerspectives(ctx context.Context, prompt string) (string, error) {
	return g.client.GenerateText(ctx, prompt)
}
