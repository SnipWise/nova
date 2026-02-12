package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/crewserver"
	"github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// Choose which agent to run (uncomment one)
	runServerAgent(ctx)
	// runCrewServerAgent(ctx)
	// runGatewayServerAgent(ctx)
}

// Example 1: ServerAgent with HTTPS
func runServerAgent(ctx context.Context) {
	agentConfig := agents.Config{
		Name:               "HTTPS Server Agent",
		EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
		SystemInstructions: "You are a helpful AI assistant running on HTTPS.",
	}

	modelConfig := models.Config{
		Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
		Temperature: models.Float64(0.7),
	}

	// Create agent with HTTPS using certificate files
	agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
		server.WithPort(8443),
		server.WithTLSCertFromFile("server.crt", "server.key"),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create server agent: %v", err)
	}

	fmt.Println("🔒 ServerAgent configured with HTTPS")
	fmt.Println("📡 Server will start on https://localhost:8443")
	fmt.Println("")
	fmt.Println("Test with: curl -k https://localhost:8443/health")
	fmt.Println("")

	// Start the server
	if err := agent.StartServer(); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// Example 2: CrewServerAgent with HTTPS
func runCrewServerAgent(ctx context.Context) {
	// Create multiple chat agents for the crew
	agentConfig1 := agents.Config{
		Name:               "Assistant 1",
		EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
		SystemInstructions: "You are Assistant 1, specialized in general questions.",
	}

	agentConfig2 := agents.Config{
		Name:               "Assistant 2",
		EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
		SystemInstructions: "You are Assistant 2, specialized in technical questions.",
	}

	modelConfig := models.Config{
		Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
		Temperature: models.Float64(0.7),
	}

	chatAgent1, _ := chat.NewAgent(ctx, agentConfig1, modelConfig)
	chatAgent2, _ := chat.NewAgent(ctx, agentConfig2, modelConfig)

	crew := map[string]*chat.Agent{
		"assistant1": chatAgent1,
		"assistant2": chatAgent2,
	}

	// Alternative: Using certificate data in memory
	certData, err := os.ReadFile("server.crt")
	if err != nil {
		log.Fatalf("❌ Failed to read certificate: %v", err)
	}

	keyData, err := os.ReadFile("server.key")
	if err != nil {
		log.Fatalf("❌ Failed to read private key: %v", err)
	}

	// Create agent with HTTPS using certificate data
	agent, err := crewserver.NewAgent(ctx,
		crewserver.WithAgentCrew(crew, "assistant1"),
		crewserver.WithPort(3500),
		crewserver.WithTLSCert(certData, keyData),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create crew server agent: %v", err)
	}

	fmt.Println("🔒 CrewServerAgent configured with HTTPS")
	fmt.Println("📡 Server will start on https://localhost:3500")
	fmt.Println("")
	fmt.Println("Test with: curl -k https://localhost:3500/health")
	fmt.Println("")

	// Start the server
	if err := agent.StartServer(); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// Example 3: GatewayServerAgent with HTTPS
func runGatewayServerAgent(ctx context.Context) {
	agentConfig := agents.Config{
		Name:               "HTTPS Gateway Agent",
		EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
		SystemInstructions: "You are a helpful AI assistant.",
	}

	modelConfig := models.Config{
		Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
		Temperature: models.Float64(0.7),
	}

	chatAgent, _ := chat.NewAgent(ctx, agentConfig, modelConfig)

	// Create agent with HTTPS
	agent, err := gatewayserver.NewAgent(ctx,
		gatewayserver.WithSingleAgent(chatAgent),
		gatewayserver.WithPort(8080),
		gatewayserver.WithTLSCertFromFile("server.crt", "server.key"),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create gateway server agent: %v", err)
	}

	fmt.Println("🔒 GatewayServerAgent configured with HTTPS")
	fmt.Println("📡 Server will start on https://localhost:8080")
	fmt.Println("📡 OpenAI-compatible endpoint: POST /v1/chat/completions")
	fmt.Println("")
	fmt.Println("Test with: curl -k https://localhost:8080/health")
	fmt.Println("")

	// Start the server
	if err := agent.StartServer(); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
