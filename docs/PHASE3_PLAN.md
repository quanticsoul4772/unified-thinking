# Phase 3: Advanced Features Implementation Plan

## Overview

Phase 3 introduces two advanced capabilities that extend unified-thinking's AI integration:
1. **Feature 5: Multimodal Embeddings** - Support for image and document embeddings
2. **Feature 6: Programmatic Tool Calling** - Agentic LLM with tool use capabilities

---

## Feature 5: Multimodal Embeddings

### Problem Statement

Current embedding system only supports text input via Voyage AI's text embedding models. Users working with visual content (diagrams, screenshots, charts) or mixed documents cannot leverage semantic search.

### Solution Architecture

Extend the `Embedder` interface and Voyage AI client to support multimodal inputs using Voyage's multimodal models.

### API Analysis: Voyage Multimodal

Voyage AI offers `voyage-multimodal-3` which embeds:
- Text (up to 32K tokens)
- Images (base64 encoded or URLs)
- PDFs (first N pages extracted)

**API Endpoint**: `https://api.voyageai.com/v1/multimodalembeddings`

**Request Format**:
```json
{
  "model": "voyage-multimodal-3",
  "inputs": [
    {
      "content": [
        {"type": "text", "text": "Description of image"},
        {"type": "image_base64", "image_base64": "iVBORw0KGgo..."}
      ]
    }
  ]
}
```

### Implementation Plan

#### 5.1 Extend Embedder Interface

**File**: `internal/embeddings/embedder.go`

```go
// MultimodalInput represents content that can be text, image, or document
type MultimodalInput struct {
    Type     string `json:"type"`      // "text", "image_base64", "image_url", "document"
    Text     string `json:"text,omitempty"`
    ImageB64 string `json:"image_base64,omitempty"`
    ImageURL string `json:"image_url,omitempty"`
    Document []byte `json:"document,omitempty"` // PDF bytes
}

// MultimodalEmbedder extends Embedder with multimodal support
type MultimodalEmbedder interface {
    Embedder

    // EmbedMultimodal generates embedding for multimodal content
    EmbedMultimodal(ctx context.Context, inputs []MultimodalInput) ([]float32, error)

    // EmbedImage generates embedding for a single image
    EmbedImage(ctx context.Context, imageBase64 string) ([]float32, error)

    // EmbedImageWithText generates embedding for image with text description
    EmbedImageWithText(ctx context.Context, imageBase64, text string) ([]float32, error)

    // SupportsMultimodal returns whether the embedder supports multimodal input
    SupportsMultimodal() bool
}
```

#### 5.2 Create VoyageMultimodalEmbedder

**File**: `internal/embeddings/voyage_multimodal.go` (NEW)

```go
package embeddings

const (
    voyageMultimodalAPIURL = "https://api.voyageai.com/v1/multimodalembeddings"
    defaultMultimodalModel = "voyage-multimodal-3"
)

// VoyageMultimodalEmbedder implements MultimodalEmbedder
type VoyageMultimodalEmbedder struct {
    *VoyageEmbedder // Embed base for text operations
    multimodalModel string
}

// NewVoyageMultimodalEmbedder creates a multimodal embedder
func NewVoyageMultimodalEmbedder(apiKey, textModel, multimodalModel string) *VoyageMultimodalEmbedder {
    base := NewVoyageEmbedder(apiKey, textModel)
    return &VoyageMultimodalEmbedder{
        VoyageEmbedder:  base,
        multimodalModel: multimodalModel,
    }
}

// EmbedMultimodal generates embedding for mixed content
func (e *VoyageMultimodalEmbedder) EmbedMultimodal(ctx context.Context, inputs []MultimodalInput) ([]float32, error) {
    // Rate limiting
    if err := e.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }

    // Build request
    contentItems := make([]map[string]interface{}, len(inputs))
    for i, input := range inputs {
        contentItems[i] = input.toAPIFormat()
    }

    // Call API...
}

// EmbedImage embeds a single image
func (e *VoyageMultimodalEmbedder) EmbedImage(ctx context.Context, imageBase64 string) ([]float32, error) {
    return e.EmbedMultimodal(ctx, []MultimodalInput{
        {Type: "image_base64", ImageB64: imageBase64},
    })
}

// EmbedImageWithText embeds image with accompanying text
func (e *VoyageMultimodalEmbedder) EmbedImageWithText(ctx context.Context, imageBase64, text string) ([]float32, error) {
    return e.EmbedMultimodal(ctx, []MultimodalInput{
        {Type: "text", Text: text},
        {Type: "image_base64", ImageB64: imageBase64},
    })
}

func (e *VoyageMultimodalEmbedder) SupportsMultimodal() bool {
    return true
}
```

