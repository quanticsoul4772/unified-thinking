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
	Mode       string                 `json:"mode"` // Which thinking mode was used
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
}
