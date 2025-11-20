# Tool Registration Architecture

## Overview

The unified-thinking MCP server uses a **separated tool definition and registration pattern** to improve maintainability and reduce code duplication. This document explains the architecture and provides guidelines for adding new tools.

## Architecture Pattern

### Three-Layer Separation

1. **Tool Definitions** (`internal/server/tools.go`) - Schema and documentation
2. **Handler Implementation** (`internal/server/handlers/*.go`) - Business logic
3. **Tool Registration** (`internal/server/server.go`) - Wiring layer

This separation follows the **Single Responsibility Principle**: each file has one clear purpose.

## File Structure

### 1. Tool Definitions (`internal/server/tools.go`)

**Purpose**: Central registry of all MCP tool schemas and documentation

**Contents**:
- Tool names and descriptions
- Parameter specifications
- Return value documentation
- Usage examples and workflows
- Integration patterns with other MCP servers

**Example**:
```go
var ToolDefinitions = []mcp.Tool{
    {
        Name: "probabilistic-reasoning",
        Description: `Perform Bayesian inference and update probabilistic beliefs.

**Parameters:**
- operation (required): "create", "update", "get", or "combine"
- belief_id (for update/get): Belief identifier
- likelihood_if_true (for update): P(E|H) - probability if hypothesis true
- likelihood_if_false (for update): P(E|¬H) - probability if hypothesis false

**Example:**
{"operation": "create", "statement": "X is true", "prior_prob": 0.5}`,
    },
}
```

**Benefits**:
- Single source of truth for tool schemas
- Easy to review all tool interfaces
- Documentation lives with definitions
- Can generate API docs from this file

### 2. Handler Implementation (`internal/server/handlers/*.go`)

**Purpose**: Modular business logic for each tool domain

**Structure**:
```
internal/server/handlers/
├── thinking.go           # Core thinking tools (think, history, search)
├── branches.go           # Branch management tools
├── validation.go         # Logical validation tools
├── probabilistic.go      # Bayesian reasoning tools
├── decision.go           # Decision-making tools
├── metacognition.go      # Self-evaluation and bias detection
├── temporal.go           # Temporal reasoning tools
├── causal.go             # Causal inference tools
├── episodic.go           # Episodic memory and learning
└── [other domain handlers]
```

**Handler Pattern**:
```go
// Handler struct with dependencies
type ProbabilisticHandler struct {
    storage               storage.Storage
    probabilisticReasoner *reasoning.ProbabilisticReasoner
    evidenceAnalyzer      *analysis.EvidenceAnalyzer
}

// Constructor
func NewProbabilisticHandler(...) *ProbabilisticHandler {
    return &ProbabilisticHandler{...}
}

// Request/Response types
type ProbabilisticReasoningRequest struct {
    Operation         string  `json:"operation"`
    LikelihoodIfTrue  float64 `json:"likelihood_if_true,omitempty"`
    LikelihoodIfFalse float64 `json:"likelihood_if_false,omitempty"`
}

type ProbabilisticReasoningResponse struct {
    Belief    *types.ProbabilisticBelief `json:"belief,omitempty"`
    Operation string                     `json:"operation"`
    Status    string                     `json:"status"`
}

// Tool handler method
func (h *ProbabilisticHandler) HandleProbabilisticReasoning(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input ProbabilisticReasoningRequest,
) (*mcp.CallToolResult, *ProbabilisticReasoningResponse, error) {
    // 1. Validation
    if err := ValidateProbabilisticReasoningRequest(&input); err != nil {
        return nil, nil, err
    }

    // 2. Business logic
    belief, err := h.probabilisticReasoner.UpdateBeliefFull(
        input.BeliefID,
        input.EvidenceID,
        input.LikelihoodIfTrue,
        input.LikelihoodIfFalse,
    )
    if err != nil {
        return nil, nil, err
    }

    // 3. Response construction
    response := &ProbabilisticReasoningResponse{
        Belief:    belief,
        Operation: input.Operation,
        Status:    "success",
    }

    return &mcp.CallToolResult{
        Content: toJSONContent(response),
    }, response, nil
}

// Validation function
func ValidateProbabilisticReasoningRequest(req *ProbabilisticReasoningRequest) error {
    if req.LikelihoodIfTrue < 0 || req.LikelihoodIfTrue > 1 {
        return &ValidationError{"likelihood_if_true", "must be between 0 and 1"}
    }
    return nil
}
```

**Benefits**:
- Clear domain separation
- Testable business logic
- Type-safe request/response handling
- Reusable validation functions

### 3. Tool Registration (`internal/server/server.go`)

**Purpose**: Wire tool definitions to handler implementations

