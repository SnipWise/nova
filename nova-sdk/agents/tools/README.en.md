# Tools Agent

## Description

The **Tools Agent** is a specialized agent for detecting and executing function calls (function calling). It allows an LLM to decide when and how to use external tools/functions to accomplish tasks.

## Features

- **Tool call detection** : Identifies when the LLM wants to call a function
- **Function execution** : Executes functions via callbacks
- **Parallel calls** : Support for parallel function calls (if the model supports it)
- **Execution loop** : Automatically executes multiple successive calls
- **User confirmation** : Human-in-the-loop to validate function calls
- **Streaming** : Support for streaming during detection and execution
- **MCP support** : Integration with MCP (Model Context Protocol) tools

## Use cases

The Tools Agent is used for:
- **Function calls** : Calculator, weather, external APIs, databases
- **Actions** : Send emails, create files, make HTTP requests
- **Human-in-the-loop** : Request confirmation before execution
- **Automation** : Chain multiple function calls automatically

## Creating a Tools Agent

### Basic syntax

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

// Agent configuration
agentConfig := agents.Config{
    Name: "ToolsAgent",
    Instructions: "You are a helpful assistant with access to tools.",
}

// Model configuration (must support function calling)
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    Name:      "hf.co/menlo/jan-nano-gguf:q4_k_m", // Supports function calling
}

// Define available tools
myTools := []*tools.Tool{
    tools.NewTool("calculate").
        SetDescription("Perform a mathematical calculation").
        AddParameter("expression", "string", "The mathematical expression to evaluate", true),

    tools.NewTool("get_weather").
        SetDescription("Get current weather for a location").
        AddParameter("location", "string", "City name", true).
        AddParameter("unit", "string", "Temperature unit (celsius/fahrenheit)", false),
}

// Create the agent with tools
agent, err := tools.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    tools.WithTools(myTools),
)
if err != nil {
    log.Fatal(err)
}
```

### Creation options

| Option | Description |
|--------|-------------|
| `WithTools(tools)` | Defines Nova SDK tools |
| `WithOpenAITools(tools)` | Defines tools in raw OpenAI format |
| `WithMCPTools(tools)` | Defines MCP (Model Context Protocol) tools |

## Tool definition

### Fluent API

```go
// Create a tool with the fluent API
calculateTool := tools.NewTool("calculate").
    SetDescription("Perform mathematical calculations").
    AddParameter("expression", "string", "Expression to evaluate (e.g., '2 + 2')", true).
    AddParameter("precision", "number", "Number of decimal places", false)

emailTool := tools.NewTool("send_email").
    SetDescription("Send an email to a recipient").
    AddParameter("to", "string", "Recipient email address", true).
    AddParameter("subject", "string", "Email subject", true).
    AddParameter("body", "string", "Email body content", true)
```

### Tool structure

```go
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]Parameter
    Required    []string
}

type Parameter struct {
    Type        string // "string", "number", "boolean", "object", "array"
    Description string
}
```

### Parameter types

- `"string"` : Text
- `"number"` : Number (integer or decimal)
- `"boolean"` : Boolean (true/false)
- `"object"` : JSON object
- `"array"` : Array

## Main methods

### DetectToolCallsLoop - Execution loop (recommended)

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Callback to execute functions
executeFunction := func(functionName string, arguments string) (string, error) {
    switch functionName {
    case "calculate":
        // Parse arguments and execute
        return `{"result": 4}`, nil
    case "get_weather":
        return `{"temperature": 22, "condition": "sunny"}`, nil
    default:
        return "", fmt.Errorf("unknown function: %s", functionName)
    }
}

// Detect and execute tool calls in a loop
userMessages := []messages.Message{
    {Role: roles.User, Content: "What's the weather in Paris and how much is 2 + 2?"},
}

result, err := agent.DetectToolCallsLoop(userMessages, executeFunction)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Finish reason:", result.FinishReason)         // "function_executed" or "stop"
fmt.Println("Tool results:", result.Results)               // ["{"result": 4}", "{"temperature": 22, ...}"]
fmt.Println("Assistant message:", result.LastAssistantMessage)
```

