// Package metacognition provides unknown unknowns detection.
// This module identifies blind spots, missing assumptions, and unconsidered factors
// in reasoning processes.
package metacognition

import (
	"context"
	"fmt"
	"strings"
	"time"

	"unified-thinking/internal/types"
)

// UnknownUnknownsDetector identifies blind spots and knowledge gaps
type UnknownUnknownsDetector struct {
	knownPatterns    map[string]*BlindSpotPattern
	domainChecklists map[string]*DomainChecklist
}

// NewUnknownUnknownsDetector creates a new detector
func NewUnknownUnknownsDetector() *UnknownUnknownsDetector {
	detector := &UnknownUnknownsDetector{
		knownPatterns:    make(map[string]*BlindSpotPattern),
		domainChecklists: make(map[string]*DomainChecklist),
	}

	// Initialize with common blind spot patterns
	detector.initializePatterns()

	return detector
}

// BlindSpot represents a detected unknown unknown
type BlindSpot struct {
	ID          string
	Type        BlindSpotType
	Description string
	Severity    float64 // 0-1, higher = more critical
	Indicators  []string
	Suggestions []string
	DetectedAt  time.Time
	Metadata    map[string]interface{}
}

// BlindSpotType categorizes types of blind spots
type BlindSpotType string

const (
	BlindSpotMissingAssumption   BlindSpotType = "missing_assumption"
	BlindSpotUnconsideredFactor  BlindSpotType = "unconsidered_factor"
	BlindSpotNarrowFraming       BlindSpotType = "narrow_framing"
	BlindSpotConfirmationBias    BlindSpotType = "confirmation_bias"
	BlindSpotIncompleteAnalysis  BlindSpotType = "incomplete_analysis"
	BlindSpotMissingPerspective  BlindSpotType = "missing_perspective"
	BlindSpotOverconfidence      BlindSpotType = "overconfidence"
	BlindSpotUnconsideredRisk    BlindSpotType = "unconsidered_risk"
)

// BlindSpotPattern represents a known pattern that often indicates blind spots
type BlindSpotPattern struct {
	ID          string
	Type        BlindSpotType
	Description string
	Indicators  []string // Keywords or patterns that suggest this blind spot
	Questions   []string // Questions to ask to probe for this blind spot
}

// DomainChecklist provides domain-specific considerations
type DomainChecklist struct {
	Domain      string
	Categories  []string
	MustConsider []string // Things that must be considered in this domain
	CommonGaps  []string // Common blind spots in this domain
}

// GapAnalysisRequest contains parameters for gap analysis
type GapAnalysisRequest struct {
	Content     string
	Domain      string
	Context     string
	Assumptions []string
	Confidence  float64
}

// GapAnalysisResult contains detected blind spots and gaps
type GapAnalysisResult struct {
	BlindSpots          []*BlindSpot
	MissingConsiderations []string
	UnchallengedAssumptions []string
	SuggestedQuestions  []string
	OverallRisk         float64 // 0-1, risk of missing critical factors
	Confidence          float64
	Analysis            string
}

