// Package modes - Tests for ToolRegistry
package modes

import (
	"context"
	"fmt"
	"testing"
)

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(registry.tools) != 0 {
		t.Errorf("expected empty tools map, got %d", len(registry.tools))
	}
}

func TestToolRegistry_Register(t *testing.T) {
	registry := NewToolRegistry()

	// Test valid registration
	err := registry.Register(ToolSpec{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{"type": "object"},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test duplicate registration (should succeed/overwrite)
	err = registry.Register(ToolSpec{
		Name:        "test-tool",
		Description: "Updated description",
		InputSchema: map[string]interface{}{"type": "object"},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			return "new result", nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error on re-registration: %v", err)
	}

	// Test empty name
	err = registry.Register(ToolSpec{
		Name: "",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	})
	if err == nil {
		t.Error("expected error for empty name")
	}

	// Test nil handler
	err = registry.Register(ToolSpec{
		Name:    "no-handler",
		Handler: nil,
	})
	if err == nil {
		t.Error("expected error for nil handler")
	}
}

func TestToolRegistry_Get(t *testing.T) {
	registry := NewToolRegistry()

	_ = registry.Register(ToolSpec{
		Name:        "get-test",
		Description: "Test for Get",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	})

	// Test successful get
	spec, ok := registry.Get("get-test")
	if !ok {
		t.Fatal("expected to find tool")
	}
	if spec.Name != "get-test" {
		t.Errorf("got name %q, want %q", spec.Name, "get-test")
	}

	// Test non-existent
	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent tool")
	}
}

func TestToolRegistry_Unregister(t *testing.T) {
	registry := NewToolRegistry()

	_ = registry.Register(ToolSpec{
		Name:    "to-remove",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) { return nil, nil },
	})

	// Verify exists
	_, ok := registry.Get("to-remove")
	if !ok {
		t.Fatal("tool should exist before unregister")
	}

	// Unregister
	registry.Unregister("to-remove")

	// Verify removed
	_, ok = registry.Get("to-remove")
	if ok {
		t.Error("tool should not exist after unregister")
	}

	// Unregister non-existent (should not panic)
	registry.Unregister("nonexistent")
}

func TestToolRegistry_List(t *testing.T) {
	registry := NewToolRegistry()

	// Empty list
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("expected empty list, got %d", len(names))
	}

	// Add tools
	_ = registry.Register(ToolSpec{
		Name:    "tool-a",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) { return nil, nil },
	})
	_ = registry.Register(ToolSpec{
		Name:    "tool-b",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) { return nil, nil },
	})

	names = registry.List()
	if len(names) != 2 {
		t.Errorf("expected 2 tools, got %d", len(names))
	}

	// Verify names present
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if !nameSet["tool-a"] || !nameSet["tool-b"] {
		t.Error("expected both tools in list")
	}
}

func TestToolRegistry_Execute(t *testing.T) {
	registry := NewToolRegistry()

	_ = registry.Register(ToolSpec{
		Name: "echo",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			return input["message"], nil
		},
	})

	_ = registry.Register(ToolSpec{
		Name: "error-tool",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("intentional error")
		},
	})

	ctx := context.Background()

	// Test successful execution
	result, err := registry.Execute(ctx, "echo", map[string]interface{}{"message": "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("got %v, want %q", result, "hello")
	}

	// Test error tool
	_, err = registry.Execute(ctx, "error-tool", nil)
	if err == nil {
		t.Error("expected error from error-tool")
	}

	// Test unknown tool
	_, err = registry.Execute(ctx, "unknown", nil)
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestToolRegistry_GetToolsForClaude(t *testing.T) {
	registry := NewToolRegistry()

	_ = registry.Register(ToolSpec{
		Name:        "claude-tool",
		Description: "A tool for Claude",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]string{"type": "string"},
			},
		},
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) { return nil, nil },
	})

	tools := registry.GetToolsForClaude()
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	tool := tools[0]
	if tool["name"] != "claude-tool" {
		t.Errorf("got name %v, want claude-tool", tool["name"])
	}
	if tool["description"] != "A tool for Claude" {
		t.Errorf("got description %v, want 'A tool for Claude'", tool["description"])
	}
	if tool["input_schema"] == nil {
		t.Error("expected input_schema to be present")
	}
}

