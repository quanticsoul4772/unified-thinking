# Testing Guide

## Overview

This guide provides comprehensive testing strategies, examples, and best practices for the unified-thinking MCP server. The project maintains **85.7% overall test coverage** across all major components.

## Test Structure

### Test Organization

```
unified-thinking/
├── internal/
│   ├── analysis/
│   │   ├── evidence.go
│   │   └── evidence_test.go           # 89.3% coverage
│   ├── config/
│   │   └── config_test.go             # 97.3% coverage
│   ├── memory/
│   │   └── episodic_test.go           # 91.0% coverage
│   ├── metrics/
│   │   └── collector_test.go          # 100% coverage
│   ├── modes/
│   │   ├── linear_test.go
│   │   ├── tree_test.go
│   │   └── auto_test.go               # 90.5% coverage
│   ├── reasoning/
│   │   ├── probabilistic_test.go
│   │   ├── probabilistic_validation_test.go
│   │   └── concurrent_test.go         # 94.8% coverage
│   ├── server/
│   │   ├── handlers/
│   │   │   ├── probabilistic_test.go
│   │   │   └── decision_test.go
│   │   ├── server_test.go
│   │   └── validation_test.go
│   ├── storage/
│   │   ├── memory_test.go
│   │   └── sqlite_test.go             # 87.9% coverage
│   ├── types/
│   │   └── types_test.go              # 100% coverage
│   └── validation/
│       └── logic_test.go              # 88.8% coverage
```

### Running Tests

```bash
# Run all tests
make test
# Or: go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage report
make test-coverage
# Or: go test -coverprofile=coverage.out ./...
#     go tool cover -html=coverage.out

# Run with race detector
make test-race
# Or: go test -race ./...

# Run short tests only (skip long-running)
make test-short
# Or: go test -short ./...

# Test specific package
go test -v ./internal/reasoning/

# Test specific function
go test -v -run TestProbabilisticReasoner_UpdateBeliefFull ./internal/reasoning/
```

## Test Categories

### 1. Unit Tests

**Purpose**: Test individual functions and methods in isolation

**Example**: Testing Bayesian update logic

```go
// File: internal/reasoning/probabilistic_test.go

func TestProbabilisticReasoner_UpdateBeliefFull(t *testing.T) {
    pr := reasoning.NewProbabilisticReasoner()

    // Create initial belief
    belief, err := pr.CreateBelief("It will rain tomorrow", 0.5)
    require.NoError(t, err)

    // Test case 1: Supporting evidence
    t.Run("supporting evidence increases probability", func(t *testing.T) {
        updated, err := pr.UpdateBeliefFull(
            belief.ID,
            "weather-forecast-1",
            0.8,  // P(forecast says rain | will rain) = 0.8
            0.2,  // P(forecast says rain | won't rain) = 0.2
        )

        require.NoError(t, err)
        assert.Greater(t, updated.Probability, 0.5, "Probability should increase")

        // Exact calculation: P(H|E) = 0.8 * 0.5 / (0.8 * 0.5 + 0.2 * 0.5) = 0.8
        assert.InDelta(t, 0.8, updated.Probability, 0.001)
    })

    // Test case 2: Refuting evidence
    t.Run("refuting evidence decreases probability", func(t *testing.T) {
        updated, err := pr.UpdateBeliefFull(
            belief.ID,
            "weather-forecast-2",
            0.2,  // P(forecast says clear | will rain) = 0.2
            0.8,  // P(forecast says clear | won't rain) = 0.8
        )

        require.NoError(t, err)
        assert.Less(t, updated.Probability, belief.Probability)
    })

    // Test case 3: Uninformative evidence (equal likelihoods)
    t.Run("uninformative evidence does not change probability", func(t *testing.T) {
        updated, err := pr.UpdateBeliefFull(
            belief.ID,
            "uninformative-1",
            0.5,  // P(E|H) = 0.5
            0.5,  // P(E|¬H) = 0.5
        )

        require.NoError(t, err)
        assert.Equal(t, belief.Probability, updated.Probability)
        assert.True(t, updated.Metadata["last_update_uninformative"].(bool))
    })

    // Test case 4: Invalid inputs
    t.Run("rejects invalid likelihood values", func(t *testing.T) {
        _, err := pr.UpdateBeliefFull(belief.ID, "invalid-1", 1.5, 0.2)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "must be between 0 and 1")
    })
}
```

**Key Patterns**:
- Use table-driven tests for multiple scenarios
- Test happy path AND error cases
- Verify exact calculations for mathematical functions
- Use `require` for setup, `assert` for verification

### 2. Integration Tests

**Purpose**: Test interaction between multiple components

**Example**: Testing evidence pipeline integration

