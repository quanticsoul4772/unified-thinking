package server

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/claudecode/format"
)

// Global response formatter configuration
var (
	responseFormatter     format.ResponseFormatter
	responseFormatterOnce sync.Once
)

// initResponseFormatter initializes the global response formatter based on environment
func initResponseFormatter() {
	level := format.FormatLevel(os.Getenv("RESPONSE_FORMAT"))
	if level == "" {
		level = format.FormatFull // Default to full output
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

	responseFormatter = format.NewFormatter(level, opts)
}

// getResponseFormatter returns the global response formatter
func getResponseFormatter() format.ResponseFormatter {
	responseFormatterOnce.Do(initResponseFormatter)
	return responseFormatter
}

// toJSONContent converts any data structure to MCP TextContent with JSON
// This is consumed by Claude AI directly - no human-readable formatting needed
// Applies response formatting based on RESPONSE_FORMAT environment variable
func toJSONContent(data interface{}) []mcp.Content {
	// Apply formatting if not at full level
	formatter := getResponseFormatter()
	if formatter.Level() != format.FormatFull {
		formatted, err := formatter.Format(data)
		if err == nil {
			data = formatted
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		// Return error as JSON
		errData := map[string]string{"error": err.Error()}
		jsonData, _ = json.Marshal(errData)
	}

	return []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}
}

// toJSONContentWithFormat converts data to MCP TextContent with optional format level override
// If formatLevel is empty, uses the default from RESPONSE_FORMAT environment variable
func toJSONContentWithFormat(data interface{}, formatLevel string) []mcp.Content {
	// Apply formatting based on explicit level or default
	var formatter format.ResponseFormatter
	if formatLevel != "" {
		level := format.ParseFormatLevel(formatLevel)
		var opts format.FormatOptions
		switch level {
		case format.FormatCompact:
			opts = format.CompactOptions()
		case format.FormatMinimal:
			opts = format.MinimalOptions()
		default:
			opts = format.DefaultOptions()
		}
		formatter = format.NewFormatter(level, opts)
	} else {
		formatter = getResponseFormatter()
	}

	if formatter.Level() != format.FormatFull {
		formatted, err := formatter.Format(data)
		if err == nil {
			data = formatted
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		// Return error as JSON
		errData := map[string]string{"error": err.Error()}
		jsonData, _ = json.Marshal(errData)
	}

	return []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}
}

// ExtractFormatLevel extracts the format_level parameter from request arguments
func ExtractFormatLevel(args map[string]interface{}) string {
	if args == nil {
		return ""
	}
	if level, ok := args["format_level"].(string); ok {
		return level
	}
	return ""
}
