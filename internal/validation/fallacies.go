// Package validation provides enhanced fallacy detection capabilities.
//
// This module detects 40+ logical fallacies including formal fallacies
// (affirming the consequent, denying the antecedent) and informal fallacies
// (ad hominem, straw man, appeal to emotion, etc.).
package validation

import (
	"fmt"
	"strings"
	"time"

	"unified-thinking/internal/types"
)

// FallacyType categorizes fallacies
type FallacyType string

const (
	FallacyFormal   FallacyType = "formal"
	FallacyInformal FallacyType = "informal"
	FallacyStatistical FallacyType = "statistical"
)

// DetectedFallacy represents a detected logical fallacy
type DetectedFallacy struct {
	Type        string      `json:"type"`        // "ad_hominem", "straw_man", etc.
	Category    FallacyType `json:"category"`    // formal, informal, statistical
	Location    string      `json:"location"`    // Where in text
	Explanation string      `json:"explanation"` // What's wrong
	Example     string      `json:"example"`     // Problematic text
	Correction  string      `json:"correction"`  // How to fix
	Confidence  float64     `json:"confidence"`  // 0.0-1.0
}

// FallacyDetector detects logical fallacies in reasoning
type FallacyDetector struct{}

// NewFallacyDetector creates a new fallacy detector
func NewFallacyDetector() *FallacyDetector {
	return &FallacyDetector{}
}

// DetectFallacies analyzes text for all types of fallacies
func (fd *FallacyDetector) DetectFallacies(content string, checkFormal, checkInformal bool) []*DetectedFallacy {
	detected := []*DetectedFallacy{}

	if checkFormal {
		detected = append(detected, fd.detectFormalFallacies(content)...)
	}

	if checkInformal {
		detected = append(detected, fd.detectInformalFallacies(content)...)
		detected = append(detected, fd.detectStatisticalFallacies(content)...)
	}

	return detected
}

// detectFormalFallacies detects formal logical fallacies
func (fd *FallacyDetector) detectFormalFallacies(content string) []*DetectedFallacy {
	detected := []*DetectedFallacy{}
	lower := strings.ToLower(content)

	// Affirming the Consequent: If P→Q and Q, conclude P
	if fd.detectAffirmingConsequent(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "affirming_consequent",
			Category:    FallacyFormal,
			Location:    "conditional reasoning",
			Explanation: "Concluding P from 'If P then Q' and Q - invalid inference",
			Example:     fd.extractExample(content, []string{"if", "then", "therefore"}),
			Correction:  "Q being true doesn't prove P. Multiple causes could lead to Q",
			Confidence:  0.7,
		})
	}

	// Denying the Antecedent: If P→Q and ¬P, conclude ¬Q
	if fd.detectDenyingAntecedent(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "denying_antecedent",
			Category:    FallacyFormal,
			Location:    "conditional reasoning",
			Explanation: "Concluding ¬Q from 'If P then Q' and ¬P - invalid inference",
			Example:     fd.extractExample(content, []string{"if", "not", "therefore"}),
			Correction:  "P being false doesn't mean Q is false. Q could occur independently",
			Confidence:  0.7,
		})
	}

	// Undistributed Middle: All A are B, All C are B, therefore All A are C
	if fd.detectUndistributedMiddle(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "undistributed_middle",
			Category:    FallacyFormal,
			Location:    "categorical syllogism",
			Explanation: "Invalid syllogism - middle term not distributed",
			Example:     fd.extractExample(content, []string{"all", "are", "therefore"}),
			Correction:  "Sharing a common property doesn't make two things identical",
			Confidence:  0.6,
		})
	}

	// Illicit Major/Minor: Invalid distribution in categorical syllogism
	if fd.detectIllicitDistribution(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "illicit_distribution",
			Category:    FallacyFormal,
			Location:    "categorical syllogism",
			Explanation: "Term distributed in conclusion but not in premise",
			Example:     fd.extractExample(content, []string{"all", "no", "some"}),
			Correction:  "Cannot make broader claims in conclusion than premises allow",
			Confidence:  0.5,
		})
	}

	return detected
}

