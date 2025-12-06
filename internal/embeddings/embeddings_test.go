package embeddings

import (
	"context"
	"math"
	"os"
	"sync"
	"testing"
	"time"
)

// Test similarity functions

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		v1        []float32
		v2        []float32
		expected  float64
		tolerance float64
	}{
		{
			name:      "identical vectors",
			v1:        []float32{1, 0, 0},
			v2:        []float32{1, 0, 0},
			expected:  1.0,
			tolerance: 0.001,
		},
		{
			name:      "opposite vectors",
			v1:        []float32{1, 0, 0},
			v2:        []float32{-1, 0, 0},
			expected:  -1.0,
			tolerance: 0.001,
		},
		{
			name:      "orthogonal vectors",
			v1:        []float32{1, 0, 0},
			v2:        []float32{0, 1, 0},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "similar vectors",
			v1:        []float32{1, 1, 0},
			v2:        []float32{1, 0, 0},
			expected:  0.707,
			tolerance: 0.01,
		},
		{
			name:      "different lengths returns 0",
			v1:        []float32{1, 0},
			v2:        []float32{1, 0, 0},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "zero vector returns 0",
			v1:        []float32{0, 0, 0},
			v2:        []float32{1, 0, 0},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "empty vectors returns 0",
			v1:        []float32{},
			v2:        []float32{},
			expected:  0.0,
			tolerance: 0.001,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CosineSimilarity(tc.v1, tc.v2)
			if math.Abs(result-tc.expected) > tc.tolerance {
				t.Errorf("expected %.3f, got %.3f", tc.expected, result)
			}
		})
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name      string
		v1        []float32
		v2        []float32
		expected  float64
		tolerance float64
	}{
		{
			name:      "identical vectors",
			v1:        []float32{1, 0, 0},
			v2:        []float32{1, 0, 0},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "unit distance",
			v1:        []float32{0, 0, 0},
			v2:        []float32{1, 0, 0},
			expected:  1.0,
			tolerance: 0.001,
		},
		{
			name:      "pythagorean",
			v1:        []float32{0, 0},
			v2:        []float32{3, 4},
			expected:  5.0,
			tolerance: 0.001,
		},
		{
			name:      "different lengths returns max",
			v1:        []float32{1, 0},
			v2:        []float32{1, 0, 0},
			expected:  math.MaxFloat64,
			tolerance: 0.001,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := EuclideanDistance(tc.v1, tc.v2)
			if math.Abs(result-tc.expected) > tc.tolerance {
				t.Errorf("expected %.3f, got %.3f", tc.expected, result)
			}
		})
	}
}

func TestDotProduct(t *testing.T) {
	tests := []struct {
		name     string
		v1       []float32
		v2       []float32
		expected float64
	}{
		{
			name:     "simple dot product",
			v1:       []float32{1, 2, 3},
			v2:       []float32{4, 5, 6},
			expected: 32.0, // 1*4 + 2*5 + 3*6
		},
		{
			name:     "orthogonal vectors",
			v1:       []float32{1, 0},
			v2:       []float32{0, 1},
			expected: 0.0,
		},
		{
			name:     "different lengths returns 0",
			v1:       []float32{1, 2},
			v2:       []float32{1, 2, 3},
			expected: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := DotProduct(tc.v1, tc.v2)
			if result != tc.expected {
				t.Errorf("expected %.1f, got %.1f", tc.expected, result)
			}
		})
	}
}

func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name string
		v    []float32
	}{
		{
			name: "unit vector unchanged",
			v:    []float32{1, 0, 0},
		},
		{
			name: "normalizes to unit length",
			v:    []float32{3, 4, 0},
		},
		{
			name: "zero vector returns zero",
			v:    []float32{0, 0, 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeVector(tc.v)

			// Calculate norm of result
			var norm float64
			for _, val := range result {
				norm += float64(val * val)
			}
			norm = math.Sqrt(norm)

			// Should be either 0 or 1
			if norm > 0.001 && math.Abs(norm-1.0) > 0.001 {
				t.Errorf("expected unit norm, got %.3f", norm)
			}
		})
	}
}

// Test serialization functions

