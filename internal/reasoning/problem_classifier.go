// Package reasoning provides problem classification using semantic analysis
package reasoning

import (
	"strings"
)

// ProblemType represents different categories of problems
type ProblemType string

const (
	ProblemTypeDecomposable  ProblemType = "decomposable"
	ProblemTypeMeta          ProblemType = "meta"
	ProblemTypeCreative      ProblemType = "creative"
	ProblemTypeEmotional     ProblemType = "emotional"
	ProblemTypeTacit         ProblemType = "tacit"
	ProblemTypeChaotic       ProblemType = "chaotic"
	ProblemTypeValueConflict ProblemType = "value_conflict"
)

// ProblemClassifier uses semantic analysis to classify problem types
type ProblemClassifier struct{}

// NewProblemClassifier creates a new classifier
func NewProblemClassifier() *ProblemClassifier {
	return &ProblemClassifier{}
}

// ClassificationResult contains the classification and reasoning
type ClassificationResult struct {
	Type       ProblemType
	Confidence float64
	Reasoning  string
	Indicators []string
	Approach   string
}

// ClassifyProblem determines what type of problem this is using semantic analysis
func (pc *ProblemClassifier) ClassifyProblem(problem string) *ClassificationResult {
	problemLower := strings.ToLower(problem)

	// Priority 1: Check for meta-questions (highest priority)
	if result := pc.detectMetaQuestion(problemLower); result != nil {
		return result
	}

	// Priority 2: Check for emotional problems
	if result := pc.detectEmotionalProblem(problemLower); result != nil {
		return result
	}

	// Priority 3: Check for value conflicts
	if result := pc.detectValueConflict(problemLower); result != nil {
		return result
	}

	// Priority 4: Check for creative problems
	if result := pc.detectCreativeProblem(problemLower); result != nil {
		return result
	}

	// Priority 5: Check for tacit knowledge
	if result := pc.detectTacitKnowledge(problemLower); result != nil {
		return result
	}

	// Priority 6: Check for chaotic/unpredictable problems
	if result := pc.detectChaoticProblem(problemLower); result != nil {
		return result
	}

	// Default: Decomposable problem
	return &ClassificationResult{
		Type:       ProblemTypeDecomposable,
		Confidence: 0.8,
		Reasoning:  "Problem appears amenable to structured decomposition and analysis",
		Indicators: []string{"structured problem statement"},
		Approach:   "This problem can be decomposed into subproblems using systematic analysis",
	}
}