#### 5.3 Add Image Processing Utilities

**File**: `internal/embeddings/image_utils.go` (NEW)

```go
package embeddings

import (
    "encoding/base64"
    "fmt"
    "image"
    _ "image/jpeg"
    _ "image/png"
    "io"
    "net/http"
    "os"
)

// ImageLoader handles loading and preprocessing images
type ImageLoader struct {
    maxWidth  int
    maxHeight int
    maxBytes  int64
}

// NewImageLoader creates an image loader with size constraints
func NewImageLoader() *ImageLoader {
    return &ImageLoader{
        maxWidth:  4096,  // Voyage multimodal limit
        maxHeight: 4096,
        maxBytes:  20 * 1024 * 1024, // 20MB
    }
}

// LoadFromPath loads and encodes an image from file path
func (l *ImageLoader) LoadFromPath(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(data), nil
}

// LoadFromURL fetches and encodes an image from URL
func (l *ImageLoader) LoadFromURL(ctx context.Context, url string) (string, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    data, err := io.ReadAll(io.LimitReader(resp.Body, l.maxBytes))
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(data), nil
}

// ValidateImage checks image dimensions and format
func (l *ImageLoader) ValidateImage(data []byte) error {
    // Decode to check dimensions
    cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
    if err != nil {
        return fmt.Errorf("invalid image: %w", err)
    }

    if cfg.Width > l.maxWidth || cfg.Height > l.maxHeight {
        return fmt.Errorf("image too large: %dx%d (max %dx%d)",
            cfg.Width, cfg.Height, l.maxWidth, l.maxHeight)
    }
    return nil
}
```

#### 5.4 New MCP Tool: `embed-multimodal`

**File**: `internal/server/handlers/multimodal.go` (NEW)

```go
// EmbedMultimodalRequest for the embed-multimodal tool
type EmbedMultimodalRequest struct {
    Text      string `json:"text,omitempty"`
    ImagePath string `json:"image_path,omitempty"`
    ImageURL  string `json:"image_url,omitempty"`
    ImageB64  string `json:"image_base64,omitempty"`
}

// HandleEmbedMultimodal generates multimodal embeddings
func (h *MultimodalHandler) HandleEmbedMultimodal(ctx context.Context, req *EmbedMultimodalRequest) (*EmbedMultimodalResponse, error) {
    // Build multimodal inputs from request
    // Call embedder
    // Return embedding vector
}
```

#### 5.5 Extend Similarity Search

**File**: `internal/similarity/thought_search.go` (MODIFY)

```go
// SearchSimilarMultimodal finds thoughts similar to multimodal query
func (ts *ThoughtSearcher) SearchSimilarMultimodal(
    ctx context.Context,
    textQuery string,
    imageBase64 string,
    limit int,
    minSimilarity float32,
) ([]*SimilarThought, error) {
    // Check if embedder supports multimodal
    mmEmbedder, ok := ts.embedder.(embeddings.MultimodalEmbedder)
    if !ok {
        return nil, fmt.Errorf("embedder does not support multimodal")
    }

    // Generate query embedding
    queryEmbedding, err := mmEmbedder.EmbedImageWithText(ctx, imageBase64, textQuery)
    // ... rest of search logic
}
```

#### 5.6 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MULTIMODAL_ENABLED` | `false` | Enable multimodal embeddings |
| `MULTIMODAL_MODEL` | `voyage-multimodal-3` | Voyage multimodal model |
| `IMAGE_MAX_SIZE` | `20971520` | Max image size in bytes (20MB) |

#### 5.7 Test Plan

**File**: `internal/embeddings/voyage_multimodal_test.go` (NEW)

```go
func TestVoyageMultimodalEmbedder_EmbedImage(t *testing.T) { ... }
func TestVoyageMultimodalEmbedder_EmbedImageWithText(t *testing.T) { ... }
func TestVoyageMultimodalEmbedder_EmbedMultimodal(t *testing.T) { ... }
func TestImageLoader_LoadFromPath(t *testing.T) { ... }
func TestImageLoader_LoadFromURL(t *testing.T) { ... }
func TestImageLoader_ValidateImage(t *testing.T) { ... }
// Integration test (requires VOYAGE_API_KEY)
func TestVoyageMultimodalEmbedder_Integration(t *testing.T) { ... }
```

