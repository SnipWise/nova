# Server Agent with Tools

This example demonstrates how to create a server agent that exposes an HTTP API for chat completions with tool calling capabilities.

## Features

- HTTP REST API for chat completions
- Server-Sent Events (SSE) streaming
- Parallel tool calls with web-based confirmation
- Tools: `calculate_sum`, `say_hello`, `say_exit`

## Prerequisites

- Ollama or compatible LLM server running on `http://localhost:12434`
- The model `hf.co/menlo/jan-nano-gguf:q4_k_m` available

## Running the Server

```bash
cd samples/49-server-agent-with-tools
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Chat Completion (Streaming)

```bash
POST /completion
Content-Type: application/json

{
  "data": {
    "message": "Make the sum of 40 and 2"
  }
}
```

### Control Endpoints

- `POST /completion/stop` - Stop current streaming
- `POST /memory/reset` - Reset conversation history
- `GET /memory/messages/list` - Get all messages
- `GET /memory/messages/context-size` - Get token count
- `GET /models` - Get model information
- `GET /health` - Health check

### Tool Call Operations

- `POST /operation/validate` - Validate a pending tool call
- `POST /operation/cancel` - Cancel a pending tool call
- `POST /operation/reset` - Cancel all pending operations

## Tool Call Flow

1. Client sends a message that triggers tool calls
2. Server detects tool calls and creates pending operations
3. Server sends SSE notification with `operation_id`
4. Client validates or cancels the operation via `/operation/validate` or `/operation/cancel`
5. Server executes the tool and continues with the response

## Example with curl

```bash
# Start the server first
go run main.go

# In another terminal, send a request
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data":{"message":"Say hello to Alice and Bob"}}'
```

## Related Examples

- [51-remote-agent-stream](../51-remote-agent-stream) - Basic remote client with manual operation management
- [52-remote-interactive](../52-remote-interactive) - Interactive CLI for operation management
- [53-remote-programmatic](../53-remote-programmatic) - Automated operation handling
