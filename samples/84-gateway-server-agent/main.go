package main

import (
	"context"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	if err := os.Setenv("NOVA_LOG_LEVEL", "INFO"); err != nil {
		panic(err)
	}

	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	ctx := context.Background()

	// Create a single chat agent
	chatAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "assistant",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a helpful AI assistant. Be concise and accurate.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.7),
		},
	)
	if err != nil {
		panic(err)
	}

	// Create the gateway server with a single agent in passthrough mode.
	// Clients connect to this server as if it were the OpenAI API.
	gateway, err := gatewayserver.NewAgent(
		ctx,
		gatewayserver.WithSingleAgent(chatAgent),
		gatewayserver.WithPort(8080),
		// ToolModePassthrough is the default: tool_calls are forwarded to the client
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("🚀 Gateway server starting on http://localhost:8080")
	fmt.Println("📡 OpenAI-compatible endpoint: POST /v1/chat/completions")
	fmt.Println()
	fmt.Println("Usage with curl (non-streaming):")
	fmt.Println(`  curl http://localhost:8080/v1/chat/completions \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"model":"assistant","messages":[{"role":"user","content":"Hello!"}]}'`)
	fmt.Println()
	fmt.Println("Usage with curl (streaming):")
	fmt.Println(`  curl http://localhost:8080/v1/chat/completions \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"model":"assistant","messages":[{"role":"user","content":"Hello!"}],"stream":true}'`)
	fmt.Println()
	fmt.Println("Usage with qwen-code:")
	fmt.Println(`  OPENAI_BASE_URL=http://localhost:8080/v1 OPENAI_API_KEY=none qwen-code`)

	if err := gateway.StartServer(); err != nil {
		panic(err)
	}
}
