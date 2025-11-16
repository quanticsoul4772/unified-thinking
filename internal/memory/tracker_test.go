package memory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewSessionTracker(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker := NewSessionTracker(store)

	if tracker == nil {
		t.Fatal("NewSessionTracker returned nil")
	}

	if tracker.store != store {
		t.Error("Store not set correctly")
	}

	if tracker.activeSessions == nil {
		t.Error("activeSessions map not initialized")
	}
}

func TestStartSession(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker := NewSessionTracker(store)
	ctx := context.Background()

	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "testing",
		Goals:       []string{"goal1", "goal2"},
		Complexity:  0.5,
	}

	err := tracker.StartSession(ctx, "session_001", problem)
	if err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	// Verify session was created
	session, exists := tracker.GetActiveSession("session_001")
	if !exists {
		t.Fatal("Session not found after StartSession")
	}

	if session.SessionID != "session_001" {
		t.Errorf("Expected session_id session_001, got %s", session.SessionID)
	}

	if session.Problem != problem {
		t.Error("Problem not set correctly")
	}

	if session.Domain != "testing" {
		t.Errorf("Expected domain testing, got %s", session.Domain)
	}

	if len(session.Steps) != 0 {
		t.Errorf("Expected 0 steps initially, got %d", len(session.Steps))
	}

	if session.CurrentStep != 0 {
		t.Errorf("Expected CurrentStep 0, got %d", session.CurrentStep)
	}
}

func TestRecordStep(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker := NewSessionTracker(store)
	ctx := context.Background()

	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "testing",
	}

	tracker.StartSession(ctx, "session_001", problem)

	step := &ReasoningStep{
		Tool:       "think",
		Mode:       "linear",
		ThoughtID:  "thought_001",
		BranchID:   "branch_001",
		Confidence: 0.8,
		Duration:   2 * time.Second,
		Success:    true,
	}

	err := tracker.RecordStep(ctx, "session_001", step)
	if err != nil {
		t.Fatalf("RecordStep failed: %v", err)
	}

	// Verify step was recorded
	session, _ := tracker.GetActiveSession("session_001")

	if len(session.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(session.Steps))
	}

	recordedStep := session.Steps[0]
	if recordedStep.StepNumber != 1 {
		t.Errorf("Expected step number 1, got %d", recordedStep.StepNumber)
	}

	if recordedStep.Tool != "think" {
		t.Errorf("Expected tool 'think', got %s", recordedStep.Tool)
	}

	// Verify tracking maps updated
	if !session.ModesUsed["linear"] {
		t.Error("Mode 'linear' not tracked")
	}

	if session.ToolsUsed["think"] != 1 {
		t.Errorf("Expected tool count 1 for 'think', got %d", session.ToolsUsed["think"])
	}

	if !session.BranchesUsed["branch_001"] {
		t.Error("Branch 'branch_001' not tracked")
	}
}

func TestRecordMultipleSteps(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker := NewSessionTracker(store)
	ctx := context.Background()

	problem := &ProblemDescription{
		Description: "Multi-step problem",
		Domain:      "testing",
	}

	tracker.StartSession(ctx, "session_001", problem)

	// Record 5 steps
	for i := 0; i < 5; i++ {
		step := &ReasoningStep{
			Tool:       "think",
			Mode:       "linear",
			Confidence: 0.7,
			Success:    true,
		}
		tracker.RecordStep(ctx, "session_001", step)
	}

	session, _ := tracker.GetActiveSession("session_001")

	if len(session.Steps) != 5 {
		t.Errorf("Expected 5 steps, got %d", len(session.Steps))
	}

	if session.CurrentStep != 5 {
		t.Errorf("Expected CurrentStep 5, got %d", session.CurrentStep)
	}

	// Verify step numbers are sequential
	for i, step := range session.Steps {
		expectedNum := i + 1
		if step.StepNumber != expectedNum {
			t.Errorf("Step %d has number %d, expected %d", i, step.StepNumber, expectedNum)
		}
	}

	// Verify tool usage count
	if session.ToolsUsed["think"] != 5 {
		t.Errorf("Expected tool count 5, got %d", session.ToolsUsed["think"])
	}
}

