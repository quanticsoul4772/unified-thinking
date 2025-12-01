# CLAUDE.md

Guidance for Claude Code when working with this repository.

## Project Overview

Go-based MCP server consolidating 5 TypeScript servers (sequential-thinking, branch-thinking, unreasonable-thinking-server, mcp-logic, state-coordinator) into one unified implementation with 75 cognitive reasoning tools.

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
| `internal/embeddings/` | Voyage AI semantic embeddings (voyage-3-lite, 512d) |
| `internal/contextbridge/` | Cross-session context retrieval with caching |
| `internal/server/` | MCP server with 75 tools, 23 handler modules |
| `internal/similarity/` | Thought similarity search via embeddings |

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

```bash
# Build
make build              # Windows: bin/unified-thinking.exe
make linux              # Linux: bin/unified-thinking

# Test
make test               # All tests
make test-coverage      # With coverage report
make test-race          # With race detector
go test -v ./internal/modes/  # Specific package

# Lint
golangci-lint run       # Run linter
golangci-lint run --fix # Auto-fix

# Clean
make clean              # Remove bin/
```

**DO NOT run server manually** - Claude Desktop starts it automatically.

**Test Coverage**: 84.3% overall, 102 test files, 1,300+ tests. Key: types (100%), metrics (100%), config (97.3%), reasoning (94.8%), modes (90.5%).

## MCP Tools (75 total)

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

### Graph-of-Thoughts (8)
`got-initialize`, `got-generate`, `got-aggregate`, `got-refine`, `got-score`, `got-prune`, `got-get-state`, `got-finalize`

## Storage Architecture

### In-Memory (Default)
Thread-safe (sync.RWMutex), no persistence, zero config.

### SQLite (Optional)
Persistent, write-through cache, FTS5 search, WAL mode, trajectory persistence (schema v6).

```bash
STORAGE_TYPE=sqlite
SQLITE_PATH=./data/thoughts.db
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

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_TYPE` | `memory` | `memory` or `sqlite` |
| `SQLITE_PATH` | - | Database file path |
| `DEBUG` | `false` | Enable debug logging |
| `AUTO_VALIDATION_THRESHOLD` | `0.5` | Auto-validation confidence threshold |
| `EMBEDDINGS_ENABLED` | `false` | Enable semantic embeddings |
| `VOYAGE_API_KEY` | - | Voyage AI API key |
| `EMBEDDINGS_MODEL` | `voyage-3-lite` | `voyage-3-lite` (512d), `voyage-3` (1024d), `voyage-3-large` (2048d) |
| `NEO4J_ENABLED` | `false` | Enable knowledge graph |
| `CONTEXT_BRIDGE_ENABLED` | `false` | Enable cross-session context |

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

**Reasoning**: `internal/reasoning/probabilistic.go`, `causal.go`, `decision.go`, `temporal.go`

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
