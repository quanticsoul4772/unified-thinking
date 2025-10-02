package metacognition

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// BiasDetector identifies cognitive biases in reasoning
type BiasDetector struct {
	mu      sync.RWMutex
	counter int
}

// NewBiasDetector creates a new bias detector
func NewBiasDetector() *BiasDetector {
	return &BiasDetector{}
}

// DetectBiases analyzes thought for cognitive biases
func (bd *BiasDetector) DetectBiases(thought *types.Thought) ([]*types.CognitiveBias, error) {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	return bd.detectBiasesInternal(thought), nil
}

// detectBiasesInternal is an internal helper that assumes the lock is already held
func (bd *BiasDetector) detectBiasesInternal(thought *types.Thought) []*types.CognitiveBias {
	biases := make([]*types.CognitiveBias, 0)

	content := strings.ToLower(thought.Content)

	// Check for confirmation bias
	if bias := bd.detectConfirmationBias(thought.ID, content); bias != nil {
		biases = append(biases, bias)
	}

	// Check for anchoring bias
	if bias := bd.detectAnchoringBias(thought.ID, content); bias != nil {
		biases = append(biases, bias)
	}

	// Check for availability bias
	if bias := bd.detectAvailabilityBias(thought.ID, content); bias != nil {
		biases = append(biases, bias)
	}

	// Check for sunk cost fallacy
	if bias := bd.detectSunkCostFallacy(thought.ID, content); bias != nil {
		biases = append(biases, bias)
	}

	// Check for overconfidence bias
	if bias := bd.detectOverconfidenceBias(thought, content); bias != nil {
		biases = append(biases, bias)
	}

	// Check for recency bias
	if bias := bd.detectRecencyBias(thought.ID, content); bias != nil {
		biases = append(biases, bias)
	}

	return biases
}

// DetectBiasesInBranch analyzes branch for cognitive biases
func (bd *BiasDetector) DetectBiasesInBranch(branch *types.Branch) ([]*types.CognitiveBias, error) {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	biases := make([]*types.CognitiveBias, 0)

	// Check each thought in branch
	for _, thought := range branch.Thoughts {
		thoughtBiases := bd.detectBiasesInternal(thought)
		biases = append(biases, thoughtBiases...)
	}

	// Check for groupthink (lack of divergence in branch)
	if bias := bd.detectGroupthink(branch); bias != nil {
		biases = append(biases, bias)
	}

	return biases, nil
}

// detectConfirmationBias detects tendency to favor confirming evidence
func (bd *BiasDetector) detectConfirmationBias(thoughtID, content string) *types.CognitiveBias {
	// Indicators: only positive evidence, dismissing alternatives
	confirmingIndicators := []string{
		"confirms", "supports", "validates", "proves",
		"as expected", "obviously", "clearly shows",
	}

	disconfirmingIndicators := []string{
		"contradicts", "challenges", "questions", "doubts",
		"however", "on the other hand", "alternatively",
	}

	confirmingCount := 0
	for _, indicator := range confirmingIndicators {
		if strings.Contains(content, indicator) {
			confirmingCount++
		}
	}

	disconfirmingCount := 0
	for _, indicator := range disconfirmingIndicators {
		if strings.Contains(content, indicator) {
			disconfirmingCount++
		}
	}

	// Strong confirmation bias if only confirming evidence presented
	if confirmingCount >= 2 && disconfirmingCount == 0 {
		return bd.createBias(
			"confirmation",
			"Potential confirmation bias: primarily focusing on supporting evidence",
			thoughtID,
			"medium",
			"Actively seek disconfirming evidence and alternative explanations",
		)
	}

	return nil
}

// detectAnchoringBias detects over-reliance on initial information
func (bd *BiasDetector) detectAnchoringBias(thoughtID, content string) *types.CognitiveBias {
	// Indicators: focus on first mentioned value, initial assessment
	anchorIndicators := []string{
		"initially", "at first", "starting from", "based on the first",
		"original estimate", "preliminary",
	}

	anchorCount := 0
	for _, indicator := range anchorIndicators {
		if strings.Contains(content, indicator) {
			anchorCount++
		}
	}

	// Check if there's insufficient adjustment mentioned
	adjustmentIndicators := []string{"adjusted", "revised", "updated", "reconsidered"}
	hasAdjustment := false
	for _, indicator := range adjustmentIndicators {
		if strings.Contains(content, indicator) {
			hasAdjustment = true
			break
		}
	}

	if anchorCount >= 1 && !hasAdjustment {
		return bd.createBias(
			"anchoring",
			"Potential anchoring bias: may be over-relying on initial information",
			thoughtID,
			"low",
			"Consider adjusting initial estimates based on new information",
		)
	}

	return nil
}

// detectAvailabilityBias detects over-reliance on readily available examples
func (bd *BiasDetector) detectAvailabilityBias(thoughtID, content string) *types.CognitiveBias {
	// Indicators: recent examples, vivid stories, personal anecdotes
	availabilityIndicators := []string{
		"recently", "just heard", "saw on the news", "everyone knows",
		"common knowledge", "obviously", "i remember when",
	}

	availabilityCount := 0
	for _, indicator := range availabilityIndicators {
		if strings.Contains(content, indicator) {
			availabilityCount++
		}
	}

	// Check for lack of statistical or comprehensive evidence
	systematicIndicators := []string{"study", "research", "data", "statistics", "survey"}
	hasSystematicEvidence := false
	for _, indicator := range systematicIndicators {
		if strings.Contains(content, indicator) {
			hasSystematicEvidence = true
			break
		}
	}

	if availabilityCount >= 2 && !hasSystematicEvidence {
		return bd.createBias(
			"availability",
			"Potential availability bias: relying on readily available examples rather than comprehensive data",
			thoughtID,
			"medium",
			"Seek systematic data and statistics beyond readily available examples",
		)
	}

	return nil
}

