// Package modes - Anthropic Claude LLM client for Graph-of-Thoughts
package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// AnthropicLLMClient implements LLMClient using Anthropic's Claude API
type AnthropicLLMClient struct {
	*AnthropicBaseClient
	useStructured    bool
	webSearchEnabled bool
}

// NewAnthropicLLMClient creates a new Anthropic LLM client
func NewAnthropicLLMClient() (*AnthropicLLMClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	model := os.Getenv("GOT_MODEL")
	if model == "" {
		model = "claude-sonnet-4-5-20250929"
	}

	useStructured := os.Getenv("GOT_STRUCTURED_OUTPUT") != "false"
	webSearchEnabled := os.Getenv("WEB_SEARCH_ENABLED") == "true"

	return &AnthropicLLMClient{
		AnthropicBaseClient: NewAnthropicBaseClient(BaseClientConfig{
			APIKey: apiKey,
			Model:  model,
		}),
		useStructured:    useStructured,
		webSearchEnabled: webSearchEnabled,
	}, nil
}

// WithWebSearch returns a copy with web search enabled
func (a *AnthropicLLMClient) WithWebSearch() *AnthropicLLMClient {
	return &AnthropicLLMClient{
		AnthropicBaseClient: a.AnthropicBaseClient,
		useStructured:       a.useStructured,
		webSearchEnabled:    true,
	}
}

// WithoutStructured returns a copy with structured outputs disabled
func (a *AnthropicLLMClient) WithoutStructured() *AnthropicLLMClient {
	return &AnthropicLLMClient{
		AnthropicBaseClient: a.AnthropicBaseClient,
		useStructured:       false,
		webSearchEnabled:    a.webSearchEnabled,
	}
}

// Generate creates k diverse continuations from a prompt
func (a *AnthropicLLMClient) Generate(ctx context.Context, prompt string, k int) ([]string, error) {
	if a.useStructured {
		return a.generateStructured(ctx, prompt, k)
	}
	return a.generateLegacy(ctx, prompt, k)
}

func (a *AnthropicLLMClient) generateStructured(ctx context.Context, prompt string, k int) ([]string, error) {
	systemPrompt := fmt.Sprintf(`You are a Graph-of-Thoughts reasoning engine. Generate exactly %d diverse, distinct continuations of the given thought. Each continuation should explore a different angle, approach, or perspective. Be creative and thorough.

Use the generate_continuations tool to provide your response.`, k)

	toolDef, err := GenerateContinuationsTool.MarshalForAPI()
	if err != nil {
		return nil, fmt.Errorf("marshal tool definition: %w", err)
	}

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 2048,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", prompt)},
		Tools:     []Tool{{Name: "generate_continuations", Description: toolDef["description"].(string), InputSchema: toolDef["input_schema"]}},
		ToolChoice: map[string]string{
			"type": "tool",
			"name": "generate_continuations",
		},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	toolInput, err := extractToolInput(resp, "generate_continuations")
	if err != nil {
		return nil, err
	}

	var result GenerateContinuationsResult
	if err := json.Unmarshal(toolInput, &result); err != nil {
		return nil, fmt.Errorf("parse tool response: %w", err)
	}

	if len(result.Continuations) == 0 {
		return nil, fmt.Errorf("no continuations generated")
	}

	return result.Continuations, nil
}

func (a *AnthropicLLMClient) generateLegacy(ctx context.Context, prompt string, k int) ([]string, error) {
	systemPrompt := fmt.Sprintf(`You are a Graph-of-Thoughts reasoning engine. Generate exactly %d diverse, distinct continuations of the given thought. Each continuation should explore a different angle, approach, or perspective. Be creative and thorough.

Format your response as a JSON array of strings, one per continuation. Example:
["First continuation exploring angle A", "Second continuation exploring angle B", "Third continuation exploring angle C"]`, k)

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 2048,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", prompt)},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	text := extractTextFromResponse(resp)
	var continuations []string
	jsonStr := extractJSON(text)
	if err := json.Unmarshal([]byte(jsonStr), &continuations); err != nil {
		return nil, fmt.Errorf("parse continuations as JSON array: %w", err)
	}

	if len(continuations) != k {
		return nil, fmt.Errorf("expected %d continuations, got %d", k, len(continuations))
	}

	return continuations, nil
}

