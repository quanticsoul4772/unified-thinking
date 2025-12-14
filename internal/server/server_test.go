package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/server/handlers"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/testutil"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// setupTestServer creates a fully initialized server for testing.
// FAILS the test if required API keys are not set.
// Uses mock hypothesis generator to avoid real API calls for abductive reasoning.
func setupTestServer(t *testing.T) *UnifiedServer {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Fatal("ANTHROPIC_API_KEY not set - required for full server")
	}
	if os.Getenv("VOYAGE_API_KEY") == "" {
		t.Fatal("VOYAGE_API_KEY not set - required for embeddings")
	}

	store := storage.NewMemoryStorage()
	linear := modes.NewLinearMode(store)
	tree := modes.NewTreeMode(store)
	divergent := modes.NewDivergentMode(store)
	auto := modes.NewAutoMode(linear, tree, divergent)
	validator := validation.NewLogicValidator()

	srv, err := NewUnifiedServer(store, linear, tree, divergent, auto, validator)
	if err != nil {
		t.Fatalf("setupTestServer failed: %v", err)
	}

	// Inject mock hypothesis generator for abductive reasoning tests
	// This avoids real LLM API calls that would timeout in CI
	mockGen := testutil.NewMockHypothesisGenerator()
	abductiveReasoner := reasoning.NewAbductiveReasoner(store, mockGen)
	srv.SetAbductiveHandler(handlers.NewAbductiveHandler(abductiveReasoner, store))

	return srv
}

type stubExecutor struct {
	result    interface{}
	err       error
	lastTool  string
	lastInput map[string]interface{}
}

func (s *stubExecutor) ExecuteTool(_ context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	s.lastTool = toolName
	s.lastInput = input
	if s.err != nil {
		return nil, s.err
	}
	if s.result != nil {
		return s.result, nil
	}
	return map[string]interface{}{"id": "result-1", "confidence": 0.9}, nil
}

func TestHandleThink_LinearMode(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	input := ThinkRequest{
		Content:    "Test linear reasoning thought",
		Mode:       "linear",
		Confidence: 0.8,
	}

	result, resp, err := server.handleThink(ctx, nil, input)
	if err != nil {
		t.Fatalf("handleThink() error = %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if resp.ThoughtID == "" {
		t.Error("ThoughtID should not be empty")
	}

	if resp.Mode != "linear" {
		t.Errorf("Mode = %v, want linear", resp.Mode)
	}

	if resp.Confidence != 0.8 {
		t.Errorf("Confidence = %v, want 0.8", resp.Confidence)
	}
}

func TestHandleThink_TreeMode(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	input := ThinkRequest{
		Content:    "Branch thought for exploration",
		Mode:       "tree",
		KeyPoints:  []string{"key1", "key2"},
		Confidence: 0.85,
	}

	_, resp, err := server.handleThink(ctx, nil, input)
	if err != nil {
		t.Fatalf("handleThink() error = %v", err)
	}

	if resp.BranchID == "" {
		t.Error("BranchID should be created for tree mode")
	}

	if resp.InsightCount == 0 {
		t.Error("Insights should be generated from key points")
	}
}

func TestHandleThink_DivergentMode(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	input := ThinkRequest{
		Content:        "Creative problem solving",
		Mode:           "divergent",
		ForceRebellion: true,
		Confidence:     0.7,
	}

	_, resp, err := server.handleThink(ctx, nil, input)
	if err != nil {
		t.Fatalf("handleThink() error = %v", err)
	}

	// Note: IsRebellion flag is stored in the thought metadata, not in response
	if resp.ThoughtID == "" {
		t.Error("ThoughtID should be set")
	}
}

func TestHandleThink_AutoMode(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedModes []string // Allow multiple valid modes for semantic selection
	}{
		{
			name:          "creative keywords trigger divergent",
			content:       "Let's think creatively about this problem",
			expectedModes: []string{"divergent"},
		},
		{
			name:          "explore keywords trigger tree or divergent",
			content:       "Let's explore alternative approaches",
			expectedModes: []string{"tree", "divergent"}, // Both valid with semantic embeddings
		},
		{
			name:          "simple content defaults to linear",
			content:       "Calculate the result",
			expectedModes: []string{"linear"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServer(t)
			ctx := context.Background()

			input := ThinkRequest{
				Content:    tt.content,
				Mode:       "auto",
				Confidence: 0.8,
			}

			_, resp, err := server.handleThink(ctx, nil, input)
			if err != nil {
				t.Fatalf("handleThink() error = %v", err)
			}

			// Check if selected mode is in the list of expected modes
			found := false
			for _, expected := range tt.expectedModes {
				if resp.Mode == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Auto mode selected %v, want one of %v", resp.Mode, tt.expectedModes)
			}
		})
	}
}

