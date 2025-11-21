# Test Coverage Analysis - Unified Thinking MCP Server

**Generated**: 2025-11-20
**Current Overall Coverage**: 83.7%
**Target**: 85%+ for all packages, 90%+ overall

## Executive Summary

The unified-thinking MCP server has good overall test coverage (83.7%), but exhibits significant gaps in critical infrastructure packages. This analysis identifies 8 priority packages requiring immediate attention, with specific test scenarios to achieve 90%+ coverage.

**Critical Findings**:
- **cmd/server** (15.6%): Main entry point completely untested
- **internal/metrics** (37.9%): Probabilistic metrics module has 0% coverage
- **internal/contextbridge** (70.6%): Core embedding and degraded response paths untested
- **internal/embeddings** (73.0%): Voyage AI API integration and rate limiting untested
- **internal/server/executor** (24.3%): Critical tool routing logic has major gaps

## Priority 1: Critical Infrastructure (IMMEDIATE ACTION REQUIRED)

### 1. cmd/server/main.go - 15.6% Coverage (TARGET: 90%)

**Current State**: `main()` function has 0% coverage, only `registerPredefinedWorkflows()` is tested

**Untested Critical Code**:
- Main server initialization sequence (lines 30-153)
- Storage initialization error handling
- MCP server creation and transport setup
- Embedder initialization (Voyage API integration)
- Context bridge initialization
- Server lifecycle management

**Why This Matters**:
- Entry point failures prevent entire server from starting
- Claude Desktop will fall back to degraded mode silently
- Storage initialization errors may go undetected
- API key validation issues only discovered at runtime

**Test Scenarios to Add**:

```go
// C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\cmd\server\main_test.go

// 1. Storage initialization tests
func TestStorageInitializationPaths(t *testing.T) {
    // Test memory storage (default)
    // Test SQLite storage with valid path
    // Test SQLite storage with invalid path (should fall back)
    // Test storage initialization timeout
}

// 2. Embedder initialization tests
func TestEmbedderInitialization(t *testing.T) {
    // Test with valid VOYAGE_API_KEY
    // Test with invalid API key format
    // Test with missing API key (should disable semantic features)
    // Test embedder dimension validation
}

// 3. Context bridge initialization tests
func TestContextBridgeSetup(t *testing.T) {
    // Test enabled context bridge with embedder
    // Test disabled context bridge
    // Test context bridge with storage failures
    // Test signature extractor initialization
}

// 4. Component integration tests
func TestServerComponentWiring(t *testing.T) {
    // Verify all modes initialized correctly
    // Verify validator connected to storage
    // Verify orchestrator has tool executor
    // Verify episodic memory has storage access
}

// 5. MCP server creation tests
func TestMCPServerLifecycle(t *testing.T) {
    // Test server creation with all components
    // Test tool registration completeness (63 tools)
    // Test transport creation
    // Test graceful shutdown simulation
}
```

**Expected Coverage Gain**: 15.6% → 90% (+74.4 points)

---

### 2. internal/metrics/probabilistic.go - 0% Coverage (TARGET: 95%)

**Current State**: Entire module has zero test coverage

**Untested Functions** (All 0%):
- `NewProbabilisticMetrics()`
- `RecordUpdate()`
- `RecordUninformative()`
- `RecordError()`
- `RecordBeliefCreated()`
- `RecordBeliefsCombined()`
- `GetStats()`
- `GetUninformativeRate()`
- `GetErrorRate()`

**Why This Matters**:
- No verification that atomic operations work correctly under concurrency
- Uninformative evidence detection may fail silently
- Error rate calculation could have division by zero bugs
- Stats aggregation may have race conditions

**Test Scenarios to Add**:

```go
// C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\internal\metrics\probabilistic_test.go

// 1. Basic metric recording
func TestProbabilisticMetricsRecording(t *testing.T) {
    metrics := NewProbabilisticMetrics()

    // Record various metric types
    metrics.RecordUpdate()
    metrics.RecordUninformative()
    metrics.RecordError()
    metrics.RecordBeliefCreated()
    metrics.RecordBeliefsCombined()

    // Verify stats
    stats := metrics.GetStats()
    assert.Equal(t, int64(2), stats["updates_total"])
    assert.Equal(t, int64(1), stats["updates_uninformative"])
    assert.Equal(t, int64(1), stats["updates_error"])
}

// 2. Rate calculation edge cases
func TestUninformativeRateCalculations(t *testing.T) {
    // Test zero updates (division by zero protection)
    // Test 100% uninformative rate
    // Test 0% uninformative rate
    // Test mixed scenarios
}

// 3. Concurrent access safety
func TestProbabilisticMetricsConcurrency(t *testing.T) {
    metrics := NewProbabilisticMetrics()
    var wg sync.WaitGroup

    // Spawn 100 goroutines recording metrics concurrently
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            metrics.RecordUpdate()
            metrics.RecordUninformative()
            metrics.GetUninformativeRate()
        }()
    }

    wg.Wait()

    // Verify no data races and counts are accurate
    stats := metrics.GetStats()
    assert.Equal(t, int64(200), stats["updates_total"])
}

// 4. Error rate accuracy
func TestErrorRateCalculations(t *testing.T) {
    // Test with no errors
    // Test with all errors
    // Test error rate formula: errors / (total + errors)
}
```

**Expected Coverage Gain**: 0% → 95% (+95 points)

---

### 3. internal/server/executor.go - 24.3% Coverage (TARGET: 90%)

**Current State**: `ExecuteTool()` has only 24.3% coverage - critical tool routing logic mostly untested

**Untested Code Paths**:
- Tool name normalization and routing
- Parameter extraction error handling
- Tool-specific parameter parsing for 63 different tools
- Error propagation from nested tool calls
- Context passing between tools

**Why This Matters**:
- Tool routing errors cause silent failures
- Parameter extraction bugs lead to incorrect tool execution
- Missing error handling could panic the server
- This is the central dispatch point for ALL MCP tool calls

**Test Scenarios to Add**:

```go
// C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\internal\server\executor_test.go
// (This file exists but needs expansion)

// 1. Tool routing tests
func TestExecuteToolRouting(t *testing.T) {
    tests := []struct {
        name       string
        toolName   string
        params     map[string]interface{}
        shouldSucceed bool
    }{
        {"think tool", "think", map[string]interface{}{"content": "test", "mode": "linear"}, true},
        {"invalid tool", "nonexistent", map[string]interface{}{}, false},
        {"probabilistic reasoning", "probabilistic-reasoning", map[string]interface{}{...}, true},
        {"make-decision", "make-decision", map[string]interface{}{...}, true},
        // Test all 63 tools
    }
}

// 2. Parameter extraction edge cases
func TestParameterExtraction(t *testing.T) {
    // Missing required parameters
    // Wrong type parameters (string instead of float)
    // Nested parameter structures
    // Empty parameter maps
    // Null/nil values in parameters
}

// 3. Error handling paths
func TestExecutorErrorHandling(t *testing.T) {
    // Server not initialized
    // Tool execution panics
    // Context cancellation during execution
    // Timeout scenarios
}

// 4. Context propagation
func TestExecutorContextPropagation(t *testing.T) {
    // Verify context passed to tool handlers
    // Test context cancellation propagation
    // Test deadline enforcement
}
```

**Expected Coverage Gain**: 24.3% → 90% (+65.7 points)

---

## Priority 2: High-Impact Modules (HIGH PRIORITY)

### 4. internal/contextbridge - 70.6% Coverage (TARGET: 90%)

**Uncovered Critical Functions** (0% coverage):
- `HasEmbedder()` - Embedder availability check
- `GenerateEmbedding()` - Embedding generation API
- `buildDegradedResponse()` - Timeout/error fallback path
- `extractTextContent()` - Content extraction from results
- `ConfigFromEnv()` - Environment configuration parsing
- `RecordTimeout()` - Timeout metric tracking
- `Reset()` - Metrics reset functionality

**Why This Matters**:
- Degraded response path never tested (timeout scenarios uncovered)
- Embedding generation errors may crash instead of degrading gracefully
- Environment configuration bugs only discovered in production
- Timeout scenarios are critical for performance guarantees

**Test Scenarios to Add**:

```go
// C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\internal\contextbridge\bridge_test.go
// (Extend existing test file)

// 1. Degraded response scenarios
func TestDegradedResponseHandling(t *testing.T) {
    // Timeout exceeded
    // Embedding generation failure
    // Storage query failure
    // Verify status field shows "degraded"
}

// 2. Embedding integration
func TestEmbeddingGeneration(t *testing.T) {
    // Successful embedding generation
    // API timeout handling
    // Rate limiting response
    // Invalid embedding dimensions
}

// 3. Environment configuration
func TestConfigFromEnvironment(t *testing.T) {
    tests := []struct {
        envVars map[string]string
        expectedConfig Config
    }{
        // Default config
        // Custom min_similarity
        // Custom max_matches
        // Invalid values (should use defaults)
    }
}

// 4. Timeout and metrics
func TestTimeoutMetrics(t *testing.T) {
    // Record timeout events
    // Verify timeout count increments
    // Test metrics reset
}

// 5. Content extraction
func TestExtractTextContent(t *testing.T) {
    // Extract from think result
    // Extract from decision result
    // Extract from empty result
    // Extract from malformed result
}
```