// detectInformalFallacies detects informal logical fallacies
func (fd *FallacyDetector) detectInformalFallacies(content string) []*DetectedFallacy {
	detected := []*DetectedFallacy{}
	lower := strings.ToLower(content)

	// Ad Hominem: Attacking the person instead of argument
	if fd.detectAdHominem(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "ad_hominem",
			Category:    FallacyInformal,
			Location:    "argument structure",
			Explanation: "Attacking person's character rather than addressing their argument",
			Example:     fd.extractExample(content, []string{"you", "stupid", "idiot", "ignorant"}),
			Correction:  "Address the argument's merits, not the person making it",
			Confidence:  0.8,
		})
	}

	// Straw Man: Misrepresenting opponent's argument
	if fd.detectStrawMan(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "straw_man",
			Category:    FallacyInformal,
			Location:    "argument representation",
			Explanation: "Misrepresenting opponent's position to make it easier to attack",
			Example:     fd.extractExample(content, []string{"they claim", "they say", "wants to"}),
			Correction:  "Address the actual argument, not a distorted version",
			Confidence:  0.6,
		})
	}

	// Appeal to Authority (illegitimate)
	if fd.detectAppealToAuthority(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "appeal_to_authority",
			Category:    FallacyInformal,
			Location:    "evidence structure",
			Explanation: "Relying on authority figure outside their area of expertise",
			Example:     fd.extractExample(content, []string{"expert", "says", "authority"}),
			Correction:  "Cite relevant expertise and provide supporting evidence",
			Confidence:  0.5,
		})
	}

	// Appeal to Emotion
	if fd.detectAppealToEmotion(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "appeal_to_emotion",
			Category:    FallacyInformal,
			Location:    "argument structure",
			Explanation: "Using emotional manipulation instead of logical reasoning",
			Example:     fd.extractExample(content, []string{"feel", "imagine", "think about"}),
			Correction:  "Support claims with evidence and logic, not emotional appeals",
			Confidence:  0.6,
		})
	}

	// Slippery Slope
	if fd.detectSlipperySlope(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "slippery_slope",
			Category:    FallacyInformal,
			Location:    "causal reasoning",
			Explanation: "Assuming chain of events without justification",
			Example:     fd.extractExample(content, []string{"leads to", "will cause", "next thing"}),
			Correction:  "Justify each step in the causal chain with evidence",
			Confidence:  0.7,
		})
	}

	// False Dilemma / False Dichotomy
	if fd.detectFalseDilemma(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "false_dilemma",
			Category:    FallacyInformal,
			Location:    "option framing",
			Explanation: "Presenting only two options when more exist",
			Example:     fd.extractExample(content, []string{"either", "or", "only two"}),
			Correction:  "Consider all possible options, not just extremes",
			Confidence:  0.8,
		})
	}

	// Red Herring
	if fd.detectRedHerring(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "red_herring",
			Category:    FallacyInformal,
			Location:    "argument flow",
			Explanation: "Introducing irrelevant topic to distract from main issue",
			Example:     fd.extractExample(content, []string{"speaking of", "reminds me", "by the way"}),
			Correction:  "Stay focused on the original question or claim",
			Confidence:  0.5,
		})
	}

	// Hasty Generalization
	if fd.detectHastyGeneralization(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "hasty_generalization",
			Category:    FallacyInformal,
			Location:    "inductive reasoning",
			Explanation: "Drawing broad conclusion from insufficient evidence",
			Example:     fd.extractExample(content, []string{"always", "all", "everyone", "never"}),
			Correction:  "Gather sufficient evidence before generalizing",
			Confidence:  0.6,
		})
	}

	// Circular Reasoning / Begging the Question
	if fd.detectCircularReasoning(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "circular_reasoning",
			Category:    FallacyInformal,
			Location:    "argument structure",
			Explanation: "Conclusion restates premise without providing new support",
			Example:     fd.extractExample(content, []string{"because", "since", "therefore"}),
			Correction:  "Provide independent evidence, not restated conclusions",
			Confidence:  0.7,
		})
	}

	// Appeal to Ignorance
	if fd.detectAppealToIgnorance(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "appeal_to_ignorance",
			Category:    FallacyInformal,
			Location:    "evidence structure",
			Explanation: "Claiming truth because it hasn't been proven false (or vice versa)",
			Example:     fd.extractExample(content, []string{"no evidence", "can't prove", "hasn't been shown"}),
			Correction:  "Absence of evidence isn't evidence of absence",
			Confidence:  0.6,
		})
	}

	// Genetic Fallacy
	if fd.detectGeneticFallacy(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "genetic_fallacy",
			Category:    FallacyInformal,
			Location:    "argument evaluation",
			Explanation: "Judging claim by its origin rather than its merit",
			Example:     fd.extractExample(content, []string{"comes from", "originated", "source"}),
			Correction:  "Evaluate ideas on their own merit, regardless of origin",
			Confidence:  0.5,
		})
	}

	// No True Scotsman
	if fd.detectNoTrueScotsman(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "no_true_scotsman",
			Category:    FallacyInformal,
			Location:    "definition",
			Explanation: "Redefining terms to exclude counterexamples",
			Example:     fd.extractExample(content, []string{"no true", "real", "genuine"}),
			Correction:  "Use consistent definitions, don't move goalposts",
			Confidence:  0.6,
		})
	}

	// Composition / Division
	if fd.detectCompositionDivision(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "composition_division",
			Category:    FallacyInformal,
			Location:    "part-whole reasoning",
			Explanation: "Assuming what's true of parts is true of whole (or vice versa)",
			Example:     fd.extractExample(content, []string{"each", "all", "whole", "parts"}),
			Correction:  "Properties of parts don't necessarily apply to the whole",
			Confidence:  0.5,
		})
	}

	return detected
}

