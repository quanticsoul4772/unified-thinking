# CHANGELOG

Detailed change history for the Unified Thinking Server.

---

## LLM Client Consolidation (2025-12-04)

### Technical Debt: Unified LLM Infrastructure

**Problem**: Duplicate HTTP client implementations and API types across `llm_anthropic.go` and `llm_agentic.go` files.

**Solution**: Consolidated LLM client infrastructure using Go's embedding pattern.

**New Files Created**:
- `internal/modes/llm_types.go` (96 lines) - Unified Anthropic API types
- `internal/modes/llm_base.go` (112 lines) - Shared HTTP infrastructure

**Types Consolidated**:
- `APIRequest`, `Message`, `ContentBlock`, `Tool`, `APIResponse`, `ResponseBlock`, `Usage`
- Helper functions: `NewTextMessage()`, `NewBlockMessage()`, `TextBlock()`, `ToolUseBlock()`, `ToolResultBlock()`

**Architecture**:
```go
// Both clients now embed shared base
type AnthropicLLMClient struct {
    *AnthropicBaseClient  // Shared HTTP infrastructure
    useStructured    bool
    webSearchEnabled bool
}

type AgenticClient struct {
    *AnthropicBaseClient  // Shared HTTP infrastructure
    registry *ToolRegistry
    config   AgenticConfig
}
```

**Benefits**:
- Eliminated ~200 lines of duplicate type definitions
- Single HTTP client configuration (API URL, headers, timeout)
- Consistent error handling across all API calls
- Easier maintenance - fix once, applies everywhere

**Files Modified**:
- `internal/modes/llm_anthropic.go` - Refactored to embed `*AnthropicBaseClient`
- `internal/modes/llm_agentic.go` - Refactored to embed `*AnthropicBaseClient`
- `internal/modes/llm_agentic_test.go` - Updated to use unified types

**LLM File Structure** (7 files, well-organized):
| File | Lines | Purpose |
|------|-------|---------|
| `llm_types.go` | 96 | Unified Anthropic API types |
| `llm_base.go` | 112 | Shared HTTP infrastructure |
| `llm_anthropic.go` | 665 | AnthropicLLMClient for GoT |
| `llm_agentic.go` | 311 | AgenticClient with tool loop |
| `llm_client.go` | 55 | LLMClient interface |
| `llm_models.go` | 142 | Domain-specific model config |
| `llm_tools.go` | 216 | Structured output schemas |

---

## Phase 3: Advanced Capabilities (2025-12-03)

### Multimodal Embeddings (`embed-multimodal`)

**New Tool**: `embed-multimodal` - Generate embeddings for images using Voyage AI's multimodal model.

**Features**:
- Support for base64-encoded images (PNG, JPEG, GIF, WebP)
- Support for image URLs (auto-fetched and encoded)
- Uses `voyage-multimodal-3` model (1024 dimensions)
- Image preprocessing: resize to max 512px, optimize quality
- Batch embedding support for multiple images

**Usage**:
```json
{
  "image_base64": "iVBORw0KGgo...",
  "content_type": "image/png"
}
```

**Response**:
```json
{
  "embedding": [0.123, -0.456, ...],
  "dimensions": 1024,
  "model": "voyage-multimodal-3",
  "input_type": "base64"
}
```

**Environment Variables**:
- `MULTIMODAL_ENABLED=true` - Enable multimodal embeddings
- `MULTIMODAL_MODEL` - Model to use (default: `voyage-multimodal-3`)

**Files Created**: `internal/embeddings/multimodal.go`, `internal/embeddings/multimodal_test.go`, `internal/embeddings/image_utils.go`, `internal/embeddings/image_utils_test.go`, `internal/server/handlers/multimodal.go`, `internal/server/handlers/multimodal_test.go`

### Programmatic Tool Calling (`run-agent`, `list-agent-tools`)

**New Tools**: Agentic LLM with programmatic tool calling for complex multi-step reasoning.

**Features**:
- Autonomous task execution with tool-calling loop
- 30+ safe tools available (no side effects)
- Configurable max iterations (default: 10, max: 20)
- Execution trace with full history
- Custom system prompts supported

