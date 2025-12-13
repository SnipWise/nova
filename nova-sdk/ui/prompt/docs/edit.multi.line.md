# edit.multi.line.go - Multi-Line Text Editor

This file provides a complete multi-line text editor with full cursor navigation, line editing, and advanced text manipulation capabilities.

## Overview

The `edit.multi.line.go` file implements a full-featured multi-line text editor that runs in raw terminal mode. It allows users to create and edit text across multiple lines with complete cursor control, insert/delete operations, and keyboard navigation similar to traditional text editors.

---

## Types

### MultiLineEditor

A struct representing the state of a multi-line text editor.

```go
type MultiLineEditor struct {
    lines         [][]rune     // Buffer as lines of runes
    cursorLine    int          // Current line number (0-based)
    cursorCol     int          // Current column in the line (0-based)
    cursorVisible bool         // Cursor visibility for blinking
    cursorStyle   CursorStyle  // Cursor style (block, underline, etc.)
    cursorMutex   sync.Mutex   // Thread-safe cursor operations
    stopBlink     chan bool    // Channel to stop blinking goroutine
    needsRender   chan bool    // Signal that a render is needed
}
```

**Fields:**
- `lines` - Text buffer stored as an array of lines, each line is an array of runes
- `cursorLine` - Current line position (0 = first line)
- `cursorCol` - Current column position within the line (0 = start of line)
- `cursorVisible` - Controls cursor visibility during blinking animation
- `cursorStyle` - The visual style of the cursor (inherited from single-line editor)
- `cursorMutex` - Ensures thread-safe access to cursor state
- `stopBlink` - Channel used to stop the blinking goroutine
- `needsRender` - Buffered channel that signals when a render is needed for blinking

---

## Constructor

### NewMultiLineEditor(defaultValue string, cursorStyle CursorStyle) *MultiLineEditor

Creates a new multi-line editor with optional default text.

```go
editor := prompt.NewMultiLineEditor("Initial text\nSecond line", prompt.CursorBlock)
```

**Parameters:**
- `defaultValue` - Initial text to populate the editor (can contain `\n` for multiple lines)
- `cursorStyle` - The cursor style to use (CursorBlock, CursorBlockBlink, etc.)

**Behavior:**
- If `defaultValue` is empty, starts with one empty line
- If `defaultValue` contains text, splits it on newlines and positions cursor at the end
- Cursor is placed at the end of the last line by default

**Example:**
```go
// Start with empty editor
editor := prompt.NewMultiLineEditor("", prompt.CursorBlock)

// Start with pre-filled text
editor := prompt.NewMultiLineEditor("Hello\nWorld", prompt.CursorBlockBlink)
```

---

## Editor Methods

### StartBlinking()

Starts the cursor blinking effect if the cursor style is set to blink.

```go
editor.StartBlinking()
```

**Behavior:**
- Only activates for `CursorBlockBlink` and `CursorUnderlineBlink` styles
- Creates a goroutine with a 500ms ticker
- Toggles `cursorVisible` on every tick
- Signals renders through the `needsRender` channel
- Safe to call multiple times (checks if already running)

**Internal Details:**
- Uses a buffered channel to avoid blocking
- Goroutine exits when `stopBlink` channel receives a signal

### StopBlinking()

Stops the cursor blinking effect and cleans up the goroutine.

```go
editor.StopBlinking()
```

**Behavior:**
- Closes the `stopBlink` channel to signal the goroutine to exit
- Safe to call even if blinking is not active
- Should always be called when done with the editor (typically with `defer`)

---

### GetText() string

Returns the complete text as a single string with newlines.

```go
text := editor.GetText()
```

**Returns:** The entire buffer as a string with lines separated by `\n`

**Example:**
```go
editor := prompt.NewMultiLineEditor("Line 1\nLine 2\nLine 3", prompt.CursorBlock)
text := editor.GetText()
// text = "Line 1\nLine 2\nLine 3"
```

---

## Text Manipulation Methods

### InsertRune(r rune)

Inserts a single rune at the current cursor position.

```go
editor.InsertRune('A')
```

**Behavior:**
- Inserts the character at `cursorCol` in the current line
- Shifts existing characters to the right
- Advances cursor one position to the right

### InsertNewLine()

Inserts a new line at the current cursor position, splitting the current line.

```go
editor.InsertNewLine()
```

**Behavior:**
- Splits the current line at cursor position
- Text before cursor stays on current line
- Text after cursor moves to new line below
- Cursor moves to the beginning of the new line (column 0)

**Example:**
```go
// Before: "Hello|World" (cursor at |)
editor.InsertNewLine()
// After:
// "Hello"
// "|World" (cursor at beginning of new line)
```

