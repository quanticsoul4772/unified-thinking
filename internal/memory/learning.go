// Package memory - Pattern learning and continuous improvement
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// LearningEngine analyzes trajectories to identify successful patterns
type LearningEngine struct {
	store              *EpisodicMemoryStore
	minTrajectories    int     // Minimum trajectories needed to form a pattern
	minSuccessRate     float64 // Minimum success rate to consider a pattern
	patternCache       map[string]*TrajectoryPattern
	lastLearningRun    time.Time
	learningInterval   time.Duration
	mu                 sync.RWMutex
}

// NewLearningEngine creates a new learning engine
func NewLearningEngine(store *EpisodicMemoryStore) *LearningEngine {
	return &LearningEngine{
		store:            store,
		minTrajectories:  3,
		minSuccessRate:   0.6,
		patternCache:     make(map[string]*TrajectoryPattern),
		learningInterval: 1 * time.Hour,
	}
}

// LearnPatterns analyzes stored trajectories and extracts patterns
func (l *LearningEngine) LearnPatterns(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.store.mu.RLock()
	defer l.store.mu.RUnlock()

	// Group trajectories by problem signature
	groupedTrajectories := l.groupByProblemSignature()

	// Analyze each group
	for signature, trajectoryIDs := range groupedTrajectories {
		if len(trajectoryIDs) < l.minTrajectories {
			continue
		}

		// Get trajectories
		trajectories := make([]*ReasoningTrajectory, 0, len(trajectoryIDs))
		for _, id := range trajectoryIDs {
			if traj, exists := l.store.trajectories[id]; exists {
				trajectories = append(trajectories, traj)
			}
		}

		// Analyze for patterns
		pattern := l.analyzeTrajectoryGroup(trajectories, signature)
		if pattern != nil && pattern.SuccessRate >= l.minSuccessRate {
			l.patternCache[pattern.ID] = pattern
			l.store.patterns[pattern.ID] = pattern
		}
	}

	l.lastLearningRun = time.Now()
	return nil
}

// groupByProblemSignature groups trajectories by their problem signature
func (l *LearningEngine) groupByProblemSignature() map[string][]string {
	groups := make(map[string][]string)

	for id, trajectory := range l.store.trajectories {
		if trajectory.Problem == nil {
			continue
		}

		signature := createProblemSignature(trajectory.Problem)
		groups[signature.Hash] = append(groups[signature.Hash], id)
	}

	return groups
}

// analyzeTrajectoryGroup analyzes a group of similar trajectories
func (l *LearningEngine) analyzeTrajectoryGroup(trajectories []*ReasoningTrajectory, signatureHash string) *TrajectoryPattern {
	if len(trajectories) == 0 {
		return nil
	}

	// Separate successful from failed trajectories
	successful := make([]*ReasoningTrajectory, 0)
	for _, traj := range trajectories {
		if traj.SuccessScore >= l.minSuccessRate {
			successful = append(successful, traj)
		}
	}

	if len(successful) == 0 {
		return nil
	}

	// Calculate success rate
	successRate := float64(len(successful)) / float64(len(trajectories))

	// Find common approach in successful trajectories
	commonApproach := findCommonApproach(successful)

	// Calculate average quality
	avgQuality := 0.0
	for _, traj := range successful {
		if traj.Quality != nil {
			avgQuality += traj.Quality.OverallQuality
		}
	}
	avgQuality /= float64(len(successful))

	// Extract problem signature from first trajectory
	problemSig := createProblemSignature(trajectories[0].Problem)

	// Create pattern
	pattern := &TrajectoryPattern{
		ID:                generatePatternID(problemSig),
		Name:              generatePatternName(problemSig, commonApproach),
		Description:       generatePatternDescription(successful),
		ProblemSignature:  problemSig,
		SuccessfulApproach: commonApproach,
		AverageQuality:    avgQuality,
		SuccessRate:       successRate,
		UsageCount:        len(trajectories),
		LastUsed:          time.Now(),
		ExampleTrajectories: extractTrajectoryIDs(successful, 3),
		Tags:              extractCommonTags(successful),
	}

	return pattern
}

