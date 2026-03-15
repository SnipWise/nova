package prompt

import (
	"testing"
)

// ── processEscapeInput ────────────────────────────────────────────────────────

func TestProcessEscapeInput_RegularByte_NotHandled(t *testing.T) {
	inEscape := false
	seq := make([]byte, 0, 10)
	r := processEscapeInput('a', &inEscape, &seq, []rune("hello"), 2)
	if r.handled {
		t.Error("regular byte: want handled=false")
	}
	if inEscape {
		t.Error("inEscape must stay false")
	}
}

func TestProcessEscapeInput_EscByte_SetsInEscape(t *testing.T) {
	inEscape := false
	seq := make([]byte, 0, 10)
	r := processEscapeInput(0x1b, &inEscape, &seq, []rune("hello"), 2)
	if !r.handled {
		t.Error("ESC byte: want handled=true")
	}
	if r.render {
		t.Error("ESC byte: want render=false (no redraw needed yet)")
	}
	if !inEscape {
		t.Error("inEscape should be true after ESC byte")
	}
}

func TestProcessEscapeInput_InEscape_ShortSeq_Handled_Render(t *testing.T) {
	inEscape := true
	seq := []byte{'['}
	// send '3' → seq becomes "[3", len==2 but no terminator → not consumed
	r := processEscapeInput('3', &inEscape, &seq, []rune("hi"), 1)
	if !r.handled {
		t.Error("inEscape byte: want handled=true")
	}
	if !r.render {
		t.Error("inEscape byte: want render=true")
	}
	if !inEscape {
		t.Error("inEscape must stay true for incomplete sequence")
	}
}

func TestProcessEscapeInput_InEscape_ValidSeq_Consumed(t *testing.T) {
	inEscape := true
	seq := []byte{'['}
	buf := []rune("hello")
	// "[D" = left arrow — cursor should move left
	r := processEscapeInput('D', &inEscape, &seq, buf, 3)
	if !r.handled {
		t.Error("want handled=true")
	}
	if inEscape {
		t.Error("inEscape should be reset after consumed sequence")
	}
	if r.cursor != 2 {
		t.Errorf("cursor: want 2, got %d", r.cursor)
	}
	if len(seq) != 0 {
		t.Errorf("escapeSeq should be cleared, got %v", seq)
	}
}

// ── processEscapeSeq ─────────────────────────────────────────────────────────

func TestProcessEscapeSeq_LeftArrow(t *testing.T) {
	buf := []rune("hello")
	buf, cur, ok := processEscapeSeq("[D", buf, 3)
	if !ok {
		t.Fatal("expected consumed=true")
	}
	if cur != 2 {
		t.Errorf("cursor: want 2, got %d", cur)
	}
	if string(buf) != "hello" {
		t.Errorf("buffer must not change, got %q", string(buf))
	}
}

func TestProcessEscapeSeq_LeftArrow_AtStart(t *testing.T) {
	buf := []rune("hello")
	_, cur, ok := processEscapeSeq("[D", buf, 0)
	if !ok {
		t.Fatal("expected consumed=true")
	}
	if cur != 0 {
		t.Errorf("cursor must stay 0, got %d", cur)
	}
}

func TestProcessEscapeSeq_RightArrow(t *testing.T) {
	buf := []rune("hello")
	_, cur, ok := processEscapeSeq("[C", buf, 2)
	if !ok {
		t.Fatal("expected consumed=true")
	}
	if cur != 3 {
		t.Errorf("cursor: want 3, got %d", cur)
	}
}

func TestProcessEscapeSeq_RightArrow_AtEnd(t *testing.T) {
	buf := []rune("hello")
	_, cur, ok := processEscapeSeq("[C", buf, 5)
	if !ok {
		t.Fatal("expected consumed=true")
	}
	if cur != 5 {
		t.Errorf("cursor must stay at end (5), got %d", cur)
	}
}

func TestProcessEscapeSeq_Home(t *testing.T) {
	buf := []rune("hello")
	_, cur, ok := processEscapeSeq("[H", buf, 4)
	if !ok {
		t.Fatal("expected consumed=true")
	}
	if cur != 0 {
		t.Errorf("cursor: want 0, got %d", cur)
	}
}

func TestProcessEscapeSeq_End(t *testing.T) {
	buf := []rune("hello")
	_, cur, ok := processEscapeSeq("[F", buf, 0)
	if !ok {
		t.Fatal("expected consumed=true")
	}
	if cur != 5 {
		t.Errorf("cursor: want 5 (len), got %d", cur)
	}
}

func TestProcessEscapeSeq_Delete_MidCursor(t *testing.T) {
	buf := []rune("hello")
	buf, cur, ok := processEscapeSeq("[3~", buf, 2)
	if !ok {
		t.Fatal("expected consumed=true")
	}
	if string(buf) != "helo" {
		t.Errorf("buffer: want %q, got %q", "helo", string(buf))
	}
	if cur != 2 {
		t.Errorf("cursor must not move, got %d", cur)
	}
}

