package handlers

import (
	"testing"
	"time"
)

func TestUnmarshalParams(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		target  interface{}
		wantErr bool
	}{
		{
			name: "valid simple params",
			params: map[string]interface{}{
				"name":  "test",
				"value": 42,
			},
			target: &struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{},
			wantErr: false,
		},
		{
			name: "valid nested params",
			params: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			target: &struct {
				User struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				} `json:"user"`
			}{},
			wantErr: false,
		},
		{
			name: "valid array params",
			params: map[string]interface{}{
				"items": []interface{}{"item1", "item2", "item3"},
			},
			target: &struct {
				Items []string `json:"items"`
			}{},
			wantErr: false,
		},
		{
			name: "invalid json structure",
			params: map[string]interface{}{
				"name": func() {}, // invalid type
			},
			target:  &struct{ Name string }{},
			wantErr: true,
		},
		{
			name:    "empty params",
			params:  map[string]interface{}{},
			target:  &struct{}{},
			wantErr: false,
		},
		{
			name: "nil params",
			params: map[string]interface{}{
				"value": nil,
			},
			target: &struct {
				Value *string `json:"value"`
			}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := unmarshalParams(tt.params, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	// Test multiple calls to ensure uniqueness
	ids := make(map[string]bool)
	prefix := "test"

	for i := 0; i < 100; i++ {
		id := generateID(prefix)

		// Check ID starts with prefix
		if len(id) <= len(prefix) || id[:len(prefix)] != prefix {
			t.Errorf("generateID() = %v, should start with prefix %v", id, prefix)
		}

		// Check uniqueness
		if ids[id] {
			t.Errorf("generateID() produced duplicate ID: %v", id)
		}
		ids[id] = true

		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}
}

func TestGenerateID_DifferentPrefixes(t *testing.T) {
	prefixes := []string{"user", "order", "session", "temp"}

	for _, prefix := range prefixes {
		id := generateID(prefix)
		if len(id) <= len(prefix) || id[:len(prefix)] != prefix {
			t.Errorf("generateID() with prefix %v = %v, should start with prefix", prefix, id)
		}
	}
}

func TestGenerateID_EmptyPrefix(t *testing.T) {
	id := generateID("")
	if id == "" {
		t.Error("generateID() with empty prefix should not return empty string")
	}
}

func TestGenerateID_LongPrefix(t *testing.T) {
	longPrefix := "this-is-a-very-long-prefix-for-testing-purposes"
	id := generateID(longPrefix)
	if len(id) <= len(longPrefix) || id[:len(longPrefix)] != longPrefix {
		t.Errorf("generateID() with long prefix failed, got: %v", id)
	}
}

func TestUnmarshalParams_TargetNil(t *testing.T) {
	params := map[string]interface{}{
		"name": "test",
	}

	err := unmarshalParams(params, nil)
	if err == nil {
		t.Error("Expected error when target is nil")
	}
}

func TestUnmarshalParams_InvalidTargetType(t *testing.T) {
	params := map[string]interface{}{
		"name": "test",
	}

	// Pass a non-pointer target
	err := unmarshalParams(params, "not a pointer")
	if err == nil {
		t.Error("Expected error when target is not a pointer")
	}
}
