# Test Cases - Quick Reference Guide

**Quick access to concrete, implementable test cases for immediate use**

---

## Priority 0 (Critical) - Implement First

### 1. Argument Analysis Tests (0% coverage currently)

**File:** `internal/analysis/argument_test.go`

```go
package analysis

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDecomposeArgument_SimpleDeductive(t *testing.T) {
	aa := NewArgumentAnalyzer()

	text := "All men are mortal. Socrates is a man. Therefore, Socrates is mortal."

	decomp, err := aa.DecomposeArgument(text)

	assert.NoError(t, err)
	assert.Equal(t, "Socrates is mortal", decomp.MainClaim)
	assert.Equal(t, ArgumentDeductive, decomp.ArgumentType)
	assert.Len(t, decomp.Premises, 2)
	assert.GreaterOrEqual(t, decomp.Strength, 0.9)
}

func TestDecomposeArgument_HiddenAssumptions(t *testing.T) {
	aa := NewArgumentAnalyzer()

	text := "We should ban guns because they cause violence."

	decomp, err := aa.DecomposeArgument(text)

	assert.NoError(t, err)
	assert.NotEmpty(t, decomp.HiddenAssumptions)
	assert.Contains(t, decomp.HiddenAssumptions[0], "causal")
}

func TestGenerateCounterArguments_DenyPremise(t *testing.T) {
	aa := NewArgumentAnalyzer()

	// First create an argument
	text := "Some people say X is true. My friend agrees. Therefore X is definitely true."
	decomp, _ := aa.DecomposeArgument(text)

	// Generate counter-arguments
	counters, err := aa.GenerateCounterArguments(decomp.ID)

	assert.NoError(t, err)
	assert.NotEmpty(t, counters)

	// Should have counter using deny_premise strategy
	foundDenyStrategy := false
	for _, counter := range counters {
		if counter.Strategy == "deny_premise" {
			foundDenyStrategy = true
		}
	}
	assert.True(t, foundDenyStrategy)
}

func TestExtractPremises_TableDriven(t *testing.T) {
	aa := NewArgumentAnalyzer()

	tests := []struct {
		name          string
		text          string
		expectedCount int
		expectedType  string
	}{
		{
			name:          "factual premises",
			text:          "The sky is blue. Water is wet.",
			expectedCount: 2,
			expectedType:  "factual",
		},
		{
			name:          "value premises",
			text:          "We should help others. Kindness ought to be valued.",
			expectedCount: 2,
			expectedType:  "value",
		},
		{
			name:          "mixed premises",
			text:          "Dogs are animals. We should treat animals well.",
			expectedCount: 2,
			expectedType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			premises := aa.extractPremises(tt.text)
			assert.Len(t, premises, tt.expectedCount)
			if tt.expectedType != "" {
				assert.Equal(t, tt.expectedType, premises[0].Type)
			}
		})
	}
}
```

**Expected Results:**
- Premise extraction recall: 90%+
- Hidden assumption detection: 60%+
- Strength estimates reasonable (0.0-1.0 range, correlates with intuition)

---

### 2. Probabilistic Reasoning Accuracy Tests

**File:** `internal/reasoning/probabilistic_accuracy_test.go`

