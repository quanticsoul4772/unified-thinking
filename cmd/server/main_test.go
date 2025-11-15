package main

import (
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/server"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	closed bool
}

// ThoughtRepository methods
func (m *MockStorage) StoreThought(thought *types.Thought) error { return nil }
func (m *MockStorage) GetThought(id string) (*types.Thought, error) { return nil, nil }
func (m *MockStorage) SearchThoughts(query string, mode types.ThinkingMode, limit, offset int) []*types.Thought {
	return nil
}

// BranchRepository methods
func (m *MockStorage) StoreBranch(branch *types.Branch) error { return nil }
func (m *MockStorage) GetBranch(id string) (*types.Branch, error) { return nil, nil }
func (m *MockStorage) ListBranches() []*types.Branch { return nil }
func (m *MockStorage) GetActiveBranch() (*types.Branch, error) { return nil, nil }
func (m *MockStorage) SetActiveBranch(branchID string) error { return nil }
func (m *MockStorage) UpdateBranchAccess(branchID string) error { return nil }
func (m *MockStorage) AppendThoughtToBranch(branchID string, thought *types.Thought) error { return nil }
func (m *MockStorage) AppendInsightToBranch(branchID string, insight *types.Insight) error { return nil }
func (m *MockStorage) AppendCrossRefToBranch(branchID string, crossRef *types.CrossRef) error { return nil }
func (m *MockStorage) UpdateBranchPriority(branchID string, priority float64) error { return nil }
func (m *MockStorage) UpdateBranchConfidence(branchID string, confidence float64) error { return nil }
func (m *MockStorage) GetRecentBranches() ([]*types.Branch, error) { return nil, nil }

// InsightRepository methods
func (m *MockStorage) StoreInsight(insight *types.Insight) error { return nil }

// ValidationRepository methods
func (m *MockStorage) StoreValidation(validation *types.Validation) error { return nil }

// RelationshipRepository methods
func (m *MockStorage) StoreRelationship(relationship *types.Relationship) error { return nil }

// MetricsProvider methods
func (m *MockStorage) GetMetrics() *storage.Metrics { return nil }

// Close method for cleanup
func (m *MockStorage) Close() error { m.closed = true; return nil }

func TestMainInitialization(t *testing.T) {
	// Save original env vars
	originalDebug := os.Getenv("DEBUG")
	defer func() {
		if originalDebug != "" {
			os.Setenv("DEBUG", originalDebug)
		} else {
			os.Unsetenv("DEBUG")
		}
	}()

	tests := []struct {
		name          string
		debugEnv      string
		expectDebug   bool
		shouldSucceed bool
	}{
		{
			name:          "debug mode enabled",
			debugEnv:      "true",
			expectDebug:   true,
			shouldSucceed: true,
		},
		{
			name:          "debug mode disabled",
			debugEnv:      "false",
			expectDebug:   false,
			shouldSucceed: true,
		},
		{
			name:          "debug mode not set",
			debugEnv:      "",
			expectDebug:   false,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set debug environment
			if tt.debugEnv != "" {
				os.Setenv("DEBUG", tt.debugEnv)
			} else {
				os.Unsetenv("DEBUG")
			}

			// Test storage initialization
			store, err := storage.NewStorageFromEnv()
			if err != nil {
				t.Fatalf("Failed to initialize storage: %v", err)
			}
			defer storage.CloseStorage(store)

			// Test modes initialization
			linearMode := modes.NewLinearMode(store)
			if linearMode == nil {
				t.Error("Failed to initialize linear mode")
			}

			treeMode := modes.NewTreeMode(store)
			if treeMode == nil {
				t.Error("Failed to initialize tree mode")
			}

			divergentMode := modes.NewDivergentMode(store)
			if divergentMode == nil {
				t.Error("Failed to initialize divergent mode")
			}

			autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
			if autoMode == nil {
				t.Error("Failed to initialize auto mode")
			}

			// Test validator initialization
			validator := validation.NewLogicValidator()
			if validator == nil {
				t.Error("Failed to initialize validator")
			}

			// Test server creation
			srv := server.NewUnifiedServer(store, linearMode, treeMode, divergentMode, autoMode, validator)
			if srv == nil {
				t.Error("Failed to create unified server")
			}

			// Test orchestrator initialization
			executor := server.NewServerToolExecutor(srv)
			if executor == nil {
				t.Error("Failed to create server tool executor")
			}

			orchestrator := orchestration.NewOrchestratorWithExecutor(executor)
			if orchestrator == nil {
				t.Error("Failed to initialize orchestrator")
			}

			srv.SetOrchestrator(orchestrator)

			// Test MCP server creation
			mcpServer := mcp.NewServer(&mcp.Implementation{
				Name:    "test-unified-thinking-server",
				Version: "1.0.0-test",
			}, nil)
			if mcpServer == nil {
				t.Error("Failed to create MCP server")
			}

			// Test tool registration
			srv.RegisterTools(mcpServer)

			// Test transport creation
			transport := &mcp.StdioTransport{}
			if transport == nil {
				t.Error("Failed to create stdio transport")
			}

			// Note: We don't test the actual server.Run() as it would block
			// and require stdio interaction
		})
	}
}

