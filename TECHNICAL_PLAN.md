# Unified Thinking Server - Technical Implementation Plan (Go)

## Overview

This document provides a DETAILED technical implementation plan for consolidating 5 thinking servers into one unified Go-based MCP server.

**Target Directory**: `C:/Development/Projects/MCP/project-root/mcp-servers/unified-thinking/`

---

## Phase 0: Learning & Setup (CURRENT)

### Go MCP SDK Analysis

**Official SDK**: `github.com/modelcontextprotocol/go-sdk/mcp`

#### Key Findings:
1. ✅ Official Go SDK exists (maintained with Google)
2. ✅ Supports stdio transport (our primary need)
3. ✅ Generic type-safe tool definitions via `AddTool`
4. ✅ Similar patterns to TypeScript SDK

#### Go SDK Core APIs:
```go
// From pkg.go.dev documentation
import "github.com/modelcontextprotocol/go-sdk/mcp"

// Create server
server := mcp.NewServer(&mcp.Implementation{
    Name: "server-name",
    Version: "1.0.0",
}, nil)

// Add tool with automatic schema generation
server.AddTool("tool-name", mcp.ToolHandlerFor(...))

// Run server
transport := mcp.NewStdioTransport()
server.Serve(context.Background(), transport)
```

### Current TypeScript Servers Analysis

From examining your existing servers:

**sequential-thinking** (`sequential-thinking.js`):
- Simple state storage
- Basic tool: `solve-problem`
- Minimal functionality

**branch-thinking** (`branch-thinking/src/`):
- Complex: BranchManager class
- Multiple tool commands (list, focus, history)
- Rich data model: ThoughtBranch, Insights, CrossRefs
- Confidence scoring
- Branch state management

**unreasonable-thinking** (`Unconventional-thinking/src/index.ts`):
- Simple thought generation
- Tools: `generate_unreasonable_thought`, `branch_thought`, `list_thoughts`
- In-memory storage
- Branch counter

**mcp-logic** (`mcp-logic/build/index.js`):
- Zod schemas for validation
- Tools: `prove`, `check-well-formed`
- Logic processing

**state-coordinator** (`state-coordinator-enhanced/dist/index.js`):
- Sophisticated storage layer
- Tools: store-state, get-state, store-insight, validate-insight, link-states
- Relationship management

### Decision: Why Go?

**Advantages**:
1. ✅ Performance: Compiled, faster than Node.js
2. ✅ Type Safety: Strong typing without TypeScript overhead
3. ✅ Single Binary: Easy deployment
4. ✅ Concurrency: Better for parallel branch processing
5. ✅ Official SDK: Well-maintained by Anthropic + Google

---

## Phase 1: Project Structure Setup

### Directory Structure

```
unified-thinking/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── modes/
│   │   ├── linear.go            # Sequential thinking
│   │   ├── tree.go              # Branch thinking
│   │   ├── divergent.go         # Creative thinking
│   │   └── auto.go              # Mode detection
│   ├── storage/
│   │   ├── memory.go            # In-memory store
│   │   ├── models.go            # Data structures
│   │   └── persistence.go       # Optional: disk persistence
│   ├── validation/
│   │   └── logic.go             # Logical validation
│   ├── server/
│   │   ├── server.go            # Main server logic
│   │   ├── handlers.go          # Tool handlers
│   │   └── formatters.go        # Output formatting
│   └── types/
│       └── types.go             # Shared types
├── go.mod
├── go.sum
├── README.md
├── TECHNICAL_PLAN.md            # This file
└── examples/
    └── usage_examples.md
```

### Initial Files to Create

#### 1. `go.mod`
```go
module github.com/YOUR_USERNAME/unified-thinking-server

go 1.23

require (
    github.com/modelcontextprotocol/go-sdk v0.1.0-rc
)
```

#### 2. `internal/types/types.go` - Core Data Structures

```go
package types

import "time"

// ThinkingMode represents the type of thinking
type ThinkingMode string

const (
    ModeLinear    ThinkingMode = "linear"
    ModeTree      ThinkingMode = "tree"
    ModeDivergent ThinkingMode = "divergent"
    ModeAuto      ThinkingMode = "auto"
)

// ThoughtState represents the state of a thought or branch
type ThoughtState string

const (
    StateActive    ThoughtState = "active"
    StateSuspended ThoughtState = "suspended"
    StateCompleted ThoughtState = "completed"
    StateDeadEnd   ThoughtState = "dead_end"
)

// InsightType categorizes insights
type InsightType string

const (
    InsightBehavioralPattern  InsightType = "behavioral_pattern"
    InsightFeatureIntegration InsightType = "feature_integration"
    InsightObservation        InsightType = "observation"
    InsightConnection         InsightType = "connection"
)

// CrossRefType categorizes cross-references
type CrossRefType string

const (
    CrossRefComplementary CrossRefType = "complementary"
    CrossRefContradictory CrossRefType = "contradictory"
    CrossRefBuildsUpon    CrossRefType = "builds_upon"
    CrossRefAlternative   CrossRefType = "alternative"
)

// Thought represents a single thought in the system
type Thought struct {
    ID          string                 `json:"id"`
    Content     string                 `json:"content"`
    Mode        ThinkingMode           `json:"mode"`
    BranchID    string                 `json:"branch_id,omitempty"`
    ParentID    string                 `json:"parent_id,omitempty"`
    Type        string                 `json:"type"`
    Confidence  float64                `json:"confidence"`
    Timestamp   time.Time              `json:"timestamp"`
    KeyPoints   []string               `json:"key_points,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    
    // Flags
    IsRebellion         bool `json:"is_rebellion"`
    ChallengesAssumption bool `json:"challenges_assumption"`
}

