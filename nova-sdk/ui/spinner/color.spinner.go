package spinner

import (
	"fmt"
	"sync"
	"time"
)

// ColorSpinner represents a loading animation with color support
type ColorSpinner struct {
	frames      []string
	delay       time.Duration
	prefix      string
	suffix      string
	prefixColor string
	frameColor  string
	suffixColor string
	stop        chan bool
	done        chan bool
	state       SpinnerState
	mu          sync.RWMutex // Protects all fields for concurrent access
}

// NewWithColor creates a new ColorSpinner with a prefix
func NewWithColor(prefix string) *ColorSpinner {
	return &ColorSpinner{
		frames:      FramesBraille,
		delay:       100 * time.Millisecond,
		prefix:      prefix,
		suffix:      "",
		prefixColor: ColorWhite,
		frameColor:  ColorCyan,
		suffixColor: ColorWhite,
		stop:        make(chan bool),
		done:        make(chan bool),
		state:       StateIdle,
	}
}

// SetFrames allows customization of animation characters
func (s *ColorSpinner) SetFrames(frames []string) *ColorSpinner {
	s.mu.Lock()
	s.frames = frames
	s.mu.Unlock()
	return s
}

// SetDelay allows customization of animation speed
func (s *ColorSpinner) SetDelay(delay time.Duration) *ColorSpinner {
	s.mu.Lock()
	s.delay = delay
	s.mu.Unlock()
	return s
}

// SetSuffix adds a suffix after the animation
// Can be called before or during spinner execution
func (s *ColorSpinner) SetSuffix(suffix string) *ColorSpinner {
	s.mu.Lock()
	s.suffix = suffix
	s.mu.Unlock()
	return s
}

// UpdateSuffix updates the suffix during execution
// Alias for SetSuffix for more clarity during dynamic updates
func (s *ColorSpinner) UpdateSuffix(suffix string) {
	s.SetSuffix(suffix)
}

// SetPrefix updates the prefix
// Can be called before or during spinner execution
func (s *ColorSpinner) SetPrefix(prefix string) *ColorSpinner {
	s.mu.Lock()
	s.prefix = prefix
	s.mu.Unlock()
	return s
}

// UpdatePrefix updates the prefix during execution
// Alias for SetPrefix for more clarity during dynamic updates
func (s *ColorSpinner) UpdatePrefix(prefix string) {
	s.SetPrefix(prefix)
}

// SetPrefixColor sets the color of the prefix
func (s *ColorSpinner) SetPrefixColor(color string) *ColorSpinner {
	s.mu.Lock()
	s.prefixColor = color
	s.mu.Unlock()
	return s
}

// SetFrameColor sets the color of the animation frames
func (s *ColorSpinner) SetFrameColor(color string) *ColorSpinner {
	s.mu.Lock()
	s.frameColor = color
	s.mu.Unlock()
	return s
}

// SetSuffixColor sets the color of the suffix
func (s *ColorSpinner) SetSuffixColor(color string) *ColorSpinner {
	s.mu.Lock()
	s.suffixColor = color
	s.mu.Unlock()
	return s
}

// SetColors sets all colors at once (prefix, frame, suffix)
func (s *ColorSpinner) SetColors(prefixColor, frameColor, suffixColor string) *ColorSpinner {
	s.mu.Lock()
	s.prefixColor = prefixColor
	s.frameColor = frameColor
	s.suffixColor = suffixColor
	s.mu.Unlock()
	return s
}

// Start launches the animation in a goroutine
func (s *ColorSpinner) Start() {
	s.mu.Lock()
	s.state = StateRunning
	s.mu.Unlock()

	go func() {
		i := 0
		for {
			select {
			case <-s.stop:
				// Clear the line before exiting
				fmt.Printf("\r%s\r", clearLine())
				s.mu.Lock()
				s.state = StateStopped
				s.mu.Unlock()
				s.done <- true
				return
			default:
				// Read all fields in a thread-safe manner
				s.mu.RLock()
				prefix := s.prefix
				suffix := s.suffix
				prefixColor := s.prefixColor
				frameColor := s.frameColor
				suffixColor := s.suffixColor
				frames := s.frames
				delay := s.delay
				s.mu.RUnlock()

				frame := frames[i%len(frames)]

				if suffix != "" {
					fmt.Printf("\r%s%s%s %s%s%s %s%s%s",
						prefixColor, prefix, ColorReset,
						frameColor, frame, ColorReset,
						suffixColor, suffix, ColorReset)
				} else {
					fmt.Printf("\r%s%s%s %s%s%s",
						prefixColor, prefix, ColorReset,
						frameColor, frame, ColorReset)
				}

				time.Sleep(delay)
				i++
			}
		}
	}()
}

// Stop stops the animation
func (s *ColorSpinner) Stop() {
	s.mu.RLock()
	currentState := s.state
	s.mu.RUnlock()

	// Only stop if currently running
	if currentState != StateRunning {
		return
	}
	s.stop <- true
	<-s.done // Wait for the goroutine to finish
}

// StopWithMessage stops the animation and displays a message
func (s *ColorSpinner) StopWithMessage(message string) {
	s.Stop()
	fmt.Println(message)
}

// Success stops the animation and displays a success message
func (s *ColorSpinner) Success(message string) {
	s.StopWithMessage(fmt.Sprintf("%s✓ %s%s", ColorGreen, message, ColorReset))
}

// Error stops the animation and displays an error message
func (s *ColorSpinner) Error(message string) {
	s.StopWithMessage(fmt.Sprintf("%s✗ %s%s", ColorRed, message, ColorReset))
}

// State returns the current state of the spinner
func (s *ColorSpinner) State() SpinnerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// IsRunning returns true if the spinner is running
func (s *ColorSpinner) IsRunning() bool {
	return s.State() == StateRunning
}

// IsStopped returns true if the spinner is stopped
func (s *ColorSpinner) IsStopped() bool {
	return s.State() == StateStopped
}

// IsIdle returns true if the spinner has not started yet
func (s *ColorSpinner) IsIdle() bool {
	return s.State() == StateIdle
}
