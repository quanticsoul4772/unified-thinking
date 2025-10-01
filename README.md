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

After saving the config, restart Claude Desktop.

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

10. **get-metrics** - Get system performance and usage metrics

11. **recent-branches** - Get recently accessed branches for quick context switching
    - Returns the last 10 accessed branches with timestamps
    - Shows active branch for context
    - Enables fast branch switching without remembering IDs

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
