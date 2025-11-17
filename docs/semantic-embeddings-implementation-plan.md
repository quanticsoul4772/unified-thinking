# Semantic Embeddings Implementation Plan

## Executive Summary

This document outlines a plan to add **optional semantic embeddings support** to the unified-thinking server's episodic memory system, addressing the gap identified in our comparison with industry standards while maintaining our core design principles:

- ✅ **Maintain zero-ML-dependency default** (current hash-based similarity)
- ✅ **Add optional semantic understanding** for improved retrieval
- ✅ **Implement hybrid search** combining structured + semantic similarity
- ✅ **Keep backward compatibility** with existing data
- ✅ **Support multiple embedding providers** (API-based and local)

**Expected Impact:**
- 📈 30-50% improvement in trajectory retrieval relevance for semantically similar problems with different wording
- 🎯 Maintain 100% accuracy for exact domain/type matches (current strength)
- 🔄 Enable gradual adoption without disrupting existing deployments

---

## Current State Analysis

### Strengths of Current Implementation

Our existing hash-based similarity approach (`internal/memory/episodic.go:380-418`) excels at:

```go
func calculateProblemSimilarity(p1, p2 *ProblemDescription) float64 {
    // Domain match (30% weight)
    // Problem type match (30% weight)
    // Complexity similarity (20% weight)
    // Goal overlap (20% weight)
    return weighted_score
}
```

**Advantages:**
- ✅ **Deterministic**: Same input always produces same output
- ✅ **Fast**: No model inference overhead
- ✅ **Explainable**: Clear scoring logic
- ✅ **Zero dependencies**: No ML infrastructure required
- ✅ **Perfect for exact matches**: Domain + type matching is 100% accurate

### Limitations

**Semantic Blindness:**
```
Problem 1: "Optimize database query performance"
Problem 2: "Speed up SQL execution time"
Current similarity: 0.0 (different wording)
Semantic similarity: 0.85+ (same concept)
```

**Miss Rate Examples:**
- "fix authentication bug" vs "resolve login security issue"
- "implement caching layer" vs "add memoization for performance"
- "refactor code structure" vs "improve architectural organization"

These semantically identical problems score low due to different terminology.

---

## Design Goals

### Primary Goals

1. **Hybrid Retrieval**: Combine structured matching (current) + semantic understanding (new)
2. **Optional Feature**: Embeddings are opt-in via configuration
3. **Backward Compatible**: Existing deployments work unchanged
4. **Multiple Providers**: Support API-based (OpenAI, etc.) and local models
5. **Performance**: < 50ms embedding generation, < 10ms similarity calculation

### Non-Goals

- ❌ Replace current hash-based system entirely
- ❌ Require ML infrastructure for basic operation
- ❌ Force all users to adopt embeddings
- ❌ Support massive-scale vector search (millions of vectors)

---

## Architecture Options

### Option 1: API-Based Embeddings (Recommended for Phase 1)

**Architecture:**
```
User Problem → Embedding API (OpenAI/Cohere/Anthropic) → 1536d vector → Store in SQLite
Query → Embedding API → Vector → Hybrid search → Ranked results
```

**Pros:**
- ✅ No local ML infrastructure
- ✅ State-of-the-art embeddings (text-embedding-3-small: 1536d, $0.02/1M tokens)
- ✅ Easy to implement (Go HTTP client)
- ✅ Supports multiple providers (OpenAI, Cohere, Anthropic Claude)

**Cons:**
- ⚠️ API latency (50-200ms per request)
- ⚠️ Costs (though minimal: ~$0.20 per 10K problems)
- ⚠️ Requires internet connectivity
- ⚠️ API key management

**Go Implementation:**
```go
// Using go-openai library
import "github.com/sashabaranov/go-openai"

type OpenAIEmbedder struct {
    client *openai.Client
}

func (e *OpenAIEmbedder) Embed(text string) ([]float32, error) {
    resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
        Model: openai.SmallEmbedding3,
        Input: []string{text},
    })
    return resp.Data[0].Embedding, err
}
```

### Option 2: Local Lightweight Embeddings (Phase 2)

**Architecture:**
```
User Problem → Local SIF/Word2Vec → 300d vector → Store in SQLite
Query → Local model → Vector → Hybrid search → Results
```

**Pros:**
- ✅ No API costs
- ✅ No internet dependency
- ✅ Fast (94K embeddings/sec for Word2Vec in Go)
- ✅ Privacy (data never leaves server)

