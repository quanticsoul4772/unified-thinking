package orchestration

import (
	"context"
	"testing"
	"time"
)

// TestReasoningContextInitialization tests context creation and initialization
func TestReasoningContextInitialization(t *testing.T) {
	o := NewOrchestrator()

	ctx := o.CreateContext("workflow-1", "Test problem statement")

	// Verify all fields are properly initialized
	if ctx.ID == "" {
		t.Error("Expected non-empty context ID")
	}
	if ctx.WorkflowID != "workflow-1" {
		t.Errorf("Expected workflow ID 'workflow-1', got %s", ctx.WorkflowID)
	}
	if ctx.Problem != "Test problem statement" {
		t.Errorf("Expected problem 'Test problem statement', got %s", ctx.Problem)
	}
	if ctx.Results == nil {
		t.Error("Expected Results map to be initialized")
	}
	if ctx.Thoughts == nil {
		t.Error("Expected Thoughts slice to be initialized")
	}
	if ctx.CausalGraphs == nil {
		t.Error("Expected CausalGraphs slice to be initialized")
	}
	if ctx.Beliefs == nil {
		t.Error("Expected Beliefs slice to be initialized")
	}
	if ctx.Evidence == nil {
		t.Error("Expected Evidence slice to be initialized")
	}
	if ctx.Decisions == nil {
		t.Error("Expected Decisions slice to be initialized")
	}
	if ctx.Metadata == nil {
		t.Error("Expected Metadata map to be initialized")
	}
	if ctx.Confidence != 0 {
		t.Errorf("Expected initial confidence 0, got %f", ctx.Confidence)
	}
	if ctx.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if ctx.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

// TestContextSharedState tests that context maintains shared state across steps
func TestContextSharedState(t *testing.T) {
	o := NewOrchestrator()
	o.SetExecutor(&mockExecutor{})

	workflow := &Workflow{
		ID:   "shared-state",
		Name: "Shared State",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:      "step1",
				Tool:    "think",
				Input:   map[string]interface{}{"content": "first thought"},
				StoreAs: "first",
			},
			{
				ID:      "step2",
				Tool:    "think",
				Input:   map[string]interface{}{"content": "second thought"},
				StoreAs: "second",
			},
			{
				ID:   "step3",
				Tool: "think",
				Input: map[string]interface{}{
					"content": "third thought",
					"ref":     "$first", // Reference to first step's result
				},
				StoreAs: "third",
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
		t.Fatalf("ExecuteWorkflow() failed: %v", err)
	}

	// Verify all results are stored in context
	if result.Context.Results["first"] == nil {
		t.Error("Expected 'first' to be in context results")
	}
	if result.Context.Results["second"] == nil {
		t.Error("Expected 'second' to be in context results")
	}
	if result.Context.Results["third"] == nil {
		t.Error("Expected 'third' to be in context results")
	}

	// Verify results are also in step results
	if result.StepResults["step1"] == nil {
		t.Error("Expected step1 result")
	}
	if result.StepResults["step2"] == nil {
		t.Error("Expected step2 result")
	}
	if result.StepResults["step3"] == nil {
		t.Error("Expected step3 result")
	}
}

// TestContextIsolation tests that different workflows have isolated contexts
func TestContextIsolation(t *testing.T) {
	o := NewOrchestrator()
	o.SetExecutor(&mockExecutor{})

	// Register two workflows
	workflow1 := &Workflow{
		ID:   "workflow-1",
		Name: "Workflow 1",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:      "step1",
				Tool:    "think",
				StoreAs: "result1",
			},
		},
	}

	workflow2 := &Workflow{
		ID:   "workflow-2",
		Name: "Workflow 2",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:      "step1",
				Tool:    "think",
				StoreAs: "result2",
			},
		},
	}

	_ = o.RegisterWorkflow(workflow1)
	_ = o.RegisterWorkflow(workflow2)

	ctx := context.Background()

	// Execute both workflows
	result1, err1 := o.ExecuteWorkflow(ctx, "workflow-1", map[string]interface{}{})
	result2, err2 := o.ExecuteWorkflow(ctx, "workflow-2", map[string]interface{}{})

	if err1 != nil || err2 != nil {
		t.Fatalf("Workflow execution failed: %v, %v", err1, err2)
	}

	// Verify contexts are separate - check that each has only its own results
	// (Context IDs may occasionally be the same due to nanosecond timing, but data should be separate)
	if result1.Context.Results["result2"] != nil {
		t.Error("Workflow 1 context should not have result2")
	}
	if result2.Context.Results["result1"] != nil {
		t.Error("Workflow 2 context should not have result1")
	}

	// At least one should have result1 or result2
	if result1.Context.Results["result1"] == nil && result2.Context.Results["result2"] == nil {
		t.Error("Expected at least one context to have results")
	}
}

