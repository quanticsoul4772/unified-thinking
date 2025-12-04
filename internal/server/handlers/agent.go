// Package handlers - Agent MCP tool handler for programmatic tool calling
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
)

// AgentHandler handles agentic tool execution
type AgentHandler struct {
	registry *modes.ToolRegistry
	apiKey   string
	enabled  bool
}

// NewAgentHandler creates a new agent handler
// The handler is disabled if AGENT_ENABLED is not "true" or ANTHROPIC_API_KEY is not set
func NewAgentHandler(registry *modes.ToolRegistry) *AgentHandler {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	enabled := os.Getenv("AGENT_ENABLED") == "true" && apiKey != ""

	return &AgentHandler{
		registry: registry,
		apiKey:   apiKey,
		enabled:  enabled,
	}
}

// RunAgentRequest for the run-agent tool
type RunAgentRequest struct {
	Task          string   `json:"task"`                      // The task for the agent to complete
	MaxIterations int      `json:"max_iterations,omitempty"`  // Maximum iterations (default: 10)
	AllowedTools  []string `json:"allowed_tools,omitempty"`   // Restrict to specific tools
	StopOnError   *bool    `json:"stop_on_error,omitempty"`   // Stop on tool error (default: true)
	SystemPrompt  string   `json:"system_prompt,omitempty"`   // Custom system prompt
	Temperature   float64  `json:"temperature,omitempty"`     // Model temperature (default: 0.3)
}

// RunAgentResponse contains agentic execution results
type RunAgentResponse struct {
	FinalAnswer    string                  `json:"final_answer"`
	Status         string                  `json:"status"`         // "completed", "max_iterations", "error"
	Iterations     int                     `json:"iterations"`
	ToolsUsed      []string                `json:"tools_used"`
	TotalToolCalls int                     `json:"total_tool_calls"`
	ErrorCount     int                     `json:"error_count"`
	Trace          *modes.ExecutionTrace   `json:"trace,omitempty"`
}

// HandleRunAgent executes an agentic task
func (h *AgentHandler) HandleRunAgent(ctx context.Context, req *mcp.CallToolRequest, input RunAgentRequest) (*mcp.CallToolResult, *RunAgentResponse, error) {
	// Check if agent is enabled
	if !h.enabled {
		return nil, nil, fmt.Errorf("run-agent requires AGENT_ENABLED=true and ANTHROPIC_API_KEY")
	}

	// Validate input
	if input.Task == "" {
		return nil, nil, fmt.Errorf("task is required")
	}

	// Build config
	config := modes.DefaultAgenticConfig()
	if input.MaxIterations > 0 {
		config.MaxIterations = input.MaxIterations
	}
	if input.MaxIterations > 20 {
		config.MaxIterations = 20 // Cap at 20 for safety
	}
	if input.StopOnError != nil {
		config.StopOnError = *input.StopOnError
	}
	if input.Temperature > 0 {
		config.Temperature = input.Temperature
	}

	// Create filtered registry if tools specified
	registry := h.registry
	if len(input.AllowedTools) > 0 {
		registry = h.registry.CreateFilteredRegistry(input.AllowedTools)
	}

	// Create agentic client
	agent := modes.NewAgenticClient(h.apiKey, registry, config)

	// Execute
	var result *modes.AgenticResult
	var err error
	if input.SystemPrompt != "" {
		result, err = agent.RunWithSystem(ctx, input.Task, input.SystemPrompt)
	} else {
		result, err = agent.Run(ctx, input.Task)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// Build response
	response := &RunAgentResponse{
		FinalAnswer:    result.FinalAnswer,
		Status:         result.Status,
		Iterations:     len(result.Trace.Iterations),
		ToolsUsed:      result.Trace.ToolsUsed(),
		TotalToolCalls: result.Trace.TotalToolCalls(),
		ErrorCount:     result.Trace.ErrorCount(),
		Trace:          &result.Trace,
	}

	return &mcp.CallToolResult{Content: agentToJSONContent(response)}, response, nil
}

// ListAgentToolsRequest for the list-agent-tools tool
type ListAgentToolsRequest struct {
	Category string `json:"category,omitempty"` // Filter by category
}

// ListAgentToolsResponse contains available agent tools
type ListAgentToolsResponse struct {
	Tools         []AgentToolInfo `json:"tools"`
	TotalCount    int             `json:"total_count"`
	SafeToolCount int             `json:"safe_tool_count"`
}

// AgentToolInfo describes a tool available for agent use
type AgentToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSafe      bool   `json:"is_safe"` // Whether in safe subset
}

