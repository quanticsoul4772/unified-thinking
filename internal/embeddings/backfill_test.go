package embeddings

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockBackfillStorage implements BackfillStorage for testing
type mockBackfillStorage struct {
	signatures []*SignatureForBackfill
	listErr    error
	updateErr  error
	updated    map[string][]float32
	mu         sync.Mutex
}

func newMockBackfillStorage(signatures []*SignatureForBackfill) *mockBackfillStorage {
	return &mockBackfillStorage{
		signatures: signatures,
		updated:    make(map[string][]float32),
	}
}

func (m *mockBackfillStorage) ListSignaturesWithoutEmbeddings(limit int) ([]*SignatureForBackfill, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if limit > len(m.signatures) {
		return m.signatures, nil
	}
	return m.signatures[:limit], nil
}

func (m *mockBackfillStorage) UpdateSignatureEmbedding(trajectoryID string, embedding []float32) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	m.updated[trajectoryID] = embedding
	m.mu.Unlock()
	return nil
}

// mockBackfillEmbedder implements Embedder for testing
type mockBackfillEmbedder struct {
	embeddings map[string][]float32
	embedErr   error
	callCount  int64
}

func newMockBackfillEmbedder() *mockBackfillEmbedder {
	return &mockBackfillEmbedder{
		embeddings: make(map[string][]float32),
	}
}

func (m *mockBackfillEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	atomic.AddInt64(&m.callCount, 1)
	if m.embedErr != nil {
		return nil, m.embedErr
	}
	if emb, ok := m.embeddings[text]; ok {
		return emb, nil
	}
	// Return a default embedding
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *mockBackfillEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		result[i] = emb
	}
	return result, nil
}

func (m *mockBackfillEmbedder) Dimension() int   { return 3 }
func (m *mockBackfillEmbedder) Model() string    { return "test-model" }
func (m *mockBackfillEmbedder) Provider() string { return "test" }

func TestDefaultBackfillConfig(t *testing.T) {
	cfg := DefaultBackfillConfig()

	if cfg.BatchSize != 100 {
		t.Errorf("expected BatchSize 100, got %d", cfg.BatchSize)
	}
	if cfg.MaxConcurrency != 5 {
		t.Errorf("expected MaxConcurrency 5, got %d", cfg.MaxConcurrency)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.DryRun {
		t.Error("expected DryRun to be false")
	}
}

func TestBackfillRunner_EmptySignatures(t *testing.T) {
	storage := newMockBackfillStorage([]*SignatureForBackfill{})
	embedder := newMockBackfillEmbedder()
	runner := NewBackfillRunner(storage, embedder, nil)

	stats, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Total != 0 {
		t.Errorf("expected Total 0, got %d", stats.Total)
	}
	if stats.Processed != 0 {
		t.Errorf("expected Processed 0, got %d", stats.Processed)
	}
}

func TestBackfillRunner_Success(t *testing.T) {
	signatures := []*SignatureForBackfill{
		{TrajectoryID: "traj-1", Description: "test description 1"},
		{TrajectoryID: "traj-2", Description: "test description 2"},
	}
	storage := newMockBackfillStorage(signatures)
	embedder := newMockBackfillEmbedder()

	config := DefaultBackfillConfig()
	config.MaxConcurrency = 1 // Sequential for predictable testing

	runner := NewBackfillRunner(storage, embedder, config)

	stats, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Total != 2 {
		t.Errorf("expected Total 2, got %d", stats.Total)
	}
	if stats.Processed != 2 {
		t.Errorf("expected Processed 2, got %d", stats.Processed)
	}
	if stats.Succeeded != 2 {
		t.Errorf("expected Succeeded 2, got %d", stats.Succeeded)
	}
	if stats.Failed != 0 {
		t.Errorf("expected Failed 0, got %d", stats.Failed)
	}

	// Check storage was updated
	if len(storage.updated) != 2 {
		t.Errorf("expected 2 updates, got %d", len(storage.updated))
	}
}

func TestBackfillRunner_SkipsEmptyDescription(t *testing.T) {
	signatures := []*SignatureForBackfill{
		{TrajectoryID: "traj-1", Description: ""},
		{TrajectoryID: "traj-2", Description: "valid description"},
	}
	storage := newMockBackfillStorage(signatures)
	embedder := newMockBackfillEmbedder()

	config := DefaultBackfillConfig()
	config.MaxConcurrency = 1

	runner := NewBackfillRunner(storage, embedder, config)

	stats, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Skipped != 1 {
		t.Errorf("expected Skipped 1, got %d", stats.Skipped)
	}
	if stats.Succeeded != 1 {
		t.Errorf("expected Succeeded 1, got %d", stats.Succeeded)
	}
}

func TestBackfillRunner_DryRun(t *testing.T) {
	signatures := []*SignatureForBackfill{
		{TrajectoryID: "traj-1", Description: "test description"},
	}
	storage := newMockBackfillStorage(signatures)
	embedder := newMockBackfillEmbedder()

	config := DefaultBackfillConfig()
	config.DryRun = true

	runner := NewBackfillRunner(storage, embedder, config)

	stats, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Succeeded != 1 {
		t.Errorf("expected Succeeded 1, got %d", stats.Succeeded)
	}

	// Check storage was NOT updated
	if len(storage.updated) != 0 {
		t.Errorf("expected 0 updates in dry run, got %d", len(storage.updated))
	}
}

func TestBackfillRunner_EmbedError(t *testing.T) {
	signatures := []*SignatureForBackfill{
		{TrajectoryID: "traj-1", Description: "test description"},
	}
	storage := newMockBackfillStorage(signatures)
	embedder := newMockBackfillEmbedder()
	embedder.embedErr = errors.New("embedding failed")

	config := DefaultBackfillConfig()

	runner := NewBackfillRunner(storage, embedder, config)

	stats, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Failed != 1 {
		t.Errorf("expected Failed 1, got %d", stats.Failed)
	}
	if stats.Succeeded != 0 {
		t.Errorf("expected Succeeded 0, got %d", stats.Succeeded)
	}
}

func TestBackfillRunner_UpdateError(t *testing.T) {
	signatures := []*SignatureForBackfill{
		{TrajectoryID: "traj-1", Description: "test description"},
	}
	storage := newMockBackfillStorage(signatures)
	storage.updateErr = errors.New("update failed")
	embedder := newMockBackfillEmbedder()

	config := DefaultBackfillConfig()

	runner := NewBackfillRunner(storage, embedder, config)

	stats, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Failed != 1 {
		t.Errorf("expected Failed 1, got %d", stats.Failed)
	}
}

func TestBackfillRunner_ListError(t *testing.T) {
	storage := newMockBackfillStorage(nil)
	storage.listErr = errors.New("list failed")
	embedder := newMockBackfillEmbedder()

	runner := NewBackfillRunner(storage, embedder, nil)

	_, err := runner.Run(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, storage.listErr) && err.Error() != "failed to fetch signatures: list failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBackfillRunner_NilStorage(t *testing.T) {
	embedder := newMockBackfillEmbedder()
	runner := NewBackfillRunner(nil, embedder, nil)

	_, err := runner.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for nil storage")
	}
}

