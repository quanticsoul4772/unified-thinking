// Package modes - Shared Anthropic API client infrastructure
package modes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	anthropicAPIURL  = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	defaultTimeout   = 120 * time.Second
)

// BaseClientConfig configures the base client
type BaseClientConfig struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
}

// AnthropicBaseClient provides shared HTTP infrastructure for Anthropic API
type AnthropicBaseClient struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	httpClient  *http.Client
}

// NewAnthropicBaseClient creates a base client
func NewAnthropicBaseClient(config BaseClientConfig) *AnthropicBaseClient {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &AnthropicBaseClient{
		apiKey:      config.APIKey,
		model:       config.Model,
		maxTokens:   config.MaxTokens,
		temperature: config.Temperature,
		httpClient:  &http.Client{Timeout: timeout},
	}
}

// SendRequest sends a request to the Anthropic API
func (c *AnthropicBaseClient) SendRequest(ctx context.Context, req *APIRequest) (*APIResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &apiResp, nil
}

// APIKey returns the API key
func (c *AnthropicBaseClient) APIKey() string {
	return c.apiKey
}

// Model returns the model ID
func (c *AnthropicBaseClient) Model() string {
	return c.model
}

// MaxTokens returns the max tokens setting
func (c *AnthropicBaseClient) MaxTokens() int {
	return c.maxTokens
}

// Temperature returns the temperature setting
func (c *AnthropicBaseClient) Temperature() float64 {
	return c.temperature
}