// Branch represents a branch in tree-mode thinking
type Branch struct {
    ID             string         `json:"id"`
    ParentBranchID string         `json:"parent_branch_id,omitempty"`
    State          ThoughtState   `json:"state"`
    Priority       float64        `json:"priority"`
    Confidence     float64        `json:"confidence"`
    Thoughts       []*Thought     `json:"thoughts"`
    Insights       []*Insight     `json:"insights"`
    CrossRefs      []*CrossRef    `json:"cross_refs"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
}

// Insight represents a derived insight
type Insight struct {
    ID                  string                 `json:"id"`
    Type                InsightType            `json:"type"`
    Content             string                 `json:"content"`
    Context             []string               `json:"context"`
    ParentInsights      []string               `json:"parent_insights,omitempty"`
    ApplicabilityScore  float64                `json:"applicability_score"`
    SupportingEvidence  map[string]interface{} `json:"supporting_evidence"`
    Validations         []*Validation          `json:"validations,omitempty"`
    CreatedAt           time.Time              `json:"created_at"`
}

// CrossRef represents a cross-reference between branches
type CrossRef struct {
    ID           string       `json:"id"`
    FromBranch   string       `json:"from_branch"`
    ToBranch     string       `json:"to_branch"`
    Type         CrossRefType `json:"type"`
    Reason       string       `json:"reason"`
    Strength     float64      `json:"strength"`
    TouchPoints  []TouchPoint `json:"touchpoints,omitempty"`
    CreatedAt    time.Time    `json:"created_at"`
}

// TouchPoint represents a connection point between thoughts
type TouchPoint struct {
    FromThought string `json:"from_thought"`
    ToThought   string `json:"to_thought"`
    Connection  string `json:"connection"`
}

// Validation represents logical validation results
type Validation struct {
    ID             string                 `json:"id"`
    InsightID      string                 `json:"insight_id,omitempty"`
    ThoughtID      string                 `json:"thought_id,omitempty"`
    IsValid        bool                   `json:"is_valid"`
    ValidationData map[string]interface{} `json:"validation_data,omitempty"`
    Reason         string                 `json:"reason,omitempty"`
    CreatedAt      time.Time              `json:"created_at"`
}

// Relationship represents connections between states
type Relationship struct {
    ID          string                 `json:"id"`
    FromStateID string                 `json:"from_state_id"`
    ToStateID   string                 `json:"to_state_id"`
    Type        string                 `json:"type"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
}
```

#### 3. `internal/storage/memory.go` - Storage Layer

```go
package storage

import (
    "fmt"
    "sync"
    "time"
    
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/types"
)

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
    mu            sync.RWMutex
    thoughts      map[string]*types.Thought
    branches      map[string]*types.Branch
    insights      map[string]*types.Insight
    validations   map[string]*types.Validation
    relationships map[string]*types.Relationship
    
    activeBranchID string
    
    // Counters for ID generation
    thoughtCounter int
    branchCounter  int
    insightCounter int
    validationCounter int
    relationshipCounter int
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        thoughts:      make(map[string]*types.Thought),
        branches:      make(map[string]*types.Branch),
        insights:      make(map[string]*types.Insight),
        validations:   make(map[string]*types.Validation),
        relationships: make(map[string]*types.Relationship),
    }
}

// StoreThought stores a thought
func (s *MemoryStorage) StoreThought(thought *types.Thought) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if thought.ID == "" {
        s.thoughtCounter++
        thought.ID = fmt.Sprintf("thought-%d-%d", time.Now().Unix(), s.thoughtCounter)
    }
    
    s.thoughts[thought.ID] = thought
    return nil
}

// GetThought retrieves a thought by ID
func (s *MemoryStorage) GetThought(id string) (*types.Thought, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    thought, exists := s.thoughts[id]
    if !exists {
        return nil, fmt.Errorf("thought not found: %s", id)
    }
    return thought, nil
}

// StoreBranch stores a branch
func (s *MemoryStorage) StoreBranch(branch *types.Branch) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if branch.ID == "" {
        s.branchCounter++
        branch.ID = fmt.Sprintf("branch-%d-%d", time.Now().Unix(), s.branchCounter)
    }
    
    s.branches[branch.ID] = branch
    
    // Set as active if it's the first branch
    if s.activeBranchID == "" {
        s.activeBranchID = branch.ID
    }
    
    return nil
}

// GetBranch retrieves a branch by ID
func (s *MemoryStorage) GetBranch(id string) (*types.Branch, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    branch, exists := s.branches[id]
    if !exists {
        return nil, fmt.Errorf("branch not found: %s", id)
    }
    return branch, nil
}

// ListBranches returns all branches
func (s *MemoryStorage) ListBranches() []*types.Branch {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    branches := make([]*types.Branch, 0, len(s.branches))
    for _, branch := range s.branches {
        branches = append(branches, branch)
    }
    return branches
}

