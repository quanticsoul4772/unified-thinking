<#
.SYNOPSIS
    Build script for unified-thinking MCP server (Windows PowerShell alternative to Makefile)

.DESCRIPTION
    Provides all Makefile targets as PowerShell commands for Windows developers.
    Usage: .\build.ps1 <command>

.EXAMPLE
    .\build.ps1 build          # Build for Windows
    .\build.ps1 test           # Run tests
    .\build.ps1 test-coverage  # Run tests with coverage
    .\build.ps1 lint           # Run golangci-lint
    .\build.ps1 pre-commit     # Run pre-commit checks
    .\build.ps1 help           # Show all commands
#>

param(
    [Parameter(Position=0)]
    [ValidateSet(
        "build", "windows", "linux", "macos",
        "run", "debug",
        "test", "test-verbose", "test-coverage", "test-race", "test-short", "test-all",
        "test-storage", "test-storage-coverage",
        "benchmark", "benchmark-reasoning", "benchmark-e2e", "benchmark-all",
        "lint", "lint-fix", "fmt", "fmt-check", "vet",
        "pre-commit", "validate-ci",
        "clean", "install-deps", "verify",
        "help"
    )]
    [string]$Command = "help"
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Step { param($msg) Write-Host "`n>> $msg" -ForegroundColor Cyan }
function Write-Success { param($msg) Write-Host $msg -ForegroundColor Green }
function Write-Warn { param($msg) Write-Host $msg -ForegroundColor Yellow }
function Write-Err { param($msg) Write-Host $msg -ForegroundColor Red }

# Ensure we're in the project root
$ProjectRoot = $PSScriptRoot
Set-Location $ProjectRoot

switch ($Command) {
    # ============ BUILD COMMANDS ============
    "build" {
        Write-Step "Building for Windows..."
        if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
        go build -o bin\unified-thinking.exe .\cmd\server
        if ($LASTEXITCODE -eq 0) { Write-Success "Build complete: bin\unified-thinking.exe" }
        else { Write-Err "Build failed"; exit 1 }
    }

    "windows" {
        & $PSCommandPath build
    }

    "linux" {
        Write-Step "Cross-compiling for Linux..."
        if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
        $env:GOOS = "linux"
        $env:GOARCH = "amd64"
        go build -o bin/unified-thinking ./cmd/server
        Remove-Item Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
        if ($LASTEXITCODE -eq 0) { Write-Success "Build complete: bin/unified-thinking (Linux)" }
        else { Write-Err "Build failed"; exit 1 }
    }

    "macos" {
        Write-Step "Cross-compiling for macOS..."
        if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
        $env:GOOS = "darwin"
        $env:GOARCH = "amd64"
        go build -o bin/unified-thinking-macos ./cmd/server
        Remove-Item Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
        if ($LASTEXITCODE -eq 0) { Write-Success "Build complete: bin/unified-thinking-macos" }
        else { Write-Err "Build failed"; exit 1 }
    }

    # ============ RUN COMMANDS ============
    "run" {
        Write-Step "Running unified thinking server..."
        go run ./cmd/server/main.go
    }

    "debug" {
        Write-Step "Running in debug mode..."
        $env:DEBUG = "true"
        go run ./cmd/server/main.go
    }

    # ============ TEST COMMANDS ============
    "test" {
        Write-Step "Running tests..."
        go test ./...
        if ($LASTEXITCODE -ne 0) { exit 1 }
    }

    "test-verbose" {
        Write-Step "Running tests with verbose output..."
        go test -v ./...
        if ($LASTEXITCODE -ne 0) { exit 1 }
    }

    "test-coverage" {
        Write-Step "Running tests with coverage..."
        go test -coverprofile=coverage.out ./...
        if ($LASTEXITCODE -ne 0) { exit 1 }

        Write-Step "Coverage summary:"
        go tool cover -func=coverage.out

        Write-Step "Generating HTML coverage report..."
        go tool cover -html=coverage.out -o coverage.html
        Write-Success "Coverage report saved to coverage.html"
    }

    "test-race" {
        Write-Step "Running tests with race detector..."
        go test -race ./...
        if ($LASTEXITCODE -ne 0) { exit 1 }
    }

    "test-short" {
        Write-Step "Running short tests..."
        go test -short ./...
        if ($LASTEXITCODE -ne 0) { exit 1 }
    }

    "test-all" {
        Write-Step "Running comprehensive test suite..."
        & $PSCommandPath test-race
        & $PSCommandPath test-coverage
        Write-Success "All tests complete!"
    }

    "test-storage" {
        Write-Step "Running storage layer tests..."
        go test -v ./internal/storage/
        if ($LASTEXITCODE -ne 0) { exit 1 }
    }

    "test-storage-coverage" {
        Write-Step "Running storage tests with coverage..."
        go test -coverprofile=storage-coverage.out ./internal/storage/
        go tool cover -func=storage-coverage.out
        go tool cover -html=storage-coverage.out -o storage-coverage.html
        Write-Success "Storage coverage report saved to storage-coverage.html"
    }

    # ============ BENCHMARK COMMANDS ============
    "benchmark" {
        Write-Step "Running performance benchmarks..."
        go test -bench=. -benchmem ./internal/storage/
    }

    "benchmark-reasoning" {
        Write-Step "Running reasoning quality benchmarks..."
        go test -v ./benchmarks/
    }

    "benchmark-e2e" {
        Write-Step "Building server for E2E testing..."
        go build -o unified-thinking-server.exe ./cmd/server
        Write-Step "Running E2E benchmarks via MCP..."
        go test -v ./benchmarks/ -run TestMCP -timeout 10m
    }

    "benchmark-all" {
        & $PSCommandPath benchmark
        & $PSCommandPath benchmark-e2e
        Write-Success "All benchmarks complete!"
    }

    # ============ CODE QUALITY COMMANDS ============
    "lint" {
        Write-Step "Running golangci-lint..."
        $linterPath = $null

        # First check PATH
        $linter = Get-Command golangci-lint -ErrorAction SilentlyContinue
        if ($linter) {
            $linterPath = $linter.Source
        } else {
            # Check GOPATH/bin
            $gopath = (go env GOPATH) -replace "`n", "" -replace "`r", ""
            $gopathLinter = Join-Path $gopath "bin\golangci-lint.exe"
            if (Test-Path $gopathLinter) {
                $linterPath = $gopathLinter
            }
        }

        if (-not $linterPath) {
            Write-Warn "golangci-lint not found. Install with:"
            Write-Host "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
            exit 1
        }

        & $linterPath run --timeout=5m
        if ($LASTEXITCODE -ne 0) { exit 1 }
    }

    "lint-fix" {
        Write-Step "Running golangci-lint with auto-fix..."
        $linterPath = $null

        $linter = Get-Command golangci-lint -ErrorAction SilentlyContinue
        if ($linter) {
            $linterPath = $linter.Source
        } else {
            $gopath = (go env GOPATH) -replace "`n", "" -replace "`r", ""
            $gopathLinter = Join-Path $gopath "bin\golangci-lint.exe"
            if (Test-Path $gopathLinter) {
                $linterPath = $gopathLinter
            }
        }

        if (-not $linterPath) {
            Write-Warn "golangci-lint not found. Install with:"
            Write-Host "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
            exit 1
        }

        & $linterPath run --fix --timeout=5m
    }

    "fmt" {
        Write-Step "Formatting code..."
        go fmt ./...
        Write-Success "Code formatted"
    }

    "fmt-check" {
        Write-Step "Checking code formatting..."
        $unformatted = gofmt -l .
        if ($unformatted) {
            Write-Err "Files need formatting:"
            $unformatted | ForEach-Object { Write-Host "  $_" }
            exit 1
        }
        Write-Success "All files formatted correctly"
    }

    "vet" {
        Write-Step "Running go vet..."
        go vet ./...
        if ($LASTEXITCODE -ne 0) { exit 1 }
    }

    # ============ WORKFLOW COMMANDS ============
    "pre-commit" {
        Write-Step "Running pre-commit checks..."
        & $PSCommandPath fmt-check
        & $PSCommandPath vet
        & $PSCommandPath test-short
        Write-Success "Pre-commit checks passed!"
    }

    "validate-ci" {
        Write-Step "Running comprehensive CI validation locally..."
        if (Test-Path "scripts\validate-workflows.bat") {
            & .\scripts\validate-workflows.bat
        } else {
            Write-Warn "validate-workflows.bat not found, running manual checks..."
            & $PSCommandPath fmt-check
            & $PSCommandPath vet
            & $PSCommandPath lint
            & $PSCommandPath test
        }
    }

    # ============ MAINTENANCE COMMANDS ============
    "clean" {
        Write-Step "Cleaning..."
        if (Test-Path "bin") { Remove-Item -Recurse -Force "bin" }
        if (Test-Path "coverage.out") { Remove-Item "coverage.out" }
        if (Test-Path "coverage.html") { Remove-Item "coverage.html" }
        if (Test-Path "storage-coverage.out") { Remove-Item "storage-coverage.out" }
        if (Test-Path "storage-coverage.html") { Remove-Item "storage-coverage.html" }
        if (Test-Path "unified-thinking-server.exe") { Remove-Item "unified-thinking-server.exe" }
        Write-Success "Clean complete"
    }

    "install-deps" {
        Write-Step "Installing dependencies..."
        go mod tidy
        go mod download
        Write-Success "Dependencies installed"
    }

    "verify" {
        Write-Step "Verifying installation..."
        go version
        if (Test-Path "bin\unified-thinking.exe") {
            Write-Success "Binary found: bin\unified-thinking.exe"
        } else {
            Write-Warn "Binary NOT found - run '.\build.ps1 build'"
        }
    }

    # ============ HELP ============
    "help" {
        Write-Host @"

unified-thinking Build Script (PowerShell)
==========================================

Usage: .\build.ps1 <command>

BUILD COMMANDS:
  build              Build for Windows (bin\unified-thinking.exe)
  linux              Cross-compile for Linux
  macos              Cross-compile for macOS
  run                Run the server directly
  debug              Run with debug logging (DEBUG=true)

TEST COMMANDS:
  test               Run all tests
  test-verbose       Run tests with verbose output
  test-coverage      Run tests with coverage report
  test-race          Run tests with race detector
  test-short         Run quick tests (skip slow ones)
  test-storage       Run storage layer tests only
  test-storage-coverage  Storage tests with coverage
  test-all           Run comprehensive test suite

BENCHMARK COMMANDS:
  benchmark          Run performance benchmarks
  benchmark-reasoning Run reasoning quality benchmarks
  benchmark-e2e      Run E2E benchmarks via MCP
  benchmark-all      Run all benchmarks

CODE QUALITY:
  lint               Run golangci-lint
  lint-fix           Run golangci-lint with auto-fix
  fmt                Format code with go fmt
  fmt-check          Check code formatting
  vet                Run go vet static analysis
  pre-commit         Quick checks before commit (fmt+vet+test-short)
  validate-ci        Full CI validation locally

MAINTENANCE:
  clean              Remove build artifacts
  install-deps       Download Go dependencies
  verify             Verify installation
  help               Show this help message

EXAMPLES:
  .\build.ps1 build          # Build the binary
  .\build.ps1 test-coverage  # Run tests with coverage
  .\build.ps1 pre-commit     # Run before committing

"@ -ForegroundColor White
    }
}
