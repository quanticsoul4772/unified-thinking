package extraction

import (
	"context"
	"testing"
)

// TestRegexExtractor_URLs tests URL extraction
func TestRegexExtractor_URLs(t *testing.T) {
	extractor := NewRegexExtractor()

	content := "Check https://example.com and http://test.org for details"

	result, err := extractor.Extract(content)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Filter for URLs only
	urlCount := 0
	for _, entity := range result.Entities {
		if entity.Type == "url" {
			urlCount++
			if entity.Method != "regex" {
				t.Errorf("Expected method 'regex', got '%s'", entity.Method)
			}
			if entity.Confidence < 0.9 {
				t.Errorf("Expected high confidence for URLs, got %.2f", entity.Confidence)
			}
		}
	}

	if urlCount < 2 {
		t.Errorf("Expected at least 2 URLs, got %d", urlCount)
	}
}

// TestRegexExtractor_Emails tests email extraction
func TestRegexExtractor_Emails(t *testing.T) {
	extractor := NewRegexExtractor()

	content := "Contact user@example.com or admin@test.org for support"

	result, err := extractor.Extract(content)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	emailCount := 0
	for _, entity := range result.Entities {
		if entity.Type == "email" {
			emailCount++
		}
	}

	if emailCount < 2 {
		t.Errorf("Expected at least 2 emails, got %d", emailCount)
	}
}

// TestRegexExtractor_FilePaths tests file path extraction
func TestRegexExtractor_FilePaths(t *testing.T) {
	extractor := NewRegexExtractor()

	content := "Edit C:\\Users\\test\\file.txt or /home/user/script.py for configuration"

	result, err := extractor.Extract(content)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	fileCount := 0
	for _, entity := range result.Entities {
		if entity.Type == "file_path" {
			fileCount++
			if entity.Method != "regex" {
				t.Errorf("Expected method 'regex', got '%s'", entity.Method)
			}
		}
	}

	if fileCount < 2 {
		t.Errorf("Expected at least 2 file paths, got %d", fileCount)
	}
}

// TestRegexExtractor_Dates tests date extraction
func TestRegexExtractor_Dates(t *testing.T) {
	extractor := NewRegexExtractor()

	content := "Events on 2025-01-15 and 2025-12-31 are important"

	result, err := extractor.Extract(content)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	dateCount := 0
	for _, entity := range result.Entities {
		if entity.Type == "date" {
			dateCount++
		}
	}

	if dateCount < 2 {
		t.Errorf("Expected at least 2 dates, got %d", dateCount)
	}
}

// TestRegexExtractor_CausalRelationships tests causal relationship extraction
func TestRegexExtractor_CausalRelationships(t *testing.T) {
	extractor := NewRegexExtractor()

	tests := []struct {
		name     string
		content  string
		expected int
		relType  string
	}{
		{
			name:     "causes relationship",
			content:  "Database optimization causes better performance",
			expected: 1,
			relType:  "CAUSES",
		},
		{
			name:     "enables relationship",
			content:  "Caching enables faster response times",
			expected: 1,
			relType:  "ENABLES",
		},
		{
			name:     "leads to relationship",
			content:  "Poor indexing leads to slow queries",
			expected: 1,
			relType:  "CAUSES",
		},
		{
			name:     "contradicts relationship",
			content:  "This approach contradicts our security policy",
			expected: 1,
			relType:  "CONTRADICTS",
		},
		{
			name:     "builds upon relationship",
			content:  "The new feature builds upon existing authentication",
			expected: 1,
			relType:  "BUILDS_UPON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rels := extractor.ExtractCausalRelationships(tt.content)

			if len(rels) < tt.expected {
				t.Errorf("Expected at least %d relationships, got %d", tt.expected, len(rels))
			}

			if len(rels) > 0 && rels[0].Type != tt.relType {
				t.Errorf("Expected type %s, got %s", tt.relType, rels[0].Type)
			}
		})
	}
}

// TestHybridExtractor_RegexOnly tests extraction with LLM disabled
func TestHybridExtractor_RegexOnly(t *testing.T) {
	cfg := HybridConfig{
		EnableLLM: false,
	}
	extractor := NewHybridExtractor(cfg)

	content := "Visit https://example.com and contact admin@test.org for help with file.txt"

	result, err := extractor.Extract(content)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if result.TotalEntities < 3 {
		t.Errorf("Expected at least 3 entities (URL, email, file), got %d", result.TotalEntities)
	}

	if result.RegexEntities == 0 {
		t.Error("Expected regex entities")
	}

	if result.LLMEntities != 0 {
		t.Errorf("Expected no LLM entities with LLM disabled, got %d", result.LLMEntities)
	}
}

// TestHybridExtractor_WithLLM tests extraction with LLM enabled
func TestHybridExtractor_WithLLM(t *testing.T) {
	cfg := HybridConfig{
		EnableLLM: true,
	}
	extractor := NewHybridExtractor(cfg)

	content := "The optimization strategy enables better performance through caching"

	result, err := extractor.Extract(content)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should have entities from regex
	if result.RegexEntities == 0 {
		t.Log("Note: No regex entities found (expected for this content)")
	}

	// LLM extractor is currently a placeholder, so LLMEntities will be 0
	// In production with real LLM, this would extract concepts like "optimization", "performance", "caching"
}

// TestHybridExtractor_ExtractWithContext tests context-aware extraction
func TestHybridExtractor_ExtractWithContext(t *testing.T) {
	cfg := HybridConfig{EnableLLM: false}
	extractor := NewHybridExtractor(cfg)

	ctx := context.Background()
	content := "Test content with https://example.com"

	result, err := extractor.ExtractWithContext(ctx, content)
	if err != nil {
		t.Fatalf("ExtractWithContext failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}
}

// TestHybridExtractor_ExtractRelationships tests relationship-only extraction
func TestHybridExtractor_ExtractRelationships(t *testing.T) {
	cfg := HybridConfig{EnableLLM: false}
	extractor := NewHybridExtractor(cfg)

	content := "Indexing enables better query performance which causes faster response times"

	rels := extractor.ExtractRelationships(content)

	if len(rels) < 2 {
		t.Errorf("Expected at least 2 relationships, got %d", len(rels))
	}

	// Verify relationship types
	for _, rel := range rels {
		if rel.Type == "" {
			t.Error("Relationship type should not be empty")
		}
		if rel.From == "" || rel.To == "" {
			t.Error("Relationship should have from and to entities")
		}
	}
}

// TestShouldUseLLM tests LLM decision heuristics
func TestShouldUseLLM(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		regexEntityCount int
		expected         bool
	}{
		{
			name:             "short content",
			content:          "Short text",
			regexEntityCount: 0,
			expected:         false,
		},
		{
			name:             "many regex matches",
			content:          "Long content with many entities but regex already found 10 of them",
			regexEntityCount: 10,
			expected:         false,
		},
		{
			name:             "complex reasoning with few matches",
			content:          "This approach enables better performance because it causes less overhead, although it contradicts the previous design which suggested a different path",
			regexEntityCount: 1,
			expected:         true,
		},
		{
			name:             "simple statement",
			content:          "The system works well and provides good results for users without issues",
			regexEntityCount: 0,
			expected:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldUseLLM(tt.content, tt.regexEntityCount)
			if result != tt.expected {
				t.Errorf("ShouldUseLLM = %v, want %v for: %s", result, tt.expected, tt.content[:50])
			}
		})
	}
}
