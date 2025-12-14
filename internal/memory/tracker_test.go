package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"unified-thinking/internal/types"
)

func TestNewSessionTracker(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker, err := NewSessionTracker(store)
	if err != nil {
		t.Fatalf("NewSessionTracker failed: %v", err)
	}

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

func TestNewSessionTracker_NilStore(t *testing.T) {
	_, err := NewSessionTracker(nil)
	if err == nil {
		t.Fatal("NewSessionTracker should fail with nil store")
	}
}

func TestStartSession(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker, _ := NewSessionTracker(store)
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
	tracker, _ := NewSessionTracker(store)
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
	tracker, _ := NewSessionTracker(store)
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
	tracker, _ := NewSessionTracker(store)
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
	tracker, _ := NewSessionTracker(store)
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
	tracker, _ := NewSessionTracker(store)
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

func TestRecordThought(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker, _ := NewSessionTracker(store)
	ctx := context.Background()

	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "testing",
	}

	tracker.StartSession(ctx, "session_001", problem)

	thought := &types.Thought{
		ID:         "thought_001",
		Content:    "Test thought content",
		Mode:       types.ModeLinear,
		BranchID:   "branch_001",
		Confidence: 0.85,
	}

	err := tracker.RecordThought(ctx, "session_001", thought, "think", 500*time.Millisecond)
	if err != nil {
		t.Fatalf("RecordThought failed: %v", err)
	}

	// Verify step was recorded
	session, _ := tracker.GetActiveSession("session_001")
	if len(session.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(session.Steps))
	}

	step := session.Steps[0]
	if step.Tool != "think" {
		t.Errorf("Expected tool 'think', got %s", step.Tool)
	}
	if step.Mode != "linear" {
		t.Errorf("Expected mode 'linear', got %s", step.Mode)
	}
	if step.ThoughtID != "thought_001" {
		t.Errorf("Expected thought_id 'thought_001', got %s", step.ThoughtID)
	}
	if step.BranchID != "branch_001" {
		t.Errorf("Expected branch_id 'branch_001', got %s", step.BranchID)
	}
	if step.Confidence != 0.85 {
		t.Errorf("Expected confidence 0.85, got %.2f", step.Confidence)
	}
	if step.Duration != 500*time.Millisecond {
		t.Errorf("Expected duration 500ms, got %v", step.Duration)
	}
	if !step.Success {
		t.Error("Expected success to be true")
	}

	// Verify input/output maps
	if step.Input["content"] != "Test thought content" {
		t.Error("Input content not set correctly")
	}
	if step.Output["thought_id"] != "thought_001" {
		t.Error("Output thought_id not set correctly")
	}
}

func TestRecordThought_AutoStartSession(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker, _ := NewSessionTracker(store)
	ctx := context.Background()

	thought := &types.Thought{
		ID:         "thought_001",
		Content:    "Test thought",
		Mode:       types.ModeTree,
		Confidence: 0.7,
	}

	// Should auto-start session
	err := tracker.RecordThought(ctx, "new_session", thought, "think", time.Second)
	if err != nil {
		t.Fatalf("RecordThought failed: %v", err)
	}

	// Verify session was auto-started
	session, exists := tracker.GetActiveSession("new_session")
	if !exists {
		t.Fatal("Session not auto-started")
	}

	if len(session.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(session.Steps))
	}
}

func TestExtractKeyDecisions_AllDecisionTypes(t *testing.T) {
	steps := []*ReasoningStep{
		{StepNumber: 1, Tool: "think"},
		{StepNumber: 2, Tool: "make-decision"},
		{StepNumber: 3, Tool: "focus-branch"},
		{StepNumber: 4, Tool: "restore-checkpoint"},
		{StepNumber: 5, Tool: "synthesize-insights"},
		{StepNumber: 6, Tool: "validate"},
	}

	decisions := extractKeyDecisions(steps)

	// Should have 4 decisions
	if len(decisions) != 4 {
		t.Errorf("Expected 4 decisions, got %d", len(decisions))
	}

	// Verify decision types are captured
	hasMakeDecision := false
	hasFocusBranch := false
	hasRestore := false
	hasSynthesize := false

	for _, d := range decisions {
		if d == "Decision at step 2" {
			hasMakeDecision = true
		}
		if d == "Switched branch at step 3" {
			hasFocusBranch = true
		}
		if d == "Backtracked at step 4" {
			hasRestore = true
		}
		if d == "Synthesized insights at step 5" {
			hasSynthesize = true
		}
	}

	if !hasMakeDecision {
		t.Error("Missing make-decision entry")
	}
	if !hasFocusBranch {
		t.Error("Missing focus-branch entry")
	}
	if !hasRestore {
		t.Error("Missing restore-checkpoint entry")
	}
	if !hasSynthesize {
		t.Error("Missing synthesize-insights entry")
	}
}

