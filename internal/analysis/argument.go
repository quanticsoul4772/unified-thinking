// Package analysis provides argument decomposition and analysis capabilities.
//
// This module breaks down arguments into their constituent parts: claims, premises,
// hidden assumptions, inference chains, and identifies vulnerabilities.
package analysis

import (
	"fmt"
	"strings"
	"time"
)

// ArgumentType categorizes argument forms
type ArgumentType string

const (
	ArgumentDeductive ArgumentType = "deductive" // Necessarily follows from premises
	ArgumentInductive ArgumentType = "inductive" // Probably follows from premises
	ArgumentAbductive ArgumentType = "abductive" // Best explanation given premises
)

// Premise represents a premise in an argument
type Premise struct {
	Statement string  `json:"statement"`
	Type      string  `json:"type"`      // "factual", "value", "definitional"
	Support   string  `json:"support"`   // Evidence or reasoning
	Certainty float64 `json:"certainty"` // 0.0-1.0
}

// InferenceStep represents a step in logical inference
type InferenceStep struct {
	ID         string   `json:"id"`
	From       []string `json:"from"`       // Premise IDs
	To         string   `json:"to"`         // Conclusion
	Rule       string   `json:"rule"`       // Inference rule used
	Confidence float64  `json:"confidence"` // 0.0-1.0
}

// ArgumentDecomposition represents a fully analyzed argument
type ArgumentDecomposition struct {
	ID                string           `json:"id"`
	MainClaim         string           `json:"main_claim"`
	Premises          []*Premise       `json:"premises"`
	HiddenAssumptions []string         `json:"hidden_assumptions"`
	InferenceChain    []*InferenceStep `json:"inference_chain"`
	ArgumentType      ArgumentType     `json:"type"`
	Strength          float64          `json:"strength"` // 0.0-1.0
	Vulnerabilities   []string         `json:"vulnerabilities"`
	CreatedAt         time.Time        `json:"created_at"`
}

// CounterArgument represents an argument opposing a claim
type CounterArgument struct {
	ID                 string    `json:"id"`
	TargetClaim        string    `json:"target_claim"`
	Strategy           string    `json:"strategy"` // "deny_premise", "break_link", "reductio", "alternative_explanation"
	CounterClaim       string    `json:"counter_claim"`
	SupportingPremises []string  `json:"supporting_premises"`
	Strength           float64   `json:"strength"` // 0.0-1.0
	CreatedAt          time.Time `json:"created_at"`
}

// ArgumentAnalyzer decomposes and analyzes arguments
type ArgumentAnalyzer struct {
	arguments        map[string]*ArgumentDecomposition
	counterArguments map[string]*CounterArgument
}

// NewArgumentAnalyzer creates a new argument analyzer
func NewArgumentAnalyzer() *ArgumentAnalyzer {
	return &ArgumentAnalyzer{
		arguments:        make(map[string]*ArgumentDecomposition),
		counterArguments: make(map[string]*CounterArgument),
	}
}

// DecomposeArgument breaks down an argument into its components
func (aa *ArgumentAnalyzer) DecomposeArgument(text string) (*ArgumentDecomposition, error) {
	if text == "" {
		return nil, fmt.Errorf("argument text is required")
	}

	// Extract main claim (usually the conclusion)
	mainClaim := aa.extractMainClaim(text)

	// Extract premises
	premises := aa.extractPremises(text)

	// Identify hidden assumptions
	assumptions := aa.identifyHiddenAssumptions(text, premises)

	// Build inference chain
	inferenceChain := aa.buildInferenceChain(premises, mainClaim)

	// Determine argument type
	argType := aa.determineArgumentType(text, inferenceChain)

	// Calculate argument strength
	strength := aa.calculateArgumentStrength(premises, inferenceChain, argType)

	// Identify vulnerabilities
	vulnerabilities := aa.identifyVulnerabilities(premises, assumptions, inferenceChain)

	decomposition := &ArgumentDecomposition{
		ID:                fmt.Sprintf("arg_%d", time.Now().UnixNano()),
		MainClaim:         mainClaim,
		Premises:          premises,
		HiddenAssumptions: assumptions,
		InferenceChain:    inferenceChain,
		ArgumentType:      argType,
		Strength:          strength,
		Vulnerabilities:   vulnerabilities,
		CreatedAt:         time.Now(),
	}

	aa.arguments[decomposition.ID] = decomposition
	return decomposition, nil
}

