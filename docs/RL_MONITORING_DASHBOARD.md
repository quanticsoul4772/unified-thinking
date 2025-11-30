# Thompson Sampling RL Monitoring Dashboard

SQL queries and analysis tools for monitoring Thompson Sampling reinforcement learning performance.

## Strategy Performance Queries

### Overall Strategy Performance

View current performance for all active strategies:

```sql
SELECT
    id,
    name,
    mode,
    trials,
    successes,
    success_rate,
    alpha,
    beta
FROM rl_strategy_performance
ORDER BY success_rate DESC;
```

**Use**: Quick overview of which strategies are performing best

### Strategy Performance Over Time

Track how strategy performance evolves:

```sql
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
```

**Use**: Identify learning trends and convergence patterns

### Strategy Performance by Problem Type

Discover which strategies work best for different problem categories:

```sql
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

**Use**: Identify strategy-problem type affinities (e.g., tree mode better for causal problems)

## Learning Curve Queries

### Rolling Window Accuracy

Compute accuracy over sliding time windows:

```sql
WITH ordered_outcomes AS (
    SELECT
        o.id,
        o.success,
        o.timestamp,
        ROW_NUMBER() OVER (ORDER BY o.timestamp) as rn
    FROM rl_strategy_outcomes o
)
SELECT
    rn as trial_number,
    AVG(success) OVER (
        ORDER BY rn
        ROWS BETWEEN 9 PRECEDING AND CURRENT ROW
    ) as rolling_accuracy_10
FROM ordered_outcomes
WHERE rn >= 10
ORDER BY rn;
```

**Use**: Visualize learning curve with 10-trial rolling window

### Initial vs Final Performance

Compare early vs recent performance:

```sql
WITH ranked AS (
    SELECT
        o.success,
        o.timestamp,
        NTILE(5) OVER (ORDER BY o.timestamp) as quintile
    FROM rl_strategy_outcomes o
)
SELECT
    CASE
        WHEN quintile = 1 THEN 'Initial (first 20%)'
        WHEN quintile = 5 THEN 'Final (last 20%)'
        ELSE NULL
    END as period,
    COUNT(*) as trials,
    CAST(SUM(success) AS REAL) / COUNT(*) as accuracy
FROM ranked
WHERE quintile IN (1, 5)
GROUP BY quintile
ORDER BY quintile;
```

**Use**: Measure learning improvement (final accuracy - initial accuracy)

## Exploration/Exploitation Analysis

### Current Exploration Rate

Measure how often non-greedy strategies are chosen:

```sql
WITH best_strategy AS (
    SELECT id
    FROM rl_strategy_performance
    ORDER BY success_rate DESC
    LIMIT 1
),
recent_selections AS (
    SELECT
        o.strategy_id,
        COUNT(*) as selections
    FROM rl_strategy_outcomes o
    WHERE o.timestamp >= strftime('%s', 'now', '-7 days')
    GROUP BY o.strategy_id
)
SELECT
    SUM(CASE WHEN strategy_id != (SELECT id FROM best_strategy) THEN selections ELSE 0 END) as exploration_count,
    SUM(selections) as total_selections,
    CAST(SUM(CASE WHEN strategy_id != (SELECT id FROM best_strategy) THEN selections ELSE 0 END) AS REAL) /
        SUM(selections) as exploration_rate
FROM recent_selections;
```

**Use**: Monitor exploration/exploitation balance (target: 20-30% exploration)

### Strategy Selection Distribution

View strategy diversity (Shannon entropy):

```sql
WITH recent_selections AS (
    SELECT
        s.name,
        COUNT(*) as count
    FROM rl_strategy_outcomes o
    JOIN rl_strategies s ON o.strategy_id = s.id
    WHERE o.timestamp >= strftime('%s', 'now', '-7 days')
    GROUP BY s.name
),
total AS (
    SELECT SUM(count) as total_count FROM recent_selections
)
SELECT
    r.name,
    r.count,
    CAST(r.count AS REAL) / t.total_count as probability,
    ROUND(CAST(r.count AS REAL) / t.total_count * 100, 1) as percentage
