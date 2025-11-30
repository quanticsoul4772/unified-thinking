// Package extraction implements hybrid extraction combining regex and LLM.
package extraction

import (
	"context"
	"strings"
	"time"
)

// HybridExtractor combines regex and LLM extraction strategies
type HybridExtractor struct {
	regexExtractor *RegexExtractor
	llmExtractor   *LLMExtractor
	enableLLM      bool
}

// HybridConfig configures hybrid extraction behavior
type HybridConfig struct {
	EnableLLM bool // Enable LLM extraction for complex entities
}

// NewHybridExtractor creates a new hybrid extractor
func NewHybridExtractor(cfg HybridConfig) *HybridExtractor {
	return &HybridExtractor{
		regexExtractor: NewRegexExtractor(),
		llmExtractor:   NewLLMExtractor(),
		enableLLM:      cfg.EnableLLM,
	}
}

// Extract performs hybrid extraction (regex + optional LLM)
func (he *HybridExtractor) Extract(content string) (*ExtractionResult, error) {
	start := time.Now()

	// Step 1: Always use regex for obvious entities (fast, free)
	regexResult, err := he.regexExtractor.Extract(content)
	if err != nil {
		return nil, err
	}

	// Step 2: Optionally use LLM for complex entities (when enabled)
	var llmResult *ExtractionResult
	if he.enableLLM {
		llmResult, err = he.llmExtractor.Extract(content)
		if err != nil {
			// LLM extraction failed, continue with regex-only results
			llmResult = &ExtractionResult{
				Entities:      []*ExtractedEntity{},
				Relationships: []*ExtractedRelationship{},
			}
		}
	} else {
		llmResult = &ExtractionResult{
			Entities:      []*ExtractedEntity{},
			Relationships: []*ExtractedRelationship{},
		}
	}

	// Step 3: Merge results
	merged := mergeExtractionResults(regexResult, llmResult)
	merged.LatencyMs = time.Since(start).Milliseconds()

	return merged, nil
}

// ExtractWithContext extracts entities with context awareness
func (he *HybridExtractor) ExtractWithContext(ctx context.Context, content string) (*ExtractionResult, error) {
	// Check context deadline
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return he.Extract(content)
}

// ExtractRelationships extracts only relationships (faster than full extraction)
func (he *HybridExtractor) ExtractRelationships(content string) []*ExtractedRelationship {
	// Use regex patterns for now
	return he.regexExtractor.ExtractCausalRelationships(content)
}

// mergeExtractionResults combines regex and LLM results, deduplicating entities
func mergeExtractionResults(regexResult, llmResult *ExtractionResult) *ExtractionResult {
	merged := &ExtractionResult{
		Entities:      make([]*ExtractedEntity, 0),
		Relationships: make([]*ExtractedRelationship, 0),
		RegexEntities: regexResult.RegexEntities,
		LLMEntities:   llmResult.LLMEntities,
	}

	// Add all regex entities
	entityMap := make(map[string]*ExtractedEntity)
	for _, entity := range regexResult.Entities {
		key := entity.Type + ":" + entity.Text
		entityMap[key] = entity
	}

	// Add LLM entities (prefer LLM for duplicates due to higher confidence)
	for _, entity := range llmResult.Entities {
		key := entity.Type + ":" + entity.Text
		if existing, exists := entityMap[key]; exists {
			// Keep entity with higher confidence
			if entity.Confidence > existing.Confidence {
				entityMap[key] = entity
			}
		} else {
			entityMap[key] = entity
		}
	}

	// Convert map to slice
	for _, entity := range entityMap {
		merged.Entities = append(merged.Entities, entity)
	}

	// Merge relationships (simple append for now, could deduplicate)
	merged.Relationships = append(merged.Relationships, regexResult.Relationships...)
	merged.Relationships = append(merged.Relationships, llmResult.Relationships...)

	merged.TotalEntities = len(merged.Entities)
	merged.TotalRelationships = len(merged.Relationships)

	return merged
}

// ShouldUseLLM determines if content is complex enough to warrant LLM extraction
func ShouldUseLLM(content string, regexEntityCount int) bool {
	// Heuristics for LLM usage:
	// 1. Content is long (>200 chars) with few regex matches
	// 2. Contains complex reasoning patterns
	// 3. Has narrative structure (multiple sentences)

	if len(content) < 100 {
		return false // Too short for LLM
	}

	if regexEntityCount > 5 {
		return false // Regex already found plenty
	}

	// Check for reasoning complexity indicators
	complexityIndicators := []string{
		"because", "therefore", "however", "although",
		"causes", "enables", "leads to", "results in",
		"contradicts", "builds upon", "suggests",
	}

	lowerContent := strings.ToLower(content)
	matches := 0
	for _, indicator := range complexityIndicators {
		if strings.Contains(lowerContent, indicator) {
			matches++
		}
	}

	return matches >= 2 // Complex if 2+ reasoning indicators
}
