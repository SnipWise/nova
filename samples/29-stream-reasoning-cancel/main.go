package main

import (
	"context"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()
	agent, err := chat.NewBaseAgent(
		ctx,
		agents.Config{
			Name:               "Bob",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		openai.ChatCompletionNewParams{
			Model:       "hf.co/menlo/lucy-gguf:q4_k_m",
			Temperature: openai.Opt(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Color("🔴 Demo: Stream will be canceled after 2 seconds\n", display.ColorRed)
	display.Separator()
	display.NewLine()

	// Launch a goroutine to cancel the stream after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		display.NewLine()
		display.Color("\n🛑 Canceling stream...\n", display.ColorRed)
		agent.StopStream()
	}()

	var responseBuilder strings.Builder
	var reasoningBuilder strings.Builder

	_, _, _, err = agent.GenerateStreamCompletionWithReasoning(
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Who is James T Kirk? Write a detailed biography."),
		},

		func(partialReasoning string, finishReason string) error {
			reasoningBuilder.WriteString(partialReasoning)
			display.Color(partialReasoning, display.ColorYellow)
			if finishReason != "" {
				display.NewLine()
				display.KeyValue("Reasoning finish reason", finishReason)
			}
			return nil
		},

		func(partialResponse string, finishReason string) error {
			responseBuilder.WriteString(partialResponse)
			display.Color(partialResponse, display.ColorGreen)
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

	display.KeyValue("Context size", conversion.IntToString(agent.GetCurrentContextSize()))

	display.Separator()
	display.NewLine()

}
