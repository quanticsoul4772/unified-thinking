package reasoning_test

import (
	"fmt"
	"testing"

	"unified-thinking/internal/reasoning"
)

// BenchmarkBatchUpdateBelief compares batch API vs individual updates
func BenchmarkBatchUpdateBelief(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()
	belief, _ := pr.CreateBelief("Test hypothesis", 0.5)

	// Realistic evidence sequence
	updates := []reasoning.BeliefUpdate{
		{EvidenceID: "ev1", ProbEGivenH: 0.8, ProbEGivenNotH: 0.2},
		{EvidenceID: "ev2", ProbEGivenH: 0.9, ProbEGivenNotH: 0.1},
		{EvidenceID: "ev3", ProbEGivenH: 0.7, ProbEGivenNotH: 0.3},
		{EvidenceID: "ev4", ProbEGivenH: 0.85, ProbEGivenNotH: 0.15},
		{EvidenceID: "ev5", ProbEGivenH: 0.75, ProbEGivenNotH: 0.25},
	}

	b.Run("BatchAPI", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = pr.BatchUpdateBeliefFull(belief.ID, updates)
		}
	})

	b.Run("Individual", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, u := range updates {
				_, _ = pr.UpdateBeliefFull(belief.ID, u.EvidenceID, u.ProbEGivenH, u.ProbEGivenNotH)
			}
		}
	})
}

// BenchmarkBatchUpdateBelief_VaryingSize tests batch performance with different update counts
func BenchmarkBatchUpdateBelief_VaryingSize(b *testing.B) {
	sizes := []int{1, 5, 10, 20, 50}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			pr := reasoning.NewProbabilisticReasoner()
			belief, _ := pr.CreateBelief("Test", 0.5)

			updates := make([]reasoning.BeliefUpdate, size)
			for i := 0; i < size; i++ {
				updates[i] = reasoning.BeliefUpdate{
					EvidenceID:     fmt.Sprintf("ev%d", i),
					ProbEGivenH:    0.8,
					ProbEGivenNotH: 0.2,
				}
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = pr.BatchUpdateBeliefFull(belief.ID, updates)
			}
		})
	}
}

// BenchmarkCopyBelief measures belief copy performance
func BenchmarkCopyBelief(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()
	belief, _ := pr.CreateBelief("Test", 0.5)

	// Add some evidence
	for i := 0; i < 10; i++ {
		_, _ = pr.UpdateBeliefFull(belief.ID, fmt.Sprintf("ev%d", i), 0.8, 0.2)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = pr.GetBelief(belief.ID)
	}
}

// BenchmarkProductionScenario_Reasoning simulates realistic reasoning workload
func BenchmarkProductionScenario_Reasoning(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()
	belief, _ := pr.CreateBelief("Production hypothesis", 0.5)

	// Realistic evidence updates (10 pieces of evidence)
	updates := make([]reasoning.BeliefUpdate, 10)
	for i := 0; i < 10; i++ {
		updates[i] = reasoning.BeliefUpdate{
			EvidenceID:     fmt.Sprintf("evidence-%d", i),
			ProbEGivenH:    0.7 + float64(i%3)*0.1, // Vary likelihoods
			ProbEGivenNotH: 0.3 - float64(i%3)*0.05,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = pr.BatchUpdateBeliefFull(belief.ID, updates)
	}
}

// BenchmarkConcurrentBeliefUpdates measures performance under concurrent load
func BenchmarkConcurrentBeliefUpdates(b *testing.B) {
	pr := reasoning.NewProbabilisticReasoner()

	// Create multiple beliefs for concurrent access
	beliefs := make([]string, 10)
	for i := 0; i < 10; i++ {
		belief, _ := pr.CreateBelief(fmt.Sprintf("Belief %d", i), 0.5)
		beliefs[i] = belief.ID
	}

	updates := []reasoning.BeliefUpdate{
		{EvidenceID: "ev1", ProbEGivenH: 0.8, ProbEGivenNotH: 0.2},
		{EvidenceID: "ev2", ProbEGivenH: 0.9, ProbEGivenNotH: 0.1},
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			beliefID := beliefs[i%len(beliefs)]
			_, _ = pr.BatchUpdateBeliefFull(beliefID, updates)
			i++
		}
	})
}
