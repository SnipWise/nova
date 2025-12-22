package main

import (
	"context"
	"strings"
	"time"

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
		models.NewConfig("hf.co/menlo/lucy-gguf:q4_k_m").
			WithTemperature(0.7).
			WithTopP(0.9),
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Color("🔴 Demo: Stream will be canceled after 15 seconds\n", display.ColorRed)
	display.Separator()
	display.NewLine()

	// Launch a goroutine to cancel the stream after 2 seconds
	go func() {
		time.Sleep(15 * time.Second)
		display.NewLine()
		display.Color("\n🛑 Canceling stream...\n", display.ColorRed)
		agent.StopStream()
	}()

	var responseBuilder strings.Builder
	var reasoningBuilder strings.Builder

	// Chat with streaming and reasoning - no OpenAI types exposed
	_, err = agent.GenerateStreamCompletionWithReasoning(
		[]messages.Message{
			{Role: roles.User, Content: "Who is James T Kirk? Write a detailed biography."},
		},
		func(reasoningChunk string, finishReason string) error {
			reasoningBuilder.WriteString(reasoningChunk)
			display.Color(reasoningChunk, display.ColorYellow)
			if finishReason != "" {
				display.NewLine()
				display.KeyValue("Reasoning finish reason", finishReason)
			}
			return nil
		},
		func(responseChunk string, finishReason string) error {
			responseBuilder.WriteString(responseChunk)
			display.Color(responseChunk, display.ColorGreen)
			if finishReason != "" {
				display.NewLine()
				display.KeyValue("Response finish reason", finishReason)
			}
			return nil
		},
	)

	display.NewLine()
	display.Separator()

	if err != nil {
		// Check if it's a cancellation error
		if err.Error() == "stream canceled by user" {
			display.NewLine()
			display.Color("✅ Stream was successfully canceled!\n", display.ColorCyan)
			display.NewLine()
			display.KeyValue("Reasoning received (chars)", conversion.IntToString(reasoningBuilder.Len()))
			display.KeyValue("Response received (chars)", conversion.IntToString(responseBuilder.Len()))
		} else {
			display.Color("❌ Error: "+err.Error()+"\n", display.ColorRed)
		}
	} else {
		display.Color("✅ Stream completed normally (not canceled)\n", display.ColorCyan)
	}

	display.NewLine()
	display.Separator()

	display.KeyValue("Context size", conversion.IntToString(agent.GetContextSize()))

	display.Separator()
	display.NewLine()
}
