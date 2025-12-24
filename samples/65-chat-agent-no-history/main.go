package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Create agent with KeepConversationHistory set to false
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: false, // Disable conversation history
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.0).
			WithMaxTokens(2000),
	)
	if err != nil {
		panic(err)
	}

	// First request
	display.NewLine()
	display.Separator()
	display.Title("First Request: What is your name?")
	display.Separator()

	result1, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Hello, what is your name?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result1.Response)
	display.KeyValue("Finish reason", result1.FinishReason)

	// Check messages after first request
	messages1 := agent.GetMessages()
	display.NewLine()
	display.KeyValue("Messages count after first request", conversion.IntToString(len(messages1)))
	display.Info("Expected: 1 (only system message, no user/assistant messages)")

	// Second request - asking about previous context
	display.NewLine()
	display.Separator()
	display.Title("Second Request: Who is James T Kirk?")
	display.Separator()

	result2, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is James T Kirk?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result2.Response)
	display.KeyValue("Finish reason", result2.FinishReason)

	// Check messages after second request
	messages2 := agent.GetMessages()
	display.NewLine()
	display.KeyValue("Messages count after second request", conversion.IntToString(len(messages2)))
	display.Info("Expected: 1 (still only system message)")

	// Third request - this should NOT have context from previous requests
	display.NewLine()
	display.Separator()
	display.Title("Third Request: Who is his best friend? (no context)")
	display.Separator()

	result3, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is his best friend?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result3.Response)
	display.KeyValue("Finish reason", result3.FinishReason)
	display.Info("Note: Agent should NOT know who 'his' refers to (no context)")

	// Check messages after third request
	messages3 := agent.GetMessages()
	display.NewLine()
	display.KeyValue("Messages count after third request", conversion.IntToString(len(messages3)))
	display.Info("Expected: 1 (still only system message)")

	// Display all messages to verify
	display.NewLine()
	display.Separator()
	display.Title("All Messages in History")
	display.Separator()
	for i, msg := range messages3 {
		contentPreview := msg.Content
		if len(contentPreview) > 50 {
			contentPreview = contentPreview[:50] + "..."
		}
		display.KeyValue("Message "+conversion.IntToString(i+1), string(msg.Role)+": "+contentPreview)
	}

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("With KeepConversationHistory=false, only the system message should be kept.")
	display.Info("User and assistant messages should NOT be added to history.")
	display.Info("Each request is independent without context from previous ones.")
	display.Separator()

	fmt.Println(agent.GetConversationHistoryJSON())

}
