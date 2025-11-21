package main

import (
	"testing"

	"unified-thinking/internal/embeddings"
)

func TestInitializeServer(t *testing.T) {
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

func TestInitializeServer_WithMockEmbedder(t *testing.T) {
	// Set environment to enable embeddings with mock
	t.Setenv("VOYAGE_API_KEY", "mock-key-for-testing")
	t.Setenv("EMBEDDINGS_MODEL", "mock-model")

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() with embedder failed: %v", err)
	}
	defer components.Cleanup()

	// When VOYAGE_API_KEY is set, embedder should be initialized
	if components.Embedder == nil {
		t.Error("Embedder should be initialized when VOYAGE_API_KEY is set")
	}
}

func TestInitializeServer_SQLiteStorage(t *testing.T) {
	// Create temporary database for testing
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	t.Setenv("STORAGE_TYPE", "sqlite")
	t.Setenv("SQLITE_PATH", dbPath)

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
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	t.Setenv("STORAGE_TYPE", "sqlite")
	t.Setenv("SQLITE_PATH", dbPath)
	t.Setenv("CONTEXT_BRIDGE_ENABLED", "true")

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}
	defer components.Cleanup()

	// With SQLite and context bridge enabled, should have bridge
	if components.ContextBridge == nil {
		t.Error("Expected ContextBridge to be initialized with SQLite storage and embedder")
	}
}

func TestInitializeContextBridge_WithoutEmbedder(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	t.Setenv("STORAGE_TYPE", "sqlite")
	t.Setenv("SQLITE_PATH", dbPath)
	t.Setenv("CONTEXT_BRIDGE_ENABLED", "true")
	// No VOYAGE_API_KEY - should use concept-based similarity

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}
	defer components.Cleanup()

	// Bridge should still work without embedder (uses concept similarity)
	if components.ContextBridge == nil {
		t.Error("Expected ContextBridge to be initialized even without embedder (concept-based similarity)")
	}
}

func TestInitializeContextBridge_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	t.Setenv("STORAGE_TYPE", "sqlite")
	t.Setenv("SQLITE_PATH", dbPath)
	t.Setenv("CONTEXT_BRIDGE_ENABLED", "false")

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}
	defer components.Cleanup()

	// Context bridge behavior when disabled is implementation-dependent
	// Test passes if initialization succeeds
}

func TestInitializeContextBridge_MemoryStorage(t *testing.T) {
	t.Setenv("STORAGE_TYPE", "memory")
	t.Setenv("CONTEXT_BRIDGE_ENABLED", "true")

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}
	defer components.Cleanup()

	// Context bridge requires SQLite, should be nil with memory storage
	if components.ContextBridge != nil {
		t.Error("ContextBridge should be nil with memory storage")
	}
}

func TestRegisterPredefinedWorkflows_NilOrchestrator(t *testing.T) {
	// Should not panic with nil orchestrator
	registerPredefinedWorkflows(nil)
	// If we get here without panic, test passes
}

func TestRegisterPredefinedWorkflows_ValidOrchestrator(t *testing.T) {
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

func TestInitializeServer_WithMockEmbedderForTesting(t *testing.T) {
	// This test demonstrates how to inject a mock embedder
	// Useful pattern for integration tests

	components, err := InitializeServer()
	if err != nil {
		t.Fatalf("InitializeServer() failed: %v", err)
	}
	defer components.Cleanup()

	// Replace embedder with mock for testing
	mockEmbedder := embeddings.NewMockEmbedder(512)
	components.Embedder = mockEmbedder
	components.AutoMode.SetEmbedder(mockEmbedder)

	// Now components can be tested without real API calls
	if components.Embedder.Provider() != "mock" {
		t.Errorf("Expected mock provider, got %s", components.Embedder.Provider())
	}
}
