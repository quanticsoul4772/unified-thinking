package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Verify server defaults
	if cfg.Server.Name != "unified-thinking" {
		t.Errorf("Expected server name 'unified-thinking', got '%s'", cfg.Server.Name)
	}
	if cfg.Server.Environment != "development" {
		t.Errorf("Expected environment 'development', got '%s'", cfg.Server.Environment)
	}

	// Verify storage defaults
	if cfg.Storage.Type != "memory" {
		t.Errorf("Expected storage type 'memory', got '%s'", cfg.Storage.Type)
	}
	if !cfg.Storage.EnableIndexing {
		t.Error("Expected indexing to be enabled by default")
	}

	// Verify all features are enabled by default
	if !cfg.Features.LinearMode {
		t.Error("Expected LinearMode to be enabled")
	}
	if !cfg.Features.ProbabilisticReasoning {
		t.Error("Expected ProbabilisticReasoning to be enabled")
	}

	// Verify performance defaults
	if cfg.Performance.MaxConcurrentThoughts != 100 {
		t.Errorf("Expected MaxConcurrentThoughts 100, got %d", cfg.Performance.MaxConcurrentThoughts)
	}
	if !cfg.Performance.EnableDeepCopy {
		t.Error("Expected EnableDeepCopy to be true")
	}

	// Verify logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected log level 'info', got '%s'", cfg.Logging.Level)
	}
}

func TestLoad(t *testing.T) {
	// Clear environment
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Should return valid default config
	if cfg.Server.Name != "unified-thinking" {
		t.Errorf("Expected default server name, got '%s'", cfg.Server.Name)
	}
}

func TestLoadWithInvalidEnv(t *testing.T) {
	// Clear environment
	clearEnv(t)
	defer clearEnv(t)

	// Set invalid environment values that will cause validation to fail
	_ = os.Setenv("UT_SERVER_ENVIRONMENT", "invalid-env")

	cfg, err := Load()
	if err == nil {
		t.Error("Expected Load() to fail with invalid environment")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
	if !contains(err.Error(), "invalid configuration") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestLoadWithInvalidStorageType(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)

	_ = os.Setenv("UT_STORAGE_TYPE", "postgresql")

	cfg, err := Load()
	if err == nil {
		t.Error("Expected Load() to fail with invalid storage type")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
}

func TestLoadWithInvalidLoggingLevel(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)

	_ = os.Setenv("UT_LOGGING_LEVEL", "invalid-level")

	cfg, err := Load()
	if err == nil {
		t.Error("Expected Load() to fail with invalid logging level")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Clear environment
	clearEnv(t)

	// Set environment variables
	_ = os.Setenv("UT_SERVER_NAME", "test-server")
	_ = os.Setenv("UT_SERVER_ENVIRONMENT", "production")
	_ = os.Setenv("UT_STORAGE_MAX_THOUGHTS", "5000")
	_ = os.Setenv("UT_FEATURES_LINEAR_MODE", "false")
	_ = os.Setenv("UT_FEATURES_PROBABILISTIC_REASONING", "true")
	_ = os.Setenv("UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS", "50")
	_ = os.Setenv("UT_LOGGING_LEVEL", "debug")

	defer clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify environment overrides
	if cfg.Server.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", cfg.Server.Name)
	}
	if cfg.Server.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", cfg.Server.Environment)
	}
	if cfg.Storage.MaxThoughts != 5000 {
		t.Errorf("Expected MaxThoughts 5000, got %d", cfg.Storage.MaxThoughts)
	}
	if cfg.Features.LinearMode {
		t.Error("Expected LinearMode to be disabled")
	}
	if !cfg.Features.ProbabilisticReasoning {
		t.Error("Expected ProbabilisticReasoning to be enabled")
	}
	if cfg.Performance.MaxConcurrentThoughts != 50 {
		t.Errorf("Expected MaxConcurrentThoughts 50, got %d", cfg.Performance.MaxConcurrentThoughts)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.Logging.Level)
	}
}

