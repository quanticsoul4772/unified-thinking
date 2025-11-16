// Package memory - Retrospective analysis for reasoning sessions
package memory

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// RetrospectiveAnalysis provides detailed analysis of a completed trajectory
type RetrospectiveAnalysis struct {
	TrajectoryID        string                   `json:"trajectory_id"`
	SessionID           string                   `json:"session_id"`
	AnalysisTimestamp   time.Time                `json:"analysis_timestamp"`
	Summary             *AnalysisSummary         `json:"summary"`
	Strengths           []string                 `json:"strengths"`
	Weaknesses          []string                 `json:"weaknesses"`
	Improvements        []*ImprovementSuggestion `json:"improvements"`
	LessonsLearned      []string                 `json:"lessons_learned"`
	ComparativeAnalysis *ComparativeAnalysis     `json:"comparative_analysis,omitempty"`
	DetailedMetrics     *DetailedMetricsAnalysis `json:"detailed_metrics"`
}

// AnalysisSummary provides high-level summary of the session
type AnalysisSummary struct {
	OverallAssessment string  `json:"overall_assessment"` // "excellent", "good", "fair", "poor"
	SuccessScore      float64 `json:"success_score"`
	QualityScore      float64 `json:"quality_score"`
	Duration          string  `json:"duration"`
	StepCount         int     `json:"step_count"`
	StrategyUsed      string  `json:"strategy_used"`
	PrimaryOutcome    string  `json:"primary_outcome"`
}

// ImprovementSuggestion provides specific actionable improvement
type ImprovementSuggestion struct {
	Category       string  `json:"category"` // "efficiency", "quality", "approach", "tools"
	Priority       float64 `json:"priority"` // 0.0-1.0
	Suggestion     string  `json:"suggestion"`
	Rationale      string  `json:"rationale"`
	ExpectedImpact string  `json:"expected_impact"`
}

// ComparativeAnalysis compares this trajectory to similar ones
type ComparativeAnalysis struct {
	SimilarTrajectories int      `json:"similar_trajectories_count"`
	PercentileRank      float64  `json:"percentile_rank"` // 0-100, higher is better
	BetterThan          int      `json:"better_than_count"`
	WorseThan           int      `json:"worse_than_count"`
	KeyDifferences      []string `json:"key_differences"`
}

// DetailedMetricsAnalysis provides deep dive into quality metrics
type DetailedMetricsAnalysis struct {
	EfficiencyAnalysis   *MetricAnalysis `json:"efficiency_analysis"`
	CoherenceAnalysis    *MetricAnalysis `json:"coherence_analysis"`
	CompletenessAnalysis *MetricAnalysis `json:"completeness_analysis"`
	InnovationAnalysis   *MetricAnalysis `json:"innovation_analysis"`
	ReliabilityAnalysis  *MetricAnalysis `json:"reliability_analysis"`
}

// MetricAnalysis provides analysis for a specific metric
type MetricAnalysis struct {
	Score       float64  `json:"score"`
	Assessment  string   `json:"assessment"` // "excellent", "good", "fair", "poor"
	Explanation string   `json:"explanation"`
	Suggestions []string `json:"suggestions"`
}

// RetrospectiveAnalyzer performs post-session analysis
type RetrospectiveAnalyzer struct {
	store *EpisodicMemoryStore
}

// NewRetrospectiveAnalyzer creates a new retrospective analyzer
func NewRetrospectiveAnalyzer(store *EpisodicMemoryStore) *RetrospectiveAnalyzer {
	return &RetrospectiveAnalyzer{
		store: store,
	}
}

