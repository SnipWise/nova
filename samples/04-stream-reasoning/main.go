package main

import (
	"context"

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

	_, _, _, err = agent.GenerateStreamCompletionWithReasoning(
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("[Brief] who is James T Kirk?"),
		},

		func(partialReasoning string, finishReason string) error {
			display.Color(partialReasoning, display.ColorYellow)
			if finishReason != "" {
				display.NewLine()
				display.KeyValue("Finish reason", finishReason)
			}
			return nil
		},

		func(partialResponse string, finishReason string) error {
			display.Color(partialResponse, display.ColorGreen)
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

	display.KeyValue("Context size", conversion.IntToString(agent.GetCurrentContextSize()))

	display.Separator()

}
