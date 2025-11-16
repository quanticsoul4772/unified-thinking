package reasoning

import (
	"testing"
)

func TestNewTemporalReasoner(t *testing.T) {
	tr := NewTemporalReasoner()
	if tr == nil {
		t.Fatal("NewTemporalReasoner() returned nil")
	}
}

func TestAnalyzeTemporal(t *testing.T) {
	tr := NewTemporalReasoner()

	tests := []struct {
		name            string
		situation       string
		timeHorizon     string
		expectError     bool
		expectedHorizon string
	}{
		{
			name:        "empty situation",
			situation:   "",
			expectError: true,
		},
		{
			name:            "short-term horizon",
			situation:       "Should we implement this quick fix?",
			timeHorizon:     "days",
			expectError:     false,
			expectedHorizon: "days-weeks",
		},
		{
			name:            "medium-term horizon",
			situation:       "Planning a major feature rollout",
			timeHorizon:     "months",
			expectError:     false,
			expectedHorizon: "months",
		},
		{
			name:            "long-term horizon",
			situation:       "Strategic technology investment decision",
			timeHorizon:     "years",
			expectError:     false,
			expectedHorizon: "years",
		},
		{
			name:            "no horizon specified",
			situation:       "General decision to analyze",
			timeHorizon:     "",
			expectError:     false,
			expectedHorizon: "months", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := tr.AnalyzeTemporal(tt.situation, tt.timeHorizon)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("AnalyzeTemporal() failed: %v", err)
			}

			// Verify analysis structure
			if analysis.ID == "" {
				t.Error("Analysis ID is empty")
			}
			if analysis.ShortTermView == "" {
				t.Error("ShortTermView is empty")
			}
			if analysis.LongTermView == "" {
				t.Error("LongTermView is empty")
			}
			if analysis.TimeHorizon != tt.expectedHorizon {
				t.Errorf("Expected horizon '%s', got '%s'", tt.expectedHorizon, analysis.TimeHorizon)
			}
			if len(analysis.Tradeoffs) == 0 {
				t.Error("Tradeoffs are empty")
			}
			if analysis.Recommendation == "" {
				t.Error("Recommendation is empty")
			}
			if analysis.CreatedAt.IsZero() {
				t.Error("CreatedAt is zero")
			}
		})
	}
}

