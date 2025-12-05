// Package embeddings provides vector embedding generation including multimodal support
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

// Voyage Multimodal API constants
const (
	voyageMultimodalAPIURL = "https://api.voyageai.com/v1/multimodalembeddings"
	defaultMultimodalModel = "voyage-multimodal-3"
	multimodalDimension    = 1024 // voyage-multimodal-3 produces 1024-dim vectors
)

// VoyageMultimodalEmbedder implements MultimodalEmbedder using Voyage AI API
type VoyageMultimodalEmbedder struct {
	*VoyageEmbedder // Embed base for text operations
	multimodalModel string
}

// NewVoyageMultimodalEmbedder creates a multimodal embedder
// textModel is used for text-only operations (backwards compatibility)
// multimodalModel defaults to "voyage-multimodal-3" if empty
func NewVoyageMultimodalEmbedder(apiKey, textModel, multimodalModel string) *VoyageMultimodalEmbedder {
	base := NewVoyageEmbedder(apiKey, textModel)

	if multimodalModel == "" {
		multimodalModel = defaultMultimodalModel
	}

	return &VoyageMultimodalEmbedder{
		VoyageEmbedder:  base,
		multimodalModel: multimodalModel,
	}
}

// multimodalRequest represents the multimodal API request
type multimodalRequest struct {
	Model  string                   `json:"model"`
	Inputs []multimodalInputWrapper `json:"inputs"`
}

// multimodalInputWrapper wraps content array for a single embedding
type multimodalInputWrapper struct {
	Content []map[string]interface{} `json:"content"`
}

// multimodalResponse represents the multimodal API response
type multimodalResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
		ImagePixels int `json:"image_pixels,omitempty"`
	} `json:"usage"`
}

// EmbedMultimodal generates embedding for multimodal content
func (e *VoyageMultimodalEmbedder) EmbedMultimodal(ctx context.Context, inputs []MultimodalInput) ([]float32, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no inputs provided")
	}

	// Wait for rate limiter
	if err := e.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Convert inputs to API format
	contentItems := make([]map[string]interface{}, len(inputs))
	for i, input := range inputs {
		contentItems[i] = input.ToAPIFormat()
	}

	// Build request body
	reqBody := multimodalRequest{
		Model: e.multimodalModel,
		Inputs: []multimodalInputWrapper{
			{Content: contentItems},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", voyageMultimodalAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	// Send request with retry logic
	maxRetries := 3
	baseDelay := 2 * time.Second

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			// Recreate request for retry
			req, err = http.NewRequestWithContext(ctx, "POST", voyageMultimodalAPIURL, bytes.NewBuffer(jsonData))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+e.apiKey)
		}

		embedding, err := e.doMultimodalRequest(req)
		if err == nil {
			return embedding, nil
		}

		lastErr = err
		if !isRateLimitError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// doMultimodalRequest executes the HTTP request and parses response
func (e *VoyageMultimodalEmbedder) doMultimodalRequest(req *http.Request) ([]float32, error) {
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage multimodal API error (status %d): %s", resp.StatusCode, string(body))
	}

	var mmResp multimodalResponse
	if err := json.Unmarshal(body, &mmResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(mmResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return mmResp.Data[0].Embedding, nil
}

// EmbedImage generates embedding for a single image
func (e *VoyageMultimodalEmbedder) EmbedImage(ctx context.Context, imageBase64 string) ([]float32, error) {
	if imageBase64 == "" {
		return nil, fmt.Errorf("image data is empty")
	}

	return e.EmbedMultimodal(ctx, []MultimodalInput{
		{Type: InputTypeImageBase64, ImageB64: imageBase64},
	})
}

// EmbedImageWithText generates embedding for image with accompanying text description
func (e *VoyageMultimodalEmbedder) EmbedImageWithText(ctx context.Context, imageBase64, text string) ([]float32, error) {
	if imageBase64 == "" {
		return nil, fmt.Errorf("image data is empty")
	}

	inputs := []MultimodalInput{
		{Type: InputTypeImageBase64, ImageB64: imageBase64},
	}

	if text != "" {
		inputs = append(inputs, MultimodalInput{Type: InputTypeText, Text: text})
	}

	return e.EmbedMultimodal(ctx, inputs)
}

// EmbedImageURL generates embedding for an image from URL
func (e *VoyageMultimodalEmbedder) EmbedImageURL(ctx context.Context, imageURL string) ([]float32, error) {
	if imageURL == "" {
		return nil, fmt.Errorf("image URL is empty")
	}

	return e.EmbedMultimodal(ctx, []MultimodalInput{
		{Type: InputTypeImageURL, ImageURL: imageURL},
	})
}

// EmbedImageURLWithText generates embedding for image URL with text description
func (e *VoyageMultimodalEmbedder) EmbedImageURLWithText(ctx context.Context, imageURL, text string) ([]float32, error) {
	if imageURL == "" {
		return nil, fmt.Errorf("image URL is empty")
	}

	inputs := []MultimodalInput{
		{Type: InputTypeImageURL, ImageURL: imageURL},
	}

	if text != "" {
		inputs = append(inputs, MultimodalInput{Type: InputTypeText, Text: text})
	}

	return e.EmbedMultimodal(ctx, inputs)
}

// SupportsMultimodal returns true as this embedder supports multimodal input
func (e *VoyageMultimodalEmbedder) SupportsMultimodal() bool {
	return true
}

// MultimodalModel returns the multimodal model identifier
func (e *VoyageMultimodalEmbedder) MultimodalModel() string {
	return e.multimodalModel
}

// MultimodalDimension returns the dimension for multimodal embeddings
func (e *VoyageMultimodalEmbedder) MultimodalDimension() int {
	return multimodalDimension
}
