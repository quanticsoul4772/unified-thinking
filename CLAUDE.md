# CLAUDE.md

Guidance for Claude Code when working with this repository.

## Project Overview

Go-based MCP server consolidating 5 TypeScript servers (sequential-thinking, branch-thinking, unreasonable-thinking-server, mcp-logic, state-coordinator) into one unified implementation with 85 cognitive reasoning tools.

**Module**: `unified-thinking` | **Entry**: `cmd/server/main.go` | **SDK**: `github.com/modelcontextprotocol/go-sdk` v0.8.0

### Key Components

| Package | Purpose |
|---------|---------|
| `internal/types/` | Core data structures, Builder patterns, 50+ cognitive reasoning types |
| `internal/storage/` | Pluggable storage (in-memory default, SQLite optional) |
| `internal/modes/` | Thinking modes: linear, tree, divergent, reflection, backtracking, auto, graph-of-thoughts |
| `internal/processing/` | Dual-process reasoning (System 1/2) |
| `internal/validation/` | Logic validation, proof checking, fallacy detection |
| `internal/reasoning/` | Probabilistic, decision, causal, temporal, abductive, case-based reasoning |
| `internal/analysis/` | Evidence, contradiction, perspective, sensitivity analysis |
| `internal/metacognition/` | Self-evaluation, bias detection, unknown unknowns |
| `internal/integration/` | Cross-mode synthesis, pattern detection |
| `internal/orchestration/` | Workflow orchestration for multi-tool pipelines |
| `internal/memory/` | Episodic memory with trajectory storage and pattern learning |
| `internal/embeddings/` | Voyage AI semantic embeddings (voyage-3-lite, 512d) + reranking (rerank-2) + multimodal (voyage-multimodal-3, 1024d) |
| `internal/contextbridge/` | Cross-session context retrieval with caching |
| `internal/server/` | MCP server with 85 tools, 27 handler modules |
| `internal/similarity/` | Thought similarity search via embeddings |
| `internal/claudecode/` | Claude Code optimizations: format, errors, session, presets |
| `internal/streaming/` | MCP progress notifications for long-running tools |
| `internal/testutil/` | Testing utilities: MockLLMClient for GoT testing without API calls |

### Thinking Modes

1. **Linear** - Sequential step-by-step reasoning
2. **Tree** - Multi-branch parallel exploration with cross-references
3. **Divergent** - Creative ideation with "rebellion" capability
4. **Reflection** - Metacognitive analysis of previous reasoning
5. **Backtracking** - Checkpoint-based with restore capabilities
6. **Auto** - Automatic mode selection via embeddings or keywords
7. **Graph** - Graph-of-Thoughts with aggregation, refinement, cyclic reasoning

**Auto Mode Selection**: Uses semantic embeddings (VOYAGE_API_KEY) or keyword detection. Explicit overrides: `ForceRebellion` → divergent, `BranchID/CrossRefs/KeyPoints` → tree.

## Development Commands

### Make (Linux/macOS/WSL)
```bash
make build              # Windows: bin/unified-thinking.exe
make linux              # Linux: bin/unified-thinking
make test               # All tests
make test-coverage      # With coverage report
make clean              # Remove bin/
```

### PowerShell (Windows)
```powershell
.\build.ps1 build       # Windows: bin\unified-thinking.exe
.\build.ps1 linux       # Cross-compile for Linux
.\build.ps1 test        # All tests
.\build.ps1 test-coverage # With coverage report
.\build.ps1 pre-commit  # Quick checks (fmt + vet + test-short)
.\build.ps1 clean       # Remove bin\
.\build.ps1 help        # Show all 25+ commands
```

### Pre-commit Hooks
```bash
# Install (one-time setup)
./scripts/install-hooks.sh      # Unix
.\scripts\install-hooks.ps1     # Windows

# Hook runs automatically on git commit:
# - go fmt (formatting check)
# - go vet (static analysis)
# - go test -short (quick tests)
# - golangci-lint --fast (if available)

# Skip temporarily: git commit --no-verify
```

**DO NOT run server manually** - Claude Desktop starts it automatically.

