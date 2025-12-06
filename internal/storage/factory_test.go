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

func TestNewStorageFailFast(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "sqlite fails with invalid path (fail-fast)",
			config: Config{
				Type:       StorageTypeSQLite,
				SQLitePath: "/invalid/\x00/path/test.db", // Invalid path
			},
			wantErr: true,
		},
		{
			name: "sqlite fails with empty path (fail-fast)",
			config: Config{
				Type:       StorageTypeSQLite,
				SQLitePath: "",
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

			// With fail-fast behavior, storage should be nil on error
			if tt.wantErr && storage != nil {
				t.Error("Expected nil storage on error with fail-fast behavior")
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

	// Restore original env vars after test
	defer func() {
		os.Setenv("STORAGE_TYPE", originalStorageType)
		os.Setenv("SQLITE_PATH", originalSQLitePath)
		os.Setenv("SQLITE_TIMEOUT", originalSQLiteTimeout)
	}()

	// SQLite is now the default - tests need temp dirs for each test case
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		envVars  map[string]string
		wantErr  bool
		wantType string
	}{
		{
			name: "default (sqlite storage - required by knowledge graph)",
			envVars: map[string]string{
				// SQLite is default, must provide valid path
				"SQLITE_PATH": filepath.Join(tempDir, "default-test.db"),
			},
			wantErr:  false,
			wantType: "*storage.SQLiteStorage",
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
				"STORAGE_TYPE":   "sqlite",
				"SQLITE_PATH":    filepath.Join(tempDir, "env-test.db"),
				"SQLITE_TIMEOUT": "3000",
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
	// Test that factory fails fast with invalid SQLite paths
	config := Config{
		Type:       StorageTypeSQLite,
		SQLitePath: "", // Empty path
	}

	// This should fail with fail-fast behavior
	storage, err := NewStorage(config)

	if err == nil {
		t.Error("NewStorage() should error with invalid path (fail-fast behavior)")
	}

	if storage != nil {
		t.Error("NewStorage() should return nil on error")
		CloseStorage(storage)
	}
}

func TestFactoryFailFastBehavior(t *testing.T) {
	// Test that factory fails immediately without attempting fallback
	config := Config{
		Type:       StorageTypeSQLite,
		SQLitePath: "/invalid/\x00/path/test.db",
	}

	storage, err := NewStorage(config)

	// Should error immediately (fail-fast)
	if err == nil {
		t.Error("NewStorage() should error immediately with invalid path")
	}

	if storage != nil {
		t.Error("NewStorage() should return nil on error")
		CloseStorage(storage)
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