func TestHandleThink_ValidationErrors(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   ThinkRequest
		wantErr bool
	}{
		{
			name: "empty content",
			input: ThinkRequest{
				Content: "",
				Mode:    "linear",
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			input: ThinkRequest{
				Content: "Test",
				Mode:    "invalid_mode",
			},
			wantErr: true,
		},
		{
			name: "confidence out of range",
			input: ThinkRequest{
				Content:    "Test",
				Mode:       "linear",
				Confidence: 1.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleThink(ctx, nil, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleThink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleHistory(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create some thoughts first
	for i := 0; i < 5; i++ {
		input := ThinkRequest{
			Content:    "Test thought",
			Mode:       "linear",
			Confidence: 0.8,
		}
		server.handleThink(ctx, nil, input)
	}

	// Test history retrieval
	histReq := HistoryRequest{
		Limit:  3,
		Offset: 0,
	}

	_, resp, err := server.handleHistory(ctx, nil, histReq)
	if err != nil {
		t.Fatalf("handleHistory() error = %v", err)
	}

	if len(resp.Thoughts) != 3 {
		t.Errorf("Expected 3 thoughts, got %d", len(resp.Thoughts))
	}
}

func TestHandleListBranches(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create some branches with explicit branch IDs to ensure separate branches
	for i := 0; i < 3; i++ {
		input := ThinkRequest{
			Content:    "Branch thought",
			Mode:       "tree",
			BranchID:   fmt.Sprintf("test-branch-%d", i),
			Confidence: 0.8,
		}
		server.handleThink(ctx, nil, input)
	}

	_, resp, err := server.handleListBranches(ctx, nil, EmptyRequest{})
	if err != nil {
		t.Fatalf("handleListBranches() error = %v", err)
	}

	if len(resp.Branches) != 3 {
		t.Errorf("Expected 3 branches, got %d", len(resp.Branches))
	}
}

func TestHandleFocusBranch(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create a branch
	input := ThinkRequest{
		Content:    "Branch thought",
		Mode:       "tree",
		Confidence: 0.8,
	}
	_, thinkResp, _ := server.handleThink(ctx, nil, input)

	// Focus on the branch
	focusReq := FocusBranchRequest{
		BranchID: thinkResp.BranchID,
	}

	_, resp, err := server.handleFocusBranch(ctx, nil, focusReq)
	if err != nil {
		t.Fatalf("handleFocusBranch() error = %v", err)
	}

	if resp.ActiveBranchID != thinkResp.BranchID {
		t.Errorf("ActiveBranchID = %v, want %v", resp.ActiveBranchID, thinkResp.BranchID)
	}
}

func TestHandleValidate(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create a thought
	input := ThinkRequest{
		Content:    "If it rains then the ground is wet",
		Mode:       "linear",
		Confidence: 0.8,
	}
	_, thinkResp, _ := server.handleThink(ctx, nil, input)

	// Validate the thought
	valReq := ValidateRequest{
		ThoughtID: thinkResp.ThoughtID,
	}

	_, resp, err := server.handleValidate(ctx, nil, valReq)
	if err != nil {
		t.Fatalf("handleValidate() error = %v", err)
	}

	// Validation response contains IsValid and Reason
	if resp.Reason == "" {
		t.Error("Validation reason should not be empty")
	}
}

func TestHandleProve(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	proveReq := ProveRequest{
		Premises:   []string{"If P then Q", "P"},
		Conclusion: "Q",
	}

	_, resp, err := server.handleProve(ctx, nil, proveReq)
	if err != nil {
		t.Fatalf("handleProve() error = %v", err)
	}

	if !resp.IsProvable {
		t.Error("Modus Ponens should be provable")
	}

	if len(resp.Steps) == 0 {
		t.Error("Proof steps should not be empty")
	}
}

func TestHandleSearch(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create thoughts with specific content
	contents := []string{
		"machine learning algorithm",
		"deep learning neural network",
		"random forest classifier",
	}

	for _, content := range contents {
		input := ThinkRequest{
			Content:    content,
			Mode:       "linear",
			Confidence: 0.8,
		}
		server.handleThink(ctx, nil, input)
	}

	// Search for "learning"
	searchReq := SearchRequest{
		Query: "learning",
		Limit: 10,
	}

	_, resp, err := server.handleSearch(ctx, nil, searchReq)
	if err != nil {
		t.Fatalf("handleSearch() error = %v", err)
	}

	if len(resp.Thoughts) != 2 {
		t.Errorf("Expected 2 thoughts with 'learning', got %d", len(resp.Thoughts))
	}
}

func TestHandleGetMetrics(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create some data
	for i := 0; i < 5; i++ {
		input := ThinkRequest{
			Content:    "Test thought",
			Mode:       "linear",
			Confidence: 0.8,
		}
		server.handleThink(ctx, nil, input)
	}

	_, resp, err := server.handleGetMetrics(ctx, nil, EmptyRequest{})
	if err != nil {
		t.Fatalf("handleGetMetrics() error = %v", err)
	}

	if resp.TotalThoughts != 5 {
		t.Errorf("TotalThoughts = %d, want 5", resp.TotalThoughts)
	}
}

func TestHandleRecentBranches(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create and access multiple branches
	var branchIDs []string
	for i := 0; i < 5; i++ {
		input := ThinkRequest{
			Content:    "Branch thought",
			Mode:       "tree",
			Confidence: 0.8,
		}
		_, resp, _ := server.handleThink(ctx, nil, input)
		branchIDs = append(branchIDs, resp.BranchID)
	}

	// Get recent branches
	_, resp, err := server.handleRecentBranches(ctx, nil, EmptyRequest{})
	if err != nil {
		t.Fatalf("handleRecentBranches() error = %v", err)
	}

	if len(resp.RecentBranches) > 3 {
		t.Errorf("Expected at most 3 recent branches, got %d", len(resp.RecentBranches))
	}

	// Most recent should be first
	if len(resp.RecentBranches) > 0 && resp.RecentBranches[0].ID != branchIDs[len(branchIDs)-1] {
		t.Error("Most recent branch should be first")
	}
}

func TestHandleThink_WithValidation(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	input := ThinkRequest{
		Content:           "If it rains then the ground is wet",
		Mode:              "linear",
		Confidence:        0.8,
		RequireValidation: true,
	}

	_, resp, err := server.handleThink(ctx, nil, input)
	if err != nil {
		t.Fatalf("handleThink() error = %v", err)
	}

	// Validation should have been performed
	// Note: IsValid will depend on validator's assessment
	if resp.ThoughtID == "" {
		t.Error("ThoughtID should not be empty even with validation")
	}
}

func TestConcurrentThinkOperations(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Test concurrent think operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			input := ThinkRequest{
				Content:    "Concurrent thought",
				Mode:       "linear",
				Confidence: 0.8,
			}
			_, _, err := server.handleThink(ctx, nil, input)
			if err != nil {
				t.Errorf("Concurrent handleThink() error = %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all thoughts were stored
	histReq := HistoryRequest{Limit: 100}
	_, resp, _ := server.handleHistory(ctx, nil, histReq)

	if len(resp.Thoughts) != 10 {
		t.Errorf("Expected 10 thoughts after concurrent operations, got %d", len(resp.Thoughts))
	}
}

// TestHandleDetectBiases_WithFallacies tests the integrated bias and fallacy detection
func TestHandleDetectBiases_WithFallacies(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create a thought with content that should trigger both biases and fallacies
	thoughtContent := `This clearly shows and confirms our hypothesis was right.
	Everyone knows this is true, it's obviously the case.
	I recently heard about this, so it must be common.
	You can't trust critics because they're just negative people.
	If we don't buy this now, we'll miss out forever.`

	// Store the thought first
	thought := &types.Thought{
		ID:         "test-thought-1",
		Content:    thoughtContent,
		Mode:       types.ModeLinear,
		Confidence: 0.7,
		Timestamp:  time.Now(),
	}
	err := server.storage.StoreThought(thought)
	if err != nil {
		t.Fatalf("Failed to store thought: %v", err)
	}

	// Test detect-biases with thought ID
	input := handlers.DetectBiasesRequest{
		ThoughtID: "test-thought-1",
	}

	result, response, err := server.handleDetectBiases(ctx, nil, input)
	if err != nil {
		t.Fatalf("handleDetectBiases() error = %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if response.Status != "success" {
		t.Errorf("Status = %v, want success", response.Status)
	}

	// Check that we have both biases and fallacies
	if len(response.Biases) == 0 {
		t.Error("Expected at least one bias to be detected")
	}

	if len(response.Fallacies) == 0 {
		t.Error("Expected at least one fallacy to be detected")
	}

	// Check that combined list contains both types
	if len(response.Combined) == 0 {
		t.Error("Combined list should not be empty")
	}

	// Count should match the total
	expectedCount := len(response.Biases) + len(response.Fallacies)
	if response.Count != expectedCount {
		t.Errorf("Count = %d, want %d", response.Count, expectedCount)
	}

	// Verify combined list structure
	hasBias := false
	hasFallacy := false
	for _, issue := range response.Combined {
		if issue.Type == "bias" {
			hasBias = true
		}
		if issue.Type == "fallacy" {
			hasFallacy = true
		}
		// Verify required fields are populated
		if issue.Name == "" {
			t.Error("Issue name should not be empty")
		}
		if issue.Category == "" {
			t.Error("Issue category should not be empty")
		}
		if issue.Description == "" {
			t.Error("Issue description should not be empty")
		}
		if issue.Mitigation == "" {
			t.Error("Issue mitigation should not be empty")
		}
	}

	if !hasBias {
		t.Error("Combined list should contain at least one bias")
	}

	if !hasFallacy {
		t.Error("Combined list should contain at least one fallacy")
	}
}

// TestHandleDetectBiases_BranchAnalysis tests bias/fallacy detection on a branch
func TestHandleDetectBiases_BranchAnalysis(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// Create a branch with multiple thoughts
	branch := &types.Branch{
		ID:        "test-branch-1",
		Thoughts:  []*types.Thought{},
		CreatedAt: time.Now(),
	}

	// Add thoughts with various biases and fallacies
	thoughtContents := []string{
		"The first person I asked agreed, so everyone must agree.", // Hasty generalization
		"You're wrong because you're not an expert.",               // Ad hominem
		"We've always done it this way, so it must be right.",      // Appeal to tradition
	}

	for i, content := range thoughtContents {
		thought := &types.Thought{
			ID:         fmt.Sprintf("thought-%d", i),
			Content:    content,
			Mode:       types.ModeTree,
			BranchID:   branch.ID,
			Confidence: 0.7,
			Timestamp:  time.Now().Add(time.Duration(i) * time.Second),
		}
		branch.Thoughts = append(branch.Thoughts, thought)
		_ = server.storage.StoreThought(thought)
	}

	// Store the branch
	err := server.storage.StoreBranch(branch)
	if err != nil {
		t.Fatalf("Failed to store branch: %v", err)
	}

	// Test detect-biases with branch ID
	input := handlers.DetectBiasesRequest{
		BranchID: "test-branch-1",
	}

	result, response, err := server.handleDetectBiases(ctx, nil, input)
	if err != nil {
		t.Fatalf("handleDetectBiases() error = %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if response.Status != "success" {
		t.Errorf("Status = %v, want success", response.Status)
	}

	// Should detect multiple issues from the branch
	if len(response.Combined) == 0 {
		t.Error("Should detect issues in branch thoughts")
	}

	// Verify we're analyzing the whole branch
	if response.Count == 0 {
		t.Error("Count should be greater than 0 for branch with problematic content")
	}
}

func TestHandleExecuteWorkflow_Success(t *testing.T) {
	server := setupTestServer(t)
	exec := &stubExecutor{}
	orch := orchestration.NewOrchestratorWithExecutor(exec)

	workflow := &orchestration.Workflow{
		ID:   "wf-success",
		Name: "Workflow Success",
		Type: orchestration.WorkflowSequential,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:      "step-1",
				Tool:    "think",
				Input:   map[string]interface{}{"content": "$problem", "mode": "linear"},
				StoreAs: "analysis",
			},
		},
	}

	if err := orch.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("RegisterWorkflow() error = %v", err)
	}

	server.SetOrchestrator(orch)

	ctx := context.Background()
	req := ExecuteWorkflowRequest{
		WorkflowID: "wf-success",
		Input:      map[string]interface{}{"problem": "Investigate latency"},
	}

	result, resp, err := server.handleExecuteWorkflow(ctx, nil, req)
	if err != nil {
		t.Fatalf("handleExecuteWorkflow() returned error = %v", err)
	}

	if result == nil {
		t.Fatal("expected call result")
	}

	if resp.Status != "completed" {
		t.Fatalf("resp.Status = %s, want completed", resp.Status)
	}

	if resp.Result == nil {
		t.Fatal("expected workflow result payload")
	}

	if resp.Result.Status != "success" {
		t.Fatalf("workflow status = %s, want success", resp.Result.Status)
	}

	if exec.lastTool != "think" {
		t.Fatalf("expected executor to invoke think, got %s", exec.lastTool)
	}

	if got := exec.lastInput["content"]; got != "Investigate latency" {
		t.Fatalf("expected content to resolve template, got %v", got)
	}

	if _, ok := resp.Result.StepResults["step-1"]; !ok {
		t.Fatal("expected step result for step-1")
	}
}

func TestHandleExecuteWorkflow_ToolError(t *testing.T) {
	server := setupTestServer(t)
	exec := &stubExecutor{err: errors.New("tool failure")}
	orch := orchestration.NewOrchestratorWithExecutor(exec)

	workflow := &orchestration.Workflow{
		ID:   "wf-error",
		Name: "Workflow Error",
		Type: orchestration.WorkflowSequential,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:    "step-1",
				Tool:  "think",
				Input: map[string]interface{}{"content": "$problem"},
			},
		},
	}

	if err := orch.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("RegisterWorkflow() error = %v", err)
	}

	server.SetOrchestrator(orch)

	ctx := context.Background()
	req := ExecuteWorkflowRequest{
		WorkflowID: "wf-error",
		Input:      map[string]interface{}{"problem": "Investigate latency"},
	}

	result, resp, err := server.handleExecuteWorkflow(ctx, nil, req)
	if err != nil {
		t.Fatalf("handleExecuteWorkflow() returned error = %v", err)
	}

	if result == nil {
		t.Fatal("expected call result")
	}

	if resp.Status != "failed" {
		t.Fatalf("resp.Status = %s, want failed", resp.Status)
	}

	if resp.Result != nil {
		t.Fatal("expected no workflow result when execution fails")
	}

	if resp.Error == "" {
		t.Fatal("expected error message in response")
	}
}

func TestHandleListWorkflows_NotInitialized(t *testing.T) {
	server := setupTestServer(t)
	server.orchestrator = nil

	ctx := context.Background()
	_, _, err := server.handleListWorkflows(ctx, nil, EmptyRequest{})
	if err == nil || err.Error() != "orchestrator not initialized" {
		t.Fatalf("expected orchestrator not initialized error, got %v", err)
	}
}

func TestHandleListWorkflows_ReturnsRegistered(t *testing.T) {
	server := setupTestServer(t)
	exec := &stubExecutor{}
	orch := orchestration.NewOrchestratorWithExecutor(exec)

	workflow := &orchestration.Workflow{
		ID:   "wf-list",
		Name: "Workflow List",
		Type: orchestration.WorkflowSequential,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:    "step-1",
				Tool:  "think",
				Input: map[string]interface{}{},
			},
		},
	}

	if err := orch.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("RegisterWorkflow() error = %v", err)
	}

	server.SetOrchestrator(orch)

	ctx := context.Background()
	_, resp, err := server.handleListWorkflows(ctx, nil, EmptyRequest{})
	if err != nil {
		t.Fatalf("handleListWorkflows() error = %v", err)
	}

	if resp.Count != 1 {
		t.Fatalf("resp.Count = %d, want 1", resp.Count)
	}

	if len(resp.Workflows) != 1 {
		t.Fatalf("Workflows length = %d, want 1", len(resp.Workflows))
	}

	if resp.Workflows[0].ID != "wf-list" {
		t.Fatalf("expected workflow ID wf-list, got %s", resp.Workflows[0].ID)
	}
}