**Test Coverage**: 72.1% overall, 148 test files across 30 packages. Key: metrics (100%), presets (98.3%), config (97.3%), reinforcement (90.2%), reasoning (89.7%), analysis (89.3%).

## MCP Tools (85 total)

### Core (11)
`think`, `history`, `list-branches`, `focus-branch`, `branch-history`, `recent-branches`, `validate`, `prove`, `check-syntax`, `search`, `get-metrics`

### Probabilistic (4)
`probabilistic-reasoning`, `assess-evidence`, `detect-contradictions`, `sensitivity-analysis`

### Decision (2)
`make-decision`, `decompose-problem`

### Metacognition (3)
`self-evaluate`, `detect-biases`, `detect-blind-spots`

### Hallucination & Calibration (5)
`verify-thought`, `get-hallucination-report`, `record-prediction`, `record-outcome`, `get-calibration-report`

### Perspective & Temporal (4)
`analyze-perspectives`, `analyze-temporal`, `compare-time-horizons`, `identify-optimal-timing`

### Causal (5)
`build-causal-graph`, `simulate-intervention`, `generate-counterfactual`, `analyze-correlation-vs-causation`, `get-causal-graph`

### Integration & Orchestration (6)
`synthesize-insights`, `detect-emergent-patterns`, `execute-workflow`, `list-workflows`, `register-workflow`, `list-integration-patterns`

### Dual-Process (1)
`dual-process-think`

### Backtracking (3)
`create-checkpoint`, `restore-checkpoint`, `list-checkpoints`

### Abductive (2)
`generate-hypotheses`, `evaluate-hypotheses`

### Case-Based (2)
`retrieve-similar-cases`, `perform-cbr-cycle`

### Symbolic (2)
`prove-theorem`, `check-constraints`

### Enhanced (8)
`find-analogy`, `apply-analogy`, `decompose-argument`, `generate-counter-arguments`, `detect-fallacies`, `process-evidence-pipeline`, `analyze-temporal-causal-effects`, `analyze-decision-timing`

### Episodic Memory (5)
`start-reasoning-session`, `complete-reasoning-session`, `get-recommendations`, `search-trajectories`, `analyze-trajectory`

### Knowledge Graph (3)
`store-entity`, `search-knowledge-graph`, `create-relationship`

### Similarity (1)
`search-similar-thoughts`

### Graph-of-Thoughts (9)
`got-initialize`, `got-generate`, `got-aggregate`, `got-refine`, `got-score`, `got-prune`, `got-get-state`, `got-finalize`, `got-explore`

### Claude Code Optimization (5)
`export-session`, `import-session`, `list-presets`, `run-preset`, `format-response`

### Research (1)
`research-with-search` - Web-augmented research using Anthropic's server-side web search (requires `ANTHROPIC_API_KEY` + `WEB_SEARCH_ENABLED=true`)

### Multimodal (1)
`embed-multimodal` - Generate embeddings for images using Voyage AI's multimodal model (requires `VOYAGE_API_KEY` + `MULTIMODAL_ENABLED=true`)

### Agentic (2)
`run-agent`, `list-agent-tools` - Programmatic tool calling via agentic LLM loop (requires `ANTHROPIC_API_KEY` + `AGENT_ENABLED=true`)

## Claude Code Optimization

Tools and features specifically designed to optimize unified-thinking usage within Claude Code.

### Response Formatting

Set `RESPONSE_FORMAT` env var to control token usage:
- **full** (default): Complete response with all metadata
- **compact**: 40-60% reduction - removes context_bridge, flattens next_tools, truncates arrays to 5 items
- **minimal**: 80%+ reduction - essential fields only, arrays truncated to 3 items

Or use `format-response` tool for per-request control.

### Session Export/Import

Preserve reasoning context across sessions:
- **export-session**: Export thoughts, branches, decisions, causal graphs to portable JSON (with optional gzip compression)
- **import-session**: Restore with merge strategies: `replace`, `merge`, `append`

### Workflow Presets

8 built-in presets for common development tasks:

