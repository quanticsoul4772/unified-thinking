package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantErr  bool
		wantType string
	}{
		{
			name: "memory storage",
			config: Config{
				Type: StorageTypeMemory,
			},
			wantErr:  false,
			wantType: "*storage.MemoryStorage",
		},
		{
			name: "sqlite storage",
			config: Config{
				Type:          StorageTypeSQLite,
				SQLitePath:    filepath.Join(t.TempDir(), "factory-test.db"),
				SQLiteTimeout: 5000,
			},
			wantErr:  false,
			wantType: "*storage.SQLiteStorage",
		},
		{
			name: "unknown storage type",
			config: Config{
				Type: StorageType("unknown"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewStorage(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if storage == nil {
					t.Error("NewStorage() returned nil storage")
				}

				// Verify type
				typeName := getTypeName(storage)
				if typeName != tt.wantType {
					t.Errorf("Storage type = %v, want %v", typeName, tt.wantType)
				}

				// Clean up
				CloseStorage(storage)
			}
		})
	}
}

func TestNewStorageFallback(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		wantErr      bool
		wantFallback bool
	}{
		{
			name: "sqlite fails with fallback to memory",
			config: Config{
				Type:         StorageTypeSQLite,
				SQLitePath:   "/invalid/\x00/path/test.db", // Invalid path
				FallbackType: StorageTypeMemory,
			},
			wantErr:      false,
			wantFallback: true,
		},
		{
			name: "sqlite fails without fallback",
			config: Config{
				Type:       StorageTypeSQLite,
				SQLitePath: "/invalid/\x00/path/test.db",
			},
			wantErr:      true,
			wantFallback: false,
		},
		{
			name: "sqlite fails with same type fallback",
			config: Config{
				Type:         StorageTypeSQLite,
				SQLitePath:   "/invalid/\x00/path/test.db",
				FallbackType: StorageTypeSQLite,
			},
			wantErr:      true,
			wantFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewStorage(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantFallback {
				// Verify it fell back to memory storage
				typeName := getTypeName(storage)
				if typeName != "*storage.MemoryStorage" {
					t.Errorf("Expected fallback to MemoryStorage, got %v", typeName)
				}
				CloseStorage(storage)
			}
		})
	}
}

func TestNewStorageFromEnv(t *testing.T) {
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
		wantErr  bool
		wantType string
	}{
		{
			name:     "default (memory storage)",
			envVars:  map[string]string{},
			wantErr:  false,
			wantType: "*storage.MemoryStorage",
		},
		{
			name: "memory storage from env",
			envVars: map[string]string{
				"STORAGE_TYPE": "memory",
			},
			wantErr:  false,
			wantType: "*storage.MemoryStorage",
		},
		{
			name: "sqlite storage from env",
			envVars: map[string]string{
				"STORAGE_TYPE":     "sqlite",
				"SQLITE_PATH":      filepath.Join(t.TempDir(), "env-test.db"),
				"SQLITE_TIMEOUT":   "3000",
				"STORAGE_FALLBACK": "memory",
			},
			wantErr:  false,
			wantType: "*storage.SQLiteStorage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("STORAGE_TYPE")
			os.Unsetenv("SQLITE_PATH")
			os.Unsetenv("SQLITE_TIMEOUT")
			os.Unsetenv("STORAGE_FALLBACK")

			// Set env vars for this test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			storage, err := NewStorageFromEnv()

			if (err != nil) != tt.wantErr {
				t.Errorf("NewStorageFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if storage == nil {
					t.Error("NewStorageFromEnv() returned nil storage")
				}

				typeName := getTypeName(storage)
				if typeName != tt.wantType {
					t.Errorf("Storage type = %v, want %v", typeName, tt.wantType)
				}

				CloseStorage(storage)
			}
		})
	}
}

