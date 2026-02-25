# Server Agent

## Description

The **Server Agent** is a chat agent that exposes an HTTP/REST API with SSE (Server-Sent Events) streaming. It wraps a `chat.Agent` and can be enhanced with auxiliary agents (Tools, RAG, Compressor) for advanced features.

## Features

- **HTTP/REST API**: Exposes endpoints to interact with the agent via HTTP
- **SSE Streaming**: Real-time responses via Server-Sent Events
- **Tools Agent**: Function calling with user confirmation
- **RAG Agent**: Similarity search and context enrichment
- **Compressor Agent**: Automatic context compression when limit is reached
- **Human-in-the-loop**: Function call validation via web interface

## Creating a Server Agent

### Syntax with options

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/server"
    "github.com/snipwise/nova/nova-sdk/models"
)

// Create a simple server agent
agent, err := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithPort(8080),
)

// Create a full-featured server agent
agent, err := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithPort(8080),
    server.WithToolsAgent(toolsAgent),
    server.WithRagAgent(ragAgent),
    server.WithCompressorAgentAndContextSize(compressorAgent, 8000),
    server.WithExecuteFn(myCustomExecutor),
)
```

### Available options

| Option | Description |
|--------|-------------|
| `WithPort(port int)` | Sets the HTTP port (default: 8080) |
| `WithExecuteFn(fn)` | Custom function executor for tools |
| `WithConfirmationPromptFn(fn)` | Custom confirmation function for human-in-the-loop |
| `WithToolsAgent(toolsAgent)` | Adds an agent for function execution |
| `WithTasksAgent(tasksAgent)` | Adds a tasks agent for task planning and orchestration |
| `WithRagAgent(ragAgent)` | Adds a RAG agent for document retrieval |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | RAG with similarity configuration |
| `WithCompressorAgent(compressorAgent)` | Adds an agent for context compression |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Compressor with context size limit |

## HTTP API Routes

### Main routes

#### `POST /completion`
Generates a completion with SSE streaming.

**Request Body:**
```json
{
  "data": {
    "message": "Your question here"
  }
}
```

**Response:** Server-Sent Events (SSE)
```
data: {"message": "Response chunk..."}
data: {"message": "", "finish_reason": "stop"}
```

**Processing pipeline:**
1. Context compression if needed (CompressorAgent)
2. Function call detection and execution (ToolsAgent)
3. Relevant context search (RagAgent)
4. Response generation with streaming

#### `POST /completion/stop`
Stops the current streaming.

**Response:**
```json
{
  "status": "ok",
  "message": "Stream stopped"
}
```

### Memory management routes

#### `POST /memory/reset`
Resets the conversation history.

**Response:**
```json
{
  "status": "ok",
  "message": "Memory reset successfully"
}
```

#### `GET /memory/messages/list`
Retrieves all conversation messages.

**Response:**
```json
{
  "messages": [
    {
      "role": "user",
      "content": "Message..."
    }
  ]
}
```

#### `GET /memory/messages/context-size`
Gets the current context size.

**Response:**
```json
{
  "messages_count": 10,
  "characters_count": 1500,
  "limit": 8000
}
```

### Operation management routes (Tools)

These routes are used for function call validation (human-in-the-loop).

#### `POST /operation/validate`
Validates a pending tool call operation.

**Request Body:**
```json
{
  "operation_id": "op_12345"
}
```

**Response:** SSE
```
data: {"message": "✅ Operation op_12345 validated<br>"}
```

#### `POST /operation/cancel`
Cancels a pending tool call operation.

**Request Body:**
```json
{
  "operation_id": "op_12345"
}
```

**Response:** SSE
```
data: {"message": "⛔️ Operation op_12345 cancelled<br>"}
```

#### `POST /operation/reset`
Cancels all pending operations.

**Response:** SSE
```
data: {"message": "🔄 All pending operations cancelled (3 operations)"}
```

### Information routes

#### `GET /models`
Returns information about the models being used.

**Response:**
```json
{
  "status": "ok",
  "chat_model": "qwen2.5:1.5b",
  "embeddings_model": "mxbai-embed-large",
  "tools_model": "jan-nano"
}
```

#### `GET /health`
Checks the server health status.

**Response:**
```json
{
  "status": "ok"
}
```

## Starting the server

```go
// Start the server (blocking)
if err := agent.StartServer(); err != nil {
    log.Fatal(err)
}
```

The server starts on `http://localhost:8080` (or the configured port).

## Usage modes

### 1. HTTP/API Mode
For REST API usage with web interface.
- Tool calls require validation via `/operation/validate`
- SSE streaming for real-time responses

### 2. CLI Mode
For direct command-line usage.
```go
result, err := agent.StreamCompletion(question, callback)
```
- By default, tool calls are auto-confirmed
- Use `WithConfirmationPromptFn` to enable human-in-the-loop
- Streaming via callback

**Example with user confirmation (CLI)** :
```go
// Custom confirmation function
confirmationPrompt := func(functionName string, arguments string) tools.ConfirmationResponse {
    fmt.Printf("Execute %s with args %s? (y/n/q): ", functionName, arguments)
    var response string
    fmt.Scanln(&response)

    switch response {
    case "y":
        return tools.Confirmed
    case "n":
        return tools.Denied
    case "q":
        return tools.Quit
    default:
        return tools.Denied
    }
}

// Create agent with confirmation
agent, _ := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithToolsAgent(toolsAgent),
    server.WithExecuteFn(executeFunction),
    server.WithConfirmationPromptFn(confirmationPrompt),
)
```

## Tool Call Notifications

When a tool call is detected, an SSE notification is sent:

```json
{
  "kind": "tool_call",
  "status": "pending",
  "operation_id": "op_12345",
  "message": "Tool call detected: calculate"
}
```

The user can then validate or cancel the operation via the `/operation/*` routes.

## Complete example

```go
ctx := context.Background()

// Configuration
agentConfig := agents.Config{
    Name: "Assistant",
    Instructions: "You are a helpful assistant.",
}
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    ModelID: "qwen2.5:1.5b",
}

// Create the agent
agent, err := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithPort(8080),
    server.WithToolsAgent(toolsAgent),
)
if err != nil {
    log.Fatal(err)
}

// Start the server
log.Println("🚀 Starting server on :8080")
if err := agent.StartServer(); err != nil {
    log.Fatal(err)
}
```

## Processing pipeline (POST /completion)

```
1. Context compression (if CompressorAgent configured)
   ↓
2. Tool call detection (if ToolsAgent configured)
   ↓
3. SSE notification of detected tool calls
   ↓
4. User validation via /operation/validate or /cancel
   ↓
5. Function execution (if validated)
   ↓
6. Add result to context
   ↓
7. RAG search (if RagAgent configured)
   ↓
8. Response generation with SSE streaming
   ↓
9. State cleanup
```

## Notes

- Default port: **8080**
- Streaming format: **Server-Sent Events (SSE)**
- CORS: **Enabled** (`Access-Control-Allow-Origin: *`)
- The server uses Go's standard `http.ServeMux`
- Tool call operations are managed with Go channels for concurrency
