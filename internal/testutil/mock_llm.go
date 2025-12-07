// Package testutil provides testing utilities for unified-thinking
package testutil

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"unified-thinking/internal/modes"
)

// MockLLMClient provides a configurable mock implementation of LLMClient
// for testing GoT operations without API calls.
type MockLLMClient struct {
	mu sync.Mutex

	// Configurable responses
	GenerateResponses  [][]string              // Responses for Generate calls (indexed by call count)
	AggregateResponses []string                // Responses for Aggregate calls
	RefineResponses    []string                // Responses for Refine calls
	ScoreResponses     []float64               // Responses for Score calls
	KeyPointResponses  [][]string              // Responses for ExtractKeyPoints calls
	NoveltyResponses   []float64               // Responses for CalculateNovelty calls
	ResearchResponses  []*modes.ResearchResult // Responses for ResearchWithSearch calls

	// Error injection
	GenerateError  error
	AggregateError error
	RefineError    error
	ScoreError     error
	KeyPointError  error
	NoveltyError   error
	ResearchError  error

	// Call tracking
	GenerateCalls         []GenerateCall
	AggregateCalls        []AggregateCall
	RefineCalls           []RefineCall
	ScoreCalls            []ScoreCall
	ExtractKeyPointsCalls []ExtractKeyPointsCall
	NoveltyCalculations   []NoveltyCall
	ResearchCalls         []ResearchCall

	// Response indices
	generateIdx  int
	aggregateIdx int
	refineIdx    int
	scoreIdx     int
	keyPointIdx  int
	noveltyIdx   int
	researchIdx  int
}

// Call tracking types
type GenerateCall struct {
	Prompt string
	K      int
}

type AggregateCall struct {
	Thoughts []string
	Problem  string
}

type RefineCall struct {
	Thought         string
	Problem         string
	RefinementCount int
}

type ScoreCall struct {
	Thought  string
	Problem  string
	Criteria map[string]float64
}

type ExtractKeyPointsCall struct {
	Thought string
}

type NoveltyCall struct {
	Thought  string
	Siblings []string
}

type ResearchCall struct {
	Query   string
	Problem string
}

// NewMockLLMClient creates a new mock with sensible defaults
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		GenerateResponses: [][]string{
			{"Approach 1: Consider the problem systematically", "Approach 2: Break it into smaller parts", "Approach 3: Look for analogies"},
		},
		AggregateResponses: []string{"Synthesized insight combining all approaches"},
		RefineResponses:    []string{"Refined and improved thought"},
		ScoreResponses:     []float64{0.75},
		KeyPointResponses:  [][]string{{"Key point 1", "Key point 2"}},
		NoveltyResponses:   []float64{0.6},
		ResearchResponses:  []*modes.ResearchResult{},
	}
}

// Generate creates k diverse continuations
func (m *MockLLMClient) Generate(ctx context.Context, prompt string, k int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GenerateCalls = append(m.GenerateCalls, GenerateCall{Prompt: prompt, K: k})

	if m.GenerateError != nil {
		return nil, m.GenerateError
	}

	if m.generateIdx >= len(m.GenerateResponses) {
		// Generate default responses if we run out
		responses := make([]string, k)
		for i := 0; i < k; i++ {
			responses[i] = fmt.Sprintf("Generated thought %d for: %s", i+1, truncate(prompt, 50))
		}
		return responses, nil
	}

	responses := m.GenerateResponses[m.generateIdx]
	m.generateIdx++

	// Adjust to requested k
	if len(responses) > k {
		return responses[:k], nil
	}
	return responses, nil
}

// Aggregate synthesizes multiple thoughts
func (m *MockLLMClient) Aggregate(ctx context.Context, thoughts []string, problem string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.AggregateCalls = append(m.AggregateCalls, AggregateCall{Thoughts: thoughts, Problem: problem})

	if m.AggregateError != nil {
		return "", m.AggregateError
	}

	if m.aggregateIdx >= len(m.AggregateResponses) {
		return fmt.Sprintf("Aggregated %d thoughts about: %s", len(thoughts), truncate(problem, 50)), nil
	}

	response := m.AggregateResponses[m.aggregateIdx]
	m.aggregateIdx++
	return response, nil
}

// Refine improves a thought
func (m *MockLLMClient) Refine(ctx context.Context, thought string, problem string, refinementCount int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RefineCalls = append(m.RefineCalls, RefineCall{Thought: thought, Problem: problem, RefinementCount: refinementCount})

	if m.RefineError != nil {
		return "", m.RefineError
	}

	if m.refineIdx >= len(m.RefineResponses) {
		return fmt.Sprintf("Refined (iteration %d): %s", refinementCount+1, truncate(thought, 50)), nil
	}

	response := m.RefineResponses[m.refineIdx]
	m.refineIdx++
	return response, nil
}

