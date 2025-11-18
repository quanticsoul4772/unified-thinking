# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Unified Thinking Server is a Go-based MCP (Model Context Protocol) server that consolidates multiple cognitive thinking patterns into a single efficient server. It replaces 5 separate TypeScript servers (sequential-thinking, branch-thinking, unreasonable-thinking-server, mcp-logic, state-coordinator) with one unified implementation.

### Core Architecture

**Module Path**: `unified-thinking` (local module, not published to public registry)

**Entry Point**: `cmd/server/main.go`

**Key Components**:
- `internal/types/` - Core data structures (Thought, Branch, Insight, CrossRef, Validation) + Builder patterns + 50+ cognitive reasoning types
- `internal/storage/` - Pluggable storage layer with in-memory (default) and SQLite backends (Storage interface for testability)
- `internal/modes/` - Thinking mode implementations (linear, tree, divergent, reflection, backtracking, auto) + Mode registry
- `internal/processing/` - Dual-process reasoning (System 1 fast vs System 2 deliberate)
- `internal/validation/` - Logical validation, proof checking, fallacy detection, symbolic constraint solving
- `internal/reasoning/` - Probabilistic inference, decision analysis, causal reasoning, temporal reasoning, abductive reasoning, case-based reasoning
- `internal/analysis/` - Evidence assessment, contradiction detection, perspective analysis, sensitivity testing
- `internal/metacognition/` - Self-evaluation, cognitive bias detection, unknown unknowns identification
- `internal/integration/` - Cross-mode synthesis, emergent pattern detection, probabilistic-causal integration
- `internal/orchestration/` - Workflow orchestration for automated multi-tool pipelines
- `internal/memory/` - Episodic reasoning memory system with trajectory storage, pattern learning, adaptive recommendations, and semantic embeddings integration
- `internal/embeddings/` - Semantic embeddings for episodic memory using Voyage AI (voyage-3-lite model, 512 dimensions)
- `internal/contextbridge/` - Cross-session context retrieval with signature matching, LRU caching, and metrics
- `internal/server/` - MCP server implementation with 63 registered tools
- `internal/server/handlers/` - 21 specialized handler modules (thinking, branches, validation, search, enhanced, abductive, backtracking, calibration, case_based, causal, decision, dual_process, episodic, hallucination, helpers, metacognition, metadata, probabilistic, symbolic, temporal, unknown_unknowns)

**MCP SDK**: Uses `github.com/modelcontextprotocol/go-sdk` v0.8.0

### Thinking Modes

1. **Linear Mode** (`modes/linear.go`) - Sequential step-by-step reasoning for systematic problem solving
2. **Tree Mode** (`modes/tree.go`) - Multi-branch parallel exploration with insights and cross-references
3. **Divergent Mode** (`modes/divergent.go`) - Creative/unconventional ideation with "rebellion" capability
4. **Reflection Mode** (`modes/reflection.go`) - Metacognitive reflection on previous reasoning with insight extraction
5. **Backtracking Mode** (`modes/backtracking.go`) - Checkpoint-based reasoning with restore capabilities
6. **Auto Mode** (`modes/auto.go`) - Automatic mode selection based on input content analysis

The Auto Mode uses semantic embeddings (when VOYAGE_API_KEY is set) or keyword detection to select the best thinking mode:

**Semantic Detection** (requires VOYAGE_API_KEY):
- Compares input content against prototype embeddings for each mode
- Uses cosine similarity to find best match
- Returns confidence score based on semantic similarity
- Mode prototypes include 6 examples each for linear, tree, and divergent modes

**Keyword Detection** (fallback when no API key):
- Divergent triggers: "creative", "unconventional", "what if", "imagine", "challenge", "rebel"
- Tree triggers: branch_id provided, cross_refs present, key_points present, or keywords "branch", "explore", "alternative", "parallel"
- Linear: Default for systematic reasoning

