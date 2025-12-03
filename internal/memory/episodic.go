// Package memory implements episodic reasoning memory for the unified thinking server.
//
// The episodic memory system enables the server to learn from past reasoning sessions,
// recognize successful patterns, and provide adaptive recommendations based on historical
// performance. This transforms the server from a stateless tool provider into a learning
// cognitive partner.
//
// Key capabilities:
// - Store complete reasoning trajectories (not just individual thoughts)
// - Learn which tool combinations work best for problem types
// - Detect failure patterns and suggest improvements
// - Build a case bank of solved problems
// - Provide similarity-based retrieval of past successful approaches
package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// ReasoningTrajectory represents a complete reasoning session from problem to solution
type ReasoningTrajectory struct {
	ID           string                 `json:"id"`
	SessionID    string                 `json:"session_id"`
	ProblemID    string                 `json:"problem_id"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     time.Duration          `json:"duration"`
	Problem      *ProblemDescription    `json:"problem"`
	Approach     *ApproachDescription   `json:"approach"`
	Steps        []*ReasoningStep       `json:"steps"`
	Outcome      *OutcomeDescription    `json:"outcome"`
	Quality      *QualityMetrics        `json:"quality"`
	Tags         []string               `json:"tags"`
	Domain       string                 `json:"domain"`
	Complexity   float64                `json:"complexity"`
	SuccessScore float64                `json:"success_score"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ProblemDescription describes the initial problem
type ProblemDescription struct {
	Description  string                 `json:"description"`
	Context      string                 `json:"context"`
	Goals        []string               `json:"goals"`
	Constraints  []string               `json:"constraints"`
	InitialState map[string]interface{} `json:"initial_state"`
	ProblemType  string                 `json:"problem_type"`
	Complexity   float64                `json:"complexity"`
	Domain       string                 `json:"domain"`

	// Embedding support (optional)
	Embedding     []float32          `json:"embedding,omitempty"`      // Vector representation
	EmbeddingMeta *EmbeddingMetadata `json:"embedding_meta,omitempty"` // Metadata about embedding
}

// EmbeddingMetadata contains metadata about an embedding
type EmbeddingMetadata struct {
	Model     string    `json:"model"`     // e.g., "voyage-3-lite"
	Provider  string    `json:"provider"`  // e.g., "voyage"
	Dimension int       `json:"dimension"` // e.g., 512
	CreatedAt time.Time `json:"created_at"`
	Source    string    `json:"source"` // "description" or "description+context+goals"
}

// ApproachDescription describes the reasoning approach taken
type ApproachDescription struct {
	Strategy     string   `json:"strategy"`
	ModesUsed    []string `json:"modes_used"`
	ToolSequence []string `json:"tool_sequence"`
	KeyDecisions []string `json:"key_decisions"`
	Adaptations  []string `json:"adaptations"`
}

// ReasoningStep represents one step in the reasoning trajectory
type ReasoningStep struct {
	StepNumber   int                    `json:"step_number"`
	Timestamp    time.Time              `json:"timestamp"`
	Tool         string                 `json:"tool"`
	Mode         string                 `json:"mode,omitempty"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	ThoughtID    string                 `json:"thought_id,omitempty"`
	BranchID     string                 `json:"branch_id,omitempty"`
	Confidence   float64                `json:"confidence"`
	Duration     time.Duration          `json:"duration"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Insights     []string               `json:"insights,omitempty"`
}

// OutcomeDescription describes the final outcome
type OutcomeDescription struct {
	Status             string                 `json:"status"` // "success", "partial", "failure"
	GoalsAchieved      []string               `json:"goals_achieved"`
	GoalsFailed        []string               `json:"goals_failed"`
	FinalState         map[string]interface{} `json:"final_state"`
	Solution           string                 `json:"solution"`
	Confidence         float64                `json:"confidence"`
	ValidationResults  []*ValidationResult    `json:"validation_results,omitempty"`
	UnexpectedOutcomes []string               `json:"unexpected_outcomes,omitempty"`
}

// ValidationResult represents a validation check
type ValidationResult struct {
	Type    string  `json:"type"`
	Passed  bool    `json:"passed"`
	Score   float64 `json:"score"`
	Details string  `json:"details"`
}

// QualityMetrics tracks the quality of the reasoning process
type QualityMetrics struct {
	OverallQuality     float64 `json:"overall_quality"`
	Efficiency         float64 `json:"efficiency"`          // Steps taken vs optimal
	Coherence          float64 `json:"coherence"`           // Logical consistency
	Completeness       float64 `json:"completeness"`        // Goal coverage
	Innovation         float64 `json:"innovation"`          // Novel approaches used
	Reliability        float64 `json:"reliability"`         // Confidence in result
	BiasScore          float64 `json:"bias_score"`          // Detected biases
	FallacyCount       int     `json:"fallacy_count"`       // Logical fallacies detected
	ContradictionCount int     `json:"contradiction_count"` // Contradictions found
	SelfEvalScore      float64 `json:"self_eval_score"`     // Metacognitive assessment
}

// TrajectoryPattern represents a learned pattern from multiple trajectories
type TrajectoryPattern struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	ProblemSignature    *ProblemSignature      `json:"problem_signature"`
	SuccessfulApproach  *ApproachDescription   `json:"successful_approach"`
	AverageQuality      float64                `json:"average_quality"`
	SuccessRate         float64                `json:"success_rate"`
	UsageCount          int                    `json:"usage_count"`
	LastUsed            time.Time              `json:"last_used"`
	ExampleTrajectories []string               `json:"example_trajectories"`
	Tags                []string               `json:"tags"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// ProblemSignature is a fingerprint of a problem type for pattern matching
type ProblemSignature struct {
	Domain               string     `json:"domain"`
	ProblemType          string     `json:"problem_type"`
	ComplexityRange      [2]float64 `json:"complexity_range"`
	RequiredCapabilities []string   `json:"required_capabilities"`
	KeywordFingerprint   []string   `json:"keyword_fingerprint"`
	Hash                 string     `json:"hash"`
}

// RecommendationContext contains information for generating recommendations
type RecommendationContext struct {
	CurrentProblem      *ProblemDescription `json:"current_problem"`
	CurrentStep         int                 `json:"current_step"`
	StepsSoFar          []*ReasoningStep    `json:"steps_so_far"`
	CurrentQuality      *QualityMetrics     `json:"current_quality"`
	SimilarTrajectories []*TrajectoryMatch  `json:"similar_trajectories"`
}

// TrajectoryMatch represents a similar past trajectory
type TrajectoryMatch struct {
	Trajectory       *ReasoningTrajectory `json:"trajectory"`
	SimilarityScore  float64              `json:"similarity_score"`
	RelevanceFactors []string             `json:"relevance_factors"`
	Recommendation   string               `json:"recommendation"`
}

// Recommendation provides adaptive guidance based on episodic memory
type Recommendation struct {
	Type            string                 `json:"type"` // "tool_sequence", "approach", "warning", "optimization"
	Priority        float64                `json:"priority"`
	Suggestion      string                 `json:"suggestion"`
	Reasoning       string                 `json:"reasoning"`
	BasedOn         []string               `json:"based_on"` // Trajectory IDs
	SuccessRate     float64                `json:"success_rate"`
	EstimatedImpact float64                `json:"estimated_impact"`
	Metadata        map[string]interface{} `json:"metadata"`

	// Enhanced fields for specific tool sequences (Phase 2.1)
	ToolSequence     []ToolStep `json:"tool_sequence,omitempty"`      // Specific ordered steps
	ExampleProblems  []string   `json:"example_problems,omitempty"`   // Similar problems solved
	ConfidenceRange  [2]float64 `json:"confidence_range,omitempty"`   // Min/max confidence achieved
	AverageSteps     int        `json:"average_steps,omitempty"`      // Typical step count
	FailureRootCause string     `json:"failure_root_cause,omitempty"` // For warnings
	Alternatives     []string   `json:"alternatives,omitempty"`       // Alternative approaches to try
	PatternID        string     `json:"pattern_id,omitempty"`         // ID of matched pattern
}

// ToolStep represents a specific step in a tool sequence
type ToolStep struct {
	StepNumber  int     `json:"step_number"`
	Tool        string  `json:"tool"`
	Mode        string  `json:"mode,omitempty"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence,omitempty"`
}

