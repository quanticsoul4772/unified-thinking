// Package metrics provides metrics collection for probabilistic reasoning operations
package metrics

import "sync/atomic"

// ProbabilisticMetrics tracks metrics for Bayesian inference operations
type ProbabilisticMetrics struct {
	updatesTotal         atomic.Int64 // Total belief updates attempted
	updatesUninformative atomic.Int64 // Updates where P(E|H) ≈ P(E|¬H)
	updatesError         atomic.Int64 // Failed updates (validation errors, etc.)
	beliefsCreated       atomic.Int64 // Total beliefs created
	beliefsCombined      atomic.Int64 // Belief combination operations
}

// NewProbabilisticMetrics creates a new probabilistic metrics tracker
func NewProbabilisticMetrics() *ProbabilisticMetrics {
	return &ProbabilisticMetrics{}
}

// RecordUpdate records a successful belief update
func (m *ProbabilisticMetrics) RecordUpdate() {
	m.updatesTotal.Add(1)
}

// RecordUninformative records an update where evidence was uninformative (P(E|H) ≈ P(E|¬H))
func (m *ProbabilisticMetrics) RecordUninformative() {
	m.updatesTotal.Add(1)
	m.updatesUninformative.Add(1)
}

// RecordError records a failed update attempt
func (m *ProbabilisticMetrics) RecordError() {
	m.updatesError.Add(1)
}

// RecordBeliefCreated records a new belief creation
func (m *ProbabilisticMetrics) RecordBeliefCreated() {
	m.beliefsCreated.Add(1)
}

// RecordBeliefsCombined records a belief combination operation
func (m *ProbabilisticMetrics) RecordBeliefsCombined() {
	m.beliefsCombined.Add(1)
}

// GetStats returns current metric values
func (m *ProbabilisticMetrics) GetStats() map[string]int64 {
	return map[string]int64{
		"updates_total":         m.updatesTotal.Load(),
		"updates_uninformative": m.updatesUninformative.Load(),
		"updates_error":         m.updatesError.Load(),
		"beliefs_created":       m.beliefsCreated.Load(),
		"beliefs_combined":      m.beliefsCombined.Load(),
	}
}

// GetUninformativeRate returns the percentage of updates that were uninformative
func (m *ProbabilisticMetrics) GetUninformativeRate() float64 {
	total := m.updatesTotal.Load()
	if total == 0 {
		return 0.0
	}

	uninformative := m.updatesUninformative.Load()
	return float64(uninformative) / float64(total)
}

// GetErrorRate returns the percentage of updates that failed
func (m *ProbabilisticMetrics) GetErrorRate() float64 {
	total := m.updatesTotal.Load() + m.updatesError.Load()
	if total == 0 {
		return 0.0
	}

	errors := m.updatesError.Load()
	return float64(errors) / float64(total)
}
