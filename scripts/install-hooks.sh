#!/bin/bash
#
# Install git hooks for unified-thinking development
#

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

echo "Installing git hooks..."

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Copy pre-commit hook
cp "$SCRIPT_DIR/pre-commit" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"

echo "Installed: pre-commit hook"
echo ""
echo "Git hooks installed successfully!"
echo ""
echo "The pre-commit hook will run:"
echo "  - go fmt (formatting check)"
echo "  - go vet (static analysis)"
echo "  - go test -short (quick tests)"
echo "  - golangci-lint --fast (if available)"
echo ""
echo "To skip hooks temporarily: git commit --no-verify"
