package format

import (
	"encoding/json"
	"reflect"
)

// ResponseFormatter defines the interface for formatting tool responses
type ResponseFormatter interface {
	// Format transforms a response according to the formatter's configuration
	Format(response any) (any, error)
	// Level returns the format level of this formatter
	Level() FormatLevel
}

// NewFormatter creates a ResponseFormatter for the given level and options
func NewFormatter(level FormatLevel, opts FormatOptions) ResponseFormatter {
	opts.Level = level
	switch level {
	case FormatCompact:
		return &CompactFormatter{opts: opts}
	case FormatMinimal:
		return &MinimalFormatter{opts: opts}
	default:
		return &FullFormatter{opts: opts}
	}
}

// FullFormatter returns responses unchanged (with optional empty removal)
type FullFormatter struct {
	opts FormatOptions
}

// Format returns the response unchanged for full format
func (f *FullFormatter) Format(response any) (any, error) {
	if f.opts.OmitEmpty {
		return removeEmptyFields(response)
	}
	return response, nil
}

// Level returns FormatFull
func (f *FullFormatter) Level() FormatLevel {
	return FormatFull
}

// removeEmptyFields removes empty/null fields from a response
func removeEmptyFields(response any) (any, error) {
	// Convert to map for manipulation
	data, err := toMap(response)
	if err != nil {
		return response, nil // Return unchanged if conversion fails
	}
	return cleanMap(data), nil
}

// toMap converts any value to a map via JSON marshaling
func toMap(v any) (map[string]any, error) {
	if m, ok := v.(map[string]any); ok {
		return m, nil
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// cleanMap recursively removes empty/null values from a map
func cleanMap(m map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range m {
		if v == nil {
			continue
		}
		if isEmpty(v) {
			continue
		}
		// Recursively clean nested maps
		if nested, ok := v.(map[string]any); ok {
			cleaned := cleanMap(nested)
			if len(cleaned) > 0 {
				result[k] = cleaned
			}
			continue
		}
		// Recursively clean arrays of maps
		if arr, ok := v.([]any); ok {
			cleaned := cleanArray(arr)
			if len(cleaned) > 0 {
				result[k] = cleaned
			}
			continue
		}
		result[k] = v
	}
	return result
}

// cleanArray recursively cleans arrays
func cleanArray(arr []any) []any {
	result := make([]any, 0, len(arr))
	for _, v := range arr {
		if v == nil {
			continue
		}
		if isEmpty(v) {
			continue
		}
		if nested, ok := v.(map[string]any); ok {
			cleaned := cleanMap(nested)
			if len(cleaned) > 0 {
				result = append(result, cleaned)
			}
			continue
		}
		result = append(result, v)
	}
	return result
}

// isEmpty checks if a value should be considered empty
func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.String:
		return val.Len() == 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	}
	return false
}

// truncateArrays truncates arrays longer than maxLen
func truncateArrays(data map[string]any, maxLen int) map[string]any {
	if maxLen <= 0 {
		return data
	}

	result := make(map[string]any)
	for k, v := range data {
		if arr, ok := v.([]any); ok && len(arr) > maxLen {
			result[k] = arr[:maxLen]
			continue
		}
		if nested, ok := v.(map[string]any); ok {
			result[k] = truncateArrays(nested, maxLen)
			continue
		}
		result[k] = v
	}
	return result
}
