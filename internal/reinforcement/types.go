// Package reinforcement implements Thompson Sampling for adaptive reasoning mode selection.
//
// This package provides reinforcement learning capabilities that enable the unified-thinking
// server to learn which reasoning strategies work best for which types of problems through
// Bayesian optimization using Thompson Sampling bandits.
package reinforcement

import "unified-thinking/internal/types"

// Strategy represents a reasoning strategy with performance tracking
type Strategy struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Mode        string         `json:"mode"` // linear, tree, divergent, reflection, backtracking
	Parameters  types.Metadata `json:"parameters"`
	IsActive    bool           `json:"is_active"`

	// Thompson Sampling state
	Alpha          float64 `json:"alpha"` // Successes + 1
	Beta           float64 `json:"beta"`  // Failures + 1
	TotalTrials    int     `json:"total_trials"`
	TotalSuccesses int     `json:"total_successes"`
}

// SuccessRate computes empirical success rate
func (s *Strategy) SuccessRate() float64 {
	if s.TotalTrials == 0 {
		return 0.0
	}
	return float64(s.TotalSuccesses) / float64(s.TotalTrials)
}

// Outcome represents the result of executing a strategy
type Outcome struct {
	ID                 int            `json:"id"`
	StrategyID         string         `json:"strategy_id"`
	ProblemID          string         `json:"problem_id"`
	ProblemType        string         `json:"problem_type"`
	ProblemDescription string         `json:"problem_description"`
	Success            bool           `json:"success"`
	ConfidenceBefore   float64        `json:"confidence_before"`
	ConfidenceAfter    float64        `json:"confidence_after"`
	ExecutionTimeNs    int64          `json:"execution_time_ns"`
	TokenCount         int            `json:"token_count"`
	ReasoningPath      types.Metadata `json:"reasoning_path"`
	Timestamp          int64          `json:"timestamp"`
	Metadata           types.Metadata `json:"metadata"`
}

// ProblemContext provides context for strategy selection
type ProblemContext struct {
	Description string         `json:"description"`
	Type        string         `json:"type"` // logical, probabilistic, causal, etc.
	Metadata    types.Metadata `json:"metadata"`
}
