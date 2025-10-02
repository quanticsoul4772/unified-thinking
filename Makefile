.PHONY: build run test test-verbose test-coverage test-race test-short test-storage test-storage-coverage benchmark test-all pre-commit clean install-deps windows linux help

# Default target
all: build

# Build for Windows
build: windows

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
	go run .\cmd\server\main.go

# Run with debug output
debug:
	@echo "Running in debug mode..."
	set DEBUG=true && go run .\cmd\server\main.go

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

# Pre-commit checks
pre-commit: test-short
	@echo "Pre-commit checks passed!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@if exist bin rmdir /s /q bin
	@echo "Clean complete"

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
	@echo "Dependencies installed"

# Verify installation
verify:
	@echo "Verifying installation..."
	go version
	@if exist bin\unified-thinking.exe (echo "Binary found: bin\unified-thinking.exe") else (echo "Binary NOT found - run 'make build'")
	@echo "Verification complete"

# Show help
help:
	@echo "Available targets:"
	@echo "  build                  - Build the server (Windows .exe)"
	@echo "  linux                  - Build for Linux"
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
	@echo "  pre-commit             - Run pre-commit checks"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean                  - Remove build artifacts"
	@echo "  install-deps           - Download Go dependencies"
	@echo "  verify                 - Verify installation"
	@echo "  help                   - Show this help message"
