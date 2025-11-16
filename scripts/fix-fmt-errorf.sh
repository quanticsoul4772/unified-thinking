#!/bin/bash
# fix-fmt-errorf.sh - Fix non-constant format string warnings in fmt.Errorf calls
# Replaces: fmt.Errorf("message: " + err.Error())
# With:     fmt.Errorf("message: %w", err)

set -e

echo "Fixing non-constant format strings in fmt.Errorf calls..."

# Pattern: fmt.Errorf("text: " + err.Error())
# Replace with: fmt.Errorf("text: %w", err)

files=(
    "internal/server/handlers/abductive.go"
    "internal/server/handlers/backtracking.go"
    "internal/server/handlers/case_based.go"
    "internal/server/handlers/dual_process.go"
    "internal/server/handlers/symbolic.go"
    "internal/server/handlers/unknown_unknowns.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "Fixing $file..."
        # Use sed to replace the pattern
        # This handles: fmt.Errorf("message: " + err.Error())
        sed -i 's/fmt\.Errorf("\([^"]*\): " + err\.Error())/fmt.Errorf("\1: %w", err)/g' "$file"
    else
        echo "Warning: $file not found"
    fi
done

echo "Done! Run 'go vet ./...' to verify fixes."
