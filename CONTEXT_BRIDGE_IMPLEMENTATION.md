# Cross-Session Context Bridging - Implementation Plan

## Objective

Add automatic context retrieval to unified-thinking MCP server. When Claude uses tools, surface similar past reasoning trajectories without explicit user action.

## Approach

Hybrid implementation combining semantic embeddings (Voyage AI) with concept-based similarity. When VOYAGE_API_KEY is set, uses cosine similarity on embeddings (70% weight) combined with Jaccard similarity on concepts (30% weight). Falls back to concept-only matching when embeddings unavailable.

## Implementation Status

**Completed:**
- Core context bridge with signature extraction, matching, and enrichment
- SQLite storage with schema migrations (v1-v5)
- LRU caching for recent signatures
- Metrics collection and exposure
- Semantic embeddings integration (Voyage AI)
- Graceful degradation with visible status in responses
- Hybrid similarity mode (embedding + concept)

**Not implemented:**
- Backfill utility for existing trajectories
- Gradual rollout (percentage-based feature flag)
- Alert thresholds and notifications

## Critical Constraints

1. Existing trajectories lack context signatures - need migration
2. SQLite full table scan at scale will degrade performance
3. Simple tokenization insufficient for technical terminology
4. Must support feature flag for rollback
5. Need metrics before launch to measure impact

## Non-Functional Requirements

| ID | Requirement | Target | Rationale |
|----|-------------|--------|-----------|
| NFR-1 | Enrichment latency p99 | < 50ms with 10,000 signatures | User experience, no perceptible delay |
| NFR-2 | Memory usage | < 500MB with full LRU cache | Prevent OOM in constrained environments |
| NFR-3 | Concurrent requests | 100 without degradation | Support high-throughput scenarios |
| NFR-4 | CPU overhead | < 10% during continuous enrichment | Don't impact other server operations |
| NFR-5 | Storage query latency | < 10ms for candidate retrieval | Fast path for common operations |
| NFR-6 | Backfill throughput | 100 trajectories/second max | Rate limit to prevent DB overload |

## Error Handling & Degradation

### Failure Modes and Recovery

| ID | Failure Mode | Behavior | Logging | Metrics | Recovery |
|----|--------------|----------|---------|---------|----------|
| EHR-1 | Storage failure | Return original response (no enrichment) | ERROR with trajectory count | `error_count` | Automatic on next request |
| EHR-2 | Timeout > 2s | Return degraded response with status | WARN with elapsed time | `timeout_count` | Automatic on next request |
| EHR-3 | Embedding timeout (500ms) | Continue with concept-only similarity | WARN with error | n/a | Automatic |
| EHR-4 | Invalid signature | Skip without error | WARN with reason | n/a | N/A |
| EHR-5 | Cache overflow | Evict LRU entries | DEBUG | n/a | Automatic |

### Degradation Strategy

**Principle**: Context enrichment is non-critical. Failures must never block or degrade the primary tool response.

1. **Visible degradation**: Return response with `context_bridge.status = "degraded"` and reason
2. **Embedding fallback**: If embedding generation fails, continue with concept-only similarity
3. **No cascading**: Enrichment failures don't affect subsequent requests
4. **Health indicator**: Metrics endpoint exposes degradation state

**Response format when degraded:**
```json
{
  "result": { ... },
  "context_bridge": {
    "version": "1.0",
    "status": "degraded",
    "reason": "timeout",
    "detail": "Context bridge timeout exceeded (2s)",
    "matches": []
  }
}
```

**Response format when embedding fails but matching succeeds:**
```json
{
  "result": { ... },
  "context_bridge": {
    "version": "1.0",
    "matches": [...],
    "embedding_status": "failed",
    "embedding_error": "context deadline exceeded",
    "similarity_mode": "concept_only"
  }
}
```

### Alert Thresholds

| Metric | Warning | Critical | Action |
|--------|---------|----------|--------|
| p99 latency | > 75ms for 5 min | > 100ms for 5 min | Investigate storage |
| Error rate | > 0.5% for 5 min | > 1% for 5 min | Check storage health |
| Cache hit rate | < 40% for 10 min | < 20% for 10 min | Review cache sizing |

## Specification by Example

### Scenario 1: Identical Problem Detection

```gherkin
Given a trajectory exists with:
  | field       | value                              |
  | description | "How to optimize database queries" |
  | success     | 0.9                                |
  | domain      | "engineering"                      |

When Claude calls think with:
  | parameter | value                              |
  | content   | "How to optimize database queries" |
  | mode      | "linear"                           |

Then response includes context_bridge with:
  | field                        | constraint              |
  | matches.length               | >= 1                    |
  | matches[0].similarity        | >= 0.95                 |
  | recommendation               | contains "high success" |
```

