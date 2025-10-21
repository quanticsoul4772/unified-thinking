# GitHub Workflows Analysis & Recommendations

## Current State

### Existing Workflows
1. **claude.yml** - Claude Code integration via @claude mentions
2. **claude-code-review.yml** - Automated PR reviews with Claude Code

### Project Characteristics
- **Language**: Go 1.23.0
- **Architecture**: MCP server with 50+ cognitive reasoning tools
- **Build Targets**: Windows (.exe), Linux (amd64), potential macOS
- **Storage**: In-memory + SQLite backends
- **Testing**: Comprehensive suite with coverage, race detection, benchmarks
- **Dependencies**: MCP SDK, modernc.org/sqlite (pure Go, no CGO)

## Recommended Workflows

### 1. CI/CD Pipeline (Critical - High Priority)

**Purpose**: Automated testing, building, and validation on every push/PR

**Workflow File**: `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.23.x', '1.24.x']

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Verify dependencies
        run: go mod verify

      - name: Run go vet
        run: go vet ./...

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          flags: unittests
          token: ${{ secrets.CODECOV_TOKEN }}

      - name: Run benchmarks (for performance regression detection)
        run: go test -bench=. -benchmem ./internal/storage/ | tee benchmark.txt

      - name: Store benchmark result
        uses: benchmark-action/github-action-benchmark@v1
        if: github.event_name == 'push' && github.ref == 'refs/heads/main'
        with:
          tool: 'go'
          output-file-path: benchmark.txt
          github-token: ${{ secrets.GITHUB_TOKEN }}
          auto-push: true

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test, lint]
    strategy:
      matrix:
        include:
          - goos: windows
            goarch: amd64
            ext: .exe
          - goos: linux
            goarch: amd64
            ext: ''
          - goos: darwin
            goarch: amd64
            ext: ''
          - goos: darwin
            goarch: arm64
            ext: ''

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          cache: true

      - name: Build binary
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          mkdir -p bin
          go build -v -o bin/unified-thinking-${{ matrix.goos }}-${{ matrix.goarch }}${{ matrix.ext }} ./cmd/server

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: unified-thinking-${{ matrix.goos }}-${{ matrix.goarch }}
          path: bin/unified-thinking-${{ matrix.goos }}-${{ matrix.goarch }}${{ matrix.ext }}
          retention-days: 30
```

**Benefits**:
- Runs tests on every push/PR
- Multi-version Go testing (1.23.x, 1.24.x)
- Race condition detection
- Code coverage tracking
- Linting with golangci-lint
- Cross-platform build verification
- Performance regression tracking

---

### 2. Release Automation (High Priority)

**Purpose**: Automated releases with cross-platform binaries

**Workflow File**: `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          cache: true

      - name: Run tests
        run: go test -race ./...

      - name: Build binaries
        run: |
          # Windows AMD64
          GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${{ github.ref_name }}" -o bin/unified-thinking-windows-amd64.exe ./cmd/server

          # Linux AMD64
          GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${{ github.ref_name }}" -o bin/unified-thinking-linux-amd64 ./cmd/server

          # Linux ARM64
          GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.Version=${{ github.ref_name }}" -o bin/unified-thinking-linux-arm64 ./cmd/server

          # macOS AMD64
          GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${{ github.ref_name }}" -o bin/unified-thinking-darwin-amd64 ./cmd/server

          # macOS ARM64 (Apple Silicon)
          GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.Version=${{ github.ref_name }}" -o bin/unified-thinking-darwin-arm64 ./cmd/server

      - name: Generate checksums
        run: |
          cd bin
          sha256sum * > SHA256SUMS
          cd ..

      - name: Generate changelog
        id: changelog
        run: |
          # Extract changelog for this version from CHANGELOG.md
          VERSION=${{ github.ref_name }}
          echo "## Release ${VERSION}" > release_notes.md
          echo "" >> release_notes.md

          # Add installation instructions
          cat >> release_notes.md << 'EOF'
          ## Installation

          Download the appropriate binary for your platform:
          - **Windows**: `unified-thinking-windows-amd64.exe`
          - **Linux**: `unified-thinking-linux-amd64` or `unified-thinking-linux-arm64`
          - **macOS**: `unified-thinking-darwin-amd64` (Intel) or `unified-thinking-darwin-arm64` (Apple Silicon)

          Configure in Claude Desktop:
          ```json
          {
            "mcpServers": {
              "unified-thinking": {
                "command": "/path/to/unified-thinking",
                "transport": "stdio",
                "env": {
                  "DEBUG": "true"
                }
              }
            }
          }
          ```

          ## Changes
          EOF

          # Extract version-specific changes from CHANGELOG.md if it exists
          if [ -f CHANGELOG.md ]; then
            sed -n "/## \[${VERSION#v}\]/,/## \[/p" CHANGELOG.md | sed '$d' >> release_notes.md
          fi

      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            bin/*
          body_path: release_notes.md
          draft: false
          prerelease: false
          generate_release_notes: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Benefits**:
- Automatic releases on version tags (v1.0.0, v2.1.3, etc.)
- Cross-platform binaries (Windows, Linux, macOS - both Intel and ARM)
- SHA256 checksums for integrity verification
- Automatic changelog generation
- Installation instructions in release notes

**Usage**:
```bash
git tag v1.0.0
git push origin v1.0.0
```

---

### 3. Code Quality & Security (High Priority)

**Purpose**: Continuous security scanning and code quality checks

**Workflow File**: `.github/workflows/security.yml`

```yaml
name: Security & Quality

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday at midnight