// detectSunkCostFallacy detects irrational commitment to past investments
func (bd *BiasDetector) detectSunkCostFallacy(thoughtID, content string) *types.CognitiveBias {
	// Indicators: references to past investment without forward-looking analysis
	sunkCostIndicators := []string{
		"already invested", "already spent", "committed so much",
		"waste", "too far in", "can't give up now",
	}

	sunkCostCount := 0
	for _, indicator := range sunkCostIndicators {
		if strings.Contains(content, indicator) {
			sunkCostCount++
		}
	}

	// Check for forward-looking reasoning
	forwardLookingIndicators := []string{"future benefit", "roi", "going forward", "from now on"}
	hasForwardLooking := false
	for _, indicator := range forwardLookingIndicators {
		if strings.Contains(content, indicator) {
			hasForwardLooking = true
			break
		}
	}

	if sunkCostCount >= 1 && !hasForwardLooking {
		return bd.createBias(
			"sunk_cost",
			"Potential sunk cost fallacy: decision may be influenced by past investments",
			thoughtID,
			"medium",
			"Focus on future costs and benefits rather than past investments",
		)
	}

	return nil
}

// detectOverconfidenceBias detects excessive confidence in judgment
func (bd *BiasDetector) detectOverconfidenceBias(thought *types.Thought, content string) *types.CognitiveBias {
	// High confidence with absolute language
	absoluteIndicators := []string{
		"certainly", "definitely", "absolutely", "without doubt",
		"impossible", "always", "never", "guaranteed",
	}

	absoluteCount := 0
	for _, indicator := range absoluteIndicators {
		if strings.Contains(content, indicator) {
			absoluteCount++
		}
	}

	// Check if confidence is high but content is short or lacks evidence
	if thought.Confidence > 0.85 && (len(content) < 100 || absoluteCount >= 2) {
		return bd.createBias(
			"overconfidence",
			"Potential overconfidence bias: high confidence with limited justification",
			thought.ID,
			"medium",
			"Acknowledge uncertainty and provide more supporting evidence",
		)
	}

	return nil
}

// detectRecencyBias detects over-weighting of recent information
func (bd *BiasDetector) detectRecencyBias(thoughtID, content string) *types.CognitiveBias {
	// Indicators: emphasis on recent events without historical context
	recencyIndicators := []string{
		"recently", "just now", "latest", "current trend",
		"nowadays", "these days", "in recent times",
	}

	recencyCount := 0
	for _, indicator := range recencyIndicators {
		if strings.Contains(content, indicator) {
			recencyCount++
		}
	}

	// Check for historical perspective
	historicalIndicators := []string{"historically", "in the past", "traditionally", "long-term"}
	hasHistorical := false
	for _, indicator := range historicalIndicators {
		if strings.Contains(content, indicator) {
			hasHistorical = true
			break
		}
	}

	if recencyCount >= 2 && !hasHistorical {
		return bd.createBias(
			"recency",
			"Potential recency bias: may be over-weighting recent information",
			thoughtID,
			"low",
			"Consider historical patterns and long-term trends",
		)
	}

	return nil
}

// detectGroupthink detects lack of critical evaluation in group thinking
func (bd *BiasDetector) detectGroupthink(branch *types.Branch) *types.CognitiveBias {
	if len(branch.Thoughts) < 3 {
		return nil // Need multiple thoughts to assess groupthink
	}

	// Check for diversity of perspectives
	disagreementIndicators := []string{
		"however", "but", "alternatively", "on the other hand",
		"disagree", "question", "challenge",
	}

	disagreementCount := 0
	for _, thought := range branch.Thoughts {
		content := strings.ToLower(thought.Content)
		for _, indicator := range disagreementIndicators {
			if strings.Contains(content, indicator) {
				disagreementCount++
				break
			}
		}
	}

	// If very few disagreements relative to thought count, potential groupthink
	if float64(disagreementCount)/float64(len(branch.Thoughts)) < 0.2 {
		return bd.createBias(
			"groupthink",
			"Potential groupthink: branch shows limited critical evaluation or alternative perspectives",
			branch.ID,
			"medium",
			"Encourage critical evaluation and consideration of alternative viewpoints",
		)
	}

	return nil
}

// createBias creates a cognitive bias record
func (bd *BiasDetector) createBias(biasType, description, detectedIn, severity, mitigation string) *types.CognitiveBias {
	bd.counter++
	return &types.CognitiveBias{
		ID:          fmt.Sprintf("bias-%d", bd.counter),
		BiasType:    biasType,
		Description: description,
		DetectedIn:  detectedIn,
		Severity:    severity,
		Mitigation:  mitigation,
		Metadata:    map[string]interface{}{},
		CreatedAt:   time.Now(),
	}
}
