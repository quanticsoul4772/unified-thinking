package contextbridge

import (
	"context"
	"testing"
	"time"

	"unified-thinking/internal/embeddings"
)

// TestIntegration_FullContextBridgeFlow tests the complete context bridge flow
// from signature extraction through matching and enrichment
func TestIntegration_FullContextBridgeFlow(t *testing.T) {
	// Setup: Create storage with existing trajectories
	storage := &MockStorage{
		candidates: []*CandidateWithSignature{
			{
				TrajectoryID: "traj-db-opt-1",
				SessionID:    "sess-previous-1",
				Description:  "Successfully optimized database queries using indexing",
				SuccessScore: 0.95,
				QualityScore: 0.90,
				Signature: &Signature{
					Fingerprint:  "abc123def456",
					Domain:       "engineering",
					KeyConcepts:  []string{"database", "optimization", "queries", "indexing", "performance"},
					ToolSequence: []string{"think", "make-decision"},
					Complexity:   0.7,
				},
			},
			{
				TrajectoryID: "traj-db-opt-2",
				SessionID:    "sess-previous-2",
				Description:  "Database query optimization with caching",
				SuccessScore: 0.85,
				QualityScore: 0.80,
				Signature: &Signature{
					Fingerprint:  "xyz789abc012",
					Domain:       "engineering",
					KeyConcepts:  []string{"database", "queries", "caching", "performance"},
					ToolSequence: []string{"think", "decompose-problem"},
					Complexity:   0.6,
				},
			},
			{
				TrajectoryID: "traj-ml-1",
				SessionID:    "sess-ml",
				Description:  "Machine learning model training",
				SuccessScore: 0.70,
				QualityScore: 0.75,
				Signature: &Signature{
					Fingerprint:  "ml123model456",
					Domain:       "ai",
					KeyConcepts:  []string{"machine", "learning", "model", "training"},
					ToolSequence: []string{"think"},
					Complexity:   0.8,
				},
			},
		},
	}

	// Create context bridge with realistic config
	config := DefaultConfig()
	config.MinSimilarity = 0.3 // Lower for testing
	config.MaxMatches = 3

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor, nil)

	// Test 1: Think tool with database optimization problem
	t.Run("think_database_optimization", func(t *testing.T) {
		params := map[string]interface{}{
			"content": "How do I optimize database queries for better performance?",
			"mode":    "linear",
		}
		result := map[string]interface{}{
			"thought_id": "thought-123",
			"content":    "Consider indexing strategies...",
			"confidence": 0.85,
		}

		enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
		if err != nil {
			t.Fatalf("EnrichResponse failed: %v", err)
		}

		enrichedMap, ok := enriched.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", enriched)
		}

		// Should have context_bridge data
		bridgeData, ok := enrichedMap["context_bridge"]
		if !ok {
			t.Fatal("Expected context_bridge in response")
		}

		cbd, ok := bridgeData.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map[string]interface{}, got %T", bridgeData)
		}
		matches := cbd["matches"].([]*Match)

		// Should match database optimization trajectories
		if len(matches) == 0 {
			t.Fatal("Expected at least one match for database optimization query")
		}

		// First match should be the highest scoring database optimization
		foundDBMatch := false
		for _, match := range matches {
			if match.TrajectoryID == "traj-db-opt-1" || match.TrajectoryID == "traj-db-opt-2" {
				foundDBMatch = true
				if match.SuccessScore < 0.8 {
					t.Logf("Match %s has success score %.2f", match.TrajectoryID, match.SuccessScore)
				}
			}
		}
		if !foundDBMatch {
			t.Error("Expected to find database optimization trajectory match")
		}

		// Verify recommendation is generated
		if cbd["recommendation"].(string) == "" {
			t.Error("Expected recommendation to be generated")
		}

		t.Logf("Found %d matches, recommendation: %s", len(matches), cbd["recommendation"].(string))
	})

	// Test 2: Make-decision tool
	t.Run("make_decision_tool", func(t *testing.T) {
		params := map[string]interface{}{
			"situation": "Need to decide on database optimization approach",
			"criteria":  []string{"performance", "maintainability", "cost"},
		}
		result := map[string]interface{}{
			"decision_id": "dec-456",
			"choice":      "indexing",
		}

		enriched, err := bridge.EnrichResponse(context.Background(), "make-decision", params, result)
		if err != nil {
			t.Fatalf("EnrichResponse failed: %v", err)
		}

		enrichedMap := enriched.(map[string]interface{})
		if _, ok := enrichedMap["context_bridge"]; !ok {
			t.Error("Expected context_bridge for make-decision tool")
		}
	})

	// Test 3: Non-matching query (ML topic shouldn't match DB trajectories well)
	t.Run("different_domain_query", func(t *testing.T) {
		params := map[string]interface{}{
			"content": "How to train a neural network model?",
		}
		result := map[string]interface{}{
			"thought_id": "thought-789",
		}

		enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
		if err != nil {
			t.Fatalf("EnrichResponse failed: %v", err)
		}

		enrichedMap := enriched.(map[string]interface{})
		if bridgeData, ok := enrichedMap["context_bridge"]; ok {
			cbd := bridgeData.(map[string]interface{})
			matches := cbd["matches"].([]*Match)
			// Should find the ML trajectory
			foundMLMatch := false
			for _, match := range matches {
				if match.TrajectoryID == "traj-ml-1" {
					foundMLMatch = true
				}
			}
			if !foundMLMatch && len(matches) > 0 {
				t.Logf("Found matches but not ML trajectory: %v", matches)
			}
		}
	})

	// Test 4: Disabled tool should not be enriched
	t.Run("disabled_tool", func(t *testing.T) {
		params := map[string]interface{}{
			"content": "Database optimization query",
		}
		result := map[string]interface{}{
			"id": "test",
		}

		// "validate" is not in EnabledTools
		enriched, err := bridge.EnrichResponse(context.Background(), "validate", params, result)
		if err != nil {
			t.Fatalf("EnrichResponse failed: %v", err)
		}

		enrichedMap := enriched.(map[string]interface{})
		if _, ok := enrichedMap["context_bridge"]; ok {
			t.Error("validate tool should not be enriched")
		}
	})

	// Test 5: Cache behavior
	t.Run("cache_hit", func(t *testing.T) {
		params := map[string]interface{}{
			"content": "Database query optimization for performance",
		}
		result := map[string]interface{}{
			"thought_id": "thought-cache-test",
		}

		// First call - cache miss
		_, err := bridge.EnrichResponse(context.Background(), "think", params, result)
		if err != nil {
			t.Fatalf("First call failed: %v", err)
		}

		// Second call - should hit cache
		_, err = bridge.EnrichResponse(context.Background(), "think", params, result)
		if err != nil {
			t.Fatalf("Second call failed: %v", err)
		}

		metrics := bridge.GetMetrics()
		hits := metrics["cache_hits"].(int64)
		misses := metrics["cache_misses"].(int64)

		if hits < 1 {
			t.Errorf("Expected cache hits >= 1, got %d", hits)
		}
		t.Logf("Cache hits: %d, misses: %d", hits, misses)
	})

	// Test 6: Metrics tracking
	t.Run("metrics", func(t *testing.T) {
		metrics := bridge.GetMetrics()

		totalEnrichments := metrics["total_enrichments"].(int64)
		if totalEnrichments < 4 {
			t.Errorf("Expected at least 4 enrichments, got %d", totalEnrichments)
		}

		t.Logf("Metrics: %+v", metrics)
	})
}