**Explicit Overrides** (highest priority):
- ForceRebellion flag forces divergent mode
- BranchID, CrossRefs, or KeyPoints force tree mode

## Development Commands

### Building
```bash
# Build for Windows (default)
make build
# Output: bin/unified-thinking.exe

# Build for Linux
make linux
# Output: bin/unified-thinking

# Install dependencies
make install-deps
```

### Running

**DO NOT run the server manually!** The server is automatically started by Claude Desktop.

For development/testing only:
```bash
# Test the server (simulates stdio communication)
go run .\cmd\server\main.go

# With debug logging
set DEBUG=true && go run .\cmd\server\main.go
```

**Note**: When run manually, the server waits for MCP protocol messages on stdin. This is only useful for protocol-level debugging.

### Testing
```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run with coverage report
make test-coverage

# Run with race detector
make test-race

# Run short tests (skip long-running)
make test-short

# Test specific package
go test -v ./internal/modes/

# Test storage layer only
make test-storage

# Run benchmarks
make benchmark

# Comprehensive test suite
make test-all
```

**Test Coverage**: The project maintains high test coverage (85.7% overall) across all major components. Key packages have excellent coverage: types (100%), metrics (100%), config (97.3%), reasoning (94.8%), memory (91.0%), modes (90.5%), analysis (89.3%), validation (88.8%), storage (87.9%), orchestration (87.7%), metacognition (87.2%).

### Linting
```bash
# Run golangci-lint (requires golangci-lint installation)
golangci-lint run

# Run with specific linters
golangci-lint run --enable-all

# Fix auto-fixable issues
golangci-lint run --fix
```

**Linter Configuration**: `.golangci.yml` configures enabled linters (errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, misspell). Compatible with golangci-lint v2.x schema.

### Cleanup
```bash
make clean  # Removes bin/ directory
```

## MCP Protocol Communication

**Key Understanding**:
- All responses are JSON-formatted for Claude AI consumption
- No human-readable text formatting is used
- Claude AI processes the structured JSON data directly
- Responses contain only the data structures defined in the response types

## MCP Tool Registration

Tools are registered in `internal/server/server.go` using the pattern:
```go
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "tool-name",
    Description: "Tool description",
}, s.handleToolName)
```

Each handler returns structured JSON via `toJSONContent(responseData)`.

**Available Tools** (63 total):

**Core Thinking Tools** (11 tools):
1. `think` - Main thinking tool (supports all modes: linear, tree, divergent, reflection, backtracking, auto)
2. `history` - View thinking history (filtered by mode/branch)
3. `list-branches` - List all branches (tree mode)
4. `focus-branch` - Switch active branch
5. `branch-history` - Get detailed branch history with insights/cross-refs
6. `recent-branches` - Get recently accessed branches for quick context switching
7. `validate` - Validate thought for logical consistency
8. `prove` - Attempt formal proof from premises to conclusion
9. `check-syntax` - Validate logical statement syntax
10. `search` - Search thoughts by query and optional mode filter
11. `get-metrics` - Get system performance and usage metrics

**Probabilistic Reasoning Tools** (4 tools):
12. `probabilistic-reasoning` - Bayesian inference and belief updates (create/update/get/combine)
13. `assess-evidence` - Evidence quality and reliability assessment
14. `detect-contradictions` - Find contradictions among thoughts or within branches
15. `sensitivity-analysis` - Test robustness of conclusions to assumption changes

**Decision & Problem-Solving Tools** (3 tools):
16. `make-decision` - Multi-criteria decision analysis with weighted scoring
17. `decompose-problem` - Break complex problems into manageable subproblems
18. `verify-thought` - Verify thought validity and structure

**Metacognition Tools** (3 tools):
19. `self-evaluate` - Metacognitive self-assessment of reasoning quality
20. `detect-biases` - Identify cognitive biases and logical fallacies
21. `detect-blind-spots` - Identify unknown unknowns and knowledge gaps