**Expected Coverage Gain**: 70.6% → 90% (+19.4 points)

---

### 5. internal/embeddings - 73.0% Coverage (TARGET: 90%)

**Uncovered Critical Functions** (0% coverage):
- `Embed()` - Single text embedding
- `EmbedBatch()` - Batch embedding API
- `embedBatchOnce()` - Retry logic
- `isRateLimitError()` - Rate limit detection
- `Dimension()`, `Model()`, `Provider()` - Metadata accessors
- `cleanup()` - Cache cleanup goroutine
- `UpdateSignatureEmbedding()` - Backfill storage operation

**Why This Matters**:
- Voyage AI API integration completely untested
- Rate limiting not validated (could hit API limits)
- Retry logic has no coverage (failures may not retry)
- Cache cleanup could have memory leaks
- Backfill operation untested (data migration risk)

**Test Scenarios to Add**:

```go
// C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\internal\embeddings\voyage_test.go
// (Extend existing test file)

// 1. API integration tests (mocked)
func TestVoyageAPIIntegration(t *testing.T) {
    // Mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request format
        // Return mock embeddings
    }))
    defer server.Close()

    // Test single embed
    // Test batch embed
    // Test API error responses
}

// 2. Rate limiting tests
func TestRateLimiting(t *testing.T) {
    // Test token bucket limiter
    // Verify 30 req/sec limit enforced
    // Test burst allowance (10 requests)
    // Test rate limit error detection
}

// 3. Retry logic tests
func TestEmbedBatchRetry(t *testing.T) {
    // Test successful retry after transient error
    // Test max retries exceeded
    // Test backoff behavior
}

// 4. Cache cleanup tests
func TestCacheCleanup(t *testing.T) {
    cache := NewEmbeddingCache(1 * time.Millisecond)
    cache.Put("key", []float32{1.0, 2.0})

    // Wait for cleanup
    time.Sleep(100 * time.Millisecond)

    // Verify expired entries removed
}

// 5. Backfill operation tests
func TestBackfillStorage(t *testing.T) {
    // Test UpdateSignatureEmbedding
    // Test concurrent updates
    // Test error handling during backfill
}
```

**Expected Coverage Gain**: 73.0% → 90% (+17 points)

---

### 6. internal/integration - 78.7% Coverage (TARGET: 90%)

**Uncovered Critical Functions** (0% coverage):
- `UpdateBeliefFromCausalGraph()` - Probabilistic-causal integration
- `calculateDecisionScoreAdjustments()` - Evidence pipeline decision updates
- `calculateScoreAdjustment()` - Score calculation logic
- `evidenceRelatesToOption()` - Evidence-option matching

**Why This Matters**:
- Evidence pipeline cannot update decisions (no coverage)
- Belief-causal integration untested (complex probability math)
- Score adjustments could have calculation bugs
- Evidence matching logic may produce false positives

**Test Scenarios to Add**:

```go
// C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\internal\integration\evidence_pipeline_test.go
// (Extend existing file)

// 1. Decision score adjustment tests
func TestDecisionScoreAdjustments(t *testing.T) {
    pipeline := NewEvidencePipeline(...)

    // Create decision with options
    decision := &types.Decision{...}

    // Create evidence supporting option 1
    evidence := &types.Evidence{Content: "supports option A", ...}

    // Process evidence
    result := pipeline.ProcessEvidence(...)

    // Verify score adjustments calculated correctly
    adjustments := result.DecisionUpdates[decisionID]
    assert.Greater(t, adjustments["option_a"], 0.0)
}

// 2. Evidence-option relationship tests
func TestEvidenceRelatesToOption(t *testing.T) {
    tests := []struct {
        evidenceContent string
        optionName      string
        shouldMatch     bool
    }{
        {"PostgreSQL is fast", "PostgreSQL", true},
        {"avoid MongoDB", "MongoDB", true},
        {"security is important", "secure-option", true},
        {"unrelated content", "some-option", false},
    }
}

// 3. Probabilistic-causal integration tests
func TestUpdateBeliefFromCausalGraph(t *testing.T) {
    // Build causal graph
    graph := buildTestCausalGraph()

    // Create belief
    belief := createTestBelief()

    // Update belief from graph
    result := integration.UpdateBeliefFromCausalGraph(graph, belief, evidence)

    // Verify posterior probability calculated correctly
    assert.InDelta(t, expectedPosterior, result.Posterior, 0.01)
}
```

