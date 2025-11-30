// Package benchmarks provides SQLite storage for benchmark results.
package benchmarks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "modernc.org/sqlite"
)

// BenchmarkStorage manages persistent storage of benchmark results
type BenchmarkStorage struct {
	db *sql.DB
}

// NewBenchmarkStorage creates a new benchmark storage
func NewBenchmarkStorage(dbPath string) (*BenchmarkStorage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create schema
	schema := `
	CREATE TABLE IF NOT EXISTS benchmark_runs (
		run_id TEXT PRIMARY KEY,
		suite_name TEXT NOT NULL,
		git_commit TEXT,
		timestamp INTEGER NOT NULL,
		overall_accuracy REAL NOT NULL,
		overall_ece REAL NOT NULL,
		avg_latency_ns INTEGER NOT NULL,
		total_problems INTEGER NOT NULL,
		correct_problems INTEGER NOT NULL,
		total_tokens INTEGER NOT NULL,
		metadata TEXT
	);

	CREATE TABLE IF NOT EXISTS benchmark_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id TEXT NOT NULL,
		problem_id TEXT NOT NULL,
		correct INTEGER NOT NULL,
		score REAL NOT NULL,
		confidence REAL NOT NULL,
		latency_ns INTEGER NOT NULL,
		tokens INTEGER NOT NULL,
		mode TEXT NOT NULL,
		response TEXT,
		error TEXT,
		FOREIGN KEY (run_id) REFERENCES benchmark_runs(run_id)
	);

	CREATE INDEX IF NOT EXISTS idx_runs_timestamp ON benchmark_runs(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_runs_suite ON benchmark_runs(suite_name);
	CREATE INDEX IF NOT EXISTS idx_results_run ON benchmark_results(run_id);
	`

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &BenchmarkStorage{db: db}, nil
}

// StoreBenchmarkRun saves a benchmark run to the database
func (s *BenchmarkStorage) StoreBenchmarkRun(run *BenchmarkRun) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Ignore error - commit might have happened
	}()

	// Calculate total tokens
	totalTokens := 0
	for _, result := range run.Results {
		totalTokens += result.Tokens
	}

	// Serialize metadata
	metadataJSON, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert benchmark run
	_, err = tx.Exec(`
		INSERT INTO benchmark_runs (
			run_id, suite_name, git_commit, timestamp, overall_accuracy,
			overall_ece, avg_latency_ns, total_problems, correct_problems,
			total_tokens, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		run.RunID,
		run.SuiteName,
		run.GitCommit,
		run.Timestamp.Unix(),
		run.OverallAccuracy,
		run.OverallECE,
		run.AvgLatency.Nanoseconds(),
		run.TotalProblems,
		run.CorrectProblems,
		totalTokens,
		string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to insert run: %w", err)
	}

	// Insert individual results
	for _, result := range run.Results {
		_, err = tx.Exec(`
			INSERT INTO benchmark_results (
				run_id, problem_id, correct, score, confidence,
				latency_ns, tokens, mode, response, error
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			run.RunID,
			result.ProblemID,
			boolToInt(result.Correct),
			result.Score,
			result.Confidence,
			result.Latency.Nanoseconds(),
			result.Tokens,
			result.Mode,
			result.Response,
			result.Error,
		)
		if err != nil {
			return fmt.Errorf("failed to insert result for %s: %w", result.ProblemID, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLatestRun retrieves the most recent benchmark run for a suite
func (s *BenchmarkStorage) GetLatestRun(suiteName string) (*BenchmarkRun, error) {
	var runID, gitCommit, metadataJSON string
	var timestamp int64
	var overallAccuracy, overallECE float64
	var avgLatencyNs int64
	var totalProblems, correctProblems, totalTokens int

	err := s.db.QueryRow(`
		SELECT run_id, suite_name, git_commit, timestamp, overall_accuracy,
			overall_ece, avg_latency_ns, total_problems, correct_problems,
			total_tokens, metadata
		FROM benchmark_runs
		WHERE suite_name = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, suiteName).Scan(
		&runID, &suiteName, &gitCommit, &timestamp, &overallAccuracy,
		&overallECE, &avgLatencyNs, &totalProblems, &correctProblems,
		&totalTokens, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest run: %w", err)
	}

	var run BenchmarkRun
	if err := json.Unmarshal([]byte(metadataJSON), &run); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &run, nil
}

// Close closes the database connection
func (s *BenchmarkStorage) Close() error {
	return s.db.Close()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
