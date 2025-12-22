package prompt

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"
)

// Terminal control sequences
const (
	// Cursor movement
	cursorLeft  = "\033[D"
	cursorRight = "\033[C"
	cursorHome  = "\033[H"
	cursorEnd   = "\033[F"

	// Line editing
	clearLine      = "\033[2K"
	clearToEnd     = "\033[K"
	saveCursor     = "\033[s"
	restoreCursor  = "\033[u"
	hideCursor     = "\033[?25l"
	showCursor     = "\033[?25h"
	carriageReturn = "\r"

	ANSIReset = "\033[0m" // Reset all attributes 
)

// CursorStyle defines how the cursor should be displayed
type CursorStyle int

const (
	// CursorBlock displays a solid block cursor (default)
	CursorBlock CursorStyle = iota
	// CursorBlockBlink displays a blinking block cursor
	CursorBlockBlink
	// CursorUnderline displays an underline cursor
	CursorUnderline
	// CursorUnderlineBlink displays a blinking underline cursor
	CursorUnderlineBlink
)

var (
	// Default cursor style
	defaultCursorStyle = CursorBlock
)

// enableRawMode enables raw mode on the terminal
func enableRawMode() (*exec.Cmd, error) {
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("raw mode not supported on Windows")
	}

	// Save current terminal state
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = os.Stdin
	savedState, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Enable raw mode
	rawCmd := exec.Command("stty", "raw", "-echo")
	rawCmd.Stdin = os.Stdin
	if err := rawCmd.Run(); err != nil {
		return nil, err
	}

	// Return command to restore state
	restoreCmd := exec.Command("stty", strings.TrimSpace(string(savedState)))
	restoreCmd.Stdin = os.Stdin
	return restoreCmd, nil
}

// disableRawMode disables raw mode and restores terminal
func disableRawMode(restoreCmd *exec.Cmd) {
	if restoreCmd != nil {
		restoreCmd.Run()
	}
}

// editLine provides an interactive line editor with arrow key support
func editLine(prompt string, defaultValue string, cursorStyle CursorStyle) (string, error) {
	// Enable raw mode
	restoreCmd, err := enableRawMode()
	if err != nil {
		return "", err
	}
	defer disableRawMode(restoreCmd)

	// Print prompt and save the position
	fmt.Print(prompt)
	fmt.Print(saveCursor) // Save cursor position after prompt

	// Hide the system cursor since we're drawing our own
	fmt.Print(hideCursor)
	defer fmt.Print(showCursor)

	// Initialize with default value if provided
	buffer := []rune(defaultValue)
	cursor := len(buffer)

	// Cursor blinking state
	var cursorVisible bool = true
	var cursorMutex sync.Mutex
	var stopBlink chan bool

	// Start cursor blinking if needed
	if cursorStyle == CursorBlockBlink || cursorStyle == CursorUnderlineBlink {
		stopBlink = make(chan bool)
		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					cursorMutex.Lock()
					cursorVisible = !cursorVisible
					renderLine(buffer, cursor, cursorVisible, cursorStyle)
					cursorMutex.Unlock()
				case <-stopBlink:
					return
				}
			}
		}()
	}

	// Render the initial line
	renderLine(buffer, cursor, cursorVisible, cursorStyle)

	// Read input byte by byte
	inputBuf := make([]byte, 1)
	escapeSeq := make([]byte, 0, 10)
	inEscape := false

	defer func() {
		if stopBlink != nil {
			close(stopBlink)
		}
	}()

	for {
		n, err := os.Stdin.Read(inputBuf)
		if err != nil || n == 0 {
			continue
		}

		b := inputBuf[0]

		// Reset cursor visibility on any keypress
		cursorMutex.Lock()
		cursorVisible = true

		// Handle escape sequences
		if inEscape {
			escapeSeq = append(escapeSeq, b)

			// Check if we have a complete escape sequence
			if len(escapeSeq) >= 2 {
				seq := string(escapeSeq)

				// Arrow keys and other special keys
				if strings.HasPrefix(seq, "[") {
					switch {
					case strings.HasSuffix(seq, "D"): // Left arrow
						if cursor > 0 {
							cursor--
						}
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "C"): // Right arrow
						if cursor < len(buffer) {
							cursor++
						}
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "H"): // Home
						cursor = 0
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "F"): // End
						cursor = len(buffer)
						inEscape = false
						escapeSeq = escapeSeq[:0]

					case strings.HasSuffix(seq, "~"): // Delete, Page Up/Down, etc.
						if strings.HasPrefix(seq, "[3~") { // Delete
							if cursor < len(buffer) {
								buffer = append(buffer[:cursor], buffer[cursor+1:]...)
							}
						}
						inEscape = false
						escapeSeq = escapeSeq[:0]
					}
				}
			}

			renderLine(buffer, cursor, cursorVisible, cursorStyle)
			cursorMutex.Unlock()
			continue
		}

		// Check for escape sequence start
		if b == 0x1b { // ESC
			inEscape = true
			escapeSeq = escapeSeq[:0]
			cursorMutex.Unlock()
			continue
		}

		// Handle special characters
		switch b {
		case 0x03: // Ctrl+C
			cursorMutex.Unlock()
			fmt.Println()
			disableRawMode(restoreCmd)
			os.Exit(0)

		case 0x04: // Ctrl+D (EOF)
			if len(buffer) == 0 {
				cursorMutex.Unlock()
				fmt.Println()
				return "", fmt.Errorf("EOF")
			}

		case 0x0d, 0x0a: // Enter (CR or LF)
			cursorMutex.Unlock()
			fmt.Println()
			return string(buffer), nil

		case 0x7f, 0x08: // Backspace or Delete
			if cursor > 0 {
				buffer = append(buffer[:cursor-1], buffer[cursor:]...)
				cursor--
			}

		case 0x01: // Ctrl+A (Home)
			cursor = 0

		case 0x05: // Ctrl+E (End)
			cursor = len(buffer)

		case 0x02: // Ctrl+B (Left)
			if cursor > 0 {
				cursor--
			}

		case 0x06: // Ctrl+F (Right)
			if cursor < len(buffer) {
				cursor++
			}

		case 0x0b: // Ctrl+K (Kill line from cursor)
			buffer = buffer[:cursor]

		case 0x15: // Ctrl+U (Kill line before cursor)
			buffer = buffer[cursor:]
			cursor = 0

		default:
			// Insert printable characters
			if unicode.IsPrint(rune(b)) {
				// Insert at cursor position
				buffer = append(buffer[:cursor], append([]rune{rune(b)}, buffer[cursor:]...)...)
				cursor++
			}
		}

		renderLine(buffer, cursor, cursorVisible, cursorStyle)
		cursorMutex.Unlock()
	}
}

