package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Color codes for terminal output (same as spinner package)
const (
	// Reset and modifiers
	ColorReset     = "\033[0m"
	ColorBold      = "\033[1m"
	ColorDim       = "\033[2m"
	ColorItalic    = "\033[3m"
	ColorUnderline = "\033[4m"
	ColorBlink     = "\033[5m"
	ColorReverse   = "\033[7m"
	ColorHidden    = "\033[8m"

	// Standard colors
	ColorBlack   = "\033[30m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorPurple  = "\033[35m" // Alias for Magenta
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
	ColorBrightPurple  = "\033[95m" // Alias for Bright Magenta
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	// Background colors
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

// ColorInput represents a user input prompt with color support
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

// NewWithColor creates a new colored input prompt with a message
func NewWithColor(message string) *ColorInput {
	return &ColorInput{
		message:       message,
		defaultValue:  "",
		validator:     nil,
		messageColor:  ColorCyan,
		defaultColor:  ColorGray,
		inputColor:    ColorWhite,
		errorColor:    ColorRed,
		successColor:  ColorGreen,
		promptSymbol:  "❯",
		successSymbol: "✓",
		errorSymbol:   "✗",
	}
}

// SetDefault sets a default value for the input
func (i *ColorInput) SetDefault(value string) *ColorInput {
	i.defaultValue = value
	return i
}

// SetValidator sets a validation function
func (i *ColorInput) SetValidator(validator func(string) error) *ColorInput {
	i.validator = validator
	return i
}

// SetMessageColor sets the color of the message
func (i *ColorInput) SetMessageColor(color string) *ColorInput {
	i.messageColor = color
	return i
}

// SetDefaultColor sets the color of the default value display
func (i *ColorInput) SetDefaultColor(color string) *ColorInput {
	i.defaultColor = color
	return i
}

// SetInputColor sets the color of user input
func (i *ColorInput) SetInputColor(color string) *ColorInput {
	i.inputColor = color
	return i
}

// SetErrorColor sets the color of error messages
func (i *ColorInput) SetErrorColor(color string) *ColorInput {
	i.errorColor = color
	return i
}

// SetSuccessColor sets the color of success messages
func (i *ColorInput) SetSuccessColor(color string) *ColorInput {
	i.successColor = color
	return i
}

// SetColors sets all colors at once
func (i *ColorInput) SetColors(messageColor, defaultColor, inputColor, errorColor, successColor string) *ColorInput {
	i.messageColor = messageColor
	i.defaultColor = defaultColor
	i.inputColor = inputColor
	i.errorColor = errorColor
	i.successColor = successColor
	return i
}

// SetSymbols sets custom symbols for prompt, success, and error
func (i *ColorInput) SetSymbols(prompt, success, error string) *ColorInput {
	i.promptSymbol = prompt
	i.successSymbol = success
	i.errorSymbol = error
	return i
}

// Run displays the prompt and returns the user input
func (i *ColorInput) Run() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display the prompt message with colors
		if i.defaultValue != "" {
			fmt.Printf("%s%s %s%s %s[%s]%s: ",
				i.messageColor, i.promptSymbol, i.message, ColorReset,
				i.defaultColor, i.defaultValue, ColorReset)
		} else {
			fmt.Printf("%s%s %s%s: ",
				i.messageColor, i.promptSymbol, i.message, ColorReset)
		}

		// Print input color
		fmt.Print(i.inputColor)

		// Read user input
		input, err := reader.ReadString('\n')
		fmt.Print(ColorReset) // Reset color after input

		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		// Clean the input
		input = strings.TrimSpace(input)

		// Use default value if input is empty
		if input == "" && i.defaultValue != "" {
			input = i.defaultValue
		}

		// Validate if a validator is set
		if i.validator != nil {
			if err := i.validator(input); err != nil {
				fmt.Printf("%s%s %s%s\n", i.errorColor, i.errorSymbol, err.Error(), ColorReset)
				continue
			}
		}

		return input, nil
	}
}

// ColorConfirm represents a yes/no confirmation prompt with color support
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

