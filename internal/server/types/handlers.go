// Package types provides common types and interfaces for MCP tool handlers.
package types

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Handler represents a tool handler function signature.
// This maintains backward compatibility with MCP's handler pattern.
// The generic parameters TReq and TResp represent the request and response types.
type HandlerFunc[TReq any, TResp any] func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input TReq,
) (*mcp.CallToolResult, *TResp, error)

// ToolRegistration encapsulates a tool and its handler for registration.
// This provides a cleaner way to organize tool definitions and their implementations.
type ToolRegistration struct {
	Tool    *mcp.Tool
	Handler interface{} // Stores the handler function
}
