package contextbridge

import "sync/atomic"

// Metrics tracks context bridge performance and usage
type Metrics struct {
	TotalEnrichments int64
	MatchesFound     int64
	CacheHits        int64
	CacheMisses      int64
	TotalLatencyMs   int64
	MaxLatencyMs     int64
	ErrorCount       int64
	TimeoutCount     int64
}

// RecordEnrichment records a completed enrichment operation
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

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit() {
	atomic.AddInt64(&m.CacheHits, 1)
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss() {
	atomic.AddInt64(&m.CacheMisses, 1)
}

// RecordError records an error
func (m *Metrics) RecordError() {
	atomic.AddInt64(&m.ErrorCount, 1)
}

// RecordTimeout records a timeout
func (m *Metrics) RecordTimeout() {
	atomic.AddInt64(&m.TimeoutCount, 1)
}

// Snapshot returns current metrics as a map
func (m *Metrics) Snapshot() map[string]interface{} {
	total := atomic.LoadInt64(&m.TotalEnrichments)
	avgLatency := int64(0)
	if total > 0 {
		avgLatency = atomic.LoadInt64(&m.TotalLatencyMs) / total
	}

	cacheHits := atomic.LoadInt64(&m.CacheHits)
	cacheMisses := atomic.LoadInt64(&m.CacheMisses)
	cacheTotal := cacheHits + cacheMisses
	cacheHitRate := 0.0
	if cacheTotal > 0 {
		cacheHitRate = float64(cacheHits) / float64(cacheTotal)
	}

	return map[string]interface{}{
		"total_enrichments": total,
		"total_matches":     atomic.LoadInt64(&m.MatchesFound),
		"cache_hits":        cacheHits,
		"cache_misses":      cacheMisses,
		"cache_hit_rate":    cacheHitRate,
		"avg_latency_ms":    avgLatency,
		"max_latency_ms":    atomic.LoadInt64(&m.MaxLatencyMs),
		"error_count":       atomic.LoadInt64(&m.ErrorCount),
		"timeout_count":     atomic.LoadInt64(&m.TimeoutCount),
	}
}

// Reset resets all metrics to zero
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.TotalEnrichments, 0)
	atomic.StoreInt64(&m.MatchesFound, 0)
	atomic.StoreInt64(&m.CacheHits, 0)
	atomic.StoreInt64(&m.CacheMisses, 0)
	atomic.StoreInt64(&m.TotalLatencyMs, 0)
	atomic.StoreInt64(&m.MaxLatencyMs, 0)
	atomic.StoreInt64(&m.ErrorCount, 0)
	atomic.StoreInt64(&m.TimeoutCount, 0)
}
