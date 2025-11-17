package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetrospectiveAnalyzer_AnalyzeTrajectory(t *testing.T) {
	store := NewEpisodicMemoryStore()
	analyzer := NewRetrospectiveAnalyzer(store)
	ctx := context.Background()

	// Create a sample trajectory with various characteristics
	trajectory := &ReasoningTrajectory{
		ID:        "test-trajectory-1",
		SessionID: "test-session",
		ProblemID: "test-problem",
		StartTime: time.Now().Add(-10 * time.Minute),
		EndTime:   time.Now(),
		Duration:  10 * time.Minute,
		Problem: &ProblemDescription{
			Description: "Test problem",
			Goals:       []string{"Goal 1", "Goal 2", "Goal 3"},
			Complexity:  0.7,
			Domain:      "test-domain",
		},
		Approach: &ApproachDescription{
			Strategy:     "systematic",
			ModesUsed:    []string{"linear", "tree", "divergent"},
			ToolSequence: []string{"think", "analyze", "synthesize", "validate", "think", "prove"},
			KeyDecisions: []string{"Use tree mode for exploration", "Apply validation"},
		},
		Steps: []*ReasoningStep{
			{
				StepNumber: 1,
				Tool:       "think",
				Input:      map[string]interface{}{"mode": "linear"},
				Output:     map[string]interface{}{"thought_id": "t1"},
				Timestamp:  time.Now().Add(-9 * time.Minute),
				Success:    true,
			},
			{
				StepNumber: 2,
				Tool:       "analyze",
				Input:      map[string]interface{}{"thought_id": "t1"},
				Output:     map[string]interface{}{"analysis": "detailed"},
				Timestamp:  time.Now().Add(-8 * time.Minute),
				Success:    true,
			},
			{
				StepNumber: 3,
				Tool:       "synthesize",
				Input:      map[string]interface{}{"inputs": []string{"t1"}},
				Output:     map[string]interface{}{"synthesis": "complete"},
				Timestamp:  time.Now().Add(-7 * time.Minute),
				Success:    true,
			},
			{
				StepNumber: 4,
				Tool:       "validate",
				Input:      map[string]interface{}{"thought_id": "t1"},
				Output:     map[string]interface{}{"valid": true},
				Timestamp:  time.Now().Add(-6 * time.Minute),
				Success:    true,
			},
			{
				StepNumber: 5,
				Tool:       "think",
				Input:      map[string]interface{}{"mode": "tree"},
				Output:     map[string]interface{}{"thought_id": "t2"},
				Timestamp:  time.Now().Add(-5 * time.Minute),
				Success:    true,
			},
			{
				StepNumber: 6,
				Tool:       "prove",
				Input:      map[string]interface{}{"premises": []string{"p1", "p2"}},
				Output:     map[string]interface{}{"proof": "valid"},
				Timestamp:  time.Now().Add(-4 * time.Minute),
				Success:    true,
			},
		},
		Outcome: &OutcomeDescription{
			Status:             "success",
			Solution:           "Problem solved successfully",
			GoalsAchieved:      []string{"Goal 1", "Goal 2"},
			GoalsFailed:        []string{"Goal 3"},
			UnexpectedOutcomes: []string{"Found alternative approach"},
			Confidence:         0.85,
		},
		Quality: &QualityMetrics{
			OverallQuality:     0.8,
			Efficiency:         0.8,
			Coherence:          0.9,
			Completeness:       0.67, // 2 of 3 goals achieved
			Innovation:         0.7,
			Reliability:        0.85,
			BiasScore:          0.1,
			FallacyCount:       0,
			ContradictionCount: 0,
			SelfEvalScore:      0.8,
		},
		SuccessScore: 0.8,
		Domain:       "test-domain",
		Complexity:   0.7,
		Tags:         []string{"problem-solving", "validation", "proof"},
	}

	// Store the trajectory first
	err := store.StoreTrajectory(ctx, trajectory)
	require.NoError(t, err)

	// Store some similar trajectories for comparative analysis
	similarTraj1 := *trajectory
	similarTraj1.ID = "similar-1"
	similarTraj1.SessionID = "similar-1"
	similarTraj1.Quality.Efficiency = 0.6
	similarTraj1.Duration = 15 * time.Minute
	store.StoreTrajectory(ctx, &similarTraj1)

	similarTraj2 := *trajectory
	similarTraj2.ID = "similar-2"
	similarTraj2.SessionID = "similar-2"
	similarTraj2.Quality.Efficiency = 0.9
	similarTraj2.Duration = 5 * time.Minute
	store.StoreTrajectory(ctx, &similarTraj2)

	// Analyze the trajectory
	report, err := analyzer.AnalyzeTrajectory(ctx, trajectory.ID)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify summary
	assert.Equal(t, trajectory.ID, report.TrajectoryID)
	assert.Equal(t, "test-session", report.SessionID)
	assert.Equal(t, "excellent", report.Summary.OverallAssessment) // 0.8 meets threshold for excellent
	assert.Equal(t, 0.8, report.Summary.SuccessScore)
	assert.Equal(t, 0.8, report.Summary.QualityScore)
	assert.Equal(t, "10m0s", report.Summary.Duration)
	assert.Equal(t, "systematic", report.Summary.StrategyUsed)

	// Verify strengths
	assert.NotEmpty(t, report.Strengths)
	foundCoherenceStrength := false
	for _, strength := range report.Strengths {
		if containsStr(strength, "coherence") {
			foundCoherenceStrength = true
			break
		}
	}
	assert.True(t, foundCoherenceStrength, "Should identify coherence as a strength")

	// Verify weaknesses
	// With these metrics, we might not have significant weaknesses
	// Completeness at 67% may not trigger weakness (threshold is <50%)
	// but we should have improvements suggested

	// Verify improvements
	assert.NotEmpty(t, report.Improvements)
	foundGoalImprovement := false
	for _, improvement := range report.Improvements {
		if improvement.Category == "approach" || improvement.Category == "completeness" {
			foundGoalImprovement = true
			assert.Contains(t, improvement.Suggestion, "goal")
			break
		}
	}
	assert.True(t, foundGoalImprovement, "Should suggest improvements for goal achievement")

	// Verify lessons learned
	assert.NotEmpty(t, report.LessonsLearned)

	// Verify comparative analysis exists (we have similar trajectories)
	assert.NotNil(t, report.ComparativeAnalysis)
	assert.GreaterOrEqual(t, report.ComparativeAnalysis.SimilarTrajectories, 2)

	// Verify detailed metrics
	assert.NotNil(t, report.DetailedMetrics)
	assert.NotNil(t, report.DetailedMetrics.EfficiencyAnalysis)
	assert.NotNil(t, report.DetailedMetrics.CoherenceAnalysis)
	assert.NotNil(t, report.DetailedMetrics.CompletenessAnalysis)
	assert.NotNil(t, report.DetailedMetrics.InnovationAnalysis)
	assert.NotNil(t, report.DetailedMetrics.ReliabilityAnalysis)
}

