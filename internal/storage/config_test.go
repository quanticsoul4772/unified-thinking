package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != StorageTypeMemory {
		t.Errorf("Default Type = %v, want %v", config.Type, StorageTypeMemory)
	}

	if config.SQLitePath != "./data/unified-thinking.db" {
		t.Errorf("Default SQLitePath = %v, want './data/unified-thinking.db'", config.SQLitePath)
	}

	if config.SQLiteTimeout != 5000 {
		t.Errorf("Default SQLiteTimeout = %v, want 5000", config.SQLiteTimeout)
	}

	if config.FallbackType != StorageTypeMemory {
		t.Errorf("Default FallbackType = %v, want %v", config.FallbackType, StorageTypeMemory)
	}
}

func TestConfigFromEnv(t *testing.T) {
	// Save original env vars
	originalStorageType := os.Getenv("STORAGE_TYPE")
	originalSQLitePath := os.Getenv("SQLITE_PATH")
	originalSQLiteTimeout := os.Getenv("SQLITE_TIMEOUT")
	originalFallback := os.Getenv("STORAGE_FALLBACK")

	// Restore original env vars after test
	defer func() {
		os.Setenv("STORAGE_TYPE", originalStorageType)
		os.Setenv("SQLITE_PATH", originalSQLitePath)
		os.Setenv("SQLITE_TIMEOUT", originalSQLiteTimeout)
		os.Setenv("STORAGE_FALLBACK", originalFallback)
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*testing.T, Config)
	}{
		{
			name:    "default config when no env vars",
			envVars: map[string]string{},
			validate: func(t *testing.T, cfg Config) {
				if cfg.Type != StorageTypeMemory {
					t.Errorf("Type = %v, want %v", cfg.Type, StorageTypeMemory)
				}
				if cfg.SQLitePath != "./data/unified-thinking.db" {
					t.Errorf("SQLitePath = %v, want default", cfg.SQLitePath)
				}
				if cfg.SQLiteTimeout != 5000 {
					t.Errorf("SQLiteTimeout = %v, want 5000", cfg.SQLiteTimeout)
				}
				if cfg.FallbackType != StorageTypeMemory {
					t.Errorf("FallbackType = %v, want %v", cfg.FallbackType, StorageTypeMemory)
				}
			},
		},
		{
			name: "memory storage type",
			envVars: map[string]string{
				"STORAGE_TYPE": "memory",
			},
			validate: func(t *testing.T, cfg Config) {
				if cfg.Type != StorageTypeMemory {
					t.Errorf("Type = %v, want memory", cfg.Type)
				}
			},
		},
		{
			name: "sqlite storage type",
			envVars: map[string]string{
				"STORAGE_TYPE": "sqlite",
			},
			validate: func(t *testing.T, cfg Config) {
				if cfg.Type != StorageTypeSQLite {
					t.Errorf("Type = %v, want sqlite", cfg.Type)
				}
			},
		},
		{
			name: "custom sqlite path",
			envVars: map[string]string{
				"SQLITE_PATH": "/custom/path/db.sqlite",
			},
			validate: func(t *testing.T, cfg Config) {
				if cfg.SQLitePath != "/custom/path/db.sqlite" {
					t.Errorf("SQLitePath = %v, want /custom/path/db.sqlite", cfg.SQLitePath)
				}
			},
		},
		{
			name: "custom sqlite timeout",
			envVars: map[string]string{
				"SQLITE_TIMEOUT": "10000",
			},
			validate: func(t *testing.T, cfg Config) {
				if cfg.SQLiteTimeout != 10000 {
					t.Errorf("SQLiteTimeout = %v, want 10000", cfg.SQLiteTimeout)
				}
			},
		},
		{
			name: "invalid timeout (non-numeric)",
			envVars: map[string]string{
				"SQLITE_TIMEOUT": "invalid",
			},
			validate: func(t *testing.T, cfg Config) {
				// Should fall back to default
				if cfg.SQLiteTimeout != 5000 {
					t.Errorf("SQLiteTimeout = %v, want 5000 (default)", cfg.SQLiteTimeout)
				}
			},
		},
		{
			name: "negative timeout",
			envVars: map[string]string{
				"SQLITE_TIMEOUT": "-1000",
			},
			validate: func(t *testing.T, cfg Config) {
				// Should fall back to default (negative not valid)
				if cfg.SQLiteTimeout != 5000 {
					t.Errorf("SQLiteTimeout = %v, want 5000 (default)", cfg.SQLiteTimeout)
				}
			},
		},
		{
			name: "zero timeout",
			envVars: map[string]string{
				"SQLITE_TIMEOUT": "0",
			},
			validate: func(t *testing.T, cfg Config) {
				// Should fall back to default (zero not valid)
				if cfg.SQLiteTimeout != 5000 {
					t.Errorf("SQLiteTimeout = %v, want 5000 (default)", cfg.SQLiteTimeout)
				}
			},
		},
		{
			name: "custom fallback type",
			envVars: map[string]string{
				"STORAGE_FALLBACK": "sqlite",
			},
			validate: func(t *testing.T, cfg Config) {
				if cfg.FallbackType != StorageTypeSQLite {
					t.Errorf("FallbackType = %v, want sqlite", cfg.FallbackType)
				}
			},
		},
		{
			name: "all custom values",
			envVars: map[string]string{
				"STORAGE_TYPE":     "sqlite",
				"SQLITE_PATH":      "/tmp/test.db",
				"SQLITE_TIMEOUT":   "8000",
				"STORAGE_FALLBACK": "memory",
			},
			validate: func(t *testing.T, cfg Config) {
				if cfg.Type != StorageTypeSQLite {
					t.Errorf("Type = %v, want sqlite", cfg.Type)
				}
				if cfg.SQLitePath != "/tmp/test.db" {
					t.Errorf("SQLitePath = %v, want /tmp/test.db", cfg.SQLitePath)
				}
				if cfg.SQLiteTimeout != 8000 {
					t.Errorf("SQLiteTimeout = %v, want 8000", cfg.SQLiteTimeout)
				}
				if cfg.FallbackType != StorageTypeMemory {
					t.Errorf("FallbackType = %v, want memory", cfg.FallbackType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			os.Unsetenv("STORAGE_TYPE")
			os.Unsetenv("SQLITE_PATH")
			os.Unsetenv("SQLITE_TIMEOUT")
			os.Unsetenv("STORAGE_FALLBACK")

			// Set env vars for this test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config := ConfigFromEnv()
			tt.validate(t, config)
		})
	}
}

func TestConfigFromEnvCreatesDirectory(t *testing.T) {
	// Save original env vars
	originalStorageType := os.Getenv("STORAGE_TYPE")
	originalSQLitePath := os.Getenv("SQLITE_PATH")

	defer func() {
		os.Setenv("STORAGE_TYPE", originalStorageType)
		os.Setenv("SQLITE_PATH", originalSQLitePath)
	}()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "subdir", "nested", "test.db")

	os.Setenv("STORAGE_TYPE", "sqlite")
	os.Setenv("SQLITE_PATH", dbPath)

	config := ConfigFromEnv()

	// Verify parent directory was created
	parentDir := filepath.Dir(dbPath)
	info, err := os.Stat(parentDir)

	if err != nil {
		t.Errorf("Parent directory not created: %v", err)
	} else if !info.IsDir() {
		t.Error("Parent path exists but is not a directory")
	}

	if config.SQLitePath != dbPath {
		t.Errorf("SQLitePath = %v, want %v", config.SQLitePath, dbPath)
	}
}

