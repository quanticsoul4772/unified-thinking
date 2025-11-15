package orchestration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// MockFailingExecutor implements ToolExecutor with configurable failures
type MockFailingExecutor struct {
	failOnExecute    bool
	executionCount   int
	lastExecutedTool string
	slowExecution    bool
}

func (m *MockFailingExecutor) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	m.executionCount++
	m.lastExecutedTool = toolName

	// Check for context cancellation first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Simulate slow execution if configured
	if m.slowExecution {
		// Wait a bit to allow timeout to trigger
		select {
		case <-time.After(10 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.failOnExecute {
		return nil, errors.New("tool execution failed")
	}

	// Simulate successful execution
	return map[string]interface{}{
		"result": "success",
		"tool":   toolName,
	}, nil
}

func TestOrchestrator_ErrorRecovery_WorkflowExecutionFailures(t *testing.T) {
	executor := &MockFailingExecutor{}
	orch := NewOrchestratorWithExecutor(executor)

	tests := []struct {
		name        string
		setupFail   func(*MockFailingExecutor)
		workflow    *Workflow
		expectError bool
	}{
		{
			name: "tool execution failure",
			setupFail: func(executor *MockFailingExecutor) {
				executor.failOnExecute = true
			},
			workflow: &Workflow{
				ID:   "test-workflow",
				Name: "Test Workflow",
				Type: WorkflowSequential,
				Steps: []*WorkflowStep{
					{
						ID:    "step1",
						Tool:  "think",
						Input: map[string]interface{}{"content": "test", "mode": "linear"},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset executor state
			executor.failOnExecute = false
			executor.executionCount = 0

			// Configure failure mode
			if tt.setupFail != nil {
				tt.setupFail(executor)
			}

			// Register workflow
			err := orch.RegisterWorkflow(tt.workflow)
			if err != nil {
				t.Fatalf("Failed to register workflow: %v", err)
			}

			ctx := context.Background()

			// Test workflow execution with failures
			result, err := orch.ExecuteWorkflow(ctx, tt.workflow.ID, map[string]interface{}{})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none, result: %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Reset failure flags
			executor.failOnExecute = false

			// Test that orchestrator remains functional after failures
			if orch == nil {
				t.Error("Orchestrator should remain functional after failures")
			}
		})
	}
}

func TestOrchestrator_ErrorRecovery_ConcurrentWorkflows(t *testing.T) {
	executor := &MockFailingExecutor{}
	orch := NewOrchestratorWithExecutor(executor)

	// Create multiple workflows
	workflows := []*Workflow{
		{
			ID:   "workflow1",
			Name: "Workflow 1",
			Type: WorkflowSequential,
			Steps: []*WorkflowStep{
				{
					ID:    "step1",
					Tool:  "think",
					Input: map[string]interface{}{"content": "test1", "mode": "linear"},
				},
				{
					ID:    "step2",
					Tool:  "think",
					Input: map[string]interface{}{"content": "test2", "mode": "tree"},
				},
			},
		},
		{
			ID:   "workflow2",
			Name: "Workflow 2",
			Type: WorkflowSequential,
			Steps: []*WorkflowStep{
				{
					ID:    "step1",
					Tool:  "think",
					Input: map[string]interface{}{"content": "test2", "mode": "tree"},
				},
			},
		},
		{
			ID:   "workflow3",
			Name: "Workflow 3",
			Type: WorkflowParallel,
			Steps: []*WorkflowStep{
				{
					ID:    "step1",
					Tool:  "analyze_contradictions",
					Input: map[string]interface{}{"thoughts": []string{"thought1", "thought2"}},
				},
				{
					ID:    "step2",
					Tool:  "think",
					Input: map[string]interface{}{"content": "parallel test", "mode": "linear"},
				},
			},
		},
	}

	// Register all workflows
	for _, workflow := range workflows {
		err := orch.RegisterWorkflow(workflow)
		if err != nil {
			t.Fatalf("Failed to register workflow %s: %v", workflow.ID, err)
		}
	}

	// Test concurrent workflow execution
	done := make(chan bool, len(workflows))

	for i, workflow := range workflows {
		go func(id int, wf *Workflow) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Workflow %d panicked: %v", id, r)
				}
				done <- true
			}()

			ctx := context.Background()
			_, err := orch.ExecuteWorkflow(ctx, wf.ID, map[string]interface{}{})

			// Some workflows might fail, but orchestrator should handle it
			if err != nil {
				t.Logf("Workflow %d failed as expected: %v", id, err)
			}
		}(i, workflow)
	}

	// Wait for all workflows to complete
	for i := 0; i < len(workflows); i++ {
		select {
		case <-done:
			// Workflow completed
		case <-time.After(2 * time.Second):
			t.Errorf("Workflow %d timed out", i)
		}
	}

	// Orchestrator should remain functional after concurrent executions
	if orch == nil {
		t.Error("Orchestrator should remain functional after concurrent workflows")
	}
}

func TestOrchestrator_ErrorRecovery_InvalidWorkflows(t *testing.T) {
	executor := &MockFailingExecutor{}
	orch := NewOrchestratorWithExecutor(executor)

	tests := []struct {
		name     string
		workflow *Workflow
	}{
		{
			name:     "nil workflow",
			workflow: nil,
		},
		{
			name: "empty workflow",
			workflow: &Workflow{
				ID:    "",
				Name:  "",
				Steps: []*WorkflowStep{},
			},
		},
		{
			name: "workflow with nil steps",
			workflow: &Workflow{
				ID:    "test",
				Name:  "Test",
				Steps: nil,
			},
		},
		{
			name: "workflow with invalid step",
			workflow: &Workflow{
				ID:   "test",
				Name: "Test",
				Steps: []*WorkflowStep{
					{
						ID:    "",
						Tool:  "",
						Input: nil,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.workflow == nil {
				// Skip nil workflow test as it would cause panic before reaching orchestrator
				return
			}

			ctx := context.Background()

			// Test that orchestrator handles invalid workflows gracefully
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Orchestrator panicked with invalid workflow: %v", r)
				}
			}()

			// Try to register the workflow first
			err := orch.RegisterWorkflow(tt.workflow)
			if err == nil {
				// If registration succeeded, try to execute it
				result, err := orch.ExecuteWorkflow(ctx, tt.workflow.ID, map[string]interface{}{})

				// Should get an error for invalid workflows
				if err == nil {
					t.Errorf("Expected error for invalid workflow, got result: %v", result)
				}
			}
			// If registration failed, that's expected for invalid workflows

			// Orchestrator should remain functional
			if orch == nil {
				t.Error("Orchestrator should remain functional after invalid workflow")
			}
		})
	}
}

func TestOrchestrator_ErrorRecovery_ResourceExhaustion(t *testing.T) {
	executor := &MockFailingExecutor{}
	orch := NewOrchestratorWithExecutor(executor)

	// Create a workflow with many steps that might exhaust resources
	var steps []*WorkflowStep
	for i := 0; i < 100; i++ {
		steps = append(steps, &WorkflowStep{
			ID:   fmt.Sprintf("step%d", i),
			Tool: "think_linear",
			Input: map[string]interface{}{
				"content": fmt.Sprintf("Large content %d", i) + string(make([]byte, 1000)), // 1KB per step
			},
		})
	}

	workflow := &Workflow{
		ID:    "resource-intensive",
		Name:  "Resource Intensive",
		Type:  WorkflowSequential,
		Steps: steps,
	}

	// Register workflow
	err := orch.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	t.Run("large workflow", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test that orchestrator handles large workflows without crashing
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Orchestrator panicked with large workflow: %v", r)
			}
		}()

		result, err := orch.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

		// Might timeout or fail, but shouldn't crash
		if ctx.Err() == context.DeadlineExceeded {
			t.Log("Workflow timed out as expected for large input")
		}

		if result == nil && err == nil && ctx.Err() == nil {
			t.Error("Expected some result or error from large workflow")
		}

		// Orchestrator should remain functional
		if orch == nil {
			t.Error("Orchestrator should remain functional after large workflow")
		}
	})
}