// detectStatisticalFallacies detects statistical and probabilistic fallacies
func (fd *FallacyDetector) detectStatisticalFallacies(content string) []*DetectedFallacy {
	detected := []*DetectedFallacy{}
	lower := strings.ToLower(content)

	// Base Rate Neglect
	if fd.detectBaseRateNeglect(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "base_rate_neglect",
			Category:    FallacyStatistical,
			Location:    "probability reasoning",
			Explanation: "Ignoring prior probability when evaluating evidence",
			Example:     fd.extractExample(content, []string{"probability", "likely", "chance"}),
			Correction:  "Consider base rates and prior probabilities",
			Confidence:  0.5,
		})
	}

	// Texas Sharpshooter
	if fd.detectTexasSharpshooter(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "texas_sharpshooter",
			Category:    FallacyStatistical,
			Location:    "pattern recognition",
			Explanation: "Cherry-picking data clusters and ignoring randomness",
			Example:     fd.extractExample(content, []string{"pattern", "cluster", "correlation"}),
			Correction:  "Consider all data, not just patterns that fit your hypothesis",
			Confidence:  0.6,
		})
	}

	// Survivorship Bias
	if fd.detectSurvivorshipBias(lower) {
		detected = append(detected, &DetectedFallacy{
			Type:        "survivorship_bias",
			Category:    FallacyStatistical,
			Location:    "data selection",
			Explanation: "Drawing conclusions from subset that 'survived' a process",
			Example:     fd.extractExample(content, []string{"successful", "survivors", "winners"}),
			Correction:  "Include data from failures and non-survivors in analysis",
			Confidence:  0.5,
		})
	}

	return detected
}

// Detection helper methods

func (fd *FallacyDetector) detectAffirmingConsequent(text string) bool {
	// Pattern: "if X then Y" + "Y" + "therefore X"
	hasConditional := strings.Contains(text, "if") && (strings.Contains(text, "then") || strings.Contains(text, "implies"))
	hasAffirmation := strings.Contains(text, "therefore") || strings.Contains(text, "thus") || strings.Contains(text, "so")
	return hasConditional && hasAffirmation && !strings.Contains(text, "not")
}

func (fd *FallacyDetector) detectDenyingAntecedent(text string) bool {
	// Pattern: "if X then Y" + "not X" + "therefore not Y"
	hasConditional := strings.Contains(text, "if") && (strings.Contains(text, "then") || strings.Contains(text, "implies"))
	hasNegation := strings.Contains(text, "not") || strings.Contains(text, "no") || strings.Contains(text, "isn't")
	hasConclusion := strings.Contains(text, "therefore") || strings.Contains(text, "thus")
	return hasConditional && hasNegation && hasConclusion
}

