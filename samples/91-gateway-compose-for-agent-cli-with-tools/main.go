package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
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
		chat.BeforeCompletion(func(agent *chat.Agent) {
			display.Styledln("🔧 [CODER AGENT] Processing request...", display.ColorCyan)
		}),
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
		chat.BeforeCompletion(func(agent *chat.Agent) {
			display.Styledln("💬 [GENERIC AGENT] Processing request...", display.ColorGreen)
		}),
	)
}

func getClientSideToolsAgent(ctx context.Context, engineURL string) (*tools.Agent, error) {
	modelID := env.GetEnvOrDefault("CLIENT_SIDE_TOOLS_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	return tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "client-side-tools",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a helpful assistant that can use tools when needed.",
			KeepConversationHistory: false, // Tools agent doesn't need history
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.0),
		},
		tools.BeforeCompletion(func(agent *tools.Agent) {
			display.Styledln("🔀 [CLIENT-SIDE TOOLS] Detecting tool calls...", display.ColorYellow)
		}),
	)
}

func getToolsAgent(ctx context.Context, engineURL string) (*tools.Agent, error) {
	modelID := env.GetEnvOrDefault("TOOLS_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")

	getToolsIndex := func() []*tools.Tool {

		calculateSumTool := tools.NewTool("calculate_sum").
			SetDescription("Calculate the sum of two numbers").
			AddParameter("a", "number", "The first number", true).
			AddParameter("b", "number", "The second number", true)

		sayHelloTool := tools.NewTool("say_hello").
			SetDescription("Say hello to the given name").
			AddParameter("name", "string", "The name to greet", true)

		return []*tools.Tool{
			calculateSumTool,
			sayHelloTool,
		}
	}

	executeFunction := func(functionName string, arguments string) (string, error) {

		display.Colorf(display.ColorGreen, "🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

		switch functionName {
		case "say_hello":
			var args struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal([]byte(arguments), &args); err != nil {
				return `{"error": "Invalid arguments for say_hello"}`, nil
			}
			hello := fmt.Sprintf("👋 Hello, %s!🙂", args.Name)
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

		default:
			return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
		}
	}

	return tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "tools",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a helpful assistant that can use tools when needed.",
			KeepConversationHistory: false, // Tools agent doesn't need history
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.0),
		},
		tools.BeforeCompletion(func(agent *tools.Agent) {
			display.Styledln("🔀 [TOOLS] Detecting tool calls...", display.ColorYellow)
		}),
		tools.WithTools(getToolsIndex()),
		//tools.WithExecuteFunction(executeFunction),
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
	// Create the client-side tools agent
	// ------------------------------------------------
	// The client-side tools agent detects when tools are needed
	// and returns tool_calls to the client in OpenAI format.
	// The client (qwen-code, aider, continue.dev, etc.) then:
	// 1. Executes the tools locally
	// 2. Sends results back as messages with role "tool"
	// 3. The gateway continues the completion with those results
	//
	// This is the standard "client-side tool execution" pattern used by
	// most AI coding assistants.
	clientSideToolsAgent, err := getClientSideToolsAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	// ------------------------------------------------
	// Create the orchestrator agent
	// ------------------------------------------------
	orchestratorModelID := env.GetEnvOrDefault("ORCHESTRATOR_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	orchestratorInstructions, err := files.ReadTextFile("orchestrator.instructions.md")
	if err != nil {
		panic(err)
	}
	orchestratorAgent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:               "orchestrator-agent",
			EngineURL:          engineURL,
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
		gatewayserver.WithClientSideToolsAgent(clientSideToolsAgent),
		gatewayserver.WithOrchestratorAgent(orchestratorAgent),
		gatewayserver.WithMatchAgentIdToTopicFn(matchAgentFunction),
		gatewayserver.WithCompressorAgentAndContextSize(compressorAgent, 16384),
		gatewayserver.WithExecuteFn(fn func(string, string) (string, error))

		// Agent execution order (default):
		// 1. ClientSideTools - Detects tool calls and returns them to client
		// 2. ServerSideTools - (not configured in this example)
		// 3. Orchestrator - Routes to appropriate agent based on topic

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
	fmt.Println("👥 Crew agents: coder, generic")
	fmt.Println("🔧 Client-side tools: enabled (tools executed by client)")
	fmt.Println("🎯 Orchestrator: enabled (topic-based routing)")
	fmt.Println("🗜️  Compressor: enabled (context compression)")
	fmt.Println()
	fmt.Println("Usage with qwen-code (with tools):")
	fmt.Println(`  OPENAI_BASE_URL=http://localhost:8080/v1 OPENAI_API_KEY=none OPENAI_MODEL=crew qwen-code`)
	fmt.Println()
	fmt.Println("Usage with aider (with tools):")
	fmt.Println(`  OPENAI_API_BASE=http://localhost:8080/v1 OPENAI_API_KEY=none aider --model crew`)
	fmt.Println()
	fmt.Println("Usage with curl (no tools):")
	fmt.Println(`  curl http://localhost:8080/v1/chat/completions \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"model":"crew","messages":[{"role":"user","content":"Write a Go function"}],"stream":true}'`)
	fmt.Println()
	fmt.Println("📖 The gateway automatically detects tool calls and returns them to the client")

	if err := gateway.StartServer(); err != nil {
		panic(err)
	}
}
