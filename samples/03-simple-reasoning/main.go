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
			Model:       "hf.co/menlo/lucy-gguf:q4_k_m",
			Temperature: openai.Opt(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	response, reasoning, finishReason, err := agent.GenerateCompletionWithReasoning([]openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("[Brief] who is James T Kirk?"),
	})
	if err != nil {
		panic(err)
	}

	display.Colorln(reasoning, display.ColorYellow)
	display.NewLine()
	display.Colorln(response, display.ColorGreen)

	display.NewLine()

	display.KeyValue("Finish reason", finishReason)

}
