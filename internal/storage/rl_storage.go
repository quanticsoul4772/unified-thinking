// Package storage provides RL-specific storage methods for Thompson Sampling.
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"unified-thinking/internal/reinforcement"
)

// StoreRLStrategy stores a reasoning strategy
func (s *SQLiteStorage) StoreRLStrategy(strategy *reinforcement.Strategy) error {
	paramsJSON, err := json.Marshal(strategy.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO rl_strategies (id, name, description, mode, parameters, created_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, strategy.ID, strategy.Name, strategy.Description, strategy.Mode,
		string(paramsJSON), time.Now().Unix(), boolToInt(strategy.IsActive))

	return err
}

// GetRLStrategy retrieves a strategy by ID
func (s *SQLiteStorage) GetRLStrategy(id string) (*reinforcement.Strategy, error) {
	var name, description, mode, paramsJSON string
	var createdAt int64
	var isActive int

	err := s.db.QueryRow(`
		SELECT name, description, mode, parameters, created_at, is_active
		FROM rl_strategies
		WHERE id = ?
	`, id).Scan(&name, &description, &mode, &paramsJSON, &createdAt, &isActive)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("strategy not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query strategy: %w", err)
	}

	var params map[string]interface{}
	if paramsJSON != "" {
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			params = make(map[string]interface{})
		}
	}

	strategy := &reinforcement.Strategy{
		ID:          id,
		Name:        name,
		Description: description,
		Mode:        mode,
		Parameters:  params,
		IsActive:    isActive == 1,
	}

	// Load Thompson state
	var alpha, beta float64
	var totalTrials, totalSuccesses int
	err = s.db.QueryRow(`
		SELECT alpha, beta, total_trials, total_successes
		FROM rl_thompson_state
		WHERE strategy_id = ?
	`, id).Scan(&alpha, &beta, &totalTrials, &totalSuccesses)

	if err == nil {
		strategy.Alpha = alpha
		strategy.Beta = beta
		strategy.TotalTrials = totalTrials
		strategy.TotalSuccesses = totalSuccesses
	}

	return strategy, nil
}

// GetAllRLStrategies retrieves all active strategies with their Thompson state
func (s *SQLiteStorage) GetAllRLStrategies() ([]*reinforcement.Strategy, error) {
	rows, err := s.db.Query(`
		SELECT
			s.id, s.name, s.description, s.mode, s.parameters, s.is_active,
			COALESCE(ts.alpha, 1.0), COALESCE(ts.beta, 1.0),
			COALESCE(ts.total_trials, 0), COALESCE(ts.total_successes, 0)
		FROM rl_strategies s
		LEFT JOIN rl_thompson_state ts ON s.id = ts.strategy_id
		WHERE s.is_active = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query strategies: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var strategies []*reinforcement.Strategy
	for rows.Next() {
		var id, name, description, mode, paramsJSON string
		var isActive int
		var alpha, beta float64
		var totalTrials, totalSuccesses int

		err := rows.Scan(&id, &name, &description, &mode, &paramsJSON, &isActive,
			&alpha, &beta, &totalTrials, &totalSuccesses)
		if err != nil {
			return nil, fmt.Errorf("failed to scan strategy: %w", err)
		}

		var params map[string]interface{}
		if paramsJSON != "" {
			_ = json.Unmarshal([]byte(paramsJSON), &params)
		}

		strategies = append(strategies, &reinforcement.Strategy{
			ID:             id,
			Name:           name,
			Description:    description,
			Mode:           mode,
			Parameters:     params,
			IsActive:       isActive == 1,
			Alpha:          alpha,
			Beta:           beta,
			TotalTrials:    totalTrials,
			TotalSuccesses: totalSuccesses,
		})
	}

	return strategies, rows.Err()
}

// RecordRLOutcome stores a strategy execution outcome
func (s *SQLiteStorage) RecordRLOutcome(outcome *reinforcement.Outcome) error {
	reasoningPathJSON, _ := json.Marshal(outcome.ReasoningPath)
	metadataJSON, _ := json.Marshal(outcome.Metadata)

	_, err := s.db.Exec(`
		INSERT INTO rl_strategy_outcomes (
			strategy_id, problem_id, problem_type, problem_description,
			success, confidence_before, confidence_after, execution_time_ns,
			token_count, reasoning_path, timestamp, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		outcome.StrategyID,
		outcome.ProblemID,
		outcome.ProblemType,
		outcome.ProblemDescription,
		boolToInt(outcome.Success),
		outcome.ConfidenceBefore,
		outcome.ConfidenceAfter,
		outcome.ExecutionTimeNs,
		outcome.TokenCount,
		string(reasoningPathJSON),
		outcome.Timestamp,
		string(metadataJSON),
	)

	return err
}

// UpdateThompsonState updates alpha/beta parameters for a strategy
func (s *SQLiteStorage) UpdateThompsonState(strategyID string, alpha, beta float64, totalTrials, totalSuccesses int) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO rl_thompson_state
		(strategy_id, alpha, beta, total_trials, total_successes, last_updated)
		VALUES (?, ?, ?, ?, ?, ?)
	`, strategyID, alpha, beta, totalTrials, totalSuccesses, time.Now().Unix())

	return err
}

// IncrementThompsonAlpha increments alpha (success)
func (s *SQLiteStorage) IncrementThompsonAlpha(strategyID string) error {
	_, err := s.db.Exec(`
		UPDATE rl_thompson_state
		SET alpha = alpha + 1,
		    total_successes = total_successes + 1,
		    total_trials = total_trials + 1,
		    last_updated = ?
		WHERE strategy_id = ?
	`, time.Now().Unix(), strategyID)

	return err
}

// IncrementThompsonBeta increments beta (failure)
func (s *SQLiteStorage) IncrementThompsonBeta(strategyID string) error {
	_, err := s.db.Exec(`
		UPDATE rl_thompson_state
		SET beta = beta + 1,
		    total_trials = total_trials + 1,
		    last_updated = ?
		WHERE strategy_id = ?
	`, time.Now().Unix(), strategyID)

	return err
}

// GetRLStrategyPerformance retrieves aggregated performance metrics
func (s *SQLiteStorage) GetRLStrategyPerformance() ([]*reinforcement.Strategy, error) {
	rows, err := s.db.Query(`
		SELECT id, name, mode, trials, successes, success_rate, alpha, beta
		FROM rl_strategy_performance
		ORDER BY success_rate DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query performance: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var strategies []*reinforcement.Strategy
	for rows.Next() {
		var id, name, mode string
		var trials, successes int
		var successRate, alpha, beta float64

		err := rows.Scan(&id, &name, &mode, &trials, &successes, &successRate, &alpha, &beta)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance: %w", err)
		}

		strategies = append(strategies, &reinforcement.Strategy{
			ID:             id,
			Name:           name,
			Mode:           mode,
			Alpha:          alpha,
			Beta:           beta,
			TotalTrials:    trials,
			TotalSuccesses: successes,
			IsActive:       true,
		})
	}

	return strategies, rows.Err()
}
