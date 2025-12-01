// Package reinforcement implements Thompson Sampling for strategy selection.
package reinforcement

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ThompsonSelector implements Thompson Sampling for multi-armed bandits
type ThompsonSelector struct {
	strategies map[string]*Strategy
	rng        *rand.Rand
	mu         sync.RWMutex
}

// NewThompsonSelector creates a new Thompson Sampling selector
func NewThompsonSelector(seed int64) *ThompsonSelector {
	return &ThompsonSelector{
		strategies: make(map[string]*Strategy),
		rng:        rand.New(rand.NewSource(seed)), // #nosec G404 - Thompson Sampling RL algorithm, not security-sensitive
	}
}

// AddStrategy registers a new strategy with uniform prior (α=1, β=1)
func (ts *ThompsonSelector) AddStrategy(strategy *Strategy) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Initialize with uniform prior if not set
	if strategy.Alpha == 0 {
		strategy.Alpha = 1.0
	}
	if strategy.Beta == 0 {
		strategy.Beta = 1.0
	}

	ts.strategies[strategy.ID] = strategy
}

// SelectStrategy uses Thompson Sampling to select the best strategy
//
// Algorithm:
//  1. For each strategy, sample θ ~ Beta(α, β)
//  2. Select strategy with highest sampled θ
//  3. This balances exploration (uncertain strategies get chances) and
//     exploitation (proven strategies selected more often)
func (ts *ThompsonSelector) SelectStrategy(ctx ProblemContext) (*Strategy, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if len(ts.strategies) == 0 {
		return nil, fmt.Errorf("no strategies available")
	}

	var bestStrategy *Strategy
	maxSample := -1.0

	// Sample from each strategy's Beta distribution
	for _, strategy := range ts.strategies {
		if !strategy.IsActive {
			continue
		}

		// Sample θ ~ Beta(α, β)
		sample := SampleBeta(strategy.Alpha, strategy.Beta, ts.rng)

		if sample > maxSample {
			maxSample = sample
			bestStrategy = strategy
		}
	}

	if bestStrategy == nil {
		return nil, fmt.Errorf("no active strategies available")
	}

	return bestStrategy, nil
}

// RecordOutcome updates strategy parameters based on outcome
//
// Bayesian update:
// - Success: α ← α + 1 (increase successes)
// - Failure: β ← β + 1 (increase failures)
func (ts *ThompsonSelector) RecordOutcome(strategyID string, success bool) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	strategy, exists := ts.strategies[strategyID]
	if !exists {
		return fmt.Errorf("strategy not found: %s", strategyID)
	}

	// Bayesian update
	if success {
		strategy.Alpha += 1.0
		strategy.TotalSuccesses++
	} else {
		strategy.Beta += 1.0
	}

	strategy.TotalTrials++

	return nil
}

// GetStrategy retrieves a strategy by ID
func (ts *ThompsonSelector) GetStrategy(id string) (*Strategy, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	strategy, exists := ts.strategies[id]
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", id)
	}

	return strategy, nil
}

// GetAllStrategies returns all registered strategies
func (ts *ThompsonSelector) GetAllStrategies() []*Strategy {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	strategies := make([]*Strategy, 0, len(ts.strategies))
	for _, strategy := range ts.strategies {
		strategies = append(strategies, strategy)
	}

	return strategies
}

// ResetStrategy resets a strategy's state to uniform prior
func (ts *ThompsonSelector) ResetStrategy(id string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	strategy, exists := ts.strategies[id]
	if !exists {
		return fmt.Errorf("strategy not found: %s", id)
	}

	strategy.Alpha = 1.0
	strategy.Beta = 1.0
	strategy.TotalTrials = 0
	strategy.TotalSuccesses = 0

	return nil
}

// GetBestStrategy returns the strategy with highest empirical success rate
// This is useful for pure exploitation (no exploration)
func (ts *ThompsonSelector) GetBestStrategy() *Strategy {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	var best *Strategy
	maxRate := -1.0

	for _, strategy := range ts.strategies {
		if !strategy.IsActive {
			continue
		}

		rate := strategy.SuccessRate()
		if rate > maxRate {
			maxRate = rate
			best = strategy
		}
	}

	return best
}

// GetStrategyDistribution returns probability mass for each strategy
// Based on current Beta distributions (for monitoring)
func (ts *ThompsonSelector) GetStrategyDistribution(samples int) map[string]float64 {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	distribution := make(map[string]float64)
	counts := make(map[string]int)

	// Monte Carlo estimation
	for i := 0; i < samples; i++ {
		var bestID string
		maxSample := -1.0

		for id, strategy := range ts.strategies {
			if !strategy.IsActive {
				continue
			}

			sample := SampleBeta(strategy.Alpha, strategy.Beta, ts.rng)
			if sample > maxSample {
				maxSample = sample
				bestID = id
			}
		}

		if bestID != "" {
			counts[bestID]++
		}
	}

	// Convert counts to probabilities
	for id, count := range counts {
		distribution[id] = float64(count) / float64(samples)
	}

	return distribution
}

// NewThompsonSelectorWithTime creates a selector with time-based seed
func NewThompsonSelectorWithTime() *ThompsonSelector {
	return NewThompsonSelector(time.Now().UnixNano())
}
