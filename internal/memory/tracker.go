// Package memory - Trajectory tracking for active reasoning sessions
package memory

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// SessionTracker tracks active reasoning sessions and builds trajectories
type SessionTracker struct {
	activeSessions map[string]*ActiveSession
	store          *EpisodicMemoryStore
	mu             sync.RWMutex
}

// ActiveSession represents an ongoing reasoning session
type ActiveSession struct {
	SessionID    string
	ProblemID    string
	StartTime    time.Time
	Problem      *ProblemDescription
	Steps        []*ReasoningStep
	CurrentStep  int
	ModesUsed    map[string]bool
	ToolsUsed    map[string]int
	BranchesUsed map[string]bool
	Domain       string
	Tags         []string
	Metadata     map[string]interface{}
	mu           sync.Mutex
}

// NewSessionTracker creates a new session tracker
func NewSessionTracker(store *EpisodicMemoryStore) *SessionTracker {
	return &SessionTracker{
		activeSessions: make(map[string]*ActiveSession),
		store:          store,
	}
}

// StartSession begins tracking a new reasoning session
func (t *SessionTracker) StartSession(ctx context.Context, sessionID string, problem *ProblemDescription) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	session := &ActiveSession{
		SessionID:    sessionID,
		ProblemID:    computeProblemHash(problem),
		StartTime:    time.Now(),
		Problem:      problem,
		Steps:        make([]*ReasoningStep, 0),
		CurrentStep:  0,
		ModesUsed:    make(map[string]bool),
		ToolsUsed:    make(map[string]int),
		BranchesUsed: make(map[string]bool),
		Domain:       problem.Domain,
		Tags:         make([]string, 0),
		Metadata:     make(map[string]interface{}),
	}

	t.activeSessions[sessionID] = session
	return nil
}

