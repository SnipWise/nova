package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	envFile := "huggingface.env"
	// Load environment variables from env file
	if err := godotenv.Load(envFile); err != nil {
		log.Error("Warning: Error loading env file: %v\n", err)
	}

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:11434/v1")
	modelId := env.GetEnvOrDefault("CHAT_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")

	ctx := context.Background()
	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-assistant",
			EngineURL:          engineURL,
			SystemInstructions: "You are Bob, a helpful AI assistant.",
			APIKey:             env.GetEnvOrDefault("TOKEN", ""),
			KeepConversationHistory: true,
		},
		models.Config{
			Name:              modelId,
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},
	)
	if err != nil {
		panic(err)
	}

	messages := []messages.Message{
		{
			Content: `
			Make the sum of 40 and 2,
			then say hello to Bob and to Sam,
			make the sum of 5 and 37
			Say hello to Alice
			`,
			Role: roles.User,
		},
	}

	result, err := agent.DetectToolCallsLoop(messages, executeFunction)
	if err != nil {
		panic(err)
	}

	display.KeyValue("Finish Reason", result.FinishReason)
	for _, value := range result.Results {
		display.KeyValue("Result for tool", value)
	}
	display.KeyValue("Assistant Message", result.LastAssistantMessage)

}

func GetToolsIndex() []*tools.Tool {

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

func executeFunction(functionName string, arguments string) (string, error) {

	display.Colorf(display.ColorGreen, "🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

	// here human check

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
