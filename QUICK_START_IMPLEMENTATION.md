# Quick Start Implementation Guide
## Unified Thinking MCP Server Improvements

**Estimated Time**: 30 hours (1 week)
**Difficulty**: Medium
**Prerequisites**: Go 1.21+, MCP SDK knowledge

---

## Day 1: Bug Fixes (3 hours)

### Task 1: Fix focus-branch Error (2 hours)

**Step 1.1**: Update `internal/storage/memory.go` (line 239-252)

```go
// SetActiveBranch sets the active branch and updates access tracking
func (s *MemoryStorage) SetActiveBranch(branchID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	branch, exists := s.branches[branchID]
	if !exists {
		// Provide helpful error with available branches
		availableBranches := make([]string, 0, len(s.branches))
		for id := range s.branches {
			availableBranches = append(availableBranches, id)
		}

		if len(availableBranches) == 0 {
			return fmt.Errorf("branch not found: %s (no branches exist yet, create one first)", branchID)
		}

		return fmt.Errorf("branch not found: %s (available branches: %v)", branchID, availableBranches)
	}

	s.activeBranchID = branchID
	branch.LastAccessedAt = time.Now()
	s.trackRecentBranch(branchID)
	return nil
}
```

**Step 1.2**: Add validation to `internal/server/server.go` (after line 316)

```go
// Verify branch exists before attempting to focus
_, err := s.storage.GetBranch(input.BranchID)
if err != nil {
	// Get list of available branches for helpful error
	branches := s.storage.ListBranches()
	branchIDs := make([]string, len(branches))
	for i, b := range branches {
		branchIDs[i] = b.ID
	}
	return nil, nil, fmt.Errorf("cannot focus on branch %s: branch does not exist. Available branches: %v", input.BranchID, branchIDs)
}
```

**Step 1.3**: Test the fix
```bash
cd internal/storage
go test -run TestSetActiveBranch -v

cd ../server
go test -run TestFocusBranch -v
```

**Checkpoint**: focus-branch should now provide helpful error messages ✅

---

### Task 2: Verify Syntax Checker (30 minutes)

**Step 2.1**: Create test case in `internal/validation/logic_test.go`

```go
func TestCheckWellFormed_UnbalancedParentheses(t *testing.T) {
	validator := NewLogicValidator()

	testCases := []struct {
		statement string
		shouldBeWellFormed bool
	}{
		{"If (A then B", false}, // Unbalanced
		{"If (A) then B", true}, // Balanced
		{"(A and B) or C", true}, // Balanced
		{"A and (B or C", false}, // Unbalanced
	}

	for _, tc := range testCases {
		checks := validator.CheckWellFormed([]string{tc.statement})
		if checks[0].IsWellFormed != tc.shouldBeWellFormed {
			t.Errorf("Statement %q: expected well-formed=%v, got %v. Issues: %v",
				tc.statement, tc.shouldBeWellFormed, checks[0].IsWellFormed, checks[0].Issues)
		}
	}
}
```

**Step 2.2**: Run test
```bash
cd internal/validation
go test -run TestCheckWellFormed_UnbalancedParentheses -v
```

**Checkpoint**: Test should pass (parentheses check already works) ✅

---

## Day 2-3: Cognitive Tools Foundation (10 hours)

### Task 3: Update Server Structure (1 hour)

**Step 3.1**: Add imports to `internal/server/server.go` (after line 21)

```go
import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)
```

**Step 3.2**: Update UnifiedServer struct (after line 39)

```go
type UnifiedServer struct {
	storage            *storage.MemoryStorage
	linear             *modes.LinearMode
	tree               *modes.TreeMode
	divergent          *modes.DivergentMode
	auto               *modes.AutoMode
	validator          *validation.LogicValidator

	// Cognitive reasoning analyzers
	probReasoner          *reasoning.ProbabilisticReasoner
	decisionMaker         *reasoning.DecisionMaker
	problemDecomposer     *reasoning.ProblemDecomposer
	evidenceAnalyzer      *analysis.EvidenceAnalyzer
	contradictionDetector *analysis.ContradictionDetector
	sensitivityAnalyzer   *analysis.SensitivityAnalyzer
	selfEvaluator         *metacognition.SelfEvaluator
	biasDetector          *metacognition.BiasDetector
}
```

**Step 3.3**: Update NewUnifiedServer constructor (after line 42)

