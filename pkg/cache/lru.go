// Package cache provides generic cache implementations for the unified-thinking server.
package cache

import (
	"sync"
	"time"
)

// Config configures an LRU cache
type Config struct {
	MaxEntries int           // Maximum cache entries (0 = unlimited)
	TTL        time.Duration // Entry TTL (0 = no expiry)
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		MaxEntries: 1000,
		TTL:        time.Hour,
	}
}

// entry is a doubly-linked list node for LRU tracking
type entry[K comparable, V any] struct {
	key    K
	value  V
	expiry time.Time
	prev   *entry[K, V]
	next   *entry[K, V]
}

// LRU is a generic LRU cache with optional TTL expiry
type LRU[K comparable, V any] struct {
	mu sync.RWMutex

	// LRU data structures
	cache map[K]*entry[K, V]
	head  *entry[K, V] // Most recently used
	tail  *entry[K, V] // Least recently used

	// Configuration
	maxEntries int
	ttl        time.Duration

	// Metrics
	hits      int64
	misses    int64
	evictions int64
	expiries  int64
}

// New creates a new generic LRU cache
func New[K comparable, V any](config *Config) *LRU[K, V] {
	if config == nil {
		config = DefaultConfig()
	}

	return &LRU[K, V]{
		cache:      make(map[K]*entry[K, V]),
		maxEntries: config.MaxEntries,
		ttl:        config.TTL,
	}
}

// Get retrieves a value from cache, returns zero value and false if not found or expired
func (c *LRU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, exists := c.cache[key]
	if !exists {
		c.misses++
		var zero V
		return zero, false
	}

	// Check TTL expiry
	if c.ttl > 0 && time.Now().After(e.expiry) {
		c.removeEntry(e)
		c.expiries++
		c.misses++
		var zero V
		return zero, false
	}

	// Move to front (most recently used)
	c.moveToFront(e)
	c.hits++

	return e.value, true
}

// Set stores a value in cache, evicting LRU entries if needed
func (c *LRU[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if e, exists := c.cache[key]; exists {
		e.value = value
		if c.ttl > 0 {
			e.expiry = time.Now().Add(c.ttl)
		}
		c.moveToFront(e)
		return
	}

	// Evict if at capacity
	if c.maxEntries > 0 && len(c.cache) >= c.maxEntries {
		c.evictLRU()
	}

	// Create new entry
	var expiry time.Time
	if c.ttl > 0 {
		expiry = time.Now().Add(c.ttl)
	}

	e := &entry[K, V]{
		key:    key,
		value:  value,
		expiry: expiry,
	}

	c.cache[key] = e
	c.addToFront(e)
}

// Delete removes a key from the cache
func (c *LRU[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, exists := c.cache[key]; exists {
		c.removeEntry(e)
	}
}

// Size returns current cache size
func (c *LRU[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Clear removes all entries
func (c *LRU[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[K]*entry[K, V])
	c.head = nil
	c.tail = nil
}

// Stats returns cache statistics
func (c *LRU[K, V]) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hitRate := float64(0)
	total := c.hits + c.misses
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return map[string]interface{}{
		"size":      len(c.cache),
		"max_size":  c.maxEntries,
		"hits":      c.hits,
		"misses":    c.misses,
		"hit_rate":  hitRate,
		"evictions": c.evictions,
		"expiries":  c.expiries,
		"ttl":       c.ttl.String(),
	}
}

// Cleanup removes all expired entries and returns count removed
func (c *LRU[K, V]) Cleanup() int {
	if c.ttl == 0 {
		return 0 // No TTL configured
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	// Traverse from tail (oldest) to head
	e := c.tail
	for e != nil {
		prev := e.prev
		if now.After(e.expiry) {
			c.removeEntry(e)
			c.expiries++
			removed++
		}
		e = prev
	}

	return removed
}

// Keys returns all keys in the cache (most recent first)
func (c *LRU[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.cache))
	e := c.head
	for e != nil {
		keys = append(keys, e.key)
		e = e.next
	}
	return keys
}

// Entry represents a cache entry with its key, value, and expiry time
type Entry[K comparable, V any] struct {
	Key    K
	Value  V
	Expiry time.Time
}

// Entries returns all non-expired entries in the cache (most recent first)
// This is useful for persistence scenarios
func (c *LRU[K, V]) Entries() []Entry[K, V] {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	entries := make([]Entry[K, V], 0, len(c.cache))
	e := c.head
	for e != nil {
		// Skip expired entries
		if c.ttl > 0 && now.After(e.expiry) {
			e = e.next
			continue
		}
		entries = append(entries, Entry[K, V]{
			Key:    e.key,
			Value:  e.value,
			Expiry: e.expiry,
		})
		e = e.next
	}
	return entries
}

// SetWithExpiry stores a value with a specific expiry time (for loading from persistence)
func (c *LRU[K, V]) SetWithExpiry(key K, value V, expiry time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if e, exists := c.cache[key]; exists {
		e.value = value
		e.expiry = expiry
		c.moveToFront(e)
		return
	}

	// Evict if at capacity
	if c.maxEntries > 0 && len(c.cache) >= c.maxEntries {
		c.evictLRU()
	}

	e := &entry[K, V]{
		key:    key,
		value:  value,
		expiry: expiry,
	}

	c.cache[key] = e
	c.addToFront(e)
}

// Internal methods

func (c *LRU[K, V]) addToFront(e *entry[K, V]) {
	e.prev = nil
	e.next = c.head

	if c.head != nil {
		c.head.prev = e
	}
	c.head = e

	if c.tail == nil {
		c.tail = e
	}
}

func (c *LRU[K, V]) moveToFront(e *entry[K, V]) {
	if e == c.head {
		return // Already at front
	}

	// Remove from current position
	if e.prev != nil {
		e.prev.next = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	}
	if e == c.tail {
		c.tail = e.prev
	}

	// Add to front
	e.prev = nil
	e.next = c.head
	if c.head != nil {
		c.head.prev = e
	}
	c.head = e
}

func (c *LRU[K, V]) removeEntry(e *entry[K, V]) {
	delete(c.cache, e.key)

	if e.prev != nil {
		e.prev.next = e.next
	} else {
		c.head = e.next
	}

	if e.next != nil {
		e.next.prev = e.prev
	} else {
		c.tail = e.prev
	}
}

func (c *LRU[K, V]) evictLRU() {
	if c.tail == nil {
		return
	}

	c.removeEntry(c.tail)
	c.evictions++
}
