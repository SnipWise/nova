# Chat Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration](#3-agent-configuration)
4. [Model Configuration](#4-model-configuration)
5. [Completion Methods](#5-completion-methods)
6. [Streaming Completion](#6-streaming-completion)
7. [Reasoning Support](#7-reasoning-support)
8. [Conversation History](#8-conversation-history)
9. [User Message Directives](#9-user-message-directives)
10. [Lifecycle Hooks (ChatAgentOption)](#10-lifecycle-hooks-chatagentoption)
11. [Context and State Management](#11-context-and-state-management)
12. [JSON Export and Debugging](#12-json-export-and-debugging)
13. [API Reference](#13-api-reference)

---

## 1. Introduction

### What is a Chat Agent?

The `chat.Agent` is the core conversational agent provided by the Nova SDK (`github.com/snipwise/nova`). It wraps the OpenAI-compatible API behind a simplified Go interface, handling message formatting, conversation history, streaming, and model configuration transparently.

A `chat.Agent` is designed for direct programmatic use -- you call methods in your Go code and receive responses directly, without any HTTP layer. For HTTP-based usage, see the `ServerAgent` or `CrewServerAgent` guides.

### When to use a Chat Agent

| Scenario | Recommended agent |
|---|---|
| Simple conversational AI in a Go application | `chat.Agent` |
| Multi-turn conversations with history | `chat.Agent` with `KeepConversationHistory: true` |
| Single-turn question/answer | `chat.Agent` with `KeepConversationHistory: false` |
| Streaming responses to a terminal or UI | `chat.Agent` with `GenerateStreamCompletion` |
| Reasoning/chain-of-thought models | `chat.Agent` with `GenerateCompletionWithReasoning` |
| Function calling / tool use | `tools.Agent` |
| HTTP API with SSE streaming | `ServerAgent` or `CrewServerAgent` |

### Key capabilities

- **Standard completion**: Send messages and receive a complete response.
- **Streaming completion**: Receive response chunks in real-time via callbacks.
- **Reasoning support**: Retrieve both the reasoning chain and the final response from reasoning-capable models.
- **Conversation history**: Optionally maintain full conversation history across multiple calls.
- **User message directives**: Inject pre/post directives into user messages for consistent framing.
- **Lifecycle hooks**: Execute custom logic before and after each completion call.
- **JSON export**: Export conversation history as JSON for debugging or persistence.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := chat.NewAgent(
        ctx,
        agents.Config{
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a helpful assistant.",
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.7),
            MaxTokens:   models.Int(2000),
        },
    )
    if err != nil {
        panic(err)
    }

    result, err := agent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: "What is the capital of France?"},
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Response)
    // Output: Paris is the capital of France.
}
```

---

## 3. Agent Configuration

The `agents.Config` struct controls the agent's identity and behavior:

```go
agents.Config{
    Name:                    "my-agent",           // Agent name (optional, for identification)
    Description:             "A helpful assistant", // Agent description (optional)
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1", // LLM engine URL (required)
    APIKey:                  "your-api-key",        // API key (optional, depends on engine)
    SystemInstructions:      "You are a helpful assistant.", // System prompt (recommended)
    KeepConversationHistory: true,                  // Maintain conversation history (default: false)
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | No | Agent identifier, used for logging and multi-agent setups. |
| `Description` | `string` | No | Human-readable description. |
| `EngineURL` | `string` | Yes | URL of the OpenAI-compatible LLM engine. |
| `APIKey` | `string` | No | API key for authenticated engines. |
| `SystemInstructions` | `string` | Recommended | System prompt that defines the agent's personality and behavior. |
| `KeepConversationHistory` | `bool` | No | When `true`, all user and assistant messages are kept across calls. Default: `false`. |

---

## 4. Model Configuration

The `models.Config` struct controls the model's generation parameters:

```go
models.Config{
    Name:             "ai/qwen2.5:1.5B-F16",    // Model ID (required)
    Temperature:      models.Float64(0.7),        // Creativity (0.0 = deterministic, 1.0+ = creative)
    MaxTokens:        models.Int(2000),            // Maximum response length
    TopP:             models.Float64(0.9),         // Nucleus sampling
    FrequencyPenalty: models.Float64(0.0),         // Penalize repeated tokens
    PresencePenalty:  models.Float64(0.0),         // Penalize tokens already present
    Stop:             []string{"\n\n"},            // Stop sequences
    ReasoningEffort:  models.String("medium"),     // Reasoning effort for reasoning models
}
```

All parameters except `Name` are optional and use pointer types (`*float64`, `*int64`) so that `nil` means "use the model's default". Helper functions are provided:

- `models.Float64(v)` returns `*float64`
- `models.Int(v)` returns `*int64`
- `models.String(v)` returns `*string`
- `models.Bool(v)` returns `*bool`

---

## 5. Completion Methods

### GenerateCompletion

The simplest way to get a response from the agent:

```go
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Hello, who are you?"},
})
if err != nil {
    // handle error
}

fmt.Println(result.Response)     // The agent's response text
fmt.Println(result.FinishReason) // "stop", "length", etc.
```

**Return type:** `*CompletionResult`

```go
type CompletionResult struct {
    Response     string // The generated response text
    FinishReason string // Why generation stopped ("stop", "length", etc.)
}
```

### Sending multiple messages

You can send multiple messages in a single call to provide context:

```go
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "My name is Alice."},
    {Role: roles.Assistant, Content: "Nice to meet you, Alice!"},
    {Role: roles.User, Content: "What is my name?"},
})
```

---

## 6. Streaming Completion

### GenerateStreamCompletion

Streaming allows you to receive response chunks as they are generated, providing a real-time experience:

```go
result, err := agent.GenerateStreamCompletion(
    []messages.Message{
        {Role: roles.User, Content: "Tell me a story about a cat."},
    },
    func(chunk string, finishReason string) error {
        fmt.Print(chunk) // Print each chunk as it arrives
        return nil
    },
)
if err != nil {
    // handle error
}
fmt.Println()
fmt.Println("Finish reason:", result.FinishReason)
```

The callback function receives:
- `chunk`: A piece of the response text (may be empty on the final call).
- `finishReason`: Empty for intermediate chunks, set to `"stop"`, `"length"`, etc. on the final chunk.

The method also returns the complete `*CompletionResult` after streaming finishes.

### Stopping a stream

You can interrupt an ongoing stream from another goroutine:

```go
agent.StopStream()
```

---

## 7. Reasoning Support

For models that support chain-of-thought reasoning (e.g., DeepSeek-R1, QwQ), the agent provides reasoning-specific methods that return both the reasoning chain and the final response.

### GenerateCompletionWithReasoning

```go
result, err := agent.GenerateCompletionWithReasoning([]messages.Message{
    {Role: roles.User, Content: "What is 15% of 240?"},
})
if err != nil {
    // handle error
}

fmt.Println("Reasoning:", result.Reasoning)  // The model's reasoning chain
fmt.Println("Response:", result.Response)     // The final answer
fmt.Println("Finish:", result.FinishReason)
```

**Return type:** `*ReasoningResult`

```go
type ReasoningResult struct {
    Response     string // The final response
    Reasoning    string // The reasoning/thinking chain
    FinishReason string // Why generation stopped
}
```

### GenerateStreamCompletionWithReasoning

Streaming variant with separate callbacks for reasoning and response:

```go
result, err := agent.GenerateStreamCompletionWithReasoning(
    []messages.Message{
        {Role: roles.User, Content: "Explain quantum computing."},
    },
    // Reasoning callback
    func(chunk string, finishReason string) error {
        fmt.Print(chunk) // Stream reasoning chunks
        return nil
    },
    // Response callback
    func(chunk string, finishReason string) error {
        fmt.Print(chunk) // Stream response chunks
        return nil
    },
)
```

---

## 8. Conversation History

### Enabling conversation history

Set `KeepConversationHistory: true` in the agent configuration to maintain context across multiple calls:

```go
agent, err := chat.NewAgent(ctx,
    agents.Config{
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "You are Bob, a helpful assistant.",
        KeepConversationHistory: true,
    },
    models.Config{
        Name: "ai/qwen2.5:1.5B-F16",
    },
)
```

With history enabled, the agent remembers previous interactions:

```go
// First call
agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Who is James T Kirk?"},
})

// Second call - the agent knows the context from the first call
result, _ := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Who is his best friend?"},
})
// The agent can answer about Spock because it remembers Kirk
```

### Without conversation history

When `KeepConversationHistory` is `false` (default), each call is independent. The agent does not remember previous interactions.

### Managing messages

```go
// Get all messages in history
msgs := agent.GetMessages()

// Get the approximate context size (in characters)
size := agent.GetContextSize()

// Clear all messages except the system instruction
agent.ResetMessages()

// Remove the last N messages
agent.RemoveLastNMessages(2)

// Manually add a message to history
agent.AddMessage(roles.User, "A manual message")

// Add multiple messages at once
agent.AddMessages([]messages.Message{
    {Role: roles.User, Content: "First message"},
    {Role: roles.Assistant, Content: "First response"},
})
```

### Updating system instructions

```go
agent.SetSystemInstructions("You are now a pirate assistant. Arrr!")
```

---

## 9. User Message Directives

User message directives allow you to automatically prepend or append text to every user message. This is useful for consistently framing user input with additional context or instructions.

### Setting directives

```go
// Add a prefix to every user message
agent.SetUserMessagePreDirectives("Always respond in formal English.")

// Add a suffix to every user message
agent.SetUserMessagePostDirectives("Keep your response under 100 words.")
```

### How directives work

When a user sends "What is Go?", the actual message sent to the model becomes:

```
Always respond in formal English.

What is Go?

Keep your response under 100 words.
```

### Getting current directives

```go
pre := agent.GetUserMessagePreDirectives()
post := agent.GetUserMessagePostDirectives()
```

---

## 10. Lifecycle Hooks (ChatAgentOption)

Lifecycle hooks allow you to execute custom logic before and after each completion call. They are configured as functional options when creating the agent.

### ChatAgentOption

```go
type ChatAgentOption func(*Agent)
```

Options are passed as variadic arguments to `NewAgent`:

```go
agent, err := chat.NewAgent(ctx, agentConfig, modelConfig,
    chat.BeforeCompletion(fn),
    chat.AfterCompletion(fn),
)
```

### BeforeCompletion

Called before each completion (standard and streaming). The hook receives a reference to the agent, allowing you to inspect or modify its state.

```go
chat.BeforeCompletion(func(a *chat.Agent) {
    fmt.Println("About to call the LLM...")
    fmt.Printf("Current context size: %d\n", a.GetContextSize())
    fmt.Printf("Messages count: %d\n", len(a.GetMessages()))
})
```

**Use cases:**
- Logging and monitoring
- Metrics collection (measure request frequency)
- Context size checks before each call
- Dynamic system instruction updates

### AfterCompletion

Called after each completion (standard and streaming), once the full response has been received. The hook receives a reference to the agent.

```go
chat.AfterCompletion(func(a *chat.Agent) {
    fmt.Println("LLM call completed.")
    fmt.Printf("Updated context size: %d\n", a.GetContextSize())
    fmt.Printf("Messages count: %d\n", len(a.GetMessages()))
})
```

**Use cases:**
- Logging response metrics
- Post-processing (e.g., save conversation to database)
- Context size monitoring after response
- Triggering downstream actions

### Complete example with hooks

```go
agent, err := chat.NewAgent(
    ctx,
    agents.Config{
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "You are Bob, a helpful AI assistant.",
        KeepConversationHistory: true,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
        MaxTokens:   models.Int(2000),
    },
    chat.BeforeCompletion(func(a *chat.Agent) {
        fmt.Printf("[BEFORE] Context: %d chars, Messages: %d\n",
            a.GetContextSize(), len(a.GetMessages()))
    }),
    chat.AfterCompletion(func(a *chat.Agent) {
        fmt.Printf("[AFTER] Context: %d chars, Messages: %d\n",
            a.GetContextSize(), len(a.GetMessages()))
    }),
)
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. Hooks are only called when they have been set. The `...ChatAgentOption` parameter is variadic, so existing code without hooks continues to work without any changes.

### Hooks apply to all completion methods

Both `BeforeCompletion` and `AfterCompletion` hooks are triggered by all four completion methods:

| Method | BeforeCompletion | AfterCompletion |
|---|---|---|
| `GenerateCompletion` | Yes | Yes |
| `GenerateCompletionWithReasoning` | Yes | Yes |
| `GenerateStreamCompletion` | Yes | Yes |
| `GenerateStreamCompletionWithReasoning` | Yes | Yes |

---

## 11. Context and State Management

### Getting and setting context

The agent carries a `context.Context` for cancellation and value propagation:

```go
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

### Getting and setting configuration

```go
// Agent configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

// Model configuration
modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)
```

### Agent metadata

```go
agent.Kind()       // Returns agents.Chat
agent.GetName()    // Returns the agent name from config
agent.GetModelID() // Returns the model name from model config
```

---

## 12. JSON Export and Debugging

### Export conversation to JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
if err != nil {
    // handle error
}
fmt.Println(jsonStr)
```

Output:

```json
[
  {
    "role": "system",
    "content": "You are a helpful assistant."
  },
  {
    "role": "user",
    "content": "Hello"
  },
  {
    "role": "assistant",
    "content": "Hello! How can I help you?"
  }
]
```

### Raw request/response JSON

For debugging, you can access the raw JSON sent to and received from the LLM engine:

```go
// Raw (unformatted) JSON
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

// Pretty-printed JSON
prettyReq, err := agent.GetLastRequestJSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 13. API Reference

### Constructor

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    opts ...ChatAgentOption,
) (*Agent, error)
```

Creates a new chat agent. The `opts` parameter accepts zero or more `ChatAgentOption` functional options.

---

### Types

```go
// ChatAgentOption is a functional option for configuring an Agent during creation
type ChatAgentOption func(*Agent)

// CompletionResult represents the result of a chat completion
type CompletionResult struct {
    Response     string
    FinishReason string
}

// ReasoningResult represents the result of a chat completion with reasoning
type ReasoningResult struct {
    Response     string
    Reasoning    string
    FinishReason string
}

// StreamCallback is a function called for each chunk of streaming response
type StreamCallback func(chunk string, finishReason string) error
```

---

### Option Functions

| Function | Description |
|---|---|
| `BeforeCompletion(fn func(*Agent))` | Sets a hook called before each completion (standard and streaming). |
| `AfterCompletion(fn func(*Agent))` | Sets a hook called after each completion (standard and streaming). |

---

### Methods

| Method | Description |
|---|---|
| `GenerateCompletion(msgs []messages.Message) (*CompletionResult, error)` | Send messages and get a complete response. |
| `GenerateCompletionWithReasoning(msgs []messages.Message) (*ReasoningResult, error)` | Send messages and get a response with reasoning chain. |
| `GenerateStreamCompletion(msgs []messages.Message, cb StreamCallback) (*CompletionResult, error)` | Send messages and stream the response via callback. |
| `GenerateStreamCompletionWithReasoning(msgs []messages.Message, reasoningCb StreamCallback, responseCb StreamCallback) (*ReasoningResult, error)` | Send messages and stream both reasoning and response. |
| `GetMessages() []messages.Message` | Get all conversation messages. |
| `GetContextSize() int` | Get the approximate context size in characters. |
| `ResetMessages()` | Clear all messages except system instruction. |
| `RemoveLastNMessages(n int)` | Remove the last N messages from history. |
| `AddMessage(role roles.Role, content string)` | Add a single message to history. |
| `AddMessages(msgs []messages.Message)` | Add multiple messages to history. |
| `SetSystemInstructions(instructions string)` | Update the system instructions. |
| `SetUserMessagePreDirectives(directives string)` | Set text prepended to every user message. |
| `GetUserMessagePreDirectives() string` | Get the current pre-directives. |
| `SetUserMessagePostDirectives(directives string)` | Set text appended to every user message. |
| `GetUserMessagePostDirectives() string` | Get the current post-directives. |
| `StopStream()` | Interrupt the current streaming operation. |
| `ExportMessagesToJSON() (string, error)` | Export conversation history as JSON. |
| `GetConfig() agents.Config` | Get the agent configuration. |
| `SetConfig(config agents.Config)` | Update the agent configuration. |
| `GetModelConfig() models.Config` | Get the model configuration. |
| `SetModelConfig(config models.Config)` | Update the model configuration. |
| `GetContext() context.Context` | Get the agent's context. |
| `SetContext(ctx context.Context)` | Update the agent's context. |
| `GetLastRequestRawJSON() string` | Get the raw JSON of the last request. |
| `GetLastResponseRawJSON() string` | Get the raw JSON of the last response. |
| `GetLastRequestJSON() (string, error)` | Get the pretty-printed JSON of the last request. |
| `GetLastResponseJSON() (string, error)` | Get the pretty-printed JSON of the last response. |
| `Kind() agents.Kind` | Returns `agents.Chat`. |
| `GetName() string` | Returns the agent name. |
| `GetModelID() string` | Returns the model name. |
