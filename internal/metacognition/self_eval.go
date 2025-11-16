// Package metacognition provides self-evaluation and bias detection capabilities
// for cognitive reasoning processes.
package metacognition

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// SelfEvaluator performs metacognitive self-assessment
type SelfEvaluator struct {
	mu      sync.RWMutex
	counter int
}

// NewSelfEvaluator creates a new self-evaluator
func NewSelfEvaluator() *SelfEvaluator {
	return &SelfEvaluator{}
}

// EvaluateThought performs self-evaluation on a thought
func (se *SelfEvaluator) EvaluateThought(thought *types.Thought) (*types.SelfEvaluation, error) {
	se.mu.Lock()
	defer se.mu.Unlock()

	se.counter++

	qualityScore := se.assessQuality(thought)
	completenessScore := se.assessCompleteness(thought)
	coherenceScore := se.assessCoherence(thought)

	strengths := se.identifyStrengths(thought, qualityScore, completenessScore, coherenceScore)
	weaknesses := se.identifyWeaknesses(thought, qualityScore, completenessScore, coherenceScore)
	improvements := se.suggestImprovements(weaknesses)

	evaluation := &types.SelfEvaluation{
		ID:                     fmt.Sprintf("eval-%d", se.counter),
		ThoughtID:              thought.ID,
		QualityScore:           qualityScore,
		CompletenessScore:      completenessScore,
		CoherenceScore:         coherenceScore,
		Strengths:              strengths,
		Weaknesses:             weaknesses,
		ImprovementSuggestions: improvements,
		Metadata:               map[string]interface{}{},
		CreatedAt:              time.Now(),
	}

	return evaluation, nil
}

// EvaluateBranch performs self-evaluation on a branch
func (se *SelfEvaluator) EvaluateBranch(branch *types.Branch) (*types.SelfEvaluation, error) {
	se.mu.Lock()
	defer se.mu.Unlock()

	se.counter++

	// Aggregate assessment from all thoughts in branch
	qualityScore := se.assessBranchQuality(branch)
	completenessScore := se.assessBranchCompleteness(branch)
	coherenceScore := se.assessBranchCoherence(branch)

	strengths := se.identifyBranchStrengths(branch, qualityScore, completenessScore, coherenceScore)
	weaknesses := se.identifyBranchWeaknesses(branch, qualityScore, completenessScore, coherenceScore)
	improvements := se.suggestImprovements(weaknesses)

	evaluation := &types.SelfEvaluation{
		ID:                     fmt.Sprintf("eval-%d", se.counter),
		BranchID:               branch.ID,
		QualityScore:           qualityScore,
		CompletenessScore:      completenessScore,
		CoherenceScore:         coherenceScore,
		Strengths:              strengths,
		Weaknesses:             weaknesses,
		ImprovementSuggestions: improvements,
		Metadata:               map[string]interface{}{},
		CreatedAt:              time.Now(),
	}

	return evaluation, nil
}

// assessQuality evaluates the quality of reasoning in a thought
func (se *SelfEvaluator) assessQuality(thought *types.Thought) float64 {
	score := 0.5 // Base score

	// Check for evidence-based reasoning
	if strings.Contains(strings.ToLower(thought.Content), "evidence") ||
		strings.Contains(strings.ToLower(thought.Content), "data") ||
		strings.Contains(strings.ToLower(thought.Content), "study") {
		score += 0.15
	}

	// Check for logical structure
	if strings.Contains(strings.ToLower(thought.Content), "therefore") ||
		strings.Contains(strings.ToLower(thought.Content), "because") ||
		strings.Contains(strings.ToLower(thought.Content), "thus") {
		score += 0.10
	}

	// Check for consideration of alternatives
	if strings.Contains(strings.ToLower(thought.Content), "however") ||
		strings.Contains(strings.ToLower(thought.Content), "alternatively") ||
		strings.Contains(strings.ToLower(thought.Content), "on the other hand") {
		score += 0.15
	}

	// Confidence alignment (penalize overconfidence with little content)
	if thought.Confidence > 0.8 && len(thought.Content) < 100 {
		score -= 0.10
	}

	// Key points indicate structured thinking
	if len(thought.KeyPoints) > 0 {
		score += 0.10
	}

	return clamp(score, 0, 1)
}

// assessCompleteness evaluates thoroughness of thinking
func (se *SelfEvaluator) assessCompleteness(thought *types.Thought) float64 {
	score := 0.5

	// Length indicates depth
	length := len(thought.Content)
	if length > 200 {
		score += 0.15
	} else if length > 100 {
		score += 0.10
	} else if length < 50 {
		score -= 0.15
	}

	// Key points indicate comprehensive coverage
	keyPointScore := float64(len(thought.KeyPoints)) * 0.05
	score += keyPointScore

	// Check for consideration of multiple aspects
	if strings.Contains(strings.ToLower(thought.Content), "consider") {
		score += 0.10
	}

	return clamp(score, 0, 1)
}

