# Display Package Documentation

## Overview

The `display` package provides a comprehensive set of utilities for formatting and displaying colored terminal output. It includes functions for printing messages with various styles, colors, and semantic meanings (success, error, warning, etc.), as well as structured output like tables, boxes, banners, and more.

## File: display.go

### Color Constants

The package provides extensive ANSI escape code constants for terminal styling.

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

### Symbol Constants

Common Unicode symbols for terminal output:

- `SymbolSuccess` - ✓ (checkmark)
- `SymbolError` - ✗ (cross)
- `SymbolWarning` - ⚠ (warning triangle)
- `SymbolInfo` - ℹ (info)
- `SymbolDebug` - ● (filled circle)
- `SymbolArrow` - → (right arrow)
- `SymbolBullet` - • (bullet point)
- `SymbolCheck` - ✓ (checkmark, alias)
- `SymbolCross` - ✗ (cross, alias)
- `SymbolStar` - ★ (star)
- `SymbolHeart` - ♥ (heart)
- `SymbolDiamond` - ◆ (diamond)

### Message Types

```go
type MessageType int

const (
    MessagePlain
    MessageSuccess
    MessageError
    MessageWarning
    MessageInfo
    MessageDebug
)
```

Enumeration representing different message types for semantic output.

## Basic Print Functions

### Print

```go
func Print(message string)
```

Prints a message without a newline.

**Parameters:**
- `message` - Text to print

### Println

```go
func Println(message string)
```

Prints a message with a newline.

**Parameters:**
- `message` - Text to print

### Printf

```go
func Printf(format string, args ...any)
```

Prints a formatted message using Printf-style formatting.

**Parameters:**
- `format` - Format string
- `args` - Values for formatting

## Color Functions

### Color

```go
func Color(message string, color string)
```

Prints a colored message without newline.

**Parameters:**
- `message` - Text to print
- `color` - ANSI color code (use package constants)

### Colorln

```go
func Colorln(message string, color string)
```

Prints a colored message with newline.

**Parameters:**
- `message` - Text to print
- `color` - ANSI color code

### Colorf

```go
func Colorf(color string, format string, args ...any)
```

Prints a formatted colored message.

**Parameters:**
- `color` - ANSI color code
- `format` - Format string
- `args` - Values for formatting

## Text Style Functions

### Bold, Boldln

```go
func Bold(message string)
func Boldln(message string)
```

Print bold text (without/with newline).

### Italic, Italicln

```go
func Italic(message string)
func Italicln(message string)
```

Print italic text (without/with newline).

### Underline, Underlineln

```go
func Underline(message string)
func Underlineln(message string)
```

Print underlined text (without/with newline).

## Semantic Message Functions

### Success, Successf

```go
func Success(message string)
func Successf(format string, args ...any)
```

Prints a success message in green with a checkmark (✓).

**Parameters:**
- `message` / `format` - Success message
- `args` - Format arguments (Successf only)

**Example:**
```go
display.Success("Operation completed successfully")
display.Successf("Processed %d items", count)
```

### Error, Errorf

```go
func Error(message string)
func Errorf(format string, args ...any)
```

Prints an error message in red with a cross (✗).

**Parameters:**
- `message` / `format` - Error message
- `args` - Format arguments (Errorf only)

### Warning, Warningf

```go
func Warning(message string)
func Warningf(format string, args ...any)
```

Prints a warning message in yellow with a warning symbol (⚠).

**Parameters:**
- `message` / `format` - Warning message
- `args` - Format arguments (Warningf only)

### Info, Infof

```go
func Info(message string)
func Infof(format string, args ...any)
```

Prints an info message in cyan with an info symbol (ℹ).

**Parameters:**
- `message` / `format` - Info message
- `args` - Format arguments (Infof only)

### Debug, Debugf

```go
func Debug(message string)
func Debugf(format string, args ...any)
```

Prints a debug message in gray with a bullet symbol (●).

**Parameters:**
- `message` / `format` - Debug message
- `args` - Format arguments (Debugf only)

## Header and Title Functions

### Header, Headerf

```go
func Header(message string)
func Headerf(format string, args ...any)
```

Prints a header in bold bright cyan.