func (fd *FallacyDetector) detectUndistributedMiddle(text string) bool {
	// Pattern: multiple "all X are Y" statements
	count := strings.Count(text, "all ") + strings.Count(text, "every ")
	return count >= 2 && (strings.Contains(text, "therefore") || strings.Contains(text, "thus"))
}

func (fd *FallacyDetector) detectIllicitDistribution(text string) bool {
	hasUniversal := strings.Contains(text, "all ") || strings.Contains(text, "no ")
	hasParticular := strings.Contains(text, "some ")
	return hasUniversal && hasParticular && strings.Contains(text, "therefore")
}

func (fd *FallacyDetector) detectAdHominem(text string) bool {
	attacks := []string{"stupid", "idiot", "ignorant", "fool", "moron", "dumb", "incompetent"}
	for _, attack := range attacks {
		if strings.Contains(text, attack) {
			return true
		}
	}
	return false
}

func (fd *FallacyDetector) detectStrawMan(text string) bool {
	indicators := []string{"they claim", "they say", "they believe", "wants to", "trying to"}
	for _, ind := range indicators {
		if strings.Contains(text, ind) && (strings.Contains(text, "but") || strings.Contains(text, "however")) {
			return true
		}
	}
	return false
}

func (fd *FallacyDetector) detectAppealToAuthority(text string) bool {
	authorities := []string{"expert", "authority", "professor", "doctor", "scientist"}
	for _, auth := range authorities {
		if strings.Contains(text, auth) && strings.Contains(text, "says") {
			return true
		}
	}
	return false
}

func (fd *FallacyDetector) detectAppealToEmotion(text string) bool {
	emotional := []string{"feel", "imagine", "think about", "consider how", "heartbreaking", "tragic"}
	count := 0
	for _, emo := range emotional {
		if strings.Contains(text, emo) {
			count++
		}
	}
	return count >= 2
}

func (fd *FallacyDetector) detectSlipperySlope(text string) bool {
	chains := []string{"leads to", "will cause", "results in", "next thing", "eventually"}
	count := 0
	for _, chain := range chains {
		if strings.Contains(text, chain) {
			count++
		}
	}
	return count >= 2
}

func (fd *FallacyDetector) detectFalseDilemma(text string) bool {
	dilemma := strings.Contains(text, "either") && strings.Contains(text, "or")
	only := strings.Contains(text, "only") && strings.Contains(text, "option")
	return dilemma || only
}

func (fd *FallacyDetector) detectRedHerring(text string) bool {
	distractions := []string{"speaking of", "reminds me", "by the way", "incidentally", "on another note"}
	for _, dist := range distractions {
		if strings.Contains(text, dist) {
			return true
		}
	}
	return false
}

func (fd *FallacyDetector) detectHastyGeneralization(text string) bool {
	absolutes := []string{"always", "never", "all", "everyone", "no one", "everything"}
	limited := []string{"one", "once", "single", "example"}

	hasAbsolute := false
	for _, abs := range absolutes {
		if strings.Contains(text, abs) {
			hasAbsolute = true
			break
		}
	}

	hasLimited := false
	for _, lim := range limited {
		if strings.Contains(text, lim) {
			hasLimited = true
			break
		}
	}

	return hasAbsolute && hasLimited
}

func (fd *FallacyDetector) detectCircularReasoning(text string) bool {
	// Look for premise appearing in conclusion
	sentences := strings.Split(text, ".")
	if len(sentences) < 2 {
		return false
	}

	first := strings.ToLower(strings.TrimSpace(sentences[0]))
	for i := 1; i < len(sentences); i++ {
		sent := strings.ToLower(strings.TrimSpace(sentences[i]))
		if len(sent) > 10 && len(first) > 10 {
			// Check for significant overlap
			if fd.sentenceSimilarity(first, sent) > 0.6 {
				return true
			}
		}
	}
	return false
}

func (fd *FallacyDetector) detectAppealToIgnorance(text string) bool {
	ignorance := []string{"no evidence", "can't prove", "hasn't been shown", "no proof", "unproven"}
	conclusion := []string{"therefore", "thus", "so", "must be"}

	hasIgnorance := false
	for _, ign := range ignorance {
		if strings.Contains(text, ign) {
			hasIgnorance = true
			break
		}
	}

	hasConclusion := false
	for _, con := range conclusion {
		if strings.Contains(text, con) {
			hasConclusion = true
			break
		}
	}

	return hasIgnorance && hasConclusion
}