func TestBackfillRunner_NilEmbedder(t *testing.T) {
	storage := newMockBackfillStorage([]*SignatureForBackfill{})
	runner := NewBackfillRunner(storage, nil, nil)

	_, err := runner.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for nil embedder")
	}
}

func TestBackfillRunner_ContextCancellation(t *testing.T) {
	signatures := make([]*SignatureForBackfill, 10)
	for i := range signatures {
		signatures[i] = &SignatureForBackfill{
			TrajectoryID: "traj-" + string(rune('a'+i)),
			Description:  "description",
		}
	}
	storage := newMockBackfillStorage(signatures)
	embedder := newMockBackfillEmbedder()

	config := DefaultBackfillConfig()
	config.MaxConcurrency = 1 // Sequential to make cancellation predictable

	runner := NewBackfillRunner(storage, embedder, config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := runner.Run(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestBackfillRunner_Concurrency(t *testing.T) {
	signatures := make([]*SignatureForBackfill, 20)
	for i := range signatures {
		signatures[i] = &SignatureForBackfill{
			TrajectoryID: string(rune('a' + i)),
			Description:  "description",
		}
	}
	storage := newMockBackfillStorage(signatures)
	embedder := newMockBackfillEmbedder()

	config := DefaultBackfillConfig()
	config.MaxConcurrency = 5

	runner := NewBackfillRunner(storage, embedder, config)

	stats, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Succeeded != 20 {
		t.Errorf("expected Succeeded 20, got %d", stats.Succeeded)
	}
}

func TestNewBackfillStorageAdapter(t *testing.T) {
	called := false
	adapter := NewBackfillStorageAdapter(
		func(limit int) (interface{}, error) {
			called = true
			return []*SignatureForBackfill{}, nil
		},
		func(trajectoryID string, embedding []float32) error {
			return nil
		},
	)

	_, err := adapter.ListSignaturesWithoutEmbeddings(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected list function to be called")
	}
}
