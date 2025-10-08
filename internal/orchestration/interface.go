package orchestration

import "context"

// ToolExecutor provides access to all tool handlers for workflow execution
type ToolExecutor interface {
	// ExecuteTool runs a tool by name with the given input and returns the result
	ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error)
}
