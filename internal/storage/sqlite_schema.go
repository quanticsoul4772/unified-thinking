// Package storage provides SQLite schema definitions and migrations.
package storage

import (
	"database/sql"
	"fmt"
)

const schemaVersion = 7 // Updated to add Thompson Sampling RL tables

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
-- Note: No foreign key to trajectories table since episodic memory is in-memory
CREATE TABLE IF NOT EXISTS context_signatures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trajectory_id TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    fingerprint_prefix TEXT NOT NULL,  -- First 8 chars for indexing
    domain TEXT,
    key_concepts TEXT,      -- JSON array
    tool_sequence TEXT,     -- JSON array
    complexity REAL,
    embedding BLOB,         -- Semantic embedding for similarity (serialized float32 vector)
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- Indexes for context signature lookups
CREATE INDEX IF NOT EXISTS idx_context_domain ON context_signatures(domain);
CREATE INDEX IF NOT EXISTS idx_context_prefix ON context_signatures(fingerprint_prefix);
CREATE INDEX IF NOT EXISTS idx_context_trajectory ON context_signatures(trajectory_id);

-- Trajectories table for episodic memory persistence
-- Stores complete trajectory as JSON to avoid import cycles
CREATE TABLE IF NOT EXISTS trajectories (
    id TEXT PRIMARY KEY,
    trajectory_json TEXT NOT NULL,   -- Complete ReasoningTrajectory as JSON
    created_at INTEGER NOT NULL
);

-- Index for retrieval
CREATE INDEX IF NOT EXISTS idx_trajectories_created ON trajectories(created_at DESC);

-- Thompson Sampling RL tables for adaptive mode selection
-- Strategies: Available reasoning strategies
CREATE TABLE IF NOT EXISTS rl_strategies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    mode TEXT NOT NULL,
    parameters TEXT,
    created_at INTEGER NOT NULL,
    is_active INTEGER DEFAULT 1
);

-- Strategy outcomes: Historical performance data
CREATE TABLE IF NOT EXISTS rl_strategy_outcomes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    strategy_id TEXT NOT NULL,
    problem_id TEXT NOT NULL,
    problem_type TEXT,
    problem_description TEXT,
    success INTEGER NOT NULL,
    confidence_before REAL,
    confidence_after REAL,
    execution_time_ns INTEGER,
    token_count INTEGER,
    reasoning_path TEXT,
    timestamp INTEGER NOT NULL,
    metadata TEXT,
    FOREIGN KEY (strategy_id) REFERENCES rl_strategies(id)
);

-- Thompson state: Beta distribution parameters
CREATE TABLE IF NOT EXISTS rl_thompson_state (
    strategy_id TEXT PRIMARY KEY,
    alpha REAL NOT NULL DEFAULT 1.0,
    beta REAL NOT NULL DEFAULT 1.0,
    total_trials INTEGER DEFAULT 0,
    total_successes INTEGER DEFAULT 0,
    last_updated INTEGER NOT NULL,
    FOREIGN KEY (strategy_id) REFERENCES rl_strategies(id)
);

