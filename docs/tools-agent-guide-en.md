# Tools Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration](#3-agent-configuration)
4. [Model Configuration](#4-model-configuration)
5. [Defining Tools](#5-defining-tools)
6. [Detecting Tool Calls](#6-detecting-tool-calls)
7. [Tool Calls with Confirmation](#7-tool-calls-with-confirmation)
8. [Streaming Tool Calls](#8-streaming-tool-calls)
9. [Conversation History and Messages](#9-conversation-history-and-messages)
10. [Options: ToolAgentOption and ToolsAgentOption](#10-options-toolagentoption-and-toolsagentoption)
11. [Lifecycle Hooks (ToolsAgentOption)](#11-lifecycle-hooks-toolsagentoption)
12. [Tool Call State](#12-tool-call-state)
13. [Context and State Management](#13-context-and-state-management)
14. [JSON Export and Debugging](#14-json-export-and-debugging)
15. [API Reference](#15-api-reference)

---

## 1. Introduction

### What is a Tools Agent?

The `tools.Agent` is a specialized agent provided by the Nova SDK (`github.com/snipwise/nova`) that enables function calling (tool use) with LLMs. It sends messages to the LLM along with tool definitions, detects when the LLM wants to call a tool, executes the tool via a callback, and feeds the result back to the LLM.

### When to use a Tools Agent

| Scenario | Recommended agent |
|---|---|
| Function calling / tool use | `tools.Agent` |
| Multi-step tool execution loops | `tools.Agent` |
| Tool calls with user confirmation | `tools.Agent` |
| Free-form conversational AI | `chat.Agent` |
| Structured data extraction | `structured.Agent[T]` |
| Intent detection and routing | `orchestrator.Agent` |
| Context compression | `compressor.Agent` |
| Embedding generation and similarity search | `rag.Agent` |

### Key capabilities

- **Tool definition**: Define tools with a fluent builder API or use OpenAI/MCP tool formats directly.
- **Tool call detection loop**: Automatically detect and execute tool calls in a loop until the LLM stops.
- **Parallel tool calls**: Support for LLMs that can call multiple tools simultaneously.
- **Confirmation workflow**: Optionally require user confirmation before executing tools (Confirmed/Denied/Quit).
- **Streaming**: Stream the LLM's final response while processing tool calls.
- **Conversation history**: Optionally maintain conversation history across calls.
- **Lifecycle hooks**: Execute custom logic before and after each tool call detection.
- **MCP tool support**: Use MCP (Model Context Protocol) tools alongside native tools.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := tools.NewAgent(
        ctx,
        agents.Config{
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a helpful assistant.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.0),
        },
        tools.WithTools([]*tools.Tool{
            tools.NewTool("calculate_sum").
                SetDescription("Calculate the sum of two numbers").
                AddParameter("a", "number", "The first number", true).
                AddParameter("b", "number", "The second number", true),
        }),
    )
    if err != nil {
        panic(err)
    }

    result, err := agent.DetectToolCallsLoop(
        []messages.Message{
            {Role: roles.User, Content: "What is 40 + 2?"},
        },
        func(functionName string, arguments string) (string, error) {
            var args struct {
                A float64 `json:"a"`
                B float64 `json:"b"`
            }
            json.Unmarshal([]byte(arguments), &args)
            return fmt.Sprintf(`{"result": %g}`, args.A+args.B), nil
        },
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("Results:", result.Results)
    fmt.Println("Assistant:", result.LastAssistantMessage)
}
```

---

## 3. Agent Configuration

```go
agents.Config{
    Name:                    "tools-agent",
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
    APIKey:                  "your-api-key",
    SystemInstructions:      "You are a helpful assistant with tool access.",
    KeepConversationHistory: false,
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | No | Agent identifier for logging. |
| `EngineURL` | `string` | Yes | URL of the OpenAI-compatible LLM engine. |
| `APIKey` | `string` | No | API key for authenticated engines. |
| `SystemInstructions` | `string` | Recommended | System prompt. |
| `KeepConversationHistory` | `bool` | No | Keep messages across calls. Default: `false`. |

---

## 4. Model Configuration

```go
models.Config{
    Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
    Temperature:       models.Float64(0.0),
    ParallelToolCalls: models.Bool(false),
}
```

| Field | Description |
|---|---|
| `ParallelToolCalls` | Enable/disable parallel tool calls. Not all LLMs support this. |

---

## 5. Defining Tools

### Fluent builder API

```go
tool := tools.NewTool("calculate_sum").
    SetDescription("Calculate the sum of two numbers").
    AddParameter("a", "number", "The first number", true).
    AddParameter("b", "number", "The second number", true)
```

Supported parameter types: `"string"`, `"number"`, `"boolean"`, `"object"`, `"array"`.

### Passing tools to the agent

```go
agent, err := tools.NewAgent(ctx, agentConfig, modelConfig,
    tools.WithTools([]*tools.Tool{tool1, tool2}),
)
```

### Using OpenAI tools directly

```go
tools.WithOpenAITools([]openai.ChatCompletionToolUnionParam{...})
```

### Using MCP tools

```go
tools.WithMCPTools(mcpToolsList)
```

---

## 6. Detecting Tool Calls

### DetectToolCallsLoop

The primary method for tool call detection. Runs a loop: sends messages to the LLM, detects tool calls, executes them, feeds results back, until the LLM stops.

```go
result, err := agent.DetectToolCallsLoop(
    []messages.Message{
        {Role: roles.User, Content: "What is 40 + 2?"},
    },
    func(functionName string, arguments string) (string, error) {
        // Execute the tool and return the result as JSON
        return `{"result": 42}`, nil
    },
)
```

### DetectParallelToolCalls

For LLMs that support calling multiple tools at once:

```go
result, err := agent.DetectParallelToolCalls(userMessages, toolCallback)
```

### ToolCallResult

All detection methods return `*ToolCallResult`:

```go
type ToolCallResult struct {
    FinishReason         string
    Results              []string
    LastAssistantMessage string
}
```

---

## 7. Tool Calls with Confirmation

### DetectToolCallsLoopWithConfirmation

Adds a confirmation step before each tool execution:

```go
result, err := agent.DetectToolCallsLoopWithConfirmation(
    userMessages,
    toolCallback,
    func(functionName string, arguments string) tools.ConfirmationResponse {
        fmt.Printf("Execute %s(%s)? ", functionName, arguments)
        // Return tools.Confirmed, tools.Denied, or tools.Quit
        return tools.Confirmed
    },
)
```

### ConfirmationResponse values

| Value | Description |
|---|---|
| `tools.Confirmed` | Execute the tool call. |
| `tools.Denied` | Skip execution but continue the loop. |
| `tools.Quit` | Stop the entire agent execution. |

### DetectParallelToolCallsWithConfirmation

Same as parallel but with confirmation:

```go
result, err := agent.DetectParallelToolCallsWithConfirmation(
    userMessages, toolCallback, confirmationCallback,
)
```

---

## 8. Streaming Tool Calls

### DetectToolCallsLoopStream

Streams the LLM's final response while handling tool calls:

```go
result, err := agent.DetectToolCallsLoopStream(
    userMessages,
    toolCallback,
    func(chunk string) error {
        fmt.Print(chunk) // Print each chunk as it arrives
        return nil
    },
)
```

### DetectToolCallsLoopWithConfirmationStream

Combines streaming with confirmation:

```go
result, err := agent.DetectToolCallsLoopWithConfirmationStream(
    userMessages,
    toolCallback,
    confirmationCallback,
    streamCallback,
)
```

---

## 9. Conversation History and Messages

### Managing messages

```go
msgs := agent.GetMessages()
agent.AddMessage(roles.User, "A message")
agent.AddMessages([]messages.Message{...})
agent.ResetMessages()
agent.GetContextSize()
```

### Export to JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
```

---

## 10. Options: ToolAgentOption and ToolsAgentOption

The tools agent supports two distinct option types, both passed as variadic `...any` arguments to `NewAgent`:

### ToolAgentOption (OpenAI params level)

`ToolAgentOption` operates on `*openai.ChatCompletionNewParams` and configures the LLM request parameters (tools, etc.):

```go
tools.WithTools([]*tools.Tool{...})        // Set tools using fluent API
tools.WithOpenAITools(openaiTools)          // Set tools using OpenAI format
tools.WithMCPTools(mcpTools)                // Set tools using MCP format
```

### ToolsAgentOption (agent level)

`ToolsAgentOption` operates on the high-level `*Agent` and configures lifecycle hooks:

```go
tools.BeforeCompletion(func(a *tools.Agent) { ... })
tools.AfterCompletion(func(a *tools.Agent) { ... })
```

### Mixing both option types

```go
agent, err := tools.NewAgent(
    ctx, agentConfig, modelConfig,
    // ToolAgentOption
    tools.WithTools(myTools),
    // ToolsAgentOption
    tools.BeforeCompletion(func(a *tools.Agent) {
        fmt.Println("Before tool call detection...")
    }),
    tools.AfterCompletion(func(a *tools.Agent) {
        fmt.Println("After tool call detection...")
    }),
)
```

---

## 11. Lifecycle Hooks (ToolsAgentOption)

Lifecycle hooks allow you to execute custom logic before and after each tool call detection. They are triggered in all 6 detection methods.

### ToolsAgentOption

```go
type ToolsAgentOption func(*Agent)
```

### BeforeCompletion

Called before each tool call detection. The hook receives a reference to the agent.

```go
tools.BeforeCompletion(func(a *tools.Agent) {
    fmt.Printf("[BEFORE] Agent: %s, Messages: %d\n",
        a.GetName(), len(a.GetMessages()))
})
```

### AfterCompletion

Called after each tool call detection, once the result is ready. The hook receives a reference to the agent.

```go
tools.AfterCompletion(func(a *tools.Agent) {
    fmt.Printf("[AFTER] Agent: %s, Messages: %d\n",
        a.GetName(), len(a.GetMessages()))
})
```

### Hooks are triggered in all detection methods

| Method | Hooks triggered |
|---|---|
| `DetectParallelToolCalls` | Yes |
| `DetectParallelToolCallsWithConfirmation` | Yes |
| `DetectToolCallsLoop` | Yes |
| `DetectToolCallsLoopWithConfirmation` | Yes |
| `DetectToolCallsLoopStream` | Yes |
| `DetectToolCallsLoopWithConfirmationStream` | Yes |

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. Existing code without hooks continues to work without any changes.

---

## 12. Tool Call State

### GetLastStateToolCalls

Access the state of the last tool call execution:

```go
state := agent.GetLastStateToolCalls()
// state.Confirmation: Confirmed, Denied, or Quit
// state.ExecutionResult.Content: the tool result
// state.ExecutionResult.ExecFinishReason: "function_executed", "user_denied", "user_quit", "error", "exit_loop"
```

### ResetLastStateToolCalls

```go
agent.ResetLastStateToolCalls()
```

---

## 13. Context and State Management

### Getting and setting context

```go
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

### Getting and setting configuration

```go
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)
```

### Agent metadata

```go
agent.Kind()       // Returns agents.Tools
agent.GetName()    // Returns the agent name
agent.GetModelID() // Returns the model name
```

---

## 14. JSON Export and Debugging

```go
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

prettyReq, err := agent.GetLastRequestSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 15. API Reference

### Constructor

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    options ...any,
) (*Agent, error)
```

Creates a new tools agent. The `options` parameter accepts both `ToolAgentOption` (OpenAI params configuration) and `ToolsAgentOption` (agent-level hooks). The constructor separates them internally using type assertion.

---

### Types

```go
type ToolCallResult struct {
    FinishReason         string
    Results              []string
    LastAssistantMessage string
}

type ToolCallback func(functionName string, arguments string) (string, error)
type ConfirmationCallback func(functionName string, arguments string) ConfirmationResponse
type StreamCallback func(chunk string) error

type ConfirmationResponse int
const (
    Confirmed ConfirmationResponse = iota
    Denied
    Quit
)

// ToolAgentOption configures OpenAI params (tools, etc.)
type ToolAgentOption func(*openai.ChatCompletionNewParams)

// ToolsAgentOption configures the high-level Agent (lifecycle hooks)
type ToolsAgentOption func(*Agent)
```

---

### Option Functions

| Function | Type | Description |
|---|---|---|
| `WithTools(tools []*Tool)` | `ToolAgentOption` | Set tools using the fluent builder API. |
| `WithOpenAITools(tools []openai.ChatCompletionToolUnionParam)` | `ToolAgentOption` | Set tools using OpenAI format. |
| `WithMCPTools(tools []mcp.Tool)` | `ToolAgentOption` | Set tools using MCP format. |
| `BeforeCompletion(fn func(*Agent))` | `ToolsAgentOption` | Hook called before each tool call detection. |
| `AfterCompletion(fn func(*Agent))` | `ToolsAgentOption` | Hook called after each tool call detection. |

---

### Methods

| Method | Description |
|---|---|
| `DetectToolCallsLoop(msgs, callback) (*ToolCallResult, error)` | Detect and execute tool calls in a loop. |
| `DetectToolCallsLoopWithConfirmation(msgs, callback, confirm) (*ToolCallResult, error)` | Same with user confirmation. |
| `DetectToolCallsLoopStream(msgs, callback, stream) (*ToolCallResult, error)` | Same with streaming. |
| `DetectToolCallsLoopWithConfirmationStream(msgs, callback, confirm, stream) (*ToolCallResult, error)` | Same with confirmation and streaming. |
| `DetectParallelToolCalls(msgs, callback) (*ToolCallResult, error)` | Detect parallel tool calls. |
| `DetectParallelToolCallsWithConfirmation(msgs, callback, confirm) (*ToolCallResult, error)` | Same with confirmation. |
| `GetMessages() []messages.Message` | Get all conversation messages. |
| `AddMessage(role, content)` | Add a single message. |
| `AddMessages(msgs)` | Add multiple messages. |
| `ResetMessages()` | Clear messages except system instruction. |
| `GetContextSize() int` | Get approximate context size. |
| `ExportMessagesToJSON() (string, error)` | Export conversation as JSON. |
| `GetLastStateToolCalls() LastToolCallsState` | Get last tool call state. |
| `ResetLastStateToolCalls()` | Reset last tool call state. |
| `GetConfig() agents.Config` | Get agent configuration. |
| `SetConfig(config)` | Update agent configuration. |
| `GetModelConfig() models.Config` | Get model configuration. |
| `SetModelConfig(config)` | Update model configuration. |
| `GetContext() context.Context` | Get agent context. |
| `SetContext(ctx)` | Update agent context. |
| `GetLastRequestRawJSON() string` | Raw JSON of last request. |
| `GetLastResponseRawJSON() string` | Raw JSON of last response. |
| `GetLastRequestSON() (string, error)` | Pretty JSON of last request. |
| `GetLastResponseJSON() (string, error)` | Pretty JSON of last response. |
| `Kind() agents.Kind` | Returns `agents.Tools`. |
| `GetName() string` | Returns agent name. |
| `GetModelID() string` | Returns model name. |
