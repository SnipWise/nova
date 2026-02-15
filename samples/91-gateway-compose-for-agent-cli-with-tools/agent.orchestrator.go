package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
)

func GetOrchestratorAgent(ctx context.Context, engineURL string) (*orchestrator.Agent, error) {

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
		orchestrator.BeforeCompletion(func(agent *orchestrator.Agent) {
			fmt.Println("🔶 Orchestrator processing request...")
		}),
	)
	

}
