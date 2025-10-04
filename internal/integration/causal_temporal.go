// Package integration provides causal-temporal integration.
//
// This module combines causal reasoning with temporal analysis to show how
// causal effects evolve over different time horizons.
package integration

import (
	"fmt"
	"time"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

// CausalTemporalIntegration combines causal and temporal reasoning
type CausalTemporalIntegration struct {
	causalReasoner   *reasoning.CausalReasoner
	temporalReasoner *reasoning.TemporalReasoner
}

// NewCausalTemporalIntegration creates a new causal-temporal integrator
func NewCausalTemporalIntegration(
	causalReasoner *reasoning.CausalReasoner,
	temporalReasoner *reasoning.TemporalReasoner,
) *CausalTemporalIntegration {
	return &CausalTemporalIntegration{
		causalReasoner:   causalReasoner,
		temporalReasoner: temporalReasoner,
	}
}

// TemporalCausalEffect represents how a causal effect evolves over time
type TemporalCausalEffect struct {
	ID             string                    `json:"id"`
	GraphID        string                    `json:"graph_id"`
	Variable       string                    `json:"variable"`
	TimeHorizons   map[string]*HorizonEffect `json:"time_horizons"` // "short_term", "medium_term", "long_term"
	OverallPattern string                    `json:"overall_pattern"` // "increasing", "decreasing", "stable", "oscillating"
	PeakEffect     string                    `json:"peak_effect"`     // When effect is strongest
	Recommendation string                    `json:"recommendation"`
	CreatedAt      time.Time                 `json:"created_at"`
}

// HorizonEffect describes effect at a specific time horizon
type HorizonEffect struct {
	TimeFrame       string                `json:"time_frame"`       // "days-weeks", "months", "years"
	DirectEffects   []*types.PredictedEffect `json:"direct_effects"`
	IndirectEffects []*types.PredictedEffect `json:"indirect_effects"`
	Magnitude       string                `json:"magnitude"`        // "weak", "moderate", "strong"
	Certainty       float64               `json:"certainty"`        // 0.0-1.0
}

// AnalyzeTemporalCausalEffects analyzes how intervention effects evolve over time
func (cti *CausalTemporalIntegration) AnalyzeTemporalCausalEffects(
	graphID string,
	variableID string,
	interventionType string,
) (*TemporalCausalEffect, error) {
	// Step 1: Simulate intervention
	intervention, err := cti.causalReasoner.SimulateIntervention(graphID, variableID, interventionType)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate intervention: %w", err)
	}

	// Step 2: Analyze effects across time horizons
	timeHorizons := map[string]*HorizonEffect{
		"short_term":  cti.analyzeShortTermEffects(intervention),
		"medium_term": cti.analyzeMediumTermEffects(intervention),
		"long_term":   cti.analyzeLongTermEffects(intervention),
	}

	// Step 3: Determine overall pattern
	pattern := cti.determineTemporalPattern(timeHorizons)

	// Step 4: Identify peak effect timing
	peakEffect := cti.identifyPeakEffectTime(timeHorizons)

	// Step 5: Generate recommendation
	recommendation := cti.generateTemporalRecommendation(pattern, peakEffect, timeHorizons)

	result := &TemporalCausalEffect{
		ID:             fmt.Sprintf("tce_%d", time.Now().UnixNano()),
		GraphID:        graphID,
		Variable:       variableID,
		TimeHorizons:   timeHorizons,
		OverallPattern: pattern,
		PeakEffect:     peakEffect,
		Recommendation: recommendation,
		CreatedAt:      time.Now(),
	}

	return result, nil
}

// AnalyzeDecisionTiming determines optimal timing for intervention
func (cti *CausalTemporalIntegration) AnalyzeDecisionTiming(
	situation string,
	causalGraphID string,
) (map[string]interface{}, error) {
	// Get causal graph
	graph, err := cti.causalReasoner.GetGraph(causalGraphID)
	if err != nil {
		return nil, err
	}

	// Analyze temporal aspects
	temporalAnalysis, err := cti.temporalReasoner.AnalyzeTemporal(situation, "")
	if err != nil {
		return nil, err
	}

	// Identify time-sensitive variables in causal graph
	timeSensitive := cti.identifyTimeSensitiveVariables(graph)

	// Determine optimal timing windows
	timingWindows := cti.determineTimingWindows(graph, temporalAnalysis, timeSensitive)

	result := map[string]interface{}{
		"graph_id":           causalGraphID,
		"temporal_analysis":  temporalAnalysis,
		"time_sensitive_vars": timeSensitive,
		"timing_windows":     timingWindows,
		"recommendation":     cti.synthesizeTimingRecommendation(timingWindows),
	}

	return result, nil
}

