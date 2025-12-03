package server

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToJSONContentWithFormat(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		formatLevel string
		wantSmaller bool // Should result in smaller output than full
	}{
		{
			name: "full format (default)",
			data: map[string]interface{}{
				"thought_id": "test-1",
				"mode":       "linear",
				"confidence": 0.8,
				"metadata": map[string]interface{}{
					"context_bridge": map[string]interface{}{
						"related": []string{"a", "b", "c"},
					},
				},
			},
			formatLevel: "",
			wantSmaller: false,
		},
		{
			name: "compact format",
			data: map[string]interface{}{
				"thought_id": "test-1",
				"mode":       "linear",
				"confidence": 0.8,
				"metadata": map[string]interface{}{
					"context_bridge": map[string]interface{}{
						"related": []string{"a", "b", "c"},
					},
				},
			},
			formatLevel: "compact",
			wantSmaller: true,
		},
		{
			name: "minimal format",
			data: map[string]interface{}{
				"thought_id": "test-1",
				"mode":       "linear",
				"confidence": 0.8,
				"metadata": map[string]interface{}{
					"extra_data": "should be removed",
				},
			},
			formatLevel: "minimal",
			wantSmaller: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toJSONContentWithFormat(tt.data, tt.formatLevel)

			if len(result) != 1 {
				t.Errorf("Expected 1 content item, got %d", len(result))
				return
			}

			// Get the text content
			textContent, ok := result[0].(*mcp.TextContent)
			if !ok {
				t.Error("Expected TextContent type")
				return
			}

			// Verify it's valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
				t.Errorf("Result is not valid JSON: %v", err)
			}

			// For compact/minimal, verify context_bridge is removed (if it was present)
			if tt.formatLevel == "compact" || tt.formatLevel == "minimal" {
				if _, hasContextBridge := parsed["context_bridge"]; hasContextBridge {
					t.Error("Expected context_bridge to be removed in compact/minimal format")
				}
			}
		})
	}
}

func TestExtractFormatLevel(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		expected string
	}{
		{
			name:     "nil args",
			args:     nil,
			expected: "",
		},
		{
			name:     "empty args",
			args:     map[string]interface{}{},
			expected: "",
		},
		{
			name: "format_level present",
			args: map[string]interface{}{
				"format_level": "compact",
			},
			expected: "compact",
		},
		{
			name: "format_level with other fields",
			args: map[string]interface{}{
				"content":      "test",
				"format_level": "minimal",
			},
			expected: "minimal",
		},
		{
			name: "format_level wrong type",
			args: map[string]interface{}{
				"format_level": 123,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFormatLevel(tt.args)
			if result != tt.expected {
				t.Errorf("ExtractFormatLevel() = %q, want %q", result, tt.expected)
			}
		})
	}
}
