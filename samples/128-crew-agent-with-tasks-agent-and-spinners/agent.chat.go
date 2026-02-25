package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/models"
)

// crewAgentKeys lists the config keys that represent chat agents in the crew.
// These must match the agent IDs referenced in the routing rules.
var crewAgentKeys = []string{"coder", "generic"}

// createChatAgent creates the chat agents that form the crew
// and returns them as a map ready for crew.WithAgentCrew.
// Each agent is built from its entry in config.yaml + its .md instructions file.
func createChatAgent(ctx context.Context, cfg *AppConfig) (map[string]*chat.Agent, error) {
	crew := make(map[string]*chat.Agent, len(crewAgentKeys))

	for _, key := range crewAgentKeys {
		ac, err := cfg.getAgentConfig(key)
		if err != nil {
			return nil, fmt.Errorf("crew agent %q: %w", key, err)
		}

		agent, err := chat.NewAgent(
			ctx,
			agents.Config{
				Name:                    key,
				EngineURL:               cfg.EngineURL,
				SystemInstructions:      ac.Instructions,
				KeepConversationHistory: true,
			},
			models.Config{
				Name:        ac.Model,
				Temperature: models.Float64(ac.Temperature),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("crew agent %q: %w", key, err)
		}

		crew[key] = agent
	}

	return crew, nil
}
