// Package reasoning provides analogical reasoning capabilities.
//
// This module enables cross-domain mapping and similarity-based inference,
// allowing the system to learn from past situations and transfer solutions
// across different problem domains.
package reasoning

import (
	"fmt"
	"strings"
	"time"

	"unified-thinking/internal/types"
)

// AnalogicalReasoner performs cross-domain analogical reasoning
type AnalogicalReasoner struct {
	analogies map[string]*types.Analogy
}

// NewAnalogicalReasoner creates a new analogical reasoner
func NewAnalogicalReasoner() *AnalogicalReasoner {
	return &AnalogicalReasoner{
		analogies: make(map[string]*types.Analogy),
	}
}

// FindAnalogy identifies analogies between source and target domains
func (ar *AnalogicalReasoner) FindAnalogy(sourceDomain, targetProblem string, constraints []string) (*types.Analogy, error) {
	if sourceDomain == "" || targetProblem == "" {
		return nil, fmt.Errorf("source domain and target problem are required")
	}

	// Extract key concepts from source and target
	sourceConcepts := ar.extractConcepts(sourceDomain)
	targetConcepts := ar.extractConcepts(targetProblem)

	// Build mapping between concepts
	mapping := ar.buildMapping(sourceConcepts, targetConcepts, constraints)

	// Calculate analogy strength based on mapping quality
	strength := ar.calculateAnalogyStrength(mapping, sourceConcepts, targetConcepts)

	// Generate insight from analogy
	insight := ar.generateInsight(sourceDomain, targetProblem, mapping)

	analogy := &types.Analogy{
		ID:           fmt.Sprintf("analogy_%d", time.Now().UnixNano()),
		SourceDomain: sourceDomain,
		TargetDomain: targetProblem,
		Mapping:      mapping,
		Insight:      insight,
		Strength:     strength,
		Metadata: map[string]interface{}{
			"constraints":      constraints,
			"source_concepts":  sourceConcepts,
			"target_concepts":  targetConcepts,
			"mapping_coverage": float64(len(mapping)) / float64(len(sourceConcepts)),
		},
		CreatedAt: time.Now(),
	}

	ar.analogies[analogy.ID] = analogy
	return analogy, nil
}

// ApplyAnalogy applies an existing analogy to a new context
func (ar *AnalogicalReasoner) ApplyAnalogy(analogyID, targetContext string) (map[string]interface{}, error) {
	analogy, exists := ar.analogies[analogyID]
	if !exists {
		return nil, fmt.Errorf("analogy %s not found", analogyID)
	}

	// Extract concepts from new target context
	targetConcepts := ar.extractConcepts(targetContext)

	// Apply mapping to new context
	result := make(map[string]interface{})
	result["analogy_id"] = analogyID
	result["source_domain"] = analogy.SourceDomain
	result["target_context"] = targetContext
	result["strength"] = analogy.Strength

	// Transfer solutions/insights based on mapping
	transferredInsights := ar.transferInsights(analogy, targetConcepts)
	result["transferred_insights"] = transferredInsights

	// Identify potential adaptation needs
	adaptations := ar.identifyAdaptations(analogy, targetConcepts)
	result["recommended_adaptations"] = adaptations

	return result, nil
}

// GetAnalogy retrieves an analogy by ID
func (ar *AnalogicalReasoner) GetAnalogy(id string) (*types.Analogy, error) {
	analogy, exists := ar.analogies[id]
	if !exists {
		return nil, fmt.Errorf("analogy %s not found", id)
	}
	return analogy, nil
}

// ListAnalogies returns all stored analogies
func (ar *AnalogicalReasoner) ListAnalogies() []*types.Analogy {
	analogies := make([]*types.Analogy, 0, len(ar.analogies))
	for _, a := range ar.analogies {
		analogies = append(analogies, a)
	}
	return analogies
}

// extractConcepts extracts key concepts from a domain description
func (ar *AnalogicalReasoner) extractConcepts(text string) []string {
	// Simple concept extraction based on keywords and noun phrases
	concepts := []string{}
	words := strings.Fields(strings.ToLower(text))

	// Common domain concepts that often map well
	keywordPatterns := []string{
		"flow", "structure", "system", "process", "component",
		"relationship", "hierarchy", "network", "pattern", "cycle",
		"input", "output", "transformation", "feedback", "control",
		"resource", "constraint", "goal", "obstacle", "solution",
	}

	for _, word := range words {
		for _, pattern := range keywordPatterns {
			if strings.Contains(word, pattern) && !ar.contains(concepts, word) {
				concepts = append(concepts, word)
			}
		}
	}

	// Also extract quoted phrases as concepts
	parts := strings.Split(text, "\"")
	for i := 1; i < len(parts); i += 2 {
		concept := strings.TrimSpace(parts[i])
		if concept != "" && !ar.contains(concepts, concept) {
			concepts = append(concepts, concept)
		}
	}

	return concepts
}

