# Production Profiling Plan

**Date**: 2025-11-21  
**Purpose**: Measure actual performance impact of Tier 1 optimizations and guide Tier 2/3 implementation

---

## Executive Summary

This plan establishes comprehensive profiling to:
1. Validate Tier 1 optimization gains (expected 10-50% improvements)
2. Identify remaining bottlenecks for Tier 2/3 prioritization
3. Establish production performance baselines

---

## Profiling Strategy

### Phase 1: Benchmark Existing Code (BASELINE)

Create benchmarks for all optimized hot paths:

1. **Storage Layer Benchmarks**
   - `BenchmarkSearchThoughts` - Measures copyThoughtOptimized impact
   - `BenchmarkIndexThoughtContent` - Measures tokenization optimization
   - `BenchmarkSearchThoughts_SmallResultSet` - Tests map/slice threshold
   - `BenchmarkSearchThoughts_LargeResultSet` - Tests at scale

2. **Reasoning Layer Benchmarks**
   - `BenchmarkBatchUpdateBelief` - Measures batch API improvement
   - `BenchmarkSequentialUpdates_Batch` - Compares batch vs individual
   - `BenchmarkShallowCopyMap` - Measures copy optimization

3. **Integration Benchmarks**
   - `BenchmarkEndToEnd_Think` - Full workflow performance
   - `BenchmarkEndToEnd_Search` - Search workflow performance

### Phase 2: CPU Profiling

Identify actual CPU hotspots in realistic workloads:

```bash
# Run with CPU profiling
go test -bench=. -benchtime=10s -cpuprofile=cpu.prof ./internal/...

# Analyze CPU profile
go tool pprof -http=:8080 cpu.prof

# Look for:
# - Functions consuming >5% CPU time
# - Unexpected allocation sites
# - Lock contention points
```

### Phase 3: Memory Profiling

Identify allocation patterns and memory pressure:

```bash
# Run with memory profiling
go test -bench=. -benchtime=10s -memprofile=mem.prof ./internal/...

# Analyze memory profile
go tool pprof -http=:8080 -alloc_space mem.prof

# Look for:
# - Functions allocating >1MB
# - Frequent small allocations
# - Slice growth patterns
```

### Phase 4: Mutex Profiling

Identify lock contention after mutex removal optimization:

```bash
# Run with mutex profiling
go test -bench=. -benchtime=10s -mutexprofile=mutex.prof ./internal/...

# Analyze mutex profile
go tool pprof -http=:8080 mutex.prof

# Validate:
# - SQLiteStorage lock removal effectiveness
# - No new contention introduced
```

---

## Benchmark Implementation

### Priority 1: Hot Path Benchmarks

These directly measure Tier 1 optimization impact:

```go
// internal/storage/memory_bench_test.go

func BenchmarkSearchThoughts_SmallResultSet(b *testing.B) {
    // Tests copyThoughtOptimized on small datasets
    s := NewMemoryStorage()
    // Setup: 100 thoughts
    for i := 0; i < 100; i++ {
        s.StoreThought(&types.Thought{
            Content: fmt.Sprintf("thought %d", i),
            Mode: types.ModeLinear,
        })
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = s.SearchThoughts("thought", types.ModeLinear, 10, 0)
    }
}

func BenchmarkSearchThoughts_LargeResultSet(b *testing.B) {
    // Tests at scale (1000+ thoughts)
    s := NewMemoryStorage()
    for i := 0; i < 1000; i++ {
        s.StoreThought(&types.Thought{
            Content: fmt.Sprintf("thought %d", i),
            Mode: types.ModeLinear,
        })
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = s.SearchThoughts("thought", types.ModeLinear, 100, 0)
    }
}

func BenchmarkIndexThoughtContent(b *testing.B) {
    // Measures tokenization optimization
    s := NewMemoryStorage()
    thought := &types.Thought{
        Content: "This is a sample thought with multiple words for tokenization testing purposes",
        Mode: types.ModeLinear,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        s.StoreThought(thought)
    }
}

func BenchmarkCopyThought(b *testing.B) {
    thought := &types.Thought{
        Content: "Test",
        Mode: types.ModeLinear,
        KeyPoints: []string{"point1", "point2"},
        Metadata: map[string]interface{}{
            "key1": "value1",
            "key2": 42,
        },
    }
    
    b.Run("Original", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = copyThought(thought)
        }
    })
    
    b.Run("Optimized", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = copyThoughtOptimized(thought)
        }
    })
}
```

### Priority 2: Batch API Benchmarks

```go
// internal/reasoning/probabilistic_bench_test.go

func BenchmarkBatchUpdateBelief(b *testing.B) {
    pr := NewProbabilisticReasoner()
    belief, _ := pr.CreateBelief("Test", 0.5)
    
    updates := []BeliefUpdate{
        {EvidenceID: "ev1", ProbEGivenH: 0.8, ProbEGivenNotH: 0.2},
        {EvidenceID: "ev2", ProbEGivenH: 0.9, ProbEGivenNotH: 0.1},
        {EvidenceID: "ev3", ProbEGivenH: 0.7, ProbEGivenNotH: 0.3},
        {EvidenceID: "ev4", ProbEGivenH: 0.85, ProbEGivenNotH: 0.15},
        {EvidenceID: "ev5", ProbEGivenH: 0.75, ProbEGivenNotH: 0.25},
    }
    
    b.Run("BatchAPI", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, _ = pr.BatchUpdateBeliefFull(belief.ID, updates)
        }
    })
    
    b.Run("Individual", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            for _, u := range updates {
                _, _ = pr.UpdateBeliefFull(belief.ID, u.EvidenceID, u.ProbEGivenH, u.ProbEGivenNotH)
            }
        }
    })
}
```