**Cons:**
- ⚠️ Lower quality than API embeddings
- ⚠️ Requires model files (~500MB for good Word2Vec)
- ⚠️ More complex setup

**Go Implementation:**
```go
// Using SIF (Smooth Inverse Frequency) in Go
type SIFEmbedder struct {
    wordVectors map[string][]float32
    wordFreqs   map[string]float64
}

func (e *SIFEmbedder) Embed(text string) ([]float32, error) {
    // Tokenize, weight by IDF, compute weighted average
    // See: https://blogs.nlmatics.com/nlp/sentence-embeddings/2020/08/07/...
}
```

### Option 3: Hybrid Local + API Fallback (Phase 3)

**Architecture:**
```
Try Local → If confident, use local → Else, use API → Cache result
```

**Best of both worlds**: Fast + cheap for common cases, high-quality for edge cases.

---

## Recommended Approach: Hybrid Search with RRF

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Problem Query                            │
└────────────────┬────────────────────────────────────────────┘
                 │
        ┌────────┴────────┐
        │                 │
        ▼                 ▼
┌───────────────┐   ┌─────────────────┐
│ Structured    │   │ Semantic Vector │
│ Search        │   │ Search          │
│ (Current)     │   │ (New)           │
└───────┬───────┘   └────────┬────────┘
        │                    │
        │  Domain: 1.0       │  Cosine: 0.87
        │  Type: 1.0         │
        │  Complexity: 0.9   │
        │  Goals: 0.6        │
        │  → Score: 0.85     │  → Score: 0.87
        │                    │
        └─────────┬──────────┘
                  │
                  ▼
        ┌─────────────────────┐
        │ Reciprocal Rank     │
        │ Fusion (RRF)        │
        │                     │
        │ Final Score =       │
        │  (1/k+rank₁) +      │
        │  (1/k+rank₂)        │
        └──────────┬──────────┘
                   │
                   ▼
        ┌──────────────────────┐
        │ Ranked Results       │
        │ 1. Problem A (0.92)  │
        │ 2. Problem B (0.81)  │
        │ 3. Problem C (0.73)  │
        └──────────────────────┘
```

### Reciprocal Rank Fusion (RRF)

**Algorithm:**
```go
func HybridSearch(query *ProblemDescription, limit int) ([]*TrajectoryMatch, error) {
    // 1. Structured search
    structuredResults := structuredSearch(query)

    // 2. Vector search (if embeddings enabled)
    var vectorResults []*TrajectoryMatch
    if embeddingsEnabled {
        vectorResults = vectorSearch(query)
    }

    // 3. Reciprocal Rank Fusion
    const k = 60 // RRF parameter (typical: 60)
    scores := make(map[string]float64)

    for rank, result := range structuredResults {
        scores[result.ID] += 1.0 / (k + rank + 1)
    }

    for rank, result := range vectorResults {
        scores[result.ID] += 1.0 / (k + rank + 1)
    }

    // 4. Re-rank by combined score
    return rankByScore(scores, limit)
}
```

**Why RRF?**
- ✅ No need to normalize scores from different search methods
- ✅ Well-studied algorithm (used by Elasticsearch, Weaviate, Pinecone)
- ✅ Balances contribution from both methods
- ✅ Simple to implement and understand

---

## Technical Specifications

### Data Model Extensions

#### 1. ProblemDescription (Updated)

```go
type ProblemDescription struct {
    Description  string                 `json:"description"`
    Context      string                 `json:"context"`
    Goals        []string               `json:"goals"`
    Constraints  []string               `json:"constraints"`
    InitialState map[string]interface{} `json:"initial_state"`
    ProblemType  string                 `json:"problem_type"`
    Complexity   float64                `json:"complexity"`
    Domain       string                 `json:"domain"`

    // NEW: Embedding support
    Embedding    []float32              `json:"embedding,omitempty"`     // Vector representation
    EmbeddingMeta *EmbeddingMetadata    `json:"embedding_meta,omitempty"` // Metadata
}

type EmbeddingMetadata struct {
    Model     string    `json:"model"`      // e.g., "text-embedding-3-small"
    Provider  string    `json:"provider"`   // e.g., "openai", "local-sif"
    Dimension int       `json:"dimension"`  // e.g., 1536
    CreatedAt time.Time `json:"created_at"`
    Source    string    `json:"source"`     // "description" or "description+context+goals"
}
```

#### 2. SQLite Schema Updates

```sql
-- Add embedding column to trajectories table
ALTER TABLE trajectories ADD COLUMN embedding BLOB;
ALTER TABLE trajectories ADD COLUMN embedding_model TEXT;
ALTER TABLE trajectories ADD COLUMN embedding_provider TEXT;
ALTER TABLE trajectories ADD COLUMN embedding_dimension INTEGER;

