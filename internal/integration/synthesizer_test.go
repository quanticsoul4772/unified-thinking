package integration

import (
	"strings"
	"testing"
)

func TestNewSynthesizer(t *testing.T) {
	s := NewSynthesizer()
	if s == nil {
		t.Fatal("NewSynthesizer() returned nil")
	}
	if s.syntheses == nil {
		t.Fatal("syntheses map not initialized")
	}
}

func TestSynthesizeInsights(t *testing.T) {
	s := NewSynthesizer()

	tests := []struct {
		name         string
		inputs       []*Input
		context      string
		expectError  bool
		minSynergies int
		minConflicts int
	}{
		{
			name:        "insufficient inputs",
			inputs:      []*Input{{ID: "1", Mode: "causal", Content: "Test", Confidence: 0.8}},
			expectError: true,
		},
		{
			name: "causal + temporal synergy",
			inputs: []*Input{
				{ID: "1", Mode: "causal", Content: "A causes B which leads to C", Confidence: 0.8},
				{ID: "2", Mode: "temporal", Content: "Short-term: immediate impact. Long-term: sustained effects", Confidence: 0.7},
			},
			context:      "Decision analysis",
			expectError:  false,
			minSynergies: 1,
		},
		{
			name: "causal + probabilistic synergy",
			inputs: []*Input{
				{ID: "3", Mode: "causal", Content: "Temperature increases sales", Confidence: 0.9},
				{ID: "4", Mode: "probabilistic", Content: "Probability of increase: 75%", Confidence: 0.75},
			},
			context:      "Sales forecast",
			expectError:  false,
			minSynergies: 1,
		},
		{
			name: "perspective + temporal synergy",
			inputs: []*Input{
				{ID: "5", Mode: "perspective", Content: "Stakeholder A wants speed, Stakeholder B wants quality", Confidence: 0.8},
				{ID: "6", Mode: "temporal", Content: "Speed provides short-term wins but quality matters long-term", Confidence: 0.7},
			},
			context:      "Project planning",
			expectError:  false,
			minSynergies: 1,
		},
		{
			name: "perspective + causal synergy",
			inputs: []*Input{
				{ID: "7", Mode: "perspective", Content: "Users care about performance", Confidence: 0.9},
				{ID: "8", Mode: "causal", Content: "Database optimization improves performance", Confidence: 0.85},
			},
			context:      "System design",
			expectError:  false,
			minSynergies: 1,
		},
		{
			name: "temporal conflicts",
			inputs: []*Input{
				{ID: "9", Mode: "temporal", Content: "Short-term vs long-term tradeoff between cost and quality", Confidence: 0.7},
				{ID: "10", Mode: "causal", Content: "Quality reduces maintenance costs", Confidence: 0.8},
			},
			context:      "Budget decision",
			expectError:  false,
			minConflicts: 1,
		},
		{
			name: "perspective conflicts",
			inputs: []*Input{
				{ID: "11", Mode: "perspective", Content: "Engineering wants stability, Marketing wants new features - conflicting priorities", Confidence: 0.8},
				{ID: "12", Mode: "temporal", Content: "Features provide short-term value, stability is long-term", Confidence: 0.75},
			},
			context:      "Roadmap planning",
			expectError:  false,
			minConflicts: 1,
		},
		{
			name: "confidence variation conflict",
			inputs: []*Input{
				{ID: "13", Mode: "causal", Content: "High certainty causal relationship", Confidence: 0.95},
				{ID: "14", Mode: "probabilistic", Content: "Very uncertain probability estimate", Confidence: 0.3},
			},
			context:      "Risk assessment",
			expectError:  false,
			minConflicts: 1,
		},
		{
			name: "low confidence consensus",
			inputs: []*Input{
				{ID: "15", Mode: "causal", Content: "Unclear causal mechanism", Confidence: 0.4},
				{ID: "16", Mode: "probabilistic", Content: "Low probability estimate", Confidence: 0.35},
			},
			context:      "Uncertain scenario",
			expectError:  false,
			minConflicts: 1,
		},
		{
			name: "multiple modes with generic synergy",
			inputs: []*Input{
				{ID: "17", Mode: "mode1", Content: "Insight from mode 1", Confidence: 0.8},
				{ID: "18", Mode: "mode2", Content: "Insight from mode 2", Confidence: 0.75},
			},
			context:      "Generic analysis",
			expectError:  false,
			minSynergies: 1, // Should get generic synergy
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synthesis, err := s.SynthesizeInsights(tt.inputs, tt.context)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("SynthesizeInsights() failed: %v", err)
			}

			// Verify synthesis structure
			if synthesis.ID == "" {
				t.Error("Synthesis ID is empty")
			}
			if len(synthesis.Sources) != len(tt.inputs) {
				t.Errorf("Expected %d sources, got %d", len(tt.inputs), len(synthesis.Sources))
			}
			if synthesis.IntegratedView == "" {
				t.Error("IntegratedView is empty")
			}
			if len(synthesis.Synergies) < tt.minSynergies {
				t.Errorf("Expected at least %d synergies, got %d", tt.minSynergies, len(synthesis.Synergies))
			}
			if len(synthesis.Conflicts) < tt.minConflicts {
				t.Errorf("Expected at least %d conflicts, got %d", tt.minConflicts, len(synthesis.Conflicts))
			}
			if synthesis.Confidence < 0 || synthesis.Confidence > 1 {
				t.Errorf("Confidence %f out of range [0,1]", synthesis.Confidence)
			}
			if synthesis.CreatedAt.IsZero() {
				t.Error("CreatedAt is zero")
			}

			// Verify metadata
			if synthesis.Metadata == nil {
				t.Error("Metadata is nil")
			}
			if _, ok := synthesis.Metadata["modes"]; !ok {
				t.Error("Metadata missing 'modes' field")
			}
			if synthesis.Metadata["context"] != tt.context {
				t.Errorf("Expected context %q, got %q", tt.context, synthesis.Metadata["context"])
			}
		})
	}
}

