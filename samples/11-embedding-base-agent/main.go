package main

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"

	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()
	agent, err := rag.NewBaseAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		openai.EmbeddingNewParams{
			Model: "ai/mxbai-embed-large",
		},
	)
	if err != nil {
		panic(err)
	}

	embeddingVector, err := agent.GenerateEmbeddingVector("I love Hawaiian pizza!")

	if err != nil {
		panic(err)
	}

	for _, v := range embeddingVector {
		display.Colorf(display.ColorCyan, "%v", v)
	}

}