// Aggregate synthesizes multiple thoughts into one
func (a *AnthropicLLMClient) Aggregate(ctx context.Context, thoughts []string, problem string) (string, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Synthesize the given thoughts into a single, unified insight that captures the best aspects of all inputs. The synthesis should be coherent, comprehensive, and more valuable than any individual thought.`

	thoughtsList := ""
	for i, thought := range thoughts {
		thoughtsList += fmt.Sprintf("\n%d. %s", i+1, thought)
	}

	userPrompt := fmt.Sprintf("Problem context: %s\n\nThoughts to synthesize:%s\n\nProvide a synthesized insight that combines the best elements:", problem, thoughtsList)

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", userPrompt)},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return "", err
	}

	return extractTextFromResponse(resp), nil
}

// Refine improves a thought through self-critique
func (a *AnthropicLLMClient) Refine(ctx context.Context, thought string, problem string, refinementCount int) (string, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Critically examine the given thought and improve it. Identify weaknesses, gaps, or opportunities for enhancement. Provide a refined version that is more accurate, comprehensive, or effective.`

	userPrompt := fmt.Sprintf("Problem context: %s\n\nCurrent thought (refinement #%d): %s\n\nProvide an improved version:", problem, refinementCount+1, thought)

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", userPrompt)},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return "", err
	}

	return extractTextFromResponse(resp), nil
}

// Score evaluates thought quality
func (a *AnthropicLLMClient) Score(ctx context.Context, thought string, problem string, criteria map[string]float64) (float64, map[string]float64, error) {
	if a.useStructured {
		return a.scoreStructured(ctx, thought, problem, criteria)
	}
	return a.scoreLegacy(ctx, thought, problem, criteria)
}

func (a *AnthropicLLMClient) scoreStructured(ctx context.Context, thought string, problem string, criteria map[string]float64) (float64, map[string]float64, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Evaluate the given thought across multiple criteria and provide scores from 0.0 to 1.0.

Use the score_thought tool to provide your evaluation.`

	userPrompt := fmt.Sprintf("Problem context: %s\n\nThought to evaluate: %s", problem, thought)

	toolDef, err := ScoreThoughtTool.MarshalForAPI()
	if err != nil {
		return 0, nil, fmt.Errorf("marshal tool definition: %w", err)
	}

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 256,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", userPrompt)},
		Tools:     []Tool{{Name: "score_thought", Description: toolDef["description"].(string), InputSchema: toolDef["input_schema"]}},
		ToolChoice: map[string]string{
			"type": "tool",
			"name": "score_thought",
		},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return 0, nil, err
	}

	toolInput, err := extractToolInput(resp, "score_thought")
	if err != nil {
		return 0, nil, err
	}

	var result ScoreThoughtResult
	if err := json.Unmarshal(toolInput, &result); err != nil {
		return 0, nil, fmt.Errorf("parse tool response: %w", err)
	}

	scores := map[string]float64{
		"confidence":   result.Confidence,
		"validity":     result.Validity,
		"relevance":    result.Relevance,
		"novelty":      result.Novelty,
		"depth_factor": result.DepthFactor,
	}

	overall := 0.0
	for criterion, weight := range criteria {
		score, ok := scores[criterion]
		if !ok {
			return 0, nil, fmt.Errorf("missing score for criterion: %s", criterion)
		}
		overall += score * weight
	}

	return overall, scores, nil
}

func (a *AnthropicLLMClient) scoreLegacy(ctx context.Context, thought string, problem string, criteria map[string]float64) (float64, map[string]float64, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Evaluate the given thought across multiple criteria. Respond ONLY with a JSON object containing scores from 0.0 to 1.0 for each criterion.

Example format:
{"confidence": 0.8, "validity": 0.9, "relevance": 0.7, "novelty": 0.6, "depth_factor": 0.8}`

	userPrompt := fmt.Sprintf("Problem context: %s\n\nThought to evaluate: %s\n\nProvide scores (0.0-1.0) for: confidence, validity, relevance, novelty, depth_factor", problem, thought)

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 256,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", userPrompt)},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return 0, nil, err
	}

	text := extractTextFromResponse(resp)
	var scores map[string]float64
	jsonStr := extractJSON(text)
	if err := json.Unmarshal([]byte(jsonStr), &scores); err != nil {
		return 0, nil, fmt.Errorf("parse scores as JSON object: %w", err)
	}

	requiredCriteria := []string{"confidence", "validity", "relevance", "novelty", "depth_factor"}
	for _, criterion := range requiredCriteria {
		if _, ok := scores[criterion]; !ok {
			return 0, nil, fmt.Errorf("missing required criterion in scores: %s", criterion)
		}
	}

	overall := 0.0
	for criterion, weight := range criteria {
		score, ok := scores[criterion]
		if !ok {
			return 0, nil, fmt.Errorf("missing score for criterion: %s", criterion)
		}
		overall += score * weight
	}

	return overall, scores, nil
}

// ExtractKeyPoints identifies key insights from a thought
func (a *AnthropicLLMClient) ExtractKeyPoints(ctx context.Context, thought string) ([]string, error) {
	if a.useStructured {
		return a.extractKeyPointsStructured(ctx, thought)
	}
	return a.extractKeyPointsLegacy(ctx, thought)
}

