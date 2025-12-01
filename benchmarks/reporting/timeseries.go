// Package reporting provides timeseries storage for benchmark results.
package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// TimeSeriesEntry represents a single benchmark run in time series
type TimeSeriesEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	GitCommit     string                 `json:"git_commit"`
	Accuracy      map[string]float64     `json:"accuracy"`    // suite_name -> accuracy
	ECE           map[string]float64     `json:"ece"`         // suite_name -> ECE
	AvgLatency    map[string]int64       `json:"avg_latency"` // suite_name -> latency (nanoseconds)
	TotalCorrect  int                    `json:"total_correct"`
	TotalProblems int                    `json:"total_problems"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// TimeSeries stores historical benchmark results
type TimeSeries struct {
	Entries []*TimeSeriesEntry `json:"entries"`
}

// AddEntry adds a new entry to the time series
func (ts *TimeSeries) AddEntry(entry *TimeSeriesEntry) {
	ts.Entries = append(ts.Entries, entry)

	// Sort by timestamp (newest first)
	sort.Slice(ts.Entries, func(i, j int) bool {
		return ts.Entries[i].Timestamp.After(ts.Entries[j].Timestamp)
	})
}

// GetTrend returns accuracy trend for a specific suite
func (ts *TimeSeries) GetTrend(suiteName string, limit int) []float64 {
	trend := make([]float64, 0, limit)

	for i, entry := range ts.Entries {
		if i >= limit {
			break
		}

		if acc, exists := entry.Accuracy[suiteName]; exists {
			trend = append(trend, acc)
		}
	}

	// Reverse to get chronological order
	for i, j := 0, len(trend)-1; i < j; i, j = i+1, j-1 {
		trend[i], trend[j] = trend[j], trend[i]
	}

	return trend
}

// DetectRegression compares current to baseline
func (ts *TimeSeries) DetectRegression(suiteName string, currentAccuracy float64, threshold float64) (bool, string) {
	if len(ts.Entries) == 0 {
		return false, "No baseline available"
	}

	// Get baseline (most recent entry)
	baseline := 0.0
	for _, entry := range ts.Entries {
		if acc, exists := entry.Accuracy[suiteName]; exists {
			baseline = acc
			break
		}
	}

	if baseline == 0 {
		return false, fmt.Sprintf("No baseline for suite: %s", suiteName)
	}

	degradation := (baseline - currentAccuracy) / baseline

	if degradation > threshold {
		return true, fmt.Sprintf("Regression: %.2f%% vs baseline %.2f%% (%.1f%% drop)",
			currentAccuracy*100, baseline*100, degradation*100)
	}

	return false, fmt.Sprintf("OK: %.2f%% vs baseline %.2f%%", currentAccuracy*100, baseline*100)
}

// SaveToFile saves time series to JSON file
func (ts *TimeSeries) SaveToFile(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal time series: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadFromFile loads time series from JSON file
func LoadFromFile(path string) (*TimeSeries, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty time series if file doesn't exist
			return &TimeSeries{Entries: make([]*TimeSeriesEntry, 0)}, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var ts TimeSeries
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal time series: %w", err)
	}

	return &ts, nil
}

// GetLatestEntry returns the most recent entry
func (ts *TimeSeries) GetLatestEntry() *TimeSeriesEntry {
	if len(ts.Entries) == 0 {
		return nil
	}
	return ts.Entries[0]
}

// GetAverageAccuracy computes rolling average over last N entries
func (ts *TimeSeries) GetAverageAccuracy(suiteName string, windowSize int) float64 {
	if len(ts.Entries) == 0 {
		return 0.0
	}

	sum := 0.0
	count := 0

	for i, entry := range ts.Entries {
		if i >= windowSize {
			break
		}

		if acc, exists := entry.Accuracy[suiteName]; exists {
			sum += acc
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return sum / float64(count)
}
