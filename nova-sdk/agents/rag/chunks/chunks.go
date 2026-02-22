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

// SplitMarkdownBySection splits markdown content by headers of a specific level only.
// It returns sections separated by headers of the specified level.
//
// Parameters:
//   - sectionLevel: The header level to split on (1 for #, 2 for ##, 3 for ###, etc.)
//   - markdown: The markdown content to parse.
//
// Returns:
//   - []string: A slice of strings representing sections split at the specified header level.
//
// Example:
//   markdown := "# Title\nContent\n## Subtitle\nMore\n# Another Title\nEnd"
//   sections := SplitMarkdownBySection(1, markdown)  // Splits only on # headers
//   // Returns: ["# Title\nContent\n## Subtitle\nMore", "# Another Title\nEnd"]
//   sections := SplitMarkdownBySection(2, markdown)  // Splits only on ## headers
//   // Returns: ["# Title\nContent", "## Subtitle\nMore\n# Another Title\nEnd"]
func SplitMarkdownBySection(sectionLevel int, markdown string) []string {
	if markdown == "" || sectionLevel < 1 {
		return []string{}
	}

	// Create regex that matches any markdown header
	headerRegex := regexp.MustCompile(`(?m)^\s*(#+)\s+.*$`)

	// Find all headers with their hash counts
	lines := strings.Split(markdown, "\n")
	var headerPositions []int
	currentPos := 0

	for _, line := range lines {
		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			// Count the number of # characters
			hashCount := len(strings.TrimSpace(matches[1]))

			// Only include headers that match our desired level
			if hashCount == sectionLevel {
				headerPositions = append(headerPositions, currentPos)
			}
		}
		// Add line length + newline character
		currentPos += len(line) + 1
	}

	if len(headerPositions) == 0 {
		// No headers of this level found, return entire content
		return []string{strings.TrimSpace(markdown)}
	}

	var sections []string

	// Handle content before first header of this level
	if headerPositions[0] > 0 {
		preHeader := strings.TrimSpace(markdown[:headerPositions[0]])
		if preHeader != "" {
			sections = append(sections, preHeader)
		}
	}

	// Split by headers of the specified level
	for i, pos := range headerPositions {
		start := pos
		var end int

		if i < len(headerPositions)-1 {
			// Not the last header, end at next header of same level
			end = headerPositions[i+1]
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

// ChunkXML splits XML content into chunks based on a specified target tag.
// Each chunk contains a complete XML element matching the target tag, with all its attributes preserved.
//
// Parameters:
//   - xml: The input XML content to be chunked.
//   - targetTag: The name of the XML tag to extract (e.g., "item").
//
// Returns:
//   - []string: A slice of strings, where each string is a complete XML element.
//
// Example:
//   xml := `<menu>
//     <item id="1">
//       <name>Margherita Pizza</name>
//       <price currency="USD">12.99</price>
//     </item>
//     <item id="2">
//       <name>Caesar Salad</name>
//       <price currency="USD">8.50</price>
//     </item>
//   </menu>`
//
//   chunks := ChunkXML(xml, "item")
//   // Returns: [
//   //   `<item id="1"><name>Margherita Pizza</name><price currency="USD">12.99</price></item>`,
//   //   `<item id="2"><name>Caesar Salad</name><price currency="USD">8.50</price></item>`
//   // ]
func ChunkXML(xml string, targetTag string) []string {
	if xml == "" || targetTag == "" {
		return []string{}
	}

	var chunks []string

	// Build regex to match both self-closing and content tags
	// Pattern explanation:
	// - <targetTag: Opening tag
	// - (?:\s[^>]*)?: Optional attributes (non-capturing group)
	// - (?:/>|>[\s\S]*?</targetTag>): Either self-closing /> or content with closing tag
	pattern := `<` + regexp.QuoteMeta(targetTag) + `(?:\s[^>]*)?(?:/>|>[\s\S]*?</` + regexp.QuoteMeta(targetTag) + `>)`

	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(xml, -1)

	for _, match := range matches {
		trimmed := strings.TrimSpace(match)
		if trimmed != "" {
			chunks = append(chunks, trimmed)
		}
	}

	return chunks
}

// ChunkYAML splits YAML content into chunks based on a specified target key.
// Each chunk contains a complete YAML block starting with the target key and all its nested content.
//
// The target key can be either a simple key (e.g., "snippet") or a list item key (e.g., "- id").
//
// Parameters:
//   - yaml: The input YAML content to be chunked.
//   - targetKey: The key to split on (e.g., "snippet" or "- id").
//
// Returns:
//   - []string: A slice of strings, where each string is a complete YAML block.
//
// Example with simple key:
//
//	yaml := `snippets:
//	  snippet:
//	    name: "Hello"
//	    language: "go"
//	  snippet:
//	    name: "World"
//	    language: "python"`
//
//	chunks := ChunkYAML(yaml, "snippet")
//	// Returns two chunks, one for each snippet block
//
// Example with list item key:
//
//	yaml := `snippets:
//	  - id: 1
//	    name: hello_world
//	    code: |
//	      print("Hello")
//	  - id: 2
//	    name: goodbye
//	    code: |
//	      print("Bye")`
//
//	chunks := ChunkYAML(yaml, "- id")
//	// Returns two chunks, one for each list item
func ChunkYAML(yaml string, targetKey string) []string {
	if yaml == "" || targetKey == "" {
		return []string{}
	}

	lines := strings.Split(yaml, "\n")

	// Build regex to match the target key at a consistent indentation level
	// For "- id", match lines like "  - id: value"
	// For "snippet", match lines like "  snippet:"
	var pattern string
	if strings.HasPrefix(targetKey, "- ") {
		// List item key: "- id" matches "  - id:" or "  - id: value"
		keyPart := regexp.QuoteMeta(strings.TrimPrefix(targetKey, "- "))
		pattern = `^(\s*)- ` + keyPart + `\s*:`
	} else {
		// Simple key: "snippet" matches "  snippet:"
		pattern = `^(\s*)` + regexp.QuoteMeta(targetKey) + `\s*:`
	}
	re := regexp.MustCompile(pattern)

	// Find the indentation level of the first match
	targetIndent := -1
	for _, line := range lines {
		if matches := re.FindStringSubmatch(line); matches != nil {
			targetIndent = len(matches[1])
			break
		}
	}

	if targetIndent < 0 {
		// No match found
		return []string{}
	}

	// Collect chunks by splitting at each occurrence of the target key at the same indentation
	var chunks []string
	var currentChunk []string

	for _, line := range lines {
		if matches := re.FindStringSubmatch(line); matches != nil && len(matches[1]) == targetIndent {
			// Found a new target key at the expected indentation
			if len(currentChunk) > 0 {
				chunks = append(chunks, strings.TrimRight(strings.Join(currentChunk, "\n"), "\n "))
			}
			currentChunk = []string{line}
		} else if len(currentChunk) > 0 {
			// Continue accumulating lines for the current chunk
			currentChunk = append(currentChunk, line)
		}
	}

	// Don't forget the last chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.TrimRight(strings.Join(currentChunk, "\n"), "\n "))
	}

	return chunks
}