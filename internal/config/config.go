// Package config provides configuration management for the unified thinking server.
//
// Configuration can be loaded from multiple sources (in order of precedence):
// 1. Environment variables (highest priority)
// 2. Configuration file (JSON/YAML)
// 3. Default values (lowest priority)
//
// Feature flags allow enabling/disabling specific capabilities at runtime.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the complete server configuration
type Config struct {
	// Server settings
	Server ServerConfig `json:"server"`

	// Storage settings
	Storage StorageConfig `json:"storage"`

	// Feature flags
	Features FeatureFlags `json:"features"`

	// Performance settings
	Performance PerformanceConfig `json:"performance"`

	// Logging settings
	Logging LoggingConfig `json:"logging"`
}

// ServerConfig contains server-level configuration
type ServerConfig struct {
	// Name of the server (for logging/identification)
	Name string `json:"name"`

	// Version of the server
	Version string `json:"version"`

	// Environment (development, staging, production)
	Environment string `json:"environment"`
}

// StorageConfig contains storage-level configuration
type StorageConfig struct {
	// Type of storage backend (currently only "memory" is supported)
	Type string `json:"type"`

	// MaxThoughts limits the number of thoughts stored (0 = unlimited)
	MaxThoughts int `json:"max_thoughts"`

	// MaxBranches limits the number of branches (0 = unlimited)
	MaxBranches int `json:"max_branches"`

	// EnableIndexing enables/disables content indexing for search
	EnableIndexing bool `json:"enable_indexing"`
}

// FeatureFlags controls which features are enabled
type FeatureFlags struct {
	// Core thinking modes
	LinearMode    bool `json:"linear_mode"`
	TreeMode      bool `json:"tree_mode"`
	DivergentMode bool `json:"divergent_mode"`
	AutoMode      bool `json:"auto_mode"`

	// Validation features
	LogicalValidation bool `json:"logical_validation"`
	ProofGeneration   bool `json:"proof_generation"`
	SyntaxChecking    bool `json:"syntax_checking"`

	// Advanced reasoning
	ProbabilisticReasoning bool `json:"probabilistic_reasoning"`
	DecisionMaking         bool `json:"decision_making"`
	ProblemDecomposition   bool `json:"problem_decomposition"`

	// Analysis capabilities
	EvidenceAssessment     bool `json:"evidence_assessment"`
	ContradictionDetection bool `json:"contradiction_detection"`
	SensitivityAnalysis    bool `json:"sensitivity_analysis"`

	// Metacognition
	SelfEvaluation bool `json:"self_evaluation"`
	BiasDetection  bool `json:"bias_detection"`

	// Search and history
	SearchEnabled  bool `json:"search_enabled"`
	HistoryEnabled bool `json:"history_enabled"`
	MetricsEnabled bool `json:"metrics_enabled"`
}

