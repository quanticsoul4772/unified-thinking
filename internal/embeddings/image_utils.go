// Package embeddings provides image loading and processing utilities
package embeddings

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ImageLoader handles loading and preprocessing images for embedding
type ImageLoader struct {
	maxWidth  int
	maxHeight int
	maxBytes  int64

	// HTTP client for URL fetching
	client *http.Client
}

// ImageLoaderConfig holds configuration for ImageLoader
type ImageLoaderConfig struct {
	MaxWidth  int   // Maximum image width (default: 4096)
	MaxHeight int   // Maximum image height (default: 4096)
	MaxBytes  int64 // Maximum file size in bytes (default: 20MB)
}

// DefaultImageLoaderConfig returns the default configuration
func DefaultImageLoaderConfig() ImageLoaderConfig {
	return ImageLoaderConfig{
		MaxWidth:  4096,               // Voyage multimodal limit
		MaxHeight: 4096,               // Voyage multimodal limit
		MaxBytes:  20 * 1024 * 1024,   // 20MB
	}
}

// NewImageLoader creates an image loader with default constraints
func NewImageLoader() *ImageLoader {
	return NewImageLoaderWithConfig(DefaultImageLoaderConfig())
}

// NewImageLoaderWithConfig creates an image loader with custom configuration
func NewImageLoaderWithConfig(cfg ImageLoaderConfig) *ImageLoader {
	return &ImageLoader{
		maxWidth:  cfg.MaxWidth,
		maxHeight: cfg.MaxHeight,
		maxBytes:  cfg.MaxBytes,
		client: &http.Client{
			Timeout: 30 * 1000 * 1000 * 1000, // 30 seconds in nanoseconds
		},
	}
}

// LoadFromPath loads and encodes an image from file path
func (l *ImageLoader) LoadFromPath(path string) (string, error) {
	// Clean and validate path
	cleanPath := filepath.Clean(path)

	// Check file exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("image file not found: %s", cleanPath)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Check file size
	if info.Size() > l.maxBytes {
		return "", fmt.Errorf("image too large: %d bytes (max %d)", info.Size(), l.maxBytes)
	}

	// Read file
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	// Validate image format and dimensions
	if err := l.ValidateImage(data); err != nil {
		return "", err
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(data), nil
}

// LoadFromURL fetches and encodes an image from URL
func (l *ImageLoader) LoadFromURL(ctx context.Context, url string) (string, error) {
	// Validate URL scheme
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("invalid URL scheme: must be http or https")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to avoid blocks
	req.Header.Set("User-Agent", "unified-thinking/1.0 (image-loader)")

	// Send request
	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: HTTP %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !isImageContentType(contentType) {
		return "", fmt.Errorf("invalid content type: %s (expected image/*)", contentType)
	}

	// Read with size limit
	data, err := io.ReadAll(io.LimitReader(resp.Body, l.maxBytes+1))
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Check size limit
	if int64(len(data)) > l.maxBytes {
		return "", fmt.Errorf("image too large: exceeds %d bytes", l.maxBytes)
	}

	// Validate image format and dimensions
	if err := l.ValidateImage(data); err != nil {
		return "", err
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(data), nil
}

// LoadFromBytes encodes image bytes to base64 after validation
func (l *ImageLoader) LoadFromBytes(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("image data is empty")
	}

	if int64(len(data)) > l.maxBytes {
		return "", fmt.Errorf("image too large: %d bytes (max %d)", len(data), l.maxBytes)
	}

	if err := l.ValidateImage(data); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// ValidateImage checks image format and dimensions
func (l *ImageLoader) ValidateImage(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("image data is empty")
	}

	// Decode image config (doesn't decode full image, just header)
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("invalid image format: %w", err)
	}

	// Check dimensions
	if cfg.Width > l.maxWidth || cfg.Height > l.maxHeight {
		return fmt.Errorf("image too large: %dx%d (max %dx%d)",
			cfg.Width, cfg.Height, l.maxWidth, l.maxHeight)
	}

	// Validate supported format
	if !isSupportedFormat(format) {
		return fmt.Errorf("unsupported image format: %s (supported: jpeg, png, gif, webp)", format)
	}

	return nil
}

// GetImageInfo returns information about an image without loading it fully
func (l *ImageLoader) GetImageInfo(data []byte) (*ImageInfo, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &ImageInfo{
		Width:  cfg.Width,
		Height: cfg.Height,
		Format: format,
		Size:   int64(len(data)),
	}, nil
}

// ImageInfo contains metadata about an image
type ImageInfo struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"` // jpeg, png, gif, webp
	Size   int64  `json:"size"`   // Size in bytes
}

// isImageContentType checks if content type is a valid image type
func isImageContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "image/")
}

// isSupportedFormat checks if the image format is supported
func isSupportedFormat(format string) bool {
	switch strings.ToLower(format) {
	case "jpeg", "jpg", "png", "gif", "webp":
		return true
	default:
		return false
	}
}

// DetectMIMEType returns the MIME type for image data
func DetectMIMEType(data []byte) string {
	if len(data) < 8 {
		return ""
	}

	// Check magic bytes
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg"
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
		return "image/png"
	case bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")):
		return "image/gif"
	case bytes.HasPrefix(data, []byte("RIFF")) && len(data) > 12 && string(data[8:12]) == "WEBP":
		return "image/webp"
	default:
		return ""
	}
}

// DecodeBase64 decodes a base64-encoded image string
func DecodeBase64(encoded string) ([]byte, error) {
	// Handle data URL prefix if present
	if idx := strings.Index(encoded, ","); idx != -1 {
		encoded = encoded[idx+1:]
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Try URL-safe encoding
		data, err = base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	}

	return data, nil
}
