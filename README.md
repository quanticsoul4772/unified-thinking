# unified thinking server

a model context protocol (mcp) server that consolidates multiple cognitive thinking patterns into a single, efficient go-based implementation.

## features

### thinking modes

- **linear mode**: sequential step-by-step reasoning for systematic problem solving
- **tree mode**: multi-branch parallel exploration with insights and cross-references
- **divergent mode**: creative/unconventional ideation with "rebellion" capability
- **reflection mode**: metacognitive reflection on previous reasoning with insight extraction
- **backtracking mode**: checkpoint-based reasoning with ability to restore previous states
- **auto mode**: automatic mode selection based on input content

### core capabilities

- multi-mode thinking (linear, tree, divergent, reflection, backtracking, auto)
- branch management and exploration
- insight generation and tracking
- cross-reference support between branches
- logical validation and consistency checking
- formal proof attempts with universal instantiation
- syntax validation for logical statements
- search across all thoughts with full-text indexing
- full history tracking with branch ancestry
- checkpoint creation and restoration
- optional persistence with sqlite (in-memory by default)

### advanced cognitive reasoning

the server includes 62 specialized tools across 10 major categories:

#### probabilistic reasoning (4 tools)
- bayesian inference with prior and posterior belief updates
- evidence-based probability updates with quality assessment
- belief combination operations (and/or logic)
- confidence estimation from evidence aggregation

#### causal reasoning (5 tools)
- causal graph construction from observations
- intervention simulation using do-calculus
- counterfactual scenario generation
- correlation vs causation analysis
- graph retrieval and inspection

#### decision support (3 tools)
- multi-criteria decision analysis with weighted scoring
- problem decomposition into manageable subproblems
- sensitivity analysis for robustness testing

#### advanced reasoning modes (9 tools)
- dual-process reasoning (system 1 fast intuition vs system 2 deliberate analysis)
- abductive reasoning (hypothesis generation and evaluation)
- case-based reasoning (similarity retrieval and adaptation)
- symbolic constraint solving with inequality detection

#### metacognition (3 tools)
- self-evaluation of thought quality, completeness, and coherence
- cognitive bias detection (confirmation, anchoring, availability, sunk cost, overconfidence, recency, groupthink)
- unknown unknowns detection (identifying knowledge gaps and blind spots)

#### analysis tools (8 tools)
- contradiction detection across thoughts
- multi-perspective stakeholder analysis
- temporal reasoning (short-term vs long-term implications)
- evidence quality assessment with reliability scoring
- hallucination detection and reporting
- confidence calibration tracking

#### integration & orchestration (6 tools)
- cross-mode insight synthesis
- emergent pattern detection across reasoning modes
- workflow execution for automated multi-tool pipelines
- custom workflow registration
- integration pattern discovery

#### validation & logic (4 tools)
- logical consistency validation
- formal proof attempts (modus ponens, modus tollens, universal instantiation, etc.)
- logical statement syntax checking
- theorem proving with constraint checking

#### enhanced tools (8 tools)
- cross-domain analogical reasoning (find and apply analogies)
- argument decomposition (premises, claims, assumptions, inference chains)
- counter-argument generation using multiple strategies
- fallacy detection (formal and informal logical fallacies)
- evidence pipeline integration (auto-update beliefs/graphs/decisions)
- temporal-causal effect analysis (short/medium/long-term impacts)
- decision timing optimization based on causal factors

#### episodic memory & learning (4 tools)
- reasoning session tracking (build complete trajectory from problem to solution)
- pattern learning (recognizes successful approaches across sessions)
- adaptive recommendations (suggests proven strategies based on similar past problems)
- trajectory search (learn from historical successes and failures)

## installation

### prerequisites

- go 1.23 or higher
- git

### build

```bash
go mod download
go build -o bin/unified-thinking.exe ./cmd/server
```

or using make:

```bash
make build
```

## documentation

comprehensive documentation is available:

- **[API Reference](API_REFERENCE.md)** - complete reference for all 50 mcp tools with parameters, examples, and response formats
- **[Architecture Diagrams](ARCHITECTURE.md)** - visual representations of system architecture, data flows, and component interactions
- **[Project Index](PROJECT_INDEX.md)** - comprehensive project structure, architecture, and component overview
- **[MCP Integration Test Report](MCP_INTEGRATION_TEST_REPORT.md)** - end-to-end validation results and production readiness certification
- **[Architecture Guide](CLAUDE.md)** - detailed technical architecture and implementation guide

## configuration

### basic configuration (in-memory)

add to your claude desktop config (`%appdata%\claude\claude_desktop_config.json` on windows):

```json
{
  "mcpservers": {
    "unified-thinking": {
      "command": "/path/to/unified-thinking/bin/unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "debug": "true"
      }
    }
  }
}
```

### configuration with sqlite persistence

for persistent storage across sessions:

```json
{
  "mcpservers": {
    "unified-thinking": {
      "command": "/path/to/unified-thinking/bin/unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "debug": "true",
        "storage_type": "sqlite",
        "sqlite_path": "c:\\users\\yourname\\appdata\\roaming\\claude\\unified-thinking.db",
        "storage_fallback": "memory"
      }
    }
  }
}
```

