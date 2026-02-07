package main

import (
	"context"
	"encoding/json"
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
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// AgentRoutingConfig represents the routing configuration
type AgentRoutingConfig struct {
	Routing []struct {
		Topics []string `json:"topics"`
		Agent  string   `json:"agent"`
	} `json:"routing"`
	DefaultAgent string `json:"default_agent"`
}

// loadRoutingConfig loads the agent routing configuration from a JSON file
func loadRoutingConfig(filename string) (*AgentRoutingConfig, error) {
	data, err := files.ReadTextFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read routing config: %w", err)
	}

	var config AgentRoutingConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, fmt.Errorf("failed to parse routing config: %w", err)
	}

	return &config, nil
}

// createMatchAgentFunction creates a routing function based on the configuration
func createMatchAgentFunction(config *AgentRoutingConfig) func(string, string) string {
	return func(currentAgentId, topic string) string {
		fmt.Println("🔵 Matching agent for topic:", topic)
		topicLower := strings.ToLower(topic)

		// Search through routing rules
		for _, rule := range config.Routing {
			for _, configTopic := range rule.Topics {
				if strings.ToLower(configTopic) == topicLower {
					return rule.Agent
				}
			}
		}

		// Return default agent if no match found
		return config.DefaultAgent
	}
}

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
	// Load routing configuration and create routing function
	// ------------------------------------------------
	routingConfig, err := loadRoutingConfig("agent-routing.json")
	if err != nil {
		panic(err)
	}

	matchAgentFunction := createMatchAgentFunction(routingConfig)

	// ------------------------------------------------
	// Tools agent - Not needed in Passthrough mode
	// ------------------------------------------------
	// In Passthrough mode, the client (qwen-code, aider, etc.)
	// manages tools and their execution.
	// The gateway simply forwards tool calls and results between
	// the client and the LLM backend.
	// ✋ **Important**: you need an agent (in the agents list) with a LLM with tool support
	// then the orchestrator will select that agent when the conversation involves tools, and the gateway will forward tool calls to the client.

	// ------------------------------------------------
	// Create the orchestrator agent
	// ------------------------------------------------
	orchestratorModelID := env.GetEnvOrDefault("ORCHESTRATOR_MODEL_ID", "hf.co/menlo/lucy-gguf:q4_k_m")
	orchestratorInstructions, err := files.ReadTextFile("orchestrator.instructions.md")
	if err != nil {
		panic(err)
	}
	orchestratorAgent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:      "orchestrator-agent",
			EngineURL: engineURL,
			SystemInstructions: orchestratorInstructions,
		},
		models.Config{
			Name:        orchestratorModelID,
			Temperature: models.Float64(0.0),
		},
		orchestrator.BeforeCompletion(func(agent *orchestrator.Agent) {
			fmt.Println("🔶 Orchestrator processing request...")
		}),
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
			messsagesFromCli := agent.GetMessages()
			for _, msg := range messsagesFromCli {
				var color string
				switch msg.Role {
				case "system":
					color = display.ColorRed
				case "user":
					color = display.ColorGreen
				case "assistant":
					color = display.ColorMagenta
				default:
					color = display.ColorBrightYellow
				}
				display.Styledln(fmt.Sprintf("   - %s: %s", msg.Role, msg.Content), color)
			}
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
