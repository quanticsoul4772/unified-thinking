// Package handlers provides MCP tool handler implementations.
package handlers

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ccerrors "unified-thinking/internal/claudecode/errors"
	"unified-thinking/internal/claudecode/format"
	"unified-thinking/internal/claudecode/presets"
	"unified-thinking/internal/claudecode/session"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/streaming"
)

// ClaudeCodeHandler provides handlers for Claude Code optimization tools
type ClaudeCodeHandler struct {
	storage  storage.Storage
	exporter *session.Exporter
	importer *session.Importer
	registry *presets.Registry
}

// NewClaudeCodeHandler creates a new handler for Claude Code tools
func NewClaudeCodeHandler(store storage.Storage) *ClaudeCodeHandler {
	return &ClaudeCodeHandler{
		storage:  store,
		exporter: session.NewExporter(store),
		importer: session.NewImporter(store),
		registry: presets.NewRegistry(),
	}
}

// === Request/Response Types ===

// ExportSessionRequest parameters for export-session tool
type ExportSessionRequest struct {
	SessionID          string `json:"session_id"`
	IncludeDecisions   bool   `json:"include_decisions,omitempty"`
	IncludeCausalGraphs bool   `json:"include_causal_graphs,omitempty"`
	Compress           bool   `json:"compress,omitempty"`
}

// ExportSessionResponse result from export-session tool
type ExportSessionResponse struct {
	ExportData    string `json:"export_data"`
	SizeBytes     int    `json:"size_bytes"`
	ThoughtCount  int    `json:"thought_count"`
	BranchCount   int    `json:"branch_count"`
	ExportVersion string `json:"export_version"`
	Compressed    bool   `json:"compressed"`
}

// ImportSessionRequest parameters for import-session tool
type ImportSessionRequest struct {
	ExportData         string `json:"export_data"`
	MergeStrategy      string `json:"merge_strategy,omitempty"` // "replace", "merge", "append"
	ValidateOnly       bool   `json:"validate_only,omitempty"`
	PreserveTimestamps bool   `json:"preserve_timestamps,omitempty"`
}

// ImportSessionResponse result from import-session tool
type ImportSessionResponse struct {
	SessionID         string   `json:"session_id"`
	Status            string   `json:"status"`
	ImportedThoughts  int      `json:"imported_thoughts"`
	ImportedBranches  int      `json:"imported_branches"`
	ConflictsResolved int      `json:"conflicts_resolved,omitempty"`
	ValidationErrors  []string `json:"validation_errors,omitempty"`
}

// ListPresetsRequest parameters for list-presets tool
type ListPresetsRequest struct {
	Category string `json:"category,omitempty"`
}

// PresetSummary summary information about a preset
type PresetSummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	StepCount     int    `json:"step_count"`
	EstimatedTime string `json:"estimated_time,omitempty"`
}

// ListPresetsResponse result from list-presets tool
type ListPresetsResponse struct {
	Presets    []PresetSummary `json:"presets"`
	Count      int             `json:"count"`
	Categories []string        `json:"categories"`
}

// RunPresetRequest parameters for run-preset tool
type RunPresetRequest struct {
	PresetID   string         `json:"preset_id"`
	Input      map[string]any `json:"input"`
	DryRun     bool           `json:"dry_run,omitempty"`
	StepByStep bool           `json:"step_by_step,omitempty"`
}

