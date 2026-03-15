package display

import (
	"fmt"
	"regexp"
	"strings"
)

// renderCodeBlockBorder prints the opening or closing fence of a code block.
// Returns the updated inCodeBlock state.
func renderCodeBlockBorder(trimmed string, inCodeBlock bool) bool {
	inCodeBlock = !inCodeBlock
	if inCodeBlock {
		lang := strings.TrimPrefix(trimmed, "```")
		if lang != "" {
			fmt.Printf("%s%s┌─ Code: %s%s\n", ColorGray, ColorDim, lang, ColorReset)
		} else {
			fmt.Printf("%s%s┌─ Code%s\n", ColorGray, ColorDim, ColorReset)
		}
	} else {
		fmt.Printf("%s%s└─%s\n", ColorGray, ColorDim, ColorReset)
	}
	return inCodeBlock
}

// renderCodeLine prints a single line inside a code block.
func renderCodeLine(line string) {
	codeLine := strings.TrimRight(line, " \t")
	if codeLine == "" {
		fmt.Printf("%s│%s\n", ColorGray, ColorReset)
	} else {
		fmt.Printf("%s│ %s%s\n", ColorGray, codeLine, ColorReset)
	}
}

// renderHeader prints a markdown header (H1–H4+).
func renderHeader(trimmed string) {
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
		fmt.Printf(fmtBlockHeaderNL, ColorBold, ColorBrightCyan, headerText, ColorReset)
		fmt.Println(strings.Repeat("═", len(headerText)))
	case 2:
		fmt.Printf(fmtBlockHeaderNL, ColorBold, ColorBrightBlue, headerText, ColorReset)
		fmt.Println(strings.Repeat("─", len(headerText)))
	case 3:
		fmt.Printf(fmtBlockHeaderNL, ColorBold, ColorCyan, headerText, ColorReset)
	default:
		fmt.Printf("\n%s%s%s\n", ColorBrightCyan, headerText, ColorReset)
	}
}

// renderBlockquote prints a blockquote line.
func renderBlockquote(trimmed string) {
	quoteText := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
	quoteText = formatInlineMarkdown(quoteText)
	fmt.Printf("%s│ %s%s\n", ColorGray, quoteText, ColorReset)
}

// renderUnorderedListItem prints an unordered list bullet item.
func renderUnorderedListItem(line, trimmed string) {
	listLevel := countLeadingSpaces(line) / 2
	listText := regexp.MustCompile(`^[-*+]\s+`).ReplaceAllString(trimmed, "")
	listText = formatInlineMarkdown(listText)
	indent := strings.Repeat("  ", listLevel)
	fmt.Printf("%s%s%s %s%s\n", indent, ColorBrightYellow, SymbolBullet, listText, ColorReset)
}

// renderOrderedListItem prints a numbered list item.
func renderOrderedListItem(line, trimmed string) {
	listLevel := countLeadingSpaces(line) / 2
	re := regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
	matches := re.FindStringSubmatch(trimmed)
	if len(matches) >= 3 {
		num := matches[1]
		listText := formatInlineMarkdown(matches[2])
		indent := strings.Repeat("  ", listLevel)
		fmt.Printf("%s%s%s.%s %s%s\n", indent, ColorBrightYellow, num, ColorReset, listText, ColorReset)
	}
}

// renderTaskListItem prints a task list item with a check or empty box.
func renderTaskListItem(trimmed string) {
	re := regexp.MustCompile(`^[-*+]\s+\[([ xX])\]\s+(.+)$`)
	matches := re.FindStringSubmatch(trimmed)
	if len(matches) >= 3 {
		checked := matches[1] != " "
		taskText := formatInlineMarkdown(matches[2])
		if checked {
			fmt.Printf("%s%s %s%s\n", ColorGreen, SymbolCheck, taskText, ColorReset)
		} else {
			fmt.Printf("%s☐ %s%s\n", ColorGray, taskText, ColorReset)
		}
	}
}

// countLeadingSpaces counts the number of leading spaces in a line.
func countLeadingSpaces(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

// renderMarkdownLine processes one line of markdown and prints it.
// Returns the updated inCodeBlock and inList state.
func renderMarkdownLine(line string, inCodeBlock bool, inList bool) (bool, bool) {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "```") {
		inCodeBlock = renderCodeBlockBorder(trimmed, inCodeBlock)
		return inCodeBlock, inList
	}

	if inCodeBlock {
		renderCodeLine(line)
		return inCodeBlock, inList
	}

	if trimmed == "" {
		fmt.Println()
		return inCodeBlock, false
	}

	if strings.HasPrefix(trimmed, "#") {
		renderHeader(trimmed)
		return inCodeBlock, inList
	}

	if matched, _ := regexp.MatchString(`^(\*\*\*+|---+|___+)$`, trimmed); matched {
		fmt.Println(strings.Repeat("─", 80))
		return inCodeBlock, inList
	}

	if strings.HasPrefix(trimmed, ">") {
		renderBlockquote(trimmed)
		return inCodeBlock, inList
	}

	if matched, _ := regexp.MatchString(`^[-*+]\s`, trimmed); matched {
		renderUnorderedListItem(line, trimmed)
		return inCodeBlock, true
	}

	if matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed); matched {
		renderOrderedListItem(line, trimmed)
		return inCodeBlock, true
	}

	if matched, _ := regexp.MatchString(`^[-*+]\s+\[[ xX]\]`, trimmed); matched {
		renderTaskListItem(trimmed)
		return inCodeBlock, inList
	}

	formatted := formatInlineMarkdown(trimmed)
	if inList {
		fmt.Printf("  %s\n", formatted)
	} else {
		fmt.Println(formatted)
	}
	return inCodeBlock, inList
}
