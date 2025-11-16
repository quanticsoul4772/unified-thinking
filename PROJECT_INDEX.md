# Unified Thinking Server - Project Index

## Project Overview

**Repository**: unified-thinking
**Language**: Go 1.23+
**Type**: Model Context Protocol (MCP) Server
**Purpose**: Consolidates multiple cognitive thinking patterns into a single efficient Go-based implementation
**MCP SDK**: `github.com/modelcontextprotocol/go-sdk` v0.8.0

## Test Coverage

**Total Test Count**: 415 tests
**Pass Rate**: 100% (415/415)
**Coverage by Package**:
- `internal/types`: 100.0%
- `internal/metrics`: 100.0%
- `internal/validation`: 91.2%
- `internal/metacognition`: 90.2%
- `internal/analysis`: 89.3%
- `internal/orchestration`: 87.2%
- `internal/reasoning`: 86.5%
- `internal/processing`: 83.3%
- `internal/integration`: 82.2%
- `internal/storage`: 81.0%
- `internal/modes`: 77.8%
- `internal/config`: 64.5%
- `internal/server/handlers`: 51.6%
- `internal/server`: 45.5%
- `cmd/server`: 25.0%
- **Overall**: ~75% test coverage

## Project Structure

```
unified-thinking/
├── cmd/
│   └── server/
│       ├── main.go                    # Entry point, server initialization
│       └── main_test.go               # Server initialization tests
│
├── internal/
│   ├── analysis/                      # Evidence, contradiction, perspective analysis
│   │   ├── argument.go                # Argument structure analysis
│   │   ├── argument_test.go           # Argument analysis tests
│   │   ├── contradiction.go           # Contradiction detection logic
│   │   ├── contradiction_edge_test.go # Edge case tests for contradiction detection
│   │   ├── concurrent_test.go         # Concurrency tests
│   │   ├── evidence.go                # Evidence quality assessment
│   │   ├── evidence_test.go           # Evidence assessment tests
│   │   ├── perspective.go             # Multi-stakeholder perspective analysis
│   │   ├── perspective_test.go        # Perspective analysis tests
│   │   └── sensitivity.go             # Sensitivity analysis implementation
│   │
│   ├── config/                        # Configuration management
│   │   └── config.go                  # Configuration structures and loading
│   │
│   ├── integration/                   # Cross-mode synthesis and integration
│   │   ├── causal_temporal.go         # Causal-temporal integration
│   │   ├── causal_temporal_test.go    # Causal-temporal integration tests
│   │   ├── edge_case_test.go          # Edge case tests for integration
│   │   ├── evidence_pipeline.go       # Evidence processing pipeline
│   │   ├── evidence_pipeline_test.go  # Evidence pipeline tests
│   │   ├── probabilistic_causal.go    # Probabilistic-causal integration
│   │   ├── probabilistic_causal_test.go # Probabilistic-causal tests
│   │   ├── synthesizer.go             # Cross-mode insight synthesis
│   │   └── synthesizer_test.go        # Synthesizer tests
│   │
│   ├── metacognition/                 # Self-evaluation, bias detection
│   │   ├── bias_detection.go          # 15+ cognitive bias detection
│   │   ├── bias_detection_test.go     # Bias detection tests
│   │   ├── concurrent_test.go         # Concurrency tests
│   │   ├── self_eval.go               # Self-evaluation logic
│   │   ├── self_eval_test.go          # Self-evaluation tests
│   │   ├── unknown_unknowns.go        # Unknown unknowns detection
│   │   └── unknown_unknowns_test.go   # Unknown unknowns tests
│   │
│   ├── metrics/                       # System metrics and monitoring
│   │   ├── collector.go               # Metrics collection
│   │   └── collector_test.go          # Metrics tests
│   │
│   ├── modes/                         # Thinking mode implementations
│   │   ├── auto.go                    # Automatic mode selection
│   │   ├── auto_test.go               # Auto mode tests
│   │   ├── backtracking.go            # Checkpoint-based reasoning
│   │   ├── backtracking_test.go       # Backtracking tests
│   │   ├── divergent.go               # Creative/unconventional ideation
│   │   ├── divergent_test.go          # Divergent mode tests
│   │   ├── error_handling_test.go     # Error handling tests
│   │   ├── error_recovery_test.go     # Error recovery tests
│   │   ├── linear.go                  # Sequential step-by-step reasoning
│   │   ├── linear_test.go             # Linear mode tests
│   │   ├── reflection.go              # Metacognitive reflection
│   │   ├── reflection_test.go         # Reflection mode tests
│   │   ├── registry.go                # Dynamic mode registration
│   │   ├── registry_test.go           # Registry tests
│   │   ├── shared.go                  # Shared types and interfaces
│   │   ├── tree.go                    # Multi-branch parallel exploration
│   │   └── tree_test.go               # Tree mode tests
│   │
│   ├── orchestration/                 # Workflow automation and coordination
│   │   ├── context_test.go            # Context management tests
│   │   ├── error_recovery_test.go     # Error recovery tests
│   │   ├── helpers_test.go            # Helper function tests
│   │   ├── interface.go               # ToolExecutor abstraction
│   │   ├── workflow_execution_test.go # Workflow execution tests
│   │   └── workflow_test.go           # Workflow structure tests
│   │
│   ├── processing/                    # Dual-process reasoning
│   │   ├── dual_process.go            # System 1/2 executor
│   │   └── dual_process_test.go       # Dual-process tests
│   │
│   ├── reasoning/                     # Advanced reasoning capabilities
│   │   ├── abductive.go               # Hypothesis generation
│   │   ├── abductive_test.go          # Abductive reasoning tests
│   │   ├── analogical.go              # Analogical reasoning
│   │   ├── analogical_test.go         # Analogical reasoning tests
│   │   ├── case_based.go              # Case-based reasoning (CBR)
│   │   ├── causal.go                  # Pearl's causal inference framework
│   │   ├── causal_test.go             # Causal reasoning tests
│   │   ├── concurrent_test.go         # Concurrency tests
│   │   ├── decision.go                # Multi-criteria decision analysis
│   │   ├── probabilistic.go           # Bayesian inference
│   │   ├── probabilistic_test.go      # Probabilistic reasoning tests
│   │   ├── temporal.go                # Temporal analysis
│   │   └── temporal_test.go           # Temporal reasoning tests
│   │
│   ├── server/                        # MCP server implementation
│   │   ├── error_recovery_test.go     # Error recovery tests
│   │   ├── executor.go                # Tool executor for orchestration
│   │   ├── executor_test.go           # Executor tests
│   │   ├── formatters.go              # Response formatters
│   │   ├── integration_patterns.go    # Integration pattern definitions
│   │   ├── server_init_test.go        # Server initialization tests
│   │   ├── server_test.go             # Server core tests
│   │   ├── validation_test.go         # Validation tests
│   │   ├── handlers/                  # Specialized tool handlers
│   │   │   ├── abductive.go           # Abductive reasoning handler
│   │   │   ├── abductive_test.go      # Abductive handler tests
│   │   │   ├── backtracking.go        # Backtracking handler
│   │   │   ├── branches.go            # Branch management handler
│   │   │   ├── calibration.go         # Calibration tracking handler
│   │   │   ├── calibration_test.go    # Calibration tests
│   │   │   ├── case_based.go          # CBR handler
│   │   │   ├── dual_process.go        # Dual-process handler
│   │   │   ├── enhanced.go            # Enhanced validation handler
│   │   │   ├── enhanced_validation_test.go # Enhanced validation tests
│   │   │   ├── hallucination.go       # Hallucination detection handler
│   │   │   ├── hallucination_test.go  # Hallucination tests
│   │   │   ├── helpers.go             # Shared helper functions
│   │   │   ├── helpers_test.go        # Helper function tests
│   │   │   ├── metadata.go            # Metadata management handler
│   │   │   ├── metadata_test.go       # Metadata tests
│   │   │   ├── search.go              # Search handler
│   │   │   ├── symbolic.go            # Symbolic reasoning handler
│   │   │   ├── thinking.go            # Main thinking handler
│   │   │   ├── unknown_unknowns.go    # Unknown unknowns handler
│   │   │   ├── validation.go          # Validation handler
│   │   │   └── validation_test.go     # Validation handler tests
│   │   └── types/                     # Server-specific types
│   │       ├── common.go              # Common server types
│   │       └── handlers.go            # Handler-specific types
│   │
│   ├── storage/                       # Pluggable storage layer
│   │   ├── config.go                  # Storage configuration
│   │   ├── config_test.go             # Configuration tests
│   │   ├── copy.go                    # Deep copy utilities
│   │   ├── error_recovery_test.go     # Error recovery tests
│   │   ├── factory.go                 # Storage factory pattern
│   │   ├── factory_test.go            # Factory tests
│   │   ├── interface.go               # Storage interface definition
│   │   ├── memory.go                  # In-memory implementation
│   │   ├── memory_test.go             # Memory storage tests
│   │   ├── optimization_test.go       # Optimization tests
│   │   ├── sqlite.go                  # SQLite implementation
│   │   ├── sqlite_schema.go           # SQLite schema definitions
│   │   └── sqlite_test.go             # SQLite storage tests
│   │
│   ├── types/                         # Core data structures
│   │   ├── builders.go                # Builder patterns for object construction
│   │   ├── builders_test.go           # Builder tests
│   │   ├── types.go                   # 50+ core types
│   │   └── types_test.go              # Type tests
│   │
│   └── validation/                    # Logical validation and proof
│       ├── calibration.go             # Confidence calibration
│       ├── calibration_test.go        # Calibration tests
│       ├── fallacies.go               # 20+ fallacy detection
│       ├── fallacies_test.go          # Fallacy detection tests
│       ├── hallucination.go           # Hallucination detection
│       ├── hallucination_test.go      # Hallucination tests
│       ├── logic.go                   # Logical validation
│       ├── logic_test.go              # Logic tests
│       ├── symbolic.go                # Symbolic constraint solving
│       └── symbolic_test.go           # Symbolic tests
│
├── .golangci.yml                      # Linter configuration
├── AGENTS.md                          # Agent coordination documentation
├── CLAUDE.md                          # Claude Code specific instructions
├── coverage_func.txt                  # Coverage function output
├── go.mod                             # Go module dependencies
├── go.sum                             # Go module checksums
├── Makefile                           # Build automation
├── PROJECT_INDEX.md                   # This file
└── README.md                          # Project documentation
```