// Score evaluates thought quality
func (m *MockLLMClient) Score(ctx context.Context, thought string, problem string, criteria map[string]float64) (float64, map[string]float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ScoreCalls = append(m.ScoreCalls, ScoreCall{Thought: thought, Problem: problem, Criteria: criteria})

	if m.ScoreError != nil {
		return 0, nil, m.ScoreError
	}

	var score float64
	if m.scoreIdx >= len(m.ScoreResponses) {
		score = 0.7 // Default score
	} else {
		score = m.ScoreResponses[m.scoreIdx]
		m.scoreIdx++
	}

	// Generate per-criterion scores
	criteriaScores := make(map[string]float64)
	for name := range criteria {
		criteriaScores[name] = score
	}

	return score, criteriaScores, nil
}

// ExtractKeyPoints identifies key insights
func (m *MockLLMClient) ExtractKeyPoints(ctx context.Context, thought string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ExtractKeyPointsCalls = append(m.ExtractKeyPointsCalls, ExtractKeyPointsCall{Thought: thought})

	if m.KeyPointError != nil {
		return nil, m.KeyPointError
	}

	if m.keyPointIdx >= len(m.KeyPointResponses) {
		return []string{"Key insight from: " + truncate(thought, 30)}, nil
	}

	response := m.KeyPointResponses[m.keyPointIdx]
	m.keyPointIdx++
	return response, nil
}

// CalculateNovelty measures uniqueness
func (m *MockLLMClient) CalculateNovelty(ctx context.Context, thought string, siblings []string) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NoveltyCalculations = append(m.NoveltyCalculations, NoveltyCall{Thought: thought, Siblings: siblings})

	if m.NoveltyError != nil {
		return 0, m.NoveltyError
	}

	if m.noveltyIdx >= len(m.NoveltyResponses) {
		return 0.5, nil
	}

	response := m.NoveltyResponses[m.noveltyIdx]
	m.noveltyIdx++
	return response, nil
}

// ResearchWithSearch performs web-augmented research
func (m *MockLLMClient) ResearchWithSearch(ctx context.Context, query string, problem string) (*modes.ResearchResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ResearchCalls = append(m.ResearchCalls, ResearchCall{Query: query, Problem: problem})

	if m.ResearchError != nil {
		return nil, m.ResearchError
	}

	if m.researchIdx >= len(m.ResearchResponses) {
		return &modes.ResearchResult{
			Findings:    "Mock research result for: " + query,
			KeyInsights: []string{"Key insight 1", "Key insight 2"},
			Confidence:  0.8,
			Citations:   []modes.Citation{{Title: "Example Source", URL: "https://example.com"}},
			Searches:    1,
		}, nil
	}

	response := m.ResearchResponses[m.researchIdx]
	m.researchIdx++
	return response, nil
}

// Reset clears all call history and resets indices
func (m *MockLLMClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GenerateCalls = nil
	m.AggregateCalls = nil
	m.RefineCalls = nil
	m.ScoreCalls = nil
	m.ExtractKeyPointsCalls = nil
	m.NoveltyCalculations = nil
	m.ResearchCalls = nil

	m.generateIdx = 0
	m.aggregateIdx = 0
	m.refineIdx = 0
	m.scoreIdx = 0
	m.keyPointIdx = 0
	m.noveltyIdx = 0
	m.researchIdx = 0
}

// WithGenerateError configures Generate to return an error
func (m *MockLLMClient) WithGenerateError(err error) *MockLLMClient {
	m.GenerateError = err
	return m
}

// WithScoreResponses configures Score responses
func (m *MockLLMClient) WithScoreResponses(scores ...float64) *MockLLMClient {
	m.ScoreResponses = scores
	return m
}

// WithGenerateResponses configures Generate responses
func (m *MockLLMClient) WithGenerateResponses(responses ...[]string) *MockLLMClient {
	m.GenerateResponses = responses
	return m
}

// Assertions for tests

// AssertGenerateCalled checks that Generate was called
func (m *MockLLMClient) AssertGenerateCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.GenerateCalls) > 0
}

// AssertGenerateCalledWith checks that Generate was called with specific k
func (m *MockLLMClient) AssertGenerateCalledWith(k int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.GenerateCalls {
		if call.K == k {
			return true
		}
	}
	return false
}

