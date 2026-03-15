package prompt

import (
	"os"
	"strings"
	"unicode"
)

// multiLineKeyResult holds the outcome of processing a key press in the multi-line editor.
// exit=true means Ctrl+C was pressed; the caller must call os.Exit.
// done=true means Ctrl+D was pressed; text contains the editor content to return.
type multiLineKeyResult struct {
	exit bool
	done bool
	text string
}

// appendAndDispatchEscape appends b to escapeSeq and dispatches the sequence when complete.
// Returns the updated slice and done=true when the sequence was consumed (caller resets inEscape).
// A "complete" sequence ends with a letter or '~'; an unknown prefix also resets the accumulator.
func appendAndDispatchEscape(escapeSeq []byte, b byte, editor *MultiLineEditor) ([]byte, bool) {
	escapeSeq = append(escapeSeq, b)
	if len(escapeSeq) < 2 {
		return escapeSeq, false
	}
	seq := string(escapeSeq)
	if !strings.HasPrefix(seq, "[") {
		return escapeSeq[:0], true // unrecognised prefix — reset
	}
	switch {
	case strings.HasSuffix(seq, "A"): // Up arrow
		editor.MoveUp()
	case strings.HasSuffix(seq, "B"): // Down arrow
		editor.MoveDown()
	case strings.HasSuffix(seq, "D"): // Left arrow
		editor.MoveLeft()
	case strings.HasSuffix(seq, "C"): // Right arrow
		editor.MoveRight()
	case strings.HasSuffix(seq, "H"): // Home
		editor.MoveHome()
	case strings.HasSuffix(seq, "F"): // End
		editor.MoveEnd()
	case strings.HasSuffix(seq, "~"): // Delete / Page Up/Down
		if strings.HasPrefix(seq, "[3~") {
			editor.Delete()
		}
	default:
		return escapeSeq, false // incomplete — keep accumulating
	}
	return escapeSeq[:0], true
}

// processMultiLineKey handles a single control byte or printable character.
// For Ctrl+C it returns exit=true without calling os.Exit, letting the caller clean up.
// For Ctrl+D it returns done=true with the current editor text.
func processMultiLineKey(b byte, editor *MultiLineEditor) multiLineKeyResult {
	switch b {
	case 0x03: // Ctrl+C
		return multiLineKeyResult{exit: true}

	case 0x04: // Ctrl+D (submit / EOF)
		return multiLineKeyResult{done: true, text: editor.GetText()}

	case 0x0d, 0x0a: // Enter (CR or LF)
		editor.InsertNewLine()

	case 0x7f, 0x08: // Backspace
		editor.Backspace()

	case 0x01: // Ctrl+A (Home)
		editor.MoveHome()

	case 0x05: // Ctrl+E (End)
		editor.MoveEnd()

	case 0x02: // Ctrl+B (Left)
		editor.MoveLeft()

	case 0x06: // Ctrl+F (Right)
		editor.MoveRight()

	case 0x0b: // Ctrl+K (kill to end of line)
		editor.KillToEnd()

	case 0x15: // Ctrl+U (kill to start of line)
		editor.KillToStart()

	case 0x10: // Ctrl+P (previous line / Up)
		editor.MoveUp()

	case 0x0e: // Ctrl+N (next line / Down)
		editor.MoveDown()

	default:
		if unicode.IsPrint(rune(b)) {
			editor.InsertRune(rune(b))
		}
	}
	return multiLineKeyResult{}
}

// startInputReader launches a goroutine that reads stdin byte-by-byte and sends each byte
// to the returned channel. The goroutine runs until the process exits.
func startInputReader() chan byte {
	inputChan := make(chan byte)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				continue
			}
			inputChan <- buf[0]
		}
	}()
	return inputChan
}

// readInputOrBlink waits for either a keypress or a blink-render signal.
// When a blink signal fires, it renders and returns (0, true) so the caller can continue.
// When a keypress arrives, it returns (b, false).
func readInputOrBlink(
	inputChan chan byte,
	editor *MultiLineEditor,
	linesRendered *int,
	cursorLineAfterRender *int,
) (byte, bool) {
	if editor.needsRender == nil {
		return <-inputChan, false
	}
	select {
	case b := <-inputChan:
		return b, false
	case <-editor.needsRender:
		editor.cursorMutex.Lock()
		*linesRendered, *cursorLineAfterRender = renderMultiLine(editor, *linesRendered, *cursorLineAfterRender)
		editor.cursorMutex.Unlock()
		return 0, true
	}
}
