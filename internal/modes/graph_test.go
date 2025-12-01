// Package modes - Graph-of-Thoughts tests
package modes

import (
	"testing"

	"unified-thinking/internal/storage"
)

func TestGraphController_Initialize(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, err := gc.Initialize("test-graph", "Initial thought", nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if state.ID != "test-graph" {
		t.Errorf("Expected ID 'test-graph', got '%s'", state.ID)
	}

	if len(state.Vertices) != 1 {
		t.Errorf("Expected 1 vertex, got %d", len(state.Vertices))
	}

	if len(state.RootIDs) != 1 {
		t.Errorf("Expected 1 root, got %d", len(state.RootIDs))
	}

	if len(state.ActiveIDs) != 1 {
		t.Errorf("Expected 1 active vertex, got %d", len(state.ActiveIDs))
	}

	rootID := state.RootIDs[0]
	rootVertex, ok := state.Vertices[rootID]
	if !ok {
		t.Fatal("Root vertex not found")
	}

	if rootVertex.Content != "Initial thought" {
		t.Errorf("Expected content 'Initial thought', got '%s'", rootVertex.Content)
	}

	if rootVertex.Type != ThoughtTypeInitial {
		t.Errorf("Expected type initial, got %s", rootVertex.Type)
	}

	if rootVertex.Depth != 0 {
		t.Errorf("Expected depth 0, got %d", rootVertex.Depth)
	}
}

func TestGraphController_AddVertex(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial thought", nil)

	newVertex := NewThoughtVertex("test-vertex", "New thought", ThoughtTypeGenerated, 0.7)
	err := gc.AddVertex(state.ID, newVertex)
	if err != nil {
		t.Fatalf("AddVertex failed: %v", err)
	}

	if len(state.Vertices) != 2 {
		t.Errorf("Expected 2 vertices, got %d", len(state.Vertices))
	}

	retrieved, ok := state.Vertices["test-vertex"]
	if !ok {
		t.Fatal("Added vertex not found")
	}

	if retrieved.Content != "New thought" {
		t.Errorf("Expected content 'New thought', got '%s'", retrieved.Content)
	}
}

func TestGraphController_AddEdge(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial thought", nil)
	rootID := state.RootIDs[0]

	childVertex := NewThoughtVertex("child-vertex", "Child thought", ThoughtTypeGenerated, 0.8)
	_ = gc.AddVertex(state.ID, childVertex)

	edge := NewThoughtEdge("edge-1", rootID, "child-vertex", EdgeTypeDerivesFrom, 0.9)
	err := gc.AddEdge(state.ID, edge)
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	if len(state.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(state.Edges))
	}

	// Check parent-child relationships
	rootVertex := state.Vertices[rootID]
	if len(rootVertex.ChildIDs) != 1 {
		t.Errorf("Expected 1 child, got %d", len(rootVertex.ChildIDs))
	}

	if rootVertex.ChildIDs[0] != "child-vertex" {
		t.Errorf("Expected child 'child-vertex', got '%s'", rootVertex.ChildIDs[0])
	}

	child := state.Vertices["child-vertex"]
	if len(child.ParentIDs) != 1 {
		t.Errorf("Expected 1 parent, got %d", len(child.ParentIDs))
	}

	if child.Depth != 1 {
		t.Errorf("Expected depth 1, got %d", child.Depth)
	}
}

