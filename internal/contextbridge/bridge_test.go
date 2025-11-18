package contextbridge

import (
	"context"
	"testing"
	"time"
)

// MockStorage implements SignatureStorage for testing
type MockStorage struct {
	candidates []*CandidateWithSignature
}

func (m *MockStorage) FindCandidatesWithSignatures(domain string, fingerprintPrefix string, limit int) ([]*CandidateWithSignature, error) {
	return m.candidates, nil
}

func TestContextBridge_EnrichResponse_Disabled(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = false

	storage := &MockStorage{}
	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	result := map[string]interface{}{"thought_id": "test-123"}
	params := map[string]interface{}{"content": "test content"}

	enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
	if err != nil {
		t.Fatalf("EnrichResponse failed: %v", err)
	}

	// Should return original result when disabled
	enrichedMap, ok := enriched.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", enriched)
	}
	if enrichedMap["thought_id"] != "test-123" {
		t.Error("Expected original result when bridge is disabled")
	}
}

func TestContextBridge_EnrichResponse_NotEnabledTool(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true
	config.EnabledTools = []string{"think"} // Only think is enabled

	storage := &MockStorage{}
	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	result := map[string]interface{}{"thought_id": "test-123"}
	params := map[string]interface{}{"content": "test content"}

	// Call with non-enabled tool
	enriched, err := bridge.EnrichResponse(context.Background(), "validate", params, result)
	if err != nil {
		t.Fatalf("EnrichResponse failed: %v", err)
	}

	// Should return original result for non-enabled tool
	enrichedMap, ok := enriched.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", enriched)
	}
	if enrichedMap["thought_id"] != "test-123" {
		t.Error("Expected original result for non-enabled tool")
	}
}

func TestContextBridge_EnrichResponse_WithMatches(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.3 // Lower threshold for easier matching

	// Create mock storage with matching candidate
	// Use concepts that will definitely match the query
	storage := &MockStorage{
		candidates: []*CandidateWithSignature{
			{
				TrajectoryID: "traj-1",
				SessionID:    "sess-1",
				Description:  "Database optimization",
				SuccessScore: 0.9,
				QualityScore: 0.85,
				Signature: &Signature{
					Fingerprint:  "abc123",
					Domain:       "",
					KeyConcepts:  []string{"optimize", "database", "queries", "performance"},
					ToolSequence: []string{"think"},
					Complexity:   0.6,
				},
			},
		},
	}

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	result := map[string]interface{}{"thought_id": "test-123"}
	params := map[string]interface{}{
		"content": "optimize database queries performance",
	}

	enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
	if err != nil {
		t.Fatalf("EnrichResponse failed: %v", err)
	}

	// Should have enriched response
	enrichedMap, ok := enriched.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map response, got %T", enriched)
	}

	// Check for context_bridge key
	contextBridge, ok := enrichedMap["context_bridge"]
	if !ok {
		t.Fatal("Expected context_bridge in enriched response")
	}

	bridgeData, ok := contextBridge.(*ContextBridgeData)
	if !ok {
		t.Fatalf("Expected *ContextBridgeData, got %T", contextBridge)
	}

	if len(bridgeData.Matches) == 0 {
		t.Error("Expected at least one match")
	}

	if bridgeData.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", bridgeData.Version)
	}
}

func TestContextBridge_EnrichResponse_NoMatches(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.9 // High threshold

	// Create mock storage with non-matching candidate
	storage := &MockStorage{
		candidates: []*CandidateWithSignature{
			{
				TrajectoryID: "traj-1",
				Signature: &Signature{
					KeyConcepts:  []string{"completely", "different", "concepts"},
					Domain:       "other",
					ToolSequence: []string{"other-tool"},
					Complexity:   0.1,
				},
			},
		},
	}

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	result := map[string]interface{}{"thought_id": "test-123"}
	params := map[string]interface{}{
		"content": "How to optimize database queries for performance",
	}

	enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
	if err != nil {
		t.Fatalf("EnrichResponse failed: %v", err)
	}

	// Should return original result when no matches found
	enrichedMap, ok := enriched.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", enriched)
	}
	if enrichedMap["thought_id"] != "test-123" {
		t.Error("Expected original result when no matches found")
	}
}