### DetectToolCallsLoopWithConfirmation - With user confirmation

```go
// Confirmation callback (human-in-the-loop)
confirmationPrompt := func(functionName string, arguments string) tools.ConfirmationResponse {
    fmt.Printf("Execute %s with args %s? (y/n/q): ", functionName, arguments)
    var response string
    fmt.Scanln(&response)

    switch response {
    case "y":
        return tools.Confirmed
    case "n":
        return tools.Denied
    case "q":
        return tools.Quit
    default:
        return tools.Denied
    }
}

// Detect and execute with confirmation
result, err := agent.DetectToolCallsLoopWithConfirmation(
    userMessages,
    executeFunction,
    confirmationPrompt,
)
```

**Confirmation responses**:
- `tools.Confirmed` : Execute the function
- `tools.Denied` : Don't execute, but continue
- `tools.Quit` : Stop entire execution

### DetectParallelToolCalls - Parallel calls

**Note** : Not all LLMs support parallel calls.

```go
// Execute multiple tool calls in parallel
result, err := agent.DetectParallelToolCalls(userMessages, executeFunction)

// With confirmation
result, err := agent.DetectParallelToolCallsWithConfirmation(
    userMessages,
    executeFunction,
    confirmationPrompt,
)
```

### Streaming

```go
// Streaming during tool call detection
streamCallback := func(chunk string) error {
    fmt.Print(chunk)
    return nil
}

result, err := agent.DetectToolCallsLoopStream(
    userMessages,
    executeFunction,
    streamCallback,
)

// With confirmation
result, err := agent.DetectToolCallsLoopWithConfirmationStream(
    userMessages,
    executeFunction,
    confirmationPrompt,
    streamCallback,
)
```

### Message management

```go
// Add a message
agent.AddMessage(roles.User, "Question...")

// Add multiple messages
messages := []messages.Message{
    {Role: roles.User, Content: "Question 1"},
    {Role: roles.Assistant, Content: "Answer 1"},
}
agent.AddMessages(messages)

// Get all messages
allMessages := agent.GetMessages()

// Reset messages
agent.ResetMessages()

// Export to JSON
jsonData, err := agent.ExportMessagesToJSON()

// Context size
contextSize := agent.GetContextSize()
```

### Tool call state

```go
// Get the last function call state
state := agent.GetLastStateToolCalls()

// State contains:
// - Confirmation: tools.ConfirmationResponse
// - ExecutionResult.Content: Execution result
// - ExecutionResult.ExecFinishReason: "function_executed", "user_denied", "user_quit", etc.
// - ExecutionResult.ShouldStop: Whether execution should stop

fmt.Println("Confirmation:", state.Confirmation)
fmt.Println("Finish reason:", state.ExecutionResult.ExecFinishReason)

// Reset state
agent.ResetLastStateToolCalls()
```

### Getters and Setters

```go
// Configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)

// Information
name := agent.GetName()
modelID := agent.GetModelID()
kind := agent.Kind() // Returns agents.Tools

// Context
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requests/Responses (debugging)
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
prettyRequest, _ := agent.GetLastRequestSON()
prettyResponse, _ := agent.GetLastResponseJSON()
```

## ToolCallResult structure

```go
type ToolCallResult struct {
    FinishReason         string   // "function_executed", "stop", "user_denied", "user_quit"
    Results              []string // JSON results of executed functions
    LastAssistantMessage string   // Last assistant message
}
```

## Usage with other agents

The Tools Agent is typically used with Server or Crew agents:

```go
// Create the tools agent
toolsAgent, _ := tools.NewAgent(ctx, agentConfig, modelConfig, tools.WithTools(myTools))

// Execution function
executeFn := func(functionName string, arguments string) (string, error) {
    // Implementation...
    return result, nil
}

// Use with Server Agent
serverAgent, _ := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithToolsAgent(toolsAgent),
    server.WithExecuteFn(executeFn),
)

// Use with Crew Agent
crewAgent, _ := crew.NewAgent(
    ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithToolsAgent(toolsAgent),
    crew.WithExecuteFn(executeFn),
    crew.WithConfirmationPromptFn(confirmationPrompt),
)
```

## Complete example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Configuration
    agentConfig := agents.Config{
        Name:         "Calculator",
        Instructions: "You are a helpful calculator assistant.",
    }
    modelConfig := models.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        Name:      "hf.co/menlo/jan-nano-gguf:q4_k_m",
    }

    // Define tools
    myTools := []*tools.Tool{
        tools.NewTool("add").
            SetDescription("Add two numbers").
            AddParameter("a", "number", "First number", true).
            AddParameter("b", "number", "Second number", true),

        tools.NewTool("multiply").
            SetDescription("Multiply two numbers").
            AddParameter("a", "number", "First number", true).
            AddParameter("b", "number", "Second number", true),
    }

    // Create the agent
    agent, err := tools.NewAgent(ctx, agentConfig, modelConfig, tools.WithTools(myTools))
    if err != nil {
        log.Fatal(err)
    }

    // Execution function
    executeFunction := func(functionName string, arguments string) (string, error) {
        var args map[string]float64
        if err := json.Unmarshal([]byte(arguments), &args); err != nil {
            return "", err
        }

        var result float64
        switch functionName {
        case "add":
            result = args["a"] + args["b"]
        case "multiply":
            result = args["a"] * args["b"]
        default:
            return "", fmt.Errorf("unknown function: %s", functionName)
        }

        return fmt.Sprintf(`{"result": %f}`, result), nil
    }

    // Detect and execute
    userMessages := []messages.Message{
        {Role: roles.User, Content: "How much is 5 + 3?"},
    }

    result, err := agent.DetectToolCallsLoop(userMessages, executeFunction)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Finish reason: %s\n", result.FinishReason)
    fmt.Printf("Results: %v\n", result.Results)
    fmt.Printf("Assistant: %s\n", result.LastAssistantMessage)
}
```

## MCP (Model Context Protocol) Support

The Tools Agent supports MCP tools:

```go
import "github.com/mark3labs/mcp-go/mcp"

// MCP tools
mcpTools := []mcp.Tool{
    // Your MCP tools
}

// Create agent with MCP tools
agent, err := tools.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    tools.WithMCPTools(mcpTools),
)
```

## Notes

- **Kind** : Returns `agents.Tools`
- **Required models** : Only models supporting function calling work
- **Parallel calls** : Not all LLMs support this feature
- **Automatic loop** : `DetectToolCallsLoop` executes multiple successive calls
- **Confirmation** : `WithConfirmation` adds human-in-the-loop
- **Streaming** : Compatible with response streaming
- **Persistent state** : `GetLastStateToolCalls()` allows maintaining state between invocations

## Recommendations

### Recommended models for function calling

- **hf.co/menlo/jan-nano-gguf:q4_k_m** : Small, fast, good support
- **qwen2.5:1.5b** : Size/performance balance
- **Avoid** : Models without native function calling support

### Best practices

1. **Clear descriptions** : Precisely describe what each tool does
2. **Explicit parameters** : Indicate types and constraints
3. **Error handling** : Return clear JSON errors
4. **Confirmation** : Use `WithConfirmation` for sensitive actions
5. **Validation** : Validate arguments before execution

```go
executeFunction := func(functionName string, arguments string) (string, error) {
    // Parse
    var args map[string]any
    if err := json.Unmarshal([]byte(arguments), &args); err != nil {
        return `{"error": "Invalid JSON arguments"}`, err
    }

    // Validate
    if functionName == "delete_file" {
        if args["path"] == "" {
            return `{"error": "path is required"}`, fmt.Errorf("missing path")
        }
    }

    // Execute
    // ...

    return `{"success": true}`, nil
}
```
