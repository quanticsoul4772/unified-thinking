package knowledge

import (
	"context"
	"testing"
)

// TestTruncate tests the truncate helper function
func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "needs truncation",
			input:  "hello world",
			maxLen: 5,
			want:   "hello...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "very long string",
			input:  "this is a very long string that needs to be truncated",
			maxLen: 20,
			want:   "this is a very long ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// TestMapExtractedType tests the entity type mapping function
func TestMapExtractedType(t *testing.T) {
	tests := []struct {
		name          string
		extractedType string
		want          EntityType
	}{
		{
			name:          "url type",
			extractedType: "url",
			want:          EntityTypeTool,
		},
		{
			name:          "file_path type",
			extractedType: "file_path",
			want:          EntityTypeTool,
		},
		{
			name:          "identifier type",
			extractedType: "identifier",
			want:          EntityTypeTool,
		},
		{
			name:          "email type",
			extractedType: "email",
			want:          EntityTypePerson,
		},
		{
			name:          "unknown type",
			extractedType: "unknown",
			want:          EntityTypeConcept,
		},
		{
			name:          "empty type",
			extractedType: "",
			want:          EntityTypeConcept,
		},
		{
			name:          "arbitrary type",
			extractedType: "foo",
			want:          EntityTypeConcept,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapExtractedType(tt.extractedType)
			if got != tt.want {
				t.Errorf("mapExtractedType(%q) = %v, want %v", tt.extractedType, got, tt.want)
			}
		})
	}
}

// TestMapRelationshipType tests the relationship type mapping function
func TestMapRelationshipType(t *testing.T) {
	tests := []struct {
		name          string
		extractedType string
		want          RelationshipType
	}{
		{
			name:          "CAUSES",
			extractedType: "CAUSES",
			want:          RelationshipCauses,
		},
		{
			name:          "ENABLES",
			extractedType: "ENABLES",
			want:          RelationshipEnables,
		},
		{
			name:          "CONTRADICTS",
			extractedType: "CONTRADICTS",
			want:          RelationshipContradicts,
		},
		{
			name:          "BUILDS_UPON",
			extractedType: "BUILDS_UPON",
			want:          RelationshipBuildsUpon,
		},
		{
			name:          "unknown type",
			extractedType: "UNKNOWN",
			want:          RelationshipRelatesTo,
		},
		{
			name:          "empty type",
			extractedType: "",
			want:          RelationshipRelatesTo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapRelationshipType(tt.extractedType)
			if got != tt.want {
				t.Errorf("mapRelationshipType(%q) = %v, want %v", tt.extractedType, got, tt.want)
			}
		})
	}
}

// TestKnowledgeGraph_DisabledPaths tests methods when KG is disabled
func TestKnowledgeGraph_DisabledPaths(t *testing.T) {
	// Create disabled knowledge graph
	cfg := KnowledgeGraphConfig{
		Enabled: false,
	}

	kg, err := NewKnowledgeGraph(cfg)
	if err != nil {
		t.Fatalf("NewKnowledgeGraph() error = %v", err)
	}

	t.Run("IsEnabled returns false", func(t *testing.T) {
		if kg.IsEnabled() {
			t.Error("IsEnabled() = true, want false")
		}
	})

	t.Run("Close succeeds when disabled", func(t *testing.T) {
		err := kg.Close(context.Background())
		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})

	t.Run("StoreEntity returns error when disabled", func(t *testing.T) {
		entity := &Entity{ID: "test", Label: "test", Type: EntityTypeConcept}
		err := kg.StoreEntity(context.Background(), entity, "content")
		if err == nil {
			t.Error("StoreEntity() error = nil, want error")
		}
	})

	t.Run("GetEntity returns error when disabled", func(t *testing.T) {
		_, err := kg.GetEntity(context.Background(), "test")
		if err == nil {
			t.Error("GetEntity() error = nil, want error")
		}
	})

	t.Run("SearchSemantic returns error when disabled", func(t *testing.T) {
		_, err := kg.SearchSemantic(context.Background(), "query", 10, 0.5)
		if err == nil {
			t.Error("SearchSemantic() error = nil, want error")
		}
	})

	t.Run("SearchGraph returns error when disabled", func(t *testing.T) {
		_, err := kg.SearchGraph(context.Background(), "entity", 2, nil)
		if err == nil {
			t.Error("SearchGraph() error = nil, want error")
		}
	})

	t.Run("HybridSearch returns error when disabled", func(t *testing.T) {
		_, err := kg.HybridSearch(context.Background(), "query", 10, 2)
		if err == nil {
			t.Error("HybridSearch() error = nil, want error")
		}
	})

	t.Run("HybridSearchWithThreshold returns error when disabled", func(t *testing.T) {
		_, err := kg.HybridSearchWithThreshold(context.Background(), "query", 10, 2, 0.5)
		if err == nil {
			t.Error("HybridSearchWithThreshold() error = nil, want error")
		}
	})

	t.Run("CreateRelationship returns error when disabled", func(t *testing.T) {
		rel := &Relationship{ID: "rel", FromID: "a", ToID: "b", Type: RelationshipRelatesTo}
		err := kg.CreateRelationship(context.Background(), rel)
		if err == nil {
			t.Error("CreateRelationship() error = nil, want error")
		}
	})

	t.Run("GetEmbeddingCacheStats returns error when disabled", func(t *testing.T) {
		_, err := kg.GetEmbeddingCacheStats()
		if err == nil {
			t.Error("GetEmbeddingCacheStats() error = nil, want error")
		}
	})

	t.Run("SetReranker succeeds even when disabled", func(t *testing.T) {
		kg.SetReranker(nil) // Should not panic
	})
}

