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

	txtChunks := []string{
		"Squirrels run in the forest",
		"Birds fly in the sky",
		"Frogs swim in the pond",
		"Fishes swim in the sea",
		"Lions roar in the savannah",
		"Eagles soar above the mountains",
		"Dolphins leap out of the ocean",
		"Bears fish in the river",
	}
	for _, chunk := range txtChunks {
		err := agent.GenerateThenSaveEmbeddingVector(chunk)
		if err != nil {
			panic(err)
		}
	}
	query := "Which animals swim?"

	similarities, err := agent.SearchSimilarities(query, 0.6)

	if err != nil {
		panic(err)
	}

	display.Colorf(display.ColorGreen, "Similarities for query: %s\n", query)
	for _, sim := range similarities {
		display.Colorf(display.ColorCyan, "Content: %s\n", sim.Prompt)
		display.Colorf(display.ColorYellow, "Score: %f\n", sim.CosineSimilarity)
	}

}
