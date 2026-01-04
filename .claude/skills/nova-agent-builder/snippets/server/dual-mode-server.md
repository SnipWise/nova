---
id: dual-mode-server
name: Dual Mode Server Agent (CLI + HTTP)
category: server
complexity: advanced
sample_source: 70
description: Server agent that works in both CLI interactive mode and HTTP API mode
---

# Dual Mode Server Agent (CLI + HTTP)

## Description

Creates a versatile Nova server agent that can run in two modes:
1. **CLI Mode**: Interactive terminal chat with StreamCompletion
2. **HTTP Mode**: REST API server with SSE streaming endpoints

The same agent code works in both modes, making it perfect for development (CLI) and production (HTTP).

## Use Cases

- Development: Test agents interactively in CLI
- Production: Deploy same agent as HTTP API
- Demos: Quick CLI demonstrations
- Testing: CLI for unit tests, HTTP for integration tests
- Unified codebase for chat and API modes

## Complete Code

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// === CREATE SERVER AGENT ===
	// Works in both CLI and HTTP modes
	agent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
		":3500",         // Port for HTTP mode (ignored in CLI mode)
		executeFunction, // Tool executor
	)
	if err != nil {
		panic(err)
	}

	// === OPTIONAL: ADD TOOLS AGENT ===
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-tools",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true),
		},
		tools.WithTools(GetToolsIndex()),
	)
	if err != nil {
		panic(err)
	}
	agent.SetToolsAgent(toolsAgent)

	// === OPTIONAL: ADD RAG AGENT ===
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "rag-agent",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
		},
		models.Config{
			Name: "ai/mxbai-embed-large",
		},
	)
	if err != nil {
		panic(err)
	}

	// Load documents with chunking
	contents, err := files.GetContentFiles("./data", ".md")
	if err != nil {
		panic(err)
	}

	for idx, content := range contents {
		// Split by markdown sections for better chunking
		piecesOfDoc := chunks.SplitMarkdownBySections(content)

		for chunkIdx, piece := range piecesOfDoc {
			display.Colorf(display.ColorYellow,
				"generating vectors... (docs %d/%d) [chunks: %d/%d]\n",
				idx+1, len(contents), chunkIdx+1, len(piecesOfDoc))

			err := ragAgent.SaveEmbedding(piece)
			if err != nil {
				display.Errorf("failed to save embedding: %v\n", err)
			}
		}
	}
	agent.SetRagAgent(ragAgent)

	// === CLI MODE EXECUTION ===
	// Interactive loop using StreamCompletion
	fmt.Println("🤖 Server Agent in CLI Mode with StreamCompletion")
	fmt.Println("Type 'exit' to quit, 'server' to switch to HTTP mode")
	fmt.Println("---")

	for {
		input := prompt.NewWithColor("🧑 You: ")
		question, err := input.RunWithEdit()
		if err != nil {
			display.Errorf("Error reading input: %v", err)
			continue
		}

		if question == "exit" {
			display.Infof("👋 Goodbye!")
			break
		}

		if question == "server" {
			// Switch to HTTP server mode
			display.Infof("Starting HTTP server on :3500...")
			if err := agent.StartServer(); err != nil {
				display.Errorf("Server error: %v", err)
			}
			break
		}

		if question == "" {
			continue
		}

		// Use StreamCompletion method (same API as crew agent)
		fmt.Print("🤖 Bob: ")
		_, err = agent.StreamCompletion(question, func(chunk string, finishReason string) error {
			if chunk != "" {
				fmt.Print(chunk)
			}
			if finishReason == "stop" {
				fmt.Println()
			}
			return nil
		})

		if err != nil {
			display.Errorf("❌ Error: %v", err)
		}
	}
}

