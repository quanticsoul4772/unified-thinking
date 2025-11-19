.PHONY: build run test test-verbose test-coverage test-race test-short test-storage test-storage-coverage benchmark test-all pre-commit clean install-deps windows linux macos darwin lint lint-fix fmt-check vet validate-ci help

# Detect OS
UNAME_S := $(shell uname -s 2>/dev/null || echo Windows)

# Default target - builds for current OS
all: build

# Build for current OS
ifeq ($(UNAME_S),Darwin)
build: macos
else ifeq ($(UNAME_S),Linux)
build: linux
else
build: windows
endif

# Build for macOS
macos darwin:
	@echo "Building for macOS..."
	@mkdir -p bin
	go build -o bin/unified-thinking ./cmd/server

# Build for Windows
windows:
	@echo "Building for Windows..."
	@if not exist bin mkdir bin
	go build -o bin\unified-thinking.exe .\cmd\server

# Build for Linux
linux:
	@echo "Building for Linux..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/unified-thinking ./cmd/server

# Run the server directly
run:
	@echo "Running unified thinking server..."
	go run ./cmd/server/main.go

# Run with debug output
ifeq ($(UNAME_S),Windows)
debug:
	@echo "Running in debug mode..."
	set DEBUG=true && go run ./cmd/server/main.go
else
debug:
	@echo "Running in debug mode..."
	DEBUG=true go run ./cmd/server/main.go
endif

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

# Run short tests only (skip long-running tests)
test-short:
	@echo "Running short tests..."
	go test -short ./...

# Run storage tests only
test-storage:
	@echo "Running storage layer tests..."
	go test -v ./internal/storage/

# Run storage tests with coverage
test-storage-coverage:
	@echo "Running storage tests with coverage..."
	go test -coverprofile=storage-coverage.out ./internal/storage/
	go tool cover -func=storage-coverage.out
	go tool cover -html=storage-coverage.out -o storage-coverage.html
	@echo "Storage coverage report saved to storage-coverage.html"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./internal/storage/

# Run all test suites (comprehensive)
test-all: test-race test-coverage
	@echo "All tests complete!"

# Code quality checks
lint:
	@echo "Running golangci-lint..."
	golangci-lint run --timeout=5m

lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	golangci-lint run --fix --timeout=5m

# Check code formatting (cross-platform)
ifeq ($(UNAME_S),Windows)
fmt-check:
	@echo "Checking code formatting..."
	@for /f %%i in ('gofmt -l . ^| find /c /v ""') do @set COUNT=%%i
	@if %COUNT% gtr 0 (echo Files need formatting: && gofmt -l . && exit 1) else (echo All files formatted correctly)
else
fmt-check:
	@echo "Checking code formatting..."
	@test -z "$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)
	@echo "All files formatted correctly"
endif

vet:
	@echo "Running go vet..."
	go vet ./...

# Comprehensive pre-commit validation (matches CI)
validate-ci:
	@echo "Running comprehensive CI validation locally..."
	@.\scripts\validate-workflows.bat

# Pre-commit checks (quick validation before commit)
pre-commit: fmt-check vet test-short
	@echo "Pre-commit checks passed!"

# Clean build artifacts
ifeq ($(UNAME_S),Windows)
clean:
	@echo "Cleaning..."
	@if exist bin rmdir /s /q bin
	@echo "Clean complete"
else
clean:
	@echo "Cleaning..."
	@rm -rf bin
	@echo "Clean complete"
endif

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
	@echo "Dependencies installed"

# Verify installation
ifeq ($(UNAME_S),Windows)
verify:
	@echo "Verifying installation..."
	go version
	@if exist bin\unified-thinking.exe (echo "Binary found: bin\unified-thinking.exe") else (echo "Binary NOT found - run 'make build'")
	@echo "Verification complete"
else
verify:
	@echo "Verifying installation..."
	go version
	@test -f bin/unified-thinking && echo "Binary found: bin/unified-thinking" || echo "Binary NOT found - run 'make build'"
	@echo "Verification complete"
endif

# Show help
help:
	@echo "Available targets:"
	@echo "  build                  - Build for current OS (auto-detected)"
	@echo "  macos/darwin           - Build for macOS"
	@echo "  linux                  - Build for Linux"
	@echo "  windows                - Build for Windows"
	@echo "  run                    - Run the server directly"
	@echo "  debug                  - Run with debug logging"
	@echo ""
	@echo "Testing:"
	@echo "  test                   - Run all tests"
	@echo "  test-verbose           - Run tests with verbose output"
	@echo "  test-coverage          - Run tests with coverage report"
	@echo "  test-race              - Run tests with race detector"
	@echo "  test-short             - Run quick tests (skip slow ones)"
	@echo "  test-storage           - Run storage layer tests only"
	@echo "  test-storage-coverage  - Storage tests with coverage"
	@echo "  benchmark              - Run performance benchmarks"
	@echo "  test-all               - Run all test suites (comprehensive)"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint                   - Run golangci-lint"
	@echo "  lint-fix               - Run golangci-lint with auto-fix"
	@echo "  fmt-check              - Check code formatting"
	@echo "  vet                    - Run go vet static analysis"
	@echo "  pre-commit             - Quick checks before commit (fmt+vet+test-short)"
	@echo "  validate-ci            - Full CI validation locally (all checks)"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean                  - Remove build artifacts"
	@echo "  install-deps           - Download Go dependencies"
	@echo "  verify                 - Verify installation"
	@echo "  help                   - Show this help message"