func TestSerializeDeserializeFloat32(t *testing.T) {
	tests := []struct {
		name string
		vec  []float32
	}{
		{
			name: "simple vector",
			vec:  []float32{1.0, 2.0, 3.0},
		},
		{
			name: "negative values",
			vec:  []float32{-1.5, 0.0, 1.5},
		},
		{
			name: "small values",
			vec:  []float32{0.001, 0.002, 0.003},
		},
		{
			name: "empty vector",
			vec:  []float32{},
		},
		{
			name: "single element",
			vec:  []float32{3.14159},
		},
		{
			name: "large vector",
			vec:  make([]float32, 512), // Typical embedding size
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Fill large vector with test data
			if tc.name == "large vector" {
				for i := range tc.vec {
					tc.vec[i] = float32(i) * 0.001
				}
			}

			bytes := SerializeFloat32(tc.vec)
			result := DeserializeFloat32(bytes)

			if len(result) != len(tc.vec) {
				t.Errorf("length mismatch: expected %d, got %d", len(tc.vec), len(result))
				return
			}

			for i := range tc.vec {
				if tc.vec[i] != result[i] {
					t.Errorf("value mismatch at index %d: expected %f, got %f", i, tc.vec[i], result[i])
				}
			}
		})
	}
}

func TestSerializeFloat32_Nil(t *testing.T) {
	result := SerializeFloat32(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

func TestDeserializeFloat32_Nil(t *testing.T) {
	result := DeserializeFloat32(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

// Test config

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if cfg.Provider != "voyage" {
		t.Errorf("expected Provider 'voyage', got '%s'", cfg.Provider)
	}
	if cfg.Model != "voyage-3-lite" {
		t.Errorf("expected Model 'voyage-3-lite', got '%s'", cfg.Model)
	}
	if !cfg.UseHybridSearch {
		t.Error("expected UseHybridSearch to be true by default")
	}
	if cfg.RRFParameter != 60 {
		t.Errorf("expected RRFParameter 60, got %d", cfg.RRFParameter)
	}
	if cfg.MinSimilarity != 0.5 {
		t.Errorf("expected MinSimilarity 0.5, got %f", cfg.MinSimilarity)
	}
	if !cfg.CacheEmbeddings {
		t.Error("expected CacheEmbeddings to be true by default")
	}
	if cfg.CacheTTL != 24*time.Hour {
		t.Errorf("expected CacheTTL 24h, got %v", cfg.CacheTTL)
	}
	if cfg.BatchSize != 100 {
		t.Errorf("expected BatchSize 100, got %d", cfg.BatchSize)
	}
	if cfg.MaxConcurrent != 5 {
		t.Errorf("expected MaxConcurrent 5, got %d", cfg.MaxConcurrent)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", cfg.Timeout)
	}
}

func TestConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("EMBEDDINGS_ENABLED", "true")
	os.Setenv("EMBEDDINGS_PROVIDER", "test-provider")
	os.Setenv("EMBEDDINGS_MODEL", "test-model")
	os.Setenv("VOYAGE_API_KEY", "test-key")
	os.Setenv("EMBEDDINGS_HYBRID_SEARCH", "true")
	os.Setenv("EMBEDDINGS_RRF_K", "100")
	os.Setenv("EMBEDDINGS_MIN_SIMILARITY", "0.75")
	os.Setenv("EMBEDDINGS_CACHE_ENABLED", "false")
	os.Setenv("EMBEDDINGS_CACHE_TTL", "1h")
	os.Setenv("EMBEDDINGS_BATCH_SIZE", "50")
	os.Setenv("EMBEDDINGS_MAX_CONCURRENT", "10")
	os.Setenv("EMBEDDINGS_TIMEOUT", "60s")

	defer func() {
		// Clean up
		os.Unsetenv("EMBEDDINGS_ENABLED")
		os.Unsetenv("EMBEDDINGS_PROVIDER")
		os.Unsetenv("EMBEDDINGS_MODEL")
		os.Unsetenv("VOYAGE_API_KEY")
		os.Unsetenv("EMBEDDINGS_HYBRID_SEARCH")
		os.Unsetenv("EMBEDDINGS_RRF_K")
		os.Unsetenv("EMBEDDINGS_MIN_SIMILARITY")
		os.Unsetenv("EMBEDDINGS_CACHE_ENABLED")
		os.Unsetenv("EMBEDDINGS_CACHE_TTL")
		os.Unsetenv("EMBEDDINGS_BATCH_SIZE")
		os.Unsetenv("EMBEDDINGS_MAX_CONCURRENT")
		os.Unsetenv("EMBEDDINGS_TIMEOUT")
	}()

	cfg := ConfigFromEnv()

	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.Provider != "test-provider" {
		t.Errorf("expected Provider 'test-provider', got '%s'", cfg.Provider)
	}
	if cfg.Model != "test-model" {
		t.Errorf("expected Model 'test-model', got '%s'", cfg.Model)
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("expected APIKey 'test-key', got '%s'", cfg.APIKey)
	}
	if cfg.RRFParameter != 100 {
		t.Errorf("expected RRFParameter 100, got %d", cfg.RRFParameter)
	}
	if cfg.MinSimilarity != 0.75 {
		t.Errorf("expected MinSimilarity 0.75, got %f", cfg.MinSimilarity)
	}
	if cfg.CacheEmbeddings {
		t.Error("expected CacheEmbeddings to be false")
	}
	if cfg.CacheTTL != time.Hour {
		t.Errorf("expected CacheTTL 1h, got %v", cfg.CacheTTL)
	}
	if cfg.BatchSize != 50 {
		t.Errorf("expected BatchSize 50, got %d", cfg.BatchSize)
	}
	if cfg.MaxConcurrent != 10 {
		t.Errorf("expected MaxConcurrent 10, got %d", cfg.MaxConcurrent)
	}
	if cfg.Timeout != 60*time.Second {
		t.Errorf("expected Timeout 60s, got %v", cfg.Timeout)
	}
}

// Benchmarks

func BenchmarkCosineSimilarity(b *testing.B) {
	v1 := make([]float32, 512)
	v2 := make([]float32, 512)
	for i := range v1 {
		v1[i] = float32(i) * 0.001
		v2[i] = float32(i) * 0.002
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CosineSimilarity(v1, v2)
	}
}

func BenchmarkSerializeFloat32(b *testing.B) {
	vec := make([]float32, 512)
	for i := range vec {
		vec[i] = float32(i) * 0.001
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SerializeFloat32(vec)
	}
}

func BenchmarkDeserializeFloat32(b *testing.B) {
	vec := make([]float32, 512)
	for i := range vec {
		vec[i] = float32(i) * 0.001
	}
	bytes := SerializeFloat32(vec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeserializeFloat32(bytes)
	}
}

// Test token bucket rate limiter

func TestTokenBucketLimiter_Basic(t *testing.T) {
	// Create limiter with 10 tokens/sec, burst of 5
	limiter := newTokenBucketLimiter(10, 5)

	ctx := context.Background()

	// Should be able to get 5 tokens immediately (burst)
	for i := 0; i < 5; i++ {
		err := limiter.Wait(ctx)
		if err != nil {
			t.Errorf("expected no error on token %d, got %v", i, err)
		}
	}
}

func TestTokenBucketLimiter_ContextCancellation(t *testing.T) {
	// Create limiter with very low rate
	limiter := newTokenBucketLimiter(1, 0)

	// Use a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := limiter.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestTokenBucketLimiter_Timeout(t *testing.T) {
	// Create limiter with very low rate
	limiter := newTokenBucketLimiter(1, 0)

	// Use a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestTokenBucketLimiter_Refill(t *testing.T) {
	// Create limiter with 100 tokens/sec, burst of 1
	limiter := newTokenBucketLimiter(100, 1)

	ctx := context.Background()

	// Use the initial token
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("expected no error on first token, got %v", err)
	}

	// Wait a bit for refill
	time.Sleep(20 * time.Millisecond)

	// Should be able to get another token after refill
	start := time.Now()
	err = limiter.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected no error after refill, got %v", err)
	}

	// Should have been nearly instant (< 20ms) since we waited for refill
	if elapsed > 20*time.Millisecond {
		t.Errorf("expected fast response after refill, took %v", elapsed)
	}
}

func TestTokenBucketLimiter_Concurrent(t *testing.T) {
	// Create limiter with high rate for concurrent test
	limiter := newTokenBucketLimiter(1000, 100)

	ctx := context.Background()
	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	// Launch 50 concurrent waiters
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := limiter.Wait(ctx)
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if errCount > 0 {
		t.Errorf("expected no errors, got %d", errCount)
	}
}

func TestNewVoyageEmbedder_HasRateLimiter(t *testing.T) {
	embedder := NewVoyageEmbedder("test-key", "voyage-3-lite")

	if embedder.rateLimiter == nil {
		t.Error("expected rate limiter to be initialized")
	}
}

func BenchmarkTokenBucketLimiter(b *testing.B) {
	limiter := newTokenBucketLimiter(10000, 1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Wait(ctx)
	}
}