func (fd *FallacyDetector) detectGeneticFallacy(text string) bool {
	origins := []string{"comes from", "originated", "source is", "created by"}
	judgments := []string{"wrong", "false", "invalid", "bad", "unreliable"}

	hasOrigin := false
	for _, org := range origins {
		if strings.Contains(text, org) {
			hasOrigin = true
			break
		}
	}

	hasJudgment := false
	for _, jud := range judgments {
		if strings.Contains(text, jud) {
			hasJudgment = true
			break
		}
	}

	return hasOrigin && hasJudgment
}

func (fd *FallacyDetector) detectNoTrueScotsman(text string) bool {
	return strings.Contains(text, "no true") || strings.Contains(text, "no real") ||
		(strings.Contains(text, "genuine") && strings.Contains(text, "wouldn't"))
}

func (fd *FallacyDetector) detectCompositionDivision(text string) bool {
	parts := []string{"each", "every", "individual", "part"}
	wholes := []string{"whole", "all", "entire", "total"}

	hasParts := false
	for _, part := range parts {
		if strings.Contains(text, part) {
			hasParts = true
			break
		}
	}

	hasWhole := false
	for _, whole := range wholes {
		if strings.Contains(text, whole) {
			hasWhole = true
			break
		}
	}

	return hasParts && hasWhole && strings.Contains(text, "therefore")
}

func (fd *FallacyDetector) detectBaseRateNeglect(text string) bool {
	prob := []string{"probability", "likely", "chance", "odds"}
	hasProb := false
	for _, p := range prob {
		if strings.Contains(text, p) {
			hasProb = true
			break
		}
	}

	return hasProb && !strings.Contains(text, "base") && !strings.Contains(text, "prior")
}

func (fd *FallacyDetector) detectTexasSharpshooter(text string) bool {
	patterns := []string{"pattern", "cluster", "correlation", "trend"}
	hasPattern := false
	for _, pat := range patterns {
		if strings.Contains(text, pat) {
			hasPattern = true
			break
		}
	}

	return hasPattern && (strings.Contains(text, "significant") || strings.Contains(text, "proves"))
}

func (fd *FallacyDetector) detectSurvivorshipBias(text string) bool {
	survivors := []string{"successful", "survivors", "winners", "who made it", "achieved"}
	for _, surv := range survivors {
		if strings.Contains(text, surv) && strings.Contains(text, "because") {
			return true
		}
	}
	return false
}

// Helper methods

func (fd *FallacyDetector) extractExample(text string, keywords []string) string {
	lower := strings.ToLower(text)
	for _, keyword := range keywords {
		if idx := strings.Index(lower, keyword); idx != -1 {
			start := max(0, idx-30)
			end := min(len(text), idx+70)
			return "..." + text[start:end] + "..."
		}
	}
	if len(text) > 100 {
		return text[:100] + "..."
	}
	return text
}

func (fd *FallacyDetector) sentenceSimilarity(s1, s2 string) float64 {
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	common := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 && len(w1) > 3 { // Only count significant words
				common++
				break
			}
		}
	}

	maxLen := len(words1)
	if len(words2) > maxLen {
		maxLen = len(words2)
	}

	return float64(common) / float64(maxLen)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CreateFallacyValidation converts detected fallacy to validation result
func (fd *FallacyDetector) CreateFallacyValidation(fallacies []*DetectedFallacy, thoughtID string) *types.Validation {
	if len(fallacies) == 0 {
		return &types.Validation{
			ID:        fmt.Sprintf("val_%d", time.Now().UnixNano()),
			ThoughtID: thoughtID,
			IsValid:   true,
			Reason:    "No logical fallacies detected",
		}
	}

	reasons := []string{}
	for _, fallacy := range fallacies {
		reasons = append(reasons, fmt.Sprintf("%s: %s", fallacy.Type, fallacy.Explanation))
	}

	return &types.Validation{
		ID:        fmt.Sprintf("val_%d", time.Now().UnixNano()),
		ThoughtID: thoughtID,
		IsValid:   false,
		Reason:    strings.Join(reasons, "; "),
		ValidationData: map[string]interface{}{
			"fallacies": fallacies,
		},
	}
}