### Backspace()

Deletes the character before the cursor.

```go
editor.Backspace()
```

**Behavior:**
- If cursor is in the middle of a line: deletes character before cursor, moves cursor left
- If cursor is at the beginning of a line (column 0):
  - Merges current line with previous line
  - Cursor moves to the end of the previous line
  - Current line is removed

**Example:**
```go
// Case 1: Middle of line
// "Hel|lo" → "He|lo"

// Case 2: Beginning of line
// "Hello"
// "|World" → "Hello|World"
```

### Delete()

Deletes the character at the cursor position.

```go
editor.Delete()
```

**Behavior:**
- If cursor is on a character: deletes that character, cursor stays in place
- If cursor is at the end of a line:
  - Merges next line into current line
  - Next line is removed
  - Cursor stays at the end of the merged content

**Example:**
```go
// Case 1: On character
// "Hel|lo" → "Hel|o"

// Case 2: End of line
// "Hello|"
// "World" → "Hello|World"
```

---

## Navigation Methods

### MoveLeft()

Moves the cursor one position to the left.

```go
editor.MoveLeft()
```

**Behavior:**
- If not at beginning of line: moves cursor left one column
- If at beginning of line (column 0):
  - Moves to previous line
  - Positions cursor at the end of that line

### MoveRight()

Moves the cursor one position to the right.

```go
editor.MoveRight()
```

**Behavior:**
- If not at end of line: moves cursor right one column
- If at end of line:
  - Moves to next line
  - Positions cursor at the beginning of that line (column 0)

### MoveUp()

Moves the cursor one line up.

```go
editor.MoveUp()
```

**Behavior:**
- Moves to the previous line
- Attempts to maintain the same column position
- If the previous line is shorter, moves cursor to the end of that line
- No effect if already on the first line

### MoveDown()

Moves the cursor one line down.

```go
editor.MoveDown()
```

**Behavior:**
- Moves to the next line
- Attempts to maintain the same column position
- If the next line is shorter, moves cursor to the end of that line
- No effect if already on the last line

### MoveHome()

Moves the cursor to the beginning of the current line.

```go
editor.MoveHome()
```

**Behavior:**
- Sets `cursorCol` to 0
- Current line remains unchanged

### MoveEnd()

Moves the cursor to the end of the current line.

```go
editor.MoveEnd()
```

**Behavior:**
- Sets `cursorCol` to the length of the current line
- Current line remains unchanged

---

## Line Editing Methods

### KillToEnd()

Deletes from the cursor position to the end of the current line.

```go
editor.KillToEnd()
```

**Behavior:**
- Removes all characters from cursor to end of line
- Cursor position remains unchanged
- Does not affect other lines

**Example:**
```go
// "Hello| World" → "Hello|"
```

### KillToStart()

Deletes from the beginning of the line to the cursor position.

```go
editor.KillToStart()
```

**Behavior:**
- Removes all characters from start of line to cursor
- Cursor moves to column 0
- Does not affect other lines

**Example:**
```go
// "Hello| World" → "| World"
```

---

## ColorInput Method

### RunWithMultiLineEdit() (string, error)

Displays the prompt with full multi-line editing support.

```go
input := prompt.NewWithColor("Enter your message").
    SetDefault("Initial text\nSecond line")
text, err := input.RunWithMultiLineEdit()
```

**Features:**
- Full multi-line text editing
- Arrow key navigation (up, down, left, right)
- Home/End keys for line navigation
- Enter creates new lines
- Backspace and Delete work across lines
- Various Ctrl shortcuts
- Customizable cursor styles with blinking
- Support for default multi-line values
- Validation support

**Keyboard Controls:**

| Key | Action | Alternative |
|-----|--------|-------------|
| `Enter` | Insert new line | - |
| `Ctrl+D` | Submit input | - |
| `↑` | Move up one line | `Ctrl+P` |
| `↓` | Move down one line | `Ctrl+N` |
| `←` | Move left one character | `Ctrl+B` |
| `→` | Move right one character | `Ctrl+F` |
| `Home` | Move to beginning of line | `Ctrl+A` |
| `End` | Move to end of line | `Ctrl+E` |
| `Backspace` | Delete before cursor | - |
| `Delete` | Delete at cursor | - |
| `Ctrl+K` | Delete to end of line | - |
| `Ctrl+U` | Delete to beginning of line | - |
| `Ctrl+C` | Cancel and exit | - |

**Prompt Format:**
```
❯ Your message [default value] (Ctrl+D to submit)
```

**Returns:**
- The complete text as a string with newlines
- Error if input fails

