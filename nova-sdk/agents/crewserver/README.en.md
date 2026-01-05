# Crew Server Agent

## Description

The **Crew Server Agent** is an HTTP agent that exposes one or more chat agents via a REST API with SSE (Server-Sent Events) streaming. It can manage multiple agents simultaneously and route requests to the appropriate agent based on context.

## Key Features

- **Multi-agent** : Management of a collection of chat agents
- **Intelligent routing** : Automatic selection of the appropriate agent via an orchestrator
- **HTTP/REST API** : Exposure of an API with SSE streaming
- **Tool calls** : Support for function calling with real-time notifications
- **RAG** : Retrieval of relevant context via a RAG agent
- **Compression** : Automatic context compression when limit is reached
- **Human confirmation** : Validation of critical tool calls

## Configuration

### Basic creation with a single agent

```go
crewAgent, err := crewserver.NewAgent(
    ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
)
```

### Creation with multiple agents

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "cook":    cookAgent,
}

crewAgent, err := crewserver.NewAgent(
    ctx,
    crewserver.WithAgentCrew(agentCrew, "coder"), // "coder" is the default agent
    crewserver.WithPort(3500),
)
```

### Available options

- `WithSingleAgent(chatAgent)` - Creates a crew with a single agent
- `WithAgentCrew(agentCrew, selectedAgentId)` - Sets a collection of agents and the default agent
- `WithPort(port)` - Sets the HTTP port (default: 3500)
- `WithToolsAgent(toolsAgent)` - Adds a tools agent for function calling
- `WithRagAgent(ragAgent)` - Adds a RAG agent for context retrieval
- `WithRagAgentAndSimilarityConfig(ragAgent, similarityLimit, maxSimilarities)` - RAG with similarity configuration
- `WithCompressorAgent(compressorAgent)` - Adds a compressor agent
- `WithCompressorAgentAndContextSize(compressorAgent, contextSizeLimit)` - Compressor with size limit
- `WithOrchestratorAgent(orchestratorAgent)` - Adds an orchestrator for automatic routing
- `WithMatchAgentIdToTopicFn(fn)` - Sets the topic -> agent ID mapping function
- `WithExecuteFn(fn)` - Sets the tool execution function

## REST API

### Available routes

#### POST /completion

Generates a completion with SSE streaming.

**Request:**
```json
{
  "data": {
    "message": "Your question here"
  }
}
```

**Response:** SSE stream with events:
```
data: {"message": "response chunk"}
data: {"message": "another chunk"}
data: {"message": "", "finish_reason": "stop"}
```

**Tool notifications:**
```
data: {"kind": "tool_call", "status": "pending", "operation_id": "123", "message": "Calling function X"}
```

#### POST /completion/stop

Stops the current streaming.

**Response:**
```json
{
  "status": "ok",
  "message": "Stream stopped"
}
```

#### POST /memory/reset

Resets the conversation history (keeps system instruction).

**Response:**
```json
{
  "status": "ok",
  "message": "Memory reset successfully"
}
```

#### GET /memory/messages/list

Lists all conversation messages.

**Response:**
```json
{
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ]
}
```

#### GET /memory/messages/context-size

Returns the approximate context size.

**Response:**
```json
{
  "context_size": 1234
}
```

#### POST /operation/validate

Validates a pending operation (human confirmation).

**Request:**
```json
{
  "operation_id": "123"
}
```

**Response:**
```json
{
  "status": "ok",
  "message": "Operation validated"
}
```

#### POST /operation/cancel

Cancels a pending operation.

**Request:**
```json
{
  "operation_id": "123"
}
```

**Response:**
```json
{
  "status": "ok",
  "message": "Operation cancelled"
}
```

#### POST /operation/reset

Resets all pending operations.

**Response:**
```json
{
  "status": "ok",
  "message": "Operations reset successfully"
}
```

#### GET /models

Returns information about the models used.

**Response:**
```json
{
  "chat_model": "ai/qwen2.5:1.5B-F16",
  "tools_model": "hf.co/menlo/jan-nano-gguf:q4_k_m",
  "rag_model": "ai/mxbai-embed-large",
  "compressor_model": "ai/qwen2.5:1.5B-F16"
}
```

#### GET /health

Server health check.

**Response:**
```json
{
  "status": "ok",
  "message": "Server is healthy"
}
```

## Request processing flow

1. **Context compression** (if configured and limit reached)
2. **Tool call detection and execution** (if tools agent configured)
3. **RAG context addition** (if RAG agent configured)
4. **Routing to appropriate agent** (if orchestrator configured)
5. **Completion generation** with SSE streaming
6. **Cleanup** of tool state

## Intelligent routing

With an orchestrator, the crew can automatically detect the topic and route to the specialized agent:

```go
// Topic -> agent ID mapping function
matchAgentFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding", "programming", "development":
        return "coder"
    case "philosophy", "thinking", "ideas":
        return "thinker"
    case "cooking", "recipe", "food":
        return "cook"
    default:
        return "generic"
    }
}

crewAgent, err := crewserver.NewAgent(
    ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.WithMatchAgentIdToTopicFn(matchAgentFn),
)
```

## Special commands

The crew server supports internal commands:

- `[agent-list]` - Lists all available agents
- `[select-agent <id>]` - Manually selects an agent

## Starting the server

```go
if err := crewAgent.StartServer(); err != nil {
    log.Fatal(err)
}
```

The server starts on `http://localhost:3500` (default).