// RecordStep records a reasoning step in an active session
func (t *SessionTracker) RecordStep(ctx context.Context, sessionID string, step *ReasoningStep) error {
	t.mu.RLock()
	session, exists := t.activeSessions[sessionID]
	t.mu.RUnlock()

	if !exists {
		// Auto-start session if it doesn't exist
		problem := &ProblemDescription{
			Description: "Implicit problem from tool usage",
			Domain:      "unknown",
		}
		if err := t.StartSession(ctx, sessionID, problem); err != nil {
			log.Printf("Warning: failed to auto-start session: %v", err)
		}

		t.mu.RLock()
		session = t.activeSessions[sessionID]
		t.mu.RUnlock()
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// Set step number
	step.StepNumber = len(session.Steps) + 1
	step.Timestamp = time.Now()

	// Add to session
	session.Steps = append(session.Steps, step)
	session.CurrentStep = step.StepNumber

	// Update tracking maps
	if step.Mode != "" {
		session.ModesUsed[step.Mode] = true
	}
	if step.Tool != "" {
		session.ToolsUsed[step.Tool]++
	}
	if step.BranchID != "" {
		session.BranchesUsed[step.BranchID] = true
	}

	return nil
}

// RecordThought is a convenience method to record a thought as a step
func (t *SessionTracker) RecordThought(ctx context.Context, sessionID string, thought *types.Thought, tool string, duration time.Duration) error {
	step := &ReasoningStep{
		Tool:       tool,
		Mode:       string(thought.Mode),
		ThoughtID:  thought.ID,
		BranchID:   thought.BranchID,
		Confidence: thought.Confidence,
		Duration:   duration,
		Success:    true,
		Input: map[string]interface{}{
			"content": thought.Content,
		},
		Output: map[string]interface{}{
			"thought_id": thought.ID,
			"confidence": thought.Confidence,
		},
	}

	return t.RecordStep(ctx, sessionID, step)
}

// CompleteSession marks a session as complete and builds the trajectory
func (t *SessionTracker) CompleteSession(ctx context.Context, sessionID string, outcome *OutcomeDescription) (*ReasoningTrajectory, error) {
	t.mu.Lock()
	session, exists := t.activeSessions[sessionID]
	if !exists {
		t.mu.Unlock()
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	delete(t.activeSessions, sessionID)
	t.mu.Unlock()

	session.mu.Lock()
	defer session.mu.Unlock()

	endTime := time.Now()
	duration := endTime.Sub(session.StartTime)

	// Build tool sequence
	toolSequence := make([]string, 0, len(session.Steps))
	for _, step := range session.Steps {
		toolSequence = append(toolSequence, step.Tool)
	}

	// Extract modes used
	modesUsed := make([]string, 0, len(session.ModesUsed))
	for mode := range session.ModesUsed {
		modesUsed = append(modesUsed, mode)
	}

	// Build approach description
	approach := &ApproachDescription{
		Strategy:     inferStrategy(session.Steps),
		ModesUsed:    modesUsed,
		ToolSequence: toolSequence,
		KeyDecisions: extractKeyDecisions(session.Steps),
	}

	// Calculate quality metrics
	quality := calculateQualityMetrics(session, outcome)

	// Calculate success score
	successScore := calculateSuccessScore(outcome, quality)

	// Infer tags
	tags := inferTags(session, outcome)

	// Build trajectory
	trajectory := &ReasoningTrajectory{
		SessionID:    sessionID,
		ProblemID:    session.ProblemID,
		StartTime:    session.StartTime,
		EndTime:      endTime,
		Duration:     duration,
		Problem:      session.Problem,
		Approach:     approach,
		Steps:        session.Steps,
		Outcome:      outcome,
		Quality:      quality,
		Tags:         tags,
		Domain:       session.Domain,
		Complexity:   session.Problem.Complexity,
		SuccessScore: successScore,
		Metadata:     session.Metadata,
	}

	// Store in episodic memory
	if t.store != nil {
		if err := t.store.StoreTrajectory(ctx, trajectory); err != nil {
			log.Printf("Warning: failed to store trajectory: %v", err)
		}
	}

	return trajectory, nil
}

// GetActiveSession retrieves an active session
func (t *SessionTracker) GetActiveSession(sessionID string) (*ActiveSession, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	session, exists := t.activeSessions[sessionID]
	return session, exists
}

// ListActiveSessions returns all active session IDs
func (t *SessionTracker) ListActiveSessions() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ids := make([]string, 0, len(t.activeSessions))
	for id := range t.activeSessions {
		ids = append(ids, id)
	}
	return ids
}

// Helper functions

func inferStrategy(steps []*ReasoningStep) string {
	if len(steps) == 0 {
		return "unknown"
	}

	// Analyze tool usage patterns
	hasLinear := false
	hasTree := false
	hasDivergent := false
	hasBacktracking := false

	for _, step := range steps {
		switch step.Mode {
		case "linear":
			hasLinear = true
		case "tree":
			hasTree = true
		case "divergent":
			hasDivergent = true
		}
		if step.Tool == "create-checkpoint" || step.Tool == "restore-checkpoint" {
			hasBacktracking = true
		}
	}

	// Determine strategy
	if hasBacktracking {
		return "exploratory-with-backtracking"
	} else if hasTree && hasDivergent {
		return "multi-branch-creative"
	} else if hasTree {
		return "parallel-exploration"
	} else if hasDivergent {
		return "creative-divergent"
	} else if hasLinear {
		return "systematic-linear"
	}

	return "mixed-approach"
}

func extractKeyDecisions(steps []*ReasoningStep) []string {
	decisions := make([]string, 0)

	for _, step := range steps {
		// Look for decision-making tools
		switch step.Tool {
		case "make-decision":
			decisions = append(decisions, fmt.Sprintf("Decision at step %d", step.StepNumber))
		case "focus-branch":
			decisions = append(decisions, fmt.Sprintf("Switched branch at step %d", step.StepNumber))
		case "restore-checkpoint":
			decisions = append(decisions, fmt.Sprintf("Backtracked at step %d", step.StepNumber))
		case "synthesize-insights":
			decisions = append(decisions, fmt.Sprintf("Synthesized insights at step %d", step.StepNumber))
		}
	}

	return decisions
}

func calculateQualityMetrics(session *ActiveSession, outcome *OutcomeDescription) *QualityMetrics {
	// Start with basic metrics
	metrics := &QualityMetrics{
		OverallQuality:     0.5, // Default
		Efficiency:         calculateEfficiency(session),
		Coherence:          0.7, // Default (would be calculated from contradiction detection)
		Completeness:       calculateCompleteness(session, outcome),
		Innovation:         calculateInnovation(session),
		Reliability:        outcome.Confidence,
		BiasScore:          0.0,
		FallacyCount:       0,
		ContradictionCount: 0,
		SelfEvalScore:      0.0,
	}

	// Calculate overall quality as weighted average
	metrics.OverallQuality = (metrics.Efficiency*0.2 +
		metrics.Coherence*0.2 +
		metrics.Completeness*0.2 +
		metrics.Innovation*0.1 +
		metrics.Reliability*0.3)

	return metrics
}

func calculateEfficiency(session *ActiveSession) float64 {
	if len(session.Steps) == 0 {
		return 0.0
	}

	// Fewer steps for same result = more efficient
	// Assume optimal is around 5-10 steps
	optimalSteps := 7.0
	actualSteps := float64(len(session.Steps))

	if actualSteps <= optimalSteps {
		return 1.0
	}

	// Diminishing returns for more steps
	efficiency := optimalSteps / actualSteps
	if efficiency < 0.3 {
		efficiency = 0.3 // Minimum efficiency
	}

	return efficiency
}

func calculateCompleteness(session *ActiveSession, outcome *OutcomeDescription) float64 {
	if session.Problem == nil || len(session.Problem.Goals) == 0 {
		return 0.5 // Unknown
	}

	if outcome == nil {
		return 0.0
	}

	totalGoals := len(session.Problem.Goals)
	achievedGoals := len(outcome.GoalsAchieved)

	if totalGoals == 0 {
		return 1.0
	}

	return float64(achievedGoals) / float64(totalGoals)
}

func calculateInnovation(session *ActiveSession) float64 {
	// Check for use of advanced/creative tools
	innovativeTools := map[string]bool{
		"find-analogy":             true,
		"generate-hypotheses":      true,
		"detect-emergent-patterns": true,
		"think":                    false, // Base tool
	}

	innovationCount := 0
	for tool, count := range session.ToolsUsed {
		if innovativeTools[tool] && count > 0 {
			innovationCount++
		}
	}

	// Also check for divergent mode usage
	if session.ModesUsed["divergent"] {
		innovationCount++
	}

	// Normalize to 0-1 range
	maxInnovation := 5.0
	innovation := float64(innovationCount) / maxInnovation
	if innovation > 1.0 {
		innovation = 1.0
	}

	return innovation
}

func calculateSuccessScore(outcome *OutcomeDescription, quality *QualityMetrics) float64 {
	if outcome == nil {
		return 0.0
	}

	baseScore := 0.0
	switch outcome.Status {
	case "success":
		baseScore = 0.9
	case "partial":
		baseScore = 0.6
	case "failure":
		baseScore = 0.2
	default:
		baseScore = 0.5
	}

	// Adjust by quality
	successScore := baseScore*0.6 + quality.OverallQuality*0.4

	// Adjust by confidence
	successScore = successScore*0.7 + outcome.Confidence*0.3

	return successScore
}

func inferTags(session *ActiveSession, outcome *OutcomeDescription) []string {
	tags := make([]string, 0)

	// Add domain tag
	if session.Domain != "" {
		tags = append(tags, session.Domain)
	}

	// Add strategy tag
	if len(session.Steps) > 0 {
		strategy := inferStrategy(session.Steps)
		tags = append(tags, strategy)
	}

	// Add outcome tag
	if outcome != nil {
		tags = append(tags, outcome.Status)
	}

	// Add complexity tag
	if session.Problem != nil {
		if session.Problem.Complexity > 0.7 {
			tags = append(tags, "high-complexity")
		} else if session.Problem.Complexity < 0.3 {
			tags = append(tags, "low-complexity")
		} else {
			tags = append(tags, "medium-complexity")
		}
	}

	// Add mode tags
	for mode := range session.ModesUsed {
		tags = append(tags, "mode:"+mode)
	}

	return tags
}