// NewColorConfirm creates a new colored confirmation prompt
func NewColorConfirm(message string) *ColorConfirm {
	return &ColorConfirm{
		message:       message,
		defaultValue:  false,
		messageColor:  ColorCyan,
		optionColor:   ColorGray,
		successColor:  ColorGreen,
		errorColor:    ColorRed,
		promptSymbol:  "❯",
		successSymbol: "✓",
		errorSymbol:   "✗",
	}
}

// SetDefault sets the default value for the confirmation
func (c *ColorConfirm) SetDefault(value bool) *ColorConfirm {
	c.defaultValue = value
	return c
}

// SetMessageColor sets the color of the message
func (c *ColorConfirm) SetMessageColor(color string) *ColorConfirm {
	c.messageColor = color
	return c
}

// SetOptionColor sets the color of the options display
func (c *ColorConfirm) SetOptionColor(color string) *ColorConfirm {
	c.optionColor = color
	return c
}

// SetSuccessColor sets the color of success indicator
func (c *ColorConfirm) SetSuccessColor(color string) *ColorConfirm {
	c.successColor = color
	return c
}

// SetErrorColor sets the color of error messages
func (c *ColorConfirm) SetErrorColor(color string) *ColorConfirm {
	c.errorColor = color
	return c
}

// SetColors sets all colors at once
func (c *ColorConfirm) SetColors(messageColor, optionColor, successColor, errorColor string) *ColorConfirm {
	c.messageColor = messageColor
	c.optionColor = optionColor
	c.successColor = successColor
	c.errorColor = errorColor
	return c
}

// SetSymbols sets custom symbols
func (c *ColorConfirm) SetSymbols(prompt, success, error string) *ColorConfirm {
	c.promptSymbol = prompt
	c.successSymbol = success
	c.errorSymbol = error
	return c
}

// Run displays the confirmation prompt and returns the user's choice
func (c *ColorConfirm) Run() (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display the prompt with default indicator
		defaultStr := "y/N"
		defaultColor := c.errorColor
		if c.defaultValue {
			defaultStr = "Y/n"
			defaultColor = c.successColor
		}

		fmt.Printf("%s%s %s%s %s(%s)%s: ",
			c.messageColor, c.promptSymbol, c.message, ColorReset,
			defaultColor, defaultStr, ColorReset)

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("error reading input: %w", err)
		}

		// Clean the input
		input = strings.ToLower(strings.TrimSpace(input))

		// Use default value if input is empty
		if input == "" {
			return c.defaultValue, nil
		}

		// Parse the response
		switch input {
		case "y", "yes", "o", "oui":
			return true, nil
		case "n", "no", "non":
			return false, nil
		default:
			fmt.Printf("%s%s Please answer with 'y' or 'n'%s\n",
				c.errorColor, c.errorSymbol, ColorReset)
			continue
		}
	}
}

// ColorSelect represents a selection prompt with color support
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

// NewColorSelect creates a new colored select prompt
func NewColorSelect(message string, choices []Choice) *ColorSelect {
	return &ColorSelect{
		message:       message,
		choices:       choices,
		defaultValue:  "",
		messageColor:  ColorCyan,
		choiceColor:   ColorWhite,
		defaultColor:  ColorYellow,
		numberColor:   ColorGray,
		errorColor:    ColorRed,
		promptSymbol:  "❯",
		defaultSymbol: "●",
		errorSymbol:   "✗",
	}
}

// SetDefault sets the default choice by value
func (s *ColorSelect) SetDefault(value string) *ColorSelect {
	s.defaultValue = value
	return s
}

// SetMessageColor sets the color of the message
func (s *ColorSelect) SetMessageColor(color string) *ColorSelect {
	s.messageColor = color
	return s
}

// SetChoiceColor sets the color of choice labels
func (s *ColorSelect) SetChoiceColor(color string) *ColorSelect {
	s.choiceColor = color
	return s
}

// SetDefaultColor sets the color of the default choice indicator
func (s *ColorSelect) SetDefaultColor(color string) *ColorSelect {
	s.defaultColor = color
	return s
}

// SetNumberColor sets the color of choice numbers
func (s *ColorSelect) SetNumberColor(color string) *ColorSelect {
	s.numberColor = color
	return s
}

