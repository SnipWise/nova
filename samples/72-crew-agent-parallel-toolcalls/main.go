package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"

	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/crewserver"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	log := logger.GetLoggerFromEnv()

	if err := godotenv.Load(".env"); err != nil {
		log.Info("Note: .env file not found (using environment variables)")
	}

	ctx := context.Background()
	httpPort := env.GetEnvIntOrDefault("HTTP_PORT", 3500)
	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")

	// Create chat agent
	chatAgent, err := CreateChatAgent(ctx, engineURL)
	if err != nil {
		return
	}

	// Create tools agent with ParallelToolCalls enabled
	toolsAgent, err := CreateToolsAgent(ctx, engineURL)
	if err != nil {
		return
	}

	// Create a simple crew with one agent
	agentCrew := map[string]*chat.Agent{
		"calculator": chatAgent,
	}

	// Create the crew server agent
	// Because the tools agent has ParallelToolCalls: true,
	// the server will use DetectParallelToolCalls (single-pass parallel detection)
	// instead of DetectToolCallsLoopWithConfirmation (loop mode)
	crewServerAgent, err := crewserver.NewAgent(
		ctx,
		crewserver.WithAgentCrew(agentCrew, "calculator"),
		crewserver.WithPort(httpPort),
		crewserver.WithExecuteFn(executeFunction),
		crewserver.WithToolsAgent(toolsAgent),
		// NOTE: No WithConfirmationPromptFn → parallel calls run without confirmation
		// If you add WithConfirmationPromptFn, it will use
		// DetectParallelToolCallsWithConfirmation instead
	)
	if err != nil {
		display.Errorf("❌ Error creating crew server agent: %v", err)
		return
	}

	display.Infof("🚀 Crew server agent with parallel tool calls listening on port %s", crewServerAgent.GetPort())

	if err := crewServerAgent.StartServer(); err != nil {
		display.Errorf("❌ Error starting server: %v", err)
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	display.Colorf(display.ColorGreen, "🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

	switch functionName {
	case "add_numbers":
		var args struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for add_numbers"}`, nil
		}
		result := args.A + args.B
		return fmt.Sprintf(`{"result": %g}`, result), nil

	case "multiply_numbers":
		var args struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for multiply_numbers"}`, nil
		}
		result := args.A * args.B
		return fmt.Sprintf(`{"result": %g}`, result), nil

	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("👋 Hello, %s!", args.Name)
		return fmt.Sprintf(`{"message": "%s"}`, hello), nil

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
