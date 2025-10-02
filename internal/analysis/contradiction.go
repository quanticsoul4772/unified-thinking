package analysis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// ContradictionDetector finds contradictions between thoughts
type ContradictionDetector struct {
	mu      sync.RWMutex
	counter int
}

// NewContradictionDetector creates a new contradiction detector
func NewContradictionDetector() *ContradictionDetector {
	return &ContradictionDetector{}
}

// DetectContradictions finds contradictions among a set of thoughts
func (cd *ContradictionDetector) DetectContradictions(thoughts []*types.Thought) ([]*types.Contradiction, error) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	contradictions := make([]*types.Contradiction, 0)

	// Compare all pairs of thoughts
	for i := 0; i < len(thoughts); i++ {
		for j := i + 1; j < len(thoughts); j++ {
			if contradiction := cd.checkPairForContradiction(thoughts[i], thoughts[j]); contradiction != nil {
				contradictions = append(contradictions, contradiction)
			}
		}
	}

	return contradictions, nil
}

// checkPairForContradiction checks if two thoughts contradict each other
func (cd *ContradictionDetector) checkPairForContradiction(t1, t2 *types.Thought) *types.Contradiction {
	content1 := strings.ToLower(t1.Content)
	content2 := strings.ToLower(t2.Content)

	// Check for direct negation patterns
	if contradiction := cd.detectDirectNegation(t1, t2, content1, content2); contradiction != nil {
		return contradiction
	}

	// Check for contradictory absolutes
	if contradiction := cd.detectContradictoryAbsolutes(t1, t2, content1, content2); contradiction != nil {
		return contradiction
	}

	// Check for contradictory claims about same subject
	if contradiction := cd.detectContradictoryClaims(t1, t2, content1, content2); contradiction != nil {
		return contradiction
	}

	return nil
}

// detectDirectNegation finds direct "X" vs "not X" contradictions
func (cd *ContradictionDetector) detectDirectNegation(t1, t2 *types.Thought, content1, content2 string) *types.Contradiction {
	// Extract key assertions
	assertions1 := cd.extractAssertions(content1)
	assertions2 := cd.extractAssertions(content2)

	negations := []string{"not ", "no ", "never ", "cannot ", "can't ", "won't ", "don't ", "doesn't "}

	for _, a1 := range assertions1 {
		for _, a2 := range assertions2 {
			// Check if one is negation of the other
			for _, neg := range negations {
				// a1 is "X", a2 is "not X"
				if strings.HasPrefix(a2, neg) {
					stripped := strings.TrimPrefix(a2, neg)
					if strings.Contains(a1, stripped) || strings.Contains(stripped, a1) {
						return cd.createContradiction(t1, t2, "direct_negation", "high",
							fmt.Sprintf("Direct contradiction: '%s' vs '%s'", a1, a2))
					}
				}
				// a2 is "X", a1 is "not X"
				if strings.HasPrefix(a1, neg) {
					stripped := strings.TrimPrefix(a1, neg)
					if strings.Contains(a2, stripped) || strings.Contains(stripped, a2) {
						return cd.createContradiction(t1, t2, "direct_negation", "high",
							fmt.Sprintf("Direct contradiction: '%s' vs '%s'", a1, a2))
					}
				}
			}
		}
	}

	return nil
}

// detectContradictoryAbsolutes finds contradictory absolute statements
func (cd *ContradictionDetector) detectContradictoryAbsolutes(t1, t2 *types.Thought, content1, content2 string) *types.Contradiction {
	// Always vs Never
	if strings.Contains(content1, "always") && strings.Contains(content2, "never") {
		return cd.createContradiction(t1, t2, "contradictory_absolutes", "high",
			"Contradictory absolutes: 'always' vs 'never'")
	}
	if strings.Contains(content1, "never") && strings.Contains(content2, "always") {
		return cd.createContradiction(t1, t2, "contradictory_absolutes", "high",
			"Contradictory absolutes: 'never' vs 'always'")
	}

	// All vs None
	if strings.Contains(content1, "all ") && strings.Contains(content2, "none ") {
		return cd.createContradiction(t1, t2, "contradictory_absolutes", "high",
			"Contradictory absolutes: 'all' vs 'none'")
	}
	if strings.Contains(content1, "none ") && strings.Contains(content2, "all ") {
		return cd.createContradiction(t1, t2, "contradictory_absolutes", "high",
			"Contradictory absolutes: 'none' vs 'all'")
	}

	// Must vs Cannot
	if strings.Contains(content1, "must ") && strings.Contains(content2, "cannot ") {
		return cd.createContradiction(t1, t2, "contradictory_modals", "medium",
			"Contradictory modals: 'must' vs 'cannot'")
	}
	if strings.Contains(content1, "cannot ") && strings.Contains(content2, "must ") {
		return cd.createContradiction(t1, t2, "contradictory_modals", "medium",
			"Contradictory modals: 'cannot' vs 'must'")
	}

	return nil
}

