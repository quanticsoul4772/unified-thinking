// Package embeddings provides vector embedding generation for semantic search
package embeddings

import (
	"context"
	"os"
	"strconv"
	"time"
)

// Embedder generates vector embeddings from text
type Embedder interface {
	// Embed generates embedding for single text
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts (more efficient)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimension returns the embedding dimension
	Dimension() int

	// Model returns the model identifier
	Model() string

	// Provider returns the provider name
	Provider() string
}

// MultimodalInputType defines the type of multimodal content
type MultimodalInputType string

const (
	// InputTypeText represents text content
	InputTypeText MultimodalInputType = "text"
	// InputTypeImageBase64 represents base64-encoded image data
	InputTypeImageBase64 MultimodalInputType = "image_base64"
	// InputTypeImageURL represents an image URL
	InputTypeImageURL MultimodalInputType = "image_url"
)

// MultimodalInput represents content that can be text, image, or document
type MultimodalInput struct {
	Type     MultimodalInputType `json:"type"`                   // "text", "image_base64", "image_url"
	Text     string              `json:"text,omitempty"`         // Text content
	ImageB64 string              `json:"image_base64,omitempty"` // Base64-encoded image
	ImageURL string              `json:"image_url,omitempty"`    // Image URL
}

// ToAPIFormat converts the input to Voyage API format
func (m *MultimodalInput) ToAPIFormat() map[string]interface{} {
	result := make(map[string]interface{})
	result["type"] = string(m.Type)

	switch m.Type {
	case InputTypeText:
		result["text"] = m.Text
	case InputTypeImageBase64:
		result["image_base64"] = m.ImageB64
	case InputTypeImageURL:
		result["image_url"] = m.ImageURL
	}

	return result
}

// MultimodalEmbedder extends Embedder with multimodal support
type MultimodalEmbedder interface {
	Embedder

	// EmbedMultimodal generates embedding for multimodal content
	EmbedMultimodal(ctx context.Context, inputs []MultimodalInput) ([]float32, error)

	// EmbedImage generates embedding for a single image
	EmbedImage(ctx context.Context, imageBase64 string) ([]float32, error)

	// EmbedImageWithText generates embedding for image with text description
	EmbedImageWithText(ctx context.Context, imageBase64, text string) ([]float32, error)

	// SupportsMultimodal returns whether the embedder supports multimodal input
	SupportsMultimodal() bool
}

// EmbeddingMetadata contains metadata about an embedding
type EmbeddingMetadata struct {
	Model     string    `json:"model"`     // e.g., "voyage-3-lite"
	Provider  string    `json:"provider"`  // e.g., "voyage"
	Dimension int       `json:"dimension"` // e.g., 1024
	CreatedAt time.Time `json:"created_at"`
	Source    string    `json:"source"` // "description" or "description+context+goals"
}

// Config holds embedding configuration
type Config struct {
	Enabled  bool   `json:"enabled"`           // Master switch
	Provider string `json:"provider"`          // "voyage" for Voyage AI
	Model    string `json:"model"`             // "voyage-3-lite" or "voyage-3-large"
	APIKey   string `json:"api_key,omitempty"` // API key for provider

	// Hybrid search settings
	UseHybridSearch bool    `json:"use_hybrid_search"` // Enable RRF
	RRFParameter    int     `json:"rrf_k"`             // Default: 60
	MinSimilarity   float64 `json:"min_similarity"`    // Minimum similarity threshold (default: 0.5)

	// Caching (Legacy - use LRU cache settings for new features)
	CacheEmbeddings bool          `json:"cache_embeddings"` // Cache computed embeddings
	CacheTTL        time.Duration `json:"cache_ttl"`        // Cache expiration

	// LRU Cache settings
	CacheMaxEntries int    `json:"cache_max_entries"` // Max cache entries (0 = unlimited)
	CachePersist    bool   `json:"cache_persist"`     // Persist cache to disk
	CachePath       string `json:"cache_path"`        // Path for cache persistence

	// Performance
	BatchSize     int           `json:"batch_size"`     // Batch embedding requests
	MaxConcurrent int           `json:"max_concurrent"` // Concurrent API calls
	Timeout       time.Duration `json:"timeout"`        // API call timeout
}