func TestRetrospectiveAnalyzer_EdgeCases(t *testing.T) {
	store := NewEpisodicMemoryStore()
	analyzer := NewRetrospectiveAnalyzer(store)
	ctx := context.Background()

	t.Run("empty trajectory", func(t *testing.T) {
		emptyTraj := &ReasoningTrajectory{
			ID:        "empty-traj",
			SessionID: "empty",
			ProblemID: "test",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  0,
			Problem: &ProblemDescription{
				Description: "Test",
			},
			Approach: &ApproachDescription{
				Strategy: "unknown",
			},
			Steps:   []*ReasoningStep{},
			Outcome: &OutcomeDescription{Status: "failure"},
			Quality: &QualityMetrics{
				OverallQuality: 0.0,
				Efficiency:     0.0,
				Coherence:      0.0,
				Completeness:   0.0,
				Innovation:     0.0,
				Reliability:    0.0,
			},
			SuccessScore: 0.0,
		}

		err := store.StoreTrajectory(ctx, emptyTraj)
		require.NoError(t, err)

		report, err := analyzer.AnalyzeTrajectory(ctx, emptyTraj.ID)
		require.NoError(t, err)
		require.NotNil(t, report)
		assert.Equal(t, "poor", report.Summary.OverallAssessment)
	})

	t.Run("perfect trajectory", func(t *testing.T) {
		perfectTraj := &ReasoningTrajectory{
			ID:        "perfect-traj",
			SessionID: "perfect",
			ProblemID: "test",
			StartTime: time.Now().Add(-5 * time.Minute),
			EndTime:   time.Now(),
			Duration:  5 * time.Minute,
			Problem: &ProblemDescription{
				Description: "Test",
				Goals:       []string{"Goal 1", "Goal 2"},
			},
			Approach: &ApproachDescription{
				Strategy:     "optimal",
				ModesUsed:    []string{"auto"},
				ToolSequence: []string{"think", "validate"},
			},
			Steps: []*ReasoningStep{
				{StepNumber: 1, Tool: "think", Success: true},
				{StepNumber: 2, Tool: "validate", Success: true},
			},
			Outcome: &OutcomeDescription{
				Status:        "success",
				GoalsAchieved: []string{"Goal 1", "Goal 2"},
				GoalsFailed:   []string{},
				Confidence:    1.0,
			},
			Quality: &QualityMetrics{
				OverallQuality:     1.0,
				Efficiency:         1.0,
				Coherence:          1.0,
				Completeness:       1.0,
				Innovation:         1.0,
				Reliability:        1.0,
				BiasScore:          0.0,
				FallacyCount:       0,
				ContradictionCount: 0,
				SelfEvalScore:      1.0,
			},
			SuccessScore: 1.0,
		}

		err := store.StoreTrajectory(ctx, perfectTraj)
		require.NoError(t, err)

		report, err := analyzer.AnalyzeTrajectory(ctx, perfectTraj.ID)
		require.NoError(t, err)
		require.NotNil(t, report)
		assert.Equal(t, "excellent", report.Summary.OverallAssessment)
		assert.NotEmpty(t, report.Strengths)
		assert.Empty(t, report.Weaknesses) // No weaknesses in perfect trajectory
	})

	t.Run("non-existent trajectory", func(t *testing.T) {
		report, err := analyzer.AnalyzeTrajectory(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Nil(t, report)
		assert.Contains(t, err.Error(), "trajectory not found")
	})
}

func TestRetrospectiveAnalyzer_MetricAnalysis(t *testing.T) {
	store := NewEpisodicMemoryStore()
	analyzer := NewRetrospectiveAnalyzer(store)
	ctx := context.Background()

	testCases := []struct {
		name               string
		quality            *QualityMetrics
		expectedAssessment string
	}{
		{
			name: "excellent metrics",
			quality: &QualityMetrics{
				OverallQuality:     0.93,
				Efficiency:         0.95,
				Coherence:          0.92,
				Completeness:       0.98,
				Innovation:         0.88,
				Reliability:        0.96,
				BiasScore:          0.05,
				FallacyCount:       0,
				ContradictionCount: 0,
				SelfEvalScore:      0.9,
			},
			expectedAssessment: "excellent",
		},
		{
			name: "good metrics",
			quality: &QualityMetrics{
				OverallQuality:     0.78,
				Efficiency:         0.75,
				Coherence:          0.78,
				Completeness:       0.82,
				Innovation:         0.70,
				Reliability:        0.80,
				BiasScore:          0.2,
				FallacyCount:       1,
				ContradictionCount: 0,
				SelfEvalScore:      0.7,
			},
			expectedAssessment: "good",
		},
		{
			name: "fair metrics",
			quality: &QualityMetrics{
				OverallQuality:     0.57,
				Efficiency:         0.55,
				Coherence:          0.60,
				Completeness:       0.58,
				Innovation:         0.50,
				Reliability:        0.62,
				BiasScore:          0.35,
				FallacyCount:       2,
				ContradictionCount: 1,
				SelfEvalScore:      0.5,
			},
			expectedAssessment: "fair",
		},
		{
			name: "poor metrics",
			quality: &QualityMetrics{
				OverallQuality:     0.30,
				Efficiency:         0.30,
				Coherence:          0.35,
				Completeness:       0.25,
				Innovation:         0.20,
				Reliability:        0.38,
				BiasScore:          0.6,
				FallacyCount:       5,
				ContradictionCount: 3,
				SelfEvalScore:      0.3,
			},
			expectedAssessment: "poor",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trajectory := &ReasoningTrajectory{
				ID:        tc.name + "-id",
				SessionID: tc.name,
				ProblemID: "test",
				StartTime: time.Now().Add(-10 * time.Minute),
				EndTime:   time.Now(),
				Duration:  10 * time.Minute,
				Problem: &ProblemDescription{
					Description: "Test problem",
				},
				Approach: &ApproachDescription{
					Strategy:     "test",
					ToolSequence: []string{"think"},
				},
				Steps: []*ReasoningStep{
					{StepNumber: 1, Tool: "think", Success: true},
				},
				Outcome: &OutcomeDescription{
					Status: "success",
				},
				Quality:      tc.quality,
				SuccessScore: tc.quality.OverallQuality,
			}

			err := store.StoreTrajectory(ctx, trajectory)
			require.NoError(t, err)

			report, err := analyzer.AnalyzeTrajectory(ctx, trajectory.ID)
			require.NoError(t, err)
			require.NotNil(t, report)

			// Check overall assessment
			assert.Equal(t, tc.expectedAssessment, report.Summary.OverallAssessment)

			// Check detailed metrics
			assert.NotNil(t, report.DetailedMetrics)
			assert.Equal(t, tc.quality.Efficiency, report.DetailedMetrics.EfficiencyAnalysis.Score)
			assert.Equal(t, tc.quality.Coherence, report.DetailedMetrics.CoherenceAnalysis.Score)
			assert.Equal(t, tc.quality.Completeness, report.DetailedMetrics.CompletenessAnalysis.Score)
			assert.Equal(t, tc.quality.Innovation, report.DetailedMetrics.InnovationAnalysis.Score)
			assert.Equal(t, tc.quality.Reliability, report.DetailedMetrics.ReliabilityAnalysis.Score)
		})
	}
}

