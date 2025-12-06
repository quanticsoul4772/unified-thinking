package knowledge

import (
	"testing"
)

// TestTruncate tests the truncate helper function
func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "needs truncation",
			input:  "hello world",
			maxLen: 5,
			want:   "hello...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "very long string",
			input:  "this is a very long string that needs to be truncated",
			maxLen: 20,
			want:   "this is a very long ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// TestMapExtractedType tests the entity type mapping function
func TestMapExtractedType(t *testing.T) {
	tests := []struct {
		name          string
		extractedType string
		want          EntityType
	}{
		{
			name:          "url type",
			extractedType: "url",
			want:          EntityTypeTool,
		},
		{
			name:          "file_path type",
			extractedType: "file_path",
			want:          EntityTypeTool,
		},
		{
			name:          "identifier type",
			extractedType: "identifier",
			want:          EntityTypeTool,
		},
		{
			name:          "email type",
			extractedType: "email",
			want:          EntityTypePerson,
		},
		{
			name:          "unknown type",
			extractedType: "unknown",
			want:          EntityTypeConcept,
		},
		{
			name:          "empty type",
			extractedType: "",
			want:          EntityTypeConcept,
		},
		{
			name:          "arbitrary type",
			extractedType: "foo",
			want:          EntityTypeConcept,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapExtractedType(tt.extractedType)
			if got != tt.want {
				t.Errorf("mapExtractedType(%q) = %v, want %v", tt.extractedType, got, tt.want)
			}
		})
	}
}

// TestMapRelationshipType tests the relationship type mapping function
func TestMapRelationshipType(t *testing.T) {
	tests := []struct {
		name          string
		extractedType string
		want          RelationshipType
	}{
		{
			name:          "CAUSES",
			extractedType: "CAUSES",
			want:          RelationshipCauses,
		},
		{
			name:          "ENABLES",
			extractedType: "ENABLES",
			want:          RelationshipEnables,
		},
		{
			name:          "CONTRADICTS",
			extractedType: "CONTRADICTS",
			want:          RelationshipContradicts,
		},
		{
			name:          "BUILDS_UPON",
			extractedType: "BUILDS_UPON",
			want:          RelationshipBuildsUpon,
		},
		{
			name:          "unknown type",
			extractedType: "UNKNOWN",
			want:          RelationshipRelatesTo,
		},
		{
			name:          "empty type",
			extractedType: "",
			want:          RelationshipRelatesTo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapRelationshipType(tt.extractedType)
			if got != tt.want {
				t.Errorf("mapRelationshipType(%q) = %v, want %v", tt.extractedType, got, tt.want)
			}
		})
	}
}
