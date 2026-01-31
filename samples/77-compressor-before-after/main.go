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

	// Create chat agent for generating conversation context
	chatAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "Bob",
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	// Create compressor agent with BeforeCompletion and AfterCompletion hooks
	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "Compressor",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: compressor.Instructions.Expert,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
		},
		// Existing base option: set the compression prompt
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
		// BeforeCompletion hook: called before each compression
		compressor.BeforeCompletion(func(a *compressor.Agent) {
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " (" + a.GetModelID() + ")")
		}),
		// AfterCompletion hook: called after each compression
		compressor.AfterCompletion(func(a *compressor.Agent) {
			display.Info("<< [AfterCompletion] Agent: " + a.GetName() + " (" + a.GetModelID() + ")")
		}),
	)
	if err != nil {
		panic(err)
	}

	// === Test 1: Standard compression with hooks ===
	display.NewLine()
	display.Separator()
	display.Title("Standard compression with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	// Generate some conversation to compress
	result, err := chatAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is James T Kirk?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Chat response", result.Response)
	display.KeyValue("Context size", conversion.IntToString(chatAgent.GetContextSize()))
	display.Separator()

	// Compress the context (standard - non-streaming)
	display.Info("Compressing context (standard)...")
	display.NewLine()

	compressedResult, err := compressorAgent.CompressContext(chatAgent.GetMessages())
	if err != nil {
		panic(err)
	}

	display.KeyValue("Compressed text", compressedResult.CompressedText)
	display.KeyValue("Finish reason", compressedResult.FinishReason)

	// === Test 2: Streaming compression with hooks ===
	display.NewLine()
	display.Separator()
	display.Title("Streaming compression with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	// Add another message to the conversation
	result2, err := chatAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Tell me about his best friend Spock."},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Chat response", result2.Response)
	display.KeyValue("Context size", conversion.IntToString(chatAgent.GetContextSize()))
	display.Separator()

	// Compress the context (streaming)
	display.Info("Compressing context (streaming)...")
	display.NewLine()

	streamResult, err := compressorAgent.CompressContextStream(
		chatAgent.GetMessages(),
		func(partialResponse string, finishReason string) error {
			fmt.Print(partialResponse)
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.KeyValue("Compressed text length", conversion.IntToString(len(streamResult.CompressedText)))
	display.KeyValue("Finish reason", streamResult.FinishReason)

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("Both standard and streaming compressions triggered the BeforeCompletion and AfterCompletion hooks.")
	display.Separator()
}