-- Indexes for RL queries
CREATE INDEX IF NOT EXISTS idx_rl_outcomes_strategy ON rl_strategy_outcomes(strategy_id);
CREATE INDEX IF NOT EXISTS idx_rl_outcomes_type ON rl_strategy_outcomes(problem_type);
CREATE INDEX IF NOT EXISTS idx_rl_outcomes_timestamp ON rl_strategy_outcomes(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_rl_outcomes_success ON rl_strategy_outcomes(success);

-- View for strategy performance monitoring
CREATE VIEW IF NOT EXISTS rl_strategy_performance AS
SELECT
    s.id,
    s.name,
    s.mode,
    COALESCE(ts.total_trials, 0) as trials,
    COALESCE(ts.total_successes, 0) as successes,
    COALESCE(CAST(ts.total_successes AS REAL) / NULLIF(ts.total_trials, 0), 0.0) as success_rate,
    ts.alpha,
    ts.beta,
    ts.last_updated
FROM rl_strategies s
LEFT JOIN rl_thompson_state ts ON s.id = ts.strategy_id
WHERE s.is_active = 1;

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
		-- Note: This version had a foreign key that caused issues
		CREATE TABLE IF NOT EXISTS context_signatures (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			trajectory_id TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			fingerprint_prefix TEXT NOT NULL,
			domain TEXT,
			key_concepts TEXT,
			tool_sequence TEXT,
			complexity REAL,
			created_at INTEGER DEFAULT (strftime('%s', 'now'))
		);

		CREATE INDEX IF NOT EXISTS idx_context_domain ON context_signatures(domain);
		CREATE INDEX IF NOT EXISTS idx_context_prefix ON context_signatures(fingerprint_prefix);
		CREATE INDEX IF NOT EXISTS idx_context_trajectory ON context_signatures(trajectory_id);
		`

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply v2->v3 migration: %w", err)
		}
	}

	// Migration from v3 to v4: Remove foreign key constraint from context_signatures
	// The foreign key referenced trajectories(id) but episodic memory stores trajectories in-memory
	if fromVersion < 4 && toVersion >= 4 {
		migration := `
		-- Recreate context_signatures without foreign key constraint
		-- SQLite doesn't support ALTER TABLE DROP CONSTRAINT, so we recreate
		CREATE TABLE IF NOT EXISTS context_signatures_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			trajectory_id TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			fingerprint_prefix TEXT NOT NULL,
			domain TEXT,
			key_concepts TEXT,
			tool_sequence TEXT,
			complexity REAL,
			created_at INTEGER DEFAULT (strftime('%s', 'now'))
		);

		-- Copy existing data if any
		INSERT OR IGNORE INTO context_signatures_new
			SELECT id, trajectory_id, fingerprint, fingerprint_prefix, domain,
			       key_concepts, tool_sequence, complexity, created_at
			FROM context_signatures;

		-- Drop old table and indexes
		DROP TABLE IF EXISTS context_signatures;

		-- Rename new table
		ALTER TABLE context_signatures_new RENAME TO context_signatures;

		-- Recreate indexes
		CREATE INDEX IF NOT EXISTS idx_context_domain ON context_signatures(domain);
		CREATE INDEX IF NOT EXISTS idx_context_prefix ON context_signatures(fingerprint_prefix);
		CREATE INDEX IF NOT EXISTS idx_context_trajectory ON context_signatures(trajectory_id);
		`

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply v3->v4 migration: %w", err)
		}
	}

	// Migration from v4 to v5: Add embedding column to context_signatures
	if fromVersion < 5 && toVersion >= 5 {
		migration := `
		-- Add embedding column for semantic similarity (v5)
		ALTER TABLE context_signatures ADD COLUMN embedding BLOB;
		`

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply v4->v5 migration: %w", err)
		}
	}

	// Migration from v5 to v6: Add trajectories table for episodic memory persistence
	if fromVersion < 6 && toVersion >= 6 {
		migration := `
		-- Trajectories table for episodic memory persistence (v6)
		-- Stores complete trajectory as JSON to avoid import cycles
		CREATE TABLE IF NOT EXISTS trajectories (
			id TEXT PRIMARY KEY,
			trajectory_json TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_trajectories_created ON trajectories(created_at DESC);
		`

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply v5->v6 migration: %w", err)
		}
	}

	// Migration from v6 to v7: Add Thompson Sampling RL tables
	if fromVersion < 7 && toVersion >= 7 {
		migration := `
		-- Thompson Sampling RL tables (v7)
		CREATE TABLE IF NOT EXISTS rl_strategies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			mode TEXT NOT NULL,
			parameters TEXT,
			created_at INTEGER NOT NULL,
			is_active INTEGER DEFAULT 1
		);

		CREATE TABLE IF NOT EXISTS rl_strategy_outcomes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			strategy_id TEXT NOT NULL,
			problem_id TEXT NOT NULL,
			problem_type TEXT,
			problem_description TEXT,
			success INTEGER NOT NULL,
			confidence_before REAL,
			confidence_after REAL,
			execution_time_ns INTEGER,
			token_count INTEGER,
			reasoning_path TEXT,
			timestamp INTEGER NOT NULL,
			metadata TEXT,
			FOREIGN KEY (strategy_id) REFERENCES rl_strategies(id)
		);

		CREATE TABLE IF NOT EXISTS rl_thompson_state (
			strategy_id TEXT PRIMARY KEY,
			alpha REAL NOT NULL DEFAULT 1.0,
			beta REAL NOT NULL DEFAULT 1.0,
			total_trials INTEGER DEFAULT 0,
			total_successes INTEGER DEFAULT 0,
			last_updated INTEGER NOT NULL,
			FOREIGN KEY (strategy_id) REFERENCES rl_strategies(id)
		);

		CREATE INDEX IF NOT EXISTS idx_rl_outcomes_strategy ON rl_strategy_outcomes(strategy_id);
		CREATE INDEX IF NOT EXISTS idx_rl_outcomes_type ON rl_strategy_outcomes(problem_type);
		CREATE INDEX IF NOT EXISTS idx_rl_outcomes_timestamp ON rl_strategy_outcomes(timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_rl_outcomes_success ON rl_strategy_outcomes(success);

		CREATE VIEW IF NOT EXISTS rl_strategy_performance AS
		SELECT
			s.id,
			s.name,
			s.mode,
			COALESCE(ts.total_trials, 0) as trials,
			COALESCE(ts.total_successes, 0) as successes,
			COALESCE(CAST(ts.total_successes AS REAL) / NULLIF(ts.total_trials, 0), 0.0) as success_rate,
			ts.alpha,
			ts.beta,
			ts.last_updated
		FROM rl_strategies s
		LEFT JOIN rl_thompson_state ts ON s.id = ts.strategy_id
		WHERE s.is_active = 1;

		-- Initialize default strategies
		INSERT OR IGNORE INTO rl_strategies (id, name, description, mode, parameters, created_at, is_active) VALUES
		('strategy_linear', 'Linear Sequential', 'Step-by-step systematic reasoning', 'linear', '{}', strftime('%s', 'now'), 1),
		('strategy_tree', 'Tree Exploration', 'Multi-branch parallel exploration', 'tree', '{}', strftime('%s', 'now'), 1),
		('strategy_divergent', 'Divergent Creative', 'Creative unconventional thinking', 'divergent', '{"force_rebellion": false}', strftime('%s', 'now'), 1),
		('strategy_reflection', 'Reflective Analysis', 'Metacognitive reflection', 'reflection', '{}', strftime('%s', 'now'), 1),
		('strategy_backtracking', 'Checkpoint Backtracking', 'Iterative refinement with rollback', 'backtracking', '{}', strftime('%s', 'now'), 1);

		-- Initialize Thompson state with uniform priors
		INSERT OR IGNORE INTO rl_thompson_state (strategy_id, alpha, beta, total_trials, total_successes, last_updated) VALUES
		('strategy_linear', 1.0, 1.0, 0, 0, strftime('%s', 'now')),
		('strategy_tree', 1.0, 1.0, 0, 0, strftime('%s', 'now')),
		('strategy_divergent', 1.0, 1.0, 0, 0, strftime('%s', 'now')),
		('strategy_reflection', 1.0, 1.0, 0, 0, strftime('%s', 'now')),
		('strategy_backtracking', 1.0, 1.0, 0, 0, strftime('%s', 'now'));
		`

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply v6->v7 migration: %w", err)
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