jobs:
  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          cache: true

      - name: Run Gosec Security Scanner
        uses: securego/gosec@v2.20.0
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'

      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec-results.sarif

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

  dependency-review:
    name: Dependency Review
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Dependency Review
        uses: actions/dependency-review-action@v4
        with:
          fail-on-severity: moderate

  codeql:
    name: CodeQL Analysis
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      actions: read
      contents: read

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: go

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
```

**Benefits**:
- Security vulnerability scanning with gosec
- Dependency vulnerability checking with govulncheck
- Advanced code analysis with CodeQL
- Weekly scheduled scans
- Integration with GitHub Security tab

---

### 4. Dependency Management

**Purpose**: Keep dependencies up-to-date automatically

**Workflow File**: `.github/dependabot.yml`

```yaml
version: 2
updates:
  # Go modules
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
    open-pull-requests-limit: 5
    reviewers:
      - "quanticsoul4772"
    commit-message:
      prefix: "deps"
      include: "scope"
    labels:
      - "dependencies"
      - "go"

  # GitHub Actions
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
    open-pull-requests-limit: 3
    commit-message:
      prefix: "ci"
    labels:
      - "dependencies"
      - "github-actions"
```

**Benefits**:
- Automatic dependency updates
- Separate updates for Go modules and GitHub Actions
- Weekly schedule to avoid noise
- Automatic PR creation with labels

---

### 5. Documentation & CHANGELOG Validation (Medium Priority)

**Purpose**: Ensure documentation quality and changelog updates

**Workflow File**: `.github/workflows/docs.yml`

```yaml
name: Documentation

on:
  pull_request:
    branches: [ main, develop ]
    paths:
      - '**.md'
      - 'docs/**'
      - '.github/workflows/docs.yml'

jobs:
  markdown-lint:
    name: Markdown Lint
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Lint Markdown files
        uses: articulate/actions-markdownlint@v1
        with:
          config: .markdownlint.json
          ignore: vendor
          files: '*.md docs/**/*.md'

  changelog-check:
    name: Changelog Check
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Check if CHANGELOG.md was updated
        run: |
          # Skip for dependabot PRs
          if [[ "${{ github.actor }}" == "dependabot[bot]" ]]; then
            echo "Skipping changelog check for dependabot"
            exit 0
          fi

          # Check if CHANGELOG.md was modified
          git diff --name-only origin/${{ github.base_ref }}...HEAD | grep -q "CHANGELOG.md" || {
            echo "::warning::CHANGELOG.md was not updated. Consider adding an entry for this PR."
            exit 0
          }

          echo "✓ CHANGELOG.md was updated"

  docs-build-test:
    name: Test Documentation Links
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Check markdown links
        uses: gaurav-nelson/github-action-markdown-link-check@v1
        with:
          use-quiet-mode: 'yes'
          use-verbose-mode: 'no'
          config-file: '.markdown-link-check.json'
```

**Benefits**:
- Markdown linting for consistent documentation
- Changelog update reminders on PRs
- Broken link detection
- Documentation quality gates

---

### 6. Performance Benchmarking (Medium Priority)

**Purpose**: Track performance over time and detect regressions

**Workflow File**: `.github/workflows/benchmark.yml`

```yaml
name: Benchmark

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

permissions:
  contents: write
  pull-requests: write

jobs:
  benchmark:
    name: Performance Benchmark
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          cache: true

      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -benchtime=5s ./... | tee benchmark-output.txt

      - name: Store benchmark result
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: benchmark-output.txt
          github-token: ${{ secrets.GITHUB_TOKEN }}
          auto-push: ${{ github.event_name == 'push' && github.ref == 'refs/heads/main' }}
          alert-threshold: '150%'
          comment-on-alert: true
          fail-on-alert: false
          alert-comment-cc-users: '@quanticsoul4772'
```

**Benefits**:
- Automated benchmark execution
- Performance regression detection (alerts on >150% slowdown)
- Historical performance tracking
- PR comments for performance impacts

---

### 7. Storage Backend Testing (Medium Priority)

**Purpose**: Dedicated workflow for storage layer validation

**Workflow File**: `.github/workflows/storage-test.yml`

```yaml
name: Storage Layer Tests

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'internal/storage/**'
      - '.github/workflows/storage-test.yml'
  pull_request:
    branches: [ main, develop ]
    paths:
      - 'internal/storage/**'

