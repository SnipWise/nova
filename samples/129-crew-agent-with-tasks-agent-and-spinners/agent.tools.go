package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
)

// createToolsAgent creates the Tools Agent for file system operations.
// The tool index is built dynamically from the tools section of config.yaml.
func createToolsAgent(ctx context.Context, cfg *AppConfig) (*tools.Agent, error) {
	ac, err := cfg.getAgentConfig("tools")
	if err != nil {
		return nil, err
	}

	return tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "file-tools-agent",
			EngineURL:          cfg.EngineURL,
			SystemInstructions: ac.Instructions,
		},
		models.Config{
			Name:        ac.Model,
			Temperature: models.Float64(ac.Temperature),
		},
		tools.WithTools(buildToolsIndex(cfg.Tools)),
	)
}

// buildToolsIndex creates the tools list from config.yaml tool definitions.
func buildToolsIndex(toolConfigs []ToolConfig) []*tools.Tool {
	var toolsList []*tools.Tool
	for _, tc := range toolConfigs {
		tool := tools.NewTool(tc.Name).
			SetDescription(tc.Description)
		for _, p := range tc.Parameters {
			tool.AddParameter(p.Name, p.Type, p.Description, p.Required)
		}
		toolsList = append(toolsList, tool)
	}
	return toolsList
}
