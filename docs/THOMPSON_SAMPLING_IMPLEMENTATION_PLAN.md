# Thompson Sampling Reinforcement Learning - Implementation Plan

## Executive Summary

Implement Thompson Sampling bandit algorithm to enable adaptive reasoning mode selection, improving from 23% baseline accuracy to 30-35% through continuous learning from benchmark outcomes.

**Timeline**: 2 weeks
**Complexity**: Low-Medium
**Expected Impact**: 15-35% accuracy improvement
**Risk**: Low (easy rollback, <1ms overhead)

## Current State Analysis

### What We Have
- ✓ Benchmark framework with 114 problems
- ✓ Baseline metrics: 23% accuracy (26/105 correct)
- ✓ Per-mode performance data available
- ✓ SQLite storage infrastructure
- ✓ Existing auto mode selection in `modes/auto.go`
- ✓ Result storage and retrieval

### What We're Missing
- Thompson Sampling algorithm implementation
- Strategy outcome tracking
- Adaptive mode selection based on learned patterns
- Performance monitoring dashboard

### Current Mode Performance (from benchmarks)
- **Linear Mode**: Used 100% of time (simplistic DirectExecutor)
- **Actual Capabilities**: 6 modes (linear, tree, divergent, reflection, backtracking, auto)
- **Opportunity**: Learn which mode works best for which problem type

## Architecture Design

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Thompson Sampling RL                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Strategy   │    │   Outcome    │    │  Thompson    │  │
│  │   Registry   │───▶│   Tracker    │───▶│   Selector   │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         │                    │                    │          │
│         ▼                    ▼                    ▼          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              SQLite Storage                           │  │
│  │  • strategies         (mode definitions)              │  │
│  │  • strategy_outcomes  (execution results)             │  │
│  │  • thompson_state     (alpha/beta parameters)         │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │    Auto Mode           │
              │  (modes/auto.go)       │
              │  • Queries Thompson    │
              │  • Selects best mode   │
              │  • Records outcome     │
              └────────────────────────┘
```

### Data Flow

```
Problem Input
    │
    ▼
Thompson Selector
    │ (samples from Beta distributions)
    ▼
Selected Mode (linear/tree/divergent)
    │
    ▼
Execute Reasoning
    │
    ▼
Outcome (success/failure)
    │
    ▼
Update Thompson State
    │ (Bayesian update: α += success, β += failure)
    ▼
Improved Future Selections
```

## Database Schema

### Schema v7 - Thompson Sampling Tables

```sql
-- Strategies table: Available reasoning strategies
CREATE TABLE IF NOT EXISTS rl_strategies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    mode TEXT NOT NULL,                -- linear, tree, divergent, reflection
    parameters TEXT,                    -- JSON: strategy-specific params
    created_at INTEGER NOT NULL,
    is_active INTEGER DEFAULT 1
);

-- Strategy outcomes: Historical performance data
CREATE TABLE IF NOT EXISTS rl_strategy_outcomes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    strategy_id TEXT NOT NULL,
    problem_id TEXT NOT NULL,
    problem_type TEXT,                  -- logical, probabilistic, causal, etc.
    problem_description TEXT,
    success INTEGER NOT NULL,           -- 1 for success, 0 for failure
    confidence_before REAL,
    confidence_after REAL,
    execution_time_ns INTEGER,
    token_count INTEGER,
    reasoning_path TEXT,                -- JSON: full reasoning trace
    timestamp INTEGER NOT NULL,
    metadata TEXT,                      -- JSON: additional context
    FOREIGN KEY (strategy_id) REFERENCES rl_strategies(id)
);

