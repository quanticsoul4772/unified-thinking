package validation

import (
	"testing"
)

func TestFallacyDetector_DetectFormalFallacies(t *testing.T) {
	fd := NewFallacyDetector()

	tests := []struct {
		name          string
		content       string
		expectedType  string
		shouldDetect  bool
	}{
		{
			name:         "affirming consequent",
			content:      "If it rains, the ground is wet. The ground is wet. Therefore, it rained.",
			expectedType: "affirming_consequent",
			shouldDetect: true,
		},
		{
			name:         "denying antecedent",
			content:      "If it rains, the ground is wet. It's not raining. Therefore, the ground is not wet.",
			expectedType: "denying_antecedent",
			shouldDetect: true,
		},
		{
			name:         "undistributed middle",
			content:      "All cats are mammals. All dogs are mammals. Therefore, all cats are dogs.",
			expectedType: "undistributed_middle",
			shouldDetect: true,
		},
		{
			name:         "valid argument",
			content:      "All humans are mortal. Socrates is human. Therefore, Socrates is mortal.",
			expectedType: "",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallacies := fd.DetectFallacies(tt.content, true, false)

			detected := false
			for _, f := range fallacies {
				if f.Type == tt.expectedType {
					detected = true
					break
				}
			}

			if detected != tt.shouldDetect {
				t.Errorf("Detection mismatch for %q: got %v, want %v", tt.name, detected, tt.shouldDetect)
			}
		})
	}
}

func TestFallacyDetector_DetectInformalFallacies(t *testing.T) {
	fd := NewFallacyDetector()

	tests := []struct {
		name          string
		content       string
		expectedType  string
		shouldDetect  bool
	}{
		{
			name:         "ad hominem",
			content:      "Your argument is wrong because you're an idiot who doesn't understand logic.",
			expectedType: "ad_hominem",
			shouldDetect: true,
		},
		{
			name:         "straw man",
			content:      "They claim we need education reform, but they really want to destroy our schools.",
			expectedType: "straw_man",
			shouldDetect: true,
		},
		{
			name:         "false dilemma",
			content:      "Either you support this policy or you hate America. There are no other options.",
			expectedType: "false_dilemma",
			shouldDetect: true,
		},
		{
			name:         "slippery slope",
			content:      "If we allow this, it will lead to chaos, which leads to anarchy, which results in societal collapse.",
			expectedType: "slippery_slope",
			shouldDetect: true,
		},
		{
			name:         "appeal to emotion",
			content:      "Think about the children! Imagine how they would feel! This is heartbreaking!",
			expectedType: "appeal_to_emotion",
			shouldDetect: true,
		},
		{
			name:         "hasty generalization",
			content:      "I met one rude French person once, so all French people are rude.",
			expectedType: "hasty_generalization",
			shouldDetect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallacies := fd.DetectFallacies(tt.content, false, true)

			detected := false
			for _, f := range fallacies {
				if f.Type == tt.expectedType {
					detected = true
					if f.Category != FallacyInformal {
						t.Errorf("Wrong category: got %v, want %v", f.Category, FallacyInformal)
					}
					break
				}
			}

			if detected != tt.shouldDetect {
				t.Errorf("Detection mismatch for %q: got %v, want %v", tt.name, detected, tt.shouldDetect)
			}
		})
	}
}

