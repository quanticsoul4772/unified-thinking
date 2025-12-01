// Package streaming provides MCP progress notification support for long-running tools.
//
// This package implements progress reporting using the standard MCP notifications/progress
// mechanism, enabling real-time updates during tool execution while maintaining backward
// compatibility with clients that don't support progress notifications.
//
// Key components:
//   - ProgressUpdate: Structured progress data sent via notifications
//   - StreamingConfig: Per-tool configuration for streaming behavior
//   - ProgressReporter: Interface for reporting progress from handlers
//   - NotifyingReporter: Implementation using MCP ServerSession.NotifyProgress
package streaming

import (
	"time"
)

// ProgressUpdate represents a single progress notification payload.
// This structure is sent via MCP notifications/progress to inform the client
// about the current state of a long-running operation.
type ProgressUpdate struct {
	// Step is the current step number (1-indexed for display)
	Step int `json:"step"`

	// TotalSteps is the total number of steps, or 0 if unknown
	TotalSteps int `json:"total_steps"`

	// StepName is a short identifier for the current step
	StepName string `json:"step_name"`

	// Percentage is the completion percentage (0.0-100.0), or -1 if unknown
	Percentage float64 `json:"percentage"`

	// Message is a human-readable status message
	Message string `json:"message"`

	// PartialData contains intermediate results for progressive rendering (optional)
	PartialData any `json:"partial_data,omitempty"`

	// Timestamp is when this update was generated
	Timestamp time.Time `json:"timestamp"`
}

// StreamingConfig provides per-tool configuration for streaming behavior.
type StreamingConfig struct {
	// Enabled indicates whether streaming is enabled for this tool
	Enabled bool `json:"enabled"`

	// MinInterval is the minimum time between notifications (debouncing)
	// Use this to prevent flooding the client with too many updates
	MinInterval time.Duration `json:"min_interval"`

	// SendPartialData indicates whether to include intermediate results
	SendPartialData bool `json:"send_partial_data"`

	// AutoProgress indicates whether to automatically calculate percentage
	// from step/totalSteps when not explicitly provided
	AutoProgress bool `json:"auto_progress"`
}

// DefaultConfig returns the default streaming configuration.
func DefaultConfig() StreamingConfig {
	return StreamingConfig{
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: false,
		AutoProgress:    true,
	}
}

// ConfigOption is a functional option for customizing StreamingConfig.
type ConfigOption func(*StreamingConfig)

// WithMinInterval sets the minimum interval between notifications.
func WithMinInterval(d time.Duration) ConfigOption {
	return func(c *StreamingConfig) {
		c.MinInterval = d
	}
}

// WithPartialData enables or disables partial data in progress updates.
func WithPartialData(enabled bool) ConfigOption {
	return func(c *StreamingConfig) {
		c.SendPartialData = enabled
	}
}

// WithAutoProgress enables or disables automatic percentage calculation.
func WithAutoProgress(enabled bool) ConfigOption {
	return func(c *StreamingConfig) {
		c.AutoProgress = enabled
	}
}

// NewConfig creates a StreamingConfig with the given options.
func NewConfig(opts ...ConfigOption) StreamingConfig {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// ToolConfigs defines default streaming configurations for streaming-enabled tools.
var ToolConfigs = map[string]StreamingConfig{
	// P0 - Essential (Day 6)
	"execute-workflow": {
		Enabled:         true,
		MinInterval:     50 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	},
	"run-preset": {
		Enabled:         true,
		MinInterval:     50 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	},
	"got-generate": {
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	},

	// P1 - Important (Day 7)
	"got-aggregate": {
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	},
	"think": {
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: false, // Tree mode branches may be large
		AutoProgress:    true,
	},
	"perform-cbr-cycle": {
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	},

	// P2 - Enhancement (Day 8)
	"synthesize-insights": {
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: false,
		AutoProgress:    true,
	},
	"analyze-perspectives": {
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	},
	"build-causal-graph": {
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	},
	"evaluate-hypotheses": {
		Enabled:         true,
		MinInterval:     100 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	},
}

// GetToolConfig returns the streaming configuration for a tool.
// Returns DefaultConfig() if the tool doesn't have a specific configuration.
func GetToolConfig(toolName string) StreamingConfig {
	if cfg, ok := ToolConfigs[toolName]; ok {
		return cfg
	}
	// Return a disabled default for tools not in the streaming list
	cfg := DefaultConfig()
	cfg.Enabled = false
	return cfg
}

// IsStreamingEnabled checks if streaming is enabled for a tool.
func IsStreamingEnabled(toolName string) bool {
	cfg, ok := ToolConfigs[toolName]
	return ok && cfg.Enabled
}