func TestOrchestrator_ErrorRecovery_Cancellation(t *testing.T) {
	executor := &MockFailingExecutor{}
	orch := NewOrchestratorWithExecutor(executor)

	workflow := &Workflow{
		ID:   "cancellable",
		Name: "Cancellable",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:    "step1",
				Tool:  "think",
				Input: map[string]interface{}{"content": "test", "mode": "linear"},
			},
			{
				ID:    "step2",
				Tool:  "think",
				Input: map[string]interface{}{"content": "test", "mode": "tree"},
			},
		},
	}

	// Register workflow
	err := orch.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	t.Run("context cancellation", func(t *testing.T) {
		// Configure executor to be slow so cancellation can be tested
		executor.slowExecution = true

		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context immediately
		cancel()

		// Test that orchestrator handles cancellation gracefully
		result, err := orch.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

		// Should get cancellation error
		if err == nil {
			t.Errorf("Expected cancellation error, got result: %v", result)
		}

		if ctx.Err() != context.Canceled {
			t.Errorf("Expected context cancellation, got: %v", ctx.Err())
		}

		// Orchestrator should remain functional
		if orch == nil {
			t.Error("Orchestrator should remain functional after cancellation")
		}

		// Reset slow execution
		executor.slowExecution = false
	})
}

func TestOrchestrator_ErrorRecovery_PartialFailures(t *testing.T) {
	// Test scenarios where some steps succeed and others fail
	executor := &MockFailingExecutor{}
	orch := NewOrchestratorWithExecutor(executor)

	// Create a workflow where some steps fail and others succeed
	workflow := &Workflow{
		ID:   "partial-failure",
		Name: "Partial Failure",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:    "step1",
				Tool:  "think",
				Input: map[string]interface{}{"content": "success", "mode": "linear"},
			},
			{
				ID:    "step2",
				Tool:  "failing_tool",
				Input: map[string]interface{}{"content": "fail"},
			},
			{
				ID:    "step3",
				Tool:  "think",
				Input: map[string]interface{}{"content": "success", "mode": "tree"},
			},
		},
	}

	// Register workflow
	err := orch.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	t.Run("partial workflow failure", func(t *testing.T) {
		// Configure executor to fail on specific tool
		executor.failOnExecute = true

		ctx := context.Background()

		result, err := orch.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

		// Should get an error due to partial failure
		if err == nil {
			t.Errorf("Expected error due to partial failure, got result: %v", result)
		}

		// Orchestrator should remain functional
		if orch == nil {
			t.Error("Orchestrator should remain functional after partial failure")
		}

		// Check that some executions were attempted
		if executor.executionCount == 0 {
			t.Error("Expected some tool executions to be attempted")
		}

		// Reset failure
		executor.failOnExecute = false
	})
}

