package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

func main() {
	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	envFile := "cerebras.env"
	// Load environment variables from env file
	if err := godotenv.Load(envFile); err != nil {
		log.Error("Warning: Error loading env file: %v\n", err)
	}

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:11434/v1")
	modelId := env.GetEnvOrDefault("CHAT_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")

	ctx := context.Background()
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-assistant",
			EngineURL:          engineURL,
			SystemInstructions: "You are Bob, a helpful AI assistant.",
			APIKey:             env.GetEnvOrDefault("CEREBRAS_API_KEY", ""),
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        modelId,
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		panic(err)
	}

	displayContextSize := func(agent *chat.Agent, result *chat.CompletionResult) {
		fmt.Println()
		fmt.Println("Finish reason:\n", result.FinishReason)
		fmt.Printf("Context size: %d characters\n", agent.GetContextSize())
		fmt.Println(strings.Repeat("-", 40))
	}

	// Chat with streaming
	result, err := agent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is James T Kirk?"},
		},
		func(chunk string, finishReason string) error {
			if chunk != "" {
				fmt.Print(chunk)
			}
			if finishReason == "stop" {
				fmt.Println()
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	displayContextSize(agent, result)

	result, err = agent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is his best friend?"},
		},
		func(chunk string, finishReason string) error {
			// Simple callback that receives strings only
			if chunk != "" {
				fmt.Print(chunk)
			}
			if finishReason == "stop" {
				fmt.Println()
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	displayContextSize(agent, result)

	fmt.Println(agent.GetLastRequestJSON())
	fmt.Println(strings.Repeat("-", 40))

	fmt.Println(agent.GetLastResponseJSON())

}
