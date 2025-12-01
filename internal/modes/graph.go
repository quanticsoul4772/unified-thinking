// Package modes - Graph-of-Thoughts controller
package modes

import (
	"fmt"
	"time"

	"github.com/dominikbraun/graph"
	"unified-thinking/internal/storage"
)

// GraphController manages Graph-of-Thoughts reasoning
type GraphController struct {
	storage storage.Storage
	states  map[string]*GraphState // Active graph states
}

// NewGraphController creates a new graph controller
func NewGraphController(store storage.Storage) *GraphController {
	return &GraphController{
		storage: store,
		states:  make(map[string]*GraphState),
	}
}

// Initialize creates a new graph with an initial thought
func (gc *GraphController) Initialize(id, initialContent string, config *GraphConfig) (*GraphState, error) {
	if config == nil {
		config = DefaultGraphConfig()
	}

	// Create directed graph
	g := graph.New(VertexHash, graph.Directed())

	// Create initial vertex
	initialVertex := NewThoughtVertex(
		fmt.Sprintf("%s-vertex-0", id),
		initialContent,
		ThoughtTypeInitial,
		0.8, // Default confidence for initial thoughts
	)
	initialVertex.Depth = 0

	// Add vertex to graph
	if err := g.AddVertex(initialVertex); err != nil {
		return nil, fmt.Errorf("failed to add initial vertex: %w", err)
	}

	// Create graph state
	state := &GraphState{
		ID:          id,
		Graph:       g,
		Vertices:    map[string]*ThoughtVertex{initialVertex.ID: initialVertex},
		Edges:       make(map[string]*ThoughtEdge),
		RootIDs:     []string{initialVertex.ID},
		ActiveIDs:   []string{initialVertex.ID},
		TerminalIDs: []string{},
		Config:      config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store state
	gc.states[id] = state

	return state, nil
}

// GetState retrieves a graph state
func (gc *GraphController) GetState(id string) (*GraphState, error) {
	state, ok := gc.states[id]
	if !ok {
		return nil, fmt.Errorf("graph state not found: %s", id)
	}
	return state, nil
}

// AddVertex adds a new thought vertex to the graph
func (gc *GraphController) AddVertex(stateID string, vertex *ThoughtVertex) error {
	state, err := gc.GetState(stateID)
	if err != nil {
		return err
	}

	// Check vertex limit
	if len(state.Vertices) >= state.Config.MaxVertices {
		return fmt.Errorf("max vertices reached (%d)", state.Config.MaxVertices)
	}

	// Add to graph
	if err := state.Graph.AddVertex(vertex); err != nil {
		return fmt.Errorf("failed to add vertex: %w", err)
	}

	// Update state
	state.Vertices[vertex.ID] = vertex
	state.UpdatedAt = time.Now()

	return nil
}

// AddEdge adds a relationship between thoughts
func (gc *GraphController) AddEdge(stateID string, edge *ThoughtEdge) error {
	state, err := gc.GetState(stateID)
	if err != nil {
		return err
	}

	// Verify vertices exist
	if _, ok := state.Vertices[edge.FromID]; !ok {
		return fmt.Errorf("source vertex not found: %s", edge.FromID)
	}
	if _, ok := state.Vertices[edge.ToID]; !ok {
		return fmt.Errorf("target vertex not found: %s", edge.ToID)
	}

	// Add to graph (EdgeWeight expects int, we'll store weight in our edge struct)
	if err := state.Graph.AddEdge(edge.FromID, edge.ToID); err != nil {
		return fmt.Errorf("failed to add edge: %w", err)
	}

	// Update vertex relationships
	fromVertex := state.Vertices[edge.FromID]
	toVertex := state.Vertices[edge.ToID]

	if !containsString(fromVertex.ChildIDs, edge.ToID) {
		fromVertex.ChildIDs = append(fromVertex.ChildIDs, edge.ToID)
	}
	if !containsString(toVertex.ParentIDs, edge.FromID) {
		toVertex.ParentIDs = append(toVertex.ParentIDs, edge.FromID)
	}

	// Update depth of target vertex
	if toVertex.Depth == 0 || toVertex.Depth > fromVertex.Depth+1 {
		toVertex.Depth = fromVertex.Depth + 1
	}

	// Update state
	state.Edges[edge.ID] = edge
	state.UpdatedAt = time.Now()

	return nil
}

// SetActiveVertices updates the active frontier
func (gc *GraphController) SetActiveVertices(stateID string, vertexIDs []string) error {
	state, err := gc.GetState(stateID)
	if err != nil {
		return err
	}

	// Verify all vertices exist
	for _, id := range vertexIDs {
		if _, ok := state.Vertices[id]; !ok {
			return fmt.Errorf("vertex not found: %s", id)
		}
	}

	// Check active limit
	if len(vertexIDs) > state.Config.MaxActiveVertices {
		return fmt.Errorf("too many active vertices (limit: %d)", state.Config.MaxActiveVertices)
	}

	state.ActiveIDs = vertexIDs
	state.UpdatedAt = time.Now()

	return nil
}

// SetTerminalVertices marks final conclusions
func (gc *GraphController) SetTerminalVertices(stateID string, vertexIDs []string) error {
	state, err := gc.GetState(stateID)
	if err != nil {
		return err
	}

	// Verify all vertices exist
	for _, id := range vertexIDs {
		if _, ok := state.Vertices[id]; !ok {
			return fmt.Errorf("vertex not found: %s", id)
		}
	}

	state.TerminalIDs = vertexIDs
	state.UpdatedAt = time.Now()

	return nil
}

// RemoveVertex removes a vertex and its edges
func (gc *GraphController) RemoveVertex(stateID, vertexID string) error {
	state, err := gc.GetState(stateID)
	if err != nil {
		return err
	}

	vertex, ok := state.Vertices[vertexID]
	if !ok {
		return fmt.Errorf("vertex not found: %s", vertexID)
	}

	// Remove edges from graph first (dominikbraun/graph requires this)
	edgesToRemove := []string{}
	for edgeID, edge := range state.Edges {
		if edge.FromID == vertexID || edge.ToID == vertexID {
			edgesToRemove = append(edgesToRemove, edgeID)
			// Remove edge from graph
			_ = state.Graph.RemoveEdge(edge.FromID, edge.ToID)
		}
	}
	for _, edgeID := range edgesToRemove {
		delete(state.Edges, edgeID)
	}

	// Now remove from graph
	if err := state.Graph.RemoveVertex(vertexID); err != nil {
		return fmt.Errorf("failed to remove vertex: %w", err)
	}

	// Update parent/child references
	for _, parentID := range vertex.ParentIDs {
		if parent, ok := state.Vertices[parentID]; ok {
			parent.ChildIDs = removeFromSlice(parent.ChildIDs, vertexID)
		}
	}
	for _, childID := range vertex.ChildIDs {
		if child, ok := state.Vertices[childID]; ok {
			child.ParentIDs = removeFromSlice(child.ParentIDs, vertexID)
		}
	}

	// Remove from state
	delete(state.Vertices, vertexID)

	// Remove from active/terminal lists
	state.ActiveIDs = removeFromSlice(state.ActiveIDs, vertexID)
	state.TerminalIDs = removeFromSlice(state.TerminalIDs, vertexID)

	state.UpdatedAt = time.Now()

	return nil
}

// GetVerticesByDepth returns all vertices at a given depth
func (gc *GraphController) GetVerticesByDepth(stateID string, depth int) ([]*ThoughtVertex, error) {
	state, err := gc.GetState(stateID)
	if err != nil {
		return nil, err
	}

	vertices := []*ThoughtVertex{}
	for _, v := range state.Vertices {
		if v.Depth == depth {
			vertices = append(vertices, v)
		}
	}

	return vertices, nil
}

// GetChildVertices returns all children of a vertex
func (gc *GraphController) GetChildVertices(stateID, vertexID string) ([]*ThoughtVertex, error) {
	state, err := gc.GetState(stateID)
	if err != nil {
		return nil, err
	}

	vertex, ok := state.Vertices[vertexID]
	if !ok {
		return nil, fmt.Errorf("vertex not found: %s", vertexID)
	}

	children := make([]*ThoughtVertex, 0, len(vertex.ChildIDs))
	for _, childID := range vertex.ChildIDs {
		if child, ok := state.Vertices[childID]; ok {
			children = append(children, child)
		}
	}

	return children, nil
}

// Helper functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
