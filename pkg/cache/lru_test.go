package cache

import (
	"sync"
	"testing"
	"time"
)

func TestLRU_BasicOperations(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: time.Hour})

	// Test Set and Get
	c.Set("key1", 100)
	c.Set("key2", 200)

	val, found := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if val != 100 {
		t.Errorf("expected 100, got %d", val)
	}

	val, found = c.Get("key2")
	if !found {
		t.Fatal("expected to find key2")
	}
	if val != 200 {
		t.Errorf("expected 200, got %d", val)
	}
}

func TestLRU_NotFound(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: time.Hour})

	val, found := c.Get("nonexistent")
	if found {
		t.Error("expected not found for nonexistent key")
	}
	if val != 0 {
		t.Errorf("expected zero value, got %d", val)
	}
}

func TestLRU_Update(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: time.Hour})

	c.Set("key1", 100)
	c.Set("key1", 200) // Update same key

	val, found := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if val != 200 {
		t.Errorf("expected 200, got %d", val)
	}
	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}
}

func TestLRU_Delete(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: time.Hour})

	c.Set("key1", 100)
	c.Set("key2", 200)

	c.Delete("key1")

	_, found := c.Get("key1")
	if found {
		t.Error("expected key1 to be deleted")
	}

	val, found := c.Get("key2")
	if !found {
		t.Fatal("expected to find key2")
	}
	if val != 200 {
		t.Errorf("expected 200, got %d", val)
	}
}

func TestLRU_Eviction(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 3, TTL: time.Hour})

	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)

	// Access key1 to make it most recently used
	c.Get("key1")

	// Add key4, should evict key2 (least recently used)
	c.Set("key4", 4)

	// key2 should be evicted
	_, found := c.Get("key2")
	if found {
		t.Error("expected key2 to be evicted")
	}

	// Others should still exist
	if _, found := c.Get("key1"); !found {
		t.Error("expected key1 to exist")
	}
	if _, found := c.Get("key3"); !found {
		t.Error("expected key3 to exist")
	}
	if _, found := c.Get("key4"); !found {
		t.Error("expected key4 to exist")
	}
}

func TestLRU_Expiration(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: 50 * time.Millisecond})

	c.Set("key1", 100)

	// Should be found immediately
	_, found := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1 before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not be found after expiration
	_, found = c.Get("key1")
	if found {
		t.Error("expected not found after expiration")
	}
}

func TestLRU_NoTTL(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: 0})

	c.Set("key1", 100)

	// Should be found
	_, found := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1 with no TTL")
	}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Should still be found (no TTL)
	_, found = c.Get("key1")
	if !found {
		t.Error("expected key1 to persist with no TTL")
	}
}

func TestLRU_Clear(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: time.Hour})

	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)

	if c.Size() != 3 {
		t.Errorf("expected size 3, got %d", c.Size())
	}

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", c.Size())
	}

	_, found := c.Get("key1")
	if found {
		t.Error("expected key1 to be cleared")
	}
}

func TestLRU_Size(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 100, TTL: time.Hour})

	if c.Size() != 0 {
		t.Errorf("expected initial size 0, got %d", c.Size())
	}

	for i := 0; i < 10; i++ {
		c.Set(string(rune('a'+i)), i)
	}

	if c.Size() != 10 {
		t.Errorf("expected size 10, got %d", c.Size())
	}
}

func TestLRU_Keys(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: time.Hour})

	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)

	// Access key1 to make it most recent
	c.Get("key1")

	keys := c.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}

	// Most recent should be first
	if keys[0] != "key1" {
		t.Errorf("expected key1 first, got %s", keys[0])
	}
}

func TestLRU_Stats(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: time.Hour})

	c.Set("key1", 1)
	c.Get("key1") // hit
	c.Get("key2") // miss

	stats := c.Stats()

	if stats["hits"].(int64) != 1 {
		t.Errorf("expected 1 hit, got %d", stats["hits"])
	}
	if stats["misses"].(int64) != 1 {
		t.Errorf("expected 1 miss, got %d", stats["misses"])
	}
	if stats["size"].(int) != 1 {
		t.Errorf("expected size 1, got %d", stats["size"])
	}
}

func TestLRU_Cleanup(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: 50 * time.Millisecond})

	c.Set("key1", 1)
	c.Set("key2", 2)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	removed := c.Cleanup()
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	if c.Size() != 0 {
		t.Errorf("expected size 0 after cleanup, got %d", c.Size())
	}
}

func TestLRU_CleanupNoTTL(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 10, TTL: 0})

	c.Set("key1", 1)

	removed := c.Cleanup()
	if removed != 0 {
		t.Errorf("expected 0 removed with no TTL, got %d", removed)
	}
}

func TestLRU_Concurrent(t *testing.T) {
	c := New[int, int](&Config{MaxEntries: 1000, TTL: time.Hour})

	var wg sync.WaitGroup
	n := 100

	// Concurrent writes
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Set(i, i*2)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Get(i)
		}(i)
	}

	wg.Wait()

	// Should have all values
	if c.Size() != n {
		t.Errorf("expected size %d, got %d", n, c.Size())
	}
}

func TestLRU_UnlimitedEntries(t *testing.T) {
	c := New[string, int](&Config{MaxEntries: 0, TTL: time.Hour})

	// Add many entries
	for i := 0; i < 100; i++ {
		c.Set(string(rune('a'+i)), i)
	}

	// All should exist (no eviction)
	if c.Size() != 100 {
		t.Errorf("expected size 100, got %d", c.Size())
	}
}

func TestLRU_DefaultConfig(t *testing.T) {
	c := New[string, int](nil)

	c.Set("key1", 1)
	val, found := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	stats := c.Stats()
	if stats["max_size"].(int) != 1000 {
		t.Errorf("expected default max_size 1000, got %d", stats["max_size"])
	}
}

func TestLRU_ComplexValueType(t *testing.T) {
	type Complex struct {
		Name  string
		Count int
		Data  []byte
	}

	c := New[string, *Complex](&Config{MaxEntries: 10, TTL: time.Hour})

	c.Set("key1", &Complex{Name: "test", Count: 42, Data: []byte{1, 2, 3}})

	val, found := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if val.Name != "test" {
		t.Errorf("expected Name 'test', got '%s'", val.Name)
	}
	if val.Count != 42 {
		t.Errorf("expected Count 42, got %d", val.Count)
	}
}

// Benchmarks

func BenchmarkLRU_Get(b *testing.B) {
	c := New[string, int](&Config{MaxEntries: 1000, TTL: time.Hour})
	c.Set("key", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key")
	}
}

func BenchmarkLRU_Set(b *testing.B) {
	c := New[string, int](&Config{MaxEntries: 10000, TTL: time.Hour})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key", i)
	}
}

func BenchmarkLRU_SetWithEviction(b *testing.B) {
	c := New[int, int](&Config{MaxEntries: 100, TTL: time.Hour})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i)
	}
}

func BenchmarkLRU_Concurrent(b *testing.B) {
	c := New[int, int](&Config{MaxEntries: 10000, TTL: time.Hour})

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				c.Set(i, i)
			} else {
				c.Get(i)
			}
			i++
		}
	})
}
