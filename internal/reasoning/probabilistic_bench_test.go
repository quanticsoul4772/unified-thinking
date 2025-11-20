package reasoning_test

import (
	"fmt"
	"testing"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

// BenchmarkCreateBelief measures the performance of creating new beliefs
func BenchmarkCreateBelief(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pr.CreateBelief(fmt.Sprintf("Belief %d", i), 0.5)
	}
}

// BenchmarkUpdateBeliefFull measures the performance of Bayesian updates
func BenchmarkUpdateBeliefFull(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()

	// Setup: create a belief
	belief, _ := pr.CreateBelief("Benchmark test", 0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pr.UpdateBeliefFull(
			belief.ID,
			fmt.Sprintf("ev-%d", i),
			0.8, // P(E|H)
			0.2, // P(E|¬H)
		)
	}
}

// BenchmarkUpdateBeliefWithEvidence measures the performance of evidence-based updates
func BenchmarkUpdateBeliefWithEvidence(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()

	// Setup: create a belief
	belief, _ := pr.CreateBelief("Benchmark test", 0.5)

	// Create evidence
	evidence := &types.Evidence{
		ID:            "ev-bench",
		Content:       "Benchmark evidence",
		Source:        "Test",
		SupportsClaim: true,
		OverallScore:  0.75,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evidence.ID = fmt.Sprintf("ev-%d", i)
		_, _ = pr.UpdateBeliefWithEvidence(belief.ID, evidence)
	}
}

// BenchmarkCombineBeliefs_And measures the performance of combining beliefs with AND
func BenchmarkCombineBeliefs_And(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()

	// Setup: create 10 beliefs
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		belief, _ := pr.CreateBelief(fmt.Sprintf("Belief %d", i), 0.7)
		ids[i] = belief.ID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pr.CombineBeliefs(ids, "and")
	}
}

// BenchmarkCombineBeliefs_Or measures the performance of combining beliefs with OR
func BenchmarkCombineBeliefs_Or(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()

	// Setup: create 10 beliefs
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		belief, _ := pr.CreateBelief(fmt.Sprintf("Belief %d", i), 0.3)
		ids[i] = belief.ID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pr.CombineBeliefs(ids, "or")
	}
}

// BenchmarkSequentialUpdates measures performance of multiple sequential updates
// (simulates real-world usage pattern)
func BenchmarkSequentialUpdates(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		belief, _ := pr.CreateBelief("Sequential test", 0.5)

		// Apply 5 sequential updates
		for j := 0; j < 5; j++ {
			_, _ = pr.UpdateBeliefFull(
				belief.ID,
				fmt.Sprintf("ev-%d-%d", i, j),
				0.7+float64(j)*0.05,
				0.3-float64(j)*0.05,
			)
		}
	}
}

// BenchmarkGetMetrics measures the performance of metrics collection
func BenchmarkGetMetrics(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()

	// Setup: create some activity
	for i := 0; i < 100; i++ {
		belief, _ := pr.CreateBelief(fmt.Sprintf("Belief %d", i), 0.5)
		_, _ = pr.UpdateBeliefFull(belief.ID, fmt.Sprintf("ev-%d", i), 0.8, 0.2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pr.GetMetrics()
	}
}

// BenchmarkLikelihoodEstimator measures estimator performance
func BenchmarkLikelihoodEstimator(b *testing.B) {
	estimator := reasoning.NewStandardEstimator(nil)

	evidence := &types.Evidence{
		ID:            "ev-1",
		Content:       "Test evidence",
		Source:        "Test",
		SupportsClaim: true,
		OverallScore:  0.75,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = estimator.EstimateLikelihoods(evidence)
	}
}

// BenchmarkLikelihoodEstimator_Profiles compares different evidence profiles
func BenchmarkLikelihoodEstimator_Profiles(b *testing.B) {
	profiles := []*reasoning.EvidenceProfile{
		reasoning.DefaultProfile(),
		reasoning.ScientificProfile(),
		reasoning.AnecdotalProfile(),
	}

	evidence := &types.Evidence{
		ID:            "ev-1",
		SupportsClaim: true,
		OverallScore:  0.8,
	}

	for _, profile := range profiles {
		b.Run(profile.Name, func(b *testing.B) {
			estimator := reasoning.NewStandardEstimator(profile)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = estimator.EstimateLikelihoods(evidence)
			}
		})
	}
}