// detectMetaQuestion detects questions about reasoning/processes/questions themselves
func (pc *ProblemClassifier) detectMetaQuestion(problemLower string) *ClassificationResult {
	// Semantic pattern 1: Questions that query their own domain
	// "What problems..." asking about problems
	// "What questions..." asking about questions
	// "What cannot be..." asking about limits

	hasMetaStructure := false
	indicators := make([]string, 0)

	// Pattern: "what X [verb] Y" where X and Y are in same domain
	metaPatterns := []struct {
		query  string
		domain string
	}{
		{"what problems", "problems"},
		{"what questions", "questions"},
		{"which problems", "problems"},
		{"which questions", "questions"},
		{"what situations", "situations"},
		{"which situations", "situations"},
		{"what scenarios", "scenarios"},
		{"which scenarios", "scenarios"},
		{"what cases", "cases"},
		{"which cases", "cases"},
		{"what circumstances", "circumstances"},
		{"which circumstances", "circumstances"},
		{"what approaches", "approaches"},
		{"which approaches", "approaches"},
		{"what methods", "methods"},
		{"which methods", "methods"},
		{"what types of", "categorization"},
		{"what kind of", "categorization"},
	}

	for _, pattern := range metaPatterns {
		if strings.Contains(problemLower, pattern.query) {
			// Check if it's asking about limits/resistance/impossibility
			if strings.Contains(problemLower, "resist") ||
				strings.Contains(problemLower, "cannot") ||
				strings.Contains(problemLower, "can't") ||
				strings.Contains(problemLower, "don't") ||
				strings.Contains(problemLower, "doesn't") ||
				strings.Contains(problemLower, "won't") ||
				strings.Contains(problemLower, "wouldn't") ||
				strings.Contains(problemLower, "fail") ||
				strings.Contains(problemLower, "beyond") ||
				strings.Contains(problemLower, "limit") ||
				strings.Contains(problemLower, "impossible") ||
				strings.Contains(problemLower, "not suitable") ||
				strings.Contains(problemLower, "not work") ||
				strings.Contains(problemLower, "not benefit") ||
				strings.Contains(problemLower, "aren't suitable") ||
				strings.Contains(problemLower, "isn't suitable") ||
				strings.Contains(problemLower, "aren't work") ||
				strings.Contains(problemLower, "doesn't work") ||
				strings.Contains(problemLower, "isn't benefit") ||
				strings.Contains(problemLower, "aren't benefit") ||
				strings.Contains(problemLower, "inappropriate") ||
				strings.Contains(problemLower, "unsuitable") {
				hasMetaStructure = true
				indicators = append(indicators, "meta-question about "+pattern.domain+" limits")
			}
		}
	}

	// Pattern: Questions about reasoning/decomposition/analysis itself
	reasoningTerms := []string{
		"reasoning",
		"decompos",
		"analysis",
		"structured thinking",
		"formal",
		"systematic",
		"methodical",
	}

	for _, term := range reasoningTerms {
		if strings.Contains(problemLower, term) {
			// If combined with resistance/limits
			if strings.Contains(problemLower, "resist") ||
				strings.Contains(problemLower, "cannot") ||
				strings.Contains(problemLower, "can't") ||
				strings.Contains(problemLower, "don't") ||
				strings.Contains(problemLower, "doesn't") ||
				strings.Contains(problemLower, "won't") ||
				strings.Contains(problemLower, "wouldn't") ||
				strings.Contains(problemLower, "beyond") ||
				strings.Contains(problemLower, "limit") ||
				strings.Contains(problemLower, "fail") ||
				strings.Contains(problemLower, "not suitable") ||
				strings.Contains(problemLower, "not work") ||
				strings.Contains(problemLower, "not benefit") ||
				strings.Contains(problemLower, "aren't suitable") ||
				strings.Contains(problemLower, "isn't suitable") ||
				strings.Contains(problemLower, "aren't work") ||
				strings.Contains(problemLower, "doesn't work") ||
				strings.Contains(problemLower, "isn't benefit") ||
				strings.Contains(problemLower, "aren't benefit") ||
				strings.Contains(problemLower, "inappropriate") ||
				strings.Contains(problemLower, "unsuitable") {
				hasMetaStructure = true
				indicators = append(indicators, "questioning "+term+" itself")
			}
		}
	}

	if !hasMetaStructure {
		return nil
	}

	return &ClassificationResult{
		Type:       ProblemTypeMeta,
		Confidence: 0.9,
		Reasoning:  "This is a meta-question that asks about the limits or nature of reasoning/analysis itself",
		Indicators: indicators,
		Approach:   "This requires philosophical analysis rather than decomposition. Use reflective thinking or analyze-perspectives tool.",
	}
}

