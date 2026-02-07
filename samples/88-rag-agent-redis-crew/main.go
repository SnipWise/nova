package main

import (
	"context"
	"strings"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/stores"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	// Create logger
	log := logger.GetLoggerFromEnv()

	envFile := ".env"
	if err := godotenv.Load(envFile); err != nil {
		log.Error("Warning: Error loading env file: %v\n", err)
	}

	ctx := context.Background()

	display.Title("NOVA Crew with Shared Redis Knowledge Base")
	display.Separator()

	// Configuration
	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	chatModel := "huggingface.co/menlo/jan-nano-gguf:q4_k_m"
	embeddingModel := "ai/mxbai-embed-large:latest"

	display.NewLine()
	display.Title("1. Creating Shared RAG Agent with Redis")

	// Create shared RAG agent with Redis backend
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "KnowledgeBase",
			EngineURL: engineURL,
		},
		models.Config{
			Name: embeddingModel,
		},
		rag.WithRedisStore(stores.RedisConfig{
			Address:   "localhost:6379",
			Password:  "",
			DB:        0,
			IndexName: "crew_knowledge_base",
		}, 1024),
	)
	if err != nil {
		display.Errorf("❌ Failed to create RAG agent: %v", err)
		return
	}

	display.Infof("✅ Shared RAG Agent created")
	display.Infof("   Redis: localhost:6379")
	display.Infof("   Index: crew_knowledge_base")
	display.Infof("   All crew members will share this knowledge base")
	display.Separator()

	display.NewLine()
	display.Title("2. Loading Knowledge into Redis")

	// Load company knowledge
	knowledgeBase := []string{
		"The company was founded in 2020 by Alice Johnson and Bob Smith",
		"Our main product is a cloud-based project management tool called TaskMaster",
		"We have offices in San Francisco, London, and Tokyo",
		"The company employs 150 people across three continents",
		"Our annual revenue for 2023 was $12 million",
		"We recently launched a mobile app for iOS and Android",
		"The company's mission is to simplify team collaboration",
		"We offer 24/7 customer support in multiple languages",
	}

	display.Infof("📝 Loading %d knowledge items into Redis...", len(knowledgeBase))
	for i, knowledge := range knowledgeBase {
		err := ragAgent.SaveEmbedding(knowledge)
		if err != nil {
			display.Errorf("   ❌ Error saving item %d: %v", i+1, err)
			return
		}
		display.Infof("   ✓ Saved: \"%s\"", knowledge)
	}
	display.Infof("✅ Knowledge base loaded into Redis")
	display.Separator()

	display.NewLine()
	display.Title("3. Creating Crew Members")

	// Create chat agents for crew members
	researchAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "Researcher",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a research analyst. Analyze information and provide concise, factual summaries.",
			KeepConversationHistory: false,
		},
		models.Config{
			Name:        chatModel,
			Temperature: models.Float64(0.0),
			MaxTokens:   models.Int(500),
		},
	)
	if err != nil {
		display.Errorf("❌ Failed to create research agent: %v", err)
		return
	}

	writerAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "Writer",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a marketing writer. Create engaging, professional descriptions.",
			KeepConversationHistory: false,
		},
		models.Config{
			Name:        chatModel,
			Temperature: models.Float64(0.5),
			MaxTokens:   models.Int(500),
		},
	)
	if err != nil {
		display.Errorf("❌ Failed to create writer agent: %v", err)
		return
	}

	display.Infof("✅ Created Researcher agent")
	display.Infof("✅ Created Writer agent")
	display.Separator()

	display.NewLine()
	display.Title("4. Organizing Team")

	// Note: Both agents will use the same Redis knowledge base
	// This demonstrates how multiple agents can share persistent knowledge
	display.Infof("💡 Both agents share the same Redis knowledge base")
	display.Infof("   This enables consistent information across the team")
	display.Separator()

	display.NewLine()
	display.Title("5. Task 1: Research Company Information")

	// Task 1: Research using RAG
	query1 := "company founders and locations"
	display.Infof("🔍 Searching knowledge base for: \"%s\"", query1)

	results, err := ragAgent.SearchSimilar(query1, 0.3)
	if err != nil {
		display.Errorf("❌ Search error: %v", err)
		return
	}

	display.Infof("   Found %d relevant pieces of information:", len(results))
	var contextParts []string
	for i, result := range results {
		display.Colorf(display.ColorCyan, "   %d. %s", i+1, result.Prompt)
		display.Colorf(display.ColorYellow, "      Similarity: %.4f", result.Similarity)
		contextParts = append(contextParts, result.Prompt)
	}
	context1 := strings.Join(contextParts, " ")

	// Use research agent with context
	display.NewLine()
	display.Infof("📊 Researcher agent analyzing information...")
	userMessage := "Based on this context: " + context1 + "\n\nAnswer briefly: Who founded the company and where are the offices located?"

	response1, err := researchAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: userMessage},
	})
	if err != nil {
		display.Errorf("❌ Error: %v", err)
		return
	}

	display.NewLine()
	display.Colorf(display.ColorGreen, "🤖 Researcher: %s", response1.Response)
	display.Separator()

	display.NewLine()
	display.Title("6. Task 2: Product Information")

	query2 := "product and services"
	display.Infof("🔍 Searching knowledge base for: \"%s\"", query2)

	results2, err := ragAgent.SearchSimilar(query2, 0.3)
	if err != nil {
		display.Errorf("❌ Search error: %v", err)
		return
	}

	display.Infof("   Found %d relevant pieces of information:", len(results2))
	var contextParts2 []string
	for i, result := range results2 {
		display.Colorf(display.ColorCyan, "   %d. %s", i+1, result.Prompt)
		display.Colorf(display.ColorYellow, "      Similarity: %.4f", result.Similarity)
		contextParts2 = append(contextParts2, result.Prompt)
	}
	context2 := strings.Join(contextParts2, " ")

	display.NewLine()
	display.Infof("✍️  Writer agent creating description...")
	userMessage2 := "Based on this context: " + context2 + "\n\nWrite a brief, engaging product description (2-3 sentences) for marketing purposes."

	response2, err := writerAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: userMessage2},
	})
	if err != nil {
		display.Errorf("❌ Error: %v", err)
		return
	}

	display.NewLine()
	display.Colorf(display.ColorGreen, "🤖 Writer: %s", response2.Response)
	display.Separator()

	display.NewLine()
	display.Title("7. Benefits of Redis-backed RAG with Crew")
	display.Infof("💡 Key advantages demonstrated:")
	display.Infof("   ✅ Shared knowledge across all crew members")
	display.Infof("   ✅ Persistent knowledge survives restarts")
	display.Infof("   ✅ Multiple crews can access same knowledge base")
	display.Infof("   ✅ Fast semantic search with HNSW indexing")
	display.Infof("   ✅ Scalable to large knowledge bases")
	display.Separator()

	display.NewLine()
	display.Success("Demo completed successfully! 🎉")
	display.Infof("The knowledge base remains in Redis for future crew tasks.")
}