// GetActiveBranch returns the currently active branch
func (s *MemoryStorage) GetActiveBranch() (*types.Branch, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    if s.activeBranchID == "" {
        return nil, fmt.Errorf("no active branch")
    }
    
    return s.branches[s.activeBranchID], nil
}

// SetActiveBranch sets the active branch
func (s *MemoryStorage) SetActiveBranch(branchID string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.branches[branchID]; !exists {
        return fmt.Errorf("branch not found: %s", branchID)
    }
    
    s.activeBranchID = branchID
    return nil
}

// StoreInsight stores an insight
func (s *MemoryStorage) StoreInsight(insight *types.Insight) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if insight.ID == "" {
        s.insightCounter++
        insight.ID = fmt.Sprintf("insight-%d", s.insightCounter)
    }
    
    s.insights[insight.ID] = insight
    return nil
}

// StoreValidation stores a validation result
func (s *MemoryStorage) StoreValidation(validation *types.Validation) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if validation.ID == "" {
        s.validationCounter++
        validation.ID = fmt.Sprintf("validation-%d", s.validationCounter)
    }
    
    s.validations[validation.ID] = validation
    return nil
}

// StoreRelationship stores a relationship
func (s *MemoryStorage) StoreRelationship(rel *types.Relationship) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if rel.ID == "" {
        s.relationshipCounter++
        rel.ID = fmt.Sprintf("rel-%d", s.relationshipCounter)
    }
    
    s.relationships[rel.ID] = rel
    return nil
}

// SearchThoughts searches thoughts by content or type
func (s *MemoryStorage) SearchThoughts(query string, mode types.ThinkingMode) []*types.Thought {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    var results []*types.Thought
    for _, thought := range s.thoughts {
        // Simple search - in production, use proper text search
        if (query == "" || contains(thought.Content, query)) &&
           (mode == "" || thought.Mode == mode) {
            results = append(results, thought)
        }
    }
    return results
}

func contains(s, substr string) bool {
    // Simplified - in production use proper string search
    return true // Placeholder
}
```

---

## Phase 2: Implement Core Modes

### Linear Mode (`internal/modes/linear.go`)

```go
package modes

import (
    "context"
    "time"
    
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/storage"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/types"
)

// LinearMode implements sequential step-by-step reasoning
type LinearMode struct {
    storage *storage.MemoryStorage
}

// NewLinearMode creates a new linear mode handler
func NewLinearMode(storage *storage.MemoryStorage) *LinearMode {
    return &LinearMode{storage: storage}
}

// ProcessThought processes a thought in linear mode
func (m *LinearMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
    thought := &types.Thought{
        Content:    input.Content,
        Mode:       types.ModeLinear,
        Type:       input.Type,
        Confidence: input.Confidence,
        Timestamp:  time.Now(),
    }
    
    // Store the thought
    if err := m.storage.StoreThought(thought); err != nil {
        return nil, err
    }
    
    // Create simple state tracking
    result := &ThoughtResult{
        ThoughtID:  thought.ID,
        Mode:       string(types.ModeLinear),
        Status:     "processed",
        Confidence: thought.Confidence,
    }
    
    return result, nil
}

// GetHistory returns the linear history of thoughts
func (m *LinearMode) GetHistory(ctx context.Context) ([]*types.Thought, error) {
    thoughts := m.storage.SearchThoughts("", types.ModeLinear)
    return thoughts, nil
}
```

### Tree Mode (`internal/modes/tree.go`)

```go
package modes

import (
    "context"
    "fmt"
    "time"
    
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/storage"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/types"
)

// TreeMode implements multi-branch exploration
type TreeMode struct {
    storage *storage.MemoryStorage
}

// NewTreeMode creates a new tree mode handler
func NewTreeMode(storage *storage.MemoryStorage) *TreeMode {
    return &TreeMode{storage: storage}
}

