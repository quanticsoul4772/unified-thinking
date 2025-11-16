// Package validation provides hallucination detection and semantic uncertainty measurement.
package validation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// SemanticUncertainty represents different types of uncertainty
type SemanticUncertainty struct {
	Overall            float64         `json:"overall"`             // Overall uncertainty score (0-1)
	Aleatory           float64         `json:"aleatory"`            // Irreducible randomness
	Epistemic          float64         `json:"epistemic"`           // Lack of knowledge
	ConfidenceMismatch float64         `json:"confidence_mismatch"` // Difference between stated and measured confidence
	Type               UncertaintyType `json:"type"`
}

// UncertaintyType categorizes the kind of uncertainty
type UncertaintyType string

const (
	UncertaintyLow      UncertaintyType = "low"      // <0.3
	UncertaintyModerate UncertaintyType = "moderate" // 0.3-0.6
	UncertaintyHigh     UncertaintyType = "high"     // 0.6-0.8
	UncertaintyCritical UncertaintyType = "critical" // >0.8
)

// FactualClaim represents an extracted claim that can be verified
type FactualClaim struct {
	ID                    string               `json:"id"`
	Text                  string               `json:"text"`
	ClaimType             ClaimType            `json:"claim_type"`
	VerificationStatus    VerificationStatus   `json:"verification_status"`
	SemanticUncertainty   float64              `json:"semantic_uncertainty"`
	SupportingEvidence    []string             `json:"supporting_evidence,omitempty"`
	ContradictingEvidence []string             `json:"contradicting_evidence,omitempty"`
	Sources               []VerificationSource `json:"sources,omitempty"`
	VerifiedAt            time.Time            `json:"verified_at"`
}

// ClaimType categorizes the kind of claim
type ClaimType string

const (
	ClaimFactual     ClaimType = "factual"     // Verifiable fact
	ClaimOpinion     ClaimType = "opinion"     // Subjective statement
	ClaimPrediction  ClaimType = "prediction"  // Future-oriented
	ClaimDefinition  ClaimType = "definition"  // Definitional statement
	ClaimCausal      ClaimType = "causal"      // X causes Y
	ClaimStatistical ClaimType = "statistical" // Numerical claim
)

// VerificationStatus indicates the verification result
type VerificationStatus string

const (
	StatusVerified      VerificationStatus = "verified"      // Confirmed true
	StatusFalse         VerificationStatus = "false"         // Confirmed false (hallucination)
	StatusContradictory VerificationStatus = "contradictory" // Mixed evidence
	StatusUnverifiable  VerificationStatus = "unverifiable"  // Cannot verify
	StatusPending       VerificationStatus = "pending"       // Not yet verified
)

// VerificationSource represents where verification info came from
type VerificationSource struct {
	Type       string    `json:"type"`       // "memory", "search", "external_api"
	Source     string    `json:"source"`     // Specific source identifier
	Confidence float64   `json:"confidence"` // How confident in this source
	Timestamp  time.Time `json:"timestamp"`
}

// HallucinationReport contains the full analysis of a thought
type HallucinationReport struct {
	ThoughtID           string              `json:"thought_id"`
	OverallRisk         float64             `json:"overall_risk"` // 0-1, higher = more likely hallucination
	SemanticUncertainty SemanticUncertainty `json:"semantic_uncertainty"`
	Claims              []FactualClaim      `json:"claims"`
	HallucinationCount  int                 `json:"hallucination_count"`
	VerifiedCount       int                 `json:"verified_count"`
	Recommendations     []string            `json:"recommendations"`
	AnalyzedAt          time.Time           `json:"analyzed_at"`
	VerificationLevel   VerificationLevel   `json:"verification_level"` // How deep the analysis was
}

// VerificationLevel indicates depth of verification
type VerificationLevel string

const (
	VerificationFast   VerificationLevel = "fast"   // Quick inline checks only
	VerificationDeep   VerificationLevel = "deep"   // Comprehensive async verification
	VerificationHybrid VerificationLevel = "hybrid" // Both fast + deep
)

// HallucinationDetector provides hallucination detection capabilities
type HallucinationDetector struct {
	mu                 sync.RWMutex
	verificationCache  map[string]*HallucinationReport
	fastCheckThreshold time.Duration
	asyncVerifyQueue   chan *verificationTask
	workersRunning     bool
	knowledgeSources   []KnowledgeSource
}

