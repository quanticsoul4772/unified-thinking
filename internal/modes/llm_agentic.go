// Package modes - AgenticClient for programmatic tool calling
package modes

import (
	"context"
	"fmt"
	"os"
	"time"
)

// AgenticConfig configures the agentic client
type AgenticConfig struct {
	MaxIterations   int
	MaxToolsPerTurn int
	StopOnError     bool
	Temperature     float64
	Model           string
	MaxTokens       int
}

// DefaultAgenticConfig returns sensible defaults
func DefaultAgenticConfig() AgenticConfig {
	return AgenticConfig{
		MaxIterations:   10,
		MaxToolsPerTurn: 5,
		StopOnError:     true,
		Temperature:     0.3,
		Model:           "claude-sonnet-4-5-20250929",
		MaxTokens:       4096,
	}
}

// AgenticClient wraps Claude API with tool-calling loop
type AgenticClient struct {
	*AnthropicBaseClient
	registry *ToolRegistry
	config   AgenticConfig
}

// NewAgenticClient creates an agentic client
func NewAgenticClient(apiKey string, registry *ToolRegistry, config AgenticConfig) *AgenticClient {
	if config.MaxIterations <= 0 {
		config.MaxIterations = 10
	}
	if config.MaxToolsPerTurn <= 0 {
		config.MaxToolsPerTurn = 5
	}
	if config.Model == "" {
		config.Model = "claude-sonnet-4-5-20250929"
	}
	if config.MaxTokens <= 0 {
		config.MaxTokens = 4096
	}

	return &AgenticClient{
		AnthropicBaseClient: NewAnthropicBaseClient(BaseClientConfig{
			APIKey:      apiKey,
			Model:       config.Model,
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
		}),
		registry: registry,
		config:   config,
	}
}

// NewAgenticClientFromEnv creates an agentic client using environment variables
func NewAgenticClientFromEnv(registry *ToolRegistry) (*AgenticClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	config := DefaultAgenticConfig()
	if model := os.Getenv("AGENT_MODEL"); model != "" {
		config.Model = model
	}

	return NewAgenticClient(apiKey, registry, config), nil
}

// ExecutionTrace records the agentic execution history
type ExecutionTrace struct {
	Iterations  []Iteration `json:"iterations"`
	TotalTokens int         `json:"total_tokens"`
	Duration    string      `json:"duration"`
}