// buildMapping creates concept mappings between source and target
func (ar *AnalogicalReasoner) buildMapping(source, target []string, constraints []string) map[string]string {
	mapping := make(map[string]string)

	// Direct matching (case-insensitive)
	for _, src := range source {
		for _, tgt := range target {
			if strings.EqualFold(src, tgt) {
				mapping[src] = tgt
				break
			}
		}
	}

	// Semantic similarity matching (simplified - checks for word overlap)
	for _, src := range source {
		if _, mapped := mapping[src]; mapped {
			continue
		}

		bestMatch := ""
		bestScore := 0.0

		for _, tgt := range target {
			if _, used := ar.isValueInMap(mapping, tgt); used {
				continue
			}

			score := ar.semanticSimilarity(src, tgt)
			if score > bestScore && score > 0.3 { // Threshold
				bestScore = score
				bestMatch = tgt
			}
		}

		if bestMatch != "" {
			mapping[src] = bestMatch
		}
	}

	// Apply constraints to refine mapping
	for _, constraint := range constraints {
		ar.applyConstraint(mapping, constraint)
	}

	return mapping
}

// semanticSimilarity calculates similarity between two concepts
func (ar *AnalogicalReasoner) semanticSimilarity(concept1, concept2 string) float64 {
	// Simplified similarity: word overlap and length ratio
	words1 := strings.Fields(strings.ToLower(concept1))
	words2 := strings.Fields(strings.ToLower(concept2))

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	overlap := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 || strings.HasPrefix(w1, w2) || strings.HasPrefix(w2, w1) {
				overlap++
				break
			}
		}
	}

	maxLen := len(words1)
	if len(words2) > maxLen {
		maxLen = len(words2)
	}

	return float64(overlap) / float64(maxLen)
}

// calculateAnalogyStrength determines how strong the analogy is
func (ar *AnalogicalReasoner) calculateAnalogyStrength(mapping map[string]string, source, target []string) float64 {
	if len(source) == 0 {
		return 0.0
	}

	// Factors: coverage, bidirectionality, concept depth
	coverage := float64(len(mapping)) / float64(len(source))

	// Penalize if target has many unmapped concepts
	targetCoverage := float64(len(mapping)) / float64(max(len(target), 1))

	// Average the coverage metrics
	strength := (coverage + targetCoverage) / 2.0

	// Boost if high overlap
	if coverage > 0.7 {
		strength += 0.1
	}

	// Cap at 1.0
	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

// generateInsight creates insight from analogy mapping
func (ar *AnalogicalReasoner) generateInsight(source, target string, mapping map[string]string) string {
	if len(mapping) == 0 {
		return "No clear analogical mapping found between domains"
	}

	insights := []string{}

	// Describe the analogy
	insights = append(insights, fmt.Sprintf("The %s is analogous to %s", source, target))

	// Highlight key mappings
	if len(mapping) <= 5 {
		for src, tgt := range mapping {
			insights = append(insights, fmt.Sprintf("- '%s' maps to '%s'", src, tgt))
		}
	} else {
		count := 0
		for src, tgt := range mapping {
			if count < 3 {
				insights = append(insights, fmt.Sprintf("- '%s' maps to '%s'", src, tgt))
				count++
			}
		}
		insights = append(insights, fmt.Sprintf("... and %d more mappings", len(mapping)-3))
	}

	// Suggest insight
	insights = append(insights, "\nThis suggests that solutions from the source domain may transfer to the target domain with appropriate adaptation")

	return strings.Join(insights, "\n")
}

// transferInsights transfers insights from source to target based on analogy
func (ar *AnalogicalReasoner) transferInsights(analogy *types.Analogy, targetConcepts []string) []string {
	insights := []string{}

	// For each mapped concept, suggest how source domain insights apply
	for src, tgt := range analogy.Mapping {
		// Check if target concept is present
		if ar.contains(targetConcepts, tgt) {
			insight := fmt.Sprintf("Apply %s-based strategies to %s", src, tgt)
			insights = append(insights, insight)
		}
	}

	return insights
}

// identifyAdaptations identifies needed adaptations for analogy application
func (ar *AnalogicalReasoner) identifyAdaptations(analogy *types.Analogy, targetConcepts []string) []string {
	adaptations := []string{}

	// Find unmapped target concepts
	mappedTargets := make(map[string]bool)
	for _, tgt := range analogy.Mapping {
		mappedTargets[tgt] = true
	}

	for _, concept := range targetConcepts {
		if !mappedTargets[concept] {
			adaptation := fmt.Sprintf("Develop approach for '%s' (not present in source domain)", concept)
			adaptations = append(adaptations, adaptation)
		}
	}

	return adaptations
}

// applyConstraint refines mapping based on constraint
func (ar *AnalogicalReasoner) applyConstraint(mapping map[string]string, constraint string) {
	// Parse constraint (simplified - expects "src->tgt" format)
	parts := strings.Split(constraint, "->")
	if len(parts) == 2 {
		src := strings.TrimSpace(parts[0])
		tgt := strings.TrimSpace(parts[1])
		mapping[src] = tgt
	}
}

// Helper functions

func (ar *AnalogicalReasoner) contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func (ar *AnalogicalReasoner) isValueInMap(m map[string]string, value string) (string, bool) {
	for k, v := range m {
		if v == value {
			return k, true
		}
	}
	return "", false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