func TestCompleteSession(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker := NewSessionTracker(store)
	ctx := context.Background()

	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "testing",
		Goals:       []string{"goal1", "goal2"},
		Complexity:  0.5,
	}

	tracker.StartSession(ctx, "session_001", problem)

	// Add some steps
	for i := 0; i < 3; i++ {
		step := &ReasoningStep{
			Tool:       "think",
			Mode:       "linear",
			Confidence: 0.8,
			Success:    true,
		}
		tracker.RecordStep(ctx, "session_001", step)
	}

	outcome := &OutcomeDescription{
		Status:        "success",
		GoalsAchieved: []string{"goal1", "goal2"},
		GoalsFailed:   []string{},
		Solution:      "Problem solved",
		Confidence:    0.85,
	}

	trajectory, err := tracker.CompleteSession(ctx, "session_001", outcome)
	if err != nil {
		t.Fatalf("CompleteSession failed: %v", err)
	}

	// Verify trajectory was built
	if trajectory == nil {
		t.Fatal("CompleteSession returned nil trajectory")
	}

	if trajectory.SessionID != "session_001" {
		t.Errorf("Expected session_id session_001, got %s", trajectory.SessionID)
	}

	if trajectory.Problem != problem {
		t.Error("Problem not set correctly in trajectory")
	}

	if trajectory.Outcome != outcome {
		t.Error("Outcome not set correctly in trajectory")
	}

	if len(trajectory.Steps) != 3 {
		t.Errorf("Expected 3 steps in trajectory, got %d", len(trajectory.Steps))
	}

	if trajectory.Approach == nil {
		t.Fatal("Approach not generated")
	}

	if trajectory.Quality == nil {
		t.Fatal("Quality metrics not calculated")
	}

	if trajectory.SuccessScore <= 0 {
		t.Error("Success score not calculated")
	}

	// Verify session removed from active sessions
	_, exists := tracker.GetActiveSession("session_001")
	if exists {
		t.Error("Session still active after completion")
	}

	// Verify trajectory stored
	stored, exists := store.trajectories[trajectory.ID]
	if !exists {
		t.Error("Trajectory not stored in episodic memory")
	}

	if stored.SessionID != "session_001" {
		t.Error("Stored trajectory has wrong session_id")
	}
}

func TestInferStrategy(t *testing.T) {
	tests := []struct {
		name     string
		steps    []*ReasoningStep
		expected string
	}{
		{
			name: "systematic linear",
			steps: []*ReasoningStep{
				{Mode: "linear", Tool: "think"},
				{Mode: "linear", Tool: "validate"},
			},
			expected: "systematic-linear",
		},
		{
			name: "parallel exploration",
			steps: []*ReasoningStep{
				{Mode: "tree", Tool: "think"},
				{Mode: "tree", Tool: "list-branches"},
			},
			expected: "parallel-exploration",
		},
		{
			name: "creative divergent",
			steps: []*ReasoningStep{
				{Mode: "divergent", Tool: "think"},
			},
			expected: "creative-divergent",
		},
		{
			name: "backtracking",
			steps: []*ReasoningStep{
				{Mode: "linear", Tool: "create-checkpoint"},
				{Mode: "linear", Tool: "restore-checkpoint"},
			},
			expected: "exploratory-with-backtracking",
		},
		{
			name: "multi-branch creative",
			steps: []*ReasoningStep{
				{Mode: "tree", Tool: "think"},
				{Mode: "divergent", Tool: "think"},
			},
			expected: "multi-branch-creative",
		},
		{
			name:     "empty steps",
			steps:    []*ReasoningStep{},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := inferStrategy(tt.steps)

			if strategy != tt.expected {
				t.Errorf("Expected strategy %s, got %s", tt.expected, strategy)
			}
		})
	}
}

