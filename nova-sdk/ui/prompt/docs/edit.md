# edit.go - Interactive Line Editing

This file provides advanced line editing capabilities with arrow key support and customizable cursors.

## Overview

The `edit.go` file implements a raw terminal mode editor that allows users to edit input using arrow keys, navigate with Home/End, and use various keyboard shortcuts. It also provides customizable cursor styles including blinking cursors.

---

## Cursor Styles

### CursorStyle

An enum type defining cursor appearance.

```go
type CursorStyle int

const (
    CursorBlock          CursorStyle = iota  // Solid block cursor (default)
    CursorBlockBlink                         // Blinking block cursor
    CursorUnderline                          // Underline cursor
    CursorUnderlineBlink                     // Blinking underline cursor
)
```

### Visual Representation

- **CursorBlock**: Character with inverted colors `█`
- **CursorBlockBlink**: Same as block but blinks every 500ms
- **CursorUnderline**: Underlined character `_`
- **CursorUnderlineBlink**: Same as underline but blinks every 500ms

---

## Functions

### SetCursorStyle(style CursorStyle)

Sets the default cursor style for all inputs globally.

```go
prompt.SetCursorStyle(prompt.CursorBlockBlink)
```

This affects all subsequent calls to `RunWithEdit()` unless overridden per-input.

---

## ColorInput Methods

### SetCursorStyle(style CursorStyle) *ColorInput

Sets the cursor style for a specific input, overriding the global default.

```go
input := prompt.NewWithColor("Enter text").
    SetCursorStyle(prompt.CursorBlockBlink)
```

**Example with all cursor styles:**

```go
// Solid block (default)
input1 := prompt.NewWithColor("Name").
    SetCursorStyle(prompt.CursorBlock)

// Blinking block
input2 := prompt.NewWithColor("Email").
    SetCursorStyle(prompt.CursorBlockBlink)

// Solid underline
input3 := prompt.NewWithColor("Phone").
    SetCursorStyle(prompt.CursorUnderline)

// Blinking underline
input4 := prompt.NewWithColor("Address").
    SetCursorStyle(prompt.CursorUnderlineBlink)
```

### RunWithEdit() (string, error)

Displays the prompt with full line editing support.

```go
input := prompt.NewWithColor("Edit this text").
    SetDefault("Initial value")
result, err := input.RunWithEdit()
```

**Features:**
- Arrow key navigation (left/right)
- Home/End keys for quick navigation
- Character insertion at cursor position
- Backspace and Delete support
- Multiple Ctrl shortcuts
- Visual cursor with customizable style
- Support for default values that can be edited
- Validation support

---

## Keyboard Shortcuts

### Navigation

| Key | Action | Alternative |
|-----|--------|-------------|
| `←` | Move cursor left | `Ctrl+B` |
| `→` | Move cursor right | `Ctrl+F` |
| `Home` | Move to beginning | `Ctrl+A` |
| `End` | Move to end | `Ctrl+E` |

### Editing

| Key | Action |
|-----|--------|
| `Backspace` | Delete character before cursor |
| `Delete` | Delete character at cursor |
| `Ctrl+K` | Kill (delete) from cursor to end of line |
| `Ctrl+U` | Kill (delete) from beginning to cursor |
| Any printable character | Insert at cursor position |

### Control

| Key | Action |
|-----|--------|
| `Enter` | Submit input |
| `Ctrl+C` | Cancel and exit |
| `Ctrl+D` | EOF (if line is empty) |

---

## Usage Examples

### Basic Interactive Input

```go
input := prompt.NewWithColor("🤖 Ask me something?")
question, err := input.RunWithEdit()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("You asked: %s\n", question)
```

### Edit Default Value

```go
input := prompt.NewWithColor("📝 Edit your name").
    SetDefault("John Doe")
name, err := input.RunWithEdit()
// User can use arrows to edit "John Doe" before submitting
```

### With Validation

```go
input := prompt.NewWithColor("Enter email").
    SetDefault("user@example.com").
    SetValidator(func(s string) error {
        if !strings.Contains(s, "@") {
            return fmt.Errorf("invalid email format")
        }
        return nil
    })

email, err := input.RunWithEdit()
// User can edit with arrows, validation runs on submit
```

### Blinking Cursor

```go
input := prompt.NewWithColor("⚡ Type here").
    SetCursorStyle(prompt.CursorBlockBlink).
    SetDefault("Watch the cursor blink!")

result, err := input.RunWithEdit()
```

### Underline Cursor

```go
input := prompt.NewWithColor("📏 Enter text").
    SetCursorStyle(prompt.CursorUnderline)

result, err := input.RunWithEdit()
```

### Global Cursor Style

