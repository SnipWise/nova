# Server Agent - Functional Presentation

## Overview

The **Server Agent** is an HTTP/REST API agent from the N.O.V.A. SDK that wraps a conversational agent (Chat Agent) and exposes its capabilities through a modern web interface with real-time streaming support.

It is a sophisticated orchestration component capable of coordinating multiple specialized agents to provide a complete conversational API with tool calling, document retrieval (RAG), and intelligent context management.

## Main Features

### 1. REST API with SSE Streaming

- **RESTful HTTP endpoints** for chat completions
- **Real-time streaming** via Server-Sent Events (SSE)
- **Progressive responses** for a smooth user experience
- **Real-time error handling** via SSE

### 2. Tool Calling with Confirmation Workflow

- **Automatic detection** of tool calls in user requests
- **Parallel calls** of multiple tools
- **Web-based validation workflow**: tools require approval before execution
- **Notification system** via SSE to inform the client of pending operations
- **Operation management**: validation, cancellation, reset

### 3. RAG (Retrieval-Augmented Generation)

- **Similarity search** in a vector document database
- **Automatic injection** of relevant context before generation
- **Flexible configuration**: similarity threshold, maximum number of documents
- **Enhanced responses** with external information

### 4. Intelligent Context Compression

- **Automatic management** of token limits
- **On-demand or automatic compression** beyond a threshold
- **Safety thresholds**: warning at 80%, reset at 90%
- **Preservation of essential context** while respecting model limits

### 5. Conversational Memory Management

- **Message history** persisted during the session
- **Token counting** to monitor usage
- **History reset**
- **JSON export** of conversations
- **Listing** of all messages

## Architecture

### Main Structure

```go
type ServerAgent struct {
    chatAgent       *chat.Agent      // Main conversational agent
    toolsAgent      *tools.Agent     // Optional: tool detection/execution
    ragAgent        *rag.Agent       // Optional: document retrieval
    compressorAgent *compressor.Agent // Optional: context compression

    // RAG Configuration
    similarityLimit float64  // Default: 0.6
    maxSimilarities int      // Default: 3

    // Compression Configuration
    contextSizeLimit int     // Default: 8000

    // Server Configuration
    port string
    ctx  context.Context
    log  logger.Logger

    // Operation Management
    pendingOperations       map[string]*PendingOperation
    operationsMutex         sync.RWMutex
    stopStreamChan          chan bool
    currentNotificationChan chan ToolCallNotification
    notificationChanMutex   sync.Mutex

    // Custom Tool Executor
    executeFn func(string, string) (string, error)
}
```

### Specialized Agents

The Server Agent orchestrates four types of agents:

1. **Chat Agent** (required): conversational text generation
2. **Tools Agent** (optional): function detection and execution
3. **RAG Agent** (optional): relevant document search
4. **Compressor Agent** (optional): context compression

## HTTP Endpoints

### Completion

#### `POST /completion`

Stream a chat completion with SSE.

**Request:**

```json
{
  "data": {
    "message": "Your question here"
  }
}
```

**Response:** SSE stream with text chunks

```
data: {"message": "text chunk"}
data: {"message": "more text"}
data: {"message": "", "finish_reason": "stop"}
```

**Tool notifications (if tools agent configured):**

```json
data: {
  "kind": "tool_call",
  "status": "pending",
  "operation_id": "op_0x140003dcbe0",
  "message": "Tool call detected: function_name"
}
```

#### `POST /completion/stop`

Stop the current streaming operation.

### Memory Management

#### `POST /memory/reset`

Clear conversation history.

**Response:**

```json
{
  "status": "ok",
  "message": "Memory reset"
}
```

#### `GET /memory/messages/list`

Retrieve all messages from history.

**Response:**

```json
{
  "messages": [
    { "role": "user", "content": "..." },
    { "role": "assistant", "content": "..." }
  ]
}
```

#### `GET /memory/messages/context-size`

Get token count and statistics.

**Response:**

```json
{
  "context_size": 1234,
  "message": "Current context size: 1234 tokens"
}
```

### Tool Operation Management

#### `POST /operation/validate`

Approve a pending tool call.

**Request:**

```json
{
  "operation_id": "op_0x140003dcbe0"
}
```

**Response:**

```json
{
  "status": "validated",
  "operation_id": "op_0x140003dcbe0"
}
```

#### `POST /operation/cancel`

Reject a pending tool call.

**Request:**

```json
{
  "operation_id": "op_0x140003dcbe0"
}
```

**Response:**

```json
{
  "status": "cancelled",
  "operation_id": "op_0x140003dcbe0"
}
```

#### `POST /operation/reset`

Cancel all pending operations.

**Response:**

