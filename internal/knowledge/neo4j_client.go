// Package knowledge provides knowledge graph integration using Neo4j and semantic search.
package knowledge

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
)

// Neo4jClient manages connections to Neo4j database
type Neo4jClient struct {
	driver  neo4j.DriverWithContext
	uri     string
	timeout time.Duration
}

// Neo4jConfig holds Neo4j connection configuration
type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
	Timeout  time.Duration
}

// DefaultConfig returns default Neo4j configuration from environment variables
func DefaultConfig() Neo4jConfig {
	cfg := Neo4jConfig{
		URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Username: getEnv("NEO4J_USERNAME", "neo4j"),
		Password: getEnv("NEO4J_PASSWORD", "password"),
		Database: getEnv("NEO4J_DATABASE", "neo4j"),
		Timeout:  5 * time.Second,
	}

	if timeoutStr := os.Getenv("NEO4J_TIMEOUT_MS"); timeoutStr != "" {
		if ms, err := strconv.Atoi(timeoutStr); err == nil && ms > 0 {
			cfg.Timeout = time.Duration(ms) * time.Millisecond
		}
	}

	return cfg
}

// NewNeo4jClient creates a new Neo4j client with connection pooling
func NewNeo4jClient(cfg Neo4jConfig) (*Neo4jClient, error) {
	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.Username, cfg.Password, ""),
		func(config *config.Config) {
			config.MaxConnectionPoolSize = 50
			config.ConnectionAcquisitionTimeout = cfg.Timeout
			config.SocketConnectTimeout = cfg.Timeout
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	client := &Neo4jClient{
		driver:  driver,
		uri:     cfg.URI,
		timeout: cfg.Timeout,
	}

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	return client, nil
}

// Close closes the Neo4j driver and releases resources
func (c *Neo4jClient) Close(ctx context.Context) error {
	if c.driver != nil {
		return c.driver.Close(ctx)
	}
	return nil
}

// ExecuteWrite executes a write transaction with retry logic
func (c *Neo4jClient) ExecuteWrite(ctx context.Context, database string, work neo4j.ManagedTransactionWork) (interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer func() { _ = session.Close(ctx) }()

	return session.ExecuteWrite(ctx, work)
}

// ExecuteRead executes a read transaction with retry logic
func (c *Neo4jClient) ExecuteRead(ctx context.Context, database string, work neo4j.ManagedTransactionWork) (interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
		AccessMode:   neo4j.AccessModeRead,
	})
	defer func() { _ = session.Close(ctx) }()

	return session.ExecuteRead(ctx, work)
}

// Run executes a single Cypher query (convenience method for simple queries)
func (c *Neo4jClient) Run(ctx context.Context, database string, query string, params map[string]interface{}) (neo4j.ResultWithContext, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer func() { _ = session.Close(ctx) }()

	return session.Run(ctx, query, params)
}

// VerifyConnectivity checks if the client can connect to Neo4j
func (c *Neo4jClient) VerifyConnectivity(ctx context.Context) error {
	return c.driver.VerifyConnectivity(ctx)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
