// Package benchmarks provides testing infrastructure for measuring reasoning quality.
package benchmarks

import "time"

// Problem represents a benchmark problem to solve
type Problem struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Expected    interface{}            `json:"expected"`
	Category    string                 `json:"category"`
	Difficulty  string                 `json:"difficulty"` // easy, medium, hard
	Metadata    map[string]interface{} `json:"metadata"`
}

// Result represents the outcome of executing a benchmark problem
type Result struct {
	ProblemID  string                 `json:"problem_id"`
	Correct    bool                   `json:"correct"`
	Score      float64                `json:"score"`      // 0.0 to 1.0
	Confidence float64                `json:"confidence"` // Reported by system
	Latency    time.Duration          `json:"latency"`
	Mode       string                 `json:"mode"`   // Which thinking mode was used
	Tokens     int                    `json:"tokens"` // Estimated token count
	Response   string                 `json:"response"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// BenchmarkSuite represents a collection of related problems
type BenchmarkSuite struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Problems    []*Problem `json:"problems"`
	Category    string     `json:"category"`
}

// Evaluator evaluates benchmark responses
type Evaluator interface {
	// Evaluate compares the response to expected answer
	Evaluate(response interface{}, expected interface{}) (correct bool, score float64, err error)
	// Name returns the evaluator name
	Name() string
}

// BenchmarkRun represents a complete benchmark execution
type BenchmarkRun struct {
	RunID            string         `json:"run_id"`
	SuiteName        string         `json:"suite_name"`
	GitCommit        string         `json:"git_commit"`
	Timestamp        time.Time      `json:"timestamp"`
	Results          []*Result      `json:"results"`
	OverallAccuracy  float64        `json:"overall_accuracy"`
	OverallECE       float64        `json:"overall_ece"` // Expected Calibration Error
	AvgLatency       time.Duration  `json:"avg_latency"`
	TotalProblems    int            `json:"total_problems"`
	CorrectProblems  int            `json:"correct_problems"`
	ModeDistribution map[string]int `json:"mode_distribution"`

	// Thompson Sampling RL metrics
	RLEnabled               bool                   `json:"rl_enabled"`
	StrategyDistribution    map[string]int         `json:"strategy_distribution,omitempty"`     // Count by strategy
	StrategySuccessRate     map[string]float64     `json:"strategy_success_rate,omitempty"`     // Success rate by strategy
	LearningCurve           []float64              `json:"learning_curve,omitempty"`            // Accuracy over time (rolling window)
	InitialAccuracy         float64                `json:"initial_accuracy,omitempty"`          // First 20% of problems
	FinalAccuracy           float64                `json:"final_accuracy,omitempty"`            // Last 20% of problems
	AccuracyImprovement     float64                `json:"accuracy_improvement,omitempty"`      // Final - Initial
	ExplorationRate         float64                `json:"exploration_rate,omitempty"`          // % of time non-greedy strategy chosen
	StrategyDiversity       float64                `json:"strategy_diversity,omitempty"`        // Shannon entropy of strategy distribution
	RLMetadata              map[string]interface{} `json:"rl_metadata,omitempty"`
}
