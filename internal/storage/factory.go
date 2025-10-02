// Package storage provides factory for creating storage backends.
package storage

import (
	"fmt"
	"io"
	"log"
)

// NewStorage creates a storage backend based on configuration
// Implements graceful fallback to memory storage on errors
func NewStorage(cfg Config) (Storage, error) {
	switch cfg.Type {
	case StorageTypeMemory:
		log.Println("Initializing in-memory storage")
		return NewMemoryStorage(), nil

	case StorageTypeSQLite:
		log.Printf("Initializing SQLite storage at %s", cfg.SQLitePath)
		sqliteStore, err := NewSQLiteStorage(cfg.SQLitePath, cfg.SQLiteTimeout)
		if err != nil {
			if cfg.FallbackType != "" && cfg.FallbackType != cfg.Type {
				log.Printf("SQLite initialization failed: %v. Falling back to %s", err, cfg.FallbackType)
				return NewStorage(Config{Type: cfg.FallbackType})
			}
			return nil, fmt.Errorf("sqlite initialization failed: %w", err)
		}
		return sqliteStore, nil

	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.Type)
	}
}

// NewStorageFromEnv creates storage from environment variables
// This is the recommended way to initialize storage for MCP servers
func NewStorageFromEnv() (Storage, error) {
	cfg := ConfigFromEnv()
	return NewStorage(cfg)
}

// CloseStorage safely closes storage if it implements io.Closer
func CloseStorage(s Storage) error {
	if closer, ok := s.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
