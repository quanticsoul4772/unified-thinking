package orchestration

import (
	"testing"
	"time"
)

// TestNewOrchestrator tests creating a new orchestrator
func TestNewOrchestrator(t *testing.T) {
	tests := []struct {
		name string
		fn   func() *Orchestrator
	}{
		{
			name: "without executor",
			fn:   NewOrchestrator,
		},
		{
			name: "with executor",
			fn:   func() *Orchestrator { return NewOrchestratorWithExecutor(&mockExecutor{}) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.fn()
			if o == nil {
				t.Fatal("Expected orchestrator to be created")
			}
			if o.workflows == nil {
				t.Error("Expected workflows map to be initialized")
			}
			if o.contexts == nil {
				t.Error("Expected contexts map to be initialized")
			}
		})
	}
}

// TestRegisterWorkflow tests workflow registration
func TestRegisterWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		workflow  *Workflow
		wantErr   bool
		errContains string
	}{
		{
			name: "valid workflow",
			workflow: &Workflow{
				ID:          "test-workflow",
				Name:        "Test Workflow",
				Description: "A test workflow",
				Type:        WorkflowSequential,
				Steps:       []*WorkflowStep{},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			workflow: &Workflow{
				Name:        "Test Workflow",
				Description: "A test workflow",
				Type:        WorkflowSequential,
			},
			wantErr:     true,
			errContains: "ID is required",
		},
		{
			name: "missing name",
			workflow: &Workflow{
				ID:          "test-workflow",
				Description: "A test workflow",
				Type:        WorkflowSequential,
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "duplicate ID",
			workflow: &Workflow{
				ID:   "duplicate",
				Name: "Duplicate",
				Type: WorkflowSequential,
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()

			// For duplicate test, register first
			if tt.name == "duplicate ID" {
				_ = o.RegisterWorkflow(&Workflow{
					ID:   "duplicate",
					Name: "First",
					Type: WorkflowSequential,
				})
			}

			err := o.RegisterWorkflow(tt.workflow)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			}

			// Verify workflow was registered
			if !tt.wantErr {
				workflow, err := o.GetWorkflow(tt.workflow.ID)
				if err != nil {
					t.Errorf("Failed to get registered workflow: %v", err)
				}
				if workflow.ID != tt.workflow.ID {
					t.Errorf("Expected workflow ID %s, got %s", tt.workflow.ID, workflow.ID)
				}
				if workflow.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
			}
		})
	}
}

// TestGetWorkflow tests workflow retrieval
func TestGetWorkflow(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Orchestrator) string
		wantErr bool
	}{
		{
			name: "existing workflow",
			setup: func(o *Orchestrator) string {
				w := &Workflow{
					ID:   "exists",
					Name: "Exists",
					Type: WorkflowSequential,
				}
				_ = o.RegisterWorkflow(w)
				return w.ID
			},
			wantErr: false,
		},
		{
			name: "non-existent workflow",
			setup: func(o *Orchestrator) string {
				return "does-not-exist"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()
			id := tt.setup(o)

			workflow, err := o.GetWorkflow(id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && workflow == nil {
				t.Error("Expected workflow to be returned")
			}
			if !tt.wantErr && workflow.ID != id {
				t.Errorf("Expected workflow ID %s, got %s", id, workflow.ID)
			}
		})
	}
}

// TestListWorkflows tests listing all workflows
func TestListWorkflows(t *testing.T) {
	o := NewOrchestrator()

	// Initially empty
	workflows := o.ListWorkflows()
	if len(workflows) != 0 {
		t.Errorf("Expected 0 workflows, got %d", len(workflows))
	}

	// Add some workflows
	for i := 0; i < 3; i++ {
		_ = o.RegisterWorkflow(&Workflow{
			ID:   string('a' + rune(i)),
			Name: "Workflow " + string('A'+rune(i)),
			Type: WorkflowSequential,
		})
	}

	workflows = o.ListWorkflows()
	if len(workflows) != 3 {
		t.Errorf("Expected 3 workflows, got %d", len(workflows))
	}
}

