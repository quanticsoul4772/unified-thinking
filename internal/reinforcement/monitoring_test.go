package reinforcement

import (
	"math"
	"strings"
	"testing"
)

// TestComputePerformanceMetrics tests performance metrics calculation
func TestComputePerformanceMetrics(t *testing.T) {
	strategy := &Strategy{
		ID:             "test_strategy",
		Name:           "Test Strategy",
		Mode:           "linear",
		Alpha:          10.0,
		Beta:           5.0,
		TotalTrials:    15,
		TotalSuccesses: 10,
		IsActive:       true,
	}

	metrics := ComputePerformanceMetrics(strategy)

	if metrics.StrategyID != "test_strategy" {
		t.Errorf("Expected ID test_strategy, got %s", metrics.StrategyID)
	}

	if metrics.TotalTrials != 15 {
		t.Errorf("Expected 15 trials, got %d", metrics.TotalTrials)
	}

	expectedSuccessRate := 10.0 / 15.0
	if metrics.SuccessRate != expectedSuccessRate {
		t.Errorf("Expected success rate %.3f, got %.3f", expectedSuccessRate, metrics.SuccessRate)
	}

	expectedRate := 10.0 / (10.0 + 5.0)
	if metrics.ExpectedRate != expectedRate {
		t.Errorf("Expected rate %.3f, got %.3f", expectedRate, metrics.ExpectedRate)
	}
}

