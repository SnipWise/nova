package display

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

// captureOutput captures everything written to os.Stdout during f().
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

// stripANSI removes ANSI escape sequences so tests can compare plain text.
var ansiRe = regexp.MustCompile(`\033\[[0-9;?]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// ── countLeadingSpaces ────────────────────────────────────────────────────────

func TestCountLeadingSpaces_None(t *testing.T) {
	if n := countLeadingSpaces("hello"); n != 0 {
		t.Errorf("want 0, got %d", n)
	}
}

func TestCountLeadingSpaces_Two(t *testing.T) {
	if n := countLeadingSpaces("  hello"); n != 2 {
		t.Errorf("want 2, got %d", n)
	}
}

func TestCountLeadingSpaces_Four(t *testing.T) {
	if n := countLeadingSpaces("    hi"); n != 4 {
		t.Errorf("want 4, got %d", n)
	}
}

func TestCountLeadingSpaces_EmptyString(t *testing.T) {
	if n := countLeadingSpaces(""); n != 0 {
		t.Errorf("want 0, got %d", n)
	}
}

func TestCountLeadingSpaces_OnlySpaces(t *testing.T) {
	if n := countLeadingSpaces("   "); n != 3 {
		t.Errorf("want 3, got %d", n)
	}
}

// ── renderCodeBlockBorder ────────────────────────────────────────────────────

func TestRenderCodeBlockBorder_Opening_NoLang(t *testing.T) {
	out := captureOutput(func() {
		renderCodeBlockBorder("```", false)
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "┌─ Code") {
		t.Errorf("expected opening border, got %q", plain)
	}
	if strings.Contains(plain, "Code:") {
		t.Errorf("no language: should not contain 'Code:', got %q", plain)
	}
}

