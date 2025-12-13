# Color Spinner Package Documentation

## Overview

The `color.spinner.go` file extends the basic spinner functionality with comprehensive color support using ANSI escape codes. It provides a colored version of the terminal spinner with independent color controls for prefix, frame, and suffix.

## File: color.spinner.go

### Color Constants

The package provides extensive ANSI color code constants for terminal output customization.

#### Reset and Modifiers

- `ColorReset` - Reset all styles and colors to default
- `ColorBold` - Bold text
- `ColorDim` - Dimmed text
- `ColorItalic` - Italic text
- `ColorUnderline` - Underlined text
- `ColorBlink` - Blinking text
- `ColorReverse` - Reversed foreground/background colors
- `ColorHidden` - Hidden text

#### Standard Foreground Colors

- `ColorBlack`, `ColorRed`, `ColorGreen`, `ColorYellow`
- `ColorBlue`, `ColorMagenta`, `ColorCyan`, `ColorWhite`
- `ColorGray`
- `ColorPurple` - Alias for Magenta

#### Bright Foreground Colors

- `ColorBrightBlack`, `ColorBrightRed`, `ColorBrightGreen`, `ColorBrightYellow`
- `ColorBrightBlue`, `ColorBrightMagenta`, `ColorBrightCyan`, `ColorBrightWhite`
- `ColorBrightPurple` - Alias for Bright Magenta

#### Background Colors

Standard backgrounds:
- `BgBlack`, `BgRed`, `BgGreen`, `BgYellow`
- `BgBlue`, `BgMagenta`, `BgCyan`, `BgWhite`

Bright backgrounds:
- `BgBrightBlack`, `BgBrightRed`, `BgBrightGreen`, `BgBrightYellow`
- `BgBrightBlue`, `BgBrightMagenta`, `BgBrightCyan`, `BgBrightWhite`

### ColorSpinner

The `ColorSpinner` struct extends the basic spinner with color support for each component.

**Structure:**
```go
type ColorSpinner struct {
    frames      []string        // Animation frames
    delay       time.Duration   // Delay between frames
    prefix      string          // Text before spinner
    suffix      string          // Text after spinner
    prefixColor string          // ANSI color for prefix
    frameColor  string          // ANSI color for spinner frames
    suffixColor string          // ANSI color for suffix
    stop        chan bool       // Stop signal channel
    done        chan bool       // Completion signal channel
    state       SpinnerState    // Current state
    mu          sync.RWMutex    // Mutex for thread safety
}
```

### Constructor

#### NewWithColor

```go
func NewWithColor(prefix string) *ColorSpinner
```

Creates a new ColorSpinner with default color settings.

**Default Settings:**
- Frames: `FramesBraille`
- Delay: `100ms`
- Prefix Color: `ColorWhite`
- Frame Color: `ColorCyan`
- Suffix Color: `ColorWhite`
- State: `StateIdle`

**Parameters:**
- `prefix` - Initial prefix message

**Returns:**
- `*ColorSpinner` - A new colored spinner instance

### Configuration Methods

All configuration methods support method chaining by returning the spinner instance.

#### SetFrames

```go
func (s *ColorSpinner) SetFrames(frames []string) *ColorSpinner
```

Sets custom animation frames. Thread-safe operation.

**Parameters:**
- `frames` - Array of strings for animation frames

#### SetDelay

```go
func (s *ColorSpinner) SetDelay(delay time.Duration) *ColorSpinner
```

Sets the animation speed. Thread-safe operation.

**Parameters:**
- `delay` - Duration between frame transitions

#### SetPrefix

```go
func (s *ColorSpinner) SetPrefix(prefix string) *ColorSpinner
```

Sets or updates the prefix text. Can be called before or during execution. Thread-safe operation.

**Parameters:**
- `prefix` - New prefix text

#### UpdatePrefix

```go
func (s *ColorSpinner) UpdatePrefix(prefix string)
```

Alias for `SetPrefix` with clearer semantics for dynamic updates during execution.

**Parameters:**
- `prefix` - New prefix text

#### SetSuffix

```go
func (s *ColorSpinner) SetSuffix(suffix string) *ColorSpinner
```

Sets or updates the suffix text. Can be called before or during execution. Thread-safe operation.

**Parameters:**
- `suffix` - New suffix text

#### UpdateSuffix

```go
func (s *ColorSpinner) UpdateSuffix(suffix string)
```

Alias for `SetSuffix` with clearer semantics for dynamic updates during execution.

**Parameters:**
- `suffix` - New suffix text

### Color Configuration Methods

#### SetPrefixColor

```go
func (s *ColorSpinner) SetPrefixColor(color string) *ColorSpinner
```

Sets the color of the prefix text. Thread-safe operation.

**Parameters:**
- `color` - ANSI color code (use package color constants)

#### SetFrameColor

```go
func (s *ColorSpinner) SetFrameColor(color string) *ColorSpinner
```

