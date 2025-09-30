# Go Installation Guide for Windows

## Quick Install (Recommended)

### Option 1: Run the Helper Script
```bash
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
.\install-go.bat
```

The script will:
1. Download Go 1.23.4 for Windows
2. Run the installer
3. Guide you through verification

---

## Option 2: Manual Installation

### Step 1: Download
1. Open your browser
2. Go to: **https://go.dev/dl/**
3. Download: **go1.23.4.windows-amd64.msi** (or latest version)

### Step 2: Install
1. Run the downloaded `.msi` file
2. Accept the license agreement
3. Use the default installation path: `C:\Program Files\Go`
4. Click "Install"
5. Click "Finish" when complete

### Step 3: Verify Installation
Open a **NEW** command prompt and run:
```bash
go version
```

You should see something like:
```
go version go1.23.4 windows/amd64
```

**Important:** You must open a NEW command prompt after installation for the PATH to update!

---

## After Go is Installed

### Build the Unified Thinking Server

1. Navigate to the project:
   ```bash
   cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

3. Build the server:
   ```bash
   go build -o bin\unified-thinking.exe .\cmd\server
   ```

   Or use make:
   ```bash
   make install-deps
   make build
   ```

4. Verify the build:
   ```bash
   dir bin\unified-thinking.exe
   ```

### Configure Claude Desktop

Add to `%APPDATA%\Claude\claude_desktop_config.json`:

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

### Test the Server

1. Restart Claude Desktop completely
2. Test with these prompts:
   - "Think step by step about how to organize a project"
   - "Explore multiple branches of solving climate change"
   - "What's a creative, unconventional solution to traffic?"

---

## Troubleshooting

### "go is not recognized" after installation
- **Solution 1:** Open a NEW command prompt (close old ones)
- **Solution 2:** Log out and log back in to Windows
- **Solution 3:** Restart your computer
- **Solution 4:** Manually add Go to PATH:
  1. Search for "Environment Variables" in Windows
  2. Edit System Environment Variables
  3. Add: `C:\Program Files\Go\bin` to PATH

### Build errors
- Make sure you're in the correct directory
- Run `go mod tidy` first
- Check that all .go files were created correctly

### Server won't start in Claude
- Verify the executable path in config is correct
- Check that the .exe file exists in the bin folder
- Try running the .exe manually first to see errors

---

## Quick Commands Reference

```bash
# Check Go installation
go version

# Navigate to project
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking

# Install dependencies
go mod download

# Build server
go build -o bin\unified-thinking.exe .\cmd\server

# Or use Makefile
make build

# Run server directly (for testing)
go run .\cmd\server\main.go

# Run with debug output
set DEBUG=true && go run .\cmd\server\main.go
```

---

## What You'll Get

Once installed and configured, you'll have access to a unified thinking server with:

- **Linear Mode**: Step-by-step reasoning
- **Tree Mode**: Multi-branch exploration with insights
- **Divergent Mode**: Creative/rebellious thinking
- **Auto Mode**: Automatic mode selection

And these tools:
- think, history, list-branches, focus-branch, branch-history
- validate, prove, check-syntax, search

---

Need help? Check:
- Official Go docs: https://go.dev/doc/install
- Project README: C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\README.md
- Technical plan: C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\TECHNICAL_PLAN.md
