# Benchmark Framework Implementation Plan

## Executive Summary

Implement a native Go benchmark framework to objectively measure unified-thinking server reasoning quality, validate improvements, and prevent regressions. The framework will integrate with existing CI/CD, focus on cognitive capabilities specific to the thinking modes, and provide continuous validation of episodic learning improvements.

**Timeline**: 4 weeks
**Complexity**: Medium
**Value**: High (enables data-driven optimization)

## Goals

1. Measure reasoning quality across all thinking modes
2. Validate episodic memory learning effectiveness
3. Detect regressions in CI/CD pipeline
4. Track improvement trends over time
5. Provide objective metrics for feature development

## Architecture

### Components

```
benchmarks/
├── datasets/           # Benchmark problem sets
│   ├── reasoning/      # Logic and multi-step reasoning
│   ├── causal/         # Causal inference problems
│   ├── probabilistic/  # Bayesian reasoning tests
│   └── metacognitive/  # Self-evaluation tests
├── evaluators/         # Metric computation
│   ├── accuracy.go     # Correctness evaluation
│   ├── calibration.go  # Confidence alignment
│   └── efficiency.go   # Performance metrics
├── runners/            # Test execution
│   ├── suite.go        # Benchmark suite runner
│   └── parallel.go     # Parallel execution
├── reporting/          # Results output
│   ├── markdown.go     # Report generation
│   └── timeseries.go   # Trend tracking
└── integration/        # CI/CD integration
    └── github_action.go
```

### Data Flow

```
1. Load Dataset → 2. Execute via MCP → 3. Evaluate Response → 4. Compute Metrics → 5. Generate Report
```

## Implementation Phases

### Phase 1: Foundation (Week 1)

**Goal**: Basic benchmark infrastructure with one complete test suite

**Tasks**:
1. Create `benchmarks/` directory structure
2. Implement `benchmarks/suite.go` - benchmark suite runner
3. Create `benchmarks/datasets/reasoning/logic_puzzles.json` - 50 logic problems
4. Implement `benchmarks/evaluators/accuracy.go` - exact match evaluator
5. Add `make benchmark` target to Makefile
6. Create first test suite in `benchmarks/reasoning_test.go`

**Deliverables**:
- Working benchmark command: `make benchmark`
- 50 logic puzzle tests with accuracy measurement
- Baseline performance report

**Validation Criteria**:
- Tests run successfully via `go test`
- Report shows accuracy percentage
- Execution completes in under 2 minutes

### Phase 2: Core Metrics (Week 2)

**Goal**: Add confidence calibration and efficiency metrics

**Tasks**:
1. Implement `benchmarks/evaluators/calibration.go` - ECE computation
2. Implement `benchmarks/evaluators/efficiency.go` - latency/token tracking
3. Create `benchmarks/datasets/probabilistic/bayes_problems.json` - 30 Bayesian problems
4. Create `benchmarks/datasets/causal/causal_inference.json` - 25 causal reasoning tests
5. Add calibration suite in `benchmarks/probabilistic_test.go`
6. Add causal reasoning suite in `benchmarks/causal_test.go`

**Deliverables**:
- Confidence calibration measurement (Expected Calibration Error)
- Performance metrics (latency, tokens per task)
- 3 complete test suites (logic, probabilistic, causal)

**Validation Criteria**:
- ECE calculated correctly (validated against known examples)
- Latency tracking works
- 105 total test problems executing

### Phase 3: Episodic Learning Validation (Week 3)

**Goal**: Validate that episodic memory improves performance over time

**Tasks**:
1. Create `benchmarks/datasets/learning/repeated_problems.json` - same problems multiple times
2. Implement `benchmarks/evaluators/learning.go` - improvement trend analysis
3. Add `benchmarks/learning_test.go` - tests episodic memory effectiveness
4. Implement trajectory analysis tools
5. Add statistical significance testing (t-tests)

**Deliverables**:
- Learning effectiveness tests showing improvement curves
- Statistical validation of learning (p-values)
- Visualization of performance trends

**Validation Criteria**:
- Can detect 10%+ improvement over 10 iterations
- Statistical tests work correctly
- Clear visualization of learning trends

### Phase 4: CI/CD Integration & Reporting (Week 4)

**Goal**: Automate benchmarks in CI/CD with regression detection

