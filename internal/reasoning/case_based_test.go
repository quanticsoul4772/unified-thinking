package reasoning

import (
	"context"
	"testing"
	"time"

	"unified-thinking/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestNewCaseBasedReasoner(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)

	assert.NotNil(t, cbr)
	assert.NotNil(t, cbr.storage)
	assert.NotNil(t, cbr.cases)
	assert.NotNil(t, cbr.caseIndex)
	assert.NotNil(t, cbr.analogical)
}

func TestCBR_StoreCase(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	problem := &ProblemDescription{
		Description: "Need to optimize database queries",
		Context:     "High-traffic web application",
		Goals:       []string{"Reduce query time", "Improve throughput"},
		Constraints: []string{"Limited budget", "No downtime"},
		Features:    map[string]interface{}{"domain": "database", "priority": "high"},
	}

	solution := &SolutionDescription{
		Description: "Add database indexes and implement query caching",
		Approach:    "Incremental optimization",
		Steps:       []string{"Analyze slow queries", "Add indexes", "Implement caching"},
		Rationale:   "Indexes speed up lookups, caching reduces repeated queries",
	}

	outcome := &Outcome{
		Success:       true,
		Effectiveness: 0.85,
		TimeToSolve:   24 * time.Hour,
		LessonsLearned: []string{"Index selection is critical", "Cache invalidation is tricky"},
	}

	c := &Case{
		Problem:  problem,
		Solution: solution,
		Outcome:  outcome,
		Domain:   "database-optimization",
	}

	err := cbr.StoreCase(ctx, c)

	assert.NoError(t, err)
	assert.NotEmpty(t, c.ID)
	assert.NotZero(t, c.CreatedAt)
	assert.Equal(t, c.CreatedAt, c.UpdatedAt)
	assert.Contains(t, cbr.cases, c.ID)
}

func TestCBR_Retrieve_SimilarCases(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	// Store similar cases
	case1 := &Case{
		Problem: &ProblemDescription{
			Description: "Optimize database query performance",
			Context:     "E-commerce platform",
			Goals:       []string{"Faster queries"},
			Features:    map[string]interface{}{"type": "database"},
		},
		Solution: &SolutionDescription{
			Description: "Use indexes",
		},
		Outcome: &Outcome{
			Success:       true,
			Effectiveness: 0.9,
		},
		Domain: "database",
	}

	case2 := &Case{
		Problem: &ProblemDescription{
			Description: "Improve API response times",
			Context:     "Mobile app backend",
			Goals:       []string{"Reduce latency"},
			Features:    map[string]interface{}{"type": "api"},
		},
		Solution: &SolutionDescription{
			Description: "Add caching layer",
		},
		Outcome: &Outcome{
			Success:       true,
			Effectiveness: 0.8,
		},
		Domain: "api",
	}

	cbr.StoreCase(ctx, case1)
	cbr.StoreCase(ctx, case2)

	// Retrieve cases similar to database problem
	req := &RetrieveRequest{
		Problem: &ProblemDescription{
			Description: "Database queries are slow",
			Context:     "Web application",
			Goals:       []string{"Speed up queries"},
			Features:    map[string]interface{}{"type": "database"},
		},
		Domain:        "database",
		MaxCases:      5,
		MinSimilarity: 0.1,
	}

	result, err := cbr.Retrieve(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.Retrieved, 0)
	assert.NotEmpty(t, result.Cases)

	// Most similar should be case1
	mostSimilar := result.Cases[0]
	assert.Greater(t, mostSimilar.Similarity, 0.0)
}

func TestCBR_Retrieve_NoMatches(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	// Store a case
	c := &Case{
		Problem: &ProblemDescription{
			Description: "Database optimization",
			Features:    map[string]interface{}{"type": "database"},
		},
		Solution: &SolutionDescription{
			Description: "Use indexes",
		},
		Outcome: &Outcome{Success: true, Effectiveness: 0.9},
		Domain:  "database",
	}
	cbr.StoreCase(ctx, c)

	// Try to retrieve completely different problem
	req := &RetrieveRequest{
		Problem: &ProblemDescription{
			Description: "UI design for mobile app",
			Features:    map[string]interface{}{"type": "design"},
		},
		MinSimilarity: 0.8, // High threshold
	}

	result, err := cbr.Retrieve(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Retrieved)
}

