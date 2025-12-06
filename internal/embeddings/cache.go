package embeddings

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Deprecated: EmbeddingCache is deprecated in favor of LRUEmbeddingCache which provides
// LRU eviction, disk persistence, and better memory management.
// Use NewLRUEmbeddingCache instead of NewEmbeddingCache.
//
// This simple cache has unbounded growth (no max size) and no persistence.
// It will be removed in a future version.
type EmbeddingCache struct {
	mu    sync.RWMutex
	cache map[string]*cachedEmbedding
	ttl   time.Duration
}

// cachedEmbedding represents a cached embedding with expiry
type cachedEmbedding struct {
	embedding []float32
	expiry    time.Time
}

// Deprecated: NewEmbeddingCache is deprecated. Use NewLRUEmbeddingCache instead.
func NewEmbeddingCache(ttl time.Duration) *EmbeddingCache {
	if ttl == 0 {
		ttl = 24 * time.Hour // Default to 24 hours
	}
	return &EmbeddingCache{
		cache: make(map[string]*cachedEmbedding),
		ttl:   ttl,
	}
}

// Get retrieves an embedding from cache
func (c *EmbeddingCache) Get(text string) ([]float32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.hashText(text)
	cached, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cached.expiry) {
		// Don't delete here to avoid write lock in read operation
		return nil, false
	}

	return cached.embedding, true
}

// Set stores an embedding in cache
func (c *EmbeddingCache) Set(text string, embedding []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.hashText(text)
	c.cache[key] = &cachedEmbedding{
		embedding: embedding,
		expiry:    time.Now().Add(c.ttl),
	}

	// Clean up expired entries periodically
	if len(c.cache)%100 == 0 {
		go c.cleanup()
	}
}

// Clear removes all cached embeddings
func (c *EmbeddingCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cachedEmbedding)
}

// Size returns the number of cached embeddings
func (c *EmbeddingCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// cleanup removes expired entries
func (c *EmbeddingCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, cached := range c.cache {
		if now.After(cached.expiry) {
			delete(c.cache, key)
		}
	}
}

// hashText creates a consistent hash key for text
func (c *EmbeddingCache) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
