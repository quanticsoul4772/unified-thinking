package benchmarks

import (
	"os"
	"testing"
	"time"
)

func TestBenchmarkStorage(t *testing.T) {
	// Create temporary database
	dbPath := "test_benchmarks.db"
	defer os.Remove(dbPath)

	storage, err := NewBenchmarkStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Create test benchmark run
	run := &BenchmarkRun{
		RunID:           "test_run_001",
		SuiteName:       "Test Suite",
		GitCommit:       "abc123",
		Timestamp:       time.Now(),
		OverallAccuracy: 0.75,
		OverallECE:      0.15,
		AvgLatency:      100 * time.Millisecond,
		TotalProblems:   10,
		CorrectProblems: 7,
		Results: []*Result{
			{
				ProblemID:  "p1",
				Correct:    true,
				Score:      1.0,
				Confidence: 0.9,
				Latency:    50 * time.Millisecond,
				Mode:       "linear",
				Tokens:     25,
				Response:   "correct answer",
			},
			{
				ProblemID:  "p2",
				Correct:    false,
				Score:      0.0,
				Confidence: 0.6,
				Latency:    150 * time.Millisecond,
				Mode:       "tree",
				Tokens:     30,
				Response:   "wrong answer",
			},
		},
	}

	// Store run
	if err := storage.StoreBenchmarkRun(run); err != nil {
		t.Fatalf("Failed to store run: %v", err)
	}

	// Retrieve latest run
	retrieved, err := storage.GetLatestRun("Test Suite")
	if err != nil {
		t.Fatalf("Failed to get latest run: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected to retrieve run, got nil")
	}

	// Verify key fields
	if retrieved.RunID != run.RunID {
		t.Errorf("RunID = %v, want %v", retrieved.RunID, run.RunID)
	}

	if retrieved.SuiteName != run.SuiteName {
		t.Errorf("SuiteName = %v, want %v", retrieved.SuiteName, run.SuiteName)
	}

	if retrieved.OverallAccuracy != run.OverallAccuracy {
		t.Errorf("OverallAccuracy = %v, want %v", retrieved.OverallAccuracy, run.OverallAccuracy)
	}

	if len(retrieved.Results) != len(run.Results) {
		t.Errorf("Results count = %d, want %d", len(retrieved.Results), len(run.Results))
	}

	t.Logf("Successfully stored and retrieved benchmark run %s", run.RunID)
}

func TestBenchmarkStorage_NoResults(t *testing.T) {
	dbPath := "test_empty.db"
	defer os.Remove(dbPath)

	storage, err := NewBenchmarkStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Query for non-existent suite
	run, err := storage.GetLatestRun("NonExistent")
	if err != nil {
		t.Fatalf("GetLatestRun should not error on no results: %v", err)
	}

	if run != nil {
		t.Error("Expected nil run for non-existent suite")
	}
}
