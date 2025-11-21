// Package main provides the entry point for the Unified Thinking MCP server.
//
// This server is designed to be spawned as a child process by Claude Desktop
// and communicates via stdio using the Model Context Protocol. It should not
// be run manually by users.
//
// The server consolidates multiple cognitive thinking patterns (linear, tree,
// divergent, and auto modes) into a single Go-based MCP server, providing
// 10 tools for thought processing, validation, search, and metrics.
//
// Environment variables:
//   - DEBUG: Set to "true" to enable debug logging
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Enable debug logging if DEBUG env var is set
	if os.Getenv("DEBUG") == "true" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Starting Unified Thinking Server in debug mode...")
	}

	// Initialize all server components
	components, err := InitializeServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	defer func() {
		if err := components.Cleanup(); err != nil {
			log.Printf("Warning: failed to cleanup: %v", err)
		}
	}()

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "unified-thinking-server",
		Version: "1.0.0",
	}, nil)
	log.Println("Created MCP server")

	// Register tools
	components.Server.RegisterTools(mcpServer)
	log.Println("Registered tools: think, history, list-branches, focus-branch, branch-history, validate, prove, check-syntax, search, get-metrics, execute-workflow, list-workflows, register-workflow, and 21 additional reasoning tools")

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
