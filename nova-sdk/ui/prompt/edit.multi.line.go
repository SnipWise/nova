package prompt

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

// MultiLineEditor represents a multi-line text editor state
type MultiLineEditor struct {
	lines         [][]rune // Buffer as lines of runes
	cursorLine    int      // Current line number (0-based)
	cursorCol     int      // Current column in the line (0-based)
	cursorVisible bool     // Cursor visibility for blinking
	cursorStyle   CursorStyle
	cursorMutex   sync.Mutex
	stopBlink     chan bool
	needsRender   chan bool // Signal that a render is needed (for blinking)
}

// NewMultiLineEditor creates a new multi-line editor with optional default text
func NewMultiLineEditor(defaultValue string, cursorStyle CursorStyle) *MultiLineEditor {
	editor := &MultiLineEditor{
		lines:         make([][]rune, 0),
		cursorLine:    0,
		cursorCol:     0,
		cursorVisible: true,
		cursorStyle:   cursorStyle,
	}

	// Initialize with default value if provided
	if defaultValue != "" {
		lines := strings.Split(defaultValue, "\n")
		for _, line := range lines {
			editor.lines = append(editor.lines, []rune(line))
		}
		// Position cursor at the end
		editor.cursorLine = len(editor.lines) - 1
		editor.cursorCol = len(editor.lines[editor.cursorLine])
	} else {
		// Start with one empty line
		editor.lines = append(editor.lines, []rune{})
	}

	return editor
}

// StartBlinking starts the cursor blinking effect if needed
func (e *MultiLineEditor) StartBlinking() {
	if e.cursorStyle == CursorBlockBlink || e.cursorStyle == CursorUnderlineBlink {
		e.stopBlink = make(chan bool)
		e.needsRender = make(chan bool, 1) // Buffered to avoid blocking
		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					e.cursorMutex.Lock()
					e.cursorVisible = !e.cursorVisible
					e.cursorMutex.Unlock()
					// Signal that a render is needed
					select {
					case e.needsRender <- true:
					default: // Don't block if channel is full
					}
				case <-e.stopBlink:
					return
				}
			}
		}()
	}
}

// StopBlinking stops the cursor blinking effect
func (e *MultiLineEditor) StopBlinking() {
	if e.stopBlink != nil {
		close(e.stopBlink)
		e.stopBlink = nil
	}
}

