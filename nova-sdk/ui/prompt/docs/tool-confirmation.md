# tool.confirmation.go - Tool Execution Confirmation

This file provides a standardized confirmation prompt for AI agent tool execution.

## Overview

The `tool.confirmation.go` file implements a human-in-the-loop confirmation system for AI agents. It allows users to approve, deny, or quit when an AI agent wants to execute a tool/function.

This is commonly used in agent systems where you want human oversight before potentially dangerous or important operations are performed.

---

## Function

### HumanConfirmation(text string) tools.ConfirmationResponse

Prompts the user to confirm a tool execution with three options: yes, no, or quit.

#### Parameters

- `text` (string): The confirmation message to display to the user. Typically describes what the tool will do.

#### Returns

- `tools.ConfirmationResponse`: One of three values:
  - `tools.Confirmed`: User approved the action
  - `tools.Denied`: User rejected the action
  - `tools.Quit`: User wants to quit the entire operation

#### Signature

```go
func HumanConfirmation(text string) tools.ConfirmationResponse
```

---

## Confirmation Response Types

The function returns values from the `tools` package:

```go
// From github.com/snipwise/nova/nova-sdk/agents/tools
type ConfirmationResponse int

const (
    Confirmed ConfirmationResponse = iota
    Denied
    Quit
)
```

---

## Usage Examples

### Basic Tool Confirmation

```go
import (
    "fmt"
    "github.com/snipwise/nova/nova-sdk/ui/prompt"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
)

response := prompt.HumanConfirmation("Allow agent to delete file 'config.json'?")

switch response {
case tools.Confirmed:
    fmt.Println("User approved: Deleting file...")
    // Execute the tool
case tools.Denied:
    fmt.Println("User denied: Operation cancelled")
    // Skip this tool
case tools.Quit:
    fmt.Println("User quit: Stopping agent")
    // Stop the entire agent loop
}
```

### In Agent Tool Loop

Example usage in an agent loop:

```go
// Example: Using HumanConfirmation in a tool execution loop
for _, tool := range toolsToExecute {
    message := fmt.Sprintf("Execute tool '%s' with args: %v?",
        tool.Name, tool.Args)

    response := prompt.HumanConfirmation(message)

    switch response {
    case tools.Confirmed:
        result := executeTool(tool)
        fmt.Printf("Tool executed: %v\n", result)

    case tools.Denied:
        fmt.Printf("Skipped tool: %s\n", tool.Name)
        continue

    case tools.Quit:
        fmt.Println("Stopping agent execution")
        return
    }
}
```

### File Operations

Example for file deletion confirmation:

```go
// Example function (not part of the package)
func deleteFile(path string) error {
    message := fmt.Sprintf("🗑️  Delete file '%s'?", path)
    response := prompt.HumanConfirmation(message)

    if response == tools.Confirmed {
        return os.Remove(path)
    } else if response == tools.Quit {
        return fmt.Errorf("operation cancelled by user")
    }

    return nil // Denied - skip silently
}
```

### Database Operations

Example for database query confirmation:

```go
// Example function (not part of the package)
func executeQuery(query string) error {
    message := fmt.Sprintf("💾 Execute SQL query?\n%s", query)
    response := prompt.HumanConfirmation(message)

    switch response {
    case tools.Confirmed:
        return db.Exec(query)
    case tools.Denied:
        log.Println("Query execution denied")
        return nil
    case tools.Quit:
        return fmt.Errorf("database operations halted")
    }

    return nil
}
```

### API Calls

Example for API call confirmation:

```go
// Example function (not part of the package)
func callExternalAPI(endpoint string, data interface{}) error {
    message := fmt.Sprintf("🌐 Make API call to %s?", endpoint)
    response := prompt.HumanConfirmation(message)

    if response != tools.Confirmed {
        if response == tools.Quit {
            return fmt.Errorf("API operations stopped")
        }
        return nil // Denied
    }

    // Make the API call
    return makeHTTPRequest(endpoint, data)
}
```

---

## Visual Appearance

The confirmation prompt appears with:

**Message**: Bright cyan color
**Choices**: White text with gray key indicators
**Default**: Highlighted in bright yellow with ● symbol
**Prompt symbol**: ❯
**Error symbol**: ✗

Example output:
```
❯ Allow agent to delete file 'config.json'?
  y) yes ●
  n) no
  q) quit
Enter choice [y/n/q] (default: y):
```