**Tasks**:
1. Create `.github/workflows/benchmarks.yml` - GitHub Actions workflow
2. Implement `benchmarks/reporting/markdown.go` - report generation
3. Implement `benchmarks/reporting/timeseries.go` - trend storage
4. Add regression detection logic (2% threshold)
5. Create benchmark result storage (SQLite or JSON)
6. Generate visual reports (ASCII charts or markdown tables)

**Deliverables**:
- Automated benchmark runs on PR and weekly schedule
- Regression detection blocking PRs with >2% degradation
- Historical trend tracking
- Markdown reports committed to repo

**Validation Criteria**:
- GitHub Actions workflow runs successfully
- Regression detection catches intentional degradation
- Reports are readable and actionable
- Historical data accumulates correctly

## Benchmark Datasets

### 1. Logic & Reasoning (50 problems)
- Propositional logic problems
- Syllogistic reasoning
- Logical puzzles (Knights and Knaves)
- Deductive reasoning chains

**Evaluation**: Exact match on conclusion

### 2. Probabilistic Reasoning (30 problems)
- Bayesian updates given evidence
- Base rate problems
- Medical test scenarios
- Sequential probability updates

**Evaluation**: Numerical accuracy within 5%, confidence calibration

### 3. Causal Reasoning (25 problems)
- Correlation vs causation scenarios
- Intervention prediction
- Counterfactual reasoning
- Confounding variable identification

**Evaluation**: Correctness of causal graph structure and intervention predictions

### 4. Metacognitive Tasks (20 problems)
- Confidence estimation tasks
- Bias identification
- Unknown unknowns detection
- Self-evaluation accuracy

**Evaluation**: Calibration quality, bias detection precision

### 5. Learning Effectiveness (20 repeated problems)
- Same problems presented multiple times
- Track improvement over iterations
- Measure learning rate

**Evaluation**: Performance improvement trend, statistical significance

## Metrics Framework

### Primary Metrics

**Accuracy**: Percentage of correct responses
```go
accuracy = correct_responses / total_responses
```

**Confidence Calibration (ECE)**: Expected Calibration Error
```go
ECE = Σ |accuracy_i - confidence_i| * (n_i / N)
// Bucket predictions by confidence, compare to actual accuracy
```

**Efficiency**: Average latency and tokens per task
```go
avg_latency = total_time / num_tasks
tokens_per_task = total_tokens / num_tasks
```

### Secondary Metrics

**Learning Rate**: Improvement velocity
```go
learning_rate = (accuracy_final - accuracy_initial) / num_iterations
```

**Strategy Distribution**: Mode usage patterns
```go
mode_distribution = count_per_mode / total_tasks
```

**Error Types**: Categorized failure analysis
- Logical errors
- Confidence mis-calibration
- Timeout/resource errors

## Technical Specifications

### Benchmark Suite Structure

```go
type BenchmarkSuite struct {
    Name     string
    Problems []Problem
    Evaluator Evaluator
}

type Problem struct {
    ID          string
    Description string
    Input       map[string]interface{}
    Expected    interface{}
    Metadata    map[string]interface{}
}

type Evaluator interface {
    Evaluate(response interface{}, expected interface{}) Result
}

type Result struct {
    Correct    bool
    Score      float64
    Confidence float64
    Latency    time.Duration
    Tokens     int
    Error      string
}
```

### MCP Client Integration

```go
import "github.com/modelcontextprotocol/go-sdk/mcp"

func executeBenchmark(problem Problem) Result {
    // 1. Start MCP server process
    // 2. Send tool call via stdio
    // 3. Parse response
    // 4. Evaluate against expected
    // 5. Return result
}
```

### Regression Detection

```go
func detectRegression(current, baseline float64, threshold float64) bool {
    degradation := (baseline - current) / baseline
    return degradation > threshold // e.g., 0.02 for 2%
}
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Benchmark Tests

on:
  pull_request:
  push:
    branches: [main]
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Run benchmarks
        run: make benchmark
      - name: Check regression
        run: go run benchmarks/regression_check.go
      - name: Upload results
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: benchmarks/results/
```

### Regression Thresholds

- **Critical**: >5% degradation - Block merge, immediate attention
- **Warning**: 2-5% degradation - Review required, investigation needed
- **Acceptable**: <2% degradation - Normal variance, allow merge