func TestRenderCodeBlockBorder_Opening_WithLang(t *testing.T) {
	out := captureOutput(func() {
		renderCodeBlockBorder("```go", false)
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "Code: go") {
		t.Errorf("expected 'Code: go' in output, got %q", plain)
	}
}

func TestRenderCodeBlockBorder_Closing(t *testing.T) {
	out := captureOutput(func() {
		renderCodeBlockBorder("```", true)
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "└─") {
		t.Errorf("expected closing border '└─', got %q", plain)
	}
}

func TestRenderCodeBlockBorder_TogglesState_OpenToClose(t *testing.T) {
	newState := renderCodeBlockBorder("```", false)
	if !newState {
		t.Error("opening: expected returned state=true")
	}
}

func TestRenderCodeBlockBorder_TogglesState_CloseToOpen(t *testing.T) {
	newState := renderCodeBlockBorder("```", true)
	if newState {
		t.Error("closing: expected returned state=false")
	}
}

// ── renderCodeLine ────────────────────────────────────────────────────────────

func TestRenderCodeLine_NonEmpty(t *testing.T) {
	out := captureOutput(func() { renderCodeLine("  x := 1") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "│") {
		t.Errorf("expected '│' prefix, got %q", plain)
	}
	if !strings.Contains(plain, "x := 1") {
		t.Errorf("expected code content, got %q", plain)
	}
}

func TestRenderCodeLine_Empty(t *testing.T) {
	out := captureOutput(func() { renderCodeLine("") })
	plain := stripANSI(out)
	// Empty line should print just the vertical bar
	trimmed := strings.TrimRight(plain, "\n")
	if trimmed != "│" {
		t.Errorf("empty code line: want '│', got %q", trimmed)
	}
}

func TestRenderCodeLine_TrailingSpaces_Stripped(t *testing.T) {
	out := captureOutput(func() { renderCodeLine("hello   ") })
	plain := stripANSI(out)
	if strings.Contains(plain, "hello   ") {
		t.Errorf("trailing spaces should be stripped, got %q", plain)
	}
	if !strings.Contains(plain, "hello") {
		t.Errorf("content should still be present, got %q", plain)
	}
}

// ── renderHeader ──────────────────────────────────────────────────────────────

func TestRenderHeader_H1(t *testing.T) {
	out := captureOutput(func() { renderHeader("# Title One") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Title One") {
		t.Errorf("H1: expected 'Title One', got %q", plain)
	}
	if !strings.Contains(plain, "═") {
		t.Errorf("H1: expected '═' underline, got %q", plain)
	}
}

func TestRenderHeader_H2(t *testing.T) {
	out := captureOutput(func() { renderHeader("## Section Two") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Section Two") {
		t.Errorf("H2: expected 'Section Two', got %q", plain)
	}
	if !strings.Contains(plain, "─") {
		t.Errorf("H2: expected '─' underline, got %q", plain)
	}
}

func TestRenderHeader_H3(t *testing.T) {
	out := captureOutput(func() { renderHeader("### Sub Section") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Sub Section") {
		t.Errorf("H3: expected 'Sub Section', got %q", plain)
	}
	if strings.Contains(plain, "═") || strings.Contains(plain, "─") {
		t.Errorf("H3: should not have underline, got %q", plain)
	}
}

func TestRenderHeader_H4Plus(t *testing.T) {
	out := captureOutput(func() { renderHeader("#### Deep Header") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Deep Header") {
		t.Errorf("H4+: expected 'Deep Header', got %q", plain)
	}
}

// ── renderBlockquote ─────────────────────────────────────────────────────────

func TestRenderBlockquote_Simple(t *testing.T) {
	out := captureOutput(func() { renderBlockquote("> A wise saying") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "│") {
		t.Errorf("expected '│' prefix, got %q", plain)
	}
	if !strings.Contains(plain, "A wise saying") {
		t.Errorf("expected quote text, got %q", plain)
	}
}

func TestRenderBlockquote_StripsLeadingChevron(t *testing.T) {
	out := captureOutput(func() { renderBlockquote(">no space") })
	plain := stripANSI(out)
	if strings.Contains(plain, ">") {
		t.Errorf("'>' should be stripped from output, got %q", plain)
	}
	if !strings.Contains(plain, "no space") {
		t.Errorf("expected quote text, got %q", plain)
	}
}

// ── renderUnorderedListItem ───────────────────────────────────────────────────

func TestRenderUnorderedListItem_Level0_Dash(t *testing.T) {
	out := captureOutput(func() { renderUnorderedListItem("- item one", "- item one") })
	plain := stripANSI(out)
	if !strings.Contains(plain, SymbolBullet) {
		t.Errorf("expected bullet '•', got %q", plain)
	}
	if !strings.Contains(plain, "item one") {
		t.Errorf("expected item text, got %q", plain)
	}
}

func TestRenderUnorderedListItem_Level0_Asterisk(t *testing.T) {
	out := captureOutput(func() { renderUnorderedListItem("* item", "* item") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "item") {
		t.Errorf("expected item text, got %q", plain)
	}
}

func TestRenderUnorderedListItem_Level1_Indent(t *testing.T) {
	out := captureOutput(func() { renderUnorderedListItem("  - sub item", "- sub item") })
	plain := stripANSI(out)
	// Level 1 → 2 leading spaces of indent
	if !strings.HasPrefix(plain, "  ") {
		t.Errorf("expected 2-space indent for level 1, got %q", plain)
	}
}

// ── renderOrderedListItem ─────────────────────────────────────────────────────

func TestRenderOrderedListItem_Simple(t *testing.T) {
	out := captureOutput(func() { renderOrderedListItem("1. first item", "1. first item") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "1.") {
		t.Errorf("expected '1.' in output, got %q", plain)
	}
	if !strings.Contains(plain, "first item") {
		t.Errorf("expected item text, got %q", plain)
	}
}

func TestRenderOrderedListItem_HighNumber(t *testing.T) {
	out := captureOutput(func() { renderOrderedListItem("42. answer", "42. answer") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "42.") {
		t.Errorf("expected '42.' in output, got %q", plain)
	}
}

func TestRenderOrderedListItem_Indented(t *testing.T) {
	out := captureOutput(func() { renderOrderedListItem("    2. nested", "2. nested") })
	plain := stripANSI(out)
	// 4 spaces → level 2 → 4 spaces indent
	if !strings.HasPrefix(plain, "    ") {
		t.Errorf("expected 4-space indent for level 2, got %q", plain)
	}
}

func TestRenderOrderedListItem_NoOutput_IfMalformed(t *testing.T) {
	out := captureOutput(func() { renderOrderedListItem("not a list", "not a list") })
	if out != "" {
		t.Errorf("malformed input: expected no output, got %q", out)
	}
}

// ── renderTaskListItem ────────────────────────────────────────────────────────

func TestRenderTaskListItem_Checked_LowerX(t *testing.T) {
	out := captureOutput(func() { renderTaskListItem("- [x] done task") })
	plain := stripANSI(out)
	if !strings.Contains(plain, SymbolCheck) {
		t.Errorf("checked task: expected '✓', got %q", plain)
	}
	if !strings.Contains(plain, "done task") {
		t.Errorf("expected task text, got %q", plain)
	}
}

func TestRenderTaskListItem_Checked_UpperX(t *testing.T) {
	out := captureOutput(func() { renderTaskListItem("- [X] also done") })
	plain := stripANSI(out)
	if !strings.Contains(plain, SymbolCheck) {
		t.Errorf("checked task (uppercase X): expected '✓', got %q", plain)
	}
}

func TestRenderTaskListItem_Unchecked(t *testing.T) {
	out := captureOutput(func() { renderTaskListItem("- [ ] pending task") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "☐") {
		t.Errorf("unchecked task: expected '☐', got %q", plain)
	}
	if !strings.Contains(plain, "pending task") {
		t.Errorf("expected task text, got %q", plain)
	}
}

func TestRenderTaskListItem_NoOutput_IfMalformed(t *testing.T) {
	out := captureOutput(func() { renderTaskListItem("not a task") })
	if out != "" {
		t.Errorf("malformed input: expected no output, got %q", out)
	}
}

// ── formatInlineMarkdown ──────────────────────────────────────────────────────

func TestFormatInlineMarkdown_PlainText(t *testing.T) {
	result := stripANSI(formatInlineMarkdown("hello world"))
	if result != "hello world" {
		t.Errorf("plain text: want %q, got %q", "hello world", result)
	}
}

func TestFormatInlineMarkdown_Bold(t *testing.T) {
	result := formatInlineMarkdown("**bold**")
	if !strings.Contains(stripANSI(result), "bold") {
		t.Errorf("bold: expected 'bold' in output, got %q", result)
	}
	if strings.Contains(result, "**") {
		t.Errorf("bold markers '**' should be removed, got %q", result)
	}
}

func TestFormatInlineMarkdown_Italic(t *testing.T) {
	result := formatInlineMarkdown("*italic*")
	if !strings.Contains(stripANSI(result), "italic") {
		t.Errorf("italic: expected 'italic' in output, got %q", result)
	}
	if strings.Contains(result, "*italic*") {
		t.Errorf("italic markers should be removed, got %q", result)
	}
}

func TestFormatInlineMarkdown_InlineCode(t *testing.T) {
	result := formatInlineMarkdown("`code`")
	if !strings.Contains(stripANSI(result), "code") {
		t.Errorf("inline code: expected 'code' in output, got %q", result)
	}
	if strings.Contains(result, "`code`") {
		t.Errorf("backtick markers should be removed, got %q", result)
	}
}

func TestFormatInlineMarkdown_Link(t *testing.T) {
	result := formatInlineMarkdown("[click here](https://example.com)")
	plain := stripANSI(result)
	if !strings.Contains(plain, "click here") {
		t.Errorf("link text missing, got %q", plain)
	}
	if !strings.Contains(plain, "https://example.com") {
		t.Errorf("link URL missing, got %q", plain)
	}
}

func TestFormatInlineMarkdown_EscapedStar(t *testing.T) {
	result := formatInlineMarkdown(`\*literal star\*`)
	plain := stripANSI(result)
	if !strings.Contains(plain, "*") {
		t.Errorf("escaped '*' should appear as literal, got %q", plain)
	}
}

// ── renderMarkdownLine ────────────────────────────────────────────────────────

func TestRenderMarkdownLine_CodeBlockToggle(t *testing.T) {
	// Opening fence switches inCodeBlock false→true
	newCode, _ := renderMarkdownLine("```go", false, false)
	if !newCode {
		t.Error("opening ```: expected inCodeBlock=true")
	}
	// Closing fence switches inCodeBlock true→false
	newCode, _ = renderMarkdownLine("```", true, false)
	if newCode {
		t.Error("closing ```: expected inCodeBlock=false")
	}
}

func TestRenderMarkdownLine_CodeLine_InsideBlock(t *testing.T) {
	out := captureOutput(func() { renderMarkdownLine("  x := 1", true, false) })
	plain := stripANSI(out)
	if !strings.Contains(plain, "│") {
		t.Errorf("code line: expected '│' prefix, got %q", plain)
	}
	if !strings.Contains(plain, "x := 1") {
		t.Errorf("code line: expected content, got %q", plain)
	}
}

func TestRenderMarkdownLine_EmptyLine_ResetsInList(t *testing.T) {
	_, newList := renderMarkdownLine("", false, true)
	if newList {
		t.Error("empty line: expected inList=false")
	}
}

func TestRenderMarkdownLine_Header(t *testing.T) {
	out := captureOutput(func() { renderMarkdownLine("## Hello", false, false) })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Hello") {
		t.Errorf("header: expected 'Hello', got %q", plain)
	}
}

func TestRenderMarkdownLine_HorizontalRule(t *testing.T) {
	out := captureOutput(func() { renderMarkdownLine("---", false, false) })
	plain := stripANSI(out)
	if !strings.Contains(plain, "─") {
		t.Errorf("horizontal rule: expected '─', got %q", plain)
	}
}

func TestRenderMarkdownLine_Blockquote(t *testing.T) {
	out := captureOutput(func() { renderMarkdownLine("> quote", false, false) })
	plain := stripANSI(out)
	if !strings.Contains(plain, "│") {
		t.Errorf("blockquote: expected '│', got %q", plain)
	}
}

func TestRenderMarkdownLine_UnorderedList_SetsInList(t *testing.T) {
	out := captureOutput(func() { renderMarkdownLine("- item", false, false) })
	plain := stripANSI(out)
	if !strings.Contains(plain, SymbolBullet) {
		t.Errorf("unordered list: expected bullet, got %q", plain)
	}
	_, newList := renderMarkdownLine("- item", false, false)
	if !newList {
		t.Error("unordered list: expected inList=true")
	}
}

func TestRenderMarkdownLine_OrderedList_SetsInList(t *testing.T) {
	_, newList := renderMarkdownLine("1. first", false, false)
	if !newList {
		t.Error("ordered list: expected inList=true")
	}
}

func TestRenderMarkdownLine_TaskList_RenderedAsBullet(t *testing.T) {
	// Task list items match the unordered-list pattern first (same behaviour as Markdown()).
	// They are rendered as bullet items, not as check/uncheck symbols.
	out := captureOutput(func() { renderMarkdownLine("- [x] done", false, false) })
	plain := stripANSI(out)
	if !strings.Contains(plain, SymbolBullet) {
		t.Errorf("task list via renderMarkdownLine: expected bullet symbol, got %q", plain)
	}
}

func TestRenderMarkdownLine_Paragraph_NotInList(t *testing.T) {
	out := captureOutput(func() { renderMarkdownLine("hello world", false, false) })
	plain := stripANSI(out)
	if !strings.Contains(plain, "hello world") {
		t.Errorf("paragraph: expected text, got %q", plain)
	}
	if strings.HasPrefix(plain, "  ") {
		t.Errorf("paragraph not in list: must not have indent, got %q", plain)
	}
}

func TestRenderMarkdownLine_Paragraph_InList_Indented(t *testing.T) {
	out := captureOutput(func() { renderMarkdownLine("continuation", false, true) })
	plain := stripANSI(out)
	if !strings.HasPrefix(plain, "  ") {
		t.Errorf("paragraph in list: expected 2-space indent, got %q", plain)
	}
}

func TestRenderMarkdownLine_InCodeBlock_PassthroughState(t *testing.T) {
	// Inside a code block, inCodeBlock stays true and inList is preserved
	newCode, newList := renderMarkdownLine("some code", true, true)
	if !newCode {
		t.Error("code block: inCodeBlock should stay true")
	}
	if !newList {
		t.Error("code block: inList should be preserved")
	}
}

// ── MarkdownChunkParser methods (markdown.chunk.blocks.go) ────────────────────

func newTestParser() *MarkdownChunkParser {
	return NewMarkdownChunkParser()
}

func TestProcessCodeBlockDelimiter_OpenNoLang(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processCodeBlockDelimiter("```") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "┌─ Code") {
		t.Errorf("opening no-lang: expected '┌─ Code', got %q", plain)
	}
	if !p.inCodeBlock {
		t.Error("inCodeBlock should be true after opening")
	}
}

func TestProcessCodeBlockDelimiter_OpenWithLang(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processCodeBlockDelimiter("```go") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Code: go") {
		t.Errorf("opening with lang: expected 'Code: go', got %q", plain)
	}
	if p.codeBlockLang != "go" {
		t.Errorf("codeBlockLang: want 'go', got %q", p.codeBlockLang)
	}
}

func TestProcessCodeBlockDelimiter_Close(t *testing.T) {
	p := newTestParser()
	p.inCodeBlock = true
	p.codeBlockLang = "go"
	out := captureOutput(func() { p.processCodeBlockDelimiter("```") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "└─") {
		t.Errorf("closing: expected '└─', got %q", plain)
	}
	if p.inCodeBlock {
		t.Error("inCodeBlock should be false after closing")
	}
	if p.codeBlockLang != "" {
		t.Errorf("codeBlockLang should be cleared, got %q", p.codeBlockLang)
	}
}

func TestProcessCodeLine_NonEmpty(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processCodeLine("  x := 1") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "│") {
		t.Errorf("expected '│' prefix, got %q", plain)
	}
	if !strings.Contains(plain, "x := 1") {
		t.Errorf("expected code content, got %q", plain)
	}
}

func TestProcessCodeLine_Empty(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processCodeLine("") })
	plain := strings.TrimRight(stripANSI(out), "\n")
	if plain != "│" {
		t.Errorf("empty code line: want '│', got %q", plain)
	}
}

func TestProcessHeaderChunk_H1(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processHeaderChunk("# Hello") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Hello") {
		t.Errorf("H1: expected 'Hello', got %q", plain)
	}
	if !strings.Contains(plain, "═") {
		t.Errorf("H1: expected '═' underline, got %q", plain)
	}
}

func TestProcessHeaderChunk_H2(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processHeaderChunk("## Section") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Section") {
		t.Errorf("H2: expected 'Section', got %q", plain)
	}
	if !strings.Contains(plain, "─") {
		t.Errorf("H2: expected '─' underline, got %q", plain)
	}
}

func TestProcessHeaderChunk_H3(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processHeaderChunk("### Sub") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "Sub") {
		t.Errorf("H3: expected 'Sub', got %q", plain)
	}
}

func TestProcessHeaderChunk_Empty_Skipped(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processHeaderChunk("#") })
	if out != "" {
		t.Errorf("empty header: expected no output, got %q", out)
	}
}

func TestProcessBlockquoteChunk_Simple(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processBlockquoteChunk("> A quote") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "│") {
		t.Errorf("expected '│' prefix, got %q", plain)
	}
	if !strings.Contains(plain, "A quote") {
		t.Errorf("expected quote text, got %q", plain)
	}
}

func TestProcessUnorderedListItemChunk_SetsInList(t *testing.T) {
	p := newTestParser()
	captureOutput(func() { p.processUnorderedListItemChunk("- item", "- item") })
	if !p.inList {
		t.Error("expected inList=true after unordered list item")
	}
}

func TestProcessUnorderedListItemChunk_Output(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processUnorderedListItemChunk("- hello", "- hello") })
	plain := stripANSI(out)
	if !strings.Contains(plain, SymbolBullet) {
		t.Errorf("expected bullet symbol, got %q", plain)
	}
	if !strings.Contains(plain, "hello") {
		t.Errorf("expected item text, got %q", plain)
	}
}

