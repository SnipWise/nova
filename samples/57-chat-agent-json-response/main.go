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

	// Create a chat agent
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.8).
			WithMaxTokens(2000),
	)
	if err != nil {
		panic(err)
	}

	display.Separator()

	// First completion
	display.Info("Sending first message...")
	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is James T Kirk?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result.Response)
	display.KeyValue("Finish reason", result.FinishReason)

	
	// Get full request and response JSON
	display.NewLine()
	display.Title("Full Request JSON")
	display.Separator()
	reqJSON, err := agent.GetLastRequestJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(reqJSON)

	display.NewLine()
	display.Title("Full Response JSON")
	display.Separator()
	respJSON, err := agent.GetLastResponseJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(respJSON)

	// Second completion to show cumulative token usage
	display.NewLine()
	display.Separator()
	display.Info("Sending second message...")
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "What is his ship called?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result.Response)
	display.KeyValue("Finish reason", result.FinishReason)



	// Show conversation history
	display.NewLine()
	display.Separator()
	display.Title("Conversation History")
	display.Separator()

	historyJSON, err := agent.ExportMessagesToJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(historyJSON)

	display.NewLine()
	// Get full request and response JSON
	display.NewLine()
	display.Title("Full Request JSON")
	display.Separator()
	reqJSON, err = agent.GetLastRequestJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(reqJSON)

	display.NewLine()
	display.Title("Full Response JSON")
	display.Separator()
	respJSON, err = agent.GetLastResponseJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(respJSON)
}
