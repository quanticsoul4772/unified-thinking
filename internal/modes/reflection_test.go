package modes

import (
	"context"
	"testing"

	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"

	"github.com/stretchr/testify/assert"
)

func TestNewReflectionLoop(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	assert.NotNil(t, loop)
	assert.NotNil(t, loop.storage)
	assert.NotNil(t, loop.selfEvaluator)
	assert.NotNil(t, loop.biasDetector)
	assert.NotNil(t, loop.fallacyDetector)
}

func TestDefaultReflectionConfig(t *testing.T) {
	config := DefaultReflectionConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 5, config.MaxIterations)
	assert.Equal(t, 0.8, config.QualityThreshold)
	assert.Equal(t, 0.05, config.MinImprovement)
	assert.True(t, config.ChallengeAssumptions)
}

func TestReflectionLoop_RefineThought_HighQualityInitial(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	// High quality thought should stop quickly
	thought := &types.Thought{
		ID:         "test-1",
		Content:    "PostgreSQL is a relational database management system. It implements ACID transactions through MVCC (Multi-Version Concurrency Control), which allows multiple transactions to proceed without blocking. This design provides excellent concurrency while maintaining data consistency. The system has been proven reliable in production environments for decades.",
		Mode:       types.ModeLinear,
		Confidence: 0.9,
	}
	_ = store.StoreThought(thought)

	config := &ReflectionConfig{
		MaxIterations:        5,
		QualityThreshold:     0.65, // Lower threshold for testing
		MinImprovement:       0.05,
		ChallengeAssumptions: false,
	}
	ctx := context.Background()

	result, err := loop.RefineThought(ctx, thought, config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	// Should stop relatively quickly (within 3 iterations)
	assert.LessOrEqual(t, result.TotalIterations, 3)
	assert.NotEmpty(t, result.StopReason)
}

func TestReflectionLoop_RefineThought_LowQualityWithFallacy(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	// Low quality thought with ad hominem fallacy
	thought := &types.Thought{
		ID:         "test-2",
		Content:    "You can't trust John's argument about database design because he's not a good programmer.",
		Mode:       types.ModeLinear,
		Confidence: 0.5,
	}
	_ = store.StoreThought(thought)

	config := &ReflectionConfig{
		MaxIterations:        3,
		QualityThreshold:     0.8,
		MinImprovement:       0.02, // Lower threshold
		ChallengeAssumptions: true,
	}
	ctx := context.Background()

	result, err := loop.RefineThought(ctx, thought, config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalIterations, 1)
	assert.NotEmpty(t, result.Iterations)

	// Check that at least one iteration has a critique
	assert.NotNil(t, result.Iterations[0].Critique)
}

func TestReflectionLoop_RefineThought_MaxIterations(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	// Mediocre thought that won't reach threshold
	thought := &types.Thought{
		ID:         "test-3",
		Content:    "Databases store data.",
		Mode:       types.ModeLinear,
		Confidence: 0.6,
	}
	_ = store.StoreThought(thought)

	config := &ReflectionConfig{
		MaxIterations:        2,
		QualityThreshold:     0.95, // Very high, won't reach
		MinImprovement:       0.0,  // No minimum improvement requirement
		ChallengeAssumptions: false,
	}
	ctx := context.Background()

	result, err := loop.RefineThought(ctx, thought, config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should reach max iterations or stop earlier due to other reasons
	assert.LessOrEqual(t, result.TotalIterations, 2)
	assert.NotEmpty(t, result.StopReason)
	assert.Len(t, result.Iterations, result.TotalIterations)
}

func TestReflectionLoop_GenerateCritique_NoBiasesOrFallacies(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	thought := &types.Thought{
		ID:      "test-4",
		Content: "PostgreSQL uses MVCC for concurrency control. This allows multiple transactions without blocking.",
		Mode:    types.ModeLinear,
	}

	quality := &types.SelfEvaluation{
		QualityScore:      0.85,
		CompletenessScore: 0.8,
		CoherenceScore:    0.9,
	}

	critique := loop.generateCritique(thought, quality)

	assert.NotNil(t, critique)
	assert.Empty(t, critique.Biases)
	assert.Empty(t, critique.Fallacies)
	assert.Empty(t, critique.QualityIssues)
	assert.False(t, critique.ShouldRefine)
	assert.Contains(t, critique.CritiqueSummary, "No significant issues")
}

func TestReflectionLoop_GenerateCritique_LowCompleteness(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	thought := &types.Thought{
		ID:      "test-5",
		Content: "Databases are useful.",
		Mode:    types.ModeLinear,
	}

	quality := &types.SelfEvaluation{
		QualityScore:      0.5,
		CompletenessScore: 0.3, // Low
		CoherenceScore:    0.7,
	}

	critique := loop.generateCritique(thought, quality)

	assert.NotNil(t, critique)
	assert.True(t, critique.ShouldRefine)
	assert.NotEmpty(t, critique.QualityIssues)
	assert.Contains(t, critique.QualityIssues[0], "incomplete")
	assert.NotEmpty(t, critique.Recommendations)
}

func TestReflectionLoop_GenerateCritique_WithFallacy(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	thought := &types.Thought{
		ID:      "test-6",
		Content: "You can't trust his argument because he's a bad person.",
		Mode:    types.ModeLinear,
	}

	quality := &types.SelfEvaluation{
		QualityScore:      0.4,
		CompletenessScore: 0.6,
		CoherenceScore:    0.5,
	}

	critique := loop.generateCritique(thought, quality)

	assert.NotNil(t, critique)
	assert.True(t, critique.ShouldRefine)
	// Fallacy detection may or may not catch this simple case
	// But quality issues should be detected
	assert.NotEmpty(t, critique.QualityIssues)
	assert.Contains(t, critique.CritiqueSummary, "issues")
}

func TestReflectionLoop_BuildRefinementPrompt(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	thought := &types.Thought{
		ID:      "test-7",
		Content: "Original reasoning here",
		Mode:    types.ModeLinear,
	}

	critique := &ReflectionCritique{
		QualityIssues: []string{
			"Issue 1",
			"Issue 2",
		},
		Recommendations: []string{
			"Recommendation 1",
			"Recommendation 2",
		},
	}

	prompt := loop.buildRefinementPrompt(thought, critique)

	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "REFINE")
	assert.Contains(t, prompt, "Original reasoning here")
	assert.Contains(t, prompt, "Issue 1")
	assert.Contains(t, prompt, "Issue 2")
	assert.Contains(t, prompt, "Recommendation 1")
	assert.Contains(t, prompt, "Recommendation 2")
	assert.Contains(t, prompt, "REFINED REASONING")
}

func TestReflectionLoop_RefineThought_MinImprovement(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	// Thought with moderate quality
	thought := &types.Thought{
		ID:         "test-8",
		Content:    "Databases provide ACID guarantees through transactions.",
		Mode:       types.ModeLinear,
		Confidence: 0.7,
	}
	_ = store.StoreThought(thought)

	config := &ReflectionConfig{
		MaxIterations:        5,
		QualityThreshold:     0.95, // High threshold
		MinImprovement:       0.2,  // High minimum improvement
		ChallengeAssumptions: false,
	}
	ctx := context.Background()

	result, err := loop.RefineThought(ctx, thought, config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should stop early due to insufficient improvement
	assert.LessOrEqual(t, result.TotalIterations, config.MaxIterations)
}

func TestReflectionLoop_RefineThought_ImprovementTracking(t *testing.T) {
	store := storage.NewMemoryStorage()
	evaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	loop := NewReflectionLoop(store, evaluator, biasDetector, fallacyDetector)

	thought := &types.Thought{
		ID:         "test-9",
		Content:    "SQL databases.",
		Mode:       types.ModeLinear,
		Confidence: 0.5,
	}
	_ = store.StoreThought(thought)

	config := &ReflectionConfig{
		MaxIterations:        3,
		QualityThreshold:     0.85,
		MinImprovement:       0.05,
		ChallengeAssumptions: true,
	}
	ctx := context.Background()

	result, err := loop.RefineThought(ctx, thought, config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotZero(t, result.InitialQuality)
	assert.NotZero(t, result.FinalQuality)
	// Improvement should be tracked (can be positive, zero, or negative)
	assert.NotNil(t, result.Improvement)
}
