package contextbridge

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ConceptExtractor interface for swappable extraction strategies
type ConceptExtractor interface {
	Extract(text string) []string
}

// SimpleExtractor uses basic tokenization and stop word filtering
type SimpleExtractor struct {
	stopWords map[string]bool
}

// NewSimpleExtractor creates a new simple concept extractor
func NewSimpleExtractor() *SimpleExtractor {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"as": true, "is": true, "was": true, "be": true, "been": true,
		"are": true, "were": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true,
		"would": true, "could": true, "should": true, "may": true, "might": true,
		"must": true, "shall": true, "can": true, "need": true, "dare": true,
		"this": true, "that": true, "these": true, "those": true, "it": true,
		"its": true, "they": true, "them": true, "their": true, "we": true,
		"us": true, "our": true, "you": true, "your": true, "he": true,
		"him": true, "his": true, "she": true, "her": true, "i": true,
		"me": true, "my": true, "what": true, "which": true, "who": true,
		"whom": true, "when": true, "where": true, "why": true, "how": true,
		"all": true, "each": true, "every": true, "both": true, "few": true,
		"more": true, "most": true, "other": true, "some": true, "such": true,
		"no": true, "nor": true, "not": true, "only": true, "own": true,
		"same": true, "so": true, "than": true, "too": true, "very": true,
		"just": true, "also": true, "now": true, "then": true, "here": true,
		"there": true, "about": true, "after": true, "before": true, "between": true,
		"into": true, "through": true, "during": true, "above": true, "below": true,
		"up": true, "down": true, "out": true, "off": true, "over": true,
		"under": true, "again": true, "further": true, "once": true,
	}
	return &SimpleExtractor{stopWords: stopWords}
}

// Extract extracts key concepts from text
func (e *SimpleExtractor) Extract(text string) []string {
	words := strings.Fields(strings.ToLower(text))

	concepts := make([]string, 0)
	seen := make(map[string]bool)

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}/<>@#$%^&*-_=+`~")

		if len(word) > 3 && !e.stopWords[word] && !seen[word] {
			concepts = append(concepts, word)
			seen[word] = true
		}
	}

	return concepts
}

// ExtractSignature creates a signature from tool call parameters
func ExtractSignature(toolName string, params map[string]interface{}, extractor ConceptExtractor) (*Signature, error) {
	sig := &Signature{
		ToolSequence: []string{toolName},
		KeyConcepts:  []string{},
	}

	// Extract text content from various parameter names
	contentSources := []string{"content", "description", "problem", "situation", "question", "query", "input"}
	var textContent string

	for _, key := range contentSources {
		if val, ok := params[key].(string); ok && val != "" {
			textContent = val
			break
		}
	}

	if textContent == "" {
		return nil, nil // No extractable content
	}

	// Generate fingerprint
	normalizedText := strings.ToLower(strings.TrimSpace(textContent))
	hash := sha256.Sum256([]byte(normalizedText))
	sig.Fingerprint = hex.EncodeToString(hash[:])

	// Extract concepts
	sig.KeyConcepts = extractor.Extract(textContent)

	// Extract domain if present
	if domain, ok := params["domain"].(string); ok {
		sig.Domain = domain
	}

	// Estimate complexity based on content length and concept count
	wordCount := len(strings.Fields(textContent))
	conceptCount := len(sig.KeyConcepts)

	sig.Complexity = 0.3 + (float64(wordCount)/200.0)*0.4 + (float64(conceptCount)/20.0)*0.3
	if sig.Complexity > 1.0 {
		sig.Complexity = 1.0
	}

	return sig, nil
}