```json
{
  "status": "ok",
  "message": "All pending operations cancelled"
}
```

### Information

#### `GET /models`

Information about used models.

**Response:**

```json
{
  "chat_model": "hf.co/menlo/jan-nano-gguf:q4_k_m",
  "tools_model": "hf.co/menlo/jan-nano-gguf:q4_k_m",
  "embeddings_model": "all-minilm:l6-v2"
}
```

#### `GET /health`

Server health check.

**Response:**

```json
{
  "status": "ok",
  "message": "Server is running"
}
```

## Request Execution Flow

### Complete `/completion` Request Processing

1. **Context Compression** (if compressor agent configured)
   - Check context size
   - Compress if threshold exceeded

2. **Request Parsing**
   - JSON decoding of message

3. **SSE Streaming Setup**
   - `text/event-stream` headers
   - Notification channel creation

4. **Tool Detection** (if tools agent configured)
   - Call `DetectParallelToolCallsWithConfirmation()`
   - Send SSE notifications for pending operations
   - Wait for validation/cancellation
   - Execute approved tools
   - Add results to context

5. **RAG Similarity Search** (if RAG agent configured)
   - Search for relevant documents
   - Inject as system message

6. **Streaming Completion Generation**
   - Stream chunks via SSE
   - Handle stop signals
   - Send finish reason

7. **Error Handling**
   - Stream errors via SSE

## Usage Examples

### 1. Basic Server Agent (Chat only)

```go
import (
    "context"
    "log"

    "nova-sdk/agents"
    "nova-sdk/agents/server"
    "nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // The executeFunction parameter is now OPTIONAL
    // You can either:
    // 1. Omit it completely (uses default executeFunction)
    // 2. Pass nil (uses default executeFunction)
    // 3. Pass a custom function
    agent, err := server.NewAgent(
        ctx,
        agents.Config{
            Name:               "bob-server-agent",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are Bob, a helpful AI assistant.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.4),
        },
        ":3500",
        // executeFunction is optional - omitted here, will use default
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Fatal(agent.StartServer())
}
```

**Usage:**

```bash
curl -X POST http://localhost:3500/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "Hello!"}}'
```

### 2. Server Agent with Tools

```go
func executeFunction(functionName, arguments string) (string, error) {
    // Your tool implementations
    switch functionName {
    case "sayHello":
        return "Hello " + arguments, nil
    default:
        return "", fmt.Errorf("unknown function: %s", functionName)
    }
}

func main() {
    ctx := context.Background()

    // Create server agent
    serverAgent, err := server.NewAgent(
        ctx,
        agentConfig,
        modelConfig,
        ":8080",
        executeFunction,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create tools agent
    toolsAgent, err := tools.NewAgent(
        ctx,
        agentConfig,
        models.Config{
            Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature:       models.Float64(0.0),
            ParallelToolCalls: models.Bool(true),
        },
        tools.WithTools(GetToolsIndex()),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Attach tools agent
    serverAgent.SetToolsAgent(toolsAgent)

    log.Fatal(serverAgent.StartServer())
}
```

**Tool call workflow:**

```bash
# 1. Trigger a tool call
curl -X POST http://localhost:8080/completion \
  -d '{"data":{"message":"Say hello to Alice"}}'

# SSE response contains:
# data: {"kind":"tool_call","status":"pending","operation_id":"op_0x140003dcbe0",...}

# 2. Validate the operation
curl -X POST http://localhost:8080/operation/validate \
  -d '{"operation_id":"op_0x140003dcbe0"}'

# Or 3. Cancel the operation
curl -X POST http://localhost:8080/operation/cancel \
  -d '{"operation_id":"op_0x140003dcbe0"}'
```

### 3. Full-Featured Server Agent (Tools + RAG + Compression)

```go
func main() {
    ctx := context.Background()

    // Create server agent
    serverAgent, err := server.NewAgent(
        ctx,
        agentConfig,
        modelConfig,
        ":8080",
        executeFunction,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Add tools agent
    toolsAgent, err := tools.NewAgent(
        ctx,
        toolsConfig,
        toolsModelConfig,
        tools.WithTools(GetToolsIndex()),
    )
    serverAgent.SetToolsAgent(toolsAgent)

    // Add RAG agent
    ragAgent, err := rag.NewAgent(ctx, ragConfig, embeddingsModelConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Load documents
    documents := []string{
        "Content of document 1...",
        "Content of document 2...",
    }

    for _, content := range documents {
        chunks := chunks.SplitMarkdownBySections(content)
        for _, chunk := range chunks {
            ragAgent.SaveEmbedding(chunk)
        }
    }
    serverAgent.SetRagAgent(ragAgent)
    serverAgent.SetSimilarityLimit(0.6)
    serverAgent.SetMaxSimilarities(3)

    // Add compressor agent
    compressorAgent, err := compressor.NewAgent(
        ctx,
        compressorConfig,
        compressorModelConfig,
        compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    )
    if err != nil {
        log.Fatal(err)
    }
    serverAgent.SetCompressorAgent(compressorAgent)
    serverAgent.SetContextSizeLimit(3000)

    log.Fatal(serverAgent.StartServer())
}
```