func TestRetrospectiveAnalyzer_ComparativeAnalysis(t *testing.T) {
	store := NewEpisodicMemoryStore()
	analyzer := NewRetrospectiveAnalyzer(store)
	ctx := context.Background()

	// Store multiple trajectories for comparison
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		traj := &ReasoningTrajectory{
			ID:        fmt.Sprintf("session-%d", i),
			SessionID: fmt.Sprintf("session-%d", i),
			ProblemID: fmt.Sprintf("problem-%d", i%3), // 3 different problem types
			StartTime: baseTime.Add(time.Duration(-i) * time.Hour),
			EndTime:   baseTime.Add(time.Duration(-i)*time.Hour + 10*time.Minute),
			Duration:  10 * time.Minute,
			Problem: &ProblemDescription{
				Description: fmt.Sprintf("Problem %d", i),
				Domain:      "test-domain",
				Complexity:  float64(i%5) / 5.0, // Varying complexity
			},
			Approach: &ApproachDescription{
				Strategy:     []string{"systematic", "exploration", "creative"}[i%3],
				ToolSequence: []string{"think", "validate"},
			},
			Steps: []*ReasoningStep{
				{StepNumber: 1, Tool: "think", Success: true},
			},
			Outcome: &OutcomeDescription{
				Status:     []string{"success", "partial", "failure"}[i%3],
				Confidence: float64(9-i) / 10.0,
			},
			Quality: &QualityMetrics{
				OverallQuality:     float64(10-i) / 10.0,
				Efficiency:         float64(10-i) / 10.0,
				Coherence:          0.8,
				Completeness:       float64(i+5) / 15.0,
				Innovation:         float64(i%4) / 4.0,
				Reliability:        0.7,
				BiasScore:          0.1,
				FallacyCount:       0,
				ContradictionCount: 0,
				SelfEvalScore:      0.7,
			},
			SuccessScore: float64(10-i) / 10.0,
			Domain:       "test-domain",
			Tags:         []string{"test"},
		}
		err := store.StoreTrajectory(ctx, traj)
		require.NoError(t, err)
	}

	// Analyze a new trajectory
	testTraj := &ReasoningTrajectory{
		ID:        "test-comparative",
		SessionID: "test-session",
		ProblemID: "test-problem",
		StartTime: baseTime.Add(-30 * time.Minute),
		EndTime:   baseTime.Add(-20 * time.Minute),
		Duration:  10 * time.Minute,
		Problem: &ProblemDescription{
			Description: "Test problem for comparison",
			Domain:      "test-domain",
			Complexity:  0.5,
		},
		Approach: &ApproachDescription{
			Strategy:     "systematic",
			ToolSequence: []string{"think", "validate"},
		},
		Steps: []*ReasoningStep{
			{StepNumber: 1, Tool: "think", Success: true},
			{StepNumber: 2, Tool: "validate", Success: true},
		},
		Outcome: &OutcomeDescription{
			Status:     "success",
			Confidence: 0.75,
		},
		Quality: &QualityMetrics{
			OverallQuality:     0.72,
			Efficiency:         0.75,
			Coherence:          0.8,
			Completeness:       0.7,
			Innovation:         0.5,
			Reliability:        0.7,
			BiasScore:          0.15,
			FallacyCount:       0,
			ContradictionCount: 0,
			SelfEvalScore:      0.7,
		},
		SuccessScore: 0.75,
		Domain:       "test-domain",
		Tags:         []string{"test"},
	}

	err := store.StoreTrajectory(ctx, testTraj)
	require.NoError(t, err)

	report, err := analyzer.AnalyzeTrajectory(ctx, testTraj.ID)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.NotNil(t, report.ComparativeAnalysis)

	// Verify comparative analysis
	assert.Greater(t, report.ComparativeAnalysis.SimilarTrajectories, 0)
	assert.GreaterOrEqual(t, report.ComparativeAnalysis.PercentileRank, 0.0)
	assert.LessOrEqual(t, report.ComparativeAnalysis.PercentileRank, 100.0)
	assert.NotEmpty(t, report.ComparativeAnalysis.KeyDifferences)

	t.Logf("Percentile rank: %.1f%%", report.ComparativeAnalysis.PercentileRank)
	t.Logf("Better than %d trajectories", report.ComparativeAnalysis.BetterThan)
	t.Logf("Worse than %d trajectories", report.ComparativeAnalysis.WorseThan)
}