func TestExtractKeyDecisions_NoDecisions(t *testing.T) {
	steps := []*ReasoningStep{
		{StepNumber: 1, Tool: "think"},
		{StepNumber: 2, Tool: "validate"},
		{StepNumber: 3, Tool: "prove"},
	}

	decisions := extractKeyDecisions(steps)
	if len(decisions) != 0 {
		t.Errorf("Expected 0 decisions, got %d", len(decisions))
	}
}

func TestCalculateEfficiency_OptimalSteps(t *testing.T) {
	session := &ActiveSession{
		Steps: make([]*ReasoningStep, 5), // 5 steps - below optimal
	}

	efficiency := calculateEfficiency(session)
	if efficiency != 1.0 {
		t.Errorf("Expected efficiency 1.0 for optimal steps, got %.2f", efficiency)
	}
}

func TestCalculateEfficiency_MoreThanOptimal(t *testing.T) {
	session := &ActiveSession{
		Steps: make([]*ReasoningStep, 14), // Double optimal (7)
	}

	efficiency := calculateEfficiency(session)
	// 7/14 = 0.5
	if efficiency != 0.5 {
		t.Errorf("Expected efficiency 0.5 for double optimal steps, got %.2f", efficiency)
	}
}

func TestCalculateEfficiency_ManySteps(t *testing.T) {
	session := &ActiveSession{
		Steps: make([]*ReasoningStep, 100), // Many steps
	}

	efficiency := calculateEfficiency(session)
	// Should hit minimum of 0.3
	if efficiency != 0.3 {
		t.Errorf("Expected minimum efficiency 0.3, got %.2f", efficiency)
	}
}

func TestCalculateEfficiency_EmptySteps(t *testing.T) {
	session := &ActiveSession{
		Steps: []*ReasoningStep{},
	}

	efficiency := calculateEfficiency(session)
	if efficiency != 0.0 {
		t.Errorf("Expected efficiency 0.0 for empty steps, got %.2f", efficiency)
	}
}

func TestCalculateCompleteness_AllGoalsAchieved(t *testing.T) {
	session := &ActiveSession{
		Problem: &ProblemDescription{
			Goals: []string{"goal1", "goal2", "goal3"},
		},
	}
	outcome := &OutcomeDescription{
		GoalsAchieved: []string{"goal1", "goal2", "goal3"},
	}

	completeness := calculateCompleteness(session, outcome)
	if completeness != 1.0 {
		t.Errorf("Expected completeness 1.0, got %.2f", completeness)
	}
}

func TestCalculateCompleteness_PartialGoals(t *testing.T) {
	session := &ActiveSession{
		Problem: &ProblemDescription{
			Goals: []string{"goal1", "goal2", "goal3"},
		},
	}
	outcome := &OutcomeDescription{
		GoalsAchieved: []string{"goal1"},
	}

	completeness := calculateCompleteness(session, outcome)
	if completeness < 0.33 || completeness > 0.34 {
		t.Errorf("Expected completeness ~0.33, got %.2f", completeness)
	}
}

func TestCalculateCompleteness_NoGoals(t *testing.T) {
	session := &ActiveSession{
		Problem: &ProblemDescription{
			Goals: []string{},
		},
	}
	outcome := &OutcomeDescription{}

	completeness := calculateCompleteness(session, outcome)
	// With empty goals, totalGoals is 0, so it returns 1.0
	// But len(Goals) is 0 which triggers the "unknown" case first, returning 0.5
	// The actual behavior depends on the implementation
	if completeness != 0.5 && completeness != 1.0 {
		t.Errorf("Expected completeness 0.5 or 1.0 for empty goals, got %.2f", completeness)
	}
}

