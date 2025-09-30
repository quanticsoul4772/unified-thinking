# 🚀 Quick Start Guide - Unified Thinking Server

## 📦 What We've Built

A complete Go-based MCP server that consolidates 5 separate thinking servers into one unified, efficient solution.

### File Structure (All Created ✅)

```
unified-thinking/
├── 📄 Documentation
│   ├── README.md           - Complete usage guide
│   ├── INSTALL.md          - Detailed installation instructions
│   ├── CHECKLIST.md        - Step-by-step installation checklist
│   ├── TECHNICAL_PLAN.md   - Architecture & design docs
│   └── Makefile            - Build automation
│
├── 🔧 Installation Scripts
│   ├── install-go.ps1      - PowerShell Go installer
│   ├── install-go.bat      - Batch Go installer
│   └── build.bat           - Server build script
│
├── 💻 Source Code
│   ├── cmd/server/
│   │   └── main.go         - Entry point
│   ├── internal/
│   │   ├── types/types.go       - Data structures
│   │   ├── storage/memory.go    - In-memory storage
│   │   ├── modes/               - Thinking modes
│   │   │   ├── linear.go        - Sequential
│   │   │   ├── tree.go          - Multi-branch
│   │   │   ├── divergent.go     - Creative
│   │   │   ├── auto.go          - Auto-detect
│   │   │   └── shared.go        - Common types
│   │   ├── validation/logic.go  - Logical validation
│   │   └── server/server.go     - MCP server
│   └── go.mod              - Go dependencies
│
└── 📋 Config
    └── .gitignore          - Git ignore rules
```

---

## ⚡ 3-Step Installation

### Step 1: Install Go (5 minutes)

**Option A - Automated (Recommended):**
```powershell
# Open PowerShell as Administrator
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
.\install-go.ps1
```

**Option B - Manual:**
1. Download from: https://go.dev/dl/
2. Run installer (use defaults)
3. **Close all terminals and open a NEW one**

**Verify:**
```bash
go version
# Should show: go version go1.23.4 windows/amd64
```

---

### Step 2: Build Server (2 minutes)

**Option A - Automated:**
```bash
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
.\build.bat
```

**Option B - Manual:**
```bash
cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
go mod download
go build -o bin\unified-thinking.exe .\cmd\server
```

**Verify:**
```bash
dir bin\unified-thinking.exe
# Should show the .exe file
```

---

### Step 3: Configure Claude (1 minute)

**Edit:** `%APPDATA%\Claude\claude_desktop_config.json`

**Add:**
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

**Then:**
- Save the file
- Restart Claude Desktop completely

---

## ✨ What You Get

### 4 Thinking Modes

| Mode | When to Use | Trigger Words |
|------|-------------|---------------|
| **Linear** | Step-by-step reasoning | "step by step", "systematically" |
| **Tree** | Multi-branch exploration | "explore branches", "alternatives" |
| **Divergent** | Creative solutions | "creative", "unconventional", "what if" |
| **Auto** | Let Claude decide | (default) |

### 9 Available Tools

1. **think** - Main thinking (auto-selects mode)
2. **history** - View past thoughts
3. **list-branches** - See all branches
4. **focus-branch** - Switch active branch
5. **branch-history** - Detailed branch info
6. **validate** - Logical consistency check
7. **prove** - Logical proof attempts
8. **check-syntax** - Statement validation
9. **search** - Find past thoughts

---

## 🧪 Test It Out

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

---

## 🔍 Verification Checklist

- [ ] `go version` works in a new terminal
- [ ] `bin\unified-thinking.exe` exists
- [ ] Claude Desktop config updated
- [ ] Claude Desktop restarted
- [ ] Server appears in Claude's available tools
- [ ] Test prompts work correctly

---

## ⚠️ Common Issues & Solutions

### "go is not recognized"
**Solution:** Close ALL terminals, open NEW one. If still failing, log out/in or restart.

### Build fails
**Solution:** Run `go mod tidy` first, then try building again.

### Server not in Claude
**Solution:** 
1. Check config file path is correct
2. Use double backslashes: `\\`
3. Completely restart Claude Desktop
4. Check for typos in JSON

### Server appears but doesn't respond
**Solution:**
1. Try running: `.\bin\unified-thinking.exe` manually
2. Check for error messages
3. Verify DEBUG=true in config

---

## 📚 Documentation

| File | Purpose |
|------|---------|
| `README.md` | Complete usage documentation |
| `INSTALL.md` | Detailed installation guide |
| `CHECKLIST.md` | Step-by-step checklist |
| `TECHNICAL_PLAN.md` | Architecture details |

---

## 🎯 Next Steps After Installation

1. **Test each mode** with different prompts
2. **Explore tree branching** with complex problems
3. **Try divergent mode** for creative thinking
4. **Use validation** on logical statements
5. **Read TECHNICAL_PLAN.md** to understand architecture

---

## 🔄 Replacing Old Servers

This replaces:
- ❌ sequential-thinking
- ❌ branch-thinking
- ❌ unreasonable-thinking-server
- ❌ mcp-logic (partially)
- ❌ state-coordinator (partially)

You can safely remove these from your config once unified-thinking is working!

---

## 📊 Performance Notes

- Uses **in-memory storage** (fast, but resets on restart)
- Written in **Go** (faster than Node.js equivalents)
- **Single binary** (no dependencies to install)
- **Auto mode detection** (smarter than manual selection)

---

## 🆘 Need Help?

1. **Check CHECKLIST.md** - Step-by-step walkthrough
2. **Read INSTALL.md** - Detailed installation info
3. **Review error messages** - Often self-explanatory
4. **Verify each step** - Use the verification checklist
5. **Test incrementally** - Make sure each phase works

---

## ✅ Success Indicators

You'll know everything is working when:
- Go version command works
- Binary builds without errors
- Server appears in Claude Desktop
- Thinking prompts get intelligent responses
- Different modes activate appropriately

---

## 🎉 You're All Set!

Once installed, just use Claude naturally. The server will:
- **Auto-detect** which thinking mode to use
- **Track thoughts** across sessions
- **Generate insights** in tree mode
- **Validate logic** when requested
- **Branch creatively** in divergent mode

Enjoy your unified thinking server! 🚀

---

*Version: 1.0.0*
*Created: 2025-09-29*
*Status: Ready for Installation*