FROM recent_selections r, total t
ORDER BY r.count DESC;
```

**Use**: Ensure strategies are being explored (not just greedy selection)

## Convergence Detection

### Thompson State Stability

Detect when learning has converged:

```sql
SELECT
    s.name,
    ts.alpha,
    ts.beta,
    ts.total_trials,
    CAST(ts.alpha AS REAL) / (ts.alpha + ts.beta) as expected_success_rate,
    ts.total_successes / CAST(ts.total_trials AS REAL) as empirical_success_rate,
    ABS(CAST(ts.alpha AS REAL) / (ts.alpha + ts.beta) -
        ts.total_successes / CAST(ts.total_trials AS REAL)) as convergence_gap
FROM rl_strategies s
JOIN rl_thompson_state ts ON s.id = ts.strategy_id
WHERE ts.total_trials >= 20
ORDER BY convergence_gap;
```

**Use**: Identify when strategies have converged (small convergence gap)

### Recent Improvement Trend

Check if accuracy is still improving:

```sql
WITH recent_windows AS (
    SELECT
        o.timestamp,
        o.success,
        NTILE(10) OVER (ORDER BY o.timestamp) as decile
    FROM rl_strategy_outcomes o
    WHERE o.timestamp >= strftime('%s', 'now', '-30 days')
)
SELECT
    decile,
    COUNT(*) as trials,
    CAST(SUM(success) AS REAL) / COUNT(*) as accuracy
FROM recent_windows
GROUP BY decile
ORDER BY decile;
```

**Use**: Detect if learning has plateaued (accuracy stable across deciles)

## Problem-Type Affinity Analysis

### Best Strategy by Problem Type

Identify optimal strategies for each problem category:

```sql
WITH strategy_type_performance AS (
    SELECT
        s.name,
        s.mode,
        o.problem_type,
        COUNT(*) as trials,
        CAST(SUM(o.success) AS REAL) / COUNT(*) as success_rate,
        ROW_NUMBER() OVER (
            PARTITION BY o.problem_type
            ORDER BY CAST(SUM(o.success) AS REAL) / COUNT(*) DESC
        ) as rank
    FROM rl_strategy_outcomes o
    JOIN rl_strategies s ON o.strategy_id = s.id
    GROUP BY s.name, s.mode, o.problem_type
    HAVING trials >= 3
)
SELECT
    problem_type,
    name as best_strategy,
    mode,
    trials,
    ROUND(success_rate * 100, 1) as success_rate_pct
FROM strategy_type_performance
WHERE rank = 1
ORDER BY problem_type;
```

**Use**: Discover which strategies excel at which problem types

### Problem Type Difficulty Analysis

Identify challenging problem categories:

```sql
SELECT
    problem_type,
    COUNT(*) as total_problems,
    SUM(success) as successes,
    CAST(SUM(success) AS REAL) / COUNT(*) as success_rate,
    AVG(confidence_after) as avg_confidence,
    AVG(execution_time_ns) / 1000000.0 as avg_execution_ms
FROM rl_strategy_outcomes
GROUP BY problem_type
HAVING total_problems >= 3
ORDER BY success_rate;
```

**Use**: Find problem types that need better strategies or more training

## Performance Metrics

### Execution Efficiency

Track execution time and token usage:

```sql
SELECT
    s.name,
    COUNT(*) as executions,
    ROUND(AVG(o.execution_time_ns) / 1000000.0, 2) as avg_latency_ms,
    ROUND(AVG(o.token_count), 1) as avg_tokens,
    ROUND(MIN(o.execution_time_ns) / 1000000.0, 2) as min_latency_ms,
    ROUND(MAX(o.execution_time_ns) / 1000000.0, 2) as max_latency_ms