**Example:**
```go
input := prompt.NewWithColor("📝 Enter your notes").
    SetDefault("TODO:\n- Item 1\n- Item 2").
    SetCursorStyle(prompt.CursorBlockBlink)

notes, err := input.RunWithMultiLineEdit()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("You entered:\n%s\n", notes)
```

---

## Internal Rendering Functions

### renderMultiLine(editor *MultiLineEditor, previousRenderedLines int, previousCursorLine int) (int, int)

Renders the entire multi-line editor state to the terminal.

**Parameters:**
- `editor` - The editor to render
- `previousRenderedLines` - Number of lines rendered in the last render
- `previousCursorLine` - Where the cursor was positioned after the last render

**Returns:**
- `linesRendered` - Total number of lines rendered
- `cursorLineAfterRender` - Where the cursor is positioned after rendering

**Behavior:**
1. Moves cursor up to line 0 of the editor
2. Clears each line and renders content
3. Renders the cursor on the current line
4. Clears any extra lines from previous render
5. Positions the cursor on the correct line

**Terminal Control Sequences Used:**
- `\033[%dA` - Move cursor up
- `\r` - Carriage return (move to start of line)
- `\033[J` - Clear from cursor to end of screen
- `\r\n` - Move to next line

### renderLineWithCursor(line []rune, cursorCol int, cursorVisible bool, cursorStyle CursorStyle)

Renders a single line with the cursor at the specified position.

**Parameters:**
- `line` - The line content as runes
- `cursorCol` - Column position of the cursor
- `cursorVisible` - Whether the cursor should be visible (for blinking)
- `cursorStyle` - The style of cursor to render

**Behavior:**
1. Prints text before cursor
2. Renders cursor based on style and visibility:
   - **Block**: Inverted background (`\033[7m`)
   - **Underline**: Underlined character (`\033[4m`)
   - **Hidden** (during blink): Plain character
3. Prints text after cursor

---

## Usage Examples

### Basic Multi-Line Input

```go
input := prompt.NewWithColor("Enter your message")
message, err := input.RunWithMultiLineEdit()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Message:\n%s\n", message)
```

### With Default Multi-Line Value

```go
defaultTemplate := `Hello,

This is a template.

Best regards`

input := prompt.NewWithColor("Edit email template").
    SetDefault(defaultTemplate)

template, err := input.RunWithMultiLineEdit()
```

### With Blinking Cursor

```go
input := prompt.NewWithColor("Write your code").
    SetCursorStyle(prompt.CursorBlockBlink)

code, err := input.RunWithMultiLineEdit()
```

### With Validation

```go
input := prompt.NewWithColor("Enter JSON").
    SetValidator(func(s string) error {
        var js json.RawMessage
        if err := json.Unmarshal([]byte(s), &js); err != nil {
            return fmt.Errorf("invalid JSON: %w", err)
        }
        return nil
    })

jsonText, err := input.RunWithMultiLineEdit()
```

### Code Editor Example

```go
defaultCode := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`

input := prompt.NewWithColor("Edit Go code").
    SetDefault(defaultCode).
    SetCursorStyle(prompt.CursorBlockBlink).
    SetInputColor(prompt.ColorBrightCyan)

code, err := input.RunWithMultiLineEdit()
if err != nil {
    log.Fatal(err)
}

// Save to file
os.WriteFile("main.go", []byte(code), 0644)
```

### Configuration File Editor

```go
configTemplate := `# Application Configuration
server:
  host: localhost
  port: 8080

database:
  driver: postgres
  connection: user:pass@localhost/db`

input := prompt.NewWithColor("Edit configuration").
    SetDefault(configTemplate).
    SetMessageColor(prompt.ColorBrightYellow)

config, err := input.RunWithMultiLineEdit()
```

---

## Comparison: Single-Line vs Multi-Line

| Feature | RunWithEdit() | RunWithMultiLineEdit() |
|---------|---------------|------------------------|
| Multiple lines | ❌ | ✅ |
| Enter key | Submits | Creates new line |
| Submit key | `Enter` | `Ctrl+D` |
| Up/Down arrows | Not used | Navigate lines |
| Line wrapping | Terminal handles | Manual (new lines) |
| Best for | Short inputs, filenames | Code, templates, messages |
| Cursor navigation | Left/Right only | All directions |

---

## Advanced Features

### Concurrent Rendering

The multi-line editor handles concurrent events:
- User keyboard input (via `inputChan`)
- Cursor blink events (via `needsRender`)

This is managed with a select statement that waits for either event and renders appropriately.

### Thread Safety

- Uses `cursorMutex` to protect cursor state
- Ensures cursor visibility changes from blink goroutine are atomic
- Prevents race conditions between input handling and blinking

### Memory Efficiency

- Lines are stored as slices of runes (not strings) for efficient editing
- Only modified lines are updated in memory
- Terminal rendering is optimized with escape sequences

### Cursor Positioning Logic

The editor tracks:
1. `linesRendered` - How many lines were drawn
2. `cursorLineAfterRender` - Where the cursor ended up

This allows the next render to:
1. Move up to the first line
2. Redraw all lines
3. Position cursor correctly

---

## Platform Compatibility

### Supported Platforms

- ✅ **Linux**: Full support
- ✅ **macOS**: Full support
- ⚠️ **Windows**: Limited support (raw mode not available)

### Windows Fallback

On Windows, raw terminal mode is not supported. Use single-line `RunWithEdit()` or basic `Run()` instead:

```go
input := prompt.NewWithColor("Enter text")

