package prompt

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
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
	restoreCmd, err := enableRawMode()
	if err != nil {
		return "", err
	}
	defer disableRawMode(restoreCmd)

	fmt.Print(prompt)
	fmt.Print(saveCursor)
	fmt.Print(hideCursor)
	defer fmt.Print(showCursor)

	buffer := []rune(defaultValue)
	cursor := len(buffer)

	var cursorVisible bool = true
	var cursorMutex sync.Mutex

	stopBlink := startBlinker(cursorStyle, &buffer, &cursor, &cursorVisible, &cursorMutex)
	defer func() {
		if stopBlink != nil {
			close(stopBlink)
		}
	}()

	renderLine(buffer, cursor, cursorVisible, cursorStyle)

	inputBuf := make([]byte, 1)
	escapeSeq := make([]byte, 0, 10)
	inEscape := false

	for {
		n, err := os.Stdin.Read(inputBuf)
		if err != nil || n == 0 {
			continue
		}

		b := inputBuf[0]
		cursorMutex.Lock()
		cursorVisible = true

		escResult := processEscapeInput(b, &inEscape, &escapeSeq, buffer, cursor)
		if escResult.handled {
			buffer, cursor = escResult.buffer, escResult.cursor
			renderLine(buffer, cursor, cursorVisible, cursorStyle)
			cursorMutex.Unlock()
			continue
		}

		result := processControlKey(b, buffer, cursor)
		buffer, cursor = result.buffer, result.cursor

		if result.ctrlC {
			cursorMutex.Unlock()
			fmt.Println()
			disableRawMode(restoreCmd)
			os.Exit(0)
		}

		if result.done {
			cursorMutex.Unlock()
			fmt.Println()
			return result.output, result.err
		}

		renderLine(buffer, cursor, cursorVisible, cursorStyle)
		cursorMutex.Unlock()
	}
}

// renderLine renders the current line with the cursor at the correct position
func renderLine(buffer []rune, cursor int, cursorVisible bool, cursorStyle CursorStyle) {
	fmt.Print(restoreCursor + clearToEnd)

	if cursor > 0 {
		fmt.Print(string(buffer[:cursor]))
	}

	if cursorVisible || (cursorStyle != CursorBlockBlink && cursorStyle != CursorUnderlineBlink) {
		renderCursorChar(buffer, cursor, cursorStyle)
	} else {
		renderHiddenCursorChar(buffer, cursor)
	}

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
			return "", fmt.Errorf(errReadInput, err)
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