func TestLoadFromEnvComprehensive(t *testing.T) {
	// Test all environment variables for complete coverage
	clearEnv(t)
	defer clearEnv(t)

	// Set ALL environment variables
	_ = os.Setenv("UT_SERVER_NAME", "comprehensive-test")
	_ = os.Setenv("UT_SERVER_VERSION", "3.0.0")
	_ = os.Setenv("UT_SERVER_ENVIRONMENT", "staging")
	_ = os.Setenv("UT_STORAGE_TYPE", "memory")
	_ = os.Setenv("UT_STORAGE_MAX_THOUGHTS", "2000")
	_ = os.Setenv("UT_STORAGE_MAX_BRANCHES", "200")
	_ = os.Setenv("UT_STORAGE_ENABLE_INDEXING", "false")
	_ = os.Setenv("UT_FEATURES_LINEAR_MODE", "false")
	_ = os.Setenv("UT_FEATURES_TREE_MODE", "false")
	_ = os.Setenv("UT_FEATURES_DIVERGENT_MODE", "false")
	_ = os.Setenv("UT_FEATURES_AUTO_MODE", "false")
	_ = os.Setenv("UT_FEATURES_LOGICAL_VALIDATION", "false")
	_ = os.Setenv("UT_FEATURES_PROBABILISTIC_REASONING", "false")
	_ = os.Setenv("UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS", "25")
	_ = os.Setenv("UT_PERFORMANCE_ENABLE_DEEP_COPY", "false")
	_ = os.Setenv("UT_PERFORMANCE_CACHE_SIZE", "250")
	_ = os.Setenv("UT_LOGGING_LEVEL", "ERROR")
	_ = os.Setenv("UT_LOGGING_FORMAT", "JSON")
	_ = os.Setenv("UT_LOGGING_ENABLE_TIMESTAMPS", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify all server settings
	if cfg.Server.Name != "comprehensive-test" {
		t.Errorf("Expected server name 'comprehensive-test', got '%s'", cfg.Server.Name)
	}
	if cfg.Server.Version != "3.0.0" {
		t.Errorf("Expected version '3.0.0', got '%s'", cfg.Server.Version)
	}
	if cfg.Server.Environment != "staging" {
		t.Errorf("Expected environment 'staging', got '%s'", cfg.Server.Environment)
	}

	// Verify all storage settings
	if cfg.Storage.Type != "memory" {
		t.Errorf("Expected storage type 'memory', got '%s'", cfg.Storage.Type)
	}
	if cfg.Storage.MaxThoughts != 2000 {
		t.Errorf("Expected MaxThoughts 2000, got %d", cfg.Storage.MaxThoughts)
	}
	if cfg.Storage.MaxBranches != 200 {
		t.Errorf("Expected MaxBranches 200, got %d", cfg.Storage.MaxBranches)
	}
	if cfg.Storage.EnableIndexing {
		t.Error("Expected EnableIndexing to be false")
	}

	// Verify all feature flags
	if cfg.Features.LinearMode {
		t.Error("Expected LinearMode to be false")
	}
	if cfg.Features.TreeMode {
		t.Error("Expected TreeMode to be false")
	}
	if cfg.Features.DivergentMode {
		t.Error("Expected DivergentMode to be false")
	}
	if cfg.Features.AutoMode {
		t.Error("Expected AutoMode to be false")
	}
	if cfg.Features.LogicalValidation {
		t.Error("Expected LogicalValidation to be false")
	}
	if cfg.Features.ProbabilisticReasoning {
		t.Error("Expected ProbabilisticReasoning to be false")
	}

	// Verify all performance settings
	if cfg.Performance.MaxConcurrentThoughts != 25 {
		t.Errorf("Expected MaxConcurrentThoughts 25, got %d", cfg.Performance.MaxConcurrentThoughts)
	}
	if cfg.Performance.EnableDeepCopy {
		t.Error("Expected EnableDeepCopy to be false")
	}
	if cfg.Performance.CacheSize != 250 {
		t.Errorf("Expected CacheSize 250, got %d", cfg.Performance.CacheSize)
	}

	// Verify all logging settings (should be lowercased)
	if cfg.Logging.Level != "error" {
		t.Errorf("Expected log level 'error', got '%s'", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Expected log format 'json', got '%s'", cfg.Logging.Format)
	}
	if cfg.Logging.EnableTimestamps {
		t.Error("Expected EnableTimestamps to be false")
	}
}

func TestLoadFromEnvWithInvalidNumbers(t *testing.T) {
	// Test that invalid numbers are ignored (don't cause errors)
	clearEnv(t)
	defer clearEnv(t)

	_ = os.Setenv("UT_STORAGE_MAX_THOUGHTS", "not-a-number")
	_ = os.Setenv("UT_STORAGE_MAX_BRANCHES", "invalid")
	_ = os.Setenv("UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS", "abc")
	_ = os.Setenv("UT_PERFORMANCE_CACHE_SIZE", "xyz")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Should keep default values when parsing fails
	defaultCfg := Default()
	if cfg.Storage.MaxThoughts != defaultCfg.Storage.MaxThoughts {
		t.Errorf("Expected default MaxThoughts %d, got %d", defaultCfg.Storage.MaxThoughts, cfg.Storage.MaxThoughts)
	}
	if cfg.Storage.MaxBranches != defaultCfg.Storage.MaxBranches {
		t.Errorf("Expected default MaxBranches %d, got %d", defaultCfg.Storage.MaxBranches, cfg.Storage.MaxBranches)
	}
	if cfg.Performance.CacheSize != defaultCfg.Performance.CacheSize {
		t.Errorf("Expected default CacheSize %d, got %d", defaultCfg.Performance.CacheSize, cfg.Performance.CacheSize)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"server": {
			"name": "file-server",
			"version": "2.0.0",
			"environment": "staging"
		},
		"storage": {
			"type": "memory",
			"max_thoughts": 1000,
			"max_branches": 100,
			"enable_indexing": false
		},
		"features": {
			"linear_mode": true,
			"tree_mode": false,
			"divergent_mode": true,
			"probabilistic_reasoning": false
		},
		"performance": {
			"max_concurrent_thoughts": 25,
			"enable_deep_copy": false,
			"cache_size": 500
		},
		"logging": {
			"level": "warn",
			"format": "json",
			"enable_timestamps": false
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Clear environment
	clearEnv(t)

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() failed: %v", err)
	}

	// Verify file values
	if cfg.Server.Name != "file-server" {
		t.Errorf("Expected server name 'file-server', got '%s'", cfg.Server.Name)
	}
	if cfg.Server.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", cfg.Server.Version)
	}
	if cfg.Server.Environment != "staging" {
		t.Errorf("Expected environment 'staging', got '%s'", cfg.Server.Environment)
	}
	if cfg.Storage.MaxThoughts != 1000 {
		t.Errorf("Expected MaxThoughts 1000, got %d", cfg.Storage.MaxThoughts)
	}
	if cfg.Storage.EnableIndexing {
		t.Error("Expected indexing to be disabled")
	}
	if cfg.Features.TreeMode {
		t.Error("Expected TreeMode to be disabled")
	}
	if cfg.Features.ProbabilisticReasoning {
		t.Error("Expected ProbabilisticReasoning to be disabled")
	}
	if cfg.Performance.MaxConcurrentThoughts != 25 {
		t.Errorf("Expected MaxConcurrentThoughts 25, got %d", cfg.Performance.MaxConcurrentThoughts)
	}
	if cfg.Performance.EnableDeepCopy {
		t.Error("Expected EnableDeepCopy to be false")
	}
	if cfg.Logging.Level != "warn" {
		t.Errorf("Expected log level 'warn', got '%s'", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Expected log format 'json', got '%s'", cfg.Logging.Format)
	}
}

func TestLoadFromFileWithEnvOverride(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"server": {
			"name": "file-server",
			"environment": "staging"
		},
		"features": {
			"linear_mode": false
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variable to override file
	clearEnv(t)
	_ = os.Setenv("UT_SERVER_NAME", "env-server")
	_ = os.Setenv("UT_FEATURES_LINEAR_MODE", "true")
	defer clearEnv(t)

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() failed: %v", err)
	}

	// Environment should override file
	if cfg.Server.Name != "env-server" {
		t.Errorf("Expected server name 'env-server' (env override), got '%s'", cfg.Server.Name)
	}
	if !cfg.Features.LinearMode {
		t.Error("Expected LinearMode to be enabled (env override)")
	}
	// File values should be preserved where not overridden
	if cfg.Server.Environment != "staging" {
		t.Errorf("Expected environment 'staging' (from file), got '%s'", cfg.Server.Environment)
	}
}

func TestLoadFromFileNonExistent(t *testing.T) {
	clearEnv(t)
	tmpDir := t.TempDir()

	// Use an absolute path that exists as a parent but has a nonexistent file
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.json")

	cfg, err := LoadFromFile(nonExistentPath)
	if err == nil {
		t.Error("Expected LoadFromFile() to fail for non-existent file")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
	if !contains(err.Error(), "failed to read config file") {
		t.Errorf("Expected read error, got: %v", err)
	}
}

func TestLoadFromFileInvalidJSON(t *testing.T) {
	clearEnv(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	invalidJSON := `{
		"server": {
			"name": "test",
			invalid-json-here
		}
	}`

	if err := os.WriteFile(configPath, []byte(invalidJSON), 0600); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err == nil {
		t.Error("Expected LoadFromFile() to fail for invalid JSON")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
	if !contains(err.Error(), "failed to parse config file") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestLoadFromFileDirectoryTraversal(t *testing.T) {
	clearEnv(t)

	// Create a path with .. that will be cleaned differently
	// The path "foo/../bar/config.json" becomes "bar/config.json" after Clean
	traversalPath := "foo/../bar/config.json"

	cfg, err := LoadFromFile(traversalPath)
	if err == nil {
		t.Error("Expected LoadFromFile() to fail for directory traversal path")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
	if !contains(err.Error(), "directory traversal") {
		t.Errorf("Expected directory traversal error, got: %v", err)
	}
}

func TestLoadFromFileCleanPath(t *testing.T) {
	clearEnv(t)
	tmpDir := t.TempDir()

	// Create a valid config file with a clean path
	configPath := filepath.Join(tmpDir, "config.json")
	validJSON := `{
		"server": {"name": "test", "environment": "development"},
		"storage": {"type": "memory"},
		"performance": {"max_concurrent_thoughts": 100},
		"logging": {"level": "info", "format": "text"}
	}`
	if err := os.WriteFile(configPath, []byte(validJSON), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load with clean path should work
	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() with clean path failed: %v", err)
	}
	if cfg == nil {
		t.Error("Expected non-nil config")
	}
}

func TestLoadFromFileValidationFailure(t *testing.T) {
	clearEnv(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.json")

	// Create config with invalid values
	invalidConfig := `{
		"server": {
			"name": "test",
			"environment": "invalid-environment"
		},
		"storage": {
			"type": "memory"
		},
		"performance": {
			"max_concurrent_thoughts": 100
		},
		"logging": {
			"level": "info",
			"format": "text"
		}
	}`

	if err := os.WriteFile(configPath, []byte(invalidConfig), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err == nil {
		t.Error("Expected LoadFromFile() to fail validation")
	}
	if cfg != nil {
		t.Error("Expected nil config on validation error")
	}
	if !contains(err.Error(), "invalid configuration") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestLoadFromFileEnvValidationFailure(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create valid config file
	validConfig := `{
		"server": {
			"name": "test",
			"environment": "development"
		},
		"storage": {
			"type": "memory"
		},
		"performance": {
			"max_concurrent_thoughts": 100
		},
		"logging": {
			"level": "info",
			"format": "text"
		}
	}`

	if err := os.WriteFile(configPath, []byte(validConfig), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set env var that will cause validation to fail
	_ = os.Setenv("UT_SERVER_ENVIRONMENT", "invalid-from-env")

	cfg, err := LoadFromFile(configPath)
	if err == nil {
		t.Error("Expected LoadFromFile() to fail when env override causes validation error")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			cfg:     Default(),
			wantErr: false,
		},
		{
			name: "empty server name",
			cfg: &Config{
				Server:      ServerConfig{Name: "", Environment: "development"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: true,
			errMsg:  "server.name cannot be empty",
		},
		{
			name: "invalid environment",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "invalid"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: true,
			errMsg:  "server.environment must be one of",
		},
		{
			name: "invalid storage type",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "postgresql"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: true,
			errMsg:  "storage.type must be 'memory'",
		},
		{
			name: "negative max thoughts",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory", MaxThoughts: -1},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: true,
			errMsg:  "storage.max_thoughts cannot be negative",
		},
		{
			name: "invalid max concurrent thoughts",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 0},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: true,
			errMsg:  "performance.max_concurrent_thoughts must be >= 1",
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "verbose", Format: "text"},
			},
			wantErr: true,
			errMsg:  "logging.level must be one of",
		},
		{
			name: "invalid log format",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "info", Format: "xml"},
			},
			wantErr: true,
			errMsg:  "logging.format must be 'text' or 'json'",
		},
		{
			name: "negative max branches",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory", MaxBranches: -5},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: true,
			errMsg:  "storage.max_branches cannot be negative",
		},
		{
			name: "negative cache size",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100, CacheSize: -10},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: true,
			errMsg:  "performance.cache_size cannot be negative",
		},
		{
			name: "valid production environment",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "production"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: false,
		},
		{
			name: "valid staging environment",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "staging"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: false,
		},
		{
			name: "valid json log format",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "debug", Format: "json"},
			},
			wantErr: false,
		},
		{
			name: "valid warn log level",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "warn", Format: "text"},
			},
			wantErr: false,
		},
		{
			name: "valid error log level",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory"},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 100},
				Logging:     LoggingConfig{Level: "error", Format: "text"},
			},
			wantErr: false,
		},
		{
			name: "zero max thoughts is valid",
			cfg: &Config{
				Server:      ServerConfig{Name: "test", Environment: "development"},
				Storage:     StorageConfig{Type: "memory", MaxThoughts: 0, MaxBranches: 0},
				Features:    FeatureFlags{},
				Performance: PerformanceConfig{MaxConcurrentThoughts: 1, CacheSize: 0},
				Logging:     LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestIsFeatureEnabled(t *testing.T) {
	cfg := Default()

	// Test ALL feature names and aliases for complete coverage
	tests := []struct {
		name     string
		feature  string
		expected bool
	}{
		// Core modes
		{"linear mode", "linear", true},
		{"linear mode alias", "linear_mode", true},
		{"tree mode", "tree", true},
		{"tree mode alias", "tree_mode", true},
		{"divergent mode", "divergent", true},
		{"divergent mode alias", "divergent_mode", true},
		{"auto mode", "auto", true},
		{"auto mode alias", "auto_mode", true},

		// Validation features
		{"validation", "validation", true},
		{"logical_validation", "logical_validation", true},
		{"proof generation", "proof", true},
		{"proof generation alias", "proof_generation", true},
		{"syntax checking", "syntax", true},
		{"syntax checking alias", "syntax_checking", true},

		// Advanced reasoning
		{"probabilistic", "probabilistic", true},
		{"probabilistic alias", "probabilistic_reasoning", true},
		{"decision making", "decision", true},
		{"decision making alias", "decision_making", true},
		{"problem decomposition", "decompose", true},
		{"problem decomposition alias", "problem_decomposition", true},

		// Analysis capabilities
		{"evidence assessment", "evidence", true},
		{"evidence assessment alias", "evidence_assessment", true},
		{"contradiction detection", "contradictions", true},
		{"contradiction detection alias", "contradiction_detection", true},
		{"sensitivity analysis", "sensitivity", true},
		{"sensitivity analysis alias", "sensitivity_analysis", true},

		// Metacognition
		{"self evaluation", "evaluate", true},
		{"self evaluation alias", "self_evaluation", true},
		{"bias detection", "biases", true},
		{"bias detection alias", "bias_detection", true},

		// Utilities
		{"search enabled", "search", true},
		{"search enabled alias", "search_enabled", true},
		{"history enabled", "history", true},
		{"history enabled alias", "history_enabled", true},
		{"metrics enabled", "metrics", true},
		{"metrics enabled alias", "metrics_enabled", true},

		// Unknown features
		{"unknown feature", "unknown", false},
		{"empty feature", "", false},
		{"random string", "xyz123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled := cfg.IsFeatureEnabled(tt.feature)
			if enabled != tt.expected {
				t.Errorf("IsFeatureEnabled(%q) = %v, want %v", tt.feature, enabled, tt.expected)
			}
		})
	}
}

func TestIsFeatureEnabledDisabled(t *testing.T) {
	cfg := Default()

	// Disable all features and verify
	cfg.Features.LinearMode = false
	cfg.Features.TreeMode = false
	cfg.Features.DivergentMode = false
	cfg.Features.AutoMode = false
	cfg.Features.LogicalValidation = false
	cfg.Features.ProofGeneration = false
	cfg.Features.SyntaxChecking = false
	cfg.Features.ProbabilisticReasoning = false
	cfg.Features.DecisionMaking = false
	cfg.Features.ProblemDecomposition = false
	cfg.Features.EvidenceAssessment = false
	cfg.Features.ContradictionDetection = false
	cfg.Features.SensitivityAnalysis = false
	cfg.Features.SelfEvaluation = false
	cfg.Features.BiasDetection = false
	cfg.Features.SearchEnabled = false
	cfg.Features.HistoryEnabled = false
	cfg.Features.MetricsEnabled = false

	// Test each disabled feature
	disabledFeatures := []string{
		"linear", "tree", "divergent", "auto",
		"validation", "proof", "syntax",
		"probabilistic", "decision", "decompose",
		"evidence", "contradictions", "sensitivity",
		"evaluate", "biases",
		"search", "history", "metrics",
	}

	for _, feature := range disabledFeatures {
		if cfg.IsFeatureEnabled(feature) {
			t.Errorf("Expected feature %q to be disabled", feature)
		}
	}
}

func TestIsFeatureEnabledCaseInsensitive(t *testing.T) {
	cfg := Default()

	// Test case insensitivity
	tests := []string{
		"LINEAR",
		"Linear",
		"LINEAR_MODE",
		"Tree",
		"TREE_MODE",
		"DIVERGENT",
		"Auto",
		"PROBABILISTIC",
	}

	for _, feature := range tests {
		if !cfg.IsFeatureEnabled(feature) {
			t.Errorf("IsFeatureEnabled(%q) should be true (case insensitive)", feature)
		}
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"on", true},
		{"enabled", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"disabled", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBool(tt.input)
			if result != tt.expected {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	cfg := Default()
	data, err := cfg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON() returned empty data")
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	if !contains(jsonStr, "server") {
		t.Error("JSON should contain 'server' field")
	}
	if !contains(jsonStr, "features") {
		t.Error("JSON should contain 'features' field")
	}
}

func TestSaveToFile(t *testing.T) {
	cfg := Default()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved-config.json")

	err := cfg.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("SaveToFile() failed: %v", err)
	}

	// Verify file exists and can be read back
	loadedCfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() after save failed: %v", err)
	}

	if loadedCfg.Server.Name != cfg.Server.Name {
		t.Errorf("Loaded config doesn't match saved config: %s != %s", loadedCfg.Server.Name, cfg.Server.Name)
	}
}

func TestSaveToFileInvalidPath(t *testing.T) {
	cfg := Default()

	// Try to save to a path that doesn't exist
	err := cfg.SaveToFile("/nonexistent/directory/config.json")
	if err == nil {
		t.Error("Expected SaveToFile() to fail for invalid path")
	}
	if !contains(err.Error(), "failed to write config file") {
		t.Errorf("Expected write error, got: %v", err)
	}
}

func TestSaveToFileOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Save first config
	cfg1 := Default()
	cfg1.Server.Name = "first-config"
	err := cfg1.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("First SaveToFile() failed: %v", err)
	}

	// Overwrite with second config
	cfg2 := Default()
	cfg2.Server.Name = "second-config"
	err = cfg2.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("Second SaveToFile() failed: %v", err)
	}

	// Load and verify it's the second config
	clearEnv(t)
	loadedCfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() failed: %v", err)
	}

	if loadedCfg.Server.Name != "second-config" {
		t.Errorf("Expected 'second-config', got '%s'", loadedCfg.Server.Name)
	}
}

func TestSaveToFileSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	configPath := filepath.Join(subDir, "config.json")

	cfg := Default()
	err := cfg.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("SaveToFile() to subdirectory failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created in subdirectory")
	}
}

// Helper functions

func clearEnv(t *testing.T) {
	t.Helper()
	envVars := []string{
		"UT_SERVER_NAME",
		"UT_SERVER_VERSION",
		"UT_SERVER_ENVIRONMENT",
		"UT_STORAGE_TYPE",
		"UT_STORAGE_MAX_THOUGHTS",
		"UT_STORAGE_MAX_BRANCHES",
		"UT_STORAGE_ENABLE_INDEXING",
		"UT_FEATURES_LINEAR_MODE",
		"UT_FEATURES_TREE_MODE",
		"UT_FEATURES_DIVERGENT_MODE",
		"UT_FEATURES_AUTO_MODE",
		"UT_FEATURES_LOGICAL_VALIDATION",
		"UT_FEATURES_PROBABILISTIC_REASONING",
		"UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS",
		"UT_PERFORMANCE_ENABLE_DEEP_COPY",
		"UT_PERFORMANCE_CACHE_SIZE",
		"UT_LOGGING_LEVEL",
		"UT_LOGGING_FORMAT",
		"UT_LOGGING_ENABLE_TIMESTAMPS",
	}

	for _, v := range envVars {
		_ = os.Unsetenv(v)
	}
}

func contains(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