func TestHandleRegisterWorkflow(t *testing.T) {
	server := setupTestServer(t)
	exec := &stubExecutor{}
	orch := orchestration.NewOrchestratorWithExecutor(exec)
	server.SetOrchestrator(orch)

	ctx := context.Background()

	req := RegisterWorkflowRequest{
		Workflow: &orchestration.Workflow{
			ID:   "new-workflow",
			Name: "New Workflow",
			Type: orchestration.WorkflowSequential,
			Steps: []*orchestration.WorkflowStep{
				{
					ID:    "step-1",
					Tool:  "think",
					Input: map[string]interface{}{},
				},
			},
		},
	}

	result, resp, err := server.handleRegisterWorkflow(ctx, nil, req)
	if err != nil {
		t.Fatalf("handleRegisterWorkflow() error = %v", err)
	}

	if result == nil {
		t.Fatal("expected call result")
	}

	if !resp.Success {
		t.Fatalf("expected workflow registration success, got %+v", resp)
	}

	// Attempt duplicate registration
	reqDuplicate := RegisterWorkflowRequest{
		Workflow: &orchestration.Workflow{
			ID:   "new-workflow",
			Name: "Duplicate",
			Type: orchestration.WorkflowSequential,
			Steps: []*orchestration.WorkflowStep{
				{
					ID:    "step-1",
					Tool:  "think",
					Input: map[string]interface{}{},
				},
			},
		},
	}

	_, duplicateResp, err := server.handleRegisterWorkflow(ctx, nil, reqDuplicate)
	if err != nil {
		t.Fatalf("handleRegisterWorkflow() duplicate error = %v", err)
	}

	if duplicateResp.Success {
		t.Fatal("expected duplicate registration to fail")
	}

	if duplicateResp.Error == "" {
		t.Fatal("expected duplicate registration error message")
	}
}
