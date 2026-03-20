package prompt

import (
	"strings"
	"testing"
)

// ── renderLineWithCursor ──────────────────────────────────────────────────────

func TestRenderLineWithCursor_TextBeforeCursor(t *testing.T) {
	line := []rune("hello")
	out := capturePromptOutput(func() { renderLineWithCursor(line, 3, true, CursorBlock) })
	plain := stripANSIMulti(out)
	// "hel" appears before the cursor
	if !strings.HasPrefix(plain, "hel") {
		t.Errorf("expected 'hel' before cursor, got %q", plain)
	}
}

func TestRenderLineWithCursor_TextAfterCursor(t *testing.T) {
	line := []rune("hello")
	out := capturePromptOutput(func() { renderLineWithCursor(line, 1, true, CursorBlock) })
	plain := stripANSIMulti(out)
	// "llo" appears after the cursor character
	if !strings.HasSuffix(strings.TrimRight(plain, "\n"), "llo") {
		t.Errorf("expected 'llo' after cursor, got %q", plain)
	}
}

func TestRenderLineWithCursor_NoPrefixAtColumn0(t *testing.T) {
	line := []rune("hello")
	// Cursor at 0: nothing printed before cursor
	out := capturePromptOutput(func() { renderLineWithCursor(line, 0, true, CursorBlock) })
	// Should not start with the raw char 'h' before ANSI code — block cursor wraps it
	_ = out // just verify no panic
}

func TestRenderLineWithCursor_VisibleCursor_EmitsEscapeCode(t *testing.T) {
	line := []rune("hi")
	out := capturePromptOutput(func() { renderLineWithCursor(line, 0, true, CursorBlock) })
	if !strings.Contains(out, "\033[7m") {
		t.Errorf("visible block cursor: expected inverse-video escape, got %q", out)
	}
}

func TestRenderLineWithCursor_HiddenCursor_NoEscapeCode(t *testing.T) {
	line := []rune("hi")
	// cursorVisible=false with blinking style → hidden cursor (no escape codes for cursor)
	out := capturePromptOutput(func() { renderLineWithCursor(line, 0, false, CursorBlockBlink) })
	// Should contain 'h' but without \033[7m
	if strings.Contains(out, "\033[7m") {
		t.Errorf("hidden cursor: must not emit inverse-video escape, got %q", out)
	}
	if !strings.Contains(out, "h") {
		t.Errorf("hidden cursor: expected 'h' in output, got %q", out)
	}
}

func TestRenderLineWithCursor_UnderlineCursor_EmitsEscapeCode(t *testing.T) {
	line := []rune("test")
	out := capturePromptOutput(func() { renderLineWithCursor(line, 2, true, CursorUnderline) })
	if !strings.Contains(out, "\033[4m") {
		t.Errorf("underline cursor: expected underline escape, got %q", out)
	}
}

func TestRenderLineWithCursor_CursorAtEnd_NoTrailingText(t *testing.T) {
	line := []rune("ab")
	// cursorCol == len(line) → cursor at end, no text after
	out := capturePromptOutput(func() { renderLineWithCursor(line, 2, true, CursorBlock) })
	plain := stripANSIMulti(out)
	// Should start with "ab" then a space (block cursor at end)
	if !strings.HasPrefix(plain, "ab") {
		t.Errorf("cursor at end: expected 'ab' prefix, got %q", plain)
	}
}

func TestRenderLineWithCursor_EmptyLine_CursorAtEnd(t *testing.T) {
	// Empty line, cursor at 0 = at end — should not panic
	out := capturePromptOutput(func() { renderLineWithCursor([]rune{}, 0, true, CursorBlock) })
	if !strings.Contains(out, "\033[7m") {
		t.Errorf("empty line block cursor: expected inverse-video escape, got %q", out)
	}
}

// ── appendAndDispatchEscape ───────────────────────────────────────────────────

func TestAppendAndDispatchEscape_IncompleteSeq_NotDone(t *testing.T) {
	editor := NewMultiLineEditor("", CursorBlock)
	seq, done := appendAndDispatchEscape([]byte{}, '[', editor)
	if done {
		t.Error("single '[' should not be done")
	}
	if len(seq) != 1 || seq[0] != '[' {
		t.Errorf("expected seq=['['], got %v", seq)
	}
}

func TestAppendAndDispatchEscape_UpArrow_MovesUp(t *testing.T) {
	editor := NewMultiLineEditor("line1\nline2", CursorBlock)
	// NewMultiLineEditor places cursor at end of last line (line 1)
	seq, _ := appendAndDispatchEscape([]byte{}, '[', editor)
	seq, done := appendAndDispatchEscape(seq, 'A', editor)
	if !done {
		t.Fatal("[A: expected done=true")
	}
	if len(seq) != 0 {
		t.Errorf("[A: expected empty seq after dispatch, got %v", seq)
	}
	if editor.cursorLine != 0 {
		t.Errorf("[A: expected cursorLine=0 after MoveUp, got %d", editor.cursorLine)
	}
}

func TestAppendAndDispatchEscape_DownArrow_MovesDown(t *testing.T) {
	editor := NewMultiLineEditor("line1\nline2", CursorBlock)
	editor.cursorLine = 0
	seq, _ := appendAndDispatchEscape([]byte{}, '[', editor)
	_, done := appendAndDispatchEscape(seq, 'B', editor)
	if !done {
		t.Fatal("[B: expected done=true")
	}
	if editor.cursorLine != 1 {
		t.Errorf("[B: expected cursorLine=1 after MoveDown, got %d", editor.cursorLine)
	}
}

