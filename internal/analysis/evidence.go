// Package analysis provides analytical capabilities for evidence assessment,
// sensitivity analysis, and cross-thought analysis.
package analysis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// EvidenceAnalyzer assesses the quality and strength of evidence
type EvidenceAnalyzer struct {
	mu      sync.RWMutex
	counter int
}

// NewEvidenceAnalyzer creates a new evidence analyzer
func NewEvidenceAnalyzer() *EvidenceAnalyzer {
	return &EvidenceAnalyzer{}
}

// AssessEvidence evaluates evidence and assigns quality scores
func (ea *EvidenceAnalyzer) AssessEvidence(content, source string, claimID string, supportsClaim bool) (*types.Evidence, error) {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	ea.counter++

	// Assess quality based on content characteristics
	quality := ea.determineQuality(content, source)

	// Calculate reliability based on source and content
	reliability := ea.calculateReliability(content, source)

	// Calculate relevance based on content length and specificity
	relevance := ea.calculateRelevance(content)

	// Overall score is weighted combination
	overallScore := ea.calculateOverallScore(quality, reliability, relevance)

	evidence := &types.Evidence{
		ID:            fmt.Sprintf("evidence-%d", ea.counter),
		Content:       content,
		Source:        source,
		Quality:       quality,
		Reliability:   reliability,
		Relevance:     relevance,
		OverallScore:  overallScore,
		SupportsClaim: supportsClaim,
		ClaimID:       claimID,
		Metadata:      map[string]interface{}{},
		CreatedAt:     time.Now(),
	}

	return evidence, nil
}

// determineQuality assigns evidence quality category
func (ea *EvidenceAnalyzer) determineQuality(content, source string) types.EvidenceQuality {
	contentLower := strings.ToLower(content)
	sourceLower := strings.ToLower(source)

	// Strong evidence indicators
	strongIndicators := []string{"study", "research", "data", "experiment", "peer-reviewed",
		"statistical", "empirical", "meta-analysis", "systematic review"}

	// Weak evidence indicators
	weakIndicators := []string{"anecdote", "opinion", "feels like", "seems", "probably",
		"might", "maybe", "rumor", "hearsay"}

	strongCount := 0
	for _, indicator := range strongIndicators {
		if strings.Contains(contentLower, indicator) || strings.Contains(sourceLower, indicator) {
			strongCount++
		}
	}

	weakCount := 0
	for _, indicator := range weakIndicators {
		if strings.Contains(contentLower, indicator) {
			weakCount++
		}
	}

	// Determine quality based on indicators
	if weakCount > strongCount {
		if weakCount >= 2 {
			return types.EvidenceAnecdotal
		}
		return types.EvidenceWeak
	}

	if strongCount >= 3 {
		return types.EvidenceStrong
	} else if strongCount >= 1 {
		return types.EvidenceModerate
	}

	return types.EvidenceWeak
}

// calculateReliability estimates reliability score
func (ea *EvidenceAnalyzer) calculateReliability(content, source string) float64 {
	score := 0.5 // Base reliability

	// Increase for credible sources
	sourceLower := strings.ToLower(source)
	credibleSources := []string{"university", "institute", "journal", "government",
		"research", "official", "expert"}

	for _, indicator := range credibleSources {
		if strings.Contains(sourceLower, indicator) {
			score += 0.1
			break
		}
	}

	// Increase for specific, detailed content
	if len(content) > 200 {
		score += 0.1
	}

	// Decrease for vague language
	contentLower := strings.ToLower(content)
	vagueTerms := []string{"maybe", "probably", "possibly", "might", "could be"}
	vagueCount := 0
	for _, term := range vagueTerms {
		if strings.Contains(contentLower, term) {
			vagueCount++
		}
	}
	score -= float64(vagueCount) * 0.05

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// calculateRelevance estimates relevance score
func (ea *EvidenceAnalyzer) calculateRelevance(content string) float64 {
	// Base relevance on content length and specificity
	score := 0.5

	// Longer content tends to be more detailed and relevant
	length := len(content)
	if length > 100 && length < 500 {
		score += 0.2
	} else if length >= 500 {
		score += 0.3
	}

	// Check for specific terms and numbers (indicates precision)
	hasNumbers := strings.ContainsAny(content, "0123456789")
	if hasNumbers {
		score += 0.1
	}

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateOverallScore computes weighted overall score
func (ea *EvidenceAnalyzer) calculateOverallScore(quality types.EvidenceQuality, reliability, relevance float64) float64 {
	// Quality weight: 40%, Reliability: 35%, Relevance: 25%
	qualityScore := 0.0
	switch quality {
	case types.EvidenceStrong:
		qualityScore = 1.0
	case types.EvidenceModerate:
		qualityScore = 0.7
	case types.EvidenceWeak:
		qualityScore = 0.4
	case types.EvidenceAnecdotal:
		qualityScore = 0.2
	}

	overall := (qualityScore * 0.4) + (reliability * 0.35) + (relevance * 0.25)
	return overall
}

// AggregateEvidence combines multiple pieces of evidence
func (ea *EvidenceAnalyzer) AggregateEvidence(evidences []*types.Evidence) *EvidenceAggregation {
	ea.mu.RLock()
	defer ea.mu.RUnlock()

	if len(evidences) == 0 {
		return &EvidenceAggregation{
			TotalCount:       0,
			SupportingCount:  0,
			RefutingCount:    0,
			AverageQuality:   0,
			OverallStrength:  0,
		}
	}

	supportingCount := 0
	refutingCount := 0
	totalQualityScore := 0.0
	totalStrength := 0.0

	for _, ev := range evidences {
		if ev.SupportsClaim {
			supportingCount++
			totalStrength += ev.OverallScore
		} else {
			refutingCount++
			totalStrength -= ev.OverallScore
		}

		// Map quality to numeric score for averaging
		qualityScore := 0.0
		switch ev.Quality {
		case types.EvidenceStrong:
			qualityScore = 1.0
		case types.EvidenceModerate:
			qualityScore = 0.7
		case types.EvidenceWeak:
			qualityScore = 0.4
		case types.EvidenceAnecdotal:
			qualityScore = 0.2
		}
		totalQualityScore += qualityScore
	}

	avgQuality := totalQualityScore / float64(len(evidences))

	// Normalize overall strength to [-1, 1] range then map to [0, 1]
	normalizedStrength := totalStrength / float64(len(evidences))
	overallStrength := (normalizedStrength + 1.0) / 2.0

	return &EvidenceAggregation{
		TotalCount:      len(evidences),
		SupportingCount: supportingCount,
		RefutingCount:   refutingCount,
		AverageQuality:  avgQuality,
		OverallStrength: overallStrength,
	}
}

// EvidenceAggregation represents aggregated evidence assessment
type EvidenceAggregation struct {
	TotalCount      int     `json:"total_count"`
	SupportingCount int     `json:"supporting_count"`
	RefutingCount   int     `json:"refuting_count"`
	AverageQuality  float64 `json:"average_quality"`
	OverallStrength float64 `json:"overall_strength"` // 0.0-1.0
}