// ProcessThought processes a thought in tree mode with branching
func (m *TreeMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
    // Determine branch (use provided or active)
    branchID := input.BranchID
    if branchID == "" {
        if activeBranch, err := m.storage.GetActiveBranch(); err == nil {
            branchID = activeBranch.ID
        } else {
            // Create new branch
            branch := &types.Branch{
                State:      types.StateActive,
                Priority:   1.0,
                Confidence: input.Confidence,
                Thoughts:   make([]*types.Thought, 0),
                Insights:   make([]*types.Insight, 0),
                CrossRefs:  make([]*types.CrossRef, 0),
                CreatedAt:  time.Now(),
                UpdatedAt:  time.Now(),
            }
            if err := m.storage.StoreBranch(branch); err != nil {
                return nil, err
            }
            branchID = branch.ID
        }
    }
    
    // Create thought
    thought := &types.Thought{
        Content:    input.Content,
        Mode:       types.ModeTree,
        BranchID:   branchID,
        ParentID:   input.ParentID,
        Type:       input.Type,
        Confidence: input.Confidence,
        KeyPoints:  input.KeyPoints,
        Timestamp:  time.Now(),
    }
    
    if err := m.storage.StoreThought(thought); err != nil {
        return nil, err
    }
    
    // Update branch
    branch, err := m.storage.GetBranch(branchID)
    if err != nil {
        return nil, err
    }
    branch.Thoughts = append(branch.Thoughts, thought)
    
    // Generate insights from key points
    if len(input.KeyPoints) > 0 {
        insight := &types.Insight{
            Type:               types.InsightObservation,
            Content:            fmt.Sprintf("Key points identified: %v", input.KeyPoints),
            Context:            []string{input.Type},
            ApplicabilityScore: input.Confidence,
            CreatedAt:          time.Now(),
        }
        if err := m.storage.StoreInsight(insight); err != nil {
            return nil, err
        }
        branch.Insights = append(branch.Insights, insight)
    }
    
    // Handle cross-references
    for _, xref := range input.CrossRefs {
        crossRef := &types.CrossRef{
            FromBranch: branchID,
            ToBranch:   xref.ToBranch,
            Type:       types.CrossRefType(xref.Type),
            Reason:     xref.Reason,
            Strength:   xref.Strength,
            CreatedAt:  time.Now(),
        }
        if err := m.storage.StoreInsight(insight); err != nil {
            return nil, err
        }
        branch.CrossRefs = append(branch.CrossRefs, crossRef)
    }
    
    // Update branch metrics
    m.updateBranchMetrics(branch)
    branch.UpdatedAt = time.Now()
    
    result := &ThoughtResult{
        ThoughtID:   thought.ID,
        BranchID:    branchID,
        Mode:        string(types.ModeTree),
        BranchState: string(branch.State),
        Priority:    branch.Priority,
        InsightCount: len(branch.Insights),
        CrossRefCount: len(branch.CrossRefs),
    }
    
    return result, nil
}

func (m *TreeMode) updateBranchMetrics(branch *types.Branch) {
    // Calculate average confidence
    if len(branch.Thoughts) > 0 {
        totalConf := 0.0
        for _, t := range branch.Thoughts {
            totalConf += t.Confidence
        }
        branch.Confidence = totalConf / float64(len(branch.Thoughts))
    }
    
    // Calculate priority
    insightScore := float64(len(branch.Insights)) * 0.1
    crossRefScore := 0.0
    for _, xref := range branch.CrossRefs {
        crossRefScore += xref.Strength * 0.1
    }
    branch.Priority = branch.Confidence + insightScore + crossRefScore
}

// ListBranches returns all branches
func (m *TreeMode) ListBranches(ctx context.Context) ([]*types.Branch, error) {
    branches := m.storage.ListBranches()
    return branches, nil
}

// GetBranchHistory returns detailed history of a branch
func (m *TreeMode) GetBranchHistory(ctx context.Context, branchID string) (*BranchHistory, error) {
    branch, err := m.storage.GetBranch(branchID)
    if err != nil {
        return nil, err
    }
    
    history := &BranchHistory{
        BranchID:   branch.ID,
        State:      string(branch.State),
        Priority:   branch.Priority,
        Confidence: branch.Confidence,
        Thoughts:   branch.Thoughts,
        Insights:   branch.Insights,
        CrossRefs:  branch.CrossRefs,
    }
    
    return history, nil
}

// SetActiveBranch changes the active branch
func (m *TreeMode) SetActiveBranch(ctx context.Context, branchID string) error {
    return m.storage.SetActiveBranch(branchID)
}
```

### Divergent Mode (`internal/modes/divergent.go`)

```go
package modes

import (
    "context"
    "fmt"
    "math/rand"
    "time"
    
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/storage"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/types"
)

// DivergentMode implements creative/rebellious ideation
type DivergentMode struct {
    storage *storage.MemoryStorage
}

// NewDivergentMode creates a new divergent mode handler
func NewDivergentMode(storage *storage.MemoryStorage) *DivergentMode {
    return &DivergentMode{storage: storage}
}

// ProcessThought generates creative/unconventional thoughts
func (m *DivergentMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
    // Generate creative thought
    creativeContent := m.generateCreativeThought(input.Content, input.ForceRebellion)
    
    thought := &types.Thought{
        Content:             creativeContent,
        Mode:                types.ModeDivergent,
        Type:                input.Type,
        Confidence:          input.Confidence,
        Timestamp:           time.Now(),
        IsRebellion:         input.ForceRebellion || rand.Float64() > 0.5,
        ChallengesAssumption: rand.Float64() > 0.3,
    }
    
    if input.PreviousThoughtID != "" {
        thought.ParentID = input.PreviousThoughtID
    }
    
    if err := m.storage.StoreThought(thought); err != nil {
        return nil, err
    }
    
    result := &ThoughtResult{
        ThoughtID:           thought.ID,
        Mode:                string(types.ModeDivergent),
        Content:             creativeContent,
        IsRebellion:         thought.IsRebellion,
        ChallengesAssumption: thought.ChallengesAssumption,
    }
    
    return result, nil
}

func (m *DivergentMode) generateCreativeThought(problem string, forceRebellion bool) string {
    approaches := []string{
        fmt.Sprintf("What if we completely eliminated the concept of %s?", problem),
        fmt.Sprintf("Imagine if %s operated in reverse - what opportunities would that create?", problem),
        fmt.Sprintf("If we had infinite resources and no physical limitations, how would we solve %s?", problem),
        fmt.Sprintf("What if we combined %s with its exact opposite?", problem),
        fmt.Sprintf("How would an alien civilization with completely different logic solve %s?", problem),
    }
    
    return approaches[rand.Intn(len(approaches))]
}

