package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func CreateThinkerAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	temperature := env.GetEnvFloatOrDefault("THINKER_TEMPERATURE", 0.8)
	topP := env.GetEnvFloatOrDefault("THINKER_TOP_P", 0.9)

	thinkerAgentModelID := env.GetEnvOrDefault("THINKER_MODEL_ID", "hf.co/menlo/lucy-gguf:q4_k_m")
	directivesPathFile := env.GetEnvOrDefault("THINKER_DIRECTIVES_PATH_FILE", "./directives/thinker.agent.system.instructions.md")
	thinkerAgentSystemInstructionsContent, err := files.ReadTextFile(directivesPathFile)
	if err != nil {
		return nil, err
	}

	// Create thinker agent
	thinkerAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "thinker",
			EngineURL:          engineURL,
			SystemInstructions: thinkerAgentSystemInstructionsContent,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        thinkerAgentModelID,
			Temperature: models.Float64(temperature),
			TopP:        models.Float64(topP),
		},
	)
	if err != nil {
		display.Errorf("❌ Error creating thinker agent: %v", err)
		return nil, err
	}
	display.Infof("✅ Thinker agent created for thoughtful assistance")

	return thinkerAgent, nil
}