func TestNormalizeTimeHorizon(t *testing.T) {
	tr := NewTemporalReasoner()

	tests := []struct {
		input    string
		expected string
	}{
		{"days", "days-weeks"},
		{"weeks", "days-weeks"},
		{"immediate", "days-weeks"},
		{"short term", "days-weeks"},
		{"months", "months"},
		{"quarter", "months"},
		{"medium term", "months"},
		{"years", "years"},
		{"long term", "years"},
		{"strategic", "years"},
		{"decades", "decades"},
		{"generation", "decades"},
		{"", "months"},        // Default
		{"unknown", "months"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := tr.normalizeTimeHorizon(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeTimeHorizon(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAnalyzeShortTerm(t *testing.T) {
	tr := NewTemporalReasoner()

	tests := []struct {
		name      string
		situation string
	}{
		{
			name:      "cost implications",
			situation: "This will require significant upfront cost and budget allocation",
		},
		{
			name:      "implementation effort",
			situation: "We need to implement and deploy this new system quickly",
		},
		{
			name:      "user impact",
			situation: "This change will affect all users and customers immediately",
		},
		{
			name:      "operational impact",
			situation: "Our operational processes and workflows will need adjustment",
		},
		{
			name:      "risk focus",
			situation: "There are significant risks and challenges we need to address",
		},
		{
			name:      "generic situation",
			situation: "A general business decision",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := tr.analyzeShortTerm(tt.situation)

			if view == "" {
				t.Error("analyzeShortTerm() returned empty string")
			}

			if len(view) < 20 {
				t.Errorf("Short-term view seems too brief: %q", view)
			}
		})
	}
}

func TestAnalyzeLongTerm(t *testing.T) {
	tr := NewTemporalReasoner()

	tests := []struct {
		name      string
		situation string
		horizon   string
	}{
		{
			name:      "scalability focus",
			situation: "This needs to scale with our growth and expansion",
			horizon:   "years",
		},
		{
			name:      "sustainability focus",
			situation: "We need a sustainable long-term solution",
			horizon:   "years",
		},
		{
			name:      "strategic focus",
			situation: "This aligns with our long-term strategy and future vision",
			horizon:   "years",
		},
		{
			name:      "cultural impact",
			situation: "This will shape our organizational culture and team dynamics",
			horizon:   "months",
		},
		{
			name:      "technical debt",
			situation: "We need to manage technical debt and legacy maintenance",
			horizon:   "years",
		},
		{
			name:      "generic situation",
			situation: "A general decision",
			horizon:   "months",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := tr.analyzeLongTerm(tt.situation, tt.horizon)

			if view == "" {
				t.Error("analyzeLongTerm() returned empty string")
			}

			if len(view) < 20 {
				t.Errorf("Long-term view seems too brief: %q", view)
			}
		})
	}
}

func TestIdentifyTradeoffs(t *testing.T) {
	tr := NewTemporalReasoner()

	tests := []struct {
		name         string
		situation    string
		minTradeoffs int
	}{
		{
			name:         "speed tradeoff",
			situation:    "We need rapid deployment but also careful planning",
			minTradeoffs: 1,
		},
		{
			name:         "cost tradeoff",
			situation:    "Lower initial cost vs higher long-term expenses",
			minTradeoffs: 1,
		},
		{
			name:         "simplicity tradeoff",
			situation:    "Simple solution now vs comprehensive system later",
			minTradeoffs: 1,
		},
		{
			name:         "automation tradeoff",
			situation:    "Manual process to start, automation for the future",
			minTradeoffs: 1,
		},
		{
			name:         "no specific tradeoffs",
			situation:    "A general decision without clear temporal tensions",
			minTradeoffs: 1, // Should generate generic tradeoffs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortTerm := "Short-term view"
			longTerm := "Long-term view"

			tradeoffs := tr.identifyTradeoffs(tt.situation, shortTerm, longTerm)

			if len(tradeoffs) < tt.minTradeoffs {
				t.Errorf("Expected at least %d tradeoffs, got %d", tt.minTradeoffs, len(tradeoffs))
			}

			for _, tradeoff := range tradeoffs {
				if tradeoff == "" {
					t.Error("Empty tradeoff found")
				}
			}
		})
	}
}

func TestGenerateRecommendation(t *testing.T) {
	tr := NewTemporalReasoner()

	tests := []struct {
		name      string
		shortTerm string
		longTerm  string
		horizon   string
	}{
		{
			name:      "strong short-term focus",
			shortTerm: "Immediate impact; quick wins; rapid results; urgent needs; current challenges",
			longTerm:  "Future considerations",
			horizon:   "days-weeks",
		},
		{
			name:      "strong long-term focus",
			shortTerm: "Initial setup",
			longTerm:  "Strategic alignment; scalability; sustainability; cultural impact; ecosystem effects; technical foundation",
			horizon:   "years",
		},
		{
			name:      "balanced approach",
			shortTerm: "Near-term costs; implementation effort",
			longTerm:  "Long-term benefits; strategic value",
			horizon:   "months",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tradeoffs := []string{"Tradeoff 1", "Tradeoff 2"}
			recommendation := tr.generateRecommendation(tt.shortTerm, tt.longTerm, tradeoffs, tt.horizon)

			if recommendation == "" {
				t.Error("generateRecommendation() returned empty string")
			}

			if len(recommendation) < 50 {
				t.Errorf("Recommendation seems too brief: %q", recommendation)
			}

			// Verify it contains "Recommendation:"
			if !contains(recommendation, "Recommendation:") {
				t.Error("Recommendation should start with 'Recommendation:'")
			}
		})
	}
}

func TestCompareTimeHorizons(t *testing.T) {
	tr := NewTemporalReasoner()

	situation := "Should we invest in this new technology platform?"

	analyses, err := tr.CompareTimeHorizons(situation)
	if err != nil {
		t.Fatalf("CompareTimeHorizons() failed: %v", err)
	}

	expectedHorizons := []string{"days-weeks", "months", "years"}
	for _, horizon := range expectedHorizons {
		analysis, ok := analyses[horizon]
		if !ok {
			t.Errorf("Missing analysis for horizon: %s", horizon)
			continue
		}

		if analysis.TimeHorizon != horizon {
			t.Errorf("Expected horizon %s, got %s", horizon, analysis.TimeHorizon)
		}

		if analysis.ShortTermView == "" {
			t.Error("ShortTermView is empty")
		}
		if analysis.LongTermView == "" {
			t.Error("LongTermView is empty")
		}
	}
}

func TestCompareTimeHorizons_EmptySituation(t *testing.T) {
	tr := NewTemporalReasoner()

	_, err := tr.CompareTimeHorizons("")
	if err == nil {
		t.Error("Expected error for empty situation, got nil")
	}
}

func TestIdentifyOptimalTiming(t *testing.T) {
	tr := NewTemporalReasoner()

	tests := []struct {
		name            string
		situation       string
		constraints     []string
		expectError     bool
		expectedKeyword string // Word that should appear in recommendation
	}{
		{
			name:        "empty situation",
			situation:   "",
			expectError: true,
		},
		{
			name:            "urgent situation",
			situation:       "Critical security issue requires immediate action",
			expectError:     false,
			expectedKeyword: "immediately",
		},
		{
			name:            "deadline constraint",
			situation:       "Planning for upcoming initiative",
			constraints:     []string{"Must complete by end of quarter", "Deadline: June 30"},
			expectError:     false,
			expectedKeyword: "deadline",
		},
		{
			name:            "strategic timing",
			situation:       "Long-term strategic investment in future technology",
			expectError:     false,
			expectedKeyword: "strategic",
		},
		{
			name:            "flexible timing",
			situation:       "General improvement initiative",
			expectError:     false,
			expectedKeyword: "flexible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendation, err := tr.IdentifyOptimalTiming(tt.situation, tt.constraints)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("IdentifyOptimalTiming() failed: %v", err)
			}

			if recommendation == "" {
				t.Error("Recommendation is empty")
			}

			if tt.expectedKeyword != "" && !containsIgnoreCase(recommendation, tt.expectedKeyword) {
				t.Errorf("Expected recommendation to contain '%s', got: %q", tt.expectedKeyword, recommendation)
			}
		})
	}
}

func TestConcurrentTemporalAnalysis(t *testing.T) {
	tr := NewTemporalReasoner()

	situations := []string{
		"Decision 1: Urgent technical issue",
		"Decision 2: Strategic investment",
		"Decision 3: Operational improvement",
		"Decision 4: Customer feature request",
		"Decision 5: Long-term platform migration",
	}

	done := make(chan bool)
	errors := make(chan error, len(situations))

	for _, situation := range situations {
		go func(s string) {
			_, err := tr.AnalyzeTemporal(s, "months")
			if err != nil {
				errors <- err
			}
			done <- true
		}(situation)
	}

	// Wait for all to complete
	for i := 0; i < len(situations); i++ {
		<-done
	}

	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent analysis error: %v", err)
	}
}

// Helper function (containsIgnoreCase uses toLower and contains from causal_test.go)
func containsIgnoreCase(s, substr string) bool {
	return contains(toLower(s), toLower(substr))
}