```go
// Set globally for all inputs
prompt.SetCursorStyle(prompt.CursorBlockBlink)

// All inputs will use blinking cursor
input1 := prompt.NewWithColor("First name")
name1, _ := input1.RunWithEdit()

input2 := prompt.NewWithColor("Last name")
name2, _ := input2.RunWithEdit()

// Override for specific input
input3 := prompt.NewWithColor("Email").
    SetCursorStyle(prompt.CursorUnderline)
email, _ := input3.RunWithEdit()
```

### Colored Input with Editing

```go
input := prompt.NewWithColor("🎨 Favorite color?").
    SetMessageColor(prompt.ColorBrightMagenta).
    SetInputColor(prompt.ColorBrightCyan).
    SetDefault("blue").
    SetCursorStyle(prompt.CursorBlockBlink)

color, err := input.RunWithEdit()
```

---

## Advanced Features

### Cursor Blinking Mechanism

Blinking cursors use a goroutine with a 500ms ticker:
- Cursor alternates between visible and hidden states
- Visibility resets on any keypress
- Goroutine is properly cleaned up on exit
- Thread-safe using mutex

### Terminal Raw Mode

The editor switches the terminal to raw mode:
- Disables line buffering
- Disables echo
- Captures individual keystrokes
- Properly restores terminal on exit
- Handles Ctrl+C gracefully

### Escape Sequence Handling

Supports complex terminal escape sequences:
- Arrow keys: `ESC[A` (up), `ESC[B` (down), `ESC[C` (right), `ESC[D` (left)
- Function keys: Home (`ESC[H`), End (`ESC[F`)
- Delete key: `ESC[3~`

---

## Platform Compatibility

### Supported Platforms

- ✅ **Linux**: Full support
- ✅ **macOS**: Full support
- ⚠️ **Windows**: Limited support (raw mode not available)

### Windows Note

On Windows, the raw terminal mode is not supported. Use the basic `Run()` method instead of `RunWithEdit()`:

```go
input := prompt.NewWithColor("Enter name")

// On Windows, use Run() instead
name, err := input.Run()
```

---

## Comparison: Run() vs RunWithEdit()

| Feature | Run() | RunWithEdit() |
|---------|-------|---------------|
| Arrow key navigation | ❌ | ✅ |
| Home/End keys | ❌ | ✅ |
| Insert at cursor | ❌ | ✅ |
| Edit default values | ❌ | ✅ |
| Ctrl shortcuts | ❌ | ✅ |
| Custom cursor styles | ❌ | ✅ |
| Blinking cursor | ❌ | ✅ |
| Windows support | ✅ | ⚠️ Limited |
| Simpler implementation | ✅ | ❌ |
| Raw mode required | ❌ | ✅ |

---

## Best Practices

1. **Use RunWithEdit() for better UX**: Provides a much better editing experience
2. **Provide meaningful defaults**: Users can easily edit them with arrows
3. **Choose appropriate cursor style**: Blinking cursors draw attention but can be distracting
4. **Test on target platforms**: Ensure raw mode works on your deployment environment
5. **Fallback for Windows**: Provide alternative input method for Windows users
6. **Keep validation simple**: Complex validation works better with separate confirmation

---

## Implementation Details

### Internal Functions

#### enableRawMode() (*exec.Cmd, error)

Enables raw terminal mode using `stty` command:
- Saves current terminal state
- Enables raw mode
- Disables echo
- Returns restore command

#### disableRawMode(restoreCmd *exec.Cmd)

Restores terminal to previous state.

#### editLine(prompt, defaultValue string, cursorStyle CursorStyle) (string, error)

Core editing loop:
- Handles all keyboard input
- Manages cursor position
- Renders line with cursor
- Returns final input

#### renderLine(buffer []rune, cursor int, cursorVisible bool, cursorStyle CursorStyle)

Renders the current line:
- Clears previous line
- Displays text before cursor
- Displays cursor with appropriate style
- Displays text after cursor

---

## Error Handling

```go
input := prompt.NewWithColor("Enter value")
result, err := input.RunWithEdit()
if err != nil {
    if err.Error() == "EOF" {
        fmt.Println("User pressed Ctrl+D")
    } else if err.Error() == "raw mode not supported on Windows" {
        // Fallback to Run()
        result, err = input.Run()
    } else {
        log.Fatal(err)
    }
}
```

---

## Performance Notes

- Blinking cursors add minimal overhead (~goroutine + timer)
- Raw mode is more efficient than buffered mode for interactive input
- No significant performance difference between cursor styles
- Terminal rendering is optimized with escape sequences

---

## Security Considerations

- Input is not echoed in password mode (not implemented yet)
- Terminal state is always restored, even on panic
- Ctrl+C properly exits and cleans up
- No buffer overflow issues (Go slices handle this)