## Core Components

### 1. Thinking Modes (`internal/modes/`)

Six cognitive thinking patterns with automatic selection:

| Mode | File | Purpose | Triggers |
|------|------|---------|----------|
| **Linear** | `linear.go` | Sequential step-by-step reasoning | Default fallback mode |
| **Tree** | `tree.go` | Multi-branch parallel exploration | `branch_id`, `cross_refs`, "branch", "explore" |
| **Divergent** | `divergent.go` | Creative/unconventional ideation | "creative", "what if", "challenge", "rebel" |
| **Reflection** | `reflection.go` | Metacognitive reflection on previous reasoning | "reflect", "review", "what did we learn" |
| **Backtracking** | `backtracking.go` | Checkpoint-based reasoning with restore capability | "checkpoint", "restore", "go back" |
| **Auto** | `auto.go` | Automatic mode selection | Analyzes content and selects appropriate mode |

### 2. Storage Layer (`internal/storage/`)

Pluggable storage architecture with two implementations:

- **In-Memory** (`memory.go`): Thread-safe with sync.RWMutex, no persistence, fast
- **SQLite** (`sqlite.go`): Persistent across restarts, write-through caching, FTS5 full-text search, WAL mode

Configuration via environment variables: `STORAGE_TYPE`, `SQLITE_PATH`, `SQLITE_TIMEOUT`, `STORAGE_FALLBACK`