## Storage Schema

### Results Storage (SQLite)

```sql
CREATE TABLE benchmark_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT UNIQUE NOT NULL,
    suite_name TEXT NOT NULL,
    git_commit TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    overall_accuracy REAL NOT NULL,
    overall_ece REAL NOT NULL,
    avg_latency_ms REAL NOT NULL,
    total_problems INTEGER NOT NULL,
    metadata TEXT  -- JSON
);

CREATE TABLE benchmark_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    problem_id TEXT NOT NULL,
    correct INTEGER NOT NULL,
    score REAL NOT NULL,
    confidence REAL NOT NULL,
    latency_ms REAL NOT NULL,
    tokens INTEGER NOT NULL,
    mode TEXT NOT NULL,
    error TEXT,
    FOREIGN KEY (run_id) REFERENCES benchmark_runs(run_id)
);

CREATE INDEX idx_runs_timestamp ON benchmark_runs(timestamp DESC);
CREATE INDEX idx_runs_commit ON benchmark_runs(git_commit);
CREATE INDEX idx_results_run ON benchmark_results(run_id);
```

## Success Criteria

### Phase 1 Success
- Benchmark command executes successfully
- At least 50 logic problems tested
- Accuracy metric computed correctly
- Results saved to file

### Phase 2 Success
- 3 complete test suites running (105 total problems)
- Confidence calibration (ECE) calculated
- Performance metrics tracked
- Report shows all metrics

### Phase 3 Success
- Learning tests demonstrate improvement over iterations
- Statistical significance validated
- Trend visualization working
- Can measure 10%+ improvements

### Phase 4 Success
- GitHub Actions workflow runs automatically
- Regression detection blocks PRs correctly
- Historical data accumulates
- Reports are actionable

## Risk Mitigation

### Technical Risks

**Risk**: Benchmark datasets don't reflect real usage
**Mitigation**: Start with standard benchmarks, add custom datasets from real trajectories

**Risk**: MCP server communication overhead skews latency measurements
**Mitigation**: Measure server-internal timing separately from E2E latency

**Risk**: False positives in regression detection from variance
**Mitigation**: Use statistical tests (t-tests) and run multiple iterations

### Process Risks

**Risk**: Benchmarks become stale and unmaintained
**Mitigation**: Automated CI/CD runs, weekly reviews, clear ownership

**Risk**: Too many benchmarks slow down CI/CD
**Mitigation**: Tiered approach - fast tests on PR, full suite on merge/weekly

## Resource Requirements

**Development Time**: 4 weeks (1 developer)

**Infrastructure**:
- GitHub Actions minutes (included in free tier)
- Storage for results (minimal - SQLite file <10MB)
- No additional compute or cloud resources

**Third-Party Dependencies**:
- None required (Go stdlib + existing MCP SDK)
- Optional: Visualization tools if needed

## Next Steps

1. **Create benchmarks directory structure**
2. **Implement Phase 1 (Week 1)** - Get first benchmark running
3. **Review results and iterate** - Adjust based on findings
4. **Continue to Phases 2-4** - Build out complete framework

## Appendix: Example Benchmark

```go
// benchmarks/reasoning_test.go
package benchmarks

import (
    "testing"
)

func TestLogicReasoning(t *testing.T) {
    suite := LoadSuite("datasets/reasoning/logic_puzzles.json")

    results := make([]Result, 0, len(suite.Problems))
    for _, problem := range suite.Problems {
        result := executeProblem(problem)
        results = append(results, result)
    }

    accuracy := computeAccuracy(results)
    avgLatency := computeAvgLatency(results)

    t.Logf("Accuracy: %.2f%%", accuracy*100)
    t.Logf("Avg Latency: %v", avgLatency)

    // Store results for trend tracking
    storeResults("logic_reasoning", results)

    // Check regression
    baseline := loadBaseline("logic_reasoning")
    if accuracy < baseline-0.02 {
        t.Errorf("Regression detected: %.2f%% vs baseline %.2f%%",
            accuracy*100, baseline*100)
    }
}
```

## References

- Big-Bench Hard: https://github.com/suzgunmirac/BIG-Bench-Hard
- GSM8K: https://github.com/openai/grade-school-math
- HELM: https://crfm.stanford.edu/helm/
- OpenAI Evals: https://github.com/openai/evals