```go
package reasoning

import (
	"math"
	"testing"
	"github.com/stretchr/testify/assert"
)

// Test Case PR-001: Medical Testing (Classic Bayes Problem)
func TestBayesUpdate_MedicalTesting(t *testing.T) {
	pr := NewProbabilisticReasoner()

	// Disease base rate: 1%
	belief, err := pr.CreateBelief("Patient has disease", 0.01)
	assert.NoError(t, err)

	// Test sensitivity: 95% (P(positive | disease))
	// Test false positive rate: 5% (P(positive | no disease))
	// Overall P(positive) = 0.95*0.01 + 0.05*0.99 = 0.059

	updated, err := pr.UpdateBelief(belief.ID, "test_positive", 0.95, 0.059)

	assert.NoError(t, err)

	// P(disease | positive) = 0.95 * 0.01 / 0.059 ≈ 0.161
	expected := 0.161
	assert.InDelta(t, expected, updated.Probability, 0.01, "Bayes calculation should be accurate to 0.01")
}

// Test Case PR-002: Sequential Updates
func TestBayesUpdate_Sequential(t *testing.T) {
	pr := NewProbabilisticReasoner()

	belief, _ := pr.CreateBelief("Hypothesis H", 0.5)

	// First update
	belief, _ = pr.UpdateBelief(belief.ID, "E1", 0.8, 0.6)
	prob1 := belief.Probability

	// Second update
	belief, _ = pr.UpdateBelief(belief.ID, "E2", 0.7, 0.5)
	prob2 := belief.Probability

	// If likelihood > prior, probability should increase
	assert.Greater(t, prob1, 0.5, "First update should increase probability")

	// Evidence should accumulate
	assert.Len(t, belief.Evidence, 2)
}

// Test Case PR-003: Probability Bounds
func TestProbabilityBounds(t *testing.T) {
	pr := NewProbabilisticReasoner()

	tests := []struct {
		name          string
		prior         float64
		likelihood    float64
		evidenceProb  float64
		shouldBeValid bool
	}{
		{"valid middle", 0.5, 0.7, 0.6, true},
		{"invalid prior", -0.1, 0.5, 0.5, false},
		{"invalid prior high", 1.5, 0.5, 0.5, false},
		{"zero prior", 0.0, 0.5, 0.5, true},
		{"one prior", 1.0, 0.5, 0.5, true},
		{"zero evidence prob", 0.5, 0.5, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			belief, err := pr.CreateBelief("test", tt.prior)
			if !tt.shouldBeValid && tt.prior < 0 || tt.prior > 1 {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			_, err = pr.UpdateBelief(belief.ID, "evidence", tt.likelihood, tt.evidenceProb)
			if tt.shouldBeValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Test Case PR-004: Combine Beliefs - AND
func TestCombineBeliefs_AND(t *testing.T) {
	pr := NewProbabilisticReasoner()

	b1, _ := pr.CreateBelief("Event A", 0.8)
	b2, _ := pr.CreateBelief("Event B", 0.9)

	combined, err := pr.CombineBeliefs([]string{b1.ID, b2.ID}, "and")

	assert.NoError(t, err)
	assert.InDelta(t, 0.72, combined, 0.001, "P(A and B) = P(A) * P(B) for independent events")
}

// Test Case PR-005: Combine Beliefs - OR
func TestCombineBeliefs_OR(t *testing.T) {
	pr := NewProbabilisticReasoner()

	b1, _ := pr.CreateBelief("Event A", 0.3)
	b2, _ := pr.CreateBelief("Event B", 0.4)

	combined, err := pr.CombineBeliefs([]string{b1.ID, b2.ID}, "or")

	assert.NoError(t, err)
	// P(A or B) = 1 - (1-0.3)*(1-0.4) = 1 - 0.42 = 0.58
	assert.InDelta(t, 0.58, combined, 0.001)
}

// Test Case PR-006: Monty Hall Problem
func TestBayesUpdate_MontyHall(t *testing.T) {
	pr := NewProbabilisticReasoner()

	// Initially, car is behind any door with P=1/3
	belief, _ := pr.CreateBelief("Car is behind door 1", 1.0/3.0)

	// Monty opens door 3 (revealing goat)
	// P(Monty opens 3 | car behind 1) = 0.5 (he could open 2 or 3)
	// P(Monty opens 3) = 0.5 * (1/3) + 1 * (1/3) + 0 * (1/3) = 0.5

	updated, _ := pr.UpdateBelief(belief.ID, "monty_opens_3", 0.5, 0.5)

	// Should stay at 1/3
	assert.InDelta(t, 1.0/3.0, updated.Probability, 0.01)

	// Door 2 probability should be 2/3 (but we'd need separate belief tracking)
}

// Benchmark Bayes calculation performance
func BenchmarkBayesUpdate(b *testing.B) {
	pr := NewProbabilisticReasoner()
	belief, _ := pr.CreateBelief("Hypothesis", 0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pr.UpdateBelief(belief.ID, "evidence", 0.7, 0.6)
	}
}
```

**Expected Results:**
- All Bayes calculations accurate to 0.01
- Handles edge cases (0, 1 probabilities)
- No probability values outside [0, 1]
- Monty Hall problem: Correct posteriors

---

### 3. Causal Reasoning Validation Tests

**File:** `internal/reasoning/causal_accuracy_test.go`

