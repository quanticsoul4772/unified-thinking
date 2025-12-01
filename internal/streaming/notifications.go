package streaming

import (
	"context"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NotifyingReporter implements ProgressReporter using MCP ServerSession.NotifyProgress.
// It sends progress notifications to the connected client in real-time.
type NotifyingReporter struct {
	session       *mcp.ServerSession
	progressToken any
	config        StreamingConfig
	toolName      string

	// Rate limiting state
	lastReport time.Time
	mu         sync.Mutex

	// Step tracking for auto-progress
	currentStep int
	totalSteps  int
}

// NewNotifyingReporter creates a reporter that sends MCP progress notifications.
// session is the active ServerSession, progressToken is from CallToolParamsRaw.GetProgressToken(),
// and config controls streaming behavior.
func NewNotifyingReporter(session *mcp.ServerSession, progressToken any, config StreamingConfig, toolName string) *NotifyingReporter {
	return &NotifyingReporter{
		session:       session,
		progressToken: progressToken,
		config:        config,
		toolName:      toolName,
	}
}

// ReportProgress sends a progress notification with current/total values.
func (r *NotifyingReporter) ReportProgress(current, total float64, message string) error {
	if !r.shouldSend() {
		return nil
	}

	r.mu.Lock()
	r.lastReport = time.Now()
	r.mu.Unlock()

	params := &mcp.ProgressNotificationParams{
		ProgressToken: r.progressToken,
		Progress:      current,
		Total:         total,
		Message:       message,
	}

	return r.session.NotifyProgress(context.Background(), params)
}

// ReportStep sends a step-based progress notification.
func (r *NotifyingReporter) ReportStep(step int, totalSteps int, stepName string, message string) error {
	r.mu.Lock()
	stepChanged := step != r.currentStep || totalSteps != r.totalSteps
	r.currentStep = step
	r.totalSteps = totalSteps
	r.mu.Unlock()

	// Always report step changes, even if rate limited
	if !stepChanged && !r.shouldSend() {
		return nil
	}

	r.mu.Lock()
	r.lastReport = time.Now()
	r.mu.Unlock()

	// Build message with step info
	fullMessage := message
	if stepName != "" {
		if message != "" {
			fullMessage = stepName + ": " + message
		} else {
			fullMessage = stepName
		}
	}

	// Calculate progress values
	var progress, total float64
	if totalSteps > 0 {
		progress = float64(step)
		total = float64(totalSteps)
	} else {
		// Indeterminate progress
		progress = float64(step)
		total = 0
	}

	params := &mcp.ProgressNotificationParams{
		ProgressToken: r.progressToken,
		Progress:      progress,
		Total:         total,
		Message:       fullMessage,
	}

	return r.session.NotifyProgress(context.Background(), params)
}

// ReportPartialResult sends intermediate results if enabled.
// Note: MCP progress notifications don't have a dedicated field for partial data,
// so we include it in the message as JSON or skip if not configured.
func (r *NotifyingReporter) ReportPartialResult(stepName string, data any) error {
	if !r.config.SendPartialData {
		return nil
	}

	if !r.shouldSend() {
		return nil
	}

	r.mu.Lock()
	r.lastReport = time.Now()
	step := r.currentStep
	totalSteps := r.totalSteps
	r.mu.Unlock()

	// For partial results, we use the current step progress
	var progress, total float64
	if totalSteps > 0 {
		progress = float64(step)
		total = float64(totalSteps)
	} else {
		progress = float64(step)
		total = 0
	}

	message := stepName + " (partial result available)"

	params := &mcp.ProgressNotificationParams{
		ProgressToken: r.progressToken,
		Progress:      progress,
		Total:         total,
		Message:       message,
	}

	return r.session.NotifyProgress(context.Background(), params)
}

// IsEnabled returns true if the reporter has a valid session and progress token.
func (r *NotifyingReporter) IsEnabled() bool {
	return r.session != nil && r.progressToken != nil && r.config.Enabled
}

// GetProgressToken returns the client's progress token.
func (r *NotifyingReporter) GetProgressToken() any {
	return r.progressToken
}

// shouldSend checks if we should send a notification based on rate limiting.
func (r *NotifyingReporter) shouldSend() bool {
	if !r.IsEnabled() {
		return false
	}

	if r.config.MinInterval == 0 {
		return true
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return time.Since(r.lastReport) >= r.config.MinInterval
}

// CreateReporter creates the appropriate ProgressReporter based on the request context.
// If the client provided a progress token and the tool supports streaming, returns a
// NotifyingReporter. Otherwise, returns a DefaultReporter (no-op).
//
// This is the main entry point for handlers to get a reporter.
func CreateReporter(req *mcp.CallToolRequest, toolName string) ProgressReporter {
	config := GetToolConfig(toolName)

	// If streaming is not enabled for this tool, return a no-op reporter
	if !config.Enabled {
		return NewDefaultReporter(nil)
	}

	// Check if the client provided a progress token
	var progressToken any
	if req != nil && req.Params != nil {
		progressToken = req.Params.GetProgressToken()
	}

	// If no progress token, streaming is not requested by the client
	if progressToken == nil {
		return NewDefaultReporter(nil)
	}

	// Check if we have a valid session
	session := req.Session
	if session == nil {
		return NewDefaultReporter(progressToken)
	}

	// Create the notifying reporter
	return NewNotifyingReporter(session, progressToken, config, toolName)
}

// CreateReporterWithConfig creates a reporter with a custom configuration.
func CreateReporterWithConfig(req *mcp.CallToolRequest, toolName string, config StreamingConfig) ProgressReporter {
	// Check if the client provided a progress token
	var progressToken any
	if req != nil && req.Params != nil {
		progressToken = req.Params.GetProgressToken()
	}

	// If no progress token or streaming disabled, return no-op
	if progressToken == nil || !config.Enabled {
		return NewDefaultReporter(progressToken)
	}

	// Check if we have a valid session
	session := req.Session
	if session == nil {
		return NewDefaultReporter(progressToken)
	}

	return NewNotifyingReporter(session, progressToken, config, toolName)
}

// InjectReporter creates a reporter and injects it into the context.
// This is a convenience function for handlers that want both the reporter
// and a context with the reporter pre-injected.
func InjectReporter(ctx context.Context, req *mcp.CallToolRequest, toolName string) (context.Context, ProgressReporter) {
	reporter := CreateReporter(req, toolName)
	config := GetToolConfig(toolName)

	ctx = WithReporter(ctx, reporter)
	ctx = WithConfig(ctx, config)
	ctx = WithToolName(ctx, toolName)

	return ctx, reporter
}
