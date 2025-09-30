# Unified Thinking Server

A comprehensive MCP server that consolidates multiple cognitive thinking patterns into a single, efficient Go-based server.

## Features

### Thinking Modes

- **Linear Mode**: Sequential step-by-step reasoning for systematic problem solving
- **Tree Mode**: Multi-branch parallel exploration with insights and cross-references
- **Divergent Mode**: Creative/unconventional ideation with "rebellion" capability
- **Auto Mode**: Automatic mode selection based on input content

### Capabilities

- Multi-mode thinking (linear, tree, divergent, auto)
- Branch management and exploration
- Insight generation and tracking
- Cross-reference support between branches
- Logical validation and consistency checking
- Formal proof attempts
- Syntax validation for logical statements
- Search across all thoughts
- Full history tracking

## Installation

### Prerequisites

- Go 1.23 or higher
- Git

### Build

```bash
go mod download
go build -o bin/unified-thinking.exe ./cmd/server
```

Or using make:

```bash
make build
```

## How It Works

**Important**: This is an MCP (Model Context Protocol) server. It is:
- **NOT** a standalone application you run manually
- Automatically started by Claude Desktop when the app launches
- Runs as a child process communicating via stdio
- Managed entirely by Claude Desktop's lifecycle
- Terminated when Claude Desktop closes

## Configuration

Add to your Claude Desktop config (`%APPDATA%\Claude\claude_desktop_config.json` on Windows):

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/unified-thinking/bin/unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

After saving the config:
1. Restart Claude Desktop
2. The server will start automatically
3. All tools will be available to Claude AI
4. No manual server management required

## Usage

All tools are accessed by Claude AI automatically through the MCP protocol. Responses are structured JSON data consumed directly by Claude.

### Available Tools

1. **think** - Main thinking tool
   ```json
   {
     "content": "Your thinking prompt",
     "mode": "auto|linear|tree|divergent",
     "confidence": 0.8,
     "key_points": ["point1", "point2"],
     "require_validation": true
   }
   ```

2. **history** - View thinking history
   ```json
   {
     "mode": "linear|tree|divergent",
     "branch_id": "optional"
   }
   ```

3. **list-branches** - List all branches (tree mode)

4. **focus-branch** - Switch active branch
   ```json
   {
     "branch_id": "branch-xxx"
   }
   ```

5. **branch-history** - Get detailed branch history
   ```json
   {
     "branch_id": "branch-xxx"
   }
   ```

6. **validate** - Validate a thought for logical consistency
   ```json
   {
     "thought_id": "thought-xxx"
   }
   ```

7. **prove** - Attempt to prove a logical conclusion
   ```json
   {
     "premises": ["All humans are mortal", "Socrates is human"],
     "conclusion": "Socrates is mortal"
   }
   ```

8. **check-syntax** - Validate logical statement syntax
   ```json
   {
     "statements": ["Statement 1", "Statement 2"]
   }
   ```

9. **search** - Search thoughts
   ```json
   {
     "query": "search term",
     "mode": "optional mode filter"
   }
   ```

### Example Prompts

**Auto Mode (Recommended)**:
```
"Analyze this problem using the best thinking approach"
```

**Explicit Linear Mode**:
```
"Think step by step about solving this"
```

**Explicit Tree Mode**:
```
"Explore multiple branches of this idea with cross-references"
```

**Explicit Divergent Mode**:
```
"What's a creative, unconventional solution to this?"
"Challenge all assumptions about this problem" (with force_rebellion)
```

## Architecture

```
unified-thinking/
├── cmd/server/          # Main entry point
├── internal/
│   ├── types/          # Core data structures
│   ├── storage/        # In-memory storage
│   ├── modes/          # Thinking mode implementations
│   │   ├── linear.go
│   │   ├── tree.go
│   │   ├── divergent.go
│   │   └── auto.go
│   ├── validation/     # Logic validation
│   └── server/         # MCP server implementation
└── TECHNICAL_PLAN.md   # Detailed technical documentation
```

## Development

### Build

```bash
# Build the server binary
make build

# Clean build artifacts
make clean
```

### Testing (Development Only)

**Warning**: Do not run the server manually in production. It's managed by Claude Desktop.

```bash
# For protocol debugging only (waits for MCP messages on stdin)
go run ./cmd/server/main.go

# With debug logging
DEBUG=true go run ./cmd/server/main.go
```

### Testing

```bash
# Run tests
make test

# Run with verbose output
go test -v ./...
```

## Migration from Old Servers

This server replaces:
- sequential-thinking
- branch-thinking
- unreasonable-thinking-server
- mcp-logic (partial)
- state-coordinator (partial)

### Tool Mapping

| Old Server | Old Tool | New Tool | Notes |
|------------|----------|----------|-------|
| sequential-thinking | solve-problem | think (mode: linear) | Use mode="linear" |
| branch-thinking | branch-thinking | think (mode: tree) | Use mode="tree" |
| unreasonable-thinking | generate_unreasonable_thought | think (mode: divergent, force_rebellion: true) | Use mode="divergent" |
| mcp-logic | prove | prove | Same tool |
| mcp-logic | check-well-formed | check-syntax | Similar functionality |

## Troubleshooting

### Server won't start

1. Check that Go is installed: `go version`
2. Verify the binary was built: Check `bin/` directory
3. Enable debug mode: Set `DEBUG=true` in env

### Tools not appearing

1. Restart Claude Desktop completely
2. Check config file syntax
3. Verify the executable path is correct

### Performance issues

- The server uses in-memory storage
- For long sessions, consider periodic restarts
- Monitor memory usage if processing many thoughts

## Contributing

See [TECHNICAL_PLAN.md](TECHNICAL_PLAN.md) for detailed architecture and implementation notes.

## License

MIT License

## Version

v1.0.0 - Initial unified release
