package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"

	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	ctx := context.Background()

	// Configuration from environment variables
	// These are automatically injected by Docker Agentic Compose
	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	modelID := env.GetEnvOrDefault("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
	agentName := env.GetEnvOrDefault("AGENT_NAME", "Bob")


	systemInstructions, err := files.ReadTextFile("./chat.system.instructions.md")
	if err != nil {
		display.Errorf("failed to read roleplay system instructions file: %v", err)
		return
	}

	fmt.Println(systemInstructions)
	display.Separator()

	log.Info("🚀 Starting Dockerized Chat Agent")
	log.Info("Engine URL: %s", engineURL)
	log.Info("Model: %s", modelID)
	log.Info("Agent Name: %s", agentName)

	// Model Configuration from environment variables
	temperature := env.GetEnvFloatOrDefault("TEMPERATURE", 0.7)
	maxTokens := env.GetEnvIntOrDefault("MAX_TOKENS", 32768)
	topK := env.GetEnvIntOrDefault("TOP_K", 20)
	minP := env.GetEnvFloatOrDefault("MIN_P", 0.01)
	topP := env.GetEnvFloatOrDefault("TOP_P", 0.8)

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
			Temperature: models.Float64(temperature),
			MaxTokens:   models.Int(int64(maxTokens)),
			TopK:        models.Int(int64(topK)),
			MinP:        models.Float64(minP),
			TopP:        models.Float64(topP),
		},
	)
	if err != nil {
		display.Errorf("❌ Error: %v", err)
		return
	}

	display.Info("✅ Agent initialized successfully!")
	display.Separator()

	for {

		markdownParser := display.NewMarkdownChunkParser()

		input := prompt.NewWithColor("🤖 Ask me something?").
			SetMessageColor(prompt.ColorBrightCyan).
			SetInputColor(prompt.ColorBrightWhite)

		question, err := input.Run()
		if err != nil {
			display.Errorf("❌ Error: %v", err)
			return
		}

		if strings.HasPrefix(question, "/bye") {
			fmt.Println("Goodbye!")
			break
		}

		if strings.HasPrefix(question, "/messages") {
			// Display all messages
			display.NewLine()
			display.Separator()
			display.Info("Conversation history:")
			messages := agent.GetMessages()
			for i, msg := range messages {
				fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, msg.Content)
			}
			display.Separator()
			continue
		}

		result, err := agent.GenerateStreamCompletion(
			[]messages.Message{
				{
					Role:    roles.User,
					Content: question,
				},
			},
			func(chunk string, finishReason string) error {

				if chunk != "" {
					display.MarkdownChunk(markdownParser, chunk)
				}
				if finishReason == "stop" {
					markdownParser.Flush()
					markdownParser.Reset()
					fmt.Println()
				}
				return nil
			},
		)
		if err != nil {
			panic(err)
		}
		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Context size", fmt.Sprintf("%d characters", agent.GetContextSize()))
		display.Separator()

	}

}
