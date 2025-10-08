// Package reasoning provides abductive reasoning capabilities.
// Abductive reasoning is "inference to the best explanation" - generating and evaluating
// hypotheses that best explain a set of observations.
package reasoning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"unified-thinking/internal/storage"
)

// AbductiveReasoner performs inference to the best explanation
type AbductiveReasoner struct {
	storage storage.Storage
}

// NewAbductiveReasoner creates a new abductive reasoner
func NewAbductiveReasoner(store storage.Storage) *AbductiveReasoner {
	return &AbductiveReasoner{
		storage: store,
	}
}

// Observation represents a fact that needs explaining
type Observation struct {
	ID          string
	Description string
	Confidence  float64 // How certain we are about this observation (0-1)
	Timestamp   time.Time
	Context     string
	Metadata    map[string]interface{}
}

// Hypothesis represents a possible explanation
type Hypothesis struct {
	ID                 string
	Description        string
	Observations       []string  // IDs of observations this explains
	ExplanatoryPower   float64   // How well it explains observations (0-1)
	Parsimony          float64   // Simplicity score (0-1, higher = simpler)
	PriorProbability   float64   // Prior plausibility (0-1)
	PosteriorProbability float64 // Updated probability after evidence
	Assumptions        []string  // Required assumptions
	Predictions        []string  // Testable predictions
	CompetingIDs       []string  // IDs of competing hypotheses
	Status             HypothesisStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Metadata           map[string]interface{}
}

// HypothesisStatus represents the state of a hypothesis
type HypothesisStatus string

const (
	StatusProposed  HypothesisStatus = "proposed"
	StatusEvaluated HypothesisStatus = "evaluated"
	StatusSupported HypothesisStatus = "supported"
	StatusRefuted   HypothesisStatus = "refuted"
	StatusRevised   HypothesisStatus = "revised"
)

// AbductiveInference represents the result of abductive reasoning
type AbductiveInference struct {
	ID                string
	Observations      []*Observation
	Hypotheses        []*Hypothesis
	BestHypothesis    *Hypothesis
	RankedHypotheses  []*Hypothesis // Sorted by posterior probability
	InferenceMethod   string
	Confidence        float64 // Overall confidence in the inference
	Timestamp         time.Time
	Metadata          map[string]interface{}
}

// GenerateHypothesesRequest contains parameters for hypothesis generation
type GenerateHypothesesRequest struct {
	Observations     []*Observation
	MaxHypotheses    int     // Maximum number to generate
	MinParsimony     float64 // Minimum simplicity threshold
	RequireTestable  bool    // Whether hypotheses must be testable
	Context          string
}

// EvaluateHypothesesRequest contains parameters for hypothesis evaluation
type EvaluateHypothesesRequest struct {
	Observations []*Observation
	Hypotheses   []*Hypothesis
	Method       EvaluationMethod
	Weights      *EvaluationWeights
}

// EvaluationMethod specifies how to evaluate hypotheses
type EvaluationMethod string

const (
	MethodBayesian    EvaluationMethod = "bayesian"     // Bayesian inference
	MethodParsimony   EvaluationMethod = "parsimony"    // Occam's Razor
	MethodExplanatory EvaluationMethod = "explanatory"  // Explanatory power
	MethodCombined    EvaluationMethod = "combined"     // Weighted combination
)

// EvaluationWeights defines relative importance of evaluation criteria
type EvaluationWeights struct {
	ExplanatoryPower float64 // Default: 0.4
	Parsimony        float64 // Default: 0.3
	PriorProbability float64 // Default: 0.3
}

// DefaultEvaluationWeights returns standard weights
func DefaultEvaluationWeights() *EvaluationWeights {
	return &EvaluationWeights{
		ExplanatoryPower: 0.4,
		Parsimony:        0.3,
		PriorProbability: 0.3,
	}
}