// BranchThought creates a new creative branch from an existing thought
func (m *DivergentMode) BranchThought(ctx context.Context, thoughtID string, direction string) (*ThoughtResult, error) {
    sourceThought, err := m.storage.GetThought(thoughtID)
    if err != nil {
        return nil, err
    }
    
    branchedContent := m.generateBranchedThought(sourceThought, direction)
    
    thought := &types.Thought{
        Content:             branchedContent,
        Mode:                types.ModeDivergent,
        ParentID:            thoughtID,
        Type:                "branched_" + direction,
        Confidence:          0.7,
        Timestamp:           time.Now(),
        IsRebellion:         direction == "opposite",
        ChallengesAssumption: true,
    }
    
    if err := m.storage.StoreThought(thought); err != nil {
        return nil, err
    }
    
    result := &ThoughtResult{
        ThoughtID: thought.ID,
        Mode:      string(types.ModeDivergent),
        Content:   branchedContent,
        Direction: direction,
    }
    
    return result, nil
}

func (m *DivergentMode) generateBranchedThought(source *types.Thought, direction string) string {
    switch direction {
    case "more_extreme":
        return fmt.Sprintf("Taking it further: %s AND multiply it by 1000x", source.Content)
    case "opposite":
        return fmt.Sprintf("Complete reversal: What if the exact opposite of \"%s\" is the answer?", source.Content)
    case "tangential":
        return fmt.Sprintf("Unexpected connection: %s but in a completely different context", source.Content)
    default:
        return fmt.Sprintf("Building on: %s in a new direction", source.Content)
    }
}
```

### Auto Mode (`internal/modes/auto.go`)

```go
package modes

import (
    "context"
    "strings"
    
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/types"
)

// AutoMode implements automatic mode detection
type AutoMode struct {
    linear    *LinearMode
    tree      *TreeMode
    divergent *DivergentMode
}

// NewAutoMode creates a new auto mode detector
func NewAutoMode(linear *LinearMode, tree *TreeMode, divergent *DivergentMode) *AutoMode {
    return &AutoMode{
        linear:    linear,
        tree:      tree,
        divergent: divergent,
    }
}

// ProcessThought automatically selects the best mode and processes
func (m *AutoMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
    mode := m.detectMode(input)
    
    switch mode {
    case types.ModeLinear:
        return m.linear.ProcessThought(ctx, input)
    case types.ModeTree:
        return m.tree.ProcessThought(ctx, input)
    case types.ModeDivergent:
        return m.divergent.ProcessThought(ctx, input)
    default:
        return m.linear.ProcessThought(ctx, input)
    }
}

func (m *AutoMode) detectMode(input ThoughtInput) types.ThinkingMode {
    content := strings.ToLower(input.Content)
    
    // Divergent indicators
    divergentKeywords := []string{"creative", "unconventional", "what if", "imagine", "challenge", "rebel"}
    for _, kw := range divergentKeywords {
        if strings.Contains(content, kw) {
            return types.ModeDivergent
        }
    }
    
    // Tree indicators
    if input.BranchID != "" || len(input.CrossRefs) > 0 || len(input.KeyPoints) > 0 {
        return types.ModeTree
    }
    
    treeKeywords := []string{"branch", "explore", "alternative", "parallel"}
    for _, kw := range treeKeywords {
        if strings.Contains(content, kw) {
            return types.ModeTree
        }
    }
    
    // Default to linear
    return types.ModeLinear
}
```

---

## Phase 3: Validation Layer

### Logic Validation (`internal/validation/logic.go`)

```go
package validation

import (
    "fmt"
    "strings"
    
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/types"
)

// LogicValidator implements logical validation
type LogicValidator struct {}

// NewLogicValidator creates a new logic validator
func NewLogicValidator() *LogicValidator {
    return &LogicValidator{}
}

// ValidateThought validates a thought for logical consistency
func (v *LogicValidator) ValidateThought(thought *types.Thought) (*types.Validation, error) {
    // Simplified validation - in production, use proper logic engine
    isValid := v.checkBasicLogic(thought.Content)
    
    validation := &types.Validation{
        ThoughtID: thought.ID,
        IsValid:   isValid,
        Reason:    v.getValidationReason(isValid),
    }
    
    return validation, nil
}

// Prove attempts to prove a conclusion from premises
func (v *LogicValidator) Prove(premises []string, conclusion string) *ProofResult {
    // Simplified proof - in production, use proper theorem prover
    result := &ProofResult{
        Premises:   premises,
        Conclusion: conclusion,
        IsProvable: v.simpleProof(premises, conclusion),
    }
    
    return result
}

// CheckWellFormed validates statement syntax
func (v *LogicValidator) CheckWellFormed(statements []string) []StatementCheck {
    checks := make([]StatementCheck, len(statements))
    for i, stmt := range statements {
        checks[i] = StatementCheck{
            Statement:   stmt,
            IsWellFormed: v.checkSyntax(stmt),
        }
    }
    return checks
}

func (v *LogicValidator) checkBasicLogic(content string) bool {
    // Simplified - check for basic contradictions
    lower := strings.ToLower(content)
    
    // Check for obvious contradictions
    if strings.Contains(lower, "always") && strings.Contains(lower, "never") {
        return false
    }
    
    return true
}

