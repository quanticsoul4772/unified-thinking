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
