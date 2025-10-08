// Package processing provides dual-process thinking architecture.
// Implements System 1 (fast, intuitive) and System 2 (slow, analytical) processing.
package processing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// ProcessingSystem represents the dual-process thinking system
type ProcessingSystem int

const (
	System1 ProcessingSystem = 1 // Fast, intuitive, heuristic-based
	System2 ProcessingSystem = 2 // Slow, analytical, deliberate
)

// DualProcessExecutor wraps existing thinking modes with System 1/2 logic
type DualProcessExecutor struct {
	storage storage.Storage
	modes   map[types.ThinkingMode]modes.ThinkingMode
}

// NewDualProcessExecutor creates a new dual-process executor
func NewDualProcessExecutor(store storage.Storage, modeRegistry map[types.ThinkingMode]modes.ThinkingMode) *DualProcessExecutor {
	return &DualProcessExecutor{
		storage: store,
		modes:   modeRegistry,
	}
}

// ProcessingRequest contains parameters for dual-process execution
type ProcessingRequest struct {
	Content          string
	Mode             types.ThinkingMode
	Confidence       float64
	KeyPoints        []string
	ForceSystem      ProcessingSystem // 0 = auto, 1 = System1, 2 = System2
	MaxSystem1Time   time.Duration    // Max time for System1 processing
	EscalateOnLowConf bool             // Escalate to System2 if confidence < threshold
	ConfidenceThreshold float64        // Threshold for escalation
}

// ProcessingResult contains the result of dual-process execution
type ProcessingResult struct {
	Thought        *types.Thought
	SystemUsed     ProcessingSystem
	Escalated      bool
	System1Time    time.Duration
	System2Time    time.Duration
	TotalTime      time.Duration
	ComplexityScore float64
	EscalationReason string
}

// ProcessThought executes dual-process thinking
func (dpe *DualProcessExecutor) ProcessThought(ctx context.Context, req *ProcessingRequest) (*ProcessingResult, error) {
	startTime := time.Now()

	complexity := dpe.calculateComplexity(req)

	result := &ProcessingResult{
		ComplexityScore: complexity,
	}

	// Step 1: Determine which system to use
	system := dpe.selectSystem(req, complexity)

	// If forced to specific system, use it
	if req.ForceSystem != 0 {
		system = req.ForceSystem
	}

	// Step 2: Execute System 1 (fast path)
	if system == System1 {
		thought, err := dpe.executeSystem1(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("System 1 failed: %w", err)
		}

		result.Thought = thought
		result.SystemUsed = System1
		result.System1Time = time.Since(startTime)
		result.TotalTime = result.System1Time

		// Check if escalation needed
		if dpe.shouldEscalate(thought, req) {
			escalatedThought, escalationReason, err := dpe.escalateToSystem2(ctx, req, thought)
			if err != nil {
				// If escalation fails, return System 1 result with note
				result.EscalationReason = fmt.Sprintf("Escalation attempted but failed: %v", err)
				return result, nil
			}

			result.Thought = escalatedThought
			result.SystemUsed = System2
			result.Escalated = true
			result.EscalationReason = escalationReason
			result.System2Time = time.Since(startTime) - result.System1Time
			result.TotalTime = time.Since(startTime)
		}
	} else {
		// Execute System 2 directly
		thought, err := dpe.executeSystem2(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("System 2 failed: %w", err)
		}

		result.Thought = thought
		result.SystemUsed = System2
		result.System2Time = time.Since(startTime)
		result.TotalTime = result.System2Time
	}

	return result, nil
}