// GetCallCounts returns call counts for all methods
func (m *MockLLMClient) GetCallCounts() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]int{
		"Generate":           len(m.GenerateCalls),
		"Aggregate":          len(m.AggregateCalls),
		"Refine":             len(m.RefineCalls),
		"Score":              len(m.ScoreCalls),
		"ExtractKeyPoints":   len(m.ExtractKeyPointsCalls),
		"CalculateNovelty":   len(m.NoveltyCalculations),
		"ResearchWithSearch": len(m.ResearchCalls),
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Verify interface compliance
var _ modes.LLMClient = (*MockLLMClient)(nil)

// Predefined mock configurations for common test scenarios

// NewQuickMockLLM returns a mock optimized for fast tests
func NewQuickMockLLM() *MockLLMClient {
	return &MockLLMClient{
		GenerateResponses:  [][]string{{"Quick thought 1", "Quick thought 2"}},
		AggregateResponses: []string{"Quick aggregate"},
		RefineResponses:    []string{"Quick refine"},
		ScoreResponses:     []float64{0.8},
		KeyPointResponses:  [][]string{{"Quick point"}},
		NoveltyResponses:   []float64{0.7},
	}
}

// NewDeterministicMockLLM returns a mock with predictable outputs
func NewDeterministicMockLLM() *MockLLMClient {
	mock := NewMockLLMClient()
	mock.ScoreResponses = []float64{0.9, 0.8, 0.7, 0.6, 0.5}
	mock.NoveltyResponses = []float64{1.0, 0.8, 0.6, 0.4, 0.2}
	return mock
}

// NewFailingMockLLM returns a mock that fails on all operations
func NewFailingMockLLM(err error) *MockLLMClient {
	return &MockLLMClient{
		GenerateError:  err,
		AggregateError: err,
		RefineError:    err,
		ScoreError:     err,
		KeyPointError:  err,
		NoveltyError:   err,
		ResearchError:  err,
	}
}

// NewRealisticMockLLM returns a mock with realistic, varied responses
func NewRealisticMockLLM() *MockLLMClient {
	return &MockLLMClient{
		GenerateResponses: [][]string{
			{
				"We could approach this by decomposing the problem into smaller, manageable components.",
				"An alternative perspective is to consider the historical context and patterns.",
				"Perhaps we should examine the edge cases and failure modes first.",
			},
			{
				"Building on the previous analysis, we can identify three key factors.",
				"The root cause appears to be related to the system architecture.",
				"A more creative solution would involve rethinking our assumptions.",
			},
		},
		AggregateResponses: []string{
			"Synthesizing these perspectives, the core insight is that a multi-faceted approach yields better results than any single strategy.",
			"The combined analysis reveals that the problem has both technical and organizational dimensions.",
		},
		RefineResponses: []string{
			"Upon reflection, the initial thought can be strengthened by considering counter-examples and addressing potential objections.",
			"Refinement reveals additional nuances that make the solution more robust.",
		},
		ScoreResponses: []float64{0.85, 0.78, 0.92, 0.65, 0.88},
		KeyPointResponses: [][]string{
			{"Decomposition is key", "Historical patterns matter", "Edge cases reveal insights"},
			{"Multi-faceted approach preferred", "Technical and organizational factors"},
		},
		NoveltyResponses: []float64{0.9, 0.6, 0.75, 0.4, 0.95},
	}
}

// ResponseBuilder provides fluent API for building mock responses
type ResponseBuilder struct {
	mock *MockLLMClient
}

// NewResponseBuilder creates a builder for configuring mock responses
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{mock: NewMockLLMClient()}
}

// WithGenerate sets Generate responses
func (b *ResponseBuilder) WithGenerate(responses ...[]string) *ResponseBuilder {
	b.mock.GenerateResponses = responses
	return b
}

// WithAggregate sets Aggregate responses
func (b *ResponseBuilder) WithAggregate(responses ...string) *ResponseBuilder {
	b.mock.AggregateResponses = responses
	return b
}

// WithRefine sets Refine responses
func (b *ResponseBuilder) WithRefine(responses ...string) *ResponseBuilder {
	b.mock.RefineResponses = responses
	return b
}

// WithScores sets Score responses
func (b *ResponseBuilder) WithScores(scores ...float64) *ResponseBuilder {
	b.mock.ScoreResponses = scores
	return b
}

// Build returns the configured mock
func (b *ResponseBuilder) Build() *MockLLMClient {
	return b.mock
}

// ContainsString checks if any call contained a specific substring
func (m *MockLLMClient) ContainsString(calls []string, substr string) bool {
	for _, call := range calls {
		if strings.Contains(call, substr) {
			return true
		}
	}
	return false
}

// MockHypothesisGenerator provides a mock implementation for testing abductive reasoning
// without making real LLM API calls.
type MockHypothesisGenerator struct {
	mu sync.Mutex

	// Response configuration
	Response string
	Error    error

	// Call tracking
	Calls []string
}

// NewMockHypothesisGenerator creates a mock with a sensible default response
func NewMockHypothesisGenerator() *MockHypothesisGenerator {
	return &MockHypothesisGenerator{
		Response: `{"hypotheses": [{"description": "Mock hypothesis: System issue detected", "assumptions": ["System is running"], "predictions": ["Issue will recur under similar conditions"], "parsimony": 0.8, "prior_probability": 0.6}]}`,
	}
}

// GenerateHypotheses implements the reasoning.HypothesisGenerator interface
func (m *MockHypothesisGenerator) GenerateHypotheses(ctx context.Context, prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, prompt)

	if m.Error != nil {
		return "", m.Error
	}
	return m.Response, nil
}

// WithResponse sets the response for GenerateHypotheses
func (m *MockHypothesisGenerator) WithResponse(response string) *MockHypothesisGenerator {
	m.Response = response
	return m
}

// WithError sets an error to be returned
func (m *MockHypothesisGenerator) WithError(err error) *MockHypothesisGenerator {
	m.Error = err
	return m
}

// Reset clears call history
func (m *MockHypothesisGenerator) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = nil
}
