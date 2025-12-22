# color.prompt.go - Colored Prompts

This file provides colored, styled prompt functionality for enhanced CLI user interfaces.

## Overview

The `color.prompt.go` file implements terminal prompts with color support and customizable styling. These prompts provide a more visually appealing and professional CLI experience compared to basic prompts.

## Color Constants

### Text Colors

```go
const (
    // Standard colors
    ColorBlack   = "\033[30m"
    ColorRed     = "\033[31m"
    ColorGreen   = "\033[32m"
    ColorYellow  = "\033[33m"
    ColorBlue    = "\033[34m"
    ColorMagenta = "\033[35m"
    ColorPurple  = "\033[35m"  // Alias for Magenta
    ColorCyan    = "\033[36m"
    ColorWhite   = "\033[37m"
    ColorGray    = "\033[90m"

    // Bright colors
    ColorBrightBlack   = "\033[90m"
    ColorBrightRed     = "\033[91m"
    ColorBrightGreen   = "\033[92m"
    ColorBrightYellow  = "\033[93m"
    ColorBrightBlue    = "\033[94m"
    ColorBrightMagenta = "\033[95m"
    ColorBrightPurple  = "\033[95m"
    ColorBrightCyan    = "\033[96m"
    ColorBrightWhite   = "\033[97m"
)
```

### Text Modifiers

```go
const (
    ColorReset     = "\033[0m"
    ColorBold      = "\033[1m"
    ColorDim       = "\033[2m"
    ColorItalic    = "\033[3m"
    ColorUnderline = "\033[4m"
    ColorBlink     = "\033[5m"
    ColorReverse   = "\033[7m"
    ColorHidden    = "\033[8m"
)
```

### Background Colors

```go
const (
    // Standard background colors
    BgBlack   = "\033[40m"
    BgRed     = "\033[41m"
    BgGreen   = "\033[42m"
    BgYellow  = "\033[43m"
    BgBlue    = "\033[44m"
    BgMagenta = "\033[45m"
    BgCyan    = "\033[46m"
    BgWhite   = "\033[47m"

    // Bright background colors
    BgBrightBlack   = "\033[100m"
    BgBrightRed     = "\033[101m"
    BgBrightGreen   = "\033[102m"
    BgBrightYellow  = "\033[103m"
    BgBrightBlue    = "\033[104m"
    BgBrightMagenta = "\033[105m"
    BgBrightCyan    = "\033[106m"
    BgBrightWhite   = "\033[107m"
)
```

---

## Types

### ColorInput

Represents a colored user input prompt with customizable styling.

```go
type ColorInput struct {
    message       string
    defaultValue  string
    validator     func(string) error
    messageColor  string
    defaultColor  string
    inputColor    string
    errorColor    string
    successColor  string
    promptSymbol  string
    successSymbol string
    errorSymbol   string
}
```

#### Methods

##### NewWithColor(message string) *ColorInput

Creates a new colored input prompt with default colors.

```go
input := prompt.NewWithColor("Enter your name")
```

**Default colors:**
- Message: Cyan
- Default value: Gray
- Input: White
- Error: Red
- Success: Green
- Prompt symbol: ❯
- Success symbol: ✓
- Error symbol: ✗

##### SetDefault(value string) *ColorInput

Sets a default value for the input.

```go
input := prompt.NewWithColor("Enter port").
    SetDefault("8080")
```

##### SetValidator(validator func(string) error) *ColorInput

Sets a validation function.

```go
input := prompt.NewWithColor("Enter email").
    SetValidator(func(s string) error {
        if !strings.Contains(s, "@") {
            return fmt.Errorf("invalid email address")
        }
        return nil
    })
```

##### SetMessageColor(color string) *ColorInput

Sets the color of the message.

```go
input := prompt.NewWithColor("Warning!").
    SetMessageColor(prompt.ColorBrightYellow)
```

##### SetDefaultColor(color string) *ColorInput

Sets the color of the default value display.

```go
input := prompt.NewWithColor("Enter name").
    SetDefaultColor(prompt.ColorGray)
```

