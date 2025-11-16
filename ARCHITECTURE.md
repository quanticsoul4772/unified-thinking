# Unified Thinking Server - Architecture Diagrams

Visual representations of the system architecture, data flows, and component interactions.

---

## Table of Contents

1. [High-Level System Architecture](#high-level-system-architecture)
2. [MCP Protocol Communication Flow](#mcp-protocol-communication-flow)
3. [Storage Architecture](#storage-architecture)
4. [Thinking Modes Architecture](#thinking-modes-architecture)
5. [Tool Organization](#tool-organization)
6. [Data Flow Diagrams](#data-flow-diagrams)
7. [Integration Patterns](#integration-patterns)

---

## High-Level System Architecture

```mermaid
flowchart TB
    subgraph "Claude Desktop"
        Client[Claude AI Client]
    end

    subgraph "MCP Protocol Layer"
        Protocol[stdio Transport]
    end

    subgraph "Unified Thinking Server"
        Server[MCP Server Core]

        subgraph "Handlers (50 Tools)"
            CoreHandlers[Core Thinking<br/>11 tools]
            ProbHandlers[Probabilistic<br/>4 tools]
            CausalHandlers[Causal<br/>5 tools]
            MetaHandlers[Metacognition<br/>3 tools]
            AdvHandlers[Advanced<br/>9 tools]
            IntHandlers[Integration<br/>6 tools]
            OtherHandlers[Other<br/>12 tools]
        end

        subgraph "Processing Layer"
            Modes[Thinking Modes<br/>6 modes]
            DualProcess[Dual-Process<br/>System 1/2]
            Validation[Validation<br/>Logic & Proof]
        end

        subgraph "Reasoning Engines"
            Prob[Probabilistic<br/>Bayesian]
            Causal[Causal<br/>Pearl's Framework]
            Temporal[Temporal<br/>Analysis]
            Abductive[Abductive<br/>Hypotheses]
            CBR[Case-Based<br/>Reasoning]
            Symbolic[Symbolic<br/>Constraints]
        end

        subgraph "Analysis Layer"
            Evidence[Evidence<br/>Assessment]
            Contradiction[Contradiction<br/>Detection]
            Perspective[Perspective<br/>Analysis]
        end

        subgraph "Metacognition Layer"
            SelfEval[Self-<br/>Evaluation]
            BiasDetect[Bias<br/>Detection]
            BlindSpots[Blind Spot<br/>Detection]
        end

        subgraph "Storage Layer"
            StorageInterface[Storage Interface]
            Memory[In-Memory<br/>Storage]
            SQLite[SQLite<br/>Storage]
        end
    end

    Client <-->|JSON-RPC| Protocol
    Protocol <-->|stdio| Server

    Server --> CoreHandlers & ProbHandlers & CausalHandlers & MetaHandlers & AdvHandlers & IntHandlers & OtherHandlers

    CoreHandlers --> Modes
    CoreHandlers --> Validation
    ProbHandlers --> Prob
    CausalHandlers --> Causal
    AdvHandlers --> DualProcess & Abductive & CBR & Symbolic
    MetaHandlers --> SelfEval & BiasDetect & BlindSpots

    Modes --> StorageInterface
    Prob --> StorageInterface
    Causal --> StorageInterface
    Temporal --> StorageInterface
    Evidence --> StorageInterface

    StorageInterface -->|Factory Pattern| Memory
    StorageInterface -->|Factory Pattern| SQLite

    style Server fill:#e1f5ff
    style StorageInterface fill:#fff4e1
    style Modes fill:#e8f5e9
```

---

## MCP Protocol Communication Flow

```mermaid
sequenceDiagram
    participant Claude as Claude Desktop
    participant Transport as stdio Transport
    participant Server as MCP Server
    participant Handler as Tool Handler
    participant Mode as Thinking Mode
    participant Storage as Storage Layer

    Claude->>Transport: JSON-RPC Request<br/>(tool call)
    Transport->>Server: Parse & Route
    Server->>Handler: Invoke Handler

    alt Core Thinking Tool
        Handler->>Mode: ProcessThought()
        Mode->>Storage: StoreThought()
        Storage-->>Mode: thought_id
        Mode-->>Handler: ThoughtResult
    else Reasoning Tool
        Handler->>Mode: Execute Reasoning
        Mode->>Storage: Store Results
        Storage-->>Mode: result_id
        Mode-->>Handler: ReasoningResult
    else Query Tool
        Handler->>Storage: Query Data
        Storage-->>Handler: QueryResult
    end

    Handler-->>Server: Structured Response
    Server-->>Transport: JSON-RPC Response
    Transport-->>Claude: Tool Result

    Note over Claude,Storage: All communication via JSON-RPC over stdio
```

---

## Storage Architecture

```mermaid
flowchart TB
    subgraph "Application Layer"
        Handlers[Tool Handlers]
        Modes[Thinking Modes]
        Reasoning[Reasoning Engines]
    end

    subgraph "Storage Interface Layer"
        Interface[Storage Interface]

        subgraph "Repository Interfaces"
            ThoughtRepo[ThoughtRepository]
            BranchRepo[BranchRepository]
            InsightRepo[InsightRepository]
            ValidationRepo[ValidationRepository]
            RelationshipRepo[RelationshipRepository]
            MetricsProvider[MetricsProvider]
        end
    end

    subgraph "Storage Implementations"
        direction LR

        subgraph "In-Memory Storage"
            MemThoughts[Thoughts Map]
            MemBranches[Branches Map]
            MemMutex[RWMutex<br/>Thread Safety]
        end

        subgraph "SQLite Storage"
            SQLiteDB[(SQLite DB)]
            Cache[Write-Through<br/>Cache]
            FTS5[FTS5 Full-Text<br/>Search]
            WAL[WAL Mode<br/>Concurrent Reads]
        end
    end

    subgraph "Configuration"
        Factory[Storage Factory]
        Config[Environment<br/>Variables]
    end

    Handlers & Modes & Reasoning --> Interface

    Interface --> ThoughtRepo & BranchRepo & InsightRepo & ValidationRepo & RelationshipRepo & MetricsProvider

    ThoughtRepo & BranchRepo --> Factory

    Config --> Factory
    Factory -->|STORAGE_TYPE=memory| MemThoughts
    Factory -->|STORAGE_TYPE=sqlite| SQLiteDB

    MemThoughts --> MemBranches
    MemBranches --> MemMutex

    SQLiteDB --> Cache
    SQLiteDB --> FTS5
    SQLiteDB --> WAL

    style Interface fill:#e1f5ff
    style Factory fill:#fff4e1
    style MemMutex fill:#e8f5e9
    style Cache fill:#e8f5e9
```

### Storage Factory Pattern

```mermaid
flowchart LR
    ENV[Environment<br/>Variables]
    Factory[Storage Factory]
    Memory[In-Memory<br/>Storage]
    SQLite[SQLite<br/>Storage]
    Fallback[Fallback<br/>Handler]

    ENV -->|STORAGE_TYPE=memory| Factory
    ENV -->|STORAGE_TYPE=sqlite| Factory
    ENV -->|STORAGE_FALLBACK=memory| Fallback

    Factory -->|Create| Memory
    Factory -->|Create| SQLite

    SQLite -.->|Error| Fallback
    Fallback --> Memory

    style Factory fill:#fff4e1
    style Fallback fill:#ffe1e1
```

---

## Thinking Modes Architecture

```mermaid
flowchart TB
    Input[User Input]
    Auto[Auto Mode<br/>Keyword Detection]
    Registry[Mode Registry]

    subgraph "Thinking Modes"
        Linear[Linear Mode<br/>Sequential Reasoning]
        Tree[Tree Mode<br/>Parallel Exploration]
        Divergent[Divergent Mode<br/>Creative Ideation]
        Reflection[Reflection Mode<br/>Metacognitive Review]
        Backtracking[Backtracking Mode<br/>Checkpoint-Based]
    end

    subgraph "Mode Components"
        direction LR
        ThoughtProcessor[ProcessThought]
        InsightGen[Insight Generation]
        CrossRef[Cross-Reference]
        Priority[Priority Calc]
    end

    Storage[(Storage Layer)]
    Output[Structured Output]

    Input --> Auto
    Auto --> Registry

    Registry -->|Select Mode| Linear
    Registry -->|Select Mode| Tree
    Registry -->|Select Mode| Divergent
    Registry -->|Select Mode| Reflection
    Registry -->|Select Mode| Backtracking

    Linear --> ThoughtProcessor
    Tree --> ThoughtProcessor & InsightGen & CrossRef & Priority
    Divergent --> ThoughtProcessor
    Reflection --> ThoughtProcessor & InsightGen
    Backtracking --> ThoughtProcessor

    ThoughtProcessor --> Storage
    InsightGen --> Storage
    CrossRef --> Storage
    Priority --> Storage

    Storage --> Output

    style Auto fill:#e1f5ff
    style Registry fill:#fff4e1
    style Tree fill:#e8f5e9
```

### Auto Mode Selection Logic

```mermaid
flowchart TD
    Start[Input Content]

    CheckBranch{Has<br/>branch_id?}
    CheckDivergent{Contains divergent<br/>keywords?}
    CheckReflection{Contains reflection<br/>keywords?}
    CheckBacktrack{Contains backtrack<br/>keywords?}
    CheckTree{Has key_points or<br/>tree keywords?}

    Linear[Linear Mode]
    Tree[Tree Mode]
    Divergent[Divergent Mode]
    Reflection[Reflection Mode]
    Backtrack[Backtracking Mode]

    Start --> CheckBranch
    CheckBranch -->|Yes| Tree
    CheckBranch -->|No| CheckDivergent

    CheckDivergent -->|Yes| Divergent
    CheckDivergent -->|No| CheckReflection

    CheckReflection -->|Yes| Reflection
    CheckReflection -->|No| CheckBacktrack

    CheckBacktrack -->|Yes| Backtrack
    CheckBacktrack -->|No| CheckTree

    CheckTree -->|Yes| Tree
    CheckTree -->|No| Linear

    style Start fill:#e1f5ff
    style Divergent fill:#ffe1f0
    style Tree fill:#e8f5e9
```

---

## Tool Organization

```mermaid
mindmap
  root((Unified Thinking<br/>50 Tools))
    Core Thinking<br/>11 tools
      think
      history
      list-branches
      focus-branch
      branch-history
      recent-branches
      validate
      prove
      check-syntax
      search
      get-metrics
    Probabilistic<br/>4 tools
      probabilistic-reasoning
      assess-evidence
      detect-contradictions
      sensitivity-analysis
    Decision<br/>3 tools
      make-decision
      decompose-problem
      verify-thought
    Metacognition<br/>3 tools
      self-evaluate
      detect-biases
      detect-blind-spots
    Calibration<br/>4 tools
      get-hallucination-report
      record-prediction
      record-outcome
      get-calibration-report
    Temporal<br/>4 tools
      analyze-perspectives
      analyze-temporal
      compare-time-horizons
      identify-optimal-timing
    Causal<br/>5 tools
      build-causal-graph
      simulate-intervention
      generate-counterfactual
      analyze-correlation-vs-causation
      get-causal-graph
    Integration<br/>6 tools
      synthesize-insights
      detect-emergent-patterns
      execute-workflow
      list-workflows
      register-workflow
      list-integration-patterns
    Advanced<br/>10 tools
      dual-process-think
      create-checkpoint
      restore-checkpoint
      list-checkpoints
      generate-hypotheses
      evaluate-hypotheses
      retrieve-similar-cases
      perform-cbr-cycle
      prove-theorem
      check-constraints
```

---

## Data Flow Diagrams

### Think Tool Data Flow

```mermaid
flowchart LR
    Input[User Request]
    AutoMode[Auto Mode<br/>Selection]
    ModeExec[Mode<br/>Execution]

    subgraph "Processing"
        Parse[Parse Content]
        Analyze[Analyze<br/>Complexity]
        Generate[Generate<br/>Thought]
    end

    subgraph "Storage Operations"
        Store[Store Thought]
        UpdateBranch[Update Branch<br/>if Tree Mode]
        GenInsights[Generate<br/>Insights]
    end

    subgraph "Validation"
        AutoVal[Auto-Validation<br/>if confidence < 0.5]
        Quality[Quality<br/>Assessment]
    end

    subgraph "Response Generation"
        Metadata[Generate<br/>Metadata]
        Suggestions[Next Tool<br/>Suggestions]
        Export[Export<br/>Formats]
    end

    Output[JSON Response]

    Input --> AutoMode
    AutoMode --> ModeExec
    ModeExec --> Parse
    Parse --> Analyze
    Analyze --> Generate

    Generate --> Store
    Store --> UpdateBranch
    UpdateBranch --> GenInsights

    GenInsights --> AutoVal
    AutoVal --> Quality

    Quality --> Metadata
    Metadata --> Suggestions
    Suggestions --> Export

    Export --> Output

    style AutoMode fill:#e1f5ff
    style AutoVal fill:#fff4e1
    style Output fill:#e8f5e9
```

### Probabilistic Reasoning Data Flow

```mermaid
flowchart TB
    subgraph "Create Belief"
        CreateInput[Statement +<br/>Prior Prob]
        CreateBelief[Create Belief<br/>Object]
        StoreB[Store Belief]
    end

    subgraph "Update Belief"
        Evidence[Evidence +<br/>Likelihood]
        Bayesian[Bayesian<br/>Inference]
        Calculate[Calculate<br/>Posterior]
        UpdateB[Update Belief]
    end

    subgraph "Combine Beliefs"
        Multiple[Multiple<br/>Beliefs]
        Logic[AND/OR<br/>Logic]
        Combine[Combine<br/>Probabilities]
        StoreC[Store Combined]
    end

    Storage[(Storage)]
    Response[Response with<br/>Updated Probability]

    CreateInput --> CreateBelief
    CreateBelief --> StoreB
    StoreB --> Storage

    Evidence --> Bayesian
    Bayesian --> Calculate
    Calculate --> UpdateB
    UpdateB --> Storage

    Multiple --> Logic
    Logic --> Combine
    Combine --> StoreC
    StoreC --> Storage

    Storage --> Response

    style Bayesian fill:#e1f5ff
    style Calculate fill:#fff4e1
```

### Causal Reasoning Data Flow

```mermaid
flowchart TB
    subgraph "Build Graph"
        Observations[Causal<br/>Observations]
        Extract[Extract Variables<br/>& Links]
        BuildDAG[Build DAG]
        StoreGraph[Store Graph]
    end

    subgraph "Simulate Intervention"
        GetGraph[Get Causal<br/>Graph]
        Surgery[Graph Surgery<br/>Remove Incoming Edges]
        Propagate[Propagate<br/>Effects]
        Calculate[Calculate<br/>Probabilities]
    end

    subgraph "Counterfactual"
        Scenario[What-If<br/>Scenario]
        Changes[Variable<br/>Changes]
        Simulate[Simulate<br/>Outcomes]
    end

    Storage[(Storage)]
    Export[Export to<br/>Memory MCP]

    Observations --> Extract
    Extract --> BuildDAG
    BuildDAG --> StoreGraph
    StoreGraph --> Storage

    GetGraph --> Surgery
    Surgery --> Propagate
    Propagate --> Calculate
    Calculate --> Storage

    Scenario --> Changes
    Changes --> Simulate
    Simulate --> Storage

    Storage --> Export

    style Surgery fill:#e1f5ff
    style Propagate fill:#fff4e1
```

---

## Integration Patterns

### Research-Enhanced Thinking Pattern

```mermaid
sequenceDiagram
    participant User
    participant UT as Unified Thinking
    participant BS as Brave Search
    participant Memory

    User->>BS: Search for information
    BS-->>User: Search results

    User->>UT: think (with context)
    UT-->>User: Thought + low confidence

    Note over UT: Auto-suggests web search

    User->>BS: Validate findings
    BS-->>User: Additional evidence

    User->>UT: assess-evidence
    UT-->>User: Evidence quality scores

    User->>UT: think (with evidence)
    UT-->>User: Updated thought + high confidence

    User->>Memory: Store findings
```

### Knowledge-Backed Decision Making Pattern

```mermaid
sequenceDiagram
    participant User
    participant Memory as Memory MCP
    participant Conv as Conversation
    participant UT as Unified Thinking
    participant Obs as Obsidian

    User->>Memory: traverse_graph (find related concepts)
    Memory-->>User: Related entities

    User->>Conv: conversation_search (past decisions)
    Conv-->>User: Historical context

    User->>UT: make-decision (with context)
    UT-->>User: Decision + export formats

    User->>Memory: create_entities (decision rationale)
    Memory-->>User: Entities created

    User->>Obs: create-note (use export format)
    Obs-->>User: Note created
```

### Causal Model to Knowledge Graph Pattern

```mermaid
sequenceDiagram
    participant User
    participant BS as Brave Search
    participant UT as Unified Thinking
    participant Memory as Memory MCP

    User->>BS: Research causal relationships
    BS-->>User: Evidence

    User->>UT: build-causal-graph
    UT-->>User: Graph + export formats

    Note over UT: Provides memory_entities<br/>and memory_relations

    User->>Memory: create_entities (from export)
    Memory-->>User: Entities created

    User->>Memory: create_relations (from export)
    Memory-->>User: Relations created

    User->>UT: simulate-intervention
    UT-->>User: Predicted effects
```

---

## Component Interaction Matrix

```
┌─────────────────────────────────────────────────────────────────┐
│                    Component Interactions                       │
├──────────────┬────────┬────────┬────────┬────────┬─────────────┤
│  Component   │ Modes  │ Storage│Validate│Reasoning│  Analysis  │
├──────────────┼────────┼────────┼────────┼────────┼─────────────┤
│ Handlers     │   ✓    │   ✓    │   ✓    │   ✓    │     ✓      │
│ Modes        │   -    │   ✓    │   ✓    │   -    │     -      │
│ Storage      │   -    │   -    │   -    │   -    │     -      │
│ Validation   │   -    │   ✓    │   -    │   -    │     ✓      │
│ Reasoning    │   -    │   ✓    │   ✓    │   -    │     ✓      │
│ Analysis     │   -    │   ✓    │   ✓    │   ✓    │     -      │
│ Metacognition│   -    │   ✓    │   ✓    │   -    │     ✓      │
│ Integration  │   ✓    │   ✓    │   -    │   ✓    │     ✓      │
└──────────────┴────────┴────────┴────────┴────────┴─────────────┘

Legend:
  ✓ = Direct dependency/interaction
  - = No direct interaction
```

---

## Package Dependencies

```
unified-thinking/
│
├── cmd/server (main entry point)
│   └── depends on: server, storage, orchestration
│
├── internal/types (core data structures)
│   └── no dependencies (foundation package)
│
├── internal/storage (storage layer)
│   └── depends on: types
│
├── internal/modes (thinking modes)
│   └── depends on: types, storage
│
├── internal/processing (dual-process)
│   └── depends on: types, storage, modes
│
├── internal/reasoning (probabilistic, causal, etc.)
│   └── depends on: types, storage
│
├── internal/analysis (evidence, contradiction, etc.)
│   └── depends on: types, storage, reasoning
│
├── internal/metacognition (self-eval, biases)
│   └── depends on: types, storage, validation
│
├── internal/validation (logic, proofs)
│   └── depends on: types, storage
│
├── internal/integration (synthesis, patterns)
│   └── depends on: types, storage, modes, reasoning
│
├── internal/orchestration (workflows)
│   └── depends on: types, all reasoning packages
│
└── internal/server (MCP server & handlers)
    └── depends on: all packages
```

---

## Thread Safety Model

```mermaid
flowchart TB
    subgraph "Concurrent Requests"
        R1[Request 1]
        R2[Request 2]
        R3[Request 3]
    end

    subgraph "Handler Layer"
        H1[Handler Instance]
        H2[Handler Instance]
        H3[Handler Instance]
    end

    subgraph "Storage Layer (Thread-Safe)"
        Lock[RWMutex]

        subgraph "Read Operations"
            ReadOp1[Read 1]
            ReadOp2[Read 2]
            ReadOp3[Read 3]
        end

        subgraph "Write Operations"
            WriteOp[Write<br/>Exclusive Lock]
        end
    end

    R1 --> H1
    R2 --> H2
    R3 --> H3

    H1 & H2 & H3 --> Lock

    Lock -->|Shared Lock| ReadOp1 & ReadOp2 & ReadOp3
    Lock -->|Exclusive Lock| WriteOp

    style Lock fill:#e1f5ff
    style WriteOp fill:#ffe1e1
```

---

## Performance Characteristics

```
┌─────────────────────────────────────────────────────────────┐
│              Operation Performance Profile                   │
├────────────────────────┬──────────┬─────────┬───────────────┤
│      Operation         │ In-Memory│ SQLite  │  Complexity   │
├────────────────────────┼──────────┼─────────┼───────────────┤
│ Store Thought          │  < 1ms   │ 1-5ms   │    O(1)       │
│ Get Thought            │  < 1ms   │ < 1ms   │    O(1)       │
│ Search Thoughts        │  1-10ms  │ 5-20ms  │    O(n)       │
│ List Branches          │  < 1ms   │ 1-3ms   │    O(b)       │
│ Validate Thought       │  1-5ms   │ 1-5ms   │    O(n)       │
│ Build Causal Graph     │  5-20ms  │ 10-30ms │    O(v+e)     │
│ Simulate Intervention  │  10-50ms │ 20-60ms │    O(v+e)     │
│ Synthesize Insights    │  20-100ms│ 30-120ms│    O(n²)      │
│ Detect Contradictions  │  10-50ms │ 15-60ms │    O(n²)      │
└────────────────────────┴──────────┴─────────┴───────────────┘

Legend:
  n = number of thoughts
  b = number of branches
  v = number of variables in graph
  e = number of edges in graph
```

---

## Deployment Architecture

```mermaid
flowchart TB
    subgraph "User Machine"
        Desktop[Claude Desktop<br/>Application]

        subgraph "MCP Server Process"
            Server[Unified Thinking<br/>Server]
            Storage[Storage Backend]

            subgraph "Storage Options"
                Memory[In-Memory]
                DB[(SQLite DB<br/>File)]
            end
        end

        Config[claude_desktop_config.json]
    end

    Desktop -->|Launches on Startup| Server
    Config -->|Configuration| Server

    Server --> Storage
    Storage -->|STORAGE_TYPE=memory| Memory
    Storage -->|STORAGE_TYPE=sqlite| DB

    Desktop <-->|JSON-RPC via stdio| Server

    style Desktop fill:#e1f5ff
    style Server fill:#e8f5e9
    style Config fill:#fff4e1
```

---

## Error Handling Flow

```mermaid
flowchart TB
    Request[Incoming Request]
    Validate[Validate<br/>Parameters]

    ValidCheck{Valid?}
    Execute[Execute<br/>Handler]

    ExecCheck{Success?}
    Storage[Storage<br/>Operation]

    StorageCheck{Success?}
    Fallback[Fallback<br/>Handler]

    Success[Success<br/>Response]
    Error[Error<br/>Response]

    Request --> Validate
    Validate --> ValidCheck

    ValidCheck -->|No| Error
    ValidCheck -->|Yes| Execute

    Execute --> ExecCheck
    ExecCheck -->|No| Error
    ExecCheck -->|Yes| Storage

    Storage --> StorageCheck
    StorageCheck -->|No| Fallback
    StorageCheck -->|Yes| Success

    Fallback -->|Retry| Storage
    Fallback -->|Give Up| Error

    style Error fill:#ffe1e1
    style Success fill:#e8f5e9
    style Fallback fill:#fff4e1
```

---

## Resource Management

```mermaid
flowchart LR
    subgraph "Resource Limits"
        MaxSearch[MaxSearchResults<br/>1000]
        MaxIndex[MaxIndexSize<br/>100000]
        MaxBranch[MaxBranches<br/>Unlimited]
    end

    subgraph "Memory Management"
        direction TB
        Cache[Write-Through<br/>Cache]
        Evict[LRU Eviction]
        GC[Go Garbage<br/>Collector]
    end

    subgraph "Cleanup"
        Session[Session<br/>Cleanup]
        Reset[Storage<br/>Reset]
    end

    MaxSearch & MaxIndex & MaxBranch --> Cache
    Cache --> Evict
    Evict --> GC

    Session --> Reset
    Reset -.-> Cache

    style MaxSearch fill:#ffe1e1
    style MaxIndex fill:#ffe1e1
    style Cache fill:#e8f5e9
```

---

## Summary

### Key Architectural Principles

1. **Modular Design**: Clear separation of concerns across 15 packages
2. **Interface-Based**: Storage and reasoning components use interfaces for testability
3. **Thread-Safe**: RWMutex protection for concurrent access
4. **Pluggable Storage**: Factory pattern enables multiple storage backends
5. **MCP Protocol**: Standard JSON-RPC communication via stdio transport
6. **Resource-Conscious**: Limits and caching prevent resource exhaustion
7. **Extensible**: Mode registry and tool registration support additions
8. **Metadata-Driven**: Rich metadata guides tool usage and integration

### Performance Optimizations

- **Write-through caching** for SQLite performance
- **FTS5 full-text search** for fast queries
- **WAL mode** for concurrent database reads
- **Prepared statements** for SQL efficiency
- **Deep copy strategy** prevents data races
- **Resource limits** prevent DoS attacks

### Production Readiness

- Comprehensive test coverage (75% overall, 81.2% handlers)
- Thread-safe concurrent operations
- Graceful error handling and fallbacks
- Resource limits and DoS protection
- Multiple storage backends (memory + SQLite)
- Extensive validation and quality checks

---

**For More Information**:
- [API Reference](API_REFERENCE.md) - Complete tool documentation
- [Project Index](PROJECT_INDEX.md) - Project structure and components
- [Integration Test Report](MCP_INTEGRATION_TEST_REPORT.md) - End-to-end validation
- [README](README.md) - Quick start and configuration
