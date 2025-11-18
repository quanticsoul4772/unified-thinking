package contextbridge

import (
	"testing"
	"time"
)

func TestLRUCache_Basic(t *testing.T) {
	cache := NewLRUCache(3, time.Hour)

	// Test put and get
	matches := []*Match{{TrajectoryID: "1", Similarity: 0.9}}
	cache.Put("key1", matches)

	got := cache.Get("key1")
	if len(got) != 1 || got[0].TrajectoryID != "1" {
		t.Errorf("Get() = %v, want matches with TrajectoryID=1", got)
	}

	// Test cache miss
	if got := cache.Get("nonexistent"); got != nil {
		t.Errorf("Get(nonexistent) = %v, want nil", got)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(2, time.Hour)

	cache.Put("key1", []*Match{{TrajectoryID: "1"}})
	cache.Put("key2", []*Match{{TrajectoryID: "2"}})
	cache.Put("key3", []*Match{{TrajectoryID: "3"}})

	// key1 should be evicted (LRU)
	if got := cache.Get("key1"); got != nil {
		t.Error("key1 should be evicted")
	}

	// key2 and key3 should still exist
	if got := cache.Get("key2"); got == nil {
		t.Error("key2 should exist")
	}
	if got := cache.Get("key3"); got == nil {
		t.Error("key3 should exist")
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache(2, time.Hour)

	cache.Put("key1", []*Match{{TrajectoryID: "1"}})
	cache.Put("key2", []*Match{{TrajectoryID: "2"}})

	// Access key1 to make it most recently used
	cache.Get("key1")

	// Add key3 - key2 should be evicted (LRU)
	cache.Put("key3", []*Match{{TrajectoryID: "3"}})

	// key1 should still exist (was accessed)
	if got := cache.Get("key1"); got == nil {
		t.Error("key1 should exist after access")
	}

	// key2 should be evicted
	if got := cache.Get("key2"); got != nil {
		t.Error("key2 should be evicted")
	}
}

func TestLRUCache_TTLExpiration(t *testing.T) {
	cache := NewLRUCache(10, 10*time.Millisecond)

	cache.Put("key1", []*Match{{TrajectoryID: "1"}})

	// Should exist immediately
	if got := cache.Get("key1"); got == nil {
		t.Error("key1 should exist before expiration")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should be expired
	if got := cache.Get("key1"); got != nil {
		t.Error("key1 should be expired")
	}
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(10, time.Hour)

	cache.Put("key1", []*Match{{TrajectoryID: "1"}})
	cache.Put("key1", []*Match{{TrajectoryID: "updated"}})

	got := cache.Get("key1")
	if len(got) != 1 || got[0].TrajectoryID != "updated" {
		t.Errorf("Get() = %v, want updated value", got)
	}

	// Should still be only 1 entry
	if cache.Len() != 1 {
		t.Errorf("Len() = %d, want 1", cache.Len())
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10, time.Hour)

	cache.Put("key1", []*Match{{TrajectoryID: "1"}})
	cache.Put("key2", []*Match{{TrajectoryID: "2"}})

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Len() after Clear() = %d, want 0", cache.Len())
	}

	if got := cache.Get("key1"); got != nil {
		t.Error("key1 should not exist after Clear()")
	}
}

func TestLRUCache_Stats(t *testing.T) {
	cache := NewLRUCache(5, time.Hour)

	cache.Put("key1", []*Match{{TrajectoryID: "1"}})
	cache.Put("key2", []*Match{{TrajectoryID: "2"}})

	stats := cache.Stats()

	if stats["size"] != 2 {
		t.Errorf("Stats()[size] = %d, want 2", stats["size"])
	}
	if stats["capacity"] != 5 {
		t.Errorf("Stats()[capacity] = %d, want 5", stats["capacity"])
	}
}

func BenchmarkLRUCache_Put(b *testing.B) {
	cache := NewLRUCache(1000, time.Hour)
	matches := []*Match{{TrajectoryID: "1", Similarity: 0.9}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put("key", matches)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000, time.Hour)
	matches := []*Match{{TrajectoryID: "1", Similarity: 0.9}}
	cache.Put("key", matches)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}
