package display

import (
	"fmt"
	"regexp"
	"strings"
)

// MarkdownChunkParser holds state for streaming markdown parsing
type MarkdownChunkParser struct {
	buffer         strings.Builder
	inCodeBlock    bool
	codeBlockLang  string
	lineBuffer     string
	lastWasNewline bool
	inList         bool
	alreadyPrinted bool // Track if current incomplete line was already printed
	// Note: pendingChars, lastFormatted, lastRawLine, lastDisplayedLen removed in Option 1 fix
}

// NewMarkdownChunkParser creates a new streaming markdown parser
func NewMarkdownChunkParser() *MarkdownChunkParser {
	return &MarkdownChunkParser{
		buffer:         strings.Builder{},
		lastWasNewline: true,
	}
}

// MarkdownChunk processes and displays a chunk of markdown text with streaming support
func MarkdownChunk(parser *MarkdownChunkParser, chunk string) {
	if parser == nil {
		return
	}

	parser.buffer.WriteString(chunk)

	// Split on newlines to process complete lines only
	content := parser.buffer.String()
	lines := strings.Split(content, "\n")

	// Process all complete lines (all except the last which might be incomplete)
	for i := 0; i < len(lines)-1; i++ {
		parser.lineBuffer = lines[i]
		parser.processLine()
		parser.lineBuffer = ""
		parser.alreadyPrinted = false // Reset for next line
	}

	// Keep the last incomplete line in the buffer
	parser.buffer.Reset()
	lastLine := lines[len(lines)-1]
	parser.buffer.WriteString(lastLine)
	parser.lineBuffer = lastLine

	// For incomplete lines that don't look like markdown structures,
	// print them directly for real-time streaming
	if lastLine != "" && !parser.inCodeBlock {
		trimmed := strings.TrimSpace(lastLine)
		// Check if this looks like a markdown structure that needs a complete line
		isMarkdownStructure := strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "```") ||
			strings.HasPrefix(trimmed, ">") ||
			strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasPrefix(trimmed, "+") ||
			regexp.MustCompile(`^\d+\.`).MatchString(trimmed) ||
			regexp.MustCompile(`^(\*\*\*+|---+|___+)`).MatchString(trimmed)

		// Also check if this COULD BECOME a markdown structure
		// (e.g., "1" could become "1. " for ordered list)
		couldBecomeMarkdown := regexp.MustCompile(`^\d+$`).MatchString(trimmed) || // Just digits (could become "1.")
			trimmed == "" // Empty lines should wait for next chunk

		// If it's not a markdown structure and won't become one, print the new chunk directly
		if !isMarkdownStructure && !couldBecomeMarkdown {
			fmt.Print(chunk)
			parser.alreadyPrinted = true
		}
	}
}

// processLine processes a complete line of markdown.
// Each branch delegates to a method in markdown.chunk.blocks.go.
func (p *MarkdownChunkParser) processLine() {
	line := p.lineBuffer
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "```") {
		p.processCodeBlockDelimiter(trimmed)
		return
	}

	if p.inCodeBlock {
		p.processCodeLine(line)
		return
	}

	if trimmed == "" {
		fmt.Println()
		p.inList = false
		return
	}

	if strings.HasPrefix(trimmed, "#") {
		p.processHeaderChunk(trimmed)
		return
	}

	if matched, _ := regexp.MatchString(`^(\*\*\*+|---+|___+)$`, trimmed); matched {
		p.eraseAndPrint(strings.Repeat("─", 80) + "\n")
		return
	}

	if strings.HasPrefix(trimmed, ">") {
		p.processBlockquoteChunk(trimmed)
		return
	}

	if matched, _ := regexp.MatchString(`^[-*+]\s`, trimmed); matched {
		p.processUnorderedListItemChunk(line, trimmed)
		return
	}

	if matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed); matched {
		p.processOrderedListItemChunk(line, trimmed)
		return
	}

	if matched, _ := regexp.MatchString(`^[-*+]\s+\[[ xX]\]`, trimmed); matched {
		p.processTaskListItemChunk(trimmed)
		return
	}

	p.processParagraph(trimmed)
}

// tryProcessInline - Not used in Option 1 fix (characters are displayed directly in MarkdownChunk)
func (p *MarkdownChunkParser) tryProcessInline() {
	// Option 1 fix: Do nothing - characters are already displayed in MarkdownChunk
	// This function is kept for compatibility but no longer performs any logic
	return
}

// eraseAndPrint prints the formatted version (line already cleared in processLine) - Option 1 fix
func (p *MarkdownChunkParser) eraseAndPrint(formatted string) {
	// Just print the formatted content (line already cleared in processLine)
	fmt.Print(formatted)
}

// Flush processes any remaining buffered content
func (p *MarkdownChunkParser) Flush() {
	if p.lineBuffer != "" {
		p.processLine()
	}
}

// Reset resets the parser state
func (p *MarkdownChunkParser) Reset() {
	p.buffer.Reset()
	p.lineBuffer = ""
	p.inCodeBlock = false
	p.codeBlockLang = ""
	p.lastWasNewline = true
	p.inList = false
	p.alreadyPrinted = false
	// Note: pendingChars, lastFormatted, lastRawLine, lastDisplayedLen removed in Option 1 fix
}