##### SetInputColor(color string) *ColorInput

Sets the color of user input.

```go
input := prompt.NewWithColor("Enter text").
    SetInputColor(prompt.ColorBrightCyan)
```

##### SetErrorColor(color string) *ColorInput

Sets the color of error messages.

```go
input := prompt.NewWithColor("Enter value").
    SetErrorColor(prompt.ColorBrightRed)
```

##### SetSuccessColor(color string) *ColorInput

Sets the color of success messages.

```go
input := prompt.NewWithColor("Enter value").
    SetSuccessColor(prompt.ColorBrightGreen)
```

##### SetColors(messageColor, defaultColor, inputColor, errorColor, successColor string) *ColorInput

Sets all colors at once.

```go
input := prompt.NewWithColor("Enter value").
    SetColors(
        prompt.ColorCyan,
        prompt.ColorGray,
        prompt.ColorWhite,
        prompt.ColorRed,
        prompt.ColorGreen,
    )
```

##### SetSymbols(prompt, success, error string) *ColorInput

Sets custom symbols for prompt, success, and error.

```go
input := prompt.NewWithColor("Enter value").
    SetSymbols("→", "✔", "✘")
```

##### Run() (string, error)

Displays the prompt and returns the user input (basic mode).

```go
name, err := input.Run()
```

##### RunWithEdit() (string, error)

Displays the prompt with full line editing support (see `edit.go` documentation).

```go
name, err := input.RunWithEdit()
```

---

### ColorConfirm

Represents a colored yes/no confirmation prompt.

```go
type ColorConfirm struct {
    message       string
    defaultValue  bool
    messageColor  string
    optionColor   string
    successColor  string
    errorColor    string
    promptSymbol  string
    successSymbol string
    errorSymbol   string
}
```

#### Methods

##### NewColorConfirm(message string) *ColorConfirm

Creates a new colored confirmation prompt.

```go
confirm := prompt.NewColorConfirm("Delete files?")
```

##### SetDefault(value bool) *ColorConfirm

Sets the default value.

```go
confirm := prompt.NewColorConfirm("Proceed?").
    SetDefault(true)
```

##### SetMessageColor(color string) *ColorConfirm

Sets the message color.

##### SetOptionColor(color string) *ColorConfirm

Sets the options (y/n) color.

##### SetSuccessColor(color string) *ColorConfirm

Sets the success indicator color.

##### SetErrorColor(color string) *ColorConfirm

Sets the error message color.

##### SetColors(messageColor, optionColor, successColor, errorColor string) *ColorConfirm

Sets all colors at once.

##### SetSymbols(prompt, success, error string) *ColorConfirm

Sets custom symbols.

##### Run() (bool, error)

Displays the confirmation and returns the choice.

```go
yes, err := confirm.Run()
if yes {
    // Proceed
}
```

---

### ColorSelect

Represents a colored selection prompt.

```go
type ColorSelect struct {
    message       string
    choices       []Choice
    defaultValue  string
    messageColor  string
    choiceColor   string
    defaultColor  string
    numberColor   string
    errorColor    string
    promptSymbol  string
    defaultSymbol string
    errorSymbol   string
}
```

#### Methods

##### NewColorSelect(message string, choices []Choice) *ColorSelect

Creates a new colored select prompt.

```go
choices := []prompt.Choice{
    {Label: "Option 1", Value: "opt1"},
    {Label: "Option 2", Value: "opt2"},
}
select := prompt.NewColorSelect("Choose option", choices)
```

##### Color Customization Methods

- `SetMessageColor(color string)`
- `SetChoiceColor(color string)`
- `SetDefaultColor(color string)`
- `SetNumberColor(color string)`
- `SetErrorColor(color string)`
- `SetColors(messageColor, choiceColor, defaultColor, numberColor, errorColor string)`
- `SetSymbols(prompt, defaultMark, error string)`

##### Run() (string, error)

Displays the selection and returns the chosen value.

---

### ColorSelectKey

