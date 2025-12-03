package streaming

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === Types Tests ===

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 100*time.Millisecond, cfg.MinInterval)
	assert.False(t, cfg.SendPartialData)
	assert.True(t, cfg.AutoProgress)
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig(
		WithMinInterval(50*time.Millisecond),
		WithPartialData(true),
		WithAutoProgress(false),
	)

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 50*time.Millisecond, cfg.MinInterval)
	assert.True(t, cfg.SendPartialData)
	assert.False(t, cfg.AutoProgress)
}

func TestGetToolConfig(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		wantEnabled bool
	}{
		{"execute-workflow is P0", "execute-workflow", true},
		{"run-preset is P0", "run-preset", true},
		{"got-generate is P0", "got-generate", true},
		{"got-aggregate is P1", "got-aggregate", true},
		{"think is P1", "think", true},
		{"perform-cbr-cycle is P1", "perform-cbr-cycle", true},
		{"synthesize-insights is P2", "synthesize-insights", true},
		{"analyze-perspectives is P2", "analyze-perspectives", true},
		{"build-causal-graph is P2", "build-causal-graph", true},
		{"evaluate-hypotheses is P2", "evaluate-hypotheses", true},
		{"unknown tool is disabled", "unknown-tool", false},
		{"history is not streaming", "history", false},
		{"validate is not streaming", "validate", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetToolConfig(tt.toolName)
			assert.Equal(t, tt.wantEnabled, cfg.Enabled, "tool %s", tt.toolName)
		})
	}
}

func TestIsStreamingEnabled(t *testing.T) {
	assert.True(t, IsStreamingEnabled("execute-workflow"))
	assert.True(t, IsStreamingEnabled("got-generate"))
	assert.False(t, IsStreamingEnabled("unknown"))
	assert.False(t, IsStreamingEnabled(""))
}

// === DefaultReporter Tests ===

func TestDefaultReporter(t *testing.T) {
	r := NewDefaultReporter("test-token")

	// All methods should return nil (no-op)
	assert.NoError(t, r.ReportProgress(50, 100, "test"))
	assert.NoError(t, r.ReportStep(1, 5, "step1", "message"))
	assert.NoError(t, r.ReportPartialResult("step1", map[string]any{"data": "test"}))

	// IsEnabled should return false
	assert.False(t, r.IsEnabled())

	// Token should be preserved
	assert.Equal(t, "test-token", r.GetProgressToken())
}

func TestDefaultReporterNilToken(t *testing.T) {
	r := NewDefaultReporter(nil)

	assert.False(t, r.IsEnabled())
	assert.Nil(t, r.GetProgressToken())
}

// === RateLimitedReporter Tests ===

func TestRateLimitedReporter(t *testing.T) {
	// Create a mock reporter that tracks calls
	mock := &mockReporter{}

	config := StreamingConfig{
		Enabled:         true,
		MinInterval:     50 * time.Millisecond,
		SendPartialData: true,
		AutoProgress:    true,
	}

	r := NewRateLimitedReporter(mock, config)

	// First call should go through
	err := r.ReportProgress(1, 10, "first")
	require.NoError(t, err)
	assert.Equal(t, 1, mock.progressCalls)

	// Immediate second call should be rate limited
	err = r.ReportProgress(2, 10, "second")
	require.NoError(t, err)
	assert.Equal(t, 1, mock.progressCalls, "should be rate limited")

	// Wait for rate limit to expire
	time.Sleep(60 * time.Millisecond)

	// Now it should go through
	err = r.ReportProgress(3, 10, "third")
	require.NoError(t, err)
	assert.Equal(t, 2, mock.progressCalls)
}

func TestRateLimitedReporterStepChanges(t *testing.T) {
	mock := &mockReporter{}
	config := StreamingConfig{
		Enabled:         true,
		MinInterval:     1 * time.Second, // Long interval
		SendPartialData: true,
		AutoProgress:    true,
	}

	r := NewRateLimitedReporter(mock, config)

	// Step changes should always go through
	err := r.ReportStep(1, 5, "step1", "first")
	require.NoError(t, err)
	assert.Equal(t, 1, mock.stepCalls)

	// Different step should go through even within rate limit
	err = r.ReportStep(2, 5, "step2", "second")
	require.NoError(t, err)
	assert.Equal(t, 2, mock.stepCalls)

	// Same step should be rate limited
	err = r.ReportStep(2, 5, "step2", "same step")
	require.NoError(t, err)
	assert.Equal(t, 2, mock.stepCalls, "same step should be rate limited")
}

