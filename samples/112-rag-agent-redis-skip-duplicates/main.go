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
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	// Create logger
	log := logger.GetLoggerFromEnv()

	envFile := ".env"
	if err := godotenv.Load(envFile); err != nil {
		log.Warn("Warning: Error loading env file: %v\n", err)
	}

	ctx := context.Background()

	display.Title("RAG Agent with Redis - Skip Duplicates Demo")
	display.Separator()
	display.Infof("This example demonstrates DocumentLoadModeSkipDuplicates")
	display.Infof("Run this program multiple times - it will NOT create duplicates!")
	display.Separator()

	// Configuration
	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	embeddingModel := "ai/mxbai-embed-large:latest"

	display.NewLine()
	display.Title("1. Sample Documents to Load")

	// Sample documents
	documents := []string{
		"Squirrels run in the forest and collect acorns for winter",
		"Birds fly in the sky and migrate south during winter",
		"Frogs swim in the pond and catch insects with their tongues",
		"Bears hibernate in caves during the cold winter months",
		"Rabbits hop through meadows and live in underground burrows",
	}

	display.Infof("📝 Documents to load:")
	for i, doc := range documents {
		display.Infof("   %d. %s", i+1, doc)
	}
	display.Separator()

	display.NewLine()
	display.Title("2. Creating RAG Agent with Redis + SkipDuplicates")

	// Create RAG agent with Redis and DocumentLoadModeSkipDuplicates
	// NOTE: Set LOG_LEVEL=debug in your .env file to see skip messages!
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
		display.Errorf("❌ Failed to create RAG agent: %v", err)
		return
	}

	display.Infof("✅ RAG Agent created successfully")
	display.Infof("   Redis: localhost:6379")
	display.Infof("   Index: skip_duplicates_demo")
	display.Infof("   Mode: DocumentLoadModeSkipDuplicates")
	display.NewLine()
	display.Infof("💡 Set LOG_LEVEL=debug in .env to see skip messages!")
	display.Infof("   You'll see: 'Document X already exists, skipping (duplicate)'")
	display.Separator()

	display.NewLine()
	display.Title("3. Testing Similarity Search")

	// Test similarity search
	query := "What do animals do in winter?"
	display.Infof("🔍 Query: %s", query)
	display.NewLine()

	results, err := ragAgent.SearchTopN(query, 0.3, 3)
	if err != nil {
		display.Errorf("❌ Failed to search: %v", err)
		return
	}

	display.Infof("📊 Top %d results (similarity > 0.3):", len(results))
	for i, result := range results {
		display.Infof("   %d. [%.3f] %s", i+1, result.Similarity, result.Prompt)
	}
	display.Separator()

	display.NewLine()
	display.Title("✅ Demo Complete!")
	display.Separator()
	display.Infof("🔄 Try running this program again!")
	display.Infof("   ➡️  You'll see debug logs showing documents being skipped")
	display.Infof("   ➡️  The total count in Redis will remain the same")
	display.Infof("   ➡️  No duplicates will be created!")
	display.NewLine()
	display.Infof("💡 Check debug logs above for messages like:")
	display.Infof("   'Document X already exists, skipping (duplicate)'")
	display.Separator()

	fmt.Println()
}
