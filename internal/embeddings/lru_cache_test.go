package embeddings

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// mustNewLRUEmbeddingCache creates a cache and fails the test if there's an error
func mustNewLRUEmbeddingCache(t *testing.T, config *LRUCacheConfig) *LRUEmbeddingCache {
	t.Helper()
	cache, err := NewLRUEmbeddingCache(config)
	if err != nil {
		t.Fatalf("failed to create LRU cache: %v", err)
	}
	return cache
}

// mustNewLRUEmbeddingCacheBenchmark creates a cache for benchmarks and fails if there's an error
func mustNewLRUEmbeddingCacheBenchmark(b *testing.B, config *LRUCacheConfig) *LRUEmbeddingCache {
	b.Helper()
	cache, err := NewLRUEmbeddingCache(config)
	if err != nil {
		b.Fatalf("failed to create LRU cache: %v", err)
	}
	return cache
}

func TestLRUEmbeddingCache_BasicOperations(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 100,
		TTL:        time.Hour,
	})
	defer cache.Close()

	// Test Set and Get
	embedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	cache.Set("hello world", embedding)

	got, ok := cache.Get("hello world")
	if !ok {
		t.Fatal("expected to find embedding")
	}

	if len(got) != len(embedding) {
		t.Fatalf("expected %d values, got %d", len(embedding), len(got))
	}

	for i := range embedding {
		if got[i] != embedding[i] {
			t.Errorf("embedding[%d]: expected %f, got %f", i, embedding[i], got[i])
		}
	}

	// Test miss
	_, ok = cache.Get("not found")
	if ok {
		t.Error("expected miss for unknown key")
	}
}

func TestLRUEmbeddingCache_LRUEviction(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 3,
		TTL:        time.Hour,
	})
	defer cache.Close()

	// Fill cache
	cache.Set("first", []float32{1.0})
	cache.Set("second", []float32{2.0})
	cache.Set("third", []float32{3.0})

	if cache.Size() != 3 {
		t.Fatalf("expected size 3, got %d", cache.Size())
	}

	// Access first to make it MRU
	cache.Get("first")

	// Add fourth - should evict second (LRU)
	cache.Set("fourth", []float32{4.0})

	if cache.Size() != 3 {
		t.Fatalf("expected size 3 after eviction, got %d", cache.Size())
	}

	// first should still exist (was accessed)
	_, ok := cache.Get("first")
	if !ok {
		t.Error("expected first to still exist")
	}

	// second should be evicted (LRU)
	_, ok = cache.Get("second")
	if ok {
		t.Error("expected second to be evicted")
	}

	// third should still exist
	_, ok = cache.Get("third")
	if !ok {
		t.Error("expected third to still exist")
	}

	// fourth should exist
	_, ok = cache.Get("fourth")
	if !ok {
		t.Error("expected fourth to exist")
	}

	// Check stats
	stats := cache.Stats()
	if stats["evictions"].(int64) != 1 {
		t.Errorf("expected 1 eviction, got %d", stats["evictions"])
	}
}

func TestLRUEmbeddingCache_TTLExpiry(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 100,
		TTL:        100 * time.Millisecond,
	})
	defer cache.Close()

	cache.Set("expires", []float32{1.0, 2.0})

	// Should exist immediately
	_, ok := cache.Get("expires")
	if !ok {
		t.Error("expected entry to exist before expiry")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get("expires")
	if ok {
		t.Error("expected entry to be expired")
	}

	stats := cache.Stats()
	if stats["expiries"].(int64) != 1 {
		t.Errorf("expected 1 expiry, got %d", stats["expiries"])
	}
}

func TestLRUEmbeddingCache_Update(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 100,
		TTL:        time.Hour,
	})
	defer cache.Close()

	// Set initial value
	cache.Set("key", []float32{1.0})

	// Update value
	cache.Set("key", []float32{2.0, 3.0})

	got, ok := cache.Get("key")
	if !ok {
		t.Fatal("expected to find updated entry")
	}

	if len(got) != 2 || got[0] != 2.0 || got[1] != 3.0 {
		t.Errorf("expected [2.0, 3.0], got %v", got)
	}

	// Size should still be 1
	if cache.Size() != 1 {
		t.Errorf("expected size 1, got %d", cache.Size())
	}
}

func TestLRUEmbeddingCache_Clear(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 100,
		TTL:        time.Hour,
	})
	defer cache.Close()

	cache.Set("one", []float32{1.0})
	cache.Set("two", []float32{2.0})
	cache.Set("three", []float32{3.0})

	if cache.Size() != 3 {
		t.Fatalf("expected size 3, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.Size())
	}

	_, ok := cache.Get("one")
	if ok {
		t.Error("expected empty cache after clear")
	}
}