func TestProcessEscapeSeq_Delete_AtEnd(t *testing.T) {
	buf := []rune("hello")
	buf, cur, ok := processEscapeSeq("[3~", buf, 5)
	if !ok {
		t.Fatal("expected consumed=true")
	}
	if string(buf) != "hello" {
		t.Errorf("buffer must not change at end, got %q", string(buf))
	}
	if cur != 5 {
		t.Errorf("cursor must not move, got %d", cur)
	}
}

func TestProcessEscapeSeq_OtherTilde_Ignored(t *testing.T) {
	// [5~ (Page Up) — recognised as ~ suffix but not [3~ prefix, so no buffer change
	buf := []rune("hello")
	buf, cur, ok := processEscapeSeq("[5~", buf, 2)
	if !ok {
		t.Fatal("expected consumed=true (tilde suffix is consumed)")
	}
	if string(buf) != "hello" {
		t.Errorf("buffer must not change for Page Up, got %q", string(buf))
	}
	if cur != 2 {
		t.Errorf("cursor must not move, got %d", cur)
	}
}

func TestProcessEscapeSeq_UnknownSuffix_NotConsumed(t *testing.T) {
	buf := []rune("hello")
	_, _, ok := processEscapeSeq("[A", buf, 2) // up arrow — not handled
	if ok {
		t.Error("expected consumed=false for unhandled sequence [A")
	}
}

func TestProcessEscapeSeq_NoLeadingBracket_NotConsumed(t *testing.T) {
	buf := []rune("hello")
	_, _, ok := processEscapeSeq("OD", buf, 2) // SS3 sequence, no "["
	if ok {
		t.Error("expected consumed=false for sequence without leading [")
	}
}

func TestProcessEscapeSeq_PartialSequence_NotConsumed(t *testing.T) {
	// "[3" is incomplete (no terminator yet)
	buf := []rune("hello")
	_, _, ok := processEscapeSeq("[3", buf, 2)
	if ok {
		t.Error("expected consumed=false for incomplete sequence [3")
	}
}

// ── processControlKey ────────────────────────────────────────────────────────

func TestProcessControlKey_CtrlC(t *testing.T) {
	buf := []rune("hello")
	r := processControlKey(0x03, buf, 2)
	if !r.ctrlC {
		t.Error("expected ctrlC=true")
	}
	if r.done {
		t.Error("ctrlC must not set done")
	}
}

func TestProcessControlKey_CtrlD_EmptyBuffer(t *testing.T) {
	r := processControlKey(0x04, []rune{}, 0)
	if !r.done {
		t.Error("expected done=true for Ctrl+D on empty buffer")
	}
	if r.err == nil || r.err.Error() != "EOF" {
		t.Errorf("expected err=EOF, got %v", r.err)
	}
}

func TestProcessControlKey_CtrlD_NonEmptyBuffer(t *testing.T) {
	r := processControlKey(0x04, []rune("hi"), 2)
	if r.done {
		t.Error("expected done=false for Ctrl+D on non-empty buffer")
	}
	if r.ctrlC {
		t.Error("expected ctrlC=false")
	}
}

func TestProcessControlKey_Enter_CR(t *testing.T) {
	buf := []rune("hello")
	r := processControlKey(0x0d, buf, 3)
	if !r.done {
		t.Error("expected done=true for Enter")
	}
	if r.output != "hello" {
		t.Errorf("output: want %q, got %q", "hello", r.output)
	}
	if r.err != nil {
		t.Errorf("expected no error, got %v", r.err)
	}
}

func TestProcessControlKey_Enter_LF(t *testing.T) {
	buf := []rune("world")
	r := processControlKey(0x0a, buf, 0)
	if !r.done {
		t.Error("expected done=true for LF")
	}
	if r.output != "world" {
		t.Errorf("output: want %q, got %q", "world", r.output)
	}
}

func TestProcessControlKey_Backspace_DEL(t *testing.T) {
	buf := []rune("hello")
	r := processControlKey(0x7f, buf, 3)
	if string(r.buffer) != "helo" {
		t.Errorf("buffer: want %q, got %q", "helo", string(r.buffer))
	}
	if r.cursor != 2 {
		t.Errorf("cursor: want 2, got %d", r.cursor)
	}
}

func TestProcessControlKey_Backspace_BS(t *testing.T) {
	buf := []rune("hello")
	r := processControlKey(0x08, buf, 5)
	if string(r.buffer) != "hell" {
		t.Errorf("buffer: want %q, got %q", "hell", string(r.buffer))
	}
	if r.cursor != 4 {
		t.Errorf("cursor: want 4, got %d", r.cursor)
	}
}