jobs:
  test-memory-backend:
    name: Test In-Memory Storage
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          cache: true

      - name: Run memory storage tests
        run: go test -v -race -coverprofile=memory-coverage.out ./internal/storage/
        env:
          STORAGE_TYPE: memory

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./memory-coverage.out
          flags: storage-memory
          token: ${{ secrets.CODECOV_TOKEN }}

  test-sqlite-backend:
    name: Test SQLite Storage
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          cache: true

      - name: Run SQLite storage tests
        run: go test -v -race -coverprofile=sqlite-coverage.out ./internal/storage/
        env:
          STORAGE_TYPE: sqlite
          SQLITE_PATH: ":memory:"

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./sqlite-coverage.out
          flags: storage-sqlite
          token: ${{ secrets.CODECOV_TOKEN }}

  test-storage-migrations:
    name: Test Storage Migrations
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          cache: true

      - name: Test schema migrations
        run: |
          go test -v ./internal/storage/ -run TestMigration
```

**Benefits**:
- Dedicated storage testing
- Both backends tested independently
- Migration testing
- Storage-specific coverage tracking

---

## Implementation Priority

### Phase 1: Essential (Week 1)
1. **CI/CD Pipeline** - Critical for code quality
2. **Security Scanning** - Essential for safety
3. **Dependabot** - Keep dependencies current

### Phase 2: Release Management (Week 2)
4. **Release Automation** - Streamline releases
5. **Benchmark Tracking** - Performance monitoring

### Phase 3: Documentation & Quality (Week 3)
6. **Documentation Validation** - Maintain doc quality
7. **Storage Testing** - Deep storage validation

## Configuration Files Needed

### 1. `.golangci.yml` - Linter configuration

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gofmt
    - goimports
    - misspell
    - goconst
    - gocyclo
    - gosec
    - unconvert
    - dupl

linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 3
    min-occurrences: 3
  misspell:
    locale: US
  errcheck:
    check-blank: true

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
        - gosec
```

### 2. `.markdownlint.json` - Markdown linter config

```json
{
  "default": true,
  "MD013": false,
  "MD033": false,
  "MD041": false
}
```

### 3. `.markdown-link-check.json` - Link checker config

```json
{
  "ignorePatterns": [
    {
      "pattern": "^http://localhost"
    }
  ],
  "timeout": "20s",
  "retryOn429": true,
  "retryCount": 3,
  "fallbackRetryDelay": "30s"
}
```

## Secrets Configuration

Add these secrets in GitHub repository settings:

1. **CODECOV_TOKEN** - For code coverage reporting
   - Get from codecov.io after connecting your repository

2. **CLAUDE_CODE_OAUTH_TOKEN** - Already configured for Claude workflows

## Monitoring & Badges

Add to README.md:

```markdown
[![CI](https://github.com/quanticsoul4772/unified-thinking/workflows/CI/badge.svg)](https://github.com/quanticsoul4772/unified-thinking/actions/workflows/ci.yml)
[![Security](https://github.com/quanticsoul4772/unified-thinking/workflows/Security%20&%20Quality/badge.svg)](https://github.com/quanticsoul4772/unified-thinking/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/quanticsoul4772/unified-thinking/branch/main/graph/badge.svg)](https://codecov.io/gh/quanticsoul4772/unified-thinking)
[![Go Report Card](https://goreportcard.com/badge/github.com/quanticsoul4772/unified-thinking)](https://goreportcard.com/report/github.com/quanticsoul4772/unified-thinking)
[![License](https://img.shields.io/github/license/quanticsoul4772/unified-thinking)](LICENSE)
```

## Estimated Impact

### Code Quality
- **Before**: Manual testing, no automated validation
- **After**: Automated testing on every commit, 70%+ coverage tracking

### Release Process
- **Before**: Manual builds, no distribution
- **After**: One-command releases with multi-platform binaries

### Security
- **Before**: No automated security scanning
- **After**: Weekly scans, vulnerability alerts, dependency monitoring

### Performance
- **Before**: No performance tracking
- **After**: Benchmark tracking, regression detection

## Next Steps

1. Review and approve workflow suggestions
2. Create configuration files (.golangci.yml, etc.)
3. Set up Codecov account and token
4. Implement Phase 1 workflows
5. Test workflows on a feature branch
6. Roll out remaining workflows
7. Add status badges to README

## Notes

- All workflows use modern GitHub Actions (v4/v5)
- Workflows are optimized for caching to reduce build times
- Security workflows integrate with GitHub Security tab
- Release workflow supports semantic versioning
- Benchmark workflow provides performance regression alerts
