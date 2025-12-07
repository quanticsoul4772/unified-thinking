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
	storage   storage.Storage
	llmClient HypothesisGenerator
}

// HypothesisGenerator interface for LLM-based hypothesis generation
type HypothesisGenerator interface {
	GenerateHypotheses(ctx context.Context, prompt string) (string, error)
}

// NewAbductiveReasoner creates a new abductive reasoner
func NewAbductiveReasoner(store storage.Storage, llm HypothesisGenerator) *AbductiveReasoner {
	return &AbductiveReasoner{
		storage:   store,
		llmClient: llm,
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
	ID                   string
	Description          string
	Observations         []string // IDs of observations this explains
	ExplanatoryPower     float64  // How well it explains observations (0-1)
	Parsimony            float64  // Simplicity score (0-1, higher = simpler)
	PriorProbability     float64  // Prior plausibility (0-1)
	PosteriorProbability float64  // Updated probability after evidence
	Assumptions          []string // Required assumptions
	Predictions          []string // Testable predictions
	CompetingIDs         []string // IDs of competing hypotheses
	Status               HypothesisStatus
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Metadata             map[string]interface{}
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
	ID               string
	Observations     []*Observation
	Hypotheses       []*Hypothesis
	BestHypothesis   *Hypothesis
	RankedHypotheses []*Hypothesis // Sorted by posterior probability
	InferenceMethod  string
	Confidence       float64 // Overall confidence in the inference
	Timestamp        time.Time
	Metadata         map[string]interface{}
}

// GenerateHypothesesRequest contains parameters for hypothesis generation
type GenerateHypothesesRequest struct {
	Observations    []*Observation
	MaxHypotheses   int     // Maximum number to generate
	MinParsimony    float64 // Minimum simplicity threshold
	RequireTestable bool    // Whether hypotheses must be testable
	Context         string
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
	MethodBayesian    EvaluationMethod = "bayesian"    // Bayesian inference
	MethodParsimony   EvaluationMethod = "parsimony"   // Occam's Razor
	MethodExplanatory EvaluationMethod = "explanatory" // Explanatory power
	MethodCombined    EvaluationMethod = "combined"    // Weighted combination
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

	if ar.llmClient == nil {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY required: hypothesis generation uses LLM")
	}

	// Build prompt for LLM hypothesis generation
	prompt := ar.buildHypothesisPrompt(req)

	// Generate hypotheses using LLM
	response, err := ar.llmClient.GenerateHypotheses(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM hypothesis generation failed: %w", err)
	}

	// Parse LLM response into structured hypotheses
	hypotheses, err := ar.parseHypothesesFromLLM(response, req.Observations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hypotheses: %w", err)
	}

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

// findCommonThemes extracts meaningful common themes from observation descriptions
// Uses bigrams and filters stop words for better theme extraction
func (ar *AbductiveReasoner) findCommonThemes(observations []*Observation) []string {
	// Track both single words and bigrams
	wordCount := make(map[string]int)
	bigramCount := make(map[string]int)

	for _, obs := range observations {
		// Clean and tokenize
		text := strings.ToLower(obs.Description)
		// Remove punctuation
		text = strings.Map(func(r rune) rune {
			if r == '.' || r == ',' || r == ':' || r == ';' || r == '!' || r == '?' {
				return ' '
			}
			return r
		}, text)

		words := strings.Fields(text)

		// Count meaningful single words
		seenWords := make(map[string]bool)
		for _, word := range words {
			// Clean the word
			word = strings.Trim(word, "\"'()[]{}*-_")
			if len(word) < 3 { // Skip very short words
				continue
			}
			if ar.isStopWord(word) {
				continue
			}
			if !seenWords[word] {
				seenWords[word] = true
				wordCount[word]++
			}
		}

		// Count bigrams (pairs of meaningful words that are adjacent in original)
		for i := 0; i < len(words)-1; i++ {
			w1 := strings.Trim(words[i], "\"'()[]{}*-_")
			w2 := strings.Trim(words[i+1], "\"'()[]{}*-_")

			// Both words should be meaningful
			if len(w1) >= 3 && len(w2) >= 3 && !ar.isStopWord(w1) && !ar.isStopWord(w2) {
				bigram := w1 + " " + w2
				bigramCount[bigram]++
			}
		}
	}

	// Calculate threshold - at least 2 observations or 1/3 of total
	threshold := len(observations) / 3
	if threshold < 2 {
		threshold = 2
	}
	if len(observations) < 3 {
		threshold = 1
	}

	themes := make([]string, 0)

	// First add bigrams that appear frequently (more specific)
	for bigram, count := range bigramCount {
		if count >= threshold {
			themes = append(themes, bigram)
		}
	}

	// Then add single words that appear frequently
	for word, count := range wordCount {
		if count >= threshold && len(word) > 3 {
			// Don't add if already covered by a bigram
			alreadyCovered := false
			for _, theme := range themes {
				if strings.Contains(theme, word) {
					alreadyCovered = true
					break
				}
			}
			if !alreadyCovered {
				themes = append(themes, word)
			}
		}
	}

	// Sort themes by specificity (longer = more specific)
	sort.Slice(themes, func(i, j int) bool {
		return len(themes[i]) > len(themes[j])
	})

	// Limit to top themes
	if len(themes) > 5 {
		themes = themes[:5]
	}

	// If still no themes, try to extract domain-specific terms
	if len(themes) == 0 {
		themes = ar.extractDomainTerms(observations)
	}

	// Final fallback - extract the most unique/specific term
	if len(themes) == 0 {
		// Get the longest word that isn't a stop word
		var longest string
		for word, count := range wordCount {
			if count >= 1 && len(word) > len(longest) {
				longest = word
			}
		}
		if longest != "" {
			themes = []string{longest}
		} else {
			themes = []string{"underlying mechanism"}
		}
	}

	return themes
}

// extractDomainTerms tries to find domain-specific terminology in observations
func (ar *AbductiveReasoner) extractDomainTerms(observations []*Observation) []string {
	// Look for capitalized words (proper nouns, acronyms)
	properNouns := make(map[string]int)

	for _, obs := range observations {
		words := strings.Fields(obs.Description)
		for _, word := range words {
			// Check for capitalized words (not at start of sentence)
			if len(word) >= 2 {
				// Check if word is capitalized
				isCapitalized := word[0] >= 'A' && word[0] <= 'Z'
				isAllCaps := strings.ToUpper(word) == word && len(word) >= 2

				if isCapitalized || isAllCaps {
					cleanWord := strings.Trim(word, "\"'()[]{}*-_.,;:!?")
					if len(cleanWord) >= 2 && !ar.isStopWord(strings.ToLower(cleanWord)) {
						properNouns[cleanWord]++
					}
				}
			}
		}
	}

	terms := make([]string, 0)
	for term, count := range properNouns {
		if count >= 1 {
			terms = append(terms, term)
		}
	}

	// Sort by frequency
	sort.Slice(terms, func(i, j int) bool {
		return properNouns[terms[i]] > properNouns[terms[j]]
	})

	if len(terms) > 3 {
		terms = terms[:3]
	}

	return terms
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

// expandedStopWords contains a comprehensive set of common words that should be filtered out
// when extracting meaningful themes from observations
var expandedStopWords = map[string]bool{
	// Articles and determiners
	"the": true, "a": true, "an": true, "this": true, "that": true, "these": true, "those": true,
	"my": true, "your": true, "his": true, "her": true, "its": true, "our": true, "their": true,
	"some": true, "any": true, "no": true, "every": true, "each": true, "all": true, "both": true,

	// Pronouns
	"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
	"me": true, "him": true, "us": true, "them": true, "who": true, "whom": true, "what": true,
	"which": true, "whose": true, "myself": true, "yourself": true, "itself": true,

	// Common verbs (to be, to have, to do, modals)
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
	"am": true, "have": true, "has": true, "had": true, "having": true, "do": true, "does": true,
	"did": true, "doing": true, "done": true, "will": true, "would": true, "shall": true,
	"should": true, "can": true, "could": true, "may": true, "might": true, "must": true,

	// Prepositions
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true, "with": true,
	"by": true, "from": true, "up": true, "about": true, "into": true, "over": true, "after": true,
	"beneath": true, "under": true, "above": true, "between": true, "among": true, "through": true,
	"during": true, "before": true, "behind": true, "below": true, "against": true,

	// Conjunctions
	"and": true, "or": true, "but": true, "nor": true, "so": true, "yet": true,
	"because": true, "although": true, "unless": true, "since": true, "while": true, "if": true,
	"then": true, "than": true, "when": true, "where": true, "whether": true,

	// Adverbs
	"very": true, "really": true, "quite": true, "rather": true, "too": true, "also": true,
	"just": true, "only": true, "even": true, "still": true, "already": true, "always": true,
	"never": true, "often": true, "sometimes": true, "usually": true, "now": true,
	"here": true, "there": true, "how": true, "why": true, "well": true,
	"more": true, "most": true, "less": true, "least": true,

	// Common adjectives (non-descriptive)
	"other": true, "another": true, "such": true, "own": true, "same": true, "different": true,
	"new": true, "old": true, "good": true, "bad": true, "great": true, "small": true,
	"large": true, "big": true, "little": true, "high": true, "low": true, "long": true,
	"first": true, "last": true, "next": true, "many": true, "much": true, "few": true,

	// Numbers and quantifiers
	"one": true, "two": true, "three": true, "four": true, "five": true, "six": true,
	"seven": true, "eight": true, "nine": true, "ten": true, "several": true, "none": true,

	// Common nouns that rarely carry specific meaning
	"thing": true, "things": true, "stuff": true, "way": true, "ways": true, "time": true,
	"times": true, "year": true, "years": true, "day": true, "days": true, "part": true,
	"parts": true, "place": true, "places": true, "case": true, "cases": true, "point": true,
	"points": true, "fact": true, "facts": true, "kind": true, "kinds": true, "type": true,
	"types": true, "form": true, "forms": true, "number": true, "numbers": true, "amount": true,
	"level": true, "levels": true, "area": true, "areas": true, "side": true, "sides": true,
	"end": true, "ends": true, "matter": true, "matters": true, "issue": true, "issues": true,
	"question": true, "questions": true, "problem": true, "problems": true, "result": true,
	"results": true, "example": true, "examples": true, "reason": true, "reasons": true,
	"idea": true, "ideas": true, "situation": true, "state": true, "states": true,

	// Domain-generic terms that don't add meaning
	"research": true, "study": true, "studies": true, "data": true, "information": true,
	"system": true, "systems": true, "process": true, "processes": true, "method": true,
	"methods": true, "approach": true, "approaches": true, "analysis": true, "work": true,
	"works": true, "effect": true, "effects": true, "impact": true, "change": true,
	"changes": true, "factor": true, "factors": true, "aspect": true, "aspects": true,

	// Common verbs and actions
	"make": true, "made": true, "get": true, "got": true, "go": true, "went": true,
	"come": true, "came": true, "take": true, "took": true, "give": true, "gave": true,
	"find": true, "found": true, "know": true, "knew": true, "think": true, "thought": true,
	"see": true, "saw": true, "want": true, "use": true, "used": true, "try": true,
	"need": true, "seem": true, "help": true, "show": true, "shows": true, "shown": true,
	"lead": true, "leads": true, "led": true, "cause": true, "causes": true, "caused": true,
}

// isStopWord checks if word is a common stop word using comprehensive list
func (ar *AbductiveReasoner) isStopWord(word string) bool {
	return expandedStopWords[word]
}