**Pattern**:
```go
func (s *UnifiedServer) RegisterTools(mcpServer *mcp.Server) error {
    // Get tool definitions
    for _, toolDef := range ToolDefinitions {
        switch toolDef.Name {
        case "probabilistic-reasoning":
            mcp.AddTool(mcpServer, &toolDef, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
                var input handlers.ProbabilisticReasoningRequest
                if err := json.Unmarshal([]byte(req.Params.Arguments), &input); err != nil {
                    return nil, fmt.Errorf("invalid input: %w", err)
                }

                result, _, err := s.probabilisticHandler.HandleProbabilisticReasoning(ctx, req, input)
                return result, err
            })

        case "make-decision":
            mcp.AddTool(mcpServer, &toolDef, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
                var input handlers.DecisionRequest
                if err := json.Unmarshal([]byte(req.Params.Arguments), &input); err != nil {
                    return nil, fmt.Errorf("invalid input: %w", err)
                }

                result, _, err := s.decisionHandler.HandleMakeDecision(ctx, req, input)
                return result, err
            })

        // ... other tools
        }
    }

    return nil
}
```

**Why This Pattern?**

The Go MCP SDK requires compile-time type information for `mcp.AddTool()`, which prevents dynamic registration. The switch statement provides:
- Type-safe unmarshaling of request arguments
- Direct handler invocation with correct types
- Clear mapping between tool names and handlers

**Benefits**:
- Centralized registration logic
- Easy to see all tool → handler mappings
- Type safety at registration time
- Single place to modify wiring

## Adding a New Tool

### Step 1: Define Tool Schema (`internal/server/tools.go`)

```go
var ToolDefinitions = []mcp.Tool{
    // ... existing tools
    {
        Name: "my-new-tool",
        Description: `Brief description of what the tool does.

**Parameters:**
- param1 (required): Description
- param2 (optional): Description

**Returns:**
- field1: Description
- field2: Description

**Example:**
{"param1": "value", "param2": "value"}`,
    },
}
```

### Step 2: Create Handler (`internal/server/handlers/my_domain.go`)

```go
package handlers

// Request/Response types
type MyToolRequest struct {
    Param1 string `json:"param1"`
    Param2 string `json:"param2,omitempty"`
}

type MyToolResponse struct {
    Field1 string `json:"field1"`
    Field2 string `json:"field2"`
    Status string `json:"status"`
}

// Handler struct
type MyDomainHandler struct {
    storage storage.Storage
    // ... other dependencies
}

func NewMyDomainHandler(store storage.Storage) *MyDomainHandler {
    return &MyDomainHandler{storage: store}
}

// Tool handler
func (h *MyDomainHandler) HandleMyTool(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input MyToolRequest,
) (*mcp.CallToolResult, *MyToolResponse, error) {
    // Validate
    if err := ValidateMyToolRequest(&input); err != nil {
        return nil, nil, err
    }

    // Business logic
    result, err := h.processRequest(input)
    if err != nil {
        return nil, nil, err
    }

    // Response
    response := &MyToolResponse{
        Field1: result.Field1,
        Field2: result.Field2,
        Status: "success",
    }

    return &mcp.CallToolResult{
        Content: toJSONContent(response),
    }, response, nil
}

// Validation
func ValidateMyToolRequest(req *MyToolRequest) error {
    if req.Param1 == "" {
        return &ValidationError{"param1", "param1 is required"}
    }
    return nil
}
```

### Step 3: Initialize Handler (`internal/server/server.go`)

```go
type UnifiedServer struct {
    // ... existing handlers
    myDomainHandler *handlers.MyDomainHandler
}

func NewUnifiedServer(...) *UnifiedServer {
    s := &UnifiedServer{
        // ... existing initialization
        myDomainHandler: handlers.NewMyDomainHandler(store),
    }
    return s
}
```

### Step 4: Register Tool (`internal/server/server.go`)

```go
func (s *UnifiedServer) RegisterTools(mcpServer *mcp.Server) error {
    for _, toolDef := range ToolDefinitions {
        switch toolDef.Name {
        // ... existing cases

        case "my-new-tool":
            mcp.AddTool(mcpServer, &toolDef, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
                var input handlers.MyToolRequest
                if err := json.Unmarshal([]byte(req.Params.Arguments), &input); err != nil {
                    return nil, fmt.Errorf("invalid input: %w", err)
                }

                result, _, err := s.myDomainHandler.HandleMyTool(ctx, req, input)
                return result, err
            })
        }
    }
    return nil
}
```

### Step 5: Write Tests (`internal/server/handlers/my_domain_test.go`)

```go
package handlers_test

func TestMyDomainHandler_HandleMyTool(t *testing.T) {
    // Setup
    store := storage.NewMemoryStorage()
    handler := handlers.NewMyDomainHandler(store)

    // Test valid request
    req := handlers.MyToolRequest{
        Param1: "test",
    }

    result, response, err := handler.HandleMyTool(context.Background(), nil, req)

    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "success", response.Status)
}
```

