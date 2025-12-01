package session

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// Exporter handles session export operations
type Exporter struct {
	storage storage.Storage
}

// NewExporter creates a new session exporter
func NewExporter(store storage.Storage) *Exporter {
	return &Exporter{storage: store}
}

// Export exports a session to a SessionExport structure
func (e *Exporter) Export(sessionID string, opts ExportOptions) (*SessionExport, error) {
	export := &SessionExport{
		Version:       SchemaVersion,
		ExportedAt:    time.Now(),
		SessionID:     sessionID,
		ExportOptions: opts,
		ToolsUsed:     make([]string, 0),
		Thoughts:      make([]types.Thought, 0),
		Branches:      make([]types.Branch, 0),
		Insights:      make([]types.Insight, 0),
	}

	// Export all thoughts
	thoughts, err := e.exportThoughts()
	if err != nil {
		return nil, fmt.Errorf("failed to export thoughts: %w", err)
	}
	export.Thoughts = thoughts
	export.ThoughtCount = len(thoughts)

	// Export all branches
	branches, err := e.exportBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to export branches: %w", err)
	}
	export.Branches = branches
	export.BranchCount = len(branches)

	// Extract insights from branches
	for _, branch := range branches {
		if branch.Insights != nil {
			for _, insight := range branch.Insights {
				if insight != nil {
					export.Insights = append(export.Insights, *insight)
				}
			}
		}
	}

	// Track tools used (based on thought types/modes)
	toolsUsed := make(map[string]bool)
	for _, t := range thoughts {
		toolsUsed["think"] = true
		if t.Mode != "" {
			toolsUsed[string(t.Mode)] = true
		}
	}
	for tool := range toolsUsed {
		export.ToolsUsed = append(export.ToolsUsed, tool)
	}

	return export, nil
}

// ExportToJSON exports a session to a JSON string
func (e *Exporter) ExportToJSON(sessionID string, opts ExportOptions) (*ExportResult, error) {
	export, err := e.Export(sessionID, opts)
	if err != nil {
		return nil, err
	}

	var data []byte
	if opts.Compress {
		data, err = compressJSON(export)
	} else {
		data, err = json.Marshal(export)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to serialize export: %w", err)
	}

	result := &ExportResult{
		ExportData:    string(data),
		SizeBytes:     len(data),
		ThoughtCount:  export.ThoughtCount,
		BranchCount:   export.BranchCount,
		ExportVersion: SchemaVersion,
		Compressed:    opts.Compress,
	}

	// If compressed, encode as base64
	if opts.Compress {
		result.ExportData = base64.StdEncoding.EncodeToString(data)
		result.SizeBytes = len(result.ExportData)
	}

	return result, nil
}

// exportThoughts retrieves all thoughts from storage
func (e *Exporter) exportThoughts() ([]types.Thought, error) {
	// Search for all thoughts (empty query returns all)
	results := e.storage.SearchThoughts("", "", 10000, 0) // Large limit to get all

	thoughts := make([]types.Thought, 0, len(results))
	for _, t := range results {
		if t != nil {
			thoughts = append(thoughts, *t)
		}
	}
	return thoughts, nil
}

// exportBranches retrieves all branches from storage
func (e *Exporter) exportBranches() ([]types.Branch, error) {
	branchList := e.storage.ListBranches()

	branches := make([]types.Branch, 0, len(branchList))
	for _, b := range branchList {
		if b != nil {
			branches = append(branches, *b)
		}
	}
	return branches, nil
}

// compressJSON compresses JSON data using gzip
func compressJSON(v any) ([]byte, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