// GetText returns the complete text as a single string with newlines
func (e *MultiLineEditor) GetText() string {
	var result strings.Builder
	for i, line := range e.lines {
		result.WriteString(string(line))
		if i < len(e.lines)-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// InsertRune inserts a rune at the current cursor position
func (e *MultiLineEditor) InsertRune(r rune) {
	line := e.lines[e.cursorLine]
	e.lines[e.cursorLine] = append(line[:e.cursorCol], append([]rune{r}, line[e.cursorCol:]...)...)
	e.cursorCol++
}

// InsertNewLine inserts a new line at the current cursor position
func (e *MultiLineEditor) InsertNewLine() {
	currentLine := e.lines[e.cursorLine]

	// Split the current line at cursor position
	beforeCursor := make([]rune, e.cursorCol)
	copy(beforeCursor, currentLine[:e.cursorCol])

	afterCursor := make([]rune, len(currentLine)-e.cursorCol)
	copy(afterCursor, currentLine[e.cursorCol:])

	// Update current line with text before cursor
	e.lines[e.cursorLine] = beforeCursor

	// Insert new line with text after cursor
	newLines := make([][]rune, len(e.lines)+1)
	copy(newLines, e.lines[:e.cursorLine+1])
	newLines[e.cursorLine+1] = afterCursor
	copy(newLines[e.cursorLine+2:], e.lines[e.cursorLine+1:])
	e.lines = newLines

	// Move cursor to beginning of new line
	e.cursorLine++
	e.cursorCol = 0
}

// Backspace deletes the character before the cursor
func (e *MultiLineEditor) Backspace() {
	if e.cursorCol > 0 {
		// Delete character in current line
		line := e.lines[e.cursorLine]
		e.lines[e.cursorLine] = append(line[:e.cursorCol-1], line[e.cursorCol:]...)
		e.cursorCol--
	} else if e.cursorLine > 0 {
		// Merge with previous line
		prevLine := e.lines[e.cursorLine-1]
		currentLine := e.lines[e.cursorLine]

		// New cursor position will be at the end of previous line
		e.cursorCol = len(prevLine)

		// Merge lines
		e.lines[e.cursorLine-1] = append(prevLine, currentLine...)

		// Remove current line
		e.lines = append(e.lines[:e.cursorLine], e.lines[e.cursorLine+1:]...)

		// Move cursor up
		e.cursorLine--
	}
}

// Delete deletes the character at the cursor
func (e *MultiLineEditor) Delete() {
	line := e.lines[e.cursorLine]
	if e.cursorCol < len(line) {
		e.lines[e.cursorLine] = append(line[:e.cursorCol], line[e.cursorCol+1:]...)
	} else if e.cursorLine < len(e.lines)-1 {
		// Merge with next line
		nextLine := e.lines[e.cursorLine+1]
		e.lines[e.cursorLine] = append(line, nextLine...)
		e.lines = append(e.lines[:e.cursorLine+1], e.lines[e.cursorLine+2:]...)
	}
}

// MoveLeft moves the cursor one position to the left
func (e *MultiLineEditor) MoveLeft() {
	if e.cursorCol > 0 {
		e.cursorCol--
	} else if e.cursorLine > 0 {
		// Move to end of previous line
		e.cursorLine--
		e.cursorCol = len(e.lines[e.cursorLine])
	}
}

// MoveRight moves the cursor one position to the right
func (e *MultiLineEditor) MoveRight() {
	if e.cursorCol < len(e.lines[e.cursorLine]) {
		e.cursorCol++
	} else if e.cursorLine < len(e.lines)-1 {
		// Move to beginning of next line
		e.cursorLine++
		e.cursorCol = 0
	}
}

// MoveUp moves the cursor one line up
func (e *MultiLineEditor) MoveUp() {
	if e.cursorLine > 0 {
		e.cursorLine--
		// Adjust column if new line is shorter
		if e.cursorCol > len(e.lines[e.cursorLine]) {
			e.cursorCol = len(e.lines[e.cursorLine])
		}
	}
}

// MoveDown moves the cursor one line down
func (e *MultiLineEditor) MoveDown() {
	if e.cursorLine < len(e.lines)-1 {
		e.cursorLine++
		// Adjust column if new line is shorter
		if e.cursorCol > len(e.lines[e.cursorLine]) {
			e.cursorCol = len(e.lines[e.cursorLine])
		}
	}
}

// MoveHome moves the cursor to the beginning of the current line
func (e *MultiLineEditor) MoveHome() {
	e.cursorCol = 0
}

// MoveEnd moves the cursor to the end of the current line
func (e *MultiLineEditor) MoveEnd() {
	e.cursorCol = len(e.lines[e.cursorLine])
}

// KillToEnd deletes from cursor to end of line
func (e *MultiLineEditor) KillToEnd() {
	e.lines[e.cursorLine] = e.lines[e.cursorLine][:e.cursorCol]
}

// KillToStart deletes from start of line to cursor
func (e *MultiLineEditor) KillToStart() {
	line := e.lines[e.cursorLine]
	e.lines[e.cursorLine] = line[e.cursorCol:]
	e.cursorCol = 0
}

// renderMultiLine renders the multi-line editor
// Returns the cursor line position after render
func renderMultiLine(editor *MultiLineEditor, previousRenderedLines int, previousCursorLine int) (int, int) {
	// Calculate where we currently are
	// After the last render, we were on line previousCursorLine
	// We need to move to line 0

	// Move up to line 0
	if previousCursorLine > 0 {
		fmt.Printf("\033[%dA", previousCursorLine)
	}

	// Move to start of line
	fmt.Print(carriageReturn)

	// Now render all lines
	for lineNum, line := range editor.lines {
		// Clear the current line
		fmt.Print(clearToEnd)

		if lineNum == editor.cursorLine {
			// Render line with cursor
			renderLineWithCursor(line, editor.cursorCol, editor.cursorVisible, editor.cursorStyle)
		} else {
			// Render line without cursor
			fmt.Print(string(line))
		}

		// Move to next line if not the last one
		if lineNum < len(editor.lines)-1 {
			fmt.Print("\r\n")
		}
	}

	// Clear everything below
	fmt.Print("\033[J")

	// After rendering all lines, the cursor is at the end of the last line (line len(editor.lines)-1)
	// But we need to position it on editor.cursorLine
	// So we need to move up from the last line to editor.cursorLine
	linesToMoveUp := len(editor.lines) - 1 - editor.cursorLine
	if linesToMoveUp > 0 {
		fmt.Printf("\033[%dA", linesToMoveUp)
	}

	// Return where we are now: total lines rendered and cursor line position
	return len(editor.lines), editor.cursorLine
}

// renderLineWithCursor renders a single line with the cursor
func renderLineWithCursor(line []rune, cursorCol int, cursorVisible bool, cursorStyle CursorStyle) {
	// Display text before cursor
	if cursorCol > 0 {
		fmt.Print(string(line[:cursorCol]))
	}

	// Display cursor based on style and visibility
	if cursorVisible || (cursorStyle != CursorBlockBlink && cursorStyle != CursorUnderlineBlink) {
		switch cursorStyle {
		case CursorBlock, CursorBlockBlink:
			if cursorCol < len(line) {
				// Show cursor as inverse video (background highlight) of the character
				fmt.Printf("\033[7m%c\033[0m", line[cursorCol])
			} else {
				// Cursor at the end - show a block cursor
				fmt.Print("\033[7m \033[0m")
			}

		case CursorUnderline, CursorUnderlineBlink:
			if cursorCol < len(line) {
				// Show cursor as underlined character
				fmt.Printf("\033[4m%c\033[0m", line[cursorCol])
			} else {
				// Cursor at the end - show an underscore
				fmt.Print("\033[4m \033[0m")
			}
		}
	} else {
		// Cursor is hidden (for blinking effect) - just print the character without styling
		if cursorCol < len(line) {
			fmt.Printf("%c", line[cursorCol])
		} else {
			fmt.Print(" ")
		}
	}

	// Print remaining text after cursor
	if cursorCol+1 < len(line) {
		fmt.Print(string(line[cursorCol+1:]))
	}
}

// editMultiLine provides an interactive multi-line editor
func editMultiLine(prompt string, defaultValue string, cursorStyle CursorStyle) (string, error) {
	// Enable raw mode
	restoreCmd, err := enableRawMode()
	if err != nil {
		return "", err
	}
	defer disableRawMode(restoreCmd)

	// Print prompt
	fmt.Print(prompt)

	// Hide the system cursor since we're drawing our own
	fmt.Print(hideCursor)
	defer fmt.Print(showCursor)

	// Create editor
	editor := NewMultiLineEditor(defaultValue, cursorStyle)

	// Start cursor blinking if needed
	editor.StartBlinking()
	defer editor.StopBlinking()

	// Track state from last render
	var linesRendered, cursorLineAfterRender int

	// Render the initial state
	linesRendered, cursorLineAfterRender = renderMultiLine(editor, 0, 0)

	// Channel to receive input bytes
	inputChan := make(chan byte)
	go func() {
		inputBuf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(inputBuf)
			if err != nil || n == 0 {
				continue
			}
			inputChan <- inputBuf[0]
		}
	}()

	// Read input byte by byte
	escapeSeq := make([]byte, 0, 10)
	inEscape := false

	for {
		var b byte

		// Wait for either input or blink render request
		if editor.needsRender != nil {
			select {
			case b = <-inputChan:
				// Got input from stdin
			case <-editor.needsRender:
				// Cursor blink triggered a render
				editor.cursorMutex.Lock()
				linesRendered, cursorLineAfterRender = renderMultiLine(editor, linesRendered, cursorLineAfterRender)
				editor.cursorMutex.Unlock()
				continue
			}
		} else {
			// No blinking, just wait for input
			b = <-inputChan
		}

		// Reset cursor visibility on any keypress
		editor.cursorMutex.Lock()
		editor.cursorVisible = true

		// Handle escape sequences
		if inEscape {
			escapeSeq = append(escapeSeq, b)

			// Check if we have a complete escape sequence
			if len(escapeSeq) >= 2 {
				seq := string(escapeSeq)

				// Arrow keys and other special keys
				if strings.HasPrefix(seq, "[") {
					switch {
					case strings.HasSuffix(seq, "A"): // Up arrow
						editor.MoveUp()
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "B"): // Down arrow
						editor.MoveDown()
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "D"): // Left arrow
						editor.MoveLeft()
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "C"): // Right arrow
						editor.MoveRight()
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "H"): // Home
						editor.MoveHome()
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "F"): // End
						editor.MoveEnd()
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "~"): // Delete, Page Up/Down, etc.
						if strings.HasPrefix(seq, "[3~") { // Delete
							editor.Delete()
						}
						inEscape = false
						escapeSeq = escapeSeq[:0]
					}
				}
			}

			linesRendered, cursorLineAfterRender = renderMultiLine(editor, linesRendered, cursorLineAfterRender)
			editor.cursorMutex.Unlock()
			continue
		}

		// Check for escape sequence start
		if b == 0x1b { // ESC
			inEscape = true
			escapeSeq = escapeSeq[:0]
			editor.cursorMutex.Unlock()
			continue
		}

		// Handle special characters
		switch b {
		case 0x03: // Ctrl+C
			editor.cursorMutex.Unlock()
			fmt.Println()
			disableRawMode(restoreCmd)
			os.Exit(0)

		case 0x04: // Ctrl+D (Submit/EOF)
			editor.cursorMutex.Unlock()
			fmt.Println()
			return editor.GetText(), nil

		case 0x0d, 0x0a: // Enter (new line)
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

		case 0x0b: // Ctrl+K (Kill line from cursor)
			editor.KillToEnd()

		case 0x15: // Ctrl+U (Kill line before cursor)
			editor.KillToStart()

		case 0x10: // Ctrl+P (Previous line / Up)
			editor.MoveUp()

		case 0x0e: // Ctrl+N (Next line / Down)
			editor.MoveDown()

		default:
			// Insert printable characters
			if unicode.IsPrint(rune(b)) {
				editor.InsertRune(rune(b))
			}
		}

		linesRendered, cursorLineAfterRender = renderMultiLine(editor, linesRendered, cursorLineAfterRender)
		editor.cursorMutex.Unlock()
	}
}

