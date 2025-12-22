# Spinner Package Documentation

## Overview

The `spinner` package provides a simple, thread-safe terminal spinner (loading animation) for Go applications. It allows you to display animated loading indicators with customizable prefixes, suffixes, and animation styles.

## File: spinner.go

### Main Components

#### SpinnerState

`SpinnerState` is an enumeration representing the current state of a spinner.

**States:**
- `StateIdle` - Spinner has been created but not yet started
- `StateRunning` - Spinner is currently running
- `StateStopped` - Spinner has been stopped

**Methods:**
- `String() string` - Returns a textual representation of the state ("idle", "running", "stopped", or "unknown")

#### Spinner

The main `Spinner` struct represents a loading animation with the following features:
- Thread-safe operations using `sync.RWMutex`
- Customizable animation frames
- Adjustable animation speed
- Dynamic prefix and suffix updates
- State management

**Structure:**
```go
type Spinner struct {
    frames []string        // Animation frames to cycle through
    delay  time.Duration   // Delay between frames
    prefix string          // Text displayed before the spinner
    suffix string          // Text displayed after the spinner
    stop   chan bool       // Channel to stop the spinner
    done   chan bool       // Channel to signal completion
    state  SpinnerState    // Current state
    mu     sync.RWMutex    // Mutex for thread-safe access
}
```

### Predefined Frame Sets

The package includes several predefined animation frame sets:

- **FramesBraille** - Uses Braille characters for an elegant animation: `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏`
- **FramesDots** - Rotating dots animation: `⣾ ⣽ ⣻ ⢿ ⡿ ⣟ ⣯ ⣷`
- **FramesASCII** - Classic ASCII characters (universally compatible): `| / - \`
- **FramesProgressive** - Progressive dots: `. .. ... .... .....`
- **FramesArrows** - Rotating arrows: `← ↖ ↑ ↗ → ↘ ↓ ↙`
- **FramesCircle** - Partial circles: `◐ ◓ ◑ ◒`
- **FramesPulsingStar** - Pulsing star animation: `✦ ✶ ✷ ✸ ✹ ✸ ✷ ✶`

### Constructor

#### New

```go
func New(prefix string) *Spinner
```

Creates a new spinner with the specified prefix message. Default settings:
- Frames: `FramesBraille`
- Delay: `100ms`
- State: `StateIdle`

**Parameters:**
- `prefix` - Initial prefix message to display

**Returns:**
- `*Spinner` - A new spinner instance

### Configuration Methods

All configuration methods support method chaining by returning the spinner instance.

#### SetFrames

```go
func (s *Spinner) SetFrames(frames []string) *Spinner
```

Customizes the animation frames.

**Parameters:**
- `frames` - Array of strings representing each frame of the animation

#### SetDelay

```go
func (s *Spinner) SetDelay(delay time.Duration) *Spinner
```

Customizes the animation speed (delay between frames).

**Parameters:**
- `delay` - Duration between frame transitions

#### SetPrefix

```go
func (s *Spinner) SetPrefix(prefix string) *Spinner
```

Sets or updates the prefix text. Can be called before or during spinner execution.

**Parameters:**
- `prefix` - New prefix text

#### UpdatePrefix

```go
func (s *Spinner) UpdatePrefix(prefix string)
```

Alias for `SetPrefix`. Provides more clarity when updating the prefix during execution.

**Parameters:**
- `prefix` - New prefix text

#### SetSuffix

```go
func (s *Spinner) SetSuffix(suffix string) *Spinner
```

Sets or updates the suffix text (displayed after the spinner). Can be called before or during spinner execution.

**Parameters:**
- `suffix` - New suffix text

#### UpdateSuffix

```go
func (s *Spinner) UpdateSuffix(suffix string)
```

Alias for `SetSuffix`. Provides more clarity when updating the suffix during execution.

**Parameters:**
- `suffix` - New suffix text

### Control Methods

#### Start

```go
func (s *Spinner) Start()
```

Launches the spinner animation in a background goroutine. The animation continues until `Stop()` is called.

#### Stop

```go
func (s *Spinner) Stop()
```

Stops the spinner animation and clears the spinner line. Only has effect if the spinner is currently running. This method blocks until the spinner goroutine has fully stopped.

#### StopWithMessage

```go
func (s *Spinner) StopWithMessage(message string)
```

Stops the spinner and displays a message on a new line.

**Parameters:**
- `message` - Message to display after stopping

#### Success

```go
func (s *Spinner) Success(message string)
```

Stops the spinner and displays a success message with a checkmark (`✓`).

**Parameters:**
- `message` - Success message to display

#### Error

```go
func (s *Spinner) Error(message string)
```

Stops the spinner and displays an error message with a cross mark (`✗`).

**Parameters:**
- `message` - Error message to display

### State Query Methods

#### State

```go
func (s *Spinner) State() SpinnerState
```

Returns the current state of the spinner.

**Returns:**
- `SpinnerState` - Current state (Idle, Running, or Stopped)

#### IsRunning

```go
func (s *Spinner) IsRunning() bool
```

Checks if the spinner is currently running.

**Returns:**
- `bool` - `true` if running, `false` otherwise

#### IsStopped

```go
func (s *Spinner) IsStopped() bool
```

Checks if the spinner has been stopped.

**Returns:**
- `bool` - `true` if stopped, `false` otherwise

#### IsIdle

```go
func (s *Spinner) IsIdle() bool
```

Checks if the spinner has not been started yet.

**Returns:**
- `bool` - `true` if idle, `false` otherwise

## Usage Example

```go
package main

import (
    "time"
    "your-module/spinner"
)

func main() {
    // Create and configure a spinner
    sp := spinner.New("Loading").
        SetFrames(spinner.FramesDots).
        SetDelay(80 * time.Millisecond)

    // Start the animation
    sp.Start()

    // Simulate work
    time.Sleep(2 * time.Second)

    // Update the message dynamically
    sp.UpdateSuffix("processing data...")
    time.Sleep(2 * time.Second)

    // Stop with success
    sp.Success("Operation completed!")
}
```

## Thread Safety

The spinner is designed to be thread-safe:
- The `prefix` and `suffix` fields are protected by a read-write mutex
- Methods can be safely called from different goroutines
- State transitions are properly synchronized

## Notes

- The spinner clears the current line when stopped
- Only one spinner animation should be displayed per terminal line
- The animation runs in a background goroutine
- Stopping a non-running spinner is a no-op (safe to call)