// DetectBlindSpots analyzes content for unknown unknowns
func (uud *UnknownUnknownsDetector) DetectBlindSpots(ctx context.Context, req *GapAnalysisRequest) (*GapAnalysisResult, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content required")
	}

	result := &GapAnalysisResult{
		BlindSpots:              make([]*BlindSpot, 0),
		MissingConsiderations:   make([]string, 0),
		UnchallengedAssumptions: make([]string, 0),
		SuggestedQuestions:      make([]string, 0),
		Confidence:              0.7,
	}

	// 1. Check for pattern-based blind spots
	patternSpots := uud.detectPatternBasedBlindSpots(req)
	result.BlindSpots = append(result.BlindSpots, patternSpots...)

	// 2. Analyze for missing assumptions
	implicitAssumptions := uud.extractImplicitAssumptions(req.Content)
	result.UnchallengedAssumptions = implicitAssumptions

	// 3. Check domain-specific considerations
	if req.Domain != "" {
		domainGaps := uud.checkDomainConsiderations(req)
		result.MissingConsiderations = append(result.MissingConsiderations, domainGaps...)
	}

	// 4. Identify framing limitations
	framingIssues := uud.analyzeFraming(req.Content)
	result.BlindSpots = append(result.BlindSpots, framingIssues...)

	// 5. Check for overconfidence indicators
	if req.Confidence > 0.9 {
		overconfidenceSpot := uud.checkOverconfidence(req)
		if overconfidenceSpot != nil {
			result.BlindSpots = append(result.BlindSpots, overconfidenceSpot)
		}
	}

	// 6. Generate probing questions
	result.SuggestedQuestions = uud.generateProbingQuestions(req, result.BlindSpots)

	// 7. Calculate overall risk
	result.OverallRisk = uud.calculateOverallRisk(result.BlindSpots, req.Confidence)

	// 8. Generate analysis summary
	result.Analysis = uud.summarizeAnalysis(result)

	return result, nil
}

// Helper methods

func (uud *UnknownUnknownsDetector) detectPatternBasedBlindSpots(req *GapAnalysisRequest) []*BlindSpot {
	spots := make([]*BlindSpot, 0)
	contentLower := strings.ToLower(req.Content)

	for _, pattern := range uud.knownPatterns {
		match := false
		matchedIndicators := make([]string, 0)

		for _, indicator := range pattern.Indicators {
			if strings.Contains(contentLower, strings.ToLower(indicator)) {
				match = true
				matchedIndicators = append(matchedIndicators, indicator)
			}
		}

		if match {
			spot := &BlindSpot{
				ID:          fmt.Sprintf("blindspot-%d", time.Now().UnixNano()),
				Type:        pattern.Type,
				Description: pattern.Description,
				Severity:    0.6, // Default severity
				Indicators:  matchedIndicators,
				Suggestions: pattern.Questions,
				DetectedAt:  time.Now(),
				Metadata:    map[string]interface{}{"pattern_id": pattern.ID},
			}
			spots = append(spots, spot)
		}
	}

	return spots
}

func (uud *UnknownUnknownsDetector) extractImplicitAssumptions(content string) []string {
	assumptions := make([]string, 0)

	// Look for absolute statements (often hide assumptions)
	absoluteWords := []string{"always", "never", "must", "can't", "impossible", "certain", "definitely"}
	for _, word := range absoluteWords {
		if strings.Contains(strings.ToLower(content), word) {
			assumptions = append(assumptions, fmt.Sprintf("Absolute statement using '%s' may hide underlying assumptions", word))
		}
	}

	// Look for causal claims without justification
	causalWords := []string{"because", "therefore", "thus", "hence", "consequently"}
	for _, word := range causalWords {
		if strings.Contains(strings.ToLower(content), word) {
			assumptions = append(assumptions, fmt.Sprintf("Causal claim with '%s' - is the causal link justified?", word))
		}
	}

	// Look for generalizations
	if strings.Contains(content, "all ") || strings.Contains(content, "every ") {
		assumptions = append(assumptions, "Generalization detected - are there exceptions?")
	}

	return assumptions
}

func (uud *UnknownUnknownsDetector) checkDomainConsiderations(req *GapAnalysisRequest) []string {
	gaps := make([]string, 0)

	checklist, exists := uud.domainChecklists[req.Domain]
	if !exists {
		return gaps
	}

	contentLower := strings.ToLower(req.Content)

	// Check if required considerations are mentioned
	for _, mustConsider := range checklist.MustConsider {
		if !strings.Contains(contentLower, strings.ToLower(mustConsider)) {
			gaps = append(gaps, fmt.Sprintf("Missing consideration: %s", mustConsider))
		}
	}

	return gaps
}