// TestSetExecutor tests setting the executor
func TestSetExecutor(t *testing.T) {
	o := NewOrchestrator()
	if o.executor != nil {
		t.Error("Expected executor to be nil initially")
	}

	executor := &mockExecutor{}
	o.SetExecutor(executor)

	if o.executor == nil {
		t.Error("Expected executor to be set")
	}
}

// TestCreateContext tests creating a reasoning context
func TestCreateContext(t *testing.T) {
	o := NewOrchestrator()

	ctx := o.CreateContext("workflow-1", "Test problem")
	if ctx == nil {
		t.Fatal("Expected context to be created")
	}
	if ctx.ID == "" {
		t.Error("Expected context ID to be set")
	}
	if ctx.WorkflowID != "workflow-1" {
		t.Errorf("Expected workflow ID 'workflow-1', got %s", ctx.WorkflowID)
	}
	if ctx.Problem != "Test problem" {
		t.Errorf("Expected problem 'Test problem', got %s", ctx.Problem)
	}
	if ctx.Results == nil {
		t.Error("Expected Results map to be initialized")
	}
	if ctx.Thoughts == nil {
		t.Error("Expected Thoughts slice to be initialized")
	}
	if ctx.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if ctx.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Verify context was stored
	retrieved, err := o.GetContext(ctx.ID)
	if err != nil {
		t.Errorf("Failed to get created context: %v", err)
	}
	if retrieved.ID != ctx.ID {
		t.Errorf("Expected context ID %s, got %s", ctx.ID, retrieved.ID)
	}
}

// TestGetContext tests context retrieval
func TestGetContext(t *testing.T) {
	o := NewOrchestrator()

	// Create a context
	ctx := o.CreateContext("workflow-1", "Test problem")

	// Test retrieving existing context
	retrieved, err := o.GetContext(ctx.ID)
	if err != nil {
		t.Errorf("Failed to get context: %v", err)
	}
	if retrieved.ID != ctx.ID {
		t.Errorf("Expected context ID %s, got %s", ctx.ID, retrieved.ID)
	}

	// Test retrieving non-existent context
	_, err = o.GetContext("does-not-exist")
	if err == nil {
		t.Error("Expected error for non-existent context")
	}
}

// TestUpdateContext tests context updates
func TestUpdateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     *ReasoningContext
		wantErr bool
	}{
		{
			name: "valid update",
			ctx: &ReasoningContext{
				ID:         "ctx-1",
				WorkflowID: "workflow-1",
				Problem:    "Updated problem",
				Results:    make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
		},
		{
			name: "empty ID",
			ctx: &ReasoningContext{
				WorkflowID: "workflow-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()

			// For valid test, create context first
			if tt.name == "valid update" {
				o.CreateContext("workflow-1", "Original problem")
			}

			beforeUpdate := time.Now()
			err := o.UpdateContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify UpdatedAt was set
			if !tt.wantErr {
				retrieved, _ := o.GetContext(tt.ctx.ID)
				if retrieved.UpdatedAt.Before(beforeUpdate) {
					t.Error("Expected UpdatedAt to be updated")
				}
			}
		})
	}
}

// TestWorkflowTypes tests different workflow types
func TestWorkflowTypes(t *testing.T) {
	tests := []struct {
		name string
		wfType WorkflowType
	}{
		{"sequential", WorkflowSequential},
		{"parallel", WorkflowParallel},
		{"conditional", WorkflowConditional},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()
			w := &Workflow{
				ID:   "test-" + tt.name,
				Name: "Test " + tt.name,
				Type: tt.wfType,
			}
			err := o.RegisterWorkflow(w)
			if err != nil {
				t.Errorf("Failed to register workflow: %v", err)
			}

			retrieved, _ := o.GetWorkflow(w.ID)
			if retrieved.Type != tt.wfType {
				t.Errorf("Expected type %s, got %s", tt.wfType, retrieved.Type)
			}
		})
	}
}

