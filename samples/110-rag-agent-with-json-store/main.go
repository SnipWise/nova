package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	storePathFile := "./store/animals.json"

	// Initial documents to load
	txtChunks := []string{
		"Squirrels run in the forest",
		"Birds fly in the sky",
		"Frogs swim in the pond",
		"Fishes swim in the sea",
		"Lions roar in the savannah",
		"Eagles soar above the mountains",
		"Dolphins leap out of the ocean",
		"Bears fish in the river",
		"Tigers prowl in the jungle",
		"Whales sing in the ocean",
		"Owls hoot at night",
		"Monkeys swing in the trees",
		"Butterflies flutter in the garden",
		"Bees buzz around flowers",
	}

	// Create a RAG agent with JSON store and initial documents
	// The WithJsonStore option will load existing data from the file if it exists
	// The WithDocuments option will initialize the store with predefined documents
	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
		},
		models.Config{
			Name: "ai/mxbai-embed-large:latest",
		},
		rag.WithJsonStore(storePathFile),
		// DocumentLoadModeOverwrite: will clear existing data and load new documents
		// DocumentLoadModeMerge: will merge new documents with existing data (default)
		// DocumentLoadModeSkip: will skip loading if store already has data
		// DocumentLoadModeError: will error if store already has data
		// DocumentLoadModeSkipDuplicates: will skip loading documents that are already in the store (based on content hash)
		rag.WithDocuments(txtChunks, rag.DocumentLoadModeSkipDuplicates),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ RAG Agent created with JSON store and initial documents")

	// Note: The store is automatically persisted when using WithJsonStore + WithDocuments
	// and new documents are added. With DocumentLoadModeSkip, documents are only added
	// if the store is empty. On subsequent runs, the existing store is loaded and no
	// persistence happens (since no new documents are added).
	fmt.Printf("📁 Store file location: %s\n", storePathFile)

	fmt.Println(strings.Repeat("=", 60))

	// Test similarity search
	queries := []string{
		"What animals live in water?",
		"Which creatures can fly?",
		"Animals in the forest",
		"What animals are in the jungle?",
		"Who sings in the ocean?",
		"Which animals are active at night?",
		"Which animals are found in the trees?",
		"Which animals buzz around flowers?",
		"Which animals are big cats?",
	}

	for _, query := range queries {
		fmt.Printf("\n🔍 Query: %s\n", query)
		fmt.Println(strings.Repeat("-", 60))

		similarities, err := agent.SearchSimilar(query, 0.6)
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

	fmt.Println("\n💡 Tips:")
	fmt.Println("- Change DocumentLoadModeOverwrite to DocumentLoadModeMerge to add documents")
	fmt.Println("- Use DocumentLoadModeSkip to keep existing data unchanged")
	fmt.Println("- Use DocumentLoadModeError to prevent accidental overwrites")
}