// SetErrorColor sets the color of error messages
func (s *ColorSelect) SetErrorColor(color string) *ColorSelect {
	s.errorColor = color
	return s
}

// SetColors sets all colors at once
func (s *ColorSelect) SetColors(messageColor, choiceColor, defaultColor, numberColor, errorColor string) *ColorSelect {
	s.messageColor = messageColor
	s.choiceColor = choiceColor
	s.defaultColor = defaultColor
	s.numberColor = numberColor
	s.errorColor = errorColor
	return s
}

// SetSymbols sets custom symbols
func (s *ColorSelect) SetSymbols(prompt, defaultMark, error string) *ColorSelect {
	s.promptSymbol = prompt
	s.defaultSymbol = defaultMark
	s.errorSymbol = error
	return s
}

// Run displays the selection prompt and returns the selected value
func (s *ColorSelect) Run() (string, error) {
	if len(s.choices) == 0 {
		return "", fmt.Errorf("no choices available")
	}

	reader := bufio.NewReader(os.Stdin)

	// Display choices
	fmt.Printf("%s%s %s%s\n", s.messageColor, s.promptSymbol, s.message, ColorReset)
	for i, choice := range s.choices {
		prefix := fmt.Sprintf("%s%d)%s", s.numberColor, i+1, ColorReset)
		if choice.Value == s.defaultValue {
			fmt.Printf("  %s %s%s%s %s%s%s\n",
				prefix,
				s.choiceColor, choice.Label, ColorReset,
				s.defaultColor, s.defaultSymbol, ColorReset)
		} else {
			fmt.Printf("  %s %s%s%s\n",
				prefix,
				s.choiceColor, choice.Label, ColorReset)
		}
	}

	for {
		// Display the prompt
		if s.defaultValue != "" {
			fmt.Printf("%sEnter choice [1-%d] (default: %s)%s: ",
				s.numberColor, len(s.choices), s.getDefaultLabel(), ColorReset)
		} else {
			fmt.Printf("%sEnter choice [1-%d]%s: ",
				s.numberColor, len(s.choices), ColorReset)
		}

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		// Clean the input
		input = strings.TrimSpace(input)

		// Use default value if input is empty
		if input == "" && s.defaultValue != "" {
			return s.defaultValue, nil
		}

		// Parse the input as a number
		var choiceNum int
		_, err = fmt.Sscanf(input, "%d", &choiceNum)
		if err != nil || choiceNum < 1 || choiceNum > len(s.choices) {
			fmt.Printf("%s%s Please enter a number between 1 and %d%s\n",
				s.errorColor, s.errorSymbol, len(s.choices), ColorReset)
			continue
		}

		return s.choices[choiceNum-1].Value, nil
	}
}

// getDefaultLabel returns the label of the default choice
func (s *ColorSelect) getDefaultLabel() string {
	for i, choice := range s.choices {
		if choice.Value == s.defaultValue {
			return fmt.Sprintf("%d", i+1)
		}
	}
	return ""
}

// ColorSelectKey represents a selection prompt with keyboard shortcuts
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

// NewColorSelectKey creates a new colored select prompt with keyboard shortcuts
// Each choice should have a single-character Value (the key to press)
func NewColorSelectKey(message string, choices []Choice) *ColorSelectKey {
	return &ColorSelectKey{
		message:       message,
		choices:       choices,
		defaultValue:  "",
		messageColor:  ColorCyan,
		choiceColor:   ColorWhite,
		defaultColor:  ColorYellow,
		keyColor:      ColorGray,
		errorColor:    ColorRed,
		promptSymbol:  "❯",
		defaultSymbol: "●",
		errorSymbol:   "✗",
	}
}

// SetDefault sets the default choice by value
func (s *ColorSelectKey) SetDefault(value string) *ColorSelectKey {
	s.defaultValue = value
	return s
}

// SetMessageColor sets the color of the message
func (s *ColorSelectKey) SetMessageColor(color string) *ColorSelectKey {
	s.messageColor = color
	return s
}

