# Compressor Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration](#3-agent-configuration)
4. [Model Configuration](#4-model-configuration)
5. [Built-in Instructions and Prompts](#5-built-in-instructions-and-prompts)
6. [Compressing Context](#6-compressing-context)
7. [Streaming Compression](#7-streaming-compression)
8. [Options: AgentOption and CompressorAgentOption](#8-options-agentoption-and-compressoragentoption)
9. [Lifecycle Hooks (CompressorAgentOption)](#9-lifecycle-hooks-compressoragentoption)
10. [Context and State Management](#10-context-and-state-management)
11. [JSON Export and Debugging](#11-json-export-and-debugging)
12. [API Reference](#12-api-reference)

---

## 1. Introduction

### What is a Compressor Agent?

The `compressor.Agent` is a specialized agent provided by the Nova SDK (`github.com/snipwise/nova`) that compresses conversation context. It takes a list of messages (typically from a chat agent) and produces a concise summary that preserves key facts, decisions, and context needed for continuation.

This is essential for managing token limits in long-running conversations: instead of sending the full history to the LLM, you compress it and use the summary as a new system message.

### When to use a Compressor Agent

| Scenario | Recommended agent |
|---|---|
| Compress conversation context to reduce token usage | `compressor.Agent` |
| Summarize long conversations for persistence | `compressor.Agent` |
| Free-form conversational AI | `chat.Agent` |
| Structured data extraction | `structured.Agent[T]` |
| Function calling / tool use | `tools.Agent` |
| Intent detection and routing | `orchestrator.Agent` |

### Key capabilities

- **Standard and streaming compression**: Compress context in a single call or stream the result chunk by chunk.
- **Built-in instructions and prompts**: Pre-defined system instructions (Minimalist, Expert, Effective) and compression prompts (Minimalist, Structured, UltraShort, ContinuityFocus).
- **Custom compression prompts**: Override the default compression prompt at creation time or at runtime.
- **Two option types**: `AgentOption` for base-level configuration (e.g., compression prompt) and `CompressorAgentOption` for high-level agent hooks.
- **Lifecycle hooks**: Execute custom logic before and after each compression.
- **JSON debugging**: Inspect raw request/response payloads.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := compressor.NewAgent(
        ctx,
        agents.Config{
            Name:               "Compressor",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: compressor.Instructions.Expert,
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.0),
        },
        compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    )
    if err != nil {
        panic(err)
    }

    // Messages to compress (typically from a chat agent via chatAgent.GetMessages())
    messagesToCompress := []messages.Message{
        {Role: roles.System, Content: "You are a helpful assistant."},
        {Role: roles.User, Content: "Who is James T Kirk?"},
        {Role: roles.Assistant, Content: "James T. Kirk is a fictional character in the Star Trek franchise. He is the captain of the USS Enterprise."},
        {Role: roles.User, Content: "Who is his best friend?"},
        {Role: roles.Assistant, Content: "His best friend is Spock, a half-Vulcan, half-human science officer aboard the Enterprise."},
    }

    result, err := agent.CompressContext(messagesToCompress)
    if err != nil {
        panic(err)
    }

    fmt.Println("Compressed:", result.CompressedText)
    fmt.Println("Finish reason:", result.FinishReason)
}
```

---

## 3. Agent Configuration

```go
agents.Config{
    Name:               "Compressor",                                      // Agent name (optional)
    EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",      // LLM engine URL (required)
    APIKey:             "your-api-key",                                     // API key (optional)
    SystemInstructions: compressor.Instructions.Expert,                     // System prompt (recommended)
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | No | Agent identifier for logging. |
| `EngineURL` | `string` | Yes | URL of the OpenAI-compatible LLM engine. |
| `APIKey` | `string` | No | API key for authenticated engines. |
| `SystemInstructions` | `string` | Recommended | System prompt defining the compression behavior. Use one of the built-in `Instructions` or provide your own. |

---

## 4. Model Configuration

```go
models.Config{
    Name:        "ai/qwen2.5:1.5B-F16",    // Model ID (required)
    Temperature: models.Float64(0.0),        // 0.0 for deterministic compression
    MaxTokens:   models.Int(2000),           // Max response length
}
```

### Recommended settings

- **Temperature**: `0.0` for deterministic, consistent compression.
- **Model**: Smaller models (0.5B-1.5B) work well for compression tasks and are faster.

---

## 5. Built-in Instructions and Prompts

The compressor package provides pre-defined system instructions and compression prompts.

### System Instructions (`compressor.Instructions`)

| Instruction | Description |
|---|---|
| `Instructions.Minimalist` | Simple, concise instruction for basic summarization. |
| `Instructions.Expert` | Detailed instruction with formatting guidelines, compression rules, and output structure. |
| `Instructions.Effective` | Balanced instruction that preserves key information, decisions, preferences, and emotional context. |

```go
// Use a built-in instruction as SystemInstructions
agents.Config{
    SystemInstructions: compressor.Instructions.Expert,
}
```

### Compression Prompts (`compressor.Prompts`)

The compression prompt is the user-level instruction sent alongside the conversation to compress. It is set via `WithCompressionPrompt` or `SetCompressionPrompt`.

| Prompt | Description |
|---|---|
| `Prompts.Minimalist` | (Default) Concise instruction to summarize preserving key facts. |
| `Prompts.Structured` | Requests a structured summary with topics, decisions, and context (under 200 words). |
| `Prompts.UltraShort` | Extracts only key facts, decisions, and essential context. |
| `Prompts.ContinuityFocus` | Focuses on preserving information needed to continue the discussion naturally. |

```go
// Set at creation time
compressor.WithCompressionPrompt(compressor.Prompts.Structured)

// Or change at runtime
agent.SetCompressionPrompt(compressor.Prompts.UltraShort)
```

---

## 6. Compressing Context

### CompressContext

The primary method for compressing a list of messages:

```go
result, err := agent.CompressContext(messagesToCompress)
if err != nil {
    // handle error
}

fmt.Println(result.CompressedText) // The compressed summary
fmt.Println(result.FinishReason)   // "stop", "length", etc.
```

**Return values:**
- `result *CompressionResult`: Contains the compressed text and finish reason.
- `err error`: Error if compression failed.

### Typical workflow with a chat agent

```go
// 1. Get messages from a chat agent
msgs := chatAgent.GetMessages()

// 2. Compress the context
result, err := compressorAgent.CompressContext(msgs)

// 3. Reset the chat agent and use the compressed context
chatAgent.ResetMessages()
chatAgent.AddMessage(roles.System, result.CompressedText)

// 4. Continue the conversation with reduced token usage
```

---

## 7. Streaming Compression

### CompressContextStream

For real-time output, use streaming compression with a callback:

```go
result, err := agent.CompressContextStream(
    messagesToCompress,
    func(chunk string, finishReason string) error {
        fmt.Print(chunk) // Print each chunk as it arrives
        return nil
    },
)
if err != nil {
    // handle error
}

fmt.Println(result.CompressedText) // Full compressed text
fmt.Println(result.FinishReason)   // "stop", "length", etc.
```

**Parameters:**
- `messagesList []messages.Message`: The messages to compress.
- `callback StreamCallback`: A function called for each chunk. Return a non-nil error to stop streaming.

**Return values:**
- `result *CompressionResult`: The full compressed result (accumulated from all chunks).
- `err error`: Error if compression or streaming failed.

---

## 8. Options: AgentOption and CompressorAgentOption

The compressor agent supports two distinct option types, both passed as variadic `...any` arguments to `NewAgent`:

### AgentOption (base-level)

`AgentOption` operates on the internal `*BaseAgent` and configures low-level behavior:

```go
// Set a custom compression prompt at creation time
compressor.WithCompressionPrompt(compressor.Prompts.Structured)
```

### CompressorAgentOption (agent-level)

`CompressorAgentOption` operates on the high-level `*Agent` and configures lifecycle hooks:

```go
// Set lifecycle hooks at creation time
compressor.BeforeCompletion(func(a *compressor.Agent) { ... })
compressor.AfterCompletion(func(a *compressor.Agent) { ... })
```

### Mixing both option types

Both option types can be passed together to `NewAgent`. The constructor uses type assertion to separate them:

```go
agent, err := compressor.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    // AgentOption (base-level)
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    // CompressorAgentOption (agent-level)
    compressor.BeforeCompletion(func(a *compressor.Agent) {
        fmt.Println("Before compression...")
    }),
    compressor.AfterCompletion(func(a *compressor.Agent) {
        fmt.Println("After compression...")
    }),
)
```

---

## 9. Lifecycle Hooks (CompressorAgentOption)

Lifecycle hooks allow you to execute custom logic before and after each compression (both standard and streaming). They are configured as functional options when creating the agent.

### CompressorAgentOption

```go
type CompressorAgentOption func(*Agent)
```

Options are passed as variadic arguments to `NewAgent` alongside `AgentOption`:

```go
agent, err := compressor.NewAgent(ctx, agentConfig, modelConfig,
    compressor.BeforeCompletion(fn),
    compressor.AfterCompletion(fn),
)
```

### BeforeCompletion

Called before each compression (standard or streaming). The hook receives a reference to the agent.

```go
compressor.BeforeCompletion(func(a *compressor.Agent) {
    fmt.Println("About to compress context...")
    fmt.Printf("Agent: %s (%s)\n", a.GetName(), a.GetModelID())
})
```

**Use cases:**
- Logging and monitoring
- Metrics collection (e.g., track compression count)
- Pre-compression state inspection

### AfterCompletion

Called after each compression, once the result is ready. The hook receives a reference to the agent.

```go
compressor.AfterCompletion(func(a *compressor.Agent) {
    fmt.Println("Compression completed.")
    fmt.Printf("Agent: %s (%s)\n", a.GetName(), a.GetModelID())
})
```

**Use cases:**
- Logging compression results
- Post-compression metrics
- Triggering downstream actions
- Auditing/tracking

### Complete example with hooks

```go
agent, err := compressor.NewAgent(
    ctx,
    agents.Config{
        Name:               "Compressor",
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: compressor.Instructions.Expert,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
    },
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    compressor.BeforeCompletion(func(a *compressor.Agent) {
        fmt.Printf("[BEFORE] Agent: %s (%s)\n", a.GetName(), a.GetModelID())
    }),
    compressor.AfterCompletion(func(a *compressor.Agent) {
        fmt.Printf("[AFTER] Agent: %s (%s)\n", a.GetName(), a.GetModelID())
    }),
)
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. The `...any` parameter is variadic, so existing code without hooks continues to work without any changes.

---

## 10. Context and State Management

### Getting and setting context

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

### Changing compression prompt at runtime

```go
agent.SetCompressionPrompt(compressor.Prompts.Structured)
```

### Agent metadata

```go
agent.GetKind()    // Returns agents.Compressor
agent.GetName()    // Returns the agent name from config
agent.GetModelID() // Returns the model name from model config
```

---

## 11. JSON Export and Debugging

### Raw request/response JSON

```go
// Raw (unformatted) JSON
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

// Pretty-printed JSON
prettyReq, err := agent.GetLastRequestJSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 12. API Reference

### Constructor

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    options ...any,
) (*Agent, error)
```

Creates a new compressor agent. The `options` parameter accepts both `AgentOption` (base-level) and `CompressorAgentOption` (agent-level hooks). The constructor separates them internally using type assertion.

---

### Types

```go
// CompressionResult represents the result of a context compression
type CompressionResult struct {
    CompressedText string
    FinishReason   string
}

// StreamCallback is a function called for each chunk of streaming response
type StreamCallback func(chunk string, finishReason string) error

// AgentOption configures the internal BaseAgent (e.g., compression prompt)
type AgentOption func(*BaseAgent)

// CompressorAgentOption configures the high-level Agent (e.g., lifecycle hooks)
type CompressorAgentOption func(*Agent)
```

---

### Built-in Constants

```go
// System instructions
compressor.Instructions.Minimalist   // Simple summarization instruction
compressor.Instructions.Expert       // Detailed compression specialist instruction
compressor.Instructions.Effective    // Balanced instruction with structured output

// Compression prompts
compressor.Prompts.Minimalist        // (Default) Concise summarization prompt
compressor.Prompts.Structured        // Structured summary with topics and decisions
compressor.Prompts.UltraShort        // Key facts and decisions only
compressor.Prompts.ContinuityFocus   // Focus on conversation continuity
```

---

### Option Functions

| Function | Type | Description |
|---|---|---|
| `WithCompressionPrompt(prompt string)` | `AgentOption` | Sets the compression prompt used when compressing context. |
| `BeforeCompletion(fn func(*Agent))` | `CompressorAgentOption` | Sets a hook called before each compression (standard and streaming). |
| `AfterCompletion(fn func(*Agent))` | `CompressorAgentOption` | Sets a hook called after each compression (standard and streaming). |

---

### Methods

| Method | Description |
|---|---|
| `CompressContext(msgs []messages.Message) (*CompressionResult, error)` | Compress messages and return the result. |
| `CompressContextStream(msgs []messages.Message, cb StreamCallback) (*CompressionResult, error)` | Compress messages with streaming output via callback. |
| `SetCompressionPrompt(prompt string)` | Change the compression prompt at runtime. |
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
| `GetKind() agents.Kind` | Returns `agents.Compressor`. |
| `GetName() string` | Returns the agent name. |
| `GetModelID() string` | Returns the model name. |
