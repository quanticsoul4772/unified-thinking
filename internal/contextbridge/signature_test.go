package contextbridge

import (
	"testing"
)

func TestSimpleExtractor_Extract(t *testing.T) {
	extractor := NewSimpleExtractor()

	tests := []struct {
		name     string
		text     string
		minCount int
		contains []string
		excludes []string
	}{
		{
			name:     "basic extraction",
			text:     "How to optimize database queries for performance",
			minCount: 3,
			contains: []string{"optimize", "database", "queries", "performance"},
			excludes: []string{"to", "for", "how"},
		},
		{
			name:     "removes stop words",
			text:     "The quick brown fox jumps over the lazy dog",
			minCount: 4,
			contains: []string{"quick", "brown", "jumps", "lazy"},
			excludes: []string{"the", "over"},
		},
		{
			name:     "handles punctuation",
			text:     "Hello, world! How are you? Fine, thanks.",
			minCount: 2,
			contains: []string{"hello", "world", "fine", "thanks"},
			excludes: []string{",", "!", "?"},
		},
		{
			name:     "removes short words",
			text:     "I am a go developer",
			minCount: 1,
			contains: []string{"developer"},
			excludes: []string{"go", "am"},
		},
		{
			name:     "deduplicates",
			text:     "database database database optimization",
			minCount: 2,
			contains: []string{"database", "optimization"},
		},
		{
			name:     "empty input",
			text:     "",
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			concepts := extractor.Extract(tt.text)

			if len(concepts) < tt.minCount {
				t.Errorf("Extract() returned %d concepts, want at least %d", len(concepts), tt.minCount)
			}

			conceptSet := make(map[string]bool)
			for _, c := range concepts {
				conceptSet[c] = true
			}

			for _, want := range tt.contains {
				if !conceptSet[want] {
					t.Errorf("Extract() missing expected concept %q, got %v", want, concepts)
				}
			}

			for _, exclude := range tt.excludes {
				if conceptSet[exclude] {
					t.Errorf("Extract() should not contain %q, got %v", exclude, concepts)
				}
			}
		})
	}
}

func TestExtractSignature(t *testing.T) {
	extractor := NewSimpleExtractor()

	tests := []struct {
		name       string
		toolName   string
		params     map[string]interface{}
		wantNil    bool
		wantDomain string
	}{
		{
			name:     "basic signature",
			toolName: "think",
			params: map[string]interface{}{
				"content": "How to optimize database queries for performance",
				"mode":    "linear",
			},
			wantNil: false,
		},
		{
			name:     "with domain",
			toolName: "think",
			params: map[string]interface{}{
				"content": "Best practices for API design",
				"domain":  "engineering",
			},
			wantNil:    false,
			wantDomain: "engineering",
		},
		{
			name:     "description parameter",
			toolName: "decompose-problem",
			params: map[string]interface{}{
				"description": "Complex system architecture",
			},
			wantNil: false,
		},
		{
			name:     "missing content",
			toolName: "think",
			params: map[string]interface{}{
				"mode": "linear",
			},
			wantNil: true,
		},
		{
			name:     "empty content",
			toolName: "think",
			params: map[string]interface{}{
				"content": "",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := ExtractSignature(tt.toolName, tt.params, extractor)

			if err != nil {
				t.Errorf("ExtractSignature() error = %v", err)
				return
			}

			if tt.wantNil {
				if sig != nil {
					t.Errorf("ExtractSignature() = %v, want nil", sig)
				}
				return
			}

			if sig == nil {
				t.Error("ExtractSignature() = nil, want signature")
				return
			}

			if sig.Fingerprint == "" {
				t.Error("ExtractSignature() fingerprint is empty")
			}

			if len(sig.ToolSequence) == 0 {
				t.Error("ExtractSignature() tool sequence is empty")
			}

			if sig.ToolSequence[0] != tt.toolName {
				t.Errorf("ExtractSignature() tool sequence = %v, want [%s]", sig.ToolSequence, tt.toolName)
			}

			if tt.wantDomain != "" && sig.Domain != tt.wantDomain {
				t.Errorf("ExtractSignature() domain = %v, want %v", sig.Domain, tt.wantDomain)
			}

			if sig.Complexity <= 0 || sig.Complexity > 1 {
				t.Errorf("ExtractSignature() complexity = %v, want between 0 and 1", sig.Complexity)
			}
		})
	}
}

func TestExtractSignature_Fingerprint_Deterministic(t *testing.T) {
	extractor := NewSimpleExtractor()

	params := map[string]interface{}{
		"content": "Test content for fingerprinting",
	}

	sig1, _ := ExtractSignature("think", params, extractor)
	sig2, _ := ExtractSignature("think", params, extractor)

	if sig1.Fingerprint != sig2.Fingerprint {
		t.Errorf("Fingerprints should be deterministic: %s != %s", sig1.Fingerprint, sig2.Fingerprint)
	}
}

func TestExtractSignature_Fingerprint_CaseInsensitive(t *testing.T) {
	extractor := NewSimpleExtractor()

	params1 := map[string]interface{}{
		"content": "Test Content",
	}
	params2 := map[string]interface{}{
		"content": "test content",
	}

	sig1, _ := ExtractSignature("think", params1, extractor)
	sig2, _ := ExtractSignature("think", params2, extractor)

	if sig1.Fingerprint != sig2.Fingerprint {
		t.Errorf("Fingerprints should be case-insensitive: %s != %s", sig1.Fingerprint, sig2.Fingerprint)
	}
}