**Safe Tool Subset** (available by default):
- Reasoning: `think`, `decompose-problem`, `make-decision`, `dual-process-think`
- Analysis: `analyze-perspectives`, `detect-biases`, `detect-fallacies`
- Evidence: `assess-evidence`, `probabilistic-reasoning`, `detect-contradictions`
- Causal: `build-causal-graph`, `generate-hypotheses`, `evaluate-hypotheses`
- Search: `search-similar-thoughts`, `search-knowledge-graph`, `retrieve-similar-cases`
- Synthesis: `synthesize-insights`, `detect-emergent-patterns`
- Graph-of-Thoughts: `got-generate`, `got-aggregate`, `got-refine`, `got-score`

**Excluded Tools** (not available to prevent recursion/side effects):
- Storage: `store-entity`, `create-relationship`
- Sessions: `export-session`, `import-session`
- Orchestration: `run-agent`, `run-preset`, `execute-workflow`
- State: `create-checkpoint`, `restore-checkpoint`, `got-initialize`, `got-prune`, `got-finalize`

**Usage**:
```json
{
  "task": "Analyze the pros and cons of microservices architecture for a startup",
  "max_iterations": 5,
  "allowed_tools": ["think", "analyze-perspectives", "make-decision"]
}
```

**Response**:
```json
{
  "final_answer": "Based on my analysis...",
  "status": "completed",
  "iterations": 3,
  "tools_used": ["think", "analyze-perspectives", "make-decision"],
  "total_tool_calls": 7,
  "error_count": 0,
  "trace": { ... }
}
```

**Environment Variables**:
- `AGENT_ENABLED=true` - Enable agentic tool calling
- `AGENT_MODEL` - Model to use (default: `claude-sonnet-4-5-20250929`)
- `ANTHROPIC_API_KEY` - Required for agent execution

**Files Created**: `internal/modes/tool_registry.go`, `internal/modes/tool_registry_test.go`, `internal/modes/llm_agentic.go`, `internal/modes/llm_agentic_test.go`, `internal/server/handlers/agent.go`

**Tool Count**: 82 → 85 tools

---

## Phase 2: Quality Improvements (2025-12-03)

### Voyage AI Reranking

**Problem**: Embedding-based similarity search returns results based purely on vector distance, which may not reflect relevance for specific queries.

**Solution**: Integrated Voyage AI's rerank API to optimize search results using cross-encoder relevance scoring.

**Features**:
- Automatic reranking for `search-similar-thoughts` and `search-knowledge-graph` tools
- Pipeline: Query → Embedding Search (2x limit) → Rerank → Return Top Results
- Model selection: `rerank-2` (quality) or `rerank-2-lite` (speed)
- Graceful fallback to embedding-only scoring if reranking fails

**Environment Variables**:
- `RERANK_ENABLED` - Enable/disable reranking (default: `true` when VOYAGE_API_KEY set)
- `RERANK_MODEL` - Model to use: `rerank-2` (default) or `rerank-2-lite`

**Files Created**: `internal/embeddings/reranker.go`, `internal/embeddings/reranker_test.go`
**Files Modified**: `internal/similarity/thought_search.go`, `internal/knowledge/knowledge_graph.go`, `cmd/server/initializer.go`

### Domain-Specific Model Configuration

**Problem**: All tasks used the same model configuration, regardless of task type (code vs research vs quick queries).

**Solution**: Added automatic task domain detection and per-domain model configuration.

**Features**:
- Automatic domain detection via keyword analysis
- Domain-specific temperature settings (code: 0.3, research: 0.7, quick/default: 0.5)
- Configurable models per domain via environment variables
- Default to Haiku for quick tasks, Sonnet for complex tasks

**Domains**:
| Domain | Keywords | Default Model | Temperature |
|--------|----------|---------------|-------------|
| `code` | debug, function, api, error, implement, refactor | claude-sonnet-4-5-20250929 | 0.3 |
| `research` | research, analyze, study, hypothesis, experiment | claude-sonnet-4-5-20250929 | 0.7 |
| `quick` | simple, brief, summarize, what is | claude-3-5-haiku-20241022 | 0.5 |
| `default` | - | claude-sonnet-4-5-20250929 | 0.5 |

**Environment Variables**:
- `GOT_MODEL` - Default model for all domains
- `GOT_MODEL_CODE` - Override for code domain
- `GOT_MODEL_RESEARCH` - Override for research domain
- `GOT_MODEL_QUICK` - Override for quick domain