// AnalyzeTrajectory performs comprehensive retrospective analysis
func (r *RetrospectiveAnalyzer) AnalyzeTrajectory(ctx context.Context, trajectoryID string) (*RetrospectiveAnalysis, error) {
	r.store.mu.RLock()
	trajectory, exists := r.store.trajectories[trajectoryID]
	r.store.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("trajectory not found: %s", trajectoryID)
	}

	// Initialize arrays to prevent nil values (MCP requires arrays, not null)
	strengths := r.identifyStrengths(trajectory)
	if strengths == nil {
		strengths = []string{}
	}
	weaknesses := r.identifyWeaknesses(trajectory)
	if weaknesses == nil {
		weaknesses = []string{}
	}
	improvements := r.generateImprovements(trajectory)
	if improvements == nil {
		improvements = []*ImprovementSuggestion{}
	}
	lessons := r.extractLessons(trajectory)
	if lessons == nil {
		lessons = []string{}
	}

	analysis := &RetrospectiveAnalysis{
		TrajectoryID:      trajectoryID,
		SessionID:         trajectory.SessionID,
		AnalysisTimestamp: time.Now(),
		Summary:           r.generateSummary(trajectory),
		Strengths:         strengths,
		Weaknesses:        weaknesses,
		Improvements:      improvements,
		LessonsLearned:    lessons,
		DetailedMetrics:   r.analyzeMetrics(trajectory),
	}

	// Add comparative analysis if similar trajectories exist
	if comparative := r.performComparativeAnalysis(ctx, trajectory); comparative != nil {
		analysis.ComparativeAnalysis = comparative
	}

	// Final defensive check - ensure NO nil arrays (triple-check for MCP validation)
	if analysis.Strengths == nil {
		analysis.Strengths = []string{}
	}
	if analysis.Weaknesses == nil {
		analysis.Weaknesses = []string{}
	}
	if analysis.Improvements == nil {
		analysis.Improvements = []*ImprovementSuggestion{}
	}
	if analysis.LessonsLearned == nil {
		analysis.LessonsLearned = []string{}
	}

	return analysis, nil
}

// generateSummary creates high-level summary
func (r *RetrospectiveAnalyzer) generateSummary(trajectory *ReasoningTrajectory) *AnalysisSummary {
	assessment := "fair"
	if trajectory.SuccessScore >= 0.8 {
		assessment = "excellent"
	} else if trajectory.SuccessScore >= 0.6 {
		assessment = "good"
	} else if trajectory.SuccessScore < 0.4 {
		assessment = "poor"
	}

	qualityScore := 0.5
	if trajectory.Quality != nil {
		qualityScore = trajectory.Quality.OverallQuality
	}

	strategy := "unknown"
	if trajectory.Approach != nil {
		strategy = trajectory.Approach.Strategy
	}

	outcome := "unknown"
	if trajectory.Outcome != nil {
		outcome = trajectory.Outcome.Status
	}

	return &AnalysisSummary{
		OverallAssessment: assessment,
		SuccessScore:      trajectory.SuccessScore,
		QualityScore:      qualityScore,
		Duration:          trajectory.Duration.String(),
		StepCount:         len(trajectory.Steps),
		StrategyUsed:      strategy,
		PrimaryOutcome:    outcome,
	}
}

// identifyStrengths identifies what went well
func (r *RetrospectiveAnalyzer) identifyStrengths(trajectory *ReasoningTrajectory) []string {
	strengths := make([]string, 0)

	if trajectory.Quality == nil {
		return strengths
	}

	q := trajectory.Quality

	if q.Efficiency >= 0.8 {
		strengths = append(strengths, fmt.Sprintf("Excellent efficiency (%.1f%% - used optimal number of steps)", q.Efficiency*100))
	}

	if q.Coherence >= 0.8 {
		strengths = append(strengths, fmt.Sprintf("Strong logical coherence (%.1f%% - reasoning was consistent)", q.Coherence*100))
	}

	if q.Completeness >= 0.8 {
		strengths = append(strengths, fmt.Sprintf("High completeness (%.1f%% - achieved most goals)", q.Completeness*100))
	}

	if q.Innovation >= 0.6 {
		strengths = append(strengths, fmt.Sprintf("Good innovation (%.1f%% - used creative approaches)", q.Innovation*100))
	}

	if q.Reliability >= 0.8 {
		strengths = append(strengths, fmt.Sprintf("High confidence in results (%.1f%%)", q.Reliability*100))
	}

	if q.FallacyCount == 0 && q.ContradictionCount == 0 {
		strengths = append(strengths, "No logical fallacies or contradictions detected")
	}

	if len(trajectory.Steps) > 0 {
		successfulSteps := 0
		for _, step := range trajectory.Steps {
			if step.Success {
				successfulSteps++
			}
		}
		successRate := float64(successfulSteps) / float64(len(trajectory.Steps))
		if successRate >= 0.9 {
			strengths = append(strengths, fmt.Sprintf("High step success rate (%.0f%%)", successRate*100))
		}
	}

	return strengths
}