func TestRetrospectiveAnalyzer_StrategicInsights(t *testing.T) {
	store := NewEpisodicMemoryStore()
	analyzer := NewRetrospectiveAnalyzer(store)
	ctx := context.Background()

	// Test trajectory with specific patterns
	trajectory := &ReasoningTrajectory{
		ID:        "strategic-test",
		SessionID: "strategic-test",
		ProblemID: "test-problem",
		StartTime: time.Now().Add(-30 * time.Minute),
		EndTime:   time.Now(),
		Duration:  30 * time.Minute,
		Problem: &ProblemDescription{
			Description: "Complex problem requiring multiple approaches",
			Goals:       []string{"Understand", "Analyze", "Solve", "Validate"},
			Complexity:  0.9,
		},
		Approach: &ApproachDescription{
			Strategy:  "multi-modal",
			ModesUsed: []string{"linear", "tree", "divergent", "reflection"},
			ToolSequence: []string{
				"think", "think", "branch", "synthesize",
				"validate", "detect-contradictions", "think",
				"prove", "make-decision",
			},
			KeyDecisions: []string{
				"Switch to tree mode for exploration",
				"Use validation to ensure correctness",
				"Apply formal proof for critical claims",
			},
		},
		Steps: []*ReasoningStep{
			{StepNumber: 1, Tool: "think", Success: true},
			{StepNumber: 2, Tool: "think", Success: true},
			{StepNumber: 3, Tool: "branch", Success: true},
			{StepNumber: 4, Tool: "synthesize", Success: true},
			{StepNumber: 5, Tool: "validate", Success: true},
			{StepNumber: 6, Tool: "detect-contradictions", Success: true},
			{StepNumber: 7, Tool: "think", Success: true},
			{StepNumber: 8, Tool: "prove", Success: true},
			{StepNumber: 9, Tool: "make-decision", Success: true},
		},
		Outcome: &OutcomeDescription{
			Status:             "success",
			GoalsAchieved:      []string{"Understand", "Analyze", "Solve"},
			GoalsFailed:        []string{"Validate"},
			UnexpectedOutcomes: []string{"Alternative solution path", "Performance optimization"},
			Confidence:         0.88,
		},
		Quality: &QualityMetrics{
			OverallQuality:     0.81,
			Efficiency:         0.6, // Took longer due to exploration
			Coherence:          0.95,
			Completeness:       0.75,
			Innovation:         0.85,
			Reliability:        0.90,
			BiasScore:          0.1,
			FallacyCount:       0,
			ContradictionCount: 0,
			SelfEvalScore:      0.85,
		},
		SuccessScore: 0.85,
	}

	err := store.StoreTrajectory(ctx, trajectory)
	require.NoError(t, err)

	report, err := analyzer.AnalyzeTrajectory(ctx, trajectory.ID)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify strategic insights are captured
	assert.NotEmpty(t, report.LessonsLearned)

	// Check for strategy-related lessons
	foundStrategyLesson := false
	for _, lesson := range report.LessonsLearned {
		if containsStr(lesson, "multi-modal") || containsStr(lesson, "effective") {
			foundStrategyLesson = true
			break
		}
	}
	assert.True(t, foundStrategyLesson, "Should identify lessons about strategy usage")

	// Check for innovation recognition
	foundInnovationStrength := false
	for _, strength := range report.Strengths {
		if containsStr(strength, "innovation") || containsStr(strength, "creative") {
			foundInnovationStrength = true
			break
		}
	}
	assert.True(t, foundInnovationStrength, "Should recognize innovation as strength")

	// Check for efficiency improvement suggestion
	foundEfficiencyImprovement := false
	for _, improvement := range report.Improvements {
		if improvement.Category == "efficiency" {
			foundEfficiencyImprovement = true
			assert.Greater(t, improvement.Priority, 0.0)
			break
		}
	}
	assert.True(t, foundEfficiencyImprovement, "Should suggest efficiency improvements")
}

// Helper function to check if string contains substring
func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr ||
		 (len(s) > len(substr) &&
		  (s[:len(substr)] == substr ||
		   s[len(s)-len(substr):] == substr ||
		   findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}