### 3. Reasoning Engines (`internal/reasoning/`)

Advanced reasoning capabilities:

- **Probabilistic** (`probabilistic.go`): Bayesian inference, prior/posterior updates
- **Causal** (`causal.go`): Pearl's causal framework, do-calculus, counterfactuals
- **Temporal** (`temporal.go`): Short-term vs long-term analysis
- **Abductive** (`abductive.go`): Hypothesis generation and evaluation
- **Case-Based** (`case_based.go`): Similarity retrieval and adaptation
- **Decision** (`decision.go`): Multi-criteria weighted scoring

### 4. Analysis Tools (`internal/analysis/`)

- **Evidence** (`evidence.go`): Quality assessment with reliability scoring
- **Contradiction** (`contradiction.go`): Detection across thoughts with subject-predicate analysis
- **Perspective** (`perspective.go`): Multi-stakeholder viewpoint analysis
- **Sensitivity** (`sensitivity.go`): Robustness testing for conclusions

### 5. Metacognition (`internal/metacognition/`)

Self-awareness and quality assessment:

- **Self-Evaluation** (`self_eval.go`): Thought quality, completeness, coherence assessment
- **Bias Detection** (`bias_detection.go`): 15+ cognitive biases and logical fallacies
- **Unknown Unknowns** (`unknown_unknowns.go`): Knowledge gap and blind spot identification

