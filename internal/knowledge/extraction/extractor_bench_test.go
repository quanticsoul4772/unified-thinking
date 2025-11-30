package extraction

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkRegexExtractor_Extract benchmarks regex extraction
func BenchmarkRegexExtractor_Extract(b *testing.B) {
	extractor := NewRegexExtractor()

	content := `
The system at https://example.com uses database optimization techniques to improve performance.
Contact admin@example.com for details about the config.json file located at C:\Program Files\app\config.json.
The deployment on 2025-01-15 at 14:30:00 improved response time to 150ms (from 500ms).
Version 1.2.3 includes UUID tracking with IDs like 550e8400-e29b-41d4-a716-446655440000.
The server at 192.168.1.100 handles requests efficiently.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.Extract(content)
		if err != nil {
			b.Fatalf("Extract failed: %v", err)
		}
	}
}

// BenchmarkRegexExtractor_CausalRelationships benchmarks relationship extraction
func BenchmarkRegexExtractor_CausalRelationships(b *testing.B) {
	extractor := NewRegexExtractor()

	content := `
Database optimization causes better performance which enables faster response times.
Poor indexing leads to slow queries that contradict our performance targets.
The new caching layer builds upon the existing architecture.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractor.ExtractCausalRelationships(content)
	}
}

// BenchmarkHybridExtractor_RegexOnly benchmarks hybrid with regex only
func BenchmarkHybridExtractor_RegexOnly(b *testing.B) {
	cfg := HybridConfig{EnableLLM: false}
	extractor := NewHybridExtractor(cfg)

	content := "Visit https://example.com and edit /path/to/file.txt on 2025-01-15"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.Extract(content)
		if err != nil {
			b.Fatalf("Extract failed: %v", err)
		}
	}
}

// BenchmarkHybridExtractor_LargeContent benchmarks extraction on large content
func BenchmarkHybridExtractor_LargeContent(b *testing.B) {
	cfg := HybridConfig{EnableLLM: false}
	extractor := NewHybridExtractor(cfg)

	// Generate large content with mixed entities
	var contentBuilder strings.Builder
	for i := 0; i < 100; i++ {
		contentBuilder.WriteString(fmt.Sprintf("Visit https://example%d.com ", i))
		contentBuilder.WriteString(fmt.Sprintf("contact user%d@test.org ", i))
		contentBuilder.WriteString(fmt.Sprintf("on 2025-01-%02d ", (i%28)+1))
		contentBuilder.WriteString("for optimization details. ")
	}
	content := contentBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.Extract(content)
		if err != nil {
			b.Fatalf("Extract failed: %v", err)
		}
	}
}

// BenchmarkShouldUseLLM benchmarks LLM decision heuristic
func BenchmarkShouldUseLLM(b *testing.B) {
	content := "This complex reasoning approach enables better performance because it causes less overhead, although it contradicts previous designs"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ShouldUseLLM(content, 2)
	}
}
