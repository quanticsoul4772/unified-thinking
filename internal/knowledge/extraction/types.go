// Package extraction provides entity and relationship extraction from reasoning content.
package extraction

// ExtractedEntity represents an entity identified in content
type ExtractedEntity struct {
	Text        string                 `json:"text"`
	Type        string                 `json:"type"`
	Confidence  float64                `json:"confidence"`
	StartOffset int                    `json:"start_offset"`
	EndOffset   int                    `json:"end_offset"`
	Method      string                 `json:"method"` // "regex" or "llm"
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExtractedRelationship represents a relationship identified in content
type ExtractedRelationship struct {
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Type       string                 `json:"type"`
	Strength   float64                `json:"strength"`
	Confidence float64                `json:"confidence"`
	Method     string                 `json:"method"` // "regex" or "llm"
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ExtractionResult contains all extracted entities and relationships
type ExtractionResult struct {
	Entities           []*ExtractedEntity       `json:"entities"`
	Relationships      []*ExtractedRelationship `json:"relationships"`
	TotalEntities      int                      `json:"total_entities"`
	TotalRelationships int                      `json:"total_relationships"`
	RegexEntities      int                      `json:"regex_entities"`
	LLMEntities        int                      `json:"llm_entities"`
	CacheHit           bool                     `json:"cache_hit"`
	LatencyMs          int64                    `json:"latency_ms"`
}

// Extractor interface for entity and relationship extraction
type Extractor interface {
	Extract(content string) (*ExtractionResult, error)
}
