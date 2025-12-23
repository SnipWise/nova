package main

import (
	"context"
	"fmt"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Create RAG agent
	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
		},
		models.NewConfig("ai/mxbai-embed-large"),
	)
	if err != nil {
		panic(err)
	}

	display.Title("RAG Agent with Embedding Metrics")
	display.Separator()

	// Create metrics tracker
	metrics := rag.NewMetrics()

	// Sample knowledge base
	documents := []string{
		"Paris is the capital of France and known for the Eiffel Tower.",
		"Berlin is the capital of Germany and has a rich history.",
		"Rome is the capital of Italy and home to the Colosseum.",
		"Madrid is the capital of Spain and known for its art museums.",
		"London is the capital of the United Kingdom and has Big Ben.",
	}

	display.Info("Creating embeddings for knowledge base...")
	display.KeyValue("  Documents", fmt.Sprintf("%d", len(documents)))
	display.NewLine()

	// Generate and store embeddings
	for i, doc := range documents {
		startTime := time.Now()

		// Save embedding (generates and stores in one step)
		err := agent.SaveEmbedding(doc)
		if err != nil {
			panic(err)
		}

		// Get the embedding dimensions for metrics
		vector, err := agent.GenerateEmbedding(doc)
		if err != nil {
			panic(err)
		}

		elapsed := time.Since(startTime)

		// Record metrics
		metrics.RecordEmbedding(doc, len(vector), elapsed)

		display.Success(fmt.Sprintf("✓ Embedded document %d (%d chars, %d dims, %dms)",
			i+1, len(doc), len(vector), elapsed.Milliseconds()))
	}

	// Display embedding metrics
	display.NewLine()
	display.Separator()
	display.Title("📊 Embedding Generation Metrics")
	display.Separator()

	display.Info("Generation Statistics:")
	display.KeyValue("  Total Embeddings", fmt.Sprintf("%d", metrics.TotalEmbeddings))
	display.KeyValue("  Avg Dimensions", fmt.Sprintf("%d", metrics.AvgDimensions()))
	display.KeyValue("  Total Time", fmt.Sprintf("%dms", metrics.TotalProcessTime.Milliseconds()))
	display.KeyValue("  Avg Time/Embedding", fmt.Sprintf("%dms", metrics.AvgEmbeddingTime().Milliseconds()))
	display.KeyValue("  Avg Chars/Document", fmt.Sprintf("%d", metrics.AvgCharsPerDocument()))
	display.KeyValue("  Total Characters", fmt.Sprintf("%d", metrics.TotalCharacters))

	// Perform similarity searches
	display.NewLine()
	display.Separator()
	display.Title("🔍 Similarity Search with Metrics")
	display.Separator()

	queries := []string{
		"What is the capital of France?",
		"Tell me about Germany's capital",
		"Ancient Roman architecture",
	}

	for _, query := range queries {
		display.Info(fmt.Sprintf("Query: %s", query))

		startTime := time.Now()

		// Search for top 3 similar documents using the query text directly
		// SearchTopN generates the embedding internally
		results, err := agent.SearchTopN(query, 0.5, 3)
		if err != nil {
			panic(err)
		}

		elapsed := time.Since(startTime)
		metrics.RecordSearch(elapsed)

		display.KeyValue("  Search Time", fmt.Sprintf("%dms", elapsed.Milliseconds()))
		display.KeyValue("  Results Found", fmt.Sprintf("%d", len(results)))

		// Show top result
		if len(results) > 0 {
			display.KeyValue("  Top Match", results[0].Prompt)
			display.KeyValue("  Similarity", fmt.Sprintf("%.4f", results[0].Similarity))
		}
		display.NewLine()
	}

	// Search performance metrics
	display.Separator()
	display.Title("🎯 Search Performance Metrics")
	display.Separator()

	display.Info("Search Statistics:")
	display.KeyValue("  Total Searches", fmt.Sprintf("%d", metrics.SearchOperations))
	display.KeyValue("  Total Search Time", fmt.Sprintf("%dms", metrics.TotalSearchTime.Milliseconds()))
	display.KeyValue("  Avg Search Time", fmt.Sprintf("%dms", metrics.AvgSearchTime().Milliseconds()))
	display.KeyValue("  Vector Store Size", fmt.Sprintf("%d documents", metrics.TotalEmbeddings))

	// Overall efficiency metrics
	display.NewLine()
	display.Separator()
	display.Title("⚡ Overall Efficiency")
	display.Separator()

	display.Info("System Performance:")
	display.KeyValue("  Total Operations", fmt.Sprintf("%d", metrics.TotalOperations()))
	display.KeyValue("  Total Time", fmt.Sprintf("%dms", metrics.TotalTime().Milliseconds()))
	display.KeyValue("  Avg Operation Time", fmt.Sprintf("%dms", metrics.AvgOperationTime().Milliseconds()))
	display.KeyValue("  Throughput", fmt.Sprintf("%.2f ops/sec", metrics.Throughput()))

	// Cost estimation (example)
	display.NewLine()
	display.Info("Cost Estimation (example):")
	// Assuming 0.0001 USD per 1000 chars for embeddings
	costPerThousand := 0.0001
	display.KeyValue("  Embedding Cost", fmt.Sprintf("$%.6f", metrics.EstimateCost(costPerThousand)))
	display.KeyValue("  Cost per Document", fmt.Sprintf("$%.6f", metrics.CostPerDocument(costPerThousand)))

	// Show telemetry from last embedding operation
	display.NewLine()
	display.Separator()
	display.Title("📄 Last Embedding Request JSON")
	display.Separator()
	reqJSON, _ := agent.GetLastEmbeddingRequestJSON()
	fmt.Println(reqJSON)

	display.NewLine()
	display.Separator()
	display.Title("📄 Last Embedding Response JSON")
	display.Separator()
	respJSON, _ := agent.GetLastEmbeddingResponseJSON()
	fmt.Println(respJSON)

	display.NewLine()
	display.Success("RAG agent metrics example completed!")

	// reqJSON, _ = agent.GetLastEmbeddingRequestJSON()
	// fmt.Println(reqJSON)

}