func TestToolRegistry_CreateFilteredRegistry(t *testing.T) {
	registry := NewToolRegistry()

	_ = registry.Register(ToolSpec{
		Name:    "tool-1",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) { return nil, nil },
	})
	_ = registry.Register(ToolSpec{
		Name:    "tool-2",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) { return nil, nil },
	})
	_ = registry.Register(ToolSpec{
		Name:    "tool-3",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) { return nil, nil },
	})

	// Filter to subset
	filtered := registry.CreateFilteredRegistry([]string{"tool-1", "tool-3"})
	names := filtered.List()
	if len(names) != 2 {
		t.Fatalf("expected 2 tools in filtered registry, got %d", len(names))
	}

	// Verify correct tools present
	_, ok1 := filtered.Get("tool-1")
	_, ok2 := filtered.Get("tool-2")
	_, ok3 := filtered.Get("tool-3")

	if !ok1 {
		t.Error("expected tool-1 in filtered registry")
	}
	if ok2 {
		t.Error("tool-2 should not be in filtered registry")
	}
	if !ok3 {
		t.Error("expected tool-3 in filtered registry")
	}

	// Empty filter returns all
	allTools := registry.CreateFilteredRegistry(nil)
	if len(allTools.List()) != 3 {
		t.Errorf("expected 3 tools when no filter, got %d", len(allTools.List()))
	}
}

func TestToolRegistry_Clone(t *testing.T) {
	registry := NewToolRegistry()

	_ = registry.Register(ToolSpec{
		Name:    "original",
		Handler: func(ctx context.Context, input map[string]interface{}) (interface{}, error) { return "original", nil },
	})

	clone := registry.Clone()

	// Verify clone has same tool
	_, ok := clone.Get("original")
	if !ok {
		t.Error("expected clone to have original tool")
	}

	// Modify original, clone should not change
	registry.Unregister("original")

	_, ok = clone.Get("original")
	if !ok {
		t.Error("clone should still have tool after original modified")
	}
}

func TestSafeToolSubset(t *testing.T) {
	// Verify SafeToolSubset is not empty
	if len(SafeToolSubset) == 0 {
		t.Error("SafeToolSubset should not be empty")
	}

	// Verify excluded tools are not in safe subset
	safeSet := make(map[string]bool)
	for _, name := range SafeToolSubset {
		safeSet[name] = true
	}

	for _, excluded := range ExcludedTools {
		if safeSet[excluded] {
			t.Errorf("excluded tool %q should not be in safe subset", excluded)
		}
	}

	// Verify some expected tools are present
	expectedSafe := []string{"think", "detect-biases", "detect-fallacies", "decompose-problem"}
	for _, name := range expectedSafe {
		if !safeSet[name] {
			t.Errorf("expected %q to be in safe subset", name)
		}
	}
}

func TestBuildSchemaFromStruct(t *testing.T) {
	schema := BuildSchemaFromStruct(
		"Test schema",
		map[string]PropertyDef{
			"query": {Type: "string", Description: "The search query"},
			"limit": {Type: "integer", Description: "Max results", Default: 10},
		},
		[]string{"query"},
	)

	if schema["type"] != "object" {
		t.Errorf("expected type object, got %v", schema["type"])
	}
	if schema["description"] != "Test schema" {
		t.Errorf("expected description 'Test schema', got %v", schema["description"])
	}

	required, ok := schema["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "query" {
		t.Errorf("expected required [query], got %v", schema["required"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties map")
	}
	if len(props) != 2 {
		t.Errorf("expected 2 properties, got %d", len(props))
	}
}

func TestToJSON(t *testing.T) {
	// Test object
	result := ToJSON(map[string]string{"key": "value"})
	if result != `{"key":"value"}` {
		t.Errorf("unexpected JSON: %s", result)
	}

	// Test array
	result = ToJSON([]int{1, 2, 3})
	if result != `[1,2,3]` {
		t.Errorf("unexpected JSON: %s", result)
	}

	// Test string
	result = ToJSON("hello")
	if result != `"hello"` {
		t.Errorf("unexpected JSON: %s", result)
	}
}
