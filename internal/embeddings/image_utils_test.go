// Package embeddings - Tests for ImageLoader
package embeddings

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// createTestPNG creates a small valid PNG image for testing
func createTestPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with solid color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestNewImageLoader(t *testing.T) {
	loader := NewImageLoader()
	if loader == nil {
		t.Fatal("expected non-nil loader")
	}

	if loader.maxWidth != 4096 {
		t.Errorf("maxWidth = %d, want 4096", loader.maxWidth)
	}
	if loader.maxHeight != 4096 {
		t.Errorf("maxHeight = %d, want 4096", loader.maxHeight)
	}
	if loader.maxBytes != 20*1024*1024 {
		t.Errorf("maxBytes = %d, want %d", loader.maxBytes, 20*1024*1024)
	}
}

func TestNewImageLoaderWithConfig(t *testing.T) {
	cfg := ImageLoaderConfig{
		MaxWidth:  100,
		MaxHeight: 200,
		MaxBytes:  1024,
	}

	loader := NewImageLoaderWithConfig(cfg)

	if loader.maxWidth != 100 {
		t.Errorf("maxWidth = %d, want 100", loader.maxWidth)
	}
	if loader.maxHeight != 200 {
		t.Errorf("maxHeight = %d, want 200", loader.maxHeight)
	}
	if loader.maxBytes != 1024 {
		t.Errorf("maxBytes = %d, want 1024", loader.maxBytes)
	}
}

func TestImageLoader_ValidateImage(t *testing.T) {
	loader := NewImageLoaderWithConfig(ImageLoaderConfig{
		MaxWidth:  100,
		MaxHeight: 100,
		MaxBytes:  1024 * 1024,
	})

	tests := []struct {
		name      string
		imageData []byte
		wantErr   bool
	}{
		{
			name:      "valid small PNG",
			imageData: createTestPNG(50, 50),
			wantErr:   false,
		},
		{
			name:      "empty data",
			imageData: []byte{},
			wantErr:   true,
		},
		{
			name:      "invalid image data",
			imageData: []byte("not an image"),
			wantErr:   true,
		},
		{
			name:      "image too wide",
			imageData: createTestPNG(200, 50),
			wantErr:   true,
		},
		{
			name:      "image too tall",
			imageData: createTestPNG(50, 200),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.ValidateImage(tt.imageData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageLoader_LoadFromPath(t *testing.T) {
	loader := NewImageLoader()

	// Create a temp file with a valid PNG
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.png")
	imgData := createTestPNG(100, 100)

	if err := os.WriteFile(tmpFile, imgData, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test loading valid file
	t.Run("valid file", func(t *testing.T) {
		b64, err := loader.LoadFromPath(tmpFile)
		if err != nil {
			t.Fatalf("LoadFromPath failed: %v", err)
		}

		// Verify it's valid base64
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			t.Fatalf("failed to decode base64: %v", err)
		}

		if !bytes.Equal(decoded, imgData) {
			t.Error("decoded data doesn't match original")
		}
	})

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		_, err := loader.LoadFromPath("/nonexistent/path/image.png")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestImageLoader_LoadFromBytes(t *testing.T) {
	loader := NewImageLoader()

	t.Run("valid image", func(t *testing.T) {
		imgData := createTestPNG(100, 100)
		b64, err := loader.LoadFromBytes(imgData)
		if err != nil {
			t.Fatalf("LoadFromBytes failed: %v", err)
		}

		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			t.Fatalf("failed to decode base64: %v", err)
		}

		if !bytes.Equal(decoded, imgData) {
			t.Error("decoded data doesn't match original")
		}
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := loader.LoadFromBytes([]byte{})
		if err == nil {
			t.Error("expected error for empty data")
		}
	})

	t.Run("invalid image", func(t *testing.T) {
		_, err := loader.LoadFromBytes([]byte("not an image"))
		if err == nil {
			t.Error("expected error for invalid image")
		}
	})
}

func TestImageLoader_GetImageInfo(t *testing.T) {
	loader := NewImageLoader()

	t.Run("valid PNG", func(t *testing.T) {
		imgData := createTestPNG(150, 200)
		info, err := loader.GetImageInfo(imgData)
		if err != nil {
			t.Fatalf("GetImageInfo failed: %v", err)
		}

		if info.Width != 150 {
			t.Errorf("Width = %d, want 150", info.Width)
		}
		if info.Height != 200 {
			t.Errorf("Height = %d, want 200", info.Height)
		}
		if info.Format != "png" {
			t.Errorf("Format = %s, want png", info.Format)
		}
		if info.Size != int64(len(imgData)) {
			t.Errorf("Size = %d, want %d", info.Size, len(imgData))
		}
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := loader.GetImageInfo([]byte{})
		if err == nil {
			t.Error("expected error for empty data")
		}
	})
}

func TestImageLoader_LoadFromURL(t *testing.T) {
	loader := NewImageLoader()

	// Test invalid URL scheme
	t.Run("invalid scheme", func(t *testing.T) {
		_, err := loader.LoadFromURL(context.Background(), "ftp://example.com/image.png")
		if err == nil {
			t.Error("expected error for invalid URL scheme")
		}
	})
}

func TestDetectMIMEType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantMIME string
	}{
		{
			name:     "PNG",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantMIME: "image/png",
		},
		{
			name:     "JPEG",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46},
			wantMIME: "image/jpeg",
		},
		{
			name:     "GIF89a",
			data:     []byte("GIF89a" + "additional data"),
			wantMIME: "image/gif",
		},
		{
			name:     "GIF87a",
			data:     []byte("GIF87a" + "additional data"),
			wantMIME: "image/gif",
		},
		{
			name:     "unknown",
			data:     []byte("unknown format data"),
			wantMIME: "",
		},
		{
			name:     "too short",
			data:     []byte{0x00, 0x01},
			wantMIME: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime := DetectMIMEType(tt.data)
			if mime != tt.wantMIME {
				t.Errorf("DetectMIMEType() = %v, want %v", mime, tt.wantMIME)
			}
		})
	}
}

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		wantErr bool
	}{
		{
			name:    "standard base64",
			encoded: base64.StdEncoding.EncodeToString([]byte("hello")),
			wantErr: false,
		},
		{
			name:    "data URL format",
			encoded: "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("test")),
			wantErr: false,
		},
		{
			name:    "invalid base64",
			encoded: "not valid base64!@#$",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeBase64(tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeBase64() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		format    string
		supported bool
	}{
		{"jpeg", true},
		{"jpg", true},
		{"png", true},
		{"gif", true},
		{"webp", true},
		{"bmp", false},
		{"tiff", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if isSupportedFormat(tt.format) != tt.supported {
				t.Errorf("isSupportedFormat(%q) = %v, want %v", tt.format, !tt.supported, tt.supported)
			}
		})
	}
}
