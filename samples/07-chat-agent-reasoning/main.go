package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Create a simple agent without exposing OpenAI SDK types
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful AI assistant that thinks step by step.",
		},
		models.Config{
			Name: "ai/qwen2.5:1.5B-F16",
			Temperature:        models.Float64(0.7),
			TopP:               models.Float64(0.9),
			ReasoningEffort: models.String(models.ReasoningEffortMedium),
		},			
	)
	if err != nil {
		panic(err)
	}

	display.Info("Streaming with reasoning:")
	display.NewLine()

	// Chat with streaming and reasoning - no OpenAI types exposed
	_, err = agent.GenerateStreamCompletionWithReasoning(
		[]messages.Message{
			{Role: roles.User, Content: "What is 15 * 24?"},
		},
		func(reasoningChunk string, finishReason string) error {
			display.Color(reasoningChunk, display.ColorYellow)
			if finishReason != "" {
				display.NewLine()
				display.KeyValue("Finish reason", finishReason)
			}
			return nil
		},
		func(responseChunk string, finishReason string) error {
			display.Color(responseChunk, display.ColorGreen)
			if finishReason != "" {
				display.NewLine()
				display.KeyValue("Finish reason", finishReason)
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()

	display.KeyValue("Context size", conversion.IntToString(agent.GetContextSize()))

	display.Separator()
}
