package memory

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"unified-thinking/internal/embeddings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmbedder implements embeddings.Embedder for testing
type MockEmbedder struct {
	mu          sync.RWMutex
	embedFn     func(ctx context.Context, text string) ([]float32, error)
	embeddings  map[string][]float32
	dimension   int
	model       string
	provider    string
	embedCount  int
	shouldError bool
	errorMsg    string
}

func NewMockEmbedder() *MockEmbedder {
	return &MockEmbedder{
		embeddings: make(map[string][]float32),
		dimension:  512,
		model:      "test-model",
		provider:   "test-provider",
	}
}

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	m.mu.Lock()
	m.embedCount++
	shouldError := m.shouldError
	errorMsg := m.errorMsg
	embedFn := m.embedFn
	m.mu.Unlock()

	if shouldError {
		return nil, errors.New(errorMsg)
	}
	if embedFn != nil {
		return embedFn(ctx, text)
	}

	// Check cache with read lock
	m.mu.RLock()
	if emb, ok := m.embeddings[text]; ok {
		m.mu.RUnlock()
		return emb, nil
	}
	m.mu.RUnlock()

	// Generate deterministic embedding based on text hash
	emb := make([]float32, m.dimension)
	for i := range emb {
		emb[i] = float32(i%10) / 10.0 * float32(len(text)%10+1)
	}

	// Store with write lock
	m.mu.Lock()
	m.embeddings[text] = emb
	m.mu.Unlock()

	return emb, nil
}

func (m *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		results[i] = emb
	}
	return results, nil
}

func (m *MockEmbedder) Dimension() int {
	return m.dimension
}

func (m *MockEmbedder) Model() string {
	return m.model
}

func (m *MockEmbedder) Provider() string {
	return m.provider
}

func TestGetProvider_WithEmbedder(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   embeddings.DefaultConfig(),
	}

	assert.Equal(t, "test-provider", ei.GetProvider())
}

func TestGetProvider_WithoutEmbedder(t *testing.T) {
	store := NewEpisodicMemoryStore()

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: nil,
	}

	assert.Equal(t, "none", ei.GetProvider())
}

func TestGetEmbedder(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
	}

	assert.Equal(t, mockEmbedder, ei.GetEmbedder())
}

func TestGetEmbedder_Nil(t *testing.T) {
	store := NewEpisodicMemoryStore()

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: nil,
	}

	assert.Nil(t, ei.GetEmbedder())
}

func TestGenerateAndStoreEmbedding_Disabled(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.Enabled = false

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: NewMockEmbedder(),
		config:   config,
	}

	ctx := context.Background()
	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "test",
	}

	err := ei.GenerateAndStoreEmbedding(ctx, problem)
	require.NoError(t, err)

	// Embedding should not be set
	assert.Nil(t, problem.Embedding)
}

func TestGenerateAndStoreEmbedding_NoEmbedder(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.Enabled = true

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: nil, // No embedder
		config:   config,
	}

	ctx := context.Background()
	problem := &ProblemDescription{
		Description: "Test problem",
	}

	err := ei.GenerateAndStoreEmbedding(ctx, problem)
	require.NoError(t, err)

	// Should skip without error
	assert.Nil(t, problem.Embedding)
}

func TestGenerateAndStoreEmbedding_Success(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	config := embeddings.DefaultConfig()
	config.Enabled = true

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
		cache:    embeddings.NewLRUEmbeddingCache(&embeddings.LRUCacheConfig{TTL: time.Hour}),
	}

	ctx := context.Background()
	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "test-domain",
		Goals:       []string{"goal1", "goal2"},
		Context:     "Some context",
		ProblemType: "testing",
	}

	err := ei.GenerateAndStoreEmbedding(ctx, problem)
	require.NoError(t, err)

	// Embedding should be set
	assert.NotNil(t, problem.Embedding)
	assert.Len(t, problem.Embedding, mockEmbedder.dimension)

	// Metadata should be set
	assert.NotNil(t, problem.EmbeddingMeta)
	assert.Equal(t, "test-model", problem.EmbeddingMeta.Model)
	assert.Equal(t, "test-provider", problem.EmbeddingMeta.Provider)
	assert.Equal(t, 512, problem.EmbeddingMeta.Dimension)
}

func TestGenerateAndStoreEmbedding_Error(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	mockEmbedder.shouldError = true
	mockEmbedder.errorMsg = "embedding API error"

	config := embeddings.DefaultConfig()
	config.Enabled = true

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
	}

	ctx := context.Background()
	problem := &ProblemDescription{
		Description: "Test problem",
	}

	err := ei.GenerateAndStoreEmbedding(ctx, problem)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "embedding API error")
}

