package main

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/stores"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	envFile := ".env"
	// Load environment variables from env file
	if err := godotenv.Load(envFile); err != nil {
		log.Error("Warning: Error loading env file: %v\n", err)
	}

	ctx := context.Background()

	display.Title("NOVA RAG Agent with Redis Vector Store")
	display.Separator()

	// Configuration for the embedding model
	agentConfig := agents.Config{
		EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
	}

	modelConfig := models.Config{
		Name: "ai/mxbai-embed-large:latest", // 1024 dimensions
	}

	display.NewLine()
	display.Title("1. Creating RAG Agent with Redis Store")

	// Create RAG agent with Redis backend
	agent, err := rag.NewAgent(
		ctx,
		agentConfig,
		modelConfig,
		rag.WithRedisStore(stores.RedisConfig{
			Address:   "localhost:6379",
			Password:  "",
			DB:        0,
			IndexName: "nova_rag_demo_index",
		}, 1024), // dimension must match embedding model (mxbai-embed-large = 1024)
	)
	if err != nil {
		display.Errorf("❌ Failed to create RAG agent: %v", err)
		return
	}

	display.Infof("✅ RAG Agent created with Redis vector store")
	display.Infof("   Redis: localhost:6379")
	display.Infof("   Index: nova_rag_demo_index")
	display.Infof("   Model: %s (1024 dimensions)", modelConfig.Name)
	display.Separator()

	// Sample data: animals and their habitats
	display.NewLine()
	display.Title("2. Storing Vector Embeddings in Redis")

	txtChunks := []string{
		"Squirrels run in the forest",
		"Birds fly in the sky",
		"Frogs swim in the pond",
		"Fishes swim in the sea",
		"Lions roar in the savannah",
		"Eagles soar above the mountains",
		"Dolphins leap out of the ocean",
		"Bears fish in the river",
		"Whales dive deep in the ocean",
		"Rabbits hop through the meadow",
	}

	display.Infof("📝 Storing %d text chunks as embeddings...", len(txtChunks))

	for i, chunk := range txtChunks {
		err := agent.SaveEmbedding(chunk)
		if err != nil {
			display.Errorf("   ❌ Error saving chunk %d: %v", i+1, err)
			return
		}
		display.Infof("   ✓ Saved: \"%s\"", chunk)
	}

	display.Infof("✅ All embeddings saved to Redis")
	display.Separator()

	// Search for similar content
	display.NewLine()
	display.Title("3. Searching Similar Content")

	queries := []string{
		"Which animals swim?",
		"Animals in the water",
		"Flying creatures",
	}

	for _, query := range queries {
		display.Infof("\n🔍 Query: \"%s\"", query)
		display.Infof("   Searching with similarity threshold >= 0.3...")

		similarities, err := agent.SearchSimilar(query, 0.3)
		if err != nil {
			display.Errorf("   ❌ Search error: %v", err)
			continue
		}

		if len(similarities) == 0 {
			display.Infof("   No results found")
			continue
		}

		display.Infof("   Found %d matches:", len(similarities))
		for i, sim := range similarities {
			display.Colorf(display.ColorCyan, "   %d. \"%s\"", i+1, sim.Prompt)
			display.Colorf(display.ColorYellow, "      Similarity: %.4f", sim.Similarity)
		}
	}

	display.Separator()

	// Demonstrate persistence
	display.NewLine()
	display.Title("4. Redis Persistence Demo")
	display.Infof("💡 Key benefits of Redis vector store:")
	display.Infof("   ✅ Data persists across application restarts")
	display.Infof("   ✅ Can be shared across multiple applications")
	display.Infof("   ✅ Fast HNSW-based similarity search")
	display.Infof("   ✅ Scales to millions of vectors")
	display.Separator()

	display.NewLine()
	display.Title("5. Viewing Data in Redis")
	display.Infof("📊 You can inspect the data using Redis CLI:")
	display.Infof("   docker exec -it nova-redis-vector-store redis-cli")
	display.Infof("")
	display.Infof("   Try these commands:")
	display.Infof("   • FT.INFO nova_rag_demo_index     # View index details")
	display.Infof("   • KEYS doc:*                      # List all documents")
	display.Infof("   • HGETALL doc:<uuid>              # View a specific document")
	display.Separator()

	display.NewLine()
	display.Success("Demo completed successfully! 🎉")
	display.Infof("Vectors are stored in Redis and will persist after this program exits.")
	display.Infof("Run this program again to search the same data without re-indexing!")
}
