# Streaming Response Support

The unified-thinking MCP server supports real-time progress notifications for long-running tools using the MCP `notifications/progress` mechanism.

## Overview

Streaming enables clients to receive progress updates during tool execution, providing:
- Real-time feedback on multi-step operations
- Step-by-step progress tracking
- Improved user experience for complex reasoning tasks

**Backward Compatible**: Clients that don't provide a `progressToken` simply don't receive notifications; tools execute normally.

## Architecture

### Package Structure

```
internal/streaming/
├── streaming.go       # Package docs, constants, enabled tools list
├── types.go           # ProgressUpdate, StreamingConfig, tool configs
├── reporter.go        # ProgressReporter interface and implementations
├── context.go         # Context helpers for reporter injection
├── notifications.go   # MCP integration via ServerSession.NotifyProgress
└── streaming_test.go  # Comprehensive tests
```

### Core Components

#### ProgressReporter Interface

```go
type ProgressReporter interface {
    ReportProgress(current, total float64, message string) error
    ReportStep(step int, totalSteps int, stepName string, message string) error
    ReportPartialResult(stepName string, data any) error
    IsEnabled() bool
    GetProgressToken() any
}
```

#### Implementations

| Reporter | Purpose |
|----------|---------|
| `DefaultReporter` | No-op reporter when streaming disabled |
| `RateLimitedReporter` | Wraps reporter with debouncing |
| `NotifyingReporter` | Sends MCP progress notifications |
| `StepReporter` | Convenience wrapper for step-based workflows |

## Streaming-Enabled Tools

### P0 - Essential (High Value)

| Tool | Steps | Description |
|------|-------|-------------|
| `execute-workflow` | Variable | Reports each workflow step execution |
| `run-preset` | 5-6 | Reports preset step progress (dry-run or execute) |
| `got-generate` | k iterations | Reports each vertex generation |

### P1 - Important

| Tool | Steps | Description |
|------|-------|-------------|
| `got-aggregate` | Source count | Reports collect/merge/complete phases |
| `think` | 2-3 | Reports mode selection, processing, validation |
| `perform-cbr-cycle` | 4 | Reports retrieve/reuse/revise/retain phases |

### P2 - Enhancement

| Tool | Steps | Description |
|------|-------|-------------|
| `synthesize-insights` | Input count + 2 | Reports analysis start and completion |
| `analyze-perspectives` | Stakeholder count + 1 | Reports perspective analysis |
| `build-causal-graph` | Observation count + 2 | Reports parsing and graph building |
| `evaluate-hypotheses` | Hypothesis count + 2 | Reports evaluation progress |

## Usage

### Handler Pattern

```go
func (h *Handler) HandleMyTool(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input MyToolRequest,
) (*mcp.CallToolResult, *MyToolResponse, error) {
    // Create reporter (no-op if client doesn't want streaming)
    reporter := streaming.CreateReporter(req, "my-tool")

    totalSteps := calculateSteps(input)

    // Report step progress
    if reporter.IsEnabled() {
        _ = reporter.ReportStep(1, totalSteps, "init", "Initializing...")
    }

    // Do work...
    result := doWork()

    // Report completion
    if reporter.IsEnabled() {
        _ = reporter.ReportStep(totalSteps, totalSteps, "complete", "Done!")
    }

    return nil, &MyToolResponse{Result: result}, nil
}
```

### Using StepReporter

For workflows with predefined steps:

```go
steps := []string{"analyze", "process", "synthesize", "validate"}
reporter := streaming.CreateReporter(req, "my-tool")
sr := streaming.NewStepReporter(reporter, steps)

sr.StartStep("Starting analysis...")
// work...
sr.StartStep("Processing data...")
// work...
```

### Context Integration

For nested function calls:

```go
// In handler
reporter := streaming.CreateReporter(req, "my-tool")
ctx = streaming.WithReporter(ctx, reporter)

// In nested function
func processNested(ctx context.Context) {
    r := streaming.GetReporter(ctx)
    r.ReportProgress(50, 100, "Halfway done")
}
```

## Configuration

### Per-Tool Configuration

Each streaming-enabled tool has a configuration:

```go
type StreamingConfig struct {
    Enabled         bool          // Whether streaming is enabled
    MinInterval     time.Duration // Minimum time between notifications
    SendPartialData bool          // Whether to send intermediate results
    AutoProgress    bool          // Auto-calculate progress percentages
}
```

Default configurations are defined in `internal/streaming/types.go`:

```go
var ToolConfigs = map[string]StreamingConfig{
    "execute-workflow": {
        Enabled:         true,
        MinInterval:     100 * time.Millisecond,
        SendPartialData: true,
        AutoProgress:    true,
    },
    // ... more tools
}
```

### Rate Limiting

Progress notifications are rate-limited to prevent flooding:

- Default minimum interval: 100ms
- Step changes always go through (bypass rate limit)
- Partial data can be filtered via `SendPartialData` config

## MCP Protocol Integration

### Client Requirements

Clients must include a `progressToken` in tool call parameters to receive notifications:

```json
{
  "method": "tools/call",
  "params": {
    "name": "execute-workflow",
    "arguments": { ... },
    "_meta": {
      "progressToken": "unique-token-123"
    }
  }
}
```

### Notification Format

Progress notifications use the MCP `notifications/progress` method:

```json
{
  "method": "notifications/progress",
  "params": {
    "progressToken": "unique-token-123",
    "progress": 2,
    "total": 5,
    "message": "Executing step 2 of 5: process"
  }
}
```

### go-sdk Integration

The implementation uses:
- `CallToolParamsRaw.GetProgressToken()` to retrieve the client's token
- `ServerSession.NotifyProgress(ctx, *ProgressNotificationParams)` to send notifications

## Testing

Run streaming tests:

```bash
go test ./internal/streaming/... -v
```

The test suite includes:
- Unit tests for all reporter types
- Rate limiting behavior tests
- Concurrency safety tests
- Context integration tests
- Benchmarks for performance validation

## Adding Streaming to New Tools

1. **Add to StreamingEnabledTools** in `internal/streaming/streaming.go`:
   ```go
   var StreamingEnabledTools = []string{
       // ... existing tools
       "my-new-tool",
   }
   ```

2. **Add configuration** in `internal/streaming/types.go`:
   ```go
   var ToolConfigs = map[string]StreamingConfig{
       // ... existing configs
       "my-new-tool": {
           Enabled:         true,
           MinInterval:     100 * time.Millisecond,
           SendPartialData: false,
           AutoProgress:    true,
       },
   }
   ```

3. **Update handler** to use streaming:
   ```go
   reporter := streaming.CreateReporter(req, "my-new-tool")
   if reporter.IsEnabled() {
       _ = reporter.ReportStep(step, total, stepName, message)
   }
   ```

4. **Update tests** to verify streaming behavior (optional but recommended)

## Performance Considerations

- Rate limiting prevents notification floods
- No-op reporter has zero overhead when streaming disabled
- Context injection is lightweight (single value)
- Notifications are non-blocking (fire-and-forget)
