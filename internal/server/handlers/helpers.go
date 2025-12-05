package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"unified-thinking/internal/server/types"
)

// unmarshalParams unmarshals MCP parameters into a struct.
// Deprecated: Use types.UnmarshalRequest[T] for new code. This function
// is retained for backward compatibility with existing handlers.
func unmarshalParams(params map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal params: %w", err)
	}

	return nil
}

// unmarshalRequest is the generic version of unmarshalParams.
// It uses the centralized types.UnmarshalRequest[T] adapter.
func unmarshalRequest[T any](params map[string]interface{}) (T, error) {
	return types.UnmarshalRequest[T](params)
}

// generateID generates a unique ID with a prefix
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
