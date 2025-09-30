# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Unified Thinking Server is a Go-based MCP (Model Context Protocol) server that consolidates multiple cognitive thinking patterns into a single efficient server. It replaces 5 separate TypeScript servers (sequential-thinking, branch-thinking, unreasonable-thinking-server, mcp-logic, state-coordinator) with one unified implementation.

### Core Architecture

**Module Path**: `unified-thinking` (local module, not published to public registry)

**Entry Point**: `cmd/server/main.go`

**Key Components**:
- `internal/types/` - Core data structures (Thought, Branch, Insight, CrossRef, Validation)
- `internal/storage/` - In-memory storage with thread-safe operations
- `internal/modes/` - Thinking mode implementations (linear, tree, divergent, auto)
- `internal/validation/` - Logical validation and proof checking
- `internal/server/` - MCP server implementation and tool handlers

**MCP SDK**: Uses `github.com/modelcontextprotocol/go-sdk` v0.8.0

### Thinking Modes

1. **Linear Mode** (`modes/linear.go`) - Sequential step-by-step reasoning for systematic problem solving
2. **Tree Mode** (`modes/tree.go`) - Multi-branch parallel exploration with insights and cross-references
3. **Divergent Mode** (`modes/divergent.go`) - Creative/unconventional ideation with "rebellion" capability
4. **Auto Mode** (`modes/auto.go`) - Automatic mode selection based on input content analysis

The Auto Mode uses keyword detection to intelligently select the best thinking mode:
- Divergent triggers: "creative", "unconventional", "what if", "imagine", "challenge", "rebel"
- Tree triggers: branch_id provided, cross_refs present, key_points present, or keywords "branch", "explore", "alternative", "parallel"
- Linear: Default fallback for systematic reasoning

## Development Commands

### Building
```bash
# Build for Windows (default)
make build
# Output: bin/unified-thinking.exe

# Build for Linux
make linux
# Output: bin/unified-thinking

# Install dependencies
make install-deps
```

### Running

**DO NOT run the server manually!** The server is automatically started by Claude Desktop.

For development/testing only:
```bash
# Test the server (simulates stdio communication)
go run .\cmd\server\main.go

# With debug logging
set DEBUG=true && go run .\cmd\server\main.go
```

**Note**: When run manually, the server waits for MCP protocol messages on stdin. This is only useful for protocol-level debugging.

### Testing
```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Test specific package
go test -v ./internal/modes/
```

### Cleanup
```bash
make clean  # Removes bin/ directory
```

## MCP Protocol Communication

**Key Understanding**:
- All responses are JSON-formatted for Claude AI consumption
- No human-readable text formatting is used
- Claude AI processes the structured JSON data directly
- Responses contain only the data structures defined in the response types

## MCP Tool Registration

Tools are registered in `internal/server/server.go` using the pattern:
```go
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "tool-name",
    Description: "Tool description",
}, s.handleToolName)
```

Each handler returns structured JSON via `toJSONContent(responseData)`.

**Available Tools**:
1. `think` - Main thinking tool (supports all modes)
2. `history` - View thinking history (filtered by mode/branch)
3. `list-branches` - List all branches (tree mode)
4. `focus-branch` - Switch active branch
5. `branch-history` - Get detailed branch history with insights/cross-refs
6. `validate` - Validate thought for logical consistency
7. `prove` - Attempt formal proof from premises to conclusion
8. `check-syntax` - Validate logical statement syntax
9. `search` - Search thoughts by query and optional mode filter

## Storage Architecture

Uses in-memory storage (`storage/memory.go`) with sync.RWMutex for thread safety. Storage is NOT persisted to disk - all data is lost on server restart.

**Storage Operations**:
- Thread-safe with RWMutex locking
- Auto-generates IDs using counters + timestamps
- Maintains active branch state
- Simple substring-based search (production should use proper text search)

**Key Methods**:
- `StoreThought()`, `GetThought()`, `SearchThoughts()`
- `StoreBranch()`, `GetBranch()`, `ListBranches()`, `GetActiveBranch()`, `SetActiveBranch()`
- `StoreInsight()`, `StoreValidation()`, `StoreRelationship()`

## Data Flow

1. **Tool Call** → `server/server.go` handler receives request
2. **Mode Selection** → Auto mode detects or explicit mode used
3. **Processing** → Selected mode's `ProcessThought()` executes
4. **Storage** → Thought/Branch/Insight stored in memory
5. **Validation** (optional) → Logic validator checks consistency
6. **Response** → Result returned to MCP client

## Important Implementation Details

### Branch Metrics Calculation
When processing thoughts in tree mode (`modes/tree.go`):
- Branch confidence = average of all thought confidences in branch
- Branch priority = confidence + (insight_count × 0.1) + (sum of cross_ref strengths × 0.1)

### Cross-References
Cross-references link branches together with typed relationships:
- `complementary` - Ideas that work well together
- `contradictory` - Conflicting approaches
- `builds_upon` - Extends another branch's ideas
- `alternative` - Different approach to same problem

TouchPoints within cross-refs specify exact thought-to-thought connections.

### Validation
Basic logical validation checks for:
- Obvious contradictions (e.g., "always" and "never" in same statement)
- Statement completeness and syntax

Production implementations should integrate proper theorem provers or logic engines.

## Migration from Old Servers

This server replaces multiple TypeScript servers. Tool mapping:

| Old Server | Old Tool | New Tool | Usage |
|------------|----------|----------|-------|
| sequential-thinking | solve-problem | think | `mode: "linear"` |
| branch-thinking | branch-thinking | think | `mode: "tree"` |
| unreasonable-thinking | generate_unreasonable_thought | think | `mode: "divergent"`, `force_rebellion: true` |
| mcp-logic | prove | prove | Same interface |
| mcp-logic | check-well-formed | check-syntax | Similar functionality |

## How MCP Servers Work

**IMPORTANT**: MCP servers are NOT standalone executables. They are:
- Automatically started by Claude Desktop when the app launches
- Run as child processes communicating via stdio (standard input/output)
- Managed entirely by the Claude Desktop application lifecycle
- Terminated when Claude Desktop closes

The server binary (`bin/unified-thinking.exe`) should **NEVER** be run manually by users.

## Configuration

Add to Claude Desktop config (`%APPDATA%\Claude\claude_desktop_config.json` on Windows):
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\Development\\Projects\\MCP\\project-root\\mcp-servers\\unified-thinking\\bin\\unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

When Claude Desktop starts, it will:
1. Read this configuration
2. Spawn the server process using the specified command
3. Establish stdio communication channel
4. Keep the server running for the entire session
5. Terminate the server when Claude Desktop closes

## Code Style & Conventions

- Use Go standard formatting (`go fmt`)
- Package-level documentation in godoc style
- Error handling: return errors, don't panic (except in main.go for fatal errors)
- Logging: Use `log.Println()` for debug info (when DEBUG=true)
- JSON tags: Use standard Go JSON tags for request/response structs
- Thread safety: Use sync.RWMutex for shared state access

## Key Files to Understand

1. `cmd/server/main.go` - Entry point, initializes all components
2. `internal/types/types.go` - All core data structures and constants
3. `internal/server/server.go` - Tool registration and request handlers
4. `internal/modes/shared.go` - Shared types for mode implementations
5. `internal/storage/memory.go` - Storage layer with thread-safe operations

## Technical Constraints

- Go 1.23+ required
- Windows primary target (Makefile uses Windows commands by default)
- MCP SDK v0.8.0 - uses `mcp.AddTool()` and `transport.Run()` patterns
- No external databases - pure in-memory storage
- stdio transport only (no HTTP/SSE)
