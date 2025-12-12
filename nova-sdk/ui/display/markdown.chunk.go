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

	// Note: We don't reset lastFormatted, lastRawLine, lastDisplayedLen here
	// because eraseAndPrint() needs them to properly clear the line
	// They will be reset in eraseAndPrint() after use

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
			// 🤓 Fix by @k33g: Avoid printing trailing spaces in empty code lines
			//p.eraseAndPrint(fmt.Sprintf("%s│ %s%s\n", ColorGray, codeLine, ColorReset))
			p.eraseAndPrint(fmt.Sprintf("%s│ %s%s", ColorGray, codeLine, ColorReset))
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

// tryProcessInline attempts to colorize inline markdown patterns as they're typed
func (p *MarkdownChunkParser) tryProcessInline() {
	line := p.lineBuffer

	// Don't process inside code blocks - wait for newline to display properly
	if p.inCodeBlock {
		// Don't display anything until we have the complete line
		// This prevents issues with line clearing in eraseAndPrint
		return
	}

	// Check if line looks like it might be starting a header or list
	// If so, don't display anything yet - wait for the newline to process properly
	trimmed := strings.TrimSpace(line)

	// Empty header check
	if regexp.MustCompile(`^#{1,6}\s*$`).MatchString(trimmed) {
		// Line is just "#", "##", "###", etc. with optional trailing spaces
		// Don't display - wait to see if content comes on next line
		return
	}

	// Code block check - if line contains ```, don't display inline
	// Let processLine() handle the formatting when newline arrives
	if strings.Contains(trimmed, "```") {
		// Line contains code block delimiter, wait for newline to process
		return
	}

	// List item check - if line starts with list markers, don't display inline
	// Let processLine() handle the formatting when newline arrives
	if regexp.MustCompile(`^[-*+]\s`).MatchString(trimmed) ||
		regexp.MustCompile(`^\d+\.\s`).MatchString(trimmed) ||
		regexp.MustCompile(`^[-*+]\s+\[[ xX]\]`).MatchString(trimmed) {
		// Line starts with list marker, wait for newline to process
		return
	}

	// Check if we're in the middle of building a markdown pattern
	// Count unclosed delimiters
	isBuilding := false

	// Count ** pairs
	boldCount := strings.Count(line, "**")
	if boldCount%2 == 1 {
		isBuilding = true // Odd number means one is unclosed
	}

	// Count ` pairs
	codeCount := strings.Count(line, "`")
	if codeCount%2 == 1 {
		isBuilding = true // Odd number means one is unclosed
	}

	// Count single * (after removing **) for italic
	tempLine := strings.ReplaceAll(line, "**", "")
	singleStarCount := strings.Count(tempLine, "*")
	if singleStarCount%2 == 1 {
		isBuilding = true
	}

	// Check if we have complete inline markdown patterns
	hasPattern := false
	if regexp.MustCompile(`\*\*[^*]+\*\*`).MatchString(line) ||
		regexp.MustCompile(`__[^_]+__`).MatchString(line) ||
		regexp.MustCompile(`\*[^*]+\*`).MatchString(line) ||
		regexp.MustCompile(`_[^_]+_`).MatchString(line) ||
		regexp.MustCompile("`[^`]+`").MatchString(line) {
		hasPattern = true
	}

	if hasPattern {
		// We have at least one complete pattern
		formatted := formatInlineMarkdown(line)

		// Always redraw the whole line to ensure proper formatting
		// This ensures that markdown delimiters (*, **, etc.) are properly hidden
		if formatted != p.lastFormatted {
			fmt.Print("\r\033[K" + formatted)
			p.lastFormatted = formatted
			p.lastRawLine = line
			p.lastDisplayedLen = len(line)
		}
	} else if isBuilding {
		// We're building a pattern but it's not complete yet
		// Don't display anything to avoid showing raw markdown
		p.lastDisplayedLen = len(line)
	} else {
		// No patterns at all, just print the new character normally
		lastChar := string(line[len(line)-1])
		fmt.Print(lastChar)
		p.lastFormatted = ""
		p.lastRawLine = ""
		p.lastDisplayedLen = len(line)
	}
}

// eraseAndPrint erases the pending characters and prints the formatted version
func (p *MarkdownChunkParser) eraseAndPrint(formatted string) {
	// Always clear the line to prevent duplication from tryProcessInline()
	// Use the lastDisplayedLen to know how much to erase if available
	if p.lastDisplayedLen > 0 || len(p.lineBuffer) > 0 {
		// Move cursor back to start of line and clear it
		fmt.Print("\r\033[K")
	}

	// Print the formatted content
	fmt.Print(formatted)

	// Reset tracking variables
	p.pendingChars = 0
	p.lastFormatted = ""
	p.lastRawLine = ""
	p.lastDisplayedLen = 0
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
