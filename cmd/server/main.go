// Package main provides the entry point for the Unified Thinking MCP server.
//
// This server is designed to be spawned as a child process by Claude Desktop
// and communicates via stdio using the Model Context Protocol. It should not
// be run manually by users.
//
// The server consolidates multiple cognitive thinking patterns (linear, tree,
// divergent, and auto modes) into a single Go-based MCP server, providing
// 9 tools for thought processing, validation, and search.
//
// Environment variables:
//   - DEBUG: Set to "true" to enable debug logging
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/server"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

func main() {
	// Enable debug logging if DEBUG env var is set
	if os.Getenv("DEBUG") == "true" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Starting Unified Thinking Server in debug mode...")
	}

	// Initialize storage
	store := storage.NewMemoryStorage()
	log.Println("Initialized memory storage")

	// Initialize modes
	linearMode := modes.NewLinearMode(store)
	treeMode := modes.NewTreeMode(store)
	divergentMode := modes.NewDivergentMode(store)
	autoMode := modes.NewAutoMode(linearMode, treeMode, divergentMode)
	log.Println("Initialized thinking modes: linear, tree, divergent, auto")

	// Initialize validator
	validator := validation.NewLogicValidator()
	log.Println("Initialized logic validator")

	// Create unified server
	srv := server.NewUnifiedServer(store, linearMode, treeMode, divergentMode, autoMode, validator)
	log.Println("Created unified server")

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "unified-thinking-server",
		Version: "1.0.0",
	}, nil)
	log.Println("Created MCP server")

	// Register tools
	srv.RegisterTools(mcpServer)
	log.Println("Registered tools: think, history, list-branches, focus-branch, branch-history, validate, prove, check-syntax, search")

	// Create stdio transport
	transport := &mcp.StdioTransport{}
	log.Println("Created stdio transport")

	// Run server
	ctx := context.Background()
	log.Println("Starting MCP server...")
	if err := mcpServer.Run(ctx, transport); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
