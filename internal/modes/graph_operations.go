// Package modes - Graph-of-Thoughts operations (Generate, Aggregate, Refine, Score, Prune)
package modes

import (
	"context"
	"fmt"
	"time"
)

// Generate creates k diverse continuations from active or specified vertices
func (gc *GraphController) Generate(ctx context.Context, stateID string, llm LLMClient, req GenerateRequest) ([]*ThoughtVertex, error) {
	state, err := gc.GetState(stateID)
	if err != nil {
		return nil, err
	}

	// Use active vertices if no sources specified
	sourceIDs := req.SourceVertexIDs
	if len(sourceIDs) == 0 {
		sourceIDs = state.ActiveIDs
	}
	if len(sourceIDs) == 0 {
		return nil, fmt.Errorf("no source vertices for generation")
	}

	// Verify k is reasonable
	if req.K <= 0 || req.K > 10 {
		return nil, fmt.Errorf("k must be between 1 and 10, got %d", req.K)
	}

	newVertices := []*ThoughtVertex{}

	// Generate from each source
	for _, sourceID := range sourceIDs {
		source, ok := state.Vertices[sourceID]
		if !ok {
			continue
		}

		// Check depth limit
		if req.MaxDepth > 0 && source.Depth >= req.MaxDepth {
			continue
		}

		// Generate k continuations
		continuations, err := llm.Generate(ctx, source.Content, req.K)
		if err != nil {
			return nil, fmt.Errorf("LLM generation failed: %w", err)
		}

		// Create vertices for continuations
		for i, content := range continuations {
			vertexID := fmt.Sprintf("%s-gen-%s-%d", stateID, sourceID, i)

			vertex := NewThoughtVertex(vertexID, content, ThoughtTypeGenerated, 0.7)
			vertex.Depth = source.Depth + 1

			// Add to graph
			if err := gc.AddVertex(stateID, vertex); err != nil {
				return nil, fmt.Errorf("failed to add generated vertex: %w", err)
			}

			// Create edge from source
			edgeID := fmt.Sprintf("edge-%s-%s", sourceID, vertexID)
			edge := NewThoughtEdge(edgeID, sourceID, vertexID, EdgeTypeDerivesFrom, 0.9)
			if err := gc.AddEdge(stateID, edge); err != nil {
				return nil, fmt.Errorf("failed to add edge: %w", err)
			}

			newVertices = append(newVertices, vertex)
		}
	}

	// Update active vertices to newly generated ones
	newActiveIDs := make([]string, 0, len(newVertices))
	for _, v := range newVertices {
		newActiveIDs = append(newActiveIDs, v.ID)
	}
	if len(newActiveIDs) > 0 {
		_ = gc.SetActiveVertices(stateID, newActiveIDs)
	}

	return newVertices, nil
}

// Aggregate merges multiple vertices into a single synthesized thought
func (gc *GraphController) Aggregate(ctx context.Context, stateID string, llm LLMClient, req AggregateRequest) (*ThoughtVertex, error) {
	state, err := gc.GetState(stateID)
	if err != nil {
		return nil, err
	}

	// Verify minimum paths
	if len(req.VertexIDs) < state.Config.AggregateMinPaths {
		return nil, fmt.Errorf("need at least %d vertices to aggregate, got %d", state.Config.AggregateMinPaths, len(req.VertexIDs))
	}

	// Collect thoughts
	thoughts := make([]string, 0, len(req.VertexIDs))
	maxDepth := 0
	for _, vertexID := range req.VertexIDs {
		vertex, ok := state.Vertices[vertexID]
		if !ok {
			return nil, fmt.Errorf("vertex not found: %s", vertexID)
		}
		thoughts = append(thoughts, vertex.Content)
		if vertex.Depth > maxDepth {
			maxDepth = vertex.Depth
		}
	}

	// Synthesize
	aggregatedContent, err := llm.Aggregate(ctx, thoughts, req.Problem)
	if err != nil {
		return nil, fmt.Errorf("LLM aggregation failed: %w", err)
	}

	// Create aggregated vertex
	vertexID := fmt.Sprintf("%s-agg-%d", stateID, time.Now().UnixNano())
	vertex := NewThoughtVertex(vertexID, aggregatedContent, ThoughtTypeAggregated, 0.85)
	vertex.Depth = maxDepth + 1

	// Add to graph
	if err := gc.AddVertex(stateID, vertex); err != nil {
		return nil, fmt.Errorf("failed to add aggregated vertex: %w", err)
	}

	// Create edges from all source vertices
	for _, sourceID := range req.VertexIDs {
		edgeID := fmt.Sprintf("edge-%s-%s", sourceID, vertexID)
		edge := NewThoughtEdge(edgeID, sourceID, vertexID, EdgeTypeAggregates, 0.8)
		if err := gc.AddEdge(stateID, edge); err != nil {
			return nil, fmt.Errorf("failed to add aggregation edge: %w", err)
		}
	}

	return vertex, nil
}