## Configuration Methods

### Server Configuration

```go
agent.SetPort(":8080")
port := agent.GetPort()
```

### Agent Configuration

```go
agent.SetToolsAgent(toolsAgent)
agent.SetRagAgent(ragAgent)
agent.SetCompressorAgent(compressorAgent)
```

### RAG Configuration

```go
agent.SetSimilarityLimit(0.6)      // Similarity threshold (0.0 - 1.0)
agent.SetMaxSimilarities(3)        // Max number of documents to retrieve
```

### Compression Configuration

```go
agent.SetContextSizeLimit(8000)    // Token limit
agent.CompressChatAgentContext()   // Force compression
agent.CompressChatAgentContextIfOverLimit() // Conditional compression
```

### Memory Management

```go
agent.ResetMessages()              // Clear history
agent.AddMessage(role, content)    // Add a message
messages := agent.GetMessages()    // Retrieve all messages
size := agent.GetContextSize()     // Get context size
json := agent.ExportMessagesToJSON() // Export to JSON
```

### Custom Execute Function

```go
agent.SetExecuteFunction(func(functionName, arguments string) (string, error) {
    // Your execution logic
    return result, nil
})
```

## SSE Message Format

### Regular text chunk

```
data: {"message": "text chunk"}
```

### Tool call notification

```json
data: {
  "kind": "tool_call",
  "status": "pending",
  "operation_id": "op_0x140003dcbe0",
  "function_name": "sayHello",
  "arguments": "{\"name\":\"Alice\"}",
  "message": "Tool call detected: sayHello"
}
```

### Finish response

```
data: {"message": "", "finish_reason": "stop"}
```

### Error

```
data: {"error": "error message"}
```

## Security and Concurrency

### Thread-safety

- **Mutex-protected operations map** for concurrent tool calls
- **Channel-based communication** for operation responses
- **Notification channel locking** to prevent race conditions

### Operation Management

- Each tool call receives a unique ID
- Pending operations are stored in a thread-safe map
- Response channels enable asynchronous communication
- Timeouts and cancellations handled properly

## Logging

The Server Agent uses the N.O.V.A. SDK logging system:

```bash
# Control via environment variable
export NOVA_LOG_LEVEL=DEBUG  # DEBUG, INFO, WARN, ERROR
```

## Default Values

| Parameter           | Default Value   | Description                        |
| ------------------- | --------------- | ---------------------------------- |
| Port                | Set at creation | HTTP server port                   |
| Similarity Limit    | 0.6             | RAG similarity threshold (0.0-1.0) |
| Max Similarities    | 3               | Max number of RAG documents        |
| Context Size Limit  | 8000            | Token limit before compression     |
| Compression Warning | 80%             | Warning for approaching limit      |
| Compression Reset   | 90%             | Forced reset                       |

## Test Scripts

The directory contains several test scripts:

- **`stream.sh`** - Basic streaming completion test
- **`call-tool.sh`** - Trigger tool calls
- **`validate.sh`** - Approve pending operation
- **`cancel.sh`** - Reject pending operation
- **`reset.sh`** - Cancel all operations
- **`test-api.sh`** - Complete test of all endpoints

## Sample Directory

Complete examples are available in `/samples`:

- **`49-server-agent-stream`** - Basic server agent with chat only
- **`50-server-agent-with-tools`** - Server agent with tool calling
- **`54-server-agent-tools-rag-compress`** - Full configuration with all agents

## Future Features

Based on task files in `/nova-sdk/agents/server/tasks/`:

- **Multi-chat agents capability** - Orchestration of multiple chat agents
- **Custom endpoint addition** - Extend the API with custom endpoints
- Additional RAG and compressor enhancements

## Conclusion

The **Server Agent** is a powerful and flexible component of the N.O.V.A. SDK that transforms any conversational agent into a complete REST API with:

- ✅ Real-time streaming via SSE
- ✅ Tool calling with validation workflow
- ✅ Relevant document retrieval (RAG)
- ✅ Intelligent context compression
- ✅ Complete conversational memory management
- ✅ Modular and extensible architecture

It enables rapid deployment of conversational AI assistants with advanced capabilities, while maintaining fine-grained control over sensitive operations like tool execution.
