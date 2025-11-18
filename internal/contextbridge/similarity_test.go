package contextbridge

import (
	"math"
	"testing"
)

func TestWeightedSimilarity_Calculate(t *testing.T) {
	calc := NewDefaultSimilarity()

	tests := []struct {
		name    string
		sig1    *Signature
		sig2    *Signature
		minSim  float64
		maxSim  float64
	}{
		{
			name: "identical signatures",
			sig1: &Signature{
				KeyConcepts:  []string{"database", "optimization", "query"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.6,
			},
			sig2: &Signature{
				KeyConcepts:  []string{"database", "optimization", "query"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.6,
			},
			minSim: 0.95,
			maxSim: 1.0,
		},
		{
			name: "similar concepts different domain",
			sig1: &Signature{
				KeyConcepts:  []string{"database", "optimization", "query"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.6,
			},
			sig2: &Signature{
				KeyConcepts:  []string{"database", "optimization", "query"},
				Domain:       "data-science",
				ToolSequence: []string{"think"},
				Complexity:   0.6,
			},
			minSim: 0.7,
			maxSim: 0.85,
		},
		{
			name: "partial concept overlap",
			sig1: &Signature{
				KeyConcepts:  []string{"database", "optimization", "query"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.6,
			},
			sig2: &Signature{
				KeyConcepts:  []string{"database", "performance", "indexing"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.7,
			},
			minSim: 0.4,
			maxSim: 0.7,
		},
		{
			name: "completely different",
			sig1: &Signature{
				KeyConcepts:  []string{"database", "optimization", "query"},
				Domain:       "engineering",
				ToolSequence: []string{"think"},
				Complexity:   0.6,
			},
			sig2: &Signature{
				KeyConcepts:  []string{"machine", "learning", "model"},
				Domain:       "ai",
				ToolSequence: []string{"decompose-problem"},
				Complexity:   0.4,
			},
			minSim: 0.0,
			maxSim: 0.3,
		},
		{
			name: "empty concepts both",
			sig1: &Signature{
				KeyConcepts:  []string{},
				Domain:       "",
				ToolSequence: []string{},
				Complexity:   0.5,
			},
			sig2: &Signature{
				KeyConcepts:  []string{},
				Domain:       "",
				ToolSequence: []string{},
				Complexity:   0.5,
			},
			minSim: 0.5,
			maxSim: 0.7,
		},
		{
			name: "nil signature",
			sig1: nil,
			sig2: &Signature{
				KeyConcepts: []string{"test"},
			},
			minSim: 0.0,
			maxSim: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := calc.Calculate(tt.sig1, tt.sig2)

			if sim < tt.minSim || sim > tt.maxSim {
				t.Errorf("Calculate() = %v, want between %v and %v", sim, tt.minSim, tt.maxSim)
			}
		})
	}
}

func TestJaccardSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want float64
	}{
		{
			name: "identical sets",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: 1.0,
		},
		{
			name: "no overlap",
			a:    []string{"a", "b", "c"},
			b:    []string{"d", "e", "f"},
			want: 0.0,
		},
		{
			name: "partial overlap",
			a:    []string{"a", "b", "c"},
			b:    []string{"b", "c", "d"},
			want: 0.5, // 2 / 4
		},
		{
			name: "both empty",
			a:    []string{},
			b:    []string{},
			want: 1.0,
		},
		{
			name: "one empty",
			a:    []string{"a", "b"},
			b:    []string{},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jaccardSimilarity(tt.a, tt.b)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("jaccardSimilarity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOverlapRatio(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want float64
	}{
		{
			name: "identical sets",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: 1.0,
		},
		{
			name: "no overlap",
			a:    []string{"a", "b", "c"},
			b:    []string{"d", "e", "f"},
			want: 0.0,
		},
		{
			name: "partial overlap different sizes",
			a:    []string{"a", "b"},
			b:    []string{"a", "b", "c", "d"},
			want: 0.5, // 2 / 4
		},
		{
			name: "one empty",
			a:    []string{"a", "b"},
			b:    []string{},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := overlapRatio(tt.a, tt.b)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("overlapRatio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkWeightedSimilarity_Calculate(b *testing.B) {
	calc := NewDefaultSimilarity()
	sig1 := &Signature{
		KeyConcepts:  []string{"database", "optimization", "query", "performance", "indexing"},
		Domain:       "engineering",
		ToolSequence: []string{"think", "make-decision"},
		Complexity:   0.6,
	}
	sig2 := &Signature{
		KeyConcepts:  []string{"database", "tuning", "performance", "scaling", "caching"},
		Domain:       "engineering",
		ToolSequence: []string{"think", "analyze-perspectives"},
		Complexity:   0.7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.Calculate(sig1, sig2)
	}
}