// identifyWeaknesses identifies areas for improvement
func (r *RetrospectiveAnalyzer) identifyWeaknesses(trajectory *ReasoningTrajectory) []string {
	weaknesses := make([]string, 0)

	if trajectory.Quality == nil {
		return weaknesses
	}

	q := trajectory.Quality

	if q.Efficiency < 0.5 {
		weaknesses = append(weaknesses, fmt.Sprintf("Low efficiency (%.1f%% - too many steps for the problem)", q.Efficiency*100))
	}

	if q.Coherence < 0.5 {
		weaknesses = append(weaknesses, fmt.Sprintf("Poor logical coherence (%.1f%% - reasoning had inconsistencies)", q.Coherence*100))
	}

	if q.Completeness < 0.5 {
		weaknesses = append(weaknesses, fmt.Sprintf("Low completeness (%.1f%% - many goals unachieved)", q.Completeness*100))
	}

	if q.Innovation < 0.3 {
		weaknesses = append(weaknesses, fmt.Sprintf("Limited innovation (%.1f%% - relied on basic approaches)", q.Innovation*100))
	}

	if q.Reliability < 0.5 {
		weaknesses = append(weaknesses, fmt.Sprintf("Low confidence in results (%.1f%%)", q.Reliability*100))
	}

	if q.FallacyCount > 3 {
		weaknesses = append(weaknesses, fmt.Sprintf("Multiple logical fallacies detected (%d)", q.FallacyCount))
	}

	if q.ContradictionCount > 2 {
		weaknesses = append(weaknesses, fmt.Sprintf("Several contradictions found (%d)", q.ContradictionCount))
	}

	if q.BiasScore > 0.5 {
		weaknesses = append(weaknesses, "Significant cognitive biases detected")
	}

	if len(trajectory.Steps) > 0 {
		failedSteps := 0
		for _, step := range trajectory.Steps {
			if !step.Success {
				failedSteps++
			}
		}
		if failedSteps > len(trajectory.Steps)/4 {
			weaknesses = append(weaknesses, fmt.Sprintf("High failure rate (%.0f%% of steps failed)", float64(failedSteps)/float64(len(trajectory.Steps))*100))
		}
	}

	return weaknesses
}

