package modes

import (
	"context"
	"log"
	"strings"

	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/types"
)

// Mode prototype examples for semantic matching
var modePrototypes = map[types.ThinkingMode][]string{
	types.ModeLinear: {
		"Step-by-step systematic problem solving",
		"Sequential reasoning approach",
		"Methodical analysis of the problem",
		"Break down the problem into steps",
		"Logical progression through the solution",
		"Analyze this systematically",
	},
	types.ModeTree: {
		"Explore multiple branches of thought",
		"Compare different alternatives",
		"Consider parallel approaches",
		"Branching analysis of options",
		"Evaluate multiple possibilities simultaneously",
		"What are the different options we could pursue",
	},
	types.ModeDivergent: {
		"Creative ideation and brainstorming",
		"Challenge all assumptions",
		"Think outside the box",
		"Unconventional radical approaches",
		"What if we completely reimagined this",
		"Break the rules and find new solutions",
	},
}

// AutoMode implements automatic mode detection
type AutoMode struct {
	linear          *LinearMode
	tree            *TreeMode
	divergent       *DivergentMode
	embedder        embeddings.Embedder
	prototypeEmbeds map[types.ThinkingMode][]float32 // Averaged prototype embeddings
}

// NewAutoMode creates a new auto mode detector
func NewAutoMode(linear *LinearMode, tree *TreeMode, divergent *DivergentMode) *AutoMode {
	return &AutoMode{
		linear:    linear,
		tree:      tree,
		divergent: divergent,
	}
}

// SetEmbedder sets the embedder for semantic mode detection
func (m *AutoMode) SetEmbedder(embedder embeddings.Embedder) {
	m.embedder = embedder
	if embedder != nil {
		m.initializePrototypes()
	}
}

// initializePrototypes computes averaged embeddings for each mode's prototypes
func (m *AutoMode) initializePrototypes() {
	if m.embedder == nil {
		log.Println("ERROR: embedder is nil in initializePrototypes")
		return
	}

	m.prototypeEmbeds = make(map[types.ThinkingMode][]float32)
	ctx := context.Background()

	for mode, examples := range modePrototypes {
		embeds, err := m.embedder.EmbedBatch(ctx, examples)
		if err != nil {
			log.Printf("ERROR: failed to embed prototypes for mode %s: %v", mode, err)
			continue
		}

		if len(embeds) == 0 {
			log.Printf("ERROR: EmbedBatch returned empty for mode %s", mode)
			continue
		}

		// Average the embeddings
		dim := len(embeds[0])
		averaged := make([]float32, dim)
		for _, emb := range embeds {
			for i, v := range emb {
				averaged[i] += v
			}
		}
		for i := range averaged {
			averaged[i] /= float32(len(embeds))
		}

		m.prototypeEmbeds[mode] = averaged
	}

	if len(m.prototypeEmbeds) == 3 {
		log.Println("Semantic auto mode selection enabled")
	} else {
		log.Printf("ERROR: Only %d/3 mode prototypes initialized, semantic detection disabled", len(m.prototypeEmbeds))
	}
}

// ProcessThought automatically selects the best mode and processes
func (m *AutoMode) ProcessThought(ctx context.Context, input ThoughtInput) (*ThoughtResult, error) {
	mode := m.detectMode(input)

	switch mode {
	case types.ModeLinear:
		return m.linear.ProcessThought(ctx, input)
	case types.ModeTree:
		return m.tree.ProcessThought(ctx, input)
	case types.ModeDivergent:
		return m.divergent.ProcessThought(ctx, input)
	default:
		return m.linear.ProcessThought(ctx, input)
	}
}

func (m *AutoMode) detectMode(input ThoughtInput) types.ThinkingMode {
	mode, confidence := m.detectModeWithConfidence(input)
	if confidence > 0 {
		log.Printf("Auto mode selected: %s (confidence: %.2f)", mode, confidence)
	}
	return mode
}

// detectModeWithConfidence returns the selected mode and confidence score
func (m *AutoMode) detectModeWithConfidence(input ThoughtInput) (types.ThinkingMode, float64) {
	content := strings.ToLower(input.Content)

	// Check ForceRebellion flag first - explicit request for divergent mode
	if input.ForceRebellion {
		return types.ModeDivergent, 1.0
	}

	// Tree structural indicators (checked before keywords to prioritize explicit structure)
	if input.BranchID != "" || len(input.CrossRefs) > 0 || len(input.KeyPoints) > 0 {
		return types.ModeTree, 1.0
	}

	// Use semantic detection if embedder is available
	if m.embedder != nil && len(m.prototypeEmbeds) == 3 {
		mode, confidence := m.detectModeSemantic(input.Content)
		if confidence < 0 {
			log.Printf("ERROR: Semantic mode detection failed, embedding error occurred")
		}
		return mode, confidence
	}

	// Keyword-based detection
	// Divergent indicators (checked before tree keywords to prioritize creative thinking)
	divergentKeywords := []string{
		"creative", "unconventional", "what if", "imagine", "challenge",
		"rebel", "outside the box", "innovative", "radical",
	}
	for _, kw := range divergentKeywords {
		if strings.Contains(content, kw) {
			return types.ModeDivergent, 0.8
		}
	}

	// Tree keywords (for exploration and branching)
	treeKeywords := []string{
		"branch", "explore", "alternative", "parallel", "compare",
		"multiple", "options", "possibilities",
	}
	for _, kw := range treeKeywords {
		if strings.Contains(content, kw) {
			return types.ModeTree, 0.8
		}
	}

	// Default to linear
	return types.ModeLinear, 0.5
}

// detectModeSemantic uses embeddings to determine the best mode
func (m *AutoMode) detectModeSemantic(content string) (types.ThinkingMode, float64) {
	ctx := context.Background()

	// Embed the input content
	inputEmbed, err := m.embedder.Embed(ctx, content)
	if err != nil {
		log.Printf("ERROR: failed to embed input for mode detection: %v", err)
		// Return error state - caller should handle this
		return types.ModeLinear, -1.0 // Negative confidence indicates error
	}

	// Calculate similarity to each mode prototype
	var bestMode types.ThinkingMode
	var bestSimilarity float64 = -1

	for mode, prototypeEmbed := range m.prototypeEmbeds {
		similarity := embeddings.CosineSimilarity(inputEmbed, prototypeEmbed)

		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMode = mode
		}
	}

	// Convert similarity to confidence (normalize from typical range)
	// Cosine similarity for semantic text typically ranges 0.3-0.8
	confidence := (bestSimilarity - 0.3) / 0.5 // Maps 0.3-0.8 to 0-1
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return bestMode, confidence
}