// ToolSequencePattern groups similar tool sequences for analysis
type ToolSequencePattern struct {
	SequenceHash    string                 `json:"sequence_hash"`
	Tools           []string               `json:"tools"`
	Trajectories    []*ReasoningTrajectory `json:"trajectories"`
	SuccessRate     float64                `json:"success_rate"`
	AverageQuality  float64                `json:"average_quality"`
	UsageCount      int                    `json:"usage_count"`
	ExampleProblems []string               `json:"example_problems"`
}

// EpisodicMemoryStore manages storage and retrieval of reasoning trajectories
type EpisodicMemoryStore struct {
	trajectories         map[string]*ReasoningTrajectory
	patterns             map[string]*TrajectoryPattern
	problemIndex         map[string][]string   // problem_hash -> trajectory_ids
	domainIndex          map[string][]string   // domain -> trajectory_ids
	tagIndex             map[string][]string   // tag -> trajectory_ids
	toolSequenceIndex    map[string][]string   // tool_sequence_hash -> trajectory_ids
	embeddingIntegration *EmbeddingIntegration // Optional embedding-based search
	signatureIntegration *SignatureIntegration // Optional context signature storage
	storageBackend       interface{}           // Optional persistent storage backend (storage.Storage with trajectory methods)
	mu                   sync.RWMutex
}