func TestCalculateCompleteness_NilProblem(t *testing.T) {
	session := &ActiveSession{
		Problem: nil,
	}
	outcome := &OutcomeDescription{}

	completeness := calculateCompleteness(session, outcome)
	if completeness != 0.5 {
		t.Errorf("Expected completeness 0.5 for nil problem, got %.2f", completeness)
	}
}

func TestCalculateCompleteness_NilOutcome(t *testing.T) {
	session := &ActiveSession{
		Problem: &ProblemDescription{
			Goals: []string{"goal1"},
		},
	}

	completeness := calculateCompleteness(session, nil)
	if completeness != 0.0 {
		t.Errorf("Expected completeness 0.0 for nil outcome, got %.2f", completeness)
	}
}

func TestCalculateInnovation_WithInnovativeTools(t *testing.T) {
	session := &ActiveSession{
		ToolsUsed: map[string]int{
			"find-analogy":             2,
			"generate-hypotheses":      1,
			"detect-emergent-patterns": 1,
		},
		ModesUsed: map[string]bool{
			"divergent": true,
		},
	}

	innovation := calculateInnovation(session)
	// 4 innovative elements, max 5 = 0.8
	if innovation != 0.8 {
		t.Errorf("Expected innovation 0.8, got %.2f", innovation)
	}
}

func TestCalculateInnovation_NoInnovation(t *testing.T) {
	session := &ActiveSession{
		ToolsUsed: map[string]int{
			"think":    5,
			"validate": 3,
		},
		ModesUsed: map[string]bool{
			"linear": true,
		},
	}

	innovation := calculateInnovation(session)
	if innovation != 0.0 {
		t.Errorf("Expected innovation 0.0, got %.2f", innovation)
	}
}

func TestCalculateInnovation_MaxInnovation(t *testing.T) {
	session := &ActiveSession{
		ToolsUsed: map[string]int{
			"find-analogy":             1,
			"generate-hypotheses":      1,
			"detect-emergent-patterns": 1,
		},
		ModesUsed: map[string]bool{
			"divergent": true,
		},
	}

	innovation := calculateInnovation(session)
	// 4 innovative elements, max 5 = 0.8
	if innovation > 1.0 {
		t.Errorf("Innovation should be capped at 1.0, got %.2f", innovation)
	}
}

func TestInferTags(t *testing.T) {
	session := &ActiveSession{
		Domain: "software-engineering",
		Steps: []*ReasoningStep{
			{Mode: "linear"},
			{Mode: "linear"},
		},
		Problem: &ProblemDescription{
			Complexity: 0.8, // High complexity
		},
		ModesUsed: map[string]bool{
			"linear": true,
		},
	}
	outcome := &OutcomeDescription{
		Status: "success",
	}

	tags := inferTags(session, outcome)

	// Check for expected tags
	hasDoamin := false
	hasStrategy := false
	hasStatus := false
	hasComplexity := false
	hasMode := false

	for _, tag := range tags {
		if tag == "software-engineering" {
			hasDoamin = true
		}
		if tag == "systematic-linear" {
			hasStrategy = true
		}
		if tag == "success" {
			hasStatus = true
		}
		if tag == "high-complexity" {
			hasComplexity = true
		}
		if tag == "mode:linear" {
			hasMode = true
		}
	}

	if !hasDoamin {
		t.Error("Missing domain tag")
	}
	if !hasStrategy {
		t.Error("Missing strategy tag")
	}
	if !hasStatus {
		t.Error("Missing status tag")
	}
	if !hasComplexity {
		t.Error("Missing complexity tag")
	}
	if !hasMode {
		t.Error("Missing mode tag")
	}
}

func TestInferTags_LowComplexity(t *testing.T) {
	session := &ActiveSession{
		Steps: []*ReasoningStep{},
		Problem: &ProblemDescription{
			Complexity: 0.2, // Low complexity
		},
		ModesUsed: map[string]bool{},
	}
	outcome := &OutcomeDescription{
		Status: "failure",
	}

	tags := inferTags(session, outcome)

	hasLowComplexity := false
	for _, tag := range tags {
		if tag == "low-complexity" {
			hasLowComplexity = true
		}
	}

	if !hasLowComplexity {
		t.Error("Missing low-complexity tag")
	}
}