// assessCoherence evaluates logical consistency
func (se *SelfEvaluator) assessCoherence(thought *types.Thought) float64 {
	score := 0.7 // Assume coherent by default

	content := strings.ToLower(thought.Content)

	// Penalize contradictory language
	if strings.Contains(content, "but") && strings.Contains(content, "however") {
		// Multiple hedges might indicate confusion
		score -= 0.10
	}

	// Penalize excessive hedging (uncertainty)
	hedges := []string{"maybe", "perhaps", "possibly", "might", "could"}
	hedgeCount := 0
	for _, hedge := range hedges {
		if strings.Contains(content, hedge) {
			hedgeCount++
		}
	}
	if hedgeCount > 2 {
		score -= 0.15
	}

	// Reward clear structure
	if strings.Contains(content, "first") || strings.Contains(content, "second") ||
		strings.Contains(content, "finally") {
		score += 0.10
	}

	return clamp(score, 0, 1)
}

// assessBranchQuality assesses quality of a whole branch
func (se *SelfEvaluator) assessBranchQuality(branch *types.Branch) float64 {
	if len(branch.Thoughts) == 0 {
		return 0.0
	}

	totalQuality := 0.0
	for _, thought := range branch.Thoughts {
		totalQuality += se.assessQuality(thought)
	}

	avgQuality := totalQuality / float64(len(branch.Thoughts))

	// Bonus for having insights
	if len(branch.Insights) > 0 {
		avgQuality += 0.05
	}

	// Bonus for cross-references (integration)
	if len(branch.CrossRefs) > 0 {
		avgQuality += 0.05
	}

	return clamp(avgQuality, 0, 1)
}

// assessBranchCompleteness assesses completeness of a branch
func (se *SelfEvaluator) assessBranchCompleteness(branch *types.Branch) float64 {
	if len(branch.Thoughts) == 0 {
		return 0.0
	}

	totalCompleteness := 0.0
	for _, thought := range branch.Thoughts {
		totalCompleteness += se.assessCompleteness(thought)
	}

	avgCompleteness := totalCompleteness / float64(len(branch.Thoughts))

	// Bonus for multiple thoughts (depth)
	if len(branch.Thoughts) >= 5 {
		avgCompleteness += 0.10
	} else if len(branch.Thoughts) >= 3 {
		avgCompleteness += 0.05
	}

	return clamp(avgCompleteness, 0, 1)
}

// assessBranchCoherence assesses coherence of a branch
func (se *SelfEvaluator) assessBranchCoherence(branch *types.Branch) float64 {
	if len(branch.Thoughts) == 0 {
		return 0.0
	}

	totalCoherence := 0.0
	for _, thought := range branch.Thoughts {
		totalCoherence += se.assessCoherence(thought)
	}

	return totalCoherence / float64(len(branch.Thoughts))
}

// identifyStrengths identifies strengths in thinking
func (se *SelfEvaluator) identifyStrengths(thought *types.Thought, quality, completeness, coherence float64) []string {
	strengths := make([]string, 0)

	// TIER 1: High quality (>=0.7)
	if quality >= 0.7 {
		strengths = append(strengths, "High-quality reasoning with evidence-based approach")
	} else if quality >= 0.6 {
		// TIER 2: Good quality (0.6-0.69)
		strengths = append(strengths, "Good quality reasoning with logical structure")
	} else if quality >= 0.5 {
		// TIER 3: Acceptable quality (0.5-0.59)
		strengths = append(strengths, "Basic reasoning structure present")
	}

	// TIER 1: Thorough completeness (>=0.7)
	if completeness >= 0.7 {
		strengths = append(strengths, "Thorough and comprehensive analysis")
	} else if completeness >= 0.6 {
		// TIER 2: Reasonably complete (0.6-0.69)
		strengths = append(strengths, "Reasonably complete analysis")
	} else if completeness >= 0.5 {
		// TIER 3: Covers basics (0.5-0.59)
		strengths = append(strengths, "Covers key aspects of the problem")
	}

	// TIER 1: Exceptional coherence (>=0.8)
	if coherence >= 0.8 {
		strengths = append(strengths, "Exceptionally clear and coherent structure")
	} else if coherence >= 0.7 {
		// TIER 2: Clear coherence (0.7-0.79)
		strengths = append(strengths, "Clear and logically coherent structure")
	} else if coherence >= 0.6 {
		// TIER 3: Coherent with minor issues (0.6-0.69)
		strengths = append(strengths, "Coherent with minor inconsistencies")
	}

	if thought.Confidence > 0.7 && quality > 0.7 {
		strengths = append(strengths, "Well-justified confidence level")
	}
	if len(thought.KeyPoints) >= 3 {
		strengths = append(strengths, "Well-structured with multiple key points")
	}

	return strengths
}