// generateImprovements creates actionable suggestions
func (r *RetrospectiveAnalyzer) generateImprovements(trajectory *ReasoningTrajectory) []*ImprovementSuggestion {
	improvements := make([]*ImprovementSuggestion, 0)

	if trajectory.Quality == nil {
		return improvements
	}

	q := trajectory.Quality

	// Efficiency improvements
	if q.Efficiency < 0.7 {
		improvements = append(improvements, &ImprovementSuggestion{
			Category:       "efficiency",
			Priority:       1.0 - q.Efficiency,
			Suggestion:     "Reduce unnecessary steps by planning the approach before executing",
			Rationale:      fmt.Sprintf("Current efficiency: %.1f%%. Optimal is around 7-10 steps for most problems.", q.Efficiency*100),
			ExpectedImpact: "Could reduce session time by 20-30%",
		})
	}

	// Coherence improvements
	if q.Coherence < 0.7 {
		improvements = append(improvements, &ImprovementSuggestion{
			Category:       "quality",
			Priority:       1.0 - q.Coherence,
			Suggestion:     "Use 'validate' tool more frequently to check logical consistency",
			Rationale:      fmt.Sprintf("Coherence score: %.1f%%. Regular validation helps maintain logical consistency.", q.Coherence*100),
			ExpectedImpact: "Improve logical consistency by 15-25%",
		})
	}

	// Completeness improvements
	if q.Completeness < 0.7 {
		improvements = append(improvements, &ImprovementSuggestion{
			Category:       "approach",
			Priority:       1.0 - q.Completeness,
			Suggestion:     "Create explicit goal tracking and verify each goal before completing session",
			Rationale:      fmt.Sprintf("Completeness: %.1f%%. Many goals were not achieved.", q.Completeness*100),
			ExpectedImpact: "Achieve 90%+ of stated goals",
		})
	}

	// Innovation improvements
	if q.Innovation < 0.4 {
		improvements = append(improvements, &ImprovementSuggestion{
			Category:       "tools",
			Priority:       0.5,
			Suggestion:     "Try using 'divergent' mode or 'find-analogy' for creative problem-solving",
			Rationale:      fmt.Sprintf("Innovation score: %.1f%%. Consider more advanced reasoning tools.", q.Innovation*100),
			ExpectedImpact: "Discover novel solutions and approaches",
		})
	}

	// Strategy-specific improvements
	if trajectory.Approach != nil {
		if trajectory.Approach.Strategy == "systematic-linear" && trajectory.SuccessScore < 0.6 {
			improvements = append(improvements, &ImprovementSuggestion{
				Category:       "approach",
				Priority:       0.7,
				Suggestion:     "Try 'tree' mode for parallel exploration of multiple approaches",
				Rationale:      "Linear approach didn't work well for this problem type",
				ExpectedImpact: "Explore multiple solutions simultaneously",
			})
		}
	}

	// Sort by priority (descending)
	sort.Slice(improvements, func(i, j int) bool {
		return improvements[i].Priority > improvements[j].Priority
	})

	return improvements
}

// extractLessons identifies key takeaways
func (r *RetrospectiveAnalyzer) extractLessons(trajectory *ReasoningTrajectory) []string {
	lessons := make([]string, 0)

	if trajectory.Outcome == nil {
		return lessons
	}

	// Success lessons
	if trajectory.SuccessScore >= 0.7 {
		if trajectory.Approach != nil && len(trajectory.Approach.ToolSequence) > 0 {
			lessons = append(lessons, fmt.Sprintf("Tool sequence '%v' worked well for %s problems",
				trajectory.Approach.ToolSequence, trajectory.Domain))
		}

		if trajectory.Approach != nil {
			lessons = append(lessons, fmt.Sprintf("Strategy '%s' was effective for this problem type",
				trajectory.Approach.Strategy))
		}
	}

	// Failure lessons
	if trajectory.SuccessScore < 0.4 {
		lessons = append(lessons, fmt.Sprintf("Approach failed - avoid similar strategy for %s domain",
			trajectory.Domain))
	}

	// Unexpected outcomes
	if len(trajectory.Outcome.UnexpectedOutcomes) > 0 {
		lessons = append(lessons, fmt.Sprintf("Watch for unexpected outcomes: %v",
			trajectory.Outcome.UnexpectedOutcomes))
	}

	// Time-based lessons
	if trajectory.Duration > 30*time.Minute {
		lessons = append(lessons, "Consider breaking long sessions into checkpoints for better manageability")
	}

	return lessons
}

