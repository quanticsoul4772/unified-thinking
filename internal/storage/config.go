// Package storage provides configuration for storage backends.
package storage

import (
	"log"
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
}

// DefaultConfig returns default configuration with in-memory storage
func DefaultConfig() Config {
	return Config{
		Type:          StorageTypeMemory,
		SQLitePath:    "./data/unified-thinking.db",
		SQLiteTimeout: 5000,
	}
}

// ConfigFromEnv reads storage configuration from environment variables
// Supports:
//   - STORAGE_TYPE: "memory" (default) or "sqlite"
//   - SQLITE_PATH: Path to SQLite database file
//   - SQLITE_TIMEOUT: Busy timeout in milliseconds
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
		if err := os.MkdirAll(dir, 0750); err != nil {
			log.Printf("warning: failed to create SQLite directory %s: %v (factory will handle this)", dir, err)
		}
	}

	// Read SQLITE_TIMEOUT
	if timeout := os.Getenv("SQLITE_TIMEOUT"); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil && val > 0 {
			cfg.SQLiteTimeout = val
		}
	}

	return cfg
}
