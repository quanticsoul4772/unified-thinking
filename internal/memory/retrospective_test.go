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

	// Create a sample trajectory with various characteristics
	trajectory := &ReasoningTrajectory{
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
			},
			{
				StepNumber: 2,
				Tool:       "analyze",
				Input:      map[string]interface{}{"thought_id": "t1"},
				Output:     map[string]interface{}{"analysis": "detailed"},
				Timestamp:  time.Now().Add(-8 * time.Minute),
			},
			{
				StepNumber: 3,
				Tool:       "synthesize",
				Input:      map[string]interface{}{"inputs": []string{"t1"}},
				Output:     map[string]interface{}{"synthesis": "complete"},
				Timestamp:  time.Now().Add(-7 * time.Minute),
			},
			{
				StepNumber: 4,
				Tool:       "validate",
				Input:      map[string]interface{}{"thought_id": "t1"},
				Output:     map[string]interface{}{"valid": true},
				Timestamp:  time.Now().Add(-6 * time.Minute),
			},
			{
				StepNumber: 5,
				Tool:       "think",
				Input:      map[string]interface{}{"mode": "tree"},
				Output:     map[string]interface{}{"thought_id": "t2"},
				Timestamp:  time.Now().Add(-5 * time.Minute),
			},
			{
				StepNumber: 6,
				Tool:       "prove",
				Input:      map[string]interface{}{"premises": []string{"p1", "p2"}},
				Output:     map[string]interface{}{"proof": "valid"},
				Timestamp:  time.Now().Add(-4 * time.Minute),
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
			Efficiency:   0.8,
			Coherence:    0.9,
			Completeness: 0.67, // 2 of 3 goals achieved
			Innovation:   0.7,
			Reliability:  0.85,
		},
		SuccessScore: 0.8,
		Domain:       "test-domain",
		Complexity:   0.7,
		Tags:         []string{"problem-solving", "validation", "proof"},
	}

	// Store some similar trajectories for comparative analysis
	similarTraj1 := *trajectory
	similarTraj1.SessionID = "similar-1"
	similarTraj1.Quality.Efficiency = 0.6
	similarTraj1.Duration = 15 * time.Minute
	store.StoreTrajectory(context.Background(), &similarTraj1)

	similarTraj2 := *trajectory
	similarTraj2.SessionID = "similar-2"
	similarTraj2.Quality.Efficiency = 0.9
	similarTraj2.Duration = 5 * time.Minute
	store.StoreTrajectory(context.Background(), &similarTraj2)

	// Analyze the trajectory
	report, err := analyzer.AnalyzeTrajectory(context.Background(), trajectory)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify summary
	assert.Equal(t, "test-session", report.SessionID)
	assert.Equal(t, "success", report.Summary.OverallStatus)
	assert.Equal(t, 0.8, report.Summary.SuccessScore)
	assert.Equal(t, 10*time.Minute, report.Summary.Duration)
	assert.Equal(t, "systematic", report.Summary.Strategy)
	assert.Contains(t, report.Summary.KeyOutcomes, "2/3 goals achieved")

	// Verify strengths
	assert.NotEmpty(t, report.Strengths)
	foundCoherenceStrength := false
	for _, strength := range report.Strengths {
		if strength.Category == "coherence" {
			foundCoherenceStrength = true
			assert.Equal(t, 0.9, strength.Score)
			assert.Contains(t, strength.Description, "Logical consistency")
			break
		}
	}
	assert.True(t, foundCoherenceStrength, "Should identify coherence as a strength")

	// Verify weaknesses
	assert.NotEmpty(t, report.Weaknesses)
	foundCompletenessWeakness := false
	for _, weakness := range report.Weaknesses {
		if weakness.Category == "completeness" {
			foundCompletenessWeakness = true
			assert.Contains(t, weakness.RootCause, "goal achievement")
			break
		}
	}
	assert.True(t, foundCompletenessWeakness, "Should identify completeness as a weakness")

	// Verify improvements
	assert.NotEmpty(t, report.Improvements)
	foundGoalImprovement := false
	for _, improvement := range report.Improvements {
		if improvement.Category == "completeness" {
			foundGoalImprovement = true
			assert.Contains(t, improvement.Suggestion, "goal")
			break
		}
	}
	assert.True(t, foundGoalImprovement, "Should suggest improvements for goal achievement")

	// Verify lessons learned
	assert.NotEmpty(t, report.LessonsLearned)

	// Verify comparative analysis
	assert.NotNil(t, report.ComparativeAnalysis)
	assert.GreaterOrEqual(t, report.ComparativeAnalysis.SimilarSessionsCount, 2)
	assert.NotEmpty(t, report.ComparativeAnalysis.PercentileRanking)

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

	t.Run("empty trajectory", func(t *testing.T) {
		emptyTraj := &ReasoningTrajectory{
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
			Quality: &QualityMetrics{},
		}

		report, err := analyzer.AnalyzeTrajectory(context.Background(), emptyTraj)
		require.NoError(t, err)
		require.NotNil(t, report)
		assert.Equal(t, "poor", report.Summary.QualityScore)
	})

	t.Run("perfect trajectory", func(t *testing.T) {
		perfectTraj := &ReasoningTrajectory{
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
				{StepNumber: 1, Tool: "think"},
				{StepNumber: 2, Tool: "validate"},
			},
			Outcome: &OutcomeDescription{
				Status:        "success",
				GoalsAchieved: []string{"Goal 1", "Goal 2"},
				GoalsFailed:   []string{},
				Confidence:    1.0,
			},
			Quality: &QualityMetrics{
				Efficiency:   1.0,
				Coherence:    1.0,
				Completeness: 1.0,
				Innovation:   1.0,
				Reliability:  1.0,
			},
			SuccessScore: 1.0,
		}

		report, err := analyzer.AnalyzeTrajectory(context.Background(), perfectTraj)
		require.NoError(t, err)
		require.NotNil(t, report)
		assert.Equal(t, "excellent", report.Summary.QualityScore)
		assert.NotEmpty(t, report.Strengths)
		assert.Empty(t, report.Weaknesses) // No weaknesses in perfect trajectory
	})

	t.Run("nil trajectory", func(t *testing.T) {
		report, err := analyzer.AnalyzeTrajectory(context.Background(), nil)
		assert.Error(t, err)
		assert.Nil(t, report)
	})
}