func TestRateLimitedReporterPartialData(t *testing.T) {
	mock := &mockReporter{}

	// Config with partial data disabled
	configDisabled := StreamingConfig{
		Enabled:         true,
		MinInterval:     0,
		SendPartialData: false,
		AutoProgress:    true,
	}

	r := NewRateLimitedReporter(mock, configDisabled)
	err := r.ReportPartialResult("step1", "data")
	require.NoError(t, err)
	assert.Equal(t, 0, mock.partialCalls, "partial data should be filtered")

	// Config with partial data enabled
	configEnabled := StreamingConfig{
		Enabled:         true,
		MinInterval:     0,
		SendPartialData: true,
		AutoProgress:    true,
	}

	r2 := NewRateLimitedReporter(mock, configEnabled)
	err = r2.ReportPartialResult("step1", "data")
	require.NoError(t, err)
	assert.Equal(t, 1, mock.partialCalls, "partial data should be sent")
}

// === StepReporter Tests ===

func TestStepReporter(t *testing.T) {
	mock := &mockReporter{}
	steps := []string{"analyze", "process", "validate", "complete"}

	sr := NewStepReporter(mock, steps)

	assert.Equal(t, 4, sr.TotalSteps())
	assert.Equal(t, 0, sr.CurrentStep())

	// Start first step
	err := sr.StartStep("Analyzing data")
	require.NoError(t, err)
	assert.Equal(t, 1, sr.CurrentStep())
	assert.Equal(t, 1, mock.stepCalls)

	// Complete current step (no advance)
	err = sr.CompleteStep("Analysis complete")
	require.NoError(t, err)
	assert.Equal(t, 1, sr.CurrentStep())
	assert.Equal(t, 2, mock.stepCalls)

	// Start second step
	err = sr.StartStep("Processing")
	require.NoError(t, err)
	assert.Equal(t, 2, sr.CurrentStep())
	assert.Equal(t, 3, mock.stepCalls)

	// Report partial
	err = sr.ReportPartial(map[string]any{"progress": 50})
	require.NoError(t, err)
	assert.Equal(t, 1, mock.partialCalls)
}

func TestStepReporterConcurrency(t *testing.T) {
	mock := &mockReporter{}
	steps := []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "s10"}

	sr := NewStepReporter(mock, steps)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sr.StartStep("concurrent step")
		}()
	}
	wg.Wait()

	// Should have 10 step calls (concurrent but safe)
	assert.Equal(t, 10, mock.stepCalls)
}

// === Context Tests ===

func TestContextReporter(t *testing.T) {
	ctx := context.Background()

	// Without reporter, should get default
	r := GetReporter(ctx)
	assert.False(t, r.IsEnabled())

	// Add reporter to context
	mock := &mockReporter{enabled: true}
	ctx = WithReporter(ctx, mock)

	// Now should get the mock
	r = GetReporter(ctx)
	assert.True(t, r.IsEnabled())
}

func TestContextConfig(t *testing.T) {
	ctx := context.Background()

	// Without config, should get default
	cfg := GetConfig(ctx)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 100*time.Millisecond, cfg.MinInterval)

	// Add custom config
	custom := StreamingConfig{
		Enabled:     false,
		MinInterval: 500 * time.Millisecond,
	}
	ctx = WithConfig(ctx, custom)

	// Now should get custom
	cfg = GetConfig(ctx)
	assert.False(t, cfg.Enabled)
	assert.Equal(t, 500*time.Millisecond, cfg.MinInterval)
}

func TestContextToolName(t *testing.T) {
	ctx := context.Background()

	// Without tool name, should get empty
	name := GetToolName(ctx)
	assert.Empty(t, name)

	// Add tool name
	ctx = WithToolName(ctx, "execute-workflow")

	// Now should get it
	name = GetToolName(ctx)
	assert.Equal(t, "execute-workflow", name)
}