---

## Success Metrics

### Tier 1 Validation Targets

| Optimization | Metric | Target | Measurement |
|--------------|--------|--------|-------------|
| Deep Copy | ns/op | -50% | BenchmarkCopyThought |
| Batch API | ns/op | -40% | BenchmarkBatchUpdateBelief |
| Tokenization | ns/op | -25% | BenchmarkIndexThoughtContent |
| Search | allocs/op | -30% | BenchmarkSearchThoughts |

### Tier 2/3 Decision Criteria

Implement only if profiling shows:
- Function consumes >5% CPU time
- Allocation site >1MB total
- Lock contention >100ms total blocking time

---

## Production Simulation

### Realistic Workload Scenarios

1. **Scenario: Heavy Search Usage**
   ```go
   // Simulates Claude AI performing multiple searches
   func BenchmarkProductionScenario_Search(b *testing.B) {
       s := setupProductionDataset(10000) // 10k thoughts
       queries := []string{"problem", "solution", "analysis", "decision"}
       
       b.ResetTimer()
       for i := 0; i < b.N; i++ {
           for _, q := range queries {
               _ = s.SearchThoughts(q, "", 20, 0)
           }
       }
   }
   ```

2. **Scenario: Sequential Reasoning**
   ```go
   // Simulates multi-step reasoning with belief updates
   func BenchmarkProductionScenario_Reasoning(b *testing.B) {
       pr := NewProbabilisticReasoner()
       belief, _ := pr.CreateBelief("Hypothesis", 0.5)
       
       updates := generateRealisticEvidenceUpdates(10)
       
       b.ResetTimer()
       for i := 0; i < b.N; i++ {
           _, _ = pr.BatchUpdateBeliefFull(belief.ID, updates)
       }
   }
   ```

3. **Scenario: Memory Pressure**
   ```go
   // Tests under memory constraints
   func BenchmarkProductionScenario_MemoryPressure(b *testing.B) {
       s := NewMemoryStorage()
       
       b.Run("IndexGrowth", func(b *testing.B) {
           for i := 0; i < b.N; i++ {
               // Add thoughts until index reaches capacity
               for j := 0; j < 1000; j++ {
                   s.StoreThought(generateThought())
               }
           }
       })
   }
   ```

---

## Analysis Workflow

### Step 1: Run Baseline Benchmarks

```bash
# Capture baseline before Tier 1
git checkout 6014a6c  # Commit before Tier 1
go test -bench=. -benchmem ./internal/... > baseline.txt

# Capture after Tier 1
git checkout b8a8704  # Current commit
go test -bench=. -benchmem ./internal/... > tier1.txt

# Compare
benchstat baseline.txt tier1.txt
```

### Step 2: Profile Hot Paths

```bash
# CPU profile
go test -bench=BenchmarkSearchThoughts -cpuprofile=cpu_search.prof
go tool pprof -http=:8080 cpu_search.prof

# Memory profile
go test -bench=BenchmarkSearchThoughts -memprofile=mem_search.prof
go tool pprof -http=:8080 -alloc_space mem_search.prof
```

### Step 3: Identify Tier 2/3 Candidates

Review profiling output for:
1. **Map/Slice conversion** - Check if `SearchThoughts` shows map allocation overhead
2. **Index caching** - Check if `searchByIndex` appears in CPU profile
3. **Branch priority** - Check if branch calculations appear in hot path

### Step 4: Generate Report

Document findings:
- Actual % improvements vs targets
- New bottlenecks discovered
- Tier 2/3 priority ranking based on data
- Recommendations for further optimization

---

## Expected Outcomes

### If Tier 1 Meets Targets (>30% improvement)

- **Action**: Implement Tier 2 selectively based on profiling
- **Priority**: Focus on remaining >5% CPU consumers
- **Tier 3**: Defer until production load testing

### If Tier 1 Falls Short (<20% improvement)

- **Action**: Investigate why optimizations didn't help
- **Possible causes**:
  - Hot paths not in benchmarked code
  - Bottleneck elsewhere (I/O, network, JSON)
  - Allocation pressure from different source
- **Next steps**: Expand profiling scope

---

## Timeline

1. **Day 1**: Implement benchmarks (2-3 hours)
2. **Day 2**: Run profiling suite (1 hour)
3. **Day 3**: Analyze results and create report (2 hours)
4. **Day 4**: Implement Tier 2/3 based on data (4-6 hours)

---

## Tools & Commands Reference

```bash
# Install benchstat for comparison
go install golang.org/x/perf/cmd/benchstat@latest

# Run comprehensive benchmark suite
go test -bench=. -benchmem -benchtime=5s ./internal/... | tee bench_results.txt

# Profile specific benchmark
go test -bench=BenchmarkSearchThoughts -cpuprofile=cpu.prof -memprofile=mem.prof

# Analyze profiles
go tool pprof -http=:8080 cpu.prof
go tool pprof -http=:8080 -alloc_space mem.prof
go tool pprof -http=:8080 -inuse_space mem.prof

# Compare benchmarks
benchstat before.txt after.txt

# Check for race conditions
go test -race ./...

# Memory leak detection
go test -memprofile=mem.prof -bench=. && go tool pprof -base mem.prof

# Escape analysis
go build -gcflags='-m -m' ./... 2>&1 | grep "escapes to heap"
```

---

## Conclusion

This profiling plan provides data-driven guidance for:
- Validating Tier 1 optimization effectiveness
- Prioritizing Tier 2/3 implementation
- Identifying unexpected bottlenecks
- Establishing production performance baselines

Execute this plan before implementing Tier 2/3 to ensure optimization efforts target actual bottlenecks, not speculative improvements.
