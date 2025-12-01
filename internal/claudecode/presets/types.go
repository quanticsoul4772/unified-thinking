// Package presets provides built-in workflow presets for common development tasks.
//
// Presets are predefined sequences of tool invocations that implement best practices
// for common tasks like code review, debugging, and architecture decisions.
package presets

// WorkflowPreset defines a reusable workflow pattern
type WorkflowPreset struct {
	// ID is the unique identifier for the preset
	ID string `json:"id"`
	// Name is the human-readable name
	Name string `json:"name"`
	// Description explains what the preset does
	Description string `json:"description"`
	// Category groups related presets (code, architecture, research, etc.)
	Category string `json:"category"`
	// Steps are the ordered tool invocations
	Steps []PresetStep `json:"steps"`
	// InputSchema describes required/optional inputs
	InputSchema map[string]ParamSpec `json:"input_schema"`
	// OutputFormat describes the expected output structure
	OutputFormat string `json:"output_format"`
	// EstimatedTime is the expected execution time
	EstimatedTime string `json:"estimated_time"`
	// Tags for searchability
	Tags []string `json:"tags,omitempty"`
}

// PresetStep represents a single step in a preset workflow
type PresetStep struct {
	// StepID is the unique identifier within the preset
	StepID string `json:"step_id"`
	// Tool is the tool to invoke
	Tool string `json:"tool"`
	// Description explains what this step does
	Description string `json:"description"`
	// InputMap maps preset inputs to tool parameters
	InputMap map[string]string `json:"input_map"`
	// StaticInputs are fixed parameters for this step
	StaticInputs map[string]any `json:"static_inputs,omitempty"`
	// Condition defines when to execute this step
	Condition *StepCondition `json:"condition,omitempty"`
	// StoreAs saves the result with this key for later steps
	StoreAs string `json:"store_as,omitempty"`
	// DependsOn lists step IDs that must complete first
	DependsOn []string `json:"depends_on,omitempty"`
	// Optional if true, failures don't stop the workflow
	Optional bool `json:"optional,omitempty"`
}

// StepCondition defines conditional execution
type StepCondition struct {
	// Type is the condition type (confidence_threshold, result_match, etc.)
	Type string `json:"type"`
	// Field is the field to check from previous step
	Field string `json:"field,omitempty"`
	// Operator is the comparison operator (gt, lt, eq, contains, etc.)
	Operator string `json:"operator"`
	// Value is the value to compare against
	Value any `json:"value"`
	// SourceStep is the step to get the field from
	SourceStep string `json:"source_step,omitempty"`
}

// ParamSpec describes a preset parameter
type ParamSpec struct {
	// Type is the parameter type (string, number, boolean, array, object)
	Type string `json:"type"`
	// Required indicates if the parameter must be provided
	Required bool `json:"required"`
	// Default is the default value if not provided
	Default any `json:"default,omitempty"`
	// Description explains the parameter
	Description string `json:"description"`
	// Examples show valid values
	Examples []any `json:"examples,omitempty"`
}

// PresetResult contains the result of executing a preset
type PresetResult struct {
	// PresetID is the executed preset
	PresetID string `json:"preset_id"`
	// StepsCompleted is the number of steps completed
	StepsCompleted int `json:"steps_completed"`
	// StepsTotal is the total number of steps
	StepsTotal int `json:"steps_total"`
	// Results contains the result of each step
	Results []StepResult `json:"results"`
	// FinalOutput is the aggregated output
	FinalOutput any `json:"final_output,omitempty"`
	// Status is success, partial, or failed
	Status string `json:"status"`
	// DurationMs is the total execution time
	DurationMs int64 `json:"duration_ms"`
	// Error contains error details if failed
	Error string `json:"error,omitempty"`
}

// StepResult contains the result of a single step
type StepResult struct {
	// Step is the step number (1-based)
	Step int `json:"step"`
	// StepID is the step identifier
	StepID string `json:"step_id"`
	// Tool is the tool that was executed
	Tool string `json:"tool"`
	// Result is the tool's response
	Result any `json:"result"`
	// DurationMs is the step execution time
	DurationMs int64 `json:"duration_ms"`
	// Status is success or failed
	Status string `json:"status"`
	// Error contains error details if failed
	Error string `json:"error,omitempty"`
	// Skipped indicates if the step was skipped due to condition
	Skipped bool `json:"skipped,omitempty"`
}

// PresetSummary is a brief description for listing
type PresetSummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	StepCount     int    `json:"step_count"`
	EstimatedTime string `json:"estimated_time"`
}

// ToSummary creates a PresetSummary from a WorkflowPreset
func (p *WorkflowPreset) ToSummary() PresetSummary {
	return PresetSummary{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Category:      p.Category,
		StepCount:     len(p.Steps),
		EstimatedTime: p.EstimatedTime,
	}
}
