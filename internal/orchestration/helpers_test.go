package orchestration

import (
	"context"
	"fmt"
	"time"
)

// Test helper functions and mocks

// mockExecutor is a simple mock for testing
type mockExecutor struct {
	calls []mockCall
}

type mockCall struct {
	toolName string
	input    map[string]interface{}
}

func (m *mockExecutor) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	m.calls = append(m.calls, mockCall{toolName: toolName, input: input})
	return map[string]interface{}{
		"id":         "mock-result-1",
		"confidence": 0.9,
	}, nil
}

// mockExecutorWithError simulates tool execution failures
type mockExecutorWithError struct {
	shouldFail bool
	failOn     string
}

func (m *mockExecutorWithError) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	if m.shouldFail && (m.failOn == "" || m.failOn == toolName) {
		return nil, fmt.Errorf("mock error for tool %s", toolName)
	}
	return map[string]interface{}{
		"id":         "result-" + toolName,
		"confidence": 0.85,
	}, nil
}

// mockToolSpecificExecutor returns tool-specific results for testing context updates
type mockToolSpecificExecutor struct{}

func (m *mockToolSpecificExecutor) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "think":
		return map[string]interface{}{
			"thought_id": "thought-123",
			"confidence": 0.9,
		}, nil
	case "build-causal-graph":
		return map[string]interface{}{
			"id":         "graph-456",
			"confidence": 0.85,
		}, nil
	case "probabilistic-reasoning":
		return map[string]interface{}{
			"id":          "belief-789",
			"probability": 0.75,
		}, nil
	case "assess-evidence":
		return map[string]interface{}{
			"id":    "evidence-101",
			"score": 0.8,
		}, nil
	case "make-decision":
		return map[string]interface{}{
			"id":         "decision-202",
			"confidence": 0.95,
		}, nil
	default:
		return map[string]interface{}{
			"id":         "result-" + toolName,
			"confidence": 0.8,
		}, nil
	}
}

// mockConfidenceExecutor returns predefined confidence values
type mockConfidenceExecutor struct {
	confidences []float64
	callCount   int
}

func (m *mockConfidenceExecutor) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	confidence := 0.5
	if m.callCount < len(m.confidences) {
		confidence = m.confidences[m.callCount]
	}
	m.callCount++

	return map[string]interface{}{
		"id":         "result",
		"confidence": confidence,
	}, nil
}

// mockDelayedExecutor simulates slow tool execution
type mockDelayedExecutor struct {
	delay time.Duration
}

func (m *mockDelayedExecutor) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	time.Sleep(m.delay)
	return map[string]interface{}{
		"id":         "result-" + toolName,
		"confidence": 0.85,
	}, nil
}
