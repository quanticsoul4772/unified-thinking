// Package streaming provides MCP progress notification support for long-running tools.
//
// This package enables real-time progress updates during tool execution using the
// standard MCP notifications/progress mechanism. It's designed to be:
//
//   - Backward Compatible: Clients that don't provide a progressToken simply don't
//     receive notifications; the tool executes normally.
//
//   - Non-Intrusive: Handlers can call progress methods without checking if streaming
//     is enabled; the DefaultReporter handles disabled cases as no-ops.
//
//   - Rate Limited: Built-in debouncing prevents notification floods.
//
//   - Configurable: Per-tool configuration controls behavior like partial data sending.
//
// # Basic Usage
//
// In a handler, create a reporter and report progress:
//
//	func (h *Handler) Handle(ctx context.Context, req *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, *Output, error) {
//	    // Create a reporter (will be no-op if client doesn't want streaming)
//	    reporter := streaming.CreateReporter(req, "my-tool")
//
//	    // Report step-based progress
//	    reporter.ReportStep(1, 3, "analyze", "Analyzing input...")
//
//	    // Do work...
//
//	    reporter.ReportStep(2, 3, "process", "Processing data...")
//
//	    // Do more work...
//
//	    reporter.ReportStep(3, 3, "complete", "Done!")
//
//	    return nil, &Output{...}, nil
//	}
//
// # Using StepReporter
//
// For step-based workflows, StepReporter provides convenient tracking:
//
//	steps := []string{"analyze", "process", "synthesize", "validate"}
//	reporter := streaming.CreateReporter(req, "my-tool")
//	sr := streaming.NewStepReporter(reporter, steps)
//
//	sr.StartStep("Starting analysis...")
//	// work...
//	sr.StartStep("Processing data...")
//	// work...
//
// # Context Integration
//
// The reporter can be stored in context for nested function calls:
//
//	ctx, reporter := streaming.InjectReporter(ctx, req, "my-tool")
//
//	// Later, in a nested function:
//	r := streaming.GetReporter(ctx)
//	r.ReportProgress(50, 100, "Halfway done")
//
// # Streaming-Enabled Tools
//
// The following tools support streaming (P0 = essential, P1 = important, P2 = enhancement):
//
// P0 Tools (enabled by default):
//   - execute-workflow: Variable steps, sends step results
//   - run-preset: 5-6 steps, sends step results
//   - got-generate: k iterations, sends new vertices
//
// P1 Tools:
//   - got-aggregate: Source count steps, sends merged result
//   - think (tree mode): Branch count steps, sends branch results
//   - perform-cbr-cycle: 4 steps (retrieve, reuse, revise, retain)
//
// P2 Tools:
//   - synthesize-insights: Input count steps
//   - analyze-perspectives: Stakeholder count steps
//   - build-causal-graph: Variable + link count steps
//   - evaluate-hypotheses: Hypothesis count steps
package streaming

// Version is the streaming package version.
const Version = "1.0.0"

// StreamingEnabledTools lists all tools that support streaming progress notifications.
var StreamingEnabledTools = []string{
	// P0 - Essential
	"execute-workflow",
	"run-preset",
	"got-generate",

	// P1 - Important
	"got-aggregate",
	"think",
	"perform-cbr-cycle",

	// P2 - Enhancement
	"synthesize-insights",
	"analyze-perspectives",
	"build-causal-graph",
	"evaluate-hypotheses",
}
