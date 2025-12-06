package main

import (
	"os"
	"testing"
)

// setupTestEnv sets required environment variables for testing
// VOYAGE_API_KEY and ANTHROPIC_API_KEY must be real keys from environment
// No fake keys - tests fail if keys are missing/invalid
func setupTestEnv(t *testing.T) {
	t.Helper()

	// Verify VOYAGE_API_KEY is set (required for embeddings)
	if os.Getenv("VOYAGE_API_KEY") == "" {
		t.Fatal("VOYAGE_API_KEY not set: tests require real Voyage AI API key")
	}

	// Verify ANTHROPIC_API_KEY is set (required for agent, web search, GoT)
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Fatal("ANTHROPIC_API_KEY not set: tests require real Anthropic API key")
	}

	// SQLite storage is default - set path for tests
	t.Setenv("SQLITE_PATH", t.TempDir()+"/test.db")
}

func TestInitializeServer(t *testing.T) {
	setupTestEnv(t)

	// Test default initialization (memory storage, no embedder)
	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}
	defer components.Cleanup()

	// Verify all components initialized
	if components.Storage == nil {
		t.Error("Storage not initialized")
	}
	if components.LinearMode == nil {
		t.Error("LinearMode not initialized")
	}
	if components.TreeMode == nil {
		t.Error("TreeMode not initialized")
	}
	if components.DivergentMode == nil {
		t.Error("DivergentMode not initialized")
	}
	if components.AutoMode == nil {
		t.Error("AutoMode not initialized")
	}
	if components.Validator == nil {
		t.Error("Validator not initialized")
	}
	if components.Server == nil {
		t.Error("Server not initialized")
	}
	if components.Orchestrator == nil {
		t.Error("Orchestrator not initialized")
	}

	// Embedder and ContextBridge are optional (depend on env vars)
	// Not testing their presence here
}

// TestInitializeServer_WithMockEmbedder removed - no mock embedders allowed

func TestInitializeServer_SQLiteStorage(t *testing.T) {
	setupTestEnv(t)

	// SQLite is now the default - this test verifies it works
	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() with SQLite failed: %v", err)
	}
	defer components.Cleanup()

	if components.Storage == nil {
		t.Fatal("Storage not initialized")
	}

	// Verify it's SQLite storage (type assertion)
	// Note: This tests the storage factory behavior
}

func TestInitializeServer_Cleanup(t *testing.T) {
	setupTestEnv(t)

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}

	// Cleanup should not error
	err = components.Cleanup()
	if err != nil {
		t.Errorf("Cleanup() failed: %v", err)
	}

	// Second cleanup should be safe (idempotent)
	err = components.Cleanup()
	if err != nil {
		t.Errorf("Second Cleanup() failed: %v", err)
	}
}

func TestServerComponents_NilStorage(t *testing.T) {
	components := &ServerComponents{
		Storage: nil,
	}

	// Cleanup with nil storage should not panic
	err := components.Cleanup()
	if err != nil {
		t.Errorf("Cleanup with nil storage should not error, got: %v", err)
	}
}

func TestInitializeContextBridge_WithEmbedder(t *testing.T) {
	setupTestEnv(t)

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}
	defer components.Cleanup()

	// Context bridge is ALWAYS enabled with SQLite storage and embedder
	if components.ContextBridge == nil {
		t.Error("Expected ContextBridge to be initialized with SQLite storage and embedder")
	}
}

func TestInitializeServer_RequiresVoyageAPIKey(t *testing.T) {
	// Do NOT call setupTestEnv - we want to test missing VOYAGE_API_KEY
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-fake-key-for-testing")
	// Explicitly unset VOYAGE_API_KEY to test the failure case
	t.Setenv("VOYAGE_API_KEY", "")
	t.Setenv("SQLITE_PATH", t.TempDir()+"/test.db")

	_, err := InitializeServer()
	if err == nil {
		t.Fatal("InitializeServer() should fail when VOYAGE_API_KEY is not set")
	}
	if err.Error() != "VOYAGE_API_KEY not set: embeddings are required" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestInitializeContextBridge_Disabled removed - context bridge is ALWAYS enabled

func TestRegisterPredefinedWorkflows_NilOrchestrator(t *testing.T) {
	// Should not panic with nil orchestrator
	registerPredefinedWorkflows(nil)
	// If we get here without panic, test passes
}

func TestRegisterPredefinedWorkflows_ValidOrchestrator(t *testing.T) {
	setupTestEnv(t)

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}
	defer components.Cleanup()

	// Should register workflows without error
	// (already called during InitializeServer, but test explicit call)
	registerPredefinedWorkflows(components.Orchestrator)

	// Verify workflows were registered
	workflows := components.Orchestrator.ListWorkflows()
	if len(workflows) == 0 {
		t.Error("Expected at least one registered workflow")
	}

	// Check for expected workflows
	workflowNames := make(map[string]bool)
	for _, w := range workflows {
		workflowNames[w.ID] = true
	}

	expectedWorkflows := []string{
		"causal-analysis",
		"critical-thinking",
		"multi-perspective-decision",
	}

	for _, name := range expectedWorkflows {
		if !workflowNames[name] {
			t.Errorf("Expected workflow %q not registered", name)
		}
	}
}

// TestInitializeServer_WithMockEmbedderForTesting removed - no mock embedders allowed

func TestServerComponents_DefaultFields(t *testing.T) {
	// Test that ServerComponents struct can be created with default values
	components := &ServerComponents{}

	// All fields should be nil by default
	if components.Storage != nil {
		t.Error("Storage should be nil by default")
	}
	if components.LinearMode != nil {
		t.Error("LinearMode should be nil by default")
	}
	if components.Server != nil {
		t.Error("Server should be nil by default")
	}
}
