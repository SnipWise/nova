package display

import (
	"fmt"
	"regexp"
	"strings"
)

// MarkdownChunkParser holds state for streaming markdown parsing
type MarkdownChunkParser struct {
	buffer           strings.Builder
	inCodeBlock      bool
	codeBlockLang    string
	lineBuffer       string
	lastWasNewline   bool
	inList           bool
	pendingChars     int    // Number of characters pending colorization
	lastFormatted    string // Last formatted output to prevent duplicate rendering
	lastRawLine      string // Last raw line that was formatted
	lastDisplayedLen int    // Length of the last displayed line (raw)

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

	// Reset state when processing a complete line
	p.lastFormatted = ""
	p.lastRawLine = ""
	p.lastDisplayedLen = 0

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
		p.eraseAndPrint(fmt.Sprintf("%s│ %s%s\n", ColorGray, trimmed, ColorReset))
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
	if p.inList {
		p.eraseAndPrint(fmt.Sprintf("  %s\n", formatted))
	} else {
		p.eraseAndPrint(fmt.Sprintf("%s\n", formatted))
	}
}

// tryProcessInline attempts to colorize inline markdown patterns as they're typed
func (p *MarkdownChunkParser) tryProcessInline() {
	line := p.lineBuffer

	// Don't process inside code blocks
	if p.inCodeBlock {
		fmt.Print(string(line[len(line)-1]))
		p.lastDisplayedLen++
		return
	}

	// Check if we have complete inline markdown patterns
	hasPattern := false

	// Check for complete patterns
	if regexp.MustCompile(`\*\*[^*]+\*\*`).MatchString(line) ||
	   regexp.MustCompile(`__[^_]+__`).MatchString(line) ||
	   regexp.MustCompile(`\*[^*]+\*`).MatchString(line) ||
	   regexp.MustCompile(`_[^_]+_`).MatchString(line) ||
	   regexp.MustCompile("`[^`]+`").MatchString(line) {
		hasPattern = true
	}

	if hasPattern {
		// Format the line
		formatted := formatInlineMarkdown(line)

		// Check if this is a new character being added after the pattern
		currentLen := len(line)
		if currentLen > p.lastDisplayedLen && p.lastDisplayedLen > 0 {
			// We've already displayed part of this line with formatting
			// Just append the new character without redrawing everything
			lastChar := string(line[len(line)-1])
			fmt.Print(lastChar)
			p.lastDisplayedLen = currentLen
		} else {
			// First time formatting this line, or pattern just completed
			// Erase the current line and print the formatted version
			fmt.Print("\r\033[K" + formatted)
			p.lastFormatted = formatted
			p.lastRawLine = line
			p.lastDisplayedLen = currentLen
		}
	} else {
		// No complete patterns yet, just print the new character
		lastChar := string(line[len(line)-1])
		fmt.Print(lastChar)
		p.lastFormatted = ""
		p.lastRawLine = ""
		p.lastDisplayedLen = len(line)
	}
}

// eraseAndPrint erases the pending characters and prints the formatted version
func (p *MarkdownChunkParser) eraseAndPrint(formatted string) {
	lineLen := len(p.lineBuffer)

	if lineLen > 0 {
		// Move cursor back to start of line and clear it
		fmt.Print("\r\033[K")
	}

	// Print the formatted content
	fmt.Print(formatted)

	// Reset pending counter
	p.pendingChars = 0
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
	p.pendingChars = 0
	p.lastFormatted = ""
	p.lastRawLine = ""
	p.lastDisplayedLen = 0
}
