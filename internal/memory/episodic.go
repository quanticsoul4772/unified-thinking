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
	"fmt"
	"log"
	"sort"
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

	// Store the trajectory
	s.trajectories[trajectory.ID] = trajectory

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
func (s *EpisodicMemoryStore) GetRecommendations(ctx context.Context, recCtx *RecommendationContext) ([]*Recommendation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Initialize with capacity based on typical recommendation count (never return nil)
	recommendations := make([]*Recommendation, 0, 5)

	// Analyze similar trajectories for patterns
	for _, match := range recCtx.SimilarTrajectories {
		if match.Trajectory.SuccessScore > 0.7 {
			// Recommend successful tool sequences
			if match.Trajectory.Approach != nil {
				rec := &Recommendation{
					Type:            "tool_sequence",
					Priority:        match.SimilarityScore * match.Trajectory.SuccessScore,
					Suggestion:      fmt.Sprintf("Consider using: %v", match.Trajectory.Approach.ToolSequence),
					Reasoning:       fmt.Sprintf("Similar problem solved successfully with %.0f%% success rate", match.Trajectory.SuccessScore*100),
					BasedOn:         []string{match.Trajectory.ID},
					SuccessRate:     match.Trajectory.SuccessScore,
					EstimatedImpact: 0.0,
					Metadata:        make(map[string]interface{}),
				}
				recommendations = append(recommendations, rec)
			}
		} else if match.Trajectory.SuccessScore < 0.4 {
			// Warn about failed approaches
			if match.Trajectory.Approach != nil {
				rec := &Recommendation{
					Type:            "warning",
					Priority:        match.SimilarityScore * 0.8,
					Suggestion:      fmt.Sprintf("Avoid approach: %s", match.Trajectory.Approach.Strategy),
					Reasoning:       fmt.Sprintf("Similar problem failed with this approach (%.0f%% success)", match.Trajectory.SuccessScore*100),
					BasedOn:         []string{match.Trajectory.ID},
					SuccessRate:     0.0,
					EstimatedImpact: 0.0,
					Metadata:        make(map[string]interface{}),
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	// Sort by priority
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority > recommendations[j].Priority
	})

	return recommendations, nil
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
