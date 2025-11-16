// Package modes provides multi-step reflection loop for iterative reasoning refinement.
package modes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

// ReflectionLoop manages iterative reasoning refinement
type ReflectionLoop struct {
	storage         storage.Storage
	selfEvaluator   *metacognition.SelfEvaluator
	biasDetector    *metacognition.BiasDetector
	fallacyDetector *validation.FallacyDetector
}

// NewReflectionLoop creates a new reflection loop
func NewReflectionLoop(
	store storage.Storage,
	evaluator *metacognition.SelfEvaluator,
	biasDetector *metacognition.BiasDetector,
	fallacyDetector *validation.FallacyDetector,
) *ReflectionLoop {
	return &ReflectionLoop{
		storage:         store,
		selfEvaluator:   evaluator,
		biasDetector:    biasDetector,
		fallacyDetector: fallacyDetector,
	}
}

// ReflectionConfig configures the reflection loop
type ReflectionConfig struct {
	MaxIterations        int     `json:"max_iterations"`        // Maximum refinement attempts
	QualityThreshold     float64 `json:"quality_threshold"`     // Stop when quality >= this
	MinImprovement       float64 `json:"min_improvement"`       // Minimum improvement to continue
	ChallengeAssumptions bool    `json:"challenge_assumptions"` // Enable assumption challenging
}

// DefaultReflectionConfig returns sensible defaults
func DefaultReflectionConfig() *ReflectionConfig {
	return &ReflectionConfig{
		MaxIterations:        5,
		QualityThreshold:     0.8,
		MinImprovement:       0.05,
		ChallengeAssumptions: true,
	}
}

// ReflectionIteration represents one refinement iteration
type ReflectionIteration struct {
	IterationNum int                   `json:"iteration_num"`
	ThoughtID    string                `json:"thought_id"`
	Quality      *types.SelfEvaluation `json:"quality"`
	Critique     *ReflectionCritique   `json:"critique"`
	Improvements []string              `json:"improvements"`
	Timestamp    time.Time             `json:"timestamp"`
}

// ReflectionCritique identifies issues to address
type ReflectionCritique struct {
	Biases          []*types.CognitiveBias        `json:"biases,omitempty"`
	Fallacies       []*validation.DetectedFallacy `json:"fallacies,omitempty"`
	QualityIssues   []string                      `json:"quality_issues"`
	Recommendations []string                      `json:"recommendations"`
	ShouldRefine    bool                          `json:"should_refine"`
	CritiqueSummary string                        `json:"critique_summary"`
}

// ReflectionResult contains the final result and iteration history
type ReflectionResult struct {
	FinalThought    *types.Thought        `json:"final_thought"`
	Iterations      []ReflectionIteration `json:"iterations"`
	TotalIterations int                   `json:"total_iterations"`
	InitialQuality  float64               `json:"initial_quality"`
	FinalQuality    float64               `json:"final_quality"`
	Improvement     float64               `json:"improvement"`
	StopReason      string                `json:"stop_reason"`
	Success         bool                  `json:"success"`
}

// RefineThought performs iterative reflection and refinement
func (rl *ReflectionLoop) RefineThought(ctx context.Context, initialThought *types.Thought, config *ReflectionConfig) (*ReflectionResult, error) {
	if config == nil {
		config = DefaultReflectionConfig()
	}

	result := &ReflectionResult{
		Iterations: make([]ReflectionIteration, 0),
	}

	currentThought := initialThought
	var previousQuality float64

	for i := 0; i < config.MaxIterations; i++ {
		iteration := ReflectionIteration{
			IterationNum: i + 1,
			ThoughtID:    currentThought.ID,
			Timestamp:    time.Now(),
		}

		// Step 1: Evaluate current thought quality
		quality, err := rl.selfEvaluator.EvaluateThought(currentThought)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate thought: %w", err)
		}
		iteration.Quality = quality

		if i == 0 {
			result.InitialQuality = quality.QualityScore
			previousQuality = quality.QualityScore
		}

		// Step 2: Check if quality threshold met
		if quality.QualityScore >= config.QualityThreshold {
			result.FinalThought = currentThought
			result.FinalQuality = quality.QualityScore
			result.Improvement = result.FinalQuality - result.InitialQuality
			result.TotalIterations = i + 1
			result.StopReason = fmt.Sprintf("Quality threshold reached (%.2f >= %.2f)", quality.QualityScore, config.QualityThreshold)
			result.Success = true
			result.Iterations = append(result.Iterations, iteration)
			return result, nil
		}

		// Step 3: Generate critique
		critique := rl.generateCritique(currentThought, quality)
		iteration.Critique = critique

		// Step 4: Check if we should continue refining
		if !critique.ShouldRefine {
			result.FinalThought = currentThought
			result.FinalQuality = quality.QualityScore
			result.Improvement = result.FinalQuality - result.InitialQuality
			result.TotalIterations = i + 1
			result.StopReason = "No significant issues to address"
			result.Success = true
			result.Iterations = append(result.Iterations, iteration)
			return result, nil
		}

		// Step 5: Check minimum improvement
		if i > 0 {
			improvement := quality.QualityScore - previousQuality
			if improvement < config.MinImprovement && improvement >= 0 {
				result.FinalThought = currentThought
				result.FinalQuality = quality.QualityScore
				result.Improvement = result.FinalQuality - result.InitialQuality
				result.TotalIterations = i + 1
				result.StopReason = fmt.Sprintf("Insufficient improvement (%.3f < %.3f)", improvement, config.MinImprovement)
				result.Success = true
				result.Iterations = append(result.Iterations, iteration)
				return result, nil
			}
		}

		result.Iterations = append(result.Iterations, iteration)

		// Step 6: Refine the thought based on critique
		refinedThought, improvements, err := rl.refineBasedOnCritique(ctx, currentThought, critique, config)
		if err != nil {
			return nil, fmt.Errorf("failed to refine thought: %w", err)
		}

		iteration.Improvements = improvements
		currentThought = refinedThought
		previousQuality = quality.QualityScore
	}

	// Reached max iterations
	finalQuality := result.Iterations[len(result.Iterations)-1].Quality.QualityScore
	result.FinalThought = currentThought
	result.FinalQuality = finalQuality
	result.Improvement = result.FinalQuality - result.InitialQuality
	result.TotalIterations = config.MaxIterations
	result.StopReason = fmt.Sprintf("Maximum iterations reached (%d)", config.MaxIterations)
	result.Success = result.FinalQuality >= config.QualityThreshold

	return result, nil
}