**Expected Coverage Gain**: 78.7% → 90% (+11.3 points)

---

### 7. internal/storage - 78.4% Coverage (TARGET: 90%)

**Coverage Analysis**: Storage layer has decent coverage but gaps exist in:
- Error recovery paths
- SQLite connection pool edge cases
- Transaction rollback scenarios
- FTS5 search edge cases

**Test Scenarios to Add**:

```go
// C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\internal\storage\sqlite_test.go
// (Extend existing file)

// 1. Connection pool exhaustion
func TestConnectionPoolLimits(t *testing.T) {
    // Exhaust connection pool
    // Verify proper queuing
    // Test timeout behavior
}

// 2. Transaction rollback scenarios
func TestTransactionRollback(t *testing.T) {
    // Start transaction
    // Induce error mid-transaction
    // Verify rollback occurred
    // Verify no partial data written
}

// 3. FTS5 edge cases
func TestFullTextSearchEdgeCases(t *testing.T) {
    // Empty search query
    // Special characters in query
    // Very long queries
    // Unicode characters
    // Search result ranking verification
}

// 4. Corruption recovery
func TestDatabaseCorruptionRecovery(t *testing.T) {
    // Simulate corrupted database file
    // Verify fallback to memory storage
    // Test error logging
}
```

**Expected Coverage Gain**: 78.4% → 90% (+11.6 points)

---

### 8. internal/server - 79.3% Coverage (TARGET: 90%)

**Low Coverage Handlers** (< 50% coverage):
- `handleMakeDecision()` - 36.4%
- `handleDecomposeProblem()` - 36.4%
- `handleAnalyzePerspectives()` - 36.4%
- `handleBuildCausalGraph()` - 36.4%
- `handlePerformCBRCycle()` - 46.7%
- `initializeEpisodicMemory()` - 43.5%

**Uncovered Functions** (0% coverage):
- `SetContextBridge()`
- `GetContextBridge()`
- `GetToolByName()`
- `GetToolNames()`

**Why This Matters**:
- Complex decision-making handlers lack error path coverage
- Context bridge integration not verified
- Tool lookup utilities completely untested
- Episodic memory initialization has significant gaps

**Test Scenarios to Add**:

```go
// C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking\internal\server\server_coverage_test.go
// (Extend existing comprehensive test file)

// 1. Decision handler comprehensive tests
func TestHandleMakeDecisionComprehensive(t *testing.T) {
    // Valid decision request
    // Missing criteria
    // Missing options
    // Invalid weights (sum != 1.0)
    // Zero options
    // Conflicting criteria
}

// 2. Problem decomposition edge cases
func TestHandleDecomposeProblemEdgeCases(t *testing.T) {
    // Simple problem (no decomposition needed)
    // Deeply nested problem
    // Circular dependencies
    // Very long problem description
}

// 3. Context bridge integration
func TestContextBridgeIntegration(t *testing.T) {
    srv := createTestServer()
    bridge := createMockContextBridge()

    srv.SetContextBridge(bridge)
    retrieved := srv.GetContextBridge()

    assert.NotNil(t, retrieved)
}

// 4. Tool lookup utilities
func TestToolLookupUtilities(t *testing.T) {
    srv := createTestServer()

    // Test GetToolByName for all 63 tools
    tool := srv.GetToolByName("think")
    assert.NotNil(t, tool)

    // Test GetToolNames
    names := srv.GetToolNames()
    assert.Equal(t, 63, len(names))
}

// 5. Episodic memory initialization paths
func TestEpisodicMemoryInitialization(t *testing.T) {
    // With embeddings enabled
    // With embeddings disabled
    // With signature extraction enabled
    // With storage failures
}
```

**Expected Coverage Gain**: 79.3% → 90% (+10.7 points)

---

## Priority 3: Acceptable Coverage (ENHANCEMENT)

### 9. internal/processing - 83.3% Coverage (TARGET: 90%)