func TestProcessControlKey_Backspace_AtStart(t *testing.T) {
	buf := []rune("hello")
	r := processControlKey(0x7f, buf, 0)
	if string(r.buffer) != "hello" {
		t.Errorf("buffer must not change at start, got %q", string(r.buffer))
	}
	if r.cursor != 0 {
		t.Errorf("cursor must stay 0, got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlA_Home(t *testing.T) {
	r := processControlKey(0x01, []rune("hello"), 4)
	if r.cursor != 0 {
		t.Errorf("cursor: want 0, got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlE_End(t *testing.T) {
	r := processControlKey(0x05, []rune("hello"), 1)
	if r.cursor != 5 {
		t.Errorf("cursor: want 5, got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlB_Left(t *testing.T) {
	r := processControlKey(0x02, []rune("hello"), 3)
	if r.cursor != 2 {
		t.Errorf("cursor: want 2, got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlB_Left_AtStart(t *testing.T) {
	r := processControlKey(0x02, []rune("hello"), 0)
	if r.cursor != 0 {
		t.Errorf("cursor must stay 0, got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlF_Right(t *testing.T) {
	r := processControlKey(0x06, []rune("hello"), 2)
	if r.cursor != 3 {
		t.Errorf("cursor: want 3, got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlF_Right_AtEnd(t *testing.T) {
	r := processControlKey(0x06, []rune("hello"), 5)
	if r.cursor != 5 {
		t.Errorf("cursor must stay at end (5), got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlK_KillToEnd(t *testing.T) {
	r := processControlKey(0x0b, []rune("hello"), 2)
	if string(r.buffer) != "he" {
		t.Errorf("buffer: want %q, got %q", "he", string(r.buffer))
	}
	if r.cursor != 2 {
		t.Errorf("cursor must not move, got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlK_AtEnd(t *testing.T) {
	r := processControlKey(0x0b, []rune("hello"), 5)
	if string(r.buffer) != "hello" {
		t.Errorf("buffer must not change when at end, got %q", string(r.buffer))
	}
}

func TestProcessControlKey_CtrlU_KillToStart(t *testing.T) {
	r := processControlKey(0x15, []rune("hello"), 3)
	if string(r.buffer) != "lo" {
		t.Errorf("buffer: want %q, got %q", "lo", string(r.buffer))
	}
	if r.cursor != 0 {
		t.Errorf("cursor: want 0, got %d", r.cursor)
	}
}

func TestProcessControlKey_CtrlU_AtStart(t *testing.T) {
	r := processControlKey(0x15, []rune("hello"), 0)
	if string(r.buffer) != "hello" {
		t.Errorf("buffer must not change when at start, got %q", string(r.buffer))
	}
	if r.cursor != 0 {
		t.Errorf("cursor must stay 0, got %d", r.cursor)
	}
}

func TestProcessControlKey_PrintableChar_Insert_MidBuffer(t *testing.T) {
	r := processControlKey('x', []rune("hello"), 2)
	if string(r.buffer) != "hexllo" {
		t.Errorf("buffer: want %q, got %q", "hexllo", string(r.buffer))
	}
	if r.cursor != 3 {
		t.Errorf("cursor: want 3, got %d", r.cursor)
	}
}

func TestProcessControlKey_PrintableChar_Insert_AtEnd(t *testing.T) {
	r := processControlKey('!', []rune("hello"), 5)
	if string(r.buffer) != "hello!" {
		t.Errorf("buffer: want %q, got %q", "hello!", string(r.buffer))
	}
	if r.cursor != 6 {
		t.Errorf("cursor: want 6, got %d", r.cursor)
	}
}

func TestProcessControlKey_PrintableChar_Insert_AtStart(t *testing.T) {
	r := processControlKey('A', []rune("hello"), 0)
	if string(r.buffer) != "Ahello" {
		t.Errorf("buffer: want %q, got %q", "Ahello", string(r.buffer))
	}
	if r.cursor != 1 {
		t.Errorf("cursor: want 1, got %d", r.cursor)
	}
}

func TestProcessControlKey_NonPrintable_NoEffect(t *testing.T) {
	// 0x00 (NUL) is not printable and not handled — should be a no-op
	r := processControlKey(0x00, []rune("hello"), 2)
	if string(r.buffer) != "hello" {
		t.Errorf("buffer must not change for NUL, got %q", string(r.buffer))
	}
	if r.cursor != 2 {
		t.Errorf("cursor must not move for NUL, got %d", r.cursor)
	}
	if r.done || r.ctrlC {
		t.Error("NUL must not trigger done or ctrlC")
	}
}

func TestProcessControlKey_EnterOnEmptyBuffer(t *testing.T) {
	r := processControlKey(0x0d, []rune{}, 0)
	if !r.done {
		t.Error("expected done=true for Enter on empty buffer")
	}
	if r.output != "" {
		t.Errorf("output: want empty string, got %q", r.output)
	}
}