func TestDetectSynergies(t *testing.T) {
	s := NewSynthesizer()

	tests := []struct {
		name             string
		inputs           []*Input
		expectedKeywords []string // Keywords that should appear in synergies
	}{
		{
			name: "causal + temporal",
			inputs: []*Input{
				{Mode: "causal", Content: "A causes B"},
				{Mode: "temporal", Content: "Effects over time"},
			},
			expectedKeywords: []string{"causal", "temporal"},
		},
		{
			name: "causal + probabilistic",
			inputs: []*Input{
				{Mode: "causal", Content: "Causal structure"},
				{Mode: "probabilistic", Content: "Probability distribution"},
			},
			expectedKeywords: []string{"causal", "probabilistic"},
		},
		{
			name: "perspective + temporal",
			inputs: []*Input{
				{Mode: "perspective", Content: "Stakeholder views"},
				{Mode: "temporal", Content: "Time horizons"},
			},
			expectedKeywords: []string{"stakeholder", "temporal"},
		},
		{
			name: "perspective + causal",
			inputs: []*Input{
				{Mode: "perspective", Content: "What matters to stakeholders"},
				{Mode: "causal", Content: "What can be changed"},
			},
			expectedKeywords: []string{"perspective", "causal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synergies := s.detectSynergies(tt.inputs, "test context")

			if len(synergies) == 0 {
				t.Error("No synergies detected")
			}

			// Check for expected keywords
			synergiesText := strings.ToLower(strings.Join(synergies, " "))
			for _, keyword := range tt.expectedKeywords {
				if !strings.Contains(synergiesText, strings.ToLower(keyword)) {
					t.Errorf("Expected synergy to mention %q, synergies: %v", keyword, synergies)
				}
			}
		})
	}
}

