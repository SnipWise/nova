# Remote Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration](#3-agent-configuration)
4. [Server Health and Models](#4-server-health-and-models)
5. [Generating Completions](#5-generating-completions)
6. [Streaming Completions](#6-streaming-completions)
7. [Completions with Reasoning](#7-completions-with-reasoning)
8. [Conversation History and Messages](#8-conversation-history-and-messages)
9. [Tool Call Operations](#9-tool-call-operations)
10. [Lifecycle Hooks (RemoteAgentOption)](#10-lifecycle-hooks-remoteagentoption)
11. [Context Management](#11-context-management)
12. [JSON Export](#12-json-export)
13. [API Reference](#13-api-reference)

---

## 1. Introduction

### What is a Remote Agent?

The `remote.Agent` is a specialized agent provided by the Nova SDK (`github.com/snipwise/nova`) that communicates with a Nova server agent via HTTP. Instead of calling the LLM directly, it sends requests to a remote server that runs the actual LLM agent, and streams back the responses via Server-Sent Events (SSE).

This is useful for client-server architectures where the LLM runs on a dedicated server and multiple clients connect to it.

### When to use a Remote Agent

| Scenario | Recommended agent |
|---|---|
| Client-server architecture with shared LLM | `remote.Agent` |
| Web/mobile frontend connecting to LLM backend | `remote.Agent` |
| Tool calls with server-side validation | `remote.Agent` |
| Direct local LLM access | `chat.Agent`, `tools.Agent`, etc. |

### Key capabilities

- **HTTP-based communication**: Connects to a Nova server agent via REST/SSE.
- **Streaming support**: Receives responses as Server-Sent Events for real-time display.
- **Server-side tool calls**: Supports tool call detection with validation/cancellation.
- **Health checks**: Verify server availability before sending requests.
- **Model discovery**: Query which models the server is using.
- **Lifecycle hooks**: Execute custom logic before and after each completion.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/remote"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

func main() {
    ctx := context.Background()

    agent, err := remote.NewAgent(
        ctx,
        agents.Config{
            Name: "Remote Client",
        },
        "http://localhost:8080",
    )
    if err != nil {
        panic(err)
    }

    if !agent.IsHealthy() {
        fmt.Println("Server not available")
        return
    }

    result, err := agent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: "What is the capital of France?"},
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Response:", result.Response)
}
```

---

## 3. Agent Configuration

```go
agents.Config{
    Name: "Remote Client",    // Agent name (optional)
}
```

The remote agent requires a `baseURL` parameter (the server address) instead of an `EngineURL`. It does not need `SystemInstructions` or model configuration since those are managed server-side.

```go
agent, err := remote.NewAgent(
    ctx,
    agents.Config{Name: "My Client"},
    "http://localhost:8080",       // Server URL (required)
)
```

---

## 4. Server Health and Models

### Health check

```go
if agent.IsHealthy() {
    fmt.Println("Server is healthy")
}

// Or get detailed health status
health, err := agent.GetHealth()
fmt.Println(health.Status) // "ok"
```

### Model information

```go
modelsInfo, err := agent.GetModelsInfo()
fmt.Println("Chat model:", modelsInfo.ChatModel)
fmt.Println("Embeddings model:", modelsInfo.EmbeddingsModel)
fmt.Println("Tools model:", modelsInfo.ToolsModel)

// Or just the chat model
modelID := agent.GetModelID()
```

---

## 5. Generating Completions

### GenerateCompletion

Send messages and get the complete response:

```go
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "What is 2 + 2?"},
})
if err != nil {
    // handle error
}

fmt.Println(result.Response)     // "4"
fmt.Println(result.FinishReason) // "stop"
```

**Note:** This method uses streaming internally and collects the full response.

---

## 6. Streaming Completions

### GenerateStreamCompletion

Stream the response chunk by chunk:

```go
result, err := agent.GenerateStreamCompletion(
    []messages.Message{
        {Role: roles.User, Content: "Tell me a story."},
    },
    func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil
    },
)
```

### Stopping a stream

```go
agent.StopStream()
```

---

## 7. Completions with Reasoning

### GenerateCompletionWithReasoning

```go
result, err := agent.GenerateCompletionWithReasoning(userMessages)
fmt.Println(result.Response)
fmt.Println(result.Reasoning)   // Currently empty (server doesn't support it yet)
fmt.Println(result.FinishReason)
```

### GenerateStreamCompletionWithReasoning

```go
result, err := agent.GenerateStreamCompletionWithReasoning(
    userMessages,
    reasoningCallback,  // Currently not used
    responseCallback,
)
```

**Note:** Reasoning is not yet supported by the server API. These methods delegate to the standard completion methods.

---

## 8. Conversation History and Messages

### Get messages from server

```go
msgs := agent.GetMessages()
for _, msg := range msgs {
    fmt.Printf("[%s] %s\n", msg.Role, msg.Content)
}
```

### Get context size

```go
tokens := agent.GetContextSize()
fmt.Printf("Context size: %d tokens\n", tokens)
```

### Reset conversation

```go
agent.ResetMessages()
```

### Export to JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
```

**Note:** Messages are managed server-side. `AddMessage` and `AddMessages` are no-ops for the remote agent.

---

## 9. Tool Call Operations

When the server detects tool calls, it sends notifications via SSE. You can set a callback to handle them and validate/cancel operations.

### Set tool call callback

```go
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("Tool call: %s (operation: %s)\n", message, operationID)
    return nil
})
```

### Validate an operation

```go
err := agent.ValidateOperation(operationID)
```

### Cancel an operation

```go
err := agent.CancelOperation(operationID)
```

### Reset all operations

```go
err := agent.ResetOperations()
```

---

## 10. Lifecycle Hooks (RemoteAgentOption)

Lifecycle hooks allow you to execute custom logic before and after each completion. They are configured as functional options when creating the agent.

### RemoteAgentOption

```go
type RemoteAgentOption func(*Agent)
```

Options are passed as variadic arguments to `NewAgent`:

```go
agent, err := remote.NewAgent(ctx, agentConfig, baseURL,
    remote.BeforeCompletion(fn),
    remote.AfterCompletion(fn),
)
```

### BeforeCompletion

Called before each completion. The hook receives a reference to the agent.

```go
remote.BeforeCompletion(func(a *remote.Agent) {
    fmt.Printf("[BEFORE] Agent: %s\n", a.GetName())
})
```

### AfterCompletion

Called after each completion. The hook receives a reference to the agent.

```go
remote.AfterCompletion(func(a *remote.Agent) {
    fmt.Printf("[AFTER] Agent: %s\n", a.GetName())
})
```

### Hook placement

The hooks are in `GenerateStreamCompletion`, which is the base method that all other completion methods delegate to:

| Method | Hooks triggered |
|---|---|
| `GenerateStreamCompletion` | Yes (directly) |
| `GenerateCompletion` | Yes (via `GenerateStreamCompletion`) |
| `GenerateCompletionWithReasoning` | Yes (via `GenerateCompletion` -> `GenerateStreamCompletion`) |
| `GenerateStreamCompletionWithReasoning` | Yes (via `GenerateStreamCompletion`) |

This ensures exactly one before/after hook per call, regardless of which method is used.

### Complete example

```go
callCount := 0

agent, err := remote.NewAgent(
    ctx,
    agents.Config{Name: "Remote Client"},
    "http://localhost:8080",
    remote.BeforeCompletion(func(a *remote.Agent) {
        callCount++
        fmt.Printf("[BEFORE] Call #%d\n", callCount)
    }),
    remote.AfterCompletion(func(a *remote.Agent) {
        fmt.Printf("[AFTER] Call #%d\n", callCount)
    }),
)
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. Existing code without hooks continues to work without any changes.

---

## 11. Context Management

### Getting and setting context

```go
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

### Agent metadata

```go
agent.Kind()       // Returns agents.Remote
agent.GetName()    // Returns the agent name
agent.GetModelID() // Returns the chat model from server
```

---

## 12. JSON Export

```go
jsonStr, err := agent.ExportMessagesToJSON()
```

---

## 13. API Reference

### Constructor

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    baseURL string,
    opts ...RemoteAgentOption,
) (*Agent, error)
```

Creates a new remote agent. The `baseURL` is the server address. The `opts` parameter accepts zero or more `RemoteAgentOption` functional options.

---

### Types

```go
type CompletionResult struct {
    Response     string
    FinishReason string
}

type ReasoningResult struct {
    Response     string
    Reasoning    string
    FinishReason string
}

type StreamCallback func(chunk string, finishReason string) error
type ToolCallCallback func(operationID string, message string) error

type ModelsInfo struct {
    Status          string
    ChatModel       string
    EmbeddingsModel string
    ToolsModel      string
}

type HealthStatus struct {
    Status string
}

type RemoteAgentOption func(*Agent)
```

---

### Option Functions

| Function | Type | Description |
|---|---|---|
| `BeforeCompletion(fn func(*Agent))` | `RemoteAgentOption` | Sets a hook called before each completion. |
| `AfterCompletion(fn func(*Agent))` | `RemoteAgentOption` | Sets a hook called after each completion. |

---

### Methods

| Method | Description |
|---|---|
| `GenerateCompletion(msgs) (*CompletionResult, error)` | Send messages and get complete response. |
| `GenerateStreamCompletion(msgs, callback) (*CompletionResult, error)` | Stream the response via callback. |
| `GenerateCompletionWithReasoning(msgs) (*ReasoningResult, error)` | Completion with reasoning (not yet supported server-side). |
| `GenerateStreamCompletionWithReasoning(msgs, reasoningCb, responseCb) (*ReasoningResult, error)` | Streaming with reasoning. |
| `StopStream()` | Stop the current streaming operation. |
| `SetToolCallCallback(callback)` | Set callback for tool call notifications. |
| `ValidateOperation(operationID) error` | Validate a pending tool call. |
| `CancelOperation(operationID) error` | Cancel a pending tool call. |
| `ResetOperations() error` | Cancel all pending operations. |
| `GetMessages() []messages.Message` | Get conversation messages from server. |
| `GetContextSize() int` | Get context size in tokens from server. |
| `ResetMessages()` | Reset conversation on server. |
| `ExportMessagesToJSON() (string, error)` | Export conversation as JSON. |
| `IsHealthy() bool` | Check if server is healthy. |
| `GetHealth() (*HealthStatus, error)` | Get detailed health status. |
| `GetModelsInfo() (*ModelsInfo, error)` | Get server model information. |
| `GetContext() context.Context` | Get agent context. |
| `SetContext(ctx)` | Update agent context. |
| `Kind() agents.Kind` | Returns `agents.Remote`. |
| `GetName() string` | Returns agent name. |
| `GetModelID() string` | Returns chat model from server. |