func TestContextBridge_EnrichResponse_CacheHit(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.5

	storage := &MockStorage{
		candidates: []*CandidateWithSignature{
			{
				TrajectoryID: "traj-1",
				SessionID:    "sess-1",
				SuccessScore: 0.9,
				Signature: &Signature{
					KeyConcepts:  []string{"database", "optimization"},
					ToolSequence: []string{"think"},
					Complexity:   0.6,
				},
			},
		},
	}

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	result := map[string]interface{}{"thought_id": "test-123"}
	params := map[string]interface{}{
		"content": "database optimization",
	}

	// First call - cache miss
	_, err := bridge.EnrichResponse(context.Background(), "think", params, result)
	if err != nil {
		t.Fatalf("First EnrichResponse failed: %v", err)
	}

	// Second call with same params - should hit cache
	_, err = bridge.EnrichResponse(context.Background(), "think", params, result)
	if err != nil {
		t.Fatalf("Second EnrichResponse failed: %v", err)
	}

	// Check metrics
	metrics := bridge.GetMetrics()
	cacheHits := metrics["cache_hits"].(int64)
	cacheMisses := metrics["cache_misses"].(int64)

	if cacheHits < 1 {
		t.Errorf("Expected at least 1 cache hit, got %d", cacheHits)
	}

	if cacheMisses < 1 {
		t.Errorf("Expected at least 1 cache miss, got %d", cacheMisses)
	}
}

func TestContextBridge_GenerateRecommendation(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true

	storage := &MockStorage{}
	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	tests := []struct {
		name           string
		successScores  []float64
		wantContains   string
	}{
		{
			name:          "high success",
			successScores: []float64{0.9, 0.85, 0.95},
			wantContains:  "high success",
		},
		{
			name:          "low success",
			successScores: []float64{0.2, 0.3, 0.1},
			wantContains:  "low success",
		},
		{
			name:          "medium success",
			successScores: []float64{0.5, 0.6, 0.55},
			wantContains:  "Related past sessions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := make([]*Match, len(tt.successScores))
			for i, score := range tt.successScores {
				matches[i] = &Match{
					TrajectoryID: "traj",
					SuccessScore: score,
				}
			}

			rec := bridge.generateRecommendation(matches)
			if rec == "" {
				t.Error("Expected non-empty recommendation")
			}
		})
	}
}

func TestContextBridge_Metrics(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true

	storage := &MockStorage{
		candidates: []*CandidateWithSignature{
			{
				TrajectoryID: "traj-1",
				SuccessScore: 0.9,
				Signature: &Signature{
					KeyConcepts: []string{"test"},
				},
			},
		},
	}

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	result := map[string]interface{}{"id": "test"}
	params := map[string]interface{}{"content": "test content for metrics"}

	// Make a few calls
	for i := 0; i < 3; i++ {
		_, _ = bridge.EnrichResponse(context.Background(), "think", params, result)
	}

	metrics := bridge.GetMetrics()

	if metrics["total_enrichments"].(int64) != 3 {
		t.Errorf("Expected 3 total enrichments, got %v", metrics["total_enrichments"])
	}

	if metrics["enabled"].(bool) != true {
		t.Error("Expected enabled to be true")
	}

	// Should have cache stats
	cacheStats := metrics["cache_stats"].(map[string]int)
	if cacheStats["capacity"] != config.CacheSize {
		t.Errorf("Expected cache capacity %d, got %d", config.CacheSize, cacheStats["capacity"])
	}
}

func BenchmarkContextBridge_EnrichResponse(b *testing.B) {
	config := DefaultConfig()
	config.Enabled = true
	config.MinSimilarity = 0.5

	// Create mock storage with candidates
	candidates := make([]*CandidateWithSignature, 50)
	for i := 0; i < 50; i++ {
		candidates[i] = &CandidateWithSignature{
			TrajectoryID: "traj",
			SuccessScore: 0.8,
			Signature: &Signature{
				KeyConcepts:  []string{"database", "optimization", "query"},
				ToolSequence: []string{"think"},
				Complexity:   0.6,
			},
		}
	}

	storage := &MockStorage{candidates: candidates}
	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	result := map[string]interface{}{"thought_id": "test-123"}
	params := map[string]interface{}{
		"content": "How to optimize database queries for performance",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.EnrichResponse(context.Background(), "think", params, result)
	}
}