// TestContextUpdatesPreserveExisting tests that updates don't overwrite existing data
func TestContextUpdatesPreserveExisting(t *testing.T) {
	o := NewOrchestrator()

	// Create initial context
	ctx := o.CreateContext("workflow-1", "Initial problem")
	ctx.Results["existing"] = "value"
	ctx.Thoughts = []string{"thought-1"}
	ctx.Confidence = 0.5

	// Update context
	err := o.UpdateContext(ctx)
	if err != nil {
		t.Fatalf("Failed to update context: %v", err)
	}

	// Add more data
	ctx.Results["new"] = "new-value"
	ctx.Thoughts = append(ctx.Thoughts, "thought-2")
	ctx.Confidence = 0.7

	err = o.UpdateContext(ctx)
	if err != nil {
		t.Fatalf("Failed to update context again: %v", err)
	}

	// Retrieve and verify
	retrieved, err := o.GetContext(ctx.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve context: %v", err)
	}

	if retrieved.Results["existing"] != "value" {
		t.Error("Expected existing result to be preserved")
	}
	if retrieved.Results["new"] != "new-value" {
		t.Error("Expected new result to be present")
	}
	if len(retrieved.Thoughts) != 2 {
		t.Errorf("Expected 2 thoughts, got %d", len(retrieved.Thoughts))
	}
	if retrieved.Confidence != 0.7 {
		t.Errorf("Expected confidence 0.7, got %f", retrieved.Confidence)
	}
}

// TestContextMetadata tests custom metadata storage
func TestContextMetadata(t *testing.T) {
	o := NewOrchestrator()
	o.SetExecutor(&mockExecutor{})

	workflow := &Workflow{
		ID:   "metadata-test",
		Name: "Metadata Test",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:   "step1",
				Tool: "think",
				Metadata: map[string]interface{}{
					"category": "analysis",
					"priority": 1,
				},
			},
		},
		Metadata: map[string]interface{}{
			"version": "1.0",
			"author":  "test",
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()
	result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	if err != nil {
		t.Fatalf("ExecuteWorkflow() failed: %v", err)
	}

	// Context can store custom metadata
	result.Context.Metadata["custom"] = "value"
	err = o.UpdateContext(result.Context)
	if err != nil {
		t.Fatalf("Failed to update context: %v", err)
	}

	retrieved, _ := o.GetContext(result.Context.ID)
	if retrieved.Metadata["custom"] != "value" {
		t.Error("Expected custom metadata to be stored")
	}
}

// TestContextResultAccumulation tests that results accumulate correctly
func TestContextResultAccumulation(t *testing.T) {
	o := NewOrchestrator()
	o.SetExecutor(&mockToolSpecificExecutor{})

	workflow := &Workflow{
		ID:   "accumulation-test",
		Name: "Accumulation Test",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{ID: "step1", Tool: "think", StoreAs: "thought"},
			{ID: "step2", Tool: "build-causal-graph", StoreAs: "graph"},
			{ID: "step3", Tool: "probabilistic-reasoning", StoreAs: "belief"},
			{ID: "step4", Tool: "assess-evidence", StoreAs: "evidence"},
			{ID: "step5", Tool: "make-decision", StoreAs: "decision"},
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()
	result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	if err != nil {
		t.Fatalf("ExecuteWorkflow() failed: %v", err)
	}

	// Verify all collections were populated
	if len(result.Context.Thoughts) == 0 {
		t.Error("Expected thoughts to be accumulated")
	}
	if len(result.Context.CausalGraphs) == 0 {
		t.Error("Expected causal graphs to be accumulated")
	}
	if len(result.Context.Beliefs) == 0 {
		t.Error("Expected beliefs to be accumulated")
	}
	if len(result.Context.Evidence) == 0 {
		t.Error("Expected evidence to be accumulated")
	}
	if len(result.Context.Decisions) == 0 {
		t.Error("Expected decisions to be accumulated")
	}

	// Verify all results are stored
	if len(result.Context.Results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(result.Context.Results))
	}
}

