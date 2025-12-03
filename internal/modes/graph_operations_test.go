// Package modes - Graph operations tests
package modes

import (
	"context"
	"fmt"
	"testing"

	"unified-thinking/internal/storage"
)

// testLLMClient provides deterministic responses for testing
type testLLMClient struct{}

func (t *testLLMClient) Generate(ctx context.Context, prompt string, k int) ([]string, error) {
	results := make([]string, k)
	for i := 0; i < k; i++ {
		results[i] = fmt.Sprintf("Test continuation %d from: %s", i+1, truncateTest(prompt, 30))
	}
	return results, nil
}

func (t *testLLMClient) Aggregate(ctx context.Context, thoughts []string, problem string) (string, error) {
	return fmt.Sprintf("Aggregated %d thoughts", len(thoughts)), nil
}

func (t *testLLMClient) Refine(ctx context.Context, thought string, problem string, refinementCount int) (string, error) {
	return fmt.Sprintf("Refined v%d: %s", refinementCount+1, thought), nil
}

func (t *testLLMClient) Score(ctx context.Context, thought string, problem string, criteria map[string]float64) (float64, map[string]float64, error) {
	breakdown := map[string]float64{
		"confidence":   0.8,
		"validity":     0.9,
		"relevance":    0.7,
		"novelty":      0.6,
		"depth_factor": 0.8,
	}
	overall := 0.76
	return overall, breakdown, nil
}

func (t *testLLMClient) ExtractKeyPoints(ctx context.Context, thought string) ([]string, error) {
	return []string{"Test key point 1", "Test key point 2"}, nil
}

func (t *testLLMClient) CalculateNovelty(ctx context.Context, thought string, siblings []string) (float64, error) {
	if len(siblings) == 0 {
		return 1.0, nil
	}
	return 0.7, nil
}