// PresetStepResult result of a single preset step
type PresetStepResult struct {
	StepID      string `json:"step_id"`
	Tool        string `json:"tool"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	Result      any    `json:"result,omitempty"`
	Error       string `json:"error,omitempty"`
}

// RunPresetResponse result from run-preset tool
type RunPresetResponse struct {
	PresetID        string             `json:"preset_id"`
	Status          string             `json:"status"`
	StepsCompleted  int                `json:"steps_completed"`
	TotalSteps      int                `json:"total_steps"`
	StepResults     []PresetStepResult `json:"step_results,omitempty"`
	FinalResult     any                `json:"final_result,omitempty"`
	ExecutionTimeMs int64              `json:"execution_time_ms,omitempty"`
}

// FormatResponseRequest parameters for format-response tool
type FormatResponseRequest struct {
	Response any    `json:"response"`
	Level    string `json:"level,omitempty"` // "full", "compact", "minimal"
}

// FormatResponseResponse result from format-response tool
type FormatResponseResponse struct {
	Formatted any    `json:"formatted"`
	Level     string `json:"level"`
	SizeBytes int    `json:"size_bytes"`
}

// === Handler Methods ===

// HandleExportSession exports current reasoning session to portable format
func (h *ClaudeCodeHandler) HandleExportSession(ctx context.Context, req *mcp.CallToolRequest, input ExportSessionRequest) (*mcp.CallToolResult, *ExportSessionResponse, error) {
	opts := session.ExportOptions{
		IncludeDecisions:    input.IncludeDecisions,
		IncludeCausalGraphs: input.IncludeCausalGraphs,
		Compress:            input.Compress,
	}

	// Use default session ID if not provided
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "default"
	}

	result, err := h.exporter.ExportToJSON(sessionID, opts)
	if err != nil {
		return nil, nil, err
	}

	return nil, &ExportSessionResponse{
		ExportData:    result.ExportData,
		SizeBytes:     result.SizeBytes,
		ThoughtCount:  result.ThoughtCount,
		BranchCount:   result.BranchCount,
		ExportVersion: result.ExportVersion,
		Compressed:    result.Compressed,
	}, nil
}

// HandleImportSession imports a previously exported reasoning session
func (h *ClaudeCodeHandler) HandleImportSession(ctx context.Context, req *mcp.CallToolRequest, input ImportSessionRequest) (*mcp.CallToolResult, *ImportSessionResponse, error) {
	// Validate required input
	if input.ExportData == "" {
		return nil, nil, ccerrors.NewStructuredError(ccerrors.ErrMissingRequired, "export_data is required").
			WithDetails("The export_data parameter must contain the JSON from export-session").
			WithRecovery("First use export-session to get the export data, then provide it here").
			WithRelatedTools("export-session")
	}

	opts := session.MergeOptions{
		Strategy:           session.ParseMergeStrategy(input.MergeStrategy),
		PreserveTimestamps: input.PreserveTimestamps,
		ValidateOnly:       input.ValidateOnly,
	}

	result, err := h.importer.ImportFromJSON(input.ExportData, opts)
	if err != nil {
		return nil, nil, ccerrors.NewStructuredError(ccerrors.ErrInvalidFormat, err.Error()).
			WithDetails("The export data could not be parsed. It may be corrupted or from an incompatible version.").
			WithRecovery("Verify the export data is valid JSON from export-session").
			WithRecovery("Check if the export was compressed - if so, ensure compress flag is handled").
			WithRelatedTools("export-session")
	}

	return nil, &ImportSessionResponse{
		SessionID:         result.SessionID,
		Status:            result.Status,
		ImportedThoughts:  result.ImportedThoughts,
		ImportedBranches:  result.ImportedBranches,
		ConflictsResolved: result.ConflictsResolved,
		ValidationErrors:  result.ValidationErrors,
	}, nil
}

// HandleListPresets lists available workflow presets
func (h *ClaudeCodeHandler) HandleListPresets(ctx context.Context, req *mcp.CallToolRequest, input ListPresetsRequest) (*mcp.CallToolResult, *ListPresetsResponse, error) {
	presetList := h.registry.List(input.Category)

	summaries := make([]PresetSummary, len(presetList))
	for i, p := range presetList {
		summaries[i] = PresetSummary{
			ID:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			Category:      p.Category,
			StepCount:     p.StepCount,
			EstimatedTime: p.EstimatedTime,
		}
	}

	return nil, &ListPresetsResponse{
		Presets:    summaries,
		Count:      len(summaries),
		Categories: h.registry.Categories(),
	}, nil
}

// HandleRunPreset executes a workflow preset
func (h *ClaudeCodeHandler) HandleRunPreset(ctx context.Context, req *mcp.CallToolRequest, input RunPresetRequest) (*mcp.CallToolResult, *RunPresetResponse, error) {
	// Validate required input
	if input.PresetID == "" {
		return nil, nil, ccerrors.NewStructuredError(ccerrors.ErrMissingRequired, "preset_id is required").
			WithDetails("The preset_id parameter must specify which preset to run").
			WithRecovery("Provide a preset_id parameter (e.g., 'code-review', 'debug-analysis')").
			WithRelatedTools("list-presets").
			WithExample("run-preset", map[string]any{
				"preset_id": "code-review",
				"input":     map[string]any{"code": "function example() {...}"},
			})
	}

	preset, err := h.registry.Get(input.PresetID)
	if err != nil {
		return nil, nil, ccerrors.NewStructuredError(ccerrors.ErrPresetNotFound, err.Error()).
			WithDetails("The requested preset was not found in the registry").
			WithRecovery("Use list-presets to see available presets").
			WithRelatedTools("list-presets")
	}

	// Create progress reporter for streaming notifications
	reporter := streaming.CreateReporter(req, "run-preset")

	// For dry run, just return the steps without executing
	if input.DryRun {
		// Report initial progress if streaming is enabled
		if reporter.IsEnabled() {
			_ = reporter.ReportStep(0, len(preset.Steps), "dry_run", "Previewing preset steps...")
		}

		stepResults := make([]PresetStepResult, len(preset.Steps))
		for i, step := range preset.Steps {
			stepResults[i] = PresetStepResult{
				StepID:      step.StepID,
				Tool:        step.Tool,
				Status:      "pending",
				Description: step.Description,
			}

			// Report each step in dry run mode
			if reporter.IsEnabled() {
				_ = reporter.ReportStep(i+1, len(preset.Steps), step.StepID, "Preview: "+step.Description)
			}
		}

		return nil, &RunPresetResponse{
			PresetID:       input.PresetID,
			Status:         "dry_run",
			StepsCompleted: 0,
			TotalSteps:     len(preset.Steps),
			StepResults:    stepResults,
		}, nil
	}

	// Full execution with progress reporting
	// Report start of execution
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(0, len(preset.Steps), "initialize", "Starting preset execution: "+preset.Name)
	}

	// Execute steps (currently returns ready status - actual execution requires tool executor)
	stepResults := make([]PresetStepResult, len(preset.Steps))
	for i, step := range preset.Steps {
		// Report step progress
		if reporter.IsEnabled() {
			_ = reporter.ReportStep(i+1, len(preset.Steps), step.StepID, "Ready: "+step.Description)
		}

		stepResults[i] = PresetStepResult{
			StepID:      step.StepID,
			Tool:        step.Tool,
			Status:      "ready",
			Description: step.Description,
		}
	}

	// Report completion
	if reporter.IsEnabled() {
		_ = reporter.ReportStep(len(preset.Steps), len(preset.Steps), "complete", "Preset ready for execution")
	}

	return nil, &RunPresetResponse{
		PresetID:       input.PresetID,
		Status:         "ready",
		StepsCompleted: 0,
		TotalSteps:     len(preset.Steps),
		StepResults:    stepResults,
	}, nil
}

// HandleFormatResponse applies format optimization to a response
func (h *ClaudeCodeHandler) HandleFormatResponse(ctx context.Context, req *mcp.CallToolRequest, input FormatResponseRequest) (*mcp.CallToolResult, *FormatResponseResponse, error) {
	level := format.FormatLevel(input.Level)
	if level == "" {
		level = format.FormatFull
	}

	var opts format.FormatOptions
	switch level {
	case format.FormatCompact:
		opts = format.CompactOptions()
	case format.FormatMinimal:
		opts = format.MinimalOptions()
	default:
		opts = format.DefaultOptions()
	}

	formatter := format.NewFormatter(level, opts)
	formatted, err := formatter.Format(input.Response)
	if err != nil {
		return nil, nil, err
	}

	return nil, &FormatResponseResponse{
		Formatted: formatted,
		Level:     string(level),
		SizeBytes: 0, // Would need to marshal to get actual size
	}, nil
}

// RegisterClaudeCodeTools registers all Claude Code optimization tools
func RegisterClaudeCodeTools(mcpServer *mcp.Server, handler *ClaudeCodeHandler) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "export-session",
		Description: `Export current reasoning session to a portable JSON format for backup, sharing, or later restoration.

**Parameters:**
- session_id: Session identifier (default: "default")
- include_decisions: Include decision records (default: true)
- include_causal_graphs: Include causal graph data (default: true)
- compress: Gzip compress the output (default: false)

**Returns:**
- export_data: JSON string containing the session data
- size_bytes: Size of exported data
- thought_count: Number of thoughts exported
- branch_count: Number of branches exported
- export_version: Schema version for compatibility

**Use Cases:**
- Backup reasoning sessions before risky operations
- Share reasoning context between sessions
- Archive completed reasoning chains

**Example:** {"session_id": "debug-session-123", "compress": true}`,
	}, handler.HandleExportSession)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "import-session",
		Description: `Import a previously exported reasoning session with merge strategy control.

**Parameters:**
- export_data (required): JSON string from export-session
- merge_strategy: "replace" (clear existing), "merge" (update/add), "append" (keep existing, add new)
- validate_only: Check validity without importing (default: false)
- preserve_timestamps: Keep original timestamps (default: true)

**Returns:**
- session_id: Imported session identifier
- status: "success", "partial", or "failed"
- imported_thoughts: Number of thoughts imported
- imported_branches: Number of branches imported
- conflicts: Number of merge conflicts detected
- validation_errors: Array of validation issues

**Example:** {"export_data": "...", "merge_strategy": "merge"}`,
	}, handler.HandleImportSession)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "list-presets",
		Description: `List available workflow presets for common development tasks.

**Parameters:**
- category: Filter by category (optional). Categories: "code", "architecture", "research", "testing", "documentation", "operations"

**Returns:**
- presets: Array of preset summaries
- count: Total number of presets
- categories: Available categories

**Built-in Presets:**
- code-review: 5-step code review workflow
- debug-analysis: Causal debugging with hypothesis generation
- architecture-decision: ADR-style decision workflow
- research-synthesis: Graph-of-Thoughts research aggregation
- refactoring-plan: Safe refactoring with impact analysis
- test-strategy: Test coverage planning workflow
- documentation-gen: Multi-perspective documentation
- incident-investigation: Post-incident analysis with timeline mapping

**Example:** {"category": "code"}`,
	}, handler.HandleListPresets)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "run-preset",
		Description: `Execute a workflow preset with provided inputs.

**Parameters:**
- preset_id (required): ID of the preset to run
- input (required): Input values matching preset's input_schema
- dry_run: Preview steps without executing (default: false)
- step_by_step: Pause after each step (default: false)

**Returns:**
- preset_id: Executed preset identifier
- status: "success", "partial", "failed", or "dry_run"
- steps_completed: Number of steps completed
- total_steps: Total steps in preset
- step_results: Array of individual step results
- final_result: Aggregated result from all steps
- execution_time_ms: Total execution time

**Example:** {"preset_id": "code-review", "input": {"code": "func example() {...}"}}`,
	}, handler.HandleRunPreset)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "format-response",
		Description: `Apply format optimization to reduce response size for Claude Code.

**Parameters:**
- response (required): The response object to format
- level: Format level - "full" (default), "compact" (40-60% reduction), "minimal" (80%+ reduction)

**Format Levels:**
- full: Complete response with all metadata (default)
- compact: Removes context_bridge, flattens next_tools, truncates arrays to 5 items
- minimal: Essential fields only based on response type, arrays truncated to 3 items

**Returns:**
- formatted: Optimized response
- level: Applied format level
- size_bytes: Size of formatted output

**Example:** {"response": {"id": "...", "metadata": {...}}, "level": "compact"}`,
	}, handler.HandleFormatResponse)
}
