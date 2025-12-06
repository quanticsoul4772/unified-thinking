// Package types provides common types and interfaces for MCP tool handlers.
package types

import (
	"encoding/json"
	"fmt"
)

// UnmarshalRequest converts untyped MCP params to a typed request struct.
// This is the primary adapter for bridging the untyped MCP SDK boundary
// to our typed internal handler methods.
//
// Usage:
//
//	req, err := types.UnmarshalRequest[MyRequest](params)
//	if err != nil {
//	    return nil, fmt.Errorf("invalid request: %w", err)
//	}
//	// Use typed req from here
func UnmarshalRequest[T any](params map[string]interface{}) (T, error) {
	var req T

	// Handle nil params - return zero value of T
	if params == nil {
		return req, nil
	}

	// Marshal params to JSON, then unmarshal to typed struct
	// This leverages JSON struct tags for field mapping
	data, err := json.Marshal(params)
	if err != nil {
		return req, fmt.Errorf("marshal params: %w", err)
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return req, fmt.Errorf("unmarshal to %T: %w", req, err)
	}

	return req, nil
}

// UnmarshalRequestWithValidation converts params to typed request and validates.
// Use this when the request type implements the Validatable interface.
func UnmarshalRequestWithValidation[T Validatable](params map[string]interface{}) (T, error) {
	req, err := UnmarshalRequest[T](params)
	if err != nil {
		return req, err
	}

	if err := req.Validate(); err != nil {
		return req, fmt.Errorf("validation failed: %w", err)
	}

	return req, nil
}

// Validatable is implemented by request types that support validation.
type Validatable interface {
	Validate() error
}

// MarshalResponse converts a typed response to JSON bytes for MCP content.
// This is the companion to UnmarshalRequest for the response path.
func MarshalResponse(resp interface{}) ([]byte, error) {
	if resp == nil {
		return []byte("{}"), nil
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}

	return data, nil
}

// MarshalResponseString converts a typed response to a JSON string.
// Convenience wrapper around MarshalResponse for common use case.
func MarshalResponseString(resp interface{}) (string, error) {
	data, err := MarshalResponse(resp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONContent marshals a response and wraps it in MCP text content format.
// This is the standard pattern for returning tool results.
func ToJSONContent(resp interface{}) string {
	data, err := MarshalResponseString(resp)
	if err != nil {
		// Return error as JSON for consistent format
		return fmt.Sprintf(`{"error": "marshal failed: %s"}`, err.Error())
	}
	return data
}

// ExtractString safely extracts a string value from params map.
// Returns empty string if key doesn't exist or value isn't a string.
func ExtractString(params map[string]interface{}, key string) string {
	if params == nil {
		return ""
	}
	if v, ok := params[key].(string); ok {
		return v
	}
	return ""
}

// ExtractInt safely extracts an int value from params map.
// Handles both int and float64 (JSON number default) types.
// Returns 0 if key doesn't exist or value isn't numeric.
func ExtractInt(params map[string]interface{}, key string) int {
	if params == nil {
		return 0
	}
	switch v := params[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// ExtractFloat safely extracts a float64 value from params map.
// Returns 0 if key doesn't exist or value isn't numeric.
func ExtractFloat(params map[string]interface{}, key string) float64 {
	if params == nil {
		return 0
	}
	switch v := params[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}

// ExtractBool safely extracts a bool value from params map.
// Returns false if key doesn't exist or value isn't a bool.
func ExtractBool(params map[string]interface{}, key string) bool {
	if params == nil {
		return false
	}
	if v, ok := params[key].(bool); ok {
		return v
	}
	return false
}

// ExtractStringSlice safely extracts a []string from params map.
// Returns nil if key doesn't exist or value can't be converted.
func ExtractStringSlice(params map[string]interface{}, key string) []string {
	if params == nil {
		return nil
	}
	v, ok := params[key]
	if !ok {
		return nil
	}

	// Try direct string slice
	if ss, ok := v.([]string); ok {
		return ss
	}

	// Try interface slice (common from JSON unmarshal)
	if is, ok := v.([]interface{}); ok {
		result := make([]string, 0, len(is))
		for _, item := range is {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	return nil
}

// ExtractMap safely extracts a nested map from params.
// Returns nil if key doesn't exist or value isn't a map.
func ExtractMap(params map[string]interface{}, key string) map[string]interface{} {
	if params == nil {
		return nil
	}
	if v, ok := params[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

// ToParams converts a typed struct to map[string]interface{} for handler calls.
// This is primarily useful in tests to maintain type safety while calling
// handlers that accept untyped params at the MCP boundary.
//
// Usage in tests:
//
//	req := MyRequest{Content: "test", Mode: "linear"}
//	params := types.ToParams(req)
//	result, err := handler.HandleMyTool(ctx, params)
func ToParams[T any](req T) map[string]interface{} {
	data, err := json.Marshal(req)
	if err != nil {
		// Return empty map on marshal error (shouldn't happen with valid structs)
		return make(map[string]interface{})
	}

	var params map[string]interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return make(map[string]interface{})
	}

	return params
}

// MustToParams is like ToParams but panics on error.
// Use in tests where marshal failure indicates a programming error.
func MustToParams[T any](req T) map[string]interface{} {
	data, err := json.Marshal(req)
	if err != nil {
		panic(fmt.Sprintf("MustToParams: marshal failed: %v", err))
	}

	var params map[string]interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		panic(fmt.Sprintf("MustToParams: unmarshal failed: %v", err))
	}

	return params
}