func (v *LogicValidator) simpleProof(premises []string, conclusion string) bool {
    // Simplified proof checking
    // In production, implement proper formal logic
    return true
}

func (v *LogicValidator) checkSyntax(statement string) bool {
    // Simplified syntax check
    return len(statement) > 0 && strings.TrimSpace(statement) != ""
}

func (v *LogicValidator) getValidationReason(isValid bool) string {
    if isValid {
        return "Thought passes basic logical consistency checks"
    }
    return "Potential logical inconsistency detected"
}

type ProofResult struct {
    Premises   []string
    Conclusion string
    IsProvable bool
}

type StatementCheck struct {
    Statement    string
    IsWellFormed bool
}
```

---

## Phase 4: Server Implementation

### Main Server (`cmd/server/main.go`)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/modes"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/server"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/storage"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/validation"
)

func main() {
    // Initialize storage
    store := storage.NewMemoryStorage()
    
    // Initialize modes
    linearMode := modes.NewLinearMode(store)
    treeMode := modes.NewTreeMode(store)
    divergentMode := modes.NewDivergentMode(store)
    autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
    
    // Initialize validator
    validator := validation.NewLogicValidator()
    
    // Create server
    srv := server.NewUnifiedServer(store, linearMode, treeMode, divergentMode, autoMode, validator)
    
    // Create MCP server
    mcpServer := mcp.NewServer(&mcp.Implementation{
        Name:    "unified-thinking-server",
        Version: "1.0.0",
    }, nil)
    
    // Register tools
    srv.RegisterTools(mcpServer)
    
    // Create stdio transport
    transport := mcp.NewStdioTransport()
    
    // Run server
    ctx := context.Background()
    if err := mcpServer.Serve(ctx, transport); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

### Server Logic (`internal/server/server.go`)

```go
package server

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/modes"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/storage"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/validation"
    "github.com/YOUR_USERNAME/unified-thinking-server/internal/types"
)

// UnifiedServer is the main server implementation
type UnifiedServer struct {
    storage   *storage.MemoryStorage
    linear    *modes.LinearMode
    tree      *modes.TreeMode
    divergent *modes.DivergentMode
    auto      *modes.AutoMode
    validator *validation.LogicValidator
}

// NewUnifiedServer creates a new unified thinking server
func NewUnifiedServer(
    store *storage.MemoryStorage,
    linear *modes.LinearMode,
    tree *modes.TreeMode,
    divergent *modes.DivergentMode,
    auto *modes.AutoMode,
    validator *validation.LogicValidator,
) *UnifiedServer {
    return &UnifiedServer{
        storage:   store,
        linear:    linear,
        tree:      tree,
        divergent: divergent,
        auto:      auto,
        validator: validator,
    }
}

// RegisterTools registers all tools with the MCP server
func (s *UnifiedServer) RegisterTools(mcpServer *mcp.Server) {
    // Main thinking tool
    mcpServer.AddTool("think", mcp.ToolHandlerFor(s.handleThink))
    
    // History tools
    mcpServer.AddTool("history", mcp.ToolHandlerFor(s.handleHistory))
    
    // Branch management
    mcpServer.AddTool("list-branches", mcp.ToolHandlerFor(s.handleListBranches))
    mcpServer.AddTool("focus-branch", mcp.ToolHandlerFor(s.handleFocusBranch))
    mcpServer.AddTool("branch-history", mcp.ToolHandlerFor(s.handleBranchHistory))
    
    // Validation
    mcpServer.AddTool("validate", mcp.ToolHandlerFor(s.handleValidate))
    mcpServer.AddTool("prove", mcp.ToolHandlerFor(s.handleProve))
    
    // Search
    mcpServer.AddTool("search", mcp.ToolHandlerFor(s.handleSearch))
}

// Tool input/output types

type ThinkRequest struct {
    Content             string            `json:"content"`
    Mode                string            `json:"mode"` // linear, tree, divergent, auto
    Type                string            `json:"type,omitempty"`
    BranchID            string            `json:"branch_id,omitempty"`
    ParentID            string            `json:"parent_id,omitempty"`
    Confidence          float64           `json:"confidence,omitempty"`
    KeyPoints           []string          `json:"key_points,omitempty"`
    RequireValidation   bool              `json:"require_validation,omitempty"`
    ChallengeAssumptions bool             `json:"challenge_assumptions,omitempty"`
    ForceRebellion      bool              `json:"force_rebellion,omitempty"`
    CrossRefs           []CrossRefInput   `json:"cross_refs,omitempty"`
}

type CrossRefInput struct {
    ToBranch string  `json:"to_branch"`
    Type     string  `json:"type"`
    Reason   string  `json:"reason"`
    Strength float64 `json:"strength"`
}

type ThinkResponse struct {
    ThoughtID    string  `json:"thought_id"`
    Mode         string  `json:"mode"`
    BranchID     string  `json:"branch_id,omitempty"`
    Status       string  `json:"status"`
    Priority     float64 `json:"priority,omitempty"`
    Confidence   float64 `json:"confidence"`
    InsightCount int     `json:"insight_count,omitempty"`
    IsValid      bool    `json:"is_valid,omitempty"`
}

