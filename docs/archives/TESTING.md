# Testing Guide - Unified Thinking Server

**Current Test Coverage**: 73.6% overall
- Storage: 80.5%
- Validation: 94.2%
- Modes: High coverage
- Analysis: Partial coverage
- Orchestration: Partial coverage

---

## Quick Start

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run with coverage report
make test-coverage

# Run with race detector
make test-race

# Run specific package
go test -v ./internal/modes/

# Run benchmarks
make benchmark
```

---

## Test Organization

### Test Files by Package

```
internal/
├── modes/           # Thinking modes (linear, tree, divergent, auto)
│   ├── *_test.go
│   └── error_handling_test.go
├── storage/         # Storage backends (memory, SQLite)
│   ├── memory_test.go
│   ├── sqlite_test.go
│   ├── optimization_test.go
│   └── factory_test.go
├── validation/      # Logic validation and fallacy detection
│   ├── logic_test.go
│   ├── logic_validation_test.go
│   └── fallacies_test.go
├── reasoning/       # Probabilistic, causal, temporal reasoning
│   ├── probabilistic_test.go
│   ├── probabilistic_validation_test.go
│   ├── causal_test.go
│   ├── analogical_test.go
│   ├── temporal_test.go
│   └── concurrent_test.go
├── analysis/        # Evidence, contradiction, perspective analysis
│   ├── evidence_test.go
│   ├── perspective_test.go
│   └── concurrent_test.go
├── metacognition/   # Self-evaluation and bias detection
│   ├── self_eval_test.go
│   ├── bias_detection_test.go
│   └── concurrent_test.go
├── integration/     # Cross-mode synthesis
│   └── synthesizer_test.go
├── orchestration/   # Workflow orchestration
│   ├── workflow_test.go
│   ├── workflow_execution_test.go
│   ├── context_test.go
│   └── helpers_test.go
├── server/          # MCP server and handlers
│   ├── server_test.go
│   ├── validation_test.go
│   ├── auto_validation_test.go
│   └── handlers/*_test.go
└── types/           # Data structures and builders
    ├── types_test.go
    └── builders_test.go
```

---

## Testing Best Practices

### 1. Table-Driven Tests

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"basic case", "input1", "output1", false},
        {"error case", "bad", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### 2. Concurrent Testing

Use `go test -race` to detect race conditions:

```go
func TestConcurrent(t *testing.T) {
    storage := memory.NewMemoryStorage()
    var wg sync.WaitGroup

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            thought := types.NewThought().
                Content(fmt.Sprintf("thought-%d", id)).
                Build()
            storage.StoreThought(thought)
        }(i)
    }

    wg.Wait()
}
```

### 3. Test Helpers

Create reusable test helpers:

```go
func createTestThought(t *testing.T, content string) *types.Thought {
    t.Helper()
    return types.NewThought().
        Content(content).
        Mode(types.ModeLinear).
        Confidence(0.8).
        Build()
}
```

---

## Coverage Goals

### Target Coverage by Package

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| types | High | 90% | Medium |
| storage | 80.5% | 85% | Medium |
| modes | High | 90% | High |
| validation | 94.2% | 95% | Low |
| reasoning | Partial | 80% | High |
| analysis | Partial | 80% | High |
| metacognition | Partial | 80% | High |
| integration | Partial | 75% | Medium |
| orchestration | Partial | 80% | High |
| server | High | 85% | High |

### Measuring Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage by function
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Coverage for specific package
go test -coverprofile=storage.out ./internal/storage/
go tool cover -html=storage.out
```

---

## Benchmarking

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkBayesUpdate -benchmem ./internal/reasoning/

# Compare benchmarks
go test -bench=. -benchmem ./internal/storage/ > old.txt
# Make changes
go test -bench=. -benchmem ./internal/storage/ > new.txt
benchcmp old.txt new.txt
```

### Writing Benchmarks

```go
func BenchmarkBayesUpdate(b *testing.B) {
    pr := reasoning.NewProbabilisticReasoner()
    belief, _ := pr.CreateBelief("test", 0.5)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        pr.UpdateBelief(belief.ID, "evidence", 0.8, 0.6)
    }
}
```

---

## Integration Testing

### Testing Workflows

```go
func TestWorkflowExecution(t *testing.T) {
    orchestrator := orchestration.NewOrchestrator()

    workflow := &orchestration.Workflow{
        ID:   "test-workflow",
        Type: orchestration.WorkflowSequential,
        Steps: []*orchestration.WorkflowStep{
            {ID: "step1", Tool: "think"},
            {ID: "step2", Tool: "validate", DependsOn: []string{"step1"}},
        },
    }

    result, err := orchestrator.ExecuteWorkflow(context.Background(), workflow, map[string]interface{}{
        "problem": "test problem",
    })

    assert.NoError(t, err)
    assert.Equal(t, "success", result.Status)
}
```

### Testing Storage Backends

```go
func TestStorageBackend(t *testing.T) {
    // Test both memory and SQLite
    backends := map[string]storage.Storage{
        "memory": memory.NewMemoryStorage(),
        "sqlite": sqlite.NewSQLiteStorage(":memory:"),
    }

    for name, store := range backends {
        t.Run(name, func(t *testing.T) {
            // Test storage operations
            thought := createTestThought(t, "test")
            err := store.StoreThought(thought)
            assert.NoError(t, err)

            retrieved, err := store.GetThought(thought.ID)
            assert.NoError(t, err)
            assert.Equal(t, thought.Content, retrieved.Content)
        })
    }
}
```

---

## Test Data

### Using Test Fixtures

```go
// testdata/logical_reasoning.json
{
    "tests": [
        {
            "name": "modus_ponens",
            "premises": ["If P then Q", "P"],
            "conclusion": "Q",
            "expected_valid": true
        }
    ]
}

// Load in tests
func loadTestData(t *testing.T, filename string) []TestCase {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    if err != nil {
        t.Fatal(err)
    }

    var tests []TestCase
    json.Unmarshal(data, &tests)
    return tests
}
```

---

## Continuous Integration

### Pre-Commit Checks

```bash
# Run before committing
make pre-commit

# Which runs:
# - go test -short ./...  (fast tests)
# - go vet ./...
# - go fmt ./...
```

### CI Pipeline (Recommended)

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Run tests
        run: make test

      - name: Run race detector
        run: make test-race

      - name: Generate coverage
        run: make test-coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

---

## Troubleshooting Tests

### Common Issues

**Tests hang or timeout**:
```bash
# Run with timeout
go test -timeout 30s ./...

# Identify slow tests
go test -v -timeout 30s ./... | grep -E '(PASS|FAIL).*[0-9]+\.[0-9]+s'
```

**Race conditions detected**:
```bash
# Run with race detector
go test -race ./...

# Focus on specific package
go test -race -v ./internal/storage/
```

**Flaky tests**:
```bash
# Run tests multiple times
go test -count=10 ./internal/modes/

# Run with verbose output to see patterns
go test -v -count=10 ./internal/modes/ 2>&1 | tee test-output.log
```

**Coverage not updating**:
```bash
# Clean test cache
go clean -testcache

# Rebuild and retest
make clean build test-coverage
```

---

## Test Patterns

### Testing Error Conditions

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {"empty input", "", ErrEmptyInput},
        {"invalid format", "bad", ErrInvalidFormat},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Process(tt.input)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got error %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### Testing with Mocks

```go
type MockStorage struct {
    mock.Mock
}

func (m *MockStorage) StoreThought(t *types.Thought) error {
    args := m.Called(t)
    return args.Error(0)
}

func TestWithMock(t *testing.T) {
    mockStore := new(MockStorage)
    mockStore.On("StoreThought", mock.Anything).Return(nil)

    // Use mockStore in test
    err := mockStore.StoreThought(&types.Thought{})
    assert.NoError(t, err)
    mockStore.AssertExpectations(t)
}
```

---

## Performance Testing

### Memory Profiling

```bash
# Generate memory profile
go test -memprofile=mem.prof -bench=. ./internal/storage/

# Analyze profile
go tool pprof mem.prof
# In pprof: top, list <function>
```

### CPU Profiling

```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof -bench=. ./internal/storage/

# Analyze profile
go tool pprof cpu.prof
# In pprof: top, web
```

---

## Test Coverage Reports

### Viewing Coverage

```bash
# Generate and view HTML coverage
make test-coverage
# Opens coverage.html in browser

# Terminal-friendly coverage
go tool cover -func=coverage.out | grep -E 'total:|^unified-thinking'
```

### Coverage by Package

```bash
# Storage layer coverage
make test-storage-coverage

# Generate report for all packages
for pkg in $(go list ./internal/...); do
    go test -coverprofile=profile.out $pkg
    echo "Coverage for $pkg:"
    go tool cover -func=profile.out | tail -1
done
```

---

## Additional Resources

- **Go Testing Documentation**: https://golang.org/pkg/testing/
- **Testify Library**: https://github.com/stretchr/testify
- **Table-Driven Tests**: https://github.com/golang/go/wiki/TableDrivenTests
- **Test Coverage**: https://go.dev/blog/cover

---

**Last Updated**: 2025-10-07
**Test Coverage Goal**: 90%
**Current Coverage**: 73.6%
