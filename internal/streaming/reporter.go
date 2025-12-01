package streaming

import (
	"sync"
	"time"
)

// ProgressReporter is the interface for reporting progress during tool execution.
// Handlers should use this interface to send progress updates to the client.
type ProgressReporter interface {
	// ReportProgress sends a progress update with current/total and a message.
	// If total is 0, the progress is indeterminate.
	ReportProgress(current, total float64, message string) error

	// ReportStep sends a step-based progress update.
	// step is the current step (1-indexed), totalSteps is the total (0 if unknown).
	ReportStep(step int, totalSteps int, stepName string, message string) error

	// ReportPartialResult sends an intermediate result for progressive rendering.
	// Only sent if StreamingConfig.SendPartialData is true.
	ReportPartialResult(stepName string, data any) error

	// IsEnabled returns whether streaming is enabled for this reporter.
	// Handlers can check this to skip expensive progress generation.
	IsEnabled() bool

	// GetProgressToken returns the client's progress token, or nil if not provided.
	GetProgressToken() any
}

// DefaultReporter is a no-op implementation for when streaming is disabled.
// It implements all methods but does nothing, allowing handlers to call
// progress methods without checking if streaming is enabled.
type DefaultReporter struct {
	progressToken any
}

// NewDefaultReporter creates a no-op progress reporter.
func NewDefaultReporter(progressToken any) *DefaultReporter {
	return &DefaultReporter{
		progressToken: progressToken,
	}
}

// ReportProgress does nothing in the default reporter.
func (r *DefaultReporter) ReportProgress(current, total float64, message string) error {
	return nil
}

// ReportStep does nothing in the default reporter.
func (r *DefaultReporter) ReportStep(step int, totalSteps int, stepName string, message string) error {
	return nil
}

// ReportPartialResult does nothing in the default reporter.
func (r *DefaultReporter) ReportPartialResult(stepName string, data any) error {
	return nil
}

// IsEnabled returns false for the default reporter.
func (r *DefaultReporter) IsEnabled() bool {
	return false
}

// GetProgressToken returns the stored progress token.
func (r *DefaultReporter) GetProgressToken() any {
	return r.progressToken
}

// RateLimitedReporter wraps a ProgressReporter with rate limiting.
// It ensures that notifications are not sent more frequently than MinInterval.
type RateLimitedReporter struct {
	delegate    ProgressReporter
	config      StreamingConfig
	lastReport  time.Time
	mu          sync.Mutex
	currentStep int
	totalSteps  int
}

// NewRateLimitedReporter wraps a reporter with rate limiting.
func NewRateLimitedReporter(delegate ProgressReporter, config StreamingConfig) *RateLimitedReporter {
	return &RateLimitedReporter{
		delegate: delegate,
		config:   config,
	}
}

// ReportProgress sends a progress update if the rate limit allows.
func (r *RateLimitedReporter) ReportProgress(current, total float64, message string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.shouldReport() {
		return nil
	}

	r.lastReport = time.Now()
	return r.delegate.ReportProgress(current, total, message)
}

// ReportStep sends a step-based progress update if the rate limit allows.
func (r *RateLimitedReporter) ReportStep(step int, totalSteps int, stepName string, message string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Always report step changes even if rate limited
	if step != r.currentStep || totalSteps != r.totalSteps || r.shouldReport() {
		r.currentStep = step
		r.totalSteps = totalSteps
		r.lastReport = time.Now()
		return r.delegate.ReportStep(step, totalSteps, stepName, message)
	}

	return nil
}

// ReportPartialResult sends partial data if enabled and rate limit allows.
func (r *RateLimitedReporter) ReportPartialResult(stepName string, data any) error {
	if !r.config.SendPartialData {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.shouldReport() {
		return nil
	}

	r.lastReport = time.Now()
	return r.delegate.ReportPartialResult(stepName, data)
}

// IsEnabled returns whether the underlying reporter is enabled.
func (r *RateLimitedReporter) IsEnabled() bool {
	return r.delegate.IsEnabled()
}

// GetProgressToken returns the progress token from the underlying reporter.
func (r *RateLimitedReporter) GetProgressToken() any {
	return r.delegate.GetProgressToken()
}

// shouldReport checks if enough time has passed since the last report.
// Must be called with the mutex held.
func (r *RateLimitedReporter) shouldReport() bool {
	if r.config.MinInterval == 0 {
		return true
	}
	return time.Since(r.lastReport) >= r.config.MinInterval
}

// StepReporter is a convenience wrapper for step-based progress reporting.
// It automatically tracks step numbers and calculates percentages.
type StepReporter struct {
	reporter   ProgressReporter
	totalSteps int
	stepNames  []string
	current    int
	mu         sync.Mutex
}

// NewStepReporter creates a reporter for a known number of steps.
func NewStepReporter(reporter ProgressReporter, stepNames []string) *StepReporter {
	return &StepReporter{
		reporter:   reporter,
		totalSteps: len(stepNames),
		stepNames:  stepNames,
		current:    0,
	}
}

// StartStep begins a new step and reports progress.
func (r *StepReporter) StartStep(message string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.current++
	stepName := ""
	if r.current <= len(r.stepNames) {
		stepName = r.stepNames[r.current-1]
	}

	return r.reporter.ReportStep(r.current, r.totalSteps, stepName, message)
}

// CompleteStep marks the current step as complete without advancing.
func (r *StepReporter) CompleteStep(message string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stepName := ""
	if r.current <= len(r.stepNames) {
		stepName = r.stepNames[r.current-1]
	}

	return r.reporter.ReportStep(r.current, r.totalSteps, stepName, message)
}

// ReportPartial sends partial results for the current step.
func (r *StepReporter) ReportPartial(data any) error {
	r.mu.Lock()
	stepName := ""
	if r.current <= len(r.stepNames) {
		stepName = r.stepNames[r.current-1]
	}
	r.mu.Unlock()

	return r.reporter.ReportPartialResult(stepName, data)
}

// IsEnabled returns whether the underlying reporter is enabled.
func (r *StepReporter) IsEnabled() bool {
	return r.reporter.IsEnabled()
}

// CurrentStep returns the current step number (1-indexed).
func (r *StepReporter) CurrentStep() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.current
}

// TotalSteps returns the total number of steps.
func (r *StepReporter) TotalSteps() int {
	return r.totalSteps
}
