package main

import (
	"context"
	"fmt"

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
			Model:       "ai/qwen2.5:1.5B-F16",
			Temperature: openai.Opt(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	_, finishReason, err := agent.GenerateStreamCompletion(
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("[Brief] who is James T Kirk?"),
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
	display.KeyValue("Context size", conversion.IntToString(agent.GetCurrentContextSize()))

	_, finishReason, err = agent.GenerateStreamCompletion(
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("[Brief] who is his best friend?"),
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
	display.KeyValue("Context size", conversion.IntToString(agent.GetCurrentContextSize()))

}