func TestGenerateAndStoreEmbedding_EmptyEmbedding(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	mockEmbedder.embedFn = func(ctx context.Context, text string) ([]float32, error) {
		return []float32{}, nil // Return empty embedding
	}

	config := embeddings.DefaultConfig()
	config.Enabled = true

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
	}

	ctx := context.Background()
	problem := &ProblemDescription{
		Description: "Test problem",
	}

	err := ei.GenerateAndStoreEmbedding(ctx, problem)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty array")
}

func TestProblemToText(t *testing.T) {
	store := NewEpisodicMemoryStore()
	ei := &EmbeddingIntegration{store: store}

	tests := []struct {
		name     string
		problem  *ProblemDescription
		expected string
	}{
		{
			name: "description only",
			problem: &ProblemDescription{
				Description: "Test problem",
			},
			expected: "Test problem",
		},
		{
			name: "with context",
			problem: &ProblemDescription{
				Description: "Test problem",
				Context:     "Some context",
			},
			expected: "Test problem. Context: Some context",
		},
		{
			name: "with goals",
			problem: &ProblemDescription{
				Description: "Test problem",
				Goals:       []string{"goal1", "goal2"},
			},
			expected: "Test problem. Goals: goal1, goal2",
		},
		{
			name: "with domain",
			problem: &ProblemDescription{
				Description: "Test problem",
				Domain:      "test-domain",
			},
			expected: "Test problem. Domain: test-domain",
		},
		{
			name: "with problem type",
			problem: &ProblemDescription{
				Description: "Test problem",
				ProblemType: "debugging",
			},
			expected: "Test problem. Type: debugging",
		},
		{
			name: "full problem",
			problem: &ProblemDescription{
				Description: "Test problem",
				Context:     "Some context",
				Goals:       []string{"goal1", "goal2"},
				Domain:      "test-domain",
				ProblemType: "debugging",
			},
			expected: "Test problem. Context: Some context. Goals: goal1, goal2. Domain: test-domain. Type: debugging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ei.problemToText(tt.problem)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetrieveSimilarWithHybridSearch_Disabled(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.Enabled = false

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: nil,
		config:   config,
	}

	ctx := context.Background()

	// Store some trajectories
	for i := 0; i < 3; i++ {
		traj := &ReasoningTrajectory{
			SessionID: "session_" + string(rune('a'+i)),
			Domain:    "test",
			Problem: &ProblemDescription{
				Domain:      "test",
				ProblemType: "testing",
			},
			SuccessScore: 0.8,
		}
		store.StoreTrajectory(ctx, traj)
	}

	problem := &ProblemDescription{
		Domain:      "test",
		ProblemType: "testing",
	}

	// Should fall back to hash-based search
	matches, err := ei.RetrieveSimilarWithHybridSearch(ctx, problem, 10)
	require.NoError(t, err)
	assert.NotNil(t, matches)
}

func TestRetrieveSimilarWithHybridSearch_GeneratesEmbedding(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	config := embeddings.DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.0 // Accept all matches for testing

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
		cache:    embeddings.NewLRUEmbeddingCache(&embeddings.LRUCacheConfig{TTL: time.Hour}),
	}

	ctx := context.Background()
	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "test",
	}

	// No embedding initially
	assert.Nil(t, problem.Embedding)

	matches, err := ei.RetrieveSimilarWithHybridSearch(ctx, problem, 10)
	require.NoError(t, err)
	assert.NotNil(t, matches)

	// Embedding should now be generated
	assert.NotNil(t, problem.Embedding)
}

func TestVectorSearch_Empty(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	config := embeddings.DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.5

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
	}

	ctx := context.Background()
	problem := &ProblemDescription{
		Embedding: make([]float32, 512),
	}

	matches := ei.vectorSearch(ctx, problem, 10)
	assert.Empty(t, matches)
}

func TestVectorSearch_WithMatches(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	config := embeddings.DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.0 // Accept all for testing

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
	}

	ctx := context.Background()

	// Create query embedding
	queryEmbedding := make([]float32, 512)
	for i := range queryEmbedding {
		queryEmbedding[i] = 0.5
	}

	// Store trajectories with embeddings
	for i := 0; i < 5; i++ {
		traj := &ReasoningTrajectory{
			ID:        "traj_" + string(rune('a'+i)),
			SessionID: "session_" + string(rune('a'+i)),
			Domain:    "test",
			Problem: &ProblemDescription{
				Domain:    "test",
				Embedding: make([]float32, 512),
			},
			SuccessScore: 0.8,
		}
		// Set similar embedding
		for j := range traj.Problem.Embedding {
			traj.Problem.Embedding[j] = 0.5 + float32(i)*0.01
		}
		store.trajectories[traj.ID] = traj
	}

	problem := &ProblemDescription{
		Embedding: queryEmbedding,
	}

	matches := ei.vectorSearch(ctx, problem, 10)
	assert.NotEmpty(t, matches)

	// Should be sorted by similarity
	for i := 0; i < len(matches)-1; i++ {
		assert.GreaterOrEqual(t, matches[i].SimilarityScore, matches[i+1].SimilarityScore)
	}
}

