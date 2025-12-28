package main

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
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
	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
		},
		models.Config{
			Name: "ai/mxbai-embed-large",
		},
	)
	if err != nil {
		panic(err)
	}

	txtChunks := []string{
		"Squirrels run in the forest",
		"Birds fly in the sky",
		"Frogs swim in the pond",
		"Fishes swim in the sea",
		"Lions roar in the savannah",
		"Eagles soar above the mountains",
		"Dolphins leap out of the ocean",
		"Bears fish in the river",
	}
	for _, chunk := range txtChunks {
		err := agent.SaveEmbedding(chunk)
		if err != nil {
			panic(err)
		}
	}
	query := "Which animals swim?"

	similarities, err := agent.SearchSimilar(query, 0.6)

	if err != nil {
		panic(err)
	}

	display.Colorf(display.ColorGreen, "Similarities for query: %s\n", query)
	for _, sim := range similarities {
		display.Colorf(display.ColorCyan, "Content: %s\n", sim.Prompt)
		display.Colorf(display.ColorYellow, "Score: %f\n", sim.Similarity)
	}

}
