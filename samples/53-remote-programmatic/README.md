# Remote Agent - Programmatic Operation Management

This example demonstrates automated handling of tool call operations using a remote agent.

## Features

- Auto-validation of tool calls
- Auto-cancellation of all operations
- Template for building custom approval logic
- Demonstrates goroutine-based operation management

## Prerequisites

- Server agent running on `http://localhost:8080` (see example 50)

## Running the Example

First, start the server in another terminal:
```bash
cd ../50-server-agent-with-tools
go run main.go
```

Then run this automated client:
```bash
go run main.go
```

## What It Does

### Example 1: Auto-Validate
Demonstrates how to automatically validate operations as they are detected using the `SetToolCallCallback`:
- Registers a callback to capture operation IDs in real-time
- Sets up a goroutine to auto-validate operations
- Validates each operation with a small delay for demonstration
- Operations are approved automatically as they appear
- Useful for trusted operations

**Key Feature:** Uses `agent.SetToolCallCallback()` to capture operation IDs directly from SSE events.

### Example 2: Auto-Cancel
Demonstrates how to automatically cancel operations as they are detected:
- Sets up a callback to auto-cancel operations
- Sends a question that triggers tool calls
- Operations are cancelled immediately as they appear
- Useful for rejecting all operations or implementing blacklists

## Building Custom Logic

This example serves as a template for building your own approval logic using the callback system:

```go
// Set up callback with custom approval logic
agent.SetToolCallCallback(func(operationID string, message string) error {
    // Extract function name from message
    functionName := extractFunctionName(message)

    // Your custom logic here
    safeOperations := []string{"say_hello", "calculate_sum"}

    for _, safe := range safeOperations {
        if strings.Contains(functionName, safe) {
            // Auto-approve safe operations
            return agent.ValidateOperation(operationID)
        }
    }

    // Auto-cancel unsafe operations
    return agent.CancelOperation(operationID)
})

// Start streaming - operations will be handled automatically
agent.GenerateStreamCompletion(messages, streamCallback)
```

**How it works:**
1. `SetToolCallCallback` registers your function
2. During streaming, when the server detects a tool call, it sends an SSE event
3. The remote agent receives the event and calls your callback
4. Your callback validates or cancels the operation
5. The server proceeds based on your decision

## Cancellation Strategies

### Strategy 1: Cancel via Callback (Recommended)
Cancel operations as they are detected in real-time:

```go
agent.SetToolCallCallback(func(operationID string, message string) error {
    return agent.CancelOperation(operationID)
})
```

**Advantages:**
- ✅ Cancels ALL operations, even those detected during streaming
- ✅ Real-time cancellation
- ✅ No operations are missed
- ✅ Clean and simple

### Strategy 2: Cancel via ResetOperations()
Cancel all pending operations at a specific point:

```go
// Start streaming in background
go func() {
    agent.GenerateStreamCompletion(messages, callback)
}()

time.Sleep(2 * time.Second)
agent.ResetOperations()  // Cancel pending operations
```

**Limitations:**
- ⚠️ Only cancels operations detected **before** the call
- ⚠️ New operations detected **after** the call are not cancelled
- ⚠️ Requires careful timing

**Recommendation:** Use Strategy 1 (callback) when you want to cancel all operations. Use Strategy 2 (ResetOperations) only when you need to cancel at a specific point in time (e.g., timeout scenarios).

## Use Cases

### Production Automation
Build systems that automatically approve safe operations and flag dangerous ones.

### Whitelisting
Auto-approve operations from a whitelist, reject everything else.

### Rate Limiting
Track operation frequency and auto-cancel if threshold exceeded.

### Timeout Management
Auto-cancel operations that haven't been manually approved within N seconds.

### Audit Logging
Log all operations and their approval status before processing.

## Advanced Patterns

### Conditional Approval
```go
func approveBasedOnContext(agent *remote.Agent, opID string, metadata map[string]interface{}) {
    if metadata["user_role"] == "admin" {
        agent.ValidateOperation(opID)
    } else if metadata["requires_approval"] == true {
        // Send to approval queue
        notifyApprovers(opID, metadata)
    } else {
        agent.CancelOperation(opID)
    }
}
```

### Batch Processing
```go
pendingOps := []string{}

// Collect operations
// ...

// Process in batch
for _, opID := range pendingOps {
    if shouldProcess(opID) {
        agent.ValidateOperation(opID)
        time.Sleep(100 * time.Millisecond) // Rate limiting
    }
}
```

### Webhook Integration
```go
func notifyAndWaitForApproval(agent *remote.Agent, opID string) {
    // Send webhook
    sendWebhook("https://approval.example.com/webhook", opID)

    // Wait for approval via callback
    approved := <-approvalChannel

    if approved {
        agent.ValidateOperation(opID)
    } else {
        agent.CancelOperation(opID)
    }
}
```

## How It Works

The automatic validation is powered by the `SetToolCallCallback` method:

1. **Server detects tool call** → Sends SSE event with `kind: "tool_call"`
2. **Remote agent receives event** → Extracts `operation_id` and `message`
3. **Callback is invoked** → Your custom logic executes
4. **Decision is made** → Validate or cancel the operation
5. **Server proceeds** → Executes approved tools or skips cancelled ones

This provides a clean, event-driven approach to operation management without manual polling or log parsing.

## Managing Callbacks

### Clearing a Callback

When you're done with a callback or want to switch strategies, clear it by passing `nil`:

```go
// Example 1 with auto-validation
agent.SetToolCallCallback(autoValidateCallback)
agent.GenerateStreamCompletion(messages1, streamCallback)

// Clear callback before Example 2
agent.SetToolCallCallback(nil)

// Example 2 without callback
agent.GenerateStreamCompletion(messages2, streamCallback)
```

**Why this is important:**
- Prevents callbacks from being invoked in subsequent operations
- Avoids errors like "send on closed channel" when using goroutines
- Allows switching between different operation management strategies

### Replacing a Callback

Setting a new callback automatically replaces the previous one:

```go
// First strategy: auto-validate all
agent.SetToolCallCallback(func(opID, msg string) error {
    return agent.ValidateOperation(opID)
})

// Later: switch to whitelist strategy
agent.SetToolCallCallback(func(opID, msg string) error {
    if isWhitelisted(msg) {
        return agent.ValidateOperation(opID)
    }
    return agent.CancelOperation(opID)
})
```

## Related Examples

- [50-server-agent-with-tools](../50-server-agent-with-tools) - Server that this client connects to
- [51-remote-agent-stream](../51-remote-agent-stream) - Basic remote agent usage
- [52-remote-interactive](../52-remote-interactive) - Interactive operation management