// identifyWeaknesses identifies weaknesses in thinking
func (se *SelfEvaluator) identifyWeaknesses(thought *types.Thought, quality, completeness, coherence float64) []string {
	weaknesses := make([]string, 0)

	// TIER 1: Critical quality issues (<0.4)
	if quality < 0.4 {
		weaknesses = append(weaknesses, "Reasoning lacks evidence and logical structure")
	} else if quality < 0.6 {
		// TIER 2: Moderate quality issues (0.4-0.59)
		weaknesses = append(weaknesses, "Could strengthen reasoning with more evidence")
	}

	// TIER 1: Critical completeness issues (<0.4)
	if completeness < 0.4 {
		weaknesses = append(weaknesses, "Analysis is incomplete or superficial")
	} else if completeness < 0.6 {
		// TIER 2: Moderate completeness issues (0.4-0.59)
		weaknesses = append(weaknesses, "Could expand analysis to cover more dimensions")
	}

	// TIER 1: Critical coherence issues (<0.5)
	if coherence < 0.5 {
		weaknesses = append(weaknesses, "Logical coherence needs significant improvement")
	} else if coherence < 0.7 {
		// TIER 2: Moderate coherence issues (0.5-0.69)
		weaknesses = append(weaknesses, "Some inconsistencies in logical flow")
	}

	if thought.Confidence > 0.8 && quality < 0.6 {
		weaknesses = append(weaknesses, "Confidence may be overestimated relative to reasoning quality")
	}
	if len(thought.Content) < 50 {
		weaknesses = append(weaknesses, "Content is too brief to support robust conclusions")
	}

	return weaknesses
}

// identifyBranchStrengths identifies strengths in branch
func (se *SelfEvaluator) identifyBranchStrengths(branch *types.Branch, quality, completeness, coherence float64) []string {
	strengths := se.identifyStrengths(&types.Thought{
		Content:    fmt.Sprintf("Branch with %d thoughts", len(branch.Thoughts)),
		Confidence: branch.Confidence,
		KeyPoints:  []string{},
	}, quality, completeness, coherence)

	if len(branch.Insights) > 0 {
		strengths = append(strengths, fmt.Sprintf("Generated %d valuable insights", len(branch.Insights)))
	}
	if len(branch.CrossRefs) > 0 {
		strengths = append(strengths, "Integrated with other branches through cross-references")
	}

	return strengths
}

// identifyBranchWeaknesses identifies weaknesses in branch
func (se *SelfEvaluator) identifyBranchWeaknesses(branch *types.Branch, quality, completeness, coherence float64) []string {
	weaknesses := se.identifyWeaknesses(&types.Thought{
		Content:    fmt.Sprintf("Branch with %d thoughts", len(branch.Thoughts)),
		Confidence: branch.Confidence,
		KeyPoints:  []string{},
	}, quality, completeness, coherence)

	if len(branch.Thoughts) < 2 {
		weaknesses = append(weaknesses, "Branch is underdeveloped with too few thoughts")
	}
	if len(branch.Insights) == 0 && len(branch.Thoughts) >= 3 {
		weaknesses = append(weaknesses, "No insights generated despite multiple thoughts")
	}

	return weaknesses
}

// suggestImprovements suggests improvements based on weaknesses
func (se *SelfEvaluator) suggestImprovements(weaknesses []string) []string {
	improvements := make([]string, 0)

	for _, weakness := range weaknesses {
		if strings.Contains(weakness, "evidence") {
			improvements = append(improvements, "Add supporting evidence and data")
		}
		if strings.Contains(weakness, "incomplete") || strings.Contains(weakness, "superficial") {
			improvements = append(improvements, "Expand analysis to cover additional dimensions")
		}
		if strings.Contains(weakness, "coherence") {
			improvements = append(improvements, "Improve logical flow and structure")
		}
		if strings.Contains(weakness, "confidence") {
			improvements = append(improvements, "Adjust confidence level to match reasoning quality")
		}
		if strings.Contains(weakness, "brief") {
			improvements = append(improvements, "Provide more detailed explanation and justification")
		}
		if strings.Contains(weakness, "insights") {
			improvements = append(improvements, "Synthesize insights from the thoughts")
		}
	}

	return improvements
}

// clamp restricts value to range [min, max]
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