// detectEmotionalProblem detects problems involving personal emotions and relationships
func (pc *ProblemClassifier) detectEmotionalProblem(problemLower string) *ClassificationResult {
	// Semantic pattern: Personal pronouns + emotional context

	hasPersonalPronoun := strings.Contains(problemLower, " i ") ||
		strings.Contains(problemLower, "my ") ||
		strings.Contains(problemLower, " me ") ||
		strings.HasPrefix(problemLower, "i ")

	if !hasPersonalPronoun {
		return nil
	}

	indicators := make([]string, 0)
	emotionalScore := 0

	// Emotional verbs and states
	emotionalTerms := []struct {
		term     string
		category string
	}{
		{"feel", "emotional state"},
		{"forgive", "relationship"},
		{"cope", "emotional processing"},
		{"grief", "loss"},
		{"loss", "bereavement"},
		{"betray", "relationship"},
		{"hurt", "emotional pain"},
		{"love", "relationship"},
		{"trust", "relationship"},
		{"guilt", "emotional state"},
		{"shame", "emotional state"},
		{"anger", "emotional state"},
		{"sad", "emotional state"},
		{"anxious", "emotional state"},
		{"worry", "emotional state"},
	}

	for _, term := range emotionalTerms {
		if strings.Contains(problemLower, term.term) {
			emotionalScore++
			indicators = append(indicators, term.category)
		}
	}

	// Decision questions with "should I" in emotional context
	if strings.Contains(problemLower, "should i") && emotionalScore > 0 {
		return &ClassificationResult{
			Type:       ProblemTypeEmotional,
			Confidence: 0.85,
			Reasoning:  "This involves emotional processing and personal decisions that resist formalization",
			Indicators: indicators,
			Approach:   "Emotional problems require human wisdom and reflection, not algorithms. Consider: analyze-perspectives (grief counselor, therapist), analyze-temporal (short vs long-term healing).",
		}
	}

	if emotionalScore >= 2 {
		return &ClassificationResult{
			Type:       ProblemTypeEmotional,
			Confidence: 0.8,
			Reasoning:  "Multiple emotional indicators suggest this problem involves personal feelings and relationships",
			Indicators: indicators,
			Approach:   "Seek human wisdom and emotional reflection rather than systematic decomposition.",
		}
	}

	return nil
}

// detectValueConflict detects problems involving incommensurable values
func (pc *ProblemClassifier) detectValueConflict(problemLower string) *ClassificationResult {
	// Semantic pattern: X vs Y where X and Y are value terms

	if !strings.Contains(problemLower, " vs ") && !strings.Contains(problemLower, " versus ") {
		return nil
	}

	valueTerms := []string{
		"privacy", "security", "freedom", "safety",
		"justice", "mercy", "efficiency", "fairness",
		"growth", "sustainability", "profit", "ethics",
		"innovation", "stability", "individual", "collective",
	}

	valueCount := 0
	indicators := make([]string, 0)

	for _, term := range valueTerms {
		if strings.Contains(problemLower, term) {
			valueCount++
			indicators = append(indicators, term)
		}
	}

	if valueCount >= 2 {
		return &ClassificationResult{
			Type:       ProblemTypeValueConflict,
			Confidence: 0.9,
			Reasoning:  "This involves conflicting values that may be incommensurable",
			Indicators: indicators,
			Approach:   "Value conflicts require reflective equilibrium, not optimization. Use: analyze-perspectives (philosopher, ethicist), analyze-temporal, make-decision with weighted criteria.",
		}
	}

	return nil
}

// detectCreativeProblem detects problems requiring novel/unprecedented solutions
func (pc *ProblemClassifier) detectCreativeProblem(problemLower string) *ClassificationResult {
	indicators := make([]string, 0)
	creativityScore := 0

	// Check for technical/domain contexts that should NOT be classified as creative
	// These are structured domains where "design" and "new" have technical meanings
	technicalContexts := []string{
		// Architecture domain
		"architecture", "microservice", "api", "interface", "system",
		"component", "module", "service", "integration", "infrastructure",
		// Debugging domain
		"debug", "bug", "error", "fix", "crash", "trace", "exception",
		// Research domain
		"research", "study", "analyze", "experiment", "methodology",
		// Proof domain
		"prove", "theorem", "lemma", "proof", "verify", "formal",
	}

	hasTechnicalContext := false
	for _, ctx := range technicalContexts {
		if strings.Contains(problemLower, ctx) {
			hasTechnicalContext = true
			break
		}
	}

	// If technical context is present, don't classify as creative
	// Let it fall through to decomposable (which will get domain-aware decomposition)
	if hasTechnicalContext {
		return nil
	}

	// Semantic pattern: Requests for novelty
	noveltyTerms := []string{
		"new", "novel", "original", "creative",
		"innovative", "unprecedented", "never seen",
		"invent", "design", "imagine", "create from scratch",
	}

	for _, term := range noveltyTerms {
		if strings.Contains(problemLower, term) {
			creativityScore++
			indicators = append(indicators, "novelty: "+term)
		}
	}

	// Pattern: "How to create X that doesn't exist yet"
	if strings.Contains(problemLower, "doesn't exist") ||
		strings.Contains(problemLower, "never existed") ||
		strings.Contains(problemLower, "has never") {
		creativityScore += 2
		indicators = append(indicators, "unprecedented situation")
	}

	if creativityScore >= 2 {
		return &ClassificationResult{
			Type:       ProblemTypeCreative,
			Confidence: 0.85,
			Reasoning:  "This requires creative synthesis and novel thinking, not decomposition of known patterns",
			Indicators: indicators,
			Approach:   "Use divergent thinking and iteration. Try: think (mode='divergent', force_rebellion=true), find-analogy, generate creative alternatives.",
		}
	}

	return nil
}