func TestStreamingContext(t *testing.T) {
	mock := &mockReporter{enabled: true}
	config := StreamingConfig{Enabled: true}

	sc := NewStreamingContext(mock, config, "test-tool")

	assert.True(t, sc.IsEnabled())

	// Test convenience methods
	err := sc.ReportProgress(50, 100, "test")
	require.NoError(t, err)
	assert.Equal(t, 1, mock.progressCalls)

	err = sc.ReportStep(1, 3, "s1", "msg")
	require.NoError(t, err)
	assert.Equal(t, 1, mock.stepCalls)

	// Test ToContext/FromContext
	ctx := sc.ToContext(context.Background())
	sc2 := FromContext(ctx)

	assert.Equal(t, "test-tool", sc2.ToolName)
	assert.True(t, sc2.Config.Enabled)
}

func TestStreamingContextDisabled(t *testing.T) {
	// With nil reporter
	sc := NewStreamingContext(nil, StreamingConfig{Enabled: true}, "test")
	assert.False(t, sc.IsEnabled())

	// Reporter methods should not panic
	err := sc.ReportProgress(50, 100, "test")
	assert.NoError(t, err)

	err = sc.ReportStep(1, 3, "s1", "msg")
	assert.NoError(t, err)

	err = sc.ReportPartialResult("s1", "data")
	assert.NoError(t, err)
}

// === Integration Tests ===

func TestProgressUpdateFields(t *testing.T) {
	update := ProgressUpdate{
		Step:        2,
		TotalSteps:  5,
		StepName:    "processing",
		Percentage:  40.0,
		Message:     "Processing data...",
		PartialData: map[string]any{"processed": 100},
		Timestamp:   time.Now(),
	}

	assert.Equal(t, 2, update.Step)
	assert.Equal(t, 5, update.TotalSteps)
	assert.Equal(t, "processing", update.StepName)
	assert.Equal(t, 40.0, update.Percentage)
	assert.Equal(t, "Processing data...", update.Message)
	assert.NotNil(t, update.PartialData)
}

func TestToolConfigsComplete(t *testing.T) {
	// Verify all streaming-enabled tools have configs
	for _, tool := range StreamingEnabledTools {
		cfg := GetToolConfig(tool)
		assert.True(t, cfg.Enabled, "tool %s should be enabled", tool)
		assert.True(t, cfg.MinInterval > 0, "tool %s should have MinInterval", tool)
	}
}

// === Mock Reporter ===

type mockReporter struct {
	enabled       bool
	progressCalls int
	stepCalls     int
	partialCalls  int
	mu            sync.Mutex
}

func (m *mockReporter) ReportProgress(current, total float64, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progressCalls++
	return nil
}

func (m *mockReporter) ReportStep(step int, totalSteps int, stepName string, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stepCalls++
	return nil
}

func (m *mockReporter) ReportPartialResult(stepName string, data any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.partialCalls++
	return nil
}

func (m *mockReporter) IsEnabled() bool {
	return m.enabled
}

func (m *mockReporter) GetProgressToken() any {
	return "mock-token"
}

// === Benchmark Tests ===

func BenchmarkDefaultReporterReportProgress(b *testing.B) {
	r := NewDefaultReporter("token")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ReportProgress(float64(i), 1000, "test")
	}
}

func BenchmarkRateLimitedReporterReportProgress(b *testing.B) {
	mock := &mockReporter{}
	config := StreamingConfig{
		Enabled:     true,
		MinInterval: time.Microsecond,
	}
	r := NewRateLimitedReporter(mock, config)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ReportProgress(float64(i), 1000, "test")
	}
}

func BenchmarkStepReporterStartStep(b *testing.B) {
	mock := &mockReporter{}
	steps := make([]string, 1000)
	for i := range steps {
		steps[i] = "step"
	}
	sr := NewStepReporter(mock, steps)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sr.StartStep("message")
	}
}

func BenchmarkGetReporterFromContext(b *testing.B) {
	mock := &mockReporter{}
	ctx := WithReporter(context.Background(), mock)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetReporter(ctx)
	}
}