### 6. Validation (`internal/validation/`)

Logic and proof verification:

- **Logic** (`logic.go`): Consistency checking, formal proof attempts
- **Fallacies** (`fallacies.go`): 20+ logical fallacy detection
- **Symbolic** (`symbolic.go`): Constraint satisfaction solving
- **Calibration** (`calibration.go`): Confidence calibration tracking
- **Hallucination** (`hallucination.go`): Semantic uncertainty measurement

### 7. Integration (`internal/integration/`)

Cross-mode synthesis:

- **Synthesizer** (`synthesizer.go`): Multi-mode insight integration
- **Causal-Temporal** (`causal_temporal.go`): Causal reasoning + temporal analysis
- **Probabilistic-Causal** (`probabilistic_causal.go`): Bayesian + causal integration
- **Evidence Pipeline** (`evidence_pipeline.go`): Evidence processing workflows

### 8. Orchestration (`internal/orchestration/`)

Automated multi-tool workflows:

- Workflow definition and registration
- Sequential and parallel execution
- Conditional step execution
- ReasoningContext for shared state
- ToolExecutor abstraction for tool invocation

## MCP Tools (50 Total)

### Core Thinking Tools (11)
1. `think` - Main thinking tool with mode selection
2. `history` - View thinking history
3. `list-branches` - List all branches
4. `focus-branch` - Switch active branch
5. `branch-history` - Get detailed branch history
6. `recent-branches` - Get recently accessed branches
7. `validate` - Validate thought logical consistency
8. `prove` - Attempt formal proof
9. `check-syntax` - Validate logical statement syntax
10. `search` - Search thoughts
11. `get-metrics` - System performance metrics

### Probabilistic Reasoning (4)
12. `probabilistic-reasoning` - Bayesian inference
13. `assess-evidence` - Evidence quality assessment
14. `detect-contradictions` - Find contradictions
15. `sensitivity-analysis` - Robustness testing

### Decision & Problem-Solving (3)
16. `make-decision` - Multi-criteria decision analysis
17. `decompose-problem` - Break down complex problems
18. `verify-thought` - Verify thought validity

### Metacognition (3)
19. `self-evaluate` - Metacognitive self-assessment
20. `detect-biases` - Identify cognitive biases and fallacies
21. `detect-blind-spots` - Identify unknown unknowns

### Hallucination & Calibration (4)
22. `get-hallucination-report` - Retrieve hallucination reports
23. `record-prediction` - Record prediction for calibration
24. `record-outcome` - Record prediction outcomes
25. `get-calibration-report` - Retrieve calibration analysis

### Perspective & Temporal Analysis (4)
26. `analyze-perspectives` - Multi-stakeholder analysis
27. `analyze-temporal` - Short-term vs long-term
28. `compare-time-horizons` - Compare across time horizons
29. `identify-optimal-timing` - Determine optimal timing

### Causal Reasoning (5)
30. `build-causal-graph` - Construct causal graphs
31. `simulate-intervention` - Simulate interventions with do-calculus
32. `generate-counterfactual` - Generate "what if" scenarios
33. `analyze-correlation-vs-causation` - Distinguish correlation from causation
34. `get-causal-graph` - Retrieve causal graph

### Integration & Orchestration (6)
35. `synthesize-insights` - Synthesize insights from multiple modes
36. `detect-emergent-patterns` - Detect patterns across modes
37. `execute-workflow` - Execute predefined workflows
38. `list-workflows` - List available workflows
39. `register-workflow` - Register custom workflows
40. `list-integration-patterns` - Discover integration patterns

