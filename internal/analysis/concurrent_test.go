package analysis

import (
	"sync"
	"testing"

	"unified-thinking/internal/types"
)

// TestEvidenceAnalyzerConcurrency tests concurrent access to EvidenceAnalyzer
func TestEvidenceAnalyzerConcurrency(t *testing.T) {
	ea := NewEvidenceAnalyzer()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Assess evidence
				_, err := ea.AssessEvidence("Test evidence", "Test source", "claim-1", true)
				if err != nil {
					t.Errorf("AssessEvidence failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all evidence was created
	expected := numGoroutines * operationsPerGoroutine
	if ea.counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, ea.counter)
	}
}

// TestContradictionDetectorConcurrency tests concurrent access to ContradictionDetector
func TestContradictionDetectorConcurrency(t *testing.T) {
	cd := NewContradictionDetector()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 50

	// Create test thoughts
	thoughts := []*types.Thought{
		{
			ID:      "thought-1",
			Content: "All remote work is always productive",
			Mode:    "linear",
		},
		{
			ID:      "thought-2",
			Content: "Remote work never increases productivity",
			Mode:    "linear",
		},
	}

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Detect contradictions
				_, err := cd.DetectContradictions(thoughts)
				if err != nil {
					t.Errorf("DetectContradictions failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all detections were processed
	expected := numGoroutines * operationsPerGoroutine
	if cd.counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, cd.counter)
	}
}

// TestSensitivityAnalyzerConcurrency tests concurrent access to SensitivityAnalyzer
func TestSensitivityAnalyzerConcurrency(t *testing.T) {
	sa := NewSensitivityAnalyzer()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 50

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Analyze sensitivity
				_, err := sa.AnalyzeSensitivity(
					"Test claim",
					[]string{"assumption1", "assumption2"},
					0.8,
				)
				if err != nil {
					t.Errorf("AnalyzeSensitivity failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all analyses were created
	expected := numGoroutines * operationsPerGoroutine
	if sa.counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, sa.counter)
	}
}