func TestRetrospectiveAnalyzer_MetricAnalysis(t *testing.T) {
	store := NewEpisodicMemoryStore()
	analyzer := NewRetrospectiveAnalyzer(store)

	testCases := []struct {
		name           string
		quality        *QualityMetrics
		expectedAssess string
	}{
		{
			name: "excellent metrics",
			quality: &QualityMetrics{
				Efficiency:   0.95,
				Coherence:    0.92,
				Completeness: 0.98,
				Innovation:   0.88,
				Reliability:  0.96,
			},
			expectedAssess: "excellent",
		},
		{
			name: "good metrics",
			quality: &QualityMetrics{
				Efficiency:   0.75,
				Coherence:    0.78,
				Completeness: 0.82,
				Innovation:   0.70,
				Reliability:  0.80,
			},
			expectedAssess: "good",
		},
		{
			name: "fair metrics",
			quality: &QualityMetrics{
				Efficiency:   0.55,
				Coherence:    0.60,
				Completeness: 0.58,
				Innovation:   0.50,
				Reliability:  0.62,
			},
			expectedAssess: "fair",
		},
		{
			name: "poor metrics",
			quality: &QualityMetrics{
				Efficiency:   0.30,
				Coherence:    0.35,
				Completeness: 0.25,
				Innovation:   0.20,
				Reliability:  0.38,
			},
			expectedAssess: "poor",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trajectory := &ReasoningTrajectory{
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
					{StepNumber: 1, Tool: "think"},
				},
				Outcome: &OutcomeDescription{
					Status: "success",
				},
				Quality: tc.quality,
			}

			report, err := analyzer.AnalyzeTrajectory(context.Background(), trajectory)
			require.NoError(t, err)
			require.NotNil(t, report)

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

	// Store multiple trajectories for comparison
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		traj := &ReasoningTrajectory{
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
				{StepNumber: 1, Tool: "think"},
			},
			Outcome: &OutcomeDescription{
				Status:     []string{"success", "partial", "failure"}[i%3],
				Confidence: float64(9-i) / 10.0,
			},
			Quality: &QualityMetrics{
				Efficiency:   float64(10-i) / 10.0,
				Coherence:    0.8,
				Completeness: float64(i+5) / 15.0,
				Innovation:   float64(i%4) / 4.0,
				Reliability:  0.7,
			},
			SuccessScore: float64(10-i) / 10.0,
			Domain:       "test-domain",
			Tags:         []string{"test"},
		}
		err := store.StoreTrajectory(context.Background(), traj)
		require.NoError(t, err)
	}

	// Analyze a new trajectory
	testTraj := &ReasoningTrajectory{
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
			{StepNumber: 1, Tool: "think"},
			{StepNumber: 2, Tool: "validate"},
		},
		Outcome: &OutcomeDescription{
			Status:     "success",
			Confidence: 0.75,
		},
		Quality: &QualityMetrics{
			Efficiency:   0.75,
			Coherence:    0.8,
			Completeness: 0.7,
			Innovation:   0.5,
			Reliability:  0.7,
		},
		SuccessScore: 0.75,
		Domain:       "test-domain",
		Tags:         []string{"test"},
	}

	report, err := analyzer.AnalyzeTrajectory(context.Background(), testTraj)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.NotNil(t, report.ComparativeAnalysis)

	// Verify comparative analysis
	assert.Greater(t, report.ComparativeAnalysis.SimilarSessionsCount, 0)
	assert.NotEmpty(t, report.ComparativeAnalysis.PercentileRanking)
	assert.NotEmpty(t, report.ComparativeAnalysis.KeyDifferences)

	// Check percentile rankings
	for metric, percentile := range report.ComparativeAnalysis.PercentileRanking {
		assert.GreaterOrEqual(t, percentile, 0.0)
		assert.LessOrEqual(t, percentile, 100.0)
		t.Logf("%s percentile: %.1f%%", metric, percentile)
	}
}