func truncateTest(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func TestGenerate(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	state, _ := gc.Initialize("test-graph", "Initial problem", nil)

	req := GenerateRequest{
		K:       3,
		Problem: "Initial problem",
	}

	vertices, err := gc.Generate(context.Background(), state.ID, llm, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(vertices) != 3 {
		t.Errorf("Expected 3 vertices, got %d", len(vertices))
	}

	for _, v := range vertices {
		if v.Type != ThoughtTypeGenerated {
			t.Errorf("Expected type generated, got %s", v.Type)
		}
		if v.Depth != 1 {
			t.Errorf("Expected depth 1, got %d", v.Depth)
		}
	}
}

func TestGenerate_MultipleRounds(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	state, _ := gc.Initialize("test-graph", "Initial problem", nil)

	// Round 1
	req1 := GenerateRequest{K: 2, Problem: "Initial problem"}
	round1, _ := gc.Generate(context.Background(), state.ID, llm, req1)

	if len(round1) != 2 {
		t.Errorf("Round 1: Expected 2 vertices, got %d", len(round1))
	}

	// Round 2 - should generate from round1 vertices
	req2 := GenerateRequest{K: 2, Problem: "Initial problem"}
	round2, _ := gc.Generate(context.Background(), state.ID, llm, req2)

	if len(round2) != 4 {
		t.Errorf("Round 2: Expected 4 vertices (2 sources × 2), got %d", len(round2))
	}

	for _, v := range round2 {
		if v.Depth != 2 {
			t.Errorf("Expected depth 2, got %d", v.Depth)
		}
	}
}

func TestAggregate(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	state, _ := gc.Initialize("test-graph", "Initial problem", nil)

	// Generate some vertices to aggregate
	req := GenerateRequest{K: 3, Problem: "Initial problem"}
	vertices, _ := gc.Generate(context.Background(), state.ID, llm, req)

	vertexIDs := make([]string, len(vertices))
	for i, v := range vertices {
		vertexIDs[i] = v.ID
	}

	// Aggregate
	aggReq := AggregateRequest{
		VertexIDs: vertexIDs,
		Problem:   "Initial problem",
	}

	aggregated, err := gc.Aggregate(context.Background(), state.ID, llm, aggReq)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if aggregated.Type != ThoughtTypeAggregated {
		t.Errorf("Expected type aggregated, got %s", aggregated.Type)
	}

	if len(aggregated.ParentIDs) != 3 {
		t.Errorf("Expected 3 parents, got %d", len(aggregated.ParentIDs))
	}

	if aggregated.Depth != 2 {
		t.Errorf("Expected depth 2 (max parent depth + 1), got %d", aggregated.Depth)
	}
}

func TestAggregate_MinPaths(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	config := DefaultGraphConfig()
	config.AggregateMinPaths = 2

	state, _ := gc.Initialize("test-graph", "Initial problem", config)

	v1 := NewThoughtVertex("v1", "Vertex 1", ThoughtTypeGenerated, 0.8)
	_ = gc.AddVertex(state.ID, v1)

	// Try to aggregate with only 1 vertex (should fail)
	aggReq := AggregateRequest{
		VertexIDs: []string{"v1"},
		Problem:   "Initial problem",
	}

	_, err := gc.Aggregate(context.Background(), state.ID, llm, aggReq)
	if err == nil {
		t.Error("Expected error when aggregating with < AggregateMinPaths vertices")
	}
}

func TestRefine(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	state, _ := gc.Initialize("test-graph", "Initial problem", nil)
	rootID := state.RootIDs[0]

	refReq := RefineRequest{
		VertexID: rootID,
		Problem:  "Initial problem",
	}

	refined, err := gc.Refine(context.Background(), state.ID, llm, refReq)
	if err != nil {
		t.Fatalf("Refine failed: %v", err)
	}

	if refined.Type != ThoughtTypeRefined {
		t.Errorf("Expected type refined, got %s", refined.Type)
	}

	if refined.RefinedCount != 1 {
		t.Errorf("Expected refined_count 1, got %d", refined.RefinedCount)
	}

	if len(refined.ParentIDs) != 1 {
		t.Errorf("Expected 1 parent, got %d", len(refined.ParentIDs))
	}

	if refined.ParentIDs[0] != rootID {
		t.Errorf("Expected parent to be root, got %s", refined.ParentIDs[0])
	}
}

func TestRefine_MaxRefinements(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	config := DefaultGraphConfig()
	config.MaxRefinements = 2

	state, _ := gc.Initialize("test-graph", "Initial problem", config)
	rootID := state.RootIDs[0]

	// Refine 1
	refReq1 := RefineRequest{VertexID: rootID, Problem: "Initial problem"}
	refined1, _ := gc.Refine(context.Background(), state.ID, llm, refReq1)

	// Refine 2
	refReq2 := RefineRequest{VertexID: refined1.ID, Problem: "Initial problem"}
	refined2, _ := gc.Refine(context.Background(), state.ID, llm, refReq2)

	// Refine 3 (should fail)
	refReq3 := RefineRequest{VertexID: refined2.ID, Problem: "Initial problem"}
	_, err := gc.Refine(context.Background(), state.ID, llm, refReq3)
	if err == nil {
		t.Error("Expected error when exceeding MaxRefinements")
	}
}

func TestScore(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	state, _ := gc.Initialize("test-graph", "Initial problem", nil)
	rootID := state.RootIDs[0]

	scoreReq := ScoreRequest{
		VertexID: rootID,
		Problem:  "Initial problem",
	}

	breakdown, err := gc.Score(context.Background(), state.ID, llm, scoreReq)
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	if breakdown.Overall < 0 || breakdown.Overall > 1 {
		t.Errorf("Overall score should be 0-1, got %f", breakdown.Overall)
	}

	if breakdown.Confidence < 0 || breakdown.Confidence > 1 {
		t.Errorf("Confidence should be 0-1, got %f", breakdown.Confidence)
	}

	// Verify vertex score was updated
	vertex := state.Vertices[rootID]
	if vertex.Score != breakdown.Overall {
		t.Errorf("Vertex score not updated: expected %f, got %f", breakdown.Overall, vertex.Score)
	}
}

func TestPrune(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial problem", nil)

	// Create vertices with different scores
	v1 := NewThoughtVertex("v1", "High quality", ThoughtTypeGenerated, 0.9)
	v1.Score = 0.8
	v2 := NewThoughtVertex("v2", "Low quality", ThoughtTypeGenerated, 0.5)
	v2.Score = 0.2
	v3 := NewThoughtVertex("v3", "Medium quality", ThoughtTypeGenerated, 0.7)
	v3.Score = 0.5

	_ = gc.AddVertex(state.ID, v1)
	_ = gc.AddVertex(state.ID, v2)
	_ = gc.AddVertex(state.ID, v3)

	// Prune with threshold 0.3
	removed, err := gc.Prune(context.Background(), state.ID, 0.3)
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}

	if removed != 1 {
		t.Errorf("Expected 1 vertex pruned, got %d", removed)
	}

	// v2 should be removed
	if _, ok := state.Vertices["v2"]; ok {
		t.Error("Low quality vertex should have been pruned")
	}

	// v1 and v3 should remain
	if _, ok := state.Vertices["v1"]; !ok {
		t.Error("High quality vertex should not be pruned")
	}
	if _, ok := state.Vertices["v3"]; !ok {
		t.Error("Medium quality vertex should not be pruned")
	}
}

func TestPrune_PreservesRootsAndTerminals(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial problem", nil)
	rootID := state.RootIDs[0]

	// Set low score on root
	state.Vertices[rootID].Score = 0.1

	// Prune
	removed, _ := gc.Prune(context.Background(), state.ID, 0.5)

	if removed != 0 {
		t.Errorf("Expected 0 vertices pruned (root protected), got %d", removed)
	}

	if _, ok := state.Vertices[rootID]; !ok {
		t.Error("Root vertex should never be pruned")
	}
}

// Tests for got-explore auto-orchestrated tool (Phase 2.2)

func TestExplore_BasicWorkflow(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	req := ExploreRequest{
		InitialThought: "Debug flaky CI tests",
		Problem:        "Find root cause of intermittent CI failures",
	}

	result, err := gc.Explore(context.Background(), "test-explore", llm, req)
	if err != nil {
		t.Fatalf("Explore failed: %v", err)
	}

	// Verify basic structure
	if result.GraphID != "test-explore" {
		t.Errorf("Expected graph_id 'test-explore', got '%s'", result.GraphID)
	}

	if result.Problem != req.Problem {
		t.Errorf("Expected problem '%s', got '%s'", req.Problem, result.Problem)
	}

	// Should have completed at least 1 iteration
	if result.Iterations < 1 {
		t.Errorf("Expected at least 1 iteration, got %d", result.Iterations)
	}

	// Should have generated some vertices
	if result.TotalGenerated == 0 {
		t.Error("Expected some vertices to be generated")
	}

	// Should have an exploration path
	if len(result.ExplorationPath) == 0 {
		t.Error("Expected exploration path to have steps")
	}

	// First step should be initialize
	if result.ExplorationPath[0].Action != "initialize" {
		t.Errorf("Expected first action to be 'initialize', got '%s'", result.ExplorationPath[0].Action)
	}

	// Last step should be finalize
	lastStep := result.ExplorationPath[len(result.ExplorationPath)-1]
	if lastStep.Action != "finalize" {
		t.Errorf("Expected last action to be 'finalize', got '%s'", lastStep.Action)
	}
}

func TestExplore_WithConfig(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	config := &ExploreConfig{
		K:              2,    // Generate 2 continuations
		MaxIterations:  1,    // Only 1 iteration
		PruneThreshold: 0.5,  // Higher threshold
		RefineTopN:     2,    // Refine top 2
		ScoreAll:       true, // Score all vertices
	}

	req := ExploreRequest{
		InitialThought: "Optimize database queries",
		Problem:        "Improve query performance",
		Config:         config,
	}

	result, err := gc.Explore(context.Background(), "test-explore-config", llm, req)
	if err != nil {
		t.Fatalf("Explore with config failed: %v", err)
	}

	// Should respect MaxIterations
	if result.Iterations != 1 {
		t.Errorf("Expected 1 iteration, got %d", result.Iterations)
	}

	// Should generate k vertices per source
	// First iteration: 1 source × 2 = 2 vertices
	if result.TotalGenerated != 2 {
		t.Errorf("Expected 2 generated vertices (1 source × k=2), got %d", result.TotalGenerated)
	}
}

func TestExplore_ValidationErrors(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	tests := []struct {
		name    string
		graphID string
		req     ExploreRequest
	}{
		{
			name:    "missing graph_id",
			graphID: "",
			req:     ExploreRequest{InitialThought: "test", Problem: "test"},
		},
		{
			name:    "missing initial_thought",
			graphID: "test",
			req:     ExploreRequest{Problem: "test"},
		},
		{
			name:    "missing problem",
			graphID: "test",
			req:     ExploreRequest{InitialThought: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gc.Explore(context.Background(), tt.graphID, llm, tt.req)
			if err == nil {
				t.Errorf("Expected error for %s", tt.name)
			}
		})
	}
}

func TestExplore_ExplorationPathActions(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := &testLLMClient{}

	req := ExploreRequest{
		InitialThought: "Analyze system architecture",
		Problem:        "Design microservices",
		Config: &ExploreConfig{
			K:             2,
			MaxIterations: 1,
		},
	}

	result, err := gc.Explore(context.Background(), "test-actions", llm, req)
	if err != nil {
		t.Fatalf("Explore failed: %v", err)
	}

	// Track which actions we've seen
	actionsSeen := make(map[string]bool)
	for _, step := range result.ExplorationPath {
		actionsSeen[step.Action] = true
	}

	// Should have these actions
	requiredActions := []string{"initialize", "generate", "score", "finalize"}
	for _, action := range requiredActions {
		if !actionsSeen[action] {
			t.Errorf("Expected action '%s' in exploration path", action)
		}
	}
}

func TestGetTopScoredVertices(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial", nil)

	// Add vertices with different scores
	vertices := []struct {
		id    string
		score float64
	}{
		{"v1", 0.5},
		{"v2", 0.9},
		{"v3", 0.7},
		{"v4", 0.3},
		{"v5", 0.8},
	}

	for _, v := range vertices {
		vertex := NewThoughtVertex(v.id, "Content", ThoughtTypeGenerated, 0.8)
		vertex.Score = v.score
		_ = gc.AddVertex(state.ID, vertex)
	}

	// Get top 3
	top3 := getTopScoredVertices(state, 3)
	if len(top3) != 3 {
		t.Fatalf("Expected 3 vertices, got %d", len(top3))
	}

	// Should be sorted by score descending
	expectedOrder := []string{"v2", "v5", "v3"}
	for i, v := range top3 {
		if v.ID != expectedOrder[i] {
			t.Errorf("Position %d: expected '%s', got '%s'", i, expectedOrder[i], v.ID)
		}
	}

	// Get top 10 (more than available)
	topAll := getTopScoredVertices(state, 10)
	if len(topAll) != 5 {
		t.Errorf("Expected 5 vertices (all with scores), got %d", len(topAll))
	}

	// Get top 0
	topNone := getTopScoredVertices(state, 0)
	if len(topNone) != 0 {
		t.Errorf("Expected 0 vertices, got %d", len(topNone))
	}
}

func TestGetTopScoredVertices_NilState(t *testing.T) {
	result := getTopScoredVertices(nil, 3)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for nil state, got %d", len(result))
	}
}

