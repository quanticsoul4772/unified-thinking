// Package benchmarks provides problem execution via the unified-thinking server.
package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// DirectExecutor executes problems directly against storage (for testing)
type DirectExecutor struct {
	storage storage.Storage
}

// NewDirectExecutor creates a new direct executor
func NewDirectExecutor(store storage.Storage) *DirectExecutor {
	return &DirectExecutor{storage: store}
}

// Execute runs a problem and evaluates the result
func (e *DirectExecutor) Execute(problem *Problem, evaluator Evaluator) (*Result, error) {
	start := time.Now()

	// Extract input parameters
	content, ok := problem.Input["content"].(string)
	if !ok {
		content = problem.Description
	}

	mode, ok := problem.Input["mode"].(string)
	if !ok {
		mode = "auto"
	}

	// Create thought using the thinking system
	thought := types.NewThought().
		Content(content).
		Mode(types.ThinkingMode(mode)).
		Confidence(0.8).
		Build()

	// Store the thought
	if err := e.storage.StoreThought(thought); err != nil {
		return nil, fmt.Errorf("failed to store thought: %w", err)
	}

	latency := time.Since(start)

	// Evaluate response
	// For now, extract conclusion from thought content
	response := thought.Content
	correct, score, err := evaluator.Evaluate(response, problem.Expected)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	result := &Result{
		ProblemID:  problem.ID,
		Correct:    correct,
		Score:      score,
		Confidence: thought.Confidence,
		Latency:    latency,
		Mode:       string(thought.Mode),
		Response:   response,
	}

	return result, nil
}

// MCPExecutor executes problems via MCP protocol (for integration testing)
// This is a placeholder for future stdio/SSE MCP client integration
type MCPExecutor struct {
	serverPath string
}

// NewMCPExecutor creates a new MCP executor
func NewMCPExecutor(serverPath string) *MCPExecutor {
	return &MCPExecutor{serverPath: serverPath}
}

// Execute runs a problem via MCP protocol
func (e *MCPExecutor) Execute(problem *Problem, evaluator Evaluator) (*Result, error) {
	start := time.Now()

	// TODO: Implement actual MCP communication via stdio
	// For Phase 1, we'll use the DirectExecutor
	// Phase 2 will add proper MCP client integration

	// Placeholder response
	ctx := context.Background()
	_ = ctx

	// Create MCP tool call request
	toolCall := map[string]interface{}{
		"tool": "think",
		"args": map[string]interface{}{
			"content": problem.Description,
			"mode":    problem.Input["mode"],
		},
	}

	_ = toolCall

	latency := time.Since(start)

	// For now, return a placeholder
	result := &Result{
		ProblemID:  problem.ID,
		Correct:    false,
		Score:      0.0,
		Confidence: 0.0,
		Latency:    latency,
		Mode:       "auto",
		Response:   "",
		Error:      "MCP executor not yet implemented",
	}

	return result, nil
}

// MockExecutor provides deterministic results for testing the framework itself
type MockExecutor struct {
	results map[string]*Result
}

// NewMockExecutor creates a mock executor with predefined results
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		results: make(map[string]*Result),
	}
}

// AddResult adds a predefined result for a problem ID
func (e *MockExecutor) AddResult(problemID string, result *Result) {
	e.results[problemID] = result
}

// Execute returns the predefined result for a problem
func (e *MockExecutor) Execute(problem *Problem, evaluator Evaluator) (*Result, error) {
	if result, exists := e.results[problem.ID]; exists {
		return result, nil
	}

	// Default result if not predefined
	return &Result{
		ProblemID:  problem.ID,
		Correct:    false,
		Score:      0.0,
		Confidence: 0.5,
		Latency:    10 * time.Millisecond,
		Mode:       "auto",
		Response:   "mock response",
	}, nil
}

// ResultToJSON converts a result to JSON string
func ResultToJSON(result *Result) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RunToJSON converts a benchmark run to JSON string
func RunToJSON(run *BenchmarkRun) (string, error) {
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
