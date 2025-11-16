# Windows Setup Guide for Local Validation

This guide helps you run code quality checks on Windows before pushing to GitHub.

## Quick Start (PowerShell)

You can run all checks using Go commands directly - **no make required**:

### 1. Format Code
```powershell
# Check formatting
gofmt -l .

# Auto-fix formatting (recommended)
gofmt -w .
```

### 2. Static Analysis
```powershell
go vet ./...
```

### 3. Run Tests
```powershell
# Basic tests (works on Windows)
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Note: Race detector requires CGO on Windows (not typically available)
# Skip -race flag on Windows unless you have CGO enabled
```

### 4. Lint Code (after setup below)
```powershell
# Add Go bin to PATH first (see Setup section)
golangci-lint run --timeout=5m
```

### 5. Build
```powershell
go build -o bin\unified-thinking.exe .\cmd\server
```

## One-Time Setup

### Install golangci-lint

Already installed at: `C:\Users\rbsmi\go\bin\golangci-lint.exe`

**Option A: Add to PowerShell PATH (Permanent)**
```powershell
# Add Go bin to your PATH
$env:Path += ";C:\Users\rbsmi\go\bin"

# To make permanent, add to PowerShell profile:
Add-Content $PROFILE '$env:Path += ";C:\Users\rbsmi\go\bin"'
```

**Option B: Use full path**
```powershell
C:\Users\rbsmi\go\bin\golangci-lint.exe run --timeout=5m
```

**Option C: Create alias in PowerShell**
```powershell
# Add to PowerShell profile:
Add-Content $PROFILE 'Set-Alias golangci-lint "C:\Users\rbsmi\go\bin\golangci-lint.exe"'
```

### Install Make (Optional)

If you want to use `make` commands:

**Option 1: Chocolatey**
```powershell
choco install make
```

**Option 2: Use Git Bash**
- Git for Windows includes `make`
- Open Git Bash and run: `make pre-commit`

## Pre-Commit Checklist (PowerShell)

Before committing, run these commands:

```powershell
# 1. Format code
gofmt -w .

# 2. Static analysis
go vet ./...

# 3. Quick tests
go test -short ./...

# 4. Lint (if PATH is set)
golangci-lint run --timeout=5m

# 5. Verify build
go build -o bin\unified-thinking.exe .\cmd\server
```

## Pre-Push Validation (PowerShell)

Before pushing, run full validation:

```powershell
# Run the validation script
.\scripts\validate-workflows.bat
```

**Expected Results**:
- ✅ Go version check
- ✅ Dependencies verified
- ✅ Format check (after running gofmt -w .)
- ✅ Go vet
- ⚠️  golangci-lint (needs PATH setup)
- ✅ Tests (without -race on Windows)
- ✅ Build

## Using Makefile (Alternative)

If you install `make` or use Git Bash:

```bash
# Quick pre-commit checks
make pre-commit

# Full CI validation
make validate-ci

# Individual checks
make lint
make vet
make test
make build
```

## Common Issues

### "make not recognized"
- Use PowerShell commands above, OR
- Install make via Chocolatey, OR
- Use Git Bash instead

### "golangci-lint not recognized"
- Add `C:\Users\rbsmi\go\bin` to your PATH
- Or use full path: `C:\Users\rbsmi\go\bin\golangci-lint.exe`

### "race requires cgo"
- This is normal on Windows
- Use `go test ./...` instead of `go test -race ./...`
- CI/CD on Linux will run race detector

### Formatting issues
- Run `gofmt -w .` to auto-fix all formatting
- This is safe and recommended

## PowerShell Script (Copy & Paste)

Save this as `validate.ps1` for easy validation:

```powershell
# validate.ps1 - Quick validation for Windows
Write-Host "Running pre-commit validation..." -ForegroundColor Blue

Write-Host "`n=== Formatting ===" -ForegroundColor Yellow
$files = gofmt -l .
if ($files) {
    Write-Host "Files need formatting:" -ForegroundColor Red
    $files
    Write-Host "Run: gofmt -w ." -ForegroundColor Cyan
    exit 1
} else {
    Write-Host "✓ All files formatted" -ForegroundColor Green
}

Write-Host "`n=== Go Vet ===" -ForegroundColor Yellow
go vet ./...
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Go vet passed" -ForegroundColor Green
} else {
    Write-Host "✗ Go vet failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n=== Tests ===" -ForegroundColor Yellow
go test -short ./...
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Tests passed" -ForegroundColor Green
} else {
    Write-Host "✗ Tests failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n=== Build ===" -ForegroundColor Yellow
go build -o bin\unified-thinking-test.exe .\cmd\server
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Build successful" -ForegroundColor Green
    Remove-Item bin\unified-thinking-test.exe -ErrorAction SilentlyContinue
} else {
    Write-Host "✗ Build failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n✓ All checks passed! Safe to commit." -ForegroundColor Green
```

**Usage**:
```powershell
.\validate.ps1
```

## Summary

**Before Commit** (Quick - 10 seconds):
```powershell
gofmt -w . && go vet ./... && go test -short ./...
```

**Before Push** (Full - 1-2 minutes):
```powershell
.\scripts\validate-workflows.bat
```

**One-Time Setup**:
1. ✅ golangci-lint already installed
2. Add `C:\Users\rbsmi\go\bin` to PATH (optional but recommended)

That's it! You can now validate code locally on Windows. 🎉
