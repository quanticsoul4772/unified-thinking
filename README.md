# Unified Thinking Server

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Test Coverage](https://img.shields.io/badge/coverage-78.3%25-brightgreen)](https://github.com/quanticsoul4772/unified-thinking)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-compatible-purple)](https://modelcontextprotocol.io/)
[![Tools](https://img.shields.io/badge/tools-75-blue)](https://github.com/quanticsoul4772/unified-thinking)
[![Benchmarks](https://img.shields.io/badge/benchmarks-114%20problems-blue)](https://github.com/quanticsoul4772/unified-thinking)

A Model Context Protocol (MCP) server that consolidates multiple cognitive thinking patterns into a single Go-based implementation.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/quanticsoul4772/unified-thinking.git
cd unified-thinking

# Build the server
make build

# Configure Claude Desktop (see Installation section)
# Restart Claude Desktop to load the server
```

## features

### thinking modes

- **linear mode**: sequential step-by-step reasoning for systematic problem solving
- **tree mode**: multi-branch parallel exploration with insights and cross-references
- **divergent mode**: creative/unconventional ideation with "rebellion" capability
- **reflection mode**: metacognitive reflection on previous reasoning with insight extraction
- **backtracking mode**: checkpoint-based reasoning with ability to restore previous states
- **auto mode**: automatic mode selection based on input content
- **graph mode**: arbitrary graph structures with thought aggregation, refinement, and cyclic reasoning (Graph-of-Thoughts)

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

the server includes 75 specialized tools across 15 major categories:

#### probabilistic reasoning (4 tools)
- bayesian inference with mathematically correct two-likelihood updates (P(E|H) and P(E|¬H))
- evidence-based probability updates with configurable likelihood estimation
- belief combination operations (and/or logic)
- confidence estimation from evidence aggregation
- metrics tracking: update counts, uninformative evidence rate, error rate

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

#### episodic memory & learning (5 tools)
- reasoning session tracking (build complete trajectory from problem to solution)
- pattern learning (recognizes successful approaches across sessions)
- adaptive recommendations (suggests proven strategies based on similar past problems)
- trajectory search (learn from historical successes and failures)
- retrospective analysis (post-session analysis with actionable improvements)
- **Semantic embeddings** - optional hybrid search combining hash-based and vector similarity (Voyage AI)
- **Automatic knowledge graph population** - reasoning sessions automatically extract entities to knowledge graph

#### knowledge graph & semantic memory (3 tools)
- **store-entity** - store entities in Neo4j with semantic indexing via chromem-go vector search
- **search-knowledge-graph** - hybrid search combining semantic similarity (Voyage AI) + graph traversal (Neo4j)
- **create-relationship** - create typed relationships (CAUSES, ENABLES, CONTRADICTS, BUILDS_UPON, etc.)
- **Automatic extraction** - entities automatically extracted from reasoning sessions and stored in knowledge graph
- **Persistent vector storage** - chromem-go collections persist to disk (survive restarts)

#### thought similarity search (1 tool)
- **search-similar-thoughts** - semantic search over past thoughts using Voyage AI embeddings
- finds semantically similar reasoning chains for reuse
- reduces redundant computation through pattern matching

#### graph-of-thoughts (8 tools)
- **got-initialize** - start new graph with initial thought
- **got-generate** - create k diverse continuations using Anthropic Claude API
- **got-aggregate** - merge parallel reasoning paths via LLM synthesis
- **got-refine** - iteratively improve thoughts through self-critique (max 3 iterations)
- **got-score** - multi-criteria quality evaluation (confidence, validity, relevance, novelty, depth)
- **got-prune** - remove low-quality vertices below threshold
- **got-get-state** - retrieve current graph state with all vertices
- **got-finalize** - mark terminal vertices and extract conclusions
- **Multiple parents** - vertices can have multiple parents (key advantage over tree-of-thoughts)
- **Cyclic reasoning** - supports feedback loops from conclusions to premises
- **LLM-powered** - uses Claude Sonnet 4.5 for generation, synthesis, and refinement

#### context bridge (automatic)
- **cross-session context retrieval** - automatically surfaces similar past reasoning trajectories
- **hybrid similarity matching** - 70% embedding similarity + 30% concept overlap
- **graceful degradation** - continues with concept-only matching if embeddings timeout
- **performance metrics** - p50/p95/p99 latency, cache stats, error/timeout counts
- **visible response enrichment** - always shows context_bridge status in tool responses

## installation

### prerequisites

- Go 1.24 or higher
- Git

### macOS / Linux

#### Install Go

**macOS (using Homebrew):**
```bash
brew install go
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install golang-go
```

**Verify Go installation:**
```bash
go version  # Should show 1.24 or higher
```

#### Build the Server

```bash
# Clone the repository
git clone https://github.com/quanticsoul4772/unified-thinking.git
cd unified-thinking

# Download dependencies and build
go mod download
go build -o bin/unified-thinking ./cmd/server
```

Or using the Makefile (auto-detects your OS):

```bash
make build
```

The binary will be created at `bin/unified-thinking` (no .exe extension).

### Windows

#### Install Go

Download and install Go from [go.dev/dl](https://go.dev/dl/)

**Verify Go installation:**
```cmd
go version  # Should show 1.24 or higher
```

#### Build the Server

```bash
# Clone the repository
git clone https://github.com/quanticsoul4772/unified-thinking.git
cd unified-thinking

# Download dependencies and build
go mod download
go build -o bin/unified-thinking.exe ./cmd/server
```

Or using the Makefile:

```bash
make build
```

The binary will be created at `bin\unified-thinking.exe`.

## documentation

- **[API Reference](API_REFERENCE.md)** - complete reference for all 63 MCP tools with parameters, examples, and response formats
- **[Configuration Guide](docs/CONFIGURATION.md)** - detailed configuration options and examples
- **[Embeddings Guide](docs/EMBEDDINGS.md)** - semantic embeddings setup and usage

## configuration

**Config file locations:**
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux:** `~/.config/Claude/claude_desktop_config.json`

### minimal configuration (in-memory only)

Basic setup with no persistence - data lost on restart:

**macOS / Linux:**
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/unified-thinking/bin/unified-thinking",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\path\\to\\unified-thinking\\bin\\unified-thinking.exe",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

### recommended configuration (sqlite persistence)

Persistent storage for thoughts and trajectories across sessions:

**macOS / Linux:**
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/unified-thinking/bin/unified-thinking",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "~/Library/Application Support/Claude/unified-thinking.db"
      }
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\path\\to\\unified-thinking\\bin\\unified-thinking.exe",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "C:\\Users\\YourName\\AppData\\Roaming\\Claude\\unified-thinking.db"
      }
    }
  }
}
```

### full configuration (sqlite + semantic embeddings + knowledge graph)

Complete setup with persistence, semantic similarity, and knowledge graph:

**macOS / Linux:**
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/unified-thinking/bin/unified-thinking",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "~/Library/Application Support/Claude/unified-thinking.db",
        "VOYAGE_API_KEY": "your-voyage-api-key-here",
        "EMBEDDINGS_MODEL": "voyage-3-lite",
        "NEO4J_ENABLED": "true",
        "NEO4J_URI": "neo4j+s://your-instance.databases.neo4j.io",
        "NEO4J_USERNAME": "neo4j",
        "NEO4J_PASSWORD": "your-neo4j-password",
        "NEO4J_DATABASE": "neo4j",
        "ANTHROPIC_API_KEY": "your-anthropic-api-key-here"
      }
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\path\\to\\unified-thinking\\bin\\unified-thinking.exe",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "C:\\Users\\YourName\\AppData\\Roaming\\Claude\\unified-thinking.db",
        "VOYAGE_API_KEY": "your-voyage-api-key-here",
        "EMBEDDINGS_MODEL": "voyage-3-lite",
        "ANTHROPIC_API_KEY": "your-anthropic-api-key-here",
        "NEO4J_ENABLED": "true",
        "NEO4J_URI": "neo4j+s://your-instance.databases.neo4j.io",
        "NEO4J_USERNAME": "neo4j",
        "NEO4J_PASSWORD": "your-neo4j-password",
        "NEO4J_DATABASE": "neo4j"
      }
    }
  }
}
```

### environment variables

**Storage Configuration**:
- `STORAGE_TYPE`: `memory` (default) or `sqlite`
- `SQLITE_PATH`: Path to database file (created automatically if it doesn't exist)
- `SQLITE_TIMEOUT`: Connection timeout in milliseconds (default: `5000`)
- `DEBUG`: Enable debug logging (`true` or `false`)
- `AUTO_VALIDATION_THRESHOLD`: Confidence threshold for auto-validation (default: `0.5`)

**Semantic Embeddings** (optional - improves similarity search):
- `VOYAGE_API_KEY`: Your Voyage AI API key from https://dashboard.voyageai.com/
- `EMBEDDINGS_MODEL`: Model to use (default: `voyage-3-lite`, also: `voyage-3`, `voyage-3-large`)
- `EMBEDDINGS_RRF_K`: Reciprocal rank fusion parameter (default: `60`)
- `EMBEDDINGS_MIN_SIMILARITY`: Minimum cosine similarity threshold (default: `0.5`)
- `EMBEDDINGS_CACHE_ENABLED`: Cache embeddings for reuse (default: `true`)
- `EMBEDDINGS_CACHE_TTL`: Cache time-to-live (default: `24h`)

**Context Bridge** (automatic cross-session context):
- `CONTEXT_BRIDGE_ENABLED`: Enable cross-session context retrieval (default: `true`)
- `CONTEXT_BRIDGE_MIN_SIMILARITY`: Minimum similarity for matches (default: `0.7`)
- `CONTEXT_BRIDGE_MAX_MATCHES`: Maximum matches to return (default: `3`)
- `CONTEXT_BRIDGE_CACHE_SIZE`: LRU cache size (default: `100`)
- `CONTEXT_BRIDGE_CACHE_TTL`: Cache time-to-live (default: `15m`)
- `CONTEXT_BRIDGE_TIMEOUT`: Timeout per enrichment (default: `2s`)

**Knowledge Graph** (Neo4j integration for semantic memory):
- `NEO4J_ENABLED`: Enable knowledge graph integration (`true` or `false`, default: `false`)
- `NEO4J_URI`: Neo4j connection URI (e.g., `neo4j+s://your-instance.databases.neo4j.io`)
- `NEO4J_USERNAME`: Neo4j username (default: `neo4j`)
- `NEO4J_PASSWORD`: Neo4j password (JWT token for Aura, or password for self-hosted)
- `NEO4J_DATABASE`: Database name (default: `neo4j`)
- `NEO4J_TIMEOUT_MS`: Connection timeout in milliseconds (default: `5000`)
- `VECTOR_STORE_PATH`: Persistent chromem-go vector storage path (defaults to `{SQLITE_PATH}_vectors`)