**Hallucination & Calibration Tools** (4 tools):
22. `get-hallucination-report` - Retrieve hallucination detection reports
23. `record-prediction` - Record a prediction for calibration tracking
24. `record-outcome` - Record prediction outcomes
25. `get-calibration-report` - Retrieve confidence calibration analysis

**Perspective & Temporal Analysis Tools** (4 tools):
26. `analyze-perspectives` - Multi-stakeholder perspective analysis
27. `analyze-temporal` - Short-term vs long-term temporal reasoning
28. `compare-time-horizons` - Compare implications across time horizons
29. `identify-optimal-timing` - Determine optimal timing for decisions

**Causal Reasoning Tools** (5 tools):
30. `build-causal-graph` - Construct causal graphs from observations
31. `simulate-intervention` - Simulate interventions using do-calculus
32. `generate-counterfactual` - Generate "what if" scenarios
33. `analyze-correlation-vs-causation` - Distinguish correlation from causation
34. `get-causal-graph` - Retrieve previously built causal graph

**Integration & Orchestration Tools** (6 tools):
35. `synthesize-insights` - Synthesize insights from multiple reasoning modes
36. `detect-emergent-patterns` - Detect emergent patterns across modes
37. `execute-workflow` - Execute predefined multi-tool workflows
38. `list-workflows` - List available automated workflows
39. `register-workflow` - Register custom workflows for automation
40. `list-integration-patterns` - Discover integration patterns

**Dual-Process Reasoning Tools** (1 tool):
41. `dual-process-think` - System 1 (fast) vs System 2 (deliberate) reasoning

**Backtracking Tools** (3 tools):
42. `create-checkpoint` - Create reasoning checkpoint
43. `restore-checkpoint` - Restore from checkpoint
44. `list-checkpoints` - List available checkpoints

**Abductive Reasoning Tools** (2 tools):
45. `generate-hypotheses` - Generate explanatory hypotheses
46. `evaluate-hypotheses` - Evaluate hypothesis plausibility

**Case-Based Reasoning Tools** (2 tools):
47. `retrieve-similar-cases` - Retrieve similar cases from memory
48. `perform-cbr-cycle` - Execute full CBR cycle (retrieve, reuse, revise, retain)

**Symbolic Reasoning Tools** (2 tools):
49. `prove-theorem` - Formal theorem proving
50. `check-constraints` - Check symbolic constraint satisfaction

**Enhanced Tools** (8 tools):
51. `find-analogy` - Find analogies between source and target domains for cross-domain reasoning
52. `apply-analogy` - Apply an existing analogy to a new context
53. `decompose-argument` - Break down arguments into premises, claims, assumptions, and inference chains
54. `generate-counter-arguments` - Generate counter-arguments using multiple strategies
55. `detect-fallacies` - Detect formal and informal logical fallacies (ad hominem, straw man, false dichotomy, etc.)
56. `process-evidence-pipeline` - Automatically update beliefs, causal graphs, and decisions from new evidence
57. `analyze-temporal-causal-effects` - Analyze temporal progression of causal effects (short/medium/long-term)
58. `analyze-decision-timing` - Determine optimal timing for decisions based on causal and temporal factors

**Episodic Memory & Learning Tools** (5 tools):
59. `start-reasoning-session` - Start tracking a reasoning session to learn from experience
60. `complete-reasoning-session` - Complete session, calculate quality metrics, trigger pattern learning
61. `get-recommendations` - Get adaptive recommendations based on similar past problems
62. `search-trajectories` - Search past reasoning sessions to learn from successes and failures
63. `analyze-trajectory` - Perform retrospective analysis of completed session (strengths, weaknesses, improvements, lessons learned)

## Storage Architecture

The server supports **pluggable storage backends** via the `storage.Storage` interface with two implementations:

### In-Memory Storage (Default)
- Backend: `storage/memory.go`
- Thread-safe with sync.RWMutex
- No persistence - data lost on restart
- Fast, zero configuration
- Ideal for development and testing

