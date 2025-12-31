# Remote Agent API Reference

## Overview

The Remote Agent provides a complete interface to interact with a Server Agent via HTTP/REST. It includes all standard chat agent methods plus additional operation management methods for handling tool calls.

## Agent Creation

```go
agent, err := remote.NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    baseURL string,
)
```

**Parameters:**

- `ctx`: Context for cancellation and timeouts
- `agentConfig`: Agent configuration (name, system instructions)
- `baseURL`: Server URL (e.g., "http://localhost:8080")

**Returns:** `(*remote.Agent, error)`

### Setting Tool Call Callback

```go
agent.SetToolCallCallback(func(operationID string, message string) error {
    // Handle tool call detection
    // operationID: The unique ID of the detected operation
    // message: Description of the tool call
    return nil
})
```

**Use this to:**

- Capture operation IDs as they are detected
- Implement automatic validation/cancellation logic
- Track tool calls for auditing

**Example:**

```go
agent.SetToolCallCallback(func(operationID string, message string) error {
    log.Printf("Tool call detected: %s (ID: %s)", message, operationID)
    // Auto-validate
    return agent.ValidateOperation(operationID)
})
```

**Important Notes:**

- Pass `nil` to clear the callback: `agent.SetToolCallCallback(nil)`
- Only one callback can be active at a time (setting a new one replaces the previous)
- The callback is invoked synchronously during streaming
- If the callback returns an error, streaming stops

## Core Methods

### Agent Information

#### `Kind() agents.Kind`

Returns the agent type (always `agents.Remote`)

#### `GetName() string`

Returns the agent name

#### `GetModelID() string`

Returns the model ID by fetching from the server.

**HTTP:** `GET /models`

### Server Information

#### `GetModelsInfo() (*ModelsInfo, error)`

Returns detailed information about all models used by the server.

**HTTP:** `GET /models`

**Returns:** `*ModelsInfo` containing:

- `Status`: Server status
- `ChatModel`: Model used for chat completions
- `EmbeddingsModel`: Model used for embeddings
- `ToolsModel`: Model used for tool calling

**Example:**

```go
info, err := agent.GetModelsInfo()
if err != nil {
    log.Printf("Failed to get models info: %v", err)
} else {
    fmt.Printf("Chat Model: %s\n", info.ChatModel)
    fmt.Printf("Embeddings Model: %s\n", info.EmbeddingsModel)
    fmt.Printf("Tools Model: %s\n", info.ToolsModel)
}
```

#### `GetHealth() (*HealthStatus, error)`

Checks if the server is healthy and reachable.

**HTTP:** `GET /health`

**Returns:** `*HealthStatus` containing:

- `Status`: Server status ("ok" if healthy)

**Example:**

```go
health, err := agent.GetHealth()
if err != nil {
    log.Printf("Health check failed: %v", err)
} else {
    fmt.Printf("Server status: %s\n", health.Status)
}
```

#### `IsHealthy() bool`

Convenience method that returns true if the server is healthy.

**Example:**

```go
if agent.IsHealthy() {
    fmt.Println("Server is ready")
} else {
    fmt.Println("Server is not available")
}
```

### Message Operations

#### `GetMessages() []messages.Message`

Fetches all conversation messages from the server.

**HTTP:** `GET /memory/messages/list`

#### `GetContextSize() int`

Returns the approximate token count of the current context.

**HTTP:** `GET /memory/messages/context-size`

#### `AddMessage(role roles.Role, content string)`

No-op for remote agent. Messages are managed server-side.

#### `ResetMessages()`

Clears all messages except system instruction on the server.

**HTTP:** `POST /memory/reset`

### Completion Generation

#### `GenerateCompletion(userMessages []messages.Message) (*CompletionResult, error)`

Sends messages and returns the complete response (non-streaming).

**Parameters:**

- `userMessages`: Array of messages to send

**Returns:** `*CompletionResult` containing:

- `Response`: Complete response text
- `FinishReason`: Reason completion ended

#### `GenerateStreamCompletion(userMessages []messages.Message, callback StreamCallback) (*CompletionResult, error)`

Sends messages and streams the response via callback.

**Parameters:**

- `userMessages`: Array of messages to send
- `callback`: Function called for each chunk: `func(chunk string, finishReason string) error`

**Returns:** `*CompletionResult` with the complete response

**HTTP:** `POST /completion` (Server-Sent Events)

**Callback receives:**

- `chunk`: Text chunk (empty for tool call notifications)
- `finishReason`: "stop" when complete, "" otherwise

#### `GenerateCompletionWithReasoning(userMessages []messages.Message) (*ReasoningResult, error)`

Similar to `GenerateCompletion` but returns reasoning (not yet fully supported by server).

#### `GenerateStreamCompletionWithReasoning(userMessages []messages.Message, reasoningCallback StreamCallback, responseCallback StreamCallback) (*ReasoningResult, error)`

Similar to `GenerateStreamCompletion` but with separate callbacks for reasoning and response.

### Stream Control

#### `StopStream()`

Interrupts the current streaming operation.

