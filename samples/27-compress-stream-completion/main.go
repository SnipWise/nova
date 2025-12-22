package main

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	compressorAgent, err := compressor.NewBaseAgent(
		ctx,
		agents.Config{
			Name:               "Compressor",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: compressor.Instructions.Expert,
		},
		openai.ChatCompletionNewParams{
			Model:       "ai/qwen2.5:1.5B-F16",
			Temperature: openai.Opt(0.0),
		},
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
	)
	if err != nil {
		panic(err)
	}

	chatAgent, err := chat.NewBaseAgent(
		ctx,
		agents.Config{
			Name:               "Bob",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		openai.ChatCompletionNewParams{
			Model:       "ai/qwen2.5:latest",
			Temperature: openai.Opt(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	_, finishReason, err := chatAgent.GenerateStreamCompletion(
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Who is James T Kirk?"),
		},
		func(partialResponse string, finisReason string) error {
			fmt.Print(partialResponse)
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()
	display.KeyValue("Finish reason", finishReason)
	display.KeyValue("Context size", conversion.IntToString(chatAgent.GetCurrentContextSize()))
	display.Separator()

	// The following is a compressed summary of the previous conversation (streamed)
	display.Info("Compressing context (streaming)...")
	display.NewLine()

	newContext, _, err := compressorAgent.CompressContextStream(
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
		openai.SystemMessage(newContext),
	)

	listOfMessages := messages.ConvertFromOpenAIMessages(chatAgent.GetMessages())

	for _, msg := range listOfMessages {
		display.Color(fmt.Sprintf("[%s] %s\n", msg.Role, msg.Content), display.ColorBrightPurple)
	}

	display.KeyValue("New context size", conversion.IntToString(chatAgent.GetCurrentContextSize()))
	display.Separator()

	_, finishReason, err = chatAgent.GenerateStreamCompletion(
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Who is his best friend?"),
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
	display.KeyValue("Finish reason", finishReason)
	display.KeyValue("Context size", conversion.IntToString(chatAgent.GetCurrentContextSize()))
	display.Separator()

}
