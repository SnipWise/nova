package display

import (
	"regexp"
	"strings"
)

// Markdown renders and prints formatted markdown content with colors.
// Per-line rendering is delegated to renderMarkdownLine (markdown.blocks.go).
func Markdown(content string) {
	inCodeBlock, inList := false, false
	for _, line := range strings.Split(content, "\n") {
		inCodeBlock, inList = renderMarkdownLine(line, inCodeBlock, inList)
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
