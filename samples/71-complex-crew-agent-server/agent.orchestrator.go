package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func CreateOrchestratorAgent(ctx context.Context, engineURL string) (*orchestrator.Agent, error) {

	orchestratorAgentModelID := env.GetEnvOrDefault("ORCHESTRATOR_MODEL_ID", "hf.co/menlo/lucy-gguf:q4_k_m")
	directivesPathFile := env.GetEnvOrDefault("ORCHESTRATOR_DIRECTIVES_PATH_FILE", "./directives/orchestrator.agent.system.instructions.md")
	orchestratorAgentSystemInstructionsContent, err := files.ReadTextFile(directivesPathFile)
	if err != nil {
		return nil, err
	}

	// Create orchestrator agent
	orchestratorAgent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:               "orchestrator-agent",
			EngineURL:          engineURL,
			SystemInstructions: orchestratorAgentSystemInstructionsContent,
		},
		models.Config{
			Name:        orchestratorAgentModelID,
			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		display.Errorf("❌ Error creating orchestrator agent: %v", err)
		return nil, err
	}
	display.Infof("✅ Orchestrator agent created for agent orchestration")

	return orchestratorAgent, nil
}