func TestFallacyDetector_DetectStatisticalFallacies(t *testing.T) {
	fd := NewFallacyDetector()

	tests := []struct {
		name          string
		content       string
		expectedType  string
		shouldDetect  bool
	}{
		{
			name:         "post hoc - lucky socks",
			content:      "After I wore my lucky socks, we won the game, so the socks caused the win.",
			expectedType: "post_hoc_ergo_propter_hoc",
			shouldDetect: true,
		},
		{
			name:         "post hoc - policy",
			content:      "Since implementing the new policy, crime rates dropped, therefore the policy worked.",
			expectedType: "post_hoc_ergo_propter_hoc",
			shouldDetect: true,
		},
		{
			name:         "post hoc - correlation causation",
			content:      "Ice cream sales and crime rates are correlated, so ice cream causes crime.",
			expectedType: "post_hoc_ergo_propter_hoc",
			shouldDetect: true,
		},
		{
			name:         "post hoc - temporal sequence",
			content:      "Following the CEO's arrival, profits increased, thus the CEO led to success.",
			expectedType: "post_hoc_ergo_propter_hoc",
			shouldDetect: true,
		},
		{
			name:         "valid causal claim with mechanism",
			content:      "The medication blocks the enzyme responsible for inflammation, which reduces pain.",
			expectedType: "post_hoc_ergo_propter_hoc",
			shouldDetect: false,  // This is valid - has causal mechanism
		},
		{
			name:         "base rate neglect",
			content:      "The test is 99% accurate, so if you test positive, you definitely have the disease.",
			expectedType: "base_rate_neglect",
			shouldDetect: true,
		},
		{
			name:         "texas sharpshooter",
			content:      "I found a significant pattern in the data that proves my theory.",
			expectedType: "texas_sharpshooter",
			shouldDetect: true,
		},
		{
			name:         "survivorship bias",
			content:      "All successful entrepreneurs dropped out of college, so dropping out leads to success.",
			expectedType: "survivorship_bias",
			shouldDetect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallacies := fd.DetectFallacies(tt.content, false, true)

			detected := false
			for _, f := range fallacies {
				if f.Type == tt.expectedType {
					detected = true
					if f.Category != FallacyStatistical {
						t.Errorf("Wrong category: got %v, want %v", f.Category, FallacyStatistical)
					}
					break
				}
			}

			if detected != tt.shouldDetect {
				t.Errorf("Detection mismatch for %q: got %v, want %v", tt.name, detected, tt.shouldDetect)
			}
		})
	}
}

func TestFallacyDetector_CreateFallacyValidation(t *testing.T) {
	fd := NewFallacyDetector()

	// No fallacies case
	validation := fd.CreateFallacyValidation([]*DetectedFallacy{}, "thought_123")
	if !validation.IsValid {
		t.Errorf("Validation should be valid when no fallacies detected")
	}

	// With fallacies case
	fallacies := []*DetectedFallacy{
		{
			Type:        "ad_hominem",
			Category:    FallacyInformal,
			Explanation: "Personal attack detected",
		},
		{
			Type:        "false_dilemma",
			Category:    FallacyInformal,
			Explanation: "Only two options presented",
		},
	}

	validation = fd.CreateFallacyValidation(fallacies, "thought_123")
	if validation.IsValid {
		t.Errorf("Validation should be invalid when fallacies detected")
	}

	if validation.ThoughtID != "thought_123" {
		t.Errorf("ThoughtID mismatch: got %s, want %s", validation.ThoughtID, "thought_123")
	}

	// Check validation data contains fallacies
	if validation.ValidationData == nil {
		t.Errorf("ValidationData should not be nil")
	} else if _, ok := validation.ValidationData["fallacies"]; !ok {
		t.Errorf("ValidationData should contain fallacies key")
	}
}

func TestFallacyDetector_ExtractExample(t *testing.T) {
	fd := NewFallacyDetector()

	content := "This is a long text with many words. If we find the keyword, we should extract context around it for the example."
	keywords := []string{"keyword"}

	example := fd.extractExample(content, keywords)

	if len(example) == 0 {
		t.Errorf("Expected non-empty example")
	}

	if example[0:3] != "..." {
		t.Errorf("Example should start with ellipsis")
	}
}

func TestFallacyDetector_SentenceSimilarity(t *testing.T) {
	fd := NewFallacyDetector()

	tests := []struct {
		s1              string
		s2              string
		minSimilarity   float64
	}{
		{
			s1:            "the quick brown fox jumps",
			s2:            "the quick brown fox leaps",
			minSimilarity: 0.6,
		},
		{
			s1:            "completely different sentence",
			s2:            "another unrelated phrase",
			minSimilarity: 0.0,
		},
		{
			s1:            "this is a test",
			s2:            "this is a test",
			minSimilarity: 0.9,
		},
	}

	for _, tt := range tests {
		sim := fd.sentenceSimilarity(tt.s1, tt.s2)
		if sim < tt.minSimilarity {
			t.Errorf("Similarity between %q and %q is %.3f, expected at least %.3f",
				tt.s1, tt.s2, sim, tt.minSimilarity)
		}
	}
}