// calculateComplexity estimates problem complexity
func (dpe *DualProcessExecutor) calculateComplexity(req *ProcessingRequest) float64 {
	score := 0.0

	// Factor 1: Content length (longer = more complex)
	contentLength := float64(len(req.Content))
	if contentLength < 100 {
		score += 0.1
	} else if contentLength < 300 {
		score += 0.3
	} else if contentLength < 600 {
		score += 0.5
	} else {
		score += 0.7
	}

	// Factor 2: Number of key points (more = more complex)
	if len(req.KeyPoints) > 5 {
		score += 0.3
	} else if len(req.KeyPoints) > 2 {
		score += 0.2
	} else if len(req.KeyPoints) > 0 {
		score += 0.1
	}

	// Factor 3: Complexity keywords
	contentLower := strings.ToLower(req.Content)
	complexityKeywords := []string{
		"why", "how", "analyze", "compare", "evaluate", "design",
		"complex", "multiple", "tradeoff", "balance", "optimize",
		"prove", "demonstrate", "derive", "explain in detail",
	}

	keywordCount := 0
	for _, keyword := range complexityKeywords {
		if strings.Contains(contentLower, keyword) {
			keywordCount++
		}
	}

	if keywordCount > 3 {
		score += 0.3
	} else if keywordCount > 1 {
		score += 0.2
	} else if keywordCount > 0 {
		score += 0.1
	}

	// Normalize to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// selectSystem chooses System 1 or System 2 based on complexity
func (dpe *DualProcessExecutor) selectSystem(req *ProcessingRequest, complexity float64) ProcessingSystem {
	// Simple heuristic: complexity < 0.4 = System 1, otherwise System 2
	if complexity < 0.4 {
		return System1
	}
	return System2
}

// executeSystem1 performs fast, heuristic-based processing
func (dpe *DualProcessExecutor) executeSystem1(ctx context.Context, req *ProcessingRequest) (*types.Thought, error) {
	// System 1: Fast pattern matching and heuristics
	// Use existing mode but with optimization flags

	mode := dpe.modes[req.Mode]
	if mode == nil {
		mode = dpe.modes[types.ModeLinear] // Fallback to linear
	}

	// Create thought with System 1 metadata
	result, err := mode.ProcessThought(ctx, modes.ThoughtInput{
		Content:    req.Content,
		Confidence: req.Confidence,
		KeyPoints:  req.KeyPoints,
	})

	if err != nil {
		return nil, err
	}

	// Retrieve the thought from storage to modify metadata
	thought, err := dpe.storage.GetThought(result.ThoughtID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve thought: %w", err)
	}

	// Add System 1 metadata
	if thought.Metadata == nil {
		thought.Metadata = make(map[string]interface{})
	}
	thought.Metadata["processing_system"] = "System1"
	thought.Metadata["processing_mode"] = "fast_heuristic"
	thought.Metadata["escalation_available"] = true

	// Store updated thought
	if err := dpe.storage.StoreThought(thought); err != nil {
		return nil, fmt.Errorf("failed to update thought: %w", err)
	}

	return thought, nil
}

// executeSystem2 performs slow, analytical processing
func (dpe *DualProcessExecutor) executeSystem2(ctx context.Context, req *ProcessingRequest) (*types.Thought, error) {
	// System 2: Slow, deliberate analytical processing
	// Use existing mode with full analysis

	mode := dpe.modes[req.Mode]
	if mode == nil {
		mode = dpe.modes[types.ModeLinear]
	}

	// Create thought with System 2 metadata
	result, err := mode.ProcessThought(ctx, modes.ThoughtInput{
		Content:    req.Content,
		Confidence: req.Confidence,
		KeyPoints:  req.KeyPoints,
	})

	if err != nil {
		return nil, err
	}

	// Retrieve the thought from storage to modify metadata
	thought, err := dpe.storage.GetThought(result.ThoughtID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve thought: %w", err)
	}

	// Add System 2 metadata
	if thought.Metadata == nil {
		thought.Metadata = make(map[string]interface{})
	}
	thought.Metadata["processing_system"] = "System2"
	thought.Metadata["processing_mode"] = "analytical_deliberate"
	thought.Metadata["full_analysis"] = true

	// Store updated thought
	if err := dpe.storage.StoreThought(thought); err != nil {
		return nil, fmt.Errorf("failed to update thought: %w", err)
	}

	return thought, nil
}

// shouldEscalate determines if System 1 result needs System 2 processing
func (dpe *DualProcessExecutor) shouldEscalate(thought *types.Thought, req *ProcessingRequest) bool {
	// Escalate if:
	// 1. Confidence below threshold
	if req.EscalateOnLowConf && thought.Confidence < req.ConfidenceThreshold {
		return true
	}

	// 2. Thought content suggests uncertainty
	contentLower := strings.ToLower(thought.Content)
	uncertaintyMarkers := []string{
		"maybe", "perhaps", "might", "could be", "unsure",
		"not certain", "unclear", "ambiguous",
	}

	for _, marker := range uncertaintyMarkers {
		if strings.Contains(contentLower, marker) {
			return true
		}
	}

	// 3. Content is too short for complex question
	complexity := dpe.calculateComplexity(req)
	if complexity > 0.6 && len(thought.Content) < 200 {
		return true
	}

	return false
}

// escalateToSystem2 re-processes thought with System 2
func (dpe *DualProcessExecutor) escalateToSystem2(ctx context.Context, req *ProcessingRequest, system1Thought *types.Thought) (*types.Thought, string, error) {
	// Determine escalation reason
	reason := "Low confidence"
	if system1Thought.Confidence < req.ConfidenceThreshold {
		reason = fmt.Sprintf("Low confidence (%.2f < %.2f)", system1Thought.Confidence, req.ConfidenceThreshold)
	}

	// Enhance request with System 1 context
	enhancedReq := &ProcessingRequest{
		Content:    fmt.Sprintf("SYSTEM 1 RESULT:\n%s\n\nREFINE AND ANALYZE IN DEPTH:", system1Thought.Content),
		Mode:       req.Mode,
		Confidence: req.Confidence,
		KeyPoints:  req.KeyPoints,
	}

	thought, err := dpe.executeSystem2(ctx, enhancedReq)
	if err != nil {
		return nil, reason, err
	}

	// Link to original System 1 thought
	if thought.Metadata == nil {
		thought.Metadata = make(map[string]interface{})
	}
	thought.Metadata["escalated_from_system1"] = true
	thought.Metadata["system1_thought_id"] = system1Thought.ID
	thought.Metadata["escalation_reason"] = reason
	thought.ParentID = system1Thought.ID

	return thought, reason, nil
}
