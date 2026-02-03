package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
)

func getCoderAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {
	modelID := env.GetEnvOrDefault("CODER_MODEL_ID", "hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m")
	return chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "coder",
			EngineURL:               engineURL,
			SystemInstructions:      "You are an expert programming assistant. You write clean, efficient, and well-documented code.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.8),
		},
	)
}

func getGenericAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {
	modelID := env.GetEnvOrDefault("GENERIC_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	return chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "generic",
			EngineURL:               engineURL,
			SystemInstructions:      "You respond appropriately to different types of questions. Always start with the most important information.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.8),
		},
	)
}

func main() {
	if err := os.Setenv("NOVA_LOG_LEVEL", "INFO"); err != nil {
		panic(err)
	}

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")

	ctx := context.Background()

	// ------------------------------------------------
	// Create the agent crew
	// ------------------------------------------------
	coderAgent, err := getCoderAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	genericAgent, err := getGenericAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	agentCrew := map[string]*chat.Agent{
		"coder":   coderAgent,
		"generic": genericAgent,
	}

	// ------------------------------------------------
	// Topic-to-agent routing function
	// ------------------------------------------------
	matchAgentFunction := func(currentAgentId, topic string) string {
		fmt.Println("🔵 Matching agent for topic:", topic)
		switch strings.ToLower(topic) {
		case "coding", "programming", "development", "code", "software", "debugging", "technology":
			return "coder"
		default:
			return "generic"
		}
	}

	// ------------------------------------------------
	// Tools agent - Not needed in Passthrough mode
	// ------------------------------------------------
	// In Passthrough mode, the client (qwen-code, aider, etc.)
	// manages tools and their execution.
	// The gateway simply forwards tool calls and results between
	// the client and the LLM backend.

	// ------------------------------------------------
	// Create the orchestrator agent
	// ------------------------------------------------
	orchestratorModelID := env.GetEnvOrDefault("ORCHESTRATOR_MODEL_ID", "hf.co/menlo/lucy-gguf:q4_k_m")
	orchestratorAgent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:      "orchestrator-agent",
			EngineURL: engineURL,
			SystemInstructions: `You identify the topic of a conversation in one word.
			Possible topics: Computing, Programming, Technology, Health, Science, Mathematics, Philosophy, Food, Education.
			Respond in JSON with the field 'topic_discussion'.`,
		},
		models.Config{
			Name:        orchestratorModelID,
			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	// ------------------------------------------------
	// Create the compressor agent
	// ------------------------------------------------
	compressorModelID := env.GetEnvOrDefault("COMPRESSOR_MODEL_ID", "ai/qwen2.5:0.5B-F16")

	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "compressor-agent",
			EngineURL:          engineURL,
			SystemInstructions: compressor.Instructions.Minimalist,
		},
		models.Config{
			Name:        compressorModelID,
			Temperature: models.Float64(0.0),
		},
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
	)
	if err != nil {
		panic(err)
	}

	// ------------------------------------------------
	// Create the gateway server
	// ------------------------------------------------
	// This creates an OpenAI-compatible API backed by the crew.
	// External clients (qwen-code, aider, etc.) see a single "model".
	gateway, err := gatewayserver.NewAgent(
		ctx,
		gatewayserver.WithAgentCrew(agentCrew, "generic"),
		gatewayserver.WithPort(8080),
		gatewayserver.WithOrchestratorAgent(orchestratorAgent),
		gatewayserver.WithMatchAgentIdToTopicFn(matchAgentFunction),
		gatewayserver.WithCompressorAgentAndContextSize(compressorAgent, 16384),

		// ToolModePassthrough (default): the client handles tools
		// The gateway forwards tool calls from the LLM to the client,
		// and forwards tool results from the client back to the LLM.
		// The client (qwen-code, aider, etc.) manages tool execution.

		gatewayserver.BeforeCompletion(func(agent *gatewayserver.GatewayServerAgent) {
			fmt.Printf("📥 Request received (current agent: %s)\n", agent.GetSelectedAgentId())
		}),
		gatewayserver.AfterCompletion(func(agent *gatewayserver.GatewayServerAgent) {
			fmt.Printf("📤 Response sent (agent used: %s)\n", agent.GetSelectedAgentId())
		}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("🚀 Gateway crew server starting on http://localhost:8080")
	fmt.Println("📡 OpenAI-compatible endpoint: POST /v1/chat/completions")
	fmt.Println("👥 Crew agents: coder, thinker, generic")
	fmt.Println("🔧 Tools mode: passthrough (client-side)")
	fmt.Println()
	fmt.Println("Usage with qwen-code:")
	fmt.Println(`  OPENAI_BASE_URL=http://localhost:8080/v1 OPENAI_API_KEY=none OPENAI_MODEL=crew qwen-code`)
	fmt.Println()
	fmt.Println("Usage with curl:")
	fmt.Println(`  curl http://localhost:8080/v1/chat/completions \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"model":"crew","messages":[{"role":"user","content":"Write a Go function"}],"stream":true}'`)
	fmt.Println()
	fmt.Println("📖 See README-tools.md for detailed documentation on tools usage")

	if err := gateway.StartServer(); err != nil {
		panic(err)
	}
}
