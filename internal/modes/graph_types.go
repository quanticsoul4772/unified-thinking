// Package modes - Graph-of-Thoughts data structures
package modes

import (
	"time"

	"github.com/dominikbraun/graph"
)

// ThoughtType categorizes vertices in the graph
type ThoughtType string

const (
	ThoughtTypeInitial    ThoughtType = "initial"    // Starting thoughts
	ThoughtTypeGenerated  ThoughtType = "generated"  // Derived thoughts
	ThoughtTypeAggregated ThoughtType = "aggregated" // Merged thoughts
	ThoughtTypeRefined    ThoughtType = "refined"    // Improved thoughts
)

// EdgeType categorizes relationships between thoughts
type EdgeType string

const (
	EdgeTypeDerivesFrom  EdgeType = "derives_from"  // Parent-child derivation
	EdgeTypeAggregates   EdgeType = "aggregates"    // Merges multiple thoughts
	EdgeTypeRefines      EdgeType = "refines"       // Iterative improvement
	EdgeTypeContradicts  EdgeType = "contradicts"   // Conflicting thoughts
	EdgeTypeSupports     EdgeType = "supports"      // Supporting evidence
)

// ThoughtVertex represents a thought node in the graph
type ThoughtVertex struct {
	ID           string                 `json:"id"`
	Content      string                 `json:"content"`
	Type         ThoughtType            `json:"type"`
	Confidence   float64                `json:"confidence"`   // 0.0-1.0
	Score        float64                `json:"score"`        // Quality score 0.0-1.0
	Depth        int                    `json:"depth"`        // Distance from root
	ParentIDs    []string               `json:"parent_ids"`   // Multiple parents allowed
	ChildIDs     []string               `json:"child_ids"`    // Children
	KeyPoints    []string               `json:"key_points"`   // Extracted insights
	RefinedCount int                    `json:"refined_count"` // Refinement iterations
	CreatedAt    time.Time              `json:"created_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ThoughtEdge represents a relationship between thoughts
type ThoughtEdge struct {
	ID        string    `json:"id"`
	FromID    string    `json:"from_id"`
	ToID      string    `json:"to_id"`
	Type      EdgeType  `json:"type"`
	Weight    float64   `json:"weight"`    // Relationship strength 0.0-1.0
	CreatedAt time.Time `json:"created_at"`
}

// GraphState encapsulates the entire thought graph
type GraphState struct {
	ID          string                            `json:"id"`
	Graph       graph.Graph[string, *ThoughtVertex] `json:"-"` // Don't serialize graph directly
	Vertices    map[string]*ThoughtVertex         `json:"vertices"`
	Edges       map[string]*ThoughtEdge           `json:"edges"`
	RootIDs     []string                          `json:"root_ids"`     // Initial thoughts
	ActiveIDs   []string                          `json:"active_ids"`   // Current frontier
	TerminalIDs []string                          `json:"terminal_ids"` // Final conclusions
	Config      *GraphConfig                      `json:"config"`
	CreatedAt   time.Time                         `json:"created_at"`
	UpdatedAt   time.Time                         `json:"updated_at"`
}

// GraphConfig controls graph behavior and limits
type GraphConfig struct {
	MaxVertices       int     `json:"max_vertices"`        // 50 - total graph limit
	MaxActiveVertices int     `json:"max_active_vertices"` // 10 - concurrent exploration
	MaxDepth          int     `json:"max_depth"`           // 7 - reasoning depth limit
	MaxRefinements    int     `json:"max_refinements"`     // 3 - self-improvement iterations
	PruneThreshold    float64 `json:"prune_threshold"`     // 0.3 - minimum quality score
	AggregateMinPaths int     `json:"aggregate_min_paths"` // 2 - minimum paths to aggregate
}

// DefaultGraphConfig returns sensible defaults
func DefaultGraphConfig() *GraphConfig {
	return &GraphConfig{
		MaxVertices:       50,
		MaxActiveVertices: 10,
		MaxDepth:          7,
		MaxRefinements:    3,
		PruneThreshold:    0.3,
		AggregateMinPaths: 2,
	}
}

// ScoreBreakdown provides detailed quality metrics
type ScoreBreakdown struct {
	Confidence  float64 `json:"confidence"`   // 25% weight - LLM self-assessment
	Validity    float64 `json:"validity"`     // 30% weight - logical consistency
	Relevance   float64 `json:"relevance"`    // 25% weight - semantic similarity to problem
	Novelty     float64 `json:"novelty"`      // 10% weight - uniqueness vs siblings
	DepthFactor float64 `json:"depth_factor"` // 10% weight - penalty for very deep thoughts
	Overall     float64 `json:"overall"`      // Weighted sum
}

// VertexHash is the hash function for graph vertices
func VertexHash(v *ThoughtVertex) string {
	return v.ID
}

// NewThoughtVertex creates a new vertex with defaults
func NewThoughtVertex(id, content string, thoughtType ThoughtType, confidence float64) *ThoughtVertex {
	return &ThoughtVertex{
		ID:           id,
		Content:      content,
		Type:         thoughtType,
		Confidence:   confidence,
		Score:        0.0,
		Depth:        0,
		ParentIDs:    []string{},
		ChildIDs:     []string{},
		KeyPoints:    []string{},
		RefinedCount: 0,
		CreatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}
}

// NewThoughtEdge creates a new edge
func NewThoughtEdge(id, fromID, toID string, edgeType EdgeType, weight float64) *ThoughtEdge {
	return &ThoughtEdge{
		ID:        id,
		FromID:    fromID,
		ToID:      toID,
		Type:      edgeType,
		Weight:    weight,
		CreatedAt: time.Now(),
	}
}
