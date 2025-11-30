// Package reinforcement provides statistical sampling for Thompson Sampling.
package reinforcement

import (
	"math"
	"math/rand"
)

// SampleBeta samples from Beta(α, β) distribution using relationship to Gamma
//
// Beta(α, β) = X / (X + Y) where X ~ Gamma(α, 1) and Y ~ Gamma(β, 1)
//
// This approach is numerically stable and efficient for all parameter values.
func SampleBeta(alpha, beta float64, rng *rand.Rand) float64 {
	if alpha <= 0 || beta <= 0 {
		// Invalid parameters - return uniform sample
		return rng.Float64()
	}

	x := SampleGamma(alpha, 1.0, rng)
	y := SampleGamma(beta, 1.0, rng)

	// Handle edge case where both are 0
	if x+y == 0 {
		return 0.5
	}

	return x / (x + y)
}

// SampleGamma samples from Gamma(α, β) distribution using Marsaglia-Tsang method
//
// The Marsaglia-Tsang algorithm is efficient and numerically stable for α ≥ 1.
// For α < 1, we use the transformation property: Gamma(α) = Gamma(α+1) * U^(1/α)
// where U ~ Uniform(0,1).
//
// Reference: Marsaglia, G. and Tsang, W.W. (2000). A Simple Method for Generating
// Gamma Variables. ACM Transactions on Mathematical Software, 26(3):363-372.
func SampleGamma(alpha, beta float64, rng *rand.Rand) float64 {
	if alpha >= 1.0 {
		// Marsaglia-Tsang algorithm for α ≥ 1
		d := alpha - 1.0/3.0
		c := 1.0 / math.Sqrt(9.0*d)

		for {
			// Sample from standard normal
			x := rng.NormFloat64()

			// Compute v = (1 + c*x)^3
			v := 1.0 + c*x
			if v <= 0 {
				continue // Reject and resample
			}
			v = v * v * v

			// Acceptance test
			u := rng.Float64()

			// Fast acceptance
			if u < 1.0-0.0331*x*x*x*x {
				return d * v / beta
			}

			// Slow acceptance (logarithmic check)
			if math.Log(u) < 0.5*x*x+d*(1.0-v+math.Log(v)) {
				return d * v / beta
			}

			// Reject and continue
		}
	} else {
		// For α < 1, use transformation:
		// Gamma(α) = Gamma(α+1) * U^(1/α)
		gamma := SampleGamma(alpha+1.0, beta, rng)
		u := rng.Float64()
		return gamma * math.Pow(u, 1.0/alpha)
	}
}

// SampleBetaBatch samples multiple values from Beta(α, β)
// This is useful for vectorized operations
func SampleBetaBatch(alpha, beta float64, n int, rng *rand.Rand) []float64 {
	samples := make([]float64, n)
	for i := 0; i < n; i++ {
		samples[i] = SampleBeta(alpha, beta, rng)
	}
	return samples
}

// BetaMean returns the theoretical mean of Beta(α, β)
// Mean = α / (α + β)
func BetaMean(alpha, beta float64) float64 {
	return alpha / (alpha + beta)
}

// BetaVariance returns the theoretical variance of Beta(α, β)
// Variance = αβ / [(α+β)²(α+β+1)]
func BetaVariance(alpha, beta float64) float64 {
	sum := alpha + beta
	return (alpha * beta) / (sum * sum * (sum + 1))
}

// BetaMode returns the mode of Beta(α, β) distribution
// Mode = (α-1) / (α+β-2) for α,β > 1
// Returns -1 if no mode exists (α or β ≤ 1)
func BetaMode(alpha, beta float64) float64 {
	if alpha <= 1 || beta <= 1 {
		return -1 // No mode exists
	}
	return (alpha - 1) / (alpha + beta - 2)
}
