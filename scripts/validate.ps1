# Pre-push validation script for Windows PowerShell
# Run this before pushing to ensure the pipeline will pass

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Running Pre-Push Validation" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check for syntax errors that would cause typecheck to fail
Write-Host "1. Checking for common syntax errors..." -ForegroundColor Yellow

$doubleUnderscore = Get-ChildItem -Path "internal" -Filter "*.go" -Recurse | Select-String -Pattern "_ = _ ="
if ($doubleUnderscore) {
    Write-Host "ERROR: Found double '_ = _ =' patterns (invalid syntax)" -ForegroundColor Red
    $doubleUnderscore | ForEach-Object { Write-Host "  $($_.Path):$($_.LineNumber): $($_.Line)" -ForegroundColor Red }
    exit 1
}
Write-Host "   ✓ No double _ = _ = patterns found" -ForegroundColor Green

$errColonEquals = Get-ChildItem -Path "internal" -Filter "*.go" -Recurse | Select-String -Pattern "_ = err :="
if ($errColonEquals) {
    Write-Host "ERROR: Found '_ = err :=' patterns (invalid syntax)" -ForegroundColor Red
    $errColonEquals | ForEach-Object { Write-Host "  $($_.Path):$($_.LineNumber): $($_.Line)" -ForegroundColor Red }
    exit 1
}
Write-Host "   ✓ No _ = err := patterns found" -ForegroundColor Green

Write-Host ""

# Run go vet
Write-Host "2. Running go vet..." -ForegroundColor Yellow
$vetOutput = go vet ./... 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: go vet failed" -ForegroundColor Red
    Write-Host $vetOutput -ForegroundColor Red
    exit 1
}
Write-Host "   ✓ go vet passed" -ForegroundColor Green
Write-Host ""

# Check if code builds
Write-Host "3. Building code..." -ForegroundColor Yellow
$buildOutput = go build ./... 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    Write-Host $buildOutput -ForegroundColor Red
    exit 1
}
Write-Host "   ✓ Build successful" -ForegroundColor Green
Write-Host ""

# Run tests
Write-Host "4. Running tests..." -ForegroundColor Yellow
$testOutput = go test -short ./... 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Tests failed" -ForegroundColor Red
    Write-Host $testOutput -ForegroundColor Red
    exit 1
}
Write-Host "   ✓ Tests passed" -ForegroundColor Green
Write-Host ""

Write-Host "========================================" -ForegroundColor Green
Write-Host "✓ All validations passed!" -ForegroundColor Green
Write-Host "Safe to push to remote" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
