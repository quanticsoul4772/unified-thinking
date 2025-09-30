.PHONY: build run test clean install-deps windows linux

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
	go test -v ./...

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
	@echo "  build       - Build the server (Windows .exe)"
	@echo "  linux       - Build for Linux"
	@echo "  run         - Run the server directly"
	@echo "  debug       - Run with debug logging"
	@echo "  test        - Run all tests"
	@echo "  clean       - Remove build artifacts"
	@echo "  install-deps - Download Go dependencies"
	@echo "  verify      - Verify installation"
	@echo "  help        - Show this help message"