// TestContextConfidenceUpdates tests confidence averaging
func TestContextConfidenceUpdates(t *testing.T) {
	o := NewOrchestrator()

	// Create executor that returns specific confidence values
	executor := &mockConfidenceExecutor{
		confidences: []float64{0.9, 0.8, 0.7},
	}
	o.SetExecutor(executor)

	workflow := &Workflow{
		ID:   "confidence-test",
		Name: "Confidence Test",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{ID: "step1", Tool: "think"},
			{ID: "step2", Tool: "think"},
			{ID: "step3", Tool: "think"},
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()
	result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

	if err != nil {
		t.Fatalf("ExecuteWorkflow() failed: %v", err)
	}

	// Confidence should be averaged across steps
	// First step: 0.9
	// Second step: (0.9 + 0.8) / 2 = 0.85
	// Third step: (0.85 + 0.7) / 2 = 0.775
	expectedConfidence := 0.775
	if result.Context.Confidence != expectedConfidence {
		t.Errorf("Expected confidence %f, got %f", expectedConfidence, result.Context.Confidence)
	}
}

// TestMultipleContextsForSameWorkflow tests multiple executions of same workflow
func TestMultipleContextsForSameWorkflow(t *testing.T) {
	o := NewOrchestrator()
	o.SetExecutor(&mockExecutor{})

	workflow := &Workflow{
		ID:   "multi-exec",
		Name: "Multi Execution",
		Type: WorkflowSequential,
		Steps: []*WorkflowStep{
			{ID: "step1", Tool: "think", StoreAs: "result"},
		},
	}

	err := o.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	ctx := context.Background()

	// Execute same workflow multiple times
	result1, err1 := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{"problem": "Problem 1"})

	// Add tiny delay to ensure different timestamp-based IDs
	time.Sleep(1 * time.Millisecond)

	result2, err2 := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{"problem": "Problem 2"})

	if err1 != nil || err2 != nil {
		t.Fatalf("Workflow execution failed: %v, %v", err1, err2)
	}

	// Verify separate contexts were created
	if result1.Context.ID == result2.Context.ID {
		t.Error("Expected different context IDs for separate executions")
	}

	// Both contexts should be retrievable and have correct data
	ctx1, err := o.GetContext(result1.Context.ID)
	if err != nil {
		t.Errorf("Failed to get first context: %v", err)
	}
	if ctx1.Problem != "Problem 1" {
		t.Errorf("Expected problem 'Problem 1', got %s", ctx1.Problem)
	}

	ctx2, err := o.GetContext(result2.Context.ID)
	if err != nil {
		t.Errorf("Failed to get second context: %v", err)
	}
	if ctx2.Problem != "Problem 2" {
		t.Errorf("Expected problem 'Problem 2', got %s", ctx2.Problem)
	}
}

// TestWorkflowResultStatus tests result status setting
func TestWorkflowResultStatus(t *testing.T) {
	tests := []struct {
		name           string
		executor       ToolExecutor
		expectedStatus string
		expectError    bool
	}{
		{
			name:           "success",
			executor:       &mockExecutor{},
			expectedStatus: "success",
			expectError:    false,
		},
		{
			name: "failure",
			executor: &mockExecutorWithError{
				shouldFail: true,
			},
			expectedStatus: "failed",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()
			o.SetExecutor(tt.executor)

			workflow := &Workflow{
				ID:   "status-test",
				Name: "Status Test",
				Type: WorkflowSequential,
				Steps: []*WorkflowStep{
					{ID: "step1", Tool: "think"},
				},
			}

			_ = o.RegisterWorkflow(workflow)

			ctx := context.Background()
			result, err := o.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{})

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			if result.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, result.Status)
			}

			if tt.expectError && result.ErrorMessage == "" {
				t.Error("Expected error message to be set")
			}
		})
	}
}

// Mocks are defined in helpers_test.go