// === TOOL DEFINITIONS ===
func GetToolsIndex() []*tools.Tool {
	calculateSumTool := tools.NewTool("calculate_sum").
		SetDescription("Calculate the sum of two numbers").
		AddParameter("a", "number", "The first number", true).
		AddParameter("b", "number", "The second number", true)

	sayHelloTool := tools.NewTool("say_hello").
		SetDescription("Say hello to the given name").
		AddParameter("name", "string", "The name to greet", true)

	getCurrentTimeTool := tools.NewTool("get_current_time").
		SetDescription("Get the current time")

	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
		getCurrentTimeTool,
	}
}

// === TOOL EXECUTOR ===
func executeFunction(functionName string, arguments string) (string, error) {
	fmt.Printf("\n🔧 Executing: %s\n", functionName)

	switch functionName {
	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("👋 Hello, %s! Nice to meet you!", args.Name)
		return fmt.Sprintf(`{"message": "%s"}`, hello), nil

	case "calculate_sum":
		var args struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for calculate_sum"}`, nil
		}
		sum := args.A + args.B
		return fmt.Sprintf(`{"result": %g}`, sum), nil

	case "get_current_time":
		return `{"time": "2025-01-01 12:00:00 UTC"}`, nil

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
```

## Configuration

```yaml
# Agent Configuration
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
CHAT_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TOOLS_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
EMBEDDING_MODEL: "ai/mxbai-embed-large"

# Server Configuration
SERVER_PORT: ":3500"
NOVA_LOG_LEVEL: "INFO"

# Temperature Settings
CHAT_TEMPERATURE: 0.4
TOOLS_TEMPERATURE: 0.0
```

## Key API

### server.NewAgent

```go
import "github.com/snipwise/nova/nova-sdk/agents/server"

// Create dual-mode agent
agent, err := server.NewAgent(
    ctx,
    agents.Config{
        Name:               "agent-name",
        EngineURL:          engineURL,
        SystemInstructions: instructions,
    },
    models.Config{
        Name:        modelName,
        Temperature: models.Float64(0.4),
    },
    ":3500",         // HTTP port (ignored in CLI mode)
    executeFunction, // Tool executor function
)
```

### CLI Mode: StreamCompletion

```go
// Interactive streaming chat (CLI mode)
_, err = agent.StreamCompletion(
    userQuestion,
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
```

### HTTP Mode: StartServer

```go
// Start HTTP server (blocking)
if err := agent.StartServer(); err != nil {
    log.Fatal(err)
}
```

## HTTP Endpoints

When running in server mode, the following endpoints are available:

### POST /chat
```bash
curl -X POST http://localhost:3500/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, who are you?"}'
```

### POST /chat/stream (SSE)
```bash
curl -X POST http://localhost:3500/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "Tell me about Go programming"}' \
  --no-buffer
```

### GET /health
```bash
curl http://localhost:3500/health
```

## Customization

### Environment-Based Mode Selection

```go
import (
    "github.com/snipwise/nova/nova-sdk/toolbox/env"
)

func main() {
    // ...agent setup...

    mode := env.GetEnvOrDefault("AGENT_MODE", "cli")

    if mode == "server" {
        // HTTP Server Mode
        display.Infof("Starting HTTP server on :3500...")
        if err := agent.StartServer(); err != nil {
            log.Fatal(err)
        }
    } else {
        // CLI Mode
        runCLILoop(agent)
    }
}

func runCLILoop(agent *server.Agent) {
    for {
        question := prompt.New("You: ").Run()
        if question == "exit" {
            break
        }

        fmt.Print("Bot: ")
        agent.StreamCompletion(question, func(chunk, _ string) error {
            fmt.Print(chunk)
            return nil
        })
        fmt.Println()
    }
}
```

### Custom Confirmation Prompt (CLI Only)

```go
// Optional: Set custom confirmation for tool execution in CLI mode
agent.SetConfirmationPromptFunction(customConfirmationPrompt)

func customConfirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
    display.Colorf(display.ColorYellow, "⚠️  Tool call: %s\n", functionName)
    display.Infof("Arguments: %s", arguments)

    choice := prompt.HumanConfirmation(fmt.Sprintf("Execute %s?", functionName))
    return choice
}
```

### Docker Deployment

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o agent main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/agent .
COPY ./data ./data

# Environment configuration
ENV AGENT_MODE=server
ENV SERVER_PORT=:3500
ENV ENGINE_URL=http://host.docker.internal:12434/engines/llama.cpp/v1

EXPOSE 3500
CMD ["./agent"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  agent:
    build: .
    ports:
      - "3500:3500"
    environment:
      - AGENT_MODE=server
      - ENGINE_URL=http://host.docker.internal:12434/engines/llama.cpp/v1
      - CHAT_MODEL=hf.co/menlo/jan-nano-gguf:q4_k_m
      - EMBEDDING_MODEL=ai/mxbai-embed-large
    volumes:
      - ./data:/app/data
    restart: unless-stopped
```

## Advanced Features

### Multi-Agent Setup

```go
// Create multiple specialized agents
chatAgent, _ := server.NewAgent(ctx, chatConfig, chatModel, ":3501", nil)
toolsAgent, _ := server.NewAgent(ctx, toolsConfig, toolsModel, ":3502", executeTools)
ragAgent, _ := server.NewAgent(ctx, ragConfig, ragModel, ":3503", nil)

// Run in separate goroutines
go chatAgent.StartServer()
go toolsAgent.StartServer()
go ragAgent.StartServer()
```

### Graceful Shutdown

```go
import (
    "os/signal"
    "syscall"
)

func main() {
    // ...agent setup...

    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Start server in goroutine
    go func() {
        if err := agent.StartServer(); err != nil {
            log.Println("Server error:", err)
        }
    }()

    // Wait for shutdown signal
    <-sigChan
    log.Println("Shutting down gracefully...")

    // Cleanup
    agent.Close()
}
```

## Important Notes

### DO:
- Use `server.NewAgent()` for dual-mode capability
- Call `agent.StreamCompletion()` for CLI interactive mode
- Call `agent.StartServer()` for HTTP server mode
- Set port with `:3500` format (colon required)
- Use environment variables to switch between modes
- Test in CLI mode before deploying to HTTP
- Add RAG and tools agents for full functionality
- Use `chunks.SplitMarkdownBySections()` for better document chunking

### DON'T:
- Don't call both `StreamCompletion()` loop and `StartServer()` simultaneously
- Don't forget the colon in port (`:3500` not `3500`)
- Don't skip error handling for tool execution
- Don't ignore finish reason in streaming callbacks
- Don't forget to set `AGENT_MODE` environment variable in production

## CLI vs HTTP Comparison

| Feature | CLI Mode | HTTP Mode |
|---------|----------|-----------|
| Method | `StreamCompletion()` | `StartServer()` |
| Input | Terminal prompt | HTTP POST /chat |
| Output | stdout streaming | SSE streaming |
| Tools | Interactive confirmation | Auto-execute |
| Use Case | Development, demos | Production API |
| Port | Ignored | Required |

## Troubleshooting

### Port Already in Use
```bash
# Check what's using the port
lsof -i :3500

# Kill the process
kill -9 <PID>

# Or change port in config
SERVER_PORT=:3501
```

### RAG Documents Not Loading
```go
// Verify data directory exists
if _, err := os.Stat("./data"); os.IsNotExist(err) {
    log.Fatal("./data directory not found")
}

// Check file count
contents, _ := files.GetContentFiles("./data", ".md")
fmt.Printf("Loaded %d documents\n", len(contents))
```

### Tools Not Executing
```go
// Verify tools agent is set
if agent.GetToolsAgent() == nil {
    log.Println("Warning: No tools agent set")
}

// Enable debug logging
os.Setenv("NOVA_LOG_LEVEL", "DEBUG")
```

## Mode Selection Examples

```bash
# CLI Mode (development)
go run main.go

# HTTP Mode (production)
AGENT_MODE=server go run main.go

# Docker CLI Mode (debugging)
docker run -it agent /bin/sh

# Docker HTTP Mode (production)
docker compose up -d
```