FROM rl_strategy_outcomes o
JOIN rl_strategies s ON o.strategy_id = s.id
GROUP BY s.name
ORDER BY avg_latency_ms;
```

**Use**: Identify performance bottlenecks and efficiency differences

### Confidence Calibration

Check if confidence scores match actual success rates:

```sql
WITH confidence_buckets AS (
    SELECT
        ROUND(confidence_after * 10) * 10 as confidence_bucket,
        success
    FROM rl_strategy_outcomes
    WHERE confidence_after IS NOT NULL
)
SELECT
    confidence_bucket || '-' || (confidence_bucket + 10) || '%' as confidence_range,
    COUNT(*) as total,
    CAST(SUM(success) AS REAL) / COUNT(*) as actual_success_rate,
    ABS(CAST(SUM(success) AS REAL) / COUNT(*) - confidence_bucket / 100.0) as calibration_error
FROM confidence_buckets
GROUP BY confidence_bucket
ORDER BY confidence_bucket;
```

**Use**: Detect overconfidence or underconfidence in strategy outcomes

## Recent Activity

### Last 24 Hours Summary

Quick status check:

```sql
SELECT
    'Total Trials' as metric,
    COUNT(*) as value,
    NULL as details
FROM rl_strategy_outcomes
WHERE timestamp >= strftime('%s', 'now', '-1 day')

UNION ALL

SELECT
    'Success Rate' as metric,
    ROUND(CAST(SUM(success) AS REAL) / COUNT(*) * 100, 1) as value,
    '%' as details
FROM rl_strategy_outcomes
WHERE timestamp >= strftime('%s', 'now', '-1 day')

UNION ALL

SELECT
    'Unique Strategies Used' as metric,
    COUNT(DISTINCT strategy_id) as value,
    NULL as details
FROM rl_strategy_outcomes
WHERE timestamp >= strftime('%s', 'now', '-1 day')

UNION ALL

SELECT
    'Avg Latency (ms)' as metric,
    ROUND(AVG(execution_time_ns) / 1000000.0, 2) as value,
    'ms' as details
FROM rl_strategy_outcomes
WHERE timestamp >= strftime('%s', 'now', '-1 day');
```

**Use**: Daily status dashboard

### Most Active Strategies (Last Week)

See which strategies are being selected most frequently:

```sql
SELECT
    s.name,
    s.mode,
    COUNT(*) as selections,
    CAST(SUM(o.success) AS REAL) / COUNT(*) as success_rate,
    ts.alpha,
    ts.beta
FROM rl_strategy_outcomes o
JOIN rl_strategies s ON o.strategy_id = s.id
JOIN rl_thompson_state ts ON s.id = ts.strategy_id
WHERE o.timestamp >= strftime('%s', 'now', '-7 days')
GROUP BY s.name, s.mode, ts.alpha, ts.beta
ORDER BY selections DESC
LIMIT 5;
```

**Use**: Monitor which strategies Thompson is favoring

## Diagnostic Queries

### Failed Problems Analysis

Find problems that all strategies fail:

```sql
SELECT
    problem_id,
    problem_type,
    problem_description,
    COUNT(DISTINCT strategy_id) as strategies_tried,
    SUM(success) as successes,
    CAST(SUM(success) AS REAL) / COUNT(*) as success_rate
FROM rl_strategy_outcomes
GROUP BY problem_id, problem_type, problem_description
HAVING strategies_tried >= 2 AND successes = 0
ORDER BY strategies_tried DESC;
```

**Use**: Identify problems that need new approaches or are incorrectly specified

### Regret Analysis

Measure cumulative regret (how often suboptimal strategies chosen):

```sql
WITH best_per_type AS (
    SELECT
        problem_type,
        MAX(CAST(SUM(success) AS REAL) / COUNT(*)) as best_success_rate
    FROM rl_strategy_outcomes
    GROUP BY problem_type, strategy_id
    GROUP BY problem_type
),
actual_performance AS (
    SELECT
        o.problem_type,
        o.success,
        b.best_success_rate
    FROM rl_strategy_outcomes o
    JOIN best_per_type b ON o.problem_type = b.problem_type
)
SELECT
    problem_type,
    COUNT(*) as trials,
    SUM(success) as actual_successes,
    MAX(best_success_rate) * COUNT(*) as optimal_successes,
    MAX(best_success_rate) * COUNT(*) - SUM(success) as cumulative_regret