func TestRegisterPredefinedWorkflows(t *testing.T) {
	// Create a mock orchestrator
	executor := server.NewServerToolExecutor(nil) // nil server for testing
	orchestrator := orchestration.NewOrchestratorWithExecutor(executor)

	if orchestrator == nil {
		t.Fatal("Failed to create orchestrator")
	}

	// Test workflow registration
	registerPredefinedWorkflows(orchestrator)

	// Verify workflows were registered by listing them
	workflows := orchestrator.ListWorkflows()
	if len(workflows) == 0 {
		t.Error("No workflows were registered")
	}

	// Check for expected workflow IDs
	expectedWorkflows := []string{
		"causal-analysis",
		"critical-thinking",
		"multi-perspective-decision",
	}

	registeredWorkflowIDs := make(map[string]bool)
	for _, workflow := range workflows {
		registeredWorkflowIDs[workflow.ID] = true
	}

	for _, expectedID := range expectedWorkflows {
		if !registeredWorkflowIDs[expectedID] {
			t.Errorf("Expected workflow %s was not registered", expectedID)
		}
	}
}

func TestWorkflowStructure(t *testing.T) {
	executor := server.NewServerToolExecutor(nil)
	orchestrator := orchestration.NewOrchestratorWithExecutor(executor)

	registerPredefinedWorkflows(orchestrator)

	workflows := orchestrator.ListWorkflows()

	// Test causal analysis workflow structure
	causalWorkflow := findWorkflowByID(workflows, "causal-analysis")
	if causalWorkflow == nil {
		t.Fatal("causal-analysis workflow not found")
	}

	if len(causalWorkflow.Steps) != 3 {
		t.Errorf("causal-analysis workflow should have 3 steps, got %d", len(causalWorkflow.Steps))
	}

	// Verify step dependencies
	steps := causalWorkflow.Steps
	if steps[1].DependsOn == nil || len(steps[1].DependsOn) == 0 {
		t.Error("Second step should depend on first step")
	}

	if steps[2].DependsOn == nil || len(steps[2].DependsOn) == 0 {
		t.Error("Third step should depend on second step")
	}

	// Test critical thinking workflow structure
	criticalWorkflow := findWorkflowByID(workflows, "critical-thinking")
	if criticalWorkflow == nil {
		t.Fatal("critical-thinking workflow not found")
	}

	if len(criticalWorkflow.Steps) != 3 {
		t.Errorf("critical-thinking workflow should have 3 steps, got %d", len(criticalWorkflow.Steps))
	}

	// Test conditional step
	proveStep := findStepByID(criticalWorkflow.Steps, "prove")
	if proveStep == nil {
		t.Fatal("prove step not found in critical-thinking workflow")
	}

	if proveStep.Condition == nil {
		t.Error("prove step should have a condition")
	}

	if proveStep.Condition.Type != "result_match" {
		t.Errorf("Expected condition type 'result_match', got %s", proveStep.Condition.Type)
	}
}

func TestParallelWorkflow(t *testing.T) {
	executor := server.NewServerToolExecutor(nil)
	orchestrator := orchestration.NewOrchestratorWithExecutor(executor)

	registerPredefinedWorkflows(orchestrator)

	workflows := orchestrator.ListWorkflows()

	// Test multi-perspective decision workflow
	decisionWorkflow := findWorkflowByID(workflows, "multi-perspective-decision")
	if decisionWorkflow == nil {
		t.Fatal("multi-perspective-decision workflow not found")
	}

	if decisionWorkflow.Type != orchestration.WorkflowParallel {
		t.Errorf("Expected parallel workflow type, got %v", decisionWorkflow.Type)
	}

	if len(decisionWorkflow.Steps) != 3 {
		t.Errorf("decision workflow should have 3 steps, got %d", len(decisionWorkflow.Steps))
	}

	// Check that the final step depends on the parallel steps
	finalStep := findStepByID(decisionWorkflow.Steps, "make-decision")
	if finalStep == nil {
		t.Fatal("make-decision step not found")
	}

	if len(finalStep.DependsOn) != 2 {
		t.Errorf("Final step should depend on 2 parallel steps, got %d", len(finalStep.DependsOn))
	}
}

func TestErrorHandling(t *testing.T) {
	// Test with nil orchestrator
	var nilOrchestrator *orchestration.Orchestrator

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("registerPredefinedWorkflows panicked with nil orchestrator: %v", r)
		}
	}()

	registerPredefinedWorkflows(nilOrchestrator)
}

func TestStorageCleanup(t *testing.T) {
	mockStorage := &MockStorage{}

	// Simulate the defer cleanup from main
	err := storage.CloseStorage(mockStorage)
	if err != nil {
		t.Errorf("Failed to close storage: %v", err)
	}

	if !mockStorage.closed {
		t.Error("Storage was not properly closed")
	}
}

// Helper functions
func findWorkflowByID(workflows []*orchestration.Workflow, id string) *orchestration.Workflow {
	for _, workflow := range workflows {
		if workflow.ID == id {
			return workflow
		}
	}
	return nil
}

func findStepByID(steps []*orchestration.WorkflowStep, id string) *orchestration.WorkflowStep {
	for _, step := range steps {
		if step.ID == id {
			return step
		}
	}
	return nil
}

// Benchmark tests
func BenchmarkWorkflowRegistration(b *testing.B) {
	executor := server.NewServerToolExecutor(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orchestrator := orchestration.NewOrchestratorWithExecutor(executor)
		registerPredefinedWorkflows(orchestrator)
	}
}

func BenchmarkServerCreation(b *testing.B) {
	store := &MockStorage{}

	linearMode := modes.NewLinearMode(store)
	treeMode := modes.NewTreeMode(store)
	divergentMode := modes.NewDivergentMode(store)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	validator := validation.NewLogicValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv := server.NewUnifiedServer(store, linearMode, treeMode, divergentMode, autoMode, validator)
		if srv == nil {
			b.Fatal("Failed to create server")
		}
	}
}