| Preset | Category | Steps | Description |
|--------|----------|-------|-------------|
| `code-review` | code | 5 | Multi-aspect code review |
| `debug-analysis` | code | 4 | Causal debugging with hypothesis generation |
| `refactoring-plan` | code | 4 | Safe refactoring with impact analysis |
| `architecture-decision` | architecture | 5 | ADR-style decision workflow |
| `research-synthesis` | research | 4 | Graph-of-Thoughts research aggregation |
| `test-strategy` | testing | 4 | Test coverage planning |
| `documentation-gen` | documentation | 4 | Multi-perspective documentation |
| `incident-investigation` | operations | 5 | Post-incident analysis with timeline mapping |

Use `list-presets` to view available presets, `run-preset` to execute (supports dry_run and step_by_step modes).

### Structured Errors

Tool errors include recovery guidance:
```json
{
  "code": "MISSING_REQUIRED",
  "message": "preset_id is required",
  "details": "The preset_id parameter must specify which preset to run",
  "recovery_suggestions": ["Use list-presets to see available presets"],
  "related_tools": ["list-presets"],
  "example": {"tool": "run-preset", "input": {"preset_id": "code-review", ...}}
}
```

### Streaming Progress Notifications

Long-running tools support real-time progress updates via MCP `notifications/progress`. Clients providing a `progressToken` receive step-by-step updates.

**Streaming-Enabled Tools:**

| Priority | Tools |
|----------|-------|
| P0 (Essential) | `execute-workflow`, `run-preset`, `got-generate` |
| P1 (Important) | `got-aggregate`, `think`, `perform-cbr-cycle` |
| P2 (Enhancement) | `synthesize-insights`, `analyze-perspectives`, `build-causal-graph`, `evaluate-hypotheses` |

**Features:**
- Rate-limited notifications (100ms default) to prevent flooding
- Step changes bypass rate limit for immediate feedback
- No-op when client doesn't provide `progressToken` (backward compatible)
- Per-tool configuration for interval, partial data, auto-progress

See [docs/STREAMING.md](./docs/STREAMING.md) for detailed documentation.

### Graph-of-Thoughts Exploration

The `got-explore` tool orchestrates a complete GoT workflow in a single call, reducing 6+ tool calls to 1:

```json
{
  "initial_thought": "How to optimize database query performance?",
  "problem": "Slow queries affecting user experience",
  "config": {
    "k": 3,
    "max_iterations": 1,
    "prune_threshold": 0.3,
    "refine_top_n": 1,
    "use_fast_scoring": true
  }
}
```

**Configuration Options**:

| Option | Default | Description |
|--------|---------|-------------|
| `k` | 3 | Number of diverse continuations per vertex |
| `max_iterations` | 1 | Exploration depth (more = slower but deeper) |
| `prune_threshold` | 0.3 | Minimum score to keep vertices |
| `refine_top_n` | 1 | How many top vertices to refine |
| `score_all` | false | Score all vertices vs just generated ones |
| `use_fast_scoring` | true | Use local heuristics (fast) vs LLM scoring (quality) |
| `skip_refine` | false | Skip refinement step entirely |
| `parallel_scoring` | true | Parallel LLM scoring when not using fast scoring |

**Performance Modes**:
- **Default** (~5 seconds): Fast local scoring, single iteration
- **Thorough** (slower): LLM scoring, multiple iterations - use `ThoroughExploreConfig`

**Workflow Steps**:
1. Initialize graph with initial thought
2. Generate k diverse continuations
3. Score all vertices (fast heuristics or LLM)
4. Prune low-quality thoughts
5. Refine top-scoring vertices
6. Repeat for max_iterations
7. Finalize and return conclusions

### Domain-Aware Problem Decomposition

The `decompose-problem` tool now supports domain-specific templates:

| Domain | Steps | Keywords |
|--------|-------|----------|
| `debugging` | 6 | debug, bug, error, fix, crash, trace, exception |
| `proof` | 7 | prove, theorem, lemma, axiom, formal, verify |
| `architecture` | 6 | architect, design, system, component, module, api |
| `research` | 7 | research, study, analyze, explore, benchmark |
| `general` | 5 | Default for unclassified problems |