// handleThink is the main thinking tool handler
func (s *UnifiedServer) handleThink(ctx context.Context, req ThinkRequest) (*ThinkResponse, error) {
    // Convert to internal format
    input := modes.ThoughtInput{
        Content:           req.Content,
        Type:              req.Type,
        BranchID:          req.BranchID,
        ParentID:          req.ParentID,
        Confidence:        req.Confidence,
        KeyPoints:         req.KeyPoints,
        ForceRebellion:    req.ForceRebellion,
        CrossRefs:         convertCrossRefs(req.CrossRefs),
    }
    
    if input.Confidence == 0 {
        input.Confidence = 0.8 // Default
    }
    
    // Select mode
    var result *modes.ThoughtResult
    var err error
    
    mode := types.ThinkingMode(req.Mode)
    if mode == "" || mode == types.ModeAuto {
        result, err = s.auto.ProcessThought(ctx, input)
    } else {
        switch mode {
        case types.ModeLinear:
            result, err = s.linear.ProcessThought(ctx, input)
        case types.ModeTree:
            result, err = s.tree.ProcessThought(ctx, input)
        case types.ModeDivergent:
            result, err = s.divergent.ProcessThought(ctx, input)
        default:
            return nil, fmt.Errorf("unknown mode: %s", mode)
        }
    }
    
    if err != nil {
        return nil, err
    }
    
    // Optional validation
    isValid := true
    if req.RequireValidation {
        thought, _ := s.storage.GetThought(result.ThoughtID)
        if thought != nil {
            validation, _ := s.validator.ValidateThought(thought)
            if validation != nil {
                isValid = validation.IsValid
            }
        }
    }
    
    response := &ThinkResponse{
        ThoughtID:    result.ThoughtID,
        Mode:         result.Mode,
        BranchID:     result.BranchID,
        Status:       "success",
        Priority:     result.Priority,
        Confidence:   result.Confidence,
        InsightCount: result.InsightCount,
        IsValid:      isValid,
    }
    
    return response, nil
}

type HistoryRequest struct {
    Mode     string `json:"mode,omitempty"`
    BranchID string `json:"branch_id,omitempty"`
}

type HistoryResponse struct {
    Thoughts []*types.Thought `json:"thoughts"`
}

func (s *UnifiedServer) handleHistory(ctx context.Context, req HistoryRequest) (*HistoryResponse, error) {
    var thoughts []*types.Thought
    
    if req.BranchID != "" {
        branch, err := s.storage.GetBranch(req.BranchID)
        if err != nil {
            return nil, err
        }
        thoughts = branch.Thoughts
    } else {
        mode := types.ThinkingMode(req.Mode)
        thoughts = s.storage.SearchThoughts("", mode)
    }
    
    return &HistoryResponse{Thoughts: thoughts}, nil
}

type ListBranchesResponse struct {
    Branches      []*types.Branch `json:"branches"`
    ActiveBranchID string         `json:"active_branch_id"`
}

func (s *UnifiedServer) handleListBranches(ctx context.Context) (*ListBranchesResponse, error) {
    branches := s.storage.ListBranches()
    
    activeBranch, _ := s.storage.GetActiveBranch()
    activeID := ""
    if activeBranch != nil {
        activeID = activeBranch.ID
    }
    
    return &ListBranchesResponse{
        Branches:       branches,
        ActiveBranchID: activeID,
    }, nil
}

type FocusBranchRequest struct {
    BranchID string `json:"branch_id"`
}

type FocusBranchResponse struct {
    Status         string `json:"status"`
    ActiveBranchID string `json:"active_branch_id"`
}

func (s *UnifiedServer) handleFocusBranch(ctx context.Context, req FocusBranchRequest) (*FocusBranchResponse, error) {
    if err := s.storage.SetActiveBranch(req.BranchID); err != nil {
        return nil, err
    }
    
    return &FocusBranchResponse{
        Status:         "success",
        ActiveBranchID: req.BranchID,
    }, nil
}

type BranchHistoryRequest struct {
    BranchID string `json:"branch_id"`
}

func (s *UnifiedServer) handleBranchHistory(ctx context.Context, req BranchHistoryRequest) (*modes.BranchHistory, error) {
    return s.tree.GetBranchHistory(ctx, req.BranchID)
}

type ValidateRequest struct {
    ThoughtID string `json:"thought_id"`
}

type ValidateResponse struct {
    IsValid bool   `json:"is_valid"`
    Reason  string `json:"reason"`
}

func (s *UnifiedServer) handleValidate(ctx context.Context, req ValidateRequest) (*ValidateResponse, error) {
    thought, err := s.storage.GetThought(req.ThoughtID)
    if err != nil {
        return nil, err
    }
    
    validation, err := s.validator.ValidateThought(thought)
    if err != nil {
        return nil, err
    }
    
    return &ValidateResponse{
        IsValid: validation.IsValid,
        Reason:  validation.Reason,
    }, nil
}

type ProveRequest struct {
    Premises   []string `json:"premises"`
    Conclusion string   `json:"conclusion"`
}

type ProveResponse struct {
    IsProvable bool     `json:"is_provable"`
    Premises   []string `json:"premises"`
    Conclusion string   `json:"conclusion"`
}

func (s *UnifiedServer) handleProve(ctx context.Context, req ProveRequest) (*ProveResponse, error) {
    result := s.validator.Prove(req.Premises, req.Conclusion)
    
    return &ProveResponse{
        IsProvable: result.IsProvable,
        Premises:   result.Premises,
        Conclusion: result.Conclusion,
    }, nil
}