### Dual-Process Reasoning (1)
41. `dual-process-think` - System 1 (fast) vs System 2 (deliberate)

### Backtracking (3)
42. `create-checkpoint` - Create reasoning checkpoint
43. `restore-checkpoint` - Restore from checkpoint
44. `list-checkpoints` - List available checkpoints

### Abductive Reasoning (2)
45. `generate-hypotheses` - Generate explanatory hypotheses
46. `evaluate-hypotheses` - Evaluate hypothesis plausibility

### Case-Based Reasoning (2)
47. `retrieve-similar-cases` - Retrieve similar cases
48. `perform-cbr-cycle` - Execute full CBR cycle

### Symbolic Reasoning (2)
49. `prove-theorem` - Formal theorem proving
50. `check-constraints` - Check constraint satisfaction

## Key Data Structures (`internal/types/types.go`)

### Core Types
- `Thought` - Individual reasoning unit with content, mode, confidence
- `Branch` - Collection of thoughts in tree mode with insights and cross-refs
- `Insight` - Key realization or breakthrough
- `CrossRef` - Link between branches with typed relationships
- `Validation` - Logical validation result
- `Relationship` - Relationship between thoughts
- `ThinkingMode` - Enum for thinking modes (Linear, Tree, Divergent, etc.)

### Reasoning Types (50+)
- `CausalGraph`, `CausalVariable`, `CausalLink`, `CausalIntervention`
- `ProbabilisticBelief`, `EvidencePiece`, `BeliefUpdate`
- `TemporalAnalysis`, `PerspectiveAnalysis`, `StakeholderView`
- `Decision`, `DecisionOption`, `DecisionCriterion`
- `Hypothesis`, `HypothesisEvaluation`
- `Case`, `CaseSolution`, `CaseAdaptation`
- `Checkpoint`, `BacktrackingState`
- `HallucinationReport`, `CalibrationReport`
- And 30+ more specialized types

### Builder Patterns (`internal/types/builders.go`)

Fluent API for object construction:
```go
thought := NewThought().
    Content("Example").
    Mode(ModeLinear).
    Confidence(0.9).
    Build()
```

## Development Workflow

### Build Commands
```bash
make build          # Build Windows binary (bin/unified-thinking.exe)
make linux          # Build Linux binary
make install-deps   # Install dependencies
make clean          # Remove build artifacts
```

### Testing Commands
```bash
make test           # Run all tests
make test-coverage  # Generate coverage report
make test-race      # Run with race detector
make test-short     # Run short tests only
make test-storage   # Test storage layer only
make benchmark      # Run benchmarks
make test-all       # Comprehensive test suite
```

### Code Quality
```bash
golangci-lint run   # Run all linters
golangci-lint run --fix  # Auto-fix issues
```

## Recent Improvements (Commit 876aa00)

### Test Suite Expansion
- Added 19 new test files across all major components
- Increased test count from ~250 to 415 tests
- Achieved 100% pass rate
- Test coverage improvements:
  - `internal/types`: 100.0%
  - `internal/metrics`: 100.0%
  - `internal/validation`: 91.2%
  - `internal/metacognition`: 90.2%

### Implementation Fixes
1. **Contradiction Detection** (`internal/analysis/contradiction.go`):
   - Enhanced case-insensitive subject extraction
   - Added compound noun phrase detection
   - Implemented context-aware subject comparison (`isSameSubjectContext`)
   - Added predicate similarity checking (`hasSimilarPredicate`)

2. **Causal-Temporal Integration** (`internal/integration/causal_temporal.go`):
   - Fixed `generateTemporalRecommendation` - added "immediate" and "delayed" patterns
   - Fixed `identifyTimeSensitiveVariables` - lowered threshold to ≥2 links
   - Fixed `determineTimingWindows` - added empty list and pattern validation
   - Fixed `synthesizeTimingRecommendation` - complete rewrite with pattern detection

3. **Server Initialization** (`cmd/server/main.go`):
   - Added nil orchestrator guard in `registerPredefinedWorkflows`

