@echo off
REM validate-workflows.bat - Run all CI/CD checks locally before pushing (Windows version)
REM This script replicates GitHub Actions workflow checks locally to catch issues early

setlocal enabledelayedexpansion
set FAILURES=0
set CRITICAL_FAILURES=0

echo ===============================================
echo   Local Workflow Validation Suite
echo   Replicates GitHub Actions CI/CD checks
echo ===============================================
echo.

REM 1. Check Go installation
echo === Environment Checks ===
echo ^> Running: Go version check
go version
if errorlevel 1 (
    echo X FAIL: Go version check
    set /a FAILURES+=1
    set /a CRITICAL_FAILURES+=1
) else (
    echo PASS: Go version check
)
echo.

REM 2. Verify dependencies
echo === Dependency Checks ===
echo ^> Running: Download dependencies
go mod download
if errorlevel 1 (
    echo X FAIL: Download dependencies
    set /a FAILURES+=1
    set /a CRITICAL_FAILURES+=1
) else (
    echo PASS: Download dependencies
)
echo.

echo ^> Running: Verify dependencies
go mod verify
if errorlevel 1 (
    echo X FAIL: Verify dependencies
    set /a FAILURES+=1
    set /a CRITICAL_FAILURES+=1
) else (
    echo PASS: Verify dependencies
)
echo.

REM 3. Format check
echo === Code Formatting ===
echo ^> Running: Format check (gofmt)
gofmt -l . > nul 2>&1
if errorlevel 1 (
    echo WARNING: Some files need formatting
    gofmt -l .
    set /a FAILURES+=1
    set /a CRITICAL_FAILURES+=1
) else (
    REM Check if output is empty
    for /f %%i in ('gofmt -l . ^| find /c /v ""') do set COUNT=%%i
    if !COUNT! gtr 0 (
        echo X FAIL: Files need formatting:
        gofmt -l .
        set /a FAILURES+=1
        set /a CRITICAL_FAILURES+=1
    ) else (
        echo PASS: Format check
    )
)
echo.

REM 4. Go vet
echo === Static Analysis (go vet) ===
echo ^> Running: Go vet analysis
go vet ./...
if errorlevel 1 (
    echo X FAIL: Go vet analysis
    set /a FAILURES+=1
    set /a CRITICAL_FAILURES+=1
) else (
    echo PASS: Go vet analysis
)
echo.

REM 5. Lint with golangci-lint
echo === Linting (golangci-lint) ===
where golangci-lint >nul 2>&1
if errorlevel 1 (
    echo WARNING: golangci-lint not installed
    echo Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    echo Or: choco install golangci-lint
    set /a FAILURES+=1
) else (
    echo ^> Running: golangci-lint (5min timeout)
    golangci-lint run --timeout=5m
    if errorlevel 1 (
        echo X FAIL: golangci-lint
        set /a FAILURES+=1
    ) else (
        echo PASS: golangci-lint
    )
)
echo.

REM 6. Run tests (without race on Windows - requires CGO)
echo === Unit Tests ===
echo ^> Running: Tests (without race detector on Windows)
go test -v -coverprofile=coverage.out ./...
if errorlevel 1 (
    echo X FAIL: Tests
    set /a FAILURES+=1
    set /a CRITICAL_FAILURES+=1
) else (
    echo PASS: Tests (race detector runs in CI/CD on Linux)
)
echo.

REM 7. Coverage report
if exist coverage.out (
    echo === Coverage Report ===
    go tool cover -func=coverage.out | findstr "total:"
    echo.
)

REM 8. Security scan (optional)
echo === Security Scanning (optional) ===
where gosec >nul 2>&1
if errorlevel 1 (
    echo INFO: gosec not installed (optional)
    echo Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest
) else (
    echo ^> Running: Security scan (gosec)
    gosec -no-fail -quiet ./...
    if errorlevel 1 (
        echo X FAIL: Security scan
        set /a FAILURES+=1
    ) else (
        echo PASS: Security scan
    )
)
echo.

REM 9. Vulnerability check (optional)
where govulncheck >nul 2>&1
if errorlevel 1 (
    echo INFO: govulncheck not installed (optional)
    echo Install with: go install golang.org/x/vuln/cmd/govulncheck@latest
) else (
    echo ^> Running: Vulnerability check
    govulncheck ./...
    if errorlevel 1 (
        echo WARNING: Vulnerabilities found
    ) else (
        echo PASS: Vulnerability check
    )
)
echo.

REM 10. Build check
echo === Build Verification ===
echo ^> Running: Build for current platform
go build -v -o bin\unified-thinking-test.exe .\cmd\server
if errorlevel 1 (
    echo X FAIL: Build verification
    set /a FAILURES+=1
    set /a CRITICAL_FAILURES+=1
) else (
    echo PASS: Build verification
    if exist bin\unified-thinking-test.exe del bin\unified-thinking-test.exe
)
echo.

REM Summary
echo ===============================================
echo           Validation Summary
echo ===============================================
echo.

if %CRITICAL_FAILURES% equ 0 (
    echo [OK] ALL CRITICAL CHECKS PASSED
    if %FAILURES% gtr 0 (
        echo.
        echo Note: %FAILURES% optional check^(s^) did not pass ^(golangci-lint, gosec, etc.^)
        echo These are handled by CI/CD and are not blocking.
    )
    echo.
    echo Safe to push to GitHub!
    exit /b 0
) else (
    echo [ERROR] %CRITICAL_FAILURES% CRITICAL CHECK^(S^) FAILED
    if %FAILURES% gtr %CRITICAL_FAILURES% (
        echo Also: Optional checks did not pass ^(not blocking^)
    )
    echo.
    echo Fix critical issues before pushing to GitHub.
    exit /b 1
)
