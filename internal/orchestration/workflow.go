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
	executor  ToolExecutor // Add executor for tool execution
	mu        sync.RWMutex
}

// NewOrchestrator creates a new workflow orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		workflows: make(map[string]*Workflow),
		contexts:  make(map[string]*ReasoningContext),
		executor:  nil, // Executor must be set via SetExecutor
	}
}

// NewOrchestratorWithExecutor creates a new workflow orchestrator with a tool executor
func NewOrchestratorWithExecutor(executor ToolExecutor) *Orchestrator {
	return &Orchestrator{
		workflows: make(map[string]*Workflow),
		contexts:  make(map[string]*ReasoningContext),
		executor:  executor,
	}
}

// SetExecutor sets the tool executor for the orchestrator
func (o *Orchestrator) SetExecutor(executor ToolExecutor) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.executor = executor
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
			_ = o.UpdateContext(reasoningCtx)
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

	_ = o.UpdateContext(reasoningCtx)
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
				_ = o.UpdateContext(reasoningCtx)
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

// executeStep executes a single workflow step using the tool executor
func (o *Orchestrator) executeStep(ctx context.Context, step *WorkflowStep, input map[string]interface{}, reasoningCtx *ReasoningContext) (interface{}, error) {
	// Check if executor is available
	if o.executor == nil {
		return nil, fmt.Errorf("no tool executor configured for orchestrator")
	}

	// Prepare step input by merging workflow input with step-specific input
	toolInput := make(map[string]interface{})

	// Start with workflow input
	for k, v := range input {
		toolInput[k] = v
	}

	// Override with step-specific input
	for k, v := range step.Input {
		// Resolve template references (supports both {{variable}} and $variable syntax)
		resolvedValue := o.resolveTemplateValue(v, input, reasoningCtx)
		toolInput[k] = resolvedValue
	}

	// Execute the tool
	result, err := o.executor.ExecuteTool(ctx, step.Tool, toolInput)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool %s: %w", step.Tool, err)
	}

	// Apply output transformation if specified
	if step.Transform != nil {
		result = applyTransform(result, step.Transform)
	}

	// Update context with tool-specific results
	switch step.Tool {
	case "think":
		if thoughtResult, ok := result.(map[string]interface{}); ok {
			if thoughtID, ok := thoughtResult["thought_id"].(string); ok {
				reasoningCtx.Thoughts = append(reasoningCtx.Thoughts, thoughtID)
			}
		}
	case "build-causal-graph":
		if graphResult, ok := result.(map[string]interface{}); ok {
			if graphID, ok := graphResult["id"].(string); ok {
				reasoningCtx.CausalGraphs = append(reasoningCtx.CausalGraphs, graphID)
			}
		}
	case "probabilistic-reasoning":
		if beliefResult, ok := result.(map[string]interface{}); ok {
			if beliefID, ok := beliefResult["id"].(string); ok {
				reasoningCtx.Beliefs = append(reasoningCtx.Beliefs, beliefID)
			}
		}
	case "assess-evidence":
		if evidenceResult, ok := result.(map[string]interface{}); ok {
			if evidenceID, ok := evidenceResult["id"].(string); ok {
				reasoningCtx.Evidence = append(reasoningCtx.Evidence, evidenceID)
			}
		}
	case "make-decision":
		if decisionResult, ok := result.(map[string]interface{}); ok {
			if decisionID, ok := decisionResult["id"].(string); ok {
				reasoningCtx.Decisions = append(reasoningCtx.Decisions, decisionID)
			}
		}
	}

	// Update overall confidence based on tool results
	if confidenceVal := extractConfidence(result); confidenceVal > 0 {
		// Weighted average of confidence
		if reasoningCtx.Confidence == 0 {
			reasoningCtx.Confidence = confidenceVal
		} else {
			reasoningCtx.Confidence = (reasoningCtx.Confidence + confidenceVal) / 2
		}
	}

	return result, nil
}

