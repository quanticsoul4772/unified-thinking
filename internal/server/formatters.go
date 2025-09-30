package server

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// toJSONContent converts any data structure to MCP TextContent with JSON
// This is consumed by Claude AI directly - no human-readable formatting needed
func toJSONContent(data interface{}) []mcp.Content {
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
