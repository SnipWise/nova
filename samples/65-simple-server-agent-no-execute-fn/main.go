package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// Create the server agent WITHOUT custom executeFunction
	// The server will use the default executeFunction from the ServerAgent
	agent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-server-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
		server.WithPort(3500),
		// No executeFunction parameter - will use default!
	)
	if err != nil {
		panic(err)
	}

	// Start the HTTP server
	fmt.Printf("🚀 Starting server agent on http://localhost%s\n", agent.GetPort())
	log.Fatal(agent.StartServer())
}