// resolveTemplateValue resolves template references in workflow step input values.
// Supports both {{variable}} and $variable syntax for backward compatibility.
// Also handles nested references like {{causal_graph.id}} or $causal_graph.id.
func (o *Orchestrator) resolveTemplateValue(value interface{}, workflowInput map[string]interface{}, reasoningCtx *ReasoningContext) interface{} {
	// Handle string values that might contain template references
	strVal, ok := value.(string)
	if !ok {
		// Not a string, check if it's a slice that might contain template strings
		if slice, ok := value.([]interface{}); ok {
			resolved := make([]interface{}, len(slice))
			for i, item := range slice {
				resolved[i] = o.resolveTemplateValue(item, workflowInput, reasoningCtx)
			}
			return resolved
		}
		// Return value as-is if not a string or slice
		return value
	}

	// Convert {{variable}} syntax to $variable for unified processing
	templateValue := convertTemplateToReference(strVal)

	// Handle $variable references
	if len(templateValue) > 0 && templateValue[0] == '$' {
		refKey := templateValue[1:] // Remove the $ prefix

		// First, check reasoning context results
		if val, exists := reasoningCtx.Results[refKey]; exists {
			return val
		}

		// Try to extract nested field (e.g., "causal_graph.id")
		parts := splitReference(refKey)
		if len(parts) > 1 {
			// Check reasoning context for nested values
			if rootVal, exists := reasoningCtx.Results[parts[0]]; exists {
				nestedVal := extractNestedValue(rootVal, parts[1:])
				if nestedVal != nil {
					return nestedVal
				}
			}

			// Check workflow input for nested values
			if rootVal, exists := workflowInput[parts[0]]; exists {
				nestedVal := extractNestedValue(rootVal, parts[1:])
				if nestedVal != nil {
					return nestedVal
				}
			}
		}

		// Check workflow input as fallback
		if val, exists := workflowInput[refKey]; exists {
			return val
		}

		// Reference not found - return original string to avoid silent failures
		// This helps with debugging workflow parameter issues
		return strVal
	}

	// No template reference found, return value as-is
	return strVal
}

// convertTemplateToReference converts {{variable}} syntax to $variable syntax.
// This normalizes template references for consistent processing.
// Examples:
//   - "{{problem}}" -> "$problem"
//   - "{{causal_graph.id}}" -> "$causal_graph.id"
//   - "regular string" -> "regular string"
//   - "$variable" -> "$variable" (already in correct format)
func convertTemplateToReference(template string) string {
	// Check if string uses {{variable}} syntax
	if len(template) > 4 && template[0:2] == "{{" && template[len(template)-2:] == "}}" {
		// Extract variable name and convert to $variable syntax
		variableName := template[2 : len(template)-2]
		return "$" + variableName
	}
	// Return as-is if not a template or already using $ syntax
	return template
}

// splitReference splits a reference string by dots
func splitReference(ref string) []string {
	parts := []string{}
	current := ""
	for _, ch := range ref {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// extractNestedValue extracts a nested value from an interface using a path
func extractNestedValue(val interface{}, path []string) interface{} {
	if len(path) == 0 {
		return val
	}

	switch v := val.(type) {
	case map[string]interface{}:
		if nextVal, exists := v[path[0]]; exists {
			return extractNestedValue(nextVal, path[1:])
		}
	}

	return nil
}

// applyTransform applies an output transformation to a result
func applyTransform(result interface{}, transform *OutputTransform) interface{} {
	switch transform.Type {
	case "extract_field":
		if field, ok := transform.Config["field"].(string); ok {
			if m, ok := result.(map[string]interface{}); ok {
				if val, exists := m[field]; exists {
					return val
				}
			}
		}
	case "map":
		// Apply a mapping function (simplified for now)
		if m, ok := result.(map[string]interface{}); ok {
			mapped := make(map[string]interface{})
			for k, v := range m {
				if mapConfig, ok := transform.Config[k]; ok {
					mapped[mapConfig.(string)] = v
				} else {
					mapped[k] = v
				}
			}
			return mapped
		}
	case "filter":
		// Filter results based on criteria
		if fields, ok := transform.Config["fields"].([]string); ok {
			if m, ok := result.(map[string]interface{}); ok {
				filtered := make(map[string]interface{})
				for _, field := range fields {
					if val, exists := m[field]; exists {
						filtered[field] = val
					}
				}
				return filtered
			}
		}
	}

	return result
}

// extractConfidence extracts confidence value from a result
func extractConfidence(result interface{}) float64 {
	switch v := result.(type) {
	case map[string]interface{}:
		if conf, ok := v["confidence"].(float64); ok {
			return conf
		}
		// Try other common confidence field names
		if conf, ok := v["probability"].(float64); ok {
			return conf
		}
		if conf, ok := v["score"].(float64); ok {
			return conf
		}
	}
	return 0
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
