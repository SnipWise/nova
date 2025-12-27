---
id: basic-remote
name: Basic Remote Agent
category: remote
complexity: beginner
sample_source: 71
description: Client agent that connects to a remote server agent via HTTP
---

# Basic Remote Agent

## Description

Creates a remote agent client that connects to a Server Agent via HTTP. The remote agent provides the same interface as local agents but executes operations on a remote server, enabling distributed AI architectures.

## Use Cases

- Distributed AI systems
- Client-server AI architectures
- Microservices with AI capabilities
- Centralized AI processing
- Load-balanced AI inference

## Complete Code

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/remote"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// === REMOTE AGENT CONFIGURATION ===
	agent, err := remote.NewAgent(
		ctx,
		agents.Config{
			Name: "Interactive Remote Client",  // Client name
		},
		"http://localhost:3500",  // Server URL
	)
	if err != nil {
		panic(err)
	}

	// Display connection info
	display.Colorf(display.ColorCyan, "🌐 Connected to remote agent at %s\n", "http://localhost:3500")
	display.Colorf(display.ColorCyan, "Agent: %s\n", agent.GetName())
	display.Colorf(display.ColorCyan, "Model: %s\n\n", agent.GetModelID())

	// Interactive loop
	for {
		input := prompt.NewWithColor("🤖 Ask me something?").
			SetMessageColor(prompt.ColorBrightCyan).
			SetInputColor(prompt.ColorBrightWhite)

		question, err := input.Run()
		if err != nil {
			log.Fatal(err)
		}

		// Exit command
		if strings.HasPrefix(question, "/bye") {
			fmt.Println("Goodbye!")
			break
		}

		// Stream completion from remote server
		result, err := agent.GenerateStreamCompletion(
			[]messages.Message{
				{
					Role:    roles.User,
					Content: question,
				},
			},
			func(chunk string, finishReason string) error {
				// Print streaming chunks
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
			panic(err)
		}

		// Display metadata
		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Context size", fmt.Sprintf("%d characters", agent.GetContextSize()))
		display.Separator()
	}
}
```

## Configuration

```yaml
SERVER_URL: "http://localhost:3500"
CLIENT_NAME: "Interactive Remote Client"
```

## How It Works

### 1. Connection
```go
agent, err := remote.NewAgent(ctx, config, serverURL)
```

The remote agent connects to a Server Agent running on the specified URL.

### 2. Same Interface as Local Agents
```go
// All these methods work remotely:
agent.GenerateCompletion(messages)
agent.GenerateStreamCompletion(messages, callback)
agent.GetMessages()
agent.GetContextSize()
agent.ResetMessages()
```

### 3. Server-Sent Events (SSE)
The remote agent uses SSE for real-time streaming from the server.

## Server Requirements

The remote agent requires a Server Agent to be running:

```go
// Server side (must be running first)
serverAgent, _ := server.NewAgent(ctx, config, models, ":3500")
serverAgent.StartServer()
```

See `server/basic-server.md` for server setup.

## Customization

### Non-Streaming Mode

```go
// Use GenerateCompletion for non-streaming
result, err := agent.GenerateCompletion(
	[]messages.Message{
		{Role: roles.User, Content: "Hello!"},
	},
)
fmt.Println(result.Response)
```

### Conversation History

```go
// Get all messages
messages := agent.GetMessages()
for _, msg := range messages {
	fmt.Printf("%s: %s\n", msg.Role, msg.Content)
}

// Check context size
contextSize := agent.GetContextSize()
fmt.Printf("Context: %d characters\n", contextSize)

// Reset conversation
agent.ResetMessages()
```

### Custom Streaming Handler

```go
var fullResponse strings.Builder

agent.GenerateStreamCompletion(
	messages,
	func(chunk string, finishReason string) error {
		// Accumulate chunks
		fullResponse.WriteString(chunk)

		// Display with color
		fmt.Print(colorize(chunk))

		// Log finish reason
		if finishReason != "" {
			log.Printf("Finished: %s", finishReason)
		}

		return nil
	},
)

fmt.Println("\nFull response:", fullResponse.String())
```

### Different Server URL

```go
// Production server
agent, _ := remote.NewAgent(ctx, config, "https://api.example.com")

// Local development
agent, _ := remote.NewAgent(ctx, config, "http://localhost:8080")

// Environment variable
serverURL := os.Getenv("SERVER_URL")
agent, _ := remote.NewAgent(ctx, config, serverURL)
```

## Interactive Commands

Add custom commands to the interactive loop:

```go
for {
	question, _ := input.Run()

	switch {
	case strings.HasPrefix(question, "/bye"):
		fmt.Println("Goodbye!")
		return

	case strings.HasPrefix(question, "/reset"):
		agent.ResetMessages()
		fmt.Println("Conversation reset")
		continue

	case strings.HasPrefix(question, "/history"):
		displayHistory(agent.GetMessages())
		continue

	case strings.HasPrefix(question, "/export"):
		json, _ := agent.ExportMessagesToJSON()
		fmt.Println(json)
		continue

	default:
		// Normal query processing
		agent.GenerateStreamCompletion(...)
	}
}
```

## Error Handling

```go
agent, err := remote.NewAgent(ctx, config, serverURL)
if err != nil {
	log.Fatalf("Failed to connect to server: %v", err)
}

// Check connection
if !agent.IsHealthy() {
	log.Fatal("Server is not healthy")
}

// Handle streaming errors
_, err = agent.GenerateStreamCompletion(messages, func(chunk, reason string) error {
	if chunk == "ERROR" {
		return fmt.Errorf("server error detected")
	}
	fmt.Print(chunk)
	return nil
})
if err != nil {
	log.Printf("Streaming error: %v", err)
}
```

## Health Check

```go
// Check if server is available
if agent.IsHealthy() {
	fmt.Println("✅ Server is healthy")
} else {
	fmt.Println("❌ Server is not available")
}

// Get server models info
modelsInfo, err := agent.GetModelsInfo()
if err == nil {
	fmt.Printf("Chat Model: %s\n", modelsInfo.ChatModel)
	fmt.Printf("Tools Model: %s\n", modelsInfo.ToolsModel)
}
```

## Important Notes

- **Server must be running** before creating remote agent
- Remote agent uses the **same interface** as local agents
- **SSE streaming** provides real-time responses
- Connection is **stateless** - conversation state is on server
- **Network latency** affects response time
- Server handles all **model inference** and **memory management**
- Remote agent is a **thin client** - minimal resource usage

## Performance Considerations

- **Latency**: Network round-trip time affects response speed
- **Streaming**: Use streaming for better perceived performance
- **Connection pooling**: Reuse agent instance for multiple queries
- **Error handling**: Network failures require retry logic
- **Timeouts**: Set appropriate timeout values for long-running queries

## Use with Server Features

### Tools (Function Calling)

If the server has tools enabled, the remote agent will receive tool notifications:

```go
// Server will send tool notifications via SSE
// Client must validate operations (see server-with-tools.md)
```

### RAG (Document Retrieval)

If the server has RAG enabled, documents are automatically retrieved:

```go
// RAG happens server-side transparently
// Client just receives enhanced responses
```

### Context Compression

If the server has compression enabled, compression happens automatically:

```go
// Compression is managed by the server
// Client sees reduced context size after compression
```

## Related Patterns

- For server setup: See `server/basic-server.md`
- For advanced features: See `remote/advanced-remote.md`
- For programmatic usage: See `remote/programmatic-remote.md`