```go
func NewUnifiedServer(
	store *storage.MemoryStorage,
	linear *modes.LinearMode,
	tree *modes.TreeMode,
	divergent *modes.DivergentMode,
	auto *modes.AutoMode,
	validator *validation.LogicValidator,
) *UnifiedServer {
	return &UnifiedServer{
		storage:               store,
		linear:                linear,
		tree:                  tree,
		divergent:             divergent,
		auto:                  auto,
		validator:             validator,
		probReasoner:          reasoning.NewProbabilisticReasoner(),
		decisionMaker:         reasoning.NewDecisionMaker(),
		problemDecomposer:     reasoning.NewProblemDecomposer(),
		evidenceAnalyzer:      analysis.NewEvidenceAnalyzer(),
		contradictionDetector: analysis.NewContradictionDetector(),
		sensitivityAnalyzer:   analysis.NewSensitivityAnalyzer(),
		selfEvaluator:         metacognition.NewSelfEvaluator(),
		biasDetector:          metacognition.NewBiasDetector(),
	}
}
```

**Checkpoint**: Server compiles with new analyzers ✅

---

### Task 4: Register New Tools (2 hours)

**Step 4.1**: Add tool registrations to `RegisterTools()` (after line 114)

```go
func (s *UnifiedServer) RegisterTools(mcpServer *mcp.Server) {
	// ... existing 11 tools ...

	// Cognitive reasoning tools
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "probabilistic-reasoning",
		Description: "Perform Bayesian inference and update beliefs based on evidence",
	}, s.handleProbabilisticReasoning)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "assess-evidence",
		Description: "Assess quality, reliability, and relevance of evidence",
	}, s.handleAssessEvidence)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "detect-contradictions",
		Description: "Detect contradictions between multiple thoughts",
	}, s.handleDetectContradictions)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "make-decision",
		Description: "Structured multi-criteria decision analysis",
	}, s.handleMakeDecision)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "decompose-problem",
		Description: "Break down complex problems into manageable subproblems",
	}, s.handleDecomposeProblem)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "sensitivity-analysis",
		Description: "Test robustness of conclusions to assumption changes",
	}, s.handleSensitivityAnalysis)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "self-evaluate",
		Description: "Perform metacognitive self-assessment of reasoning quality",
	}, s.handleSelfEvaluate)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "detect-biases",
		Description: "Identify cognitive biases in reasoning",
	}, s.handleDetectBiases)
}
```

**Step 4.2**: Test compilation
```bash
cd internal/server
go build
```

**Checkpoint**: Server compiles with 19 tool registrations ✅

---

### Task 5: Create Validation File (4 hours)

**Step 5.1**: Create `internal/server/validation.go`

Copy all validation function templates from COMPREHENSIVE_IMPROVEMENT_PLAN.md, Appendix B.

**Key functions to implement**:
- `ValidateProbabilisticReasoningRequest()`
- `ValidateAssessEvidenceRequest()`
- `ValidateDetectContradictionsRequest()`
- `ValidateMakeDecisionRequest()`
- `ValidateDecomposeProblemRequest()`
- `ValidateSensitivityAnalysisRequest()`
- `ValidateSelfEvaluateRequest()`
- `ValidateDetectBiasesRequest()`

**Step 5.2**: Test validation functions
```bash
cd internal/server
go test -run TestValidation -v
```

**Checkpoint**: All validation functions working ✅

---

### Task 6: Implement Handler Stubs (3 hours)

**Step 6.1**: Create stub handlers in `internal/server/server.go` (at end of file)

For each of the 8 tools, create a stub handler following this pattern:

```go
// ============================================
// TOOL: probabilistic-reasoning
// ============================================

type ProbabilisticReasoningRequest struct {
	Action       string   `json:"action"`
	Statement    string   `json:"statement,omitempty"`
	PriorProb    float64  `json:"prior_prob,omitempty"`
	BeliefID     string   `json:"belief_id,omitempty"`
	EvidenceID   string   `json:"evidence_id,omitempty"`
	Likelihood   float64  `json:"likelihood,omitempty"`
	EvidenceProb float64  `json:"evidence_prob,omitempty"`
	BeliefIDs    []string `json:"belief_ids,omitempty"`
	Operation    string   `json:"operation,omitempty"`
}

type ProbabilisticReasoningResponse struct {
	Action       string                         `json:"action"`
	Belief       *types.ProbabilisticBelief     `json:"belief,omitempty"`
	Probability  float64                        `json:"probability,omitempty"`
	Message      string                         `json:"message"`
}

func (s *UnifiedServer) handleProbabilisticReasoning(ctx context.Context, req *mcp.CallToolRequest, input ProbabilisticReasoningRequest) (*mcp.CallToolResult, *ProbabilisticReasoningResponse, error) {
	// TODO: Implement handler logic
	return nil, nil, fmt.Errorf("not yet implemented")
}
```