// performComparativeAnalysis compares to similar trajectories
func (r *RetrospectiveAnalyzer) performComparativeAnalysis(ctx context.Context, trajectory *ReasoningTrajectory) *ComparativeAnalysis {
	if trajectory.Problem == nil {
		return nil
	}

	// Find similar trajectories
	similar, err := r.store.RetrieveSimilarTrajectories(ctx, trajectory.Problem, 20)
	if err != nil || len(similar) < 3 {
		return nil
	}

	// Calculate percentile rank
	betterThan := 0
	worseThan := 0
	for _, match := range similar {
		if match.Trajectory.ID == trajectory.ID {
			continue
		}
		if trajectory.SuccessScore > match.Trajectory.SuccessScore {
			betterThan++
		} else {
			worseThan++
		}
	}

	total := betterThan + worseThan
	percentile := 0.0
	if total > 0 {
		percentile = (float64(betterThan) / float64(total)) * 100
	}

	// Identify key differences
	differences := r.identifyKeyDifferences(trajectory, similar)

	return &ComparativeAnalysis{
		SimilarTrajectories: len(similar),
		PercentileRank:      percentile,
		BetterThan:          betterThan,
		WorseThan:           worseThan,
		KeyDifferences:      differences,
	}
}

// identifyKeyDifferences finds what made this trajectory different
func (r *RetrospectiveAnalyzer) identifyKeyDifferences(trajectory *ReasoningTrajectory, similar []*TrajectoryMatch) []string {
	differences := make([]string, 0)

	// Compare tool usage
	avgToolCount := 0
	for _, match := range similar {
		if match.Trajectory.Approach != nil {
			avgToolCount += len(match.Trajectory.Approach.ToolSequence)
		}
	}
	if len(similar) > 0 {
		avgToolCount /= len(similar)
	}

	if trajectory.Approach != nil {
		currentToolCount := len(trajectory.Approach.ToolSequence)
		if currentToolCount > int(float64(avgToolCount)*1.3) {
			differences = append(differences, fmt.Sprintf("Used significantly more tools (%d vs avg %d)",
				currentToolCount, avgToolCount))
		} else if currentToolCount < int(float64(avgToolCount)*0.7) {
			differences = append(differences, fmt.Sprintf("Used fewer tools (%d vs avg %d)",
				currentToolCount, avgToolCount))
		}
	}

	// Compare success rate
	avgSuccess := 0.0
	for _, match := range similar {
		avgSuccess += match.Trajectory.SuccessScore
	}
	if len(similar) > 0 {
		avgSuccess /= float64(len(similar))
	}

	if trajectory.SuccessScore > avgSuccess*1.2 {
		differences = append(differences, fmt.Sprintf("Significantly more successful than average (%.0f%% vs %.0f%%)",
			trajectory.SuccessScore*100, avgSuccess*100))
	} else if trajectory.SuccessScore < avgSuccess*0.8 {
		differences = append(differences, fmt.Sprintf("Less successful than average (%.0f%% vs %.0f%%)",
			trajectory.SuccessScore*100, avgSuccess*100))
	}

	return differences
}

// analyzeMetrics provides detailed metric analysis
func (r *RetrospectiveAnalyzer) analyzeMetrics(trajectory *ReasoningTrajectory) *DetailedMetricsAnalysis {
	if trajectory.Quality == nil {
		return &DetailedMetricsAnalysis{}
	}

	q := trajectory.Quality

	return &DetailedMetricsAnalysis{
		EfficiencyAnalysis:   r.analyzeEfficiency(q.Efficiency, len(trajectory.Steps)),
		CoherenceAnalysis:    r.analyzeCoherence(q.Coherence, q.ContradictionCount, q.FallacyCount),
		CompletenessAnalysis: r.analyzeCompleteness(q.Completeness, trajectory),
		InnovationAnalysis:   r.analyzeInnovation(q.Innovation, trajectory),
		ReliabilityAnalysis:  r.analyzeReliability(q.Reliability),
	}
}

// analyzeEfficiency provides efficiency metric analysis
func (r *RetrospectiveAnalyzer) analyzeEfficiency(score float64, stepCount int) *MetricAnalysis {
	assessment := r.scoreToAssessment(score)

	explanation := fmt.Sprintf("Used %d steps. Optimal is around 7-10 steps for most problems.", stepCount)

	suggestions := make([]string, 0)
	if score < 0.7 {
		suggestions = append(suggestions, "Plan the approach before executing")
		suggestions = append(suggestions, "Use 'decompose-problem' to break down complex issues")
		suggestions = append(suggestions, "Avoid unnecessary validation steps")
	}

	return &MetricAnalysis{
		Score:       score,
		Assessment:  assessment,
		Explanation: explanation,
		Suggestions: suggestions,
	}
}

