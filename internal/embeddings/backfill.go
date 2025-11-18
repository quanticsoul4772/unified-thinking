package embeddings

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// BackfillStorage defines the interface for storage operations needed by backfill
type BackfillStorage interface {
	ListSignaturesWithoutEmbeddings(limit int) ([]*SignatureForBackfill, error)
	UpdateSignatureEmbedding(trajectoryID string, embedding []float32) error
}

// SignatureForBackfill represents a signature that needs embedding backfill
// This mirrors the storage type to avoid import cycles
type SignatureForBackfill struct {
	TrajectoryID string
	Description  string
}

// BackfillStats tracks backfill operation statistics
type BackfillStats struct {
	Total     int64
	Processed int64
	Succeeded int64
	Failed    int64
	Skipped   int64
	Duration  time.Duration
}

// BackfillConfig configures the backfill operation
type BackfillConfig struct {
	BatchSize      int           // Number of signatures to process per batch
	MaxConcurrency int           // Maximum concurrent embedding operations
	Timeout        time.Duration // Timeout per embedding operation
	DryRun         bool          // If true, don't actually update the database
}

// DefaultBackfillConfig returns default backfill configuration
func DefaultBackfillConfig() *BackfillConfig {
	return &BackfillConfig{
		BatchSize:      100,
		MaxConcurrency: 5,
		Timeout:        30 * time.Second,
		DryRun:         false,
	}
}

// BackfillRunner runs the embedding backfill operation
type BackfillRunner struct {
	storage  BackfillStorage
	embedder Embedder
	config   *BackfillConfig
}

// NewBackfillRunner creates a new backfill runner
func NewBackfillRunner(storage BackfillStorage, embedder Embedder, config *BackfillConfig) *BackfillRunner {
	if config == nil {
		config = DefaultBackfillConfig()
	}
	return &BackfillRunner{
		storage:  storage,
		embedder: embedder,
		config:   config,
	}
}

// Run executes the backfill operation
func (r *BackfillRunner) Run(ctx context.Context) (*BackfillStats, error) {
	start := time.Now()
	stats := &BackfillStats{}

	if r.storage == nil {
		return stats, fmt.Errorf("storage is nil")
	}
	if r.embedder == nil {
		return stats, fmt.Errorf("embedder is nil")
	}

	// This is a simplified approach - in production you'd want pagination
	// For now, we fetch signatures and process them with the given limit
	signatures, err := r.fetchSignatures()
	if err != nil {
		return stats, fmt.Errorf("failed to fetch signatures: %w", err)
	}

	atomic.StoreInt64(&stats.Total, int64(len(signatures)))

	if len(signatures) == 0 {
		log.Printf("No signatures found needing embedding backfill")
		stats.Duration = time.Since(start)
		return stats, nil
	}

	log.Printf("Starting backfill for %d signatures (concurrency=%d, dry_run=%v)",
		len(signatures), r.config.MaxConcurrency, r.config.DryRun)

	// Process with concurrency control
	semaphore := make(chan struct{}, r.config.MaxConcurrency)
	var wg sync.WaitGroup

	for _, sig := range signatures {
		select {
		case <-ctx.Done():
			wg.Wait()
			stats.Duration = time.Since(start)
			return stats, ctx.Err()
		default:
		}

		wg.Add(1)
		go func(s *SignatureForBackfill) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			r.processSignature(ctx, s, stats)
		}(sig)
	}

	wg.Wait()
	stats.Duration = time.Since(start)

	log.Printf("Backfill complete: processed=%d, succeeded=%d, failed=%d, skipped=%d, duration=%v",
		stats.Processed, stats.Succeeded, stats.Failed, stats.Skipped, stats.Duration)

	return stats, nil
}

// fetchSignatures retrieves signatures that need backfill
func (r *BackfillRunner) fetchSignatures() ([]*SignatureForBackfill, error) {
	// Call the storage method - we need to adapt the storage type
	sigs, err := r.storage.ListSignaturesWithoutEmbeddings(r.config.BatchSize)
	if err != nil {
		return nil, err
	}

	// Convert storage type to our type
	result := make([]*SignatureForBackfill, len(sigs))
	for i, sig := range sigs {
		result[i] = &SignatureForBackfill{
			TrajectoryID: sig.TrajectoryID,
			Description:  sig.Description,
		}
	}
	return result, nil
}

// processSignature generates and stores embedding for a single signature
func (r *BackfillRunner) processSignature(ctx context.Context, sig *SignatureForBackfill, stats *BackfillStats) {
	atomic.AddInt64(&stats.Processed, 1)

	// Skip if no description
	if sig.Description == "" {
		atomic.AddInt64(&stats.Skipped, 1)
		log.Printf("[SKIP] Trajectory %s: no description for embedding", sig.TrajectoryID)
		return
	}

	// Create context with timeout
	embedCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	// Generate embedding
	embedding, err := r.embedder.Embed(embedCtx, sig.Description)
	if err != nil {
		atomic.AddInt64(&stats.Failed, 1)
		log.Printf("[FAIL] Trajectory %s: embedding generation failed: %v", sig.TrajectoryID, err)
		return
	}

	if len(embedding) == 0 {
		atomic.AddInt64(&stats.Failed, 1)
		log.Printf("[FAIL] Trajectory %s: empty embedding returned", sig.TrajectoryID)
		return
	}

	// Update in database (unless dry run)
	if r.config.DryRun {
		atomic.AddInt64(&stats.Succeeded, 1)
		log.Printf("[DRY-RUN] Trajectory %s: would update with %d-dim embedding", sig.TrajectoryID, len(embedding))
		return
	}

	if err := r.storage.UpdateSignatureEmbedding(sig.TrajectoryID, embedding); err != nil {
		atomic.AddInt64(&stats.Failed, 1)
		log.Printf("[FAIL] Trajectory %s: database update failed: %v", sig.TrajectoryID, err)
		return
	}

	atomic.AddInt64(&stats.Succeeded, 1)
	log.Printf("[OK] Trajectory %s: updated with %d-dim embedding", sig.TrajectoryID, len(embedding))
}

// BackfillStorageAdapter adapts storage.SQLiteStorage to BackfillStorage interface
type BackfillStorageAdapter struct {
	listFunc   func(limit int) (interface{}, error)
	updateFunc func(trajectoryID string, embedding []float32) error
}

// NewBackfillStorageAdapter creates an adapter from function callbacks
// This avoids import cycles between embeddings and storage packages
func NewBackfillStorageAdapter(
	listFunc func(limit int) (interface{}, error),
	updateFunc func(trajectoryID string, embedding []float32) error,
) BackfillStorage {
	return &backfillStorageWrapper{
		listFunc:   listFunc,
		updateFunc: updateFunc,
	}
}

type backfillStorageWrapper struct {
	listFunc   func(limit int) (interface{}, error)
	updateFunc func(trajectoryID string, embedding []float32) error
}

func (w *backfillStorageWrapper) ListSignaturesWithoutEmbeddings(limit int) ([]*SignatureForBackfill, error) {
	result, err := w.listFunc(limit)
	if err != nil {
		return nil, err
	}
	// Type assertion - the caller must ensure the correct type
	return result.([]*SignatureForBackfill), nil
}

func (w *backfillStorageWrapper) UpdateSignatureEmbedding(trajectoryID string, embedding []float32) error {
	return w.updateFunc(trajectoryID, embedding)
}