-- Enable sqlite-vec extension
LOAD EXTENSION 'vec0';

-- Create vector search index (using sqlite-vec)
CREATE VIRTUAL TABLE trajectory_embeddings USING vec0(
    trajectory_id TEXT PRIMARY KEY,
    embedding FLOAT[1536]  -- Dimension depends on model
);

-- Index for quick filtering before vector search
CREATE INDEX idx_traj_domain_type ON trajectories(domain, problem_type);
```

#### 3. Configuration

```go
type EmbeddingConfig struct {
    Enabled  bool   `json:"enabled"`           // Master switch
    Provider string `json:"provider"`          // "openai", "cohere", "anthropic", "local-sif", "local-word2vec"
    Model    string `json:"model"`             // Provider-specific model name
    APIKey   string `json:"api_key,omitempty"` // For API providers

    // Hybrid search settings
    UseHybridSearch  bool    `json:"use_hybrid_search"`  // Enable RRF
    RRFParameter     int     `json:"rrf_k"`              // Default: 60
    VectorWeight     float64 `json:"vector_weight"`      // For weighted combination (alternative to RRF)
    StructuredWeight float64 `json:"structured_weight"`  // For weighted combination

    // Caching
    CacheEmbeddings bool          `json:"cache_embeddings"` // Cache computed embeddings
    CacheTTL        time.Duration `json:"cache_ttl"`        // Cache expiration

    // Performance
    BatchSize       int           `json:"batch_size"`       // Batch embedding requests
    MaxConcurrent   int           `json:"max_concurrent"`   // Concurrent API calls
}
```

**Environment Variables:**
```bash
# Enable embeddings
EMBEDDINGS_ENABLED=true
EMBEDDINGS_PROVIDER=openai  # openai, cohere, anthropic, local-sif
EMBEDDINGS_MODEL=text-embedding-3-small
OPENAI_API_KEY=sk-...

# Hybrid search
EMBEDDINGS_HYBRID_SEARCH=true
EMBEDDINGS_RRF_K=60

# Performance
EMBEDDINGS_BATCH_SIZE=100
EMBEDDINGS_MAX_CONCURRENT=5
EMBEDDINGS_CACHE_ENABLED=true
EMBEDDINGS_CACHE_TTL=24h
```

### Component Design

#### 1. Embedder Interface

```go
// Embedder generates vector embeddings from text
type Embedder interface {
    // Embed generates embedding for single text
    Embed(ctx context.Context, text string) ([]float32, error)

    // EmbedBatch generates embeddings for multiple texts (more efficient)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

    // Dimension returns the embedding dimension
    Dimension() int

    // Model returns the model identifier
    Model() string

    // Provider returns the provider name
    Provider() string
}
```

#### 2. Provider Implementations

```go
// OpenAI Embedder
type OpenAIEmbedder struct {
    client    *openai.Client
    model     string
    dimension int
}

func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
    return &OpenAIEmbedder{
        client:    openai.NewClient(apiKey),
        model:     model,
        dimension: 1536, // text-embedding-3-small
    }
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
    resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
        Model: openai.EmbeddingModel(e.model),
        Input: []string{text},
    })
    if err != nil {
        return nil, err
    }
    return resp.Data[0].Embedding, nil
}

func (e *OpenAIEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
    resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
        Model: openai.EmbeddingModel(e.model),
        Input: texts,
    })
    if err != nil {
        return nil, err
    }

    embeddings := make([][]float32, len(resp.Data))
    for i, data := range resp.Data {
        embeddings[i] = data.Embedding
    }
    return embeddings, nil
}
```

```go
// Local SIF Embedder (Phase 2)
type SIFEmbedder struct {
    wordVectors map[string][]float32
    wordFreqs   map[string]float64
    dimension   int
}

func NewSIFEmbedder(vectorFile, freqFile string) (*SIFEmbedder, error) {
    // Load pre-trained word vectors (GloVe, Word2Vec, etc.)
    // Load word frequencies for IDF weighting
    return &SIFEmbedder{
        dimension: 300, // Typical for GloVe/Word2Vec
    }, nil
}