**Usage**:
```json
{
  "problem": "Debug the flaky test in CI",
  "domain": "debugging"  // Optional: auto-detected if not specified
}
```

**Response includes**:
- `detected_domain`: Domain used for decomposition
- `domain_was_explicit`: Whether domain was specified or auto-detected
- Domain-specific subproblems and dependencies

### Enhanced Episodic Memory

The `get-recommendations` tool now returns:
- **Specific tool sequences** from similar successful trajectories
- **Concrete examples** with success rates
- **Approach warnings** for historically unsuccessful patterns

### Auto-Calibrated Bias Detection

The `detect-biases` tool now tracks historical false positive rates and suppresses low-confidence detections. Requires minimum sample size before calibration activates.

### Automatic Confidence Tracking

Predictions are automatically recorded on `think` calls and outcomes recorded on `validate` calls, enabling passive calibration improvement.

### Voyage AI Reranking

Search results are now automatically reranked using Voyage AI's rerank models when `VOYAGE_API_KEY` is set:

**Affected Tools**:
- `search-similar-thoughts` - Reranks thought similarity results
- `search-knowledge-graph` - Reranks semantic search results

**Pipeline**: Query → Embedding Search (2x limit) → Rerank → Return Top Results

**Models**:
- `rerank-2` (default): Best quality, ~150ms latency
- `rerank-2-lite`: Faster, ~50ms latency

**Disable**: Set `RERANK_ENABLED=false` to use embedding-only scoring.

### LRU Embedding Cache

The embedding cache now uses LRU (Least Recently Used) eviction with optional disk persistence:

**Features**:
- **LRU Eviction**: Automatically evicts least-used entries when max size is reached
- **Disk Persistence**: Save cache to disk, load on startup (gzip compressed)
- **TTL Expiry**: Entries automatically expire after configurable time
- **Auto-Save**: Dirty caches save automatically every 5 minutes
- **Thread-Safe**: Full concurrent read/write support

**Configuration**:
```bash
EMBEDDINGS_CACHE_MAX_ENTRIES=10000   # Max cache entries (default: 10000, 0 = unlimited)
EMBEDDINGS_CACHE_PERSIST=true         # Enable disk persistence
EMBEDDINGS_CACHE_PATH=/path/to/cache  # Cache file path (auto-enables persist)
```

**Cache Stats**: Available via `cache.Stats()`:
```json
{
  "size": 5000,
  "max_size": 10000,
  "hits": 12500,
  "misses": 250,
  "hit_rate": 0.98,
  "evictions": 100,
  "expiries": 50
}
```

**Memory Estimate**: ~20MB for 10K entries with 512-dimensional embeddings.

### Domain-Specific Model Configuration

The system automatically detects task domains and applies optimized model configurations:

| Domain | Keywords | Temperature | Use Case |
|--------|----------|-------------|----------|
| `code` | debug, function, api, error, implement | 0.3 | Code generation, debugging, refactoring |
| `research` | research, analyze, study, hypothesis, experiment | 0.7 | Research tasks, creative exploration |
| `quick` | simple, brief, summarize, what is | 0.5 | Fast responses, simple questions |
| `default` | - | 0.5 | General-purpose reasoning |

**Override via environment variables**: `GOT_MODEL_CODE`, `GOT_MODEL_RESEARCH`, `GOT_MODEL_QUICK`, `GOT_MODEL`

### Multimodal Embeddings

The `embed-multimodal` tool generates embeddings for images using Voyage AI's multimodal model:

```json
{
  "image_base64": "iVBORw0KGgo...",
  "content_type": "image/png"
}
```

**Supported formats**: PNG, JPEG, GIF, WebP
**Model**: `voyage-multimodal-3` (1024 dimensions)
**Enable**: Set `MULTIMODAL_ENABLED=true` and `VOYAGE_API_KEY`

### Agentic Tool Calling

The `run-agent` tool executes complex multi-step reasoning tasks autonomously:

```json
{
  "task": "Analyze the tradeoffs between SQL and NoSQL for our use case",
  "max_iterations": 5
}
```

**Features**:
- Autonomous iteration until task completion or max iterations
- 30+ safe tools available (read-only, no side effects)
- Full execution trace with thoughts and tool calls
- Custom system prompts supported

**Safe Tools**: `think`, `decompose-problem`, `analyze-perspectives`, `detect-biases`, `build-causal-graph`, `generate-hypotheses`, `search-similar-thoughts`, etc.

**Excluded Tools**: `store-entity`, `run-agent`, `run-preset`, `create-checkpoint` (side effects or recursion risk)

**Enable**: Set `AGENT_ENABLED=true` and `ANTHROPIC_API_KEY`

## Storage Architecture

### SQLite (Default)
Persistent, write-through cache, FTS5 search, WAL mode, trajectory persistence (schema v6).

```bash
STORAGE_TYPE=sqlite  # Default
SQLITE_PATH=./data/unified-thinking.db
```

### In-Memory (Optional)
Thread-safe (sync.RWMutex), no persistence - use for testing only.

```bash
STORAGE_TYPE=memory
```

**Key Methods**: `StoreThought()`, `GetThought()`, `SearchThoughts()`, `StoreBranch()`, `GetBranch()`, `ListBranches()`

## Configuration