### Scenario 2: Similar Domain Match

```gherkin
Given a trajectory exists with:
  | field       | value                          |
  | description | "SQL query performance tuning" |
  | domain      | "engineering"                  |

When Claude calls think with:
  | parameter | value                                  |
  | content   | "Best practices database optimization" |
  | domain    | "engineering"                          |

Then response includes context_bridge with:
  | field                        | constraint |
  | matches.length               | >= 1       |
  | matches[0].similarity        | 0.6 - 0.8  |
```

### Scenario 3: Unrelated Topics - No Match

```gherkin
Given a trajectory exists with:
  | field       | value                       |
  | description | "Frontend component design" |
  | domain      | "frontend"                  |

When Claude calls think with:
  | parameter | value                         |
  | content   | "Database query optimization" |
  | domain    | "backend"                     |

Then response includes context_bridge with:
  | field          | constraint                        |
  | matches.length | 0 OR matches[0].similarity < 0.3 |
```

### Scenario 4: Feature Flag Disabled

```gherkin
Given context bridge is disabled via feature flag

When Claude calls think with any parameters

Then response does NOT include context_bridge field
And original response is returned unchanged
```

### Scenario 5: Storage Failure Graceful Degradation

```gherkin
Given storage is unavailable

When Claude calls think with:
  | content | "Any problem description" |

Then response does NOT include context_bridge field
And original response is returned unchanged
And error is logged at ERROR level
And error_count metric is incremented
```

### Expected Similarity Ranges

| Content A | Content B | Expected Similarity | Rationale |
|-----------|-----------|---------------------|-----------|
| "optimize database queries" | "optimize database queries" | 0.95 - 1.0 | Identical |
| "optimize database queries" | "speed up SQL execution" | 0.7 - 0.9 | Same domain + intent |
| "Go error handling patterns" | "Golang panic recovery" | 0.5 - 0.7 | Related concepts |
| "database optimization" | "frontend component design" | 0.0 - 0.2 | Different domains |
| "API rate limiting" | "throttling HTTP requests" | 0.6 - 0.8 | Semantic similarity |

## Architecture

### Component Structure

```
server/
  internal/
    contextbridge/
      signature.go       - Extract problem fingerprints
      similarity.go      - Calculate trajectory similarity  
      matcher.go         - Match signatures to trajectories
      middleware.go      - Enrich tool responses
      config.go         - Feature flag and configuration
    storage/
      migrations/
        006_context_signatures.sql
```

### Design Decisions

**Similarity Algorithm**: Hybrid embedding + concept scoring

When VOYAGE_API_KEY is set (semantic mode):
- Embedding cosine similarity (weight: 0.7)
- Concept Jaccard similarity (weight: 0.3)

When no API key (concept-only mode):
- Concept overlap: Jaccard similarity (weight: 0.5)
- Domain match: binary (weight: 0.2)
- Tool sequence: overlap ratio (weight: 0.2)
- Complexity: distance metric (weight: 0.1)

**Embedding Configuration**:
- Provider: Voyage AI
- Model: voyage-3-lite (512 dimensions)
- Storage: SQLite BLOB column (serialized float32)
- Sub-timeout: 500ms for embedding generation

**Performance Strategy**:
- Index on domain and fingerprint prefix
- Limit candidates to 50 before similarity calculation
- Cache recent signature lookups (LRU, 100 entries)
- Share embedder instance between auto mode and context bridge

**Extensibility Points**:
- SimilarityCalculator interface for swappable algorithms
- ConceptExtractor interface for swappable NLP
- Embedder interface for swappable embedding providers

## Implementation Phases

### Phase 1: Data Layer (Est: 2-3 days, 8-12 hours)

#### 1.1 Database Schema

File: `server/internal/storage/migrations/006_context_signatures.sql`

```sql
-- Note: No foreign key on trajectory_id since episodic memory stores trajectories in-memory
CREATE TABLE IF NOT EXISTS context_signatures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trajectory_id TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    fingerprint_prefix TEXT NOT NULL,  -- First 8 chars for indexing
    domain TEXT,
    key_concepts TEXT,      -- JSON array
    tool_sequence TEXT,     -- JSON array
    complexity REAL,
    embedding BLOB,         -- Semantic embedding (serialized float32 vector)
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX idx_context_domain ON context_signatures(domain);
CREATE INDEX idx_context_prefix ON context_signatures(fingerprint_prefix);
CREATE INDEX idx_context_trajectory ON context_signatures(trajectory_id);
```