func (e *SIFEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
    // 1. Tokenize text
    tokens := tokenize(text)

    // 2. Get word vectors and weight by IDF
    weighted := make([]float32, e.dimension)
    totalWeight := 0.0

    for _, token := range tokens {
        if vec, ok := e.wordVectors[token]; ok {
            weight := 1.0 / (1.0 + e.wordFreqs[token]) // SIF weighting
            for i, v := range vec {
                weighted[i] += float32(weight) * v
            }
            totalWeight += weight
        }
    }

    // 3. Normalize by total weight
    if totalWeight > 0 {
        for i := range weighted {
            weighted[i] /= float32(totalWeight)
        }
    }

    // 4. Remove first principal component (SIF algorithm)
    // (Requires computing PC from corpus - skip for simplicity)

    return weighted, nil
}
```

#### 3. Vector Similarity Functions

```go
// CosineSimilarity computes cosine similarity between two vectors
func CosineSimilarity(v1, v2 []float32) float64 {
    if len(v1) != len(v2) {
        return 0.0
    }

    var dotProduct, norm1, norm2 float64
    for i := range v1 {
        dotProduct += float64(v1[i] * v2[i])
        norm1 += float64(v1[i] * v1[i])
        norm2 += float64(v2[i] * v2[i])
    }

    if norm1 == 0 || norm2 == 0 {
        return 0.0
    }

    return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// EuclideanDistance computes L2 distance
func EuclideanDistance(v1, v2 []float32) float64 {
    if len(v1) != len(v2) {
        return math.MaxFloat64
    }

    var sum float64
    for i := range v1 {
        diff := float64(v1[i] - v2[i])
        sum += diff * diff
    }

    return math.Sqrt(sum)
}
```

#### 4. Hybrid Search Implementation

```go
// HybridSearcher combines structured and vector search
type HybridSearcher struct {
    store    *EpisodicMemoryStore
    embedder Embedder
    config   *EmbeddingConfig
}

func (h *HybridSearcher) Search(ctx context.Context, problem *ProblemDescription, limit int) ([]*TrajectoryMatch, error) {
    // 1. Structured search (current implementation)
    structuredResults := h.structuredSearch(problem)

    // 2. Vector search (if embeddings enabled)
    var vectorResults []*TrajectoryMatch
    if h.config.Enabled && h.config.UseHybridSearch {
        vectorResults, _ = h.vectorSearch(ctx, problem)
    }

    // 3. Combine using RRF
    if len(vectorResults) == 0 {
        return structuredResults[:min(limit, len(structuredResults))], nil
    }

    return h.reciprocalRankFusion(structuredResults, vectorResults, limit)
}

func (h *HybridSearcher) reciprocalRankFusion(
    structured, vector []*TrajectoryMatch,
    limit int,
) ([]*TrajectoryMatch, error) {
    const k = 60 // RRF parameter

    scores := make(map[string]float64)
    trajectories := make(map[string]*TrajectoryMatch)

    // Score structured results
    for rank, match := range structured {
        id := match.Trajectory.ID
        scores[id] += 1.0 / float64(k+rank+1)
        trajectories[id] = match
    }

    // Score vector results
    for rank, match := range vector {
        id := match.Trajectory.ID
        scores[id] += 1.0 / float64(k+rank+1)
        if _, exists := trajectories[id]; !exists {
            trajectories[id] = match
        }
    }

    // Sort by combined score
    type scoredMatch struct {
        match *TrajectoryMatch
        score float64
    }

    scored := make([]scoredMatch, 0, len(scores))
    for id, score := range scores {
        scored = append(scored, scoredMatch{
            match: trajectories[id],
            score: score,
        })
    }

    sort.Slice(scored, func(i, j int) bool {
        return scored[i].score > scored[j].score
    })

    // Return top limit results
    results := make([]*TrajectoryMatch, 0, min(limit, len(scored)))
    for i := 0; i < min(limit, len(scored)); i++ {
        results = append(results, scored[i].match)
    }

    return results, nil
}

func (h *HybridSearcher) vectorSearch(ctx context.Context, problem *ProblemDescription) ([]*TrajectoryMatch, error) {
    // 1. Generate embedding for query
    queryText := h.problemToText(problem)
    queryEmb, err := h.embedder.Embed(ctx, queryText)
    if err != nil {
        return nil, err
    }

    // 2. Find similar trajectories using cosine similarity
    allTrajectories := h.store.GetAllTrajectories()
    matches := make([]*TrajectoryMatch, 0)

    for _, traj := range allTrajectories {
        if traj.Problem == nil || traj.Problem.Embedding == nil {
            continue
        }

        similarity := CosineSimilarity(queryEmb, traj.Problem.Embedding)
        if similarity > 0.5 { // Minimum threshold
            matches = append(matches, &TrajectoryMatch{
                Trajectory:      traj,
                SimilarityScore: similarity,
                RelevanceFactors: []string{fmt.Sprintf("Vector similarity: %.2f", similarity)},
            })
        }
    }

    // 3. Sort by similarity
    sort.Slice(matches, func(i, j int) bool {
        return matches[i].SimilarityScore > matches[j].SimilarityScore
    })

    return matches, nil
}

func (h *HybridSearcher) problemToText(p *ProblemDescription) string {
    // Combine description, context, and goals into single text
    parts := []string{p.Description}
    if p.Context != "" {
        parts = append(parts, p.Context)
    }
    if len(p.Goals) > 0 {
        parts = append(parts, "Goals: "+strings.Join(p.Goals, ", "))
    }
    return strings.Join(parts, ". ")
}
```

---

## Implementation Phases

### Phase 1: API-Based Embeddings (4-6 weeks)

**Week 1-2: Foundation**
- [ ] Add embeddings configuration to server
- [ ] Implement Embedder interface
- [ ] Implement OpenAIEmbedder
- [ ] Add embedding fields to ProblemDescription
- [ ] Update SQLite schema

**Week 3-4: Core Functionality**
- [ ] Implement embedding generation on problem storage
- [ ] Add vector similarity search
- [ ] Implement RRF hybrid search
- [ ] Add embedding caching layer

**Week 5-6: Integration & Testing**
- [ ] Update MCP tool handlers
- [ ] Add configuration options
- [ ] Write comprehensive tests
- [ ] Performance benchmarks
- [ ] Documentation

**Deliverables:**
- ✅ Optional OpenAI embedding support
- ✅ Hybrid search with RRF
- ✅ Backward compatible
- ✅ Configuration via environment variables

### Phase 2: Local Embeddings (6-8 weeks)

**Week 1-3: Research & Implementation**
- [ ] Evaluate SIF vs Word2Vec for Go
- [ ] Implement chosen local embedder
- [ ] Create model download/setup tooling
- [ ] Optimize for performance

**Week 4-5: Integration**
- [ ] Add local embedder to provider registry
- [ ] Implement automatic provider fallback
- [ ] Add benchmarks comparing local vs API

**Week 6-8: Polish**
- [ ] Documentation for local setup
- [ ] Performance tuning
- [ ] Testing across providers

**Deliverables:**
- ✅ Local SIF or Word2Vec embedder
- ✅ No API dependency option
- ✅ Provider abstraction working

### Phase 3: Advanced Features (4-6 weeks)

**Optional enhancements:**
- [ ] Multiple embedding models simultaneously
- [ ] Embedding model fine-tuning on user data
- [ ] Advanced hybrid ranking algorithms
- [ ] Embedding compression (PCA/quantization)
- [ ] Batch processing optimizations

---

## Performance Considerations

### Latency Analysis

**Current System (No Embeddings):**
```
Problem storage: ~1ms
Similarity search: ~5ms (for 1000 trajectories)
Total: ~6ms
```

**With API Embeddings:**
```
Problem storage:
  - Hash embedding: ~1ms
  - API call: 50-200ms (depends on API)
  - Store: ~2ms
  Total: 53-203ms

Similarity search:
  - Structured: ~5ms
  - API call: 50-200ms
  - Vector search: ~10ms (in-memory cosine)
  - RRF: ~2ms
  Total: 67-217ms
```

**With Local Embeddings:**
```
Problem storage:
  - SIF embedding: ~0.01ms (94K/sec)
  - Store: ~2ms
  Total: ~3ms

Similarity search:
  - Structured: ~5ms
  - SIF embedding: ~0.01ms
  - Vector search: ~10ms
  - RRF: ~2ms
  Total: ~17ms
```

### Optimization Strategies

1. **Embedding Cache**
   - Cache embeddings by hash of input text
   - TTL: 24 hours default
   - Expected hit rate: 30-50% for repeated queries

2. **Batch Processing**
   - Batch API calls (up to 100 texts per request)
   - Reduces API overhead 10x

3. **Pre-filtering**
   - Use structured filters before vector search
   - Reduces vector comparison count 5-10x

4. **Approximate Nearest Neighbors (Future)**
   - Use sqlite-vec's indexing for >10K trajectories
   - Trade 5% accuracy for 100x speed

### Storage Impact

**Per Trajectory:**
```
Current: ~2KB (JSON data)
With embeddings: ~2KB + 6KB (1536 floats × 4 bytes) = ~8KB
Increase: 4x
```

**For 10,000 trajectories:**
```
Current: 20MB
With embeddings: 80MB
Still well within SQLite capacity
```

---

## Migration Strategy

### Backward Compatibility

1. **Embedding fields are optional**
   - `Embedding []float32 json:"embedding,omitempty"`
   - Existing data loads without embeddings

2. **Graceful degradation**
   - If embeddings disabled, use structured search only
   - If embedding generation fails, log warning and continue

3. **Progressive enhancement**
   - New trajectories get embeddings if enabled
   - Old trajectories can be backfilled on-demand

### Backfill Strategy

```go
func BackfillEmbeddings(ctx context.Context, store *EpisodicMemoryStore, embedder Embedder) error {
    trajectories := store.GetAllTrajectories()

    for _, traj := range trajectories {
        if traj.Problem != nil && traj.Problem.Embedding == nil {
            text := problemToText(traj.Problem)
            emb, err := embedder.Embed(ctx, text)
            if err != nil {
                log.Printf("Failed to embed trajectory %s: %v", traj.ID, err)
                continue
            }

            traj.Problem.Embedding = emb
            traj.Problem.EmbeddingMeta = &EmbeddingMetadata{
                Model:     embedder.Model(),
                Provider:  embedder.Provider(),
                Dimension: embedder.Dimension(),
                CreatedAt: time.Now(),
                Source:    "backfill",
            }

            store.UpdateTrajectory(ctx, traj)
        }
    }

    return nil
}
```

**Backfill Options:**
1. **On-demand**: Embed when accessed
2. **Batch**: Background job processes N per minute
3. **Manual**: Admin tool triggers backfill

---

## Testing Strategy

### Unit Tests

```go
func TestOpenAIEmbedder(t *testing.T) {
    embedder := NewOpenAIEmbedder(apiKey, "text-embedding-3-small")

    // Test single embedding
    emb, err := embedder.Embed(ctx, "Test problem")
    assert.NoError(t, err)
    assert.Equal(t, 1536, len(emb))

    // Test batch embedding
    embs, err := embedder.EmbedBatch(ctx, []string{"Problem 1", "Problem 2"})
    assert.NoError(t, err)
    assert.Equal(t, 2, len(embs))
}

func TestCosineSimilarity(t *testing.T) {
    v1 := []float32{1, 0, 0}
    v2 := []float32{1, 0, 0}
    v3 := []float32{0, 1, 0}

    assert.Equal(t, 1.0, CosineSimilarity(v1, v2))  // Identical
    assert.Equal(t, 0.0, CosineSimilarity(v1, v3))  // Orthogonal
}

func TestHybridSearch(t *testing.T) {
    searcher := NewHybridSearcher(store, embedder, config)

    problem := &ProblemDescription{
        Description: "Optimize database queries",
    }

    results, err := searcher.Search(ctx, problem, 5)
    assert.NoError(t, err)
    assert.LessOrEqual(t, len(results), 5)

    // Results should be ranked by combined score
    for i := 1; i < len(results); i++ {
        assert.GreaterOrEqual(t, results[i-1].SimilarityScore, results[i].SimilarityScore)
    }
}
```

### Integration Tests

```go
func TestEmbeddingWorkflow(t *testing.T) {
    // 1. Store problem with embedding
    problem := &ProblemDescription{
        Description: "Fix authentication bug",
        Domain: "security",
    }

    sessionID := "test-session"
    tracker.StartSession(ctx, sessionID, problem)
    tracker.RecordStep(ctx, sessionID, &ReasoningStep{...})
    trajectory, err := tracker.CompleteSession(ctx, sessionID, outcome)

    // 2. Verify embedding was generated
    assert.NotNil(t, trajectory.Problem.Embedding)
    assert.Equal(t, 1536, len(trajectory.Problem.Embedding))

    // 3. Search with similar problem
    similarProblem := &ProblemDescription{
        Description: "Resolve login security issue", // Semantically similar
    }

    matches, err := searcher.Search(ctx, similarProblem, 5)
    assert.NoError(t, err)

    // 4. Original problem should be in top results
    found := false
    for _, match := range matches {
        if match.Trajectory.ID == trajectory.ID {
            found = true
            assert.Greater(t, match.SimilarityScore, 0.7) // High similarity
            break
        }
    }
    assert.True(t, found)
}
```

### Benchmark Tests

```go
func BenchmarkStructuredSearch(b *testing.B) {
    for i := 0; i < b.N; i++ {
        searcher.structuredSearch(problem)
    }
}

func BenchmarkVectorSearch(b *testing.B) {
    for i := 0; i < b.N; i++ {
        searcher.vectorSearch(ctx, problem)
    }
}

func BenchmarkHybridSearch(b *testing.B) {
    for i := 0; i < b.N; i++ {
        searcher.Search(ctx, problem, 10)
    }
}

func BenchmarkEmbeddingGeneration(b *testing.B) {
    for i := 0; i < b.N; i++ {
        embedder.Embed(ctx, "Test problem description")
    }
}
```

### Quality Tests

```go
func TestSemanticRetrieval(t *testing.T) {
    testCases := []struct {
        stored  string
        query   string
        minSim  float64
    }{
        {
            stored: "Optimize database query performance",
            query:  "Speed up SQL execution time",
            minSim: 0.75,
        },
        {
            stored: "Fix authentication bug in login flow",
            query:  "Resolve security issue with user sign-in",
            minSim: 0.70,
        },
        {
            stored: "Implement caching layer for API responses",
            query:  "Add memoization to improve API performance",
            minSim: 0.65,
        },
    }

    for _, tc := range testCases {
        // Store trajectory with first description
        storeTrajectory(tc.stored)

        // Search with semantic variation
        matches := search(tc.query)

        // Should find the stored trajectory
        assert.Greater(t, matches[0].SimilarityScore, tc.minSim)
    }
}
```

---

## Success Metrics

### Quantitative Metrics

1. **Retrieval Quality**
   - **Recall@5**: % of relevant trajectories in top 5 results
     - Target: >85% (up from ~60% with structured only)
   - **MRR (Mean Reciprocal Rank)**: Average rank of first relevant result
     - Target: >0.75
   - **NDCG@10**: Ranking quality metric
     - Target: >0.80

2. **Performance**
   - **Embedding generation latency**: <100ms (API) or <1ms (local)
   - **Search latency**: <200ms total (including hybrid)
   - **Storage overhead**: <5x (currently ~4x expected)

3. **Adoption**
   - **Opt-in rate**: % of users enabling embeddings
     - Target: >30% in first 6 months
   - **API cost per user**: <$1/month for typical usage
   - **Cache hit rate**: >40%

### Qualitative Metrics

1. **User Feedback**
   - Survey: "Are recommendations more relevant?"
   - Target: >75% positive

2. **Developer Experience**
   - Setup time: <15 minutes for API embeddings
   - Documentation clarity: Reviewed by 3 external developers

3. **Reliability**
   - Embedding service uptime: >99.5%
   - Graceful degradation: 100% (never block core functionality)

---

## Dependencies

### Required Go Libraries

```go
// go.mod additions
require (
    github.com/sashabaranov/go-openai v1.20.0     // OpenAI API
    github.com/cohere-ai/cohere-go v2.0.0         // Cohere API (optional)
    github.com/anthropics/anthropic-sdk-go v0.1.0 // Claude API (optional)

    // For local embeddings (Phase 2)
    github.com/nlpodyssey/spago v1.1.0            // NLP toolkit
    github.com/james-bowman/nlp v0.1.0            // Text processing

    // Already have
    modernc.org/sqlite v1.28.0                    // SQLite driver
)
```

### External Dependencies

**Phase 1 (API):**
- OpenAI API account (free tier: $5 credit)
- OR Cohere API account (free tier: 100K tokens/month)
- OR Anthropic Claude API account

**Phase 2 (Local):**
- Pre-trained word vectors (GloVe/Word2Vec)
  - Download: ~500MB
  - Memory: ~1GB loaded
- OR SIF implementation with smaller vectors (~50MB)

### SQLite Extensions

**sqlite-vec** (recommended):
```bash
# Installation
wget https://github.com/asg017/sqlite-vec/releases/download/v0.1.0/vec0.so
# Or compile from source
```

**Load in Go:**
```go
import (
    "database/sql"
    _ "modernc.org/sqlite"
)

db, _ := sql.Open("sqlite", "file:memory.db?_loc=auto")
db.Exec("SELECT load_extension('vec0')")
```

---

## Risk Analysis

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| API rate limiting | Medium | Medium | Implement backoff, caching, batching |
| Embedding quality varies | Low | Medium | Test multiple providers, allow switching |
| SQLite vec extension issues | Low | High | Have pure-Go fallback (brute-force search) |
| Performance degradation | Medium | Medium | Extensive benchmarking, optimization |
| Storage bloat | Low | Low | Embeddings are optional, can be disabled |

### Operational Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| API costs exceed budget | Low | Medium | Set rate limits, monitor usage |
| API downtime | Medium | Low | Graceful degradation, fallback to structured |
| Configuration complexity | Medium | Low | Sane defaults, clear documentation |
| User confusion | Low | Low | Feature is opt-in, well-documented |

### Mitigation Strategies

1. **Graceful Degradation**
   ```go
   if embedder != nil && config.Enabled {
       // Try embedding
       emb, err := embedder.Embed(ctx, text)
       if err != nil {
           log.Printf("Embedding failed: %v. Falling back to structured search", err)
           // Continue without embeddings
       }
   }
   ```

2. **Cost Controls**
   ```go
   type RateLimiter struct {
       maxPerHour int
       current    int
   }

   func (r *RateLimiter) Allow() bool {
       return r.current < r.maxPerHour
   }
   ```

3. **Monitoring**
   ```go
   type EmbeddingMetrics struct {
       TotalEmbeddings   int64
       FailedEmbeddings  int64
       AverageLatency    time.Duration
       CacheHitRate      float64
       EstimatedCost     float64
   }
   ```

---

## Documentation Plan

### User Documentation

1. **Quick Start Guide**
   - Enable embeddings in 5 minutes
   - Copy-paste configuration
   - Verify it's working

2. **Configuration Reference**
   - All environment variables
   - Provider-specific settings
   - Performance tuning

3. **Comparison Guide**
   - When to use embeddings vs structured
   - Cost-benefit analysis
   - Performance characteristics

### Developer Documentation

1. **Architecture Overview**
   - Component diagram
   - Data flow
   - Integration points

2. **API Reference**
   - Embedder interface
   - Configuration structs
   - Public functions

3. **Testing Guide**
   - How to run tests
   - How to add provider
   - Benchmark guidelines

### Examples

```markdown
# Enabling Embeddings

## OpenAI (Recommended)

1. Get API key from platform.openai.com
2. Set environment variables:
   ```bash
   export EMBEDDINGS_ENABLED=true
   export EMBEDDINGS_PROVIDER=openai
   export EMBEDDINGS_MODEL=text-embedding-3-small
   export OPENAI_API_KEY=sk-...
   ```
3. Restart server
4. Verify:
   ```bash
   # Check logs for "Embeddings enabled: openai/text-embedding-3-small"
   ```

## Cost Estimate

- Model: text-embedding-3-small
- Price: $0.02 per 1M tokens
- Average problem: ~100 tokens
- 10,000 problems/month: ~1M tokens = $0.02/month

Negligible cost for typical usage!
```

---

## Conclusion

### Summary

This implementation plan provides a **pragmatic path** to adding semantic embeddings support while maintaining our core design principles:

✅ **Maintains simplicity**: Embeddings are optional, not required
✅ **Backward compatible**: Existing deployments work unchanged
✅ **Flexible**: Supports multiple providers (API and local)
✅ **Performant**: Hybrid search with RRF balances quality and speed
✅ **Cost-effective**: API embeddings cost <$1/month for typical usage
✅ **Production-ready**: Comprehensive testing, monitoring, and documentation

### Recommended Next Steps

1. **Approve architecture** (1 week)
   - Review this plan with team
   - Select primary embedding provider
   - Define success criteria

2. **Phase 1 Implementation** (4-6 weeks)
   - Start with OpenAI API embeddings
   - Implement hybrid search with RRF
   - Comprehensive testing

3. **Pilot Deployment** (2 weeks)
   - Deploy to test environment
   - Gather user feedback
   - Measure performance metrics

4. **Phase 2 Planning** (ongoing)
   - Evaluate local embedding options
   - Plan advanced features
   - Iterate based on feedback

### Expected Outcomes

After Phase 1 completion:
- 📈 30-50% improvement in retrieval relevance for semantic queries
- 🚀 Maintain <200ms search latency
- 💰 <$1/month API costs for typical users
- 🎯 >85% Recall@5 on semantic similarity tests
- ✅ Zero breaking changes to existing functionality

This positions unified-thinking as a **best-in-class** episodic memory system with both structured precision AND semantic understanding.

---

**Document Version**: 1.0
**Date**: 2025-11-17
**Status**: Proposal - Awaiting Approval
**Next Review**: After Phase 1 completion
