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

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
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

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
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

	// Test various feature names and aliases
	tests := []struct {
		name     string
		feature  string
		expected bool
	}{
		{"linear mode", "linear", true},
		{"linear mode alias", "linear_mode", true},
		{"tree mode", "tree", true},
		{"probabilistic", "probabilistic", true},
		{"probabilistic alias", "probabilistic_reasoning", true},
		{"unknown feature", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled := cfg.IsFeatureEnabled(tt.feature)
			if enabled != tt.expected {
				t.Errorf("IsFeatureEnabled(%q) = %v, want %v", tt.feature, enabled, tt.expected)
			}
		})
	}

	// Test with disabled features
	cfg.Features.LinearMode = false
	if cfg.IsFeatureEnabled("linear") {
		t.Error("Expected linear mode to be disabled")
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
		os.Unsetenv(v)
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