### SQLite Storage (Optional)
- Backend: `storage/sqlite.go`
- Persistent across server restarts
- Write-through caching for performance
- Full-text search via FTS5
- WAL mode for concurrent reads
- Graceful fallback to memory on errors

**Configuration** via environment variables:
```bash
STORAGE_TYPE=memory     # Default - in-memory only
STORAGE_TYPE=sqlite     # Persistent SQLite storage

SQLITE_PATH=./data/thoughts.db  # Database file path
SQLITE_TIMEOUT=5000             # Connection timeout (ms)
STORAGE_FALLBACK=memory         # Fallback if SQLite fails
```

**Storage Factory** (`storage/factory.go`):
- `NewStorageFromEnv()` - Creates storage from environment variables
- `NewStorage(cfg Config)` - Creates storage from explicit configuration
- Automatic fallback handling

**Storage Operations**:
- Thread-safe with RWMutex locking
- Auto-generates IDs using counters + timestamps
- Maintains active branch state
- Full-text search (FTS5 in SQLite, substring in memory)

**Key Methods**:
- `StoreThought()`, `GetThought()`, `SearchThoughts()`
- `StoreBranch()`, `GetBranch()`, `ListBranches()`, `GetActiveBranch()`, `SetActiveBranch()`
- `StoreInsight()`, `StoreValidation()`, `StoreRelationship()`

## Data Flow

### Basic Thinking Flow
1. **Tool Call** → `server/server.go` handler receives request
2. **Mode Selection** → Auto mode detects or explicit mode used
3. **Processing** → Selected mode's `ProcessThought()` executes
4. **Storage** → Thought/Branch/Insight persisted via storage backend (memory or SQLite)
5. **Validation** (optional) → Logic validator checks consistency
6. **Response** → Result returned to MCP client

### Workflow Orchestration Flow
1. **Workflow Call** → `execute-workflow` with workflow_id and input
2. **Context Creation** → ReasoningContext tracks shared state across steps
3. **Step Execution** → Sequential or parallel execution of workflow steps
4. **Tool Invocation** → Each step calls appropriate cognitive tool
5. **Result Transformation** → Output transformations and conditional logic
6. **Context Updates** → Results stored in context for dependent steps
7. **Final Synthesis** → Aggregated results returned with execution metadata

Workflow orchestration enables automated multi-tool reasoning pipelines without manual coordination.

## Important Implementation Details

### Storage Architecture
- **Interface-based**: All code depends on `storage.Storage` interface for testability
- **Thread-safe**: RWMutex protection with deep copy strategy
- **Resource limits**: MaxSearchResults=1000, MaxIndexSize=100000 to prevent DoS
- **Optimized appends**: Direct append methods avoid full get-modify-store cycles

### SQLite Persistence Implementation
When using SQLite storage (`STORAGE_TYPE=sqlite`):

**Architecture**:
- **Write-through cache**: All writes go to DB first, then update in-memory cache
- **Cache-first reads**: Reads hit cache first (fast path), DB on miss (warm cache)
- **Deep copying**: All returns are deep copies to prevent data races
- **Prepared statements**: Pre-compiled SQL statements for performance

**Schema Design** (`sqlite_schema.go`):
- Core tables: `thoughts`, `branches`, `insights`, `cross_refs`, `validations`, `relationships`
- FTS5 virtual table: `thoughts_fts` for full-text search on thought content
- Indexes: Optimized for common query patterns (mode filtering, branch lookups, timestamps)
- JSON columns: `key_points`, `metadata` stored as JSON for complex types

**Performance Optimizations**:
- WAL mode: Concurrent reads while writing
- 64MB cache size: Reduces disk I/O
- Batch inserts: Transaction batching for multi-record writes
- Connection pooling: Reuses DB connections
- FTS5 tokenization: Fast full-text search with relevance ranking

