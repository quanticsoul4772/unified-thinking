package reinforcement

import (
	"math"
	"math/rand"
	"testing"
)

func TestSampleBeta_UniformPrior(t *testing.T) {
	// Beta(1,1) is uniform distribution [0,1]
	rng := rand.New(rand.NewSource(42))
	samples := SampleBetaBatch(1.0, 1.0, 10000, rng)

	// Check mean ≈ 0.5
	mean := computeMean(samples)
	if math.Abs(mean-0.5) > 0.02 {
		t.Errorf("Beta(1,1) mean = %v, want ~0.5", mean)
	}

	// Check variance ≈ 1/12 ≈ 0.0833
	variance := computeVariance(samples, mean)
	expectedVar := BetaVariance(1.0, 1.0)
	if math.Abs(variance-expectedVar) > 0.01 {
		t.Errorf("Beta(1,1) variance = %v, want ~%v", variance, expectedVar)
	}
}

func TestSampleBeta_SkewedDistributions(t *testing.T) {
	tests := []struct {
		name     string
		alpha    float64
		beta     float64
		wantMean float64
	}{
		{
			name:     "skewed right - Beta(5,2)",
			alpha:    5.0,
			beta:     2.0,
			wantMean: 5.0 / 7.0, // ≈ 0.714
		},
		{
			name:     "skewed left - Beta(2,5)",
			alpha:    2.0,
			beta:     5.0,
			wantMean: 2.0 / 7.0, // ≈ 0.286
		},
		{
			name:     "highly confident - Beta(20,2)",
			alpha:    20.0,
			beta:     2.0,
			wantMean: 20.0 / 22.0, // ≈ 0.909
		},
		{
			name:     "low confidence - Beta(2,20)",
			alpha:    2.0,
			beta:     20.0,
			wantMean: 2.0 / 22.0, // ≈ 0.091
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			samples := SampleBetaBatch(tt.alpha, tt.beta, 10000, rng)

			mean := computeMean(samples)
			tolerance := 0.02

			if math.Abs(mean-tt.wantMean) > tolerance {
				t.Errorf("Beta(%v,%v) mean = %v, want %v (tolerance: %v)",
					tt.alpha, tt.beta, mean, tt.wantMean, tolerance)
			}
		})
	}
}

func TestSampleBeta_InvalidParameters(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	// Negative alpha
	sample := SampleBeta(-1.0, 1.0, rng)
	if sample < 0 || sample > 1 {
		t.Errorf("Invalid parameters should return valid probability, got %v", sample)
	}

	// Zero parameters
	sample = SampleBeta(0.0, 1.0, rng)
	if sample < 0 || sample > 1 {
		t.Errorf("Zero alpha should return valid probability, got %v", sample)
	}
}

func TestSampleGamma_AlphaGreaterThanOne(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	alpha := 5.0
	beta := 1.0

	samples := make([]float64, 10000)
	for i := 0; i < len(samples); i++ {
		samples[i] = SampleGamma(alpha, beta, rng)
	}

	// Gamma(α,β) mean = α/β
	mean := computeMean(samples)
	expectedMean := alpha / beta
	if math.Abs(mean-expectedMean) > 0.2 {
		t.Errorf("Gamma(%v,%v) mean = %v, want ~%v", alpha, beta, mean, expectedMean)
	}

	// Gamma(α,β) variance = α/β²
	variance := computeVariance(samples, mean)
	expectedVar := alpha / (beta * beta)
	if math.Abs(variance-expectedVar) > 0.5 {
		t.Errorf("Gamma(%v,%v) variance = %v, want ~%v", alpha, beta, variance, expectedVar)
	}
}

func TestSampleGamma_AlphaLessThanOne(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	alpha := 0.5
	beta := 1.0

	samples := make([]float64, 10000)
	for i := 0; i < len(samples); i++ {
		samples[i] = SampleGamma(alpha, beta, rng)
	}

	// Check mean
	mean := computeMean(samples)
	expectedMean := alpha / beta
	tolerance := 0.1

	if math.Abs(mean-expectedMean) > tolerance {
		t.Errorf("Gamma(%v,%v) mean = %v, want ~%v (tolerance: %v)",
			alpha, beta, mean, expectedMean, tolerance)
	}
}

func TestBetaMean(t *testing.T) {
	tests := []struct {
		alpha float64
		beta  float64
		want  float64
	}{
		{alpha: 1, beta: 1, want: 0.5},
		{alpha: 5, beta: 2, want: 5.0 / 7.0},
		{alpha: 2, beta: 5, want: 2.0 / 7.0},
		{alpha: 10, beta: 10, want: 0.5},
	}

	for _, tt := range tests {
		got := BetaMean(tt.alpha, tt.beta)
		if math.Abs(got-tt.want) > 0.0001 {
			t.Errorf("BetaMean(%v, %v) = %v, want %v", tt.alpha, tt.beta, got, tt.want)
		}
	}
}

func TestBetaVariance(t *testing.T) {
	// Beta(1,1) variance = 1/12
	got := BetaVariance(1.0, 1.0)
	want := 1.0 / 12.0
	if math.Abs(got-want) > 0.0001 {
		t.Errorf("BetaVariance(1,1) = %v, want %v", got, want)
	}
}

func TestBetaMode(t *testing.T) {
	tests := []struct {
		alpha float64
		beta  float64
		want  float64
	}{
		{alpha: 5, beta: 2, want: 4.0 / 5.0}, // (5-1)/(5+2-2) = 0.8
		{alpha: 2, beta: 5, want: 1.0 / 5.0}, // (2-1)/(2+5-2) = 0.2
		{alpha: 1, beta: 1, want: -1},        // No mode
		{alpha: 0.5, beta: 2, want: -1},      // No mode
	}

	for _, tt := range tests {
		got := BetaMode(tt.alpha, tt.beta)
		if tt.want == -1 {
			if got != -1 {
				t.Errorf("BetaMode(%v, %v) = %v, want %v (no mode)", tt.alpha, tt.beta, got, tt.want)
			}
		} else {
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("BetaMode(%v, %v) = %v, want %v", tt.alpha, tt.beta, got, tt.want)
			}
		}
	}
}

// Helper functions for statistical tests

func computeMean(samples []float64) float64 {
	sum := 0.0
	for _, s := range samples {
		sum += s
	}
	return sum / float64(len(samples))
}

func computeVariance(samples []float64, mean float64) float64 {
	sumSq := 0.0
	for _, s := range samples {
		diff := s - mean
		sumSq += diff * diff
	}
	return sumSq / float64(len(samples))
}

// Benchmark Beta sampling performance
func BenchmarkSampleBeta(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = SampleBeta(5.0, 2.0, rng)
	}
}

// Benchmark Gamma sampling performance
func BenchmarkSampleGamma(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = SampleGamma(5.0, 1.0, rng)
	}
}