## Design Rationale

### Why Not Dynamic Registration?

**Attempted Approach** (Commit 5807e7b, reverted in e2a241d):
```go
// DOES NOT WORK - Go type system limitation
func RegisterToolDynamically(name string, handler interface{}) {
    mcp.AddTool(server, tool, handler) // Requires compile-time type info!
}
```

**Problem**: `mcp.AddTool()` signature:
```go
func AddTool[TArgs, TResult any](
    server *Server,
    tool *Tool,
    handler func(context.Context, TArgs) (*CallToolResult, error)
) error
```

The generic type parameters `TArgs` and `TResult` must be known at compile time. Go's reflection cannot provide this information dynamically.

### Why Not Interface-Based?

**Alternative Considered**:
```go
type ToolHandler interface {
    Handle(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
```

**Problem**: Loses type safety for request/response. Every handler would need runtime type assertions, increasing error risk.

### Current Pattern Benefits

✅ **Type Safety**: Compile-time verification of request/response types
✅ **Clear Wiring**: Easy to see tool → handler mappings
✅ **Maintainability**: Definitions separate from implementation
✅ **Testability**: Handlers are independently testable
✅ **Documentation**: Schema and docs in one place
✅ **Modularity**: Domain-specific handlers are isolated

## Best Practices

### 1. Tool Naming

- Use kebab-case: `my-tool-name`
- Be descriptive: `analyze-temporal` not `analyze`
- Group related tools: `make-decision`, `decompose-problem`

### 2. Request/Response Types

- Define explicit structs (no `map[string]interface{}`)
- Use JSON tags: `json:"field_name"`
- Mark optional fields: `omitempty`
- Validate all required fields

### 3. Error Handling

```go
// Good: Specific error with context
return nil, nil, &ValidationError{"param", "param must be between 0 and 1"}

// Bad: Generic error
return nil, nil, fmt.Errorf("invalid")
```

### 4. Handler Structure

```go
// Good: Clean separation
func (h *Handler) HandleTool(ctx, req, input) (result, response, error) {
    // 1. Validate
    // 2. Process
    // 3. Return
}

// Bad: Mixed concerns
func (h *Handler) HandleTool(ctx, req) (result, error) {
    // Parsing + validation + processing all mixed
}
```

### 5. Testing

- Test validation logic separately
- Test business logic with mock storage
- Test error cases explicitly
- Test response format compliance

## Tool Categories

The 63 tools are organized into logical groups:

1. **Core Thinking** (11): think, history, branches, validation
2. **Probabilistic** (4): Bayesian inference, evidence assessment
3. **Decision** (3): Decision-making, problem decomposition
4. **Metacognition** (3): Self-evaluation, bias detection
5. **Hallucination** (4): Verification, calibration tracking
6. **Temporal** (4): Time-horizon analysis, optimal timing
7. **Causal** (5): Causal graphs, interventions, counterfactuals
8. **Integration** (6): Synthesis, workflows, patterns
9. **Advanced** (10): Dual-process, backtracking, abductive, CBR
10. **Enhanced** (8): Analogies, arguments, evidence pipeline
11. **Episodic** (5): Memory, learning, recommendations

## Performance Considerations

### Handler Initialization

- Handlers are initialized once in `NewUnifiedServer()`
- No per-request allocation
- Shared dependencies (storage, reasoners)

### Registration Performance

- Tool definitions are pre-compiled
- Switch statement is O(n) but n=63 and only runs once
- No reflection overhead during request handling

### Request Handling

- Direct function calls (no reflection)
- Type-safe unmarshaling
- Minimal allocation overhead

## Future Improvements

### Code Generation

**Potential**: Generate registration code from tool definitions
```go
//go:generate go run tools/generate_registration.go
```

This would eliminate the manual switch statement while maintaining type safety.

### Handler Registry

**Pattern**: Auto-discovery of handlers via struct tags or build-time analysis
```go
type Handler struct {
    ToolName string `tool:"my-tool"`
    // ...
}
```

Could reduce boilerplate but requires reflection or code generation.

## Conclusion

The three-layer separation (definitions → handlers → registration) provides a maintainable, type-safe architecture for the 63-tool MCP server. While the switch statement in `RegisterTools()` is verbose, it ensures:

1. **Type Safety**: Compile-time verification of all tool interfaces
2. **Clarity**: Explicit mapping between tools and handlers
3. **Maintainability**: Clear separation of concerns
4. **Testability**: Independently testable components

This architecture successfully refactored a 2,225-line monolithic `server.go` into a modular, organized system with 701 fewer lines while maintaining all functionality and improving code quality.
