package streaming

import (
	"context"
	"log"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey int

const (
	// reporterKey is the context key for the ProgressReporter.
	reporterKey contextKey = iota

	// configKey is the context key for StreamingConfig.
	configKey

	// toolNameKey is the context key for the current tool name.
	toolNameKey
)

// WithReporter returns a new context with the given ProgressReporter.
func WithReporter(ctx context.Context, reporter ProgressReporter) context.Context {
	return context.WithValue(ctx, reporterKey, reporter)
}

// GetReporter retrieves the ProgressReporter from the context.
// Returns a DefaultReporter if no reporter is set, ensuring handlers
// can always call progress methods safely.
func GetReporter(ctx context.Context) ProgressReporter {
	if reporter, ok := ctx.Value(reporterKey).(ProgressReporter); ok {
		return reporter
	}
	return NewDefaultReporter(nil)
}

// WithConfig returns a new context with the given StreamingConfig.
func WithConfig(ctx context.Context, config StreamingConfig) context.Context {
	return context.WithValue(ctx, configKey, config)
}

// GetConfig retrieves the StreamingConfig from the context.
// Returns DefaultConfig() if not set.
func GetConfig(ctx context.Context) StreamingConfig {
	if cfg, ok := ctx.Value(configKey).(StreamingConfig); ok {
		return cfg
	}
	return DefaultConfig()
}

// WithToolName returns a new context with the given tool name.
func WithToolName(ctx context.Context, toolName string) context.Context {
	return context.WithValue(ctx, toolNameKey, toolName)
}

// GetToolName retrieves the tool name from the context.
// Returns empty string if not set.
func GetToolName(ctx context.Context) string {
	if name, ok := ctx.Value(toolNameKey).(string); ok {
		return name
	}
	return ""
}

// StreamingContext bundles all streaming-related context values.
// This is a convenience struct for cases where you need to pass
// multiple streaming values together.
type StreamingContext struct {
	Reporter ProgressReporter
	Config   StreamingConfig
	ToolName string
}

// NewStreamingContext creates a StreamingContext with the given values.
func NewStreamingContext(reporter ProgressReporter, config StreamingConfig, toolName string) *StreamingContext {
	return &StreamingContext{
		Reporter: reporter,
		Config:   config,
		ToolName: toolName,
	}
}

// FromContext extracts a StreamingContext from a context.Context.
func FromContext(ctx context.Context) *StreamingContext {
	return &StreamingContext{
		Reporter: GetReporter(ctx),
		Config:   GetConfig(ctx),
		ToolName: GetToolName(ctx),
	}
}

// ToContext adds all StreamingContext values to a context.Context.
func (sc *StreamingContext) ToContext(ctx context.Context) context.Context {
	ctx = WithReporter(ctx, sc.Reporter)
	ctx = WithConfig(ctx, sc.Config)
	ctx = WithToolName(ctx, sc.ToolName)
	return ctx
}

// IsEnabled returns whether streaming is enabled in this context.
func (sc *StreamingContext) IsEnabled() bool {
	return sc.Config.Enabled && sc.Reporter != nil && sc.Reporter.IsEnabled()
}

// ReportProgress is a convenience method to report progress.
func (sc *StreamingContext) ReportProgress(current, total float64, message string) error {
	if sc.Reporter == nil {
		return nil
	}
	return sc.Reporter.ReportProgress(current, total, message)
}

// ReportStep is a convenience method to report step progress.
func (sc *StreamingContext) ReportStep(step, totalSteps int, stepName, message string) error {
	if sc.Reporter == nil {
		return nil
	}
	return sc.Reporter.ReportStep(step, totalSteps, stepName, message)
}

// ReportPartialResult is a convenience method to report partial results.
func (sc *StreamingContext) ReportPartialResult(stepName string, data any) error {
	if sc.Reporter == nil {
		return nil
	}
	return sc.Reporter.ReportPartialResult(stepName, data)
}

// ProgressError wraps a progress reporting error with context.
type ProgressError struct {
	Operation string
	Err       error
}

func (e *ProgressError) Error() string {
	return "streaming: " + e.Operation + " failed: " + e.Err.Error()
}

func (e *ProgressError) Unwrap() error {
	return e.Err
}

// CheckReport returns a ProgressError if the error is non-nil.
// This provides a safe alternative to panicking on progress failures.
// Example: if err := CheckReport("step", reporter.ReportStep(...)); err != nil { log error }
func CheckReport(operation string, err error) error {
	if err != nil {
		return &ProgressError{Operation: operation, Err: err}
	}
	return nil
}

// LogReportError logs a progress reporting error if non-nil.
// This is the recommended way to handle non-critical progress failures.
// Returns true if an error occurred (for conditional logic).
func LogReportError(err error) bool {
	if err != nil {
		// Log to stderr - progress failures shouldn't crash the server
		// but should be visible for debugging
		log.Printf("[streaming] progress report failed: %v", err)
		return true
	}
	return false
}
