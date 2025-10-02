package reasoning

import (
	"sync"
	"testing"

	"unified-thinking/internal/types"
)

// TestProbabilisticReasonerConcurrency tests concurrent access to ProbabilisticReasoner
func TestProbabilisticReasonerConcurrency(t *testing.T) {
	pr := NewProbabilisticReasoner()

	// Create multiple goroutines that create and update beliefs concurrently
	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Create belief
				_, err := pr.CreateBelief("Test belief", 0.5)
				if err != nil {
					t.Errorf("CreateBelief failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all beliefs were created
	// We should have numGoroutines * operationsPerGoroutine beliefs
	// Since CreateBelief increments counter, we can check the counter value
	expected := numGoroutines * operationsPerGoroutine
	if pr.counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, pr.counter)
	}
}

// TestProbabilisticReasonerConcurrentReadWrite tests concurrent reads and writes
func TestProbabilisticReasonerConcurrentReadWrite(t *testing.T) {
	pr := NewProbabilisticReasoner()

	// Create some initial beliefs
	beliefIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		belief, err := pr.CreateBelief("Initial belief", 0.5)
		if err != nil {
			t.Fatalf("CreateBelief failed: %v", err)
		}
		beliefIDs[i] = belief.ID
	}

	var wg sync.WaitGroup
	numReaders := 5
	numWriters := 5
	iterations := 100

	// Start readers
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Read random belief
				beliefID := beliefIDs[j%len(beliefIDs)]
				_, err := pr.GetBelief(beliefID)
				if err != nil {
					t.Errorf("GetBelief failed: %v", err)
				}
			}
		}()
	}

	// Start writers
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Update random belief
				beliefID := beliefIDs[j%len(beliefIDs)]
				_, err := pr.UpdateBelief(beliefID, "evidence", 0.8, 0.5)
				if err != nil {
					t.Errorf("UpdateBelief failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()
}

// TestDecisionMakerConcurrency tests concurrent access to DecisionMaker
func TestDecisionMakerConcurrency(t *testing.T) {
	dm := NewDecisionMaker()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 50

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Create decision
				_, err := dm.CreateDecision(
					"Test decision",
					[]*types.DecisionOption{
						{ID: "opt1", Name: "Option 1", Scores: map[string]float64{"c1": 0.8}},
					},
					[]*types.DecisionCriterion{
						{ID: "c1", Name: "Criterion 1", Weight: 1.0, Maximize: true},
					},
				)
				if err != nil {
					t.Errorf("CreateDecision failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all decisions were created
	expected := numGoroutines * operationsPerGoroutine
	if dm.counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, dm.counter)
	}
}

// TestProblemDecomposerConcurrency tests concurrent access to ProblemDecomposer
func TestProblemDecomposerConcurrency(t *testing.T) {
	pd := NewProblemDecomposer()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 50

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Decompose problem
				_, err := pd.DecomposeProblem("Test problem")
				if err != nil {
					t.Errorf("DecomposeProblem failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all decompositions were created
	expected := numGoroutines * operationsPerGoroutine
	if pd.counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, pd.counter)
	}
}