FROM actual_performance
GROUP BY problem_type
ORDER BY cumulative_regret DESC;
```

**Use**: Measure cost of exploration (regret = missed successes due to non-optimal choices)

## Alert Queries

### Low Exploration Alert

Trigger when exploration rate drops too low:

```sql
WITH recent_best AS (
    SELECT id FROM rl_strategy_performance
    ORDER BY success_rate DESC LIMIT 1
),
last_100 AS (
    SELECT strategy_id
    FROM rl_strategy_outcomes
    ORDER BY timestamp DESC
    LIMIT 100
)
SELECT
    COUNT(*) as non_greedy_selections,
    100 as total_selections,
    CAST(COUNT(*) AS REAL) / 100 * 100 as exploration_rate_pct
FROM last_100
WHERE strategy_id != (SELECT id FROM recent_best);
```

**Alert**: If exploration rate < 15%, increase exploration bonus

### Performance Degradation Alert

Detect if recent performance is declining:

```sql
WITH time_periods AS (
    SELECT
        o.success,
        CASE
            WHEN o.timestamp >= strftime('%s', 'now', '-7 days') THEN 'last_week'
            WHEN o.timestamp >= strftime('%s', 'now', '-14 days') THEN 'prev_week'
            ELSE 'older'
        END as period
    FROM rl_strategy_outcomes o
)
SELECT
    period,
    COUNT(*) as trials,
    CAST(SUM(success) AS REAL) / COUNT(*) as accuracy
FROM time_periods
WHERE period IN ('last_week', 'prev_week')
GROUP BY period
ORDER BY period DESC;
```

**Alert**: If last_week accuracy < prev_week accuracy - 0.05, investigate

### Stagnation Detection

Identify strategies that haven't been updated recently:

```sql
SELECT
    s.name,
    ts.total_trials,
    ts.last_updated,
    (strftime('%s', 'now') - ts.last_updated) / 86400.0 as days_since_update
FROM rl_strategies s
JOIN rl_thompson_state ts ON s.id = ts.strategy_id
WHERE ts.total_trials < 10
   OR (strftime('%s', 'now') - ts.last_updated) > 604800
ORDER BY days_since_update DESC;
```

**Alert**: Strategies not explored in 7+ days may need manual testing

## Benchmark Integration Queries

### Benchmark Run History

Track benchmark performance over time:

```sql
SELECT
    run_id,
    suite_name,
    timestamp,
    total_problems,
    correct_problems,
    ROUND(overall_accuracy * 100, 2) as accuracy_pct,
    ROUND(overall_ece, 4) as calibration_ece,
    avg_latency
FROM benchmark_runs
ORDER BY timestamp DESC
LIMIT 20;
```

**Note**: Requires benchmark results to be persisted to database

### Strategy Selection Trends in Benchmarks

Analyze which strategies benchmarks favor over time:

```sql
SELECT
    br.timestamp,
    br.suite_name,
    s.name as strategy,
    COUNT(*) as selections,
    CAST(SUM(o.success) AS REAL) / COUNT(*) as success_rate
FROM rl_strategy_outcomes o
JOIN rl_strategies s ON o.strategy_id = s.id
JOIN (
    SELECT DISTINCT timestamp, suite_name
    FROM benchmark_runs
) br ON DATE(o.timestamp, 'unixepoch') = DATE(br.timestamp)
GROUP BY br.timestamp, br.suite_name, s.name
ORDER BY br.timestamp DESC, selections DESC;
```

**Use**: See how strategy preferences evolve across benchmark runs

## Configuration Queries

### Current Thompson Parameters

View current α/β parameters:

```sql
SELECT
    s.name,
    s.mode,
    ts.alpha,
    ts.beta,
    ts.total_trials,
    ts.total_successes,
    ROUND(ts.alpha / (ts.alpha + ts.beta), 3) as expected_rate,
    ROUND(CAST(ts.total_successes AS REAL) / ts.total_trials, 3) as empirical_rate
