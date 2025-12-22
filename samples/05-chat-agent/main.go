package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Create a simple agent without exposing OpenAI SDK types
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.0).
			WithMaxTokens(2000),
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
}
