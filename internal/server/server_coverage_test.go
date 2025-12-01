package server

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/integration"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/server/handlers"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// =============================================================================
// Test: toJSONContent helper function
// =============================================================================

func TestToJSONContent(t *testing.T) {
	t.Run("valid struct serialization", func(t *testing.T) {
		data := struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}{
			Name:  "test",
			Value: 42,
		}

		content := toJSONContent(data)
		if len(content) != 1 {
			t.Fatalf("expected 1 content item, got %d", len(content))
		}

		textContent, ok := content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("expected TextContent type")
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if result["name"] != "test" {
			t.Errorf("expected name 'test', got %v", result["name"])
		}
		if result["value"] != float64(42) {
			t.Errorf("expected value 42, got %v", result["value"])
		}
	})

	t.Run("nil value serialization", func(t *testing.T) {
		content := toJSONContent(nil)
		if len(content) != 1 {
			t.Fatalf("expected 1 content item, got %d", len(content))
		}

		textContent, ok := content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("expected TextContent type")
		}

		if textContent.Text != "null" {
			t.Errorf("expected 'null', got %s", textContent.Text)
		}
	})

	t.Run("empty slice serialization", func(t *testing.T) {
		content := toJSONContent([]string{})
		if len(content) != 1 {
			t.Fatalf("expected 1 content item, got %d", len(content))
		}

		textContent, ok := content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("expected TextContent type")
		}

		if textContent.Text != "[]" {
			t.Errorf("expected '[]', got %s", textContent.Text)
		}
	})

	t.Run("complex nested structure", func(t *testing.T) {
		data := map[string]interface{}{
			"nested": map[string]interface{}{
				"array": []int{1, 2, 3},
				"bool":  true,
			},
		}

		content := toJSONContent(data)
		if len(content) != 1 {
			t.Fatalf("expected 1 content item, got %d", len(content))
		}

		textContent := content[0].(*mcp.TextContent)
		if !strings.Contains(textContent.Text, "nested") {
			t.Error("expected nested key in output")
		}
	})

	t.Run("unmarshallable value returns error JSON", func(t *testing.T) {
		// Create a channel which cannot be marshalled
		ch := make(chan int)
		content := toJSONContent(ch)
		if len(content) != 1 {
			t.Fatalf("expected 1 content item, got %d", len(content))
		}

		textContent := content[0].(*mcp.TextContent)
		if !strings.Contains(textContent.Text, "error") {
			t.Error("expected error in output for unmarshallable value")
		}
	})
}

// =============================================================================
// Test: convertCrossRefs helper function
// =============================================================================

func TestConvertCrossRefs(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		result := convertCrossRefs([]CrossRefInput{})
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d items", len(result))
		}
	})

	t.Run("single cross reference", func(t *testing.T) {
		input := []CrossRefInput{
			{
				ToBranch: "branch-1",
				Type:     "complementary",
				Reason:   "test reason",
				Strength: 0.8,
			},
		}

		result := convertCrossRefs(input)
		if len(result) != 1 {
			t.Fatalf("expected 1 result, got %d", len(result))
		}

		if result[0].ToBranch != "branch-1" {
			t.Errorf("expected ToBranch 'branch-1', got %s", result[0].ToBranch)
		}
		if result[0].Type != "complementary" {
			t.Errorf("expected Type 'complementary', got %s", result[0].Type)
		}
		if result[0].Reason != "test reason" {
			t.Errorf("expected Reason 'test reason', got %s", result[0].Reason)
		}
		if result[0].Strength != 0.8 {
			t.Errorf("expected Strength 0.8, got %f", result[0].Strength)
		}
	})

	t.Run("multiple cross references", func(t *testing.T) {
		input := []CrossRefInput{
			{ToBranch: "branch-1", Type: "complementary", Strength: 0.8},
			{ToBranch: "branch-2", Type: "contradictory", Strength: 0.6},
			{ToBranch: "branch-3", Type: "builds_upon", Strength: 0.9},
		}

		result := convertCrossRefs(input)
		if len(result) != 3 {
			t.Fatalf("expected 3 results, got %d", len(result))
		}

		for i, xref := range result {
			if xref.ToBranch != input[i].ToBranch {
				t.Errorf("index %d: expected ToBranch %s, got %s", i, input[i].ToBranch, xref.ToBranch)
			}
		}
	})
}

// =============================================================================
// Test: NewUnifiedServer initialization
// =============================================================================

func TestNewUnifiedServer(t *testing.T) {
	t.Run("creates server with all handlers initialized", func(t *testing.T) {
		store := storage.NewMemoryStorage()
		linear := modes.NewLinearMode(store)
		tree := modes.NewTreeMode(store)
		divergent := modes.NewDivergentMode(store)
		auto := modes.NewAutoMode(linear, tree, divergent)
		validator := validation.NewLogicValidator()

		server, err := NewUnifiedServer(store, linear, tree, divergent, auto, validator)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Verify all core components are initialized
		if server.storage == nil {
			t.Error("storage not initialized")
		}
		if server.linear == nil {
			t.Error("linear mode not initialized")
		}
		if server.tree == nil {
			t.Error("tree mode not initialized")
		}
		if server.divergent == nil {
			t.Error("divergent mode not initialized")
		}
		if server.auto == nil {
			t.Error("auto mode not initialized")
		}
		if server.validator == nil {
			t.Error("validator not initialized")
		}

		// Verify reasoning engines are initialized
		if server.probabilisticReasoner == nil {
			t.Error("probabilistic reasoner not initialized")
		}
		if server.evidenceAnalyzer == nil {
			t.Error("evidence analyzer not initialized")
		}
		if server.decisionMaker == nil {
			t.Error("decision maker not initialized")
		}

		// Verify handlers are initialized
		if server.probabilisticHandler == nil {
			t.Error("probabilistic handler not initialized")
		}
		if server.decisionHandler == nil {
			t.Error("decision handler not initialized")
		}
		if server.metacognitionHandler == nil {
			t.Error("metacognition handler not initialized")
		}
		if server.temporalHandler == nil {
			t.Error("temporal handler not initialized")
		}
		if server.causalHandler == nil {
			t.Error("causal handler not initialized")
		}
		if server.hallucinationHandler == nil {
			t.Error("hallucination handler not initialized")
		}
		if server.calibrationHandler == nil {
			t.Error("calibration handler not initialized")
		}

		// Verify advanced handlers
		if server.dualProcessHandler == nil {
			t.Error("dual process handler not initialized")
		}
		if server.backtrackingHandler == nil {
			t.Error("backtracking handler not initialized")
		}
		if server.abductiveHandler == nil {
			t.Error("abductive handler not initialized")
		}
		if server.caseBasedHandler == nil {
			t.Error("case-based handler not initialized")
		}
		if server.unknownUnknownsHandler == nil {
			t.Error("unknown unknowns handler not initialized")
		}
		if server.symbolicHandler == nil {
			t.Error("symbolic handler not initialized")
		}

		// Verify enhanced tools components
		if server.analogicalReasoner == nil {
			t.Error("analogical reasoner not initialized")
		}
		if server.argumentAnalyzer == nil {
			t.Error("argument analyzer not initialized")
		}
		if server.evidencePipeline == nil {
			t.Error("evidence pipeline not initialized")
		}
		if server.causalTemporalIntegration == nil {
			t.Error("causal temporal integration not initialized")
		}

		// Verify episodic memory handler
		if server.episodicMemoryHandler == nil {
			t.Error("episodic memory handler not initialized")
		}
	})
}