// detectContradictoryClaims finds contradictory claims about the same subject
func (cd *ContradictionDetector) detectContradictoryClaims(t1, t2 *types.Thought, content1, content2 string) *types.Contradiction {
	// Extract subjects (simple noun extraction)
	subjects1 := cd.extractSubjects(content1)
	subjects2 := cd.extractSubjects(content2)

	// Find common subjects
	for _, s1 := range subjects1 {
		for _, s2 := range subjects2 {
			if s1 == s2 {
				// Same subject - check for contradictory predicates
				if cd.hasContradictoryPredicates(content1, content2, s1) {
					return cd.createContradiction(t1, t2, "contradictory_claims", "medium",
						fmt.Sprintf("Contradictory claims about '%s'", s1))
				}
			}
		}
	}

	return nil
}

// extractAssertions extracts key assertions from content
func (cd *ContradictionDetector) extractAssertions(content string) []string {
	// Split by common sentence delimiters
	sentences := strings.FieldsFunc(content, func(r rune) bool {
		return r == '.' || r == '!' || r == '?' || r == ';'
	})

	assertions := make([]string, 0)
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 10 { // Minimum meaningful assertion length
			assertions = append(assertions, trimmed)
		}
	}
	return assertions
}

// extractSubjects extracts potential subjects from content (simplified)
func (cd *ContradictionDetector) extractSubjects(content string) []string {
	// Very simplified - extract nouns after "the", "a", "an"
	subjects := make([]string, 0)
	words := strings.Fields(content)

	articles := map[string]bool{"the": true, "a": true, "an": true}

	for i := 0; i < len(words)-1; i++ {
		if articles[words[i]] {
			subject := strings.TrimRight(words[i+1], ".,!?;:")
			if len(subject) > 2 {
				subjects = append(subjects, subject)
			}
		}
	}

	return subjects
}

// hasContradictoryPredicates checks if predicates contradict for same subject
func (cd *ContradictionDetector) hasContradictoryPredicates(content1, content2, subject string) bool {
	// Find sentence containing subject in each content
	sentences1 := cd.extractAssertions(content1)
	sentences2 := cd.extractAssertions(content2)

	for _, s1 := range sentences1 {
		if strings.Contains(s1, subject) {
			for _, s2 := range sentences2 {
				if strings.Contains(s2, subject) {
					// Check for contradictory verbs
					if cd.hasContradictoryVerbs(s1, s2) {
						return true
					}
				}
			}
		}
	}

	return false
}

// hasContradictoryVerbs checks for contradictory action verbs
func (cd *ContradictionDetector) hasContradictoryVerbs(s1, s2 string) bool {
	// Pairs of contradictory verbs
	contradictoryPairs := map[string][]string{
		"is":       {"is not", "isn't", "are not", "aren't"},
		"can":      {"cannot", "can't"},
		"will":     {"will not", "won't"},
		"should":   {"should not", "shouldn't"},
		"does":     {"does not", "doesn't"},
		"has":      {"has not", "hasn't"},
		"increases": {"decreases"},
		"improves":  {"worsens", "degrades"},
		"supports":  {"opposes", "contradicts"},
	}

	for positive, negatives := range contradictoryPairs {
		if strings.Contains(s1, positive) {
			for _, negative := range negatives {
				if strings.Contains(s2, negative) {
					return true
				}
			}
		}
		if strings.Contains(s2, positive) {
			for _, negative := range negatives {
				if strings.Contains(s1, negative) {
					return true
				}
			}
		}
	}

	return false
}

// createContradiction creates a contradiction record
func (cd *ContradictionDetector) createContradiction(t1, t2 *types.Thought, contradictionType, severity, description string) *types.Contradiction {
	cd.counter++
	return &types.Contradiction{
		ID:              fmt.Sprintf("contradiction-%d", cd.counter),
		ThoughtID1:      t1.ID,
		ThoughtID2:      t2.ID,
		ContradictoryAt: fmt.Sprintf("%s: %s", contradictionType, description),
		Severity:        severity,
		DetectedAt:      time.Now(),
	}
}
