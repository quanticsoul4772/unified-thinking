package format

// MinimalFormatter returns only essential result fields for 80%+ size reduction
type MinimalFormatter struct {
	opts FormatOptions
}

// essentialFields defines the fields to keep for each response type
var essentialFields = map[string][]string{
	// Thinking responses
	"think": {"thought_id", "mode", "confidence", "is_valid"},
	// Branch responses
	"branch": {"branch_id", "state", "confidence"},
	// Decision responses
	"decision": {"decision_id", "selected_option", "confidence"},
	// Search responses
	"search": {"results", "count"},
	// Validation responses
	"validation": {"is_valid", "reason"},
	// Session responses
	"session": {"session_id", "status"},
	// Preset responses
	"preset": {"preset_id", "status", "steps_completed"},
	// Generic default
	"default": {"id", "status", "confidence", "result"},
}

// Format transforms a response to minimal format
func (f *MinimalFormatter) Format(response any) (any, error) {
	// Convert to map for manipulation
	data, err := toMap(response)
	if err != nil {
		return response, nil // Return unchanged if conversion fails
	}

	// Unwrap result if present
	if innerResult, hasResult := data["result"]; hasResult {
		if innerMap, ok := innerResult.(map[string]any); ok {
			data = innerMap
		}
	}

	// Detect response type and get essential fields
	responseType := detectResponseType(data)
	fields := essentialFields[responseType]
	if fields == nil {
		fields = essentialFields["default"]
	}

	// Build minimal response
	result := make(map[string]any)
	for _, field := range fields {
		if v, exists := data[field]; exists && !isEmpty(v) {
			result[field] = v
		}
	}

	// Always include error information if present
	if errMsg, hasErr := data["error"]; hasErr {
		result["error"] = errMsg
	}
	if errCode, hasCode := data["error_code"]; hasCode {
		result["error_code"] = errCode
	}

	// Truncate arrays if configured
	if f.opts.MaxArrayLength > 0 {
		result = truncateArrays(result, f.opts.MaxArrayLength)
	}

	// Ensure we have at least something
	if len(result) == 0 {
		// Return a simplified version of the original
		return simplifyResponse(data), nil
	}

	return result, nil
}

// Level returns FormatMinimal
func (f *MinimalFormatter) Level() FormatLevel {
	return FormatMinimal
}

// detectResponseType identifies the type of response based on fields present
func detectResponseType(data map[string]any) string {
	// Check for specific identifier fields
	if _, has := data["thought_id"]; has {
		return "think"
	}
	if _, has := data["branch_id"]; has {
		return "branch"
	}
	if _, has := data["decision_id"]; has {
		return "decision"
	}
	if _, has := data["session_id"]; has {
		return "session"
	}
	if _, has := data["preset_id"]; has {
		return "preset"
	}
	if _, has := data["is_valid"]; has {
		return "validation"
	}
	if _, has := data["results"]; has {
		return "search"
	}
	return "default"
}

// simplifyResponse creates a minimal representation of any response
func simplifyResponse(data map[string]any) map[string]any {
	result := make(map[string]any)

	// Priority fields to include
	priorityFields := []string{"id", "status", "success", "result", "confidence", "count", "error"}

	for _, field := range priorityFields {
		if v, exists := data[field]; exists && !isEmpty(v) {
			// For complex values, simplify further
			if nested, ok := v.(map[string]any); ok {
				result[field] = simplifyNested(nested)
			} else {
				result[field] = v
			}
		}
	}

	// If still empty, just return success indicator
	if len(result) == 0 {
		result["status"] = "ok"
	}

	return result
}

// simplifyNested reduces a nested map to its key values
func simplifyNested(data map[string]any) any {
	// If it has an id, return just that
	if id, has := data["id"]; has {
		return id
	}
	// If it has a name, return that
	if name, has := data["name"]; has {
		return name
	}
	// Otherwise return count of fields
	return len(data)
}
