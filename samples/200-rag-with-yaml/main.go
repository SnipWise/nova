package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
)

func main() {
	ctx := context.Background()

	storePathFile := "./store/snippets.json"
	knowledgeBase, err := files.ReadTextFile("snippets.swiftlang.yaml")
	if err != nil {
		panic(err)
	}

	// Chunk the YAML content by list item key "- id"
	txtChunks := chunks.ChunkYAML(knowledgeBase, "- id")
	_ = txtChunks // Use txtChunks if you want to load them into the store on agent creation

	// Create a RAG agent with JSON store and initial documents
	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
		},
		models.Config{
			Name: "ai/mxbai-embed-large:latest",
		},
		rag.WithJsonStore(storePathFile),
		rag.WithDocuments(txtChunks, rag.DocumentLoadModeSkip),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("✅ RAG Agent created with JSON store and YAML documents")

	fmt.Println(strings.Repeat("=", 60))

	// Test similarity search
	queries := []string{
		"How to print Hello World in Swift?",
		"How to declare variables and constants?",
		"How to use optionals in Swift?",
	}

	for _, query := range queries {
		fmt.Printf("\n🔍 Query: %s\n", query)
		fmt.Println(strings.Repeat("-", 60))

		similarities, err := agent.SearchTopN(query, 0.6, 2)
		if err != nil {
			fmt.Printf("❌ Error searching: %v\n", err)
			continue
		}

		if len(similarities) == 0 {
			fmt.Println("No similar documents found (threshold: 0.6)")
		} else {
			for i, sim := range similarities {
				fmt.Printf("%d. [%.3f] %s\n", i+1, sim.Similarity, sim.Prompt)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n📝 Demo: Loading existing store on next run")
	fmt.Println("Try running this program again - it will load the existing store from the JSON file!")
	fmt.Println("The store file is located at:", storePathFile)
}