// Refine iteratively improves a vertex through self-critique
func (gc *GraphController) Refine(ctx context.Context, stateID string, llm LLMClient, req RefineRequest) (*ThoughtVertex, error) {
	state, err := gc.GetState(stateID)
	if err != nil {
		return nil, err
	}

	source, ok := state.Vertices[req.VertexID]
	if !ok {
		return nil, fmt.Errorf("vertex not found: %s", req.VertexID)
	}

	// Check refinement limit
	if source.RefinedCount >= state.Config.MaxRefinements {
		return nil, fmt.Errorf("max refinements reached (%d)", state.Config.MaxRefinements)
	}

	// Refine
	refinedContent, err := llm.Refine(ctx, source.Content, req.Problem, source.RefinedCount)
	if err != nil {
		return nil, fmt.Errorf("LLM refinement failed: %w", err)
	}

	// Create refined vertex
	vertexID := fmt.Sprintf("%s-refined-%d", req.VertexID, source.RefinedCount+1)
	vertex := NewThoughtVertex(vertexID, refinedContent, ThoughtTypeRefined, 0.8)
	vertex.Depth = source.Depth
	vertex.RefinedCount = source.RefinedCount + 1

	// Add to graph
	if err := gc.AddVertex(stateID, vertex); err != nil {
		return nil, fmt.Errorf("failed to add refined vertex: %w", err)
	}

	// Create refinement edge (self-loop pattern)
	edgeID := fmt.Sprintf("edge-%s-%s", req.VertexID, vertexID)
	edge := NewThoughtEdge(edgeID, req.VertexID, vertexID, EdgeTypeRefines, 0.9)
	if err := gc.AddEdge(stateID, edge); err != nil {
		return nil, fmt.Errorf("failed to add refinement edge: %w", err)
	}

	return vertex, nil
}

// Score evaluates vertex quality with multi-criteria breakdown
func (gc *GraphController) Score(ctx context.Context, stateID string, llm LLMClient, req ScoreRequest) (*ScoreBreakdown, error) {
	state, err := gc.GetState(stateID)
	if err != nil {
		return nil, err
	}

	vertex, ok := state.Vertices[req.VertexID]
	if !ok {
		return nil, fmt.Errorf("vertex not found: %s", req.VertexID)
	}

	// Score criteria weights
	criteria := map[string]float64{
		"confidence":   0.25,
		"validity":     0.30,
		"relevance":    0.25,
		"novelty":      0.10,
		"depth_factor": 0.10,
	}

	// Get LLM-based scores
	_, breakdown, err := llm.Score(ctx, vertex.Content, req.Problem, criteria)
	if err != nil {
		return nil, fmt.Errorf("LLM scoring failed: %w", err)
	}

	// Build score breakdown
	result := &ScoreBreakdown{
		Confidence:  breakdown["confidence"],
		Validity:    breakdown["validity"],
		Relevance:   breakdown["relevance"],
		Novelty:     breakdown["novelty"],
		DepthFactor: breakdown["depth_factor"],
	}

	// Calculate weighted sum
	result.Overall = (result.Confidence * criteria["confidence"]) +
		(result.Validity * criteria["validity"]) +
		(result.Relevance * criteria["relevance"]) +
		(result.Novelty * criteria["novelty"]) +
		(result.DepthFactor * criteria["depth_factor"])

	// Update vertex score
	vertex.Score = result.Overall

	return result, nil
}

// Prune removes low-quality vertices below threshold
func (gc *GraphController) Prune(ctx context.Context, stateID string, threshold float64) (int, error) {
	state, err := gc.GetState(stateID)
	if err != nil {
		return 0, err
	}

	// Use config threshold if not specified
	if threshold == 0 {
		threshold = state.Config.PruneThreshold
	}

	// Identify vertices to prune (exclude roots and terminals)
	toPrune := []string{}
	for vertexID, vertex := range state.Vertices {
		// Never prune roots or terminals
		if containsStr(state.RootIDs, vertexID) || containsStr(state.TerminalIDs, vertexID) {
			continue
		}

		// Prune if score is below threshold
		if vertex.Score > 0 && vertex.Score < threshold {
			toPrune = append(toPrune, vertexID)
		}
	}

	// Remove vertices
	removed := 0
	for _, vertexID := range toPrune {
		if err := gc.RemoveVertex(stateID, vertexID); err == nil {
			removed++
		}
	}

	return removed, nil
}

// Helper function
func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