// PerformanceConfig contains performance tuning options
type PerformanceConfig struct {
	// MaxConcurrentThoughts limits concurrent thought processing
	MaxConcurrentThoughts int `json:"max_concurrent_thoughts"`

	// EnableDeepCopy controls whether storage returns deep copies (thread safety)
	EnableDeepCopy bool `json:"enable_deep_copy"`

	// CacheSize sets the size of various internal caches (0 = no caching)
	CacheSize int `json:"cache_size"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	// Level sets the logging level (debug, info, warn, error)
	Level string `json:"level"`

	// Format sets the log format (text, json)
	Format string `json:"format"`

	// EnableTimestamps adds timestamps to log entries
	EnableTimestamps bool `json:"enable_timestamps"`
}

// Default returns the default configuration with all features enabled
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Name:        "unified-thinking",
			Version:     "1.0.0",
			Environment: "development",
		},
		Storage: StorageConfig{
			Type:           "memory",
			MaxThoughts:    0, // Unlimited
			MaxBranches:    0, // Unlimited
			EnableIndexing: true,
		},
		Features: FeatureFlags{
			// Core modes - all enabled by default
			LinearMode:    true,
			TreeMode:      true,
			DivergentMode: true,
			AutoMode:      true,

			// Validation - all enabled
			LogicalValidation: true,
			ProofGeneration:   true,
			SyntaxChecking:    true,

			// Advanced reasoning - all enabled
			ProbabilisticReasoning: true,
			DecisionMaking:         true,
			ProblemDecomposition:   true,

			// Analysis - all enabled
			EvidenceAssessment:     true,
			ContradictionDetection: true,
			SensitivityAnalysis:    true,

			// Metacognition - all enabled
			SelfEvaluation: true,
			BiasDetection:  true,

			// Utilities - all enabled
			SearchEnabled:  true,
			HistoryEnabled: true,
			MetricsEnabled: true,
		},
		Performance: PerformanceConfig{
			MaxConcurrentThoughts: 100,
			EnableDeepCopy:        true,
			CacheSize:             1000,
		},
		Logging: LoggingConfig{
			Level:            "info",
			Format:           "text",
			EnableTimestamps: true,
		},
	}
}

// Load loads configuration from environment variables and applies defaults
func Load() (*Config, error) {
	cfg := Default()

	// Load from environment variables
	if err := cfg.loadFromEnv(); err != nil {
		return nil, fmt.Errorf("failed to load from environment: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(path string) (*Config, error) {
	// Start with defaults
	cfg := Default()

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	if err := cfg.loadFromEnv(); err != nil {
		return nil, fmt.Errorf("failed to load from environment: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables
// Environment variables follow the pattern: UT_<SECTION>_<KEY>
// Example: UT_SERVER_NAME, UT_FEATURES_LINEAR_MODE
func (c *Config) loadFromEnv() error {
	// Server settings
	if v := os.Getenv("UT_SERVER_NAME"); v != "" {
		c.Server.Name = v
	}
	if v := os.Getenv("UT_SERVER_VERSION"); v != "" {
		c.Server.Version = v
	}
	if v := os.Getenv("UT_SERVER_ENVIRONMENT"); v != "" {
		c.Server.Environment = v
	}

	// Storage settings
	if v := os.Getenv("UT_STORAGE_TYPE"); v != "" {
		c.Storage.Type = v
	}
	if v := os.Getenv("UT_STORAGE_MAX_THOUGHTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Storage.MaxThoughts = n
		}
	}
	if v := os.Getenv("UT_STORAGE_MAX_BRANCHES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Storage.MaxBranches = n
		}
	}
	if v := os.Getenv("UT_STORAGE_ENABLE_INDEXING"); v != "" {
		c.Storage.EnableIndexing = parseBool(v)
	}

	// Feature flags
	if v := os.Getenv("UT_FEATURES_LINEAR_MODE"); v != "" {
		c.Features.LinearMode = parseBool(v)
	}
	if v := os.Getenv("UT_FEATURES_TREE_MODE"); v != "" {
		c.Features.TreeMode = parseBool(v)
	}
	if v := os.Getenv("UT_FEATURES_DIVERGENT_MODE"); v != "" {
		c.Features.DivergentMode = parseBool(v)
	}
	if v := os.Getenv("UT_FEATURES_AUTO_MODE"); v != "" {
		c.Features.AutoMode = parseBool(v)
	}
	if v := os.Getenv("UT_FEATURES_LOGICAL_VALIDATION"); v != "" {
		c.Features.LogicalValidation = parseBool(v)
	}
	if v := os.Getenv("UT_FEATURES_PROBABILISTIC_REASONING"); v != "" {
		c.Features.ProbabilisticReasoning = parseBool(v)
	}

	// Performance settings
	if v := os.Getenv("UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Performance.MaxConcurrentThoughts = n
		}
	}
	if v := os.Getenv("UT_PERFORMANCE_ENABLE_DEEP_COPY"); v != "" {
		c.Performance.EnableDeepCopy = parseBool(v)
	}
	if v := os.Getenv("UT_PERFORMANCE_CACHE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Performance.CacheSize = n
		}
	}

	// Logging settings
	if v := os.Getenv("UT_LOGGING_LEVEL"); v != "" {
		c.Logging.Level = strings.ToLower(v)
	}
	if v := os.Getenv("UT_LOGGING_FORMAT"); v != "" {
		c.Logging.Format = strings.ToLower(v)
	}
	if v := os.Getenv("UT_LOGGING_ENABLE_TIMESTAMPS"); v != "" {
		c.Logging.EnableTimestamps = parseBool(v)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Name == "" {
		return fmt.Errorf("server.name cannot be empty")
	}
	if c.Server.Environment != "development" && c.Server.Environment != "staging" && c.Server.Environment != "production" {
		return fmt.Errorf("server.environment must be one of: development, staging, production")
	}

	// Validate storage config
	if c.Storage.Type != "memory" {
		return fmt.Errorf("storage.type must be 'memory' (only supported type)")
	}
	if c.Storage.MaxThoughts < 0 {
		return fmt.Errorf("storage.max_thoughts cannot be negative")
	}
	if c.Storage.MaxBranches < 0 {
		return fmt.Errorf("storage.max_branches cannot be negative")
	}

	// Validate performance config
	if c.Performance.MaxConcurrentThoughts < 1 {
		return fmt.Errorf("performance.max_concurrent_thoughts must be >= 1")
	}
	if c.Performance.CacheSize < 0 {
		return fmt.Errorf("performance.cache_size cannot be negative")
	}

	// Validate logging config
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}
	if c.Logging.Format != "text" && c.Logging.Format != "json" {
		return fmt.Errorf("logging.format must be 'text' or 'json'")
	}

	return nil
}

// IsFeatureEnabled checks if a specific feature is enabled
func (c *Config) IsFeatureEnabled(feature string) bool {
	switch strings.ToLower(feature) {
	case "linear", "linear_mode":
		return c.Features.LinearMode
	case "tree", "tree_mode":
		return c.Features.TreeMode
	case "divergent", "divergent_mode":
		return c.Features.DivergentMode
	case "auto", "auto_mode":
		return c.Features.AutoMode
	case "validation", "logical_validation":
		return c.Features.LogicalValidation
	case "proof", "proof_generation":
		return c.Features.ProofGeneration
	case "syntax", "syntax_checking":
		return c.Features.SyntaxChecking
	case "probabilistic", "probabilistic_reasoning":
		return c.Features.ProbabilisticReasoning
	case "decision", "decision_making":
		return c.Features.DecisionMaking
	case "decompose", "problem_decomposition":
		return c.Features.ProblemDecomposition
	case "evidence", "evidence_assessment":
		return c.Features.EvidenceAssessment
	case "contradictions", "contradiction_detection":
		return c.Features.ContradictionDetection
	case "sensitivity", "sensitivity_analysis":
		return c.Features.SensitivityAnalysis
	case "evaluate", "self_evaluation":
		return c.Features.SelfEvaluation
	case "biases", "bias_detection":
		return c.Features.BiasDetection
	case "search", "search_enabled":
		return c.Features.SearchEnabled
	case "history", "history_enabled":
		return c.Features.HistoryEnabled
	case "metrics", "metrics_enabled":
		return c.Features.MetricsEnabled
	default:
		return false
	}
}

// parseBool parses a boolean from string (handles various formats)
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "on" || s == "enabled"
}

// ToJSON serializes the configuration to JSON
func (c *Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// SaveToFile saves the configuration to a JSON file
func (c *Config) SaveToFile(path string) error {
	data, err := c.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
