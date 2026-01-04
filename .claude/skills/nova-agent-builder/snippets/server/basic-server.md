---
id: basic-server
name: Basic Server Agent
category: server
complexity: intermediate
sample_source: 70
description: HTTP/REST API server agent with streaming support via SSE
---

# Basic Server Agent

## Description

Creates an HTTP server agent that exposes a conversational AI through REST API endpoints with Server-Sent Events (SSE) streaming for real-time responses.

## Use Cases

- REST API for chat applications
- Web-based conversational interfaces
- Microservices architecture
- Real-time streaming chat responses
- Server-side AI integration

## Complete Code

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// === CONFIGURATION - CUSTOMIZE HERE ===
	agent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:                    "bob-server-agent",                              // Agent name
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",   // LLM Engine URL
			SystemInstructions:      "You are Bob, a helpful AI assistant.",          // System instructions
			KeepConversationHistory: true,                                            // Keep conversation context
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",    // Model to use
			Temperature: models.Float64(0.4),                    // Creativity level
		},
		":3500",  // HTTP port
		// executeFunction is optional - omitted here, will use default
	)
	if err != nil {
		panic(err)
	}

	// Start the HTTP server
	fmt.Printf("🚀 Starting server agent on http://localhost%s\n", agent.GetPort())
	log.Fatal(agent.StartServer())
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TEMPERATURE: 0.4
PORT: ":3500"
```

## API Endpoints

### POST /completion

Stream a chat completion with SSE.

**Request:**

```bash
curl -N -X POST http://localhost:3500/completion \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "message": "Hello! How are you?"
    }
  }'
```

**Response (SSE Stream):**

```
data: {"chunk": "Hello"}
data: {"chunk": "!"}
data: {"chunk": " I"}
data: {"chunk": "'m"}
data: {"finish_reason": "stop"}
```

### POST /completion/stop

Stop the current streaming operation.

```bash
curl -X POST http://localhost:3500/completion/stop
```

### POST /memory/reset

Clear conversation history.

```bash
curl -X POST http://localhost:3500/memory/reset
```

### GET /memory/messages/list

Get all conversation messages.

```bash
curl http://localhost:3500/memory/messages/list
```

### GET /memory/messages/context-size

Get token count for the conversation.

```bash
curl http://localhost:3500/memory/messages/context-size
```

### GET /health

Health check endpoint.

```bash
curl http://localhost:3500/health
```

## Customization

### Custom Port

```go
agent.SetPort(":8080")
```

### Environment Variables

```go
import "github.com/joho/godotenv"

func main() {
    godotenv.Load()

    engineURL := os.Getenv("ENGINE_URL")
    modelName := os.Getenv("MODEL_NAME")
    port := os.Getenv("SERVER_PORT")

    agent, err := server.NewAgent(
        ctx,
        agents.Config{
            EngineURL: engineURL,
            // ...
        },
        models.Config{
            Name: modelName,
        },
        port,
    )
}
```

### Using Different Models

```go
// For better performance
models.Config{
    Name:        "ai/qwen2.5:1.5B-F16",
    Temperature: models.Float64(0.7),
}

// For function calling support
models.Config{
    Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
    Temperature: models.Float64(0.0),
}
```

## Client Example (JavaScript)

```javascript
const evtSource = new EventSource("http://localhost:3500/completion", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    data: { message: "Hello!" },
  }),
});

evtSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.chunk) {
    console.log(data.chunk);
  }
  if (data.finish_reason) {
    evtSource.close();
  }
};
```

## Important Notes

### DO:
- Use `server.NewAgent()` to create HTTP API agents
- Set `KeepConversationHistory: true` for stateful conversations
- Call `agent.StartServer()` to start HTTP server (blocking)
- Use SSE endpoints for streaming responses
- Set appropriate temperature (0.0-0.8 depending on use case)
- Enable logging with `NOVA_LOG_LEVEL=INFO` or `DEBUG`
- Use `/memory/reset` to clear history between sessions
- Include colon in port format: `:3500` not `3500`

### DON'T:
- Don't forget the colon prefix for ports (`:3500`)
- Don't use very high temperature (> 0.9) for consistent responses
- Don't ignore errors from `StartServer()` (use `log.Fatal()`)
- Don't skip health check endpoint for production monitoring
- Don't forget to handle graceful shutdown in production
- Don't expose server without authentication in production

### Dual-Mode Pattern:
The `server.Agent` can run in **two modes**:
1. **HTTP Mode**: Call `agent.StartServer()` for REST API
2. **CLI Mode**: Call `agent.StreamCompletion()` for interactive terminal

See `dual-mode-server.md` for complete dual-mode implementation.

### Production Deployment:
```go
// Use environment variables for configuration
import "github.com/snipwise/nova/nova-sdk/toolbox/env"

engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
modelName := env.GetEnvOrDefault("MODEL_NAME", "hf.co/menlo/jan-nano-gguf:q4_k_m")
port := env.GetEnvOrDefault("SERVER_PORT", ":3500")
```

### Key Methods:
```go
// Start HTTP server (blocking)
agent.StartServer()

// Get current port
port := agent.GetPort()

// Set custom port
agent.SetPort(":8080")

// Stream completion (CLI mode)
agent.StreamCompletion(question, callbackFunc)
```

## Related Patterns

- **Dual-mode agent**: See `dual-mode-server.md` (CLI + HTTP modes)
- **Tools support**: See `server-with-tools.md`
- **RAG support**: See `server-with-rag.md`
- **Context compression**: See `server-with-compressor.md`
- **Full-featured server**: See `server-full-featured.md` (tools + RAG + compression)