func TestGraphController_MultipleParents(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial thought", nil)
	rootID := state.RootIDs[0]

	// Add two child vertices
	child1 := NewThoughtVertex("child-1", "Child 1", ThoughtTypeGenerated, 0.8)
	child2 := NewThoughtVertex("child-2", "Child 2", ThoughtTypeGenerated, 0.8)
	_ = gc.AddVertex(state.ID, child1)
	_ = gc.AddVertex(state.ID, child2)

	edge1 := NewThoughtEdge("edge-1", rootID, "child-1", EdgeTypeDerivesFrom, 0.9)
	edge2 := NewThoughtEdge("edge-2", rootID, "child-2", EdgeTypeDerivesFrom, 0.9)
	_ = gc.AddEdge(state.ID, edge1)
	_ = gc.AddEdge(state.ID, edge2)

	// Add aggregated vertex with two parents
	aggregated := NewThoughtVertex("aggregated", "Aggregated thought", ThoughtTypeAggregated, 0.9)
	_ = gc.AddVertex(state.ID, aggregated)

	edge3 := NewThoughtEdge("edge-3", "child-1", "aggregated", EdgeTypeAggregates, 0.8)
	edge4 := NewThoughtEdge("edge-4", "child-2", "aggregated", EdgeTypeAggregates, 0.8)
	_ = gc.AddEdge(state.ID, edge3)
	_ = gc.AddEdge(state.ID, edge4)

	// Verify multiple parents
	agg := state.Vertices["aggregated"]
	if len(agg.ParentIDs) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(agg.ParentIDs))
	}

	if agg.Depth != 2 {
		t.Errorf("Expected depth 2, got %d", agg.Depth)
	}
}

func TestGraphController_SetActiveVertices(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial thought", nil)

	v1 := NewThoughtVertex("v1", "Vertex 1", ThoughtTypeGenerated, 0.8)
	v2 := NewThoughtVertex("v2", "Vertex 2", ThoughtTypeGenerated, 0.8)
	_ = gc.AddVertex(state.ID, v1)
	_ = gc.AddVertex(state.ID, v2)

	err := gc.SetActiveVertices(state.ID, []string{"v1", "v2"})
	if err != nil {
		t.Fatalf("SetActiveVertices failed: %v", err)
	}

	if len(state.ActiveIDs) != 2 {
		t.Errorf("Expected 2 active vertices, got %d", len(state.ActiveIDs))
	}
}

func TestGraphController_MaxVertices(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	config := DefaultGraphConfig()
	config.MaxVertices = 3

	state, _ := gc.Initialize("test-graph", "Initial thought", config)

	v1 := NewThoughtVertex("v1", "Vertex 1", ThoughtTypeGenerated, 0.8)
	v2 := NewThoughtVertex("v2", "Vertex 2", ThoughtTypeGenerated, 0.8)

	_ = gc.AddVertex(state.ID, v1)
	_ = gc.AddVertex(state.ID, v2)

	v3 := NewThoughtVertex("v3", "Vertex 3", ThoughtTypeGenerated, 0.8)
	err := gc.AddVertex(state.ID, v3)
	if err == nil {
		t.Error("Expected error when exceeding max vertices")
	}
}

func TestGraphController_RemoveVertex(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial thought", nil)
	rootID := state.RootIDs[0]

	child := NewThoughtVertex("child", "Child thought", ThoughtTypeGenerated, 0.8)
	_ = gc.AddVertex(state.ID, child)

	edge := NewThoughtEdge("edge-1", rootID, "child", EdgeTypeDerivesFrom, 0.9)
	_ = gc.AddEdge(state.ID, edge)

	err := gc.RemoveVertex(state.ID, "child")
	if err != nil {
		t.Fatalf("RemoveVertex failed: %v", err)
	}

	if len(state.Vertices) != 1 {
		t.Errorf("Expected 1 vertex after removal, got %d", len(state.Vertices))
	}

	if len(state.Edges) != 0 {
		t.Errorf("Expected 0 edges after removal, got %d", len(state.Edges))
	}

	rootVertex := state.Vertices[rootID]
	if len(rootVertex.ChildIDs) != 0 {
		t.Errorf("Expected 0 children after removal, got %d", len(rootVertex.ChildIDs))
	}
}

