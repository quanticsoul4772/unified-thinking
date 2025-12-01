package format

// CompactFormatter removes context_bridge and flattens metadata for 40-60% size reduction
type CompactFormatter struct {
	opts FormatOptions
}

// Format transforms a response to compact format
func (f *CompactFormatter) Format(response any) (any, error) {
	// Convert to map for manipulation
	data, err := toMap(response)
	if err != nil {
		return response, nil // Return unchanged if conversion fails
	}

	result := make(map[string]any)

	// Check if this is a wrapped response with "result" field
	if innerResult, hasResult := data["result"]; hasResult {
		innerMap, ok := innerResult.(map[string]any)
		if ok {
			// Flatten the result to top level
			for k, v := range innerMap {
				result[k] = v
			}
		} else {
			result["result"] = innerResult
		}
	} else {
		// Copy all non-context_bridge fields
		for k, v := range data {
			if k == "context_bridge" {
				continue // Remove context_bridge entirely
			}
			result[k] = v
		}
	}

	// Flatten suggested_next_tools from metadata to next_tools
	if f.opts.FlattenNextTools {
		result = f.flattenNextTools(result)
	}

	// Remove metadata if not needed
	if !f.opts.IncludeMetadata {
		delete(result, "metadata")
	}

	// Remove timing information
	if !f.opts.IncludeTimings {
		delete(result, "duration_ms")
		delete(result, "timing")
		delete(result, "timings")
	}

	// Truncate arrays if configured
	if f.opts.MaxArrayLength > 0 {
		result = truncateArrays(result, f.opts.MaxArrayLength)
	}

	// Remove empty fields if configured
	if f.opts.OmitEmpty {
		result = cleanMap(result)
	}

	return result, nil
}

// Level returns FormatCompact
func (f *CompactFormatter) Level() FormatLevel {
	return FormatCompact
}

// flattenNextTools extracts suggested_next_tools from metadata to top-level next_tools
func (f *CompactFormatter) flattenNextTools(data map[string]any) map[string]any {
	metadata, ok := data["metadata"].(map[string]any)
	if !ok {
		return data
	}

	// Extract suggested_next_tools
	if nextTools, hasTools := metadata["suggested_next_tools"]; hasTools {
		// Convert to simple array of tool names
		if toolsArr, isArr := nextTools.([]any); isArr {
			simplified := make([]string, 0, len(toolsArr))
			for _, t := range toolsArr {
				if toolMap, isMap := t.(map[string]any); isMap {
					if name, hasName := toolMap["tool"].(string); hasName {
						simplified = append(simplified, name)
					}
				} else if name, isStr := t.(string); isStr {
					simplified = append(simplified, name)
				}
			}
			if len(simplified) > 0 {
				data["next_tools"] = simplified
			}
		}
		delete(metadata, "suggested_next_tools")
	}

	// Remove action_recommendations (verbose)
	delete(metadata, "action_recommendations")

	// Remove export_formats (rarely needed)
	delete(metadata, "export_formats")

	// Update or remove metadata
	if len(metadata) == 0 {
		delete(data, "metadata")
	} else {
		data["metadata"] = metadata
	}

	return data
}