**Graph-of-Thoughts** (LLM-powered graph reasoning):
- `ANTHROPIC_API_KEY`: Anthropic API key (REQUIRED for GoT tools, server fails if missing)
- `GOT_MODEL`: Model to use (default: `claude-sonnet-4-5-20250929`)

**Important Notes**:
- **Trajectory persistence requires SQLite**: Set `STORAGE_TYPE=sqlite` to enable episodic memory persistence
- **Knowledge graph requires Neo4j + Voyage AI**: Both `NEO4J_ENABLED=true` and `VOYAGE_API_KEY` must be set
- **Graph-of-Thoughts requires Anthropic API**: Server will not start without `ANTHROPIC_API_KEY`
- **Restart required**: Changes to configuration require restarting Claude Desktop to take effect

## recent updates

### new features (latest: graph-of-thoughts)

- **graph-of-thoughts (GoT)** (3 commits, 2,029 lines):
  - arbitrary graph structures vs tree-only (key advantage: multiple parents per thought)
  - 8 new MCP tools: got-initialize, got-generate, got-aggregate, got-refine, got-score, got-prune, got-get-state, got-finalize
  - LLM-powered operations using Anthropic Claude Sonnet 4.5 API
  - generate: creates k diverse continuations from active vertices
  - aggregate: synthesizes parallel reasoning paths into unified insights
  - refine: iterative self-improvement through critique (max 3 iterations)
  - score: multi-criteria evaluation (confidence 25%, validity 30%, relevance 25%, novelty 10%, depth 10%)
  - prune: removes low-quality thoughts while preserving roots and terminals
  - cyclic reasoning: supports feedback loops from conclusions back to premises
  - research-backed: 61-69% error reduction vs tree-of-thoughts on sorting tasks
  - test suite: 25 unit tests validating graph operations and LLM integration