type SearchRequest struct {
    Query string `json:"query"`
    Mode  string `json:"mode,omitempty"`
}

type SearchResponse struct {
    Thoughts []*types.Thought `json:"thoughts"`
}

func (s *UnifiedServer) handleSearch(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
    mode := types.ThinkingMode(req.Mode)
    thoughts := s.storage.SearchThoughts(req.Query, mode)
    
    return &SearchResponse{Thoughts: thoughts}, nil
}

func convertCrossRefs(input []CrossRefInput) []modes.CrossRefInput {
    result := make([]modes.CrossRefInput, len(input))
    for i, xref := range input {
        result[i] = modes.CrossRefInput{
            ToBranch: xref.ToBranch,
            Type:     xref.Type,
            Reason:   xref.Reason,
            Strength: xref.Strength,
        }
    }
    return result
}
```

---

## Phase 5: Testing & Documentation

### Build Configuration

#### `Makefile`
```makefile
.PHONY: build run test clean

build:
	go build -o bin/unified-thinking ./cmd/server

run:
	go run ./cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

install-deps:
	go mod tidy
	go mod download
```

### README.md

```markdown
# Unified Thinking Server

A comprehensive MCP server that provides multiple cognitive modes for AI-assisted thinking.

## Features

- **Linear Mode**: Sequential step-by-step reasoning
- **Tree Mode**: Multi-branch parallel exploration
- **Divergent Mode**: Creative/unconventional ideation
- **Auto Mode**: Automatic mode selection
- **Validation**: Integrated logical consistency checking

## Installation

```bash
git clone <repo>
cd unified-thinking
go mod download
make build
```

## Usage

### Configuration

Add to Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:/Development/Projects/MCP/project-root/mcp-servers/unified-thinking/bin/unified-thinking",
      "transport": "stdio",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

### Example Queries

```
# Auto mode (recommended)
"Analyze the problem using the best thinking approach"

# Explicit modes
"Think linearly about this step by step" -> Linear mode
"Explore multiple branches of this idea" -> Tree mode
"What's a creative solution to this?" -> Divergent mode
```

## Tools

- `think` - Main thinking tool (supports all modes)
- `history` - View thinking history
- `list-branches` - List all branches
- `focus-branch` - Switch active branch
- `branch-history` - View branch details
- `validate` - Validate logical consistency
- `prove` - Prove logical statements
- `search` - Search thoughts

## Architecture

See [TECHNICAL_PLAN.md](TECHNICAL_PLAN.md) for detailed architecture.
```

---

## Phase 6: Migration from Old Servers

### Migration Guide

Create `MIGRATION.md`:

```markdown
# Migration Guide

## Replacing Old Servers

### Before
```json
{
  "sequential-thinking": {...},
  "branch-thinking": {...},
  "unreasonable-thinking-server": {...},
  "mcp-logic": {...},
  "state-coordinator": {...}
}
```

### After
```json
{
  "unified-thinking": {
    "command": "path/to/unified-thinking",
    "transport": "stdio"
  }
}
```

### Tool Mapping

| Old Server | Old Tool | New Tool | Notes |
|------------|----------|----------|-------|
| sequential | solve-problem | think (mode: linear) | Use mode="linear" |
| branch | branch-thinking | think (mode: tree) | Use mode="tree" |
| unreasonable | generate_unreasonable_thought | think (mode: divergent, force_rebellion: true) | Use mode="divergent" |
| mcp-logic | prove | prove | Same tool |
| mcp-logic | check-well-formed | validate | Similar functionality |
| state-coordinator | store-state | (automatic) | Built-in storage |
```

---

## Implementation Checklist

### Phase 1: Setup ✅
- [ ] Create directory structure
- [ ] Initialize go.mod
- [ ] Define types
- [ ] Implement storage layer

### Phase 2: Modes
- [ ] Implement LinearMode
- [ ] Implement TreeMode
- [ ] Implement DivergentMode
- [ ] Implement AutoMode

### Phase 3: Validation
- [ ] Implement LogicValidator
- [ ] Add validation to storage

### Phase 4: Server
- [ ] Implement server handlers
- [ ] Register all tools
- [ ] Test stdio transport

### Phase 5: Testing
- [ ] Write unit tests
- [ ] Integration tests
- [ ] Manual testing with Claude Desktop

### Phase 6: Documentation
- [ ] Write README
- [ ] Write migration guide
- [ ] Add examples

### Phase 7: Deployment
- [ ] Build binary
- [ ] Test in Claude Desktop
- [ ] Update main config
- [ ] Remove old servers

---

## Next Steps

1. Review this plan
2. Start with Phase 1 (types + storage)
3. Implement one mode at a time
4. Test each phase before moving forward
5. Keep old servers until new one is fully tested

---

## Questions to Answer Before Starting

1. **Persistence**: Should we persist to disk or stay in-memory?
2. **Configuration**: Any runtime configuration needs?
3. **Logging**: What level of logging/debugging?
4. **Error Handling**: How verbose should error messages be?

---

*Last Updated: $(date)*
*Status: PLANNING COMPLETE - READY FOR IMPLEMENTATION*
