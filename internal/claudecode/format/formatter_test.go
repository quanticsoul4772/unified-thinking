package format

import (
	"encoding/json"
	"testing"
)

func TestFormatLevelValues(t *testing.T) {
	tests := []struct {
		level    FormatLevel
		expected string
	}{
		{FormatFull, "full"},
		{FormatCompact, "compact"},
		{FormatMinimal, "minimal"},
	}

	for _, tt := range tests {
		if string(tt.level) != tt.expected {
			t.Errorf("FormatLevel %v: got %s, want %s", tt.level, string(tt.level), tt.expected)
		}
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Level != FormatFull {
		t.Errorf("DefaultOptions Level: got %v, want %v", opts.Level, FormatFull)
	}
	if !opts.IncludeMetadata {
		t.Error("DefaultOptions IncludeMetadata should be true")
	}
	if !opts.IncludeTimings {
		t.Error("DefaultOptions IncludeTimings should be true")
	}
	if opts.MaxArrayLength != 0 {
		t.Errorf("DefaultOptions MaxArrayLength: got %d, want 0", opts.MaxArrayLength)
	}
	if opts.OmitEmpty {
		t.Error("DefaultOptions OmitEmpty should be false")
	}
}

func TestCompactOptions(t *testing.T) {
	opts := CompactOptions()

	if opts.Level != FormatCompact {
		t.Errorf("CompactOptions Level: got %v, want %v", opts.Level, FormatCompact)
	}
	if !opts.IncludeMetadata {
		t.Error("CompactOptions IncludeMetadata should be true")
	}
	if opts.IncludeTimings {
		t.Error("CompactOptions IncludeTimings should be false")
	}
	if opts.MaxArrayLength != 5 {
		t.Errorf("CompactOptions MaxArrayLength: got %d, want 5", opts.MaxArrayLength)
	}
	if !opts.OmitEmpty {
		t.Error("CompactOptions OmitEmpty should be true")
	}
	if !opts.FlattenNextTools {
		t.Error("CompactOptions FlattenNextTools should be true")
	}
}

func TestMinimalOptions(t *testing.T) {
	opts := MinimalOptions()

	if opts.Level != FormatMinimal {
		t.Errorf("MinimalOptions Level: got %v, want %v", opts.Level, FormatMinimal)
	}
	if opts.IncludeMetadata {
		t.Error("MinimalOptions IncludeMetadata should be false")
	}
	if opts.IncludeTimings {
		t.Error("MinimalOptions IncludeTimings should be false")
	}
	if opts.MaxArrayLength != 3 {
		t.Errorf("MinimalOptions MaxArrayLength: got %d, want 3", opts.MaxArrayLength)
	}
	if !opts.OmitEmpty {
		t.Error("MinimalOptions OmitEmpty should be true")
	}
	// MinimalOptions does NOT flatten next tools (only compact does)
	if opts.FlattenNextTools {
		t.Error("MinimalOptions FlattenNextTools should be false")
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		level         FormatLevel
		expectedLevel FormatLevel
	}{
		{FormatFull, FormatFull},
		{FormatCompact, FormatCompact},
		{FormatMinimal, FormatMinimal},
		{"unknown", FormatFull}, // default to full
	}

	for _, tt := range tests {
		var opts FormatOptions
		switch tt.level {
		case FormatFull:
			opts = DefaultOptions()
		case FormatCompact:
			opts = CompactOptions()
		case FormatMinimal:
			opts = MinimalOptions()
		default:
			opts = DefaultOptions()
		}
		formatter := NewFormatter(tt.level, opts)
		if formatter.Level() != tt.expectedLevel {
			t.Errorf("NewFormatter(%v).Level(): got %v, want %v", tt.level, formatter.Level(), tt.expectedLevel)
		}
	}
}

func TestFullFormatterFormat(t *testing.T) {
	formatter := &FullFormatter{opts: DefaultOptions()}

	input := map[string]any{
		"id":      "test-123",
		"content": "test content",
		"metadata": map[string]any{
			"key": "value",
		},
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("FullFormatter.Format() error: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("FullFormatter.Format() result should be a map")
	}

	if resultMap["id"] != "test-123" {
		t.Errorf("Expected id 'test-123', got %v", resultMap["id"])
	}
}

func TestCompactFormatterFormat(t *testing.T) {
	formatter := &CompactFormatter{opts: CompactOptions()}

	input := map[string]any{
		"id":       "test-123",
		"content":  "test content",
		"metadata": map[string]any{"key": "value"},
		"empty":    "",
		"items":    []any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("CompactFormatter.Format() error: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("CompactFormatter.Format() result should be a map")
	}

	// CompactOptions has IncludeMetadata=true, so metadata should be preserved
	// (it only removes suggested_next_tools from metadata via FlattenNextTools)
	if _, exists := resultMap["metadata"]; !exists {
		t.Error("CompactFormatter with IncludeMetadata=true should keep metadata")
	}

	// Should truncate arrays to MaxArrayLength=5
	if items, ok := resultMap["items"].([]any); ok {
		if len(items) > 5 {
			t.Errorf("CompactFormatter should truncate arrays to 5 items, got %d", len(items))
		}
	}
}

func TestMinimalFormatterFormat(t *testing.T) {
	formatter := &MinimalFormatter{opts: MinimalOptions()}

	// MinimalFormatter extracts only essential fields based on detected response type
	// Since this input has "id" field, it will be detected as "default" type
	// and only keep: id, status, confidence, result
	input := map[string]any{
		"id":         "test-123",
		"content":    "test content",
		"status":     "ok",
		"confidence": 0.9,
		"metadata":   map[string]any{"key": "value"},
		"items":      []any{1, 2, 3, 4, 5},
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("MinimalFormatter.Format() error: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("MinimalFormatter.Format() result should be a map")
	}

	// MinimalFormatter keeps only essential fields for detected type
	// For "default" type: id, status, confidence, result
	if _, exists := resultMap["id"]; !exists {
		t.Error("MinimalFormatter should keep 'id' field")
	}
	if _, exists := resultMap["status"]; !exists {
		t.Error("MinimalFormatter should keep 'status' field")
	}
	if _, exists := resultMap["confidence"]; !exists {
		t.Error("MinimalFormatter should keep 'confidence' field")
	}

	// Should NOT have metadata (not in essential fields)
	if _, exists := resultMap["metadata"]; exists {
		t.Error("MinimalFormatter should not include metadata")
	}

	// Should NOT have content (not in essential fields for default type)
	if _, exists := resultMap["content"]; exists {
		t.Error("MinimalFormatter should not include content for default type")
	}
}

func TestFormatterJSONOutput(t *testing.T) {
	formatters := []ResponseFormatter{
		&FullFormatter{opts: DefaultOptions()},
		&CompactFormatter{opts: CompactOptions()},
		&MinimalFormatter{opts: MinimalOptions()},
	}

	input := map[string]any{
		"id":      "test",
		"content": "test content",
	}

	for _, formatter := range formatters {
		result, err := formatter.Format(input)
		if err != nil {
			t.Errorf("%v.Format() error: %v", formatter.Level(), err)
			continue
		}

		// Ensure result is JSON-serializable
		_, err = json.Marshal(result)
		if err != nil {
			t.Errorf("%v.Format() result is not JSON-serializable: %v", formatter.Level(), err)
		}
	}
}

func TestCompactFormatterOmitEmpty(t *testing.T) {
	formatter := &CompactFormatter{opts: CompactOptions()}

	input := map[string]any{
		"id":         "test",
		"content":    "has content",
		"empty":      "",
		"null_value": nil,
		"zero":       0,
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("CompactFormatter.Format() error: %v", err)
	}

	resultMap := result.(map[string]any)

	// Empty string should be omitted
	if _, exists := resultMap["empty"]; exists {
		t.Error("CompactFormatter should omit empty strings")
	}

	// nil should be omitted
	if _, exists := resultMap["null_value"]; exists {
		t.Error("CompactFormatter should omit nil values")
	}

	// Zero should NOT be omitted (only empty strings and nil)
	if _, exists := resultMap["zero"]; !exists {
		t.Error("CompactFormatter should keep zero values")
	}
}

func TestFullFormatterLevel(t *testing.T) {
	formatter := &FullFormatter{opts: DefaultOptions()}
	if formatter.Level() != FormatFull {
		t.Errorf("FullFormatter.Level() should return FormatFull, got %v", formatter.Level())
	}
}

func TestCompactFormatterLevel(t *testing.T) {
	formatter := &CompactFormatter{opts: CompactOptions()}
	if formatter.Level() != FormatCompact {
		t.Errorf("CompactFormatter.Level() should return FormatCompact, got %v", formatter.Level())
	}
}

func TestMinimalFormatterLevel(t *testing.T) {
	formatter := &MinimalFormatter{opts: MinimalOptions()}
	if formatter.Level() != FormatMinimal {
		t.Errorf("MinimalFormatter.Level() should return FormatMinimal, got %v", formatter.Level())
	}
}

func TestTruncateArrays(t *testing.T) {
	data := map[string]any{
		"short":  []any{1, 2},
		"long":   []any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"nested": map[string]any{"arr": []any{1, 2, 3, 4, 5}},
	}

	result := truncateArrays(data, 3)

	short := result["short"].([]any)
	if len(short) != 2 {
		t.Errorf("Short array should not be truncated, got %d", len(short))
	}

	long := result["long"].([]any)
	if len(long) != 3 {
		t.Errorf("Long array should be truncated to 3, got %d", len(long))
	}

	nested := result["nested"].(map[string]any)
	nestedArr := nested["arr"].([]any)
	if len(nestedArr) != 3 {
		t.Errorf("Nested array should be truncated to 3, got %d", len(nestedArr))
	}
}

func TestTruncateArraysZeroMax(t *testing.T) {
	data := map[string]any{
		"arr": []any{1, 2, 3},
	}

	result := truncateArrays(data, 0)

	// Should return unchanged
	arr := result["arr"].([]any)
	if len(arr) != 3 {
		t.Errorf("Array should not be truncated when maxLen=0, got %d", len(arr))
	}
}

// TestIsValid tests the FormatLevel.IsValid method
func TestIsValid(t *testing.T) {
	tests := []struct {
		level FormatLevel
		valid bool
	}{
		{FormatFull, true},
		{FormatCompact, true},
		{FormatMinimal, true},
		{"unknown", false},
		{"", false},
		{"FULL", false}, // case-sensitive
	}

	for _, tt := range tests {
		if got := tt.level.IsValid(); got != tt.valid {
			t.Errorf("FormatLevel(%q).IsValid() = %v, want %v", tt.level, got, tt.valid)
		}
	}
}

// TestParseFormatLevel tests the ParseFormatLevel function
func TestParseFormatLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected FormatLevel
	}{
		{"full", FormatFull},
		{"compact", FormatCompact},
		{"minimal", FormatMinimal},
		{"", FormatFull},      // default
		{"unknown", FormatFull}, // default
		{"COMPACT", FormatFull}, // case-sensitive, defaults to full
	}

	for _, tt := range tests {
		if got := ParseFormatLevel(tt.input); got != tt.expected {
			t.Errorf("ParseFormatLevel(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

// TestRemoveEmptyFields tests the removeEmptyFields function
func TestRemoveEmptyFields(t *testing.T) {
	input := map[string]any{
		"id":      "test",
		"empty":   "",
		"null":    nil,
		"arr":     []any{},
		"nested":  map[string]any{"key": "value", "empty": ""},
	}

	result, err := removeEmptyFields(input)
	if err != nil {
		t.Fatalf("removeEmptyFields() error = %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("removeEmptyFields() should return a map")
	}

	// Should have id
	if _, exists := resultMap["id"]; !exists {
		t.Error("removeEmptyFields should keep 'id'")
	}

	// Should not have empty
	if _, exists := resultMap["empty"]; exists {
		t.Error("removeEmptyFields should remove empty strings")
	}

	// Should not have null
	if _, exists := resultMap["null"]; exists {
		t.Error("removeEmptyFields should remove nil values")
	}

	// Should not have empty array
	if _, exists := resultMap["arr"]; exists {
		t.Error("removeEmptyFields should remove empty arrays")
	}
}

// TestFullFormatterWithOmitEmpty tests FullFormatter with OmitEmpty option
func TestFullFormatterWithOmitEmpty(t *testing.T) {
	opts := DefaultOptions()
	opts.OmitEmpty = true
	formatter := &FullFormatter{opts: opts}

	input := map[string]any{
		"id":    "test",
		"empty": "",
		"value": "content",
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("FullFormatter.Format() error = %v", err)
	}

	resultMap := result.(map[string]any)

	// Should not have empty string
	if _, exists := resultMap["empty"]; exists {
		t.Error("FullFormatter with OmitEmpty should remove empty strings")
	}

	// Should have non-empty fields
	if _, exists := resultMap["id"]; !exists {
		t.Error("FullFormatter with OmitEmpty should keep non-empty fields")
	}
}

// TestDetectResponseType tests the detectResponseType function
func TestDetectResponseType(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected string
	}{
		{
			name:     "think response",
			data:     map[string]any{"thought_id": "t1", "content": "test"},
			expected: "think",
		},
		{
			name:     "branch response",
			data:     map[string]any{"branch_id": "b1", "state": "active"},
			expected: "branch",
		},
		{
			name:     "decision response",
			data:     map[string]any{"decision_id": "d1", "selected_option": "A"},
			expected: "decision",
		},
		{
			name:     "session response",
			data:     map[string]any{"session_id": "s1", "status": "active"},
			expected: "session",
		},
		{
			name:     "preset response",
			data:     map[string]any{"preset_id": "p1", "status": "running"},
			expected: "preset",
		},
		{
			name:     "validation response",
			data:     map[string]any{"is_valid": true, "reason": "ok"},
			expected: "validation",
		},
		{
			name:     "search response",
			data:     map[string]any{"results": []any{}, "count": 0},
			expected: "search",
		},
		{
			name:     "default response",
			data:     map[string]any{"id": "x", "value": 123},
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectResponseType(tt.data)
			if got != tt.expected {
				t.Errorf("detectResponseType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSimplifyResponse tests the simplifyResponse function
func TestSimplifyResponse(t *testing.T) {
	t.Run("with priority fields", func(t *testing.T) {
		data := map[string]any{
			"id":         "test-123",
			"status":     "ok",
			"confidence": 0.9,
			"extra":      "ignored",
		}

		result := simplifyResponse(data)

		if result["id"] != "test-123" {
			t.Error("simplifyResponse should keep 'id'")
		}
		if result["status"] != "ok" {
			t.Error("simplifyResponse should keep 'status'")
		}
		if _, exists := result["extra"]; exists {
			t.Error("simplifyResponse should not include non-priority fields")
		}
	})

	t.Run("with nested map", func(t *testing.T) {
		data := map[string]any{
			"result": map[string]any{
				"id":   "nested-id",
				"name": "test-name",
			},
		}

		result := simplifyResponse(data)

		// Should have simplified nested result
		if _, exists := result["result"]; !exists {
			t.Error("simplifyResponse should keep 'result' field")
		}
	})

	t.Run("empty data returns status ok", func(t *testing.T) {
		data := map[string]any{}

		result := simplifyResponse(data)

		if result["status"] != "ok" {
			t.Errorf("simplifyResponse for empty data should return status=ok, got %v", result["status"])
		}
	})
}

// TestSimplifyNested tests the simplifyNested function
func TestSimplifyNested(t *testing.T) {
	t.Run("with id", func(t *testing.T) {
		data := map[string]any{"id": "test-id", "other": "value"}
		result := simplifyNested(data)
		if result != "test-id" {
			t.Errorf("simplifyNested with 'id' should return id, got %v", result)
		}
	})

	t.Run("with name", func(t *testing.T) {
		data := map[string]any{"name": "test-name", "other": "value"}
		result := simplifyNested(data)
		if result != "test-name" {
			t.Errorf("simplifyNested with 'name' should return name, got %v", result)
		}
	})

	t.Run("without id or name", func(t *testing.T) {
		data := map[string]any{"key1": "v1", "key2": "v2", "key3": "v3"}
		result := simplifyNested(data)
		if result != 3 {
			t.Errorf("simplifyNested without id/name should return field count, got %v", result)
		}
	})
}

// TestFlattenNextTools tests the flattenNextTools function
func TestFlattenNextTools(t *testing.T) {
	t.Run("with suggested_next_tools as map array", func(t *testing.T) {
		formatter := &CompactFormatter{opts: CompactOptions()}
		data := map[string]any{
			"id": "test",
			"metadata": map[string]any{
				"suggested_next_tools": []any{
					map[string]any{"tool": "think", "priority": "high"},
					map[string]any{"tool": "validate", "priority": "low"},
				},
				"action_recommendations": []any{"do this", "do that"},
				"export_formats":         []string{"json", "xml"},
				"other_field":            "preserved",
			},
		}

		result := formatter.flattenNextTools(data)

		// Should have flattened next_tools
		nextTools, exists := result["next_tools"]
		if !exists {
			t.Error("flattenNextTools should create 'next_tools' field")
		}
		toolsArr, ok := nextTools.([]string)
		if !ok {
			t.Fatalf("next_tools should be []string, got %T", nextTools)
		}
		if len(toolsArr) != 2 || toolsArr[0] != "think" || toolsArr[1] != "validate" {
			t.Errorf("next_tools = %v, want [think, validate]", toolsArr)
		}

		// Should have metadata with only other_field left
		metadata, hasMetadata := result["metadata"].(map[string]any)
		if !hasMetadata {
			t.Error("flattenNextTools should keep metadata with remaining fields")
		} else {
			if _, exists := metadata["action_recommendations"]; exists {
				t.Error("flattenNextTools should remove action_recommendations from metadata")
			}
			if _, exists := metadata["export_formats"]; exists {
				t.Error("flattenNextTools should remove export_formats from metadata")
			}
			if _, exists := metadata["other_field"]; !exists {
				t.Error("flattenNextTools should preserve other_field in metadata")
			}
		}
	})

	t.Run("with suggested_next_tools as string array", func(t *testing.T) {
		formatter := &CompactFormatter{opts: CompactOptions()}
		data := map[string]any{
			"metadata": map[string]any{
				"suggested_next_tools": []any{"think", "validate"},
			},
		}

		result := formatter.flattenNextTools(data)

		nextTools := result["next_tools"].([]string)
		if len(nextTools) != 2 {
			t.Errorf("next_tools should have 2 items, got %d", len(nextTools))
		}
	})

	t.Run("without metadata", func(t *testing.T) {
		formatter := &CompactFormatter{opts: CompactOptions()}
		data := map[string]any{"id": "test"}

		result := formatter.flattenNextTools(data)

		if _, exists := result["next_tools"]; exists {
			t.Error("flattenNextTools without metadata should not create next_tools")
		}
	})

	t.Run("empty metadata after cleanup", func(t *testing.T) {
		formatter := &CompactFormatter{opts: CompactOptions()}
		data := map[string]any{
			"id": "test",
			"metadata": map[string]any{
				"action_recommendations": []any{"rec1"},
				"export_formats":         []string{"json"},
			},
		}

		result := formatter.flattenNextTools(data)

		// Metadata should be removed entirely if empty after cleanup
		if _, exists := result["metadata"]; exists {
			t.Error("flattenNextTools should remove empty metadata")
		}
	})
}

// TestMinimalFormatterWithEmptyResult tests MinimalFormatter when result is empty
func TestMinimalFormatterWithEmptyResult(t *testing.T) {
	formatter := &MinimalFormatter{opts: MinimalOptions()}

	// Input that will produce empty essential fields
	input := map[string]any{
		"unknown_field": "value",
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("MinimalFormatter.Format() error = %v", err)
	}

	resultMap := result.(map[string]any)

	// Should return simplified response with at least "status": "ok"
	if _, exists := resultMap["status"]; !exists {
		t.Error("MinimalFormatter with empty result should return status field")
	}
}

// TestMinimalFormatterWithError tests MinimalFormatter preserves error info
func TestMinimalFormatterWithError(t *testing.T) {
	formatter := &MinimalFormatter{opts: MinimalOptions()}

	input := map[string]any{
		"error":      "something went wrong",
		"error_code": "ERR_001",
		"metadata":   map[string]any{"ignored": true},
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("MinimalFormatter.Format() error = %v", err)
	}

	resultMap := result.(map[string]any)

	if resultMap["error"] != "something went wrong" {
		t.Error("MinimalFormatter should preserve error message")
	}
	if resultMap["error_code"] != "ERR_001" {
		t.Error("MinimalFormatter should preserve error_code")
	}
}

// TestMinimalFormatterWithWrappedResult tests MinimalFormatter unwrapping
func TestMinimalFormatterWithWrappedResult(t *testing.T) {
	formatter := &MinimalFormatter{opts: MinimalOptions()}

	input := map[string]any{
		"result": map[string]any{
			"thought_id": "t123",
			"mode":       "linear",
			"confidence": 0.85,
			"metadata":   "ignored",
		},
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("MinimalFormatter.Format() error = %v", err)
	}

	resultMap := result.(map[string]any)

	// Should unwrap result and detect "think" type
	if resultMap["thought_id"] != "t123" {
		t.Error("MinimalFormatter should unwrap result and keep thought_id")
	}
	if _, exists := resultMap["metadata"]; exists {
		t.Error("MinimalFormatter should not include metadata for think type")
	}
}

// TestCompactFormatterWithWrappedResult tests CompactFormatter flattening wrapped results
func TestCompactFormatterWithWrappedResult(t *testing.T) {
	formatter := &CompactFormatter{opts: CompactOptions()}

	input := map[string]any{
		"result": map[string]any{
			"id":      "r123",
			"content": "test content",
		},
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("CompactFormatter.Format() error = %v", err)
	}

	resultMap := result.(map[string]any)

	// Should flatten result to top level
	if resultMap["id"] != "r123" {
		t.Error("CompactFormatter should flatten result.id to top level")
	}
	if resultMap["content"] != "test content" {
		t.Error("CompactFormatter should flatten result.content to top level")
	}
}

// TestCompactFormatterWithNonMapResult tests CompactFormatter with non-map result
func TestCompactFormatterWithNonMapResult(t *testing.T) {
	formatter := &CompactFormatter{opts: CompactOptions()}

	input := map[string]any{
		"result": "simple string result",
	}

	result, err := formatter.Format(input)
	if err != nil {
		t.Fatalf("CompactFormatter.Format() error = %v", err)
	}

	resultMap := result.(map[string]any)

	// Should keep non-map result as is
	if resultMap["result"] != "simple string result" {
		t.Errorf("CompactFormatter should keep non-map result, got %v", resultMap["result"])
	}
}

// TestCleanArrayWithNestedMaps tests cleanArray with nested map elements
func TestCleanArrayWithNestedMaps(t *testing.T) {
	arr := []any{
		map[string]any{"id": "1", "empty": ""},
		nil,
		"",
		map[string]any{},
		map[string]any{"name": "test"},
	}

	result := cleanArray(arr)

	// Should have only 2 elements (id map and name map)
	if len(result) != 2 {
		t.Errorf("cleanArray should return 2 non-empty elements, got %d", len(result))
	}
}

// TestToMapWithStruct tests toMap with a struct input
func TestToMapWithStruct(t *testing.T) {
	type testStruct struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}

	input := testStruct{ID: "t1", Content: "test"}

	result, err := toMap(input)
	if err != nil {
		t.Fatalf("toMap() error = %v", err)
	}

	if result["id"] != "t1" {
		t.Errorf("toMap should convert struct, got id=%v", result["id"])
	}
}

// TestIsEmptyWithPointer tests isEmpty with pointer/interface values
func TestIsEmptyWithPointer(t *testing.T) {
	var nilPtr *string = nil
	if !isEmpty(nilPtr) {
		t.Error("isEmpty should return true for nil pointer")
	}

	s := "test"
	ptr := &s
	if isEmpty(ptr) {
		t.Error("isEmpty should return false for non-nil pointer")
	}
}