func TestLRUEmbeddingCache_Persistence(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "test_cache.gob")

	// Create and populate cache
	cache1 := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries:    100,
		TTL:           time.Hour,
		PersistPath:   cachePath,
		CompressCache: true,
	})

	cache1.Set("persistent1", []float32{1.0, 1.1, 1.2})
	cache1.Set("persistent2", []float32{2.0, 2.1, 2.2})

	// Save and close
	if err := cache1.Close(); err != nil {
		t.Fatalf("failed to close cache: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file should exist after save")
	}

	// Create new cache and load
	cache2 := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries:    100,
		TTL:           time.Hour,
		PersistPath:   cachePath,
		CompressCache: true,
	})
	defer cache2.Close()

	// Verify loaded data
	got1, ok := cache2.Get("persistent1")
	if !ok {
		t.Fatal("expected persistent1 to be loaded")
	}
	if len(got1) != 3 || got1[0] != 1.0 {
		t.Errorf("unexpected persistent1 value: %v", got1)
	}

	got2, ok := cache2.Get("persistent2")
	if !ok {
		t.Fatal("expected persistent2 to be loaded")
	}
	if len(got2) != 3 || got2[0] != 2.0 {
		t.Errorf("unexpected persistent2 value: %v", got2)
	}
}

func TestLRUEmbeddingCache_PersistenceUncompressed(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "uncompressed.gob")

	// Create with uncompressed storage
	cache1 := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries:    100,
		TTL:           time.Hour,
		PersistPath:   cachePath,
		CompressCache: false,
	})

	cache1.Set("test", []float32{1.0, 2.0, 3.0})
	cache1.Close()

	// Reload
	cache2 := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries:    100,
		TTL:           time.Hour,
		PersistPath:   cachePath,
		CompressCache: false,
	})
	defer cache2.Close()

	got, ok := cache2.Get("test")
	if !ok {
		t.Fatal("expected entry to be loaded")
	}
	if len(got) != 3 {
		t.Errorf("expected 3 values, got %d", len(got))
	}
}

func TestLRUEmbeddingCache_ConcurrentAccess(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 1000,
		TTL:        time.Hour,
	})
	defer cache.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	opsPerGoroutine := 100

	// Concurrent writes
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := string(rune('a'+id)) + string(rune('0'+i%10))
				cache.Set(key, []float32{float32(id), float32(i)})
			}
		}(g)
	}

	// Concurrent reads
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := string(rune('a'+id)) + string(rune('0'+i%10))
				cache.Get(key)
			}
		}(g)
	}

	wg.Wait()

	// Verify no panic and reasonable state
	if cache.Size() > numGoroutines*10 {
		t.Errorf("cache size %d seems too large", cache.Size())
	}
}

func TestLRUEmbeddingCache_Stats(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 10,
		TTL:        time.Hour,
	})
	defer cache.Close()

	// Generate some hits and misses
	cache.Set("exists", []float32{1.0})
	cache.Get("exists")       // hit
	cache.Get("exists")       // hit
	cache.Get("missing")      // miss
	cache.Get("also missing") // miss

	stats := cache.Stats()

	if stats["hits"].(int64) != 2 {
		t.Errorf("expected 2 hits, got %d", stats["hits"])
	}
	if stats["misses"].(int64) != 2 {
		t.Errorf("expected 2 misses, got %d", stats["misses"])
	}

	hitRate := stats["hit_rate"].(float64)
	expectedRate := 0.5 // 2 hits / 4 total
	if hitRate != expectedRate {
		t.Errorf("expected hit rate %f, got %f", expectedRate, hitRate)
	}
}

func TestLRUEmbeddingCache_Cleanup(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 100,
		TTL:        50 * time.Millisecond,
	})
	defer cache.Close()

	// Add entries
	cache.Set("one", []float32{1.0})
	cache.Set("two", []float32{2.0})
	cache.Set("three", []float32{3.0})

	if cache.Size() != 3 {
		t.Fatalf("expected size 3, got %d", cache.Size())
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Manual cleanup
	removed := cache.Cleanup()
	if removed != 3 {
		t.Errorf("expected 3 removed, got %d", removed)
	}

	if cache.Size() != 0 {
		t.Errorf("expected size 0 after cleanup, got %d", cache.Size())
	}
}

func TestLRUEmbeddingCache_NoTTL(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 100,
		TTL:        0, // No TTL
	})
	defer cache.Close()

	cache.Set("forever", []float32{1.0})

	// Should never expire
	time.Sleep(10 * time.Millisecond)

	_, ok := cache.Get("forever")
	if !ok {
		t.Error("entry should not expire with TTL=0")
	}

	// Cleanup should do nothing
	removed := cache.Cleanup()
	if removed != 0 {
		t.Errorf("expected 0 removed with no TTL, got %d", removed)
	}
}

