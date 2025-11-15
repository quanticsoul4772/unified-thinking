package handlers

import (
	"strings"
	"testing"
)

// TestUnmarshalParams tests the unmarshalParams helper function
func TestUnmarshalParams(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"name":  "test",
				"value": 42,
			},
			wantErr: false,
		},
		{
			name:    "empty params",
			params:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "numeric value as float",
			params: map[string]interface{}{
				"name":  "test",
				"value": 42.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := unmarshalParams(tt.params, &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateID tests the generateID helper function
func TestGenerateID(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{name: "simple prefix", prefix: "test"},
		{name: "empty prefix", prefix: ""},
		{name: "observation prefix", prefix: "obs"},
		{name: "hypothesis prefix", prefix: "hyp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := generateID(tt.prefix)

			if id == "" {
				t.Error("generateID() should return non-empty ID")
			}

			if tt.prefix != "" && !strings.HasPrefix(id, tt.prefix+"-") {
				t.Errorf("generateID() = %v, should have prefix %v", id, tt.prefix)
			}

			// IDs are based on nanosecond timestamps, so they should be unique
			// We just verify the format is correct
		})
	}
}