// GenerateCounterArguments generates counter-arguments for a claim
func (aa *ArgumentAnalyzer) GenerateCounterArguments(argumentID string) ([]*CounterArgument, error) {
	argument, exists := aa.arguments[argumentID]
	if !exists {
		return nil, fmt.Errorf("argument %s not found", argumentID)
	}

	counterArgs := []*CounterArgument{}

	// Strategy 1: Deny key premise
	if len(argument.Premises) > 0 {
		weakestPremise := aa.findWeakestPremise(argument.Premises)
		if weakestPremise != nil {
			counter := &CounterArgument{
				ID:           fmt.Sprintf("counter_%d_1", time.Now().UnixNano()),
				TargetClaim:  argument.MainClaim,
				Strategy:     "deny_premise",
				CounterClaim: fmt.Sprintf("The premise '%s' is questionable because...", weakestPremise.Statement),
				SupportingPremises: []string{
					"Available evidence contradicts this premise",
					"Alternative explanations exist",
				},
				Strength:  1.0 - weakestPremise.Certainty,
				CreatedAt: time.Now(),
			}
			counterArgs = append(counterArgs, counter)
			aa.counterArguments[counter.ID] = counter
		}
	}

	// Strategy 2: Break inference link
	if len(argument.InferenceChain) > 0 {
		weakestInference := aa.findWeakestInference(argument.InferenceChain)
		if weakestInference != nil {
			counter := &CounterArgument{
				ID:           fmt.Sprintf("counter_%d_2", time.Now().UnixNano()),
				TargetClaim:  argument.MainClaim,
				Strategy:     "break_link",
				CounterClaim: fmt.Sprintf("The inference from premises to '%s' doesn't necessarily follow", weakestInference.To),
				SupportingPremises: []string{
					"Additional unstated assumptions are required",
					"The logical connection is not guaranteed",
				},
				Strength:  1.0 - weakestInference.Confidence,
				CreatedAt: time.Now(),
			}
			counterArgs = append(counterArgs, counter)
			aa.counterArguments[counter.ID] = counter
		}
	}

	// Strategy 3: Reductio ad absurdum (if applicable)
	if argument.ArgumentType == ArgumentDeductive {
		counter := &CounterArgument{
			ID:           fmt.Sprintf("counter_%d_3", time.Now().UnixNano()),
			TargetClaim:  argument.MainClaim,
			Strategy:     "reductio",
			CounterClaim: fmt.Sprintf("If '%s' were true, it would lead to absurd consequences", argument.MainClaim),
			SupportingPremises: []string{
				"Accepting this claim leads to logical contradictions",
				"The implications are clearly false",
			},
			Strength:  0.7,
			CreatedAt: time.Now(),
		}
		counterArgs = append(counterArgs, counter)
		aa.counterArguments[counter.ID] = counter
	}

	// Strategy 4: Alternative explanation
	if argument.ArgumentType == ArgumentAbductive {
		counter := &CounterArgument{
			ID:           fmt.Sprintf("counter_%d_4", time.Now().UnixNano()),
			TargetClaim:  argument.MainClaim,
			Strategy:     "alternative_explanation",
			CounterClaim: "There is a better explanation for the observed evidence",
			SupportingPremises: []string{
				"Alternative explanation is simpler (Occam's razor)",
				"Alternative better accounts for all evidence",
			},
			Strength:  0.6,
			CreatedAt: time.Now(),
		}
		counterArgs = append(counterArgs, counter)
		aa.counterArguments[counter.ID] = counter
	}

	return counterArgs, nil
}

// GetArgument retrieves an argument by ID
func (aa *ArgumentAnalyzer) GetArgument(id string) (*ArgumentDecomposition, error) {
	arg, exists := aa.arguments[id]
	if !exists {
		return nil, fmt.Errorf("argument %s not found", id)
	}
	return arg, nil
}

// GetCounterArgument retrieves a counter-argument by ID
func (aa *ArgumentAnalyzer) GetCounterArgument(id string) (*CounterArgument, error) {
	counter, exists := aa.counterArguments[id]
	if !exists {
		return nil, fmt.Errorf("counter-argument %s not found", id)
	}
	return counter, nil
}

