package embeddings

import (
	"math"
)

// CosineSimilarity computes cosine similarity between two vectors
// Returns a value between -1 and 1, where 1 means identical direction
func CosineSimilarity(v1, v2 []float32) float64 {
	if len(v1) != len(v2) {
		return 0.0
	}

	var dotProduct, norm1, norm2 float64
	for i := range v1 {
		dotProduct += float64(v1[i] * v2[i])
		norm1 += float64(v1[i] * v1[i])
		norm2 += float64(v2[i] * v2[i])
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// EuclideanDistance computes L2 distance between two vectors
// Lower values mean more similar
func EuclideanDistance(v1, v2 []float32) float64 {
	if len(v1) != len(v2) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range v1 {
		diff := float64(v1[i] - v2[i])
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// DotProduct computes the dot product between two vectors
func DotProduct(v1, v2 []float32) float64 {
	if len(v1) != len(v2) {
		return 0.0
	}

	var product float64
	for i := range v1 {
		product += float64(v1[i] * v2[i])
	}

	return product
}

// NormalizeVector normalizes a vector to unit length
func NormalizeVector(v []float32) []float32 {
	var norm float64
	for _, val := range v {
		norm += float64(val * val)
	}

	if norm == 0 {
		return v
	}

	norm = math.Sqrt(norm)
	normalized := make([]float32, len(v))
	for i, val := range v {
		normalized[i] = float32(float64(val) / norm)
	}

	return normalized
}

// SerializeFloat32 converts a slice of float32 to bytes for storage
func SerializeFloat32(vec []float32) []byte {
	if len(vec) == 0 {
		return nil
	}

	// Each float32 is 4 bytes
	bytes := make([]byte, len(vec)*4)
	for i, v := range vec {
		// Convert float32 to bytes (little-endian)
		bits := math.Float32bits(v)
		bytes[i*4] = byte(bits)
		bytes[i*4+1] = byte(bits >> 8)
		bytes[i*4+2] = byte(bits >> 16)
		bytes[i*4+3] = byte(bits >> 24)
	}

	return bytes
}

// DeserializeFloat32 converts bytes back to a slice of float32
func DeserializeFloat32(bytes []byte) []float32 {
	if len(bytes) == 0 {
		return nil
	}

	// Each float32 is 4 bytes
	vec := make([]float32, len(bytes)/4)
	for i := 0; i < len(vec); i++ {
		// Convert bytes to float32 (little-endian)
		bits := uint32(bytes[i*4]) |
			uint32(bytes[i*4+1])<<8 |
			uint32(bytes[i*4+2])<<16 |
			uint32(bytes[i*4+3])<<24
		vec[i] = math.Float32frombits(bits)
	}

	return vec
}
