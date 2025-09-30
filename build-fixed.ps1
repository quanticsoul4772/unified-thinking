# Build script for Unified Thinking Server (PowerShell version)
# Handles fetching the latest MCP SDK

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Building Unified Thinking Server" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
$goExists = Get-Command go -ErrorAction SilentlyContinue
if (-not $goExists) {
    Write-Host "ERROR: Go is not installed or not in PATH!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please:" -ForegroundColor Yellow
    Write-Host "1. Install Go from https://go.dev/dl/" -ForegroundColor White
    Write-Host "2. Close this window and open a NEW PowerShell" -ForegroundColor White
    Write-Host "3. Run this script again" -ForegroundColor White
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "Go is installed:" -ForegroundColor Green
go version
Write-Host ""

# Navigate to project directory
Set-Location "C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking"
Write-Host "Working directory: $(Get-Location)" -ForegroundColor White
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Step 1: Fetching MCP SDK..." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Create a clean go.mod
Write-Host "Initializing go.mod..." -ForegroundColor White
@"
module unified-thinking

go 1.23
"@ | Out-File -FilePath "go.mod" -Encoding UTF8 -NoNewline

# Try to fetch the SDK
Write-Host "Fetching latest MCP SDK from GitHub..." -ForegroundColor White
try {
    $output = go get github.com/modelcontextprotocol/go-sdk/mcp@latest 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Latest tag not found, trying main branch..." -ForegroundColor Yellow
        $output = go get github.com/modelcontextprotocol/go-sdk/mcp@main 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to fetch SDK"
        }
    }
    Write-Host "SDK fetched successfully!" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Failed to fetch MCP SDK" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""
Write-Host "Running go mod tidy..." -ForegroundColor White
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: go mod tidy failed" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "Dependencies configured successfully!" -ForegroundColor Green
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Step 2: Building server..." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build the server
go build -o bin\unified-thinking.exe .\cmd\server
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please check for errors above." -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Build Complete!" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Verify the executable was created
if (Test-Path "bin\unified-thinking.exe") {
    Write-Host "Executable created: bin\unified-thinking.exe" -ForegroundColor Green
    $fileSize = (Get-Item "bin\unified-thinking.exe").Length
    Write-Host "File size: $fileSize bytes" -ForegroundColor White
    Write-Host ""
    
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "Next Steps:" -ForegroundColor Cyan
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "1. Update your Claude Desktop config:" -ForegroundColor White
    Write-Host "   File: $env:APPDATA\Claude\claude_desktop_config.json" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "2. Add this configuration:" -ForegroundColor White
    Write-Host @"
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
"@ -ForegroundColor Yellow
    Write-Host ""
    Write-Host "3. Restart Claude Desktop completely" -ForegroundColor White
    Write-Host ""
    Write-Host "4. Test with prompts like:" -ForegroundColor White
    Write-Host "   - 'Think step by step about...'" -ForegroundColor Cyan
    Write-Host "   - 'Explore multiple branches of...'" -ForegroundColor Cyan
    Write-Host "   - 'What's a creative solution to...'" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host "ERROR: Executable not found after build!" -ForegroundColor Red
    Write-Host "Something went wrong during the build process." -ForegroundColor Red
    Write-Host ""
}

Read-Host "Press Enter to exit"
