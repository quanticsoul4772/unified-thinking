// Package contextbridge provides cross-session context retrieval for tool responses.
package contextbridge

// Signature represents a fingerprint of a problem for similarity matching
type Signature struct {
	Fingerprint  string    `json:"fingerprint"`
	Domain       string    `json:"domain"`
	KeyConcepts  []string  `json:"key_concepts"`
	ToolSequence []string  `json:"tool_sequence"`
	Complexity   float64   `json:"complexity"`
	Embedding    []float32 `json:"embedding,omitempty"` // Semantic embedding for similarity
}

// CandidateWithSignature combines trajectory metadata with its signature
// to enable single-query retrieval and avoid N+1 performance issues
type CandidateWithSignature struct {
	TrajectoryID string
	SessionID    string
	Description  string
	SuccessScore float64
	QualityScore float64
	Signature    *Signature
}

// Match represents a matched trajectory with similarity score
type Match struct {
	TrajectoryID string  `json:"trajectory_id"`
	SessionID    string  `json:"session_id"`
	Similarity   float64 `json:"similarity"`
	Summary      string  `json:"summary"`
	SuccessScore float64 `json:"success_score"`
	QualityScore float64 `json:"quality_score"`
}

// ContextBridgeData represents the context bridge enrichment in a response
type ContextBridgeData struct {
	Version        string   `json:"version"`
	Matches        []*Match `json:"matches"`
	Recommendation string   `json:"recommendation,omitempty"`
}