// DefaultConfig returns default embedding configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:         false, // Opt-in feature
		Provider:        "voyage",
		Model:           "voyage-3-lite",
		UseHybridSearch: true,
		RRFParameter:    60,
		MinSimilarity:   0.5,
		CacheEmbeddings: true,
		CacheTTL:        24 * time.Hour,
		CacheMaxEntries: 10000, // 10K entries ~= 20MB for 512d embeddings
		CachePersist:    false, // Opt-in for disk persistence
		CachePath:       "",    // Must be set if CachePersist is true
		BatchSize:       100,
		MaxConcurrent:   5,
		Timeout:         30 * time.Second,
	}
}

// ConfigFromEnv creates config from environment variables
func ConfigFromEnv() *Config {
	cfg := DefaultConfig()

	// Read from environment
	if os.Getenv("EMBEDDINGS_ENABLED") == "true" {
		cfg.Enabled = true
	}

	if provider := os.Getenv("EMBEDDINGS_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}

	if model := os.Getenv("EMBEDDINGS_MODEL"); model != "" {
		cfg.Model = model
	}

	if apiKey := os.Getenv("VOYAGE_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}

	if os.Getenv("EMBEDDINGS_HYBRID_SEARCH") == "true" {
		cfg.UseHybridSearch = true
	}

	if k := os.Getenv("EMBEDDINGS_RRF_K"); k != "" {
		if val, err := strconv.Atoi(k); err == nil {
			cfg.RRFParameter = val
		}
	}

	if minSim := os.Getenv("EMBEDDINGS_MIN_SIMILARITY"); minSim != "" {
		if val, err := strconv.ParseFloat(minSim, 64); err == nil {
			cfg.MinSimilarity = val
		}
	}

	if os.Getenv("EMBEDDINGS_CACHE_ENABLED") == "false" {
		cfg.CacheEmbeddings = false
	}

	if ttl := os.Getenv("EMBEDDINGS_CACHE_TTL"); ttl != "" {
		if duration, err := time.ParseDuration(ttl); err == nil {
			cfg.CacheTTL = duration
		}
	}

	if batchSize := os.Getenv("EMBEDDINGS_BATCH_SIZE"); batchSize != "" {
		if val, err := strconv.Atoi(batchSize); err == nil {
			cfg.BatchSize = val
		}
	}

	if maxConcurrent := os.Getenv("EMBEDDINGS_MAX_CONCURRENT"); maxConcurrent != "" {
		if val, err := strconv.Atoi(maxConcurrent); err == nil {
			cfg.MaxConcurrent = val
		}
	}

	if timeout := os.Getenv("EMBEDDINGS_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			cfg.Timeout = duration
		}
	}

	// LRU Cache settings
	if maxEntries := os.Getenv("EMBEDDINGS_CACHE_MAX_ENTRIES"); maxEntries != "" {
		if val, err := strconv.Atoi(maxEntries); err == nil {
			cfg.CacheMaxEntries = val
		}
	}

	if os.Getenv("EMBEDDINGS_CACHE_PERSIST") == "true" {
		cfg.CachePersist = true
	}

	if cachePath := os.Getenv("EMBEDDINGS_CACHE_PATH"); cachePath != "" {
		cfg.CachePath = cachePath
		cfg.CachePersist = true // Auto-enable persistence if path is set
	}

	return cfg
}

// ToLRUCacheConfig converts embedding Config to LRUCacheConfig
func (c *Config) ToLRUCacheConfig() *LRUCacheConfig {
	persistPath := ""
	if c.CachePersist && c.CachePath != "" {
		persistPath = c.CachePath
	}

	return &LRUCacheConfig{
		MaxEntries:    c.CacheMaxEntries,
		TTL:           c.CacheTTL,
		PersistPath:   persistPath,
		SaveInterval:  5 * time.Minute,
		CompressCache: true,
	}
}
