---
id: server-full-featured
name: Full-Featured Server Agent
category: server
complexity: advanced
sample_source: 54
description: Complete HTTP server agent with tools, RAG, and context compression
---

# Full-Featured Server Agent

## Description

Creates a complete HTTP server agent with all capabilities: function calling (tools), RAG (Retrieval-Augmented Generation), and automatic context compression. This is a production-ready configuration combining all server agent features.

## Use Cases

- Production AI API servers
- Enterprise chatbot backends
- Intelligent knowledge bases with actions
- Long-running conversational services
- Full-stack AI microservices

## Complete Code

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	os.Setenv("NOVA_LOG_LEVEL", "INFO")
	ctx := context.Background()

	// === 1. TOOLS AGENT ===
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "Bob Tools",
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true,
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

	// === 2. RAG AGENT ===
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name: "ai/mxbai-embed-large",
		},
	)
	if err != nil {
		panic(err)
	}

	// Index documents
	contents, err := files.GetContentFiles("./data", ".md")
	if err != nil {
		panic(err)
	}
	for idx, content := range contents {
		piecesOfDoc := chunks.SplitMarkdownBySections(content)
		for chunkIdx, piece := range piecesOfDoc {
			display.Colorf(display.ColorYellow,
				"generating vectors... (docs %d/%d) [chunks: %d/%d]\n",
				idx+1, len(contents), chunkIdx+1, len(piecesOfDoc))
			ragAgent.SaveEmbedding(piece)
		}
	}

	// === 3. COMPRESSOR AGENT ===
	// Best practice: Use Effective instructions and UltraShort prompts
	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:      "compressor-agent",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			// RECOMMENDED: Use Effective for balanced compression
			SystemInstructions: compressor.Instructions.Effective,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
		// RECOMMENDED: Use UltraShort for maximum token reduction
		compressor.WithCompressionPrompt(compressor.Prompts.UltraShort),
	)
	if err != nil {
		panic(err)
	}

	// === 4. SERVER AGENT (Main Chat) ===
	serverAgent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:                    "Bob",
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
		// Optional configuration via functional options
		server.WithPort(":8080"),
		server.WithExecuteFn(executeFunction),
		server.WithToolsAgent(toolsAgent),
		server.WithRagAgent(ragAgent),
		server.WithCompressorAgent(compressorAgent),
	)
	if err != nil {
		panic(err)
	}

	// === 5. CONFIGURE ===
	serverAgent.SetContextSizeLimit(3000)
	serverAgent.SetSimilarityLimit(0.6)
	serverAgent.SetMaxSimilarities(3)

	display.Colorf(display.ColorCyan, "🚀 Server starting on http://localhost%s\n", serverAgent.GetPort())
	if err := serverAgent.StartServer(); err != nil {
		panic(err)
	}
}

func GetToolsIndex() []*tools.Tool {
	return []*tools.Tool{
		tools.NewTool("calculate_sum").
			SetDescription("Calculate the sum of two numbers").
			AddParameter("a", "number", "The first number", true).
			AddParameter("b", "number", "The second number", true),
		tools.NewTool("say_hello").
			SetDescription("Say hello to the given name").
			AddParameter("name", "string", "The name to greet", true),
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	display.Colorf(display.ColorGreen, "🟢 Executing: %s(%s)\n", functionName, arguments)

	switch functionName {
	case "say_hello":
		var args struct{ Name string `json:"name"` }
		json.Unmarshal([]byte(arguments), &args)
		return fmt.Sprintf(`{"message": "👋 Hello, %s!"}`, args.Name), nil

	case "calculate_sum":
		var args struct{ A, B float64 `json:"a,b"` }
		json.Unmarshal([]byte(arguments), &args)
		return fmt.Sprintf(`{"result": %g}`, args.A+args.B), nil

	default:
		return "", fmt.Errorf("unknown function: %s", functionName)
	}
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
PORT: ":8080"
CHAT_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TOOLS_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
EMBEDDING_MODEL: "ai/mxbai-embed-large"
CONTEXT_SIZE_LIMIT: 3000
SIMILARITY_THRESHOLD: 0.6
```

## Features Combined

This server provides:
- ✅ **Chat**: Conversational AI via HTTP/SSE
- ✅ **Tools**: Function calling with validation
- ✅ **RAG**: Document retrieval for context
- ✅ **Compression**: Automatic context management

## Related Patterns

- Basic server: `basic-server.md`
- Tools only: `server-with-tools.md`
- RAG only: `server-with-rag.md`
- Compressor only: `server-with-compressor.md`
