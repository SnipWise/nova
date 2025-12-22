package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
)

func (ca *CompositeAgent) initializeToolsAgent(ctx context.Context, engineURL string, toolsCatalog []mcp.Tool) error {

	toolsAgentSystemInstructionsContent := `
        You are an assistant that decides which tool to use based on user input.
        - Analyze the user's request carefully
        - Choose the most appropriate tool from the available options
        - Provide clear reasoning for your choice
	`

	toolsAgentModel := env.GetEnvOrDefault("TOOLS_AGENT_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	toolsAgentSystemInstructions := env.GetEnvOrDefault("TOOLS_AGENT_SYSTEM_INSTRUCTIONS", toolsAgentSystemInstructionsContent)

	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          engineURL,
			SystemInstructions: toolsAgentSystemInstructions,
		},
		models.Config{
			Name:              toolsAgentModel,
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true),
		},

		tools.WithMCPTools(toolsCatalog),
	)
	if err != nil {
		return err
	}

	ca.toolsAgent = toolsAgent
	return nil
}
