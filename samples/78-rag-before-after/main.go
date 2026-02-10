package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	embeddingCount := 0

	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "RAG",
			//EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			EngineURL: "http://localhost:12434/v1",
		},
		models.Config{
			Name: "ai/mxbai-embed-large",
		},
		// BeforeCompletion hook: called before each embedding generation
		rag.BeforeCompletion(func(a *rag.Agent) {
			embeddingCount++
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " (" + a.GetModelID() + ") - Embedding #" + conversion.IntToString(embeddingCount))
		}),
		// AfterCompletion hook: called after each embedding generation
		rag.AfterCompletion(func(a *rag.Agent) {
			display.Info("<< [AfterCompletion] Agent: " + a.GetName() + " (" + a.GetModelID() + ") - Embedding #" + conversion.IntToString(embeddingCount))
		}),
	)
	if err != nil {
		panic(err)
	}

	// === Test 1: Generate a single embedding with hooks ===
	display.NewLine()
	display.Separator()
	display.Title("Single embedding generation with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	embedding, err := agent.GenerateEmbedding("James T Kirk is the captain of the USS Enterprise.")
	if err != nil {
		panic(err)
	}

	display.KeyValue("Embedding dimension", conversion.IntToString(len(embedding)))
	display.KeyValue("First 3 values", fmt.Sprintf("[%f, %f, %f]", embedding[0], embedding[1], embedding[2]))

	// === Test 2: Save embeddings (triggers hooks for each save) ===
	display.NewLine()
	display.Separator()
	display.Title("Save embeddings into memory vector store (hooks triggered for each)")
	display.Separator()

	documents := []string{
		"Spock is a half-Vulcan, half-human science officer aboard the Enterprise.",
		"Leonard McCoy, known as Bones, is the chief medical officer of the Enterprise.",
		"Nyota Uhura is the communications officer aboard the USS Enterprise.",
	}

	for _, doc := range documents {
		err := agent.SaveEmbedding(doc)
		if err != nil {
			panic(err)
		}
		display.KeyValue("Saved", doc[:50]+"...")
	}

	// === Test 3: Search similar (triggers hooks for embedding generation) ===
	display.NewLine()
	display.Separator()
	display.Title("Search similar documents (hooks triggered for query embedding)")
	display.Separator()

	query := "Who is the doctor on the Enterprise?"
	display.KeyValue("Query", query)

	results, err := agent.SearchSimilar(query, 0.5)
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		display.KeyValue("Match", result.Prompt)
		display.KeyValue("Similarity", fmt.Sprintf("%.4f", result.Similarity))
		display.Info(strings.Repeat("-", 40))
	}

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("Total embedding generations: " + conversion.IntToString(embeddingCount))
	display.Info("BeforeCompletion and AfterCompletion hooks were triggered for each GenerateEmbedding call.")
	display.Separator()
}
