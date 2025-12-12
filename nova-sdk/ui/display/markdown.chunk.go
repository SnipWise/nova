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
	}

	// Keep the last incomplete line in the buffer
	parser.buffer.Reset()
	parser.buffer.WriteString(lines[len(lines)-1])
	parser.lineBuffer = lines[len(lines)-1]
}

/* func MarkdownChunk(parser *MarkdownChunkParser, chunk string) {
	if parser == nil {
		return
	}

	for _, ch := range chunk {
		parser.buffer.WriteRune(ch)
		parser.lineBuffer += string(ch)

		if ch == '\n' {
			// Process complete line
			parser.processLine()
			parser.lineBuffer = ""
			parser.lastWasNewline = true
		} else {
			parser.lastWasNewline = false

			// Try to process inline markdown as we type
			parser.tryProcessInline()
		}
	}
}
*/
// processLine processes a complete line of markdown
func (p *MarkdownChunkParser) processLine() {
	line := p.lineBuffer
	trimmed := strings.TrimSpace(line)

	// Handle code block delimiters
	if strings.HasPrefix(trimmed, "```") {
		p.inCodeBlock = !p.inCodeBlock
		if p.inCodeBlock {
			p.codeBlockLang = strings.TrimPrefix(trimmed, "```")
			if p.codeBlockLang != "" {
				p.eraseAndPrint(fmt.Sprintf("%s%s┌─ Code: %s%s\n", ColorGray, ColorDim, p.codeBlockLang, ColorReset))
			} else {
				p.eraseAndPrint(fmt.Sprintf("%s%s┌─ Code%s\n", ColorGray, ColorDim, ColorReset))
			}
		} else {
			p.eraseAndPrint(fmt.Sprintf("%s%s└─%s\n", ColorGray, ColorDim, ColorReset))
			p.codeBlockLang = ""
		}
		return
	}

	// Inside code block
	if p.inCodeBlock {
		// Use original line to preserve indentation, but trim trailing spaces only
		codeLine := strings.TrimRight(line, " \t")
		// Check if the line is truly empty (not just whitespace)
		if strings.TrimSpace(line) == "" {
			p.eraseAndPrint(fmt.Sprintf("%s│%s\n", ColorGray, ColorReset))

		} else {
			p.eraseAndPrint(fmt.Sprintf("%s│ %s%s\n", ColorGray, codeLine, ColorReset))
		}
		return
	}

	// Empty line
	if trimmed == "" {
		fmt.Println()
		p.inList = false
		return
	}

	// Headers
	if strings.HasPrefix(trimmed, "#") {
		level := 0
		for _, ch := range trimmed {
			if ch == '#' {
				level++
			} else {
				break
			}
		}
		headerText := strings.TrimSpace(trimmed[level:])

		// Only treat as header if there's actual content after the #
		// If empty, skip it (might be malformed or content on next line)
		if headerText == "" {
			// Don't display empty headers - just skip this line
			return
		}

		// Apply inline formatting to header text
		headerText = formatInlineMarkdown(headerText)

		switch level {
		case 1:
			p.eraseAndPrint(fmt.Sprintf("\n%s%s%s%s\n", ColorBold, ColorBrightCyan, headerText, ColorReset))
			fmt.Println(strings.Repeat("═", len(headerText)))
		case 2:
			p.eraseAndPrint(fmt.Sprintf("\n%s%s%s%s\n", ColorBold, ColorBrightBlue, headerText, ColorReset))
			fmt.Println(strings.Repeat("─", len(headerText)))
		case 3:
			p.eraseAndPrint(fmt.Sprintf("\n%s%s%s%s\n", ColorBold, ColorCyan, headerText, ColorReset))
		default:
			p.eraseAndPrint(fmt.Sprintf("\n%s%s%s\n", ColorBrightCyan, headerText, ColorReset))
		}
		return
	}

	// Horizontal rule
	if matched, _ := regexp.MatchString(`^(\*\*\*+|---+|___+)$`, trimmed); matched {
		p.eraseAndPrint(strings.Repeat("─", 80) + "\n")
		return
	}

	// Blockquotes
	if strings.HasPrefix(trimmed, ">") {
		quoteText := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
		quoteText = formatInlineMarkdown(quoteText)
		p.eraseAndPrint(fmt.Sprintf("%s│ %s%s\n", ColorGray, quoteText, ColorReset))
		return
	}

	// Unordered lists
	if matched, _ := regexp.MatchString(`^[-*+]\s`, trimmed); matched {
		listLevel := countLeadingSpaces(line) / 2
		listText := regexp.MustCompile(`^[-*+]\s+`).ReplaceAllString(trimmed, "")
		listText = formatInlineMarkdown(listText)
		indent := strings.Repeat("  ", listLevel)
		p.eraseAndPrint(fmt.Sprintf("%s%s%s %s%s\n", indent, ColorBrightYellow, SymbolBullet, listText, ColorReset))
		p.inList = true
		return
	}

	// Ordered lists
	if matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed); matched {
		listLevel := countLeadingSpaces(line) / 2
		re := regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) >= 3 {
			num := matches[1]
			listText := formatInlineMarkdown(matches[2])
			indent := strings.Repeat("  ", listLevel)
			p.eraseAndPrint(fmt.Sprintf("%s%s%s.%s %s%s\n", indent, ColorBrightYellow, num, ColorReset, listText, ColorReset))
			p.inList = true
		}
		return
	}

	// Task lists
	if matched, _ := regexp.MatchString(`^[-*+]\s+\[[ xX]\]`, trimmed); matched {
		re := regexp.MustCompile(`^[-*+]\s+\[([ xX])\]\s+(.+)$`)
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) >= 3 {
			checked := matches[1] != " "
			taskText := formatInlineMarkdown(matches[2])
			if checked {
				p.eraseAndPrint(fmt.Sprintf("%s%s %s%s\n", ColorGreen, SymbolCheck, taskText, ColorReset))
			} else {
				p.eraseAndPrint(fmt.Sprintf("%s☐ %s%s\n", ColorGray, taskText, ColorReset))
			}
		}
		return
	}

	// Regular paragraph with inline formatting
	formatted := formatInlineMarkdown(trimmed)

	// Always use eraseAndPrint to avoid duplication issues
	// This ensures we always display the complete, properly formatted line
	if p.inList {
		p.eraseAndPrint(fmt.Sprintf("  %s\n", formatted))
	} else {
		p.eraseAndPrint(fmt.Sprintf("%s\n", formatted))
	}
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
	// Note: pendingChars, lastFormatted, lastRawLine, lastDisplayedLen removed in Option 1 fix
}