```go
// File: internal/integration/evidence_pipeline_test.go

func TestEvidencePipeline_ProcessEvidence(t *testing.T) {
    // Setup: Create all required components
    pr := reasoning.NewProbabilisticReasoner()
    cr := reasoning.NewCausalReasoner()
    dm := reasoning.NewDecisionMaker()
    ea := analysis.NewEvidenceAnalyzer()

    pipeline := integration.NewEvidencePipeline(pr, cr, dm, ea)

    // Create initial belief
    belief, err := pr.CreateBelief("Database optimization will improve performance", 0.6)
    require.NoError(t, err)

    // Create decision
    decision, err := dm.CreateDecision(&reasoning.DecisionRequest{
        Question: "Should we optimize the database?",
        Options: []reasoning.Option{
            {ID: "opt1", Name: "Optimize now"},
            {ID: "opt2", Name: "Optimize later"},
        },
        Criteria: []reasoning.Criterion{
            {ID: "perf", Name: "Performance", Weight: 0.7},
        },
    })
    require.NoError(t, err)

    // Process new evidence
    result, err := pipeline.ProcessEvidence(&types.Evidence{
        ID:            "benchmark-results",
        Content:       "Benchmark shows 40% improvement with indexing",
        Source:        "Performance Testing Team",
        SupportsClaim: true,
        OverallScore:  0.85,
    }, belief.ID, "", decision.ID)

    require.NoError(t, err)

    // Verify belief was updated
    assert.NotNil(t, result.UpdatedBelief)
    assert.Greater(t, result.UpdatedBelief.Probability, 0.6)

    // Verify decision was updated
    assert.NotNil(t, result.UpdatedDecision)
    assert.Equal(t, "opt1", result.UpdatedDecision.RecommendedOptionID)
}
```

**Key Patterns**:
- Test realistic workflows involving multiple systems
- Verify data flows correctly between components
- Check side effects (database updates, cache invalidation)
- Use real implementations, not mocks

### 3. Handler Tests

**Purpose**: Test MCP tool handlers end-to-end

**Example**: Testing probabilistic-reasoning tool handler

```go
// File: internal/server/handlers/probabilistic_test.go

func TestProbabilisticHandler_HandleProbabilisticReasoning(t *testing.T) {
    store := storage.NewMemoryStorage()
    pr := reasoning.NewProbabilisticReasoner()
    ea := analysis.NewEvidenceAnalyzer()
    cd := analysis.NewContradictionDetector()

    handler := handlers.NewProbabilisticHandler(store, pr, ea, cd)

    t.Run("create operation", func(t *testing.T) {
        input := handlers.ProbabilisticReasoningRequest{
            Operation: "create",
            Statement: "The feature will be popular",
            PriorProb: 0.7,
        }

        result, response, err := handler.HandleProbabilisticReasoning(
            context.Background(),
            nil,
            input,
        )

        require.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, "success", response.Status)
        assert.NotNil(t, response.Belief)
        assert.Equal(t, "The feature will be popular", response.Belief.Statement)
        assert.Equal(t, 0.7, response.Belief.Probability)
    })

    t.Run("update operation", func(t *testing.T) {
        // First create a belief
        createInput := handlers.ProbabilisticReasoningRequest{
            Operation: "create",
            Statement: "Test belief",
            PriorProb: 0.5,
        }

        _, createResp, err := handler.HandleProbabilisticReasoning(
            context.Background(),
            nil,
            createInput,
        )
        require.NoError(t, err)

        // Then update it
        updateInput := handlers.ProbabilisticReasoningRequest{
            Operation:         "update",
            BeliefID:          createResp.Belief.ID,
            EvidenceID:        "ev-001",
            LikelihoodIfTrue:  0.9,
            LikelihoodIfFalse: 0.1,
        }

        result, response, err := handler.HandleProbabilisticReasoning(
            context.Background(),
            nil,
            updateInput,
        )

        require.NoError(t, err)
        assert.Equal(t, "success", response.Status)
        assert.Greater(t, response.Belief.Probability, 0.5)
    })

    t.Run("validation errors", func(t *testing.T) {
        testCases := []struct {
            name          string
            input         handlers.ProbabilisticReasoningRequest
            expectedError string
        }{
            {
                name: "missing statement for create",
                input: handlers.ProbabilisticReasoningRequest{
                    Operation: "create",
                    PriorProb: 0.5,
                },
                expectedError: "statement is required",
            },
            {
                name: "invalid prior probability",
                input: handlers.ProbabilisticReasoningRequest{
                    Operation: "create",
                    Statement: "Test",
                    PriorProb: 1.5,
                },
                expectedError: "prior_prob must be between 0 and 1",
            },
            {
                name: "invalid likelihood",
                input: handlers.ProbabilisticReasoningRequest{
                    Operation:         "update",
                    BeliefID:          "belief-1",
                    EvidenceID:        "ev-1",
                    LikelihoodIfTrue:  -0.1,
                    LikelihoodIfFalse: 0.5,
                },
                expectedError: "likelihood_if_true must be between 0 and 1",
            },
        }

        for _, tc := range testCases {
            t.Run(tc.name, func(t *testing.T) {
                _, _, err := handler.HandleProbabilisticReasoning(
                    context.Background(),
                    nil,
                    tc.input,
                )

                require.Error(t, err)
                assert.Contains(t, err.Error(), tc.expectedError)
            })
        }
    })
}
```

