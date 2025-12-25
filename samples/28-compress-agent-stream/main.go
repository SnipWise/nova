package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Create compressor agent using simplified API
	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "Compressor",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: compressor.Instructions.Expert,
		},
		models.Config{
			Name: "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
		},
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
	)
	if err != nil {
		panic(err)
	}

	// Create chat agent using simplified API
	chatAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "Bob",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.NewConfig("ai/qwen2.5:latest").
			WithTemperature(0.0),
	)
	if err != nil {
		panic(err)
	}

	// First conversation
	result, err := chatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is James T Kirk?"},
		},
		func(partialResponse string, finishReason string) error {
			fmt.Print(partialResponse)
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()
	display.KeyValue("Finish reason", result.FinishReason)
	display.KeyValue("Context size", conversion.IntToString(chatAgent.GetContextSize()))
	display.Separator()

	// Compress context (streaming)
	display.Info("Compressing context (streaming)...")
	display.NewLine()

	// newContext
	newContext, err := compressorAgent.CompressContextStream(
		chatAgent.GetMessages(),
		func(partialResponse string, finishReason string) error {
			display.Color(partialResponse, display.ColorCyan)
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()

	// Reset chat agent messages and add new compressed context
	chatAgent.ResetMessages()

	chatAgent.AddMessage(
		roles.System,
		newContext.CompressedText,
	)

	listOfMessages := chatAgent.GetMessages()
	for _, msg := range listOfMessages {
		display.Color(fmt.Sprintf("[%s] %s\n", msg.Role, msg.Content), display.ColorBrightPurple)
	}

	display.KeyValue("New context size", conversion.IntToString(chatAgent.GetContextSize()))
	display.Separator()

	// Second conversation using compressed context
	result, err = chatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is his best friend?"},
		},
		func(partialResponse string, finishReason string) error {
			fmt.Print(partialResponse)
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()
	display.KeyValue("Finish reason", result.FinishReason)
	display.KeyValue("Context size", conversion.IntToString(chatAgent.GetContextSize()))
	display.Separator()
}
