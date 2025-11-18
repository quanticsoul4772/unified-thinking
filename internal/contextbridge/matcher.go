package contextbridge

import (
	"fmt"
	"log"
	"sort"
)

// SignatureStorage defines the storage interface needed by the matcher
type SignatureStorage interface {
	FindCandidatesWithSignatures(domain string, fingerprintPrefix string, limit int) ([]*CandidateWithSignature, error)
}

// Matcher finds similar trajectories based on signatures
type Matcher struct {
	storage    SignatureStorage
	similarity SimilarityCalculator
	extractor  ConceptExtractor
}

// NewMatcher creates a new signature matcher
func NewMatcher(storage SignatureStorage, similarity SimilarityCalculator, extractor ConceptExtractor) *Matcher {
	return &Matcher{
		storage:    storage,
		similarity: similarity,
		extractor:  extractor,
	}
}

// FindMatches finds trajectories similar to the given signature
func (m *Matcher) FindMatches(sig *Signature, minSimilarity float64, maxMatches int) ([]*Match, error) {
	if sig == nil {
		return nil, nil
	}

	// Validate matcher dependencies
	if m.storage == nil {
		return nil, fmt.Errorf("matcher storage is nil")
	}
	if m.similarity == nil {
		return nil, fmt.Errorf("matcher similarity calculator is nil")
	}

	// Get candidates with signatures in single query (avoids N+1)
	prefix := ""
	if len(sig.Fingerprint) >= 8 {
		prefix = sig.Fingerprint[:8]
	}

	// Single query returns candidates WITH their signatures
	candidates, err := m.storage.FindCandidatesWithSignatures(sig.Domain, prefix, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to find candidates: %w", err)
	}
	log.Printf("[DEBUG] Matcher retrieved %d candidates from storage", len(candidates))

	// Calculate similarity for each candidate - no additional queries needed
	// Pre-allocate with reasonable capacity to reduce allocations
	matches := make([]*Match, 0, min(len(candidates), maxMatches))
	for _, candidate := range candidates {
		if candidate == nil || candidate.Signature == nil {
			continue
		}

		similarity := m.similarity.Calculate(sig, candidate.Signature)

		if similarity >= minSimilarity {
			matches = append(matches, &Match{
				TrajectoryID: candidate.TrajectoryID,
				SessionID:    candidate.SessionID,
				Similarity:   similarity,
				Summary:      candidate.Description,
				SuccessScore: candidate.SuccessScore,
				QualityScore: candidate.QualityScore,
			})
		}
	}

	// Sort by similarity desc
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Similarity > matches[j].Similarity
	})

	// Return top N
	if len(matches) > maxMatches {
		matches = matches[:maxMatches]
	}

	return matches, nil
}

// GetExtractor returns the concept extractor
func (m *Matcher) GetExtractor() ConceptExtractor {
	return m.extractor
}
