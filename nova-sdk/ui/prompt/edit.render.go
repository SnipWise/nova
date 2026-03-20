package prompt

import (
	"fmt"
	"sync"
	"time"
)

// startBlinker launches the cursor-blink goroutine for blinking cursor styles.
// It returns the stop channel that the caller must close to stop the goroutine.
// Returns nil when cursorStyle is non-blinking (no goroutine is started).
func startBlinker(
	cursorStyle CursorStyle,
	buffer *[]rune,
	cursor *int,
	cursorVisible *bool,
	mu *sync.Mutex,
) chan bool {
	if cursorStyle != CursorBlockBlink && cursorStyle != CursorUnderlineBlink {
		return nil
	}
	stopCh := make(chan bool)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				*cursorVisible = !*cursorVisible
				renderLine(*buffer, *cursor, *cursorVisible, cursorStyle)
				mu.Unlock()
			case <-stopCh:
				return
			}
		}
	}()
	return stopCh
}

// renderCursorChar prints the cursor character with its visual style applied.
// Called when the cursor is visible (non-blinking styles, or blinking styles in the "on" phase).
func renderCursorChar(buffer []rune, cursor int, cursorStyle CursorStyle) {
	switch cursorStyle {
	case CursorBlock, CursorBlockBlink:
		if cursor < len(buffer) {
			fmt.Printf("\033[7m%c\033[0m", buffer[cursor])
		} else {
			fmt.Print("\033[7m \033[0m")
		}
	case CursorUnderline, CursorUnderlineBlink:
		if cursor < len(buffer) {
			fmt.Printf("\033[4m%c\033[0m", buffer[cursor])
		} else {
			fmt.Print("\033[4m \033[0m")
		}
	}
}

// renderHiddenCursorChar prints the character at the cursor position without any styling.
// Called during the "off" phase of blinking cursor styles.
func renderHiddenCursorChar(buffer []rune, cursor int) {
	if cursor < len(buffer) {
		fmt.Printf("%c", buffer[cursor])
	} else {
		fmt.Print(" ")
	}
}