// Private helper methods

func (cti *CausalTemporalIntegration) analyzeShortTermEffects(intervention *types.CausalIntervention) *HorizonEffect {
	// Short term: immediate and 1-hop effects
	directEffects := []*types.PredictedEffect{}
	indirectEffects := []*types.PredictedEffect{}

	for _, effect := range intervention.PredictedEffects {
		if effect.PathLength == 1 {
			directEffects = append(directEffects, effect)
		} else if effect.PathLength == 2 {
			indirectEffects = append(indirectEffects, effect)
		}
	}

	magnitude := cti.calculateMagnitude(directEffects)
	certainty := cti.calculateCertainty(directEffects)

	return &HorizonEffect{
		TimeFrame:       "days-weeks",
		DirectEffects:   directEffects,
		IndirectEffects: indirectEffects,
		Magnitude:       magnitude,
		Certainty:       certainty,
	}
}

func (cti *CausalTemporalIntegration) analyzeMediumTermEffects(intervention *types.CausalIntervention) *HorizonEffect {
	// Medium term: 2-3 hop effects
	directEffects := []*types.PredictedEffect{}
	indirectEffects := []*types.PredictedEffect{}

	for _, effect := range intervention.PredictedEffects {
		if effect.PathLength == 2 {
			directEffects = append(directEffects, effect)
		} else if effect.PathLength == 3 {
			indirectEffects = append(indirectEffects, effect)
		}
	}

	magnitude := cti.calculateMagnitude(directEffects)
	certainty := cti.calculateCertainty(directEffects) * 0.8 // Reduce certainty for medium term

	return &HorizonEffect{
		TimeFrame:       "months",
		DirectEffects:   directEffects,
		IndirectEffects: indirectEffects,
		Magnitude:       magnitude,
		Certainty:       certainty,
	}
}

func (cti *CausalTemporalIntegration) analyzeLongTermEffects(intervention *types.CausalIntervention) *HorizonEffect {
	// Long term: 3+ hop effects, secondary effects
	directEffects := []*types.PredictedEffect{}
	indirectEffects := []*types.PredictedEffect{}

	for _, effect := range intervention.PredictedEffects {
		if effect.PathLength >= 3 {
			directEffects = append(directEffects, effect)
		} else if effect.PathLength >= 4 {
			indirectEffects = append(indirectEffects, effect)
		}
	}

	magnitude := cti.calculateMagnitude(directEffects)
	certainty := cti.calculateCertainty(directEffects) * 0.6 // Further reduce certainty for long term

	return &HorizonEffect{
		TimeFrame:       "years",
		DirectEffects:   directEffects,
		IndirectEffects: indirectEffects,
		Magnitude:       magnitude,
		Certainty:       certainty,
	}
}

func (cti *CausalTemporalIntegration) determineTemporalPattern(horizons map[string]*HorizonEffect) string {
	short := horizons["short_term"]
	medium := horizons["medium_term"]
	long := horizons["long_term"]

	shortMag := cti.magnitudeValue(short.Magnitude)
	mediumMag := cti.magnitudeValue(medium.Magnitude)
	longMag := cti.magnitudeValue(long.Magnitude)

	if shortMag < mediumMag && mediumMag < longMag {
		return "increasing" // Effects strengthen over time
	} else if shortMag > mediumMag && mediumMag > longMag {
		return "decreasing" // Effects weaken over time
	} else if shortMag > mediumMag && longMag > mediumMag {
		return "oscillating" // Effects vary over time
	} else {
		return "stable" // Consistent effects
	}
}

func (cti *CausalTemporalIntegration) identifyPeakEffectTime(horizons map[string]*HorizonEffect) string {
	maxMagnitude := 0.0
	peakTime := "short_term"

	for name, horizon := range horizons {
		mag := cti.magnitudeValue(horizon.Magnitude) * horizon.Certainty
		if mag > maxMagnitude {
			maxMagnitude = mag
			peakTime = name
		}
	}

	switch peakTime {
	case "short_term":
		return "Effects peak in days to weeks"
	case "medium_term":
		return "Effects peak in months"
	case "long_term":
		return "Effects peak after years"
	default:
		return "Peak effect time uncertain"
	}
}