**Parameters:**
- `message` / `format` - Header text
- `args` - Format arguments (Headerf only)

### Subheader, Subheaderf

```go
func Subheader(message string)
func Subheaderf(format string, args ...any)
```

Prints a subheader in bright blue.

**Parameters:**
- `message` / `format` - Subheader text
- `args` - Format arguments (Subheaderf only)

### Title, Titlef

```go
func Title(message string)
func Titlef(format string, args ...any)
```

Prints a title with an underline separator.

**Parameters:**
- `message` / `format` - Title text
- `args` - Format arguments (Titlef only)

**Output:**
```
Title Text
──────────
```

## Separator Functions

### Separator

```go
func Separator()
```

Prints a horizontal line separator (80 characters wide).

### SeparatorWithChar

```go
func SeparatorWithChar(char string, length int)
```

Prints a separator with custom character and length.

**Parameters:**
- `char` - Character to repeat
- `length` - Number of repetitions

## List and Bullet Functions

### Bullet, Bulletf

```go
func Bullet(message string)
func Bulletf(format string, args ...any)
```

Prints a bulleted item with a bullet symbol (•).

**Parameters:**
- `message` / `format` - Item text
- `args` - Format arguments (Bulletf only)

### ColoredBullet

```go
func ColoredBullet(message string, color string)
```

Prints a colored bulleted item.

**Parameters:**
- `message` - Item text
- `color` - ANSI color code

### List, Listf

```go
func List(index int, message string)
func Listf(index int, format string, args ...any)
```

Prints a numbered list item.

**Parameters:**
- `index` - Item number
- `message` / `format` - Item text
- `args` - Format arguments (Listf only)

**Example:**
```go
display.List(1, "First item")
display.List(2, "Second item")
```

### ColoredList

```go
func ColoredList(index int, message string, color string)
```

Prints a colored numbered list item.

**Parameters:**
- `index` - Item number
- `message` - Item text
- `color` - ANSI color code

## Arrow Function

### Arrow, Arrowf

```go
func Arrow(message string)
func Arrowf(format string, args ...any)
```

Prints a message with an arrow prefix (→) in cyan.

**Parameters:**
- `message` / `format` - Message text
- `args` - Format arguments (Arrowf only)

## Box and Banner Functions

### Box

```go
func Box(message string)
```

Prints a message enclosed in a box.

**Parameters:**
- `message` - Message to display in box

**Output:**
```
┌─────────────┐
│ Message here │
└─────────────┘
```

### ColoredBox

```go
func ColoredBox(message string, color string)
```

Prints a colored message in a box.

**Parameters:**
- `message` - Message to display
- `color` - ANSI color code for the box

### Banner

```go
func Banner(message string)
```

Prints a prominent banner message in yellow/white.

**Parameters:**
- `message` - Banner text

**Output:**
```
╔═══════════════╗
║ Message here  ║
╚═══════════════╝
```

## Highlight Function

### Highlight

```go
func Highlight(message string, fgColor, bgColor string)
```

Prints a message with foreground and background colors.

**Parameters:**
- `message` - Text to highlight
- `fgColor` - Foreground color code
- `bgColor` - Background color code

## Progress and Status Functions

### Step, Stepf

```go
func Step(current, total int, message string)
func Stepf(current, total int, format string, args ...any)
```

Prints a step indicator for multi-step processes.

**Parameters:**
- `current` - Current step number
- `total` - Total number of steps
- `message` / `format` - Step description
- `args` - Format arguments (Stepf only)

**Output:**
```
[2/5] Processing data
```

### Progress, Progressf

```go
func Progress(message string)
func Progressf(format string, args ...any)
```

Prints a progress message with hourglass symbol (⏳) in yellow.

**Parameters:**
- `message` / `format` - Progress description
- `args` - Format arguments (Progressf only)

### Done, Donef

```go
func Done(message string)
func Donef(format string, args ...any)
```

Prints a completion message with checkmark in green.

**Parameters:**
- `message` / `format` - Completion message
- `args` - Format arguments (Donef only)

## Styled Output Functions

### Styled, Styledln, Styledf

```go
func Styled(message string, styles ...string)
func Styledln(message string, styles ...string)
func Styledf(format string, styles []string, args ...any)
```

