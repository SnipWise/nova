package prompt

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// capturePromptOutput captures text written to os.Stdout during f().
func capturePromptOutput(f func()) string {
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

// ── renderCursorChar ──────────────────────────────────────────────────────────

func TestRenderCursorChar_Block_MidBuffer(t *testing.T) {
	buf := []rune("hello")
	out := capturePromptOutput(func() { renderCursorChar(buf, 2, CursorBlock) })
	// Should contain the character at position 2 ('l') wrapped in inverse-video codes
	if !strings.Contains(out, "l") {
		t.Errorf("expected character 'l', got %q", out)
	}
	if !strings.Contains(out, "\033[7m") {
		t.Errorf("CursorBlock: expected inverse-video escape \\033[7m, got %q", out)
	}
}

func TestRenderCursorChar_Block_AtEnd(t *testing.T) {
	buf := []rune("hello")
	out := capturePromptOutput(func() { renderCursorChar(buf, 5, CursorBlock) })
	// At end: block cursor shown as a space with inverse video
	if !strings.Contains(out, "\033[7m") {
		t.Errorf("CursorBlock at end: expected \\033[7m, got %q", out)
	}
	if !strings.Contains(out, " ") {
		t.Errorf("CursorBlock at end: expected space, got %q", out)
	}
}

func TestRenderCursorChar_BlockBlink_MidBuffer(t *testing.T) {
	buf := []rune("world")
	out := capturePromptOutput(func() { renderCursorChar(buf, 1, CursorBlockBlink) })
	if !strings.Contains(out, "o") {
		t.Errorf("CursorBlockBlink: expected 'o', got %q", out)
	}
	if !strings.Contains(out, "\033[7m") {
		t.Errorf("CursorBlockBlink: expected inverse-video, got %q", out)
	}
}

func TestRenderCursorChar_Underline_MidBuffer(t *testing.T) {
	buf := []rune("hello")
	out := capturePromptOutput(func() { renderCursorChar(buf, 0, CursorUnderline) })
	if !strings.Contains(out, "h") {
		t.Errorf("CursorUnderline: expected 'h', got %q", out)
	}
	if !strings.Contains(out, "\033[4m") {
		t.Errorf("CursorUnderline: expected underline escape \\033[4m, got %q", out)
	}
}

func TestRenderCursorChar_Underline_AtEnd(t *testing.T) {
	buf := []rune("hello")
	out := capturePromptOutput(func() { renderCursorChar(buf, 5, CursorUnderline) })
	if !strings.Contains(out, "\033[4m") {
		t.Errorf("CursorUnderline at end: expected \\033[4m, got %q", out)
	}
}

func TestRenderCursorChar_UnderlineBlink_MidBuffer(t *testing.T) {
	buf := []rune("abc")
	out := capturePromptOutput(func() { renderCursorChar(buf, 2, CursorUnderlineBlink) })
	if !strings.Contains(out, "c") {
		t.Errorf("CursorUnderlineBlink: expected 'c', got %q", out)
	}
	if !strings.Contains(out, "\033[4m") {
		t.Errorf("CursorUnderlineBlink: expected underline, got %q", out)
	}
}

func TestRenderCursorChar_EmptyBuffer(t *testing.T) {
	// Empty buffer + cursor at 0 → cursor at end → show styled space
	out := capturePromptOutput(func() { renderCursorChar([]rune{}, 0, CursorBlock) })
	if !strings.Contains(out, "\033[7m") {
		t.Errorf("empty buffer CursorBlock: expected inverse-video, got %q", out)
	}
}

// ── renderHiddenCursorChar ────────────────────────────────────────────────────

func TestRenderHiddenCursorChar_MidBuffer(t *testing.T) {
	buf := []rune("hello")
	out := capturePromptOutput(func() { renderHiddenCursorChar(buf, 2) })
	// Should print the character without any escape codes
	if out != "l" {
		t.Errorf("hidden cursor mid: want %q, got %q", "l", out)
	}
}

func TestRenderHiddenCursorChar_AtStart(t *testing.T) {
	buf := []rune("hello")
	out := capturePromptOutput(func() { renderHiddenCursorChar(buf, 0) })
	if out != "h" {
		t.Errorf("hidden cursor at start: want %q, got %q", "h", out)
	}
}

func TestRenderHiddenCursorChar_AtEnd(t *testing.T) {
	buf := []rune("hello")
	out := capturePromptOutput(func() { renderHiddenCursorChar(buf, 5) })
	// At end: prints a plain space
	if out != " " {
		t.Errorf("hidden cursor at end: want %q, got %q", " ", out)
	}
}

func TestRenderHiddenCursorChar_EmptyBuffer(t *testing.T) {
	out := capturePromptOutput(func() { renderHiddenCursorChar([]rune{}, 0) })
	if out != " " {
		t.Errorf("hidden cursor empty buffer: want %q, got %q", " ", out)
	}
}

func TestRenderHiddenCursorChar_NoEscapeCodes(t *testing.T) {
	buf := []rune("test")
	out := capturePromptOutput(func() { renderHiddenCursorChar(buf, 1) })
	if strings.Contains(out, "\033") {
		t.Errorf("hidden cursor must not emit escape codes, got %q", out)
	}
}

// ── startBlinker ──────────────────────────────────────────────────────────────

func TestStartBlinker_NonBlinkStyle_ReturnsNil(t *testing.T) {
	buf := []rune("hello")
	cursor := 0
	visible := true
	var mu sync.Mutex

	ch := startBlinker(CursorBlock, &buf, &cursor, &visible, &mu)
	if ch != nil {
		close(ch)
		t.Error("non-blink style: want nil channel")
	}

	ch2 := startBlinker(CursorUnderline, &buf, &cursor, &visible, &mu)
	if ch2 != nil {
		close(ch2)
		t.Error("CursorUnderline: want nil channel")
	}
}

func TestStartBlinker_BlinkStyle_ReturnsChannel(t *testing.T) {
	buf := []rune("hi")
	cursor := 0
	visible := true
	var mu sync.Mutex

	ch := startBlinker(CursorBlockBlink, &buf, &cursor, &visible, &mu)
	if ch == nil {
		t.Fatal("CursorBlockBlink: want non-nil channel")
	}
	// Give goroutine a moment to start, then stop it cleanly
	time.Sleep(10 * time.Millisecond)
	close(ch)
}

func TestStartBlinker_UnderlineBlink_ReturnsChannel(t *testing.T) {
	buf := []rune("hi")
	cursor := 0
	visible := true
	var mu sync.Mutex

	ch := startBlinker(CursorUnderlineBlink, &buf, &cursor, &visible, &mu)
	if ch == nil {
		t.Fatal("CursorUnderlineBlink: want non-nil channel")
	}
	close(ch)
}
