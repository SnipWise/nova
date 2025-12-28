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

### GET /memory/messages/tokens
Get token count for the conversation.

```bash
curl http://localhost:3500/memory/messages/tokens
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
const evtSource = new EventSource('http://localhost:3500/completion', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    data: { message: 'Hello!' }
  })
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

- The server agent wraps a chat agent and exposes it via HTTP
- SSE (Server-Sent Events) provides real-time streaming
- The `executeFunction` parameter is optional - uses default if omitted
- Conversation history is maintained in server memory
- Use `/memory/reset` to clear history between conversations
- Port must include the colon prefix (e.g., `:3500`)

## Related Patterns

- For tools support: See `server-with-tools.md`
- For RAG support: See `server-with-rag.md`
- For context compression: See `server-with-compressor.md`
- For full-featured server: See `server-full-featured.md`