func (a *AnthropicLLMClient) extractKeyPointsStructured(ctx context.Context, thought string) ([]string, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Extract 3-5 key points or insights from the given thought.

Use the extract_key_points tool to provide your response.`

	toolDef, err := ExtractKeyPointsTool.MarshalForAPI()
	if err != nil {
		return nil, fmt.Errorf("marshal tool definition: %w", err)
	}

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 512,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", thought)},
		Tools:     []Tool{{Name: "extract_key_points", Description: toolDef["description"].(string), InputSchema: toolDef["input_schema"]}},
		ToolChoice: map[string]string{
			"type": "tool",
			"name": "extract_key_points",
		},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	toolInput, err := extractToolInput(resp, "extract_key_points")
	if err != nil {
		return nil, err
	}

	var result ExtractKeyPointsResult
	if err := json.Unmarshal(toolInput, &result); err != nil {
		return nil, fmt.Errorf("parse tool response: %w", err)
	}

	if len(result.KeyPoints) == 0 {
		return nil, fmt.Errorf("no key points extracted")
	}

	return result.KeyPoints, nil
}

func (a *AnthropicLLMClient) extractKeyPointsLegacy(ctx context.Context, thought string) ([]string, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Extract 3-5 key points or insights from the given thought. Respond ONLY with a JSON array of strings.

Example: ["Key point 1", "Key point 2", "Key point 3"]`

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 512,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", thought)},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	text := extractTextFromResponse(resp)
	var keyPoints []string
	jsonStr := extractJSON(text)
	if err := json.Unmarshal([]byte(jsonStr), &keyPoints); err != nil {
		return nil, fmt.Errorf("parse key points as JSON array: %w", err)
	}

	if len(keyPoints) == 0 {
		return nil, fmt.Errorf("no key points extracted")
	}

	return keyPoints, nil
}

// CalculateNovelty measures uniqueness vs siblings
func (a *AnthropicLLMClient) CalculateNovelty(ctx context.Context, thought string, siblings []string) (float64, error) {
	if len(siblings) == 0 {
		return 1.0, nil
	}

	if a.useStructured {
		return a.calculateNoveltyStructured(ctx, thought, siblings)
	}
	return a.calculateNoveltyLegacy(ctx, thought, siblings)
}

func (a *AnthropicLLMClient) calculateNoveltyStructured(ctx context.Context, thought string, siblings []string) (float64, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Rate how novel/unique the given thought is compared to its siblings on a scale from 0.0 (identical/derivative) to 1.0 (highly novel/unique).

Use the calculate_novelty tool to provide your response.`

	siblingsList := ""
	for i, sibling := range siblings {
		siblingsList += fmt.Sprintf("\n%d. %s", i+1, sibling)
	}

	userPrompt := fmt.Sprintf("Target thought: %s\n\nSibling thoughts:%s", thought, siblingsList)

	toolDef, err := CalculateNoveltyTool.MarshalForAPI()
	if err != nil {
		return 0, fmt.Errorf("marshal tool definition: %w", err)
	}

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 64,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", userPrompt)},
		Tools:     []Tool{{Name: "calculate_novelty", Description: toolDef["description"].(string), InputSchema: toolDef["input_schema"]}},
		ToolChoice: map[string]string{
			"type": "tool",
			"name": "calculate_novelty",
		},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return 0, err
	}

	toolInput, err := extractToolInput(resp, "calculate_novelty")
	if err != nil {
		return 0, err
	}

	var result CalculateNoveltyResult
	if err := json.Unmarshal(toolInput, &result); err != nil {
		return 0, fmt.Errorf("parse tool response: %w", err)
	}

	return result.NoveltyScore, nil
}

func (a *AnthropicLLMClient) calculateNoveltyLegacy(ctx context.Context, thought string, siblings []string) (float64, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Rate how novel/unique the given thought is compared to its siblings on a scale from 0.0 (identical/derivative) to 1.0 (highly novel/unique). Respond with ONLY a decimal number.`

	siblingsList := ""
	for i, sibling := range siblings {
		siblingsList += fmt.Sprintf("\n%d. %s", i+1, sibling)
	}

	userPrompt := fmt.Sprintf("Target thought: %s\n\nSibling thoughts:%s\n\nNovelty score (0.0-1.0):", thought, siblingsList)

	req := &APIRequest{
		Model:     a.Model(),
		MaxTokens: 32,
		System:    systemPrompt,
		Messages:  []Message{NewTextMessage("user", userPrompt)},
	}

	resp, err := a.SendRequest(ctx, req)
	if err != nil {
		return 0, err
	}

	text := extractTextFromResponse(resp)
	text = strings.TrimSpace(text)

	var novelty float64
	if _, err := fmt.Sscanf(text, "%f", &novelty); err != nil {
		return 0, fmt.Errorf("parse novelty score as float: %w", err)
	}

	if novelty < 0.0 || novelty > 1.0 {
		return 0, fmt.Errorf("novelty score out of range [0.0-1.0]: %f", novelty)
	}

	return novelty, nil
}

