package server

import (
	"context"
	"reflect"
	"testing"
)

func TestGetStringField(t *testing.T) {
	value := getStringField(map[string]interface{}{"content": "hello"}, "content")
	if value != "hello" {
		t.Fatalf("expected hello, got %s", value)
	}

	empty := getStringField(map[string]interface{}{"content": 42}, "content")
	if empty != "" {
		t.Fatalf("expected empty string for non-string value, got %s", empty)
	}
}

func TestGetFloatField(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{name: "float64", input: 1.25, want: 1.25},
		{name: "float32", input: float32(2.5), want: 2.5},
		{name: "int", input: 3, want: 3},
		{name: "int64", input: int64(4), want: 4},
		{name: "invalid", input: "oops", want: 0},
	}

	for _, tc := range tests {
		got := getFloatField(map[string]interface{}{"value": tc.input}, "value")
		if got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}

func TestGetBoolField(t *testing.T) {
	value := getBoolField(map[string]interface{}{"flag": true}, "flag")
	if !value {
		t.Fatal("expected true value")
	}

	missing := getBoolField(map[string]interface{}{}, "flag")
	if missing {
		t.Fatal("expected false for missing key")
	}
}

func TestGetStringSliceField(t *testing.T) {
	fromInterface := getStringSliceField(map[string]interface{}{"values": []interface{}{"a", "b"}}, "values")
	if !reflect.DeepEqual(fromInterface, []string{"a", "b"}) {
		t.Fatalf("expected slice from interface, got %v", fromInterface)
	}

	fromStrings := getStringSliceField(map[string]interface{}{"values": []string{"x", "y"}}, "values")
	if !reflect.DeepEqual(fromStrings, []string{"x", "y"}) {
		t.Fatalf("expected slice from strings, got %v", fromStrings)
	}

	missing := getStringSliceField(map[string]interface{}{}, "values")
	if missing != nil {
		t.Fatalf("expected nil for missing key, got %v", missing)
	}
}

func TestGetCriteria(t *testing.T) {
	input := map[string]interface{}{
		"criteria": []interface{}{
			map[string]interface{}{"name": "impact"},
			map[string]interface{}{"name": "cost"},
		},
	}

	got := _getCriteria(input)
	if len(got) != 2 {
		t.Fatalf("expected two criteria, got %d", len(got))
	}

	if got[0]["name"] != "impact" || got[1]["name"] != "cost" {
		t.Fatalf("unexpected criteria contents: %v", got)
	}
}

func TestExecuteToolAssessEvidence(t *testing.T) {
	executor := &serverToolExecutor{}
	payload := map[string]interface{}{
		"claim":    "claim",
		"evidence": []interface{}{"a", "b"},
	}

	result, err := executor.ExecuteTool(context.Background(), "assess-evidence", payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	if data["claim"] != "claim" {
		t.Fatalf("unexpected claim: %v", data["claim"])
	}

	evidence, ok := data["evidence"].([]string)
	if !ok {
		t.Fatalf("expected evidence slice, got %T", data["evidence"])
	}

	if !reflect.DeepEqual(evidence, []string{"a", "b"}) {
		t.Fatalf("unexpected evidence values: %v", evidence)
	}

	if data["strength"] != 0.7 {
		t.Fatalf("unexpected strength: %v", data["strength"])
	}
}

func TestExecuteToolMakeDecision(t *testing.T) {
	executor := &serverToolExecutor{}
	payload := map[string]interface{}{
		"situation": "choose path",
		"options":   []string{"optA", "optB"},
	}

	result, err := executor.ExecuteTool(context.Background(), "make-decision", payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	if data["selected"] != "optA" {
		t.Fatalf("expected first option selected, got %v", data["selected"])
	}
}

func TestExecuteToolUnsupported(t *testing.T) {
	executor := &serverToolExecutor{}
	_, err := executor.ExecuteTool(context.Background(), "unknown-tool", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for unsupported tool")
	}
}
