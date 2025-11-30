// Package benchmarks provides RL-aware problem execution.
package benchmarks

import (
	"fmt"
	"time"

	"unified-thinking/internal/reinforcement"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// RLExecutor executes problems with Thompson Sampling RL tracking
type RLExecutor struct {
	storage           storage.Storage
	rlStorage         RLStorage
	outcomeThreshold  float64 // Confidence threshold for success
	trackOutcomes     bool    // Whether to record outcomes
	strategyID        string  // Last selected strategy ID
	strategySelections map[string]int
	strategySuccesses  map[string]int
}

// RLStorage defines the interface for RL persistence
type RLStorage interface {
	GetAllRLStrategies() ([]*reinforcement.Strategy, error)
	IncrementThompsonAlpha(strategyID string) error
	IncrementThompsonBeta(strategyID string) error
	RecordRLOutcome(outcome *reinforcement.Outcome) error
}

// NewRLExecutor creates a new RL-aware executor
func NewRLExecutor(store storage.Storage, rlStorage RLStorage, outcomeThreshold float64) *RLExecutor {
	return &RLExecutor{
		storage:            store,
		rlStorage:          rlStorage,
		outcomeThreshold:   outcomeThreshold,
		trackOutcomes:      true,
		strategySelections: make(map[string]int),
		strategySuccesses:  make(map[string]int),
	}
}

// Execute runs a problem with RL tracking
func (e *RLExecutor) Execute(problem *Problem, evaluator Evaluator) (*Result, error) {
	start := time.Now()

	// Extract input parameters
	content, ok := problem.Input["content"].(string)
	if !ok {
		content = problem.Description
	}

	mode, ok := problem.Input["mode"].(string)
	if !ok {
		mode = "auto" // Auto mode will use Thompson selector
	}

	// Create thought using the thinking system
	thought := types.NewThought().
		Content(content).
		Mode(types.ThinkingMode(mode)).
		Confidence(0.8).
		Build()

	// Store the thought
	if err := e.storage.StoreThought(thought); err != nil {
		return nil, fmt.Errorf("failed to store thought: %w", err)
	}

	latency := time.Since(start)

	// Evaluate response
	response := thought.Content
	correct, score, err := evaluator.Evaluate(response, problem.Expected)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	// Estimate tokens
	tokens := estimateTokens(content) + estimateTokens(response)

	result := &Result{
		ProblemID:  problem.ID,
		Correct:    correct,
		Score:      score,
		Confidence: thought.Confidence,
		Latency:    latency,
		Mode:       string(thought.Mode),
		Response:   response,
		Tokens:     tokens,
		Metadata:   make(map[string]interface{}),
	}

	// Record RL outcome if tracking is enabled
	if e.trackOutcomes && e.rlStorage != nil {
		e.recordRLOutcome(problem, result, latency.Nanoseconds())
	}

	return result, nil
}

// recordRLOutcome records the execution outcome for Thompson Sampling
func (e *RLExecutor) recordRLOutcome(problem *Problem, result *Result, executionTimeNs int64) {
	// Determine problem type
	problemType := problem.Category
	if problemType == "" {
		problemType = "general"
	}

	// Determine success based on confidence threshold
	success := result.Confidence >= e.outcomeThreshold

	// Get strategy ID from metadata (set by auto mode during execution)
	strategyID, ok := result.Metadata["strategy_id"].(string)
	if !ok || strategyID == "" {
		// If no strategy ID, try to infer from mode
		strategyID = fmt.Sprintf("strategy_%s", result.Mode)
	}

	// Track strategy selection
	e.strategySelections[strategyID]++
	if success {
		e.strategySuccesses[strategyID]++
	}
	e.strategyID = strategyID

	// Update Thompson state in database
	var err error
	if success {
		err = e.rlStorage.IncrementThompsonAlpha(strategyID)
	} else {
		err = e.rlStorage.IncrementThompsonBeta(strategyID)
	}

	if err != nil {
		// Don't fail the benchmark, just log
		fmt.Printf("Warning: failed to update Thompson state: %v\n", err)
	}

	// Record full outcome for analysis
	outcome := &reinforcement.Outcome{
		StrategyID:         strategyID,
		ProblemID:          problem.ID,
		ProblemType:        problemType,
		ProblemDescription: problem.Description,
		Success:            success,
		ConfidenceBefore:   0.8, // Default before execution
		ConfidenceAfter:    result.Confidence,
		ExecutionTimeNs:    executionTimeNs,
		TokenCount:         result.Tokens,
		Timestamp:          time.Now().Unix(),
		Metadata: map[string]interface{}{
			"category":   problem.Category,
			"difficulty": problem.Difficulty,
			"correct":    result.Correct,
			"score":      result.Score,
		},
	}

	if err := e.rlStorage.RecordRLOutcome(outcome); err != nil {
		fmt.Printf("Warning: failed to record RL outcome: %v\n", err)
	}

	// Store strategy ID in result metadata for reporting
	result.Metadata["strategy_id"] = strategyID
	result.Metadata["rl_success"] = success
}

// GetStrategyStats returns current strategy statistics
func (e *RLExecutor) GetStrategyStats() (map[string]int, map[string]int) {
	return e.strategySelections, e.strategySuccesses
}

// ResetStats clears strategy statistics
func (e *RLExecutor) ResetStats() {
	e.strategySelections = make(map[string]int)
	e.strategySuccesses = make(map[string]int)
}
