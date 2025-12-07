package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"unified-thinking/internal/types"
)

// PerspectiveGenerator interface for generating perspectives via LLM
type PerspectiveGenerator interface {
	GeneratePerspectives(ctx context.Context, prompt string) (string, error)
}

// LLMPerspectiveAnalyzer uses LLM for generating context-aware perspectives
type LLMPerspectiveAnalyzer struct {
	generator PerspectiveGenerator
	fallback  *PerspectiveAnalyzer // Fallback to template-based when LLM unavailable
	counter   int
}

// NewLLMPerspectiveAnalyzer creates a new LLM-enhanced perspective analyzer
func NewLLMPerspectiveAnalyzer(generator PerspectiveGenerator, fallback *PerspectiveAnalyzer) *LLMPerspectiveAnalyzer {
	return &LLMPerspectiveAnalyzer{
		generator: generator,
		fallback:  fallback,
	}
}

// AnalyzePerspectives generates perspectives using LLM when available
func (pa *LLMPerspectiveAnalyzer) AnalyzePerspectives(ctx context.Context, situation string, stakeholderHints []string) ([]*types.Perspective, error) {
	if situation == "" {
		return nil, fmt.Errorf("situation cannot be empty")
	}

	// If no LLM generator, use template-based fallback
	if pa.generator == nil {
		log.Printf("[WARN] No LLM generator configured for perspective analysis, using template fallback")
		return pa.fallback.AnalyzePerspectives(situation, stakeholderHints)
	}

	// Build prompt for LLM
	prompt := pa.buildPerspectivePrompt(situation, stakeholderHints)

	// Generate perspectives via LLM
	response, err := pa.generator.GeneratePerspectives(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM perspective generation failed: %w", err)
	}

	// Parse response
	perspectives, err := pa.parsePerspectivesFromLLM(response, stakeholderHints)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return perspectives, nil
}

// buildPerspectivePrompt creates an LLM prompt for generating stakeholder perspectives
func (pa *LLMPerspectiveAnalyzer) buildPerspectivePrompt(situation string, stakeholders []string) string {
	stakeholderList := ""
	if len(stakeholders) > 0 {
		stakeholderList = "Stakeholders to analyze:\n"
		for i, s := range stakeholders {
			stakeholderList += fmt.Sprintf("%d. %s\n", i+1, s)
		}
	} else {
		stakeholderList = "Identify 3-5 relevant stakeholders for this situation."
	}

	return fmt.Sprintf(`You are an expert at stakeholder analysis and perspective-taking.

Given this situation:
%s

%s

For each stakeholder, generate a detailed, context-specific perspective that includes:
1. viewpoint: A unique first-person perspective that reflects their specific role, interests, and concerns about THIS situation (not generic boilerplate)
2. concerns: 3-5 specific concerns THIS stakeholder would have about THIS situation
3. priorities: 3 priorities THIS stakeholder would focus on
4. constraints: 2-3 constraints or limitations THIS stakeholder faces
5. confidence: Your confidence in this perspective (0.0-1.0)

IMPORTANT:
- Each perspective MUST be specific to the situation described
- Viewpoints should be written in first person as if you ARE that stakeholder
- Include domain-specific terminology and reasoning appropriate to the stakeholder
- Do NOT use generic templates - each response must reflect deep understanding of the stakeholder's unique position

Return ONLY valid JSON in this format:
{
  "perspectives": [
    {
      "stakeholder": "stakeholder name",
      "viewpoint": "First-person perspective specific to this situation...",
      "concerns": ["specific concern 1", "specific concern 2", "specific concern 3"],
      "priorities": ["priority 1", "priority 2", "priority 3"],
      "constraints": ["constraint 1", "constraint 2"],
      "confidence": 0.8
    }
  ]
}`, situation, stakeholderList)
}

// parsePerspectivesFromLLM parses LLM JSON response into Perspective structs
func (pa *LLMPerspectiveAnalyzer) parsePerspectivesFromLLM(response string, stakeholderHints []string) ([]*types.Perspective, error) {
	// Extract JSON from response (handle markdown code blocks)
	jsonStr := response

	// Remove markdown code blocks if present
	if idx := strings.Index(response, "```json\n"); idx >= 0 {
		start := idx + 8 // len("```json\n")
		if end := strings.Index(response[start:], "\n```"); end >= 0 {
			jsonStr = response[start : start+end]
		}
	} else if idx := strings.Index(response, "```json"); idx >= 0 {
		start := idx + 7
		if end := strings.Index(response[start:], "```"); end >= 0 {
			jsonStr = response[start : start+end]
		}
	} else if idx := strings.Index(response, "```\n"); idx >= 0 {
		start := idx + 4
		if end := strings.Index(response[start:], "\n```"); end >= 0 {
			jsonStr = response[start : start+end]
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)

	// Parse JSON
	var parsed struct {
		Perspectives []struct {
			Stakeholder string   `json:"stakeholder"`
			Viewpoint   string   `json:"viewpoint"`
			Concerns    []string `json:"concerns"`
			Priorities  []string `json:"priorities"`
			Constraints []string `json:"constraints"`
			Confidence  float64  `json:"confidence"`
		} `json:"perspectives"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w (response: %s)", err, jsonStr)
	}

	// Convert to Perspective structs
	perspectives := make([]*types.Perspective, len(parsed.Perspectives))
	now := time.Now()

	for i, p := range parsed.Perspectives {
		pa.counter++
		perspectives[i] = &types.Perspective{
			ID:          fmt.Sprintf("perspective-llm-%d-%d", now.UnixNano(), pa.counter),
			Stakeholder: p.Stakeholder,
			Viewpoint:   p.Viewpoint,
			Concerns:    p.Concerns,
			Priorities:  p.Priorities,
			Constraints: p.Constraints,
			Confidence:  p.Confidence,
			Metadata:    map[string]interface{}{"source": "llm"},
			CreatedAt:   now,
		}
	}

	return perspectives, nil
}

// HasGenerator returns true if an LLM generator is configured
func (pa *LLMPerspectiveAnalyzer) HasGenerator() bool {
	return pa.generator != nil
}
