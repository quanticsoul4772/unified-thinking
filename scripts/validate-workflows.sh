#!/bin/bash
# validate-workflows.sh - Run all CI/CD checks locally before pushing
# This script replicates GitHub Actions workflow checks locally to catch issues early

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track failures
FAILURES=0

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Local Workflow Validation Suite          ║${NC}"
echo -e "${BLUE}║  Replicates GitHub Actions CI/CD checks   ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# Function to run a check and track failures
run_check() {
    local name="$1"
    local command="$2"

    echo -e "${YELLOW}► Running: ${name}${NC}"
    if eval "$command"; then
        echo -e "${GREEN}✓ PASS: ${name}${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}✗ FAIL: ${name}${NC}"
        echo ""
        ((FAILURES++))
        return 1
    fi
}

# 1. Check Go installation
echo -e "${BLUE}═══ Environment Checks ═══${NC}"
run_check "Go version check" "go version"

# 2. Verify dependencies
echo -e "${BLUE}═══ Dependency Checks ═══${NC}"
run_check "Download dependencies" "go mod download"
run_check "Verify dependencies" "go mod verify"

# 3. Format check
echo -e "${BLUE}═══ Code Formatting ═══${NC}"
run_check "Format check (gofmt)" "test -z \"\$(gofmt -l .)\""

# 4. Go vet
echo -e "${BLUE}═══ Static Analysis (go vet) ═══${NC}"
run_check "Go vet analysis" "go vet ./..."

# 5. Lint with golangci-lint (matches ci.yml)
echo -e "${BLUE}═══ Linting (golangci-lint) ═══${NC}"
if command -v golangci-lint &> /dev/null; then
    run_check "golangci-lint (5min timeout)" "golangci-lint run --timeout=5m"
else
    echo -e "${YELLOW}⚠ WARNING: golangci-lint not installed${NC}"
    echo -e "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    echo -e "Or: brew install golangci-lint (macOS)"
    echo -e "Or: choco install golangci-lint (Windows)"
    ((FAILURES++))
    echo ""
fi

# 6. Run tests (matches ci.yml test job)
echo -e "${BLUE}═══ Unit Tests ═══${NC}"
run_check "Tests with race detector" "go test -v -race -coverprofile=coverage.out -covermode=atomic ./..."

# 7. Generate coverage report
if [ -f coverage.out ]; then
    echo -e "${BLUE}═══ Coverage Report ═══${NC}"
    go tool cover -func=coverage.out | tail -n 1
    echo ""
fi

# 8. Security scan (optional - matches security.yml)
echo -e "${BLUE}═══ Security Scanning (optional) ═══${NC}"
if command -v gosec &> /dev/null; then
    run_check "Security scan (gosec)" "gosec -no-fail -quiet ./..."
else
    echo -e "${YELLOW}⚠ INFO: gosec not installed (optional)${NC}"
    echo -e "Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"
    echo ""
fi

# 9. Vulnerability check (optional - matches security.yml)
if command -v govulncheck &> /dev/null; then
    run_check "Vulnerability check" "govulncheck ./..."
else
    echo -e "${YELLOW}⚠ INFO: govulncheck not installed (optional)${NC}"
    echo -e "Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
    echo ""
fi

# 10. Build check (matches ci.yml build job)
echo -e "${BLUE}═══ Build Verification ═══${NC}"
run_check "Build for current platform" "go build -v -o bin/unified-thinking-test ./cmd/server"
if [ -f bin/unified-thinking-test ]; then
    rm bin/unified-thinking-test
fi

# 11. Validate workflow YAML syntax
echo -e "${BLUE}═══ Workflow YAML Validation ═══${NC}"
if command -v yamllint &> /dev/null; then
    run_check "YAML syntax check" "yamllint .github/workflows/*.yml"
else
    echo -e "${YELLOW}⚠ INFO: yamllint not installed (optional)${NC}"
    echo -e "Install with: pip install yamllint"
    echo ""
fi

# Summary
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Validation Summary               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"

if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}✓ ALL CHECKS PASSED${NC}"
    echo -e "Safe to push to GitHub!"
    echo ""
    exit 0
else
    echo -e "${RED}✗ ${FAILURES} CHECK(S) FAILED${NC}"
    echo -e "Fix issues before pushing to GitHub."
    echo ""
    exit 1
fi