func TestDetectConflicts(t *testing.T) {
	s := NewSynthesizer()

	tests := []struct {
		name             string
		inputs           []*Input
		expectConflicts  bool
		expectedKeywords []string
	}{
		{
			name: "temporal tradeoff",
			inputs: []*Input{
				{Mode: "temporal", Content: "Short-term vs long-term tradeoff exists", Confidence: 0.8},
			},
			expectConflicts:  true,
			expectedKeywords: []string{"tradeoff"},
		},
		{
			name: "perspective conflict",
			inputs: []*Input{
				{Mode: "perspective", Content: "Stakeholders have conflicting priorities", Confidence: 0.8},
			},
			expectConflicts:  true,
			expectedKeywords: []string{"conflict"},
		},
		{
			name: "confidence variation",
			inputs: []*Input{
				{Mode: "mode1", Content: "High confidence analysis", Confidence: 0.95},
				{Mode: "mode2", Content: "Low confidence analysis", Confidence: 0.35},
			},
			expectConflicts:  true,
			expectedKeywords: []string{"confidence", "variation"},
		},
		{
			name: "low confidence consensus",
			inputs: []*Input{
				{Mode: "mode1", Content: "Uncertain", Confidence: 0.3},
				{Mode: "mode2", Content: "Unclear", Confidence: 0.4},
			},
			expectConflicts:  true,
			expectedKeywords: []string{"confidence", "uncertainty"},
		},
		{
			name: "no conflicts",
			inputs: []*Input{
				{Mode: "mode1", Content: "Clear analysis", Confidence: 0.8},
				{Mode: "mode2", Content: "Clear analysis", Confidence: 0.85},
			},
			expectConflicts: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflicts := s.detectConflicts(tt.inputs, "test context")

			if tt.expectConflicts {
				if len(conflicts) == 0 {
					t.Error("Expected conflicts but found none")
				}

				// Check for expected keywords
				conflictsText := strings.ToLower(strings.Join(conflicts, " "))
				for _, keyword := range tt.expectedKeywords {
					if !strings.Contains(conflictsText, strings.ToLower(keyword)) {
						t.Errorf("Expected conflict to mention %q, conflicts: %v", keyword, conflicts)
					}
				}
			} else {
				if len(conflicts) > 0 {
					t.Errorf("Expected no conflicts but found: %v", conflicts)
				}
			}
		})
	}
}

func TestGenerateIntegratedView(t *testing.T) {
	s := NewSynthesizer()

	inputs := []*Input{
		{ID: "1", Mode: "causal", Content: "A causes B", Confidence: 0.8},
		{ID: "2", Mode: "temporal", Content: "Effects unfold over time", Confidence: 0.75},
	}
	synergies := []string{"Modes complement each other"}
	conflicts := []string{"Some tension exists"}
	context := "Test analysis"

	view := s.generateIntegratedView(inputs, context, synergies, conflicts)

	if view == "" {
		t.Fatal("Generated view is empty")
	}

	// Check for required sections
	requiredSections := []string{
		"Test analysis",
		"Key Insights",
		"Complementary Insights",
		"Tensions to Address",
		"Synthesized Conclusion",
	}

	for _, section := range requiredSections {
		if !strings.Contains(view, section) {
			t.Errorf("Integrated view missing section: %s", section)
		}
	}

	// Verify inputs are mentioned
	for _, input := range inputs {
		if !strings.Contains(view, input.Mode) {
			t.Errorf("Integrated view missing mode: %s", input.Mode)
		}
	}
}