// GenerateHypotheses creates possible explanations for observations
func (ar *AbductiveReasoner) GenerateHypotheses(ctx context.Context, req *GenerateHypothesesRequest) ([]*Hypothesis, error) {
	if len(req.Observations) == 0 {
		return nil, fmt.Errorf("no observations provided")
	}

	hypotheses := make([]*Hypothesis, 0)

	// Strategy 1: Single-cause hypotheses (one explanation for all observations)
	singleCause := ar.generateSingleCauseHypothesis(req.Observations)
	if singleCause != nil {
		hypotheses = append(hypotheses, singleCause)
	}

	// Strategy 2: Multiple-cause hypotheses (different explanations for different observations)
	multipleCauses := ar.generateMultipleCauseHypotheses(req.Observations)
	hypotheses = append(hypotheses, multipleCauses...)

	// Strategy 3: Pattern-based hypotheses (identify patterns in observations)
	patterns := ar.generatePatternHypotheses(req.Observations)
	hypotheses = append(hypotheses, patterns...)

	// Filter by parsimony threshold
	if req.MinParsimony > 0 {
		filtered := make([]*Hypothesis, 0)
		for _, h := range hypotheses {
			if h.Parsimony >= req.MinParsimony {
				filtered = append(filtered, h)
			}
		}
		hypotheses = filtered
	}

	// Limit to max hypotheses
	if req.MaxHypotheses > 0 && len(hypotheses) > req.MaxHypotheses {
		hypotheses = hypotheses[:req.MaxHypotheses]
	}

	return hypotheses, nil
}

// generateSingleCauseHypothesis creates a hypothesis that explains all observations with one cause
func (ar *AbductiveReasoner) generateSingleCauseHypothesis(observations []*Observation) *Hypothesis {
	obsIDs := make([]string, len(observations))
	for i, obs := range observations {
		obsIDs[i] = obs.ID
	}

	// Find common themes in observation descriptions
	commonThemes := ar.findCommonThemes(observations)

	description := fmt.Sprintf("Single common cause: %s", strings.Join(commonThemes, ", "))

	return &Hypothesis{
		ID:               fmt.Sprintf("hyp-single-%d", time.Now().UnixNano()),
		Description:      description,
		Observations:     obsIDs,
		ExplanatoryPower: 0.7, // Default, will be refined
		Parsimony:        0.9, // High parsimony (single cause)
		PriorProbability: 0.5,
		Assumptions:      []string{"All observations share a common cause"},
		Status:           StatusProposed,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}
}

// generateMultipleCauseHypotheses creates hypotheses with multiple independent causes
func (ar *AbductiveReasoner) generateMultipleCauseHypotheses(observations []*Observation) []*Hypothesis {
	hypotheses := make([]*Hypothesis, 0)

	// Group observations by similarity
	groups := ar.groupObservations(observations)

	if len(groups) <= 1 {
		return hypotheses
	}

	// Create hypothesis for each grouping
	obsIDs := make([]string, len(observations))
	for i, obs := range observations {
		obsIDs[i] = obs.ID
	}

	description := fmt.Sprintf("Multiple independent causes (%d groups)", len(groups))

	h := &Hypothesis{
		ID:               fmt.Sprintf("hyp-multi-%d", time.Now().UnixNano()),
		Description:      description,
		Observations:     obsIDs,
		ExplanatoryPower: 0.6,
		Parsimony:        0.5, // Lower parsimony (multiple causes)
		PriorProbability: 0.4,
		Assumptions:      []string{fmt.Sprintf("Observations cluster into %d independent groups", len(groups))},
		Status:           StatusProposed,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         map[string]interface{}{"group_count": len(groups)},
	}

	hypotheses = append(hypotheses, h)

	return hypotheses
}

// generatePatternHypotheses identifies patterns in observations
func (ar *AbductiveReasoner) generatePatternHypotheses(observations []*Observation) []*Hypothesis {
	hypotheses := make([]*Hypothesis, 0)

	// Temporal pattern: check if observations follow a sequence
	if ar.hasTemporalPattern(observations) {
		obsIDs := make([]string, len(observations))
		for i, obs := range observations {
			obsIDs[i] = obs.ID
		}

		h := &Hypothesis{
			ID:               fmt.Sprintf("hyp-pattern-%d", time.Now().UnixNano()),
			Description:      "Temporal sequence pattern detected",
			Observations:     obsIDs,
			ExplanatoryPower: 0.65,
			Parsimony:        0.75,
			PriorProbability: 0.45,
			Assumptions:      []string{"Observations follow a temporal sequence"},
			Predictions:      []string{"Future observations will continue the pattern"},
			Status:           StatusProposed,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			Metadata:         make(map[string]interface{}),
		}

		hypotheses = append(hypotheses, h)
	}

	return hypotheses
}

