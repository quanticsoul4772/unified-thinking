# Type Safety Enhancement Plan

## Implementation Status

| Phase | Status | Progress |
|-------|--------|----------|
| Phase 1: Centralized Types | **COMPLETED** | 100% |
| Phase 2: Handler Refactor | **COMPLETED** | 100% |
| Phase 3: ContextBridge | **COMPLETED** | 100% |
| Phase 4: Orchestration | **COMPLETED** | 100% |
| Phase 5: Metadata Distinct Type | **COMPLETED** | 100% |

**Last Updated**: December 4, 2024

---

## Problem Statement

**Excessive `map[string]interface{}` Usage**: ~980 occurrences across the codebase reduce type safety and make refactoring error-prone.

### Current Distribution

| Directory | Count | Priority |
|-----------|-------|----------|
| `internal/server` | 334 | HIGH |
| `internal/orchestration` | 166 | MEDIUM |
| `internal/modes` | 96 | MEDIUM |
| `internal/contextbridge` | 83 | HIGH |
| `internal/storage` | 58 | LOW |
| `internal/reasoning` | 48 | MEDIUM |
| `internal/types` | 43 | LOW |
| `benchmarks/` | 30+ | LOW |
| Other | ~120 | LOW |

---

## Root Cause Analysis

### 1. Legacy Handler Pattern
**Problem**: Many handlers accept `params map[string]interface{}` instead of typed structs.

**Example (OLD pattern - `abductive.go:57`)**:
```go
func (h *AbductiveHandler) HandleGenerateHypotheses(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
    var req GenerateHypothesesRequest
    if err := unmarshalParams(params, &req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    // ... uses typed req from here
}
```

**Example (NEW pattern - `thinking.go:93`)**:
```go
func (h *ThinkingHandler) HandleThink(ctx context.Context, req *mcp.CallToolRequest, input ThinkRequest) (*mcp.CallToolResult, *ThinkResponse, error) {
    // Already typed - no unmarshal needed
}
```

### 2. Response Building
**Problem**: Response maps built with `map[string]interface{}` instead of typed response structs.

### 3. ContextBridge Generic Interface
**Problem**: `EnrichResponse` returns `interface{}` and uses `map[string]interface{}` for composition.

### 4. MCP SDK Constraint
**Constraint**: MCP SDK's `CallToolRequest.Params` returns `map[string]interface{}` - this is an external boundary we cannot change.

---

## Design Solution

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     External Boundary (MCP SDK)                  │
│                  params map[string]interface{}                   │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Typed Adapter Layer (NEW)                     │
│            internal/server/types/request_types.go                │
│                                                                  │
│   - UnmarshalTypedRequest[T any](params) (T, error)             │
│   - Per-handler typed request structs                            │
│   - Validation at boundary                                       │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Handler Layer (TYPED)                        │
│                                                                  │
│   HandleFoo(ctx, req TypedRequest) (*TypedResponse, error)      │
│   - All internal logic uses typed structs                        │
│   - No map[string]interface{} after adapter layer               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Response Layer (TYPED)                        │
│                                                                  │
│   TypedResponse → toJSONContent() → mcp.CallToolResult          │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 1: Create Centralized Request/Response Types

**Location**: `internal/server/types/`

**New Files**:
```
internal/server/types/
├── requests.go       # All handler request types
├── responses.go      # All handler response types
├── adapters.go       # Generic unmarshal adapter
└── validation.go     # Request validation utilities
```

**Generic Adapter Pattern**:
```go
// internal/server/types/adapters.go

// UnmarshalRequest converts untyped params to typed request
func UnmarshalRequest[T any](params map[string]interface{}) (T, error) {
    var req T
    data, err := json.Marshal(params)
    if err != nil {
        return req, fmt.Errorf("marshal params: %w", err)
    }
    if err := json.Unmarshal(data, &req); err != nil {
        return req, fmt.Errorf("unmarshal to %T: %w", req, err)
    }
    return req, nil
}

// Alternative: Use mapstructure for better performance
func UnmarshalRequestFast[T any](params map[string]interface{}) (T, error) {
    var req T
    decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
        Result:           &req,
        TagName:          "json",
        WeaklyTypedInput: true,
    })
    if err := decoder.Decode(params); err != nil {
        return req, fmt.Errorf("decode to %T: %w", req, err)
    }
    return req, nil
}
```

### Phase 2: Refactor Handlers (High Priority)

