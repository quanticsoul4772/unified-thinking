// Package types provides common types and interfaces for MCP tool handlers.
package types

// EmptyRequest represents handlers that don't need input parameters.
// Used by tools like list-branches, get-metrics, and recent-branches.
type EmptyRequest struct{}

// StatusResponse represents a simple status response for operations
// that primarily perform actions rather than return data.
type StatusResponse struct {
	Status string `json:"status"`
}