// EvaluateHypotheses ranks hypotheses by their explanatory quality
func (ar *AbductiveReasoner) EvaluateHypotheses(ctx context.Context, req *EvaluateHypothesesRequest) ([]*Hypothesis, error) {
	if len(req.Hypotheses) == 0 {
		return nil, fmt.Errorf("no hypotheses to evaluate")
	}

	weights := req.Weights
	if weights == nil {
		weights = DefaultEvaluationWeights()
	}

	// Evaluate each hypothesis
	for _, h := range req.Hypotheses {
		ar.evaluateHypothesis(h, req.Observations, req.Method, weights)
	}

	// Sort by posterior probability (highest first)
	sort.Slice(req.Hypotheses, func(i, j int) bool {
		return req.Hypotheses[i].PosteriorProbability > req.Hypotheses[j].PosteriorProbability
	})

	return req.Hypotheses, nil
}

// evaluateHypothesis calculates scores for a single hypothesis
func (ar *AbductiveReasoner) evaluateHypothesis(h *Hypothesis, observations []*Observation, method EvaluationMethod, weights *EvaluationWeights) {
	// Calculate explanatory power
	h.ExplanatoryPower = ar.calculateExplanatoryPower(h, observations)

	// Calculate parsimony (simplicity)
	h.Parsimony = ar.calculateParsimony(h)

	// Calculate posterior probability based on method
	switch method {
	case MethodBayesian:
		h.PosteriorProbability = ar.calculateBayesianProbability(h, observations)
	case MethodParsimony:
		h.PosteriorProbability = h.Parsimony
	case MethodExplanatory:
		h.PosteriorProbability = h.ExplanatoryPower
	case MethodCombined:
		h.PosteriorProbability = weights.ExplanatoryPower*h.ExplanatoryPower +
			weights.Parsimony*h.Parsimony +
			weights.PriorProbability*h.PriorProbability
	default:
		h.PosteriorProbability = weights.ExplanatoryPower*h.ExplanatoryPower +
			weights.Parsimony*h.Parsimony +
			weights.PriorProbability*h.PriorProbability
	}

	h.Status = StatusEvaluated
	h.UpdatedAt = time.Now()
}

// calculateExplanatoryPower measures how well hypothesis explains observations
func (ar *AbductiveReasoner) calculateExplanatoryPower(h *Hypothesis, observations []*Observation) float64 {
	if len(h.Observations) == 0 {
		return 0.0
	}

	// Simple heuristic: proportion of observations explained * average observation confidence
	explained := float64(len(h.Observations))
	total := float64(len(observations))
	coverage := explained / total

	// Weight by observation confidence
	totalConfidence := 0.0
	for _, obs := range observations {
		for _, obsID := range h.Observations {
			if obs.ID == obsID {
				totalConfidence += obs.Confidence
			}
		}
	}
	avgConfidence := totalConfidence / explained

	return coverage * avgConfidence
}

// calculateParsimony measures hypothesis simplicity (Occam's Razor)
func (ar *AbductiveReasoner) calculateParsimony(h *Hypothesis) float64 {
	// Parsimony inversely related to number of assumptions
	assumptionPenalty := 1.0 / (1.0 + float64(len(h.Assumptions)))

	// Simpler descriptions score higher
	descriptionComplexity := float64(len(strings.Fields(h.Description)))
	descriptionScore := 1.0 / (1.0 + math.Log(1.0+descriptionComplexity))

	return (assumptionPenalty + descriptionScore) / 2.0
}

// calculateBayesianProbability applies Bayes' theorem
func (ar *AbductiveReasoner) calculateBayesianProbability(h *Hypothesis, observations []*Observation) float64 {
	// P(H|E) = P(E|H) * P(H) / P(E)
	// Where:
	// P(H|E) = posterior probability
	// P(E|H) = likelihood (how well hypothesis predicts evidence)
	// P(H) = prior probability
	// P(E) = marginal probability of evidence

	likelihood := h.ExplanatoryPower
	prior := h.PriorProbability

	// Simplified: assume P(E) = 0.5 (evidence is equally likely under any hypothesis)
	marginal := 0.5

	posterior := (likelihood * prior) / marginal

	// Clamp to [0, 1]
	if posterior > 1.0 {
		posterior = 1.0
	}
	if posterior < 0.0 {
		posterior = 0.0
	}

	return posterior
}