**Minor Gaps**: Mostly edge cases in dual-process reasoning

**Quick Wins**:
- Test System 1 → System 2 escalation edge cases
- Test complexity threshold boundaries
- Test timeout handling in deliberate processing

**Expected Gain**: +6.7 points

---

## Summary of Recommendations

### Immediate Action Items (Next Sprint)

1. **cmd/server/main_test.go** - Add 5 test suites covering main initialization (Target: +74 points)
2. **internal/metrics/probabilistic_test.go** - Add complete test coverage (Target: +95 points)
3. **internal/server/executor_test.go** - Expand tool routing tests (Target: +66 points)

**Total Expected Gain**: +235 points across 3 packages

### High Priority (Following Sprint)

4. **internal/contextbridge** - Add degraded response and embedding tests (+19 points)
5. **internal/embeddings** - Add Voyage API integration tests (+17 points)
6. **internal/integration** - Add evidence pipeline tests (+11 points)

**Total Expected Gain**: +47 points across 3 packages

### Enhancement (Ongoing)

7. **internal/storage** - Fill SQLite edge case gaps (+12 points)
8. **internal/server** - Expand handler coverage (+11 points)
9. **internal/processing** - Complete dual-process tests (+7 points)

**Total Expected Gain**: +30 points across 3 packages

---

## Test Quality Standards

When implementing these tests, ensure:

1. **Independence**: Tests don't share state or depend on execution order
2. **Coverage of Error Paths**: Test both happy path and all error scenarios
3. **Edge Cases**: Test boundary conditions, empty inputs, null values
4. **Concurrency Safety**: Test thread-safe components under concurrent load
5. **Table-Driven Design**: Use table-driven tests for multiple scenarios
6. **Clear Assertions**: Use specific assertions with descriptive failure messages
7. **Mock External Dependencies**: Mock Voyage API, SQLite errors, network failures
8. **Performance**: Keep unit tests fast (< 100ms each)

---

## Projected Coverage After Implementation

| Package | Current | Target | Gain |
|---------|---------|--------|------|
| cmd/server | 15.6% | 90% | +74.4 |
| internal/metrics | 37.9% | 95% | +57.1 |
| internal/server/executor | 24.3% | 90% | +65.7 |
| internal/contextbridge | 70.6% | 90% | +19.4 |
| internal/embeddings | 73.0% | 90% | +17.0 |
| internal/integration | 78.7% | 90% | +11.3 |
| internal/storage | 78.4% | 90% | +11.6 |
| internal/server | 79.3% | 90% | +10.7 |
| internal/processing | 83.3% | 90% | +6.7 |

**Overall Target**: 90%+ coverage (currently 83.7%)

**Estimated Implementation Effort**:
- Immediate: 16 hours (3 packages)
- High Priority: 12 hours (3 packages)
- Enhancement: 8 hours (3 packages)
- **Total**: ~36 hours of focused test development

---

## Appendix: Critical Untested Functions

### Embeddings (0% coverage)
- `voyage.Embed()` - Single embedding generation
- `voyage.EmbedBatch()` - Batch embedding with retry
- `voyage.isRateLimitError()` - Rate limit detection
- `embeddings.cleanup()` - Background cache cleanup

### Context Bridge (0% coverage)
- `bridge.buildDegradedResponse()` - Timeout fallback
- `bridge.extractTextContent()` - Result content extraction
- `bridge.GenerateEmbedding()` - Embedding API
- `ConfigFromEnv()` - Environment parsing

### Metrics (0% coverage - entire module)
- All ProbabilisticMetrics methods (9 functions)

### Integration (0% coverage)
- `probabilistic_causal.UpdateBeliefFromCausalGraph()`
- `evidence_pipeline.calculateDecisionScoreAdjustments()`
- `evidence_pipeline.calculateScoreAdjustment()`
- `evidence_pipeline.evidenceRelatesToOption()`

### Server (0% coverage)
- `server.SetContextBridge()`
- `server.GetContextBridge()`
- `tools.GetToolByName()`
- `tools.GetToolNames()`

### Main Entry Point (0% coverage)
- `main.main()` - Entire server initialization sequence

---

## Notes

- This analysis is based on coverage data from 2025-11-20
- All file paths are absolute Windows paths
- Test files should follow Go naming convention: `*_test.go`
- Use table-driven tests for multiple scenarios
- Mock external dependencies (Voyage API, SQLite errors)
- Focus on error paths and edge cases - happy paths often already tested
