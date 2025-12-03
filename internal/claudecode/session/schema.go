// Package session provides session export and import functionality for context preservation.
//
// This package enables exporting reasoning sessions to JSON and importing them back,
// allowing context to be preserved across Claude Code restarts.
package session

import (
	"time"
	"unified-thinking/internal/types"
)

// SchemaVersion is the current export schema version
const SchemaVersion = "1.0"

// SessionExport represents a complete exported session
type SessionExport struct {
	// Schema information
	Version    string    `json:"version"`
	ExportedAt time.Time `json:"exported_at"`

	// Session identification
	SessionID   string `json:"session_id"`
	Description string `json:"description,omitempty"`

	// Core reasoning data
	Thoughts []types.Thought `json:"thoughts"`
	Branches []types.Branch  `json:"branches"`
	Insights []types.Insight `json:"insights,omitempty"`

	// Optional components
	Decisions    []DecisionExport    `json:"decisions,omitempty"`
	CausalGraphs []CausalGraphExport `json:"causal_graphs,omitempty"`
	Checkpoints  []CheckpointExport  `json:"checkpoints,omitempty"`
	Beliefs      []BeliefExport      `json:"beliefs,omitempty"`

	// Session metadata
	ToolsUsed    []string      `json:"tools_used"`
	Duration     time.Duration `json:"duration"`
	ThoughtCount int           `json:"thought_count"`
	BranchCount  int           `json:"branch_count"`

	// Export options used
	ExportOptions ExportOptions `json:"export_options"`
}

// ExportOptions configures what to include in the export
type ExportOptions struct {
	IncludeDecisions    bool `json:"include_decisions"`
	IncludeCausalGraphs bool `json:"include_causal_graphs"`
	IncludeCheckpoints  bool `json:"include_checkpoints"`
	IncludeBeliefs      bool `json:"include_beliefs"`
	Compress            bool `json:"compress"`
}

// DefaultExportOptions returns the default export options
func DefaultExportOptions() ExportOptions {
	return ExportOptions{
		IncludeDecisions:    true,
		IncludeCausalGraphs: true,
		IncludeCheckpoints:  true,
		IncludeBeliefs:      true,
		Compress:            false,
	}
}

// DecisionExport represents an exported decision
type DecisionExport struct {
	ID             string            `json:"id"`
	Question       string            `json:"question"`
	Options        []OptionExport    `json:"options"`
	Criteria       []CriterionExport `json:"criteria"`
	SelectedOption string            `json:"selected_option,omitempty"`
	Confidence     float64           `json:"confidence"`
	Reasoning      string            `json:"reasoning,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	Metadata       map[string]any    `json:"metadata,omitempty"`
}

// OptionExport represents an exported decision option
type OptionExport struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Scores      map[string]float64 `json:"scores"`
	TotalScore  float64            `json:"total_score"`
}

// CriterionExport represents an exported decision criterion
type CriterionExport struct {
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	Maximize bool    `json:"maximize"`
}

// CausalGraphExport represents an exported causal graph
type CausalGraphExport struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Nodes       []NodeExport `json:"nodes"`
	Edges       []EdgeExport `json:"edges"`
	CreatedAt   time.Time    `json:"created_at"`
}

// NodeExport represents a causal graph node
type NodeExport struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type,omitempty"`
	Value    any            `json:"value,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// EdgeExport represents a causal graph edge
type EdgeExport struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Type     string  `json:"type,omitempty"`
	Strength float64 `json:"strength,omitempty"`
}

// CheckpointExport represents an exported checkpoint
type CheckpointExport struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	ThoughtIDs  []string       `json:"thought_ids"`
	BranchIDs   []string       `json:"branch_ids"`
	CreatedAt   time.Time      `json:"created_at"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// BeliefExport represents an exported probabilistic belief
type BeliefExport struct {
	ID          string         `json:"id"`
	Statement   string         `json:"statement"`
	Probability float64        `json:"probability"`
	PriorProb   float64        `json:"prior_prob"`
	Evidence    []string       `json:"evidence"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	SessionID         string   `json:"session_id"`
	ImportedThoughts  int      `json:"imported_thoughts"`
	ImportedBranches  int      `json:"imported_branches"`
	ImportedInsights  int      `json:"imported_insights"`
	ImportedDecisions int      `json:"imported_decisions,omitempty"`
	MergedCount       int      `json:"merged_count"`
	ConflictsResolved int      `json:"conflicts_resolved"`
	ValidationErrors  []string `json:"validation_errors,omitempty"`
	Status            string   `json:"status"` // success, partial, failed
}

// ExportResult represents the result of an export operation
type ExportResult struct {
	ExportData    string `json:"export_data"` // JSON string or base64
	SizeBytes     int    `json:"size_bytes"`
	ThoughtCount  int    `json:"thought_count"`
	BranchCount   int    `json:"branch_count"`
	ExportVersion string `json:"export_version"`
	Compressed    bool   `json:"compressed"`
}