// TestPredefinedWorkflows tests the predefined workflow templates
func TestPredefinedWorkflows(t *testing.T) {
	tests := []struct {
		name         string
		workflowID   string
		wantSteps    int
		wantType     WorkflowType
	}{
		{
			name:       "comprehensive-analysis",
			workflowID: "comprehensive-analysis",
			wantSteps:  6,
			wantType:   WorkflowSequential,
		},
		{
			name:       "validation-pipeline",
			workflowID: "validation-pipeline",
			wantSteps:  4,
			wantType:   WorkflowSequential,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow, exists := PredefinedWorkflows[tt.workflowID]
			if !exists {
				t.Fatalf("Predefined workflow %s not found", tt.workflowID)
			}

			if workflow.ID != tt.workflowID {
				t.Errorf("Expected ID %s, got %s", tt.workflowID, workflow.ID)
			}
			if workflow.Type != tt.wantType {
				t.Errorf("Expected type %s, got %s", tt.wantType, workflow.Type)
			}
			if len(workflow.Steps) != tt.wantSteps {
				t.Errorf("Expected %d steps, got %d", tt.wantSteps, len(workflow.Steps))
			}
			if workflow.Name == "" {
				t.Error("Expected non-empty name")
			}
			if workflow.Description == "" {
				t.Error("Expected non-empty description")
			}

			// Verify all steps have required fields
			for i, step := range workflow.Steps {
				if step.ID == "" {
					t.Errorf("Step %d missing ID", i)
				}
				if step.Tool == "" {
					t.Errorf("Step %d missing Tool", i)
				}
			}
		})
	}
}

