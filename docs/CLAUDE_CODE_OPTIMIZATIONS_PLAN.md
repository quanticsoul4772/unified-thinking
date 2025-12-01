# Claude Code Optimizations Implementation Plan

**Version:** 1.0  
**Created:** December 1, 2025  
**Status:** Ready for Implementation  
**Estimated Effort:** 69 hours (~12 working days)  
**New MCP Tools:** 4 (export-session, import-session, list-presets, run-preset)

---

## Executive Summary

This document outlines the implementation plan for optimizing the unified-thinking MCP server for Claude Code and Claude Desktop usage. The optimizations focus on five key areas:

1. **Response Format Optimization** - Reduce token consumption by 40-60%
2. **Workflow Presets** - 8 built-in presets for common development tasks
3. **Structured Error System** - Actionable recovery suggestions for all errors
4. **Session Export/Import** - Preserve reasoning context across restarts
5. **Documentation Improvements** - Quick reference and workflow examples

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Component 1: Response Format Optimization](#component-1-response-format-optimization)
3. [Component 2: Workflow Presets](#component-2-workflow-presets)
4. [Component 3: Structured Error System](#component-3-structured-error-system)
5. [Component 4: Session Export/Import](#component-4-session-exportimport)
6. [Component 5: Documentation Improvements](#component-5-documentation-improvements)
7. [Implementation Timeline](#implementation-timeline)
8. [Success Metrics](#success-metrics)
9. [Risk Mitigation](#risk-mitigation)

---

## Architecture Overview

### New Package Structure

```
internal/claudecode/
├── format/
│   ├── formatter.go       # ResponseFormatter interface
│   ├── compact.go         # Compact format (40-60% reduction)
│   ├── minimal.go         # Minimal format (result only)
│   ├── options.go         # FormatOptions type
│   └── formatter_test.go
├── presets/
│   ├── registry.go        # Preset registration and lookup
│   ├── code_review.go     # Code review workflow
│   ├── debug.go           # Debug analysis workflow
│   ├── architecture.go    # ADR generation workflow
│   ├── research.go        # Research synthesis workflow
│   ├── refactoring.go     # Refactoring plan workflow
│   ├── testing.go         # Test strategy workflow
│   ├── documentation.go   # Documentation generation workflow
│   ├── incident.go        # Incident investigation workflow
│   └── presets_test.go
├── errors/
│   ├── structured.go      # StructuredError type
│   ├── codes.go           # Error code constants
│   ├── recovery.go        # Recovery suggestion generator
│   └── errors_test.go
├── session/
│   ├── exporter.go        # Session serialization
│   ├── importer.go        # Session restoration
│   ├── schema.go          # Export schema versioning
│   ├── merge.go           # Import merge strategies
│   └── session_test.go
└── claudecode.go          # Package initialization
```

### Integration Points

| Component | Integrates With | Method |
|-----------|-----------------|--------|
| Format Options | server.go | Tool handler wrapper |
| Presets | orchestration/ | Workflow registry |
| Errors | All 23 handlers | Replace error returns |
| Session | storage/ | Serialize/deserialize |

---

## Component 1: Response Format Optimization

### Purpose
Reduce token consumption in agentic workflows where Claude Code makes many sequential tool calls.

### Format Levels

```go
type FormatLevel string

const (
    FormatFull    FormatLevel = "full"    // Default - all metadata
    FormatCompact FormatLevel = "compact" // Reduced - no context_bridge details
    FormatMinimal FormatLevel = "minimal" // Result only - no metadata
)

type FormatOptions struct {
    Level           FormatLevel `json:"level"`
    IncludeMetadata bool        `json:"include_metadata"`
    IncludeTimings  bool        `json:"include_timings"`
    MaxArrayLength  int         `json:"max_array_length"` // Truncate long arrays
}
```

### Format Comparison

**Full Format (Current - Default):**
```json
{
    "context_bridge": {
        "match_count": 0,
        "matches": [],
        "recommendation": "",
        "similarity_mode": "semantic_embedding",
        "status": "no_matches",
        "version": "1.0"
    },
    "result": {
        "thought_id": "thought-1234567890-1",
        "mode": "linear",
        "status": "success",
        "confidence": 0.85,
        "is_valid": true,
        "metadata": {
            "suggested_next_tools": [...],
            "action_recommendations": [...],
            "export_formats": {...}
        }
    }
}
```

**Compact Format (40-60% smaller):**
```json
{
    "thought_id": "thought-1234567890-1",
    "mode": "linear",
    "confidence": 0.85,
    "is_valid": true,
    "next_tools": ["validate", "think"]
}
```

**Minimal Format (80%+ smaller):**
```json
{
    "thought_id": "thought-1234567890-1",
    "confidence": 0.85
}
```

### Implementation

```go
// format/formatter.go
type ResponseFormatter interface {
    Format(response any) (any, error)
}

// format/compact.go
type CompactFormatter struct {
    opts FormatOptions
}

func (f *CompactFormatter) Format(response any) (any, error) {
    // Remove context_bridge entirely
    // Flatten metadata.suggested_next_tools to next_tools array
    // Remove empty arrays and null values
    // Truncate arrays longer than MaxArrayLength
}
```

### Tool Parameter Addition

All tools gain an optional `format` parameter:

```json
{
    "tool": "think",
    "params": {
        "content": "Analyze this code",
        "mode": "linear",
        "format": "compact"
    }
}
```

---

## Component 2: Workflow Presets

### Purpose
Provide ready-to-use tool sequences for common Claude Code tasks, teaching best practices through example.

### Preset Schema

```go
type WorkflowPreset struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Description string              `json:"description"`
    Category    string              `json:"category"`
    Steps       []PresetStep        `json:"steps"`
    InputSchema map[string]ParamSpec `json:"input_schema"`
    OutputFormat string             `json:"output_format"`
    EstimatedTime string            `json:"estimated_time"`
}

type PresetStep struct {
    Tool        string            `json:"tool"`
    Description string            `json:"description"`
    InputMap    map[string]string `json:"input_map"`
    Condition   string            `json:"condition,omitempty"`
    StoreAs     string            `json:"store_as"`
}
```

### Built-in Presets

#### 1. code-review
**Category:** Code  
**Purpose:** Systematic code review with documented reasoning

| Step | Tool | Description |
|------|------|-------------|
| 1 | decompose-problem | Break code into reviewable chunks |
| 2 | think (linear) | Analyze each chunk |
| 3 | detect-fallacies | Check for logical errors |
| 4 | detect-biases | Identify review blind spots |
| 5 | make-decision | Approve/request changes |

#### 2. debug-analysis
**Category:** Code  
**Purpose:** Structured debugging with causal analysis

| Step | Tool | Description |
|------|------|-------------|
| 1 | start-reasoning-session | Track debug session |
| 2 | build-causal-graph | Map potential causes |
| 3 | generate-hypotheses | Create hypothesis list |
| 4 | simulate-intervention | Test fix effectiveness |
| 5 | complete-reasoning-session | Document solution |

#### 3. architecture-decision
**Category:** Architecture  
**Purpose:** Generate Architecture Decision Records

| Step | Tool | Description |
|------|------|-------------|
| 1 | analyze-perspectives | All stakeholder views |
| 2 | analyze-temporal | Short vs long-term |
| 3 | make-decision | Multi-criteria evaluation |
| 4 | detect-blind-spots | Find gaps |
| 5 | think (linear) | Generate ADR |

#### 4. research-synthesis
**Category:** Research  
**Purpose:** Deep research with Graph-of-Thoughts

| Step | Tool | Description |
|------|------|-------------|
| 1 | got-initialize | Start exploration |
| 2 | got-generate (k=3) | Multiple paths |
| 3 | got-aggregate | Synthesize findings |
| 4 | got-refine | Polish synthesis |
| 5 | got-finalize | Extract conclusions |

#### 5. refactoring-plan
**Category:** Code  
**Purpose:** Safe refactoring with risk assessment

| Step | Tool | Description |
|------|------|-------------|
| 1 | decompose-problem | Break into phases |
| 2 | build-causal-graph | Map dependencies |
| 3 | sensitivity-analysis | Assess risk |
| 4 | analyze-temporal | Plan timing |
| 5 | make-decision | Prioritize order |

#### 6. test-strategy
**Category:** Testing  
**Purpose:** Comprehensive test planning

| Step | Tool | Description |
|------|------|-------------|
| 1 | decompose-problem | Identify components |
| 2 | analyze-perspectives | Tester/dev/user views |
| 3 | detect-blind-spots | Find untested scenarios |
| 4 | generate-hypotheses | Predict failures |
| 5 | think (tree) | Generate test cases |

#### 7. documentation-gen
**Category:** Documentation  
**Purpose:** Generate docs from code analysis

| Step | Tool | Description |
|------|------|-------------|
| 1 | think (linear) | Analyze code |
| 2 | analyze-perspectives | Consider audiences |
| 3 | decompose-argument | Structure outline |
| 4 | synthesize-insights | Combine perspectives |
| 5 | think (linear) | Generate docs |

#### 8. incident-investigation
**Category:** Operations  
**Purpose:** Post-incident analysis (postmortem)

| Step | Tool | Description |
|------|------|-------------|
| 1 | start-reasoning-session | Track investigation |
| 2 | build-causal-graph | Map timeline/causes |
| 3 | generate-counterfactual | "What if" scenarios |
| 4 | detect-biases | Check hindsight bias |
| 5 | make-decision | Recommend prevention |
| 6 | complete-reasoning-session | Generate postmortem |

### New MCP Tools

**list-presets**
```json
{
    "tool": "list-presets",
    "params": {
        "category": "code"  // Optional filter
    }
}
```

**run-preset**
```json
{
    "tool": "run-preset",
    "params": {
        "preset": "code-review",
        "context": {
            "code": "function example() {...}",
            "focus": "performance"
        },
        "format": "compact"
    }
}
```

---

## Component 3: Structured Error System

### Purpose
Enable Claude Code to automatically recover from errors without human intervention.

### Error Code Categories

```go
// errors/codes.go
const (
    // Resource errors (1xxx)
    ErrThoughtNotFound     = "ERR_1001_THOUGHT_NOT_FOUND"
    ErrBranchNotFound      = "ERR_1002_BRANCH_NOT_FOUND"
    ErrSessionNotFound     = "ERR_1003_SESSION_NOT_FOUND"
    ErrGraphNotFound       = "ERR_1004_GRAPH_NOT_FOUND"
    ErrCheckpointNotFound  = "ERR_1005_CHECKPOINT_NOT_FOUND"
    ErrDecisionNotFound    = "ERR_1006_DECISION_NOT_FOUND"
    
    // Validation errors (2xxx)
    ErrInvalidParameter    = "ERR_2001_INVALID_PARAMETER"
    ErrMissingRequired     = "ERR_2002_MISSING_REQUIRED"
    ErrInvalidMode         = "ERR_2003_INVALID_MODE"
    ErrInvalidConfidence   = "ERR_2004_INVALID_CONFIDENCE"
    ErrInvalidFormat       = "ERR_2005_INVALID_FORMAT"
    
    // State errors (3xxx)
    ErrSessionActive       = "ERR_3001_SESSION_ALREADY_ACTIVE"
    ErrSessionNotActive    = "ERR_3002_SESSION_NOT_ACTIVE"
    ErrBranchLocked        = "ERR_3003_BRANCH_LOCKED"
    ErrGraphFinalized      = "ERR_3004_GRAPH_ALREADY_FINALIZED"
    
    // External errors (4xxx)
    ErrEmbeddingFailed     = "ERR_4001_EMBEDDING_FAILED"
    ErrNeo4jConnection     = "ERR_4002_NEO4J_CONNECTION"
    ErrLLMFailed           = "ERR_4003_LLM_CALL_FAILED"
    ErrStorageFailed       = "ERR_4004_STORAGE_OPERATION"
    
    // Limit errors (5xxx)
    ErrRateLimited         = "ERR_5001_RATE_LIMITED"
    ErrContextTooLarge     = "ERR_5002_CONTEXT_TOO_LARGE"
    ErrTooManyBranches     = "ERR_5003_TOO_MANY_BRANCHES"
    ErrMaxDepthReached     = "ERR_5004_MAX_DEPTH_REACHED"
)
```

### StructuredError Type

```go
type StructuredError struct {
    Code                string   `json:"error_code"`
    Message             string   `json:"message"`
    Details             string   `json:"details,omitempty"`
    RecoverySuggestions []string `json:"recovery_suggestions"`
    RelatedTools        []string `json:"related_tools,omitempty"`
    ExampleFix          any      `json:"example_fix,omitempty"`
}

// Builder pattern
func NewStructuredError(code, message string) *StructuredError {
    return &StructuredError{Code: code, Message: message}
}

func (e *StructuredError) WithDetails(d string) *StructuredError {
    e.Details = d
    return e
}

func (e *StructuredError) WithExample(tool string, params map[string]any) *StructuredError {
    e.ExampleFix = map[string]any{"tool": tool, "params": params}
    return e
}
```

### Example Error Responses

**Resource Not Found:**
```json
{
    "error_code": "ERR_1001_THOUGHT_NOT_FOUND",
    "message": "Thought with ID 'thought-abc123' not found",
    "details": "The thought may have been deleted or the ID is malformed",
    "recovery_suggestions": [
        "Use 'search' tool to find thoughts matching your criteria",
        "Use 'history' tool to list recent thoughts",
        "Verify the thought_id format matches 'thought-XXXXXXXXXX-N'"
    ],
    "related_tools": ["search", "history", "list-branches"],
    "example_fix": {
        "tool": "search",
        "params": {"query": "your search terms", "limit": 10}
    }
}
```

**Missing Parameter:**
```json
{
    "error_code": "ERR_2002_MISSING_REQUIRED",
    "message": "Required parameter 'content' is missing",
    "details": "The 'think' tool requires a 'content' parameter",
    "recovery_suggestions": [
        "Add the 'content' parameter with your reasoning prompt"
    ],
    "related_tools": [],
    "example_fix": {
        "tool": "think",
        "params": {
            "content": "Your reasoning prompt here",
            "mode": "linear"
        }
    }
}
```

**External Service Failed:**
```json
{
    "error_code": "ERR_4001_EMBEDDING_FAILED",
    "message": "Failed to generate embeddings",
    "details": "Voyage AI API returned error: rate limited",
    "recovery_suggestions": [
        "Wait 60 seconds and retry the operation",
        "Check VOYAGE_API_KEY environment variable",
        "Operation will continue with hash-based fallback"
    ],
    "related_tools": ["get-metrics"],
    "example_fix": null
}
```

---

## Component 4: Session Export/Import

### Purpose
Preserve reasoning context across Claude Code session restarts.

### Export Schema

```go
type SessionExport struct {
    Version      string             `json:"version"`       // "1.0"
    ExportedAt   time.Time          `json:"exported_at"`
    SessionID    string             `json:"session_id"`
    
    // Core Data
    Thoughts     []types.Thought    `json:"thoughts"`
    Branches     []types.Branch     `json:"branches"`
    Insights     []types.Insight    `json:"insights,omitempty"`
    
    // Optional Components
    Decisions    []DecisionExport   `json:"decisions,omitempty"`
    CausalGraphs []CausalGraphExport `json:"causal_graphs,omitempty"`
    Checkpoints  []CheckpointExport `json:"checkpoints,omitempty"`
    
    // Metadata
    ToolsUsed    []string           `json:"tools_used"`
    Duration     time.Duration      `json:"duration"`
    ThoughtCount int                `json:"thought_count"`
}
```

### Merge Strategies

| Strategy | Behavior |
|----------|----------|
| `replace` | Clear existing data, import new |
| `merge` | Update existing, add new (default) |
| `append` | Keep existing, add new with new IDs |

### New MCP Tools

**export-session**
```json
{
    "tool": "export-session",
    "params": {
        "session_id": "debug-session-001",
        "include_decisions": true,
        "include_causal_graphs": true,
        "compress": false
    }
}
```

Response:
```json
{
    "export_data": "base64-encoded-json-or-raw-json",
    "size_bytes": 12458,
    "thought_count": 15,
    "export_version": "1.0"
}
```

**import-session**
```json
{
    "tool": "import-session",
    "params": {
        "data": "base64-or-json-string",
        "merge_strategy": "merge",
        "validate_only": false
    }
}
```

Response:
```json
{
    "session_id": "debug-session-001",
    "imported_thoughts": 15,
    "imported_branches": 3,
    "merged_count": 2,
    "conflicts_resolved": 0,
    "status": "success"
}
```

---

## Component 5: Documentation Improvements

### CLAUDE.md Additions

#### Quick Reference Table

```markdown
## Quick Reference

| Tool | Purpose | Key Params |
|------|---------|------------|
| think | Main reasoning | content, mode, confidence |
| validate | Check logic | thought_id |
| search | Find thoughts | query, limit |
| make-decision | Multi-criteria | question, options, criteria |
| decompose-problem | Break down | problem |
| build-causal-graph | Causality | description, observations |
| got-initialize | Start GoT | graph_id, initial_thought |
| start-reasoning-session | Track session | session_id, description |
| export-session | Save state | session_id |
| run-preset | Execute workflow | preset, context |
```

#### Workflow Examples

```markdown
## Workflow Examples

### Code Review Workflow
```
1. decompose-problem {"problem": "Review PR #123 changes"}
2. think {"content": "Analyze chunk 1...", "mode": "linear"}
3. detect-fallacies {"content": "Implementation logic..."}
4. make-decision {"question": "Approve PR?", ...}
```

### Debug Workflow
```
1. start-reasoning-session {"session_id": "bug-123", "domain": "debugging"}
2. build-causal-graph {"description": "Bug causes", "observations": [...]}
3. generate-hypotheses {"observations": [...]}
4. complete-reasoning-session {"status": "success", "solution": "..."}
```
```

#### Performance Tips

```markdown
## Performance Tips

1. **Use compact format for agentic workflows:**
   ```json
   {"tool": "think", "params": {"content": "...", "format": "compact"}}
   ```

2. **Chain presets for complex tasks:**
   - Start with `run-preset` for common patterns
   - Fall back to individual tools for customization

3. **Export sessions before long breaks:**
   - Prevents context loss on restart
   - Enables resumption of complex analysis
```

---

## Implementation Timeline

### Phase 1: Foundation (Days 1-3)

**Day 1 - Core Types (4 hours)**
- [ ] Create internal/claudecode/ package structure
- [ ] Define FormatLevel enum and FormatOptions
- [ ] Define StructuredError type with builder
- [ ] Define SessionExport schema

**Day 2 - Formatters (6 hours)**
- [ ] Implement ResponseFormatter interface
- [ ] Implement CompactFormatter
- [ ] Implement MinimalFormatter
- [ ] Add formatter unit tests (90%+ coverage)

**Day 3 - Error System (5 hours)**
- [ ] Implement all error code constants
- [ ] Implement RecoveryGenerator (10 error mappings)
- [ ] Add handler integration helper
- [ ] Add error system unit tests

### Phase 2: Core Features (Days 4-7)

**Day 4 - Format Integration (6 hours)**
- [ ] Add format parameter parsing to server
- [ ] Integrate with tool handler wrapper
- [ ] Update 5 high-frequency tools
- [ ] Add integration tests

**Day 5 - Presets (7 hours)**
- [ ] Design preset workflow schema
- [ ] Implement preset registry
- [ ] Create presets: code-review, debug-analysis, architecture-decision
- [ ] Add preset execution tests

**Day 6 - Session Management (6 hours)**
- [ ] Implement session exporter
- [ ] Implement session importer
- [ ] Add merge strategies
- [ ] Add version migration support

**Day 7 - Error Integration (8 hours)**
- [ ] Update all 23 handler files
- [ ] Replace raw errors with StructuredError
- [ ] Add recovery suggestions (top 15 errors)
- [ ] Add integration tests

### Phase 3: Polish & Testing (Days 8-10)

**Day 8 - Remaining Presets (6 hours)**
- [ ] Create: research-synthesis, refactoring-plan
- [ ] Create: test-strategy, documentation-gen, incident-investigation
- [ ] Implement list-presets tool
- [ ] Implement run-preset tool

**Day 9 - Session Tools (7 hours)**
- [ ] Implement export-session MCP tool
- [ ] Implement import-session MCP tool
- [ ] Add full round-trip integration tests
- [ ] Extend format support to all 75 tools

**Day 10 - Testing & Cleanup (6 hours)**
- [ ] Comprehensive integration testing
- [ ] Performance benchmarking
- [ ] Bug fixes
- [ ] Code review

### Phase 4: Documentation (Days 11-12)

**Day 11 - CLAUDE.md (4 hours)**
- [ ] Add Quick Reference Table (all 79 tools)
- [ ] Add Workflow Examples section
- [ ] Add Performance Tips section
- [ ] Add Session Management section

**Day 12 - Other Docs (4 hours)**
- [ ] Update README.md (tool count, features)
- [ ] Update API_REFERENCE.md (4 new tools)
- [ ] Create PRESETS.md
- [ ] Create UPGRADE.md migration guide

---

## Success Metrics

### Quantitative Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Response size (compact) | 40-60% reduction | JSON bytes comparison |
| Token consumption | 30-50% reduction | Token count per workflow |
| Error recovery rate | 80%+ actionable | Audit all error paths |
| Session fidelity | 100% round-trip | Export/import diff test |
| Preset coverage | 8 presets | Implementation count |
| Test coverage | 85%+ new code | go test -cover |

### Acceptance Criteria

**Phase 1 Complete:**
- [ ] Package compiles with all types defined
- [ ] FormatOptions and StructuredError functional

**Phase 2 Complete:**
- [ ] format=compact reduces responses by 40%+
- [ ] 3 presets execute via execute-workflow
- [ ] Session export produces valid JSON
- [ ] Structured errors include suggestions

**Phase 3 Complete:**
- [ ] All 8 presets implemented
- [ ] All handlers use StructuredError
- [ ] Export/import passes all tests
- [ ] All 79 tools support format

**Phase 4 Complete:**
- [ ] CLAUDE.md has quick reference
- [ ] All docs updated
- [ ] CHANGELOG entry added

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Format changes break clients | High | Default to "full", test backward compat |
| Preset workflows too rigid | Medium | Allow parameter overrides per step |
| Session export too large | Medium | Add compression, optional field exclusion |
| Error suggestions incorrect | Low | Start conservative, iterate on feedback |
| Timeline slips | Medium | Phase 1-2 are MVP, 3-4 can defer |

---

## Appendix A: File Checklist

### New Files to Create

```
internal/claudecode/claudecode.go
internal/claudecode/format/formatter.go
internal/claudecode/format/compact.go
internal/claudecode/format/minimal.go
internal/claudecode/format/options.go
internal/claudecode/format/formatter_test.go
internal/claudecode/presets/registry.go
internal/claudecode/presets/code_review.go
internal/claudecode/presets/debug.go
internal/claudecode/presets/architecture.go
internal/claudecode/presets/research.go
internal/claudecode/presets/refactoring.go
internal/claudecode/presets/testing.go
internal/claudecode/presets/documentation.go
internal/claudecode/presets/incident.go
internal/claudecode/presets/presets_test.go
internal/claudecode/errors/structured.go
internal/claudecode/errors/codes.go
internal/claudecode/errors/recovery.go
internal/claudecode/errors/errors_test.go
internal/claudecode/session/exporter.go
internal/claudecode/session/importer.go
internal/claudecode/session/schema.go
internal/claudecode/session/merge.go
internal/claudecode/session/session_test.go
internal/server/handlers/presets.go
internal/server/handlers/presets_test.go
internal/server/handlers/session_export.go
internal/server/handlers/session_export_test.go
docs/PRESETS.md
docs/UPGRADE.md
```

### Files to Modify

```
internal/server/server.go          # Add format wrapper
internal/server/tools.go           # Add 4 new tools
internal/server/handlers/*.go      # All 23 handlers for structured errors
CLAUDE.md                          # Quick reference, examples
README.md                          # Tool count, features
API_REFERENCE.md                   # New tool documentation
CHANGELOG.md                       # Version entry
```

---

## Appendix B: API Contracts

### export-session

**Request:**
```json
{
    "session_id": "string (required)",
    "include_decisions": "boolean (default: true)",
    "include_causal_graphs": "boolean (default: true)",
    "include_checkpoints": "boolean (default: true)",
    "compress": "boolean (default: false)"
}
```

**Response:**
```json
{
    "export_data": "string (JSON or base64)",
    "size_bytes": "number",
    "thought_count": "number",
    "branch_count": "number",
    "export_version": "string"
}
```

### import-session

**Request:**
```json
{
    "data": "string (JSON or base64, required)",
    "merge_strategy": "replace|merge|append (default: merge)",
    "validate_only": "boolean (default: false)"
}
```

**Response:**
```json
{
    "session_id": "string",
    "imported_thoughts": "number",
    "imported_branches": "number",
    "merged_count": "number",
    "conflicts_resolved": "number",
    "validation_errors": "array (if validate_only)",
    "status": "success|partial|failed"
}
```

### list-presets

**Request:**
```json
{
    "category": "string (optional filter)"
}
```

**Response:**
```json
{
    "presets": [
        {
            "id": "string",
            "name": "string",
            "description": "string",
            "category": "string",
            "step_count": "number",
            "estimated_time": "string"
        }
    ],
    "count": "number"
}
```

### run-preset

**Request:**
```json
{
    "preset": "string (preset ID, required)",
    "context": "object (input parameters)",
    "format": "full|compact|minimal (default: full)",
    "stop_on_error": "boolean (default: true)"
}
```

**Response:**
```json
{
    "preset_id": "string",
    "steps_completed": "number",
    "steps_total": "number",
    "results": [
        {
            "step": "number",
            "tool": "string",
            "result": "object",
            "duration_ms": "number"
        }
    ],
    "final_output": "object",
    "status": "success|partial|failed"
}
```

---

*Document generated using unified-thinking reasoning session: claude-code-optimization-plan-2024*
