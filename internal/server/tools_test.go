package server

import (
	"testing"
)

func TestGetToolByName_ExistingTool(t *testing.T) {
	// Test with known tool names
	knownTools := []string{
		"think",
		"history",
		"list-branches",
		"validate",
		"prove",
		"search",
		"get-metrics",
	}

	for _, name := range knownTools {
		tool, found := GetToolByName(name)
		if !found {
			t.Errorf("GetToolByName(%q) should find the tool", name)
			continue
		}
		if tool == nil {
			t.Errorf("GetToolByName(%q) returned nil tool but found=true", name)
			continue
		}
		if tool.Name != name {
			t.Errorf("GetToolByName(%q) returned tool with name %q", name, tool.Name)
		}
	}
}

func TestGetToolByName_NonExistentTool(t *testing.T) {
	tool, found := GetToolByName("non-existent-tool-xyz")
	if found {
		t.Error("GetToolByName should return false for non-existent tool")
	}
	if tool != nil {
		t.Error("GetToolByName should return nil for non-existent tool")
	}
}

func TestGetToolByName_EmptyName(t *testing.T) {
	tool, found := GetToolByName("")
	if found {
		t.Error("GetToolByName should return false for empty name")
	}
	if tool != nil {
		t.Error("GetToolByName should return nil for empty name")
	}
}

func TestGetToolNames(t *testing.T) {
	names := GetToolNames()

	if len(names) == 0 {
		t.Fatal("GetToolNames should return at least one tool name")
	}

	// Verify count matches ToolDefinitions
	if len(names) != len(ToolDefinitions) {
		t.Errorf("GetToolNames returned %d names, but ToolDefinitions has %d entries",
			len(names), len(ToolDefinitions))
	}
}

func TestGetToolNames_ContainsKnownTools(t *testing.T) {
	names := GetToolNames()

	knownTools := []string{
		"think",
		"history",
		"list-branches",
		"focus-branch",
		"branch-history",
		"recent-branches",
		"validate",
		"prove",
		"check-syntax",
		"search",
		"get-metrics",
	}

	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, expected := range knownTools {
		if !nameMap[expected] {
			t.Errorf("GetToolNames should include %q", expected)
		}
	}
}

func TestGetToolNames_NoDuplicates(t *testing.T) {
	names := GetToolNames()

	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			t.Errorf("GetToolNames contains duplicate: %q", name)
		}
		seen[name] = true
	}
}

func TestGetToolNames_AllNonEmpty(t *testing.T) {
	names := GetToolNames()

	for i, name := range names {
		if name == "" {
			t.Errorf("GetToolNames[%d] is empty", i)
		}
	}
}

func TestToolDefinitions_HasRequiredFields(t *testing.T) {
	for i, tool := range ToolDefinitions {
		if tool.Name == "" {
			t.Errorf("ToolDefinitions[%d]: Name should not be empty", i)
		}
		if tool.Description == "" {
			t.Errorf("ToolDefinitions[%d] (%s): Description should not be empty", i, tool.Name)
		}
	}
}

func TestToolDefinitions_Consistency(t *testing.T) {
	// Verify GetToolByName returns tools from ToolDefinitions
	for _, tool := range ToolDefinitions {
		found, ok := GetToolByName(tool.Name)
		if !ok {
			t.Errorf("GetToolByName should find tool %q from ToolDefinitions", tool.Name)
			continue
		}
		if found.Description != tool.Description {
			t.Errorf("GetToolByName(%q) description mismatch", tool.Name)
		}
	}
}