**Key Patterns**:
- Test all operation types (create, update, get, combine)
- Verify response format and content
- Test validation logic with invalid inputs
- Use table-driven tests for error cases

### 4. Storage Tests

**Purpose**: Test storage layer behavior and persistence

**Example**: Testing SQLite storage with foreign keys

```go
// File: internal/storage/sqlite_test.go

func TestSQLiteStorage_ThoughtPersistence(t *testing.T) {
    // Create temporary database
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    store, err := storage.NewSQLiteStorage(dbPath, 5000)
    require.NoError(t, err)
    defer store.Close()

    // Create a branch first (foreign key requirement)
    branch := &types.Branch{
        State:      "active",
        Priority:   0.8,
        Confidence: 0.7,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }

    err = store.StoreBranch(branch)
    require.NoError(t, err)

    // Create thought with valid branch reference
    thought := &types.Thought{
        Content:    "Test thought",
        Mode:       types.ModeLinear,
        BranchID:   branch.ID,
        Confidence: 0.8,
        Timestamp:  time.Now(),
        KeyPoints:  []string{"key1", "key2"},
        Metadata:   map[string]interface{}{"test": "value"},
    }

    err = store.StoreThought(thought)
    require.NoError(t, err)

    // Retrieve and verify
    retrieved, err := store.GetThought(thought.ID)
    require.NoError(t, err)

    assert.Equal(t, thought.Content, retrieved.Content)
    assert.Equal(t, thought.BranchID, retrieved.BranchID)
    assert.Equal(t, thought.KeyPoints, retrieved.KeyPoints)
    assert.Equal(t, thought.Metadata, retrieved.Metadata)

    // Test full-text search
    results := store.SearchThoughts("Test", "", 10, 0)
    assert.Len(t, results, 1)
    assert.Equal(t, thought.ID, results[0].ID)
}

func TestSQLiteStorage_CacheWarming(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    // Create database and add 100 thoughts
    store, err := storage.NewSQLiteStorage(dbPath, 5000)
    require.NoError(t, err)

    for i := 0; i < 100; i++ {
        thought := &types.Thought{
            Content:    fmt.Sprintf("Thought %d", i),
            Mode:       types.ModeLinear,
            Confidence: 0.8,
            Timestamp:  time.Now(),
        }
        err := store.StoreThought(thought)
        require.NoError(t, err)
    }

    store.Close()

    // Reopen with custom warm cache limit
    t.Setenv("WARM_CACHE_LIMIT", "50")

    store, err = storage.NewSQLiteStorage(dbPath, 5000)
    require.NoError(t, err)
    defer store.Close()

    // Verify cache was warmed (check metrics or search performance)
    metrics := store.GetMetrics()
    assert.Greater(t, metrics.TotalThoughts, int64(0))
}
```

**Key Patterns**:
- Use `t.TempDir()` for test databases
- Test persistence across close/reopen cycles
- Verify foreign key constraints
- Test cache behavior and performance

### 5. Concurrent Safety Tests

**Purpose**: Test thread safety under concurrent access

**Example**: Testing concurrent probabilistic updates

```go
// File: internal/reasoning/concurrent_test.go

func TestProbabilisticReasoner_ConcurrentUpdates(t *testing.T) {
    pr := reasoning.NewProbabilisticReasoner()

    // Create initial belief
    belief, err := pr.CreateBelief("Concurrent test", 0.5)
    require.NoError(t, err)

    // Run 100 concurrent updates
    const numGoroutines = 100
    var wg sync.WaitGroup
    wg.Add(numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        go func(idx int) {
            defer wg.Done()

            _, err := pr.UpdateBeliefFull(
                belief.ID,
                fmt.Sprintf("evidence-%d", idx),
                0.7,
                0.3,
            )

            assert.NoError(t, err)
        }(i)
    }

    wg.Wait()

    // Verify final state is consistent
    final, err := pr.GetBelief(belief.ID)
    require.NoError(t, err)
    assert.NotNil(t, final)
    assert.Len(t, final.Evidence, numGoroutines)
}
```

