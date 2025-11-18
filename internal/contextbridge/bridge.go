package contextbridge

import (
	"context"
	"log"
	"time"
)

// ContextBridge enriches tool responses with similar past trajectories
type ContextBridge struct {
	config    *Config
	matcher   *Matcher
	extractor ConceptExtractor
	cache     *LRUCache
	metrics   *Metrics
}

// New creates a new context bridge
func New(config *Config, matcher *Matcher, extractor ConceptExtractor) *ContextBridge {
	return &ContextBridge{
		config:    config,
		matcher:   matcher,
		extractor: extractor,
		cache:     NewLRUCache(config.CacheSize, config.CacheTTL),
		metrics:   &Metrics{},
	}
}

// EnrichResponse adds context matches to a tool response
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

	// Check timeout
	deadline := start.Add(cb.config.Timeout)

	// Extract signature
	sig, err := ExtractSignature(toolName, params, cb.extractor)
	if err != nil || sig == nil {
		return result, nil // Don't fail on enrichment errors
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

	// Check if we've exceeded timeout
	if time.Now().After(deadline) {
		cb.metrics.RecordTimeout()
		log.Printf("[WARN] Context bridge timeout before matching for tool %s", toolName)
		return result, nil
	}

	// Find matches
	matches, err := cb.matcher.FindMatches(sig, cb.config.MinSimilarity, cb.config.MaxMatches)
	if err != nil {
		cb.metrics.RecordError()
		log.Printf("[ERROR] Context matching failed: %v", err)
		return result, nil
	}

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
	if len(matches) == 0 {
		return result
	}

	return map[string]interface{}{
		"result": result,
		"context_bridge": &ContextBridgeData{
			Version:        "1.0",
			Matches:        matches,
			Recommendation: cb.generateRecommendation(matches),
		},
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
func (cb *ContextBridge) GetMetrics() map[string]interface{} {
	metricsData := cb.metrics.Snapshot()
	metricsData["cache_stats"] = cb.cache.Stats()
	metricsData["enabled"] = cb.config.Enabled
	return metricsData
}

// IsEnabled returns whether the context bridge is enabled
func (cb *ContextBridge) IsEnabled() bool {
	return cb.config.Enabled
}

// GetConfig returns the current configuration
func (cb *ContextBridge) GetConfig() *Config {
	return cb.config
}
