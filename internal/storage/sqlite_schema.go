// Package storage provides SQLite schema definitions and migrations.
package storage

import (
	"database/sql"
	"fmt"
)

const schemaVersion = 3 // Updated for context signatures support

// Schema defines the complete database schema
const schema = `
-- Schema metadata for versioning
CREATE TABLE IF NOT EXISTS schema_metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Branches table (MUST be created BEFORE thoughts due to foreign key constraint)
CREATE TABLE IF NOT EXISTS branches (
    id TEXT PRIMARY KEY,
    parent_branch_id TEXT,
    state TEXT NOT NULL,
    priority REAL NOT NULL DEFAULT 0.0,
    confidence REAL NOT NULL DEFAULT 0.0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    last_accessed_at INTEGER NOT NULL,
    FOREIGN KEY (parent_branch_id) REFERENCES branches(id) ON DELETE SET NULL
);

-- Thoughts table
CREATE TABLE IF NOT EXISTS thoughts (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    mode TEXT NOT NULL,
    branch_id TEXT,
    parent_id TEXT,
    type TEXT NOT NULL DEFAULT '',
    confidence REAL NOT NULL,
    timestamp INTEGER NOT NULL,
    key_points TEXT,
    metadata TEXT,
    is_rebellion INTEGER DEFAULT 0,
    challenges_assumption INTEGER DEFAULT 0,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE SET NULL
);

-- Insights table
CREATE TABLE IF NOT EXISTS insights (
    id TEXT PRIMARY KEY,
    branch_id TEXT,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    context TEXT,
    parent_insights TEXT,
    applicability_score REAL NOT NULL,
    supporting_evidence TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE SET NULL
);

-- Cross-references table
CREATE TABLE IF NOT EXISTS cross_refs (
    id TEXT PRIMARY KEY,
    from_branch TEXT NOT NULL,
    to_branch TEXT NOT NULL,
    type TEXT NOT NULL,
    reason TEXT NOT NULL,
    strength REAL NOT NULL,
    touchpoints TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (from_branch) REFERENCES branches(id) ON DELETE CASCADE,
    FOREIGN KEY (to_branch) REFERENCES branches(id) ON DELETE CASCADE
);

-- Validations table
CREATE TABLE IF NOT EXISTS validations (
    id TEXT PRIMARY KEY,
    insight_id TEXT,
    thought_id TEXT,
    is_valid INTEGER NOT NULL,
    validation_data TEXT,
    reason TEXT,
    created_at INTEGER NOT NULL
);

-- Relationships table
CREATE TABLE IF NOT EXISTS relationships (
    id TEXT PRIMARY KEY,
    from_state_id TEXT NOT NULL,
    to_state_id TEXT NOT NULL,
    type TEXT NOT NULL,
    metadata TEXT,
    created_at INTEGER NOT NULL
);

-- Embeddings table for vector storage
CREATE TABLE IF NOT EXISTS embeddings (
    id TEXT PRIMARY KEY,
    problem_id TEXT NOT NULL UNIQUE,  -- References episodic memory problem descriptions
    embedding BLOB NOT NULL,           -- Serialized float32 vector
    model TEXT NOT NULL,               -- e.g., 'voyage-3-lite'
    provider TEXT NOT NULL,            -- e.g., 'voyage'
    dimension INTEGER NOT NULL,        -- Vector dimension (e.g., 512)
    source TEXT NOT NULL,              -- What was embedded (e.g., 'description+context+goals')
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Index for fast problem_id lookups
CREATE INDEX IF NOT EXISTS idx_embeddings_problem ON embeddings(problem_id);
CREATE INDEX IF NOT EXISTS idx_embeddings_created ON embeddings(created_at DESC);

-- Context signatures table for cross-session bridging
CREATE TABLE IF NOT EXISTS context_signatures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trajectory_id TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    fingerprint_prefix TEXT NOT NULL,  -- First 8 chars for indexing
    domain TEXT,
    key_concepts TEXT,      -- JSON array
    tool_sequence TEXT,     -- JSON array
    complexity REAL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (trajectory_id) REFERENCES trajectories(id) ON DELETE CASCADE
);

-- Indexes for context signature lookups
CREATE INDEX IF NOT EXISTS idx_context_domain ON context_signatures(domain);
CREATE INDEX IF NOT EXISTS idx_context_prefix ON context_signatures(fingerprint_prefix);
CREATE INDEX IF NOT EXISTS idx_context_trajectory ON context_signatures(trajectory_id);

-- Full-text search index for thought content
CREATE VIRTUAL TABLE IF NOT EXISTS thoughts_fts USING fts5(
    id UNINDEXED,
    content,
    content='thoughts',
    content_rowid='rowid'
);

-- Triggers to keep FTS index synchronized
CREATE TRIGGER IF NOT EXISTS thoughts_fts_insert AFTER INSERT ON thoughts BEGIN
    INSERT INTO thoughts_fts(rowid, id, content) VALUES (new.rowid, new.id, new.content);
END;

CREATE TRIGGER IF NOT EXISTS thoughts_fts_update AFTER UPDATE ON thoughts BEGIN
    UPDATE thoughts_fts SET content = new.content WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS thoughts_fts_delete AFTER DELETE ON thoughts BEGIN
    DELETE FROM thoughts_fts WHERE rowid = old.rowid;
END;

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_thoughts_mode ON thoughts(mode);
CREATE INDEX IF NOT EXISTS idx_thoughts_branch ON thoughts(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_thoughts_timestamp ON thoughts(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_insights_branch ON insights(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_branches_accessed ON branches(last_accessed_at DESC);
CREATE INDEX IF NOT EXISTS idx_branches_priority ON branches(priority DESC);
CREATE INDEX IF NOT EXISTS idx_crossrefs_from ON cross_refs(from_branch);
CREATE INDEX IF NOT EXISTS idx_crossrefs_to ON cross_refs(to_branch);
`

