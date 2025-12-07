package contextbridge

import (
	"context"
	"fmt"
	"log"
	"time"

	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/types"
)

// ContextBridge enriches tool responses with similar past trajectories
type ContextBridge struct {
	config    *Config
	matcher   *Matcher
	extractor ConceptExtractor
	embedder  embeddings.Embedder
	cache     *LRUCache
	metrics   *Metrics
}

// New creates a new context bridge
func New(config *Config, matcher *Matcher, extractor ConceptExtractor, embedder embeddings.Embedder) *ContextBridge {
	return &ContextBridge{
		config:    config,
		matcher:   matcher,
		extractor: extractor,
		embedder:  embedder,
		cache:     NewLRUCache(config.CacheSize, config.CacheTTL),
		metrics:   &Metrics{},
	}
}

// HasEmbedder returns true if an embedder is configured
func (cb *ContextBridge) HasEmbedder() bool {
	return cb.embedder != nil
}

// GenerateEmbedding generates an embedding for text content
func (cb *ContextBridge) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if cb.embedder == nil {
		return nil, nil
	}
	return cb.embedder.Embed(ctx, text)
}

// EnrichResponse adds context matches to a tool response
func (cb *ContextBridge) EnrichResponse(
	ctx context.Context,
	toolName string,
	params types.Metadata,
	result interface{},
) (interface{}, error) {
	// Fast path - tool not in enabled list
	if !cb.isEnabledTool(toolName) {
		return result, nil
	}

	start := time.Now()

	// Check timeout
	deadline := start.Add(cb.config.Timeout)

	// Extract signature
	sig, err := ExtractSignature(toolName, params, cb.extractor)
	if err != nil {
		return nil, fmt.Errorf("signature extraction failed: %w", err)
	}
	if sig == nil {
		// No extractable content - return result as-is (not an error)
		return result, nil
	}

	// Generate embedding for semantic similarity if embedder is available
	// Use a separate context with timeout for embedding generation
	if cb.embedder != nil && len(sig.Embedding) == 0 {
		// Extract content for embedding (use full param content)
		content := extractTextContent(params)
		if content != "" {
			// Use a shorter timeout for embedding (500ms) to leave time for matching
			embedCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			embedding, embedErr := cb.embedder.Embed(embedCtx, content)
			cancel()

			if embedErr != nil {
				// FAIL FAST: Embedding is required, do not fallback
				return nil, fmt.Errorf("embedding generation failed: %w", embedErr)
			} else if len(embedding) > 0 {
				sig.Embedding = embedding
				log.Printf("[DEBUG] Generated embedding for signature (%d dimensions)", len(embedding))
			}
		}
	}

	// Check cache
	cacheKey := sig.Fingerprint
	if cached := cb.cache.Get(cacheKey); cached != nil {
		cb.metrics.RecordCacheHit()
		elapsed := time.Since(start).Milliseconds()
		cb.metrics.RecordEnrichment(elapsed, len(cached))
		return cb.buildEnrichedResponse(result, cached), nil
	}
	cb.metrics.RecordCacheMiss()

	// Check if we've exceeded timeout - FAIL FAST
	if time.Now().After(deadline) {
		cb.metrics.RecordTimeout()
		return nil, fmt.Errorf("context bridge timeout exceeded (%v) for tool %s", cb.config.Timeout, toolName)
	}

	// Find matches
	log.Printf("[DEBUG] Context bridge searching for matches (domain=%s, minSimilarity=%.2f, maxMatches=%d)",
		sig.Domain, cb.config.MinSimilarity, cb.config.MaxMatches)
	matches, err := cb.matcher.FindMatches(sig, cb.config.MinSimilarity, cb.config.MaxMatches)
	if err != nil {
		cb.metrics.RecordError()
		return nil, fmt.Errorf("context matching failed: %w", err)
	}
	log.Printf("[DEBUG] Context bridge found %d matches above threshold", len(matches))

	// Record metrics
	elapsed := time.Since(start).Milliseconds()
	cb.metrics.RecordEnrichment(elapsed, len(matches))

	if elapsed > 100 {
		log.Printf("[WARN] Context enrichment took %dms for tool %s", elapsed, toolName)
	}

	// Cache result
	if len(matches) > 0 {
		cb.cache.Put(cacheKey, matches)
	}

	return cb.buildEnrichedResponse(result, matches), nil
}

// buildEnrichedResponse creates the enriched response with context bridge data
func (cb *ContextBridge) buildEnrichedResponse(result interface{}, matches []*Match) interface{} {
	// Always return context_bridge structure for visibility (even with no matches)
	bridgeData := map[string]interface{}{
		"version":        "1.0",
		"matches":        matches,
		"match_count":    len(matches),
		"recommendation": cb.generateRecommendation(matches),
	}

	// Add similarity mode - always semantic_embedding since we fail fast if embedder fails
	if cb.embedder != nil {
		bridgeData["similarity_mode"] = "semantic_embedding"
	} else {
		bridgeData["similarity_mode"] = "concept_only"
	}

	// Add status field for clarity
	if len(matches) == 0 {
		bridgeData["status"] = "no_matches"
	} else {
		bridgeData["status"] = "matches_found"
	}

	return map[string]interface{}{
		"result":         result,
		"context_bridge": bridgeData,
	}
}

// generateRecommendation creates a recommendation based on match quality
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

// isEnabledTool checks if context bridging is enabled for the given tool
func (cb *ContextBridge) isEnabledTool(toolName string) bool {
	for _, tool := range cb.config.EnabledTools {
		if tool == toolName {
			return true
		}
	}
	return false
}

// GetMetrics returns the current metrics snapshot
func (cb *ContextBridge) GetMetrics() types.Metadata {
	metricsData := cb.metrics.Snapshot()
	metricsData["cache_stats"] = cb.cache.Stats()
	return metricsData
}

// extractTextContent extracts text content from params for embedding generation
func extractTextContent(params types.Metadata) string {
	// Check common content field names
	contentFields := []string{"content", "description", "query", "problem", "text", "message"}

	for _, field := range contentFields {
		if val, ok := params[field]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}

	return ""
}

// GetConfig returns the current configuration
func (cb *ContextBridge) GetConfig() *Config {
	return cb.config
}