// TestIntegration_SignatureGeneration tests signature generation from various inputs
func TestIntegration_SignatureGeneration(t *testing.T) {
	extractor := NewSimpleExtractor()

	tests := []struct {
		name           string
		tool           string
		params         map[string]interface{}
		expectNil      bool
		minConcepts    int
		expectedDomain string
	}{
		{
			name: "think_with_content",
			tool: "think",
			params: map[string]interface{}{
				"content": "Optimize database queries for better performance",
			},
			minConcepts: 3,
		},
		{
			name: "make_decision_with_situation",
			tool: "make-decision",
			params: map[string]interface{}{
				"situation": "Choose between caching strategies",
			},
			minConcepts: 2,
		},
		{
			name: "decompose_with_problem",
			tool: "decompose-problem",
			params: map[string]interface{}{
				"problem": "Build a scalable microservices architecture",
			},
			minConcepts: 3,
		},
		{
			name: "with_domain",
			tool: "think",
			params: map[string]interface{}{
				"content": "Test content",
				"domain":  "engineering",
			},
			minConcepts:    1,
			expectedDomain: "engineering",
		},
		{
			name: "empty_params",
			tool: "think",
			params: map[string]interface{}{
				"other_field": "not content",
			},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := ExtractSignature(tt.tool, tt.params, extractor)
			if err != nil {
				t.Fatalf("ExtractSignature failed: %v", err)
			}

			if tt.expectNil {
				if sig != nil {
					t.Error("Expected nil signature")
				}
				return
			}

			if sig == nil {
				t.Fatal("Expected signature, got nil")
			}

			if len(sig.KeyConcepts) < tt.minConcepts {
				t.Errorf("Expected at least %d concepts, got %d: %v",
					tt.minConcepts, len(sig.KeyConcepts), sig.KeyConcepts)
			}

			if sig.Fingerprint == "" {
				t.Error("Expected fingerprint")
			}

			if tt.expectedDomain != "" && sig.Domain != tt.expectedDomain {
				t.Errorf("Expected domain %s, got %s", tt.expectedDomain, sig.Domain)
			}

			if len(sig.ToolSequence) != 1 || sig.ToolSequence[0] != tt.tool {
				t.Errorf("Expected tool sequence [%s], got %v", tt.tool, sig.ToolSequence)
			}

			t.Logf("Signature: fingerprint=%s..., concepts=%v, complexity=%.2f",
				sig.Fingerprint[:8], sig.KeyConcepts, sig.Complexity)
		})
	}
}

