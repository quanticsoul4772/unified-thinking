// Package modes - Graph-of-Thoughts operations (Generate, Aggregate, Refine, Score, Prune)
package modes

import (
	"context"
	"fmt"
	"strings"
	"sync"
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

// ExploreRequest encapsulates the auto-orchestrated exploration parameters
type ExploreRequest struct {
	InitialThought string         `json:"initial_thought"`
	Problem        string         `json:"problem"`
	Config         *ExploreConfig `json:"config,omitempty"`
}

// ExploreConfig controls the exploration workflow
type ExploreConfig struct {
	K               int     `json:"k"`                // Continuations per step (default: 3)
	MaxIterations   int     `json:"max_iterations"`   // Max exploration cycles (default: 1)
	PruneThreshold  float64 `json:"prune_threshold"`  // Score threshold for pruning (default: 0.3)
	RefineTopN      int     `json:"refine_top_n"`     // Refine top N vertices (default: 1)
	ScoreAll        bool    `json:"score_all"`        // Score all vertices, not just active (default: false)
	UseFastScoring  bool    `json:"use_fast_scoring"` // Use local heuristics instead of LLM for scoring (default: true)
	SkipRefine      bool    `json:"skip_refine"`      // Skip the refinement step to save LLM calls (default: false)
	ParallelScoring bool    `json:"parallel_scoring"` // Parallelize LLM scoring calls (default: true)
}

// ExploreResult contains the orchestrated exploration results
type ExploreResult struct {
	GraphID         string            `json:"graph_id"`
	Problem         string            `json:"problem"`
	Iterations      int               `json:"iterations"`
	TotalGenerated  int               `json:"total_generated"`
	TotalPruned     int               `json:"total_pruned"`
	TotalRefined    int               `json:"total_refined"`
	BestVertices    []*ThoughtVertex  `json:"best_vertices"`
	Conclusions     []*ThoughtVertex  `json:"conclusions"`
	ExplorationPath []ExplorationStep `json:"exploration_path"`
}

// ExplorationStep records a single step in the exploration workflow
type ExplorationStep struct {
	Step        int    `json:"step"`
	Action      string `json:"action"`       // "generate", "score", "prune", "refine", "finalize"
	VertexCount int    `json:"vertex_count"` // Vertices affected
	Details     string `json:"details"`      // Human-readable description
}

// DefaultExploreConfig returns sensible defaults for exploration
// Optimized for speed by default - uses local scoring to avoid LLM API calls
func DefaultExploreConfig() *ExploreConfig {
	return &ExploreConfig{
		K:               3,
		MaxIterations:   1, // Reduced from 2 to minimize API calls
		PruneThreshold:  0.3,
		RefineTopN:      1,
		ScoreAll:        false,
		UseFastScoring:  true,  // Use local heuristics by default for speed
		SkipRefine:      false, // Refinement adds value, keep it
		ParallelScoring: true,  // Parallelize LLM calls when UseFastScoring=false
	}
}

// ThoroughExploreConfig returns config for thorough exploration with LLM scoring
// Use this when quality matters more than speed
func ThoroughExploreConfig() *ExploreConfig {
	return &ExploreConfig{
		K:               3,
		MaxIterations:   2,
		PruneThreshold:  0.3,
		RefineTopN:      2,
		ScoreAll:        false,
		UseFastScoring:  false, // Use LLM for accurate scoring
		SkipRefine:      false,
		ParallelScoring: true, // Still parallelize for speed
	}
}

// Explore orchestrates a complete Graph-of-Thoughts workflow automatically
// This combines: initialize → generate → score → prune → refine → finalize
func (gc *GraphController) Explore(ctx context.Context, graphID string, llm LLMClient, req ExploreRequest) (*ExploreResult, error) {
	config := req.Config
	if config == nil {
		config = DefaultExploreConfig()
	}

	// Validate inputs
	if graphID == "" {
		return nil, fmt.Errorf("graph_id is required")
	}
	if req.InitialThought == "" {
		return nil, fmt.Errorf("initial_thought is required")
	}
	if req.Problem == "" {
		return nil, fmt.Errorf("problem is required")
	}

	result := &ExploreResult{
		GraphID:         graphID,
		Problem:         req.Problem,
		ExplorationPath: []ExplorationStep{},
	}
	stepNum := 0

	// Step 1: Initialize the graph
	graphConfig := DefaultGraphConfig()
	graphConfig.PruneThreshold = config.PruneThreshold

	state, err := gc.Initialize(graphID, req.InitialThought, graphConfig)
	if err != nil {
		return nil, fmt.Errorf("initialization failed: %w", err)
	}

	stepNum++
	result.ExplorationPath = append(result.ExplorationPath, ExplorationStep{
		Step:        stepNum,
		Action:      "initialize",
		VertexCount: 1,
		Details:     fmt.Sprintf("Created graph with initial thought: %s...", truncateStr(req.InitialThought, 50)),
	})

	// Exploration loop
	for iteration := 0; iteration < config.MaxIterations; iteration++ {
		// Step 2: Generate k diverse continuations
		genReq := GenerateRequest{
			K:       config.K,
			Problem: req.Problem,
		}

		generated, err := gc.Generate(ctx, graphID, llm, genReq)
		if err != nil {
			return nil, fmt.Errorf("generation failed at iteration %d: %w", iteration, err)
		}
		result.TotalGenerated += len(generated)
		result.Iterations = iteration + 1

		stepNum++
		result.ExplorationPath = append(result.ExplorationPath, ExplorationStep{
			Step:        stepNum,
			Action:      "generate",
			VertexCount: len(generated),
			Details:     fmt.Sprintf("Iteration %d: Generated %d continuations", iteration+1, len(generated)),
		})

		if len(generated) == 0 {
			break // No more to generate
		}

		// Step 3: Score all active vertices
		verticesToScore := state.ActiveIDs
		if config.ScoreAll {
			verticesToScore = make([]string, 0, len(state.Vertices))
			for id := range state.Vertices {
				verticesToScore = append(verticesToScore, id)
			}
		}

		scoredCount := 0
		scoringMethod := "llm"

		if config.UseFastScoring {
			// Fast local scoring - no LLM calls, ~1ms per vertex
			scoringMethod = "fast"
			for _, vertexID := range verticesToScore {
				if vertex, ok := state.Vertices[vertexID]; ok {
					vertex.Score = fastScoreVertex(vertex, req.Problem)
					scoredCount++
				}
			}
		} else if config.ParallelScoring && len(verticesToScore) > 1 {
			// Parallel LLM scoring - significantly faster than sequential
			scoringMethod = "llm-parallel"
			var wg sync.WaitGroup
			var mu sync.Mutex
			results := make(map[string]bool)

			for _, vertexID := range verticesToScore {
				wg.Add(1)
				go func(vID string) {
					defer wg.Done()
					scoreReq := ScoreRequest{
						VertexID: vID,
						Problem:  req.Problem,
					}
					_, err := gc.Score(ctx, graphID, llm, scoreReq)
					mu.Lock()
					results[vID] = (err == nil)
					mu.Unlock()
				}(vertexID)
			}
			wg.Wait()

			for _, success := range results {
				if success {
					scoredCount++
				}
			}
		} else {
			// Sequential LLM scoring (slowest)
			for _, vertexID := range verticesToScore {
				scoreReq := ScoreRequest{
					VertexID: vertexID,
					Problem:  req.Problem,
				}
				if _, err := gc.Score(ctx, graphID, llm, scoreReq); err == nil {
					scoredCount++
				}
			}
		}

		stepNum++
		result.ExplorationPath = append(result.ExplorationPath, ExplorationStep{
			Step:        stepNum,
			Action:      "score",
			VertexCount: scoredCount,
			Details:     fmt.Sprintf("Scored %d vertices (%s)", scoredCount, scoringMethod),
		})

		// Step 4: Prune low-quality vertices
		pruned, err := gc.Prune(ctx, graphID, config.PruneThreshold)
		if err != nil {
			return nil, fmt.Errorf("pruning failed: %w", err)
		}
		result.TotalPruned += pruned

		if pruned > 0 {
			stepNum++
			result.ExplorationPath = append(result.ExplorationPath, ExplorationStep{
				Step:        stepNum,
				Action:      "prune",
				VertexCount: pruned,
				Details:     fmt.Sprintf("Pruned %d vertices below threshold %.2f", pruned, config.PruneThreshold),
			})
		}

		// Refresh state after pruning
		state, _ = gc.GetState(graphID)

		// Step 5: Refine top N vertices (skip if SkipRefine is true)
		if !config.SkipRefine {
			topVertices := getTopScoredVertices(state, config.RefineTopN)
			refinedCount := 0
			for _, vertex := range topVertices {
				refReq := RefineRequest{
					VertexID: vertex.ID,
					Problem:  req.Problem,
				}
				if _, err := gc.Refine(ctx, graphID, llm, refReq); err == nil {
					refinedCount++
				}
			}
			result.TotalRefined += refinedCount

			if refinedCount > 0 {
				stepNum++
				result.ExplorationPath = append(result.ExplorationPath, ExplorationStep{
					Step:        stepNum,
					Action:      "refine",
					VertexCount: refinedCount,
					Details:     fmt.Sprintf("Refined %d top-scoring vertices", refinedCount),
				})
			}

			// Refresh state after refinement
			state, _ = gc.GetState(graphID)
		}
	}

	// Step 6: Finalize - select best vertices as conclusions
	state, _ = gc.GetState(graphID)
	bestVertices := getTopScoredVertices(state, 3) // Top 3 as conclusions

	terminalIDs := make([]string, len(bestVertices))
	for i, v := range bestVertices {
		terminalIDs[i] = v.ID
	}

	if len(terminalIDs) > 0 {
		_ = gc.SetTerminalVertices(graphID, terminalIDs)
	}

	stepNum++
	result.ExplorationPath = append(result.ExplorationPath, ExplorationStep{
		Step:        stepNum,
		Action:      "finalize",
		VertexCount: len(bestVertices),
		Details:     fmt.Sprintf("Finalized with %d best conclusions", len(bestVertices)),
	})

	result.BestVertices = bestVertices
	result.Conclusions = bestVertices

	return result, nil
}

// getTopScoredVertices returns the n highest-scored vertices
func getTopScoredVertices(state *GraphState, n int) []*ThoughtVertex {
	if state == nil || len(state.Vertices) == 0 {
		return []*ThoughtVertex{}
	}

	// Collect all vertices with scores
	vertices := make([]*ThoughtVertex, 0, len(state.Vertices))
	for _, v := range state.Vertices {
		if v.Score > 0 {
			vertices = append(vertices, v)
		}
	}

	// Sort by score descending (simple bubble sort for small n)
	for i := 0; i < len(vertices); i++ {
		for j := i + 1; j < len(vertices); j++ {
			if vertices[j].Score > vertices[i].Score {
				vertices[i], vertices[j] = vertices[j], vertices[i]
			}
		}
	}

	// Return top n
	if n > len(vertices) {
		n = len(vertices)
	}
	return vertices[:n]
}

// truncateStr truncates a string to maxLen with ellipsis
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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

// fastScoreVertex applies local heuristics to score a vertex without LLM calls
// This provides reasonable scoring in ~1ms vs ~15-30s for LLM scoring
func fastScoreVertex(vertex *ThoughtVertex, problem string) float64 {
	if vertex == nil || vertex.Content == "" {
		return 0.0
	}

	content := strings.ToLower(vertex.Content)
	problemLower := strings.ToLower(problem)
	score := 0.5 // Base score

	// Length bonus: prefer substantive thoughts (100-500 chars optimal)
	length := len(vertex.Content)
	if length >= 100 && length <= 500 {
		score += 0.15
	} else if length >= 50 && length <= 800 {
		score += 0.08
	} else if length < 20 {
		score -= 0.2 // Penalize very short
	}

	// Relevance: keyword overlap with problem
	problemWords := strings.Fields(problemLower)
	matchCount := 0
	for _, word := range problemWords {
		if len(word) > 3 && strings.Contains(content, word) {
			matchCount++
		}
	}
	if len(problemWords) > 0 {
		relevance := float64(matchCount) / float64(len(problemWords))
		score += relevance * 0.2
	}

	// Structure indicators: lists, steps, examples
	structureIndicators := []string{
		"1.", "2.", "first", "second", "then", "next",
		"because", "therefore", "however", "example",
		"specifically", "consider", "approach",
	}
	structureCount := 0
	for _, indicator := range structureIndicators {
		if strings.Contains(content, indicator) {
			structureCount++
		}
	}
	score += float64(structureCount) * 0.03
	if score > 1.0 {
		score = 0.95
	}

	// Depth factor based on graph position
	if vertex.Depth > 0 {
		depthBonus := float64(vertex.Depth) * 0.05
		if depthBonus > 0.15 {
			depthBonus = 0.15
		}
		score += depthBonus
	}

	// Normalize to 0-1 range
	if score < 0 {
		score = 0.1
	}
	if score > 1 {
		score = 0.98
	}

	return score
}
