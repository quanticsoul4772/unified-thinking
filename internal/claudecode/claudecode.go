// Package claudecode provides optimizations for Claude Code and Claude Desktop integration.
//
// This package includes:
//   - Response format optimization (compact, minimal) for reduced token consumption
//   - Workflow presets for common development tasks
//   - Structured error system with recovery suggestions
//   - Session export/import for context preservation
//
// Key components:
//   - format/: Response formatters (CompactFormatter, MinimalFormatter)
//   - presets/: Built-in workflow presets (code-review, debug-analysis, etc.)
//   - errors/: Structured error types with recovery suggestions
//   - session/: Session export and import functionality
package claudecode

import (
	"unified-thinking/internal/claudecode/errors"
	"unified-thinking/internal/claudecode/format"
)

// Version is the current version of the claudecode package
const Version = "1.0.0"

// NewFormatter creates a new response formatter for the given format level
func NewFormatter(level format.FormatLevel) format.ResponseFormatter {
	return format.NewFormatter(level, format.DefaultOptions())
}

// NewFormatterWithOptions creates a new response formatter with custom options
func NewFormatterWithOptions(level format.FormatLevel, opts format.FormatOptions) format.ResponseFormatter {
	return format.NewFormatter(level, opts)
}

// NewError creates a new structured error with the given code and message
func NewError(code, message string) *errors.StructuredError {
	return errors.NewStructuredError(code, message)
}

// WrapError wraps an existing error with structured error information
func WrapError(code string, err error) *errors.StructuredError {
	return errors.WrapError(code, err)
}