func TestAppendAndDispatchEscape_LeftArrow_MovesLeft(t *testing.T) {
	editor := NewMultiLineEditor("hello", CursorBlock)
	// cursorCol = 5 (end of "hello")
	seq, _ := appendAndDispatchEscape([]byte{}, '[', editor)
	_, done := appendAndDispatchEscape(seq, 'D', editor)
	if !done {
		t.Fatal("[D: expected done=true")
	}
	if editor.cursorCol != 4 {
		t.Errorf("[D: expected cursorCol=4, got %d", editor.cursorCol)
	}
}

func TestAppendAndDispatchEscape_RightArrow_MovesRight(t *testing.T) {
	editor := NewMultiLineEditor("hello", CursorBlock)
	editor.cursorCol = 2
	seq, _ := appendAndDispatchEscape([]byte{}, '[', editor)
	_, done := appendAndDispatchEscape(seq, 'C', editor)
	if !done {
		t.Fatal("[C: expected done=true")
	}
	if editor.cursorCol != 3 {
		t.Errorf("[C: expected cursorCol=3, got %d", editor.cursorCol)
	}
}

func TestAppendAndDispatchEscape_Delete_RemovesChar(t *testing.T) {
	editor := NewMultiLineEditor("hello", CursorBlock)
	editor.cursorCol = 0
	// Feed "[3~" byte by byte
	seq := []byte{}
	for _, b := range []byte("[3~") {
		seq, _ = appendAndDispatchEscape(seq, b, editor)
	}
	if string(editor.lines[0]) != "ello" {
		t.Errorf("[3~: expected 'ello', got %q", string(editor.lines[0]))
	}
}

func TestAppendAndDispatchEscape_UnknownPrefix_Done(t *testing.T) {
	// Non-'[' prefix (e.g. ESC O P = F1 on some terminals) should reset immediately
	editor := NewMultiLineEditor("", CursorBlock)
	seq, done := appendAndDispatchEscape([]byte{'O'}, 'P', editor)
	if !done {
		t.Error("non-'[' prefix: expected done=true")
	}
	if len(seq) != 0 {
		t.Errorf("non-'[' prefix: expected empty seq, got %v", seq)
	}
}

// ── processMultiLineKey ───────────────────────────────────────────────────────

func TestProcessMultiLineKey_CtrlC_Exit(t *testing.T) {
	editor := NewMultiLineEditor("", CursorBlock)
	result := processMultiLineKey(0x03, editor)
	if !result.exit {
		t.Error("Ctrl+C: expected exit=true")
	}
	if result.done {
		t.Error("Ctrl+C: done should be false")
	}
}

func TestProcessMultiLineKey_CtrlD_Done(t *testing.T) {
	editor := NewMultiLineEditor("hello", CursorBlock)
	result := processMultiLineKey(0x04, editor)
	if !result.done {
		t.Error("Ctrl+D: expected done=true")
	}
	if result.text != "hello" {
		t.Errorf("Ctrl+D: expected text='hello', got %q", result.text)
	}
	if result.exit {
		t.Error("Ctrl+D: exit should be false")
	}
}

func TestProcessMultiLineKey_Enter_InsertsNewLine(t *testing.T) {
	editor := NewMultiLineEditor("hello", CursorBlock)
	editor.cursorCol = 3
	result := processMultiLineKey(0x0d, editor)
	if result.exit || result.done {
		t.Error("Enter: should not set exit or done")
	}
	if len(editor.lines) != 2 {
		t.Errorf("Enter: expected 2 lines, got %d", len(editor.lines))
	}
}

func TestProcessMultiLineKey_Backspace_DeletesChar(t *testing.T) {
	editor := NewMultiLineEditor("hello", CursorBlock)
	// cursor at end (col 5)
	result := processMultiLineKey(0x7f, editor)
	if result.exit || result.done {
		t.Error("Backspace: should not set exit or done")
	}
	if string(editor.lines[0]) != "hell" {
		t.Errorf("Backspace: expected 'hell', got %q", string(editor.lines[0]))
	}
}

func TestProcessMultiLineKey_PrintableChar_InsertsRune(t *testing.T) {
	editor := NewMultiLineEditor("", CursorBlock)
	result := processMultiLineKey('x', editor)
	if result.exit || result.done {
		t.Error("printable: should not set exit or done")
	}
	if string(editor.lines[0]) != "x" {
		t.Errorf("printable: expected 'x', got %q", string(editor.lines[0]))
	}
}

func TestProcessMultiLineKey_NonPrintable_NoChange(t *testing.T) {
	editor := NewMultiLineEditor("hello", CursorBlock)
	result := processMultiLineKey(0x00, editor) // NUL — not printable, not handled
	if result.exit || result.done {
		t.Error("non-printable: should not set exit or done")
	}
	if string(editor.lines[0]) != "hello" {
		t.Errorf("non-printable: content should be unchanged, got %q", string(editor.lines[0]))
	}
}

// stripANSIMulti removes ANSI escape sequences (same pattern as edit.render_test.go helper).
func stripANSIMulti(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			// skip until letter
			i += 2
			for i < len(s) && (s[i] < 'A' || s[i] > 'z') {
				i++
			}
			i++ // skip terminator
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}
