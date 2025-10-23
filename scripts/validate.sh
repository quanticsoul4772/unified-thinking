#!/bin/bash
# Pre-push validation script
# Run this before pushing to ensure the pipeline will pass

set -e  # Exit on first error

echo "========================================"
echo "Running Pre-Push Validation"
echo "========================================"
echo ""

# Check for syntax errors that would cause typecheck to fail
echo "1. Checking for common syntax errors..."
if grep -rn "_ = _ =" internal/ --include="*.go" 2>/dev/null; then
    echo "ERROR: Found double '_ = _ =' patterns (invalid syntax)"
    exit 1
fi
echo "   ✓ No double _ = _ = patterns found"

if grep -rn "_ = err :=" internal/ --include="*.go" 2>/dev/null; then
    echo "ERROR: Found '_ = err :=' patterns (invalid syntax)"
    exit 1
fi
echo "   ✓ No _ = err := patterns found"

if grep -rn "_ = err =" internal/ --include="*.go" | grep -v "if _ = err" 2>/dev/null; then
    echo "ERROR: Found '_ = err =' patterns (likely invalid syntax)"
    exit 1
fi
echo "   ✓ No _ = err = patterns found"

echo ""

# Run go vet
echo "2. Running go vet..."
if ! go vet ./...; then
    echo "ERROR: go vet failed"
    exit 1
fi
echo "   ✓ go vet passed"
echo ""

# Check if code builds
echo "3. Building code..."
if ! go build ./...; then
    echo "ERROR: Build failed"
    exit 1
fi
echo "   ✓ Build successful"
echo ""

# Run tests
echo "4. Running tests..."
if ! go test -short ./...; then
    echo "ERROR: Tests failed"
    exit 1
fi
echo "   ✓ Tests passed"
echo ""

echo "========================================"
echo "✓ All validations passed!"
echo "Safe to push to remote"
echo "========================================"
