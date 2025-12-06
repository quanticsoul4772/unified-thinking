<#
.SYNOPSIS
    Install git hooks for unified-thinking development (Windows)

.DESCRIPTION
    Copies pre-commit hook to .git/hooks directory

.EXAMPLE
    .\scripts\install-hooks.ps1
#>

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$HooksDir = Join-Path $ProjectRoot ".git\hooks"

Write-Host "Installing git hooks..." -ForegroundColor Cyan

# Create hooks directory if it doesn't exist
if (-not (Test-Path $HooksDir)) {
    New-Item -ItemType Directory -Path $HooksDir | Out-Null
}

# Copy pre-commit hook
Copy-Item "$ScriptDir\pre-commit" "$HooksDir\pre-commit" -Force

Write-Host "Installed: pre-commit hook" -ForegroundColor Green
Write-Host ""
Write-Host "Git hooks installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "The pre-commit hook will run:"
Write-Host "  - go fmt (formatting check)"
Write-Host "  - go vet (static analysis)"
Write-Host "  - go test -short (quick tests)"
Write-Host "  - golangci-lint --fast (if available)"
Write-Host ""
Write-Host "To skip hooks temporarily: git commit --no-verify"
