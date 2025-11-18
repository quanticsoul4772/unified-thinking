package contextbridge

import "math"

// SimilarityCalculator interface for swappable similarity algorithms
type SimilarityCalculator interface {
	Calculate(sig1, sig2 *Signature) float64
}

// WeightedSimilarity calculates similarity using weighted multi-factor scoring
type WeightedSimilarity struct {
	ConceptWeight    float64
	DomainWeight     float64
	ToolWeight       float64
	ComplexityWeight float64
}

// NewDefaultSimilarity creates a similarity calculator with default weights
func NewDefaultSimilarity() *WeightedSimilarity {
	return &WeightedSimilarity{
		ConceptWeight:    0.5,
		DomainWeight:     0.2,
		ToolWeight:       0.2,
		ComplexityWeight: 0.1,
	}
}

// Calculate computes similarity between two signatures
func (ws *WeightedSimilarity) Calculate(sig1, sig2 *Signature) float64 {
	if sig1 == nil || sig2 == nil {
		return 0.0
	}

	conceptSim := jaccardSimilarity(sig1.KeyConcepts, sig2.KeyConcepts)

	domainSim := 0.0
	if sig1.Domain == sig2.Domain && sig1.Domain != "" {
		domainSim = 1.0
	}

	toolSim := overlapRatio(sig1.ToolSequence, sig2.ToolSequence)

	complexitySim := 1.0 - math.Abs(sig1.Complexity-sig2.Complexity)

	return (conceptSim * ws.ConceptWeight) +
		(domainSim * ws.DomainWeight) +
		(toolSim * ws.ToolWeight) +
		(complexitySim * ws.ComplexityWeight)
}

// jaccardSimilarity calculates Jaccard similarity between two string slices
func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	setA := make(map[string]bool)
	for _, item := range a {
		setA[item] = true
	}

	intersection := 0
	for _, item := range b {
		if setA[item] {
			intersection++
		}
	}

	union := len(setA)
	for _, item := range b {
		if !setA[item] {
			union++
		}
	}

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// overlapRatio calculates overlap ratio between two string slices
func overlapRatio(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	setA := make(map[string]bool)
	for _, item := range a {
		setA[item] = true
	}

	overlap := 0
	for _, item := range b {
		if setA[item] {
			overlap++
		}
	}

	maxLen := math.Max(float64(len(a)), float64(len(b)))
	return float64(overlap) / maxLen
}