func TestCBR_Reuse_DirectStrategy(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	similarCase := &SimilarCase{
		Case: &Case{
			Problem: &ProblemDescription{
				Description: "Optimize queries",
			},
			Solution: &SolutionDescription{
				Description: "Add database indexes",
				Steps:       []string{"Identify slow queries", "Create indexes"},
			},
		},
		Similarity: 0.9,
	}

	targetProblem := &ProblemDescription{
		Description: "Improve query speed",
	}

	result, err := cbr.Reuse(ctx, similarCase, targetProblem, AdaptDirect)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, AdaptDirect, result.Strategy)
	assert.Equal(t, similarCase.Case, result.OriginalCase)
	assert.Equal(t, similarCase.Similarity, result.Confidence)
	assert.NotEmpty(t, result.Adaptations)
}

func TestCBR_Reuse_SubstituteStrategy(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	similarCase := &SimilarCase{
		Case: &Case{
			Problem: &ProblemDescription{
				Description: "Optimize database performance",
			},
			Solution: &SolutionDescription{
				Description: "database needs optimization",
				Steps:       []string{"Analyze database"},
			},
		},
		Similarity: 0.8,
	}

	targetProblem := &ProblemDescription{
		Description: "Improve API performance",
	}

	result, err := cbr.Reuse(ctx, similarCase, targetProblem, AdaptSubstitute)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, AdaptSubstitute, result.Strategy)
	assert.Less(t, result.Confidence, similarCase.Similarity) // Reduced due to adaptation
}

func TestCBR_Revise_Success(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	reuseResult := &ReuseResult{
		OriginalCase: &Case{},
		AdaptedSolution: &SolutionDescription{
			Description: "Optimized solution",
			Rationale:   "Original rationale",
		},
		Strategy:   AdaptDirect,
		Confidence: 0.8,
	}

	result, err := cbr.Revise(ctx, reuseResult, "Solution worked perfectly", true)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.Confidence, reuseResult.Confidence) // Increased on success
	assert.NotEmpty(t, result.Changes)
}

func TestCBR_Revise_Failure(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	reuseResult := &ReuseResult{
		OriginalCase: &Case{},
		AdaptedSolution: &SolutionDescription{
			Description: "Attempted solution",
			Rationale:   "Original rationale",
		},
		Strategy:   AdaptDirect,
		Confidence: 0.8,
	}

	result, err := cbr.Revise(ctx, reuseResult, "Solution didn't work due to constraints", false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Less(t, result.Confidence, reuseResult.Confidence) // Decreased on failure
	assert.Contains(t, result.RevisedSolution.Rationale, "failure")
	assert.NotEmpty(t, result.Changes)
}

func TestCBR_Retain(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	problem := &ProblemDescription{
		Description: "Test problem",
		Goals:       []string{"Goal 1"},
	}

	solution := &SolutionDescription{
		Description: "Test solution",
	}

	outcome := &Outcome{
		Success:       true,
		Effectiveness: 0.85,
	}

	newCase, err := cbr.Retain(ctx, problem, solution, outcome, "test-domain")

	assert.NoError(t, err)
	assert.NotNil(t, newCase)
	assert.NotEmpty(t, newCase.ID)
	assert.Equal(t, problem, newCase.Problem)
	assert.Equal(t, solution, newCase.Solution)
	assert.Equal(t, outcome, newCase.Outcome)
	assert.Equal(t, "test-domain", newCase.Domain)
	assert.Equal(t, 0.85, newCase.SuccessRate)
	assert.Contains(t, cbr.cases, newCase.ID)
}

func TestCBR_PerformCBRCycle(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)
	ctx := context.Background()

	// First, store a successful case
	existingCase := &Case{
		Problem: &ProblemDescription{
			Description: "Slow database queries",
			Goals:       []string{"Improve speed"},
		},
		Solution: &SolutionDescription{
			Description: "Add indexes and caching",
			Steps:       []string{"Analyze", "Index", "Cache"},
		},
		Outcome: &Outcome{
			Success:       true,
			Effectiveness: 0.9,
		},
		Domain:      "database",
		SuccessRate: 0.9,
	}
	cbr.StoreCase(ctx, existingCase)

	// Now perform CBR cycle for similar problem
	newProblem := &ProblemDescription{
		Description: "Database performance issues",
		Goals:       []string{"Speed up queries"},
	}

	cycle, err := cbr.PerformCBRCycle(ctx, newProblem, "database")

	assert.NoError(t, err)
	assert.NotNil(t, cycle)
	assert.NotNil(t, cycle.Retrieved)
	if cycle.Retrieved.Retrieved > 0 {
		assert.NotNil(t, cycle.Reused)
		assert.Greater(t, cycle.Reused.Confidence, 0.0)
	}
}

