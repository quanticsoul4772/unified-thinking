// Package handlers - Multimodal Embeddings MCP tool handler
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/embeddings"
)

// MultimodalHandler handles multimodal embedding operations
type MultimodalHandler struct {
	embedder    embeddings.MultimodalEmbedder
	imageLoader *embeddings.ImageLoader
}

// NewMultimodalHandler creates a new multimodal handler
func NewMultimodalHandler(embedder embeddings.MultimodalEmbedder) *MultimodalHandler {
	return &MultimodalHandler{
		embedder:    embedder,
		imageLoader: embeddings.NewImageLoader(),
	}
}

// EmbedMultimodalRequest for embed-multimodal tool
type EmbedMultimodalRequest struct {
	Text      string `json:"text,omitempty"`         // Text to embed
	ImagePath string `json:"image_path,omitempty"`   // Path to image file
	ImageURL  string `json:"image_url,omitempty"`    // URL of image
	ImageB64  string `json:"image_base64,omitempty"` // Base64-encoded image
}

// EmbedMultimodalResponse for embed-multimodal tool
type EmbedMultimodalResponse struct {
	Embedding []float32             `json:"embedding"`
	Dimension int                   `json:"dimension"`
	Model     string                `json:"model"`
	InputType string                `json:"input_type"` // "text", "image", "multimodal"
	ImageInfo *embeddings.ImageInfo `json:"image_info,omitempty"`
}

// HandleEmbedMultimodal generates embeddings for multimodal content
func (h *MultimodalHandler) HandleEmbedMultimodal(ctx context.Context, req *mcp.CallToolRequest, request EmbedMultimodalRequest) (*mcp.CallToolResult, *EmbedMultimodalResponse, error) {
	// Check if embedder is available
	if h.embedder == nil {
		return nil, nil, fmt.Errorf("embed-multimodal requires MULTIMODAL_ENABLED=true and VOYAGE_API_KEY")
	}

	// Build multimodal inputs
	inputs, inputType, imageInfo, err := h.buildInputs(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	if len(inputs) == 0 {
		return nil, nil, fmt.Errorf("at least one input (text, image_path, image_url, or image_base64) is required")
	}

	// Generate embedding
	embedding, err := h.embedder.EmbedMultimodal(ctx, inputs)
	if err != nil {
		return nil, nil, fmt.Errorf("embedding failed: %w", err)
	}

	response := &EmbedMultimodalResponse{
		Embedding: embedding,
		Dimension: len(embedding),
		Model:     h.embedder.Model(),
		InputType: inputType,
		ImageInfo: imageInfo,
	}

	return &mcp.CallToolResult{Content: multimodalToJSONContent(response)}, response, nil
}

// buildInputs constructs MultimodalInput slice from the request
func (h *MultimodalHandler) buildInputs(ctx context.Context, request EmbedMultimodalRequest) ([]embeddings.MultimodalInput, string, *embeddings.ImageInfo, error) {
	var inputs []embeddings.MultimodalInput
	var inputType string
	var imageInfo *embeddings.ImageInfo

	hasText := request.Text != ""
	hasImage := request.ImagePath != "" || request.ImageURL != "" || request.ImageB64 != ""

	// Process text
	if hasText {
		inputs = append(inputs, embeddings.MultimodalInput{
			Type: embeddings.InputTypeText,
			Text: request.Text,
		})
		inputType = "text"
	}

	// Process image from path
	if request.ImagePath != "" {
		imageB64, err := h.imageLoader.LoadFromPath(request.ImagePath)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to load image from path: %w", err)
		}

		// Get image info
		data, _ := embeddings.DecodeBase64(imageB64)
		imageInfo, _ = h.imageLoader.GetImageInfo(data)

		inputs = append(inputs, embeddings.MultimodalInput{
			Type:     embeddings.InputTypeImageBase64,
			ImageB64: imageB64,
		})
		inputType = "image"
	}

	// Process image from URL
	if request.ImageURL != "" {
		inputs = append(inputs, embeddings.MultimodalInput{
			Type:     embeddings.InputTypeImageURL,
			ImageURL: request.ImageURL,
		})
		inputType = "image"
	}

	// Process image from base64
	if request.ImageB64 != "" {
		// Validate the image
		data, err := embeddings.DecodeBase64(request.ImageB64)
		if err != nil {
			return nil, "", nil, fmt.Errorf("invalid base64 image: %w", err)
		}

		if err := h.imageLoader.ValidateImage(data); err != nil {
			return nil, "", nil, fmt.Errorf("invalid image: %w", err)
		}

		imageInfo, _ = h.imageLoader.GetImageInfo(data)

		inputs = append(inputs, embeddings.MultimodalInput{
			Type:     embeddings.InputTypeImageBase64,
			ImageB64: request.ImageB64,
		})
		inputType = "image"
	}

	// Determine combined input type
	if hasText && hasImage {
		inputType = "multimodal"
	}

	return inputs, inputType, imageInfo, nil
}

// multimodalToJSONContent converts response to JSON content
func multimodalToJSONContent(data interface{}) []mcp.Content {
	jsonData, err := json.Marshal(data)
	if err != nil {
		errData := map[string]string{"error": err.Error()}
		jsonData, _ = json.Marshal(errData)
	}

	return []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}
}

// RegisterMultimodalTools registers all multimodal MCP tools
func RegisterMultimodalTools(mcpServer *mcp.Server, handler *MultimodalHandler) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "embed-multimodal",
		Description: `Generate embeddings for multimodal content (text, images, or both).

Requires MULTIMODAL_ENABLED=true and VOYAGE_API_KEY environment variables.
Uses Voyage AI's voyage-multimodal-3 model for unified text+image embeddings.

**Parameters:**
- text (optional): Text to embed
- image_path (optional): Local file path to an image (JPEG, PNG, GIF, WebP)
- image_url (optional): URL of an image to embed
- image_base64 (optional): Base64-encoded image data

At least one parameter must be provided. Multiple can be combined for multimodal embeddings.

**Image Constraints:**
- Maximum dimensions: 4096x4096 pixels
- Maximum file size: 20MB
- Supported formats: JPEG, PNG, GIF, WebP

**Returns:**
- embedding: Array of float32 values (1024 dimensions for multimodal model)
- dimension: Number of dimensions in the embedding
- model: Model used for embedding
- input_type: "text", "image", or "multimodal"
- image_info: Image metadata (width, height, format, size) if image was provided

**Examples:**
- Text only: {"text": "A sunset over the ocean"}
- Image from path: {"image_path": "/path/to/image.jpg"}
- Image from URL: {"image_url": "https://example.com/image.png"}
- Multimodal: {"text": "A beautiful sunset", "image_path": "/path/to/sunset.jpg"}

**Use Cases:**
- Semantic search across images and text
- Image-text similarity scoring
- Building multimodal knowledge graphs
- Visual content analysis and retrieval`,
	}, handler.HandleEmbedMultimodal)
}