**Files Created**: `internal/modes/llm_models.go`, `internal/modes/llm_models_test.go`

---

## Phase 1: AI Integrations (2025-12-03)

### Structured Outputs for Graph-of-Thoughts

**Problem**: GoT operations used fragile JSON parsing from LLM text responses, causing parsing failures and inconsistent outputs.

**Solution**: Implemented Anthropic's tool use API with `strict: true` for guaranteed schema compliance.

**Features**:
- JSON schemas for all GoT operations (Generate, Aggregate, Refine, Score, ExtractKeyPoints, CalculateNovelty)
- Zero parsing failures - API enforces schema at generation time
- Optional fallback to text mode via `GOT_STRUCTURED_OUTPUT=false`

**Files Created**: `internal/modes/llm_tools.go` (tool definitions with JSON schemas)
**Files Modified**: `internal/modes/llm_anthropic.go` (major rewrite for structured outputs)

### Web Search Integration (`research-with-search`)

**New Tool**: `research-with-search` - Web-augmented research using Anthropic's server-side web search.

**Features**:
- Multi-turn agentic search with automatic query refinement
- Citation extraction with URLs and titles
- Confidence scoring based on source quality
- Integrated with existing LLM client architecture

**Usage**:
```json
{
  "query": "What are the latest advances in quantum computing?",
  "problem": "Research for technology report"
}
```

**Response**:
```json
{
  "findings": "...",
  "key_insights": ["...", "..."],
  "confidence": 0.85,
  "citations": [{"url": "...", "title": "..."}],
  "searches_performed": 3,
  "status": "completed"
}
```

**Environment Variables**:
- `WEB_SEARCH_ENABLED=true` - Enable web search capability
- `GOT_STRUCTURED_OUTPUT=false` - Disable structured outputs (fallback to text parsing)

**Files Created**: `internal/server/handlers/research.go`
**Files Modified**: `internal/modes/llm_client.go`, `internal/server/server.go`

**Tool Count**: 81 → 82 tools

---

## Tool Quality Improvements (2025-12-03)

### 5 Critical Pain Points Fixed

Based on real-world testing with dark matter research funding analysis, identified and fixed 5 tool deficiencies.

#### 1. `analyze-perspectives` - Stakeholder-Specific Viewpoints

**Problem**: All stakeholders returned identical template responses.

**Solution**: Added 30+ stakeholder-specific viewpoint generators with unique framing, questions, and concerns.

**Categories**: Scientist, Policymaker, Taxpayer, Investor, Consumer, Executive, Engineer, Regulator, Academic, Journalist, Environmentalist, Union, Community, Ethicist, Competitor, Patient, Parent, Entrepreneur, Activist, Historian

**Files Modified**: `internal/analysis/perspective.go`

#### 2. `generate-hypotheses` - Multiple Meaningful Hypotheses

**Problem**: Returned single hypothesis with stopword ("matter") as the common theme.

**Solution**:
- Extended stop words list (200+ words)
- Added noun phrase extraction
- Generates 3+ distinct hypothesis types (causal mechanism, common underlying factor, systemic pattern)
- Each hypothesis includes predictions for validation

**Files Modified**: `internal/reasoning/abductive.go`

#### 3. `build-causal-graph` - Enhanced Causal Pattern Detection

**Problem**: Only captured 16% of causal relationships due to limited keyword patterns.

**Solution**: Expanded causal vocabulary:
- Attraction/enablement: "attracts", "draws", "enables", "facilitates"
- Support/funding: "funds", "finances", "supports", "sustains"
- Conditional: "requires", "needs", "depends on"
- Correlation: "correlates with", "associated with", "linked to"

**Files Modified**: `internal/reasoning/causal.go`

#### 4. `think` - Confidence Parameter Honored

**Problem**: Input confidence parameter was ignored in output.

**Solution**: Response now returns `max(input_confidence, calculated_confidence)`.

**Files Modified**: `internal/server/handlers/thinking.go`

#### 5. `decompose-problem` - Entity Extraction & Template Parameterization

**Problem**: Returned generic template descriptions regardless of input.

