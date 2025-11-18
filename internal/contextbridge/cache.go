package contextbridge

import (
	"container/list"
	"sync"
	"time"
)

// LRUCache provides a thread-safe LRU cache for signature lookups
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	order    *list.List
}

type cacheEntry struct {
	key       string
	value     []*Match
	expiresAt time.Time
}

// NewLRUCache creates a new LRU cache with the specified capacity and TTL
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) []*Match {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*cacheEntry)

		// Check if expired
		if time.Now().After(entry.expiresAt) {
			c.removeElement(elem)
			return nil
		}

		// Move to front (most recently used)
		c.order.MoveToFront(elem)
		return entry.value
	}

	return nil
}

// Put adds a value to the cache
func (c *LRUCache) Put(key string, value []*Match) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update if exists
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		entry.value = value
		entry.expiresAt = time.Now().Add(c.ttl)
		return
	}

	// Evict if at capacity
	if c.order.Len() >= c.capacity {
		c.evictOldest()
	}

	// Add new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	elem := c.order.PushFront(entry)
	c.items[key] = elem
}

// evictOldest removes the least recently used item
func (c *LRUCache) evictOldest() {
	elem := c.order.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

// removeElement removes an element from the cache
func (c *LRUCache) removeElement(elem *list.Element) {
	c.order.Remove(elem)
	entry := elem.Value.(*cacheEntry)
	delete(c.items, entry.key)
}

// Len returns the current number of items in the cache
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.order = list.New()
}

// Stats returns cache statistics
func (c *LRUCache) Stats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]int{
		"size":     c.order.Len(),
		"capacity": c.capacity,
	}
}