func TestConfigFromEnvDirectoryCreationFailure(t *testing.T) {
	// This test verifies that ConfigFromEnv doesn't panic on directory creation failure
	// We can't easily trigger this without platform-specific code, so we just verify
	// it doesn't panic with an invalid path

	originalStorageType := os.Getenv("STORAGE_TYPE")
	originalSQLitePath := os.Getenv("SQLITE_PATH")

	defer func() {
		os.Setenv("STORAGE_TYPE", originalStorageType)
		os.Setenv("SQLITE_PATH", originalSQLitePath)
	}()

	// Use a path with null byte (invalid on most systems)
	os.Setenv("STORAGE_TYPE", "sqlite")
	os.Setenv("SQLITE_PATH", "/invalid/\x00/path/test.db")

	// Should not panic even if directory creation fails
	config := ConfigFromEnv()

	// Config should still be created with the invalid path
	// (it will fail later when trying to open the database)
	if config.Type != StorageTypeSQLite {
		t.Errorf("Type = %v, want sqlite", config.Type)
	}
}

func TestStorageTypeConstants(t *testing.T) {
	// Verify storage type constants are defined correctly
	if StorageTypeMemory != "memory" {
		t.Errorf("StorageTypeMemory = %v, want 'memory'", StorageTypeMemory)
	}

	if StorageTypeSQLite != "sqlite" {
		t.Errorf("StorageTypeSQLite = %v, want 'sqlite'", StorageTypeSQLite)
	}
}

func TestConfigStructFields(t *testing.T) {
	// Test that Config struct can be created and all fields are accessible
	config := Config{
		Type:          StorageTypeSQLite,
		SQLitePath:    "/test/path.db",
		SQLiteTimeout: 3000,
		FallbackType:  StorageTypeMemory,
	}

	if config.Type != StorageTypeSQLite {
		t.Errorf("Type field not set correctly")
	}
	if config.SQLitePath != "/test/path.db" {
		t.Errorf("SQLitePath field not set correctly")
	}
	if config.SQLiteTimeout != 3000 {
		t.Errorf("SQLiteTimeout field not set correctly")
	}
	if config.FallbackType != StorageTypeMemory {
		t.Errorf("FallbackType field not set correctly")
	}
}