// findCommonApproach identifies common strategies in successful trajectories
func findCommonApproach(trajectories []*ReasoningTrajectory) *ApproachDescription {
	if len(trajectories) == 0 {
		return nil
	}

	// Count tool usage frequency
	toolFrequency := make(map[string]int)
	modeFrequency := make(map[string]int)
	strategyFrequency := make(map[string]int)

	for _, traj := range trajectories {
		if traj.Approach == nil {
			continue
		}

		// Count tools
		for _, tool := range traj.Approach.ToolSequence {
			toolFrequency[tool]++
		}

		// Count modes
		for _, mode := range traj.Approach.ModesUsed {
			modeFrequency[mode]++
		}

		// Count strategy
		strategyFrequency[traj.Approach.Strategy]++
	}

	// Find most common strategy
	mostCommonStrategy := findMostFrequent(strategyFrequency)

	// Find most common tools (appearing in >50% of trajectories)
	threshold := len(trajectories) / 2
	commonTools := make([]string, 0)
	for tool, count := range toolFrequency {
		if count >= threshold {
			commonTools = append(commonTools, tool)
		}
	}

	// Find most common modes
	commonModes := make([]string, 0)
	for mode, count := range modeFrequency {
		if count >= threshold {
			commonModes = append(commonModes, mode)
		}
	}

	return &ApproachDescription{
		Strategy:     mostCommonStrategy,
		ModesUsed:    commonModes,
		ToolSequence: commonTools,
		KeyDecisions: []string{"Pattern derived from successful trajectories"},
	}
}

// createProblemSignature creates a signature for problem matching
func createProblemSignature(problem *ProblemDescription) *ProblemSignature {
	if problem == nil {
		return &ProblemSignature{
			Domain:      "unknown",
			ProblemType: "unknown",
			Hash:        "unknown",
		}
	}

	// Extract keywords from description
	keywords := extractKeywords(problem.Description)

	// Infer capabilities needed
	capabilities := inferRequiredCapabilities(problem)

	signature := &ProblemSignature{
		Domain:               problem.Domain,
		ProblemType:          problem.ProblemType,
		ComplexityRange:      [2]float64{maxFloat(0, problem.Complexity-0.2), minFloat(1.0, problem.Complexity+0.2)},
		RequiredCapabilities: capabilities,
		KeywordFingerprint:   keywords,
	}

	// Compute hash
	signature.Hash = computeSignatureHash(signature)

	return signature
}

// extractKeywords extracts important keywords from text
func extractKeywords(text string) []string {
	// Simple keyword extraction (in production, use NLP)
	// For now, return empty to avoid dependencies
	return []string{}
}

// inferRequiredCapabilities infers what capabilities are needed for a problem
func inferRequiredCapabilities(problem *ProblemDescription) []string {
	capabilities := make([]string, 0)

	// Check goals for capability hints
	for _, goal := range problem.Goals {
		if contains(goal, "decide") || contains(goal, "choose") {
			capabilities = append(capabilities, "decision-making")
		}
		if contains(goal, "analyze") || contains(goal, "understand") {
			capabilities = append(capabilities, "analysis")
		}
		if contains(goal, "create") || contains(goal, "generate") {
			capabilities = append(capabilities, "generation")
		}
		if contains(goal, "validate") || contains(goal, "verify") {
			capabilities = append(capabilities, "validation")
		}
	}

	// Check problem type
	switch problem.ProblemType {
	case "causal-analysis":
		capabilities = append(capabilities, "causal-reasoning")
	case "probabilistic":
		capabilities = append(capabilities, "probabilistic-reasoning")
	case "creative":
		capabilities = append(capabilities, "divergent-thinking")
	case "logical":
		capabilities = append(capabilities, "logical-reasoning")
	}

	return capabilities
}

// computeSignatureHash creates a hash for the signature
func computeSignatureHash(sig *ProblemSignature) string {
	data := fmt.Sprintf("%s|%s|%.1f-%.1f", sig.Domain, sig.ProblemType, sig.ComplexityRange[0], sig.ComplexityRange[1])
	return computeProblemHash(&ProblemDescription{
		Domain:      sig.Domain,
		ProblemType: sig.ProblemType,
		Description: data,
	})
}