// Private helper methods

func (aa *ArgumentAnalyzer) extractMainClaim(text string) string {
	lower := strings.ToLower(text)

	// Look for conclusion indicators
	indicators := []string{"therefore", "thus", "hence", "so", "consequently", "in conclusion"}

	for _, indicator := range indicators {
		if idx := strings.Index(lower, indicator); idx != -1 {
			// Extract sentence after indicator
			remaining := text[idx+len(indicator):]
			sentences := strings.Split(remaining, ".")
			if len(sentences) > 0 {
				return strings.TrimSpace(sentences[0])
			}
		}
	}

	// If no indicators, use last sentence as likely conclusion
	sentences := strings.Split(text, ".")
	if len(sentences) > 0 {
		last := strings.TrimSpace(sentences[len(sentences)-1])
		if last != "" {
			return last
		}
		if len(sentences) > 1 {
			return strings.TrimSpace(sentences[len(sentences)-2])
		}
	}

	return "Main claim could not be determined"
}

func (aa *ArgumentAnalyzer) extractPremises(text string) []*Premise {
	premises := []*Premise{}
	sentences := strings.Split(text, ".")

	// Premise indicators
	indicators := []string{"because", "since", "given that", "assuming", "based on", "as"}

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		lower := strings.ToLower(sentence)

		// Check if this is a premise
		isPremise := false
		for _, indicator := range indicators {
			if strings.Contains(lower, indicator) {
				isPremise = true
				break
			}
		}

		// Also treat factual statements as premises
		if !isPremise && (strings.Contains(lower, "is") || strings.Contains(lower, "are") || strings.Contains(lower, "has")) {
			isPremise = true
		}

		if isPremise {
			premiseType := "factual"
			if strings.Contains(lower, "should") || strings.Contains(lower, "ought") {
				premiseType = "value"
			} else if strings.Contains(lower, "defined as") || strings.Contains(lower, "means") {
				premiseType = "definitional"
			}

			certainty := 0.8 // Default certainty
			if strings.Contains(lower, "possibly") || strings.Contains(lower, "might") {
				certainty = 0.5
			} else if strings.Contains(lower, "definitely") || strings.Contains(lower, "certainly") {
				certainty = 0.95
			}

			premises = append(premises, &Premise{
				Statement: sentence,
				Type:      premiseType,
				Support:   "Stated in argument",
				Certainty: certainty,
			})
		}
	}

	return premises
}

func (aa *ArgumentAnalyzer) identifyHiddenAssumptions(text string, premises []*Premise) []string {
	assumptions := []string{}

	// Check for causal assumptions
	if strings.Contains(strings.ToLower(text), "causes") || strings.Contains(strings.ToLower(text), "leads to") {
		assumptions = append(assumptions, "Assumes causal relationship (correlation ≠ causation)")
	}

	// Check for generalization assumptions
	if strings.Contains(strings.ToLower(text), "all") || strings.Contains(strings.ToLower(text), "every") {
		assumptions = append(assumptions, "Assumes universal generalization holds without exceptions")
	}

	// Check for value assumptions
	hasValuePremise := false
	for _, p := range premises {
		if p.Type == "value" {
			hasValuePremise = true
			break
		}
	}
	if hasValuePremise {
		assumptions = append(assumptions, "Assumes shared value system or ethical framework")
	}

	// Check for analogy assumptions
	if strings.Contains(strings.ToLower(text), "like") || strings.Contains(strings.ToLower(text), "similar to") {
		assumptions = append(assumptions, "Assumes relevant similarity between compared cases")
	}

	// Check for authority assumptions
	if strings.Contains(strings.ToLower(text), "expert") || strings.Contains(strings.ToLower(text), "authority") {
		assumptions = append(assumptions, "Assumes authority is reliable and relevant to topic")
	}

	return assumptions
}

func (aa *ArgumentAnalyzer) buildInferenceChain(premises []*Premise, conclusion string) []*InferenceStep {
	chain := []*InferenceStep{}

	if len(premises) == 0 {
		return chain
	}

	// Simple chain: all premises support conclusion
	premiseIDs := []string{}
	for i := range premises {
		premiseIDs = append(premiseIDs, fmt.Sprintf("P%d", i+1))
	}

	step := &InferenceStep{
		ID:         "step_1",
		From:       premiseIDs,
		To:         conclusion,
		Rule:       "combined_premises",
		Confidence: aa.calculateInferenceConfidence(premises),
	}

	chain = append(chain, step)
	return chain
}

