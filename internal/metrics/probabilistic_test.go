package metrics_test

import (
	"sync"
	"testing"

	"unified-thinking/internal/metrics"
)

func TestNewProbabilisticMetrics(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()
	if m == nil {
		t.Fatal("NewProbabilisticMetrics returned nil")
	}

	// Verify initial state
	stats := m.GetStats()
	if stats["updates_total"] != 0 {
		t.Errorf("Expected initial updates_total = 0, got %d", stats["updates_total"])
	}
	if stats["beliefs_created"] != 0 {
		t.Errorf("Expected initial beliefs_created = 0, got %d", stats["beliefs_created"])
	}
}

func TestProbabilisticMetrics_RecordUpdate(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	// Record 5 updates
	for i := 0; i < 5; i++ {
		m.RecordUpdate()
	}

	stats := m.GetStats()
	if stats["updates_total"] != 5 {
		t.Errorf("Expected updates_total = 5, got %d", stats["updates_total"])
	}
	if stats["updates_uninformative"] != 0 {
		t.Errorf("Expected updates_uninformative = 0, got %d", stats["updates_uninformative"])
	}
}

func TestProbabilisticMetrics_RecordUninformative(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	// Record 3 uninformative updates
	for i := 0; i < 3; i++ {
		m.RecordUninformative()
	}

	stats := m.GetStats()
	if stats["updates_total"] != 3 {
		t.Errorf("Expected updates_total = 3, got %d", stats["updates_total"])
	}
	if stats["updates_uninformative"] != 3 {
		t.Errorf("Expected updates_uninformative = 3, got %d", stats["updates_uninformative"])
	}
}

func TestProbabilisticMetrics_RecordError(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	// Record 2 errors
	m.RecordError()
	m.RecordError()

	stats := m.GetStats()
	if stats["updates_error"] != 2 {
		t.Errorf("Expected updates_error = 2, got %d", stats["updates_error"])
	}
}

func TestProbabilisticMetrics_RecordBeliefCreated(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	// Create 10 beliefs
	for i := 0; i < 10; i++ {
		m.RecordBeliefCreated()
	}

	stats := m.GetStats()
	if stats["beliefs_created"] != 10 {
		t.Errorf("Expected beliefs_created = 10, got %d", stats["beliefs_created"])
	}
}

func TestProbabilisticMetrics_RecordBeliefsCombined(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	// Combine beliefs 4 times
	for i := 0; i < 4; i++ {
		m.RecordBeliefsCombined()
	}

	stats := m.GetStats()
	if stats["beliefs_combined"] != 4 {
		t.Errorf("Expected beliefs_combined = 4, got %d", stats["beliefs_combined"])
	}
}

func TestProbabilisticMetrics_GetUninformativeRate(t *testing.T) {
	tests := []struct {
		name          string
		totalUpdates  int
		uninformative int
		expectedRate  float64
	}{
		{"no updates", 0, 0, 0.0},
		{"all informative", 10, 0, 0.0},
		{"half uninformative", 10, 5, 0.5},
		{"all uninformative", 10, 10, 1.0},
		{"one uninformative", 100, 1, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metrics.NewProbabilisticMetrics()

			// Record normal updates
			for i := 0; i < tt.totalUpdates-tt.uninformative; i++ {
				m.RecordUpdate()
			}

			// Record uninformative updates
			for i := 0; i < tt.uninformative; i++ {
				m.RecordUninformative()
			}

			rate := m.GetUninformativeRate()
			if rate != tt.expectedRate {
				t.Errorf("Expected rate = %.2f, got %.2f", tt.expectedRate, rate)
			}
		})
	}
}

func TestProbabilisticMetrics_GetErrorRate(t *testing.T) {
	tests := []struct {
		name         string
		totalUpdates int
		errors       int
		expectedRate float64
	}{
		{"no operations", 0, 0, 0.0},
		{"no errors", 10, 0, 0.0},
		{"half errors", 10, 10, 0.5},
		{"all errors", 0, 10, 1.0},
		{"10% errors", 90, 10, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metrics.NewProbabilisticMetrics()

			// Record successful updates
			for i := 0; i < tt.totalUpdates; i++ {
				m.RecordUpdate()
			}

			// Record errors
			for i := 0; i < tt.errors; i++ {
				m.RecordError()
			}

			rate := m.GetErrorRate()
			if rate != tt.expectedRate {
				t.Errorf("Expected error rate = %.2f, got %.2f", tt.expectedRate, rate)
			}
		})
	}
}