**Step 6.2**: Test compilation
```bash
cd internal/server
go build
```

**Checkpoint**: All 8 handler stubs compile ✅

---

## Day 4-5: Handler Implementation (12 hours)

**Time per handler**: 1.5 hours × 8 handlers = 12 hours

### Handler Template

For each handler, follow this pattern:

```go
func (s *UnifiedServer) handle<ToolName>(ctx context.Context, req *mcp.CallToolRequest, input <ToolName>Request) (*mcp.CallToolResult, *<ToolName>Response, error) {
	// Step 1: Validate input
	if err := Validate<ToolName>Request(&input); err != nil {
		return nil, nil, err
	}

	// Step 2: Call underlying analyzer
	result, err := s.<analyzer>.<Method>(...)
	if err != nil {
		return nil, nil, err
	}

	// Step 3: Format response
	response := &<ToolName>Response{
		// Map result fields to response fields
	}

	// Step 4: Return MCP result
	return &mcp.CallToolResult{
		Content: toJSONContent(response),
	}, response, nil
}
```

### Implementation Order

1. **probabilistic-reasoning** (1.5h) - See detailed example in COMPREHENSIVE_IMPROVEMENT_PLAN.md Section 4
2. **assess-evidence** (1.5h)
3. **detect-contradictions** (1.5h)
4. **make-decision** (1.5h)
5. **decompose-problem** (1.5h)
6. **sensitivity-analysis** (1.5h)
7. **self-evaluate** (1.5h)
8. **detect-biases** (1.5h)

### Testing After Each Handler

```bash
# Unit test individual handler
go test -run TestHandle<ToolName> -v

# Integration test
go test -run TestIntegration<ToolName> -v
```

**Checkpoint**: All 8 handlers implemented and tested ✅

---

## Day 6: Integration Testing (4 hours)

### Task 7: Create Integration Tests (3 hours)

**Step 7.1**: Add tests to `internal/server/server_test.go`

```go
func TestProbabilisticReasoning_FullWorkflow(t *testing.T) {
	// Test create → update → combine workflow
	server := setupTestServer(t)

	// Create belief
	createReq := ProbabilisticReasoningRequest{
		Action: "create_belief",
		Statement: "Test hypothesis",
		PriorProb: 0.7,
	}
	createResp, err := server.handleProbabilisticReasoning(context.Background(), nil, createReq)
	if err != nil {
		t.Fatalf("Create belief failed: %v", err)
	}

	// Update belief
	updateReq := ProbabilisticReasoningRequest{
		Action: "update_belief",
		BeliefID: createResp.Belief.ID,
		EvidenceID: "evidence-1",
		Likelihood: 0.9,
		EvidenceProb: 0.5,
	}
	updateResp, err := server.handleProbabilisticReasoning(context.Background(), nil, updateReq)
	if err != nil {
		t.Fatalf("Update belief failed: %v", err)
	}

	// Verify posterior probability increased
	if updateResp.Belief.Probability <= 0.7 {
		t.Errorf("Expected probability increase, got %.2f", updateResp.Belief.Probability)
	}
}

// Similar tests for all 8 tools
```

**Step 7.2**: Run full test suite
```bash
cd internal/server
go test -v
```

**Checkpoint**: All integration tests passing ✅

---

### Task 8: Manual Testing in Claude Desktop (1 hour)

**Step 8.1**: Build and run server
```bash
go build -o unified-thinking.exe cmd/unified-thinking/main.go
./unified-thinking.exe
```

**Step 8.2**: Test each tool in Claude Desktop

1. Test probabilistic-reasoning:
```json
{
  "action": "create_belief",
  "statement": "The feature will ship on time",
  "prior_prob": 0.6
}
```

2. Test assess-evidence:
```json
{
  "content": "Multiple peer-reviewed studies show...",
  "source": "Journal of Science",
  "claim_id": "claim-1",
  "supports_claim": true
}
```

3. Test remaining 6 tools...

**Checkpoint**: All tools functional in Claude Desktop ✅

---

## Day 7: Documentation & Polish (3 hours)

### Task 9: Update Documentation (2 hours)