```go
package reasoning

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// Test Case CG-001: Smoking and Lung Cancer
func TestBuildCausalGraph_SmokingCancer(t *testing.T) {
	cr := NewCausalReasoner()

	observations := []string{
		"Smoking causes lung cancer",
		"Smoking increases tar exposure",
		"Lung cancer increases mortality risk",
	}

	graph, err := cr.BuildCausalGraph("Smoking health effects", observations)

	assert.NoError(t, err)
	assert.NotNil(t, graph)

	// Should extract key variables
	variableNames := extractVariableNames(graph.Variables)
	assert.Contains(t, variableNames, "smoking")
	assert.Contains(t, variableNames, "lung cancer")
	assert.Contains(t, variableNames, "mortality")

	// Should identify causal links
	assert.GreaterOrEqual(t, len(graph.Links), 2)

	// Smoking → Lung cancer link should exist
	foundSmokingToCancer := false
	for _, link := range graph.Links {
		fromVar := findVariable(graph.Variables, link.From)
		toVar := findVariable(graph.Variables, link.To)
		if fromVar != nil && toVar != nil &&
			fromVar.Name == "smoking" && toVar.Name == "lung cancer" {
			foundSmokingToCancer = true
			assert.Equal(t, "positive", link.Type)
		}
	}
	assert.True(t, foundSmokingToCancer, "Should identify smoking → lung cancer link")
}

// Test Case CG-002: Confounding Variable
func TestBuildCausalGraph_Confounding(t *testing.T) {
	cr := NewCausalReasoner()

	observations := []string{
		"Ice cream sales correlate with drowning rates",
		"Hot weather increases ice cream sales",
		"Hot weather leads to more swimming",
		"More swimming increases drowning risk",
	}

	graph, err := cr.BuildCausalGraph("Ice cream and drowning", observations)

	assert.NoError(t, err)

	// Should identify hot weather as common cause
	variableNames := extractVariableNames(graph.Variables)
	assert.Contains(t, variableNames, "hot weather")

	// Hot weather should have outgoing links to both ice cream and drowning (indirectly)
	hotWeatherVar := findVariableByName(graph.Variables, "hot weather")
	assert.NotNil(t, hotWeatherVar)

	outgoingLinks := 0
	for _, link := range graph.Links {
		if link.From == hotWeatherVar.ID {
			outgoingLinks++
		}
	}
	assert.GreaterOrEqual(t, outgoingLinks, 1, "Confounding variable should have causal links")
}

// Test Case IS-001: Intervention Simulation
func TestSimulateIntervention_DownstreamEffects(t *testing.T) {
	cr := NewCausalReasoner()

	// Build graph
	observations := []string{
		"Smoking causes lung cancer",
		"Lung cancer increases mortality",
	}
	graph, _ := cr.BuildCausalGraph("Smoking effects", observations)

	// Simulate intervention: reduce smoking
	intervention, err := cr.SimulateIntervention(graph.ID, "smoking", "decrease")

	assert.NoError(t, err)
	assert.NotNil(t, intervention)

	// Should predict decreased lung cancer
	foundLungCancerEffect := false
	for _, effect := range intervention.PredictedEffects {
		if effect.Variable == "lung cancer" {
			foundLungCancerEffect = true
			assert.Equal(t, "decrease", effect.Effect)
			assert.Greater(t, effect.Probability, 0.5)
		}
	}
	assert.True(t, foundLungCancerEffect)

	// Should predict decreased mortality (downstream effect)
	foundMortalityEffect := false
	for _, effect := range intervention.PredictedEffects {
		if effect.Variable == "mortality" {
			foundMortalityEffect = true
			assert.Equal(t, "decrease", effect.Effect)
		}
	}
	assert.True(t, foundMortalityEffect)
}

// Test Case CF-001: Counterfactual Generation
func TestGenerateCounterfactual_WhatIfScenario(t *testing.T) {
	cr := NewCausalReasoner()

	observations := []string{
		"Study time increases exam score",
		"Exam score determines graduation",
	}
	graph, _ := cr.BuildCausalGraph("Academic performance", observations)

	// What if student studied more?
	changes := map[string]string{
		"study time": "increase",
	}

	counterfactual, err := cr.GenerateCounterfactual(
		graph.ID,
		"What if I had studied 10 more hours?",
		changes,
	)

	assert.NoError(t, err)
	assert.NotNil(t, counterfactual)

	// Should predict increased exam score
	assert.Contains(t, counterfactual.Outcomes, "exam score")
	assert.Greater(t, counterfactual.Plausibility, 0.5)
}

// Test Case CC-001: Correlation vs Causation
func TestAnalyzeCorrelationVsCausation_Clear(t *testing.T) {
	cr := NewCausalReasoner()

	tests := []struct {
		name        string
		observation string
		expected    string
	}{
		{
			name:        "clear causation",
			observation: "Pressing the brake pedal causes the car to slow down",
			expected:    "causal",
		},
		{
			name:        "correlation language",
			observation: "Income is correlated with education level",
			expected:    "correlation",
		},
		{
			name:        "ambiguous",
			observation: "People who exercise are healthier",
			expected:    "unclear",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := cr.AnalyzeCorrelationVsCausation(tt.observation)

			assert.NoError(t, err)
			assert.Contains(t, analysis, tt.expected)
		})
	}
}

// Helper functions
func extractVariableNames(vars []*types.CausalVariable) []string {
	names := make([]string, len(vars))
	for i, v := range vars {
		names[i] = v.Name
	}
	return names
}

func findVariable(vars []*types.CausalVariable, id string) *types.CausalVariable {
	for _, v := range vars {
		if v.ID == id {
			return v
		}
	}
	return nil
}

func findVariableByName(vars []*types.CausalVariable, name string) *types.CausalVariable {
	for _, v := range vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}
```

