// Package orchestration provides workflow orchestration for automated tool chaining.
//
// This package enables automatic coordination of multiple reasoning tools to execute
// complex analysis workflows without manual intervention. Workflows can be sequential
// or parallel, with conditional execution based on intermediate results.
package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkflowType defines the type of workflow execution pattern
type WorkflowType string

const (
	WorkflowSequential WorkflowType = "sequential" // Execute steps in order
	WorkflowParallel   WorkflowType = "parallel"   // Execute steps concurrently
	WorkflowConditional WorkflowType = "conditional" // Execute based on conditions
)

// Workflow represents a coordinated sequence of tool executions
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        WorkflowType           `json:"type"`
	Steps       []*WorkflowStep        `json:"steps"`
	Context     *ReasoningContext      `json:"context,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Tool        string                 `json:"tool"`         // Tool name to execute
	Input       map[string]interface{} `json:"input"`        // Tool input parameters
	DependsOn   []string               `json:"depends_on,omitempty"` // Step IDs this depends on
	Condition   *StepCondition         `json:"condition,omitempty"`  // Conditional execution
	Transform   *OutputTransform       `json:"transform,omitempty"`  // Output transformation
	StoreAs     string                 `json:"store_as,omitempty"`   // Store result in context
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StepCondition defines when a step should execute
type StepCondition struct {
	Type       string  `json:"type"`        // "confidence_threshold", "result_match", etc.
	Field      string  `json:"field,omitempty"`
	Operator   string  `json:"operator"`    // "gt", "lt", "eq", "contains", etc.
	Value      interface{} `json:"value"`
}

// OutputTransform defines how to transform step output
type OutputTransform struct {
	Type   string                 `json:"type"`   // "extract_field", "map", "filter"
	Config map[string]interface{} `json:"config"` // Transform configuration
}

// ReasoningContext tracks shared state across workflow execution
type ReasoningContext struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflow_id"`
	Problem      string                 `json:"problem"`
	Results      map[string]interface{} `json:"results"`       // Step results
	Thoughts     []string               `json:"thoughts"`      // Thought IDs
	CausalGraphs []string               `json:"causal_graphs"` // Causal graph IDs
	Beliefs      []string               `json:"beliefs"`       // Belief IDs
	Evidence     []string               `json:"evidence"`      // Evidence IDs
	Decisions    []string               `json:"decisions"`     // Decision IDs
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// WorkflowResult contains the outcome of workflow execution
type WorkflowResult struct {
	WorkflowID   string                 `json:"workflow_id"`
	Status       string                 `json:"status"` // "success", "partial", "failed"
	StepResults  map[string]interface{} `json:"step_results"`
	Context      *ReasoningContext      `json:"context"`
	Duration     time.Duration          `json:"duration"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Orchestrator manages workflow execution and coordination
type Orchestrator struct {
	workflows map[string]*Workflow
	contexts  map[string]*ReasoningContext
	mu        sync.RWMutex
}

// NewOrchestrator creates a new workflow orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		workflows: make(map[string]*Workflow),
		contexts:  make(map[string]*ReasoningContext),
	}
}

// RegisterWorkflow adds a new workflow to the orchestrator
func (o *Orchestrator) RegisterWorkflow(workflow *Workflow) error {
	if workflow.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.workflows[workflow.ID]; exists {
		return fmt.Errorf("workflow %s already exists", workflow.ID)
	}

	workflow.CreatedAt = time.Now()
	o.workflows[workflow.ID] = workflow
	return nil
}

// GetWorkflow retrieves a workflow by ID
func (o *Orchestrator) GetWorkflow(id string) (*Workflow, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	workflow, exists := o.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}

	return workflow, nil
}

// ListWorkflows returns all registered workflows
func (o *Orchestrator) ListWorkflows() []*Workflow {
	o.mu.RLock()
	defer o.mu.RUnlock()

	workflows := make([]*Workflow, 0, len(o.workflows))
	for _, w := range o.workflows {
		workflows = append(workflows, w)
	}
	return workflows
}

// CreateContext creates a new reasoning context for workflow execution
func (o *Orchestrator) CreateContext(workflowID, problem string) *ReasoningContext {
	ctx := &ReasoningContext{
		ID:           fmt.Sprintf("ctx_%d", time.Now().UnixNano()),
		WorkflowID:   workflowID,
		Problem:      problem,
		Results:      make(map[string]interface{}),
		Thoughts:     []string{},
		CausalGraphs: []string{},
		Beliefs:      []string{},
		Evidence:     []string{},
		Decisions:    []string{},
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	o.mu.Lock()
	o.contexts[ctx.ID] = ctx
	o.mu.Unlock()

	return ctx
}

// GetContext retrieves a reasoning context by ID
func (o *Orchestrator) GetContext(id string) (*ReasoningContext, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	ctx, exists := o.contexts[id]
	if !exists {
		return nil, fmt.Errorf("context %s not found", id)
	}

	return ctx, nil
}

// UpdateContext updates a reasoning context
func (o *Orchestrator) UpdateContext(ctx *ReasoningContext) error {
	if ctx == nil || ctx.ID == "" {
		return fmt.Errorf("invalid context")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	ctx.UpdatedAt = time.Now()
	o.contexts[ctx.ID] = ctx
	return nil
}

// ExecuteWorkflow executes a workflow with the given input
func (o *Orchestrator) ExecuteWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (*WorkflowResult, error) {
	startTime := time.Now()

	workflow, err := o.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	// Create reasoning context
	problem, _ := input["problem"].(string)
	reasoningCtx := o.CreateContext(workflowID, problem)

	result := &WorkflowResult{
		WorkflowID:  workflowID,
		StepResults: make(map[string]interface{}),
		Context:     reasoningCtx,
		Metadata:    make(map[string]interface{}),
	}

	// Execute based on workflow type
	switch workflow.Type {
	case WorkflowSequential:
		err = o.executeSequential(ctx, workflow, input, reasoningCtx, result)
	case WorkflowParallel:
		err = o.executeParallel(ctx, workflow, input, reasoningCtx, result)
	case WorkflowConditional:
		err = o.executeConditional(ctx, workflow, input, reasoningCtx, result)
	default:
		return nil, fmt.Errorf("unknown workflow type: %s", workflow.Type)
	}

	result.Duration = time.Since(startTime)

	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = err.Error()
		return result, err
	}

	result.Status = "success"
	return result, nil
}

// executeSequential executes workflow steps in sequence
func (o *Orchestrator) executeSequential(ctx context.Context, workflow *Workflow, input map[string]interface{}, reasoningCtx *ReasoningContext, result *WorkflowResult) error {
	for _, step := range workflow.Steps {
		// Check condition if present
		if step.Condition != nil && !o.evaluateCondition(step.Condition, reasoningCtx) {
			continue
		}

		// Execute step
		stepResult, err := o.executeStep(ctx, step, input, reasoningCtx)
		if err != nil {
			return fmt.Errorf("step %s failed: %w", step.ID, err)
		}

		// Store result
		if step.StoreAs != "" {
			reasoningCtx.Results[step.StoreAs] = stepResult
			o.UpdateContext(reasoningCtx)
		}

		result.StepResults[step.ID] = stepResult
	}

	return nil
}

// executeParallel executes workflow steps in parallel
func (o *Orchestrator) executeParallel(ctx context.Context, workflow *Workflow, input map[string]interface{}, reasoningCtx *ReasoningContext, result *WorkflowResult) error {
	var wg sync.WaitGroup
	errors := make(chan error, len(workflow.Steps))
	stepResults := make(map[string]interface{})
	var resultMu sync.Mutex

	for _, step := range workflow.Steps {
		wg.Add(1)
		go func(s *WorkflowStep) {
			defer wg.Done()

			// Check condition if present
			if s.Condition != nil && !o.evaluateCondition(s.Condition, reasoningCtx) {
				return
			}

			// Execute step
			stepResult, err := o.executeStep(ctx, s, input, reasoningCtx)
			if err != nil {
				errors <- fmt.Errorf("step %s failed: %w", s.ID, err)
				return
			}

			// Store result
			resultMu.Lock()
			if s.StoreAs != "" {
				reasoningCtx.Results[s.StoreAs] = stepResult
			}
			stepResults[s.ID] = stepResult
			resultMu.Unlock()
		}(step)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	// Update results
	for id, res := range stepResults {
		result.StepResults[id] = res
	}

	o.UpdateContext(reasoningCtx)
	return nil
}

// executeConditional executes workflow with conditional branching
func (o *Orchestrator) executeConditional(ctx context.Context, workflow *Workflow, input map[string]interface{}, reasoningCtx *ReasoningContext, result *WorkflowResult) error {
	// Build dependency graph
	dependencies := make(map[string][]string)
	for _, step := range workflow.Steps {
		dependencies[step.ID] = step.DependsOn
	}

	// Execute steps respecting dependencies
	executed := make(map[string]bool)
	for len(executed) < len(workflow.Steps) {
		progress := false

		for _, step := range workflow.Steps {
			if executed[step.ID] {
				continue
			}

			// Check if dependencies are met
			canExecute := true
			for _, dep := range step.DependsOn {
				if !executed[dep] {
					canExecute = false
					break
				}
			}

			if !canExecute {
				continue
			}

			// Check condition
			if step.Condition != nil && !o.evaluateCondition(step.Condition, reasoningCtx) {
				executed[step.ID] = true
				progress = true
				continue
			}

			// Execute step
			stepResult, err := o.executeStep(ctx, step, input, reasoningCtx)
			if err != nil {
				return fmt.Errorf("step %s failed: %w", step.ID, err)
			}

			// Store result
			if step.StoreAs != "" {
				reasoningCtx.Results[step.StoreAs] = stepResult
				o.UpdateContext(reasoningCtx)
			}

			result.StepResults[step.ID] = stepResult
			executed[step.ID] = true
			progress = true
		}

		if !progress {
			return fmt.Errorf("workflow deadlock: circular dependencies or all steps blocked by conditions")
		}
	}

	return nil
}

// executeStep executes a single workflow step (stub - to be implemented with actual tool calls)
func (o *Orchestrator) executeStep(ctx context.Context, step *WorkflowStep, input map[string]interface{}, reasoningCtx *ReasoningContext) (interface{}, error) {
	// This is a stub - actual implementation will call the appropriate tool handler
	// For now, we just return a placeholder
	return map[string]interface{}{
		"step_id": step.ID,
		"tool":    step.Tool,
		"status":  "executed",
	}, nil
}

// evaluateCondition checks if a step condition is met
func (o *Orchestrator) evaluateCondition(condition *StepCondition, ctx *ReasoningContext) bool {
	// Extract value from context or result
	var value interface{}

	if condition.Field != "" {
		value = ctx.Results[condition.Field]
	}

	// Evaluate based on operator
	switch condition.Operator {
	case "gt":
		if v, ok := value.(float64); ok {
			threshold, _ := condition.Value.(float64)
			return v > threshold
		}
	case "lt":
		if v, ok := value.(float64); ok {
			threshold, _ := condition.Value.(float64)
			return v < threshold
		}
	case "eq":
		return value == condition.Value
	case "contains":
		if v, ok := value.(string); ok {
			search, _ := condition.Value.(string)
			return len(search) > 0 && len(v) > 0 && contains(v, search)
		}
	}

	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// PredefinedWorkflows contains commonly used workflow templates
var PredefinedWorkflows = map[string]*Workflow{
	"comprehensive-analysis": {
		ID:          "comprehensive-analysis",
		Name:        "Comprehensive Analysis",
		Description: "Full analysis pipeline with causal, temporal, and probabilistic reasoning",
		Type:        WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:      "decompose",
				Tool:    "decompose-problem",
				Input:   map[string]interface{}{},
				StoreAs: "decomposition",
			},
			{
				ID:      "causal",
				Tool:    "build-causal-graph",
				Input:   map[string]interface{}{},
				StoreAs: "causal_graph",
			},
			{
				ID:      "temporal",
				Tool:    "analyze-temporal",
				Input:   map[string]interface{}{},
				StoreAs: "temporal_analysis",
			},
			{
				ID:      "perspectives",
				Tool:    "analyze-perspectives",
				Input:   map[string]interface{}{},
				StoreAs: "perspectives",
			},
			{
				ID:      "synthesis",
				Tool:    "synthesize-insights",
				Input:   map[string]interface{}{},
				StoreAs: "synthesis",
			},
			{
				ID:      "decision",
				Tool:    "make-decision",
				Input:   map[string]interface{}{},
				StoreAs: "decision",
			},
		},
	},
	"validation-pipeline": {
		ID:          "validation-pipeline",
		Name:        "Validation Pipeline",
		Description: "Thorough validation with bias detection and self-evaluation",
		Type:        WorkflowSequential,
		Steps: []*WorkflowStep{
			{
				ID:      "detect-contradictions",
				Tool:    "detect-contradictions",
				Input:   map[string]interface{}{},
				StoreAs: "contradictions",
			},
			{
				ID:      "detect-biases",
				Tool:    "detect-biases",
				Input:   map[string]interface{}{},
				StoreAs: "biases",
			},
			{
				ID:      "self-evaluate",
				Tool:    "self-evaluate",
				Input:   map[string]interface{}{},
				StoreAs: "evaluation",
			},
			{
				ID:      "sensitivity",
				Tool:    "sensitivity-analysis",
				Input:   map[string]interface{}{},
				StoreAs: "sensitivity",
			},
		},
	},
}
