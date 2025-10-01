# Quick Start Guide - Unified Thinking Server

## Overview

A complete Go-based MCP server that consolidates 5 separate thinking servers into one unified, efficient solution.

## Installation

### Step 1: Install Go (5 minutes)

#### Option A - Automated (Recommended)

Open PowerShell as Administrator and run:

```powershell
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
.\install-go.ps1
```

#### Option B - Manual

1. Download from: https://go.dev/dl/
2. Run installer (use defaults)
3. Close all terminals and open a new one

#### Verify Installation

```bash
go version
# Should show: go version go1.23.4 windows/amd64
```

### Step 2: Build Server (2 minutes)

#### Option A - Automated

```bash
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
.\build.bat
```

#### Option B - Manual

```bash
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
go mod download
go build -o bin\unified-thinking.exe .\cmd\server
```

#### Verify Build

```bash
dir bin\unified-thinking.exe
# Should show the .exe file
```

### Step 3: Configure Claude Desktop (1 minute)

Edit: `%APPDATA%\Claude\claude_desktop_config.json`

Add the following configuration:

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

Then:
- Save the file
- Restart Claude Desktop completely

## Features

### Thinking Modes

| Mode | When to Use | Trigger Words |
|------|-------------|---------------|
| Linear | Step-by-step reasoning | "step by step", "systematically" |
| Tree | Multi-branch exploration | "explore branches", "alternatives" |
| Divergent | Creative solutions | "creative", "unconventional", "what if" |
| Auto | Let Claude decide | (default) |

### Available Tools

1. **think** - Main thinking (auto-selects mode)
2. **history** - View past thoughts
3. **list-branches** - See all branches
4. **focus-branch** - Switch active branch
5. **branch-history** - Detailed branch info
6. **validate** - Logical consistency check
7. **prove** - Logical proof attempts
8. **check-syntax** - Statement validation
9. **search** - Find past thoughts

## Testing the Installation

Try these prompts after installation:

### Test 1: Linear Mode
```
Think step by step about how to organize a software project
```

### Test 2: Tree Mode
```
Explore multiple branches of solving climate change
```

### Test 3: Divergent Mode
```
What's a creative, unconventional solution to urban traffic?
```

### Test 4: Auto Mode (Default)
```
Help me analyze the best approach to learning Go programming
```

## Verification Checklist

- [ ] `go version` works in a new terminal
- [ ] `bin\unified-thinking.exe` exists
- [ ] Claude Desktop config updated
- [ ] Claude Desktop restarted
- [ ] Server appears in Claude's available tools
- [ ] Test prompts work correctly

## Troubleshooting

### "go is not recognized"
Solution: Close all terminals, open a new one. If still failing, log out/in or restart.

### Build fails
Solution: Run `go mod tidy` first, then try building again.

### Server not appearing in Claude
Solution:
1. Check config file path is correct
2. Use double backslashes: `\\`
3. Completely restart Claude Desktop
4. Check for typos in JSON

### Server appears but doesn't respond
Solution:
1. Try running: `.\bin\unified-thinking.exe` manually
2. Check for error messages
3. Verify DEBUG=true in config

## Documentation

| File | Purpose |
|------|---------|
| README.md | Complete usage documentation |
| CLAUDE.md | Developer and AI assistant guidance |
| QUICKSTART.md | This quick start guide |
| CHANGELOG.md | Version history |

## Replacing Old Servers

This server consolidates and replaces:
- sequential-thinking
- branch-thinking
- unreasonable-thinking-server
- mcp-logic (partially)
- state-coordinator (partially)

You can safely remove these from your config once unified-thinking is working.

## Performance Notes

- Uses in-memory storage (fast, but resets on restart)
- Written in Go (faster than Node.js equivalents)
- Single binary (no dependencies to install)
- Auto mode detection (intelligent mode selection)

## Success Indicators

You'll know everything is working when:
- Go version command works
- Binary builds without errors
- Server appears in Claude Desktop
- Thinking prompts get intelligent responses
- Different modes activate appropriately

## Next Steps

After successful installation:

1. Test each mode with different prompts
2. Explore tree branching with complex problems
3. Try divergent mode for creative thinking
4. Use validation on logical statements
5. Read CLAUDE.md to understand the architecture

---

Version: 1.0.0
Created: 2025-09-29
Status: Production Ready