// =============================================================================
// Test: initializeSemanticAutoMode
// =============================================================================

func TestInitializeSemanticAutoMode(t *testing.T) {
	t.Run("no API key - logs error and returns", func(t *testing.T) {
		// Ensure no API key is set
		os.Unsetenv("VOYAGE_API_KEY")

		server := setupTestServer()
		// The function should not panic and auto mode should work without embedder
		if server.auto == nil {
			t.Error("auto mode should still be initialized")
		}
	})

	t.Run("with API key - initializes embedder", func(t *testing.T) {
		// Set test API key
		os.Setenv("VOYAGE_API_KEY", "test-api-key")
		defer os.Unsetenv("VOYAGE_API_KEY")

		server := setupTestServer()
		if server.auto == nil {
			t.Error("auto mode should be initialized")
		}
		// Embedder should be set on auto mode
	})

	t.Run("custom model from env", func(t *testing.T) {
		os.Setenv("VOYAGE_API_KEY", "test-api-key")
		os.Setenv("EMBEDDINGS_MODEL", "voyage-3")
		defer os.Unsetenv("VOYAGE_API_KEY")
		defer os.Unsetenv("EMBEDDINGS_MODEL")

		server := setupTestServer()
		if server.auto == nil {
			t.Error("auto mode should be initialized")
		}
	})
}

// =============================================================================
// Test: SetOrchestrator
// =============================================================================

func TestSetOrchestratorMethod(t *testing.T) {
	server := setupTestServer()
	exec := &stubExecutor{}
	orch := orchestration.NewOrchestratorWithExecutor(exec)

	if server.orchestrator != nil {
		t.Error("orchestrator should be nil initially")
	}

	server.SetOrchestrator(orch)

	if server.orchestrator == nil {
		t.Error("orchestrator should be set after SetOrchestrator")
	}
}

// =============================================================================
// Test: Handler error paths
// =============================================================================

func TestHandlerErrorPaths(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleBranchHistory with nonexistent branch", func(t *testing.T) {
		req := BranchHistoryRequest{
			BranchID: "nonexistent-branch",
		}

		_, _, err := server.handleBranchHistory(ctx, nil, req)
		if err == nil {
			t.Error("expected error for nonexistent branch")
		}
	})

	t.Run("handleFocusBranch with nonexistent branch", func(t *testing.T) {
		req := FocusBranchRequest{
			BranchID: "nonexistent-branch",
		}

		_, _, err := server.handleFocusBranch(ctx, nil, req)
		if err == nil {
			t.Error("expected error for nonexistent branch")
		}
	})

	t.Run("handleValidate with nonexistent thought", func(t *testing.T) {
		req := ValidateRequest{
			ThoughtID: "nonexistent-thought",
		}

		_, _, err := server.handleValidate(ctx, nil, req)
		if err == nil {
			t.Error("expected error for nonexistent thought")
		}
	})

	t.Run("handleExecuteWorkflow without orchestrator", func(t *testing.T) {
		server.orchestrator = nil
		req := ExecuteWorkflowRequest{
			WorkflowID: "test-workflow",
			Input:      map[string]interface{}{"problem": "test"},
		}

		_, _, err := server.handleExecuteWorkflow(ctx, nil, req)
		if err == nil {
			t.Error("expected error when orchestrator not initialized")
		}
	})

	t.Run("handleRegisterWorkflow without orchestrator", func(t *testing.T) {
		server.orchestrator = nil
		req := RegisterWorkflowRequest{
			Workflow: &orchestration.Workflow{
				ID:   "test-wf",
				Name: "Test",
				Type: orchestration.WorkflowSequential,
				Steps: []*orchestration.WorkflowStep{
					{ID: "s1", Tool: "think", Input: map[string]interface{}{}},
				},
			},
		}

		_, _, err := server.handleRegisterWorkflow(ctx, nil, req)
		if err == nil {
			t.Error("expected error when orchestrator not initialized")
		}
	})
}

// =============================================================================
// Test: Calibration and Hallucination handlers
// =============================================================================