**Data Migration**:
- Schema versioning: `schema_version` table tracks migrations
- Automatic upgrades: Applies migrations on startup if needed
- Backward compatible: Old data remains accessible after upgrades

### Builder Patterns
Use builders from `internal/types/builders.go` for object construction:
```go
thought := types.NewThought().
    Content("Example").
    Mode(types.ModeLinear).
    Confidence(0.9).
    Build()
```

### Mode Registry
Modes implement `ThinkingMode` interface and can be registered dynamically:
- `Name()` - Returns mode identifier
- `CanHandle()` - Determines if mode can process input
- `ProcessThought()` - Executes mode logic

### Branch Metrics Calculation
When processing thoughts in tree mode (`modes/tree.go`):
- Branch confidence = average of all thought confidences in branch
- Branch priority = confidence + (insight_count × 0.1) + (sum of cross_ref strengths × 0.1)

### Cross-References
Cross-references link branches together with typed relationships:
- `complementary` - Ideas that work well together
- `contradictory` - Conflicting approaches
- `builds_upon` - Extends another branch's ideas
- `alternative` - Different approach to same problem

TouchPoints within cross-refs specify exact thought-to-thought connections.

### Workflow Orchestration
The orchestration system (`internal/orchestration/`) enables automated multi-tool reasoning:
- **Sequential workflows**: Execute steps in order with dependency tracking
- **Parallel workflows**: Concurrent execution of independent steps
- **Conditional workflows**: Execute steps based on intermediate results
- **ReasoningContext**: Shared state tracking across workflow execution
- **Built-in workflows**: Pre-configured workflows like "comprehensive-analysis"
- **Custom workflows**: Register custom workflows via `register-workflow` tool

Example workflow structure:
```go
workflow := &Workflow{
    Type: WorkflowSequential,
    Steps: []*WorkflowStep{
        {Tool: "think", Input: {...}},
        {Tool: "assess-evidence", DependsOn: ["step1"]},
        {Tool: "make-decision", DependsOn: ["step1", "step2"]},
    },
}
```

### Causal Reasoning
Implements Pearl's causal inference framework (`internal/reasoning/causal.go`):
- **Causal graphs**: Build directed acyclic graphs (DAGs) from observations
- **Graph surgery**: Proper do-calculus for interventions (removing incoming edges)
- **Counterfactuals**: Generate "what if" scenarios by changing variables
- **Intervention simulation**: Test effects of interventions on outcomes
- Correctly distinguishes correlation from causation using structural causal models

### Fallacy Detection
Enhanced validation (`internal/validation/fallacies.go`) detects 20+ logical fallacies:
- **Formal fallacies**: Affirming consequent, denying antecedent, circular reasoning
- **Informal fallacies**: Ad hominem, straw man, false dilemma, slippery slope
- **Evidence fallacies**: Hasty generalization, cherry picking, anecdotal evidence
- **Appeal fallacies**: Appeal to authority, emotion, tradition, nature
- Integrated with cognitive bias detection for comprehensive reasoning quality assessment

## Migration from Old Servers

This server replaces multiple TypeScript servers. Tool mapping:

| Old Server | Old Tool | New Tool | Usage |
|------------|----------|----------|-------|
| sequential-thinking | solve-problem | think | `mode: "linear"` |
| branch-thinking | branch-thinking | think | `mode: "tree"` |
| unreasonable-thinking | generate_unreasonable_thought | think | `mode: "divergent"`, `force_rebellion: true` |
| mcp-logic | prove | prove | Same interface |
| mcp-logic | check-well-formed | check-syntax | Similar functionality |

## How MCP Servers Work

**IMPORTANT**: MCP servers are NOT standalone executables. They are:
- Automatically started by Claude Desktop when the app launches
- Run as child processes communicating via stdio (standard input/output)
- Managed entirely by the Claude Desktop application lifecycle
- Terminated when Claude Desktop closes

