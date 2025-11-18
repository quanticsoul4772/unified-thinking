package contextbridge

import (
	"os"
	"strconv"
	"time"
)

// Config holds context bridge configuration
type Config struct {
	Enabled       bool
	MinSimilarity float64
	MaxMatches    int
	EnabledTools  []string
	CacheSize     int
	CacheTTL      time.Duration
	Timeout       time.Duration
}

// DefaultConfig returns the default configuration - always enabled
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true, // Always enabled by default
		MinSimilarity: 0.7,
		MaxMatches:    3,
		EnabledTools: []string{
			"think",
			"make-decision",
			"decompose-problem",
			"analyze-perspectives",
			"build-causal-graph",
		},
		CacheSize: 100,
		CacheTTL:  15 * time.Minute,
		Timeout:   2 * time.Second, // Increased to accommodate embedding API latency
	}
}

// ConfigFromEnv creates configuration from environment variables
func ConfigFromEnv() *Config {
	config := DefaultConfig()

	// Emergency disable flag (for rollback scenarios only)
	if disabled := os.Getenv("CONTEXT_BRIDGE_DISABLED"); disabled == "true" || disabled == "1" {
		config.Enabled = false
	}

	// Min similarity threshold
	if minSim := os.Getenv("CONTEXT_BRIDGE_MIN_SIMILARITY"); minSim != "" {
		if val, err := strconv.ParseFloat(minSim, 64); err == nil && val > 0 && val <= 1 {
			config.MinSimilarity = val
		}
	}

	// Max matches
	if maxMatches := os.Getenv("CONTEXT_BRIDGE_MAX_MATCHES"); maxMatches != "" {
		if val, err := strconv.Atoi(maxMatches); err == nil && val > 0 && val <= 10 {
			config.MaxMatches = val
		}
	}

	// Cache size
	if cacheSize := os.Getenv("CONTEXT_BRIDGE_CACHE_SIZE"); cacheSize != "" {
		if val, err := strconv.Atoi(cacheSize); err == nil && val > 0 && val <= 1000 {
			config.CacheSize = val
		}
	}

	// Cache TTL
	if cacheTTL := os.Getenv("CONTEXT_BRIDGE_CACHE_TTL"); cacheTTL != "" {
		if val, err := time.ParseDuration(cacheTTL); err == nil && val > 0 {
			config.CacheTTL = val
		}
	}

	// Timeout
	if timeout := os.Getenv("CONTEXT_BRIDGE_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil && val > 0 {
			config.Timeout = val
		}
	}

	return config
}