// initializeSchema creates all tables and indexes
func initializeSchema(db *sql.DB) error {
	// Execute schema
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Check schema version
	var currentVersion int
	err := db.QueryRow("SELECT value FROM schema_metadata WHERE key = 'version'").Scan(&currentVersion)
	if err == sql.ErrNoRows {
		// First time initialization
		_, err = db.Exec("INSERT INTO schema_metadata (key, value) VALUES ('version', ?)", schemaVersion)
		if err != nil {
			return fmt.Errorf("failed to set schema version: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query schema version: %w", err)
	} else if currentVersion < schemaVersion {
		// Run migrations
		if err := runMigrations(db, currentVersion, schemaVersion); err != nil {
			return fmt.Errorf("failed to run migrations from v%d to v%d: %w", currentVersion, schemaVersion, err)
		}
		// Update version
		_, err = db.Exec("UPDATE schema_metadata SET value = ? WHERE key = 'version'", schemaVersion)
		if err != nil {
			return fmt.Errorf("failed to update schema version: %w", err)
		}
	} else if currentVersion > schemaVersion {
		return fmt.Errorf("database version (%d) is newer than application version (%d)", currentVersion, schemaVersion)
	}

	return nil
}

// runMigrations applies database migrations
func runMigrations(db *sql.DB, fromVersion, toVersion int) error {
	// Migration from v1 to v2: Add embeddings table
	if fromVersion < 2 && toVersion >= 2 {
		migration := `
		-- Embeddings table for vector storage (v2)
		CREATE TABLE IF NOT EXISTS embeddings (
			id TEXT PRIMARY KEY,
			problem_id TEXT NOT NULL UNIQUE,
			embedding BLOB NOT NULL,
			model TEXT NOT NULL,
			provider TEXT NOT NULL,
			dimension INTEGER NOT NULL,
			source TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_embeddings_problem ON embeddings(problem_id);
		CREATE INDEX IF NOT EXISTS idx_embeddings_created ON embeddings(created_at DESC);
		`

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply v1->v2 migration: %w", err)
		}
	}

	// Migration from v2 to v3: Add context signatures table
	if fromVersion < 3 && toVersion >= 3 {
		migration := `
		-- Context signatures table for cross-session bridging (v3)
		CREATE TABLE IF NOT EXISTS context_signatures (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			trajectory_id TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			fingerprint_prefix TEXT NOT NULL,
			domain TEXT,
			key_concepts TEXT,
			tool_sequence TEXT,
			complexity REAL,
			created_at INTEGER DEFAULT (strftime('%s', 'now')),
			FOREIGN KEY (trajectory_id) REFERENCES trajectories(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_context_domain ON context_signatures(domain);
		CREATE INDEX IF NOT EXISTS idx_context_prefix ON context_signatures(fingerprint_prefix);
		CREATE INDEX IF NOT EXISTS idx_context_trajectory ON context_signatures(trajectory_id);
		`

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply v2->v3 migration: %w", err)
		}
	}

	return nil
}

// configureSQLite sets optimal pragmas for performance and safety
func configureSQLite(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",        // Write-Ahead Logging for concurrent reads
		"PRAGMA synchronous = NORMAL",      // Balance safety vs performance
		"PRAGMA cache_size = -64000",       // 64MB cache
		"PRAGMA foreign_keys = ON",         // Enforce referential integrity
		"PRAGMA temp_store = MEMORY",       // Keep temp tables in memory
		"PRAGMA mmap_size = 268435456",     // 256MB memory-mapped I/O
		"PRAGMA page_size = 8192",          // 8KB page size
		"PRAGMA auto_vacuum = INCREMENTAL", // Incremental vacuum mode
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}

	return nil
}
