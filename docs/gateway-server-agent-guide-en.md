# Gateway Server Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration (Options)](#3-agent-configuration-options)
4. [Crew Management](#4-crew-management)
5. [HTTP Completion Pipeline (handleChatCompletions)](#5-http-completion-pipeline-handlechatcompletions)
6. [HTTP Server and Routes](#6-http-server-and-routes)
7. [Tool Execution Modes](#7-tool-execution-modes)
8. [Intelligent Routing (Orchestrator)](#8-intelligent-routing-orchestrator)
9. [RAG Integration](#9-rag-integration)
10. [Context Compression](#10-context-compression)
11. [Lifecycle Hooks (BeforeCompletion / AfterCompletion)](#11-lifecycle-hooks-beforecompletion--aftercompletion)
12. [Conversation Management](#12-conversation-management)
13. [OpenAI-Compatible Types](#13-openai-compatible-types)
14. [Testing](#14-testing)
15. [API Reference](#15-api-reference)

---

## 1. Introduction

### What is a Gateway Server Agent?

The `gatewayserver.GatewayServerAgent` is a high-level composite agent provided by the Nova SDK (`github.com/snipwise/nova`) that exposes an **OpenAI-compatible HTTP API** (`POST /v1/chat/completions`) backed by a **crew of N.O.V.A. agents**. External clients (such as `qwen-code`, `aider`, `continue.dev`, or any OpenAI SDK) see a single "model", while internally the gateway routes requests to specialized agents.

Unlike the `crewserver.CrewServerAgent` which uses a custom SSE protocol, the Gateway Server Agent speaks the **standard OpenAI Chat Completions API format**, making it a drop-in replacement for the OpenAI API.

### When to use a Gateway Server Agent

| Scenario | Recommended agent |
|---|---|
| OpenAI-compatible API for external tools (qwen-code, aider, etc.) | `gatewayserver.GatewayServerAgent` |
| Passthrough tool_calls to client (client manages tool execution) | `gatewayserver.GatewayServerAgent` with `ToolModePassthrough` |
| Server-side tool execution with OpenAI API format | `gatewayserver.GatewayServerAgent` with `ToolModeAutoExecute` |
| Custom SSE protocol with web-based tool confirmation | `crewserver.CrewServerAgent` |
| CLI-only multi-agent pipeline (no HTTP) | `crew.CrewAgent` |
| Simple direct LLM access | `chat.Agent` |

### Key capabilities

- **OpenAI-compatible API**: Full `POST /v1/chat/completions` support (streaming SSE + non-streaming JSON).
- **Two tool modes**: Passthrough (client executes tools) and auto-execute (server executes tools).
- **Multi-agent crew**: Manage multiple `chat.Agent` instances, each specialized for a topic.
- **Intelligent routing**: Automatically route questions to the most appropriate agent via an orchestrator.
- **Full pipeline**: Context compression, tool calls, RAG injection, and streaming completion.
- **Standard SSE streaming**: `data: {json}\n\n` chunks + `data: [DONE]\n\n` terminator.
- **Models endpoint**: `GET /v1/models` lists all crew agents as available models.
- **Lifecycle hooks**: Execute custom logic before and after each completion request.
- **Functional options pattern**: Configurable via `GatewayServerAgentOption` functions.

---

## 2. Quick Start

### Minimal example with a single agent

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    chatAgent, _ := chat.NewAgent(ctx,
        agents.Config{
            Name:               "assistant",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a helpful assistant.",
        },
        models.Config{
            Name:        "my-model",
            Temperature: models.Float64(0.7),
        },
    )

    gateway, _ := gatewayserver.NewAgent(ctx,
        gatewayserver.WithSingleAgent(chatAgent),
        gatewayserver.WithPort(8080),
    )

    fmt.Println("Gateway server starting on http://localhost:8080")
    log.Fatal(gateway.StartServer())
}
```

**Usage with curl:**

```bash
# Non-streaming
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"assistant","messages":[{"role":"user","content":"Hello!"}]}'

# Streaming
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"assistant","messages":[{"role":"user","content":"Hello!"}],"stream":true}'
```

**Usage with qwen-code:**

```bash
OPENAI_BASE_URL=http://localhost:8080/v1 OPENAI_API_KEY=none qwen-code
```

### Example with multiple agents

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "generic": genericAgent,
}

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    gatewayserver.WithPort(8080),
    gatewayserver.WithOrchestratorAgent(orchestratorAgent),
    gatewayserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
        switch strings.ToLower(topic) {
        case "coding", "programming":
            return "coder"
        case "philosophy", "thinking":
            return "thinker"
        default:
            return "generic"
        }
    }),
)

log.Fatal(gateway.StartServer())
```

---

## 3. Agent Configuration (Options)

Options are passed as variadic arguments to `NewAgent`:

```go
gateway, err := gatewayserver.NewAgent(ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    gatewayserver.WithPort(8080),
    gatewayserver.WithToolsAgent(toolsAgent),
    gatewayserver.WithToolMode(gatewayserver.ToolModeAutoExecute),
    gatewayserver.WithExecuteFn(executeFn),
    gatewayserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 5),
    gatewayserver.WithCompressorAgentAndContextSize(compressorAgent, 7000),
    gatewayserver.WithOrchestratorAgent(orchestratorAgent),
    gatewayserver.WithMatchAgentIdToTopicFn(matchFn),
    gatewayserver.BeforeCompletion(beforeFn),
    gatewayserver.AfterCompletion(afterFn),
)
```

| Option | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Sets the crew of agents and the initially selected agent. **Mandatory** (or `WithSingleAgent`). |
| `WithSingleAgent(chatAgent)` | Creates a crew with a single agent (ID: `"single"`). **Mandatory** (or `WithAgentCrew`). |
| `WithPort(port)` | Sets the HTTP server port as int (default: `8080`). |
| `WithToolsAgent(toolsAgent)` | Attaches a tools agent for function calling. |
| `WithToolMode(mode)` | Sets tool execution mode: `ToolModePassthrough` (default) or `ToolModeAutoExecute`. |
| `WithExecuteFn(fn)` | Sets the function executor for server-side tool execution. |
| `WithConfirmationPromptFn(fn)` | Sets a custom confirmation prompt function for tool calls. |
| `WithMatchAgentIdToTopicFn(fn)` | Sets the function mapping detected topics to agent IDs. |
| `WithRagAgent(ragAgent)` | Attaches a RAG agent for document retrieval. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attaches a RAG agent with similarity configuration. |
| `WithCompressorAgent(compressorAgent)` | Attaches a compressor agent for context compression. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attaches a compressor with a context size limit. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attaches an orchestrator agent for topic detection and routing. |
| `BeforeCompletion(fn)` | Sets a hook called before each completion request. |
| `AfterCompletion(fn)` | Sets a hook called after each completion request. |

### Default values

| Parameter | Default |
|---|---|
| Port | `:8080` |
| ToolMode | `ToolModePassthrough` |
| `SimilarityLimit` | `0.6` |
| `MaxSimilarities` | `3` |
| `ContextSizeLimit` | `8000` |

---

## 4. Crew Management

### Static crew (at creation)

```go
gatewayserver.WithAgentCrew(map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
}, "coder")
```

### Dynamic crew management

```go
// Add an agent at runtime
err := gateway.AddChatAgentToCrew("cook", cookAgent)

// Remove an agent (cannot remove the currently active one)
err := gateway.RemoveChatAgentFromCrew("thinker")

// Get all agents
agents := gateway.GetChatAgents()

// Replace the entire crew
gateway.SetChatAgents(newCrew)
```

### Switching agents manually

```go
// Get currently selected agent
id := gateway.GetSelectedAgentId()

// Switch to a different agent
err := gateway.SetSelectedAgentId("coder")
```

**Note:** Only one agent is active at a time. `GetName()`, `GetModelID()`, `GetMessages()`, etc. all operate on the currently active agent.

---

## 5. HTTP Completion Pipeline (handleChatCompletions)

The `handleChatCompletions` HTTP handler is the main entry point for completion requests. It processes `POST /v1/chat/completions` requests.

### Pipeline steps

1. **BeforeCompletion hook** (if set)
2. **Parse request** (decode OpenAI-format JSON body)
3. **Resolve model** (match `model` field to a crew agent or use current)
4. **Sync messages** (import conversation history from the request)
5. **Context compression** (if compressor configured and context exceeds limit)
6. **RAG context injection** (if RAG agent configured)
7. **Intelligent routing** (if orchestrator configured, detect topic and switch agent)
8. **Tool handling** (dispatch based on tool mode):
   - **Passthrough**: Forward tool definitions to LLM, return tool_calls to client
   - **Auto-execute**: Execute tools server-side, loop until final response
9. **Generate completion** (streaming SSE or non-streaming JSON)
10. **Cleanup tool state**
11. **AfterCompletion hook** (if set)

### Request format (OpenAI-compatible)

```json
POST /v1/chat/completions
{
    "model": "assistant",
    "messages": [
        {"role": "system", "content": "You are helpful."},
        {"role": "user", "content": "Hello!"}
    ],
    "stream": false,
    "temperature": 0.7,
    "tools": [...]
}
```

### Non-streaming response format

```json
{
    "id": "chatcmpl-abc123",
    "object": "chat.completion",
    "created": 1700000000,
    "model": "assistant",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "Hello! How can I help you?"
            },
            "finish_reason": "stop"
        }
    ],
    "usage": {
        "prompt_tokens": 10,
        "completion_tokens": 8,
        "total_tokens": 18
    }
}
```

### Streaming response format (SSE)

```
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"assistant","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"assistant","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"assistant","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

---

## 6. HTTP Server and Routes

### Starting the server

```go
err := gateway.StartServer()
```

### Available routes

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/chat/completions` | Generate a completion (streaming or non-streaming) |
| `GET` | `/v1/models` | List available models (one per crew agent) |
| `GET` | `/health` | Health check |

### Models endpoint

`GET /v1/models` returns all crew agents as available models:

```json
{
    "object": "list",
    "data": [
        {"id": "coder", "object": "model", "created": 1700000000, "owned_by": "nova-gateway"},
        {"id": "thinker", "object": "model", "created": 1700000000, "owned_by": "nova-gateway"},
        {"id": "generic", "object": "model", "created": 1700000000, "owned_by": "nova-gateway"}
    ]
}
```

### CORS

All responses include CORS headers allowing all origins. Preflight `OPTIONS` requests are handled automatically.

---

## 7. Tool Execution Modes

The Gateway Server Agent supports two distinct tool execution modes, controlled via `WithToolMode`:

### ToolModePassthrough (default)

In passthrough mode, the gateway forwards tool definitions to the LLM backend and returns `tool_calls` to the client. The client is responsible for executing the tools and sending results back in subsequent requests. This is the mode used by tools like `qwen-code` and `aider`.

```go
gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithPort(8080),
    // ToolModePassthrough is the default
)
```

**Client-side tool call flow:**

1. Client sends request with `tools` array
2. Gateway forwards tools to LLM
3. LLM returns `tool_calls` in the response (or SSE stream)
4. Client executes the tools locally
5. Client sends a new request including `tool` role messages with results
6. Gateway completes with the final response

**Non-streaming response with tool_calls:**

```json
{
    "id": "chatcmpl-abc123",
    "choices": [{
        "index": 0,
        "message": {
            "role": "assistant",
            "tool_calls": [{
                "id": "call_xyz",
                "type": "function",
                "function": {"name": "calculate_sum", "arguments": "{\"a\":3,\"b\":5}"}
            }]
        },
        "finish_reason": "tool_calls"
    }]
}
```

**Streaming response with tool_calls:**

```
data: {"choices":[{"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"call_xyz","type":"function","function":{"name":"calculate_sum","arguments":""}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"a\":3,\"b\":5}"}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]
```

### ToolModeAutoExecute

In auto-execute mode, the gateway handles tool execution server-side using the configured `ExecuteFn`. The client only sees the final response and is not aware of tool calls.

```go
gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithToolsAgent(toolsAgent),
    gatewayserver.WithToolMode(gatewayserver.ToolModeAutoExecute),
    gatewayserver.WithExecuteFn(func(name string, args string) (string, error) {
        switch name {
        case "calculate_sum":
            // Execute the function
            return `{"result": 8}`, nil
        default:
            return `{"error": "unknown function"}`, fmt.Errorf("unknown: %s", name)
        }
    }),
)
```

**Server-side tool execution flow:**

1. Client sends request (no `tools` array needed)
2. Gateway detects tool calls using the `tools.Agent`
3. Gateway executes each tool via `ExecuteFn`
4. Gateway feeds results back to the LLM
5. Steps 2-4 repeat until LLM produces a final answer
6. Client receives only the final response

---

## 8. Intelligent Routing (Orchestrator)

When an orchestrator agent is attached, the gateway can automatically route questions to the most appropriate specialized agent.

### Setup

```go
orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:               "orchestrator",
        EngineURL:          engineURL,
        SystemInstructions: `Identify the main topic in one word.
            Possible topics: Technology, Philosophy, Cooking, etc.
            Respond in JSON with 'topic_discussion'.`,
    },
    models.Config{Name: "my-model", Temperature: models.Float64(0.0)},
)

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    gatewayserver.WithOrchestratorAgent(orchestratorAgent),
    gatewayserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
        switch strings.ToLower(topic) {
        case "coding", "programming":
            return "coder"
        case "cooking", "food":
            return "cook"
        default:
            return "generic"
        }
    }),
)
```

### How it works

1. The orchestrator analyzes the user's question and detects the topic.
2. The `matchAgentIdToTopicFn` maps the topic to an agent ID.
3. The gateway switches to the matched agent if different from the current one.
4. The completion is generated by the newly selected agent.

### Direct topic detection

```go
agentId, err := gateway.DetectTopicThenGetAgentId("Write a Python function")
// agentId = "coder"
```

---

## 9. RAG Integration

```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, ragModelConfig)

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 5),
)
```

During the completion pipeline, the gateway performs a similarity search and injects relevant context into the conversation before generating the completion.

---

## 10. Context Compression

```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, compressorModelConfig,
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
)

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)
```

At the beginning of each completion request, the context is compressed if it exceeds the configured limit.

### Manual compression

```go
// Compress only if over limit
newSize, err := gateway.CompressChatAgentContextIfOverLimit()
```

---

## 11. Lifecycle Hooks (BeforeCompletion / AfterCompletion)

Lifecycle hooks allow you to execute custom logic before and after each HTTP completion request (`POST /v1/chat/completions`). They are configured as `GatewayServerAgentOption` functional options.

### BeforeCompletion

Called at the very beginning of each `handleChatCompletions` HTTP handler, before request parsing. The hook receives a reference to the gateway server agent.

```go
gatewayserver.BeforeCompletion(func(a *gatewayserver.GatewayServerAgent) {
    fmt.Printf("[BEFORE] Agent: %s\n", a.GetName())
})
```

### AfterCompletion

Called at the very end of each `handleChatCompletions` HTTP handler, after cleanup. The hook receives a reference to the gateway server agent.

```go
gatewayserver.AfterCompletion(func(a *gatewayserver.GatewayServerAgent) {
    fmt.Printf("[AFTER] Agent: %s\n", a.GetName())
})
```

### Complete example

```go
callCount := 0

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithPort(8080),
    gatewayserver.BeforeCompletion(func(a *gatewayserver.GatewayServerAgent) {
        callCount++
        fmt.Printf("[BEFORE] Call #%d - Agent: %s\n", callCount, a.GetName())
    }),
    gatewayserver.AfterCompletion(func(a *gatewayserver.GatewayServerAgent) {
        fmt.Printf("[AFTER] Call #%d - Agent: %s\n", callCount, a.GetName())
    }),
)

log.Fatal(gateway.StartServer())
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. Existing code without hooks continues to work without any changes.

---

## 12. Conversation Management

All conversation methods operate on the **currently active** chat agent:

```go
// Get messages
msgs := gateway.GetMessages()

// Get context size
size := gateway.GetContextSize()

// Reset conversation
gateway.ResetMessages()

// Add a message
gateway.AddMessage(roles.User, "Hello")

// Stop streaming
gateway.StopStream()
```

---

## 13. OpenAI-Compatible Types

The `gatewayserver` package exports all OpenAI-compatible types for use in tests and custom integrations:

### Request types

| Type | Description |
|---|---|
| `ChatCompletionRequest` | The main request body for `POST /v1/chat/completions` |
| `ChatCompletionMessage` | A message in the conversation (role, content, tool_calls, tool_call_id) |
| `MessageContent` | Message content, supports both string and array formats (multi-modal) |
| `ToolDefinition` | A tool available for the model to call |
| `FunctionDefinition` | Describes a callable function (name, description, parameters) |
| `ToolCall` | A tool call made by the assistant |
| `FunctionCall` | Contains the function name and JSON arguments |

The `MessageContent` type automatically handles the different content formats of the OpenAI API:
- Simple string: `"Hello"`
- Array of strings: `["Hello", "world"]`
- Array of multi-modal parts: `[{"type": "text", "text": "Hello"}]`

Use `NewMessageContent("text")` to create a new message content.

### Response types (non-streaming)

| Type | Description |
|---|---|
| `ChatCompletionResponse` | Complete response with id, object, model, choices, usage |
| `ChatCompletionChoice` | A single choice with message and finish_reason |
| `Usage` | Token usage statistics |

### Response types (streaming)

| Type | Description |
|---|---|
| `ChatCompletionChunk` | A single SSE chunk in streaming mode |
| `ChatCompletionChunkChoice` | A single choice with delta and finish_reason |
| `ChatCompletionDelta` | Incremental content (role, content, tool_calls) |

### Other types

| Type | Description |
|---|---|
| `ModelsResponse` | Response for `GET /v1/models` |
| `ModelEntry` | A single model entry |
| `APIError` | OpenAI-compatible error response |
| `APIErrorDetail` | Error details (message, type, code) |

---

## 14. Testing

### Unit tests

The package includes comprehensive unit tests in `gateway_test.go` with a fake LLM backend. Run them with:

```bash
go test ./nova-sdk/agents/gatewayserver/ -v
```

The test suite covers:
- Request/response serialization
- Full HTTP round-trip (non-streaming and streaming)
- SSE parsing and `data: [DONE]` termination
- Models endpoint
- Health endpoint
- Tool call types

### Public test helpers

For integration testing, the package exposes public wrappers around the private HTTP handlers:

```go
// Create a gateway agent and use these methods in tests:
gateway.HandleChatCompletionsForTest(w, r)
gateway.HandleListModelsForTest(w, r)
gateway.HandleHealthForTest(w, r)
```

### Manual testing with curl

See `samples/84-gateway-server-agent/test.sh` (single agent) and `samples/85-gateway-server-agent-crew/test.sh` (crew) for complete curl-based test scripts.

---

## 15. API Reference

### Constructor

```go
func NewAgent(ctx context.Context, options ...GatewayServerAgentOption) (*GatewayServerAgent, error)
```

### Types

```go
type GatewayServerAgentOption func(*GatewayServerAgent) error

type ToolMode int
const (
    ToolModePassthrough ToolMode = iota
    ToolModeAutoExecute
)
```

### Option Functions

| Function | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Sets the crew and initial agent. |
| `WithSingleAgent(chatAgent)` | Creates a single-agent crew. |
| `WithPort(port)` | Sets the HTTP server port (default: 8080). |
| `WithToolsAgent(toolsAgent)` | Attaches tools agent. |
| `WithToolMode(mode)` | Sets tool execution mode. |
| `WithExecuteFn(fn)` | Sets tool execution function. |
| `WithConfirmationPromptFn(fn)` | Sets custom tool confirmation function. |
| `WithMatchAgentIdToTopicFn(fn)` | Sets topic-to-agent mapping function. |
| `WithRagAgent(ragAgent)` | Attaches RAG agent. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attaches RAG with config. |
| `WithCompressorAgent(compressorAgent)` | Attaches compressor agent. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attaches compressor with limit. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attaches orchestrator agent. |
| `BeforeCompletion(fn func(*GatewayServerAgent))` | Sets hook before each completion. |
| `AfterCompletion(fn func(*GatewayServerAgent))` | Sets hook after each completion. |

### Methods

| Method | Description |
|---|---|
| `StartServer() error` | Start the HTTP server with all routes. |
| `GetPort() string` | Get the HTTP port. |
| `GetToolMode() ToolMode` | Get the current tool execution mode. |
| `SetToolMode(mode)` | Set the tool execution mode. |
| `StopStream()` | Stop the current streaming operation. |
| `GetMessages() []messages.Message` | Get messages from active agent. |
| `GetContextSize() int` | Get context size of active agent. |
| `ResetMessages()` | Reset active agent's conversation. |
| `AddMessage(role, content)` | Add message to active agent. |
| `GetChatAgents() map[string]*chat.Agent` | Get all crew agents. |
| `SetChatAgents(crew)` | Replace entire crew. |
| `AddChatAgentToCrew(id, agent) error` | Add agent to crew. |
| `RemoveChatAgentFromCrew(id) error` | Remove agent from crew. |
| `GetSelectedAgentId() string` | Get active agent ID. |
| `SetSelectedAgentId(id) error` | Switch active agent. |
| `DetectTopicThenGetAgentId(query) (string, error)` | Detect topic and return matching agent ID. |
| `SetOrchestratorAgent(orchestratorAgent)` | Set orchestrator agent. |
| `GetOrchestratorAgent() OrchestratorAgent` | Get orchestrator agent. |
| `SetRagAgent(ragAgent)` | Set RAG agent. |
| `GetRagAgent() *rag.Agent` | Get RAG agent. |
| `SetSimilarityLimit(limit)` | Set similarity threshold. |
| `GetSimilarityLimit() float64` | Get similarity threshold. |
| `SetMaxSimilarities(n)` | Set max similarities. |
| `GetMaxSimilarities() int` | Get max similarities. |
| `SetCompressorAgent(compressorAgent)` | Set compressor agent. |
| `GetCompressorAgent() *compressor.Agent` | Get compressor agent. |
| `SetContextSizeLimit(limit)` | Set context size limit. |
| `GetContextSizeLimit() int` | Get context size limit. |
| `CompressChatAgentContextIfOverLimit() (int, error)` | Compress if over limit. |
| `Kind() agents.Kind` | Returns `agents.ChatServer`. |
| `GetName() string` | Returns active agent name. |
| `GetModelID() string` | Returns active agent model ID. |
| `HandleChatCompletionsForTest(w, r)` | Test helper: exposes chat completions handler. |
| `HandleListModelsForTest(w, r)` | Test helper: exposes models handler. |
| `HandleHealthForTest(w, r)` | Test helper: exposes health handler. |
