# Spinner Package Documentation

Welcome to the Nova SDK Spinner package documentation. This package provides elegant, thread-safe terminal spinner animations for Go applications.

## Overview

The spinner package offers two main implementations:

1. **Basic Spinner** - Simple, lightweight terminal spinner
2. **Color Spinner** - Enhanced spinner with full ANSI color support

Both implementations are thread-safe, customizable, and easy to use with a fluent API design.

## Documentation Files

- **[spinner.md](spinner.md)** - Documentation for the basic `Spinner` type
  - Simple terminal spinner without color support
  - Lightweight and universally compatible
  - Predefined animation frame sets
  - Thread-safe operations
  - State management

- **[color.spinner.md](color.spinner.md)** - Documentation for the `ColorSpinner` type
  - Full ANSI color support
  - Independent color control for prefix, frames, and suffix
  - Extensive color constants (standard, bright, backgrounds)
  - Text modifiers (bold, dim, italic, etc.)
  - All features of basic spinner plus colors

## Quick Start

### Basic Spinner

```go
import "your-module/spinner"

sp := spinner.New("Loading")
sp.Start()

// Do work...

sp.Success("Done!")
```

### Color Spinner

```go
import "your-module/spinner"

sp := spinner.NewWithColor("Processing").
    SetColors(
        spinner.ColorBrightYellow,  // Prefix
        spinner.ColorBrightCyan,    // Frame
        spinner.ColorBrightGreen,   // Suffix
    )
sp.Start()

// Do work...

sp.Success("Complete!")
```

## Key Features

### Animation Styles

Choose from several predefined frame sets:
- **Braille** - Elegant Braille characters (default)
- **Dots** - Rotating dots
- **ASCII** - Classic ASCII characters (|/-\)
- **Progressive** - Progressive dots (. .. ...)
- **Arrows** - Rotating arrows (← ↖ ↑ ↗ →)
- **Circle** - Partial circles (◐ ◓ ◑ ◒)
- **Pulsing Star** - Pulsing star animation

### Customization

- **Custom Frames** - Define your own animation frames
- **Adjustable Speed** - Set delay between frames
- **Dynamic Text** - Update prefix/suffix during animation
- **Color Control** - (ColorSpinner) Individual colors for each component
- **Method Chaining** - Fluent API for easy configuration

### Thread Safety

Both spinner implementations are fully thread-safe:
- Protected by read-write mutexes
- Safe concurrent access to all methods
- Can update text/colors during animation
- Proper state synchronization

### State Management

Track spinner state with built-in helpers:
- `IsIdle()` - Not yet started
- `IsRunning()` - Currently animating
- `IsStopped()` - Animation stopped

### Completion Methods

Clean ways to end the animation:
- `Stop()` - Stop and clear line
- `StopWithMessage(msg)` - Stop with custom message
- `Success(msg)` - Stop with success indicator (✓)
- `Error(msg)` - Stop with error indicator (✗)

## When to Use Which?

### Use Basic Spinner When:
- Maximum terminal compatibility is required
- No color output is needed
- Minimal dependencies are preferred
- Running in environments without ANSI support

### Use Color Spinner When:
- Rich visual feedback is desired
- Color differentiation aids user experience
- Running in modern terminal environments
- Want to match brand/theme colors

## Examples

### Progress Tracking

```go
sp := spinner.New("Downloading")
sp.Start()

for i := 0; i <= 100; i += 10 {
    sp.UpdateSuffix(fmt.Sprintf("%d%%", i))
    time.Sleep(500 * time.Millisecond)
}

sp.Success("Download complete!")
```

### Multi-Stage Process

```go
sp := spinner.NewWithColor("Initializing").
    SetFrameColor(spinner.ColorBrightCyan)

sp.Start()
time.Sleep(1 * time.Second)

sp.UpdatePrefix("Processing")
sp.UpdateSuffix("stage 1/3")
time.Sleep(2 * time.Second)

sp.UpdateSuffix("stage 2/3")
time.Sleep(2 * time.Second)

sp.UpdateSuffix("stage 3/3")
time.Sleep(2 * time.Second)

sp.Success("All stages complete!")
```

### Error Handling

```go
sp := spinner.New("Connecting to server")
sp.Start()

err := connectToServer()
if err != nil {
    sp.Error(fmt.Sprintf("Connection failed: %v", err))
    return
}

sp.Success("Connected successfully!")
```

## API Reference

For detailed API documentation, please refer to:
- [Basic Spinner API](spinner.md)
- [Color Spinner API](color.spinner.md)

## Package Structure

```
spinner/
├── spinner.go           # Basic spinner implementation
├── color.spinner.go     # Color spinner implementation
└── doc/
    ├── README.md        # This file
    ├── spinner.md       # Basic spinner documentation
    └── color.spinner.md # Color spinner documentation
```

## License

Part of the Nova SDK project.