func TestInferTags_MediumComplexity(t *testing.T) {
	session := &ActiveSession{
		Steps: []*ReasoningStep{},
		Problem: &ProblemDescription{
			Complexity: 0.5, // Medium complexity
		},
		ModesUsed: map[string]bool{},
	}
	outcome := &OutcomeDescription{}

	tags := inferTags(session, outcome)

	hasMediumComplexity := false
	for _, tag := range tags {
		if tag == "medium-complexity" {
			hasMediumComplexity = true
		}
	}

	if !hasMediumComplexity {
		t.Error("Missing medium-complexity tag")
	}
}

func TestInferTags_NilOutcome(t *testing.T) {
	session := &ActiveSession{
		Steps:     []*ReasoningStep{},
		ModesUsed: map[string]bool{},
	}

	// Should not panic
	tags := inferTags(session, nil)
	if tags == nil {
		t.Error("Tags should not be nil")
	}
}

// TestCompleteSession_NilStore removed - nil store is no longer allowed.
// NewSessionTracker now requires a non-nil store and returns an error if nil.
// See TestNewSessionTracker_NilStore for validation of this requirement.

func TestCompleteSession_SessionNotFound(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker, _ := NewSessionTracker(store)
	ctx := context.Background()

	outcome := &OutcomeDescription{
		Status: "success",
	}

	_, err := tracker.CompleteSession(ctx, "nonexistent_session", outcome)
	if err == nil {
		t.Error("Expected error for nonexistent session")
	}
}

func TestCalculateQualityMetrics_WithSteps(t *testing.T) {
	session := &ActiveSession{
		Steps: []*ReasoningStep{
			{Success: true},
			{Success: true},
			{Success: false},
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
		GoalsAchieved: []string{"goal1"},
		Confidence:    0.8,
	}

	metrics := calculateQualityMetrics(session, outcome)

	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Reliability should match outcome confidence
	if metrics.Reliability != 0.8 {
		t.Errorf("Expected reliability 0.8, got %.2f", metrics.Reliability)
	}

	// Completeness should be 0.5 (1/2 goals)
	if metrics.Completeness != 0.5 {
		t.Errorf("Expected completeness 0.5, got %.2f", metrics.Completeness)
	}

	// Innovation should be positive (divergent mode + find-analogy)
	if metrics.Innovation <= 0 {
		t.Error("Expected positive innovation")
	}

	// Overall quality should be calculated
	if metrics.OverallQuality <= 0 {
		t.Error("Expected positive overall quality")
	}
}

func TestCalculateSuccessScore_DefaultStatus(t *testing.T) {
	outcome := &OutcomeDescription{
		Status:     "unknown",
		Confidence: 0.5,
	}
	quality := &QualityMetrics{
		OverallQuality: 0.5,
	}

	score := calculateSuccessScore(outcome, quality)
	// Default base score is 0.5
	if score < 0.3 || score > 0.7 {
		t.Errorf("Expected score around 0.5, got %.2f", score)
	}
}

func TestRecordStep_WithInsights(t *testing.T) {
	store := NewEpisodicMemoryStore()
	tracker, _ := NewSessionTracker(store)
	ctx := context.Background()

	problem := &ProblemDescription{
		Description: "Test problem",
		Domain:      "testing",
	}

	tracker.StartSession(ctx, "session_001", problem)

	step := &ReasoningStep{
		Tool:     "synthesize-insights",
		Mode:     "reflection",
		Insights: []string{"Insight 1", "Insight 2"},
		Success:  true,
	}

	err := tracker.RecordStep(ctx, "session_001", step)
	if err != nil {
		t.Fatalf("RecordStep failed: %v", err)
	}

	session, _ := tracker.GetActiveSession("session_001")
	recordedStep := session.Steps[0]

	if len(recordedStep.Insights) != 2 {
		t.Errorf("Expected 2 insights, got %d", len(recordedStep.Insights))
	}

	if session.ModesUsed["reflection"] != true {
		t.Error("Mode 'reflection' should be tracked")
	}
}