**Step 9.1**: Update `README.md`

Add to features section:
```markdown
### Cognitive Reasoning Tools (New!)

- **probabilistic-reasoning**: Bayesian inference and belief updates
- **assess-evidence**: Evidence quality assessment
- **detect-contradictions**: Cross-thought contradiction detection
- **make-decision**: Multi-criteria decision analysis
- **decompose-problem**: Problem breakdown with dependencies
- **sensitivity-analysis**: Robustness testing
- **self-evaluate**: Metacognitive self-assessment
- **detect-biases**: Cognitive bias identification
```

**Step 9.2**: Create `docs/TOOLS.md`

Document all 19 tools with:
- Description
- Parameters
- Return values
- Usage examples

**Step 9.3**: Create `examples/` directory with usage examples

**Checkpoint**: Documentation complete ✅

---

### Task 10: Final Testing & Polish (1 hour)

**Step 10.1**: Run complete test suite
```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Step 10.2**: Verify 99% test pass rate target

**Step 10.3**: Code review checklist
- [ ] All handlers implemented
- [ ] All validation functions working
- [ ] Error messages helpful
- [ ] Documentation complete
- [ ] Examples provided
- [ ] Tests passing
- [ ] Performance acceptable (<1s per operation)

**Checkpoint**: System production-ready ✅

---

## Bonus: Prove Tool Enhancement (2 hours)

### Task 11: Fix Universal Instantiation (Optional)

See COMPREHENSIVE_IMPROVEMENT_PLAN.md Section 3 for detailed implementation.

**Location**: `internal/validation/logic.go` lines 371-435

**Key improvements**:
- Add singular/plural pattern matching
- Handle irregular plurals (men/man, people/person)
- Support "is human" in addition to "is a human"
- Fix verb conjugation (are → is)

**Testing**:
```bash
cd internal/validation
go test -run TestProve_UniversalInstantiation -v
```

---

## Verification Checklist

After completing all tasks, verify:

### Code Quality
- [ ] All files compile without errors
- [ ] All unit tests passing (100% coverage maintained)
- [ ] All integration tests passing (>98% pass rate)
- [ ] No lint warnings
- [ ] Code follows existing patterns

### Functionality
- [ ] focus-branch error fixed with helpful messages
- [ ] All 8 cognitive tools registered
- [ ] All 8 handlers functional
- [ ] All tools work in Claude Desktop
- [ ] Error handling comprehensive

### Documentation
- [ ] README.md updated
- [ ] TOOLS.md created
- [ ] Examples provided
- [ ] Code comments clear
- [ ] API documentation complete

### Performance
- [ ] All operations <1 second
- [ ] No memory leaks
- [ ] No race conditions
- [ ] Concurrent access safe

---

## Common Issues & Solutions

### Issue 1: Import Errors
**Problem**: Cannot find package
**Solution**: Run `go mod tidy` to resolve dependencies

### Issue 2: Type Mismatch
**Problem**: Handler signature doesn't match MCP SDK
**Solution**: Check that handler returns `(*mcp.CallToolResult, *ResponseType, error)`

### Issue 3: Test Failures
**Problem**: Integration tests fail
**Solution**: Ensure test server is properly initialized with all analyzers

### Issue 4: Performance Issues
**Problem**: Operations taking >1 second
**Solution**: Add profiling and identify bottlenecks

---

## Success Metrics

After implementation, verify these metrics:

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| MCP Tools | 11 | 19 | 19 ✅ |
| Test Pass Rate | 97.6% | >98% | 99% |
| Tool Coverage | Basic | Advanced | Complete ✅ |
| Documentation | Good | Excellent | Complete ✅ |
| Performance | <1s | <1s | <1s ✅ |

---

## Next Steps After Implementation

1. **Beta Testing**: Deploy to beta users
2. **Monitoring**: Track tool usage metrics
3. **Feedback**: Collect user feedback
4. **Iteration**: Refine based on usage patterns
5. **Production**: Full production deployment

---

## Resources

- **Detailed Plan**: COMPREHENSIVE_IMPROVEMENT_PLAN.md (72KB)
- **Executive Summary**: EXECUTIVE_SUMMARY.md
- **Test Results**: MANUAL_TEST_RESULTS.md
- **Issues List**: ISSUES.md

---

**Quick Start Guide Version**: 1.0
**Last Updated**: 2025-01-10
**Estimated Time**: 30 hours (1 week)
**Difficulty**: Medium