func BenchmarkSignatureExtraction(b *testing.B) {
	extractor := NewSimpleExtractor()
	params := map[string]interface{}{
		"content": "How to optimize database queries for large-scale applications with high throughput requirements",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractSignature("think", params, extractor)
	}
}

func TestContextBridge_IsEnabled(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true

	storage := &MockStorage{}
	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	if !bridge.IsEnabled() {
		t.Error("Expected IsEnabled() to return true")
	}

	config.Enabled = false
	bridge2 := New(config, matcher, extractor)
	if bridge2.IsEnabled() {
		t.Error("Expected IsEnabled() to return false")
	}
}

func TestContextBridge_GetConfig(t *testing.T) {
	config := DefaultConfig()
	config.MinSimilarity = 0.8
	config.MaxMatches = 5

	storage := &MockStorage{}
	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor)

	got := bridge.GetConfig()
	if got.MinSimilarity != 0.8 {
		t.Errorf("Expected MinSimilarity 0.8, got %v", got.MinSimilarity)
	}
	if got.MaxMatches != 5 {
		t.Errorf("Expected MaxMatches 5, got %v", got.MaxMatches)
	}
}

func TestMatcher_FindMatches(t *testing.T) {
	storage := &MockStorage{
		candidates: []*CandidateWithSignature{
			{
				TrajectoryID: "traj-1",
				SessionID:    "sess-1",
				Description:  "Database optimization",
				SuccessScore: 0.9,
				QualityScore: 0.85,
				Signature: &Signature{
					KeyConcepts:  []string{"database", "optimization", "performance"},
					Domain:       "engineering",
					ToolSequence: []string{"think"},
					Complexity:   0.6,
				},
			},
			{
				TrajectoryID: "traj-2",
				SessionID:    "sess-2",
				Description:  "Machine learning",
				SuccessScore: 0.7,
				QualityScore: 0.8,
				Signature: &Signature{
					KeyConcepts:  []string{"machine", "learning", "model"},
					Domain:       "ai",
					ToolSequence: []string{"think"},
					Complexity:   0.8,
				},
			},
		},
	}

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)

	querySig := &Signature{
		Fingerprint:  "abc123",
		KeyConcepts:  []string{"database", "optimization", "query"},
		Domain:       "engineering",
		ToolSequence: []string{"think"},
		Complexity:   0.6,
	}

	matches, err := matcher.FindMatches(querySig, 0.5, 3)
	if err != nil {
		t.Fatalf("FindMatches failed: %v", err)
	}

	if len(matches) == 0 {
		t.Error("Expected at least one match")
	}

	// First match should be the database optimization trajectory
	if matches[0].TrajectoryID != "traj-1" {
		t.Errorf("Expected traj-1 as first match, got %s", matches[0].TrajectoryID)
	}

	// Matches should be sorted by similarity
	for i := 1; i < len(matches); i++ {
		if matches[i].Similarity > matches[i-1].Similarity {
			t.Error("Matches not sorted by similarity (desc)")
		}
	}
}

func TestConfig_FromEnv(t *testing.T) {
	// Test with defaults
	config := DefaultConfig()

	if config.Enabled {
		t.Error("Expected Enabled to be false by default")
	}
	if config.MinSimilarity != 0.7 {
		t.Errorf("Expected MinSimilarity 0.7, got %v", config.MinSimilarity)
	}
	if config.MaxMatches != 3 {
		t.Errorf("Expected MaxMatches 3, got %v", config.MaxMatches)
	}
	if config.CacheSize != 100 {
		t.Errorf("Expected CacheSize 100, got %v", config.CacheSize)
	}
	if config.Timeout != 100*time.Millisecond {
		t.Errorf("Expected Timeout 100ms, got %v", config.Timeout)
	}
}
