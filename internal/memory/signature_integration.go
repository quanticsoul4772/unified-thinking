package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"unified-thinking/internal/embeddings"
)

// SignatureStorage defines the interface for storing context signatures
type SignatureStorage interface {
	StoreContextSignature(trajectoryID string, sig *ContextSignature) error
}

// ContextSignature represents a signature for cross-session context retrieval
type ContextSignature struct {
	TrajectoryID string
	Fingerprint  string
	Domain       string
	KeyConcepts  []string
	ToolSequence []string
	Complexity   float64
	Embedding    []float32 // Semantic embedding for similarity matching
}

// SignatureIntegration handles generating and storing context signatures for trajectories
type SignatureIntegration struct {
	storage   SignatureStorage
	extractor ConceptExtractor
	embedder  embeddings.Embedder // Optional embedder for semantic similarity
}

// ConceptExtractor extracts key concepts from text
type ConceptExtractor interface {
	Extract(text string) []string
}

// SimpleConceptExtractor uses basic tokenization and stop word filtering
type SimpleConceptExtractor struct {
	stopWords map[string]bool
}

// NewSimpleConceptExtractor creates a new simple concept extractor
func NewSimpleConceptExtractor() *SimpleConceptExtractor {
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
	return &SimpleConceptExtractor{stopWords: stopWords}
}

// Extract extracts key concepts from text
func (e *SimpleConceptExtractor) Extract(text string) []string {
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

// NewSignatureIntegration creates a new signature integration
func NewSignatureIntegration(storage SignatureStorage, extractor ConceptExtractor) *SignatureIntegration {
	if extractor == nil {
		extractor = NewSimpleConceptExtractor()
	}
	return &SignatureIntegration{
		storage:   storage,
		extractor: extractor,
	}
}

// SetEmbedder sets the embedder for semantic similarity
func (si *SignatureIntegration) SetEmbedder(embedder embeddings.Embedder) {
	si.embedder = embedder
}

// GenerateAndStoreSignature generates a context signature from a trajectory and stores it
// Embeddings are generated asynchronously to avoid blocking the main flow
func (si *SignatureIntegration) GenerateAndStoreSignature(trajectory *ReasoningTrajectory) error {
	if si.storage == nil || trajectory == nil {
		return nil
	}

	// Generate signature from trajectory
	sig := si.generateSignature(trajectory)
	if sig == nil {
		log.Printf("No signature generated for trajectory %s (no problem description)", trajectory.ID)
		return nil
	}

	// Store the signature immediately (without embedding)
	if err := si.storage.StoreContextSignature(trajectory.ID, sig); err != nil {
		log.Printf("Failed to store signature for trajectory %s: %v", trajectory.ID, err)
		return err
	}

	prefixLen := 8
	if len(sig.Fingerprint) < prefixLen {
		prefixLen = len(sig.Fingerprint)
	}
	log.Printf("Stored context signature for trajectory %s (fingerprint prefix: %s)",
		trajectory.ID, sig.Fingerprint[:prefixLen])

	// Generate embedding asynchronously if embedder is available
	if si.embedder != nil && trajectory.Problem != nil {
		go si.generateAndUpdateEmbedding(trajectory.ID, trajectory.Problem.Description, sig)
	}

	return nil
}

// generateAndUpdateEmbedding generates an embedding and updates the stored signature
func (si *SignatureIntegration) generateAndUpdateEmbedding(trajectoryID string, text string, sig *ContextSignature) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	embedding, err := si.embedder.Embed(ctx, text)
	if err != nil {
		log.Printf("[WARN] Failed to generate embedding for trajectory %s: %v", trajectoryID, err)
		return
	}

	if len(embedding) == 0 {
		log.Printf("[WARN] Empty embedding returned for trajectory %s", trajectoryID)
		return
	}

	// Update the signature with the embedding
	sig.Embedding = embedding
	if err := si.storage.StoreContextSignature(trajectoryID, sig); err != nil {
		log.Printf("[WARN] Failed to update signature with embedding for trajectory %s: %v", trajectoryID, err)
		return
	}

	log.Printf("[DEBUG] Updated signature with embedding for trajectory %s (%d dimensions)",
		trajectoryID, len(embedding))
}

// generateSignature creates a context signature from a trajectory
func (si *SignatureIntegration) generateSignature(trajectory *ReasoningTrajectory) *ContextSignature {
	if trajectory.Problem == nil || trajectory.Problem.Description == "" {
		return nil
	}

	sig := &ContextSignature{
		TrajectoryID: trajectory.ID,
		KeyConcepts:  []string{},
		ToolSequence: []string{},
	}

	// Generate fingerprint from problem description
	normalizedText := strings.ToLower(strings.TrimSpace(trajectory.Problem.Description))
	hash := sha256.Sum256([]byte(normalizedText))
	sig.Fingerprint = hex.EncodeToString(hash[:])

	// Extract concepts from problem description and context
	text := trajectory.Problem.Description
	if trajectory.Problem.Context != "" {
		text += " " + trajectory.Problem.Context
	}
	sig.KeyConcepts = si.extractor.Extract(text)

	// Set domain
	sig.Domain = trajectory.Domain
	if sig.Domain == "" && trajectory.Problem.Domain != "" {
		sig.Domain = trajectory.Problem.Domain
	}

	// Get tool sequence from approach
	if trajectory.Approach != nil && len(trajectory.Approach.ToolSequence) > 0 {
		sig.ToolSequence = trajectory.Approach.ToolSequence
	} else if len(trajectory.Steps) > 0 {
		// Extract from steps
		tools := make([]string, 0, len(trajectory.Steps))
		seen := make(map[string]bool)
		for _, step := range trajectory.Steps {
			if step.Tool != "" && !seen[step.Tool] {
				tools = append(tools, step.Tool)
				seen[step.Tool] = true
			}
		}
		sig.ToolSequence = tools
	}

	// Use trajectory complexity or estimate from problem
	sig.Complexity = trajectory.Complexity
	if sig.Complexity == 0 && trajectory.Problem.Complexity > 0 {
		sig.Complexity = trajectory.Problem.Complexity
	}
	if sig.Complexity == 0 {
		// Estimate from content
		wordCount := len(strings.Fields(text))
		conceptCount := len(sig.KeyConcepts)
		sig.Complexity = 0.3 + (float64(wordCount)/200.0)*0.4 + (float64(conceptCount)/20.0)*0.3
		if sig.Complexity > 1.0 {
			sig.Complexity = 1.0
		}
	}

	return sig
}