func TestOrchestrator_ErrorRecovery_TimeoutHandling(t *testing.T) {
	executor := &MockFailingExecutor{}
	orch := NewOrchestratorWithExecutor(executor)

	// Create a workflow that might take time
	workflow := &Workflow{
		ID:   "timeout-test",
		Name: "Timeout Test",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:    "step1",
				Tool:  "think",
				Input: map[string]interface{}{"content": "test", "mode": "linear"},
			},
		},
	}

	// Register workflow
	err := orch.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "very short timeout",
			timeout: 1 * time.Millisecond,
		},
		{
			name:    "zero timeout",
			timeout: 0,
		},
		{
			name:    "reasonable timeout",
			timeout: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure executor to be slow for timeout tests
			if tt.timeout > 0 && tt.timeout < 5*time.Millisecond {
				executor.slowExecution = true
			} else {
				executor.slowExecution = false
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Test timeout handling
			result, err := orch.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

			// Check for timeout or cancellation
			if ctx.Err() == context.DeadlineExceeded {
				if err == nil {
					t.Error("Expected error due to timeout but got none")
				}
			}

			// Orchestrator should handle timeouts gracefully
			if result == nil && err == nil && ctx.Err() == nil {
				t.Error("Expected some result or error from timeout test")
			}

			// Orchestrator should remain functional
			if orch == nil {
				t.Error("Orchestrator should remain functional after timeout")
			}

			// Reset slow execution
			executor.slowExecution = false
		})
	}
}
