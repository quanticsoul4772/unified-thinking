package modes

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/reinforcement"
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

// AutoMode implements automatic mode detection with optional Thompson Sampling RL
type AutoMode struct {
	linear          *LinearMode
	tree            *TreeMode
	divergent       *DivergentMode
	embedder        embeddings.Embedder
	prototypeEmbeds map[types.ThinkingMode][]float32

	// Thompson Sampling RL
	rlEnabled        bool
	thompsonSelector *reinforcement.ThompsonSelector
	outcomeThreshold float64                 // Confidence threshold for success (default 0.7)
	selectedStrategy *reinforcement.Strategy // Track last selected strategy for outcome recording
	storage          RLStorage               // Storage interface for loading/persisting RL state
}

// RLStorage defines the interface for RL persistence
type RLStorage interface {
	GetAllRLStrategies() ([]*reinforcement.Strategy, error)
	IncrementThompsonAlpha(strategyID string) error
	IncrementThompsonBeta(strategyID string) error
	RecordRLOutcome(outcome *reinforcement.Outcome) error
}

// NewAutoMode creates a new auto mode detector
func NewAutoMode(linear *LinearMode, tree *TreeMode, divergent *DivergentMode) *AutoMode {
	return &AutoMode{
		linear:           linear,
		tree:             tree,
		divergent:        divergent,
		outcomeThreshold: 0.7, // Default threshold
	}
}

// SetRLStorage enables Thompson Sampling RL by providing storage backend
// RL is ALWAYS enabled when SQLite storage is available (no flag required)
func (m *AutoMode) SetRLStorage(storage RLStorage) error {
	m.storage = storage

	// Load outcome threshold from environment (default: 0.7)
	if thresholdStr := os.Getenv("RL_OUTCOME_THRESHOLD"); thresholdStr != "" {
		if threshold, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			m.outcomeThreshold = threshold
		}
	}

	// Initialize Thompson selector
	m.thompsonSelector = reinforcement.NewThompsonSelectorWithTime()

	// Load strategies from storage
	strategies, err := storage.GetAllRLStrategies()
	if err != nil {
		log.Printf("Warning: failed to load RL strategies: %v", err)
		// Disable RL if we can't load strategies
		m.rlEnabled = false
		return err
	}

	if len(strategies) == 0 {
		log.Println("Warning: no RL strategies found in database")
		m.rlEnabled = false
		return nil
	}

	// Register strategies with Thompson selector
	for _, strategy := range strategies {
		m.thompsonSelector.AddStrategy(strategy)
	}

	// RL is now enabled
	m.rlEnabled = true

	log.Printf("Thompson Sampling RL enabled with %d strategies (threshold: %.2f)",
		len(strategies), m.outcomeThreshold)

	return nil
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
	startTime := time.Now()

	mode := m.detectMode(input)

	// Execute with selected mode
	result, err := m.executeMode(ctx, mode, input)

	// Record outcome if RL enabled and we have a selected strategy
	if m.rlEnabled && m.selectedStrategy != nil && result != nil {
		executionTimeNs := time.Since(startTime).Nanoseconds()
		m.recordOutcome(input, result, executionTimeNs)
	}

	return result, err
}

// executeMode executes the thought with the specified mode
func (m *AutoMode) executeMode(ctx context.Context, mode types.ThinkingMode, input ThoughtInput) (*ThoughtResult, error) {
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

	// Use Thompson Sampling RL if enabled
	if m.rlEnabled && m.thompsonSelector != nil {
		mode, confidence := m.detectModeRL(input)
		if mode != "" {
			return types.ThinkingMode(mode), confidence
		}
		// Fall through to semantic/keyword detection if RL fails
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

// detectModeRL uses Thompson Sampling to select mode
func (m *AutoMode) detectModeRL(input ThoughtInput) (string, float64) {
	// Determine problem type from content
	problemType := detectProblemType(input.Content)

	// Create problem context
	ctx := reinforcement.ProblemContext{
		Description: input.Content,
		Type:        problemType,
	}

	// Let Thompson selector choose strategy
	strategy, err := m.thompsonSelector.SelectStrategy(ctx)
	if err != nil {
		log.Printf("ERROR: Thompson selector failed: %v", err)
		return "", 0
	}

	// Store selected strategy for outcome recording
	m.selectedStrategy = strategy

	// Log selection
	log.Printf("Thompson selected: %s (α=%.2f, β=%.2f, rate=%.2f)",
		strategy.Name, strategy.Alpha, strategy.Beta, strategy.SuccessRate())

	// Return mode with high confidence (we trust Thompson)
	return strategy.Mode, 0.95
}

// detectProblemType analyzes content to determine problem category
func detectProblemType(content string) string {
	lower := strings.ToLower(content)

	// Causal reasoning indicators
	causalKeywords := []string{"cause", "effect", "intervention", "why", "because", "leads to"}
	for _, kw := range causalKeywords {
		if strings.Contains(lower, kw) {
			return "causal"
		}
	}

	// Probabilistic reasoning indicators
	probKeywords := []string{"probability", "likely", "chance", "uncertain", "odds", "bayesian"}
	for _, kw := range probKeywords {
		if strings.Contains(lower, kw) {
			return "probabilistic"
		}
	}

	// Logical reasoning indicators
	logicKeywords := []string{"if", "then", "therefore", "implies", "conclude", "prove"}
	for _, kw := range logicKeywords {
		if strings.Contains(lower, kw) {
			return "logical"
		}
	}

	// Default to general
	return "general"
}

// recordOutcome records the execution outcome for Thompson Sampling
func (m *AutoMode) recordOutcome(input ThoughtInput, result *ThoughtResult, executionTimeNs int64) {
	if m.storage == nil || m.selectedStrategy == nil {
		return
	}

	// Determine success based on confidence threshold
	success := result.Confidence >= m.outcomeThreshold

	// Update Thompson state in database
	var err error
	if success {
		err = m.storage.IncrementThompsonAlpha(m.selectedStrategy.ID)
	} else {
		err = m.storage.IncrementThompsonBeta(m.selectedStrategy.ID)
	}

	if err != nil {
		log.Printf("Warning: failed to update Thompson state: %v", err)
	}

	// Record full outcome for analysis
	outcome := &reinforcement.Outcome{
		StrategyID:         m.selectedStrategy.ID,
		ProblemID:          "", // Could generate from hash of content
		ProblemType:        detectProblemType(input.Content),
		ProblemDescription: input.Content,
		Success:            success,
		ConfidenceBefore:   input.Confidence,
		ConfidenceAfter:    result.Confidence,
		ExecutionTimeNs:    executionTimeNs,
		TokenCount:         0, // Could estimate from content length
		Timestamp:          time.Now().Unix(),
	}

	if err := m.storage.RecordRLOutcome(outcome); err != nil {
		log.Printf("Warning: failed to record RL outcome: %v", err)
	}

	// Update in-memory Thompson selector
	if err := m.thompsonSelector.RecordOutcome(m.selectedStrategy.ID, success); err != nil {
		log.Printf("Warning: failed to update Thompson selector: %v", err)
	}

	// Log outcome
	if success {
		log.Printf("RL outcome recorded: SUCCESS (confidence %.2f >= %.2f)",
			result.Confidence, m.outcomeThreshold)
	} else {
		log.Printf("RL outcome recorded: FAILURE (confidence %.2f < %.2f)",
			result.Confidence, m.outcomeThreshold)
	}

	// Clear selected strategy
	m.selectedStrategy = nil
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