Prints text with custom ANSI style combinations.

**Parameters:**
- `message` / `format` - Text to display
- `styles` - ANSI codes to combine
- `args` - Format arguments (Styledf only)

**Example:**
```go
display.Styled("Important", ColorBold, ColorRed, ColorUnderline)
```

## Indentation Functions

### Indent, Indentf

```go
func Indent(level int, message string)
func Indentf(level int, format string, args ...any)
```

Prints an indented message (2 spaces per level).

**Parameters:**
- `level` - Indentation level
- `message` / `format` - Text to indent
- `args` - Format arguments (Indentf only)

### ColoredIndent

```go
func ColoredIndent(level int, message string, color string)
```

Prints a colored indented message.

**Parameters:**
- `level` - Indentation level
- `message` - Text to indent
- `color` - ANSI color code

## Table and Key-Value Functions

### Table, Tablef

```go
func Table(key, value string)
func Tablef(key string, format string, args ...any)
```

Prints a 2-column table row with key and value.

**Parameters:**
- `key` - Left column (20 chars wide)
- `value` / `format` - Right column value
- `args` - Format arguments (Tablef only)

**Output:**
```
  Name:                John Doe
  Age:                 30
```

### KeyValue, KeyValuef

```go
func KeyValue(key, value string)
func KeyValuef(key, format string, args ...any)
```

Prints a key-value pair.

**Parameters:**
- `key` - Key name (in cyan)
- `value` / `format` - Value
- `args` - Format arguments (KeyValuef only)

**Output:**
```
name: John Doe
```

## Object-like Output Functions

### ObjectStart

```go
func ObjectStart(name string)
```

Prints the start of an object-like structure.

**Parameters:**
- `name` - Object name

**Output:**
```
ObjectName {
```

### ObjectEnd

```go
func ObjectEnd()
```

Prints the end of an object-like structure.

**Output:**
```
}
```

### Field, Fieldf

```go
func Field(key, value string)
func Fieldf(key, format string, args ...any)
```

Prints a field within an object-like structure.

**Parameters:**
- `key` - Field name
- `value` / `format` - Field value
- `args` - Format arguments (Fieldf only)

**Example:**
```go
display.ObjectStart("User")
display.Field("name", "Alice")
display.Field("age", "25")
display.ObjectEnd()
```

**Output:**
```
User {
  name: Alice
  age: 25
}
```

## Utility Functions

### NewLine

```go
func NewLine(count ...int)
```

Prints one or more newlines.

**Parameters:**
- `count` - Optional number of newlines (default: 1)

### Clear

```go
func Clear()
```

Clears the current terminal line. Useful for updating spinner or progress output.

## Usage Examples

### Basic Messaging

```go
display.Success("File saved successfully")
display.Error("Failed to connect to server")
display.Warning("Disk space running low")
display.Info("Server started on port 8080")
```

### Structured Output

```go
display.Title("User Profile")
display.Table("Name", "Alice Johnson")
display.Table("Email", "alice@example.com")
display.Table("Role", "Administrator")
```

### Multi-step Process

```go
display.Step(1, 3, "Initializing")
time.Sleep(1 * time.Second)
display.Step(2, 3, "Processing data")
time.Sleep(1 * time.Second)
display.Step(3, 3, "Finalizing")
display.Done("Process completed")
```

### Lists and Bullets

```go
display.Header("Available Options")
display.Bullet("Install dependencies")
display.Bullet("Run tests")
display.Bullet("Build project")
```

### Custom Styling

```go
display.Styled("CRITICAL", ColorBold, ColorRed, BgYellow)
display.Highlight(" Important ", ColorWhite, BgRed)
```

## Terminal Compatibility

The ANSI color codes and symbols work in most modern terminals:
- Linux/Unix terminals
- macOS Terminal and iTerm2
- Windows Terminal and PowerShell 7+
- VSCode integrated terminal
- Most CI/CD environments

## Notes

- All color output is automatically reset to prevent color bleeding
- Unicode symbols require UTF-8 terminal support
- Functions ending with 'f' support Printf-style formatting
- Functions ending with 'ln' automatically add newlines
