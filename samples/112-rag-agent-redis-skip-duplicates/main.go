package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/stores"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

func main() {
	// This example demonstrates DocumentLoadModeSkipDuplicates
	// Run this program multiple times - it will NOT create duplicates!

	// Create logger
	log := logger.GetLoggerFromEnv()

	envFile := ".env"
	if err := godotenv.Load(envFile); err != nil {
		log.Warn("Warning: Error loading env file: %v\n", err)
	}

	ctx := context.Background()

	// Configuration
	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	embeddingModel := "ai/mxbai-embed-large:latest"

	// documents
	documents := []string{
		"Squirrels run in the forest and collect acorns for winter",
		"Birds fly in the sky and migrate south during winter",
		"Frogs swim in the pond and catch insects with their tongues",
		"Bears hibernate in caves during the cold winter months",
		"Rabbits hop through meadows and live in underground burrows",
	}

	// Create RAG agent with Redis and DocumentLoadModeSkipDuplicates
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "SkipDuplicatesDemo",
			EngineURL: engineURL,
		},
		models.Config{
			Name: embeddingModel,
		},
		rag.WithRedisStore(stores.RedisConfig{
			Address:   "localhost:6379",
			Password:  "",
			DB:        0,
			IndexName: "skip_duplicates_demo",
		}, 1024),
		rag.WithDocuments(documents, rag.DocumentLoadModeSkipDuplicates),
	)
	if err != nil {
		fmt.Printf("❌ Failed to create RAG agent: %v\n", err)
		return
	}

	fmt.Println("✅ RAG Agent created with Redis store and initial documents")
	fmt.Println()

	fmt.Println("Testing Similarity Search...")

	// Test similarity search
	query := "What do animals do in winter?"
	fmt.Printf("🔍 Query: %s\n", query)
	fmt.Println()

	results, err := ragAgent.SearchTopN(query, 0.3, 3)
	if err != nil {
		fmt.Printf("❌ Failed to search: %v\n", err)
		return
	}

	fmt.Printf("📊 Top %d results (similarity > 0.3):\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. [%.3f] %s\n", i+1, result.Similarity, result.Prompt)
	}

	fmt.Println()
}