The server binary (`bin/unified-thinking.exe`) should **NEVER** be run manually by users.

## Configuration

Add to Claude Desktop config (`%APPDATA%\Claude\claude_desktop_config.json` on Windows):

### Default Configuration (In-Memory)
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\Development\\Projects\\MCP\\project-root\\mcp-servers\\unified-thinking\\bin\\unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

### With SQLite Persistence
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\Development\\Projects\\MCP\\project-root\\mcp-servers\\unified-thinking\\bin\\unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "C:\\Users\\YourName\\AppData\\Roaming\\Claude\\unified-thinking.db",
        "STORAGE_FALLBACK": "memory"
      }
    }
  }
}
```

### With Semantic Embeddings (Recommended)
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\Development\\Projects\\MCP\\project-root\\mcp-servers\\unified-thinking\\bin\\unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "C:\\Users\\YourName\\AppData\\Roaming\\Claude\\unified-thinking.db",
        "STORAGE_FALLBACK": "memory",
        "EMBEDDINGS_ENABLED": "true",
        "EMBEDDINGS_PROVIDER": "voyage",
        "EMBEDDINGS_MODEL": "voyage-3-lite",
        "VOYAGE_API_KEY": "your-voyage-api-key",
        "EMBEDDINGS_HYBRID_SEARCH": "true",
        "EMBEDDINGS_RRF_K": "60",
        "EMBEDDINGS_MIN_SIMILARITY": "0.5",
        "EMBEDDINGS_CACHE_ENABLED": "true",
        "EMBEDDINGS_CACHE_TTL": "24h"
      }
    }
  }
}
```

**Environment Variables**:
- `STORAGE_TYPE`: `memory` (default) or `sqlite`
- `SQLITE_PATH`: Path to SQLite database file (created if not exists)
- `SQLITE_TIMEOUT`: Connection timeout in milliseconds (default: 5000)
- `STORAGE_FALLBACK`: Fallback storage type if primary fails (default: none)
- `DEBUG`: Enable debug logging (`true` or `false`)
- `AUTO_VALIDATION_THRESHOLD`: Confidence threshold below which auto-validation triggers (default: 0.5, range: 0.0-1.0)

