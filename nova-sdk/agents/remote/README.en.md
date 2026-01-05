# Remote Agent

## Description

The **Remote Agent** is an HTTP client that connects to a remote Server Agent or Crew Server Agent. It allows interaction with hosted agents via a REST API with SSE (Server-Sent Events) support for real-time streaming.

## Features

- **Lightweight HTTP client** : Connect to remote agents via HTTP/REST
- **SSE streaming** : Real-time streaming support with Server-Sent Events
- **Tool call management** : Remote function call validation and cancellation
- **Health checks** : Server availability verification
- **History management** : Access to server-side conversation history
- **Callbacks** : Customizable notifications for function calls

## Use cases

The Remote Agent is used for:
- **Client-server applications** : Frontend connecting to AI backend
- **Distributed microservices** : Communication between services via HTTP
- **User interfaces** : Web apps, mobile applications, remote CLIs
- **Load balancing** : Connection to multiple server instances
- **Decoupled architecture** : Separation between client and AI logic

## Creating a Remote Agent

### Basic syntax

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/remote"
)

ctx := context.Background()

// Create remote agent
agent, err := remote.NewAgent(
    ctx,
    agents.Config{
        Name: "Remote Client",
    },
    "http://localhost:8080", // Server URL
)
if err != nil {
    log.Fatal(err)
}
```

## Main methods

### GenerateCompletion - Non-streaming completion

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Send question and receive complete response
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "What is the capital of France?"},
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Response: %s\n", result.Response)
fmt.Printf("Finish Reason: %s\n", result.FinishReason)
```

### GenerateStreamCompletion - Streaming completion

```go
// Real-time streaming
_, err := agent.GenerateStreamCompletion(
    []messages.Message{
        {Role: roles.User, Content: "Tell me a story."},
    },
    func(chunk string, finishReason string) error {
        if chunk != "" {
            fmt.Print(chunk) // Display as received
        }
        if finishReason == "stop" {
            fmt.Println()
        }
        return nil
    },
)
```

### Health and server information

```go
// Check server health
if agent.IsHealthy() {
    fmt.Println("✅ Server is healthy")
} else {
    fmt.Println("❌ Server is not available")
    return
}

// Get detailed status
health, err := agent.GetHealth()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %s\n", health.Status)

// Get model information
modelsInfo, err := agent.GetModelsInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Chat Model: %s\n", modelsInfo.ChatModel)
fmt.Printf("Embeddings Model: %s\n", modelsInfo.EmbeddingsModel)
fmt.Printf("Tools Model: %s\n", modelsInfo.ToolsModel)
```

### History management

```go
// Get all conversation messages (server-side)
messages := agent.GetMessages()
for _, msg := range messages {
    fmt.Printf("%s: %s\n", msg.Role, msg.Content)
}

// Get context size in tokens
contextSize := agent.GetContextSize()
fmt.Printf("Context: %d tokens\n", contextSize)

// Reset conversation
agent.ResetMessages()

// Export to JSON
jsonData, err := agent.ExportMessagesToJSON()
if err == nil {
    fmt.Println(jsonData)
}
```

### Tool call management

```go
// Validate a pending function call
err := agent.ValidateOperation("operation-id-12345")
if err != nil {
    log.Fatal(err)
}

// Cancel a pending function call
err = agent.CancelOperation("operation-id-12345")
if err != nil {
    log.Fatal(err)
}

// Reset all pending operations
err = agent.ResetOperations()
if err != nil {
    log.Fatal(err)
}
```

### Tool call callback

```go
// Set callback for tool call notifications
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("🔔 Tool call detected: %s\n", message)
    fmt.Printf("📝 Operation ID: %s\n", operationID)

    // Auto-validate (or ask user for confirmation)
    return agent.ValidateOperation(operationID)

    // Or cancel
    // return agent.CancelOperation(operationID)
})

// Callback will be called automatically during streaming
agent.GenerateStreamCompletion(messages, streamCallback)
```

### Stream control

```go
// Stop current streaming
agent.StopStream()
```

### Getters