func TestRetrospectiveAnalyzer_StrategicInsights(t *testing.T) {
	store := NewEpisodicMemoryStore()
	analyzer := NewRetrospectiveAnalyzer(store)

	// Test trajectory with specific patterns
	trajectory := &ReasoningTrajectory{
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
		Steps: make([]*ReasoningStep, 9),
		Outcome: &OutcomeDescription{
			Status:          "success",
			GoalsAchieved:   []string{"Understand", "Analyze", "Solve"},
			GoalsFailed:     []string{"Validate"},
			UnexpectedFinds: []string{"Alternative solution path", "Performance optimization"},
			Confidence:      0.88,
		},
		Quality: &QualityMetrics{
			Efficiency:   0.6, // Took longer due to exploration
			Coherence:    0.95,
			Completeness: 0.75,
			Innovation:   0.85,
			Reliability:  0.90,
		},
		SuccessScore: 0.85,
	}

	report, err := analyzer.AnalyzeTrajectory(context.Background(), trajectory)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify strategic insights are captured
	assert.NotEmpty(t, report.LessonsLearned)

	// Check for mode-related lessons
	foundModeLesson := false
	for _, lesson := range report.LessonsLearned {
		if lesson.Category == "approach" {
			foundModeLesson = true
			assert.Contains(t, lesson.Takeaway, "mode")
			break
		}
	}
	assert.True(t, foundModeLesson, "Should identify lessons about mode usage")

	// Check for innovation recognition
	foundInnovationStrength := false
	for _, strength := range report.Strengths {
		if strength.Category == "innovation" {
			foundInnovationStrength = true
			assert.Greater(t, strength.Score, 0.8)
			break
		}
	}
	assert.True(t, foundInnovationStrength, "Should recognize innovation as strength")

	// Check for efficiency improvement suggestion
	foundEfficiencyImprovement := false
	for _, improvement := range report.Improvements {
		if improvement.Category == "efficiency" {
			foundEfficiencyImprovement = true
			assert.Equal(t, "high", improvement.Priority)
			break
		}
	}
	assert.True(t, foundEfficiencyImprovement, "Should suggest efficiency improvements")
}