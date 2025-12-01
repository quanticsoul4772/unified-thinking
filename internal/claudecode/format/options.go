// Package format provides response formatting optimizations for Claude Code integration.
//
// This package supports multiple format levels to reduce token consumption:
//   - Full: Default format with all metadata (100% size)
//   - Compact: Reduced format without context_bridge (40-60% size)
//   - Minimal: Result-only format with essential fields (20% size)
package format

// FormatLevel represents the level of detail in response formatting
type FormatLevel string

const (
	// FormatFull includes all metadata and context (default)
	FormatFull FormatLevel = "full"
	// FormatCompact removes context_bridge and flattens metadata
	FormatCompact FormatLevel = "compact"
	// FormatMinimal includes only essential result fields
	FormatMinimal FormatLevel = "minimal"
)

// FormatOptions configures response formatting behavior
type FormatOptions struct {
	// Level specifies the format detail level
	Level FormatLevel `json:"level"`
	// IncludeMetadata controls whether metadata is included (default: true for full/compact)
	IncludeMetadata bool `json:"include_metadata"`
	// IncludeTimings controls whether timing information is included
	IncludeTimings bool `json:"include_timings"`
	// MaxArrayLength truncates arrays longer than this (0 = no limit)
	MaxArrayLength int `json:"max_array_length"`
	// OmitEmpty removes empty/null fields from output
	OmitEmpty bool `json:"omit_empty"`
	// FlattenNextTools moves suggested_next_tools to top level as next_tools
	FlattenNextTools bool `json:"flatten_next_tools"`
}

// DefaultOptions returns the default format options for the full level
func DefaultOptions() FormatOptions {
	return FormatOptions{
		Level:            FormatFull,
		IncludeMetadata:  true,
		IncludeTimings:   true,
		MaxArrayLength:   0,
		OmitEmpty:        false,
		FlattenNextTools: false,
	}
}

// CompactOptions returns format options optimized for compact output
func CompactOptions() FormatOptions {
	return FormatOptions{
		Level:            FormatCompact,
		IncludeMetadata:  true,
		IncludeTimings:   false,
		MaxArrayLength:   5,
		OmitEmpty:        true,
		FlattenNextTools: true,
	}
}

// MinimalOptions returns format options for minimal output
func MinimalOptions() FormatOptions {
	return FormatOptions{
		Level:            FormatMinimal,
		IncludeMetadata:  false,
		IncludeTimings:   false,
		MaxArrayLength:   3,
		OmitEmpty:        true,
		FlattenNextTools: false,
	}
}

// IsValid checks if the format level is valid
func (l FormatLevel) IsValid() bool {
	switch l {
	case FormatFull, FormatCompact, FormatMinimal:
		return true
	default:
		return false
	}
}

// ParseFormatLevel parses a string into a FormatLevel, defaulting to FormatFull
func ParseFormatLevel(s string) FormatLevel {
	switch s {
	case "compact":
		return FormatCompact
	case "minimal":
		return FormatMinimal
	case "full", "":
		return FormatFull
	default:
		return FormatFull
	}
}