-- Thompson state: Current Beta distribution parameters
CREATE TABLE IF NOT EXISTS rl_thompson_state (
    strategy_id TEXT PRIMARY KEY,
    alpha REAL NOT NULL DEFAULT 1.0,   -- Successes + 1
    beta REAL NOT NULL DEFAULT 1.0,    -- Failures + 1
    total_trials INTEGER DEFAULT 0,
    total_successes INTEGER DEFAULT 0,
    last_updated INTEGER NOT NULL,
    FOREIGN KEY (strategy_id) REFERENCES rl_strategies(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_outcomes_strategy ON rl_strategy_outcomes(strategy_id);
CREATE INDEX IF NOT EXISTS idx_outcomes_type ON rl_strategy_outcomes(problem_type);
CREATE INDEX IF NOT EXISTS idx_outcomes_timestamp ON rl_strategy_outcomes(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_outcomes_success ON rl_strategy_outcomes(success);

-- View for strategy performance
CREATE VIEW IF NOT EXISTS rl_strategy_performance AS
SELECT
    s.id,
    s.name,
    s.mode,
    COALESCE(ts.total_trials, 0) as trials,
    COALESCE(ts.total_successes, 0) as successes,
    COALESCE(CAST(ts.total_successes AS REAL) / NULLIF(ts.total_trials, 0), 0) as success_rate,
    ts.alpha,
    ts.beta,
    ts.last_updated
FROM rl_strategies s
LEFT JOIN rl_thompson_state ts ON s.id = ts.strategy_id
WHERE s.is_active = 1;
```

## Implementation Phases

### Phase 1: Core Algorithm (Week 1, Days 1-3)

**Tasks**:
1. Create `internal/reinforcement/` package
2. Implement Beta distribution sampling
   - Gamma distribution sampler (Marsaglia-Tsang algorithm)
   - Beta(α,β) = Gamma(α,1) / (Gamma(α,1) + Gamma(β,1))
3. Implement Thompson Sampling selector
   - Sample from each strategy's Beta distribution
   - Select strategy with highest sample
4. Add comprehensive unit tests

**Deliverables**:
- `internal/reinforcement/beta_sampling.go`
- `internal/reinforcement/thompson.go`
- `internal/reinforcement/thompson_test.go`
- Tests validate sampling correctness and selection logic

**Validation Criteria**:
- Beta sampling produces valid probability distributions
- Thompson selector chooses probabilistically
- Performance <1ms per selection
- Tests have >90% coverage

### Phase 2: Storage Integration (Week 1, Days 4-5)

**Tasks**:
1. Add schema v7 to `internal/storage/sqlite_schema.go`
2. Implement strategy CRUD operations
3. Implement outcome tracking
4. Implement state management (α/β updates)
5. Add migration from v6 to v7

**Deliverables**:
- Updated `sqlite_schema.go` with RL tables
- `internal/storage/rl_storage.go` - RL-specific storage methods
- Migration path for existing databases

**Validation Criteria**:
- Schema creates successfully
- Can insert/query strategies, outcomes, state
- Bayesian updates persist correctly
- Migration works on existing v6 databases

### Phase 3: Auto Mode Integration (Week 2, Days 1-2)

**Tasks**:
1. Modify `modes/auto.go` to query Thompson selector
2. Wire RL selector into mode selection logic
3. Record outcomes after reasoning execution
4. Add configuration flags (enable/disable RL)

**Deliverables**:
- Updated `modes/auto.go` with RL integration
- Configuration via environment variables
- Outcome recording hooks

**Validation Criteria**:
- Auto mode queries Thompson selector
- Outcomes recorded to database
- Can enable/disable RL via config
- Fallback to original logic if RL disabled

### Phase 4: Benchmark Integration (Week 2, Days 3-4)

**Tasks**:
1. Connect benchmark framework to RL outcome tracking
2. Run benchmarks with RL enabled
3. Compare performance: random vs Thompson
4. Add RL metrics to benchmark reports

**Deliverables**:
- Benchmark executor records RL outcomes
- Performance comparison suite
- RL metrics in efficiency reports

**Validation Criteria**:
- Benchmarks populate RL tables
- Can demonstrate learning over iterations
- Reports show strategy selection distribution
- Improvement curve visible

### Phase 5: Monitoring & Optimization (Week 2, Day 5)

**Tasks**:
1. Create performance dashboard queries
2. Add strategy analysis tools
3. Implement exploration/exploitation balance monitoring
4. Performance optimization if needed

**Deliverables**:
- SQL queries for strategy performance analysis
- Dashboard views for monitoring
- Documentation for operators

**Validation Criteria**:
- Can visualize strategy performance
- Can detect convergence to best strategies
- Can identify problem-type affinities

## Technical Specifications

### Beta Distribution Sampling

```go
package reinforcement

import (
    "math"
    "math/rand"
)

// SampleBeta samples from Beta(α, β) distribution
func SampleBeta(alpha, beta float64, rng *rand.Rand) float64 {
    x := SampleGamma(alpha, 1, rng)
    y := SampleGamma(beta, 1, rng)
    return x / (x + y)
}

// SampleGamma samples from Gamma(α, β) using Marsaglia-Tsang method
func SampleGamma(alpha, beta float64, rng *rand.Rand) float64 {
    // Marsaglia and Tsang's Method
    if alpha >= 1 {
        d := alpha - 1.0/3.0
        c := 1.0 / math.Sqrt(9.0*d)

        for {
            x := rng.NormFloat64()
            v := math.Pow(1.0+c*x, 3)
            if v <= 0 {
                continue
            }

            u := rng.Float64()
            if u < 1-0.0331*math.Pow(x, 4) {
                return d * v / beta
            }

            if math.Log(u) < 0.5*x*x+d*(1-v+math.Log(v)) {
                return d * v / beta
            }
        }
    } else {
        // For α < 1, use transformation
        return SampleGamma(alpha+1, beta, rng) * math.Pow(rng.Float64(), 1.0/alpha)
    }
}
```

### Thompson Selector Interface

```go
package reinforcement

// ThompsonSelector implements Thompson Sampling strategy selection
type ThompsonSelector struct {
    storage  StrategyStorage
    rng      *rand.Rand
}

// Strategy represents a reasoning strategy
type Strategy struct {
    ID          string
    Name        string
    Mode        string
    Parameters  map[string]interface{}
    Alpha       float64  // Successes + 1
    Beta        float64  // Failures + 1
}

// SelectStrategy samples and selects best strategy
func (ts *ThompsonSelector) SelectStrategy(problemContext ProblemContext) (*Strategy, error) {
    // Get all active strategies
    strategies, err := ts.storage.GetActiveStrategies()
    if err != nil {
        return nil, err
    }

    // Sample from each Beta distribution
    bestStrategy := strategies[0]
    maxSample := 0.0

    for _, strategy := range strategies {
        sample := SampleBeta(strategy.Alpha, strategy.Beta, ts.rng)
        if sample > maxSample {
            maxSample = sample
            bestStrategy = strategy
        }
    }

    return bestStrategy, nil
}

// RecordOutcome updates strategy state based on outcome
func (ts *ThompsonSelector) RecordOutcome(strategyID string, success bool) error {
    if success {
        return ts.storage.IncrementAlpha(strategyID)
    }
    return ts.storage.IncrementBeta(strategyID)
}
```

### Integration Points

**1. Auto Mode Enhancement** (`modes/auto.go`):
```go
func (m *AutoMode) ProcessThought(thought *types.Thought, store storage.Storage) (*types.Thought, error) {
    // Check if RL is enabled
    if m.rlEnabled {
        // Query Thompson selector
        strategy, err := m.thompsonSelector.SelectStrategy(ProblemContext{
            Content: thought.Content,
            Type: detectProblemType(thought.Content),
        })
        if err == nil {
            // Use selected mode
            thought.Mode = types.ThinkingMode(strategy.Mode)
        }
    } else {
        // Fall back to original semantic/keyword detection
        thought.Mode = m.detectMode(thought)
    }

    // Execute with selected mode...
    result, err := m.executeMode(thought, store)

    // Record outcome if RL enabled
    if m.rlEnabled && result != nil {
        success := result.Confidence > 0.7  // Example threshold
        _ = m.thompsonSelector.RecordOutcome(strategy.ID, success)
    }

    return result, err
}
```

**2. Benchmark Integration** (`benchmarks/suite.go`):
```go
func RunSuiteWithRL(suite *BenchmarkSuite, selector *ThompsonSelector, evaluator Evaluator, executor ProblemExecutor) (*BenchmarkRun, error) {
    for _, problem := range suite.Problems {
        // Let Thompson select strategy
        strategy, _ := selector.SelectStrategy(ProblemContext{
            Description: problem.Description,
            Type: problem.Category,
        })

        // Execute with selected strategy
        result, _ := executor.Execute(problem, evaluator)

        // Record outcome for learning
        selector.RecordOutcome(strategy.ID, result.Correct)
    }
}
```

## Configuration

### Environment Variables

```bash
# Thompson Sampling Configuration
RL_ENABLED=true                    # Enable/disable RL
RL_MIN_SAMPLES=10                  # Minimum samples before exploitation
RL_EXPLORATION_BONUS=0.1           # UCB exploration bonus (optional)
RL_OUTCOME_THRESHOLD=0.7           # Confidence threshold for success
```

### Initial Strategies

Populate on first run:

```sql
INSERT INTO rl_strategies (id, name, description, mode, parameters) VALUES
('strategy_linear', 'Linear Sequential', 'Step-by-step systematic reasoning', 'linear', '{}'),
('strategy_tree', 'Tree Exploration', 'Multi-branch parallel exploration', 'tree', '{}'),
('strategy_divergent', 'Divergent Creative', 'Creative unconventional thinking', 'divergent', '{"force_rebellion": false}'),
('strategy_reflection', 'Reflective Analysis', 'Metacognitive reflection', 'reflection', '{}'),
('strategy_backtracking', 'Checkpoint Backtracking', 'Iterative refinement with rollback', 'backtracking', '{}');

-- Initialize Thompson state (uniform prior: α=1, β=1)
INSERT INTO rl_thompson_state (strategy_id, alpha, beta, total_trials, total_successes, last_updated) VALUES
('strategy_linear', 1.0, 1.0, 0, 0, strftime('%s', 'now')),
('strategy_tree', 1.0, 1.0, 0, 0, strftime('%s', 'now')),
('strategy_divergent', 1.0, 1.0, 0, 0, strftime('%s', 'now')),
('strategy_reflection', 1.0, 1.0, 0, 0, strftime('%s', 'now')),
('strategy_backtracking', 1.0, 1.0, 0, 0, strftime('%s', 'now'));
```

## Implementation Details

### Directory Structure

```
internal/
├── reinforcement/
│   ├── beta_sampling.go        # Beta/Gamma distribution sampling
│   ├── beta_sampling_test.go   # Statistical tests
│   ├── thompson.go             # Thompson Sampling selector
│   ├── thompson_test.go        # Selection logic tests
│   ├── types.go                # Strategy, Outcome types
│   └── storage.go              # StrategyStorage interface

storage/
├── rl_storage.go               # RL-specific SQLite methods
└── rl_storage_test.go          # Storage tests
```

### Key Algorithms

**Marsaglia-Tsang Gamma Sampling**:
- Efficient for α ≥ 1 (most cases)
- Falls back to transformation for α < 1
- Average iterations: 1.3 per sample
- Performance: ~100ns per sample

**Thompson Selection**:
- Sample from each strategy's Beta(α, β)
- Select argmax(samples)
- O(n) where n = number of strategies
- Performance: <1ms for 10 strategies

**Bayesian Update**:
- Success: α ← α + 1
- Failure: β ← β + 1
- O(1) update time
- Triggers automatically via SQL

## Testing Strategy

### Unit Tests

**Beta Sampling** (`beta_sampling_test.go`):
- Test mean/variance match theoretical values
- Test edge cases (α=1, β=1, α>>β, α<<β)
- Statistical tests with χ² goodness-of-fit
- Performance benchmarks

**Thompson Selector** (`thompson_test.go`):
- Test selection probability distribution
- Test exploration vs exploitation balance
- Test outcome recording
- Test multi-strategy scenarios

### Integration Tests

**Auto Mode Integration**:
- Test RL-enabled vs disabled modes
- Test fallback on RL failure
- Test outcome recording pipeline

**Benchmark Integration**:
- Run 100 iterations on small dataset
- Verify learning curve (accuracy increases)
- Measure convergence time
- Compare to random selection baseline

### Performance Tests

Benchmarks must validate:
- Selection latency <1ms
- Update latency <10ms (including DB write)
- Memory overhead <10MB
- No degradation to existing functionality

## Success Metrics

### Phase 1 Success
- ✓ Beta sampling produces valid distributions
- ✓ Statistical tests pass (mean/variance correct)
- ✓ Thompson selector chooses probabilistically
- ✓ All unit tests pass with >90% coverage

### Phase 2 Success
- ✓ Schema creates successfully
- ✓ Can store strategies, outcomes, state
- ✓ Bayesian updates persist correctly
- ✓ Migration from v6 to v7 works

### Phase 3 Success
- ✓ Auto mode uses Thompson selector when enabled
- ✓ Outcomes recorded after execution
- ✓ Can enable/disable via environment variable
- ✓ Graceful fallback if RL fails

### Phase 4 Success
- ✓ Benchmarks demonstrate learning (accuracy improves)
- ✓ Strategy distribution shifts from uniform to best
- ✓ 10%+ improvement after 50 iterations
- ✓ Reports show RL metrics

### Phase 5 Success
- ✓ Can query strategy performance
- ✓ Can visualize learning curves
- ✓ Can monitor exploration/exploitation ratio
- ✓ Documentation complete

## Performance Targets

### Latency
- Strategy selection: <1ms (median)
- Outcome recording: <10ms (95th percentile)
- Database overhead: <5% of total execution time

### Accuracy Improvement
- Baseline: 23% (current)
- Target after 50 iterations: 30-35%
- Target after 200 iterations: 35-40%
- Long-term ceiling: 45-50% (limited by problem difficulty)

### Learning Efficiency
- Initial exploration: 30-50 trials
- Convergence: 100-150 trials
- Strategy diversity: >20% exploration rate maintained

## Risk Mitigation

### Technical Risks

**Risk**: Sampling produces invalid probabilities
**Mitigation**: Comprehensive statistical tests, clamp to [0,1]

**Risk**: Database writes slow down reasoning
**Mitigation**: Async outcome recording, batch writes, prepared statements

**Risk**: Overfitting to benchmark dataset
**Mitigation**: Track performance on held-out validation set

### Operational Risks

**Risk**: RL degrades performance initially (exploration phase)
**Mitigation**: Feature flag to disable, warm-start with reasonable priors

**Risk**: State corruption leads to poor selection
**Mitigation**: State validation, reset mechanism, backups

## Monitoring & Observability

### Key Metrics to Track

**Strategy Performance**:
- Success rate per strategy
- Trials per strategy
- α/β parameters over time

**Learning Progress**:
- Accuracy over time (moving average)
- Strategy diversity (entropy)
- Convergence detection

**System Health**:
- Selection latency (p50, p95, p99)
- Outcome recording latency
- Database size growth

### Dashboard Queries

```sql
-- Strategy performance over time
SELECT
    s.name,
    DATE(o.timestamp, 'unixepoch') as date,
    COUNT(*) as trials,
    SUM(o.success) as successes,
    CAST(SUM(o.success) AS REAL) / COUNT(*) as success_rate
FROM rl_strategy_outcomes o
JOIN rl_strategies s ON o.strategy_id = s.id
GROUP BY s.name, DATE(o.timestamp, 'unixepoch')
ORDER BY date DESC, s.name;

-- Strategy performance by problem type
SELECT
    s.name,
    o.problem_type,
    COUNT(*) as trials,
    CAST(SUM(o.success) AS REAL) / COUNT(*) as success_rate
FROM rl_strategy_outcomes o
JOIN rl_strategies s ON o.strategy_id = s.id
GROUP BY s.name, o.problem_type
HAVING trials >= 5
ORDER BY o.problem_type, success_rate DESC;
```

## Dependencies

**New**:
- None (uses Go stdlib + existing SQLite)

**Existing**:
- `modernc.org/sqlite` - Already in use
- `math/rand` - Stdlib
- `math` - Stdlib

## Rollback Plan

If RL underperforms:
1. Set `RL_ENABLED=false` to disable
2. Auto mode falls back to semantic/keyword detection
3. RL tables remain for future analysis
4. No data loss, immediate rollback

## Future Enhancements

**Phase 6** (Post-MVP):
- Contextual bandits (incorporate problem features)
- Multi-armed bandit variants (UCB, Exp3)
- Strategy parameter optimization (not just mode selection)
- Transfer learning across problem domains

## Appendix: Expected Results

### Learning Curve Projection

```
Iterations | Expected Accuracy | Strategy Distribution
-----------|-------------------|---------------------
0-20       | 23-25%           | Uniform (exploring)
20-50      | 25-30%           | Shifting to best
50-100     | 30-35%           | Mostly best (80%)
100-200    | 35-40%           | Converged (90%)
200+       | 40-45%           | Optimal (95%)
```

### ROI Analysis

**Investment**: 2 weeks development time

**Returns**:
- 15-35% accuracy improvement (measurable)
- Continuous learning capability
- Problem-type affinity discovery
- Foundation for advanced RL

**Payback**: Immediate (first 100 benchmark runs show improvement)

## Next Steps

1. **Review and approve this plan**
2. **Create feature branch** (`feature/thompson-sampling-rl`)
3. **Begin Phase 1** (Beta sampling implementation)
4. **Iterate with validation** after each phase
5. **Merge to main** after Phase 4 success validation

Ready to begin implementation?
