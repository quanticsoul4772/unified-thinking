# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Unified Thinking Server is a Go-based MCP (Model Context Protocol) server that consolidates multiple cognitive thinking patterns into a single efficient server. It replaces 5 separate TypeScript servers (sequential-thinking, branch-thinking, unreasonable-thinking-server, mcp-logic, state-coordinator) with one unified implementation.

### Core Architecture

**Module Path**: `unified-thinking` (local module, not published to public registry)

**Entry Point**: `cmd/server/main.go`

**Key Components**:
- `internal/types/` - Core data structures (Thought, Branch, Insight, CrossRef, Validation) + Builder patterns
- `internal/storage/` - Pluggable storage layer with in-memory (default) and SQLite backends (Storage interface for testability)
- `internal/modes/` - Thinking mode implementations (linear, tree, divergent, auto) + Mode registry
- `internal/validation/` - Logical validation and proof checking
- `internal/server/` - MCP server implementation
- `internal/server/handlers/` - Focused handler modules (thinking, branches, validation, search)

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

The server supports **pluggable storage backends** via the `storage.Storage` interface with two implementations:

### In-Memory Storage (Default)
- Backend: `storage/memory.go`
- Thread-safe with sync.RWMutex
- No persistence - data lost on restart
- Fast, zero configuration
- Ideal for development and testing

### SQLite Storage (Optional)
- Backend: `storage/sqlite.go`
- Persistent across server restarts
- Write-through caching for performance
- Full-text search via FTS5
- WAL mode for concurrent reads
- Graceful fallback to memory on errors

**Configuration** via environment variables:
```bash
STORAGE_TYPE=memory     # Default - in-memory only
STORAGE_TYPE=sqlite     # Persistent SQLite storage

SQLITE_PATH=./data/thoughts.db  # Database file path
SQLITE_TIMEOUT=5000             # Connection timeout (ms)
STORAGE_FALLBACK=memory         # Fallback if SQLite fails
```

**Storage Factory** (`storage/factory.go`):
- `NewStorageFromEnv()` - Creates storage from environment variables
- `NewStorage(cfg Config)` - Creates storage from explicit configuration
- Automatic fallback handling

**Storage Operations**:
- Thread-safe with RWMutex locking
- Auto-generates IDs using counters + timestamps
- Maintains active branch state
- Full-text search (FTS5 in SQLite, substring in memory)

**Key Methods**:
- `StoreThought()`, `GetThought()`, `SearchThoughts()`
- `StoreBranch()`, `GetBranch()`, `ListBranches()`, `GetActiveBranch()`, `SetActiveBranch()`
- `StoreInsight()`, `StoreValidation()`, `StoreRelationship()`

## Data Flow

1. **Tool Call** → `server/server.go` handler receives request
2. **Mode Selection** → Auto mode detects or explicit mode used
3. **Processing** → Selected mode's `ProcessThought()` executes
4. **Storage** → Thought/Branch/Insight persisted via storage backend (memory or SQLite)
5. **Validation** (optional) → Logic validator checks consistency
6. **Response** → Result returned to MCP client

## Important Implementation Details

### Storage Architecture
- **Interface-based**: All code depends on `storage.Storage` interface for testability
- **Thread-safe**: RWMutex protection with deep copy strategy
- **Resource limits**: MaxSearchResults=1000, MaxIndexSize=100000 to prevent DoS
- **Optimized appends**: Direct append methods avoid full get-modify-store cycles

### SQLite Persistence Implementation
When using SQLite storage (`STORAGE_TYPE=sqlite`):

**Architecture**:
- **Write-through cache**: All writes go to DB first, then update in-memory cache
- **Cache-first reads**: Reads hit cache first (fast path), DB on miss (warm cache)
- **Deep copying**: All returns are deep copies to prevent data races
- **Prepared statements**: Pre-compiled SQL statements for performance

**Schema Design** (`sqlite_schema.go`):
- Core tables: `thoughts`, `branches`, `insights`, `cross_refs`, `validations`, `relationships`
- FTS5 virtual table: `thoughts_fts` for full-text search on thought content
- Indexes: Optimized for common query patterns (mode filtering, branch lookups, timestamps)
- JSON columns: `key_points`, `metadata` stored as JSON for complex types

**Performance Optimizations**:
- WAL mode: Concurrent reads while writing
- 64MB cache size: Reduces disk I/O
- Batch inserts: Transaction batching for multi-record writes
- Connection pooling: Reuses DB connections
- FTS5 tokenization: Fast full-text search with relevance ranking

**Data Migration**:
- Schema versioning: `schema_version` table tracks migrations
- Automatic upgrades: Applies migrations on startup if needed
- Backward compatible: Old data remains accessible after upgrades

### Builder Patterns
Use builders from `internal/types/builders.go` for object construction:
```go
thought := types.NewThought().
    Content("Example").
    Mode(types.ModeLinear).
    Confidence(0.9).
    Build()
```

### Mode Registry
Modes implement `ThinkingMode` interface and can be registered dynamically:
- `Name()` - Returns mode identifier
- `CanHandle()` - Determines if mode can process input
- `ProcessThought()` - Executes mode logic

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

### Default Configuration (In-Memory)
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

### With SQLite Persistence
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\Development\\Projects\\MCP\\project-root\\mcp-servers\\unified-thinking\\bin\\unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "C:\\Users\\YourName\\AppData\\Roaming\\Claude\\unified-thinking.db",
        "STORAGE_FALLBACK": "memory"
      }
    }
  }
}
```

**Environment Variables**:
- `STORAGE_TYPE`: `memory` (default) or `sqlite`
- `SQLITE_PATH`: Path to SQLite database file (created if not exists)
- `SQLITE_TIMEOUT`: Connection timeout in milliseconds (default: 5000)
- `STORAGE_FALLBACK`: Fallback storage type if primary fails (default: none)
- `DEBUG`: Enable debug logging (`true` or `false`)

When Claude Desktop starts, it will:
1. Read this configuration
2. Spawn the server process using the specified command
3. Establish stdio communication channel
4. Initialize storage backend based on environment variables
5. Keep the server running for the entire session
6. Terminate the server when Claude Desktop closes

## Code Style & Conventions

- Use Go standard formatting (`go fmt`)
- Package-level documentation in godoc style
- Error handling: return errors, don't panic (except in main.go for fatal errors)
- Logging: Use `log.Println()` for debug info (when DEBUG=true)
- JSON tags: Use standard Go JSON tags for request/response structs
- Thread safety: Use sync.RWMutex for shared state access

## Key Files to Understand

1. `cmd/server/main.go` - Entry point, initializes storage and server components
2. `internal/types/types.go` - All core data structures and constants
3. `internal/server/server.go` - Tool registration and request handlers
4. `internal/modes/shared.go` - Shared types for mode implementations
5. `internal/storage/factory.go` - Storage factory and configuration
6. `internal/storage/memory.go` - In-memory storage implementation
7. `internal/storage/sqlite.go` - SQLite storage with write-through cache
8. `internal/storage/sqlite_schema.go` - Database schema and migrations

## Technical Constraints

- Go 1.23+ required
- Windows primary target (Makefile uses Windows commands by default)
- MCP SDK v0.8.0 - uses `mcp.AddTool()` and `transport.Run()` patterns
- Storage: In-memory (default) or SQLite (optional)
- SQLite backend uses modernc.org/sqlite (pure Go, no CGO)
- stdio transport only (no HTTP/SSE)
