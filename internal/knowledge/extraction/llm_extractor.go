// Package extraction implements LLM-based entity extraction for complex entities.
package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// LLMExtractor uses an LLM to extract complex entities and relationships
type LLMExtractor struct {
	// For now, this is a placeholder for future LLM integration
	// In production, this would use an LLM client (OpenAI, Anthropic, etc.)
}

// NewLLMExtractor creates a new LLM-based extractor
func NewLLMExtractor() *LLMExtractor {
	return &LLMExtractor{}
}

// Extract extracts complex entities using LLM (placeholder for now)
func (le *LLMExtractor) Extract(content string) (*ExtractionResult, error) {
	result := &ExtractionResult{
		Entities:      []*ExtractedEntity{},
		Relationships: []*ExtractedRelationship{},
		LLMEntities:   0,
		RegexEntities: 0,
	}

	// Placeholder: In production, this would call an LLM with a structured prompt
	// For now, we'll return an empty result to maintain API compatibility

	return result, nil
}

// ExtractWithPrompt would extract entities using a custom prompt (future implementation)
func (le *LLMExtractor) ExtractWithPrompt(ctx context.Context, content string, prompt string) (*ExtractionResult, error) {
	// Future implementation will:
	// 1. Format content with extraction prompt
	// 2. Call LLM API with JSON mode
	// 3. Parse structured JSON response
	// 4. Return entities and relationships

	return &ExtractionResult{
		Entities:      []*ExtractedEntity{},
		Relationships: []*ExtractedRelationship{},
	}, nil
}

// buildExtractionPrompt creates the LLM prompt for entity extraction
func buildExtractionPrompt(content string) string {
	return fmt.Sprintf(`Extract entities and relationships from the following text.

Text:
%s

Extract:
1. Entities: Identify concepts, tools, files, decisions, people, and strategies
2. Relationships: Identify causal (causes, enables), contradictory, and builds-upon relationships

Return a JSON object with this structure:
{
  "entities": [
    {
      "text": "entity name",
      "type": "concept|tool|file|decision|person|strategy",
      "confidence": 0.0-1.0
    }
  ],
  "relationships": [
    {
      "from": "entity1",
      "to": "entity2",
      "type": "CAUSES|ENABLES|CONTRADICTS|BUILDS_UPON",
      "strength": 0.0-1.0,
      "confidence": 0.0-1.0
    }
  ]
}`, content)
}

// parseL LMResponse parses the LLM's JSON response into extraction result
func parseLLMResponse(responseJSON string) (*ExtractionResult, error) {
	var parsed struct {
		Entities []struct {
			Text       string  `json:"text"`
			Type       string  `json:"type"`
			Confidence float64 `json:"confidence"`
		} `json:"entities"`
		Relationships []struct {
			From       string  `json:"from"`
			To         string  `json:"to"`
			Type       string  `json:"type"`
			Strength   float64 `json:"strength"`
			Confidence float64 `json:"confidence"`
		} `json:"relationships"`
	}

	if err := json.Unmarshal([]byte(responseJSON), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	result := &ExtractionResult{
		Entities:      make([]*ExtractedEntity, 0, len(parsed.Entities)),
		Relationships: make([]*ExtractedRelationship, 0, len(parsed.Relationships)),
	}

	for _, e := range parsed.Entities {
		entity := &ExtractedEntity{
			Text:       e.Text,
			Type:       strings.ToLower(e.Type),
			Confidence: e.Confidence,
			Method:     "llm",
		}
		result.Entities = append(result.Entities, entity)
		result.LLMEntities++
	}

	for _, r := range parsed.Relationships {
		rel := &ExtractedRelationship{
			From:       r.From,
			To:         r.To,
			Type:       strings.ToUpper(r.Type),
			Strength:   r.Strength,
			Confidence: r.Confidence,
			Method:     "llm",
		}
		result.Relationships = append(result.Relationships, rel)
	}

	result.TotalEntities = len(result.Entities)
	result.TotalRelationships = len(result.Relationships)

	return result, nil
}
