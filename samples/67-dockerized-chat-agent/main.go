package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	envFile := ".env"
	// Load environment variables from env file (optional in Docker)
	if err := godotenv.Load(envFile); err != nil {
		log.Info("Note: .env file not found (using Docker environment variables)\n")
	}

	ctx := context.Background()

	// Configuration from environment variables
	// These are automatically injected by Docker Agentic Compose
	engineURL := getEnv("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	modelID := getEnv("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
	agentName := getEnv("AGENT_NAME", "Bob")
	systemInstructions := getEnv("SYSTEM_INSTRUCTIONS", "You are Bob, a helpful AI assistant.")

	log.Info("🚀 Starting Dockerized Chat Agent")
	log.Info("Engine URL: %s", engineURL)
	log.Info("Model: %s", modelID)
	log.Info("Agent Name: %s", agentName)

	// Create chat agent with configuration from environment
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    agentName,
			EngineURL:               engineURL,
			SystemInstructions:      systemInstructions,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.0),
			MaxTokens:   models.Int(2000),
		},
	)
	if err != nil {
		display.Errorf("❌ Error: %v", err)
		panic(err)
	}

	display.Info("✅ Agent initialized successfully!")
	display.Separator()

	// Simple chat demonstration
	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Hello, what is your name?"},
	})
	if err != nil {
		display.Errorf("❌ Error: %v", err)
		panic(err)
	}

	display.KeyValue("Response", result.Response)
	display.KeyValue("Finish reason", result.FinishReason)

	// Continue the conversation
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "[Brief] Tell me about Docker containers."},
	})
	if err != nil {
		display.Errorf("❌ Error: %v", err)
		panic(err)
	}

	display.NewLine()
	display.Separator()
	display.KeyValue("Response", result.Response)
	display.KeyValue("Finish reason", result.FinishReason)

	// Context is maintained automatically
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "[Brief] What are the benefits?"},
	})
	if err != nil {
		display.Errorf("❌ Error: %v", err)
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
		display.Errorf("❌ Error: %v", err)
		panic(err)
	}
	display.NewLine()
	display.Separator()
	display.Info("Exported conversation:")
	fmt.Println(jsonData)

	display.NewLine()
	display.Separator()
	display.Success("✅ Dockerized chat agent demo completed!")

}