// generatePatternID generates a unique pattern ID
func generatePatternID(sig *ProblemSignature) string {
	return fmt.Sprintf("pattern_%s_%d", sig.Hash[:8], time.Now().Unix())
}

// generatePatternName creates a readable name for the pattern
func generatePatternName(sig *ProblemSignature, approach *ApproachDescription) string {
	name := sig.Domain
	if name == "" {
		name = "General"
	}

	if sig.ProblemType != "" {
		name += " " + sig.ProblemType
	}

	if approach != nil && approach.Strategy != "" {
		name += " using " + approach.Strategy
	}

	return name
}

// generatePatternDescription generates a description
func generatePatternDescription(trajectories []*ReasoningTrajectory) string {
	if len(trajectories) == 0 {
		return "No description available"
	}

	avgSuccess := 0.0
	for _, traj := range trajectories {
		avgSuccess += traj.SuccessScore
	}
	avgSuccess /= float64(len(trajectories))

	return fmt.Sprintf("Pattern learned from %d successful trajectories with %.0f%% average success rate",
		len(trajectories), avgSuccess*100)
}

// extractTrajectoryIDs extracts up to limit trajectory IDs
func extractTrajectoryIDs(trajectories []*ReasoningTrajectory, limit int) []string {
	ids := make([]string, 0, min(len(trajectories), limit))
	for i, traj := range trajectories {
		if i >= limit {
			break
		}
		ids = append(ids, traj.ID)
	}
	return ids
}

// extractCommonTags finds tags that appear frequently
func extractCommonTags(trajectories []*ReasoningTrajectory) []string {
	tagCount := make(map[string]int)

	for _, traj := range trajectories {
		for _, tag := range traj.Tags {
			tagCount[tag]++
		}
	}

	// Keep tags that appear in >50% of trajectories
	threshold := len(trajectories) / 2
	commonTags := make([]string, 0)
	for tag, count := range tagCount {
		if count >= threshold {
			commonTags = append(commonTags, tag)
		}
	}

	return commonTags
}

// findMostFrequent finds the most frequent item in a frequency map
func findMostFrequent(frequency map[string]int) string {
	maxCount := 0
	mostFrequent := ""

	for item, count := range frequency {
		if count > maxCount {
			maxCount = count
			mostFrequent = item
		}
	}

	return mostFrequent
}

// GetLearnedPatterns retrieves patterns matching a problem
func (l *LearningEngine) GetLearnedPatterns(ctx context.Context, problem *ProblemDescription) ([]*TrajectoryPattern, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	signature := createProblemSignature(problem)
	patterns := make([]*TrajectoryPattern, 0)

	// Find matching patterns
	for _, pattern := range l.patternCache {
		if matchesSignature(signature, pattern.ProblemSignature) {
			patterns = append(patterns, pattern)
		}
	}

	// Sort by success rate
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].SuccessRate > patterns[j].SuccessRate
	})

	return patterns, nil
}

// matchesSignature checks if two signatures match
func matchesSignature(sig1, sig2 *ProblemSignature) bool {
	// Domain match
	if sig1.Domain != sig2.Domain {
		return false
	}

	// Problem type match (or either unknown)
	if sig1.ProblemType != sig2.ProblemType &&
		sig1.ProblemType != "unknown" && sig2.ProblemType != "unknown" {
		return false
	}

	// Complexity overlap
	if sig1.ComplexityRange[1] < sig2.ComplexityRange[0] ||
		sig1.ComplexityRange[0] > sig2.ComplexityRange[1] {
		return false
	}

	return true
}

// Helper functions

func contains(text, substring string) bool {
	return len(text) > 0 && len(substring) > 0 && 
		containsHelper(text, substring)
}

func containsHelper(text, substring string) bool {
	for i := 0; i <= len(text)-len(substring); i++ {
		match := true
		for j := 0; j < len(substring); j++ {
			if toLower(text[i+j]) != toLower(substring[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxFloat is defined in episodic.go