// TestIsConverged tests convergence detection
func TestIsConverged(t *testing.T) {
	tests := []struct {
		name       string
		trials     int
		gap        float64
		threshold  float64
		expected   bool
	}{
		{"converged_sufficient_trials", 50, 0.03, 0.05, true},
		{"not_converged_large_gap", 50, 0.15, 0.05, false},
		{"not_converged_insufficient_trials", 10, 0.03, 0.05, false},
		{"converged_exactly_at_threshold", 20, 0.049, 0.05, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PerformanceMetrics{
				TotalTrials:    tt.trials,
				ConvergenceGap: tt.gap,
			}

			result := metrics.IsConverged(tt.threshold)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestComputeExplorationMetrics tests exploration metrics calculation
func TestComputeExplorationMetrics(t *testing.T) {
	selections := map[string]int{
		"strategy_1": 50, // Greedy
		"strategy_2": 30,
		"strategy_3": 20,
	}

	metrics := ComputeExplorationMetrics(selections, "strategy_1")

	if metrics.TotalSelections != 100 {
		t.Errorf("Expected 100 total, got %d", metrics.TotalSelections)
	}

	if metrics.GreedySelections != 50 {
		t.Errorf("Expected 50 greedy, got %d", metrics.GreedySelections)
	}

	if metrics.ExplorationCount != 50 {
		t.Errorf("Expected 50 exploration, got %d", metrics.ExplorationCount)
	}

	if metrics.ExplorationRate != 0.5 {
		t.Errorf("Expected 0.5 rate, got %.2f", metrics.ExplorationRate)
	}

	if metrics.UniqueStrategies != 3 {
		t.Errorf("Expected 3 unique strategies, got %d", metrics.UniqueStrategies)
	}

	// Shannon entropy should be > 0 for diverse distribution
	if metrics.StrategyDiversity <= 0 {
		t.Errorf("Expected positive entropy, got %.3f", metrics.StrategyDiversity)
	}
}

// TestNeedsMoreExploration tests exploration rate checking
func TestNeedsMoreExploration(t *testing.T) {
	tests := []struct {
		name     string
		rate     float64
		minRate  float64
		expected bool
	}{
		{"below_threshold", 0.10, 0.15, true},
		{"above_threshold", 0.25, 0.15, false},
		{"at_threshold", 0.15, 0.15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &ExplorationMetrics{
				ExplorationRate: tt.rate,
			}

			result := metrics.NeedsMoreExploration(tt.minRate)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsExploringTooMuch tests excessive exploration detection
func TestIsExploringTooMuch(t *testing.T) {
	tests := []struct {
		name     string
		rate     float64
		maxRate  float64
		expected bool
	}{
		{"below_max", 0.40, 0.50, false},
		{"above_max", 0.60, 0.50, true},
		{"at_max", 0.50, 0.50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &ExplorationMetrics{
				ExplorationRate: tt.rate,
			}

			result := metrics.IsExploringTooMuch(tt.maxRate)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestComputeLearningMetrics tests learning progress calculation
func TestComputeLearningMetrics(t *testing.T) {
	// Simulate improving learning: 50% initial → 80% final
	outcomes := make([]bool, 100)
	// First 20: 10/20 = 50%
	for i := 0; i < 10; i++ {
		outcomes[i] = true
	}
	// Last 20: 16/20 = 80%
	for i := 80; i < 96; i++ {
		outcomes[i] = true
	}

	metrics := ComputeLearningMetrics(outcomes, 20)

	if metrics.TotalTrials != 100 {
		t.Errorf("Expected 100 trials, got %d", metrics.TotalTrials)
	}

	if metrics.InitialAccuracy != 0.5 {
		t.Errorf("Expected initial 0.5, got %.2f", metrics.InitialAccuracy)
	}

	if metrics.CurrentAccuracy != 0.8 {
		t.Errorf("Expected current 0.8, got %.2f", metrics.CurrentAccuracy)
	}

	expectedImprovement := 0.3
	if math.Abs(metrics.AccuracyImprovement-expectedImprovement) > 0.01 {
		t.Errorf("Expected improvement %.2f, got %.2f", expectedImprovement, metrics.AccuracyImprovement)
	}

	if metrics.Trend != "improving" {
		t.Errorf("Expected improving trend, got %s", metrics.Trend)
	}
}

// TestLearningMetrics_Declining tests declining performance detection
func TestLearningMetrics_Declining(t *testing.T) {
	outcomes := make([]bool, 100)
	// First 20: 16/20 = 80%
	for i := 0; i < 16; i++ {
		outcomes[i] = true
	}
	// Last 20: 10/20 = 50%
	for i := 80; i < 90; i++ {
		outcomes[i] = true
	}

	metrics := ComputeLearningMetrics(outcomes, 20)

	if metrics.Trend != "declining" {
		t.Errorf("Expected declining trend, got %s", metrics.Trend)
	}

	if metrics.AccuracyImprovement >= 0 {
		t.Errorf("Expected negative improvement, got %.2f", metrics.AccuracyImprovement)
	}
}

// TestLearningMetrics_Stable tests stable performance detection
func TestLearningMetrics_Stable(t *testing.T) {
	outcomes := make([]bool, 100)
	// Consistent 50% throughout
	for i := 0; i < 50; i++ {
		outcomes[i*2] = true
	}

	metrics := ComputeLearningMetrics(outcomes, 20)

	if metrics.Trend != "stable" {
		t.Errorf("Expected stable trend, got %s", metrics.Trend)
	}

	if math.Abs(metrics.AccuracyImprovement) > 0.05 {
		t.Errorf("Expected near-zero improvement, got %.2f", metrics.AccuracyImprovement)
	}
}

// TestFormatPerformanceReport tests report formatting
func TestFormatPerformanceReport(t *testing.T) {
	metrics := &PerformanceMetrics{
		StrategyName:   "Test Strategy",
		Mode:           "linear",
		TotalTrials:    50,
		TotalSuccesses: 40,
		SuccessRate:    0.8,
		Alpha:          41.0,
		Beta:           11.0,
		ExpectedRate:   0.788,
		ConvergenceGap: 0.012,
	}

	report := FormatPerformanceReport(metrics)

	if !strings.Contains(report, "Test Strategy") {
		t.Error("Report should contain strategy name")
	}

	if !strings.Contains(report, "80.00%") {
		t.Error("Report should contain success rate")
	}

	if !strings.Contains(report, "α=41.00") {
		t.Error("Report should contain alpha parameter")
	}

	if !strings.Contains(report, "Converged") {
		t.Error("Report should show converged status")
	}
}

// TestFormatExplorationReport tests exploration report formatting
func TestFormatExplorationReport(t *testing.T) {
	metrics := &ExplorationMetrics{
		TotalSelections:   100,
		GreedySelections:  60,
		ExplorationCount:  40,
		ExplorationRate:   0.4,
		StrategyDiversity: 1.25,
		UniqueStrategies:  3,
	}

	report := FormatExplorationReport(metrics)

	if !strings.Contains(report, "100") {
		t.Error("Report should contain total selections")
	}

	if !strings.Contains(report, "40.0%") {
		t.Error("Report should contain exploration rate")
	}

	if !strings.Contains(report, "Balanced") {
		t.Error("Report should show balanced status")
	}
}

// TestFormatLearningReport tests learning report formatting
func TestFormatLearningReport(t *testing.T) {
	metrics := &LearningMetrics{
		InitialAccuracy:     0.5,
		CurrentAccuracy:     0.7,
		AccuracyImprovement: 0.2,
		TotalTrials:         100,
		LearningRate:        0.002,
		Trend:               "improving",
		HasConverged:        false,
	}

	report := FormatLearningReport(metrics)

	if !strings.Contains(report, "50.00%") {
		t.Error("Report should contain initial accuracy")
	}

	if !strings.Contains(report, "70.00%") {
		t.Error("Report should contain current accuracy")
	}

	if !strings.Contains(report, "+20.00%") {
		t.Error("Report should show positive improvement")
	}

	if !strings.Contains(report, "improving") {
		t.Error("Report should show improving trend")
	}
}

// TestExplorationMetrics_EdgeCases tests edge cases
func TestExplorationMetrics_EdgeCases(t *testing.T) {
	// Empty selections
	metrics := ComputeExplorationMetrics(map[string]int{}, "none")
	if metrics.TotalSelections != 0 {
		t.Error("Empty selections should have 0 total")
	}

	// Single strategy (no exploration possible)
	single := map[string]int{"strategy_1": 100}
	metrics = ComputeExplorationMetrics(single, "strategy_1")
	if metrics.ExplorationRate != 0.0 {
		t.Error("Single strategy should have 0 exploration")
	}

	if metrics.StrategyDiversity != 0.0 {
		t.Error("Single strategy should have 0 diversity")
	}
}

// TestLearningMetrics_EdgeCases tests learning metrics edge cases
func TestLearningMetrics_EdgeCases(t *testing.T) {
	// Empty outcomes
	metrics := ComputeLearningMetrics([]bool{}, 10)
	if metrics.TotalTrials != 0 {
		t.Error("Empty outcomes should have 0 trials")
	}

	// Single outcome
	single := []bool{true}
	metrics = ComputeLearningMetrics(single, 10)
	if metrics.TotalTrials != 1 {
		t.Error("Single outcome should have 1 trial")
	}

	// All failures
	failures := make([]bool, 50)
	metrics = ComputeLearningMetrics(failures, 10)
	if metrics.InitialAccuracy != 0.0 {
		t.Error("All failures should have 0 initial accuracy")
	}
	if metrics.CurrentAccuracy != 0.0 {
		t.Error("All failures should have 0 current accuracy")
	}
}
