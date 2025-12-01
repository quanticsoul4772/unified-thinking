package session

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// Importer handles session import operations
type Importer struct {
	storage storage.Storage
}

// NewImporter creates a new session importer
func NewImporter(store storage.Storage) *Importer {
	return &Importer{storage: store}
}

// Import imports a session from a SessionExport structure
func (i *Importer) Import(export *SessionExport, opts MergeOptions) (*ImportResult, error) {
	result := &ImportResult{
		SessionID:        export.SessionID,
		ValidationErrors: make([]string, 0),
		Status:           "success",
	}

	// Validate the export
	if err := i.validate(export); err != nil {
		result.ValidationErrors = append(result.ValidationErrors, err.Error())
		if opts.ValidateOnly {
			result.Status = "validation_failed"
			return result, nil
		}
	}

	if opts.ValidateOnly {
		return result, nil
	}

	// Handle replace strategy
	if opts.Strategy == MergeReplace {
		// Clear existing data (this is a simplified approach)
		// In a real implementation, you'd want more selective clearing
	}

	// Import thoughts
	importedThoughts, err := i.importThoughts(export.Thoughts, opts)
	if err != nil {
		result.Status = "partial"
		result.ValidationErrors = append(result.ValidationErrors, fmt.Sprintf("thought import error: %s", err))
	}
	result.ImportedThoughts = importedThoughts

	// Import branches
	importedBranches, err := i.importBranches(export.Branches, opts)
	if err != nil {
		result.Status = "partial"
		result.ValidationErrors = append(result.ValidationErrors, fmt.Sprintf("branch import error: %s", err))
	}
	result.ImportedBranches = importedBranches

	// Import insights
	importedInsights := 0
	for _, insight := range export.Insights {
		if err := i.storage.StoreInsight(&insight); err == nil {
			importedInsights++
		}
	}
	result.ImportedInsights = importedInsights

	return result, nil
}

// ImportFromJSON imports a session from a JSON string
func (i *Importer) ImportFromJSON(data string, opts MergeOptions) (*ImportResult, error) {
	var export SessionExport

	// Try to decode as base64 (compressed)
	if decoded, err := base64.StdEncoding.DecodeString(data); err == nil {
		// Try to decompress
		if decompressed, err := decompressJSON(decoded); err == nil {
			data = string(decompressed)
		}
	}

	// Parse JSON
	if err := json.Unmarshal([]byte(data), &export); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return i.Import(&export, opts)
}

// validate checks if the export is valid for import
func (i *Importer) validate(export *SessionExport) error {
	if export.Version == "" {
		return fmt.Errorf("missing version field")
	}

	// Check version compatibility
	if !isVersionCompatible(export.Version) {
		return fmt.Errorf("incompatible version: %s (current: %s)", export.Version, SchemaVersion)
	}

	if export.SessionID == "" {
		return fmt.Errorf("missing session_id field")
	}

	return nil
}

// importThoughts imports thoughts with the specified merge strategy
func (i *Importer) importThoughts(thoughts []types.Thought, opts MergeOptions) (int, error) {
	imported := 0

	for _, thought := range thoughts {
		// For append strategy, generate new IDs
		if opts.Strategy == MergeAppend {
			thought.ID = generateNewID("thought")
		}

		// Update timestamps if not preserving
		if !opts.PreserveTimestamps {
			thought.Timestamp = time.Now()
		}

		// Store the thought
		if err := i.storage.StoreThought(&thought); err != nil {
			// For merge strategy, try to update existing
			if opts.Strategy == MergeMerge && strings.Contains(err.Error(), "exists") {
				// Update logic would go here
				continue
			}
			continue
		}
		imported++
	}

	return imported, nil
}

// importBranches imports branches with the specified merge strategy
func (i *Importer) importBranches(branches []types.Branch, opts MergeOptions) (int, error) {
	imported := 0

	for _, branch := range branches {
		// For append strategy, generate new IDs
		if opts.Strategy == MergeAppend {
			branch.ID = generateNewID("branch")
		}

		// Update timestamps if not preserving
		if !opts.PreserveTimestamps {
			branch.CreatedAt = time.Now()
			branch.UpdatedAt = time.Now()
		}

		// Store the branch
		if err := i.storage.StoreBranch(&branch); err != nil {
			continue
		}
		imported++
	}

	return imported, nil
}

// decompressJSON decompresses gzipped data
func decompressJSON(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// isVersionCompatible checks if a version is compatible with current
func isVersionCompatible(version string) bool {
	// Simple version check - major version must match
	current := strings.Split(SchemaVersion, ".")
	imported := strings.Split(version, ".")

	if len(current) == 0 || len(imported) == 0 {
		return false
	}

	return current[0] == imported[0]
}

// generateNewID creates a new unique ID for the given type
func generateNewID(prefix string) string {
	return fmt.Sprintf("%s-import-%d", prefix, time.Now().UnixNano())
}