// TestRLContextRetriever_DisabledPaths tests RLContextRetriever when KG is disabled
func TestRLContextRetriever_DisabledPaths(t *testing.T) {
	// Create disabled knowledge graph
	cfg := KnowledgeGraphConfig{
		Enabled: false,
	}

	kg, err := NewKnowledgeGraph(cfg)
	if err != nil {
		t.Fatalf("NewKnowledgeGraph() error = %v", err)
	}

	retriever := NewRLContextRetriever(kg)

	t.Run("GetSimilarProblems returns nil when disabled", func(t *testing.T) {
		problems, err := retriever.GetSimilarProblems(context.Background(), "test", 5)
		if err != nil {
			t.Errorf("GetSimilarProblems() error = %v, want nil", err)
		}
		if problems != nil {
			t.Errorf("GetSimilarProblems() = %v, want nil", problems)
		}
	})

	t.Run("GetStrategyPerformance returns nil when disabled", func(t *testing.T) {
		perf, err := retriever.GetStrategyPerformance(context.Background(), "test")
		if err != nil {
			t.Errorf("GetStrategyPerformance() error = %v, want nil", err)
		}
		if perf != nil {
			t.Errorf("GetStrategyPerformance() = %v, want nil", perf)
		}
	})

	t.Run("EnrichProblemContext returns original when disabled", func(t *testing.T) {
		enriched, err := retriever.EnrichProblemContext(context.Background(), "test problem")
		if err != nil {
			t.Errorf("EnrichProblemContext() error = %v, want nil", err)
		}
		if enriched != "test problem" {
			t.Errorf("EnrichProblemContext() = %q, want %q", enriched, "test problem")
		}
	})

	t.Run("RecordStrategyOutcome returns nil when disabled", func(t *testing.T) {
		err := retriever.RecordStrategyOutcome(context.Background(), "problem", "strategy", true, 0.9)
		if err != nil {
			t.Errorf("RecordStrategyOutcome() error = %v, want nil", err)
		}
	})
}

// TestRLContextRetriever_Construction tests retriever construction
func TestRLContextRetriever_Construction(t *testing.T) {
	cfg := KnowledgeGraphConfig{Enabled: false}
	kg, _ := NewKnowledgeGraph(cfg)

	t.Run("NewRLContextRetriever", func(t *testing.T) {
		retriever := NewRLContextRetriever(kg)
		if retriever == nil {
			t.Fatal("NewRLContextRetriever() = nil")
		}
		if retriever.similarityThreshold != 0.7 {
			t.Errorf("default threshold = %f, want 0.7", retriever.similarityThreshold)
		}
	})

	t.Run("NewRLContextRetrieverWithThreshold", func(t *testing.T) {
		retriever := NewRLContextRetrieverWithThreshold(kg, 0.5)
		if retriever == nil {
			t.Fatal("NewRLContextRetrieverWithThreshold() = nil")
		}
		if retriever.similarityThreshold != 0.5 {
			t.Errorf("threshold = %f, want 0.5", retriever.similarityThreshold)
		}
	})
}

// TestTrajectoryExtractor_DisabledPaths tests TrajectoryExtractor when KG is disabled
func TestTrajectoryExtractor_DisabledPaths(t *testing.T) {
	cfg := KnowledgeGraphConfig{Enabled: false}
	kg, _ := NewKnowledgeGraph(cfg)

	extractor := NewTrajectoryExtractor(kg, false)

	t.Run("NewTrajectoryExtractor succeeds", func(t *testing.T) {
		if extractor == nil {
			t.Error("NewTrajectoryExtractor() = nil")
		}
	})

	t.Run("ExtractFromTrajectory returns nil when disabled", func(t *testing.T) {
		err := extractor.ExtractFromTrajectory(context.Background(), "traj-1", "problem", []string{"step1"})
		if err != nil {
			t.Errorf("ExtractFromTrajectory() error = %v, want nil", err)
		}
	})
}
