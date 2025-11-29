// Package benchmarks provides problem execution via the unified-thinking server.
package benchmarks

import (
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
type MCPExecutor struct {
	client     *MCPClient
	serverPath string
	env        []string
}

// NewMCPExecutor creates a new MCP executor
func NewMCPExecutor(serverPath string, env []string) *MCPExecutor {
	return &MCPExecutor{
		serverPath: serverPath,
		env:        env,
	}
}

// Execute runs a problem via MCP protocol
func (e *MCPExecutor) Execute(problem *Problem, evaluator Evaluator) (*Result, error) {
	start := time.Now()

	// Lazy initialization - reuse client for multiple problems
	if e.client == nil {
		client, err := NewMCPClient(e.serverPath, e.env)
		if err != nil {
			return &Result{
				ProblemID: problem.ID,
				Correct:   false,
				Score:     0.0,
				Latency:   time.Since(start),
				Error:     fmt.Sprintf("Failed to create MCP client: %v", err),
			}, nil
		}
		if err := client.Start(); err != nil {
			return &Result{
				ProblemID: problem.ID,
				Correct:   false,
				Score:     0.0,
				Latency:   time.Since(start),
				Error:     fmt.Sprintf("Failed to start MCP server: %v", err),
			}, nil
		}
		e.client = client
	}

	// Extract parameters
	content, ok := problem.Input["content"].(string)
	if !ok {
		content = problem.Description
	}

	mode, ok := problem.Input["mode"].(string)
	if !ok {
		mode = "auto"
	}

	// Call think tool via MCP
	args := map[string]interface{}{
		"content": content,
		"mode":    mode,
	}

	resp, err := e.client.CallTool("think", args)
	if err != nil {
		return &Result{
			ProblemID: problem.ID,
			Correct:   false,
			Score:     0.0,
			Latency:   time.Since(start),
			Error:     fmt.Sprintf("MCP call failed: %v", err),
		}, nil
	}

	latency := time.Since(start)

	// Extract thought data from MCP response
	// Data is in structuredContent field
	var thoughtID string
	var responseText string
	var confidence float64
	var thinkMode string

	if structured, ok := resp.Content["structuredContent"].(map[string]interface{}); ok {
		thoughtID, _ = structured["thought_id"].(string)
		confidence, _ = structured["confidence"].(float64)
		thinkMode, _ = structured["mode"].(string)

		// Response text is in the content array as JSON
		if contentArray, ok := resp.Content["content"].([]interface{}); ok && len(contentArray) > 0 {
			if contentItem, ok := contentArray[0].(map[string]interface{}); ok {
				responseText, _ = contentItem["text"].(string)
			}
		}
	}

	// Evaluate response
	correct, score, err := evaluator.Evaluate(responseText, problem.Expected)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	result := &Result{
		ProblemID:  problem.ID,
		Correct:    correct,
		Score:      score,
		Confidence: confidence,
		Latency:    latency,
		Mode:       thinkMode,
		Response:   responseText,
		Metadata: map[string]interface{}{
			"thought_id": thoughtID,
		},
	}

	return result, nil
}

// Close gracefully shuts down the MCP client
func (e *MCPExecutor) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
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