func TestVectorSearch_WithLimit(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	config := embeddings.DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.0

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
	}

	ctx := context.Background()

	// Create query embedding
	queryEmbedding := make([]float32, 512)
	for i := range queryEmbedding {
		queryEmbedding[i] = 0.5
	}

	// Store many trajectories
	for i := 0; i < 20; i++ {
		traj := &ReasoningTrajectory{
			ID:        "traj_" + string(rune('a'+i)),
			SessionID: "session_" + string(rune('a'+i)),
			Domain:    "test",
			Problem: &ProblemDescription{
				Domain:    "test",
				Embedding: queryEmbedding, // Same embedding = high similarity
			},
		}
		store.trajectories[traj.ID] = traj
	}

	problem := &ProblemDescription{
		Embedding: queryEmbedding,
	}

	// Limit to 5
	matches := ei.vectorSearch(ctx, problem, 5)
	assert.Len(t, matches, 5)
}

func TestVectorSearch_NilProblem(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.0

	ei := &EmbeddingIntegration{
		store:  store,
		config: config,
	}

	ctx := context.Background()

	// Store trajectory with nil problem
	store.trajectories["traj_1"] = &ReasoningTrajectory{
		ID:      "traj_1",
		Problem: nil,
	}

	// Store trajectory with nil embedding
	store.trajectories["traj_2"] = &ReasoningTrajectory{
		ID: "traj_2",
		Problem: &ProblemDescription{
			Embedding: nil,
		},
	}

	problem := &ProblemDescription{
		Embedding: make([]float32, 512),
	}

	// Should not panic and should skip trajectories without embeddings
	matches := ei.vectorSearch(ctx, problem, 10)
	assert.Empty(t, matches)
}

