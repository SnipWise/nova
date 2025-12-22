package chunks

import (
	"regexp"
	"strings"
)

// MarkdownChunk represents a parsed markdown section with hierarchical context
type MarkdownChunk struct {
	Header         string
	Content        string
	Level          int
	Prefix         string
	ParentLevel    int
	ParentHeader   string
	ParentPrefix   string
	Hierarchy        string
	SimpleMetaData string                 // Additional metadata if needed
	Metadata       map[string]interface{} // additional metadata
	KeyWords       []string               // Keywords that could be extracted from the content
}

// ParseMarkdownHierarchy parses the given markdown content and returns a slice of Chunk structs.
// Each Chunk represents a header and its associated content, along with its hierarchical lineage.
//
// The function processes the markdown content line by line, identifying headers and their levels
// using a regular expression. It then collects the content associated with each header and
// determines the parent header to build the hierarchical structure.
//
// Parameters:
//   - content: A string containing the markdown content to be parsed.
//
// Returns:
//   - A slice of Chunk structs, each representing a header and its associated content, along with
//     its hierarchical lineage.

//  ParseMarkdownHierarchy parses the given markdown content and returns a slice of MarkdownChunk structs preserving the hierarchical context
func ParseMarkdownHierarchy(content string) []MarkdownChunk {
	lines := strings.Split(content, "\n")
	var chunks []MarkdownChunk
	var stack []MarkdownChunk

	headerRegex := regexp.MustCompile(`^(#+)\s+(.*)$`)


	for i := range lines {
		line := lines[i]
		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			header := matches[2]
			prefix := matches[1]

			// Find content for this header
			contentLines := []string{}
			for j := i + 1; j < len(lines); j++ {
				if headerRegex.MatchString(lines[j]) {
					break
				}
				contentLines = append(contentLines, lines[j])
			}
			content := strings.Join(contentLines, "\n")

			// Determine parent header
			var parent MarkdownChunk
			for len(stack) > 0 && stack[len(stack)-1].Level >= level {
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 {
				parent = stack[len(stack)-1]
			}

			// Build hierarchy
			hierarchy := buildHierarchy(stack, header)

			chunk := MarkdownChunk{
				Level:        level,
				Prefix:       prefix,
				Header:       header,
				Content:      strings.TrimSpace(content),
				ParentPrefix: parent.Prefix,
				ParentLevel:  parent.Level,
				ParentHeader: parent.Header,
				Hierarchy:      hierarchy,
			}
			//if chunk.Content != "" {
			chunks = append(chunks, chunk)
			stack = append(stack, chunk)
			//}
		}
	}

	return chunks
}

// buildHierarchy creates a hierarchical path string from the header stack
func buildHierarchy(stack []MarkdownChunk, currentHeader string) string {
	var hierarchy []string
	for _, chunk := range stack {
		hierarchy = append(hierarchy, chunk.Header)
	}
	hierarchy = append(hierarchy, currentHeader)
	return strings.Join(hierarchy, " > ")
}


// ChunkWithMarkdownHierarchy processes markdown content into formatted chunks with hierarchical context
func ChunkWithMarkdownHierarchy(content string) []string {
	// Parse the markdown content and return the chunks with hierarchy
	var chunks []string
	markdownChunks := ParseMarkdownHierarchy(content)
	for _, chunk := range markdownChunks {
		chunkContent := "TITLE: " + chunk.Prefix + " " + chunk.Header + "\n" +
			"HIERARCHY: " + chunk.Hierarchy + "\n" +
			"CONTENT: " + chunk.Content
		chunks = append(chunks, chunkContent)
	}
	return chunks
}