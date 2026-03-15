package prompt

import (
	"fmt"
	"strings"
	"unicode"
)

// processEscapeSeq processes a complete ANSI escape sequence.
// Returns the updated buffer, cursor, and whether the sequence was recognized and consumed.
// If consumed is false, the caller should keep accumulating bytes.
func processEscapeSeq(seq string, buffer []rune, cursor int) ([]rune, int, bool) {
	if !strings.HasPrefix(seq, "[") {
		return buffer, cursor, false
	}
	switch {
	case strings.HasSuffix(seq, "D"): // Left arrow
		if cursor > 0 {
			cursor--
		}
	case strings.HasSuffix(seq, "C"): // Right arrow
		if cursor < len(buffer) {
			cursor++
		}
	case strings.HasSuffix(seq, "H"): // Home
		cursor = 0
	case strings.HasSuffix(seq, "F"): // End
		cursor = len(buffer)
	case strings.HasSuffix(seq, "~"): // Delete, Page Up/Down, etc.
		if strings.HasPrefix(seq, "[3~") { // Delete key
			if cursor < len(buffer) {
				buffer = append(buffer[:cursor], buffer[cursor+1:]...)
			}
		}
	default:
		return buffer, cursor, false
	}
	return buffer, cursor, true
}

// escapeResult holds the outcome of processEscapeInput.
// handled=true means the byte was consumed by escape logic; the caller should
// continue to the next loop iteration.
// render=true means the caller should renderLine before continuing.
type escapeResult struct {
	buffer  []rune
	cursor  int
	handled bool
	render  bool
}

// processEscapeInput handles one byte of escape-sequence input.
// It manages the inEscape flag and accumulates bytes in escapeSeq until a
// complete sequence is recognised by processEscapeSeq.
// Returns an escapeResult indicating whether the byte was consumed.
func processEscapeInput(b byte, inEscape *bool, escapeSeq *[]byte, buffer []rune, cursor int) escapeResult {
	if *inEscape {
		*escapeSeq = append(*escapeSeq, b)
		if len(*escapeSeq) >= 2 {
			var consumed bool
			buffer, cursor, consumed = processEscapeSeq(string(*escapeSeq), buffer, cursor)
			if consumed {
				*inEscape = false
				*escapeSeq = (*escapeSeq)[:0]
			}
		}
		return escapeResult{buffer: buffer, cursor: cursor, handled: true, render: true}
	}
	if b == 0x1b {
		*inEscape = true
		*escapeSeq = (*escapeSeq)[:0]
		return escapeResult{buffer: buffer, cursor: cursor, handled: true, render: false}
	}
	return escapeResult{buffer: buffer, cursor: cursor}
}

// keyResult holds the outcome of processing a key press.
type keyResult struct {
	buffer []rune
	cursor int
	done   bool   // true if the edit loop should terminate (Enter, Ctrl+D+empty)
	ctrlC  bool   // true if Ctrl+C was pressed; caller must handle os.Exit
	output string // value to return when done=true and err=nil
	err    error  // error to return when done=true
}

// processControlKey handles a single control byte or printable character.
// For Ctrl+C it sets ctrlC=true and returns without calling os.Exit,
// allowing the caller to clean up before exiting.
func processControlKey(b byte, buffer []rune, cursor int) keyResult {
	switch b {
	case 0x03: // Ctrl+C
		return keyResult{buffer: buffer, cursor: cursor, ctrlC: true}

	case 0x04: // Ctrl+D (EOF)
		if len(buffer) == 0 {
			return keyResult{buffer: buffer, cursor: cursor, done: true, err: fmt.Errorf("EOF")}
		}

	case 0x0d, 0x0a: // Enter (CR or LF)
		return keyResult{buffer: buffer, cursor: cursor, done: true, output: string(buffer)}

	case 0x7f, 0x08: // Backspace
		if cursor > 0 {
			buffer = append(buffer[:cursor-1], buffer[cursor:]...)
			cursor--
		}

	case 0x01: // Ctrl+A (Home)
		cursor = 0

	case 0x05: // Ctrl+E (End)
		cursor = len(buffer)

	case 0x02: // Ctrl+B (move left one char)
		if cursor > 0 {
			cursor--
		}

	case 0x06: // Ctrl+F (move right one char)
		if cursor < len(buffer) {
			cursor++
		}

	case 0x0b: // Ctrl+K (kill to end of line)
		buffer = buffer[:cursor]

	case 0x15: // Ctrl+U (kill to start of line)
		buffer = buffer[cursor:]
		cursor = 0

	default:
		if unicode.IsPrint(rune(b)) {
			buffer = append(buffer[:cursor], append([]rune{rune(b)}, buffer[cursor:]...)...)
			cursor++
		}
	}

	return keyResult{buffer: buffer, cursor: cursor}
}