func (aa *ArgumentAnalyzer) determineArgumentType(text string, chain []*InferenceStep) ArgumentType {
	lower := strings.ToLower(text)

	// Deductive indicators
	if strings.Contains(lower, "necessarily") || strings.Contains(lower, "must") {
		return ArgumentDeductive
	}

	// Abductive indicators (inference to best explanation)
	if strings.Contains(lower, "best explanation") || strings.Contains(lower, "explains") {
		return ArgumentAbductive
	}

	// Default to inductive
	return ArgumentInductive
}

func (aa *ArgumentAnalyzer) calculateArgumentStrength(premises []*Premise, chain []*InferenceStep, argType ArgumentType) float64 {
	if len(premises) == 0 {
		return 0.0
	}

	// Average premise certainty
	premiseCertainty := 0.0
	for _, p := range premises {
		premiseCertainty += p.Certainty
	}
	premiseCertainty /= float64(len(premises))

	// Average inference confidence
	inferenceConfidence := 1.0
	if len(chain) > 0 {
		inferenceConfidence = 0.0
		for _, step := range chain {
			inferenceConfidence += step.Confidence
		}
		inferenceConfidence /= float64(len(chain))
	}

	// Combine based on argument type
	var strength float64
	switch argType {
	case ArgumentDeductive:
		// Deductive: strength = weakest link (minimum)
		strength = min(premiseCertainty, inferenceConfidence)
	case ArgumentInductive:
		// Inductive: strength = combined probability
		strength = premiseCertainty * inferenceConfidence
	case ArgumentAbductive:
		// Abductive: strength = explanatory power (average with boost)
		strength = (premiseCertainty + inferenceConfidence) / 2.0
		if strength > 0.5 {
			strength += 0.1 // Boost for explanatory coherence
		}
	}

	// Cap at 1.0
	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

func (aa *ArgumentAnalyzer) identifyVulnerabilities(premises []*Premise, assumptions []string, chain []*InferenceStep) []string {
	vulnerabilities := []string{}

	// Check for weak premises
	for _, p := range premises {
		if p.Certainty < 0.6 {
			vulnerabilities = append(vulnerabilities, fmt.Sprintf("Weak premise: '%s' (certainty: %.2f)", p.Statement, p.Certainty))
		}
	}

	// Unstated assumptions are vulnerabilities
	if len(assumptions) > 0 {
		vulnerabilities = append(vulnerabilities, fmt.Sprintf("%d unstated assumptions that could be challenged", len(assumptions)))
	}

	// Check for weak inferences
	for _, step := range chain {
		if step.Confidence < 0.7 {
			vulnerabilities = append(vulnerabilities, fmt.Sprintf("Weak inference: %s to '%s' (confidence: %.2f)", step.Rule, step.To, step.Confidence))
		}
	}

	// Check for missing alternative explanations
	vulnerabilities = append(vulnerabilities, "Alternative explanations not considered")

	return vulnerabilities
}

func (aa *ArgumentAnalyzer) calculateInferenceConfidence(premises []*Premise) float64 {
	if len(premises) == 0 {
		return 0.0
	}

	// Confidence decreases with more premises (more links to break)
	avgCertainty := 0.0
	for _, p := range premises {
		avgCertainty += p.Certainty
	}
	avgCertainty /= float64(len(premises))

	// Penalty for many premises
	penalty := 1.0 - (float64(len(premises)) * 0.05)
	if penalty < 0.5 {
		penalty = 0.5
	}

	return avgCertainty * penalty
}

func (aa *ArgumentAnalyzer) findWeakestPremise(premises []*Premise) *Premise {
	if len(premises) == 0 {
		return nil
	}

	weakest := premises[0]
	for _, p := range premises {
		if p.Certainty < weakest.Certainty {
			weakest = p
		}
	}

	return weakest
}

func (aa *ArgumentAnalyzer) findWeakestInference(chain []*InferenceStep) *InferenceStep {
	if len(chain) == 0 {
		return nil
	}

	weakest := chain[0]
	for _, step := range chain {
		if step.Confidence < weakest.Confidence {
			weakest = step
		}
	}

	return weakest
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