**Expected Results:**
- Variable extraction: 80% recall
- Causal link identification: 75% precision
- Intervention predictions: Correct direction in 90%+ cases
- Counterfactuals: Plausible outcomes

---

## Priority 1 (High) - Implement Second

### 4. Fallacy Detection Precision/Recall Tests

**File:** `internal/validation/fallacies_accuracy_test.go`

```go
package validation

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// Test suite for all 40+ fallacy types
func TestDetectFallacies_FormalFallacies(t *testing.T) {
	fd := NewFallacyDetector()

	tests := []struct {
		name             string
		content          string
		expectedFallacy  string
		expectedCategory FallacyType
		minConfidence    float64
	}{
		{
			name:             "affirming consequent",
			content:          "If it rains, the ground is wet. The ground is wet. Therefore, it rained.",
			expectedFallacy:  "affirming_consequent",
			expectedCategory: FallacyFormal,
			minConfidence:    0.7,
		},
		{
			name:             "denying antecedent",
			content:          "If P then Q. Not P. Therefore, not Q.",
			expectedFallacy:  "denying_antecedent",
			expectedCategory: FallacyFormal,
			minConfidence:    0.7,
		},
		{
			name:             "undistributed middle",
			content:          "All cats are mammals. All dogs are mammals. Therefore, all cats are dogs.",
			expectedFallacy:  "undistributed_middle",
			expectedCategory: FallacyFormal,
			minConfidence:    0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected := fd.DetectFallacies(tt.content, true, true)

			foundExpected := false
			for _, fallacy := range detected {
				if fallacy.Type == tt.expectedFallacy {
					foundExpected = true
					assert.Equal(t, tt.expectedCategory, fallacy.Category)
					assert.GreaterOrEqual(t, fallacy.Confidence, tt.minConfidence)
				}
			}
			assert.True(t, foundExpected, "Should detect %s", tt.expectedFallacy)
		})
	}
}

func TestDetectFallacies_InformalFallacies(t *testing.T) {
	fd := NewFallacyDetector()

	tests := []struct {
		name            string
		content         string
		expectedFallacy string
		minConfidence   float64
	}{
		{
			name:            "ad hominem",
			content:         "Your argument is wrong because you're stupid",
			expectedFallacy: "ad_hominem",
			minConfidence:   0.8,
		},
		{
			name:            "straw man",
			content:         "They claim we should regulate banks. But they want to destroy free markets!",
			expectedFallacy: "straw_man",
			minConfidence:   0.6,
		},
		{
			name:            "false dilemma",
			content:         "You're either with us or against us",
			expectedFallacy: "false_dilemma",
			minConfidence:   0.8,
		},
		{
			name:            "slippery slope",
			content:         "If we allow same-sex marriage, next people will marry animals",
			expectedFallacy: "slippery_slope",
			minConfidence:   0.7,
		},
		{
			name:            "appeal to authority",
			content:         "Einstein believed in God, so atheism must be wrong",
			expectedFallacy: "appeal_to_authority",
			minConfidence:   0.5,
		},
		{
			name:            "appeal to emotion",
			content:         "Think about the children! Imagine how they feel! This is heartbreaking!",
			expectedFallacy: "appeal_to_emotion",
			minConfidence:   0.6,
		},
		{
			name:            "red herring",
			content:         "We're discussing taxes, but speaking of which, did you see the game?",
			expectedFallacy: "red_herring",
			minConfidence:   0.5,
		},
		{
			name:            "hasty generalization",
			content:         "I met one rude person from that city. Everyone from there must be rude.",
			expectedFallacy: "hasty_generalization",
			minConfidence:   0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected := fd.DetectFallacies(tt.content, false, true)

			foundExpected := false
			for _, fallacy := range detected {
				if fallacy.Type == tt.expectedFallacy {
					foundExpected = true
					assert.GreaterOrEqual(t, fallacy.Confidence, tt.minConfidence)
					assert.NotEmpty(t, fallacy.Explanation)
					assert.NotEmpty(t, fallacy.Correction)
				}
			}
			assert.True(t, foundExpected, "Should detect %s", tt.expectedFallacy)
		})
	}
}

// Precision/Recall calculation test
func TestFallacyDetection_PrecisionRecall(t *testing.T) {
	fd := NewFallacyDetector()

	// Test set with labeled fallacies
	testCases := []struct {
		content           string
		hasFallacy        bool
		expectedFallacies []string
	}{
		{
			content:           "You're stupid, so your argument is wrong",
			hasFallacy:        true,
			expectedFallacies: []string{"ad_hominem"},
		},
		{
			content:           "All swans I've seen are white, so all swans are white",
			hasFallacy:        true,
			expectedFallacies: []string{"hasty_generalization"},
		},
		{
			content:           "The data shows a clear correlation between X and Y",
			hasFallacy:        false,
			expectedFallacies: []string{},
		},
		// Add 50+ more labeled examples...
	}

	truePositives := 0
	falsePositives := 0
	falseNegatives := 0

	for _, tc := range testCases {
		detected := fd.DetectFallacies(tc.content, true, true)

		if tc.hasFallacy {
			if len(detected) > 0 {
				// Check if any expected fallacy was found
				foundExpected := false
				for _, d := range detected {
					for _, exp := range tc.expectedFallacies {
						if d.Type == exp {
							foundExpected = true
						}
					}
				}
				if foundExpected {
					truePositives++
				} else {
					falsePositives++
				}
			} else {
				falseNegatives++
			}
		} else {
			if len(detected) > 0 {
				falsePositives++
			}
		}
	}

	precision := float64(truePositives) / float64(truePositives+falsePositives)
	recall := float64(truePositives) / float64(truePositives+falseNegatives)
	f1 := 2 * (precision * recall) / (precision + recall)

	t.Logf("Precision: %.2f", precision)
	t.Logf("Recall: %.2f", recall)
	t.Logf("F1 Score: %.2f", f1)

	assert.GreaterOrEqual(t, precision, 0.80, "Precision should be >= 0.80")
	assert.GreaterOrEqual(t, recall, 0.75, "Recall should be >= 0.75")
	assert.GreaterOrEqual(t, f1, 0.77, "F1 score should be >= 0.77")
}
```