**Target**: 16 handlers in `internal/server/handlers/` using old pattern

| Handler | Functions | Priority |
|---------|-----------|----------|
| `abductive.go` | 2 | HIGH |
| `backtracking.go` | 3 | HIGH |
| `case_based.go` | 2 | HIGH |
| `dual_process.go` | 1 | MEDIUM |
| `episodic.go` | 5 | HIGH |
| `symbolic.go` | 2 | MEDIUM |
| `unknown_unknowns.go` | 1 | LOW |

**Refactoring Pattern**:

**BEFORE**:
```go
func (h *Handler) HandleFoo(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
    var req FooRequest
    if err := unmarshalParams(params, &req); err != nil {
        return nil, err
    }
    // ... logic using req
}
```

**AFTER**:
```go
// Handler method becomes internal (private)
func (h *Handler) handleFoo(ctx context.Context, req FooRequest) (*FooResponse, error) {
    // ... logic using typed req, returns typed response
}

// Public method wraps with adapter
func (h *Handler) HandleFoo(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
    req, err := types.UnmarshalRequest[FooRequest](params)
    if err != nil {
        return nil, err
    }
    resp, err := h.handleFoo(ctx, req)
    if err != nil {
        return nil, err
    }
    return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}
```

### Phase 3: Refactor ContextBridge

**Problem**: 83 occurrences in `internal/contextbridge/`

**Current**:
```go
func (cb *ContextBridge) EnrichResponse(
    result interface{},
    params map[string]interface{},
) (interface{}, error)
```

**Solution**: Define typed context bridge data structures

```go
// internal/contextbridge/types.go

// ContextBridgeData is the structured context data
type ContextBridgeData struct {
    SimilarThoughts   []SimilarThought `json:"similar_thoughts,omitempty"`
    RelatedPatterns   []RelatedPattern `json:"related_patterns,omitempty"`
    RelevantDecisions []Decision       `json:"relevant_decisions,omitempty"`
}

// EnrichedResponse wraps any result with context bridge data
type EnrichedResponse[T any] struct {
    Result        T                  `json:"result"`
    ContextBridge *ContextBridgeData `json:"context_bridge,omitempty"`
}

// EnrichTypedResponse returns typed enriched response
func (cb *ContextBridge) EnrichTypedResponse[T any](
    result T,
    params ContextBridgeParams,
) (*EnrichedResponse[T], error) {
    // ... typed implementation
}
```

### Phase 4: Refactor Orchestration (166 occurrences)

**Pattern**: Workflow step definitions use `map[string]interface{}` for flexible inputs.

**Solution**: Define step input/output interfaces

```go
// internal/orchestration/types.go

// StepInput defines typed input for workflow steps
type StepInput interface {
    StepName() string
    Validate() error
}

// StepOutput defines typed output from workflow steps
type StepOutput interface {
    StepName() string
    Success() bool
}

// GenericStepData for truly dynamic data
type GenericStepData struct {
    Name   string         `json:"name"`
    Data   map[string]any `json:"data"` // Use any for Go 1.18+
}
```

### Phase 5: Standardize on `map[string]any`

**Go 1.18+ Migration**: Replace `map[string]interface{}` with `map[string]any` for cleaner syntax where untyped maps are still necessary.

**Scope**: 115 occurrences already use `any`, 980 use `interface{}`

**Command**:
```bash
# Find and replace (with review)
find . -name "*.go" -exec sed -i 's/map\[string\]interface{}/map[string]any/g' {} \;
```

---

## Implementation Plan

### Priority Matrix

| Phase | Impact | Effort | Risk | Priority |
|-------|--------|--------|------|----------|
| 1. Centralized Types | HIGH | LOW | LOW | P0 |
| 2. Handler Refactor | HIGH | MEDIUM | MEDIUM | P0 |
| 3. ContextBridge | MEDIUM | MEDIUM | LOW | P1 |
| 4. Orchestration | MEDIUM | HIGH | MEDIUM | P2 |
| 5. Syntax Standardize | LOW | LOW | LOW | P3 |

### Phase 1: Centralized Types (1-2 days)

**Tasks**:
1. Create `internal/server/types/` package
2. Define generic `UnmarshalRequest[T]` adapter
3. Consolidate existing request structs from handlers
4. Add validation utilities
5. Add tests for adapter

**Files to Create**:
- `internal/server/types/adapters.go`
- `internal/server/types/adapters_test.go`