// PerformAbductiveInference executes full abductive reasoning workflow
func (ar *AbductiveReasoner) PerformAbductiveInference(ctx context.Context, observations []*Observation, maxHypotheses int) (*AbductiveInference, error) {
	// Generate hypotheses
	genReq := &GenerateHypothesesRequest{
		Observations:  observations,
		MaxHypotheses: maxHypotheses,
		MinParsimony:  0.3,
	}

	hypotheses, err := ar.GenerateHypotheses(ctx, genReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hypotheses: %w", err)
	}

	// Evaluate hypotheses
	evalReq := &EvaluateHypothesesRequest{
		Observations: observations,
		Hypotheses:   hypotheses,
		Method:       MethodCombined,
		Weights:      DefaultEvaluationWeights(),
	}

	ranked, err := ar.EvaluateHypotheses(ctx, evalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate hypotheses: %w", err)
	}

	// Select best hypothesis
	var best *Hypothesis
	if len(ranked) > 0 {
		best = ranked[0]
	}

	// Calculate overall confidence
	confidence := 0.0
	if best != nil {
		confidence = best.PosteriorProbability
	}

	inference := &AbductiveInference{
		ID:               fmt.Sprintf("inference-%d", time.Now().UnixNano()),
		Observations:     observations,
		Hypotheses:       hypotheses,
		BestHypothesis:   best,
		RankedHypotheses: ranked,
		InferenceMethod:  string(MethodCombined),
		Confidence:       confidence,
		Timestamp:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}

	return inference, nil
}

// Helper methods

// findCommonThemes extracts common keywords from observation descriptions
func (ar *AbductiveReasoner) findCommonThemes(observations []*Observation) []string {
	wordCount := make(map[string]int)

	for _, obs := range observations {
		words := strings.Fields(strings.ToLower(obs.Description))
		for _, word := range words {
			// Skip common stop words
			if ar.isStopWord(word) {
				continue
			}
			wordCount[word]++
		}
	}

	// Find words that appear in multiple observations
	threshold := len(observations) / 2
	themes := make([]string, 0)
	for word, count := range wordCount {
		if count >= threshold {
			themes = append(themes, word)
		}
	}

	if len(themes) == 0 {
		themes = []string{"common factor"}
	}

	return themes
}

// groupObservations clusters observations by similarity
func (ar *AbductiveReasoner) groupObservations(observations []*Observation) [][]string {
	// Simple grouping: observations with similar timestamps or context
	groups := make([][]string, 0)

	// Group by time proximity (within 1 hour)
	timeGroups := make(map[int64][]string)
	for _, obs := range observations {
		hourBucket := obs.Timestamp.Unix() / 3600
		timeGroups[hourBucket] = append(timeGroups[hourBucket], obs.ID)
	}

	for _, group := range timeGroups {
		if len(group) > 0 {
			groups = append(groups, group)
		}
	}

	if len(groups) == 0 {
		// Default: treat each observation as its own group
		for _, obs := range observations {
			groups = append(groups, []string{obs.ID})
		}
	}

	return groups
}

// hasTemporalPattern checks if observations follow a time sequence
func (ar *AbductiveReasoner) hasTemporalPattern(observations []*Observation) bool {
	if len(observations) < 2 {
		return false
	}

	// Sort by timestamp
	sorted := make([]*Observation, len(observations))
	copy(sorted, observations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	// Check if intervals are relatively consistent
	intervals := make([]time.Duration, 0)
	for i := 1; i < len(sorted); i++ {
		interval := sorted[i].Timestamp.Sub(sorted[i-1].Timestamp)
		intervals = append(intervals, interval)
	}

	// Calculate coefficient of variation
	if len(intervals) == 0 {
		return false
	}

	avg := time.Duration(0)
	for _, interval := range intervals {
		avg += interval
	}
	avg = avg / time.Duration(len(intervals))

	variance := time.Duration(0)
	for _, interval := range intervals {
		diff := interval - avg
		variance += diff * diff / time.Duration(len(intervals))
	}

	stdDev := time.Duration(math.Sqrt(float64(variance)))
	cv := float64(stdDev) / float64(avg)

	// Pattern exists if coefficient of variation < 0.5
	return cv < 0.5
}

// isStopWord checks if word is a common stop word
func (ar *AbductiveReasoner) isStopWord(word string) bool {
	stopWords := []string{"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by", "is", "was", "are", "were", "be", "been", "being"}
	for _, sw := range stopWords {
		if word == sw {
			return true
		}
	}
	return false
}