---

## Feature 6: Programmatic Tool Calling

### Problem Statement

Current LLM client only supports single-turn responses. Claude can reason about problems but cannot take actions (call tools, execute workflows) based on that reasoning. Users must manually invoke tools based on LLM suggestions.

### Solution Architecture

Implement an agentic LLM wrapper that:
1. Exposes unified-thinking tools as Claude tools
2. Handles tool call responses in a loop
3. Limits iterations to prevent runaway execution
4. Returns structured results with execution traces

### API Analysis: Anthropic Tool Use

Claude supports tool use via:
```json
{
  "model": "claude-sonnet-4-5-20250929",
  "max_tokens": 4096,
  "tools": [
    {
      "name": "search_similar_thoughts",
      "description": "Find similar thoughts via semantic search",
      "input_schema": { ... }
    }
  ],
  "messages": [...]
}
```

Response may include `tool_use` blocks:
```json
{
  "content": [
    {"type": "text", "text": "Let me search for similar thoughts..."},
    {"type": "tool_use", "id": "call_123", "name": "search_similar_thoughts", "input": {...}}
  ]
}
```

### Implementation Plan

#### 6.1 Define Tool Registry

**File**: `internal/modes/tool_registry.go` (NEW)

```go
package modes

import "unified-thinking/internal/server"

// ToolSpec defines a tool's interface for Claude
type ToolSpec struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"input_schema"`
    Handler     ToolHandler            // Internal handler reference
}

// ToolHandler executes a tool call
type ToolHandler func(ctx context.Context, input map[string]interface{}) (interface{}, error)

// ToolRegistry manages available tools
type ToolRegistry struct {
    tools   map[string]*ToolSpec
    server  *server.UnifiedServer
}

// NewToolRegistry creates registry from unified server
func NewToolRegistry(srv *server.UnifiedServer) *ToolRegistry {
    reg := &ToolRegistry{
        tools:  make(map[string]*ToolSpec),
        server: srv,
    }
    reg.registerBuiltinTools()
    return reg
}

// registerBuiltinTools registers unified-thinking tools
func (r *ToolRegistry) registerBuiltinTools() {
    // Register subset of tools suitable for agentic use
    r.register(ToolSpec{
        Name: "think",
        Description: "Process a thought with structured reasoning",
        InputSchema: thinkSchema,
        Handler: r.handleThink,
    })

    r.register(ToolSpec{
        Name: "search_similar_thoughts",
        Description: "Find semantically similar thoughts",
        InputSchema: searchSimilarSchema,
        Handler: r.handleSearchSimilar,
    })

    r.register(ToolSpec{
        Name: "build_causal_graph",
        Description: "Build causal graph from observations",
        InputSchema: causalGraphSchema,
        Handler: r.handleBuildCausalGraph,
    })

    // ... register other safe tools
}

// GetToolsForClaude returns tools in Claude API format
func (r *ToolRegistry) GetToolsForClaude() []map[string]interface{} {
    tools := make([]map[string]interface{}, 0, len(r.tools))
    for _, spec := range r.tools {
        tools = append(tools, map[string]interface{}{
            "name":         spec.Name,
            "description":  spec.Description,
            "input_schema": spec.InputSchema,
        })
    }
    return tools
}

// Execute runs a tool by name
func (r *ToolRegistry) Execute(ctx context.Context, name string, input map[string]interface{}) (interface{}, error) {
    spec, ok := r.tools[name]
    if !ok {
        return nil, fmt.Errorf("unknown tool: %s", name)
    }
    return spec.Handler(ctx, input)
}
```

#### 6.2 Create AgenticClient

**File**: `internal/modes/llm_agentic.go` (NEW)

```go
package modes

import (
    "context"
    "fmt"
)

// AgenticConfig configures the agentic client
type AgenticConfig struct {
    MaxIterations    int     // Maximum tool-calling iterations (default: 10)
    MaxToolsPerTurn  int     // Max tools per single response (default: 5)
    StopOnError      bool    // Stop if tool execution fails
    Temperature      float64 // Model temperature
    Model            string  // Model ID
}

// DefaultAgenticConfig returns sensible defaults
func DefaultAgenticConfig() AgenticConfig {
    return AgenticConfig{
        MaxIterations:   10,
        MaxToolsPerTurn: 5,
        StopOnError:     true,
        Temperature:     0.3,
        Model:           "claude-sonnet-4-5-20250929",
    }
}

