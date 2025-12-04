// Package embeddings - Voyage AI reranker for search result relevance optimization
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

// Voyage Rerank API constants
const (
	voyageRerankURL = "https://api.voyageai.com/v1/rerank"
)

// Reranker interface for result reranking
type Reranker interface {
	// Rerank reorders documents by relevance to query
	Rerank(ctx context.Context, query string, documents []string, topK int) ([]RerankResult, error)
	// Model returns the reranker model name
	Model() string
}

// RerankResult represents a reranked document
type RerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
	Document       string  `json:"document,omitempty"` // Optional, if return_documents=true
}

// VoyageReranker implements Reranker using Voyage AI
type VoyageReranker struct {
	client      *http.Client
	apiKey      string
	model       string
	rateLimiter *tokenBucketLimiter
}

// NewVoyageReranker creates a new Voyage AI reranker
// Models: rerank-2 (8K context, recommended), rerank-2-lite (4K context, faster)
func NewVoyageReranker(apiKey, model string) *VoyageReranker {
	if model == "" {
		model = "rerank-2" // Default to full model
	}
	return &VoyageReranker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:      apiKey,
		model:       model,
		rateLimiter: newTokenBucketLimiter(defaultRateLimit, defaultBurstLimit),
	}
}

// rerankRequest represents the API request
type rerankRequest struct {
	Model           string   `json:"model"`
	Query           string   `json:"query"`
	Documents       []string `json:"documents"`
	TopK            int      `json:"top_k,omitempty"`
	ReturnDocuments bool     `json:"return_documents,omitempty"`
}

// rerankResponse represents the API response
type rerankResponse struct {
	Results []RerankResult `json:"results"`
	Model   string         `json:"model"`
	Usage   struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// Rerank reorders documents by relevance to the query
func (r *VoyageReranker) Rerank(ctx context.Context, query string, documents []string, topK int) ([]RerankResult, error) {
	if len(documents) == 0 {
		return nil, nil
	}

	if query == "" {
		return nil, fmt.Errorf("query is required for reranking")
	}

	// Rate limiting
	if err := r.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Ensure topK doesn't exceed document count
	if topK <= 0 || topK > len(documents) {
		topK = len(documents)
	}

	reqBody := rerankRequest{
		Model:           r.model,
		Query:           query,
		Documents:       documents,
		TopK:            topK,
		ReturnDocuments: false, // We already have the documents
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", voyageRerankURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage rerank API error (status %d): %s", resp.StatusCode, string(body))
	}

	var rerankResp rerankResponse
	if err := json.Unmarshal(body, &rerankResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return rerankResp.Results, nil
}

// Model returns the reranker model name
func (r *VoyageReranker) Model() string {
	return r.model
}

// RerankWithDocuments returns results with original documents attached
func (r *VoyageReranker) RerankWithDocuments(ctx context.Context, query string, documents []string, topK int) ([]RerankResult, error) {
	results, err := r.Rerank(ctx, query, documents, topK)
	if err != nil {
		return nil, err
	}

	// Attach original documents to results
	for i := range results {
		if results[i].Index < len(documents) {
			results[i].Document = documents[results[i].Index]
		}
	}

	return results, nil
}