```go
// Basic information
name := agent.GetName()
modelID := agent.GetModelID() // Retrieved from server
kind := agent.Kind() // Returns agents.Remote

// Context
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

## Result structures

### CompletionResult

```go
type CompletionResult struct {
    Response     string // Complete response
    FinishReason string // "stop", "length", etc.
}
```

### ReasoningResult

```go
type ReasoningResult struct {
    Response     string // Complete response
    Reasoning    string // Reasoning (not currently supported)
    FinishReason string // "stop", "length", etc.
}
```

### ModelsInfo

```go
type ModelsInfo struct {
    Status           string // "ok"
    ChatModel        string // Chat model used
    EmbeddingsModel  string // Embeddings model
    ToolsModel       string // Tools model
}
```

### HealthStatus

```go
type HealthStatus struct {
    Status string // "ok" if healthy
}
```

## Complete example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/remote"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

func main() {
    ctx := context.Background()

    // Create remote client
    agent, err := remote.NewAgent(
        ctx,
        agents.Config{
            Name: "Remote Client",
        },
        "http://localhost:8080",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Check server health
    if !agent.IsHealthy() {
        log.Fatal("Server is not available")
    }
    fmt.Println("✅ Connected to server")

    // Get server information
    modelsInfo, _ := agent.GetModelsInfo()
    fmt.Printf("Using model: %s\n\n", modelsInfo.ChatModel)

    // Set callback for tool calls
    agent.SetToolCallCallback(func(operationID string, message string) error {
        fmt.Printf("\n🔔 Tool call: %s\n", message)
        fmt.Printf("📝 Validating operation: %s\n\n", operationID)
        return agent.ValidateOperation(operationID)
    })

    // Example 1: Simple completion
    fmt.Println("=== Simple Question ===")
    result, err := agent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: "What is 2 + 2?"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Answer: %s\n\n", result.Response)

    // Example 2: Streaming with tool calls
    fmt.Println("=== Question with Tools ===")
    fmt.Print("Response: ")
    _, err = agent.GenerateStreamCompletion(
        []messages.Message{
            {Role: roles.User, Content: "Calculate 40 + 2 and say hello to Alice"},
        },
        func(chunk string, finishReason string) error {
            if chunk != "" {
                fmt.Print(chunk)
            }
            if finishReason == "stop" {
                fmt.Println()
            }
            return nil
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Display history
    fmt.Println("\n=== Conversation History ===")
    messages := agent.GetMessages()
    for i, msg := range messages {
        fmt.Printf("[%d] %s: %s\n", i+1, msg.Role, msg.Content)
    }

    // Display context size
    fmt.Printf("\nContext size: %d tokens\n", agent.GetContextSize())

    // Reset conversation
    agent.ResetMessages()
    fmt.Println("Conversation reset")
}
```

## HTTP endpoints used

The Remote Agent communicates with the following server endpoints:

### Completion
- `POST /completion` - SSE streaming completion

### Memory
- `GET /memory/messages/list` - Message list
- `GET /memory/messages/context-size` - Context size
- `POST /memory/reset` - Reset history

### Operations (Tool calls)
- `POST /operation/validate` - Validate function call
- `POST /operation/cancel` - Cancel function call
- `POST /operation/reset` - Reset all operations

### Information
- `GET /models` - Model information
- `GET /health` - Health check

### Control
- `POST /completion/stop` - Stop streaming

## SSE (Server-Sent Events) format

The Remote Agent parses SSE events in this format:

```
data: {"message": "text chunk", "finish_reason": ""}
data: {"message": "", "finish_reason": "stop"}
data: {"kind": "tool_call", "operation_id": "abc123", "message": "Tool detected"}
data: {"error": "error message"}
```

## Important notes

- **Kind** : Returns `agents.Remote`
- **Local history** : Remote Agent does NOT maintain local history
  - `AddMessage()` and `AddMessages()` are no-ops
  - History is entirely server-managed
  - `GetMessages()` retrieves history from server
- **Streaming** : Uses Server-Sent Events (SSE) for real-time
- **Tool calls** : Require manual validation or via callback
- **Connection** : Uses standard HTTP client
- **Timeouts** : Managed by standard Go context

## Recommendations

### Best practices

1. **Health checks** : Always verify `IsHealthy()` before use
2. **Error handling** : Check network and server errors
3. **Callbacks** : Use `SetToolCallCallback` to handle tool calls automatically
4. **Context** : Use context with timeout to avoid blocking
5. **Reconnection** : Implement retry logic for connection loss

### Example with timeout

```go
import "time"

// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

agent, err := remote.NewAgent(ctx, agentConfig, serverURL)
if err != nil {
    log.Fatal(err)
}
```

### Example with retry

```go
func connectWithRetry(baseURL string, maxRetries int) (*remote.Agent, error) {
    for i := 0; i < maxRetries; i++ {
        agent, err := remote.NewAgent(context.Background(), agents.Config{
            Name: "Client",
        }, baseURL)

        if err == nil && agent.IsHealthy() {
            return agent, nil
        }

        fmt.Printf("Attempt %d failed, retrying...\n", i+1)
        time.Sleep(2 * time.Second)
    }
    return nil, fmt.Errorf("failed to connect after %d attempts", maxRetries)
}
```

### Automatic tool call validation

```go
// Auto-validate all tool calls
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("Auto-validating: %s\n", message)
    return agent.ValidateOperation(operationID)
})
```

### Manual validation with confirmation

```go
// Ask user for confirmation
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("Tool call: %s\n", message)
    fmt.Printf("Validate? (y/n): ")

    var response string
    fmt.Scanln(&response)

    if response == "y" {
        return agent.ValidateOperation(operationID)
    }
    return agent.CancelOperation(operationID)
})
```

## Differences with local agents

| Feature | Remote Agent | Local Agents |
|---------|--------------|--------------|
| History | Server-managed | Locally managed |
| AddMessage() | No-op | Works |
| Streaming | Via SSE | Direct |
| Tool calls | Validation required | Direct execution |
| Configuration | Server-side | Client-side |
| Latency | Network | Minimal |

## Troubleshooting

### Server unavailable

```go
if !agent.IsHealthy() {
    fmt.Println("Server is down. Check:")
    fmt.Println("1. Server is running")
    fmt.Println("2. URL is correct")
    fmt.Println("3. Firewall allows connection")
}
```

### Tool calls not executing

```go
// Verify callback is set
agent.SetToolCallCallback(func(operationID, message string) error {
    // IMPORTANT: Validate the operation
    return agent.ValidateOperation(operationID)
})
```

### Timeout during streaming

```go
// Use context with longer timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

agent.SetContext(ctx)
```
