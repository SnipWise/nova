# Examples Overview

This directory contains the basic remote agent example with manual tool call management.

## main.go - Basic Remote Agent Usage

**What it demonstrates:**
- Simple streaming completions
- Non-streaming completions
- Tool calls with manual validation (via curl/scripts)
- Getting conversation history
- Exporting to JSON
- Resetting conversation

**Run:**
```bash
go run main.go
```

## Related Examples

For more advanced usage patterns, see:

### [52-remote-interactive](../52-remote-interactive)
Interactive command-line interface for managing operations in real-time.

**Features:**
- CLI commands for validation/cancellation
- Real-time operation tracking
- No need to switch terminals

**Use when:** You want interactive control over operations without using curl commands.

### [53-remote-programmatic](../53-remote-programmatic)
Automated operation handling with custom logic.

**Features:**
- Auto-validation patterns
- Auto-cancellation patterns
- Template for building custom approval workflows

**Use when:** You want to build automated systems that handle operations programmatically.

## Use Cases by Example

### Use Case 1: Manual Approval (This Example)
Best for development and testing where you want full control over each operation.

```bash
# Terminal 1: Start server
cd ../50-server-agent-with-tools
go run main.go

# Terminal 2: Run client
cd ../51-remote-agent-stream
go run main.go

# Terminal 3: Validate operations
./validate-operation.sh op_0x12345
```

**When to use:**
- Learning how tool calls work
- Debugging issues
- Maximum control over each operation

### Use Case 2: Interactive Control (Example 52)
Best for scenarios where you want a simple CLI to manage operations.

```bash
cd ../52-remote-interactive
go run main.go

# In the interactive prompt:
> v op_0x12345  # Validate
> c op_0x67890  # Cancel
> r             # Reset all
```

**When to use:**
- Development with frequent tool testing
- Manual approval workflow
- Security-sensitive operations

### Use Case 3: Automated Approval (Example 53)
Best for production scenarios with custom approval logic.

```bash
cd ../53-remote-programmatic
go run main.go
```

**When to use:**
- Production systems
- Auto-approving safe operations
- Building custom workflows
- Integration with existing systems

## Custom Implementation Example

Here's how you might build your own auto-approval logic using the callback system:

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/remote"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

func main() {
    agent, _ := remote.NewAgent(
        context.Background(),
        agents.Config{Name: "Auto-Approver"},
        "http://localhost:8080",
    )

    // Set up auto-approval callback
    agent.SetToolCallCallback(func(operationID string, message string) error {
        // Extract function name from message
        functionName := extractFunctionName(message)

        if shouldApprove(functionName) {
            fmt.Printf("✅ Approving: %s (ID: %s)\n", functionName, operationID)
            return agent.ValidateOperation(operationID)
        } else {
            fmt.Printf("❌ Rejecting: %s (ID: %s)\n", functionName, operationID)
            return agent.CancelOperation(operationID)
        }
    })

    // Send a question that triggers tool calls
    messages := []messages.Message{{
        Content: "Say hello to Alice and calculate 10 + 20",
        Role:    roles.User,
    }}

    // Start streaming - operations will be handled automatically
    agent.GenerateStreamCompletion(messages, func(chunk string, finishReason string) error {
        if chunk != "" {
            fmt.Print(chunk)
        }
        return nil
    })
}

func shouldApprove(functionName string) bool {
    // Whitelist of safe operations
    safeOperations := []string{"say_hello", "calculate_sum"}

    for _, safe := range safeOperations {
        if strings.Contains(functionName, safe) {
            return true
        }
    }
    return false
}

func extractFunctionName(message string) string {
    // Extract function name from message like "Tool call detected: say_hello"
    parts := strings.Split(message, ": ")
    if len(parts) > 1 {
        return parts[1]
    }
    return ""
}
```

**Key Features:**
- Uses `SetToolCallCallback` to capture operations in real-time
- Implements whitelist-based approval logic
- Operations are validated or cancelled automatically
- No need for manual curl commands or terminal switching

## Helper Scripts

This example includes helper scripts for manual operation management:

- `run-demo.sh` - Checks server and runs main.go
- `validate-operation.sh <op_id>` - Validate specific operation
- `cancel-operation.sh <op_id>` - Cancel specific operation

## Tips

1. **For Development**: Use this example with the helper scripts
2. **For Interactive Testing**: See example 52
3. **For Production**: See example 53 for automation patterns
4. **For Debugging**: Enable logs with `os.Setenv("NOVA_LOG_LEVEL", "DEBUG")`

## API Documentation

See [API.md](API.md) for complete API reference including all operation management methods.

## Quick Start

See [USAGE.md](USAGE.md) for step-by-step guide to getting started.
