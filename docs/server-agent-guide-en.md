# Server Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration](#3-agent-configuration)
4. [Server Options](#4-server-options)
5. [Starting the Server](#5-starting-the-server)
6. [Completion Pipeline](#6-completion-pipeline)
7. [CLI Completion (StreamCompletion)](#7-cli-completion-streamcompletion)
8. [Direct Completion Methods](#8-direct-completion-methods)
9. [Tool Call Integration](#9-tool-call-integration)
10. [RAG Integration](#10-rag-integration)
11. [Context Compression](#11-context-compression)
12. [Lifecycle Hooks (BeforeCompletion / AfterCompletion)](#12-lifecycle-hooks-beforecompletion--aftercompletion)
13. [Conversation Management](#13-conversation-management)
14. [API Reference](#14-api-reference)

---

## 1. Introduction

### What is a Server Agent?

The `server.ServerAgent` is a high-level agent provided by the Nova SDK (`github.com/snipwise/nova`) that wraps a `chat.Agent` and exposes it as an HTTP server with SSE streaming. It orchestrates tool calls, RAG context injection, and context compression in a single pipeline.

### When to use a Server Agent

| Scenario | Recommended agent |
|---|---|
| HTTP server exposing an LLM via REST/SSE | `server.ServerAgent` |
| LLM with tool calls, RAG, and compression via HTTP | `server.ServerAgent` |
| CLI usage with the same pipeline (tools + RAG + compression) | `server.ServerAgent` (via `StreamCompletion`) |
| Simple direct LLM access | `chat.Agent`, `tools.Agent`, etc. |

### Key capabilities

- **HTTP server with SSE streaming**: Serves completions via `POST /completion` with Server-Sent Events.
- **Full pipeline**: Compression, tool calls, RAG context injection, and streaming completion.
- **Dual mode**: Works as an HTTP server (`StartServer`) and as a CLI library (`StreamCompletion`).
- **Tool call notifications**: Sends tool call notifications via SSE for web-based human-in-the-loop.
- **Lifecycle hooks**: Execute custom logic before and after each completion.
- **Functional options pattern**: Configurable via `ServerAgentOption` functions.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/server"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := server.NewAgent(
        ctx,
        agents.Config{
            Name:               "My Server",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a helpful AI assistant.",
        },
        models.Config{
            Name:        "my-model",
            Temperature: models.Float64(0.4),
        },
        server.WithPort(3500),
    )
    if err != nil {
        panic(err)
    }

    log.Fatal(agent.StartServer())
}
```

---

## 3. Agent Configuration

The server agent requires both an `agents.Config` and a `models.Config`:

```go
agents.Config{
    Name:               "My Server",           // Agent name
    EngineURL:          "http://localhost:...", // LLM engine URL
    SystemInstructions: "You are...",          // System prompt
}

models.Config{
    Name:        "model-name",       // Model identifier
    Temperature: models.Float64(0.4), // Temperature
}
```

---

## 4. Server Options

Options are passed as variadic arguments to `NewAgent`:

```go
agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithPort(3500),
    server.WithToolsAgent(toolsAgent),
    server.WithExecuteFn(executeFn),
    server.WithRagAgent(ragAgent),
    server.WithCompressorAgent(compressorAgent),
    server.BeforeCompletion(beforeFn),
    server.AfterCompletion(afterFn),
)
```

| Option | Description |
|---|---|
| `WithPort(port int)` | Sets the HTTP server port (default: 8080). |
| `WithExecuteFn(fn)` | Sets the function executor for tool calls. |
| `WithConfirmationPromptFn(fn)` | Sets the confirmation prompt for human-in-the-loop. |
| `WithTLSCert(certData, keyData []byte)` | Enables HTTPS with PEM-encoded certificate and key data in memory. |
| `WithTLSCertFromFile(certPath, keyPath string)` | Enables HTTPS with certificate and key file paths. |
| `WithToolsAgent(toolsAgent)` | Attaches a tools agent for function calling. |
| `WithCompressorAgent(compressorAgent)` | Attaches a compressor agent for context compression. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attaches a compressor agent with a context size limit. |
| `WithRagAgent(ragAgent)` | Attaches a RAG agent for document retrieval. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attaches a RAG agent with similarity configuration. |
| `BeforeCompletion(fn)` | Sets a hook called before each completion. |
| `AfterCompletion(fn)` | Sets a hook called after each completion. |

### HTTPS Support

The Server Agent supports HTTPS for secure communication. When TLS certificates are provided, the server will automatically use HTTPS instead of HTTP.

**Option 1: Using certificate files** (recommended for production):

```go
agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithPort(443),
    server.WithTLSCertFromFile("server.crt", "server.key"),
)
```

**Option 2: Using certificate data in memory**:

```go
certData, _ := os.ReadFile("server.crt")
keyData, _ := os.ReadFile("server.key")

agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithPort(443),
    server.WithTLSCert(certData, keyData),
)
```

**Important notes**:
- HTTPS is **optional** - without TLS certificates, the server runs on HTTP (backward compatible)
- For production, use certificates from a trusted CA (e.g., Let's Encrypt)
- For development/testing, you can use self-signed certificates
- See `/samples/90-https-server-example` for a complete example

---

## 5. Starting the Server

```go
log.Fatal(agent.StartServer())
```

### HTTP endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/completion` | Send a completion request (SSE streaming response). |
| `POST` | `/completion/stop` | Stop the current streaming operation. |
| `POST` | `/memory/reset` | Reset conversation history. |
| `GET` | `/memory/messages/list` | Get conversation messages. |
| `GET` | `/memory/messages/context-size` | Get context size in tokens. |
| `POST` | `/operation/validate` | Validate a pending tool call operation. |
| `POST` | `/operation/cancel` | Cancel a pending tool call operation. |
| `POST` | `/operation/reset` | Reset all pending operations. |
| `GET` | `/models` | Get model information. |
| `GET` | `/health` | Health check. |

---

## 6. Completion Pipeline

The `POST /completion` handler (`handleCompletion`) follows this pipeline:

1. **BeforeCompletion hook** (if set)
2. **Compress context** if compressor agent is configured and context exceeds limit
3. **Parse request** (extracts the question)
4. **Setup SSE streaming**
5. **Tool call detection and execution** (if tools agent is configured)
6. **RAG context injection** (if RAG agent is configured)
7. **Generate streaming completion**
8. **Cleanup tool state**
9. **AfterCompletion hook** (if set)

---

## 7. CLI Completion (StreamCompletion)

The `StreamCompletion` method provides the same pipeline for CLI usage:

```go
result, err := agent.StreamCompletion(
    "What is 2 + 2?",
    func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil
    },
)
```

The CLI pipeline mirrors the HTTP pipeline:

1. **BeforeCompletion hook** (if set)
2. **Compress context** if needed
3. **Tool call detection and execution**
4. **RAG context injection**
5. **Generate streaming completion**
6. **AfterCompletion hook** (if set)

---

## 8. Direct Completion Methods

The server agent also exposes direct completion methods that delegate to the internal `chat.Agent`:

```go
// Non-streaming
result, err := agent.GenerateCompletion(userMessages)

// Streaming
result, err := agent.GenerateStreamCompletion(userMessages, callback)

// With reasoning
result, err := agent.GenerateCompletionWithReasoning(userMessages)
result, err := agent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCb, responseCb)
```

**Note:** These methods bypass the full pipeline (no compression, no tool calls, no RAG). They delegate directly to the underlying `chat.Agent`. The lifecycle hooks are **not** triggered by these methods.

---

## 9. Tool Call Integration

```go
toolsAgent, _ := tools.NewAgent(ctx, toolsConfig, toolsModelConfig,
    tools.WithTools(myTools),
)

agent, _ := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithToolsAgent(toolsAgent),
    server.WithExecuteFn(func(name string, args string) (string, error) {
        // Execute the tool and return result
        return `{"result": "ok"}`, nil
    }),
)
```

---

## 10. RAG Integration

```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, ragModelConfig)

agent, _ := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithRagAgentAndSimilarityConfig(ragAgent, 0.3, 3),
)
```

---

## 11. Context Compression

```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, compressorModelConfig)

agent, _ := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithCompressorAgentAndContextSize(compressorAgent, 4096),
)
```

---

## 12. Lifecycle Hooks (BeforeCompletion / AfterCompletion)

Lifecycle hooks allow you to execute custom logic before and after each completion. They are configured as `ServerAgentOption` functional options.

### BeforeCompletion

Called before each completion (HTTP `handleCompletion` and CLI `StreamCompletion`). The hook receives a reference to the server agent.

```go
server.BeforeCompletion(func(a *server.ServerAgent) {
    fmt.Printf("[BEFORE] Agent: %s\n", a.GetName())
})
```

### AfterCompletion

Called after each completion. The hook receives a reference to the server agent.

```go
server.AfterCompletion(func(a *server.ServerAgent) {
    fmt.Printf("[AFTER] Agent: %s\n", a.GetName())
})
```

### Hook placement

| Method | Hooks triggered |
|---|---|
| `handleCompletion` (HTTP POST /completion) | Yes |
| `StreamCompletion` (CLI) | Yes |
| `GenerateCompletion` | No (delegates to chat agent) |
| `GenerateStreamCompletion` | No (delegates to chat agent) |
| `GenerateCompletionWithReasoning` | No (delegates to chat agent) |
| `GenerateStreamCompletionWithReasoning` | No (delegates to chat agent) |

The hooks are in the full pipeline methods (`handleCompletion` and `StreamCompletion`) which orchestrate compression, tool calls, RAG, and completion. The `Generate*` methods delegate directly to the internal chat agent and do not trigger server-level hooks.

### Complete example

```go
callCount := 0

agent, err := server.NewAgent(
    ctx,
    agents.Config{
        Name:               "My Server",
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "You are a helpful assistant.",
    },
    models.Config{
        Name:        "my-model",
        Temperature: models.Float64(0.4),
    },
    server.WithPort(3500),
    server.BeforeCompletion(func(a *server.ServerAgent) {
        callCount++
        fmt.Printf("[BEFORE] Call #%d\n", callCount)
    }),
    server.AfterCompletion(func(a *server.ServerAgent) {
        fmt.Printf("[AFTER] Call #%d\n", callCount)
    }),
)
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. Existing code without hooks continues to work without any changes.

---

## 13. Conversation Management

```go
// Get messages
msgs := agent.GetMessages()

// Get context size
tokens := agent.GetContextSize()

// Reset conversation
agent.ResetMessages()

// Add a message
agent.AddMessage(roles.User, "Hello")

// Export to JSON
jsonStr, err := agent.ExportMessagesToJSON()

// Stop streaming
agent.StopStream()
```

---

## 14. API Reference

### Constructor

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    options ...ServerAgentOption,
) (*ServerAgent, error)
```

### Types

```go
type ServerAgentOption func(*ServerAgent) error
```

### Option Functions

| Function | Description |
|---|---|
| `WithPort(port int)` | Sets the HTTP server port. |
| `WithExecuteFn(fn)` | Sets the function executor for tool calls. |
| `WithConfirmationPromptFn(fn)` | Sets the confirmation prompt function. |
| `WithToolsAgent(toolsAgent)` | Attaches a tools agent. |
| `WithCompressorAgent(compressorAgent)` | Attaches a compressor agent. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attaches a compressor agent with context size limit. |
| `WithRagAgent(ragAgent)` | Attaches a RAG agent. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attaches a RAG agent with similarity config. |
| `BeforeCompletion(fn func(*ServerAgent))` | Sets a hook called before each completion. |
| `AfterCompletion(fn func(*ServerAgent))` | Sets a hook called after each completion. |

### Methods

| Method | Description |
|---|---|
| `StartServer() error` | Starts the HTTP server with all routes. |
| `StreamCompletion(question, callback) (*chat.CompletionResult, error)` | Full pipeline completion for CLI usage. |
| `GenerateCompletion(msgs) (*chat.CompletionResult, error)` | Direct completion (delegates to chat agent). |
| `GenerateStreamCompletion(msgs, callback) (*chat.CompletionResult, error)` | Direct streaming completion (delegates to chat agent). |
| `GenerateCompletionWithReasoning(msgs) (*chat.ReasoningResult, error)` | Direct completion with reasoning. |
| `GenerateStreamCompletionWithReasoning(msgs, reasoningCb, responseCb) (*chat.ReasoningResult, error)` | Direct streaming with reasoning. |
| `StopStream()` | Stop the current streaming operation. |
| `GetMessages() []messages.Message` | Get conversation messages. |
| `GetContextSize() int` | Get context size in tokens. |
| `ResetMessages()` | Reset conversation history. |
| `AddMessage(role, content)` | Add a message to conversation. |
| `ExportMessagesToJSON() (string, error)` | Export conversation as JSON. |
| `Kind() agents.Kind` | Returns `agents.ChatServer`. |
| `GetName() string` | Returns agent name. |
| `GetModelID() string` | Returns model ID. |
| `SetPort(port string)` | Set HTTP port. |
| `GetPort() string` | Get HTTP port. |
| `SetToolsAgent(toolsAgent)` | Set tools agent. |
| `GetToolsAgent() *tools.Agent` | Get tools agent. |
| `SetRagAgent(ragAgent)` | Set RAG agent. |
| `GetRagAgent() *rag.Agent` | Get RAG agent. |
| `SetCompressorAgent(compressorAgent)` | Set compressor agent. |
| `GetCompressorAgent() *compressor.Agent` | Get compressor agent. |
| `SetContextSizeLimit(limit)` | Set context size limit for compression. |
| `GetContextSizeLimit() int` | Get context size limit. |
