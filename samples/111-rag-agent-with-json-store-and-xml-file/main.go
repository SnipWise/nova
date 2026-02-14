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

	storePathFile := "./store/menu.json"
	knowledgeBase, err := files.ReadTextFile("menu.xml")
	if err != nil {
		panic(err)
	}

	// Initial documents to load
	txtChunks := chunks.ChunkXML(knowledgeBase, "item")

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
		rag.WithDocuments(txtChunks, rag.DocumentLoadModeSkip),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ RAG Agent created with JSON store and initial documents")

	// Persist the store to file
	// This saves the embeddings to the JSON file for future use
	err = agent.PersistStore(storePathFile)
	if err != nil {
		fmt.Printf("❌ Failed to persist RAG store to %s: %v\n", storePathFile, err)
	} else {
		fmt.Printf("💾 Successfully saved RAG store to %s\n", storePathFile)
	}

	fmt.Println(strings.Repeat("=", 60))

	// Test similarity search
	queries := []string{
		"What is the price of Grilled Salmon?",
		"What is the price of Chocolate Mousse?",
		"What is the price of Chicken Shawarma?",
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

	fmt.Println("\n💡 Tips:")
	fmt.Println("- Change DocumentLoadModeOverwrite to DocumentLoadModeMerge to add documents")
	fmt.Println("- Use DocumentLoadModeSkip to keep existing data unchanged")
	fmt.Println("- Use DocumentLoadModeError to prevent accidental overwrites")
}
