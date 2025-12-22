package stores

import (
	"math"
	"sort"
)

// getTopNVectorRecords returns the top N vector records sorted by highest cosine similarity
func getTopNVectorRecords(records []VectorRecord, max int) []VectorRecord {
	// Sort the records slice in descending order based on CosineDistance
	sort.Slice(records, func(i, j int) bool {
		return records[i].CosineSimilarity > records[j].CosineSimilarity
	})

	// Return the first max records or all if less than three
	if len(records) < max {
		return records
	}
	return records[:max]
}

// --- Cosine similarity ---

// dotProduct calculates the dot product of two equal-length vectors
func dotProduct(v1 []float64, v2 []float64) float64 {
	// Calculate the dot product of two vectors
	sum := 0.0
	for i := range v1 {
		sum += v1[i] * v2[i]
	}
	return sum
}

// cosineSimilarity calculates the cosine similarity between two vectors (0 to 1 scale)
// Returns values between 0 and 1:
//   - 1.0: vectors are identical or perfectly aligned (maximum similarity, close distance)
//   - 0.0: vectors are orthogonal/perpendicular (no similarity)
//   - Values close to 1: vectors are very similar (close distance)
// Note: Cosine similarity measures the angle between vectors, not their magnitude.
// Two vectors can have different lengths but still be considered similar if they point in the same direction.
func cosineSimilarity(v1, v2 []float64) float64 {
	// Calculate the cosine distance between two vectors
	product := dotProduct(v1, v2)

	norm1 := math.Sqrt(dotProduct(v1, v1))
	norm2 := math.Sqrt(dotProduct(v2, v2))
	if norm1 <= 0.0 || norm2 <= 0.0 {
		// Handle potential division by zero
		return 0.0
	}
	return product / (norm1 * norm2)
}