**Solution**:
- Added `ExtractProblemEntities()` function
- Extracts: technical terms, stakeholders, constraints, key concepts
- Domain-specific parameterization (Architecture, Debugging, Research, Proof)
- Metadata includes `extracted_entities` for transparency

**Example**:
- Input: "OAuth 2.0, SAML, enterprise customers, API clients, backward compatibility"
- Before: "Clarify functional requirements..."
- After: "Clarify functional requirements for OAuth 2.0, SAML, addressing backward compatibility constraints for enterprise customers, api clients"

**Files Modified**: `internal/reasoning/domain_templates.go`

---

## Performance & Quality Improvements (2024-12-03)

### `got-explore` Performance Fix

**Problem**: `got-explore` was taking ~4 minutes per call, causing MCP client timeouts.

**Root Cause**: Sequential LLM API calls (10+ requests per exploration).

**Solution**:
- Added fast local heuristic scoring as default (`UseFastScoring: true`)
- Added parallel LLM scoring option (`ParallelScoring: true`)
- Reduced default iterations from 2 to 1
- Added `ThoroughExploreConfig()` for when quality matters more than speed

**Result**: ~5 seconds vs ~4 minutes (98% reduction)

**New Config Options** (`ExploreConfig`):
```go
UseFastScoring  bool // Use local heuristics instead of LLM (default: true)
SkipRefine      bool // Skip refinement step (default: false)
ParallelScoring bool // Parallel LLM scoring if not using fast scoring (default: true)
```

### Domain-Aware Problem Classification Fix

**Problem**: Architecture problems ("Design the microservice architecture...") incorrectly classified as "creative".

**Root Cause**: "design" and "new" keywords in problem triggered creative detection before domain detection.

**Solution**: Added technical context whitelist in `detectCreativeProblem()` that bypasses creative classification for:
- Architecture: architecture, microservice, api, interface, system, component, module, service, integration, infrastructure
- Debugging: debug, bug, error, fix, crash, trace, exception
- Research: research, study, analyze, experiment, methodology
- Proof: prove, theorem, lemma, proof, verify, formal

**Files Modified**:
- `internal/reasoning/problem_classifier.go` - Technical context detection
- `internal/reasoning/problem_classifier_test.go` - 11 new test cases

### Code Quality

- Applied `gofmt -w .` across entire codebase
- Removed development artifacts (`IMPLEMENTATION_PLAN.md`)
- All 30 packages pass tests

---

## Claude Code Optimization

**New Package**: `internal/claudecode/`

Optimizations specifically designed for Claude Code integration.

**Features**:
- **Response Formatting** (`format/`): Compact (40-60% token reduction) and minimal (80%+ reduction) response modes via `RESPONSE_FORMAT` env var or `format-response` tool
- **Session Export/Import** (`session/`): Preserve reasoning context across sessions with JSON export (optional gzip) and merge strategies (`replace`, `merge`, `append`)
- **Workflow Presets** (`presets/`): 8 built-in presets for common development tasks (code-review, debug-analysis, architecture-decision, etc.)
- **Structured Errors** (`errors/`): Tool errors include recovery guidance, related tools, and usage examples

**5 MCP Tools**: `export-session`, `import-session`, `list-presets`, `run-preset`, `format-response`

---

## Streaming Progress Notifications

**New Package**: `internal/streaming/`

Real-time progress updates for long-running tools via MCP `notifications/progress`.

**Features**:
- Rate-limited notifications (100ms default) to prevent flooding
- Step changes bypass rate limit for immediate feedback
- Backward compatible (no-op when client doesn't provide `progressToken`)
- Per-tool configuration for interval, partial data, auto-progress

**Priority Tools**:
- P0 (Essential): `execute-workflow`, `run-preset`, `got-generate`
- P1 (Important): `got-aggregate`, `think`, `perform-cbr-cycle`
- P2 (Enhancement): `synthesize-insights`, `analyze-perspectives`, `build-causal-graph`, `evaluate-hypotheses`

**Technical Debt Fixes**:
- Removed `panic()` calls from streaming code - replaced with safe error handling utilities
- Converted `log.Fatal` in server.go to proper error propagation
- Graph-of-Thoughts now fails fast if `ANTHROPIC_API_KEY` is missing

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

**Overall at time**: 83.7% → 84.3% (see README for current coverage)

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
