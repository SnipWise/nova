package display

import (
	"fmt"
	"regexp"
	"strings"
)

// processCodeBlockDelimiter handles the opening or closing ``` fence.
func (p *MarkdownChunkParser) processCodeBlockDelimiter(trimmed string) {
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
}

// processCodeLine renders one line inside a code block.
func (p *MarkdownChunkParser) processCodeLine(line string) {
	codeLine := strings.TrimRight(line, " \t")
	if strings.TrimSpace(line) == "" {
		p.eraseAndPrint(fmt.Sprintf("%s│%s\n", ColorGray, ColorReset))
	} else {
		p.eraseAndPrint(fmt.Sprintf("%s│ %s%s\n", ColorGray, codeLine, ColorReset))
	}
}

// processHeaderChunk renders a markdown header with inline formatting.
// Skips headers with no text after the # prefix.
func (p *MarkdownChunkParser) processHeaderChunk(trimmed string) {
	level := 0
	for _, ch := range trimmed {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	headerText := strings.TrimSpace(trimmed[level:])
	if headerText == "" {
		return
	}
	headerText = formatInlineMarkdown(headerText)
	switch level {
	case 1:
		p.eraseAndPrint(fmt.Sprintf(fmtBlockHeaderNL, ColorBold, ColorBrightCyan, headerText, ColorReset))
		fmt.Println(strings.Repeat("═", len(headerText)))
	case 2:
		p.eraseAndPrint(fmt.Sprintf(fmtBlockHeaderNL, ColorBold, ColorBrightBlue, headerText, ColorReset))
		fmt.Println(strings.Repeat("─", len(headerText)))
	case 3:
		p.eraseAndPrint(fmt.Sprintf(fmtBlockHeaderNL, ColorBold, ColorCyan, headerText, ColorReset))
	default:
		p.eraseAndPrint(fmt.Sprintf("\n%s%s%s\n", ColorBrightCyan, headerText, ColorReset))
	}
}

// processBlockquoteChunk renders a blockquote line.
func (p *MarkdownChunkParser) processBlockquoteChunk(trimmed string) {
	quoteText := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
	quoteText = formatInlineMarkdown(quoteText)
	p.eraseAndPrint(fmt.Sprintf("%s│ %s%s\n", ColorGray, quoteText, ColorReset))
}

// processUnorderedListItemChunk renders an unordered list bullet item.
func (p *MarkdownChunkParser) processUnorderedListItemChunk(line, trimmed string) {
	listLevel := countLeadingSpaces(line) / 2
	listText := regexp.MustCompile(`^[-*+]\s+`).ReplaceAllString(trimmed, "")
	listText = formatInlineMarkdown(listText)
	indent := strings.Repeat("  ", listLevel)
	p.eraseAndPrint(fmt.Sprintf("%s%s%s %s%s\n", indent, ColorBrightYellow, SymbolBullet, listText, ColorReset))
	p.inList = true
}

// processOrderedListItemChunk renders a numbered list item.
func (p *MarkdownChunkParser) processOrderedListItemChunk(line, trimmed string) {
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
}

// processTaskListItemChunk renders a task list item with a check or empty box.
func (p *MarkdownChunkParser) processTaskListItemChunk(trimmed string) {
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
}

// processParagraph renders a regular paragraph line with inline formatting.
// Respects alreadyPrinted (set during streaming) and inList indentation.
func (p *MarkdownChunkParser) processParagraph(trimmed string) {
	formatted := formatInlineMarkdown(trimmed)
	if !p.alreadyPrinted {
		if p.inList {
			p.eraseAndPrint(fmt.Sprintf("  %s\n", formatted))
		} else {
			p.eraseAndPrint(fmt.Sprintf("%s\n", formatted))
		}
	} else {
		fmt.Println()
	}
}
