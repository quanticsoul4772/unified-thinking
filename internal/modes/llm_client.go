// Package modes - LLM client interface for Graph-of-Thoughts
package modes

import "context"

// LLMClient defines the interface for LLM interactions
type LLMClient interface {
	// Generate creates k diverse continuations from a prompt
	Generate(ctx context.Context, prompt string, k int) ([]string, error)

	// Aggregate synthesizes multiple thoughts into one
	Aggregate(ctx context.Context, thoughts []string, problem string) (string, error)

	// Refine improves a thought through self-critique
	Refine(ctx context.Context, thought string, problem string, refinementCount int) (string, error)

	// Score evaluates thought quality (returns 0.0-1.0)
	Score(ctx context.Context, thought string, problem string, criteria map[string]float64) (float64, map[string]float64, error)

	// ExtractKeyPoints identifies key insights from a thought
	ExtractKeyPoints(ctx context.Context, thought string) ([]string, error)

	// CalculateNovelty measures uniqueness vs siblings
	CalculateNovelty(ctx context.Context, thought string, siblings []string) (float64, error)

	// ResearchWithSearch performs web-augmented research (optional, may return error if not supported)
	ResearchWithSearch(ctx context.Context, query string, problem string) (*ResearchResult, error)
}

// GenerateRequest encapsulates generation parameters
type GenerateRequest struct {
	SourceVertexIDs []string // Vertices to expand from
	K               int      // Number of continuations per source
	Problem         string   // Original problem context
	MaxDepth        int      // Depth limit
}

// AggregateRequest encapsulates aggregation parameters
type AggregateRequest struct {
	VertexIDs []string // Vertices to merge
	Problem   string   // Original problem context
}

// RefineRequest encapsulates refinement parameters
type RefineRequest struct {
	VertexID string // Vertex to refine
	Problem  string // Original problem context
}

// ScoreRequest encapsulates scoring parameters
type ScoreRequest struct {
	VertexID string // Vertex to score
	Problem  string // Original problem context
}