// ResearchWithSearch performs web-augmented research
func (a *AnthropicLLMClient) ResearchWithSearch(ctx context.Context, query string, problem string) (*ResearchResult, error) {
	if !a.webSearchEnabled {
		return nil, fmt.Errorf("web search is not enabled (set WEB_SEARCH_ENABLED=true)")
	}

	systemPrompt := `You are a research assistant with access to web search. Research the given query thoroughly using web search when needed for current information. Provide well-sourced findings.

Use the research_response tool to provide your findings.`

	userPrompt := query
	if problem != "" {
		userPrompt = fmt.Sprintf("Problem context: %s\n\nResearch query: %s", problem, query)
	}

	responseTool, err := ResearchWithSearchTool.MarshalForAPI()
	if err != nil {
		return nil, fmt.Errorf("marshal tool definition: %w", err)
	}

	result, citations, searches, err := a.runResearchLoop(ctx, systemPrompt, userPrompt, responseTool)
	if err != nil {
		return nil, err
	}

	return &ResearchResult{
		Findings:    result.Findings,
		KeyInsights: result.KeyInsights,
		Confidence:  result.Confidence,
		Citations:   citations,
		Searches:    searches,
	}, nil
}

// runResearchLoop handles multi-turn research with web search
func (a *AnthropicLLMClient) runResearchLoop(ctx context.Context, systemPrompt, userPrompt string, responseTool map[string]any) (*ResearchResponseResult, []Citation, int, error) {
	var citations []Citation
	searchCount := 0
	messages := []Message{NewTextMessage("user", userPrompt)}

	for i := 0; i < 10; i++ {
		req := &APIRequest{
			Model:     a.Model(),
			MaxTokens: 4096,
			System:    systemPrompt,
			Messages:  messages,
			Tools: []Tool{
				{Name: "web_search", Description: WebSearchServerTool["description"].(string), InputSchema: WebSearchServerTool["input_schema"]},
				{Name: "research_response", Description: responseTool["description"].(string), InputSchema: responseTool["input_schema"]},
			},
		}

		resp, err := a.SendRequest(ctx, req)
		if err != nil {
			return nil, nil, 0, err
		}

		for _, block := range resp.Content {
			if block.Type == "tool_use" {
				switch block.Name {
				case "web_search":
					searchCount++
				case "research_response":
					inputBytes, err := json.Marshal(block.Input)
					if err != nil {
						return nil, nil, 0, fmt.Errorf("marshal tool input: %w", err)
					}
					var result ResearchResponseResult
					if err := json.Unmarshal(inputBytes, &result); err != nil {
						return nil, nil, 0, fmt.Errorf("parse research response: %w", err)
					}
					return &result, citations, searchCount, nil
				}
			}
		}

		if resp.StopReason == "end_turn" {
			break
		}

		if resp.StopReason == "tool_use" {
			var assistantBlocks []ContentBlock
			for _, block := range resp.Content {
				assistantBlocks = append(assistantBlocks, ContentBlock{
					Type:  block.Type,
					Text:  block.Text,
					ID:    block.ID,
					Name:  block.Name,
					Input: block.Input,
				})
			}
			messages = append(messages, NewBlockMessage("assistant", assistantBlocks))
		}
	}

	return nil, nil, 0, fmt.Errorf("research did not complete within iteration limit")
}

// Helper functions

func extractToolInput(resp *APIResponse, toolName string) (json.RawMessage, error) {
	for _, block := range resp.Content {
		if block.Type == "tool_use" && block.Name == toolName {
			return json.Marshal(block.Input)
		}
	}
	return nil, fmt.Errorf("no tool_use block for %s in response", toolName)
}

func extractTextFromResponse(resp *APIResponse) string {
	for _, block := range resp.Content {
		if block.Type == "text" {
			return stripMarkdownCodeBlocks(block.Text)
		}
	}
	return ""
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	startObj := strings.Index(s, "{")
	startArr := strings.Index(s, "[")

	start := -1
	endChar := byte('}')
	if startObj >= 0 && (startArr < 0 || startObj < startArr) {
		start = startObj
		endChar = '}'
	} else if startArr >= 0 {
		start = startArr
		endChar = ']'
	}

	if start < 0 {
		return s
	}

	depth := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		if escape {
			escape = false
			continue
		}
		c := s[i]
		if c == '\\' && inString {
			escape = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch c {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 && c == endChar {
				return s[start : i+1]
			}
		}
	}
	return s[start:]
}

func stripMarkdownCodeBlocks(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
