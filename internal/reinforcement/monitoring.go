// Package reinforcement provides monitoring tools for Thompson Sampling RL.
package reinforcement

import (
	"fmt"
	"math"
)

// PerformanceMetrics aggregates strategy performance data
type PerformanceMetrics struct {
	StrategyID     string
	StrategyName   string
	Mode           string
	TotalTrials    int
	TotalSuccesses int
	SuccessRate    float64
	Alpha          float64
	Beta           float64
	ExpectedRate   float64 // E[Beta(α,β)] = α/(α+β)
	ConvergenceGap float64 // |ExpectedRate - EmpiricalRate|
	LastUpdated    int64
}

// ComputePerformanceMetrics calculates comprehensive metrics for a strategy
func ComputePerformanceMetrics(strategy *Strategy) *PerformanceMetrics {
	successRate := strategy.SuccessRate()
	expectedRate := strategy.Alpha / (strategy.Alpha + strategy.Beta)
	convergenceGap := math.Abs(expectedRate - successRate)

	return &PerformanceMetrics{
		StrategyID:     strategy.ID,
		StrategyName:   strategy.Name,
		Mode:           strategy.Mode,
		TotalTrials:    strategy.TotalTrials,
		TotalSuccesses: strategy.TotalSuccesses,
		SuccessRate:    successRate,
		Alpha:          strategy.Alpha,
		Beta:           strategy.Beta,
		ExpectedRate:   expectedRate,
		ConvergenceGap: convergenceGap,
	}
}

// IsConverged checks if a strategy has converged based on convergence gap
func (pm *PerformanceMetrics) IsConverged(threshold float64) bool {
	if pm.TotalTrials < 20 {
		return false // Need minimum trials for convergence
	}
	return pm.ConvergenceGap < threshold
}

// ExplorationMetrics tracks exploration/exploitation balance
type ExplorationMetrics struct {
	TotalSelections   int
	GreedySelections  int
	ExplorationCount  int
	ExplorationRate   float64
	StrategyDiversity float64 // Shannon entropy
	UniqueStrategies  int
}

// ComputeExplorationMetrics calculates exploration metrics
func ComputeExplorationMetrics(selections map[string]int, greedyStrategyID string) *ExplorationMetrics {
	totalSelections := 0
	greedySelections := 0
	uniqueStrategies := 0

	for strategyID, count := range selections {
		totalSelections += count
		if strategyID == greedyStrategyID {
			greedySelections = count
		}
		if count > 0 {
			uniqueStrategies++
		}
	}

	explorationCount := totalSelections - greedySelections
	explorationRate := 0.0
	if totalSelections > 0 {
		explorationRate = float64(explorationCount) / float64(totalSelections)
	}

	// Compute Shannon entropy for diversity
	entropy := 0.0
	if totalSelections > 0 {
		for _, count := range selections {
			if count > 0 {
				p := float64(count) / float64(totalSelections)
				entropy -= p * math.Log2(p)
			}
		}
	}

	return &ExplorationMetrics{
		TotalSelections:   totalSelections,
		GreedySelections:  greedySelections,
		ExplorationCount:  explorationCount,
		ExplorationRate:   explorationRate,
		StrategyDiversity: entropy,
		UniqueStrategies:  uniqueStrategies,
	}
}

// NeedsMoreExploration checks if exploration rate is too low
func (em *ExplorationMetrics) NeedsMoreExploration(minRate float64) bool {
	return em.ExplorationRate < minRate
}

// IsExploringTooMuch checks if exploration rate is too high
func (em *ExplorationMetrics) IsExploringTooMuch(maxRate float64) bool {
	return em.ExplorationRate > maxRate
}

// LearningMetrics tracks learning progress
type LearningMetrics struct {
	InitialAccuracy     float64
	CurrentAccuracy     float64
	AccuracyImprovement float64
	TotalTrials         int
	LearningRate        float64 // Improvement per trial
	HasConverged        bool
	Trend               string // "improving", "declining", "stable"
}