func (uud *UnknownUnknownsDetector) analyzeFraming(content string) []*BlindSpot {
	spots := make([]*BlindSpot, 0)

	// Check for narrow framing indicators
	narrowIndicators := []string{
		"only option",
		"no alternative",
		"must be",
		"single solution",
	}

	contentLower := strings.ToLower(content)
	for _, indicator := range narrowIndicators {
		if strings.Contains(contentLower, indicator) {
			spot := &BlindSpot{
				ID:          fmt.Sprintf("blindspot-%d", time.Now().UnixNano()),
				Type:        BlindSpotNarrowFraming,
				Description: "Content suggests narrow framing - may be missing alternative approaches",
				Severity:    0.7,
				Indicators:  []string{indicator},
				Suggestions: []string{
					"What alternative approaches exist?",
					"What if the opposite were true?",
					"Who might frame this differently?",
				},
				DetectedAt: time.Now(),
				Metadata:   make(map[string]interface{}),
			}
			spots = append(spots, spot)
			break // Only add one framing spot
		}
	}

	return spots
}

func (uud *UnknownUnknownsDetector) checkOverconfidence(req *GapAnalysisRequest) *BlindSpot {
	// High confidence with short content suggests possible overconfidence
	contentLength := len(req.Content)
	if contentLength < 200 && req.Confidence > 0.9 {
		return &BlindSpot{
			ID:          fmt.Sprintf("blindspot-%d", time.Now().UnixNano()),
			Type:        BlindSpotOverconfidence,
			Description: "High confidence with limited analysis - potential overconfidence",
			Severity:    0.8,
			Indicators:  []string{fmt.Sprintf("Confidence: %.2f, Content length: %d", req.Confidence, contentLength)},
			Suggestions: []string{
				"What could go wrong?",
				"What are you most uncertain about?",
				"What evidence would change your mind?",
			},
			DetectedAt: time.Now(),
			Metadata:   map[string]interface{}{"confidence": req.Confidence, "content_length": contentLength},
		}
	}

	return nil
}

func (uud *UnknownUnknownsDetector) generateProbingQuestions(req *GapAnalysisRequest, blindSpots []*BlindSpot) []string {
	questions := make([]string, 0)

	// Standard probing questions
	questions = append(questions,
		"What assumptions are you making that might not hold?",
		"What factors haven't been considered?",
		"Who might disagree with this analysis and why?",
		"What would need to be true for this to be wrong?",
	)

	// Add questions from detected blind spots
	for _, spot := range blindSpots {
		questions = append(questions, spot.Suggestions...)
	}

	// Limit to 10 questions
	if len(questions) > 10 {
		questions = questions[:10]
	}

	return questions
}

func (uud *UnknownUnknownsDetector) calculateOverallRisk(blindSpots []*BlindSpot, confidence float64) float64 {
	if len(blindSpots) == 0 {
		return 0.1 // Minimal risk
	}

	// Average severity of blind spots
	totalSeverity := 0.0
	for _, spot := range blindSpots {
		totalSeverity += spot.Severity
	}
	avgSeverity := totalSeverity / float64(len(blindSpots))

	// Adjust for confidence (high confidence with blind spots is risky)
	risk := avgSeverity
	if confidence > 0.8 {
		risk = avgSeverity * 1.2 // Increase risk if overconfident
	}

	// Cap at 1.0
	if risk > 1.0 {
		risk = 1.0
	}

	return risk
}