// RunWithMultiLineEdit displays the prompt with full multi-line editing support
// Press Ctrl+D to submit the input
// Allows using arrow keys, Home, End, Delete, and various Ctrl shortcuts
func (i *ColorInput) RunWithMultiLineEdit() (string, error) {
	cursorStyle := i.getCursorStyle()

	for {
		// Build the prompt string
		var promptStr string
		if i.defaultValue != "" {
			promptStr = fmt.Sprintf("%s%s %s%s %s[%s]%s %s(Ctrl+D to submit)%s\n",
				i.messageColor, i.promptSymbol, i.message, ColorReset,
				i.defaultColor, i.defaultValue, ColorReset,
				i.defaultColor, ColorReset)
		} else {
			promptStr = fmt.Sprintf("%s%s %s%s %s(Ctrl+D to submit)%s\n",
				i.messageColor, i.promptSymbol, i.message, ColorReset,
				i.defaultColor, ColorReset)
		}

		// Use the multi-line editor
		input, err := editMultiLine(promptStr, i.defaultValue, cursorStyle)
		fmt.Print(ColorReset) // Reset color after input

		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		// Clean the input (trim leading/trailing whitespace from entire text)
		input = strings.TrimSpace(input)

		// Use default value if input is empty
		if input == "" && i.defaultValue != "" {
			input = i.defaultValue
		}

		// Validate if a validator is set
		if i.validator != nil {
			if err := i.validator(input); err != nil {
				fmt.Printf("%s%s %s%s\n", i.errorColor, i.errorSymbol, err.Error(), ColorReset)
				continue
			}
		}

		fmt.Print(carriageReturn)
		fmt.Print(ANSIReset)
		return input, nil
	}
}
