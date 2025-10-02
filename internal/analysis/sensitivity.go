package analysis

import (
	"fmt"
	"math"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// SensitivityAnalyzer performs robustness testing of conclusions
type SensitivityAnalyzer struct {
	mu      sync.RWMutex
	counter int
}

// NewSensitivityAnalyzer creates a new sensitivity analyzer
func NewSensitivityAnalyzer() *SensitivityAnalyzer {
	return &SensitivityAnalyzer{}
}

// AnalyzeSensitivity tests how robust a claim is to changes in assumptions
func (sa *SensitivityAnalyzer) AnalyzeSensitivity(targetClaim string, assumptions []string, baseConfidence float64) (*types.SensitivityAnalysis, error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	sa.counter++

	variations := make([]*types.Variation, 0)

	// Test each assumption
	for i, assumption := range assumptions {
		// Create variation where assumption is weakened
		variation := sa.createVariation(
			fmt.Sprintf("variation-%d-%d", sa.counter, i+1),
			fmt.Sprintf("Weaken assumption: %s", assumption),
			baseConfidence,
		)
		variations = append(variations, variation)

		// Create variation where assumption is strengthened
		variation2 := sa.createVariation(
			fmt.Sprintf("variation-%d-%d-alt", sa.counter, i+1),
			fmt.Sprintf("Strengthen assumption: %s", assumption),
			baseConfidence,
		)
		variations = append(variations, variation2)
	}

	// Calculate overall robustness
	robustness := sa.calculateRobustness(variations)

	// Identify key assumptions (those with highest impact)
	keyAssumptions := sa.identifyKeyAssumptions(assumptions, variations)

	analysis := &types.SensitivityAnalysis{
		ID:             fmt.Sprintf("sensitivity-%d", sa.counter),
		TargetClaim:    targetClaim,
		Variations:     variations,
		Robustness:     robustness,
		KeyAssumptions: keyAssumptions,
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
	}

	return analysis, nil
}

// createVariation generates a variation scenario
func (sa *SensitivityAnalyzer) createVariation(id, assumptionChange string, baseConfidence float64) *types.Variation {
	// Simulate impact - in real system, would need actual reasoning
	// For now, estimate impact based on assumption change type
	impactMagnitude := 0.0
	impact := ""

	if len(assumptionChange) > 50 {
		// Complex assumption change likely has higher impact
		impactMagnitude = 0.3 + (float64(len(assumptionChange)%30) / 100.0)
		impact = "Moderate to high impact on conclusion confidence"
	} else {
		impactMagnitude = 0.1 + (float64(len(assumptionChange)%20) / 100.0)
		impact = "Low to moderate impact on conclusion confidence"
	}

	// Cap impact magnitude
	if impactMagnitude > 0.8 {
		impactMagnitude = 0.8
	}

	return &types.Variation{
		ID:              id,
		AssumptionChange: assumptionChange,
		Impact:          impact,
		ImpactMagnitude: impactMagnitude,
	}
}

// calculateRobustness computes overall robustness score
func (sa *SensitivityAnalyzer) calculateRobustness(variations []*types.Variation) float64 {
	if len(variations) == 0 {
		return 1.0 // No variations tested = maximally robust (trivially)
	}

	// Robustness is inverse of average impact magnitude
	totalImpact := 0.0
	for _, v := range variations {
		totalImpact += v.ImpactMagnitude
	}
	avgImpact := totalImpact / float64(len(variations))

	// Robustness = 1 - average impact
	robustness := 1.0 - avgImpact
	return math.Max(0, math.Min(1, robustness))
}

// identifyKeyAssumptions finds assumptions with highest impact
func (sa *SensitivityAnalyzer) identifyKeyAssumptions(assumptions []string, variations []*types.Variation) []string {
	// Build map of assumption to maximum impact
	assumptionImpacts := make(map[string]float64)

	for i, assumption := range assumptions {
		maxImpact := 0.0
		// Find variations related to this assumption
		for j, v := range variations {
			// Variations are created in pairs for each assumption
			if j/2 == i {
				if v.ImpactMagnitude > maxImpact {
					maxImpact = v.ImpactMagnitude
				}
			}
		}
		assumptionImpacts[assumption] = maxImpact
	}

	// Select assumptions with impact > 0.3 as "key"
	keyAssumptions := make([]string, 0)
	for assumption, impact := range assumptionImpacts {
		if impact > 0.3 {
			keyAssumptions = append(keyAssumptions, assumption)
		}
	}

	// If none are key, return top 2
	if len(keyAssumptions) == 0 && len(assumptions) > 0 {
		// Simple heuristic: first 2 assumptions
		limit := 2
		if len(assumptions) < 2 {
			limit = len(assumptions)
		}
		return assumptions[:limit]
	}

	return keyAssumptions
}