**Embeddings Environment Variables** (for semantic search in episodic memory):
- `EMBEDDINGS_ENABLED`: Enable semantic embeddings (`true` or `false`, default: false)
- `EMBEDDINGS_PROVIDER`: Embedding provider (`voyage` supported)
- `EMBEDDINGS_MODEL`: Model to use (`voyage-3-lite` (512d), `voyage-3` (1024d), `voyage-3-large` (2048d))
- `VOYAGE_API_KEY`: Your Voyage AI API key (get from https://dashboard.voyageai.com/)
- `EMBEDDINGS_HYBRID_SEARCH`: Combine semantic + structured search (`true` or `false`)
- `EMBEDDINGS_RRF_K`: RRF fusion parameter (default: 60)
- `EMBEDDINGS_MIN_SIMILARITY`: Minimum cosine similarity threshold (default: 0.5)
- `EMBEDDINGS_CACHE_ENABLED`: Cache embeddings (`true` or `false`)
- `EMBEDDINGS_CACHE_TTL`: Cache time-to-live (e.g., `24h`, `1h`)

**Context Bridge Environment Variables** (for cross-session context retrieval):
- `CONTEXT_BRIDGE_ENABLED`: Enable context bridge (`true` or `false`, default: false)
- `CONTEXT_BRIDGE_MIN_SIMILARITY`: Minimum similarity threshold (default: 0.7)
- `CONTEXT_BRIDGE_MAX_MATCHES`: Maximum matches to return (default: 3)
- `CONTEXT_BRIDGE_CACHE_SIZE`: LRU cache size (default: 100)
- `CONTEXT_BRIDGE_CACHE_TTL`: Cache TTL (default: 15m)
- `CONTEXT_BRIDGE_TIMEOUT`: Timeout per enrichment (default: 2s)

**Hybrid Similarity Mode**: When `VOYAGE_API_KEY` is set, the context bridge uses a hybrid similarity algorithm:
- 70% weight: Cosine similarity on Voyage AI embeddings (voyage-3-lite, 512 dimensions)
- 30% weight: Jaccard similarity on extracted concepts

This enables matching between semantically similar but lexically different problems (e.g., "database queries running slowly" matches "SQL performance optimization needed").

**Graceful Degradation**: The context bridge never blocks or errors the primary tool response:
- If embedding generation times out (500ms sub-timeout), falls back to concept-only similarity
- If overall enrichment times out (2s), returns degraded response with status
- All failures are logged and metriced for observability

**Response Formats**:

Normal response with matches:
```json
{
  "result": { ... },
  "context_bridge": {
    "version": "1.0",
    "matches": [...],
    "recommendation": "Similar past reasoning had high success rates.",
    "similarity_mode": "semantic_embedding"
  }
}
```

Response when embedding fails (uses concept similarity):
```json
{
  "result": { ... },
  "context_bridge": {
    "version": "1.0",
    "matches": [...],
    "embedding_status": "failed",
    "embedding_error": "context deadline exceeded",
    "similarity_mode": "concept_only"
  }
}
```

Response when timeout occurs:
```json
{
  "result": { ... },
  "context_bridge": {
    "version": "1.0",
    "status": "degraded",
    "reason": "timeout",
    "detail": "Context bridge timeout exceeded (2s)",
    "matches": []
  }
}
```

### Auto-Validation Feature
When a thought is processed with confidence below the threshold (default 0.5):
1. The system automatically runs self-evaluation to assess quality
2. If significant issues are found (quality/completeness/coherence < 0.5)
3. The thought is automatically retried with `ChallengeAssumptions=true`
4. Metadata tracks auto-validation and scores for transparency

When Claude Desktop starts, it will:
1. Read this configuration
2. Spawn the server process using the specified command
3. Establish stdio communication channel
4. Initialize storage backend based on environment variables
5. Keep the server running for the entire session
6. Terminate the server when Claude Desktop closes

## Code Style & Conventions

- Use Go standard formatting (`go fmt`)
- Package-level documentation in godoc style
- Error handling: return errors, don't panic (except in main.go for fatal errors)
- Logging: Use `log.Println()` for debug info (when DEBUG=true)
- JSON tags: Use standard Go JSON tags for request/response structs
- Thread safety: Use sync.RWMutex for shared state access

## Key Files to Understand

### Core Infrastructure
1. `cmd/server/main.go` - Entry point, initializes storage and server components
2. `internal/types/types.go` - All core data structures and constants (50+ types)
3. `internal/server/server.go` - Tool registration and request handlers (50 tools)
4. `internal/server/handlers/` - Modular handlers for thinking, branches, validation, search, enhanced cognitive tools

### Storage Layer
5. `internal/storage/factory.go` - Storage factory and configuration
6. `internal/storage/memory.go` - In-memory storage implementation
7. `internal/storage/sqlite.go` - SQLite storage with write-through cache
8. `internal/storage/sqlite_schema.go` - Database schema and migrations

### Thinking Modes
9. `internal/modes/shared.go` - Shared types for mode implementations
10. `internal/modes/auto.go` - Automatic mode selection logic
11. `internal/modes/registry.go` - Dynamic mode registration system

### Cognitive Reasoning
12. `internal/reasoning/probabilistic.go` - Bayesian inference implementation
13. `internal/reasoning/causal.go` - Pearl's causal inference framework with do-calculus
14. `internal/reasoning/decision.go` - Multi-criteria decision analysis
15. `internal/reasoning/temporal.go` - Temporal reasoning (short/long-term analysis)

### Analysis & Validation
16. `internal/validation/logic.go` - Logical validation and proof checking
17. `internal/validation/fallacies.go` - Fallacy detection (20+ types)
18. `internal/analysis/evidence.go` - Evidence quality assessment
19. `internal/analysis/contradiction.go` - Contradiction detection across thoughts
20. `internal/analysis/perspective.go` - Multi-stakeholder perspective analysis

### Metacognition
21. `internal/metacognition/self_eval.go` - Self-evaluation implementation
22. `internal/metacognition/bias_detection.go` - Cognitive bias detection (15+ biases)

### Integration & Orchestration
23. `internal/integration/synthesizer.go` - Cross-mode insight synthesis
24. `internal/orchestration/workflow.go` - Workflow orchestration system
25. `internal/orchestration/interface.go` - ToolExecutor abstraction for workflow steps

## Recent Updates

### Code Quality Improvements

**Refactoring (Commits 5807e7b, e2a241d)**
- Attempted modularization of monolithic server.go (2,225 lines)
- Created tools.go with centralized tool definitions (800 lines)
- Discovered Go type system limitation: mcp.AddTool requires compile-time type info, preventing dynamic registration
- Removed unsuccessful refactoring experiments to maintain clean codebase

**Test Coverage Improvements (Commits 7ebdf3b, d2a9280)**
- Memory module coverage: 35.3% → 67.8% (+32.5 percentage points)
- Handler module coverage: 43.0% → 47.2% (+4.2 percentage points)
- Added comprehensive tests for temporal and causal handlers
- Fixed test API mismatches and compilation errors

**Code Formatting & Linting (Latest cleanup)**
- Applied gofmt -s to entire codebase (23 files formatted)
- Resolved all golangci-lint v2.x compatibility issues
- Removed unused functions and imports
- Maintained clean, consistent code style throughout

## Recent Fixes

### Episodic Memory Fixes (Commit e7cc5c7)

**Issue 1: SQLite Foreign Key Constraint Failure**
- **Problem**: SQLite initialization failing with "no such column: branch_id", server falling back to memory storage
- **Root Cause**: Duplicate `thoughts` table definition with one created before `branches` table, causing foreign key validation failure with `PRAGMA foreign_keys=ON`
- **Fix**: Removed duplicate table definition, ensured correct table ordering (branches before thoughts)
- **Files Modified**: `internal/storage/sqlite_schema.go`
- **Impact**: Episodic memory now persists correctly across Claude Desktop restarts

**Issue 2: Nil Array Validation Errors**
- **Problem**: Three episodic memory tools failing with MCP validation error: "type null, want array"
- **Affected Tools**: `get-recommendations`, `search-trajectories`, `start-reasoning-session`
- **Root Cause**: MCP tool registration functions creating empty response structs with nil array fields for schema validation
- **Fix**:
  - Modified tool registration to properly unmarshal responses for MCP validation
  - Initialized all array fields in response structs to prevent nil values
  - Added defensive initialization in nested structs (Recommendation, TrajectoryPattern)
- **Files Modified**:
  - `internal/server/handlers/episodic.go` (3 tool registration functions)
  - `internal/memory/episodic.go` (Recommendation struct initialization)
  - `internal/memory/learning.go` (TrajectoryPattern struct initialization)
- **Impact**: All episodic memory tools now return valid JSON arrays instead of null values

**Verification**:
- SQLite: Check logs for "SQLite storage initialized successfully" (no "falling back to memory")
- Tools: All episodic memory tools return `[]` (empty arrays) instead of `null` when no data exists
- Persistence: Reasoning sessions survive Claude Desktop restarts

## Technical Constraints

- Go 1.24+ required
- Windows primary target (Makefile uses Windows commands by default)
- MCP SDK v0.8.0 - uses `mcp.AddTool()` and `transport.Run()` patterns
- Storage: In-memory (default) or SQLite (optional)
- SQLite backend uses modernc.org/sqlite (pure Go, no CGO)
- stdio transport only (no HTTP/SSE)
