# CrewServerAgent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Architecture Overview](#2-architecture-overview)
3. [Quick Start: Single Agent](#3-quick-start-single-agent)
4. [Multi-Agent Crew with Orchestrator](#4-multi-agent-crew-with-orchestrator)
5. [Adding Tools (Function Calling)](#5-adding-tools-function-calling)
6. [Parallel Tool Calls](#6-parallel-tool-calls)
7. [Confirmation Workflow (Human-in-the-Loop)](#7-confirmation-workflow-human-in-the-loop)
8. [RAG Integration](#8-rag-integration)
9. [Context Compression](#9-context-compression)
10. [Complete Example](#10-complete-example)
11. [API Reference](#11-api-reference)
12. [Configuration Reference](#12-configuration-reference)

---

## 1. Introduction

### What is CrewServerAgent?

`CrewServerAgent` is an HTTP server agent provided by the Nova SDK (`github.com/snipwise/nova`). It exposes a set of REST/SSE endpoints that orchestrate one or more AI chat agents behind a single HTTP interface. Clients interact with the server through standard HTTP requests and receive streaming responses via Server-Sent Events (SSE).

Unlike a simple `ServerAgent` that wraps a single chat agent, `CrewServerAgent` can manage an entire **crew** of specialized chat agents and dynamically route incoming requests to the most appropriate one based on topic detection. It also integrates optional components -- tools (function calling), RAG (Retrieval-Augmented Generation), and context compression -- into a unified request pipeline.

### When to use CrewServerAgent (vs. simple ServerAgent)

| Scenario | Recommended agent |
|---|---|
| Single chat agent exposed as an API | `ServerAgent` |
| Multiple specialized agents with routing | `CrewServerAgent` |
| Agent API with function calling (tools) | `CrewServerAgent` |
| Agent API with RAG document retrieval | `CrewServerAgent` |
| Agent API with context compression | `CrewServerAgent` |
| Full-featured agent backend (tools + RAG + compression + routing) | `CrewServerAgent` |

### Key capabilities

- **Multi-agent crew**: Register multiple `chat.Agent` instances and route between them at runtime.
- **Orchestrator-based routing**: Automatically detect the topic of a user query and forward it to the most relevant agent.
- **Function calling (tools)**: Detect tool calls in user queries, execute them, and inject results into the conversation.
- **Parallel tool calls**: Execute multiple tool calls detected in a single pass, with optional confirmation.
- **Human-in-the-loop confirmation**: Web-based or custom confirmation workflows before tool execution.
- **RAG integration**: Perform similarity search against a vector store and inject relevant context before completion.
- **Context compression**: Automatically compress conversation history when it exceeds a configurable size limit.
- **SSE streaming**: All completion responses are streamed to the client as Server-Sent Events.
- **CORS support**: Built-in CORS middleware for browser-based frontends.

---

## 2. Architecture Overview

### Components

A `CrewServerAgent` can be composed of the following components, all of which are optional except for at least one chat agent:

| Component | Type | Purpose |
|---|---|---|
| **Chat Agents** | `map[string]*chat.Agent` | One or more specialized chat agents that generate completions. |
| **Tools Agent** | `*tools.Agent` | Detects tool calls (function calling) in user messages and invokes them. |
| **RAG Agent** | `*rag.Agent` | Searches a vector store for relevant documents and injects them as context. |
| **Compressor Agent** | `*compressor.Agent` | Compresses conversation history when context size exceeds a limit. |
| **Orchestrator Agent** | `agents.OrchestratorAgent` | Detects the topic/intent of a user query for routing to the appropriate chat agent. |
| **Execute Function** | `func(string, string) (string, error)` | User-defined function that performs the actual work when a tool call is detected. |
| **Match Function** | `func(string, string) string` | Maps a detected topic to an agent ID, enabling dynamic agent switching. |

### Request flow

When a `POST /completion` request arrives, the server processes it through the following pipeline:

```
HTTP Request (POST /completion)
       |
       v
  1. Parse request body (extract user message)
       |
       v
  2. Setup SSE streaming headers
       |
       v
  3. Compress context if needed
     (sends SSE notification to client)
       |
       v
  4. Setup notification channel for tool calls
       |
       v
  5. Tool call detection & execution
     - Detect tool calls in user message + history
     - If confirmation required: send SSE notification, wait for /operation/validate or /operation/cancel
     - Execute tool functions
     - Inject results into chat context
       |
       v
  6. Close notification channel
       |
       v
  7. Generate streaming completion (if needed)
     a. Add RAG context (similarity search)
     b. Route to appropriate agent (orchestrator topic detection)
     c. Stream completion via SSE
       |
       v
  8. Cleanup tool state
```

The streaming completion step (7) is skipped if a tool call was successfully executed and returned results, since the tool results are already streamed to the client. It proceeds when no tool calls were detected, or when the user denied/quit the confirmation prompt.

---

## 3. Quick Start: Single Agent

The simplest way to use `CrewServerAgent` is with a single chat agent. This is equivalent to a basic server agent, but gives you access to the full crew infrastructure for future expansion.

### Minimal example

```go
package main

import (
    "context"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Create a chat agent
    chatAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "assistant",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a helpful assistant.",
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.7),
        },
    )
    if err != nil {
        log.Fatalf("Failed to create chat agent: %v", err)
    }

    // Create the crew server agent with a single agent
    crewServerAgent, err := crewserver.NewAgent(
        ctx,
        crewserver.WithSingleAgent(chatAgent),
        crewserver.WithPort(3500),
    )
    if err != nil {
        log.Fatalf("Failed to create crew server agent: %v", err)
    }

    // Start the HTTP server
    log.Println("Server starting on http://localhost:3500")
    if err := crewServerAgent.StartServer(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

### Testing with curl

Send a completion request:

```bash
curl -N -X POST http://localhost:3500/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "What is the capital of France?"}}'
```

The response is streamed as SSE events. Each event has this format:

```
data: {"message":"Paris"}

data: {"message":" is the"}

data: {"message":" capital of France."}

data: {"message":"","finish_reason":"stop"}
```

Check server health:

```bash
curl http://localhost:3500/health
# {"status":"ok"}
```

Reset conversation memory:

```bash
curl -X POST http://localhost:3500/memory/reset
# {"status":"ok","message":"Memory reset successfully"}
```

---

## 4. Multi-Agent Crew with Orchestrator

The real power of `CrewServerAgent` emerges when you register multiple specialized chat agents and use an orchestrator to route queries dynamically.

### Defining multiple chat agents

Each chat agent is created independently with its own system instructions, model, and configuration:

```go
// Create a coding expert agent
coderAgent, _ := chat.NewAgent(ctx,
    agents.Config{
        Name:               "coder",
        EngineURL:          engineURL,
        SystemInstructions: "You are an expert Go programmer...",
        KeepConversationHistory: true,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.3),
    },
)

// Create a thinking/reasoning agent
thinkerAgent, _ := chat.NewAgent(ctx,
    agents.Config{
        Name:               "thinker",
        EngineURL:          engineURL,
        SystemInstructions: "You are a deep thinker who excels at analysis...",
        KeepConversationHistory: true,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.9),
    },
)

// Create a general expert agent
expertAgent, _ := chat.NewAgent(ctx,
    agents.Config{
        Name:               "expert",
        EngineURL:          engineURL,
        SystemInstructions: "You are a knowledgeable expert assistant...",
        KeepConversationHistory: true,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.7),
    },
)
```

### Registering the crew with WithAgentCrew

Place all agents into a map keyed by their ID, then pass it to `WithAgentCrew` along with the ID of the default agent:

```go
agentCrew := map[string]*chat.Agent{
    "expert":  expertAgent,
    "thinker": thinkerAgent,
    "coder":   coderAgent,
}
```

### Creating an orchestrator agent

The orchestrator is a lightweight agent that classifies user queries by topic. It uses the `orchestrator` package:

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
)

orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:               "orchestrator-agent",
        EngineURL:          engineURL,
        SystemInstructions: "Classify the user query into one of these topics: code_generation, complex_thinking, code_question.",
    },
    models.Config{
        Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
        Temperature: models.Float64(0.0),
    },
)
```

### Routing with WithMatchAgentIdToTopicFn

The match function receives the current agent ID and the detected topic, and returns the agent ID that should handle the query. This is also a good place to transfer conversation history between agents when switching:

```go
matchFn := func(currentAgentId, topic string) string {
    var agentId string
    switch strings.ToLower(topic) {
    case "code_generation", "write code", "create code":
        agentId = "coder"
    case "complex_thinking", "reasoning", "analysis":
        agentId = "thinker"
    case "code_question", "explain", "how to":
        agentId = "expert"
    default:
        agentId = "expert"
    }

    // Transfer conversation history when switching agents
    if agentId != currentAgentId {
        history := agentCrew[currentAgentId].GetMessages()
        agentCrew[agentId].AddMessages(history)
    }

    return agentId
}
```

### Assembling the crew server

```go
crewServerAgent, err := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "expert"),
    crewserver.WithPort(3500),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.WithMatchAgentIdToTopicFn(matchFn),
)
if err != nil {
    log.Fatalf("Failed to create crew server: %v", err)
}

crewServerAgent.StartServer()
```

### Checking the current agent

You can query which agent is currently active:

```bash
curl http://localhost:3500/current-agent
# {"agent_id":"expert","model_id":"ai/qwen2.5:1.5B-F16","agent_name":"expert"}
```

---

## 5. Adding Tools (Function Calling)

Tools enable your agent to detect when a user query requires calling a function and to execute that function automatically.

### Creating a tools agent

Define your tools using `tools.NewTool`, then create a `tools.Agent`:

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents/tools"
)

// Define tools
calculateSum := tools.NewTool("calculate_sum").
    SetDescription("Calculate the sum of two numbers").
    AddParameter("a", "number", "The first number", true).
    AddParameter("b", "number", "The second number", true)

saveFile := tools.NewTool("save_snippet").
    SetDescription("Save snippet content to a file").
    AddParameter("file_path", "string", "The file path", true).
    AddParameter("content", "string", "The content to write", true)

toolsList := []*tools.Tool{calculateSum, saveFile}

// Create the tools agent
toolsAgent, err := tools.NewAgent(ctx,
    agents.Config{
        Name:               "tools-agent",
        EngineURL:          engineURL,
        SystemInstructions: "You detect tool calls from user queries.",
        KeepConversationHistory: false,
    },
    models.Config{
        Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature:       models.Float64(0.0),
        ParallelToolCalls: models.Bool(false),
    },
    tools.WithTools(toolsList),
)
```

### Defining the execute function

The execute function receives the function name and a JSON string of arguments, and returns a JSON result:

```go
func executeFunction(functionName string, arguments string) (string, error) {
    switch functionName {
    case "calculate_sum":
        var args struct {
            A float64 `json:"a"`
            B float64 `json:"b"`
        }
        if err := json.Unmarshal([]byte(arguments), &args); err != nil {
            return `{"error": "Invalid arguments"}`, err
        }
        return fmt.Sprintf(`{"result": %g}`, args.A+args.B), nil

    case "save_snippet":
        var args struct {
            FilePath string `json:"file_path"`
            Content  string `json:"content"`
        }
        if err := json.Unmarshal([]byte(arguments), &args); err != nil {
            return `{"error": "Invalid arguments"}`, err
        }
        // Perform file write here...
        return `{"message": "file saved"}`, nil

    default:
        return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
    }
}
```

### Attaching tools to the crew server

```go
crewServerAgent, err := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(executeFunction),
)
```

### Tool execution flow

When a completion request arrives and a tools agent is configured:

1. The tools agent receives the user message along with the current chat history.
2. It determines whether the message requires a tool call.
3. If a tool call is detected, the execute function is invoked.
4. The result is injected into the chat agent's context as a system message.
5. The result is also streamed to the client via SSE.
6. If no tool call is detected (or the user denied confirmation), the request proceeds to the standard streaming completion.

### Default behavior

When no `WithConfirmationPromptFn` is provided, the server uses the built-in **web confirmation prompt** (`webConfirmationPrompt`). This means every tool call triggers a notification via SSE and waits for the client to call `/operation/validate` or `/operation/cancel` before proceeding.

---

## 6. Parallel Tool Calls

By default, tool calls are detected in a loop -- the agent sends the query, detects one tool call, executes it, and loops back to check for more. With **parallel tool calls**, the model can detect multiple tool calls in a single inference pass and execute them all at once.

### Enabling parallel tool calls

Set `ParallelToolCalls: models.Bool(true)` on the tools agent's model configuration:

```go
toolsAgent, _ := tools.NewAgent(ctx,
    agents.Config{
        Name:               "tools-agent",
        EngineURL:          engineURL,
        SystemInstructions: "You detect tool calls from user queries.",
        KeepConversationHistory: false,
    },
    models.Config{
        Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature:       models.Float64(0.0),
        ParallelToolCalls: models.Bool(true),
    },
    tools.WithTools(toolsList),
)
```

### Parallel vs. loop detection

| Mode | Behavior |
|---|---|
| **Loop** (`ParallelToolCalls = false` or `nil`) | Detects one tool call at a time, executes it, then loops to detect more. Continues until no more tool calls are found or the user quits. |
| **Parallel** (`ParallelToolCalls = true`) | Detects all tool calls in a single inference pass and executes them all. Single-pass, no loop. |

### Branching logic

The `CrewServerAgent` automatically selects the correct detection method based on the combination of `ParallelToolCalls` and `ConfirmationPromptFn`:

| ParallelToolCalls | ConfirmationPromptFn | Method used |
|---|---|---|
| `true` | Not set | `DetectParallelToolCalls` |
| `true` | Set (custom) | `DetectParallelToolCallsWithConfirmation` |
| `false` or `nil` | Not set | `DetectToolCallsLoopWithConfirmation` (web-based) |
| `false` or `nil` | Set (custom) | `DetectToolCallsLoopWithConfirmation` (custom) |

### Example: parallel without confirmation

```go
toolsAgent, _ := tools.NewAgent(ctx,
    agents.Config{
        Name:               "tools-agent",
        EngineURL:          engineURL,
        SystemInstructions: "Detect and execute tool calls.",
        KeepConversationHistory: false,
    },
    models.Config{
        Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature:       models.Float64(0.0),
        ParallelToolCalls: models.Bool(true),
    },
    tools.WithTools(toolsList),
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(executeFunction),
    // No WithConfirmationPromptFn => uses DetectParallelToolCalls
)
```

### Example: parallel with confirmation

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(executeFunction),
    crewserver.WithConfirmationPromptFn(myConfirmationFn),
    // ParallelToolCalls=true + confirmation => uses DetectParallelToolCallsWithConfirmation
)
```

---

## 7. Confirmation Workflow (Human-in-the-Loop)

The confirmation workflow allows a human to approve or deny tool executions before they run. This is essential for safety-critical tools (file writes, API calls, database mutations).

### Web-based confirmation (default)

When no `WithConfirmationPromptFn` is provided, the server uses the built-in web confirmation mechanism:

1. A tool call is detected.
2. The server sends an SSE notification to the client with the operation ID, function name, and arguments:
   ```json
   {"kind": "tool_call", "status": "pending", "operation_id": "op_xxx", "message": "Tool call detected: calculate_sum"}
   ```
3. The server **blocks** and waits for the client to respond.
4. The client calls one of these endpoints:
   - `POST /operation/validate` with `{"operation_id": "op_xxx"}` to approve.
   - `POST /operation/cancel` with `{"operation_id": "op_xxx"}` to deny.
   - `POST /operation/reset` to cancel all pending operations.

### Custom confirmation function

You can provide your own confirmation logic using `WithConfirmationPromptFn`. The function receives the function name and arguments, and returns a `tools.ConfirmationResponse`:

```go
import "github.com/snipwise/nova/nova-sdk/agents/tools"

customConfirmation := func(functionName string, arguments string) tools.ConfirmationResponse {
    // Your custom logic here
    // For example, auto-approve safe functions:
    if functionName == "calculate_sum" {
        return tools.Confirmed
    }
    // Deny unknown functions:
    return tools.Denied
}

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(executeFunction),
    crewserver.WithConfirmationPromptFn(customConfirmation),
)
```

The `tools.ConfirmationResponse` type supports the following values:

| Value | Meaning |
|---|---|
| `tools.Confirmed` | Approve the tool call; execution proceeds. |
| `tools.Denied` | Deny the tool call; the loop continues to check for more. |
| `tools.Quit` | Quit the tool call loop entirely. |

### Combining with parallel tool calls

When `ParallelToolCalls` is enabled and a `ConfirmationPromptFn` is provided, the server uses `DetectParallelToolCallsWithConfirmation`. Each detected tool call in the parallel batch is passed through the confirmation function before execution.

### Operation endpoints

These endpoints are used by the web-based confirmation workflow:

| Method | Path | Description |
|---|---|---|
| `POST` | `/operation/validate` | Approve a pending tool call operation. |
| `POST` | `/operation/cancel` | Deny a pending tool call operation. |
| `POST` | `/operation/reset` | Cancel all pending operations (sends `Quit` to all). |

Request body for `/operation/validate` and `/operation/cancel`:

```json
{
    "operation_id": "op_xxx"
}
```

---

## 8. RAG Integration

RAG (Retrieval-Augmented Generation) allows the crew server to enrich the chat context with relevant documents retrieved from a vector store before generating a completion.

### Creating a RAG agent

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents/rag"
)

ragAgent, err := rag.NewAgent(ctx,
    agents.Config{
        Name:      "rag-agent",
        EngineURL: engineURL,
    },
    models.Config{
        Name: "ai/mxbai-embed-large",
    },
)
```

You will typically load documents into the RAG agent's vector store before starting the server. See the Nova SDK documentation for `rag.Agent` for details on document ingestion.

### Attaching RAG to the crew server

**Basic attachment** (uses default similarity settings):

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.WithRagAgent(ragAgent),
)
```

**With custom similarity configuration**:

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 7),
    // similarityLimit = 0.4 (minimum cosine similarity)
    // maxSimilarities = 7   (maximum number of documents to retrieve)
)
```

### Default similarity settings

| Setting | Default | Description |
|---|---|---|
| `SimilarityLimit` | `0.6` | Minimum cosine similarity score for a document to be included. |
| `MaxSimilarities` | `3` | Maximum number of similar documents to retrieve. |

### How RAG context is injected

During the completion pipeline (step 7a), after tool calls have been processed:

1. The user's question is used as the search query.
2. The RAG agent performs a similarity search: `ragAgent.SearchTopN(question, similarityLimit, maxSimilarities)`.
3. All matching documents are concatenated and injected into the chat agent's context as a system message:
   ```
   Relevant information to help you answer the question:
   <document 1>
   ---
   <document 2>
   ---
   ...
   ```
4. The chat agent then generates its completion with the enriched context.

---

## 9. Context Compression

Long conversations accumulate large context windows. The compressor agent automatically compresses conversation history when it exceeds a configurable size limit, keeping the context manageable without losing important information.

### Creating a compressor agent

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
)

compressorAgent, err := compressor.NewAgent(ctx,
    agents.Config{
        Name:               "compressor-agent",
        EngineURL:          engineURL,
        SystemInstructions: compressor.Instructions.Effective,
    },
    models.Config{
        Name:        "ai/qwen2.5:0.5B-F16",
        Temperature: models.Float64(0.0),
    },
    compressor.WithCompressionPrompt(compressor.Prompts.UltraShort),
)
```

### Attaching the compressor to the crew server

**Basic attachment** (uses default context size limit of 8000 characters):

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.WithCompressorAgent(compressorAgent),
)
```

**With custom context size limit**:

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.WithCompressorAgentAndContextSize(compressorAgent, 32000),
)
```

### Automatic compression with SSE notifications

Compression is triggered automatically at the beginning of each completion request (step 3 in the pipeline). When the current context size exceeds the limit:

1. The server sends an SSE notification to the client:
   ```json
   {"role": "information", "content": "Context size limit reached. Compressing conversation history..."}
   ```
2. The compressor agent compresses all messages into a summary.
3. The chat agent's messages are reset and replaced with the compressed summary as a system message.
4. A success notification is sent:
   ```json
   {"role": "information", "content": "Compression completed. Context reduced from 35000 to 4200 bytes."}
   ```

If the compressed context still exceeds 80% of the limit, an error is reported. If it exceeds 90%, the messages are fully reset as a safety measure.

### Monitoring context size

Use the context size endpoint to check the current state:

```bash
curl http://localhost:3500/memory/messages/context-size
# {"messages_count": 12, "characters_count": 5400, "limit": 32000}
```

---

## 10. Complete Example

The following example combines all features: multi-agent crew, orchestrator routing, tools with parallel execution, RAG document retrieval, and context compression. It mirrors the structure of sample 71 in the Nova repository.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/models"
)