// SetChoiceColor sets the color of choice labels
func (s *ColorSelectKey) SetChoiceColor(color string) *ColorSelectKey {
	s.choiceColor = color
	return s
}

// SetDefaultColor sets the color of the default choice indicator
func (s *ColorSelectKey) SetDefaultColor(color string) *ColorSelectKey {
	s.defaultColor = color
	return s
}

// SetKeyColor sets the color of keyboard shortcuts
func (s *ColorSelectKey) SetKeyColor(color string) *ColorSelectKey {
	s.keyColor = color
	return s
}

// SetErrorColor sets the color of error messages
func (s *ColorSelectKey) SetErrorColor(color string) *ColorSelectKey {
	s.errorColor = color
	return s
}

// SetColors sets all colors at once
func (s *ColorSelectKey) SetColors(messageColor, choiceColor, defaultColor, keyColor, errorColor string) *ColorSelectKey {
	s.messageColor = messageColor
	s.choiceColor = choiceColor
	s.defaultColor = defaultColor
	s.keyColor = keyColor
	s.errorColor = errorColor
	return s
}

// SetSymbols sets custom symbols
func (s *ColorSelectKey) SetSymbols(prompt, defaultMark, error string) *ColorSelectKey {
	s.promptSymbol = prompt
	s.defaultSymbol = defaultMark
	s.errorSymbol = error
	return s
}

// Run displays the selection prompt and returns the selected value
func (s *ColorSelectKey) Run() (string, error) {
	if len(s.choices) == 0 {
		return "", fmt.Errorf("no choices available")
	}

	reader := bufio.NewReader(os.Stdin)

	// Build valid keys map
	validKeys := make(map[string]string)
	var keyList []string
	for _, choice := range s.choices {
		validKeys[strings.ToLower(choice.Value)] = choice.Value
		keyList = append(keyList, choice.Value)
	}

	// Display message and choices
	fmt.Printf("%s%s %s%s\n", s.messageColor, s.promptSymbol, s.message, ColorReset)
	for _, choice := range s.choices {
		if choice.Value == s.defaultValue {
			fmt.Printf("  %s%s)%s %s%s%s %s%s%s\n",
				s.keyColor, choice.Value, ColorReset,
				s.choiceColor, choice.Label, ColorReset,
				s.defaultColor, s.defaultSymbol, ColorReset)
		} else {
			fmt.Printf("  %s%s)%s %s%s%s\n",
				s.keyColor, choice.Value, ColorReset,
				s.choiceColor, choice.Label, ColorReset)
		}
	}

	for {
		// Display the prompt
		if s.defaultValue != "" {
			fmt.Printf("%sEnter choice [%s] (default: %s)%s: ",
				s.keyColor, strings.Join(keyList, "/"), s.defaultValue, ColorReset)
		} else {
			fmt.Printf("%sEnter choice [%s]%s: ",
				s.keyColor, strings.Join(keyList, "/"), ColorReset)
		}

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		// Clean the input
		input = strings.ToLower(strings.TrimSpace(input))

		// Use default value if input is empty
		if input == "" && s.defaultValue != "" {
			return s.defaultValue, nil
		}

		// Check if the input is a valid key
		if selectedValue, exists := validKeys[input]; exists {
			return selectedValue, nil
		}

		fmt.Printf("%s%s Please enter one of: %s%s\n",
			s.errorColor, s.errorSymbol, strings.Join(keyList, ", "), ColorReset)
	}
}

// ColorMultiChoice represents a multi-choice prompt with color support
type ColorMultiChoice struct {
	message       string
	choices       []Choice
	defaultValues []string
	messageColor  string
	choiceColor   string
	defaultColor  string
	numberColor   string
	errorColor    string
	promptSymbol  string
	defaultSymbol string
	errorSymbol   string
}

// NewColorMultiChoice creates a new colored multi-choice prompt
func NewColorMultiChoice(message string, choices []Choice) *ColorMultiChoice {
	return &ColorMultiChoice{
		message:       message,
		choices:       choices,
		defaultValues: []string{},
		messageColor:  ColorCyan,
		choiceColor:   ColorWhite,
		defaultColor:  ColorYellow,
		numberColor:   ColorGray,
		errorColor:    ColorRed,
		promptSymbol:  "❯",
		defaultSymbol: "●",
		errorSymbol:   "✗",
	}
}

