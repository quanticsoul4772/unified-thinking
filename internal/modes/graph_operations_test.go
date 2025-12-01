// Package modes - Graph operations tests
package modes

import (
	"context"
	"testing"

	"unified-thinking/internal/storage"
)

func TestGenerate(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)
	llm := NewMockLLMClient()

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
	llm := NewMockLLMClient()

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
	llm := NewMockLLMClient()

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
	llm := NewMockLLMClient()

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
	llm := NewMockLLMClient()

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
	llm := NewMockLLMClient()

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
	llm := NewMockLLMClient()

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
