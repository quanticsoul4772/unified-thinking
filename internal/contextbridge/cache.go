package contextbridge

import (
	"time"

	"unified-thinking/pkg/cache"
)

// LRUCache provides a thread-safe LRU cache for signature lookups
// This is a thin wrapper around the generic cache.LRU for backwards compatibility
type LRUCache struct {
	inner *cache.LRU[string, []*Match]
}

// NewLRUCache creates a new LRU cache with the specified capacity and TTL
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		inner: cache.New[string, []*Match](&cache.Config{
			MaxEntries: capacity,
			TTL:        ttl,
		}),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) []*Match {
	val, found := c.inner.Get(key)
	if !found {
		return nil
	}
	return val
}

// Put adds a value to the cache
func (c *LRUCache) Put(key string, value []*Match) {
	c.inner.Set(key, value)
}

// Len returns the current number of items in the cache
func (c *LRUCache) Len() int {
	return c.inner.Size()
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.inner.Clear()
}

// Stats returns cache statistics
func (c *LRUCache) Stats() map[string]int {
	stats := c.inner.Stats()
	return map[string]int{
		"size":     stats["size"].(int),
		"capacity": stats["max_size"].(int),
	}
}