func TestGraphController_GetVerticesByDepth(t *testing.T) {
	store := storage.NewMemoryStorage()
	gc := NewGraphController(store)

	state, _ := gc.Initialize("test-graph", "Initial thought", nil)
	rootID := state.RootIDs[0]

	// Add depth 1 vertices
	v1 := NewThoughtVertex("v1", "Vertex 1", ThoughtTypeGenerated, 0.8)
	v2 := NewThoughtVertex("v2", "Vertex 2", ThoughtTypeGenerated, 0.8)
	_ = gc.AddVertex(state.ID, v1)
	_ = gc.AddVertex(state.ID, v2)

	edge1 := NewThoughtEdge("e1", rootID, "v1", EdgeTypeDerivesFrom, 0.9)
	edge2 := NewThoughtEdge("e2", rootID, "v2", EdgeTypeDerivesFrom, 0.9)
	_ = gc.AddEdge(state.ID, edge1)
	_ = gc.AddEdge(state.ID, edge2)

	depth0, _ := gc.GetVerticesByDepth(state.ID, 0)
	if len(depth0) != 1 {
		t.Errorf("Expected 1 vertex at depth 0, got %d", len(depth0))
	}

	depth1, _ := gc.GetVerticesByDepth(state.ID, 1)
	if len(depth1) != 2 {
		t.Errorf("Expected 2 vertices at depth 1, got %d", len(depth1))
	}
}

func TestDefaultGraphConfig(t *testing.T) {
	config := DefaultGraphConfig()

	if config.MaxVertices != 50 {
		t.Errorf("Expected MaxVertices 50, got %d", config.MaxVertices)
	}

	if config.MaxActiveVertices != 10 {
		t.Errorf("Expected MaxActiveVertices 10, got %d", config.MaxActiveVertices)
	}

	if config.MaxDepth != 7 {
		t.Errorf("Expected MaxDepth 7, got %d", config.MaxDepth)
	}

	if config.MaxRefinements != 3 {
		t.Errorf("Expected MaxRefinements 3, got %d", config.MaxRefinements)
	}

	if config.PruneThreshold != 0.3 {
		t.Errorf("Expected PruneThreshold 0.3, got %f", config.PruneThreshold)
	}

	if config.AggregateMinPaths != 2 {
		t.Errorf("Expected AggregateMinPaths 2, got %d", config.AggregateMinPaths)
	}
}

func TestNewThoughtVertex(t *testing.T) {
	vertex := NewThoughtVertex("test-id", "Test content", ThoughtTypeGenerated, 0.75)

	if vertex.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", vertex.ID)
	}

	if vertex.Content != "Test content" {
		t.Errorf("Expected content 'Test content', got '%s'", vertex.Content)
	}

	if vertex.Type != ThoughtTypeGenerated {
		t.Errorf("Expected type generated, got %s", vertex.Type)
	}

	if vertex.Confidence != 0.75 {
		t.Errorf("Expected confidence 0.75, got %f", vertex.Confidence)
	}

	if vertex.ParentIDs == nil || len(vertex.ParentIDs) != 0 {
		t.Error("Expected empty ParentIDs slice")
	}

	if vertex.Metadata == nil {
		t.Error("Expected non-nil Metadata map")
	}
}

func TestNewThoughtEdge(t *testing.T) {
	edge := NewThoughtEdge("edge-1", "from", "to", EdgeTypeDerivesFrom, 0.8)

	if edge.ID != "edge-1" {
		t.Errorf("Expected ID 'edge-1', got '%s'", edge.ID)
	}

	if edge.FromID != "from" {
		t.Errorf("Expected FromID 'from', got '%s'", edge.FromID)
	}

	if edge.ToID != "to" {
		t.Errorf("Expected ToID 'to', got '%s'", edge.ToID)
	}

	if edge.Type != EdgeTypeDerivesFrom {
		t.Errorf("Expected type derives_from, got %s", edge.Type)
	}

	if edge.Weight != 0.8 {
		t.Errorf("Expected weight 0.8, got %f", edge.Weight)
	}
}