- **knowledge graph with neo4j + chromem-go** (16 commits, 6,744 lines):
  - automatic entity extraction from reasoning sessions (always enabled when Neo4j available)
  - hybrid search combining semantic similarity (Voyage AI) + graph traversal (Neo4j)
  - persistent vector storage (chromem-go persists to disk, survives restarts)
  - 3 new MCP tools: store-entity, search-knowledge-graph, create-relationship
  - 7 entity types: Concept, Person, Tool, File, Decision, Strategy, Problem
  - 7 relationship types: CAUSES, ENABLES, CONTRADICTS, BUILDS_UPON, RELATES_TO, HAS_OBSERVATION, USED_IN_CONTEXT
  - regex-based entity extraction (10 patterns) with LLM integration ready
  - integration with Thompson Sampling RL and episodic memory
  - test suite: 21 integration tests (100% pass rate with Neo4j)
  - production-verified with Neo4j Aura cloud instances

### previous features
- **context bridge for cross-session learning**: automatic retrieval of similar past reasoning trajectories
  - hybrid similarity: 70% embedding cosine similarity + 30% concept jaccard similarity
  - graceful degradation: continues with concept-only matching if embeddings fail
  - performance metrics: p50/p95/p99 latency, cache stats, error/timeout counts
  - visible status in responses: always shows context_bridge field with match status
  - proactive rate limiting: token bucket rate limiter for Voyage AI API calls
  - backfill utility: batch processing for adding embeddings to existing trajectories