#### 1.2 Storage Interface Extension

File: `server/internal/storage/storage.go`

Add methods:
```go
type Storage interface {
    // ... existing methods ...

    StoreContextSignature(trajectoryID string, sig *contextbridge.Signature) error
    // FindCandidatesWithSignatures returns candidates with their signatures in a single query
    // This avoids N+1 query pattern - do NOT use separate GetSignatureByTrajectoryID calls
    FindCandidatesWithSignatures(domain string, fingerprintPrefix string, limit int) ([]*contextbridge.CandidateWithSignature, error)
}

// CandidateWithSignature combines trajectory metadata with its signature
// to enable single-query retrieval and avoid N+1 performance issues
type CandidateWithSignature struct {
    TrajectoryID string
    SessionID    string
    Description  string
    SuccessScore float64
    QualityScore float64
    Signature    *Signature
}
```

#### 1.3 SQLite Implementation

File: `server/internal/storage/sqlite_storage.go`

```go
func (s *SQLiteStorage) StoreContextSignature(trajectoryID string, sig *contextbridge.Signature) error {
    concepts, _ := json.Marshal(sig.KeyConcepts)
    tools, _ := json.Marshal(sig.ToolSequence)

    prefix := ""
    if len(sig.Fingerprint) >= 8 {
        prefix = sig.Fingerprint[:8]
    }

    _, err := s.db.Exec(`
        INSERT INTO context_signatures
        (trajectory_id, fingerprint, fingerprint_prefix, domain, key_concepts, tool_sequence, complexity)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, trajectoryID, sig.Fingerprint, prefix, sig.Domain, concepts, tools, sig.Complexity)

    return err
}