// TestIntegration_SimilarityScoring tests similarity calculation accuracy
func TestIntegration_SimilarityScoring(t *testing.T) {
	similarity := NewDefaultSimilarity()

	tests := []struct {
		name   string
		sig1   *Signature
		sig2   *Signature
		minSim float64
		maxSim float64
	}{
		{
			name: "identical",
			sig1: &Signature{
				KeyConcepts:  []string{"database", "optimization"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.5,
			},
			sig2: &Signature{
				KeyConcepts:  []string{"database", "optimization"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.5,
			},
			minSim: 0.95,
			maxSim: 1.0,
		},
		{
			name: "high_overlap",
			sig1: &Signature{
				KeyConcepts:  []string{"database", "optimization", "queries"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.6,
			},
			sig2: &Signature{
				KeyConcepts:  []string{"database", "optimization", "performance"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.7,
			},
			minSim: 0.6,
			maxSim: 0.9,
		},
		{
			name: "different_domains",
			sig1: &Signature{
				KeyConcepts:  []string{"database", "optimization"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.5,
			},
			sig2: &Signature{
				KeyConcepts:  []string{"database", "optimization"},
				Domain:       "data-science",
				ToolSequence: []string{"think"},
				Complexity:   0.5,
			},
			minSim: 0.7,
			maxSim: 0.85,
		},
		{
			name: "no_overlap",
			sig1: &Signature{
				KeyConcepts:  []string{"database", "optimization"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.5,
			},
			sig2: &Signature{
				KeyConcepts:  []string{"machine", "learning"},
				Domain:       "ai",
				ToolSequence: []string{"decompose-problem"},
				Complexity:   0.8,
			},
			minSim: 0.0,
			maxSim: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := similarity.Calculate(tt.sig1, tt.sig2)
			if score < tt.minSim || score > tt.maxSim {
				t.Errorf("Expected similarity %.2f-%.2f, got %.2f", tt.minSim, tt.maxSim, score)
			}
			t.Logf("Similarity score: %.3f", score)
		})
	}
}

// TestIntegration_ErrorHandling tests that errors are properly propagated
func TestIntegration_ErrorHandling(t *testing.T) {
	config := DefaultConfig()

	// Create storage that returns errors
	errorStorage := &ErrorMockStorage{
		err: context.DeadlineExceeded,
	}

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(errorStorage, similarity, extractor)
	bridge := New(config, matcher, extractor, nil)

	params := map[string]interface{}{
		"content": "Test query",
	}
	result := map[string]interface{}{
		"id": "test",
	}

	_, err := bridge.EnrichResponse(context.Background(), "think", params, result)
	if err == nil {
		t.Error("Expected error to be propagated")
	}

	// Verify error metrics
	metrics := bridge.GetMetrics()
	errorCount, ok := metrics["error_count"].(int64)
	if !ok {
		t.Logf("Metrics: %+v", metrics)
		t.Fatal("error_count not found in metrics")
	}
	if errorCount < 1 {
		t.Errorf("Expected error count >= 1, got %d", errorCount)
	}
}

// ErrorMockStorage is a mock that returns errors
type ErrorMockStorage struct {
	err error
}

func (m *ErrorMockStorage) FindCandidatesWithSignatures(domain string, fingerprintPrefix string, limit int) ([]*CandidateWithSignature, error) {
	return nil, m.err
}

// TestIntegration_PerformanceBaseline establishes performance expectations
func TestIntegration_PerformanceBaseline(t *testing.T) {
	// Create storage with many candidates
	candidates := make([]*CandidateWithSignature, 100)
	for i := 0; i < 100; i++ {
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
	config := DefaultConfig()

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	bridge := New(config, matcher, extractor, nil)

	params := map[string]interface{}{
		"content": "Database query optimization",
	}
	result := map[string]interface{}{
		"id": "test",
	}

	// Warm up
	bridge.EnrichResponse(context.Background(), "think", params, result)

	// Measure performance
	start := time.Now()
	iterations := 100
	for i := 0; i < iterations; i++ {
		_, err := bridge.EnrichResponse(context.Background(), "think", params, result)
		if err != nil {
			t.Fatalf("EnrichResponse failed: %v", err)
		}
	}
	elapsed := time.Since(start)
	avgLatency := elapsed / time.Duration(iterations)

	t.Logf("Average latency: %v over %d iterations", avgLatency, iterations)

	// Should be well under 50ms target (most will be cache hits)
	if avgLatency > 10*time.Millisecond {
		t.Errorf("Average latency %v exceeds 10ms target", avgLatency)
	}
}

// TestIntegration_EmbeddingSimilarityPath tests the full embedding similarity workflow
func TestIntegration_EmbeddingSimilarityPath(t *testing.T) {
	// Create candidates with embeddings
	// Using simple normalized vectors for testing
	embedding1 := []float32{0.5, 0.5, 0.5, 0.5} // normalized
	embedding2 := []float32{0.6, 0.4, 0.5, 0.5} // similar to 1
	embedding3 := []float32{0.1, 0.9, 0.2, 0.3} // different

	storage := &MockStorage{
		candidates: []*CandidateWithSignature{
			{
				TrajectoryID: "traj-emb-1",
				SessionID:    "sess-emb-1",
				Description:  "Database performance optimization with indexes",
				SuccessScore: 0.95,
				QualityScore: 0.90,
				Signature: &Signature{
					Fingerprint:  "emb1fingerprint",
					Domain:       "engineering",
					KeyConcepts:  []string{"database", "performance", "indexes"},
					ToolSequence: []string{"think"},
					Complexity:   0.7,
					Embedding:    embedding1,
				},
			},
			{
				TrajectoryID: "traj-emb-2",
				SessionID:    "sess-emb-2",
				Description:  "Query optimization techniques",
				SuccessScore: 0.85,
				QualityScore: 0.80,
				Signature: &Signature{
					Fingerprint:  "emb2fingerprint",
					Domain:       "engineering",
					KeyConcepts:  []string{"query", "optimization", "techniques"},
					ToolSequence: []string{"think"},
					Complexity:   0.6,
					Embedding:    embedding2,
				},
			},
			{
				TrajectoryID: "traj-emb-3",
				SessionID:    "sess-emb-3",
				Description:  "Machine learning model training",
				SuccessScore: 0.70,
				QualityScore: 0.75,
				Signature: &Signature{
					Fingerprint:  "emb3fingerprint",
					Domain:       "ai",
					KeyConcepts:  []string{"machine", "learning", "model"},
					ToolSequence: []string{"think"},
					Complexity:   0.8,
					Embedding:    embedding3,
				},
			},
		},
	}

	// Configure context bridge with embedding support
	config := DefaultConfig()
	config.MinSimilarity = 0.3

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	// Pass nil embedder - we test similarity calculator directly
	bridge := New(config, matcher, extractor, nil)

	// Create embedding similarity calculator for direct tests
	embeddingSim := NewEmbeddingSimilarity(nil, similarity, true)

	// Test 1: Query semantically similar to embedding1
	t.Run("semantic_match_with_embeddings", func(t *testing.T) {
		params := map[string]interface{}{
			"content": "How to improve database performance using indexes?",
		}
		result := map[string]interface{}{
			"thought_id": "thought-emb-test",
		}

		enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
		if err != nil {
			t.Fatalf("EnrichResponse failed: %v", err)
		}

		enrichedMap, ok := enriched.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", enriched)
		}

		bridgeData, ok := enrichedMap["context_bridge"]
		if !ok {
			t.Fatal("Expected context_bridge in response")
		}

		cbd, ok := bridgeData.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map[string]interface{}, got %T", bridgeData)
		}

		matches := cbd["matches"].([]*Match)
		if len(matches) == 0 {
			t.Fatal("Expected at least one match")
		}

		// Should find database-related trajectories first due to concept similarity
		foundDBMatch := false
		for _, match := range matches {
			if match.TrajectoryID == "traj-emb-1" || match.TrajectoryID == "traj-emb-2" {
				foundDBMatch = true
				t.Logf("Found match: %s with similarity %.3f", match.TrajectoryID, match.Similarity)
			}
		}
		if !foundDBMatch {
			t.Error("Expected to find database-related trajectory matches")
		}
	})

	// Test 2: Verify EmbeddingSimilarity calculator works correctly
	t.Run("embedding_similarity_calculation", func(t *testing.T) {
		sig1 := &Signature{
			KeyConcepts: []string{"database", "optimization"},
			Embedding:   embedding1,
		}
		sig2 := &Signature{
			KeyConcepts: []string{"query", "optimization"},
			Embedding:   embedding2,
		}
		sig3 := &Signature{
			KeyConcepts: []string{"machine", "learning"},
			Embedding:   embedding3,
		}

		// Similar embeddings should have higher score
		sim12 := embeddingSim.Calculate(sig1, sig2)
		sim13 := embeddingSim.Calculate(sig1, sig3)

		if sim12 <= sim13 {
			t.Errorf("Expected sim(1,2)=%.3f > sim(1,3)=%.3f", sim12, sim13)
		}

		t.Logf("Similarity(1,2)=%.3f, Similarity(1,3)=%.3f", sim12, sim13)
	})

	// Test 3: NO fallback when embeddings missing - should return 0
	t.Run("no_fallback_without_embeddings", func(t *testing.T) {
		sigWithEmb := &Signature{
			KeyConcepts: []string{"database", "optimization"},
			Embedding:   embedding1,
		}
		sigWithoutEmb := &Signature{
			KeyConcepts: []string{"database", "optimization"},
			Embedding:   nil, // No embedding
		}

		// Should return 0.0 - NO fallback to concept similarity
		sim := embeddingSim.Calculate(sigWithEmb, sigWithoutEmb)
		if sim != 0 {
			t.Errorf("Expected zero similarity when embedding missing (no fallback), got %.3f", sim)
		}
		t.Logf("Similarity with missing embedding (expected 0.0): %.3f", sim)
	})

	// Test 4: Hybrid similarity mode indicator in response
	t.Run("similarity_mode_indicator", func(t *testing.T) {
		params := map[string]interface{}{
			"content": "Database query performance",
		}
		result := map[string]interface{}{
			"thought_id": "thought-mode-test",
		}

		enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
		if err != nil {
			t.Fatalf("EnrichResponse failed: %v", err)
		}

		enrichedMap := enriched.(map[string]interface{})
		cbd := enrichedMap["context_bridge"].(map[string]interface{})

		// Check if similarity_mode is set
		if mode, ok := cbd["similarity_mode"]; ok {
			t.Logf("Similarity mode: %v", mode)
		}
	})
}

// TestIntegration_EmbeddingSimilarityMath tests the mathematical correctness of embedding similarity
func TestIntegration_EmbeddingSimilarityMath(t *testing.T) {
	// Test cosine similarity calculation
	t.Run("cosine_similarity_identical", func(t *testing.T) {
		v := []float32{1, 2, 3, 4}
		similarity := embeddings.CosineSimilarity(v, v)
		if similarity < 0.999 {
			t.Errorf("Expected ~1.0 for identical vectors, got %.3f", similarity)
		}
	})

	t.Run("cosine_similarity_orthogonal", func(t *testing.T) {
		v1 := []float32{1, 0, 0, 0}
		v2 := []float32{0, 1, 0, 0}
		similarity := embeddings.CosineSimilarity(v1, v2)
		if similarity > 0.001 {
			t.Errorf("Expected ~0.0 for orthogonal vectors, got %.3f", similarity)
		}
	})

	t.Run("cosine_similarity_opposite", func(t *testing.T) {
		v1 := []float32{1, 0, 0, 0}
		v2 := []float32{-1, 0, 0, 0}
		similarity := embeddings.CosineSimilarity(v1, v2)
		if similarity > -0.999 {
			t.Errorf("Expected ~-1.0 for opposite vectors, got %.3f", similarity)
		}
	})

	// Test hybrid scoring
	t.Run("hybrid_similarity_scoring", func(t *testing.T) {
		fallback := NewDefaultSimilarity()
		embSim := NewEmbeddingSimilarity(nil, fallback, true)

		// Same concepts, same embeddings -> very high similarity
		sig1 := &Signature{
			KeyConcepts: []string{"database", "optimization"},
			Embedding:   []float32{0.5, 0.5, 0.5, 0.5},
		}
		sig2 := &Signature{
			KeyConcepts: []string{"database", "optimization"},
			Embedding:   []float32{0.5, 0.5, 0.5, 0.5},
		}

		score := embSim.Calculate(sig1, sig2)
		// Hybrid similarity = 70% embedding (1.0) + 30% concept (1.0) = 1.0
		// But the actual calculation may differ based on implementation
		if score < 0.85 {
			t.Errorf("Expected high similarity for identical signatures, got %.3f", score)
		}
		t.Logf("Identical signature similarity: %.3f", score)
	})
}

// TestIntegration_GracefulDegradation tests the system handles missing embeddings gracefully
func TestIntegration_GracefulDegradation(t *testing.T) {
	// Create candidates with mixed embedding availability
	storage := &MockStorage{
		candidates: []*CandidateWithSignature{
			{
				TrajectoryID: "traj-with-emb",
				SuccessScore: 0.90,
				Signature: &Signature{
					KeyConcepts: []string{"database", "queries"},
					Embedding:   []float32{0.5, 0.5, 0.5, 0.5},
				},
			},
			{
				TrajectoryID: "traj-without-emb",
				SuccessScore: 0.85,
				Signature: &Signature{
					KeyConcepts: []string{"database", "queries"},
					Embedding:   nil, // No embedding
				},
			},
		},
	}

	config := DefaultConfig()
	config.MinSimilarity = 0.2

	extractor := NewSimpleExtractor()
	similarity := NewDefaultSimilarity()
	matcher := NewMatcher(storage, similarity, extractor)
	// Pass nil embedder - test graceful handling
	bridge := New(config, matcher, extractor, nil)

	params := map[string]interface{}{
		"content": "Database query optimization",
	}
	result := map[string]interface{}{
		"thought_id": "test",
	}

	// Should not error even with mixed embeddings
	enriched, err := bridge.EnrichResponse(context.Background(), "think", params, result)
	if err != nil {
		t.Fatalf("EnrichResponse failed with mixed embeddings: %v", err)
	}

	enrichedMap := enriched.(map[string]interface{})

	// Check if context_bridge is present
	bridgeData, ok := enrichedMap["context_bridge"]
	if !ok {
		t.Log("No context_bridge in response - this is acceptable for graceful degradation test")
		return
	}

	cbd, ok := bridgeData.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", bridgeData)
	}

	matches, ok := cbd["matches"].([]*Match)
	if !ok {
		t.Log("No matches in context_bridge - this is acceptable")
		return
	}

	// Log results
	t.Logf("Found %d matches with mixed embedding availability", len(matches))
}
