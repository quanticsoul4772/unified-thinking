package testutil

import (
	"context"
	"errors"
	"testing"
)

func TestMockLLMClient_Generate(t *testing.T) {
	mock := NewMockLLMClient()
	ctx := context.Background()

	results, err := mock.Generate(ctx, "test prompt", 3)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if !mock.AssertGenerateCalled() {
		t.Error("Expected Generate to be called")
	}

	if !mock.AssertGenerateCalledWith(3) {
		t.Error("Expected Generate to be called with k=3")
	}
}

func TestMockLLMClient_GenerateWithError(t *testing.T) {
	expectedErr := errors.New("API error")
	mock := NewMockLLMClient().WithGenerateError(expectedErr)
	ctx := context.Background()

	_, err := mock.Generate(ctx, "test", 2)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockLLMClient_Score(t *testing.T) {
	mock := NewMockLLMClient().WithScoreResponses(0.9, 0.7, 0.5)
	ctx := context.Background()

	criteria := map[string]float64{"relevance": 0.5, "clarity": 0.5}

	// First call
	score1, _, err := mock.Score(ctx, "thought 1", "problem", criteria)
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}
	if score1 != 0.9 {
		t.Errorf("Expected score 0.9, got %f", score1)
	}

	// Second call
	score2, _, err := mock.Score(ctx, "thought 2", "problem", criteria)
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}
	if score2 != 0.7 {
		t.Errorf("Expected score 0.7, got %f", score2)
	}
}

func TestMockLLMClient_Aggregate(t *testing.T) {
	mock := NewMockLLMClient()
	ctx := context.Background()

	thoughts := []string{"thought 1", "thought 2", "thought 3"}
	result, err := mock.Aggregate(ctx, thoughts, "problem")
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	if len(mock.AggregateCalls) != 1 {
		t.Errorf("Expected 1 Aggregate call, got %d", len(mock.AggregateCalls))
	}
}

func TestMockLLMClient_Refine(t *testing.T) {
	mock := NewMockLLMClient()
	ctx := context.Background()

	result, err := mock.Refine(ctx, "original thought", "problem", 0)
	if err != nil {
		t.Fatalf("Refine failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestMockLLMClient_Reset(t *testing.T) {
	mock := NewMockLLMClient()
	ctx := context.Background()

	// Make some calls
	mock.Generate(ctx, "test", 2)
	mock.Aggregate(ctx, []string{"a"}, "b")
	mock.Score(ctx, "t", "p", nil)

	// Reset
	mock.Reset()

	// Verify all cleared
	counts := mock.GetCallCounts()
	for method, count := range counts {
		if count != 0 {
			t.Errorf("Expected %s call count to be 0 after reset, got %d", method, count)
		}
	}
}

func TestNewQuickMockLLM(t *testing.T) {
	mock := NewQuickMockLLM()
	ctx := context.Background()

	results, err := mock.Generate(ctx, "test", 2)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestNewDeterministicMockLLM(t *testing.T) {
	mock := NewDeterministicMockLLM()
	ctx := context.Background()

	// Scores should be deterministic
	score1, _, _ := mock.Score(ctx, "t", "p", nil)
	score2, _, _ := mock.Score(ctx, "t", "p", nil)
	score3, _, _ := mock.Score(ctx, "t", "p", nil)

	if score1 != 0.9 || score2 != 0.8 || score3 != 0.7 {
		t.Errorf("Unexpected scores: %f, %f, %f", score1, score2, score3)
	}
}

func TestNewFailingMockLLM(t *testing.T) {
	expectedErr := errors.New("mock failure")
	mock := NewFailingMockLLM(expectedErr)
	ctx := context.Background()

	_, err := mock.Generate(ctx, "test", 1)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Generate: expected error %v, got %v", expectedErr, err)
	}

	_, err = mock.Aggregate(ctx, nil, "")
	if !errors.Is(err, expectedErr) {
		t.Errorf("Aggregate: expected error %v, got %v", expectedErr, err)
	}

	_, err = mock.Refine(ctx, "", "", 0)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Refine: expected error %v, got %v", expectedErr, err)
	}

	_, _, err = mock.Score(ctx, "", "", nil)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Score: expected error %v, got %v", expectedErr, err)
	}
}

func TestResponseBuilder(t *testing.T) {
	mock := NewResponseBuilder().
		WithGenerate([]string{"a", "b"}, []string{"c", "d"}).
		WithScores(0.95, 0.85).
		Build()

	ctx := context.Background()

	// First generate call
	results1, _ := mock.Generate(ctx, "test", 2)
	if len(results1) != 2 || results1[0] != "a" {
		t.Errorf("Unexpected first generate results: %v", results1)
	}

	// Second generate call
	results2, _ := mock.Generate(ctx, "test", 2)
	if len(results2) != 2 || results2[0] != "c" {
		t.Errorf("Unexpected second generate results: %v", results2)
	}

	// First score
	score1, _, _ := mock.Score(ctx, "t", "p", nil)
	if score1 != 0.95 {
		t.Errorf("Expected score 0.95, got %f", score1)
	}
}

func TestMockLLMClient_CalculateNovelty(t *testing.T) {
	mock := NewMockLLMClient()
	ctx := context.Background()

	novelty, err := mock.CalculateNovelty(ctx, "new thought", []string{"sibling1", "sibling2"})
	if err != nil {
		t.Fatalf("CalculateNovelty failed: %v", err)
	}

	if novelty < 0 || novelty > 1 {
		t.Errorf("Novelty should be between 0 and 1, got %f", novelty)
	}

	if len(mock.NoveltyCalculations) != 1 {
		t.Errorf("Expected 1 novelty calculation, got %d", len(mock.NoveltyCalculations))
	}
}

func TestMockLLMClient_ExtractKeyPoints(t *testing.T) {
	mock := NewMockLLMClient()
	ctx := context.Background()

	keyPoints, err := mock.ExtractKeyPoints(ctx, "complex thought with many points")
	if err != nil {
		t.Fatalf("ExtractKeyPoints failed: %v", err)
	}

	if len(keyPoints) == 0 {
		t.Error("Expected at least one key point")
	}
}

func TestMockLLMClient_ResearchWithSearch(t *testing.T) {
	mock := NewMockLLMClient()
	ctx := context.Background()

	result, err := mock.ResearchWithSearch(ctx, "test query", "test problem")
	if err != nil {
		t.Fatalf("ResearchWithSearch failed: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}

	if len(mock.ResearchCalls) != 1 {
		t.Errorf("Expected 1 research call, got %d", len(mock.ResearchCalls))
	}
}

func TestMockLLMClient_ConcurrentAccess(t *testing.T) {
	mock := NewMockLLMClient()
	ctx := context.Background()

	// Run concurrent calls
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			mock.Generate(ctx, "test", 2)
			mock.Score(ctx, "t", "p", nil)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	counts := mock.GetCallCounts()
	if counts["Generate"] != 10 {
		t.Errorf("Expected 10 Generate calls, got %d", counts["Generate"])
	}
	if counts["Score"] != 10 {
		t.Errorf("Expected 10 Score calls, got %d", counts["Score"])
	}
}