// SetDefaults sets the default choices by values
func (m *ColorMultiChoice) SetDefaults(values []string) *ColorMultiChoice {
	m.defaultValues = values
	return m
}

// SetMessageColor sets the color of the message
func (m *ColorMultiChoice) SetMessageColor(color string) *ColorMultiChoice {
	m.messageColor = color
	return m
}

// SetChoiceColor sets the color of choice labels
func (m *ColorMultiChoice) SetChoiceColor(color string) *ColorMultiChoice {
	m.choiceColor = color
	return m
}

// SetDefaultColor sets the color of default choice indicators
func (m *ColorMultiChoice) SetDefaultColor(color string) *ColorMultiChoice {
	m.defaultColor = color
	return m
}

// SetNumberColor sets the color of choice numbers
func (m *ColorMultiChoice) SetNumberColor(color string) *ColorMultiChoice {
	m.numberColor = color
	return m
}

// SetErrorColor sets the color of error messages
func (m *ColorMultiChoice) SetErrorColor(color string) *ColorMultiChoice {
	m.errorColor = color
	return m
}

// SetColors sets all colors at once
func (m *ColorMultiChoice) SetColors(messageColor, choiceColor, defaultColor, numberColor, errorColor string) *ColorMultiChoice {
	m.messageColor = messageColor
	m.choiceColor = choiceColor
	m.defaultColor = defaultColor
	m.numberColor = numberColor
	m.errorColor = errorColor
	return m
}

// SetSymbols sets custom symbols
func (m *ColorMultiChoice) SetSymbols(prompt, defaultMark, error string) *ColorMultiChoice {
	m.promptSymbol = prompt
	m.defaultSymbol = defaultMark
	m.errorSymbol = error
	return m
}

// Run displays the multi-choice prompt and returns the selected values
func (m *ColorMultiChoice) Run() ([]string, error) {
	if len(m.choices) == 0 {
		return nil, fmt.Errorf("no choices available")
	}

	reader := bufio.NewReader(os.Stdin)

	// Display choices
	fmt.Printf("%s%s %s%s\n", m.messageColor, m.promptSymbol, m.message, ColorReset)
	for i, choice := range m.choices {
		prefix := fmt.Sprintf("%s%d)%s", m.numberColor, i+1, ColorReset)
		isDefault := false
		for _, dv := range m.defaultValues {
			if choice.Value == dv {
				isDefault = true
				break
			}
		}
		if isDefault {
			fmt.Printf("  %s %s%s%s %s%s%s\n",
				prefix,
				m.choiceColor, choice.Label, ColorReset,
				m.defaultColor, m.defaultSymbol, ColorReset)
		} else {
			fmt.Printf("  %s %s%s%s\n",
				prefix,
				m.choiceColor, choice.Label, ColorReset)
		}
	}

	for {
		// Display the prompt
		fmt.Printf("%sEnter choices [1-%d] (comma-separated, or press Enter for defaults)%s: ",
			m.numberColor, len(m.choices), ColorReset)

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading input: %w", err)
		}

		// Clean the input
		input = strings.TrimSpace(input)

		// Use default values if input is empty
		if input == "" && len(m.defaultValues) > 0 {
			return m.defaultValues, nil
		}

		// Parse the input
		parts := strings.Split(input, ",")
		var selectedValues []string
		valid := true

		for _, part := range parts {
			part = strings.TrimSpace(part)
			var choiceNum int
			_, err = fmt.Sscanf(part, "%d", &choiceNum)
			if err != nil || choiceNum < 1 || choiceNum > len(m.choices) {
				fmt.Printf("%s%s Invalid choice: %s. Please enter numbers between 1 and %d%s\n",
					m.errorColor, m.errorSymbol, part, len(m.choices), ColorReset)
				valid = false
				break
			}
			selectedValues = append(selectedValues, m.choices[choiceNum-1].Value)
		}

		if !valid {
			continue
		}

		return selectedValues, nil
	}
}