func TestCalculateQualityMetrics(t *testing.T) {
	session := &ActiveSession{
		Steps: []*ReasoningStep{
			{Success: true},
			{Success: true},
			{Success: true},
		},
		Problem: &ProblemDescription{
			Goals: []string{"goal1", "goal2"},
		},
		ModesUsed: map[string]bool{
			"divergent": true,
		},
		ToolsUsed: map[string]int{
			"find-analogy": 1,
		},
	}

	outcome := &OutcomeDescription{
		GoalsAchieved: []string{"goal1", "goal2"},
		Confidence:    0.9,
	}

	metrics := calculateQualityMetrics(session, outcome)

	if metrics == nil {
		t.Fatal("calculateQualityMetrics returned nil")
	}

	if metrics.Efficiency <= 0 {
		t.Error("Efficiency not calculated")
	}

	if metrics.Completeness != 1.0 {
		t.Errorf("Expected completeness 1.0 (all goals achieved), got %.2f", metrics.Completeness)
	}

	if metrics.Reliability != 0.9 {
		t.Errorf("Expected reliability 0.9, got %.2f", metrics.Reliability)
	}

	if metrics.Innovation <= 0 {
		t.Error("Innovation not calculated (should detect divergent mode)")
	}

	if metrics.OverallQuality <= 0 {
		t.Error("Overall quality not calculated")
	}
}

func TestCalculateSuccessScore(t *testing.T) {
	tests := []struct {
		name     string
		outcome  *OutcomeDescription
		quality  *QualityMetrics
		minScore float64
		maxScore float64
	}{
		{
			name: "perfect success",
			outcome: &OutcomeDescription{
				Status:     "success",
				Confidence: 1.0,
			},
			quality: &QualityMetrics{
				OverallQuality: 1.0,
			},
			minScore: 0.9,
			maxScore: 1.0,
		},
		{
			name: "partial success",
			outcome: &OutcomeDescription{
				Status:     "partial",
				Confidence: 0.6,
			},
			quality: &QualityMetrics{
				OverallQuality: 0.5,
			},
			minScore: 0.5,
			maxScore: 0.7,
		},
		{
			name: "failure",
			outcome: &OutcomeDescription{
				Status:     "failure",
				Confidence: 0.3,
			},
			quality: &QualityMetrics{
				OverallQuality: 0.2,
			},
			minScore: 0.1,
			maxScore: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateSuccessScore(tt.outcome, tt.quality)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Expected success score between %.2f and %.2f, got %.2f",
					tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestListActiveSessions(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker := NewSessionTracker(store)
	ctx := context.Background()

	// Start multiple sessions
	for i := 0; i < 3; i++ {
		sessionID := fmt.Sprintf("session_%d", i)
		problem := &ProblemDescription{Description: "test"}
		tracker.StartSession(ctx, sessionID, problem)
	}

	sessions := tracker.ListActiveSessions()

	if len(sessions) != 3 {
		t.Errorf("Expected 3 active sessions, got %d", len(sessions))
	}

	// Complete one session
	outcome := &OutcomeDescription{Status: "success"}
	tracker.CompleteSession(ctx, "session_0", outcome)

	sessions = tracker.ListActiveSessions()

	if len(sessions) != 2 {
		t.Errorf("Expected 2 active sessions after completion, got %d", len(sessions))
	}
}

func TestAutoStartSession(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker := NewSessionTracker(store)
	ctx := context.Background()

	// Record step without starting session (should auto-start)
	step := &ReasoningStep{
		Tool:    "think",
		Success: true,
	}

	err := tracker.RecordStep(ctx, "new_session", step)
	if err != nil {
		t.Fatalf("RecordStep failed: %v", err)
	}

	// Verify session was auto-started
	session, exists := tracker.GetActiveSession("new_session")
	if !exists {
		t.Fatal("Session not auto-started")
	}

	if len(session.Steps) != 1 {
		t.Errorf("Expected 1 step after auto-start, got %d", len(session.Steps))
	}
}
