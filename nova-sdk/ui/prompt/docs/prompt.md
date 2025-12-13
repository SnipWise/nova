# prompt.go - Basic Prompts

This file provides basic, non-colored prompt functionality for user input.

## Overview

The `prompt.go` file implements simple terminal prompts for user input, confirmations, selections, and multi-choice selections. These are the basic building blocks for CLI interactions.

## Types

### Input

Represents a user input prompt.

```go
type Input struct {
    message      string
    defaultValue string
    validator    func(string) error
}
```

#### Methods

##### New(message string) *Input

Creates a new input prompt with a message.

```go
input := prompt.New("Enter your name")
```

##### SetDefault(value string) *Input

Sets a default value for the input.

```go
input := prompt.New("Enter your name").
    SetDefault("John Doe")
```

##### SetValidator(validator func(string) error) *Input

Sets a validation function. The validator should return an error if the input is invalid.

```go
input := prompt.New("Enter age").
    SetValidator(func(s string) error {
        age, err := strconv.Atoi(s)
        if err != nil || age < 0 {
            return fmt.Errorf("please enter a valid age")
        }
        return nil
    })
```

##### Run() (string, error)

Displays the prompt and returns the user input.

```go
name, err := input.Run()
if err != nil {
    log.Fatal(err)
}
```

**Features:**
- Displays the prompt message
- Shows default value in brackets if set
- Validates input if validator is set
- Re-prompts on validation failure
- Returns cleaned (trimmed) input

---

### Confirm

Represents a yes/no confirmation prompt.

```go
type Confirm struct {
    message      string
    defaultValue bool
}
```

#### Methods

##### NewConfirm(message string) *Confirm

Creates a new confirmation prompt.

```go
confirm := prompt.NewConfirm("Do you want to continue?")
```

##### SetDefault(value bool) *Confirm

Sets the default value for the confirmation.

```go
confirm := prompt.NewConfirm("Delete file?").
    SetDefault(false)
```

##### Run() (bool, error)

Displays the confirmation prompt and returns the user's choice.

```go
yes, err := confirm.Run()
if err != nil {
    log.Fatal(err)
}
if yes {
    fmt.Println("Proceeding...")
}
```

**Accepted responses:**
- Yes: `y`, `yes`, `o`, `oui`
- No: `n`, `no`, `non`
- Empty: uses default value

---

### Select

Represents a selection prompt from multiple choices.

```go
type Select struct {
    message      string
    choices      []Choice
    defaultValue string
}
```

#### Choice

```go
type Choice struct {
    Label string  // Display text
    Value string  // Return value
}
```

#### Methods

##### NewSelect(message string, choices []Choice) *Select

Creates a new select prompt.

```go
choices := []prompt.Choice{
    {Label: "Small", Value: "s"},
    {Label: "Medium", Value: "m"},
    {Label: "Large", Value: "l"},
}
select := prompt.NewSelect("Choose size", choices)
```

##### SetDefault(value string) *Select

Sets the default choice by value.

```go
select := prompt.NewSelect("Choose size", choices).
    SetDefault("m")
```

##### Run() (string, error)

Displays the selection prompt and returns the selected value.

```go
size, err := select.Run()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Selected size: %s\n", size)
```

**Features:**
- Displays numbered choices
- Shows which choice is default
- Accepts numeric input (1-N)
- Validates choice range

---

### MultiChoice

Represents a multi-choice prompt.

```go
type MultiChoice struct {
    message       string
    choices       []Choice
    defaultValues []string
}
```

#### Methods

##### NewMultiChoice(message string, choices []Choice) *MultiChoice

Creates a new multi-choice prompt.

```go
choices := []prompt.Choice{
    {Label: "JavaScript", Value: "js"},
    {Label: "Python", Value: "py"},
    {Label: "Go", Value: "go"},
}
multi := prompt.NewMultiChoice("Select languages", choices)
```

##### SetDefaults(values []string) *MultiChoice

Sets the default choices by values.

```go
multi := prompt.NewMultiChoice("Select languages", choices).
    SetDefaults([]string{"js", "go"})
```

##### Run() ([]string, error)

Displays the multi-choice prompt and returns the selected values.

```go
languages, err := multi.Run()
if err != nil {
    log.Fatal(err)
}
for _, lang := range languages {
    fmt.Printf("Selected: %s\n", lang)
}
```

**Features:**
- Displays numbered choices
- Shows which choices are default
- Accepts comma-separated input (e.g., "1,3,5")
- Validates all choices

---

## Usage Examples

### Basic Input

```go
input := prompt.New("What is your name?")
name, err := input.Run()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Hello, %s!\n", name)
```

### Input with Default and Validation

```go
input := prompt.New("Enter port number").
    SetDefault("8080").
    SetValidator(func(s string) error {
        port, err := strconv.Atoi(s)
        if err != nil || port < 1 || port > 65535 {
            return fmt.Errorf("invalid port number")
        }
        return nil
    })

port, err := input.Run()
if err != nil {
    log.Fatal(err)
}
```

### Confirmation

```go
confirm := prompt.NewConfirm("Delete all files?").
    SetDefault(false)

if yes, _ := confirm.Run(); yes {
    // Delete files
    fmt.Println("Files deleted")
} else {
    fmt.Println("Operation cancelled")
}
```

### Selection

```go
choices := []prompt.Choice{
    {Label: "Development", Value: "dev"},
    {Label: "Staging", Value: "stage"},
    {Label: "Production", Value: "prod"},
}

select := prompt.NewSelect("Choose environment", choices).
    SetDefault("dev")

env, err := select.Run()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Deploying to: %s\n", env)
```

### Multi-Choice

```go
choices := []prompt.Choice{
    {Label: "Unit Tests", Value: "unit"},
    {Label: "Integration Tests", Value: "integration"},
    {Label: "E2E Tests", Value: "e2e"},
}

multi := prompt.NewMultiChoice("Select tests to run", choices).
    SetDefaults([]string{"unit"})

tests, err := multi.Run()
if err != nil {
    log.Fatal(err)
}

for _, test := range tests {
    fmt.Printf("Running %s tests...\n", test)
}
```

---

## Notes

- All prompts use `bufio.Reader` for input reading
- Input is automatically trimmed of whitespace
- Empty input uses default value if available
- Validation errors are displayed with an ✗ symbol
- These are basic prompts without color support (see `color.prompt.go` for colored versions)