// AgenticClient wraps AnthropicClient with tool-calling loop
type AgenticClient struct {
    client   *AnthropicClient
    registry *ToolRegistry
    config   AgenticConfig
}

// NewAgenticClient creates an agentic client
func NewAgenticClient(apiKey string, registry *ToolRegistry, config AgenticConfig) *AgenticClient {
    return &AgenticClient{
        client:   NewAnthropicClient(apiKey),
        registry: registry,
        config:   config,
    }
}

// ExecutionTrace records the agentic execution history
type ExecutionTrace struct {
    Iterations  []Iteration `json:"iterations"`
    TotalTokens int         `json:"total_tokens"`
    Duration    string      `json:"duration"`
}

// Iteration represents one round of thought + tool calls
type Iteration struct {
    Index      int           `json:"index"`
    Thought    string        `json:"thought"`
    ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
    ToolErrors []string      `json:"tool_errors,omitempty"`
}

// ToolCall represents a single tool invocation
type ToolCall struct {
    ID     string      `json:"id"`
    Name   string      `json:"name"`
    Input  interface{} `json:"input"`
    Output interface{} `json:"output"`
    Error  string      `json:"error,omitempty"`
}

// AgenticResult is the final result of agentic execution
type AgenticResult struct {
    FinalAnswer string         `json:"final_answer"`
    Trace       ExecutionTrace `json:"trace"`
    Status      string         `json:"status"` // "completed", "max_iterations", "error"
}

// Run executes an agentic task with tool use
func (a *AgenticClient) Run(ctx context.Context, task string) (*AgenticResult, error) {
    trace := ExecutionTrace{}
    messages := []map[string]interface{}{
        {"role": "user", "content": task},
    }

    tools := a.registry.GetToolsForClaude()

    for i := 0; i < a.config.MaxIterations; i++ {
        // Call Claude with tools
        resp, err := a.client.callWithTools(ctx, messages, tools, a.config)
        if err != nil {
            return nil, fmt.Errorf("iteration %d: %w", i, err)
        }

        iteration := Iteration{Index: i}

        // Process response content
        var toolCalls []ToolUseBlock
        for _, block := range resp.Content {
            switch block.Type {
            case "text":
                iteration.Thought = block.Text
            case "tool_use":
                toolCalls = append(toolCalls, block.ToolUse)
            }
        }

        // If no tool calls, we're done
        if len(toolCalls) == 0 {
            trace.Iterations = append(trace.Iterations, iteration)
            return &AgenticResult{
                FinalAnswer: iteration.Thought,
                Trace:       trace,
                Status:      "completed",
            }, nil
        }

        // Execute tool calls
        var toolResults []map[string]interface{}
        for _, tc := range toolCalls {
            call := ToolCall{
                ID:    tc.ID,
                Name:  tc.Name,
                Input: tc.Input,
            }

            result, err := a.registry.Execute(ctx, tc.Name, tc.Input)
            if err != nil {
                call.Error = err.Error()
                iteration.ToolErrors = append(iteration.ToolErrors, err.Error())

                if a.config.StopOnError {
                    trace.Iterations = append(trace.Iterations, iteration)
                    return &AgenticResult{
                        FinalAnswer: fmt.Sprintf("Tool error: %v", err),
                        Trace:       trace,
                        Status:      "error",
                    }, nil
                }

                toolResults = append(toolResults, map[string]interface{}{
                    "type":        "tool_result",
                    "tool_use_id": tc.ID,
                    "is_error":    true,
                    "content":     err.Error(),
                })
            } else {
                call.Output = result
                toolResults = append(toolResults, map[string]interface{}{
                    "type":        "tool_result",
                    "tool_use_id": tc.ID,
                    "content":     toJSON(result),
                })
            }

            iteration.ToolCalls = append(iteration.ToolCalls, call)
        }

        trace.Iterations = append(trace.Iterations, iteration)

        // Add assistant message and tool results
        messages = append(messages, map[string]interface{}{
            "role":    "assistant",
            "content": resp.Content,
        })
        messages = append(messages, map[string]interface{}{
            "role":    "user",
            "content": toolResults,
        })
    }

    return &AgenticResult{
        FinalAnswer: "Reached maximum iterations",
        Trace:       trace,
        Status:      "max_iterations",
    }, nil
}
```

#### 6.3 Extend AnthropicClient

**File**: `internal/modes/llm_anthropic.go` (MODIFY)

```go
// Add method for tool-calling requests
func (c *AnthropicClient) callWithTools(
    ctx context.Context,
    messages []map[string]interface{},
    tools []map[string]interface{},
    config AgenticConfig,
) (*MessageResponse, error) {
    reqBody := map[string]interface{}{
        "model":       config.Model,
        "max_tokens":  4096,
        "messages":    messages,
        "tools":       tools,
        "temperature": config.Temperature,
    }

    // HTTP request to Anthropic API...
}