**environment variables**:
- `storage_type`: `memory` (default) or `sqlite`
- `sqlite_path`: path to sqlite database file (created automatically)
- `sqlite_timeout`: connection timeout in milliseconds (default: 5000)
- `storage_fallback`: storage to use if primary fails (e.g., `memory`)
- `debug`: enable debug logging (`true` or `false`)
- `auto_validation_threshold`: confidence threshold for auto-validation (default: 0.5)

after saving the config, restart claude desktop.

## available tools (50 total)

### core thinking tools (11 tools)

1. **think** - main thinking tool
   ```json
   {
     "content": "your thinking prompt",
     "mode": "auto|linear|tree|divergent|reflection|backtracking",
     "confidence": 0.8,
     "key_points": ["point1", "point2"],
     "require_validation": true
   }
   ```

2. **history** - view thinking history
3. **list-branches** - list all branches
4. **focus-branch** - switch active branch
5. **branch-history** - get detailed branch history
6. **recent-branches** - get recently accessed branches
7. **validate** - validate thought logical consistency
8. **prove** - attempt to prove a logical conclusion
9. **check-syntax** - validate logical statement syntax
10. **search** - search thoughts
11. **get-metrics** - system performance and usage metrics

### probabilistic reasoning tools (4 tools)

12. **probabilistic-reasoning** - bayesian inference and belief updates
    ```json
    {
      "operation": "create|update|get|combine",
      "statement": "it will rain tomorrow",
      "prior_prob": 0.3
    }
    ```

13. **assess-evidence** - evidence quality assessment
14. **detect-contradictions** - find contradictions among thoughts
15. **sensitivity-analysis** - test robustness of conclusions

### decision & problem-solving tools (3 tools)

16. **make-decision** - multi-criteria decision analysis
    ```json
    {
      "question": "which option should we choose?",
      "options": [{"name": "option a", "scores": {"cost": 0.8}}],
      "criteria": [{"name": "cost", "weight": 0.6, "maximize": false}]
    }
    ```

17. **decompose-problem** - break down complex problems
18. **verify-thought** - verify thought validity and structure

### metacognition tools (3 tools)

19. **self-evaluate** - metacognitive self-assessment
20. **detect-biases** - identify cognitive biases and logical fallacies
21. **detect-blind-spots** - identify unknown unknowns and knowledge gaps

### hallucination & calibration tools (4 tools)

22. **get-hallucination-report** - retrieve hallucination detection reports
23. **record-prediction** - record a prediction for calibration tracking
24. **record-outcome** - record prediction outcomes
25. **get-calibration-report** - retrieve confidence calibration analysis

### perspective & temporal analysis tools (4 tools)

26. **analyze-perspectives** - multi-stakeholder perspective analysis
27. **analyze-temporal** - short-term vs long-term implications
28. **compare-time-horizons** - compare across time horizons
29. **identify-optimal-timing** - determine optimal decision timing

### causal reasoning tools (5 tools)

30. **build-causal-graph** - construct causal graphs from observations
31. **simulate-intervention** - simulate interventions with do-calculus
32. **generate-counterfactual** - generate "what if" scenarios
33. **analyze-correlation-vs-causation** - distinguish correlation from causation
34. **get-causal-graph** - retrieve previously built causal graph

### integration & orchestration tools (6 tools)

35. **synthesize-insights** - synthesize insights from multiple modes
36. **detect-emergent-patterns** - detect patterns across reasoning modes
37. **execute-workflow** - execute predefined multi-tool workflows
38. **list-workflows** - list available automated workflows
39. **register-workflow** - register custom workflows
40. **list-integration-patterns** - discover integration patterns

### dual-process reasoning tools (1 tool)

41. **dual-process-think** - system 1 (fast) vs system 2 (deliberate) reasoning
    ```json
    {
      "content": "problem to analyze",
      "force_system": "system1|system2"
    }
    ```

### backtracking tools (3 tools)

42. **create-checkpoint** - create reasoning checkpoint
43. **restore-checkpoint** - restore from checkpoint
44. **list-checkpoints** - list available checkpoints

### abductive reasoning tools (2 tools)

45. **generate-hypotheses** - generate explanatory hypotheses
46. **evaluate-hypotheses** - evaluate hypothesis plausibility

### case-based reasoning tools (2 tools)

47. **retrieve-similar-cases** - retrieve similar cases from memory
48. **perform-cbr-cycle** - execute full cbr cycle (retrieve, reuse, revise, retain)

### symbolic reasoning tools (2 tools)

49. **prove-theorem** - formal theorem proving
50. **check-constraints** - check symbolic constraint satisfaction

## architecture

