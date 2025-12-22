package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Input represents a user input prompt
type Input struct {
	message      string
	defaultValue string
	validator    func(string) error
}

// New creates a new input prompt with a message
func New(message string) *Input {
	return &Input{
		message:      message,
		defaultValue: "",
		validator:    nil,
	}
}

// SetDefault sets a default value for the input
func (i *Input) SetDefault(value string) *Input {
	i.defaultValue = value
	return i
}

// SetValidator sets a validation function
// The validator should return an error if the input is invalid
func (i *Input) SetValidator(validator func(string) error) *Input {
	i.validator = validator
	return i
}

// Run displays the prompt and returns the user input
func (i *Input) Run() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display the prompt message
		if i.defaultValue != "" {
			fmt.Printf("%s [%s]: ", i.message, i.defaultValue)
		} else {
			fmt.Printf("%s: ", i.message)
		}

		// Read user input
		input, err := reader.ReadString('\n')
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
				fmt.Printf("✗ %s\n", err.Error())
				continue
			}
		}

		return input, nil
	}
}

// Confirm creates a yes/no confirmation prompt
type Confirm struct {
	message      string
	defaultValue bool
}

// NewConfirm creates a new confirmation prompt
func NewConfirm(message string) *Confirm {
	return &Confirm{
		message:      message,
		defaultValue: false,
	}
}

// SetDefault sets the default value for the confirmation
func (c *Confirm) SetDefault(value bool) *Confirm {
	c.defaultValue = value
	return c
}

// Run displays the confirmation prompt and returns the user's choice
func (c *Confirm) Run() (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display the prompt with default indicator
		defaultStr := "y/N"
		if c.defaultValue {
			defaultStr = "Y/n"
		}
		fmt.Printf("%s (%s): ", c.message, defaultStr)

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
			fmt.Println("✗ Please answer with 'y' or 'n'")
			continue
		}
	}
}

// Choice represents a choice in a select prompt
type Choice struct {
	Label string
	Value string
}

// Select represents a selection prompt from multiple choices
type Select struct {
	message      string
	choices      []Choice
	defaultValue string
}

// NewSelect creates a new select prompt
func NewSelect(message string, choices []Choice) *Select {
	return &Select{
		message:      message,
		choices:      choices,
		defaultValue: "",
	}
}

// SetDefault sets the default choice by value
func (s *Select) SetDefault(value string) *Select {
	s.defaultValue = value
	return s
}

// Run displays the selection prompt and returns the selected value
func (s *Select) Run() (string, error) {
	if len(s.choices) == 0 {
		return "", fmt.Errorf("no choices available")
	}

	reader := bufio.NewReader(os.Stdin)

	// Display choices
	fmt.Println(s.message)
	for i, choice := range s.choices {
		prefix := fmt.Sprintf("%d)", i+1)
		if choice.Value == s.defaultValue {
			fmt.Printf("  %s %s (default)\n", prefix, choice.Label)
		} else {
			fmt.Printf("  %s %s\n", prefix, choice.Label)
		}
	}

	for {
		// Display the prompt
		if s.defaultValue != "" {
			fmt.Printf("Enter choice [1-%d] (default: %s): ", len(s.choices), s.getDefaultLabel())
		} else {
			fmt.Printf("Enter choice [1-%d]: ", len(s.choices))
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
			fmt.Printf("✗ Please enter a number between 1 and %d\n", len(s.choices))
			continue
		}

		return s.choices[choiceNum-1].Value, nil
	}
}

// getDefaultLabel returns the label of the default choice
func (s *Select) getDefaultLabel() string {
	for i, choice := range s.choices {
		if choice.Value == s.defaultValue {
			return fmt.Sprintf("%d", i+1)
		}
	}
	return ""
}

// MultiChoice represents a multi-choice prompt
type MultiChoice struct {
	message       string
	choices       []Choice
	defaultValues []string
}

// NewMultiChoice creates a new multi-choice prompt
func NewMultiChoice(message string, choices []Choice) *MultiChoice {
	return &MultiChoice{
		message:       message,
		choices:       choices,
		defaultValues: []string{},
	}
}

// SetDefaults sets the default choices by values
func (m *MultiChoice) SetDefaults(values []string) *MultiChoice {
	m.defaultValues = values
	return m
}

// Run displays the multi-choice prompt and returns the selected values
func (m *MultiChoice) Run() ([]string, error) {
	if len(m.choices) == 0 {
		return nil, fmt.Errorf("no choices available")
	}

	reader := bufio.NewReader(os.Stdin)

	// Display choices
	fmt.Println(m.message)
	for i, choice := range m.choices {
		prefix := fmt.Sprintf("%d)", i+1)
		isDefault := false
		for _, dv := range m.defaultValues {
			if choice.Value == dv {
				isDefault = true
				break
			}
		}
		if isDefault {
			fmt.Printf("  %s %s (default)\n", prefix, choice.Label)
		} else {
			fmt.Printf("  %s %s\n", prefix, choice.Label)
		}
	}

	for {
		// Display the prompt
		fmt.Printf("Enter choices [1-%d] (comma-separated, or press Enter for defaults): ", len(m.choices))

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
				fmt.Printf("✗ Invalid choice: %s. Please enter numbers between 1 and %d\n", part, len(m.choices))
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