// Iteration represents one round of thought + tool calls
type Iteration struct {
	Index      int        `json:"index"`
	Thought    string     `json:"thought,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolErrors []string   `json:"tool_errors,omitempty"`
}

// ToolCall represents a single tool invocation
type ToolCall struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Input  any    `json:"input"`
	Output any    `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// AgenticResult is the final result of agentic execution
type AgenticResult struct {
	FinalAnswer string         `json:"final_answer"`
	Trace       ExecutionTrace `json:"trace"`
	Status      string         `json:"status"` // "completed", "max_iterations", "error"
}

// Run executes an agentic task with tool use
func (a *AgenticClient) Run(ctx context.Context, task string) (*AgenticResult, error) {
	return a.RunWithSystem(ctx, task, "")
}

// RunWithSystem executes an agentic task with a custom system prompt
func (a *AgenticClient) RunWithSystem(ctx context.Context, task, systemPrompt string) (*AgenticResult, error) {
	startTime := time.Now()
	trace := ExecutionTrace{
		Iterations: make([]Iteration, 0),
	}

	if systemPrompt == "" {
		systemPrompt = defaultAgentSystemPrompt
	}

	messages := []Message{NewTextMessage("user", task)}
	tools := a.registry.GetToolsForClaude()

	// Convert to Tool slice
	var toolDefs []Tool
	for _, t := range tools {
		toolDefs = append(toolDefs, Tool{
			Name:        t["name"].(string),
			Description: t["description"].(string),
			InputSchema: t["input_schema"],
		})
	}

	for i := 0; i < a.config.MaxIterations; i++ {
		req := &APIRequest{
			Model:       a.Model(),
			MaxTokens:   a.MaxTokens(),
			System:      systemPrompt,
			Messages:    messages,
			Tools:       toolDefs,
			Temperature: a.Temperature(),
		}

		resp, err := a.SendRequest(ctx, req)
		if err != nil {
			return &AgenticResult{
				FinalAnswer: fmt.Sprintf("API error: %v", err),
				Trace:       trace,
				Status:      "error",
			}, nil
		}

		trace.TotalTokens += resp.Usage.InputTokens + resp.Usage.OutputTokens
		iteration := Iteration{Index: i}

		var toolCalls []ResponseBlock
		var assistantBlocks []ContentBlock

		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				iteration.Thought = block.Text
				assistantBlocks = append(assistantBlocks, ContentBlock{
					Type: "text",
					Text: block.Text,
				})
			case "tool_use":
				toolCalls = append(toolCalls, block)
				assistantBlocks = append(assistantBlocks, ContentBlock{
					Type:  "tool_use",
					ID:    block.ID,
					Name:  block.Name,
					Input: block.Input,
				})
			}
		}

		// If no tool calls, we're done
		if len(toolCalls) == 0 || resp.StopReason == "end_turn" {
			trace.Iterations = append(trace.Iterations, iteration)
			trace.Duration = time.Since(startTime).String()
			return &AgenticResult{
				FinalAnswer: iteration.Thought,
				Trace:       trace,
				Status:      "completed",
			}, nil
		}

		// Execute tool calls (up to MaxToolsPerTurn)
		var toolResults []ContentBlock
		for j, tc := range toolCalls {
			if j >= a.config.MaxToolsPerTurn {
				break
			}

			call := ToolCall{
				ID:    tc.ID,
				Name:  tc.Name,
				Input: tc.Input,
			}

			result, err := a.registry.Execute(ctx, tc.Name, tc.Input)
			if err != nil {
				call.Error = err.Error()
				iteration.ToolErrors = append(iteration.ToolErrors, err.Error())

				if a.config.StopOnError {
					trace.Iterations = append(trace.Iterations, iteration)
					trace.Duration = time.Since(startTime).String()
					return &AgenticResult{
						FinalAnswer: fmt.Sprintf("Tool error in %s: %v", tc.Name, err),
						Trace:       trace,
						Status:      "error",
					}, nil
				}

				toolResults = append(toolResults, ContentBlock{
					Type:      "tool_result",
					ToolUseID: tc.ID,
					IsError:   true,
					Content:   err.Error(),
				})
			} else {
				call.Output = result
				toolResults = append(toolResults, ContentBlock{
					Type:      "tool_result",
					ToolUseID: tc.ID,
					Content:   ToJSON(result),
				})
			}

			iteration.ToolCalls = append(iteration.ToolCalls, call)
		}

		trace.Iterations = append(trace.Iterations, iteration)

		// Add assistant message and tool results to continue conversation
		messages = append(messages, NewBlockMessage("assistant", assistantBlocks))
		messages = append(messages, NewBlockMessage("user", toolResults))
	}

	trace.Duration = time.Since(startTime).String()
	return &AgenticResult{
		FinalAnswer: "Reached maximum iterations without completing task",
		Trace:       trace,
		Status:      "max_iterations",
	}, nil
}

const defaultAgentSystemPrompt = `You are an intelligent reasoning agent with access to specialized cognitive tools. Your goal is to help the user by using these tools effectively.

Guidelines:
1. Break down complex problems into smaller steps
2. Use appropriate tools for each step
3. Synthesize results into a coherent answer
4. Be thorough but efficient - don't use tools unnecessarily
5. If you encounter errors, explain what went wrong and try alternative approaches

Available tool categories:
- Reasoning: think, decompose-problem, make-decision
- Analysis: analyze-perspectives, detect-biases, detect-fallacies
- Evidence: assess-evidence, probabilistic-reasoning
- Causal: build-causal-graph, generate-hypotheses
- Synthesis: synthesize-insights, detect-emergent-patterns
- Search: search-similar-thoughts, search-knowledge-graph

When you have enough information to answer, provide a clear, comprehensive response.`

// ToolsUsed returns a deduplicated list of tools used in the trace
func (t *ExecutionTrace) ToolsUsed() []string {
	used := make(map[string]bool)
	for _, iter := range t.Iterations {
		for _, tc := range iter.ToolCalls {
			used[tc.Name] = true
		}
	}

	names := make([]string, 0, len(used))
	for name := range used {
		names = append(names, name)
	}
	return names
}

// TotalToolCalls returns the total number of tool calls made
func (t *ExecutionTrace) TotalToolCalls() int {
	count := 0
	for _, iter := range t.Iterations {
		count += len(iter.ToolCalls)
	}
	return count
}

// ErrorCount returns the number of tool errors encountered
func (t *ExecutionTrace) ErrorCount() int {
	count := 0
	for _, iter := range t.Iterations {
		count += len(iter.ToolErrors)
	}
	return count
}

