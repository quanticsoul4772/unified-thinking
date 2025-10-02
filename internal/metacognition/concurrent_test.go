package metacognition

import (
	"sync"
	"testing"

	"unified-thinking/internal/types"
)

// TestSelfEvaluatorConcurrency tests concurrent access to SelfEvaluator
func TestSelfEvaluatorConcurrency(t *testing.T) {
	se := NewSelfEvaluator()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Create test thought
	thought := &types.Thought{
		ID:         "thought-1",
		Content:    "Test thought for evaluation",
		Mode:       "linear",
		Confidence: 0.8,
	}

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Evaluate thought
				_, err := se.EvaluateThought(thought)
				if err != nil {
					t.Errorf("EvaluateThought failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all evaluations were created
	expected := numGoroutines * operationsPerGoroutine
	if se.counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, se.counter)
	}
}

// TestBiasDetectorConcurrency tests concurrent access to BiasDetector
func TestBiasDetectorConcurrency(t *testing.T) {
	bd := NewBiasDetector()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 50

	// Create test thought with bias indicators to ensure counter increments
	thought := &types.Thought{
		ID:         "thought-1",
		Content:    "I'm absolutely certain this confirms my hypothesis. Everyone knows this is true and definitely correct.",
		Mode:       "linear",
		Confidence: 0.95,
	}

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Detect biases
				_, err := bd.DetectBiases(thought)
				if err != nil {
					t.Errorf("DetectBiases failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify counter incremented (confirms biases were detected)
	// Each operation detects 1 bias (overconfidence)
	expected := numGoroutines * operationsPerGoroutine
	if bd.counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, bd.counter)
	}
}

// TestBiasDetectorConcurrentReadWrite tests concurrent reads and writes
func TestBiasDetectorConcurrentReadWrite(t *testing.T) {
	bd := NewBiasDetector()

	// Create test thoughts with clear bias indicators
	thoughts := []*types.Thought{
		{
			ID:         "thought-1",
			Content:    "I'm absolutely certain, definitely guaranteed, this will work without doubt",
			Mode:       "linear",
			Confidence: 0.95,
		},
		{
			ID:      "thought-2",
			Content: "We've already invested so much money and committed resources, can't give up now",
			Mode:    "linear",
		},
	}

	var wg sync.WaitGroup
	numReaders := 5
	numWriters := 5
	iterations := 50

	// Start readers (detect biases)
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				thought := thoughts[j%len(thoughts)]
				_, err := bd.DetectBiases(thought)
				if err != nil {
					t.Errorf("DetectBiases failed: %v", err)
				}
			}
		}()
	}

	// Start writers (also detect biases, which increments counter)
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				thought := thoughts[j%len(thoughts)]
				_, err := bd.DetectBiases(thought)
				if err != nil {
					t.Errorf("DetectBiases failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify counter incremented (thought-1 detects overconfidence, thought-2 detects sunk cost)
	// Total operations = (numReaders + numWriters) * iterations = 500
	// Each operation detects 1 bias per thought
	expectedCount := (numReaders + numWriters) * iterations
	if bd.counter != expectedCount {
		t.Errorf("Expected counter %d, got %d", expectedCount, bd.counter)
	}
}