### Test Files Added
- `cmd/server/main_test.go` - Server initialization tests
- `internal/analysis/argument_test.go` - Argument analysis tests
- `internal/analysis/contradiction_edge_test.go` - Edge case tests
- `internal/integration/causal_temporal_test.go` - Causal-temporal tests
- `internal/integration/edge_case_test.go` - Integration edge cases
- `internal/integration/evidence_pipeline_test.go` - Evidence pipeline tests
- `internal/metrics/collector_test.go` - Metrics collection tests
- `internal/modes/error_recovery_test.go` - Mode error recovery tests
- `internal/orchestration/error_recovery_test.go` - Orchestration error recovery
- `internal/server/error_recovery_test.go` - Server error recovery
- `internal/server/executor_test.go` - Tool executor tests
- `internal/server/handlers/abductive_test.go` - Abductive handler tests
- `internal/server/handlers/calibration_test.go` - Calibration tests
- `internal/server/handlers/enhanced_validation_test.go` - Enhanced validation tests
- `internal/server/handlers/hallucination_test.go` - Hallucination detection tests
- `internal/server/handlers/helpers_test.go` - Helper function tests
- `internal/server/handlers/metadata_test.go` - Metadata handler tests
- `internal/server/server_init_test.go` - Server initialization tests
- `internal/storage/error_recovery_test.go` - Storage error recovery tests

## Performance Characteristics

- **In-Memory Storage**: O(1) access, O(n) search with substring matching
- **SQLite Storage**: O(log n) indexed access, O(1) FTS5 full-text search
- **Thinking Modes**: O(1) mode selection, O(n) thought processing
- **Validation**: O(n²) contradiction detection, O(n) logical validation
- **Causal Reasoning**: O(V+E) graph traversal, O(n) intervention simulation
- **Thread Safety**: RWMutex for concurrent read/write operations
- **Resource Limits**: MaxSearchResults=1000, MaxIndexSize=100000

## Configuration

### Environment Variables
- `STORAGE_TYPE`: `memory` (default) or `sqlite`
- `SQLITE_PATH`: Path to SQLite database file
- `SQLITE_TIMEOUT`: Connection timeout in milliseconds (default: 5000)
- `STORAGE_FALLBACK`: Fallback storage type (e.g., `memory`)
- `DEBUG`: Enable debug logging (`true` or `false`)
- `AUTO_VALIDATION_THRESHOLD`: Confidence threshold for auto-validation (default: 0.5)

### Claude Desktop Configuration

Basic (In-Memory):
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\path\\to\\unified-thinking.exe",
      "transport": "stdio",
      "env": {"DEBUG": "true"}
    }
  }
}
```

With SQLite Persistence:
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\path\\to\\unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "C:\\Users\\You\\unified-thinking.db",
        "STORAGE_FALLBACK": "memory"
      }
    }
  }
}
```

## Dependencies

### Core Dependencies
- `github.com/modelcontextprotocol/go-sdk` v0.8.0 - MCP protocol implementation
- `modernc.org/sqlite` - Pure Go SQLite (no CGO)

### Development Dependencies
- `golangci-lint` - Code linting and quality checks
- `go test` - Testing framework
- `go cover` - Coverage analysis

## Migration Notes

This server replaces 5 separate TypeScript servers:
- `sequential-thinking` → `unified-thinking` (mode: "linear")
- `branch-thinking` → `unified-thinking` (mode: "tree")
- `unreasonable-thinking-server` → `unified-thinking` (mode: "divergent", `force_rebellion: true`)
- `mcp-logic` → `unified-thinking` (`prove`, `check-syntax` tools)
- `state-coordinator` → Built into storage layer

## Future Enhancements

Potential areas for expansion:
- [ ] Distributed storage backend (Redis, PostgreSQL)
- [ ] Real-time collaboration features
- [ ] Advanced visualization exports
- [ ] Machine learning integration for pattern recognition
- [ ] Multi-language support beyond Go
- [ ] GraphQL API for external integrations
- [ ] Enhanced workflow debugging tools
- [ ] Performance profiling and optimization

## License

MIT License

## Contributors

See git commit history for detailed contribution records.

---

**Last Updated**: November 2025
**Project Status**: Active Development
**Maintainer**: Project Team