// ComputeLearningMetrics calculates learning progress metrics
func ComputeLearningMetrics(outcomes []bool, windowSize int) *LearningMetrics {
	if len(outcomes) == 0 {
		return &LearningMetrics{}
	}

	// Initial accuracy (first windowSize or 20% of outcomes)
	initialWindow := min(windowSize, max(1, len(outcomes)/5))
	initialCorrect := 0
	for i := 0; i < initialWindow; i++ {
		if outcomes[i] {
			initialCorrect++
		}
	}
	initialAccuracy := float64(initialCorrect) / float64(initialWindow)

	// Current accuracy (last windowSize outcomes)
	currentWindow := min(windowSize, len(outcomes))
	currentCorrect := 0
	startIdx := len(outcomes) - currentWindow
	for i := startIdx; i < len(outcomes); i++ {
		if outcomes[i] {
			currentCorrect++
		}
	}
	currentAccuracy := float64(currentCorrect) / float64(currentWindow)

	// Improvement and learning rate
	improvement := currentAccuracy - initialAccuracy
	learningRate := 0.0
	if len(outcomes) > 1 {
		learningRate = improvement / float64(len(outcomes))
	}

	// Determine trend
	trend := "stable"
	if math.Abs(improvement) > 0.05 {
		if improvement > 0 {
			trend = "improving"
		} else {
			trend = "declining"
		}
	}

	// Check convergence (stable for last 20% of trials)
	hasConverged := len(outcomes) >= 50 && math.Abs(improvement) < 0.02

	return &LearningMetrics{
		InitialAccuracy:     initialAccuracy,
		CurrentAccuracy:     currentAccuracy,
		AccuracyImprovement: improvement,
		TotalTrials:         len(outcomes),
		LearningRate:        learningRate,
		HasConverged:        hasConverged,
		Trend:               trend,
	}
}

// FormatPerformanceReport generates a text summary of performance metrics
func FormatPerformanceReport(metrics *PerformanceMetrics) string {
	return fmt.Sprintf(`Strategy: %s (%s)
  Trials: %d
  Success Rate: %.2f%% (%d/%d)
  Thompson Parameters: α=%.2f, β=%.2f
  Expected Rate: %.2f%%
  Convergence Gap: %.4f
  Status: %s
`,
		metrics.StrategyName,
		metrics.Mode,
		metrics.TotalTrials,
		metrics.SuccessRate*100,
		metrics.TotalSuccesses,
		metrics.TotalTrials,
		metrics.Alpha,
		metrics.Beta,
		metrics.ExpectedRate*100,
		metrics.ConvergenceGap,
		convergenceStatus(metrics))
}

func convergenceStatus(m *PerformanceMetrics) string {
	if m.TotalTrials < 20 {
		return "Insufficient data (need 20+ trials)"
	}
	if m.IsConverged(0.05) {
		return "Converged"
	}
	if m.ConvergenceGap < 0.10 {
		return "Converging"
	}
	return "Exploring"
}

// FormatExplorationReport generates a text summary of exploration metrics
func FormatExplorationReport(metrics *ExplorationMetrics) string {
	status := "Balanced"
	if metrics.NeedsMoreExploration(0.15) {
		status = "Low exploration - consider exploration bonus"
	} else if metrics.IsExploringTooMuch(0.50) {
		status = "High exploration - may be wasteful"
	}

	return fmt.Sprintf(`Exploration/Exploitation Balance:
  Total Selections: %d
  Greedy Selections: %d (%.1f%%)
  Exploration: %d (%.1f%%)
  Unique Strategies: %d
  Diversity (entropy): %.3f
  Status: %s
`,
		metrics.TotalSelections,
		metrics.GreedySelections,
		float64(metrics.GreedySelections)/float64(metrics.TotalSelections)*100,
		metrics.ExplorationCount,
		metrics.ExplorationRate*100,
		metrics.UniqueStrategies,
		metrics.StrategyDiversity,
		status)
}

// FormatLearningReport generates a text summary of learning progress
func FormatLearningReport(metrics *LearningMetrics) string {
	improvementStr := fmt.Sprintf("%.2f%%", metrics.AccuracyImprovement*100)
	if metrics.AccuracyImprovement > 0 {
		improvementStr = "+" + improvementStr
	}

	return fmt.Sprintf(`Learning Progress:
  Initial Accuracy: %.2f%%
  Current Accuracy: %.2f%%
  Improvement: %s
  Learning Rate: %.4f per trial
  Total Trials: %d
  Trend: %s
  Converged: %v
`,
		metrics.InitialAccuracy*100,
		metrics.CurrentAccuracy*100,
		improvementStr,
		metrics.LearningRate,
		metrics.TotalTrials,
		metrics.Trend,
		metrics.HasConverged)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
