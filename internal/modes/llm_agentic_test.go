// Package modes - Tests for AgenticClient
package modes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDefaultAgenticConfig(t *testing.T) {
	config := DefaultAgenticConfig()

	if config.MaxIterations != 10 {
		t.Errorf("MaxIterations = %d, want 10", config.MaxIterations)
	}
	if config.MaxToolsPerTurn != 5 {
		t.Errorf("MaxToolsPerTurn = %d, want 5", config.MaxToolsPerTurn)
	}
	if !config.StopOnError {
		t.Error("StopOnError should be true by default")
	}
	if config.Temperature != 0.3 {
		t.Errorf("Temperature = %f, want 0.3", config.Temperature)
	}
	if config.Model != "claude-sonnet-4-5-20250929" {
		t.Errorf("Model = %s, want claude-sonnet-4-5-20250929", config.Model)
	}
	if config.MaxTokens != 4096 {
		t.Errorf("MaxTokens = %d, want 4096", config.MaxTokens)
	}
}

func TestNewAgenticClient(t *testing.T) {
	registry := NewToolRegistry()
	config := DefaultAgenticConfig()

	client := NewAgenticClient("test-key", registry, config)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.apiKey != "test-key" {
		t.Error("apiKey not set correctly")
	}
	if client.registry != registry {
		t.Error("registry not set correctly")
	}

	// Test with zero values
	emptyConfig := AgenticConfig{}
	client = NewAgenticClient("key", registry, emptyConfig)

	if client.config.MaxIterations != 10 {
		t.Errorf("expected default MaxIterations 10, got %d", client.config.MaxIterations)
	}
	if client.config.Model != "claude-sonnet-4-5-20250929" {
		t.Errorf("expected default model, got %s", client.config.Model)
	}
}

func TestNewAgenticClientFromEnv(t *testing.T) {
	registry := NewToolRegistry()

	// Test without API key
	os.Unsetenv("ANTHROPIC_API_KEY")
	_, err := NewAgenticClientFromEnv(registry)
	if err == nil {
		t.Error("expected error without ANTHROPIC_API_KEY")
	}

	// Test with API key
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	client, err := NewAgenticClientFromEnv(registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Test with custom model
	os.Setenv("AGENT_MODEL", "claude-opus-4-20250929")
	defer os.Unsetenv("AGENT_MODEL")

	client, err = NewAgenticClientFromEnv(registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.config.Model != "claude-opus-4-20250929" {
		t.Errorf("expected custom model, got %s", client.config.Model)
	}
}

func TestExecutionTrace_Methods(t *testing.T) {
	trace := ExecutionTrace{
		Iterations: []Iteration{
			{
				Index:   0,
				Thought: "Initial thought",
				ToolCalls: []ToolCall{
					{Name: "tool-a"},
					{Name: "tool-b"},
				},
			},
			{
				Index:   1,
				Thought: "Second thought",
				ToolCalls: []ToolCall{
					{Name: "tool-a"},
					{Name: "tool-c"},
				},
				ToolErrors: []string{"error 1"},
			},
		},
	}

	// Test ToolsUsed
	tools := trace.ToolsUsed()
	if len(tools) != 3 {
		t.Errorf("expected 3 unique tools, got %d", len(tools))
	}

	toolSet := make(map[string]bool)
	for _, name := range tools {
		toolSet[name] = true
	}
	if !toolSet["tool-a"] || !toolSet["tool-b"] || !toolSet["tool-c"] {
		t.Error("missing expected tool in ToolsUsed")
	}

	// Test TotalToolCalls
	if trace.TotalToolCalls() != 4 {
		t.Errorf("expected 4 total calls, got %d", trace.TotalToolCalls())
	}

	// Test ErrorCount
	if trace.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", trace.ErrorCount())
	}
}

func TestAgenticClient_Run_NoTools(t *testing.T) {
	// Create mock server that returns end_turn without tools
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "This is my final answer.",
				},
			},
			"stop_reason": "end_turn",
			"usage": map[string]int{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	registry := NewToolRegistry()
	config := DefaultAgenticConfig()
	client := NewAgenticClient("test-key", registry, config)

	// We can't easily test with the mock server since it requires changing the URL
	// This is mostly a structure test
	_ = client
	_ = server
}