func (uud *UnknownUnknownsDetector) summarizeAnalysis(result *GapAnalysisResult) string {
	summary := fmt.Sprintf("Detected %d potential blind spots.\n", len(result.BlindSpots))

	if len(result.BlindSpots) > 0 {
		summary += "Key concerns:\n"
		for i, spot := range result.BlindSpots {
			if i >= 3 {
				break // Limit to top 3
			}
			summary += fmt.Sprintf("- %s: %s\n", spot.Type, spot.Description)
		}
	}

	if len(result.MissingConsiderations) > 0 {
		summary += fmt.Sprintf("\n%d missing domain-specific considerations identified.\n", len(result.MissingConsiderations))
	}

	summary += fmt.Sprintf("\nOverall risk level: %.2f\n", result.OverallRisk)

	if result.OverallRisk > 0.7 {
		summary += "⚠️ HIGH RISK - Significant blind spots detected. Recommend deeper analysis."
	} else if result.OverallRisk > 0.4 {
		summary += "⚠️ MODERATE RISK - Some blind spots detected. Consider addressing them."
	} else {
		summary += "✓ LOW RISK - Few blind spots detected, but remain vigilant."
	}

	return summary
}

func (uud *UnknownUnknownsDetector) initializePatterns() {
	// Pattern 1: Confirmation bias
	uud.knownPatterns["confirmation_bias"] = &BlindSpotPattern{
		ID:          "confirmation_bias",
		Type:        BlindSpotConfirmationBias,
		Description: "May be seeking evidence that confirms existing beliefs",
		Indicators:  []string{"confirms", "proves my point", "as expected", "validates"},
		Questions: []string{
			"What evidence would contradict this?",
			"What alternative explanations exist?",
		},
	}

	// Pattern 2: Incomplete analysis
	uud.knownPatterns["incomplete"] = &BlindSpotPattern{
		ID:          "incomplete",
		Type:        BlindSpotIncompleteAnalysis,
		Description: "Analysis appears incomplete or rushed",
		Indicators:  []string{"quick analysis", "briefly", "at a glance", "seems like"},
		Questions: []string{
			"What deeper analysis is needed?",
			"What details are missing?",
		},
	}

	// Pattern 3: Missing perspectives
	uud.knownPatterns["single_perspective"] = &BlindSpotPattern{
		ID:          "single_perspective",
		Type:        BlindSpotMissingPerspective,
		Description: "Only one perspective considered",
		Indicators:  []string{"obviously", "clearly", "everyone knows", "it's clear that"},
		Questions: []string{
			"Who might see this differently?",
			"What cultural or contextual factors are missing?",
		},
	}

	// Initialize domain checklists
	uud.domainChecklists["software"] = &DomainChecklist{
		Domain: "software",
		MustConsider: []string{
			"scalability", "security", "performance", "maintainability",
			"testing", "documentation", "error handling",
		},
		CommonGaps: []string{
			"Not considering edge cases",
			"Ignoring backward compatibility",
			"Underestimating technical debt",
		},
	}

	uud.domainChecklists["business"] = &DomainChecklist{
		Domain: "business",
		MustConsider: []string{
			"cost", "revenue", "market", "competition",
			"customer", "risk", "timeline",
		},
		CommonGaps: []string{
			"Not considering opportunity cost",
			"Ignoring market dynamics",
			"Underestimating implementation time",
		},
	}
}

// IdentifyKnowledgeGaps analyzes thought for missing knowledge
func (uud *UnknownUnknownsDetector) IdentifyKnowledgeGaps(thought *types.Thought) ([]string, error) {
	gaps := make([]string, 0)

	// Check for uncertainty markers
	uncertaintyMarkers := []string{"unsure", "unclear", "don't know", "not certain", "maybe", "perhaps"}
	contentLower := strings.ToLower(thought.Content)

	for _, marker := range uncertaintyMarkers {
		if strings.Contains(contentLower, marker) {
			gaps = append(gaps, fmt.Sprintf("Expressed uncertainty: '%s'", marker))
		}
	}

	// Check for questions in content (indicates gaps)
	if strings.Contains(thought.Content, "?") {
		gaps = append(gaps, "Questions present - indicates knowledge gaps")
	}

	// Low confidence suggests gaps
	if thought.Confidence < 0.5 {
		gaps = append(gaps, fmt.Sprintf("Low confidence (%.2f) suggests significant knowledge gaps", thought.Confidence))
	}

	return gaps, nil
}