---

## Implementation Details

The function uses:
- `ColorSelectKey`: Keyboard shortcut selection prompt
- **Styled with custom colors**: Cyan message, white choices, yellow default
- **Single-key selection**: Press `y`, `n`, or `q` (no Enter needed)
- **Default to "yes"**: Makes it quick to approve multiple actions
- **Fatal error on prompt failure**: Uses `log.Fatal()` if prompt fails

### Default Behavior

- Default choice: **Yes (y)**
- Pressing Enter without input: Confirms (yes)
- Invalid input: Re-prompts

---

## Integration with AI Agents

This function is designed to work with the Nova SDK agent system:

```go
// Example integration (pseudocode - adapt to your agent implementation)
import (
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/ui/prompt"
)

// Configure agent with human confirmation
agent := agents.NewAgent(config)

// Set confirmation callback
agent.SetToolConfirmation(prompt.HumanConfirmation)

// Agent will now ask for confirmation before each tool execution
agent.Run()
```

---

## Customization

If you need different behavior, you can create your own confirmation function:

```go
// Example custom confirmation function (not part of the package)
func MyCustomConfirmation(text string) tools.ConfirmationResponse {
    choices := []prompt.Choice{
        {Label: "Approve", Value: "a"},
        {Label: "Reject", Value: "r"},
        {Label: "Stop All", Value: "s"},
    }

    selectPrompt := prompt.NewColorSelectKey(text, choices).
        SetDefault("r"). // Default to reject for safety
        SetMessageColor(prompt.ColorBrightRed)

    selected, err := selectPrompt.Run()
    if err != nil {
        return tools.Denied
    }

    switch selected {
    case "a":
        return tools.Confirmed
    case "r":
        return tools.Denied
    case "s":
        return tools.Quit
    }

    return tools.Denied
}
```

---

## Best Practices

1. **Clear messages**: Describe exactly what the tool will do
   ```go
   // Good
   HumanConfirmation("Delete 150 files from /tmp?")

   // Bad
   HumanConfirmation("Proceed?")
   ```

2. **Include relevant details**: Help users make informed decisions
   ```go
   message := fmt.Sprintf(
       "Send email to %s with subject '%s'?",
       recipient, subject,
   )
   HumanConfirmation(message)
   ```

3. **Use emojis for context**: Makes messages more scannable
   ```go
   HumanConfirmation("🗑️  Delete database table 'users'?")
   HumanConfirmation("📧 Send notification to 500 users?")
   HumanConfirmation("💰 Process payment of $150.00?")
   ```

4. **Handle Quit properly**: Stop the entire operation, not just skip
   ```go
   if response == tools.Quit {
       return fmt.Errorf("operation cancelled by user")
   }
   ```

5. **Log denials**: Keep audit trail
   ```go
   if response == tools.Denied {
       log.Printf("User denied: %s", toolName)
   }
   ```

---

## Security Considerations

This confirmation system is crucial for:

- **Preventing unintended actions**: AI agents might hallucinate dangerous operations
- **Audit trail**: User explicitly approves each action
- **Human oversight**: Critical operations require human judgment
- **Safety net**: Users can quit entire agent loop if things go wrong

**Important**: This is not a security boundary. The agent code must still validate and sanitize all inputs.

---

## Error Handling

The function uses `log.Fatal()` on prompt errors. In production, you may want to handle this differently:

```go
// Example wrapper with custom error handling (not part of the package)
func SafeHumanConfirmation(text string) (tools.ConfirmationResponse, error) {
    choices := []prompt.Choice{
        {Label: "yes", Value: "y"},
        {Label: "no", Value: "n"},
        {Label: "quit", Value: "q"},
    }

    selectPrompt := prompt.NewColorSelectKey(text, choices).
        SetDefault("y")

    selected, err := selectPrompt.Run()
    if err != nil {
        return tools.Denied, err
    }

    switch selected {
    case "q":
        return tools.Quit, nil
    case "n":
        return tools.Denied, nil
    case "y":
        return tools.Confirmed, nil
    default:
        return tools.Denied, nil
    }
}
```

---

## Related

- See `color-prompt.md` for `ColorSelectKey` documentation
- See Nova SDK agent documentation for tool execution flow
- See `tools` package for `ConfirmationResponse` type definition