// FindCandidatesWithSignatures returns candidates with their signatures in a single query
// This avoids N+1 query pattern for better performance
func (s *SQLiteStorage) FindCandidatesWithSignatures(domain string, fingerprintPrefix string, limit int) ([]*contextbridge.CandidateWithSignature, error) {
    query := `
        SELECT
            cs.trajectory_id,
            t.session_id,
            t.description,
            t.success_score,
            t.quality_score,
            cs.fingerprint,
            cs.domain,
            cs.key_concepts,
            cs.tool_sequence,
            cs.complexity
        FROM context_signatures cs
        JOIN trajectories t ON cs.trajectory_id = t.id
        WHERE (cs.domain = ? OR cs.fingerprint_prefix = ?)
        ORDER BY t.created_at DESC
        LIMIT ?
    `

    rows, err := s.db.Query(query, domain, fingerprintPrefix, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var candidates []*contextbridge.CandidateWithSignature
    for rows.Next() {
        var c contextbridge.CandidateWithSignature
        var conceptsJSON, toolsJSON string

        err := rows.Scan(
            &c.TrajectoryID,
            &c.SessionID,
            &c.Description,
            &c.SuccessScore,
            &c.QualityScore,
            &c.Signature.Fingerprint,
            &c.Signature.Domain,
            &conceptsJSON,
            &toolsJSON,
            &c.Signature.Complexity,
        )
        if err != nil {
            return nil, err
        }

        // Parse JSON arrays
        c.Signature = &contextbridge.Signature{}
        json.Unmarshal([]byte(conceptsJSON), &c.Signature.KeyConcepts)
        json.Unmarshal([]byte(toolsJSON), &c.Signature.ToolSequence)

        candidates = append(candidates, &c)
    }

    return candidates, rows.Err()
}
```

#### 1.4 Migration for Existing Trajectories

File: `server/internal/storage/backfill.go`

```go
// BackfillContextSignatures generates signatures for existing trajectories
func (s *SQLiteStorage) BackfillContextSignatures() error {
    // Get all trajectories without signatures
    // For each, attempt to extract signature from description/metadata
    // Store signature
    // Log failures for manual review
}
```

**Deliverables**:
- Schema migration file
- Storage interface methods
- SQLite implementation
- Backfill utility
- Unit tests for storage methods

**Success Criteria**:
- Signatures stored and retrieved without errors
- Index performance measured (query should be <10ms on 1000 trajectories)
- Backfill runs without crashes

### Phase 2: Core Logic (Est: 3-4 days, 12-16 hours)

#### 2.1 Signature Extraction

File: `server/internal/contextbridge/signature.go`

```go
package contextbridge

import (
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

type Signature struct {
    Fingerprint  string    `json:"fingerprint"`
    Domain       string    `json:"domain"`
    KeyConcepts  []string  `json:"key_concepts"`
    ToolSequence []string  `json:"tool_sequence"`
    Complexity   float64   `json:"complexity"`
    Embedding    []float32 `json:"embedding,omitempty"` // Semantic embedding for similarity
}

// ConceptExtractor interface for swappable extraction strategies
type ConceptExtractor interface {
    Extract(text string) []string
}

// SimpleExtractor uses basic tokenization and stop word filtering
type SimpleExtractor struct {
    stopWords map[string]bool
}

func NewSimpleExtractor() *SimpleExtractor {
    stopWords := map[string]bool{
        "the": true, "a": true, "an": true, "and": true, "or": true,
        "but": true, "in": true, "on": true, "at": true, "to": true,
        "for": true, "of": true, "with": true, "by": true, "from": true,
        "as": true, "is": true, "was": true, "be": true, "been": true,
    }
    return &SimpleExtractor{stopWords: stopWords}
}

func (e *SimpleExtractor) Extract(text string) []string {
    words := strings.Fields(strings.ToLower(text))
    
    concepts := make([]string, 0)
    seen := make(map[string]bool)
    
    for _, word := range words {
        // Remove punctuation
        word = strings.Trim(word, ".,!?;:\"'()[]{}") 
        
        if len(word) > 3 && !e.stopWords[word] && !seen[word] {
            concepts = append(concepts, word)
            seen[word] = true
        }
    }
    
    return concepts
}

// ExtractSignature creates a signature from tool call parameters
func ExtractSignature(toolName string, params map[string]interface{}, extractor ConceptExtractor) (*Signature, error) {
    sig := &Signature{
        ToolSequence: []string{toolName},
    }
    
    // Extract text content from various parameter names
    contentSources := []string{"content", "description", "problem", "situation", "question"}
    var textContent string
    
    for _, key := range contentSources {
        if val, ok := params[key].(string); ok && val != "" {
            textContent = val
            break
        }
    }
    
    if textContent == "" {
        return nil, nil // No extractable content
    }
    
    // Generate fingerprint
    hash := sha256.Sum256([]byte(strings.ToLower(textContent)))
    sig.Fingerprint = hex.EncodeToString(hash[:])
    
    // Extract concepts
    sig.KeyConcepts = extractor.Extract(textContent)
    
    // Extract domain if present
    if domain, ok := params["domain"].(string); ok {
        sig.Domain = domain
    }
    
    // Estimate complexity based on content length and concept count
    wordCount := len(strings.Fields(textContent))
    sig.Complexity = 0.3 + (float64(wordCount)/200.0)*0.4 + (float64(len(sig.KeyConcepts))/20.0)*0.3
    if sig.Complexity > 1.0 {
        sig.Complexity = 1.0
    }
    
    return sig, nil
}
```

#### 2.2 Similarity Calculation

File: `server/internal/contextbridge/similarity.go`

```go
package contextbridge

import "math"

type SimilarityCalculator interface {
    Calculate(sig1, sig2 *Signature) float64
}

type WeightedSimilarity struct {
    ConceptWeight    float64
    DomainWeight     float64
    ToolWeight       float64
    ComplexityWeight float64
}

func NewDefaultSimilarity() *WeightedSimilarity {
    return &WeightedSimilarity{
        ConceptWeight:    0.5,
        DomainWeight:     0.2,
        ToolWeight:       0.2,
        ComplexityWeight: 0.1,
    }
}

func (ws *WeightedSimilarity) Calculate(sig1, sig2 *Signature) float64 {
    conceptSim := jaccardSimilarity(sig1.KeyConcepts, sig2.KeyConcepts)
    
    domainSim := 0.0
    if sig1.Domain == sig2.Domain && sig1.Domain != "" {
        domainSim = 1.0
    }
    
    toolSim := overlapRatio(sig1.ToolSequence, sig2.ToolSequence)
    
    complexitySim := 1.0 - math.Abs(sig1.Complexity-sig2.Complexity)
    
    return (conceptSim * ws.ConceptWeight) +
           (domainSim * ws.DomainWeight) +
           (toolSim * ws.ToolWeight) +
           (complexitySim * ws.ComplexityWeight)
}

func jaccardSimilarity(a, b []string) float64 {
    if len(a) == 0 && len(b) == 0 {
        return 1.0
    }
    
    setA := make(map[string]bool)
    for _, item := range a {
        setA[item] = true
    }
    
    intersection := 0
    for _, item := range b {
        if setA[item] {
            intersection++
        }
    }
    
    union := len(setA)
    for _, item := range b {
        if !setA[item] {
            union++
        }
    }
    
    if union == 0 {
        return 0.0
    }
    
    return float64(intersection) / float64(union)
}

func overlapRatio(a, b []string) float64 {
    if len(a) == 0 || len(b) == 0 {
        return 0.0
    }
    
    setA := make(map[string]bool)
    for _, item := range a {
        setA[item] = true
    }
    
    overlap := 0
    for _, item := range b {
        if setA[item] {
            overlap++
        }
    }
    
    maxLen := math.Max(float64(len(a)), float64(len(b)))
    return float64(overlap) / maxLen
}
```

#### 2.3 Matcher

File: `server/internal/contextbridge/matcher.go`

```go
package contextbridge

import (
    "server/internal/storage"
    "sort"
)

type Match struct {
    TrajectoryID   string  `json:"trajectory_id"`
    SessionID      string  `json:"session_id"`
    Similarity     float64 `json:"similarity"`
    Summary        string  `json:"summary"`
    SuccessScore   float64 `json:"success_score"`
    QualityScore   float64 `json:"quality_score"`
}

type Matcher struct {
    storage    storage.Storage
    similarity SimilarityCalculator
    extractor  ConceptExtractor
}

func NewMatcher(storage storage.Storage, similarity SimilarityCalculator, extractor ConceptExtractor) *Matcher {
    return &Matcher{
        storage:    storage,
        similarity: similarity,
        extractor:  extractor,
    }
}

func (m *Matcher) FindMatches(sig *Signature, minSimilarity float64, maxMatches int) ([]*Match, error) {
    // Get candidates with signatures in single query (avoids N+1)
    prefix := ""
    if len(sig.Fingerprint) >= 8 {
        prefix = sig.Fingerprint[:8]
    }

    // Single query returns candidates WITH their signatures
    candidates, err := m.storage.FindCandidatesWithSignatures(sig.Domain, prefix, 50)
    if err != nil {
        return nil, err
    }

    // Calculate similarity for each candidate - no additional queries needed
    matches := make([]*Match, 0)
    for _, candidate := range candidates {
        if candidate.Signature == nil {
            continue
        }

        similarity := m.similarity.Calculate(sig, candidate.Signature)

        if similarity >= minSimilarity {
            matches = append(matches, &Match{
                TrajectoryID: candidate.TrajectoryID,
                SessionID:    candidate.SessionID,
                Similarity:   similarity,
                Summary:      candidate.Description,
                SuccessScore: candidate.SuccessScore,
                QualityScore: candidate.QualityScore,
            })
        }
    }

    // Sort by similarity desc
    sort.Slice(matches, func(i, j int) bool {
        return matches[i].Similarity > matches[j].Similarity
    })

    // Return top N
    if len(matches) > maxMatches {
        matches = matches[:maxMatches]
    }

    return matches, nil
}
```

**Deliverables**:
- Signature extraction with interface for swappable extractors
- Similarity calculation with configurable weights
- Matcher that ties it together
- Comprehensive unit tests with edge cases

**Success Criteria**:
- Signature extraction handles missing/malformed params
- Similarity scores align with manual judgment on test cases
- Matcher returns results sorted by relevance

### Phase 3: Integration (Est: 2-3 days, 8-12 hours)

#### 3.1 Configuration

File: `server/internal/contextbridge/config.go`

```go
package contextbridge

type Config struct {
    Enabled         bool
    MinSimilarity   float64
    MaxMatches      int
    EnabledTools    []string
    CacheSize       int
}

func DefaultConfig() *Config {
    return &Config{
        Enabled:       false,  // Feature flag - off by default
        MinSimilarity: 0.7,
        MaxMatches:    3,
        EnabledTools: []string{
            "think",
            "make-decision",
            "decompose-problem",
            "analyze-perspectives",
            "build-causal-graph",
        },
        CacheSize: 100,
    }
}
```

#### 3.2 Middleware

File: `server/internal/contextbridge/middleware.go`

```go
package contextbridge

import (
    "context"
    "log"
    "time"
)

type ContextBridge struct {
    config    *Config
    matcher   *Matcher
    extractor ConceptExtractor
    cache     *LRUCache  // Simple LRU for recent signatures
}

func New(config *Config, matcher *Matcher, extractor ConceptExtractor) *ContextBridge {
    return &ContextBridge{
        config:    config,
        matcher:   matcher,
        extractor: extractor,
        cache:     NewLRUCache(config.CacheSize),
    }
}

// EnrichResponse adds context matches to tool response
func (cb *ContextBridge) EnrichResponse(
    ctx context.Context,
    toolName string,
    params map[string]interface{},
    result interface{},
) (interface{}, error) {
    
    // Fast path - feature disabled or tool not enabled
    if !cb.config.Enabled || !cb.isEnabledTool(toolName) {
        return result, nil
    }
    
    start := time.Now()
    defer func() {
        elapsed := time.Since(start)
        if elapsed > 100*time.Millisecond {
            log.Printf("[WARN] Context enrichment took %v for tool %s", elapsed, toolName)
        }
    }()
    
    // Extract signature
    sig, err := ExtractSignature(toolName, params, cb.extractor)
    if err != nil || sig == nil {
        return result, nil // Don't fail on enrichment errors
    }
    
    // Check cache
    cacheKey := sig.Fingerprint
    if cached := cb.cache.Get(cacheKey); cached != nil {
        return cb.buildEnrichedResponse(result, cached.([]*Match)), nil
    }
    
    // Find matches
    matches, err := cb.matcher.FindMatches(sig, cb.config.MinSimilarity, cb.config.MaxMatches)
    if err != nil {
        log.Printf("[ERROR] Context matching failed: %v", err)
        return result, nil
    }
    
    // Cache result
    if len(matches) > 0 {
        cb.cache.Put(cacheKey, matches)
    }
    
    return cb.buildEnrichedResponse(result, matches), nil
}

func (cb *ContextBridge) buildEnrichedResponse(result interface{}, matches []*Match) interface{} {
    if len(matches) == 0 {
        return result
    }
    
    return map[string]interface{}{
        "result": result,
        "context_bridge": map[string]interface{}{
            "matches":        matches,
            "recommendation": cb.generateRecommendation(matches),
        },
    }
}

func (cb *ContextBridge) generateRecommendation(matches []*Match) string {
    if len(matches) == 0 {
        return ""
    }
    
    avgSuccess := 0.0
    for _, m := range matches {
        avgSuccess += m.SuccessScore
    }
    avgSuccess /= float64(len(matches))
    
    if avgSuccess > 0.8 {
        return "Similar past reasoning had high success rates."
    } else if avgSuccess < 0.4 {
        return "Similar past reasoning had low success rates - consider alternative approaches."
    }
    
    return "Related past sessions found."
}

func (cb *ContextBridge) isEnabledTool(toolName string) bool {
    for _, tool := range cb.config.EnabledTools {
        if tool == toolName {
            return true
        }
    }
    return false
}
```

#### 3.3 Wire into Server

File: `server/cmd/server/main.go`

```go
// Add to main()
contextBridgeConfig := contextbridge.DefaultConfig()
// Load from env or config file if present

extractor := contextbridge.NewSimpleExtractor()
similarity := contextbridge.NewDefaultSimilarity()
matcher := contextbridge.NewMatcher(storage, similarity, extractor)
bridge := contextbridge.New(contextBridgeConfig, matcher, extractor)

// Modify tool handlers
func handleThink(storage storage.Storage, bridge *contextbridge.ContextBridge) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse request
        var params map[string]interface{}
        json.NewDecoder(r.Body).Decode(&params)
        
        // Execute tool
        result, err := executeThink(storage, params)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Enrich with context
        enriched, _ := bridge.EnrichResponse(r.Context(), "think", params, result)
        
        json.NewEncoder(w).Encode(enriched)
    }
}
```

**Deliverables**:
- Configuration struct with feature flag
- Middleware with LRU cache
- Integration into existing tool handlers
- Environment variable support for config

**Success Criteria**:
- Feature flag defaults to disabled
- Enrichment adds <50ms latency (measured)
- Cache reduces repeated lookups
- Original response preserved if enrichment fails

### Phase 4: Testing (Est: 2 days, 6-8 hours)

#### 4.1 Unit Tests

```go
// Test signature extraction
func TestSignatureExtraction(t *testing.T) {
    cases := []struct{
        name string
        params map[string]interface{}
        wantConcepts []string
    }{
        {
            name: "think with content",
            params: map[string]interface{}{
                "content": "How to optimize database queries for performance",
                "mode": "linear",
            },
            wantConcepts: []string{"optimize", "database", "queries", "performance"},
        },
        {
            name: "missing content",
            params: map[string]interface{}{
                "mode": "linear",
            },
            wantConcepts: nil,
        },
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            sig, err := ExtractSignature("think", tc.params, NewSimpleExtractor())
            // ... assertions ...
        })
    }
}

