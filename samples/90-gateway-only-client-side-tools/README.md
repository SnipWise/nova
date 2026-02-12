# Example 90: Gateway with Client-Side Tool Execution

This example demonstrates a **minimal gateway setup** that supports client-side tool execution. This is the simplest possible configuration for supporting tools like those used by qwen-code, aider, continue.dev, and other AI coding assistants.

## What's Included

- ✅ **1 Chat Agent**: Handles all regular conversations
- ✅ **1 Client-Side Tools Agent**: Detects tool calls and returns them to the client
- ❌ **No Orchestrator**: Single agent, no routing needed
- ❌ **No Compressor**: Keep it simple
- ❌ **No RAG**: Pure conversation

## How It Works

```
┌─────────────┐
│   Client    │ (qwen-code, aider, etc.)
│             │
│  1. Send    │ Request with tools definitions
│     ↓       │
│  2. Recv    │ tool_calls in OpenAI format
│     ↓       │
│  3. Execute │ Tools locally
│     ↓       │
│  4. Send    │ Tool results (role: "tool")
│     ↓       │
│  5. Recv    │ Final response
└─────────────┘
       ↕
┌─────────────┐
│   Gateway   │
│             │
│ • Detects   │ Tool calls needed?
│ • Returns   │ tool_calls to client
│ • Continues │ After tool results
└─────────────┘
```

## Architecture

```go
Gateway
├── ChatAgent (assistant)          // Handles conversations
└── ClientSideToolsAgent           // Detects tool calls
    └── Returns tool_calls to client
        └── Client executes tools
            └── Client sends results back
                └── Gateway continues with results
```

## Running the Example

### 1. Start your LLM backend

```bash
# Example with llama.cpp server
llama-server -m /path/to/model.gguf --port 12434
```

### 2. Start the gateway

```bash
cd samples/90-gateway-only-client-side-tools
go run main.go
```

### 3. Use with a tool-capable client

**With qwen-code:**
```bash
OPENAI_BASE_URL=http://localhost:8080/v1 \
OPENAI_API_KEY=none \
OPENAI_MODEL=assistant \
qwen-code
```

**With aider:**
```bash
OPENAI_API_BASE=http://localhost:8080/v1 \
OPENAI_API_KEY=none \
aider --model assistant
```

**With curl (test without tools):**
```bash
curl http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "assistant",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENGINE_URL` | `http://localhost:12434/engines/llama.cpp/v1` | LLM backend URL |
| `MODEL_ID` | `hf.co/menlo/jan-nano-gguf:q4_k_m` | Chat agent model |
| `CLIENT_TOOLS_MODEL_ID` | Same as `MODEL_ID` | Tools detection model |
| `NOVA_LOG_LEVEL` | `INFO` | Log level |

## Key Differences from Example 89

| Feature | Example 89 (Full) | Example 90 (Minimal) |
|---------|-------------------|----------------------|
| Chat Agents | Multiple (coder, generic) | Single (assistant) |
| Orchestrator | ✅ Yes | ❌ No |
| Compressor | ✅ Yes | ❌ No |
| RAG | ✅ Yes | ❌ No |
| Client Tools | ✅ Yes | ✅ Yes |
| Complexity | High (production-ready) | Low (learning/simple use) |

## When to Use This Example

Use this minimal setup when:
- ✅ You want the simplest possible client-side tool calling setup
- ✅ You're learning how client-side tools work
- ✅ You have a single-purpose assistant
- ✅ You don't need routing or orchestration
- ✅ You want minimal overhead

Use Example 89 (full) when:
- ⚠️ You need multiple specialized agents
- ⚠️ You need topic-based routing
- ⚠️ You need context compression
- ⚠️ You need RAG capabilities
- ⚠️ You're building a production system

## Code Structure

```go
// 1. Create chat agent
chatAgent := chat.NewAgent(...)

// 2. Create client-side tools agent
clientSideToolsAgent := tools.NewAgent(...)

// 3. Create gateway with minimal config
gateway := gatewayserver.NewAgent(
    ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithClientSideToolsAgent(clientSideToolsAgent),
    gatewayserver.WithPort(8080),
)

// 4. Start server
gateway.StartServer()
```

That's it! No complexity, no extra features, just pure client-side tool calling.

## Testing Tool Calls

You can test tool calling by using a client like qwen-code or aider with file operations:

```bash
# With qwen-code
OPENAI_BASE_URL=http://localhost:8080/v1 \
OPENAI_API_KEY=none \
OPENAI_MODEL=assistant \
qwen-code

# Then ask it to perform file operations:
> Create a new file hello.go with a Hello World function
```

The gateway will:
1. Detect that tools are needed (file operations)
2. Return tool_calls to qwen-code
3. qwen-code executes the file operations
4. qwen-code sends results back
5. Gateway continues with the completion

## Next Steps

- See Example 89 for a full-featured gateway with orchestration
- See the N.O.V.A. SDK documentation for more configuration options
- Try adding server-side tools with `WithToolsAgent()`
- Experiment with custom agent execution order with `WithAgentExecutionOrder()`