**Key Patterns**:
- Use `sync.WaitGroup` to coordinate goroutines
- Test with `-race` flag to detect data races
- Verify final state is consistent
- Test high concurrency levels (100+ goroutines)

## Test Best Practices

### 1. Arrange-Act-Assert Pattern

```go
func TestExample(t *testing.T) {
    // Arrange: Set up test fixtures
    store := storage.NewMemoryStorage()
    handler := handlers.NewHandler(store)
    input := createTestInput()

    // Act: Execute the code under test
    result, err := handler.Process(input)

    // Assert: Verify the results
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### 2. Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    testCases := []struct {
        name          string
        input         Input
        expectedError string
    }{
        {"valid input", validInput, ""},
        {"missing field", incompleteInput, "field is required"},
        {"invalid range", outOfRangeInput, "must be between"},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := Validate(tc.input)

            if tc.expectedError == "" {
                assert.NoError(t, err)
            } else {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tc.expectedError)
            }
        })
    }
}
```

### 3. Test Helpers

```go
// helper functions reduce boilerplate
func createTestBelief(t *testing.T, pr *reasoning.ProbabilisticReasoner) *types.ProbabilisticBelief {
    t.Helper()

    belief, err := pr.CreateBelief("Test belief", 0.5)
    require.NoError(t, err)

    return belief
}

func assertProbabilityInRange(t *testing.T, prob float64, min, max float64) {
    t.Helper()

    assert.GreaterOrEqual(t, prob, min)
    assert.LessOrEqual(t, prob, max)
}
```

### 4. Cleanup with defer

```go
func TestWithResources(t *testing.T) {
    store, err := storage.NewSQLiteStorage(dbPath, 5000)
    require.NoError(t, err)
    defer store.Close() // Always cleanup

    // Test code...
}
```

## Coverage Goals

### Package Coverage Targets

| Package          | Current | Target | Priority |
|------------------|---------|--------|----------|
| types            | 100%    | 100%   | High     |
| metrics          | 100%    | 100%   | High     |
| config           | 97.3%   | 95%+   | Medium   |
| reasoning        | 94.8%   | 90%+   | High     |
| memory           | 91.0%   | 90%+   | Medium   |
| modes            | 90.5%   | 85%+   | Medium   |
| analysis         | 89.3%   | 85%+   | Medium   |
| validation       | 88.8%   | 85%+   | Medium   |
| storage          | 87.9%   | 85%+   | High     |
| orchestration    | 87.7%   | 85%+   | Medium   |
| metacognition    | 87.2%   | 85%+   | Medium   |

**Overall Target**: Maintain > 85% coverage

### Uncovered Code Patterns

**Acceptable uncovered code**:
1. Error paths that are theoretically unreachable
2. Defensive nil checks after validation
3. Debug logging statements
4. Main function initialization

**Unacceptable uncovered code**:
1. Business logic
2. Data transformations
3. Validation logic
4. Error handling

## Continuous Integration

### Pre-Commit Checks

```bash
# Run before committing
make test          # All tests must pass
make test-race     # No data races
make test-coverage # Check coverage report
```

### CI Pipeline

```yaml
# .github/workflows/test.yml (example)
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run tests
        run: make test

      - name: Run tests with race detector
        run: make test-race

      - name: Generate coverage
        run: make test-coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Debugging Failed Tests

### 1. Verbose Output

```bash
# See detailed test output
go test -v ./internal/reasoning/

# Run specific test with verbose output
go test -v -run TestProbabilisticReasoner_UpdateBeliefFull ./internal/reasoning/
```

### 2. Test Logging

```go
func TestWithLogging(t *testing.T) {
    // Enable debug logging for this test
    t.Setenv("DEBUG", "true")

    // Your test code...

    t.Logf("Intermediate state: %+v", state)
}
```

### 3. Test Data Inspection

```go
func TestDebug(t *testing.T) {
    result, err := Function()

    // Print full structure for debugging
    t.Logf("Result: %#v", result)
    t.Logf("Error: %v", err)
}
```

## Conclusion

This testing guide provides a comprehensive framework for maintaining high-quality code through systematic testing. Key principles:

1. **Test Coverage**: Maintain > 85% coverage across all packages
2. **Test Categories**: Unit, integration, handler, storage, concurrent
3. **Best Practices**: AAA pattern, table-driven tests, cleanup
4. **CI Integration**: Automated testing on every commit
5. **Debugging**: Verbose output and detailed logging

Regular testing ensures the unified-thinking MCP server remains reliable, correct, and maintainable as it evolves.
