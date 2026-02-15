# Crew Server Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration (Options)](#3-agent-configuration-options)
4. [Crew Management](#4-crew-management)
5. [HTTP Completion Pipeline (handleCompletion)](#5-http-completion-pipeline-handlecompletion)
6. [HTTP Server and Routes](#6-http-server-and-routes)
7. [Intelligent Routing (Orchestrator)](#7-intelligent-routing-orchestrator)
8. [Tool Call Integration](#8-tool-call-integration)
9. [RAG Integration](#9-rag-integration)
10. [Context Compression](#10-context-compression)
11. [Lifecycle Hooks (BeforeCompletion / AfterCompletion)](#11-lifecycle-hooks-beforecompletion--aftercompletion)
12. [Direct Completion Methods](#12-direct-completion-methods)
13. [Conversation Management](#13-conversation-management)
14. [Special Commands](#14-special-commands)
15. [API Reference](#15-api-reference)

---

## 1. Introduction

### What is a Crew Server Agent?

The `crewserver.CrewServerAgent` is a high-level composite agent provided by the Nova SDK (`github.com/snipwise/nova`) that combines a **crew of multiple chat agents** with an **HTTP server** exposing SSE (Server-Sent Events) streaming endpoints. It extends the `BaseServerAgent` with crew-specific functionality: multi-agent routing, tool call management with web-based confirmation, RAG context injection, context compression, and intelligent routing.

### When to use a Crew Server Agent

| Scenario | Recommended agent |
|---|---|
| HTTP server with multiple specialized agents and topic routing | `crewserver.CrewServerAgent` |
| Web-based tool call confirmation with SSE streaming | `crewserver.CrewServerAgent` |
| HTTP API with full pipeline: tools + RAG + compression + routing | `crewserver.CrewServerAgent` |
| CLI-only multi-agent pipeline (no HTTP) | `crew.CrewAgent` |
| Simple HTTP server with a single agent | `server.ServerAgent` |
| Simple direct LLM access | `chat.Agent` |

### Key capabilities

- **HTTP server with SSE streaming**: Built-in HTTP server with CORS support and real-time streaming.
- **Multi-agent crew**: Manage multiple `chat.Agent` instances, each specialized for a topic.
- **Intelligent routing**: Automatically route questions to the most appropriate agent via an orchestrator.
- **Full pipeline**: Context compression, tool calls (with web confirmation), RAG injection, and streaming completion.
- **Dynamic crew management**: Add or remove agents at runtime.
- **Lifecycle hooks**: Execute custom logic before and after each HTTP completion request.
- **Parallel tool calls**: Support for both sequential and parallel tool call execution.
- **Functional options pattern**: Configurable via `CrewServerAgentOption` functions.

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
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
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
            Temperature: models.Float64(0.4),
        },
    )

    crewServerAgent, _ := crewserver.NewAgent(ctx,
        crewserver.WithSingleAgent(chatAgent),
        crewserver.WithPort(3500),
    )

    fmt.Printf("Starting server on http://localhost%s\n", crewServerAgent.GetPort())
    log.Fatal(crewServerAgent.StartServer())
}
```

### Example with multiple agents

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "generic": genericAgent,
}

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithPort(9090),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
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

log.Fatal(crewServerAgent.StartServer())
```

---

## 3. Agent Configuration (Options)

Options are passed as variadic arguments to `NewAgent`:

```go
crewServerAgent, err := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithPort(9090),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithRagAgent(ragAgent),
    crewserver.WithCompressorAgent(compressorAgent),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.BeforeCompletion(beforeFn),
    crewserver.AfterCompletion(afterFn),
)
```

| Option | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Sets the crew of agents and the initially selected agent. **Mandatory** (or `WithSingleAgent`). |
| `WithSingleAgent(chatAgent)` | Creates a crew with a single agent (ID: `"single"`). **Mandatory** (or `WithAgentCrew`). |
| `WithPort(port)` | Sets the HTTP server port as int (default: `3500`). |
| `WithMatchAgentIdToTopicFn(fn)` | Sets the function mapping detected topics to agent IDs. |
| `WithExecuteFn(fn)` | Sets the function executor for tool calls. |
| `WithConfirmationPromptFn(fn)` | Sets a custom confirmation prompt function for tool calls (replaces web-based confirmation). |
| `WithTLSCert(certData, keyData []byte)` | Enables HTTPS with PEM-encoded certificate and key data in memory. |
| `WithTLSCertFromFile(certPath, keyPath string)` | Enables HTTPS with certificate and key file paths. |
| `WithToolsAgent(toolsAgent)` | Attaches a tools agent for function calling. |
| `WithCompressorAgent(compressorAgent)` | Attaches a compressor agent for context compression. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attaches a compressor with a context size limit. |
| `WithRagAgent(ragAgent)` | Attaches a RAG agent for document retrieval. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attaches a RAG agent with similarity configuration. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attaches an orchestrator agent for topic detection and routing. |
| `BeforeCompletion(fn)` | Sets a hook called before each `handleCompletion` call. |
| `AfterCompletion(fn)` | Sets a hook called after each `handleCompletion` call. |

### Default values

| Parameter | Default |
|---|---|
| Port | `:3500` |
| `SimilarityLimit` | `0.6` (inherited from `BaseServerAgent`) |
| `MaxSimilarities` | `3` (inherited from `BaseServerAgent`) |
| `ContextSizeLimit` | `8000` (inherited from `BaseServerAgent`) |

### HTTPS Support

The Crew Server Agent supports HTTPS for secure communication. When TLS certificates are provided, the server will automatically use HTTPS instead of HTTP.

```go
// Option 1: Using certificate files (recommended)
crewServerAgent, err := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithPort(443),
    crewserver.WithTLSCertFromFile("server.crt", "server.key"),
)

// Option 2: Using certificate data in memory
certData, _ := os.ReadFile("server.crt")
keyData, _ := os.ReadFile("server.key")

crewServerAgent, err := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithPort(443),
    crewserver.WithTLSCert(certData, keyData),
)
```

**Important notes**:
- HTTPS is **optional** - without TLS certificates, the server runs on HTTP (backward compatible)
- For production, use certificates from a trusted CA (e.g., Let's Encrypt)
- See `/samples/90-https-server-example` for a complete example

---

## 4. Crew Management

### Static crew (at creation)

```go
crewserver.WithAgentCrew(map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
}, "coder")
```

### Dynamic crew management

```go
// Add an agent at runtime
err := crewServerAgent.AddChatAgentToCrew("cook", cookAgent)

// Remove an agent (cannot remove the currently active one)
err := crewServerAgent.RemoveChatAgentFromCrew("thinker")

// Get all agents
agents := crewServerAgent.GetChatAgents()

// Replace the entire crew
crewServerAgent.SetChatAgents(newCrew)
```

### Switching agents manually

```go
// Get currently selected agent
id := crewServerAgent.GetSelectedAgentId()

// Switch to a different agent
err := crewServerAgent.SetSelectedAgentId("coder")
```

**Note:** Only one agent is active at a time. `GetName()`, `GetModelID()`, `GetMessages()`, etc. all operate on the currently active agent.

---

## 5. HTTP Completion Pipeline (handleCompletion)

The `handleCompletion` HTTP handler is the main entry point for completion requests. It processes `POST /completion` requests with SSE streaming.

### Pipeline steps

1. **BeforeCompletion hook** (if set)
2. **Parse request** (extract question from JSON body)
3. **Setup SSE streaming** (headers and flusher)
4. **Context compression** (if compressor configured and context exceeds limit, with SSE notifications)
5. **Setup notification channel** (for tool call notifications)
6. **Tool call detection and execution** (if tools agent configured, with web confirmation)
7. **Close notification channel**
8. **Generate streaming completion** (if needed: RAG + routing + stream)
9. **Cleanup tool state**
10. **AfterCompletion hook** (if set)

### Request format

```json
POST /completion
{
    "data": {
        "message": "Your question here"
    }
}
```

### SSE response format

```
data: {"message": "chunk of text"}

data: {"message": "", "finish_reason": "stop"}
```

---

## 6. HTTP Server and Routes

### Starting the server

```go
err := crewServerAgent.StartServer()
```

### Available routes

| Method | Path | Description |
|---|---|---|
| `POST` | `/completion` | Generate a streaming completion (SSE) |
| `POST` | `/completion/stop` | Stop the current streaming operation |
| `POST` | `/memory/reset` | Reset conversation history |
| `GET` | `/memory/messages/list` | List all conversation messages |
| `GET` | `/memory/messages/context-size` | Get current context size |
| `POST` | `/operation/validate` | Validate a pending tool call operation |
| `POST` | `/operation/cancel` | Cancel a pending tool call operation |
| `POST` | `/operation/reset` | Reset pending operations |
| `GET` | `/models` | Get model information |
| `GET` | `/health` | Health check |
| `GET` | `/current-agent` | Get current agent information (ID, name, model) |

### Custom routes

The `Mux` field is exposed after `StartServer` initializes it, allowing custom routes:

```go
// The Mux is available after StartServer sets it up
// For custom routes, you can add handlers before calling StartServer
// by accessing agent.Mux after it's created in StartServer
```

### CORS

All responses include CORS headers allowing all origins. Preflight `OPTIONS` requests are handled automatically.

---

## 7. Intelligent Routing (Orchestrator)

When an orchestrator agent is attached, the crew server agent can automatically route questions to the most appropriate specialized agent.

### Automatic Routing Configuration (Recommended)

When you use `WithOrchestratorAgent`, the crew server agent **automatically configures routing** using the orchestrator's `GetAgentForTopic` method. You don't need to provide `WithMatchAgentIdToTopicFn` unless you have custom routing logic.

**Option 1: Orchestrator with built-in routing configuration**

```go
// Load routing configuration from JSON
routingConfig, _ := loadRoutingConfig("agent-routing.json")

orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:               "orchestrator",
        EngineURL:          engineURL,
        SystemInstructions: `Identify the main topic in one word.
            Possible topics: Technology, Philosophy, Cooking, etc.
            Respond in JSON with 'topic_discussion'.`,
    },
    models.Config{Name: "my-model", Temperature: models.Float64(0.0)},
    orchestrator.WithRoutingConfig(*routingConfig), // Configure routing in orchestrator
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithOrchestratorAgent(orchestratorAgent), // Auto-configures routing with GetAgentForTopic
)
```

**Routing configuration format (agent-routing.json):**

```json
{
    "routing": [
        {
            "topics": ["coding", "programming", "development"],
            "agent": "coder"
        },
        {
            "topics": ["cooking", "food", "recipe"],
            "agent": "cook"
        }
    ],
    "default_agent": "generic"
}
```

**Option 2: Orchestrator with custom inline matching function**

If you need custom routing logic beyond simple topic matching, you can still provide `WithMatchAgentIdToTopicFn`:

```go
orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:               "orchestrator",
        EngineURL:          engineURL,
        SystemInstructions: `Identify the main topic in one word...`,
    },
    models.Config{Name: "my-model", Temperature: models.Float64(0.0)},
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    // Override auto-configuration with custom logic
    crewserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
        // Custom routing logic with additional conditions
        if currentAgentId == "coder" && strings.ToLower(topic) == "philosophy" {
            // Keep using coder if already coding
            return currentAgentId
        }
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

1. The orchestrator analyzes the user's question and detects the topic using `IdentifyIntent` or `IdentifyTopicFromText`.
2. **If no custom `matchAgentIdToTopicFn` is provided**: The crew server agent automatically calls `orchestratorAgent.GetAgentForTopic(topic)` to get the agent ID.
3. **If a custom `matchAgentIdToTopicFn` is provided**: The crew server agent uses your custom function instead.
4. The crew server agent switches to the matched agent if different from the current one.
5. The completion is generated by the newly selected agent.

### Direct topic detection

```go
agentId, err := crewServerAgent.DetectTopicThenGetAgentId("Write a Python function")
// agentId = "coder"
```

---

## 8. Tool Call Integration

The crew server agent supports two modes of tool call confirmation:

### Web-based confirmation (default)

When no custom `confirmationPromptFn` is provided, tool calls trigger a web-based confirmation flow via SSE notifications:

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(func(name string, args string) (string, error) {
        return `{"result": "ok"}`, nil
    }),
)
```

The web client receives tool call notifications via SSE and can validate or cancel operations via:
- `POST /operation/validate` - Approve the tool call
- `POST /operation/cancel` - Reject the tool call

### Custom confirmation function

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(executeFn),
    crewserver.WithConfirmationPromptFn(func(name string, args string) tools.ConfirmationResponse {
        return tools.Confirm // Always confirm
    }),
)
```

### Parallel tool calls

When the tools agent is configured with `ParallelToolCalls: true`, the crew server agent automatically uses the parallel detection methods:

```go
toolsAgent, _ := tools.NewAgent(ctx, toolsConfig,
    models.Config{
        Name:              "my-model",
        ParallelToolCalls: models.Bool(true),
    },
    tools.WithTools(myTools),
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(executeFn),
)
```

---

## 9. RAG Integration

```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, ragModelConfig)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 5),
)
```

During the completion pipeline, the crew server agent performs a similarity search and injects relevant context into the conversation before generating the completion.

---

## 10. Context Compression

```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, compressorModelConfig,
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)
```

At the beginning of each completion request, the context is compressed if it exceeds the configured limit. The client receives SSE notifications about the compression process.

### Manual compression

```go
// Compress only if over limit
newSize, err := crewServerAgent.CompressChatAgentContextIfOverLimit()

// Force compression
newSize, err := crewServerAgent.CompressChatAgentContext()
```

---

## 11. Lifecycle Hooks (BeforeCompletion / AfterCompletion)

Lifecycle hooks allow you to execute custom logic before and after each HTTP completion request (`POST /completion`). They are configured as `CrewServerAgentOption` functional options.

### BeforeCompletion

Called at the very beginning of each `handleCompletion` HTTP handler, before request parsing. The hook receives a reference to the crew server agent.

```go
crewserver.BeforeCompletion(func(a *crewserver.CrewServerAgent) {
    fmt.Printf("[BEFORE] Agent: %s\n", a.GetName())
})
```

### AfterCompletion

Called at the very end of each `handleCompletion` HTTP handler, after cleanup. The hook receives a reference to the crew server agent.

```go
crewserver.AfterCompletion(func(a *crewserver.CrewServerAgent) {
    fmt.Printf("[AFTER] Agent: %s\n", a.GetName())
})
```

### Hook placement

| Method | Hooks triggered |
|---|---|
| `POST /completion` (handleCompletion) | Yes |
| `GenerateCompletion` | No (delegates to current chat agent) |
| `GenerateStreamCompletion` | No (delegates to current chat agent) |
| `GenerateCompletionWithReasoning` | No (delegates to current chat agent) |
| `GenerateStreamCompletionWithReasoning` | No (delegates to current chat agent) |

The hooks are in `handleCompletion`, which is the HTTP completion pipeline. The `Generate*` methods delegate directly to the currently active `chat.Agent` and do not trigger crew-level hooks.

### Complete example

```go
callCount := 0

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.BeforeCompletion(func(a *crewserver.CrewServerAgent) {
        callCount++
        fmt.Printf("[BEFORE] Call #%d - Agent: %s\n", callCount, a.GetName())
    }),
    crewserver.AfterCompletion(func(a *crewserver.CrewServerAgent) {
        fmt.Printf("[AFTER] Call #%d - Agent: %s\n", callCount, a.GetName())
    }),
)

log.Fatal(crewServerAgent.StartServer())
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. Existing code without hooks continues to work without any changes.

---

## 12. Direct Completion Methods

The crew server agent exposes direct completion methods that delegate to the currently active `chat.Agent`:

```go
// Non-streaming
result, err := crewServerAgent.GenerateCompletion(userMessages)

// Streaming
result, err := crewServerAgent.GenerateStreamCompletion(userMessages, callback)

// With reasoning
result, err := crewServerAgent.GenerateCompletionWithReasoning(userMessages)
result, err := crewServerAgent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCb, responseCb)
```

**Note:** These methods bypass the full HTTP pipeline (no compression, no tool calls, no RAG, no routing, no SSE). They delegate directly to the currently active chat agent. Lifecycle hooks are **not** triggered.

---

## 13. Conversation Management

All conversation methods operate on the **currently active** chat agent:

```go
// Get messages
msgs := crewServerAgent.GetMessages()

// Get context size
size := crewServerAgent.GetContextSize()

// Reset conversation
crewServerAgent.ResetMessages()

// Add a message
crewServerAgent.AddMessage(roles.User, "Hello")

// Export to JSON
jsonStr, err := crewServerAgent.ExportMessagesToJSON()

// Stop streaming
crewServerAgent.StopStream()
```

---

## 14. Special Commands

The crew server agent supports special command prefixes in the question:

### Select agent

Send a message prefixed with `[select-agent <id>]` to manually switch the active agent:

```
[select-agent coder]
```

### List agents

Send `[agent-list]` to get a list of all available agents:

```
[agent-list]
```

These commands are processed before the standard completion flow and return their results via SSE.

---

## 15. API Reference

### Constructor

```go
func NewAgent(ctx context.Context, options ...CrewServerAgentOption) (*CrewServerAgent, error)
```

### Types

```go
type CrewServerAgentOption func(*CrewServerAgent) error
```

### Option Functions

| Function | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Sets the crew and initial agent. |
| `WithSingleAgent(chatAgent)` | Creates a single-agent crew. |
| `WithPort(port)` | Sets the HTTP server port (default: 3500). |
| `WithMatchAgentIdToTopicFn(fn)` | Sets topic-to-agent mapping function. |
| `WithExecuteFn(fn)` | Sets tool execution function. |
| `WithConfirmationPromptFn(fn)` | Sets custom tool confirmation function. |
| `WithToolsAgent(toolsAgent)` | Attaches tools agent. |
| `WithCompressorAgent(compressorAgent)` | Attaches compressor agent. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attaches compressor with limit. |
| `WithRagAgent(ragAgent)` | Attaches RAG agent. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attaches RAG with config. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attaches orchestrator agent. |
| `BeforeCompletion(fn func(*CrewServerAgent))` | Sets hook before each handleCompletion. |
| `AfterCompletion(fn func(*CrewServerAgent))` | Sets hook after each handleCompletion. |

### Methods

| Method | Description |
|---|---|
| `StartServer() error` | Start the HTTP server with all routes. |
| `SetPort(port)` | Set the HTTP port. |
| `GetPort() string` | Get the HTTP port. |
| `GenerateCompletion(msgs) (*chat.CompletionResult, error)` | Direct completion (delegates to active agent). |
| `GenerateStreamCompletion(msgs, callback) (*chat.CompletionResult, error)` | Direct streaming (delegates to active agent). |
| `GenerateCompletionWithReasoning(msgs) (*chat.ReasoningResult, error)` | Direct completion with reasoning. |
| `GenerateStreamCompletionWithReasoning(msgs, reasoningCb, responseCb) (*chat.ReasoningResult, error)` | Direct streaming with reasoning. |
| `StopStream()` | Stop the current streaming operation. |
| `GetMessages() []messages.Message` | Get messages from active agent. |
| `GetContextSize() int` | Get context size of active agent. |
| `ResetMessages()` | Reset active agent's conversation. |
| `AddMessage(role, content)` | Add message to active agent. |
| `ExportMessagesToJSON() (string, error)` | Export active agent's conversation. |
| `GetChatAgents() map[string]*chat.Agent` | Get all crew agents. |
| `SetChatAgents(crew)` | Replace entire crew. |
| `AddChatAgentToCrew(id, agent) error` | Add agent to crew. |
| `RemoveChatAgentFromCrew(id) error` | Remove agent from crew. |
| `GetSelectedAgentId() string` | Get active agent ID. |
| `SetSelectedAgentId(id) error` | Switch active agent. |
| `DetectTopicThenGetAgentId(query) (string, error)` | Detect topic and return matching agent ID. |
| `SetOrchestratorAgent(orchestratorAgent)` | Set orchestrator agent. |
| `GetOrchestratorAgent() OrchestratorAgent` | Get orchestrator agent. |
| `SetToolsAgent(toolsAgent)` | Set tools agent. |
| `GetToolsAgent() *tools.Agent` | Get tools agent. |
| `SetExecuteFunction(fn)` | Set tool execution function. |
| `SetConfirmationPromptFn(fn)` | Set tool confirmation function. |
| `GetConfirmationPromptFn() func(...)` | Get tool confirmation function. |
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
| `CompressChatAgentContext() (int, error)` | Force compression. |
| `Kind() agents.Kind` | Returns `agents.ChatServer`. |
| `GetName() string` | Returns active agent name. |
| `GetModelID() string` | Returns active agent model ID. |
