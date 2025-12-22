package main

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()
	agent, err := chat.NewBaseAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		openai.ChatCompletionNewParams{
			Model:       "ai/qwen2.5:1.5B-F16",
			Temperature: openai.Opt(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	response, finishReason, err := agent.GenerateCompletion([]openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello, what is your name?"),
	})

	if err != nil {
		panic(err)
	}
	// display.NewLine()
	// display.Separator()
	display.KeyValue("Response", response)
	display.KeyValue("Finish reason", finishReason)

	response, finishReason, err = agent.GenerateCompletion([]openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("[Brief] who is James T Kirk?"),
	})

	if err != nil {
		panic(err)
	}
	display.NewLine()
	display.Separator()
	display.KeyValue("Response", response)
	display.KeyValue("Finish reason", finishReason)

	response, finishReason, err = agent.GenerateCompletion([]openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("[Brief] who is his best friend?"),
	})

	if err != nil {
		panic(err)
	}
	display.NewLine()
	display.Separator()
	display.KeyValue("Response", response)
	display.KeyValue("Finish reason", finishReason)
}