func TestProbabilisticMetrics_ConcurrentAccess(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 5) // 5 operations per goroutine

	// Concurrent updates
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			m.RecordUpdate()
		}()
	}

	// Concurrent uninformative
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			m.RecordUninformative()
		}()
	}

	// Concurrent errors
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			m.RecordError()
		}()
	}

	// Concurrent belief creation
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			m.RecordBeliefCreated()
		}()
	}

	// Concurrent belief combination
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			m.RecordBeliefsCombined()
		}()
	}

	wg.Wait()

	// Verify counts
	stats := m.GetStats()
	if stats["updates_total"] != 200 { // 100 normal + 100 uninformative
		t.Errorf("Expected updates_total = 200, got %d", stats["updates_total"])
	}
	if stats["updates_uninformative"] != 100 {
		t.Errorf("Expected updates_uninformative = 100, got %d", stats["updates_uninformative"])
	}
	if stats["updates_error"] != 100 {
		t.Errorf("Expected updates_error = 100, got %d", stats["updates_error"])
	}
	if stats["beliefs_created"] != 100 {
		t.Errorf("Expected beliefs_created = 100, got %d", stats["beliefs_created"])
	}
	if stats["beliefs_combined"] != 100 {
		t.Errorf("Expected beliefs_combined = 100, got %d", stats["beliefs_combined"])
	}
}

func TestProbabilisticMetrics_GetStats(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	// Record mixed operations
	m.RecordUpdate()
	m.RecordUpdate()
	m.RecordUninformative()
	m.RecordError()
	m.RecordBeliefCreated()
	m.RecordBeliefsCombined()

	stats := m.GetStats()

	// Verify all keys present
	expectedKeys := []string{
		"updates_total",
		"updates_uninformative",
		"updates_error",
		"beliefs_created",
		"beliefs_combined",
	}

	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("Missing key in stats: %s", key)
		}
	}

	// Verify values
	if stats["updates_total"] != 3 { // 2 normal + 1 uninformative
		t.Errorf("Expected updates_total = 3, got %d", stats["updates_total"])
	}
}

func TestProbabilisticMetrics_RatesWithZeroOperations(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	// Test division by zero handling
	uninformativeRate := m.GetUninformativeRate()
	if uninformativeRate != 0.0 {
		t.Errorf("Expected uninformative rate = 0.0 with no operations, got %.2f", uninformativeRate)
	}

	errorRate := m.GetErrorRate()
	if errorRate != 0.0 {
		t.Errorf("Expected error rate = 0.0 with no operations, got %.2f", errorRate)
	}
}

func TestProbabilisticMetrics_MixedOperations(t *testing.T) {
	m := metrics.NewProbabilisticMetrics()

	// Simulate realistic usage pattern
	for i := 0; i < 100; i++ {
		m.RecordBeliefCreated()

		if i%10 == 0 {
			m.RecordUninformative() // 10% uninformative
		} else {
			m.RecordUpdate()
		}

		if i%50 == 0 {
			m.RecordBeliefsCombined()
		}

		if i%25 == 0 {
			m.RecordError() // 4% errors
		}
	}

	stats := m.GetStats()

	// Verify realistic statistics
	if stats["beliefs_created"] != 100 {
		t.Errorf("Expected 100 beliefs created, got %d", stats["beliefs_created"])
	}

	// 10 uninformative + 90 normal = 100 total updates
	if stats["updates_total"] != 100 {
		t.Errorf("Expected 100 total updates, got %d", stats["updates_total"])
	}

	uninformativeRate := m.GetUninformativeRate()
	if uninformativeRate < 0.09 || uninformativeRate > 0.11 {
		t.Errorf("Expected ~10%% uninformative rate, got %.2f%%", uninformativeRate*100)
	}
}