// detectTacitKnowledge detects problems requiring expert intuition/judgment
func (pc *ProblemClassifier) detectTacitKnowledge(problemLower string) *ClassificationResult {
	indicators := make([]string, 0)
	tacitScore := 0

	// Semantic pattern: Questions about when/how in context-dependent situations
	tacitTerms := []string{
		"when should i", "when to", "how do i know when",
		"gut feeling", "intuition", "instinct",
		"sense", "feel for", "judgment",
		"experience tells", "expert",
	}

	for _, term := range tacitTerms {
		if strings.Contains(problemLower, term) {
			tacitScore++
			indicators = append(indicators, term)
		}
	}

	// Pattern: "when to X" in situational context
	if strings.Contains(problemLower, "when to") || strings.Contains(problemLower, "when should") {
		situationalTerms := []string{
			"bail out", "give up", "pivot",
			"double down", "cut losses", "take risk",
		}

		for _, term := range situationalTerms {
			if strings.Contains(problemLower, term) {
				tacitScore += 2
				indicators = append(indicators, "situational judgment: "+term)
			}
		}
	}

	if tacitScore >= 2 {
		return &ClassificationResult{
			Type:       ProblemTypeTacit,
			Confidence: 0.8,
			Reasoning:  "This relies on tacit knowledge and expert intuition that resists formalization",
			Indicators: indicators,
			Approach:   "Study expert behaviors and case patterns rather than abstract principles. Use: retrieve-similar-cases, perform-cbr-cycle, analyze-perspectives (domain expert).",
		}
	}

	return nil
}

// detectChaoticProblem detects fundamentally unpredictable/chaotic problems
func (pc *ProblemClassifier) detectChaoticProblem(problemLower string) *ClassificationResult {
	indicators := make([]string, 0)
	chaosScore := 0

	// Semantic pattern: Prediction + unpredictability context
	chaosTerms := []string{
		"predict", "forecast", "anticipate",
	}

	unpredictableContexts := []string{
		"market", "stock", "weather", "complex system",
		"emergent", "chaotic", "turbulent", "volatile",
		"black swan", "unexpected",
	}

	hasPrediction := false
	for _, term := range chaosTerms {
		if strings.Contains(problemLower, term) {
			hasPrediction = true
			break
		}
	}

	if hasPrediction {
		for _, context := range unpredictableContexts {
			if strings.Contains(problemLower, context) {
				chaosScore++
				indicators = append(indicators, context)
			}
		}
	}

	if chaosScore >= 1 {
		return &ClassificationResult{
			Type:       ProblemTypeChaotic,
			Confidence: 0.75,
			Reasoning:  "This system exhibits chaotic or fundamentally unpredictable behavior",
			Indicators: indicators,
			Approach:   "Focus on resilience and adaptation rather than prediction. Use: build-causal-graph, simulate-intervention, analyze-temporal (short-term only).",
		}
	}

	return nil
}
