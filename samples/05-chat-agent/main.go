package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {

	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	envFile := ".env"
	// Load environment variables from env file
	if err := godotenv.Load(envFile); err != nil {
		log.Error("Warning: Error loading env file: %v\n", err)
	}

	ctx := context.Background()

	// Create a simple agent without exposing OpenAI SDK types
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
			MaxTokens:   models.Int(2000),
		},
	)
	if err != nil {
		panic(err)
	}

	// Simple chat using only Message structs
	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Hello, what is your name?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result.Response)
	display.KeyValue("Finish reason", result.FinishReason)

	// Continue the conversation
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "[Brief] who is James T Kirk?"},
	})
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()
	display.KeyValue("Response", result.Response)
	display.KeyValue("Finish reason", result.FinishReason)

	// Context is maintained automatically
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "[Brief] who is his best friend?"},
	})
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()
	display.KeyValue("Response", result.Response)
	display.KeyValue("Finish reason", result.FinishReason)

	// Display all messages
	display.NewLine()
	display.Separator()
	display.Info("Conversation history:")
	messages := agent.GetMessages()
	for i, msg := range messages {
		fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, msg.Content)
	}

	// Export conversation to JSON
	jsonData, err := agent.ExportMessagesToJSON()
	if err != nil {
		panic(err)
	}
	display.NewLine()
	display.Separator()
	display.Info("Exported conversation:")
	fmt.Println(jsonData)


	display.Separator()
	fmt.Println(agent.GetLastRequestJSON())
	display.Separator()
	fmt.Println(agent.GetLastResponseJSON())
}
