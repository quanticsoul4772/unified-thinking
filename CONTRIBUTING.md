# Contributing to Unified Thinking Server

First off, thank you for considering contributing to the Unified Thinking Server! It's people like you that make this project such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* Use a clear and descriptive title
* Describe the exact steps which reproduce the problem
* Provide specific examples to demonstrate the steps
* Describe the behavior you observed after following the steps
* Explain which behavior you expected to see instead and why
* Include details about your configuration and environment

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* Use a clear and descriptive title
* Provide a step-by-step description of the suggested enhancement
* Provide specific examples to demonstrate the steps
* Describe the current behavior and explain which behavior you expected to see instead
* Explain why this enhancement would be useful

### Pull Requests

Please follow these steps to have your contribution considered:

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code follows the existing code style
6. Issue that pull request!

## Development Process

### Setting Up Your Development Environment

1. **Prerequisites**
   - Go 1.24 or higher
   - Git
   - Make (optional but recommended)

2. **Clone the repository**
   ```bash
   git clone https://github.com/quanticsoul4772/unified-thinking.git
   cd unified-thinking
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Build the project**
   ```bash
   make build
   # or
   go build -o bin/unified-thinking ./cmd/server
   ```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/modes/

# Run with race detector
go test -race ./...
```

### Code Style

We use standard Go formatting and linting:

* Run `gofmt -s -w .` before committing
* Run `golangci-lint run` to check for common issues
* Follow Go best practices and idioms
* Write clear, self-documenting code with comments where necessary

### Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line
* Consider starting the commit message with an applicable emoji:
  - 🎨 `:art:` when improving the format/structure of the code
  - ⚡ `:zap:` when improving performance
  - 🐛 `:bug:` when fixing a bug
  - ✨ `:sparkles:` when introducing new features
  - 📝 `:memo:` when writing docs
  - 🔧 `:wrench:` when updating configuration files
  - ✅ `:white_check_mark:` when adding tests
  - 🔒 `:lock:` when dealing with security

### Testing

* Write unit tests for all new code
* Maintain or improve code coverage (currently at 73.9%)
* Test edge cases and error conditions
* Use table-driven tests where appropriate
* Mock external dependencies appropriately

## Project Structure

```
unified-thinking/
├── cmd/server/          # Main entry point
├── internal/
│   ├── types/           # Core data structures
│   ├── storage/         # Storage implementations
│   ├── modes/           # Thinking mode implementations
│   ├── reasoning/       # Reasoning engines
│   ├── analysis/        # Analysis tools
│   ├── metacognition/   # Self-evaluation
│   ├── validation/      # Validation logic
│   ├── integration/     # Cross-mode integration
│   ├── orchestration/   # Workflow automation
│   ├── memory/          # Episodic memory
│   └── server/          # MCP server implementation
│       └── handlers/    # Request handlers
└── scripts/             # Build and validation scripts
```

## Areas for Contribution

We're particularly interested in contributions in these areas:

* **New Thinking Modes**: Implement additional cognitive reasoning patterns
* **Performance Improvements**: Optimize existing algorithms and data structures
* **Test Coverage**: Improve test coverage, especially for handlers (currently 47.2%)
* **Documentation**: Improve documentation, examples, and tutorials
* **Bug Fixes**: Help us squash bugs
* **Cross-Platform Support**: Improve support for Linux and macOS
* **MCP Tool Extensions**: Add new cognitive reasoning tools

## Getting Help

If you need help, you can:

* Open an issue with the question label
* Check the [documentation](README.md)
* Review the [technical architecture guide](CLAUDE.md)

## Recognition

Contributors who submit accepted pull requests will be added to our [Contributors](#) list (coming soon).

Thank you for contributing! 🎉