FROM rl_strategies s
JOIN rl_thompson_state ts ON s.id = ts.strategy_id
WHERE s.is_active = 1
ORDER BY ts.total_trials DESC;
```

**Use**: Understand current Thompson beliefs about strategy quality

### Reset Strategies

Reset all strategies to uniform prior (α=1, β=1):

```sql
UPDATE rl_thompson_state
SET alpha = 1.0,
    beta = 1.0,
    total_trials = 0,
    total_successes = 0,
    last_updated = strftime('%s', 'now');
```

**Use**: Start fresh learning (caution: loses all learned knowledge)

## Export Queries

### CSV Export for Visualization

Export data for external dashboards (Grafana, Tableau, etc.):

```sql
SELECT
    o.timestamp,
    s.name as strategy,
    s.mode,
    o.problem_type,
    o.success,
    o.confidence_before,
    o.confidence_after,
    o.execution_time_ns / 1000000.0 as execution_ms,
    o.token_count
FROM rl_strategy_outcomes o
JOIN rl_strategies s ON o.strategy_id = s.id
ORDER BY o.timestamp;
```

**Format**: Save as CSV for time-series analysis in external tools

## Monitoring Best Practices

### Daily Checks
1. Run "Last 24 Hours Summary" query
2. Check exploration rate (should be 15-30%)
3. Review strategy performance trends
4. Check for performance degradation alerts

### Weekly Analysis
1. Compare current week vs previous week accuracy
2. Review problem-type affinity table
3. Check convergence status
4. Analyze failed problems list

### Monthly Review
1. Full learning curve analysis
2. Cumulative regret calculation
3. Strategy efficiency comparison
4. Consider resetting underperforming strategies

### Performance Targets

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| Exploration Rate | 20-30% | <15% or >50% |
| Learning Improvement | +10% in 50 trials | <5% improvement |
| Convergence Gap | <0.05 | >0.15 |
| Strategy Diversity | >1.0 (entropy) | <0.5 (too concentrated) |
| Success Rate Trend | Increasing | Declining 2 weeks+ |

## Integration with MCP Tools

### Query via sqlite3 CLI

```bash
sqlite3 unified-thinking.db < dashboard_query.sql
```

### Query via Go Code

```go
import "database/sql"

rows, err := db.Query(`
    SELECT id, name, trials, success_rate
    FROM rl_strategy_performance
    ORDER BY success_rate DESC
`)
```

### Future: Dedicated Dashboard Tool

Consider creating MCP tool for dashboard queries:

```go
// get-rl-dashboard tool
type DashboardRequest struct {
    View string `json:"view"` // "performance", "learning", "exploration", etc.
    Days int    `json:"days"` // Time window
}

type DashboardResponse struct {
    View string                 `json:"view"`
    Data map[string]interface{} `json:"data"`
}
```

## Visualization Recommendations

### Learning Curve Plot
- X-axis: Trial number (or time)
- Y-axis: Rolling window accuracy
- Lines: One per strategy
- Annotations: Mark when strategies converged

### Strategy Distribution Pie Chart
- Sectors: Each strategy
- Size: Selection frequency
- Color: Success rate (green=high, red=low)

### Exploration/Exploitation Timeline
- X-axis: Time
- Y-axis: Exploration rate %
- Target zone: 20-30% shaded
- Trend line: Show movement toward exploitation

### Problem-Type Heatmap
- Rows: Strategies
- Columns: Problem types
- Cell color: Success rate
- Use: Quickly identify strategy-type affinities