// KnowledgeSource represents a source for fact verification
type KnowledgeSource interface {
	Verify(ctx context.Context, claim string) (*VerificationResult, error)
	Type() string
	Confidence() float64
}

// VerificationResult contains verification outcome
type VerificationResult struct {
	IsVerified            bool
	Confidence            float64
	SupportingEvidence    []string
	ContradictingEvidence []string
	Source                string
}

type verificationTask struct {
	thoughtID string
	thought   *types.Thought
	resultCh  chan *HallucinationReport
}

// NewHallucinationDetector creates a new hallucination detector
func NewHallucinationDetector() *HallucinationDetector {
	hd := &HallucinationDetector{
		verificationCache:  make(map[string]*HallucinationReport),
		fastCheckThreshold: 100 * time.Millisecond,
		asyncVerifyQueue:   make(chan *verificationTask, 100),
		knowledgeSources:   []KnowledgeSource{},
	}

	// Start async worker pool
	hd.startWorkers(3)

	return hd
}

// RegisterKnowledgeSource adds a knowledge source for verification
func (hd *HallucinationDetector) RegisterKnowledgeSource(source KnowledgeSource) {
	hd.mu.Lock()
	defer hd.mu.Unlock()
	hd.knowledgeSources = append(hd.knowledgeSources, source)
}

// VerifyThought performs hybrid verification on a thought
func (hd *HallucinationDetector) VerifyThought(ctx context.Context, thought *types.Thought) (*HallucinationReport, error) {
	// Check cache first
	if report, ok := hd.getCachedReport(thought.ID); ok {
		return report, nil
	}

	// Fast inline checks
	fastReport := hd.fastVerification(thought)

	// If fast check shows high risk, return immediately with warning
	if fastReport.OverallRisk > 0.7 {
		fastReport.Recommendations = append(fastReport.Recommendations,
			"High hallucination risk detected. Deep verification recommended.")
		hd.cacheReport(fastReport)
		return fastReport, nil
	}

	// Queue for deep async verification
	hd.queueDeepVerification(thought)

	// Return fast report with note that deep verification is pending
	fastReport.Recommendations = append(fastReport.Recommendations,
		"Fast verification complete. Deep verification in progress.")
	fastReport.VerificationLevel = VerificationHybrid

	hd.cacheReport(fastReport)
	return fastReport, nil
}

// fastVerification performs quick inline checks (< 100ms)
func (hd *HallucinationDetector) fastVerification(thought *types.Thought) *HallucinationReport {
	report := &HallucinationReport{
		ThoughtID:         thought.ID,
		AnalyzedAt:        time.Now(),
		VerificationLevel: VerificationFast,
		Claims:            []FactualClaim{},
	}

	// Extract potential factual claims (simple heuristics)
	claims := hd.extractClaims(thought.Content)
	report.Claims = claims

	// Quick coherence checks
	coherence := hd.checkInternalCoherence(thought.Content)

	// Quick confidence calibration check
	confidenceMismatch := hd.estimateConfidenceMismatch(thought)

	// Calculate semantic uncertainty
	report.SemanticUncertainty = SemanticUncertainty{
		Overall:            (1.0-coherence)*0.5 + confidenceMismatch*0.5,
		ConfidenceMismatch: confidenceMismatch,
		Type:               hd.categorizeUncertainty((1.0-coherence)*0.5 + confidenceMismatch*0.5),
	}

	// Calculate overall risk
	report.OverallRisk = report.SemanticUncertainty.Overall

	// Add recommendations based on confidence mismatch or overall risk
	if confidenceMismatch > 0.3 {
		report.Recommendations = append(report.Recommendations,
			"Confidence-language mismatch detected. Review stated confidence level.")
	}

	if report.OverallRisk > 0.5 {
		report.Recommendations = append(report.Recommendations,
			"Moderate hallucination risk. Consider requesting evidence.")
	}

	return report
}

// extractClaims extracts factual claims from text (simple heuristic version)
func (hd *HallucinationDetector) extractClaims(content string) []FactualClaim {
	var claims []FactualClaim

	// Simple sentence splitting
	sentences := strings.Split(content, ". ")

	for i, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) < 5 {
			continue
		}

		// Classify claim type based on keywords
		claimType := hd.classifyClaimType(sentence)

		// Skip opinions for now (can't verify subjectively)
		if claimType == ClaimOpinion {
			continue
		}

		claim := FactualClaim{
			ID:                  fmt.Sprintf("claim-%d", i),
			Text:                sentence,
			ClaimType:           claimType,
			VerificationStatus:  StatusPending,
			SemanticUncertainty: 0.5, // Default uncertainty
			VerifiedAt:          time.Now(),
		}

		claims = append(claims, claim)
	}

	return claims
}

