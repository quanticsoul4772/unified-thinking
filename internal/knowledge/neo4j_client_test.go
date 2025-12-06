package knowledge

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestDefaultConfig tests default configuration from environment
func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected Neo4jConfig
	}{
		{
			name: "default values",
			env:  map[string]string{},
			expected: Neo4jConfig{
				URI:      "bolt://localhost:7687",
				Username: "neo4j",
				Password: "password",
				Database: "neo4j",
				Timeout:  5 * time.Second,
			},
		},
		{
			name: "custom values from env",
			env: map[string]string{
				"NEO4J_URI":        "bolt://remote:7687",
				"NEO4J_USERNAME":   "admin",
				"NEO4J_PASSWORD":   "secret",
				"NEO4J_DATABASE":   "graph",
				"NEO4J_TIMEOUT_MS": "10000",
			},
			expected: Neo4jConfig{
				URI:      "bolt://remote:7687",
				Username: "admin",
				Password: "secret",
				Database: "graph",
				Timeout:  10 * time.Second,
			},
		},
		{
			name: "invalid timeout falls back to default",
			env: map[string]string{
				"NEO4J_TIMEOUT_MS": "invalid",
			},
			expected: Neo4jConfig{
				URI:      "bolt://localhost:7687",
				Username: "neo4j",
				Password: "password",
				Database: "neo4j",
				Timeout:  5 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear Neo4j env vars first to ensure clean state
			neo4jVars := []string{"NEO4J_URI", "NEO4J_USERNAME", "NEO4J_PASSWORD", "NEO4J_DATABASE", "NEO4J_TIMEOUT_MS"}
			original := make(map[string]string)
			for _, k := range neo4jVars {
				original[k] = os.Getenv(k)
				os.Unsetenv(k)
			}
			defer func() {
				for k, v := range original {
					if v != "" {
						os.Setenv(k, v)
					} else {
						os.Unsetenv(k)
					}
				}
			}()

			// Set test-specific environment
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			cfg := DefaultConfig()

			if cfg.URI != tt.expected.URI {
				t.Errorf("URI = %s, want %s", cfg.URI, tt.expected.URI)
			}
			if cfg.Username != tt.expected.Username {
				t.Errorf("Username = %s, want %s", cfg.Username, tt.expected.Username)
			}
			if cfg.Password != tt.expected.Password {
				t.Errorf("Password = %s, want %s", cfg.Password, tt.expected.Password)
			}
			if cfg.Database != tt.expected.Database {
				t.Errorf("Database = %s, want %s", cfg.Database, tt.expected.Database)
			}
			if cfg.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout = %v, want %v", cfg.Timeout, tt.expected.Timeout)
			}
		})
	}
}

// TestNewNeo4jClient_ConnectionFailure tests connection error handling
func TestNewNeo4jClient_ConnectionFailure(t *testing.T) {
	cfg := Neo4jConfig{
		URI:      "bolt://nonexistent:7687",
		Username: "neo4j",
		Password: "password",
		Database: "neo4j",
		Timeout:  1 * time.Second,
	}

	client, err := NewNeo4jClient(cfg)
	if err == nil {
		if client != nil {
			client.Close(context.Background())
		}
		t.Skip("Test requires Neo4j to be unavailable at bolt://nonexistent:7687")
	}

	if client != nil {
		t.Error("Expected nil client on connection failure")
	}
}

// TestNeo4jClient_VerifyConnectivity requires a running Neo4j instance
func TestNeo4jClient_VerifyConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := DefaultConfig()
	client, err := NewNeo4jClient(cfg)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}
	defer client.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.VerifyConnectivity(ctx)
	if err != nil {
		t.Errorf("VerifyConnectivity failed: %v", err)
	}
}

// TestNeo4jClient_Close tests cleanup
func TestNeo4jClient_Close(t *testing.T) {
	client := &Neo4jClient{}

	ctx := context.Background()
	err := client.Close(ctx)
	if err != nil {
		t.Errorf("Close with nil driver should not error: %v", err)
	}
}