// HandleListAgentTools lists tools available for agentic use
func (h *AgentHandler) HandleListAgentTools(ctx context.Context, req *mcp.CallToolRequest, input ListAgentToolsRequest) (*mcp.CallToolResult, *ListAgentToolsResponse, error) {
	// Get all registered tools
	toolNames := h.registry.List()

	// Build safe set
	safeSet := make(map[string]bool)
	for _, name := range modes.SafeToolSubset {
		safeSet[name] = true
	}

	// Build response
	tools := make([]AgentToolInfo, 0, len(toolNames))
	safeCount := 0
	for _, name := range toolNames {
		spec, ok := h.registry.Get(name)
		if !ok {
			continue
		}

		isSafe := safeSet[name]
		if isSafe {
			safeCount++
		}

		tools = append(tools, AgentToolInfo{
			Name:        name,
			Description: spec.Description,
			IsSafe:      isSafe,
		})
	}

	response := &ListAgentToolsResponse{
		Tools:         tools,
		TotalCount:    len(tools),
		SafeToolCount: safeCount,
	}

	return &mcp.CallToolResult{Content: agentToJSONContent(response)}, response, nil
}

// agentToJSONContent converts response to JSON content
func agentToJSONContent(data interface{}) []mcp.Content {
	jsonData, err := json.Marshal(data)
	if err != nil {
		errData := map[string]string{"error": err.Error()}
		jsonData, _ = json.Marshal(errData)
	}

	return []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}
}

// RegisterAgentTools registers agent MCP tools
func RegisterAgentTools(mcpServer *mcp.Server, handler *AgentHandler) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "run-agent",
		Description: `Execute a task using an agentic LLM with access to unified-thinking tools.

Requires AGENT_ENABLED=true and ANTHROPIC_API_KEY environment variables.

The agent can use cognitive reasoning tools to break down problems, analyze evidence,
generate hypotheses, and synthesize insights. It operates in a loop, calling tools
until the task is complete or max iterations reached.

**Parameters:**
- task (required): The task for the agent to complete
- max_iterations (optional): Maximum tool-calling iterations (default: 10, max: 20)
- allowed_tools (optional): List of specific tools to allow (default: all safe tools)
- stop_on_error (optional): Stop if a tool call fails (default: true)
- system_prompt (optional): Custom system prompt for the agent
- temperature (optional): Model temperature (default: 0.3)

**Safe Tools (available by default):**
- Reasoning: think, decompose-problem, make-decision, dual-process-think
- Analysis: analyze-perspectives, detect-biases, detect-fallacies
- Evidence: assess-evidence, probabilistic-reasoning, detect-contradictions
- Causal: build-causal-graph, generate-hypotheses, evaluate-hypotheses
- Search: search-similar-thoughts, search-knowledge-graph, retrieve-similar-cases
- Synthesis: synthesize-insights, detect-emergent-patterns
- Graph-of-Thoughts: got-generate, got-aggregate, got-refine, got-score

**Excluded Tools (not available):**
- Storage: store-entity, create-relationship (side effects)
- Sessions: export-session, import-session (session management)
- Orchestration: run-agent, run-preset, execute-workflow (recursion risk)
- State: create-checkpoint, restore-checkpoint (state modification)

**Returns:**
- final_answer: The agent's final response
- status: "completed", "max_iterations", or "error"
- iterations: Number of iterations used
- tools_used: List of tools that were called
- total_tool_calls: Total number of tool invocations
- error_count: Number of tool errors encountered
- trace: Full execution trace (iterations, thoughts, tool calls)

**Example:**
{
  "task": "Analyze the pros and cons of using microservices architecture for a startup",
  "max_iterations": 5
}

**Use Cases:**
- Complex multi-step reasoning tasks
- Research synthesis across multiple domains
- Decision analysis with multiple perspectives
- Hypothesis generation and evaluation
- Problem decomposition and systematic analysis`,
	}, handler.HandleRunAgent)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "list-agent-tools",
		Description: `List all tools available for agentic use.

Shows which tools can be used by the run-agent tool, including which
are in the "safe" subset (no side effects) and which are excluded.

**Parameters:**
- category (optional): Filter by category (not yet implemented)

**Returns:**
- tools: List of tool info (name, description, is_safe)
- total_count: Total registered tools
- safe_tool_count: Number of safe tools`,
	}, handler.HandleListAgentTools)
}