Claude Desktop config (`%APPDATA%\Claude\claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\path\\to\\bin\\unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "C:\\Users\\YourName\\AppData\\Roaming\\Claude\\unified-thinking.db",
        "EMBEDDINGS_ENABLED": "true",
        "VOYAGE_API_KEY": "your-key"
      }
    }
  }
}
```

### Environment Variables

**Required (fail-fast if missing):**

| Variable | Description |
|----------|-------------|
| `VOYAGE_API_KEY` | Voyage AI API key (REQUIRED for embeddings) |
| `ANTHROPIC_API_KEY` | Anthropic API key (REQUIRED for GoT, agent, web search) |
| `NEO4J_URI` | Neo4j connection URI (REQUIRED for knowledge graph) |
| `NEO4J_USERNAME` | Neo4j username |
| `NEO4J_PASSWORD` | Neo4j password |

**Storage:**

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_TYPE` | `sqlite` | `sqlite` (recommended) or `memory` |
| `SQLITE_PATH` | `./data/unified-thinking.db` | Database file path |

**Configuration:**

| Variable | Default | Description |
|----------|---------|-------------|
| `DEBUG` | `false` | Enable debug logging |
| `AUTO_VALIDATION_THRESHOLD` | `0.5` | Auto-validation confidence threshold |
| `RESPONSE_FORMAT` | `full` | Response format: `full`, `compact`, `minimal` |
| `EMBEDDINGS_MODEL` | `voyage-3-lite` | `voyage-3-lite` (512d), `voyage-3` (1024d), `voyage-3-large` (2048d) |
| `EMBEDDINGS_CACHE_MAX_ENTRIES` | `10000` | Max LRU cache entries (0 = unlimited) |
| `EMBEDDINGS_CACHE_PERSIST` | `false` | Persist embedding cache to disk |
| `EMBEDDINGS_CACHE_PATH` | - | Path for cache persistence |
| `RERANK_MODEL` | `rerank-2` | `rerank-2` (quality) or `rerank-2-lite` (speed) |
| `GOT_MODEL` | `claude-sonnet-4-5-20250929` | Default model for GoT and auto mode |
| `GOT_MODEL_CODE` | `claude-sonnet-4-5-20250929` | Model for code-related tasks (temperature: 0.3) |
| `GOT_MODEL_RESEARCH` | `claude-sonnet-4-5-20250929` | Model for research tasks (temperature: 0.7) |
| `GOT_MODEL_QUICK` | `claude-3-5-haiku-20241022` | Model for quick/simple tasks |
| `MULTIMODAL_MODEL` | `voyage-multimodal-3` | Voyage multimodal model (1024d) |
| `AGENT_MODEL` | `claude-sonnet-4-5-20250929` | Model for agentic tasks |
| `NEO4J_DATABASE` | `neo4j` | Neo4j database name |

## Data Flow

1. **Tool Call** → handler receives request
2. **Mode Selection** → auto-detect or explicit
3. **Processing** → mode's `ProcessThought()` executes
4. **Storage** → persisted via backend
5. **Validation** (optional) → logic check
6. **Response** → JSON to MCP client

## Implementation Details

### Builder Pattern
```go
thought := types.NewThought().
    Content("Example").
    Mode(types.ModeLinear).
    Confidence(0.9).
    Build()
```

### Mode Registry
Modes implement `ThinkingMode`: `Name()`, `CanHandle()`, `ProcessThought()`

### Branch Metrics
- Confidence = avg of thought confidences
- Priority = confidence + (insights × 0.1) + (cross_ref strengths × 0.1)

### Cross-Reference Types
`complementary`, `contradictory`, `builds_upon`, `alternative`

### Causal Reasoning
Pearl's framework: DAGs, graph surgery (do-calculus), counterfactuals, intervention simulation.

### Fallacy Detection
20+ types: formal (affirming consequent, circular), informal (ad hominem, straw man), evidence (hasty generalization), appeal (authority, emotion).

### LLM Client Architecture
Unified Anthropic API infrastructure using Go's embedding pattern:

```
AnthropicBaseClient (llm_base.go)
├── AnthropicLLMClient (llm_anthropic.go) - Graph-of-Thoughts scoring/generation
└── AgenticClient (llm_agentic.go) - Programmatic tool-calling loop
```

| File | Purpose |
|------|---------|
| `llm_types.go` | Unified API types: APIRequest, Message, ContentBlock, Tool, APIResponse |
| `llm_base.go` | Shared HTTP infrastructure: AnthropicBaseClient, SendRequest |
| `llm_anthropic.go` | AnthropicLLMClient with structured outputs, caching, domain-specific models |
| `llm_agentic.go` | AgenticClient with tool-calling loop, execution traces |
| `llm_client.go` | LLMClient interface definition |
| `llm_models.go` | Domain-specific model configuration (code/research/quick) |
| `llm_tools.go` | Structured output tool schemas |

## Migration from Old Servers

| Old Server | Old Tool | New Tool | Mode |
|------------|----------|----------|------|
| sequential-thinking | solve-problem | think | `linear` |
| branch-thinking | branch-thinking | think | `tree` |
| unreasonable-thinking | generate_unreasonable_thought | think | `divergent` |
| mcp-logic | prove | prove | - |

## Key Files

**Infrastructure**: `cmd/server/main.go`, `internal/types/types.go`, `internal/server/server.go`

**Storage**: `internal/storage/factory.go`, `memory.go`, `sqlite.go`, `sqlite_schema.go`

**Modes**: `internal/modes/auto.go`, `registry.go`, `linear.go`, `tree.go`, `divergent.go`, `graph*.go`

**LLM**: `internal/modes/llm_types.go`, `llm_base.go`, `llm_anthropic.go`, `llm_agentic.go`, `llm_client.go`, `llm_models.go`, `llm_tools.go`

**Reasoning**: `internal/reasoning/probabilistic.go`, `causal.go`, `decision.go`, `temporal.go`, `domain_templates.go`

**Testing**: `internal/testutil/mock_llm.go` (MockLLMClient for GoT testing without API)

**Dev Tools**: `build.ps1` (PowerShell), `scripts/pre-commit`, `scripts/install-hooks.{sh,ps1}`

**Validation**: `internal/validation/logic.go`, `fallacies.go`

## Code Style

- `go fmt` standard formatting
- Package-level godoc documentation
- Return errors, don't panic (except main.go fatal)
- `log.Println()` for debug (DEBUG=true)
- sync.RWMutex for thread safety

## Technical Constraints

- Go 1.24+
- Windows primary (Makefile defaults)
- MCP SDK v0.8.0
- modernc.org/sqlite (pure Go, no CGO)
- stdio transport only

---

For detailed change history, see [CHANGELOG.md](./CHANGELOG.md).