// NewEpisodicMemoryStore creates a new episodic memory store
func NewEpisodicMemoryStore() *EpisodicMemoryStore {
	return &EpisodicMemoryStore{
		trajectories:      make(map[string]*ReasoningTrajectory),
		patterns:          make(map[string]*TrajectoryPattern),
		problemIndex:      make(map[string][]string),
		domainIndex:       make(map[string][]string),
		tagIndex:          make(map[string][]string),
		toolSequenceIndex: make(map[string][]string),
	}
}

// StoreTrajectory stores a complete reasoning trajectory
func (s *EpisodicMemoryStore) StoreTrajectory(ctx context.Context, trajectory *ReasoningTrajectory) error {
	// Generate embedding for the problem if we have integration
	log.Printf("StoreTrajectory called, embedding integration: %v, problem: %v", s.embeddingIntegration != nil, trajectory.Problem != nil)
	if s.embeddingIntegration != nil && trajectory.Problem != nil {
		log.Printf("Calling GenerateAndStoreEmbedding")
		if err := s.embeddingIntegration.GenerateAndStoreEmbedding(ctx, trajectory.Problem); err != nil {
			return fmt.Errorf("failed to generate embedding for trajectory: %w", err)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if trajectory.ID == "" {
		trajectory.ID = generateTrajectoryID(trajectory)
	}

	// Store the trajectory in memory
	s.trajectories[trajectory.ID] = trajectory

	// Persist to storage backend if available (using JSON to avoid import cycle)
	if s.storageBackend != nil {
		type trajectoryJSONStorer interface {
			StoreTrajectoryJSON(id string, trajectoryJSON string) error
		}
		if storer, ok := s.storageBackend.(trajectoryJSONStorer); ok {
			// Serialize trajectory to JSON
			trajectoryJSON, err := json.Marshal(trajectory)
			if err != nil {
				log.Printf("Warning: failed to marshal trajectory for persistence: %v", err)
			} else if err := storer.StoreTrajectoryJSON(trajectory.ID, string(trajectoryJSON)); err != nil {
				log.Printf("Warning: failed to persist trajectory to storage backend: %v", err)
			}
		}
	}

	// Update indexes
	if trajectory.Problem != nil {
		problemHash := computeProblemHash(trajectory.Problem)
		s.problemIndex[problemHash] = append(s.problemIndex[problemHash], trajectory.ID)
	}

	if trajectory.Domain != "" {
		s.domainIndex[trajectory.Domain] = append(s.domainIndex[trajectory.Domain], trajectory.ID)
	}

	for _, tag := range trajectory.Tags {
		s.tagIndex[tag] = append(s.tagIndex[tag], trajectory.ID)
	}

	if trajectory.Approach != nil && len(trajectory.Approach.ToolSequence) > 0 {
		seqHash := computeToolSequenceHash(trajectory.Approach.ToolSequence)
		s.toolSequenceIndex[seqHash] = append(s.toolSequenceIndex[seqHash], trajectory.ID)
	}

	// Generate and store context signature if we have integration
	if s.signatureIntegration != nil {
		log.Printf("[DEBUG] Generating context signature for trajectory %s", trajectory.ID)
		if err := s.signatureIntegration.GenerateAndStoreSignature(trajectory); err != nil {
			return fmt.Errorf("failed to store context signature for trajectory %s: %w", trajectory.ID, err)
		}
		log.Printf("[DEBUG] Context signature stored for trajectory %s", trajectory.ID)
	} else {
		log.Printf("[DEBUG] No signature integration configured for trajectory %s", trajectory.ID)
	}

	return nil
}

// RetrieveSimilarTrajectories finds similar past trajectories
func (s *EpisodicMemoryStore) RetrieveSimilarTrajectories(ctx context.Context, problem *ProblemDescription, limit int) ([]*TrajectoryMatch, error) {
	// If we have embedding integration, use hybrid search
	if s.embeddingIntegration != nil {
		return s.embeddingIntegration.RetrieveSimilarWithHybridSearch(ctx, problem, limit)
	}

	// Otherwise, use hash-based search
	return s.RetrieveSimilarHashBased(problem, limit)
}

// RetrieveSimilarHashBased performs hash-based similarity search (internal method, exported for embedding integration)
func (s *EpisodicMemoryStore) RetrieveSimilarHashBased(problem *ProblemDescription, limit int) ([]*TrajectoryMatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Initialize with capacity based on expected limit (never return nil)
	matches := make([]*TrajectoryMatch, 0, limit)

	// Find candidates by problem similarity
	problemHash := computeProblemHash(problem)
	candidateIDs := s.problemIndex[problemHash]

	// Also include trajectories from same domain
	if problem.Domain != "" {
		candidateIDs = append(candidateIDs, s.domainIndex[problem.Domain]...)
	}

	// Remove duplicates
	seen := make(map[string]bool, len(candidateIDs))
	uniqueIDs := make([]string, 0, len(candidateIDs))
	for _, id := range candidateIDs {
		if !seen[id] {
			seen[id] = true
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	// Calculate similarity scores
	for _, id := range uniqueIDs {
		trajectory, exists := s.trajectories[id]
		if !exists {
			continue
		}

		similarity := calculateProblemSimilarity(problem, trajectory.Problem)
		if similarity > 0.3 { // Minimum similarity threshold
			match := &TrajectoryMatch{
				Trajectory:       trajectory,
				SimilarityScore:  similarity,
				RelevanceFactors: identifyRelevanceFactors(problem, trajectory),
			}
			matches = append(matches, match)
		}
	}

	// Sort by similarity (descending)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].SimilarityScore > matches[j].SimilarityScore
	})

	// Apply limit
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	return matches, nil
}

// GetRecommendations generates adaptive recommendations based on context
// Enhanced in Phase 2.1 to provide specific tool sequences with success rates
func (s *EpisodicMemoryStore) GetRecommendations(ctx context.Context, recCtx *RecommendationContext) ([]*Recommendation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Initialize with capacity based on typical recommendation count (never return nil)
	recommendations := make([]*Recommendation, 0, 5)

	// Group similar trajectories by tool sequence patterns
	successfulPatterns := s.groupByToolSequencePattern(recCtx.SimilarTrajectories, true) // success > 0.7
	failedPatterns := s.groupByToolSequencePattern(recCtx.SimilarTrajectories, false)    // success < 0.4

	// Generate recommendations from successful patterns (grouped by tool sequence)
	for _, pattern := range successfulPatterns {
		if len(pattern.Tools) == 0 {
			continue
		}

		// Build specific tool steps from the best trajectory in this pattern
		toolSteps := s.extractToolSteps(pattern.Trajectories[0])

		// Collect example problems solved with this pattern
		examples := make([]string, 0, 3)
		trajectoryIDs := make([]string, 0, len(pattern.Trajectories))
		minConf, maxConf := 1.0, 0.0
		totalSteps := 0

		for i, traj := range pattern.Trajectories {
			trajectoryIDs = append(trajectoryIDs, traj.ID)
			if traj.Problem != nil && i < 3 {
				// Truncate long descriptions
				desc := traj.Problem.Description
				if len(desc) > 80 {
					desc = desc[:77] + "..."
				}
				examples = append(examples, desc)
			}
			// Track confidence range
			if traj.Outcome != nil {
				if traj.Outcome.Confidence < minConf {
					minConf = traj.Outcome.Confidence
				}
				if traj.Outcome.Confidence > maxConf {
					maxConf = traj.Outcome.Confidence
				}
			}
			totalSteps += len(traj.Steps)
		}
		avgSteps := 0
		if len(pattern.Trajectories) > 0 {
			avgSteps = totalSteps / len(pattern.Trajectories)
		}

		// Build descriptive suggestion with specific steps
		suggestion := s.buildToolSequenceSuggestion(pattern.Tools, toolSteps)

		rec := &Recommendation{
			Type:            "tool_sequence",
			Priority:        pattern.SuccessRate * (1.0 + 0.1*float64(pattern.UsageCount)), // Boost by usage count
			Suggestion:      suggestion,
			Reasoning:       fmt.Sprintf("Proven sequence: %d similar problems solved with %.0f%% success rate", pattern.UsageCount, pattern.SuccessRate*100),
			BasedOn:         trajectoryIDs,
			SuccessRate:     pattern.SuccessRate,
			EstimatedImpact: pattern.AverageQuality,
			Metadata:        map[string]interface{}{"pattern_usage": pattern.UsageCount},
			ToolSequence:    toolSteps,
			ExampleProblems: examples,
			ConfidenceRange: [2]float64{minConf, maxConf},
			AverageSteps:    avgSteps,
			PatternID:       pattern.SequenceHash,
		}
		recommendations = append(recommendations, rec)
	}

	// Generate warnings from failed patterns with root cause analysis
	for _, pattern := range failedPatterns {
		if len(pattern.Tools) == 0 {
			continue
		}

		trajectoryIDs := make([]string, 0, len(pattern.Trajectories))
		for _, traj := range pattern.Trajectories {
			trajectoryIDs = append(trajectoryIDs, traj.ID)
		}

		// Analyze root cause from failed trajectories
		rootCause := s.analyzeFailureRootCause(pattern.Trajectories)
		alternatives := s.suggestAlternatives(pattern.Tools, successfulPatterns)

		rec := &Recommendation{
			Type:             "warning",
			Priority:         0.8 * (1.0 + 0.05*float64(len(pattern.Trajectories))), // Boost by failure count
			Suggestion:       fmt.Sprintf("Avoid sequence: %v", pattern.Tools),
			Reasoning:        fmt.Sprintf("Failed in %d similar cases (%.0f%% success rate). %s", pattern.UsageCount, pattern.SuccessRate*100, rootCause),
			BasedOn:          trajectoryIDs,
			SuccessRate:      pattern.SuccessRate,
			EstimatedImpact:  0.0,
			Metadata:         map[string]interface{}{"failure_count": pattern.UsageCount},
			ToolSequence:     s.extractToolSteps(pattern.Trajectories[0]),
			FailureRootCause: rootCause,
			Alternatives:     alternatives,
			PatternID:        pattern.SequenceHash,
		}
		recommendations = append(recommendations, rec)
	}

	// Add approach-level recommendations for high-performing strategies
	for _, match := range recCtx.SimilarTrajectories {
		if match.Trajectory.SuccessScore > 0.85 && match.Trajectory.Approach != nil && match.Trajectory.Approach.Strategy != "" {
			// Check if we already have a tool_sequence recommendation for this
			hasToolSeq := false
			for _, rec := range recommendations {
				if rec.Type == "tool_sequence" {
					for _, id := range rec.BasedOn {
						if id == match.Trajectory.ID {
							hasToolSeq = true
							break
						}
					}
				}
				if hasToolSeq {
					break
				}
			}

			if !hasToolSeq {
				rec := &Recommendation{
					Type:            "approach",
					Priority:        match.SimilarityScore * match.Trajectory.SuccessScore * 0.9, // Slightly lower than tool_sequence
					Suggestion:      fmt.Sprintf("Consider strategy: %s", match.Trajectory.Approach.Strategy),
					Reasoning:       fmt.Sprintf("High success (%.0f%%) on similar problem: %s", match.Trajectory.SuccessScore*100, s.truncateDescription(match.Trajectory.Problem)),
					BasedOn:         []string{match.Trajectory.ID},
					SuccessRate:     match.Trajectory.SuccessScore,
					EstimatedImpact: match.Trajectory.Quality.OverallQuality,
					Metadata:        make(map[string]interface{}),
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	// Sort by priority (descending)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority > recommendations[j].Priority
	})

	// Limit to top recommendations (avoid overwhelming the user)
	if len(recommendations) > 7 {
		recommendations = recommendations[:7]
	}

	return recommendations, nil
}

// groupByToolSequencePattern groups trajectories by their tool sequence patterns
func (s *EpisodicMemoryStore) groupByToolSequencePattern(matches []*TrajectoryMatch, successful bool) []*ToolSequencePattern {
	patternMap := make(map[string]*ToolSequencePattern)

	for _, match := range matches {
		traj := match.Trajectory

		// Filter by success criteria
		if successful && traj.SuccessScore <= 0.7 {
			continue
		}
		if !successful && traj.SuccessScore >= 0.4 {
			continue
		}

		if traj.Approach == nil || len(traj.Approach.ToolSequence) == 0 {
			continue
		}

		// Compute pattern hash
		seqHash := computeToolSequenceHash(traj.Approach.ToolSequence)

		if pattern, exists := patternMap[seqHash]; exists {
			// Add to existing pattern
			pattern.Trajectories = append(pattern.Trajectories, traj)
			pattern.UsageCount++
			// Update aggregate metrics
			pattern.SuccessRate = (pattern.SuccessRate*float64(pattern.UsageCount-1) + traj.SuccessScore) / float64(pattern.UsageCount)
			if traj.Quality != nil {
				pattern.AverageQuality = (pattern.AverageQuality*float64(pattern.UsageCount-1) + traj.Quality.OverallQuality) / float64(pattern.UsageCount)
			}
		} else {
			// Create new pattern
			quality := 0.0
			if traj.Quality != nil {
				quality = traj.Quality.OverallQuality
			}
			patternMap[seqHash] = &ToolSequencePattern{
				SequenceHash:   seqHash,
				Tools:          traj.Approach.ToolSequence,
				Trajectories:   []*ReasoningTrajectory{traj},
				SuccessRate:    traj.SuccessScore,
				AverageQuality: quality,
				UsageCount:     1,
			}
		}
	}

	// Convert map to slice and sort by usage count
	patterns := make([]*ToolSequencePattern, 0, len(patternMap))
	for _, pattern := range patternMap {
		patterns = append(patterns, pattern)
	}

	sort.Slice(patterns, func(i, j int) bool {
		// Sort by success rate * usage count for best recommendations first
		scoreI := patterns[i].SuccessRate * float64(patterns[i].UsageCount)
		scoreJ := patterns[j].SuccessRate * float64(patterns[j].UsageCount)
		return scoreI > scoreJ
	})

	return patterns
}

// extractToolSteps extracts detailed tool steps from a trajectory
func (s *EpisodicMemoryStore) extractToolSteps(traj *ReasoningTrajectory) []ToolStep {
	if traj == nil || len(traj.Steps) == 0 {
		return nil
	}

	steps := make([]ToolStep, 0, len(traj.Steps))
	for _, step := range traj.Steps {
		desc := s.generateStepDescription(step)
		steps = append(steps, ToolStep{
			StepNumber:  step.StepNumber,
			Tool:        step.Tool,
			Mode:        step.Mode,
			Description: desc,
			Confidence:  step.Confidence,
		})
	}

	return steps
}

// generateStepDescription generates a human-readable description for a reasoning step
func (s *EpisodicMemoryStore) generateStepDescription(step *ReasoningStep) string {
	if step == nil {
		return ""
	}

	// Build description based on tool type and insights
	desc := ""
	switch step.Tool {
	case "think":
		if step.Mode != "" {
			desc = fmt.Sprintf("Analyze using %s mode", step.Mode)
		} else {
			desc = "Analyze the problem"
		}
	case "decompose-problem":
		desc = "Break down into subproblems"
	case "build-causal-graph":
		desc = "Map causal relationships"
	case "simulate-intervention":
		desc = "Test intervention effects"
	case "generate-hypotheses":
		desc = "Generate possible explanations"
	case "evaluate-hypotheses":
		desc = "Evaluate and rank hypotheses"
	case "detect-biases":
		desc = "Check for cognitive biases"
	case "validate":
		desc = "Validate logical consistency"
	case "synthesize-insights":
		desc = "Combine insights across analyses"
	case "make-decision":
		desc = "Evaluate options and decide"
	default:
		desc = fmt.Sprintf("Execute %s", step.Tool)
	}

	// Add insights if available
	if len(step.Insights) > 0 {
		desc += fmt.Sprintf(" [%s]", step.Insights[0])
	}

	return desc
}

// buildToolSequenceSuggestion builds a clear, actionable suggestion for a tool sequence
func (s *EpisodicMemoryStore) buildToolSequenceSuggestion(tools []string, steps []ToolStep) string {
	if len(steps) == 0 {
		return fmt.Sprintf("Use sequence: %v", tools)
	}

	// Build step-by-step suggestion
	var suggestion string
	if len(steps) <= 3 {
		// Short sequence: list all steps
		parts := make([]string, 0, len(steps))
		for _, step := range steps {
			parts = append(parts, fmt.Sprintf("%d. %s (%s)", step.StepNumber, step.Description, step.Tool))
		}
		suggestion = "Recommended steps: " + strings.Join(parts, " → ")
	} else {
		// Longer sequence: summarize
		suggestion = fmt.Sprintf("Recommended %d-step sequence: Start with %s, then %s, and conclude with %s",
			len(steps),
			steps[0].Tool,
			steps[len(steps)/2].Tool,
			steps[len(steps)-1].Tool)
	}

	return suggestion
}

// analyzeFailureRootCause analyzes common failure patterns in trajectories
func (s *EpisodicMemoryStore) analyzeFailureRootCause(trajectories []*ReasoningTrajectory) string {
	if len(trajectories) == 0 {
		return "Unknown cause"
	}

	// Analyze patterns in failed trajectories
	lowConfidenceCount := 0
	shortSequenceCount := 0
	noValidationCount := 0
	errorMessages := make(map[string]int)

	for _, traj := range trajectories {
		// Check for low confidence
		if traj.Outcome != nil && traj.Outcome.Confidence < 0.5 {
			lowConfidenceCount++
		}

		// Check for short sequences (potentially incomplete analysis)
		if len(traj.Steps) < 3 {
			shortSequenceCount++
		}

		// Check for missing validation
		hasValidation := false
		for _, step := range traj.Steps {
			if step.Tool == "validate" || step.Tool == "self-evaluate" {
				hasValidation = true
				break
			}
			// Track error messages
			if step.ErrorMessage != "" {
				errorMessages[step.ErrorMessage]++
			}
		}
		if !hasValidation {
			noValidationCount++
		}
	}

	// Determine primary root cause
	total := len(trajectories)
	if float64(shortSequenceCount)/float64(total) > 0.5 {
		return "Insufficient analysis depth - try more thorough exploration"
	}
	if float64(noValidationCount)/float64(total) > 0.6 {
		return "Missing validation step - add validate or self-evaluate"
	}
	if float64(lowConfidenceCount)/float64(total) > 0.5 {
		return "Low confidence in results - consider alternative approaches"
	}

	// Check for common error messages
	for msg, count := range errorMessages {
		if float64(count)/float64(total) > 0.3 {
			return fmt.Sprintf("Common error: %s", msg)
		}
	}

	return "Approach may not fit this problem type"
}

// suggestAlternatives suggests alternative tool sequences based on successful patterns
func (s *EpisodicMemoryStore) suggestAlternatives(failedTools []string, successPatterns []*ToolSequencePattern) []string {
	alternatives := make([]string, 0, 3)

	for _, pattern := range successPatterns {
		if len(pattern.Tools) == 0 {
			continue
		}

		// Check if this is a different approach (not just a subset/superset)
		if !s.toolSequencesOverlap(failedTools, pattern.Tools) {
			alt := fmt.Sprintf("Try: %v (%.0f%% success rate)", pattern.Tools, pattern.SuccessRate*100)
			alternatives = append(alternatives, alt)
			if len(alternatives) >= 3 {
				break
			}
		}
	}

	return alternatives
}

// toolSequencesOverlap checks if two tool sequences are substantially similar
func (s *EpisodicMemoryStore) toolSequencesOverlap(seq1, seq2 []string) bool {
	if len(seq1) == 0 || len(seq2) == 0 {
		return false
	}

	// Simple overlap check: >50% common tools means similar
	common := 0
	toolSet := make(map[string]bool)
	for _, t := range seq1 {
		toolSet[t] = true
	}
	for _, t := range seq2 {
		if toolSet[t] {
			common++
		}
	}

	minLen := len(seq1)
	if len(seq2) < minLen {
		minLen = len(seq2)
	}

	return float64(common)/float64(minLen) > 0.5
}

// truncateDescription truncates a problem description for display
func (s *EpisodicMemoryStore) truncateDescription(problem *ProblemDescription) string {
	if problem == nil {
		return ""
	}
	desc := problem.Description
	if len(desc) > 60 {
		return desc[:57] + "..."
	}
	return desc
}

// SetEmbeddingIntegration sets the embedding integration for hybrid search
func (s *EpisodicMemoryStore) SetEmbeddingIntegration(ei *EmbeddingIntegration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.embeddingIntegration = ei
}

// SetSignatureIntegration sets the signature integration for context bridge
func (s *EpisodicMemoryStore) SetSignatureIntegration(si *SignatureIntegration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.signatureIntegration = si
}

// SetStorageBackend sets the persistent storage backend and loads existing trajectories
func (s *EpisodicMemoryStore) SetStorageBackend(backend interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storageBackend = backend

	// Load existing trajectories from storage
	type trajectoryJSONLoader interface {
		GetAllTrajectoriesJSON() (map[string]string, error)
	}
	if loader, ok := backend.(trajectoryJSONLoader); ok {
		trajectoriesJSON, err := loader.GetAllTrajectoriesJSON()
		if err != nil {
			return fmt.Errorf("failed to load trajectories from storage: %w", err)
		}

		// Deserialize and rebuild indexes
		loadedCount := 0
		for id, trajectoryJSON := range trajectoriesJSON {
			var trajectory ReasoningTrajectory
			if err := json.Unmarshal([]byte(trajectoryJSON), &trajectory); err != nil {
				log.Printf("Warning: failed to unmarshal trajectory %s: %v", id, err)
				continue
			}

			// Store in memory
			s.trajectories[id] = &trajectory

			// Rebuild indexes
			if trajectory.Problem != nil {
				problemHash := computeProblemHash(trajectory.Problem)
				s.problemIndex[problemHash] = append(s.problemIndex[problemHash], id)
			}

			if trajectory.Domain != "" {
				s.domainIndex[trajectory.Domain] = append(s.domainIndex[trajectory.Domain], id)
			}

			for _, tag := range trajectory.Tags {
				s.tagIndex[tag] = append(s.tagIndex[tag], id)
			}

			if trajectory.Approach != nil && len(trajectory.Approach.ToolSequence) > 0 {
				seqHash := computeToolSequenceHash(trajectory.Approach.ToolSequence)
				s.toolSequenceIndex[seqHash] = append(s.toolSequenceIndex[seqHash], id)
			}

			loadedCount++
		}

		log.Printf("Loaded %d trajectories from persistent storage", loadedCount)
	}

	return nil
}

// GetAllTrajectories returns all stored trajectories (for search operations)
func (s *EpisodicMemoryStore) GetAllTrajectories() []*ReasoningTrajectory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trajectories := make([]*ReasoningTrajectory, 0, len(s.trajectories))
	for _, traj := range s.trajectories {
		trajectories = append(trajectories, traj)
	}
	return trajectories
}

// Helper functions

func generateTrajectoryID(trajectory *ReasoningTrajectory) string {
	if trajectory.SessionID != "" && trajectory.ProblemID != "" {
		return fmt.Sprintf("traj_%s_%s_%d", trajectory.SessionID, trajectory.ProblemID, trajectory.StartTime.Unix())
	}
	return fmt.Sprintf("traj_%d", time.Now().UnixNano())
}

// ComputeProblemHash computes a hash for a problem description (exported for handlers)
func ComputeProblemHash(problem *ProblemDescription) string {
	data := fmt.Sprintf("%s|%s|%s", problem.Domain, problem.ProblemType, problem.Description)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func computeProblemHash(problem *ProblemDescription) string {
	return ComputeProblemHash(problem)
}

func computeToolSequenceHash(sequence []string) string {
	data := ""
	for _, tool := range sequence {
		data += tool + "|"
	}
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func calculateProblemSimilarity(p1, p2 *ProblemDescription) float64 {
	if p1 == nil || p2 == nil {
		return 0.0
	}

	score := 0.0
	checks := 0.0

	// Domain match
	if p1.Domain == p2.Domain && p1.Domain != "" {
		score += 0.3
	}
	checks += 0.3

	// Problem type match
	if p1.ProblemType == p2.ProblemType && p1.ProblemType != "" {
		score += 0.3
	}
	checks += 0.3

	// Complexity similarity
	if p1.Complexity > 0 && p2.Complexity > 0 {
		complexityDiff := 1.0 - (abs(p1.Complexity-p2.Complexity) / maxFloat(p1.Complexity, p2.Complexity))
		score += complexityDiff * 0.2
	}
	checks += 0.2

	// Goal overlap
	if len(p1.Goals) > 0 && len(p2.Goals) > 0 {
		overlap := calculateSetOverlap(p1.Goals, p2.Goals)
		score += overlap * 0.2
	}
	checks += 0.2

	if checks > 0 {
		return score / checks
	}
	return 0.0
}

func identifyRelevanceFactors(problem *ProblemDescription, trajectory *ReasoningTrajectory) []string {
	factors := []string{}

	if problem.Domain == trajectory.Domain {
		factors = append(factors, "Same domain")
	}

	if problem.ProblemType == trajectory.Problem.ProblemType {
		factors = append(factors, "Same problem type")
	}

	if trajectory.SuccessScore > 0.8 {
		factors = append(factors, "High success rate")
	}

	return factors
}

func calculateSetOverlap(set1, set2 []string) float64 {
	if len(set1) == 0 || len(set2) == 0 {
		return 0.0
	}

	set2Map := make(map[string]bool)
	for _, item := range set2 {
		set2Map[item] = true
	}

	overlap := 0
	for _, item := range set1 {
		if set2Map[item] {
			overlap++
		}
	}

	return float64(overlap) / float64(max(len(set1), len(set2)))
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