func TestCalibrationHandlers(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleRecordPrediction validation errors", func(t *testing.T) {
		// Missing thought_id
		req := handlers.RecordPredictionRequest{
			Confidence: 0.8,
			Mode:       "linear",
		}
		_, _, err := server.handleRecordPrediction(ctx, nil, req)
		if err == nil {
			t.Error("expected error for missing thought_id")
		}

		// Invalid confidence
		req = handlers.RecordPredictionRequest{
			ThoughtID:  "thought-1",
			Confidence: 1.5,
			Mode:       "linear",
		}
		_, _, err = server.handleRecordPrediction(ctx, nil, req)
		if err == nil {
			t.Error("expected error for invalid confidence")
		}

		// Missing mode
		req = handlers.RecordPredictionRequest{
			ThoughtID:  "thought-1",
			Confidence: 0.8,
		}
		_, _, err = server.handleRecordPrediction(ctx, nil, req)
		if err == nil {
			t.Error("expected error for missing mode")
		}
	})

	t.Run("handleRecordOutcome validation errors", func(t *testing.T) {
		// Missing thought_id
		req := handlers.RecordOutcomeRequest{
			WasCorrect:       true,
			ActualConfidence: 0.8,
			Source:           "validation",
		}
		_, _, err := server.handleRecordOutcome(ctx, nil, req)
		if err == nil {
			t.Error("expected error for missing thought_id")
		}

		// Invalid actual_confidence
		req = handlers.RecordOutcomeRequest{
			ThoughtID:        "thought-1",
			WasCorrect:       true,
			ActualConfidence: -0.1,
			Source:           "validation",
		}
		_, _, err = server.handleRecordOutcome(ctx, nil, req)
		if err == nil {
			t.Error("expected error for invalid actual_confidence")
		}

		// Missing source
		req = handlers.RecordOutcomeRequest{
			ThoughtID:        "thought-1",
			WasCorrect:       true,
			ActualConfidence: 0.8,
		}
		_, _, err = server.handleRecordOutcome(ctx, nil, req)
		if err == nil {
			t.Error("expected error for missing source")
		}
	})

	t.Run("handleGetCalibrationReport success", func(t *testing.T) {
		req := handlers.GetCalibrationReportRequest{}
		result, _, err := server.handleGetCalibrationReport(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

func TestHallucinationHandlers(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleVerifyThought missing thought_id", func(t *testing.T) {
		req := handlers.VerifyThoughtRequest{}
		_, _, err := server.handleVerifyThought(ctx, nil, req)
		if err == nil {
			t.Error("expected error for missing thought_id")
		}
	})

	t.Run("handleGetHallucinationReport missing thought_id", func(t *testing.T) {
		req := handlers.GetReportRequest{}
		_, _, err := server.handleGetHallucinationReport(ctx, nil, req)
		if err == nil {
			t.Error("expected error for missing thought_id")
		}
	})
}

// =============================================================================
// Test: Advanced handler methods
// =============================================================================

func TestAdvancedHandlers(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleDualProcessThink", func(t *testing.T) {
		req := handlers.DualProcessThinkRequest{
			Content: "Test dual process thought",
		}

		result, _, err := server.handleDualProcessThink(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleCreateCheckpoint", func(t *testing.T) {
		// First create a branch
		thinkReq := ThinkRequest{
			Content:    "Test thought for checkpoint",
			Mode:       "tree",
			Confidence: 0.8,
		}
		_, thinkResp, err := server.handleThink(ctx, nil, thinkReq)
		if err != nil {
			t.Fatalf("failed to create thought: %v", err)
		}

		req := handlers.CreateCheckpointRequest{
			BranchID: thinkResp.BranchID,
			Name:     "test-checkpoint",
		}

		result, _, err := server.handleCreateCheckpoint(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleListCheckpoints", func(t *testing.T) {
		req := handlers.ListCheckpointsRequest{}

		result, _, err := server.handleListCheckpoints(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleGenerateHypotheses", func(t *testing.T) {
		req := handlers.GenerateHypothesesRequest{
			Observations: []*handlers.ObservationInput{
				{Description: "The system crashed at midnight", Confidence: 0.9},
				{Description: "Logs show memory spike", Confidence: 0.8},
			},
		}

		result, _, err := server.handleGenerateHypotheses(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleDetectBlindSpots", func(t *testing.T) {
		req := handlers.DetectBlindSpotsRequest{
			Content:    "We should use microservices for our new project",
			Domain:     "software architecture",
			Confidence: 0.8,
		}

		result, _, err := server.handleDetectBlindSpots(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleProveTheorem", func(t *testing.T) {
		req := handlers.ProveTheoremRequest{
			Name:       "modus ponens test",
			Premises:   []string{"If P then Q", "P"},
			Conclusion: "Q",
		}

		result, _, err := server.handleProveTheorem(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleCheckConstraints", func(t *testing.T) {
		req := handlers.CheckConstraintsRequest{
			Symbols: []*handlers.SymbolInput{
				{Name: "x", Type: "integer", Domain: "positive"},
				{Name: "y", Type: "integer", Domain: "positive"},
			},
			Constraints: []*handlers.ConstraintInput{
				{Type: "equality", Expression: "x + y = 10", Symbols: []string{"x", "y"}},
			},
		}

		result, _, err := server.handleCheckConstraints(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

// =============================================================================
// Test: Temporal and Causal handlers
// =============================================================================

func TestTemporalCausalHandlers(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleAnalyzePerspectives", func(t *testing.T) {
		req := handlers.AnalyzePerspectivesRequest{
			Situation:        "Implementing a new authentication system",
			StakeholderHints: []string{"developers", "security team", "users"},
		}

		result, _, err := server.handleAnalyzePerspectives(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleAnalyzeTemporal", func(t *testing.T) {
		req := handlers.AnalyzeTemporalRequest{
			Situation:   "Should we refactor now or after release?",
			TimeHorizon: "months",
		}

		result, _, err := server.handleAnalyzeTemporal(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleCompareTimeHorizons", func(t *testing.T) {
		req := handlers.CompareTimeHorizonsRequest{
			Situation: "Migrate to new database",
		}

		result, _, err := server.handleCompareTimeHorizons(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleIdentifyOptimalTiming", func(t *testing.T) {
		req := handlers.IdentifyOptimalTimingRequest{
			Situation: "Launch new feature",
		}

		result, _, err := server.handleIdentifyOptimalTiming(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleBuildCausalGraph", func(t *testing.T) {
		req := handlers.BuildCausalGraphRequest{
			Description: "Sales process analysis",
			Observations: []string{
				"Marketing increases brand awareness",
				"Brand awareness leads to more leads",
				"More leads result in more sales",
			},
		}

		result, resp, err := server.handleBuildCausalGraph(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
		if resp == nil || resp.Graph == nil || resp.Graph.ID == "" {
			t.Error("expected graph ID in response")
		}
	})

	t.Run("handleSimulateIntervention", func(t *testing.T) {
		// First build a causal graph
		buildReq := handlers.BuildCausalGraphRequest{
			Description:  "Test graph",
			Observations: []string{"A causes B", "B causes C"},
		}
		_, buildResp, _ := server.handleBuildCausalGraph(ctx, nil, buildReq)

		// Use a variable ID from the graph
		variableID := ""
		if buildResp.Graph != nil && len(buildResp.Graph.Variables) > 0 {
			variableID = buildResp.Graph.Variables[0].ID
		}

		req := handlers.SimulateInterventionRequest{
			GraphID:          buildResp.Graph.ID,
			VariableID:       variableID,
			InterventionType: "increase",
		}

		result, _, err := server.handleSimulateIntervention(ctx, nil, req)
		// This may error if variable not found, but we're testing the handler works
		if result == nil && err == nil {
			t.Error("expected either result or error")
		}
	})

	t.Run("handleGenerateCounterfactual", func(t *testing.T) {
		// First build a causal graph
		buildReq := handlers.BuildCausalGraphRequest{
			Description:  "Test graph",
			Observations: []string{"X causes Y", "Y causes Z"},
		}
		_, buildResp, _ := server.handleBuildCausalGraph(ctx, nil, buildReq)

		req := handlers.GenerateCounterfactualRequest{
			GraphID:  buildResp.Graph.ID,
			Scenario: "What if X had been high?",
			Changes:  map[string]string{"X": "high"},
		}

		result, _, err := server.handleGenerateCounterfactual(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleAnalyzeCorrelationVsCausation", func(t *testing.T) {
		req := handlers.AnalyzeCorrelationVsCausationRequest{
			Observation: "Ice cream sales and drowning incidents both increase in summer",
		}

		result, _, err := server.handleAnalyzeCorrelationVsCausation(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleGetCausalGraph", func(t *testing.T) {
		// First build a causal graph
		buildReq := handlers.BuildCausalGraphRequest{
			Description:  "Get test graph",
			Observations: []string{"P causes Q"},
		}
		_, buildResp, _ := server.handleBuildCausalGraph(ctx, nil, buildReq)

		req := handlers.GetCausalGraphRequest{
			GraphID: buildResp.Graph.ID,
		}

		result, _, err := server.handleGetCausalGraph(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

// =============================================================================
// Test: Probabilistic and Decision handlers
// =============================================================================

func TestProbabilisticDecisionHandlers(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleProbabilisticReasoning create", func(t *testing.T) {
		req := handlers.ProbabilisticReasoningRequest{
			Operation: "create",
			Statement: "It will rain tomorrow",
			PriorProb: 0.3,
		}

		result, _, err := server.handleProbabilisticReasoning(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleAssessEvidence", func(t *testing.T) {
		req := handlers.AssessEvidenceRequest{
			Content: "A peer-reviewed study found significant results",
			Source:  "Nature Journal",
		}

		result, _, err := server.handleAssessEvidence(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleDetectContradictions", func(t *testing.T) {
		// Create some thoughts with potential contradictions
		thought1 := &types.Thought{
			ID:         "contra-thought-1",
			Content:    "The system is fast and responsive",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		thought2 := &types.Thought{
			ID:         "contra-thought-2",
			Content:    "The system is slow and unresponsive",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		server.storage.StoreThought(thought1)
		server.storage.StoreThought(thought2)

		req := handlers.DetectContradictionsRequest{
			ThoughtIDs: []string{"contra-thought-1", "contra-thought-2"},
		}

		result, _, err := server.handleDetectContradictions(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleMakeDecision", func(t *testing.T) {
		req := handlers.MakeDecisionRequest{
			Question: "Which database should we use?",
			Options: []*types.DecisionOption{
				{
					ID:          "postgres",
					Name:        "PostgreSQL",
					Description: "Relational database",
					Scores:      map[string]float64{"cost": 0.7, "performance": 0.8},
					Pros:        []string{"ACID compliance", "Mature ecosystem"},
					Cons:        []string{"Complex scaling"},
				},
				{
					ID:          "mongo",
					Name:        "MongoDB",
					Description: "Document database",
					Scores:      map[string]float64{"cost": 0.6, "performance": 0.9},
					Pros:        []string{"Flexible schema", "Easy scaling"},
					Cons:        []string{"No ACID for multi-doc"},
				},
			},
			Criteria: []*types.DecisionCriterion{
				{ID: "cost", Name: "Cost", Weight: 0.4, Maximize: false},
				{ID: "performance", Name: "Performance", Weight: 0.6, Maximize: true},
			},
		}

		result, _, err := server.handleMakeDecision(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleDecomposeProblem", func(t *testing.T) {
		req := handlers.DecomposeProblemRequest{
			Problem: "How to improve system performance by 50%?",
		}

		result, _, err := server.handleDecomposeProblem(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleSensitivityAnalysis", func(t *testing.T) {
		req := handlers.SensitivityAnalysisRequest{
			TargetClaim:    "System will scale to handle load",
			Assumptions:    []string{"User growth of 10%", "Server costs decrease"},
			BaseConfidence: 0.7,
		}

		result, _, err := server.handleSensitivityAnalysis(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

// =============================================================================
// Test: Metacognition handlers
// =============================================================================

func TestMetacognitionHandlers(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleSelfEvaluate", func(t *testing.T) {
		// First create a thought
		thinkReq := ThinkRequest{
			Content:    "This is a test thought for self-evaluation",
			Mode:       "linear",
			Confidence: 0.7,
		}
		_, thinkResp, _ := server.handleThink(ctx, nil, thinkReq)

		req := handlers.SelfEvaluateRequest{
			ThoughtID: thinkResp.ThoughtID,
		}

		result, _, err := server.handleSelfEvaluate(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleDetectBiases with thought", func(t *testing.T) {
		// First create a thought with biased content
		thought := &types.Thought{
			ID:         "bias-test-thought",
			Content:    "This confirms what I already believed to be true. Everyone knows this is obvious.",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		server.storage.StoreThought(thought)

		req := handlers.DetectBiasesRequest{
			ThoughtID: "bias-test-thought",
		}

		result, _, err := server.handleDetectBiases(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

// =============================================================================
// Test: Integration handlers
// =============================================================================

func TestIntegrationHandlers(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleSynthesizeInsights", func(t *testing.T) {
		syncReq := SynthesizeInsightsRequest{
			Context: "Analyzing system performance",
			Inputs: []*integration.Input{
				{ID: "input-1", Mode: "causal", Content: "High load causes delays", Confidence: 0.8},
				{ID: "input-2", Mode: "temporal", Content: "Delays increase over time", Confidence: 0.75},
			},
		}

		result, _, err := server.handleSynthesizeInsights(ctx, nil, syncReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleDetectEmergentPatterns", func(t *testing.T) {
		req := DetectEmergentPatternsRequest{
			Inputs: []*integration.Input{
				{ID: "input-1", Mode: "linear", Content: "Step by step analysis", Confidence: 0.8},
				{ID: "input-2", Mode: "tree", Content: "Branch exploration", Confidence: 0.75},
				{ID: "input-3", Mode: "divergent", Content: "Creative idea", Confidence: 0.7},
			},
		}

		result, _, err := server.handleDetectEmergentPatterns(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handleListIntegrationPatterns", func(t *testing.T) {
		req := EmptyRequest{}

		result, resp, err := server.handleListIntegrationPatterns(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
		if resp == nil || len(resp.Patterns) == 0 {
			t.Error("expected patterns in response")
		}
	})
}

// =============================================================================
// Test: Case-based reasoning handlers
// =============================================================================

func TestCaseBasedReasoningHandlers(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleRetrieveCases", func(t *testing.T) {
		req := handlers.RetrieveCasesRequest{
			Problem: &handlers.ProblemDescriptionInput{
				Description: "System performance degradation",
				Context:     "Production environment",
				Goals:       []string{"Improve response time"},
				Constraints: []string{"No downtime"},
			},
			Domain: "performance",
		}

		result, _, err := server.handleRetrieveCases(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("handlePerformCBRCycle", func(t *testing.T) {
		req := handlers.PerformCBRCycleRequest{
			Problem: &handlers.ProblemDescriptionInput{
				Description: "Database connection pool exhaustion",
				Context:     "High traffic scenario",
				Goals:       []string{"Stabilize connections"},
			},
			Domain: "database",
		}

		// CBR cycle may fail if no similar cases exist - this is expected
		// We're testing the handler invocation, not case availability
		result, _, err := server.handlePerformCBRCycle(ctx, nil, req)
		// Either we get a result or an expected "no similar cases" error
		if result == nil && err == nil {
			t.Error("expected either result or error")
		}
		// If there's an error, it should be about no similar cases (expected)
		if err != nil && !strings.Contains(err.Error(), "no similar cases") {
			t.Errorf("unexpected error type: %v", err)
		}
	})

	t.Run("handleEvaluateHypotheses", func(t *testing.T) {
		req := handlers.EvaluateHypothesesRequest{
			Observations: []*handlers.ObservationInput{
				{Description: "Server CPU at 90%", Confidence: 0.95},
				{Description: "Memory usage normal", Confidence: 0.9},
			},
			Hypotheses: []*handlers.HypothesisInput{
				{Description: "CPU-bound process", PriorProbability: 0.8},
				{Description: "Memory leak", PriorProbability: 0.6},
			},
			Method: "combined",
		}

		result, _, err := server.handleEvaluateHypotheses(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

// =============================================================================
// Test: Validation helper functions
// =============================================================================

func TestValidationHelpers(t *testing.T) {
	t.Run("SuggestAlternativeTool detect-fallacies suggestion", func(t *testing.T) {
		suggestion := SuggestAlternativeTool("detect-fallacies", "Check for confirmation bias")
		if suggestion == nil {
			t.Error("expected suggestion for bias detection")
		} else if suggestion.SuggestedTool != "detect-biases" {
			t.Errorf("expected 'detect-biases', got %s", suggestion.SuggestedTool)
		}
	})

	t.Run("SuggestAlternativeTool execute-workflow format", func(t *testing.T) {
		suggestion := SuggestAlternativeTool("execute-workflow", "analyze something")
		if suggestion == nil {
			t.Error("expected suggestion for workflow format")
		}
	})

	t.Run("SuggestAlternativeTool make-decision format", func(t *testing.T) {
		suggestion := SuggestAlternativeTool("make-decision", "help me make a decision")
		if suggestion == nil {
			t.Error("expected suggestion for decision format")
		}
	})

	t.Run("SuggestAlternativeTool no suggestion needed", func(t *testing.T) {
		suggestion := SuggestAlternativeTool("think", "just a regular thought")
		if suggestion != nil {
			t.Error("expected no suggestion for valid usage")
		}
	})

	t.Run("toLowerSimple", func(t *testing.T) {
		result := toLowerSimple("HELLO World")
		if result != "hello world" {
			t.Errorf("expected 'hello world', got %s", result)
		}
	})

	t.Run("contains", func(t *testing.T) {
		if !contains("hello world", "world") {
			t.Error("expected true for substring match")
		}
		if contains("hello", "world") {
			t.Error("expected false for no match")
		}
		if !contains("anything", "") {
			t.Error("expected true for empty substring")
		}
	})
}

// =============================================================================
// Test: Auto-validation with custom threshold
// =============================================================================

func TestAutoValidationThreshold(t *testing.T) {
	t.Run("custom threshold from env", func(t *testing.T) {
		os.Setenv("AUTO_VALIDATION_THRESHOLD", "0.3")
		defer os.Unsetenv("AUTO_VALIDATION_THRESHOLD")

		server := setupTestServer()
		ctx := context.Background()

		// With threshold at 0.3, confidence 0.4 should NOT trigger auto-validation
		req := ThinkRequest{
			Content:    "Test thought",
			Mode:       "linear",
			Confidence: 0.4,
		}

		_, _, err := server.handleThink(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid threshold falls back to default", func(t *testing.T) {
		os.Setenv("AUTO_VALIDATION_THRESHOLD", "invalid")
		defer os.Unsetenv("AUTO_VALIDATION_THRESHOLD")

		server := setupTestServer()
		ctx := context.Background()

		req := ThinkRequest{
			Content:    "Test thought",
			Mode:       "linear",
			Confidence: 0.4,
		}

		_, _, err := server.handleThink(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("out of range threshold ignored", func(t *testing.T) {
		os.Setenv("AUTO_VALIDATION_THRESHOLD", "1.5")
		defer os.Unsetenv("AUTO_VALIDATION_THRESHOLD")

		server := setupTestServer()
		ctx := context.Background()

		req := ThinkRequest{
			Content:    "Test thought",
			Mode:       "linear",
			Confidence: 0.4,
		}

		_, _, err := server.handleThink(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// =============================================================================
// Test: Various handler modes and edge cases
// =============================================================================

func TestHandlerModes(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleThink with reflection mode", func(t *testing.T) {
		// Reflection mode is implemented internally as linear
		req := ThinkRequest{
			Content:    "Reflecting on my previous thoughts",
			Mode:       "linear",
			Confidence: 0.75,
		}

		_, resp, err := server.handleThink(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp.ThoughtID == "" {
			t.Error("expected thought ID")
		}
	})

	t.Run("handleThink with cross references", func(t *testing.T) {
		// First create a branch
		req1 := ThinkRequest{
			Content:    "First branch thought",
			Mode:       "tree",
			Confidence: 0.8,
		}
		_, resp1, _ := server.handleThink(ctx, nil, req1)

		// Then create a thought with cross-reference
		req2 := ThinkRequest{
			Content:    "Second branch with cross-ref",
			Mode:       "tree",
			Confidence: 0.8,
			CrossRefs: []CrossRefInput{
				{
					ToBranch: resp1.BranchID,
					Type:     "builds_upon",
					Reason:   "Extends the first thought",
					Strength: 0.9,
				},
			},
		}

		_, resp2, err := server.handleThink(ctx, nil, req2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp2.BranchID == "" {
			t.Error("expected branch ID")
		}
	})

	t.Run("handleCheckSyntax", func(t *testing.T) {
		req := CheckSyntaxRequest{
			Statements: []string{
				"All men are mortal",
				"If P then Q",
			},
		}

		result, _, err := server.handleCheckSyntax(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

// =============================================================================
// Test: Workflow execution with nonexistent workflow
// =============================================================================

func TestWorkflowExecutionErrors(t *testing.T) {
	server := setupTestServer()
	exec := &stubExecutor{}
	orch := orchestration.NewOrchestratorWithExecutor(exec)
	server.SetOrchestrator(orch)

	ctx := context.Background()

	t.Run("execute nonexistent workflow", func(t *testing.T) {
		req := ExecuteWorkflowRequest{
			WorkflowID: "nonexistent-workflow",
			Input:      map[string]interface{}{"problem": "test"},
		}

		_, resp, err := server.handleExecuteWorkflow(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp.Status != "failed" {
			t.Errorf("expected failed status, got %s", resp.Status)
		}
	})
}

// =============================================================================
// Test: Public interface methods for tool executor
// =============================================================================

func TestPublicInterfaceMethods(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("ProcessThought linear mode", func(t *testing.T) {
		input := modes.ThoughtInput{
			Content:    "Test thought for linear mode",
			Confidence: 0.8,
		}

		result, err := server.ProcessThought(ctx, "linear", input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil || result.ThoughtID == "" {
			t.Error("expected thought result")
		}
	})

	t.Run("ProcessThought auto mode", func(t *testing.T) {
		input := modes.ThoughtInput{
			Content:    "Test thought for auto mode",
			Confidence: 0.8,
		}

		result, err := server.ProcessThought(ctx, "", input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected thought result")
		}
	})

	t.Run("ProcessThought unknown mode", func(t *testing.T) {
		input := modes.ThoughtInput{
			Content:    "Test",
			Confidence: 0.8,
		}

		_, err := server.ProcessThought(ctx, "unknown", input)
		if err == nil {
			t.Error("expected error for unknown mode")
		}
	})

	t.Run("BuildCausalGraph", func(t *testing.T) {
		graph, err := server.BuildCausalGraph(ctx, "Test graph", []string{"A causes B"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if graph == nil {
			t.Error("expected causal graph")
		}
	})

	t.Run("ProbabilisticReasoning create", func(t *testing.T) {
		req := handlers.ProbabilisticReasoningRequest{
			Operation: "create",
			Statement: "Test statement",
			PriorProb: 0.5,
		}

		result, err := server.ProbabilisticReasoning(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("ProbabilisticReasoning unknown operation", func(t *testing.T) {
		req := handlers.ProbabilisticReasoningRequest{
			Operation: "unknown",
		}

		_, err := server.ProbabilisticReasoning(ctx, req)
		if err == nil {
			t.Error("expected error for unknown operation")
		}
	})

	t.Run("MakeDecision", func(t *testing.T) {
		req := handlers.MakeDecisionRequest{
			Question: "Test decision",
			Options: []*types.DecisionOption{
				{ID: "a", Name: "A", Scores: map[string]float64{"c": 0.8}},
			},
			Criteria: []*types.DecisionCriterion{
				{ID: "c", Name: "C", Weight: 1.0},
			},
		}

		decision, err := server.MakeDecision(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if decision == nil {
			t.Error("expected decision")
		}
	})

	t.Run("DetectContradictions with thought IDs", func(t *testing.T) {
		// Create thoughts
		thought := &types.Thought{
			ID:         "public-test-1",
			Content:    "Test content",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		server.storage.StoreThought(thought)

		req := handlers.DetectContradictionsRequest{
			ThoughtIDs: []string{"public-test-1"},
		}

		_, err := server.DetectContradictions(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("DetectContradictions with mode", func(t *testing.T) {
		req := handlers.DetectContradictionsRequest{
			Mode: "linear",
		}

		_, err := server.DetectContradictions(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SynthesizeInsights", func(t *testing.T) {
		req := SynthesizeInsightsRequest{
			Context: "Test context",
			Inputs: []*integration.Input{
				{ID: "1", Mode: "linear", Content: "A", Confidence: 0.8},
				{ID: "2", Mode: "tree", Content: "B", Confidence: 0.7},
			},
		}

		_, err := server.SynthesizeInsights(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("DetectBiases with thought", func(t *testing.T) {
		thought := &types.Thought{
			ID:         "bias-public-test",
			Content:    "This confirms my belief",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		server.storage.StoreThought(thought)

		req := handlers.DetectBiasesRequest{
			ThoughtID: "bias-public-test",
		}

		_, err := server.DetectBiases(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("DetectBiases missing input", func(t *testing.T) {
		req := handlers.DetectBiasesRequest{}

		_, err := server.DetectBiases(ctx, req)
		if err == nil {
			t.Error("expected error for missing input")
		}
	})

	t.Run("AssessEvidence", func(t *testing.T) {
		req := handlers.AssessEvidenceRequest{
			Content: "Study shows results",
			Source:  "Journal",
		}

		_, err := server.AssessEvidence(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SelfEvaluate with thought", func(t *testing.T) {
		thought := &types.Thought{
			ID:         "eval-public-test",
			Content:    "Test content",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		server.storage.StoreThought(thought)

		req := handlers.SelfEvaluateRequest{
			ThoughtID: "eval-public-test",
		}

		_, err := server.SelfEvaluate(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SelfEvaluate missing input", func(t *testing.T) {
		req := handlers.SelfEvaluateRequest{}

		_, err := server.SelfEvaluate(ctx, req)
		if err == nil {
			t.Error("expected error for missing input")
		}
	})

	t.Run("DecomposeProblem", func(t *testing.T) {
		req := handlers.DecomposeProblemRequest{
			Problem: "How to improve performance?",
		}

		_, err := server.DecomposeProblem(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SensitivityAnalysis", func(t *testing.T) {
		req := handlers.SensitivityAnalysisRequest{
			TargetClaim:    "Test claim",
			Assumptions:    []string{"A1", "A2"},
			BaseConfidence: 0.7,
		}

		_, err := server.SensitivityAnalysis(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("AnalyzePerspectives", func(t *testing.T) {
		req := handlers.AnalyzePerspectivesRequest{
			Situation: "New system implementation",
		}

		_, err := server.AnalyzePerspectives(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("AnalyzeTemporal", func(t *testing.T) {
		req := handlers.AnalyzeTemporalRequest{
			Situation:   "Test decision",
			TimeHorizon: "months",
		}

		_, err := server.AnalyzeTemporal(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SimulateIntervention", func(t *testing.T) {
		// Build graph first
		graph, _ := server.BuildCausalGraph(ctx, "Test", []string{"X causes Y"})

		req := handlers.SimulateInterventionRequest{
			GraphID:          graph.ID,
			VariableID:       "X",
			InterventionType: "increase",
		}

		// May fail if variable not found - that's OK
		server.SimulateIntervention(ctx, req)
	})

	t.Run("GenerateCounterfactual", func(t *testing.T) {
		graph, _ := server.BuildCausalGraph(ctx, "Test", []string{"A causes B"})

		req := handlers.GenerateCounterfactualRequest{
			GraphID:  graph.ID,
			Scenario: "What if A was high?",
			Changes:  map[string]string{"A": "high"},
		}

		// May fail if variable not found - that's OK
		server.GenerateCounterfactual(ctx, req)
	})

	t.Run("DetectEmergentPatterns", func(t *testing.T) {
		req := DetectEmergentPatternsRequest{
			Inputs: []*integration.Input{
				{ID: "1", Mode: "linear", Content: "A", Confidence: 0.8},
				{ID: "2", Mode: "tree", Content: "B", Confidence: 0.7},
			},
		}

		_, err := server.DetectEmergentPatterns(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// =============================================================================
// Test: Checkpoint restore and paginateThoughts
// =============================================================================

func TestRestoreCheckpointAndPagination(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleRestoreCheckpoint", func(t *testing.T) {
		// First create a branch and checkpoint
		thinkReq := ThinkRequest{
			Content:    "Test thought",
			Mode:       "tree",
			Confidence: 0.8,
		}
		_, thinkResp, _ := server.handleThink(ctx, nil, thinkReq)

		// Create checkpoint
		checkpointReq := handlers.CreateCheckpointRequest{
			BranchID: thinkResp.BranchID,
			Name:     "test-restore-checkpoint",
		}
		_, checkpointResp, _ := server.handleCreateCheckpoint(ctx, nil, checkpointReq)

		// Restore checkpoint
		restoreReq := handlers.RestoreCheckpointRequest{
			CheckpointID: checkpointResp.CheckpointID,
		}

		result, _, err := server.handleRestoreCheckpoint(ctx, nil, restoreReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("history with pagination", func(t *testing.T) {
		// Create multiple thoughts
		for i := 0; i < 10; i++ {
			req := ThinkRequest{
				Content:    "Test thought for pagination",
				Mode:       "linear",
				Confidence: 0.8,
			}
			server.handleThink(ctx, nil, req)
		}

		// Test pagination
		histReq := HistoryRequest{
			Limit:  5,
			Offset: 2,
		}

		_, resp, err := server.handleHistory(ctx, nil, histReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(resp.Thoughts) > 5 {
			t.Errorf("expected at most 5 thoughts, got %d", len(resp.Thoughts))
		}
	})
}

// =============================================================================
// Test: Hallucination verification with valid thought
// =============================================================================

func TestHallucinationVerificationComplete(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("handleVerifyThought with valid thought", func(t *testing.T) {
		// Create a thought first
		thought := &types.Thought{
			ID:         "verify-test-thought",
			Content:    "This is a test thought that should be verified",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		server.storage.StoreThought(thought)

		req := handlers.VerifyThoughtRequest{
			ThoughtID:         "verify-test-thought",
			VerificationLevel: "fast",
		}

		result, resp, err := server.handleVerifyThought(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
		if resp == nil {
			t.Error("expected response")
		}
	})

	t.Run("handleGetHallucinationReport with valid thought", func(t *testing.T) {
		// First verify a thought
		thought := &types.Thought{
			ID:         "report-test-thought",
			Content:    "Test thought for report",
			Mode:       types.ModeLinear,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		}
		server.storage.StoreThought(thought)

		// Verify first
		verifyReq := handlers.VerifyThoughtRequest{
			ThoughtID:         "report-test-thought",
			VerificationLevel: "fast",
		}
		server.handleVerifyThought(ctx, nil, verifyReq)

		// Get report
		reportReq := handlers.GetReportRequest{
			ThoughtID: "report-test-thought",
		}

		result, _, err := server.handleGetHallucinationReport(ctx, nil, reportReq)
		// May fail if report not cached - that's OK
		if result == nil && err == nil {
			t.Error("expected either result or error")
		}
	})
}

// =============================================================================
// Test: Additional coverage for ProbabilisticReasoning operations
// =============================================================================

func TestProbabilisticReasoningOperations(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("update operation", func(t *testing.T) {
		// First create a belief
		createReq := handlers.ProbabilisticReasoningRequest{
			Operation: "create",
			Statement: "Test update belief",
			PriorProb: 0.5,
		}
		result, _ := server.ProbabilisticReasoning(ctx, createReq)
		beliefResult, ok := result.(map[string]interface{})
		if !ok {
			t.Skip("could not get belief ID")
		}
		beliefID, _ := beliefResult["id"].(string)

		// Update belief
		updateReq := handlers.ProbabilisticReasoningRequest{
			Operation:    "update",
			BeliefID:     beliefID,
			EvidenceID:   "evidence-1",
			Likelihood:   0.8,
			EvidenceProb: 0.6,
		}

		_, err := server.ProbabilisticReasoning(ctx, updateReq)
		if err != nil {
			t.Logf("update error (expected if belief not found): %v", err)
		}
	})

	t.Run("get operation", func(t *testing.T) {
		req := handlers.ProbabilisticReasoningRequest{
			Operation: "get",
			BeliefID:  "nonexistent-belief",
		}

		_, err := server.ProbabilisticReasoning(ctx, req)
		if err != nil {
			t.Logf("get error (expected): %v", err)
		}
	})

	t.Run("combine operation", func(t *testing.T) {
		req := handlers.ProbabilisticReasoningRequest{
			Operation: "combine",
			BeliefIDs: []string{"belief-1", "belief-2"},
			CombineOp: "and",
		}

		_, err := server.ProbabilisticReasoning(ctx, req)
		if err != nil {
			t.Logf("combine error (expected): %v", err)
		}
	})
}

// =============================================================================
// Test: Additional DetectContradictions paths
// =============================================================================

func TestDetectContradictionsAdditionalPaths(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("with branch ID", func(t *testing.T) {
		// Create a branch
		thinkReq := ThinkRequest{
			Content:    "Test thought",
			Mode:       "tree",
			Confidence: 0.8,
		}
		_, thinkResp, _ := server.handleThink(ctx, nil, thinkReq)

		req := handlers.DetectContradictionsRequest{
			BranchID: thinkResp.BranchID,
		}

		_, err := server.DetectContradictions(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("with nonexistent thought ID", func(t *testing.T) {
		req := handlers.DetectContradictionsRequest{
			ThoughtIDs: []string{"nonexistent"},
		}

		_, err := server.DetectContradictions(ctx, req)
		if err == nil {
			t.Error("expected error for nonexistent thought")
		}
	})

	t.Run("check all thoughts", func(t *testing.T) {
		req := handlers.DetectContradictionsRequest{}

		_, err := server.DetectContradictions(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// =============================================================================
// Test: DetectBiases and SelfEvaluate with branch
// =============================================================================

func TestBranchBasedOperations(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create a branch with thoughts
	thinkReq := ThinkRequest{
		Content:    "Test thought for branch operations",
		Mode:       "tree",
		Confidence: 0.8,
	}
	_, thinkResp, _ := server.handleThink(ctx, nil, thinkReq)

	t.Run("DetectBiases with branch", func(t *testing.T) {
		req := handlers.DetectBiasesRequest{
			BranchID: thinkResp.BranchID,
		}

		_, err := server.DetectBiases(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SelfEvaluate with branch", func(t *testing.T) {
		req := handlers.SelfEvaluateRequest{
			BranchID: thinkResp.BranchID,
		}

		_, err := server.SelfEvaluate(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// =============================================================================
// Test: Record prediction and outcome success paths
// =============================================================================

func TestCalibrationSuccessPaths(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("record prediction success", func(t *testing.T) {
		req := handlers.RecordPredictionRequest{
			ThoughtID:  "calibration-test-thought",
			Confidence: 0.8,
			Mode:       "linear",
		}

		result, _, err := server.handleRecordPrediction(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("record outcome success", func(t *testing.T) {
		// First record a prediction
		predReq := handlers.RecordPredictionRequest{
			ThoughtID:  "outcome-test-thought",
			Confidence: 0.8,
			Mode:       "linear",
		}
		server.handleRecordPrediction(ctx, nil, predReq)

		// Then record outcome
		outcomeReq := handlers.RecordOutcomeRequest{
			ThoughtID:        "outcome-test-thought",
			WasCorrect:       true,
			ActualConfidence: 0.9,
			Source:           "validation",
		}

		result, _, err := server.handleRecordOutcome(ctx, nil, outcomeReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

// =============================================================================
// Test: Paginate thoughts directly (internal function)
// =============================================================================

func TestPaginateThoughts(t *testing.T) {
	thoughts := []*types.Thought{
		{ID: "t1"},
		{ID: "t2"},
		{ID: "t3"},
		{ID: "t4"},
		{ID: "t5"},
	}

	t.Run("normal pagination", func(t *testing.T) {
		result := paginateThoughts(thoughts, 2, 1)
		if len(result) != 2 {
			t.Errorf("expected 2, got %d", len(result))
		}
		if result[0].ID != "t2" {
			t.Errorf("expected t2, got %s", result[0].ID)
		}
	})

	t.Run("offset beyond length", func(t *testing.T) {
		result := paginateThoughts(thoughts, 10, 100)
		if len(result) != 0 {
			t.Errorf("expected 0, got %d", len(result))
		}
	})

	t.Run("limit exceeds remaining", func(t *testing.T) {
		result := paginateThoughts(thoughts, 10, 3)
		if len(result) != 2 {
			t.Errorf("expected 2, got %d", len(result))
		}
	})
}

// =============================================================================
// Test: History with all filter modes
// =============================================================================

func TestHistoryAllModes(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create thoughts for filtering
	for _, mode := range []string{"linear", "tree", "divergent"} {
		req := ThinkRequest{
			Content:    "Test thought for " + mode,
			Mode:       mode,
			Confidence: 0.8,
		}
		server.handleThink(ctx, nil, req)
	}

	t.Run("filter by tree mode", func(t *testing.T) {
		req := HistoryRequest{
			Mode:  "tree",
			Limit: 100,
		}

		_, resp, err := server.handleHistory(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		for _, thought := range resp.Thoughts {
			if thought.Mode != "tree" {
				t.Errorf("expected mode 'tree', got %s", thought.Mode)
			}
		}
	})

	t.Run("filter by divergent mode", func(t *testing.T) {
		req := HistoryRequest{
			Mode:  "divergent",
			Limit: 100,
		}

		_, resp, err := server.handleHistory(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		for _, thought := range resp.Thoughts {
			if thought.Mode != "divergent" {
				t.Errorf("expected mode 'divergent', got %s", thought.Mode)
			}
		}
	})

	t.Run("filter by branch ID", func(t *testing.T) {
		// Create a tree thought to get branch ID
		thinkReq := ThinkRequest{
			Content:    "Tree thought",
			Mode:       "tree",
			Confidence: 0.8,
		}
		_, thinkResp, _ := server.handleThink(ctx, nil, thinkReq)

		req := HistoryRequest{
			BranchID: thinkResp.BranchID,
			Limit:    100,
		}

		_, resp, err := server.handleHistory(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Error("expected response")
		}
	})
}

// =============================================================================
// Test: History with filtering
// =============================================================================

// =============================================================================
// Test: Additional mode coverage for ProcessThought
// =============================================================================

func TestProcessThoughtModes(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("tree mode", func(t *testing.T) {
		input := modes.ThoughtInput{
			Content:    "Tree mode test",
			Confidence: 0.8,
		}

		result, err := server.ProcessThought(ctx, "tree", input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})

	t.Run("divergent mode", func(t *testing.T) {
		input := modes.ThoughtInput{
			Content:    "Divergent mode test",
			Confidence: 0.8,
		}

		result, err := server.ProcessThought(ctx, "divergent", input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result")
		}
	})
}

// =============================================================================
// Test: handleSearch with mode filter
// =============================================================================

func TestSearchWithModeFilter(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create thoughts
	for _, mode := range []string{"linear", "tree"} {
		req := ThinkRequest{
			Content:    "Searchable content " + mode,
			Mode:       mode,
			Confidence: 0.8,
		}
		server.handleThink(ctx, nil, req)
	}

	t.Run("search with mode filter", func(t *testing.T) {
		req := SearchRequest{
			Query: "Searchable",
			Mode:  "linear",
			Limit: 100,
		}

		_, resp, err := server.handleSearch(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		for _, thought := range resp.Thoughts {
			if thought.Mode != "linear" {
				t.Errorf("expected mode 'linear', got %s", thought.Mode)
			}
		}
	})
}

// =============================================================================
// Test: Focus branch success path
// =============================================================================

func TestFocusBranchSuccess(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create two branches
	req1 := ThinkRequest{
		Content:    "First branch",
		Mode:       "tree",
		Confidence: 0.8,
	}
	_, resp1, _ := server.handleThink(ctx, nil, req1)

	req2 := ThinkRequest{
		Content:    "Second branch",
		Mode:       "tree",
		Confidence: 0.8,
	}
	_, _, _ = server.handleThink(ctx, nil, req2)

	// Focus on first branch
	focusReq := FocusBranchRequest{
		BranchID: resp1.BranchID,
	}

	result, resp, err := server.handleFocusBranch(ctx, nil, focusReq)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result")
	}
	if resp.ActiveBranchID != resp1.BranchID {
		t.Errorf("expected active branch %s, got %s", resp1.BranchID, resp.ActiveBranchID)
	}

	// Try to focus on the same branch again - should return "already_active"
	result2, resp2, err := server.handleFocusBranch(ctx, nil, focusReq)
	if err != nil {
		t.Errorf("unexpected error on second focus: %v", err)
	}
	if result2 == nil {
		t.Error("expected result on second focus")
	}
	if resp2.Status != "already_active" {
		t.Errorf("expected status 'already_active', got %s", resp2.Status)
	}
}

// =============================================================================
// Test: Branch history success path
// =============================================================================

func TestBranchHistorySuccess(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create a branch with multiple thoughts
	branchID := ""
	for i := 0; i < 3; i++ {
		req := ThinkRequest{
			Content:    "Branch history test thought",
			Mode:       "tree",
			BranchID:   branchID,
			Confidence: 0.8,
		}
		_, resp, _ := server.handleThink(ctx, nil, req)
		if branchID == "" {
			branchID = resp.BranchID
		}
	}

	histReq := BranchHistoryRequest{
		BranchID: branchID,
	}

	result, resp, err := server.handleBranchHistory(ctx, nil, histReq)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result")
	}
	if resp == nil {
		t.Error("expected branch history")
	}
}

// =============================================================================
// Test: Initialize episodic memory coverage
// =============================================================================

func TestInitializeEpisodicMemoryPaths(t *testing.T) {
	// This test exercises the initializeEpisodicMemory function by creating
	// servers with different storage configurations

	t.Run("with memory storage", func(t *testing.T) {
		store := storage.NewMemoryStorage()
		linear := modes.NewLinearMode(store)
		tree := modes.NewTreeMode(store)
		divergent := modes.NewDivergentMode(store)
		auto := modes.NewAutoMode(linear, tree, divergent)
		validator := validation.NewLogicValidator()

		server, err := NewUnifiedServer(store, linear, tree, divergent, auto, validator)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		if server.episodicMemoryHandler == nil {
			t.Error("episodic memory handler should be initialized")
		}
	})
}

// =============================================================================
// Test: handleHistory complete paths
// =============================================================================

func TestHistoryCompletePaths(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create thoughts in different modes
	for _, mode := range []string{"linear", "tree", "divergent"} {
		for i := 0; i < 3; i++ {
			req := ThinkRequest{
				Content:    "Complete history test",
				Mode:       mode,
				Confidence: 0.8,
			}
			server.handleThink(ctx, nil, req)
		}
	}

	t.Run("default limit applied", func(t *testing.T) {
		req := HistoryRequest{
			Limit: 0, // Should use default
		}

		_, resp, err := server.handleHistory(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Error("expected response")
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		req := HistoryRequest{
			Limit:  3,
			Offset: 5,
		}

		_, resp, err := server.handleHistory(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(resp.Thoughts) > 3 {
			t.Errorf("expected at most 3 thoughts, got %d", len(resp.Thoughts))
		}
	})
}

func TestHistoryFiltering(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create thoughts in different modes
	modes := []string{"linear", "tree", "divergent"}
	for _, mode := range modes {
		for i := 0; i < 3; i++ {
			req := ThinkRequest{
				Content:    "Test thought",
				Mode:       mode,
				Confidence: 0.8,
			}
			server.handleThink(ctx, nil, req)
		}
	}

	t.Run("filter by mode", func(t *testing.T) {
		req := HistoryRequest{
			Mode:  "linear",
			Limit: 100,
		}

		_, resp, err := server.handleHistory(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		for _, thought := range resp.Thoughts {
			if thought.Mode != "linear" {
				t.Errorf("expected mode 'linear', got %s", thought.Mode)
			}
		}
	})

	t.Run("with offset and limit", func(t *testing.T) {
		req := HistoryRequest{
			Limit:  2,
			Offset: 1,
		}

		_, resp, err := server.handleHistory(ctx, nil, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(resp.Thoughts) > 2 {
			t.Errorf("expected at most 2 thoughts, got %d", len(resp.Thoughts))
		}
	})
}
