# Installation Checklist

## ✅ Phase 1: Install Go

### Method A: PowerShell Script (Recommended)
- [ ] Open PowerShell as Administrator
- [ ] Navigate to: `C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking`
- [ ] Run: `.\install-go.ps1`
- [ ] Follow the installation wizard
- [ ] **IMPORTANT:** Close PowerShell and open a NEW window

### Method B: Manual Download
- [ ] Visit: https://go.dev/dl/
- [ ] Download: `go1.23.4.windows-amd64.msi`
- [ ] Run the installer (use default settings)
- [ ] **IMPORTANT:** Close all command prompts and open a NEW one

### Verify Go Installation
- [ ] Open a **NEW** PowerShell or Command Prompt
- [ ] Run: `go version`
- [ ] Expected output: `go version go1.23.4 windows/amd64`
- [ ] If "go is not recognized", log out and log back in or restart computer

---

## ✅ Phase 2: Build Unified Thinking Server

### Option A: Use Build Script
- [ ] Open Command Prompt
- [ ] Navigate to: `C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking`
- [ ] Run: `build.bat`
- [ ] Wait for build to complete
- [ ] Verify: `bin\unified-thinking.exe` exists

### Option B: Manual Build
- [ ] Open Command Prompt
- [ ] Navigate to: `C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking`
- [ ] Run: `go mod download`
- [ ] Run: `go build -o bin\unified-thinking.exe .\cmd\server`
- [ ] Verify: Check that `bin\unified-thinking.exe` was created

---

## ✅ Phase 3: Configure Claude Desktop

- [ ] Locate Claude config file: `%APPDATA%\Claude\claude_desktop_config.json`
- [ ] Open it in a text editor (Notepad, VS Code, etc.)
- [ ] Add the unified-thinking server configuration:

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

**Notes:**
- If you already have other servers, add this to the existing `"mcpServers"` object
- Make sure the path uses double backslashes: `\\`
- Verify the path matches where your executable actually is

---

## ✅ Phase 4: Test the Server

- [ ] **Restart Claude Desktop** completely (close and reopen)
- [ ] Try these test prompts:

### Test 1: Linear Mode
```
Think step by step about how to organize a large project
```
Expected: Sequential reasoning

### Test 2: Tree Mode
```
Explore multiple branches of solving climate change
```
Expected: Multiple branches with insights

### Test 3: Divergent Mode
```
What's a creative, unconventional solution to traffic congestion?
```
Expected: Creative/rebellious thinking

### Test 4: Auto Mode
```
Analyze the best approach to learn a new programming language
```
Expected: Automatic mode selection

---

## 🔧 Troubleshooting

### Go not found after installation
- [ ] Close ALL command prompts/PowerShell windows
- [ ] Open a NEW window
- [ ] If still not found, log out and log back in
- [ ] If still not found, restart computer
- [ ] Check PATH manually:
  - Search "Environment Variables" in Windows
  - Verify `C:\Program Files\Go\bin` is in PATH

### Build fails
- [ ] Make sure Go is installed: `go version`
- [ ] Check you're in the right directory
- [ ] Try: `go mod tidy` then rebuild
- [ ] Check for error messages and Google them

### Server doesn't appear in Claude
- [ ] Verify `bin\unified-thinking.exe` exists
- [ ] Check config file syntax (valid JSON)
- [ ] Verify paths use double backslashes: `\\`
- [ ] Restart Claude Desktop completely
- [ ] Check Claude Desktop logs for errors

### Server appears but doesn't work
- [ ] Try running manually: `.\bin\unified-thinking.exe`
- [ ] Check for error messages
- [ ] Enable DEBUG mode in config
- [ ] Check logs in Claude Desktop

---

## 📝 Quick Reference Commands

```bash
# Check Go installation
go version

# Navigate to project
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking

# Download dependencies
go mod download

# Build server
go build -o bin\unified-thinking.exe .\cmd\server

# Or use scripts
.\install-go.ps1  # Install Go
.\build.bat       # Build server

# Test run server directly
.\bin\unified-thinking.exe
```

---

## 📚 Documentation Files

- `README.md` - Complete usage documentation
- `INSTALL.md` - Detailed installation instructions
- `TECHNICAL_PLAN.md` - Technical architecture details
- `install-go.ps1` - PowerShell installation script
- `install-go.bat` - Batch installation script
- `build.bat` - Build automation script

---

## ✨ What You'll Get

Once everything is working, you'll have:

### 4 Thinking Modes
- **Linear**: Step-by-step sequential reasoning
- **Tree**: Multi-branch parallel exploration
- **Divergent**: Creative/unconventional thinking
- **Auto**: Automatic mode detection

### 9 Available Tools
1. `think` - Main thinking tool
2. `history` - View past thoughts
3. `list-branches` - See all branches
4. `focus-branch` - Switch active branch
5. `branch-history` - Detailed branch info
6. `validate` - Check logical consistency
7. `prove` - Attempt logical proofs
8. `check-syntax` - Validate statements
9. `search` - Search all thoughts

---

## 🎯 Success Criteria

You'll know everything is working when:
- [ ] `go version` shows Go 1.23.4 (or higher)
- [ ] `bin\unified-thinking.exe` exists and runs
- [ ] Claude Desktop shows the unified-thinking server
- [ ] You can use thinking prompts and get responses
- [ ] The server uses different modes appropriately

---

## 📞 Need Help?

If you get stuck:
1. Check this checklist thoroughly
2. Read the error messages carefully
3. Check `INSTALL.md` for detailed instructions
4. Verify each step completed successfully
5. Try the troubleshooting section above

---

Last Updated: $(date)
Version: 1.0.0