func TestConfigFromEnvEmptyStorageType(t *testing.T) {
	originalStorageType := os.Getenv("STORAGE_TYPE")
	defer os.Setenv("STORAGE_TYPE", originalStorageType)

	// Set empty storage type
	os.Setenv("STORAGE_TYPE", "")

	config := ConfigFromEnv()

	// Should fall back to default (memory)
	if config.Type != StorageTypeMemory {
		t.Errorf("Type = %v, want memory (default)", config.Type)
	}
}

func TestConfigFromEnvEmptySQLitePath(t *testing.T) {
	originalSQLitePath := os.Getenv("SQLITE_PATH")
	defer os.Setenv("SQLITE_PATH", originalSQLitePath)

	// Set empty SQLite path
	os.Setenv("SQLITE_PATH", "")

	config := ConfigFromEnv()

	// Should use default path
	if config.SQLitePath != "./data/unified-thinking.db" {
		t.Errorf("SQLitePath = %v, want default", config.SQLitePath)
	}
}

func TestConfigFromEnvEmptyTimeout(t *testing.T) {
	originalSQLiteTimeout := os.Getenv("SQLITE_TIMEOUT")
	defer os.Setenv("SQLITE_TIMEOUT", originalSQLiteTimeout)

	// Set empty timeout
	os.Setenv("SQLITE_TIMEOUT", "")

	config := ConfigFromEnv()

	// Should use default timeout
	if config.SQLiteTimeout != 5000 {
		t.Errorf("SQLiteTimeout = %v, want 5000 (default)", config.SQLiteTimeout)
	}
}

func TestConfigFromEnvEmptyFallback(t *testing.T) {
	originalFallback := os.Getenv("STORAGE_FALLBACK")
	defer os.Setenv("STORAGE_FALLBACK", originalFallback)

	// Set empty fallback
	os.Setenv("STORAGE_FALLBACK", "")

	config := ConfigFromEnv()

	// Should use default fallback
	if config.FallbackType != StorageTypeMemory {
		t.Errorf("FallbackType = %v, want memory (default)", config.FallbackType)
	}
}

func TestConfigFromEnvCustomStorageType(t *testing.T) {
	originalStorageType := os.Getenv("STORAGE_TYPE")
	defer os.Setenv("STORAGE_TYPE", originalStorageType)

	// Set custom storage type (might not be valid, but should be accepted)
	os.Setenv("STORAGE_TYPE", "custom")

	config := ConfigFromEnv()

	// Should accept custom value (validation happens in factory)
	if config.Type != StorageType("custom") {
		t.Errorf("Type = %v, want 'custom'", config.Type)
	}
}

func TestConfigFromEnvWithRelativePath(t *testing.T) {
	originalSQLitePath := os.Getenv("SQLITE_PATH")
	defer os.Setenv("SQLITE_PATH", originalSQLitePath)

	os.Setenv("SQLITE_PATH", "data/test.db")

	config := ConfigFromEnv()

	if config.SQLitePath != "data/test.db" {
		t.Errorf("SQLitePath = %v, want 'data/test.db'", config.SQLitePath)
	}
}

func TestConfigFromEnvWithAbsolutePath(t *testing.T) {
	originalSQLitePath := os.Getenv("SQLITE_PATH")
	defer os.Setenv("SQLITE_PATH", originalSQLitePath)

	tempDir := t.TempDir()
	absolutePath := filepath.Join(tempDir, "test.db")

	os.Setenv("SQLITE_PATH", absolutePath)

	config := ConfigFromEnv()

	if config.SQLitePath != absolutePath {
		t.Errorf("SQLitePath = %v, want %v", config.SQLitePath, absolutePath)
	}
}

func TestConfigFromEnvLargeTimeout(t *testing.T) {
	originalSQLiteTimeout := os.Getenv("SQLITE_TIMEOUT")
	defer os.Setenv("SQLITE_TIMEOUT", originalSQLiteTimeout)

	// Set very large timeout
	os.Setenv("SQLITE_TIMEOUT", "3600000") // 1 hour

	config := ConfigFromEnv()

	if config.SQLiteTimeout != 3600000 {
		t.Errorf("SQLiteTimeout = %v, want 3600000", config.SQLiteTimeout)
	}
}

func TestConfigFromEnvMemoryFallback(t *testing.T) {
	originalFallback := os.Getenv("STORAGE_FALLBACK")
	defer os.Setenv("STORAGE_FALLBACK", originalFallback)

	os.Setenv("STORAGE_FALLBACK", "memory")

	config := ConfigFromEnv()

	if config.FallbackType != StorageTypeMemory {
		t.Errorf("FallbackType = %v, want memory", config.FallbackType)
	}
}

func TestConfigFromEnvSQLiteFallback(t *testing.T) {
	originalFallback := os.Getenv("STORAGE_FALLBACK")
	defer os.Setenv("STORAGE_FALLBACK", originalFallback)

	os.Setenv("STORAGE_FALLBACK", "sqlite")

	config := ConfigFromEnv()

	if config.FallbackType != StorageTypeSQLite {
		t.Errorf("FallbackType = %v, want sqlite", config.FallbackType)
	}
}
