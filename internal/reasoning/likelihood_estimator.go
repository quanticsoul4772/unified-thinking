package reasoning

import (
	"fmt"

	"unified-thinking/internal/types"
)

// LikelihoodEstimator converts evidence quality scores into conditional probabilities
// P(E|H) and P(E|¬H) for Bayesian belief updates.
type LikelihoodEstimator interface {
	// EstimateLikelihoods returns P(E|H) and P(E|¬H) based on evidence characteristics
	EstimateLikelihoods(evidence *types.Evidence) (ifTrue, ifFalse float64, err error)
}

// EvidenceProfile configures likelihood estimation parameters for a specific domain.
// Different domains may require different calibration of how evidence quality maps
// to conditional probabilities.
type EvidenceProfile struct {
	Name string // Domain name (e.g., "scientific", "anecdotal", "expert-opinion")

	// Supporting evidence parameters (when evidence.SupportsClaim = true)
	SupportHigh float64 // Added to 0.5 base for P(E|H) when evidence supports claim
	SupportLow  float64 // Subtracted from 0.5 base for P(E|¬H) when evidence supports claim

	// Refuting evidence parameters (when evidence.SupportsClaim = false)
	RefuteHigh float64 // Subtracted from 0.5 base for P(E|H) when evidence refutes claim
	RefuteLow  float64 // Added to 0.5 base for P(E|¬H) when evidence refutes claim
}

// DefaultProfile returns the standard evidence profile used for general-purpose reasoning.
//
// Calibration rationale:
// - SupportHigh = 0.4: High-quality supporting evidence → P(E|H) ranges from 0.5 to 0.9
// - SupportLow = 0.3: High-quality supporting evidence → P(E|¬H) ranges from 0.5 to 0.2
// - RefuteHigh = 0.4: High-quality refuting evidence → P(E|H) ranges from 0.5 to 0.1
// - RefuteLow = 0.3: High-quality refuting evidence → P(E|¬H) ranges from 0.5 to 0.8
//
// These values ensure:
// 1. Evidence quality of 1.0 produces strong likelihood ratios (LR = P(E|H)/P(E|¬H))
// 2. Evidence quality of 0.0 produces neutral likelihoods (both 0.5)
// 3. Likelihoods stay within valid [0,1] range
func DefaultProfile() *EvidenceProfile {
	return &EvidenceProfile{
		Name:        "default",
		SupportHigh: 0.4,
		SupportLow:  0.3,
		RefuteHigh:  0.4,
		RefuteLow:   0.3,
	}
}

// ScientificProfile returns a profile calibrated for scientific evidence.
// Scientific evidence has higher discriminatory power due to rigorous methodology.
func ScientificProfile() *EvidenceProfile {
	return &EvidenceProfile{
		Name:        "scientific",
		SupportHigh: 0.45, // Higher range: 0.5 to 0.95
		SupportLow:  0.35, // Lower range: 0.5 to 0.15
		RefuteHigh:  0.45,
		RefuteLow:   0.35,
	}
}

// AnecdotalProfile returns a profile for anecdotal evidence.
// Anecdotal evidence has lower discriminatory power and should be weighted cautiously.
func AnecdotalProfile() *EvidenceProfile {
	return &EvidenceProfile{
		Name:        "anecdotal",
		SupportHigh: 0.25, // Narrower range: 0.5 to 0.75
		SupportLow:  0.20, // Narrower range: 0.5 to 0.30
		RefuteHigh:  0.25,
		RefuteLow:   0.20,
	}
}

// StandardEstimator implements LikelihoodEstimator with configurable evidence profiles.
type StandardEstimator struct {
	profile *EvidenceProfile
}

// NewStandardEstimator creates a new likelihood estimator with the given profile.
// If profile is nil, uses DefaultProfile().
func NewStandardEstimator(profile *EvidenceProfile) *StandardEstimator {
	if profile == nil {
		profile = DefaultProfile()
	}
	return &StandardEstimator{profile: profile}
}

// EstimateLikelihoods converts evidence quality score to conditional probabilities.
//
// The estimation works as follows:
//
// For supporting evidence (SupportsClaim = true):
//   P(E|H) = 0.5 + (OverallScore × SupportHigh)
//   P(E|¬H) = 0.5 - (OverallScore × SupportLow)
//
// For refuting evidence (SupportsClaim = false):
//   P(E|H) = 0.5 - (OverallScore × RefuteHigh)
//   P(E|¬H) = 0.5 + (OverallScore × RefuteLow)
//
// This ensures:
// - High-quality evidence has strong likelihood ratios (far from 1.0)
// - Low-quality evidence has weak likelihood ratios (close to 1.0)
// - All likelihoods stay within [0, 1] range
func (e *StandardEstimator) EstimateLikelihoods(evidence *types.Evidence) (float64, float64, error) {
	// Validate evidence score
	if evidence.OverallScore < 0 || evidence.OverallScore > 1 {
		return 0, 0, fmt.Errorf("evidence score must be in [0,1], got: %f", evidence.OverallScore)
	}

	var ifTrue, ifFalse float64

	if evidence.SupportsClaim {
		// Evidence supports the hypothesis
		ifTrue = 0.5 + (evidence.OverallScore * e.profile.SupportHigh)
		ifFalse = 0.5 - (evidence.OverallScore * e.profile.SupportLow)
	} else {
		// Evidence refutes the hypothesis
		ifTrue = 0.5 - (evidence.OverallScore * e.profile.RefuteHigh)
		ifFalse = 0.5 + (evidence.OverallScore * e.profile.RefuteLow)
	}

	// Sanity check: ensure likelihoods are in valid range
	if ifTrue < 0 || ifTrue > 1 || ifFalse < 0 || ifFalse > 1 {
		return 0, 0, fmt.Errorf("estimated likelihoods out of range: P(E|H)=%f, P(E|¬H)=%f (profile: %s)",
			ifTrue, ifFalse, e.profile.Name)
	}

	return ifTrue, ifFalse, nil
}

// GetProfile returns the current evidence profile
func (e *StandardEstimator) GetProfile() *EvidenceProfile {
	return e.profile
}

// SetProfile updates the evidence profile (useful for domain-specific calibration)
func (e *StandardEstimator) SetProfile(profile *EvidenceProfile) {
	if profile != nil {
		e.profile = profile
	}
}
