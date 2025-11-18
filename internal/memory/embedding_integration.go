package memory

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/storage"
)

// EmbeddingIntegration handles embedding generation and storage for episodic memory
type EmbeddingIntegration struct {
	store       *EpisodicMemoryStore
	sqliteStore *storage.SQLiteStorage
	embedder    embeddings.Embedder
	config      *embeddings.Config
	cache       *embeddings.EmbeddingCache
}

// GetProvider returns the embedding provider name
func (ei *EmbeddingIntegration) GetProvider() string {
	if ei.embedder != nil {
		return ei.embedder.Provider()
	}
	return "none"
}

// GetEmbedder returns the underlying embedder for use by other components
func (ei *EmbeddingIntegration) GetEmbedder() embeddings.Embedder {
	return ei.embedder
}

// NewEmbeddingIntegration creates a new embedding integration
func NewEmbeddingIntegration(store *EpisodicMemoryStore, sqliteStore *storage.SQLiteStorage) (*EmbeddingIntegration, error) {
	// Get configuration from environment
	config := embeddings.ConfigFromEnv()

	if !config.Enabled {
		log.Println("Embeddings disabled, using hash-based search only")
		return nil, nil // Return nil when disabled, don't create a broken integration
	}

	// Create embedder
	var embedder embeddings.Embedder

	switch config.Provider {
	case "voyage":
		if config.APIKey == "" {
			// Try to get from environment if not set
			config.APIKey = os.Getenv("VOYAGE_API_KEY")
		}
		if config.APIKey == "" {
			return nil, fmt.Errorf("VOYAGE_API_KEY not set")
		}
		embedder = embeddings.NewVoyageEmbedder(config.APIKey, config.Model)
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", config.Provider)
	}

	// Create cache
	cache := embeddings.NewEmbeddingCache(config.CacheTTL)

	return &EmbeddingIntegration{
		store:       store,
		sqliteStore: sqliteStore,
		embedder:    embedder,
		config:      config,
		cache:       cache,
	}, nil
}

// GenerateAndStoreEmbedding generates and stores an embedding for a problem
func (ei *EmbeddingIntegration) GenerateAndStoreEmbedding(ctx context.Context, problem *ProblemDescription) error {
	if !ei.config.Enabled || ei.embedder == nil {
		return nil // Embeddings disabled, skip
	}

	// Generate embedding text
	text := ei.problemToText(problem)

	// Use a background context with timeout to ensure embedding completes
	// even if the original request context is cancelled
	embedCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate embedding
	log.Printf("Generating embedding for problem (text length: %d)", len(text))
	embedding, err := ei.embedder.Embed(embedCtx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}
	if len(embedding) == 0 {
		return fmt.Errorf("embedding returned empty array")
	}
	log.Printf("Successfully generated embedding (dimension: %d)", len(embedding))

	// Store embedding in problem (for in-memory usage)
	problem.Embedding = embedding
	problem.EmbeddingMeta = &EmbeddingMetadata{
		Model:     ei.embedder.Model(),
		Provider:  ei.embedder.Provider(),
		Dimension: ei.embedder.Dimension(),
		Source:    "description+context+goals",
	}

	// Store in SQLite if available
	if ei.sqliteStore != nil {
		problemID := ComputeProblemHash(problem)
		err = ei.sqliteStore.StoreEmbedding(
			problemID,
			embedding,
			ei.embedder.Model(),
			ei.embedder.Provider(),
			ei.embedder.Dimension(),
			"description+context+goals",
		)
		if err != nil {
			return fmt.Errorf("failed to store embedding in SQLite: %w", err)
		}
	}

	return nil
}

// RetrieveSimilarWithHybridSearch performs hybrid search combining structured and semantic search
func (ei *EmbeddingIntegration) RetrieveSimilarWithHybridSearch(ctx context.Context, problem *ProblemDescription, limit int) ([]*TrajectoryMatch, error) {
	// If embeddings disabled, fall back to hash-based search
	if !ei.config.Enabled || ei.embedder == nil {
		return ei.store.RetrieveSimilarHashBased(problem, limit)
	}

	// Ensure problem has embedding
	if problem.Embedding == nil || len(problem.Embedding) == 0 {
		if err := ei.GenerateAndStoreEmbedding(ctx, problem); err != nil {
			return nil, fmt.Errorf("failed to generate embedding for query: %w", err)
		}
	}

	// 1. Get hash-based search results
	structuredResults, err := ei.store.RetrieveSimilarHashBased(problem, limit*2)
	if err != nil {
		return nil, fmt.Errorf("structured search failed: %w", err)
	}

	// 2. Get vector search results (if embeddings available)
	var vectorResults []*TrajectoryMatch
	if problem.Embedding != nil && len(problem.Embedding) > 0 {
		vectorResults = ei.vectorSearch(ctx, problem, limit*2)
	}

	// 3. If no vector results, return structured results
	if len(vectorResults) == 0 {
		if len(structuredResults) > limit {
			return structuredResults[:limit], nil
		}
		return structuredResults, nil
	}

	// 4. Combine using RRF (Reciprocal Rank Fusion)
	return ei.reciprocalRankFusion(structuredResults, vectorResults, limit)
}