- **semantic embeddings for episodic memory**: optional hybrid search combining hash-based and vector similarity
  - voyage ai integration (200m free tokens, anthropic's recommended provider)
  - rrf (reciprocal rank fusion) for optimal result combination
  - blob-based vector storage in sqlite (no cgo required)
  - transparent fallback to hash-based search when embeddings disabled

### performance improvements
- **production optimizations**: performance tuning completed
  - tier 1-3 optimizations: memory allocation patterns, locking efficiency, hot path optimization
  - production profiling: identified and resolved critical bottlenecks
  - improved memory allocation and eliminated redundant locking
- context bridge metrics now exposed in get-metrics response
- context bridge always returns structure for visibility (even with no matches)
- proactive rate limiting prevents api throttling (30 req/sec with burst of 10)
- converted all fmt.Printf debug statements to log.Printf for consistency
- applied gofmt -s across entire codebase for consistent formatting

### bug fixes
- **trajectory persistence**: fixed episodic memory trajectories not persisting across restarts (schema v6)
  - implemented json-based trajectory storage in sqlite
  - auto-load existing trajectories on server initialization
  - rebuild all indexes (problem, domain, tag, tool sequence) from persisted data
- fixed sqlite foreign key constraint failures in episodic memory
- resolved nil array validation errors in mcp tool responses
- fixed race conditions in concurrent sqlite tests
- fixed context bridge not showing metrics in get-metrics
- fixed context bridge not returning structure when no matches found

### code quality
- 73% test coverage for embeddings package
- integration tests for embedding similarity path
- cleaned up test files and improved edge case coverage
- maintained backward compatibility while improving internals

## available tools (66 total)

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

16. **make-decision** - multi-criteria decision analysis with persistent storage
    ```json
    {
      "question": "which option should we choose?",
      "options": [{"name": "option a", "scores": {"cost": 0.8}}],
      "criteria": [{"name": "cost", "weight": 0.6, "maximize": false}]
    }
    ```
    decisions are stored and can be re-evaluated when new evidence arrives via `process-evidence-pipeline`

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

### enhanced tools (8 tools)

51. **find-analogy** - find analogies between source and target domains for cross-domain reasoning
52. **apply-analogy** - apply an existing analogy to a new context
53. **decompose-argument** - break down arguments into premises, claims, assumptions, and inference chains
54. **generate-counter-arguments** - generate counter-arguments using multiple strategies
55. **detect-fallacies** - detect formal and informal logical fallacies (ad hominem, straw man, false dichotomy, etc.)
56. **process-evidence-pipeline** - automatically update beliefs, causal graphs, and decisions from new evidence
57. **analyze-temporal-causal-effects** - analyze temporal progression of causal effects (short/medium/long-term)
58. **analyze-decision-timing** - determine optimal timing for decisions based on causal and temporal factors

### episodic memory & learning tools (5 tools)

59. **start-reasoning-session** - start tracking a reasoning session to build episodic memory
    ```json
    {
      "session_id": "debug_2024_001",
      "description": "optimize database query performance",
      "goals": ["reduce query time", "improve user experience"],
      "domain": "software-engineering",
      "complexity": 0.6
    }
    ```

60. **complete-reasoning-session** - complete session, calculate quality metrics, trigger pattern learning, **automatically extract entities to knowledge graph**
61. **get-recommendations** - get adaptive recommendations based on similar past problems
62. **search-trajectories** - search past reasoning sessions to learn from successes and failures
63. **analyze-trajectory** - perform retrospective analysis of completed session (strengths, weaknesses, improvements)

### knowledge graph tools (3 tools)

64. **store-entity** - store entities in knowledge graph with semantic indexing
    ```json
    {
      "entity_id": "concept-optimization",
      "label": "Database Optimization",
      "type": "Concept",
      "content": "Techniques for optimizing database queries...",
      "description": "Core optimization strategies",
      "metadata": {"category": "performance"}
    }
    ```

65. **search-knowledge-graph** - hybrid semantic + graph traversal search
    ```json
    {
      "query": "database performance",
      "search_type": "semantic|graph|hybrid",
      "limit": 10,
      "max_hops": 2,
      "min_similarity": 0.5
    }
    ```

66. **create-relationship** - create typed relationships between entities
    ```json
    {
      "from_id": "entity-1",
      "to_id": "entity-2",
      "type": "CAUSES|ENABLES|CONTRADICTS|BUILDS_UPON",
      "strength": 0.9,
      "confidence": 0.95
    }
    ```

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
│   ├── memory/             # episodic memory and pattern learning
│   │   ├── episodic.go     # reasoning trajectory storage
│   │   ├── learning.go     # pattern recognition
│   │   └── retrospective.go # post-session analysis
│   ├── embeddings/         # semantic embeddings for similarity
│   │   ├── voyage.go       # voyage ai embedder with rate limiting
│   │   └── backfill.go     # batch embedding generation utility
│   ├── contextbridge/      # cross-session context retrieval
│   │   ├── bridge.go       # response enrichment and metrics
│   │   ├── matcher.go      # trajectory similarity matching
│   │   ├── similarity.go   # hybrid similarity calculation
│   │   └── config.go       # feature configuration
│   ├── knowledge/          # knowledge graph integration
│   │   ├── neo4j_client.go      # Neo4j connection management
│   │   ├── graph_store.go       # entity/relationship CRUD
│   │   ├── vector_store.go      # chromem-go semantic search
│   │   ├── embedding_cache.go   # SQLite-based embedding cache
│   │   ├── knowledge_graph.go   # unified hybrid search API
│   │   ├── episodic_integration.go  # automatic extraction from trajectories
│   │   ├── rl_integration.go    # Thompson Sampling RL context
│   │   ├── schema.go            # Neo4j schema definitions
│   │   └── extraction/          # entity extraction pipeline
│   │       ├── regex_extractor.go   # pattern-based extraction
│   │       ├── llm_extractor.go     # LLM-based extraction (ready)
│   │       └── hybrid_extractor.go  # intelligent routing
│   └── server/             # mcp server implementation
│       └── handlers/       # specialized tool handlers (21 files)
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
- **memory**: episodic memory for session tracking, pattern learning, and adaptive recommendations
- **embeddings**: semantic embedding generation with Voyage AI for similarity search
- **contextbridge**: automatic cross-session context retrieval with hybrid similarity matching
- **knowledge**: Neo4j graph database + chromem-go vector search for semantic memory and entity relationships

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

**test coverage**: 78.3% overall | 100% pass rate | 135 test files

### coverage by package

| package | coverage |
|---------|----------|
| `internal/types` | 100.0% |
| `internal/metrics` | 100.0% |
| `internal/config` | 97.3% |
| `internal/reasoning` | 94.8% |
| `internal/memory` | 91.0% |
| `internal/modes` | 90.5% |
| `internal/analysis` | 89.3% |
| `internal/validation` | 88.8% |
| `internal/storage` | 87.9% |
| `internal/orchestration` | 87.7% |
| `internal/metacognition` | 87.2% |
| `internal/processing` | 83.3% |
| `internal/server` | 79.3% |
| `internal/server/handlers` | 79.0% |
| `internal/integration` | 78.7% |
| `internal/embeddings` | 75.6% |
| `internal/knowledge/extraction` | 71.7% |
| `cmd/server` | 70.6% |
| `internal/contextbridge` | 70.6% |
| `internal/knowledge` | 64.9% (requires Neo4j) |

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

- 66 specialized mcp tools across 14 categories
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
- episodic memory with session tracking and pattern learning
- adaptive recommendations based on historical reasoning sessions
- retrospective analysis for continuous improvement
- knowledge graph integration (Neo4j + chromem-go)
- automatic entity extraction from reasoning sessions
- hybrid semantic + graph traversal search
- pluggable storage (in-memory or sqlite)
- persistent vector storage with chromem-go
- thread-safe operations
- test coverage: 135 test files, 78.3% overall

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
- neo4j property graph model with cypher queries
- chromem-go pure go vector database (no cgo)
- persistent vector collections (survive restarts)
- automatic entity extraction with regex patterns
- hybrid similarity search (semantic + graph)

## Contributing

See [Contributing Guidelines](CONTRIBUTING.md) for details.

- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Issue Templates](.github/ISSUE_TEMPLATE)
- [PR Template](.github/pull_request_template.md)

## Security

For security issues, see [Security Policy](SECURITY.md). Do not report security vulnerabilities through public GitHub issues.

## License

MIT License - see [LICENSE](LICENSE) file.

## Acknowledgments

- Built with the [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)

## Support

- Issues: [GitHub Issues](https://github.com/quanticsoul4772/unified-thinking/issues)
- Discussions: [GitHub Discussions](https://github.com/quanticsoul4772/unified-thinking/discussions)
