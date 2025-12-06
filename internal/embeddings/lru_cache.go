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

	"unified-thinking/pkg/cache"
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

// LRUEmbeddingCache provides LRU-evicting, disk-persistent embedding cache
// It composes the generic cache.LRU for core LRU functionality and adds
// embedding-specific features like disk persistence and text hashing.
type LRUEmbeddingCache struct {
	mu sync.RWMutex

	// Core LRU cache (delegates all cache operations)
	inner *cache.LRU[string, []float32]

	// Configuration for persistence
	persistPath   string
	compressCache bool

	// Dirty tracking for optimized saves
	dirty bool

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

// NewLRUEmbeddingCache creates a new LRU embedding cache.
// Returns an error if persistence is configured but loading fails.
func NewLRUEmbeddingCache(config *LRUCacheConfig) (*LRUEmbeddingCache, error) {
	if config == nil {
		config = DefaultLRUCacheConfig()
	}

	// Create the inner generic cache
	inner := cache.New[string, []float32](&cache.Config{
		MaxEntries: config.MaxEntries,
		TTL:        config.TTL,
	})

	c := &LRUEmbeddingCache{
		inner:         inner,
		persistPath:   config.PersistPath,
		compressCache: config.CompressCache,
		saveInterval:  config.SaveInterval,
		stopChan:      make(chan struct{}),
	}

	// Load from disk if path configured - fail fast if load fails
	if config.PersistPath != "" {
		if err := c.Load(); err != nil {
			return nil, fmt.Errorf("failed to load cache from disk: %w", err)
		}
	}

	// Start auto-save goroutine if interval configured
	if config.SaveInterval > 0 && config.PersistPath != "" {
		c.startAutoSave()
	}

	return c, nil
}

// Get retrieves an embedding from cache, returns nil if not found or expired
func (c *LRUEmbeddingCache) Get(text string) ([]float32, bool) {
	key := c.hashText(text)
	return c.inner.Get(key)
}

// Set stores an embedding in cache, evicting LRU entries if needed
func (c *LRUEmbeddingCache) Set(text string, embedding []float32) {
	key := c.hashText(text)
	c.inner.Set(key, embedding)

	c.mu.Lock()
	c.dirty = true
	c.mu.Unlock()
}

// Size returns current cache size
func (c *LRUEmbeddingCache) Size() int {
	return c.inner.Size()
}

// Clear removes all entries
func (c *LRUEmbeddingCache) Clear() {
	c.inner.Clear()

	c.mu.Lock()
	c.dirty = true
	c.mu.Unlock()
}

// Stats returns cache statistics
func (c *LRUEmbeddingCache) Stats() map[string]interface{} {
	stats := c.inner.Stats()
	stats["persistent"] = c.persistPath != ""
	return stats
}

// Save persists cache to disk
func (c *LRUEmbeddingCache) Save() error {
	if c.persistPath == "" {
		return nil // No persistence configured
	}

	// Get all entries from the generic cache
	entries := c.inner.Entries()

	// Convert to persisted format
	persistedEntries := make([]persistedEntry, 0, len(entries))
	for _, e := range entries {
		persistedEntries = append(persistedEntries, persistedEntry{
			Key:       e.Key,
			Embedding: e.Value,
			Expiry:    e.Expiry,
		})
	}

	data := persistedCache{
		Entries:   persistedEntries,
		CreatedAt: time.Now(),
		Version:   1,
	}

	// Validate and clean the persist path
	cleanPath := filepath.Clean(c.persistPath)
	if cleanPath == "." || cleanPath == "/" {
		return fmt.Errorf("invalid cache path: %s", c.persistPath)
	}

	// Ensure directory exists
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write to temp file first for atomic save
	// #nosec G304 - path is from configuration, validated above
	tempPath := cleanPath + ".tmp"
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
	if err := os.Rename(tempPath, cleanPath); err != nil {
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

	// Validate and clean the persist path
	cleanPath := filepath.Clean(c.persistPath)
	if cleanPath == "." || cleanPath == "/" {
		return fmt.Errorf("invalid cache path: %s", c.persistPath)
	}

	// #nosec G304 - path is from configuration, validated above
	file, err := os.Open(cleanPath)
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

	// Load entries into the generic cache using SetWithExpiry
	now := time.Now()
	loaded := 0

	for _, entry := range data.Entries {
		// Skip expired entries
		if !entry.Expiry.IsZero() && now.After(entry.Expiry) {
			continue
		}

		c.inner.SetWithExpiry(entry.Key, entry.Embedding, entry.Expiry)
		loaded++
	}

	c.mu.Lock()
	c.dirty = false
	c.mu.Unlock()

	fmt.Printf("LRU cache: loaded %d entries from disk\n", loaded)
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
	removed := c.inner.Cleanup()
	if removed > 0 {
		c.mu.Lock()
		c.dirty = true
		c.mu.Unlock()
	}
	return removed
}
