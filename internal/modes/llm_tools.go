// Package modes - Anthropic Claude tool definitions for structured outputs
package modes

import "encoding/json"

// Tool definitions for structured outputs (Anthropic tool use)
// Using strict: true guarantees schema-compliant responses

// ToolDefinition represents an Anthropic tool definition
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolChoice specifies which tool the model should use
type ToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// Tool definitions for GoT operations with strict schemas

// GenerateContinuationsTool - for generating diverse thought continuations
var GenerateContinuationsTool = ToolDefinition{
	Name:        "generate_continuations",
	Description: "Generate diverse thought continuations for Graph-of-Thoughts reasoning",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"continuations": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of diverse thought continuations",
			},
		},
		"required":             []string{"continuations"},
		"additionalProperties": false,
	},
}

// ScoreThoughtTool - for scoring thought quality
var ScoreThoughtTool = ToolDefinition{
	Name:        "score_thought",
	Description: "Score thought quality across multiple criteria (0.0-1.0)",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"confidence": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Confidence in the thought's correctness",
			},
			"validity": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Logical validity and soundness",
			},
			"relevance": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Relevance to the problem context",
			},
			"novelty": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Uniqueness and originality",
			},
			"depth_factor": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Depth and thoroughness of analysis",
			},
		},
		"required":             []string{"confidence", "validity", "relevance", "novelty", "depth_factor"},
		"additionalProperties": false,
	},
}

// ExtractKeyPointsTool - for extracting key insights
var ExtractKeyPointsTool = ToolDefinition{
	Name:        "extract_key_points",
	Description: "Extract key insights and points from a thought",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"key_points": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"minItems":    1,
				"maxItems":    5,
				"description": "List of key points or insights (1-5 items)",
			},
		},
		"required":             []string{"key_points"},
		"additionalProperties": false,
	},
}

// CalculateNoveltyTool - for novelty scoring
var CalculateNoveltyTool = ToolDefinition{
	Name:        "calculate_novelty",
	Description: "Calculate novelty score comparing thought to siblings",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"novelty_score": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Novelty score from 0.0 (derivative) to 1.0 (highly unique)",
			},
		},
		"required":             []string{"novelty_score"},
		"additionalProperties": false,
	},
}

// ResearchWithSearchTool - for web-augmented research
var ResearchWithSearchTool = ToolDefinition{
	Name:        "research_response",
	Description: "Provide research findings with citations",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"findings": map[string]interface{}{
				"type":        "string",
				"description": "Research findings and analysis",
			},
			"key_insights": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Key insights from research",
			},
			"confidence": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Confidence in findings based on source quality",
			},
		},
		"required":             []string{"findings", "key_insights", "confidence"},
		"additionalProperties": false,
	},
}

// Structured output response types

// GenerateContinuationsResult is the parsed result from generate_continuations tool
type GenerateContinuationsResult struct {
	Continuations []string `json:"continuations"`
}

// ScoreThoughtResult is the parsed result from score_thought tool
type ScoreThoughtResult struct {
	Confidence  float64 `json:"confidence"`
	Validity    float64 `json:"validity"`
	Relevance   float64 `json:"relevance"`
	Novelty     float64 `json:"novelty"`
	DepthFactor float64 `json:"depth_factor"`
}

// ExtractKeyPointsResult is the parsed result from extract_key_points tool
type ExtractKeyPointsResult struct {
	KeyPoints []string `json:"key_points"`
}

// CalculateNoveltyResult is the parsed result from calculate_novelty tool
type CalculateNoveltyResult struct {
	NoveltyScore float64 `json:"novelty_score"`
}

// ResearchResponseResult is the parsed result from research_response tool
type ResearchResponseResult struct {
	Findings    string   `json:"findings"`
	KeyInsights []string `json:"key_insights"`
	Confidence  float64  `json:"confidence"`
}

// Citation represents a web search citation
type Citation struct {
	Type           string `json:"type,omitempty"`
	URL            string `json:"url"`
	Title          string `json:"title"`
	EncryptedIndex string `json:"encrypted_index,omitempty"`
	CitedText      string `json:"cited_text,omitempty"`
}

// ResearchResult combines research findings with citations
type ResearchResult struct {
	Findings    string     `json:"findings"`
	KeyInsights []string   `json:"key_insights"`
	Confidence  float64    `json:"confidence"`
	Citations   []Citation `json:"citations"`
	Searches    int        `json:"searches_performed"`
}

// Helper to marshal tool definition to JSON for API requests
func (t ToolDefinition) MarshalForAPI() (map[string]interface{}, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
