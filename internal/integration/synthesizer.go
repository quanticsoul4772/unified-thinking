// Package integration provides cross-mode synthesis capabilities.
package integration

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// Synthesizer integrates insights from different reasoning modes
type Synthesizer struct {
	mu         sync.RWMutex
	syntheses  map[string]*types.Synthesis
	counter    int
}

// NewSynthesizer creates a new cross-mode synthesizer
func NewSynthesizer() *Synthesizer {
	return &Synthesizer{
		syntheses: make(map[string]*types.Synthesis),
	}
}

// Input represents a piece of reasoning from a specific mode
type Input struct {
	ID          string
	Mode        string // "causal", "temporal", "perspective", "probabilistic", etc.
	Content     string
	Confidence  float64
	Metadata    map[string]interface{}
}

// SynthesizeInsights combines insights from multiple reasoning modes
func (s *Synthesizer) SynthesizeInsights(inputs []*Input, context string) (*types.Synthesis, error) {
	if len(inputs) < 2 {
		return nil, fmt.Errorf("synthesis requires at least 2 inputs from different modes")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++

	// Extract source IDs
	sources := make([]string, len(inputs))
	for i, input := range inputs {
		sources[i] = input.ID
	}

	// Detect synergies (how inputs complement each other)
	synergies := s.detectSynergies(inputs, context)

	// Detect conflicts (contradictions or tensions)
	conflicts := s.detectConflicts(inputs, context)

	// Generate integrated view
	integratedView := s.generateIntegratedView(inputs, context, synergies, conflicts)

	// Calculate synthesis confidence
	confidence := s.calculateSynthesisConfidence(inputs, synergies, conflicts)

	synthesis := &types.Synthesis{
		ID:             fmt.Sprintf("synthesis-%d", s.counter),
		Sources:        sources,
		IntegratedView: integratedView,
		Synergies:      synergies,
		Conflicts:      conflicts,
		Confidence:     confidence,
		Metadata:       map[string]interface{}{},
		CreatedAt:      time.Now(),
	}

	// Store modes used
	modes := make([]string, 0, len(inputs))
	for _, input := range inputs {
		modes = append(modes, input.Mode)
	}
	synthesis.Metadata["modes"] = modes
	synthesis.Metadata["context"] = context

	s.syntheses[synthesis.ID] = synthesis
	return synthesis, nil
}

// detectSynergies identifies how different modes complement each other
func (s *Synthesizer) detectSynergies(inputs []*Input, context string) []string {
	synergies := make([]string, 0)
	modes := make(map[string]*Input)

	// Map inputs by mode
	for _, input := range inputs {
		modes[input.Mode] = input
	}

	// Check for known complementary patterns

	// Causal + Temporal synergy
	if causal, hasCausal := modes["causal"]; hasCausal {
		if temporal, hasTemporal := modes["temporal"]; hasTemporal {
			synergies = append(synergies, fmt.Sprintf(
				"Causal reasoning (%s) reveals mechanisms, while temporal analysis (%s) shows how effects unfold over time",
				extractKey(causal.Content, 30),
				extractKey(temporal.Content, 30),
			))
		}
	}

	// Causal + Probabilistic synergy
	if causal, hasCausal := modes["causal"]; hasCausal {
		if prob, hasProb := modes["probabilistic"]; hasProb {
			synergies = append(synergies, fmt.Sprintf(
				"Causal structure from %s provides qualitative relationships; probabilistic analysis from %s quantifies uncertainty",
				extractKey(causal.Content, 30),
				extractKey(prob.Content, 30),
			))
		}
	}

	// Perspective + Temporal synergy
	if perspective, hasPerspective := modes["perspective"]; hasPerspective {
		if temporal, hasTemporal := modes["temporal"]; hasTemporal {
			synergies = append(synergies, fmt.Sprintf(
				"Multiple stakeholder views (%s) combined with temporal analysis (%s) reveal how different groups prioritize short vs long term",
				extractKey(perspective.Content, 30),
				extractKey(temporal.Content, 30),
			))
		}
	}

	// Perspective + Causal synergy
	if perspective, hasPerspective := modes["perspective"]; hasPerspective {
		if causal, hasCausal := modes["causal"]; hasCausal {
			synergies = append(synergies, fmt.Sprintf(
				"Stakeholder perspectives (%s) identify what matters; causal analysis (%s) shows what can be changed",
				extractKey(perspective.Content, 30),
				extractKey(causal.Content, 30),
			))
		}
	}

	// Generic multi-mode synergy if no specific patterns found
	if len(synergies) == 0 {
		synergies = append(synergies, fmt.Sprintf(
			"Multiple reasoning modes (%s) provide complementary lenses on the situation",
			strings.Join(getModes(inputs), ", "),
		))
	}

	return synergies
}

// detectConflicts identifies contradictions or tensions between modes
func (s *Synthesizer) detectConflicts(inputs []*Input, context string) []string {
	conflicts := make([]string, 0)
	modes := make(map[string]*Input)

	// Map inputs by mode
	for _, input := range inputs {
		modes[input.Mode] = input
	}

	// Check for known conflicting patterns

	// Temporal conflicts (short-term vs long-term)
	if temporal, hasTemporal := modes["temporal"]; hasTemporal {
		if strings.Contains(strings.ToLower(temporal.Content), "tradeoff") ||
		   strings.Contains(strings.ToLower(temporal.Content), "tension") {
			conflicts = append(conflicts,
				"Temporal analysis reveals inherent short-term vs long-term tradeoffs that require prioritization")
		}
	}

	// Perspective conflicts
	if perspective, hasPerspective := modes["perspective"]; hasPerspective {
		if strings.Contains(strings.ToLower(perspective.Content), "conflict") ||
		   strings.Contains(strings.ToLower(perspective.Content), "disagreement") ||
		   strings.Contains(strings.ToLower(perspective.Content), "opposing") {
			conflicts = append(conflicts,
				"Different stakeholder perspectives have conflicting priorities or concerns that must be balanced")
		}
	}

	// Confidence conflicts
	confidences := make([]float64, 0, len(inputs))
	for _, input := range inputs {
		confidences = append(confidences, input.Confidence)
	}

	minConfidence, maxConfidence := findMinMax(confidences)
	if maxConfidence - minConfidence > 0.4 {
		conflicts = append(conflicts, fmt.Sprintf(
			"Significant confidence variation across modes (%.2f to %.2f) suggests underlying uncertainty",
			minConfidence, maxConfidence,
		))
	}

	// Low confidence consensus
	avgConfidence := average(confidences)
	if avgConfidence < 0.5 {
		conflicts = append(conflicts,
			"Low average confidence across modes indicates high uncertainty requiring more investigation")
	}

	return conflicts
}

// generateIntegratedView creates a unified conclusion from multiple inputs
func (s *Synthesizer) generateIntegratedView(inputs []*Input, context string, synergies, conflicts []string) string {
	var builder strings.Builder

	// Start with context if provided
	if context != "" {
		builder.WriteString(fmt.Sprintf("Integrated analysis of: %s\n\n", context))
	}

	// Synthesize key insights from each mode
	builder.WriteString("Key Insights:\n")
	for i, input := range inputs {
		builder.WriteString(fmt.Sprintf("%d. [%s mode] %s (confidence: %.2f)\n",
			i+1, input.Mode, extractKey(input.Content, 120), input.Confidence))
	}

	// Describe synergies
	if len(synergies) > 0 {
		builder.WriteString("\nComplementary Insights:\n")
		for i, synergy := range synergies {
			builder.WriteString(fmt.Sprintf("- %s\n", synergy))
			if i >= 2 { // Limit synergies to avoid verbosity
				break
			}
		}
	}

	// Address conflicts
	if len(conflicts) > 0 {
		builder.WriteString("\nTensions to Address:\n")
		for i, conflict := range conflicts {
			builder.WriteString(fmt.Sprintf("- %s\n", conflict))
			if i >= 2 { // Limit conflicts to avoid verbosity
				break
			}
		}
	}

	// Generate synthesis
	builder.WriteString("\nSynthesized Conclusion:\n")
	builder.WriteString(s.generateConclusion(inputs, synergies, conflicts))

	return builder.String()
}

// generateConclusion creates the final synthesized conclusion
func (s *Synthesizer) generateConclusion(inputs []*Input, synergies, conflicts []string) string {
	modes := getModes(inputs)
	avgConfidence := 0.0
	for _, input := range inputs {
		avgConfidence += input.Confidence
	}
	avgConfidence /= float64(len(inputs))

	var conclusion strings.Builder

	// Lead with overall assessment
	if avgConfidence > 0.7 {
		conclusion.WriteString("High confidence synthesis: ")
	} else if avgConfidence > 0.5 {
		conclusion.WriteString("Moderate confidence synthesis: ")
	} else {
		conclusion.WriteString("Preliminary synthesis (low confidence): ")
	}

	// Describe what the synthesis reveals
	if len(conflicts) > len(synergies) {
		conclusion.WriteString(fmt.Sprintf(
			"Analysis across %d modes (%s) reveals significant tensions requiring careful navigation. ",
			len(modes), strings.Join(modes, ", "),
		))
	} else {
		conclusion.WriteString(fmt.Sprintf(
			"Analysis across %d modes (%s) provides complementary insights that strengthen understanding. ",
			len(modes), strings.Join(modes, ", "),
		))
	}

	// Add actionability
	if len(conflicts) > 0 {
		conclusion.WriteString("Recommend addressing identified conflicts before proceeding. ")
	}

	if avgConfidence < 0.6 {
		conclusion.WriteString("Further investigation recommended to increase confidence. ")
	}

	conclusion.WriteString("This multi-modal analysis provides a more complete picture than any single reasoning mode.")

	return conclusion.String()
}

// calculateSynthesisConfidence determines overall confidence of the synthesis
func (s *Synthesizer) calculateSynthesisConfidence(inputs []*Input, synergies, conflicts []string) float64 {
	// Start with average input confidence
	sum := 0.0
	for _, input := range inputs {
		sum += input.Confidence
	}
	avgConfidence := sum / float64(len(inputs))

	// Boost for synergies (complementary insights increase confidence)
	synergyBoost := float64(len(synergies)) * 0.05
	if synergyBoost > 0.15 {
		synergyBoost = 0.15 // Cap at +15%
	}

	// Penalty for conflicts (contradictions reduce confidence)
	conflictPenalty := float64(len(conflicts)) * 0.08
	if conflictPenalty > 0.25 {
		conflictPenalty = 0.25 // Cap at -25%
	}

	// Bonus for diversity of modes
	modeBonus := 0.0
	if len(inputs) >= 3 {
		modeBonus = 0.05 // +5% for 3+ modes
	}
	if len(inputs) >= 4 {
		modeBonus = 0.10 // +10% for 4+ modes
	}

	confidence := avgConfidence + synergyBoost + modeBonus - conflictPenalty

	// Clamp to [0, 1]
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// DetectEmergentPatterns identifies patterns that only become visible across modes
func (s *Synthesizer) DetectEmergentPatterns(inputs []*Input) ([]string, error) {
	if len(inputs) < 2 {
		return nil, fmt.Errorf("pattern detection requires at least 2 inputs")
	}

	patterns := make([]string, 0)

	// Pattern: Feedback loops (causal + temporal)
	hasCausal := false
	hasTemporal := false
	for _, input := range inputs {
		if input.Mode == "causal" && (strings.Contains(strings.ToLower(input.Content), "cycle") ||
		   strings.Contains(strings.ToLower(input.Content), "feedback") ||
		   strings.Contains(strings.ToLower(input.Content), "reinforcing")) {
			hasCausal = true
		}
		if input.Mode == "temporal" && (strings.Contains(strings.ToLower(input.Content), "compound") ||
		   strings.Contains(strings.ToLower(input.Content), "accumulate")) {
			hasTemporal = true
		}
	}
	if hasCausal && hasTemporal {
		patterns = append(patterns, "Feedback loop: Causal relationships create self-reinforcing patterns that compound over time")
	}

	// Pattern: Misaligned incentives (perspective + causal/temporal)
	hasPerspective := false
	hasIncentive := false
	for _, input := range inputs {
		if input.Mode == "perspective" {
			hasPerspective = true
		}
		if strings.Contains(strings.ToLower(input.Content), "incentive") ||
		   strings.Contains(strings.ToLower(input.Content), "motivation") {
			hasIncentive = true
		}
	}
	if hasPerspective && hasIncentive {
		patterns = append(patterns, "Incentive misalignment: Different stakeholders have conflicting motivations driving behavior")
	}

	// Pattern: Delayed consequences (causal + temporal)
	hasDelay := false
	for _, input := range inputs {
		contentLower := strings.ToLower(input.Content)
		if strings.Contains(contentLower, "delay") ||
		   strings.Contains(contentLower, "lag") ||
		   strings.Contains(contentLower, "long-term") {
			hasDelay = true
			// Mark as having both causal and temporal if both keywords present
			if input.Mode == "causal" || strings.Contains(contentLower, "cause") {
				hasCausal = true
			}
			if input.Mode == "temporal" || strings.Contains(contentLower, "time") {
				hasTemporal = true
			}
			break
		}
	}
	if hasDelay {
		patterns = append(patterns, "Delayed impact: Short-term actions have significant long-term causal consequences")
	}

	// Pattern: Uncertainty cascade (probabilistic + causal)
	hasProb := false
	hasUncertainty := false
	for _, input := range inputs {
		if input.Mode == "probabilistic" || strings.Contains(strings.ToLower(input.Content), "probability") {
			hasProb = true
		}
		if input.Confidence < 0.6 {
			hasUncertainty = true
		}
		if input.Mode == "causal" || strings.Contains(strings.ToLower(input.Content), "causal") {
			hasCausal = true
		}
	}
	if hasProb && hasUncertainty {
		patterns = append(patterns, "Uncertainty propagation: Initial uncertainties amplify through causal chains")
	}

	// Generic pattern if none detected
	if len(patterns) == 0 {
		patterns = append(patterns, "Cross-mode analysis reveals interconnected factors requiring holistic consideration")
	}

	return patterns, nil
}

// GetSynthesis retrieves a synthesis by ID
func (s *Synthesizer) GetSynthesis(id string) (*types.Synthesis, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	synthesis, ok := s.syntheses[id]
	if !ok {
		return nil, fmt.Errorf("synthesis not found: %s", id)
	}

	return synthesis, nil
}

// Helper functions

func extractKey(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLen {
		return text
	}
	// Try to break at sentence
	if idx := strings.Index(text, ". "); idx > 0 && idx < maxLen {
		return text[:idx]
	}
	// Otherwise just truncate
	return text[:maxLen] + "..."
}

func getModes(inputs []*Input) []string {
	modes := make([]string, 0, len(inputs))
	seen := make(map[string]bool)
	for _, input := range inputs {
		if !seen[input.Mode] {
			modes = append(modes, input.Mode)
			seen[input.Mode] = true
		}
	}
	return modes
}

func findMinMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	min := values[0]
	max := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
