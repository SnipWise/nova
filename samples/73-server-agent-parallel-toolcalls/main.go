package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
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
	chatModelID := env.GetEnvOrDefault("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
	directivesPathFile := env.GetEnvOrDefault("CHAT_DIRECTIVES_PATH_FILE", "./directives/chat.agent.system.instructions.md")

	systemInstructions, err := files.ReadTextFile(directivesPathFile)
	if err != nil {
		display.Errorf("❌ Error reading directives: %v", err)
		return
	}

	// Create tools agent with ParallelToolCalls enabled
	toolsAgent, err := CreateToolsAgent(ctx, engineURL)
	if err != nil {
		return
	}

	// Create the server agent
	// Because the tools agent has ParallelToolCalls: true,
	// the server will use DetectParallelToolCalls (single-pass parallel detection)
	// instead of DetectToolCallsLoopWithConfirmation (loop mode)
	serverAgent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:               "server-agent",
			EngineURL:          engineURL,
			SystemInstructions: systemInstructions,
		},
		models.Config{
			Name:        chatModelID,
			Temperature: models.Float64(0.7),
		},
		server.WithPort(httpPort),
		server.WithExecuteFn(executeFunction),
		server.WithToolsAgent(toolsAgent),
		// NOTE: No WithConfirmationPromptFn → parallel calls run without confirmation
		// If you add WithConfirmationPromptFn, it will use
		// DetectParallelToolCallsWithConfirmation instead
	)
	if err != nil {
		display.Errorf("❌ Error creating server agent: %v", err)
		return
	}

	display.Infof("🚀 Server agent with parallel tool calls listening on port %s", serverAgent.GetPort())

	if err := serverAgent.StartServer(); err != nil {
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
