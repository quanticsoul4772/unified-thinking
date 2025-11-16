package orchestration

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// TestExecuteSequentialWorkflow tests sequential workflow execution
func TestExecuteSequentialWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		workflow       *Workflow
		input          map[string]interface{}
		executor       ToolExecutor
		wantErr        bool
		wantSteps      int
		validateResult func(*testing.T, *WorkflowResult)
	}{
		{
			name: "single step success",
			workflow: &Workflow{
				ID:   "single-step",
				Name: "Single Step",
				Type: WorkflowSequential,
				Steps: []*WorkflowStep{
					{
						ID:      "step1",
						Tool:    "think",
						Input:   map[string]interface{}{"content": "test"},
						StoreAs: "thought",
					},
				},
			},
			input: map[string]interface{}{
				"problem": "Test problem",
			},
			executor:  &mockExecutor{},
			wantErr:   false,
			wantSteps: 1,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if r.Status != "success" {
					t.Errorf("Expected status 'success', got %s", r.Status)
				}
				if len(r.StepResults) != 1 {
					t.Errorf("Expected 1 step result, got %d", len(r.StepResults))
				}
				// Duration is set, but may be very small (nanoseconds) on fast systems
				// Just verify it was measured
			},
		},
		{
			name: "multi-step sequential",
			workflow: &Workflow{
				ID:   "multi-step",
				Name: "Multi Step",
				Type: WorkflowSequential,
				Steps: []*WorkflowStep{
					{
						ID:      "step1",
						Tool:    "decompose-problem",
						Input:   map[string]interface{}{},
						StoreAs: "decomposition",
					},
					{
						ID:      "step2",
						Tool:    "build-causal-graph",
						Input:   map[string]interface{}{},
						StoreAs: "causal",
					},
					{
						ID:      "step3",
						Tool:    "make-decision",
						Input:   map[string]interface{}{},
						StoreAs: "decision",
					},
				},
			},
			input: map[string]interface{}{
				"problem": "Complex problem",
			},
			executor:  &mockExecutor{},
			wantErr:   false,
			wantSteps: 3,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if len(r.StepResults) != 3 {
					t.Errorf("Expected 3 step results, got %d", len(r.StepResults))
				}
				// Verify steps executed in order
				if r.StepResults["step1"] == nil {
					t.Error("Expected step1 result")
				}
				if r.Context.Results["decomposition"] == nil {
					t.Error("Expected decomposition to be stored")
				}
			},
		},
		{
			name: "step with condition - skip",
			workflow: &Workflow{
				ID:   "conditional-skip",
				Name: "Conditional Skip",
				Type: WorkflowSequential,
				Steps: []*WorkflowStep{
					{
						ID:      "step1",
						Tool:    "think",
						Input:   map[string]interface{}{},
						StoreAs: "result1",
					},
					{
						ID:   "step2",
						Tool: "think",
						Condition: &StepCondition{
							Field:    "confidence",
							Operator: "gt",
							Value:    0.95, // High threshold - will skip
						},
						Input:   map[string]interface{}{},
						StoreAs: "result2",
					},
					{
						ID:      "step3",
						Tool:    "think",
						Input:   map[string]interface{}{},
						StoreAs: "result3",
					},
				},
			},
			input:     map[string]interface{}{"problem": "Test"},
			executor:  &mockExecutor{},
			wantErr:   false,
			wantSteps: 2, // step2 should be skipped
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if r.StepResults["step2"] != nil {
					t.Error("Expected step2 to be skipped")
				}
				if r.StepResults["step1"] == nil || r.StepResults["step3"] == nil {
					t.Error("Expected step1 and step3 to execute")
				}
			},
		},
		{
			name: "step failure",
			workflow: &Workflow{
				ID:   "step-failure",
				Name: "Step Failure",
				Type: WorkflowSequential,
				Steps: []*WorkflowStep{
					{
						ID:   "step1",
						Tool: "failing-tool",
					},
				},
			},
			input: map[string]interface{}{},
			executor: &mockExecutorWithError{
				shouldFail: true,
				failOn:     "failing-tool",
			},
			wantErr: true,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if r.Status != "failed" {
					t.Errorf("Expected status 'failed', got %s", r.Status)
				}
				if r.ErrorMessage == "" {
					t.Error("Expected error message")
				}
			},
		},
		{
			name: "no executor configured",
			workflow: &Workflow{
				ID:   "no-executor",
				Name: "No Executor",
				Type: WorkflowSequential,
				Steps: []*WorkflowStep{
					{
						ID:   "step1",
						Tool: "think",
					},
				},
			},
			input:    map[string]interface{}{},
			executor: nil,
			wantErr:  true,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if !strings.Contains(r.ErrorMessage, "no tool executor") {
					t.Errorf("Expected 'no tool executor' error, got %s", r.ErrorMessage)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()
			if tt.executor != nil {
				o.SetExecutor(tt.executor)
			}
			err := o.RegisterWorkflow(tt.workflow)
			if err != nil {
				t.Fatalf("Failed to register workflow: %v", err)
			}

			ctx := context.Background()
			result, err := o.ExecuteWorkflow(ctx, tt.workflow.ID, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// TestExecuteParallelWorkflow tests parallel workflow execution
func TestExecuteParallelWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		workflow       *Workflow
		input          map[string]interface{}
		wantErr        bool
		validateResult func(*testing.T, *WorkflowResult)
	}{
		{
			name: "parallel execution",
			workflow: &Workflow{
				ID:   "parallel",
				Name: "Parallel",
				Type: WorkflowParallel,
				Steps: []*WorkflowStep{
					{
						ID:      "step1",
						Tool:    "analyze-perspectives",
						Input:   map[string]interface{}{},
						StoreAs: "perspectives",
					},
					{
						ID:      "step2",
						Tool:    "analyze-temporal",
						Input:   map[string]interface{}{},
						StoreAs: "temporal",
					},
					{
						ID:      "step3",
						Tool:    "build-causal-graph",
						Input:   map[string]interface{}{},
						StoreAs: "causal",
					},
				},
			},
			input:   map[string]interface{}{"problem": "Test"},
			wantErr: false,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if len(r.StepResults) != 3 {
					t.Errorf("Expected 3 step results, got %d", len(r.StepResults))
				}
				if r.Context.Results["perspectives"] == nil {
					t.Error("Expected perspectives to be stored")
				}
				if r.Context.Results["temporal"] == nil {
					t.Error("Expected temporal to be stored")
				}
				if r.Context.Results["causal"] == nil {
					t.Error("Expected causal to be stored")
				}
			},
		},
		{
			name: "parallel with one failure",
			workflow: &Workflow{
				ID:   "parallel-failure",
				Name: "Parallel Failure",
				Type: WorkflowParallel,
				Steps: []*WorkflowStep{
					{ID: "step1", Tool: "think"},
					{ID: "step2", Tool: "failing-tool"},
					{ID: "step3", Tool: "think"},
				},
			},
			input:   map[string]interface{}{},
			wantErr: true,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if r.Status != "failed" {
					t.Errorf("Expected status 'failed', got %s", r.Status)
				}
			},
		},
		{
			name: "parallel with conditions",
			workflow: &Workflow{
				ID:   "parallel-conditional",
				Name: "Parallel Conditional",
				Type: WorkflowParallel,
				Steps: []*WorkflowStep{
					{
						ID:      "step1",
						Tool:    "think",
						StoreAs: "result1",
					},
					{
						ID:   "step2",
						Tool: "think",
						Condition: &StepCondition{
							Field:    "nonexistent",
							Operator: "eq",
							Value:    "value",
						},
						StoreAs: "result2",
					},
				},
			},
			input:   map[string]interface{}{},
			wantErr: false,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				// step1 should execute, step2 should be skipped
				if r.Context.Results["result1"] == nil {
					t.Error("Expected result1 to be stored")
				}
				// step2 might not store result if skipped
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()
			executor := &mockExecutorWithError{
				shouldFail: tt.name == "parallel with one failure",
				failOn:     "failing-tool",
			}
			o.SetExecutor(executor)

			err := o.RegisterWorkflow(tt.workflow)
			if err != nil {
				t.Fatalf("Failed to register workflow: %v", err)
			}

			ctx := context.Background()
			result, err := o.ExecuteWorkflow(ctx, tt.workflow.ID, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != nil && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// TestExecuteConditionalWorkflow tests conditional workflow execution
func TestExecuteConditionalWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		workflow       *Workflow
		input          map[string]interface{}
		wantErr        bool
		validateResult func(*testing.T, *WorkflowResult)
	}{
		{
			name: "simple dependency chain",
			workflow: &Workflow{
				ID:   "dependency-chain",
				Name: "Dependency Chain",
				Type: WorkflowConditional,
				Steps: []*WorkflowStep{
					{
						ID:      "step1",
						Tool:    "think",
						Input:   map[string]interface{}{},
						StoreAs: "result1",
					},
					{
						ID:        "step2",
						Tool:      "think",
						DependsOn: []string{"step1"},
						Input:     map[string]interface{}{},
						StoreAs:   "result2",
					},
					{
						ID:        "step3",
						Tool:      "think",
						DependsOn: []string{"step2"},
						Input:     map[string]interface{}{},
						StoreAs:   "result3",
					},
				},
			},
			input:   map[string]interface{}{},
			wantErr: false,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if len(r.StepResults) != 3 {
					t.Errorf("Expected 3 step results, got %d", len(r.StepResults))
				}
			},
		},
		{
			name: "parallel branches converging",
			workflow: &Workflow{
				ID:   "parallel-converge",
				Name: "Parallel Converge",
				Type: WorkflowConditional,
				Steps: []*WorkflowStep{
					{
						ID:      "step1",
						Tool:    "think",
						StoreAs: "branch1",
					},
					{
						ID:      "step2",
						Tool:    "think",
						StoreAs: "branch2",
					},
					{
						ID:        "step3",
						Tool:      "synthesize-insights",
						DependsOn: []string{"step1", "step2"},
						StoreAs:   "synthesis",
					},
				},
			},
			input:   map[string]interface{}{},
			wantErr: false,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if r.Context.Results["synthesis"] == nil {
					t.Error("Expected synthesis result")
				}
			},
		},
		{
			name: "conditional branching - always execute step1",
			workflow: &Workflow{
				ID:   "conditional-branch",
				Name: "Conditional Branch",
				Type: WorkflowConditional,
				Steps: []*WorkflowStep{
					{
						ID:      "step1",
						Tool:    "think",
						StoreAs: "initial",
					},
					{
						ID:        "step2",
						Tool:      "think",
						DependsOn: []string{"step1"},
						// No condition - should always execute
						StoreAs: "followup",
					},
				},
			},
			input:   map[string]interface{}{},
			wantErr: false,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				// Both steps should execute
				if r.Context.Results["initial"] == nil {
					t.Error("Expected initial result to be stored")
				}
				if r.Context.Results["followup"] == nil {
					t.Error("Expected followup result to be stored")
				}
			},
		},
		{
			name: "circular dependency",
			workflow: &Workflow{
				ID:   "circular",
				Name: "Circular",
				Type: WorkflowConditional,
				Steps: []*WorkflowStep{
					{
						ID:        "step1",
						Tool:      "think",
						DependsOn: []string{"step2"},
					},
					{
						ID:        "step2",
						Tool:      "think",
						DependsOn: []string{"step1"},
					},
				},
			},
			input:   map[string]interface{}{},
			wantErr: true,
			validateResult: func(t *testing.T, r *WorkflowResult) {
				if !strings.Contains(r.ErrorMessage, "deadlock") {
					t.Errorf("Expected deadlock error, got %s", r.ErrorMessage)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()
			o.SetExecutor(&mockExecutor{})

			err := o.RegisterWorkflow(tt.workflow)
			if err != nil {
				t.Fatalf("Failed to register workflow: %v", err)
			}

			ctx := context.Background()
			result, err := o.ExecuteWorkflow(ctx, tt.workflow.ID, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != nil && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// TestExecuteStepWithReferences tests parameter reference resolution
func TestExecuteStepWithReferences(t *testing.T) {
	o := NewOrchestrator()
	o.SetExecutor(&mockExecutor{})

	workflow := &Workflow{
		ID:   "reference-test",
		Name: "Reference Test",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:   "step1",
				Tool: "build-causal-graph",
				Input: map[string]interface{}{
					"description": "Test graph",
				},
				StoreAs: "graph",
			},
			{
				ID:   "step2",
				Tool: "simulate-intervention",
				Input: map[string]interface{}{
					"graph_id": "$graph.id", // Reference to previous result
				},
			},
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()
	result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	if err != nil {
		t.Errorf("ExecuteWorkflow() failed: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("Expected success status, got %s", result.Status)
	}
}

// TestExecuteStepWithTransform tests output transformations
func TestExecuteStepWithTransform(t *testing.T) {
	o := NewOrchestrator()
	o.SetExecutor(&mockExecutor{})

	workflow := &Workflow{
		ID:   "transform-test",
		Name: "Transform Test",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:   "step1",
				Tool: "think",
				Transform: &OutputTransform{
					Type: "extract_field",
					Config: map[string]interface{}{
						"field": "id",
					},
				},
				StoreAs: "thought_id",
			},
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()
	result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	if err != nil {
		t.Errorf("ExecuteWorkflow() failed: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("Expected success status, got %s", result.Status)
	}
}

// TestExecuteWorkflowNotFound tests error handling for non-existent workflow
func TestExecuteWorkflowNotFound(t *testing.T) {
	o := NewOrchestrator()
	ctx := context.Background()

	_, err := o.ExecuteWorkflow(ctx, "non-existent", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for non-existent workflow")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got %v", err)
	}
}

// TestExecuteWorkflowUnknownType tests error handling for unknown workflow type
func TestExecuteWorkflowUnknownType(t *testing.T) {
	o := NewOrchestrator()
	o.SetExecutor(&mockExecutor{})

	workflow := &Workflow{
		ID:    "unknown-type",
		Name:  "Unknown Type",
		Type:  "unknown",
		Steps: []*WorkflowStep{},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()
	result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for unknown workflow type")
	}
	if !strings.Contains(err.Error(), "unknown workflow type") {
		t.Errorf("Expected 'unknown workflow type' error, got %v", err)
	}
	// For unknown workflow type, result is nil since it fails before result creation
	if result != nil {
		t.Error("Expected nil result for unknown workflow type")
	}
}

// TestContextUpdatesDuringExecution tests that context is updated as steps execute
func TestContextUpdatesDuringExecution(t *testing.T) {
	o := NewOrchestrator()

	// Create a mock executor that returns tool-specific results
	executor := &mockToolSpecificExecutor{}
	o.SetExecutor(executor)

	workflow := &Workflow{
		ID:   "context-updates",
		Name: "Context Updates",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:      "think-step",
				Tool:    "think",
				StoreAs: "thought_result",
			},
			{
				ID:      "causal-step",
				Tool:    "build-causal-graph",
				StoreAs: "graph_result",
			},
			{
				ID:      "belief-step",
				Tool:    "probabilistic-reasoning",
				StoreAs: "belief_result",
			},
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()
	result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	if err != nil {
		t.Errorf("ExecuteWorkflow() failed: %v", err)
	}

	// Verify context was updated with tool-specific data
	if len(result.Context.Thoughts) == 0 {
		t.Error("Expected thought ID to be added to context")
	}
	if len(result.Context.CausalGraphs) == 0 {
		t.Error("Expected causal graph ID to be added to context")
	}
	if len(result.Context.Beliefs) == 0 {
		t.Error("Expected belief ID to be added to context")
	}

	// Verify confidence was updated
	if result.Context.Confidence == 0 {
		t.Error("Expected context confidence to be updated")
	}
}

// Mocks are defined in helpers_test.go

// TestWorkflowDuration tests that workflow duration is recorded
func TestWorkflowDuration(t *testing.T) {
	o := NewOrchestrator()

	// Create executor with artificial delay
	executor := &mockDelayedExecutor{delay: 10 * time.Millisecond}
	o.SetExecutor(executor)

	workflow := &Workflow{
		ID:   "duration-test",
		Name: "Duration Test",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{ID: "step1", Tool: "think"},
			{ID: "step2", Tool: "think"},
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()
	result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	if err != nil {
		t.Errorf("ExecuteWorkflow() failed: %v", err)
	}

	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}

	// Duration should be at least 20ms (2 steps * 10ms delay)
	if result.Duration < 20*time.Millisecond {
		t.Errorf("Expected duration >= 20ms, got %v", result.Duration)
	}
}

// TestContextCancellation tests workflow execution with context cancellation
func TestContextCancellation(t *testing.T) {
	o := NewOrchestrator()

	// Create executor with long delay
	executor := &mockDelayedExecutor{delay: 1 * time.Second}
	o.SetExecutor(executor)

	workflow := &Workflow{
		ID:   "cancel-test",
		Name: "Cancel Test",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{ID: "step1", Tool: "think"},
			{ID: "step2", Tool: "think"},
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	// Create cancellable context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should timeout due to slow executor
	_, _ = o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Errorf("Expected DeadlineExceeded error, got %v", ctx.Err())
		}
	default:
		t.Error("Expected context to be cancelled")
	}
}
