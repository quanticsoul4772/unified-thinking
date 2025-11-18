package contextbridge

import (
	"sort"
	"sync"
	"sync/atomic"
)

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

	// For percentile calculations
	latencyMu         sync.Mutex
	latencyBuffer     []int64 // Circular buffer of recent latencies
	latencyIndex      int     // Current position in buffer
	latencyBufferSize int     // Size of buffer (default 1000)
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

	// Track latency in circular buffer for percentile calculations
	m.latencyMu.Lock()
	if m.latencyBuffer == nil {
		m.latencyBufferSize = 1000
		m.latencyBuffer = make([]int64, m.latencyBufferSize)
	}
	m.latencyBuffer[m.latencyIndex] = latencyMs
	m.latencyIndex = (m.latencyIndex + 1) % m.latencyBufferSize
	m.latencyMu.Unlock()
}

// calculatePercentile calculates the p-th percentile from the latency buffer
func (m *Metrics) calculatePercentile(p float64) int64 {
	m.latencyMu.Lock()
	defer m.latencyMu.Unlock()

	if len(m.latencyBuffer) == 0 {
		return 0
	}

	// Get non-zero values (buffer may not be full yet)
	total := atomic.LoadInt64(&m.TotalEnrichments)
	count := int(total)
	if count > m.latencyBufferSize {
		count = m.latencyBufferSize
	}
	if count == 0 {
		return 0
	}

	// Copy and sort
	values := make([]int64, count)
	for i := 0; i < count; i++ {
		values[i] = m.latencyBuffer[i]
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	// Calculate percentile index
	index := int(float64(count-1) * p / 100.0)
	return values[index]
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
		"p50_latency_ms":    m.calculatePercentile(50),
		"p95_latency_ms":    m.calculatePercentile(95),
		"p99_latency_ms":    m.calculatePercentile(99),
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

	// Reset latency buffer
	m.latencyMu.Lock()
	m.latencyBuffer = nil
	m.latencyIndex = 0
	m.latencyMu.Unlock()
}
