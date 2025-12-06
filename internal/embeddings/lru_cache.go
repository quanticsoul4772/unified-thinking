package embeddings

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LRUCacheConfig configures the LRU embedding cache
type LRUCacheConfig struct {
	MaxEntries    int           // Maximum cache entries (0 = unlimited)
	TTL           time.Duration // Entry TTL (0 = no expiry)
	PersistPath   string        // File path for disk persistence (empty = no persistence)
	SaveInterval  time.Duration // Auto-save interval (0 = manual save only)
	CompressCache bool          // Use gzip compression for disk cache
}

// DefaultLRUCacheConfig returns sensible defaults
func DefaultLRUCacheConfig() *LRUCacheConfig {
	return &LRUCacheConfig{
		MaxEntries:    10000, // 10K entries ~= 20MB for 512d embeddings
		TTL:           24 * time.Hour,
		PersistPath:   "", // No persistence by default
		SaveInterval:  5 * time.Minute,
		CompressCache: true,
	}
}

// lruEntry is a doubly-linked list node for LRU tracking
type lruEntry struct {
	key       string
	embedding []float32
	expiry    time.Time
	prev      *lruEntry
	next      *lruEntry
}

// LRUEmbeddingCache provides LRU-evicting, disk-persistent embedding cache
type LRUEmbeddingCache struct {
	mu sync.RWMutex

	// LRU data structures
	cache map[string]*lruEntry
	head  *lruEntry // Most recently used
	tail  *lruEntry // Least recently used

	// Configuration
	maxEntries    int
	ttl           time.Duration
	persistPath   string
	compressCache bool

	// Metrics
	hits      int64
	misses    int64
	evictions int64
	expiries  int64
	dirty     bool // Track if cache needs saving

	// Auto-save
	saveInterval time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// persistedCache is the on-disk format
type persistedCache struct {
	Entries   []persistedEntry
	CreatedAt time.Time
	Version   int
}

type persistedEntry struct {
	Key       string
	Embedding []float32
	Expiry    time.Time
}

// NewLRUEmbeddingCache creates a new LRU embedding cache
func NewLRUEmbeddingCache(config *LRUCacheConfig) *LRUEmbeddingCache {
	if config == nil {
		config = DefaultLRUCacheConfig()
	}

	cache := &LRUEmbeddingCache{
		cache:         make(map[string]*lruEntry),
		maxEntries:    config.MaxEntries,
		ttl:           config.TTL,
		persistPath:   config.PersistPath,
		compressCache: config.CompressCache,
		saveInterval:  config.SaveInterval,
		stopChan:      make(chan struct{}),
	}

	// Load from disk if path configured
	if config.PersistPath != "" {
		if err := cache.Load(); err != nil {
			// Log but don't fail - start with empty cache
			fmt.Printf("LRU cache: failed to load from disk: %v\n", err)
		}
	}

	// Start auto-save goroutine if interval configured
	if config.SaveInterval > 0 && config.PersistPath != "" {
		cache.startAutoSave()
	}

	return cache
}

// Get retrieves an embedding from cache, returns nil if not found or expired
func (c *LRUEmbeddingCache) Get(text string) ([]float32, bool) {
	key := c.hashText(text)

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		c.misses++
		return nil, false
	}

	// Check TTL expiry
	if c.ttl > 0 && time.Now().After(entry.expiry) {
		c.removeEntry(entry)
		c.expiries++
		c.misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.moveToFront(entry)
	c.hits++

	return entry.embedding, true
}

// Set stores an embedding in cache, evicting LRU entries if needed
func (c *LRUEmbeddingCache) Set(text string, embedding []float32) {
	key := c.hashText(text)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if entry, exists := c.cache[key]; exists {
		entry.embedding = embedding
		if c.ttl > 0 {
			entry.expiry = time.Now().Add(c.ttl)
		}
		c.moveToFront(entry)
		c.dirty = true
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

	entry := &lruEntry{
		key:       key,
		embedding: embedding,
		expiry:    expiry,
	}

	c.cache[key] = entry
	c.addToFront(entry)
	c.dirty = true
}

// Size returns current cache size
func (c *LRUEmbeddingCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Clear removes all entries
func (c *LRUEmbeddingCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*lruEntry)
	c.head = nil
	c.tail = nil
	c.dirty = true
}

// Stats returns cache statistics
func (c *LRUEmbeddingCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hitRate := float64(0)
	total := c.hits + c.misses
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return map[string]interface{}{
		"size":       len(c.cache),
		"max_size":   c.maxEntries,
		"hits":       c.hits,
		"misses":     c.misses,
		"hit_rate":   hitRate,
		"evictions":  c.evictions,
		"expiries":   c.expiries,
		"ttl":        c.ttl.String(),
		"persistent": c.persistPath != "",
	}
}

// Save persists cache to disk
func (c *LRUEmbeddingCache) Save() error {
	if c.persistPath == "" {
		return nil // No persistence configured
	}

	c.mu.RLock()
	entries := make([]persistedEntry, 0, len(c.cache))
	now := time.Now()

	for _, entry := range c.cache {
		// Skip expired entries
		if c.ttl > 0 && now.After(entry.expiry) {
			continue
		}
		entries = append(entries, persistedEntry{
			Key:       entry.key,
			Embedding: entry.embedding,
			Expiry:    entry.expiry,
		})
	}
	c.mu.RUnlock()

	data := persistedCache{
		Entries:   entries,
		CreatedAt: now,
		Version:   1,
	}

	// Ensure directory exists
	dir := filepath.Dir(c.persistPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write to temp file first for atomic save
	tempPath := c.persistPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	var encoder *gob.Encoder
	var gzWriter *gzip.Writer

	if c.compressCache {
		gzWriter = gzip.NewWriter(file)
		encoder = gob.NewEncoder(gzWriter)
	} else {
		encoder = gob.NewEncoder(file)
	}

	if err := encoder.Encode(data); err != nil {
		_ = file.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to encode cache: %w", err)
	}

	if gzWriter != nil {
		if err := gzWriter.Close(); err != nil {
			_ = file.Close()
			_ = os.Remove(tempPath)
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, c.persistPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	c.mu.Lock()
	c.dirty = false
	c.mu.Unlock()

	return nil
}

// Load restores cache from disk
func (c *LRUEmbeddingCache) Load() error {
	if c.persistPath == "" {
		return nil
	}

	file, err := os.Open(c.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file yet
		}
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	var decoder *gob.Decoder
	var gzReader *gzip.Reader

	if c.compressCache {
		gzReader, err = gzip.NewReader(file)
		if err != nil {
			// Try uncompressed fallback
			if _, seekErr := file.Seek(0, 0); seekErr != nil {
				return fmt.Errorf("failed to seek file: %w", seekErr)
			}
			decoder = gob.NewDecoder(file)
		} else {
			defer func() {
				_ = gzReader.Close()
			}()
			decoder = gob.NewDecoder(gzReader)
		}
	} else {
		decoder = gob.NewDecoder(file)
	}

	var data persistedCache
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode cache: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	loaded := 0

	for _, entry := range data.Entries {
		// Skip expired entries
		if c.ttl > 0 && now.After(entry.Expiry) {
			continue
		}

		// Respect max entries
		if c.maxEntries > 0 && len(c.cache) >= c.maxEntries {
			break
		}

		lruEnt := &lruEntry{
			key:       entry.Key,
			embedding: entry.Embedding,
			expiry:    entry.Expiry,
		}

		c.cache[entry.Key] = lruEnt
		c.addToFront(lruEnt)
		loaded++
	}

	c.dirty = false

	return nil
}

// Close stops auto-save and performs final save
func (c *LRUEmbeddingCache) Close() error {
	// Stop auto-save
	close(c.stopChan)
	c.wg.Wait()

	// Final save
	return c.Save()
}

// Internal methods

func (c *LRUEmbeddingCache) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (c *LRUEmbeddingCache) addToFront(entry *lruEntry) {
	entry.prev = nil
	entry.next = c.head

	if c.head != nil {
		c.head.prev = entry
	}
	c.head = entry

	if c.tail == nil {
		c.tail = entry
	}
}

func (c *LRUEmbeddingCache) moveToFront(entry *lruEntry) {
	if entry == c.head {
		return // Already at front
	}

	// Remove from current position
	if entry.prev != nil {
		entry.prev.next = entry.next
	}
	if entry.next != nil {
		entry.next.prev = entry.prev
	}
	if entry == c.tail {
		c.tail = entry.prev
	}

	// Add to front
	entry.prev = nil
	entry.next = c.head
	if c.head != nil {
		c.head.prev = entry
	}
	c.head = entry
}

func (c *LRUEmbeddingCache) removeEntry(entry *lruEntry) {
	delete(c.cache, entry.key)

	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		c.head = entry.next
	}

	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		c.tail = entry.prev
	}
}

func (c *LRUEmbeddingCache) evictLRU() {
	if c.tail == nil {
		return
	}

	c.removeEntry(c.tail)
	c.evictions++
	c.dirty = true
}

func (c *LRUEmbeddingCache) startAutoSave() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.saveInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.mu.RLock()
				dirty := c.dirty
				c.mu.RUnlock()

				if dirty {
					if err := c.Save(); err != nil {
						fmt.Printf("LRU cache: auto-save failed: %v\n", err)
					}
				}
			case <-c.stopChan:
				return
			}
		}
	}()
}

// Cleanup removes all expired entries
func (c *LRUEmbeddingCache) Cleanup() int {
	if c.ttl == 0 {
		return 0 // No TTL configured
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	// Traverse from tail (oldest) to head
	entry := c.tail
	for entry != nil {
		prev := entry.prev
		if now.After(entry.expiry) {
			c.removeEntry(entry)
			c.expiries++
			removed++
		}
		entry = prev
	}

	if removed > 0 {
		c.dirty = true
	}

	return removed
}