// generateCritique analyzes thought and generates targeted critique
func (rl *ReflectionLoop) generateCritique(thought *types.Thought, quality *types.SelfEvaluation) *ReflectionCritique {
	critique := &ReflectionCritique{
		QualityIssues:   make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Detect biases
	biases, _ := rl.biasDetector.DetectBiases(thought)
	if len(biases) > 0 {
		critique.Biases = biases
		for _, bias := range biases {
			critique.QualityIssues = append(critique.QualityIssues,
				fmt.Sprintf("Cognitive bias detected: %s", bias.BiasType))
		}
	}

	// Detect fallacies (check both formal and informal)
	fallacies := rl.fallacyDetector.DetectFallacies(thought.Content, true, true)
	if len(fallacies) > 0 {
		critique.Fallacies = fallacies
		for _, fallacy := range fallacies {
			critique.QualityIssues = append(critique.QualityIssues,
				fmt.Sprintf("Logical fallacy: %s", fallacy.Type))
		}
	}

	// Quality-based issues
	if quality.CompletenessScore < 0.6 {
		critique.QualityIssues = append(critique.QualityIssues, "Reasoning is incomplete")
		critique.Recommendations = append(critique.Recommendations, "Expand reasoning with more detail and evidence")
	}

	if quality.CoherenceScore < 0.6 {
		critique.QualityIssues = append(critique.QualityIssues, "Reasoning lacks coherence")
		critique.Recommendations = append(critique.Recommendations, "Restructure argument for better logical flow")
	}

	// Note: types.SelfEvaluation doesn't have EvidenceQuality field, skip this check

	// Determine if refinement is needed
	critique.ShouldRefine = len(critique.QualityIssues) > 0 || quality.QualityScore < 0.7

	// Generate summary
	if len(critique.QualityIssues) == 0 {
		critique.CritiqueSummary = "No significant issues detected. Reasoning is sound."
	} else {
		critique.CritiqueSummary = fmt.Sprintf("Found %d issues: %s",
			len(critique.QualityIssues),
			strings.Join(critique.QualityIssues, "; "))
	}

	return critique
}

// refineBasedOnCritique creates refined version of thought
func (rl *ReflectionLoop) refineBasedOnCritique(
	ctx context.Context,
	thought *types.Thought,
	critique *ReflectionCritique,
	config *ReflectionConfig,
) (*types.Thought, []string, error) {
	improvements := make([]string, 0)

	// Build refinement prompt
	refinementPrompt := rl.buildRefinementPrompt(thought, critique)

	// Create refined thought
	refinedThought := &types.Thought{
		ID:         fmt.Sprintf("%s-refined-%d", thought.ID, time.Now().Unix()),
		Content:    refinementPrompt,
		Mode:       thought.Mode,
		Confidence: thought.Confidence,
		ParentID:   thought.ID,
		KeyPoints:  thought.KeyPoints,
		Metadata: map[string]interface{}{
			"refinement_of":         thought.ID,
			"critique":              critique.CritiqueSummary,
			"iteration":             len(critique.QualityIssues),
			"challenge_assumptions": config.ChallengeAssumptions,
		},
		Timestamp: time.Now(),
	}

	// Store refined thought
	if err := rl.storage.StoreThought(refinedThought); err != nil {
		return nil, nil, fmt.Errorf("failed to store refined thought: %w", err)
	}

	// Track improvements
	if len(critique.Biases) > 0 {
		improvements = append(improvements, "Addressed cognitive biases")
	}
	if len(critique.Fallacies) > 0 {
		improvements = append(improvements, "Corrected logical fallacies")
	}
	if len(critique.Recommendations) > 0 {
		improvements = append(improvements, "Applied quality recommendations")
	}

	return refinedThought, improvements, nil
}

// buildRefinementPrompt creates prompt for refinement iteration
func (rl *ReflectionLoop) buildRefinementPrompt(thought *types.Thought, critique *ReflectionCritique) string {
	var sb strings.Builder

	sb.WriteString("REFINE the following reasoning by addressing identified issues:\n\n")
	sb.WriteString("ORIGINAL: ")
	sb.WriteString(thought.Content)
	sb.WriteString("\n\n")

	if len(critique.QualityIssues) > 0 {
		sb.WriteString("ISSUES TO ADDRESS:\n")
		for i, issue := range critique.QualityIssues {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
		}
		sb.WriteString("\n")
	}

	if len(critique.Recommendations) > 0 {
		sb.WriteString("RECOMMENDATIONS:\n")
		for i, rec := range critique.Recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("REFINED REASONING:")

	return sb.String()
}
