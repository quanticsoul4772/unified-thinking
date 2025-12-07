package reasoning

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// buildHypothesisPrompt creates an LLM prompt for generating domain-specific hypotheses
func (ar *AbductiveReasoner) buildHypothesisPrompt(req *GenerateHypothesesRequest) string {
	// Build observations list
	obsList := make([]string, len(req.Observations))
	for i, obs := range req.Observations {
		conf := ""
		if obs.Confidence > 0 {
			conf = fmt.Sprintf(" (confidence: %.2f)", obs.Confidence)
		}
		obsList[i] = fmt.Sprintf("%d. %s%s", i+1, obs.Description, conf)
	}

	maxHyp := 3
	if req.MaxHypotheses > 0 {
		maxHyp = req.MaxHypotheses
	}

	contextNote := ""
	if req.Context != "" {
		contextNote = fmt.Sprintf("\n\nContext: %s", req.Context)
	}

	prompt := fmt.Sprintf(`You are an expert at abductive reasoning - generating plausible explanations for observations.

Given these observations:
%s%s

Generate %d distinct, domain-specific hypotheses that could explain these observations.

For each hypothesis, provide:
1. A clear, specific explanation (not generic templates)
2. Assumptions required for this explanation to hold
3. Testable predictions this hypothesis makes
4. Parsimony score (0-1, higher = simpler)
5. Prior probability (0-1, how likely before evidence)

Focus on:
- Domain-specific mechanisms and causes (not "a common underlying mechanism")
- Concrete, falsifiable hypotheses
- Diverse explanatory approaches (not variations of same idea)

Return ONLY valid JSON in this format:
{
  "hypotheses": [
    {
      "description": "specific explanation here",
      "assumptions": ["assumption 1", "assumption 2"],
      "predictions": ["testable prediction 1", "testable prediction 2"],
      "parsimony": 0.7,
      "prior_probability": 0.5
    }
  ]
}`, strings.Join(obsList, "\n"), contextNote, maxHyp)

	return prompt
}

// parseHypothesesFromLLM parses LLM JSON response into Hypothesis structs
func (ar *AbductiveReasoner) parseHypothesesFromLLM(response string, observations []*Observation) ([]*Hypothesis, error) {
	// Extract JSON from response (handle markdown code blocks)
	jsonStr := response
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json") + 7
		// Skip newline after ```json
		if start < len(response) && response[start] == '\n' {
			start++
		}
		// Find closing ``` after the opening one
		end := strings.Index(response[start:], "```")
		if end > 0 {
			jsonStr = response[start : start+end]
		}
	} else if strings.Contains(response, "```") {
		start := strings.Index(response, "```") + 3
		if start < len(response) && response[start] == '\n' {
			start++
		}
		end := strings.Index(response[start:], "```")
		if end > 0 {
			jsonStr = response[start : start+end]
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)

	// Parse JSON
	var parsed struct {
		Hypotheses []struct {
			Description      string   `json:"description"`
			Assumptions      []string `json:"assumptions"`
			Predictions      []string `json:"predictions"`
			Parsimony        float64  `json:"parsimony"`
			PriorProbability float64  `json:"prior_probability"`
		} `json:"hypotheses"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w (response: %s)", err, jsonStr)
	}

	// Convert to Hypothesis structs
	obsIDs := make([]string, len(observations))
	for i, obs := range observations {
		obsIDs[i] = obs.ID
	}

	hypotheses := make([]*Hypothesis, len(parsed.Hypotheses))
	now := time.Now()

	for i, h := range parsed.Hypotheses {
		hypotheses[i] = &Hypothesis{
			ID:                   fmt.Sprintf("hyp-llm-%d-%d", now.UnixNano(), i),
			Description:          h.Description,
			Observations:         obsIDs,
			ExplanatoryPower:     0.0, // Will be calculated during evaluation
			Parsimony:            h.Parsimony,
			PriorProbability:     h.PriorProbability,
			PosteriorProbability: h.PriorProbability, // Updated during evaluation
			Assumptions:          h.Assumptions,
			Predictions:          h.Predictions,
			Status:               StatusProposed,
			CreatedAt:            now,
			UpdatedAt:            now,
			Metadata:             map[string]interface{}{"source": "llm"},
		}
	}

	return hypotheses, nil
}
