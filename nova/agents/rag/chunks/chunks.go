package chunks

import (
	"regexp"
	"strings"
)

// ChunkText takes a text string and divides it into chunks of a specified size with a given overlap.
// It returns a slice of strings, where each string represents a chunk of the original text.
//
// Parameters:
//   - text: The input text to be chunked.
//   - chunkSize: The size of each chunk.
//   - overlap: The amount of overlap between consecutive chunks.
//
// Returns:
//   - []string: A slice of strings representing the chunks of the original text.
func ChunkText(text string, chunkSize, overlap int) []string {
	chunks := []string{}
	for start := 0; start < len(text); start += chunkSize - overlap {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
	}
	return chunks
}

// SplitTextWithDelimiter splits the given text using the specified delimiter and returns a slice of strings.
//
// Parameters:
//   - text: The text to be split.
//   - delimiter: The delimiter used to split the text.
//
// Returns:
//   - []string: A slice of strings containing the split parts of the text.
func SplitTextWithDelimiter(text string, delimiter string) []string {
	return strings.Split(text, delimiter)
}



// SplitMarkdownBySections splits markdown content into sections at header boundaries
func SplitMarkdownBySections(markdown string) []string {
	if markdown == "" {
		return []string{}
	}
	
	// Regex to match markdown headers (# ## ### etc. allowing leading whitespace)
	headerRegex := regexp.MustCompile(`(?m)^\s*#+\s+.*$`)
	
	// Find all header positions
	headerMatches := headerRegex.FindAllStringIndex(markdown, -1)
	
	if len(headerMatches) == 0 {
		// No headers found, return the entire content as one section
		return []string{strings.TrimSpace(markdown)}
	}
	
	var sections []string
	
	// Handle content before first header
	if headerMatches[0][0] > 0 {
		preHeader := strings.TrimSpace(markdown[:headerMatches[0][0]])
		if preHeader != "" {
			sections = append(sections, preHeader)
		}
	}
	
	// Split by headers
	for i, match := range headerMatches {
		start := match[0]
		var end int
		
		if i < len(headerMatches)-1 {
			// Not the last header, end at next header
			end = headerMatches[i+1][0]
		} else {
			// Last header, end at document end
			end = len(markdown)
		}
		
		section := strings.TrimSpace(markdown[start:end])
		if section != "" {
			sections = append(sections, section)
		}
	}
	
	return sections
}