func TestCalculateSynthesisConfidence(t *testing.T) {
	s := NewSynthesizer()

	tests := []struct {
		name              string
		inputs            []*Input
		synergies         []string
		conflicts         []string
		expectedMinConf   float64
		expectedMaxConf   float64
	}{
		{
			name: "high confidence with synergies",
			inputs: []*Input{
				{Confidence: 0.9},
				{Confidence: 0.85},
			},
			synergies:       []string{"synergy1", "synergy2"},
			conflicts:       []string{},
			expectedMinConf: 0.85,
			expectedMaxConf: 1.0,
		},
		{
			name: "moderate confidence with conflicts",
			inputs: []*Input{
				{Confidence: 0.6},
				{Confidence: 0.65},
			},
			synergies:       []string{},
			conflicts:       []string{"conflict1", "conflict2"},
			expectedMinConf: 0.4,
			expectedMaxConf: 0.7,
		},
		{
			name: "low confidence",
			inputs: []*Input{
				{Confidence: 0.3},
				{Confidence: 0.4},
			},
			synergies:       []string{},
			conflicts:       []string{},
			expectedMinConf: 0.2,
			expectedMaxConf: 0.5,
		},
		{
			name: "multiple modes bonus",
			inputs: []*Input{
				{Confidence: 0.7},
				{Confidence: 0.7},
				{Confidence: 0.7},
				{Confidence: 0.7},
			},
			synergies:       []string{},
			conflicts:       []string{},
			expectedMinConf: 0.7,
			expectedMaxConf: 0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := s.calculateSynthesisConfidence(tt.inputs, tt.synergies, tt.conflicts)

			if confidence < tt.expectedMinConf || confidence > tt.expectedMaxConf {
				t.Errorf("Expected confidence between %.2f and %.2f, got %.2f",
					tt.expectedMinConf, tt.expectedMaxConf, confidence)
			}

			// Confidence must be in valid range
			if confidence < 0 || confidence > 1 {
				t.Errorf("Confidence %f out of range [0,1]", confidence)
			}
		})
	}
}

func TestDetectEmergentPatterns(t *testing.T) {
	s := NewSynthesizer()

	tests := []struct {
		name             string
		inputs           []*Input
		expectError      bool
		expectedKeywords []string
	}{
		{
			name:        "insufficient inputs",
			inputs:      []*Input{{Mode: "test", Content: "test"}},
			expectError: true,
		},
		{
			name: "feedback loop pattern",
			inputs: []*Input{
				{Mode: "causal", Content: "Reinforcing feedback cycle between A and B"},
				{Mode: "temporal", Content: "Effects compound over time"},
			},
			expectError:      false,
			expectedKeywords: []string{"feedback", "loop"},
		},
		{
			name: "incentive misalignment",
			inputs: []*Input{
				{Mode: "perspective", Content: "Different stakeholder views"},
				{Mode: "causal", Content: "Incentives drive conflicting behavior"},
			},
			expectError:      false,
			expectedKeywords: []string{"incentive", "misalignment"},
		},
		{
			name: "delayed consequences",
			inputs: []*Input{
				{Mode: "causal", Content: "A causes B"},
				{Mode: "temporal", Content: "Long-term delayed impact"},
			},
			expectError:      false,
			expectedKeywords: []string{"delay", "impact"},
		},
		{
			name: "uncertainty cascade",
			inputs: []*Input{
				{Mode: "probabilistic", Content: "Probability estimate", Confidence: 0.4},
				{Mode: "causal", Content: "Causal chain", Confidence: 0.5},
			},
			expectError:      false,
			expectedKeywords: []string{"uncertainty", "propagation"},
		},
		{
			name: "generic pattern",
			inputs: []*Input{
				{Mode: "mode1", Content: "Analysis 1"},
				{Mode: "mode2", Content: "Analysis 2"},
			},
			expectError:      false,
			expectedKeywords: []string{"cross-mode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns, err := s.DetectEmergentPatterns(tt.inputs)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("DetectEmergentPatterns() failed: %v", err)
			}

			if len(patterns) == 0 {
				t.Error("No patterns detected")
			}

			// Check for expected keywords
			patternsText := strings.ToLower(strings.Join(patterns, " "))
			for _, keyword := range tt.expectedKeywords {
				if !strings.Contains(patternsText, strings.ToLower(keyword)) {
					t.Errorf("Expected pattern to mention %q, patterns: %v", keyword, patterns)
				}
			}
		})
	}
}