func TestProcessOrderedListItemChunk_Output(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processOrderedListItemChunk("1. first", "1. first") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "1.") {
		t.Errorf("expected '1.' in output, got %q", plain)
	}
	if !strings.Contains(plain, "first") {
		t.Errorf("expected item text, got %q", plain)
	}
	if !p.inList {
		t.Error("expected inList=true after ordered list item")
	}
}

func TestProcessTaskListItemChunk_Checked(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processTaskListItemChunk("- [x] done") })
	plain := stripANSI(out)
	if !strings.Contains(plain, SymbolCheck) {
		t.Errorf("checked task: expected check symbol, got %q", plain)
	}
}

func TestProcessTaskListItemChunk_Unchecked(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processTaskListItemChunk("- [ ] todo") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "☐") {
		t.Errorf("unchecked task: expected '☐', got %q", plain)
	}
}

func TestProcessParagraph_Normal(t *testing.T) {
	p := newTestParser()
	out := captureOutput(func() { p.processParagraph("hello world") })
	plain := stripANSI(out)
	if !strings.Contains(plain, "hello world") {
		t.Errorf("expected paragraph text, got %q", plain)
	}
}

func TestProcessParagraph_InList_Indented(t *testing.T) {
	p := newTestParser()
	p.inList = true
	out := captureOutput(func() { p.processParagraph("continuation") })
	plain := stripANSI(out)
	if !strings.HasPrefix(plain, "  ") {
		t.Errorf("inList paragraph: expected 2-space indent, got %q", plain)
	}
}

func TestProcessParagraph_AlreadyPrinted_OnlyNewline(t *testing.T) {
	p := newTestParser()
	p.alreadyPrinted = true
	out := captureOutput(func() { p.processParagraph("streamed text") })
	// Should print only a newline, not the text again
	if strings.Contains(out, "streamed text") {
		t.Errorf("alreadyPrinted: text should not be reprinted, got %q", out)
	}
	if out != "\n" {
		t.Errorf("alreadyPrinted: expected only newline, got %q", out)
	}
}