// analyzeCoherence provides coherence metric analysis
func (r *RetrospectiveAnalyzer) analyzeCoherence(score float64, contradictions, fallacies int) *MetricAnalysis {
	assessment := r.scoreToAssessment(score)

	explanation := fmt.Sprintf("Logical consistency score. Found %d contradictions and %d fallacies.",
		contradictions, fallacies)

	suggestions := make([]string, 0)
	if score < 0.7 {
		suggestions = append(suggestions, "Use 'validate' tool regularly")
		suggestions = append(suggestions, "Run 'detect-contradictions' before completing")
		suggestions = append(suggestions, "Use 'detect-fallacies' to check reasoning quality")
	}

	return &MetricAnalysis{
		Score:       score,
		Assessment:  assessment,
		Explanation: explanation,
		Suggestions: suggestions,
	}
}

// analyzeCompleteness provides completeness metric analysis
func (r *RetrospectiveAnalyzer) analyzeCompleteness(score float64, trajectory *ReasoningTrajectory) *MetricAnalysis {
	assessment := r.scoreToAssessment(score)

	goalsAchieved := 0
	goalsTotal := 0
	if trajectory.Problem != nil {
		goalsTotal = len(trajectory.Problem.Goals)
	}
	if trajectory.Outcome != nil {
		goalsAchieved = len(trajectory.Outcome.GoalsAchieved)
	}

	explanation := fmt.Sprintf("Achieved %d of %d goals (%.0f%%).",
		goalsAchieved, goalsTotal, score*100)

	suggestions := make([]string, 0)
	if score < 0.7 {
		suggestions = append(suggestions, "Track goals explicitly throughout the session")
		suggestions = append(suggestions, "Verify each goal before completing")
		suggestions = append(suggestions, "Use 'list-branches' to explore multiple solutions")
	}

	return &MetricAnalysis{
		Score:       score,
		Assessment:  assessment,
		Explanation: explanation,
		Suggestions: suggestions,
	}
}

// analyzeInnovation provides innovation metric analysis
func (r *RetrospectiveAnalyzer) analyzeInnovation(score float64, trajectory *ReasoningTrajectory) *MetricAnalysis {
	assessment := r.scoreToAssessment(score)

	explanation := "Measures use of creative and advanced reasoning tools."

	suggestions := make([]string, 0)
	if score < 0.5 {
		suggestions = append(suggestions, "Try 'divergent' mode for creative thinking")
		suggestions = append(suggestions, "Use 'find-analogy' for cross-domain insights")
		suggestions = append(suggestions, "Explore 'abductive-reasoning' for hypothesis generation")
	}

	return &MetricAnalysis{
		Score:       score,
		Assessment:  assessment,
		Explanation: explanation,
		Suggestions: suggestions,
	}
}

// analyzeReliability provides reliability metric analysis
func (r *RetrospectiveAnalyzer) analyzeReliability(score float64) *MetricAnalysis {
	assessment := r.scoreToAssessment(score)

	explanation := fmt.Sprintf("Confidence in final result: %.0f%%", score*100)

	suggestions := make([]string, 0)
	if score < 0.7 {
		suggestions = append(suggestions, "Validate conclusions before finalizing")
		suggestions = append(suggestions, "Use 'assess-evidence' to verify claims")
		suggestions = append(suggestions, "Run 'self-evaluate' for quality check")
	}

	return &MetricAnalysis{
		Score:       score,
		Assessment:  assessment,
		Explanation: explanation,
		Suggestions: suggestions,
	}
}

// scoreToAssessment converts score to assessment string
func (r *RetrospectiveAnalyzer) scoreToAssessment(score float64) string {
	if score >= 0.8 {
		return "excellent"
	} else if score >= 0.6 {
		return "good"
	} else if score >= 0.4 {
		return "fair"
	}
	return "poor"
}
