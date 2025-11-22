package storage

import (
	"fmt"
	"testing"

	"unified-thinking/internal/types"
)

// BenchmarkSearchThoughts_SmallResultSet tests search performance with small result sets
// This measures the impact of copyThoughtOptimized
func BenchmarkSearchThoughts_SmallResultSet(b *testing.B) {
	s := NewMemoryStorage()

	// Setup: 100 thoughts
	for i := 0; i < 100; i++ {
		thought := &types.Thought{
			Content: fmt.Sprintf("thought number %d with some content", i),
			Mode:    types.ModeLinear,
			Metadata: map[string]interface{}{
				"index": i,
				"type":  "test",
			},
			KeyPoints: []string{fmt.Sprintf("point-%d", i)},
		}
		_ = s.StoreThought(thought)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.SearchThoughts("thought", types.ModeLinear, 10, 0)
	}
}

// BenchmarkSearchThoughts_LargeResultSet tests search at scale
func BenchmarkSearchThoughts_LargeResultSet(b *testing.B) {
	s := NewMemoryStorage()

	// Setup: 1000 thoughts
	for i := 0; i < 1000; i++ {
		thought := &types.Thought{
			Content: fmt.Sprintf("thought number %d with searchable content", i),
			Mode:    types.ModeLinear,
			Metadata: map[string]interface{}{
				"index": i,
				"type":  "test",
			},
			KeyPoints: []string{fmt.Sprintf("point-%d", i)},
		}
		_ = s.StoreThought(thought)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.SearchThoughts("thought", types.ModeLinear, 100, 0)
	}
}

// BenchmarkIndexThoughtContent measures tokenization optimization
func BenchmarkIndexThoughtContent(b *testing.B) {
	s := NewMemoryStorage()
	thought := &types.Thought{
		Content: "This is a sample thought with multiple words for tokenization testing purposes including various punctuation marks and numbers like 123 and 456",
		Mode:    types.ModeLinear,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		thought.ID = "" // Reset ID to force new indexing
		_ = s.StoreThought(thought)
	}
}

// BenchmarkCopyThought compares original vs optimized copy
func BenchmarkCopyThought(b *testing.B) {
	thought := &types.Thought{
		Content:   "Test thought content",
		Mode:      types.ModeLinear,
		KeyPoints: []string{"point1", "point2", "point3"},
		Metadata: map[string]interface{}{
			"key1":   "value1",
			"key2":   42,
			"key3":   true,
			"nested": "not too deep",
		},
	}

	b.Run("Original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = copyThought(thought)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = copyThoughtOptimized(thought)
		}
	})
}

// BenchmarkLRUEviction measures eviction performance
func BenchmarkLRUEviction(b *testing.B) {
	s := NewMemoryStorage()

	// Fill index to near capacity
	for i := 0; i < MaxIndexSize-100; i++ {
		thought := &types.Thought{
			Content: fmt.Sprintf("unique word%d content", i),
			Mode:    types.ModeLinear,
		}
		_ = s.StoreThought(thought)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// This should trigger eviction
		thought := &types.Thought{
			Content: fmt.Sprintf("newword%d triggers eviction", i),
			Mode:    types.ModeLinear,
		}
		_ = s.StoreThought(thought)
	}
}

// BenchmarkSearchWithMapConversion measures map vs slice overhead
func BenchmarkSearchWithMapConversion(b *testing.B) {
	s := NewMemoryStorage()

	// Setup with varying result set sizes
	for i := 0; i < 500; i++ {
		thought := &types.Thought{
			Content: fmt.Sprintf("searchable content item %d", i),
			Mode:    types.ModeLinear,
		}
		_ = s.StoreThought(thought)
	}

	b.Run("Small-10", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = s.SearchThoughts("item", types.ModeLinear, 10, 0)
		}
	})

	b.Run("Medium-50", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = s.SearchThoughts("item", types.ModeLinear, 50, 0)
		}
	})

	b.Run("Large-200", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = s.SearchThoughts("item", types.ModeLinear, 200, 0)
		}
	})
}

// BenchmarkConcurrentReads measures read performance under concurrent load
func BenchmarkConcurrentReads(b *testing.B) {
	s := NewMemoryStorage()

	// Setup data
	for i := 0; i < 1000; i++ {
		thought := &types.Thought{
			Content: fmt.Sprintf("concurrent test thought %d", i),
			Mode:    types.ModeLinear,
		}
		_ = s.StoreThought(thought)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = s.SearchThoughts("test", types.ModeLinear, 10, 0)
		}
	})
}

// BenchmarkProductionScenario_Search simulates realistic search workload
func BenchmarkProductionScenario_Search(b *testing.B) {
	s := NewMemoryStorage()

	// Setup production-scale dataset
	for i := 0; i < 10000; i++ {
		thought := &types.Thought{
			Content: fmt.Sprintf("production thought %d with analysis problem solution decision reasoning", i),
			Mode:    types.ModeLinear,
			Metadata: map[string]interface{}{
				"session_id": i % 100,
				"step":       i % 10,
			},
		}
		_ = s.StoreThought(thought)
	}

	queries := []string{"problem", "solution", "analysis", "decision"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			_ = s.SearchThoughts(q, "", 20, 0)
		}
	}
}