func TestCloseStorage(t *testing.T) {
	tests := []struct {
		name    string
		storage Storage
		wantErr bool
	}{
		{
			name:    "close memory storage (no-op)",
			storage: NewMemoryStorage(),
			wantErr: false,
		},
		{
			name: "close sqlite storage",
			storage: func() Storage {
				tempDir := t.TempDir()
				dbPath := filepath.Join(tempDir, "close-test.db")
				s, err := NewSQLiteStorage(dbPath, 5000)
				if err != nil {
					t.Fatalf("Failed to create SQLite storage: %v", err)
				}
				return s
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CloseStorage(tt.storage)

			if (err != nil) != tt.wantErr {
				t.Errorf("CloseStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloseStorageNil(t *testing.T) {
	// Test that CloseStorage handles storage without Close method
	var storage Storage = NewMemoryStorage()

	err := CloseStorage(storage)
	if err != nil {
		t.Errorf("CloseStorage() on memory storage should not error, got %v", err)
	}
}

func TestFactoryWithInvalidSQLitePath(t *testing.T) {
	// Test that factory handles invalid SQLite paths correctly
	config := Config{
		Type:         StorageTypeSQLite,
		SQLitePath:   "", // Empty path
		FallbackType: StorageTypeMemory,
	}

	// This should fail and fallback to memory
	storage, err := NewStorage(config)

	if err != nil {
		t.Errorf("NewStorage() should fallback, not error: %v", err)
	}

	if storage == nil {
		t.Fatal("NewStorage() returned nil")
	}

	// Verify it fell back to memory
	typeName := getTypeName(storage)
	if typeName != "*storage.MemoryStorage" {
		t.Errorf("Expected fallback to MemoryStorage, got %v", typeName)
	}

	CloseStorage(storage)
}

func TestFactoryRecursiveFallback(t *testing.T) {
	// Test that factory doesn't infinitely recurse with invalid fallback
	config := Config{
		Type:         StorageTypeSQLite,
		SQLitePath:   "/invalid/\x00/path/test.db",
		FallbackType: StorageTypeSQLite, // Same type as primary
	}

	_, err := NewStorage(config)

	// Should error, not fallback infinitely
	if err == nil {
		t.Error("NewStorage() should error with same type fallback")
	}
}

func TestFactoryCreatesDirectoryForSQLite(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "subdir", "test.db")

	// Directory doesn't exist yet
	config := Config{
		Type:          StorageTypeSQLite,
		SQLitePath:    dbPath,
		SQLiteTimeout: 5000,
	}

	// This should fail because parent directory doesn't exist
	// SQLite won't auto-create parent directories
	_, err := NewStorage(config)

	if err == nil {
		t.Error("NewStorage() should fail when parent directory doesn't exist")
	}
}

func TestFactorySQLiteWithCustomTimeout(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "timeout-test.db")

	config := Config{
		Type:          StorageTypeSQLite,
		SQLitePath:    dbPath,
		SQLiteTimeout: 10000, // Custom timeout
	}

	storage, err := NewStorage(config)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}
	defer CloseStorage(storage)

	if storage == nil {
		t.Error("NewStorage() returned nil")
	}

	// Verify it's SQLite storage
	sqliteStorage, ok := storage.(*SQLiteStorage)
	if !ok {
		t.Error("Storage is not SQLiteStorage")
	}

	if sqliteStorage.db == nil {
		t.Error("Database connection is nil")
	}
}

func TestFactoryMemoryStorageIsNotCloser(t *testing.T) {
	storage := NewMemoryStorage()

	// Verify MemoryStorage doesn't implement io.Closer
	_, implementsCloser := interface{}(storage).(interface{ Close() error })
	if implementsCloser {
		t.Error("MemoryStorage should not implement io.Closer")
	}

	// CloseStorage should still work (no-op)
	err := CloseStorage(storage)
	if err != nil {
		t.Errorf("CloseStorage() error = %v", err)
	}
}

func TestFactorySQLiteStorageIsCloser(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "closer-test.db")

	storage, err := NewSQLiteStorage(dbPath, 5000)
	if err != nil {
		t.Fatalf("NewSQLiteStorage() error = %v", err)
	}

	// Verify SQLiteStorage implements io.Closer
	_, implementsCloser := interface{}(storage).(interface{ Close() error })
	if !implementsCloser {
		t.Error("SQLiteStorage should implement io.Closer")
	}

	// Close should work
	err = CloseStorage(storage)
	if err != nil {
		t.Errorf("CloseStorage() error = %v", err)
	}
}

// Helper function to get type name as string
func getTypeName(i interface{}) string {
	if i == nil {
		return "nil"
	}

	switch i.(type) {
	case *MemoryStorage:
		return "*storage.MemoryStorage"
	case *SQLiteStorage:
		return "*storage.SQLiteStorage"
	default:
		return "unknown"
	}
}