func TestReciprocalRankFusion_Empty(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.RRFParameter = 60

	ei := &EmbeddingIntegration{
		store:  store,
		config: config,
	}

	results, err := ei.reciprocalRankFusion([]*TrajectoryMatch{}, []*TrajectoryMatch{}, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestReciprocalRankFusion_StructuredOnly(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.RRFParameter = 60

	ei := &EmbeddingIntegration{
		store:  store,
		config: config,
	}

	structured := []*TrajectoryMatch{
		{
			Trajectory:      &ReasoningTrajectory{ID: "traj_1"},
			SimilarityScore: 0.9,
		},
		{
			Trajectory:      &ReasoningTrajectory{ID: "traj_2"},
			SimilarityScore: 0.8,
		},
	}

	results, err := ei.reciprocalRankFusion(structured, []*TrajectoryMatch{}, 10)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestReciprocalRankFusion_VectorOnly(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.RRFParameter = 60

	ei := &EmbeddingIntegration{
		store:  store,
		config: config,
	}

	vector := []*TrajectoryMatch{
		{
			Trajectory:       &ReasoningTrajectory{ID: "traj_1"},
			SimilarityScore:  0.9,
			RelevanceFactors: []string{"Semantic: 0.9"},
		},
		{
			Trajectory:       &ReasoningTrajectory{ID: "traj_2"},
			SimilarityScore:  0.8,
			RelevanceFactors: []string{"Semantic: 0.8"},
		},
	}

	results, err := ei.reciprocalRankFusion([]*TrajectoryMatch{}, vector, 10)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestReciprocalRankFusion_Combined(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.RRFParameter = 60

	ei := &EmbeddingIntegration{
		store:  store,
		config: config,
	}

	structured := []*TrajectoryMatch{
		{
			Trajectory:      &ReasoningTrajectory{ID: "traj_1"},
			SimilarityScore: 0.9,
		},
		{
			Trajectory:      &ReasoningTrajectory{ID: "traj_2"},
			SimilarityScore: 0.8,
		},
		{
			Trajectory:      &ReasoningTrajectory{ID: "traj_3"},
			SimilarityScore: 0.7,
		},
	}

	vector := []*TrajectoryMatch{
		{
			Trajectory:       &ReasoningTrajectory{ID: "traj_2"}, // Overlaps
			SimilarityScore:  0.95,
			RelevanceFactors: []string{"Semantic: 0.95"},
		},
		{
			Trajectory:       &ReasoningTrajectory{ID: "traj_4"}, // New
			SimilarityScore:  0.85,
			RelevanceFactors: []string{"Semantic: 0.85"},
		},
	}

	results, err := ei.reciprocalRankFusion(structured, vector, 10)
	require.NoError(t, err)

	// Should have 4 unique trajectories
	assert.Len(t, results, 4)

	// traj_2 should be first (appears in both lists)
	assert.Equal(t, "traj_2", results[0].Trajectory.ID)
}

func TestReciprocalRankFusion_Limit(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.RRFParameter = 60

	ei := &EmbeddingIntegration{
		store:  store,
		config: config,
	}

	structured := []*TrajectoryMatch{
		{Trajectory: &ReasoningTrajectory{ID: "traj_1"}},
		{Trajectory: &ReasoningTrajectory{ID: "traj_2"}},
		{Trajectory: &ReasoningTrajectory{ID: "traj_3"}},
	}

	vector := []*TrajectoryMatch{
		{Trajectory: &ReasoningTrajectory{ID: "traj_4"}},
		{Trajectory: &ReasoningTrajectory{ID: "traj_5"}},
	}

	results, err := ei.reciprocalRankFusion(structured, vector, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestReciprocalRankFusion_NilTrajectory(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.RRFParameter = 60

	ei := &EmbeddingIntegration{
		store:  store,
		config: config,
	}

	structured := []*TrajectoryMatch{
		{Trajectory: nil}, // Should be skipped
		{Trajectory: &ReasoningTrajectory{ID: "traj_1"}},
	}

	results, err := ei.reciprocalRankFusion(structured, []*TrajectoryMatch{}, 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestReciprocalRankFusion_ZeroK(t *testing.T) {
	store := NewEpisodicMemoryStore()
	config := embeddings.DefaultConfig()
	config.RRFParameter = 0 // Should default to 60

	ei := &EmbeddingIntegration{
		store:  store,
		config: config,
	}

	structured := []*TrajectoryMatch{
		{Trajectory: &ReasoningTrajectory{ID: "traj_1"}},
	}

	results, err := ei.reciprocalRankFusion(structured, []*TrajectoryMatch{}, 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestLoadEmbeddingsFromStorage_NilSQLite(t *testing.T) {
	store := NewEpisodicMemoryStore()

	ei := &EmbeddingIntegration{
		store:       store,
		sqliteStore: nil,
	}

	err := ei.LoadEmbeddingsFromStorage()
	require.NoError(t, err)
}

func TestEmbeddingIntegration_ConcurrentAccess(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	config := embeddings.DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.0

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
		cache:    embeddings.NewLRUEmbeddingCache(&embeddings.LRUCacheConfig{TTL: time.Hour}),
	}

	ctx := context.Background()

	// Store some trajectories
	for i := 0; i < 10; i++ {
		traj := &ReasoningTrajectory{
			ID:        "traj_" + string(rune('a'+i)),
			SessionID: "session_" + string(rune('a'+i)),
			Domain:    "test",
			Problem: &ProblemDescription{
				Domain:    "test",
				Embedding: make([]float32, 512),
			},
		}
		store.trajectories[traj.ID] = traj
	}

	done := make(chan bool, 20)

	// Concurrent embedding generation
	for i := 0; i < 5; i++ {
		go func(idx int) {
			problem := &ProblemDescription{
				Description: "problem " + string(rune('a'+idx)),
			}
			ei.GenerateAndStoreEmbedding(ctx, problem)
			done <- true
		}(i)
	}

	// Concurrent searches
	for i := 0; i < 5; i++ {
		go func() {
			problem := &ProblemDescription{
				Domain:    "test",
				Embedding: make([]float32, 512),
			}
			ei.vectorSearch(ctx, problem, 10)
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRetrieveSimilarWithHybridSearch_EmbeddingError(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	mockEmbedder.shouldError = true
	mockEmbedder.errorMsg = "API error"

	config := embeddings.DefaultConfig()
	config.Enabled = true

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
	}

	ctx := context.Background()
	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "test",
	}

	_, err := ei.RetrieveSimilarWithHybridSearch(ctx, problem, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestRetrieveSimilarWithHybridSearch_NoVectorResults(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockEmbedder := NewMockEmbedder()
	config := embeddings.DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 1.0 // Very high threshold - no vector matches

	ei := &EmbeddingIntegration{
		store:    store,
		embedder: mockEmbedder,
		config:   config,
		cache:    embeddings.NewLRUEmbeddingCache(&embeddings.LRUCacheConfig{TTL: time.Hour}),
	}

	ctx := context.Background()

	// Store trajectories for structured search
	for i := 0; i < 3; i++ {
		traj := &ReasoningTrajectory{
			SessionID: "session_" + string(rune('a'+i)),
			Domain:    "test",
			Problem: &ProblemDescription{
				Domain:      "test",
				ProblemType: "testing",
			},
			SuccessScore: 0.8,
		}
		store.StoreTrajectory(ctx, traj)
	}

	problem := &ProblemDescription{
		Domain:      "test",
		ProblemType: "testing",
	}

	matches, err := ei.RetrieveSimilarWithHybridSearch(ctx, problem, 10)
	require.NoError(t, err)
	// Should return structured results since no vector results
	assert.NotNil(t, matches)
}
