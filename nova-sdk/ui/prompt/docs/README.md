# Prompt Package Documentation

Complete documentation for the `github.com/snipwise/nova/nova-sdk/ui/prompt` package.

## Overview

The prompt package provides a comprehensive set of tools for building interactive CLI applications with user input, selections, confirmations, and advanced line editing capabilities.

## Package Contents

### Core Files

1. **[prompt.go](./prompt.md)** - Basic Prompts
   - Simple text input
   - Yes/No confirmations
   - Single selection
   - Multi-selection
   - No color support

2. **[color-prompt.go](./color-prompt.md)** - Colored Prompts
   - All features from `prompt.go` with color support
   - Customizable colors and symbols
   - Professional CLI appearance
   - ANSI color constants

3. **[edit.go](./edit.md)** - Interactive Line Editing
   - Arrow key navigation
   - Full line editing capabilities
   - Customizable cursor styles (block, underline, blinking)
   - Raw terminal mode
   - Keyboard shortcuts (Home, End, Ctrl+K, etc.)

4. **[edit.multi.line.go](./edit.multi.line.md)** - Multi-Line Text Editor
   - Full multi-line text editing
   - Cursor navigation across lines (up, down, left, right)
   - Line insertion and deletion
   - Multi-line default values
   - Submit with Ctrl+D
   - Complete text editor experience in the terminal

5. **[tool-confirmation.go](./tool-confirmation.md)** - Tool Execution Confirmation
   - Human-in-the-loop for AI agents
   - Confirm/Deny/Quit workflow
   - Integration with Nova SDK agents

---

## Quick Start

### Installation

```bash
go get github.com/snipwise/nova
```

### Import

```go
import "github.com/snipwise/nova/nova-sdk/ui/prompt"
```

---

## Feature Matrix

| Feature | Basic Prompt | Colored Prompt | With Editing | Multi-Line |
|---------|-------------|----------------|--------------|------------|
| Text Input | ✅ | ✅ | ✅ | ✅ |
| Confirmation | ✅ | ✅ | N/A | N/A |
| Selection | ✅ | ✅ | N/A | N/A |
| Multi-Selection | ✅ | ✅ | N/A | N/A |
| Colors | ❌ | ✅ | ✅ | ✅ |
| Custom Symbols | ❌ | ✅ | ✅ | ✅ |
| Validation | ✅ | ✅ | ✅ | ✅ |
| Default Values | ✅ | ✅ | ✅ | ✅ |
| Arrow Keys | ❌ | ❌ | ✅ (←→) | ✅ (←→↑↓) |
| Line Editing | ❌ | ❌ | ✅ | ✅ |
| Multiple Lines | ❌ | ❌ | ❌ | ✅ |
| Cursor Styles | ❌ | ❌ | ✅ | ✅ |
| Ctrl Shortcuts | ❌ | ❌ | ✅ | ✅ |
| Submit Key | Enter | Enter | Enter | Ctrl+D |

---

## Common Use Cases

### 1. Simple User Input

```go
input := prompt.NewWithColor("What is your name?")
name, err := input.Run()
```