// Test similarity calculation
func TestSimilarityCalculation(t *testing.T) {
    calc := NewDefaultSimilarity()
    
    sig1 := &Signature{
        KeyConcepts: []string{"database", "optimization", "query"},
        Domain: "engineering",
        Complexity: 0.6,
    }
    
    sig2Identical := &Signature{
        KeyConcepts: []string{"database", "optimization", "query"},
        Domain: "engineering",
        Complexity: 0.6,
    }
    
    sim := calc.Calculate(sig1, sig2Identical)
    if sim < 0.95 {
        t.Errorf("Expected high similarity for identical signatures, got %.2f", sim)
    }
    
    sig2Different := &Signature{
        KeyConcepts: []string{"machine", "learning", "model"},
        Domain: "ai",
        Complexity: 0.4,
    }
    
    sim = calc.Calculate(sig1, sig2Different)
    if sim > 0.3 {
        t.Errorf("Expected low similarity for different signatures, got %.2f", sim)
    }
}
```

#### 4.2 Integration Tests

```go
func TestContextBridgeEndToEnd(t *testing.T) {
    // Setup in-memory storage
    storage := setupTestStorage(t)
    defer storage.Close()
    
    // Create first trajectory with signature
    traj1 := createTestTrajectory(storage, "Database optimization", 0.9)
    sig1 := &Signature{
        Fingerprint: "abc123",
        Domain: "engineering",
        KeyConcepts: []string{"database", "optimization", "performance"},
        Complexity: 0.6,
    }
    storage.StoreContextSignature(traj1.ID, sig1)
    
    // Create bridge
    bridge := createTestBridge(storage)
    
    // Simulate similar request
    params := map[string]interface{}{
        "content": "Best practices for database performance optimization",
        "mode": "linear",
    }
    
    result := map[string]interface{}{
        "thought_id": "thought-test",
        "confidence": 0.8,
    }
    
    enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
    if err != nil {
        t.Fatalf("EnrichResponse failed: %v", err)
    }
    
    // Verify enrichment
    enrichedMap := enriched.(map[string]interface{})
    contextData := enrichedMap["context_bridge"].(map[string]interface{})
    matches := contextData["matches"].([]*Match)
    
    if len(matches) == 0 {
        t.Error("Expected matches for similar content")
    }
    
    if matches[0].Similarity < 0.7 {
        t.Errorf("Expected high similarity, got %.2f", matches[0].Similarity)
    }
}
```

#### 4.3 Performance Benchmarks

```go
func BenchmarkSignatureExtraction(b *testing.B) {
    params := map[string]interface{}{
        "content": "How to optimize database queries for large-scale applications with high throughput requirements",
    }
    extractor := NewSimpleExtractor()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ExtractSignature("think", params, extractor)
    }
}