var agentCrew map[string]*chat.Agent

func main() {
    ctx := context.Background()
    engineURL := "http://localhost:12434/engines/llama.cpp/v1"

    // -------------------------------------------------------
    // 1. Create specialized chat agents
    // -------------------------------------------------------
    expertAgent, _ := chat.NewAgent(ctx,
        agents.Config{
            Name:                   "expert",
            EngineURL:              engineURL,
            SystemInstructions:     "You are a knowledgeable expert assistant.",
            KeepConversationHistory: true,
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.7),
        },
    )

    coderAgent, _ := chat.NewAgent(ctx,
        agents.Config{
            Name:                   "coder",
            EngineURL:              engineURL,
            SystemInstructions:     "You are an expert Go programmer.",
            KeepConversationHistory: true,
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.3),
        },
    )

    thinkerAgent, _ := chat.NewAgent(ctx,
        agents.Config{
            Name:                   "thinker",
            EngineURL:              engineURL,
            SystemInstructions:     "You are a deep thinker who excels at analysis and reasoning.",
            KeepConversationHistory: true,
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.9),
        },
    )

    agentCrew = map[string]*chat.Agent{
        "expert":  expertAgent,
        "coder":   coderAgent,
        "thinker": thinkerAgent,
    }

    // -------------------------------------------------------
    // 2. Create orchestrator agent for topic detection
    // -------------------------------------------------------
    orchestratorAgent, _ := orchestrator.NewAgent(ctx,
        agents.Config{
            Name:               "orchestrator-agent",
            EngineURL:          engineURL,
            SystemInstructions: "Classify queries into: code_generation, complex_thinking, code_question.",
        },
        models.Config{
            Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
            Temperature: models.Float64(0.0),
        },
    )

    // -------------------------------------------------------
    // 3. Create tools agent
    // -------------------------------------------------------
    calculateSum := tools.NewTool("calculate_sum").
        SetDescription("Calculate the sum of two numbers").
        AddParameter("a", "number", "The first number", true).
        AddParameter("b", "number", "The second number", true)

    toolsAgent, _ := tools.NewAgent(ctx,
        agents.Config{
            Name:                   "tools-agent",
            EngineURL:              engineURL,
            SystemInstructions:     "Detect tool calls from user queries.",
            KeepConversationHistory: false,
        },
        models.Config{
            Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature:       models.Float64(0.0),
            ParallelToolCalls: models.Bool(true),
        },
        tools.WithTools([]*tools.Tool{calculateSum}),
    )

    // -------------------------------------------------------
    // 4. Create RAG agent
    // -------------------------------------------------------
    ragAgent, _ := rag.NewAgent(ctx,
        agents.Config{
            Name:      "rag-agent",
            EngineURL: engineURL,
        },
        models.Config{
            Name: "ai/mxbai-embed-large",
        },
    )
    // Load documents into ragAgent here...

    // -------------------------------------------------------
    // 5. Create compressor agent
    // -------------------------------------------------------
    compressorAgent, _ := compressor.NewAgent(ctx,
        agents.Config{
            Name:               "compressor-agent",
            EngineURL:          engineURL,
            SystemInstructions: compressor.Instructions.Effective,
        },
        models.Config{
            Name:        "ai/qwen2.5:0.5B-F16",
            Temperature: models.Float64(0.0),
        },
        compressor.WithCompressionPrompt(compressor.Prompts.UltraShort),
    )

    // -------------------------------------------------------
    // 6. Define routing function
    // -------------------------------------------------------
    matchFn := func(currentAgentId, topic string) string {
        var agentId string
        switch strings.ToLower(topic) {
        case "code_generation", "write code":
            agentId = "coder"
        case "complex_thinking", "reasoning", "analysis":
            agentId = "thinker"
        default:
            agentId = "expert"
        }

        // Transfer history when switching agents
        if agentId != currentAgentId {
            history := agentCrew[currentAgentId].GetMessages()
            agentCrew[agentId].AddMessages(history)
        }
        return agentId
    }

    // -------------------------------------------------------
    // 7. Define tool execution function
    // -------------------------------------------------------
    executeFn := func(functionName string, arguments string) (string, error) {
        switch functionName {
        case "calculate_sum":
            var args struct {
                A float64 `json:"a"`
                B float64 `json:"b"`
            }
            if err := json.Unmarshal([]byte(arguments), &args); err != nil {
                return `{"error": "Invalid arguments"}`, err
            }
            return fmt.Sprintf(`{"result": %g}`, args.A+args.B), nil
        default:
            return `{"error": "Unknown function"}`, fmt.Errorf("unknown: %s", functionName)
        }
    }

    // -------------------------------------------------------
    // 8. Assemble the crew server agent
    // -------------------------------------------------------
    crewServerAgent, err := crewserver.NewAgent(ctx,
        crewserver.WithAgentCrew(agentCrew, "expert"),
        crewserver.WithPort(3500),
        crewserver.WithOrchestratorAgent(orchestratorAgent),
        crewserver.WithMatchAgentIdToTopicFn(matchFn),
        crewserver.WithToolsAgent(toolsAgent),
        crewserver.WithExecuteFn(executeFn),
        crewserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 7),
        crewserver.WithCompressorAgentAndContextSize(compressorAgent, 32000),
    )
    if err != nil {
        log.Fatalf("Failed to create crew server: %v", err)
    }

    // -------------------------------------------------------
    // 9. Start the server
    // -------------------------------------------------------
    log.Println("Crew server starting on http://localhost:3500")
    if err := crewServerAgent.StartServer(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

---

## 11. API Reference

All endpoints are served with CORS headers enabled by default (`Access-Control-Allow-Origin: *`).

### POST /completion

Send a user message and receive a streaming completion via SSE.

**Request:**

```json
{
    "data": {
        "message": "Your question here"
    }
}
```

**Response:** Server-Sent Events stream.

Each SSE event is a `data:` line containing JSON:

| Event type | JSON fields | Description |
|---|---|---|
| Content chunk | `{"message": "..."}` | A chunk of the completion text. |
| Finish | `{"message": "", "finish_reason": "stop"}` | Signals the end of the completion. |
| Error | `{"error": "..."}` | An error occurred during completion. |
| Information | `{"role": "information", "content": "..."}` | Informational notification (e.g., compression status). |
| Tool notification | `{"kind": "tool_call", "status": "pending", "operation_id": "...", "message": "..."}` | A tool call is pending confirmation. |

---

### POST /completion/stop

Interrupt the current streaming completion.

**Request:** Empty body.

**Response:**

```json
{
    "status": "ok",
    "message": "Stream stopped"
}
```

If no stream is active:

```json
{
    "status": "ok",
    "message": "No stream to stop"
}
```

---

### POST /memory/reset

Clear all conversation messages (except the system instruction) for the current chat agent and the tools agent.

**Request:** Empty body.

**Response:**

```json
{
    "status": "ok",
    "message": "Memory reset successfully"
}
```

---

### GET /memory/messages/list

Retrieve all messages in the current chat agent's conversation history.

**Response:**

```json
{
    "messages": [
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
}
```

---

### GET /memory/messages/context-size

Get the current context size information.

**Response:**

```json
{
    "messages_count": 5,
    "characters_count": 1200,
    "limit": 32000
}
```

---

### POST /operation/validate

Approve a pending tool call operation (used in the web-based confirmation workflow).

**Request:**

```json
{
    "operation_id": "op_xxx"
}
```

**Response:** SSE stream with a confirmation message:

```
data: {"message":"Operation op_xxx validated"}
```

---

### POST /operation/cancel

Deny a pending tool call operation.

**Request:**

```json
{
    "operation_id": "op_xxx"
}
```

**Response:** SSE stream with a cancellation message:

```
data: {"message":"Operation op_xxx cancelled"}
```

---

### POST /operation/reset

Cancel all pending tool call operations. Sends a `Quit` signal to every pending operation.

**Request:** Empty body.

**Response:** SSE stream with a reset message:

```
data: {"message":"All pending operations cancelled (N operations)"}
```

---

### GET /models

Get information about the models used by the server.

**Response:**

```json
{
    "status": "ok",
    "chat_model": "ai/qwen2.5:1.5B-F16",
    "embeddings_model": "ai/mxbai-embed-large",
    "tools_model": "hf.co/menlo/jan-nano-gguf:q4_k_m"
}
```

If a component is not configured, its model is reported as `"none"`.

---

### GET /health

Health check endpoint.

**Response:**

```json
{
    "status": "ok"
}
```

---

### GET /current-agent

Get information about the currently active chat agent.

**Response:**

```json
{
    "agent_id": "expert",
    "model_id": "ai/qwen2.5:1.5B-F16",
    "agent_name": "expert"
}
```

---

## 12. Configuration Reference

All configuration is done through functional options passed to `crewserver.NewAgent(ctx, options...)`.

### WithAgentCrew

```go
func WithAgentCrew(agentCrew map[string]*chat.Agent, selectedAgentId string) CrewServerAgentOption
```

Sets the crew of chat agents and the initially selected agent ID. The `agentCrew` map must not be empty, and `selectedAgentId` must correspond to a key in the map.

**Required:** Either `WithAgentCrew` or `WithSingleAgent` must be provided.

---

### WithSingleAgent

```go
func WithSingleAgent(chatAgent *chat.Agent) CrewServerAgentOption
```

Convenience option that creates a crew with a single agent registered under the key `"single"`. Equivalent to:

```go
WithAgentCrew(map[string]*chat.Agent{"single": chatAgent}, "single")
```

**Required:** Either `WithAgentCrew` or `WithSingleAgent` must be provided.

---

### WithPort

```go
func WithPort(port int) CrewServerAgentOption
```

Sets the HTTP server port. The port is specified as an integer (e.g., `3500`).

**Default:** `3500`

---

### WithOrchestratorAgent

```go
func WithOrchestratorAgent(orchestratorAgent agents.OrchestratorAgent) CrewServerAgentOption
```

Attaches an orchestrator agent for topic detection and query routing. The orchestrator classifies user queries by topic, and the result is passed to the match function (see `WithMatchAgentIdToTopicFn`) to determine which chat agent should handle the request.

**Default:** `nil` (no routing; all queries go to the current agent).

---

### WithMatchAgentIdToTopicFn

```go
func WithMatchAgentIdToTopicFn(fn func(currentAgentId string, topic string) string) CrewServerAgentOption
```

Sets the function that maps a detected topic to a chat agent ID. The function receives:

- `currentAgentId`: The ID of the currently active agent.
- `topic`: The topic detected by the orchestrator.

It must return a valid agent ID from the crew map.

**Default:** Returns the first agent ID found in the crew map (non-deterministic).

---

### WithToolsAgent

```go
func WithToolsAgent(toolsAgent *tools.Agent) CrewServerAgentOption
```

Attaches a tools agent for function calling capabilities. When set, every completion request first passes through tool call detection.

**Default:** `nil` (no tool call detection).

---

### WithExecuteFn

```go
func WithExecuteFn(fn func(functionName string, arguments string) (string, error)) CrewServerAgentOption
```

Sets the function that executes detected tool calls. The function receives:

- `functionName`: The name of the function to execute.
- `arguments`: A JSON string containing the function arguments.

It must return:

- A JSON string with the result.
- An error if execution failed.

**Default:** A placeholder function that returns an error indicating the function is not implemented.

---

### WithConfirmationPromptFn

```go
func WithConfirmationPromptFn(fn func(functionName string, arguments string) tools.ConfirmationResponse) CrewServerAgentOption
```

Sets a custom confirmation prompt function for tool call confirmation. When provided, this function is called instead of the default web-based confirmation prompt.

When combined with `ParallelToolCalls` enabled on the tools agent, `DetectParallelToolCallsWithConfirmation` is used instead of `DetectParallelToolCalls`.

**Default:** `nil` (uses the built-in web-based confirmation via SSE notifications and `/operation/validate`, `/operation/cancel` endpoints).

---

### WithCompressorAgent

```go
func WithCompressorAgent(compressorAgent *compressor.Agent) CrewServerAgentOption
```

Attaches a compressor agent for automatic context compression. Uses the default context size limit (8000 characters).

**Default:** `nil` (no compression).

---

### WithCompressorAgentAndContextSize

```go
func WithCompressorAgentAndContextSize(compressorAgent *compressor.Agent, contextSizeLimit int) CrewServerAgentOption
```

Attaches a compressor agent and sets a custom context size limit (in characters). When the conversation history exceeds this limit, compression is triggered automatically before processing the next completion request.

**Default:** `nil` compressor, `8000` characters limit.

---

### WithRagAgent

```go
func WithRagAgent(ragAgent *rag.Agent) CrewServerAgentOption
```

Attaches a RAG agent for document retrieval. Uses the default similarity settings (`similarityLimit = 0.6`, `maxSimilarities = 3`).

**Default:** `nil` (no RAG).

---

### WithRagAgentAndSimilarityConfig

```go
func WithRagAgentAndSimilarityConfig(ragAgent *rag.Agent, similarityLimit float64, maxSimilarities int) CrewServerAgentOption
```

Attaches a RAG agent and configures similarity search settings:

- `similarityLimit`: Minimum cosine similarity score (0.0 to 1.0). Documents below this threshold are excluded.
- `maxSimilarities`: Maximum number of similar documents to retrieve.

**Default:** `nil` RAG agent, `0.6` similarity limit, `3` max similarities.