```
unified-thinking/
├── cmd/server/             # main entry point
├── internal/
│   ├── types/              # core data structures (50+ types)
│   ├── storage/            # pluggable storage (memory/sqlite)
│   │   ├── memory.go       # in-memory implementation
│   │   ├── sqlite.go       # sqlite with write-through cache
│   │   ├── factory.go      # storage factory pattern
│   │   └── config.go       # configuration management
│   ├── modes/              # thinking mode implementations
│   │   ├── linear.go       # sequential reasoning
│   │   ├── tree.go         # parallel exploration
│   │   ├── divergent.go    # creative ideation
│   │   ├── reflection.go   # metacognitive reflection
│   │   ├── backtracking.go # checkpoint-based reasoning
│   │   └── auto.go         # automatic mode selection
│   ├── processing/         # dual-process reasoning
│   │   └── dual_process.go # system 1/2 executor
│   ├── reasoning/          # probabilistic, causal, temporal
│   │   ├── probabilistic.go    # bayesian inference
│   │   ├── causal.go           # pearl's causal framework
│   │   ├── temporal.go         # temporal analysis
│   │   ├── abductive.go        # hypothesis generation
│   │   └── case_based.go       # cbr implementation
│   ├── analysis/           # evidence, contradiction, perspective
│   ├── metacognition/      # self-eval, bias detection, unknown unknowns
│   ├── validation/         # logic validation, fallacy detection, symbolic
│   ├── integration/        # cross-mode synthesis
│   ├── orchestration/      # workflow automation
│   └── server/             # mcp server implementation
│       └── handlers/       # specialized tool handlers (19 files)
```

### cognitive architecture

the server implements a modular cognitive architecture with specialized packages:

- **modes**: six thinking modes (linear, tree, divergent, reflection, backtracking, auto)
- **processing**: dual-process reasoning (fast intuitive vs slow deliberate)
- **reasoning**: probabilistic inference, causal analysis, temporal reasoning, abductive inference, case-based reasoning
- **analysis**: evidence assessment, contradiction detection, perspective analysis, sensitivity testing
- **metacognition**: self-evaluation, bias detection, unknown unknowns identification
- **validation**: logical validation, fallacy detection, symbolic constraint solving
- **integration**: cross-mode synthesis and emergent pattern detection
- **orchestration**: automated multi-tool workflow execution

all components are thread-safe, composable, and maintain backward compatibility.

## development

### build

```bash
# build the server binary
make build

# clean build artifacts
make clean
```

### testing

```bash
# run all tests
make test

# run with verbose output
go test -v ./...

# run with coverage
make test-coverage

# run benchmarks
make benchmark
```

**test coverage**: ~75% overall | 415 tests | 100% pass rate

### coverage by package

| package | coverage | status |
|---------|----------|--------|
| `internal/types` | 100.0% | ✅ excellent |
| `internal/metrics` | 100.0% | ✅ excellent |
| `internal/validation` | 91.2% | ✅ excellent |
| `internal/metacognition` | 90.2% | ✅ excellent |
| `internal/analysis` | 89.3% | ✅ excellent |
| `internal/orchestration` | 87.2% | ✅ excellent |
| `internal/reasoning` | 86.5% | ✅ excellent |
| `internal/processing` | 83.3% | ✅ excellent |
| `internal/integration` | 82.2% | ✅ excellent |
| **`internal/server/handlers`** | **81.2%** | **✅ improved (+29.6%)** |
| `internal/storage` | 81.0% | ✅ excellent |
| `internal/modes` | 77.8% | ✅ good |
| `internal/config` | 64.5% | ⚠️ adequate |
| `internal/server` | 45.5% | ⚠️ adequate |
| `cmd/server` | 25.0% | ℹ️ expected (entry point) |

**recent improvements:**
- handlers package: 51.6% → 81.2% (+29.6 percentage points)
- 100+ new test cases across 7 test files
- comprehensive edge case coverage
- zero test failures (415/415 passing)

## troubleshooting

### server won't start

1. check that go is installed: `go version`
2. verify the binary was built: check `bin/` directory
3. enable debug mode: set `debug=true` in env

### tools not appearing

1. restart claude desktop completely
2. check config file syntax
3. verify the executable path is correct

### performance issues

**in-memory mode** (default):
- data is lost on server restart
- for long sessions, consider periodic restarts
- monitor memory usage if processing thousands of thoughts

**sqlite mode** (persistent):
- data persists across restarts
- uses write-through caching for fast access
- automatic memory management via cache eviction
- enable with `storage_type=sqlite` environment variable

## technical details

### key features

- 50 specialized mcp tools across 8 categories
- 6 thinking modes with automatic mode selection
- dual-process reasoning (fast vs slow thinking)
- checkpoint-based backtracking
- probabilistic bayesian inference
- causal reasoning with do-calculus
- abductive hypothesis generation
- case-based reasoning
- symbolic constraint solving
- hallucination detection
- confidence calibration tracking
- unknown unknowns detection
- workflow orchestration for multi-tool automation
- pluggable storage (in-memory or sqlite)
- thread-safe operations
- comprehensive test coverage

### implementation highlights

- modular package design for composability
- interface-based architecture for testability
- builder patterns for complex object construction
- write-through caching for sqlite performance
- fts5 full-text search
- wal mode for concurrent database reads
- automatic schema migrations
- graceful fallback handling
- resource limits to prevent dos

## license

mit license