func BenchmarkSimilarityCalculation(b *testing.B) {
    calc := NewDefaultSimilarity()
    sig1 := generateTestSignature()
    sig2 := generateTestSignature()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        calc.Calculate(sig1, sig2)
    }
}

func BenchmarkEndToEndMatching(b *testing.B) {
    storage := setupBenchStorage(b, 1000) // 1000 trajectories
    bridge := createTestBridge(storage)
    
    params := map[string]interface{}{
        "content": "Database optimization techniques",
    }
    result := map[string]interface{}{"thought_id": "test"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bridge.EnrichResponse(context.Background(), "think", params, result)
    }
}
```

**Deliverables**:
- Unit tests for all core functions
- Integration tests for full flow
- Performance benchmarks
- Test coverage report

**Success Criteria**:
- Unit test coverage >80%
- All integration tests pass
- Signature extraction <1ms per call
- End-to-end matching <50ms with 1000 trajectories

### Phase 5: Observability (Est: 1 day, 4 hours)

#### 5.1 Metrics Collection

File: `server/internal/contextbridge/metrics.go`

```go
package contextbridge

import "sync/atomic"

type Metrics struct {
    TotalEnrichments    int64
    MatchesFound        int64
    CacheHits           int64
    CacheMisses         int64
    TotalLatencyMs      int64
    MaxLatencyMs        int64
}