Sets the color of the animation frames. Thread-safe operation.

**Parameters:**
- `color` - ANSI color code (use package color constants)

#### SetSuffixColor

```go
func (s *ColorSpinner) SetSuffixColor(color string) *ColorSpinner
```

Sets the color of the suffix text. Thread-safe operation.

**Parameters:**
- `color` - ANSI color code (use package color constants)

#### SetColors

```go
func (s *ColorSpinner) SetColors(prefixColor, frameColor, suffixColor string) *ColorSpinner
```

Sets all colors at once in a single thread-safe operation.

**Parameters:**
- `prefixColor` - ANSI color for prefix
- `frameColor` - ANSI color for frames
- `suffixColor` - ANSI color for suffix

### Control Methods

#### Start

```go
func (s *ColorSpinner) Start()
```

Launches the colored spinner animation in a background goroutine. The animation displays colored output and continues until `Stop()` is called.

**Behavior:**
- Sets state to `StateRunning`
- Spawns goroutine for animation loop
- Reads all configuration in thread-safe manner
- Automatically resets colors after each component

#### Stop

```go
func (s *ColorSpinner) Stop()
```

Stops the spinner animation and clears the line. Only has effect if currently running. Blocks until the animation goroutine has fully stopped.

**Behavior:**
- Checks current state before stopping
- Sends stop signal
- Waits for goroutine completion
- Sets state to `StateStopped`
- Clears the terminal line

#### StopWithMessage

```go
func (s *ColorSpinner) StopWithMessage(message string)
```

Stops the spinner and displays a message on a new line.

**Parameters:**
- `message` - Message to display (can include ANSI color codes)

#### Success

```go
func (s *ColorSpinner) Success(message string)
```

Stops the spinner and displays a success message in green with a checkmark.

**Parameters:**
- `message` - Success message to display

**Output Format:**
`✓ [message]` in green

#### Error

```go
func (s *ColorSpinner) Error(message string)
```

Stops the spinner and displays an error message in red with a cross mark.

**Parameters:**
- `message` - Error message to display

**Output Format:**
`✗ [message]` in red

### State Query Methods

#### State

```go
func (s *ColorSpinner) State() SpinnerState
```

Returns the current state of the spinner. Thread-safe operation.

**Returns:**
- `SpinnerState` - Current state (Idle, Running, or Stopped)

#### IsRunning

```go
func (s *ColorSpinner) IsRunning() bool
```

Checks if the spinner is currently running.

**Returns:**
- `bool` - `true` if running, `false` otherwise

#### IsStopped

```go
func (s *ColorSpinner) IsStopped() bool
```

Checks if the spinner has been stopped.

**Returns:**
- `bool` - `true` if stopped, `false` otherwise

#### IsIdle

```go
func (s *ColorSpinner) IsIdle() bool
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
    // Create a colorful spinner
    sp := spinner.NewWithColor("Downloading").
        SetFrames(spinner.FramesDots).
        SetDelay(80 * time.Millisecond).
        SetColors(
            spinner.ColorBrightYellow,  // Prefix color
            spinner.ColorBrightCyan,    // Frame color
            spinner.ColorBrightGreen,   // Suffix color
        )

    // Start the animation
    sp.Start()

    // Simulate work
    time.Sleep(2 * time.Second)

    // Update suffix with progress
    sp.UpdateSuffix("50% complete")
    time.Sleep(2 * time.Second)

    // Complete with success
    sp.Success("Download completed!")

    // Or handle an error
    // sp.Error("Download failed!")
}
```

## Advanced Example: Custom Colors

```go
// Create a spinner with custom color scheme
sp := spinner.NewWithColor("Processing").
    SetPrefixColor(spinner.ColorBold + spinner.ColorBrightMagenta).
    SetFrameColor(spinner.ColorBrightCyan).
    SetSuffixColor(spinner.ColorDim + spinner.ColorWhite)

sp.Start()
sp.UpdateSuffix("analyzing data...")
time.Sleep(3 * time.Second)
sp.Success("Analysis complete!")
```

## Thread Safety

The ColorSpinner is fully thread-safe:
- All fields are protected by `sync.RWMutex`
- Configuration can be updated during animation
- State transitions are properly synchronized
- Safe to call from multiple goroutines

## Terminal Compatibility

The ANSI color codes work in most modern terminals:
- Linux/Unix terminals
- macOS Terminal and iTerm2
- Windows Terminal and PowerShell 7+
- VSCode integrated terminal
- Most CI/CD environments

For maximum compatibility in environments without color support, consider using the basic `Spinner` instead.

## Notes

- Colors are automatically reset after each component to prevent bleeding
- All configuration methods are chainable
- Dynamic updates to text and colors are thread-safe
- The spinner automatically clears the line when stopped
- Success and error messages use predefined colors (green/red) regardless of custom settings