func TestCBR_TextSimilarity(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)

	tests := []struct {
		name     string
		text1    string
		text2    string
		expected float64
	}{
		{
			name:     "identical",
			text1:    "optimize database queries",
			text2:    "optimize database queries",
			expected: 1.0,
		},
		{
			name:     "completely_different",
			text1:    "optimize database queries",
			text2:    "design mobile interface",
			expected: 0.0,
		},
		{
			name:     "partial_overlap",
			text1:    "optimize database performance",
			text2:    "improve database speed",
			expected: 0.2, // "database" overlaps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := cbr.textSimilarity(tt.text1, tt.text2)
			assert.InDelta(t, tt.expected, similarity, 0.15)
		})
	}
}

func TestCBR_SetOverlap(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)

	tests := []struct {
		name     string
		set1     []string
		set2     []string
		minScore float64
	}{
		{
			name:     "identical_sets",
			set1:     []string{"goal1", "goal2"},
			set2:     []string{"goal1", "goal2"},
			minScore: 0.9,
		},
		{
			name:     "no_overlap",
			set1:     []string{"goal1"},
			set2:     []string{"goal2"},
			minScore: 0.0,
		},
		{
			name:     "partial_overlap",
			set1:     []string{"goal1", "goal2", "goal3"},
			set2:     []string{"goal2", "goal4"},
			minScore: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlap := cbr.setOverlap(tt.set1, tt.set2)
			assert.GreaterOrEqual(t, overlap, tt.minScore)
		})
	}
}

func TestCBR_CalculateSimilarity(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)

	p1 := &ProblemDescription{
		Description: "Optimize database queries for performance",
		Context:     "E-commerce web application",
		Goals:       []string{"Reduce query time", "Improve throughput"},
		Constraints: []string{"Limited budget"},
		Features:    map[string]interface{}{"domain": "database", "priority": "high"},
	}

	p2 := &ProblemDescription{
		Description: "Improve database query speed",
		Context:     "E-commerce platform",
		Goals:       []string{"Reduce query time"},
		Constraints: []string{"Limited budget"},
		Features:    map[string]interface{}{"domain": "database"},
	}

	similarity := cbr.calculateSimilarity(p1, p2)

	assert.Greater(t, similarity, 0.3) // Should be somewhat similar
	assert.LessOrEqual(t, similarity, 1.0)
}

func TestCBR_DeepCopySolution(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)

	original := &SolutionDescription{
		Description: "Original solution",
		Approach:    "Test approach",
		Steps:       []string{"Step 1", "Step 2"},
		Rationale:   "Test rationale",
		Assumptions: []string{"Assumption 1"},
		Resources:   []string{"Resource 1"},
	}

	copied := cbr.deepCopySolution(original)

	// Verify deep copy
	assert.Equal(t, original.Description, copied.Description)
	assert.Equal(t, len(original.Steps), len(copied.Steps))

	// Modify original
	original.Description = "Modified"
	original.Steps[0] = "Modified Step"

	// Verify copy wasn't affected
	assert.Equal(t, "Original solution", copied.Description)
	assert.Equal(t, "Step 1", copied.Steps[0])
}

func TestCBR_ExtractTags(t *testing.T) {
	store := storage.NewMemoryStorage()
	cbr := NewCaseBasedReasoner(store)

	problem := &ProblemDescription{
		Description: "Optimize database performance using indexes and caching strategies",
	}

	tags := cbr.extractTags(problem)

	assert.NotEmpty(t, tags)
	assert.LessOrEqual(t, len(tags), 10) // Limited to 10 tags
}

func TestCBR_IndexCase(t *testing.T) {
	index := NewCaseIndex()

	c := &Case{
		ID:     "case-1",
		Domain: "database",
		Tags:   []string{"optimization", "performance"},
		Problem: &ProblemDescription{
			Features: map[string]interface{}{"type": "query", "priority": "high"},
		},
	}

	index.indexCase(c)

	assert.Contains(t, index.byDomain["database"], "case-1")
	assert.Contains(t, index.byTags["optimization"], "case-1")
	assert.Contains(t, index.byTags["performance"], "case-1")
	assert.Contains(t, index.byFeatures["type"], "case-1")
	assert.Contains(t, index.byFeatures["priority"], "case-1")
}
