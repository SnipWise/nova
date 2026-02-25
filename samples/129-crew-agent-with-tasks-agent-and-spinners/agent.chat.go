package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/models"
)

// createChatAgent creates the single chat agent (coder) for the crew.
// Since there is no orchestrator, only one chat agent is needed.
func createChatAgent(ctx context.Context, cfg *AppConfig) (*chat.Agent, error) {
	ac, err := cfg.getAgentConfig("coder")
	if err != nil {
		return nil, err
	}

	return chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "coder",
			EngineURL:               cfg.EngineURL,
			SystemInstructions:      ac.Instructions,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        ac.Model,
			Temperature: models.Float64(ac.Temperature),
		},
	)
}
