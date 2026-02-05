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

	storePathFile := "./store/startrek.json"
	dataPath := "./data"

	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
		},
		models.Config{
			Name: "ai/mxbai-embed-large:latest",
		},
	)
	if err != nil {
		panic(err)
	}

	if agent.StoreFileExists(storePathFile) {
		// Load the RAG store from file
		err := agent.LoadStore(storePathFile)
		if err != nil {
			fmt.Printf("failed to load RAG store from %s: %v\n", storePathFile, err)
		}
		fmt.Printf("Successfully loaded RAG store from %s\n", storePathFile)
	} else {
		fmt.Printf("RAG store file %s does not exist. A new store will be created.\n", storePathFile)
		// Chunking + chunk enrichment
		filesContent, err := files.GetContentFilesWithNames(dataPath, ".md")
		if err != nil {
			fmt.Printf("failed to get report files: %v\n", err)
		}
		for idx, content := range filesContent {

			//contentPieces := chunks.ChunkText(content.Content, 512, 64)
			contentPieces := chunks.SplitMarkdownBySections(content.Content)

			for _, piece := range contentPieces {
				err = agent.SaveEmbedding(piece)
				if err != nil {
					fmt.Printf("failed to save embedding for document %d: %v\n", idx, err)
				} else {
					fmt.Printf("Successfully saved embedding for report %s (piece)\n", content.FileName)
				}
			}
		}
		// Save the RAG store to file
		err = agent.PersistStore(storePathFile)
		if err != nil {
			fmt.Printf("failed to persist RAG store to %s: %v\n", storePathFile, err)
		}
		fmt.Printf("Successfully saved RAG Reports store to %s\n", storePathFile)

	}

	query := "Who is Spock?"

	similarities, err := agent.SearchSimilar(query, 0.7)

	if err != nil {
		panic(err)
	}


	fmt.Printf("📝 Similarities for query: %s\n", query)
	for _, sim := range similarities {
		fmt.Println(strings.Repeat("-", 30))
		fmt.Printf("Content: %s\n", sim.Prompt)
		fmt.Printf("Score: %f\n", sim.Similarity)
	}

	query = "Who is Uhura?"

	similarities, err = agent.SearchSimilar(query, 0.7)

	if err != nil {
		panic(err)
	}

	fmt.Println(strings.Repeat("=", 30))

	fmt.Printf("📝 Similarities for query: %s\n", query)
	for _, sim := range similarities {
		fmt.Println(strings.Repeat("-", 30))
		fmt.Printf("Content: %s\n", sim.Prompt)
		fmt.Printf("Score: %f\n", sim.Similarity)
	}

}