**Expected Results:**
- Precision >= 0.80
- Recall >= 0.75
- F1 Score >= 0.77
- All 40+ fallacy types have test coverage

---

## Implementation Checklist

### Week 1: Setup
- [ ] Create test data repository
- [ ] Set up JSON test fixtures
- [ ] Create table-driven test templates
- [ ] Implement test result aggregation

### Week 2-3: P0 Tests
- [ ] Implement all Argument Analysis tests (20+ cases)
- [ ] Implement Probabilistic Reasoning accuracy tests (15+ cases)
- [ ] Implement Causal Reasoning validation tests (20+ cases)
- [ ] Run coverage analysis

### Week 4-5: P1 Tests
- [ ] Implement Fallacy Detection precision/recall suite (200+ cases)
- [ ] Implement Decision Analysis tests
- [ ] Implement Metacognition tests
- [ ] Performance benchmarks

### Week 6: Integration
- [ ] Cross-tool validation tests
- [ ] Workflow orchestration tests
- [ ] Evidence pipeline tests
- [ ] Full regression suite

---

## Quick Commands

```bash
# Run all tests
go test ./... -v

# Run specific package tests
go test ./internal/analysis -v
go test ./internal/reasoning -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run only accuracy tests
go test ./... -v -run Accuracy

# Run benchmarks
go test ./... -bench=. -benchmem

# Run table-driven tests
go test ./internal/reasoning -v -run TableDriven
```

---

## Success Metrics Dashboard

Track these metrics weekly:

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Overall Coverage | 90% | 73.6% | 🟡 |
| Argument Analysis | 80% | 0% | 🔴 |
| Probabilistic MAE | < 0.01 | TBD | ⚪ |
| Causal F1 | 0.80 | TBD | ⚪ |
| Fallacy F1 | 0.77 | TBD | ⚪ |
| Orchestration Coverage | 80% | 0% | 🔴 |

🔴 Critical  🟡 Needs Improvement  🟢 Good  ⚪ Not Measured

---

**Next Steps:**
1. Start with Argument Analysis tests (highest priority, 0% coverage)
2. Add Probabilistic Reasoning accuracy validation
3. Build up to 400+ total test cases
4. Achieve 90% coverage target