// vectorSearch performs semantic similarity search using embeddings
func (ei *EmbeddingIntegration) vectorSearch(ctx context.Context, problem *ProblemDescription, limit int) []*TrajectoryMatch {
	matches := make([]*TrajectoryMatch, 0)

	ei.store.mu.RLock()
	defer ei.store.mu.RUnlock()

	trajCount := 0
	embeddingCount := 0
	for _, traj := range ei.store.trajectories {
		trajCount++
		if traj.Problem == nil || traj.Problem.Embedding == nil {
			continue
		}
		embeddingCount++

		// Calculate similarity
		similarity := embeddings.CosineSimilarity(problem.Embedding, traj.Problem.Embedding)

		// Only include if above threshold
		if similarity >= ei.config.MinSimilarity {
			matches = append(matches, &TrajectoryMatch{
				Trajectory:      traj,
				SimilarityScore: similarity,
				RelevanceFactors: []string{
					fmt.Sprintf("Semantic similarity: %.2f", similarity),
				},
			})
		}
	}

	// Sort by similarity score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].SimilarityScore > matches[j].SimilarityScore
	})

	log.Printf("vectorSearch: %d trajectories, %d with embeddings, %d matches above threshold", trajCount, embeddingCount, len(matches))

	if len(matches) > limit {
		return matches[:limit]
	}
	return matches
}

// reciprocalRankFusion combines results using RRF algorithm
func (ei *EmbeddingIntegration) reciprocalRankFusion(structured, vector []*TrajectoryMatch, limit int) ([]*TrajectoryMatch, error) {
	k := float64(ei.config.RRFParameter) // Default: 60
	if k == 0 {
		k = 60
	}

	scores := make(map[string]float64)
	trajectories := make(map[string]*TrajectoryMatch)

	// Score structured results
	for rank, match := range structured {
		if match == nil || match.Trajectory == nil {
			continue
		}
		id := match.Trajectory.ID
		scores[id] += 1.0 / (k + float64(rank+1))
		trajectories[id] = match
	}

	// Score vector results
	for rank, match := range vector {
		if match == nil || match.Trajectory == nil {
			continue
		}
		id := match.Trajectory.ID
		scores[id] += 1.0 / (k + float64(rank+1))

		// If this trajectory wasn't in structured results, add it
		if existing, exists := trajectories[id]; !exists {
			trajectories[id] = match
		} else {
			// Merge relevance factors
			existing.SimilarityScore = match.SimilarityScore
			existing.RelevanceFactors = append(existing.RelevanceFactors, match.RelevanceFactors...)
		}
	}

	// Sort by combined RRF score
	type scoredMatch struct {
		match *TrajectoryMatch
		score float64
	}

	scored := make([]scoredMatch, 0, len(scores))
	for id, score := range scores {
		if match, exists := trajectories[id]; exists {
			match.SimilarityScore = score // Use RRF score as final similarity
			scored = append(scored, scoredMatch{
				match: match,
				score: score,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top limit results
	results := make([]*TrajectoryMatch, 0, limit)
	for i := 0; i < limit && i < len(scored); i++ {
		results = append(results, scored[i].match)
	}

	return results, nil
}

// LoadEmbeddingsFromStorage loads all embeddings from SQLite storage
func (ei *EmbeddingIntegration) LoadEmbeddingsFromStorage() error {
	if ei.sqliteStore == nil {
		return nil // No SQLite storage
	}

	embeddings, err := ei.sqliteStore.GetAllEmbeddings()
	if err != nil {
		return fmt.Errorf("failed to load embeddings: %w", err)
	}

	// Apply embeddings to trajectories
	ei.store.mu.Lock()
	defer ei.store.mu.Unlock()

	for problemID, embedding := range embeddings {
		// Find trajectories with this problem ID
		for _, trajectory := range ei.store.trajectories {
			if trajectory.Problem != nil && ComputeProblemHash(trajectory.Problem) == problemID {
				trajectory.Problem.Embedding = embedding
				// Note: We don't load metadata here, but it could be added if needed
			}
		}
	}

	log.Printf("Loaded %d embeddings from storage", len(embeddings))
	return nil
}

// problemToText converts a problem description to text for embedding
func (ei *EmbeddingIntegration) problemToText(p *ProblemDescription) string {
	text := p.Description

	if p.Context != "" {
		text += ". Context: " + p.Context
	}

	if len(p.Goals) > 0 {
		text += ". Goals: "
		for i, goal := range p.Goals {
			if i > 0 {
				text += ", "
			}
			text += goal
		}
	}

	if p.Domain != "" {
		text += ". Domain: " + p.Domain
	}

	if p.ProblemType != "" {
		text += ". Type: " + p.ProblemType
	}

	return text
}

// generateRecommendation creates a recommendation string from a trajectory
func generateRecommendation(trajectory *ReasoningTrajectory) string {
	strategy := "unknown"
	if trajectory.Approach != nil && trajectory.Approach.Strategy != "" {
		strategy = trajectory.Approach.Strategy
	}

	if trajectory.SuccessScore > 0.8 {
		return fmt.Sprintf("High success approach: %s", strategy)
	} else if trajectory.SuccessScore > 0.5 {
		return fmt.Sprintf("Moderate success approach: %s (consider improvements)", strategy)
	}
	return fmt.Sprintf("Learn from past attempt: %s", strategy)
}