func TestLRUEmbeddingCache_UnlimitedSize(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 0, // Unlimited
		TTL:        time.Hour,
	})
	defer cache.Close()

	// Add many entries
	for i := 0; i < 100; i++ {
		key := string(rune('a' + i%26))
		cache.Set(key, []float32{float32(i)})
	}

	// All should exist (no eviction)
	stats := cache.Stats()
	if stats["evictions"].(int64) != 0 {
		t.Errorf("expected 0 evictions with unlimited size, got %d", stats["evictions"])
	}
}

func TestLRUEmbeddingCache_HashConsistency(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 100,
		TTL:        time.Hour,
	})
	defer cache.Close()

	text := "the quick brown fox jumps over the lazy dog"
	embedding := []float32{0.1, 0.2, 0.3}

	cache.Set(text, embedding)

	// Same text should produce same hash and retrieve the value
	got, ok := cache.Get(text)
	if !ok {
		t.Fatal("expected to find entry with same text")
	}

	if len(got) != len(embedding) {
		t.Errorf("expected %d values, got %d", len(embedding), len(got))
	}
}

func TestLRUEmbeddingCache_DefaultConfig(t *testing.T) {
	config := DefaultLRUCacheConfig()

	if config.MaxEntries != 10000 {
		t.Errorf("expected default MaxEntries 10000, got %d", config.MaxEntries)
	}
	if config.TTL != 24*time.Hour {
		t.Errorf("expected default TTL 24h, got %v", config.TTL)
	}
	if !config.CompressCache {
		t.Error("expected default CompressCache true")
	}
	if config.SaveInterval != 5*time.Minute {
		t.Errorf("expected default SaveInterval 5m, got %v", config.SaveInterval)
	}
}

func TestLRUEmbeddingCache_NilConfig(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, nil)
	defer cache.Close()

	// Should use defaults
	cache.Set("test", []float32{1.0})
	_, ok := cache.Get("test")
	if !ok {
		t.Error("cache should work with nil config")
	}
}

func TestLRUEmbeddingCache_MissingFile(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "nonexistent.gob")

	// Should not error on missing file
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries:  100,
		TTL:         time.Hour,
		PersistPath: cachePath,
	})
	defer cache.Close()

	if cache.Size() != 0 {
		t.Errorf("expected empty cache, got size %d", cache.Size())
	}
}

func TestLRUEmbeddingCache_LRUOrder(t *testing.T) {
	cache := mustNewLRUEmbeddingCache(t, &LRUCacheConfig{
		MaxEntries: 3,
		TTL:        time.Hour,
	})
	defer cache.Close()

	// Add entries: a, b, c (a is LRU, c is MRU)
	cache.Set("a", []float32{1.0})
	cache.Set("b", []float32{2.0})
	cache.Set("c", []float32{3.0})

	// Access b - now order is: a (LRU), c, b (MRU)
	cache.Get("b")

	// Add d - should evict a (LRU)
	cache.Set("d", []float32{4.0})

	if _, ok := cache.Get("a"); ok {
		t.Error("a should have been evicted")
	}
	if _, ok := cache.Get("b"); !ok {
		t.Error("b should exist")
	}
	if _, ok := cache.Get("c"); !ok {
		t.Error("c should exist")
	}
	if _, ok := cache.Get("d"); !ok {
		t.Error("d should exist")
	}
}

func BenchmarkLRUEmbeddingCache_Set(b *testing.B) {
	cache := mustNewLRUEmbeddingCacheBenchmark(b, &LRUCacheConfig{
		MaxEntries: 10000,
		TTL:        time.Hour,
	})
	defer cache.Close()

	embedding := make([]float32, 512) // 512-dimensional embedding
	for i := range embedding {
		embedding[i] = float32(i) * 0.001
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune('a' + i%26))
		cache.Set(key, embedding)
	}
}

func BenchmarkLRUEmbeddingCache_Get(b *testing.B) {
	cache := mustNewLRUEmbeddingCacheBenchmark(b, &LRUCacheConfig{
		MaxEntries: 10000,
		TTL:        time.Hour,
	})
	defer cache.Close()

	embedding := make([]float32, 512)
	for i := 0; i < 100; i++ {
		key := string(rune('a' + i%26))
		cache.Set(key, embedding)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune('a' + i%26))
		cache.Get(key)
	}
}

func BenchmarkLRUEmbeddingCache_ConcurrentReadWrite(b *testing.B) {
	cache := mustNewLRUEmbeddingCacheBenchmark(b, &LRUCacheConfig{
		MaxEntries: 10000,
		TTL:        time.Hour,
	})
	defer cache.Close()

	embedding := make([]float32, 512)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := string(rune('a' + i%26))
			if i%2 == 0 {
				cache.Set(key, embedding)
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}