**Success Criteria**:
- Generic adapter works with all existing request types
- 100% test coverage on adapter
- No functionality changes

### Phase 2: Handler Refactor (3-5 days)

**Tasks per Handler**:
1. Move request/response types to `types/` package if not already there
2. Create private typed implementation method
3. Update public method to use adapter
4. Update tests to use typed inputs
5. Verify no regressions

**Order of Refactoring** (by impact/risk):
1. `abductive.go` - Clean example, 2 methods
2. `backtracking.go` - 3 methods, well-contained
3. `symbolic.go` - 2 methods, isolated
4. `dual_process.go` - 1 method, simple
5. `case_based.go` - 2 methods, moderate complexity
6. `episodic.go` - 5 methods, most complex
7. `unknown_unknowns.go` - 1 method, simple

**Success Criteria**:
- All handlers use typed internal methods
- All tests pass
- No `unmarshalParams` in handler logic

### Phase 3: ContextBridge Refactor (2-3 days)

**Tasks**:
1. Define `ContextBridgeData` and related types
2. Create generic `EnrichedResponse[T]` wrapper
3. Add typed `EnrichTypedResponse[T]` method
4. Migrate callers progressively
5. Deprecate old `EnrichResponse`

**Success Criteria**:
- Type-safe context enrichment
- Backward compatible during migration
- Clear deprecation path

### Phase 4: Orchestration Refactor (3-5 days)

**Tasks**:
1. Analyze workflow step patterns
2. Define `StepInput`/`StepOutput` interfaces
3. Create typed step implementations
4. Refactor workflow execution
5. Update tests

**Success Criteria**:
- Workflow steps have typed contracts
- Dynamic data contained to specific types
- Improved IDE support

### Phase 5: Syntax Standardization (1 day)

**Tasks**:
1. Global replace `interface{}` → `any` in new code
2. Replace `map[string]interface{}` → `map[string]any` where appropriate
3. Update code style guide

**Success Criteria**:
- Consistent use of Go 1.18+ syntax
- No functionality changes

---

## Metrics & Success Criteria

### Quantitative Goals

| Metric | Current | Target | Notes |
|--------|---------|--------|-------|
| `map[string]interface{}` count | 980 | <200 | Remaining = SDK boundaries |
| Typed handler methods | 0% | 100% | All internal logic typed |
| Handler test coverage | ~75% | 90%+ | Typed tests easier to write |

### Qualitative Goals

1. **Refactoring Safety**: Changes caught at compile time, not runtime
2. **IDE Support**: Better autocomplete and error detection
3. **Documentation**: Types serve as self-documenting API contracts
4. **Maintainability**: Easier to understand data flow

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking changes | HIGH | Phased rollout, maintain backward compat |
| Performance regression | LOW | Benchmark before/after, use mapstructure |
| Test updates | MEDIUM | Update tests alongside handler changes |
| SDK constraint | LOW | Accept boundary remains untyped |

---

## Non-Goals

1. **Changing MCP SDK**: External dependency, out of scope
2. **100% elimination**: Some truly dynamic data needs `map[string]any`
3. **Benchmarks refactor**: Low priority, test code

---

## Related Issues

- LLM Client Consolidation (completed) - Similar pattern of type unification
- Handler Extraction project - Types should align with extracted handlers

---

## Appendix: Files to Modify

### Phase 1-2 Handler Files

```
internal/server/handlers/
├── abductive.go          # 2 handlers → typed
├── backtracking.go       # 3 handlers → typed
├── case_based.go         # 2 handlers → typed
├── dual_process.go       # 1 handler → typed
├── episodic.go           # 5 handlers → typed
├── symbolic.go           # 2 handlers → typed
└── unknown_unknowns.go   # 1 handler → typed
```

### Phase 3 ContextBridge Files

```
internal/contextbridge/
├── bridge.go             # Main refactor target
├── types.go              # NEW: Typed structures
└── bridge_test.go        # Update tests
```

### Phase 4 Orchestration Files

```
internal/orchestration/
├── workflow.go           # Step definitions
├── executor.go           # Execution logic
├── types.go              # NEW: Step interfaces
└── *_test.go             # Update tests
```

---

## Completed Implementation Details

### Phase 5: types.Metadata Distinct Type (COMPLETED)

**Location**: `internal/types/types.go`

The `types.Metadata` type is defined as a **distinct type** (not a type alias), providing compile-time type safety:

```go
// Metadata is a distinct type for arbitrary key-value metadata.
// This provides compile-time type safety - assignments between
// map[string]any and Metadata require explicit conversion.
type Metadata map[string]any
```

**Key Distinction**:
- `type Metadata map[string]any` = Distinct type (compile-time errors on mismatch)
- `type Metadata = map[string]any` = Type alias (no compile-time checking)

**Files Modified** (51 files):
- `internal/contextbridge/bridge.go`, `metrics.go`
- `internal/embeddings/*.go`
- `internal/integration/synthesizer.go`
- `internal/knowledge/*.go`
- `internal/metacognition/*.go`
- `internal/metrics/collector.go`
- `internal/orchestration/workflow.go`
- `internal/reasoning/*.go`
- `internal/server/handlers/*.go`
- `internal/validation/symbolic.go`

**Handling JSON Unmarshal Edge Cases**:

JSON unmarshaling naturally produces `map[string]interface{}`, not `types.Metadata`. For test code and external boundaries:

```go
// internal/orchestration/workflow.go - extractResultMap helper
func extractResultMap(result interface{}) map[string]interface{} {
    switch v := result.(type) {
    case types.Metadata:
        return map[string]interface{}(v)
    case map[string]interface{}:
        return v
    }
    return nil
}
```

### Phase 1: Centralized Request/Response Types (COMPLETED)

**Location**: `internal/server/types/`

**Files Created**:
- `adapters.go` - Generic unmarshal adapter with helpers
- `adapters_test.go` - Comprehensive test coverage

**Key Functions**:

```go
// Generic request unmarshaling
func UnmarshalRequest[T any](params map[string]interface{}) (T, error)

// With validation support
func UnmarshalRequestWithValidation[T Validatable](params map[string]interface{}) (T, error)

// Response marshaling
func MarshalResponse(resp interface{}) ([]byte, error)
func ToJSONContent(resp interface{}) string

// Safe extraction helpers
func ExtractString(params map[string]interface{}, key string) string
func ExtractInt(params map[string]interface{}, key string) int
func ExtractFloat(params map[string]interface{}, key string) float64
func ExtractBool(params map[string]interface{}, key string) bool
func ExtractStringSlice(params map[string]interface{}, key string) []string
func ExtractMap(params map[string]interface{}, key string) map[string]interface{}
```

### Test Validation

All 30 packages pass tests after the type safety refactoring:
- Build: PASS
- `go vet`: PASS
- `gofmt`: All 36 files formatted
- Tests: All packages OK

### Phase 2: Handler Refactor (COMPLETED)

**Location**: `internal/server/handlers/`

All 16 handlers across 7 files have been refactored to use the typed internal method pattern.

**Refactoring Pattern Applied**:

```go
// PUBLIC: Accepts map[string]interface{} at MCP boundary
func (h *Handler) HandleFoo(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
    req, err := unmarshalRequest[FooRequest](params)
    if err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    resp, err := h.foo(ctx, req)
    if err != nil {
        return nil, err
    }
    return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// PRIVATE: Fully typed internal implementation
func (h *Handler) foo(ctx context.Context, req FooRequest) (*FooResponse, error) {
    // All business logic with typed request/response
    return &FooResponse{...}, nil
}
```

**Files Modified**:

| File | Handlers Refactored | Private Methods Added |
|------|---------------------|----------------------|
| `unknown_unknowns.go` | 1 | `detectBlindSpots` |
| `symbolic.go` | 2 | `proveTheorem`, `checkConstraints` |
| `dual_process.go` | 1 | `dualProcessThink` |
| `backtracking.go` | 3 | `createCheckpoint`, `restoreCheckpoint`, `listCheckpoints` |
| `abductive.go` | 2 | `generateHypotheses`, `evaluateHypotheses` |
| `case_based.go` | 2 | `retrieveCases`, `performCBRCycle` |
| `episodic.go` | 5 | `startSession`, `completeSession`, `getRecommendations`, `searchTrajectories`, `analyzeTrajectory` |

**Benefits Achieved**:
- All internal logic is now fully typed
- Business logic separated from MCP boundary handling
- Easier to test private methods directly with typed inputs
- Better IDE support and compile-time error detection
- No `unmarshalParams` calls in business logic

**Validation**:
- `go build ./...`: PASS
- `go vet ./...`: PASS
- `go test ./... -count=1`: All 30 packages PASS
