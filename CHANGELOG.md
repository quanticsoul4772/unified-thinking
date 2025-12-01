# CHANGELOG

Detailed change history for the Unified Thinking Server.

---

## Knowledge Graph Integration

**16 commits, 6,744 LOC**

Neo4j + chromem-go integration for semantic memory.

**Features**:
- Automatic entity extraction from reasoning sessions
- Hybrid search: semantic similarity (Voyage AI) + graph traversal (Neo4j)
- Persistent vector storage (chromem-go persists to disk)
- 3 MCP tools: `store-entity`, `search-knowledge-graph`, `create-relationship`
- 7 entity types: Concept, Person, Tool, File, Decision, Strategy, Problem
- 7 relationship types: CAUSES, ENABLES, CONTRADICTS, BUILDS_UPON, RELATES_TO, HAS_OBSERVATION, USED_IN_CONTEXT
- Regex-based extraction (10 patterns) with LLM integration ready
- Integration with Thompson Sampling RL and episodic memory

**Test Suite**: 21 integration tests (100% pass with Neo4j)

---

## Graph-of-Thoughts Implementation

**Commits**: d3db19c, 23bd01d, 173c1a2, 8f0a9ad, fde0129

Full Graph-of-Thoughts reasoning mode with LLM-powered operations.

**Files Created** (9 files, 2,029 LOC):
- `internal/modes/graph_types.go` - ThoughtVertex, ThoughtEdge, GraphState, GraphConfig
- `internal/modes/graph.go` - GraphController with state management (282 lines)
- `internal/modes/graph_operations.go` - Generate, Aggregate, Refine, Score, Prune (278 lines)
- `internal/modes/llm_client.go` - LLM interface for extensibility (61 lines)
- `internal/modes/llm_anthropic.go` - Anthropic Claude API client (319 lines)
- `internal/modes/graph_test.go` - 11 controller unit tests (241 lines)
- `internal/modes/graph_operations_test.go` - 14 operation tests (294 lines)
- `internal/server/handlers/got.go` - 8 MCP tool handlers (346 lines)
- `internal/similarity/thought_search.go` - Semantic thought search (114 lines)

**Key Features**:
- Multiple parents per vertex (key GoT advantage over tree-of-thoughts)
- LLM-powered operations via Anthropic Claude Sonnet 4.5
- Generate: k diverse continuations from active vertices
- Aggregate: Synthesize parallel paths into unified insights
- Refine: Iterative self-critique (max 3 iterations)
- Score: Multi-criteria evaluation (confidence 25%, validity 30%, relevance 25%, novelty 10%, depth 10%)
- Prune: Remove low-quality vertices while preserving roots/terminals
- Cyclic reasoning: Feedback loops from conclusions to premises

**Research**: Based on "Graph of Thoughts" paper. 61-69% error reduction vs tree-of-thoughts, 31% token cost reduction.

---

## Expert Panel Recommendations

**Commits**: 204cbb3, 839d262, 35cc95a

Improvements following review by Karl Wiegers, Martin Fowler, Michael Nygard, Lisa Crispin, and Gregor Hohpe.

### Phase 1: Observability and Documentation (204cbb3)
- Enhanced `UpdateBeliefFull` with 50+ line documentation block
- Medical test example demonstrating base rate fallacy (16.7% vs 99% intuition)
- Created `internal/metrics/probabilistic.go` with atomic metrics
- WARNING logging for uninformative evidence
- Extended `get-metrics` with probabilistic metrics

### Phase 2: Likelihood Estimator (839d262)
- `LikelihoodEstimator` interface in `internal/reasoning/likelihood_estimator.go`
- Three calibrated profiles: Default, Scientific, Anecdotal
- 5 example tests, 8 benchmarks (UpdateBeliefFull: 123ns/op)

### Phase 3: Property Tests (35cc95a)
- 7 property-based tests (1000 random cases each)
- Mathematical invariants verified: P(H|E) ∈ [0,1], P(A∧B) ≤ min(P(A),P(B))
- Architecture documentation in server.go

---

## Test Coverage Initiative

**Commits**: da6f163, Phase 3

### Quick Wins (da6f163)
- **Metrics**: 37.9% → 100.0% (+62.1%)
- **cmd/server**: 15.6% → 70.6% (+55.0%)
- Extracted `cmd/server/initializer.go` for testability
- Created mock embedder for API-free testing

### Edge Cases
- `internal/storage/edge_case_test.go` with 11 tests
- Large content (1MB), Unicode, concurrent writes
- Empty content, nil metadata, pagination

**Overall**: 83.7% → 84.3%

---

## Trajectory Persistence (0bdfc49)

Full SQLite persistence for episodic memory trajectories.

**Problem**: Data loss on Claude Desktop restart

**Solution**:
- Added `trajectories` table to schema (v6)
- JSON serialization to avoid import cycles
- `SetStorageBackend()` loads trajectories on init

**Files**: sqlite_schema.go, sqlite.go, memory.go, episodic.go, server.go

---

## Episodic Memory Fixes (e7cc5c7)

### Issue 1: SQLite Foreign Key Failure
- **Cause**: Duplicate `thoughts` table before `branches`
- **Fix**: Correct table ordering in schema

### Issue 2: Nil Array Validation
- **Affected**: get-recommendations, search-trajectories, start-reasoning-session
- **Fix**: Initialize all array fields to prevent null JSON

---

## Context Bridge Improvements

### Metrics Exposure
- Added `ContextBridge` to `MetricsResponse`
- Shows p50/p95/p99 latency, cache stats

### Always-Visible Response
- `context_bridge` field always present with status

### Rate Limiting
- Token bucket: 30 req/s, burst 10
- Prevents Voyage AI throttling

### Backfill Utility
- `BackfillRunner` for adding embeddings to existing trajectories

---

## Decision Storage

In-memory storage enabling retrieval and re-evaluation.

**Methods**:
- `GetDecision(decisionID)`
- `ListDecisions()`
- `RecalculateDecision(decisionID, scoreAdjustments)`
- `DeleteDecision(decisionID)`

---

## Code Quality

### Refactoring (5807e7b, e2a241d)
- Attempted server.go modularization (2,225 lines)
- Discovered Go type limitation: mcp.AddTool requires compile-time types

### Test Improvements (7ebdf3b, d2a9280)
- Memory: 35.3% → 67.8%
- Handlers: 43.0% → 47.2%

### Formatting
- gofmt -s applied (23 files)
- golangci-lint v2.x compatibility