func (cti *CausalTemporalIntegration) generateTemporalRecommendation(pattern, peakEffect string, horizons map[string]*HorizonEffect) string {
	recommendations := []string{}

	// Based on pattern
	switch pattern {
	case "increasing":
		recommendations = append(recommendations, "Effects strengthen over time - patience is beneficial")
	case "decreasing":
		recommendations = append(recommendations, "Initial effects are strongest - act quickly to capitalize")
	case "oscillating":
		recommendations = append(recommendations, "Effects vary - monitor closely and adapt strategy")
	case "stable":
		recommendations = append(recommendations, "Effects are consistent across timeframes")
	}

	// Based on peak effect
	recommendations = append(recommendations, peakEffect)

	// Based on certainty
	shortCertainty := horizons["short_term"].Certainty
	longCertainty := horizons["long_term"].Certainty

	if shortCertainty > 0.7 && longCertainty < 0.5 {
		recommendations = append(recommendations, "Short-term effects are predictable, but long-term outcomes are uncertain")
	} else if longCertainty > 0.7 {
		recommendations = append(recommendations, "Long-term effects are well-understood - suitable for strategic planning")
	}

	return joinStrings(recommendations, ". ")
}

func (cti *CausalTemporalIntegration) identifyTimeSensitiveVariables(graph *types.CausalGraph) []string {
	timeSensitive := []string{}

	// Variables with many outgoing links are time-sensitive (affect many things)
	outgoingCount := make(map[string]int)
	for _, link := range graph.Links {
		outgoingCount[link.From]++
	}

	for varID, count := range outgoingCount {
		if count >= 3 { // High influence
			for _, v := range graph.Variables {
				if v.ID == varID {
					timeSensitive = append(timeSensitive, v.Name)
					break
				}
			}
		}
	}

	return timeSensitive
}

func (cti *CausalTemporalIntegration) determineTimingWindows(
	graph *types.CausalGraph,
	temporal *types.TemporalAnalysis,
	timeSensitive []string,
) []map[string]string {
	windows := []map[string]string{}

	// Early window: capitalize on short-term effects
	windows = append(windows, map[string]string{
		"name":        "Early Action Window",
		"timeframe":   "Immediate to 1 month",
		"focus":       "Capture short-term benefits before they dissipate",
		"variables":   joinStrings(timeSensitive, ", "),
		"priority":    "High if effects are decreasing",
	})

	// Strategic window: leverage medium-term effects
	windows = append(windows, map[string]string{
		"name":        "Strategic Window",
		"timeframe":   "1-6 months",
		"focus":       "Position for medium-term outcomes",
		"priority":    "High if effects are stable or increasing",
	})

	// Long-term window: sustainable change
	windows = append(windows, map[string]string{
		"name":        "Long-term Commitment Window",
		"timeframe":   "6+ months",
		"focus":       "Establish sustainable patterns",
		"priority":    "Essential if long-term effects dominate",
	})

	return windows
}

func (cti *CausalTemporalIntegration) synthesizeTimingRecommendation(windows []map[string]string) string {
	if len(windows) == 0 {
		return "Unable to determine optimal timing"
	}

	// Default recommendation based on first window
	return fmt.Sprintf("Consider acting within the %s to %s",
		windows[0]["name"],
		windows[0]["focus"])
}

// Helper functions

func (cti *CausalTemporalIntegration) calculateMagnitude(effects []*types.PredictedEffect) string {
	if len(effects) == 0 {
		return "weak"
	}

	// Count strong effects
	strongCount := 0
	for _, effect := range effects {
		if effect.Magnitude > 0.7 || effect.Probability > 0.8 {
			strongCount++
		}
	}

	ratio := float64(strongCount) / float64(len(effects))
	if ratio > 0.6 {
		return "strong"
	} else if ratio > 0.3 {
		return "moderate"
	}
	return "weak"
}

func (cti *CausalTemporalIntegration) calculateCertainty(effects []*types.PredictedEffect) float64 {
	if len(effects) == 0 {
		return 0.0
	}

	total := 0.0
	for _, effect := range effects {
		total += effect.Probability
	}
	return total / float64(len(effects))
}

func (cti *CausalTemporalIntegration) magnitudeValue(magnitude string) float64 {
	switch magnitude {
	case "strong":
		return 1.0
	case "moderate":
		return 0.6
	case "weak":
		return 0.3
	default:
		return 0.0
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
