package spinner

import (
	"fmt"
	"sync"
	"time"
)

// SpinnerState represents the state of the spinner
type SpinnerState int

const (
	// StateIdle: spinner created but not yet started
	StateIdle SpinnerState = iota
	// StateRunning: spinner is running
	StateRunning
	// StateStopped: spinner is stopped
	StateStopped
)

// String returns a textual representation of the state
func (s SpinnerState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Spinner represents a loading animation
type Spinner struct {
	frames []string
	delay  time.Duration
	prefix string
	suffix string
	stop   chan bool
	done   chan bool
	state  SpinnerState
	mu     sync.RWMutex // Protects prefix and suffix for concurrent access
}

// Popular predefined frames
var (
	// FramesBraille uses Braille characters for an elegant animation
	FramesBraille = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	// FramesDots uses rotating dots
	FramesDots = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

	// FramesASCII uses classic ASCII characters (compatible everywhere)
	FramesASCII = []string{"|", "/", "-", "\\"}

	// FramesProgressive uses progressive dots
	FramesProgressive = []string{".", "..", "...", "....", "....."}

	// FramesArrows uses rotating arrows
	FramesArrows = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}

	// FramesCircle uses partial circles
	FramesCircle = []string{"◐", "◓", "◑", "◒"}

	// FramesPulsingStar uses a pulsing star animation
	FramesPulsingStar = []string{"✦", "✶", "✷", "✸", "✹", "✸", "✷", "✶"}
)

// New creates a new spinner with a prefix
func New(prefix string) *Spinner {
	return &Spinner{
		frames: FramesBraille,
		delay:  100 * time.Millisecond,
		prefix: prefix,
		suffix: "",
		stop:   make(chan bool),
		done:   make(chan bool),
		state:  StateIdle,
	}
}

// SetFrames allows customization of animation characters
func (s *Spinner) SetFrames(frames []string) *Spinner {
	s.frames = frames
	return s
}

// SetDelay allows customization of animation speed
func (s *Spinner) SetDelay(delay time.Duration) *Spinner {
	s.delay = delay
	return s
}

// SetSuffix adds a suffix after the animation
// Can be called before or during spinner execution
func (s *Spinner) SetSuffix(suffix string) *Spinner {
	s.mu.Lock()
	s.suffix = suffix
	s.mu.Unlock()
	return s
}

// UpdateSuffix updates the suffix during execution
// Alias for SetSuffix for more clarity during dynamic updates
func (s *Spinner) UpdateSuffix(suffix string) {
	s.SetSuffix(suffix)
}

// SetPrefix updates the prefix
// Can be called before or during spinner execution
func (s *Spinner) SetPrefix(prefix string) *Spinner {
	s.mu.Lock()
	s.prefix = prefix
	s.mu.Unlock()
	return s
}

// UpdatePrefix updates the prefix during execution
// Alias for SetPrefix for more clarity during dynamic updates
func (s *Spinner) UpdatePrefix(prefix string) {
	s.SetPrefix(prefix)
}

// Start launches the animation in a goroutine
func (s *Spinner) Start() {
	s.state = StateRunning
	go func() {
		i := 0
		for {
			select {
			case <-s.stop:
				// Clear the line before exiting
				fmt.Printf("\r%s\r", clearLine())
				s.state = StateStopped
				s.done <- true
				return
			default:
				// Read prefix and suffix in a thread-safe manner
				s.mu.RLock()
				prefix := s.prefix
				suffix := s.suffix
				s.mu.RUnlock()

				frame := s.frames[i%len(s.frames)]
				if suffix != "" {
					fmt.Printf("\r%s %s %s", prefix, frame, suffix)
				} else {
					fmt.Printf("\r%s %s", prefix, frame)
				}
				time.Sleep(s.delay)
				i++
			}
		}
	}()
}

// Stop stops the animation
func (s *Spinner) Stop() {
	// Only stop if currently running
	if s.state != StateRunning {
		return
	}
	s.stop <- true
	<-s.done // Wait for the goroutine to finish
}

// StopWithMessage stops the animation and displays a message
func (s *Spinner) StopWithMessage(message string) {
	s.Stop()
	fmt.Println(message)
}

// Success stops the animation and displays a success message
func (s *Spinner) Success(message string) {
	s.StopWithMessage(fmt.Sprintf("✓ %s", message))
}

// Error stops the animation and displays an error message
func (s *Spinner) Error(message string) {
	s.StopWithMessage(fmt.Sprintf("✗ %s", message))
}

// State returns the current state of the spinner
func (s *Spinner) State() SpinnerState {
	return s.state
}

// IsRunning returns true if the spinner is running
func (s *Spinner) IsRunning() bool {
	return s.state == StateRunning
}

// IsStopped returns true if the spinner is stopped
func (s *Spinner) IsStopped() bool {
	return s.state == StateStopped
}

// IsIdle returns true if the spinner has not started yet
func (s *Spinner) IsIdle() bool {
	return s.state == StateIdle
}

// clearLine returns a string of spaces to clear the line
func clearLine() string {
	return "                                                                                "
}
