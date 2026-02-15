package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
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

// createMatchAgentFunction creates a routing function based on the configuration
func createMatchAgentFunction(config *orchestrator.AgentRoutingConfig) func(string, string) string {
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

func main() {
	if err := os.Setenv("NOVA_LOG_LEVEL", "INFO"); err != nil {
		panic(err)
	}

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")

	ctx := context.Background()

	// ------------------------------------------------
	// Create the agent crew
	// ------------------------------------------------
	coderAgent, err := GetCoderAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	genericAgent, err := GetGenericAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	agentCrew := map[string]*chat.Agent{
		"coder":   coderAgent,
		"generic": genericAgent,
	}


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
	clientSideToolsAgent, err := GetClientSideToolsAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	toolsAgent, err := GetToolsAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	// ------------------------------------------------
	// Create the orchestrator agent
	// ------------------------------------------------
	orchestratorAgent, err := GetOrchestratorAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	matchAgentFunction := createMatchAgentFunction(orchestratorAgent.GetRoutingConfig())


	// ------------------------------------------------
	// Create the compressor agent
	// ------------------------------------------------

	compressorAgent, err := GetCompressorAgent(ctx, engineURL)
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
		gatewayserver.WithToolsAgent(toolsAgent),

		// gatewayserver.WithAgentExecutionOrder([]gatewayserver.AgentExecutionType{
		// 	gatewayserver.AgentExecutionServerSideTools,
		// 	gatewayserver.AgentExecutionClientSideTools, // Puis vérifier les outils client
		// 	gatewayserver.AgentExecutionOrchestrator,    // Router vers l'agent approprié d'abord
		// }),

		//gatewayserver.WithExecuteFn(fn func(string, string) (string, error))
		//gatewayserver.WithConfirmationPromptFn(fn func(string, string) tools.ConfirmationResponse)

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
