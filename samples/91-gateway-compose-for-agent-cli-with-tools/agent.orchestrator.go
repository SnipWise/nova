package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
)

func loadRoutingConfig(filename string) (*orchestrator.AgentRoutingConfig, error) {
	data, err := files.ReadTextFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read routing config: %w", err)
	}

	var config orchestrator.AgentRoutingConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, fmt.Errorf("failed to parse routing config: %w", err)
	}

	return &config, nil
}

func GetOrchestratorAgent(ctx context.Context, engineURL string) (*orchestrator.Agent, error) {

	// ------------------------------------------------
	// Load routing configuration and create routing function
	// ------------------------------------------------
	routingConfig, err := loadRoutingConfig("agent-routing.json")
	if err != nil {
		panic(err)
	}


	orchestratorModelID := env.GetEnvOrDefault("ORCHESTRATOR_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	orchestratorInstructions, err := files.ReadTextFile("orchestrator.instructions.md")
	if err != nil {
		panic(err)
	}
	return orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:               "orchestrator-agent",
			EngineURL:          engineURL,
			SystemInstructions: orchestratorInstructions,
		},
		models.Config{
			Name:        orchestratorModelID,
			Temperature: models.Float64(0.0),
		},
		orchestrator.WithRoutingConfig(*routingConfig),

		orchestrator.BeforeCompletion(func(agent *orchestrator.Agent) {
			fmt.Println("🔶 Orchestrator processing request...")
		}),
	)

}
