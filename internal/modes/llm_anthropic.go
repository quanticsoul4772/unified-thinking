// Package modes - Anthropic Claude LLM client for Graph-of-Thoughts
package modes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// AnthropicLLMClient implements LLMClient using Anthropic's Claude API
type AnthropicLLMClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewAnthropicLLMClient creates a new Anthropic LLM client
func NewAnthropicLLMClient() (*AnthropicLLMClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	model := os.Getenv("GOT_MODEL")
	if model == "" {
		model = "claude-sonnet-4-5-20250929" // Sonnet 4.5
	}

	return &AnthropicLLMClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
	}, nil
}

type anthropicRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	Messages  []anthropicMessage  `json:"messages"`
	System    string              `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Generate creates k diverse continuations from a prompt
func (a *AnthropicLLMClient) Generate(ctx context.Context, prompt string, k int) ([]string, error) {
	systemPrompt := fmt.Sprintf(`You are a Graph-of-Thoughts reasoning engine. Generate exactly %d diverse, distinct continuations of the given thought. Each continuation should explore a different angle, approach, or perspective. Be creative and thorough.

Format your response as a JSON array of strings, one per continuation. Example:
["First continuation exploring angle A", "Second continuation exploring angle B", "Third continuation exploring angle C"]`, k)

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 2048,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	response, err := a.makeRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// Parse JSON array from response
	var continuations []string
	if err := json.Unmarshal([]byte(response), &continuations); err != nil {
		return nil, fmt.Errorf("failed to parse continuations as JSON array: %w", err)
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

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	return a.makeRequest(ctx, reqBody)
}

// Refine improves a thought through self-critique
func (a *AnthropicLLMClient) Refine(ctx context.Context, thought string, problem string, refinementCount int) (string, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Critically examine the given thought and improve it. Identify weaknesses, gaps, or opportunities for enhancement. Provide a refined version that is more accurate, comprehensive, or effective.`

	userPrompt := fmt.Sprintf("Problem context: %s\n\nCurrent thought (refinement #%d): %s\n\nProvide an improved version:", problem, refinementCount+1, thought)

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	return a.makeRequest(ctx, reqBody)
}

// Score evaluates thought quality (returns 0.0-1.0)
func (a *AnthropicLLMClient) Score(ctx context.Context, thought string, problem string, criteria map[string]float64) (float64, map[string]float64, error) {
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Evaluate the given thought across multiple criteria. Respond ONLY with a JSON object containing scores from 0.0 to 1.0 for each criterion.

Example format:
{"confidence": 0.8, "validity": 0.9, "relevance": 0.7, "novelty": 0.6, "depth_factor": 0.8}`

	userPrompt := fmt.Sprintf("Problem context: %s\n\nThought to evaluate: %s\n\nProvide scores (0.0-1.0) for: confidence, validity, relevance, novelty, depth_factor", problem, thought)

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 256,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	response, err := a.makeRequest(ctx, reqBody)
	if err != nil {
		return 0, nil, err
	}

	// Parse JSON scores
	var scores map[string]float64
	if err := json.Unmarshal([]byte(response), &scores); err != nil {
		return 0, nil, fmt.Errorf("failed to parse scores as JSON object: %w", err)
	}

	// Validate all required criteria are present
	requiredCriteria := []string{"confidence", "validity", "relevance", "novelty", "depth_factor"}
	for _, criterion := range requiredCriteria {
		if _, ok := scores[criterion]; !ok {
			return 0, nil, fmt.Errorf("missing required criterion in scores: %s", criterion)
		}
	}

	// Calculate overall weighted score
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
	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Extract 3-5 key points or insights from the given thought. Respond ONLY with a JSON array of strings.

Example: ["Key point 1", "Key point 2", "Key point 3"]`

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 512,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: thought},
		},
	}

	response, err := a.makeRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// Parse JSON array
	var keyPoints []string
	if err := json.Unmarshal([]byte(response), &keyPoints); err != nil {
		return nil, fmt.Errorf("failed to parse key points as JSON array: %w", err)
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

	systemPrompt := `You are a Graph-of-Thoughts reasoning engine. Rate how novel/unique the given thought is compared to its siblings on a scale from 0.0 (identical/derivative) to 1.0 (highly novel/unique). Respond with ONLY a decimal number.`

	siblingsList := ""
	for i, sibling := range siblings {
		siblingsList += fmt.Sprintf("\n%d. %s", i+1, sibling)
	}

	userPrompt := fmt.Sprintf("Target thought: %s\n\nSibling thoughts:%s\n\nNovelty score (0.0-1.0):", thought, siblingsList)

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 32,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	response, err := a.makeRequest(ctx, reqBody)
	if err != nil {
		return 0, err
	}

	// Parse float
	var novelty float64
	response = strings.TrimSpace(response)
	if _, err := fmt.Sscanf(response, "%f", &novelty); err != nil {
		return 0, fmt.Errorf("failed to parse novelty score as float: %w", err)
	}

	// Validate range
	if novelty < 0.0 || novelty > 1.0 {
		return 0, fmt.Errorf("novelty score out of range [0.0-1.0]: %f", novelty)
	}

	return novelty, nil
}

// makeRequest sends a request to Anthropic API
func (a *AnthropicLLMClient) makeRequest(ctx context.Context, reqBody anthropicRequest) (string, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return apiResp.Content[0].Text, nil
}
