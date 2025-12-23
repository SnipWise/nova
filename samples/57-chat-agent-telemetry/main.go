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

	display.Title("Chat Agent with Telemetry")
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

	// Get telemetry data after the first request
	display.NewLine()
	display.Separator()
	display.Title("Telemetry Data - Last Request")
	display.Separator()

	// Get request metadata
	reqMetadata := agent.GetLastRequestMetadata()
	display.KeyValue("Model", reqMetadata.Model)
	display.KeyValue("Context Length", fmt.Sprintf("%d", reqMetadata.ContextLength))
	display.KeyValue("Temperature", fmt.Sprintf("%.2f", reqMetadata.Temperature))
	display.KeyValue("Max Tokens", fmt.Sprintf("%d", reqMetadata.MaxTokens))
	display.KeyValue("Request Time", reqMetadata.Timestamp.Format("2006-01-02 15:04:05"))

	// Get response metadata
	display.NewLine()
	display.Title("Telemetry Data - Last Response")
	display.Separator()

	respMetadata := agent.GetLastResponseMetadata()
	display.KeyValue("Response ID", respMetadata.ID)
	display.KeyValue("Model", respMetadata.Model)
	display.KeyValue("Finish Reason", respMetadata.FinishReason)
	display.KeyValue("Prompt Tokens", fmt.Sprintf("%d", respMetadata.PromptTokens))
	display.KeyValue("Completion Tokens", fmt.Sprintf("%d", respMetadata.CompletionTokens))
	display.KeyValue("Total Tokens", fmt.Sprintf("%d", respMetadata.TotalTokens))
	display.KeyValue("Response Time", fmt.Sprintf("%d ms", respMetadata.ResponseTime))
	display.KeyValue("Response Timestamp", respMetadata.Timestamp.Format("2006-01-02 15:04:05"))

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

	// Show updated telemetry
	display.NewLine()
	display.Separator()
	display.Title("Updated Telemetry After Second Request")
	display.Separator()

	respMetadata = agent.GetLastResponseMetadata()
	display.KeyValue("Last Response Tokens", fmt.Sprintf("%d", respMetadata.TotalTokens))
	display.KeyValue("Total Tokens Used (Session)", fmt.Sprintf("%d", agent.GetTotalTokensUsed()))
	display.KeyValue("Response Time", fmt.Sprintf("%d ms", respMetadata.ResponseTime))

	// Show conversation history
	display.NewLine()
	display.Separator()
	display.Title("Conversation History")
	display.Separator()

	historyJSON, err := agent.GetConversationHistoryJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(historyJSON)

	display.NewLine()
	display.Success("Telemetry example completed!")
}