// ContentBlock represents a response content block
type ContentBlock struct {
    Type    string       `json:"type"`
    Text    string       `json:"text,omitempty"`
    ToolUse ToolUseBlock `json:"tool_use,omitempty"`
}

// ToolUseBlock represents a tool call from Claude
type ToolUseBlock struct {
    ID    string                 `json:"id"`
    Name  string                 `json:"name"`
    Input map[string]interface{} `json:"input"`
}

// MessageResponse represents Claude's response
type MessageResponse struct {
    Content    []ContentBlock `json:"content"`
    StopReason string         `json:"stop_reason"`
    Usage      UsageInfo      `json:"usage"`
}
```

#### 6.4 New MCP Tool: `run-agent`

**File**: `internal/server/handlers/agent.go` (NEW)

```go
// RunAgentRequest for the run-agent tool
type RunAgentRequest struct {
    Task          string   `json:"task"`
    MaxIterations int      `json:"max_iterations,omitempty"` // Default: 10
    AllowedTools  []string `json:"allowed_tools,omitempty"`  // Restrict to specific tools
    StopOnError   bool     `json:"stop_on_error,omitempty"`  // Default: true
}

// RunAgentResponse contains agentic execution results
type RunAgentResponse struct {
    FinalAnswer  string                 `json:"final_answer"`
    Status       string                 `json:"status"`
    Iterations   int                    `json:"iterations"`
    ToolsUsed    []string               `json:"tools_used"`
    Trace        modes.ExecutionTrace   `json:"trace,omitempty"`
}

// HandleRunAgent executes an agentic task
func (h *AgentHandler) HandleRunAgent(ctx context.Context, req *RunAgentRequest) (*RunAgentResponse, error) {
    config := modes.DefaultAgenticConfig()
    if req.MaxIterations > 0 {
        config.MaxIterations = req.MaxIterations
    }
    config.StopOnError = req.StopOnError

    // Create registry with allowed tools
    registry := h.createRegistry(req.AllowedTools)

    // Create agentic client
    agent := modes.NewAgenticClient(h.apiKey, registry, config)

    // Execute
    result, err := agent.Run(ctx, req.Task)
    if err != nil {
        return nil, err
    }

    // Build response
    toolsUsed := make(map[string]bool)
    for _, iter := range result.Trace.Iterations {
        for _, tc := range iter.ToolCalls {
            toolsUsed[tc.Name] = true
        }
    }

    return &RunAgentResponse{
        FinalAnswer: result.FinalAnswer,
        Status:      result.Status,
        Iterations:  len(result.Trace.Iterations),
        ToolsUsed:   mapKeys(toolsUsed),
        Trace:       result.Trace,
    }, nil
}
```

#### 6.5 Safe Tool Subset

Not all unified-thinking tools should be available to the agent. Define a safe subset:

**Safe Tools (for agentic use)**:
- `think` - Process reasoning
- `search_similar_thoughts` - Semantic search
- `search_knowledge_graph` - Knowledge retrieval
- `build_causal_graph` - Causal analysis
- `generate_hypotheses` - Abductive reasoning
- `evaluate_hypotheses` - Hypothesis testing
- `analyze_perspectives` - Multi-perspective analysis
- `detect_biases` - Bias detection
- `detect_fallacies` - Fallacy detection
- `decompose_problem` - Problem decomposition

**Excluded Tools (side effects or resource-intensive)**:
- `store_entity` - Writes to knowledge graph
- `create_relationship` - Writes to knowledge graph
- `export_session` / `import_session` - Session management
- `run_preset` - Recursive agent potential
- `execute_workflow` - External orchestration

#### 6.6 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENT_ENABLED` | `false` | Enable agentic tool calling |
| `AGENT_MAX_ITERATIONS` | `10` | Maximum agent iterations |
| `AGENT_MODEL` | `claude-sonnet-4-5-20250929` | Model for agentic tasks |
| `AGENT_ALLOWED_TOOLS` | `` | Comma-separated tool allowlist (empty = all safe) |