func (m *Metrics) RecordEnrichment(latencyMs int64, matchCount int) {
    atomic.AddInt64(&m.TotalEnrichments, 1)
    atomic.AddInt64(&m.MatchesFound, int64(matchCount))
    atomic.AddInt64(&m.TotalLatencyMs, latencyMs)
    
    // Update max latency
    for {
        current := atomic.LoadInt64(&m.MaxLatencyMs)
        if latencyMs <= current {
            break
        }
        if atomic.CompareAndSwapInt64(&m.MaxLatencyMs, current, latencyMs) {
            break
        }
    }
}

func (m *Metrics) RecordCacheHit() {
    atomic.AddInt64(&m.CacheHits, 1)
}

func (m *Metrics) RecordCacheMiss() {
    atomic.AddInt64(&m.CacheMisses, 1)
}

func (m *Metrics) Snapshot() map[string]interface{} {
    total := atomic.LoadInt64(&m.TotalEnrichments)
    avgLatency := int64(0)
    if total > 0 {
        avgLatency = atomic.LoadInt64(&m.TotalLatencyMs) / total
    }
    
    return map[string]interface{}{
        "total_enrichments": total,
        "total_matches":     atomic.LoadInt64(&m.MatchesFound),
        "cache_hits":        atomic.LoadInt64(&m.CacheHits),
        "cache_misses":      atomic.LoadInt64(&m.CacheMisses),
        "avg_latency_ms":    avgLatency,
        "max_latency_ms":    atomic.LoadInt64(&m.MaxLatencyMs),
    }
}
```

#### 5.2 Metrics Endpoint

Add to server:
```go
func handleMetrics(bridge *contextbridge.ContextBridge) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        metrics := bridge.GetMetrics()
        json.NewEncoder(w).Encode(metrics)
    }
}
```

**Deliverables**:
- Metrics collection in middleware
- Metrics endpoint
- Log warnings for high latency

**Success Criteria**:
- Metrics accurate and thread-safe
- Endpoint returns current stats
- Logs identify performance degradation

## Migration Strategy

### Backfill Existing Trajectories

Run after deployment:
```bash
# Backfill signatures for existing trajectories
curl -X POST http://localhost:8080/admin/backfill-signatures
```

Handles:
- Extract signatures from trajectory descriptions
- Store in context_signatures table
- Log failures for manual review
- Report success rate

### Gradual Rollout

1. Deploy with feature flag OFF
2. Run backfill on existing data
3. Enable for 10% of requests (random sampling)
4. Monitor metrics for 24 hours
5. If metrics acceptable, enable for 50%
6. If metrics acceptable, enable for 100%

## Rollback Plan

If performance degrades:
1. Set feature flag to false via environment variable
2. Restart server (immediate)
3. No code rollback needed

If accuracy is poor:
1. Adjust MinSimilarity threshold
2. Review failed matches in logs
3. Improve concept extraction if needed

## Open Questions

1. How to handle trajectory updates - invalidate signature cache?
2. Should similarity weights be per-domain or global?
3. What engagement metrics indicate context was useful?
4. Should we dedupe matches from same session?

## Testing Plan

### Manual Test Scenarios

1. Identical problem detection
   - Create trajectory about "database optimization"
   - Later use think with "database optimization"
   - Verify match with high similarity

2. Similar but not identical
   - Create trajectory about "SQL query performance"
   - Later use think with "database optimization"
   - Verify match with moderate similarity

3. Unrelated topics
   - Create trajectory about "frontend design"
   - Later use think with "database optimization"
   - Verify no match or very low similarity

4. Domain filtering
   - Create trajectories in different domains
   - Verify same-domain matches prioritized

5. Performance under load
   - Create 1000 trajectories
   - Measure latency of enrichment
   - Verify <100ms overhead

### Success Metrics

After 1 week of deployment:
- Enrichment latency p99 <100ms
- Match quality subjective review (sample 50 matches)
- Cache hit rate >50%
- No errors/crashes related to context bridge
- Feature flag rollback used: 0 times

## Future Work

**Completed in current implementation:**
- ~~Replace Jaccard similarity with embedding-based semantic similarity~~ (done - hybrid mode with Voyage AI)

**Remaining improvements:**
1. Replace SimpleExtractor with NLP-based concept extraction
2. Add vector database for efficient similarity search at scale (current SQLite approach may not scale beyond 10k signatures)
3. Track engagement (which matches Claude reads) to improve ranking
4. Add temporal decay (recent trajectories weighted higher)
5. Support cross-domain transfer learning
6. Implement backfill utility for existing trajectories
7. Add async embedding generation to reduce latency
8. Implement rate limiting for embedding API calls