// classifyClaimType attempts to classify the type of claim
func (hd *HallucinationDetector) classifyClaimType(sentence string) ClaimType {
	lower := strings.ToLower(sentence)

	// Causal indicators
	if strings.Contains(lower, "causes") || strings.Contains(lower, "leads to") ||
		strings.Contains(lower, "results in") || strings.Contains(lower, "because") {
		return ClaimCausal
	}

	// Statistical indicators
	if strings.Contains(lower, "%") || strings.Contains(lower, "percent") ||
		strings.Contains(lower, "probability") || strings.Contains(lower, "average") {
		return ClaimStatistical
	}

	// Prediction indicators
	if strings.Contains(lower, "will") || strings.Contains(lower, "going to") ||
		strings.Contains(lower, "predict") || strings.Contains(lower, "expect") {
		return ClaimPrediction
	}

	// Definition indicators
	if strings.Contains(lower, "is defined as") || strings.Contains(lower, "means") ||
		strings.Contains(lower, "refers to") {
		return ClaimDefinition
	}

	// Opinion indicators
	if strings.Contains(lower, "i think") || strings.Contains(lower, "believe") ||
		strings.Contains(lower, "opinion") || strings.Contains(lower, "should") {
		return ClaimOpinion
	}

	// Default to factual
	return ClaimFactual
}

// checkInternalCoherence checks if the text is internally consistent
func (hd *HallucinationDetector) checkInternalCoherence(content string) float64 {
	// Simple heuristic: Check for obvious contradictions
	lower := strings.ToLower(content)

	coherenceScore := 1.0

	// Check for contradiction indicators
	if strings.Contains(lower, "however") && strings.Contains(lower, "but") {
		coherenceScore -= 0.1
	}

	// Check for uncertain language
	uncertainWords := []string{"maybe", "possibly", "might", "could", "perhaps", "unclear"}
	for _, word := range uncertainWords {
		if strings.Contains(lower, word) {
			coherenceScore -= 0.05
		}
	}

	// Ensure score stays in valid range
	if coherenceScore < 0 {
		coherenceScore = 0
	}

	return coherenceScore
}

// estimateConfidenceMismatch estimates if stated confidence matches content quality
func (hd *HallucinationDetector) estimateConfidenceMismatch(thought *types.Thought) float64 {
	// If confidence is high but content has uncertainty markers, that's a mismatch
	lower := strings.ToLower(thought.Content)

	uncertaintyMarkers := 0
	markers := []string{"maybe", "possibly", "might", "unclear", "uncertain", "not sure"}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			uncertaintyMarkers++
		}
	}

	// High confidence with uncertainty markers = mismatch
	if thought.Confidence > 0.7 && uncertaintyMarkers > 0 {
		// Scale based on number of markers, capped at reasonable max
		mismatch := float64(uncertaintyMarkers) * 0.15
		if mismatch > 0.6 {
			mismatch = 0.6
		}
		return mismatch
	}

	// Low confidence with definitive language = also mismatch
	definitiveMarkers := 0
	definitive := []string{"certainly", "definitely", "always", "never", "must"}
	for _, marker := range definitive {
		if strings.Contains(lower, marker) {
			definitiveMarkers++
		}
	}

	if thought.Confidence < 0.5 && definitiveMarkers > 0 {
		mismatch := float64(definitiveMarkers) * 0.15
		if mismatch > 0.5 {
			mismatch = 0.5
		}
		return mismatch
	}

	return 0.0
}

// categorizeUncertainty categorizes uncertainty level
func (hd *HallucinationDetector) categorizeUncertainty(uncertainty float64) UncertaintyType {
	if uncertainty < 0.3 {
		return UncertaintyLow
	} else if uncertainty < 0.6 {
		return UncertaintyModerate
	} else if uncertainty < 0.8 {
		return UncertaintyHigh
	}
	return UncertaintyCritical
}

