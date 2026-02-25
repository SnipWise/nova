# Crew Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration (Options)](#3-agent-configuration-options)
4. [Crew Management](#4-crew-management)
5. [StreamCompletion Pipeline](#5-streamcompletion-pipeline)
6. [Intelligent Routing (Orchestrator)](#6-intelligent-routing-orchestrator)
7. [Tool Call Integration](#7-tool-call-integration)
8. [RAG Integration](#8-rag-integration)
9. [Context Compression](#9-context-compression)
10. [Lifecycle Hooks (BeforeCompletion / AfterCompletion)](#10-lifecycle-hooks-beforecompletion--aftercompletion)
11. [Direct Completion Methods](#11-direct-completion-methods)
12. [Conversation Management](#12-conversation-management)
13. [Context Management](#13-context-management)
14. [Legacy Constructor (NewSimpleAgent)](#14-legacy-constructor-newsimpleagent)
15. [API Reference](#15-api-reference)

---

## 1. Introduction

### What is a Crew Agent?

The `crew.CrewAgent` is a high-level composite agent provided by the Nova SDK (`github.com/snipwise/nova`) that manages a **crew of multiple chat agents** and routes between them based on topics. It orchestrates tool calls, RAG context injection, context compression, and intelligent agent routing in a single pipeline.

### When to use a Crew Agent

| Scenario | Recommended agent |
|---|---|
| Multiple specialized agents with topic-based routing | `crew.CrewAgent` |
| Full pipeline: tools + RAG + compression + routing | `crew.CrewAgent` |
| Single agent with tools, RAG, and compression | `crew.CrewAgent` (via `WithSingleAgent`) |
| Simple direct LLM access | `chat.Agent` |

### Key capabilities

- **Multi-agent crew**: Manage multiple `chat.Agent` instances, each specialized for a topic.
- **Intelligent routing**: Automatically route questions to the most appropriate agent via an orchestrator.
- **Full pipeline**: Context compression, tool calls, RAG injection, and streaming completion.
- **Dynamic crew management**: Add or remove agents at runtime.
- **Lifecycle hooks**: Execute custom logic before and after each completion.
- **Functional options pattern**: Configurable via `CrewAgentOption` functions.

---

## 2. Quick Start

### Minimal example with a single agent

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/crew"
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

    crewAgent, _ := crew.NewAgent(ctx,
        crew.WithSingleAgent(chatAgent),
    )

    result, _ := crewAgent.StreamCompletion("Hello!", func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil
    })

    fmt.Println("\nFinish reason:", result.FinishReason)
}
```

### Example with multiple agents

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "generic": genericAgent,
}

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
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
```

---

## 3. Agent Configuration (Options)

Options are passed as variadic arguments to `NewAgent`:

```go
crewAgent, err := crew.NewAgent(ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithToolsAgent(toolsAgent),
    crew.WithRagAgent(ragAgent),
    crew.WithCompressorAgent(compressorAgent),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.BeforeCompletion(beforeFn),
    crew.AfterCompletion(afterFn),
)
```

| Option | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Sets the crew of agents and the initially selected agent. **Mandatory** (or `WithSingleAgent`). |
| `WithSingleAgent(chatAgent)` | Creates a crew with a single agent (ID: `"single"`). **Mandatory** (or `WithAgentCrew`). |
| `WithMatchAgentIdToTopicFn(fn)` | Sets the function mapping detected topics to agent IDs. |
| `WithExecuteFn(fn)` | Sets the function executor for tool calls. |
| `WithConfirmationPromptFn(fn)` | Sets the confirmation prompt for human-in-the-loop tool calls. |
| `WithToolsAgent(toolsAgent)` | Attaches a tools agent for function calling. |
| `WithTasksAgent(tasksAgent)` | Attaches a tasks agent for task planning and orchestration. |
| `WithCompressorAgent(compressorAgent)` | Attaches a compressor agent for context compression. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attaches a compressor with a context size limit. |
| `WithRagAgent(ragAgent)` | Attaches a RAG agent for document retrieval. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attaches a RAG agent with similarity configuration. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attaches an orchestrator agent for topic detection and routing. |
| `BeforeCompletion(fn)` | Sets a hook called before each `StreamCompletion` call. |
| `AfterCompletion(fn)` | Sets a hook called after each `StreamCompletion` call. |

### Default values

| Parameter | Default |
|---|---|
| `similarityLimit` | `0.6` |
| `maxSimilarities` | `3` |
| `contextSizeLimit` | `8000` |

---

## 4. Crew Management

### Static crew (at creation)

```go
crew.WithAgentCrew(map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
}, "coder")
```

### Dynamic crew management

```go
// Add an agent at runtime
err := crewAgent.AddChatAgentToCrew("cook", cookAgent)

// Remove an agent (cannot remove the currently active one)
err := crewAgent.RemoveChatAgentFromCrew("thinker")

// Get all agents
agents := crewAgent.GetChatAgents()

// Replace the entire crew
crewAgent.SetChatAgents(newCrew)
```

### Switching agents manually

```go
// Get currently selected agent
id := crewAgent.GetSelectedAgentId()

// Switch to a different agent
err := crewAgent.SetSelectedAgentId("coder")
```

**Note:** Only one agent is active at a time. `GetName()`, `GetModelID()`, `GetMessages()`, etc. all operate on the currently active agent.

---

## 5. StreamCompletion Pipeline

The `StreamCompletion` method is the main entry point for the crew agent. It orchestrates the full pipeline:

```go
result, err := crewAgent.StreamCompletion(question, func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
})
```

### Pipeline steps

1. **BeforeCompletion hook** (if set)
2. **Context compression** (if compressor agent configured and context exceeds limit)
3. **Tool call detection and execution** (if tools agent configured)
4. **RAG context injection** (if RAG agent configured)
5. **Topic detection and agent routing** (if orchestrator configured)
6. **Streaming completion** with the active agent
7. **AfterCompletion hook** (if set)

---

## 6. Intelligent Routing (Orchestrator)

When an orchestrator agent is attached, the crew agent can automatically route questions to the most appropriate specialized agent.

### Automatic Routing Configuration (Recommended)

When you use `WithOrchestratorAgent`, the crew agent **automatically configures routing** using the orchestrator's `GetAgentForTopic` method. You don't need to provide `WithMatchAgentIdToTopicFn` unless you have custom routing logic.

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

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent), // Auto-configures routing with GetAgentForTopic
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

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    // Override auto-configuration with custom logic
    crew.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
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
2. **If no custom `matchAgentIdToTopicFn` is provided**: The crew agent automatically calls `orchestratorAgent.GetAgentForTopic(topic)` to get the agent ID.
3. **If a custom `matchAgentIdToTopicFn` is provided**: The crew agent uses your custom function instead.
4. The crew agent switches to the matched agent if different from the current one.
5. The completion is generated by the newly selected agent.

### Direct topic detection

```go
agentId, err := crewAgent.DetectTopicThenGetAgentId("Write a Python function")
// agentId = "coder"
```

---

## 7. Tool Call Integration

```go
toolsAgent, _ := tools.NewAgent(ctx, toolsConfig, toolsModelConfig,
    tools.WithTools(myTools),
)

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithToolsAgent(toolsAgent),
    crew.WithExecuteFn(func(name string, args string) (string, error) {
        return `{"result": "ok"}`, nil
    }),
    crew.WithConfirmationPromptFn(func(name string, args string) tools.ConfirmationResponse {
        // Return Confirm, Deny, or Quit
        return tools.Confirm
    }),
)
```

Tool calls are detected and executed during the `StreamCompletion` pipeline. Results are injected into the chat context before generating the final completion.

---

## 8. RAG Integration

```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, ragModelConfig)

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 5),
)
```

During `StreamCompletion`, the crew agent performs a similarity search and injects relevant context into the conversation before generating the completion.

---

## 9. Context Compression

```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, compressorModelConfig,
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
)

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)
```

At the beginning of each `StreamCompletion`, the context is compressed if it exceeds the configured limit.

### Manual compression

```go
// Compress only if over limit
newSize, err := crewAgent.CompressChatAgentContextIfOverLimit()

// Force compression
newSize, err := crewAgent.CompressChatAgentContext()
```

---

## 10. Lifecycle Hooks (BeforeCompletion / AfterCompletion)

Lifecycle hooks allow you to execute custom logic before and after each `StreamCompletion` call. They are configured as `CrewAgentOption` functional options.

### BeforeCompletion

Called before each `StreamCompletion`. The hook receives a reference to the crew agent.

```go
crew.BeforeCompletion(func(a *crew.CrewAgent) {
    fmt.Printf("[BEFORE] Agent: %s\n", a.GetName())
})
```

### AfterCompletion

Called after each `StreamCompletion`. The hook receives a reference to the crew agent.

```go
crew.AfterCompletion(func(a *crew.CrewAgent) {
    fmt.Printf("[AFTER] Agent: %s\n", a.GetName())
})
```

### Hook placement

| Method | Hooks triggered |
|---|---|
| `StreamCompletion` | Yes |
| `GenerateCompletion` | No (delegates to current chat agent) |
| `GenerateStreamCompletion` | No (delegates to current chat agent) |
| `GenerateCompletionWithReasoning` | No (delegates to current chat agent) |
| `GenerateStreamCompletionWithReasoning` | No (delegates to current chat agent) |

The hooks are in `StreamCompletion`, which is the full pipeline method. The `Generate*` methods delegate directly to the currently active `chat.Agent` and do not trigger crew-level hooks.

### Complete example

```go
callCount := 0

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
    crew.BeforeCompletion(func(a *crew.CrewAgent) {
        callCount++
        fmt.Printf("[BEFORE] Call #%d - Agent: %s\n", callCount, a.GetName())
    }),
    crew.AfterCompletion(func(a *crew.CrewAgent) {
        fmt.Printf("[AFTER] Call #%d - Agent: %s\n", callCount, a.GetName())
    }),
)
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. Existing code without hooks continues to work without any changes.

---

## 11. Direct Completion Methods

The crew agent exposes direct completion methods that delegate to the currently active `chat.Agent`:

```go
// Non-streaming
result, err := crewAgent.GenerateCompletion(userMessages)

// Streaming
result, err := crewAgent.GenerateStreamCompletion(userMessages, callback)

// With reasoning
result, err := crewAgent.GenerateCompletionWithReasoning(userMessages)
result, err := crewAgent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCb, responseCb)
```

**Note:** These methods bypass the full pipeline (no compression, no tool calls, no RAG, no routing). They delegate directly to the currently active chat agent. Lifecycle hooks are **not** triggered.

---

## 12. Conversation Management

All conversation methods operate on the **currently active** chat agent:

```go
// Get messages
msgs := crewAgent.GetMessages()

// Get context size
tokens := crewAgent.GetContextSize()

// Reset conversation
crewAgent.ResetMessages()

// Add a message
crewAgent.AddMessage(roles.User, "Hello")

// Export to JSON
jsonStr, err := crewAgent.ExportMessagesToJSON()

// Stop streaming
crewAgent.StopStream()
```

---

## 13. Context Management

```go
ctx := crewAgent.GetContext()
crewAgent.SetContext(newCtx)
```

---

## 14. Legacy Constructor (NewSimpleAgent)

A simplified constructor is available for backward compatibility:

```go
crewAgent, err := crew.NewSimpleAgent(ctx, agentCrew, "generic")
```

**Note:** `NewSimpleAgent` does not support options (tools, RAG, compressor, orchestrator, hooks). Use `NewAgent` with options for full functionality.

---

## 15. API Reference

### Constructor

```go
func NewAgent(ctx context.Context, options ...CrewAgentOption) (*CrewAgent, error)
func NewSimpleAgent(ctx context.Context, agentCrew map[string]*chat.Agent, selectedAgentId string) (*CrewAgent, error)
```

### Types

```go
type CrewAgentOption func(*CrewAgent) error
```

### Option Functions

| Function | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Sets the crew and initial agent. |
| `WithSingleAgent(chatAgent)` | Creates a single-agent crew. |
| `WithMatchAgentIdToTopicFn(fn)` | Sets topic-to-agent mapping function. |
| `WithExecuteFn(fn)` | Sets tool execution function. |
| `WithConfirmationPromptFn(fn)` | Sets tool confirmation function. |
| `WithToolsAgent(toolsAgent)` | Attaches tools agent. |
| `WithTasksAgent(tasksAgent)` | Attaches tasks agent for planning and orchestration. |
| `WithCompressorAgent(compressorAgent)` | Attaches compressor agent. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attaches compressor with limit. |
| `WithRagAgent(ragAgent)` | Attaches RAG agent. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attaches RAG with config. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attaches orchestrator agent. |
| `BeforeCompletion(fn func(*CrewAgent))` | Sets hook before each StreamCompletion. |
| `AfterCompletion(fn func(*CrewAgent))` | Sets hook after each StreamCompletion. |

### Methods

| Method | Description |
|---|---|
| `StreamCompletion(question, callback) (*chat.CompletionResult, error)` | Full pipeline completion with streaming. |
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
| `SetConfirmationPromptFunction(fn)` | Set tool confirmation function. |
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
| `Kind() agents.Kind` | Returns `agents.Composite`. |
| `GetName() string` | Returns active agent name. |
| `GetModelID() string` | Returns active agent model ID. |
| `GetContext() context.Context` | Get agent context. |
| `SetContext(ctx)` | Set agent context. |