// Check platform or catch error
text, err := input.RunWithMultiLineEdit()
if err != nil && strings.Contains(err.Error(), "raw mode not supported") {
    // Fallback to single-line
    text, err = input.RunWithEdit()
}
```

---

## Best Practices

1. **Always use `Ctrl+D` to submit**: Make this clear in your prompt or documentation
2. **Provide helpful default templates**: Users can edit them with full cursor control
3. **Choose appropriate cursor style**: Blinking cursors work well for multi-line to show position
4. **Test on target platforms**: Ensure raw mode works on your deployment environment
5. **Use validation for structured input**: Validate JSON, YAML, code syntax, etc.
6. **Consider line length**: Very long lines may not display well in all terminals
7. **Clean up properly**: Always defer `editor.StopBlinking()` to prevent goroutine leaks

---

## Error Handling

```go
input := prompt.NewWithColor("Enter your text")
text, err := input.RunWithMultiLineEdit()

if err != nil {
    if err.Error() == "raw mode not supported on Windows" {
        // Fallback strategy
        text, err = input.Run()
    } else {
        log.Fatalf("Input error: %v", err)
    }
}
```

---

## Performance Notes

- Multi-line editing is very efficient for up to hundreds of lines
- Terminal rendering uses ANSI escape sequences for fast updates
- Blink goroutine has minimal overhead (~one goroutine + 500ms ticker)
- Memory usage is proportional to text size (each line is a rune slice)
- No performance difference between cursor styles

---

## Security Considerations

- Terminal state is always restored, even on panic or error
- Ctrl+C properly cleans up and exits
- No buffer overflow issues (Go slices handle bounds automatically)
- Input is echoed visibly (not suitable for passwords)
- Validation can prevent malformed input

---

## Common Use Cases

### 1. Code Snippets
```go
input := prompt.NewWithColor("Paste your code snippet")
code, _ := input.RunWithMultiLineEdit()
```

### 2. Commit Messages
```go
input := prompt.NewWithColor("Enter commit message").
    SetDefault("feat: ")
message, _ := input.RunWithMultiLineEdit()
```

### 3. Email Templates
```go
template := "Dear [Name],\n\n\n\nBest regards,\n[Your Name]"
input := prompt.NewWithColor("Edit email").SetDefault(template)
email, _ := input.RunWithMultiLineEdit()
```

### 4. Configuration Files
```go
input := prompt.NewWithColor("Edit YAML config").
    SetDefault("key: value\n")
config, _ := input.RunWithMultiLineEdit()
```

### 5. Notes and Documentation
```go
input := prompt.NewWithColor("Add notes")
notes, _ := input.RunWithMultiLineEdit()
```

---

## Troubleshooting

### Editor not responding to arrow keys
- Ensure terminal supports ANSI escape sequences
- Check that raw mode is enabled (automatic in `RunWithMultiLineEdit()`)
- Test terminal with other interactive CLI apps

### Cursor not blinking
- Verify cursor style is `CursorBlockBlink` or `CursorUnderlineBlink`
- Check that `StartBlinking()` was called
- Ensure terminal supports the cursor escape sequences

### Text rendering issues
- Some terminals have limited line length support
- Try a different terminal emulator
- Check terminal size with `stty size`

### Can't submit with Ctrl+D
- Ensure you're pressing Ctrl+D (not just D)
- On some systems, you may need to press it twice
- Check if terminal is capturing the key combination

---

## Related Documentation

- [edit.md](./edit.md) - Single-line editing
- [color-prompt.md](./color-prompt.md) - Colored prompts
- [prompt.md](./prompt.md) - Basic prompts

---

## Example Application

See the complete example at:
- `/samples/39-input-with-multiline-edit/main.go`
