// Package storage provides configuration for storage backends.
package storage

import (
	"os"
	"path/filepath"
	"strconv"
)

// StorageType represents the type of storage backend
type StorageType string

const (
	// StorageTypeMemory uses in-memory storage (default)
	StorageTypeMemory StorageType = "memory"
	// StorageTypeSQLite uses SQLite persistent storage
	StorageTypeSQLite StorageType = "sqlite"
)

// Config holds storage configuration
type Config struct {
	Type          StorageType // Storage backend type
	SQLitePath    string      // Path to SQLite database file
	SQLiteTimeout int         // SQLite busy timeout in milliseconds
	FallbackType  StorageType // Fallback storage type on errors
}

// DefaultConfig returns default configuration with in-memory storage
func DefaultConfig() Config {
	return Config{
		Type:          StorageTypeMemory,
		SQLitePath:    "./data/unified-thinking.db",
		SQLiteTimeout: 5000,
		FallbackType:  StorageTypeMemory,
	}
}

// ConfigFromEnv reads storage configuration from environment variables
// Supports:
//   - STORAGE_TYPE: "memory" (default) or "sqlite"
//   - SQLITE_PATH: Path to SQLite database file
//   - SQLITE_TIMEOUT: Busy timeout in milliseconds
//   - STORAGE_FALLBACK: Fallback storage type on errors
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	// Read STORAGE_TYPE
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		cfg.Type = StorageType(storageType)
	}

	// Read SQLITE_PATH
	if sqlitePath := os.Getenv("SQLITE_PATH"); sqlitePath != "" {
		cfg.SQLitePath = sqlitePath
	}

	// Ensure parent directory exists for SQLite
	if cfg.Type == StorageTypeSQLite {
		dir := filepath.Dir(cfg.SQLitePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			// Log warning but don't fail - factory will handle this
		}
	}

	// Read SQLITE_TIMEOUT
	if timeout := os.Getenv("SQLITE_TIMEOUT"); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil && val > 0 {
			cfg.SQLiteTimeout = val
		}
	}

	// Read STORAGE_FALLBACK
	if fallback := os.Getenv("STORAGE_FALLBACK"); fallback != "" {
		cfg.FallbackType = StorageType(fallback)
	}

	return cfg
}
