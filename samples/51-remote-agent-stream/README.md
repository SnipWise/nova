# Remote Agent Stream

This example demonstrates how to use a remote agent to connect to a server agent via HTTP and perform streaming chat completions.

## Features

- Connect to remote server agent via HTTP
- Streaming completions with real-time output
- Non-streaming completions
- Access conversation history from server
- Export conversation to JSON
- Reset conversation remotely
- Programmatic operation management (validate/cancel tool calls)

## Prerequisites

- Server agent running on `http://localhost:8080` (see example 50)
- Server must be started before running this client

## Running the Client

First, start the server in another terminal:
```bash
cd ../50-server-agent-with-tools
go run main.go
```

Then run the client:
```bash
go run main.go
```

## What the Client Does

The example demonstrates three types of interactions:

### 1. Simple Streaming Question
Asks a simple question and streams the response:
```go
agent.GenerateStreamCompletion(messages, callback)
```

### 2. Tool Calls with Confirmation
Sends a message that triggers multiple tool calls:
- The server will detect the tool calls
- Client displays operation IDs and validation commands
- You can validate using scripts or curl commands
- The client receives the streamed response

### 3. Non-Streaming Completion
Sends a question and waits for the complete response:
```go
result, err := agent.GenerateCompletion(messages)
```

## Remote Agent API

The remote agent provides the same interface as the chat agent, plus operation management methods:

### Basic Operations
```go
// Create remote agent
agent, err := remote.NewAgent(ctx, agentConfig, baseURL)

// Check server health
if agent.IsHealthy() {
    fmt.Println("Server is ready")
}

// Get detailed models information
modelsInfo, err := agent.GetModelsInfo()
fmt.Printf("Chat Model: %s\n", modelsInfo.ChatModel)

// Generate streaming completion
result, err := agent.GenerateStreamCompletion(messages, callback)

// Generate non-streaming completion
result, err := agent.GenerateCompletion(messages)

// Get conversation history from server
messages := agent.GetMessages()

// Get context size from server
tokens := agent.GetContextSize()

// Reset conversation on server
agent.ResetMessages()

// Export to JSON
json, err := agent.ExportMessagesToJSON()
```

### Operation Management
```go
// Validate a pending tool call operation
err := agent.ValidateOperation(operationID)

// Cancel a pending tool call operation
err := agent.CancelOperation(operationID)

// Cancel all pending operations
err := agent.ResetOperations()
```

## Tool Call Flow

When the client sends a message that triggers tool calls:

1. Client sends the message to the server
2. Server detects tool calls and sends SSE notifications with `operation_id`
3. Client displays the operation ID and validation commands
4. You validate or cancel operations using the provided scripts or curl commands
5. Server executes approved tools and continues streaming
6. Final result includes tool call outputs

### Validating Tool Calls

The client will display information like this when a tool call is detected:

```
🔔 Tool Call Detected: Tool call detected: say_hello
📝 Operation ID: op_0x14000126020
✅ To validate: curl -X POST http://localhost:8080/operation/validate -d '{"operation_id":"op_0x14000126020"}'
⛔️ To cancel:   curl -X POST http://localhost:8080/operation/cancel -d '{"operation_id":"op_0x14000126020"}'
```

**Option 1: Use the provided scripts (easiest)**
```bash
# In another terminal
cd samples/51-remote-agent-stream

# Validate an operation
./validate-operation.sh op_0x14000126020

# Or cancel an operation
./cancel-operation.sh op_0x14000126020
```

**Option 2: Use curl directly**
```bash
# Validate
curl -X POST http://localhost:8080/operation/validate \
  -H "Content-Type: application/json" \
  -d '{"operation_id":"op_0x14000126020"}'

# Cancel
curl -X POST http://localhost:8080/operation/cancel \
  -H "Content-Type: application/json" \
  -d '{"operation_id":"op_0x14000126020"}'
```

**Option 3: Use the remote agent API (programmatic)**
```go
// In your code
err := agent.ValidateOperation("op_0x14000126020")
// or
err := agent.CancelOperation("op_0x14000126020")
```

## Helper Scripts

- `run-demo.sh` - Checks server and runs main.go
- `validate-operation.sh <op_id>` - Validate specific operation
- `cancel-operation.sh <op_id>` - Cancel specific operation

## How It Works

1. The remote agent makes HTTP requests to the server agent
2. For streaming, it parses Server-Sent Events (SSE)
3. All conversation state is maintained on the server
4. The client is stateless and can reconnect at any time

## Related Examples

- [50-server-agent-with-tools](../50-server-agent-with-tools) - Server that this client connects to
- [52-remote-interactive](../52-remote-interactive) - Interactive CLI for operation management
- [53-remote-programmatic](../53-remote-programmatic) - Automated operation handling
- [48-parallel-toolcalls-tool-agent-with-confirmation](../48-parallel-toolcalls-tool-agent-with-confirmation) - Local version without HTTP

## For More Details

- See [API.md](API.md) for complete API reference
- See [USAGE.md](USAGE.md) for quick start guide