// renderLine renders the current line with the cursor at the correct position
func renderLine(buffer []rune, cursor int, cursorVisible bool, cursorStyle CursorStyle) {
	// Restore cursor position (after the prompt) and clear from there to end of line
	fmt.Print(restoreCursor + clearToEnd)

	// Display text before cursor
	if cursor > 0 {
		fmt.Print(string(buffer[:cursor]))
	}

	// Display cursor based on style and visibility
	if cursorVisible || (cursorStyle != CursorBlockBlink && cursorStyle != CursorUnderlineBlink) {
		switch cursorStyle {
		case CursorBlock, CursorBlockBlink:
			if cursor < len(buffer) {
				// Show cursor as inverse video (background highlight) of the character
				fmt.Printf("\033[7m%c\033[0m", buffer[cursor])
			} else {
				// Cursor at the end - show a block cursor
				fmt.Print("\033[7m \033[0m")
			}

		case CursorUnderline, CursorUnderlineBlink:
			if cursor < len(buffer) {
				// Show cursor as underlined character
				fmt.Printf("\033[4m%c\033[0m", buffer[cursor])
			} else {
				// Cursor at the end - show an underscore
				fmt.Print("\033[4m \033[0m")
			}
		}
	} else {
		// Cursor is hidden (for blinking effect) - just print the character without styling
		if cursor < len(buffer) {
			fmt.Printf("%c", buffer[cursor])
		} else {
			fmt.Print(" ")
		}
	}

	// Print remaining text after cursor
	if cursor+1 < len(buffer) {
		fmt.Print(string(buffer[cursor+1:]))
	}
}

// SetCursorStyle sets the default cursor style for all inputs
func SetCursorStyle(style CursorStyle) {
	defaultCursorStyle = style
}

// cursorStyle field for ColorInput
type editConfig struct {
	cursorStyle CursorStyle
}

var inputEditConfigs = make(map[*ColorInput]*editConfig)
var configMutex sync.Mutex

// SetCursorStyle sets the cursor style for this specific input
func (i *ColorInput) SetCursorStyle(style CursorStyle) *ColorInput {
	configMutex.Lock()
	defer configMutex.Unlock()

	if inputEditConfigs[i] == nil {
		inputEditConfigs[i] = &editConfig{}
	}
	inputEditConfigs[i].cursorStyle = style
	return i
}

// getCursorStyle returns the cursor style for this input
func (i *ColorInput) getCursorStyle() CursorStyle {
	configMutex.Lock()
	defer configMutex.Unlock()

	if cfg := inputEditConfigs[i]; cfg != nil {
		return cfg.cursorStyle
	}
	return defaultCursorStyle
}

// RunWithEdit displays the prompt with full line editing support
// Allows using arrow keys, Home, End, Delete, and various Ctrl shortcuts
func (i *ColorInput) RunWithEdit() (string, error) {
	cursorStyle := i.getCursorStyle()

	for {
		// Build the prompt string
		var promptStr string
		if i.defaultValue != "" {
			promptStr = fmt.Sprintf("%s%s %s%s %s[%s]%s: %s",
				i.messageColor, i.promptSymbol, i.message, ColorReset,
				i.defaultColor, i.defaultValue, ColorReset,
				i.inputColor)
		} else {
			promptStr = fmt.Sprintf("%s%s %s%s: %s",
				i.messageColor, i.promptSymbol, i.message, ColorReset,
				i.inputColor)
		}

		// Use the line editor
		input, err := editLine(promptStr, i.defaultValue, cursorStyle)
		fmt.Print(ColorReset) // Reset color after input

		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		// Clean the input
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
		// BEGIN: human-fixed
		fmt.Print(carriageReturn)
		fmt.Print(ANSIReset)
		// END: human-fixed
		return input, nil
	}
}