// queueDeepVerification queues a thought for deep async verification
func (hd *HallucinationDetector) queueDeepVerification(thought *types.Thought) {
	task := &verificationTask{
		thoughtID: thought.ID,
		thought:   thought,
		resultCh:  make(chan *HallucinationReport, 1),
	}

	select {
	case hd.asyncVerifyQueue <- task:
		// Successfully queued
	default:
		// Queue full, skip for now
	}
}

// startWorkers starts the background worker pool
func (hd *HallucinationDetector) startWorkers(numWorkers int) {
	hd.mu.Lock()
	if hd.workersRunning {
		hd.mu.Unlock()
		return
	}
	hd.workersRunning = true
	hd.mu.Unlock()

	for i := 0; i < numWorkers; i++ {
		go hd.verificationWorker()
	}
}

// verificationWorker processes deep verification tasks
func (hd *HallucinationDetector) verificationWorker() {
	for task := range hd.asyncVerifyQueue {
		report := hd.deepVerification(task.thought)
		hd.cacheReport(report)

		select {
		case task.resultCh <- report:
		default:
		}
	}
}

// deepVerification performs comprehensive verification (can take seconds)
func (hd *HallucinationDetector) deepVerification(thought *types.Thought) *HallucinationReport {
	ctx := context.Background()

	report := &HallucinationReport{
		ThoughtID:         thought.ID,
		AnalyzedAt:        time.Now(),
		VerificationLevel: VerificationDeep,
		Claims:            []FactualClaim{},
	}

	// Extract claims
	claims := hd.extractClaims(thought.Content)

	// Verify each claim against knowledge sources
	for i := range claims {
		claim := &claims[i]
		wasVerified := false
		wasContradicted := false

		// Verify with each knowledge source
		for _, source := range hd.knowledgeSources {
			result, err := source.Verify(ctx, claim.Text)
			if err != nil {
				continue
			}

			// Update claim based on verification
			if result.IsVerified {
				claim.VerificationStatus = StatusVerified
				claim.SupportingEvidence = result.SupportingEvidence
				claim.SemanticUncertainty = 1.0 - result.Confidence
				wasVerified = true
			} else {
				if len(result.ContradictingEvidence) > 0 {
					claim.VerificationStatus = StatusFalse
					claim.ContradictingEvidence = result.ContradictingEvidence
					claim.SemanticUncertainty = 0.9
					wasContradicted = true
				}
			}

			claim.Sources = append(claim.Sources, VerificationSource{
				Type:       source.Type(),
				Source:     result.Source,
				Confidence: result.Confidence,
				Timestamp:  time.Now(),
			})
		}

		// Count each claim only once
		if wasVerified {
			report.VerifiedCount++
		}
		if wasContradicted {
			report.HallucinationCount++
		}

		// If no verification possible, mark as unverifiable
		if claim.VerificationStatus == StatusPending {
			claim.VerificationStatus = StatusUnverifiable
			claim.SemanticUncertainty = 0.7
		}
	}

	report.Claims = claims

	// Calculate overall risk
	if len(claims) > 0 {
		totalUncertainty := 0.0
		for _, claim := range claims {
			totalUncertainty += claim.SemanticUncertainty
		}
		report.OverallRisk = totalUncertainty / float64(len(claims))
	}

	// Generate recommendations
	if report.HallucinationCount > 0 {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("Found %d potential hallucinations. Verify before using.", report.HallucinationCount))
	}

	if report.OverallRisk > 0.6 {
		report.Recommendations = append(report.Recommendations,
			"High uncertainty detected. Request additional evidence or verification.")
	}

	return report
}

// getCachedReport retrieves a cached report
func (hd *HallucinationDetector) getCachedReport(thoughtID string) (*HallucinationReport, bool) {
	hd.mu.RLock()
	defer hd.mu.RUnlock()
	report, ok := hd.verificationCache[thoughtID]
	return report, ok
}

// cacheReport stores a report in the cache
func (hd *HallucinationDetector) cacheReport(report *HallucinationReport) {
	hd.mu.Lock()
	defer hd.mu.Unlock()
	hd.verificationCache[report.ThoughtID] = report
}

// GetReport retrieves the full report for a thought (waits for deep verification if needed)
func (hd *HallucinationDetector) GetReport(thoughtID string) (*HallucinationReport, error) {
	if report, ok := hd.getCachedReport(thoughtID); ok {
		return report, nil
	}
	return nil, fmt.Errorf("no report found for thought %s", thoughtID)
}
