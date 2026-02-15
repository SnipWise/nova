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
			KeepConversationHistory: true,
		},
		models.Config{
			//Name:        "ai/qwen2.5:1.5B-F16",
			Name: "hf.co/menlo/lucy-gguf:q4_k_m",
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		panic(err)
	}

	display.Info("Streaming response:")
	display.NewLine()

	// Chat with streaming - no OpenAI types exposed
	result, err := agent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is James T Kirk?"},
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

	display.NewLine()
	display.Separator()
	display.KeyValue("Finish reason", result.FinishReason)
	display.KeyValue("Context size", fmt.Sprintf("%d characters", agent.GetContextSize()))
	display.Separator()

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

	display.NewLine()
	display.Separator()
	display.KeyValue("Finish reason", result.FinishReason)
	display.KeyValue("Context size", fmt.Sprintf("%d characters", agent.GetContextSize()))
	
	// display.KeyValue("Completion ", fmt.Sprintf("%d tokens", agent.GetLastResponseMetadata().CompletionTokens))
	// display.KeyValue("Prompt ", fmt.Sprintf("%d tokens", agent.GetLastResponseMetadata().PromptTokens))
	// display.KeyValue("Total ", fmt.Sprintf("%d tokens", agent.GetLastResponseMetadata().TotalTokens))

	// display.Separator()
	// fmt.Println(agent.GetLastRequestJSON())
	// display.Separator()
	// fmt.Println(agent.GetLastResponseJSON())


	display.Separator()
}
