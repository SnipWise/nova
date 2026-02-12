package main

import (
	"context"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// This example demonstrates a minimal gateway setup with ONLY client-side tool execution.
// No orchestrator, no compressor, no RAG - just pure client-side tool calling.
//
// Use case: You want to create a simple OpenAI-compatible API that supports tool calling
// where the client executes the tools (like qwen-code, aider, continue.dev, etc.)

func main() {
	if err := os.Setenv("NOVA_LOG_LEVEL", "INFO"); err != nil {
		panic(err)
	}

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	ctx := context.Background()

	// ------------------------------------------------
	// 1. Create a single chat agent
	// ------------------------------------------------
	// This agent will handle all non-tool-related conversations
	modelID := env.GetEnvOrDefault("MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	chatAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "assistant",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a helpful AI assistant. Use tools when appropriate to help the user.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.7),
		},
		chat.BeforeCompletion(func(agent *chat.Agent) {
			display.Styledln("💬 [CHAT AGENT] Processing request...", display.ColorGreen)
		}),
	)
	if err != nil {
		panic(err)
	}

	// ------------------------------------------------
	// 2. Create the client-side tools agent
	// ------------------------------------------------
	// This agent is ONLY used to detect when the LLM wants to call tools.
	// It doesn't execute tools itself - it returns tool_calls to the client.
	clientToolsModelID := env.GetEnvOrDefault("CLIENT_TOOLS_MODEL_ID", modelID)
	clientSideToolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "client-tools",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a tool-calling assistant. When the user needs tools, identify which tools to call and with what arguments.",
			KeepConversationHistory: false, // Tools agent doesn't need conversation history
		},
		models.Config{
			Name:        clientToolsModelID,
			Temperature: models.Float64(0.0), // Low temperature for deterministic tool selection
		},
		tools.BeforeCompletion(func(agent *tools.Agent) {
			display.Styledln("🔧 [TOOLS AGENT] Detecting tool calls...", display.ColorYellow)
		}),
	)
	if err != nil {
		panic(err)
	}

	// ------------------------------------------------
	// 3. Create the minimal gateway
	// ------------------------------------------------
	gateway, err := gatewayserver.NewAgent(
		ctx,
		// Single agent crew
		gatewayserver.WithSingleAgent(chatAgent),

		// Client-side tools support
		gatewayserver.WithClientSideToolsAgent(clientSideToolsAgent),

		// Server configuration
		gatewayserver.WithPort(8080),

		// Optional: Hooks for debugging
		gatewayserver.BeforeCompletion(func(agent *gatewayserver.GatewayServerAgent) {
			fmt.Println("📥 Request received")
		}),
		gatewayserver.AfterCompletion(func(agent *gatewayserver.GatewayServerAgent) {
			fmt.Println("📤 Response sent")
		}),
	)
	if err != nil {
		panic(err)
	}

	// ------------------------------------------------
	// Start the server
	// ------------------------------------------------
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║  🚀 Minimal Gateway with Client-Side Tools                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("📡 Server: http://localhost:8080")
	fmt.Println("🔧 Client-side tools: ENABLED")
	fmt.Println("📋 Single chat agent: assistant")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  1. Client sends request with tools definitions")
	fmt.Println("  2. Gateway detects if tools are needed")
	fmt.Println("  3. If yes: returns tool_calls to client")
	fmt.Println("  4. Client executes tools locally")
	fmt.Println("  5. Client sends results back")
	fmt.Println("  6. Gateway continues completion with results")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Usage examples:")
	fmt.Println()
	fmt.Println("  With qwen-code:")
	fmt.Println("    OPENAI_BASE_URL=http://localhost:8080/v1 \\")
	fmt.Println("    OPENAI_API_KEY=none \\")
	fmt.Println("    OPENAI_MODEL=assistant \\")
	fmt.Println("    qwen-code")
	fmt.Println()
	fmt.Println("  With aider:")
	fmt.Println("    OPENAI_API_BASE=http://localhost:8080/v1 \\")
	fmt.Println("    OPENAI_API_KEY=none \\")
	fmt.Println("    aider --model assistant")
	fmt.Println()
	fmt.Println("  With curl (no tools):")
	fmt.Println("    curl http://localhost:8080/v1/chat/completions \\")
	fmt.Println("      -H 'Content-Type: application/json' \\")
	fmt.Println("      -d '{")
	fmt.Println("        \"model\": \"assistant\",")
	fmt.Println("        \"messages\": [{\"role\": \"user\", \"content\": \"Hello!\"}],")
	fmt.Println("        \"stream\": true")
	fmt.Println("      }'")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if err := gateway.StartServer(); err != nil {
		panic(err)
	}
}
