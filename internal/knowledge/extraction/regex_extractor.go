// Package extraction implements entity extraction using regex patterns.
package extraction

import (
	"regexp"
	"strings"
)

// RegexExtractor extracts obvious entities using pattern matching
type RegexExtractor struct {
	patterns []entityPattern
}

type entityPattern struct {
	regex      *regexp.Regexp
	entityType string
	confidence float64
}

// NewRegexExtractor creates a new regex-based extractor
func NewRegexExtractor() *RegexExtractor {
	patterns := []entityPattern{
		// URLs
		{
			regex:      regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`),
			entityType: "url",
			confidence: 0.95,
		},
		// Email addresses
		{
			regex:      regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			entityType: "email",
			confidence: 0.95,
		},
		// File paths (Windows and Unix)
		{
			regex:      regexp.MustCompile(`(?:[A-Za-z]:[/\\]|/)[^\s<>"{}|^` + "`" + `\[\]]*\.[a-zA-Z0-9]+`),
			entityType: "file_path",
			confidence: 0.90,
		},
		// ISO dates (YYYY-MM-DD)
		{
			regex:      regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`),
			entityType: "date",
			confidence: 0.90,
		},
		// Times (HH:MM:SS or HH:MM)
		{
			regex:      regexp.MustCompile(`\b([01]?[0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9])?\b`),
			entityType: "time",
			confidence: 0.85,
		},
		// Code identifiers (camelCase, PascalCase, snake_case)
		{
			regex:      regexp.MustCompile(`\b[a-z][a-zA-Z0-9_]{2,}(?:[A-Z][a-z0-9]+)*\b`),
			entityType: "identifier",
			confidence: 0.70,
		},
		// UUIDs
		{
			regex:      regexp.MustCompile(`\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`),
			entityType: "uuid",
			confidence: 0.95,
		},
		// Numbers with units
		{
			regex:      regexp.MustCompile(`\b\d+\.?\d*\s*(ms|seconds?|minutes?|hours?|days?|KB|MB|GB|%)\b`),
			entityType: "measurement",
			confidence: 0.80,
		},
		// Version numbers (semver)
		{
			regex:      regexp.MustCompile(`\bv?\d+\.\d+\.\d+(?:-[a-zA-Z0-9.]+)?\b`),
			entityType: "version",
			confidence: 0.85,
		},
		// IP addresses
		{
			regex:      regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
			entityType: "ip_address",
			confidence: 0.90,
		},
	}

	return &RegexExtractor{patterns: patterns}
}

// Extract extracts obvious entities using regex patterns
func (re *RegexExtractor) Extract(content string) (*ExtractionResult, error) {
	result := &ExtractionResult{
		Entities:      []*ExtractedEntity{},
		Relationships: []*ExtractedRelationship{},
		RegexEntities: 0,
		LLMEntities:   0,
	}

	seen := make(map[string]bool) // Deduplication

	for _, pattern := range re.patterns {
		matches := pattern.regex.FindAllStringIndex(content, -1)

		for _, match := range matches {
			text := content[match[0]:match[1]]

			// Deduplicate
			key := pattern.entityType + ":" + text
			if seen[key] {
				continue
			}
			seen[key] = true

			// Skip very short identifiers (likely false positives)
			if pattern.entityType == "identifier" && len(text) < 4 {
				continue
			}

			entity := &ExtractedEntity{
				Text:        text,
				Type:        pattern.entityType,
				Confidence:  pattern.confidence,
				StartOffset: match[0],
				EndOffset:   match[1],
				Method:      "regex",
			}

			result.Entities = append(result.Entities, entity)
			result.RegexEntities++
		}
	}

	result.TotalEntities = len(result.Entities)
	return result, nil
}

// ExtractCausalRelationships extracts causal relationships using patterns
func (re *RegexExtractor) ExtractCausalRelationships(content string) []*ExtractedRelationship {
	relationships := []*ExtractedRelationship{}

	// Causal patterns
	causalPatterns := []struct {
		regex      *regexp.Regexp
		relType    string
		confidence float64
	}{
		{
			regex:      regexp.MustCompile(`(?i)(\w+(?:\s+\w+){0,3})\s+causes?\s+(\w+(?:\s+\w+){0,3})`),
			relType:    "CAUSES",
			confidence: 0.85,
		},
		{
			regex:      regexp.MustCompile(`(?i)(\w+(?:\s+\w+){0,3})\s+enables?\s+(\w+(?:\s+\w+){0,3})`),
			relType:    "ENABLES",
			confidence: 0.80,
		},
		{
			regex:      regexp.MustCompile(`(?i)(\w+(?:\s+\w+){0,3})\s+leads? to\s+(\w+(?:\s+\w+){0,3})`),
			relType:    "CAUSES",
			confidence: 0.75,
		},
		{
			regex:      regexp.MustCompile(`(?i)(\w+(?:\s+\w+){0,3})\s+contradicts?\s+(\w+(?:\s+\w+){0,3})`),
			relType:    "CONTRADICTS",
			confidence: 0.80,
		},
		{
			regex:      regexp.MustCompile(`(?i)(\w+(?:\s+\w+){0,3})\s+builds? (?:up)?on\s+(\w+(?:\s+\w+){0,3})`),
			relType:    "BUILDS_UPON",
			confidence: 0.75,
		},
	}

	for _, pattern := range causalPatterns {
		matches := pattern.regex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) < 3 {
				continue
			}

			from := strings.TrimSpace(match[1])
			to := strings.TrimSpace(match[2])

			if from == "" || to == "" {
				continue
			}

			rel := &ExtractedRelationship{
				From:       from,
				To:         to,
				Type:       pattern.relType,
				Strength:   pattern.confidence,
				Confidence: pattern.confidence,
				Method:     "regex",
			}

			relationships = append(relationships, rel)
		}
	}

	return relationships
}
