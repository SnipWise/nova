package display

import (
	"fmt"
	"regexp"
	"strings"
)

// Markdown renders and prints formatted markdown content with colors
func Markdown(content string) {
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	inList := false
	listLevel := 0

	for i := range lines {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
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
			continue
		}

		if inCodeBlock {
			// Use original line to preserve indentation, but trim trailing spaces only
			codeLine := strings.TrimRight(line, " \t")
			// For empty lines in code blocks, just show the vertical bar
			if codeLine == "" {
				fmt.Printf("%s│%s\n", ColorGray, ColorReset)
			} else {
				fmt.Printf("%s│ %s%s\n", ColorGray, codeLine, ColorReset)
			}
			continue
		}

		// Empty line
		if trimmed == "" {
			fmt.Println()
			inList = false
			continue
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
				fmt.Printf("\n%s%s%s%s\n", ColorBold, ColorBrightCyan, headerText, ColorReset)
				fmt.Println(strings.Repeat("═", len(headerText)))
			case 2:
				fmt.Printf("\n%s%s%s%s\n", ColorBold, ColorBrightBlue, headerText, ColorReset)
				fmt.Println(strings.Repeat("─", len(headerText)))
			case 3:
				fmt.Printf("\n%s%s%s%s\n", ColorBold, ColorCyan, headerText, ColorReset)
			default:
				fmt.Printf("\n%s%s%s\n", ColorBrightCyan, headerText, ColorReset)
			}
			continue
		}

		// Horizontal rule
		if matched, _ := regexp.MatchString(`^(\*\*\*+|---+|___+)$`, trimmed); matched {
			fmt.Println(strings.Repeat("─", 80))
			continue
		}

		// Blockquotes
		if strings.HasPrefix(trimmed, ">") {
			quoteText := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			quoteText = formatInlineMarkdown(quoteText)
			fmt.Printf("%s│ %s%s\n", ColorGray, quoteText, ColorReset)
			continue
		}

		// Unordered lists
		if matched, _ := regexp.MatchString(`^[-*+]\s`, trimmed); matched {
			listLevel = countLeadingSpaces(line) / 2
			listText := regexp.MustCompile(`^[-*+]\s+`).ReplaceAllString(trimmed, "")
			listText = formatInlineMarkdown(listText)
			indent := strings.Repeat("  ", listLevel)
			fmt.Printf("%s%s%s %s%s\n", indent, ColorBrightYellow, SymbolBullet, listText, ColorReset)
			inList = true
			continue
		}

		// Ordered lists
		if matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed); matched {
			listLevel = countLeadingSpaces(line) / 2
			re := regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
			matches := re.FindStringSubmatch(trimmed)
			if len(matches) >= 3 {
				num := matches[1]
				listText := formatInlineMarkdown(matches[2])
				indent := strings.Repeat("  ", listLevel)
				fmt.Printf("%s%s%s.%s %s%s\n", indent, ColorBrightYellow, num, ColorReset, listText, ColorReset)
				inList = true
			}
			continue
		}

		// Task lists
		if matched, _ := regexp.MatchString(`^[-*+]\s+\[[ xX]\]`, trimmed); matched {
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
			continue
		}

		// Regular paragraph
		formatted := formatInlineMarkdown(trimmed)
		if inList {
			fmt.Printf("  %s\n", formatted)
		} else {
			fmt.Println(formatted)
		}
	}
}

// formatInlineMarkdown handles inline markdown formatting
func formatInlineMarkdown(text string) string {
	// Escape sequences for later restoration
	text = strings.ReplaceAll(text, "\\*", "\x00STAR\x00")
	text = strings.ReplaceAll(text, "\\_", "\x00UNDERSCORE\x00")
	text = strings.ReplaceAll(text, "\\`", "\x00BACKTICK\x00")

	// Process in order to avoid conflicts:
	// 1. Bold+Code combination: **`text`** (must be processed first)
	boldCodeRe := regexp.MustCompile(`\*\*\x60([^\x60]+?)\x60\*\*`)
	text = boldCodeRe.ReplaceAllString(text, ColorBold+ColorBrightYellow+BgBlack+"$1"+ColorReset)

	// 2. Bold with ** or __ (process before italic to avoid conflicts)
	boldRe := regexp.MustCompile(`\*\*([^*]+?)\*\*`)
	text = boldRe.ReplaceAllString(text, ColorBold+"$1"+ColorReset)
	boldRe2 := regexp.MustCompile(`__([^_]+?)__`)
	text = boldRe2.ReplaceAllString(text, ColorBold+"$1"+ColorReset)

	// 3. Inline code with ` (process before italic to avoid conflicts)
	codeRe := regexp.MustCompile("`([^`]+)`")
	text = codeRe.ReplaceAllString(text, ColorBrightYellow+BgBlack+"$1"+ColorReset)

	// 4. Italic with * or _ (process after bold and code)
	italicRe := regexp.MustCompile(`\*([^*]+?)\*`)
	text = italicRe.ReplaceAllString(text, ColorItalic+"$1"+ColorReset)
	italicRe2 := regexp.MustCompile(`_([^_]+?)_`)
	text = italicRe2.ReplaceAllString(text, ColorItalic+"$1"+ColorReset)

	// 5. Strikethrough with ~~
	strikeRe := regexp.MustCompile(`~~(.+?)~~`)
	text = strikeRe.ReplaceAllString(text, ColorDim+"$1"+ColorReset)

	// 6. Links [text](url)
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = linkRe.ReplaceAllString(text, ColorBrightCyan+ColorUnderline+"$1"+ColorReset+ColorGray+" ($2)"+ColorReset)

	// 7. Images ![alt](url)
	imageRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	text = imageRe.ReplaceAllString(text, ColorBrightMagenta+"🖼  $1"+ColorReset+ColorGray+" ($2)"+ColorReset)

	// Restore escaped characters
	text = strings.ReplaceAll(text, "\x00STAR\x00", "*")
	text = strings.ReplaceAll(text, "\x00UNDERSCORE\x00", "_")
	text = strings.ReplaceAll(text, "\x00BACKTICK\x00", "`")

	return text
}

// countLeadingSpaces counts leading spaces in a line
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