Represents a selection prompt with keyboard shortcuts.

```go
type ColorSelectKey struct {
    message       string
    choices       []Choice
    defaultValue  string
    messageColor  string
    choiceColor   string
    defaultColor  string
    keyColor      string
    errorColor    string
    promptSymbol  string
    defaultSymbol string
    errorSymbol   string
}
```

#### Usage

```go
choices := []prompt.Choice{
    {Label: "Yes", Value: "y"},
    {Label: "No", Value: "n"},
    {Label: "Cancel", Value: "c"},
}
select := prompt.NewColorSelectKey("Choose action", choices)
result, err := select.Run()
```

**Note:** Each choice should have a single-character Value (the key to press).

---

### ColorMultiChoice

Represents a colored multi-choice prompt.

Similar to `ColorSelect` but allows multiple selections.

#### Methods

##### NewColorMultiChoice(message string, choices []Choice) *ColorMultiChoice

##### SetDefaults(values []string) *ColorMultiChoice

Sets default selections.

##### Run() ([]string, error)

Returns multiple selected values.

---

## Usage Examples

### Simple Colored Input

```go
input := prompt.NewWithColor("🤖 What is your name?")
name, err := input.Run()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Hello, %s!\n", name)
```

### Fully Customized Input

```go
input := prompt.NewWithColor("Enter your email").
    SetDefault("user@example.com").
    SetMessageColor(prompt.ColorBrightCyan).
    SetInputColor(prompt.ColorBrightWhite).
    SetDefaultColor(prompt.ColorGray).
    SetErrorColor(prompt.ColorBrightRed).
    SetSymbols("→", "✓", "✗").
    SetValidator(func(s string) error {
        if !strings.Contains(s, "@") {
            return fmt.Errorf("invalid email format")
        }
        return nil
    })

email, err := input.Run()
```

### Colored Confirmation

```go
confirm := prompt.NewColorConfirm("🗑️  Delete all files?").
    SetDefault(false).
    SetMessageColor(prompt.ColorBrightYellow).
    SetErrorColor(prompt.ColorBrightRed).
    SetSuccessColor(prompt.ColorBrightGreen)

if yes, _ := confirm.Run(); yes {
    fmt.Println("Deleting files...")
}
```

### Colored Selection

```go
choices := []prompt.Choice{
    {Label: "Development", Value: "dev"},
    {Label: "Staging", Value: "staging"},
    {Label: "Production", Value: "prod"},
}

select := prompt.NewColorSelect("🚀 Choose deployment environment", choices).
    SetDefault("dev").
    SetMessageColor(prompt.ColorBrightMagenta).
    SetChoiceColor(prompt.ColorWhite).
    SetDefaultColor(prompt.ColorYellow)

env, err := select.Run()
fmt.Printf("Deploying to: %s\n", env)
```

### Multi-Choice with Colors

```go
choices := []prompt.Choice{
    {Label: "JavaScript", Value: "js"},
    {Label: "Python", Value: "py"},
    {Label: "Go", Value: "go"},
    {Label: "Rust", Value: "rust"},
}

multi := prompt.NewColorMultiChoice("💻 Select programming languages", choices).
    SetDefaults([]string{"go"}).
    SetMessageColor(prompt.ColorBrightCyan).
    SetChoiceColor(prompt.ColorWhite)

languages, err := multi.Run()
for _, lang := range languages {
    fmt.Printf("Selected: %s\n", lang)
}
```

---

## Best Practices

1. **Use appropriate colors**: Choose colors that work well with both light and dark terminals
2. **Be consistent**: Use the same color scheme across your application
3. **Accessibility**: Don't rely solely on color to convey information
4. **Test your colors**: Test on different terminal emulators
5. **Use bright colors sparingly**: They can be hard to read on some terminals

## Notes

- All colored prompts support the same functionality as basic prompts
- Colors are ANSI escape sequences
- Colors are reset automatically after each prompt
- Use `ColorReset` to reset colors manually
- Supports both Run() and RunWithEdit() methods for input