#### 6.7 Test Plan

**File**: `internal/modes/llm_agentic_test.go` (NEW)

```go
func TestToolRegistry_RegisterBuiltinTools(t *testing.T) { ... }
func TestToolRegistry_Execute(t *testing.T) { ... }
func TestAgenticClient_Run_SingleIteration(t *testing.T) { ... }
func TestAgenticClient_Run_MultipleIterations(t *testing.T) { ... }
func TestAgenticClient_Run_MaxIterations(t *testing.T) { ... }
func TestAgenticClient_Run_ToolError(t *testing.T) { ... }
// Integration test (requires ANTHROPIC_API_KEY)
func TestAgenticClient_Integration(t *testing.T) { ... }
```

---

## File Summary

### Feature 5: Multimodal Embeddings

| Action | File |
|--------|------|
| CREATE | `internal/embeddings/voyage_multimodal.go` |
| CREATE | `internal/embeddings/voyage_multimodal_test.go` |
| CREATE | `internal/embeddings/image_utils.go` |
| CREATE | `internal/embeddings/image_utils_test.go` |
| CREATE | `internal/server/handlers/multimodal.go` |
| MODIFY | `internal/embeddings/embedder.go` - Add MultimodalEmbedder interface |
| MODIFY | `internal/similarity/thought_search.go` - Add SearchSimilarMultimodal |
| MODIFY | `cmd/server/initializer.go` - Initialize multimodal embedder |
| MODIFY | `internal/server/server.go` - Register embed-multimodal tool |

### Feature 6: Programmatic Tool Calling

| Action | File |
|--------|------|
| CREATE | `internal/modes/tool_registry.go` |
| CREATE | `internal/modes/tool_registry_test.go` |
| CREATE | `internal/modes/llm_agentic.go` |
| CREATE | `internal/modes/llm_agentic_test.go` |
| CREATE | `internal/server/handlers/agent.go` |
| CREATE | `internal/server/handlers/agent_test.go` |
| MODIFY | `internal/modes/llm_anthropic.go` - Add callWithTools |
| MODIFY | `cmd/server/initializer.go` - Initialize agentic client |
| MODIFY | `internal/server/server.go` - Register run-agent tool |

---

## Implementation Order

### Phase 3.1: Multimodal Embeddings (Recommended First)
1. Extend Embedder interface with MultimodalEmbedder
2. Implement VoyageMultimodalEmbedder
3. Create image utilities
4. Add embed-multimodal handler
5. Extend similarity search
6. Write tests
7. Update documentation

### Phase 3.2: Programmatic Tool Calling
1. Create ToolRegistry with safe tool subset
2. Implement AgenticClient with loop logic
3. Extend AnthropicClient for tool calling
4. Add run-agent handler
5. Write tests
6. Update documentation

---

## Risk Assessment

### Feature 5: Multimodal

| Risk | Severity | Mitigation |
|------|----------|------------|
| Large image uploads | Medium | Size validation, streaming |
| API costs (images are expensive) | Medium | Document costs, add warnings |
| Dimension mismatch with text | Low | Use same model family |

### Feature 6: Agentic Tool Calling

| Risk | Severity | Mitigation |
|------|----------|------------|
| Runaway execution | High | Max iterations, timeout |
| Tool abuse | Medium | Safe tool subset, rate limiting |
| High API costs | High | Iteration limits, monitoring |
| Recursive agent calls | High | Exclude run-agent from registry |
| Infinite loops | Medium | Cycle detection in trace |

---

## Success Criteria

### Feature 5
- [ ] EmbedImage works with JPEG/PNG
- [ ] EmbedImageWithText combines modalities
- [ ] Similarity search works across text and images
- [ ] Image size/format validation
- [ ] Tests pass with >80% coverage

### Feature 6
- [ ] Agent completes simple multi-step tasks
- [ ] Tool calls execute correctly
- [ ] Max iterations respected
- [ ] Errors handled gracefully
- [ ] Execution trace captures full history
- [ ] Tests pass with >80% coverage

---

## Documentation Updates

- CLAUDE.md: Add Multimodal Embeddings and Agentic Tool Calling sections
- CHANGELOG.md: Add Phase 3 entry
- README.md: Update tool count and capabilities
- docs/MULTIMODAL.md: Detailed multimodal guide (NEW)
- docs/AGENTIC.md: Detailed agentic guide (NEW)
