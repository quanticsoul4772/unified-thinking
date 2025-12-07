package reasoning

import (
	"context"
)

// AnthropicDecompositionGenerator adapts modes.AnthropicLLMClient for decomposition generation
// Uses TextGenerator interface defined in hypothesis_llm_adapter.go
type AnthropicDecompositionGenerator struct {
	client TextGenerator
}

// NewAnthropicDecompositionGenerator creates decomposition generator from Anthropic client
func NewAnthropicDecompositionGenerator(client TextGenerator) *AnthropicDecompositionGenerator {
	return &AnthropicDecompositionGenerator{client: client}
}

// GenerateDecomposition implements DecompositionGenerator interface
func (g *AnthropicDecompositionGenerator) GenerateDecomposition(ctx context.Context, prompt string) (string, error) {
	return g.client.GenerateText(ctx, prompt)
}