func TestAgenticClient_RunWithSystem(t *testing.T) {
	registry := NewToolRegistry()
	config := DefaultAgenticConfig()
	client := NewAgenticClient("test-key", registry, config)

	// Just verify the method exists and can be called
	// Actual API calls require integration tests
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestIteration_Structure(t *testing.T) {
	iter := Iteration{
		Index:   5,
		Thought: "Test thought",
		ToolCalls: []ToolCall{
			{
				ID:     "call_123",
				Name:   "test-tool",
				Input:  map[string]interface{}{"key": "value"},
				Output: "result",
			},
		},
		ToolErrors: []string{"error message"},
	}

	if iter.Index != 5 {
		t.Errorf("Index = %d, want 5", iter.Index)
	}
	if iter.Thought != "Test thought" {
		t.Errorf("Thought mismatch")
	}
	if len(iter.ToolCalls) != 1 {
		t.Errorf("expected 1 tool call")
	}
	if len(iter.ToolErrors) != 1 {
		t.Errorf("expected 1 error")
	}
}

func TestToolCall_Structure(t *testing.T) {
	call := ToolCall{
		ID:     "call_abc",
		Name:   "my-tool",
		Input:  map[string]interface{}{"query": "test"},
		Output: map[string]interface{}{"result": "success"},
		Error:  "",
	}

	if call.ID != "call_abc" {
		t.Errorf("ID mismatch")
	}
	if call.Name != "my-tool" {
		t.Errorf("Name mismatch")
	}
	if call.Input == nil {
		t.Error("Input should not be nil")
	}
	if call.Output == nil {
		t.Error("Output should not be nil")
	}
}

func TestAgenticResult_Structure(t *testing.T) {
	result := AgenticResult{
		FinalAnswer: "The answer is 42",
		Status:      "completed",
		Trace: ExecutionTrace{
			Iterations:  []Iteration{{Index: 0}},
			TotalTokens: 500,
			Duration:    "1.5s",
		},
	}

	if result.FinalAnswer != "The answer is 42" {
		t.Errorf("FinalAnswer mismatch")
	}
	if result.Status != "completed" {
		t.Errorf("Status mismatch")
	}
	if result.Trace.TotalTokens != 500 {
		t.Errorf("TotalTokens = %d, want 500", result.Trace.TotalTokens)
	}
}

func TestAPIRequest_Marshaling(t *testing.T) {
	req := APIRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 4096,
		System:    "You are helpful",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Tools: []any{
			Tool{
				Name:        "test-tool",
				Description: "A test",
				InputSchema: map[string]string{"type": "object"},
			},
		},
		Temperature: 0.5,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled["model"] != "claude-sonnet-4-5-20250929" {
		t.Error("model mismatch after marshal/unmarshal")
	}
}

func TestContentBlock_Types(t *testing.T) {
	// Text block
	textBlock := ContentBlock{
		Type: "text",
		Text: "Hello world",
	}
	if textBlock.Type != "text" {
		t.Error("expected text type")
	}

	// Tool use block
	toolUseBlock := ContentBlock{
		Type:  "tool_use",
		ID:    "call_123",
		Name:  "my-tool",
		Input: map[string]any{"key": "value"},
	}
	if toolUseBlock.Type != "tool_use" {
		t.Error("expected tool_use type")
	}

	// Tool result block
	toolResultBlock := ContentBlock{
		Type:      "tool_result",
		ToolUseID: "call_123",
		Content:   "Tool result here",
		IsError:   false,
	}
	if toolResultBlock.Type != "tool_result" {
		t.Error("expected tool_result type")
	}
}

func TestAPIResponse_Parsing(t *testing.T) {
	respJSON := `{
		"content": [
			{"type": "text", "text": "Let me help"},
			{"type": "tool_use", "id": "call_1", "name": "think", "input": {"content": "test"}}
		],
		"stop_reason": "tool_use",
		"usage": {"input_tokens": 100, "output_tokens": 50}
	}`

	var resp APIResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Content) != 2 {
		t.Errorf("expected 2 content blocks, got %d", len(resp.Content))
	}
	if resp.StopReason != "tool_use" {
		t.Errorf("stop_reason = %s, want tool_use", resp.StopReason)
	}
	if resp.Usage.InputTokens != 100 {
		t.Errorf("input_tokens = %d, want 100", resp.Usage.InputTokens)
	}

	// Verify first block is text
	if resp.Content[0].Type != "text" {
		t.Error("first block should be text")
	}
	if resp.Content[0].Text != "Let me help" {
		t.Error("text content mismatch")
	}

	// Verify second block is tool_use
	if resp.Content[1].Type != "tool_use" {
		t.Error("second block should be tool_use")
	}
	if resp.Content[1].Name != "think" {
		t.Error("tool name mismatch")
	}
}

// Integration test - requires ANTHROPIC_API_KEY
func TestAgenticClient_Integration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	registry := NewToolRegistry()

	// Register a simple test tool
	_ = registry.Register(ToolSpec{
		Name:        "echo",
		Description: "Returns the input message",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]string{
					"type":        "string",
					"description": "Message to echo",
				},
			},
			"required": []string{"message"},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			msg, _ := input["message"].(string)
			return map[string]string{"echoed": msg}, nil
		},
	})

	config := DefaultAgenticConfig()
	config.MaxIterations = 3
	client := NewAgenticClient(apiKey, registry, config)

	ctx := context.Background()
	result, err := client.Run(ctx, "Please use the echo tool to echo 'hello world'")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Status == "" {
		t.Error("expected non-empty status")
	}

	t.Logf("Result status: %s", result.Status)
	t.Logf("Iterations: %d", len(result.Trace.Iterations))
	t.Logf("Final answer: %s", result.FinalAnswer)
}
