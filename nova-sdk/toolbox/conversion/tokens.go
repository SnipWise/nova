package conversion

// EstimateTokenCount calculates the total number of tokens from a string
// Uses a rough approximation: 1 token ≈ 4 characters
func EstimateTokenCount(content string) int {
	return len(content) / 4
}