func TestGetSynthesis(t *testing.T) {
	s := NewSynthesizer()

	// Create a synthesis
	inputs := []*Input{
		{ID: "1", Mode: "causal", Content: "Test", Confidence: 0.8},
		{ID: "2", Mode: "temporal", Content: "Test", Confidence: 0.75},
	}
	synthesis, err := s.SynthesizeInsights(inputs, "test")
	if err != nil {
		t.Fatalf("SynthesizeInsights() failed: %v", err)
	}

	// Test retrieval
	retrieved, err := s.GetSynthesis(synthesis.ID)
	if err != nil {
		t.Fatalf("GetSynthesis() failed: %v", err)
	}

	if retrieved.ID != synthesis.ID {
		t.Errorf("Expected synthesis ID %q, got %q", synthesis.ID, retrieved.ID)
	}

	// Test nonexistent synthesis
	_, err = s.GetSynthesis("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent synthesis, got nil")
	}
}

func TestConcurrentSynthesis(t *testing.T) {
	s := NewSynthesizer()

	inputSets := [][]*Input{
		{
			{ID: "1a", Mode: "causal", Content: "Test A", Confidence: 0.8},
			{ID: "1b", Mode: "temporal", Content: "Test A", Confidence: 0.75},
		},
		{
			{ID: "2a", Mode: "perspective", Content: "Test B", Confidence: 0.85},
			{ID: "2b", Mode: "probabilistic", Content: "Test B", Confidence: 0.7},
		},
		{
			{ID: "3a", Mode: "causal", Content: "Test C", Confidence: 0.9},
			{ID: "3b", Mode: "temporal", Content: "Test C", Confidence: 0.8},
		},
	}

	done := make(chan bool)
	errors := make(chan error, len(inputSets))

	for i, inputs := range inputSets {
		go func(idx int, ins []*Input) {
			_, err := s.SynthesizeInsights(ins, "concurrent test")
			if err != nil {
				errors <- err
			}
			done <- true
		}(i, inputs)
	}

	// Wait for all to complete
	for i := 0; i < len(inputSets); i++ {
		<-done
	}

	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent synthesis error: %v", err)
	}
}

// Test helper functions

func TestExtractKey(t *testing.T) {
	tests := []struct {
		text   string
		maxLen int
		expect string
	}{
		{"Short text", 20, "Short text"},
		{"This is a longer text that needs truncation", 20, "This is a longer ..."},
		{"First sentence. Second sentence.", 25, "First sentence"},
		{"No sentence break here but very long text", 15, "No sentence bre..."},
	}

	for _, tt := range tests {
		result := extractKey(tt.text, tt.maxLen)
		if len(result) > tt.maxLen+3 { // +3 for "..."
			t.Errorf("extractKey(%q, %d) = %q (too long)", tt.text, tt.maxLen, result)
		}
	}
}

func TestFindMinMax(t *testing.T) {
	tests := []struct {
		values     []float64
		expectedMin float64
		expectedMax float64
	}{
		{[]float64{1.0, 2.0, 3.0}, 1.0, 3.0},
		{[]float64{5.5, 2.2, 8.8, 1.1}, 1.1, 8.8},
		{[]float64{7.0}, 7.0, 7.0},
		{[]float64{}, 0.0, 0.0},
	}

	for _, tt := range tests {
		min, max := findMinMax(tt.values)
		if min != tt.expectedMin || max != tt.expectedMax {
			t.Errorf("findMinMax(%v) = (%f, %f), want (%f, %f)",
				tt.values, min, max, tt.expectedMin, tt.expectedMax)
		}
	}
}

func TestAverage(t *testing.T) {
	tests := []struct {
		values   []float64
		expected float64
	}{
		{[]float64{1.0, 2.0, 3.0}, 2.0},
		{[]float64{5.0, 10.0}, 7.5},
		{[]float64{7.5}, 7.5},
		{[]float64{}, 0.0},
	}

	for _, tt := range tests {
		result := average(tt.values)
		if result != tt.expected {
			t.Errorf("average(%v) = %f, want %f", tt.values, result, tt.expected)
		}
	}
}
