package server

import (
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

func TestRegisterTools(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test requiring full server")
	}

	// Setup test server
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

	// Create a mock MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0",
	}, nil)

	// Test that RegisterTools doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterTools panicked: %v", r)
		}
	}()

	server.RegisterTools(mcpServer)

	// We can't easily check the tools since the MCP SDK doesn't expose a list method
	// But if we get here without panic, the registration worked
	t.Log("RegisterTools completed without error")
}

func TestSetOrchestrator(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test requiring full server")
	}

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

	// Test that we can set it without error
	executor := NewServerToolExecutor(server)
	orch := orchestration.NewOrchestratorWithExecutor(executor)

	// Test that SetOrchestrator doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetOrchestrator panicked: %v", r)
		}
	}()

	server.SetOrchestrator(orch)
	t.Log("SetOrchestrator completed without error")
}

func TestInitializeAdvancedHandlers(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test requiring full server")
	}

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

	// initializeAdvancedHandlers is called in NewUnifiedServer, so handlers should already be initialized
	// We can verify this indirectly by checking that server creation succeeded

	if server == nil {
		t.Fatal("expected server to be created")
	}

	t.Log("Server created successfully with advanced handlers")
}