func TestGetTopScoredVertices_NoScores(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial", nil)

	// Add vertex without score
	v := NewThoughtVertex("v1", "Content", ThoughtTypeGenerated, 0.8)
	v.Score = 0 // No score
	_ = gc.AddVertex(state.ID, v)

	result := getTopScoredVertices(state, 3)
	if len(result) != 0 {
		t.Errorf("Expected 0 vertices (no scores > 0), got %d", len(result))
	}
}

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10x", 10, "exactly10x"},
		{"this is a longer string", 10, "this is a ..."},
		{"", 5, ""},
		{"abc", 0, "..."},
	}

	for _, tt := range tests {
		result := truncateStr(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateStr(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestDefaultExploreConfig(t *testing.T) {
	config := DefaultExploreConfig()

	if config.K != 3 {
		t.Errorf("Expected K=3, got %d", config.K)
	}
	// Default MaxIterations is 1 for speed (reduced from 2)
	if config.MaxIterations != 1 {
		t.Errorf("Expected MaxIterations=1, got %d", config.MaxIterations)
	}
	if config.PruneThreshold != 0.3 {
		t.Errorf("Expected PruneThreshold=0.3, got %f", config.PruneThreshold)
	}
	if config.RefineTopN != 1 {
		t.Errorf("Expected RefineTopN=1, got %d", config.RefineTopN)
	}
	if config.ScoreAll != false {
		t.Error("Expected ScoreAll=false")
	}
	// New performance options
	if config.UseFastScoring != true {
		t.Error("Expected UseFastScoring=true by default for speed")
	}
	if config.ParallelScoring != true {
		t.Error("Expected ParallelScoring=true by default")
	}
}

func TestThoroughExploreConfig(t *testing.T) {
	config := ThoroughExploreConfig()

	if config.K != 3 {
		t.Errorf("Expected K=3, got %d", config.K)
	}
	if config.MaxIterations != 2 {
		t.Errorf("Expected MaxIterations=2 for thorough mode, got %d", config.MaxIterations)
	}
	if config.RefineTopN != 2 {
		t.Errorf("Expected RefineTopN=2 for thorough mode, got %d", config.RefineTopN)
	}
	if config.UseFastScoring != false {
		t.Error("Expected UseFastScoring=false for thorough mode (use LLM)")
	}
	if config.ParallelScoring != true {
		t.Error("Expected ParallelScoring=true even in thorough mode")
	}
}

func TestFastScoreVertex(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		problem   string
		depth     int
		expectMin float64
		expectMax float64
	}{
		{
			name:      "nil vertex",
			content:   "",
			problem:   "test problem",
			depth:     0,
			expectMin: 0.0,
			expectMax: 0.0,
		},
		{
			name:      "empty content",
			content:   "",
			problem:   "test problem",
			depth:     0,
			expectMin: 0.0,
			expectMax: 0.0,
		},
		{
			name:      "short content penalized",
			content:   "short",
			problem:   "test problem",
			depth:     0,
			expectMin: 0.1,
			expectMax: 0.5,
		},
		{
			name:      "optimal length content",
			content:   "This is a substantive thought about the problem at hand. It contains enough detail to be useful and addresses the key concerns in a structured manner.",
			problem:   "test problem",
			depth:     0,
			expectMin: 0.5,
			expectMax: 0.9,
		},
		{
			name:      "relevant keywords boost score",
			content:   "This approach addresses the database performance issue by optimizing queries.",
			problem:   "database performance optimization",
			depth:     0,
			expectMin: 0.5,
			expectMax: 0.9,
		},
		{
			name:      "structure indicators boost score",
			content:   "First, we should analyze the problem. Second, we need to consider alternatives. Therefore, the best approach is to proceed step by step.",
			problem:   "any problem",
			depth:     0,
			expectMin: 0.6,
			expectMax: 1.0,
		},
		{
			name:      "depth adds bonus",
			content:   "A reasonably detailed thought that builds on previous analysis",
			problem:   "test problem",
			depth:     3,
			expectMin: 0.5,
			expectMax: 0.95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var vertex *ThoughtVertex
			if tt.content != "" {
				vertex = &ThoughtVertex{
					Content: tt.content,
					Depth:   tt.depth,
				}
			}

			score := fastScoreVertex(vertex, tt.problem)

			if score < tt.expectMin || score > tt.expectMax {
				t.Errorf("fastScoreVertex() = %v, want between %v and %v", score, tt.expectMin, tt.expectMax)
			}
		})
	}
}