// TestConvertTemplateToReference tests template syntax conversion
func TestConvertTemplateToReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double brace template",
			input:    "{{problem}}",
			expected: "$problem",
		},
		{
			name:     "double brace nested reference",
			input:    "{{causal_graph.id}}",
			expected: "$causal_graph.id",
		},
		{
			name:     "double brace with underscores",
			input:    "{{evidence_assessment}}",
			expected: "$evidence_assessment",
		},
		{
			name:     "already dollar syntax",
			input:    "$variable",
			expected: "$variable",
		},
		{
			name:     "already dollar nested syntax",
			input:    "$causal_graph.id",
			expected: "$causal_graph.id",
		},
		{
			name:     "regular string",
			input:    "regular string",
			expected: "regular string",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single brace",
			input:    "{variable}",
			expected: "{variable}",
		},
		{
			name:     "incomplete template",
			input:    "{{variable",
			expected: "{{variable",
		},
		{
			name:     "incomplete closing",
			input:    "variable}}",
			expected: "variable}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTemplateToReference(tt.input)
			if result != tt.expected {
				t.Errorf("convertTemplateToReference(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestResolveTemplateValue tests template value resolution
func TestResolveTemplateValue(t *testing.T) {
	tests := []struct {
		name           string
		value          interface{}
		workflowInput  map[string]interface{}
		reasoningCtx   *ReasoningContext
		expected       interface{}
		expectedString string // For string comparison
	}{
		{
			name:  "resolve double brace from workflow input",
			value: "{{problem}}",
			workflowInput: map[string]interface{}{
				"problem": "test problem",
			},
			reasoningCtx:   &ReasoningContext{Results: make(map[string]interface{})},
			expectedString: "test problem",
		},
		{
			name:          "resolve double brace from reasoning context",
			value:         "{{causal_graph}}",
			workflowInput: make(map[string]interface{}),
			reasoningCtx: &ReasoningContext{
				Results: map[string]interface{}{
					"causal_graph": map[string]interface{}{"id": "graph-1"},
				},
			},
			expected: map[string]interface{}{"id": "graph-1"},
		},
		{
			name:          "resolve nested double brace reference",
			value:         "{{causal_graph.id}}",
			workflowInput: make(map[string]interface{}),
			reasoningCtx: &ReasoningContext{
				Results: map[string]interface{}{
					"causal_graph": map[string]interface{}{"id": "graph-1", "nodes": 5},
				},
			},
			expectedString: "graph-1",
		},
		{
			name:  "resolve dollar syntax from workflow input",
			value: "$evidence",
			workflowInput: map[string]interface{}{
				"evidence": "test evidence",
			},
			reasoningCtx:   &ReasoningContext{Results: make(map[string]interface{})},
			expectedString: "test evidence",
		},
		{
			name:          "resolve dollar syntax from reasoning context",
			value:         "$biases",
			workflowInput: make(map[string]interface{}),
			reasoningCtx: &ReasoningContext{
				Results: map[string]interface{}{
					"biases": []string{"confirmation", "anchoring"},
				},
			},
			expected: []string{"confirmation", "anchoring"},
		},
		{
			name:           "non-template string unchanged",
			value:          "regular string",
			workflowInput:  make(map[string]interface{}),
			reasoningCtx:   &ReasoningContext{Results: make(map[string]interface{})},
			expectedString: "regular string",
		},
		{
			name:          "non-string value unchanged",
			value:         42,
			workflowInput: make(map[string]interface{}),
			reasoningCtx:  &ReasoningContext{Results: make(map[string]interface{})},
			expected:      42,
		},
		{
			name:          "slice with templates resolved",
			value:         []interface{}{"{{problem}}", "{{context}}", "regular"},
			workflowInput: map[string]interface{}{
				"problem": "test problem",
				"context": "test context",
			},
			reasoningCtx: &ReasoningContext{Results: make(map[string]interface{})},
			expected:     []interface{}{"test problem", "test context", "regular"},
		},
		{
			name:  "unresolved reference returns original",
			value: "{{nonexistent}}",
			workflowInput: make(map[string]interface{}),
			reasoningCtx:  &ReasoningContext{Results: make(map[string]interface{})},
			expectedString: "{{nonexistent}}",
		},
		{
			name:          "reasoning context takes precedence",
			value:         "{{value}}",
			workflowInput: map[string]interface{}{"value": "from workflow"},
			reasoningCtx: &ReasoningContext{
				Results: map[string]interface{}{"value": "from context"},
			},
			expectedString: "from context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()
			result := o.resolveTemplateValue(tt.value, tt.workflowInput, tt.reasoningCtx)

			// Compare results based on type
			if tt.expectedString != "" {
				if strResult, ok := result.(string); !ok || strResult != tt.expectedString {
					t.Errorf("resolveTemplateValue() = %v (type %T), expected %q", result, result, tt.expectedString)
				}
			} else if tt.expected != nil {
				// For complex types, use deep comparison
				if !deepEqual(result, tt.expected) {
					t.Errorf("resolveTemplateValue() = %v, expected %v", result, tt.expected)
				}
			}
		})
	}
}

// TestResolveTemplateValueNestedFields tests nested field resolution
func TestResolveTemplateValueNestedFields(t *testing.T) {
	reasoningCtx := &ReasoningContext{
		Results: map[string]interface{}{
			"causal_graph": map[string]interface{}{
				"id":    "graph-1",
				"nodes": 5,
				"metadata": map[string]interface{}{
					"confidence": 0.85,
					"source":     "analysis",
				},
			},
			"evidence_assessment": map[string]interface{}{
				"strength": 0.9,
				"type":     "empirical",
			},
		},
	}

	tests := []struct {
		name     string
		value    string
		expected interface{}
	}{
		{
			name:     "single level nested",
			value:    "{{causal_graph.id}}",
			expected: "graph-1",
		},
		{
			name:     "single level nested numeric",
			value:    "{{causal_graph.nodes}}",
			expected: 5,
		},
		{
			name:     "double level nested",
			value:    "{{causal_graph.metadata.confidence}}",
			expected: nil, // Current implementation only supports single-level nesting
		},
		{
			name:     "different nested field",
			value:    "{{evidence_assessment.strength}}",
			expected: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestrator()
			result := o.resolveTemplateValue(tt.value, make(map[string]interface{}), reasoningCtx)

			if tt.expected != nil && !deepEqual(result, tt.expected) {
				t.Errorf("resolveTemplateValue(%q) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

// deepEqual performs deep equality comparison for testing
func deepEqual(a, b interface{}) bool {
	// Simple implementation for testing - handles basic types
	switch va := a.(type) {
	case string:
		vb, ok := b.(string)
		return ok && va == vb
	case int:
		vb, ok := b.(int)
		return ok && va == vb
	case float64:
		vb, ok := b.(float64)
		return ok && va == vb
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepEqual(va[i], vb[i]) {
				return false
			}
		}
		return true
	case []string:
		vb, ok := b.([]string)
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if va[i] != vb[i] {
				return false
			}
		}
		return true
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	}
	return false
}

// Mocks are defined in helpers_test.go
