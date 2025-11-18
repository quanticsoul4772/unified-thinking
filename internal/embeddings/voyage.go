package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// VoyageAI API constants
const (
	voyageAPIURL = "https://api.voyageai.com/v1/embeddings"
)

// VoyageEmbedder implements Embedder using Voyage AI API
type VoyageEmbedder struct {
	client    *http.Client
	apiKey    string
	model     string
	dimension int
	timeout   time.Duration
}

// NewVoyageEmbedder creates a new Voyage AI embedder
func NewVoyageEmbedder(apiKey, model string) *VoyageEmbedder {
	// Model dimensions from Voyage AI documentation
	dimensions := map[string]int{
		"voyage-3-lite":  512,
		"voyage-3":       1024,
		"voyage-3-large": 2048,
		"voyage-code-3":  1536,
		"voyage-finance-2": 1024,
		"voyage-law-2":   1024,
	}

	dim := dimensions[model]
	if dim == 0 {
		dim = 1024 // Default dimension
	}

	return &VoyageEmbedder{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:    apiKey,
		model:     model,
		dimension: dim,
		timeout:   30 * time.Second,
	}
}

// voyageRequest represents the API request
type voyageRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// voyageResponse represents the API response
type voyageResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// Embed generates embedding for single text
func (e *VoyageEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts
func (e *VoyageEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Create request
	reqBody := voyageRequest{
		Model: e.model,
		Input: texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", voyageAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	// Send request
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var voyageResp voyageResponse
	if err := json.Unmarshal(body, &voyageResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract embeddings
	embeddings := make([][]float32, len(voyageResp.Data))
	for _, data := range voyageResp.Data {
		if data.Index < len(embeddings) {
			embeddings[data.Index] = data.Embedding
		}
	}

	return embeddings, nil
}

// Dimension returns the embedding dimension
func (e *VoyageEmbedder) Dimension() int {
	return e.dimension
}

// Model returns the model identifier
func (e *VoyageEmbedder) Model() string {
	return e.model
}

// Provider returns the provider name
func (e *VoyageEmbedder) Provider() string {
	return "voyage"
}