**HTTP:** `POST /completion/stop`

### Export

#### `ExportMessagesToJSON() (string, error)`

Exports the conversation history to JSON format.

**Returns:** JSON string of all messages

## Operation Management Methods (NEW)

These methods allow programmatic control of tool call operations.

### `ValidateOperation(operationID string) error`

Approves a pending tool call operation, allowing it to execute.

**Parameters:**

- `operationID`: The operation ID (e.g., "op_0x14000126020")

**HTTP:** `POST /operation/validate`

**Example:**

```go
err := agent.ValidateOperation("op_0x14000126020")
if err != nil {
    log.Printf("Failed to validate: %v", err)
}
```

### `CancelOperation(operationID string) error`

Cancels a pending tool call operation, preventing execution.

**Parameters:**

- `operationID`: The operation ID

**HTTP:** `POST /operation/cancel`

**Example:**

```go
err := agent.CancelOperation("op_0x14000126020")
if err != nil {
    log.Printf("Failed to cancel: %v", err)
}
```

### `ResetOperations() error`

Cancels all pending tool call operations at once.

**HTTP:** `POST /operation/reset`

**Example:**

```go
err := agent.ResetOperations()
if err != nil {
    log.Printf("Failed to reset: %v", err)
}
```

## Type Definitions

### CompletionResult

```go
type CompletionResult struct {
    Response     string  // Complete response text
    FinishReason string  // Reason completion ended ("stop", "length", etc.)
}
```

### ReasoningResult

```go
type ReasoningResult struct {
    Response     string  // Complete response text
    Reasoning    string  // Reasoning content (if supported)
    FinishReason string  // Reason completion ended
}
```

### ModelsInfo

```go
type ModelsInfo struct {
    Status           string  // Server status
    ChatModel        string  // Model used for chat completions
    EmbeddingsModel  string  // Model used for embeddings
    ToolsModel       string  // Model used for tool calling
}
```

### HealthStatus

```go
type HealthStatus struct {
    Status string  // Server health status ("ok" if healthy)
}
```

### StreamCallback

```go
type StreamCallback func(chunk string, finishReason string) error
```

Callback function for streaming responses:

- Called for each chunk of text
- Called with finishReason when complete
- Return error to stop streaming

### ToolCallCallback

```go
type ToolCallCallback func(operationID string, message string) error
```

Callback function for tool call notifications:

- Called when a tool call is detected during streaming
- `operationID`: Unique identifier for the operation
- `message`: Description of the tool call
- Return error to stop streaming
- Use `SetToolCallCallback()` to register this callback

## Usage Patterns

### Basic Streaming

```go
result, err := agent.GenerateStreamCompletion(messages, func(chunk string, finishReason string) error {
    if chunk != "" {
        fmt.Print(chunk)
    }
    return nil
})
```

### Auto-Validate Tool Calls

```go
// Set up callback to auto-validate
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("Auto-validating: %s\n", message)
    return agent.ValidateOperation(operationID)
})

// Start streaming - operations will be validated automatically
agent.GenerateStreamCompletion(messages, streamCallback)
```

### Conditional Approval

```go
// Whitelist of safe operations
safeOps := map[string]bool{
    "say_hello":     true,
    "calculate_sum": true,
}

agent.SetToolCallCallback(func(operationID string, message string) error {
    // Extract function name from message
    functionName := extractFunctionName(message)

    if safeOps[functionName] {
        fmt.Printf("Auto-approving safe operation: %s\n", functionName)
        return agent.ValidateOperation(operationID)
    } else {
        fmt.Printf("Auto-cancelling unsafe operation: %s\n", functionName)
        return agent.CancelOperation(operationID)
    }
})
```

### Timeout-based Auto-Cancel

```go
go func() {
    time.Sleep(30 * time.Second)
    if err := agent.ResetOperations(); err != nil {
        log.Printf("Failed to reset: %v", err)
    }
}()
```

## Error Handling

All methods return errors that should be checked:

```go
if err := agent.ValidateOperation(opID); err != nil {
    switch {
    case strings.Contains(err.Error(), "not found"):
        // Operation already processed or doesn't exist
    case strings.Contains(err.Error(), "status 500"):
        // Server error
    default:
        // Other error
    }
}
```

## Server Requirements

The remote agent requires a server implementing these endpoints:

- `GET /models` - Get models information
- `GET /health` - Health check
- `POST /completion` - Streaming completions (SSE)
- `POST /completion/stop` - Stop streaming
- `GET /memory/messages/list` - List messages
- `GET /memory/messages/context-size` - Token count
- `POST /memory/reset` - Reset conversation
- `POST /operation/validate` - Validate operation
- `POST /operation/cancel` - Cancel operation
- `POST /operation/reset` - Reset all operations

See [50-server-agent-with-tools](../50-server-agent-with-tools) for reference implementation.

## Related Examples

- [50-server-agent-with-tools](../50-server-agent-with-tools) - Server implementation
- [52-remote-interactive](../52-remote-interactive) - Interactive CLI
- [53-remote-programmatic](../53-remote-programmatic) - Automated handling