**Documentation**: [color-prompt.md](./color-prompt.md#colorinput)

---

### 2. Editable Input with Default

```go
input := prompt.NewWithColor("Edit configuration path").
    SetDefault("/etc/myapp/config.json").
    SetCursorStyle(prompt.CursorBlockBlink)
path, err := input.RunWithEdit()
```

**Documentation**: [edit.md](./edit.md#runwithedit-string-error)

---

### 2b. Multi-Line Editable Input

```go
template := "Dear [Name],\n\nBest regards"
input := prompt.NewWithColor("Edit email template").
    SetDefault(template).
    SetCursorStyle(prompt.CursorBlockBlink)
text, err := input.RunWithMultiLineEdit()
```

**Documentation**: [edit.multi.line.md](./edit.multi.line.md#runwithmultilineedit-string-error)

---

### 3. Yes/No Confirmation

```go
confirm := prompt.NewColorConfirm("Delete all files?").
    SetDefault(false)
if yes, _ := confirm.Run(); yes {
    // Delete files
}
```

**Documentation**: [color-prompt.md](./color-prompt.md#colorconfirm)

---

### 4. Selection Menu

```go
choices := []prompt.Choice{
    {Label: "Development", Value: "dev"},
    {Label: "Production", Value: "prod"},
}
selectEnv := prompt.NewColorSelect("Choose environment", choices)
env, err := selectEnv.Run()
```

**Documentation**: [color-prompt.md](./color-prompt.md#colorselect)

---

### 5. Validated Input

```go
input := prompt.NewWithColor("Enter email").
    SetValidator(func(s string) error {
        if !strings.Contains(s, "@") {
            return fmt.Errorf("invalid email")
        }
        return nil
    })
email, err := input.Run()
```

**Documentation**: [prompt.md](./prompt.md#setvalidator)

---

### 6. AI Agent Tool Confirmation

```go
response := prompt.HumanConfirmation("Execute database migration?")

switch response {
case tools.Confirmed:
    // Execute
case tools.Denied:
    // Skip
case tools.Quit:
    // Stop agent
}
```

**Documentation**: [tool-confirmation.md](./tool-confirmation.md)

---

## Choosing the Right Prompt Type

### Use Basic Prompt (`prompt.go`) when:
- Building simple CLI tools
- No need for colors
- Terminal doesn't support ANSI colors
- Minimal dependencies required

### Use Colored Prompt (`color-prompt.go`) when:
- Building professional CLI applications
- Want better visual feedback
- Terminal supports colors (most modern terminals)
- Need to highlight important information

### Use RunWithEdit (`edit.go`) when:
- Users need to edit single-line inputs
- Providing default values to be modified
- Want professional editing experience (single line)
- Building configuration tools
- Terminal supports raw mode (Linux/macOS)

### Use RunWithMultiLineEdit (`edit.multi.line.go`) when:
- Users need to enter or edit multi-line text
- Editing code snippets, templates, or messages
- Need full cursor navigation (up/down/left/right)
- Building text editors or note-taking tools
- Want to edit multi-line default values
- Terminal supports raw mode (Linux/macOS)

### Use Tool Confirmation (`tool-confirmation.go`) when:
- Building AI agent systems
- Need human oversight for dangerous operations
- Want audit trail of user approvals
- Using Nova SDK agents with tools

---

## Available Colors

### Standard Colors
```go
ColorBlack, ColorRed, ColorGreen, ColorYellow
ColorBlue, ColorMagenta, ColorCyan, ColorWhite, ColorGray
```

### Bright Colors
```go
ColorBrightRed, ColorBrightGreen, ColorBrightYellow
ColorBrightBlue, ColorBrightMagenta, ColorBrightCyan, ColorBrightWhite
```

### Text Modifiers
```go
ColorBold, ColorDim, ColorItalic, ColorUnderline
ColorBlink, ColorReverse, ColorHidden
```

**Full list**: [color-prompt.md](./color-prompt.md#color-constants)

---

## Cursor Styles

```go
CursorBlock          // Solid block (default)
CursorBlockBlink     // Blinking block
CursorUnderline      // Solid underline
CursorUnderlineBlink // Blinking underline
```

**Documentation**: [edit.md](./edit.md#cursor-styles)

---

## Keyboard Shortcuts

### Single-Line Editing (RunWithEdit)

**Navigation:**
- `←` / `Ctrl+B` - Move left
- `→` / `Ctrl+F` - Move right
- `Home` / `Ctrl+A` - Beginning
- `End` / `Ctrl+E` - End

**Editing:**
- `Backspace` - Delete before cursor
- `Delete` - Delete at cursor
- `Ctrl+K` - Delete to end
- `Ctrl+U` - Delete to beginning

**Control:**
- `Enter` - Submit
- `Ctrl+C` - Cancel
- `Ctrl+D` - EOF

**Full list**: [edit.md](./edit.md#keyboard-shortcuts)

---

### Multi-Line Editing (RunWithMultiLineEdit)

**Navigation:**
- `←` / `Ctrl+B` - Move left
- `→` / `Ctrl+F` - Move right
- `↑` / `Ctrl+P` - Move up
- `↓` / `Ctrl+N` - Move down
- `Home` / `Ctrl+A` - Beginning of line
- `End` / `Ctrl+E` - End of line

**Editing:**
- `Enter` - Insert new line
- `Backspace` - Delete before cursor
- `Delete` - Delete at cursor
- `Ctrl+K` - Delete to end of line
- `Ctrl+U` - Delete to beginning of line

**Control:**
- `Ctrl+D` - Submit
- `Ctrl+C` - Cancel

**Full list**: [edit.multi.line.md](./edit.multi.line.md#keyboard-controls)

---

## Examples by Category

### Basic I/O
- [Simple input](./prompt.md#basic-input)
- [Input with default](./prompt.md#input-with-default-and-validation)
- [Confirmation](./prompt.md#confirmation)

### Selections
- [Single selection](./prompt.md#selection)
- [Multi-selection](./prompt.md#multi-choice)
- [Keyboard shortcuts selection](./color-prompt.md#colorselectkey)

### Advanced
- [Editable input (single-line)](./edit.md#edit-default-value)
- [Editable input (multi-line)](./edit.multi.line.md#with-default-multi-line-value)
- [Cursor customization](./edit.md#cursor-styles)
- [Global cursor style](./edit.md#global-cursor-style)
- [Colored + editable](./edit.md#colored-input-with-editing)
- [Multi-line with validation](./edit.multi.line.md#with-validation)

### Validation
- [Custom validation](./prompt.md#input-with-default-and-validation)
- [Email validation](./edit.md#with-validation)
- [Retry on failure](./prompt.md#setvalidator)

### AI/Agents
- [Tool confirmation](./tool-confirmation.md#basic-tool-confirmation)
- [File operations](./tool-confirmation.md#file-operations)
- [API calls](./tool-confirmation.md#api-calls)

---

## Platform Support

| Platform | Basic | Colored | Editing | Multi-Line |
|----------|-------|---------|---------|------------|
| Linux | ✅ | ✅ | ✅ | ✅ |
| macOS | ✅ | ✅ | ✅ | ✅ |
| Windows | ✅ | ✅ | ⚠️ Limited | ⚠️ Limited |

**Note**: `RunWithEdit()` and `RunWithMultiLineEdit()` have limited support on Windows due to raw terminal mode requirements. Use `Run()` as fallback.

---

## Best Practices

### 1. Always Handle Errors
```go
name, err := input.Run()
if err != nil {
    log.Fatal(err)
}
```

### 2. Provide Clear Messages
```go
// Good
prompt.NewWithColor("Enter your email address")

// Bad
prompt.NewWithColor("Email")
```

### 3. Use Appropriate Defaults
```go
confirm := prompt.NewColorConfirm("Delete files?").
    SetDefault(false) // Safe default
```

### 4. Validate User Input
```go
input.SetValidator(func(s string) error {
    if len(s) < 3 {
        return fmt.Errorf("too short")
    }
    return nil
})
```

### 5. Use Colors Consistently
```go
// Define color scheme
const (
    promptColor = prompt.ColorBrightCyan
    errorColor  = prompt.ColorBrightRed
    successColor = prompt.ColorBrightGreen
)

// Use throughout app
input.SetMessageColor(promptColor)
input.SetErrorColor(errorColor)
```

### 6. Choose Right Cursor for Context
```go
// For quick input
input.SetCursorStyle(prompt.CursorBlock)

// To draw attention
input.SetCursorStyle(prompt.CursorBlockBlink)

// For minimal distraction
input.SetCursorStyle(prompt.CursorUnderline)
```

---

## Migration Guide

### From Basic to Colored

```go
// Before
input := prompt.New("Enter name")

// After
input := prompt.NewWithColor("Enter name")
```

### From Run() to RunWithEdit()

```go
// Before
name, err := input.Run()

// After (single-line)
name, err := input.RunWithEdit()

// After (multi-line)
name, err := input.RunWithMultiLineEdit()
```

### Adding Cursor Style

```go
input := prompt.NewWithColor("Enter name").
    SetCursorStyle(prompt.CursorBlockBlink) // Add this line

// Single-line
name, err := input.RunWithEdit()

// Multi-line
name, err := input.RunWithMultiLineEdit()
```

---

## Troubleshooting

### Colors not showing
- Check if terminal supports ANSI colors
- Try a different terminal emulator
- Test with `echo -e "\033[31mRed Text\033[0m"`

### Edit mode not working
- Ensure platform is Linux or macOS
- Check if `stty` command is available
- On Windows, use `Run()` instead of `RunWithEdit()` or `RunWithMultiLineEdit()`
- Verify terminal supports raw mode

### Cursor not visible
- System cursor is hidden in edit mode
- Custom cursor is drawn by the library
- Check if `hideCursor` sequence is supported

### Validation not triggering
- Ensure validator is set before `Run()`
- Validator must return error for invalid input
- Check if validation logic is correct

---

## API Reference

### Complete Type Reference

#### Input Types
- `Input` - [prompt.md](./prompt.md#input)
- `ColorInput` - [color-prompt.md](./color-prompt.md#colorinput)

#### Confirmation Types
- `Confirm` - [prompt.md](./prompt.md#confirm)
- `ColorConfirm` - [color-prompt.md](./color-prompt.md#colorconfirm)

#### Selection Types
- `Select` - [prompt.md](./prompt.md#select)
- `ColorSelect` - [color-prompt.md](./color-prompt.md#colorselect)
- `ColorSelectKey` - [color-prompt.md](./color-prompt.md#colorselectkey)

#### Multi-Selection Types
- `MultiChoice` - [prompt.md](./prompt.md#multichoice)
- `ColorMultiChoice` - [color-prompt.md](./color-prompt.md#colormultichoice)

#### Shared Types
- `Choice` - [prompt.md](./prompt.md#choice)
- `CursorStyle` - [edit.md](./edit.md#cursorstyle)

### Complete Function Reference

#### Constructors
- `New()` - [prompt.md](./prompt.md#newmessage-string-input)
- `NewWithColor()` - [color-prompt.md](./color-prompt.md#newwithcolormessage-string-colorinput)
- `NewConfirm()` - [prompt.md](./prompt.md#newconfirmmessage-string-confirm)
- `NewColorConfirm()` - [color-prompt.md](./color-prompt.md#newcolorconfirmmessage-string-colorconfirm)
- `NewSelect()` - [prompt.md](./prompt.md#newselectmessage-string-choices-choice-select)
- `NewColorSelect()` - [color-prompt.md](./color-prompt.md#newcolorselectmessage-string-choices-choice-colorselect)
- `NewColorSelectKey()` - [color-prompt.md](./color-prompt.md#newcolorselectkeymessage-string-choices-choice-colorselectkey)
- `NewMultiChoice()` - [prompt.md](./prompt.md#newmultichoicemessage-string-choices-choice-multichoice)
- `NewColorMultiChoice()` - [color-prompt.md](./color-prompt.md#newcolormultichoicemessage-string-choices-choice-colormultichoice)

#### Global Functions
- `SetCursorStyle()` - [edit.md](./edit.md#setcursorstylestyle-cursorstyle)
- `HumanConfirmation()` - [tool-confirmation.md](./tool-confirmation.md#humanconfirmationtext-string-toolsconfirmationresponse)

---

## Contributing

When contributing to this package:

1. Follow existing patterns
2. Add documentation for new features
3. Include examples
4. Test on Linux, macOS, and Windows
5. Update this README

---

## Related Packages

- `github.com/snipwise/nova/nova-sdk/agents` - AI agent framework
- `github.com/snipwise/nova/nova-sdk/agents/tools` - Agent tools and confirmation types
- `github.com/snipwise/nova/nova-sdk/ui/display` - Output formatting and display utilities

---

## License

Part of the Nova SDK by SnipWise.

---

## Support

For issues, questions, or contributions:
- GitHub Issues: [Nova SDK Issues](https://github.com/snipwise/nova/issues)
- Documentation: This directory
- Examples:
  - Single-line editing: `/samples/38-input-with-edit/`
  - Multi-line editing: `/samples/39-input-with-multiline-edit/`
